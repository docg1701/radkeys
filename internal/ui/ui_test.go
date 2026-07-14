package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	fyneTheme "fyne.io/fyne/v2/theme"

	"github.com/docg1701/radkeys/internal/theme"
)

type dummyColorTheme struct {
	bg color.NRGBA
}

func (t *dummyColorTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if name == fyneTheme.ColorNameBackground {
		return t.bg
	}
	return color.Black
}

func (t *dummyColorTheme) Font(fyne.TextStyle) fyne.Resource     { return nil }
func (t *dummyColorTheme) Icon(fyne.ThemeIconName) fyne.Resource { return nil }
func (t *dummyColorTheme) Size(fyne.ThemeSizeName) float32       { return 0 }

func TestVariantForDefaultThemeUsesFallback(t *testing.T) {
	th := fyneTheme.DefaultTheme()
	v := variantFor(th, fyneTheme.VariantLight)
	if v != fyneTheme.VariantLight {
		t.Fatalf("default theme should use explicit fallback variant, got %v", v)
	}
}

func TestVariantForCustomThemeIgnoresFallback(t *testing.T) {
	th := theme.NewCustomTheme(theme.Presets[1]) // dracula, dark-only
	v := variantFor(th, fyneTheme.VariantLight)
	if v != fyneTheme.VariantDark {
		t.Fatalf("dark custom theme should ignore fallback and return VariantDark, got %v", v)
	}
}

func TestVariantForDarkBg(t *testing.T) {
	th := &dummyColorTheme{bg: color.NRGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xff}}
	v := variantFromBackground(th)
	if v != fyneTheme.VariantDark {
		t.Fatalf("dark background should return VariantDark, got %v", v)
	}
}

func TestVariantForLightBg(t *testing.T) {
	th := &dummyColorTheme{bg: color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff}}
	v := variantFromBackground(th)
	if v != fyneTheme.VariantLight {
		t.Fatalf("light background should return VariantLight, got %v", v)
	}
}

func TestHexUint16Validator(t *testing.T) {
	cases := []struct {
		input string
		valid bool
	}{
		{"0x1234", true},
		{"1234", true},
		{"0xABCD", true},
		{"0x12345", false},
		{"xyz", false},
		{"", false},
	}
	for _, c := range cases {
		err := hexUint16Validator(c.input)
		if c.valid && err != nil {
			t.Errorf("hexUint16Validator(%q) unexpected error: %v", c.input, err)
		}
		if !c.valid && err == nil {
			t.Errorf("hexUint16Validator(%q) expected error, got nil", c.input)
		}
	}
}
