package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
)

// Regression: the first paint must fill every block cell regardless of theme
// (the OS-theme listener early-returns for custom themes).
func TestRenderGridFillsAllBlockCells(t *testing.T) {
	a := test.NewApp()
	cfg := &config.Config{
		App: config.App{
			Layout: config.Layout{Blocks: []config.Block{{Rows: 2, Cols: 4}, {Rows: 4, Cols: 6}}},
		},
		Screens: []config.Screen{{ID: "root", Name: "Root", Buttons: []config.Button{
			{Block: 0, Row: 0, Col: 0, Label: "A", Action: config.ActionCopy},
			{Block: 1, Row: 3, Col: 5, Label: "B", Action: config.ActionCopy},
		}}},
	}
	u := &appUI{cfg: cfg, a: a, current: "root"}
	u.rebuildKeypad()
	u.refillGrids()

	if got := len(u.blockGrids[0].Objects); got != 8 {
		t.Fatalf("block 0 cells = %d, want 8", got)
	}
	if got := len(u.blockGrids[1].Objects); got != 24 {
		t.Fatalf("block 1 cells = %d, want 24", got)
	}
	btn, ok := u.blockGrids[0].Objects[0].(*widget.Button)
	if !ok {
		t.Fatalf("cell (0,0,0) is %T, want *widget.Button", u.blockGrids[0].Objects[0])
	}
	if btn.Text != "n0 · A" {
		t.Fatalf("button text = %q, want %q", btn.Text, "n0 · A")
	}
}
