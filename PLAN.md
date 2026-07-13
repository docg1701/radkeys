# RadKeys — Target Architecture + Step-by-Step Plan (post-compaction handoff)

> Final decision (Galvani, 2026-07-13): **paste without stealing focus becomes
> the FIRMWARE's responsibility**, not the host's. The RP2040-Zero becomes a
> **composite** USB device (vendor HID `[row,col]` + **HID keyboard** that sends
> Ctrl/Cmd+V on command). The App becomes a **configurator**: it stores ALL
> configuration (phrases, action for each of the 36 buttons) and **never writes
> to the device** after the single factory flash. **Startup-grab is accepted**
> (it draws the user's attention — it's a feature, not a bug). So **no
> non-activating window / layer-shell / Rust rewrite is needed** — Go+Fyne works
> on any OS, because during use there is no ping (HID in the background) and
> startup-grab is welcome.

---

## ⚠️ Constraints (mandatory, non-negotiable)

- **Always version `0.x.x`.** DO NOT move to `1.0.0` without explicit order from Galvani.
- **Hardware: only Galvani tests.** Any change in the firmware
  (`firmware/rp2040-zero/`) requires flashing + manual validation on the RP2040-Zero and
  does NOT count as verified until Galvani confirms. Never say "I tested it works" without Galvani having flashed.
- **Single executable per download.** The host (App) is one binary per OS
  (Linux flatpak + Windows mingw). No system shared-lib dependencies beyond the standard ones (GL/X11/Wayland). **GTK4/Qt are OUT** (they link system libs).
- **Device: single factory flash, never written again.** No reflash for
  configuration, no writing config to the device, no flash wear.
  All configuration lives in the App (TOML). The device receives only
  **transient** RAM commands (e.g., "fire paste"), it never persists anything.
- **Paste without stealing focus, on any macOS/Linux/Windows.** Via the device being
  a USB keyboard (native on any OS, no driver). Unicode via clipboard (host sets
  in text/copy), never via the device typing the phrase.
- **Code/comments/errors in English. Idiomatic Go. No hardcoded UI strings — use `i18n.T()`.** Version only in `var Version` in `main.go`.
- **Dev cycle (every release):** `gofmt -w . && go vet ./... &&
  golangci-lint run ./... && go test ./...` → bump `var Version` in `main.go` →
  commits (separate fixes + 1 `fix: version bump X → Y (context)`) →
  `git push origin main` → local build (Linux `-tags flatpak` + Windows mingw) →
  `git tag vX.Y.Z <sha>` (lightweight) → `git push origin vX.Y.Z` →
  `gh run watch <run> --exit-status` → `gh release upload vX.Y.Z
  dist/radkeys-linux-amd64 dist/radkeys-windows-amd64.exe --clobber`.
  Do not finish until the release has Linux+Windows binaries. **macOS: we do NOT ship a binary**
  (no Mac); the code compiles with GOOS=darwin (device-command is cross-platform).

---

## Target architecture (how it will actually work)

### Device (RP2040-Zero, composite firmware, single factory flash)
- **Vendor HID** interface (same as today): on any button click, sends
  `[row, col]` (IN report). The host reads in the background (without stealing focus).
- **HID keyboard** interface (NEW): does only one thing — when it receives a
  **vendor OUT** command "fire paste [modifier]", it sends the keystroke
  (Ctrl+V or Cmd+V: modifier down, V down, V up, modifier up) as a USB keyboard.
- **The device stores no config.** It does not know which button is paste — the host decides
  (reads `[row,col]`, sees it is paste, sends the command). Factory firmware, once,
  never touched again.

### Host (Go+Fyne app, single-binary, configurator)
- Reads vendor `[row,col]` (hid reader, same as today — background, no focus).
- **text:** host sets clipboard to phrase + shows preview (display). Nothing to
  the RIS, no focus. Unicode-safe (clipboard).
