package fyneui

import "fyne.io/fyne/v2"

func uiExec(fn func()) {
	fyne.Do(fn)
}
