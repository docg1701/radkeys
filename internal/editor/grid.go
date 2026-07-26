package editor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/gridframe"
	"github.com/docg1701/radkeys/internal/i18n"
)

// buildGrid renders one framed grid per layout block plus an out-of-grid strip.
func (e *Editor) buildGrid() fyne.CanvasObject {
	s := e.currentScreen()
	if s == nil {
		return widget.NewLabel(i18n.T("editor.no_problems"))
	}
	th := e.app.Settings().Theme()
	v := e.app.Settings().ThemeVariant()
	blocks := e.cfg.App.Layout.Blocks
	frames := make([]fyne.CanvasObject, len(blocks))
	for i, b := range blocks {
		grid := container.NewGridWithColumns(b.Cols)
		for r := 0; r < b.Rows; r++ {
			for c := 0; c < b.Cols; c++ {
				grid.Objects = append(grid.Objects, e.buildGridCell(i, r, c))
			}
		}
		frames[i] = gridframe.Frame(grid, gridframe.Caption(e.cfg.App.Layout, i), th, v)
	}
	return container.NewVBox(container.NewHBox(frames...), e.buildOutOfGridStrip(s))
}

// buildGridCell creates one grid cell button or empty-cell placeholder.
func (e *Editor) buildGridCell(block, row, col int) fyne.CanvasObject {
	s := e.currentScreen()
	b, ok := s.ButtonAt(block, row, col)
	if !ok {
		return e.emptyCell(block, row, col)
	}
	return e.filledCell(b, block, row, col)
}

// emptyCell renders a clickable "+" placeholder showing the cell's slot.
func (e *Editor) emptyCell(block, row, col int) fyne.CanvasObject {
	slot := e.cfg.App.Layout.SlotOf(block, row, col)
	btn := widget.NewButton(fmt.Sprintf("n%d %s", slot, i18n.T("editor.empty_cell")), func() {
		e.onEmptyCellClicked(block, row, col)
	})
	btn.Importance = widget.LowImportance
	return btn
}

// onEmptyCellClicked adds or moves a button to (block, row, col).
func (e *Editor) onEmptyCellClicked(block, row, col int) {
	e.clearSelection()
	e.addButton(block, row, col)
}

// filledCell renders a button that already exists on the grid.
func (e *Editor) filledCell(b config.Button, block, row, col int) fyne.CanvasObject {
	label := e.cellLabel(b, e.cfg.App.Layout.SlotOf(block, row, col))
	btn := widget.NewButton(label, func() { e.selectCell(e.current, block, row, col) })
	if e.isSelected(block, row, col) {
		btn.Importance = widget.HighImportance
	}
	if len(e.issuesAt(e.current, block, row, col)) > 0 {
		btn.Importance = widget.DangerImportance
	}
	return btn
}

// cellLabel returns the display text for a grid button: slot, label, and the
// translated action name, so each cell shows where it lives and what it does.
// Empty labels fall back to a "label required" hint.
func (e *Editor) cellLabel(b config.Button, slot int) string {
	prefix := ""
	if slot >= 0 {
		prefix = fmt.Sprintf("n%d · ", slot)
	}
	if b.Label == "" {
		return prefix + i18n.T("editor.label_required")
	}
	return fmt.Sprintf("%s%s · %s", prefix, b.Label, config.ActionLabel(b.Action))
}

// isSelected reports whether (block, row, col) on the current screen is selected.
func (e *Editor) isSelected(block, row, col int) bool {
	return e.selected != nil && e.selected.screen == e.current &&
		e.selected.block == block && e.selected.row == row && e.selected.col == col
}

// buildOutOfGridStrip lists buttons on the current screen with no valid cell.
func (e *Editor) buildOutOfGridStrip(s *config.Screen) fyne.CanvasObject {
	buttons := make([]fyne.CanvasObject, 0, len(s.Buttons))
	for _, b := range s.Buttons {
		if e.cellExists(b) {
			continue
		}
		buttons = append(buttons, e.outOfGridButton(b))
	}
	if len(buttons) == 0 {
		return container.NewVBox()
	}
	title := widget.NewLabel(fmt.Sprintf("%s:", i18n.T("editor.problems_title")))
	title.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewVBox(title, container.NewHBox(buttons...))
}

// cellExists reports whether the button's (block, row, col) is a real cell.
func (e *Editor) cellExists(b config.Button) bool {
	if b.Block < 0 || b.Block >= len(e.cfg.App.Layout.Blocks) {
		return false
	}
	blk := e.cfg.App.Layout.Blocks[b.Block]
	return b.Row >= 0 && b.Row < blk.Rows && b.Col >= 0 && b.Col < blk.Cols
}

// outOfGridButton renders one out-of-grid button with its problem message.
func (e *Editor) outOfGridButton(b config.Button) fyne.CanvasObject {
	msg := fmt.Sprintf(i18n.T("editor.out_of_grid"), b.Label)
	btn := widget.NewButton(e.cellLabel(b, -1), func() {
		e.selectCell(e.current, b.Block, b.Row, b.Col)
	})
	btn.Importance = widget.DangerImportance
	lbl := widget.NewLabel(msg)
	lbl.TextStyle = fyne.TextStyle{Italic: true}
	return container.NewVBox(btn, lbl)
}

// updateButtonsTab replaces the Buttons tab content with the rebuilt grid.
func (e *Editor) updateButtonsTab() {
	if len(e.tabs.Items) < 2 {
		return
	}
	e.tabs.Items[1].Content = e.buildButtonsTab()
	e.tabs.Refresh()
}
