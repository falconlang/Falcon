# HorizontalArrangement Simulator Review

## Overview

`HorizontalArrangement` is a layout container that arranges its child components from left to right in a single horizontal row. It extends `HVArrangement`, which itself extends `AndroidViewComponent` and also acts as a `ComponentContainer`. The component carries no events or methods of its own — its entire public API consists of inherited properties from `HVArrangement` plus the standard `AndroidViewComponent` width/height/visible interface.

In Tensor IDE, `HorizontalArrangement` is rendered as a flex-row `<div>` in `SimulationComponent.svelte`, with defaults declared in `simulation-capabilities.js`. The backend Go runtime (`simulation_wasm.go`) handles generic property get/set but has no HorizontalArrangement-specific method dispatch.

---

## Properties Analysis

### Supported

- `Visible` — controls whether the element is rendered at all (checked via `isSimulationVisible`)
- `Width` — size constants `-1` (Automatic/preferred), `-2` (Fill Parent), and percentage encoding `<= -1000` all handled in `sizeStyle()`
- `Height` — same size constant handling as `Width`
- `BackgroundColor` — applied via `baseStyle()` → `colorValue()`; the `&H` ARGB format is parsed and the RGB portion is applied as a CSS hex color

### Missing / Unsupported

| Property | Priority | Notes |
|---|---|---|
| `AlignHorizontal` | **Critical** | Controls how child components are aligned along the horizontal axis (1=Left, 2=Right, 3=Center). Default is `1` (Left) per `ComponentConstants.HORIZONTAL_ALIGNMENT_DEFAULT = GRAVITY_LEFT`. Not present in `SIMULATION_DEFAULTS`, not applied in CSS. The flex container always uses browser default (stretch/start). |
| `AlignVertical` | **Critical** | Controls vertical alignment of children within the row (1=Top, 2=Center, 3=Bottom). Default is `1` (Top) per `ComponentConstants.VERTICAL_ALIGNMENT_DEFAULT = GRAVITY_TOP`. Not present in `SIMULATION_DEFAULTS`, not applied in CSS. The CSS rule `align-items: center` is hardcoded in `.sim-arrangement--horizontal`, which incorrectly defaults to vertical center rather than the AI default of Top. |
| `Image` | **High** | `HVArrangement.Image()` — sets a background image that takes precedence over `BackgroundColor`. Not present in `SIMULATION_DEFAULTS`, not tracked in `SIMULATION_VISUAL_PROPS`, and not rendered. |
| `Enabled` | **Medium** | Listed in `SIMULATION_DEFAULTS` but `HVArrangement` (and `HorizontalArrangement`) does not expose an `Enabled` property in the AI source. `AndroidViewComponent` does not have `Enabled`. No AI event or property depends on it for arrangements. Including it wastes state and could cause confusion. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `Width` | `-1` (Automatic / `LENGTH_PREFERRED`) | `-2` (Fill Parent) | **High** | `simulation-capabilities.js` line 54: `HorizontalArrangement: { ..., Width: -2, ... }`. In App Inventor a new `HorizontalArrangement` starts with `Automatic` width (`-1`). Using `-2` (Fill Parent) causes the arrangement to always span the full container width even when it should wrap its children. |
| `Height` | `-1` (Automatic / `LENGTH_PREFERRED`) | `-2` (Fill Parent) | **Medium** | `simulation-capabilities.js` line 54: `HorizontalArrangement: { ..., Height: -2, ... }`. The AI default is also `Automatic` (`-1`). Using `-2` inflates the arrangement's height to fill its parent. Note: the AI docs say empty-arrangement height falls back to 100 px for Automatic; a `-1` default with a CSS `min-height` fallback would be more faithful. |
| `BackgroundColor` | `&H00000000` (COLOR_DEFAULT — transparent / system default) | _(not set — undefined in defaults)_ | **Medium** | `HVArrangement` calls `BackgroundColor(Component.COLOR_DEFAULT)` in its constructor. The Tensor defaults object for `HorizontalArrangement` omits `BackgroundColor` entirely, so the arrangement renders with no background rather than the transparent/system-default appearance. |
| `AlignVertical` CSS hardcode | Top (1) | Center (CSS `align-items: center`) | **Critical** | Even if `AlignVertical` were surfaced, the `.sim-arrangement--horizontal` CSS class unconditionally sets `align-items: center`. This diverges from the AI default of Top alignment and cannot be corrected at runtime without removing the class-level rule. |

