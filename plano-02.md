# Implementation Plan — Screen-Map Panel + Breadcrumb Header

## Goal

Add a **side panel** on the Shortcuts tab that renders the navigation graph
as dots-and-lines only (no labels, no glow), and a **breadcrumb header**
above the AppTabs that shows the `>`-separated path of screen names from
`u.stack + [u.current]`. Both update in real time. Zero new dependencies.

---

## Spec recap (per task description)

- **Map panel (right side, inside the Shortcuts tab):**
  collapsible via chevron toggle, resizable via HSplit divider, vertical
  scroll when content overflows. Content = one dot per screen + one line
  per `ActionNavigate` edge. No labels, no text, no edge arrows. Current
  screen = colour change on the dot only.
- **Breadcrumb header (above the AppTabs):**
  `[breadcrumb] | [device-status]`. Breadcrumb = `reverse(u.stack) + [u.current]`,
  with each id replaced by its `Screen.Name`, joined by `>`. Updates whenever
  `u.current` or `u.stack` change.
- **Status moves into the header** (was already in the top border slot per
  the current code at `internal/ui/ui.go:624`; the new design merges it with
  the breadcrumb on the same row).

## What this plan drops vs. the previous one

- ❌ Label rendering (text inside each node). Gone.
- ❌ Glow ring / pulsing highlight for the current node. Gone — colour change only.
- ❌ A `Map` tab inside `AppTabs`. Replaced by a side panel inside the Shortcuts tab.
- ⚠️ Click-to-navigate is **kept** (research brief confirms widget-level
  `Tappable` + hit-test is the lightest pattern; works fine in Fyne v2).
- ⚠️ Force-directed layout chosen over Sugiyama layered: simpler (~80 lines),
  handles cycles, no new dependency; spec is a small graph (<50 nodes today,
  ~190 with full TOML expansion).

## Layout algorithm: DIY Fruchterman-Reingold (FR), seeded

- Nav graph is a small DAG (or near-DAG) — FR works fine.
- Zero deps, ~80 lines, identical to the previous plan.
- Seeded with `rand.NewSource(42)` so `TestLayoutDeterministic` can pin positions.
- 200 iters, k = sqrt(area/n), cool by 0.95. Linear O(n²) repulsion.
- Ponytail ceiling: O(n²) is fine up to ~200 nodes; add Barnes-Hut if it ever
  shows up as a hot spot in profiling. Don't pre-empt.

## Out of scope (YAGNI)

- No persistence of the map collapse state across launches.
- No edge labels or edge arrows.
- No zoom/pan inside the scrollable map.
- No new i18n keys (status text is already in i18n; chevron is icon-only;
  breadcrumb uses existing `Screen.Name` strings).
- No new go.mod entries.

---

## Tasks

Grouped by file. Each task lists: file:line, before/after sketch, risk,
validation.

### Task 1 — Data model + graph builder

**File:** `internal/ui/map.go` (new file)
**Risk:** low — pure function over `*config.Config`, no Fyne types.
**Validation:** `go test ./internal/ui/... -run TestMapGraph -v` passes.

Before (no map code exists): nothing.

After (~50 lines):

```go
package ui

import (
    "fyne.io/fyne/v2"
    "github.com/docg1701/radkeys/internal/config"
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
```

**ponytail:** node/edge types live next to the widget that renders them.
No `internal/graph` package until a second consumer appears.

---

### Task 2 — Layout (Fruchterman-Reingold, seeded)

**File:** `internal/ui/map.go` (continue)
**Risk:** medium — algorithmic but bounded.
**Validation:** `TestLayoutDeterministic` (Task 7).

```go
import (
    "math"
    "math/rand"
)

const (
    mapNodeW = 12 // dot diameter in pixels; small for a clean side panel
    mapNodeH = 12
    mapPad  = 24  // outer padding inside the widget
)

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
```

**ponytail:** 200 iters, seeded `math/rand`, O(n²) repulsion. If profiling
later shows >16ms on 200 nodes, add Barnes-Hut quadtree. Don't pre-empt.

