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

func NormalizeUIScale(s float32) float32 {
	switch {
	case s <= 0:
		return UIScaleComfortable
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
	return NormalizeUIScale(t.Scale)
}

func (t *TokyoNightTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x1a, G: 0x1b, B: 0x26, A: 0xff}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xc0, G: 0xca, B: 0xf5, A: 0xff}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x7a, G: 0xa2, B: 0xf7, A: 0xff}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x1f, G: 0x23, B: 0x35, A: 0xff}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x24, G: 0x28, B: 0x3b, A: 0xff}
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 0x3b, G: 0x42, B: 0x61, A: 0xff}
	case theme.ColorNameFocus:
		return color.NRGBA{R: 0x7a, G: 0xa2, B: 0xf7, A: 0x66}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x9e, G: 0xce, B: 0x6a, A: 0xff}
	case theme.ColorNameError:
		return color.NRGBA{R: 0xf7, G: 0x76, B: 0x8e, A: 0xff}
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
		return 12 * s
	case theme.SizeNameInputRadius:
		return 4 * s
	}
	return theme.DefaultTheme().Size(name) * s
}

func ApplyTheme(app fyne.App, scale float32) {
	app.Settings().SetTheme(NewTokyoNightTheme(scale))
}

func ScaledWindowSize(baseW, baseH, scale float32) fyne.Size {
	s := NormalizeUIScale(scale)
	return fyne.NewSize(baseW*s, baseH*s)
}

func ScaledMinWindowSize(scale float32) fyne.Size {
	s := NormalizeUIScale(scale)
	return fyne.NewSize(960*s, 640*s)
}
