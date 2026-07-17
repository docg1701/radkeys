# ponytail-audit: RadKeys

**Date:** 2025-07-18
**Scope:** 28 Go files, 6346 lines
**Method:** repo-wide scan per ponytail-audit SKILL.md + subagent reviewer validation

---

## Findings (ranked biggest cut first)

### 1. yagni: `deviceBase` struct — fold into `diyDevice`

**File:** `internal/hid/reader_cgo.go:80-115`

`deviceBase` carries channels (`ch`, `stop`, `done`), `sync.Once`, emit/drop logic, and idempotent `Close`. Only `diyDevice` embeds it. `MockDevice` has its own independent copy of the same logic.

Extracted for a second protocol type that never arrived. The `Open()` function switches on `dev.Protocol` but only `ProtocolDIY` exists.

**What to cut:** Fold `deviceBase` fields and methods directly into `diyDevice`.
**Estimated cut:** ~40 lines.

### 2. shrink: duplicate `themeOptions`

**Files:** `internal/ui/ui.go:543`, `internal/editor/appsettings.go:62`

Both files define an identical method:

```go
func (u *appUI) themeOptions() (ids, names []string) {
    for _, p := range themes.Presets {
        ids = append(ids, p.ID())
        names = append(names, i18n.T("theme."+p.ID()))
    }
    return ids, names
}
```

`editor/appsettings.go` has the same logic on `*Editor`. 6-line duplication.

**What to cut:** Extract to `theme.PresetIDsAndNames()` in the theme package. Both callers already import `themes`. The i18n lookup can happen at the call site or via a translator parameter.
**Estimated cut:** ~6 lines dup.

### 3. shrink: `helpLine` wrapper — inline

**File:** `internal/editor/problems.go:26`

```go
func helpLine(text string) fyne.CanvasObject {
    lbl := widget.NewLabel(text)
    lbl.TextStyle = fyne.TextStyle{Italic: true}
    return lbl
}
```

Used exactly once, in `buildProblems`. 3-line wrapper for a 2-line pattern.

**What to cut:** Inline at the single call site.
**Estimated cut:** ~3 lines.

---

### 4. yagni: `ActionLabelKey` exported, single internal caller

**File:** `internal/config/config.go:85`

```go
func ActionLabelKey(id string) string { return "action." + id }
```

Zero external callers. Only `ActionLabel` (same file, line 88) uses it. Inline.
**Estimated cut:** -1 line, -1 export.

### 5. shrink: near-identical `setVendorIDFromEntry` / `setProductIDFromEntry`

**File:** `internal/editor/editor.go:312-336`

Two 12-line methods differing only by the field assigned (`VendorID` vs `ProductID`). Extract a shared helper.
**Estimated cut:** ~ -10 lines.

### 6. delete: `preset.name` English-name fallback in `FindPreset` is dead

**File:** `internal/theme/theme.go:287`

```go
if p.id == id || p.name == id {
```

All callers pass lowercase ids from config/UI dropdowns. The `p.name == id` branch never fires. The `name` field on every `preset` literal exists only for this dead path.
**Estimated cut:** -13 hardcoded strings, -1 condition.

### 7. shrink: `PresetIDs` single caller

**File:** `internal/theme/theme.go:295`

Exported, called only in `config.go:352` for an error format string. Could inline.
**Estimated cut:** ~ -5 lines.

---

## Reviewer validation

All 3 original findings **confirmed** by subagent `reviewer` (fork context) with source-code evidence. Findings 4-7 added by reviewer on missed-items pass.

---

## Not flagged (reviewed and cleared)

| Item | Reason |
|------|--------|
| `hid.hidDevice` interface (test seam) | Standard Go: one real impl (`*hid.Device`), one fake in tests. Minimal 3-method surface. |
| `CustomThemeMarker` interface | Unexported `radKeysTheme` struct can't be type-asserted across packages. Marker interface is the idiomatic Go workaround. |
| `issueFormatters` dispatch map | 18 formatters. Switch would be longer; map is cleaner and the table is the single source of truth. |
| `widgetutil` package (2 functions) | Shared between `ui` and `editor` — separate binaries. Correct Go internal package pattern. |
| `FirmwareOutdated` standalone | Used once but tested; 3-condition logic benefits from isolated test coverage. |
| `config.ActionList` + label helpers | 13 actions, 7 languages. The ceremony is proportional to the i18n surface. |
| `focus_invariant_test.go` AST guard | Safety-critical: prevents focus-stealing regression. Cost justified. |
| `labelDebounceTimer` package-level var | Already marked `ponytail:` — known acceptable shortcut. |
| `editor/` split into 6 files | One concern per file on an 800-line struct. Good organization, not over-engineering. |

---

## Summary

**net: ~ -80 lines, 0 deps removed.**  
7 findings in 6346 lines — all minor. The biggest cut (`deviceBase`, ~40 lines) is speculative abstraction for a protocol that doesn't exist. The rest is small deduplication and dead code.
