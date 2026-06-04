# Slider Simulator Review

## Overview

The `Slider` component is a draggable range control backed by Android's `SeekBar`. It exposes `MinValue`, `MaxValue`, and `ThumbPosition` properties plus three colour properties (`ColorLeft`, `ColorRight`, `ThumbColor`) and a step-count property (`NumberOfSteps`). Three events fire during interaction: `PositionChanged(thumbPosition)`, `TouchDown()`, and `TouchUp()`. The Tensor IDE simulator renders it as an HTML `<input type="range">` in `SimulationComponent.svelte`, with defaults declared in `simulation-capabilities.js` and event dispatch routed through `simulation_wasm.go`. Core numeric range properties and the three events are wired up correctly; however, all three colour properties are not applied to the rendered element, `NumberOfSteps` is entirely absent, the `Height` default is wrong, `ThumbEnabled=false` produces the wrong visual result, and the Go backend performs no clamping/recalculation when properties change programmatically.

---

## Properties Analysis

### Supported

- `MinValue` — declared in defaults (`10`), in `SIMULATION_VISUAL_PROPS`, in `coerceSimulationValue` (numeric), and correctly bound to `min` attribute.
- `MaxValue` — declared in defaults (`50`), in `SIMULATION_VISUAL_PROPS`, in `coerceSimulationValue` (numeric), and correctly bound to `max` attribute.
- `ThumbPosition` — declared in defaults (`30`), in `SIMULATION_VISUAL_PROPS`, in `coerceSimulationValue` (numeric), and correctly bound to `value` attribute; updated back via `sliderInput`.
- `ThumbEnabled` — declared in defaults (`true`), coerced as boolean, and partially applied (see issues below).
- `Enabled` — applied via HTML `disabled` attribute.
- `Visible` — controls whether the element is rendered at all.
- `Width` — size constants `-1` (auto), `-2` (fill), and percentage encoding `<= -1000` all handled in `sizeStyle()`.
- `ColorLeft` — present in `SIMULATION_VISUAL_PROPS` (tracked, but not rendered — see Missing section).
- `ColorRight` — present in `SIMULATION_VISUAL_PROPS` (tracked, but not rendered — see Missing section).

### Missing / Unsupported

| Property | Priority | Notes |
|---|---|---|
| `ThumbColor` | **High** | AI default is dark-grey (`&HFF444444`). Not in `SIMULATION_DEFAULTS` for Slider, not in `SIMULATION_VISUAL_PROPS`, not in `coerceSimulationValue`, and not applied to the rendered thumb. The thumb always renders as the hard-coded `accent-color: #2563eb` (blue). |
| `NumberOfSteps` | **High** | AI default is `100`. Controls slider precision. Entirely absent from `SIMULATION_DEFAULTS`, `SIMULATION_VISUAL_PROPS`, `coerceSimulationValue`, and the HTML `<input>`. The HTML `<input type="range">` always uses a `step` of `1`; without deriving `step = (MaxValue - MinValue) / NumberOfSteps` the reported `ThumbPosition` values will differ from what App Inventor would produce when `NumberOfSteps != (MaxValue - MinValue)`. |
| `ColorLeft` (rendering) | **High** | Present in `SIMULATION_VISUAL_PROPS` so the value is tracked in state, but the `Slider` branch of the template never reads `props.ColorLeft`. The left track always renders as the browser's default or `accent-color`. AI default is orange (`&HFFFFC800`). |
| `ColorRight` (rendering) | **High** | Same as `ColorLeft` — tracked but never applied. AI default is grey (`&HFF888888`). |
| `Width` (default) | **Low** | No explicit `Width` default is provided for Slider in `SIMULATION_DEFAULTS`; the entry uses `Visible: true, Enabled: true, Width: -1, Height: -2, ...` which is fine, but it is worth noting the AI component inherits `Width` from `AndroidViewComponent` with no special override. |

### Wrong Defaults or Types

