package ui

import (
	"math"
	"math/rand"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
)

const (
	mapNodeW      = 12
	mapNodeH      = 12
	mapPad        = 24
	mapPanelWidth = 260
)

// mapNode is a renderable screen with the position assigned by the layout
// step. The id drives navigation on click; the name is only used by the
// breadcrumb, not the renderer.
type mapNode struct {
	id   string
	name string
	pos  fyne.Position
}

type mapGraph struct {
	nodes []mapNode
	edges [][2]int // indices into nodes: [from, to]
}

// buildMapGraph walks cfg: one node per screen, one edge per ActionNavigate
// button. Targets that don't resolve are skipped (config.Issues() reports
// them at load time).
func buildMapGraph(cfg *config.Config) mapGraph {
	g := mapGraph{nodes: make([]mapNode, 0, len(cfg.Screens))}
	idx := make(map[string]int, len(cfg.Screens))
	for _, s := range cfg.Screens {
		idx[s.ID] = len(g.nodes)
		g.nodes = append(g.nodes, mapNode{id: s.ID, name: s.Name})
	}
	for _, s := range cfg.Screens {
		for _, b := range s.Buttons {
			if b.Action != config.ActionNavigate {
				continue
			}
			to, ok := idx[b.Target]
			if !ok {
				continue
			}
			g.edges = append(g.edges, [2]int{idx[s.ID], to})
		}
	}
	return g
}

// layoutFR assigns each node a position inside the (w, h) box. Seeded so
// the same graph always produces the same positions (test-friendly).
func layoutFR(g mapGraph, w, h float64) mapGraph {
	rng := rand.New(rand.NewSource(42))
	n := len(g.nodes)
	if n == 0 {
		return g
	}
	pos := make([][2]float64, n)
	for i := range pos {
		pos[i] = [2]float64{rng.Float64() * w, rng.Float64() * h}
	}
	k := math.Sqrt((w * h) / float64(n))
	t := w / 10
	const iters = 200
	for iter := 0; iter < iters; iter++ {
		disp := make([][2]float64, n)
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				if i == j {
					continue
				}
				dx := pos[i][0] - pos[j][0]
				dy := pos[i][1] - pos[j][1]
				d := math.Max(math.Hypot(dx, dy), 0.01)
				fr := (k * k) / d
				disp[i][0] += (dx / d) * fr
				disp[i][1] += (dy / d) * fr
			}
		}
		for _, e := range g.edges {
			u, v := e[0], e[1]
			dx := pos[u][0] - pos[v][0]
			dy := pos[u][1] - pos[v][1]
			d := math.Max(math.Hypot(dx, dy), 0.01)
			fa := (d * d) / k
			disp[u][0] -= (dx / d) * fa
			disp[u][1] -= (dy / d) * fa
			disp[v][0] += (dx / d) * fa
			disp[v][1] += (dy / d) * fa
		}
		for i := 0; i < n; i++ {
			d := math.Max(math.Hypot(disp[i][0], disp[i][1]), 0.01)
			pos[i][0] = clamp(pos[i][0]+disp[i][0]/d*math.Min(d, t), 0, w)
			pos[i][1] = clamp(pos[i][1]+disp[i][1]/d*math.Min(d, t), 0, h)
		}
		t *= 0.95
	}
	for i, p := range pos {
		g.nodes[i].pos = fyne.NewPos(float32(p[0]), float32(p[1]))
	}
	return g
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ---------------------------------------------------------------------------
// Custom Fyne widget
// ---------------------------------------------------------------------------

type mapWidget struct {
	widget.BaseWidget
	graph     mapGraph
	currentID string
	renderer  *mapRenderer // cached so SetCurrentScreen can mutate dot colors
	theme     fyne.Theme
	variant   fyne.ThemeVariant
}

func newMapWidget(cfg *config.Config) *mapWidget {
	w := &mapWidget{graph: layoutFR(buildMapGraph(cfg), mapPanelWidth-mapPad*2, 300-mapPad*2)}
	w.ExtendBaseWidget(w)
	return w
}

func (m *mapWidget) SetTheme(th fyne.Theme, v fyne.ThemeVariant) {
	m.theme, m.variant = th, v
	m.Refresh()
}

func (m *mapWidget) CreateRenderer() fyne.WidgetRenderer {
	m.renderer = &mapRenderer{m: m}
	return m.renderer
}

