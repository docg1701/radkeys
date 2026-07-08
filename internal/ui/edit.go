package ui

import (
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/BurntSushi/toml"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/deck"
)

// openEditor opens a separate window for editing the config screens.
func (u *appUI) openEditor() {
	w := u.fapp.NewWindow("RadKeys — Editar configuração")
	w.Resize(fyne.NewSize(900, 600))

	ed := &editor{
		app:  u,
		cfg:  cloneConfig(u.cfg),
		path: u.configPath,
		win:  w,
	}
	ed.build()
	w.Show()
}

// editor holds the editing state (a working copy of the config).
type editor struct {
	app      *appUI
	cfg      *config.Config
	path     string
	win      fyne.Window
	list     *widget.List
	sel      int // selected screen index, -1 if none
	titleEnt *widget.Entry
	btnBox   *fyne.Container
}

func (ed *editor) build() {
	ed.list = widget.NewList(
		func() int { return len(ed.cfg.Screens) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(ed.cfg.Screens[i].Title)
		},
	)
	ed.list.OnSelected = func(i int) { ed.selectScreen(i) }

	ed.titleEnt = widget.NewEntry()
	ed.titleEnt.OnChanged = func(s string) {
		if ed.sel >= 0 && ed.sel < len(ed.cfg.Screens) {
			ed.cfg.Screens[ed.sel].Title = s
			ed.list.Refresh()
		}
	}

	ed.btnBox = container.NewVBox()

	addScrBtn := widget.NewButton("+ Tela", ed.addScreen)
	delScrBtn := widget.NewButton("− Tela", ed.delScreen)
	addBtnBtn := widget.NewButton("+ Botão", ed.addButton)
	saveBtn := widget.NewButton("Salvar", ed.save)
	closeBtn := widget.NewButton("Fechar", func() { ed.win.Close() })

	left := container.NewBorder(nil, container.NewVBox(addScrBtn, delScrBtn), nil, nil, ed.list)
	right := container.NewBorder(
		container.NewVBox(widget.NewLabel("Título"), ed.titleEnt),
		container.NewVBox(addBtnBtn, saveBtn, closeBtn),
		nil, nil,
		container.NewVScroll(ed.btnBox),
	)
	ed.win.SetContent(container.NewHSplit(left, right))
	ed.win.SetOnClosed(func() { ed.win = nil })
}

func (ed *editor) selectScreen(i int) {
	ed.sel = i
	if i < 0 || i >= len(ed.cfg.Screens) {
		ed.titleEnt.SetText("")
		ed.btnBox.Objects = nil
		ed.btnBox.Refresh()
		return
	}
	s := ed.cfg.Screens[i]
	ed.titleEnt.SetText(s.Title)
	ed.rebuildButtons()
}

func (ed *editor) rebuildButtons() {
	ed.btnBox.Objects = ed.btnBox.Objects[:0]
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	s := ed.cfg.Screens[ed.sel]
	for j := range s.Buttons {
		j := j
		row := ed.buttonRow(&s.Buttons[j], func() { ed.delButton(j) })
		ed.btnBox.Objects = append(ed.btnBox.Objects, row)
	}
	ed.btnBox.Refresh()
}

func (ed *editor) buttonRow(b *config.Button, onDel func()) fyne.CanvasObject {
	idxEnt := widget.NewEntry()
	idxEnt.SetText(strconv.Itoa(b.Index))
	idxEnt.OnChanged = func(s string) {
		if v, err := strconv.Atoi(s); err == nil {
			b.Index = v
		}
	}

	labelEnt := widget.NewEntry()
	labelEnt.SetText(b.Label)
	labelEnt.OnChanged = func(s string) { b.Label = s }

	actSel := widget.NewSelect([]string{config.ActionNavigate, config.ActionText}, func(s string) {
		b.Action = s
	})
	actSel.SetSelected(b.Action)

	targetEnt := widget.NewEntry()
	targetEnt.SetText(b.Target)
	targetEnt.OnChanged = func(s string) { b.Target = s }

	contentEnt := widget.NewEntry()
	contentEnt.SetText(b.Content)
	contentEnt.OnChanged = func(s string) { b.Content = s }
	contentEnt.MultiLine = true

	delBtn := widget.NewButton("X", onDel)

	return container.NewHBox(
		widget.NewLabel("Idx"), idxEnt,
		widget.NewLabel("Label"), labelEnt,
		actSel,
		widget.NewLabel("Target"), targetEnt,
		widget.NewLabel("Content"), contentEnt,
		delBtn,
	)
}

func (ed *editor) addScreen() {
	ed.cfg.Screens = append(ed.cfg.Screens, config.Screen{
		ID:    fmt.Sprintf("screen_%d", len(ed.cfg.Screens)),
		Title: "Nova tela",
	})
	ed.list.Refresh()
}

func (ed *editor) delScreen() {
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	ed.cfg.Screens = append(ed.cfg.Screens[:ed.sel], ed.cfg.Screens[ed.sel+1:]...)
	ed.sel = -1
	ed.titleEnt.SetText("")
	ed.btnBox.Objects = nil
	ed.btnBox.Refresh()
	ed.list.Refresh()
}

func (ed *editor) addButton() {
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	s := &ed.cfg.Screens[ed.sel]
	s.Buttons = append(s.Buttons, config.Button{
		Index:  len(s.Buttons) + 3, // after the 3 fixed buttons
		Label:  "Novo",
		Action: config.ActionText,
	})
	ed.rebuildButtons()
}

func (ed *editor) delButton(j int) {
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	s := &ed.cfg.Screens[ed.sel]
	if j < 0 || j >= len(s.Buttons) {
		return
	}
	s.Buttons = append(s.Buttons[:j], s.Buttons[j+1:]...)
	ed.rebuildButtons()
}

func (ed *editor) save() {
	f, err := os.Create(ed.path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("salvar: %w", err), ed.win)
		return
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(ed.cfg); err != nil {
		dialog.ShowError(fmt.Errorf("codificar TOML: %w", err), ed.win)
		return
	}
	// Reload the config into the main app and reset the deck.
	reloaded, err := config.Load(ed.path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("recarregar config: %w", err), ed.win)
		return
	}
	ed.app.cfg = reloaded
	ed.app.deck = deck.New(reloaded)
	ed.app.renderScreen()
	dialog.ShowInformation("Salvo", "Configuração salva e recarregada.", ed.win)
}

// cloneConfig returns a deep copy of the config for editing.
func cloneConfig(c *config.Config) *config.Config {
	cp := *c
	cp.Screens = make([]config.Screen, len(c.Screens))
	for i, s := range c.Screens {
		cp.Screens[i] = s
		cp.Screens[i].Buttons = make([]config.Button, len(s.Buttons))
		copy(cp.Screens[i].Buttons, s.Buttons)
	}
	return &cp
}
