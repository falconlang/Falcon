# TimePicker Simulator Review

## Overview

`TimePicker` in App Inventor is a **button** (it extends `ButtonBase`, not a dedicated Picker base class) that, when clicked, opens an Android `TimePickerDialog` — a system modal. After the user picks a time, `AfterTimeSet` fires and `Hour`, `Minute`, and `Instant` are updated. The component inherits a large set of button-style properties (text, colors, fonts, shape, image) and button events (`Click`, `GotFocus`, `LostFocus`, `TouchDown`, `TouchUp`) from `ButtonBase` / `TouchComponent`.

In Tensor IDE the component is rendered as an HTML `<input type="time">` element. This is a functional approximation for web simulation — the native browser time-picker dialog is close enough in behaviour to the Android dialog. However, there are several gaps in properties, events, and methods documented below.

---

## Properties Analysis

### Supported

- `Visible` — default `true`, correctly inherited from `COMMON_VISIBLE_PROPS`.
- `Enabled` — default `true`, propagated to the `disabled` attribute of the `<input>`.
- `Width` / `Height` — size encoding (-1 fill, -2 wrap, -1000x percent) correctly applied via `sizeStyle()`.
- `Text` — default `'TimePicker'`, included in `SIMULATION_DEFAULTS`; written back as a "HH:MM" string on `timeChange`. This is a reasonable simulation-only label, though in AI the `Text` controls the button label.
- `Hour` — default `0`, coerced as a number via `coerceSimulationValue`, read by `timeText()`, written back on `timeChange`. Range 0–23. Correctly handled.
- `Minute` — default `0`, coerced as a number, read by `timeText()`, written back on `timeChange`. Range 0–59. Correctly handled.

### Missing / Unsupported

| Property | AI Source | Priority |
|---|---|---|
| `Instant` | Read-only `Calendar` object from `Dates.TimeInstant(hour, minute)`. Updated every time the picker dialog commits a value, and also by `SetTimeToDisplay` / `SetTimeToDisplayFromInstant`. Not modelled in Tensor at all — no `Instant` key in defaults, visual props, or state patches. | **High** |
| `FontSize` | Inherited from `ButtonBase` (`@SimpleProperty`). Used by the AI designer to control the button label font size. Not present in `SIMULATION_DEFAULTS` for `TimePicker`. | **Medium** |
| `TextColor` | Inherited from `ButtonBase`. Not present in `SIMULATION_DEFAULTS` for `TimePicker`. | **Medium** |
| `BackgroundColor` | Inherited from `TouchComponent`. Not in defaults. | **Medium** |
| `Image` | Background image for the button, inherited from `TouchComponent` (`@SimpleProperty`, asset path). Not in defaults or visual-props handling for `TimePicker`. | **Low** |
| `FontBold` | Inherited from `ButtonBase`. Not modelled. | **Low** |
| `FontItalic` | Inherited from `ButtonBase`. Not modelled. | **Low** |
| `FontTypeface` | Inherited from `ButtonBase`. Not modelled. | **Low** |
| `TextAlignment` | Inherited from `ButtonBase` (`userVisible = false` but still a designer property). Not modelled. | **Low** |
| `Shape` | Inherited from `ButtonBase` (`userVisible = false`). Not modelled. | **Low** |

### Wrong Defaults or Types

| Property | Tensor Default | AI Default / Behaviour | Priority |
|---|---|---|---|
| `Hour` initial value | `0` (hard-coded in `SIMULATION_DEFAULTS`) | In AI the initial hour is `Calendar.getInstance().get(Calendar.HOUR_OF_DAY)` — i.e. **current wall-clock hour** at construction time. | **Medium** |
| `Minute` initial value | `0` (hard-coded in `SIMULATION_DEFAULTS`) | In AI the initial minute is `Calendar.getInstance().get(Calendar.MINUTE)` — i.e. **current wall-clock minute** at construction time. | **Medium** |

---

## Events Analysis

### Supported

- `AfterTimeSet` — correctly emitted (no args) from `timeChange()` in `SimulationComponent.svelte` (line 218). Correct event name, correct no-argument signature. Backend `dispatchSimulationEvent` routes it to user handlers.

### Missing / Incorrect

