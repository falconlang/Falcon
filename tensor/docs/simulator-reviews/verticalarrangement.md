# VerticalArrangement Simulator Review

## Overview

`VerticalArrangement` is a layout container that arranges its child components top-to-bottom, left-aligned, with no built-in scrolling (the scrollable variant is `VerticalScrollArrangement`). In App Inventor it extends `HVArrangement`, which itself extends `AndroidViewComponent`. The component surfaces six designer properties (`AlignHorizontal`, `AlignVertical`, `BackgroundColor`, `Image`, `Visible`, `Width`/`Height`) inherited from `HVArrangement`, no events of its own (it fires no events), and no public methods. The Tensor IDE simulator renders it as a CSS flex-column `<div>` in `SimulationComponent.svelte`, with property defaults in `simulation-capabilities.js` and runtime state managed by `simulation_wasm.go`. The basic layout shell works but several visual properties and layout-semantics are missing or incorrect.

---

## Properties Analysis

### Supported
- `Visible` — defaults `true`; the outer `{#if visible && !nonVisible}` gate hides the div entirely when false.
- `Width` — size constants `-1` (automatic/wrap), `-2` (fill parent), and percentage encoding `<= -1000` all handled via `sizeStyle()` in `SimulationComponent.svelte` (line 51-58).
- `Height` — same as Width via `sizeStyle()`.
- `BackgroundColor` — passed through `baseStyle()` → `colorValue()` and emitted as CSS `background:` (line 64).
- `Enabled` — tracked in state; `isSimulationEnabled()` gate used by children.

### Missing / Unsupported

| Property | AI Source | Priority | Notes |
|---|---|---|---|
| `AlignHorizontal` | `HVArrangement.AlignHorizontal()` — values 1=Left, 2=Right, 3=Center; default=1 (Left) | **High** | Not in `SIMULATION_DEFAULTS.VerticalArrangement`, not in `SIMULATION_VISUAL_PROPS`, not read by `baseStyle()`, not applied as CSS `align-items` or `justify-content`. Children are always `align-items: stretch` regardless of the property value. |
| `AlignVertical` | `HVArrangement.AlignVertical()` — values 1=Top, 2=Center, 3=Bottom; default=1 (Top) | **High** | Same as above — completely absent from all layers. Vertical distribution of children within the container is always top-aligned regardless of the property value. |
| `Image` | `HVArrangement.Image()` — path to a background image asset; takes precedence over `BackgroundColor` | **Medium** | Not in `SIMULATION_DEFAULTS.VerticalArrangement`, not in `SIMULATION_VISUAL_PROPS`, and `baseStyle()` has no branch to apply a `background-image`. The AI runtime silences `BackgroundColor` when an image is set; the simulator has no equivalent logic for arrangements. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `Width` | `LENGTH_PREFERRED` = `-1` (Automatic) | `-2` (Fill Parent) | **High** | In App Inventor, `VerticalArrangement.Width` defaults to `Automatic` (`-1`). `simulation-capabilities.js` line 55 sets it to `-2` (Fill Parent). This makes every arrangement fill its parent horizontally even when it should only be as wide as its widest child, causing layouts to render wider than they should. |
| `Height` | `LENGTH_PREFERRED` = `-1` (Automatic) | `-2` (Fill Parent) | **Medium** | Similarly, AI default for `Height` is also `Automatic` (`-1`). Tensor defaults to `-2` (Fill Parent), causing the arrangement to occupy all available vertical space instead of wrapping its children. For nested arrangements this is particularly misleading. |
| `BackgroundColor` | `COLOR_DEFAULT` = `&H00000000` (transparent/default) | Not in defaults (absent) | **Medium** | The AI initialises `backgroundColor = Component.COLOR_DEFAULT` (`0x00000000`, transparent). Tensor's `SIMULATION_DEFAULTS.VerticalArrangement` has no `BackgroundColor` key at all, so the property starts as `undefined`. `baseStyle()` skips the CSS rule when `props.BackgroundColor` is falsy (line 64), which happens to render transparently and is functionally acceptable, but block-sets from a `Screen.Initialize` handler that reads `BackgroundColor` before setting it will receive `null` instead of the proper `&H00000000` value. |
| `Enabled` | Not a designer property on arrangements in AI | `true` | **Low** | `HVArrangement` does not expose an `Enabled` property. Tensor tracks it anyway via `COMMON_VISIBLE_PROPS`, which is harmless but inconsistent with the AI spec. |

