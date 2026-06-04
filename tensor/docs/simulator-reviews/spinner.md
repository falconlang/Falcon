# Spinner Simulator Review

## Overview

The Spinner component in App Inventor is a dropdown/popup picker that displays a list of string items. When the user taps it, a dialog pops up with all items listed as radio-button-style choices. The first item is auto-selected on creation; re-selecting the first item does not fire `AfterSelecting`. This review compares the AI source (`Spinner.java`) against the Tensor IDE simulator across the frontend (`SimulationComponent.svelte`), the capabilities/defaults layer (`simulation-capabilities.js`), and the Go WASM backend (`simulation_wasm.go`).

---

## Properties Analysis

### Supported
- `Elements` — stored and rendered correctly; list is updated when `ElementsFromString` changes.
- `ElementsFromString` — parsed with comma-split, trimmed, empty entries filtered out. Matches AI behaviour.
- `Selection` — read/write; setter does a linear search through elements to resolve `SelectionIndex`. Correctly sets both to empty when the value is not found in the list (Spinner-specific path in `setSelection`).
- `SelectionIndex` — read/write; 1-based; out-of-range values (< 1 or > len) set index to 0 and selection to "". Correct.
- `Prompt` — supported as the `<option disabled>` placeholder when no item is selected. Rendered in the `<select>` widget.
- `Visible` — supported via `isSimulationVisible`.
- `Enabled` — supported; disables the `<select>` element.
- `Width` / `Height` — supported via `sizeStyle()`, including -1 (preferred), -2 (fill parent), and -1000-N percent encoding.

### Missing / Unsupported

| Property | AI Default | Priority | Notes |
|---|---|---|---|
| `FontBold` | `false` | **High** | Stored on adapter in AI; ignored entirely in Tensor. Bold is never applied to the spinner text in the simulator. |
| `FontItalic` | `false` | **High** | Same as FontBold — never applied. |
| `FontSize` | `14.0 sp` | **High** | `FontSize` is not included in the `Spinner` defaults block in `simulation-capabilities.js` and is not applied to the `<select>` or picker-menu elements. The global `baseStyle()` only injects `font-size` when `props.FontSize` is truthy; since it's absent from defaults it is silently ignored. |
| `FontTypeface` | `"0"` (default) | **Medium** | Not tracked or rendered. |
| `TextColor` | `Component.COLOR_DEFAULT` | **High** | Not included in Spinner defaults. `baseStyle()` injects `color` only when `props.TextColor` is truthy — absent from defaults, so the text colour property has no effect unless explicitly set via blocks. |
| `TextAlignment` | `Component.ALIGNMENT_CENTER` (1) | **Medium** | Not in defaults and not applied to the select or picker menu. |
| `BackgroundColor` | (inherited from theme) | **Low** | Not in Spinner defaults. The component has no explicit background colour in AI but `BackgroundColor` is a standard `TouchComponent` property. |
| `WidthPercent` / `HeightPercent` | n/a | **Low** | The `-1000-N` percent encoding is handled in `sizeStyle()` correctly, but these are block-only setters and there is no separate setter path needed in the simulator — this is fine as-is. |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority | Notes |
|---|---|---|---|---|
| `FontSize` (absent) | `14.0` sp | Not present in Spinner defaults object | **High** | All other text-bearing components (Label, Button, CheckBox etc.) include `FontSize: 14` in their defaults. Spinner omits it. |
| `TextColor` (absent) | `COLOR_DEFAULT` (system color) | Not present in Spinner defaults object | **High** | Same omission pattern as FontSize. Every other text component has `TextColor` in defaults. |
| `SelectionIndex` initial | `0` | `0` | Correct | — |
| `Selection` initial | `""` | `""` | Correct | — |

---

## Events Analysis

