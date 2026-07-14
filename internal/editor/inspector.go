package editor

import (
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
	case config.ActionText:
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

// labelField edits the button label.
func (e *Editor) labelField(b config.Button) fyne.CanvasObject {
	ent := widget.NewEntry()
	ent.SetText(b.Label)
	ent.OnChanged = func(label string) {
		e.setButtonLabel(label)
		e.labelDebouncer.Add(func() {
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
	actions := e.actionOptions()
	sel := widget.NewSelect(actions, nil)
	sel.SetSelected(e.actionLabel(b.Action))
	sel.OnChanged = func(choice string) {
		e.setButtonAction(e.actionFromLabel(choice))
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
	names := e.targetOptions()
	sel := widget.NewSelect(names, nil)
	sel.SetSelected(e.targetName(b.Target))
	sel.OnChanged = func(choice string) {
		e.setButtonTarget(e.targetFromName(choice))
	}
	return widgetutil.Labeled(i18n.T("editor.target"), sel)
}

// actionOptions returns the human-readable labels for all 12 actions.
func (e *Editor) actionOptions() []string {
	return actionLabels()
}

// actionLabel returns the display label for an action id.
func (e *Editor) actionLabel(action string) string {
	return actionLabelByID(action)
}

// actionFromLabel maps a display label back to the action id.
func (e *Editor) actionFromLabel(label string) string {
	return actionIDByLabel(label)
}

// targetOptions returns human-readable names for the target dropdown.
func (e *Editor) targetOptions() []string {
	names := make([]string, 0, len(e.cfg.Screens))
	for _, s := range e.cfg.Screens {
		names = append(names, targetLabel(s))
	}
	return names
}

// targetLabel formats a screen as "id — name".
func targetLabel(s config.Screen) string {
	return s.ID + " — " + s.Name
}

// targetName returns the label for a target id, or a placeholder.
func (e *Editor) targetName(id string) string {
	for _, s := range e.cfg.Screens {
		if s.ID == id {
			return targetLabel(s)
		}
	}
	return i18n.T("editor.select_target")
}

// targetFromName maps a target label back to the screen id.
func (e *Editor) targetFromName(name string) string {
	for _, s := range e.cfg.Screens {
		if targetLabel(s) == name {
			return s.ID
		}
	}
	return ""
}
