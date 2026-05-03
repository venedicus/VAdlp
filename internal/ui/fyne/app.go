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
	w := a.NewWindow("🎬 VAdlp — Video-Audio dlp")
	w.Resize(fyne.NewSize(1240, 840))

	cfg := core.DefaultConfig()
	var issueEntries []string

	var snapMu sync.Mutex
	lastSession := core.Session{Config: cfg}

	statusBadge := NewStatusBadge("READY")
	statusBadge.SetStatus("READY")
	phaseBadge := NewPhaseBadge()

	issuesBtn := widget.NewButton("🔍 Diagnostics (0)", nil)

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
	progressFileLabel := widget.NewLabel("This file")
	progressOverallLabel := widget.NewLabel("Everything in this run")

	logs := widget.NewMultiLineEntry()
	logs.Disable()
	logs.Wrapping = fyne.TextWrapWord
	logs.TextStyle = fyne.TextStyle{Monospace: true}

	sessionPathEntry := widget.NewEntry()
	sessionPathEntry.SetPlaceHolder(`Optional: /path/session.json — auto-saved while downloading`)

	uiExec := func(fn func()) { fn() }

	addIssue := func(summary string, err error) {
		entry := summary
		if err != nil {
			entry = summary + ": " + err.Error()
		}
		timestamp := time.Now().Format("15:04:05")
		issueEntries = append(issueEntries, "["+timestamp+"] "+entry)
		issuesBtn.SetText(fmt.Sprintf("🔍 Diagnostics (%d)", len(issueEntries)))
	}

	issuesBtn.OnTapped = func() {
		content := "No problems recorded."
		if len(issueEntries) > 0 {
			content = strings.Join(issueEntries, "\n")
		}
		issueView := widget.NewMultiLineEntry()
		issueView.SetText(content)
		issueView.Disable()
		issueView.Wrapping = fyne.TextWrapWord
		issueView.TextStyle = fyne.TextStyle{Monospace: true}

		scroll := container.NewVScroll(issueView)
		scroll.SetMinSize(fyne.NewSize(720, 380))

		d := dialog.NewCustom("🔍 Diagnostics", "Close", scroll, w)
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

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Paste a video, playlist, or channel URL…")
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

	audioCheck := widget.NewCheck("Audio only (extract sound)", func(b bool) {
		cfg.AudioOnly = b
		updatePreview()
	})
	audioCheck.SetChecked(cfg.AudioOnly)

	audioFormatSelect := widget.NewSelect([]string{"", "mp3", "m4a", "opus", "wav", "flac", "vorbis"}, func(s string) {
		cfg.AudioFormat = s
		updatePreview()
	})
	audioFormatSelect.SetSelected(cfg.AudioFormat)

	pathEntry := widget.NewEntry()
	pathEntry.SetText(cfg.OutputPath)
	pathEntry.OnChanged = func(s string) {
		cfg.OutputPath = s
		updatePreview()
	}
	pickFolderBtn := widget.NewButton("📁 Browse", func() {
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

	cookiesBrowserCheck := widget.NewCheck("Use cookies from a browser profile", func(b bool) {
		cfg.UseCookiesBrowser = b
		updatePreview()
	})
	cookiesBrowserCheck.SetChecked(cfg.UseCookiesBrowser)

	cookiesBrowserSelect := widget.NewSelect([]string{"chrome", "firefox", "vivaldi", "edge", "brave"}, func(s string) {
		cfg.CookiesBrowser = s
		updatePreview()
	})
	cookiesBrowserSelect.SetSelected(cfg.CookiesBrowser)

	cookiesFileCheck := widget.NewCheck("Use cookies from a Netscape file", func(b bool) {
		cfg.UseCookiesFile = b
		updatePreview()
	})

	cookiesFileEntry := widget.NewEntry()
	cookiesFileEntry.SetPlaceHolder("/path/cookies.txt")
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

	reverseCheck := widget.NewCheck("Download playlist in reverse order", func(b bool) {
		cfg.PlaylistReverse = b
		updatePreview()
	})
	reverseCheck.SetChecked(cfg.PlaylistReverse)

	continueCheck := widget.NewCheck("Resume partial downloads", func(b bool) {
		cfg.Continue = b
		updatePreview()
	})
	continueCheck.SetChecked(cfg.Continue)

	noPartCheck := widget.NewCheck("Write directly to final file (no .part)", func(b bool) {
		cfg.NoPart = b
		updatePreview()
	})

	playlistStartEntry := widget.NewEntry()
	playlistStartEntry.SetPlaceHolder("empty = from first item")
	playlistStartEntry.OnChanged = func(s string) {
		cfg.PlaylistStart = atoiOrZero(s)
		updatePreview()
	}

	playlistEndEntry := widget.NewEntry()
	playlistEndEntry.SetPlaceHolder("empty = through last item")
	playlistEndEntry.OnChanged = func(s string) {
		cfg.PlaylistEnd = atoiOrZero(s)
		updatePreview()
	}

	maxDownloadsEntry := widget.NewEntry()
	maxDownloadsEntry.SetPlaceHolder("0 = unlimited")
	maxDownloadsEntry.OnChanged = func(s string) {
		cfg.MaxDownloads = atoiOrZero(s)
		updatePreview()
	}

	downloadArchiveEntry := widget.NewEntry()
	downloadArchiveEntry.SetPlaceHolder("archive.txt (skip finished IDs)")
	downloadArchiveEntry.OnChanged = func(s string) {
		cfg.DownloadArchive = s
		updatePreview()
	}

	noPlaylistCheck := widget.NewCheck("Single video only (--no-playlist)", func(b bool) {
		cfg.NoPlaylist = b
		updatePreview()
	})

	flatPlaylistCheck := widget.NewCheck("Flat playlist (--flat-playlist)", func(b bool) {
		cfg.FlatPlaylist = b
		updatePreview()
	})

	writeSubsCheck := widget.NewCheck("Write subtitles (--write-subs)", func(b bool) {
		cfg.WriteSubs = b
		updatePreview()
	})

	writeAutoSubCheck := widget.NewCheck("Auto subs (--write-auto-sub)", func(b bool) {
		cfg.WriteAutoSub = b
		updatePreview()
	})

	embedSubsCheck := widget.NewCheck("Embed subs (--embed-subs)", func(b bool) {
		cfg.EmbedSubs = b
		updatePreview()
	})

	subLangsEntry := widget.NewEntry()
	subLangsEntry.SetText(cfg.SubLangs)
	subLangsEntry.OnChanged = func(s string) {
		cfg.SubLangs = s
		updatePreview()
	}

	writeThumbCheck := widget.NewCheck("Thumbnail file (--write-thumbnail)", func(b bool) {
		cfg.WriteThumbnail = b
		updatePreview()
	})

	embedThumbCheck := widget.NewCheck("Embed thumbnail (--embed-thumbnail)", func(b bool) {
		cfg.EmbedThumbnail = b
		updatePreview()
	})

	embedMetaCheck := widget.NewCheck("Embed metadata (--embed-metadata)", func(b bool) {
		cfg.EmbedMetadata = b
		updatePreview()
	})

	embedChaptersCheck := widget.NewCheck("Embed chapters (--embed-chapters)", func(b bool) {
		cfg.EmbedChapters = b
		updatePreview()
	})

	writeInfoJSONCheck := widget.NewCheck("Write .info.json (for offline resume)", func(b bool) {
		cfg.WriteInfoJSON = b
		updatePreview()
	})

	loadInfoJSONEntry := widget.NewEntry()
	loadInfoJSONEntry.SetPlaceHolder("Path: --load-info-json (URL optional)")
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
	socketTimeoutEntry.SetPlaceHolder("seconds, 0 = default")
	socketTimeoutEntry.OnChanged = func(s string) {
		cfg.SocketTimeout = atoiOrZero(s)
		updatePreview()
	}

	noWarningsCheck := widget.NewCheck("No warnings (--no-warnings)", func(b bool) {
		cfg.NoWarnings = b
		updatePreview()
	})

	verboseCheck := widget.NewCheck("Verbose (-v)", func(b bool) {
		cfg.Verbose = b
		updatePreview()
	})

	quietCheck := widget.NewCheck("Quiet (-q)", func(b bool) {
		cfg.Quiet = b
		updatePreview()
	})

	windowsFilenamesCheck := widget.NewCheck("Windows-safe names (--windows-filenames)", func(b bool) {
		cfg.WindowsFilenames = b
		updatePreview()
	})

	noMtimeCheck := widget.NewCheck("No file mtime (--no-mtime)", func(b bool) {
		cfg.NoMtime = b
		updatePreview()
	})

	abortOnErrorCheck := widget.NewCheck("Abort on error (--abort-on-error)", func(b bool) {
		cfg.AbortOnError = b
		updatePreview()
	})

	ignoreErrorsCheck := widget.NewCheck("Ignore errors (--ignore-errors)", func(b bool) {
		cfg.IgnoreErrors = b
		updatePreview()
	})

	extraArgsEntry := widget.NewMultiLineEntry()
	extraArgsEntry.SetPlaceHolder("# one flag group per line, split by spaces\n# --extractor-args \"youtube:player_client=web\"")
	extraArgsEntry.SetMinRowsVisible(4)
	extraArgsEntry.OnChanged = func(s string) {
		cfg.ExtraArgs = s
		updatePreview()
	}

	syncUIFromCfg := func(c core.Config) {
		cfg = c
		urlEntry.SetText(cfg.URL)
		qualitySelect.SetSelected(cfg.Quality)
		formatSelect.SetSelected(cfg.Format)
		audioCheck.SetChecked(cfg.AudioOnly)
		audioFormatSelect.SetSelected(cfg.AudioFormat)
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
		updatePreview()
	}

	saveSessionBtn := widget.NewButton("💾 Export session…", func() {
		dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if err != nil {
				addIssue("Save session", err)
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
				addIssue("Save session", err)
				return
			}
			sessionPathEntry.SetText(path)
		}, w).Show()
	})

	loadSessionBtn := widget.NewButton("📂 Import session…", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				addIssue("Load session", err)
				return
			}
			if reader == nil {
				return
			}
			path := reader.URI().Path()
			_ = reader.Close()
			s, err := core.LoadSession(path)
			if err != nil {
				addIssue("Load session", err)
				return
			}
			resume := s.ApplyResumeHints()
			syncUIFromCfg(resume)
			snapMu.Lock()
			lastSession = s
			lastSession.Config = cfg
			snapMu.Unlock()
			sessionPathEntry.SetText(path)
			statusBadge.SetStatus("READY")
			phaseBadge.SetPhase(downloader.StageUnknown)
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".json"}))
		fd.Show()
	})

	applyResumeBtn := widget.NewButton("⏯️ Apply resume from session", func() {
		snapMu.Lock()
		s := lastSession
		snapMu.Unlock()
		if s.Version == 0 {
			addIssue("Resume", fmt.Errorf("load a session file first"))
			return
		}
		syncUIFromCfg(s.ApplyResumeHints())
	})

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

	runDownload := func(current core.Config, taskID string, qIdx, qTot int) bool {
		uiExec(func() {
			logs.SetText("")
			progressFile.SetValue(0)
			progressOverall.SetValue(0)
			statusBadge.SetStatus("RUNNING")
			phaseBadge.SetPhase(downloader.StageUnknown)
			progressFileLabel.SetText("This file")
			if qTot > 1 {
				progressOverallLabel.SetText(fmt.Sprintf("Queue: task %d of %d", qIdx, qTot))
			} else {
				progressOverallLabel.SetText("Everything in this run")
			}
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

		_, err := downloader.Run(current, func(ev downloader.Event) {
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
				uiExec(func() { progressFile.SetValue(ev.Progress) })
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
			uiExec(func() {
				statusBadge.SetStatus("ERROR")
				phaseBadge.SetPhase(downloader.StageUnknown)
				localLogs = append(localLogs, "ERROR: "+err.Error())
				logs.SetText(strings.Join(localLogs, "\n"))
			})
			addIssue("Download failed", err)
			if taskID != "" {
				updateTaskStatus(taskID, "error")
			}
			updateSnap()
			return false
		}

		uiExec(func() {
			statusBadge.SetStatus("COMPLETED")
			phaseBadge.SetPhase(downloader.StageUnknown)
			progressFile.SetValue(100)
			progressOverall.SetValue(100)
		})
		if taskID != "" {
			updateTaskStatus(taskID, "completed")
		}
		plCur, plTot = 0, 0
		filePct = 100
		updateSnap()
		return true
	}

	runBtn := widget.NewButton("⬇️ Start download", func() {
		if !running.CompareAndSwap(false, true) {
			return
		}
		go func(localCfg core.Config) {
			defer running.Store(false)
			runDownload(localCfg, "", 1, 1)
		}(cfg)
	})
	runBtn.Importance = widget.HighImportance

	addQueueBtn := widget.NewButton("➕ Add to queue", func() {
		if strings.TrimSpace(cfg.URL) == "" && strings.TrimSpace(cfg.LoadInfoJSON) == "" {
			addIssue("Queue add rejected", fmt.Errorf("enter a URL or a JSON path"))
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

	runQueueBtn := widget.NewButton("📋 Start queue", func() {
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
			allOK := true
			for i := range local {
				updateTaskStatus(local[i].ID, "running")
				if !runDownload(local[i].Config, local[i].ID, i+1, qTot) {
					allOK = false
					break
				}
			}

			uiExec(func() {
				if allOK {
					statusBadge.SetStatus("READY")
					phaseBadge.SetPhase(downloader.StageUnknown)
				}
			})
		}()
	})

	clearQueueBtn := widget.NewButton("🗑️ Clear", func() {
		if running.Load() {
			return
		}
		queueLock.Lock()
		queue = nil
		queueLock.Unlock()
		queueList.Refresh()
	})

	playlistPresetBtn := widget.NewButton("📺 YouTube playlist preset", func() {
		core.ApplyYouTubePlaylistPreset(&cfg)
		qualitySelect.SetSelected(cfg.Quality)
		formatSelect.SetSelected(cfg.Format)
		audioCheck.SetChecked(cfg.AudioOnly)
		reverseCheck.SetChecked(cfg.PlaylistReverse)
		cookiesBrowserCheck.SetChecked(cfg.UseCookiesBrowser)
		cookiesBrowserSelect.SetSelected(cfg.CookiesBrowser)
		updatePreview()
	})

	audioPresetBtn := widget.NewButton("🎵 Audio-only preset", func() {
		core.ApplyAudioOnlyPreset(&cfg)
		audioCheck.SetChecked(cfg.AudioOnly)
		formatSelect.SetSelected(cfg.Format)
		templateEntry.SetText(cfg.OutputTemplate)
		updatePreview()
	})

	templateHint := widget.NewLabel("Tip: use %(title)s, %(upload_date)s, %(uploader)s, %(ext)s in the pattern.")
	templateHint.Wrapping = fyne.TextWrapWord

	linkCard := widget.NewCard("What to download",
		"Paste any URL yt-dlp understands — one video, a playlist, or a channel.",
		widget.NewForm(widget.NewFormItem("Address", urlEntry)),
	)

	folderRow := container.NewBorder(nil, nil, nil, pickFolderBtn, pathEntry)
	outputCard := widget.NewCard("Where to save",
		"Browse if the default folder is not where you want files.",
		container.NewVBox(
			folderRow,
			widget.NewSeparator(),
			widget.NewForm(widget.NewFormItem("File name pattern", templateEntry)),
			templateHint,
		),
	)

	formatForm := widget.NewForm(
		widget.NewFormItem("Quality", qualitySelect),
		widget.NewFormItem("Container (when merging)", formatSelect),
		widget.NewFormItem("", audioCheck),
		widget.NewFormItem("Audio format (if audio only)", audioFormatSelect),
	)
	formatCard := widget.NewCard("Picture & sound",
		"Quality picks streams; container sets the merged file type.",
		container.NewVBox(
			formatForm,
			widget.NewSeparator(),
			container.NewGridWithColumns(2, playlistPresetBtn, audioPresetBtn),
		),
	)

	downloadTab := container.NewVBox(linkCard, widget.NewSeparator(), outputCard, widget.NewSeparator(), formatCard)

	networkForm := widget.NewForm(
		widget.NewFormItem("", cookiesBrowserCheck),
		widget.NewFormItem("Browser", cookiesBrowserSelect),
		widget.NewFormItem("", cookiesFileCheck),
		widget.NewFormItem("Cookies file", cookiesFileEntry),
		widget.NewFormItem("HTTP proxy", proxyEntry),
		widget.NewFormItem("Speed limit", rateEntry),
	)
	networkCard := widget.NewCard("Network & sign-in",
		"Use cookies when a site asks you to log in or age-gate content.",
		networkForm,
	)
	networkTab := container.NewVBox(networkCard)

	playlistForm := widget.NewForm(
		widget.NewFormItem("", reverseCheck),
		widget.NewFormItem("", continueCheck),
		widget.NewFormItem("", noPartCheck),
		widget.NewFormItem("Start at item #", playlistStartEntry),
		widget.NewFormItem("End at item #", playlistEndEntry),
		widget.NewFormItem("Max items (0 = no limit)", maxDownloadsEntry),
		widget.NewFormItem("Archive file", downloadArchiveEntry),
		widget.NewFormItem("", noPlaylistCheck),
		widget.NewFormItem("", flatPlaylistCheck),
	)
	playlistCard := widget.NewCard("Playlist behaviour",
		"Start/end numbers are inclusive.",
		playlistForm,
	)
	sessionHelp := widget.NewLabel("If you set a session file, it auto-saves a few times per minute and when you close the app during a download. Import it later, then tap \"Apply resume from session\".")
	sessionHelp.Wrapping = fyne.TextWrapWord
	sessionCard := widget.NewCard("Resume & sessions", "",
		container.NewVBox(
			sessionHelp,
			widget.NewForm(widget.NewFormItem("Session file (.json)", sessionPathEntry)),
			container.NewHBox(saveSessionBtn, loadSessionBtn, applyResumeBtn),
		),
	)
	playlistTab := container.NewVBox(playlistCard, widget.NewSeparator(), sessionCard)

	extrasMediaForm := widget.NewForm(
		widget.NewFormItem("", writeSubsCheck),
		widget.NewFormItem("", writeAutoSubCheck),
		widget.NewFormItem("", embedSubsCheck),
		widget.NewFormItem("Subtitle languages", subLangsEntry),
		widget.NewFormItem("", writeThumbCheck),
		widget.NewFormItem("", embedThumbCheck),
		widget.NewFormItem("", embedMetaCheck),
		widget.NewFormItem("", embedChaptersCheck),
		widget.NewFormItem("", writeInfoJSONCheck),
		widget.NewFormItem("Load info JSON path", loadInfoJSONEntry),
	)
	extrasMediaCard := widget.NewCard("Subtitles, thumbnails, metadata",
		"Embedding runs after download and needs ffmpeg.",
		extrasMediaForm,
	)

	extrasNetForm := widget.NewForm(
		widget.NewFormItem("Network retries", retriesEntry),
		widget.NewFormItem("Fragment retries", fragRetriesEntry),
		widget.NewFormItem("Parallel fragments", concFragEntry),
		widget.NewFormItem("Socket timeout (seconds)", socketTimeoutEntry),
		widget.NewFormItem("", noWarningsCheck),
		widget.NewFormItem("", verboseCheck),
		widget.NewFormItem("", quietCheck),
		widget.NewFormItem("", windowsFilenamesCheck),
		widget.NewFormItem("", noMtimeCheck),
		widget.NewFormItem("", abortOnErrorCheck),
		widget.NewFormItem("", ignoreErrorsCheck),
	)
	extrasNetCard := widget.NewCard("Reliability & output hygiene",
		"Raise retries on flaky Wi‑Fi.",
		extrasNetForm,
	)

	extrasPowerCard := widget.NewCard("Extra yt-dlp flags",
		"One group of flags per line. Lines starting with # are ignored.",
		extraArgsEntry,
	)
	extrasTab := container.NewVBox(extrasMediaCard, widget.NewSeparator(), extrasNetCard, widget.NewSeparator(), extrasPowerCard)

	queueIntro := widget.NewLabel("Each row keeps the settings from when you added it. Only tasks marked \"queued\" run when you start the queue.")
	queueIntro.Wrapping = fyne.TextWrapWord
	queueToolbar := container.NewHBox(addQueueBtn, runQueueBtn, clearQueueBtn)
	queueScroll := container.NewVScroll(queueList)
	queueCard := widget.NewCard("Batch downloads", "", container.NewVBox(queueIntro, queueToolbar, queueScroll))
	queueTab := container.NewVBox(queueCard)

	leftTabs := container.NewAppTabs(
		container.NewTabItem("📥 Download", container.NewVScroll(container.NewPadded(downloadTab))),
		container.NewTabItem("🌐 Network", container.NewVScroll(container.NewPadded(networkTab))),
		container.NewTabItem("📑 Playlist", container.NewVScroll(container.NewPadded(playlistTab))),
		container.NewTabItem("✨ Extras", container.NewVScroll(container.NewPadded(extrasTab))),
		container.NewTabItem("📋 Queue", container.NewPadded(queueTab)),
	)
	leftTabs.SetTabLocation(container.TabLocationLeading)

	activityTitle := widget.NewLabel("📊 Status & output")
	activityTitle.TextStyle = fyne.TextStyle{Bold: true}

	cmdAccordion := widget.NewAccordion(widget.NewAccordionItem("⌨️ Full command line (yt-dlp …)", commandPreview))

	progressBlock := container.NewVBox(
		widget.NewSeparator(),
		progressFileLabel,
		progressFile,
		progressOverallLabel,
		progressOverall,
	)

	logHeader := widget.NewLabel("📜 yt-dlp log")
	logHeader.TextStyle = fyne.TextStyle{Bold: true}

	topBar := container.NewHBox(
		issuesBtn,
		layout.NewSpacer(),
		statusBadge.Root,
		phaseBadge.Root,
		runBtn,
	)

	// Fixed header content (non-growing widgets)
	rightHeader := container.NewVBox(
		activityTitle,
		topBar,
		widget.NewSeparator(),
		cmdAccordion,
		progressBlock,
		logHeader,
	)

	// logs fills the remaining vertical space via Border layout
	activityPanel := container.NewBorder(
		rightHeader,                // top — fixed
		nil,                        // bottom
		nil,                        // left
		nil,                        // right
		container.NewVScroll(logs), // center — expands
	)

	content := container.NewHSplit(
		leftTabs,
		container.NewPadded(activityPanel),
	)
	content.SetOffset(0.4)

	w.SetContent(content)

	go func() {
		result := updater.CheckTools()

		if !result.YtDlp.Found {
			// Show install dialog on the main goroutine via a channel trick.
			// Fyne dialogs must be created/shown from the goroutine that owns
			// the event loop. We schedule via a short timer so w.ShowAndRun
			// has time to initialise.
			time.Sleep(300 * time.Millisecond)
			showYtDlpInstaller(w, result, addIssue, updatePreview)
		} else {
			addIssue("yt-dlp found: "+result.YtDlp.Path+" ("+result.YtDlp.Version+")", nil)
			if !result.FFmpeg.Found {
				addIssue("ffmpeg not found — install it for merging formats and audio extraction", nil)
			}
		}
	}()

	w.SetCloseIntercept(func() {
		p := strings.TrimSpace(sessionPathEntry.Text)
		if running.Load() && p != "" {
			snapMu.Lock()
			s := lastSession
			snapMu.Unlock()
			_ = core.SaveSession(p, s)
			uiExec(func() { statusBadge.SetStatus("STOPPED") })
		}
		w.Close()
	})

	w.ShowAndRun()
}

// showYtDlpInstaller displays a dialog offering to download yt-dlp when it's
// not found on the system.
func showYtDlpInstaller(w fyne.Window, result updater.CheckResult, addIssue func(string, error), updatePreview func()) {
	if result.YtDlp.Found {
		return
	}

	progressBar := widget.NewProgressBar()
	progressBar.Hide()
	statusLabel := widget.NewLabel("yt-dlp was not found on this system.")
	statusLabel.Wrapping = fyne.TextWrapWord

	infoLabel := widget.NewLabel(
		"VAdlp requires yt-dlp to download media.\n" +
			"You can install it now (downloaded from github.com/yt-dlp/yt-dlp)\n" +
			"or place it next to this application manually.",
	)
	infoLabel.Wrapping = fyne.TextWrapWord

	ffmpegNote := widget.NewLabel("")
	if !result.FFmpeg.Found {
		ffmpegNote.SetText("⚠️  ffmpeg was also not found. Install it separately for format merging and audio extraction.")
	}
	ffmpegNote.Wrapping = fyne.TextWrapWord

	content := container.NewVBox(infoLabel, ffmpegNote, widget.NewSeparator(), statusLabel, progressBar)

	d := dialog.NewCustomConfirm(
		"🔧 Install yt-dlp",
		"Install now",
		"Skip",
		content,
		func(install bool) {
			if !install {
				addIssue("yt-dlp not found — downloads will fail until it is installed", nil)
				return
			}
			// Run download in background, update UI via label/progress.
			go func() {
				progressBar.Show()
				statusLabel.SetText("Downloading yt-dlp…")

				destDir := updater.DefaultInstallDir()
				path, err := updater.DownloadYtDlp(destDir, func(pct int) {
					progressBar.SetValue(float64(pct))
				})
				if err != nil {
					statusLabel.SetText("Download failed: " + err.Error())
					addIssue("yt-dlp install failed", err)
					return
				}
				statusLabel.SetText("Installed: " + path)
				progressBar.SetValue(100)
				addIssue("yt-dlp installed at "+path, nil)
				updatePreview()
			}()
		},
		w,
	)
	d.Resize(fyne.NewSize(560, 300))
	d.Show()
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