- **copy:** host sets clipboard to previewText (same as today).
- **paste:** host reads `[row,col]` → sees it is paste → **sends "fire
  Ctrl/Cmd+V" command to the device** (vendor OUT, 1 transient byte in RAM) → **device
  sends the keystroke as a keyboard** → the RIS (focused window) pastes the clipboard at the cursor. **Keyboard never steals focus** → guaranteed on any OS.
- **navigate (prev/home/navigate):** host switches screen (internal App state). No focus, no keystroke to the RIS.
- **Focus invariant:** the App **never** raises/focuses its own window when
  handling an HID event. text/copy/navigate are silent (update the background window + clipboard, no raise). Only paste sends a keystroke to the RIS (desired).
- **Modifier per OS:** `runtime.GOOS == "darwin"` → Cmd (GUI); otherwise Ctrl. The
  command carries the modifier. (macOS becomes supported in code; we do not ship a macOS binary.)

### Guarantees (honest, no lies)
- **Device sends Ctrl/Cmd+V as a USB keyboard:** ABSOLUTELY GUARANTEED on
  any macOS/Linux/Windows — USB keyboard is native, no driver, no software. (Bullet-proof.)
- **During use (1000 clicks), cursor does not ping in the RIS:** GUARANTEED on any
  OS — text/copy/navigate are background (HID read + background render +
  clipboard-set, nothing focuses the App); only paste sends to the RIS (desired).
- **Startup-grab:** ACCEPTED (feature). No non-activating window needed → **Fyne
  works on any OS, Wayland included** (no layer-shell, no rewrite). The user clicks on the RIS after launch and then the 1000 clicks flow without ping.
- **Unicode:** via clipboard (host), not via device. ✓
- **36 configurable keys:** everything in the App (phrases, action per button). The device is generic. ✓
- **Single-binary:** host gets SIMPLER (delete the `keystroke` package and the
  OS-specific injection). ✓

---

## Current state (starting point)

> Progress snapshot — steps 0-7 COMPLETE and committed. Release
> v0.10.0 published (Linux + Windows). Steps 8-9 (0.11.0 device editing
> functions, 0.12.0 config-editor app) planned. No hardware prototype yet.

- `var Version = "0.10.0"` in `main.go` (Step 7 bump).
- **Block 1+2 (antipattern cleanup, commits 5e1af11..8085d90):** kept —
  pure config + 6×6, status label, surfaced errors, CI -race, deterministic variantFor,
  testable hid lifecycle. Fyne remains in the new architecture.
- **Step 0 — Release 0.9.1 (@ b3b26e0):** bump 0.9.0→0.9.1, tag v0.9.1, release
  with Linux+Windows. ✅ COMPLETE.
- **Step 1 — Composite USB firmware (@ 4dcc90e):** `diy.ino` rewritten as
  composite TinyUSB: vendor IN `[row,col]` + OUT `[cmd,arg]` +
  HID keyboard interface (Ctrl/Cmd+V). `PROTOCOL.md` documented. ✅ COMPLETE
  (static validation — no hardware prototype yet).
- **Step 2 — Host device-command writer (@ 1cc1ac6):** `hid.Device`
  interface with `FirePaste(Modifier)` + `Version()`. `ModifierForOS()`
  (darwin→GUI, otherwise Ctrl). Testable mock. ✅ COMPLETE (mock + cross-compile).
- **Step 3 — Rewire paste + delete keystroke (@ 21770d9):** `ui.go` paste
  via `device.FirePaste(hid.ModifierForOS())`. The `internal/keystroke`
  package REMOVED (no xdotool/SendInput/osascript). ✅ COMPLETE.
- **Step 4 — Focus invariant (@ d26a2ed):** HID_FOCUS_INVARIANT documented
  in `press()` + static guard `TestHIDPathDoesNotActivateWindow`. ✅ COMPLETE.
- **Step 5 — Version check one-shot (@ c22537d):** `FirmwareOutdated` +
  `MinFirmware 1.0` + one-time warning dialog on connect. ✅ COMPLETE.
