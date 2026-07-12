# AGENTS.md — RadKeys

> Instructions for AI coding agents. Follow exactly.

## Dev cycle (MANDATORY — follow every time)

```
1. Desenvolver
2. gofmt -w . && go vet ./... && golangci-lint run ./... && go test ./...
3. Bump version in main.go (var Version = "X.Y.Z"). NEVER hardcode version in config or other Go files.
4. Commit: fix: version bump X.Y.Z → A.B.C (context)
5. Push to main
6. Build all release binaries LOCALLY to dist/:
   go build -tags flatpak -o dist/radkeys-linux-amd64 .
   CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build -o dist/radkeys-windows-amd64.exe .
7. git tag vA.B.C <sha>       ← LIGHTWEIGHT, NOT -a, NOT -m
8. git push origin vA.B.C
9. MONITOR: gh run watch <run-id> --exit-status
    Wait until CI passes → release auto-created by CI.
10. Upload the locally-built binaries to the release:
    gh release upload vA.B.C dist/radkeys-linux-amd64 dist/radkeys-windows-amd64.exe
    The agent MUST NOT stop until all binaries are in the release.
```

## Commands

```bash
# Build native (Linux) — use flatpak tag for native file dialogs via xdg-desktop-portal
go build -tags flatpak -o dist/radkeys-linux-amd64 .

# Cross-compile Windows from Linux
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build -o dist/radkeys-windows-amd64.exe .

# macOS (on a Mac — cross-compile from Linux is impossible with CGO)
CGO_ENABLED=1 go build -o dist/radkeys-macos-amd64 .
CGO_ENABLED=1 GOARCH=arm64 go build -o dist/radkeys-macos-arm64 .

# Test
LINT_VERSION=v1.64.8
golangci-lint run ./...
go test ./... -v
gofmt -w . && go vet ./...
go mod tidy
```

## Agent responsibilities

### ✅ Always
- `gofmt -w . && go vet ./... && golangci-lint run ./... && go test ./...` before every commit.
- Conventional commits (`feat:`, `fix:`, `chore:`, `docs:`).
- Build Linux binary (priority — tested) and upload to the release.
- Build Windows binary locally (mingw) and upload to the release (fornecido, NÃO testado pelo autor).
- Monitor CI after tag push until release is published.
- Conventional commits for changelog.

### ⚠️ macOS
- NÃO entregamos binário macOS. SDK proprietário da Apple exige Mac físico — não pagamos por isso.
- Fornecemos instruções de build no README para quem tiver Mac.
- Comandos para referência:
  ```bash
  CGO_ENABLED=1 go build -o dist/radkeys-macos-amd64 .
  CGO_ENABLED=1 GOARCH=arm64 go build -o dist/radkeys-macos-arm64 .
  ```

### 🚫 Never
- Keyboard HID (F13-F24) input — rejected by product.
- `RequestAlwaysOnTop()` without verifying Fyne version (only available in ≥v2.8.0, not released yet).
- Hardcoded UI strings — use `i18n.T()`.
- Hardcoded version numbers in Go source or config files: version is set via `var Version` in `main.go` and injected at build time (`-ldflags "-X main.Version=X.Y.Z"`). Test fixtures use `"0.0.0-test"`.
- Annotated tags (`git tag -a`, `git tag -m`) — lightweight only.
- Cross-compile in CI — agent does it locally.
- End the turn before CI release is published and all binaries are uploaded.
- End the turn before the release has Linux + Windows binaries.
- Build ou upload de binário macOS — não é nossa responsabilidade.
- Screens are connected via `navigate` with `target`. Navigation is stack-based (`prev` goes back, `home` goes to root).
- `[app.fixed_buttons]` — removed. `copy`/`paste`/`prev`/`home` are normal actions.
- Firmware with fixed-size bitmap — use `(row, col)` 2-byte protocol.

## Release checklist (agent MUST complete)

- [ ] `golangci-lint run ./...` clean
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] Version bumped in `main.go` (var Version)
- [ ] `dist/radkeys-linux-amd64` built and uploaded
- [ ] `dist/radkeys-windows-amd64.exe` built (mingw) and uploaded
- [ ] `git tag vX.Y.Z` (lightweight) pushed
- [ ] CI passed → release published by CI
- [ ] Linux + Windows binaries uploaded to the release

## Testing

- Framework: Go standard `testing`.
- Every new function gets a test. Bug fixes get a regression test.
- Mock HID hardware with `hid.NewMock()`.

## Project Structure

```
radkeys/
├── main.go / go.mod / go.sum
├── radkeys.config.toml      # Config example (versionado)
├── BUILD.md                 # Guia de montagem do hardware
├── internal/
│   ├── config/              # TOML parser + validation + types
│   ├── hid/                 # HID reader (go-hid + mock)
│   ├── ui/                  # Fyne UI: preview + grid + settings + about
│   ├── i18n/                # single Go map (7 languages)
│   ├── theme/               # theme.go — 13 presets
│   └── assets/              # embedded icons
├── firmware/rp2040-zero/    # RP2040-Zero: TinyUSB, (row, col) protocol
└── research/                # investigation notes
```

> `internal/deck/` removed. Navigation is stack-based with screen ids.
> `firmware/arduino/` and `firmware/rp2040/` removed. Only RP2040-Zero.

## Code Style

- **All code, comments, error messages, and identifiers must be in English.**
- Idiomatic Go. Functions 4–20 lines. Specific names. Early return, max 2 indent levels.

```go
// GOOD: stack-based screen navigation
func (u *appUI) navigate(target string) {
    u.stack = append(u.stack, u.current)
    u.current = target
    u.renderGrid()
}
```

## Git Workflow

- `main` — stable. `feat/*` — features/fixes.
- Commits: conventional (`feat:`, `fix:`, `chore:`, `docs:`).
- Tags: **lightweight only** (`git tag vX.Y.Z <sha>`). NEVER `-a` or `-m`.

## Release

The CI `.github/workflows/build.yml`:
1. Test + vet on ubuntu (Linux only).
2. On tag push → build Linux binary + create GitHub release with:
   - `radkeys-linux-amd64`
   - `radkeys.config.toml`

The agent then uploads the locally-built Windows (and macOS if available) binaries to the same release.