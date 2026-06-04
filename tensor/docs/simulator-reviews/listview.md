# ListView Simulator Review

## Overview

ListView is a scrollable list component that displays items as tappable rows. It supports simple string lists as well as richer multi-column / image layouts via `ListViewLayout` and `ListData`. Users interact by tapping a row which fires `AfterPicking` and updates `Selection` / `SelectionIndex`.

The Tensor simulator renders a basic scrollable button-list, handles `AfterPicking`, `Selection`, `SelectionIndex`, `Elements`, `ElementsFromString`, and a small set of Go-side list-mutation methods (`AddItem`, `AddItemAtIndex`, `RemoveItemAtIndex`). Several visual properties, events, and methods present in the App Inventor (AI) specification are absent or incorrectly defaulted.

---

## Properties Analysis

### Supported
- `Visible` — read/rendered correctly
- `Width` / `Height` — size constants decoded; default `Width=-1` (fill-parent), `Height=-2` (wrap) - note: AI spec uses `-1` as fill-parent for Width which matches
- `Elements` — list rendered as clickable rows; normalised from array or comma-string
- `ElementsFromString` — parsed on both frontend and Go host
- `Selection` — set after picking, tracked in Go host
- `SelectionIndex` — set after picking, tracked in Go host; bounds checking present
- `TextColor` — applied via `baseStyle()` from `colorValue()`
- `FontSize` — applied via `baseStyle()`
- `BackgroundColor` — applied via `baseStyle()`

### Missing / Unsupported

| Property | AI Default | Priority |
|---|---|---|
| `ShowFilterBar` (search bar) | `false` | **High** — common feature, not rendered at all |
| `HintText` (filter bar hint) | `"Search list..."` | Medium — depends on ShowFilterBar |
| `SelectionColor` | `COLOR_LTGRAY` (`&HFFD3D3D3`) | **High** — selected row highlight colour is hardcoded to `#e8f0fe` CSS |
| `TextColorDetail` | `COLOR_WHITE` | Medium — detail text not rendered |
| `FontSizeDetail` | `14.0` (FONT_DEFAULT_SIZE) | Medium — detail text not rendered |
| `FontTypeface` | `"0"` (default) | Low — typeface not applied |
| `FontTypefaceDetail` | `"0"` (default) | Low — detail typeface not applied |
| `ListViewLayout` | `0` (LISTVIEW_LAYOUT_SINGLE_TEXT) | **High** — multi-column and image layouts not rendered |
| `ListData` | `""` | **High** — designer data for rich layouts ignored |
| `Orientation` | `1` (vertical) | Medium — horizontal swipe layout unsupported |
| `ImageWidth` | `200` | Medium — image columns not rendered |
| `ImageHeight` | `200` | Medium — image columns not rendered |
| `ElementColor` | `COLOR_NONE` | Medium — per-element background colour not applied |
| `DividerColor` | `COLOR_WHITE` | Low — dividers not rendered |
| `DividerThickness` | `0` | Low — dividers not rendered |
| `ElementMarginsWidth` | `0` | Low — element margin not applied |
| `ElementCornerRadius` | `0` | Low — rounded corners not applied |
| `BounceEdgeEffect` | `false` | Low — physics scroll effect; web limitation |
| `MultiSelect` | `false` | Medium — commented-out in AI source but may be enabled; not modelled |

### Wrong Defaults or Types

| Property | AI Default | Tensor Default | Priority |
|---|---|---|---|
| `BackgroundColor` | `COLOR_BLACK` (`&HFF000000`) | Not set (falls through to `transparent`) | **High** — list background should default to black per AI spec |
| `TextColor` | `COLOR_WHITE` (`&HFFFFFFFF`) | `'&HFF222222'` (dark grey) | **High** — text is nearly invisible on the black background in AI; Tensor inverts this |
| `FontSize` | `22` (DEFAULT_TEXT_SIZE) | `14` | Medium — noticeably smaller default |
| `Height` | `LENGTH_FILL_PARENT` (-1 mapped to fill) | `-2` (wrap) | Medium — AI forces fill-parent height; Tensor uses wrap-content |
| `SelectionIndex` | `0` | `0` | OK |
| `Selection` | `""` | `""` | OK |

---

## Events Analysis

### Supported
- `AfterPicking` — fired on row click via `listViewPick()` with `args: []`

### Missing / Incorrect

| Event | Args in AI | Tensor | Priority |
|---|---|---|---|
| `AfterPicking` args check | none (zero-argument event) | `args: []` — correct | OK |
| No event for `SelectionDetailText` change | n/a | n/a | n/a |

The AI specification defines only `AfterPicking` as a `@SimpleEvent` on ListView. No events are missing in terms of the public API. However:

- **`AfterPicking` args** — AI fires with zero arguments; Tensor correctly passes `args: []`. No issue.
- **`EVENT_ARGS` registry** (`simulation-capabilities.js`) does not have a `ListView` entry. This is not an error since `AfterPicking` takes no parameters, but the omission is inconsistent with the registry pattern. (Low)

---

## Methods Analysis

