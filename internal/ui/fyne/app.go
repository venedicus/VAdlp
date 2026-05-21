package fyneui

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/core"
	"vadlp/internal/downloader"
	"vadlp/internal/i18n"
	"vadlp/internal/settings"
	"vadlp/internal/updater"
)

type QueueTask struct {
	ID     string
	Name   string
	Config core.Config
	Status string
}

func Run() {
	a := app.NewWithID("vadlp.fyne")
	a.Settings().SetTheme(NewTokyoNightTheme())

	appSettings, _ := settings.Load()
	lang := appSettings.Language
	if lang == "" {
		lang = "en"
	}
	_ = i18n.Init(lang)
	tr := func(id string) string { return i18n.T(id, nil) }

	w := a.NewWindow(tr("app.title"))
	w.Resize(fyne.NewSize(1240, 840))

	cfg := appSettings.Config
	if appSettings.DenoPath != "" {
		cfg.DenoPath = appSettings.DenoPath
	}
	var journalEntries []string

	var snapMu sync.Mutex
	lastSession := core.Session{Config: cfg}

	statusBadge := NewStatusBadge("ready")
	statusBadge.SetStatusKey("ready")
	phaseBadge := NewPhaseBadge()

	journalBtn := widget.NewButton(tr("journal.title")+" (0)", nil)

	commandPreview := widget.NewMultiLineEntry()
	commandPreview.Disable()
	commandPreview.Wrapping = fyne.TextWrapWord
	commandPreview.SetMinRowsVisible(4)
	commandPreview.TextStyle = fyne.TextStyle{Monospace: true}

	progressFile := widget.NewProgressBar()
	progressFile.Min = 0
	progressFile.Max = 100
	progressOverall := widget.NewProgressBar()
	progressOverall.Min = 0
	progressOverall.Max = 100
	progressFileLabel := widget.NewLabel(tr("progress.file"))
	progressOverallLabel := widget.NewLabel(tr("progress.overall"))

	logs := widget.NewMultiLineEntry()
	logs.Disable()
	logs.Wrapping = fyne.TextWrapWord
	logs.TextStyle = fyne.TextStyle{Monospace: true}

	sessionPathEntry := widget.NewEntry()
	sessionPathEntry.SetPlaceHolder(tr("placeholder.session_path"))
	if appSettings.SessionPath != "" {
		sessionPathEntry.SetText(appSettings.SessionPath)
	}

	saveAppSettings := func() {
		appSettings.Config = cfg
		appSettings.SessionPath = strings.TrimSpace(sessionPathEntry.Text)
		size := w.Canvas().Size()
		appSettings.WindowWidth = size.Width
		appSettings.WindowHeight = size.Height
		_ = settings.Save(appSettings)
	}

	var setQuality func(string)

	addJournal := func(summary string, err error) {
		entry := summary
		if err != nil {
			entry = summary + ": " + err.Error()
		}
		timestamp := time.Now().Format("15:04:05")
		journalEntries = append(journalEntries, "["+timestamp+"] "+entry)
		journalBtn.SetText(fmt.Sprintf("%s (%d)", tr("journal.title"), len(journalEntries)))
	}

	journalBtn.OnTapped = func() {
		content := tr("journal.empty")
		if len(journalEntries) > 0 {
			content = strings.Join(journalEntries, "\n")
		}
		issueView := widget.NewMultiLineEntry()
		issueView.SetText(content)
		issueView.Disable()
		issueView.Wrapping = fyne.TextWrapWord
		issueView.TextStyle = fyne.TextStyle{Monospace: true}

		scroll := container.NewVScroll(issueView)
		scroll.SetMinSize(fyne.NewSize(720, 380))

		d := dialog.NewCustom(tr("journal.title"), tr("btn.close"), scroll, w)
		d.Resize(fyne.NewSize(760, 460))
		d.Show()
	}

	updatePreview := func() {
		prog := "yt-dlp"
		if p, err := downloader.ResolveBinary(); err == nil {
			prog = p
		}
		commandPreview.SetText(core.PreviewCommand(cfg, prog))
	}

	urlEntry := newDropEntry()
	urlEntry.OnChanged = func(s string) {
		cfg.URL = s
		updatePreview()
	}
	setupURLDrop(w, func(s string) {
		urlEntry.SetText(s)
		cfg.URL = s
		updatePreview()
	})

	formatUI, formatCard := NewDownloadFormatUI(&cfg, updatePreview, tr)

	setQuality = func(q string) {
		cfg.Quality = q
		formatUI.QualityEntry.SetText(q)
		updatePreview()
	}

	pathEntry := widget.NewEntry()
	pathEntry.SetText(cfg.OutputPath)
	pathEntry.OnChanged = func(s string) {
		cfg.OutputPath = s
		updatePreview()
	}
	pickFolderBtn := widget.NewButton(tr("btn.browse"), func() {
		dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				addJournal(tr("err.folder_picker"), err)
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

	cookiesBrowserCheck := widget.NewCheck(tr("check.cookies_browser"), func(b bool) {
		cfg.UseCookiesBrowser = b
		updatePreview()
	})
	cookiesBrowserCheck.SetChecked(cfg.UseCookiesBrowser)

	cookiesBrowserSelect := widget.NewSelect([]string{"chrome", "firefox", "vivaldi", "edge", "brave"}, func(s string) {
		cfg.CookiesBrowser = s
		updatePreview()
	})
	cookiesBrowserSelect.SetSelected(cfg.CookiesBrowser)

	cookiesFileCheck := widget.NewCheck(tr("check.cookies_file"), func(b bool) {
		cfg.UseCookiesFile = b
		updatePreview()
	})

	cookiesFileEntry := widget.NewEntry()
	cookiesFileEntry.SetPlaceHolder(tr("placeholder.cookies"))
	cookiesFileEntry.OnChanged = func(s string) {
		cfg.CookiesFile = s
		updatePreview()
	}
	pickCookiesBtn := widget.NewButton(tr("btn.browse"), func() {
		dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			path := r.URI().Path()
			_ = r.Close()
			cookiesFileEntry.SetText(path)
			cfg.CookiesFile = path
			cookiesFileCheck.SetChecked(true)
			cfg.UseCookiesFile = true
			updatePreview()
		}, w).Show()
	})

	proxyEntry := widget.NewEntry()
	proxyEntry.SetPlaceHolder(tr("placeholder.proxy"))
	proxyEntry.OnChanged = func(s string) {
		cfg.Proxy = s
		updatePreview()
	}

	rateEntry := widget.NewEntry()
	rateEntry.SetPlaceHolder(tr("placeholder.rate"))
	rateEntry.OnChanged = func(s string) {
		cfg.RateLimit = s
		updatePreview()
	}

	reverseCheck := widget.NewCheck(tr("check.reverse"), func(b bool) {
		cfg.PlaylistReverse = b
		updatePreview()
	})
	reverseCheck.SetChecked(cfg.PlaylistReverse)

	continueCheck := widget.NewCheck(tr("check.continue"), func(b bool) {
		cfg.Continue = b
		updatePreview()
	})
	continueCheck.SetChecked(cfg.Continue)

	noPartCheck := widget.NewCheck(tr("check.no_part"), func(b bool) {
		cfg.NoPart = b
		updatePreview()
	})

	playlistStartEntry := widget.NewEntry()
	playlistStartEntry.SetPlaceHolder(tr("placeholder.playlist_start"))
	playlistStartEntry.OnChanged = func(s string) {
		cfg.PlaylistStart = atoiOrZero(s)
		updatePreview()
	}

	playlistEndEntry := widget.NewEntry()
	playlistEndEntry.SetPlaceHolder(tr("placeholder.playlist_end"))
	playlistEndEntry.OnChanged = func(s string) {
		cfg.PlaylistEnd = atoiOrZero(s)
		updatePreview()
	}

	maxDownloadsEntry := widget.NewEntry()
	maxDownloadsEntry.SetPlaceHolder(tr("placeholder.max_downloads"))
	maxDownloadsEntry.OnChanged = func(s string) {
		cfg.MaxDownloads = atoiOrZero(s)
		updatePreview()
	}

	downloadArchiveEntry := widget.NewEntry()
	downloadArchiveEntry.SetPlaceHolder(tr("placeholder.archive"))
	downloadArchiveEntry.OnChanged = func(s string) {
		cfg.DownloadArchive = s
		updatePreview()
	}

	noPlaylistCheck := widget.NewCheck(tr("check.no_playlist"), func(b bool) {
		cfg.NoPlaylist = b
		updatePreview()
	})

	flatPlaylistCheck := widget.NewCheck(tr("check.flat_playlist"), func(b bool) {
		cfg.FlatPlaylist = b
		updatePreview()
	})

	writeSubsCheck := widget.NewCheck(tr("check.write_subs"), func(b bool) {
		cfg.WriteSubs = b
		updatePreview()
	})

	writeAutoSubCheck := widget.NewCheck(tr("check.write_auto_sub"), func(b bool) {
		cfg.WriteAutoSub = b
		updatePreview()
	})

	embedSubsCheck := widget.NewCheck(tr("check.embed_subs"), func(b bool) {
		cfg.EmbedSubs = b
		updatePreview()
	})

	subLangsEntry := widget.NewEntry()
	subLangsEntry.SetText(cfg.SubLangs)
	subLangsEntry.OnChanged = func(s string) {
		cfg.SubLangs = s
		updatePreview()
	}

	writeThumbCheck := widget.NewCheck(tr("check.write_thumb"), func(b bool) {
		cfg.WriteThumbnail = b
		updatePreview()
	})

	embedThumbCheck := widget.NewCheck(tr("check.embed_thumb"), func(b bool) {
		cfg.EmbedThumbnail = b
		updatePreview()
	})

	embedMetaCheck := widget.NewCheck(tr("check.embed_meta"), func(b bool) {
		cfg.EmbedMetadata = b
		updatePreview()
	})

	embedChaptersCheck := widget.NewCheck(tr("check.embed_chapters"), func(b bool) {
		cfg.EmbedChapters = b
		updatePreview()
	})

	writeInfoJSONCheck := widget.NewCheck(tr("check.write_info_json"), func(b bool) {
		cfg.WriteInfoJSON = b
		updatePreview()
	})

	loadInfoJSONEntry := widget.NewEntry()
	loadInfoJSONEntry.SetPlaceHolder(tr("placeholder.load_info_json"))
	loadInfoJSONEntry.OnChanged = func(s string) {
		cfg.LoadInfoJSON = s
		updatePreview()
	}

	retriesEntry := widget.NewEntry()
	retriesEntry.SetText(strconv.Itoa(cfg.Retries))
	retriesEntry.OnChanged = func(s string) {
		cfg.Retries = atoiOrZero(s)
		updatePreview()
	}

	fragRetriesEntry := widget.NewEntry()
	fragRetriesEntry.SetText(strconv.Itoa(cfg.FragmentRetries))
	fragRetriesEntry.OnChanged = func(s string) {
		cfg.FragmentRetries = atoiOrZero(s)
		updatePreview()
	}

	concFragEntry := widget.NewEntry()
	concFragEntry.SetText(strconv.Itoa(cfg.ConcurrentFragments))
	concFragEntry.OnChanged = func(s string) {
		cfg.ConcurrentFragments = atoiOrZero(s)
		if cfg.ConcurrentFragments < 0 {
			cfg.ConcurrentFragments = 0
		}
		updatePreview()
	}

	socketTimeoutEntry := widget.NewEntry()
	socketTimeoutEntry.SetPlaceHolder(tr("placeholder.socket_timeout"))
	socketTimeoutEntry.OnChanged = func(s string) {
		cfg.SocketTimeout = atoiOrZero(s)
		updatePreview()
	}

	noWarningsCheck := widget.NewCheck(tr("check.no_warnings"), func(b bool) {
		cfg.NoWarnings = b
		updatePreview()
	})

	verboseCheck := widget.NewCheck(tr("check.verbose"), func(b bool) {
		cfg.Verbose = b
		updatePreview()
	})

	quietCheck := widget.NewCheck(tr("check.quiet"), func(b bool) {
		cfg.Quiet = b
		updatePreview()
	})

	windowsFilenamesCheck := widget.NewCheck(tr("check.windows_filenames"), func(b bool) {
		cfg.WindowsFilenames = b
		updatePreview()
	})

	noMtimeCheck := widget.NewCheck(tr("check.no_mtime"), func(b bool) {
		cfg.NoMtime = b
		updatePreview()
	})

	abortOnErrorCheck := widget.NewCheck(tr("check.abort_on_error"), func(b bool) {
		cfg.AbortOnError = b
		updatePreview()
	})

	ignoreErrorsCheck := widget.NewCheck(tr("check.ignore_errors"), func(b bool) {
		cfg.IgnoreErrors = b
		updatePreview()
	})

	extraArgsEntry := widget.NewMultiLineEntry()
	extraArgsEntry.SetPlaceHolder(tr("placeholder.extra_args"))
	extraArgsEntry.SetMinRowsVisible(4)
	extraArgsEntry.OnChanged = func(s string) {
		cfg.ExtraArgs = s
		updatePreview()
	}

	batchURLEntry := widget.NewMultiLineEntry()
	batchURLEntry.SetPlaceHolder(tr("placeholder.batch_urls"))
	batchURLEntry.SetMinRowsVisible(3)
	batchURLEntry.SetText(cfg.BatchURLs)
	batchURLEntry.OnChanged = func(s string) {
		cfg.BatchURLs = s
		updatePreview()
	}

	sponsorBlockCheck := widget.NewCheck(tr("check.sponsorblock"), func(b bool) {
		cfg.SponsorBlockRemove = b
		updatePreview()
	})
	sponsorBlockCheck.SetChecked(cfg.SponsorBlockRemove)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(cfg.Username)
	usernameEntry.OnChanged = func(s string) {
		cfg.Username = s
		updatePreview()
	}

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetText(cfg.Password)
	passwordEntry.OnChanged = func(s string) {
		cfg.Password = s
		updatePreview()
	}

	ffmpegPathEntry := widget.NewEntry()
	ffmpegPathEntry.SetText(cfg.FFmpegLocation)
	if appSettings.FFmpegPath != "" && cfg.FFmpegLocation == "" {
		ffmpegPathEntry.SetText(appSettings.FFmpegPath)
		cfg.FFmpegLocation = appSettings.FFmpegPath
	}
	ffmpegPathEntry.OnChanged = func(s string) {
		cfg.FFmpegLocation = s
		appSettings.FFmpegPath = s
		updatePreview()
	}

	denoPathEntry := widget.NewEntry()
	denoPathEntry.SetText(cfg.DenoPath)
	if appSettings.DenoPath != "" && cfg.DenoPath == "" {
		denoPathEntry.SetText(appSettings.DenoPath)
		cfg.DenoPath = appSettings.DenoPath
	}
	denoPathEntry.OnChanged = func(s string) {
		cfg.DenoPath = s
		appSettings.DenoPath = s
		updatePreview()
	}

	parallelEntry := widget.NewEntry()
	if appSettings.QueueParallel < 1 {
		appSettings.QueueParallel = 1
	}
	parallelEntry.SetText(strconv.Itoa(appSettings.QueueParallel))
	parallelEntry.OnChanged = func(s string) {
		appSettings.QueueParallel = atoiOrZero(s)
		if appSettings.QueueParallel < 1 {
			appSettings.QueueParallel = 1
		}
	}

	syncUIFromCfg := func(c core.Config) {
		cfg = c
		urlEntry.SetText(cfg.URL)
		formatUI.SyncFromCfg(c)
		pathEntry.SetText(cfg.OutputPath)
		templateEntry.SetText(cfg.OutputTemplate)
		cookiesBrowserCheck.SetChecked(cfg.UseCookiesBrowser)
		cookiesBrowserSelect.SetSelected(cfg.CookiesBrowser)
		cookiesFileCheck.SetChecked(cfg.UseCookiesFile)
		cookiesFileEntry.SetText(cfg.CookiesFile)
		proxyEntry.SetText(cfg.Proxy)
		rateEntry.SetText(cfg.RateLimit)
		reverseCheck.SetChecked(cfg.PlaylistReverse)
		continueCheck.SetChecked(cfg.Continue)
		noPartCheck.SetChecked(cfg.NoPart)
		playlistStartEntry.SetText(itoaOrEmpty(cfg.PlaylistStart))
		playlistEndEntry.SetText(itoaOrEmpty(cfg.PlaylistEnd))
		maxDownloadsEntry.SetText(itoaOrEmpty(cfg.MaxDownloads))
		downloadArchiveEntry.SetText(cfg.DownloadArchive)
		noPlaylistCheck.SetChecked(cfg.NoPlaylist)
		flatPlaylistCheck.SetChecked(cfg.FlatPlaylist)
		writeSubsCheck.SetChecked(cfg.WriteSubs)
		writeAutoSubCheck.SetChecked(cfg.WriteAutoSub)
		embedSubsCheck.SetChecked(cfg.EmbedSubs)
		subLangsEntry.SetText(cfg.SubLangs)
		writeThumbCheck.SetChecked(cfg.WriteThumbnail)
		embedThumbCheck.SetChecked(cfg.EmbedThumbnail)
		embedMetaCheck.SetChecked(cfg.EmbedMetadata)
		embedChaptersCheck.SetChecked(cfg.EmbedChapters)
		writeInfoJSONCheck.SetChecked(cfg.WriteInfoJSON)
		loadInfoJSONEntry.SetText(cfg.LoadInfoJSON)
		retriesEntry.SetText(strconv.Itoa(cfg.Retries))
		fragRetriesEntry.SetText(strconv.Itoa(cfg.FragmentRetries))
		concFragEntry.SetText(strconv.Itoa(cfg.ConcurrentFragments))
		socketTimeoutEntry.SetText(itoaOrEmpty(cfg.SocketTimeout))
		noWarningsCheck.SetChecked(cfg.NoWarnings)
		verboseCheck.SetChecked(cfg.Verbose)
		quietCheck.SetChecked(cfg.Quiet)
		windowsFilenamesCheck.SetChecked(cfg.WindowsFilenames)
		noMtimeCheck.SetChecked(cfg.NoMtime)
		abortOnErrorCheck.SetChecked(cfg.AbortOnError)
		ignoreErrorsCheck.SetChecked(cfg.IgnoreErrors)
		extraArgsEntry.SetText(cfg.ExtraArgs)
		batchURLEntry.SetText(cfg.BatchURLs)
		sponsorBlockCheck.SetChecked(cfg.SponsorBlockRemove)
		usernameEntry.SetText(cfg.Username)
		passwordEntry.SetText(cfg.Password)
		ffmpegPathEntry.SetText(cfg.FFmpegLocation)
		denoPathEntry.SetText(cfg.DenoPath)
		updatePreview()
	}

	saveSessionBtn := widget.NewButton(tr("btn.save_session"), func() {
		dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err != nil {
				addJournal(tr("err.save_session"), err)
				return
			}
			if uc == nil {
				return
			}
			path := uc.URI().Path()
			_ = uc.Close()
			snapMu.Lock()
			s := lastSession
			s.Config = cfg
			snapMu.Unlock()
			if err := core.SaveSession(path, s); err != nil {
				addJournal(tr("err.save_session"), err)
				return
			}
			sessionPathEntry.SetText(path)
		}, w).Show()
	})

	loadSessionBtn := widget.NewButton(tr("btn.load_session"), func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				addJournal(tr("err.load_session"), err)
				return
			}
			if reader == nil {
				return
			}
			path := reader.URI().Path()
			_ = reader.Close()
			s, err := core.LoadSession(path)
			if err != nil {
				addJournal(tr("err.load_session"), err)
				return
			}
			resume := s.ApplyResumeHints()
			syncUIFromCfg(resume)
			snapMu.Lock()
			lastSession = s
			lastSession.Config = cfg
			snapMu.Unlock()
			sessionPathEntry.SetText(path)
			statusBadge.SetStatusKey("ready")
			phaseBadge.SetPhase(downloader.StageUnknown)
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		fd.Show()
	})

	applyResumeBtn := widget.NewButton(tr("btn.resume_session"), func() {
		snapMu.Lock()
		s := lastSession
		snapMu.Unlock()
		if s.Version == 0 {
			addJournal(tr("err.resume_session"), fmt.Errorf("%s", tr("err.resume_session")))
			return
		}
		syncUIFromCfg(s.ApplyResumeHints())
	})

	updatePreview()

	var (
		queue            []QueueTask
		queueLock        sync.Mutex
		running          atomic.Bool
		selectedQueueIdx = -1
		currentJobID     atomic.Value
	)

	queueList := widget.NewList(
		func() int {
			queueLock.Lock()
			defer queueLock.Unlock()
			return len(queue)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(tr("queue.task")),
				layout.NewSpacer(),
				widget.NewLabel(tr("status.queued")),
				widget.NewButton("×", nil),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			queueLock.Lock()
			item := queue[i]
			queueLock.Unlock()
			row := o.(*fyne.Container)
			row.Objects[0].(*widget.Label).SetText(item.Name)
			row.Objects[2].(*widget.Label).SetText(localizedStatus(item.Status))
			btn := row.Objects[3].(*widget.Button)
			taskID := item.ID
			if item.Status == "running" {
				btn.SetText(tr("btn.cancel_task"))
				btn.Show()
				btn.Enable()
				btn.OnTapped = func() { downloader.CancelJob(taskID) }
			} else {
				btn.Hide()
			}
		},
	)
	queueList.OnSelected = func(id widget.ListItemID) {
		selectedQueueIdx = int(id)
	}

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

	var persistMu sync.Mutex
	lastPersistWrite := time.Time{}
	persistSnapshot := func(path string, snap core.Session) {
		if strings.TrimSpace(path) == "" {
			return
		}
		persistMu.Lock()
		defer persistMu.Unlock()
		if time.Since(lastPersistWrite) < 2*time.Second {
			return
		}
		lastPersistWrite = time.Now()
		_ = core.SaveSession(path, snap)
	}

	stopBtn := widget.NewButton(tr("btn.stop"), func() {
		if v := currentJobID.Load(); v != nil {
			if id, ok := v.(string); ok && id != "" {
				downloader.CancelJob(id)
			}
		}
	})
	stopBtn.Disable()

	runBtn := widget.NewButton(tr("btn.download"), nil)
	runBtn.Importance = widget.HighImportance

	openFolderBtn := widget.NewButton(tr("btn.open_folder"), func() {
		if err := openFolder(cfg.OutputPath); err != nil {
			addJournal(tr("err.open_folder"), err)
		}
	})

	runDownload := func(current core.Config, taskID string, qIdx, qTot int) bool {
		jobID := taskID
		if jobID == "" {
			jobID = "main"
		}
		currentJobID.Store(jobID)
		defer currentJobID.Store("")

		uiExec(func() {
			logs.SetText("")
			progressFile.SetValue(0)
			progressOverall.SetValue(0)
			statusBadge.SetStatusKey("running")
			phaseBadge.SetPhase(downloader.StageUnknown)
			progressFileLabel.SetText("This file")
			if qTot > 1 {
				progressOverallLabel.SetText(fmt.Sprintf("Queue: task %d of %d", qIdx, qTot))
			} else {
				progressOverallLabel.SetText("Everything in this run")
			}
			stopBtn.Enable()
			runBtn.Disable()
		})
		defer uiExec(func() {
			stopBtn.Disable()
			runBtn.Enable()
		})

		plCur, plTot := 0, 0
		filePct := 0.0

		updateSnap := func() {
			snap := core.Session{
				Config:          current,
				PlaylistCurrent: plCur,
				PlaylistTotal:   plTot,
				QueueDone:       maxInt(0, qIdx-1),
				QueueTotal:      qTot,
			}
			snapMu.Lock()
			lastSession = snap
			snapMu.Unlock()
			persistSnapshot(strings.TrimSpace(sessionPathEntry.Text), snap)
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
			v := computeOverall()
			uiExec(func() {
				progressOverall.SetValue(v)
				if plTot > 0 && plCur > 0 {
					progressOverallLabel.SetText(fmt.Sprintf("Playlist: item %d of %d", plCur, plTot))
				} else if qTot > 1 {
					progressOverallLabel.SetText(fmt.Sprintf("Queue: task %d of %d", qIdx, qTot))
				}
			})
		}

		var localLogs []string

		_, err := downloader.Run(current, jobID, func(ev downloader.Event) {
			switch ev.Type {
			case downloader.EventLog:
				line := ev.LogLine
				if ev.Stage != downloader.StageUnknown {
					st := ev.Stage
					uiExec(func() { phaseBadge.SetPhase(st) })
				}
				if !strings.HasPrefix(strings.TrimSpace(line), "[download]") || !strings.Contains(line, "%") {
					localLogs = append(localLogs, line)
					if len(localLogs) > 450 {
						localLogs = localLogs[len(localLogs)-450:]
					}
					uiExec(func() {
						logs.SetText(strings.Join(localLogs, "\n"))
						logs.CursorRow = len(localLogs) - 1
					})
				}
			case downloader.EventProgress:
				filePct = ev.Progress
				uiExec(func() {
					progressFile.SetValue(ev.Progress)
					if ev.Speed != "" || ev.ETA != "" {
						progressFileLabel.SetText(strings.TrimSpace(ev.Speed + "  ETA " + ev.ETA))
					}
				})
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
			if downloader.IsCancelled(err) {
				st = "cancelled"
				msg = "cancelled"
			}
			uiExec(func() {
				if downloader.IsCancelled(err) {
					statusBadge.SetStatusKey("stopped")
				} else {
					statusBadge.SetStatusKey("error")
				}
				phaseBadge.SetPhase(downloader.StageUnknown)
				localLogs = append(localLogs, strings.ToUpper(msg))
				logs.SetText(strings.Join(localLogs, "\n"))
			})
			if !downloader.IsCancelled(err) {
				addJournal(tr("err.download_failed"), err)
			}
			_ = core.AppendHistory(core.HistoryItem{
				URL:    firstNonEmpty(current.URL, current.BatchURLs),
				Status: st,
				Output: current.OutputPath,
				Error:  msg,
			})
			if taskID != "" {
				updateTaskStatus(taskID, st)
			}
			updateSnap()
			return !downloader.IsCancelled(err)
		}

		uiExec(func() {
			statusBadge.SetStatusKey("completed")
			phaseBadge.SetPhase(downloader.StageUnknown)
			progressFile.SetValue(100)
			progressOverall.SetValue(100)
		})
		_ = core.AppendHistory(core.HistoryItem{
			URL:    firstNonEmpty(current.URL, current.BatchURLs),
			Status: "completed",
			Output: current.OutputPath,
		})
		if taskID != "" {
			updateTaskStatus(taskID, "completed")
		}
		plCur, plTot = 0, 0
		filePct = 100
		updateSnap()
		return true
	}

	runBtn.OnTapped = func() {
		if !running.CompareAndSwap(false, true) {
			return
		}
		go func(localCfg core.Config) {
			defer running.Store(false)
			runDownload(localCfg, "", 1, 1)
			saveAppSettings()
		}(cfg)
	}

	addQueueBtn := widget.NewButton(tr("btn.add_queue"), func() {
		if strings.TrimSpace(cfg.URL) == "" && strings.TrimSpace(cfg.LoadInfoJSON) == "" && strings.TrimSpace(cfg.BatchURLs) == "" {
			addJournal(tr("err.queue_no_url"), fmt.Errorf("%s", tr("err.queue_no_url")))
			return
		}
		queueLock.Lock()
		queue = append(queue, QueueTask{
			ID:     fmt.Sprintf("q-%d", time.Now().UnixNano()),
			Name:   firstNonEmpty(cfg.URL, cfg.LoadInfoJSON),
			Config: cfg,
			Status: "queued",
		})
		queueLock.Unlock()
		queueList.Refresh()
	})

	runQueueBtn := widget.NewButton(tr("btn.run_queue"), func() {
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

			qTot := len(local)
			workers := appSettings.QueueParallel
			if workers < 1 {
				workers = 1
			}

			runOne := func(i int, task QueueTask) bool {
				updateTaskStatus(task.ID, "running")
				return runDownload(task.Config, task.ID, i+1, qTot)
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
					statusBadge.SetStatusKey("ready")
					phaseBadge.SetPhase(downloader.StageUnknown)
				}
			})
			saveAppSettings()
		}()
	})

	removeQueueBtn := widget.NewButton(tr("btn.remove"), func() {
		if running.Load() || selectedQueueIdx < 0 {
			return
		}
		queueLock.Lock()
		if selectedQueueIdx < len(queue) {
			queue = append(queue[:selectedQueueIdx], queue[selectedQueueIdx+1:]...)
			selectedQueueIdx = -1
		}
		queueLock.Unlock()
		queueList.Refresh()
	})

	retryQueueBtn := widget.NewButton(tr("btn.retry_failed"), func() {
		queueLock.Lock()
		for i := range queue {
			if queue[i].Status == "error" || queue[i].Status == "cancelled" {
				queue[i].Status = "queued"
			}
		}
		queueLock.Unlock()
		queueList.Refresh()
	})

	moveQueueUpBtn := widget.NewButton(tr("btn.up"), func() {
		if running.Load() || selectedQueueIdx <= 0 {
			return
		}
		queueLock.Lock()
		i := selectedQueueIdx
		if i < len(queue) {
			queue[i-1], queue[i] = queue[i], queue[i-1]
			selectedQueueIdx--
		}
		queueLock.Unlock()
		queueList.Refresh()
	})

	moveQueueDownBtn := widget.NewButton(tr("btn.down"), func() {
		queueLock.Lock()
		i := selectedQueueIdx
		if !running.Load() && i >= 0 && i < len(queue)-1 {
			queue[i+1], queue[i] = queue[i], queue[i+1]
			selectedQueueIdx++
		}
		queueLock.Unlock()
		queueList.Refresh()
	})

	clearQueueBtn := widget.NewButton(tr("btn.clear_queue"), func() {
		if running.Load() {
			return
		}
		queueLock.Lock()
		queue = nil
		queueLock.Unlock()
		queueList.Refresh()
	})

	fetchFormatsBtn := widget.NewButton(tr("btn.fetch_formats"), func() {
		showFormatPicker(w, cfg, setQuality)
	})
	pasteURLBtn := widget.NewButton(tr("btn.paste"), func() {
		if text := w.Clipboard().Content(); text != "" {
			urlEntry.SetText(text)
		}
	})
	urlActions := container.NewHBox(pasteURLBtn, fetchFormatsBtn)
	urlRow := container.NewBorder(nil, nil, nil, urlActions, urlEntry)
	linkCard := widget.NewCard(tr("card.url"), tr("drop.hint"), urlRow)

	batchCard := widget.NewCard(tr("card.batch"), "", batchURLEntry)

	folderRow := container.NewBorder(nil, nil, nil, pickFolderBtn, pathEntry)
	outputCard := widget.NewCard(tr("card.output"), "",
		container.NewVBox(
			folderRow,
			widget.NewSeparator(),
			widget.NewForm(widget.NewFormItem(tr("form.filename_template"), templateEntry)),
		),
	)

	_, profileCard := NewProfileBar(w, &cfg, &appSettings, saveAppSettings, syncUIFromCfg, tr)

	downloadTab := container.NewVBox(
		profileCard,
		widget.NewSeparator(),
		linkCard,
		widget.NewSeparator(),
		batchCard,
		widget.NewSeparator(),
		outputCard,
		widget.NewSeparator(),
		formatCard,
	)

	cookiesFileRow := container.NewBorder(nil, nil, nil, pickCookiesBtn, cookiesFileEntry)
	networkForm := widget.NewForm(
		widget.NewFormItem("", cookiesBrowserCheck),
		widget.NewFormItem(tr("form.browser"), cookiesBrowserSelect),
		widget.NewFormItem("", cookiesFileCheck),
		widget.NewFormItem(tr("form.cookies_file"), cookiesFileRow),
		widget.NewFormItem(tr("form.proxy"), proxyEntry),
		widget.NewFormItem(tr("form.rate_limit"), rateEntry),
		widget.NewFormItem(tr("form.username"), usernameEntry),
		widget.NewFormItem(tr("form.password"), passwordEntry),
	)
	networkCard := widget.NewCard(tr("card.network"), "", networkForm)
	networkTab := container.NewVBox(networkCard)

	playlistForm := widget.NewForm(
		widget.NewFormItem("", reverseCheck),
		widget.NewFormItem("", continueCheck),
		widget.NewFormItem("", noPartCheck),
		widget.NewFormItem(tr("form.playlist_start"), playlistStartEntry),
		widget.NewFormItem(tr("form.playlist_end"), playlistEndEntry),
		widget.NewFormItem(tr("form.max_downloads"), maxDownloadsEntry),
		widget.NewFormItem(tr("form.archive"), downloadArchiveEntry),
		widget.NewFormItem("", noPlaylistCheck),
		widget.NewFormItem("", flatPlaylistCheck),
	)
	playlistCard := widget.NewCard(tr("card.playlist"), "", playlistForm)
	sessionCard := widget.NewCard(tr("card.session"), "",
		container.NewVBox(
			widget.NewForm(widget.NewFormItem(tr("form.session_path"), sessionPathEntry)),
			container.NewHBox(saveSessionBtn, loadSessionBtn, applyResumeBtn),
		),
	)
	playlistTab := container.NewVBox(playlistCard, widget.NewSeparator(), sessionCard)

	extrasMediaForm := widget.NewForm(
		widget.NewFormItem("", writeSubsCheck),
		widget.NewFormItem("", writeAutoSubCheck),
		widget.NewFormItem("", embedSubsCheck),
		widget.NewFormItem(tr("form.sub_langs"), subLangsEntry),
		widget.NewFormItem("", writeThumbCheck),
		widget.NewFormItem("", embedThumbCheck),
		widget.NewFormItem("", embedMetaCheck),
		widget.NewFormItem("", embedChaptersCheck),
		widget.NewFormItem("", writeInfoJSONCheck),
		widget.NewFormItem(tr("form.load_info_json"), loadInfoJSONEntry),
	)
	extrasMediaCard := widget.NewCard(tr("card.media_extras"), "", extrasMediaForm)

	extrasNetForm := widget.NewForm(
		widget.NewFormItem(tr("form.retries"), retriesEntry),
		widget.NewFormItem(tr("form.frag_retries"), fragRetriesEntry),
		widget.NewFormItem(tr("form.concurrent_frags"), concFragEntry),
		widget.NewFormItem(tr("form.socket_timeout"), socketTimeoutEntry),
		widget.NewFormItem("", noWarningsCheck),
		widget.NewFormItem("", verboseCheck),
		widget.NewFormItem("", quietCheck),
		widget.NewFormItem("", windowsFilenamesCheck),
		widget.NewFormItem("", noMtimeCheck),
		widget.NewFormItem("", abortOnErrorCheck),
		widget.NewFormItem("", ignoreErrorsCheck),
	)
	extrasNetCard := widget.NewCard(tr("card.retries"), "", extrasNetForm)

	extrasPowerCard := widget.NewCard(tr("card.extra_flags"), "", extraArgsEntry)
	extrasTab := container.NewVBox(
		extrasMediaCard,
		widget.NewSeparator(),
		extrasNetCard,
		widget.NewSeparator(),
		widget.NewCard(tr("card.sponsorblock"), "", sponsorBlockCheck),
		widget.NewSeparator(),
		extrasPowerCard,
	)

	updateYtDlpBtn := widget.NewButton(tr("btn.update_ytdlp"), func() {
		bin, err := downloader.ResolveBinary()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		go func() {
			out, err := updater.UpdateYtDlp(bin)
			uiExec(func() {
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				dialog.ShowInformation(tr("dialog.ytdlp_update"), out, w)
				updatePreview()
			})
		}()
	})

	langSelect := widget.NewSelect([]string{"en", "ru"}, func(code string) {
		appSettings.Language = code
		i18n.SetLanguage(code)
		saveAppSettings()
	})
	if appSettings.Language == "ru" {
		langSelect.SetSelected("ru")
	} else {
		langSelect.SetSelected("en")
	}

	ffmpegStatusLabel := widget.NewLabel("")
	ffmpegStatusLabel.Wrapping = fyne.TextWrapWord
	denoStatusLabel := widget.NewLabel("")
	denoStatusLabel.Wrapping = fyne.TextWrapWord

	installFFmpegBtn := widget.NewButton(tr("btn.install_ffmpeg"), nil)
	installDenoBtn := widget.NewButton(tr("btn.install_deno"), nil)

	syncToolsStatus := func() {
		r := updater.CheckTools()
		if r.FFmpeg.Found {
			ffmpegStatusLabel.SetText(i18n.T("tools.ffmpeg_ok", map[string]interface{}{
				"Path": r.FFmpeg.Path,
			}))
			installFFmpegBtn.Disable()
		} else {
			ffmpegStatusLabel.SetText(i18n.T("tools.ffmpeg_missing", nil))
			installFFmpegBtn.Enable()
		}
		if r.Deno.Found {
			denoStatusLabel.SetText(i18n.T("tools.deno_ok", map[string]interface{}{
				"Path": r.Deno.Path,
			}))
			installDenoBtn.Disable()
		} else {
			denoStatusLabel.SetText(i18n.T("tools.deno_missing", nil))
			installDenoBtn.Enable()
		}
	}

	installFFmpegBtn.OnTapped = func() {
		showFFmpegInstaller(w, addJournal, func(path string) {
			ffmpegPathEntry.SetText(path)
			cfg.FFmpegLocation = path
			appSettings.FFmpegPath = path
			updatePreview()
			syncToolsStatus()
		})
	}

	installDenoBtn.OnTapped = func() {
		showDenoInstaller(w, addJournal, func(path string) {
			denoPathEntry.SetText(path)
			cfg.DenoPath = path
			appSettings.DenoPath = path
			updatePreview()
			syncToolsStatus()
		})
	}

	syncToolsStatus()

	toolsForm := widget.NewForm(
		widget.NewFormItem(tr("form.language"), langSelect),
		widget.NewFormItem(tr("form.ffmpeg_path"), ffmpegPathEntry),
		widget.NewFormItem(tr("form.deno_path"), denoPathEntry),
		widget.NewFormItem(tr("form.queue_workers"), parallelEntry),
	)
	toolsTab := container.NewVBox(
		widget.NewCard(tr("card.binaries"), "", container.NewVBox(
			toolsForm,
			ffmpegStatusLabel,
			denoStatusLabel,
			container.NewHBox(updateYtDlpBtn, installFFmpegBtn, installDenoBtn),
		)),
	)

	historyView := widget.NewMultiLineEntry()
	historyView.Disable()
	historyView.Wrapping = fyne.TextWrapWord
	historyView.TextStyle = fyne.TextStyle{Monospace: true}
	refreshHistory := func() {
		h, err := core.LoadHistory()
		if err != nil {
			return
		}
		var lines []string
		for _, item := range h.Items {
			line := item.At.Format("2006-01-02 15:04") + "  " + strings.ToUpper(item.Status) + "  " + item.URL
			if item.Error != "" {
				line += "  (" + item.Error + ")"
			}
			lines = append(lines, line)
		}
		historyView.SetText(strings.Join(lines, "\n"))
	}
	refreshHistoryBtn := widget.NewButton(tr("btn.refresh"), refreshHistory)
	historyTab := container.NewVBox(
		widget.NewCard(tr("card.history"), "", container.NewVBox(
			container.NewHBox(refreshHistoryBtn),
			historyView,
		)),
	)

	queueToolbar := container.NewHBox(addQueueBtn, runQueueBtn, removeQueueBtn, retryQueueBtn, moveQueueUpBtn, moveQueueDownBtn, clearQueueBtn)
	queueScroll := container.NewVScroll(queueList)
	queueCard := widget.NewCard(tr("card.queue"), "", container.NewVBox(queueToolbar, queueScroll))
	queueTab := container.NewVBox(queueCard)

	leftTabs := container.NewAppTabs(
		container.NewTabItem(tr("tab.download"), container.NewVScroll(container.NewPadded(downloadTab))),
		container.NewTabItem(tr("tab.network"), container.NewVScroll(container.NewPadded(networkTab))),
		container.NewTabItem(tr("tab.playlist"), container.NewVScroll(container.NewPadded(playlistTab))),
		container.NewTabItem(tr("tab.extras"), container.NewVScroll(container.NewPadded(extrasTab))),
		container.NewTabItem(tr("tab.queue"), container.NewPadded(queueTab)),
		container.NewTabItem(tr("tab.history"), container.NewPadded(historyTab)),
		container.NewTabItem(tr("tab.tools"), container.NewPadded(toolsTab)),
	)
	leftTabs.SetTabLocation(container.TabLocationLeading)

	activityTitle := widget.NewLabel(tr("activity.title"))
	activityTitle.TextStyle = fyne.TextStyle{Bold: true}

	cmdAccordion := widget.NewAccordion(widget.NewAccordionItem(tr("command.title"), commandPreview))

	progressBlock := container.NewVBox(
		widget.NewSeparator(),
		progressFileLabel,
		progressFile,
		progressOverallLabel,
		progressOverall,
	)

	logHeader := widget.NewLabel(tr("log.title"))
	logHeader.TextStyle = fyne.TextStyle{Bold: true}

	topBar := container.NewHBox(
		journalBtn,
		openFolderBtn,
		layout.NewSpacer(),
		statusBadge.Root,
		phaseBadge.Root,
		stopBtn,
		runBtn,
	)

	rightHeader := container.NewVBox(
		activityTitle,
		topBar,
		widget.NewSeparator(),
		cmdAccordion,
		progressBlock,
		logHeader,
	)

	activityPanel := container.NewBorder(
		rightHeader,
		nil,
		nil,
		nil,
		container.NewVScroll(logs),
	)

	content := container.NewHSplit(
		leftTabs,
		container.NewPadded(activityPanel),
	)
	content.SetOffset(0.4)

	if appSettings.WindowWidth > 400 && appSettings.WindowHeight > 300 {
		w.Resize(fyne.NewSize(appSettings.WindowWidth, appSettings.WindowHeight))
	}
	w.SetContent(content)

	go func() {
		result := updater.CheckTools()
		uiExec(func() {
			if !result.YtDlp.Found {
				showYtDlpInstaller(w, result, addJournal, updatePreview)
			} else {
				addJournal(i18n.T("log.ytdlp_ok", map[string]interface{}{
					"Path":    result.YtDlp.Path,
					"Version": result.YtDlp.Version,
				}), nil)
				if !result.FFmpeg.Found {
					addJournal(tr("warn.ffmpeg"), nil)
				}
				if !result.Deno.Found {
					addJournal(tr("warn.deno"), nil)
				}
			}
			syncToolsStatus()
		})
	}()

	settingsTicker := time.NewTicker(30 * time.Second)
	go func() {
		for range settingsTicker.C {
			saveAppSettings()
		}
	}()

	w.SetCloseIntercept(func() {
		settingsTicker.Stop()
		p := strings.TrimSpace(sessionPathEntry.Text)
		if running.Load() && p != "" {
			snapMu.Lock()
			s := lastSession
			snapMu.Unlock()
			_ = core.SaveSession(p, s)
		}
		saveAppSettings()
		w.Close()
	})

	w.ShowAndRun()
}

