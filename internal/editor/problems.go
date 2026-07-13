package editor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// buildProblems creates the validation strip below the inspector.
func (e *Editor) buildProblems() fyne.CanvasObject {
	issues := e.buttonIssues()
	if len(issues) == 0 {
		return container.NewVBox()
	}
	items := make([]fyne.CanvasObject, 0, len(issues))
	for _, issue := range issues {
		items = append(items, helpLine(e.issueMessage(issue)))
	}
	return container.NewVBox(items...)
}

// helpLine creates an italic helper label.
func helpLine(text string) fyne.CanvasObject {
	lbl := widget.NewLabel(text)
	lbl.TextStyle = fyne.TextStyle{Italic: true}
	return lbl
}

// refreshProblems rebuilds the validation strip.
func (e *Editor) refreshProblems() {
	e.problemsBox = e.buildProblems()
	e.updateButtonsTab()
}

// hasBlockingIssues reports whether any issue blocks saving.
func (e *Editor) hasBlockingIssues() bool {
	return len(e.cfg.Issues()) > 0
}

// buttonIssues filters the config issues to those on the current screen.
func (e *Editor) buttonIssues() []config.Issue {
	s := e.currentScreen()
	if s == nil {
		return nil
	}
	var out []config.Issue
	for _, issue := range e.cfg.Issues() {
		if issue.ScreenID == s.ID {
			out = append(out, issue)
		}
	}
	return out
}

// issueMessage translates a config Issue into plain language.
func (e *Editor) issueMessage(issue config.Issue) string {
	key, args := e.issueKeyArgs(issue)
	if key != "" {
		return fmt.Sprintf(i18n.T(key), args...)
	}
	return issue.Error(e.cfg.App.Layout.Rows, e.cfg.App.Layout.Columns).Error()
}

// issueKeyArgs maps an IssueKind to an i18n key and format arguments.
func (e *Editor) issueKeyArgs(issue config.Issue) (string, []any) {
	switch issue.Kind {
	case config.IssueEmptyLabel:
		return "editor.label_required", nil
	case config.IssueOutOfGridRow, config.IssueOutOfGridCol:
		return "editor.out_of_grid", []any{issue.Label}
	case config.IssueDuplicatePosition:
		return "editor.duplicate_pos", []any{issue.Detail, issue.Label}
	case config.IssueNavigateUnknownTarget:
		return "editor.bad_target", []any{issue.Detail}
	case config.IssueTextRequiresContent:
		return "editor.content_required", nil
	case config.IssueNavigateRequiresTarget:
		return "editor.target_required", nil
	case config.IssueActionRejectsTarget:
		return "editor.action_rejects_target", []any{issue.Detail}
	case config.IssueActionRejectsContent:
		return "editor.action_rejects_content", []any{issue.Detail}
	case config.IssueInvalidAction:
		return "editor.invalid_action", []any{issue.Detail}
	}
	return "", nil
}