---

## Events Analysis

`VerticalArrangement` (and its parent `HVArrangement`) define **no events** in the App Inventor source. The component is purely a passive layout container.

### Supported
- N/A — No events are specified; none are implemented.

### Missing / Incorrect
- None required by spec.

> Note: `HVArrangement` does not inherit `TouchComponent` or `Touchable`; it therefore has no `Click`, `LongClick`, `Touch`, or gesture events. The simulator correctly does not emit any events for this component type.

---

## Methods Analysis

`VerticalArrangement` and `HVArrangement` define **no public `@SimpleFunction` methods** in the App Inventor source.

### Supported
- N/A — No methods are specified; none are implemented.

### Missing / Incorrect
- None required by spec.

> Internal methods like `$add()`, `setChildWidth()`, `setChildHeight()` are Android/host concerns with no direct simulation surface.

---

## Behaviour Gaps

### 1. AlignHorizontal / AlignVertical not applied to layout (High)
The most significant behaviour gap. In AI, `AlignHorizontal` (1/2/3 = left/right/center) maps to Android gravity on the inner `LinearLayout`, directly controlling where children land when the arrangement is wider than the children. `AlignVertical` (1/2/3 = top/center/bottom) similarly controls cross-axis gravity. Neither is wired in the simulator. The Svelte CSS for `.sim-arrangement` hard-codes `align-items: stretch` (line 369), which corresponds to none of the three valid App Inventor horizontal alignment values for a vertical arrangement. The correct CSS mapping is:

| AI `AlignHorizontal` value | CSS `align-items` |
|---|---|
| 1 (Left, default) | `flex-start` |
| 2 (Right) | `flex-end` |
| 3 (Center) | `center` |

| AI `AlignVertical` value | CSS `justify-content` |
|---|---|
| 1 (Top, default) | `flex-start` |
| 2 (Center) | `center` |
| 3 (Bottom) | `flex-end` |

### 2. Width/Height size-constant defaults diverge from AI spec (High)
`VerticalArrangement.Width` is `Automatic` (`-1`) and `Height` is `Automatic` (`-1`) in App Inventor. Tensor uses `-2` (Fill Parent) for both. In practice this means a plain `VerticalArrangement` in the simulator always stretches to fill its container in both dimensions, producing over-wide, over-tall previews that do not match what users will see on their devices.

### 3. Automatic width/height semantics for Fill-Parent children (Medium)
The AI javadoc (VerticalArrangement.java lines 37-43) specifies: when a `VerticalArrangement`'s `Height` is `Automatic`, any child whose `Height` is `Fill Parent` behaves as if it were `Automatic`. The simulator has no logic to detect this situation and re-interpret child sizing. A label with `Height = -2` inside an `Automatic`-height arrangement would correctly shrink to its content in AI but expand unexpectedly in the simulator.

### 4. Background image takes precedence over BackgroundColor (Medium)
`HVArrangement.updateAppearance()` (lines 423-440) clears the background colour when an image is set, and restores the default drawable when `COLOR_DEFAULT` is active. The simulator's `baseStyle()` simply emits both CSS rules independently; if an `Image` property were honoured, both `background-image` and `background` could coexist in the style attribute with unpredictable layering.

### 5. Empty arrangement minimum size (Low)
In AI, an empty `VerticalArrangement` with `Width=Automatic` renders at `EMPTY_HV_ARRANGEMENT_WIDTH = 100` dp (ComponentConstants line from grep). The simulator renders at whatever its parent dictates (or zero width if no children and no explicit width). For blocks that read `Width` at runtime on an empty arrangement the returned value will differ from AI.

