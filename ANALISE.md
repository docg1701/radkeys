# RadKeys — Arquitetura-Alvo + Plano em Etapas (handoff pós-compactação)

> Decisão final (Galvani, 2026-07-13): o **paste sem roubar foco vira
> responsabilidade do FIRMWARE**, não do host. O RP2040-Zero vira um dispositivo
> USB **composite** (vendor HID `[row,col]` + **HID teclado** que manda Ctrl/Cmd+V
> sob comando). O App vira **configurador**: guarda TODA a configuração (frases,
> ações dos 36 botões) e **nunca escreve no device** depois do flash único de
> fábrica. **Startup-grab é aceito** (chama atenção do usuário — é feature, não
> bug). Logo **não precisa janela não-ativante / layer-shell / reescrita Rust** —
> Go+Fyne atende qualquer OS, porque o durante-o-uso não pinga (HID em
> background) e o startup-grab é bem-vindo.

---

## ⚠️ Constraints (obrigatório, inegociável)

- **Versionar sempre `0.x.x`.** NÃO mudar pra `1.0.0` sem ordem explícita do Galvani.
- **Hardware: só o Galvani testa.** Qualquer mudança no firmware
  (`firmware/rp2040-zero/`) requer flash + validação manual no RP2040-Zero e
  NÃO conta como verificada até o Galvani confirmar. Nunca dizer "testei que
  funciona" sem o Galvani ter flasheado.
- **Single executable por download.** O host (App) é um binário só por OS
  (Linux flatpak + Windows mingw). Nenhuma dep de shared-lib do sistema além
  das padrão (GL/X11/Wayland). **GTK4/Qt estão FORA** (linkam libs do sistema).
- **Device: flash único de fábrica, nunca mais escrito.** Nada de reflash por
  configuração, nada de escrever config no device, nada de desgaste de flash.
  Toda configuração vive no App (TOML). O device recebe só comandos
  **transitórios** em RAM (ex.: "fire paste"), nunca persiste nada.
- **Paste sem roubar foco, em qualquer macOS/Linux/Windows.** Via device ser
  teclado USB (nativo em qualquer OS, sem driver). Unicode via clipboard (host
  seta em text/copy), nunca via device digitar frase.
- **Código/comentários/erros em inglês. Idiomatic Go. Sem string hardcoded na
  UI — use `i18n.T()`.** Versão só em `var Version` no `main.go`.
- **Dev cycle (a cada release):** `gofmt -w . && go vet ./... &&
  golangci-lint run ./... && go test ./...` → bump `var Version` em `main.go` →
  commits (fix separados + 1 `fix: version bump X → Y (context)`) →
  `git push origin main` → build local (Linux `-tags flatpak` + Windows mingw) →
  `git tag vX.Y.Z <sha>` (lightweight) → `git push origin vX.Y.Z` →
  `gh run watch <run> --exit-status` → `gh release upload vX.Y.Z
  dist/radkeys-linux-amd64 dist/radkeys-windows-amd64.exe --clobber`.
  Não encerrar até release com Linux+Windows. **macOS: NÃO entregamos binário**
  (sem Mac); o código compila GOOS=darwin (device-command é cross-platform).

---

## A arquitetura-alvo (como vai funcionar de fato)

### Device (RP2040-Zero, firmware composite, flash único de fábrica)
- Interface **vendor HID** (igual hoje): no clique de qualquer botão, manda
  `[row, col]` (report IN). O host lê em background (sem roubar foco).
- Interface **HID teclado** (NOVA): só faz uma coisa — quando recebe um
  **comando vendor OUT** "fire paste [modificador]", envia o keystroke
  (Ctrl+V ou Cmd+V: modifier down, V down, V up, modifier up) como teclado USB.
- **O device não guarda config.** Não sabe qual botão é paste — o host decide
  (lê `[row,col]`, vê que é paste, manda o comando). Firmware de fábrica, uma
  vez, nunca mais escrito.

### Host (App Go+Fyne, single-binary, configurador)
- Lê `[row,col]` vendor (hid reader, igual hoje — background, sem foco).
- **text:** host seta clipboard pra frase + mostra preview (display). Nada pro
  RIS, sem foco. Unicode-safe (clipboard).
