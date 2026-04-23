package fyneui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"ytgui/internal/downloader"
)

type StatusBadge struct {
	bg    *canvas.Rectangle
	label *widget.Label
	Root  fyne.CanvasObject
}

func NewStatusBadge(initial string) *StatusBadge {
	bg := canvas.NewRectangle(mainStatusColor(initial))
	bg.CornerRadius = 4
	lbl := widget.NewLabel(initial)
	lbl.Alignment = fyne.TextAlignCenter
	root := container.NewStack(bg, container.NewPadded(lbl))
	return &StatusBadge{bg: bg, label: lbl, Root: root}
}

func (s *StatusBadge) SetStatus(text string) {
	s.label.SetText(statusWithEmoji(text))
	s.bg.FillColor = mainStatusColor(text)
	s.bg.Refresh()
}

func statusWithEmoji(text string) string {
	switch strings.ToUpper(strings.TrimSpace(text)) {
	case "READY":
		return "⚪ " + text
	case "RUNNING":
		return "▶️ " + text
	case "COMPLETED", "COMPLETE":
		return "✅ " + text
	case "ERROR":
		return "❌ " + text
	case "STOPPED", "CANCELLED":
		return "⏹ " + text
	case "QUEUED":
		return "📋 " + text
	case "PAUSED":
		return "⏸ " + text
	default:
		return text
	}
}

func mainStatusColor(text string) color.Color {
	switch strings.ToUpper(strings.TrimSpace(text)) {
	case "READY":
		return color.NRGBA{R: 0x3c, G: 0x40, B: 0x5a, A: 0xff}
	case "RUNNING":
		return color.NRGBA{R: 0x2f, G: 0x64, B: 0xb8, A: 0xff}
	case "COMPLETED", "COMPLETE":
		return color.NRGBA{R: 0x3d, G: 0x8c, B: 0x4a, A: 0xff}
	case "ERROR":
		return color.NRGBA{R: 0xb8, G: 0x3d, B: 0x4f, A: 0xff}
	case "STOPPED", "CANCELLED":
		return color.NRGBA{R: 0xc6, G: 0x7a, B: 0x2f, A: 0xff}
	case "QUEUED":
		return color.NRGBA{R: 0x6c, G: 0x4a, B: 0xb8, A: 0xff}
	case "PAUSED":
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
	lbl := widget.NewLabel("IDLE")
	lbl.Alignment = fyne.TextAlignCenter
	root := container.NewStack(bg, container.NewPadded(lbl))
	return &PhaseBadge{bg: bg, label: lbl, Root: root}
}

func (p *PhaseBadge) SetPhase(st downloader.Stage) {
	text := "IDLE"
	emoji := "💤"
	if st != downloader.StageUnknown {
		text = string(st)
		switch st {
		case downloader.StageExtracting:
			emoji = "🔎"
		case downloader.StageDownloading:
			emoji = "⬇️"
		case downloader.StagePostProcess:
			emoji = "🛠️"
		default:
			emoji = "•"
		}
	}
	p.label.SetText(emoji + " " + text)
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
