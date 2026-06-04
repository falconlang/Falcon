# Notifier Simulator Review

## Overview

The Notifier is a non-visible App Inventor component that provides alert dialogs, toast messages, progress dialogs, and Android system log output. It has no visual presence in the designer — it is invoked entirely through method calls at runtime.

In the Tensor IDE simulator, Notifier is included in `SIMULATION_SUPPORTED_TYPES` and `SIMULATION_NONVISIBLE_TYPES`. Runtime method handling lives in `simulation_wasm.go` (`CallMethod`), and toast ("notice") effects are rendered in `SimulateOverlay.svelte`. The component has no visual rendering in `SimulationComponent.svelte` (correctly excluded via `nonVisible`).

---

## Properties Analysis

### Supported
- `Visible: false` — Notifier is correctly listed as non-visible and carries a default of `{ Visible: false }` in `SIMULATION_DEFAULTS`.

### Missing / Unsupported

| Property | AI Default | Priority | Notes |
|---|---|---|---|
| `NotifierLength` | `1` (TOAST_LENGTH_LONG) | **Medium** | Controls whether the toast duration is "short" (0) or "long" (1). Not stored in simulation state at all; no `NotifierLength` key exists in `SIMULATION_DEFAULTS.Notifier`. The notice always displays with no auto-dismiss at all (manual dismiss only), making the length property irrelevant but the absence means any block that reads or sets `NotifierLength` silently fails. |
| `BackgroundColor` | `&HFF444444` (DKGRAY) | **Medium** | The AI specifies a dark-grey background for toast alerts. The simulator's `sim-notice` uses a hardcoded `rgba(32,33,36,0.94)` rather than reading from this property. Blocks that set `BackgroundColor` are silently ignored. |
| `TextColor` | `&HFFFFFFFF` (WHITE) | **Medium** | Text color for toast alerts. The `sim-notice` uses hardcoded `color: #fff`. Blocks that set `TextColor` are silently ignored. |

### Wrong Defaults or Types

| Property | Tensor Default | AI Default | Priority | Notes |
|---|---|---|---|---|
| `Visible` | `false` | N/A (non-visible) | **Low** | The explicit `Visible: false` entry is harmless but nonstandard — non-visible components in AI have no Visible property at all. This could cause a block reading `Notifier1.Visible` to return `false` rather than raising an unsupported error. |

---

## Events Analysis

### Supported
- None of the Notifier's events are actively dispatched by the simulator, which is consistent with the fact that dialogs (`ShowMessageDialog`, `ShowChooseDialog`, `ShowTextDialog`, `ShowPasswordDialog`) are not simulated. There is no event routing bug here — it is a scope limitation.

### Missing / Incorrect

| Event | AI Args | Priority | Notes |
|---|---|---|---|
| `AfterChoosing(choice)` | `choice: String` | **Critical** | `ShowChooseDialog` is completely unimplemented. When called, it falls through to `h.Unsupported(...)`. The `AfterChoosing` event is therefore never fired. Any app logic that branches on dialog choice cannot be tested. |
| `ChoosingCanceled()` | none | **Critical** | Companion event for cancelable `ShowChooseDialog`. Never fired because the method is not implemented. |
| `AfterTextInput(response)` | `response: String` | **Critical** | `ShowTextDialog` and `ShowPasswordDialog` are completely unimplemented. The event is never fired. |
| `TextInputCanceled()` | none | **Critical** | Companion cancel event for `ShowTextDialog`/`ShowPasswordDialog`. Never fired. |

---

## Methods Analysis

### Supported
- `ShowAlert(notice: String)` — Correctly implemented. Appends a `"notice"` effect with `type`, `component`, and `text` fields. The overlay in `SimulateOverlay.svelte` renders this as a dismissible toast at the bottom of the phone frame. **Partially correct** — see Behaviour Gaps for duration and styling issues.
- `LogError(message: String)` — Falls through to `h.Unsupported(...)`, which surfaces as an unsupported-method warning in the overlay. This is a reasonable no-op for a simulator.
- `LogWarning(message: String)` — Same as above.
- `LogInfo(message: String)` — Same as above.

### Missing / Incorrect

