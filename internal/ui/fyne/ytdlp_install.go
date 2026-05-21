package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/i18n"
	"vadlp/internal/updater"
)

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
						statusLabel.SetText(i18n.T("install.download_failed", map[string]interface{}{"Error": err.Error()}))
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
