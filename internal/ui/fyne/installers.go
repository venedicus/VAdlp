package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/i18n"
	"vadlp/internal/updater"
)

func showToolInstaller(
	w fyne.Window,
	titleKey, bodyKey string,
	probeLocal func() updater.ToolStatus,
	install func(dir string, progress func(int)) (string, error),
	onDone func(path string, err error),
	addJournal func(string, error),
) {
	if st := probeLocal(); st.Found {
		dialog.ShowInformation(
			i18n.T(titleKey, nil),
			i18n.T("install.already", map[string]interface{}{"Path": st.Path}),
			w,
		)
		onDone(st.Path, nil)
		return
	}

	statusLabel := widget.NewLabel("")
	statusLabel.Wrapping = fyne.TextWrapWord
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	info := widget.NewLabel(i18n.T(bodyKey, nil))
	info.Wrapping = fyne.TextWrapWord

	var d dialog.Dialog
	installBtn := widget.NewButton(i18n.T("btn.install", nil), nil)
	installBtn.OnTapped = func() {
		installBtn.Disable()
		progressBar.Show()
		progressBar.SetValue(0)
		statusLabel.SetText(i18n.T("install.downloading", nil))
		go func() {
			destDir := updater.DefaultInstallDir()
			path, err := install(destDir, func(pct int) {
				uiExec(func() { progressBar.SetValue(float64(pct)) })
			})
			uiExec(func() {
				installBtn.Enable()
				if err != nil {
					statusLabel.SetText(err.Error())
					addJournal(i18n.T("install.failed", nil), err)
					return
				}
				statusLabel.SetText(path)
				progressBar.SetValue(100)
				addJournal(i18n.T("install.done", map[string]interface{}{"Path": path}), nil)
				onDone(path, nil)
			})
		}()
	}

	body := container.NewVBox(
		info,
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		container.NewHBox(installBtn),
	)

	d = dialog.NewCustom(i18n.T(titleKey, nil), i18n.T("btn.close", nil), body, w)
	d.Resize(fyne.NewSize(520, 260))
	d.Show()
}

func showFFmpegInstaller(w fyne.Window, addJournal func(string, error), onPath func(string)) {
	global := updater.CheckTools().FFmpeg
	if global.Found && !updater.ProbeLocalFFmpeg().Found {
		dialog.ShowInformation(
			i18n.T("install.ffmpeg.title", nil),
			i18n.T("install.found_elsewhere", map[string]interface{}{"Path": global.Path}),
			w,
		)
		onPath(global.Path)
		return
	}
	showToolInstaller(w, "install.ffmpeg.title", "install.ffmpeg.body", updater.ProbeLocalFFmpeg, updater.DownloadFFmpeg, func(path string, err error) {
		if err == nil {
			onPath(path)
		}
	}, addJournal)
}

func showDenoInstaller(w fyne.Window, addJournal func(string, error), onPath func(string)) {
	global := updater.CheckDeno()
	if global.Found && !updater.ProbeLocalDeno().Found {
		dialog.ShowInformation(
			i18n.T("install.deno.title", nil),
			i18n.T("install.found_elsewhere", map[string]interface{}{"Path": global.Path}),
			w,
		)
		onPath(global.Path)
		return
	}
	showToolInstaller(w, "install.deno.title", "install.deno.body", updater.ProbeLocalDeno, updater.DownloadDeno, func(path string, err error) {
		if err == nil {
			onPath(path)
		}
	}, addJournal)
}
