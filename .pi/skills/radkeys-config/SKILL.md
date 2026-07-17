---
name: radkeys-config
description: Creates valid radkeys.config.toml files for the RadKeys radiology shortcut deck. Use when writing, editing, or validating TOML configs for RadKeys — including screens, buttons, actions, navigation, report text content, and exec commands. Covers all 13 actions, grid layout (1-6 rows/cols), themes, languages, and validation rules.
---

# RadKeys Config

RadKeys is a radiology shortcut deck: a 6×6 (36-button) USB keypad that copies pre-written
report phrases to the clipboard and pastes them into the RIS/PACS without stealing focus.
The config file (`radkeys.config.toml`) defines the radiologist name, language, theme, USB
device, and the screen hierarchy with shortcut buttons.

## How Configs Are Used

- The Go+Fyne host binary reads `radkeys.config.toml` at startup.
- The binary looks for the config at: `$RADKEYS_CONFIG` env var → `<exe-dir>/radkeys.config.toml` → `./radkeys.config.toml`.
- The optional `radkeys-config` binary is a visual TOML editor for the same file.
- Navigation is stack-based: `navigate` pushes current screen; `prev` pops; `home` clears to root.

**IMPORTANT:** When RadKeys saves the config, it first copies to `radkeys.config.toml.bak` then rewrites.
The Go TOML encoder (BurntSushi/toml) does NOT preserve comments. The `.bak` file keeps them.
Warn the user about this.

## TOML Structure

```toml
[app]
name = "RadKeys"
radiologist = "Dr. Name"
language = "en"

[app.device]
vendor_id = 0x1234
product_id = 0xABCD
protocol = "radkeys-diy"

[app.layout]
columns = 6   # 1–6
rows = 6      # 1–6

[app.theme]
preset = "dracula"
# icon = "/path/to/icon.png"   # optional custom icon

[[screens]]
id = "root"
name = "Home"

  [[screens.buttons]]
  row = 0
  col = 0
  label = "Chest"
  action = "navigate"
  target = "chest_menu"
```

## [app] Fields

| Field | Required | Default | Notes |
|-------|----------|---------|-------|
| `name` | No | — | App display name. |
| `radiologist` | No | — | Shown in the UI. |
| `language` | No | `"en"` | One of: `en`, `pt-BR`, `pt-PT`, `es`, `fr`, `de`, `it`. |
| `theme.preset` | No | `"system"` | One of 13 presets (see below). |
| `theme.icon` | No | — | Optional path to a custom icon file. |
| `layout.columns` | No | `6` | Grid columns, range 1–6. |
| `layout.rows` | No | `6` | Grid rows, range 1–6. |
| `device.vendor_id` | Yes | — | USB vendor ID (hex, e.g. `0x1234`). |
| `device.product_id` | Yes | — | USB product ID (hex, e.g. `0xABCD`). |
| `device.protocol` | Yes | — | Must be `"radkeys-diy"` (only supported protocol). |

### Theme Presets (13 total)

`system`, `dracula`, `solarized_dark`, `monokai`, `gruvbox_dark`, `nord`,
`one_dark`, `tokyo_night`, `catppuccin_mocha`, `solarized_light`, `gruvbox_light`,
`light_gray`, `dark_gray`.

## [[screens]] Array

Each `[[screens]]` section defines one page. The first screen in the file is the **home screen**.

| Field | Required | Notes |
|-------|----------|-------|
| `id` | **Yes** | Unique string. Must be non-empty, unique across all screens. |
| `name` | **Yes** | Display name shown in the UI header. Must be non-empty. |
| `buttons` | No | Array of `[[screens.buttons]]` for this screen. |

- Screen IDs must be unique across the entire config.
- The first screen is always the root/home screen.
- At least one screen is required.

## [[screens.buttons]] Array

Each button occupies a physical `(row, col)` on the keypad grid.

| Field | Required | Notes |
|-------|----------|-------|
| `row` | **Yes** | 0-based row index (0 to rows-1). |
| `col` | **Yes** | 0-based column index (0 to cols-1). |
| `label` | **Yes** | Display text on the button. Must be non-empty. |
| `action` | **Yes** | One of 13 actions (see below). |
| `target` | Conditional | **Required** when `action = "navigate"`. **Forbidden** otherwise. Must be an existing screen `id`. |
| `content` | Conditional | **Required** when `action = "text"` or `action = "exec"`. **Forbidden** otherwise. Multi-line strings use `"""..."""`. |

### Position Rules

- Row must be `0 ≤ row < rows` (default grid is 6×6, so rows 0–5, cols 0–5).
- Column must be `0 ≤ col < columns`.
- No two buttons on the same screen may share the same `(row, col)`.

## All 13 Actions

### Report-Content Actions

| Action | Requires | Description |
|--------|----------|-------------|
| `text` | `content` | Loads the content into the preview pane. |
| `exec` | `content` | Runs a bash command from content (not a report text). |

### Clipboard Actions

| Action | Requires | Description |
|--------|----------|-------------|
| `copy` | — | Copies the preview text to the clipboard. |
| `paste` | — | Sends Ctrl/Cmd+V to the focused window (the RIS) via the device keyboard. RadKeys never steals focus. |

### Navigation Actions

| Action | Requires | Description |
|--------|----------|-------------|
| `prev` | — | Goes back to the previous screen (stack-based). |
| `home` | — | Goes to the first screen in the config (clears stack). |
| `navigate` | `target` | Goes to the screen with the given `id` (pushes current screen onto stack). |

### Editing Keystroke Actions (sent via device keyboard)