### Supported
- `AfterSelecting(selection)` — fired correctly from `pickSpinnerItem()` with the selected string as the single argument. The `EVENT_ARGS` map correctly declares `['selection']`. The Go backend receives the event via `dispatchSimulationEvent` with `args[0]` = the selection string.

### Missing / Incorrect

| Event | Status | Priority | Notes |
|---|---|---|---|
| First-item-selected suppression | **Partially wrong** | **High** | In App Inventor, selecting the first item when the spinner is initially loaded (i.e. `oldSelectionIndex == 0` and adapter just became non-empty) does **not** fire `AfterSelecting`. Tensor's `pickSpinnerItem` always fires `AfterSelecting` regardless of whether the item was the initial auto-selection. In practice this only matters when `DisplayDropdown` is called and the user immediately picks the item that was already selected; real first-load suppression is less relevant in the simulator but the per-click path is still incorrect. |
| Re-picking the currently selected item | **Wrong** | **Medium** | In AI's `onItemSelected`, picking the same item that is already selected does fire `AfterSelecting` (because the user explicitly chose it). Tensor correctly fires it too. This is fine. |
| `TouchDown` / `TouchUp` / `LongClick` / `Click` | **Missing** | **Medium** | Spinner extends `TouchComponent`. Inherited events `TouchDown`, `TouchUp`, `LongClick`, and `Click` are not wired in the frontend or in the Go backend. These are rarely used for Spinners but are part of the AI API surface. |

---

## Methods Analysis

### Supported
- `DisplayDropdown()` — implemented in `simulation_wasm.go` (`CallMethod` → `"DisplayDropdown"` → appends a `component-action "open"` effect). The frontend watches `actions[node.name].open` and sets `pickerOpen = true`. Functionally correct.

### Missing / Incorrect

| Method | Priority | Notes |
|---|---|---|
| No additional methods are missing | — | Spinner only exposes `DisplayDropdown` as a `@SimpleFunction`. All inherited methods from `TouchComponent` (e.g. `getWidth`, `getHeight`) are not part of the blocks-callable surface and do not need simulation. |

---

## Behaviour Gaps

### 1. First-item auto-selection suppression (High)
App Inventor documents: *"Spinners are created with the first item already selected. So selecting it does not generate an AfterSelecting event."* In the Java source (`onItemSelected`), when the adapter transitions from empty to non-empty and `oldSelectionIndex == 0`, the selection is updated silently without calling `AfterSelecting`. Tensor's simulator does not implement this guard. If a `Screen.Initialize` handler sets `Elements` and the user opens the dropdown and picks the first item (which was the default selection), `AfterSelecting` fires unexpectedly.

### 2. `AfterSelecting` fires even when picker is opened via `DisplayDropdown` and closed without changing selection (Medium)
In native Android, `onItemSelected` only fires when the selected item *changes* after the dialog is shown. In Tensor's HTML `<select>` widget, the `on:change` event only fires when a new option is actually chosen, so this is mostly fine. However, in the custom `sim-picker-menu` (used when `DisplayDropdown` is called), every button click fires `AfterSelecting` — including clicking the item that was already selected (`pickSpinnerItem` is called unconditionally). This is slightly inconsistent with AI behaviour.

### 3. Prompt rendering inconsistency (Medium)
In AI, `Prompt` is the **title** of the dropdown dialog window — it appears as a heading above the list, not as a selectable item. In Tensor, `Prompt` is rendered as a `<option value="-1" disabled>` placeholder in the `<select>` element. This gives the impression that Prompt is a non-selectable first choice, but it is never shown in the custom `sim-picker-menu` that appears when `DisplayDropdown()` is invoked. The two UI paths (native `<select>` vs. custom overlay) behave differently for the same property.

### 4. Custom picker-menu does not close when clicking outside (Low)
When `DisplayDropdown()` is called, `pickerOpen` is set to `true` and a `sim-picker-menu` overlay is shown. There is no `on:click` handler on the backdrop or `document` to dismiss the overlay — it only closes when an item is explicitly clicked. This is a usability gap; in practice users who click away from the dropdown in the AI emulator would dismiss it.

