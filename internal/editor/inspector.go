package editor

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
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
			e.refreshGrid()
			e.refreshProblems()
			e.updateButtonsTab()
		})
	}
	return labeled(i18n.T("editor.label"), ent)
}

// actionField edits the button action.
func (e *Editor) actionField(b config.Button) fyne.CanvasObject {
	actions := e.actionOptions()
	sel := widget.NewSelect(actions, nil)
	sel.SetSelected(e.actionLabel(b.Action))
	sel.OnChanged = func(choice string) {
		e.setButtonAction(e.actionFromLabel(choice))
	}
	return labeled(i18n.T("editor.action"), sel)
}

// contentField edits multi-line report text.
func (e *Editor) contentField(b config.Button) fyne.CanvasObject {
	ent := widget.NewMultiLineEntry()
	ent.SetText(b.Content)
	ent.OnChanged = e.setButtonContent
	ent.Wrapping = fyne.TextWrapWord
	ent.SetMinRowsVisible(12)
	return labeled(i18n.T("editor.content"), ent)
}

// targetField edits the navigate target.
func (e *Editor) targetField(b config.Button) fyne.CanvasObject {
	names := e.targetOptions()
	sel := widget.NewSelect(names, nil)
	sel.SetSelected(e.targetName(b.Target))
	sel.OnChanged = func(choice string) {
		e.setButtonTarget(e.targetFromName(choice))
	}
	return labeled(i18n.T("editor.target"), sel)
}

// actionOptions returns the human-readable labels for all 12 actions.
func (e *Editor) actionOptions() []string {
	return []string{
		i18n.T("editor.action_text"),
		i18n.T("button.copy"),
		i18n.T("button.paste"),
		i18n.T("button.back"),
		i18n.T("button.home"),
		i18n.T("editor.action_navigate"),
		i18n.T("button.select_all"),
		i18n.T("button.select_line"),
		i18n.T("button.line_start"),
		i18n.T("button.line_end"),
		i18n.T("button.backspace"),
		i18n.T("button.delete"),
	}
}

// actionLabel returns the display label for an action id.
func (e *Editor) actionLabel(action string) string {
	labels := map[string]string{
		config.ActionText:       i18n.T("editor.action_text"),
		config.ActionCopy:       i18n.T("button.copy"),
		config.ActionPaste:      i18n.T("button.paste"),
		config.ActionPrev:       i18n.T("button.back"),
		config.ActionHome:       i18n.T("button.home"),
		config.ActionNavigate:   i18n.T("editor.action_navigate"),
		config.ActionSelectAll:  i18n.T("button.select_all"),
		config.ActionSelectLine: i18n.T("button.select_line"),
		config.ActionLineStart:  i18n.T("button.line_start"),
		config.ActionLineEnd:    i18n.T("button.line_end"),
		config.ActionBackspace:  i18n.T("button.backspace"),
		config.ActionDelete:     i18n.T("button.delete"),
	}
	if label, ok := labels[action]; ok {
		return label
	}
	return action
}

// actionFromLabel maps a display label back to the action id.
func (e *Editor) actionFromLabel(label string) string {
	for _, a := range configActionOrder() {
		if e.actionLabel(a) == label {
			return a
		}
	}
	return config.ActionText
}

// configActionOrder returns the canonical action ids in display order.
func configActionOrder() []string {
	return []string{
		config.ActionText, config.ActionCopy, config.ActionPaste,
		config.ActionPrev, config.ActionHome, config.ActionNavigate,
		config.ActionSelectAll, config.ActionSelectLine,
		config.ActionLineStart, config.ActionLineEnd,
		config.ActionBackspace, config.ActionDelete,
	}
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

// labeled wraps an input under a label.
func labeled(label string, input fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(label), input)
}