---

## Events Analysis

### Supported

_(none — HorizontalArrangement defines no events in the AI source)_

### Missing / Incorrect

_(none — HorizontalArrangement has no events; this section is not applicable)_

---

## Methods Analysis

### Supported

_(none — HorizontalArrangement defines no callable methods in the AI source)_

### Missing / Incorrect

_(none — HorizontalArrangement has no methods; this section is not applicable)_

---

## Behaviour Gaps

### 1. Horizontal alignment not simulated (Critical)

App Inventor's `AlignHorizontal` maps to Android `LinearLayout` gravity along the main axis. Values are:
- `1` = Left (default)
- `2` = Right
- `3` = Horizontally Centered

None of these are modelled. The flex container in CSS has no `justify-content` rule, so children are always packed to the left (browser default for `flex-direction: row`). While this happens to match the AI default alignment, it cannot be changed at runtime by a block that sets `AlignHorizontal`.

### 2. Vertical alignment hardcoded wrong (Critical)

The CSS rule for `.sim-arrangement--horizontal` sets `align-items: center`. The AI default is `AlignVertical = 1` (Top), which maps to `align-items: flex-start`. The hardcoded value silently overrides what the AI default should be. Any block setting `AlignVertical` to `1` (Top) would have no visual effect because:
- `AlignVertical` is not in `SIMULATION_VISUAL_PROPS`
- No CSS is applied from it
- The hardcoded `align-items: center` always wins

### 3. Default Width should be Automatic (-1), not Fill Parent (-2) (High)

In a real App Inventor project, a freshly placed `HorizontalArrangement` has `Width = Automatic`. The simulator defaults it to `-2` (Fill Parent). This means every simulated `HorizontalArrangement` spans the full width of its parent even when the designer has not explicitly set Fill Parent, which visually misrepresents the layout.

### 4. `flex-wrap: wrap` on `.sim-arrangement--horizontal` (Medium)

The CSS for `.sim-arrangement--horizontal` includes `flex-wrap: wrap`. App Inventor's `HorizontalArrangement` uses a `LinearLayout` with `LAYOUT_ORIENTATION_HORIZONTAL` — children that overflow the available width are clipped or extend the scroll area (for `HorizontalScrollArrangement`). They do not wrap to a new line. This makes the simulator incorrectly show children wrapping when the total child width exceeds the container, instead of clipping or scrolling.

### 5. Background image not supported (High)

`HVArrangement.Image()` accepts a path and displays it as the arrangement's background, taking visual precedence over `BackgroundColor`. This is commonly used in UI layouts. The simulator has no mechanism to resolve or render a background image on layout containers. The `assetUrl` reactive variable in `SimulationComponent.svelte` only reads `props.Picture || props.Image`, but even if it resolved a value there is no `<img>` or CSS `background-image` applied to the arrangement's `<div>`.

### 6. Alpha channel in BackgroundColor is ignored (Low)

`colorValue()` in `SimulationComponent.svelte` strips the alpha channel from `&HAABBGGRR` values: it uses `#${ai[2]}` (the 6-digit RGB portion), discarding the leading two-digit alpha (`ai[1]`). For semi-transparent arrangement backgrounds, this will render the arrangement as fully opaque. App Inventor applies the full ARGB value including transparency.

### 7. `Enabled` property is spurious (Low)

`HVArrangement` does not extend `ToggleBase` or any component that exposes `Enabled`. No AI block can read or set `HorizontalArrangement.Enabled`. Including it in `SIMULATION_DEFAULTS` causes a phantom property to appear in runtime state that has no correspondence in the real platform.

### 8. Size constant -1 (Automatic) and empty-arrangement 100 px fallback (Low)

The AI documentation for `HorizontalArrangement` states: "If a `HorizontalArrangement`'s `Height` property is set to `Automatic` and it is empty, the `Height` will be 100." The simulator renders an empty arrangement with no minimum height when `Width`/`Height` are `-1`, which could result in a zero-height invisible box rather than the 100 px fallback.

---

## Bugs Found

