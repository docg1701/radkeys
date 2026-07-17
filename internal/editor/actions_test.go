package editor

import (
	"testing"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

func TestActionOptionsCount(t *testing.T) {
	labels := config.ActionLabels()
	if len(labels) != len(config.ActionList) {
		t.Fatalf("expected %d action labels, got %d", len(config.ActionList), len(labels))
	}
}

func TestActionLabelPaste(t *testing.T) {
	want := i18n.T("action.paste")
	if got := config.ActionLabel(config.ActionPaste); got != want {
		t.Fatalf("ActionLabel(paste) = %q, want %q", got, want)
	}
}

func TestActionFromLabelPaste(t *testing.T) {
	label := i18n.T("action.paste")
	if got := config.ActionIDFromLabel(label); got != config.ActionPaste {
		t.Fatalf("ActionIDFromLabel(%q) = %q, want %q", label, got, config.ActionPaste)
	}
}
