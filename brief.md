# RadKeys — Brief Técnico Completo

> **Versão:** 1.0  
> **Data:** 2026-07-07  
> **Autor:** Nonatinho (consultor)  
> **Cliente:** Galvani (radiologista)  
> **Repo:** https://github.com/docg1701/radkeys

---

## 1. Visão do produto

RadKeys é um aplicativo desktop portátil, open source, multiplataforma, para médicos radiologistas que **digitam laudos**. Ele exibe uma interface visual de atalhos hierárquicos (shortcut deck) com preview central. Cada atalho carrega uma frase pronta de laudo. O usuário confirma a cópia, o texto vai para a área de transferência e é colado no RIS/PACS.

A operação principal é feita por **teclado físico/dispositivo USB HID** enviando teclas F13-F24, sem roubar o foco do RIS. Cliques de mouse na janela do RadKeys mudam o foco normalmente — esse é o comportamento esperado.

---

## 2. Restrições e requisitos de negócio

| Item | Definição |
|------|-----------|
| **Open source** | Sim, MIT, repositório público no GitHub. |
| **Arquivos de configuração** | Abertos em plaintext (TOML). Compartilháveis livremente. |
| **Receita** | Venda de arquivos premium (curadoria/conveniência) + hardware opcional/afiliado/DIY. |
| **Plataformas** | Windows 10/11, macOS (Intel + Apple Silicon), Linux (X11 principalmente). |
| **Distribuição** | Executável único, sem instalação, sem bundle, sem webview. |
| **Configuração** | Um arquivo `radkeys.config.toml` no mesmo diretório do executável. |
| **Hardware USB** | Upgrade opcional; dispositivo HID genérico enviando F13-F24. |
| **Sem assinatura** | Modelo de pagamento perpétuo/único. |

---

## 3. Stack tecnológico

| Camada | Tecnologia | Justificativa |
|--------|-----------|---------------|
| Linguagem | **Go 1.22+** | Compilação nativa, binário único, cross-compile maduro. |
| GUI | **Fyne v2 (≥2.7.4)** | Toolkit Go nativo, executável único, sem webview, API de clipboard, always-on-top via `RequestAlwaysOnTop()`. |
| Hotkeys globais | **golang.design/x/hotkey** | Registra hotkeys no sistema operacional (Windows/macOS/Linux X11) sem exigir foco na janela. Integra com Fyne. |
| Parser config | **TOML** (`github.com/BurntSushi/toml` ou similar) | Fácil para humanos e LLMs. |
| Build multiplataforma | `fyne-cross` (Docker) ou builds nativos por OS. | Gera binários para Windows, macOS e Linux a partir de uma codebase. |

### O que NÃO usar

- **Tauri/Electron:** geram bundles/instaladores e dependem de webview.
- **Flutter:** overkill para um MVP desktop.
- **Python+PyInstaller:** binário pesado, instalação implícita, menos portátil.

---

## 4. Arquitetura do sistema

### 4.1 Componentes

```
┌─────────────────────────────────────┐
│           RadKeys (Fyne UI)         │
│  ┌─────────────┐  ┌──────────────┐  │
│  │ Tela de uso │  │ Tela de edição│  │
│  │ (shortcut   │  │  (config TOML)│  │
│  │  deck +     │  │               │  │
│  │  preview)   │  │               │  │
│  └─────────────┘  └──────────────┘  │
└─────────────────────────────────────┘
           │              │
           ▼              ▼
   ┌──────────────┐  ┌──────────────┐
   │ Config Model │  │ Clipboard API│
   │ (TOML)       │  │ (Fyne)       │
   └──────────────┘  └──────────────┘
           │
           ▼
   ┌──────────────┐
   │ Global     │
   │ Hotkeys    │
   │ (golang.   │
   │ design/x/  │
   │ hotkey)    │
   └──────────────┘
           ▲
           │
   ┌──────────────┐
   │ USB HID      │
   │ keyboard     │
   │ (F13-F24)    │
   └──────────────┘
```

### 4.2 Fluxo de dados

1. App inicia e carrega `radkeys.config.toml` do mesmo diretório.
2. Tela de uso é exibida always-on-top.
3. Hotkeys globais F13-F24 são registradas.
4. Usuário posiciona cursor no campo de laudo do RIS.
5. Usuário aperta tecla do dispositivo USB → hotkey global dispara → navegação na hierarquia.
6. Usuário seleciona frase → preview atualiza.
7. Usuário confirma cópia → texto vai para clipboard.
8. Usuário cola no RIS (Ctrl+V ou tecla física "colar").

