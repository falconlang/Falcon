# EmailPicker Simulator Implementation Plan

## Overview
EmailPicker is, per its `helpString`, "a kind of text box" in the **Social** category. On a real device it behaves like a single-line `TextBox`: as the user types the name or email address of a contact, the OS shows a dropdown of contact-completion suggestions drawn from the device address book. The selected/typed value is exposed via the `Text` property; when `Text` is empty the `Hint` shows faintly. It is a **visible** component with no special container relationship — it is a standalone leaf that can sit directly on a `Screen`/`Form` or inside any arrangement, exactly like `TextBox`.

Functionally it is identical to `TextBox` (single-line only — no `MultiLine`, `NumbersOnly`, or `ReadOnly` properties) plus the contacts autocomplete. The simulator already fully implements the `TextBox`/`PasswordTextBox` text-input path (rendering, events, methods, cursor effects). EmailPicker can ride almost entirely on that existing path.

## Feasibility Verdict
**Partially feasible.**

The text-box surface — appearance, text entry, `Text`/`Hint`/`HintColor`, colors, fonts, alignment, `Enabled`/`Visible`, the three events, and all four cursor/focus methods — is fully feasible and already implemented for `TextBox`. EmailPicker is, in practice, a `TextBox` clone in the simulator.

The one thing that **cannot** be reproduced is the defining feature: the **contacts/email autocomplete dropdown**. That dropdown is populated from the device's native address book, which requires:
- The Android Contacts content provider / READ_CONTACTS permission — no web equivalent.
- The browser has no Contacts API exposed to page JS (the limited `navigator.contacts.select()` Contact Picker API is Chromium-mobile-only, requires a user gesture + secure context, returns a one-shot picker dialog rather than inline type-ahead suggestions, and exposes no data the user has not explicitly granted). It does not model AI's behavior and is not portable.

Realistic simulated approximation: render EmailPicker as a normal single-line text box (an `<input type="email">`). The user can freely type any value including a full email address; `Text`/`TextChanged`/focus events all fire normally. No autocomplete suggestions are shown. This matches the visible behavior of an EmailPicker whose owner has not (yet) typed something that resolves to a known contact — i.e. the common case — and is honest: there are no contacts in a browser sandbox to complete against. The plan does **not** attempt to synthesize a fake contact list, since that would be misleading and is not part of the spec.

## Properties
Defaults from the spec's `designerProperties`. Note `&H00000000` for `BackgroundColor`/`TextColor` is AI's "Default" sentinel (fully transparent alpha) meaning "use the platform default 3-D look / black text", not literally transparent — mirror the `TextBox` treatment (opaque white background, black text).

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| Text | `""` | Behavioral/Visual | `value` bound to `<input>`; two-way via `textInput` -> `Text` patch | High |
| Hint | `""` | Visual | `placeholder` attribute | High |
| HintColor | `&HFF888888` | Visual | `--sim-placeholder-color` via existing `hintColorStyle()` | High |
| Enabled | `True` | Behavioral | `disabled={!enabled}` (reuse `enabled` derive) | High |
| Visible | `True` | Visual | existing visibility handling | High |
| BackgroundColor | `&H00000000` (Default) | Visual | `colorValue()` in `baseStyle()`; default block uses `&HFFFFFFFF` like TextBox | High |
| TextColor | `&H00000000` (Default=black) | Visual | `baseStyle()` color; default block uses `&HFF000000` | High |
| TextAlignment | `0` (left) | Visual | `textAlignStyle()` in `baseStyle()` | Medium |
| FontSize | `14.0` | Visual | `baseStyle()` font-size | Medium |
| FontBold | `False` | Visual | `baseStyle()` font-weight | Medium |
| FontItalic | `False` | Visual | `baseStyle()` font-style | Medium |
| FontTypeface | `0` | Visual | `baseStyle()` `typefaceStyle()` | Low |
| Width / Height | `""` (auto) | Visual | `sizeStyle()`; seed `Width: 160` default like TextBox | Medium |
| Left / Top | `""` | Visual | only meaningful inside `AbsoluteArrangement` (`positionStyle()`) | Low |

