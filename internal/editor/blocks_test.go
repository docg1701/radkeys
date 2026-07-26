package editor

import (
	"testing"

	"github.com/docg1701/radkeys/internal/config"
)

func blocksFixture() *config.Config {
	return &config.Config{
		App: config.App{
			Language: "en",
			Theme:    config.Theme{Preset: "system"},
			Device:   config.Device{VendorID: 0x1234, ProductID: 0xABCD, Protocol: config.ProtocolDIY},
			Layout:   config.Layout{Blocks: []config.Block{{Rows: 2, Cols: 4}, {Rows: 4, Cols: 6}, {Rows: 1, Cols: 4}}},
		},
		Screens: []config.Screen{
			{
				ID:   "root",
				Name: "Root",
				Buttons: []config.Button{
					{Block: 0, Row: 0, Col: 0, Label: "A", Action: config.ActionCopy},
					{Block: 1, Row: 0, Col: 0, Label: "B", Action: config.ActionCopy},
					{Block: 2, Row: 0, Col: 0, Label: "C", Action: config.ActionCopy},
				},
			},
		},
	}
}

func TestDeleteBlockRemovesItsButtonsAndShiftsLaterBlocks(t *testing.T) {
	cfg := blocksFixture()
	deleteBlock(cfg, 1)
	if len(cfg.App.Layout.Blocks) != 2 {
		t.Fatalf("blocks = %d, want 2", len(cfg.App.Layout.Blocks))
	}
	buttons := cfg.Screens[0].Buttons
	if len(buttons) != 2 {
		t.Fatalf("buttons = %d, want 2 (block 1 button deleted)", len(buttons))
	}
	if buttons[0].Label != "A" || buttons[0].Block != 0 {
		t.Errorf("button A = %+v, want block 0 untouched", buttons[0])
	}
	if buttons[1].Label != "C" || buttons[1].Block != 1 {
		t.Errorf("button C = %+v, want block shifted 2→1", buttons[1])
	}
	if len(cfg.Issues()) != 0 {
		t.Fatalf("config should be valid after deleteBlock: %v", cfg.Issues())
	}
}

func TestDeleteLastBlockLeavesOthersUntouched(t *testing.T) {
	cfg := blocksFixture()
	deleteBlock(cfg, 2)
	buttons := cfg.Screens[0].Buttons
	if len(buttons) != 2 || buttons[0].Block != 0 || buttons[1].Block != 1 {
		t.Fatalf("buttons = %+v, want A@0 and B@1", buttons)
	}
	if len(cfg.Issues()) != 0 {
		t.Fatalf("config should be valid: %v", cfg.Issues())
	}
}

func TestCellExists(t *testing.T) {
	e := &Editor{cfg: blocksFixture()}
	cases := []struct {
		b    config.Button
		want bool
	}{
		{config.Button{Block: 0, Row: 1, Col: 3}, true},
		{config.Button{Block: 1, Row: 3, Col: 5}, true},
		{config.Button{Block: 0, Row: 2, Col: 0}, false}, // block 0 has 2 rows
		{config.Button{Block: 0, Row: 0, Col: 4}, false}, // block 0 has 4 cols
		{config.Button{Block: 3, Row: 0, Col: 0}, false}, // no block 3
		{config.Button{Block: -1, Row: 0, Col: 0}, false},
		{config.Button{Block: 0, Row: -1, Col: 0}, false},
	}
	for _, tc := range cases {
		if got := e.cellExists(tc.b); got != tc.want {
			t.Errorf("cellExists(%+v) = %v, want %v", tc.b, got, tc.want)
		}
	}
}
