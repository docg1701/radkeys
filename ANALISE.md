# RadKeys — Status, Auditoria de Antipatterns & Backlog (handoff)

> Estado em v0.9.0 (main @ `5390049`). Este arquivo é o handoff para o próximo
> ciclo depois de compactar o contexto. **A seção 🔴 Auditoria é a prioridade
> do próximo 0.x.x** — caça às gambiarras mascaradas nos 6 releases (0.4.0→0.9.0).

## ⚠️ Constraints do próximo ciclo (obrigatório)

- **Versionar sempre `0.x.x`. NÃO mudar para `1.0.0` sem ordem explícita do Galvani.**
- **Hardware: só o Galvani testa.** Mudanças no firmware (`firmware/rp2040-zero/diy.ino`)
  requerem validação manual no RP2040-Zero e NÃO contam como verificadas até o Galvani
  confirmar. Não usar "testei que funciona" sem o Galvani.
- **Princípio do Galvani para este ciclo:** caçar (a) mocks inúteis/testes que só
  testam a si mesmos, (b) fallbacks que escondem bugs e meia-implementação,
  (c) hardcoded que finge ser configurável e quebra bizarro/silencioso em rotina,
  (d) gambiarras em vez de causa raiz. **Resolver a causa real, não parchar o sintoma.**
- **Dev cycle do AGENTS.md (a cada release):** `gofmt -w . && go vet ./... &&
  golangci-lint run ./... && go test ./...` → bump `var Version` em `main.go` →
  commits (fix separados + 1 `fix: version bump X → Y (context)`) → `git push
  origin main` → build local (Linux `-tags flatpak` + Windows mingw) →
  `git tag vX.Y.Z <sha>` (lightweight) → `git push origin vX.Y.Z` →
  `gh run watch <run> --exit-status` → `gh release upload vX.Y.Z dist/radkeys-linux-amd64
  dist/radkeys-windows-amd64.exe --clobber`. Não encerrar até release com Linux+Windows.

## Estado atual

| Item | Valor |
|------|-------|
| Versão | `0.9.0` em `main.go` (`var Version`) |
| Releases | v0.4.0 → v0.9.0 (6 releases, cada: Linux flatpak + Windows mingw + config, CI verde) |
| Cobertura | config 90% · theme 64% · i18n 69% · hid 31% · **ui 2,9%** · keystroke/main/assets 0% (48 testes) |
| CI | jobs `test` (build/test/vet + flatpak-tag), `lint` (golangci-lint v1.64.8 via curl + Linux deps, timeout 10m), `release` (`-tags flatpak`, needs test+lint). Actions Node 24, `cache:false`, Go 1.24.0. |
| Toolchain | go 1.24 · golangci-lint v1.64.8 em `$(go env GOPATH)/bin` (precisa no PATH) · mingw OK · `gh` auth `docg1701` · DISPLAY=:0 (X11, dá pra rodar o app) |

---

## 🔴 AUDITORIA DE ANTIPATTERNS — prioridade do próximo 0.x.x (caça às gambiarras)

Cada item: **o quê / causa raiz / fix real (não gambiarra).** Ordenado por impacto.

### (A) Fallback silencioso pra mock — esconde HID quebrado
- **O quê:** `main.go:34-36` — `hid.Open` falha → `log.Printf` (stderr) + `reader = hid.NewMock()`. A UI sobe em modo mock **sem nenhuma indicação visual**.
- **Causa raiz:** tratar "device não encontrado / VID-PID errado / sem permissão udev" como caso normal e degenerar silenciosamente.
- **Fix real:** indicar no UI quando está em mock (banner/label "Sem dispositivo — modo mock") para o usuário não achar que o hardware funciona. O `log.Printf` no stderr não chega ao usuário. Não remover o mock (ele é legítimo p/ dev), mas **distinguir** e avisar.

### (B) `validate()` muta o struct + default de layout 4×5 inconsistente — quebra bizarro
- **O quê:** `internal/config/config.go:108-133` — `validate()` seta `Language="en"`, `Theme="system"`, `Columns=4`, `Rows=5` como side-effect. `validate()` com mutação é antipattern. E o default de layout é **4×5**, mas `ensureConfig` (main.go) e o hardware são **6×6**.
- **Causa raiz:** default hardcoded que finge ser configurável mas é inconsistente com o hardware.
- **Cenário bizarro:** usuário com config 6×6 apaga `[app.layout]` → vira 4×5 silencioso → botões em col 4/5 dão erro "out of range [0,4)" sem o usuário entender por quê.
- **Fix real:** (1) default de layout omitido = **6×6** (consistente com hardware/ensureConfig); (2) **separar defaults de validação** — aplicar defaults em `Load` explicitamente e fazer `validate` apenas rejeitar (sem mutar); ou, se mantém mutação, default 6×6 e documentar. Não deixar 4×5.

