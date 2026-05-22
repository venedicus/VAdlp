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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/applog"
	"vadlp/internal/core"
	"vadlp/internal/downloader"
	"vadlp/internal/i18n"
	"vadlp/internal/service"
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

	appSettings, _ := settings.Load()
	bootArea := fyne.NewSize(fallbackScreenW, fallbackScreenH)
	ApplyTheme(a, EffectiveUIScale(appSettings.UIScale, bootArea))
	_ = applog.Init(appSettings.DebugLog)
	lang := appSettings.Language
	if lang == "" {
		lang = "en"
	}
	if err := i18n.Init(lang); err != nil {
		lang = "en"
		_ = i18n.Init(lang)
	}
	bind := NewLocaleBinder()
	tr := func(id string) string { return i18n.T(id, nil) }
	i18n.OnLanguageChange(func() { bind.RefreshDeferred() })

	w := a.NewWindow(tr("app.title"))

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

	commandPreview := newReadOnlyTextArea(4, true)

	progressFile := widget.NewProgressBar()
	progressFile.Min = 0
	progressFile.Max = 100
	progressOverall := widget.NewProgressBar()
	progressOverall.Min = 0
	progressOverall.Max = 100
	progressFileLabel := widget.NewLabel(tr("progress.file"))
	progressOverallLabel := widget.NewLabel(tr("progress.overall"))

	logs := newReadOnlyTextArea(10, false)

	sessionPathEntry := widget.NewEntry()
	sessionPathEntry.SetPlaceHolder(tr("placeholder.session_path"))
	if appSettings.SessionPath != "" {
		sessionPathEntry.SetText(appSettings.SessionPath)
	}

	var setQuality func(string)
	var saveAppSettings func()
	var layoutShell *layoutShell
	var activityAccordion *widget.Accordion

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
		issueView := newReadOnlyTextArea(12, true)
		issueView.SetText(content)

		scroll := container.NewVScroll(issueView)
		dlgSize := DialogSize(w, 760, 460)
		scroll.SetMinSize(fyne.NewSize(dlgSize.Width-40, dlgSize.Height-80))

		d := dialog.NewCustom(tr("journal.title"), tr("btn.close"), scroll, w)
		d.Resize(dlgSize)
		d.Show()
	}

	saveAppSettings = func() {
		appSettings.Config = cfg
		appSettings.SessionPath = strings.TrimSpace(sessionPathEntry.Text)
		size := w.Canvas().Size()
		appSettings.WindowWidth = size.Width
		appSettings.WindowHeight = size.Height
		if layoutShell != nil && layoutShell.activeSplit != nil {
			appSettings.ActivityPanelOffset = layoutShell.activeSplit.Offset
		}
		if activityAccordion != nil && len(activityAccordion.Items) > 0 {
			appSettings.ActivityPanelOpen = activityAccordion.Items[0].Open
		}
		if err := settings.Save(appSettings); err != nil {
			journalFromErr(tr, addJournal, "err.save_settings", err)
		}
	}

	updatePreview := func() {
		prog := "yt-dlp"
		if p, err := downloader.ResolveBinary(); err == nil {
			prog = p
		}
		commandPreview.SetText(core.PreviewCommand(cfg, prog))
	}

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder(tr("placeholder.url"))
	urlEntry.OnChanged = func(s string) {
		cfg.URL = s
		updatePreview()
	}
	setupURLDrop(w, func(s string) {
		urlEntry.SetText(s)
		cfg.URL = s
		updatePreview()
	})

	formatUI, formatSection := NewDownloadFormatUI(&cfg, updatePreview, tr, bind)

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

	var verboseCheck, quietCheck *widget.Check
	quietCheck = widget.NewCheck(tr("check.quiet"), func(b bool) {
		cfg.Quiet = b
		if b {
			cfg.Verbose = false
			verboseCheck.SetChecked(false)
		}
		cfg.Normalize()
		updatePreview()
	})

	verboseCheck = widget.NewCheck(tr("check.verbose"), func(b bool) {
		cfg.Verbose = b
		if b {
			cfg.Quiet = false
			quietCheck.SetChecked(false)
		}
		cfg.Normalize()
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
			return newQueueRow().object()
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			queueLock.Lock()
			item := queue[i]
			queueLock.Unlock()
			root := o.(*fyne.Container)
			dot := root.Objects[2].(*fyne.Container).Objects[0].(*canvas.Rectangle)
			cancel := root.Objects[3].(*widget.Button)
			hbox := root.Objects[4].(*fyne.Container)
			name := hbox.Objects[0].(*widget.Label)
			status := hbox.Objects[2].(*widget.Label)
			name.SetText(fmt.Sprintf("%s %d — %s", tr("queue.task"), i+1, item.Name))
			status.SetText(LocalizedStatus(item.Status, false))
			cancel.SetText(tr("btn.cancel_task"))
			dot.FillColor = StatusColor(item.Status)
			dot.Refresh()
			taskID := item.ID
			if item.Status == "running" {
				cancel.Show()
				cancel.Enable()
				cancel.OnTapped = func() { downloader.CancelJob(taskID) }
			} else {
				cancel.Hide()
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

	stopBtn := widget.NewButton(tr("btn.stop"), nil)
	stopBtn.Disable()

	runBtn := widget.NewButton(tr("btn.download"), nil)
	runBtn.Importance = widget.HighImportance

	openFolderBtn := widget.NewButton(tr("btn.open_folder"), func() {
		if err := openFolder(cfg.OutputPath); err != nil {
			addJournal(tr("err.open_folder"), err)
		}
	})

	addQueueBtn := widget.NewButton(tr("btn.add_queue"), nil)
	runQueueBtn := widget.NewButton(tr("btn.run_queue"), nil)
	removeQueueBtn := widget.NewButton(tr("btn.remove"), nil)
	retryQueueBtn := widget.NewButton(tr("btn.retry_failed"), nil)
	moveQueueUpBtn := widget.NewButton(tr("btn.up"), nil)
	moveQueueDownBtn := widget.NewButton(tr("btn.down"), nil)
	clearQueueBtn := widget.NewButton(tr("btn.clear_queue"), nil)
	pasteURLBtn := widget.NewButton(tr("btn.paste"), nil)
	fetchFormatsBtn := widget.NewButton(tr("btn.fetch_formats"), nil)

	_, profileSection, profileActions := NewProfileBar(w, &cfg, &appSettings, saveAppSettings, syncUIFromCfg, tr, bind)

	runLabel := widget.NewLabel(tr("queue.group_run"))
	runLabel.TextStyle = fyne.TextStyle{Bold: true}
	editLabel := widget.NewLabel(tr("queue.group_edit"))
	editLabel.TextStyle = fyne.TextStyle{Bold: true}
	queueRunGroup := container.NewHBox(addQueueBtn, runQueueBtn)
	queueEditGroup := container.NewHBox(removeQueueBtn, retryQueueBtn, moveQueueUpBtn, moveQueueDownBtn, clearQueueBtn)
	queueToolbar := container.NewVBox(
		runLabel, queueRunGroup,
		editLabel, queueEditGroup,
	)

	dlRunner := &downloadRunner{
		svc:                  service.New(),
		tr:                   tr,
		addJournal:           addJournal,
		saveAppSettings:      saveAppSettings,
		sessionPathEntry:     sessionPathEntry,
		snapMu:               &snapMu,
		lastSession:          &lastSession,
		logs:                 logs,
		progressFile:         progressFile,
		progressOverall:      progressOverall,
		progressFileLabel:    progressFileLabel,
		progressOverallLabel: progressOverallLabel,
		statusBadge:          statusBadge,
		phaseBadge:           phaseBadge,
		stopBtn:              stopBtn,
		runBtn:               runBtn,
		updateTaskStatus:     updateTaskStatus,
		cfg:                  &cfg,
		appSettings:          &appSettings,
		running:              &running,
		currentJobID:         &currentJobID,
	}

	onUIScaleChanged := func(storedScale float32) {
		area := WorkAreaSize(w)
		ApplyTheme(a, EffectiveUIScale(storedScale, area))
		cs := w.Canvas().Size()
		FitWindow(w, storedScale, cs.Width, cs.Height)
	}

	tabs := buildMainTabs(tabFields{
		W: w, App: a, Cfg: &cfg, AppSettings: &appSettings, Tr: tr, Bind: bind,
		AddJournal: addJournal, SaveAppSettings: saveAppSettings,
		SetWindowTitle: w.SetTitle, UpdatePreview: updatePreview,
		OnUIScaleChanged: onUIScaleChanged,
		URLEntry:         urlEntry, BatchURLEntry: batchURLEntry, PathEntry: pathEntry,
		PickFolderBtn: pickFolderBtn, TemplateEntry: templateEntry,
		FormatSection: formatSection, ProfileSection: profileSection,
		PasteURLBtn: pasteURLBtn, FetchFormatsBtn: fetchFormatsBtn, SetQuality: setQuality,
		CookiesBrowserCheck: cookiesBrowserCheck, CookiesBrowserSelect: cookiesBrowserSelect,
		CookiesFileCheck: cookiesFileCheck, CookiesFileEntry: cookiesFileEntry, PickCookiesBtn: pickCookiesBtn,
		ProxyEntry: proxyEntry, RateEntry: rateEntry, UsernameEntry: usernameEntry, PasswordEntry: passwordEntry,
		ReverseCheck: reverseCheck, ContinueCheck: continueCheck, NoPartCheck: noPartCheck,
		PlaylistStartEntry: playlistStartEntry, PlaylistEndEntry: playlistEndEntry,
		MaxDownloadsEntry: maxDownloadsEntry, DownloadArchiveEntry: downloadArchiveEntry,
		NoPlaylistCheck: noPlaylistCheck, FlatPlaylistCheck: flatPlaylistCheck,
		SessionPathEntry: sessionPathEntry, SaveSessionBtn: saveSessionBtn,
		LoadSessionBtn: loadSessionBtn, ApplyResumeBtn: applyResumeBtn,
		WriteSubsCheck: writeSubsCheck, WriteAutoSubCheck: writeAutoSubCheck, EmbedSubsCheck: embedSubsCheck,
		SubLangsEntry: subLangsEntry, WriteThumbCheck: writeThumbCheck, EmbedThumbCheck: embedThumbCheck,
		EmbedMetaCheck: embedMetaCheck, EmbedChaptersCheck: embedChaptersCheck,
		WriteInfoJSONCheck: writeInfoJSONCheck, LoadInfoJSONEntry: loadInfoJSONEntry,
		RetriesEntry: retriesEntry, FragRetriesEntry: fragRetriesEntry, ConcFragEntry: concFragEntry,
		SocketTimeoutEntry: socketTimeoutEntry, NoWarningsCheck: noWarningsCheck,
		VerboseCheck: verboseCheck, QuietCheck: quietCheck,
		WindowsFilenamesCheck: windowsFilenamesCheck, NoMtimeCheck: noMtimeCheck,
		AbortOnErrorCheck: abortOnErrorCheck, IgnoreErrorsCheck: ignoreErrorsCheck,
		SponsorBlockCheck: sponsorBlockCheck, ExtraArgsEntry: extraArgsEntry,
		FFmpegPathEntry: ffmpegPathEntry, DenoPathEntry: denoPathEntry, ParallelEntry: parallelEntry,
		QueueList: queueList, QueueToolbar: queueToolbar,
	})
	dlRunner.refreshHistory = tabs.RefreshHistory

	(&queueController{
		runner: dlRunner, queue: &queue, queueLock: &queueLock, queueList: queueList,
		selectedQueueIdx: &selectedQueueIdx, cfg: &cfg, addJournal: addJournal, tr: tr,
	}).wireButtons(runBtn, addQueueBtn, runQueueBtn, removeQueueBtn, retryQueueBtn, moveQueueUpBtn, moveQueueDownBtn, clearQueueBtn)

	stopBtn.OnTapped = func() {
		dlRunner.cancelActiveJob()
		if v := currentJobID.Load(); v != nil {
			if id, ok := v.(string); ok && id != "" {
				downloader.CancelJob(id)
			}
		}
	}

	leftTabs := tabs.AppTabs
	tabDownload := tabs.Items.Download
	tabNetwork := tabs.Items.Network
	tabPlaylist := tabs.Items.Playlist
	tabExtras := tabs.Items.Extras
	tabQueue := tabs.Items.Queue
	tabHistory := tabs.Items.History
	tabTools := tabs.Items.Tools
	syncToolsStatus := tabs.SyncToolsStatus

	cmdAccordion := widget.NewAccordion(widget.NewAccordionItem(tr("command.title"), commandPreview))

	progressBlock := container.NewVBox(
		progressFileLabel,
		progressFile,
		progressOverallLabel,
		progressOverall,
	)

	logHeader := widget.NewLabel(tr("log.title"))
	logHeader.TextStyle = fyne.TextStyle{Bold: true}

	topBar := newTopBar(journalBtn, openFolderBtn, statusBadge, phaseBadge, stopBtn, runBtn)

	logPanel := newActivityLogPanel(logHeader, logs)
	activitySplit, activityBody := newActivityBody(cmdAccordion, progressBlock, logPanel)
	activityAccordion = widget.NewAccordion(widget.NewAccordionItem(tr("activity.title"), activityBody))
	if appSettings.ActivityPanelOpen {
		activityAccordion.Open(0)
	}

	activityPanel := container.NewVBox(topBar.Object(), activityAccordion)

	layoutShell = newLayoutShell(leftTabs, container.NewPadded(activityPanel), appSettings.ActivityPanelOffset)
	if off := appSettings.ActivityPanelOffset; off > 0.05 && off < 0.95 {
		layoutShell.splitH.SetOffset(off)
		layoutShell.splitV.SetOffset(off)
	} else {
		layoutShell.splitH.SetOffset(0.4)
		layoutShell.splitV.SetOffset(0.58)
	}
	layoutShell.activityAccordion = activityAccordion
	layoutShell.activitySplit = activitySplit
	layoutShell.topBar = topBar
	layoutShell.formatUI = formatUI
	layoutShell.profileActions = profileActions
	layoutShell.commandPreview = commandPreview
	layoutShell.batchURLEntry = batchURLEntry
	layoutShell.extraArgsEntry = extraArgsEntry
	layoutShell.statusBadge = statusBadge
	layoutShell.phaseBadge = phaseBadge
	layoutShell.queueScroll = tabs.QueueScroll
	content := layoutShell.Root()

	bind.Add(func() {
		w.SetTitle(tr("app.title"))
		tabDownload.Text = tr("tab.download")
		tabNetwork.Text = tr("tab.network")
		tabPlaylist.Text = tr("tab.playlist")
		tabExtras.Text = tr("tab.extras")
		tabQueue.Text = tr("tab.queue")
		tabHistory.Text = tr("tab.history")
		tabTools.Text = tr("tab.tools")
		journalBtn.SetText(fmt.Sprintf("%s (%d)", tr("journal.title"), len(journalEntries)))
		logHeader.SetText(tr("log.title"))
		activityAccordion.Items[0].Title = tr("activity.title")
		cmdAccordion.Items[0].Title = tr("command.title")
		progressFileLabel.SetText(tr("progress.file"))
		progressOverallLabel.SetText(tr("progress.overall"))
		statusBadge.RefreshText()
		phaseBadge.RefreshText()
		queueList.Refresh()
		runLabel.SetText(tr("queue.group_run"))
		editLabel.SetText(tr("queue.group_edit"))
		tabs.RefreshToolsLabels()
	})
	bind.BindLabel(runLabel, "queue.group_run", tr)
	bind.BindLabel(editLabel, "queue.group_edit", tr)
	bind.BindButton(openFolderBtn, "btn.open_folder", tr)
	bind.BindButton(stopBtn, "btn.stop", tr)
	bind.BindButton(runBtn, "btn.download", tr)
	bind.BindButton(addQueueBtn, "btn.add_queue", tr)
	bind.BindButton(runQueueBtn, "btn.run_queue", tr)
	bind.BindButton(removeQueueBtn, "btn.remove", tr)
	bind.BindButton(retryQueueBtn, "btn.retry_failed", tr)
	bind.BindButton(moveQueueUpBtn, "btn.up", tr)
	bind.BindButton(moveQueueDownBtn, "btn.down", tr)
	bind.BindButton(clearQueueBtn, "btn.clear_queue", tr)
	bind.BindButton(pasteURLBtn, "btn.paste", tr)
	bind.BindButton(fetchFormatsBtn, "btn.fetch_formats", tr)
	bind.BindButton(pickFolderBtn, "btn.browse", tr)
	bind.BindButton(pickCookiesBtn, "btn.browse", tr)
	bind.BindButton(saveSessionBtn, "btn.save_session", tr)
	bind.BindButton(loadSessionBtn, "btn.load_session", tr)
	bind.BindButton(applyResumeBtn, "btn.resume_session", tr)
	bind.BindPlaceholder(urlEntry, "placeholder.url", tr)
	bind.BindPlaceholder(sessionPathEntry, "placeholder.session_path", tr)

	setupHotkeys(w, hotkeyTarget{app: a, urlEntry: urlEntry, runBtn: runBtn, stopBtn: stopBtn})

	w.SetContent(content)
	winSize := IdealWindowSize(WorkAreaSize(w), appSettings.WindowWidth, appSettings.WindowHeight)
	w.Resize(winSize)
	w.CenterOnScreen()
	adaptLayout(winSize, layoutShell, appSettings.ActivityPanelOffset)

	layoutStop := make(chan struct{})
	var stopLayoutOnce sync.Once
	stopLayout := func() { stopLayoutOnce.Do(func() { close(layoutStop) }) }
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		var lastBucket uint64
		for {
			select {
			case <-layoutStop:
				return
			case <-ticker.C:
				s := w.Canvas().Size()
				if s.Width < 200 || s.Height < 200 {
					continue
				}
				b := layoutSizeBucket(s.Width, s.Height)
				if b == lastBucket {
					continue
				}
				lastBucket = b
				size := s
				uiExec(func() {
					adaptLayout(size, layoutShell, appSettings.ActivityPanelOffset)
				})
			}
		}
	}()

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
		stopLayout()
		settingsTicker.Stop()
		dlRunner.cancelActiveJob()
		downloader.CancelAll()
		if len(activityAccordion.Items) > 0 {
			appSettings.ActivityPanelOpen = activityAccordion.Items[0].Open
		}
		p := strings.TrimSpace(sessionPathEntry.Text)
		if running.Load() && p != "" {
			snapMu.Lock()
			s := lastSession
			snapMu.Unlock()
			if err := core.SaveSession(p, s); err != nil {
				addJournal(tr("err.save_session"), err)
			}
		}
		saveAppSettings()
		w.Close()
	})

	w.ShowAndRun()
}
