# CircularProgress Simulator Implementation Plan

## Overview

`CircularProgress` is a visible User Interface component (categoryString `USERINTERFACE`, `nonVisible = false`). Per the spec helpString it is "A visible component that indicates the progress of an operation using an animated loop." In App Inventor it is an *indeterminate* spinner — it wraps Android's indeterminate `ProgressBar` (the circular Material spinner) and continuously animates a rotating arc/loop. It has no determinate "value" or "percent done" surface; the only thing the developer can control is its tint (`Color`), its visibility, and its position/size.

Container relationship: it is a **standalone visible leaf** — it has no children and is placed inside any standard container (Screen/Form or an arrangement). It does not itself contain anything.

The component is intentionally minimal: the spec lists **zero events** and **zero methods**. Its entire interactive surface is property reads/writes. This makes it one of the simplest possible additions to the simulator.

## Feasibility Verdict

**Feasible.**

An indeterminate circular spinner is exactly what a browser does well. A pure-CSS rotating ring (a circular element with a partially-transparent border plus `@keyframes spin` + `animation`), or an inline SVG `<circle>` with a rotating dash, reproduces the App Inventor visual faithfully with no JS animation loop and no external dependency. The single styleable property, `Color`, maps directly to the border/stroke color via the existing `colorValue()` helper. `Visible`, `Width`, `Height`, `Left`, `Top` are all handled by the shared `baseStyle()` / `sizeStyle()` / `positionStyle()` helpers already used by every other component.

