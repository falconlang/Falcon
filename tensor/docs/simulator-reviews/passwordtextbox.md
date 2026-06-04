# PasswordTextBox Simulator Review

## Overview

`PasswordTextBox` in App Inventor inherits from `TextBoxBase`, which itself extends `AndroidViewComponent`. It is a single-line text input that masks entered characters using `PasswordTransformationMethod`. It adds two own properties (`PasswordVisible`, `NumbersOnly`) on top of the full `TextBoxBase` property/event/method set. The Tensor simulator renders it as an HTML `<input>` element and shares the rendering branch with `TextBox`. This review compares every property, event, and method defined in the Java source against what the simulator supports.

---

## Properties Analysis

### Supported

- `Text` — read/write, correctly initialised to `''`, reactively bound.
- `Hint` — read/write, correctly initialised to `''`, mapped to `placeholder`.
- `Enabled` — read/write, correctly toggles `disabled` on the `<input>`.
- `Visible` — read/write, hides the entire element when false.
- `NumbersOnly` — read/write, default `false` (matches AI default). Applied as `inputmode="decimal"` on the `<input>`.
- `PasswordVisible` — read/write, default `false` (matches AI default). Toggled to `type="text"` / `type="password"` on the `<input>` (line 264 of SimulationComponent.svelte).
- `Width` — read/write, default `160` px, size constants `-1`/`-2`/`<=-1000` all handled.
- `Height` — read/write, default `-1` (automatic), same size-constant logic.
- `TextColor` — read/write, default `&HFF222222`, parsed via `colorValue()`.
- `FontSize` — read/write, default `14` px, coerced to number.
- `GotFocus` / `LostFocus` events correctly wired to `on:focus` / `on:blur`.

### Missing / Unsupported

| Property | AI Default | Priority | Notes |
|---|---|---|---|
| `BackgroundColor` | `COLOR_DEFAULT` (system drawable) | **High** | Defined in `SIMULATION_VISUAL_PROPS` and parsed via `baseStyle()`, but **not** included in `SIMULATION_DEFAULTS` for `PasswordTextBox`. The `TextBox` entry also omits it, so if a designer sets it the update path works, but the initial state has no `BackgroundColor` key and `baseStyle()` silently skips it. |
| `TextAlignment` | `0` (normal / left) | **Medium** | Not in defaults, not in `SIMULATION_VISUAL_PROPS`, no CSS mapping. Changing it in Blocks (`set PasswordTextBox1.TextAlignment to 1`) has no visual effect. |
| `FontBold` | `false` | **Medium** | Not tracked in simulation state, no CSS mapping (`font-weight`). |
| `FontItalic` | `false` | **Medium** | Not tracked in simulation state, no CSS mapping (`font-style`). |
| `FontTypeface` | `0` (default) | **Low** | Not tracked; designer fonts would be ignored. |
| `HintColor` | `COLOR_DEFAULT` (gray) | **Low** | Not tracked; the hint always uses the browser placeholder color. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `Height` | `-1` (automatic / wrap content in AI sense) | `-1` | OK (matches) | — |
| `Width` | `TEXTBOX_PREFERRED_WIDTH` (≈160 dp) | `160` | Low | Close enough for simulation purposes. |
| `TextColor` | `COLOR_DEFAULT` (system, resolves to black on light theme) | `&HFF222222` | Low | Near-black, visually close but not the same token. |
| `BackgroundColor` | `COLOR_DEFAULT` (3-D system drawable) | absent / transparent | **High** | Tensor shows a plain white input; AI shows the Android system EditText background (bevelled/shadowed). More importantly the property is missing from `SIMULATION_DEFAULTS`, so a block `set PasswordTextBox1.BackgroundColor to Default` will leave the state key undefined rather than resetting it. |

---

## Events Analysis

### Supported

- `GotFocus` — fired on `focus` event (SimulationComponent.svelte line 271). No arguments. Correct.
- `LostFocus` — fired on `blur` event (SimulationComponent.svelte line 272). No arguments. Correct.
- `TextChanged` — fired via `emitInteraction` inside `textInput()` (line 103). No arguments. Correct — AI's `TextChanged` event also takes no arguments.

### Missing / Incorrect

No events are missing from the `PasswordTextBox`-specific or `TextBoxBase`-defined set. All three public events are handled.

One nuance: in App Inventor `TextChanged` fires only when the text actually changes (`!lastText.equals(...)` guard in `TextBoxBase.onTextChanged`). The Tensor simulator fires it on every `input` DOM event, including cases where the browser may fire `input` with the same value (e.g. composition events on some IMEs). This is a **Low** priority behavioural gap, not a concrete bug.

---

## Methods Analysis

### Supported

None of the `TextBoxBase` methods are explicitly handled in `simulation_wasm.go`'s `CallMethod` switch. However there is a fallthrough to `h.Unsupported(...)`, so calling any method from Blocks will log an unsupported warning rather than silently crashing.