- **Step 6 — Final documentation (@ 8e55d53):** README/BUILD/AGENTS/PLAN/
  radkeys.config.toml updated for the composite-USB architecture;
  `PROTOCOL.md` verified against the firmware (no change). ✅ COMPLETE.
- **Step 7 — Release 0.10.0 (@ 916734f):** bump 0.9.1→0.10.0, tag v0.10.0,
  green CI, release published with Linux+Windows. ✅ COMPLETE.
- **No hardware prototype yet:** all validation so far is static
  (firmware review) + mock (host) + cross-compile (Linux flatpak, Windows
  mingw, `GOOS=darwin go vet`). The firmware will be flashed + tested on the
  RP2040-Zero only when the prototype is ready (weeks). Version `0.x.x` until
  everything is ready; `1.0.0` only after hardware approval.
- Toolchain: go 1.24 · golangci-lint v1.64.8 in `$(go env GOPATH)/bin` (add to
  PATH) · mingw OK · `gh` auth `docg1701` · DISPLAY=:0 (X11).

---

## Step-by-step plan (pi-subagents, small and actionable)

> Principle: each step = planner/worker/validator fresh-context (parent is
  single-thread writer orchestrator), focused validation, and **firmware steps
  need Galvani to flash + test on hardware**. Run
  `gofmt -w . && go vet ./... && golangci-lint run ./... && go test -race ./...`
  before committing. Follow the SKILL pi-subagents (staged fix orchestration:
  fanout read-only planning → 1 writer worker → fanout read-only validation
  → parent commits).

### Step 0 — Release 0.9.1 (cleanup already done)
- **What:** bump `0.9.0 → 0.9.1` in `main.go`, build Linux flatpak + Windows
  mingw, `git tag v0.9.1 <sha>` (lightweight), push, `gh run watch`, upload the
  binaries. The 6 antipattern commits (Block 1+2) are already in `main` and CI is green — only bump+tag+release is missing.
- **Subagent:** none (parent executes the dev cycle directly).
- **Validation:** green CI + release with Linux+Windows.
- **Why:** ship the cleanup before the firmware feature (clean changelog).

### Step 1 — Firmware: composite USB (vendor + keyboard) + fire-paste protocol
- **What:** rewrite `firmware/rp2040-zero/diy.ino` as composite TinyUSB:
  vendor IN (`[row,col]`, same as today) + vendor OUT (receives
  "fire paste [mod]" command) + **HID keyboard** interface (sends
  Ctrl/Cmd+V + release when commanded). Define the vendor OUT command protocol
  (e.g., byte 0 = cmd `0x01` fire-paste, byte 1 = modifier `0x01`=Ctrl / `0x02`=GUI/Cmd). Document the protocol in `PROTOCOL.md`.
- **Subagent:** `planner`/`reviewer` (research TinyUSB composite + HID
  keyboard on RP2040-Zero, read the current firmware and the TinyUSB/Adafruit lib) → `worker` (write the composite firmware + PROTOCOL.md). Parent validates statically (coherent HID descriptors).
- **Validation:** static firmware review (coherent HID descriptors,
  keyboard logic, OUT command handling) + cross-check of `PROTOCOL.md`.
  **No hardware prototype yet** — real flashing + testing on the RP2040-Zero is
  for when the prototype is ready (weeks); until then nothing is "I tested it on hardware". Keep `0.x.x` until hardware approval, then `1.0.0`.
- **Honest risk:** composite TinyUSB firmware is real work (HID descriptors for 2 interfaces + keyboard logic + OUT command reception). It is bounded and standard, but not trivial. Once done, never touched again.

### Step 2 — Host: device-command writer (vendor OUT fire-paste) + mock
- **What:** new code in the `hid` package (or new `device` package): function
  `FirePaste(mod Modifier)` that writes the 2-byte vendor OUT report to the
  device. Define `Modifier` (Ctrl/Cmd) via `runtime.GOOS`. Create an internal
  interface so the write can be mocked (testable without USB).
