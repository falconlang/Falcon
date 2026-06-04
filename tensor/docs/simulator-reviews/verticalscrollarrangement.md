# VerticalScrollArrangement Simulator Review

## Overview

`VerticalScrollArrangement` is a layout container in App Inventor that arranges child components vertically (top to bottom) inside a `ScrollView`, making the content scrollable when it overflows the container's bounds. It extends `HVArrangement` with `SCROLLABLE_ARRANGEMENT = true` and `LAYOUT_ORIENTATION_VERTICAL`. It has no events and no methods of its own — its entire API surface is inherited properties from `HVArrangement` and `AndroidViewComponent`.

In Tensor IDE, this component is rendered as a `<div>` with the CSS class `sim-scroll-vertical` (plus `sim-arrangement`), using `overflow-y: auto`. The implementation is largely structural/visual with no runtime method dispatch needed.

---

## Properties Analysis

### Supported

- `Visible` — supported, defaults to `true`, correctly gates rendering via `{#if visible}`.
- `Width` — supported in both defaults and `sizeStyle()`. The `-2` fill-parent / `-1` automatic encoding is handled.
- `Height` — supported in both defaults and `sizeStyle()`.
- `BackgroundColor` — supported via `baseStyle()` → `colorValue()`. AI colour strings (`&HaarrggBB`) are parsed.
- `Left` / `Top` — supported for `AbsoluteArrangement` parent positioning via `baseStyle()`.

### Missing / Unsupported

| Property | AI Default | Priority | Notes |
|---|---|---|---|
| `AlignHorizontal` | `1` (Left) | **High** | Not in `SIMULATION_DEFAULTS`, not applied to the container div. In AI, sets `gravity` on the inner `LinearLayout`. Children can be left/centre/right aligned within the scroll container. |
| `AlignVertical` | `1` (Top) | **Medium** | Not in `SIMULATION_DEFAULTS`, not applied. Controls vertical gravity within the LinearLayout; meaningful when the content is shorter than the container. |
| `Image` | `""` | **Medium** | Background image property present in AI (inherited from `HVArrangement`) and exposed in `simple_components.json`, but absent from `SIMULATION_DEFAULTS` and not rendered. The `assetUrl` reactive is bound to `props.Picture \|\| props.Image` but is only wired to the `<img>` path for `Image` components, not to the arrangement's background. |
| `HeightPercent` (write-only) | — | **Low** | AI exposes a `HeightPercent` setter. Tensor handles the encoded `-1000x` percent format in `sizeStyle()` for the `Height` property, which partially covers this, but the setter is not available as a distinct callable property from blocks. |
| `WidthPercent` (write-only) | — | **Low** | Same as above for width. |
| `Enabled` | — | **Low** | Not a real property on `HVArrangement`/`VerticalScrollArrangement` in AI (arrangements have no `Enabled` concept — only individual interactive widgets do). Tensor adds `Enabled: true` in `SIMULATION_DEFAULTS` which is harmless but spurious. |
| `Column` / `Row` | invisible | **Info** | Internal grid-layout positioning properties, invisible in designer. Not relevant to simulation. |

### Wrong Defaults or Types

| Property | Tensor Default | AI Default | Priority | Notes |
|---|---|---|---|---|
| `Width` | `-2` (fill parent) | Not set in designer defaults (Automatic / `-1`) | **High** | The AI `simple_components.json` does not include a designer-level `Width` default for `VerticalScrollArrangement`, meaning it uses the runtime default of `LENGTH_PREFERRED` (-1, Automatic). Tensor sets `-2` (fill parent / 100%), which makes the arrangement always span full width even when the user never set a width. This is incorrect — the component should default to automatic sizing. |
| `Height` | `-1` (automatic) | Not set in designer defaults (Automatic / `-1`) | **Medium** | The AI default is Automatic (`-1`). Tensor uses `-1` here which matches, but note that `sizeStyle()` treats `-1` as `return ''` (no explicit style), which in CSS means the container takes the height of its children. For a scroll arrangement, this means it will never clip/scroll because it grows to fit content. A real `VerticalScrollArrangement` on Android is typically given a fixed or fill-parent height to create a scrollable window. The simulation will silently not scroll if `Height` stays at `-1`. |
| `BackgroundColor` | Not in `SIMULATION_DEFAULTS` for VSA | AI default is `&H00000000` (transparent/default) | **Medium** | `SIMULATION_DEFAULTS` for `VerticalScrollArrangement` has no `BackgroundColor` key. The `baseStyle()` function checks `props.BackgroundColor` and if absent, emits no background rule, which renders as transparent — this accidentally matches AI's transparent default. However, the value is not available for block reads (returns `null`/`NullVal`), whereas AI would return `0x00000000`. |

---

## Events Analysis

### Supported

- None are defined in the AI source. `VerticalScrollArrangement` inherits from `HVArrangement` which itself defines **no events**. The `simple_components.json` confirms `"events": []`.

### Missing / Incorrect

- None. The component genuinely has no events in App Inventor. This is correctly reflected by the absence of any event wiring in the Tensor implementation.

---

## Methods Analysis

### Supported

