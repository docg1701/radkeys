// Package theme — custom Fyne theme with explicit colours per preset.
// Each named colour has a hardcoded value; missing colours fall back to
// theme.DefaultTheme() with the correct variant. No lighten/darken/blend.
//
// "Padrão do sistema" delegates entirely to theme.DefaultTheme().
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Theme = (*RadKeysTheme)(nil)

// PresetColours maps ThemeColorName → colour for a single variant.
// nil = fall back to DefaultTheme for that colour.
type PresetColours map[fyne.ThemeColorName]color.NRGBA

// RadKeysTheme looks up colours from a preset; missing ones go to DefaultTheme.
type RadKeysTheme struct {
	light PresetColours
	dark  PresetColours
}

func (t *RadKeysTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	colours := t.dark
	if v == theme.VariantLight {
		colours = t.light
	}
	if c, ok := colours[name]; ok {
		return c
	}
	return theme.DefaultTheme().Color(name, v)
}

func (t *RadKeysTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *RadKeysTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *RadKeysTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
