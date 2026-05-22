package fyneui

import (
	"fyne.io/fyne/v2"
)

const (
	baseWindowWidth  float32 = 1240
	baseWindowHeight float32 = 840
	screenFitRatio   float32 = 0.92
	fallbackScreenW  float32 = 1280
	fallbackScreenH  float32 = 800
)

// UIScaleAuto (0) means pick scale from the available work area.
const UIScaleAuto float32 = 0

// WorkAreaSize returns the usable area for the window (menu bars, dock excluded when known).
func WorkAreaSize(w fyne.Window) fyne.Size {
	if w != nil {
		_, area := w.Canvas().InteractiveArea()
		if area.Width >= 640 && area.Height >= 400 {
			return area
		}
		if s := w.Canvas().Size(); s.Width >= 640 && s.Height >= 400 {
			return s
		}
	}
	return fyne.NewSize(fallbackScreenW, fallbackScreenH)
}

// AutoUIScale picks a readable scale for the given work area without user tuning.
func AutoUIScale(screen fyne.Size) float32 {
	if screen.Width < 1360 || screen.Height < 820 {
		return UIScaleCompact
	}
	if screen.Width < 1680 || screen.Height < 980 {
		return UIScaleComfortable
	}
	return UIScaleLarge
}

// EffectiveUIScale resolves Auto (0) against the work area; explicit presets are normalized.
func EffectiveUIScale(stored float32, screen fyne.Size) float32 {
	if stored <= 0 || stored == UIScaleAuto {
		return AutoUIScale(screen)
	}
	return NormalizeUIScale(stored)
}

// IdealWindowSize fits the app window inside the work area. Theme scale does not inflate window pixels.
func IdealWindowSize(screen fyne.Size, savedW, savedH float32) fyne.Size {
	maxW := screen.Width * screenFitRatio
	maxH := screen.Height * screenFitRatio
	w, h := baseWindowWidth, baseWindowHeight
	if savedW > 400 && savedH > 300 {
		w, h = savedW, savedH
	}
	w = clampFloat(w, 720, maxW)
	h = clampFloat(h, 480, maxH)
	if w > maxW {
		w = maxW
	}
	if h > maxH {
		h = maxH
	}
	return fyne.NewSize(w, h)
}

// DialogSize caps dialog dimensions to a fraction of the parent window or work area.
func DialogSize(parent fyne.Window, wantW, wantH float32) fyne.Size {
	limit := WorkAreaSize(parent)
	if parent != nil {
		cs := parent.Canvas().Size()
		if cs.Width > 0 && cs.Height > 0 {
			limit = cs
		}
	}
	maxW := limit.Width * 0.94
	maxH := limit.Height * 0.92
	w := minFloat(wantW, maxW)
	h := minFloat(wantH, maxH)
	if w < 320 {
		w = minFloat(maxW, 320)
	}
	if h < 200 {
		h = minFloat(maxH, 200)
	}
	return fyne.NewSize(w, h)
}

func FitWindow(w fyne.Window, storedScale float32, savedW, savedH float32) {
	area := WorkAreaSize(w)
	size := IdealWindowSize(area, savedW, savedH)
	w.Resize(size)
	w.CenterOnScreen()
}

func clampFloat(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func minFloat(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
