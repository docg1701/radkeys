
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

- `internal/editor/grid.go:127` — `updateButtonsTab()` over-fires, rebuilding the entire Buttons tab 2-4 times per mutation. See `plans/01-raw-findings.md#B1`.
- `internal/ui/ui.go:355` — `buildSettings()` is 191 lines with an 80-line save closure that rebuilds the whole settings tab on every save. See `plans/01-raw-findings.md#B2`.
- `internal/ui/ui.go:31` — `Run()` is 128 lines, well above the 20-line guideline. See `plans/01-raw-findings.md#H1`.
- `internal/ui/ui.go:624` and `internal/editor/inspector.go:190` — identical `labeled()` helpers duplicated across packages. See `plans/01-raw-findings.md#H2`.
- `internal/ui/ui.go:584` and `internal/editor/appsettings.go:124` — identical `section()` helpers duplicated across packages. See `plans/01-raw-findings.md#H3`.
- `internal/editor/inspector.go:95` — `actionOptions()`, `actionLabel()`, and `actionFromLabel()` keep three parallel data structures for the same 12 actions. See `plans/01-raw-findings.md#H4`.
- `internal/ui/ui.go:473` — `tabs.Items[i].Content = X; tabs.Refresh()` is a fragile Fyne pattern. See `plans/01-raw-findings.md#L5`.

### Silent failures and fallback behavior

- `main.go:38` — `hid.Open` silently falls back to `hid.NewMock()` when the device is not found, violating the "fail loud" rule. See `plans/01-raw-findings.md#H5`.
- `internal/editor/editor.go:321` — `setVendorID`/`setProductID` silently ignore invalid hex input. See `plans/01-raw-findings.md#H10`.
- `internal/ui/ui.go:443` — the settings save closure silently ignores invalid VID/PID hex input. See `plans/01-raw-findings.md#H11`.

### Stale or dead PLAN.md content (this file)

- `PLAN.md:91` — old reference to `var Version = "0.10.0"`; actual is `0.12.1`. See `plans/01-raw-findings.md#H6`.
- `PLAN.md:128` — "Known bugs to fix" section lists the mock-mode log-line bug as queued, but it is already fixed in `main.go`. See `plans/01-raw-findings.md#H7`.
- `PLAN.md:159` — the historical step-by-step plan reads as actionable work and lacks a "DONE — DO NOT REDO" marker. See `plans/01-raw-findings.md#H8`.
- `PLAN.md:86` — Steps 8 and 9 (v0.11.0 editing commands and v0.12.0 config editor) were missing from the "Current state" summary. See `plans/01-raw-findings.md#H9`.
- `PLAN.md:128` — the "Known bugs" header says "(queued)" although the bug is fixed. See `plans/01-raw-findings.md#M7`.
- `PLAN.md` overall — 486 lines of documentation bloat with duplicated Current state / step-by-step content. See `plans/01-raw-findings.md#L7`.
- `PLAN.md:86` — "(starting point)" header is confusing at v0.12.1. See `plans/01-raw-findings.md#L8`.

### Editor implementation gaps

- `internal/editor/editor.go:217` — orphaned `moveButton` comment with no function body. See `plans/01-raw-findings.md#M1`.
- `internal/i18n/i18n.go:425` — `editor.move_up` and `editor.move_down` keys exist but layer reordering is not implemented. See `plans/01-raw-findings.md#M5`.
- `internal/i18n/i18n.go:528` — dead keys `editor.help_toggle`, `editor.help_label`, `editor.model_intro`, `editor.preview_jump`, `editor.last_screen`, `editor.about_model`, `editor.confirm_remove_button`, `editor.no_button_selected`, and all `editor.help.*` keys are unused. See `plans/01-raw-findings.md#M6`.
- `internal/editor/io.go:136` — `saveConfigAs` receives a `fyne.URIWriteCloser` but never closes it. See `plans/01-raw-findings.md#M8`.
- `internal/editor/inspector.go:58` — `setButtonLabel` OnChanged fires on every keystroke, triggering heavy rebuilds. See `plans/01-raw-findings.md#M9`.
- `internal/editor/grid.go:103` — `outOfGridButton` duplicates selection logic instead of calling `selectCell()`. See `plans/01-raw-findings.md#M10`.

### Code structure and anti-patterns

- `internal/ui/ui.go:197` — `press()` dispatches 12 actions via a long switch with repeated `fireDeviceCommand()` calls. See `plans/01-raw-findings.md#M2`.
- `AGENTS.md:91` — the navigation architecture statement is misplaced in the "Never" section. See `plans/01-raw-findings.md#M4`.
- `internal/hid/reader_cgo.go:168` — `readFirmwareVersion` could read a button `[row, col]` event instead of the version reply. See `plans/01-raw-findings.md#L1`.
- `internal/config/config.go:289` — `Issue.Error()` uses six nested switch functions instead of a lookup table. See `plans/01-raw-findings.md#L2`.
- `internal/theme/theme.go:218` — 13 preset globals plus a separate `Presets` slice require two edits to add a theme. See `plans/01-raw-findings.md#L3`.

### Documentation and tests

- `radkeys.config.toml` — header does not mention that saves strip comments and create a `.bak` backup. See `plans/01-raw-findings.md#L9`.
- `internal/editor/editor_test.go:154` — `TestStartupPathUsesExecutableDir` is environment-dependent and can fail if `RADKEYS_CONFIG` is set or the test binary is in the project root. See `plans/01-raw-findings.md#L10`.

### Verified correct (kept in the list for completeness, no change needed)

- `internal/ui/ui.go:305` — `pollHID` wraps events in `fyne.Do()`; this is required because `press()` touches Fyne widgets and is correct Fyne usage. See `plans/01-raw-findings.md#M3`.
- `internal/i18n/i18n.go:574` — `init()` calls `bundle.AddMessages` ~700 times; acceptable for a desktop app. See `plans/01-raw-findings.md#L4`.
- `internal/ui/ui.go:148` — `previewBg` pointer is low-risk in the current single-creation flow. See `plans/01-raw-findings.md#L6`.

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
