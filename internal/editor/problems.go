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
	if e.selected == nil || e.selected.screen != e.current {
		return container.NewVBox()
	}
	issues := e.issuesAt(e.current, e.selected.row, e.selected.col)
	if len(issues) == 0 {
		return container.NewVBox()
	}
	items := make([]fyne.CanvasObject, 0, len(issues))
	for _, issue := range issues {
		lbl := widget.NewLabel(e.issueMessage(issue))
		lbl.TextStyle = fyne.TextStyle{Italic: true}
		items = append(items, lbl)
	}
	return container.NewVBox(items...)
}

// refreshProblems rebuilds the validation strip.
func (e *Editor) refreshProblems() {
	e.problemsBox = e.buildProblems()
}

// hasBlockingIssues reports whether any issue blocks saving.
func (e *Editor) hasBlockingIssues() bool {
	return len(e.cfg.Issues()) > 0
}

// issuesAt returns every config.Issue that points at (row, col) on the given
// screen index, or nil if none.
func (e *Editor) issuesAt(screenIdx, row, col int) []config.Issue {
	sid := e.cfg.Screens[screenIdx].ID
	var out []config.Issue
	for _, issue := range e.cfg.Issues() {
		if issue.ScreenID == sid && issue.Row == row && issue.Col == col {
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