### Missing / Incorrect

| Method | Signature | Priority | Notes |
|---|---|---|---|
| `RequestFocus()` | `void` | **High** | Commonly used in real apps to move focus programmatically. Calling it marks the component as unsupported. No DOM `focus()` call is made via the effects system. |
| `MoveCursorTo(position int)` | `void` | **Medium** | Not implemented. Less commonly used but present in AI. |
| `MoveCursorToEnd()` | `void` | **Medium** | Not implemented. |
| `MoveCursorToStart()` | `void` | **Medium** | Not implemented. |

For `PasswordTextBox` specifically, there are no additional own methods beyond what `TextBoxBase` provides.

---

## Behaviour Gaps

### 1. `NumbersOnly` uses `inputmode` instead of restricting `type`

**Priority: High**

In App Inventor, `NumbersOnly=true` on a `PasswordTextBox` changes the Android `InputType` to `TYPE_CLASS_NUMBER | TYPE_NUMBER_VARIATION_PASSWORD` (or `_NORMAL` if also `PasswordVisible`). This means the soft keyboard shows a numeric pad. In the simulator, `inputmode="decimal"` is applied (line 268 of SimulationComponent.svelte), which is a hint to the browser but does not prevent alphabetic input on desktop. The field remains `type="password"` or `type="text"`, so a user can still type letters. This diverges from the real behaviour where the keyboard is hard-locked to numbers.

Additionally, the `inputmode` is applied to **both** `TextBox` and `PasswordTextBox` identically (the condition reads `props.NumbersOnly ? 'decimal' : 'text'`). For `PasswordTextBox` with `NumbersOnly=true` and `PasswordVisible=false`, AI uses `TYPE_NUMBER_VARIATION_PASSWORD`, which shows a numeric keypad while still masking digits. The browser has no equivalent for a masked numeric field, which is a known platform gap, but the current implementation does not even attempt to use `type="number"` when `PasswordVisible=true && NumbersOnly=true`.

### 2. `PasswordVisible` toggle does not reset transformation method after `NumbersOnly` is also set

**Priority: Medium**

In AI, the combination of `PasswordVisible` and `NumbersOnly` produces four distinct Android `InputType` combinations. In Tensor, the only interaction is `type="password"` vs `type="text"` — `NumbersOnly` does not influence `type` at all. Specifically, `PasswordVisible=true, NumbersOnly=true` should present a visible-text numeric field; currently it shows `type="text"` with `inputmode="decimal"`, which is partially correct but omits the constraint.

### 3. `BackgroundColor` property missing from `SIMULATION_DEFAULTS`

**Priority: High**

`simulation-capabilities.js` line 43 defines `PasswordTextBox` defaults without `BackgroundColor`. When the interpreter calls `h.SetProperty(name, type, "BackgroundColor", value)`, it writes to the runtime state and the patch, so dynamic Blocks changes are forwarded to the frontend. However, on initial load `deriveStateFromDesignerProps` only copies keys that appear in `SIMULATION_VISUAL_PROPS`, and `BackgroundColor` is in that set — so designer-set values _do_ flow through. The gap is specifically that the fallback/initial state has no `BackgroundColor` entry, meaning `baseStyle()` omits the `background:` rule entirely, rendering a transparent `<input>` over whatever parent background exists instead of the white/default system background.

### 4. `TextChanged` event argument count

**Priority: Low**

App Inventor's `TextChanged` takes zero arguments. The Tensor implementation fires `emitEvent` with `args: []` which is correct. No issue here — this item is confirmed correct.

### 5. Font properties are visual-only in the simulator

**Priority: Medium**

`FontBold`, `FontItalic`, and `FontTypeface` are missing from `SIMULATION_VISUAL_PROPS` and from `SIMULATION_DEFAULTS`. A block that sets `set PasswordTextBox1.FontBold to true` will call `h.setProperty` in the Go host and update the runtime state, but because `FontBold` is not in `SIMULATION_VISUAL_PROPS`, `deriveStateFromDesignerProps` will never push it to the frontend, and the frontend `baseStyle()` has no mapping for it. The text will always render in the browser default weight/style regardless of what Blocks sets.

### 6. `RequestFocus` effect not implemented

**Priority: High**

Real apps frequently call `RequestFocus()` to move the cursor into the password field (e.g. after showing an error). In Tensor, `CallMethod` falls through to `h.Unsupported("method", "PasswordTextBox1.RequestFocus")`, which adds an orange dashed outline to the component in the preview and logs a warning, but does not attempt to focus the DOM element. An `effects` entry of type `"component-action"` with `action: "focus"` could be added and handled in the frontend.

---

## Bugs Found

### Bug 1: `inputmode` attribute applied to both `TextBox` and `PasswordTextBox` without `PasswordVisible` interaction

**File:** `src/lib/SimulationComponent.svelte`, line 268  
**Severity: High**

```svelte
inputmode={props.NumbersOnly ? 'decimal' : 'text'}
```

