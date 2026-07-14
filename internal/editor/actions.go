package editor

import (
	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// actionDef pairs a canonical action id with the i18n key used for its label.
// A single ordered slice is the only place that must change when adding a new
// action (H4).
type actionDef struct {
	ID       string
	LabelKey string
}

// actionDefs is the single source of truth for all 12 button actions in their
// editor display order. To add a 13th action, append one line here.
var actionDefs = []actionDef{
	{config.ActionText, "editor.action_text"},
	{config.ActionExec, "editor.action_exec"},
	{config.ActionCopy, "button.copy"},
	{config.ActionPaste, "button.paste"},
	{config.ActionPrev, "button.back"},
	{config.ActionHome, "button.home"},
	{config.ActionNavigate, "editor.action_navigate"},
	{config.ActionSelectAll, "button.select_all"},
	{config.ActionSelectLine, "button.select_line"},
	{config.ActionLineStart, "button.line_start"},
	{config.ActionLineEnd, "button.line_end"},
	{config.ActionBackspace, "button.backspace"},
	{config.ActionDelete, "button.delete"},
}

// actionLabelByID returns the localized label for a given action id.
func actionLabelByID(id string) string {
	for _, def := range actionDefs {
		if def.ID == id {
			return i18n.T(def.LabelKey)
		}
	}
	return id
}

// actionIDByLabel returns the canonical id for a localized label.
func actionIDByLabel(label string) string {
	for _, def := range actionDefs {
		if i18n.T(def.LabelKey) == label {
			return def.ID
		}
	}
	return config.ActionText
}

// actionLabels returns the localized labels in display order for the action
// dropdown.
func actionLabels() []string {
	labels := make([]string, len(actionDefs))
	for i, def := range actionDefs {
		labels[i] = i18n.T(def.LabelKey)
	}
	return labels
}
