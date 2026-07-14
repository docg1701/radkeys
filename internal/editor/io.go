package editor

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// buildMenu creates the File menu.
func (e *Editor) buildMenu() *fyne.MainMenu {
	file := fyne.NewMenu(i18n.T("editor.file_menu"),
		fyne.NewMenuItem(i18n.T("editor.new"), e.newConfig),
		fyne.NewMenuItem(i18n.T("editor.open"), e.openConfig),
		fyne.NewMenuItem(i18n.T("editor.save"), e.saveConfig),
		fyne.NewMenuItem(i18n.T("editor.save_as"), e.saveConfigAs),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(i18n.T("button.close"), e.onCloseIntercept),
	)
	return fyne.NewMainMenu(file)
}

// onCloseIntercept asks for confirmation when there are unsaved changes.
func (e *Editor) onCloseIntercept() {
	if !e.dirty {
		e.win.Close()
		return
	}
	e.confirmDiscard(func() { e.win.Close() })
}

// confirmDiscard asks before discarding unsaved changes; runs onDiscard on "Discard".
func (e *Editor) confirmDiscard(onDiscard func()) {
	msg := widget.NewLabel(i18n.T("editor.confirm_discard"))
	msg.Wrapping = fyne.TextWrapWord
	d := dialog.NewCustomConfirm(
		i18n.T("editor.confirm_discard_title"),
		i18n.T("editor.discard"),
		i18n.T("editor.cancel"),
		msg,
		func(ok bool) {
			if ok {
				onDiscard()
			}
		},
		e.win,
	)
	d.Resize(fyne.NewSize(480, 180))
	d.Show()
}

// newConfig starts a fresh default config.
func (e *Editor) newConfig() {
	e.confirmDiscardAsync(func() {
		e.cfg = defaultConfig()
		e.path = ""
		e.current = 0
		e.selected = nil
		e.clearDirty()
		e.rebuildTabs()
	})
}

// confirmDiscardAsync runs onOK after confirming unsaved changes, if any.
func (e *Editor) confirmDiscardAsync(onOK func()) {
	if !e.dirty {
		onOK()
		return
	}
	e.confirmDiscard(onOK)
}

// openConfig loads an existing TOML file.
func (e *Editor) openConfig() {
	e.confirmDiscardAsync(func() {
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
		defer rc.Close()
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

// defaultConfig returns a blank starter config with one empty screen.
func defaultConfig() *config.Config {
	return &config.Config{
		App: config.App{
			Name:     "RadKeys",
			Language: "en",
			Device:   config.Device{VendorID: 0x1234, ProductID: 0xABCD, Protocol: config.ProtocolDIY},
			Layout:   config.Layout{Columns: 6, Rows: 6},
			Theme:    config.Theme{Preset: "system"},
		},
		Screens: []config.Screen{{ID: "root", Name: "Home"}},
	}
}

// StartupPath resolves the file to open at launch.
func StartupPath() string {
	if p := os.Getenv("RADKEYS_CONFIG"); p != "" {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "radkeys.config.toml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "radkeys.config.toml"
}

// LoadStartup loads the config at path or returns a fresh default.
func LoadStartup(path string) (*config.Config, error) {
	if _, err := os.Stat(path); err != nil {
		return defaultConfig(), fmt.Errorf("no config found at %s", path)
	}
	cfg, err := config.Load(path)
	if err != nil {
		log.Printf("radkeys-config: cannot load %s: %v", path, err)
		return defaultConfig(), err
	}
	return cfg, nil
}