| Action | Requires | Description |
|--------|----------|-------------|
| `select_all` | — | Sends Ctrl/Cmd+A (select all). |
| `select_line` | — | Selects the current line (Home then Shift+End). |
| `line_start` | — | Sends Home key (jump to line start). |
| `line_end` | — | Sends End key (jump to line end). |
| `backspace` | — | Sends Backspace (delete backward). |
| `delete` | — | Sends Delete Forward. |

### Action Constraints Summary

- `navigate` → `target` is **mandatory**, `content` is **forbidden**.
- `text` and `exec` → `content` is **mandatory**, `target` is **forbidden**.
- All other actions (copy, paste, prev, home, select_all, select_line, line_start, line_end, backspace, delete) → `target` and `content` are **forbidden**.

## Validation — All Failure Conditions

These are checked at load time. If any fail, the config is rejected with a specific error.

| What | Error |
|------|-------|
| `protocol` is not `"radkeys-diy"` | `[app.device] protocol must be "radkeys-diy", got "<value>"` |
| Unknown language | `[app] language "<value>" is not supported (use one of: en, pt-BR, pt-PT, es, fr, de, it)` |
| Unknown theme preset | `[app.theme] preset "<value>" is unknown` |
| `columns` < 1 or > 6 | `[app.layout] columns=<N> out of range [1,6]` |
| `rows` < 1 or > 6 | `[app.layout] rows=<N> out of range [1,6]` |
| Zero screens | `no screens defined — add at least one [[screens]]` |
| Empty screen `id` | `screen has empty id` |
| Duplicate screen `id` | `duplicate screen id "<id>"` |
| Empty screen `name` | `screen "<id>" has empty name` |
| Empty button `label` | `screen "<id>", button at (row=N, col=N): label is required` |
| `row` out of grid bounds | `screen "<id>", button "<label>": row=N out of range [0,rows)` |
| `col` out of grid bounds | `screen "<id>", button "<label>": col=N out of range [0,cols)` |
| Two buttons same `(row,col)` | `screen "<id>": buttons "<A>" and "<B>" both occupy (row=N, col=N)` |
| Unknown action string | `screen "<id>", button "<label>": invalid action "<action>"` |
| `navigate` without `target` | `screen "<id>", button "<label>": navigate requires target` |
| Non-navigate action with `target` | `screen "<id>", button "<label>": action "<action>" does not accept target` |
| `text` or `exec` without `content` | `screen "<id>", button "<label>": text requires content` (or `exec requires content`) |
| Non-text/exec action with `content` | `screen "<id>", button "<label>": action "<action>" does not accept content` |
| `navigate` to non-existent `target` | `screen "<id>", button "<label>": target "<target>" does not exist` |

## Defaults

When the following fields are omitted, defaults are applied **before** validation:

| Field | Default |
|-------|---------|
| `app.language` | `"en"` |
| `app.theme.preset` | `"system"` |
| `app.layout.columns` | `6` |
| `app.layout.rows` | `6` |

## Best Practices for Radiology Config Design

1. **Home screen first.** The first `[[screens]]` in the file IS the home screen. Always make it a modality selector (X-Ray, CT, MRI, US, etc.).

2. **Row 4 + Row 5 on EVERY screen.** Row 4 (cols 0–3): SLine (`select_line`), LStart (`line_start`), LEnd (`line_end`), Del (`delete`). Row 5 (all 6 cols): Back (`prev`), Home (`home`), SelAll (`select_all`), Copy (`copy`), Paste (`paste`), Bksp (`backspace`). This gives the radiologist full editing control from any screen. exec buttons appear ONLY on the root-level utilities screen.

3. **Content buttons occupy the main grid.** Use rows 0–4 for report text content. Group related findings together.

4. **Use descriptive screen IDs.** Pattern: `<modality>_<region>[_<subtype>]`. Examples: `xray_chest`, `ct_abdomen_pelvis`, `mri_brain`, `us_thyroid`.

5. **Context-aware report templates.** Each text button should contain a complete, ready-to-paste radiology report fragment. Use `"""..."""` for multi-line content. Keep content concise (the device keyboard types it out, so shorter = faster).

6. **Nesting depth.** Keep screen hierarchy ≤ 6 levels deep. Deep nesting is fine for detailed anatomical subcategories (e.g., root → X-Ray → Extremities → Hand → Trauma → Distal Radius Fracture). Use `prev` to bubble back up — navigation is stack-based so the user can always go back.

7. **Workflow pattern.** Standard button layout for ALL screens:
   - Row 0–3: content buttons — report texts, navigate shortcuts (no exec outside utilities)
   - Row 4: SLine | LStart | LEnd | Del (all 4 editing keystrokes)
   - Row 5: Back | Home | SelAll | Copy | Paste | Bksp (all 6 positions)
   On menu/navigation screens, rows 0–3 contain `navigate` buttons. On report screens, rows 0–3 contain `text` buttons grouped by finding category.

8. **All 13 actions must be represented.** The config as a whole should exercise every action at least once. The `utilities` screen should demonstrate a full 6×6 grid with all 36 positions filled and all 13 actions present. See `references/example.toml` for the canonical layout.

9. **Placeholder screens are OK.** When building out the config, placeholder menus (with just Home/Back buttons) are valid. The user fills them in later.

10. **Exec buttons for integration.** Use `exec` sparingly — it runs arbitrary bash. Useful for launching external tools or scripts from the deck.

## Reference Example

See `references/example.toml` for a heavily-commented, complex, realistic radiology config
with **39 screens, up to 6 levels of navigation depth**, every action exercised, and all
39 screens have both a 4-button row-4 editing bar (SLine, LStart, LEnd, Del) and a
6-button row-5 bar (Back, Home, SelAll, Copy, Paste, Bksp).`exec` buttons appear ONLY
on the `utilities` screen (L2 root level) — all other screens use only navigation,
editing keystroke, text, and clipboard actions.