| Property | Tensor Default | AI Default | Priority | Notes |
|---|---|---|---|---|
| `Height` | `-2` (fill parent / `height: 100%`) | Fixed intrinsic height (~32 dp, `@Override` removes the designer setter) | **High** | `-2` causes the slider to stretch to fill its parent container's full height. The AI component explicitly removes the `Height` designer annotation so users cannot change it; the Android `SeekBar` has a small, fixed intrinsic height. The Tensor default should be `-1` (automatic) so the slider adopts the browser's natural height for `<input type="range">`. |
| `ColorLeft` | Not set (no key in `SIMULATION_DEFAULTS`) | Orange `&HFFFFC800` | **Medium** | If code reads `ColorLeft` on a freshly created Slider before it has been set, it will get `null` / `undefined` instead of the AI-spec orange. |
| `ColorRight` | Not set | Grey `&HFF888888` | **Medium** | Same as above. |
| `ThumbColor` | Not set | Dark grey `&HFF444444` | **Medium** | Same as above. `ThumbColor` is also absent from `SIMULATION_VISUAL_PROPS` so it would be ignored even if written. |
| `NumberOfSteps` | Not set | `100` | **Medium** | Missing from defaults entirely, so a block that reads `Slider1.NumberOfSteps` would return `null`. |

---

## Events Analysis

### Supported

- `PositionChanged(thumbPosition)` — fired via `sliderInput` → `emitInteraction` with `args: [value]`. Argument is a JavaScript `Number` (float), matching AI's `float thumbPosition`. Declared correctly in `EVENT_ARGS.Slider.PositionChanged`. State update (`ThumbPosition`) is emitted together with the event in the same `emitInteraction` call. This is correct.
- `TouchDown()` — fired on `pointerdown`. No arguments, matching AI's `TouchDown()`. Not listed in `EVENT_ARGS` which is correct (empty args).
- `TouchUp()` — fired on `pointerup`. No arguments, matching AI's `TouchUp()`. Not listed in `EVENT_ARGS` which is correct.

### Missing / Incorrect

| Event | Priority | Notes |
|---|---|---|
| `TouchDown` / `TouchUp` triggered when `ThumbEnabled=false` | **Medium** | In the AI source, when `ThumbEnabled` is false the `OnTouchListener` consumes touch events, preventing `onStartTrackingTouch` / `onStopTrackingTouch` from firing, so `TouchDown` and `TouchUp` are NOT dispatched. In Tensor, the slider is rendered with `disabled={... || props.ThumbEnabled === false}`. HTML `disabled` inputs do not fire pointer events at all, so this is effectively correct for `TouchDown`/`TouchUp`. However, see the visual bug in the Bugs section. |
| `PositionChanged` not fired on programmatic `ThumbPosition` set | **Low** | In the AI source, `setSeekbarPosition()` changes the seekbar progress which calls `onProgressChanged`, which fires `PositionChanged`. In Tensor, when `SetProperty("ThumbPosition", ...)` is called from blocks, the state is updated but no `PositionChanged` event is dispatched. This matches the `notice` flag logic in App Inventor (programmatic sets via block setter do trigger the event in AI, however). This is a minor deviation. |

---

## Methods Analysis

### Supported

The Slider component defines no `@SimpleFunction`-annotated methods in the AI source. All interactions are property-based. No method support is required.

### Missing / Incorrect

None — no methods to implement.

---

## Behaviour Gaps

### 1. ThumbPosition Clamping (Critical)

In the AI source (`Slider.java` line 260):
```java
thumbPosition = Math.max(Math.min(position, maxValue), minValue);
```
When blocks set `Slider1.ThumbPosition` to a value outside `[MinValue, MaxValue]`, App Inventor clamps the value silently. In Tensor the `simulationHost.SetProperty` method has no Slider-specific logic and writes the raw value directly to state. The HTML `<input type="range">` will clamp visually at the browser level, but the stored `ThumbPosition` in state will be out-of-range, so a subsequent `get Slider1.ThumbPosition` block would return the unclamped value.

### 2. MinValue / MaxValue Mutual Clamping (High)

In the AI source:
- `MinValue(float value)` sets `maxValue = Math.max(value, maxValue)` — if the new min exceeds the current max, the max is raised to match.
- `MaxValue(float value)` sets `minValue = Math.min(value, minValue)` — if the new max is below the current min, the min is lowered to match.
- Both recalculate `thumbPosition` from the current seekbar progress after the change.

In Tensor, `SetProperty` stores each property independently. Setting `MinValue = 60` when `MaxValue = 50` will leave `MaxValue < MinValue`, producing an inverted HTML range that browsers clamp to `value == min` and may render incorrectly.

### 3. ColorLeft / ColorRight Not Rendered (High)

`ColorLeft` and `ColorRight` are included in `SIMULATION_VISUAL_PROPS` so state changes will be tracked, but the Slider branch of `SimulationComponent.svelte` never reads those props and never generates any CSS to style the track. The standard browser range track is always shown, ignoring these properties completely. The AI default colours (orange left track, grey right track) are never applied.

Implementation note: CSS custom properties and a linear-gradient background on the range track (or `-webkit-slider-runnable-track` pseudo-element) would be needed, or an SVG/canvas fallback.

