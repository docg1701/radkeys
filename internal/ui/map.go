package ui

import (
	"math"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
)

const (
	mapNodeW       = 12
	mapNodeH       = 12
	mapNodeSpacing = 24 // 12px dot + 12px gap between centers
	mapPad         = 24
	mapMinWidth    = 260 // floor when the graph has few nodes
	mapRowH        = 32  // vertical spacing between rows
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
	nodes       []mapNode
	edges       [][2]int // indices into nodes: [from, to]
	maxPerLevel int      // widest level — drives min panel width
	totalRows   int      // sub-rows after wrapping — drives min height
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

// layoutLayered assigns node positions as a vertical cascade. The first
// screen (root) sits at the top; each BFS depth becomes a row below it.
// Nodes at the same depth spread evenly across the available width.
// Rows are distributed evenly across the full height h so the graph
// always fills the panel, never squishes at the top.
func layoutLayered(g mapGraph, w, h float64) mapGraph {
	n := len(g.nodes)
	if n == 0 {
		return g
	}

	// adjacency list from edges
	children := make([][]int, n)
	for _, e := range g.edges {
		children[e[0]] = append(children[e[0]], e[1])
	}

	// BFS from node 0 (the first screen = home)
	depth := make([]int, n)
	for i := range depth {
		depth[i] = -1
	}
	queue := []int{0}
	depth[0] = 0
	maxDepth := 0
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range children[u] {
			if depth[v] == -1 {
				depth[v] = depth[u] + 1
				if depth[v] > maxDepth {
					maxDepth = depth[v]
				}
				queue = append(queue, v)
			}
		}
	}

	// unvisited nodes go to a final row below everything
	for i := range depth {
		if depth[i] == -1 {
			maxDepth++
			depth[i] = maxDepth
		}
	}

	// group nodes by depth, track widest level
	levels := make([][]int, maxDepth+1)
	for i, d := range depth {
		levels[d] = append(levels[d], i)
	}

	// Build flat list of rows (a depth may span multiple rows if too wide).
	type row struct {
		nodes []int
	}
	var rows []row
	maxNodesPerRow := int(math.Max(1, w/mapNodeSpacing))
	for d := 0; d <= maxDepth; d++ {
		ids := levels[d]
		if len(ids) == 0 {
			continue
		}
		sorted := make([]int, len(ids))
		copy(sorted, ids)
		slices.Sort(sorted)
		if len(ids) > g.maxPerLevel {
			g.maxPerLevel = len(ids)
		}
		for start := 0; start < len(sorted); start += maxNodesPerRow {
			end := start + maxNodesPerRow
			if end > len(sorted) {
				end = len(sorted)
			}
			rows = append(rows, row{nodes: sorted[start:end]})
		}
	}

	// Distribute rows evenly across the full height h so the graph always
	// fills the panel, never squishes at the top.
	g.totalRows = len(rows)
	usableH := h - mapPad*2
	if usableH < mapNodeH {
		usableH = mapNodeH
	}
	rowSpacing := float64(mapRowH) // fallback for zero-height edge case
	if len(rows) > 0 {
		rowSpacing = usableH / float64(len(rows))
	}
	// Position each sub-row: center of each row band minus half node height.
	for i, r := range rows {
		y := mapPad + float64(i)*rowSpacing + rowSpacing/2 - mapNodeH/2
		colW := w / float64(len(r.nodes))
		for j, id := range r.nodes {
			x := float64(j)*colW + colW/2 - mapNodeW/2
			g.nodes[id].pos = fyne.NewPos(float32(x), float32(y))
		}
	}
	return g
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
	g := buildMapGraph(cfg)
	// Provisional layout — real positions are computed in the renderer's Layout().
	g = layoutLayered(g, float64(mapMinWidth-mapPad*2), 300-mapPad*2)
	w := &mapWidget{graph: g}
	w.ExtendBaseWidget(w)
	return w
}

// relayout re-runs the layered layout with the given canvas size so node
// positions always match the actual panel dimensions.
func (m *mapWidget) relayout(w, h float32) {
	m.graph = layoutLayered(m.graph, float64(w-mapPad*2), float64(h-mapPad*2))
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
	// Width: enough to fit the widest level without wrapping.
	w := float32(mapMinWidth)
	if m.graph.maxPerLevel > 0 {
		need := float32(m.graph.maxPerLevel)*mapNodeSpacing + mapPad*2
		if need > w {
			w = need
		}
	}
	// Height: enough rows to show the full graph.
	h := float32(m.graph.totalRows)*mapRowH + mapPad*2
	if h < 200 {
		h = 200
	}
	return fyne.NewSize(w, h)
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
	lineObjs []*canvas.Line            // parallel to m.graph.edges
}

func (r *mapRenderer) MinSize() fyne.Size { return r.m.MinSize() }

func (r *mapRenderer) Layout(size fyne.Size) {
	r.m.relayout(size.Width, size.Height)
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
		p.X = mapPad + float32(min(max(float64(p.X), 0), w))
		p.Y = mapPad + float32(min(max(float64(p.Y), 0), h))
		c.Resize(fyne.NewSize(mapNodeW, mapNodeH))
		c.Move(p)
	}
	for i, ln := range r.lineObjs {
		e := r.m.graph.edges[i]
		a := r.m.graph.nodes[e[0]].pos.Add(fyne.NewPos(mapPad+mapNodeW/2, mapPad+mapNodeH/2))
		b := r.m.graph.nodes[e[1]].pos.Add(fyne.NewPos(mapPad+mapNodeW/2, mapPad+mapNodeH/2))
		ln.Position1 = a
		ln.Position2 = b
		ln.Refresh()
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
		r.lineObjs = append(r.lineObjs, line)
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
