# RadKeys — Brief Técnico Completo

> **Versão:** 2.0
> **Data:** 2026-07-08
> **Autor:** Nonatinho (consultor)
> **Cliente:** Galvani (radiologista)
> **Repo:** https://github.com/docg1701/radkeys

---

## 1. Visão do produto

RadKeys é um aplicativo desktop portátil, open source, multiplataforma, para
médicos radiologistas que **digitam laudos**. Exibe uma interface com um
**preview de texto em cima** (metade da tela) e um **keypad virtual embaixo**
(metade da tela) que reflete o dispositivo USB de botões. Cada botão carrega
uma frase pronta de laudo; o usuário confirma a cópia, o texto vai para a
área de transferência e é colado no RIS/PACS.

A operação é feita por um **dispositivo USB HID custom**, lido diretamente
pelo RadKeys — o hardware **não envia teclas**, logo **não rouba o foco do RIS**.

### Modelo de botões (N botões físicos via hidapi)

| Botões | Função | Configurável? |
|--------|--------|---------------|
| **3 fixos** (índices 0/1/2 por padrão) | `copy`, `level_up`, `go_home` | Fixa (global) |
| **N−3 configuráveis** | `navigate` (sub-tela) ou `text` (frase) | Por tela |

- **N** = botões do hardware, descoberto em runtime. Alvo: **24** (DIY).
- **Hierarquia de telas**: `navigate` entra em sub-tela; `level_up`/`go_home` retornam.

---

## 2. Restrições e requisitos

| Item | Definição |
|------|-----------|
| Open source | MIT |
| Config | `radkeys.config.toml` (plaintext TOML) |
| Plataformas | Windows 10/11, macOS (Intel+AS), Linux |
| Distribuição | Executável único, sem instalação |
| Hardware | HID custom: Stream Deck/clone (Elgato) ou DIY 24 (Arduino) |
| Botões | 3 fixos + (N−3) configuráveis; alvo 24 |

---

## 3. Stack tecnológico

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.22+ |
| GUI | Fyne v2.7.4 (estável) |
| HID | `github.com/sstallion/go-hid` v0.15.0 (hidapi, CGO) |
| Config | `github.com/BurntSushi/toml` |
| Build | `fyne-cross` ou nativo por OS |

### Notas de API (validadas)

