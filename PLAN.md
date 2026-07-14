
# RadKeys — Current State, Architecture, Known Issues, and Correction Steps

> Living document for the current main branch. Historical releases are in git log.

---

## Current State

- Version: `0.12.1` (`var Version = "0.12.1"` in `main.go:23`).
- Branch: `main`. CI is green. Validation is static + mock + cross-compile; no hardware prototype exists yet.
- Shipped binaries per release: `radkeys-linux-amd64`, `radkeys-windows-amd64.exe`, `radkeys-config-linux-amd64`, `radkeys-config-windows-amd64.exe`. macOS is supported in source only; no binary is shipped because cross-compiling CGO from Linux is impossible.
- `0.x.x` only. Move to `1.0.0` only after Galvani approves the hardware prototype.
- All configuration lives in `radkeys.config.toml`. The device is flashed once and never reflashed or reconfigured from the host.
- This plan does not repeat release history, build commands, or the dev cycle; those remain in `git log` and `AGENTS.md`.
- It also does not redesign product behavior: the 6×6 grid, 12 actions, and composite-USB approach are locked in.
- Out of scope for this plan: new hardware platforms, new network protocols, cloud sync, telemetry, or non-HID input methods.
- macOS support remains source-only; shipping a macOS binary is still not planned.
- The `research/` briefs are kept for reference but are not part of active work.

---

## Current Architecture

### Device (RP2040-Zero)

The RP2040-Zero runs a composite USB firmware with two HID interfaces:

- **Vendor HID interface** (usage page `0xFF00`): sends 2-byte IN reports `[row, col]` on every physical button press.
- **HID keyboard interface**: injects keystrokes into the currently focused window when the host sends a 2-byte vendor OUT report `[cmd, arg]`.

The device stores no configuration. It receives only transient RAM commands. Supported commands are `FIRE_PASTE` (0x01), `GET_VERSION` (0x02), `SELECT_ALL` (0x03), `SELECT_LINE` (0x04), `LINE_START` (0x05), `LINE_END` (0x06), `BACKSPACE` (0x07), and `DELETE` (0x08). The modifier byte is `0x01` for Ctrl (Linux/Windows) and `0x02` for GUI/Cmd (macOS). `SELECT_LINE` uses a fixed Shift modifier.

The firmware replies to `GET_VERSION` with a 2-byte IN report `[major, minor]` once at connect. Current firmware reports `[1, 0]`. The protocol is documented in `firmware/rp2040-zero/PROTOCOL.md` and matches `firmware/rp2040-zero/diy.ino`.

### Host (Go + Fyne)

The host is a configurator, not a keystroke injector:

- `text` loads the phrase into the preview.
- `copy` copies the preview text to the host clipboard.
- `paste` sends `FIRE_PASTE` with the OS modifier to the device; the device types Ctrl/Cmd+V into the focused window.
- `select_all`, `select_line`, `line_start`, `line_end`, `backspace`, `delete` send the corresponding device-keyboard command.
- `navigate`, `prev`, and `home` switch screens using an internal stack; they never touch the OS focus or clipboard.

Modifier selection is `runtime.GOOS == "darwin"` ? GUI : Ctrl. The `internal/keystroke/` package was removed; the device is the keyboard.

### Focus Invariant

The HID event path (`pollHID → press(fromUI=false)`) must never raise, activate, or focus the RadKeys window. It is documented on `appUI.press` and enforced statically by `TestHIDPathDoesNotActivateWindow`, which parses `ui.go` and fails if any HID-path method calls `u.win.Show/ShowAndRun/SetContent/RequestFocus`.

### Surface Area

- 12 button actions: text, copy, paste, prev, home, navigate, select_all, select_line, line_start, line_end, backspace, delete.
- 13 themes, including system default and 12 named presets.
- 7 UI languages.
- Optional separate binary: `radkeys-config` (visual TOML editor in `cmd/radkeys-config`). It loads the config from its own directory, supports File→Open/Save/Save As, validates before saving, and never exposes raw TOML syntax to the user.

### Validation Model

Because the hardware prototype does not exist yet, all validation is:

1. **Static** — source review of `firmware/rp2040-zero/diy.ino` against `PROTOCOL.md`.
2. **Mock** — host code tested with `hid.NewMock()` and `fakeHIDDevice` in unit tests.
3. **Cross-compile** — Linux flatpak, Windows mingw, and `GOOS=darwin go vet` to confirm macOS compatibility.

