package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type hotkeyTarget struct {
	app      fyne.App
	urlEntry *widget.Entry
	runBtn   *widget.Button
	stopBtn  *widget.Button
}

func setupHotkeys(w fyne.Window, h hotkeyTarget) {
	c, ok := w.Canvas().(desktop.Canvas)
	if !ok {
		return
	}
	c.SetOnKeyDown(func(ev *fyne.KeyEvent) {
		mods := currentModifiers()
		switch ev.Name {
		case fyne.KeyV:
			if mods&fyne.KeyModifierControl != 0 {
				if text := h.app.Clipboard().Content(); text != "" {
					h.urlEntry.SetText(text)
				}
			}
		case fyne.KeyReturn:
			if mods&fyne.KeyModifierControl != 0 && h.runBtn != nil && h.runBtn.OnTapped != nil {
				h.runBtn.OnTapped()
			}
		case fyne.KeyEscape:
			if h.stopBtn != nil && h.stopBtn.OnTapped != nil {
				h.stopBtn.OnTapped()
			}
		}
	})
}

func currentModifiers() fyne.KeyModifier {
	if d, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		return d.CurrentKeyModifiers()
	}
	return 0
}