- **Subagent:** `worker` (implements the writer + mockable interface + unit
  tests of the writer with the mock).
- **Validation:** `go test -race ./internal/hid/` (mock), Linux/Windows build,
  `GOOS=darwin go vet ./internal/hid/` (compiles on mac). Real write on the
  device is tested by Galvani with the Step 1 firmware.
- **No hardware:** host code is tested with mock; real integration is Galvani.

### Step 3 — Host: rewire paste + delete keystroke injection
- **What:** in `internal/ui/ui.go`, `case config.ActionPaste`: replace
  `keystroke.SendCtrlV()` with `device.FirePaste(hid.ModifierForOS())`. **Delete the entire
  `internal/keystroke` package** (SendCtrlV + keystroke_darwin/linux/windows.go)
  — no more OS injection; the device is the keyboard. Confirm that text/copy
  still set the clipboard (host-side, Unicode-safe) and navigate still switches
  screen. Update i18n/tests that referenced keystroke.
- **Subagent:** `reviewer` (audit what depends on `keystroke`) → `worker`
  (rewire + delete + adjust imports/tests).
- **Validation:** `go build -tags flatpak`, `go test -race ./...`,
  `GOOS=windows ... go build`, `GOOS=darwin go vet ./...` (confirms it compiles
  without keystroke). App runs (DISPLAY=:0) with mock. **No hardware prototype
  yet** — real paste test (device sends Ctrl+V → RIS pastes) is for when the
  prototype is ready. Keep `0.x.x` until hardware approval.
- **Note:** macOS becomes supported in code (no per-OS keystroke;
  device-command is cross-platform). We do not ship a macOS binary.

### Step 4 — Host: HID no-focus-steal invariant
- **What:** ensure `press()`/`pollHID()` **never** raise/focus the RadKeys window
  when handling a `[row,col]`. Audit `ui.go` (no `RequestFocus`,
  `Show`, raise, `SetContent` re-trigger that focuses the window on the HID path).
  Add a documented invariant (comment +, if possible, a test/guard).
- **Subagent:** `reviewer` (audit HID paths for raise/focus) → `worker`
  (fix if needed + document the invariant).
- **Validation:** code review (documented invariant + static guard
  `TestHIDPathDoesNotActivateWindow`) + `go test -race ./...` with mock. **No
  hardware prototype yet** — visual confirmation "1000 clicks, cursor in RIS
  without ping" (text/copy/navigate silent; only paste sends to RIS) is for when
  the prototype is ready. Test on Linux Xorg, Linux Wayland, Windows (macOS if a
  Mac is available). Keep `0.x.x` until hardware approval.
- **No firmware:** host only. Visual "no ping" confirmation is for when the
  prototype is ready.

### Step 5 — (Optional) Firmware version check one-shot
- **What:** when connecting the device, the App reads the firmware version
  **once** and warns if it is old ("update the firmware once"). Does not nag during use.
- **Subagent:** `worker` (read device version via vendor + dialog/warning).
- **Validation:** Galvani tests (old firmware warns, new one stays silent).
- **Can be skipped** if Galvani prefers no check.

### Step 6 — Final documentation (everything updated)
- **What:** at the end of development (before release 0.10.0), update ALL
  project documentation to reflect the new architecture and real state. At a
  minimum: `README.md` (architecture paste-via-firmware-USB-keyboard,
  app=configurator, single-binary, startup-grab accepted, macOS supported in code
  without shipped binary, `keystroke` package removed, one-shot firmware version
  check), `BUILD.md` (hardware assembly + note about composite USB
  vendor+keyboard device + single factory flash), `PROTOCOL.md`
  (referenced by Step 1 — confirm it matches the final firmware),
  `radkeys.config.toml` (versioned example coherent with current fields/uses),
  and this `PLAN.md` itself (mark steps 0-7 as done, reflect the reframe
  "no hardware prototype yet → only static/mock validation; real flash when the
  prototype is ready; `0.x.x` until hardware approval, `1.0.0` only after", and
  rewrite all stale GATE/flash language in the validations of steps 1/3/4). Check
  any other `.md` in the repo and update if stale. No doc may contradict the
  shipped code.