### Supported
- `AddItem(mainText, detailText, imageName)` — partially; Go host only appends `args[0]` (mainText), `detailText` and `imageName` are silently dropped (see below)
- `AddItemAtIndex(index, mainText, detailText, imageName)` — partially; same issue
- `RemoveItemAtIndex(index)` — correctly removes by 1-based index
- `Refresh()` — no-op (acceptable, adapter is always up-to-date)

### Missing / Incorrect

| Method | AI Signature | Tensor Status | Priority |
|---|---|---|---|
| `AddItem` | `(mainText, detailText, imageName)` | Only `args[0]` used; 2nd and 3rd args ignored | **High** — users building multi-column lists via `AddItem` get only the main text |
| `AddItemAtIndex` | `(index, mainText, detailText, imageName)` | Only first two args used (`args[0]` = index, `args[1]` = mainText); detail and image dropped | **High** — same as above |
| `AddItems(itemsList)` | Accepts a `List<Object>` | Not implemented; falls through to `Unsupported` | **High** — batch-add method missing entirely |
| `AddItemsAtIndex(index, itemsList)` | `(index, YailList)` | Not implemented | High |
| `CreateElement(mainText, detailText, imageName)` | Returns a dictionary | Not implemented | Medium — needed for constructing rich elements before adding |
| `GetMainText(element)` | Returns String | Not implemented | Medium |
| `GetDetailText(element)` | Returns String | Not implemented | Medium |
| `GetImageName(element)` | Returns String | Not implemented | Medium |

---

## Behaviour Gaps

### 1. Background Color Default Inversion (Critical)
AI initialises `BackgroundColor` to `COLOR_BLACK` (`&HFF000000`) and `TextColor` to `COLOR_WHITE` (`&HFFFFFFFF`). Tensor defaults `TextColor` to `'&HFF222222'` (near-black text) but does not set a background colour at all (falls through to CSS `transparent`/white). On a real device this produces nearly invisible white text on black; in the simulator it is near-black text on white — an opposite visual. Any app that relies on the default dark-list look will appear completely wrong.

### 2. SelectionColor Hardcoded in CSS
The selected row background is hardcoded to `#e8f0fe` (a light blue). AI's default `SelectionColor` is `COLOR_LTGRAY` (`&HFFD3D3D3`). Blocks that call `set ListView.SelectionColor to ...` have no effect in the simulator because the property is neither tracked nor applied to the `selected` CSS class.

### 3. Rich ListView Layouts Not Rendered
`ListViewLayout` constants `TWO_TEXT`, `TWO_TEXT_LINEAR`, `IMAGE_SINGLE_TEXT`, `IMAGE_TWO_TEXT`, `IMAGE_TOP_TWO_TEXT` are completely ignored. The frontend always renders a flat list of plain text strings regardless of the chosen layout. Any `DetailText` or `Image` fields in YailDictionary items are silently discarded during `normalizeElements()` (which calls `String(item)`, losing the dictionary structure).

### 4. `ListData` Designer Property Ignored
The designer-facing `ListData` JSON property (used to pre-populate rich rows via the Properties panel) is not listed in `SIMULATION_VISUAL_PROPS` and is not processed by `deriveStateFromDesignerProps`. Apps that use the `ListData` property to populate their ListView at design time will show an empty list in the simulator.

### 5. ShowFilterBar Not Rendered
The search/filter text box at the top of the ListView is never rendered. Blocks that set `ShowFilterBar` to `true` produce no visible UI change.

### 6. `AddItem` Drops detailText and imageName (Bug — see Bugs section)

### 7. `setElements` Only Guards `Spinner`, Not `ListView`
In `simulation_wasm.go` `setElements()` has special out-of-bounds SelectionIndex fixup logic but wraps it in `if componentType != "Spinner" { return }`. This means the SelectionIndex clamp is **only** applied to Spinner. For ListView, after a batch element replacement via `AddItems` or direct `Elements` assignment, `SelectionIndex` is not automatically reset to `0` if the current selection is now out of range. AI's `updateAdapterData()` always calls `SelectionIndex(0)` after any data change. The Go host's `setElements` should do the same for ListView.

### 8. `selectionByIndex` Assumes Flat String Elements
`selectionByIndex` in the frontend calls `elements[index]` where `elements` comes from `normalizeElements()`. For rich dictionary items (two-text / image layouts), the main text would need to be extracted from the `Text1` key; instead the entire serialised object string is used. This is consistent with the missing multi-layout support, but it means that even if `ListData` were populated, selection text would be wrong.

### 9. Orientation Property Ignored
`Orientation` (vertical vs horizontal scrolling) is not tracked in state or applied to the `sim-listview` CSS class. Horizontal ListView (carousel) is never rendered.

### 10. Height Default Mismatch
AI overrides `Height(LENGTH_PREFERRED)` to `LENGTH_FILL_PARENT` in its `Height()` setter, meaning ListView always fills its parent vertically by default. Tensor sets `Height: -2` (wrap-content/auto), so a ListView with no explicit height will be much shorter than on a real device.

---

## Bugs Found

### Bug 1 — `AddItem` Ignores detailText and imageName (Critical)
**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, lines 258–260

