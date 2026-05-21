package fyneui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func newDropEntry() *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder("https://…")
	return e
}

func setupURLDrop(w fyne.Window, apply func(text string)) {
	w.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		for _, u := range uris {
			s := strings.TrimSpace(u.String())
			if s == "" {
				continue
			}
			if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
				apply(s)
				return
			}
		}
		if len(uris) > 0 {
			apply(strings.TrimSpace(uris[0].String()))
		}
	})
}
