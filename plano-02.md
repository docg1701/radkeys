# Implementation Plan — Screen-Map Panel + Breadcrumb Header (v0.16.0)

## Goal

Add a **side panel** on the Shortcuts tab that renders the navigation graph
as dots-and-lines only (no labels, no text), and a **breadcrumb header**
above the AppTabs that shows the `>`-separated path of screen names from
`u.stack + [u.current]`. Both are purely visual — no click-to-navigate,
no buttons in the map content. Zero new dependencies.

## Spec

- **Map panel (right side, inside the Shortcuts tab):**
  collapsible via `◀` chevron toggle (sets HSplit offset to 1.0 / 0.75),
  resizable via HSplit divider drag, vertical scroll (auto-hide via
  `container.NewVScroll`). Content = one dot per screen + one line per
  `ActionNavigate` edge. No labels, no text, no edge arrows. Current
  screen = color change on the dot only (primary vs foreground).
  **NOT clickable** — dots/lines are `canvas.Circle`/`canvas.Line`
  primitives, no `Tappable`, no `onNavigate`.
- **Breadcrumb header (above AppTabs):**
  `[breadcrumb]` (center) and `[device-status]` (right) via
  `container.NewBorder`. No `|` separator — the Border layout keeps
  them naturally apart. Breadcrumb = `u.stack + [u.current]`, each id
  replaced by `Screen.Name`, joined by ` > `. Updates every `renderGrid`.
- **Status moves into the header** — replaces the previous top-border slot.

## Layout algorithm: BFS layered (vertical cascade)

- BFS from `cfg.Screens[0]` (home/root screen) assigns each node a depth.
- Nodes at the same depth form a horizontal row; rows stack vertically.
- Isolated nodes (not reachable from root) go to a final row below.
- Stable horizontal order via sorted ids (deterministic, testable).
- Zero deps, ~50 lines.

Fruchterman-Reingold was considered and rejected — it produces a scattered
"hairball" distribution that looks chaotic for small DAGs. The layered layout
gives a clean vertical cascade with home at the top, which matches the mental
model of navigation.

## Out of scope (YAGNI)

- No click-to-navigate on map dots. The map is purely a visual aid.
- No labels or text in the map — dots and lines only.
- No persistence of collapse state across launches.
- No zoom/pan inside the scrollable map.
- No new i18n keys.
- No new go.mod entries.

---

## Files changed

### New: `internal/ui/map.go` (~220 lines)

- `mapNode`, `mapGraph` types
- `buildMapGraph(cfg)` — walks screens + `ActionNavigate` buttons
- `layoutLayered(g, w, h)` — BFS-based vertical cascade
- `clamp` helper
- `mapWidget` — custom Fyne widget, embeds `widget.BaseWidget`
- `mapRenderer` — `canvas.Circle` dots + `canvas.Line` edges, in-place
  color swap via `SetCurrentScreen`
- No `Tapped`, no `onNavigate` — the map is purely visual

### New: `internal/ui/map_test.go` (~110 lines)

- `TestMapGraphCapturesAllNavigateEdges`
- `TestMapGraphSkipsUnknownTargets`
- `TestLayoutDeterministic`
- `TestLayoutLayeredRootAtTop`
- `TestBreadcrumb`, `TestBreadcrumbEmptyStack`, `TestBreadcrumbUnknownIDFallsBackToRaw`
- `makeLinearScreens` helper

### Modified: `internal/ui/ui.go` (~70 lines added)

- **Struct fields:** `navMap *mapWidget`, `mapSplit *container.Split`,
  `mapVisible bool`, `breadcrumbLabel *widget.Label`
- **Constants:** `mapOffsetCollapsed = 1.0`, `mapOffsetExpanded = 0.75`
- **`rebuildTabs`:** wraps shortcuts in `u.shortcutsTab(main)`, uses
  `u.headerBar()` for the top border
- **`shortcutsTab(main)`:** creates HSplit with map panel on the right,
  collapse `◀` toggle button above the scrollable map
- **`headerBar()`:** `NewBorder(nil, nil, nil, u.status, breadcrumbLabel)`
  — breadcrumb center, status right, no separator
- **`breadcrumb()`:** `strings.Join(names, " > ")` from `u.stack + [u.current]`
- **`renderGrid`:** prepends `navMap.SetCurrentScreen(u.current)` +
  `breadcrumbLabel.SetText(u.breadcrumb())`
- **`osThemeSettledListener` + `applySettings`:** both call
  `navMap.SetTheme(th, v)` before `renderGrid`
- **Import:** `"strings"` added

### Modified: `main.go`

- `var Version = "0.16.0"` (bumped from 0.15.4)

---

## Dev cycle (completed)

```bash
gofmt -w . && go vet ./... && golangci-lint run ./... && go test ./...
# Bump version: main.go var Version = "0.16.0"
git commit -m "feat: version bump 0.15.4 → 0.16.0 (screen-map panel + breadcrumb header)"
git push origin main
# Build 4 binaries to dist/
go build -tags flatpak -o dist/radkeys-linux-amd64 .
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=... go build -o dist/radkeys-windows-amd64.exe .
go build -tags flatpak -o dist/radkeys-config-linux-amd64 ./cmd/radkeys-config
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=... go build -o dist/radkeys-config-windows-amd64.exe ./cmd/radkeys-config
git tag v0.16.0 HEAD
git push origin v0.16.0
gh run watch <run-id> --exit-status
gh release upload v0.16.0 dist/* --clobber
```

## Estimated diff size

- `map.go`: ~220 lines
- `map_test.go`: ~110 lines
- `ui.go`: ~70 lines added
- `main.go`: 1 line
- **Total: ~400 LOC**
