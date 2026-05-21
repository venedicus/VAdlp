package fyneui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/downloader"
	"vadlp/internal/i18n"
)

type StatusBadge struct {
	bg    *canvas.Rectangle
	label *widget.Label
	Root  fyne.CanvasObject
}

func NewStatusBadge(key string) *StatusBadge {
	bg := canvas.NewRectangle(mainStatusColor(key))
	bg.CornerRadius = 4
	lbl := widget.NewLabel(statusLabel(key))
	lbl.Alignment = fyne.TextAlignCenter
	root := container.NewStack(bg, container.NewPadded(lbl))
	return &StatusBadge{bg: bg, label: lbl, Root: root}
}

func statusLabel(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "" {
		k = "ready"
	}
	return strings.ToUpper(i18n.T("status."+k, nil))
}

func (s *StatusBadge) SetStatusKey(key string) {
	s.label.SetText(statusLabel(key))
	s.bg.FillColor = mainStatusColor(key)
	s.bg.Refresh()
}

func mainStatusColor(key string) color.Color {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "ready":
		return color.NRGBA{R: 0x3c, G: 0x40, B: 0x5a, A: 0xff}
	case "running":
		return color.NRGBA{R: 0x2f, G: 0x64, B: 0xb8, A: 0xff}
	case "completed", "complete":
		return color.NRGBA{R: 0x3d, G: 0x8c, B: 0x4a, A: 0xff}
	case "error":
		return color.NRGBA{R: 0xb8, G: 0x3d, B: 0x4f, A: 0xff}
	case "stopped", "cancelled":
		return color.NRGBA{R: 0xc6, G: 0x7a, B: 0x2f, A: 0xff}
	case "queued":
		return color.NRGBA{R: 0x6c, G: 0x4a, B: 0xb8, A: 0xff}
	case "paused":
		return color.NRGBA{R: 0x3d, G: 0x6e, B: 0x8c, A: 0xff}
	default:
		return color.NRGBA{R: 0x3c, G: 0x40, B: 0x5a, A: 0xff}
	}
}

type PhaseBadge struct {
	bg    *canvas.Rectangle
	label *widget.Label
	Root  fyne.CanvasObject
}

func NewPhaseBadge() *PhaseBadge {
	bg := canvas.NewRectangle(phaseColor(downloader.StageUnknown))
	bg.CornerRadius = 4
	lbl := widget.NewLabel(phaseLabel(downloader.StageUnknown))
	lbl.Alignment = fyne.TextAlignCenter
	root := container.NewStack(bg, container.NewPadded(lbl))
	return &PhaseBadge{bg: bg, label: lbl, Root: root}
}

func phaseLabel(st downloader.Stage) string {
	switch st {
	case downloader.StageExtracting:
		return i18n.T("phase.extracting", nil)
	case downloader.StageDownloading:
		return i18n.T("phase.downloading", nil)
	case downloader.StagePostProcess:
		return i18n.T("phase.postprocess", nil)
	default:
		return i18n.T("phase.idle", nil)
	}
}

func (p *PhaseBadge) SetPhase(st downloader.Stage) {
	p.label.SetText(phaseLabel(st))
	p.bg.FillColor = phaseColor(st)
	p.bg.Refresh()
}

func phaseColor(st downloader.Stage) color.Color {
	switch st {
	case downloader.StageExtracting:
		return color.NRGBA{R: 0x8c, G: 0x6a, B: 0x24, A: 0xff}
	case downloader.StageDownloading:
		return color.NRGBA{R: 0x2a, G: 0x55, B: 0xa8, A: 0xff}
	case downloader.StagePostProcess:
		return color.NRGBA{R: 0x6b, G: 0x3d, B: 0xa8, A: 0xff}
	default:
		return color.NRGBA{R: 0x34, G: 0x38, B: 0x4d, A: 0xff}
	}
}
