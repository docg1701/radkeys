# Research: Fyne "always on top" window API — status in released v2.7.4

Investigation date: 2026-07-07. Fyne latest released tag at this date: **v2.7.4** (published 2026-05-12). v2.8.0 is at the **release-candidate** stage (v2.8.0-rc1) and has **not** been published as a final/stable release yet.

## Summary

PR #6184 ("Add more requests to desktop Window - Always on Top and Position") **was merged** on 2026-03-28 (merge commit `6fff42b`), but into the **`develop` branch** — i.e. the upcoming **v2.8.0** line — not into the `release/v2.7.x` branch that v2.7.4 was cut from. Therefore **no released Fyne version (including v2.7.4) contains `RequestAlwaysOnTop` / `RequestPosition` / the `desktop.Window` interface.** In v2.7.4 the `fyne.io/fyne/v2/driver/desktop` package contains only `driver.go` (the `Driver` interface: `CreateSplashWindow`, `CurrentKeyModifiers`) and there is **no `window.go`** and **no symbol containing "AlwaysOnTop"** — confirmed by reading the installed module cache directly. The claim that `desktop.Window.RequestAlwaysOnTop()` exists in v2.7.4 is **false**.

## Findings

1. **PR #6184 is merged, but into `develop` (the v2.8 line), not into any 2.7.x release.**
   - PR state: `merged`; author `andydotxyz`; created 2026-03-09, merged 2026-03-28T15:07:07Z; merge commit `6fff42b93136a644b93badd6ddb481458e39c108`; `+45 -1 in 3 files`; description "Fixes #1129 and #1155".
   - The PR is part of a deliberate "request" pattern for desktop window features that all land on `develop` for 2.8: sibling PR #6153 ("Support opening on secondary screen", merged 2026-03-09 into `fyne-io:develop`) introduced `RequestFullScreenSecondary` and created the new `driver/desktop/window.go` + `desktop.Window` interface; #6184 then added `RequestPosition` (commit `ae3338c`) and `RequestAlwaysOnTop`. PR #6270 (merged 2026-05-03) explicitly states its base is `fyne-io:develop`, confirming the 2.8 feature branch.
   - [PR #6184](https://github.com/fyne-io/fyne/pull/6184), [commit ae3338c](https://github.com/fyne-io/fyne/commit/ae3338c02c1a7635a65edd3ca8eb6c9273a9db58), [PR #6153](https://github.com/fyne-io/fyne/pull/6153), [PR #6270](https://github.com/fyne-io/fyne/pull/6270)

2. **v2.7.4 was cut from `release/v2.7.x` as a bugfix-only release and predates/excludes the feature.**
   - v2.7.4 release: "bug fixes and performance improvements aplenty", published 2026-05-12. The release build is `d9f0beb Merge branch 'release/v2.7.x'`. Its changelog lists only fixes (SIGSEGV in `glfwPollEvents`, raster stretching, infinite progress bar, etc.) — no new public window API.
   - Although PR #6184 merged (2026-03-28) *before* the v2.7.4 date (2026-05-12), it was merged to `develop`, not backported to `release/v2.7.x`, so it is absent from the v2.7.4 tag.
   - [v2.7.4 release](https://github.com/fyne-io/fyne/releases), [v2.7.3...v2.7.4 compare](https://github.com/fyne-io/fyne/compare/v2.7.3...v2.7.4), [commit d9f0beb](https://github.com/fyne-io/fyne/commit/d9f0bebaa389f90e52882ee830323c9a73cd5a8e)

3. **Local module-cache verification confirms NO always-on-top API in v2.7.4.**
   - Reading `/home/galvani/go/pkg/mod/fyne.io/fyne/v2@v2.7.4/driver/desktop/driver.go` shows the `desktop` package contains **only** the `Driver` interface (`CreateSplashWindow() fyne.Window`, `CurrentKeyModifiers() fyne.KeyModifier` — the latter `// Since: 2.4`).
   - The file `driver/desktop/window.go` does **not exist** in v2.7.4 (read attempt returned ENOENT). There is no `desktop.Window` interface and no `RequestAlwaysOnTop`/`RequestPosition`/`RequestFullScreenSecondary` symbol anywhere in v2.7.4 — consistent with the background fact you already verified. The module download cache lists only `v2.7.4` (no v2.8 cached locally).
   - [pkg.go.dev desktop package @ v2.7.4](https://pkg.go.dev/fyne.io/fyne/v2/driver/desktop) (shows only `Driver` type for the released version)

4. **The exact public API exists only on `develop` (for the upcoming v2.8.0).**
   - Package path: `fyne.io/fyne/v2/driver/desktop`; new type: `Window` (an interface, in the new file `driver/desktop/window.go`). Methods added by the "request" series: `RequestFullScreenSecondary()` (#6153), `RequestPosition(...)` (commit `ae3338c`, +7 to `driver/desktop/window.go`, +13 to `internal/driver/glfw/window_desktop.go`), and `RequestAlwaysOnTop()` (#6184). The PR checklist states "Public APIs match existing style and have `Since:` line" — these are tagged `Since: 2.8`.
   - **Caveat on exact signatures:** the web search snippets exposed the PR metadata and file-level diffs but **not the verbatim method signatures**. The method *names*, *package*, and *type* are confirmed; the precise parameter list of `RequestAlwaysOnTop` (no-arg toggle vs. `RequestAlwaysOnTop(bool)`) and `RequestPosition` (likely `RequestPosition(fyne.Position)`) should be verified against the `develop` branch source or the v2.8.0 release once published. Confidence on names/package/type: high; confidence on exact parameter lists: medium.
   - [PR #6184 files](https://github.com/fyne-io/fyne/pull/6184/files), [commit ae3338c](https://github.com/fyne-io/fyne/commit/ae3338c02c1a7635a65edd3ca8eb6c9273a9db58)

5. **v2.8.0 is not yet final as of 2026-07-07.**
   - The Releases page still shows v2.7.4 as "Latest". The v2.8.0 schedule (Wiki) targeted 2026-07-03 for the 2.8 final, but issue #6400 ("Arm32 build broken in v2.8.0-rc1", created 2026-07-05, still open 2026-07-06) shows v2.8.0 is at the **rc1** stage with a known arm32 blocker (upstream go-gl/glfw#416, worked around via the `fyne-io/glfw` fork). So the `RequestAlwaysOnTop` API is not yet in a stable release.
   - [Releases](https://github.com/fyne-io/fyne/releases), [Wiki Releases schedule](https://github.com/fyne-io/fyne/wiki/Releases), [issue #6400](https://github.com/fyne-io/fyne/issues/6400)

6. **`SetMaster()` is NOT a workaround for always-on-top.**
   - `fyne.Window.SetMaster()` only marks a window so that closing it exits the app; it has no effect on window z-order. No existing public v2.7.4 API raises a window above others persistently (`RequestFocus()` raises+focuses once but does not pin it on top).
   - [fyne.Window docs](https://docs.fyne.io/api/v2/fyne/window/), [Window handling docs](https://docs.fyne.io/started/windows/)

7. **GLFW `Floating` hint is the underlying mechanism used by the new API.**
   - GLFW's `GLFW_FLOATING` window hint/attribute "specifies whether the windowed mode window will be floating above other regular windows, also called topmost or always-on-top." The community patch shown in issue #3429 calls `w.viewport.SetAttrib(glfw.Floating, 1)` inside `runOnMainWhenCreated`. This is exactly the pattern the merged Fyne PR implements internally in `internal/driver/glfw/window_desktop.go`. Supported by GLFW on Windows, macOS, and Linux (X11/Wayland) — though on Linux some window managers may ignore the floating hint.
   - [GLFW window guide](https://www.glfw.org/docs/3.3/window_guide.html), [go-gl/glfw Floating hint](https://github.com/go-gl/glfw/blob/9c147ed2fc8c/v3.3/glfw/window.go), [issue #3429 community code](https://github.com/fyne-io/fyne/issues/3429), [issue #1129](https://github.com/fyne-io/fyne/issues/1129)

## Viable workarounds in Fyne v2.7.4 (no public always-on-top API)

| # | Approach | Feasibility | Risk |
|---|----------|-------------|------|
| A | **`go.mod` replace → pin to `develop` (or v2.8.0-rc1 / final once out)** and use `desktop.Window.RequestAlwaysOnTop()`. Type-assert: `if dw, ok := w.(desktop.Window); ok { dw.RequestAlwaysOnTop() }`. | High — this is the real, supported API. | Medium: pre-release; arm32 broken in rc1 (#6400); API could still change before 2.8.0 final. Best waited for v2.8.0 stable. |
| B | **Maintain a small fork/patch** of `internal/driver/glfw/window.go` (or `window_desktop.go`) adding a `Topmost(bool)` / `Floating(bool)` method that calls `w.viewport.SetAttrib(glfw.Floating, b)` inside `runOnMainWhenCreated` (the exact code shown in issue #3429). Reference the fork via `go.mod` `replace`. | High on Windows/macOS/Linux desktop (GLFW_FLOATING supported). | Medium: keeps a fork in sync with upstream; touches unexported `viewport`/`runOnMainWhenCreated`; Linux WM may ignore the hint. Drops away once you move to v2.8.0. |
| C | **`unsafe` reflection to reach the unexported `glfw.window.viewport`** field and call `SetAttrib(glfw.Floating, 1)`. | Technically possible. | **High — do not use.** Fragile across Fyne/go-gl versions, breaks on non-glfw drivers (mobile/wasm), no thread-safety guarantees. Violates the internal boundary. |
| D | **Platform-native OS calls** bypassing Fyne: Win32 `SetWindowPos(HWND_TOPMOST)` (Windows), `[NSWindow setLevel:]` (macOS), X11 `_NET_WM_STATE_ABOVE` / Wayland (Linux). Requires cgo and obtaining the HWND/NSWindow/X11 window id. | Works, used by other Go GUIs. | High effort; not cross-platform; obtaining the native handle from Fyne is itself unsupported (no public API to get the raw window handle in v2.7.4). |
| E | `SetMaster()` / `RequestFocus()` / `CenterOnScreen()`. | None for pinning on top. | N/A — these do not keep a window above others. |

## Recommended path for a Go app on Fyne v2.7.4 needing always-on-top

1. **Preferred:** wait for / upgrade to **Fyne v2.8.0 stable** (imminent; currently rc1) and use the official API:
   ```go
   import "fyne.io/fyne/v2/driver/desktop"
   // w is a fyne.Window from the glfw (desktop) driver
   if dw, ok := w.(desktop.Window); ok {
       dw.RequestAlwaysOnTop() // exact signature: verify against v2.8.0 godoc
   }
   ```
   This is the only supported, cross-platform (Windows/macOS/Linux) solution.
2. **If you must stay on v2.7.4 today:** use a `go.mod` `replace` directive pointing to a minimal fork of Fyne v2.7.4 that adds the `glfw.Floating` attribute call (Approach B). Keep the patch isolated so it is trivial to delete once you upgrade to v2.8.0. Test on your real target Linux window manager (GNOME/KDE) since floating is WM-dependent.
3. **Avoid** `unsafe` reflection (C) and hand-rolled native calls (D) unless you have a single fixed platform and accept the maintenance cost.

## Confidence

- PR #6184 merged into `develop` (for v2.8.0): **High** (PR metadata + merge commit + sibling-PR base `fyne-io:develop` + local cache).
- No `RequestAlwaysOnTop` / `desktop.Window` in released v2.7.4: **Very High** (direct read of installed module cache: `driver/desktop` has only `driver.go`; `window.go` ENOENT).
- Exact method signatures on `develop`: **Medium** (names/package/type confirmed; verbatim parameter lists not retrieved from snippets — verify against v2.8.0 godoc/source).
- v2.8.0 not yet stable as of 2026-07-07: **High** (Releases page shows v2.7.4 as Latest; #6400 confirms rc1 stage).

## Sources

- Kept:
  - [PR #6184 — Add more requests to desktop Window: Always on Top and Position](https://github.com/fyne-io/fyne/pull/6184) — primary source for the merge state/date/branch of the feature.
  - [Commit ae3338c — "Also allow a position to be requested" (Fixes #1155)](https://github.com/fyne-io/fyne/commit/ae3338c02c1a7635a65edd3ca8eb6c9273a9db58) — shows `driver/desktop/window.go` and `internal/driver/glfw/window_desktop.go` are the files touched.
  - [PR #6153 — Support opening on secondary screen](https://github.com/fyne-io/fyne/pull/6153) — introduces the `desktop.Window` "request" pattern on `develop`.
  - [PR #6270 — Add a secondary monitor check](https://github.com/fyne-io/fyne/pull/6270) — confirms `fyne-io:develop` as the 2.8 feature base.
  - [v2.7.4 release page](https://github.com/fyne-io/fyne/releases) and [v2.7.3...v2.7.4 compare](https://github.com/fyne-io/fyne/compare/v2.7.3...v2.7.4) — v2.7.4 is bugfix-only, cut 2026-05-12.
  - [Commit d9f0beb — Merge branch 'release/v2.7.x'](https://github.com/fyne-io/fyne/commit/d9f0bebaa389f90e52882ee830323c9a73cd5a8e) — confirms the v2.7.4 release branch.
  - [Wiki Releases schedule](https://github.com/fyne-io/fyne/wiki/Releases) — v2.8.0 timeline (feature freeze 31 May, code freeze 26 Jun, target 3 Jul 2026).
  - [Issue #6400 — Arm32 build broken in v2.8.0-rc1](https://github.com/fyne-io/fyne/issues/6400) — v2.8.0 still at rc1 as of 2026-07-05/06.
  - [Issue #1129 — Ability to create always-on-top window](https://github.com/fyne-io/fyne/issues/1129) — long-standing feature request closed by #6184.
  - [Issue #1155 — Reposition Window programmatically](https://github.com/fyne-io/fyne/issues/1155) — closed by #6184 (RequestPosition).
  - [Issue #3429 — Add option to make window floating / always on top](https://github.com/fyne-io/fyne/issues/3429) — community `glfw.Floating` patch and discussion.
  - [GLFW window guide (GLFW_FLOATING)](https://www.glfw.org/docs/3.3/window_guide.html) and [go-gl/glfw Floating hint](https://github.com/go-gl/glfw/blob/9c147ed2fc8c/v3.3/glfw/window.go) — underlying mechanism.
  - [pkg.go.dev desktop package](https://pkg.go.dev/fyne.io/fyne/v2/driver/desktop) and [fyne.Window docs](https://docs.fyne.io/api/v2/fyne/window/) — confirm no public always-on-top API in released versions.
  - Local read of `…/v2@v2.7.4/driver/desktop/driver.go` and ENOENT for `…/driver/desktop/window.go` — confirms the absence in the installed v2.7.4.
- Dropped:
  - Various SEO/aggregator mirrors of the GitHub pages (newreleases.io, deepwiki) — redundant with the primary GitHub sources.
  - pkg.go.dev `fyne` root package page — not relevant to the desktop window API.

## Gaps

- **Verbatim method signatures** of `RequestAlwaysOnTop` / `RequestPosition` / `RequestFullScreenSecondary` on the `develop` `desktop.Window` interface were not extractable from web-search snippets. Recommend reading `driver/desktop/window.go` at the `develop` branch HEAD (or the v2.8.0 godoc once published) to confirm the exact parameter lists before coding against them.
- **v2.8.0 final release date** is not confirmed published as of 2026-07-07 (still rc1 with an open arm32 blocker). Re-check the Releases page before committing to a v2.8.0 dependency.
- **Linux WM behavior** for `GLFW_FLOATING` is environment-dependent (GNOME/KDE/X11/Wayland) — needs empirical testing on the target desktop.

## Supervisor coordination
No decision needed; no files edited (research-only). Findings written to the authoritative output path only.