Real end-to-end paste/focus behavior is only verified once Galvani flashes the RP2040-Zero and tests it against a live RIS/PACS. CI runs on Ubuntu only; Windows and macOS are not exercised in CI.

---

## Known Issues (all)

### Editor refresh and duplication

- `internal/editor/grid.go:127` — `updateButtonsTab()` over-fires: `refresh()` calls `refreshGrid()` + `refreshInspector()` + `refreshLayerBar()` + `refreshProblems()`, each of which calls `updateButtonsTab()` → `buildButtonsTab()` → full rebuild of `layerBar` + `inspector` + `problemsBox` + `gridBox`. A single `addButton()` or `removeButton()` triggers 4 full tab rebuilds. (B1)
- `internal/ui/ui.go:355` — `buildSettings()` is 191 lines with an 80-line save closure that mutates `tabs.Items[1].Content = u.buildSettings()` and `tabs.Items[2].Content = u.buildAbout()`, then conditionally replaces `tabs.Items[0]` with a new `VSplit`. Full teardown+rebuild on every save; 9.5× the AGENTS.md 20-line guideline. (B2)
- `internal/ui/ui.go:31` — `Run()` is 128 lines (app creation, theme resolution, i18n setup, window creation, UI construction, grid render, mock status, settings listener, firmware check, device open, pollHID launch, close handler) — all in one function, well over the 20-line guideline. (H1)
- `internal/ui/ui.go:624` and `internal/editor/inspector.go:190` — identical `labeled()` helper duplicated across packages. (H2)
- `internal/ui/ui.go:584` and `internal/editor/appsettings.go:124` — identical `section()` helper duplicated across packages. (H3)
- `internal/editor/inspector.go:95` — `actionOptions()` (hardcoded `[]string` of i18n labels), `actionLabel()` (hardcoded `map[string]string` action→label), and `actionFromLabel()` (iterates `configActionOrder()` calling `actionLabel()` to reverse-map) keep three parallel data structures for the same 12 actions; adding a 13th requires editing all three. (H4)
- `internal/ui/ui.go:473` — `tabs.Items[i].Content = X; tabs.Refresh()` is a fragile Fyne pattern (mutating `AppTabs.Items` after `SetContent`). (L5)

### Silent failures and fallback behavior

- `main.go:38` — `hid.Open` silently falls back to `hid.NewMock()` when the device is not found, violating the AGENTS.md "fail loud" rule. The user sees a translated status bar but the terminal log is English-only; the real error is swallowed. (H5)
- `internal/editor/editor.go:321` — `setVendorID`/`setProductID` return early without `setDirty()` on parse error; the user can type "xyz" and the old value persists with no error feedback and ambiguous dirty state. Contradicts PLAN.md step 9's "schema-driven, constrained inputs only" claim. (H10)
- `internal/ui/ui.go:443` — the settings save closure silently ignores invalid VID/PID hex input (e.g. `"0x12345"` overflows uint16 → ParseUint errors → old value kept, no user feedback). Inconsistent with how columns/rows handle invalid input (which falls back to 1 and updates the entry). (H11)

### Stale or dead PLAN.md content (this file)

- `PLAN.md:91` (the obsolete file replaced by this document) — referenced `var Version = "0.10.0"` while the real source had `0.12.1`. (H6)
- `PLAN.md:128` (the obsolete file) — the "Known bugs to fix" section listed the mock-mode log-line bug as queued, but the `"using mock (click UI buttons)"` fragment had already been removed from `main.go:44`. (H7)
- `PLAN.md:159` (the obsolete file) — the historical step-by-step plan read as actionable work and lacked a "DONE — DO NOT REDO" marker, risking re-implementation of completed work. (H8)
- `PLAN.md:86` (the obsolete file) — the "Current state" summary skipped Steps 8 (v0.11.0 editing commands) and 9 (v0.12.0 config editor) even though both shipped. (H9)
- `PLAN.md:128` (the obsolete file) — the "Known bugs" header said "(queued — not part of the current step)" although the bug it described was already fixed. (M7)
- `PLAN.md` overall (the obsolete file) — 486 lines of documentation bloat with duplicated Current state / step-by-step / History sections. (L7)
- `PLAN.md:86` (the obsolete file) — "(starting point)" header was confusing at v0.12.1 since the project was not starting. (L8)

### Editor implementation gaps

