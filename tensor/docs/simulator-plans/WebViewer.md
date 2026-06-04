# WebViewer Simulator Implementation Plan

## Overview

`WebViewer` (category **USERINTERFACE**, `nonVisible = false`) is a visible component that embeds a web browser surface inside an App Inventor screen. On Android it wraps a native `WebView`: it loads a `HomeUrl`, optionally follows links, lets the user fill in web forms, exposes navigation history (`GoBack`/`GoForward`/`Reload`), and provides a two-way string channel (`WebViewString`) so app blocks and JavaScript running in the page can exchange a single string via `window.AppInventor.getWebViewString()` / `setWebViewString(...)`. The spec's `helpString` explicitly warns it "is not a full browser."

Container relationship: standalone visible component. It is a leaf — it contains no children and may be placed inside `Form`/`Screen` or any arrangement, exactly like `Image` or `ListView`. There is no parent/child containment rule to encode.

In the simulator it currently renders as a dashed "unsupported" placeholder because `WebViewer` is absent from `SIMULATION_SUPPORTED_TYPES`.

## Feasibility Verdict

**Partially feasible.**

The visible surface maps naturally to an HTML `<iframe>`, which is the closest web-platform analog to an Android `WebView`. Loading a `HomeUrl`, sizing it, showing/hiding it, navigating via `GoToUrl`/`GoHome`, and reloading are all directly implementable. However, several core behaviors are blocked by browser security (the same-origin policy and `X-Frame-Options` / `Content-Security-Policy: frame-ancestors`), which a client-side simulator cannot bypass:

- **Cross-origin pages frequently refuse to load.** Major sites (Google, YouTube, GitHub, most banking/login pages) send `X-Frame-Options: DENY/SAMEORIGIN` or a restrictive `frame-ancestors` CSP. The `<iframe>` will render blank with a console error the simulator cannot intercept. A real device `WebView` has no such restriction. This is the single biggest fidelity gap and cannot be worked around in-browser.
- **History introspection is impossible.** The page inside a cross-origin `<iframe>` does not expose its history length or navigation state. `iframe.contentWindow.history.back()` throws a `SecurityException` for cross-origin frames. So `GoBack`/`GoForward`/`CanGoBack`/`CanGoForward` can only be approximated against a **simulator-maintained navigation stack** of URLs we explicitly set, not the page's own internal navigation.
- **`CurrentUrl` / `CurrentPageTitle` cannot be read for cross-origin frames.** `iframe.contentWindow.location.href` and `.document.title` throw on cross-origin access. We can only report the URL we last *set*, and we cannot read the page's `<title>` at all unless it is same-origin.
- **Link clicks inside the frame are invisible to us.** We cannot intercept in-frame navigations on cross-origin content, so `FollowLinks`, `BeforePageLoad`, and updating `CurrentUrl` after a user clicks a link inside the page are not observable. `BeforePageLoad`/`PageLoaded` can only fire for navigations *we* initiate.
- **`WebViewString` bridge requires `window.AppInventor`.** The real bridge is a JS interface injected into the page. We can inject `window.AppInventor` only into **same-origin** iframe content (e.g. a `srcdoc` page or a same-origin asset). For arbitrary cross-origin pages, `getWebViewString`/`setWebViewString` are unreachable, so `WebViewStringChange` cannot fire from page JS. `postMessage` could be used if the loaded page cooperates, but arbitrary third-party pages will not.
- **`RunJavaScript` on cross-origin frames is blocked** by the same-origin policy. It can only run against same-origin / `srcdoc` content.
- **`IgnoreSslErrors`, `ClearCaches`, `ClearCookies`, `ClearLocations`, `UsesCamera/Location/Microphone`, `PromptforPermission`, `StopLoading`** have no browser equivalent the simulator controls; the browser governs TLS, cache, cookies, and permission prompts for iframe content itself.