No property is un-honorable in the browser except the implicit contacts source (not a designer property — it has no UI knob). `Column`/`Row` block-only getters are layout-internal and not simulated (consistent with all other components).

## Events
| Event | Args | How/when the simulator fires it |
|---|---|---|
| TextChanged | none | On `<input>` `input` — `emitInteraction([{property:'Text'}], {event:'TextChanged'})`, identical to TextBox. Also fired by the Go host on programmatic `Text` writes (the existing `componentType == "TextBox" \|\| "PasswordTextBox"` branch must be extended to include EmailPicker). |
| GotFocus | none | `<input>` `focus` -> `focusEvent()` -> `emitEvent('GotFocus')` (reused). |
| LostFocus | none | `<input>` `blur` -> `blurEvent()` -> `emitEvent('LostFocus')` (reused). |

All three events are fully fireable in the browser. No event depends on the contacts dropdown (AI raises no event for "suggestion chosen" — choosing a suggestion just sets `Text`, which fires `TextChanged`, which we already cover).

## Methods
| Method | Signature | Simulated behavior |
|---|---|---|
| RequestFocus | `RequestFocus()` | Emits a `focus` component-action effect; overlay token -> `focusCurrentInput()` focuses the `<input>`. Reused verbatim from `callTextBoxMethod`. |
| MoveCursorToStart | `MoveCursorToStart()` | `cursor-start` effect -> `setTextCursor(0)`. Reused. |
| MoveCursorToEnd | `MoveCursorToEnd()` | `cursor-end` effect -> cursor to end. Reused. |
| MoveCursorTo | `MoveCursorTo(position: number)` | `cursor-position` effect with 1-indexed `position`; `setTextCursor(position-1)` clamps. Reused. |