| Method | AI Signature | Priority | Notes |
|---|---|---|---|
| `ShowMessageDialog(message, title, buttonText)` | 3 × String → void | **Critical** | Not handled in `CallMethod`. Falls to `h.Unsupported(...)`. Any app that uses a one-button confirmation dialog cannot be simulated. |
| `ShowChooseDialog(message, title, button1Text, button2Text, cancelable)` | 4 × String + bool → void | **Critical** | Not handled. Falls to `h.Unsupported(...)`. The two-choice dialog — one of the most commonly used Notifier features — cannot be simulated at all. |
| `ShowTextDialog(message, title, cancelable)` | 2 × String + bool → void | **Critical** | Not handled. Falls to `h.Unsupported(...)`. Text-input dialogs cannot be tested. |
| `ShowPasswordDialog(message, title, cancelable)` | 2 × String + bool → void | **High** | Not handled. Falls to `h.Unsupported(...)`. Shares the same simulation gap as `ShowTextDialog`. |
| `ShowProgressDialog(message, title)` | 2 × String → void | **High** | Not handled. Falls to `h.Unsupported(...)`. Apps that display a loading spinner while performing async-like operations cannot be tested. |
| `DismissProgressDialog()` | void → void | **High** | Not handled. Companion to the missing `ShowProgressDialog`. |
| `LogError(message)` | String → void | **Low** | Reported as unsupported instead of routed to the simulator console/logs. A `host.logs` entry would be more useful and faithful. |
| `LogWarning(message)` | String → void | **Low** | Same as `LogError`. |
| `LogInfo(message)` | String → void | **Low** | Same as `LogError`. |

---

## Behaviour Gaps

### 1. Toast auto-dismiss not implemented (High)

In App Inventor, `ShowAlert` produces an Android `Toast` that disappears automatically after a short (`TOAST_LENGTH_SHORT` ≈ 2 s) or long (`TOAST_LENGTH_LONG` ≈ 3.5 s) interval, driven by the `NotifierLength` property. In the simulator, the notice banner is persistent — it only disappears when the user clicks the "x" dismiss button. There is no auto-timeout at all.

**Location:** `SimulateOverlay.svelte`, `applyEffects()` — the `"notice"` effect pushes to `notices[]` but never schedules a removal timeout.

### 2. Toast styling ignores BackgroundColor and TextColor (Medium)

`toastNow()` in the Java source applies `textView.setBackgroundColor(backgroundColor)` and `textView.setTextColor(textColor)` using the `BackgroundColor` and `TextColor` properties (defaulting to DKGRAY and WHITE respectively). The simulator hardcodes `background: rgba(32,33,36,0.94)` and `color: #fff` in `.sim-notice` CSS, completely ignoring the runtime property values.

**Location:** `SimulateOverlay.svelte` lines 598–615.

### 3. Toast capped at 3 visible notices (Low)

The overlay truncates to 3 notices with `notices.slice(-2)` (keeping the last 2 plus the new one = max 3). App Inventor queues toasts sequentially without a hard cap. While capping at 3 is a reasonable UX choice for the simulator, the silent discard of older notices may confuse app authors.

**Location:** `SimulateOverlay.svelte` line 176, `notices.slice(-2)`.

### 4. Dialog methods are fire-and-forget unsupported markers (Critical)

`ShowMessageDialog`, `ShowChooseDialog`, `ShowTextDialog`, `ShowPasswordDialog`, `ShowProgressDialog` all fall into `h.Unsupported("method", componentName+"."+methodName)`. This means:
- The unsupported-method indicator appears in the overlay.
- Execution continues past the call immediately, as if the dialog was instantly dismissed.
- No events (`AfterChoosing`, `AfterTextInput`, `TextInputCanceled`, `ChoosingCanceled`) are ever fired.

Apps designed around dialogs are completely non-functional in the simulator — there is no workaround short of not using Notifier dialogs.

### 5. Log methods reported as unsupported instead of logged (Low)

`LogError`, `LogWarning`, and `LogInfo` call `h.Unsupported(...)` in `CallMethod`, causing them to appear in the unsupported-method panel. The correct behaviour would be to append to `host.logs` (which the overlay already surfaces via `result.logs` in the session result fields, though the overlay does not currently display logs). At minimum, they should be no-ops rather than unsupported markers, since logging is side-effect-only and does not affect app state.

**Location:** `simulation_wasm.go` — the default fall-through in `CallMethod` at line 242.

### 6. ShowChooseDialog cancel semantics not modelled (High)

In App Inventor, when `cancelable=true` the CANCEL button raises **both** `ChoosingCanceled()` (new event, AI2 nb186+) **and** `AfterChoosing("Cancel")` (for backward compatibility). Any simulator implementation must fire both events in the correct order. This dual-fire behaviour is easy to miss.

### 7. TextInputCanceled return value not modelled (Medium)

`TextInputCanceled()` in the Java source has a boolean return value — `EventDispatcher.dispatchEvent(...)` returns `true` if any handler consumed the event. The dialog implementation then uses this return value to decide whether to also fire `AfterTextInput("Cancel")` (backward compatibility). A simulator that does not model this conditional chain will produce incorrect event sequences for apps that handle `TextInputCanceled`.

---

## Bugs Found

### Bug 1 — `ShowAlert` effect type mismatch guard (Low)

In `simulation_wasm.go` line 186, the Notifier branch is guarded by:

```go
if (componentType == "Notifier" || componentType == "") && method == "ShowAlert" {
```

