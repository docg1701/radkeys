package editor

import (
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// buildMenuBar returns a custom menu bar using PopUpMenu (not native
// SetMainMenu) so the drop-down width can be controlled.
func (e *Editor) buildMenuBar() fyne.CanvasObject {
	labels := []struct {
		label  string
		action func()
	}{
		{i18n.T("editor.new"), e.newConfig},
		{i18n.T("editor.open"), e.openConfig},
		{i18n.T("editor.save"), e.saveConfig},
		{i18n.T("editor.save_as"), e.saveConfigAs},
		{"", nil}, // separator placeholder
		{i18n.T("editor.close_file"), e.closeFile},
		{i18n.T("editor.quit"), e.onCloseIntercept},
	}
	maxW := float32(0)
	for _, it := range labels {
		if it.label == "" {
			continue
		}
		w := fyne.MeasureText(it.label, 0, fyne.TextStyle{}).Width
		if w > maxW {
			maxW = w
		}
	}
	popW := maxW * 2 // at least 2× longest label

	var fileBtn *widget.Button
	fileBtn = widget.NewButton(i18n.T("editor.file_menu"), func() {
		items := make([]*fyne.MenuItem, 0, len(labels))
		for _, it := range labels {
			if it.label == "" {
				items = append(items, fyne.NewMenuItemSeparator())
			} else {
				items = append(items, fyne.NewMenuItem(it.label, it.action))
			}
		}
		menu := fyne.NewMenu("", items...)
		pop := widget.NewPopUpMenu(menu, e.win.Canvas())
		pop.Resize(fyne.NewSize(popW, pop.MinSize().Height))
		pop.ShowAtPosition(fyne.NewPos(0, fileBtn.Position().Y+fileBtn.Size().Height))
	})
	return container.NewHBox(fileBtn)
}

// closeFile resets to a blank default config without closing the window.
func (e *Editor) closeFile() {
	e.confirmDiscard(func() {
		e.cfg = config.DefaultConfig()
		e.path = ""
		e.current = 0
		e.selected = nil
		e.clearDirty()
		e.rebuildTabs()
	})
}

// onCloseIntercept asks for confirmation when there are unsaved changes.
func (e *Editor) onCloseIntercept() {
	e.confirmDiscard(e.win.Close)
}

// confirmDiscard runs action when there are no unsaved changes, otherwise
// shows a confirm dialog and runs action only on "Discard". Used by every
// path that can lose edits (close, new, open, close-file).
func (e *Editor) confirmDiscard(action func()) {
	if !e.dirty {
		action()
		return
	}
	msg := widget.NewLabel(i18n.T("editor.confirm_discard"))
	msg.Wrapping = fyne.TextWrapWord
	d := dialog.NewCustomConfirm(
		i18n.T("editor.confirm_discard_title"),
		i18n.T("editor.discard"),
		i18n.T("editor.cancel"),
		msg,
		func(ok bool) {
			if ok {
				action()
			}
		},
		e.win,
	)
	d.Resize(fyne.NewSize(480, 180))
	d.Show()
}

// newConfig starts a fresh default config.
func (e *Editor) newConfig() {
	e.confirmDiscard(func() {
		e.cfg = config.DefaultConfig()
		e.path = ""
		e.current = 0
		e.selected = nil
		e.clearDirty()
		e.rebuildTabs()
	})
}

// openConfig loads an existing TOML file.
func (e *Editor) openConfig() {
	e.confirmDiscard(func() {
		fd := dialog.NewFileOpen(e.onFileOpened, e.win)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".toml"}))
		fd.Resize(fyne.NewSize(900, 650))
		fd.Show()
	})
}

// onFileOpened handles the result of the file-open dialog.
func (e *Editor) onFileOpened(rc fyne.URIReadCloser, err error) {
	if err != nil || rc == nil {
		return
	}
	path := rc.URI().Path()
	cfg, err := config.Load(path)
	if err != nil {
		dialog.ShowError(err, e.win)
		return
	}
	e.cfg = cfg
	e.path = path
	e.current = 0
	e.selected = nil
	e.app.Preferences().SetString("lastFile", path)
	i18n.SetLanguage(e.cfg.App.Language)
	e.clearDirty()
	e.rebuildTabs()
}

// saveConfig saves to the current path after validating.
func (e *Editor) saveConfig() {
	if e.hasBlockingIssues() {
		e.showSaveBlocked()
		return
	}
	if e.path == "" {
		e.saveConfigAs()
		return
	}
	if err := e.cfg.Save(e.path); err != nil {
		dialog.ShowError(err, e.win)
		return
	}
	e.app.Preferences().SetString("lastFile", e.path)
	e.clearDirty()
}

// saveConfigAs asks for a path and saves.
func (e *Editor) saveConfigAs() {
	if e.hasBlockingIssues() {
		e.showSaveBlocked()
		return
	}
	fd := dialog.NewFileSave(func(rc fyne.URIWriteCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Printf("radkeys-config: close save dialog writer: %v", err)
			}
		}()
		path := rc.URI().Path()
		if filepath.Ext(path) != ".toml" {
			path += ".toml"
		}
		if err := e.cfg.Save(path); err != nil {
			dialog.ShowError(err, e.win)
			return
		}
		e.path = path
		e.app.Preferences().SetString("lastFile", path)
		e.clearDirty()
	}, e.win)
	fd.SetFilter(storage.NewExtensionFileFilter([]string{".toml"}))
	fd.Resize(fyne.NewSize(900, 650))
	fd.Show()
}

// showSaveBlocked explains why saving is blocked.
func (e *Editor) showSaveBlocked() {
	body := widget.NewLabel(i18n.T("editor.save_blocked_message"))
	body.Wrapping = fyne.TextWrapWord
	d := dialog.NewCustom(i18n.T("editor.save_blocked_title"), i18n.T("editor.cancel"), body, e.win)
	d.Resize(fyne.NewSize(500, 200))
	d.Show()
}
