package fyneui

import (
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
	key   string
}

func NewStatusBadge(key string) *StatusBadge {
	bg := canvas.NewRectangle(StatusColor(key))
	bg.CornerRadius = 4
	lbl := widget.NewLabel(statusLabel(key))
	lbl.Alignment = fyne.TextAlignCenter
	root := container.NewStack(bg, container.NewPadded(lbl))
	return &StatusBadge{bg: bg, label: lbl, Root: root, key: key}
}

func statusLabel(key string) string {
	k := strings.ToLower(strings.TrimSpace(key))
	if k == "" {
		k = "ready"
	}
	return strings.ToUpper(i18n.T("status."+k, nil))
}

func (s *StatusBadge) SetStatusKey(key string) {
	s.key = key
	s.label.SetText(statusLabel(key))
	s.bg.FillColor = StatusColor(key)
	s.bg.Refresh()
}

func (s *StatusBadge) RefreshText() {
	s.SetStatusKey(s.key)
}

type PhaseBadge struct {
	bg    *canvas.Rectangle
	label *widget.Label
	Root  fyne.CanvasObject
	stage downloader.Stage
}

func NewPhaseBadge() *PhaseBadge {
	bg := canvas.NewRectangle(PhaseColor(downloader.StageUnknown))
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
	p.stage = st
	p.label.SetText(phaseLabel(st))
	p.bg.FillColor = PhaseColor(st)
	p.bg.Refresh()
}

func (p *PhaseBadge) RefreshText() {
	p.SetPhase(p.stage)
}