### (C) Gambiarra do tema (v0.7.0) — sintoma parcheado, causa raiz intacta
- **O quê:** `internal/ui/ui.go:80-95` — adicionei `Settings().AddListener` que re-renderiza quando a variante do OS "assenta". Isso corrigiu o fundo preto do system-default na reabertura.
- **Causa raiz (NÃO corrigida):** `variantFor` (`ui.go:487-495`) depende de `fyne.CurrentApp().Settings().ThemeVariant()` que é **assíncrono** (não-pronto no startup — M12) e usa `th == fyneTheme.DefaultTheme()` (comparação de interface frágil — M11). O listener é uma **gambiarra** que reage ao sintoma; a fragilidade estrutural permanece. Se o Fyne mudar o timing ou o listener não disparar, o bug volta.
- **Fix real:** fazer `variantFor` robusto: não depender de `ThemeVariant()` async no startup (derivar variante da cor de fundo resolvida para todos os temas, inclusive system — requer ler a cor que o Fyne resolve, ou um marker method em vez de `==`); remover o global `fyne.CurrentApp()` (passar a variante ou o app explicitamente). O listener pode ficar como cinto-de-segurança, mas a causa raiz é `variantFor`.

### (D) Teste de regressão que provavelmente não pega o bug (teatro sem `-race`)
- **O quê:** `internal/hid/hid_test.go` `TestMockReaderPutConcurrentCloseNoPanic` — 200 iterações Put+Close concorrentes, asserção "no panic".
- **Causa raiz:** a race window (send-on-closed-channel) é minúscula; **sem `-race`**, o panic é probabilístico e pode nunca disparar em 200 runs. O CI roda `go test ./...` **sem `-race`** → o teste provavelmente **passa no código bugado também**. O comentário do teste admite "rode com -race", mas o CI não passa.
- **Fix real:** CI rodar `go test -race ./...` (ou pelo menos `./internal/hid/`). Sem isso, o teste é teatro. (Custo: -race deixa o CI um pouco mais lento.)

### (E) Mock testa mock — o reader real não é testado
- **O quê:** `internal/hid/` — os testes só cobrem `MockReader`. O `diyReader` real (CGO) nunca é exercitado. O ciclo de vida do mock **não espelha** o real: o real fecha `ch` via `defer close(d.ch)` no `loop()` (stop OU erro de leitura); o mock fecha no `Close()`.
- **Causa raiz:** não há um fake do `*hid.Device` pra exercitar o `diyReader.loop()` (path "loop morre em erro de leitura → ch fecha → pollHID sai" NÃO testado).
- **Fix real:** introduzir uma interface interna `device` atrás do `baseReader` e um fake que simula timeout/erro/report, pra testar o `diyReader` de verdade (sem USB). Enquanto não tiver, o mock é só auto-validação.

### (F) `isLight` mutável em `Color()` — potencial data race
- **O quê:** `internal/theme/theme.go` `radKeysTheme.Color()` seta `t.isLight = v == VariantLight` como side-effect; `sign()` lê. Se `Color()` for chamada concorrentemente (Fyne pode), há data race em `t.isLight`.
- **Causa raiz:** campo mutável usado como estado temporário num método "de leitura."
- **Fix real:** computar `sign` localmente de `v` dentro de `Color()` (não armazenar `isLight`).

### (G) `shift()` overflow uint8 não-guardado
- **O quê:** `internal/theme/theme.go` `shift` faz `d := uint8(255 * math.Abs(factor))`. Para `|factor| > 1`, `uint8` faz wrap → cor errada silenciosa. Deixei sem guardião com o comentário "nenhum caller passa >1".
- **Causa raiz:** gambiarra (assumir uso) em vez de guardião.
- **Fix real:** clamp `factor` a `[0,1]` (ou saturar `d` a 255) para a função ser segura de reusar.

### (H) `emit` dropa evento silencioso ("shouldn't happen")
- **O quê:** `internal/hid/reader_cgo.go:65-70` — `select { case b.ch <- e: default: }` dropa se o buffer de 64 enche, comentário esperançoso "shouldn't happen".
- **Causa raiz:** mascarar perda de evento (num device de input médico, perder um key press é ruim).
- **Fix real:** logar warning quando dropar (M6), ou aumentar buffer, ou usar send bloqueante com shutdown — mas decidir o trade-off, não esconder.