```go
case "AddItem":
    if len(args) >= 1 {
        elements = append(elements, args[0])
```

AI signature: `AddItem(String mainText, String detailText, String imageName)`. The implementation only uses `args[0]` and treats the internal list as a plain string list. For single-text layouts this happens to be correct, but for any multi-column layout the detail and image fields are lost. The AI implementation checks the existing item type and wraps in a `YailDictionary` when appropriate; Tensor does not.

### Bug 2 — `AddItemAtIndex` Has Off-by-One on Argument Position (Critical)
**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, lines 263–272

```go
case "AddItemAtIndex":
    if len(args) >= 2 {
        index := int(args[0].AsNum()) - 1
        ...
        elements = append(elements[:index], append([]runtime.Value{args[1]}, elements[index:]...)...)
```

AI signature: `AddItemAtIndex(int index, String mainText, String detailText, String imageName)` — 4 arguments. Tensor uses `args[0]` as index and `args[1]` as mainText, which is correct for 1-based index conversion, **but** `detailText` (`args[2]`) and `imageName` (`args[3]`) are dropped. Same structural bug as Bug 1.

### Bug 3 — `setElements` Does Not Reset SelectionIndex for ListView (High)
**File:** `/home/kumaraswamy/Documents/falcon/lang/simulation_wasm.go`, lines 121–133

```go
func (h *simulationHost) setElements(componentName, componentType string, value runtime.Value) {
    h.setProperty(componentName, "Elements", value)
    if componentType != "Spinner" {
        return   // <-- ListView exits here without bounds-checking SelectionIndex
    }
    ...
```

After `AddItem`, `RemoveItemAtIndex`, or direct `Elements` assignment the SelectionIndex is never reset, so it can silently point to a non-existent row.

### Bug 4 — `ListData` Not in `SIMULATION_VISUAL_PROPS` (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 63–100

`ListData` is absent from `SIMULATION_VISUAL_PROPS`, so `deriveStateFromDesignerProps` never propagates it to simulation state. Apps that configure list content at design-time via `ListData` see an empty list.

### Bug 5 — Wrong `TextColor` Default (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 53

```js
ListView: { ..., TextColor: '&HFF222222', ... }
```

AI default is `Component.DEFAULT_VALUE_COLOR_WHITE` = `&HFFFFFFFF`. The simulator uses a dark grey that is only legible because the simulator background is also light (not the correct black). This is the wrong default on both counts.

### Bug 6 — `BackgroundColor` Missing from Defaults (High)
**File:** `/home/kumaraswamy/Documents/falcon/tensor/src/lib/simulation-capabilities.js`, line 53

`BackgroundColor` is not listed in the `ListView` defaults object. AI default is `COLOR_BLACK`. Without this key, `baseStyle()` emits no `background:` rule and the browser renders a transparent/white background.

---

## Android/App Inventor Standards Compliance

| Aspect | Status |
|---|---|
| 1-based SelectionIndex | Correct — `selectionByIndex(index)` returns `index + 1` |
| SelectionIndex=0 means no selection | Correct |
| `AfterPicking` fired after selection changes | Correct |
| `HEIGHT_PREFERRED` mapped to fill-parent | Not applied — Tensor uses wrap-content default |
| Color format `&HaaRRGGBB` | Parsed correctly by `colorValue()` (alpha stripped, uses RGB portion) |
| `ElementsFromString` comma-split | Correct |
| `Selection` setter searches elements list | Correct in Go host |
| `SelectionIndex` out-of-range → index=0, selection="" | Correct in Go host |
| `RemoveItemAtIndex` range check | Correct (silent no-op if out of range) |
| `Refresh()` deprecated no-op | Correctly treated as no-op |
| Rich layout rendering (multi-column, images) | Not implemented |
| Filter bar (ShowFilterBar) | Not implemented |
| Horizontal orientation | Not implemented |

---

## Summary

The simulator provides a working baseline for simple string-only ListView usage: items display, clicking a row fires `AfterPicking` and updates `Selection`/`SelectionIndex`, and basic list mutation methods work. However significant gaps remain.

**Total issues found: 22** (6 bugs + 16 missing/wrong properties/methods/behaviours)

### Top 3 Action Items

1. **Fix default colors** — Set `BackgroundColor: '&HFF000000'` and `TextColor: '&HFFFFFFFF'` in `SIMULATION_DEFAULTS.ListView`. This is a one-line fix per property and makes the simulator visually match a real device for the most common case. Without it every ListView looks wrong by default. (**Critical**)

2. **Fix `AddItem` / `AddItemAtIndex` to handle detailText and imageName, and implement `AddItems`** — In `callListViewMethod` wrap new items as a dictionary when the list already contains dictionaries (matching the AI logic). Also add the `AddItems` case. This unblocks a large class of common apps that build lists dynamically. (**High**)

3. **Add `ListData` to `SIMULATION_VISUAL_PROPS` and parse it** — `ListData` is the primary designer-time mechanism for populating multi-column lists. Adding it to the visual props set and converting its JSON array to `Elements` in `deriveStateFromDesignerProps` would give basic read-only display of rich list content even before full layout rendering is implemented. (**High**)
