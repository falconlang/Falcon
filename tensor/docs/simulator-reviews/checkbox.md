# CheckBox Simulator Review

## Overview

`CheckBox` in App Inventor extends `ToggleBase<android.widget.CheckBox>`, which itself extends `AndroidViewComponent`. The component is a toggleable boolean widget that fires `Changed`, `GotFocus`, and `LostFocus` events. It exposes a rich set of inherited appearance properties (font styling, background/text color, bold, italic, typeface) in addition to its own `Checked` boolean property.

The Tensor IDE simulation handles `CheckBox` rendering in `SimulationComponent.svelte`, defaults/coercion in `simulation-capabilities.js`, and backend state/method dispatch in `simulation_wasm.go`. There are no CheckBox-specific method handlers in the WASM backend because CheckBox has no methods; all property and event handling flows through the shared infrastructure.

---

## Properties Analysis

### Supported

- `Visible` — supported via `COMMON_VISIBLE_PROPS`, correctly defaults to `true`.
- `Enabled` — supported via `COMMON_VISIBLE_PROPS`, correctly defaults to `true`. Rendered as `disabled={!enabled}` on the `<input>`.
- `Width` / `Height` — inherited from `COMMON_VISIBLE_PROPS`, both default to `-1` (fill-parent/automatic). Passed through `sizeStyle()`.
- `Text` — supported, defaults to `''`. Rendered as `<span>{props.Text || ''}</span>`.
- `Checked` — supported, defaults to `false`. Correct per AI source (`defaultValue = "False"`). Coercion via `coerceSimulationValue` handles string `"true"/"false"` and numeric `1/0`.
- `TextColor` — supported, defaults to `'&HFF222222'`. Rendered via `colorValue()` in `baseStyle()`.
- `FontSize` — supported, defaults to `14`. Rendered as `font-size: Xpx` in `baseStyle()`.
- `BackgroundColor` — supported in `baseStyle()` for visual rendering.

### Missing / Unsupported

| Property | AI Default | Priority | Notes |
|---|---|---|---|
| `FontBold` | `false` | Medium | Declared in `ToggleBase` with `@DesignerProperty`. Not in `SIMULATION_DEFAULTS`, `SIMULATION_VISUAL_PROPS`, or rendered in `baseStyle()`. Blocks are able to set `FontBold` but it will never affect the rendered checkbox label. |
| `FontItalic` | `false` | Medium | Same as `FontBold` — declared in `ToggleBase`, not tracked or rendered. |
| `FontTypeface` | `"0"` (default) | Low | Designer-settable typeface (default/serif/sans-serif/monospace). Not tracked or applied. |
| `WidthPercent` | n/a (method) | Low | `WidthPercent` and `HeightPercent` are settable methods on `AndroidViewComponent`. The `-1000x` percent encoding is supported in `sizeStyle()`, but there is no designer-facing `WidthPercent`/`HeightPercent` property entry. This is consistent across all components and is not specific to CheckBox. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `TextColor` | `Component.COLOR_BLACK` = `&HFF000000` | `'&HFF222222'` | Low | The AI source defaults `TextColor` to pure black (`#FF000000`). Tensor uses a slightly softened dark (`#222222`). Visually near-identical but not spec-faithful. This is a shared deviation across Label, TextBox, etc. |
| `BackgroundColor` | `Component.COLOR_NONE` (transparent) | Not in `SIMULATION_DEFAULTS` for CheckBox | Low | `ToggleBase.initToggle()` sets `BackgroundColor(Component.COLOR_NONE)`. Tensor has no `BackgroundColor` key in the `CheckBox` defaults object, so it falls back to the CSS inherited background (white/transparent). Functionally acceptable but not explicit. |

---

## Events Analysis

### Supported

- `Changed` — fired by `checkboxInput()` via `emitInteraction(...)` with `{ event: 'Changed', args: [] }`. The WASM dispatcher receives this and runs any registered `Changed` handler. This is correct: `ToggleBase.onCheckedChanged` calls `Changed()` with no arguments.

### Missing / Incorrect

