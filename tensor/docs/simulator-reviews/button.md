# Button Simulator Review

## Overview

The Button component is a clickable widget inherited from `ButtonBase` (which itself extends `TouchComponent`). It is one of the most fundamental UI components in App Inventor. The Tensor IDE simulator renders it as a plain HTML `<button>` element in `SimulationComponent.svelte`, with property defaults in `simulation-capabilities.js` and event dispatch routed through `simulation_wasm.go`. The simulation correctly handles the core click interaction but is missing a significant portion of the visual/style properties and all focus/touch events defined in the AI specification.

---

## Properties Analysis

### Supported
- `Text` — rendered as button label (fallback `'Button'` shown in template)
- `TextColor` — applied via `baseStyle()` → `colorValue()`
- `FontSize` — applied via `baseStyle()`
- `BackgroundColor` — applied via `baseStyle()` → `colorValue()`
- `Width` / `Height` — size constants `-1` (auto), `-2` (fill), and percentage encoding `≤ -1000` all handled in `sizeStyle()`
- `Enabled` — maps to HTML `disabled` attribute; checked before firing `Click`
- `Visible` — controls whether the element is rendered at all

### Missing / Unsupported

| Property | Priority | Notes |
|---|---|---|
| `Image` | **High** | `TouchComponent.Image()` — button background image. In AI it takes precedence over `BackgroundColor`. No rendering, no asset resolution for Button. (`assetUrl` resolves only `props.Picture \|\| props.Image`, but Button defaults object has neither key and the template ignores them.) |
| `Shape` | **High** | `ButtonBase.Shape()` — controls default/rounded/rectangular/oval shapes (values 0–3). Entirely missing from defaults and rendering. |
| `FontBold` | **Medium** | `ButtonBase.FontBold()` — bold font style. Not in defaults, not applied in CSS. |
| `FontItalic` | **Medium** | `ButtonBase.FontItalic()` — italic font style. Not in defaults, not applied in CSS. |
| `FontTypeface` | **Medium** | `ButtonBase.FontTypeface()` — default/serif/sans-serif/monospace or custom .ttf. Not applied. |
| `TextAlignment` | **Medium** | `ButtonBase.TextAlignment()` — center (1) is the AI default; left (0), right (2) also possible. Not rendered. |
| `ShowFeedback` | **Low** | `TouchComponent.ShowFeedback()` — whether visual feedback is shown on press. No equivalent. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority |
|---|---|---|---|
| `Text` | `""` (empty string) | `'Button'` (fallback in template `{props.Text \|\| 'Button'}`) | **High** — The AI default for `Text` is an empty string (see `@DesignerProperty defaultValue = ""`). Tensor uses an empty string in `SIMULATION_DEFAULTS` correctly, but the Svelte template falls back to `'Button'` when `Text` is falsy (including an empty string). This means a freshly-placed button with `Text=""` incorrectly displays the word "Button" in the simulator. |
| `TextColor` | `COLOR_DEFAULT` (`&H00000000`) | `&HFFFFFFFF` (white) | **High** — AI default is `Component.DEFAULT_VALUE_COLOR_DEFAULT` = `&H00000000` (the "use system default" sentinel, which on Android renders as a dark system color). Tensor hardcodes white (`&HFFFFFFFF`), which is correct aesthetically for the dark default button but misrepresents the actual default sentinel value. Programmatic reads of `TextColor` will return the wrong value if not overridden. |
| `BackgroundColor` | `COLOR_DEFAULT` (`&H00000000`) | Not set in defaults (absent from `SIMULATION_DEFAULTS.Button`) | **Medium** — AI default is the "system default" sentinel. Tensor has no `BackgroundColor` key in `Button` defaults, so `baseStyle()` emits no background rule, leaving the CSS class default (`#5b6470`). This is mostly fine visually, but the property is unreadable via blocks until set. |

---

## Events Analysis

### Supported
- `Click` — fired via `buttonClick()` → `emitEvent(node.name, 'Click')` when the button is enabled and visible. Correctly passes no arguments (AI: `EventDispatcher.dispatchEvent(this, "Click")` with no args).

### Missing / Incorrect

