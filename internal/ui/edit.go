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

// buildEditor returns the configuration tab content: list of screens on the
// left, form on the right, save/close at the bottom.
func (u *appUI) buildEditor() fyne.CanvasObject {
	ed := &editor{app: u, cfg: cloneConfig(u.cfg), path: u.configPath}
	ed.sel = -1

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
	ed.idEnt = widget.NewEntry()
	ed.idEnt.OnChanged = func(s string) {
		if ed.sel >= 0 && ed.sel < len(ed.cfg.Screens) {
			ed.cfg.Screens[ed.sel].ID = s
		}
	}

	ed.btnList = container.NewVBox()

	addScr := widget.NewButtonWithIcon("Nova tela", theme.ContentAddIcon(), ed.addScreen)
	delScr := widget.NewButtonWithIcon("Remover tela", theme.ContentRemoveIcon(), ed.delScreen)
	addBtn := widget.NewButtonWithIcon("Novo botão", theme.ContentAddIcon(), ed.addButton)
	saveBtn := widget.NewButtonWithIcon("Salvar", theme.DocumentSaveIcon(), ed.save)

	left := container.NewBorder(nil, container.NewVBox(addScr, delScr), nil, nil, ed.list)

	form := container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("ID", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			ed.idEnt,
			widget.NewLabelWithStyle("Título", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			ed.titleEnt,
			widget.NewSeparator(),
			widget.NewLabelWithStyle("Botões", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		), nil, nil, nil,
		container.NewBorder(nil, container.NewVBox(addBtn, saveBtn), nil, nil,
			container.NewVScroll(ed.btnList)),
	)

	return container.NewHSplit(left, form)
}

type editor struct {
	app      *appUI
	cfg      *config.Config
	path     string
	list     *widget.List
	sel      int
	titleEnt *widget.Entry
	idEnt    *widget.Entry
	btnList  *fyne.Container
}

func (ed *editor) selectScreen(i int) {
	ed.sel = i
	if i < 0 || i >= len(ed.cfg.Screens) {
		ed.titleEnt.SetText("")
		ed.idEnt.SetText("")
		ed.btnList.Objects = ed.btnList.Objects[:0]
		ed.btnList.Refresh()
		return
	}
	s := ed.cfg.Screens[i]
	ed.titleEnt.SetText(s.Title)
	ed.idEnt.SetText(s.ID)
	ed.rebuildButtons()
}

func (ed *editor) rebuildButtons() {
	ed.btnList.Objects = ed.btnList.Objects[:0]
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	for j := range ed.cfg.Screens[ed.sel].Buttons {
		j := j
		ed.btnList.Objects = append(ed.btnList.Objects, ed.buttonCard(j))
	}
	ed.btnList.Refresh()
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

	targetEnt := widget.NewEntry()
	targetEnt.SetText(b.Target)
	targetEnt.OnChanged = func(s string) { b.Target = s }

	contentEnt := widget.NewEntry()
	contentEnt.SetText(b.Content)
	contentEnt.OnChanged = func(s string) { b.Content = s }
	contentEnt.MultiLine = true
	contentEnt.Wrapping = fyne.TextWrapWord

	delBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() { ed.delButton(j) })

	form := container.NewVBox(
		container.NewHBox(
			widget.NewLabel("Índice"), idxEnt,
			widget.NewLabel("Rótulo"), labelEnt,
			actSel,
			delBtn,
		),
		widget.NewLabel("Target"),
		targetEnt,
		widget.NewLabel("Conteúdo"),
		contentEnt,
	)
	return widget.NewCard("", fmt.Sprintf("Botão %d", j), form)
}

func (ed *editor) addScreen() {
	ed.cfg.Screens = append(ed.cfg.Screens, config.Screen{
		ID:    fmt.Sprintf("tela_%d", len(ed.cfg.Screens)),
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
	ed.idEnt.SetText("")
	ed.btnList.Objects = ed.btnList.Objects[:0]
	ed.btnList.Refresh()
	ed.list.Refresh()
}

func (ed *editor) addButton() {
	if ed.sel < 0 || ed.sel >= len(ed.cfg.Screens) {
		return
	}
	s := &ed.cfg.Screens[ed.sel]
	s.Buttons = append(s.Buttons, config.Button{
		Index:  len(s.Buttons) + 3,
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
	ed.app.cfg = reloaded
	ed.app.deck = deck.New(reloaded)
	ed.app.renderScreen()
	dialog.ShowInformation("Salvo", "Configuração salva e recarregada.", ed.app.win)
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
