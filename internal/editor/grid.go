package editor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// buildGrid renders the current screen grid plus an out-of-grid strip.
func (e *Editor) buildGrid() fyne.CanvasObject {
	s := e.currentScreen()
	if s == nil {
		return widget.NewLabel(i18n.T("editor.no_problems"))
	}
	cols := e.cfg.App.Layout.Columns
	rows := e.cfg.App.Layout.Rows
	grid := container.NewGridWithColumns(cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			grid.Objects = append(grid.Objects, e.buildGridCell(r, c))
		}
	}
	return container.NewVBox(grid, e.buildOutOfGridStrip(s))
}

// buildGridCell creates one grid cell button or empty-cell placeholder.
func (e *Editor) buildGridCell(row, col int) fyne.CanvasObject {
	s := e.currentScreen()
	b, ok := s.ButtonAt(row, col)
	if !ok {
		return e.emptyCell(row, col)
	}
	return e.filledCell(b, row, col)
}

// emptyCell renders a clickable "+" placeholder.
func (e *Editor) emptyCell(row, col int) fyne.CanvasObject {
	btn := widget.NewButton(i18n.T("editor.empty_cell"), func() {
		e.onEmptyCellClicked(row, col)
	})
	btn.Importance = widget.LowImportance
	return btn
}

// onEmptyCellClicked adds or moves a button to (row, col).
func (e *Editor) onEmptyCellClicked(row, col int) {
	e.clearSelection()
	e.addButton(row, col)
}

// filledCell renders a button that already exists on the grid.
func (e *Editor) filledCell(b config.Button, row, col int) fyne.CanvasObject {
	label := e.cellLabel(b)
	btn := widget.NewButton(label, func() { e.selectCell(e.current, row, col) })
	if e.isSelected(row, col) {
		btn.Importance = widget.HighImportance
	}
	if _, bad := e.cellProblem(e.current, row, col); bad {
		btn.Importance = widget.DangerImportance
	}
	return btn
}

// cellLabel returns the display text for a grid button: the label plus the
// translated action name as a compact indicator, so each cell shows what its
// button does at a glance (the PLAN's "label + an action indicator"). Empty
// labels fall back to a "label required" hint.
func (e *Editor) cellLabel(b config.Button) string {
	if b.Label == "" {
		return i18n.T("editor.label_required")
	}
	return fmt.Sprintf("%s · %s", b.Label, e.actionLabel(b.Action))
}

// isSelected reports whether (row, col) on the current screen is selected.
func (e *Editor) isSelected(row, col int) bool {
	return e.selected != nil && e.selected.screen == e.current && e.selected.row == row && e.selected.col == col
}

// buildOutOfGridStrip lists buttons on the current screen outside the layout.
func (e *Editor) buildOutOfGridStrip(s *config.Screen) fyne.CanvasObject {
	cols := e.cfg.App.Layout.Columns
	rows := e.cfg.App.Layout.Rows
	var buttons []fyne.CanvasObject
	for _, b := range s.Buttons {
		if b.Row >= rows || b.Col >= cols || b.Row < 0 || b.Col < 0 {
			buttons = append(buttons, e.outOfGridButton(b))
		}
	}
	if len(buttons) == 0 {
		return container.NewVBox()
	}
	title := widget.NewLabel(fmt.Sprintf("%s:", i18n.T("editor.problems_title")))
	title.TextStyle = fyne.TextStyle{Bold: true}
	return container.NewVBox(title, container.NewHBox(buttons...))
}

// outOfGridButton renders one out-of-grid button with its problem message.
func (e *Editor) outOfGridButton(b config.Button) fyne.CanvasObject {
	msg := fmt.Sprintf(i18n.T("editor.out_of_grid"), b.Label)
	btn := widget.NewButton(e.cellLabel(b), func() {
		e.selected = &cellKey{screen: e.current, row: b.Row, col: b.Col}
		e.refreshInspector()
		e.refreshGrid()
	})
	btn.Importance = widget.DangerImportance
	lbl := widget.NewLabel(msg)
	lbl.TextStyle = fyne.TextStyle{Italic: true}
	return container.NewVBox(btn, lbl)
}

// refreshGrid rebuilds the grid and out-of-grid strip.
func (e *Editor) refreshGrid() {
	if e.tabs == nil {
		return
	}
	e.gridBox = e.buildGrid()
	e.updateButtonsTab()
}

// updateButtonsTab replaces the Buttons tab content with the rebuilt grid.
func (e *Editor) updateButtonsTab() {
	if len(e.tabs.Items) < 2 {
		return
	}
	e.tabs.Items[1].Content = e.buildButtonsTab()
	e.tabs.Refresh()
}

// cellProblem returns the first validation problem for a cell, if any.
func (e *Editor) cellProblem(screenIdx, row, col int) (string, bool) {
	issues := e.cfg.Issues()
	for _, issue := range issues {
		if issue.ScreenID != e.cfg.Screens[screenIdx].ID {
			continue
		}
		if issue.Row != row || issue.Col != col {
			continue
		}
		return e.issueMessage(issue), true
	}
	return "", false
}
