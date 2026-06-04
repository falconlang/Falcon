# HorizontalScrollArrangement Simulator Review

## Overview

`HorizontalScrollArrangement` is a layout container in App Inventor that places its children side-by-side from left to right inside a horizontally scrollable `HorizontalScrollView`. It extends `HVArrangement` which provides all of its properties (alignment, background color, background image) and layout logic.

In the Tensor IDE simulator the component is rendered as a `<div class="sim-arrangement sim-arrangement--horizontal sim-scroll-horizontal">` in `SimulationComponent.svelte`. Default state is minimal (only `Visible`, `Enabled`, `Width: -1`, `Height: -2`) and is defined in `simulation-capabilities.js`. No method calls or events are dispatched for this component type in `simulation_wasm.go`.

---

## Properties Analysis

### Supported

- `Visible` — controlled through `isSimulationVisible()`. Hides the div when false.
- `Enabled` — controlled through `isSimulationEnabled()`. Tracked in state; no visual effect on a container (correct, since the arrangement itself does not have an interactable role).
- `Width` — interpreted via `sizeStyle()`, supporting `-1` (automatic), `-2` (fill parent), pixel values, and the `-1000 - n%` percent encoding.
- `Height` — interpreted via `sizeStyle()` with the same encoding.
- `BackgroundColor` — read from state and applied via `colorValue()` as CSS `background` on the div. Accepts `&HAARRGGBB` notation (alpha silently dropped — see Bugs).

### Missing / Unsupported

| Property | AI Default | Priority | Notes |
|---|---|---|---|
| `AlignHorizontal` | 1 (Left) | **High** | Not tracked in defaults or `SIMULATION_VISUAL_PROPS`; the flex container is hardcoded to `align-items: center` with no dynamic mapping from the stored value |
| `AlignVertical` | 1 (Top) | **High** | Same — no dynamic `justify-content` or `align-items` mapping from stored property |
| `Image` (BackgroundImage) | `""` | **Medium** | `HVArrangement.Image` is a `@DesignerProperty` / `@SimpleProperty`. Neither tracked in `SIMULATION_DEFAULTS` nor `SIMULATION_VISUAL_PROPS` for this type; setting it via a block has no visual effect |

### Wrong Defaults or Types

| Property | Tensor Default | Correct AI Default | Priority |
|---|---|---|---|
| `Width` | `-1` (automatic) | `ComponentConstants.EMPTY_HV_ARRANGEMENT_WIDTH` which is `LENGTH_FILL_PARENT` (`-2`) for a scrollable arrangement — in practice the outer scroll view should fill the parent by default | **Medium** — the current default of `-1` (automatic, wraps content) differs from AI behaviour where the scroll container fills available horizontal width. Visually, `overflow-x: auto` on a narrow auto-width div may not show scroll indicators where expected. Note: AI actually defaults `HorizontalScrollArrangement` with `EMPTY_HV_ARRANGEMENT_WIDTH` which is `LENGTH_FILL_PARENT` in the component constants |
| `Height` | `-2` (fill parent) | `ComponentConstants.EMPTY_HV_ARRANGEMENT_HEIGHT` which evaluates to `LENGTH_FILL_PARENT` (`-2`) | **Low** — matches |
| `BackgroundColor` | Not set (no entry in defaults) | `Component.COLOR_DEFAULT` (which is `0x00000000` — transparent) | **Medium** — because `BackgroundColor` is absent from `SIMULATION_DEFAULTS` for `HorizontalScrollArrangement`, `props.BackgroundColor` is `undefined`. The `baseStyle()` function guards with `props.BackgroundColor ?` so the background is effectively transparent (browser default), which is acceptable, but a block that reads `HorizontalScrollArrangement1.BackgroundColor` will receive `NullVal()` instead of the AI's `COLOR_DEFAULT` value |

---

## Events Analysis

### Supported

None. App Inventor's `HorizontalScrollArrangement` (via `HVArrangement`) does not define any `@SimpleEvent` annotations. There are no user-exposed events for this component, so this is **correct**.

