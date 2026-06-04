# Image Simulator Review

## Overview

The `Image` component is a simple image-display widget in App Inventor (`com.google.appinventor.components.runtime.Image`). It wraps an Android `ImageView` and exposes properties for picture path, rotation, scaling, clickability, alternate text, and a legacy animation helper. It raises a single user event (`Click`) and exposes no return-value methods. The Tensor IDE simulator renders it as an HTML `<button class="sim-image">` element in `SimulationComponent.svelte`, with defaults in `simulation-capabilities.js` and no Image-specific logic in `simulation_wasm.go`. The core picture-display and click interaction are functional, but several properties are missing, one important default is wrong, and the render element choice introduces subtle behaviour gaps.

---

## Properties Analysis

### Supported

- `Picture` — stored in state, resolved to an asset URL via `resolveAssetUrl()`, rendered as `<img src={assetUrl}>`.
- `Clickable` — default `false`; checked in `imageClick()` before firing the `Click` event; also used as the `disabled` attribute on the wrapping `<button>`.
- `RotationAngle` — applied as `transform: rotate(Ndeg)` in inline style; default `0`.
- `Width` / `Height` — handled by the shared `sizeStyle()` helper; size constants `-1` (auto), `-2` (fill), and percentage encoding `≤ -1000` all handled.
- `Visible` — controls whether the element is rendered at all.
- `AlternateText` — read from `props.AlternateText` and used as the `<img alt>` attribute (line 301 of `SimulationComponent.svelte`). This is correctly handled in the frontend.

### Missing / Unsupported

| Property | Priority | Notes |
|---|---|---|
| `ScalePictureToFit` | **High** | AI property (boolean, default `false`). When `true`, sets `ImageView.ScaleType.FIT_XY` (stretch to fill, ignoring aspect ratio); when `false`, uses `FIT_CENTER` (preserve aspect ratio). The simulator CSS always applies `object-fit: contain` (equivalent to `FIT_CENTER`), so `ScalePictureToFit = true` cannot be honoured. Not in defaults, not in `SIMULATION_VISUAL_PROPS`, not rendered. |
| `Scaling` | **Medium** | Deprecated but still a valid `@SimpleProperty`. Integer 0 = proportional (`FIT_CENTER`), 1 = fit (`FIT_XY`). Overlaps `ScalePictureToFit` semantics. Not exposed or simulated. |
| `Animation` | **Low** | `@SimpleProperty` setter that calls `AnimationUtil.ApplyAnimation()` with strings like `ScrollRightSlow`, `ScrollLeft`, `Stop`. Entirely absent. Not feasible to simulate faithfully in CSS without significant effort, but the property should at least be accepted without marking the component as unsupported. |
| `AlternateText` (defaults entry) | **Low** | The `SIMULATION_DEFAULTS` object for `Image` (line 48 of `simulation-capabilities.js`) does not include `AlternateText: ''`, so it is absent from the initial state snapshot. The frontend reads `props.AlternateText` at render time; if the user never sets it the property simply resolves to `undefined`, which then coerces to `'undefined'` in the `alt` attribute via the `|| props.Picture || node.name` fallback chain — giving a misleading `alt` text. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `Enabled` | not an AI property on Image | `true` (via `COMMON_VISIBLE_PROPS`) | **Medium** | `Image` in App Inventor does **not** inherit from `ButtonBase` and does not have an `Enabled` property. The Java source shows no `Enabled` setter; `AndroidViewComponent` does not declare one either. Including `Enabled` in the defaults is harmless but technically incorrect and could allow blocks to appear to set `Enabled` without effect, confusing users. |
| `AlternateText` | `""` (empty string) | missing from defaults | **Low** | See above. |

---

## Events Analysis

### Supported

- `Click` — fired by `imageClick()` (line 129 of `SimulationComponent.svelte`) when the image is visible, not disabled, and `Clickable !== false`. Dispatched via `emitEvent(node.name, 'Click')`. No arguments — matches the AI signature `public void Click()`. Correctly guarded.

### Missing / Incorrect

