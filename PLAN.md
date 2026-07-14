# RadKeys — PLAN.md

> Living plan for the next development cycle.
> Steps 1–15 executed (commits 575225a..HEAD). Release v0.13.10 shipped.
> What's done lives in `git log`, not here.

## Current state

- Version: `0.13.10` (`var Version = "0.13.10"` in `main.go:23`).
- Branch `main`. CI green. No hardware prototype yet.
- 13 actions (all shipped), 13 themes, 7 languages, 4 binaries per release.
- `0.x.x` only. `1.0.0` only after Galvani approves the hardware prototype.
- Dev cycle and build commands in `AGENTS.md`. Do not repeat here.

## Pending

| ID | Description | Gate |
|----|-------------|------|
| L1 | Firmware: `GET_VERSION` reply uses the same report ID (0) as button events. Host mitigates with retry but protocol ambiguity remains. Real fix is a distinct report ID or sentinel byte in the firmware — requires flash + test on RP2040-Zero. | **1.0.0** |

### L1 — firmware version vs button event ambiguity (NOT YET IMPLEMENTED)

**What's done:** host retries `GET_VERSION` 3 times on connect (`reader_cgo.go`). Mitigates a single stray button press but does not eliminate the root ambiguity: a `[row=1, col=0]` event is indistinguishable from `[major=1, minor=0]` on the wire.

**What's missing:** firmware-side fix (report ID or sentinel) documented in `firmware/rp2040-zero/PROTOCOL_FUTURE.md`. Implement when Galvani has hardware to validate.

**Gate for 1.0.0:** firmware-side fix validated on RP2040-Zero.

### F1 — `exec` action ✅ (implemented)

New button action that runs `bash -c "<command>"`. Fire-and-forget via `exec.Command`.

### E1–E4 — Config editor UX issues ✅ (implemented)

- **E1:** menu items padded with 8 non-breaking spaces for visual width.
- **E2:** already fixed (io.go uses `i18n.T("button.close")` since before step 15).
- **E3:** menu split into "Close file" (resets to blank) and "Quit" (exits).
- **E4:** `refreshTitle()` shows file path or "unsaved" label.

---

## Notes

- Firmware changes always require flash + validation on RP2040-Zero. No hardware = static review only.
- `golangci-lint` not installed in dev environment; `go vet` covers the essentials.
- macOS: source-only, no binary shipped. Build instructions in README.
- `research/`: kept for reference, not active development.
