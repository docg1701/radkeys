package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	fyneTheme "fyne.io/fyne/v2/theme"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/hid"
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

func TestIndexOf(t *testing.T) {
	options := []string{"a", "b", "c"}
	if got := indexOf(options, "b"); got != 1 {
		t.Fatalf("indexOf(..., \"b\") = %d, want 1", got)
	}
	if got := indexOf(options, "z"); got != -1 {
		t.Fatalf("indexOf(..., \"z\") = %d, want -1", got)
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

func TestDeviceCommands(t *testing.T) {
	want := map[string]struct {
		cmd hid.Command
		arg byte
	}{
		config.ActionPaste:      {hid.CmdFirePaste, byte(hid.ModifierForOS())},
		config.ActionSelectAll:  {hid.CmdSelectAll, byte(hid.ModifierForOS())},
		config.ActionSelectLine: {hid.CmdSelectLine, 0x00},
		config.ActionLineStart:  {hid.CmdLineStart, 0x00},
		config.ActionLineEnd:    {hid.CmdLineEnd, 0x00},
		config.ActionBackspace:  {hid.CmdBackspace, 0x00},
		config.ActionDelete:     {hid.CmdDelete, 0x00},
	}
	if len(deviceCommands) != len(want) {
		t.Fatalf("deviceCommands has %d entries, want %d", len(deviceCommands), len(want))
	}
	for action, expected := range want {
		def, ok := deviceCommands[action]
		if !ok {
			t.Fatalf("deviceCommands missing %q", action)
		}
		if def.cmd != expected.cmd {
			t.Errorf("deviceCommands[%q].cmd = %v, want %v", action, def.cmd, expected.cmd)
		}
		if got := def.arg(); got != expected.arg {
			t.Errorf("deviceCommands[%q].arg() = 0x%02x, want 0x%02x", action, got, expected.arg)
		}
	}
}