- **copy:** host seta clipboard pro previewText (igual hoje).
- **paste:** host lê `[row,col]` → vê que é paste → **manda comando "fire
  Ctrl/Cmd+V" pro device** (vendor OUT, 1 byte transitório em RAM) → **device
  manda o keystroke como teclado** → o RIS (janela focada) cola o clipboard no
  cursor. **Teclado nunca rouba foco** → garantido qualquer OS.
- **navigate (prev/home/navigate):** host muda de screen (estado interno do
  App). Sem foco, sem keystroke pro RIS.
- **Invariante de foco:** o App **nunca** ergue/foca a própria janela ao
  handle evento HID. text/copy/navigate são silenciosos (atualizam a janela de
  fundo + clipboard, sem raise). Só o paste manda keystroke pro RIS (desejado).
- **Modifier por OS:** `runtime.GOOS == "darwin"` → Cmd (GUI); senão Ctrl. O
  comando carrega o modificador. (macOS passa a funcionar no código; não
  shipamos binário macOS.)

### Garantias (honestas, sem mentira)
- **Device manda Ctrl/Cmd+V como teclado USB:** ABSOLUTAMENTE GARANTIDO em
  qualquer macOS/Linux/Windows — teclado USB é nativo, sem driver, sem
  software. (Bala-proof.)
- **Durante o uso (1000 cliques), cursor não pinga no RIS:** GARANTIDO qualquer
  OS — text/copy/navigate são background (HID read + render de fundo +
  clipboard-set, nada foca o App); só paste manda keystroke pro RIS (desejado).
- **Startup-grab:** ACEITO (feature). Não precisa janela não-ativante → **Fyne
  serve em qualquer OS, Wayland incluído** (sem layer-shell, sem reescrita). O
  usuário clica no RIS depois do lançamento e aí os 1000 cliques fluem sem ping.
- **Unicode:** via clipboard (host), não via device. ✓
- **36 teclas configuráveis:** tudo no App (frases, ação por botão). O device é
  genérico. ✓
- **Single-binary:** host fica MAIS simples (deleta o pacote `keystroke` e a
  injeção OS-específica). ✓

---

## Estado atual (ponto de partida)

- `var Version = "0.9.0"` em `main.go` (intacto).
- **Bloco 1 — antipattern cleanup Fyne-side (4 commits, CI verde @ 4bba0f9):**
  config `validate` pura + default 6×6; status label (mock mode + erros
  acionáveis surfados na UI); erros `_=` logados; `ensureConfig` fail-loud;
  `showConfigError` i18n-ado; CI `go test -race` + CGO + timeout; teste
  concorrente do hid endurecido. **Mantém** (Fyne segue na arquitetura nova).
- **Bloco 2 — causa-raiz estrutural (2 commits @ 8085d90):**
  `variantFor` determinístico (marker interface + fallback explícito, sem
  interface `==` nem global async); `isLight` race removido (sign local);
  `shift()` guardião de overflow; `emit()` loga drops com throttle;
  `hidDevice` interface + fake pro `diyReader.loop` (lifecycle testável).
  **Mantém** (Fyne theme + hid reader vendor seguem).
- Esses 6 commits estão em `main`, **sem bump de versão / sem tag / sem
  release** ainda.
- Toolchain: go 1.24 · golangci-lint v1.64.8 em `$(go env GOPATH)/bin` (botar no
  PATH) · mingw OK · `gh` auth `docg1701` · DISPLAY=:0 (X11).

---

## Plano em etapas (pi-subagents, pequenas e exequíveis)

> Princípio: cada etapa = planejador/worker/validador fresh-context (pai é
  orchestrator single-thread writer), validação focada, e **etapas de firmware
  precisam do Galvani flashear + testar no hardware**. Rodar
  `gofmt -w . && go vet ./... && golangci-lint run ./... && go test -race ./...`
  antes de commitar. Seguir o SKILL pi-subagents (staged fix orchestration:
  fanout planejamento só-leitura → 1 worker escritor → fanout validação
  só-leitura → pai commita).