| Event | Priority | Notes |
|---|---|---|
| `Click` guard uses `!enabled` | **Medium** | `imageClick()` checks `!enabled` (line 128), but `Enabled` is not a real Image property in App Inventor. The effective guard should only be `!visible` and `props.Clickable === false`. As a result, if a blocks program sets `Image1.Enabled = false` (which would be accepted because `Enabled` is in the state/defaults), the click is silently suppressed even though App Inventor has no such behaviour. |
| No event args registered in `EVENT_ARGS` | **Low** | `Click` takes no arguments, so no entry in `EVENT_ARGS` is needed. This is fine. |

---

## Methods Analysis

### Supported

The `Image` component defines no `@SimpleFunction` methods in the AI source. All interaction is purely through properties and the single event. The absence of any Image-specific method handling in `simulation_wasm.go` is therefore **correct**.

### Missing / Incorrect

None — the AI Java source contains zero `@SimpleFunction` methods for Image.

---

## Behaviour Gaps

### 1. Render element is `<button>` regardless of `Clickable` (High)
The Image is always rendered as `<button class="sim-image">`. When `Clickable = false`, the button is `disabled`, which suppresses the click correctly, but the element still renders with button styling (border, background, cursor changes on hover in some browsers). In App Inventor, a non-clickable Image is a plain view with no interactive affordance. This can mislead users into thinking the image is always a button-like control.

### 2. `ScalePictureToFit` not simulated — hardcoded `object-fit: contain` (High)
The CSS rule `.sim-image img { object-fit: contain; }` (line 471) is unconditional. When a designer sets `ScalePictureToFit = true`, the image should stretch to fill the container (`object-fit: fill` or `object-fit: cover`). This cannot be toggled at runtime because the property is not tracked in state.

### 3. Asset resolution falls back to empty string — no placeholder feedback (Medium)
`resolveAssetUrl()` returns `''` when the asset is not found in the `assets` array. The template then renders `<span>{props.Picture || 'Image'}</span>` as a fallback (line 303). This is a reasonable degraded state, but the fallback text always says the raw filename or the string `'Image'`, which may not help users debug missing assets.

### 4. `RotationAngle` applied to the outer `<button>`, not the `<img>` (Medium)
The `transform: rotate()` style is applied to the entire `<button>` element (line 299), including its border and padding. In Android, `HoneycombUtil.viewSetRotate` rotates the `ImageView`'s drawing matrix around the view's center. This is equivalent to rotating the whole element, so visually it is close — but the dashed border of the sim element also rotates, which can look odd and does not match the Android appearance where only the image content rotates inside the view bounds.

### 5. `Animation` property silently unsupported (Low)
If a user program calls `set Image1.Animation to "ScrollRight"`, the property will reach `SetProperty` in the Go host, which will call `h.setProperty(componentName, "Animation", value)`, storing it in state. The frontend does not react to `Animation` changes (not in `SIMULATION_VISUAL_PROPS`), so no CSS animation is applied and no unsupported marker is shown. The behaviour is silent no-op rather than a clearly flagged unsupported operation.

### 6. `AlternateText` fallback in `alt` attribute (Low)
Line 301: `alt={props.AlternateText || props.Picture || node.name}`. If `AlternateText` is `undefined` (because it is missing from defaults), the expression evaluates through `props.Picture` (the filename) and finally `node.name` (the component name). This means the `alt` attribute of the rendered image is never empty, which violates the AI default of an empty alternate-text description and may affect screen-reader simulation fidelity.

---

## Bugs Found

### Bug 1 — `alt` attribute shows `"undefined"` string when `AlternateText` is set to the literal string `"undefined"` by a block (Low)
The expression `props.AlternateText || props.Picture || node.name` uses JavaScript truthiness. If a block program sets `AlternateText` to an empty string `""`, the empty string is falsy, so the fallback will continue to `props.Picture`. The intended AI behaviour is that `setContentDescription("")` should clear the description — meaning `alt=""` (empty). This is a correctness bug for the empty-string case.

**Location:** `SimulationComponent.svelte` line 301.

**Fix:** Use `props.AlternateText != null ? props.AlternateText : (props.Picture || '')` so an explicit empty string is respected.

