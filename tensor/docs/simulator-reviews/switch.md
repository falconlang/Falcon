# Switch Simulator Review

## Overview

`Switch` in App Inventor extends `ToggleBase<CompoundButton>`, which itself extends `AndroidViewComponent`. It renders an `androidx.appcompat.widget.SwitchCompat` toggle widget. On top of the shared `ToggleBase` properties (text, font, background/text color, enabled state) the Switch adds four colour properties controlling the thumb and track in their active/inactive states, plus the `On` boolean property.

The Tensor IDE simulation handles Switch rendering inside the same `{:else if node.type === 'CheckBox' || node.type === 'Switch'}` branch in `SimulationComponent.svelte`, with defaults/coercion in `simulation-capabilities.js` and shared backend state management in `simulation_wasm.go`. There are no Switch-specific method handlers because Switch exposes no callable methods.

---

## Properties Analysis

### Supported

- `Visible` — supported via `COMMON_VISIBLE_PROPS`, correctly defaults to `true`.
- `Enabled` — supported via `COMMON_VISIBLE_PROPS`, correctly defaults to `true`. Applied as `disabled={!enabled}` on the underlying `<input type="checkbox">`.
- `Width` / `Height` — inherited from `COMMON_VISIBLE_PROPS`, both default to `-1` (automatic). Processed by `sizeStyle()`.
- `On` — supported, defaults to `false`. Correct per AI source (`defaultValue = "False"`). Coerced from string/number by `coerceSimulationValue`. Rendered as `checked={props.On}`.
- `Text` — supported, defaults to `''`. Rendered as `<span>{props.Text || ''}</span>` beside the toggle.
- `TextColor` — supported, defaults to `'&HFF222222'`. Applied via `colorValue()` in `baseStyle()`.
- `FontSize` — supported, defaults to `14`. Applied as `font-size: Xpx` in `baseStyle()`.
- `BackgroundColor` — carried through `baseStyle()` for visual rendering.

### Missing / Unsupported

| Property | AI Default | Priority | Notes |
|---|---|---|---|
| `ThumbColorActive` | `&HFFFFFFFF` (white) | **High** | Declared in `Switch.java`. Not in `SIMULATION_DEFAULTS`, `SIMULATION_VISUAL_PROPS`, or applied in rendering. Blocks that set `ThumbColorActive` will silently have no visual effect. |
| `ThumbColorInactive` | `&HFFCCCCCC` (light gray) | **High** | Same as above. AI default is `DEFAULT_VALUE_COLOR_LTGRAY`. Not tracked or rendered. |
| `TrackColorActive` | `&HFF00FF00` (green) | **High** | Declared in `Switch.java`. Not in `SIMULATION_VISUAL_PROPS` or rendered. The track colour is hardcoded via CSS `accent-color: #2563eb` (blue) regardless. |
| `TrackColorInactive` | `&HFF444444` (dark gray) | **High** | AI default is `DEFAULT_VALUE_COLOR_DKGRAY` (`&HFF444444`). Not tracked or rendered. |
| `FontBold` | `false` | Medium | Inherited from `ToggleBase`. Not in `SIMULATION_DEFAULTS`, `SIMULATION_VISUAL_PROPS`, or rendered. Blocks setting bold on the label have no effect. |
| `FontItalic` | `false` | Medium | Inherited from `ToggleBase`. Same gap as `FontBold`. |
| `FontTypeface` | `"0"` (default) | Low | Designer-settable typeface. Not tracked or applied. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `TextColor` | `&HFF000000` (black) | `&HFF222222` | Low | AI default from `ToggleBase` is `DEFAULT_VALUE_COLOR_BLACK` (`&HFF000000`). Tensor uses a near-black `#222222`. Visually minor but technically incorrect. |
| `BackgroundColor` | `&H00FFFFFF` (transparent/none) | Not in `SIMULATION_DEFAULTS` | Low | AI default from `ToggleBase.initToggle()` calls `BackgroundColor(Component.COLOR_NONE)` which is `&H00FFFFFF`. Tensor does not initialise this default, so no background is applied (functionally correct but the explicit default is missing). |

---

## Events Analysis

### Supported

