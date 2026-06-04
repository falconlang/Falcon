# TextBox Simulator Review

## Overview

TextBox is a single-line (optionally multi-line) text input component. In App Inventor it extends `TextBoxBase`, which in turn extends `AndroidViewComponent`. The component wraps an Android `EditText` widget and inherits a rich set of properties (font, alignment, colors, hint), events (GotFocus, LostFocus, TextChanged), and methods (HideKeyboard, RequestFocus, MoveCursorTo, MoveCursorToStart, MoveCursorToEnd).

The Tensor IDE simulator renders TextBox as an HTML `<input type="text">` element in `SimulationComponent.svelte`, with defaults and type coercion in `simulation-capabilities.js`, and backend method dispatch in `simulation_wasm.go`. The scope of this review covers all three layers against the ground-truth Java source (`TextBox.java` + `TextBoxBase.java`).

---

## Properties Analysis

### Supported

- `Text` — rendered as `value` on the `<input>`, updated reactively via `textInput()`. Default `""`. Correct.
- `Hint` — rendered as `placeholder`. Default `""`. Correct.
- `NumbersOnly` — coerced to boolean; sets `inputmode="decimal"` on the input. Default `false`. Correct.
- `Enabled` — coerced to boolean; passed as `disabled` attribute. Default `true`. Correct.
- `Visible` — controls render via `{#if visible}`. Default `true`. Correct.
- `Width` — handled by `sizeStyle()` with -1 (fill), -2 (wrap), -1000x (percent), positive px. Default `160`. Correct value (matches `ComponentConstants.TEXTBOX_PREFERRED_WIDTH = 160`).
- `Height` — handled by `sizeStyle()`. Default `-1` (fill parent). Correct.
- `TextColor` — passed via `baseStyle()` as CSS `color`. Default `&HFF222222`. See wrong defaults section.
- `FontSize` — passed via `baseStyle()` as CSS `font-size` in px. Default `14`. Correct value.
- `BackgroundColor` — passed via `baseStyle()` as CSS `background`. No default set in `SIMULATION_DEFAULTS` for TextBox.

### Missing / Unsupported

- **`MultiLine`** [High] — App Inventor's `MultiLine` property (default `false`) controls whether the textbox accepts multi-line input. In Tensor the HTML `<input>` is always single-line; there is no `<textarea>` fallback and `MultiLine` is not in `SIMULATION_DEFAULTS` nor rendered. Any block that reads or sets `MultiLine` will operate on undefined state, and multi-line textboxes will appear and behave as single-line.
- **`ReadOnly`** [High] — App Inventor's `ReadOnly` property (default `false`) makes the field non-editable while keeping it enabled (different from `Enabled=false` which greys it out). Tensor has no `ReadOnly` property in defaults, coercion list, or visual props set. Setting `ReadOnly=true` via blocks has no visual effect; the input remains editable.
- **`FontBold`** [Medium] — Boolean, default `false`. Not in `SIMULATION_DEFAULTS`, not in `SIMULATION_VISUAL_PROPS`, not applied to any CSS. Block reads/writes are silently ignored.
- **`FontItalic`** [Medium] — Boolean, default `false`. Same gap as `FontBold`.
- **`FontTypeface`** [Low] — String/int typeface selector (default, serif, sans-serif, monospace). Not supported in simulation.
- **`TextAlignment`** [Medium] — Integer (0=normal/left, 1=center, 2=opposite/right), default `0`. Not in `SIMULATION_DEFAULTS`, not in `SIMULATION_VISUAL_PROPS`, not applied to any CSS. Blocks that set `TextAlignment` have no effect.
- **`HintColor`** [Low] — Color of the placeholder/hint text. App Inventor default is `&HFFAAAAAA` (gray). Tensor has no `HintColor` support; placeholder color comes from browser/CSS defaults only.
- **`BackgroundColor`** [Low] — Present in `baseStyle()` and `colorValue()` so it will render if set via blocks, but is absent from `SIMULATION_DEFAULTS` for TextBox, so the initial state is undefined. App Inventor default is `COLOR_DEFAULT` (the Android 3-D shaded look). Tensor should default this explicitly (e.g., `'&H00000000'` or a neutral value) to avoid undefined CSS.

