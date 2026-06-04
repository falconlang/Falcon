# TableArrangement Simulator Implementation Plan

## Overview
`TableArrangement` is an App Inventor LAYOUT container: "A formatting element in which to place
components that should be displayed in tabular form." It exposes a fixed grid of `Rows` x `Columns`
cells (defaults 2 x 2) and holds visible child components, one per cell. Unlike the linear
`HorizontalArrangement` / `VerticalArrangement`, the table has no `AlignHorizontal` / `AlignVertical`,
no `BackgroundColor`, and no `Image` property in its designer surface — it is a pure geometry
container.

- **Category:** LAYOUT
- **Visible:** Yes (it is a visible container; `nonVisible: false`).
- **Container relationship:** It is a child of any container (`Screen`/`Form`, the
  arrangements, scroll arrangements, `AbsoluteArrangement`) and it is a *parent* that holds visible
  children in a grid. In App Inventor each child occupies a specific `(Row, Column)` cell; a cell can
  hold at most one component; the table does not scroll and does not wrap.

## Feasibility Verdict
**Partially feasible (in a browser simulator).**

The grid container itself is trivially feasible with CSS Grid (`display: grid` +
`grid-template-columns: repeat(N, auto)` / `grid-template-rows: repeat(M, auto)`), and the
read/write of `Width`/`Height`/`Visible` rides the generic store path that every other component
already uses. The one real gap is **exact cell placement**:

- In App Inventor's `.scm`/designer model, each child carries a `Row` and `Column` cell coordinate,
  and cells may be sparse (a 2x2 table can hold a single child at cell (1,1) with the other three
  cells empty). The TableArrangement's own `Row`/`Column` block properties (both `invisible`,
  read-only) are the per-child placement getters.
- Tensor's design schema represents children as a **flat, ordered list** with no per-child
  `Row`/`Column` cell metadata (see `parseComponent` in `design-schema-tree.js`: a node is
  `{ type, name, props, children }` — no cell coordinates are parsed or stored). The simulator
  therefore cannot know which cell a given child was assigned to.

**Realistic approximation:** render an N-column CSS grid and fill cells with children in
**document order** (row-major: child 0 -> cell (0,0), child 1 -> cell (0,1), ...). This reproduces
the *tabular look* (aligned columns/rows, grid sizing behavior) for the common case where the AIA
author placed children densely from the top-left. It will diverge from the device whenever the
original layout left gaps or placed children out of document order — there is no data in the schema
to recover that. This is an honest, accepted-fidelity limitation, not a blocker; the visual result
is correct for the typical "fill the grid" usage and far better than the current dashed
"unsupported" placeholder.

## Properties
| Property | AI default | Visual/Behavioral | How to simulate | Priority |
| --- | --- | --- | --- | --- |
| `Columns` | `2` | Visual — number of grid columns | `grid-template-columns: repeat(N, max-content)` (or `auto`); clamp to >= 1 | High |
| `Rows` | `2` | Visual — number of grid rows | `grid-template-rows: repeat(M, max-content)`; clamp to >= 1 | High |
| `Visible` | `True` | Visual | Generic boolean; hides the grid (`display: none` via existing visible handling) | High |
| `Width` | (empty -> Automatic, `-1`) | Visual — pixels / `-1` auto / `-2` fill / `<=-1000` percent | Reuse `sizeStyle('Width')` exactly as the arrangements do | High |
| `Height` | (empty -> Automatic, `-1`) | Visual | Reuse `sizeStyle('Height')` | High |
| `Left` | `` (empty) | Behavioral — only meaningful inside `AbsoluteArrangement` | Reuse existing `positionStyle()` (already keyed on `parentType === 'AbsoluteArrangement'`) | Medium |
| `Top` | `` (empty) | Behavioral — `AbsoluteArrangement` offset | Reuse `positionStyle()` | Medium |
| `WidthPercent` | (write-only block prop) | Behavioral | Generic `SetProperty` store write; the JS `sizeStyle` percent path (`<= -1000`) is the rendering side — see caveat | Low |
| `HeightPercent` | (write-only block prop) | Behavioral | As `WidthPercent` | Low |
| `Column` / `Columns` (block getters) | — | Read-only (`invisible`) | Return stored `Columns`; `Column` getter has no per-child source — return stored `Columns` value, see caveat | Low |
| `Row` / `Rows` (block getters) | — | Read-only (`invisible`) | Return stored `Rows` | Low |