### (I) Erros descartados (`_ =`) que escondem falhas
- **O quê / onde:** `ui.go:98` `reader.Close()` no close da janela (M5); `main.go:116` `ensureConfig` `WriteFile` (M7 — se falha, usuário só vê "cannot read" genérico depois); `config.go:232` backup `.bak` `WriteFile`; `config.go:238` `f.Close()` no Save; `i18n.go:250` `AddMessages` (message faltante → `T()` retorna ID cru).
- **Fix real:** logar/handlear cada um. Não engolir silenciosamente.

### (J) Erros só no `log.Printf` (não chegam ao usuário)
- **O quê / onde:** `ui.go:172` paste failed; `ui.go:153` device out-of-grid; `ui.go:220/304` icon read fail; `reader_cgo.go:101` HID read failed; `main.go:35` mock fallback; `main.go:54` config-open fail. Todos vão pra stderr; o usuário não vê.
- **Fix real:** erros acionáveis (paste falhou, device fora do grid, modo mock) deveriam aparecer no UI (status label/dialog), não só no log.

### (K) Paste via UI é meia-implementação mascarada como "by design"
- **O quê:** `internal/ui/ui.go:163-168` — `press()` recusa paste via clique no UI com um dialog (`paste.via_keypad_hint`), alegando que "paste é pelo teclado físico".
- **Causa raiz:** paste precisa do RIS focado; clicar no UI dá foco ao RadKeys. Em vez de resolver (refocus janela anterior — difícil cross-platform), eu **recusei** e chamei de design. Em modo mock (sem hardware), paste é inutilizável.
- **Fix real:** ou implementar de verdade (rastrear/refocus a janela anterior antes de enviar Ctrl+V), ou **documentar explicitamente** que paste só funciona pelo teclado físico e o modo mock não suporta paste — não fingir que está resolvido.

### (L) Outros hardcoded/fallback a confirmar
- **`main.go:104-106` ensureConfig template** com VID/PID `0x1234/0xABCD` — placeholder fingindo ser a identidade do device. Combinado com (A), usuário que não troca → mock silencioso. Documentado no BUILD.md, mas o template "parece configurado".
- **`internal/keystroke/keystroke_windows.go` SendInput** — struct layout "verificado por raciocínio", **não testado no Windows** (sem Windows aqui). Se errado, paste quebra silencioso no Windows. Meia-implementação. `go vet GOOS=windows` e cross-compile passam, mas runtime não-verificado. Aceito pelo Galvani, mas é risco.
- **`internal/ui/ui.go:503-508` `showFileDialog`** — `if err != nil || rc == nil { return }` engole erro do dialog silenciosamente.

### (M) `i18n` `init()` engole `AddMessages` e `T()` fallback esconde chaves faltantes
- `i18n.go:250` `_ = bundle.AddMessages(...)` — se falha, a chave faltante faz `T()` retornar o ID cru (ex.: "tab.shortcuts" aparece no UI). Silencioso. Confirmar que `TestAllMessagesHaveAllLanguages` pega; mas o `_ =` ainda esconde falha de registro.

---

## Priorização do 0.9.1 (causa raiz, não gambiarra)

**Bloco 1 — silenciosidade que esconde bugs (maior impacto no usuário):**
1. (A) mock fallback → indicar no UI
2. (B) validate() default 6×6 + separar defaults de validação (não mutar)
3. (I)+(J) erros `_ =` e `log.Printf` → logar/handlear e surfar no UI os acionáveis
4. (D) CI `go test -race` (+ confirmar o teste concorrente pega o bug)

**Bloco 2 — causa raiz estrutural:**
5. (C)+(F) `variantFor` robusto (M11/M12) + `isLight` race — a causa real do bug do tema
6. (G) `shift()` guardião de overflow
7. (H) `emit` logar drop (M6)
8. (K) paste via UI: implementar refocus OU documentar explicitamente (não fingir)
9. (E) fake de `hid.Device` pra testar o `diyReader` real (ou declarar known-gap)

**Bloco 3 — confirmar/limpar:**
10. (L) template VID/PID + SendInput não-testado — documentar como known-limitations
11. (M) i18n `_ =` AddMessages

**Para referência (já no REMAINING abaixo, não refazer neste bloco):** H20 file-split,
M1, M4, M10, L7 (deveria-ter); M16/M9/M15 (diferido); H23/B1/Windows (externo).

