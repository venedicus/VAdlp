package fyneui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/core"
)

type DownloadFormatUI struct {
	QualityEntry      *widget.Entry
	QualityPreset     *widget.Select
	FormatSelect      *widget.Select
	FormatCustomEntry *widget.Entry
	AudioCheck        *widget.Check
	AudioFormatSelect *widget.Select
}

func NewDownloadFormatUI(
	cfg *core.Config,
	updatePreview func(),
	tr func(string) string,
) (*DownloadFormatUI, fyne.CanvasObject) {
	ui := &DownloadFormatUI{}

	ui.QualityEntry = widget.NewEntry()
	ui.QualityEntry.SetPlaceHolder(tr("placeholder.quality"))
	ui.QualityEntry.SetText(cfg.Quality)
	ui.QualityEntry.OnChanged = func(s string) {
		cfg.Quality = s
		updatePreview()
	}

	presetLabels := make([]string, len(core.QualityPresets))
	labelToValue := make(map[string]string, len(core.QualityPresets))
	for i, p := range core.QualityPresets {
		label := tr(p.Key)
		presetLabels[i] = label
		labelToValue[label] = p.Value
	}
	ui.QualityPreset = widget.NewSelect(presetLabels, func(label string) {
		if v, ok := labelToValue[label]; ok {
			ui.QualityEntry.SetText(v)
			cfg.Quality = v
			updatePreview()
		}
	})

	ui.FormatCustomEntry = widget.NewEntry()
	ui.FormatCustomEntry.SetPlaceHolder(tr("form.container_custom"))
	ui.FormatCustomEntry.OnChanged = func(s string) {
		if s == "" {
			return
		}
		cfg.Format = s
		found := false
		for _, o := range ui.FormatSelect.Options {
			if o == s {
				found = true
				break
			}
		}
		if !found {
			ui.FormatSelect.Options = append(ui.FormatSelect.Options, s)
		}
		ui.FormatSelect.SetSelected(s)
		updatePreview()
	}

	ui.FormatSelect = widget.NewSelect(core.MergeFormats, func(s string) {
		cfg.Format = s
		if s != "" {
			ui.FormatCustomEntry.SetText("")
		}
		updatePreview()
	})

	ui.AudioCheck = widget.NewCheck(tr("form.audio_only"), func(b bool) {
		cfg.AudioOnly = b
		updatePreview()
	})
	ui.AudioCheck.SetChecked(cfg.AudioOnly)

	ui.AudioFormatSelect = widget.NewSelect(
		[]string{"", "mp3", "m4a", "opus", "wav", "flac", "vorbis", "aac", "alac"},
		func(s string) {
			cfg.AudioFormat = s
			updatePreview()
		},
	)
	ui.AudioFormatSelect.SetSelected(cfg.AudioFormat)
	ui.FormatSelect.SetSelected(cfg.Format)

	presetBtns := make([]fyne.CanvasObject, 0, len(core.PresetIDs))
	for _, id := range core.PresetIDs {
		presetID := id
		presetBtns = append(presetBtns, widget.NewButton(tr("preset."+presetID), func() {
			var c core.Config
			c = *cfg
			if core.ApplyPreset(&c, presetID) {
				ui.SyncFromCfg(c)
				*cfg = c
				updatePreview()
			}
		}))
	}

	formatForm := widget.NewForm(
		widget.NewFormItem(tr("form.quality_preset"), ui.QualityPreset),
		widget.NewFormItem(tr("form.quality"), ui.QualityEntry),
		widget.NewFormItem(tr("form.container"), ui.FormatSelect),
		widget.NewFormItem(tr("form.container_custom"), ui.FormatCustomEntry),
		widget.NewFormItem("", ui.AudioCheck),
		widget.NewFormItem(tr("form.audio_format"), ui.AudioFormatSelect),
	)

	card := widget.NewCard(tr("card.format"), "",
		container.NewVBox(
			formatForm,
			widget.NewSeparator(),
			widget.NewLabel(tr("form.quick_presets")),
			container.NewGridWithColumns(3, presetBtns...),
		),
	)
	return ui, card
}

func (ui *DownloadFormatUI) SyncFromCfg(c core.Config) {
	ui.QualityEntry.SetText(c.Quality)
	ui.FormatSelect.SetSelected(c.Format)
	if c.Format != "" {
		found := false
		for _, o := range ui.FormatSelect.Options {
			if o == c.Format {
				found = true
				break
			}
		}
		if !found {
			ui.FormatSelect.Options = append(ui.FormatSelect.Options, c.Format)
		}
	}
	ui.FormatCustomEntry.SetText("")
	ui.AudioCheck.SetChecked(c.AudioOnly)
	ui.AudioFormatSelect.SetSelected(c.AudioFormat)
}