Notes:
- `TableArrangement` has **no** `BackgroundColor`, `Image`, `AlignHorizontal`, or `AlignVertical` in
  its designer surface — do **not** add those. `baseStyle()`'s alignment/background-image rules are
  no-ops for this type anyway (`alignmentStyle()` only triggers for the linear layouts), so the safer
  choice is to render with `sizeStyle` + `positionStyle` directly rather than full `baseStyle()`.
- `Left`/`Top`/`Width`/`Height` are already in `SIMULATION_VISUAL_PROPS`; only `Columns`/`Rows`
  (and optionally the percent props) are new.

## Events
| Event | Args | How/when the simulator fires it |
| --- | --- | --- |
| (none) | — | The spec lists `"events": []`. TableArrangement has no events; nothing to wire. |

## Methods
| Method | Signature | Simulated behavior (or Unsupported + why) |
| --- | --- | --- |
| (none) | — | The spec lists `"methods": []`. No `callTableArrangementMethod` handler is needed. |

## Implementation Plan

### simulation-capabilities.js
- **`SIMULATION_SUPPORTED_TYPES`:** add `'TableArrangement'`.
- **`buildSimulationDefaults()`** — add a defaults block (note: NOT spread from `ARRANGEMENT_PROPS`,
  because that injects `AlignHorizontal`/`AlignVertical`/`BackgroundColor`/`Image` which this type
  does not have):
  ```js
  TableArrangement: {
    Visible: true,
    Width: -1,
    Height: -1,
    Columns: 2,
    Rows: 2,
  },
  ```
- **`SIMULATION_VISUAL_PROPS`:** add `'Columns'` and `'Rows'`. (`Visible`, `Width`, `Height`,
  `Left`, `Top` are already present.)
- **`isNumericProp`:** add `'Columns'` and `'Rows'` so they coerce to numbers. (`Visible` already
  handled by `isBooleanProp`.)
- **`coerceSimulationValue`:** no special case needed — the `isNumericProp` branch already returns
  `coerceNumber(value)` for `Columns`/`Rows`.
- **`deriveStateFromDesignerProps`:** no derived state needed; the generic prop loop is sufficient.

### SimulationComponent.svelte
Add a render branch alongside the other arrangement branches (after `AbsoluteArrangement`,
before `Label`). Sketch:
```svelte
{:else if node.type === 'TableArrangement'}
  <div
    class="sim-table"
    class:sim-unsupported={unsupportedHere}
    style={`${containerStyle()} display: grid; grid-template-columns: repeat(${Math.max(1, numberOr(props.Columns, 2))}, max-content); grid-template-rows: repeat(${Math.max(1, numberOr(props.Rows, 2))}, max-content); align-items: start; justify-items: start;`}
    data-sim-component={node.name}
  >
    {#each node.children || [] as child (child.pathId || child.name)}
      <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
    {/each}
  </div>
{/if}
```
- **Reuse:** `containerStyle()` (already defined: `sizeStyle('Width') + sizeStyle('Height') +
  positionStyle()`) — this is exactly the right helper because it omits the background/alignment
  rules TableArrangement does not have, while keeping size + `AbsoluteArrangement` positioning.
- **Child rendering:** identical `<svelte:self>` recursion used by every container branch; children
  flow into grid cells row-major in document order (the accepted approximation).
