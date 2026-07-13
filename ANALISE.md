# RadKeys — Análise Profunda (Diagnóstico + Plano de Ação)

**Data:** 2026-07-12 · **Base:** `main` @ `1c704de` (v0.3.1) · **LOC:** ~2306 Go + 68 firmware

## Metodologia

8 revisores fresh-context (independentes, read-only) rodando em paralelo, cada um com um ângulo
distinto, inspecionando o repo real + validação empírica (`go test`, `go vet`, `gofmt`, snippet Go
para confirmar bug de cor, web research do TinyUSB/go-hid para o protocolo). Achados próprios de
primeira mão foram cruzados com os 7 relatórios por dimensão (arquitetura, tech debt, correção,
segurança, testes, firmware, build/CI). Dimensão cross-platform sintetizada pelo orquestrador
(revisor dedicado travou em tentativa de cross-compile; achados cobertos por inspeção própria +
corroboração dos revisores de tech debt/correção).

## Scores por dimensão

| Dimensão | Score | Estado |
|----------|-------|--------|
| Arquitetura & qualidade | 6/10 | Pacotes limpos, mas `ui.go` é god file (481 linhas) |
| Correção & bugs | 5/10 | Bugs reais no teardown do HID e em temas claros |
| Segurança | 5/10 | Superfície pequena, mas modelo baseado só em confiar USB |
| Testes & cobertura | 5/10 | config/i18n bons (91%); ui 3%, keystroke 0% |
| Cross-platform | 4/10 | Windows quebrado em 2 pontos; paste rouba foco |
| Dívida técnica | 5/10 | `buildSettings` 200 linhas; API Win deprecated |
| Firmware & protocolo | 2/10 | **Blocker: dispositivo não funciona com o host** |
| Build, CI & release | 5/10 | **Blocker: release Linux sem `-tags flatpak`** |

---

# 🔴 BLOCKERS (P0 — corrigir antes de qualquer release)

## B1 — Firmware envia report errado; dispositivo físico NÃO funciona com o host
**`firmware/rp2040-zero/diy.ino:59`** · confiado pelo revisor de firmware com citação de fonte

O descritor `TUD_HID_REPORT_DESC_GENERIC_INOUT(2)` **não declara report ID** (sem
`HID_REPORT_ID`). Mesmo assim o firmware chama `usb_hid.sendReport(1, report, sizeof(report))`.
O TinyUSB (`tud_hid_n_report`, `hid_device.c:104-107`) **prepend `0x01` quando `report_id != 0`**,
 enviando 3 bytes no fio: `[0x01, row, col]`. O host (`reader_cgo.go`, `diyReportLen=2`) lê só
2 bytes → `buf[0]=0x01` (Row sempre 1), `buf[1]=row` (vira Col), **coluna real truncada**.
Botões da mesma linha física ficam indistinguíveis. Confirmado contra o exemplo oficial da
Adafruit (`hid_generic_inout.ino`) que usa `sendReport(0, ...)` para o mesmo descritor.

