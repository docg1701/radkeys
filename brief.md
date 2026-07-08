# RadKeys — Brief Técnico Completo

> **Versão:** 1.3
> **Data:** 2026-07-07
> **Autor:** Nonatinho (consultor)
> **Cliente:** Galvani (radiologista)
> **Repo:** https://github.com/docg1701/radkeys
>
> **v1.3 — correção do always-on-top (alucinação removida).** O `desktop.Window.RequestAlwaysOnTop()` **não existe no Fyne v2.7.4 lançado**: o PR #6184 foi mergeado em `develop` (linha v2.8.0, ainda rc1), não em `release/v2.7.x`. MVP fica em v2.7.4 estável sem always-on-top; re-adicionar ao subir para v2.8.0 estável. Investigação em `research/fyne-always-on-top.md`.
>
> **v1.2 — modelo de input redefinido para HID custom via hidapi.** Dispositivo USB dedicado (Stream Deck/clone pelo protocolo Elgato, ou DIY de 24 com firmware RadKeys) lido diretamente pelo app. Sem teclas, sem modificadores, sem roubar foco do RIS, botões ilimitados conforme o hardware. Alvo: 24 teclas (aceita 18–20).

---

## 1. Visão do produto

RadKeys é um aplicativo desktop portátil, open source, multiplataforma, para médicos radiologistas que **digitam laudos**. Ele exibe uma interface visual de atalhos hierárquicos (shortcut deck) com preview central. Cada atalho carrega uma frase pronta de laudo. O usuário confirma a cópia, o texto vai para a área de transferência e é colado no RIS/PACS.

A operação é feita por um **dispositivo USB de botões (HID custom)**, lido diretamente pelo RadKeys — o hardware **não envia teclas** ao sistema, logo **não rouba o foco do RIS** e **não interfere** em atalhos do app focado. Cliques de mouse na janela do RadKeys mudam o foco normalmente (só no modo de edição).

### Modelo de botões (N botões físicos, lidos via hidapi)

| Botões | Função | Configurável? |
|--------|--------|---------------|
| **3 fixos** (índices configuráveis, default 0/1/2) | `copy` (clipboard), `level_up` (sobe nível), `go_home` (volta à raiz) | Fixa (global, constante em qualquer tela) |
| **N−3 configuráveis** (demais índices) | botões da tela ativa (`navigate` ou `text`) | Configurável por tela |

- **N = nº de botões do hardware**, descoberto em runtime (Elgato: via `Get Unit Information`, rows×cols; DIY: fixo 24). O RadKeys adapta-se ao hardware — **ilimitado conforme o dispositivo**.
- Alvo **24 teclas**: DIY 24 → 21 configuráveis; Stream Deck XL (32) → 29; Stream Deck original (15) → 12. Cobre o intervalo pedido (18–24 configuráveis com DIY 24 ou XL).
- As **3 funções fixas** são globais (constantes em qualquer tela).
- Os **N−3 configuráveis** mudam conforme a **tela ativa** (estado interno do RadKeys). Para organizar as centenas de frases, há **hierarquia de telas** (`navigate` entra em sub-tela, `level_up`/`go_home` retornam).

### Plataformas e mecanismos de captura

| Plataforma | Mecanismo | Backend | Requisitos do usuário |
|------------|-----------|---------|----------------------|
| **Windows** | `github.com/sstallion/go-hid` (hidapi) | WinAPI HID | Nenhum. Só baixar e executar. |
| **macOS** | `github.com/sstallion/go-hid` (hidapi) | IOKit | Nenhum em runtime (o dispositivo não é teclado, não exige Accessibility). |
| **Linux** | `github.com/sstallion/go-hid` (hidapi) | hidraw (`/dev/hidraw*`) | Permissão de leitura no `/dev/hidraw*`: grupo apropriado ou regra udev por VID/PID. |

> Como o dispositivo é **HID custom (vendor-defined)** e **não envia teclas**, nada chega ao RIS — não há vazamento por construção. Não há `RegisterHotKey`, CGEventTap nem evdev de teclado.

---

## 2. Restrições e requisitos de negócio

