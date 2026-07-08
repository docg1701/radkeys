# AGENTS.md — RadKeys

> Instructions for AI coding agents. Follow exactly. Dev cycle ends ONLY when
> the CI auto-release is complete and the GitHub release is published.

## Dev cycle (MANDATORY — follow every time)

```
1. Desenvolver → go test ./... → go vet ./... → gofmt
2. Bump version in radkeys.config.toml ([app] version)
3. Commit: fix: version bump X.Y.Z → A.B.C (context)
4. Push to main
5. git tag vA.B.C <sha>       ← LIGHTWEIGHT, NOT -a, NOT -m
6. git push origin vA.B.C
7. MONITOR: gh run watch <run-id> --exit-status
   Wait until CI passes → release auto-created by CI.
   The agent MUST NOT stop until the release is published.
```

## Commands

```bash
go build -o radkeys .        # build
./radkeys                     # run (mock without hardware)
go test ./... -v              # tests
gofmt -w . && go vet ./...    # format + vet
go mod tidy                   # deps

# Cross-compile tests (agent MUST run locally — NOT in CI)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o /dev/null ./...
CGO_ENABLED=1 GOOS=darwin  GOARCH=amd64 go build -o /dev/null ./...
CGO_ENABLED=1 GOOS=darwin  GOARCH=arm64 go build -o /dev/null ./...
```

## Testing

- Framework: Go standard `testing`.
- Every new function gets a test. Bug fixes get a regression test.
- Mock HID hardware with `hid.NewMock()`.

## Project Structure

```
radkeys/
├── main.go / go.mod
├── radkeys.config.toml      # Config de exemplo (comentado p/ humano/LLM)
├── internal/
│   ├── config/              # TOML parser + types
│   ├── deck/                # Navigation state machine
│   ├── hid/                 # HID reader (go-hid) + mock
│   ├── ui/                  # Fyne UI (Atalhos + Ajustes)
│   ├── i18n/                 # go-i18n + 7 JSON embed
│   ├── theme/               # 12 preset themes
│   └── assets/              # Icon (Obsidian) embed
├── firmware/arduino/        # Arduino Pro Micro firmware
├── firmware/rp2040/         # RP2040 firmware
└── research/                # Technical investigation notes
```

## Code Style

Go idiomático. Funções 4–20 linhas. Nomes específicos. Early return, max 2 níveis indent.

```go
// BOM
func (d *Deck) levelUp() {
    if len(d.stack) == 0 {
        d.current = d.cfg.Screens[0].ID
        return
    }
    d.current = d.stack[len(d.stack)-1]
    d.stack = d.stack[:len(d.stack)-1]
}
```

## Git Workflow

- `main` — stable. `feat/*` — features/fixes.
- Commits: conventional (`feat:`, `fix:`, `chore:`, `docs:`).
- Tags: **lightweight only** (`git tag vX.Y.Z <sha>`). NEVER `-a` or `-m`.

## Release

The CI `.github/workflows/build.yml` does:
1. Test + vet on ubuntu (Linux only — cross-compile is agent's job).
2. On tag push → build Linux binary + create GitHub release with:
   - `radkeys-linux-amd64` binary
   - `radkeys.config.toml` config template

Cross-compile for Windows/macOS is the agent's responsibility, tested locally:
```bash
GOOS=windows go build ./... && GOOS=darwin go build ./...
```

## Boundaries

### ✅ Always
- `gofmt -w . && go vet ./... && go test ./...` before commit.
- Conventional commits.
- Embed everything (icon, i18n, themes) → release = 1 binary + 1 config.
- Monitor CI after tag push until release is created.

### 🚫 Never
- Keyboard HID (F13-F24) input — rejected by product.
- `RequestAlwaysOnTop()` without verifying Fyne version.
- Hardcoded UI strings — use `i18n.T()`.
- Annotated tags (`git tag -a`, `git tag -m`).
- Cross-compile in CI — agent does it locally.
- End the turn before CI release is confirmed published.