- `Changed` — fired correctly from `checkboxInput()` when the user toggles the switch. The `emitInteraction` call simultaneously updates the `On` property and fires `Changed` with no args, matching the AI signature (`@SimpleEvent` in `Switch.java` which overrides `ToggleBase.Changed()`).

### Missing / Incorrect

| Event | AI Source | Priority | Notes |
|---|---|---|---|
| `GotFocus` | Declared in `ToggleBase` | **High** | Not wired for Switch. The `on:focus` handler exists **only** on the `TextBox`/`PasswordTextBox` branch (lines 271–272). The `CheckBox`/`Switch` branch has no focus event listeners, so `GotFocus` is never fired when the toggle gains focus. |
| `LostFocus` | Declared in `ToggleBase` | **High** | Same gap as `GotFocus` — no `on:blur` handler in the CheckBox/Switch branch. |

---

## Methods Analysis

### Supported

- (none required) — `Switch` and its `ToggleBase` superclass declare no `@SimpleFunction` callable methods. The simulation correctly routes unknown methods for Switch through `h.Unsupported()`.

### Missing / Incorrect

- No method gaps identified. Switch has no `@SimpleFunction` members in the AI source.

---

## Behaviour Gaps

### 1. Thumb and Track Colours Completely Absent (High)

The most prominent visual characteristic distinguishing a Switch from a plain CheckBox in App Inventor is its coloured thumb (white when ON, light gray when OFF) and coloured track (green when ON, dark gray when OFF). In Tensor, none of these four colour properties exist anywhere in the simulation pipeline. The rendered toggle uses browser-default or CSS `accent-color: #2563eb` appearance regardless of what any AI block sets. A user who sets `Switch1.TrackColorActive` to red in Blocks will see no change in the simulator.

### 2. `accent-color` Hardcoded to Blue (Medium)

The CSS rule `.sim-switch input { accent-color: #2563eb; }` sets the active track/thumb tint unconditionally. This is not the App Inventor default green (`#00FF00`). Even before dynamic colour support is added, the hardcoded colour should be changed to the AI default green (`#00FF00`) for the active state so the simulator looks like the platform default.

### 3. `GotFocus` / `LostFocus` Never Fire (High)

`ToggleBase` registers an `OnFocusChangeListener` in `initToggle()`, so AI's Switch fires `GotFocus` and `LostFocus` when keyboard or accessibility focus moves. In the Svelte template the CheckBox/Switch `<label>` element and its inner `<input type="checkbox">` have no `on:focus` or `on:blur` event bindings. This is a regression compared with `TextBox` which does have these handlers.

### 4. Font Style Properties Not Rendered (Medium)

`FontBold`, `FontItalic`, and `FontTypeface` from `ToggleBase` are designer-settable properties that affect the text label next to the toggle. Tensor's `baseStyle()` helper only sets `font-size`; it does not apply `font-weight`, `font-style`, or `font-family`. Blocks that modify these properties will update the runtime state but the label will not reflect the change visually.

### 5. `On` Property Not in `SIMULATION_VISUAL_PROPS` Consistently Handled but Colour Props Not Listed

`On` is correctly listed in `SIMULATION_VISUAL_PROPS` (line 69 of `simulation-capabilities.js`) and is coerced by `coerceSimulationValue`. This part of the pipeline is correct. However `ThumbColorActive`, `ThumbColorInactive`, `TrackColorActive`, and `TrackColorInactive` are not in `SIMULATION_VISUAL_PROPS`, which means even if `deriveStateFromDesignerProps` is called with those values they will be silently dropped before reaching the component state.

### 6. `BackgroundColor` Default Initialisation Missing

`ToggleBase.initToggle()` explicitly calls `BackgroundColor(Component.COLOR_NONE)` meaning the default is transparent (`&H00FFFFFF`). Tensor's `SIMULATION_DEFAULTS.Switch` does not include a `BackgroundColor` key. When `baseStyle()` tests `props.BackgroundColor`, it finds `undefined` and emits nothing, which is coincidentally the right visual behaviour, but a Block reading `Switch1.BackgroundColor` before any assignment will get `null` rather than `&H00FFFFFF`.

---

## Bugs Found