### Missing / Incorrect

No events are missing. The component specification has no events.

---

## Methods Analysis

### Supported

None. `HVArrangement` and `HorizontalScrollArrangement` define no `@SimpleFunction` methods. The `CallMethod` path in `simulation_wasm.go` will fall through to `h.Unsupported(...)` if a method is ever called on this type, which is the correct fallback.

### Missing / Incorrect

No public block-accessible methods exist in the AI source for this component, so there is nothing to implement. This is correct.

---

## Behaviour Gaps

### 1. AlignHorizontal not implemented (High)

The AI source (`HVArrangement.java`, lines 271–283) exposes `AlignHorizontal` as a `@DesignerProperty` and `@SimpleProperty` with default `ComponentConstants.HORIZONTAL_ALIGNMENT_DEFAULT` (value `1` = Left). The property accepts values 1 (Left), 2 (Right), 3 (Center).

In Tensor:
- `AlignHorizontal` is absent from `SIMULATION_DEFAULTS` for `HorizontalScrollArrangement`.
- It is absent from `SIMULATION_VISUAL_PROPS`, so changes made via blocks are not tracked or reflected.
- The CSS class `.sim-arrangement--horizontal` hardcodes `align-items: center`, which corresponds to vertical centering, not horizontal content alignment.
- There is no `justify-content` CSS mapping driven by the property value.

Result: Any arrangement with explicit horizontal alignment (e.g., right-aligned children) will always appear left-aligned in simulation.

### 2. AlignVertical not implemented (High)

Same situation as `AlignHorizontal`. The AI default is `ComponentConstants.VERTICAL_ALIGNMENT_DEFAULT` (value `1` = Top). The CSS class hardcodes `align-items: center` which centres children vertically — this is actually the wrong default (the AI default is Top = stretch/start, not center).

The `.sim-arrangement--horizontal` CSS sets `align-items: center` which simulates `AlignVertical = 2` (center) instead of the correct default `AlignVertical = 1` (top/start).

### 3. Background Image property not tracked (Medium)

`HVArrangement.Image` (path to a background asset) is a fully supported property in AI. When an image is set it takes visual precedence over `BackgroundColor`. In Tensor neither the `Image` (or `BackgroundImage`) property is tracked in `SIMULATION_VISUAL_PROPS` for arrangement types, nor is there any CSS `background-image` rendering logic in the arrangement branch of `SimulationComponent.svelte`. The `assetUrl` reactive variable is derived only for components that explicitly use `props.Picture || props.Image`, which the arrangement block does not.

### 4. Default Width should be fill-parent (-2) not automatic (-1) (Medium)

In App Inventor the `HorizontalScrollArrangement` frame container is initialised with `ComponentConstants.EMPTY_HV_ARRANGEMENT_WIDTH` which resolves to `LENGTH_FILL_PARENT` (`-2`). The Tensor default of `Width: -1` (automatic) means the scroll container only occupies as much horizontal space as its children. On a device this is unusual: a horizontal scroll arrangement typically spans the full screen width with children scrolling inside. With `Width: -1` and `overflow-x: auto`, the outer div may shrink to child content width causing the scrollbar to appear outside the intended region.

### 5. Horizontal scroll axis requires `flex-wrap: nowrap` — partially correct (Low)

The `.sim-scroll-horizontal` CSS class correctly sets `overflow-x: auto` and `flex-wrap: nowrap`. However, the parent class `.sim-arrangement--horizontal` already sets `flex-wrap: wrap`. Since `.sim-scroll-horizontal` is applied as an additional class on the same element, the later `flex-wrap: nowrap` in `.sim-scroll-horizontal` overrides the earlier `flex-wrap: wrap`. This works correctly due to CSS specificity/ordering, but relies on source-order precedence — a fragile pattern. Explicit consolidation would be safer.

### 6. colorValue() drops alpha channel (Medium)

The `colorValue()` function in `SimulationComponent.svelte` (line 45) extracts only the RGB portion of `&HAARRGGBB` format:

```js
if (ai) return `#${ai[2]}`;  // ai[1] is the alpha byte — silently discarded
```

Semi-transparent `BackgroundColor` values (e.g., `&H80FF0000` for 50% transparent red) will render as fully opaque in simulation. This affects all components with a `BackgroundColor` property including `HorizontalScrollArrangement`.

### 7. `Enabled` property has no visual effect on the container (Low)

In AI, disabling a container propagates the disabled state to its children (they become non-interactive). In the Tensor simulation the `enabled` reactive variable is computed for the arrangement but the `<div>` element has no `aria-disabled` attribute and children are not passed an "effective parent disabled" flag. Children are rendered using their own `isSimulationEnabled()` check based on their own state only, so disabling an arrangement does not disable its children.

### 8. No `data-sim-component` on the HorizontalScrollArrangement div (Low)

Looking at the Svelte template for `HorizontalArrangement` / `HorizontalScrollArrangement` (line 238–248):

```html
<div
  class="sim-arrangement sim-arrangement--horizontal"
  class:sim-scroll-horizontal={node.type === 'HorizontalScrollArrangement'}
  class:sim-unsupported={unsupportedHere}
  style={baseStyle()}
  data-sim-component={node.name}
>
```

`data-sim-component` is present. This is correct.

---

## Bugs Found

### Bug 1: Wrong default vertical alignment in CSS — AlignVertical defaults to Center instead of Top

**File:** `SimulationComponent.svelte`, line 376

```css
.sim-arrangement--horizontal {
  flex-direction: row;
  align-items: center;   /* <-- wrong: AI default is Top (align-items: flex-start) */
  flex-wrap: wrap;
}
```

The App Inventor default for `AlignVertical` is `1` = Top, which maps to `align-items: flex-start`. The current CSS uses `align-items: center` which simulates `AlignVertical = 2` (centered). Any arrangement relying on the default top-aligned children will appear with children centered vertically in simulation.

**Priority: High**

---

### Bug 2: `Width` default is `-1` (automatic) instead of `-2` (fill parent)

**File:** `simulation-capabilities.js`, line 56

```js
HorizontalScrollArrangement: { Visible: true, Enabled: true, Width: -1, Height: -2 },
```

App Inventor creates a `HorizontalScrollArrangement` whose outer `HorizontalScrollView` frame fills its parent width by default. With `Width: -1` the simulated container only wraps its content width. This means the scroll area may be narrower than the device width, placing the scrollbar in the wrong position.

**Priority: Medium**

---

### Bug 3: `AlignHorizontal` and `AlignVertical` absent from `SIMULATION_VISUAL_PROPS`

**File:** `simulation-capabilities.js`, lines 63–100

Neither `AlignHorizontal` nor `AlignVertical` appear in `SIMULATION_VISUAL_PROPS`. This means:
1. When a user sets these properties in the designer, `deriveStateFromDesignerProps()` skips them.
2. When a block calls `set HorizontalScrollArrangement1.AlignHorizontal to 3`, the Go host updates state via `SetProperty` (which has no special case for alignment), but `SIMULATION_VISUAL_PROPS` is checked by the Svelte layer to decide which state changes to track for re-render. The `statePatch` from the Go host would contain the updated value, but the Svelte component never reads `props.AlignHorizontal` to produce any CSS change.

**Priority: High**

---

### Bug 4: `Image` property absent from `SIMULATION_VISUAL_PROPS` for arrangement types

**File:** `simulation-capabilities.js`, line 63–100

`Image` (the background image path) is absent from `SIMULATION_VISUAL_PROPS`, so it is never tracked for re-render. The `SIMULATION_DEFAULTS` for `HorizontalScrollArrangement` also lacks an `Image` entry, so `GetProperty` on this field returns `NullVal()`. The Svelte arrangement branch never applies a `background-image` CSS property.

**Priority: Medium**

---

### Bug 5: `BackgroundColor` missing from `SIMULATION_DEFAULTS` for HorizontalScrollArrangement

**File:** `simulation-capabilities.js`, line 56

```js
HorizontalScrollArrangement: { Visible: true, Enabled: true, Width: -1, Height: -2 },
```

Unlike `HorizontalArrangement` (which also lacks it), the AI source sets `BackgroundColor(Component.COLOR_DEFAULT)` in the constructor (line 134 of `HVArrangement.java`). There is no `BackgroundColor` in the Tensor defaults. When a block reads `HorizontalScrollArrangement1.BackgroundColor` it receives `NullVal()` rather than the AI default `COLOR_DEFAULT` (transparent / 0x00000000).

**Priority: Medium**

---

## Android/App Inventor Standards Compliance

| Standard | Status |
|---|---|
| `&HAARRGGBB` color format | Partially — alpha byte is discarded in `colorValue()` |
| Size constants: -1 = automatic, -2 = fill parent, -(n+1000) = n% | Correctly handled in `sizeStyle()`; default `Width: -1` diverges from AI's fill-parent default |
| Horizontal orientation (row layout) | Correct — `flex-direction: row` |
| Scrollable — children do not wrap and container scrolls horizontally | Correct — `flex-wrap: nowrap; overflow-x: auto` |
| `AlignHorizontal` default = 1 (Left) | Non-compliant — not implemented; no CSS mapping; not in visual props |
| `AlignVertical` default = 1 (Top) | Non-compliant — CSS hardcodes `align-items: center` (= Top should be `flex-start`) |
| Background image (`Image` property) takes precedence over `BackgroundColor` | Not implemented |
| `BackgroundColor` default = `COLOR_DEFAULT` (0 = transparent) | Partially — no default in `SIMULATION_DEFAULTS`; reads return null |
| No events or methods on this component | Correct — no events or methods defined |
| Disabled container propagates disabled state to children | Not implemented |
| Children laid out in the inner linear layout, scroll view wraps it | Approximated via CSS flex; no explicit inner wrapper as in AI's two-layer view hierarchy |

---

## Summary

The Tensor IDE simulator renders `HorizontalScrollArrangement` correctly for its core purpose: a horizontally-scrolling row container that holds child components without wrapping. The `overflow-x: auto; flex-wrap: nowrap` CSS combination faithfully replicates the AI `HorizontalScrollView` scroll behaviour, and size encoding is correctly handled.

However, there are notable compliance gaps in alignment, default sizing, and background image support:

**Total issues found: 12** (counting the 5 concrete bugs and 7 behaviour gaps including standards compliance items listed above, where several overlap).

**Top 3 action items:**

1. **Fix `AlignVertical` default: change `align-items: center` to `align-items: flex-start`** in `.sim-arrangement--horizontal` CSS. The AI default for `AlignVertical` is `1` = Top, which maps to `flex-start`. The current CSS centres children vertically by default, silently changing the visual layout of every horizontal arrangement without the user setting any alignment. Additionally, add `AlignHorizontal` and `AlignVertical` to `SIMULATION_VISUAL_PROPS` and map their values to `align-items` / `justify-content` CSS properties in `baseStyle()` or the arrangement branch.

2. **Add `AlignHorizontal` and `AlignVertical` to `SIMULATION_VISUAL_PROPS` and implement CSS mapping.** These are frequently-used designer properties that control content layout. Without them, any arrangement configured with `AlignHorizontal = 3` (center) or `AlignVertical = 3` (bottom) renders incorrectly. Values: AlignHorizontal: 1→`justify-content: flex-start`, 2→`flex-end`, 3→`center`; AlignVertical: 1→`align-items: flex-start`, 2→`center`, 3→`flex-end`.

3. **Change `Width` default from `-1` to `-2` (fill parent) for `HorizontalScrollArrangement`** in `SIMULATION_DEFAULTS`. The AI scroll container fills its parent by default. An arrangement with `Width: -1` wraps content, so the scrollable region may be narrower than the screen width, producing a visually incorrect and confusing layout in simulation.