| Event | AI Source | Priority |
|---|---|---|
| `Click` | Inherited from `ButtonBase` / the underlying Android `Button`. The AI `TimePicker` *also* fires a `Click` event when the button is tapped (before the dialog opens). In Tensor the `<input type="time">` element does not emit a `Click` event; clicking the input just opens the browser picker. | **High** |
| `GotFocus` | `@SimpleEvent` in `ButtonBase` — fires when the cursor/focus moves over the button. The `<input type="time">` receives browser focus but `SimulationComponent.svelte` attaches no `on:focus` handler for `TimePicker` (only `TextBox` has focus handlers). | **Medium** |
| `LostFocus` | `@SimpleEvent` in `ButtonBase` — symmetric to `GotFocus`. Same gap. | **Medium** |
| `TouchDown` | `@SimpleEvent` in `TouchComponent` — fires on press. Not emitted for `TimePicker`; only wired for `Slider`. | **Low** |
| `TouchUp` | `@SimpleEvent` in `TouchComponent` — fires on release. Same gap as `TouchDown`. | **Low** |

---

## Methods Analysis

### Supported

- `LaunchPicker()` — handled in `simulation_wasm.go` (`case "LaunchPicker":`). Appends a `component-action` "open" effect, which `SimulationComponent.svelte` detects via the `actions` reactive expression (line 39) and calls `triggerNativePicker(timeInput)`. This correctly programmatically opens the browser time picker dialog. Implementation is correct.
- `SetTimeToDisplay(hour, minute)` — handled in `simulation_wasm.go` (lines 231–235). Sets `Hour` and `Minute` properties on the host state. Correct method name and argument count. **However** — see Bug #1 below.

### Missing / Incorrect

| Method | AI Source | Priority |
|---|---|---|
| `SetTimeToDisplayFromInstant(instant)` | `@SimpleFunction` in `TimePicker.java`. Accepts a `Calendar` instant, extracts hour and minute via `Dates.Hour()` / `Dates.Minute()`, updates the dialog. Completely absent from `simulation_wasm.go` — no `"SetTimeToDisplayFromInstant"` case in the `TimePicker` switch. | **High** |

---

## Behaviour Gaps

### 1. `customTime` flag / dialog pre-population semantics

In App Inventor, `SetTimeToDisplay` sets an internal `customTime = true` flag. The next time the dialog opens it shows that custom time, and then clears the flag (reverting to current time for subsequent opens). This two-step "custom then current" behaviour is not replicated. In Tensor, `SetTimeToDisplay` directly patches `Hour`/`Minute` in the persistent state; the values remain set even after the picker is opened and dismissed without selecting a new time.

**Priority: Medium**

### 2. `AfterTimeSet` fires only if `view.isShown()`

In App Inventor, the `onTimeSet` callback guards dispatch with `if (view.isShown())`. This prevents double-firing that can occur on some Android OS versions. Tensor has no such guard — `AfterTimeSet` fires unconditionally on every HTML `change` event. This is usually acceptable, but the nuance is lost.

**Priority: Low**

### 3. Initial `Instant` is not set from current time

On Android the `instant` field is populated with `Dates.TimeInstant(hour, minute)` at construction (using the actual current time). In Tensor, `Instant` is never set at all, so any block reading `TimePicker1.Instant` returns null.

**Priority: High**

### 4. `Text` property write-back on timeChange

When the user picks a time, `timeChange()` writes `Text` as a raw "HH:MM" string (e.g. `"09:35"`). In App Inventor the `Text` property of the button is the designer-set button label (e.g. "Set Time") and is never overwritten by picker activity. The write-back to `Text` is a Tensor-specific artefact that could corrupt the button label if the user has set a custom label.

**Priority: Medium**

### 5. 24-hour vs 12-hour format rendering

App Inventor uses `DateFormat.is24HourFormat(context)` — the dialog respects the device locale's 12/24-hour setting. The browser `<input type="time">` element also respects the user's locale for display, so this aspect is handled correctly by the web platform.

**Priority: Low (informational)**

### 6. Hour/Minute default initialisation (current time vs 0)

`SIMULATION_DEFAULTS` hard-codes `Hour: 0, Minute: 0`. On real App Inventor the initial values come from `Calendar.getInstance()` at component construction — the current wall-clock time. Users who read `TimePicker1.Hour` before opening the dialog will get `0` in Tensor but the actual current hour in production.

**Priority: Medium**

---

## Bugs Found

### Bug 1 — `SetTimeToDisplay` does not validate arguments (simulation_wasm.go:231–235)

```go
case "SetTimeToDisplay":
    if len(args) >= 2 {
        h.SetProperty(componentName, componentType, "Hour", args[0])
        h.SetProperty(componentName, componentType, "Minute", args[1])
    }
    return runtime.VoidVal()
```

In App Inventor, `SetTimeToDisplay` validates that `hour` is in `[0, 23]` and `minute` is in `[0, 59]`, and calls `form.dispatchErrorOccurredEvent` with `ERROR_ILLEGAL_HOUR` / `ERROR_ILLEGAL_MINUTE` if out-of-range. The Tensor implementation silently writes any value. A script passing `hour = 25` will corrupt the time state without any error signal.