### Etapa 0 — Release 0.9.1 (cleanup já pronto)
- **O quê:** bump `0.9.0 → 0.9.1` em `main.go`, build Linux flatpak + Windows
  mingw, `git tag v0.9.1 <sha>` (lightweight), push, `gh run watch`, upload dos
  binários. Os 6 commits de antipattern (Bloco 1+2) já estão em `main` e CI
  verde — só falta o bump+tag+release.
- **Subagent:** nenhum (pai executa o dev cycle direto).
- **Validação:** CI verde + release com Linux+Windows.
- **Por quê:** shipar o cleanup antes da feature de firmware (changelog limpo).

### Etapa 1 — Firmware: composite USB (vendor + teclado) + protocolo fire-paste
- **O quê:** reescrever `firmware/rp2040-zero/diy.ino` como composite TinyUSB:
  interface vendor IN (`[row,col]`, igual hoje) + interface vendor OUT (recebe
  comando "fire paste [mod]") + interface **HID teclado** (envia
  Ctrl/Cmd+V + release quando comandado). Definir o protocolo do comando
  vendor OUT (ex.: byte 0 = cmd `0x01` fire-paste, byte 1 = modificador
  `0x01`=Ctrl / `0x02`=GUI/Cmd). Documentar o protocolo num `PROTOCOL.md`.
- **Subagent:** `planner`/`reviewer` (pesquisar TinyUSB composite + HID
  keyboard no RP2040-Zero, ler o firmware atual e a lib TinyUSB/Adafruit) →
  `worker` (escrever o firmware composite + o PROTOCOL.md). O pai valida
  estaticamente (descritores HID coerentes) — **mas só o Galvani flasheia e
  testa no hardware.**
- **Validação:** revisão estática do firmware (descritores, lógica do
  teclado, handling do comando). **Galvani:** flash no RP2040-Zero, conferir
  que o device aparece como vendor+keyboard, e que mandando o comando (via
  app de teste) dispara Ctrl+V na janela focada. **HARDWARE = Galvani.**
- **Risco honesto:** firmware composite TinyUSB é trabalho real (descritores
  HID de 2 interfaces + lógica de teclado + recepção de comando OUT). É
  bounded e padrão, mas não é trivial. Uma vez feito, nunca mais se toca.

### Etapa 2 — Host: device-command writer (vendor OUT fire-paste) + mock
- **O quê:** novo código no pacote `hid` (ou novo pacote `device`): função
  `FirePaste(mod Modificador)` que escreve o report vendor OUT de 2 bytes no
  device. Definir `Modificador` (Ctrl/Cmd) por `runtime.GOOS`. Criar uma
  interface interna pra poder mockar o write (testável sem USB).
- **Subagent:** `worker` (implementa o writer + a interface mockável + testes
  unitários do writer com o mock).
- **Validação:** `go test -race ./internal/hid/` (mock), build Linux/Windows,
  `GOOS=darwin go vet ./internal/hid/` (compila no mac). Write real no device é
  testado pelo Galvani com o firmware da Etapa 1.
- **Sem hardware:** o host code é testado com mock; a integração real é Galvani.

### Etapa 3 — Host: rewire do paste + deletar a injeção de keystroke
- **O quê:** em `internal/ui/ui.go`, `case config.ActionPaste`: trocar
  `keystroke.SendCtrlV()` por `device.FirePaste(modPorOS())`. **Deletar o pacote
  `internal/keystroke`** inteiro (SendCtrlV + keystroke_darwin/linux/windows.go)
  — não injeta mais nada no OS; o device é o teclado. Confirmar que text/copy
  seguem setando clipboard (host-side, Unicode-safe) e navigate segue mudando
  screen. Atualizar i18n/tests que referenciavam keystroke.
- **Subagent:** `reviewer` (auditar o que depende de `keystroke`) → `worker`
  (rewire + deletar + ajustar imports/tests).
- **Validação:** `go build -tags flatpak`, `go test -race ./...`,
  `GOOS=windows ... go build`, `GOOS=darwin go build` (confirma que compila
  sem o keystroke). App roda (DISPLAY=:0). **Galvani:** testa paste real com o
  firmware da Etapa 1 (device manda Ctrl+V → RIS cola).