- **Subagent:** `reviewer` (audit ALL docs against the final code — find stale:
  old architecture, hardware GATEs written as if hardware existed, outdated
  config fields, mentions of the `keystroke` package, etc.) → `worker` (rewrite
  each stale doc; identifiers in English, i18n where applicable).
- **Validation:** doc code review (no contradiction with shipped code),
  `go test ./...`/builds remain green, parent checks the doc diff before
  committing. No hardware: nothing depends on flashing.
- **Why:** release 0.10.0 ships with correct docs; handoff for when the
  prototype is ready (weeks) is clean.

### Step 7 — Release 0.10.0 (firmware feature)
- **What:** bump `0.9.1 → 0.10.0`, build, tag `v0.10.0`, push, CI, upload
  Linux+Windows binaries. Release notes: "paste now via firmware
  (USB keyboard); does not steal focus; macOS supported in code; keystroke
  package removed."
- **Subagent:** none (parent executes the dev cycle).
- **Validation:** green CI + release with Linux+Windows. Real hardware
  confirmation deferred (no prototype yet); `1.0.0` only after hardware approval.

### Step 8 — Release 0.11.0: device keyboard editing functions

- **What:** add six new configurable button actions that make the device
  keyboard send editing keystrokes (extending the composite keyboard interface
  from Step 1 — no new USB interface):
  - `select_all` — Ctrl/Cmd+A (select all text). Modifier per OS.
  - `select_line` — select the current line: Home, then Shift+End (two-key
    sequence; Shift is fixed, not OS-dependent).
  - `line_start` — Home (jump to start of line).
  - `line_end` — End (jump to end of line).
  - `backspace` — Backspace (delete backward).
  - `delete` — Delete Forward (delete forward).
  Firmware: new vendor OUT commands `0x03`..`0x08` (SELECT_ALL, SELECT_LINE,
  LINE_START, LINE_END, BACKSPACE, DELETE); each arms a volatile flag (like
  `pending_paste`) and `loop()` sends the keyboard sequence (HID keycodes:
  A=0x04, Home=0x4A, End=0x4D, Backspace=0x2A, Delete Forward=0x4C; Shift=
  KEYBOARD_MODIFIER_LEFTSHIFT 0x02; Ctrl/Cmd per OS). The `arg` byte carries the
  OS modifier selector for SELECT_ALL (0x01 Ctrl / 0x02 GUI, like paste);
  unused (0x00) for the others.
  Host: new `config` actions (ActionSelectAll/SelectLine/LineStart/LineEnd/
  Backspace/Delete) + validActions; `ui.go` press() dispatches them to a device
  command writer (generalize `FirePaste` into `FireCommand(cmd, arg)` — keep
  paste working); `ModifierForOS()` for select_all, `0x00` for the rest.
  `MockDevice` records the commands; unit tests assert the bytes. i18n button
  labels in all 7 languages. `PROTOCOL.md` documents the new commands +
  sequences.
- **Subagent:** `planner`/`reviewer` (firmware keycodes/sequences + host action
  wiring) → `worker` (firmware + host + i18n + PROTOCOL.md + tests) → fresh
  validators (firmware static + host mock + PROTOCOL coherence).
- **Validation:** `go test -race ./...`, build Linux/Windows, `GOOS=darwin go
  vet`, static firmware review (no compile/hardware). Real hardware test
  deferred (no prototype yet); `1.0.0` only after hardware approval.
