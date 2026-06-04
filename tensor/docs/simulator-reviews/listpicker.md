# ListPicker Simulator Review

## Overview

ListPicker is a button-based component that, when clicked, opens a full-screen activity showing a list of choices. It inherits from `Picker → ButtonBase` in App Inventor, giving it all standard button visual/text properties plus its own list-management properties (`Elements`, `Selection`, `SelectionIndex`, `ElementsFromString`, `ShowFilterBar`, `ItemTextColor`, `ItemBackgroundColor`, `Title`). Events are `BeforePicking` (fired before the picker opens) and `AfterPicking` (fired after a choice is made). The only method is `Open()` (programmatically launch the picker).

Tensor IDE renders ListPicker as an inline dropdown menu (`sim-picker-menu`) anchored below a button; the Go host (`simulation_wasm.go`) handles `Open()`, `Selection`/`SelectionIndex` synchronisation, and `BeforePicking` emission. The core pick-and-return flow is functional, but many ButtonBase-inherited appearance properties and list-management behaviours are missing or inaccurate.

---

## Properties Analysis

### Supported

- `Text` — rendered as the button label (fallback to `props.Selection` then `'ListPicker'`).
- `Enabled` — maps to HTML `disabled` on the inner `<button>`; also checked in `openListPicker()`.
- `Visible` — gates the entire component render.
- `Width` / `Height` — size constants `-1` (auto), `-2` (fill), percentage encoding `≤ -1000` all handled via `sizeStyle()`.
- `Elements` — list of choices; rendered as dropdown items; normalised in `simulation-capabilities.js`.
- `ElementsFromString` — comma-separated string; parsed both in JS and Go host.
- `Selection` — stored and synced by `setSelection()` in Go host; updates `SelectionIndex` correctly.
- `SelectionIndex` — 1-based index; synced by `setSelectionIndex()`; clamps out-of-range to 0 with empty selection.
- `Title` — present in defaults (empty string default matches AI) and in `SIMULATION_VISUAL_PROPS`; not visually rendered in the dropdown header, but passed through.
- `BackgroundColor` — applied via `baseStyle()` to the outer `div.sim-picker` wrapper (not the inner button).
- `TextColor` — applied via `baseStyle()` to the outer wrapper (propagates to label text by CSS inheritance).
- `FontSize` — applied via `baseStyle()` to the outer wrapper.

### Missing / Unsupported

| Property | Priority | Notes |
|---|---|---|
| `ShowFilterBar` | **High** | Defined in AI Java (`showFilter`, default `false`). The search-filter bar makes the list searchable — a key interactive feature. Not in Tensor defaults, not in `SIMULATION_VISUAL_PROPS`, not rendered. |
| `ItemTextColor` | **High** | Per-item text colour (`DEFAULT_VALUE_COLOR_WHITE`, i.e. `&HFFFFFFFF`). Not in Tensor defaults at all. Dropdown items always render in the CSS-default `#202124`. |
| `ItemBackgroundColor` | **High** | Per-item background colour (`DEFAULT_VALUE_COLOR_BLACK`, i.e. `&HFF000000`). Not in Tensor defaults. Dropdown items always render with `background: #fff`. |
| `FontBold` | **Medium** | Inherited from `ButtonBase`. Default `false`. Not in defaults, not applied in CSS. |
| `FontItalic` | **Medium** | Inherited from `ButtonBase`. Default `false`. Not in defaults, not applied in CSS. |
| `FontTypeface` | **Medium** | Inherited from `ButtonBase`. Not applied. |
| `TextAlignment` | **Medium** | Inherited from `ButtonBase`. Center (1) is the AI default. Not in defaults, not rendered. |
| `Shape` | **Medium** | Inherited from `ButtonBase`. Controls default/rounded/rectangular/oval shape. Not in defaults. |
| `Image` | **Medium** | Inherited from `ButtonBase` / `TouchComponent`. Button background image. Not in defaults, not resolved in `assetUrl`. |
| `ShowFeedback` | **Low** | Inherited from `TouchComponent`. Not in defaults. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `Text` | `""` (empty string, from ButtonBase) | `'ListPicker'` | **Medium** | `ButtonBase.Text` defaults to `""`. Tensor uses `'ListPicker'` as the display string. This means a freshly dropped component shows button text in the simulator that does not match what App Inventor would show (an empty button). |
| `BackgroundColor` | `&H00000000` (COLOR_DEFAULT — system default button appearance) | not set in defaults | **Medium** | The AI default for ButtonBase's `BackgroundColor` is `COLOR_DEFAULT` (transparent/system). Tensor omits it from the ListPicker defaults object, so `baseStyle()` outputs no background rule, which results in the outer wrapper being unstyled — acceptable visually but semantically divergent. |
| `TextColor` | `&H00000000` (COLOR_DEFAULT) | not set in defaults | **Low** | Same as BackgroundColor — AI defaults to system colour. Tensor inherits from CSS. Low impact. |
| `ItemTextColor` | `&HFFFFFFFF` (white) | missing | **High** | This should default to white text on items. Tensor renders dark text instead. |
| `ItemBackgroundColor` | `&HFF000000` (black) | missing | **High** | This should default to black item background. Tensor renders white instead — the opposite. |

