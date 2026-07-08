// Package theme — custom Fyne theme that applies preset colors to the entire UI.
// Follows the Catppuccin theme pattern: every Fyne color name is resolved from
// the preset's Background, Button, and Fixed colors. No hardcoded values.
// "Padrão do sistema" delegates entirely to theme.DefaultTheme().
package theme

import (
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Theme = (*RadKeysTheme)(nil)

// RadKeysTheme applies preset colors to all Fyne theme color names.
type RadKeysTheme struct {
	bg      color.NRGBA // preset Background
	btn     color.NRGBA // preset Button
	fix     color.NRGBA // preset Fixed (accent)
	variant fyne.ThemeVariant
}

// NewCustomTheme creates a fyne.Theme from a RadKeys preset.
func NewCustomTheme(p Preset) fyne.Theme {
	if p.Name == "Padrão do sistema" {
		return theme.DefaultTheme()
	}
	v := theme.VariantDark
	if isLight(p) {
		v = theme.VariantLight
	}
	return &RadKeysTheme{
		bg:      parseHex(p.Background, 0x00, 0x00, 0x00),
		btn:     parseHex(p.Button, 0x00, 0x00, 0x00),
		fix:     parseHex(p.Fixed, 0x00, 0x00, 0x00),
		variant: v,
	}
}

func (t *RadKeysTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return t.bg
	case theme.ColorNameButton:
		return t.btn
	case theme.ColorNameDisabledButton:
		return darken(t.btn, 0.25)
	case theme.ColorNameDisabled:
		return blend(t.bg, t.fg(), 0.35)
	case theme.ColorNameFocus:
		return setAlpha(t.fix, 0x50)
	case theme.ColorNameForeground:
		return t.fg()
	case theme.ColorNameHover:
		return lighten(t.btn, 0.10)
	case theme.ColorNameInputBackground:
		return lighten(t.bg, 0.06)
	case theme.ColorNameInputBorder:
		return blend(t.bg, t.fg(), 0.15)
	case theme.ColorNamePlaceHolder:
		return blend(t.bg, t.fg(), 0.40)
	case theme.ColorNamePressed:
		return darken(t.btn, 0.12)
	case theme.ColorNamePrimary:
		return t.fix
	case theme.ColorNameScrollBar:
		return blend(t.bg, t.fg(), 0.12)
	case theme.ColorNameSelection:
		return setAlpha(t.fix, 0x40)
	case theme.ColorNameSeparator:
		return blend(t.bg, t.fg(), 0.10)
	case theme.ColorNameShadow:
		return setAlpha(t.fg(), 0x12)
	default:
		return theme.DefaultTheme().Color(name, t.variant)
	}
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

// fg returns white or near-black based on the background luminance, not on preset name.
func (t *RadKeysTheme) fg() color.NRGBA {
	if t.variant == theme.VariantLight {
		return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xFF}
	}
	return color.NRGBA{R: 0xe8, G: 0xe8, B: 0xe8, A: 0xFF}
}

func isLight(p Preset) bool {
	bg := parseHex(p.Background, 0, 0, 0)
	return 0.2126*float64(bg.R)/255.0+
		0.7152*float64(bg.G)/255.0+
		0.0722*float64(bg.B)/255.0 > 0.45
}

func lighten(c color.NRGBA, factor float64) color.NRGBA {
	return color.NRGBA{
		R: clamp(c.R + uint8(255*factor)),
		G: clamp(c.G + uint8(255*factor)),
		B: clamp(c.B + uint8(255*factor)),
		A: c.A,
	}
}

func darken(c color.NRGBA, factor float64) color.NRGBA {
	d := uint8(255 * factor)
	return color.NRGBA{
		R: satSub(c.R, d),
		G: satSub(c.G, d),
		B: satSub(c.B, d),
		A: c.A,
	}
}

func blend(bg, fg color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: uint8(float64(bg.R)*(1-t) + float64(fg.R)*t),
		G: uint8(float64(bg.G)*(1-t) + float64(fg.G)*t),
		B: uint8(float64(bg.B)*(1-t) + float64(fg.B)*t),
		A: bg.A,
	}
}

func setAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

func parseHex(s string, dr, dg, db uint8) color.NRGBA {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.NRGBA{R: dr, G: dg, B: db, A: 0xFF}
	}
	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}
}

func clamp(v uint8) uint8 {
	if v > 255 {
		return 255
	}
	return v
}

func satSub(a, b uint8) uint8 {
	if b > a {
		return 0
	}
	return a - b
}
