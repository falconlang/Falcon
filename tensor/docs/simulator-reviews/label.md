# Label Simulator Review

## Overview

The Label component in App Inventor is a read-only text display widget backed by Android's `TextView`. It supports rich styling (font face, size, bold, italic, alignment, margins) and optional HTML rendering via `HTMLFormat`. There are no user-interaction events on Label — it is a purely display component.

The Tensor IDE simulator renders Labels as a `<div class="sim-label">` element. Property state is driven by `simulation-capabilities.js` defaults and coercion, visual rendering is done in `SimulationComponent.svelte`, and runtime property get/set goes through `simulation_wasm.go`.

---

## Properties Analysis

### Supported

- `Text` — rendered as inner text of the label `<div>`. Default is `''`. Correctly supported.
- `Visible` — controls whether the element is shown at all (via `isSimulationVisible`). Correct.
- `Width` / `Height` — handled by `sizeStyle()` with -1 (auto), -2 (fill parent), -1000x (percent) and positive pixel encoding. Correct infrastructure.
- `TextColor` — applied via `baseStyle()` as CSS `color`. Correct.
- `FontSize` — applied via `baseStyle()` as CSS `font-size` in `px`. Correct (with caveat — see Behaviour Gaps).
- `BackgroundColor` — applied via `baseStyle()` as CSS `background`. Correct.

### Missing / Unsupported

| Property | Priority | Notes |
|---|---|---|
| `FontBold` | High | Not in defaults, not in `SIMULATION_VISUAL_PROPS`, not applied to style. Bold text is never rendered. |
| `FontItalic` | High | Same gap as FontBold — italic is never applied. |
| `FontTypeface` | Medium | Not tracked. AI supports Default / Serif / SansSerif / Monospace mapped to CSS font-family. |
| `TextAlignment` | Medium | Not in defaults, not in `SIMULATION_VISUAL_PROPS`, not applied via `text-align`. Default is ALIGNMENT_NORMAL (left), so it works by coincidence, but centre/right alignments silently fail. |
| `HasMargins` | Low | Not in defaults, not rendered. AI default is `true` (2 dp margin on all sides). The simulator has a fixed `gap: 8px` on parent arrangements which partially substitutes but is not equivalent. |
| `HTMLFormat` | Medium | Not tracked. When `HTMLFormat` is true AI renders the Text value as HTML; the simulator always renders it as plain text, stripping all markup. |
| `HTMLContent` | Low | Read-only getter in AI — returns the raw HTML content. Not exposed in simulator, but only matters when HTMLFormat is active. |
| `Enabled` | Low | Present in `COMMON_VISIBLE_PROPS` default and `coerceSimulationValue` boolean list, but Label in AI has no `Enabled` property (it extends `AndroidViewComponent` but Label never uses it). Including it is harmless but inaccurate. |

### Wrong Defaults or Types

| Property | Tensor Default | AI Default | Priority |
|---|---|---|---|
| `TextColor` | `&HFF222222` (near-black) | `Component.DEFAULT_VALUE_COLOR_BLACK` = `&HFF000000` (pure black) | Low |
| `BackgroundColor` | not set (no entry in Label defaults) | `Component.COLOR_NONE` = `&H00000000` (transparent) | Medium |
| `FontSize` | `14` (px) | `Component.FONT_DEFAULT_SIZE` = `14.0` sp | Low — value matches but units differ (sp vs px); on high-density displays this would diverge |

---

## Events Analysis

### Supported

None — Label has no events in App Inventor, and none are fired in the simulator. This is correct.

### Missing / Incorrect

None. Label defines no `@SimpleEvent` methods in the AI source. The simulator correctly fires no events for Label interactions.

---

## Methods Analysis

### Supported

None — Label has no callable `@SimpleFunction` methods in the AI source. The simulator correctly treats all Label method calls as unsupported (falls through to `h.Unsupported()`).

### Missing / Incorrect

None. No methods to implement.

---

## Behaviour Gaps

