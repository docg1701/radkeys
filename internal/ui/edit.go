package ui

import (
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/BurntSushi/toml"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/deck"
)

// buildEditor returns the "Editar" tab: left = list of screens, right = form
// to edit the selected screen's ID, title, and buttons. Buttons can be added
// and removed. Save writes TOML and reloads the live config + keypad.
func (u *appUI) buildEditor() fyne.CanvasObject {
	ed := &editor{app: u, cfg: cloneConfig(u.cfg), path: u.configPath, sel: -1}

	ed.list = widget.NewList(
		func() int { return len(ed.cfg.Screens) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(ed.cfg.Screens[i].Title)
		},
	)
	ed.list.OnSelected = func(i int) { ed.selectScreen(i) }

	ed.titleEnt = widget.NewEntry()
	ed.idEnt = widget.NewEntry()
	ed.btnScroll = container.NewVBox()

	ed.titleEnt.OnChanged = func(s string) {
		if ed.sel >= 0 && ed.sel < len(ed.cfg.Screens) {
			ed.cfg.Screens[ed.sel].Title = s
			ed.list.Refresh()
		}
	}
	ed.idEnt.OnChanged = func(s string) {
		if ed.sel >= 0 && ed.sel < len(ed.cfg.Screens) {
			ed.cfg.Screens[ed.sel].ID = s
		}
	}

	addScr := widget.NewButtonWithIcon("Nova tela", theme.ContentAddIcon(), ed.addScreen)
	delScr := widget.NewButtonWithIcon("Remover tela", theme.DeleteIcon(), ed.delScreen)
	addBtn := widget.NewButtonWithIcon("Novo botão", theme.ContentAddIcon(), ed.addButton)
	saveBtn := widget.NewButtonWithIcon("Salvar e aplicar", theme.DocumentSaveIcon(), ed.save)

	left := container.NewBorder(nil, container.NewVBox(addScr, delScr), nil, nil, ed.list)

	rightForm := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Tela", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			widget.NewLabel("ID"), ed.idEnt,
			widget.NewLabel("Título"), ed.titleEnt,
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Botões", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		),
		container.NewVBox(addBtn, saveBtn),
		nil, nil,
		container.NewVScroll(ed.btnScroll),
	)

	return container.NewHSplit(left, rightForm)
}

type editor struct {
	app       *appUI
	cfg       *config.Config
	path      string
	list      *widget.List
	sel       int
	titleEnt  *widget.Entry
	idEnt     *widget.Entry
	btnScroll *fyne.Container
}

func (ed *editor) selectScreen(i int) {
	ed.sel = i
	if i < 0 || i >= len(ed.cfg.Screens) {
		ed.titleEnt.SetText("")
		ed.idEnt.SetText("")
		ed.btnScroll.Objects = ed.btnScroll.Objects[:0]
		ed.btnScroll.Refresh()
		return
	}
	s := ed.cfg.Screens[i]
	ed.titleEnt.SetText(s.Title)
	ed.idEnt.SetText(s.ID)
	ed.rebuildButtons()
}

func (ed *editor) rebuildButtons() {
	ed.btnScroll.Objects = ed.btnScroll.Objects[:0]
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	for j := range ed.cfg.Screens[ed.sel].Buttons {
		j := j
		ed.btnScroll.Objects = append(ed.btnScroll.Objects, ed.buttonCard(j))
	}
	ed.btnScroll.Refresh()
}

func (ed *editor) buttonCard(j int) fyne.CanvasObject {
	b := &ed.cfg.Screens[ed.sel].Buttons[j]

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

	contentEnt := widget.NewEntry()
	contentEnt.SetText(b.Content)
	contentEnt.OnChanged = func(s string) { b.Content = s }
	contentEnt.MultiLine = true
	contentEnt.Wrapping = fyne.TextWrapWord

	targetEnt := widget.NewEntry()
	targetEnt.SetText(b.Target)
	targetEnt.OnChanged = func(s string) { b.Target = s }

	delBtn := widget.NewButtonWithIcon("Remover", theme.DeleteIcon(), func() { ed.delButton(j) })

	return widget.NewCard("", fmt.Sprintf("Botão %d", j), container.NewVBox(
		container.NewHBox(widget.NewLabel("Índice"), idxEnt, widget.NewLabel("Rótulo"), labelEnt, actSel, delBtn),
		widget.NewLabel("Conteúdo (texto do laudo)"),
		contentEnt,
		widget.NewLabel("Target (se navigate)"),
		targetEnt,
	))
}

func (ed *editor) addScreen() {
	ed.cfg.Screens = append(ed.cfg.Screens, config.Screen{
		ID:    fmt.Sprintf("tela_%d", len(ed.cfg.Screens)),
		Title: "Nova tela",
	})
	ed.list.Refresh()
	idx := len(ed.cfg.Screens) - 1
	ed.list.Select(idx)
}

func (ed *editor) delScreen() {
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	ed.cfg.Screens = append(ed.cfg.Screens[:ed.sel], ed.cfg.Screens[ed.sel+1:]...)
	ed.sel = -1
	ed.titleEnt.SetText("")
	ed.idEnt.SetText("")
	ed.btnScroll.Objects = ed.btnScroll.Objects[:0]
	ed.btnScroll.Refresh()
	ed.list.Refresh()
	ed.list.UnselectAll()
}

func (ed *editor) addButton() {
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	s := &ed.cfg.Screens[ed.sel]
	nextIdx := 3
	for _, b := range s.Buttons {
		if b.Index >= nextIdx {
			nextIdx = b.Index + 1
		}
	}
	s.Buttons = append(s.Buttons, config.Button{
		Index:   nextIdx,
		Label:   "Novo botão",
		Action:  config.ActionText,
		Content: "Texto do laudo.",
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
		dialog.ShowError(fmt.Errorf("salvar: %w", err), ed.app.win)
		return
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(ed.cfg); err != nil {
		dialog.ShowError(fmt.Errorf("codificar TOML: %w", err), ed.app.win)
		return
	}
	reloaded, err := config.Load(ed.path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("recarregar: %w", err), ed.app.win)
		return
	}
	// Atualiza a config viva e o deck.
	ed.app.cfg = reloaded
	ed.app.deck = deck.New(reloaded)
	ed.app.renderScreen()
	dialog.ShowInformation("Salvo", "Configuração salva e aplicada à tela de atalhos.", ed.app.win)
}

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
