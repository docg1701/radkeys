# RadKeys — Status & Backlog (handoff para o próximo ciclo)

> Estado em v0.9.0 (main @ `5390049`). Este arquivo é o handoff: diz o que JÁ FOI
> feito (não refazer) e o backlog restante, para retomar o desenvolvimento depois
> de compactar o contexto. A análise original de 8 dimensões foi executada e o
> diagnóstico virou 6 releases (0.4.0 → 0.9.0).

## ⚠️ Constraints do próximo ciclo (obrigatório)

- **Versionar sempre `0.x.x`. NÃO mudar para `1.0.0` sem ordem explícita do Galvani.**
- **Hardware: só o Galvani testa.** O agente não tem mão. Mudanças no firmware
  (`firmware/rp2040-zero/diy.ino`) devem ser marcadas como "requer validação
  manual no RP2040-Zero" e NÃO contar como verificadas até o Galvani confirmar.
- **Dev cycle do AGENTS.md (a cada release):** `gofmt -w . && go vet ./... &&
  golangci-lint run ./... && go test ./...` → bump `var Version` em `main.go` →
  commits (fix separados + 1 commit `fix: version bump X → Y (context)`) →
  `git push origin main` → build local (Linux `-tags flatpak` + Windows mingw)
  → `git tag vX.Y.Z <sha>` (lightweight) → `git push origin vX.Y.Z` →
  `gh run watch <run> --exit-status` → `gh release upload vX.Y.Z dist/radkeys-linux-amd64
  dist/radkeys-windows-amd64.exe --clobber`. NÃO encerrar até release publicada
  com Linux + Windows.

## Estado atual

| Item | Valor |
|------|-------|
| Versão | `0.9.0` em `main.go` (`var Version`) |
| Releases publicadas | v0.4.0, v0.5.0, v0.6.0, v0.7.0, v0.8.0, v0.9.0 (cada: Linux flatpak + Windows mingw + config, CI verde) |
| Cobertura | config 90% · theme 64% · i18n 69% · hid 31% · **ui 2,9%** · keystroke/main/assets 0% (48 testes) |
| CI | `.github/workflows/build.yml`: jobs `test` (build/test/vet **+ flatpak-tag**), `lint` (golangci-lint v1.64.8 via curl, Linux deps), `release` (build `-tags flatpak`, needs test+lint). Actions em Node 24 (`@v6`), `cache: false`, Go 1.24.0. Sem warnings Node/cache. |
| Toolchain local | go 1.24 · golangci-lint v1.64.8 em `$(go env GOPATH)/bin` (precisa no PATH) · mingw `x86_64-w64-mingw32-gcc` OK · `gh` auth `docg1701` · DISPLAY=:0 (X11, dá pra rodar o app) |

## ✅ DONE — não refazer (por release)

- **v0.4.0 (P0 blockers):** firmware `sendReport(0,...)` (B1 — o dispositivo não
  funcionava com o host: coluna truncada); CI release build com `-tags flatpak`
  (B2); theme `shift()` com `math.Abs` — botões não ficam mais pretos em temas
  claros (B3) + 2 testes de regressão.
- **v0.5.0 (P1a HID + cross-platform):** `baseReader.Close` idempotente via
  `sync.Once` (H1); loop do HID fecha o canal → `pollHID` não vaza (H2); race
  do `MockReader.Put` resolvida + teste concorrente com `-race` (H3); erros de
  leitura HID logados (H4); paste recusa clique no UI que rouba foco (H5);
  `openConfigEditor` por `runtime.GOOS` (H6 — xdg-open quebrava no Windows);
  Windows migrou `keybd_event`→`SendInput` com erro reportado (H11); config
  default `0o600` (H12); CI Node 24 + `cache:false` + Go 1.24.0 (H13 puxado).
- **v0.6.0 (P1b config + firmware):** valida idioma (H7), theme (H8), rejeita
  layout out-of-range (H9), detecta botões sobrepostos (H10); navigate no-op em
  self-target (H18); default config 6×6 (H19); `press()` loga evento HID fora
  do grid (H22 parte log); comment do firmware corrigido (M17); warning VID/PID
  no BUILD.md (H23 parte doc). + 4 testes de regressão.