- None are defined in the AI source. `simple_components.json` confirms `"methods": []`.

### Missing / Incorrect

- None. The component has no callable methods. This is correctly reflected in the Tensor `CallMethod` dispatcher — there is no case for `VerticalScrollArrangement`, so it falls through to `h.Unsupported(...)`. This is the correct behaviour.

---

## Behaviour Gaps

### 1. Scrolling does not activate unless a fixed Height is given (High)

In AI, `VerticalScrollArrangement` wraps its content in an Android `ScrollView`. The `ScrollView` clips its content to its measured bounds and enables vertical scrolling. For scrolling to work, the arrangement must have a bounded height. If `Height` is Automatic (`-1`), the Android `ScrollView` will still constrain to the screen, but in the Tensor CSS simulation, the div with `overflow-y: auto` and no explicit height will simply expand to fit all children and **never show a scrollbar**. A user who sets a fixed pixel height or `Fill Parent` (`-2`) will get correct scroll behaviour, but the default `-1` case silently breaks the scrolling semantics.

### 2. AlignHorizontal not applied to children (High)

In AI, `AlignHorizontal` sets the Android `gravity` flag on the inner `LinearLayout`, which controls how child views are aligned within the available horizontal space. In Tensor, the `sim-arrangement` CSS class uses `align-items: stretch`, which is equivalent to the default (Left/stretch). Changing `AlignHorizontal` to `3` (Center) or `2` (Right) via blocks has no visual effect in the simulation because the property is not read from `props` and not mapped to `justify-content` or `align-items`.

### 3. AlignVertical not applied to children (Medium)

Similarly, `AlignVertical` (`1` = Top, `2` = Center, `3` = Bottom) is not read. In AI this maps to `gravity` on the vertical axis of the `LinearLayout`. In Tensor the children always stack from the top, so `AlignVertical = 2` (Center) and `AlignVertical = 3` (Bottom) produce incorrect layouts.

### 4. Background Image not rendered (Medium)

The `Image` property lets users set a background image on the arrangement. In Tensor, `assetUrl` is computed as `resolveAssetUrl(assets, props.Picture || props.Image)` but this value is only used inside the `{:else if node.type === 'Image'}` branch. For an arrangement div, there is no binding of `assetUrl` to a CSS `background-image` style. Setting `Image` on a `VerticalScrollArrangement` via the designer or via blocks will have no visible effect.

### 5. BackgroundColor `COLOR_DEFAULT` (transparent) semantics (Low)

In AI, `BackgroundColor` defaults to `Component.COLOR_DEFAULT = 0x00000000`. When this value is set, `updateAppearance()` restores the original system drawable (the default Android background, typically a plain white/grey surface). In Tensor, `colorValue` for a `&H00000000` value would parse as `#000000` (black), not transparent, because the parser extracts `ai[2]` (the last 6 hex digits, `000000`), ignoring the alpha channel `00`. This means if a user explicitly reads and re-sets the default colour value, the background would incorrectly appear black rather than transparent/default.

Specifically in `colorValue()`:
```js
const ai = text.match(/^&H([0-9A-Fa-f]{2})([0-9A-Fa-f]{6})$/);
if (ai) return `#${ai[2]}`;
```
`&H00000000` → `ai[1]='00'` (alpha), `ai[2]='000000'` → returns `#000000` (opaque black). The alpha channel is discarded. For `&HFFFFFFFF` (white) this is harmless, but for colours with non-FF alpha (including the default `&H00000000`), the rendered colour is wrong.

### 6. Width default `-2` forces full-width expansion (High)

`SIMULATION_DEFAULTS` has `VerticalScrollArrangement: { Width: -2, Height: -1 }`. The `-2` fill-parent width means every `VerticalScrollArrangement` will always expand to 100% of its parent, even when the designer intention was "automatic" (shrink to content). This changes the layout geometry of any screen that uses a narrower `VerticalScrollArrangement` next to other components.

### 7. No `Enabled` property on arrangements (Low)

Tensor includes `Enabled: true` in `SIMULATION_DEFAULTS` for `VerticalScrollArrangement`, and `coerceSimulationValue` will coerce it. AI's `HVArrangement` does not expose an `Enabled` property — it is not in `simple_components.json`. This means Tensor silently accepts and stores a nonsensical property, and if user blocks write to `VerticalScrollArrangement1.Enabled`, Tensor would store it without complaint whereas AI would raise an error.

---

## Bugs Found

### Bug 1: `colorValue()` drops alpha channel for all AI colour values (Medium)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 44–45

```js
const ai = text.match(/^&H([0-9A-Fa-f]{2})([0-9A-Fa-f]{6})$/);
if (ai) return `#${ai[2]}`;
```

The regex captures alpha in `ai[1]` and RGB in `ai[2]`, but only `ai[2]` is used. For any colour with alpha < 255 (e.g. the default `&H00000000`, semi-transparent colours), the output CSS colour will be fully opaque. The fix is to return `#${ai[1]}${ai[2]}` or use `rgba()`. This affects not just `VerticalScrollArrangement` but all components that use `BackgroundColor`.