---

## 5. Formato do arquivo de configuração

### 5.1 Nome e local

- Nome: `radkeys.config.toml`
- Local: mesmo diretório do executável.
- Se não existir: app cria template mínimo.

### 5.2 Estrutura TOML

```toml
[app]
name = "RadKeys"
version = "1.0"

[app.hotkeys]
level_up = "Escape"
go_home = "Home"
copy = "Enter"

[app.global_hotkeys]
# Mapeia teclas F13-F24 para ações/navegação
F13 = { action = "navigate", target = "rx_menu" }
F14 = { action = "navigate", target = "ct_menu" }
F15 = { action = "navigate", target = "mr_menu" }
F16 = { action = "level_up" }
F17 = { action = "go_home" }
F18 = { action = "copy" }

[[screens]]
id = "root"
title = "Início"
buttons = [
  { key = "a", label = "RX", action = "navigate", target = "rx_menu" },
  { key = "b", label = "TC", action = "navigate", target = "ct_menu" },
  { key = "c", label = "RM", action = "navigate", target = "mr_menu" },
]

[[screens]]
id = "rx_menu"
title = "RX"
buttons = [
  { key = "a", label = "Tórax", action = "navigate", target = "rx_torax" },
  { key = "b", label = "Abdome", action = "navigate", target = "rx_abdome" },
  { key = "home", label = "Início", action = "navigate", target = "root" },
  { key = "esc", label = "Voltar", action = "level_up" },
]

[[screens]]
id = "rx_torax"
title = "RX Tórax"
buttons = [
  { key = "a", label = "Normal", action = "text", content = """
Radiografia de tórax em incidências PA e perfil, realizadas em aparelho digital.
Arcada costal intacta, campos pulmonares livres, seios costofrênicos agudos.
Não há evidência de derrame pleural, pneumotórax ou consolidação.
""" },
  { key = "b", label = "Derrame direito", action = "text", content = """
Radiografia de tórax demonstrando opacidade basal direita com obliteração do seio costofrênico,
sugestiva de derrame pleural. Acompanhamento/ultrassonografia de tórax recomendado.
""" },
  { key = "c", label = "Derrame esquerdo", action = "text", content = """
Radiografia de tórax demonstrando opacidade basal esquerda com obliteração do seio costofrênico,
sugestiva de derrame pleural. Acompanhamento/ultrassonografia de tórax recomendado.
""" },
  { key = "enter", label = "Copiar", action = "copy" },
  { key = "home", label = "Início", action = "navigate", target = "root" },
  { key = "esc", label = "Voltar", action = "level_up" },
]
```

### 5.3 Regras do formato

- Cada tela tem `id`, `title` e `buttons`.
- Botão tem: `key`, `label`, `action`.
- `action`: `navigate`, `text`, `level_up`, `copy`.
- `target`: próxima tela (quando `navigate`).
- `content`: texto a copiar (quando `text`).
- Lateralidade resolvida no conteúdo (botões separados: direito/esquerdo/bilateral).
- Hotkeys globais mapeiam F13-F24 para ações do mesmo tipo dos botões.

---

## 6. Interface do usuário

### 6.1 Tela de uso

Layout em tela cheia/janela grande, always-on-top:

- **Centro:** preview de uma frase por vez.
- **Ao redor:** botões de atalho mostrando tecla + label.
- **Cores/ícones:** simples, alto contraste, legível em monitor de radiologia.
- **Sem widgets editáveis focáveis na tela principal** — para minimizar mudança de foco acidental.

### 6.2 Navegação

- `level_up`: sobe um nível na hierarquia.
- `go_home`: volta à raiz.
- `navigate`: entra em uma sub-tela.
- `text`: carrega texto no preview.
- `copy`: copia o preview para a área de transferência.

### 6.3 Tela de edição

- Lista todas as telas.
- Adicionar/editar/remover telas e botões.
- Editar conteúdo de texto (multi-line).
- Salvar em TOML.
- Ao salvar, recarregar configuração ativa automaticamente.

---

## 7. Comportamento de foco

### 7.1 Promessa real

