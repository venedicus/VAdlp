package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/downloader"
	"vadlp/internal/i18n"
)

type StatusBadge struct {
	bg      *canvas.Rectangle
	label   *widget.Label
	Root    fyne.CanvasObject
	key     string
	compact bool
}

func NewStatusBadge(key string) *StatusBadge {
	bg := canvas.NewRectangle(StatusColor(key))
	bg.CornerRadius = 4
	lbl := widget.NewLabel(LocalizedStatus(key, false))
	lbl.Alignment = fyne.TextAlignCenter
	root := container.NewStack(bg, container.NewPadded(lbl))
	return &StatusBadge{bg: bg, label: lbl, Root: root, key: key}
}

func (s *StatusBadge) SetStatusKey(key string) {
	s.key = key
	s.label.SetText(LocalizedStatus(key, s.compact))
	s.bg.FillColor = StatusColor(key)
	s.bg.Refresh()
}

func (s *StatusBadge) RefreshText() {
	s.SetStatusKey(s.key)
}

func (s *StatusBadge) SetCompact(compact bool) {
	if s == nil || s.compact == compact {
		return
	}
	s.compact = compact
	s.SetStatusKey(s.key)
}

type PhaseBadge struct {
	bg      *canvas.Rectangle
	label   *widget.Label
	Root    fyne.CanvasObject
	stage   downloader.Stage
	compact bool
}

func NewPhaseBadge() *PhaseBadge {
	bg := canvas.NewRectangle(PhaseColor(downloader.StageUnknown))
	bg.CornerRadius = 4
	lbl := widget.NewLabel(phaseLabel(downloader.StageUnknown, false))
	lbl.Alignment = fyne.TextAlignCenter
	root := container.NewStack(bg, container.NewPadded(lbl))
	return &PhaseBadge{bg: bg, label: lbl, Root: root}
}

func phaseLabel(st downloader.Stage, compact bool) string {
	var text string
	switch st {
	case downloader.StageExtracting:
		text = i18n.T("phase.extracting", nil)
	case downloader.StageDownloading:
		text = i18n.T("phase.downloading", nil)
	case downloader.StagePostProcess:
		text = i18n.T("phase.postprocess", nil)
	default:
		text = i18n.T("phase.idle", nil)
	}
	if compact && len([]rune(text)) > 5 {
		r := []rune(text)
		return string(r[:5])
	}
	return text
}

func (p *PhaseBadge) SetPhase(st downloader.Stage) {
	p.stage = st
	p.label.SetText(phaseLabel(st, p.compact))
	p.bg.FillColor = PhaseColor(st)
	p.bg.Refresh()
}

func (p *PhaseBadge) RefreshText() {
	p.SetPhase(p.stage)
}

func (p *PhaseBadge) SetCompact(compact bool) {
	if p == nil || p.compact == compact {
		return
	}
	p.compact = compact
	p.SetPhase(p.stage)
}