### Bug 2 — `Clickable` coercion in `imageClick()` uses strict equality `=== false` (Medium)
`imageClick()` at line 128:
```js
if (!enabled || !visible || props.Clickable === false) return;
```
The coercion step in `coerceSimulationValue()` normalises `Clickable` to a boolean, so `props.Clickable` should always be `true` or `false` once set. However, the **default** in `SIMULATION_DEFAULTS` is `Clickable: false` (correct), and the coercion path handles it. The bug is more subtle: if `Clickable` is never set at runtime (value stays as the raw designer string `"False"`), `coerceSimulationValue` is only called during `deriveStateFromDesignerProps` — if that path is skipped, the string `"False"` is stored directly, and `"False" === false` is `false`, so the guard is bypassed and clicks fire even when `Clickable = false`. This is an edge-case correctness issue depending on initialisation path.

**Location:** `SimulationComponent.svelte` line 128; `simulation-capabilities.js` coercion path.

**Fix:** Change the guard to use the coercion helper: `!props.Clickable` (since coerced value is boolean) or explicitly check `props.Clickable !== true`.

### Bug 3 — `Enabled` in `COMMON_VISIBLE_PROPS` incorrectly applied to Image (Medium)
`Image` shares `COMMON_VISIBLE_PROPS` which includes `Enabled: true`. App Inventor's `Image` component does not define or inherit an `Enabled` property. This means the Tensor simulator accepts `Image1.Enabled` as a valid property in state, tracks it, and uses it to suppress clicks — none of which reflects AI behaviour.

**Location:** `simulation-capabilities.js` line 48 (`Image: { ...COMMON_VISIBLE_PROPS, ... }`).

**Fix:** Remove `Enabled` from the Image defaults object by not spreading `COMMON_VISIBLE_PROPS`, or create a separate `IMAGE_VISIBLE_PROPS = { Visible: true, Width: -1, Height: -1 }`.

---

## Android/App Inventor Standards Compliance

| Standard | Compliance | Notes |
|---|---|---|
| Size encoding (`-1` auto, `-2` fill, `≤ -1000` percent) | Compliant | `sizeStyle()` handles all three conventions correctly. |
| Color encoding (`&HAArrggbb`) | Compliant | `colorValue()` correctly strips the alpha byte and converts to CSS hex. |
| `Picture` default is `""` | Compliant | Both AI and Tensor default to empty string. |
| `RotationAngle` default is `0.0` | Compliant | Tensor default is `0` (numeric). |
| `Clickable` default is `false` | Compliant | Tensor default is `false`. |
| `ScalePictureToFit` default is `false` (preserve aspect) | Partially compliant | CSS uses `object-fit: contain` which matches the default `false` case, but the `true` case is unsimulated. |
| `AlternateText` default is `""` | Non-compliant | Missing from defaults; `alt` attribute never resolves to empty. |
| No `Enabled` property on Image | Non-compliant | `Enabled` is incorrectly included via `COMMON_VISIBLE_PROPS`. |
| `Click` event fires only when `Clickable = true` | Compliant (mostly) | Correct guard in `imageClick()`, though see Bug 2. |
| Asset loading from project assets | Compliant | `resolveAssetUrl()` correctly matches by name or basename. |

---

## Summary

The Image simulator covers the essential display use-case (showing an image asset with optional rotation and click handling) and is usable for basic scenarios. However, several correctness issues reduce fidelity to the App Inventor specification.

**Top 3 action items:**

1. **[High] Remove `Enabled` from Image defaults and the `imageClick` guard.** App Inventor's `Image` has no `Enabled` property. Spreading `COMMON_VISIBLE_PROPS` (which includes `Enabled: true`) onto the Image defaults is wrong. The `imageClick()` function should not check `!enabled`. Fix: create a dedicated `IMAGE_VISIBLE_PROPS` without `Enabled`, and remove the `!enabled` guard from `imageClick()`.

2. **[High] Add `ScalePictureToFit` support.** This is a designer-visible property that directly controls how the image is displayed. Add `ScalePictureToFit: false` to `SIMULATION_DEFAULTS.Image`, add `'ScalePictureToFit'` to `SIMULATION_VISUAL_PROPS`, and in the `<img>` tag toggle between `object-fit: contain` (false) and `object-fit: fill` (true) based on `props.ScalePictureToFit`.

3. **[Medium] Fix the `alt` attribute empty-string bug.** Change `alt={props.AlternateText || props.Picture || node.name}` to `alt={props.AlternateText != null ? props.AlternateText : ''}` so that an explicit empty `AlternateText` clears the description as App Inventor intends (matching `setContentDescription("")`), and add `AlternateText: ''` to `SIMULATION_DEFAULTS.Image`.
