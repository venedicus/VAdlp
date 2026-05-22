package fyneui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestAutoUIScaleSmallLaptop(t *testing.T) {
	got := AutoUIScale(fyne.NewSize(1280, 800))
	if got != UIScaleCompact {
		t.Fatalf("got %v want compact", got)
	}
}

func TestEffectiveUIScaleAuto(t *testing.T) {
	got := EffectiveUIScale(0, fyne.NewSize(1280, 800))
	if got != UIScaleCompact {
		t.Fatalf("got %v", got)
	}
}

func TestIdealWindowSizeClampsToScreen(t *testing.T) {
	screen := fyne.NewSize(1280, 800)
	size := IdealWindowSize(screen, 0, 0)
	if size.Width > screen.Width*screenFitRatio+1 {
		t.Fatalf("width %v too large for screen", size.Width)
	}
	if size.Height > screen.Height*screenFitRatio+1 {
		t.Fatalf("height %v too large for screen", size.Height)
	}
}

func TestIdealWindowSizeRespectsSavedWithinScreen(t *testing.T) {
	screen := fyne.NewSize(1440, 900)
	size := IdealWindowSize(screen, 1100, 700)
	if size.Width != 1100 || size.Height != 700 {
		t.Fatalf("got %v", size)
	}
}

func TestDialogSizeCapsToFallback(t *testing.T) {
	got := DialogSize(nil, 760, 460)
	if got.Width > fallbackScreenW*0.94+1 {
		t.Fatalf("width %v", got.Width)
	}
}
