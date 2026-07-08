# AGENTS.md — RadKeys

> Instructions for AI coding agents. Follow exactly.

## Dev cycle (MANDATORY — follow every time)

```
1. Desenvolver
2. gofmt -w . && go vet ./... && go test ./...
3. Bump version in radkeys.config.toml ([app] version)
4. Commit: fix: version bump X.Y.Z → A.B.C (context)
5. Push to main
6. Build all release binaries LOCALLY:
   go build -o radkeys-linux-amd64 .
   CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build -o radkeys-windows-amd64.exe .
7. macOS: build on a Mac (cross-compile from Linux is impossible — needs Apple SDK).
   On macOS: CGO_ENABLED=1 go build -o radkeys-macos-amd64 . && CGO_ENABLED=1 GOARCH=arm64 go build -o radkeys-macos-arm64 .
8. git tag vA.B.C <sha>       ← LIGHTWEIGHT, NOT -a, NOT -m
9. git push origin vA.B.C
10. MONITOR: gh run watch <run-id> --exit-status
    Wait until CI passes → release auto-created by CI.
11. Upload the locally-built binaries to the release:
    gh release upload vA.B.C radkeys-linux-amd64 radkeys-windows-amd64.exe
    (and macOS binaries if available)
    The agent MUST NOT stop until all binaries are in the release.
```

## Commands

```bash
# Build native (Linux)
go build -o radkeys-linux-amd64 .

# Cross-compile Windows from Linux
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build -o radkeys-windows-amd64.exe .

# macOS (on a Mac — cross-compile from Linux is impossible with CGO)
CGO_ENABLED=1 go build -o radkeys-macos-amd64 .
CGO_ENABLED=1 GOARCH=arm64 go build -o radkeys-macos-arm64 .

# Test
go test ./... -v
gofmt -w . && go vet ./...
go mod tidy
```

## Agent responsibilities

### ✅ Always
- `gofmt -w . && go vet ./... && go test ./...` before every commit.
- Conventional commits (`feat:`, `fix:`, `chore:`, `docs:`).
- Build Windows binary locally (mingw) and upload to the release.
- Build Linux binary and upload to the release.
- Monitor CI after tag push until release is published.
- Conventional commits for changelog.

### ⚠️ When possible
- Build macOS binaries on a Mac and upload to the release.
- MacOS cross-compile from Linux is **impossible** with CGO (needs Apple's proprietary SDK). Build on a Mac or skip.

### 🚫 Never
- Keyboard HID (F13-F24) input — rejected by product.
- `RequestAlwaysOnTop()` without verifying Fyne version (only available in ≥v2.8.0, not released yet).
- Hardcoded UI strings — use `i18n.T()`.
- Annotated tags (`git tag -a`, `git tag -m`) — lightweight only.
- Cross-compile in CI — agent does it locally.
- End the turn before CI release is published and all binaries are uploaded.
- End the turn before the release has Linux + Windows binaries.

## Release checklist (agent MUST complete)

- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] Version bumped in `radkeys.config.toml`
- [ ] `radkeys-linux-amd64` built and uploaded
- [ ] `radkeys-windows-amd64.exe` built (mingw) and uploaded
- [ ] macOS binaries built and uploaded (if Mac available)
- [ ] `git tag vX.Y.Z` (lightweight) pushed
- [ ] CI passed → release published by CI
- [ ] All binaries uploaded to the release

## Testing

- Framework: Go standard `testing`.
- Every new function gets a test. Bug fixes get a regression test.
- Mock HID hardware with `hid.NewMock()`.

## Project Structure

```
radkeys/
├── main.go / go.mod / go.sum
├── radkeys.config.toml      # Config example (commented for human/LLM)
├── internal/
│   ├── config/              # TOML parser + validation + types
│   ├── deck/                # Navigation state machine
│   ├── hid/                 # HID reader (go-hid + mock with build tags)
│   ├── ui/                  # Fyne UI (Atalhos + Ajustes)
│   ├── i18n/                 # go-i18n + 7 JSON embed
│   ├── theme/               # 12 preset themes
│   └── assets/              # Icon (Obsidian) embedded
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

The CI `.github/workflows/build.yml`:
1. Test + vet on ubuntu (Linux only).
2. On tag push → build Linux binary + create GitHub release with:
   - `radkeys-linux-amd64`
   - `radkeys.config.toml`

The agent then uploads the locally-built Windows (and macOS if available) binaries to the same release.