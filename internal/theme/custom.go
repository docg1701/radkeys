// Package theme — custom Fyne theme that derives every ThemeColorName from
// the preset's three colors (Background, Button, Fixed). Light/dark variant is
// auto-detected from background luminance.
//
// ZERO delegation to theme.DefaultTheme() — every color is explicit so theme
// switches leave no residue from the previous theme.
//
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

// RadKeysTheme derives every Fyne color from a preset.
type RadKeysTheme struct {
	bg  color.NRGBA
	btn color.NRGBA
	fix color.NRGBA
	fg  color.NRGBA // cached foreground
}

// NewCustomTheme returns a fyne.Theme for preset p.
func NewCustomTheme(p Preset) fyne.Theme {
	if p.ID == "system" {
		return theme.DefaultTheme()
	}
	bg := parseHex(p.Background)
	btn := parseHex(p.Button)
	fix := parseHex(p.Fixed)
	fg := color.NRGBA{R: 0xc8, G: 0xc8, B: 0xc8, A: 0xff} // muted dark fg
	if isLightNRGBA(bg) {
		fg = color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xff} // muted light fg
	}
	return &RadKeysTheme{bg: bg, btn: btn, fix: fix, fg: fg}
}

// Color returns an explicit derivation for every ThemeColorName.
// No fallback to DefaultTheme.
func (t *RadKeysTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {

	// ── structural (page / surfaces) ──────────────────────────
	case theme.ColorNameBackground:
		return t.bg
	case theme.ColorNameButton:
		return t.btn
	case theme.ColorNameHeaderBackground:
		return t.headerBg()
	case theme.ColorNameMenuBackground:
		return t.btn
	case theme.ColorNameOverlayBackground:
		return setAlpha(t.fg, 0xCC)

	// ── input fields ──────────────────────────────────────────
	case theme.ColorNameInputBackground:
		return t.inputBg()
	case theme.ColorNameInputBorder:
		return blend(t.bg, t.fg, 0.20)

	// ── text ──────────────────────────────────────────────────
	case theme.ColorNameForeground:
		return t.fg
	case theme.ColorNameDisabled:
		return blend(t.bg, t.fg, 0.38)
	case theme.ColorNamePlaceHolder:
		return blend(t.bg, t.fg, 0.42)
	case theme.ColorNameHyperlink:
		return t.fg

	// ── interactive states ────────────────────────────────────
	case theme.ColorNameHover:
		return t.hover()
	case theme.ColorNamePressed:
		return darken(t.btn, 0.12)
	case theme.ColorNameFocus:
		return setAlpha(t.fix, 0x5c)
	case theme.ColorNameSelection:
		return setAlpha(t.fix, 0x40)
	case theme.ColorNameDisabledButton:
		return blend(t.btn, t.bg, 0.45)

	// ── decoration ────────────────────────────────────────────
	case theme.ColorNamePrimary:
		return t.primary()
	case theme.ColorNameScrollBar:
		return blend(t.bg, t.fg, 0.14)
	case theme.ColorNameScrollBarBackground:
		return blend(t.bg, t.fg, 0.05)
	case theme.ColorNameSeparator:
		return blend(t.bg, t.fg, 0.12)
	case theme.ColorNameShadow:
		return setAlpha(t.fg, 0x10)

	// ── foreground-on-X (readability guaranteed) ──────────────
	case theme.ColorNameForegroundOnPrimary:
		return t.fgOnPrimary()
	case theme.ColorNameForegroundOnError:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameForegroundOnSuccess:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameForegroundOnWarning:
		return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}

	// ── semantic (standard colors, NOT derived from preset) ───
	case theme.ColorNameError:
		return color.NRGBA{R: 0xd3, G: 0x2f, B: 0x2f, A: 0xff}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x38, G: 0x8e, B: 0x3c, A: 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xf5, G: 0x7c, B: 0x00, A: 0xff}
	}

	// Safety net — should never be reached since all ThemeColorName are handled.
	// Return bg so any future Fyne additions render visibly rather than panic.
	return t.bg
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

// ---------------------------------------------------------------------------
// adaptive helpers
// ---------------------------------------------------------------------------

// headerBg returns a colour for the tab bar / header area.
// Dark theme: much lighter than bg so tab bar clearly stands out.
// Light theme: much darker than bg so tab bar clearly stands out.
func (t *RadKeysTheme) headerBg() color.NRGBA {
	if isLightNRGBA(t.bg) {
		return darken(t.bg, 0.12)
	}
	return lighten(t.bg, 0.20)
}

// inputBg returns a colour slightly distinct from bg so Entry/Select fields
// are visually separated from the page background.
func (t *RadKeysTheme) inputBg() color.NRGBA {
	if isLightNRGBA(t.bg) {
		// Light theme: inputs slightly lighter than bg.
		return lighten(t.bg, 0.03)
	}
	// Dark theme: inputs slightly brighter than bg.
	return lighten(t.bg, 0.06)
}

// hover moves btn toward the foreground direction so the hovered element
// gains contrast against its neighbours.
func (t *RadKeysTheme) hover() color.NRGBA {
	if isLightNRGBA(t.bg) {
		return blend(t.btn, t.fg, 0.10)
	}
	return lighten(t.btn, 0.08)
}

// primary returns the accent colour boosted for contrast — used both for
// the tab indicator underline and for selected tab text.
// Maximum contrast: pure white on dark, pure black on light.
func (t *RadKeysTheme) primary() color.NRGBA {
	if isLightNRGBA(t.bg) {
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	}
	return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
}

// fgOnPrimary returns the opposite of the primary colour so text on
// primary buttons is always readable.
func (t *RadKeysTheme) fgOnPrimary() color.NRGBA {
	if isLightNRGBA(t.bg) {
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	}
	return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
}

// fgOnAccent returns white or black depending on the fix (accent) luminance,
// ensuring text on primary buttons is always readable.
func (t *RadKeysTheme) fgOnAccent() color.NRGBA {
	if isLightNRGBA(t.fix) {
		return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}
	}
	return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
}

// ---------------------------------------------------------------------------
// colour operations
// ---------------------------------------------------------------------------

func lighten(c color.NRGBA, factor float64) color.NRGBA {
	d := uint8(255 * factor)
	return color.NRGBA{
		R: satAdd(c.R, d), G: satAdd(c.G, d), B: satAdd(c.B, d), A: c.A,
	}
}

func darken(c color.NRGBA, factor float64) color.NRGBA {
	d := uint8(255 * factor)
	return color.NRGBA{
		R: satSub(c.R, d), G: satSub(c.G, d), B: satSub(c.B, d), A: c.A,
	}
}

func blend(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: lerp(a.R, b.R, t), G: lerp(a.G, b.G, t), B: lerp(a.B, b.B, t), A: a.A,
	}
}

func setAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

func lerp(a, b uint8, t float64) uint8 {
	return uint8(float64(a)*(1-t) + float64(b)*t)
}

func satAdd(a, b uint8) uint8 {
	if s := uint16(a) + uint16(b); s <= 255 {
		return uint8(s)
	}
	return 255
}

func satSub(a, b uint8) uint8 {
	if b > a {
		return 0
	}
	return a - b
}

// ---------------------------------------------------------------------------
// luminance helpers
// ---------------------------------------------------------------------------

func isLightNRGBA(c color.NRGBA) bool {
	return 0.2126*float64(c.R)/255.0+
		0.7152*float64(c.G)/255.0+
		0.0722*float64(c.B)/255.0 > 0.45
}

func parseHex(s string) color.NRGBA {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}
	}
	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xff}
}