- `internal/editor/editor.go:217` — orphaned `moveButton` doc comment with no function body after it (dead documentation). (M1)
- `internal/i18n/i18n.go:425` — `editor.move_up` and `editor.move_down` keys exist in all 7 languages, but the layer-reorder UI they were meant for was never implemented. (M5)
- `internal/i18n/i18n.go:528` — dead keys `editor.help_toggle`, `editor.help_label`, `editor.model_intro`, `editor.preview_jump`, `editor.last_screen`, `editor.about_model`, `editor.confirm_remove_button`, `editor.no_button_selected`, and all `editor.help.*` (dot-separated) keys are defined in 7 languages but never referenced in the editor package. (M6)
- `internal/editor/io.go:136` — `saveConfigAs` receives a `fyne.URIWriteCloser` (`rc`) but never calls `rc.Close()`; the URI writer is opened and leaked. The actual write goes through `e.cfg.Save(path)` on the file path, so the URIWriteCloser is unused for I/O but still needs closing. (M8)
- `internal/editor/inspector.go:58` — `ent.OnChanged = e.setButtonLabel` fires on every character typed; each call triggers `setDirty()` + `refreshGrid()` + `refreshProblems()`. Typing "RX" runs the chain twice. (M9)
- `internal/editor/grid.go:103` — `outOfGridButton` sets `e.selected` directly and calls `refreshInspector()` + `refreshGrid()` manually instead of reusing `selectCell()`, duplicating the selection logic. (M10)

### Code structure and anti-patterns

- `internal/ui/ui.go:197` — `press()` dispatches 12 actions via a long switch with repeated `u.fireDeviceCommand(action, cmd, arg, fromUI)` calls; only `cmd` and `arg` vary. Acceptable for 6 device-keyboard cases but should become table-driven if more actions are added. (M2)
- `AGENTS.md:91` — the architectural statement "Screens are connected via `navigate` with `target`. Navigation is stack-based (`prev` goes back, `home` goes to root)." sits in the "🚫 Never" section, where prohibitions belong. (M4)
- `internal/hid/reader_cgo.go:168` — `readFirmwareVersion` writes `GET_VERSION` then reads 2 bytes. If a button is pressed during the 500ms version-read window, the `[row, col]` IN report could arrive before the version reply and be misinterpreted as `[major, minor]`. The protocol uses no report ID to distinguish them. Low risk today (version read happens once at connect before the event loop). (L1)
- `internal/config/config.go:289` — `Issue.Error()` dispatches through six nested switch functions (`appError` → `layoutError` → `screenError` → `buttonError` → `positionError` / `actionFieldError`) for what is essentially a `map[IssueKind]formatter`. (L2)
- `internal/theme/theme.go:218` — 13 theme preset globals plus a separate `Presets` slice; adding a 14th theme requires two edits. (L3)

### Documentation and tests

- `radkeys.config.toml` — header does not mention that `config.Save` strips comments (BurntSushi/toml limitation) and creates a `.bak` backup before rewriting, so users who hand-edit the file are surprised when their comments vanish. (L9)
- `internal/editor/editor_test.go:154` — `TestStartupPathUsesExecutableDir` expects the fallback `"radkeys.config.toml"`, but `StartupPath()` checks `RADKEYS_CONFIG` first, then the executable directory. If either is set, the test fails — environment-dependent. (L10)

### Verified correct (kept in the list for completeness, no change needed)

- `internal/ui/ui.go:305` — `pollHID` wraps every event in `fyne.Do()`; required because `press()` touches Fyne widgets, and correct Fyne usage. No change. (M3)
- `internal/i18n/i18n.go:574` — `init()` calls `bundle.AddMessages` ~700 times; acceptable for a desktop app (~1-2ms). No change. (L4)
- `internal/ui/ui.go:148` — `previewBg` is a `*canvas.Rectangle` stored as a field and mutated by the settings listener; low risk in the current single-creation flow. No change. (L6)

---

## Correction Steps

The order below is the execution order, and it is not optional. Every step in this list must be completed before the next release ships, in the order shown. The agent executing this plan must report completion of every step by number, with the resolved issue IDs and the output of the verification command. "Done" without per-step evidence is not acceptable. Conventional commits only; run `gofmt -w . && go vet ./... && go test ./...` before every non-docs commit. Severity is implicit in the position: the first steps fix user-visible failures and silent violations of the "fail loud" rule, the middle steps stabilize the codebase so later changes are safe, and the last step is firmware-only and must be validated by Galvani on the RP2040-Zero.

1. **docs: replace PLAN.md with current state + architecture + known issues**
   Resolves: H6, H7, H8, H9, M7, L7, L8.
   Replace the obsolete 486-line `PLAN.md` with this document.
   Verify: `wc -l PLAN.md` is under 280 and no stale version strings remain.