The `componentType == ""` arm is overly broad: if any unknown component's `CallMethod` is called with method name `ShowAlert` and an empty component type, it will silently emit a notice effect instead of routing to the unsupported handler. This could mask type-resolution failures.

### Bug 2 — Notice effect carries no duration or styling metadata (Medium)

The effect payload produced for `ShowAlert` is:

```go
map[string]any{
    "type":      "notice",
    "component": componentName,
    "text":      text,
}
```

It carries no `duration`, `backgroundColor`, or `textColor` fields. Even if the frontend were updated to honour these properties, the backend would need to include them. This is a design gap that makes future correct implementation harder.

### Bug 3 — `SIMULATION_DEFAULTS.Notifier` is missing required properties (Medium)

`simulation-capabilities.js` line 59:

```js
Notifier: { Visible: false },
```

Properties `NotifierLength`, `BackgroundColor`, and `TextColor` are absent. Any call to `GetProperty` on these names returns `runtime.NullVal()` from `simulation_wasm.go` line 78. If a user's block reads `Notifier1.NotifierLength` or sets `Notifier1.BackgroundColor`, no error is raised and no meaningful value is returned or stored.

### Bug 4 — `coerceSimulationValue` has no Notifier-aware entries (Low)

`simulation-capabilities.js` `coerceSimulationValue()` has no entries for `NotifierLength`, `BackgroundColor`, or `TextColor` of the Notifier. If these properties were added to the defaults, numeric coercion for `NotifierLength` and color-string coercion for `BackgroundColor`/`TextColor` would need to be added to prevent type errors at runtime.

---

## Android/App Inventor Standards Compliance

| Standard | Status |
|---|---|
| Non-visible component (no visual render) | Correct |
| `BackgroundColor` default `&HFF444444` (DKGRAY) | Not applied to toast rendering |
| `TextColor` default `&HFFFFFFFF` (WHITE) | Not applied to toast rendering |
| `NotifierLength` default `TOAST_LENGTH_LONG = 1` | Property not modelled; no auto-dismiss |
| Toast gravity: `Gravity.CENTER` | Simulator renders notices at bottom — acceptable UX deviation |
| Dialog `setCancelable(false)` — Back key cannot dismiss | Not applicable in browser simulator |
| `Html.fromHtml()` message formatting in dialogs | Dialogs not implemented; N/A |
| `ChoosingCanceled` + `AfterChoosing("Cancel")` dual-fire | Not implemented |
| `TextInputCanceled` boolean return gating `AfterTextInput` | Not implemented |
| `ShowProgressDialog` non-cancelable spinner | Not implemented |

---

## Summary

The Notifier simulator implementation is **minimally functional**: only `ShowAlert` is handled, and even that implementation has gaps (no auto-dismiss, no property-driven styling). All six dialog methods (`ShowMessageDialog`, `ShowChooseDialog`, `ShowTextDialog`, `ShowPasswordDialog`, `ShowProgressDialog`, `DismissProgressDialog`) are completely absent, causing any app that uses Notifier dialogs to silently fail — no dialog appears, no events fire, and execution falls through immediately.

**Total issues found: 19**
- Critical: 4 (missing dialog methods + missing dialog events)
- High: 6 (ShowPasswordDialog, ShowProgressDialog, DismissProgressDialog, cancel semantics, auto-dismiss, dual-fire)
- Medium: 6 (BackgroundColor, TextColor, NotifierLength properties, TextInputCanceled return, notice effect metadata, missing defaults)
- Low: 3 (LogError/LogWarning/LogInfo routing, Visible default, componentType guard)

### Top 3 Action Items

1. **Implement modal dialog simulation** — Add browser `<dialog>` or overlay-based implementations for `ShowMessageDialog`, `ShowChooseDialog`, `ShowTextDialog`, and `ShowPasswordDialog` in both `simulation_wasm.go` (produce `"dialog"` effects with type and button configuration) and `SimulateOverlay.svelte` (render the dialog, fire the appropriate events on button press). This is the single most impactful improvement.

2. **Add auto-dismiss timeout for ShowAlert** — In `SimulateOverlay.svelte` `applyEffects()`, schedule a `setTimeout` (≈2000 ms for short, ≈3500 ms for long) keyed on the notice `id` to automatically remove it from `notices[]`, matching the Android Toast behaviour. Also pass `NotifierLength`, `BackgroundColor`, and `TextColor` in the effect payload from the Go backend and apply them in the rendered `.sim-notice`.

3. **Route LogError/LogWarning/LogInfo to the simulator console** — In `simulation_wasm.go` `CallMethod`, handle these three methods by appending to `host.logs` (e.g. `"[Error] " + message`) instead of calling `h.Unsupported(...)`. This removes false-positive unsupported warnings and makes log-based debugging visible in the simulator output.
