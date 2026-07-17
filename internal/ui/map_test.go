package ui

import (
	"fmt"
	"testing"

	"github.com/docg1701/radkeys/internal/config"
)

func TestMapGraphCapturesAllNavigateEdges(t *testing.T) {
	cfg := &config.Config{
		App: config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
		Screens: []config.Screen{
			{ID: "a", Name: "A", Buttons: []config.Button{
				{Row: 0, Col: 0, Label: "go b", Action: config.ActionNavigate, Target: "b"},
				{Row: 0, Col: 1, Label: "go c", Action: config.ActionNavigate, Target: "c"},
				{Row: 0, Col: 2, Label: "copy", Action: config.ActionCopy},
			}},
			{ID: "b", Name: "B", Buttons: []config.Button{
				{Row: 0, Col: 0, Label: "back a", Action: config.ActionNavigate, Target: "a"},
			}},
			{ID: "c", Name: "C"},
		},
	}
	g := buildMapGraph(cfg)
	if len(g.nodes) != 3 {
		t.Fatalf("nodes=%d, want 3", len(g.nodes))
	}
	if len(g.edges) != 3 { // a→b, a→c, b→a
		t.Fatalf("edges=%d, want 3 (copy button must not be an edge)", len(g.edges))
	}
}

func TestMapGraphSkipsUnknownTargets(t *testing.T) {
	cfg := &config.Config{
		App: config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
		Screens: []config.Screen{{ID: "a", Name: "A", Buttons: []config.Button{
			{Row: 0, Col: 0, Label: "ghost", Action: config.ActionNavigate, Target: "nope"},
		}}},
	}
	g := buildMapGraph(cfg)
	if len(g.edges) != 0 {
		t.Fatalf("expected 0 edges (unknown target), got %d", len(g.edges))
	}
}

func TestLayoutDeterministic(t *testing.T) {
	cfg := &config.Config{
		App:     config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
		Screens: makeLinearScreens(10),
	}
	g1 := layoutFR(buildMapGraph(cfg), 400, 300)
	g2 := layoutFR(buildMapGraph(cfg), 400, 300)
	for i := range g1.nodes {
		if g1.nodes[i].pos != g2.nodes[i].pos {
			t.Errorf("node %d position drift: %v vs %v",
				i, g1.nodes[i].pos, g2.nodes[i].pos)
		}
	}
}

func TestBreadcrumb(t *testing.T) {
	cfg := &config.Config{
		App: config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
		Screens: []config.Screen{
			{ID: "home", Name: "Home"},
			{ID: "rm", Name: "RM"},
			{ID: "mi", Name: "Medicina Interna"},
			{ID: "abd", Name: "Abdome Superior"},
			{ID: "masc", Name: "Masculino"},
		},
	}
	u := &appUI{cfg: cfg, current: "masc", stack: []string{"home", "rm", "mi", "abd"}}
	want := "Home > RM > Medicina Interna > Abdome Superior > Masculino"
	if got := u.breadcrumb(); got != want {
		t.Errorf("breadcrumb() = %q, want %q", got, want)
	}
}

func TestBreadcrumbEmptyStack(t *testing.T) {
	cfg := &config.Config{
		App:     config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
		Screens: []config.Screen{{ID: "home", Name: "Home"}},
	}
	u := &appUI{cfg: cfg, current: "home", stack: nil}
	if got := u.breadcrumb(); got != "Home" {
		t.Errorf("breadcrumb() = %q, want %q", got, "Home")
	}
}

func TestBreadcrumbUnknownIDFallsBackToRaw(t *testing.T) {
	cfg := &config.Config{
		App:     config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
		Screens: []config.Screen{{ID: "home", Name: "Home"}},
	}
	u := &appUI{cfg: cfg, current: "ghost", stack: []string{"home"}}
	want := "Home > ghost"
	if got := u.breadcrumb(); got != want {
		t.Errorf("breadcrumb() = %q, want %q", got, want)
	}
}

func makeLinearScreens(n int) []config.Screen {
	out := make([]config.Screen, n)
	for i := 0; i < n; i++ {
		s := config.Screen{ID: fmt.Sprintf("s%d", i), Name: fmt.Sprintf("Screen %d", i)}
		if i > 0 {
			s.Buttons = []config.Button{
				{Row: 0, Col: 0, Label: "prev", Action: config.ActionNavigate, Target: fmt.Sprintf("s%d", i-1)},
			}
		}
		out[i] = s
	}
	return out
}
