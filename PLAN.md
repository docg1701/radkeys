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
| L1 | Firmware: `GET_VERSION` reply uses the same report ID (0) as button events. Host mitigates with retry but protocol ambiguity remains. Real fix is a distinct report ID or sentinel byte in the firmware — requires flash + test on RP2040-Zero. | **1.0.0** |
| F1 | New `exec` action — runs arbitrary bash command with user permissions. `content` = command. Host-side, fire-and-forget via `bash -c`. | Next release |
| E1 | File menu dropdown too narrow — standard Fyne popup width, visually cramped. Needs 20–30% wider. | Next release |
| E2 | File → "Sair" label hardcoded in pt-BR (`"Sair"`), does not follow UI language. Must use `i18n.T("button.close")`. | Next release |
| E3 | Both "Close" and "Sair" quit the program. "Close" should close the current `.toml` file (leave editor in no-file state); "Sair"/"Quit" should exit. | Next release |
| E4 | No visible file path indicator. Industry standard: window title includes the absolute path, or a read-only widget in the UI. | Next release |

### L1 — firmware version vs button event ambiguity

**What's done:** host retries `GET_VERSION` 3 times on connect (`reader_cgo.go`). Mitigates a single stray button press but does not eliminate the root ambiguity: a `[row=1, col=0]` event is indistinguishable from `[major=1, minor=0]` on the wire.

**What's missing:** firmware-side fix (report ID or sentinel) documented in `firmware/rp2040-zero/PROTOCOL_FUTURE.md`. Implement when Galvani has hardware to validate.

**Gate for 1.0.0:** firmware-side fix validated on RP2040-Zero.

### F1 — `exec` action

New button action that runs `bash -c "<command>"`. The command lives in `content`. Same permissions as the RadKeys process (user, not root). No shell injection risk: `exec.Command` does not interpolate strings — `-c` receives the argument as a parameter, not via concatenation.

**Not yet implemented.**

### E1–E4 — Config editor UX issues

**E1 — Menu width:** the Fyne `Menu` popup auto-sizes to the widest `MenuItem.Label`. Current labels (`New`, `Open`, `Save`, `Save As`, `Close`, `Sair`) are short and feel cramped. Fix: pad the shortest labels with Unicode non-breaking spaces (`\u00A0`), or define a custom theme that overrides menu item padding (`Size(fyne.ThemeSizeNamePadding)`). Target: 20–30% visual width increase.

**E2 — "Sair" hardcoded:** `internal/editor/io.go` `buildMenu()` has `fyne.NewMenuItem("Sair", …)` — a hardcoded pt-BR string. Replace with `i18n.T("button.close")`. The existing key `button.close` already has translations for all 7 languages.

**E3 — Close vs Quit confusion:** both menu actions call `onCloseIntercept` which closes the window. Correct semantics:
- **Close file** (`editor.close_file`, new key): reset config to a blank default, clear `e.path = ""`, disable save. Editor stays open but empty.
- **Quit** (`editor.quit`, new key): current `onCloseIntercept` behavior (discard check then exit).

New i18n keys (7 languages each):
- `editor.close_file` — en: "Close file", pt-BR: "Fechar arquivo", pt-PT: "Fechar ficheiro", es: "Cerrar archivo", fr: "Fermer le fichier", de: "Datei schließen", it: "Chiudi file"
- `editor.quit` — en: "Quit", pt-BR: "Sair", pt-PT: "Sair", es: "Salir", fr: "Quitter", de: "Beenden", it: "Esci"
- `editor.unsaved` — en: "unsaved", pt-BR: "não salvo", pt-PT: "não guardado", es: "sin guardar", fr: "non enregistré", de: "ungespeichert", it: "non salvato"

**E4 — File path indicator:** `refreshTitle()` in `editor.go` must include `e.path` when set. When `e.path == ""`, show the "unsaved" label. Example:

```go
func (e *Editor) refreshTitle() {
    if e.path != "" {
        e.win.SetTitle(fmt.Sprintf("%s — %s", i18n.T("editor.title"), e.path))
    } else {
        e.win.SetTitle(fmt.Sprintf("%s — %s", i18n.T("editor.title"), i18n.T("editor.unsaved")))
    }
}
```

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
3. In `internal/ui/ui.go` `press()`: new switch case for `exec` — `exec.Command("bash", "-c", b.Content).Start()`. Fire-and-forget, no stdout capture, non-blocking.
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

### 15. fix: editor menu width, i18n, close-vs-quit, and file path indicator

Resolves: E1, E2, E3, E4.

1. **E1 (menu width):** In `buildMenu()` (`internal/editor/io.go`), pad the shortest `MenuItem.Label` strings with `\u00A0` (Unicode non-breaking space) repeated until the popup looks balanced. Three-digit labels like "Save As" are naturally wider; "New" and "Quit" need the most padding.
2. **E2 (hardcoded "Sair"):** Replace `fyne.NewMenuItem("Sair", …)` with `fyne.NewMenuItem(i18n.T("button.close"), …)`.
3. **E3 (close vs quit):** Restructure File menu in `buildMenu()`:
   - `fyne.NewMenuItem(i18n.T("editor.close_file"), e.closeFile)` — new method `closeFile()` that calls `defaultConfig()`, sets `e.path = ""`, clears dirty, rebuilds tabs.
   - `fyne.NewMenuItem(i18n.T("editor.quit"), e.onCloseIntercept)` — current quit behavior.
   - Remove the old `fyne.NewMenuItem(i18n.T("button.close"), e.onCloseIntercept)`.
4. **E4 (file path in title):** Modify `refreshTitle()` in `editor.go` to include `e.path` when set. When `e.path == ""`, show `i18n.T("editor.unsaved")`. The dirty asterisk prefix is preserved in both cases.

New files / tests:
- `internal/editor/editor_test.go`: `TestCloseFileResetsConfig` — verifies `closeFile()` resets to default config and clears path.

New i18n keys: `editor.close_file`, `editor.quit`, `editor.unsaved` (7 languages).

**Verification:**
```
gofmt -w . && go vet ./... && go test ./...
go build -tags flatpak -o /tmp/radkeys-config-test ./cmd/radkeys-config
```

---

## Notes

- Firmware changes always require flash + validation on RP2040-Zero. No hardware = static review only.
- `golangci-lint` not installed in dev environment; `go vet` covers the essentials.
- macOS: source-only, no binary shipped. Build instructions in README.
- `research/`: kept for reference, not active development.