- **v0.7.0 (P1c CI + save + bug do tema):** **fix do system-default preto na
  reabertura** — raiz: Fyne detecta a variante do OS async; no startup
  `ThemeVariant()` retorna Dark antes da detecção; adicionado
  `Settings().AddListener` em `Run()` que re-aplica as cores quando a variante
  assenta (corrige automaticamente, sem save). + CI golangci-lint job (H14),
  steps flatpak no test (H15), Save com backup `.bak` dos comentários (H16),
  AGENTS.md corrige a claim de ldflags (H17).
- **v0.8.0 (P1d testes + refactor):** testes de color math (blend/lerp/satAdd/
  satSub/contrastOf/setAlpha), completude i18n (7 langs), `config.Save` roundtrip;
  `config.Config.Save` (H20 parcial — UI não importa mais `toml`, config é dono
  da serialização).
- **v0.9.0 (P2+P3 cleanup):** dead assets removidos (IconNames/IconData + 8
  ícones — L1); `ValidActions` desexportado (M14); `titleBase` de `cfg.App.Name`
  (M13); renames `bc`→`newBaseColors` (L3), `catMocha`→`catppuccinMocha` (L4);
  `configPath()` uma vez (L5); fixture `version` removido (M18).

Scores da análise original: arquitetura 6 · correção 5 · segurança 5 · testes 5
· cross-platform 4 · tech debt 5 · firmware 2 · build/CI 5 (a maioria melhorou
com os releases; firmware e ui-cobertura ainda são os pontos fracos).

---

## 📋 REMAINING — backlog do próximo `0.x.x`

### 🔴 Deve ter (consistência / robustez / cobertura do core)
- **M3** — `ui.go` `save()` muta `cfg.App.*` em memória ANTES de `cfg.Save`; se o
  write falhar, o cfg in-memory diverge do disco e a UI reconstrói do estado
  mutado. Fix: validar/escrever primeiro e mutar só no sucesso (ou mutar uma
  cópia e trocar o ponteiro ao sucesso).
- **M11 + M12** — `variantFor(th)` em `ui.go` ainda faz `if th == DefaultTheme()`
  (comparação de interface frágil) e usa o global `fyne.CurrentApp()`. É a raiz
  estrutural do bug do system-default (parchei o sintoma com o listener em v0.7.0,
  mas a fragilidade permanece). Fix: derivar a variante da cor de fundo
  resolvida para custom themes e passar a variante explicitamente (remover o
  global e a comparação `==`).
- **M5** — `ui.go` `w.SetOnClosed(func(){ _ = reader.Close() })` descarta o erro.
  Fix: logar não-nil.
- **M6** — `reader_cgo.go` `emit` dropa eventos silenciosamente se o buffer de 64
  estiver cheio. Fix: logar warning quando dropar (ou aumentar buffer + doc).
- **M7** — `main.go` `ensureConfig` descarta o erro de `os.WriteFile`. Fix:
  retornar/logar (se não conseguir criar o config, o usuário só vê o erro
  genérico de "cannot read" depois).
- **M8** — `keystroke_linux.go` só roda `xdotool` no paste; ausência só é logada
  na hora. Fix: `exec.LookPath("xdotool")` no startup + dialog instrucional se
  faltar.
- **H21** — testes de `press()` (todas as actions: text/copy/paste/prev/home/
  navigate) e `pollHID`+mock. `ui` está em 2,9% de cobertura — é o calo. Usa o
  `fyne.io/fyne/v2/test` (app headless) + um `hid.MockReader`; para paste,
  injetar um `keystroke` falso (variável de pacote) pra não chamar o OS.

### 🟡 Deveria ter (qualidade)
- **H20 (restante)** — split do `ui.go` (514 linhas, `buildSettings` ~200) em
  `app.go`/`shortcut.go`/`settings.go`/`about.go` (mesmo pacote, mover funções,
  sem mudança de comportamento). `config.Save` já saiu da UI (v0.8.0); falta só
  a quebra em arquivos. Risco: imports por arquivo — validar com build+lint+test
  e uma rodada manual do app.
