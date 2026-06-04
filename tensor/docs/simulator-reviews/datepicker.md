# DatePicker Simulator Review

## Overview

`DatePicker` in App Inventor is a button that, when clicked, opens an Android `DatePickerDialog` (a system modal). The user selects a date; the dialog fires `AfterDateSet` and updates the `Year`, `Month`, `Day`, `MonthInText`, and `Instant` read-only properties.

In Tensor IDE the component is rendered as an HTML `<input type="date">` element. This is a reasonable web approximation, but it diverges from the AI specification in several important ways documented below.

---

## Properties Analysis

### Supported

- `Visible` — default `true`, correctly inherited from `COMMON_VISIBLE_PROPS`.
- `Enabled` — default `true`, propagated to the `disabled` attribute of the input.
- `Width` / `Height` — size encoding (-1 fill, -2 wrap, -1000x percent) correctly applied via `sizeStyle()`.
- `Text` — included in defaults (`'DatePicker'`), written back on `dateChange`. Used as button label in AI; not meaningful for an `<input type="date">` element.
- `Day` — default `1`, coerced as number, read by `dateText()`, written back on change.
- `Month` — default `1`, coerced as number, read by `dateText()`, written back on change.
- `Year` — default `1970`, coerced as number, read by `dateText()`, written back on change.

### Missing / Unsupported

| Property | AI Source | Priority |
|---|---|---|
| `MonthInText` | Read-only string; name of the month (e.g. `"January"`) from `DateFormatSymbols`. Never written or readable in Tensor. | High |
| `Instant` | Read-only `Calendar` object representing the selected date instant. Not modelled in Tensor at all. | High |
| `FontSize` | Inherited from `ButtonBase`. Used by AI designer for the button label. Not in `SIMULATION_DEFAULTS` for `DatePicker`. | Medium |
| `TextColor` | Inherited from `ButtonBase`. Not in defaults. | Medium |
| `BackgroundColor` | Inherited from `ButtonBase`. Not in defaults. | Low |
| `FontBold` | Inherited from `ButtonBase`. | Low |
| `FontItalic` | Inherited from `ButtonBase`. | Low |
| `FontTypeface` | Inherited from `ButtonBase`. | Low |
| `Image` | Button image property inherited from `ButtonBase`. Not supported. | Low |
| `Shape` | Button shape (default, rounded, rectangular, oval) from `ButtonBase`. Not supported. | Low |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority |
|---|---|---|---|
| `Year` | Current device year at creation (e.g. 2025) | `1970` | High |
| `Month` | Current device month (1-based, 1–12) | `1` | High |
| `Day` | Current device day of month | `1` | High |
| `Text` | `'Set Date'` (inherited ButtonBase label set in the designer by default) | `'DatePicker'` | Low |

**Note on Year/Month/Day defaults:** In App Inventor, the DatePicker is initialised with the *current date* from `Calendar.getInstance()`. Tensor hard-codes epoch defaults (`1970-01-01`). When the simulator starts, the date shown will be wrong unless user code calls `SetDateToDisplay` first. This is a visible, user-facing discrepancy on every app that uses a DatePicker without pre-setting the date.

---

## Events Analysis

### Supported

- `AfterDateSet` — fired after the user confirms date selection (emitted by `dateChange()` in `SimulationComponent.svelte` with 0 arguments). This matches the AI specification exactly.

### Missing / Incorrect

| Event | AI Source | Tensor Status | Priority |
|---|---|---|---|
| `Click` (inherited ButtonBase) | Fires when the button is physically clicked, before the dialog opens. Useful for programmatic response. | Not emitted for `DatePicker`; only `AfterDateSet` path exists. Clicking the `<input>` does not emit a `Click` event. | Medium |
| `GotFocus` / `LostFocus` (ButtonBase) | Standard Android focus events inherited from ButtonBase. | Not wired for `DatePicker`. TextBox wires them; DatePicker does not. | Low |
| `TouchDown` / `TouchUp` (ButtonBase) | Inherited touch events. | Not wired. | Low |

---

## Methods Analysis

### Supported

- `LaunchPicker` — triggers `componentAction(componentName, "open")` which causes `triggerNativePicker(dateInput)` in the frontend. The native browser date picker opens. Behaviour is functionally correct.
- `SetDateToDisplay(year, month, day)` — sets `Year`, `Month`, `Day` state properties, which updates `dateText()` and therefore the visible `<input type="date">` value. Functionally correct for displaying a pre-set date.