| Event | Priority | Notes |
|---|---|---|
| `LongClick` | **High** | `Button.LongClick()` / `Button.longClick()` — triggered on long press. The HTML button has no `on:longpress` or `on:pointerdown`+timer pattern. Long-click blocks in user code will never fire. |
| `GotFocus` | **Medium** | `ButtonBase.GotFocus()` — fires when focus moves onto the button. Present in TextBox but not wired up for Button. The button element has no `on:focus` handler. |
| `LostFocus` | **Medium** | `ButtonBase.LostFocus()` — fires when focus leaves the button. Same gap as GotFocus. |
| `TouchDown` | **Medium** | `TouchComponent.TouchDown()` — fires when the button is first pressed. Slider has `on:pointerdown` but Button does not. |
| `TouchUp` | **Medium** | `TouchComponent.TouchUp()` — fires when the button is released. Same gap as TouchDown. |

---

## Methods Analysis

### Supported
- No explicit methods are defined on Button/ButtonBase in the AI source (beyond inherited VisibleComponent/AndroidViewComponent utility methods). The runtime's `CallMethod` dispatcher correctly falls through to `Unsupported` for any unrecognised call, which is the correct behaviour here.

### Missing / Incorrect
- None applicable — Button defines no `@SimpleFunction` methods. The `Initialize()` override in ButtonBase is an internal lifecycle hook, not a blocks-callable method.

---

## Behaviour Gaps

### 1. Empty Text renders as "Button" (Critical UX gap)
In `SimulationComponent.svelte` line 276:
```svelte
{props.Text || 'Button'}
```
When `Text` is an empty string (the AI default), JavaScript's `||` short-circuits to `'Button'`. AI renders an empty label. This makes every new button falsely display text in the simulator.

**Fix:** Use `props.Text ?? ''` or check `props.Text !== undefined ? props.Text : ''`.

### 2. `Image` property not rendered for Button
`TouchComponent` defines an `Image` property that sets a background image on the button. In the simulator, `assetUrl` is computed from `props.Picture || props.Image`, but the Button branch at line 275–277 only renders text — there is no `<img>` inside the `<button>`. Users who set an image on a Button will see only the text fallback.

### 3. `Shape` property completely absent
`ButtonBase.Shape()` accepts `0` (default), `1` (rounded), `2` (rectangular), `3` (oval). It's a designer-visible property. It is not tracked in `SIMULATION_DEFAULTS.Button`, not in `SIMULATION_VISUAL_PROPS`, not in `coerceSimulationValue`, and not applied in the template. A shaped button will look identical to a default one.

### 4. `LongClick` cannot be triggered
The simulation has no mechanism for a long-press gesture. On Android, `LongClick` is consumed before `Click` if a handler is registered. In Tensor, pressing and holding a button for any duration just fires `Click` on release. Long-click event handlers in user code are silently unreachable.

### 5. Color sentinel `&H00000000` (COLOR_DEFAULT) treated as white
`colorValue()` at line 42–47 of `SimulationComponent.svelte` strips the alpha channel from AI color strings:
```js
const ai = text.match(/^&H([0-9A-Fa-f]{2})([0-9A-Fa-f]{6})$/);
if (ai) return `#${ai[2]}`;
```
For `&H00000000` (COLOR_DEFAULT), this yields `#000000` (black). For `TextColor` this means the text would turn solid black, when on Android the default renders as a dark-grey system color that changes in high-contrast mode. The simulation does not special-case the sentinel value.

### 6. Font styling properties silently ignored
`FontBold`, `FontItalic`, and `FontTypeface` are all part of the AI block palette for Button. If a user writes a block like `set Button1.FontBold to true`, the runtime's `SetProperty` will store the value in state but `baseStyle()` never reads these properties, so the visual result is always plain-weight, default-font text.

### 7. Focus and Touch events not wired
`GotFocus`, `LostFocus`, `TouchDown`, and `TouchUp` are all standard events for Button on App Inventor. The TextBox component correctly wires `on:focus` / `on:blur` for `GotFocus` / `LostFocus`. The Slider correctly wires `on:pointerdown` / `on:pointerup` for `TouchDown` / `TouchUp`. The Button template has none of these handlers, creating an inconsistency even within the simulator's own implementation.

### 8. `ShowFeedback` not simulated
`TouchComponent.ShowFeedback()` controls whether the button visually "highlights" on press. The simulator always shows the `:active` transform animation regardless. This is a low-impact cosmetic gap.

---

## Bugs Found

