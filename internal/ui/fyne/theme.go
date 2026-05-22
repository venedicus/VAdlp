package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	UIScaleCompact     float32 = 1.0
	UIScaleComfortable float32 = 1.15
	UIScaleLarge       float32 = 1.28
	UIScaleExtraLarge  float32 = 1.42
)

func IsAutoUIScale(s float32) bool {
	return s <= 0 || s == UIScaleAuto
}

func NormalizeUIScale(s float32) float32 {
	switch {
	case IsAutoUIScale(s):
		return UIScaleAuto
	case s < 1.06:
		return UIScaleCompact
	case s < 1.21:
		return UIScaleComfortable
	case s < 1.35:
		return UIScaleLarge
	default:
		return UIScaleExtraLarge
	}
}

type TokyoNightTheme struct {
	Scale float32
}

func NewTokyoNightTheme(scale float32) fyne.Theme {
	return &TokyoNightTheme{Scale: NormalizeUIScale(scale)}
}

func (t *TokyoNightTheme) scale() float32 {
	s := NormalizeUIScale(t.Scale)
	if IsAutoUIScale(s) {
		return UIScaleCompact
	}
	return s
}

func (t *TokyoNightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x1a, G: 0x1b, B: 0x26, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xc0, G: 0xca, B: 0xf5, A: 0xff}
	case theme.ColorNameDisabled:
		// Read-only logs and previews use the disabled palette; keep high contrast.
		return color.NRGBA{R: 0xc0, G: 0xca, B: 0xf5, A: 0xff}
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 0x56, G: 0x5f, B: 0x89, A: 0xff}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x7a, G: 0xa2, B: 0xf7, A: 0xff}
	case theme.ColorNameHyperlink:
		return color.NRGBA{R: 0x7d, G: 0xcf, B: 0xff, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x16, G: 0x19, B: 0x28, A: 0xff}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x24, G: 0x28, B: 0x3b, A: 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x3b, G: 0x42, B: 0x61, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x7a, G: 0xa2, B: 0xf7, A: 0x66}
	case theme.ColorNameSelection:
		return color.NRGBA{R: 0x33, G: 0x46, B: 0x7c, A: 0xff}
	case theme.ColorNameSeparator:
		return color.NRGBA{R: 0x29, G: 0x2e, B: 0x42, A: 0xff}
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 0x3b, G: 0x40, B: 0x5c, A: 0xff}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x9e, G: 0xce, B: 0x6a, A: 0xff}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xf7, G: 0x76, B: 0x8e, A: 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xe0, G: 0xaf, B: 0x68, A: 0xff}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (t *TokyoNightTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *TokyoNightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *TokyoNightTheme) Size(name fyne.ThemeSizeName) float32 {
	s := t.scale()
	switch name {
	case theme.SizeNamePadding:
		return 6 * s
	case theme.SizeNameInnerPadding:
		return 4 * s
	case theme.SizeNameInlineIcon:
		return 18 * s
	case theme.SizeNameInputBorder:
		return 1 * s
	case theme.SizeNameText:
		return 14 * s
	case theme.SizeNameHeadingText:
		return 18 * s
	case theme.SizeNameSubHeadingText:
		return 16 * s
	case theme.SizeNameCaptionText:
		return 13 * s
	case theme.SizeNameInputRadius:
		return 4 * s
	}
	return theme.DefaultTheme().Size(name) * s
}

func ApplyTheme(app fyne.App, scale float32) {
	app.Settings().SetTheme(NewTokyoNightTheme(scale))
}

func ScaledMinWindowSize(scale float32) fyne.Size {
	s := NormalizeUIScale(scale)
	return fyne.NewSize(960*s, 640*s)
}