---

## Events Analysis

### Supported

- `BeforePicking` — fired by `openListPicker()` in the Svelte component (on button click) and also by the Go host when `Open()` is called programmatically. No arguments. Matches AI spec.
- `AfterPicking` — fired by `pickListItem()` after the user selects an item. No arguments. Matches AI spec. `Selection` and `SelectionIndex` are updated before the event fires.

### Missing / Incorrect

| Event | Priority | Notes |
|---|---|---|
| `GotFocus` | **Low** | Inherited from `ButtonBase` — fired when the button gains keyboard/touch focus. The ListPicker `<button>` element has no `on:focus` handler (unlike `TextBox` which does have one). |
| `LostFocus` | **Low** | Inherited from `ButtonBase`. Same issue — no `on:blur` handler on the picker button. |
| `BeforePicking` return value ignored | **High** | In the AI implementation (`Picker.click()`), `BeforePicking` is a **fallible event** — if it returns `false` (via `StopBlocksExecution`), the picker does **not** open. In Tensor, `openListPicker()` fires `BeforePicking` as a plain event and always sets `pickerOpen = true` regardless of what the event handler does. This breaks blocks that use `BeforePicking` to conditionally prevent opening. |

---

## Methods Analysis

### Supported

- `Open()` — handled in `simulation_wasm.go` (line 199-206). Fires `BeforePicking` via `runEvent` and emits a `component-action` effect with `action: "open"` to open the dropdown. Matches AI spec at a functional level.

### Missing / Incorrect

| Method | Priority | Notes |
|---|---|---|
| `Open()` — `BeforePicking` guard not respected | **High** | When called programmatically via `Open()`, the Go host fires `BeforePicking` but immediately appends the `component-action` open effect regardless of whether `BeforePicking` returned/threw a stop. The guard logic is identical to the frontend issue. |

---

## Behaviour Gaps

### 1. BeforePicking Cannot Cancel Opening (Critical)

In App Inventor, `Picker.click()` calls `BeforePicking()` which dispatches the `BeforePicking` event. If any block inside the event calls `StopBlocksExecution` (or equivalent), the method returns `false` and the intent is never started — the picker stays closed. Tensor always opens the picker after firing `BeforePicking`, making it impossible to guard picker opening from blocks.

**Code reference**: `SimulationComponent.svelte` line 155-159:
```javascript
function openListPicker() {
  if (!enabled) return;
  pickerOpen = true;          // opens unconditionally
  emitEvent(node.name, 'BeforePicking');   // event fires AFTER open
}
```
The event is emitted AFTER the picker is already opened, so even if the event handler were to set a flag, the dropdown is already visible.

### 2. ItemTextColor and ItemBackgroundColor Not Applied (High)

The AI Java source initialises `itemTextColor = DEFAULT_ITEM_TEXT_COLOR` (white, `0xFFFFFFFF`) and `itemBackgroundColor = DEFAULT_ITEM_BACKGROUND_COLOR` (black, `0xFF000000`). These are passed to the `ListPickerActivity` via `Intent` extras and applied to each list row. In Tensor the defaults are simply absent — the dropdown items render with a white background and dark text, which is the opposite of the AI defaults. Any blocks that set `ItemTextColor` or `ItemBackgroundColor` will be silently ignored (the properties are not in `SIMULATION_VISUAL_PROPS`).

