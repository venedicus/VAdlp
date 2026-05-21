package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type queueRow struct {
	root   *fyne.Container
	dot    *canvas.Rectangle
	name   *widget.Label
	status *widget.Label
	cancel *widget.Button
}

func newQueueRow() *queueRow {
	dot := canvas.NewRectangle(StatusColor("queued"))
	dot.SetMinSize(fyne.NewSize(8, 8))
	dot.CornerRadius = 4
	name := widget.NewLabel("")
	name.Truncation = fyne.TextTruncateEllipsis
	status := widget.NewLabel("")
	status.Alignment = fyne.TextAlignTrailing
	cancel := widget.NewButton("×", nil)
	cancel.Hide()
	root := container.NewBorder(
		nil, nil,
		container.NewCenter(dot),
		cancel,
		container.NewHBox(name, layout.NewSpacer(), status),
	)
	return &queueRow{root: root, dot: dot, name: name, status: status, cancel: cancel}
}

func (r *queueRow) object() fyne.CanvasObject {
	return r.root
}