2. **fix: surface hid.Open and VID/PID parse failures instead of silent fallback**
   Resolves: H5, H10, H11.
   Show a dialog when the device is not found and mock mode is entered. Mark invalid VID/PID entry values with DangerImportance and flash an error; do not silently keep the old value.
   Verify: `go test ./...` passes; manual run with a bogus VID produces a visible error state.

3. **fix: decouple editor refreshes from updateButtonsTab**
   Resolves: B1, M9, M10.
   Make each `refreshXxx` update only its own cached widget; defer a single `updateButtonsTab()` per mutation cycle. Debounce label OnChanged or make it update on focus loss. Reuse `selectCell()` in `outOfGridButton`.
   Verify: add/remove/select buttons in `radkeys-config` perform exactly one tab rebuild; no visual flicker.

4. **refactor: split Run and buildSettings into focused methods**
   Resolves: H1, B2, L5, L6.
   Extract `buildMainUI()`, `checkFirmware()`, `startHIDLoop()`, and `applySettings()` methods. Rebuild tabs by creating a new `AppTabs` container instead of mutating `Items` after `SetContent`.
   Verify: `golangci-lint run ./...` clean; `Run` and `buildSettings` are each under 40 lines; mock-mode smoke test passes.

5. **refactor: extract shared Fyne helpers and action definition table**
   Resolves: H2, H3, H4.
   Move `labeled()` and `section()` to a shared `internal/widgetutil` package. Replace the three action mappings in the editor with one ordered `actionDefs` slice.
   Verify: `go test ./...` and both Linux/Windows builds pass; no duplicate `labeled`/`section` functions remain.

6. **fix: close URIWriteCloser in saveConfigAs and harden StartupPath test**
   Resolves: M8, L10.
   Add `defer rc.Close()` in `saveConfigAs`. Clear `RADKEYS_CONFIG` or run `StartupPath` from a temp directory in the test.
   Verify: `go test ./internal/editor/...` passes with `RADKEYS_CONFIG=/tmp/fake.toml` set externally.

7. **fix: remove orphaned moveButton comment and dead editor i18n keys**
   Resolves: M1, M5, M6.
   Delete the bodyless `moveButton` comment. Remove unused `editor.move_up`, `editor.move_down`, and all unused `editor.help.*` / `editor.preview_jump` / `editor.last_screen` / `editor.about_model` / `editor.confirm_remove_button` / `editor.no_button_selected` keys.
   Verify: `grep -rn 'editor\.move_up\|editor\.help_toggle\|editor\.model_intro' internal/i18n/` returns nothing; `go test ./internal/i18n/` still passes.

8. **refactor: table-driven press dispatch and Issue.Error formatter table**
   Resolves: M2, L2.
   Replace the 12-case action switch with a `map[string]deviceCommand` table. Replace nested `Issue.Error` switch functions with a `map[IssueKind]formatter` table.
   Verify: `go test ./internal/config/...` and `go test ./internal/ui/...` pass; new table is covered by existing tests.

9. **refactor: consolidate theme preset registry**
   Resolves: L3.
   Build `Presets` from a single slice literal or `init()` registry so adding a theme requires one edit.
   Verify: `go test ./internal/theme/...` passes; adding a fake preset requires editing only one location.

10. **docs: add .bak backup note to radkeys.config.toml header**
    Resolves: L9.
    Document that `config.Save` creates `radkeys.config.toml.bak` before rewriting.
    Verify: `grep -n ".bak" radkeys.config.toml` returns a header comment.

11. **docs: move navigation architecture statement out of AGENTS.md Never section**
    Resolves: M4.
    Relocate the `navigate`/`prev`/`home` sentence to the Project section or rephrase it as a rule.
    Verify: `grep -n "navigate" AGENTS.md` no longer appears under the "Never" heading.

12. **fix: disambiguate firmware version reply from button events**
    Resolves: L1.
    Firmware step: use a distinct report ID or sentinel for `GET_VERSION` replies, or host step: retry the version read when the result looks like a button event. **This step requires Galvani to flash + test on the RP2040-Zero — no I-tested-it-without-Galvani claims.**
    Verify: static review of `diy.ino` + `PROTOCOL.md` coherence; host mock tests still pass.

---

## Notes

- No release history is kept here; use `git log --oneline` and `git tag -l` for that.
- The dev cycle, build commands, and release checklist remain in `AGENTS.md`.
- Firmware changes always require Galvani to flash and validate the RP2040-Zero prototype; until then they are considered statically reviewed only.
