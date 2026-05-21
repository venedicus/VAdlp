package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// adaptLayout adjusts the activity split on narrow windows so tabs stay readable.
func adaptLayout(canvasSize fyne.Size, split *container.Split, savedOffset float64) {
	if split == nil {
		return
	}
	w := canvasSize.Width
	off := savedOffset
	if off <= 0.05 || off >= 0.95 {
		off = 0.4
	}
	switch {
	case w < 820:
		off = 0.22
	case w < 1024:
		off = 0.30
	case w < 1200:
		if off > 0.38 {
			off = 0.38
		}
	}
	split.SetOffset(off)
}
