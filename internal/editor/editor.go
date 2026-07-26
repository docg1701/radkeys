// Package editor provides a visual editor for radkeys.config.toml.
// It is a separate, optional binary; the main RadKeys app is unchanged.
package editor

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
	themes "github.com/docg1701/radkeys/internal/theme"
)

// cellKey identifies one button cell by screen index and (block, row, col).
type cellKey struct {
	screen int
	block  int
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
	app      fyne.App
	win      fyne.Window

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
	e.win.SetMainMenu(e.buildMenu())
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

	inspectorPanel := container.NewVScroll(
		container.NewVBox(e.inspector, e.problemsBox),
	)
	gridPanel := container.NewVScroll(e.gridBox)
	split := container.NewHSplit(gridPanel, inspectorPanel)
	split.Offset = 0.60

	return container.NewBorder(e.layerBar, nil, nil, nil, split)
}

// refresh updates the window title and rebuilds the Buttons tab (which
// rebuilds the layer bar, grid, inspector, and problems strip in one pass).
func (e *Editor) refresh() {
	e.refreshTitle()
	e.updateButtonsTab()
}

// refreshTitle adds an asterisk prefix when the config is dirty and
// includes the current file path (or "unsaved" label when no path).
func (e *Editor) refreshTitle() {
	prefix := ""
	if e.dirty {
		prefix = i18n.T("editor.unsaved_title") + " "
	}
	title := i18n.T("editor.title")
	if e.path != "" {
		title = fmt.Sprintf("%s — %s", title, e.path)
	} else {
		title = fmt.Sprintf("%s — %s", title, i18n.T("editor.unsaved"))
	}
	e.win.SetTitle(prefix + title)
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

// selectedButton returns the currently selected button, if any.
func (e *Editor) selectedButton() (config.Button, bool) {
	if e.selected == nil {
		return config.Button{}, false
	}
	if e.selected.screen < 0 || e.selected.screen >= len(e.cfg.Screens) {
		return config.Button{}, false
	}
	return e.cfg.Screens[e.selected.screen].ButtonAt(e.selected.block, e.selected.row, e.selected.col)
}

// selectCell selects a cell and refreshes the dependent UI surfaces.
func (e *Editor) selectCell(screen, block, row, col int) {
	e.selected = &cellKey{screen: screen, block: block, row: row, col: col}
	e.refresh()
}

// clearSelection clears the selected cell and refreshes the dependent UI
// surfaces.
func (e *Editor) clearSelection() {
	e.selected = nil
	e.refresh()
}

// addButton creates a new Text button at (block, row, col) on the current screen.
func (e *Editor) addButton(block, row, col int) {
	s := e.currentScreen()
	if s == nil {
		return
	}
	s.Buttons = append(s.Buttons, config.Button{
		Block:  block,
		Row:    row,
		Col:    col,
		Label:  "",
		Action: config.ActionText,
	})
	e.selectCell(e.current, block, row, col)
	e.setDirty()
}

// removeButton deletes the button at (block, row, col) on the current screen.
func (e *Editor) removeButton(block, row, col int) {
	idx, ok := e.cfg.Screens[e.current].ButtonIndex(block, row, col)
	if !ok {
		return
	}
	btns := &e.cfg.Screens[e.current].Buttons
	*btns = append((*btns)[:idx], (*btns)[idx+1:]...)
	e.clearSelection()
	e.setDirty()
}

// setButtonLabel updates the selected button's label. It only mutates data
// and the dirty flag; the UI refresh is debounced by labelField.
func (e *Editor) setButtonLabel(label string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	e.cfg.Screens[e.current].Buttons[idx].Label = label
	e.setDirty()
}

// setButtonAction updates the selected button's action and clears invalid fields.
func (e *Editor) setButtonAction(action string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	b := &e.cfg.Screens[e.current].Buttons[idx]
	b.Action = action
	if action != config.ActionText && action != config.ActionExec {
		b.Content = ""
	}
	if action != config.ActionNavigate {
		b.Target = ""
	}
	e.setDirty()
	e.updateButtonsTab()
}

// setButtonContent updates the selected button's content.
func (e *Editor) setButtonContent(content string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	e.cfg.Screens[e.current].Buttons[idx].Content = content
	e.setDirty()
	e.updateButtonsTab()
}

// setButtonTarget updates the selected button's navigate target.
func (e *Editor) setButtonTarget(target string) {
	idx, ok := e.selectedIndex()
	if !ok {
		return
	}
	e.cfg.Screens[e.current].Buttons[idx].Target = target
	e.setDirty()
	e.updateButtonsTab()
}

// selectedIndex returns the index of the selected button on the current screen.
func (e *Editor) selectedIndex() (int, bool) {
	if e.selected == nil || e.selected.screen != e.current {
		return -1, false
	}
	return e.cfg.Screens[e.current].ButtonIndex(e.selected.block, e.selected.row, e.selected.col)
}

// resizeBlock sets one block's dimensions without deleting buttons.
func (e *Editor) resizeBlock(i, rows, cols int) {
	if i < 0 || i >= len(e.cfg.App.Layout.Blocks) {
		return
	}
	e.cfg.App.Layout.Blocks[i].Rows = rows
	e.cfg.App.Layout.Blocks[i].Cols = cols
	e.setDirty()
	e.updateButtonsTab()
}

// addBlock appends a 1×1 block at the end of the layout.
func (e *Editor) addBlock() {
	e.cfg.App.Layout.Blocks = append(e.cfg.App.Layout.Blocks, config.Block{Rows: 1, Cols: 1})
	e.setDirty()
	e.rebuildTabs()
}

// removeBlock deletes a block after confirmation. Buttons on it are deleted;
// buttons on later blocks follow their block (index shifted down by one).
func (e *Editor) removeBlock(i int) {
	blocks := &e.cfg.App.Layout.Blocks
	if i < 0 || i >= len(*blocks) {
		return
	}
	count := 0
	for _, s := range e.cfg.Screens {
		for _, b := range s.Buttons {
			if b.Block == i {
				count++
			}
		}
	}
	dialog.ShowConfirm(i18n.T("editor.remove_block"), fmt.Sprintf(i18n.T("editor.remove_block_confirm"), i, count), func(ok bool) {
		if !ok {
			return
		}
		deleteBlock(e.cfg, i)
		e.clearSelection()
		e.setDirty()
		e.rebuildTabs()
	}, e.win)
}

// deleteBlock removes block i: buttons on it are deleted and buttons on
// later blocks follow their block (index shifted down by one).
func deleteBlock(cfg *config.Config, i int) {
	cfg.App.Layout.Blocks = append(cfg.App.Layout.Blocks[:i], cfg.App.Layout.Blocks[i+1:]...)
	for si := range cfg.Screens {
		kept := cfg.Screens[si].Buttons[:0]
		for _, b := range cfg.Screens[si].Buttons {
			if b.Block == i {
				continue
			}
			if b.Block > i {
				b.Block--
			}
			kept = append(kept, b)
		}
		cfg.Screens[si].Buttons = kept
	}
}

// moveButtonToBlock moves the selected button to the first free cell of the
// target block; if none is free the button stays and the user is told.
func (e *Editor) moveButtonToBlock(target int) {
	idx, ok := e.selectedIndex()
	if !ok || target < 0 || target >= len(e.cfg.App.Layout.Blocks) {
		return
	}
	b := &e.cfg.Screens[e.current].Buttons[idx]
	if b.Block == target {
		return
	}
	blk := e.cfg.App.Layout.Blocks[target]
	for r := 0; r < blk.Rows; r++ {
		for c := 0; c < blk.Cols; c++ {
			if _, taken := e.cfg.Screens[e.current].ButtonAt(target, r, c); taken {
				continue
			}
			b.Block, b.Row, b.Col = target, r, c
			e.selectCell(e.current, target, r, c)
			e.setDirty()
			return
		}
	}
	dialog.ShowInformation(i18n.T("editor.no_free_cell_title"), fmt.Sprintf(i18n.T("editor.no_free_cell"), target), e.win)
}

// setAppLanguage switches the editor UI language and rebuilds tabs.
func (e *Editor) setAppLanguage(lang string) {
	e.cfg.App.Language = lang
	i18n.SetLanguage(lang)
	e.setDirty()
	e.rebuildTabs()
}

// setAppTheme stores the selected theme preset and applies it live.
func (e *Editor) setAppTheme(id string) {
	e.cfg.App.Theme.Preset = id
	e.setDirty()
	if p, ok := themes.FindPreset(id); ok {
		e.app.Settings().SetTheme(themes.NewCustomTheme(p))
	}
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

// setVendorIDFromEntry parses the hex value in the entry and updates the
// config vendor ID when valid. It marks the entry on parse failure.
func (e *Editor) setVendorIDFromEntry(entry *widget.Entry, s string) {
	e.setHexIDFromEntry(entry, s, &e.cfg.App.Device.VendorID)
}

// setProductIDFromEntry parses the hex value in the entry and updates the
// config product ID when valid. It marks the entry on parse failure.
func (e *Editor) setProductIDFromEntry(entry *widget.Entry, s string) {
	e.setHexIDFromEntry(entry, s, &e.cfg.App.Device.ProductID)
}

func (e *Editor) setHexIDFromEntry(entry *widget.Entry, s string, dst *uint16) {
	v, err := config.ParseHexUint16(s)
	if err != nil {
		entry.SetValidationError(fmt.Errorf("%s", i18n.T("settings.invalid_hex")))
		return
	}
	entry.SetValidationError(nil)
	*dst = v
	e.setDirty()
}

// setProtocol stores the device protocol.
func (e *Editor) setProtocol(p string) {
	e.cfg.App.Device.Protocol = p
	e.setDirty()
}
