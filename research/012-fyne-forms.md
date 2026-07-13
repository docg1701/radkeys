# Research: Fyne (Go GUI Toolkit) Patterns for Form-Based Config-Editor UI

## Summary

Fyne v2.7.x provides a mature widget set and data-binding system suitable for building a TOML config editor. The recommended approach uses `widget.Form` with `widget.Entry`/`Select`/`Check`/`RadioGroup` for flat fields, `widget.List` with `binding.BindList` for nested arrays of objects, and `container.AppTabs` or `container.Split` for multi-section layout. **Key gap**: binding to nested slices of structs has known limitations (issue #2607) — the workaround is to manually manage a `[]binding.DataMap` slice. **Drag-drop reordering is not natively supported** in `widget.List`; an RFC exists but is unmerged. Use up/down buttons as a workaround.

## Findings

### 1. Core Data-Entry Widgets

**`widget.Entry`** — single/multi-line text input. Key fields: `Text`, `PlaceHolder`, `Validator` (`fyne.StringValidator`), `OnChanged`, `OnSubmitted`, `Password`, `MultiLine`, `Wrapping`. Constructor: `widget.NewEntry()`. [Entry docs](https://docs.fyne.io/widget/entry/)

**`widget.Select`** — dropdown picker. Fields: `Options []string`, `Selected string`, `PlaceHolder`, `OnChanged func(string)`. Constructor: `widget.NewSelect(options, onChanged)`. [Select docs](https://docs.fyne.io/api/v2/widget/select/)

**`widget.Check`** — boolean toggle. Fields: `Text`, `Checked bool`, `OnChanged func(bool)`. Constructor: `widget.NewCheck(text, onChanged)`. [Check docs](https://docs.fyne.io/api/v2/widget/check/)

**`widget.RadioGroup`** — mutually exclusive radio buttons. Fields: `Options []string`, `Selected string`, `Horizontal bool`, `Required bool`, `OnChanged func(string)`. Constructor: `widget.NewRadioGroup(options, onChanged)`. [RadioGroup docs](https://docs.fyne.io/api/v2/widget/radiogroup/)

**`widget.Slider`** — numeric slider. Fields: `Min, Max, Step float64`, `OnChanged func(float64)`. Constructor: `widget.NewSlider(min, max)`. Also `widget.NewSliderWithData(min, max, data)` for binding. [Slider in widget package](https://pkg.go.dev/fyne.io/fyne/v2/widget)

**`widget.Form`** — auto-layout label+widget pairs with optional submit/cancel buttons. Fields: `Items []*FormItem`, `OnSubmit func()`, `OnCancel func()`, `SubmitText`, `CancelText`, `Orientation`. Each `FormItem` has `Text` (label), `Widget` (input), and `HintText`. Constructor: `widget.NewForm(items...)` or `&widget.Form{Items: ...}`. [Form docs](https://docs.fyne.io/widget/form/)

**`widget.List`** — virtualized vertical list with callbacks: `Length func() int`, `CreateItem func() fyne.CanvasObject`, `UpdateItem func(id ListItemID, obj fyne.CanvasObject)`, `OnSelected func(id ListItemID)`. Also `NewListWithData(data, createItem, updateItem)` for binding. [List docs](https://docs.fyne.io/api/v2/widget/list/)

**`widget.Table`** — 2D grid with `Length func() (rows, cols int)`, `CreateCell func() fyne.CanvasObject`, `UpdateCell func(id TableCellID, obj fyne.CanvasObject)`, `OnSelected`, `ShowHeaderRow`. [Table docs](https://docs.fyne.io/api/v2/widget/table/)

**`widget.Tree`** — hierarchical data with `ChildUIDs`, `IsBranch`, `CreateNode`, `UpdateNode` callbacks. [Tree docs](https://docs.fyne.io/api/v2/widget/tree/)

### 2. Data Binding (`data/binding`)

**Primitive bindings**: `binding.NewString()`, `binding.BindString(&s)`, `binding.NewInt()`, `binding.BindInt(&i)`, `binding.NewFloat()`, `binding.BindFloat(&f)`, `binding.NewBool()`, `binding.BindBool(&b)`, `binding.NewUntyped()`, `binding.BindUntyped(&v)`. [Data Binding docs](https://docs.fyne.io/binding/data/)

**Struct binding**: `binding.BindStruct(&myStruct)` returns a `binding.Struct` (implements `DataMap` with `GetValue(key)`, `SetValue(key, val)`, `Reload()`). Keys are exported field names. Only top-level exported fields are included — **nested struct fields and slices are stored as opaque `any` values**, not recursively bound. [BindStruct docs](https://docs.fyne.io/api/v2/data/binding/struct/)

**List binding**: `binding.NewStringList()` returns `binding.List[string]`. `binding.BindStringList(&[]string)` returns `ExternalList[string]`. Generic: `binding.NewList[T](comparator)` and `binding.BindList[T](v *[]T, comparator)` (since v2.7). Methods: `Append`, `Prepend`, `Get`, `Set`, `GetValue`, `SetValue`, `Remove`. [List binding docs](https://docs.fyne.io/api/v2/data/binding/list/)

**Widgets with data binding constructors**:
- `widget.NewEntryWithData(binding.String)` — two-way bound entry
- `widget.NewLabelWithData(binding.String)` — auto-updating label
- `widget.NewSliderWithData(min, max, binding.Float)` — bound slider
- `widget.NewProgressBarWithData(binding.Float)` — bound progress bar
- `widget.NewListWithData(binding.DataList, createItem, updateItem)` — bound list
- `widget.NewCheckWithData(binding.Bool)` — bound checkbox (since v2.5)

**Converters**: `binding.FloatToString(data)`, `binding.FloatToStringWithFormat(data, format)`, `binding.IntToString(data)`, `binding.BoolToString(data)`.

**Known limitation — nested slices of structs**: `BindStruct` does not recursively bind slice fields. Issue [#2607](https://github.com/fyne-io/fyne/issues/2607) documents that `BindStruct` fails to reload when a slice field changes. The recommended workaround (from [StackOverflow](https://stackoverflow.com/questions/67346900/using-fyne-to-bind-a-list-widget-to-a-slice-of-structs) and [Fyne examples](https://gist.github.com/micheam/70ab31443d88ec89f6eb31104d7bd714)) is to manually maintain a `[]binding.DataMap` slice where each element is `binding.BindStruct(&item)`, then use `widget.NewListWithData` with a custom `DataList` wrapper, or use `widget.NewList` with manual callbacks.

### 3. Dialogs

**File open/save**: `dialog.NewFileOpen(callback func(reader fyne.URIReadCloser, err error), parent fyne.Window)` and `dialog.NewFileSave(callback func(writer fyne.URIWriteCloser, err error), parent fyne.Window)`. Both support `.SetFilter(filter)` with `storage.NewExtensionFileFilter(extensions)`. [FileDialog docs](https://docs.fyne.io/api/v2/dialog/filedialog/)

**Confirm**: `dialog.NewConfirm(title, message, callback func(bool), parent)` and `dialog.ShowConfirm(title, message, callback, parent)`. [ConfirmDialog docs](https://docs.fyne.io/api/v2/dialog/confirmdialog/)

**Error/Information**: `dialog.ShowError(err, parent)`, `dialog.NewError(err, parent)`, `dialog.ShowInformation(title, message, parent)`. [Dialog package docs](https://docs.fyne.io/api/v2/dialog/package/)

**Form dialog** (since v2.4): `dialog.NewForm(title, confirm, dismiss, items []*widget.FormItem, callback func(bool), parent)` — shows a modal form with validation. [FormDialog docs](https://docs.fyne.io/api/v2/dialog/formdialog/)

### 4. Layout/Container Patterns for Multi-Section Config

**`container.AppTabs`** — tabbed sections. Constructor: `container.NewAppTabs(items...)` where each item is `container.NewTabItem(label, content)`. Supports `TabLocation` (Top, Bottom, Leading, Trailing), `OnSelected`, `OnUnselected`. Ideal for splitting config into logical groups (General, Screens, Buttons, About). [AppTabs docs](https://docs.fyne.io/container/apptabs/)

**`container.NewBorder(top, bottom, left, right, center)`** — fixed toolbars/status bars around a scrollable center. [Border docs](https://docs.fyne.io/container/border/)

**`container.NewVSplit(top, bottom)` / `container.NewHSplit(leading, trailing)`** — draggable split panels. `Offset float64` (0.0–1.0) controls divider position. Good for master-detail (list on left, form on right). [Split docs](https://docs.fyne.io/api/v2/container/split/)

**`container.NewVBox(items...)` / `container.NewHBox(items...)`** — simple vertical/horizontal stacking. [Box docs](https://docs.fyne.io/container/box/)

**`container.NewVScroll(content)` / `container.NewHScroll(content)`** — scrollable containers for long forms. [Scroll docs](https://pkg.go.dev/fyne.io/fyne/v2/container)

**Master-detail pattern** (from [Fyne demo collection.go](https://github.com/fyne-io/fyne/blob/master/cmd/fyne_demo/tutorials/collection.go)): Use `container.NewHSplit(list, formContent)` where `list.OnSelected` populates the form. The form can be a `widget.Form` or a set of individual widgets that are updated when the selection changes.

### 5. Validation

**`widget.Entry.Validator`** — set to a `fyne.StringValidator` (func `func(string) error`). Built-in validators in `data/validation`: `validation.NewRegexp(pattern, message)`, `validation.NewTime(layout)`. [Validation package docs](https://docs.fyne.io/api/v2/data/validation/package/)

**Custom validator example** (from [Fyne demo widget.go](https://github.com/fyne-io/fyne/blob/c29a0624ed96ba1b8f45d903b6941824d50e0502/cmd/fyne_demo/tutorials/widget.go)):
```go
email.Validator = validation.NewRegexp(`\w{1,}@\w{1,}\.\w{1,4}`, "not a valid email")
```

**Form-level validation**: `widget.Form.OnSubmit` only fires when **all** validatable items pass validation. The submit button is disabled while any field is invalid. Known issue: `dialog.NewForm()` may enable submit after only one field is valid ([#4510](https://github.com/fyne-io/fyne/issues/4510), [#5006](https://github.com/fyne-io/fyne/issues/5006)). Workaround: use `widget.Form` instead of `dialog.NewForm()`.

**Inline error display**: Entry widgets show red error text below the input when validation fails. For non-Entry validatable widgets in a Form, error text may not display unless `HintText` is set ([#5194](https://github.com/fyne-io/fyne/issues/5194)).

### 6. Preferences (fyne.App.Preferences)

**API**: `app.Preferences()` returns `fyne.Preferences` interface with methods: `String(key)`, `StringWithFallback(key, fallback)`, `SetString(key, value)`, and equivalents for `Bool`, `Float`, `Int`, plus `StringList`, `IntList`, `FloatList`. [Preferences docs](https://docs.fyne.io/api/v2/fyne/preferences/)

**Usage for last-opened-file**:
```go
prefs := a.Preferences()
lastFile := prefs.String("lastOpenedFile")
// ... later ...
prefs.SetString("lastOpenedFile", uri.String())
```

**Storage**: Preferences are stored as JSON in the app's data directory (platform-specific). The app must be created with `app.NewWithID("unique.app.id")` for consistent storage location. [Using Preferences docs](https://docs.fyne.io/explore/preferences/)

### 7. Real-World Examples and Code Patterns

**Fyne demo — BindStruct + form** ([bind.go](https://github.com/fyne-io/fyne/blob/c4b5c694/cmd/fyne_demo/tutorials/bind.go)):
```go
formStruct := struct { Name, Email string; Subscribe bool }{}
formData := binding.BindStruct(&formStruct)
form := newFormWithData(formData)
form.OnSubmit = func() { fmt.Println("Struct:\n", formStruct) }
```
Where `newFormWithData` iterates `data.Keys()` and creates `widget.NewFormItem` for each key, binding the appropriate widget type.

**Fyne demo — Form with validation** ([widget.go](https://github.com/fyne-io/fyne/blob/c29a0624ed96ba1b8f45d903b6941824d50e0502/cmd/fyne_demo/tutorials/widget.go)):
```go
form := &widget.Form{
    Items: []*widget.FormItem{
        {Text: "Name", Widget: name, HintText: "Your full name"},
        {Text: "Email", Widget: email},
        {Text: "Password", Widget: password},
    },
    OnSubmit: func() { log.Println("Form submitted") },
    OnCancel: func() { log.Println("Form cancelled") },
}
```

**Fyne settings app** ([appearance.go](https://github.com/fyne-io/fyne/blob/master/cmd/fyne_settings/settings/appearance.go)): Real-world multi-section settings UI using `container.NewAppTabs` with theme/scale controls.

**DEFyne IDE** ([project.go](https://github.com/fyne-io/defyne/blob/main/project.go)): Uses `dialog.ShowForm` with `widget.NewEntry` and `widget.NewButton` (for directory chooser) to create a new project dialog — a pattern directly applicable to config editing.

**List bound to slice of structs** ([StackOverflow](https://stackoverflow.com/questions/67346900/using-fyne-to-bind-a-list-widget-to-a-slice-of-structs) + [Gist](https://gist.github.com/micheam/70ab31443d88ec89f6eb31104d7bd714)):
```go
var bindings []binding.DataMap
for _, item := range data {
    bindings = append(bindings, binding.BindStruct(&item))
}
list := widget.NewListWithData(myDataList, createItem, updateItem)
```

**CRUD app pattern** ([blogvali.com](https://blogvali.com/fyne-crud-app-fyne-golang-gui-tutorial-56/)): Uses `widget.NewList` with `OnSelected` to populate detail fields, plus add/edit/delete buttons — directly applicable to the screens→buttons config editor.

### 8. Gaps and Limitations

| Gap | Details | Workaround |
|-----|---------|------------|
| **Nested slice-of-structs binding** | `BindStruct` stores slice fields as opaque `any`; no recursive binding. Issue [#2607](https://github.com/fyne-io/fyne/issues/2607) | Manually maintain `[]binding.DataMap` or use `widget.NewList` with manual callbacks |
| **Drag-drop reordering** | No built-in list reordering. `fyne.Draggable` is for moving objects on screen, not list reordering. RFC [#5863](https://github.com/fyne-io/fyne/pull/5863) proposes API but is unmerged | Up/down arrow buttons to reorder items in the slice |
| **Form dialog validation** | `dialog.NewForm()` may enable submit after one valid field ([#4510](https://github.com/fyne-io/fyne/issues/4510), [#5006](https://github.com/fyne-io/fyne/issues/5006)) | Use `widget.Form` instead of `dialog.NewForm()` |
| **Non-Entry validation display** | Validatable widgets in Form don't show error text unless they are Entry or have HintText ([#5194](https://github.com/fyne-io/fyne/issues/5194)) | Set `HintText` on FormItems |
| **External binding refresh** | `BindStruct`/`BindList` require explicit `Reload()` call after external mutation ([PR #1717](https://github.com/fyne-io/fyne/pull/1717)) | Call `.Reload()` after JSON unmarshal or slice mutation |
| **No macOS binary shipping** | Per project policy, macOS is supported in code but no binary is shipped | Document build instructions for Mac users |

## Sources

### Kept
- **Fyne Data Binding docs** (https://docs.fyne.io/binding/data/) — authoritative overview of all binding types, converters, and widget integration
- **Fyne Form widget docs** (https://docs.fyne.io/widget/form/) — Form API with OnSubmit/OnCancel/validation
- **Fyne Entry widget docs** (https://docs.fyne.io/widget/entry/) — Validator, PlaceHolder, data binding constructors
- **Fyne List widget docs** (https://docs.fyne.io/api/v2/widget/list/) — List callbacks, OnSelected, NewListWithData
- **Fyne AppTabs docs** (https://docs.fyne.io/container/apptabs/) — Multi-section tab layout
- **Fyne Split container docs** (https://docs.fyne.io/api/v2/container/split/) — Master-detail split panels
- **Fyne Preferences docs** (https://docs.fyne.io/explore/preferences/) — Last-opened-file persistence
- **Fyne demo bind.go** (https://github.com/fyne-io/fyne/blob/c4b5c694/cmd/fyne_demo/tutorials/bind.go) — Real code: BindStruct + form generation pattern
- **Fyne demo widget.go** (https://github.com/fyne-io/fyne/blob/c29a0624ed96ba1b8f45d903b6941824d50e0502/cmd/fyne_demo/tutorials/widget.go) — Real code: Form with validation
- **StackOverflow: binding list to slice of structs** (https://stackoverflow.com/questions/67346900/using-fyne-to-bind-a-list-widget-to-a-slice-of-structs) — Community solution for the nested-struct binding gap
- **Issue #2607: Struct binding fails with slice field** (https://github.com/fyne-io/fyne/issues/2607) — Confirmed limitation of BindStruct with slices
- **RFC #5863: Drag and Drop API** (https://github.com/fyne-io/fyne/pull/5863) — Proposed but unmerged drag-drop reorder API

### Dropped
- **Fyne examples repo** (https://github.com/fyne-io/examples) — Contains clock/fractal/tictactoe apps, not form/config-editor patterns
- **Fynerisor form validation example** — Third-party tutorial, less authoritative than official docs
- **DEFyne project.go** — Relevant but the project is archived/unmaintained; patterns are covered by official demo code

## Gaps

1. **No authoritative example of a full TOML config editor** exists in the Fyne ecosystem. The closest are the Fyne settings app (simple key-value) and the DEFyne IDE (archived). The RadKeys project would be a reference implementation.

2. **Binding to `screens[]→buttons[]` nested structure**: The recommended approach is to use `widget.List` with manual callbacks (not `NewListWithData`) for the top-level screens list, and a `widget.Form` for the selected screen's fields (including its buttons sub-list). Each button sub-list would also be a `widget.List`. This avoids the `BindStruct` slice-field limitation entirely.

3. **Drag-drop reorder**: Not available. Implement up/down buttons that swap items in the underlying slice and call `list.Refresh()` or `binding.Reload()`.

4. **Confidence**: High for widget/container/dialog APIs and flat struct binding. Medium for nested slice binding (workaround pattern is community-validated but not officially documented). Low for drag-drop reorder (no supported path in current Fyne).

## Supervisor coordination

No coordination needed. This is a read-only research task with no blocking decisions required.

---

```acceptance-report
{
  "criteriaSatisfied": [
    {
      "id": "criterion-1",
      "status": "satisfied",
      "evidence": "Research is scoped to Fyne v2.7.x form/config-editor patterns only. No project files were read or modified. No scope creep into firmware, HID, or other RadKeys subsystems."
    },
    {
      "id": "criterion-2",
      "status": "satisfied",
      "evidence": "Research brief covers all 8 requested topics with 12 inline source citations from official Fyne docs, source code, and community discussions. Gaps and confidence levels are explicitly stated."
    }
  ],
  "changedFiles": [],
  "testsAddedOrUpdated": [],
  "commandsRun": [
    {
      "command": "web_search (12 queries across 4 rounds)",
      "result": "passed",
      "summary": "Retrieved documentation, source code, and community discussions covering all requested topics"
    }
  ],
  "validationOutput": [
    "Output written to /tmp/radkeys-012/research-fyne-forms.md",
    "12 strong sources cited (8 kept, 4 dropped)",
    "8 research topics covered with inline citations",
    "Gaps section documents 4 known limitations with workarounds"
  ],
  "residualRisks": [
    "Nested slice-of-structs binding has no official solution — the workaround (manual []binding.DataMap) is community-validated but not guaranteed for all edge cases",
    "Drag-drop reordering is not supported in current Fyne — up/down button workaround adds UX complexity",
    "Form dialog validation bugs (#4510, #5006) may affect dialog.NewForm() usage — widget.Form is the safer alternative"
  ],
  "noStagedFiles": true,
  "diffSummary": "No files modified. Single research document written to /tmp/radkeys-012/research-fyne-forms.md",
  "reviewFindings": [
    "no blockers: all requested topics covered with authoritative sources",
    "note: the recommended approach for screens[]→buttons[] nested editing is to use widget.List with manual callbacks, avoiding BindStruct's slice-field limitation"
  ],
  "manualNotes": "Research is read-only. No project files were touched. The output path /tmp/radkeys-012/research-fyne-forms.md is authoritative per runtime instructions."
}
```