func (m *mapWidget) MinSize() fyne.Size {
	return fyne.NewSize(mapPanelWidth, 300)
}

// SetCurrentScreen updates the highlight. Mutates the existing dot's
// FillColor in place — does not rebuild the objects list.
func (m *mapWidget) SetCurrentScreen(id string) {
	if id == m.currentID {
		return
	}
	r := m.renderer
	if r == nil {
		m.currentID = id
		m.Refresh()
		return
	}
	th, v := r.m.theme, r.m.variant
	if th == nil {
		th = fyneTheme.DefaultTheme()
		v = fyneTheme.VariantDark
	}
	normal := th.Color(fyneTheme.ColorNameForeground, v)
	hi := th.Color(fyneTheme.ColorNamePrimary, v)
	if old, ok := r.nodeObjs[m.currentID]; ok {
		old.FillColor = normal
		canvas.Refresh(old)
	}
	if newDot, ok := r.nodeObjs[id]; ok {
		newDot.FillColor = hi
		canvas.Refresh(newDot)
	}
	m.currentID = id
}

// ---------------------------------------------------------------------------
// Renderer
// ---------------------------------------------------------------------------

type mapRenderer struct {
	m        *mapWidget
	objects  []fyne.CanvasObject
	nodeObjs map[string]*canvas.Circle // id -> dot, for in-place color swap
	lineObjs []canvas.Line             // parallel to m.graph.edges
}

func (r *mapRenderer) MinSize() fyne.Size { return r.m.MinSize() }

func (r *mapRenderer) Layout(size fyne.Size) {
	w, h := float64(size.Width-mapPad*2), float64(size.Height-mapPad*2)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	for id, c := range r.nodeObjs {
		var p fyne.Position
		for _, n := range r.m.graph.nodes {
			if n.id == id {
				p = n.pos
				break
			}
		}
		p.X = mapPad + float32(clamp(float64(p.X), 0, w))
		p.Y = mapPad + float32(clamp(float64(p.Y), 0, h))
		c.Resize(fyne.NewSize(mapNodeW, mapNodeH))
		c.Move(p)
	}
	for i, line := range r.lineObjs {
		e := r.m.graph.edges[i]
		a := r.m.graph.nodes[e[0]].pos.Add(fyne.NewPos(mapPad+mapNodeW/2, mapPad+mapNodeH/2))
		b := r.m.graph.nodes[e[1]].pos.Add(fyne.NewPos(mapPad+mapNodeW/2, mapPad+mapNodeH/2))
		line.Position1, line.Position2 = a, b
	}
}

func (r *mapRenderer) Refresh() {
	r.objects = r.objects[:0]
	r.nodeObjs = make(map[string]*canvas.Circle, len(r.m.graph.nodes))
	r.lineObjs = r.lineObjs[:0]

	th, v := r.m.theme, r.m.variant
	if th == nil {
		th = fyneTheme.DefaultTheme()
		v = fyneTheme.VariantDark
	}
	nodeColor := th.Color(fyneTheme.ColorNameForeground, v)
	edgeColor := th.Color(fyneTheme.ColorNameDisabled, v)
	hi := th.Color(fyneTheme.ColorNamePrimary, v)

	// Edges first (drawn under nodes).
	for _, e := range r.m.graph.edges {
		line := canvas.NewLine(edgeColor)
		line.StrokeWidth = 1.5
		a := r.m.graph.nodes[e[0]].pos.Add(fyne.NewPos(mapPad+mapNodeW/2, mapPad+mapNodeH/2))
		b := r.m.graph.nodes[e[1]].pos.Add(fyne.NewPos(mapPad+mapNodeW/2, mapPad+mapNodeH/2))
		line.Position1, line.Position2 = a, b
		r.lineObjs = append(r.lineObjs, *line)
		r.objects = append(r.objects, line)
	}
	// Dots.
	for _, n := range r.m.graph.nodes {
		fill := nodeColor
		if n.id == r.m.currentID {
			fill = hi
		}
		c := canvas.NewCircle(fill)
		c.Resize(fyne.NewSize(mapNodeW, mapNodeH))
		c.Move(fyne.NewPos(mapPad+n.pos.X, mapPad+n.pos.Y))
		r.nodeObjs[n.id] = c
		r.objects = append(r.objects, c)
	}
	canvas.Refresh(r.m)
}

func (r *mapRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *mapRenderer) Destroy()                     {}