### Missing / Incorrect

| Method | AI Source | Tensor Status | Priority |
|---|---|---|---|
| `SetDateToDisplayFromInstant(instant)` | Takes a `Calendar` instant and extracts year/month/day to pre-populate the picker. | Not implemented in `CallMethod` switch in `simulation_wasm.go`. Falls through to `h.Unsupported(...)`. | High |
| `SetDateToDisplay` — validation | AI wraps date construction in `GregorianCalendar` with `setLenient(false)` and dispatches `ErrorOccurredEvent` on invalid dates. | Tensor's `SetDateToDisplay` blindly applies the arguments with no validation. Invalid dates like month=13 or day=32 would be silently set. | Medium |
| `SetDateToDisplay` — `customDate` flag | AI sets a `customDate = true` flag so the picker opens on the custom date rather than today's date. On next open without `SetDateToDisplay`, it resets to today. | Tensor directly sets `Year`/`Month`/`Day` in state and re-renders; the `customDate` reset-on-open logic is absent. This is a minor semantic difference. | Low |

---

## Behaviour Gaps

### 1. Component Visual Form: Input vs Button (Critical)

App Inventor `DatePicker` extends `ButtonBase`. It looks and behaves like a styled button. Clicking anywhere on it opens the date dialog. Tensor renders it as `<input type="date">`, which looks like a text input with a calendar icon. This is visually incorrect: the designer panel shows a button but the simulator shows an input field. Users get confused about layout and dimensions.

### 2. Initial Date Defaults to Epoch (1970-01-01) (High)

As noted above, AI initialises to the current device date. Tensor defaults to `{ Year: 1970, Month: 1, Day: 1 }`. Any app that reads `DatePicker1.Year` at startup without setting a date will get `1970` instead of the current year.

### 3. `MonthInText` Not Computed (High)

The `MonthInText` property is a computed, read-only, localised month name. AI computes it from `DateFormatSymbols().getMonths()[javaMonth]`. In Tensor, if a block reads `DatePicker1.MonthInText`, `GetProperty` will return `NullVal()` because the property is never populated. This will likely produce an empty string or runtime null in user code.

### 4. `Instant` Property Not Available (High)

`Instant` returns a `Calendar` object used heavily with Clock component methods (`Clock.FormatDate`, `Clock.FormatDateTime`, etc.). In Tensor this property is never set. Blocks that read `DatePicker1.Instant` get null, breaking all Clock-interoperability patterns.

### 5. `Text` Property Written Back on Date Change (Medium)

In `dateChange()`, the implementation writes `{ property: 'Text', value: nextDate }` (e.g. `"2025-06-04"`) back to state. In AI, `Text` is the button label (e.g. `"Pick Date"`), not a formatted date string. Writing the ISO date string to `Text` will corrupt the button label if it is subsequently read. This is a functional bug.

### 6. `AfterDateSet` Has No Arguments (Correct, but Note) (Low)

AI `AfterDateSet` takes no arguments — the values are read from properties (`Year`, `Month`, `Day`). Tensor fires the event with `args: []`, which is correct. However, since `MonthInText` and `Instant` are not populated before the event fires, any handler that reads those properties will see wrong values.

### 7. Month 1-Indexing (Correct)

Tensor correctly stores and emits Month as 1-based (January = 1), matching AI's `month = javaMonth + 1`. No issue here.

### 8. `SetDateToDisplay` Does Not Trigger Re-open Correctly (Low)

In AI, calling `SetDateToDisplay` before `LaunchPicker` causes the picker to show the custom date. Calling `LaunchPicker` a second time without calling `SetDateToDisplay` again shows today's date. Tensor has no such state machine; calling `SetDateToDisplay` simply updates `Year`/`Month`/`Day` permanently until overwritten. The second open-without-set semantics differ.

---

## Bugs Found

### Bug 1: `Text` Property Corrupted on Date Selection (Medium)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 199

```js
{ component: node.name, property: 'Text', value: nextDate },
```

`nextDate` is an ISO date string like `"2025-06-04"`. This overwrites the button's `Text` label property with a date string. If user code later reads `DatePicker1.Text`, it gets an ISO date string rather than the button caption. This property should not be written back here; `Text` is the button label, not a date display field.