| Item | Definição |
|------|-----------|
| **Open source** | Sim, MIT, repositório público no GitHub. |
| **Arquivos de configuração** | Abertos em plaintext (TOML). Compartilháveis livremente. |
| **Receita** | Venda de arquivos premium (curadoria/conveniência) + hardware opcional/afiliado/DIY. |
| **Plataformas** | Windows 10/11, macOS (Intel + Apple Silicon), Linux. |
| **Distribuição** | Executável único, sem instalação, sem bundle, sem webview. |
| **Configuração** | Um arquivo `radkeys.config.toml` no mesmo diretório do executável. |
| **Hardware USB** | HID custom: Stream Deck/clone (protocolo Elgato) **ou** DIY de 24 (firmware RadKeys). |
| **Sem assinatura** | Modelo de pagamento perpétuo/único. |
| **Botões** | N botões físicos (ilimitado conforme o hardware); 3 fixos (copy/level_up/go_home) + (N−3) configuráveis por tela. Alvo 24. |

---

## 3. Stack tecnológico

| Camada | Tecnologia | Justificativa |
|--------|-----------|---------------|
| Linguagem | **Go 1.22+** | Compilação nativa, binário único, cross-compile maduro. |
| GUI | **Fyne v2.7.4** (estável) | Toolkit Go nativo, executável único, sem webview, clipboard (`App.Clipboard().SetContent()`). **Always-on-top NÃO disponível em 2.7.4** — o PR #6184 foi mergeado em `develop` (linha v2.8.0, ainda rc1), não em 2.7.x. MVP roda em janela normal; re-adicionar `desktop.Window.RequestAlwaysOnTop()` quando v2.8.0 estável sair. |
| Leitura do dispositivo | **`github.com/sstallion/go-hid` v0.15.0** (HIDAPI) | Wrapper idiomático do [libusb/hidapi](https://github.com/libusb/hidapi), cross-platform. Lê relatórios HID de dispositivos vendor-defined (Elgato e DIY). **Exige CGO + GCC** e pré-requisitos do HIDAPI. |
| Parser config | **TOML** (`github.com/BurntSushi/toml`) | Fácil para humanos e LLMs; estável. |
| Build multiplataforma | `fyne-cross` (Docker) ou builds nativos por OS | Agora com CGO + hidapi por OS (mais complexo — ver seção 10). |

### Notas críticas de API (validadas contra o código-fonte/docs)

- **`sstallion/go-hid`**: `hid.Init()` → `hid.OpenFirst(vid, pid)` (ou `hid.Enumerate`) → `d.ReadWithTimeout(buf, 50ms)` (retorna `hid.ErrTimeout` quando não há evento) → `d.Close()` / `hid.Exit()`. Linux: backend **hidraw** (default; linka `-ludev -lrt`), backend libusb opcional (`-tags libusb`). macOS: `-framework IOKit -framework CoreFoundation`. **Pré-requisitos do HIDAPI** devem estar instalados antes de `go get` ([hidapi BUILD.md](https://github.com/libusb/hidapi/blob/master/BUILD.md#prerequisites)); GCC no PATH.
- **Protocolo Elgato (Expanded family)**: input report `Report ID 0x01, Command 0x00`, payload = **1 byte por botão** (`0x00` solto, `0x01` pressionado). Host faz *HID READ* com timeout (poll recomendado **50ms**); `TIMEOUT` = sem evento. `Get Unit Information` (feature report `0x08`) retorna `rows`/`cols` → **nº de botões = rows×cols** (descoberta dinâmica). O RadKeys só implementa a **leitura de botões** (não escreve imagens; a UI vive no PC, não no display do Stream Deck).
- **Always-on-top (`desktop.Window.RequestAlwaysOnTop()`)**: **NÃO existe no Fyne v2.7.4 lançado.** O PR #6184 foi mergeado em `develop` (linha v2.8.0, ainda rc1 em 2026-07-07), não em `release/v2.7.x`. Decisão (Galvani): MVP fica em **v2.7.4 estável sem always-on-top**; re-adicionar (type assertion `w.(desktop.Window)` + chamar antes de `Show()`) ao subir para v2.8.0 estável. Detalhe e fontes em [research/fyne-always-on-top.md](research/fyne-always-on-top.md).
- **Clipboard**: `fyne.App.Clipboard()` (desde 2.6) com `SetContent(string)`; em ≥2.7.4: `a.Clipboard().SetContent(texto)`.

### O que NÃO usar

- **Teclado HID (F13–F24, modificadores):** limita a ~12 teclas inertes; modificadores vazam para o RIS. Rejeitado pelo produto.
- **Tauri/Electron:** bundles/webview. **Flutter:** overkill. **Python+PyInstaller:** pesado, menos portátil.

---

## 4. Arquitetura do sistema

### 4.1 Componentes

```
┌─────────────────────────────────────┐
│           RadKeys (Fyne UI)         │
│  ┌─────────────┐  ┌──────────────┐  │
│  │ Tela de uso │  │ Tela de edição│  │
│  │ (N botões + │  │  (config TOML)│  │
│  0 preview)    │  │               │  │
│  └─────────────┘  └──────────────┘  │
└─────────────────────────────────────┘
           │              │
           ▼              ▼
   ┌──────────────┐  ┌──────────────┐
   │ Config Model │  │ Clipboard API│
   │ (TOML)       │  │ (Fyne ≥2.6)  │
   └──────────────┘  └──────────────┘
           │
           ▼
   ┌─────────────────────────────────┐
   │ Estado de navegação            │
   │  currentScreen                 │
   └─────────────────────────────────┘
           │
           ▼
   ┌─────────────────────────────────┐
   │ HID Reader (sstallion/go-hid)   │
   │  • Protocolo Elgato (input 0x01)│
   │  • Protocolo DIY RadKeys        │
   │  Poll 50ms, ReadWithTimeout     │
   └─────────────────────────────────┘
           ▲
           │
   ┌──────────────┐
   │ USB HID custom│
   │ (vendor-     │
   │  defined)    │
   │ Stream Deck /│
   │ DIY 24       │
   └──────────────┘
```

### 4.2 Fluxo de dados

1. App inicia, carrega `radkeys.config.toml` e abre o dispositivo por VID/PID (`hid.OpenFirst` ou enumera).
2. Descobre N (Elgato: feature `0x08` rows×cols; DIY: 24).
3. Tela de uso (MVP: janela normal; always-on-top pendente do upgrade p/ Fyne v2.8.0 estável), mostrando os N botões da tela raiz (3 fixos + configuráveis).
4. Loop de poll: `d.ReadWithTimeout(buf, 50ms)`; em `ErrTimeout`, continua; senão parseia o input report e identifica o(s) botão(ões) que mudaram de estado.
5. Botão fixo → ação global (`copy`/`level_up`/`go_home`). Botão configurável → ação da tela ativa (`navigate` muda `currentScreen`; `text` carrega frase no preview).
6. `copy` → `a.Clipboard().SetContent(previewText)`.
7. Usuário cola no RIS (Ctrl+V). O foco **nunca saiu do RIS** — o hardware não envia teclas.

---

## 5. Formato do arquivo de configuração

### 5.1 Nome e local
- Nome: `radkeys.config.toml`, no mesmo diretório do executável. Se inexistente, app cria template mínimo.

### 5.2 Estrutura TOML

```toml
[app]
name = "RadKeys"
version = "1.2"

[app.device]
# Dispositivo HID custom.VID/PID.
# Elgato Stream Deck: vendor_id = 0x0fd9 (product_id por modelo: original/Mini/XL/Plus).
# DIY RadKeys: VID/PID próprios definidos no firmware.
# O RadKeys também pode enumerar por vendor_id e usar o primeiro compatível.
vendor_id  = 0x0fd9
product_id = 0x0063
protocol   = "elgato"   # "elgato" ou "radkeys-diy"

[app.fixed_buttons]
# Índices (0-based) dos 3 botões fixos globais.
copy     = 0   # copia o preview para o clipboard
level_up = 1   # sobe um nível na hierarquia
go_home  = 2   # volta à tela raiz

[[screens]]
id = "root"
title = "Início"
buttons = [
  { index = 3, label = "RX", action = "navigate", target = "rx_menu" },
  { index = 4, label = "TC", action = "navigate", target = "ct_menu" },
  { index = 5, label = "RM", action = "navigate", target = "mr_menu" },
  # ... até N-1
]

[[screens]]
id = "rx_torax"
title = "RX Tórax"
buttons = [
  { index = 3, label = "Normal", action = "text", content = """
Radiografia de tórax em incidências PA e perfil, realizadas em aparelho digital.
Arcada costal intacta, campos pulmonares livres, seios costofrênicos agudos.
Não há evidência de derrame pleural, pneumotórax ou consolidação.
""" },
  { index = 4, label = "Derrame direito", action = "text", content = """
Radiografia de tórax demonstrando opacidade basal direita com obliteração do seio costofrênico,
sugestiva de derrame pleural. Acompanhamento/ultrassonografia de tórax recomendado.
""" },
  # ... até N-1 (N−3 configuráveis)
]
```

### 5.3 Regras do formato
- `[app.device]`: VID/PID + `protocol` (`elgato` ou `radkeys-diy`).
- `[app.fixed_buttons]`: índices dos 3 botões fixos. Função constante em qualquer tela.
- Cada `[[screens]]` tem `id`, `title`, `buttons`. Botão: `index` (0..N−1, exceto os fixos), `label`, `action`.
- `action`: `navigate` (com `target`) ou `text` (com `content`).
- Lateralidade resolvida no conteúdo (botões separados: direito/esquerdo/bilateral).
- **Hierarquia:** uma categoria com mais frases que (N−3) usa sub-telas ligadas por `navigate`; `level_up`/`go_home` retornam.

---

## 6. Interface do usuário

### 6.1 Tela de uso
- Always-on-top.
- **Centro:** preview de uma frase por vez.
- **Grade dos N botões** (3 fixos + N−3 configuráveis) — **só os botões acionáveis pelo hardware atual**, sem botões órfãos. Layout reflete o keypad físico (rows×cols).
- Os 3 fixos em posição fixa, rotulados (Copy / Up / Home).
- Alto contraste, legível em monitor de radiologia.
- **Sem widgets editáveis focáveis** na tela de uso.

### 6.2 Navegação
- `copy`: copia o preview para o clipboard.
- `level_up`: sobe um nível.
- `go_home`: volta à raiz.
- `navigate`: entra em sub-tela.
- `text`: carrega texto no preview.

### 6.3 Tela de edição
- Lista telas/botões; adicionar/editar/remover (respeitando índices 0..N−1 e os 3 fixos).
- Editar conteúdo multi-line; salvar em TOML; recarrega a config ativa.
- **No modo edição a janela RadKeys tem foco** e aceita teclado comum. Roubar o foco é aceitável aqui.

---

## 7. Comportamento de foco

### 7.1 Promessa real
- **Modo USO:** sem roubar o foco do RIS. O dispositivo é HID custom e **não envia teclas** — nada chega ao sistema/RIS. A leitura é via hidapi (sem hook de teclado, sem RegisterHotKey, sem evdev de teclado).
- **Modo EDIÇÃO:** a janela RadKeys tem foco e usa teclado comum (aceitável).
- **Mouse:** clicar no RadKeys muda o foco (comportamento normal).

### 7.2 Implementação
1. Sem `RequestFocus()` na tela de uso.
2. Always-on-top: **indisponível no Fyne v2.7.4** (PR #6184 só em `develop`/v2.8.0). MVP roda em janela normal (o usuário posiciona/fixa via WM). Snippet preparado para v2.8.0 estável:
   ```go
   // ao subir para Fyne v2.8.0:
   if dw, ok := w.(desktop.Window); ok {
       dw.RequestAlwaysOnTop() // antes de w.Show()
   }
   ```
3. Abrir o dispositivo HID custom por VID/PID (`hid.OpenFirst` ou `hid.Enumerate`).
4. Loop de poll `d.ReadWithTimeout(buf, 50ms)`; parsear input report conforme `protocol`:
   - **Elgato:** `buf[0]=0x01` (report ID), `buf[1]=0x00` (cmd), payload a partir do offset 4 — 1 byte/botão.
   - **DIY RadKeys:** formato definido pelo firmware (ex.: bitmask de N bytes, 1 bit/botão).
5. Detectar transições 0→1 (press); mapear índice → ação (fixa ou da tela ativa).
6. `copy` → `a.Clipboard().SetContent(previewText)`.

---

## 8. Hardware USB

### 8.1 Opções (MVP)
1. **Comprar pronto — Stream Deck/clone (protocolo Elgato):** original Elgato (caro) ou clone compatível com o protocolo Elgato (médio custo). Pronto, sem construir. 15/32 teclas.
2. **DIY HID custom de 24 (firmware RadKeys):** alguém monta (Arduino/ESP32 como HID custom, vendor-defined). ~R$50–100 em peças. 24 teclas. Firmware open source no repo.

> Rejeitado: macro-keypad que envia teclas (F13–F24/letras/combos) — limita a ~12 inertes e/ou vaza modificadores para o RIS.

### 8.2 Protocolos
- **Elgato (Expanded family):** input report `0x01/0x00`, 1 byte/botão, poll 50ms. `Get Unit Information` (feature `0x08`) dá rows×cols. Referência: [docs Elgato](https://docs.elgato.com/streamdeck/hid/general), [.NET Haukcode.StreamDeck](https://github.com/HakanL/Haukcode.StreamDeck).
- **DIY RadKeys:** protocolo vendor-defined simples (input report = bitmask ou 1 byte/botão), firmware open source no repo. O RadKeys define o descritor HID e o formato.

### 8.3 Permissões Linux
- O dispositivo HID custom aparece em `/dev/hidraw*` (não `/dev/input/event*`).
- Permissão de leitura: grupo apropriado ou regra udev por VID/PID. Funciona em X11, Wayland e console (leitura via hidraw, independe do servidor gráfico).

---

## 9. Modelo de negócio e distribuição

### 9.1 Gratuito
- Código fonte (MIT) + firmware DIY (MIT).
- Arquivos de configuração básicos da comunidade.

### 9.2 Pago
- **Arquivos premium:** pacotes de frases por modalidade/especialidade (RX, TC, RM, US, MG).
- **Hardware:** versão pronta/afiliada (não é core).

### 9.3 Licenciamento
- MIT permite uso comercial e venda de conteúdo. Arquivos premium são plaintext abertos; o valor é a curadoria.

---

## 10. Build e deploy

### 10.1 Dependências de build
- Go 1.22+ e **GCC no PATH** (CGO obrigatório por `go-hid`).
- **Pré-requisitos do HIDAPI** por plataforma ([hidapi BUILD.md](https://github.com/libusb/hidapi/blob/master/BUILD.md#prerequisites)):
  - **Linux:** `libudev-dev` (backend hidraw). (libusb-1.0 só se usar `-tags libusb`.)
  - **macOS:** IOKit/CoreFoundation (system, sem install).
  - **Windows:** toolchain MinGW + HIDAPI.
- Fyne: GCC/MinGW (Windows), macOS SDK, Linux `libgl1-mesa-dev` + `xorg-dev`.
- `fyne-cross`: as imagens precisam incluir hidapi/udev/libGL — pode exigir imagem custom.

### 10.2 Targets

| Plataforma | Target | Artefato |
|------------|--------|----------|
| Windows x64 | `windows/amd64` | `radkeys.exe` |
| macOS Intel | `darwin/amd64` | `radkeys` |
| macOS Apple Silicon | `darwin/arm64` | `radkeys` |
| Linux x64 | `linux/amd64` | `radkeys` |

### 10.3 CI/CD sugerido
- GitHub Actions, 4 binários por release. Imagens Docker com CGO + hidapi por OS. Assets na release do GitHub. Sem app store.

### 10.4 Ação imediata
- `go.mod`: bump `fyne.io/fyne/v2` v2.5.3 → **v2.7.4** (estável; para `App.Clipboard()`); adicionar `github.com/sstallion/go-hid v0.15.0`; `go mod tidy`. Always-on-top **pendente** do upgrade p/ v2.8.0 estável. Garantir GCC + deps hidapi antes de buildar.

---

## 11. Riscos técnicos

| Risco | Prob. | Impacto | Mitigação |
|-------|-------|---------|-----------|
| CGO + pré-requisitos hidapi complicam o build/cross-compile | Alta | Médio | Imagens Docker dedicadas por OS em `fyne-cross`; documentar deps; build nativo como fallback. |
| Clone "stream deck" do AliExpress **não fala o protocolo Elgato** (cada fabricante chinês tem o seu) | Alta | Alto | Priorizar DIY 24 (protocolo controlado) e Stream Deck **original**; testar clones antes de prometer suporte; permitir VID/PID configurável para protocolos Elgato-compatíveis. |
| `/dev/hidraw*` no Linux sem permissão | Alta | Baixo | Documentar grupo/udev por VID/PID; fornecer regra udev pronta no repo. |
| Always-on-top indisponível no Fyne v2.7.4 (API só em v2.8+, ainda rc1) | Alta | Médio | MVP em janela normal; re-adicionar `RequestAlwaysOnTop()` ao subir p/ v2.8.0 estável. No Linux, `GLFW_FLOATING` é WM-dependente (testar GNOME/KDE/Wayland). |
| `go.mod` em v2.5.3 não tem `App.Clipboard()` | Alta | Médio | Bump para v2.7.4 (feito). |
| Descoberta dinâmica de nº de botões (Elgato feature `0x08`) varia entre firmwares | Média | Médio | Documentar; fallback: nº de botões configurável no TOML. |
| Binário grande (20–40 MB) por CGO/hidapi | Alta | Baixo | Aceitável para "executável único". |

---

## 12. Próximos passos

1. Bump `go.mod` (Fyne v2.7.4 + `sstallion/go-hid`); instalar GCC + deps hidapi; `go mod tidy`.
2. Implementar parser TOML (`[app.device]`, `[app.fixed_buttons]`, `[[screens]]` por `index`).
3. Implementar HID Reader: `hid.OpenFirst(VID,PID)` + loop `ReadWithTimeout(50ms)`; parser Elgato (input `0x01/0x00`) e parser DIY.
4. Implementar estado de navegação (`currentScreen`) e tela de uso (N botões + preview + 3 fixos).
5. Implementar cópia via `App.Clipboard().SetContent()`.
6. Always-on-top: **pendente** do upgrade p/ Fyne v2.8.0 estável (`desktop.Window.RequestAlwaysOnTop()` antes de `Show()`); MVP roda em janela normal.
7. Tela de edição (foco aceito neste modo).
8. Firmware DIY 24 (Arduino/ESP32 HID custom, vendor-defined) no repo.
9. Build multiplataforma (CGO + hidapi).
10. Testar com Stream Deck original e com DIY 24 nas 3 plataformas.

---

## 13. Fontes-chave (validadas)

- `sstallion/go-hid` (HIDAPI bindings): https://pkg.go.dev/github.com/sstallion/go-hid · https://github.com/sstallion/go-hid
- libusb/hidapi (+ BUILD prerequisites): https://github.com/libusb/hidapi · https://github.com/libusb/hidapi/blob/master/BUILD.md#prerequisites
- Protocolo Elgato Stream Deck (input reports, feature 0x08): https://docs.elgato.com/streamdeck/hid/general · intro: https://docs.elgato.com/streamdeck/hid/intro
- Ref. .NET Elgato (Haukcode.StreamDeck): https://github.com/HakanL/Haukcode.StreamDeck
- Fyne v2.7.4: https://pkg.go.dev/fyne.io/fyne/v2 · Releases: https://github.com/fyne-io/fyne/releases
- Fyne PR #6184 (`RequestAlwaysOnTop`) — mergeado em `develop` (v2.8.0), **não** em 2.7.x: https://github.com/fyne-io/fyne/pull/6184 · análise: [research/fyne-always-on-top.md](research/fyne-always-on-top.md)
- Fyne `Clipboard` (SetContent/Content): https://raw.githubusercontent.com/fyne-io/fyne/v2.7.4/clipboard.go
- `fyne-cross`: https://github.com/fyne-io/fyne-cross