- **Nota:** macOS passa a ser suportado no código (sem per-OS keystroke;
  device-command é cross-platform). Não shipamos binário macOS.

### Etapa 4 — Host: invariante de não-roubo-de-foco em evento HID
- **O quê:** garantir que `press()`/`pollHID()` **nunca** ergam/focam a janela do
  RadKeys ao handle um `[row,col]`. Auditar `ui.go` (nenhum `RequestFocus`,
  `Show`, raise, `SetContent` re-trigger que foca a janela em path de HID).
  Adicionar invariante documentada (comentário +, se possível, um teste/guard).
- **Subagent:** `reviewer` (auditar paths de HID em busca de raise/foco) →
  `worker` (corrigir se houver + documentar a invariante).
- **Validação:** revisão de código + **Galvani** testa o "1000 cliques, cursor
  piscando no RIS sem ping" (text/copy/navigate silenciosos; só paste manda pro
  RIS). Testar em Linux Xorg, Linux Wayland, Windows (macOS se tiver Mac).
- **Sem firmware:** só host. Mas a confirmação visual "sem ping" é Galvani.

### Etapa 5 — (Opcional) Firmware version check one-shot
- **O quê:** ao conectar o device, o App lê a versão do firmware **uma vez** e
  avisa se for antiga ("atualize o firmware uma vez"). Não enche por uso.
- **Subagent:** `worker` (ler versão do device via vendor + dialog/aviso).
- **Validação:** Galvani testa (firmware antigo avisa, novo fica silencioso).
- **Pode pular** se o Galvani preferir não ter check.

### Etapa 6 — Documentação final (tudo atualizado)
- **O quê:** ao final do desenvolvimento (antes do release 0.10.0), atualizar
  TODA a documentação do projeto pra refletir a arquitetura nova e o estado
  real. No mínimo: `README.md` (arquitetura paste-via-firmware-teclado-USB,
  app=configurador, single-binary, startup-grab aceito, macOS suportado no
  código sem binário shipado, pacote `keystroke` removido, check one-shot de
  versão do firmware), `BUILD.md` (montagem do hardware + nota do device
  composite USB vendor+teclado + flash único de fábrica), `PROTOCOL.md`
  (referenciado pela Etapa 1 — confirmar coerente com o firmware final),
  `radkeys.config.toml` (exemplo versionado coerente com os campos/usos
  atuais), e o próprio `ANALISE.md` (marcar etapas 0-7 como feitas, refletir
  o reframe "sem hardware protótipo ainda → validação estática/mock só; flash
  real quando o protótipo ficar pronto; 0.x.x até aprovação no hardware,
  1.0.0 só depois", e reescrever TODA linguagem stale de GATE/flash nas
  validações das etapas 1/3/4). Conferir qualquer outro `.md` do repo e
  atualizar se estiver stale. Nenhum doc pode contradizer o código shipped.
- **Subagent:** `reviewer` (auditar TODA doc contra o código final — achar
  stale: arquitetura antiga, GATEs de hardware como se existisse, campos de
  config defasados, menções ao pacote `keystroke`, etc.) → `worker` (reescrever
  cada doc stale; identifiers em inglês, i18n onde aplicável).
- **Validação:** revisão de código da doc (sem contradição com o código
  shipped), `go test ./...`/build seguem verdes, pai confere o diff de docs
  antes de commitar. Sem hardware: nada depende de flash.
- **Por quê:** o release 0.10.0 shipa com doc correta; handoff pra quando o
  protótipo ficar pronto (semanas) é limpo.

### Etapa 7 — Release 0.10.0 (feature de firmware)
- **O quê:** bump `0.9.1 → 0.10.0`, build, tag `v0.10.0`, push, CI, upload
  binários Linux+Windows. Release notes: "paste agora via firmware
  (teclado USB); não rouba foco; macOS suportado no código; keystroke package
  removido."
- **Subagent:** nenhum (pai executa dev cycle).
- **Validação:** CI verde + release com Linux+Windows + **Galvani confirma o
  fluxo completo no hardware** antes do tag.

---

## Ordem de execução sugerida (com pi-subagents)

