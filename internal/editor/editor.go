// Package editor provides a visual editor for radkeys.config.toml.
// It is a separate, optional binary; the main RadKeys app is unchanged.
package editor

import (
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// cellKey identifies one button cell by screen index and (row, col).
type cellKey struct {
	screen int
	row    int
	col    int
}

// Editor holds the visual editor state and cached Fyne widgets.
type Editor struct {
	cfg      *config.Config
	path     string
	dirty    bool
	current  int
	selected *cellKey
	showHelp bool
	app      fyne.App
	win      fyne.Window
	navStack []int

	tabs        *container.AppTabs
	gridBox     fyne.CanvasObject
	inspector   fyne.CanvasObject
	layerBar    fyne.CanvasObject
	problemsBox fyne.CanvasObject
	appSettings fyne.CanvasObject
}

// NewEditor creates an Editor for the given config and file path.
func NewEditor(a fyne.App, w fyne.Window, cfg *config.Config, path string) *Editor {
	e := &Editor{
		cfg:     cfg,
		path:    path,
		app:     a,
		win:     w,
		current: 0,
	}
	e.buildUI()
	return e
}

// Run shows the editor window and starts the Fyne event loop.
func (e *Editor) Run() {
	e.win.SetMainMenu(e.buildMenu())
	e.win.SetContent(e.tabs)
	e.win.SetCloseIntercept(e.onCloseIntercept)
	e.win.ShowAndRun()
}

// rebuildTabs recreates the tab labels after a language change.
func (e *Editor) rebuildTabs() {
	idx := 0
	if e.tabs != nil {
		idx = e.tabs.SelectedIndex()
	}
	e.tabs = container.NewAppTabs(
		container.NewTabItem(i18n.T("editor.tab_app_settings"), e.buildAppSettings()),
		container.NewTabItem(i18n.T("editor.tab_buttons"), e.buildButtonsTab()),
	)
	e.tabs.SelectIndex(idx)
	e.win.SetContent(e.tabs)
}

// buildUI constructs the tab container and initial child widgets.
func (e *Editor) buildUI() {
	e.tabs = container.NewAppTabs(
		container.NewTabItem(i18n.T("editor.tab_app_settings"), e.buildAppSettings()),
		container.NewTabItem(i18n.T("editor.tab_buttons"), e.buildButtonsTab()),
	)
	e.tabs.SelectIndex(0)
}

// buildButtonsTab assembles the layer bar, inspector, grid, and problems strip.
func (e *Editor) buildButtonsTab() fyne.CanvasObject {
	e.layerBar = e.buildLayerBar()
	e.inspector = e.buildInspector()
	e.problemsBox = e.buildProblems()
	e.gridBox = e.buildGrid()
	top := container.NewVBox(e.layerBar, e.inspector, e.problemsBox)
	return container.NewBorder(top, nil, nil, nil, container.NewVScroll(e.gridBox))
}

// refresh updates all mutable UI surfaces and the window title.
func (e *Editor) refresh() {
	e.refreshTitle()
	e.refreshGrid()
	e.refreshInspector()
	e.refreshLayerBar()
	e.refreshProblems()
}

// refreshTitle adds an asterisk prefix when the config is dirty.
func (e *Editor) refreshTitle() {
	prefix := ""
	if e.dirty {
		prefix = i18n.T("editor.unsaved_title") + " "
	}
	e.win.SetTitle(prefix + i18n.T("editor.title"))
}

// setDirty marks the config as changed and refreshes the title.
func (e *Editor) setDirty() {
	if !e.dirty {
		e.dirty = true
		e.refreshTitle()
	}
}

// clearDirty marks the config as saved and refreshes the title.
func (e *Editor) clearDirty() {
	e.dirty = false
	e.refreshTitle()
}

// currentScreen returns the screen at the current index, or nil.
func (e *Editor) currentScreen() *config.Screen {
	if e.current < 0 || e.current >= len(e.cfg.Screens) {
		return nil
	}
	return &e.cfg.Screens[e.current]
}

// findButton returns the index of the button at (row, col) on a screen.
func (e *Editor) findButton(screenIdx, row, col int) (int, bool) {
	if screenIdx < 0 || screenIdx >= len(e.cfg.Screens) {
		return -1, false
	}
	for i, b := range e.cfg.Screens[screenIdx].Buttons {
		if b.Row == row && b.Col == col {
			return i, true
		}
	}
	return -1, false
}

// buttonAt returns the button at (row, col) on a screen.
func (e *Editor) buttonAt(screenIdx, row, col int) (config.Button, bool) {
	idx, ok := e.findButton(screenIdx, row, col)
	if !ok {
		return config.Button{}, false
	}
	return e.cfg.Screens[screenIdx].Buttons[idx], true
}

// selectedButton returns the currently selected button, if any.
func (e *Editor) selectedButton() (config.Button, bool) {
	if e.selected == nil {
		return config.Button{}, false
	}
	return e.buttonAt(e.selected.screen, e.selected.row, e.selected.col)
}

// selectCell selects a cell and refreshes dependent widgets.
func (e *Editor) selectCell(screen, row, col int) {
	e.selected = &cellKey{screen: screen, row: row, col: col}
	e.refreshInspector()
	e.refreshGrid()
}

// clearSelection clears the selected cell and refreshes dependent widgets.
func (e *Editor) clearSelection() {
	e.selected = nil
	e.refreshInspector()
	e.refreshGrid()
}

// addButton creates a new Text button at (row, col) on the current screen.
func (e *Editor) addButton(row, col int) {
	s := e.currentScreen()
	if s == nil {
		return
	}
	s.Buttons = append(s.Buttons, config.Button{
		Row:    row,
		Col:    col,
		Label:  "",
		Action: config.ActionText,
	})
	e.selectCell(e.current, row, col)
	e.setDirty()
	e.refresh()
}

// removeButton deletes the button at (row, col) on the current screen.
func (e *Editor) removeButton(row, col int) {
	idx, ok := e.findButton(e.current, row, col)
	if !ok {
		return
	}
	btns := &e.cfg.Screens[e.current].Buttons
	*btns = append((*btns)[:idx], (*btns)[idx+1:]...)
	e.clearSelection()
	e.setDirty()
	e.refresh()
}

// moveButton moves the selected button to (row, col) on the current screen.
func (e *Editor) moveButton(toRow, toCol int) {
	if e.selected == nil || e.selected.screen != e.current {
		return
	}
	idx, ok := e.findButton(e.current, e.selected.row, e.selected.col)
	if !ok {
		return
	}
	if _, occupied := e.findButton(e.current, toRow, toCol); occupied {
		return
	}
	btns := &e.cfg.Screens[e.current].Buttons
	(*btns)[idx].Row = toRow
	(*btns)[idx].Col = toCol
	e.selected = &cellKey{screen: e.current, row: toRow, col: toCol}
	e.setDirty()
	e.refresh()
}

// setButtonLabel updates the selected button's label.
func (e *Editor) setButtonLabel(label string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	e.cfg.Screens[e.current].Buttons[idx].Label = label
	e.setDirty()
	e.refreshGrid()
	e.refreshProblems()
}

// setButtonAction updates the selected button's action and clears invalid fields.
func (e *Editor) setButtonAction(action string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	b := &e.cfg.Screens[e.current].Buttons[idx]
	b.Action = action
	if action != config.ActionText {
		b.Content = ""
	}
	if action != config.ActionNavigate {
		b.Target = ""
	}
	e.setDirty()
	e.refreshInspector()
	e.refreshGrid()
	e.refreshProblems()
}

// setButtonContent updates the selected button's content.
func (e *Editor) setButtonContent(content string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	e.cfg.Screens[e.current].Buttons[idx].Content = content
	e.setDirty()
	e.refreshProblems()
}

// setButtonTarget updates the selected button's navigate target.
func (e *Editor) setButtonTarget(target string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	e.cfg.Screens[e.current].Buttons[idx].Target = target
	e.setDirty()
	e.refreshGrid()
	e.refreshProblems()
}

// selectedIndex returns the index of the selected button on the current screen.
func (e *Editor) selectedIndex() (int, bool) {
	if e.selected == nil || e.selected.screen != e.current {
		return -1, false
	}
	return e.findButton(e.current, e.selected.row, e.selected.col)
}

// resizeGrid sets the layout size without deleting buttons.
func (e *Editor) resizeGrid(cols, rows int) {
	e.cfg.App.Layout.Columns = cols
	e.cfg.App.Layout.Rows = rows
	e.setDirty()
	e.refreshGrid()
	e.refreshProblems()
}

// setAppLanguage switches the editor UI language and rebuilds tabs.
func (e *Editor) setAppLanguage(lang string) {
	e.cfg.App.Language = lang
	i18n.SetLanguage(lang)
	e.setDirty()
	e.rebuildTabs()
}

// setAppTheme stores the selected theme preset for RadKeys.
func (e *Editor) setAppTheme(id string) {
	e.cfg.App.Theme.Preset = id
	e.setDirty()
}

// setRadiologist stores the radiologist name.
func (e *Editor) setRadiologist(name string) {
	e.cfg.App.Radiologist = name
	e.setDirty()
}

// setAppName stores the application name.
func (e *Editor) setAppName(name string) {
	e.cfg.App.Name = name
	e.setDirty()
}

// setVendorID parses a hex string and stores the USB vendor ID.
func (e *Editor) setVendorID(s string) {
	if v, err := strconv.ParseUint(strip0x(s), 16, 16); err == nil {
		e.cfg.App.Device.VendorID = uint16(v)
		e.setDirty()
	}
}

// setProductID parses a hex string and stores the USB product ID.
func (e *Editor) setProductID(s string) {
	if v, err := strconv.ParseUint(strip0x(s), 16, 16); err == nil {
		e.cfg.App.Device.ProductID = uint16(v)
		e.setDirty()
	}
}

// setProtocol stores the device protocol.
func (e *Editor) setProtocol(p string) {
	e.cfg.App.Device.Protocol = p
	e.setDirty()
}

// strip0x removes a leading "0x" for hex parsing.
func strip0x(s string) string {
	return strings.TrimPrefix(strings.ToLower(s), "0x")
}
