package editor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// buildLayerBar creates the screen/layer management bar.
func (e *Editor) buildLayerBar() fyne.CanvasObject {
	s := e.currentScreen()
	if s == nil {
		return container.NewHBox()
	}
	names := e.layerOptions()
	sel := widget.NewSelect(names, nil)
	sel.SetSelected(layerLabel(*s))
	sel.OnChanged = func(choice string) {
		e.switchToLayer(e.layerIndexFromName(choice))
	}

	add := widget.NewButton(i18n.T("editor.add_layer"), e.addLayer)
	remove := widget.NewButton(i18n.T("editor.remove_layer"), e.askRemoveLayer)
	rename := widget.NewButton(i18n.T("editor.rename_layer"), e.askRenameLayer)

	return container.NewBorder(
		nil, nil,
		widget.NewLabel(i18n.T("editor.layer")),
		container.NewHBox(add, remove, rename),
		sel,
	)
}

// refreshLayerBar rebuilds the layer management bar.
func (e *Editor) refreshLayerBar() {
	e.layerBar = e.buildLayerBar()
	e.updateButtonsTab()
}

// layerOptions returns labels for every screen.
func (e *Editor) layerOptions() []string {
	names := make([]string, 0, len(e.cfg.Screens))
	for _, s := range e.cfg.Screens {
		names = append(names, layerLabel(s))
	}
	return names
}

// layerLabel formats a screen as "id — name".
func layerLabel(s config.Screen) string {
	return s.ID + " — " + s.Name
}

// layerIndexFromName maps a layer label back to its index.
func (e *Editor) layerIndexFromName(name string) int {
	for i, s := range e.cfg.Screens {
		if layerLabel(s) == name {
			return i
		}
	}
	return e.current
}

// switchToLayer changes the current layer.
func (e *Editor) switchToLayer(idx int) {
	if idx < 0 || idx >= len(e.cfg.Screens) || idx == e.current {
		return
	}
	e.current = idx
	e.clearSelection()
	e.refresh()
}

// addLayer appends a new screen with a generated id and name.
func (e *Editor) addLayer() {
	id := e.uniqueLayerID()
	e.cfg.Screens = append(e.cfg.Screens, config.Screen{
		ID:   id,
		Name: i18n.T("editor.new_layer_name"),
	})
	e.current = len(e.cfg.Screens) - 1
	e.setDirty()
	e.refresh()
}

// uniqueLayerID generates a new screen id that does not collide.
func (e *Editor) uniqueLayerID() string {
	base := "layer"
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s%d", base, i)
		if !e.hasLayerID(candidate) {
			return candidate
		}
	}
}

// hasLayerID reports whether a screen id already exists.
func (e *Editor) hasLayerID(id string) bool {
	for _, s := range e.cfg.Screens {
		if s.ID == id {
			return true
		}
	}
	return false
}

// askRemoveLayer confirms before deleting the current layer.
func (e *Editor) askRemoveLayer() {
	if len(e.cfg.Screens) <= 1 {
		d := dialog.NewInformation(i18n.T("editor.problems_title"), i18n.T("editor.cannot_remove_last_screen"), e.win)
		d.Resize(fyne.NewSize(500, 200))
		d.Show()
		return
	}
	if e.isLayerTargeted(e.cfg.Screens[e.current].ID) {
		d := dialog.NewInformation(i18n.T("editor.problems_title"), i18n.T("editor.cannot_remove_targeted_screen"), e.win)
		d.Resize(fyne.NewSize(500, 200))
		d.Show()
		return
	}
	d := dialog.NewConfirm(i18n.T("editor.remove_layer"), i18n.T("editor.confirm_remove_screen"), func(ok bool) {
		if ok {
			e.removeLayer()
		}
	}, e.win)
	d.Resize(fyne.NewSize(450, 180))
	d.Show()
}

// isLayerTargeted reports whether any navigate button targets the given screen id.
func (e *Editor) isLayerTargeted(id string) bool {
	for _, s := range e.cfg.Screens {
		for _, b := range s.Buttons {
			if b.Action == config.ActionNavigate && b.Target == id {
				return true
			}
		}
	}
	return false
}

// removeLayer deletes the current screen.
func (e *Editor) removeLayer() {
	if len(e.cfg.Screens) <= 1 {
		return
	}
	e.cfg.Screens = append(e.cfg.Screens[:e.current], e.cfg.Screens[e.current+1:]...)
	if e.current >= len(e.cfg.Screens) {
		e.current = len(e.cfg.Screens) - 1
	}
	e.clearSelection()
	e.setDirty()
	e.refresh()
}

// askRenameLayer opens a dialog to rename the current layer.
func (e *Editor) askRenameLayer() {
	s := e.currentScreen()
	if s == nil {
		return
	}
	name := widget.NewEntry()
	name.SetText(s.Name)
	name.SetPlaceHolder(i18n.T("editor.layer_name"))
	form := dialog.NewForm(
		i18n.T("editor.rename_layer"),
		i18n.T("editor.save"),
		i18n.T("editor.cancel"),
		[]*widget.FormItem{
			{Text: i18n.T("editor.layer_name"), Widget: name, HintText: s.ID},
		},
		func(ok bool) {
			if ok {
				e.renameLayer(name.Text)
			}
		},
		e.win,
	)
	form.Resize(fyne.NewSize(500, 220))
	form.Show()
}

// renameLayer changes the current screen's name.
func (e *Editor) renameLayer(name string) {
	s := e.currentScreen()
	if s == nil {
		return
	}
	s.Name = name
	e.setDirty()
	e.refresh()
}