1. **Etapa 0** (pai direto): release 0.9.1 do cleanup já pronto.
2. **Etapa 1** (planner→worker, **validação estática**): firmware composite.
   → **Sem hardware protótipo ainda:** validação é estática (descritores HID
   coerentes, lógica do teclado, handling do comando OUT) + cross-check do
   `PROTOCOL.md`. Flash real no RP2040-Zero fica pra quando o protótipo ficar
   pronto (semanas); até lá nada é "testei que funciona no hardware".
3. **Etapa 2** (worker, mock): device-command writer (paralelo à 1? sim, sem
   conflito — 2 só toca hid, 1 só firmware; mas o protocolo da 1 define os
   bytes que a 2 escreve → fazer 1 antes de 2, ou 1+2 juntos com o planner
   definindo o protocolo primeiro).
4. **Etapa 3** (reviewer→worker): rewire paste + deleta keystroke. Depende de
   2 (o writer).
5. **Etapa 4** (reviewer→worker): invariante de foco. Depende de 3 (paste
   rewire) pra testar o fluxo real (mock).
6. **Etapa 5** (worker): version check one-shot (incluído por decisão Galvani).
7. **Etapa 6** (reviewer→worker): documentação final — TODA doc atualizada
   contra o código shipped (reescreve também a linguagem stale de GATE/flash
   das etapas 1/3/4). Depende de 1-5 prontos.
8. **Etapa 7** (pai direto): release 0.10.0.

**Dependência crítica:** sem hardware protótipo ainda, toda validação é
estática + mock + cross-compile (Linux flatpak, Windows mingw, `GOOS=darwin
go vet`). 0.x.x incremental até tudo pronto; `1.0.0` só após aprovação no
hardware protótipo. As etapas host (2-5) podem ser codadas + testadas com
mock em paralelo à firmware (1), mas o protocolo da 1 define os bytes que a 2
escreve.

---

## Como retomar (notas operacionais)

- **golangci-lint** (não tá no PATH default): `export PATH="$(go env GOPATH)/bin:$PATH"`.
- **Rodar o App** (DISPLAY=:0): `RADKEYS_CONFIG=/tmp/c.toml go run -tags flatpak .`
  — útil pra reproduzir bugs de UI. Diagnóstico: `log.Printf` temporário +
  `timeout 6s`.
- **Vet/build cross-OS sem o OS:** `GOOS=windows GOARCH=amd64 CGO_ENABLED=1
  CC=/usr/bin/x86_64-w64-mingw32-gcc go build -o dist/radkeys-windows-amd64.exe .`;
  `GOOS=darwin go vet ./...` (macOS compila, não linka sem Mac).
- **Build release:** `go build -tags flatpak -o dist/radkeys-linux-amd64 .` +
  Windows mingw.
- **CI:** `gh run list` → achar run do tag → `gh run watch <id> --exit-status`.
- **pi-subagents:** seguir SKILL pi-subagents — staged fix orchestration
  (fanout planejamento só-leitura fresh → 1 worker escritor → fanout validação
  só-leitura fresh → pai synthesize + commit). Validadores **static-only**
  (não rodam build/test — o pai já valida) pra evitar lentidão de cache frio.
  Mudanças arriscadas: `go test -race` + rodar o App. Worker NÃO toca `var
  Version` nem commita (pai commita em unidades lógicas). Etapas de firmware
  não são "testei que funciona" até o Galvani flashear.
- **Firmware → Galvani:** qualquer PR no `firmware/rp2040-zero/` é só código
  estático até o Galvani flashar + testar no RP2040-Zero.

---

## Histórico (referência, não refazer)

- **Antipattern cleanup (Bloco 1+2, commits 5e1af11..8085d90):** caça às
  gambiarras dos 6 releases (0.4.0→0.9.0) — config pura + 6×6, status label,
  erros surfados, CI -race, variantFor determinístico, isLight race,
  shift guard, emit log, hid lifecycle testável. Tudo em `main`, CI verde.
- **Decisão de arquitetura (esta seção):** paste via firmware-teclado (não
  host-injeção), app=configurador, single-binary, startup-grab aceito. Rejeitou
  GTK4/Qt (single-binary), reescrita Rust (startup-grab aceito tornou
  desnecessária), reflash por config (idiota), escrita de config no device.