- **Why:** radiologists need quick text editing (select all/line, jump
  start/end, delete) from the keypad without touching the keyboard or losing
  RIS focus — same no-focus-steal property as paste (device keyboard sends to
  the focused window).

### Step 9 — Release 0.12.0: dedicated RadKeys TOML config editor (separate, optional binary)

- **What:** build a complete, dedicated Go+Fyne app that makes editing
  `radkeys.config.toml` extremely easy for a lay user — a visual editor that
  knows the RadKeys schema (not a generic TOML editor). It is a SEPARATE,
  OPTIONAL binary (`cmd/radkeys-config/`, built for Linux + Windows like the
  RadKeys binary); RadKeys runs without it, and the TOML can still be hand-edited.

  **Launch / file handling:** on startup it auto-loads `radkeys.config.toml`
  from its own directory; File→Open loads any `.toml`; one-click Save (toolbar)
  writes back to the open file (Save As supported); recent files via
  `fyne.App.Preferences`. Dirty-state asterisk in the title +
  confirm-on-unsaved-quit.

  **UX (absolutely intuitive + graphical; the user never touches TOML syntax):**
  - Main view = the visual **6×6 grid** of the CURRENT layer, mirroring the
    physical keypad. Empty cells show "+"; filled cells show the label + an
    action icon. The user clicks a cell and the **options appear at the TOP of
    the window** (a top property bar, not a side panel): Label, an Action
    dropdown, and the per-action fields — only the fields valid for that action.
  - **Layers (= config `screens`):** the user adds / removes / renames layers
    and connects them with `navigate` buttons (a navigate button's Target is
    another layer). The top bar shows the current layer name + a Back (prev)
    button + a layer dropdown to jump to any layer + Add/Remove layer. (Config
    field stays `screens` for back-compat; the UI labels say "Layer".)
  - **Navigate like the device:** a Simulate/Preview toggle makes button clicks
    act like the real RadKeys — a `navigate` button switches to its target
    layer's grid, a `text` button shows the phrase in a preview pane, etc. — so
    the user walks the whole layer graph interactively before saving. In the
    default Edit mode, a click shows the button's options at the top.

  **The app resolves syntax; the user only makes valid choices (the core ask):**
  - The user NEVER sees or writes TOML. The editor only offers valid options:
    dropdowns for enums (action, language, theme preset, protocol, navigate
    target — populated from the actual layers), bounded inputs (row/col come
    from the grid cell, not free text), required fields enforced (label always;
    content for `text`; target for `navigate`), and inline validation that
    blocks invalid values before Save. The app generates the TOML; the output
    is guaranteed to be a valid `radkeys.config.toml`.
  - Static help: every field has a tooltip/help text (i18n, 7 languages) in
    plain language; a Help toggle reveals all explanations inline; an info
    panel explains the RadKeys model (device → layers → buttons → actions).
  - Dynamic help: the top options bar adapts to the chosen action (see above);
    invalid buttons are highlighted on the grid (duplicate position, bad
    navigate target) with a tooltip stating the problem in plain language.

  **Grid format + non-destructive resize (configuration hierarchy):**
  - The grid format (columns × rows → button count) is intuitive to adjust:
    a bounded stepper/slider (1–6 each) in the App-settings view.
  - **Configuration hierarchy:** the user sets the grid format (and the other
    app settings — device, theme, language, radiologist) FIRST; button-grid
    editing is the next step. On first run (empty/default config) a quick setup
    sets the grid format before any button editing. (Ordered sections / a
    first-run wizard — not a rigid lock; keep it simple.)
  - **Non-destructive resize (least destructive, preserve maximum data):** when
    the grid shrinks, buttons that no longer fit are NOT deleted — they move to
    a **parked/overflow pool** for that screen (preserved in the TOML as
    `[[screens.parked_buttons]]`, inactive; the RadKeys app ignores them and
    only loads `buttons`). The editor lists parked buttons with a way to place
    them back when the grid grows (auto-fill empty cells, or click-to-place).
    Growing the grid never loses data; shrinking never deletes a configured
    button. This needs a small `internal/config` addition
    (`Screen.ParkedButtons []Button`) that the app ignores on load.

  **KISS / no over-engineering:**
  - No drag-drop (Fyne lacks native list reordering — research gap
    /tmp/radkeys-012/research-fyne-forms.md). Reorder layers with up/down
    buttons; assign a button by clicking its grid cell (the grid IS the
    row/col picker — no row/col fields).
  - Schema-driven, constrained inputs only — the user can never produce an
    invalid TOML (see above); the app writes it.
  - Explicit Save (not auto-save) — config edits have consequences.
  - Explicit Save (not auto-save) — config edits have consequences.
  - Save writes a canonical, fully-commented TOML (the editor owns the format;
    comments are part of its output) and backs up the previous file to `.bak`.
    Target users edit via the editor, so canonical comments are sufficient; the
    `.bak` protects hand-editors. (A comment-preserving lib such as
    pelletier/go-toml/v2 is allowed ONLY if it stays simple; do NOT build an AST
    editor.)

  **Reuses:** `internal/config` (Load/Validate/Save + the `.bak`), `internal/i18n`
  (existing button.*/settings.*/status.* + new `editor.*` keys in 7 languages —
  see scout /tmp/radkeys-012/scout-config-schema.md), `internal/theme` (13
  presets). New `cmd/radkeys-config/main.go` + a new `internal/editor` package for
  the Fyne editor.