### Wrong Defaults or Types

- **`TextColor` default is wrong** [Medium] — Tensor default is `&HFF222222` (near-black). App Inventor default is `COLOR_DEFAULT` (`&H00000000`), which is a sentinel that resolves to black or white depending on theme. While the rendered colour is visually close, the semantic value is wrong: a block that checks `TextBox.TextColor = Default` would fail to match.
- **`Width` semantic** [Low] — Tensor stores `Width: 160` (a positive pixel value) in `SIMULATION_DEFAULTS`. App Inventor also sets preferred width to 160 via `container.setChildWidth(this, ComponentConstants.TEXTBOX_PREFERRED_WIDTH)`, so the numeric value matches. However, App Inventor expresses this as a layout hint subject to parent constraints, not as a hard pixel lock. In practice the value aligns, but the semantic is slightly different and may cause layout issues in tight containers.

---

## Events Analysis

### Supported

- **`GotFocus`** — fired on `focus` via `on:focus={() => emitEvent(node.name, 'GotFocus')}`. No arguments. Correct.
- **`LostFocus`** — fired on `blur` via `on:blur={() => emitEvent(node.name, 'LostFocus')}`. No arguments. Correct.
- **`TextChanged`** — fired via `emitInteraction(…, { component: node.name, event: 'TextChanged', args: [] })` inside `textInput()`. No arguments. Correct.

### Missing / Incorrect

- **`TextChanged` event has no arguments — correct per AI spec** — App Inventor's `TextChanged` event (defined in `TextBoxBase`) takes no parameters. Tensor passes `args: []`. This is correct; no issue.
- **No missing events** for TextBox/TextBoxBase specifically. All three declared `@SimpleEvent` handlers (`GotFocus`, `LostFocus`, `TextChanged`) are implemented.

---

## Methods Analysis

### Supported

- None of the TextBox/TextBoxBase methods are handled in `simulation_wasm.go`'s `CallMethod`. The switch statement has cases for `Notifier`, `ListPicker`, `Spinner`, `DatePicker`, `TimePicker`, `ListView`, and `TinyDB`, but no case for `TextBox`.

### Missing / Incorrect

- **`HideKeyboard()`** [High] — Falls through to `h.Unsupported("method", componentName+".HideKeyboard")`. This is a common method in App Inventor apps (called to dismiss the soft keyboard after multi-line input). In the browser simulator there is no soft keyboard to hide, but the method call should be a no-op rather than generating an unsupported-method warning that surfaces to the user. Any app using `HideKeyboard` will show a spurious orange "unsupported" outline.
- **`RequestFocus()`** [High] — Falls through to `h.Unsupported(...)`. `RequestFocus` is very commonly used in tutorials and real apps to move the cursor to a text field. It should be implemented as a best-effort `focus()` call via an effect, or at minimum be a silent no-op, rather than triggering the unsupported warning.
- **`MoveCursorTo(position)`** [Medium] — Falls through to unsupported. The method takes one integer argument (1-indexed position). Could be implemented via HTML `setSelectionRange`, or made a no-op.
- **`MoveCursorToStart()`** [Medium] — Falls through to unsupported.
- **`MoveCursorToEnd()`** [Medium] — Falls through to unsupported.

All five methods generate an "unsupported" marker in the simulation result, which causes the TextBox to receive an orange dashed outline (`class:sim-unsupported`) even when the program logic is otherwise correct. This is a usability regression for any app that focuses or hides the keyboard.

---

## Behaviour Gaps

### MultiLine not simulated

App Inventor TextBox supports multi-line input via `MultiLine=true`. When enabled, the underlying `EditText` accepts newlines, scrolls vertically, and the Done key is replaced with a Return key. Tensor always renders `<input type="text">` (single line). No `<textarea>` fallback is provided. Any app designed with a multi-line textbox will render incorrectly: text will not wrap, newlines cannot be entered, and height-auto expansion will not occur.

### ReadOnly vs Enabled distinction lost