| Event | AI Signature | Tensor Status | Priority | Notes |
|---|---|---|---|---|
| `GotFocus` | `GotFocus()` — no args | Missing for CheckBox | High | `ToggleBase` declares `GotFocus()` as a `@SimpleEvent`. The HTML `<input type="checkbox">` inside `<label>` does receive focus, but `SimulationComponent.svelte` does not attach any `on:focus` handler to the checkbox input. The TextBox block correctly emits `GotFocus`, but the checkbox branch (`{:else if node.type === 'CheckBox' || node.type === 'Switch'}`) has no focus handler. |
| `LostFocus` | `LostFocus()` — no args | Missing for CheckBox | High | Same issue as `GotFocus`. No `on:blur` handler attached to the checkbox `<input>`. Users who write `CheckBox1.LostFocus` blocks will never see them fire. |

---

## Methods Analysis

### Supported

There are no `@SimpleFunction` methods on `CheckBox` or `ToggleBase`. All interactions are property-based. This is correctly reflected — `simulation_wasm.go`'s `CallMethod` has no `CheckBox` case, and any method call would fall through to `h.Unsupported(...)`, which is the correct behaviour.

### Missing / Incorrect

None — the component has no callable methods.

---

## Behaviour Gaps

### 1. `Changed` event fires without guard on `enabled` state (High)

In `checkboxInput()`, the interaction is dispatched without checking whether the component is `enabled`. However, the `<input>` element receives `disabled={!enabled}`, which prevents the browser from firing `change` events on a disabled checkbox. This is accidentally correct for pointer-initiated changes, but if `Checked` is set programmatically via block code, the Android `OnCheckedChangeListener` also fires `Changed`. Tensor's `SetProperty` path (`simulation_wasm.go` line 100 → `setProperty`) updates state directly without firing an event — this is correct.

### 2. `Changed` event passes no new checked value to the event handler (Medium)

In App Inventor the `Changed` event takes no arguments (the block handler reads `CheckBox1.Checked` to get the new value). Tensor fires `Changed` with `args: []`, which is correct. However, the `emitInteraction` call in `checkboxInput` updates state (setting `Checked` to the new boolean) before the event fires. This matches AI behaviour where `view.isChecked()` already returns the new state by the time the handler runs.

### 3. No `GotFocus` / `LostFocus` events on the checkbox input (High)

The `<label>` wrapper receives focus indirectly through its nested `<input type="checkbox">`. The AI spec (via `ToggleBase.onFocusChange`) fires `GotFocus` and `LostFocus`. Because no `on:focus`/`on:blur` handler is attached to the nested checkbox `<input>`, these events are silently dropped. Compare with TextBox (lines 271–272 of `SimulationComponent.svelte`) which correctly wires `on:focus` and `on:blur`.

### 4. `FontSize` unit: sp vs px (Low)

App Inventor uses sp (scale-independent pixels) for font sizing. Tensor renders `font-size: ${FontSize}px` using CSS pixels. In a desktop simulator context this is acceptable, but it means the rendered size is not density-scaled. This is a systemic issue across all text components, not specific to CheckBox.

### 5. `BackgroundColor` default is `COLOR_NONE` (transparent) in AI (Low)

`ToggleBase.initToggle()` calls `BackgroundColor(Component.COLOR_NONE)`. Tensor does not include `BackgroundColor` in `SIMULATION_DEFAULTS.CheckBox`, so the background defaults to whatever the CSS cascade provides. In practice the checkbox label background will be transparent, which matches `COLOR_NONE`. This is functionally correct but could become wrong if a parent sets an explicit background and the user hasn't set `BackgroundColor` on the checkbox.

### 6. `FontBold` and `FontItalic` not tracked in `SIMULATION_VISUAL_PROPS` (Medium)

`deriveStateFromDesignerProps()` only copies keys that appear in `SIMULATION_VISUAL_PROPS`. `FontBold` and `FontItalic` are absent from this set. If a designer sets bold/italic via the properties panel, the value is not carried into the simulation state, and the label text will always render in normal weight/style.

---

## Bugs Found

### Bug 1: `GotFocus` and `LostFocus` events never fire for CheckBox

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 278–282

