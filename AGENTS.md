# AGENTS.md — RadKeys

> Instructions for AI coding agents working on this project. Follow exactly.

## Commands

```bash
# Build (Linux, needs GCC + libgl1-mesa-dev xorg-dev libudev-dev libxxf86vm-dev)
go build -o radkeys .

# Run (mock mode without hardware — UI works via mouse clicks)
./radkeys

# Tests
go test ./... -v

# Format + vet (ALWAYS run before commit)
gofmt -w . && go vet ./...

# Tidy deps
go mod tidy

# Cross-compile (needs Docker + fyne-cross)
fyne-cross linux -arch amd64
fyne-cross windows -arch amd64
fyne-cross darwin -arch amd64
fyne-cross darwin -arch arm64
```

## Testing

- Framework: Go standard `testing`.
- Location: `*_test.go` alongside source (e.g. `internal/config/config_test.go`).
- Run: `go test ./... -v`.
- Every new function gets a test. Bug fixes get a regression test.
- Mock external deps (HID hardware) with `hid.NewMock()`, never inline stubs.

## Project Structure

```
radkeys/
├── main.go                  # Entrypoint: load config → open HID → run UI
├── radkeys.config.toml      # Config de exemplo (comentado para humano/LLM)
├── internal/
│   ├── config/              # Parser TOML + validação + tipos (Config, Layout, Theme)
│   ├── deck/                # Estado de navegação (navigate/text/copy/level_up/go_home)
│   ├── hid/                 # Interface Reader + Mock + go-hid (Elgato/DIY) com build tags
│   ├── ui/                  # Fyne UI: aba Atalhos (preview+keypad) + aba Ajustes
│   ├── i18n/                 # go-i18n + arquivos JSON embed (en, pt-BR, pt-PT, es, fr, de, it)
│   ├── theme/               # 12 preset themes (10 terminal + 2 gray)
│   └── assets/              # Ícone embarcado (Obsidian icon theme)
├── firmware/
│   ├── arduino/             # Arduino Pro Micro (matriz 6×4, HID vendor-defined)
│   └── rp2040/              # RP2040 (24 GPIO diretos, Adafruit_TinyUSB)
├── research/                # Notas de investigação técnica
├── .github/workflows/       # CI: test + auto-release from tags
├── brief.md                 # Brief técnico (v2.0)
└── go.mod / go.sum
```

## Code Style

Go idiomático. Funções 4-20 linhas. Arquivos <500 linhas. Nomes específicos.
Sem `any`, sem `Dict`, sem funções sem tipo. Early return, máx 2 níveis de indentação.

```go
// BOM: nome específico, tipo explícito, early return
func (d *Deck) levelUp() {
    if len(d.stack) == 0 {
        d.current = d.cfg.Screens[0].ID
        return
    }
    last := d.stack[len(d.stack)-1]
    d.stack = d.stack[:len(d.stack)-1]
    d.current = last
}

// RUIM: nome vago, aninhamento, sem tipo
func doStuff(x interface{}) interface{} {
    if x != nil {
        if v, ok := x.(int); ok {
            return v
        }
    }
    return nil
}
```

## Git Workflow

### Branches

- `main` — estável, sempre compila e passa testes.
- `feat/*` — features e fixes. PR para `main` (ou fast-forward se solo).

### Commits (Conventional Commits)

```
feat: <descrição>        # nova funcionalidade
fix: <descrição>         # correção de bug
chore: <descrição>       # manutenção, deps, CI
docs: <descrição>        # documentação, brief, AGENTS.md
```

### Release & version bump

A versão vive em **UM lugar**: `radkeys.config.toml` → `[app] version`.
Tudo o mais é derivado/automatizado.

O **ciclo de desenvolvimento** é:
1. Desenvolver na branch `feat/*`.
2. `go test ./...` passa.
3. `gofmt -w . && go vet ./...` limpo.
4. Bump de versão em `radkeys.config.toml`.
5. Commit: `fix: version bump X.Y.Z -> A.B.C (contexto)`.
6. Push para `main`.
7. Criar tag **lightweight**: `git tag vX.Y.Z <sha>` (NÃO `git tag -a`, NÃO `-m`).
8. `git push origin vX.Y.Z`.
9. CI roda testes → se passarem, **auto-release** cria a release com:
   - Binários compilados (linux, windows, macos) como assets.
   - `radkeys.config.toml` como asset.
   - Changelog categorizado a partir dos conventional commits.
10. **NUNCA** criar/editar a release manualmente — o CI `release` job é dono.

### 🚫 NEVER (release)

- `git tag -a` / `git tag -m` — tags anotadas duplicam o título na release.
- `gh release create` / `gh release edit` — o CI é dono da release.
- Bump de versão em qualquer lugar que não `radkeys.config.toml`.
- Force-push de tag depois que o CI criou a release.

## Boundaries

### ✅ Always

- `gofmt -w . && go vet ./...` antes de commitar.
- `go test ./...` passando antes de push.
- Conventional commits (`feat:`, `fix:`, `chore:`, `docs:`).
- Validar APIs contra docs reais antes de usar (não confiar na memória de treino).
- Embed tudo no binário (ícone, traduções) — release = 1 executável + 1 config.

### ⚠️ Ask first

- Mudar dependência de versão major (Fyne, go-hid).
- Adicionar nova dependência externa.
- Mudar o protocolo HID ou o formato do config TOML.
- Mudar a arquitetura de pacotes.

### 🚫 Never

- Usar teclado HID (F13-F24) como input — foi rejeitado pelo produto.
- Usar `RequestAlwaysOnTop()` sem verificar que a versão do Fyne tem a API.
- Hardcoded de strings de UI — usar `i18n.T()`.
- Adicionar widgets editáveis focáveis na aba Atalhos.
- Criar tags anotadas (`-a`, `-m`).
- Editar a release manualmente — o CI é dono.
- Instalar Go em diretórios de gambiarra (`~/.local/go`) — usar apt ou método oficial.