### 4. ThumbColor Not Rendered (High)

`ThumbColor` is entirely absent from the simulation pipeline. The hard-coded `accent-color: #2563eb` CSS on `.sim-slider` is used unconditionally. There is no mechanism for blocks or designer values to change the thumb colour.

### 5. NumberOfSteps Not Simulated (High)

`NumberOfSteps` (AI default `100`, controls precision) is absent from all simulation files. The HTML `<input type="range">` uses an implicit `step="1"`. When `NumberOfSteps` is not `(MaxValue - MinValue)`, the position values reported by `PositionChanged` will differ from App Inventor. For example with `MinValue=0, MaxValue=150, NumberOfSteps=1000`, AI would report positions in increments of `0.15`, but Tensor would report integer positions from `0` to `150` (step 1).

### 6. ThumbEnabled Visual Behaviour (Medium)

In the AI source, `ThumbEnabled = false` sets the thumb alpha to `0` (invisible) and consumes touch events, but the track remains visible and styled normally. In Tensor, `disabled` is set on the `<input>` element which makes the entire control render at `opacity: 0.55` (via `.sim-slider:disabled { opacity: 0.55 }`). This is visually different — the track should remain fully opaque and visible; only the thumb should disappear.

### 7. Height Default Expands Slider (High)

The default `Height: -2` causes `sizeStyle('Height')` to emit `height: 100%;`, stretching the slider to fill its parent's full height. Android's `SeekBar` has a small, fixed intrinsic height (typically 32 dp / ~32 px) and users cannot change it. The Tensor default should be `-1` (auto / wrap content).

### 8. No `step` Attribute Derived from NumberOfSteps (Medium)

Even if `NumberOfSteps` were tracked, the HTML `<input type="range">` element needs a `step` attribute set to `(MaxValue - MinValue) / NumberOfSteps` to reproduce the correct discrete positions. Without this, `PositionChanged` will emit values at integer resolution by default.

---

## Bugs Found

### Bug 1 — `Height: -2` stretches Slider to fill container

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 47

```js
Slider: { Visible: true, Enabled: true, Width: -1, Height: -2, MinValue: 10, MaxValue: 50, ThumbPosition: 30, ThumbEnabled: true },
```

`Height: -2` maps to `height: 100%` in `sizeStyle()`. A Slider inside any container will grow to fill it. Should be `Height: -1`.

**Priority: High**

---

### Bug 2 — `min` fallback `|| 0` masks intentional `MinValue = 0`

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 289

```svelte
min={Number(props.MinValue) || 0}
```

`Number(0) || 0` evaluates to `0`, which happens to be correct, but `Number(false) || 0` and `Number(null) || 0` also give `0`. The real problem is that if `MinValue` is not yet in state (e.g. before designer props are applied) and the simulation default is already `10`, this line would produce `0` instead of `10` if defaults are not merged before the component renders. The pattern should be `props.MinValue != null ? Number(props.MinValue) : 0`. The same applies to the `max` and `value` attributes.

**Priority: Low**

---

### Bug 3 — `ThumbPosition` not clamped on programmatic set

**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go` — `SetProperty`

When blocks execute `set Slider1.ThumbPosition to 999`, the Go host stores `999` in state unconditionally. The visual HTML input will show the clamped position, but `get Slider1.ThumbPosition` will return `999`. App Inventor always clamps to `[MinValue, MaxValue]`. This causes a state desync.

**Priority: Critical**

---

### Bug 4 — `MinValue`/`MaxValue` not mutually clamped

**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go` — `SetProperty`

Setting `MinValue > MaxValue` (or vice versa) via blocks leaves the slider with an inverted range. The HTML `<input type="range">` with `min > max` is undefined behaviour in browsers (Chromium treats it as `max = min`). In App Inventor the other bound is adjusted automatically. No Slider-specific handling exists in `SetProperty`.

**Priority: High**

---

