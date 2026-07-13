# Research: Visual Config-File Editors for Non-Technical (Lay) Users

## Summary

The dominant UX pattern across successful config editors for lay users is **drag-and-drop onto a visual canvas that mirrors the physical device**, combined with a **property inspector panel** for configuring the selected item. The gold standard is Elgato Stream Deck's software: a left-side visual grid of LCD buttons, a right-side action palette, and a bottom property panel — no raw JSON/TOML editing required. For schema-driven config files (TOML/YAML/JSON), the emerging best practice is **auto-generating forms from a typed schema** (Pydantic Studio, MetaConfigurator), which eliminates syntax errors entirely. The most critical UX patterns for a lay-user config editor are: (1) a visual representation of the device/buttons, (2) drag-and-drop assignment, (3) pre-built action library with categories, (4) inline validation with helpful error messages, (5) undo/redo, (6) dirty-state tracking with confirm-on-unsaved-quit, and (7) profiles/pages for organization.

---

## Findings

### 1. Macro/Keypad Config Apps

**1a. Elgato Stream Deck Software — The Gold Standard**
- **UX pattern:** Three-panel layout: left = visual grid of LCD buttons (mirrors the physical device), right = action palette (categorized: System, Hotkey, Text, Open, Multi Action, etc.), bottom = property inspector for the selected action.
- **Key features:** Drag actions from the right panel onto any key. Folders (drag "Create Folder" onto a key, then drag actions into it). Pages (up to 10 per profile). Profiles (per-app or per-workflow). Smart Profiles (auto-switch based on active application). Multi Action (sequence of sub-actions on one tap). Icon customization (upload 144×144 px images or browse marketplace). Plugin ecosystem (OBS, Spotify, Discord, etc.).
- **Why it works for lay users:** Zero config files. Everything is visual drag-and-drop. The device grid IS the interface. Property inspector shows only relevant options for the selected action type. No syntax, no file paths, no JSON.
- **Source:** [Elgato Stream Deck Software](https://www.elgato.com/us/en/s/stream-deck-app) | [Quick Start Guide](https://www.elgato.com/ww/en/explorer/products/stream-deck/elgato-stream-deck-quick-start-guide/) | [Multi Action Guide](https://www.elgato.com/us/en/explorer/products/stream-deck/how-to-use-multi-actions/)

**1b. Loupedeck Software — Similar Paradigm, More Dial-Oriented**
- **UX pattern:** Three-panel layout: left = assignable content (actions, adjustments), middle = device canvas with touch buttons and dials, right = overview/navigation of workspaces and pages.
- **Key features:** Drag-and-drop actions onto touch buttons. Workspaces and pages. Custom actions: AppleScript, Multi-toggle, Multi-Action (Macro), Mouse Click/Scroll/ Move, Open App. Dial adjustments for fine-grained control. Round buttons and square buttons are customizable. Marketplace for profiles.
- **Why it works:** Same drag-and-drop visual paradigm as Stream Deck. The device representation is interactive — hover over buttons to see available actions. Workspaces allow different configurations for different software.
- **Source:** [Loupedeck User Support](https://support.loupedeck.com/getting-started.html) | [Loupedeck Setup Guide (PDF)](https://audioeffetti.com/product/documents/LOU/LDD-1903-01.pdf)

**1c. Razer Synapse 4 — Tabbed Device Config**
- **UX pattern:** Dashboard lists all connected devices. Click a device card → tabbed interface (Customize, Performance, Lighting, Calibration). Visual keyboard/mouse diagram for button mapping. Macro tab with recording.
- **Key features:** Button remapping via clicking on the visual device diagram. Hypershift (secondary layer). Profiles with per-game auto-switch. Onboard memory (save profiles to device). Cloud sync across machines. Macro recording with key sequence capture.
- **What's good:** Visual device diagram makes mapping intuitive. Per-game profiles are powerful. Cloud sync is seamless.
- **What's bad:** Synapse 4 UI regression — users report Synapse 3 had a "clean, modern, simple" interface while Synapse 4 is "awful" with a less consistent layout. Heavy, requires account login.
- **Source:** [Razer Synapse 4 Guide](https://aurasync.net/how-to-use-razer-synapse/) | [Razer Synapse 4 Review](https://www.oreateai.com/blog/indepth-experience-report-on-razers-new-blackwidow-v4-keyboard-and-razer-synapse-4-configuration-tool/7fb086fbc60d3246513f700dd26c0a5f) | [User complaint](https://insider.razer.com/the-new-razer-synapse-razer-chroma-app-open-beta-53/the-new-user-interface-is-awful-52985)

**1d. Corsair iCUE — Assignment-List Approach**
- **UX pattern:** Home screen → select keyboard → "Key Assignments" in left sidebar. A list-based assignment system: click + to add a new assignment, select assignment type (keystroke, macro, text, etc.), then click the key on a visual keyboard diagram.
- **Key features:** Macro recording with advanced settings (trigger type: press/hold, repeat: once/toggle/repeat, second action after completion). Visual keyboard diagram for key selection. Per-profile settings. Onboard memory.
- **What's good:** Clear separation between assignment creation and key selection. Macro recording is straightforward.
- **What's bad:** List-based rather than drag-and-drop. Less visual than Stream Deck. The visual keyboard diagram is for selection only, not for drag-and-drop assignment.
- **Source:** [Corsair iCUE Key Assignments](https://help.corsair.com/hc/en-us/articles/9399197817101-How-to-Assign-key-remaps-and-macros-to-your-keyboards) | [Corsair iCUE Macros Guide](https://www.corsair.com/ww/en/explorer/gamer/keyboards/how-to-add-macros-and-remap-keys-in-corsair-icue/)

**1e. Logitech G HUB — Device-Centric with Profile Management**
- **UX pattern:** Home screen shows all connected Logitech devices. Click a device → left menu (Assignments, Lighting, etc.). Button assignments via clicking on a visual device diagram. Per-game profiles with auto-detection.
- **Key features:** Button remapping with visual diagram. Macro recording with delay insertion. Onboard memory (1-5 slots). LIGHTSYNC RGB control. Per-game profile auto-switch.
- **What's good:** Clean, modern interface. Onboard memory means profiles work without software running. Per-game detection is reliable.
- **What's bad:** Occasional sync issues between software and onboard memory. Some users find the profile management confusing.
- **Source:** [Logitech G HUB Basics](https://www.logitechg.com/en-us/software/guides/g-hub-basics) | [G HUB Guide](https://gamerhardware.org/logitech-g-hub-software-guide/)

---

### 2. Gaming Peripheral Config (Keyboard/HOTAS)

**2a. QMK VIA — Web-Based Visual Keymap Configurator**
- **UX pattern:** Web app (usevia.app) or desktop app. Auto-detects VIA-compatible keyboard. Shows a visual rendering of the keyboard. Click any key → select new function from a palette of available keycodes. Changes apply on-the-fly without reflashing firmware.
- **Key features:** Key remapping with visual keyboard. Layers (Fn, etc.). Macros (record key sequences). RGB lighting control (if supported). Layout options (e.g., split spacebar, different bottom row). Onboard memory save.
- **Why it works for lay users:** Plug in keyboard → it's detected automatically. Click a key → choose what it does from a list. No firmware compilation, no code. The visual keyboard is accurate to the physical layout.
- **Limitation:** Requires VIA-compatible firmware pre-installed on the keyboard. Not all QMK keyboards support VIA.
- **Source:** [VIA Website](https://caniusevia.com/) | [VIA GitHub](https://github.com/the-via/app) | [VIA Usage Guide](https://docs.keeb.io/via) | [XDA Guide](https://www.xda-developers.com/how-configure-qmk-keyboards-via/)

**2b. Vial — Real-Time GUI Configurator (VIA Fork)**
- **UX pattern:** Desktop app (Windows/Linux/Mac) + web version (vial.rocks). Auto-detects Vial-compatible keyboard. Click a key → select replacement from a palette. Changes apply instantly in real time.
- **Key features:** Everything VIA has, plus: combos (multiple keys pressed together), tap dance (different actions on tap/hold/double-tap), macros with more flexibility. Real-time changes without any reflash.
- **Why it's better than VIA for lay users:** Even more features exposed through the same click-to-edit paradigm. The web version (vial.rocks) requires no installation.
- **Source:** [Vial Home](https://get.vial.today/) | [Vial GitHub](https://github.com/vial-kb/vial-gui) | [Vial Beginner's Guide](https://artkeeb.com/blogs/news/how-to-use-vial-a-beginners-guide-to-remapping-your-keyboard)

**2c. Thrustmaster T.A.R.G.E.T — Dual-Mode (GUI + Scripting)**
- **UX pattern:** Two modes: (1) Basic GUI profile — visual mapping of axes and buttons, easy keyboard event emulation; (2) Script Editor — full programming language for advanced users.
- **Key features:** GUI mode: select device, map buttons to keyboard keys or mouse events, configure axes. Script mode: full control with event-driven programming. Print function generates a visual reference card of all mappings.
- **What's good:** The dual-mode approach serves both beginners (GUI) and power users (scripting). The print-to-reference-card feature is unique and useful.
- **What's bad:** The GUI is functional but dated. The scripting language has a steep learning curve. Windows-only.
- **Source:** [Thrustmaster T.A.R.G.E.T](https://www.thrustmaster.com/en-us/news/t-a-r-g-e-t-advanced-programming-software/) | [T.A.R.G.E.T Basic Config Guide](https://support.thrustmaster.com/en/kb/1853-en/) | [T.A.R.G.E.T Manual (PDF)](https://ts.thrustmaster.com/download/accessories/pc/hotas/software/TARGET/TARGET_User_Manual_ENG.pdf)

---

### 3. General Config GUI Editors (TOML/JSON/YAML)

**3a. Pydantic Studio — Schema-Driven Form Generation**
- **UX pattern:** Define a Pydantic model (typed schema with descriptions, defaults, constraints) → auto-generate three frontends: (1) console wizard (Yeoman-style prompts), (2) Textual TUI (terminal form), (3) web app (React-backed). Outputs config.yaml, config.toml, or config.json.
- **Why it's important:** The schema IS the contract. Types, constraints, defaults, and descriptions are encoded once in Python. The generated forms are always valid — users cannot produce malformed output. This is the ideal pattern for config-file editors: **never let the user touch raw text**.
- **Limitation:** Requires a Pydantic schema to exist. Not a general-purpose TOML editor — it's schema-specific.
- **Source:** [Pydantic Studio GitHub](https://github.com/invoker-bot/pydantic-studio) | [Pydantic Studio PyPI](https://pypi.org/project/pydantic-studio/0.5.2/)

**3b. MetaConfigurator — JSON Schema → Auto-Generated Forms**
- **UX pattern:** Web-based tool. Define a JSON Schema → MetaConfigurator generates a graphical form for editing data files. Also includes a graphical schema editor (no need to write JSON Schema by hand). AI assistance for schema creation.
- **Why it's important:** Academic project (University of Stuttgart) that explicitly targets "both technical and non-technical users." The model-driven approach (schema → form) is the same philosophy as Pydantic Studio but for JSON Schema. Published research paper validates the approach.
- **Source:** [MetaConfigurator](https://www.metaconfigurator.org/) | [MetaConfigurator GitHub](https://github.com/metaconfigurator/meta-configurator) | [Research Paper](https://doi.org/10.1007/s13222-024-00472-7)

**3c. JSON Editor Online / Tree View Editors — Visual Tree + Code Side-by-Side**
- **UX pattern:** Split view: left = collapsible tree view of JSON/YAML structure, right = code editor. Inline editing of keys and values. Drag-and-drop reorder of nodes. Real-time validation. Type-aware (shows data types, array lengths).
- **What's good:** Tree view makes hierarchical structure visible and navigable. Collapsible nodes reduce cognitive load. Inline editing is faster than form-based for simple values.
- **What's bad:** Still exposes the raw data format. Non-technical users can still produce invalid JSON (missing commas, brackets). Not ideal for lay users who don't understand the data structure.
- **Source:** [JSON Editor Online (dataformatterpro)](https://dataformatterpro.com/json-editor/) | [JSONLint Tree Viewer](https://jsonlint.com/json-tree) | [JSONCraft Viewer](https://jsoncraft.dev/viewer/)

---

### 4. UX Patterns for Lay-User Config Editors — Synthesized Recommendations

Based on the research above, here are concrete UX recommendations for a lay-user config editor (e.g., for a radiology shortcut keypad):

**4.1. Visual Device Canvas (Highest Priority)**
- Show a grid/representation of the physical device buttons. This is the primary interaction surface.
- **Example:** Stream Deck's LCD button grid, VIA's keyboard rendering, Loupedeck's device canvas.
- **Why:** Users map mental model of the physical device to the screen. No abstraction layer needed.

**4.2. Drag-and-Drop Assignment**
- Actions are dragged from a palette onto the device canvas. Drop on a button to assign.
- **Example:** Stream Deck (drag from right panel onto key), Loupedeck (drag from left panel onto device).
- **Why:** Most intuitive mapping action. No "select key → select action" two-step confusion.

**4.3. Pre-Built Action Library with Categories**
- Actions are organized into categories (System, Text, Hotkey, App Launch, Media, etc.).
- Each action has a clear name, icon, and short description.
- **Example:** Stream Deck's categorized action panel, VIA's keycode palette.
- **Why:** Lay users don't know what's possible. A categorized library teaches them.

**4.4. Property Inspector Panel**
- When a button is selected, show a panel with its configurable properties.
- Properties change based on action type (e.g., Hotkey shows a key recorder, Text shows a text input).
- **Example:** Stream Deck's bottom property inspector, Loupedeck's right-side configuration area.
- **Why:** Context-sensitive. Users only see relevant options for the action they chose.

**4.5. Inline Validation + Helpful Error Messages**
- Validate inputs as the user types (onBlur or onInput, not on submit).
- Show error messages next to the field, not in a popup or at the top of the page.
- Use plain language: "Type the text you want to paste" not "Field 'content' is required."
- **Example:** Form validation UX patterns from UXPatterns.dev.
- **Source:** [Form Validation UX Patterns](https://uxpatterns.dev/patterns/forms/form-validation) | [Form Validation UX - 7 Patterns](https://sarvaya.in/blog/form-validation-ux-patterns-real-time-2026)

**4.6. Tooltips / Inline Help / Placeholder Examples**
- Every field should have a tooltip or help icon that explains what it does.
- Placeholder text should show a realistic example: e.g., "e.g., Normal chest, no acute findings" for a radiology shortcut.
- **Example:** Stream Deck's property inspector shows hints in fields. VIA shows key descriptions on hover.
- **Why:** Reduces learning curve. Users don't need to read a manual.

**4.7. Live Preview**
- Show what the button will look like on the device (icon, label) as the user configures it.
- For text-paste actions, show the text that will be pasted.
- **Example:** Stream Deck updates the button image in real time as you change icon/label.
- **Why:** Immediate feedback. "What you see is what you get."

**4.8. Save / Revert / Undo Patterns**
- **Explicit save** (Save button) for config editors where mistakes have consequences. Never auto-save destructive changes.
- **Undo/redo** for all editing operations. Track history per session.
- **Revert** button to discard all unsaved changes and return to last saved state.
- **Dirty state indicator** (e.g., dot on close button, asterisk on tab) when there are unsaved changes.
- **Confirm on unsaved quit** — but only when there ARE unsaved changes. Don't ask unnecessarily.
- **Source:** [Primer Save Patterns](https://primer-docs-preview.github.com/product/ui-patterns/saving/) | [Setting.page Save Patterns](https://setting.page/settings-form-design-save-patterns) | [Oracle Alta Save Model](https://www.oracle.com/webfolder/ux/middleware/alta/patterns/SaveModel.html)

**4.9. Profiles / Pages / Organization**
- Allow multiple profiles (per-user, per-workflow, per-app).
- Within a profile, allow pages (multiple screens of buttons) and folders (nested groups).
- Smart profiles that auto-switch based on the active application.
- **Example:** Stream Deck (profiles + pages + folders + Smart Profiles), Loupedeck (workspaces + pages).
- **Why:** A 6×6 grid fills up fast. Organization prevents overwhelm.

**4.10. Search / Filter for Large Configs**
- When the action library is large, provide search with autocomplete.
- Filter by category, action type, or keyword.
- **Example:** VIA's keycode search, Stream Deck's action search.
- **Why:** Scrolling through hundreds of actions is slow. Search is fast.

**4.11. Drag-and-Drop Reordering**
- Allow reordering of buttons, pages, profiles, and sub-actions within a Multi Action.
- Show visual drop indicators (ghost preview, insertion line).
- **Example:** Stream Deck's folder organization, JSON tree editors' drag-drop reorder.
- **Why:** Users think spatially. Reordering should feel like rearranging physical items.

**4.12. Preventing Invalid States**
- Disable or hide options that are not applicable to the current action type.
- Use dropdowns/selectors instead of free-text fields where possible.
- Never allow the user to create a config that the device cannot interpret.
- **Example:** Stream Deck's property inspector only shows relevant fields per action type.
- **Why:** "An ounce of prevention is worth a pound of cure." Better to prevent errors than to report them.

**4.13. First-Run Onboarding / Guided Setup**
- On first launch, show a welcome screen with a quick setup wizard.
- Offer template profiles (e.g., "Radiology Reporting," "Gaming," "Streaming").
- Highlight the drag-and-drop mechanism with an animation or tooltip.
- **Example:** Stream Deck's quick start guide, Logitech G HUB's first-run device detection.
- **Why:** First impression determines whether a lay user continues or gives up.

**4.14. Onboard Memory / Portability**
- Allow saving config directly to the device (onboard memory) so it works on any computer without the software.
- Show a clear "Save to Device" button with progress indicator.
- **Example:** Logitech G HUB (1-5 onboard slots), Razer Synapse (onboard profiles), VIA (save to keyboard).
- **Why:** The whole point of a hardware configurator is that the config travels with the device.

---

## Sources

### Kept (Strong Sources)

1. **Elgato Stream Deck Software** — The gold standard for macro-pad config UX. Three-panel drag-and-drop layout. [Source](https://www.elgato.com/us/en/s/stream-deck-app)
2. **Elgato Stream Deck Quick Start Guide** — Official documentation showing the drag-and-drop workflow. [Source](https://www.elgato.com/ww/en/explorer/products/stream-deck/elgato-stream-deck-quick-start-guide/)
3. **Elgato Multi Action Guide** — Shows how sub-action sequences work in Stream Deck. [Source](https://www.elgato.com/us/en/explorer/products/stream-deck/how-to-use-multi-actions/)
4. **Loupedeck User Support + Setup Guide** — Three-panel layout, drag-and-drop, workspaces. [Source](https://support.loupedeck.com/getting-started.html)
5. **Razer Synapse 4 Guide** — Tabbed device config, visual device diagram, profiles. [Source](https://aurasync.net/how-to-use-razer-synapse/)
6. **Corsair iCUE Key Assignments** — Assignment-list + visual keyboard selection pattern. [Source](https://help.corsair.com/hc/en-us/articles/9399197817101-How-to-Assign-key-remaps-and-macros-to-your-keyboards)
7. **Logitech G HUB Basics** — Device-centric, visual diagram, onboard memory. [Source](https://www.logitechg.com/en-us/software/guides/g-hub-basics)
8. **VIA Usage Guide** — Web-based visual keyboard configurator, click-to-remap. [Source](https://docs.keeb.io/via)
9. **Vial Home + Manual** — Real-time GUI configurator, click-to-edit paradigm. [Source](https://get.vial.today/)
10. **Pydantic Studio** — Schema-driven form generation for TOML/YAML/JSON. Three frontends. [Source](https://github.com/invoker-bot/pydantic-studio)
11. **MetaConfigurator** — JSON Schema → auto-generated forms. Academic project targeting non-technical users. [Source](https://www.metaconfigurator.org/) | [Research Paper](https://doi.org/10.1007/s13222-024-00472-7)
12. **Primer Save Patterns (GitHub)** — Design guidelines for save/unsaved/dirty state patterns. [Source](https://primer-docs-preview.github.com/product/ui-patterns/saving/)
13. **Setting.page — Save Pattern Comparison** — Inline save vs save bar vs auto-save analysis. [Source](https://setting.page/settings-form-design-save-patterns)
14. **UXPatterns.dev — Form Validation** — Field-level validation, timing, error placement patterns. [Source](https://uxpatterns.dev/patterns/forms/form-validation)
15. **Thrustmaster T.A.R.G.E.T** — Dual-mode (GUI + scripting) for HOTAS config. [Source](https://www.thrustmaster.com/en-us/news/t-a-r-g-e-t-advanced-programming-software/)

### Dropped

- **Sweetwater Stream Deck Guide** — Redundant with official Elgato sources.
- **Various JSON tree viewers** (JSONLint, JSONCraft, etc.) — Too developer-oriented, not lay-user focused.
- **DRKDS Studio** — Odoo-specific, not generalizable.
- **v0 / Lovable / Replit** — AI app builders, not config editors.
- **Ferrite** — Rust/egui text editor, not a visual config tool.
- **TOMLKit.io** — Landing page only, no substantive UX documentation.

---

## Gaps

1. **No dedicated radiology/PACS config editor exists as a reference.** The closest are Stream Deck (general macro pad) and gaming peripheral configurators. RadKeys would be novel in this space.
2. **No comprehensive academic UX research specifically on config-file editors for non-technical users.** The MetaConfigurator paper is the closest but focuses on JSON Schema tooling, not UX patterns per se.
3. **Limited data on what "good" TOML-specific visual editing looks like.** Most TOML tools are developer-oriented (VS Code extensions, CLI tools). Pydantic Studio is the only schema-driven TOML form generator found.
4. **No usability testing data comparing drag-and-drop vs. list-based vs. form-based config editors.** The recommendations above are synthesized from product examples and general form UX research, not from controlled studies of config editors.
5. **Macro recording UX patterns** — While Stream Deck and iCUE support macro recording, detailed UX analysis of the recording workflow (start/stop, delay insertion, editing recorded sequences) was not found in depth.

### Suggested Next Steps

- Prototype a Stream Deck-style three-panel layout (device grid + action palette + property inspector) for RadKeys.
- Conduct a small usability test with radiologists (the target users) comparing drag-and-drop vs. form-based button configuration.
- Evaluate Pydantic Studio's approach for auto-generating the RadKeys config form from a Go struct schema (or equivalent).
- Study Stream Deck's Multi Action UX in detail as a model for RadKeys' "sub-action sequences" (multi-step report phrases).

---

## Acceptance Report

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Research covers all four requested categories (macro/keypad config apps, gaming peripheral config, general config GUI editors, UX patterns) with concrete examples and synthesized recommendations. Scope limited to read-only web research; no files modified in the working repository."
    },
    {
      "id": "criterion-2",
      "status": "satisfied",
      "evidence": "15 strong sources cited with URLs. Findings organized by category with numbered findings. UX recommendations synthesized into 14 concrete patterns. Gaps and suggested next steps documented. Output written to /tmp/radkeys-012/research-config-editor-ux.md as required."
    }
  ],
  "changedFiles": [
    "/tmp/radkeys-012/research-config-editor-ux.md"
  ],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "web_search (15 queries across 4 search passes)",
      "result": "passed",
      "summary": "Searched for Stream Deck, Loupedeck, Razer Synapse, Corsair iCUE, Logitech G HUB, QMK VIA, Vial, Thrustmaster T.A.R.G.E.T, Pydantic Studio, MetaConfigurator, JSON Editor Online, form validation UX patterns, save patterns, and config editor UX patterns. Retrieved 50+ results, filtered to 15 strong sources."
    }
  ],
  "validationOutput": [
    "Output written to /tmp/radkeys-012/research-config-editor-ux.md",
    "15 sources cited with URLs",
    "14 UX pattern recommendations synthesized",
    "4 gaps documented with suggested next steps"
  ],
  "residualRisks": [
    "No usability testing data specifically for config editors — recommendations are synthesized from product examples and general form UX research",
    "No radiology-specific config editor exists as a reference — RadKeys would be novel in this space",
    "TOML-specific visual editing tools are scarce — most findings come from JSON/YAML tools and general macro-pad configurators"
  ],
  "noStagedFiles": true,
  "diffSummary": "New file created: /tmp/radkeys-012/research-config-editor-ux.md (comprehensive research brief on visual config-file editors for non-technical users)",
  "reviewFindings": [
    "no blockers: research complete, all four categories covered, concrete examples provided, UX patterns synthesized, gaps documented"
  ],
  "manualNotes": "Research is read-only. No files in the working repository were modified. The output path /tmp/radkeys-012/research-config-editor-ux.md was used as specified. The acceptance report is included at the end of the research document."
}
```
