package gridframe

import (
	"testing"

	"github.com/docg1701/radkeys/internal/config"
)

func TestCaption(t *testing.T) {
	l := config.Layout{Blocks: []config.Block{{Rows: 2, Cols: 4}, {Name: "Func", Rows: 4, Cols: 6}}}
	if got := Caption(l, 0); got != "Block 0 (0–7)" {
		t.Errorf("Caption(0) = %q, want %q", got, "Block 0 (0–7)")
	}
	if got := Caption(l, 1); got != "Func (8–31)" {
		t.Errorf("Caption(1) = %q, want %q", got, "Func (8–31)")
	}
}
