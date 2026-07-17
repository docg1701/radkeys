package editor

import (
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
	"github.com/docg1701/radkeys/internal/widgetutil"
)

// buildInspector creates the property inspector for the selected button.
func (e *Editor) buildInspector() fyne.CanvasObject {
	b, ok := e.selectedButton()
	if !ok {
		return container.NewVBox(
			widget.NewLabel(i18n.T("editor.click_to_edit")),
			widget.NewLabel(i18n.T("editor.click_to_add")),
		)
	}

	var specific fyne.CanvasObject
	switch b.Action {
	case config.ActionText, config.ActionExec:
		specific = e.contentField(b)
	case config.ActionNavigate:
		specific = e.targetField(b)
	default:
		specific = container.NewHBox()
	}

	removeBtn := widget.NewButton(i18n.T("editor.remove"), func() {
		if e.selected == nil {
			return
		}
		e.removeButton(e.selected.row, e.selected.col)
	})
	removeBtn.Importance = widget.DangerImportance

	return container.NewVBox(
		e.labelField(b),
		e.actionField(b),
		specific,
		container.NewHBox(removeBtn),
	)
}

// refreshInspector rebuilds the property bar.
func (e *Editor) refreshInspector() {
	e.inspector = e.buildInspector()
}

// ponytail: package-level timer — Fyne is single-threaded for UI events.
const labelDebounce = 200 * time.Millisecond

var labelDebounceTimer *time.Timer

// labelField edits the button label.
func (e *Editor) labelField(b config.Button) fyne.CanvasObject {
	ent := widget.NewEntry()
	ent.SetText(b.Label)
	ent.OnChanged = func(label string) {
		e.setButtonLabel(label)
		if labelDebounceTimer != nil {
			labelDebounceTimer.Stop()
		}
		labelDebounceTimer = time.AfterFunc(labelDebounce, func() {
			fyne.Do(func() {
				e.refreshGrid()
				e.refreshProblems()
				e.updateButtonsTab()
			})
		})
	}
	return widgetutil.Labeled(i18n.T("editor.label"), ent)
}

// actionField edits the button action.
func (e *Editor) actionField(b config.Button) fyne.CanvasObject {
	sel := widget.NewSelect(config.ActionLabels(), nil)
	sel.SetSelected(config.ActionLabel(b.Action))
	sel.OnChanged = func(choice string) {
		e.setButtonAction(config.ActionIDFromLabel(choice))
	}
	return widgetutil.Labeled(i18n.T("editor.action"), sel)
}

// contentField edits multi-line report text.
func (e *Editor) contentField(b config.Button) fyne.CanvasObject {
	ent := widget.NewMultiLineEntry()
	ent.SetText(b.Content)
	ent.OnChanged = e.setButtonContent
	ent.Wrapping = fyne.TextWrapWord
	ent.SetMinRowsVisible(12)
	return widgetutil.Labeled(i18n.T("editor.content"), ent)
}

// targetField edits the navigate target.
func (e *Editor) targetField(b config.Button) fyne.CanvasObject {
	screens := e.cfg.Screens
	ids := make([]string, len(screens))
	labels := make([]string, len(screens))
	for i, s := range screens {
		ids[i] = s.ID
		labels[i] = s.DropdownLabel()
	}
	sel := widget.NewSelect(labels, nil)
	if idx := slices.Index(ids, b.Target); idx >= 0 {
		sel.SetSelectedIndex(idx)
	} else {
		sel.PlaceHolder = i18n.T("editor.select_target")
	}
	sel.OnChanged = func(label string) {
		if i := slices.Index(labels, label); i >= 0 {
			e.setButtonTarget(ids[i])
		}
	}
	return widgetutil.Labeled(i18n.T("editor.target"), sel)
}