### 5. `setElements` guard applies only to Spinner (correct), but the first-item auto-select sync on non-empty-to-non-empty change is missing (Medium)
In AI's `Elements(YailList)` setter, when the new list is smaller than the old list **and** the current selection index equals the old list size, the selection is clamped to the new list size. Tensor's `setElements` in `simulation_wasm.go` only clamps when `currentIndex > len(elements)` — this matches the AI logic for the "index out of range" case. However it does not handle the exact equality edge case (`currentIndex == len(oldItems) && newLen < oldLen`). This is a minor deviation.

### 6. `Visible` hide/show does not save/restore Width and Height (Low)
In `Spinner.java`, `Visible(boolean)` saves `savedWidth`/`savedHeight` on hide and restores them via `Width(savedWidth)` / `Height(savedHeight)` on show. Tensor's visibility handling simply reads `props.Width` / `props.Height` from state every render, so the save/restore semantics are implicit in the state — this is acceptable in a web simulator but is a minor fidelity gap if blocks change Width while the component is hidden (the Tensor simulator will apply the new width immediately upon re-render, which actually matches the intent better).

### 7. Spinner `<select>` vs. App Inventor dialog popup (Low/cosmetic)
App Inventor renders Spinner as a `performClick()` that opens a full dialog with a radio list. Tensor renders it as a native HTML `<select>` element which on desktop shows a native OS dropdown. This is an intentional simplification for web simulation, but users familiar with the AI popup behaviour may be confused.

---

## Bugs Found

### Bug 1 — `AfterSelecting` always fires; no guard for initial first-item selection (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, line 143–153

```js
function pickSpinnerItem(index) {
  const { selection, selectionIndex } = selectionByIndex(index);
  pickerOpen = false;
  emitInteraction(
    [...],
    { component: node.name, event: 'AfterSelecting', args: [selection] },
  );
}
```

There is no check for whether `selectionIndex === Number(props.SelectionIndex)` (i.e. user re-picked the already-selected item after the spinner was freshly populated). According to AI spec, the initial selection should not fire `AfterSelecting`. There should be a guard: if this is the first selection after initial population (detectable by `props.SelectionIndex === 0` before the pick), do not emit the event.

### Bug 2 — `FontSize`, `TextColor`, `FontBold`, `FontItalic` absent from Spinner defaults cause silent no-op (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 49

```js
Spinner: { ...COMMON_VISIBLE_PROPS, Elements: [], ElementsFromString: '', Selection: '', SelectionIndex: 0, Prompt: '' },
```

`FontSize`, `TextColor`, `FontBold`, `FontItalic`, and `TextAlignment` are missing. Since `baseStyle()` only applies `color` and `font-size` when the prop is truthy, a user who sets `TextColor` via the Designer or blocks will see the value set in state but it will correctly propagate through `baseStyle()` — **however** the default values specified by AI (`FONT_DEFAULT_SIZE = 14`, `COLOR_DEFAULT`) are never applied, so the spinner always renders in the browser's native `<select>` styling regardless of what the Designer palette shows.

### Bug 3 — `spinnerChange` reads `e.currentTarget.value` as the option's `index` value, but the Prompt option has `value="-1"` (Medium)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 308–315 and 138–140

```html
{#if props.Prompt}
  <option value="-1" disabled selected={!props.Selection}>{props.Prompt}</option>
{/if}
{#each elements as item, index}
  <option value={index} ...>{item}</option>
{/each}
```

```js
function spinnerChange(e) {
  const index = Number(e.currentTarget.value);
  pickSpinnerItem(index);
}
```

