package fyneui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/applog"
	"vadlp/internal/core"
	"vadlp/internal/downloader"
	"vadlp/internal/i18n"
	"vadlp/internal/settings"
	"vadlp/internal/updater"
	"vadlp/internal/version"
)

type mainTabs struct {
	AppTabs *container.AppTabs
	Items   struct {
		Download *container.TabItem
		Network  *container.TabItem
		Playlist *container.TabItem
		Extras   *container.TabItem
		Queue    *container.TabItem
		History  *container.TabItem
		Tools    *container.TabItem
	}
	RefreshHistory    func()
	RefreshHistoryBtn *widget.Button
	SyncToolsStatus   func()
}

type tabFields struct {
	W                fyne.Window
	Cfg              *core.Config
	AppSettings      *settings.App
	Tr               func(string) string
	Bind             *LocaleBinder
	AddJournal       func(string, error)
	SaveAppSettings  func()
	SetWindowTitle   func(string)
	UpdatePreview    func()
	SyncToolsStatus  func()
	OnUIScaleChanged func(float32)
	App              fyne.App

	URLEntry        *widget.Entry
	BatchURLEntry   *widget.Entry
	PathEntry       *widget.Entry
	PickFolderBtn   *widget.Button
	TemplateEntry   *widget.Entry
	FormatSection   fyne.CanvasObject
	ProfileSection  fyne.CanvasObject
	PasteURLBtn     *widget.Button
	FetchFormatsBtn *widget.Button
	SetQuality      func(string)

	CookiesBrowserCheck  *widget.Check
	CookiesBrowserSelect *widget.Select
	CookiesFileCheck     *widget.Check
	CookiesFileEntry     *widget.Entry
	PickCookiesBtn       *widget.Button
	ProxyEntry           *widget.Entry
	RateEntry            *widget.Entry
	UsernameEntry        *widget.Entry
	PasswordEntry        *widget.Entry

	ReverseCheck         *widget.Check
	ContinueCheck        *widget.Check
	NoPartCheck          *widget.Check
	PlaylistStartEntry   *widget.Entry
	PlaylistEndEntry     *widget.Entry
	MaxDownloadsEntry    *widget.Entry
	DownloadArchiveEntry *widget.Entry
	NoPlaylistCheck      *widget.Check
	FlatPlaylistCheck    *widget.Check
	SessionPathEntry     *widget.Entry
	SaveSessionBtn       *widget.Button
	LoadSessionBtn       *widget.Button
	ApplyResumeBtn       *widget.Button

	WriteSubsCheck        *widget.Check
	WriteAutoSubCheck     *widget.Check
	EmbedSubsCheck        *widget.Check
	SubLangsEntry         *widget.Entry
	WriteThumbCheck       *widget.Check
	EmbedThumbCheck       *widget.Check
	EmbedMetaCheck        *widget.Check
	EmbedChaptersCheck    *widget.Check
	WriteInfoJSONCheck    *widget.Check
	LoadInfoJSONEntry     *widget.Entry
	RetriesEntry          *widget.Entry
	FragRetriesEntry      *widget.Entry
	ConcFragEntry         *widget.Entry
	SocketTimeoutEntry    *widget.Entry
	NoWarningsCheck       *widget.Check
	VerboseCheck          *widget.Check
	QuietCheck            *widget.Check
	WindowsFilenamesCheck *widget.Check
	NoMtimeCheck          *widget.Check
	AbortOnErrorCheck     *widget.Check
	IgnoreErrorsCheck     *widget.Check
	SponsorBlockCheck     *widget.Check
	ExtraArgsEntry        *widget.Entry

	FFmpegPathEntry *widget.Entry
	DenoPathEntry   *widget.Entry
	ParallelEntry   *widget.Entry

	QueueList    *widget.List
	QueueToolbar fyne.CanvasObject
}