### Bug 2: `sim-scroll-vertical` class applied without `sim-arrangement` base class (Low)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 224–229

```svelte
class:sim-arrangement={node.type === 'VerticalArrangement'}
class:sim-scroll-vertical={node.type === 'VerticalScrollArrangement'}
```

`VerticalArrangement` gets class `sim-arrangement`.
`VerticalScrollArrangement` gets class `sim-scroll-vertical` **but not** `sim-arrangement`.

The `sim-arrangement` CSS rule provides `display: flex; flex-direction: column; align-items: stretch; gap: 8px; min-width: 0;` which is the core layout. Without it, `VerticalScrollArrangement` relies only on the `sim-scroll-vertical` rule which only sets `overflow-y: auto; max-height: 100%`. The children would not be laid out in a flex column, breaking vertical stacking.

In practice a `<div>` defaults to `display: block` which also stacks children vertically, so this may not be visually obvious, but the gap, min-width, and stretch alignment from `sim-arrangement` are missing, causing inconsistent spacing and sizing between `VerticalArrangement` and `VerticalScrollArrangement`.

### Bug 3: `sizeStyle` for Height `-1` returns empty string, preventing scrolling (High)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 50–58

```js
if (value === -1 || value === '-1' || value === undefined || value === '') return '';
```

When `Height` is `-1` (Automatic, the default), `sizeStyle('Height')` returns `''`, so no `height:` CSS rule is emitted. Combined with `overflow-y: auto`, the div will expand to fit content rather than scroll. For `VerticalScrollArrangement`, the whole point is scrolling — a height must be constrained. This is a fundamental semantic mismatch: the simulation will never scroll unless the user explicitly sets a pixel height or fill-parent (`-2`).

### Bug 4: `AlignHorizontal` / `AlignVertical` props not read or applied for arrangements (High)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 224–236

The `baseStyle()` call for `VerticalScrollArrangement` does not inspect `props.AlignHorizontal` or `props.AlignVertical`. There is no mapping to CSS `justify-content` or `align-items`. Setting these properties via blocks or the designer panel has no effect in the simulation.

### Bug 5: `SIMULATION_DEFAULTS` sets wrong `Width` for `VerticalScrollArrangement` (High)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 57

```js
VerticalScrollArrangement: { Visible: true, Enabled: true, Width: -2, Height: -1 },
```

AI does not set a designer-level `Width` default for `VerticalScrollArrangement` (it is omitted from `simple_components.json`'s `properties` array, meaning it uses the runtime default of `-1` / Automatic). Using `-2` causes the arrangement to always fill its parent width, which is incorrect for cases where the user intends a narrower scroll panel.

---

## Android/App Inventor Standards Compliance

| Aspect | Status |
|---|---|
| Vertical-only scroll (no horizontal) | Correct — `overflow-y: auto`, no `overflow-x`. |
| Children stacked top-to-bottom | Partially correct — `<div>` default block layout stacks children, but flex column from `sim-arrangement` is missing (Bug 2). |
| `LENGTH_PREFERRED (-1)` = Automatic (shrink-wrap) | Correctly returns empty string from `sizeStyle()`, but this breaks scrolling (Bug 3). |
| `LENGTH_FILL_PARENT (-2)` = 100% | Correctly maps to `width/height: 100%` in `sizeStyle()`. |
| Percent encoding (`<= -1000`) | Correctly decoded as `(-n - 1000)%` in `sizeStyle()`. |
| `BackgroundColor` transparent default | Broken — alpha channel dropped in `colorValue()` (Bug 1). |
| `AlignHorizontal` / `AlignVertical` | Not implemented (Bug 4). |
| `Image` background | Not implemented. |
| No events | Correctly not wired. |
| No methods | Correctly falls through to unsupported. |
| Children can be any component | Correctly handled via recursive `<svelte:self>`. |

---

## Summary

`VerticalScrollArrangement` is partially implemented. The structural rendering (div, children rendering, Visible, basic background colour) works. However several important gaps remain:

**Top 3 action items:**

1. **Fix `sim-scroll-vertical` to also apply `sim-arrangement` styles** (Bug 2) and **constrain height when default `-1`** by applying a sensible `max-height` or treating `-1` as `100%` for scroll containers specifically — without a bounded height, the scroll div grows to fit content and scrolling never activates (Bug 3). These two bugs together mean `VerticalScrollArrangement` cannot actually scroll in any default configuration.

2. **Implement `AlignHorizontal` and `AlignVertical` properties** (Bug 4): read `props.AlignHorizontal` and `props.AlignVertical` in `baseStyle()` and map them to CSS `align-items` (`1`→`flex-start`, `2`→`flex-end`, `3`→`center`) and `justify-content` respectively for arrangement divs. Add these properties to `SIMULATION_DEFAULTS` and `SIMULATION_VISUAL_PROPS`.

3. **Fix the `colorValue()` alpha-channel bug** (Bug 1): include the alpha byte when converting `&HaaRRGGBB` colour values to CSS, e.g. return `#${ai[1]}${ai[2]}` or use `rgba(r,g,b,a/255)`. This affects background colours across all components, not just arrangements.
