package fyneui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2/widget"

	"vadlp/internal/core"
	"vadlp/internal/downloader"
	"vadlp/internal/i18n"
	"vadlp/internal/service"
	"vadlp/internal/settings"
)

type downloadRunner struct {
	tr                   func(string) string
	addJournal           func(string, error)
	saveAppSettings      func()
	sessionPathEntry     *widget.Entry
	snapMu               *sync.Mutex
	lastSession          *core.Session
	logs                 *widget.Entry
	progressFile         *widget.ProgressBar
	progressOverall      *widget.ProgressBar
	progressFileLabel    *widget.Label
	progressOverallLabel *widget.Label
	statusBadge          *StatusBadge
	phaseBadge           *PhaseBadge
	stopBtn              *widget.Button
	runBtn               *widget.Button
	refreshHistory       func()
	updateTaskStatus     func(string, string)
	cfg                  *core.Config
	appSettings          *settings.App
	running              *atomic.Bool
	currentJobID         *atomic.Value
	svc                  *service.Service
	jobCancel            context.CancelFunc
	cancelMu             sync.Mutex
}

func (r *downloadRunner) persistSnapshot(path string, snap core.Session) {
	if strings.TrimSpace(path) == "" {
		return
	}
	if r.svc == nil {
		r.svc = service.New()
	}
	if err := r.svc.SaveSession(path, snap); err != nil {
		r.addJournal(r.tr("err.save_session"), err)
	}
}

