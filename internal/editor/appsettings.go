package editor

import (
	"fmt"
	"slices"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
	themes "github.com/docg1701/radkeys/internal/theme"
	"github.com/docg1701/radkeys/internal/widgetutil"
)

// buildAppSettings creates the first tab with app-wide fields.
func (e *Editor) buildAppSettings() fyne.CanvasObject {
	e.appSettings = container.NewVBox(
		e.buildAppearanceGroup(),
		e.buildBlocksGroup(),
		e.buildDeviceGroup(),
	)
	return container.NewVScroll(container.NewPadded(e.appSettings))
}

// buildAppearanceGroup groups name, radiologist, language, and theme.
func (e *Editor) buildAppearanceGroup() fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetText(e.cfg.App.Name)
	name.OnChanged = e.setAppName

	rad := widget.NewEntry()
	rad.SetText(e.cfg.App.Radiologist)
	rad.OnChanged = e.setRadiologist

	lang := widget.NewSelect(i18n.Supported, nil)
	lang.SetSelected(e.cfg.App.Language)
	lang.OnChanged = e.setAppLanguage

	themeIDs, themeNames := themes.Options()
	theme := widget.NewSelect(themeNames, nil)
	theme.SetSelected(i18n.T("theme." + e.cfg.App.Theme.Preset))
	theme.OnChanged = func(choice string) {
		e.setAppTheme(themeIDs[slices.Index(themeNames, choice)])
	}
	return widgetutil.Section(i18n.T("settings.group_appearance"),
		container.NewGridWithColumns(2,
			widgetutil.Labeled(i18n.T("editor.app_name"), name),
			widgetutil.Labeled(i18n.T("settings.radiologist"), rad),
		),
		container.NewGridWithColumns(2,
			widgetutil.Labeled(i18n.T("settings.language"), lang),
			widgetutil.Labeled(i18n.T("settings.theme"), theme),
		),
	)
}

// buildBlocksGroup edits the ordered block list with user-defined names,
// per-block dimensions, and add/remove. Layout per block:
//
//	[Label "Block 0"]  [rows entry]  [cols entry]  [empty]
//	[name entry]          [empty]       [empty]       [Remove]
//
// The add-button is a full-width button in a 3-column grid (col 1).
func (e *Editor) buildBlocksGroup() fyne.CanvasObject {
	rows := make([]fyne.CanvasObject, 0, len(e.cfg.App.Layout.Blocks)+1)

	for i := 0; i < len(e.cfg.App.Layout.Blocks); i++ {
		idx := i

		capt := widget.NewLabel(fmt.Sprintf(i18n.T("block.default_name"), idx))

		nameEnt := widget.NewEntry()
		nameEnt.SetText(e.cfg.App.Layout.Blocks[idx].Name)
		nameEnt.SetPlaceHolder(i18n.T("editor.block_name_placeholder"))
		nameEnt.OnChanged = func(s string) {
			e.cfg.App.Layout.Blocks[idx].Name = s
			e.setDirty()
			e.updateButtonsTab()
		}

		rowsEnt := widget.NewEntry()
		rowsEnt.SetText(strconv.Itoa(e.cfg.App.Layout.Blocks[idx].Rows))
		rowsEnt.OnChanged = func(s string) {
			v, err := strconv.Atoi(s)
			if err != nil || v < 1 {
				rowsEnt.SetValidationError(fmt.Errorf("≥ 1"))
				return
			}
			rowsEnt.SetValidationError(nil)
			e.resizeBlock(idx, v, e.cfg.App.Layout.Blocks[idx].Cols)
		}

		colsEnt := widget.NewEntry()
		colsEnt.SetText(strconv.Itoa(e.cfg.App.Layout.Blocks[idx].Cols))
		colsEnt.OnChanged = func(s string) {
			v, err := strconv.Atoi(s)
			if err != nil || v < 1 {
				colsEnt.SetValidationError(fmt.Errorf("≥ 1"))
				return
			}
			colsEnt.SetValidationError(nil)
			e.resizeBlock(idx, e.cfg.App.Layout.Blocks[idx].Rows, v)
		}

		del := widget.NewButton(i18n.T("editor.remove"), func() { e.removeBlock(idx) })
		del.Importance = widget.DangerImportance

		rows = append(rows, container.NewGridWithColumns(4,
			container.NewVBox(capt, nameEnt),
			widgetutil.Labeled(i18n.T("settings.rows"), rowsEnt),
			widgetutil.Labeled(i18n.T("settings.columns"), colsEnt),
			container.NewBorder(nil, del, nil, nil, widget.NewLabel("")),
		))
	}

	add := widget.NewButton(i18n.T("editor.add_block"), e.addBlock)
	rows = append(rows, container.NewGridWithColumns(3,
		widget.NewLabel(""), add, widget.NewLabel(""),
	))
	return widgetutil.Section(i18n.T("editor.blocks"), rows...)
}

// buildDeviceGroup groups VID, PID, and protocol.
func (e *Editor) buildDeviceGroup() fyne.CanvasObject {
	vid := widget.NewEntry()
	vid.SetText(fmt.Sprintf("0x%04x", e.cfg.App.Device.VendorID))
	vid.SetPlaceHolder(i18n.T("editor.hex_format"))
	vid.OnChanged = func(s string) { e.setVendorIDFromEntry(vid, s) }

	pid := widget.NewEntry()
	pid.SetText(fmt.Sprintf("0x%04x", e.cfg.App.Device.ProductID))
	pid.SetPlaceHolder(i18n.T("editor.hex_format"))
	pid.OnChanged = func(s string) { e.setProductIDFromEntry(pid, s) }

	proto := widget.NewSelect([]string{config.ProtocolDIY}, nil)
	proto.SetSelected(e.cfg.App.Device.Protocol)
	proto.OnChanged = e.setProtocol

	return widgetutil.Section(i18n.T("settings.group_device"),
		container.NewGridWithColumns(3,
			widgetutil.Labeled(i18n.T("settings.vid"), vid),
			widgetutil.Labeled(i18n.T("settings.pid"), pid),
			widgetutil.Labeled(i18n.T("settings.protocol"), proto),
		),
	)
}
