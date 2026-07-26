package editor

import (
	"fmt"
	"slices"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/gridframe"
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

// buildBlocksGroup edits the ordered block list: dimensions per block,
// plus add/remove. Slot ranges in the captions update live via refresh.
func (e *Editor) buildBlocksGroup() fyne.CanvasObject {
	rows := make([]fyne.CanvasObject, 0, len(e.cfg.App.Layout.Blocks)+1)
	for i, b := range e.cfg.App.Layout.Blocks {
		idx := i
		rowsEnt := widget.NewEntry()
		rowsEnt.SetText(strconv.Itoa(b.Rows))
		rowsEnt.OnChanged = func(s string) {
			if v, err := strconv.Atoi(s); err == nil && v >= 1 {
				rowsEnt.SetValidationError(nil)
				e.resizeBlock(idx, v, e.cfg.App.Layout.Blocks[idx].Cols)
			} else {
				rowsEnt.SetValidationError(fmt.Errorf("≥ 1"))
			}
		}
		colsEnt := widget.NewEntry()
		colsEnt.SetText(strconv.Itoa(b.Cols))
		colsEnt.OnChanged = func(s string) {
			if v, err := strconv.Atoi(s); err == nil && v >= 1 {
				colsEnt.SetValidationError(nil)
				e.resizeBlock(idx, e.cfg.App.Layout.Blocks[idx].Rows, v)
			} else {
				colsEnt.SetValidationError(fmt.Errorf("≥ 1"))
			}
		}
		del := widget.NewButton(i18n.T("editor.remove"), func() { e.removeBlock(idx) })
		del.Importance = widget.DangerImportance
		rows = append(rows, container.NewGridWithColumns(4,
			widget.NewLabel(gridframe.Caption(e.cfg.App.Layout, idx)),
			widgetutil.Labeled(i18n.T("settings.rows"), rowsEnt),
			widgetutil.Labeled(i18n.T("settings.columns"), colsEnt),
			del,
		))
	}
	rows = append(rows, widget.NewButton(i18n.T("editor.add_block"), e.addBlock))
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
