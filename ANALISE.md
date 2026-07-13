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
> v0.10.0 published (Linux + Windows). No hardware prototype yet.

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
- **Step 6 — Final documentation (@ 8e55d53):** README/BUILD/AGENTS/ANALISE/
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
  and this `ANALISE.md` itself (mark steps 0-7 as done, reflect the reframe
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
- **Validation:** green CI + release with Linux+Windows + **Galvani confirms the
  full flow on hardware** before the tag.

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
