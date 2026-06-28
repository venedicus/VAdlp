package app

import (
	"github.com/energye/systray"
	"github.com/gen2brain/beeep"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"vadlp/internal/i18n"
)

// startTray runs the system tray icon on its own goroutine. systray.Run blocks
// until systray.Quit is called, so it must not run on the goroutine that owns
// the webview's event loop (wails.Run, started from main).
func (a *App) startTray() {
	go systray.Run(a.onTrayReady, func() {})
}

func (a *App) onTrayReady() {
	systray.SetIcon(generateTrayIconICO())
	systray.SetTitle("VAdlp")
	systray.SetTooltip(i18n.T("tray.tooltip", nil))

	showItem := systray.AddMenuItem(i18n.T("tray.show", nil), "")
	systray.AddSeparator()
	quitItem := systray.AddMenuItem(i18n.T("tray.quit", nil), "")

	i18n.OnLanguageChange(func() {
		systray.SetTooltip(i18n.T("tray.tooltip", nil))
		showItem.SetTitle(i18n.T("tray.show", nil))
		quitItem.SetTitle(i18n.T("tray.quit", nil))
	})

	showItem.Click(func() {
		runtime.WindowShow(a.ctx)
		runtime.WindowUnminimise(a.ctx)
	})
	quitItem.Click(func() {
		a.allowQuit.Store(true)
		runtime.Quit(a.ctx)
	})
}

func (a *App) stopTray() {
	systray.Quit()
}

func (a *App) notify(title, message string) {
	_ = beeep.Notify(title, message, "")
}