- **M1** — `appIconData` lê o icon custom sem validar tamanho/formato/path.
  Fix: limitar tamanho, checar magic PNG, resolver path absoluto.
- **M4** — no settings, VID/PID/layout inválido é silenciosamente coerced/ignorado.
  Fix: dialog de validação com o valor ofensivo.
- **M10** — `save()` reconstrói as tabs inteiras (stale widget refs). Fix:
  atualizar widgets in-place quando possível.
- **L7** — `build.yml` regex de release-notes não cobre `feat!:` (opcional).

### 🟢 Diferido — declarar como known-limitation (não fazer sem o pré-requisito)
- **M16** — debounce `delay(30)` bloqueia o scan loop no firmware. Benefício
  marginal (deck = uma tecla por vez) e **não-testável sem hardware**. Deferido
  até o Galvani validar no RP2040-Zero.
- **M9** — macOS accessibility não validada antes do Paste. Plataforma — sem Mac
  aqui. Documentar no README.
- **M15** — `hid.Open` é two-phase (retorna reader que precisa de outro
  `Open()`). Rename para `NewReader`/`Connect` toca a API pública; baixo valor.
- **M19** — teste do protocolo DIY 2-byte firmware↔host. Precisa de mock de
  `hid.Device` ou hardware.
- **L6** — `appIconData` chamado em 3 lugares (já é helper, não há duplicação
  real — reavaliar; provavelmente moot).
- **L8** — binário Windows não testado em CI. **Intencional per AGENTS.md**
  (cross-compile só localmente). Won't-do.

### ⛔ Bloqueadores externos para 1.0.0 (documentar, não resolver com código)
- **H23 real** — obter um VID/PID registrado (USB-IF) ou PID open-source alocado.
  Sem isso, `0x1234:0xABCD` são placeholders não-compliant para distribuição
  clínica. Já documentado como "trocar antes de produção" no BUILD.md.
- **B1 validação HW** — o fix `sendReport(0)` é estático + citado contra o fonte
  TinyUSB/Adafruit; precisa do Galvani testar no RP2040-Zero real end-to-end.
- **Windows runtime** — `SendInput` (H11) e `openConfigEditor` Windows não foram
  testados em runtime (sem Windows aqui). O `go vet GOOS=windows` e o cross-
  compile passam, mas o comportamento é não-verificado.

---

## 🛠️ Como retomar (notas operacionais)

- **golangci-lint** (não está no PATH default): `export PATH="$(go env GOPATH)/bin:$PATH"`.
- **Rodar o app** (DISPLAY=:0): `RADKEYS_CONFIG=/tmp/c.toml go run -tags flatpak .`
  — útil pra reproduzir bugs de UI. Para diagnóstico, adicionar `log.Printf`
  temporário em `Run()`/`save()` e rodar com `timeout 6s`.
- **Validar Windows sem Windows:** `GOOS=windows GOARCH=amd64 CGO_ENABLED=1
  CC=/usr/bin/x86_64-w64-mingw32-gcc go vet ./internal/keystroke/` (vet Linux
  não enxerga arquivos `//go:build windows`).
- **Build release local:** `go build -tags flatpak -o dist/radkeys-linux-amd64 .`
  e `CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc
  go build -o dist/radkeys-windows-amd64.exe .`
- **Monitorar CI:** `gh run list` → achar o run do tag → `gh run watch <id> --exit-status`.
- **Subagentes:** para validação adversarial de um diff, revisores fresh-context
  com instrução **static-only** (não rodar builds/testes — o pai já valida) evitam
  lentidão de cache frio. Para mudanças arriscadas, valide com `go test -race`
  e rode o app.

## Próxima versão sugerida: 0.9.1

Conteúdo: o bloco "Deve ter" (M3, M11+M12, M5, M6, M7, M8, H21) — consistência,
robustez e cobertura do core. Depois, incrementos 0.9.x com o "Deveria ter"
até o Galvani autorizar 1.0.0 (que depende de validação de hardware + VID/PID real).