**Realistic simulated approximation:** render an `<iframe>` bound to the current URL with `sandbox` flags; maintain our own navigation stack for Go*/Can* navigation and `CurrentUrl`; fire `BeforePageLoad`/`PageLoaded` around URL changes we drive; treat `WebViewString` as a plain read/write string in host state with a same-origin `postMessage`/`srcdoc` bridge for cooperating pages only; show an honest in-frame notice ("This page cannot be embedded" or "Browser preview — not a full WebView") when load is blocked. Permission/SSL/cache/cookie properties are accepted-but-no-op and surfaced via `h.Unsupported` for the methods.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `HomeUrl` | `""` | Behavioral (also initial load) | Store in state; on init and on `GoHome`, set the iframe `src`. Empty default → blank frame with a "Set HomeUrl" placeholder. | High |
| `Visible` | `True` | Visual | Standard — controls whether the branch renders (shared `isSimulationVisible`). | High |
| `Width` / `Height` | `""` (auto) | Visual | `sizeStyle()` already handles `-1`/`-2`/percent encodings. A WebViewer with auto height is usually `Fill`/explicit; render with a sensible min-height fallback. | High |
| `Left` / `Top` | `""` | Visual (AbsoluteArrangement only) | Already supported via `positionStyle()`; add `Left`/`Top` to accepted props (they are already in `SIMULATION_VISUAL_PROPS`). | Low |
| `FollowLinks` | `True` | Behavioral | Stored; cannot truly intercept cross-origin link clicks. Honored only conceptually; document as caveat. | Medium |
| `WebViewString` | (block-only, default `""`) | Behavioral | Plain string in host state, read/write; same-origin bridge for cooperating pages. | Medium |
| `IgnoreSslErrors` | `False` | Behavioral | Cannot honor — browser governs TLS for iframe content. Accept + no-op. | Low |
| `PromptforPermission` | `True` | Behavioral | Cannot honor — browser owns permission prompts. Accept + no-op. | Low |
| `UsesCamera` | `False` | Designer-only (`rw: invisible`) | Permission flag, not readable/writable at runtime. Map to iframe `allow="camera"` at best; otherwise no-op. | Low |
| `UsesLocation` | `False` | Designer-only (`rw: invisible`) | As above → `allow="geolocation"`. | Low |
| `UsesMicrophone` | `False` | Designer-only (`rw: invisible`) | As above → `allow="microphone"`. | Low |
| `CurrentUrl` | (read-only) | Computed | Return the last URL we navigated to from the host nav stack. | Medium |
| `CurrentPageTitle` | (read-only) | Computed | Cannot read cross-origin `<title>`; return the URL or `""`. Document as caveat. | Low |
| `HeightPercent` / `WidthPercent` | (write-only) | Visual | Map to percent size encoding the existing `sizeStyle()` understands. | Low |

`Enabled` is **not** an AI WebViewer property (it does not derive from a focusable control with an Enabled flag in the same way); do not add it to defaults, mirroring the `Image` correction noted in the reviews.

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `BeforePageLoad` | `url` (text) | Fired from the host immediately before we set a new URL (on `GoToUrl`, `GoHome`, `Reload`, and the initial `HomeUrl` load). Cannot fire for user-initiated in-frame link clicks on cross-origin pages. |
| `PageLoaded` | `url` (text) | Fired when the iframe `load` event fires for a navigation we initiated, or (approximation) on a short timer after we set `src` if the cross-origin `load` event is suppressed. Cannot fire for in-page link navigations. |
| `ErrorOccurred` | `errorCode` (number), `description` (text), `failingUrl` (text) | Fired when the iframe fails to load. **Browser limitation:** an `<iframe>` blocked by `X-Frame-Options`/CSP does **not** emit a catchable `error` event cross-origin, so reliable detection is not possible; we can heuristically fire it (e.g. `load` never arrives within a timeout) with a synthetic `errorCode`. Mark as best-effort. |
| `WebViewStringChange` | `value` (text) | Fired only when same-origin/`srcdoc` page JS calls `window.AppInventor.setWebViewString(...)` and we receive it via `postMessage`. Cannot fire for arbitrary cross-origin pages. |

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `GoToUrl(url)` | `(text) → void` | Push `url` onto the host nav stack, fire `BeforePageLoad(url)`, emit a `navigate` effect → overlay sets iframe `src`; `PageLoaded` on load. |
| `GoHome()` | `() → void` | Navigate to current `HomeUrl` via the same path as `GoToUrl`. |
| `GoBack()` | `() → void` | Pop the host nav stack one entry (no-op if at start), navigate to the previous URL. Approximation — does not move within the page's own in-frame history. |
| `GoForward()` | `() → void` | Advance the host nav stack (no-op if at end). Same approximation. |
| `CanGoBack()` | `() → boolean` | Return `stackIndex > 0` from the host nav stack. |
| `CanGoForward()` | `() → boolean` | Return `stackIndex < stack.length - 1`. |
| `Reload()` | `() → void` | Re-emit a `navigate` effect for the current URL with a cache-busting token so the overlay forces an iframe reload; fire `BeforePageLoad`/`PageLoaded`. |
| `StopLoading()` | `() → void` | **Unsupported** — no way to abort an in-flight iframe load programmatically. `h.Unsupported`. |
| `RunJavaScript(js)` | `(text) → void` | Same-origin/`srcdoc` only: emit a `run-js` effect; overlay evals via `contentWindow`. Cross-origin → `h.Unsupported` (SOP). |
| `ClearCaches()` | `() → void` | **Unsupported** — JS cannot clear iframe cache. `h.Unsupported`. |
| `ClearCookies()` | `() → void` | **Unsupported** — cannot clear third-party iframe cookies. `h.Unsupported`. |
| `ClearLocations()` | `() → void` | **Unsupported** — no stored-location-permission API in browser. `h.Unsupported`. |