There are no native-only capabilities involved (no sensors, permissions, files, or device APIs), and there is no determinate progress value to track, so nothing about the component is impossible or even approximate in a browser. The only fidelity gaps are cosmetic (exact easing curve / arc sweep of the Material spinner versus a CSS ring), which are negligible.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Color` | `&HFF0000FF` (blue) | Visual | Tint of the rotating ring. Read via `colorValue(props.Color, '#0000ff')`; apply as the ring's `border-top-color` (CSS ring) or `stroke` (SVG) through a CSS custom property `--cp-color`. | **High** |
| `Visible` | `True` | Visual | Already gated by the top-level `{#if visible && !nonVisible}` block via `isSimulationVisible`. No branch-specific work. | **High** |
| `Width` | `""` (automatic) | Visual | `sizeStyle('Width')` in `baseStyle()`. With the AI automatic default the spinner uses its intrinsic size; give the ring element a sensible intrinsic px size (e.g. 36px) so automatic sizing looks right. | **Medium** |
| `Height` | `""` (automatic) | Visual | `sizeStyle('Height')` in `baseStyle()`. Same intrinsic-size note as Width. | **Medium** |
| `WidthPercent` | n/a (write-only block prop) | Visual | Encoded by the host as a negative `Width` sentinel (`<= -1000`), already decoded by `sizeStyle()`. Generic store path; no extra work. | **Low** |
| `HeightPercent` | n/a (write-only block prop) | Visual | Same as `WidthPercent`, via `sizeStyle('Height')`. | **Low** |
| `Left` | `""` | Behavioral | Only meaningful inside an `AbsoluteArrangement`; handled by `positionStyle()` in `baseStyle()`. | **Low** |
| `Top` | `""` | Behavioral | Same as `Left`. | **Low** |
| `Column` / `Row` | n/a | — | `rw: invisible` getters in the spec, not exposed to blocks/designer. Ignore. | — |

All defaults are honorable in a browser. There is no `Enabled` property in the spec (the indeterminate spinner has no enabled/disabled state), so do not add one.

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| — | — | The spec lists **no events**. Nothing to wire. |

No `EVENT_ARGS` entry is needed.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| — | — | The spec lists **no methods**. No `CallMethod` / `callCircularProgressMethod` handler is required. |

## Implementation Plan

### simulation-capabilities.js

1. Add `'CircularProgress'` to `SIMULATION_SUPPORTED_TYPES`. Do **not** add it to `SIMULATION_NONVISIBLE_TYPES` (it is visible).
2. Add a defaults block inside `buildSimulationDefaults()`:
   ```js
   CircularProgress: {
     Visible: true,
     Width: -1,
     Height: -1,
     Color: '&HFF0000FF',
   },
   ```
   Use `-1` (automatic) for Width/Height since the AI designer defaults are empty/automatic; the render branch supplies an intrinsic px size. Do not include `Enabled` (no such property).
3. Add `'Color'` to `SIMULATION_VISUAL_PROPS`. (`Visible`, `Width`, `Height`, `Left`, `Top` are already present.) `Color` is the only genuinely new prop name introduced by this component.
4. `isBooleanProp`: no change (only `Visible`, already covered).
5. `isNumericProp`: no change (`Width`/`Height`/`Left`/`Top` already listed). `Color` is an ARGB color string/number handled by `colorValue()` at render time, not a coerced numeric — leave it out of `isNumericProp` so it passes through `coerceSimulationValue` unchanged (consistent with how `BackgroundColor`/`TextColor` are treated).
6. `coerceSimulationValue`: no change — `Color` falls through to the default `return value` branch, matching every other color prop.
7. `deriveStateFromDesignerProps`: no change — no derived state.

### SimulationComponent.svelte

Add a render branch before the final `{:else}` placeholder. Reuse `baseStyle()` (it already folds in `sizeStyle` for Width/Height, `positionStyle` for Left/Top, and `Visible` is gated upstream) and `colorValue()` for the tint via a CSS custom property:

```svelte
{:else if node.type === 'CircularProgress'}
  <div
    class="sim-circular-progress"
    class:sim-unsupported={unsupportedHere}
    style={baseStyle(`--cp-color: ${colorValue(props.Color, '#0000ff')};`, { typography: false, arrangement: false, backgroundImage: false })}
    role="progressbar"
    aria-label="Loading"
    data-sim-component={node.name}
  >
    <span class="sim-cp-ring" aria-hidden="true"></span>
  </div>
```

CSS (add to the `<style>` block):

```css
.sim-circular-progress {
  display: inline-grid;
  place-items: center;
  min-width: 36px;
  min-height: 36px;
}

.sim-cp-ring {
  width: 28px;
  height: 28px;
  border: 3px solid color-mix(in srgb, var(--cp-color, #0000ff) 25%, transparent);
  border-top-color: var(--cp-color, #0000ff);
  border-radius: 50%;
  animation: sim-cp-spin 0.9s linear infinite;
}

@keyframes sim-cp-spin {
  to { transform: rotate(360deg); }
}

@media (prefers-reduced-motion: reduce) {
  .sim-cp-ring { animation-duration: 2.4s; }
}
```

Notes:
- No DOM event wiring (`on:click`/`input`/etc.) — the component is non-interactive (no events in the spec).
- `baseStyle()` is called with `typography:false, arrangement:false, backgroundImage:false` because none apply; pass `--cp-color` as the `extra` argument so it lands on the wrapper and cascades to the ring.
- If `color-mix` support is a concern for the target browser baseline, fall back to a fixed translucent track (e.g. `rgba(0,0,0,0.12)`) for the non-`border-top-color` sides; the spinning arc still uses `--cp-color`.
- The ring is the intrinsic visual; `baseStyle()`'s Width/Height (when set explicitly via blocks/`WidthPercent`) sizes the wrapper, and `place-items: center` keeps the ring centered.

### SimulateOverlay.svelte

**None.** No dialogs, toasts, or host effects are involved.

### simulation_wasm.go

**Generic store path suffices.** `Color`, `Visible`, `Width`, `Height`, `Left`, `Top` all read/write through the default `GetProperty`/`SetProperty` store with no component-specific logic. `WidthPercent`/`HeightPercent` are write-only block props; if the runtime already routes percent writes to the negative `Width`/`Height` sentinel encoding used elsewhere they need no new case, otherwise they fall through to the generic `setProperty` and are simply stored (the renderer reads `Width`/`Height`). No `callCircularProgressMethod` handler, no `runEvent`, no `effects`, no `Unsupported` call. The component type just needs to be reachable through the generic path (it is, once added to the supported set).

### design-schema-tree.js

**Already handled.** `canContainDesignComponent` falls through to `return containerTypes().has(parent)` for a standard standalone visible component, so CircularProgress can be placed inside Screen/Form or any arrangement with no edit. `unsupportedSimulationComponents` stops flagging it as soon as `'CircularProgress'` is in `SIMULATION_SUPPORTED_TYPES`, which removes the dashed "unsupported" placeholder. `designTreeToInitialState` will merge the new `SIMULATION_DEFAULTS.CircularProgress` automatically. No change required.

## Dependencies & Ordering

- **External libraries:** none. Pure CSS animation.
- **Prerequisite components:** none. This is a self-contained leaf; it does not depend on Canvas/Map/Chart child plumbing or on any other component being implemented first.

## Web-Platform Limitations & Fidelity Caveats

- The CSS ring's sweep/easing will not be a pixel-perfect match for Android's Material indeterminate spinner (which uses a growing/shrinking arc with a specific cubic easing). The simulated version is a constant-width arc rotating at constant speed — visually equivalent in intent, cosmetically different in detail.
- App Inventor's indeterminate spinner has a fixed intrinsic size from the platform theme; the simulator picks a reasonable intrinsic px size (28px ring / 36px box). When the developer leaves Width/Height automatic this is an approximation of the device's density-dependent size.
- `color-mix()` is used for the translucent track ring; on browsers without it the track falls back to a fixed translucent grey (documented in the CSS note above). The animated arc color (`Color`) is always honored.
- No determinate/percent state exists in the component, so there is nothing to under-simulate there — the simulation is feature-complete relative to the spec.

## Effort Estimate

**S** — one supported-type entry + a 4-line defaults block + one `SIMULATION_VISUAL_PROPS` addition in `simulation-capabilities.js`; one render branch + a small CSS block in `SimulationComponent.svelte`; zero Go changes (generic store path); zero `design-schema-tree.js` changes. No events, methods, effects, or external dependencies.