```svelte
{:else if node.type === 'CheckBox' || node.type === 'Switch'}
  <label class="sim-check" ...>
    <input type="checkbox" checked={...} disabled={!enabled} on:change={checkboxInput} />
    <span>{props.Text || ''}</span>
  </label>
```

The `<input>` has only `on:change`. It needs `on:focus` and `on:blur` handlers equivalent to those on the TextBox (lines 271–272). Without them, any blocks that use `CheckBox1.GotFocus` or `CheckBox1.LostFocus` are dead code in the simulator.

**Fix:** Add `on:focus={() => emitEvent(node.name, 'GotFocus')}` and `on:blur={() => emitEvent(node.name, 'LostFocus')}` to the `<input type="checkbox">` element.

### Bug 2: `FontBold` / `FontItalic` designer values silently dropped

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 63–100 (`SIMULATION_VISUAL_PROPS`)

`FontBold` and `FontItalic` are absent from `SIMULATION_VISUAL_PROPS`. The `deriveStateFromDesignerProps` function skips any key not in this set (line 155), so designer-configured bold/italic never enters simulation state and the `baseStyle()` function in `SimulationComponent.svelte` never applies them, since it only reads `FontSize` and `TextColor`, not bold/italic flags.

### Bug 3: `TextColor` default diverges from App Inventor spec

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 45

```js
CheckBox: { ...COMMON_VISIBLE_PROPS, Text: '', Checked: false, TextColor: '&HFF222222', FontSize: 14 },
```

App Inventor's `ToggleBase` sets `TextColor` default to `Component.DEFAULT_VALUE_COLOR_BLACK` = `&HFF000000`. The simulator uses `&HFF222222` (a soft near-black). While visually similar, blocks that read `CheckBox1.TextColor` and compare against the App Inventor constant `BLACK` will get a mismatch.

---

## Android/App Inventor Standards Compliance

| Aspect | Status |
|---|---|
| `Checked` default (`false`) | Correct — matches `Checked(false)` in constructor |
| `Text` default (`""`) | Correct — matches `Text("")` in `initToggle()` |
| `Enabled` default (`true`) | Correct — matches `Enabled(true)` in `initToggle()` |
| `Changed` event name and arg count | Correct — zero arguments |
| `GotFocus` / `LostFocus` events | Not implemented — hardware-level focus events are absent |
| `FontBold` / `FontItalic` | Not tracked or rendered |
| `BackgroundColor` default (transparent/none) | Functionally correct via CSS cascade |
| `TextColor` default (black) | Off by `&HFF000000` vs `&HFF222222` |
| Size encoding (`-1` fill, `-2` wrap, `-1000x` percent) | Correctly handled in `sizeStyle()` |
| `FontTypeface` (0=default, 1=serif, 2=sans, 3=mono) | Not implemented |
| `WidthPercent` / `HeightPercent` percent setters | Supported via `-1000x` encoding in `sizeStyle()` |

---

## Summary

The CheckBox simulator is functional for the most common use-cases (rendering, toggling, `Changed` event dispatch). The `Checked` property, `Text`, `Enabled`, `Visible`, `Width`/`Height`, `TextColor`, and `FontSize` all work correctly.

The most impactful gaps are the missing `GotFocus` and `LostFocus` events, which are declared in `ToggleBase` and will silently fail for any app that uses focus-driven logic on a CheckBox. Font styling properties (`FontBold`, `FontItalic`, `FontTypeface`) are also absent from the simulation layer, causing designer configurations to be dropped.

**Top 3 action items:**

1. **[Critical] Add `on:focus` and `on:blur` to the checkbox `<input>` element** in `SimulationComponent.svelte` to fire `GotFocus` and `LostFocus` events — identical to how TextBox already handles focus.
2. **[High] Add `FontBold` and `FontItalic` to `SIMULATION_VISUAL_PROPS`** and apply them in `baseStyle()` (e.g. `font-weight: bold` and `font-style: italic`) so designer-configured styling is not silently dropped during simulation.
3. **[Medium] Correct `TextColor` default** from `&HFF222222` to `&HFF000000` in `SIMULATION_DEFAULTS.CheckBox` (and the other components sharing this off-spec default) to ensure block-level color comparisons work correctly.