- **DOM events:** none (container has no events).
- **CSS:** add a `.sim-table` rule mirroring `.sim-arrangement` minimal styling (no `display: flex`;
  the inline `display: grid` drives layout). `numberOr` is already an available helper.
- **Empty-table floor:** optionally mirror `emptyArrangementStyle()`'s 100px floor for an
  Automatic-sized empty table to match App Inventor's empty-arrangement minimum; low priority.

### SimulateOverlay.svelte
None. No dialogs, toasts, or runtime effects.

### simulation_wasm.go
Generic store path suffices for the read-write properties:
- **`GetProperty`:** generic store read already returns stored `Width`/`Height`/`Visible`/`Left`/
  `Top`/`Columns`/`Rows`. No new case required for the common read-write set.
- **`SetProperty`:** falls through to `h.setProperty` (generic) for `Width`/`Height`/`Visible`/
  `Left`/`Top`/`Columns`/`Rows` — these patch state and re-render the grid live.
- **`WidthPercent`/`HeightPercent` (write-only):** optional. If precise device parity is desired,
  add a small case under a `case "TableArrangement":` (or shared with arrangements) that converts the
  percentage to the `<= -1000` sentinel the JS `sizeStyle` understands (e.g. store `Width` as
  `-1000 - pct`). Otherwise leave to the generic path (it will store the raw percent and the renderer
  will treat it as a pixel value — incorrect, so prefer the conversion if these are exercised).
- **`Column`/`Row` block getters (`invisible`):** these are designer-time per-child getters with no
  runtime source in the simulator; safe to leave on the generic path (returns stored `Columns`/`Rows`
  or `NullVal`). No `callTableArrangementMethod` is needed (no methods).
- No `h.Unsupported(...)` calls are required; nothing in the surface is genuinely unsupported at the
  host level.

### design-schema-tree.js
Containment is **already handled**. `canContainDesignComponent` derives container types from
`categoryString === 'LAYOUT'` (`containerTypes()`), so `TableArrangement` is automatically a valid
parent for visible children and a valid child of other containers — no edit needed.
`designTreeToInitialState` will merge `SIMULATION_DEFAULTS.TableArrangement` with derived props once
the defaults block exists. `unsupportedSimulationComponents` will stop flagging it once it is in
`SIMULATION_SUPPORTED_TYPES`. No code change in this file.

## Dependencies & Ordering
- **External libraries:** none. CSS Grid is native.
- **Prerequisite components:** none. The arrangement branches it imitates already exist. It composes
  with any already-supported child types; no other component must land first.

## Web-Platform Limitations & Fidelity Caveats
- **No per-cell placement data (primary caveat):** Tensor's schema has no `Row`/`Column` cell
  coordinates per child, so children fill cells in document order (row-major). Sparse/out-of-order
  device layouts will not be reproduced. The `Column`/`Row` block getters cannot return a meaningful
  per-child cell index.
- **Cell sizing:** App Inventor sizes each column to its widest cell and each row to its tallest, with
  Automatic-sized children. `grid-template-*: repeat(N, max-content)` approximates this; exact pixel
  metrics (Android cell padding, child Automatic measurement) will differ slightly from device.
- **Overflow:** TableArrangement does not scroll on device; if children exceed `Rows x Columns` cells,
  device behavior is undefined/clipped. The CSS grid will instead add implicit rows. Optionally cap
  rendered children at `Rows * Columns` to match the fixed-grid contract; otherwise extra children
  spill into implicit rows (a benign divergence).
- **WidthPercent/HeightPercent:** require the percent->sentinel conversion in Go to render correctly;
  without it the renderer treats the value as pixels.

## Effort Estimate
**S** — one supported-type entry, a small defaults block, two new visual/numeric prop names, one
straightforward CSS-Grid render branch, and zero Go/overlay work beyond the generic store path; the
only nuance is documenting the document-order cell-fill approximation and (optionally) the percent
conversion.