In App Inventor `ReadOnly=true` calls `view.setEnabled(!readOnly)` on the underlying view but the intent is different from `Enabled`: `ReadOnly` prevents editing but the field visually appears interactive (not greyed out). Tensor maps `Enabled` to HTML `disabled`, but has no concept of `ReadOnly`. An app that uses `ReadOnly` to prevent editing while keeping the field visually active will instead see no effect (the field remains editable).

### NumbersOnly keyboard restriction vs inputmode

App Inventor's `NumbersOnly=true` sets `InputType.TYPE_CLASS_NUMBER | TYPE_NUMBER_FLAG_SIGNED | TYPE_NUMBER_FLAG_DECIMAL`, which restricts the soft keyboard to numeric input and also allows the minus sign and decimal point. Tensor sets `inputmode="decimal"`, which is the correct HTML equivalent for the browser's virtual keyboard hint. However, `inputmode` does not enforce input validation on desktop browsers (physical keyboards can still type letters). This is an acceptable platform limitation, but apps that depend on strict number-only enforcement will not be faithfully simulated on desktop.

### TextColor and BackgroundColor default semantics

App Inventor uses `COLOR_DEFAULT = 0x00000000` as a sentinel for "use the Android system/theme default". The simulator hardcodes `&HFF222222` as the TextColor default. If a block reads the TextColor and compares it with the system default constant, the comparison will fail. Additionally, if the designer leaves BackgroundColor at its default, Tensor produces no background CSS (since `BackgroundColor` has no entry in `SIMULATION_DEFAULTS` for TextBox), which means the CSS `.sim-textbox` rule's `background: #fff` takes over — this is visually acceptable but semantically wrong.

### FontSize in sp vs px

App Inventor measures font size in `sp` (scale-independent pixels), which respects the user's device font-size preference. The simulator applies `font-size` in `px`, which does not scale with browser text preferences. This is a platform-level difference rather than a bug, but it means accessibility scenarios (large text mode) are not simulated.

### TextAlignment not applied

App Inventor supports left (0), center (1), and right (2) text alignment via `TextAlignment`. Tensor does not include `TextAlignment` in `SIMULATION_VISUAL_PROPS` or apply any `text-align` CSS. Blocks that set alignment will silently have no effect, and all text boxes will render left-aligned.

### HideKeyboard triggers unsupported warning

As noted in Methods, `HideKeyboard()` is commonly called in multi-line textbox patterns. The unsupported warning paints the component with an orange outline, misleading the developer into thinking the component itself is broken rather than just an unimplemented method.

---

## Bugs Found

### Bug 1: `HideKeyboard`, `RequestFocus`, `MoveCursorTo*` all route to `Unsupported` [High]

**Location:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, `CallMethod` function (line 184–244).

There is no `case "TextBox":` block in the switch statement. All method calls on a TextBox fall through to `h.Unsupported("method", componentName+"."+method)` on line 242. This means any app that calls `TextBox1.HideKeyboard()`, `TextBox1.RequestFocus()`, `TextBox1.MoveCursorTo(1)`, `TextBox1.MoveCursorToStart()`, or `TextBox1.MoveCursorToEnd()` will produce an unsupported-method entry, causing the TextBox to render with an orange dashed outline (`class:sim-unsupported` on line 33 of `SimulationComponent.svelte`). The fix is to add a `case "TextBox":` (or `case "TextBox", "PasswordTextBox":`) block that makes these methods no-ops (returning `runtime.VoidVal()`).

### Bug 2: `MultiLine` not in `SIMULATION_VISUAL_PROPS` so it is silently dropped from state [High]

**Location:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, `SIMULATION_VISUAL_PROPS` set (lines 63–100) and `SIMULATION_DEFAULTS` for TextBox (line 42).

`deriveStateFromDesignerProps` (line 152) only copies properties that appear in `SIMULATION_VISUAL_PROPS`. Because `MultiLine` is absent from the set, any designer-set `MultiLine` value is silently dropped. If a block later reads `TextBox1.MultiLine`, it will get `null` (no entry in state) rather than the correct boolean. The fix is to add `'MultiLine'` and `'ReadOnly'` to `SIMULATION_VISUAL_PROPS`, add their defaults to `SIMULATION_DEFAULTS.TextBox`, and switch the `<input>` to a `<textarea>` when `MultiLine === true`.