### 6. Gap between children introduced by simulator (Low)
The `.sim-arrangement` CSS includes `gap: 8px` (line 359). App Inventor's `LinearLayout` does not insert margin between children by default; children are packed with no gap. The visual spacing looks cleaner in the simulator but misrepresents the AI layout.

### 7. `COLOR_DEFAULT` transparent background vs. white screen (Low)
`HVArrangement` initialises `BackgroundColor(Component.COLOR_DEFAULT)` which results in a transparent background (showing the parent's background through). The simulator div has no explicit background, which is also transparent. This is effectively correct, but the `colorValue()` function (line 43-48 of SimulationComponent.svelte) will return the `fallback` (`'transparent'`) only when the value is falsy/unrecognised; if `&H00000000` were set explicitly it would be parsed as `#000000` (the `ai` regex captures the RGB portion, discarding the alpha byte), making it appear as solid black rather than transparent. This would be a visible defect if `BackgroundColor` were included in the defaults.

---

## Bugs Found

### Bug 1 — `colorValue()` strips the alpha channel, making transparent colors render as opaque black (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte` lines 43-48

```js
const ai = text.match(/^&H([0-9A-Fa-f]{2})([0-9A-Fa-f]{6})$/);
if (ai) return `#${ai[2]}`;
```

`ai[1]` is the alpha byte; `ai[2]` is the RGB portion. The alpha is silently discarded. The App Inventor color `&H00RRGGBB` (fully transparent, any colour) will be rendered as `#RRGGBB` — opaque. For `BackgroundColor = &H00000000` (AI default, transparent) this returns `#000000` (solid black), obliterating the arrangement background. This is not yet triggered for `VerticalArrangement` because `BackgroundColor` is absent from its defaults, but any runtime `set BackgroundColor to` call from Blocks code will exhibit the bug. The same defect exists for every component that uses `colorValue()`.

**Fix:** Convert the alpha byte from hex to decimal and emit `rgba()`:
```js
if (ai) {
  const a = (parseInt(ai[1], 16) / 255).toFixed(3);
  const hex = ai[2];
  const r = parseInt(hex.slice(0,2), 16);
  const g = parseInt(hex.slice(2,4), 16);
  const b = parseInt(hex.slice(4,6), 16);
  return `rgba(${r},${g},${b},${a})`;
}
```

### Bug 2 — `Width` and `Height` defaults are `Fill Parent` instead of `Automatic` (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js` line 55

```js
VerticalArrangement: { Visible: true, Enabled: true, Width: -2, Height: -2 },
```

Both should be `-1` (Automatic) to match `Component.LENGTH_PREFERRED`. The current `-2` (Fill Parent) causes the arrangement to expand to fill its container, which is only correct when the user explicitly sets those properties to Fill Parent.

### Bug 3 — `AlignHorizontal` and `AlignVertical` are not in `SIMULATION_VISUAL_PROPS` (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js` lines 63-100

Neither `AlignHorizontal` nor `AlignVertical` appears in the `SIMULATION_VISUAL_PROPS` set. This means:
1. `deriveStateFromDesignerProps()` silently drops these properties when the designer writes them.
2. `coerceSimulationValue()` will not coerce them to integers.
3. They are never passed to `baseStyle()`, so even if the rendering were implemented, the values would never arrive in the component's `props` object.

### Bug 4 — `Image` property for arrangements is not in `SIMULATION_VISUAL_PROPS` (Medium)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`

`Image` (the background image path for arrangements) is absent from `SIMULATION_VISUAL_PROPS`. It is also absent from `SIMULATION_DEFAULTS.VerticalArrangement`. This means setting `Image` in the Designer or at runtime has no visual effect, and `deriveStateFromDesignerProps()` silently discards it.

Note: `assetUrl` in `SimulationComponent.svelte` line 34 resolves `props.Picture || props.Image`, but this is only used in the `{:else if node.type === 'Image'}` branch — it does not apply a CSS `background-image` for arrangement types.

### Bug 5 — `sizeStyle()` maps `-2` to `height: 100%` which fills the immediate CSS parent, not the screen (Medium)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte` line 52

```js
if (value === -2 || value === '-2') return prop === 'Width' ? 'width: 100%;' : 'height: 100%;';
```

`height: 100%` requires the parent element to have a defined height, otherwise it collapses to `0`. Nested arrangements with `Height = -2` inside a `VerticalArrangement` that itself has `Height = -2` (the current default) collapse to zero height in CSS unless the root screen element has a fixed height. App Inventor resolves Fill Parent by dividing remaining container height equally among Fill-Parent children at the Android layout pass; this is fundamentally different from CSS percentage heights. The result is that nested Fill-Parent arrangements often render invisible in the simulator.

---

## Android/App Inventor Standards Compliance

| Concern | Compliant? | Notes |
|---|---|---|
| Vertical (top-to-bottom) layout direction | Yes | `flex-direction: column` matches AI's vertical `LinearLayout`. |
| Default left-alignment of children | Partial | `align-items: stretch` (CSS default for flex-column) is not the same as `gravity_left`. Children that are narrower than the container will stretch to fill width rather than align to the left edge. |
| No scrolling | Yes | The plain `VerticalArrangement` div has no `overflow` CSS rule. Scrolling is correctly handled only by `VerticalScrollArrangement`. |
| `COLOR_DEFAULT` transparent background | Partial | Functionally transparent (no CSS background emitted) when `BackgroundColor` is absent from defaults. Would break if it were included due to the alpha-stripping bug. |
| Size constants `-1`/`-2`/`-1000x` encoding | Partial | The encoding is understood and decoded, but `-2` fill semantics differ from AI (CSS `100%` vs. equal-share at layout time). Percentage encoding (`<= -1000`) is correctly decoded as `-n - 1000` percent. |
| Child `Fill Parent` in `Automatic` parent re-interpreted as `Automatic` | No | Not implemented. See Behaviour Gap 3. |
| Empty arrangement minimum width of 100dp | No | Not implemented. |
| Background image priority over colour | No | `Image` property not wired at all for arrangements. |
| `AlignHorizontal` / `AlignVertical` | No | Not implemented at any layer. |

---

## Summary

The Tensor IDE simulator provides a working vertical flex container for `VerticalArrangement` — children stack top-to-bottom and visibility/basic sizing is functional. However, two critical gaps prevent correct layout fidelity:

1. **Wrong Width and Height defaults (`-2` Fill Parent instead of `-1` Automatic)** — Every `VerticalArrangement` in the simulator expands to fill its container, producing layouts that can look significantly different from the real App Inventor output.

2. **`AlignHorizontal` and `AlignVertical` are entirely absent** — These properties cannot be set, stored, or rendered. Users who rely on centered or right-aligned content inside arrangements will see no effect in simulation.

3. **`colorValue()` alpha-stripping bug** — Although not yet triggered by `VerticalArrangement` default state, any `set BackgroundColor to` block call that passes a colour with a non-FF alpha byte (including the AI color default `&H00000000`) will render as fully opaque rather than the intended transparency, causing visually incorrect backgrounds across all components.

**Top 3 action items:**
1. Fix `SIMULATION_DEFAULTS.VerticalArrangement` — change `Width` and `Height` from `-2` to `-1`.
2. Add `AlignHorizontal` and `AlignVertical` to `SIMULATION_VISUAL_PROPS` and `SIMULATION_DEFAULTS.VerticalArrangement` (defaults `1`/`1`), and apply them as `align-items` / `justify-content` in `baseStyle()` for arrangement types.
3. Fix `colorValue()` to preserve the alpha channel by emitting `rgba()` instead of discarding the alpha byte and emitting a plain hex colour.