- **Always-on-top**: `desktop.Window.RequestAlwaysOnTop()` **não existe em v2.7.4**
  (PR #6184 mergeado em `develop`/v2.8.0, ainda rc1). MVP sem always-on-top;
  re-adicionar ao subir para v2.8.0 estável. Ver `research/fyne-always-on-top.md`.
- **Clipboard**: `a.Clipboard().SetContent(texto)` (disponível desde 2.6).
- **go-hid**: `hid.Init()` → `hid.OpenFirst(vid,pid)` → `d.ReadWithTimeout(buf, 50ms)`
  (retorna `hid.ErrTimeout` sem evento) → `d.Close()`.
- **Elgato**: input report `0x01/0x00`, 1 byte/botão. Feature `0x08` = rows×cols.
- **DIY**: vendor-defined, report ID 1 + 24 bytes (1 por botão).

---

## 4. Arquitetura

```
┌──────────────────────────────────────┐
│            RadKeys (Fyne UI)          │
│  ┌────────────┐  ┌─────────────────┐  │
│  │ Aba        │  │ Aba "Editar"    │  │
│  │ "Atalhos"  │  │ (editor visual) │  │
│  │            │  │                 │  │
│  │ ┌────────┐ │  │  Lista de telas │  │
│  │ │Preview │ │  │  + formulário   │  │
│  │ │(topo,  │ │  │  + botões       │  │
│  │ │ 50%)   │ │  │  + salvar       │  │
│  │ ├────────┤ │  │                 │  │
│  │ │Keypad  │ │  │                 │  │
│  │ │(baixo, │ │  │                 │  │
│  │ │ 50%)   │ │  │                 │  │
│  │ │ 4×N    │ │  │                 │  │
│  │ └────────┘ │  │                 │  │
│  └────────────┘  └─────────────────┘  │
└──────────────────────────────────────┘
        │              │
        ▼              ▼
  ┌──────────┐  ┌──────────────┐
  │ Deck     │  │ Clipboard   │
  │ (naveg.) │  │ (Fyne)       │
  └──────────┘  └──────────────┘
        │
        ▼
  ┌──────────────────────┐
  │ HID Reader (go-hid)  │
  │ Elgato / DIY         │
  └──────────────────────┘
        ▲
        │
  ┌──────────────┐
  │ USB HID custom│
  │ Stream Deck /│
  │ DIY 24       │
  └──────────────┘
```

### Fluxo de dados

1. App carrega `radkeys.config.toml` (cria template se não existe).
2. Abre o dispositivo por VID/PID (`hid.OpenFirst`).
3. Tela "Atalhos": preview no topo, keypad 4×N embaixo.
4. Loop de poll `ReadWithTimeout(50ms)`: parseia input report.
5. Botão fixo → `copy`/`level_up`/`go_home`. Botão configurável → `navigate`/`text`.
6. `copy` → `Clipboard().SetContent(previewText)`.
7. Usuário cola no RIS. Foco nunca saiu do RIS.
8. Tela "Editar": edita telas/botões, salva TOML, recarrega e aplica à tela de atalhos.

---

## 5. Formato do config

### 5.1 Arquivo

`radkeys.config.toml` no mesmo diretório do executável. Se não existe, app
cria template mínimo.

### 5.2 Estrutura TOML

```toml
[app]
name = "RadKeys"
version = "2.0"

[app.device]
vendor_id  = 0x0fd9
product_id = 0x0063
protocol   = "elgato"   # "elgato" ou "radkeys-diy"

[app.fixed_buttons]
copy     = 0
level_up = 1
go_home  = 2

[app.layout]
columns = 4    # colunas do keypad virtual (default 4)
rows    = 5    # linhas do keypad virtual (default 5)

[app.theme]
background = "#1a1a1a"  # fundo do preview
button     = "#2a2a2a"  # cor do slot vazio
fixed      = "#3a3a3a"  # cor do botão fixo (reservado)

[[screens]]
id = "root"
title = "Início"
buttons = [
  { index = 3, label = "RX Tórax", action = "navigate", target = "rx_torax" },
  { index = 4, label = "Normal", action = "text", content = "..." },
]

[[screens]]
id = "rx_torax"
title = "RX Tórax"
buttons = [
  { index = 3, label = "Normal", action = "text", content = "..." },
]
```

### 5.3 Regras

- `[app.device]`: VID/PID + protocolo.
- `[app.fixed_buttons]`: índices dos 3 botões fixos globais.
- `[app.layout]`: columns/rows do keypad virtual. Default 4×5.
- `[app.theme]`: cores em hex (configuráveis; defaults se vazio).
- `[[screens]]`: `id`, `title`, `buttons[]`.
- Botão: `index`, `label`, `action` (`navigate`+`target` ou `text`+`content`).

---

## 6. Interface do usuário

### 6.1 Tela "Atalhos"

- Janela **redimensionável** (1280×800 default, sem `SetFixedSize`).
- Tema **escuro** (DarkTheme).
- **Topo**: título da tela ativa (centro, negrito).
- **Metade superior**: **preview** — `widget.Label` monospace com scroll,
  preserva quebras de linha do laudo. Fundo com cor `[app.theme].background`.
- **Metade inferior**: **keypad virtual** — grid `columns × rows` (default 4×5 = 20).
  Cada slot é `NewGridWrap(120×80)` (tamanho fixo, todos iguais). Botões
  preenchidos = `widget.Button`; slots vazios = `canvas.Rectangle` com
  `[app.theme].button`. Os 3 fixos (Copiar/Voltar/Início) sempre ocupam os
  primeiros slots.
- Sem `SetFixedSize`; sem `RequestFocus()` na tela de uso.

### 6.2 Navegação

- `copy`: copia preview para clipboard.
- `level_up`: sobe um nível.
- `go_home`: volta à raiz.
- `navigate`: entra em sub-tela.
- `text`: carrega texto no preview.

### 6.3 Tela "Editar"

- **Abas**: "Atalhos" e "Editar" (`container.NewAppTabs`).
- **Editor**: lista de telas à esquerda (`widget.List`), formulário à direita.
  - Selecionar tela → mostra ID, Título, e lista de botões (cada botão num
    `widget.NewCard` com índice, rótulo, ação, target, conteúdo multiline).
  - "Nova tela": adiciona e **seleciona** a nova tela (atualiza o formulário).
  - "Novo botão": adiciona e **rebuild** da lista de botões.
  - "Remover": remove tela/botão e atualiza.
  - "Salvar e aplicar": escreve TOML, recarrega config, recria deck,
    **re-renderiza o keypad** da aba Atalhos.
- Modo edição aceita foco (janela RadKeys tem foco; aceitável).

---

## 7. Comportamento de foco

- **Modo uso (aba Atalhos)**: sem roubar foco do RIS. Dispositivo HID custom
  não envia teclas — nada chega ao sistema. Leitura via hidapi.
- **Modo edição (aba Editar)**: janela RadKeys tem foco, aceita teclado comum.
- **Mouse**: clicar no RadKeys muda o foco (normal).

---

## 8. Hardware USB

### 8.1 Opções

1. **Comprar pronto**: Stream Deck/clone (protocolo Elgato). 15/32 teclas.
2. **DIY (primário)**: Arduino Pro Micro (ATmega32U4) + chaves de teclado
   chinês barato + caixa 3D printed + cabo USB. Matriz 6×4 (10 pinos).
   ~R$30-50. Firmware: `firmware/arduino/`.
3. **DIY (alt)**: Raspberry Pi Pico (RP2040), 24 GPIO diretos. `firmware/rp2040/`.

### 8.2 Protocolos

- **Elgato**: input report `0x01/0x00`, 1 byte/botão. Feature `0x08` = rows×cols.
- **DIY**: vendor-defined, report ID 1 + 24 bytes. `parseDIYReport` aceita
  25 bytes (com report ID) ou 24 bytes (sem).

### 8.3 Permissões Linux

- `/dev/hidraw*` (não `/dev/input/event*`). Grupo `input` ou regra udev por VID/PID.
- Funciona em X11, Wayland e console.

---

## 9. Modelo de negócio

- Gratuito: código (MIT) + firmware (MIT) + configs da comunidade.
- Pago: arquivos premium (frases curadas por modalidade).
- Hardware: versão pronta/afiliada (não é core).

---

## 10. Build e deploy

### 10.1 Dependências

- Go 1.22+ e **GCC** (CGO por go-hid).
- Linux: `libgl1-mesa-dev xorg-dev libudev-dev libxxf86vm-dev`.
- macOS: IOKit (system). Windows: MinGW.

### 10.2 CI

- `.github/workflows/build.yml`: build/test/vet em ubuntu, macos, windows.

### 10.3 Targets

| Plataforma | Artefato |
|------------|----------|
| Windows x64 | `radkeys.exe` |
| macOS Intel | `radkeys` |
| macOS AS | `radkeys` |
| Linux x64 | `radkeys` |

---

## 11. Riscos

| Risco | Mitigação |
|-------|-----------|
| CGO + hidapi complica cross-compile | Imagens Docker por OS; build nativo fallback |
| Clone não fala Elgato | Priorizar DIY (protocolo controlado) |
| Always-on-top indisponível em v2.7.4 | MVP sem; re-adicionar em v2.8.0 estável |
| `/dev/hidraw*` sem permissão | Documentar udev/grupo |

---

## 12. Próximos passos

1. ~~Bump go.mod~~ (feito: Fyne v2.7.4 + go-hid v0.15.0)
2. ~~Parser TOML~~ (feito + testes)
3. ~~HID Reader~~ (feito: Elgato + DIY + mock)
4. ~~Tela de uso~~ (feito: preview + keypad 4×N)
5. ~~Clipboard~~ (feito)
6. Always-on-top: pendente v2.8.0 estável
7. ~~Tela de edição~~ (feito: editor visual com salvar+aplicar)
8. ~~Firmware DIY~~ (feito: Arduino Pro Micro + RP2040)
9. ~~CI~~ (feito: GitHub Actions)
10. Testar com hardware real
11. Polir UI (cores custom theme completo, animações)

---

## 13. Fontes-chave

- `sstallion/go-hid`: https://pkg.go.dev/github.com/sstallion/go-hid
- libusb/hidapi: https://github.com/libusb/hidapi
- Elgato HID: https://docs.elgato.com/streamdeck/hid/general
- Fyne v2.7.4: https://pkg.go.dev/fyne.io/fyne/v2
- PR #6184 (always-on-top, em develop): https://github.com/fyne-io/fyne/pull/6184
- Análise always-on-top: `research/fyne-always-on-top.md`
- DIY Stream Deck (ref): https://github.com/Mercawa/DIYStreamDeck-HIDKeyboard
- `fyne-cross`: https://github.com/fyne-io/fyne-cross

---

## 14. Estrutura do repo

```
radkeys/
├── brief.md                        # este brief (v2.0)
├── go.mod / go.sum                 # Fyne v2.7.4 + go-hid + BurntSushi/toml
├── main.go                         # entrypoint: load config, open HID, run UI
├── radkeys.config.toml             # config de exemplo
├── README.md
├── LICENSE                         # MIT
├── .github/workflows/build.yml     # CI cross-platform
├── internal/
│   ├── config/
│   │   ├── config.go               # parser TOML + validação + Theme/Layout
│   │   └── config_test.go           # testes (6 testes, roundtrip TOML)
│   ├── deck/
│   │   ├── deck.go                 # estado de navegação (navigate/text/copy/up/home)
│   │   └── deck_test.go            # testes (8 testes)
│   ├── hid/
│   │   ├── hid.go                   # interface Reader + MockReader
│   │   ├── reader_cgo.go           # implementação go-hid (Elgato + DIY)
│   │   ├── reader_nocgo.go         # fallback sem CGO
│   │   └── hid_test.go             # testes do mock (4 testes)
│   └── ui/
│       ├── ui.go                   # tela Atalhos (preview + keypad)
│       └── edit.go                  # tela Editar (editor visual)
├── firmware/
│   ├── arduino/
│   │   ├── diy24.ino                # Arduino Pro Micro (matriz 6×4, HID vendor-defined)
│   │   └── README.md               # BOM, pinagem, udev, protocolo
│   └── rp2040/
│       ├── diy24.ino                # RP2040 (24 GPIO diretos, Adafruit_TinyUSB)
│       └── README.md
└── research/
    └── fyne-always-on-top.md        # investigação do PR #6184
```

### Testes

```
go test ./...  →  config: ok (6)  deck: ok (8)  hid: ok (4)
go vet ./...   →  clean
go build .     →  radkeys (31 MB, CGO + Fyne + go-hid)
```

### Validação

- Compila em Linux amd64 (Go 1.22.2, GCC 13.3, libudev 255, libxxf86vm OK).
- Executável roda: sem hardware → mock → UI funcional via mouse.
- `go test ./...` passa (18 testes).
- `go vet ./...` limpo.