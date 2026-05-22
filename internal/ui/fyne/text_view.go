package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// newReadOnlyTextArea is a scroll-friendly read-only block (log, preview, history).
func newReadOnlyTextArea(minRows int, monospace bool) *widget.Entry {
	e := widget.NewMultiLineEntry()
	e.Wrapping = fyne.TextWrapWord
	if minRows > 0 {
		e.SetMinRowsVisible(minRows)
	}
	if monospace {
		e.TextStyle = fyne.TextStyle{Monospace: true}
	}
	e.Disable()
	return e
}
