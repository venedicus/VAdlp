package fyneui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type TokyoNightTheme struct{}

func NewTokyoNightTheme() fyne.Theme {
	return &TokyoNightTheme{}
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
	// Keep UI labels readable; use monospace only where widgets set TextStyle.Monospace (log, command).
	return theme.DefaultTheme().Font(style)
}

func (t *TokyoNightTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *TokyoNightTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 16
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameInputBorder:
		return 1
	case theme.SizeNameText:
		return 13
	}
	return theme.DefaultTheme().Size(name)
}