### Bug 3: `ReadOnly` not in `SIMULATION_VISUAL_PROPS` — field remains editable when it should not be [High]

**Location:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js` (line 42 and `SIMULATION_VISUAL_PROPS`).

Same root cause as Bug 2. `ReadOnly=true` in designer props is silently discarded. The rendered `<input>` has no `readonly` attribute, so the textbox is always editable regardless of the property value.

### Bug 4: `TextChanged` event — text value is not passed to the Go runtime args [Low]

**Location:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, `textInput` function (lines 98–104).

```js
function textInput(e) {
  const value = e.currentTarget.value;
  emitInteraction(
    [{ component: node.name, property: 'Text', value }],
    { component: node.name, event: 'TextChanged', args: [] },
  );
}
```

`TextChanged` is fired with `args: []`. App Inventor's `TextChanged` event also takes no arguments (no parameters in `TextBoxBase.TextChanged()`), so `args: []` is correct per the spec. This is **not** a bug — noted here for completeness.

### Bug 5: `BackgroundColor` absent from `SIMULATION_DEFAULTS` for TextBox [Low]

**Location:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, `SIMULATION_DEFAULTS.TextBox` (line 42).

Other components (Screen, Form, Label) include `BackgroundColor` in their defaults. TextBox does not. If a block reads `TextBox1.BackgroundColor` before any write, the runtime host returns `NullVal()` rather than the App Inventor default sentinel `&H00000000`. This could cause unexpected behaviour in apps that check or compare the background color.

---

## Android/App Inventor Standards Compliance

| Standard | Status |
|---|---|
| `COLOR_DEFAULT` sentinel (`&H00000000`) | Not faithfully simulated — Tensor uses hardcoded `&HFF222222` for TextColor |
| Width/Height encoding (-1 fill, -2 wrap, -1000x percent, positive px) | Correctly implemented in `sizeStyle()` |
| NumbersOnly keyboard hint | Approximated with `inputmode="decimal"`; physically-typed characters not restricted on desktop |
| MultiLine / single-line behaviour | Not implemented — always single-line `<input>` |
| ReadOnly (editable=false but not disabled) | Not implemented |
| TextAlignment (0/1/2) | Not implemented |
| FontBold / FontItalic / FontTypeface | Not implemented |
| HintColor | Not implemented |
| sp font units | Not implemented (px used instead) |
| HideKeyboard no-op on browser | Should be no-op; currently generates unsupported warning |
| RequestFocus → element focus | Should attempt focus; currently generates unsupported warning |
| MoveCursorTo / MoveCursorToStart / MoveCursorToEnd | Should attempt cursor positioning; currently generates unsupported warning |

---

## Summary

The Tensor simulator provides a functional baseline for TextBox: text input, hint display, NumbersOnly keyboard mode, GotFocus/LostFocus/TextChanged events, and standard size/color theming all work correctly. However, there are significant gaps that will cause incorrect behaviour or misleading unsupported warnings for many real-world App Inventor apps.

**Top 3 action items:**

1. **Add a `case "TextBox"` block in `CallMethod` (simulation_wasm.go)** making `HideKeyboard`, `RequestFocus`, and the three `MoveCursor*` methods no-ops (`return runtime.VoidVal()`). This eliminates spurious unsupported-method warnings for the single most common TextBox method pattern in App Inventor tutorials.

2. **Add `MultiLine` and `ReadOnly` to `SIMULATION_VISUAL_PROPS` and `SIMULATION_DEFAULTS`** in `simulation-capabilities.js`, and update `SimulationComponent.svelte` to render a `<textarea>` (with `rows` derived from height) when `MultiLine === true`, and add the `readonly` HTML attribute when `ReadOnly === true`.

3. **Add `TextAlignment`, `FontBold`, and `FontItalic` to `SIMULATION_VISUAL_PROPS`** and apply the corresponding CSS (`text-align`, `font-weight`, `font-style`) in `baseStyle()`. These are visible designer properties used by a large fraction of real apps and currently have zero effect in simulation.