---

## ✅ DONE — não refazer (por release)

- **v0.4.0 (P0):** firmware `sendReport(0)` (B1); CI release `-tags flatpak` (B2); theme `shift()` `math.Abs` (B3) + testes.
- **v0.5.0 (P1a):** `baseReader.Close` `sync.Once` (H1); loop fecha `ch` (H2); race `MockReader.Put` (H3); log erros HID (H4); paste recusa UI (H5); `openConfigEditor` GOOS (H6); Windows `SendInput` (H11); config `0o600` (H12); CI Node24/cache/Go1.24 (H13).
- **v0.6.0 (P1b):** valida idioma/tema/layout/botões duplos (H7-H10); navigate no-op self (H18); default config 6×6 (H19); log out-of-grid (H22); comment firmware (M17); warning VID/PID (H23 doc) + testes.
- **v0.7.0 (P1c):** **fix do system-default preto** via `Settings().AddListener` (gambiarra — ver (C)); CI golangci-lint job (H14) + flatpak steps (H15); Save backup `.bak` (H16); AGENTS.md ldflags (H17).
- **v0.8.0 (P1d):** testes color math + completude i18n + `config.Save` roundtrip; `config.Config.Save` (H20 parcial — UI sem `toml`).
- **v0.9.0 (cleanup):** dead assets (L1); `ValidActions` desexportado (M14); `titleBase` de `cfg.App.Name` (M13); renames `bc`→`newBaseColors` (L3), `catMocha`→`catppuccinMocha` (L4); `configPath()` 1x (L5); fixture version removido (M18).

> ⚠️ Notar: alguns "DONE" acima são parches/gambiarras que a auditoria (seção 🔴)
> reabre — ex.: (C) o listener do tema, (K) paste fromUI, (H16) .bak. O próximo
> ciclo deve ir na causa raiz desses.

---

## 🟢 Diferido (declarar known-limitation; não fazer sem pré-requisito)
- **M16** debounce `delay(30)` no firmware — benefício marginal, **não-testável sem hardware**. Deferido até o Galvani validar no RP2040-Zero.
- **M9** macOS accessibility — sem Mac aqui. Documentar no README.
- **M15** `hid.Open` two-phase rename — baixo valor, toca API.
- **L6** `appIconData` "dup" — já é helper (provavelmente moot; reavaliar).
- **L8** Windows em CI — **intencional per AGENTS.md** (cross-compile só local). Won't-do.

## ⛔ Bloqueadores externos para 1.0.0 (documentar, não resolver com código)
- **H23 real** VID/PID registrado (USB-IF) — sem isso, `0x1234/0xABCD` são placeholders não-compliant pra distribuição clínica. Já no BUILD.md.
- **B1 validação HW** — o fix `sendReport(0)` é estático + citado contra fonte TinyUSB/Adafruit; precisa do Galvani testar no RP2040-Zero end-to-end.
- **Windows runtime** — `SendInput` (H11) e `openConfigEditor` Windows não testados em runtime. Cross-compile + `vet GOOS=windows` passam; comportamento não-verificado.

## 🛠️ Como retomar (notas operacionais)
- **golangci-lint** (não está no PATH default): `export PATH="$(go env GOPATH)/bin:$PATH"`.
- **Rodar o app** (DISPLAY=:0): `RADKEYS_CONFIG=/tmp/c.toml go run -tags flatpak .` — útil pra reproduzir bugs de UI. P/ diagnóstico, adicionar `log.Printf` temporário e rodar com `timeout 6s`.
- **Vet Windows sem Windows:** `GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=/usr/bin/x86_64-w64-mingw32-gcc go vet ./internal/keystroke/` (vet Linux não enxerga `//go:build windows`).
- **Build release:** `go build -tags flatpak -o dist/radkeys-linux-amd64 .` e `CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=/usr/bin/x86_64-w64-mingw32-gcc go build -o dist/radkeys-windows-amd64.exe .`
- **CI:** `gh run list` → achar run do tag → `gh run watch <id> --exit-status`.
- **Subagentes:** validação adversarial de diff = revisores fresh-context **static-only** (não rodar builds/testes — o pai já valida) pra evitar lentidão de cache frio. Mudanças arriscadas: `go test -race` + rodar o app.
- **Caça a antipatterns:** grep `'_ ='`, `'log.Printf'`, `'default:'`, `'NewMock'`, `'== 0'`, `'shouldn.t'` — as categorias (A)-(M) acima vieram daí. Releia o código tocado, não a memória.