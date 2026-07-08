package deck

import (
	"testing"

	"github.com/docg1701/radkeys/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		App: config.App{
			Name:         "RadKeys",
			Device:       config.Device{VendorID: 1, ProductID: 2, Protocol: config.ProtocolDIY},
			FixedButtons: config.FixedButtons{Copy: 0, LevelUp: 1, GoHome: 2},
		},
		Screens: []config.Screen{
			{ID: "root", Title: "Início", Buttons: []config.Button{
				{Index: 3, Label: "RX", Action: config.ActionNavigate, Target: "rx"},
				{Index: 4, Label: "Frase", Action: config.ActionText, Content: "olá"},
			}},
			{ID: "rx", Title: "RX", Buttons: []config.Button{
				{Index: 3, Label: "Tórax", Action: config.ActionText, Content: "tórax normal"},
			}},
		},
	}
}

func TestStartsAtRoot(t *testing.T) {
	d := New(testConfig())
	if d.CurrentScreen().ID != "root" {
		t.Fatalf("current = %q, want root", d.CurrentScreen().ID)
	}
}

func TestPressNavigate(t *testing.T) {
	d := New(testConfig())
	eff := d.Press(3) // root -> rx
	if eff.Type != EffectNavigate {
		t.Fatalf("want EffectNavigate, got %v", eff.Type)
	}
	if d.CurrentScreen().ID != "rx" {
		t.Fatalf("current = %q, want rx", d.CurrentScreen().ID)
	}
}

func TestPressTextLoadsPreview(t *testing.T) {
	d := New(testConfig())
	eff := d.Press(4) // text "olá"
	if eff.Type != EffectPreview || eff.Text != "olá" {
		t.Fatalf("want preview olá, got %+v", eff)
	}
	if d.Preview() != "olá" {
		t.Fatalf("preview = %q, want olá", d.Preview())
	}
}

func TestCopyReturnsPreview(t *testing.T) {
	d := New(testConfig())
	d.Press(4)        // load "olá"
	eff := d.Press(0) // fixed copy
	if eff.Type != EffectCopy || eff.Text != "olá" {
		t.Fatalf("want copy olá, got %+v", eff)
	}
}

func TestLevelUpReturnsToParent(t *testing.T) {
	d := New(testConfig())
	d.Press(3)        // root -> rx
	eff := d.Press(1) // level_up
	if eff.Type != EffectNavigate {
		t.Fatalf("want EffectNavigate, got %v", eff.Type)
	}
	if d.CurrentScreen().ID != "root" {
		t.Fatalf("current = %q, want root", d.CurrentScreen().ID)
	}
}

func TestGoHomeResetsToRoot(t *testing.T) {
	d := New(testConfig())
	d.Press(3)        // root -> rx
	eff := d.Press(2) // go_home
	if eff.Type != EffectNavigate {
		t.Fatalf("want EffectNavigate, got %v", eff.Type)
	}
	if d.CurrentScreen().ID != "root" {
		t.Fatalf("current = %q, want root", d.CurrentScreen().ID)
	}
}

func TestPressUnknownIndexIsNoop(t *testing.T) {
	d := New(testConfig())
	eff := d.Press(99)
	if eff.Type != EffectNone {
		t.Fatalf("want EffectNone, got %v", eff.Type)
	}
}

func TestLevelUpFromRootStaysRoot(t *testing.T) {
	d := New(testConfig())
	d.Press(1) // level_up at root
	if d.CurrentScreen().ID != "root" {
		t.Fatalf("current = %q, want root", d.CurrentScreen().ID)
	}
}