### Bug 1 — Wrong `TrackColorInactive` default in `Switch.java` (Documentation)
`Switch.java` line 77 calls `TrackColorInactive(Component.COLOR_GRAY)` which is `&HFF888888`, but the `@DesignerProperty` annotation on `TrackColorInactive()` specifies `defaultValue = Component.DEFAULT_VALUE_COLOR_DKGRAY` (`&HFF444444`). The constructor and the designer default are inconsistent in the upstream source. Tensor does not model either value, so it sidesteps the bug but also never presents the correct default.

### Bug 2 — `ThumbColorInactive` constructor/designer default mismatch in AI source (Documentation)
`Switch.java` line 76 calls `ThumbColorInactive(Component.COLOR_LTGRAY)` (`&HFFCCCCCC`). `@DesignerProperty` also specifies `DEFAULT_VALUE_COLOR_LTGRAY`. These match, so this is not a bug — but Tensor still does not model it.

### Bug 3 — `GotFocus`/`LostFocus` wired only to TextBox (Tensor bug, High)
In `SimulationComponent.svelte` lines 271–272:
```svelte
on:focus={() => emitEvent(node.name, 'GotFocus')}
on:blur={() => emitEvent(node.name, 'LostFocus')}
```
These handlers are inside the `{:else if node.type === 'TextBox' || node.type === 'PasswordTextBox'}` block. The `{:else if node.type === 'CheckBox' || node.type === 'Switch'}` block at line 278 has no equivalent bindings. Both CheckBox and Switch should fire these events.

### Bug 4 — `colorValue()` drops alpha channel (Low, shared with all components)
`colorValue()` at line 45 of `SimulationComponent.svelte` discards the two alpha nibbles from App Inventor's `&HAABBGGRR`-style colour string and emits only `#RRGGBB`. If a user sets `ThumbColorActive` or `TrackColorActive` to a semi-transparent colour, the alpha information would be silently ignored. This is a pre-existing shared bug but it is particularly relevant for Switch where all four colour properties carry alpha.

---

## Android/App Inventor Standards Compliance

| Aspect | Status |
|---|---|
| Toggle shape/style | Non-compliant: rendered as a plain HTML checkbox, not a pill-shaped Android SwitchCompat |
| Thumb/Track colours | Non-compliant: all four colour properties are absent |
| Default active colour | Non-compliant: blue (`#2563eb`) instead of AI green (`&HFF00FF00`) |
| `On` default (`false`) | Compliant |
| `Changed` event firing | Compliant |
| Focus events (`GotFocus`/`LostFocus`) | Non-compliant: never fired |
| Font styling (bold/italic/typeface) | Non-compliant: not rendered |
| `Enabled` disabled state | Compliant |
| `Text` label beside toggle | Compliant |

The Switch component is functionally distinguishable from CheckBox only by the property name (`On` vs `Checked`) and the `accent-color` CSS tint. The two components share a rendering branch with no Switch-specific visual differentiation beyond a CSS class `.sim-switch` that sets only `accent-color`.

---

## Summary

The Switch simulator covers the bare minimum required to handle the `On` boolean state and the `Changed` event. All four Switch-specific colour properties (`ThumbColorActive`, `ThumbColorInactive`, `TrackColorActive`, `TrackColorInactive`) are completely absent from the simulation pipeline, meaning the defining visual characteristic of the component — coloured thumb and track — is not simulated at all. Additionally, `GotFocus` and `LostFocus` are never fired because focus listeners are missing from the CheckBox/Switch rendering branch.

### Top 3 Action Items

1. **(Critical) Add ThumbColorActive, ThumbColorInactive, TrackColorActive, TrackColorInactive to `SIMULATION_VISUAL_PROPS` and `SIMULATION_DEFAULTS`** and apply them in the Switch CSS via CSS custom properties so blocks that set these colours are reflected visually. At minimum the defaults should match AI: thumb active = `&HFFFFFFFF`, thumb inactive = `&HFFCCCCCC`, track active = `&HFF00FF00`, track inactive = `&HFF444444`.

2. **(High) Wire `on:focus` and `on:blur` to the `<input type="checkbox">` inside the Switch rendering branch** to fire `GotFocus` and `LostFocus` events, matching what `ToggleBase` does on Android.

3. **(Medium) Change the hardcoded `accent-color: #2563eb` for `.sim-switch`** to green (`#00ff00`) to at least approximate the App Inventor default track colour pending full colour property support.
