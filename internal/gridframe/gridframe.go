// Package gridframe renders one keypad block as a captioned, framed grid.
// Host and editor share it so both render blocks identically.
package gridframe

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

// Caption returns the frame title. If the block has a name it is used;
// otherwise "Block i" is the fallback. The slot range always follows.
func Caption(l config.Layout, i int) string {
	start := l.SlotOf(i, 0, 0)
	end := start + l.Blocks[i].Rows*l.Blocks[i].Cols - 1
	name := l.Blocks[i].Name
	if name == "" {
		name = fmt.Sprintf(i18n.T("block.default_name"), i)
	}
	return fmt.Sprintf("%s (%d–%d)", name, start, end)
}

// Frame wraps a block grid with a caption on top and a stroke frame around it.
func Frame(grid *fyne.Container, caption string, th fyne.Theme, v fyne.ThemeVariant) fyne.CanvasObject {
	rect := canvas.NewRectangle(color.Transparent)
	rect.StrokeColor = th.Color(fyneTheme.ColorNameSeparator, v)
	rect.StrokeWidth = 1
	rect.CornerRadius = 4
	return container.NewBorder(
		widget.NewLabel(caption), nil, nil, nil,
		container.NewPadded(container.NewStack(rect, container.NewPadded(grid))),
	)
}