func (r *downloadRunner) cancelActiveJob() {
	r.cancelMu.Lock()
	cancel := r.jobCancel
	r.cancelMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (r *downloadRunner) runDownload(current core.Config, taskID string, qIdx, qTot int, focusUI bool) bool {
	jobID := taskID
	if jobID == "" {
		jobID = "main"
	}
	r.currentJobID.Store(jobID)
	defer r.currentJobID.Store("")

	if focusUI {
		uiExec(func() {
			r.logs.SetText("")
			r.progressFile.SetValue(0)
			r.progressOverall.SetValue(0)
			r.statusBadge.SetStatusKey("running")
			r.phaseBadge.SetPhase(downloader.StageUnknown)
			r.progressFileLabel.SetText(r.tr("progress.file"))
			if qTot > 1 {
				r.progressOverallLabel.SetText(i18n.T("progress.overall_queue", map[string]interface{}{
					"Current": qIdx,
					"Total":   qTot,
				}))
			} else {
				r.progressOverallLabel.SetText(r.tr("progress.overall"))
			}
			r.stopBtn.Enable()
			r.runBtn.Disable()
		})
	}
	if focusUI {
		defer uiExec(func() {
			r.stopBtn.Disable()
			r.runBtn.Enable()
		})
	}

	plCur, plTot := 0, 0
	filePct := 0.0

	var persistMu sync.Mutex
	lastPersistWrite := time.Time{}
	persistSnapshotThrottled := func(path string, snap core.Session) {
		if strings.TrimSpace(path) == "" {
			return
		}
		persistMu.Lock()
		defer persistMu.Unlock()
		if time.Since(lastPersistWrite) < 2*time.Second {
			return
		}
		lastPersistWrite = time.Now()
		r.persistSnapshot(path, snap)
	}

	updateSnap := func() {
		snap := core.Session{
			Config:          current,
			PlaylistCurrent: plCur,
			PlaylistTotal:   plTot,
			QueueDone:       maxInt(0, qIdx-1),
			QueueTotal:      qTot,
		}
		r.snapMu.Lock()
		*r.lastSession = snap
		r.snapMu.Unlock()
		persistSnapshotThrottled(strings.TrimSpace(r.sessionPathEntry.Text), snap)
	}

	computeOverall := func() float64 {
		if plTot > 0 && plCur > 0 {
			return (float64(plCur-1) + filePct/100.0) / float64(plTot) * 100.0
		}
		if qTot > 0 && qIdx > 0 {
			return (float64(qIdx-1) + filePct/100.0) / float64(qTot) * 100.0
		}
		return filePct
	}

	pushOverall := func() {
		if !focusUI {
			return
		}
		v := computeOverall()
		uiExec(func() {
			r.progressOverall.SetValue(v)
			if plTot > 0 && plCur > 0 {
				r.progressOverallLabel.SetText(i18n.T("progress.overall_playlist", map[string]interface{}{
					"Current": plCur,
					"Total":   plTot,
				}))
			} else if qTot > 1 {
				r.progressOverallLabel.SetText(i18n.T("progress.overall_queue", map[string]interface{}{
					"Current": qIdx,
					"Total":   qTot,
				}))
			}
		})
	}

	var localLogs []string

	ctx, cancel := context.WithCancel(context.Background())
	r.cancelMu.Lock()
	r.jobCancel = cancel
	r.cancelMu.Unlock()
	defer func() {
		cancel()
		r.cancelMu.Lock()
		r.jobCancel = nil
		r.cancelMu.Unlock()
	}()

	if r.svc == nil {
		r.svc = service.New()
	}
	result, err := r.svc.Download(ctx, current, jobID, func(ev downloader.Event) {
		switch ev.Type {
		case downloader.EventLog:
			line := ev.LogLine
			if focusUI && ev.Stage != downloader.StageUnknown {
				st := ev.Stage
				uiExec(func() { r.phaseBadge.SetPhase(st) })
			}
			if !strings.HasPrefix(strings.TrimSpace(line), "[download]") || !strings.Contains(line, "%") {
				localLogs = append(localLogs, line)
				if len(localLogs) > 450 {
					localLogs = localLogs[len(localLogs)-450:]
				}
				if focusUI {
					uiExec(func() {
						r.logs.SetText(strings.Join(localLogs, "\n"))
						r.logs.CursorRow = len(localLogs) - 1
					})
				}
			}
		case downloader.EventProgress:
			filePct = ev.Progress
			if focusUI {
				uiExec(func() {
					r.progressFile.SetValue(ev.Progress)
					if ev.Speed != "" || ev.ETA != "" {
						r.progressFileLabel.SetText(i18n.T("progress.speed_eta", map[string]interface{}{
							"Speed": ev.Speed,
							"ETA":   ev.ETA,
						}))
					}
				})
			}
			pushOverall()
			updateSnap()
		case downloader.EventPlaylist:
			plCur = ev.PlaylistCurrent
			plTot = ev.PlaylistTotal
			pushOverall()
			updateSnap()
		}
	})

	if err != nil {
		st := "error"
		msg := err.Error()
		if downloader.IsCancelled(err) || errors.Is(err, context.Canceled) {
			st = "cancelled"
			msg = "cancelled"
		}
		if focusUI {
			uiExec(func() {
				if downloader.IsCancelled(err) {
					r.statusBadge.SetStatusKey("stopped")
				} else {
					r.statusBadge.SetStatusKey("error")
				}
				r.phaseBadge.SetPhase(downloader.StageUnknown)
				localLogs = append(localLogs, strings.ToUpper(msg))
				r.logs.SetText(strings.Join(localLogs, "\n"))
			})
		}
		if !downloader.IsCancelled(err) && !errors.Is(err, context.Canceled) {
			journalFromErr(r.tr, r.addJournal, "err.download_failed", err)
		}
		_ = r.svc.AppendHistory(core.HistoryItem{
			URL:         firstNonEmpty(current.URL, current.BatchURLs),
			Status:      st,
			Output:      current.OutputPath,
			Error:       msg,
			DurationSec: result.DurationSec,
		})
		if r.refreshHistory != nil {
			uiExec(r.refreshHistory)
		}
		if taskID != "" {
			r.updateTaskStatus(taskID, st)
		}
		updateSnap()
		return !downloader.IsCancelled(err) && !errors.Is(err, context.Canceled)
	}

	if focusUI {
		uiExec(func() {
			r.statusBadge.SetStatusKey("completed")
			r.phaseBadge.SetPhase(downloader.StageUnknown)
			r.progressFile.SetValue(100)
			r.progressOverall.SetValue(100)
		})
	}
	_ = r.svc.AppendHistory(core.HistoryItem{
		URL:         firstNonEmpty(current.URL, current.BatchURLs),
		Status:      "completed",
		Output:      current.OutputPath,
		DurationSec: result.DurationSec,
	})
	if taskID != "" {
		r.updateTaskStatus(taskID, "completed")
	}
	if r.refreshHistory != nil {
		uiExec(r.refreshHistory)
	}
	plCur, plTot = 0, 0
	filePct = 100
	updateSnap()
	return true
}

type queueController struct {
	runner           *downloadRunner
	queue            *[]QueueTask
	queueLock        *sync.Mutex
	queueList        *widget.List
	selectedQueueIdx *int
	cfg              *core.Config
	addJournal       func(string, error)
	tr               func(string) string
}

func (q *queueController) wireButtons(
	runBtn *widget.Button,
	addQueueBtn, runQueueBtn, removeQueueBtn, retryQueueBtn, moveQueueUpBtn, moveQueueDownBtn, clearQueueBtn *widget.Button,
) {
	runBtn.OnTapped = func() {
		if !q.runner.running.CompareAndSwap(false, true) {
			return
		}
		go func(localCfg core.Config) {
			defer q.runner.running.Store(false)
			q.runner.runDownload(localCfg, "", 1, 1, true)
			q.runner.saveAppSettings()
		}(*q.cfg)
	}

	addQueueBtn.OnTapped = func() {
		if strings.TrimSpace(q.cfg.URL) == "" && strings.TrimSpace(q.cfg.LoadInfoJSON) == "" && strings.TrimSpace(q.cfg.BatchURLs) == "" {
			q.addJournal(q.tr("err.queue_no_url"), fmt.Errorf("%s", q.tr("err.queue_no_url")))
			return
		}
		q.queueLock.Lock()
		*q.queue = append(*q.queue, QueueTask{
			ID:     fmt.Sprintf("q-%d", time.Now().UnixNano()),
			Name:   firstNonEmpty(q.cfg.URL, q.cfg.LoadInfoJSON),
			Config: *q.cfg,
			Status: "queued",
		})
		q.queueLock.Unlock()
		q.queueList.Refresh()
	}

	runQueueBtn.OnTapped = func() {
		if !q.runner.running.CompareAndSwap(false, true) {
			return
		}
		go func() {
			defer q.runner.running.Store(false)
			q.queueLock.Lock()
			local := make([]QueueTask, 0, len(*q.queue))
			for _, t := range *q.queue {
				if t.Status == "queued" {
					local = append(local, t)
				}
			}
			q.queueLock.Unlock()

			qTot := len(local)
			workers := q.runner.appSettings.QueueParallel
			if workers < 1 {
				workers = 1
			}

			runOne := func(i int, task QueueTask) bool {
				q.runner.updateTaskStatus(task.ID, "running")
				focus := workers == 1
				return q.runner.runDownload(task.Config, task.ID, i+1, qTot, focus)
			}

			allOK := true
			if workers == 1 {
				for i := range local {
					if !runOne(i, local[i]) {
						allOK = false
						break
					}
				}
			} else {
				var wg sync.WaitGroup
				sem := make(chan struct{}, workers)
				var failMu sync.Mutex
				for i := range local {
					wg.Add(1)
					go func(idx int, task QueueTask) {
						defer wg.Done()
						sem <- struct{}{}
						defer func() { <-sem }()
						if !runOne(idx, task) {
							failMu.Lock()
							allOK = false
							failMu.Unlock()
						}
					}(i, local[i])
				}
				wg.Wait()
			}

			uiExec(func() {
				if allOK {
					q.runner.statusBadge.SetStatusKey("ready")
					q.runner.phaseBadge.SetPhase(downloader.StageUnknown)
				}
			})
			q.runner.saveAppSettings()
		}()
	}

	removeQueueBtn.OnTapped = func() {
		if q.runner.running.Load() || *q.selectedQueueIdx < 0 {
			return
		}
		q.queueLock.Lock()
		if *q.selectedQueueIdx < len(*q.queue) {
			*q.queue = append((*q.queue)[:*q.selectedQueueIdx], (*q.queue)[*q.selectedQueueIdx+1:]...)
			*q.selectedQueueIdx = -1
		}
		q.queueLock.Unlock()
		q.queueList.Refresh()
	}

	retryQueueBtn.OnTapped = func() {
		q.queueLock.Lock()
		for i := range *q.queue {
			if (*q.queue)[i].Status == "error" || (*q.queue)[i].Status == "cancelled" {
				(*q.queue)[i].Status = "queued"
			}
		}
		q.queueLock.Unlock()
		q.queueList.Refresh()
	}

	moveQueueUpBtn.OnTapped = func() {
		if q.runner.running.Load() || *q.selectedQueueIdx <= 0 {
			return
		}
		q.queueLock.Lock()
		i := *q.selectedQueueIdx
		if i < len(*q.queue) {
			(*q.queue)[i-1], (*q.queue)[i] = (*q.queue)[i], (*q.queue)[i-1]
			*q.selectedQueueIdx--
		}
		q.queueLock.Unlock()
		q.queueList.Refresh()
	}

	moveQueueDownBtn.OnTapped = func() {
		q.queueLock.Lock()
		i := *q.selectedQueueIdx
		if !q.runner.running.Load() && i >= 0 && i < len(*q.queue)-1 {
			(*q.queue)[i+1], (*q.queue)[i] = (*q.queue)[i], (*q.queue)[i+1]
			*q.selectedQueueIdx++
		}
		q.queueLock.Unlock()
		q.queueList.Refresh()
	}

	clearQueueBtn.OnTapped = func() {
		if q.runner.running.Load() {
			return
		}
		q.queueLock.Lock()
		*q.queue = nil
		q.queueLock.Unlock()
		q.queueList.Refresh()
	}
}