When `PasswordVisible=false` and `NumbersOnly=true`, the AI platform shows a numeric keypad but the digits are masked. In the simulator the `type` is already `"password"`, and `inputmode="decimal"` on a `type="password"` field is ignored by most browsers (browser treats password as text keyboard). The simulator should use `type="number"` (dropping masking, since there is no `type="number password"`) and note the gap, or add a comment explaining the limitation. As-is, `NumbersOnly` has zero visible effect for a masked `PasswordTextBox` in desktop browsers.

### Bug 2: `BackgroundColor` absent from `SIMULATION_DEFAULTS` for `PasswordTextBox`

**File:** `src/lib/simulation-capabilities.js`, line 43  
**Severity: High**

```js
PasswordTextBox: { ...COMMON_VISIBLE_PROPS, Text: '', Hint: '', NumbersOnly: false, PasswordVisible: false, Width: 160, TextColor: '&HFF222222', FontSize: 14 },
```

`BackgroundColor` is present for `Screen` and `Form` but absent for `PasswordTextBox` (and `TextBox`). The Android component defaults to `COLOR_DEFAULT`, which resolves to the system EditText 3-D background. In the simulator, the input will have a transparent background on first render, inheriting whatever parent color is set. This can produce invisible text inputs if the parent has a white background (white-on-white).

### Bug 3: `TextChanged` event never fires when `Text` is set programmatically

**File:** `src/lib/SimulationComponent.svelte`, `src/lib/simulation_wasm.go`  
**Severity: Medium**

In App Inventor, `TextChanged` fires both when the user types and when the program calls `set PasswordTextBox1.Text to "..."`. This is explicit in the Java docs: "Event raised when the text … is changed, either by the user **or the program**." In Tensor, `textInput()` (line 98) fires only on the DOM `input` event. When `h.SetProperty` is called from Blocks to set `Text`, no `TextChanged` event is dispatched. This diverges from the AI specification.

### Bug 4: `MoveCursorTo` / `MoveCursorToEnd` / `MoveCursorToStart` silently log unsupported

**File:** `src/lib/simulation_wasm.go`, `CallMethod`  
**Severity: Medium**

These three `@SimpleFunction` methods are annotated and user-visible in AI. When called from Blocks inside the simulator, they fall through to `h.Unsupported(...)` and mark the component with the orange dashed outline. No cursor manipulation is attempted. While a full cursor API is hard to expose through the state-patch model, at minimum `MoveCursorToEnd` and `MoveCursorToStart` could be silently no-op'd (return `VoidVal()` without calling `Unsupported`) since the visual impact is minor in a simulation context.

---

## Android/App Inventor Standards Compliance

| Concern | Status |
|---|---|
| Single-line constraint | Correct — HTML `<input>` is inherently single-line. |
| IME `ACTION_DONE` key | Not simulated. Pressing Enter on desktop does not fire a `Done` key event. Low priority (no public event exposed for this in AI). |
| Password masking default | Correct — defaults to `type="password"`. |
| `PasswordVisible` toggle | Correct — `type` attribute switches reactively. |
| `NumbersOnly` numeric keyboard lock | Partial — `inputmode` hint only; no actual enforcement of numeric input. |
| `COLOR_DEFAULT` resolution | Not simulated — AI resolves to the system EditText background drawable; Tensor uses transparent. |
| High-contrast mode | Not applicable in web simulator. |
| Large text / `BigDefaultText` | Not simulated (`FontSize` is fixed at designer value with no scale). |
| `TextChanged` on programmatic set | Missing — see Bug 3. |

---

## Summary

The `PasswordTextBox` simulation is functional for basic use-cases: the component renders, `PasswordVisible` toggles masking correctly, focus events fire, and text input synchronises `Text` state and fires `TextChanged` for user interactions. The implementation is a reasonable first pass.

**Top 3 Action Items:**

1. **[Critical] Fix `TextChanged` not firing on programmatic `Text` set (Bug 3).** App Inventor explicitly documents "either by the user or the program". Many real apps set `Text` from Blocks and expect `TextChanged` to trigger dependent logic. The Go `SetProperty` handler should call `runEvent(componentName, componentType, "TextChanged", nil)` after updating the `Text` property.

2. **[High] Add `BackgroundColor` to `SIMULATION_DEFAULTS` for `PasswordTextBox` (and `TextBox`) (Bug 2).** Without a default background, the input may be invisible against parent backgrounds. Suggested default: `'&HFFFFFFFF'` (white), which approximates the system EditText background without requiring the full Android 3-D drawable.

3. **[High] Implement `RequestFocus()` in the Go `CallMethod` handler.** This is one of the most commonly called methods on text inputs. Add a `"TextBox"/"PasswordTextBox"` case that emits a `component-action` effect with `action: "focus"`, and handle it in `SimulateOverlay.svelte` / `SimulationComponent.svelte` by calling `.focus()` on the bound DOM element.