---

### Task 3 — Custom widget skeleton

**File:** `internal/ui/map.go` (continue)
**Risk:** medium — Fyne widget lifecycle; get `ExtendBaseWidget` and the
renderer interface right.
**Validation:** smoke-render in the app, watch for redraw-loop warnings.

```go
import (
    "fyne.io/fyne/v2/canvas"
    "fyne.io/fyne/v2/widget"
)

const mapPanelWidth = 260

type mapWidget struct {
    widget.BaseWidget
    graph      mapGraph
    currentID  string
    onNavigate func(id string) // set by appUI once at construction
    theme      fyne.Theme
    variant    fyne.ThemeVariant
}

func newMapWidget(cfg *config.Config) *mapWidget {
    w := &mapWidget{graph: buildMapGraph(cfg)}
    w.ExtendBaseWidget(w)
    return w
}

func (m *mapWidget) SetTheme(th fyne.Theme, v fyne.ThemeVariant) {
    m.theme, m.variant = th, v
    m.Refresh()
}

func (m *mapWidget) CreateRenderer() fyne.WidgetRenderer {
    return &mapRenderer{m: m}
}

func (m *mapWidget) MinSize() fyne.Size {
    return fyne.NewSize(mapPanelWidth, 300)
}
```

`CreateRenderer` returns a renderer (Task 4).

---

### Task 4 — Renderer (canvas.Circle + canvas.Line only — no text)

**File:** `internal/ui/map.go` (continue)
**Risk:** medium-high — Fyne renderer API quirks (objects list rebuild,
in-place mutation for highlights, no double-buffering).
**Validation:** manual: open the app, dots + lines render, drag the divider,
switch screens, verify only the current dot's colour changes.

```go
type mapRenderer struct {
    m        *mapWidget
    objects  []fyne.CanvasObject
    nodeObjs map[string]*canvas.Circle // id -> dot, for in-place colour swap
    lineObjs []canvas.Line             // parallel to m.graph.edges
}

func (r *mapRenderer) MinSize() fyne.Size { return r.m.MinSize() }

func (r *mapRenderer) Layout(size fyne.Size) {
    // Reposition existing primitives inside the new size. Positions were
    // computed once by layoutFR at construction; we clamp them here so a
    // narrower panel doesn't push dots off-screen.
    w, h := float64(size.Width-mapPad*2), float64(size.Height-mapPad*2)
    if w < 1 {
        w = 1
    }
    if h < 1 {
        h = 1
    }
    if rect, ok := r.nodeObjs[r.m.currentID]; ok {
        _ = rect
    }
    for id, c := range r.nodeObjs {
        var p fyne.Position
        for _, n := range r.m.graph.nodes {
            if n.id == id {
                p = n.pos
                break
            }
        }
        p.X = mapPad + clamp(float64(p.X), 0, w)
        p.Y = mapPad + clamp(float64(p.Y), 0, h)
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
func (r *mapRenderer) Destroy()                      {}
```

**ponytail:** In-place highlight (mutate `c.FillColor` + `canvas.Refresh(c)`)
in `SetCurrentScreen` — avoids rebuilding N circles on every key press.

**ponytail:** No labels (per the new spec). Previous plan's `canvas.Text` is
dropped. Node size is 12px; small but visible at typical 1280px window widths.

---

### Task 5 — Click handling (Tappable + hit-test)

**File:** `internal/ui/map.go` (continue)
**Risk:** low — `Tappable` is the documented Fyne hook.
**Validation:** manual: click a dot, the keypad grid updates; the breadcrumb
header updates; the dot's colour changes.

