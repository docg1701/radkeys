package ui

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

// TestHIDPathDoesNotActivateWindow is the static guard for the
// HID_FOCUS_INVARIANT documented on appUI.press. The physical-keypad event path
// (pollHID -> press(fromUI=false) -> renderGrid/flashStatus/setStatus) must never
// activate the RadKeys window. Runtime window focus is not observable in
// headless Fyne tests, so we guard statically: parse ui.go and fail if any
// HID-path method calls u.win.{Raise, Show, ShowAndRun, SetContent, RequestFocus}.
//
// u.win is the only window handle reachable from these methods, so matching the
// selector u.win.<method> avoids false positives on widget methods of the same
// name (e.g. u.status.Show). dialog.ShowInformation(..., u.win) on the fromUI
// branch is allowed: u.win is an argument there, not the call receiver.
func TestHIDPathDoesNotActivateWindow(t *testing.T) {
	path := uiSourcePath(t)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	forbidden := map[string]bool{
		"Show":         true,
		"ShowAndRun":   true,
		"SetContent":   true,
		"RequestFocus": true,
	}
	hidPath := map[string]bool{
		"pollHID":     true,
		"press":       true,
		"renderGrid":  true,
		"flashStatus": true,
		"setStatus":   true,
	}

	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || !hidPath[fn.Name.Name] {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel == nil || !forbidden[sel.Sel.Name] {
				return true
			}
			// sel.X must be the window field access u.win.
			inner, ok := sel.X.(*ast.SelectorExpr)
			if !ok || inner.Sel == nil || inner.Sel.Name != "win" {
				return true
			}
			recv, ok := inner.X.(*ast.Ident)
			if !ok || recv.Name != "u" {
				return true
			}
			pos := fset.Position(call.Pos())
			t.Errorf("HID path method %s calls forbidden window-activation "+
				"u.win.%s() at %s:%d — violates HID_FOCUS_INVARIANT",
				fn.Name.Name, sel.Sel.Name, filepath.Base(path), pos.Line)
			return true
		})
	}
}

// uiSourcePath returns the absolute path to ui.go next to this test file.
func uiSourcePath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test source file")
	}
	return filepath.Join(filepath.Dir(file), "ui.go")
}