### 1. HTMLFormat not simulated (High)
When `HTMLFormat` is `true`, App Inventor calls `TextViewUtil.setTextHTML()` which parses basic HTML tags (`<b>`, `<i>`, `<u>`, `<br>`, `<font color>`, etc.) via `Html.fromHtml()`. The simulator renders `props.Text` as a raw text node, so any HTML tags appear as literal characters and markup is never applied.

### 2. FontBold / FontItalic ignored (High)
The simulator's `baseStyle()` function has no branch for `FontBold` or `FontItalic`. Because `SIMULATION_VISUAL_PROPS` does not include these keys, they are also stripped in `deriveStateFromDesignerProps()` — they never reach the runtime state at all. This means a Label designed with bold or italic text will always appear in normal weight/style.

### 3. TextAlignment not applied (Medium)
`TextAlignment` is not in `SIMULATION_VISUAL_PROPS` so it is dropped during designer prop derivation. The CSS `text-align` rule is never emitted. AI values are 0 = normal (left), 1 = centre, 2 = opposite (right). A centre- or right-aligned label in the designer will render left-aligned in the simulator.

### 4. FontSize units: sp vs px (Low)
App Inventor specifies `FontSize` in scale-independent pixels (sp). The simulator converts it directly to CSS `px` via `font-size: ${Number(props.FontSize) || 14}px`. On standard 1× screens this is equivalent, but the two units diverge on high-density or accessibility-scaled screens. A faithful simulation would use CSS `rem` or simulate the sp→px scaling factor.

### 5. BackgroundColor default discrepancy (Medium)
The `SIMULATION_DEFAULTS` for Label does not include a `BackgroundColor` entry, so the property starts as `undefined`. In `baseStyle()` the condition `props.BackgroundColor ? ...` evaluates falsy for `undefined`, leaving the label with the browser's default background (typically `transparent`). This happens to match the AI default (`COLOR_NONE` = transparent), but only by accident; if a designer explicitly sets `BackgroundColor` to `&H00000000` (still transparent) the code path behaves differently. More importantly, if the runtime resets the property to the AI default, the simulator will not clear a previously applied background.

### 6. TextColor: COLOR_DEFAULT theme-aware behaviour not replicated (Low)
In AI, when `TextColor` is `COLOR_DEFAULT`, the text color is resolved at runtime based on whether the form is in dark theme (`COLOR_WHITE`) or light theme (`COLOR_BLACK`). The simulator always emits the literal hex value `&HFF222222` regardless of any theme setting.

### 7. HasMargins default not simulated (Low)
AI defaults `HasMargins` to `true`, applying a 2 dp margin around each label. The simulator has no `HasMargins` state, so the margin is always absent. Because parent arrangements use `gap: 8px`, spacing between labels is usually sufficient visually, but the per-label margin is missing and cannot be toggled.

### 8. FontTypeface not mapped (Medium)
App Inventor supports four typefaces: Default (sans-serif), Serif, SansSerif, Monospace. The simulator does not track `FontTypeface` and always renders in the inherited CSS font (which is sans-serif). Labels using a serif or monospace typeface will silently appear in the wrong font.

### 9. BigDefaultText / LargeFont accessibility mode not simulated (Low)
AI has an `AccessibleComponent` interface: when `isBigText` or `form.BigDefaultText()` is true, `FontSize` at the default (14 sp) is automatically raised to 24 sp. The simulator has no concept of accessibility/large-font mode.

---

## Bugs Found

### Bug 1: FontBold/FontItalic state never stored (High)
**Location:** `simulation-capabilities.js` line 41 (Label defaults) and `SIMULATION_VISUAL_PROPS` set (lines 63-100).

`FontBold` and `FontItalic` are absent from both `SIMULATION_DEFAULTS.Label` and `SIMULATION_VISUAL_PROPS`. Consequently, `deriveStateFromDesignerProps()` silently discards these properties when building initial simulator state. Even if the Go runtime calls `SetProperty(..., "FontBold", ...)` the value is stored in the backing Go map but is never forwarded to the frontend as a visual prop update (since `SIMULATION_VISUAL_PROPS` gates what triggers re-render in `SimulateOverlay.svelte`). Bold/italic are permanently lost.

