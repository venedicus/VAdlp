package fyneui

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"ytgui/internal/core"
	"ytgui/internal/downloader"
)

type QueueTask struct {
	ID     string
	Name   string
	Config core.Config
	Status string
}

func Run() {
	a := app.NewWithID("ytgui.fyne")
	a.Settings().SetTheme(NewTokyoNightTheme())
	w := a.NewWindow("yt-dlp GUI (Fyne)")
	w.Resize(fyne.NewSize(1120, 760))

	cfg := core.DefaultConfig()
	var issueEntries []string

	statusLabel := widget.NewLabel("READY")
	issuesBtn := widget.NewButton("Issues (0)", nil)
	
	commandPreview := widget.NewMultiLineEntry()
	commandPreview.Disable()
	commandPreview.Wrapping = fyne.TextWrapWord
	commandPreview.SetMinRowsVisible(4)
	commandPreview.TextStyle = fyne.TextStyle{Monospace: true} // TUI aesthetics

	progress := widget.NewProgressBar()
	progress.Min = 0
	progress.Max = 100

	logs := widget.NewMultiLineEntry()
	logs.Disable()
	logs.SetMinRowsVisible(18)
	logs.Wrapping = fyne.TextWrapWord
	logs.TextStyle = fyne.TextStyle{Monospace: true} // TUI aesthetics

	addIssue := func(summary string, err error) {
		entry := summary
		if err != nil {
			entry = summary + ": " + err.Error()
		}
		timestamp := time.Now().Format("15:04:05")
		issueEntries = append(issueEntries, "["+timestamp+"] "+entry)
		issuesBtn.SetText(fmt.Sprintf("Issues (%d)", len(issueEntries)))
	}

	issuesBtn.OnTapped = func() {
		content := "No issues recorded."
		if len(issueEntries) > 0 {
			content = strings.Join(issueEntries, "\n")
		}
		issueView := widget.NewMultiLineEntry()
		issueView.SetText(content)
		issueView.Disable()
		issueView.SetMinRowsVisible(18)
		dialog.NewCustom("Issues", "Close", container.NewScroll(issueView), w).Show()
	}

	updatePreview := func() {
		commandPreview.SetText(core.PreviewCommand(cfg))
	}

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("https://www.youtube.com/watch?v=...")
	urlEntry.OnChanged = func(s string) {
		cfg.URL = s
		updatePreview()
	}

	qualitySelect := widget.NewSelect([]string{
		"best",
		"bestvideo[height<=1080]+bestaudio/best[height<=1080]",
		"bestvideo[height<=720]+bestaudio/best[height<=720]",
	}, func(s string) {
		cfg.Quality = s
		updatePreview()
	})
	qualitySelect.SetSelected(cfg.Quality)

	formatSelect := widget.NewSelect([]string{"mp4", "webm", ""}, func(s string) {
		cfg.Format = s
		updatePreview()
	})
	formatSelect.SetSelected(cfg.Format)

	audioCheck := widget.NewCheck("Audio only (-x)", func(b bool) {
		cfg.AudioOnly = b
		updatePreview()
	})
	audioCheck.SetChecked(cfg.AudioOnly)

	pathEntry := widget.NewEntry()
	pathEntry.SetText(cfg.OutputPath)
	pathEntry.OnChanged = func(s string) {
		cfg.OutputPath = s
		updatePreview()
	}
	pickFolderBtn := widget.NewButton("Browse", func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				addIssue("Folder picker error", err)
				return
			}
			if uri != nil {
				pathEntry.SetText(uri.Path())
				cfg.OutputPath = uri.Path()
				updatePreview()
			}
		}, w).Show()
	})

	templateEntry := widget.NewEntry()
	templateEntry.SetText(cfg.OutputTemplate)
	templateEntry.OnChanged = func(s string) {
		cfg.OutputTemplate = s
		updatePreview()
	}

	cookiesBrowserCheck := widget.NewCheck("Cookies from browser", func(b bool) {
		cfg.UseCookiesBrowser = b
		updatePreview()
	})
	cookiesBrowserCheck.SetChecked(cfg.UseCookiesBrowser)

	cookiesBrowserSelect := widget.NewSelect([]string{"chrome", "firefox", "vivaldi", "edge", "brave"}, func(s string) {
		cfg.CookiesBrowser = s
		updatePreview()
	})
	cookiesBrowserSelect.SetSelected(cfg.CookiesBrowser)

	cookiesFileCheck := widget.NewCheck("Cookies from file", func(b bool) {
		cfg.UseCookiesFile = b
		updatePreview()
	})

	cookiesFileEntry := widget.NewEntry()
	cookiesFileEntry.SetPlaceHolder("C:\\path\\cookies.txt")
	cookiesFileEntry.OnChanged = func(s string) {
		cfg.CookiesFile = s
		updatePreview()
	}

	proxyEntry := widget.NewEntry()
	proxyEntry.SetPlaceHolder("http://127.0.0.1:8080")
	proxyEntry.OnChanged = func(s string) {
		cfg.Proxy = s
		updatePreview()
	}

	rateEntry := widget.NewEntry()
	rateEntry.SetPlaceHolder("1M, 500K")
	rateEntry.OnChanged = func(s string) {
		cfg.RateLimit = s
		updatePreview()
	}

	reverseCheck := widget.NewCheck("Reverse playlist", func(b bool) {
		cfg.PlaylistReverse = b
		updatePreview()
	})
	reverseCheck.SetChecked(cfg.PlaylistReverse)

	updatePreview()

	var (
		queue     []QueueTask
		queueLock sync.Mutex
		running   atomic.Bool
	)

	queueList := widget.NewList(
		func() int {
			queueLock.Lock()
			defer queueLock.Unlock()
			return len(queue)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel("task"),
				layout.NewSpacer(),
				widget.NewLabel("queued"),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			queueLock.Lock()
			item := queue[i]
			queueLock.Unlock()
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(item.Name)
			row.Objects[2].(*widget.Label).SetText(strings.ToUpper(item.Status))
		},
	)

	updateTaskStatus := func(id string, status string) {
		queueLock.Lock()
		for i := range queue {
			if queue[i].ID == id {
				queue[i].Status = status
				break
			}
		}
		queueLock.Unlock()
		queueList.Refresh()
	}

	runCfg := func(current core.Config, taskID string) bool {
		logs.SetText("")
		progress.SetValue(0)
		statusLabel.SetText("RUNNING")
		
		var localLogs []string

		_, err := downloader.Run(current, func(ev downloader.Event) {
			switch ev.Type {
			case downloader.EventLog:
				line := ev.LogLine
				if !strings.HasPrefix(strings.TrimSpace(line), "[download]") || !strings.Contains(line, "%") {
					localLogs = append(localLogs, line)
					if len(localLogs) > 450 {
						localLogs = localLogs[len(localLogs)-450:]
					}
					logs.SetText(strings.Join(localLogs, "\n"))
					logs.CursorRow = len(localLogs) - 1
				}
			case downloader.EventProgress:
				progress.SetValue(ev.Progress)
			}
		})
		if err != nil {
			statusLabel.SetText("ERROR")
			localLogs = append(localLogs, "ERROR: "+err.Error())
			logs.SetText(strings.Join(localLogs, "\n"))
			addIssue("Download failed", err)
			if taskID != "" {
				updateTaskStatus(taskID, "error")
			}
			return false
		}
		statusLabel.SetText("COMPLETED")
		if taskID != "" {
			updateTaskStatus(taskID, "completed")
		}
		return true
	}

	runBtn := widget.NewButton("Run download", func() {
		if !running.CompareAndSwap(false, true) {
			return
		}
		go func(localCfg core.Config) {
			defer running.Store(false)
			runCfg(localCfg, "")
		}(cfg)
	})

	addQueueBtn := widget.NewButton("Add current", func() {
		if strings.TrimSpace(cfg.URL) == "" {
			addIssue("Queue add rejected", fmt.Errorf("set URL before adding task"))
			return
		}
		queueLock.Lock()
		queue = append(queue, QueueTask{
			ID:     fmt.Sprintf("q-%d", time.Now().UnixNano()),
			Name:   cfg.URL,
			Config: cfg,
			Status: "queued",
		})
		queueLock.Unlock()
		queueList.Refresh()
	})

	runQueueBtn := widget.NewButton("Run queue", func() {
		if !running.CompareAndSwap(false, true) {
			return
		}
		go func() {
			defer running.Store(false)
			queueLock.Lock()
			local := make([]QueueTask, 0, len(queue))
			for _, t := range queue {
				if t.Status == "queued" {
					local = append(local, t)
				}
			}
			queueLock.Unlock()

			for _, task := range local {
				updateTaskStatus(task.ID, "running")
				runCfg(task.Config, task.ID)
			}

			if statusLabel.Text != "ERROR" {
				statusLabel.SetText("READY")
			}
		}()
	})

	clearQueueBtn := widget.NewButton("Clear queue", func() {
		if running.Load() {
			return
		}
		queueLock.Lock()
		queue = nil
		queueLock.Unlock()
		queueList.Refresh()
	})

	playlistPresetBtn := widget.NewButton("Preset: YouTube Playlist", func() {
		core.ApplyYouTubePlaylistPreset(&cfg)
		qualitySelect.SetSelected(cfg.Quality)
		formatSelect.SetSelected(cfg.Format)
		audioCheck.SetChecked(cfg.AudioOnly)
		reverseCheck.SetChecked(cfg.PlaylistReverse)
		cookiesBrowserCheck.SetChecked(cfg.UseCookiesBrowser)
		cookiesBrowserSelect.SetSelected(cfg.CookiesBrowser)
		updatePreview()
	})

	audioPresetBtn := widget.NewButton("Preset: Audio Only", func() {
		core.ApplyAudioOnlyPreset(&cfg)
		audioCheck.SetChecked(cfg.AudioOnly)
		formatSelect.SetSelected(cfg.Format)
		templateEntry.SetText(cfg.OutputTemplate)
		updatePreview()
	})

	mainForm := widget.NewForm(
		widget.NewFormItem("URL", urlEntry),
		widget.NewFormItem("Quality (-f)", qualitySelect),
		widget.NewFormItem("Format", formatSelect),
		widget.NewFormItem("", audioCheck),
		widget.NewFormItem("Save path (-P)", container.NewBorder(nil, nil, nil, pickFolderBtn, pathEntry)),
	)

	templateHint := widget.NewLabel("Template tips: %(title)s %(upload_date)s %(uploader)s %(ext)s")
	templateHint.Wrapping = fyne.TextWrapWord

	advancedForm := widget.NewForm(
		widget.NewFormItem("Output template (-o)", templateEntry),
		widget.NewFormItem("", templateHint),
		widget.NewFormItem("", cookiesBrowserCheck),
		widget.NewFormItem("Browser", cookiesBrowserSelect),
		widget.NewFormItem("", cookiesFileCheck),
		widget.NewFormItem("Cookies file", cookiesFileEntry),
		widget.NewFormItem("Proxy (--proxy)", proxyEntry),
		widget.NewFormItem("Rate (--limit-rate)", rateEntry),
		widget.NewFormItem("", reverseCheck),
	)

	queuePanel := container.NewVBox(
		widget.NewLabel("[ QUEUE ]"),
		container.NewHBox(addQueueBtn, runQueueBtn, clearQueueBtn),
		container.NewVScroll(queueList),
	)

	leftPanel := container.NewVBox(
		widget.NewLabel("[ BASIC ]"),
		mainForm,
		widget.NewLabel("[ ADVANCED ]"),
		advancedForm,
		container.NewGridWithColumns(2, playlistPresetBtn, audioPresetBtn),
		queuePanel,
	)

	previewPanel := container.NewVBox(
		container.NewHBox(widget.NewLabel("[ PREVIEW ]"), layout.NewSpacer(), issuesBtn, runBtn, statusLabel),
		commandPreview,
		progress,
		widget.NewLabel("[ LOG ]"),
		logs,
	)

	content := container.NewHSplit(
		container.NewVScroll(leftPanel),
		container.NewPadded(previewPanel),
	)
	content.SetOffset(0.48)

	w.SetContent(content)
	w.ShowAndRun()
}