func showYtDlpInstaller(w fyne.Window, result updater.CheckResult, addJournal func(string, error), updatePreview func()) {
	if result.YtDlp.Found {
		return
	}

	progressBar := widget.NewProgressBar()
	progressBar.Hide()
	statusLabel := widget.NewLabel(i18n.T("install.ytdlp.not_found", nil))
	statusLabel.Wrapping = fyne.TextWrapWord

	infoLabel := widget.NewLabel(i18n.T("install.ytdlp.body", nil))
	infoLabel.Wrapping = fyne.TextWrapWord

	ffmpegNote := widget.NewLabel("")
	if !result.FFmpeg.Found {
		ffmpegNote.SetText(i18n.T("warn.ffmpeg", nil))
	}
	if !result.Deno.Found {
		if ffmpegNote.Text != "" {
			ffmpegNote.SetText(ffmpegNote.Text + "\n" + i18n.T("warn.deno", nil))
		} else {
			ffmpegNote.SetText(i18n.T("warn.deno", nil))
		}
	}
	ffmpegNote.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(infoLabel, ffmpegNote, widget.NewSeparator(), statusLabel, progressBar)

	d := dialog.NewCustomConfirm(
		i18n.T("install.ytdlp.title", nil),
		i18n.T("btn.install", nil),
		i18n.T("btn.skip", nil),
		content,
		func(install bool) {
			if !install {
				addJournal(i18n.T("install.ytdlp.skipped", nil), nil)
				return
			}
			go func() {
				uiExec(func() {
					progressBar.Show()
					statusLabel.SetText(i18n.T("install.downloading", nil))
				})

				destDir := updater.DefaultInstallDir()
				path, err := updater.DownloadYtDlp(destDir, func(pct int) {
					uiExec(func() { progressBar.SetValue(float64(pct)) })
				})
				uiExec(func() {
					if err != nil {
						statusLabel.SetText("Download failed: " + err.Error())
						addJournal(i18n.T("install.ytdlp.failed", nil), err)
						return
					}
					statusLabel.SetText(i18n.T("install.ytdlp.done", map[string]interface{}{"Path": path}))
					progressBar.SetValue(100)
					addJournal(i18n.T("install.ytdlp.done", map[string]interface{}{"Path": path}), nil)
					updatePreview()
				})
			}()
		},
		w,
	)
	d.Resize(fyne.NewSize(560, 300))
	d.Show()
}

func localizedStatus(status string) string {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "complete" {
		s = "completed"
	}
	key := "status." + s
	label := i18n.T(key, nil)
	if label == key {
		return strings.ToUpper(status)
	}
	return strings.ToUpper(label)
}

func atoiOrZero(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func itoaOrEmpty(v int) string {
	if v == 0 {
		return ""
	}
	return strconv.Itoa(v)
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