**Priority: High**

### Bug 2 — `SetTimeToDisplayFromInstant` is entirely absent (simulation_wasm.go)

The `TimePicker` switch in `CallMethod` has no `"SetTimeToDisplayFromInstant"` case. Any block calling this method falls through to `h.Unsupported("method", ...)`, marking the component unsupported and silently doing nothing. This method is a documented `@SimpleFunction` in the AI source.

**Priority: High**

### Bug 3 — `Text` property incorrectly overwritten on every time selection (SimulationComponent.svelte:214–216)

```js
{ component: node.name, property: 'Text', value: nextTime },
```

`timeChange()` always patches `Text` with the raw `"HH:MM"` string. There is no equivalent in App Inventor — the `Text` property is the static button label. This creates a spurious side-effect: after the first pick, subsequent reads of `TimePicker1.Text` in blocks return the last picked time string, not the designed label.

**Priority: Medium**

### Bug 4 — `Instant` property is never populated or returned (simulation-capabilities.js + simulation_wasm.go)

`Instant` is absent from `SIMULATION_DEFAULTS`, `SIMULATION_VISUAL_PROPS`, `coerceSimulationValue`, and there is no logic to construct a Calendar-equivalent value after a pick. Any block that reads `TimePicker1.Instant` returns `null`, causing downstream Clock/Date arithmetic to fail silently.

**Priority: High**

### Bug 5 — `LaunchPicker()` does not fire `Click` event first

In App Inventor, `LaunchPicker()` calls `click()` which in `ButtonBase` chains through `onClick` → `click()` (implemented in TimePicker, which calls `time.show()`). On the AI platform, a `When TimePicker1.Click` handler fires *before* the dialog opens (since `LaunchPicker` calls `click()`). In Tensor, `LaunchPicker` only emits the `component-action "open"` effect; the `Click` event is not fired at all, so any `When TimePicker1.Click` handler attached to `LaunchPicker` will not execute.

**Priority: Medium**

---

## Android/App Inventor Standards Compliance

| Area | Assessment |
|---|---|
| Dialog semantics | **Partial.** A browser `<input type="time">` is a reasonable approximation of `TimePickerDialog`, but it is always inline/overlay rather than a system modal. The programmatic `showPicker()` approach via `triggerNativePicker` is correct for the web context. |
| 24-hour locale handling | **Adequate.** The browser input type respects the user locale, equivalent to AI's `DateFormat.is24HourFormat()`. |
| ButtonBase inheritance | **Poor.** TimePicker is a Button in AI — it should render as a button that opens the dialog. Tensor renders it as a bare `<input type="time">` without a button label, ignoring all inherited button properties and events (`Click`, `GotFocus`, `LostFocus`, `Text`, `TextColor`, `FontSize`, `BackgroundColor`, `Image`). |
| Error dispatching | **Missing.** `SetTimeToDisplay` should dispatch `ErrorOccurred` for invalid hours/minutes; Tensor does not. |
| `Instant` Calendar object | **Missing.** Tensor has no representation of the `Calendar` instant object, which is required for interoperability with the Clock component. |

---

## Summary

TimePicker has a working core: picking a time via the browser input fires `AfterTimeSet` and correctly updates `Hour` and `Minute`. `LaunchPicker()` correctly triggers the picker programmatically. These are the most-used interactions and they work.

The critical gaps are:

1. **`Instant` property is completely absent** — any block using `TimePicker1.Instant` with Clock component arithmetic will silently receive `null`. This is the most impactful gap since AI developers frequently use the Instant for date/time formatting. *(Critical)*
2. **`SetTimeToDisplayFromInstant` method is unimplemented** — calling it marks the component as "unsupported" and does nothing. *(High)*
3. **`SetTimeToDisplay` has no argument validation** — out-of-range hours/minutes are written silently instead of firing `ErrorOccurred`. *(High)*

### Top 3 Action Items

1. **Implement `Instant` property** — After a time is picked (or set via `SetTimeToDisplay`), compute and store an `Instant`-equivalent value (e.g. a numeric timestamp or structured object). Expose it as a readable property. This unblocks all Clock-based blocks.
2. **Implement `SetTimeToDisplayFromInstant(instant)`** — Add a `"SetTimeToDisplayFromInstant"` case in the `TimePicker` switch in `simulation_wasm.go` that extracts hour/minute from the instant and sets them on state.
3. **Remove the spurious `Text` write-back in `timeChange()`** — The `Text` property should not be overwritten with the picked time string; this corrupts the button label. If a display of the selected time is needed, only `Hour` and `Minute` (and eventually `Instant`) should be updated.
