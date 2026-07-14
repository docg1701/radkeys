// Package widgetutil provides small Fyne layout helpers shared by ui and editor.
package widgetutil

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// IndexOf returns the index of value in options, or -1 if not found.
func IndexOf(options []string, value string) int {
	for i, o := range options {
		if o == value {
			return i
		}
	}
	return -1
}

// Labeled wraps an input under a label.
func Labeled(label string, input fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(label), input)
}

// Section creates a titled group box.
func Section(title string, rows ...fyne.CanvasObject) fyne.CanvasObject {
	header := widget.NewLabel(title)
	header.TextStyle = fyne.TextStyle{Bold: true}
	items := []fyne.CanvasObject{header}
	items = append(items, rows...)
	return container.NewVBox(items...)
}