**Fix (1 caractere):** `usb_hid.sendReport(0, report, sizeof(report));`
Corrigir também o comentário `diy.ino:19-22` ("report ID 1" → "single 2-byte vendor report, no
report ID") e `BUILD.md:146` que repete o erro. O host não precisa mudar.

## B2 — Release Linux do CI é buildado sem `-tags flatpak`
**`.github/workflows/build.yml:57`**

`CGO_ENABLED=1 go build -o /tmp/assets/radkeys-linux-amd64 .` — sem `-tags flatpak`, que
`AGENTS.md:14,29` e `README.md:74` mandam para diálogos de arquivo nativos via xdg-desktop-portal.
O binário publicado usa backend GTK em vez do portal → quebra em Wayland/sandbox. Diverge do
build local documentado.

**Fix:** `CGO_ENABLED=1 go build -tags flatpak -o /tmp/assets/radkeys-linux-amd64 .`

## B3 — Botões ficam PRETOS ao pressionar em temas claros
**`internal/theme/theme.go` `shift()`** · achado próprio, confirmado com snippet executável

`shift(c, factor)` faz `d := uint8(255 * factor)`. Para `factor = -0.08` (PRESSED em temas
claros, onde `sign()*-0.08 = +1*-0.08 = -0.08`), `uint8(-20.4)` wraps para **236**, e
`satSub(c, 236)` satura para **0**. Resultado reproduzido:

```
light_gray button [192,192,192] → PRESSED = [0,0,0]   (preto)
dracula     button [68,71,90]   → PRESSED = [88,91,110] (clareia, correto)
```

Ou seja, em `solarized_light`, `gruvbox_light` e `light_gray` (3 de 13 temas), pressionar um
botão o pisca de preto. Nenhum teste cobre `shift()`.

**Fix:** `d := uint8(255 * math.Abs(factor))` e ramificar add/sub pelo sinal (ou calcular a cor
destino diretamente). Adicionar teste de regressão para `shift` com factor negativo.

---

# 🟠 HIGH (P1 — bugs reais / risco significativo)

### H1 — `baseReader.Close()` dá panic em double-close
`reader_cgo.go:49` — `close(b.stop)` sem `sync.Once`. Segundo `Close()` → panic "close of closed
channel". `MockReader` usa `once.Do` (seguro); o real não (inconsistente). **Fix:** `sync.Once`.

### H2 — Reader real nunca fecha o canal `Events()` → `pollHID` vaza para sempre
`reader_cgo.go` `loop()` só fecha `done`, nunca `ch`. `ui.go:85` `pollHID` faz
`range u.reader.Events()` e bloqueia eternamente após erro de leitura/close. Hardware morre
silenciosamente, sem diagnóstico. **Fix:** fechar `ch` após o loop terminar (em `Close` depois de
`<-b.done`).

### H3 — Race em `MockReader.Put` → panic "send on closed channel"
`hid.go:48` — dá unlock antes do `select { case m.ch <- e: }`; `Close()` concorrente pode fechar
`ch` entre o unlock e o send. O doc diz "Safe to call before or after Close" mas não
concorrentemente. `TestMockReaderPutAfterCloseIsSafe` é sequencial (não pega). **Fix:** segure o
mutex durante o send, ou feche `ch` antes de `done` em `Close`.

### H4 — Loop do HID engole erros de leitura não-timeout silenciosamente
`reader_cgo.go:76` — erro não-timeout → `return` sem log. Device morto, app parece idle.
**Fix:** `log.Printf("radkeys: hid read failed: %v", err)` antes de retornar.

### H5 — Paste envia Ctrl+V para a janela focada, inclusive o próprio RadKeys
`ui.go:130-147` + `keystroke/*` — `press()` ActionPaste chama `keystroke.SendCtrlV()`
incondicionalmente. Clicar o botão no UI dá foco ao RadKeys → Ctrl+V vai pro app, não pro RIS.
Teclado físico preserva o foco do RIS (funciona), mas paste via UI é no-op. **Fix:** checar
janela foreground ≠ RadKeys antes de enviar (XGetInputFocus / GetForegroundWindow+PID /
AppleScript frontmost); senão, mostrar hint em vez de colar.

### H6 — `showConfigError` usa `xdg-open` em todas as plataformas → quebrado no Windows/macOS
`main.go:52` — `exec.Command("xdg-open", configPath).Start()`. Binário Windows é publicado mas o
botão "Open file to edit" silenciosamente não faz nada. **Fix:** dispatch por `runtime.GOOS`:
`xdg-open` (linux) / `cmd /c start "" <path>` (windows) / `open` (darwin).

### H7 — Config não valida idioma contra o set suportado
`config.go:103` — só default empty→"en"; aceita "xx"/"pt_BR". Idioma desconhecido → seletor vazio
na Settings, possível exibição de IDs crus. **Fix:** validar contra `i18n.Supported`.

### H8 — Config não valida theme preset
Preset desconhecido cai silenciosamente em `Presets[0]` (system) em `resolveFullTheme`.
**Fix:** validar via `themes.FindPreset` e rejeitar.

### H9 — Layout out-of-range é silenciosamente coercido → cascata confusa
`config.go:112` — `columns=20` vira 4 sem avisar; depois botão `col=5` erro "out of range [0,4)" e
o usuário não entende. **Fix:** rejeitar layout fora de [1,6] com erro explícito.

### H10 — Posições duplicadas de botões não são detectadas
Dois botões no mesmo `(row,col)` → `ButtonAt` retorna o primeiro, o segundo é mascarado em
silêncio. **Fix:** detectar重叠 em `validate()`.

### H11 — Windows paste usa `keybd_event` (deprecated) e ignora todos os erros
`keystroke_windows.go:19-27` — `keybdEvent.Call(...)` descarta retorno; `sendCtrlV()` sempre
retorna `nil`. Falha invisível. MS recomenda `SendInput`. **Fix:** `SendInput` + checar retorno
+ surfar erro. (Provavelmente `errcheck` do golangci-lint já flagaria — mas lint não roda.)

### H12 — Config default escrito `0o644` (world-readable)
`main.go:113` `ensureConfig` — contém nome do radiologista + templates de laudo. Em estação
compartilhada, outros usuários leem. O fixture de teste usa `0o600` (inconsistente).
**Fix:** `0o600` em `ensureConfig` e auditar `os.Create` em `ui.go:307`.

### H13 — CI pin Go 1.22 vs `go.mod` `go 1.24.0`
`build.yml:18,47` — depende de toolchain auto-download (rede, ~100MB, frágil em rede restrita).
**Fix:** `go-version: "1.24.0"`.

### H14 — CI nunca roda `golangci-lint`
`AGENTS.md` obriga antes de cada commit + release checklist, mas CI só roda build/test/vet.
Lint não está nem instalado localmente. Regressões de lint passam despercebidas. `errcheck`
provavelmente flag `keybdEvent.Call`. **Fix:** adicionar job `golangci-lint run ./...` no CI +
instalar localmente.

### H15 — CI nunca builda/testa com `-tags flatpak`
`build.yml:26,29,32` — caminho flatpak (xdg-desktop-portal) nunca compilado/testado em CI.
Compounds B2. **Fix:** matrix/step extra com `-tags flatpak` em build/test/vet.

### H16 — Settings Save destrói os comentários do config
`ui.go:317` `save()` — `toml.NewEncoder(f).Encode(cfg)` reescreve o TOML inteiro sem comentários.
README gabola "heavily commented" mas um Save apaga tudo. **Fix:** preservar comentários (editar
texto in-place / só atualizar campos mudados / usar template comentado), ou documentar a
perda e oferecer backup.

### H17 — `AGENTS.md` mente sobre injeção de versão
AGENTS.md diz versão "injected at build time (`-ldflags`)", mas CI não passa ldflags; a versão é
o literal em `main.go:19`. **Fix:** reconciliar — ou remover a claim de ldflags e manter o
bump no main.go, ou setar `var Version = "0.0.0-dev"` e injetar via ldflags no CI + dev cycle.

### H18 — `navigate` empilha mesmo quando target == current → stack cresce indefinidamente
`ui.go:140` — sempre `append(u.stack, u.current)`. Botão auto-referente cresce a pilha; `prev`
cicla na mesma tela. **Fix:** `if b.Target == u.current { break }`.

### H19 — `ensureConfig` default 4×3 inconsistente com docs/firmware 6×6
`main.go:69-89` — primeira execução sem config gera 4×3, não bate com BUILD.md/README/firmware.
**Fix:** default `columns = 6, rows = 6`.

### H20 — `ui.go` é god file (481 linhas); `buildSettings` ~200 linhas
Viola SRP e a própria regra do AGENTS.md (funções 4-20 linhas). UI knows serialização TOML.
**Fix:** split em `app.go`/`shortcut.go`/`settings.go`/`about.go`; mover `Save` pro pacote
`config` (`func (c *Config) Save(path) error`).

### H21 — Cobertura de testes crítica ausente
`go test -cover`: ui **3.0%**, keystroke **0%**, assets **0%**, main **0%**. `press()`,
navegação, copy/paste, `pollHID` — zero testes. `shift()`/`blend()`/`contrastOf()` sem teste
(não pegaram B3). i18n sem checagem de completude (todas as chaves × 7 idiomas). **Fix:**
testes para `press()` (todas as actions), `pollHID`+mock, color math, completude i18n.

### H22 — HID row/col confiados cegamente (sem clamp/auth)
`reader_cgo.go:74-81` — bytes crus do device viram Event sem checar contra o layout. Device
malicioso reivindicando `0x1234:0xABCD` pode dirigir o app. **Fix:** clamp ao layout, logar
out-of-range; campo opcional `serial` em `app.device` para bind a unidade específica.

### H23 — VID/PID placeholder `0x1234:0xABCD`
IDs de exemplo comuns; colisão trivial; não-USB-IF-compliant para distribuição.
**Fix:** obter VID/PID real ou documentar como protótipo-obrigatório-trocar.

---

# 🟡 MEDIUM (P2 — corrigir no próximo ciclo)

| # | Local | Problema | Fix |
|---|-------|----------|-----|
| M1 | `ui.go:84-93` | Icon path custom sem validação de tamanho/formato | Limitar tamanho, checar magic PNG, resolver path absoluto |
| M2 | `ui.go:145` | Copy escreve no clipboard global (exfiltração por monitors) | Documentar fluxo; futura paste direta via HID keyboard |
| M3 | `ui.go:317` | `save()` muta cfg antes de escrever; falha de write = mismatch | Validar/escrever primeiro, depois mutar; ou mutar cópia e swap no sucesso |
| M4 | `ui.go:300` | VID/PID/layout parse inválido silenciosamente coerced/ignorado | Dialog de validação com valor ofensivo |
| M5 | `ui.go:85` | Erro do `reader.Close()` no close da janela descartado | Logar não-nil |
| M6 | `reader_cgo.go:56` | `emit` dropa eventos silenciosamente se buffer cheio | Logar warning ou aumentar buffer; documentar trade-off |
| M7 | `main.go:83` | `ensureConfig` descarta erro de WriteFile | Retornar/logar erro |
| M8 | `keystroke_linux.go:9` | `xdotool` ausente só logado, sem check no startup | `exec.LookPath` no startup + dialog se faltar |
| M9 | `keystroke_darwin.go:8` | Permissão de accessibility não validada antes do Paste | Detectar/capturar erro e instruir usuário |
| M10 | `ui.go:340-343` | Save reconstrói tabs inteiras (stale widget refs) | Atualizar widgets in-place |
| M11 | `theme.go` `variantFor` | Comparação `th == DefaultTheme()` frágil (interface ==) | Type assertion / marker method |
| M12 | `ui.go:383` | `variantFor` depende de global `fyne.CurrentApp()` | Passar variant explicitamente |
| M13 | `config.go:41` / `ui.go:60` | `App.Name` carregado mas nunca usado; `titleBase` hardcoded | Usar `cfg.App.Name` como titleBase |
| M14 | `config.go:24` | `ValidActions` é map mutável exportado | Desexportar + helper `IsValidAction` |
| M15 | `reader_cgo.go:22` | `hid.Open` não retorna reader pronto (precisa outro `Open()`) | Renomear `NewReader`/`Connect` ou documentar two-phase |
| M16 | `diy.ino:60` | `delay(30)` debounce bloqueia scan loop | Debounce por timestamp (millis()) |
| M17 | `diy.ino:6,14,15` | Comentário "does NOT hardcode size" é falso (matriz 6×6 fixa) | Corrigir comentário |
| M18 | `config_test.go:15` | Fixture `version = "0.0.0-test"` é morto (Config não tem Version) | Remover ou comentar que é ignorado |
| M19 | firmware/host | Sem teste do protocolo DIY 2-byte | Teste table-driven simulando hidapi device |

---

# 🟢 LOW / NIT (P3 — limpeza quando conveniente)

| # | Local | Problema |
|---|-------|----------|
| L1 | `assets/assets.go` | `IconNames`/`IconData` + 8 ícones `icons/*.png` embedded são **dead code** (só `IconPNG` é usado) — remove ou usa |
| L2 | `assets/assets.go:27` | Strip manual de `.png` → `strings.TrimSuffix` |
| L3 | `theme.go:236` | Helper `bc` críptico → `newBaseColors` |
| L4 | `theme.go:228` | `catMocha` inconsistente com id `catppuccin_mocha` → `catppuccinMocha` |
| L5 | `main.go:28,34` | `configPath()` chamado 2x → calcular uma vez |
| L6 | `ui.go:50,295` | `appIconData` duplicado → helper `refreshIcon()` |
| L7 | `build.yml:94` | Regex de release notes não cobre `feat!:` (opcional) |
| L8 | Windows binary | Não testado em CI (intencional per AGENTS.md — aceitar/documented) |

---

# Plano de Ação Priorizado

## P0 — antes do próximo release (bloqueiam release correto)
1. **B1** firmware: `sendReport(1,...)` → `sendReport(0,...)` + corrigir comentários (`diy.ino`, `BUILD.md`). Testar com hardware.
2. **B2** CI: adicionar `-tags flatpak` ao build de release em `build.yml:57`.
3. **B3** theme: corrigir `shift()` para factor negativo + teste de regressão.
4. Rodar `gofmt -w . && go vet ./... && golangci-lint run ./... && go test ./...` (instalar golangci-lint).
5. Bump versão, commit `fix: ...`, tag, subir, monitorar CI, subir binários (Linux+Windows).

## P1 — próximo ciclo de desenvolvimento (bugs reais e riscos)
6. **H1/H2/H3** HID: `sync.Once` no `baseReader.Close`, fechar `ch` no fim do loop, fix race do `MockReader.Put` + teste concorrente.
7. **H4** logar erros de leitura não-timeout no loop do HID.
8. **H5** focus guard antes do Paste (não colar se RadKeys focado).
9. **H6** `showConfigError` dispatch por GOOS (xdg-open/start/open).
10. **H7/H8/H9/H10** config: validar idioma, theme, rejeitar layout out-of-range, detectar botões sobrepostos.
11. **H11** Windows: migrar `keybd_event` → `SendInput` + checar erros.
12. **H12** `ensureConfig` `0o600` (+ auditar `os.Create`).
13. **H13** CI `go-version: "1.24.0"`.
14. **H14** adicionar job `golangci-lint` no CI.
15. **H15** CI: step/matrix com `-tags flatpak` em build/test/vet.
16. **H16** Settings Save: preservar comentários do config (ou documentar perda + backup).
17. **H17** reconciliar AGENTS.md sobre injeção de versão.
18. **H18** `navigate` skip se target==current.
19. **H19** `ensureConfig` default 6×6.
20. **H21** testes: `press()` todas actions, `pollHID`+mock, color math (shift/blend/contrastOf), completude i18n.
21. **H22/H23** HID clamp ao layout + campo serial opcional; documentar/obter VID/PID real.

## P2 — quando tocar a área relevante
22. M1–M19 conforme tabela (validação de icon, mismatch do save, check xdotool/accessibility, debounce firmware, etc.).
23. **H20** refactor `ui.go` em arquivos focados + mover Save p/ config (faz junto com H16).

## P3 — limpeza de baixo risco
24. L1–L8 (remover dead code de assets, renames, configPath uma vez, etc.).

---

# O que está limpo (positivo)
- `gofmt`, `go vet`, `go test` limpos no working tree.
- Leftovers removidos confirmadamente: `internal/deck`, `firmware/arduino`, `firmware/rp2040` sumiram.
- Sem strings hardcoded na UI (tudo via `i18n.T()`); 7 idiomas com traduções completas.
- Validação de config robusta no que cobre (protocol, actions, ranges, navigate targets, duplicate screen ids).
- Abstração HID CGO vs !cgo via build tags limpa; mock bom para dev.
- `keystroke` isola por OS em arquivos separados.
- Tags são lightweight (conforme AGENTS.md); release notes geradas corretamente.
- Sem concatenação de shell para comandos externos (sem injection); sem network listeners/secrets.
- Version centralizada em `main.go` (fonte única da verdade, na prática).

# Riscos residuais a observar
- **B1 não confirmado em hardware** — a análise é estática + citação de fonte TinyUSB/Adafruit.
  Testar com o RP2040-Zero real após o fix de 1 caractere para validar o protocolo end-to-end.
- O modelo de segurança depende inteiramente de confiar no USB físico; sem autenticação de
  device, um gadget malicioso pode acionar copy/paste. Relevante se RadKeys for usado em contexto
  clínico/compartilhado.
- Windows binary nunca testado em CI nem pelo autor (política do projeto); riscos do path
  Windows (`keybd_event`, `xdg-open` em showConfigError) só são visíveis em runtime no Windows.

# Fontes dos achados
- Relatórios por dimensão: `/tmp/radkeys-analysis/0{1,2,3,4,5,7,8}-*.md` (revisores fresh-context).
- Validação empírica: `go test -cover`, snippet Go confirmando B3, web research TinyUSB/go-hid
  para B1 (citado no relatório 07-firmware.md).
- Inspeção própria do orquestrador para a dimensão cross-platform (revisor dedicado travou em
  cross-compile; achados corroborados pelos relatórios de tech debt e correção).