All four methods already exist in `callTextBoxMethod`; routing EmailPicker there is the only change. (`HideKeyboard` is handled as a no-op by the same function even though it is not in EmailPicker's method list — harmless.)

## Implementation Plan
The guiding principle: **do not write a new branch.** EmailPicker is a single-line TextBox, so add `'EmailPicker'` to the existing TextBox/PasswordTextBox type checks rather than duplicating logic.

- **simulation-capabilities.js**
  - Add `'EmailPicker'` to `SIMULATION_SUPPORTED_TYPES`. (Not non-visible — leave `SIMULATION_NONVISIBLE_TYPES` alone.)
  - Add a defaults block inside `buildSimulationDefaults()`, modeled on `TextBox` minus the props EmailPicker lacks:
    ```js
    EmailPicker: {
      ...COMMON_VISIBLE_PROPS,
      ...COMMON_TEXT_PROPS,
      Text: '',
      Hint: '',
      BackgroundColor: '&HFFFFFFFF',
      HintColor: '&HFF888888',
      Width: 160,
    },
    ```
    (`COMMON_TEXT_PROPS` already supplies `FontSize/FontBold/FontItalic/FontTypeface/TextColor/TextAlignment`; `COMMON_VISIBLE_PROPS` supplies `Visible/Enabled/Width/Height`. `HintColor` default `&HFF888888` matches the spec.)
  - `SIMULATION_VISUAL_PROPS`: no additions — every EmailPicker prop (`Text`, `Hint`, `HintColor`, `Visible`, `Enabled`, `Width`, `Height`, `Left`, `Top`, `BackgroundColor`, `TextColor`, `FontSize`, `FontBold`, `FontItalic`, `FontTypeface`, `TextAlignment`) is already in the set.
  - `isBooleanProp` / `isNumericProp` / `coerceSimulationValue`: no additions — all relevant props (`Enabled`, `Visible`, `FontBold`, `FontItalic`; `FontSize`, `Width`, `Height`, `Left`, `Top`, `TextAlignment`) are already registered.
  - `deriveStateFromDesignerProps`: no EmailPicker-specific derived state needed.

- **SimulationComponent.svelte**
  - Extend the existing branch guard from `{:else if node.type === 'TextBox' || node.type === 'PasswordTextBox'}` to also include `|| node.type === 'EmailPicker'`.
  - Since EmailPicker has no `MultiLine`, `isMultilineTextBox` stays false (it already keys on `node?.type === 'TextBox'`), so EmailPicker always takes the single-line `<input>` path — correct.
  - Optionally set the input type to `email` for EmailPicker (better mobile keyboard / built-in `@` affordance) by extending the existing `type={...}` ternary:
    ```svelte
    type={node.type === 'PasswordTextBox' && !boolValue(props.PasswordVisible, false)
      ? 'password'
      : node.type === 'EmailPicker' ? 'email' : 'text'}
    ```
    `inputmode` can stay as-is (`text`); `NumbersOnly`/`ReadOnly` props are simply absent for EmailPicker so `boolValue(..., false)` yields the right defaults. Wiring (`textInput`, `focusEvent`, `blurEvent`, `bind:this={textInputEl}`, `baseStyle()`, `hintColorStyle()`) is reused unchanged.
  - CSS: reuse the existing `.sim-textbox` class — no new styles.
  - Reuse pattern: **TextBox**.

- **SimulateOverlay.svelte**
  - None. No dialogs, toasts, or new effects. The `focus`/`cursor-*` component-action effects EmailPicker uses are already processed by the generic `component-action` handler in the overlay.

- **simulation_wasm.go**
  - `SetProperty`: extend the programmatic-`Text` -> `TextChanged` branch guard from `(componentType == "TextBox" || componentType == "PasswordTextBox")` to also include `componentType == "EmailPicker"`, so a blocks-driven `set Text` fires `TextChanged`.
  - `CallMethod`: add `"EmailPicker"` to the `case "TextBox", "PasswordTextBox":` line so it routes to `callTextBoxMethod` (covers `RequestFocus`, `MoveCursorTo`, `MoveCursorToStart`, `MoveCursorToEnd`).
  - `GetProperty`: generic store path suffices — EmailPicker has no computed getters (the `Column`/`Row` getters are not simulated).
  - No `callEmailPickerMethod` needed; no `h.Unsupported` call (the contacts dropdown has no method/property surface to mark unsupported — it is purely an implicit native behavior). Rebuild with `npm run build:wasm`.

- **design-schema-tree.js**
  - Containment already handled: EmailPicker is a leaf, not a container, so `canContainDesignComponent()` needs no entry. Once `'EmailPicker'` is in `SIMULATION_SUPPORTED_TYPES`, `unsupportedSimulationComponents()` stops reporting it and the dashed placeholder disappears; `designTreeToInitialState()` merges the new `SIMULATION_DEFAULTS.EmailPicker` automatically.

## Dependencies & Ordering
No external libraries. No prerequisite components — it rides on the already-shipped TextBox infrastructure (render branch, events, methods, cursor/focus effects). Can be implemented standalone in isolation.

## Web-Platform Limitations & Fidelity Caveats
- **No contacts autocomplete.** The defining feature — the type-ahead dropdown of contact names/emails from the device address book — does not appear. The browser sandbox has no address book and no portable Contacts API; we deliberately do not fake a contact list. This is the only behavioral divergence.
- **No "computing matches" latency.** The spec notes the dropdown "can take several seconds to appear … intermediate results"; with no dropdown there is nothing to delay. Not simulated.
- **`type="email"` browser hints differ from Android.** Desktop browsers may offer their own saved-email autofill on an `<input type="email">`; this is browser-controlled, not contacts-driven, and not equivalent to AI's behavior. It is cosmetic and harmless, but worth noting it is not the AI contacts feature.
- **Default-color sentinel.** AI `&H00000000` ("Default") is rendered as opaque white background / black text to match how AI actually paints a text box, not as literal transparency — same convention already used for TextBox.

## Effort Estimate
**S.** Five tiny edits, all extensions of existing TextBox handling: one `SIMULATION_SUPPORTED_TYPES` entry + one defaults block in `simulation-capabilities.js`, two `||`/case additions in `simulation_wasm.go`, and one branch-guard (plus optional `type="email"`) in `SimulationComponent.svelte`. No new render branch, no new effects, no new CSS, no overlay or schema-tree changes. The only non-mechanical work is acknowledging — in code review and any user-facing note — that the contacts dropdown is intentionally absent.
