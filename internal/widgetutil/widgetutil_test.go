package widgetutil

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

func TestLabeledContainsLabel(t *testing.T) {
	inp := widget.NewEntry()
	obj := Labeled("Name", inp)
	if obj == nil {
		t.Fatal("Labeled returned nil")
	}
}

func TestSectionContainsHeader(t *testing.T) {
	child := widget.NewLabel("child")
	obj := Section("Header", child)
	if obj == nil {
		t.Fatal("Section returned nil")
	}
}