func buildMainTabs(f tabFields) *mainTabs {
	tr := f.Tr
	bind := f.Bind

	f.PasteURLBtn.OnTapped = func() {
		if text := f.App.Clipboard().Content(); text != "" {
			f.URLEntry.SetText(text)
		}
	}
	f.FetchFormatsBtn.OnTapped = func() {
		showFormatPicker(f.W, *f.Cfg, f.SetQuality)
	}

	urlActions := container.NewHBox(f.PasteURLBtn, f.FetchFormatsBtn)
	urlRow := container.NewBorder(nil, nil, nil, urlActions, f.URLEntry)
	urlSec := Section(tr("card.url"), tr("drop.hint"), urlRow)
	batchSec := Section(tr("card.batch"), "", f.BatchURLEntry)
	bind.BindPlaceholder(f.BatchURLEntry, "placeholder.batch_urls", tr)

	folderRow := container.NewBorder(nil, nil, nil, f.PickFolderBtn, f.PathEntry)
	templateItem := widget.NewFormItem(tr("form.filename_template"), f.TemplateEntry)
	outputForm := widget.NewForm(templateItem)
	outputSec := Section(tr("card.output"), "", container.NewVBox(folderRow, outputForm))
	bind.BindFormItem(templateItem, "form.filename_template", tr)

	downloadTab := container.NewVBox(
		f.ProfileSection,
		urlSec.Root,
		batchSec.Root,
		outputSec.Root,
		f.FormatSection,
	)
	bind.BindSection(urlSec, "card.url", "drop.hint", tr)
	bind.BindSection(batchSec, "card.batch", "", tr)
	bind.BindSection(outputSec, "card.output", "", tr)

	cookiesFileRow := container.NewBorder(nil, nil, nil, f.PickCookiesBtn, f.CookiesFileEntry)
	fiBrowser := widget.NewFormItem(tr("form.browser"), f.CookiesBrowserSelect)
	fiCookies := widget.NewFormItem(tr("form.cookies_file"), cookiesFileRow)
	fiProxy := widget.NewFormItem(tr("form.proxy"), f.ProxyEntry)
	fiRate := widget.NewFormItem(tr("form.rate_limit"), f.RateEntry)
	fiUser := widget.NewFormItem(tr("form.username"), f.UsernameEntry)
	fiPass := widget.NewFormItem(tr("form.password"), f.PasswordEntry)
	networkForm := widget.NewForm(
		widget.NewFormItem("", f.CookiesBrowserCheck),
		fiBrowser,
		widget.NewFormItem("", f.CookiesFileCheck),
		fiCookies, fiProxy, fiRate, fiUser, fiPass,
	)
	networkSec := Section(tr("card.network"), "", networkForm)
	networkTab := container.NewVBox(networkSec.Root)
	bind.BindSection(networkSec, "card.network", "", tr)
	bind.BindCheck(f.CookiesBrowserCheck, "check.cookies_browser", tr)
	bind.BindCheck(f.CookiesFileCheck, "check.cookies_file", tr)
	bind.BindFormItem(fiBrowser, "form.browser", tr)
	bind.BindFormItem(fiCookies, "form.cookies_file", tr)
	bind.BindFormItem(fiProxy, "form.proxy", tr)
	bind.BindFormItem(fiRate, "form.rate_limit", tr)
	bind.BindFormItem(fiUser, "form.username", tr)
	bind.BindFormItem(fiPass, "form.password", tr)
	bind.BindPlaceholder(f.ProxyEntry, "placeholder.proxy", tr)
	bind.BindPlaceholder(f.RateEntry, "placeholder.rate", tr)
	bind.BindPlaceholder(f.CookiesFileEntry, "placeholder.cookies", tr)

	fiPlStart := widget.NewFormItem(tr("form.playlist_start"), f.PlaylistStartEntry)
	fiPlEnd := widget.NewFormItem(tr("form.playlist_end"), f.PlaylistEndEntry)
	fiMaxDl := widget.NewFormItem(tr("form.max_downloads"), f.MaxDownloadsEntry)
	fiArchive := widget.NewFormItem(tr("form.archive"), f.DownloadArchiveEntry)
	playlistForm := widget.NewForm(
		widget.NewFormItem("", f.ReverseCheck),
		widget.NewFormItem("", f.ContinueCheck),
		widget.NewFormItem("", f.NoPartCheck),
		fiPlStart, fiPlEnd, fiMaxDl, fiArchive,
		widget.NewFormItem("", f.NoPlaylistCheck),
		widget.NewFormItem("", f.FlatPlaylistCheck),
	)
	playlistSec := Section(tr("card.playlist"), "", playlistForm)
	fiSession := widget.NewFormItem(tr("form.session_path"), f.SessionPathEntry)
	sessionSec := Section(tr("card.session"), "",
		container.NewVBox(
			widget.NewForm(fiSession),
			container.NewHBox(f.SaveSessionBtn, f.LoadSessionBtn, f.ApplyResumeBtn),
		),
	)
	playlistTab := container.NewVBox(playlistSec.Root, sessionSec.Root)
	bind.BindSection(playlistSec, "card.playlist", "", tr)
	bind.BindSection(sessionSec, "card.session", "", tr)
	bind.BindCheck(f.ReverseCheck, "check.reverse", tr)
	bind.BindCheck(f.ContinueCheck, "check.continue", tr)
	bind.BindCheck(f.NoPartCheck, "check.no_part", tr)
	bind.BindCheck(f.NoPlaylistCheck, "check.no_playlist", tr)
	bind.BindCheck(f.FlatPlaylistCheck, "check.flat_playlist", tr)
	bind.BindFormItem(fiPlStart, "form.playlist_start", tr)
	bind.BindFormItem(fiPlEnd, "form.playlist_end", tr)
	bind.BindFormItem(fiMaxDl, "form.max_downloads", tr)
	bind.BindFormItem(fiArchive, "form.archive", tr)
	bind.BindFormItem(fiSession, "form.session_path", tr)
	bind.BindPlaceholder(f.PlaylistStartEntry, "placeholder.playlist_start", tr)
	bind.BindPlaceholder(f.PlaylistEndEntry, "placeholder.playlist_end", tr)
	bind.BindPlaceholder(f.MaxDownloadsEntry, "placeholder.max_downloads", tr)
	bind.BindPlaceholder(f.DownloadArchiveEntry, "placeholder.archive", tr)

	fiSubLangs := widget.NewFormItem(tr("form.sub_langs"), f.SubLangsEntry)
	fiLoadInfo := widget.NewFormItem(tr("form.load_info_json"), f.LoadInfoJSONEntry)
	extrasMediaForm := widget.NewForm(
		widget.NewFormItem("", f.WriteSubsCheck),
		widget.NewFormItem("", f.WriteAutoSubCheck),
		widget.NewFormItem("", f.EmbedSubsCheck),
		fiSubLangs,
		widget.NewFormItem("", f.WriteThumbCheck),
		widget.NewFormItem("", f.EmbedThumbCheck),
		widget.NewFormItem("", f.EmbedMetaCheck),
		widget.NewFormItem("", f.EmbedChaptersCheck),
		widget.NewFormItem("", f.WriteInfoJSONCheck),
		fiLoadInfo,
	)
	fiRetries := widget.NewFormItem(tr("form.retries"), f.RetriesEntry)
	fiFrag := widget.NewFormItem(tr("form.frag_retries"), f.FragRetriesEntry)
	fiConc := widget.NewFormItem(tr("form.concurrent_frags"), f.ConcFragEntry)
	fiSocket := widget.NewFormItem(tr("form.socket_timeout"), f.SocketTimeoutEntry)
	extrasNetForm := widget.NewForm(
		fiRetries, fiFrag, fiConc, fiSocket,
		widget.NewFormItem("", f.NoWarningsCheck),
		widget.NewFormItem("", f.VerboseCheck),
		widget.NewFormItem("", f.QuietCheck),
		widget.NewFormItem("", f.WindowsFilenamesCheck),
		widget.NewFormItem("", f.NoMtimeCheck),
		widget.NewFormItem("", f.AbortOnErrorCheck),
		widget.NewFormItem("", f.IgnoreErrorsCheck),
	)
	mediaSec := Section(tr("card.media_extras"), "", extrasMediaForm)
	retriesSec := Section(tr("card.retries"), "", extrasNetForm)
	sponsorSec := Section(tr("card.sponsorblock"), "", f.SponsorBlockCheck)
	extraSec := Section(tr("card.extra_flags"), "", f.ExtraArgsEntry)
	extrasTab := container.NewVBox(mediaSec.Root, retriesSec.Root, sponsorSec.Root, extraSec.Root)
	bind.BindSection(mediaSec, "card.media_extras", "", tr)
	bind.BindSection(retriesSec, "card.retries", "", tr)
	bind.BindSection(sponsorSec, "card.sponsorblock", "", tr)
	bind.BindSection(extraSec, "card.extra_flags", "", tr)
	bind.BindCheck(f.WriteSubsCheck, "check.write_subs", tr)
	bind.BindCheck(f.WriteAutoSubCheck, "check.write_auto_sub", tr)
	bind.BindCheck(f.EmbedSubsCheck, "check.embed_subs", tr)
	bind.BindCheck(f.WriteThumbCheck, "check.write_thumb", tr)
	bind.BindCheck(f.EmbedThumbCheck, "check.embed_thumb", tr)
	bind.BindCheck(f.EmbedMetaCheck, "check.embed_meta", tr)
	bind.BindCheck(f.EmbedChaptersCheck, "check.embed_chapters", tr)
	bind.BindCheck(f.WriteInfoJSONCheck, "check.write_info_json", tr)
	bind.BindCheck(f.NoWarningsCheck, "check.no_warnings", tr)
	bind.BindCheck(f.VerboseCheck, "check.verbose", tr)
	bind.BindCheck(f.QuietCheck, "check.quiet", tr)
	bind.BindCheck(f.WindowsFilenamesCheck, "check.windows_filenames", tr)
	bind.BindCheck(f.NoMtimeCheck, "check.no_mtime", tr)
	bind.BindCheck(f.AbortOnErrorCheck, "check.abort_on_error", tr)
	bind.BindCheck(f.IgnoreErrorsCheck, "check.ignore_errors", tr)
	bind.BindCheck(f.SponsorBlockCheck, "check.sponsorblock", tr)
	bind.BindFormItem(fiSubLangs, "form.sub_langs", tr)
	bind.BindFormItem(fiLoadInfo, "form.load_info_json", tr)
	bind.BindFormItem(fiRetries, "form.retries", tr)
	bind.BindFormItem(fiFrag, "form.frag_retries", tr)
	bind.BindFormItem(fiConc, "form.concurrent_frags", tr)
	bind.BindFormItem(fiSocket, "form.socket_timeout", tr)
	bind.BindPlaceholder(f.LoadInfoJSONEntry, "placeholder.load_info_json", tr)
	bind.BindPlaceholder(f.SocketTimeoutEntry, "placeholder.socket_timeout", tr)
	bind.BindPlaceholder(f.ExtraArgsEntry, "placeholder.extra_args", tr)

	updateYtDlpBtn := widget.NewButton(tr("btn.update_ytdlp"), func() {
		bin, err := downloader.ResolveBinary()
		if err != nil {
			dialog.ShowError(err, f.W)
			return
		}
		go func() {
			out, err := updater.UpdateYtDlp(bin)
			uiExec(func() {
				if err != nil {
					dialog.ShowError(err, f.W)
					return
				}
				dialog.ShowInformation(tr("dialog.ytdlp_update"), out, f.W)
				f.UpdatePreview()
			})
		}()
	})
	bind.BindButton(updateYtDlpBtn, "btn.update_ytdlp", tr)

	langSelect := widget.NewSelect([]string{"en", "ru"}, func(code string) {
		f.AppSettings.Language = code
		i18n.SetLanguage(code)
		f.SetWindowTitle(tr("app.title"))
		f.SaveAppSettings()
	})
	if f.AppSettings.Language == "ru" {
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
	bind.BindButton(installFFmpegBtn, "btn.install_ffmpeg", tr)
	bind.BindButton(installDenoBtn, "btn.install_deno", tr)

	syncToolsStatus := func() {
		r := updater.CheckTools()
		if r.FFmpeg.Found {
			ffmpegStatusLabel.SetText(i18n.T("tools.ffmpeg_ok", map[string]interface{}{"Path": r.FFmpeg.Path}))
			installFFmpegBtn.Disable()
		} else {
			ffmpegStatusLabel.SetText(i18n.T("tools.ffmpeg_missing", nil))
			installFFmpegBtn.Enable()
		}
		if r.Deno.Found {
			denoStatusLabel.SetText(i18n.T("tools.deno_ok", map[string]interface{}{"Path": r.Deno.Path}))
			installDenoBtn.Disable()
		} else {
			denoStatusLabel.SetText(i18n.T("tools.deno_missing", nil))
			installDenoBtn.Enable()
		}
	}

	installFFmpegBtn.OnTapped = func() {
		showFFmpegInstaller(f.W, f.AddJournal, func(path string) {
			f.FFmpegPathEntry.SetText(path)
			f.Cfg.FFmpegLocation = path
			f.AppSettings.FFmpegPath = path
			f.UpdatePreview()
			syncToolsStatus()
		})
	}
	installDenoBtn.OnTapped = func() {
		showDenoInstaller(f.W, f.AddJournal, func(path string) {
			f.DenoPathEntry.SetText(path)
			f.Cfg.DenoPath = path
			f.AppSettings.DenoPath = path
			f.UpdatePreview()
			syncToolsStatus()
		})
	}
	syncToolsStatus()

	fiLang := widget.NewFormItem(tr("form.language"), langSelect)
	fiFFmpeg := widget.NewFormItem(tr("form.ffmpeg_path"), f.FFmpegPathEntry)
	fiDeno := widget.NewFormItem(tr("form.deno_path"), f.DenoPathEntry)
	fiWorkers := widget.NewFormItem(tr("form.queue_workers"), f.ParallelEntry)

	scaleOpts := []struct {
		key string
		val float32
	}{
		{"ui_scale.compact", UIScaleCompact},
		{"ui_scale.comfortable", UIScaleComfortable},
		{"ui_scale.large", UIScaleLarge},
		{"ui_scale.extra_large", UIScaleExtraLarge},
	}
	scaleLabels := make([]string, len(scaleOpts))
	scaleLabelToVal := make(map[string]float32, len(scaleOpts))
	for i, o := range scaleOpts {
		scaleLabels[i] = tr(o.key)
		scaleLabelToVal[scaleLabels[i]] = o.val
	}
	uiScaleSelect := widget.NewSelect(scaleLabels, func(label string) {
		if v, ok := scaleLabelToVal[label]; ok {
			f.AppSettings.UIScale = v
			if f.OnUIScaleChanged != nil {
				f.OnUIScaleChanged(v)
			}
			f.SaveAppSettings()
		}
	})
	curScale := NormalizeUIScale(f.AppSettings.UIScale)
	for _, o := range scaleOpts {
		if NormalizeUIScale(o.val) == curScale {
			uiScaleSelect.SetSelected(tr(o.key))
			break
		}
	}
	fiScale := widget.NewFormItem(tr("form.ui_scale"), uiScaleSelect)
	bind.BindFormItem(fiScale, "form.ui_scale", tr)
	bind.Add(func() {
		for _, o := range scaleOpts {
			if NormalizeUIScale(o.val) == NormalizeUIScale(f.AppSettings.UIScale) {
				uiScaleSelect.SetSelected(tr(o.key))
				break
			}
		}
	})

	debugCheck := widget.NewCheck(tr("check.debug_log"), func(on bool) {
		f.AppSettings.DebugLog = on
		_ = applog.Init(on)
		f.SaveAppSettings()
	})
	debugCheck.SetChecked(f.AppSettings.DebugLog)
	bind.BindCheck(debugCheck, "check.debug_log", tr)
	toolsForm := widget.NewForm(
		fiLang, fiScale, fiFFmpeg, fiDeno, fiWorkers,
		widget.NewFormItem("", debugCheck),
	)
	bind.BindFormItem(fiLang, "form.language", tr)
	bind.BindFormItem(fiFFmpeg, "form.ffmpeg_path", tr)
	bind.BindFormItem(fiDeno, "form.deno_path", tr)
	bind.BindFormItem(fiWorkers, "form.queue_workers", tr)

	appVersionLabel := widget.NewLabel(i18n.T("tools.app_version", map[string]interface{}{
		"Version": version.Label(),
	}))
	appVersionLabel.Wrapping = fyne.TextWrapWord
	bind.Add(func() {
		appVersionLabel.SetText(i18n.T("tools.app_version", map[string]interface{}{
			"Version": version.Label(),
		}))
	})

	binariesSec := Section(tr("card.binaries"), "",
		container.NewVBox(
			toolsForm,
			ffmpegStatusLabel,
			denoStatusLabel,
			container.NewHBox(updateYtDlpBtn, installFFmpegBtn, installDenoBtn),
			appVersionLabel,
		),
	)
	bind.BindSection(binariesSec, "card.binaries", "", tr)
	toolsTab := container.NewVBox(binariesSec.Root)

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
			line := item.At.Format("2006-01-02 15:04") + "  " + localizedStatus(item.Status) + "  " + item.URL
			if item.DurationSec > 0 {
				line += fmt.Sprintf("  (%ds)", item.DurationSec)
			}
			if item.Error != "" {
				line += "  (" + item.Error + ")"
			}
			lines = append(lines, line)
		}
		historyView.SetText(strings.Join(lines, "\n"))
	}
	refreshHistoryBtn := widget.NewButton(tr("btn.refresh"), refreshHistory)
	bind.BindButton(refreshHistoryBtn, "btn.refresh", tr)
	historySec := Section(tr("card.history"), "",
		container.NewVBox(container.NewHBox(refreshHistoryBtn), historyView),
	)
	bind.BindSection(historySec, "card.history", "", tr)
	historyTab := container.NewVBox(historySec.Root)
	refreshHistory()

	queueScroll := container.NewVScroll(f.QueueList)
	queueScroll.SetMinSize(fyne.NewSize(0, 220*NormalizeUIScale(f.AppSettings.UIScale)))
	queueSec := Section(tr("card.queue"), "", container.NewVBox(f.QueueToolbar, queueScroll))
	bind.BindSection(queueSec, "card.queue", "", tr)
	queueTab := container.NewVBox(queueSec.Root)

	out := &mainTabs{
		RefreshHistory:    refreshHistory,
		RefreshHistoryBtn: refreshHistoryBtn,
		SyncToolsStatus:   syncToolsStatus,
	}
	out.Items.Download = container.NewTabItem(tr("tab.download"), ScrollTab(downloadTab))
	out.Items.Network = container.NewTabItem(tr("tab.network"), ScrollTab(networkTab))
	out.Items.Playlist = container.NewTabItem(tr("tab.playlist"), ScrollTab(playlistTab))
	out.Items.Extras = container.NewTabItem(tr("tab.extras"), ScrollTab(extrasTab))
	out.Items.Queue = container.NewTabItem(tr("tab.queue"), ScrollTab(queueTab))
	out.Items.History = container.NewTabItem(tr("tab.history"), ScrollTab(historyTab))
	out.Items.Tools = container.NewTabItem(tr("tab.tools"), ScrollTab(toolsTab))

	out.AppTabs = container.NewAppTabs(
		out.Items.Download, out.Items.Network, out.Items.Playlist, out.Items.Extras,
		out.Items.Queue, out.Items.History, out.Items.Tools,
	)
	out.AppTabs.SetTabLocation(container.TabLocationLeading)
	return out
}