If somehow the prompt option is not `disabled` (e.g. browser quirk) and the user selects it, `index = -1` is passed to `pickSpinnerItem(-1)`. `selectionByIndex(-1)` would return `elements[-1]` which is `undefined`, yielding `selection = ''` and `selectionIndex = 0`. While the prompt option is marked `disabled`, defensively `pickSpinnerItem` should guard against negative indices.

### Bug 4 — `setSelection` in Go backend does case-sensitive exact match only (Low)
**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, lines 168–175

```go
func selectionIndexForValue(text string, elements []runtime.Value) int {
    for index, item := range elements {
        if item.AsStr() == text {
            return index + 1
        }
    }
    return 0
}
```

App Inventor's `ElementsUtil.setSelectedIndexFromValue` also performs a case-sensitive exact-string match, so this is correct. No bug here — noted for completeness.

### Bug 5 — `Prompt` is not rendered in the custom `sim-picker-menu` (Medium)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/SimulationComponent.svelte`, lines 316–322

```html
{#if pickerOpen}
  <div class="sim-picker-menu">
    {#each elements as item, index}
      <button ...>{item}</button>
    {/each}
  </div>
{/if}
```

When `DisplayDropdown()` triggers `pickerOpen = true`, the overlay does not show the `Prompt` as a dialog title, despite it being visible in the `<select>` as a placeholder option. This creates inconsistent behaviour between the native select and the programmatic dropdown path.

---

## Android/App Inventor Standards Compliance

| Aspect | Status | Notes |
|---|---|---|
| 1-based SelectionIndex | Correct | Both frontend and Go backend use 1-based indexing. |
| Empty list sets SelectionIndex to 0 | Correct | `setElements` handles this. |
| SelectionIndex out of range sets to 0 | Correct | `setSelectionIndex` range-checks and zeroes. |
| `ElementsFromString` comma-split + trim | Correct | Matches AI's `ElementsUtil.elementsFromString`. |
| `DisplayDropdown` opens dropdown | Correct | Effect-based communication works end-to-end. |
| `AfterSelecting` arg is the selected string | Correct | Event arg is `selection` string. |
| First-item no-fire rule | **Non-compliant** | Not implemented; see Bug 1. |
| Font/text styling on spinner text | **Non-compliant** | FontBold, FontItalic, FontSize, TextColor, TextAlignment not applied. |
| Prompt as dialog title (not list item) | **Partially non-compliant** | Rendered as a `<select>` placeholder option rather than a dialog heading. |

---

## Summary

The Tensor IDE Spinner simulation covers the core interactive flow well: elements populate, selecting an item updates `Selection` and `SelectionIndex` bidirectionally, `AfterSelecting` is dispatched with the right argument, `DisplayDropdown` works, and out-of-range index handling is correct.

**Top 3 action items:**

1. **Add missing text-styling defaults and apply them** — Add `FontSize: 14`, `TextColor: '&HFF000000'` (or use COLOR_DEFAULT), `FontBold: false`, `FontItalic: false`, `TextAlignment: 1` to the Spinner entry in `SIMULATION_DEFAULTS` and propagate them to the `<select>` and `sim-picker-menu` elements via inline style (e.g. `font-weight`, `font-style`, `text-align`). This fixes Bugs 1-related rendering gaps and aligns with every other text component in the simulator. **Priority: High.**

2. **Implement the initial-selection no-fire guard for `AfterSelecting`** — In `pickSpinnerItem`, before emitting the `AfterSelecting` interaction, check whether the item being picked was the already-selected item at the time the picker was opened (specifically when `props.SelectionIndex === 0` before first pick, meaning the spinner was just populated). This matches the documented App Inventor behaviour. **Priority: High.**

3. **Unify Prompt display across both UI paths** — Either render a non-clickable header row with `props.Prompt` at the top of `sim-picker-menu` (matching the AI dialog-title semantic), or remove the Prompt `<option>` from the native `<select>` and solely rely on the custom overlay. This eliminates the inconsistency between the `<select>` path and the `DisplayDropdown` path. **Priority: Medium.**
