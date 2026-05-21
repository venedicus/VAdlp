package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type SectionMeta struct {
	Root  fyne.CanvasObject
	Title *widget.Label
	Hint  *widget.Label
}

func Section(title, hint string, body fyne.CanvasObject) SectionMeta {
	titleLbl := widget.NewLabel(title)
	titleLbl.TextStyle = fyne.TextStyle{Bold: true}
	parts := []fyne.CanvasObject{titleLbl}
	var hintLbl *widget.Label
	if hint != "" {
		hintLbl = widget.NewLabel(hint)
		hintLbl.Wrapping = fyne.TextWrapWord
		parts = append(parts, hintLbl)
	}
	parts = append(parts, body)
	return SectionMeta{
		Root:  container.NewVBox(parts...),
		Title: titleLbl,
		Hint:  hintLbl,
	}
}

func ScrollTab(content fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVScroll(container.NewPadded(content))
}

func Toolbar(groups ...fyne.CanvasObject) fyne.CanvasObject {
	rows := make([]fyne.CanvasObject, 0, len(groups))
	for _, g := range groups {
		if g == nil {
			continue
		}
		rows = append(rows, g)
	}
	return container.NewVBox(rows...)
}