## Implementation Plan

### simulation-capabilities.js

- Add `'WebViewer'` to `SIMULATION_SUPPORTED_TYPES` (it is **not** non-visible, so do **not** add it to `SIMULATION_NONVISIBLE_TYPES`).
- Add a defaults block inside `buildSimulationDefaults()`:

```js
WebViewer: {
  Visible: true,
  Width: -1,            // Fill; WebViewer is rarely auto-sized
  Height: -1,
  HomeUrl: '',
  FollowLinks: true,
  IgnoreSslErrors: false,
  PromptforPermission: true,
  UsesCamera: false,
  UsesLocation: false,
  UsesMicrophone: false,
  WebViewString: '',
  CurrentUrl: '',
},
```

- Add new prop names to `SIMULATION_VISUAL_PROPS`: `HomeUrl`, `FollowLinks`, `IgnoreSslErrors`, `PromptforPermission`, `UsesCamera`, `UsesLocation`, `UsesMicrophone`, `WebViewString`, `CurrentUrl`. (`Width`, `Height`, `Left`, `Top`, `Visible` are already present.) Props not in this set are stripped before reaching the renderer, so the iframe wiring needs `HomeUrl`/`CurrentUrl` here.
- `isBooleanProp`: add `FollowLinks`, `IgnoreSslErrors`, `PromptforPermission`, `UsesCamera`, `UsesLocation`, `UsesMicrophone`.
- `isNumericProp`: nothing new required (`Width`/`Height`/`Left`/`Top` already covered; `HeightPercent`/`WidthPercent` are write-only and can be coerced as plain numbers).
- `coerceSimulationValue`: no special case needed — `HomeUrl`/`CurrentUrl`/`WebViewString` are plain strings handled by the default `return value`. No `deriveStateFromDesignerProps` branch is required (initial `CurrentUrl` can default to `HomeUrl` in the host or via a tiny derive: when `HomeUrl` is set, also seed `CurrentUrl`).

### SimulationComponent.svelte

Add a branch (after the `Image` branch, since it is the most similar asset/leaf pattern). Reuse `baseStyle()`/`containerStyle()` for sizing/position and `colorValue()` is not needed.

```svelte
{:else if node.type === 'WebViewer'}
  <div
    class="sim-webviewer"
    class:sim-unsupported={unsupportedHere}
    style={containerStyle('min-height: 120px;')}
    data-sim-component={node.name}
  >
    {#if webViewerUrl(props)}
      <iframe
        bind:this={webViewEl}
        title={node.name}
        src={webViewerUrl(props)}
        sandbox="allow-scripts allow-forms allow-same-origin allow-popups"
        referrerpolicy="no-referrer"
        on:load={() => emitEvent(node.name, 'PageLoaded', [webViewerUrl(props)])}
      ></iframe>
    {:else}
      <div class="sim-webviewer-empty">
        <span>WebViewer</span>
        <small>Set HomeUrl to preview a page</small>
      </div>
    {/if}
    <div class="sim-webviewer-note">Browser preview — not a full WebView</div>
  </div>
```

- Helper `webViewerUrl(props)` returns `props.CurrentUrl || props.HomeUrl || ''`, normalized to `https://` if scheme-less and non-empty.
- Wire the `navigate`/`reload`/`run-js` component-actions through the existing `handleComponentActions(actions?.[node.name] ?? {})` reactive path (same mechanism `open`/`focus` use). On a `navigate` action token bump, re-read `src` from state (state is patched by the host) — the simplest path is to let `CurrentUrl` in state drive `webViewerUrl`, so the host just patches `CurrentUrl` and the iframe `src` updates reactively. A `reload` action bumps a cache-bust suffix.
- DOM events: the iframe `load` event drives `PageLoaded`. There are no click/input events to wire on the host element itself (interaction happens inside the frame, which we cannot observe).
- CSS: `.sim-webviewer` is a positioned block with `overflow:hidden`; the `iframe` fills it (`width:100%;height:100%;border:0;`); `.sim-webviewer-empty` is a centered dashed placeholder; `.sim-webviewer-note` is a small absolutely-positioned badge. Add `webViewEl` to the `let` declarations near `buttonEl`.

### SimulateOverlay.svelte