- **Dispositivo USB/teclado:** não rouba o foco do RIS. O RadKeys é notificado via hotkeys globais do sistema operacional.
- **Mouse:** clicar na janela do RadKeys muda o foco para o RadKeys. Isso é comportamento normal e documentado.

### 7.2 Implementação

1. Não chamar `RequestFocus()`.
2. Usar `desktop.Window.RequestAlwaysOnTop()` (Fyne ≥2.7.4).
3. Capturar F13-F24 via `golang.design/x/hotkey` registradas globalmente.
4. Quando hotkey dispara, executar ação correspondente (navegação, cópia, etc.).
5. Copiar para clipboard via `fyne.App.Clipboard().SetContent()`.

---

## 8. Hardware USB

### 8.1 Opções

1. **Comprar pronto:** clones de Stream Deck no AliExpress/envio direto.
2. **DIY:** Arduino Pro Micro/ESP32 + teclas mecânicas + firmware HID.
3. **Afiliado:** indicar fornecedor; usuário compra direto.

### 8.2 Protocolo

- Dispositivo se comporta como teclado USB HID.
- Cada tecla física envia F13-F24 (ou F13-F24 + modificadores).
- Não requer driver específico do RadKeys.
- O RadKeys apenas escuta as hotkeys globais correspondentes.

---

## 9. Modelo de negócio e distribuição

### 9.1 Gratuito

- Código fonte (MIT).
- Arquivos de configuração básicos da comunidade.

### 9.2 Pago

- **Arquivos premium:** pacotes de frases por modalidade/especialidade (RX, TC, RM, US, MG).
- **Hardware:** versão pronta/afiliada (não é core).

### 9.3 Licenciamento

- MIT permite uso comercial e venda de conteúdo.
- Arquivos premium são plaintext abertos; o valor é a curadoria.

---

## 10. Build e deploy

### 10.1 Dependências de build

- Go 1.22+
- GCC/MinGW (Windows)
- macOS SDK (Darwin)
- Linux: libgl1-mesa-dev, xorg-dev
- fyne-cross para builds cruzados

### 10.2 Targets

| Plataforma | Target triplet | Artefato |
|------------|---------------|----------|
| Windows x64 | `windows/amd64` | `radkeys.exe` |
| macOS Intel | `darwin/amd64` | `radkeys` |
| macOS Apple Silicon | `darwin/arm64` | `radkeys` |
| Linux x64 | `linux/amd64` | `radkeys` |

### 10.3 CI/CD sugerido

- GitHub Actions buildando os 4 binários a cada release.
- Assets anexados na release do GitHub.
- Sem loja/app store (evita bundles e revisões).

---

## 11. Riscos técnicos

| Risco | Probabilidade | Impacto | Mitigação |
|-------|--------------|---------|-----------|
| Fyne `RequestAlwaysOnTop()` não funciona em algum Linux | Média | Médio | Testar em distros comuns; fallback posicionamento manual. |
| `golang.design/x/hotkey` limitado em Wayland | Média | Médio | Documentar suporte X11; testar em distros desktop. |
| Binário grande (15-30 MB+) | Alta | Baixo | Aceitável para "executável único". |
| Cross-compile macOS exige macOS SDK | Alta | Médio | Usar fyne-cross; ou build em macOS real para release. |
| Kimi/GLM alucinam em APIs obscuras | Média | Médio | Fornecer referências (`golang.design/x/hotkey`, exemplo Fyne). |

---

## 12. Próximos passos

1. Implementar parser TOML.
2. Implementar tela de uso com navegação hierárquica e preview.
3. Integrar `golang.design/x/hotkey` para F13-F24.
4. Implementar cópia para clipboard.
5. Adicionar always-on-top.
6. Implementar tela de edição.
7. Build multiplataforma.
8. Testar com dispositivo USB HID real.

---

## 13. Fontes-chave

- Fyne: https://fyne.io/
- Fyne PR #6184 (always-on-top): https://github.com/fyne-io/fyne/pull/6184
- golang.design/x/hotkey: https://pkg.go.dev/golang.design/x/hotkey
- Exemplo Fyne + hotkey: https://github.com/golang-design/hotkey/blob/main/examples/fyne/main.go
- fyne-cross: https://github.com/fyne-io/fyne-cross
- DIY Stream Deck HID F13-F24: https://github.com/Mercawa/DIYStreamDeck-HIDKeyboard