### Bug 1 — `AlignVertical` hardcoded to Center instead of Top
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 375–378
```css
.sim-arrangement--horizontal {
  flex-direction: row;
  align-items: center;   /* <-- BUG: should be flex-start (Top) to match AI default */
  flex-wrap: wrap;
}
```
The AI default `AlignVertical = 1` (Top) means children should align at the top of the row. `align-items: center` is incorrect and cannot be overridden at runtime because the property is not surfaced.

### Bug 2 — Wrong default Width (-2 Fill Parent instead of -1 Automatic)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 54
```js
HorizontalArrangement: { Visible: true, Enabled: true, Width: -2, Height: -2 },
```
`Width: -2` causes every simulated `HorizontalArrangement` to fill its parent width unconditionally. The correct default is `Width: -1` (Automatic).

### Bug 3 — `flex-wrap: wrap` incorrectly wraps children
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 378
`flex-wrap: wrap` causes children to wrap to a second line when the total width is exceeded. App Inventor's `LinearLayout` clips/overflows rather than wrapping. This should be `flex-wrap: nowrap`.

### Bug 4 — `AlignHorizontal` and `AlignVertical` missing from `SIMULATION_VISUAL_PROPS`
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, lines 63–100
Neither `AlignHorizontal` nor `AlignVertical` appear in `SIMULATION_VISUAL_PROPS`. This means property-set calls from blocks at runtime (e.g. `set HorizontalArrangement1.AlignHorizontal to 3`) will never be forwarded to the frontend renderer. Even if the CSS were updated to use CSS variables driven by these props, they would still not be applied because the reactive `props` object would not be updated.

### Bug 5 — `Image` property not in `SIMULATION_VISUAL_PROPS`
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, lines 63–100
`Image` (background image path for the arrangement) is not tracked, so setting it from a block at runtime would have no visual effect.

---

## Android/App Inventor Standards Compliance

| Aspect | Status |
|---|---|
| Horizontal flex layout | Partially correct — children are laid out left-to-right, but wrapping and overflow behaviour diverge from Android `LinearLayout` |
| Default AlignHorizontal = Left (1) | Accidentally correct (browser `justify-content` default is `flex-start`) but not expressly modelled |
| Default AlignVertical = Top (1) | **Non-compliant** — hardcoded to `align-items: center` |
| Default Width = Automatic (-1) | **Non-compliant** — set to `-2` (Fill Parent) |
| Default Height = Automatic (-1) | **Non-compliant** — set to `-2` (Fill Parent) |
| Background color default = COLOR_DEFAULT (transparent) | **Non-compliant** — not set in defaults |
| Background image support | **Missing** |
| Alpha channel in BackgroundColor | **Non-compliant** — alpha stripped by `colorValue()` |
| No events or methods | Correct — HorizontalArrangement has none |
| Children fill-parent width behaviour | Not modelled — fill-parent width distribution among siblings in a horizontal row is not simulated |
| Percent-encoded size (≤ -1000) | Correctly decoded in `sizeStyle()` |

---

## Summary

The `HorizontalArrangement` simulator implementation establishes a basic horizontal flex container that renders children left-to-right, which is the correct structural behaviour. However, the implementation has several concrete divergences from the App Inventor specification that affect visual fidelity.

**Top 3 action items:**

1. **Fix `align-items: center` → `align-items: flex-start` in `.sim-arrangement--horizontal` CSS, and surface `AlignVertical` and `AlignHorizontal` as reactive properties.** This is the most visible correctness bug: the AI default is Top alignment, not Center. Without surfacing both alignment properties in `SIMULATION_VISUAL_PROPS` and mapping them to CSS `align-items` / `justify-content`, blocks that change alignment will silently have no effect.

2. **Change the `Width` default from `-2` to `-1` (Automatic).** Every new `HorizontalArrangement` currently takes up the full width of its parent in the simulator, which misrepresents the designer's actual layout. Correcting the default to `-1` and adding a CSS `min-height` fallback of 100 px for empty arrangements would make the simulator match the AI specification.

3. **Remove `flex-wrap: wrap` and add `Image` / `BackgroundColor` support.** Children must not wrap to a new line — App Inventor clips or scrolls instead. Additionally, the `Image` background property and the alpha channel in `BackgroundColor` should be applied so that visual styling set by a block (or the designer) is faithfully reflected in the simulation.
