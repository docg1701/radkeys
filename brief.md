# RadKeys — Technical Brief

> **Date:** 2026-07-12
> **Repo:** https://github.com/docg1701/radkeys
> **Current release:** v0.2.1 → **target: v0.3.0 (architectural refactor)**
> **Status:** ✅ Complete — pending version bump and release

---

## Architecture (v0.3.0)

### 1. Firmware protocol: `(row, col)` — 2 bytes

Firmware (RP2040-Zero, TinyUSB) sends 2 bytes per event: `[row, col]`.
No fixed-size bitmap. Grid size is configurable in `radkeys.config.toml`
without recompiling firmware.

### 2. Hierarchical navigation: `navigate` + `target`

Screens have unique IDs. Buttons navigate via `target` to any screen.
Stack-based `prev` goes back. `home` returns to the first screen.

Actions: `text`, `copy`, `paste`, `prev`, `home`, `navigate`.

### 3. No fixed buttons

`copy`, `paste`, `prev`, `home` are normal actions placed anywhere on the grid.
No `[app.fixed_buttons]` reserving hardware indices.

---

## What changed from v0.2.1

| Area | Before | After |
|------|--------|-------|
| Config model | `Screen` with `index` + `FixedButtons` | `Screen` with `(row, col)` — no fixed indices |
| Navigation | Graph with `target` IDs + deck stack | Graph with `target` IDs + stack in ui.go |
| Firmware protocol | 24-byte bitmap | 2-byte `(row, col)` |
| HID events | `Event{Index int}` | `Event{Row, Col int}` |
| Theme system | `lighten`/`darken`/`blend` magic factors | 7 base colors per preset, 28 explicit derivations |
| Theme files | `custom.go` + `presets.go` | Single `theme.go` |
| Deck package | `internal/deck/` | Removed — navigation in ui.go |
| Version | `[app] version` in config.toml | Build constant (`var Version = "dev"`, set via `-ldflags`) |
| Config errors | `log.Fatal` to stderr | Fyne dialog with "open file" button |
| Linting | None | `golangci-lint` + `taplo` (TOML) |
| Code language | Mixed English/Portuguese | English only (rule in AGENTS.md) |

---

## Project structure

```
radkeys/
├── main.go / go.mod / go.sum
├── radkeys.config.toml        # documented example config
├── .golangci.yml              # Go linter config
├── .taplo.toml                # TOML linter config
├── dist/                      # gitignored
├── internal/
│   ├── config/                # TOML parser + validation + types
│   ├── hid/                   # HID reader (go-hid + mock)
│   ├── ui/                    # Fyne UI: preview + grid + settings + about
│   ├── i18n/                  # single Go map (7 languages)
│   ├── theme/                 # theme.go — 13 presets
│   └── assets/                # embedded icons
├── firmware/
│   └── rp2040-zero/           # diy.ino (TinyUSB, row,col protocol)
├── BUILD.md                   # hardware assembly guide (Portuguese)
└── research/
```

---

## Data model (config.go)

```go
type Config struct {
    App     App      `toml:"app"`
    Screens []Screen `toml:"screens"`
}

type Screen struct {
    ID      string   `toml:"id"`
    Name    string   `toml:"name"`
    Buttons []Button `toml:"buttons"`
}

type Button struct {
    Row     int    `toml:"row"`
    Col     int    `toml:"col"`
    Label   string `toml:"label"`
    Action  string `toml:"action"`  // text | copy | paste | prev | home | navigate
    Target  string `toml:"target,omitempty"`
    Content string `toml:"content,omitempty"`
}
```

---

## Theme system

- Single file: `internal/theme/theme.go`
- Each preset defines 7 base colors: bg, fg, primary, button, header, input, hover
- All remaining 28 Fyne `ThemeColorName` values are derived from these bases
- No `DefaultTheme` fallback for colors (only system preset delegates entirely)
- Dark-only presets stay dark regardless of OS theme (variant resolution falls back to available variant)

---

## Linting (mandatory)

```bash
golangci-lint run ./...   # Go: errcheck, gofmt, goimports, misspell, staticcheck, etc.
taplo lint *.toml          # TOML: syntax validation
```

Both must pass before every commit. See `.golangci.yml` and `.taplo.toml`.

---

## Platforms

| Platform | Binary in release | Responsibility |
|----------|-------------------|----------------|
| Linux | ✅ Built and delivered | Priority — tested |
| Windows | ✅ Cross-compiled with mingw | Provided, NOT tested by author |
| macOS | ❌ Not delivered | Build instructions only |

---

## Remaining work

| Item | Status |
|------|--------|
| Always-on-top | ⏳ Blocked by Fyne v2.8.0 |
| Version bump to 0.3.0 + release cycle | ⏳ Pending |
| UI test coverage | ⏳ None yet |

---

## Version rules

- Single source of truth: `var Version = "dev"` in `main.go`, set via `-ldflags "-X main.Version=X.Y.Z"` at build time.
- Never hardcode version in Go or config files.
- Test fixtures use `"0.0.0-test"`.
