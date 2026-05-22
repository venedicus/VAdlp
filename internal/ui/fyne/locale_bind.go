package fyneui

import (
	"sync/atomic"

	"fyne.io/fyne/v2/widget"
)

type LocaleBinder struct {
	refresh    []func()
	refreshing atomic.Bool
}

func NewLocaleBinder() *LocaleBinder {
	return &LocaleBinder{}
}

func (b *LocaleBinder) Add(fn func()) {
	if fn != nil {
		b.refresh = append(b.refresh, fn)
	}
}

func (b *LocaleBinder) Refresh() {
	if b == nil {
		return
	}
	if !b.refreshing.CompareAndSwap(false, true) {
		return
	}
	defer b.refreshing.Store(false)
	for _, fn := range b.refresh {
		fn()
	}
}

// RefreshDeferred runs Refresh on the UI thread after the current event handler returns.
func (b *LocaleBinder) RefreshDeferred() {
	uiExec(func() { b.Refresh() })
}

func (b *LocaleBinder) BindLabel(lbl *widget.Label, key string, tr func(string) string) {
	if lbl == nil || key == "" {
		return
	}
	b.Add(func() { lbl.SetText(tr(key)) })
}

func (b *LocaleBinder) BindCheck(ch *widget.Check, key string, tr func(string) string) {
	if ch == nil || key == "" {
		return
	}
	b.Add(func() { ch.SetText(tr(key)) })
}

func (b *LocaleBinder) BindButton(btn *widget.Button, key string, tr func(string) string) {
	if btn == nil || key == "" {
		return
	}
	b.Add(func() { btn.SetText(tr(key)) })
}

func (b *LocaleBinder) BindPlaceholder(entry *widget.Entry, key string, tr func(string) string) {
	if entry == nil || key == "" {
		return
	}
	b.Add(func() { entry.SetPlaceHolder(tr(key)) })
}

func (b *LocaleBinder) BindFormItem(item *widget.FormItem, key string, tr func(string) string) {
	if item == nil || key == "" {
		return
	}
	b.Add(func() { item.Text = tr(key) })
}

func (b *LocaleBinder) BindSection(sec SectionMeta, titleKey, hintKey string, tr func(string) string) {
	if sec.Title == nil || titleKey == "" {
		return
	}
	b.Add(func() {
		sec.Title.SetText(tr(titleKey))
		if sec.Hint != nil && hintKey != "" {
			sec.Hint.SetText(tr(hintKey))
		}
	})
}