### 3. ShowFilterBar Has No Effect (High)

`ShowFilterBar` defaults to `false` in AI. When set to `true`, the ListPickerActivity shows a search box at the top of the list. In Tensor this property is entirely absent from defaults and `SIMULATION_VISUAL_PROPS`, so there is no search input rendered in the dropdown and no filtering behaviour.

### 4. Title Not Rendered in Dropdown UI (Medium)

`Title` is correctly stored in state (it is in `SIMULATION_VISUAL_PROPS` and the defaults object), but the dropdown menu template (lines 327-334 of `SimulationComponent.svelte`) never renders the title. In App Inventor the title appears as the activity's title bar text. In Tensor a title set by the user or blocks is silently dropped from the visual.

### 5. setElements() Does Not Adjust SelectionIndex for ListPicker (Medium)

In `simulation_wasm.go`, `setElements()` (lines 121-133) only adjusts `SelectionIndex` for the `Spinner` component type. When `Elements` changes on a `ListPicker` and the current `SelectionIndex` now exceeds the new list length, the index is not reset to 0 and `Selection` is not cleared, leaving the state inconsistent. The AI implementation's `ElementsUtil.elements()` does not automatically reset selection, but the combined setter interaction through blocks would eventually re-read the item, so this is an edge-case inconsistency rather than a strict spec violation — nevertheless the asymmetry between Spinner and ListPicker handling is a behaviour gap.

### 6. Dropdown Stays Open When Dismissed Without Selection (Low)

In App Inventor, pressing the back button on the `ListPickerActivity` dismisses the list with `RESULT_CANCELED`, which means `AfterPicking` is NOT fired and `Selection`/`SelectionIndex` are unchanged. In Tensor there is no way to dismiss the dropdown without picking an item (there is no backdrop/close button). Once `pickerOpen = true`, the only path to closing is picking an item. This means users cannot simulate "user cancelled the picker" scenarios.

### 7. Text Default Diverges from AI Spec (Medium)

`ButtonBase.Text` defaults to `""` in App Inventor. Tensor uses `'ListPicker'` as the default `Text` value. A freshly placed component will show `ListPicker` as the button label in the simulator, while in AI it would show an empty button (with just the system button styling).

### 8. BackgroundColor/TextColor defaults use COLOR_DEFAULT, not white/black (Low)

The AI `ButtonBase` defaults `BackgroundColor` to `&H00000000` (COLOR_DEFAULT — system theme button colour, not transparent). The Tensor ListPicker button is rendered with a hard-coded `background: #5b6470` CSS class (`sim-button`), ignoring the color-default convention. When a user block sets `BackgroundColor` it does apply, but the baseline appearance does not reflect the AI default.

---

## Bugs Found

### Bug 1: `openListPicker()` opens the picker BEFORE firing BeforePicking (Critical)

**File**: `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 155-159

```javascript
function openListPicker() {
  if (!enabled) return;
  pickerOpen = true;            // BUG: picker is open before event fires
  emitEvent(node.name, 'BeforePicking');
}
```

`pickerOpen` is set to `true` synchronously before the `BeforePicking` event is dispatched to the WASM runtime. This means:
1. Any UI update based on `pickerOpen` will show the dropdown before blocks run.
2. The AI spec guarantees blocks run BEFORE the picker is shown, enabling use-cases like dynamically populating `Elements` in `BeforePicking`.

The correct order is: fire `BeforePicking` → wait for block handlers to complete → then open the dropdown.

### Bug 2: `ItemTextColor` and `ItemBackgroundColor` absent from SIMULATION_VISUAL_PROPS (High)

**File**: `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`

Neither `ItemTextColor` nor `ItemBackgroundColor` appear in `SIMULATION_VISUAL_PROPS` (lines 63-100). This means even if these properties are set by blocks via `SetProperty`, the `deriveStateFromDesignerProps()` function will silently discard them, and the `coerceSimulationValue()` function won't handle them. Blocks reading these properties will receive stale or null values.

### Bug 3: `ShowFilterBar` absent from defaults, SIMULATION_VISUAL_PROPS, and coercion (High)

**File**: `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`

`ShowFilterBar` is missing from all three tracking structures. A block calling `set ListPicker1.ShowFilterBar to true` will not register the change in any meaningful way.

### Bug 4: `setElements()` skips SelectionIndex bounds check for ListPicker (Medium)

**File**: `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, lines 121-133

