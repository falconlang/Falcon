# Screen Simulator Review

## Overview

The `Screen` component (also mapped as `Form` in App Inventor's Java source `Form.java`) is the root container for every App Inventor application. It hosts all child components, manages lifecycle events (Initialize, BackPressed, OtherScreenClosed, etc.), controls screen orientation, scroll behaviour, background, title bar, alignment, and multi-screen navigation.

In the Tensor IDE simulator the Screen is rendered as a plain `<div class="sim-screen">` in `SimulationComponent.svelte`. Default state is set in `simulation-capabilities.js` and runtime method calls are dispatched through the Go WASM host in `simulation_wasm.go`.

---

## Properties Analysis

### Supported

- `BackgroundColor` — read from state, applied via `colorValue()` as CSS `background`. Accepts `&HAARRGGBB` notation.
- `Width` — interpreted via `sizeStyle()`.
- `Height` — interpreted via `sizeStyle()`.
- `Visible` — controlled through `isSimulationVisible()`.
- `Title` — listed in `SIMULATION_VISUAL_PROPS`; stored in state but **not rendered** in any simulated title bar (see Bugs section).

### Missing / Unsupported

| Property | AI Default | Priority |
|---|---|---|
| `AlignHorizontal` | 1 (left) | **High** — affects child layout; entire flex alignment is hardcoded to `align-items: stretch` |
| `AlignVertical` | 1 (top) | **High** — affects child layout; no CSS justify-content mapping |
| `Scrollable` | false | **High** — screen is always scrollable (`overflow: auto`); a non-scrollable screen should clip at device height |
| `BackgroundImage` | "" | **Medium** — no `BackgroundImage` property tracked or applied |
| `ScreenOrientation` | "unspecified" | **Medium** — no orientation simulation (landscape / portrait toggle) |
| `ShowStatusBar` | true | **Low** — cosmetic only in simulation context but affects simulated chrome height |
| `TitleVisible` | true | **Low** — property stored but title bar not rendered at all |
| `AboutScreen` | "" | **Low** — read-only informational; acceptable to omit |
| `ActionBar` | false | **Low** — affects title bar style; acceptable to omit in simulator |
| `PrimaryColor` | `#FF3F51B5` | **Low** — theme colour; not used |
| `PrimaryColorDark` | `#FF303F9F` | **Low** — theme colour; not used |
| `AccentColor` | `#FFFF4081` | **Low** — theme colour; not used |
| `HighContrast` | false | **Low** — accessibility; cascades to children |
| `BigDefaultText` | false | **Low** — accessibility; cascades to children |
| `Sizing` | "Responsive" | **Low** — fixed vs. responsive layout scaling; not simulated |
| `OpenScreenAnimation` | "default" | **Low** — screen transition animation; acceptable to omit |
| `CloseScreenAnimation` | "default" | **Low** — screen transition animation; acceptable to omit |
| `Platform` (read-only) | "Android" | **Medium** — blocks that read `Screen1.Platform` will return null/unsupported instead of "Android" |
| `PlatformVersion` (read-only) | OS version string | **Low** — blocks that read this will get null |

### Wrong Defaults or Types

| Property | Tensor Default | Correct AI Default | Priority |
|---|---|---|---|
| `BackgroundColor` | `'&HFFFFFFFF'` (white) | `Component.COLOR_DEFAULT` which maps to `COLOR_WHITE` (`0xFFFFFFFF`) in Classic theme | **Low** — values match for white background but `COLOR_DEFAULT` (`0x00000000`) has special meaning; setting `BackgroundColor` to `COLOR_DEFAULT` in AI means "use theme default" not pure white |
| `Height` | `-2` (fill parent) | Screen height is a fixed measured pixel value (not `-2`); in AI, `Screen.Height` is read-only and returns the actual device height in dp | **High** — the simulator exports `-2` as the Screen's Height property. When a block reads `Screen1.Height` expecting a pixel value (e.g., 640) it gets -2, breaking layout arithmetic |
| `Width` | `-1` (automatic) | `Screen.Width` is read-only actual pixel width | **High** — same issue as Height; blocks reading `Screen1.Width` get -1 instead of a concrete pixel count |

---

## Events Analysis

### Supported

- `Initialize` — fired at session creation via `createSimulationSession` in `simulation_wasm.go` (lines 584–597). Correct.
- `BackPressed` — listed in AI source as a `@SimpleEvent`; no explicit wiring in Svelte but can be dispatched manually through `dispatchSimulationEvent`.

### Missing / Incorrect

| Event | Signature (AI) | Priority | Notes |
|---|---|---|---|
| `OtherScreenClosed` | `(otherScreenName: String, result: Object)` | **High** | Multi-screen navigation is unsupported; closing a screen never fires this event on the caller |
| `ScreenOrientationChanged` | `()` | **Medium** | Device orientation changes are not simulated |
| `ErrorOccurred` | `(component: Component, functionName: String, errorNumber: int, message: String)` | **High** | Runtime errors in the Go interpreter surface as WASM-level failures; `ErrorOccurred` is never dispatched to user blocks. The AI spec allows user blocks to override the default error toast via this event |
| `PermissionDenied` | `(component: Component, functionName: String, permissionName: String)` | **Low** | Permission model not applicable in web simulator |
| `PermissionGranted` | `(permissionName: String)` | **Low** | Permission model not applicable in web simulator |

---

## Methods Analysis

### Supported

None of the Screen's `@SimpleFunction` methods are explicitly handled by the Go WASM host. The host's `CallMethod` dispatcher does not have a `case "Screen":` branch at all.

### Missing / Incorrect

| Method | Signature (AI) | Priority | Notes |
|---|---|---|---|
| `AskForPermission` | `(permissionName: String)` | **Low** | Not applicable in web context |
| `HideKeyboard` (implicit via `InputMethodManager`) | none in blocks | **Low** | Not applicable |

**Screen-level block operations that are missing (these are block primitives, not `@SimpleFunction` but exposed to blocks):**

| Block | Priority | Notes |
|---|---|---|
| `open another screen` / `switchForm` | **Critical** | Calling this in simulation throws unsupported; multi-screen apps cannot run |
| `open another screen with start value` | **Critical** | Same — unsupported |
| `close screen` / `finishActivity` | **High** | No-op or unsupported; blocks that close a screen silently fail |
| `close screen with value` | **High** | No-op |
| `close application` / `finishApplication` | **High** | No-op |
| `get plain start text` / `getStartText` | **High** | Always returns null/empty; apps that branch on startup text break |
| `get start value` / `getStartValue` | **High** | Same |

---

## Behaviour Gaps

### 1. Screen Height / Width return wrong sentinel values (Critical)

In AI, `Screen1.Width` and `Screen1.Height` are read-only properties that return the **actual device pixel dimensions** (e.g., 320 × 568 for a small phone). Tensor defaults both to `-1` and `-2` respectively (the "automatic" and "fill parent" sentinels used for *child* sizing). Any block that does arithmetic on `Screen1.Height` — a common pattern for placing components relative to the bottom of the screen — will receive these sentinel values and produce incorrect results.

### 2. `Scrollable` property not honoured (High)

The `.sim-screen` CSS class has `overflow: auto` hardcoded. In AI, when `Scrollable = false` the screen is not scrollable and content is clipped at the device height. This difference means that a non-scrollable screen that should overflow and clip its content appears to scroll freely in Tensor's simulator.

### 3. AlignHorizontal / AlignVertical not implemented (High)

The screen's flex container is hardcoded to `align-items: stretch` with no `justify-content` or alignment adjustment based on `AlignHorizontal` / `AlignVertical` props. AI supports left (1), center (3), right (2) for horizontal and top (1), center (2), bottom (3) for vertical. A screen designed with center-aligned content will render left-aligned in simulation.

### 4. Multi-screen navigation completely absent (Critical)

`switchForm`, `switchFormWithStartValue`, `finishActivity`, `finishActivityWithResult`, `getStartText`, `getStartValue` are all unimplemented. These are not obscure edge cases — multi-screen navigation is a core App Inventor concept taught in most intermediate tutorials. Any app with more than one screen cannot be simulated.

### 5. BackgroundColor `COLOR_DEFAULT` vs. explicit white (Medium)

AI's `BackgroundColor(Component.COLOR_DEFAULT)` sets the background to the *theme's* default (white in Classic, dark in Dark theme). Tensor stores `&HFFFFFFFF` as the default, which is visually identical for the Classic theme but means `BackgroundColor` cannot be reset to the theme default via a block — it always forces explicit white.

### 6. Title property stored but not rendered (Medium)

`Title` is included in `SIMULATION_VISUAL_PROPS` and tracked in state, but the `sim-screen` div has no title bar element. The AI specification shows the title string in a visible title bar at the top of the screen. Designers who rely on the title for screen identification have no visual feedback in simulation.

### 7. `Platform` and `PlatformVersion` read-only properties return null (Medium)

Blocks that read `Screen1.Platform` expect the string `"Android"`. Because `Platform` is not in `SIMULATION_DEFAULTS` or `SIMULATION_VISUAL_PROPS`, `GetProperty` returns `NullVal()`. This breaks apps that conditionally branch on platform.

### 8. `ErrorOccurred` event never dispatched (High)

When a runtime error occurs in the Go interpreter, `simulation_wasm.go` propagates it back to JavaScript as a top-level `simulationResult(false, ...)`. The AI specification fires the `ErrorOccurred` event on the Screen and only falls back to a toast if no handler is registered. Tensor never gives user blocks a chance to handle errors gracefully.

### 9. Alignment encoding mismatch in description (Low)

The AI comment in `AlignHorizontal` says: "1 = left, **3** = horizontally centered, 2 = right". This non-sequential mapping (1, 3, 2) must be preserved if `AlignHorizontal` is ever implemented; using sequential 1/2/3 for left/center/right would be incorrect.

### 10. SIZE_ENCODING: -1000x percent format partially implemented (Low)

`sizeStyle()` in `SimulationComponent.svelte` correctly handles the `-1000 - percentage` encoding (negative values ≤ -1000 are treated as `-(n+1000)%`). This is correctly implemented.

---

## Bugs Found

### Bug 1: `Screen.Height` defaults to `-2` (fill-parent sentinel) — read always returns -2

**File:** `simulation-capabilities.js`, line 39

```js
Screen: { ..., Height: -2, ... }
```

`Height: -2` is the child-component sentinel meaning "fill parent". For the Screen component itself, Height should be the actual simulated viewport height (e.g., 640). Any block expression like `Screen1.Height - 100` evaluates to `-102` instead of the intended value.

**Priority: High**

---

### Bug 2: `Screen.Width` defaults to `-1` (automatic sentinel) — read always returns -1

**File:** `simulation-capabilities.js`, line 39

```js
Screen: { ..., Width: -1, ... }
```

Same class of issue. `-1` means "automatic" width for a child component. For Screen itself, this should be a fixed pixel value (e.g., 320 or the simulated container width).

**Priority: High**

---

### Bug 3: Screen is always scrollable regardless of `Scrollable` property

**File:** `SimulationComponent.svelte`, line 362

```css
.sim-screen {
  overflow: auto;
  ...
}
```

The `Scrollable` property from state is never consulted. The screen will always scroll.

**Priority: High**

---

### Bug 4: `open another screen` / multi-screen blocks cause silent unsupported error

**File:** `simulation_wasm.go` — `CallMethod` has no Screen/Form case.

Calling `switchForm` or any multi-screen block via the interpreter results in `h.Unsupported("method", "Screen.switchForm")` being appended to the unsupported list. The app continues running (partially) without navigating. The user sees no informative error message in the simulation UI explaining that multi-screen navigation is not supported.

**Priority: Critical**

---

### Bug 5: `Initialize` event lookup falls back incorrectly for named screens

**File:** `simulation_wasm.go`, lines 584–597

```go
screenName := "Screen1"
if names := componentContextMap["Screen"]; len(names) > 0 {
    screenName = names[0]
}
if event := session.events[simulationEventKey{component: screenName, event: "Initialize"}]; event != nil {
    interp.RunBody(event.Body)
} else {
    for key, event := range session.events {
        if key.event == "Initialize" && (event.ComponentType == "Screen" || event.ComponentType == "Form") {
            interp.RunBody(event.Body)
            break
        }
    }
}
```

The fallback `for` loop iterates a Go `map` which has **non-deterministic iteration order**. In a project with multiple screens (if/when supported), the wrong screen's `Initialize` event could fire. For single-screen projects this is benign but the pattern is fragile.

**Priority: Medium**

---

### Bug 6: `colorValue()` silently drops the alpha channel

**File:** `SimulationComponent.svelte`, lines 43–47

```js
function colorValue(value, fallback = 'transparent') {
  const text = String(value ?? '').trim();
  const ai = text.match(/^&H([0-9A-Fa-f]{2})([0-9A-Fa-f]{6})$/);
  if (ai) return `#${ai[2]}`;   // ai[1] is the alpha byte — discarded
  ...
}
```

The AI color format is `&HAARRGGBB`. The alpha byte (`ai[1]`) is discarded and the returned CSS colour has no opacity. For fully opaque colours (`&HFF...`) this is harmless, but the AI spec allows semi-transparent backgrounds (e.g., `&H80FFFFFF`). This affects `BackgroundColor` on the Screen and all components.

**Priority: Medium**

---

### Bug 7: `Title` property has no visual effect

**File:** `SimulationComponent.svelte`, lines 224–236

The Screen `<div>` renders only its children. There is no rendered title bar element, so changes to the `Title` property (e.g., `set Screen1.Title to "My App"`) produce no visible effect in simulation even though the state is correctly updated.

**Priority: Medium**

---

## Android/App Inventor Standards Compliance

| Standard | Status |
|---|---|
| `&HAARRGGBB` color format | Partially — alpha byte ignored (Bug 6) |
| Size constants: -1 = automatic, -2 = fill parent, -(n+1000) = n% | Correctly handled for child components; incorrectly applied to Screen itself (Bugs 1–2) |
| Screen as vertical LinearLayout | Correct — `flex-direction: column` |
| Screen is the root container with full width | Correct — `width: 100%` |
| `Initialize` fires once at screen start | Correct |
| `BackPressed` event can cancel the default back action | Not simulated — no browser "back" intercept |
| AlignHorizontal values are 1/3/2 (not 1/2/3) | Not implemented; note non-sequential mapping must be respected if added |
| `Screen.Height` / `Screen.Width` return device dp dimensions | Non-compliant — returns sentinels -2 and -1 |
| `Scrollable = false` clips content at device height | Non-compliant — always scrollable |
| Multi-screen: `open screen`, `close screen`, `get start value` | Not implemented |
| `ErrorOccurred` event dispatched on runtime errors | Not implemented — errors surface only as WASM failures |
| `Platform` returns "Android" | Not implemented — returns null |

---

## Summary

The Tensor IDE Screen simulator correctly handles the basic rendering use-case: a vertically-stacked container with a configurable background colour that hosts child components and fires the `Initialize` event at startup. For simple single-screen apps this is sufficient.

However, there are significant compliance gaps that affect a large class of apps:

**Top 3 action items:**

1. **Fix `Screen.Height` and `Screen.Width` to return real pixel dimensions** (e.g., the simulated viewport height/width in px). These are read-only block-accessible properties that are widely used for responsive layout arithmetic. Currently they return the child-component sentinels `-2` and `-1`, silently corrupting any arithmetic.

2. **Implement `AlignHorizontal` and `AlignVertical`** by mapping the stored property values to CSS `justify-content` and `align-items` on the `.sim-screen` container, and **honour the `Scrollable` property** by conditionally applying `overflow: auto` vs `overflow: hidden`.

3. **Surface a user-visible "multi-screen not supported" notice** and **dispatch `ErrorOccurred`** to user blocks before falling back to a WASM-level error. Multi-screen navigation (`open another screen`, `close screen`, `get start value`) is entirely absent and fails silently, which is confusing during simulation of any intermediate-level AI project.
