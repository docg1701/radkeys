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

func TestIndexOf(t *testing.T) {
	opts := []string{"a", "b", "c"}
	if got := IndexOf(opts, "b"); got != 1 {
		t.Fatalf("IndexOf(..., \"b\") = %d, want 1", got)
	}
	if got := IndexOf(opts, "z"); got != -1 {
		t.Fatalf("IndexOf(..., \"z\") = %d, want -1", got)
	}
}