```go
func (h *simulationHost) setElements(componentName, componentType string, value runtime.Value) {
    h.setProperty(componentName, "Elements", value)
    if componentType != "Spinner" {
        return   // BUG: ListPicker and ListView do not get SelectionIndex clamped
    }
    // ... Spinner-only bounds check
}
```

When `Elements` is changed on a ListPicker and the current `SelectionIndex` is now out of bounds, neither `SelectionIndex` nor `Selection` is reset to 0/`""`. This can leave the state inconsistent.

### Bug 5: ListPicker button text precedence wrong (Low)

**File**: `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 327

```svelte
<button ...>{props.Text || props.Selection || 'ListPicker'}</button>
```

When `Text` is the empty string (the AI default), `props.Text` is falsy, so it falls through to `props.Selection`. This means after a user picks an item, the button label changes to show the current selection. This does NOT match the App Inventor behaviour — the `Text` property of the button is separate from `Selection` and does not change when the user picks an item. If the user sets `Text` to `""`, the button should display nothing, not the selection value.

---

## Android/App Inventor Standards Compliance

| Aspect | Status | Notes |
|---|---|---|
| ListPicker opens as a full-screen Activity | Partial | Tensor renders an inline dropdown. This is a necessary Web adaptation, but the "full screen takeover" feeling is absent. Acceptable for simulation. |
| Selection/SelectionIndex are 1-based | Correct | `selectionByIndex()` correctly offsets by +1. |
| SelectionIndex 0 = no selection | Correct | Handled in `setSelectionIndex()`. |
| Out-of-range SelectionIndex resets to 0 | Correct | `setSelectionIndex()` handles this. |
| Setting `Selection` to a value not in Elements sets SelectionIndex to 0 | Correct | `setSelection()` with `case "ListPicker"` sets `Selection` to the raw text but sets `SelectionIndex` via `selectionIndexForValue()` which returns 0 for non-matches. |
| Setting `Selection` to a value in Elements sets SelectionIndex to matching position | Correct | `selectionIndexForValue()` does a linear scan. |
| `ElementsFromString` with commas | Correct | Handled in both JS and Go, with whitespace trimming. |
| `BeforePicking` can prevent picker opening | **Not implemented** | Critical compliance gap. |
| `ItemTextColor` default is white on black background | **Not implemented** | Defaults are inverted. |
| `Open()` programmatic access | Correct | Works via `component-action` effect. |
| `Title` displayed at top of picker list | **Not implemented** | Stored but not rendered. |
| `ShowFilterBar` search functionality | **Not implemented** | Not present at all. |

---

## Summary

The Tensor ListPicker simulator correctly handles the core pick interaction flow (select item → update Selection/SelectionIndex → fire AfterPicking), the Open() method, and property synchronisation. However, it has significant gaps in visual fidelity and one critical behavioural bug.

**Top 3 action items:**

1. **Fix `BeforePicking` ordering and cancel semantics (Critical):** In `openListPicker()`, emit `BeforePicking` first and only set `pickerOpen = true` after the event resolves. The Go `CallMethod` for `Open()` has the same ordering issue. In App Inventor, `BeforePicking` returning false (or throwing StopBlocksExecution) must prevent the picker from opening — this is the primary use-case for that event.

2. **Add `ItemTextColor` and `ItemBackgroundColor` (High):** Add these to `SIMULATION_DEFAULTS` (with correct AI defaults: `ItemTextColor: '&HFFFFFFFF'`, `ItemBackgroundColor: '&HFF000000'`), add them to `SIMULATION_VISUAL_PROPS`, handle them in the ListPicker dropdown item rendering (apply inline `color` and `background` CSS to each `<button>` in `sim-picker-menu`), and add colour coercion in `coerceSimulationValue`.

3. **Add `ShowFilterBar` and render Title in the dropdown (High):** Add `ShowFilterBar: false` to `SIMULATION_DEFAULTS` and `SIMULATION_VISUAL_PROPS`. When `ShowFilterBar` is `true`, render a text input inside `sim-picker-menu` and filter the displayed items reactively. Additionally render `props.Title` as a header row inside the dropdown so users can see and test title-dependent behaviour.