### Bug 5 — `ColorLeft` and `ColorRight` silently accepted but never rendered

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 283–297; `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, lines 92–93

Both properties are in `SIMULATION_VISUAL_PROPS` (state changes trigger re-renders) but the Slider `{:else if node.type === 'Slider'}` branch never reads `props.ColorLeft` or `props.ColorRight` in its `style` or anywhere else. Blocks that set these properties will produce no visible change, with no unsupported warning emitted.

**Priority: High**

---

### Bug 6 — `ThumbEnabled = false` grays out entire slider instead of hiding thumb

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 292

```svelte
disabled={!enabled || props.ThumbEnabled === false}
```

Setting `ThumbEnabled = false` via blocks or designer sets the HTML `disabled` attribute, which applies the `.sim-slider:disabled { opacity: 0.55 }` rule, dimming the entire control. App Inventor makes only the thumb invisible (alpha `0`) while the track retains its normal appearance. The correct simulation would hide the thumb with CSS (e.g. `appearance: none` + custom thumb pseudo-element with `opacity: 0`) while keeping the track visible at full opacity.

**Priority: Medium**

---

### Bug 7 — `ThumbColor` missing from `SIMULATION_VISUAL_PROPS`

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`

`ThumbColor` does not appear in `SIMULATION_VISUAL_PROPS`. If blocks set `Slider1.ThumbColor`, the value is stored in the Go backend state via `setProperty` (no special handling), but the state patch will not trigger a visual re-render because the property is not in the watched set. Additionally, there is no rendering code to use it even if it were tracked.

**Priority: High**

---

## Android / App Inventor Standards Compliance

| Aspect | Status | Notes |
|---|---|---|
| Default numeric values (MinValue=10, MaxValue=50, ThumbPosition=30) | Correct | Match `Component.SLIDER_MIN_VALUE`, `SLIDER_MAX_VALUE`, `SLIDER_THUMB_VALUE`. |
| Default colour values | Non-compliant | ColorLeft (orange), ColorRight (grey), ThumbColor (dark grey) are not set as defaults. The slider always displays with browser-default track colouring. |
| Height non-configurable | Non-compliant | AI removes the Height designer annotation. Tensor sets `Height: -2` which incorrectly expands the slider. |
| ThumbPosition clamping | Non-compliant | App Inventor enforces `[MinValue, MaxValue]` clamping on every set; Tensor does not. |
| MinValue / MaxValue mutual adjustment | Non-compliant | App Inventor adjusts the opposite bound when a constraint is violated; Tensor does not. |
| NumberOfSteps / step precision | Non-compliant | `NumberOfSteps` property entirely absent. Reported positions will differ from AI when `NumberOfSteps != (MaxValue - MinValue)`. |
| ThumbEnabled visual behaviour | Partially compliant | Interaction is correctly blocked, but the visual presentation (full opacity, invisible thumb) does not match AI (track remains visible). |
| PositionChanged arg type | Compliant | Float value passed as JavaScript Number. |
| TouchDown / TouchUp no-arg events | Compliant | Correctly fire with empty args. |
| Event not fired during programmatic NumberOfSteps change | Not simulated | AI uses `notice = false` to suppress `PositionChanged` during `NumberOfSteps` set; moot since `NumberOfSteps` is not supported. |

---

## Summary

The Slider simulation correctly handles the three core numeric properties (`MinValue`, `MaxValue`, `ThumbPosition`), all three events (`PositionChanged`, `TouchDown`, `TouchUp`), and the `Enabled` / `Visible` / `ThumbEnabled` (interaction-blocking) flags. The essential drag-and-report loop works.

However, there are significant gaps:

1. **ThumbPosition is not clamped** (Critical, Bug 3) — programmatic sets outside `[MinValue, MaxValue]` corrupt stored state relative to what App Inventor would hold.
2. **ColorLeft, ColorRight, and ThumbColor are not applied to the rendered element** (High, Bugs 5 & 7) — the slider track and thumb colours are always browser-default blue, ignoring all three colour properties. App Inventor's distinctive orange/grey/dark-grey defaults are never shown.
3. **Height default `-2` stretches the slider** (High, Bug 1) — should be `-1` so the control uses its natural intrinsic height.

**Top 3 action items:**

1. Add clamping logic in `simulation_wasm.go` `SetProperty` for `ThumbPosition` (clamp to `[MinValue, MaxValue]`) and mutual-adjustment for `MinValue`/`MaxValue` changes.
2. Add `ColorLeft`, `ColorRight`, and `ThumbColor` defaults to `SIMULATION_DEFAULTS.Slider` and add `ThumbColor` to `SIMULATION_VISUAL_PROPS`; apply all three to the slider via CSS custom properties and a track gradient in the `SimulationComponent.svelte` Slider branch.
3. Change `Height` default for Slider from `-2` to `-1` in `SIMULATION_DEFAULTS`, and add `NumberOfSteps` (default `100`) to defaults, `SIMULATION_VISUAL_PROPS`, and `coerceSimulationValue`, deriving the HTML `step` attribute as `(MaxValue - MinValue) / NumberOfSteps`.