### Bug 2: `MonthInText` Never Set After Date Change (High)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, `dateChange()` function (lines 191–203)

After a date is selected, `Month` is updated in state but `MonthInText` is never derived and written. Any block reading `DatePicker1.MonthInText` after `AfterDateSet` will get `null`/empty. The fix is to compute the month name (from a month-index-to-name table) in `dateChange()` and emit it as a property update alongside `Year`/`Month`/`Day`.

### Bug 3: `SetDateToDisplayFromInstant` Falls to Unsupported (High)

**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, `CallMethod` switch for `DatePicker` (lines 213–224)

`SetDateToDisplayFromInstant` is not handled in the `DatePicker` case. The method falls through to `h.Unsupported("method", ...)`. In AI this method is a `@SimpleFunction` and is commonly used to reset the picker from a stored instant. The fix is to add a `case "SetDateToDisplayFromInstant":` block that extracts Year/Month/Day from the `Instant` argument (or at minimum from individual numeric args representing a date).

### Bug 4: Initial `Year`/`Month`/`Day` Default to Epoch Instead of Current Date (High)

**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 51

```js
DatePicker: { ...COMMON_VISIBLE_PROPS, Text: 'DatePicker', Day: 1, Month: 1, Year: 1970 },
```

Defaults should be derived from the current date at simulation-start. As a minimum, JavaScript can provide `new Date()` values. Since this file is a static module, the session initialisation code should inject current date values rather than relying on static defaults.

### Bug 5: `Instant` Never Populated (High)

**Files:** `simulation-capabilities.js` (defaults), `SimulationComponent.svelte` (`dateChange`), `simulation_wasm.go` (`SetDateToDisplay`)

`Instant` is a core AI read-only property. It is never set in any part of the Tensor pipeline — not in defaults, not after `dateChange`, not in `SetDateToDisplay`. Blocks using Clock + DatePicker patterns will fail silently.

---

## Android/App Inventor Standards Compliance

| Aspect | AI Behaviour | Tensor Behaviour | Compliant? |
|---|---|---|---|
| Component form | Styled button (ButtonBase) | HTML date input | No |
| Dialog style | Android `DatePickerDialog` (modal overlay) | Browser native date picker (varies by OS/browser) | Partial |
| Month indexing | 1-based (1=January) | 1-based | Yes |
| Date on init | Current device date | 1970-01-01 | No |
| `AfterDateSet` signature | No args | No args | Yes |
| `MonthInText` | Localised month name string | Never set | No |
| `Instant` | `Calendar` object | Never set | No |
| `SetDateToDisplay` validation | Invalid dates raise `ErrorOccurredEvent` | No validation | No |
| `SetDateToDisplayFromInstant` | Supported | Not implemented | No |
| `LaunchPicker` | Opens dialog | Opens browser native picker | Partial |
| Button-inherited properties | FontSize, TextColor, BackgroundColor, etc. | Missing from defaults | No |

---

## Summary

The DatePicker simulator provides a minimal working path: a user can see a date input, change the date, and trigger `AfterDateSet` with correct `Year`/`Month`/`Day` values. `LaunchPicker` and `SetDateToDisplay` are implemented at a basic level.

However, there are significant gaps that will break real App Inventor programs:

1. **`MonthInText` and `Instant` are never set** — these properties are read by a large portion of AI apps that use DatePicker, and returning null for both will silently corrupt user programs. This is the highest-priority fix.
2. **Initial date defaults to epoch (1970-01-01) instead of the current date** — every app that reads date properties without first calling `SetDateToDisplay` will see wrong values.
3. **`Text` property is incorrectly written back as an ISO date string on date change** — this corrupts the button label and is a concrete bug in `dateChange()`.

### Top 3 Action Items

1. **Populate `MonthInText` and `Instant`** in `dateChange()`, in `SetDateToDisplay`, and at simulation initialisation. For `Instant`, a numeric timestamp or structured object approximation is sufficient for most Clock method interactions.
2. **Fix initial `Year`/`Month`/`Day` defaults** to use the current date (inject from session creation in JS side rather than static `1970` values in the capabilities file).
3. **Remove the `Text` write-back** from `dateChange()` and implement `SetDateToDisplayFromInstant` in `simulation_wasm.go`.
