package editor

import (
	"testing"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
)

func TestActionOptionsCount(t *testing.T) {
	labels := actionLabels()
	if len(labels) != len(actionDefs) {
		t.Fatalf("expected %d action labels, got %d", len(actionDefs), len(labels))
	}
}

func TestActionLabelPaste(t *testing.T) {
	want := i18n.T("button.paste")
	if got := actionLabelByID(config.ActionPaste); got != want {
		t.Fatalf("actionLabelByID(paste) = %q, want %q", got, want)
	}
}

func TestActionFromLabelPaste(t *testing.T) {
	label := i18n.T("button.paste")
	if got := actionIDByLabel(label); got != config.ActionPaste {
		t.Fatalf("actionIDByLabel(%q) = %q, want %q", label, got, config.ActionPaste)
	}
}