### Bug 2: TextAlignment silently dropped (Medium)
**Location:** `simulation-capabilities.js` — `TextAlignment` is absent from `SIMULATION_VISUAL_PROPS`.

`deriveStateFromDesignerProps()` skips any key not in `SIMULATION_VISUAL_PROPS`. A Label with `TextAlignment: 1` (centre) in the designer will have that property stripped; the simulator state will not contain it; `baseStyle()` never emits `text-align`; the label renders left-aligned regardless of design intent.

### Bug 3: colorValue() drops the alpha channel (Medium)
**Location:** `SimulationComponent.svelte` lines 44-47, `colorValue()` function.

```js
const ai = text.match(/^&H([0-9A-Fa-f]{2})([0-9A-Fa-f]{6})$/);
if (ai) return `#${ai[2]}`;
```

The regex captures the alpha byte in `ai[1]` but discards it, always returning a fully-opaque 6-digit hex colour. This means semi-transparent background or text colours defined in the designer (e.g. `&H80FF0000` = 50% transparent red) will render as fully opaque in the simulator. This bug affects all colour properties on all components, not just Label.

### Bug 4: Label renders empty string as blank but Text prop could be a non-string (Low)
**Location:** `SimulationComponent.svelte` line 257.

```svelte
{props.Text || ''}
```

If `props.Text` is the number `0` or boolean `false` (valid after a block sets `Label.Text = 0`), the `||` fallback will replace it with `''` and render nothing. The correct idiom is `{props.Text ?? ''}` (nullish coalescing) or `{String(props.Text ?? '')}`.

### Bug 5: FontSize falls back to 14 for falsy values, hiding zero-size labels (Low)
**Location:** `SimulationComponent.svelte` line 66 inside `baseStyle()`.

```js
props.FontSize ? `font-size: ${Number(props.FontSize) || 14}px;` : '',
```

If `FontSize` is `0` (a block explicitly setting size to zero), the outer `props.FontSize` check is falsy and no `font-size` rule is emitted, so the label falls back to the inherited CSS font size rather than disappearing. This is a minor visual difference from the AI behaviour where `FontSize(0)` sets the TextView text size to 0 sp.

---

## Android/App Inventor Standards Compliance

| Concern | Status |
|---|---|
| Color format `&HAARRGGBB` → CSS | Partial — alpha channel is silently dropped (Bug 3). |
| Size constants -1 (auto) / -2 (fill) / -1000x (percent) | Correctly handled in `sizeStyle()`. |
| sp vs px font units | Treated as equivalent (diverges on non-1× screens). |
| HTML rendering via `Html.fromHtml()` | Not implemented; plain text only. |
| TextAlignment enum (0/1/2) | Not applied. |
| Font typeface enum | Not tracked or applied. |
| Margin system (2 dp default) | Not implemented. |
| Dark theme / COLOR_DEFAULT resolution | Not implemented; hardcoded near-black. |
| Accessibility large-font mode | Not implemented. |
| `Enabled` property | Included in defaults but Label does not use it in AI. |

---

## Summary

The Tensor IDE simulator implements the bare minimum for Label: it renders `Text`, respects `Visible`, and applies `TextColor`, `BackgroundColor`, `FontSize`, `Width`, and `Height`. This covers the most common use-case well.

However, several important styling properties are silently lost, and a cross-cutting bug (alpha channel in colour values) affects the fidelity of all colour properties.

**Top 3 Action Items:**

1. **Add `FontBold`, `FontItalic`, `TextAlignment`, `FontTypeface` to `SIMULATION_VISUAL_PROPS` and apply them in `baseStyle()`** (High) — these are visible in every Label the designer touches and the silent discard makes the simulator misleading.

2. **Fix `colorValue()` to honour the alpha channel** (Medium/High) — `&H80RRGGBB` colours must produce `rgba(r, g, b, a/255)` CSS; the current implementation makes all semi-transparent colours fully opaque, affecting every colour property on every component.

3. **Implement `HTMLFormat` rendering** (Medium) — use `innerHTML` when `HTMLFormat` is `true` so that styled HTML Labels display correctly in the simulator.
