# RadKeys — PLAN.md

> Living plan for the next development cycle.
> Steps 1–12 executed (commits 575225a..c612e2d). Release v0.13.10 shipped.
> What's done lives in `git log`, not here.

## Current state

- Version: `0.13.10` (`var Version = "0.13.10"` in `main.go:23`).
- Branch `main`. CI green. No hardware prototype yet.
- 13 actions (12 shipped + `exec` pending), 13 themes, 7 languages, 4 binaries per release.
- `0.x.x` only. `1.0.0` only after Galvani approves the hardware prototype.
- Dev cycle and build commands in `AGENTS.md`. Do not repeat here.

## Pending (block the next release)

| ID | Description | Gate |
|----|-------------|------|
| L1 | Firmware: `GET_VERSION` reply uses the same report ID (0) as button events. Host mitigates with retry, but protocol ambiguity remains. The real fix is a distinct report ID or sentinel byte in the firmware IN report — requires flash + test on RP2040-Zero. | **1.0.0** |
| F1 | New `exec` action — runs an arbitrary bash command with user permissions. `content` = command. Host-side, fire-and-forget via `bash -c`. | Next release |

### L1 — firmware version vs button event ambiguity

**What's done:** host retries `GET_VERSION` 3 times on connect (`reader_cgo.go`). Mitigates a single stray button press but does not eliminate the root ambiguity: a `[row=1, col=0]` event is indistinguishable from `[major=1, minor=0]` on the wire.

**What's missing:** firmware-side fix (report ID or sentinel) documented in `firmware/rp2040-zero/PROTOCOL_FUTURE.md`. Implement when Galvani has hardware to validate.

**Gate for 1.0.0:** firmware-side fix validated on RP2040-Zero.

### F1 — `exec` action

New button action that runs `bash -c "<command>"`. The command lives in `content`. Same permissions as the RadKeys process (user, not root). No shell injection risk: `exec.Command` does not interpolate strings — `-c` receives the argument as a parameter, not via concatenation.

**Not yet implemented.** Documented as step 13 below.

---

## Correction Steps

Order is mandatory. The agent must report each step by number with verification evidence.

### 13. feat: add `exec` action for arbitrary bash commands

Resolves: F1.

1. Add `ActionExec = "exec"` to `internal/config/config.go`:
   - constant in the action list
   - entry in `validActions`
   - validation: `content` must be non-empty for `exec` (same as `text`); `target` rejected
2. Add entry to `internal/editor/actions.go` (`actionDefs`), between `text` and `copy`
3. In `internal/ui/ui.go` `press()`: new switch case for `exec` → `exec.Command("bash", "-c", b.Content).Start()`. Fire-and-forget, no stdout capture, non-blocking.
4. i18n: add `button.exec` and `editor.action_exec` in 7 languages (`internal/i18n/i18n.go`):
   - en: "Execute command"
   - pt-BR/pt-PT: "Executar comando"
   - es: "Ejecutar comando"
   - fr: "Exécuter commande"
   - de: "Befehl ausführen"
   - it: "Esegui comando"
5. Test: `TestExecActionRunsCommand` — spawns `echo radkeys-test`, asserts exit code 0.

**Verification:**
```
gofmt -w . && go vet ./... && go test ./...
go build -tags flatpak -o dist/radkeys-linux-amd64 .
```

### 14. fix: disambiguate firmware GET_VERSION reply (FIRMWARE SIDE)

Resolves: L1 (complete).

Implement the change documented in `firmware/rp2040-zero/PROTOCOL_FUTURE.md`:
- `diy.ino`: add distinct report ID for `GET_VERSION` (or 3-byte sentinel if report IDs don't work)
- `reader_cgo.go`: filter by report ID in version read and event loop
- `mock.go`: update fake device to prepend report ID
- `PROTOCOL.md`: document new report IDs

**Gate: REQUIRES PHYSICAL RP2040-ZERO.** The agent must not mark this step complete without Galvani's validation on hardware.

---

## Notes

- Firmware changes always require flash + validation on RP2040-Zero. No hardware = static review only.
- `golangci-lint` not installed in dev environment; `go vet` covers the essentials.
- macOS: source-only, no binary shipped. Build instructions in README.
- `research/`: kept for reference, not active development.
