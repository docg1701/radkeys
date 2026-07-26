package gridframe

import (
	"testing"

	"github.com/docg1701/radkeys/internal/config"
)

func TestCaption(t *testing.T) {
	l := config.Layout{Blocks: []config.Block{{Rows: 2, Cols: 4}, {Rows: 4, Cols: 6}}}
	if got := Caption(l, 0); got != "Block 0 · n0–n7" {
		t.Errorf("Caption(0) = %q, want %q", got, "Block 0 · n0–n7")
	}
	if got := Caption(l, 1); got != "Block 1 · n8–n31" {
		t.Errorf("Caption(1) = %q, want %q", got, "Block 1 · n8–n31")
	}
}