### Bug 1 — Empty string Text always displays "Button" (High)
**Location:** `src/lib/SimulationComponent.svelte`, line 277
```svelte
{props.Text || 'Button'}
```
`props.Text` is `''` (the AI default) → falsy → fallback `'Button'` is shown.
This contradicts AI behavior where a Button with `Text=""` displays nothing.

### Bug 2 — `TextColor` default value is wrong (High)
**Location:** `src/lib/simulation-capabilities.js`, line 44
```js
Button: { ...COMMON_VISIBLE_PROPS, Text: 'Button', TextColor: '&HFFFFFFFF', FontSize: 14 },
```
The AI default for `TextColor` is `&H00000000` (COLOR_DEFAULT sentinel), not white. White was chosen for aesthetics against the dark button, but this means a block reading `Button1.TextColor` on a button whose color was never changed will return white instead of the system default sentinel, breaking comparisons and conditional logic in user code.

### Bug 3 — `colorValue()` maps `&H00000000` to `#000000` (Medium)
**Location:** `src/lib/SimulationComponent.svelte`, `colorValue()` function, lines 42–47
The function strips the alpha byte and uses only the RGB portion. The AI color `&H00000000` (COLOR_DEFAULT, fully transparent black) becomes `#000000` (solid black in CSS). Any component with `TextColor = COLOR_DEFAULT` that the simulator does process will render black text instead of the system-default color. The correct behavior would be to recognize `&H00000000` as a sentinel and use the CSS inherited/default color.

### Bug 4 — `buttonClick()` does not check `visible` before dispatching (Low)
**Location:** `src/lib/SimulationComponent.svelte`, line 122–125
```js
function buttonClick() {
  if (!enabled || !visible) return;
  emitEvent(node.name, 'Click');
}
```
The guard checks `visible` but the outer `{#if visible && !nonVisible}` block already prevents the entire button from rendering when not visible. The `visible` check inside `buttonClick` is therefore redundant, not harmful. However, if `enabled` is `false`, the HTML `disabled` attribute already prevents `on:click` from firing at all — making the `!enabled` check also unreachable. These double-guards are harmless but indicate the logic has not been unified.

---

## Android/App Inventor Standards Compliance

| Standard | Status |
|---|---|
| Click event with no arguments | Correct |
| LongClick returning boolean (consumed) | Not implemented |
| Default text is empty string | Wrong (defaults to `'Button'` display) |
| Default TextColor is COLOR_DEFAULT sentinel | Wrong (hardcoded white) |
| Default BackgroundColor is COLOR_DEFAULT sentinel | Missing from defaults entirely |
| Image overrides BackgroundColor | Not implemented |
| Shape property changes visual appearance | Not implemented |
| TouchDown/TouchUp events on press/release | Not implemented |
| GotFocus/LostFocus on keyboard/accessibility focus | Not implemented |
| FontBold, FontItalic, FontTypeface applied | Not implemented |
| TextAlignment applied | Not implemented |
| ShowFeedback controls press highlight | Not implemented |
| AI size constants (-1=auto, -2=fill, ≤-1000=percentage) | Correctly implemented |
| Color `&HAARRGGBB` format parsed | Correct (but alpha channel silently discarded) |

---

## Summary

The Tensor IDE Button simulator correctly implements the core click interaction, size constants, background/text color rendering, and enabled/visible toggling. However, it is significantly incomplete relative to the App Inventor specification.

**Total issues found: 19** (counting each row above plus the bugs)

### Top 3 Action Items

1. **Fix the empty-text display bug** (Bug 1, Critical UX): Change `{props.Text || 'Button'}` to `{props.Text != null ? props.Text : ''}` so a Button with `Text=""` correctly shows an empty label, matching AI behavior.

2. **Implement `LongClick`, `GotFocus`, `LostFocus`, `TouchDown`, `TouchUp` events** (High): Wire `on:pointerdown` / `on:pointerup` for TouchDown/TouchUp, `on:focus` / `on:blur` for GotFocus/LostFocus, and a pointer-hold timer (≈500 ms) for LongClick. These events are commonly used in AI apps and are completely unreachable today.

3. **Add `Image` and `Shape` property support** (High): Render a button background image when `props.Image` is set (similar to the Image component's asset resolution), and apply CSS `border-radius` or `clip-path` rules for the four Shape values. Add both properties to `SIMULATION_DEFAULTS.Button` and `SIMULATION_VISUAL_PROPS`.