```go
// Tapped satisfies fyne.Tappable. Hit-test against node bounding boxes.
// ponytail: linear scan; with ~100 nodes, <1µs.
func (m *mapWidget) Tapped(ev *fyne.PointEvent) {
    for _, n := range m.graph.nodes {
        x0, y0 := float32(mapPad)+n.pos.X, float32(mapPad)+n.pos.Y
        if ev.Position.X >= x0 && ev.Position.X <= x0+mapNodeW &&
            ev.Position.Y >= y0 && ev.Position.Y <= y0+mapNodeH {
            if m.onNavigate != nil && n.id != m.currentID {
                m.onNavigate(n.id)
            }
            return
        }
    }
}

// SetCurrentScreen updates the highlight. Mutates the existing dot's
// FillColor in place — does not rebuild the objects list.
func (m *mapWidget) SetCurrentScreen(id string) {
    if id == m.currentID {
        return
    }
    r, ok := m.Renderer().(*mapRenderer)
    if !ok {
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
```

**ponytail:** widget-level `Tappable`, not per-dot `widget.Button` overlay.
100 buttons in the Fyne tree = 100 layout passes; the hit-test is one tight
loop.

**If Fyne refuses to deliver `Tapped` events to the custom widget** (the
research brief's "canvas primitives are non-interactive" caveat), fall
back to **drop click-to-navigate** and document the limitation in a
`// ponytail:` comment. The spec marks it as optional. Don't add per-dot
buttons as a workaround — 50+ button widgets for a decorative panel is
overkill.

---

### Task 6 — Side panel integration in the Shortcuts tab

**File:** `internal/ui/ui.go:601-625` (`rebuildTabs`)
**Risk:** medium — must not break the existing `preview | keypad` vertical
split. Nested splits add layout interactions.
**Validation:** drag the keypad/preview divider — still works. Drag the new
map divider — works. Collapse + expand the map — works.

Before (line 601-625):
```go
func (u *appUI) rebuildTabs() {
    selectedIdx := 0
    if u.tabs != nil {
        selectedIdx = u.tabs.SelectedIndex()
    }

    previewArea := u.previewBox()
    keypadArea := container.NewPadded(u.keypad)
    split := container.NewVSplit(previewArea, keypadArea)

    u.tabs = container.NewAppTabs(
        container.NewTabItem(i18n.T("tab.shortcuts"), split),
        container.NewTabItem(i18n.T("tab.settings"), u.buildSettings()),
        container.NewTabItem(i18n.T("tab.about"), u.buildAbout()),
    )
    u.tabs.SelectIndex(selectedIdx)
    u.win.SetContent(container.NewBorder(u.status, nil, nil, nil, u.tabs))
}
```

After:
```go
func (u *appUI) rebuildTabs() {
    selectedIdx := 0
    if u.tabs != nil {
        selectedIdx = u.tabs.SelectedIndex()
    }

    previewArea := u.previewBox()
    keypadArea := container.NewPadded(u.keypad)
    main := container.NewVSplit(previewArea, keypadArea)
    shortcuts := u.shortcutsTab(main) // new helper — wraps main in HSplit with map

    u.tabs = container.NewAppTabs(
        container.NewTabItem(i18n.T("tab.shortcuts"), shortcuts),
        container.NewTabItem(i18n.T("tab.settings"), u.buildSettings()),
        container.NewTabItem(i18n.T("tab.about"), u.buildAbout()),
    )
    u.tabs.SelectIndex(selectedIdx)
    u.win.SetContent(container.NewBorder(u.headerBar(), nil, nil, nil, u.tabs))
}
```

New helpers on `appUI` (in `ui.go`):

```go
// shortcutsTab wraps the existing preview/keypad split with the
// collapsible map panel on the right.
func (u *appUI) shortcutsTab(main *container.Split) fyne.CanvasObject {
    if u.map == nil {
        u.map = newMapWidget(u.cfg)
        u.map.onNavigate = u.navigateTo
    }
    th, v := u.a.Settings().Theme(), u.a.Settings().ThemeVariant()
    u.map.SetTheme(th, v)

    if u.mapSplit == nil {
        u.mapSplit = container.NewHSplit(main, container.NewVScroll(u.map))
        u.mapSplit.Offset = mapOffsetExpanded // 0.75
    }
    return u.mapSplit
}

// mapCollapseBtn toggles the right panel. Lives in the panel header (not
// on the HSplit divider — Fyne has no click-to-collapse API).
func (u *appUI) mapCollapseBtn() *widget.Button {
    icon := fyneTheme.IconNameNavigateBack // ◀ when panel is open
    if !u.mapVisible {
        icon = fyneTheme.IconNameNavigateNext // ▶ when collapsed
    }
    return widget.NewButtonWithIcon("", theme.NewThemedResource(icon), func() {
        u.mapVisible = !u.mapVisible
        if u.mapSplit != nil {
            if u.mapVisible {
                u.mapSplit.Offset = mapOffsetExpanded
            } else {
                u.mapSplit.Offset = mapOffsetCollapsed // 1.0
            }
            u.mapSplit.Refresh()
        }
    })
}

const (
    mapOffsetCollapsed = 1.0
    mapOffsetExpanded  = 0.75
)
```

**ponytail:** The map's `NewVScroll` wraps the map widget; the scrollbar
appears only when the rendered content overflows the panel's height.

**ponytail:** The collapse button is icon-only — no new i18n key needed.
The spec says the chevron is a toggle, not a labelled control.

**ponytail:** `u.mapSplit` is cached on `appUI` (Task 7) so the toggle
button can mutate its `Offset` without rebuilding tabs.

---

### Task 7 — `appUI` struct additions

**File:** `internal/ui/ui.go:132-152` (`appUI` struct)
**Risk:** low — additive fields.
**Validation:** `go build ./...` clean.

Before (excerpt):
```go
type appUI struct {
    cfg         *config.Config
    configPath  string
    current     string
    stack       []string
    device      hid.Device
    a           fyne.App
    win         fyne.Window
    titleBase   string
    preview     *widget.Label
    previewText string
    version     string
    mock        bool
    closing     atomic.Bool
    status      *widget.Label
    flashTimer  *time.Timer
    tabs        *container.AppTabs
    cols        int
    rows        int
    keypad      *fyne.Container
    previewBg   *canvas.Rectangle
}
```

After (add three fields):
```go
    tabs        *container.AppTabs
    cols        int
    rows        int
    keypad      *fyne.Container
    previewBg   *canvas.Rectangle
    map         *mapWidget         // NEW
    mapSplit    *container.Split   // NEW: cached HSplit for the side panel
    mapVisible  bool               // NEW: true when panel is shown
}
```

---

### Task 8 — Breadcrumb helper

**File:** `internal/ui/ui.go` (new method on `appUI`)
**Risk:** low — pure function over `u.stack` + `u.current` + `u.cfg.Screens`.
**Validation:** `TestBreadcrumb` (Task 10).

```go
// breadcrumb returns the `>`-separated path of screen names from the back
// stack to the current screen, e.g. "Home > RM > Medicina Interna > Abdome".
// Unknown ids fall back to the raw id (never empty — visible in the UI).
func (u *appUI) breadcrumb() string {
    names := make([]string, 0, len(u.stack)+1)
    idToName := make(map[string]string, len(u.cfg.Screens))
    for _, s := range u.cfg.Screens {
        idToName[s.ID] = s.Name
    }
    for _, id := range u.stack {
        if name, ok := idToName[id]; ok {
            names = append(names, name)
        } else {
            names = append(names, id) // unknown id — show raw, no panic
        }
    }
    cur := u.current
    if name, ok := idToName[cur]; ok {
        names = append(names, name)
    } else {
        names = append(names, cur)
    }
    return strings.Join(names, " > ")
}
```

**Add import:** `"strings"` if not already present. (`grep -n '^import\|\"strings\"' internal/ui/ui.go` to confirm — currently absent.)

**ponytail:** no caching — the helper runs O(screens) per call. With ~50
screens, the call is sub-microsecond. Don't pre-empt.

---

### Task 9 — Header bar (above AppTabs)

**File:** `internal/ui/ui.go` (new method `headerBar`)
**Risk:** medium — replaces the existing top-border slot where `u.status`
lives; the device-status message must keep its current behaviour.
**Validation:** manual: in mock mode, the status string appears on the right
of the header. In real mode, no message (the label is hidden/empty).

```go
// headerBar is the top-of-window row: breadcrumb on the left, a `|`
// separator, and the device-status message on the right. Replaces the
// previous top-border that held only u.status.
func (u *appUI) headerBar() fyne.CanvasObject {
    breadcrumb := widget.NewLabel(u.breadcrumb())
    breadcrumb.TextStyle = fyne.TextStyle{Italic: true}
    sep := widget.NewLabel("|")
    return container.NewBorder(nil, nil, nil, u.status,
        container.NewHBox(breadcrumb, sep),
    )
}
```

**Behaviour notes:**
- The current `u.status` is created at `ui.go:75` and used by `setStatus` /
  `flashStatus` (`ui.go:339-368`). No change to those methods.
- When the device is in real mode and there's no message, `u.status` is
  hidden (existing behaviour: `u.status.Hide()` after timer fires). The
  right border slot then collapses to zero width — the breadcrumb fills
  the row. Visually clean.
- The `border(top=nothing, bottom=nothing, left=nothing, right=u.status,
  center=HBox(breadcrumb, sep))` lays out as: breadcrumb+sep fill the
  centre, status is right-aligned. If status is hidden, the breadcrumb
  expands to full width.

**ponytail:** Border-with-right-element is the simplest way to get
"left content + right content" in a Fyne row. Two `HBox` children with a
`layout.NewSpacer()` would also work; the Border pattern matches the
existing style at `ui.go:624` (the previous `NewBorder(u.status, ...)` call).

---

### Task 10 — Real-time refresh

**File:** `internal/ui/ui.go:293` (`renderGrid`)
**Risk:** low — additive; `renderGrid()` is already called at every site
where `u.current` or `u.stack` change.
**Validation:** manual: switch screens, both the breadcrumb and the map
highlight update within the same frame as the keypad.

Before (top of `renderGrid`, line 293-294):
```go
func (u *appUI) renderGrid() {
    s := u.currentScreen()
```

After:
```go
func (u *appUI) renderGrid() {
    if u.map != nil {
        u.map.SetCurrentScreen(u.current)
    }
    s := u.currentScreen()
```

And the breadcrumb label is refreshed via the same path. Since the header
bar calls `u.breadcrumb()` only when constructed, it won't update on
navigation. Two fixes — pick **one**:

**Option A (cheapest, preferred):** make the breadcrumb a `widget.Label`
created once, and refresh its text inside `renderGrid`:

```go
// In appUI struct:
    breadcrumbLabel *widget.Label

// In headerBar:
    u.breadcrumbLabel = widget.NewLabel(u.breadcrumb())
    u.breadcrumbLabel.TextStyle = fyne.TextStyle{Italic: true}

// At top of renderGrid, after the map refresh:
    if u.breadcrumbLabel != nil {
        u.breadcrumbLabel.SetText(u.breadcrumb())
    }
```

**Option B:** wrap breadcrumb in a custom widget that re-reads
`u.breadcrumb()` on `Refresh()`. More work, no clear win.

**Plan uses Option A.**

---

### Task 11 — Tests

**File:** `internal/ui/map_test.go` (new file)
**Risk:** low — pure logic, no Fyne widget test harness.
**Validation:** `go test ./internal/ui/... -v` passes (existing tests still
green; new tests cover graph builder, layout, breadcrumb).

```go
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
        App:     config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
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
        App: config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
        Screens: []config.Screen{{ID: "home", Name: "Home"}},
    }
    u := &appUI{cfg: cfg, current: "home", stack: nil}
    if got := u.breadcrumb(); got != "Home" {
        t.Errorf("breadcrumb() = %q, want %q", got, "Home")
    }
}

func TestBreadcrumbUnknownIDFallsBackToRaw(t *testing.T) {
    cfg := &config.Config{
        App: config.App{Layout: config.Layout{Columns: 6, Rows: 6}},
        Screens: []config.Screen{{ID: "home", Name: "Home"}},
    }
    u := &appUI{cfg: cfg, current: "ghost", stack: []string{"home"}}
    // Stack id is known; current id is unknown — should not panic, shows raw.
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
```

**Skip widget-level Fyne tests** — no headless harness in the project
(`focus_invariant_test.go` is purely static AST). Graph builder, layout,
and breadcrumb are the testable surface; the renderer is verified manually.

---

### Task 12 — `osThemeSettledListener` parity (theme update on real screens)

**File:** `internal/ui/ui.go:84-95` (`osThemeSettledListener`)
**Risk:** low — additive; the listener currently refreshes `u.previewBg`.
**Validation:** manual: change theme in settings; map dots/lines pick up
the new colour without restart.

Before (line 84-95):
```go
func (u *appUI) osThemeSettledListener(s fyne.Settings) {
    th := s.Theme()
    if _, ok := th.(themes.CustomThemeMarker); ok {
        return
    }
    v := variantFor(th, u.a.Settings().ThemeVariant())
    if u.previewBg != nil {
        u.previewBg.FillColor = th.Color(fyneTheme.ColorNameBackground, v)
        canvas.Refresh(u.previewBg)
    }
    u.renderGrid()
}
```

After: add two lines before `u.renderGrid()`:
```go
    if u.map != nil {
        u.map.SetTheme(th, v)
    }
    u.renderGrid()
```

`applySettings` (`ui.go:560-588`) does the same: theme reapplication calls
`SetTheme` on the preview, the map, then `renderGrid`. Both call sites
need the same map-theme update for symmetry.

---

## Files to Modify

- `internal/ui/ui.go`
  - struct (`~line 132-152`): add `map *mapWidget`, `mapSplit *container.Split`,
    `mapVisible bool`, `breadcrumbLabel *widget.Label`.
  - `rebuildTabs` (`~line 601-625`): use `u.headerBar()` for the top border;
    wrap the Shortcuts tab's `VSplit` in `u.shortcutsTab()`.
  - `osThemeSettledListener` (`~line 84`): re-`SetTheme` the map before `renderGrid`.
  - `applySettings` (`~line 560-588`): same — re-`SetTheme` the map.
  - `renderGrid` (`~line 293`): prepend map highlight + breadcrumb refresh.
  - new methods: `shortcutsTab`, `mapCollapseBtn`, `headerBar`, `breadcrumb`.
  - new constants: `mapOffsetCollapsed = 1.0`, `mapOffsetExpanded = 0.75`.
  - new import: `"strings"`.

## New Files

- `internal/ui/map.go` — `mapNode`, `mapGraph`, `buildMapGraph`, `layoutFR`,
  `clamp`, `mapWidget`, `mapRenderer`, `Tapped`, `SetCurrentScreen`
  (~220 lines total).
- `internal/ui/map_test.go` — `TestMapGraphCapturesAllNavigateEdges`,
  `TestMapGraphSkipsUnknownTargets`, `TestLayoutDeterministic`,
  `TestBreadcrumb`, `TestBreadcrumbEmptyStack`,
  `TestBreadcrumbUnknownIDFallsBackToRaw`, `makeLinearScreens` helper
  (~90 lines).

## Dependencies (task ordering)

```
T1 (graph)         ┐
T2 (layout FR)     ├── run together; T2 consumes T1's types
T11 (tests)        ┘   ← tests for T1+T2 land alongside

T3 (widget skel)   ┐
T4 (renderer)      ├── run together; renderer consumes widget
T5 (Tappable)      ┘

T6 (panel wiring)  ── needs T3+T4+T5; mutates rebuildTabs
T7 (struct fields) ── independent; can land first
T8 (breadcrumb)    ── independent
T9 (header bar)    ── needs T7+T8
T10 (real-time)    ── needs T6+T9
T12 (theme parity) ── needs T6
```

Critical path: T1+T2 → T11 → T3 → T4 → T5 → T6 → (T7 ∥ T8) → T9 → (T10 ∥ T12).

## Risks (top 3)

1. **Click-to-navigate on a custom widget.** *Medium.* Fyne's docs say
   canvas primitives are non-interactive, but widget-level `Tappable` on
   the `mapWidget` itself works (research brief §9, approach A). If it
   doesn't deliver events in practice, drop click-to-navigate per the
   spec's "if it's painful in Fyne, drop it but explain why" clause.
   Document the limitation in a `// ponytail:` comment; no per-dot
   button overlay as a workaround.

2. **Breadcrumb staleness.** *Low.* Mitigated by the Option A pattern in
   Task 10: `u.breadcrumbLabel` is a single `*widget.Label` created once
   in `headerBar`, refreshed inside `renderGrid`. `renderGrid` is the
   single sink for every `u.current` / `u.stack` change
   (`press`, `applySettings`, etc.). One update site = one source of
   truth.

3. **HSplit / VSplit nesting on small windows.** *Low–medium.* Nested
   splits have no min-size enforcement beyond each child's `MinSize()`.
   The map widget's `MinSize` is 260×300. On a 1024×600 screen, the
   preview/keypad split and the map split compete for vertical space.
   Mitigation: the map collapses via `SetOffset(1.0)`; the radiologist
   can hide it. Document the 1280×800 default window size (already set
   at `w.Resize(...)` in `buildMainUI`) as the recommended minimum for
   a usable map.

---

## Task 13 — Dev cycle (release 0.16.0)

**Run these commands in order. Don't run them out of order. Don't commit
before lint+vet+test pass.**

```bash
# 1. Lint + format + vet + test
gofmt -w .
go vet ./...
golangci-lint run ./...
go test ./...

# 2. Bump version: main.go:14  var Version = "0.15.3"  →  var Version = "0.16.0"

# 3. Commit
git add -A
git commit -m "feat: version bump 0.15.3 → 0.16.0 (screen-map feature: dot/line graph + breadcrumb header)"

# 4. Push
git push origin main

# 5. Build all 4 binaries to dist/
go build -tags flatpak -o dist/radkeys-linux-amd64 .
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc \
  go build -o dist/radkeys-windows-amd64.exe .
go build -tags flatpak -o dist/radkeys-config-linux-amd64 ./cmd/radkeys-config
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc \
  go build -o dist/radkeys-config-windows-amd64.exe ./cmd/radkeys-config

# 6. Lightweight tag (NOT -a, NOT -m)
git tag v0.16.0 HEAD

# 7. Push tag
git push origin v0.16.0

# 8. Watch CI to completion
gh run watch <run-id> --exit-status
#   <run-id> = the run created by the tag push (visible via `gh run list`)

# 9. Upload the 4 binaries to the release (Linux + Windows + both config binaries)
gh release upload v0.16.0 \
  dist/radkeys-linux-amd64 \
  dist/radkeys-windows-amd64.exe \
  dist/radkeys-config-linux-amd64 \
  dist/radkeys-config-windows-amd64.exe \
  --clobber
```

**Acceptance for this task:**
- `go test ./...` passes locally.
- `var Version` in `main.go:14` reads `"0.16.0"`.
- The 4 binaries exist in `dist/` with non-zero size.
- `git tag v0.16.0` is lightweight (`git cat-file -t v0.16.0` → `tag`,
  `git cat-file -p v0.16.0` shows a commit SHA, no tagger/date).
- `gh release view v0.16.0` lists all 4 binaries.
- AGENTS.md dev cycle is complete (Linux + Windows + config binaries
  shipped; CI release published).

---

## Estimated Diff Size

- `map.go`: ~220 lines new (incl. comments, layout, widget, renderer).
- `map_test.go`: ~90 lines new.
- `ui.go`: ~60 lines added (3 new fields, 1 constant, 4 new methods,
  panel-wrap in `rebuildTabs`, 2-line hook in `renderGrid`,
  1-line hook in `osThemeSettledListener`).
- `main.go`: 1 line (version bump in Task 13).
- **Total: ~370 LOC** including tests, no new imports outside
  `internal/ui/`, no new go.mod entries.
