package fyneui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"vadlp/internal/core"
	"vadlp/internal/i18n"
	"vadlp/internal/settings"
)

type ProfileBar struct {
	Select *widget.Select
}

func NewProfileBar(
	w fyne.Window,
	cfg *core.Config,
	appSettings *settings.App,
	saveAppSettings func(),
	syncUIFromCfg func(core.Config),
	tr func(string) string,
	bind *LocaleBinder,
) (*ProfileBar, fyne.CanvasObject) {
	bar := &ProfileBar{}
	bar.Select = widget.NewSelect([]string{}, nil)
	var suppressSelect bool

	refresh := func() {
		names, err := core.ListProfiles()
		if err != nil {
			return
		}
		suppressSelect = true
		bar.Select.Options = names
		if appSettings.LastProfile != "" {
			for _, n := range names {
				if n == appSettings.LastProfile {
					bar.Select.SetSelected(n)
					break
				}
			}
		}
		bar.Select.Refresh()
		suppressSelect = false
	}

	loadSelected := func() {
		name := bar.Select.Selected
		if name == "" {
			return
		}
		p, err := core.LoadProfile(name)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		core.ApplyProfileConfig(cfg, p.Config)
		syncUIFromCfg(*cfg)
		appSettings.LastProfile = name
		saveAppSettings()
	}

	bar.Select.OnChanged = func(string) {
		if suppressSelect {
			return
		}
		loadSelected()
	}

	saveAs := func(title string) {
		nameEntry := widget.NewEntry()
		descEntry := widget.NewMultiLineEntry()
		descEntry.SetMinRowsVisible(2)
		if bar.Select.Selected != "" {
			nameEntry.SetText(bar.Select.Selected)
			if p, err := core.LoadProfile(bar.Select.Selected); err == nil {
				descEntry.SetText(p.Description)
			}
		}
		dialog.ShowForm(title, tr("btn.save"), tr("btn.cancel"), []*widget.FormItem{
			widget.NewFormItem(tr("form.profile_name"), nameEntry),
			widget.NewFormItem(tr("form.profile_description"), descEntry),
		}, func(ok bool) {
			if !ok {
				return
			}
			name := strings.TrimSpace(nameEntry.Text)
			if err := core.ValidateProfileName(name); err != nil {
				dialog.ShowError(err, w)
				return
			}
			p := core.Profile{
				Name:        name,
				Description: strings.TrimSpace(descEntry.Text),
				Config:      *cfg,
			}
			if err := core.SaveProfile(p); err != nil {
				dialog.ShowError(err, w)
				return
			}
			refresh()
			bar.Select.SetSelected(name)
			appSettings.LastProfile = name
			saveAppSettings()
		}, w)
	}

	saveBtn := widget.NewButton(tr("btn.save_profile"), func() {
		if bar.Select.Selected == "" {
			saveAs(tr("dialog.save_profile"))
			return
		}
		p := core.Profile{
			Name:        bar.Select.Selected,
			Config:      *cfg,
			Description: "",
		}
		if existing, err := core.LoadProfile(bar.Select.Selected); err == nil {
			p.Description = existing.Description
		}
		if err := core.SaveProfile(p); err != nil {
			dialog.ShowError(err, w)
			return
		}
		appSettings.LastProfile = bar.Select.Selected
		saveAppSettings()
		dialog.ShowInformation(tr("card.profiles"), tr("msg.profile_saved"), w)
	})

	saveAsBtn := widget.NewButton(tr("btn.save_profile_as"), func() {
		saveAs(tr("dialog.save_profile_as"))
	})

	newBtn := widget.NewButton(tr("btn.new_profile"), func() {
		def := core.DefaultConfig()
		core.ApplyProfileConfig(cfg, def)
		syncUIFromCfg(*cfg)
		bar.Select.ClearSelected()
		appSettings.LastProfile = ""
		saveAppSettings()
	})

	deleteBtn := widget.NewButton(tr("btn.delete_profile"), func() {
		name := bar.Select.Selected
		if name == "" {
			return
		}
		dialog.ShowConfirm(tr("btn.delete_profile"), i18n.T("msg.delete_profile_confirm", map[string]interface{}{"Name": name}),
			func(ok bool) {
				if !ok {
					return
				}
				if err := core.DeleteProfile(name); err != nil {
					dialog.ShowError(err, w)
					return
				}
				if appSettings.LastProfile == name {
					appSettings.LastProfile = ""
					saveAppSettings()
				}
				refresh()
			}, w)
	})

	renameBtn := widget.NewButton(tr("btn.rename_profile"), func() {
		old := bar.Select.Selected
		if old == "" {
			return
		}
		nameEntry := widget.NewEntry()
		nameEntry.SetText(old)
		dialog.ShowForm(tr("btn.rename_profile"), tr("btn.save"), tr("btn.cancel"), []*widget.FormItem{
			widget.NewFormItem(tr("form.profile_name"), nameEntry),
		}, func(ok bool) {
			if !ok {
				return
			}
			newName := strings.TrimSpace(nameEntry.Text)
			if err := core.RenameProfile(old, newName); err != nil {
				dialog.ShowError(err, w)
				return
			}
			if appSettings.LastProfile == old {
				appSettings.LastProfile = newName
				saveAppSettings()
			}
			refresh()
			bar.Select.SetSelected(newName)
		}, w)
	})

	refresh()
	if appSettings.LastProfile != "" {
		if p, err := core.LoadProfile(appSettings.LastProfile); err == nil {
			core.ApplyProfileConfig(cfg, p.Config)
			syncUIFromCfg(*cfg)
		}
	}

	fiProfile := widget.NewFormItem(tr("form.saved_profile"), bar.Select)
	body := container.NewVBox(
		widget.NewForm(fiProfile),
		container.NewHBox(saveBtn, saveAsBtn, newBtn),
		container.NewHBox(renameBtn, deleteBtn),
	)

	section := Section(tr("card.profiles"), tr("card.profiles_hint"), body)

	if bind != nil {
		bind.BindSection(section, "card.profiles", "card.profiles_hint", tr)
		bind.BindFormItem(fiProfile, "form.saved_profile", tr)
		bind.BindButton(saveBtn, "btn.save_profile", tr)
		bind.BindButton(saveAsBtn, "btn.save_profile_as", tr)
		bind.BindButton(newBtn, "btn.new_profile", tr)
		bind.BindButton(deleteBtn, "btn.delete_profile", tr)
		bind.BindButton(renameBtn, "btn.rename_profile", tr)
	}

	return bar, section.Root
}