Needed for the `navigate` / `reload` / `run-js` effects, but these can be modeled as `component-action` effects (like `open`/`focus`/`cursor-position`) that `applyEffects()` already turns into `actionTokens` — so no new dialog/toast handling is required. If `CurrentUrl` is patched directly into state by the host (recommended), the overlay needs **no change at all** for navigation; the existing `statePatch` flow updates the iframe `src` reactively. Only `RunJavaScript` against same-origin content would need a new effect handler; given it is largely `Unsupported`, treat overlay changes as **optional / none for v1**.

### simulation_wasm.go

A `WebViewer` type needs real handlers (the generic store path is insufficient for the nav stack, computed props, and method semantics):

- **State:** track a per-component nav stack and index (e.g. a `webViewNav map[string]*webViewState` on the host, or store the stack inside `h.state` as encoded values). Seed from `HomeUrl` at session init.
- **`GetProperty`:** add a `case "WebViewer"` for `CanGoBack`-style is handled in `CallMethod`, but for read-only properties return computed values: `CurrentUrl` (top of nav stack), `CurrentPageTitle` (return the URL or `""`, with a `h.logs` note that titles are unavailable).
- **`SetProperty`:** add a `case "WebViewer"` (or inline checks) so that setting `HomeUrl` also seeds/navigates `CurrentUrl` and fires `BeforePageLoad`; `WebViewString` set is a plain store write. Most others fall through to `h.setProperty`.
- **`CallMethod`:** add `case "WebViewer": return h.callWebViewerMethod(...)`. Implement `GoToUrl`/`GoHome`/`Reload` (push + patch `CurrentUrl` + `runEvent("BeforePageLoad", [url])`; the iframe `load` in the frontend fires `PageLoaded`), `GoBack`/`GoForward` (move the index + patch `CurrentUrl`), `CanGoBack`/`CanGoForward` (return `runtime.BoolVal(...)`). For `StopLoading`, `ClearCaches`, `ClearCookies`, `ClearLocations`, and cross-origin `RunJavaScript`: call `h.Unsupported("method", componentName+"."+method)`.
- **Events:** `runEvent` for `BeforePageLoad` (host-side, with the `url` arg) and `WebViewStringChange` (fired only if the same-origin bridge delivers a value). `PageLoaded` is fired frontend-side from the iframe `load` event; `ErrorOccurred` is best-effort frontend-side on load timeout.
- Recompile with `npm run build:wasm`.

### design-schema-tree.js

No change required. `WebViewer` is a leaf with no special containment: `canContainDesignComponent` already allows any non-special-parent container to hold non-container leaves, and `unsupportedSimulationComponents()` will stop flagging it once it is in `SIMULATION_SUPPORTED_TYPES`. `designTreeToInitialState()` will merge the new `SIMULATION_DEFAULTS.WebViewer` automatically.

## Dependencies & Ordering

- **No external libraries.** The render uses a native `<iframe>` — no Leaflet/MapLibre-style heavy dependency.
- **No prerequisite components.** WebViewer is a standalone leaf; it does not depend on Canvas/Map/Chart child plumbing or on any other component being implemented first. It can be implemented independently.

## Web-Platform Limitations & Fidelity Caveats

- **Most real-world URLs will not embed.** `X-Frame-Options`/CSP `frame-ancestors` block embedding for the majority of popular sites; the frame renders blank and the simulator cannot detect it reliably. A device `WebView` has no such restriction.
- **History is faked.** `GoBack`/`GoForward`/`CanGoBack`/`CanGoForward` track only the URLs the *app* navigated to via blocks, not links the user clicked inside the page. Cross-origin `history` access throws.
- **`CurrentUrl` reflects the last app-set URL, not in-frame navigations.** `CurrentPageTitle` is essentially unavailable (cross-origin `document.title` access is blocked) and will return the URL or empty.
- **`WebViewString` bridge works only for same-origin / `srcdoc` content.** Arbitrary third-party pages cannot reach `window.AppInventor`, so `WebViewStringChange` will not fire for them and `RunJavaScript` against them is blocked by the same-origin policy.
- **`IgnoreSslErrors`, `PromptforPermission`, `ClearCaches/Cookies/Locations`, `StopLoading`, `UsesCamera/Location/Microphone` are no-ops** — TLS, cache, cookies, and permission prompts for iframe content are governed by the browser, not the app.
- **`ErrorOccurred` is best-effort** (timeout heuristic), since cross-origin embed failures do not emit a catchable `error` event.

## Effort Estimate

**M.** Frontend iframe branch + URL helper + CSS is small (~Image-sized); the host work (nav stack, computed `CurrentUrl`, `Go*`/`Can*` methods, `BeforePageLoad`/`PageLoaded` wiring, and `h.Unsupported` for the non-feasible methods) plus a `npm run build:wasm` cycle is the bulk. No new overlay dialogs and no external dependency keep it out of L territory; the breadth of honestly-unsupported surface (and the same-origin bridge if attempted) keeps it above S.
