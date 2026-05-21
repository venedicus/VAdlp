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
	presetIDs         []string
	presetBtns        []*widget.Button
	presetsLabel      *widget.Label
}

func NewDownloadFormatUI(
	cfg *core.Config,
	updatePreview func(),
	tr func(string) string,
	bind *LocaleBinder,
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

	ui.presetIDs = append([]string(nil), core.PresetIDs...)
	ui.presetBtns = make([]*widget.Button, 0, len(core.PresetIDs))
	for _, id := range core.PresetIDs {
		presetID := id
		btn := widget.NewButton(tr("preset."+presetID), func() {
			c := *cfg
			if core.ApplyPreset(&c, presetID) {
				ui.SyncFromCfg(c)
				*cfg = c
				updatePreview()
			}
		})
		ui.presetBtns = append(ui.presetBtns, btn)
	}

	ui.presetsLabel = widget.NewLabel(tr("form.quick_presets"))
	ui.presetsLabel.TextStyle = fyne.TextStyle{Bold: true}

	fiPreset := widget.NewFormItem(tr("form.quality_preset"), ui.QualityPreset)
	fiQuality := widget.NewFormItem(tr("form.quality"), ui.QualityEntry)
	fiContainer := widget.NewFormItem(tr("form.container"), ui.FormatSelect)
	fiCustom := widget.NewFormItem(tr("form.container_custom"), ui.FormatCustomEntry)
	fiAudioFmt := widget.NewFormItem(tr("form.audio_format"), ui.AudioFormatSelect)
	formatForm := widget.NewForm(
		fiPreset, fiQuality, fiContainer, fiCustom,
		widget.NewFormItem("", ui.AudioCheck),
		fiAudioFmt,
	)

	body := container.NewVBox(
		formatForm,
		ui.presetsLabel,
		container.NewGridWithColumns(3, objectsFromButtons(ui.presetBtns)...),
	)

	section := Section(tr("card.format"), "", body)

	if bind != nil {
		bind.Add(func() { ui.refreshTr(tr) })
		bind.BindSection(section, "card.format", "", tr)
		bind.BindFormItem(fiPreset, "form.quality_preset", tr)
		bind.BindFormItem(fiQuality, "form.quality", tr)
		bind.BindFormItem(fiContainer, "form.container", tr)
		bind.BindFormItem(fiCustom, "form.container_custom", tr)
		bind.BindFormItem(fiAudioFmt, "form.audio_format", tr)
	}

	return ui, section.Root
}

func objectsFromButtons(btns []*widget.Button) []fyne.CanvasObject {
	out := make([]fyne.CanvasObject, len(btns))
	for i, b := range btns {
		out[i] = b
	}
	return out
}

func (ui *DownloadFormatUI) refreshTr(tr func(string) string) {
	ui.QualityEntry.SetPlaceHolder(tr("placeholder.quality"))
	ui.FormatCustomEntry.SetPlaceHolder(tr("form.container_custom"))
	ui.AudioCheck.SetText(tr("form.audio_only"))
	ui.presetsLabel.SetText(tr("form.quick_presets"))

	labels := make([]string, len(core.QualityPresets))
	labelToValue := make(map[string]string, len(core.QualityPresets))
	for i, p := range core.QualityPresets {
		label := tr(p.Key)
		labels[i] = label
		labelToValue[label] = p.Value
	}
	ui.QualityPreset.Options = labels
	ui.QualityPreset.Refresh()

	for i, id := range ui.presetIDs {
		if i < len(ui.presetBtns) {
			ui.presetBtns[i].SetText(tr("preset." + id))
		}
	}
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