- **Subagent:** `planner`/`reviewer` (UX spec → Fyne widget mapping: `widget.Form`
  + Entry/Select/Check, `widget.List` with manual callbacks for screens/buttons
  [NOT `BindStruct` — issue #2607], `AppTabs`/`HSplit`, validation + `HintText`,
  `dialog.NewFileOpen`/`NewFileSave`, `Preferences`) → `worker` (build
  `cmd/radkeys-config` + `internal/editor`: auto-load/open/save/save-as, the
  3-panel UI, property inspector, static+dynamic help, inline validation +
  grid highlighting, `editor.*` i18n keys, reuse `config.Validate`) → fresh
  validators (UX review against the research patterns, Fyne-pattern review,
  schema-coverage review that every config field + validation rule is editable +
  guarded, and a no-over-engineering check).

- **Validation:** `go test -race ./...` (editor unit tests: load/save round-trip,
  validation surfacing, mock Fyne where feasible), build `radkeys-config` for
  Linux flatpak + Windows mingw (same toolchain as RadKeys), `GOOS=darwin go
  vet`. The editor is OPTIONAL — RadKeys still builds/runs without it. Ship
  `radkeys-config-linux-amd64` + `radkeys-config-windows-amd64.exe` in the
  0.12.0 release (alongside the RadKeys binaries + config template). Real
  usability test with radiologists deferred (no prototype pressure).

- **Why:** today the only way to add phrases/screens is hand-editing TOML (the
  Settings tab edits app settings only — scout-confirmed gap). A dedicated
  visual editor with game-inventory clarity + static/dynamic help makes the
  36-button configurability accessible to any radiologist, with zero TOML
  knowledge — this is the make-or-break UX for adoption. A separate, optional
  binary keeps RadKeys itself single-purpose.

---

## Suggested execution order (with pi-subagents)

1. **Step 0** (parent directly): release 0.9.1 of the already-done cleanup.
2. **Step 1** (planner→worker, **static validation**): composite firmware.
   → **No hardware prototype yet:** validation is static (coherent HID
   descriptors, keyboard logic, OUT command handling) + cross-check of
   `PROTOCOL.md`. Real flashing on the RP2040-Zero is for when the prototype is
   ready (weeks); until then nothing is "I tested it on hardware".
3. **Step 2** (worker, mock): device-command writer (parallel to 1? yes, no
   conflict — 2 only touches hid, 1 only firmware; but Step 1's protocol defines
   the bytes Step 2 writes → do 1 before 2, or 1+2 together with the planner
   defining the protocol first).
4. **Step 3** (reviewer→worker): rewire paste + delete keystroke. Depends on
   2 (the writer).
5. **Step 4** (reviewer→worker): focus invariant. Depends on 3 (paste
   rewired) to test the real flow (mock).
6. **Step 5** (worker): version check one-shot (included by Galvani's decision).
7. **Step 6** (reviewer→worker): final documentation — ALL docs updated
   against the shipped code (also rewrite stale GATE/flash language in steps
   1/3/4 validations). Depends on 1-5 being done.
8. **Step 7** (parent directly): release 0.10.0.
9. **Step 8** (planner→worker→validators): device keyboard editing functions
   (select_all/select_line/line_start/line_end/backspace/delete) + release
   0.11.0. Depends on Step 1 (composite keyboard) + Step 2 (device-command
   writer), which it extends.
10. **Step 9** (planner→worker→validators): dedicated RadKeys TOML config editor
    (`cmd/radkeys-config`, separate optional binary) — game-inventory 3-panel
    UI + static/dynamic help + inline validation. Depends on Step 8 (0.11.0
    actions in the action set) for full action coverage; can start in parallel
    once the action set is agreed. Release 0.12.0.

**Critical dependency:** no hardware prototype yet, so all validation is
static + mock + cross-compile (Linux flatpak, Windows mingw, `GOOS=darwin
go vet`). Incremental `0.x.x` until everything is ready; `1.0.0` only after
prototype hardware approval. Host steps (2-5) can be coded + tested with
mock in parallel with firmware (1), but Step 1's protocol defines the bytes
Step 2 writes.

---

## How to resume (operational notes)

- **golangci-lint** (not in the default PATH): `export PATH="$(go env GOPATH)/bin:$PATH"`.
- **Run the App** (DISPLAY=:0): `RADKEYS_CONFIG=/tmp/c.toml go run -tags flatpak .`
  — useful for reproducing UI bugs. Diagnosis: temporary `log.Printf` +
  `timeout 6s`.
- **Cross-OS vet/build without the OS:** `GOOS=windows GOARCH=amd64 CGO_ENABLED=1
  CC=/usr/bin/x86_64-w64-mingw32-gcc go build -o dist/radkeys-windows-amd64.exe .`;
  `GOOS=darwin go vet ./...` (macOS compiles, does not link without a Mac).
- **Release build:** `go build -tags flatpak -o dist/radkeys-linux-amd64 .` +
  Windows mingw.
- **CI:** `gh run list` → find the tag run → `gh run watch <id> --exit-status`.
- **pi-subagents:** follow the SKILL pi-subagents — staged fix orchestration
  (fanout read-only fresh planning → 1 writer worker → fanout read-only fresh
  validation → parent synthesizes + commits). Validators are **static-only**
  (do not run build/test — parent already validates) to avoid cold-cache
  slowness. Risky changes: `go test -race` + run the App. Worker does NOT touch
  `var Version` nor commits (parent commits in logical units). Firmware steps
  are not "I tested it" until Galvani flashes.
- **Firmware → Galvani:** any PR on `firmware/rp2040-zero/` is static code
  until Galvani flashes + tests on the RP2040-Zero.

---

## History (reference, do not redo)

- **Antipattern cleanup (Block 1+2, commits 5e1af11..8085d90):** hunt for the
  hacks from the 6 releases (0.4.0→0.9.0) — pure config + 6×6, status label,
  surfaced errors, CI -race, deterministic variantFor, isLight race,
  shift guard, emit log, testable hid lifecycle. All in `main`, CI green.
- **Architecture decision (this section):** paste via firmware-keyboard (not
  host-injection), app=configurator, single-binary, startup-grab accepted. Rejected
  GTK4/Qt (single-binary), Rust rewrite (startup-grab accepted made it
  unnecessary), reflash per config (stupid), writing config to device.
