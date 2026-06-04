# Chart Simulator Implementation Plan

## Overview

`Chart` is a visible component (categoryString `CHARTS`, `nonVisible = false`). Per the spec helpString it is "A component that allows visualizing data." In App Inventor the `Chart` is a drawing surface (backed by the MPAndroidChart library) that renders one or more data series — supplied by child `ChartData2D` components — as a line, scatter, area, bar, or pie chart. It owns the axes, grid, legend, background, and value formatting; the actual data points live on its children.

Container relationship: `Chart` is a **container** and the **parent** of `ChartData2D` (a series of `(x, y)` points) and `Trendline` (a best-fit line computed over a `ChartData2D`). Both children are themselves `CHARTS` components with `nonVisible = false` but they have **no independent visual** — they only exist as data attached to a parent `Chart`. `design-schema-tree.js` already encodes this containment (`CHART_CHILD_TYPES = {ChartData2D, Trendline}`, `canContainDesignComponent` returns `parent === 'Chart'`).

The interactive surface of `Chart` itself is small: one event (`EntryClick`), five axis/domain methods, and a set of mostly-cosmetic properties. The data-bearing surface (AddEntry, ImportFrom*, etc.) lives on `ChartData2D`, and the regression math lives on `Trendline`.

## Feasibility Verdict

**Partially feasible.**

The core visualization is very feasible in a browser: `<canvas>` and inline SVG both render line/scatter/area/bar/pie charts well, and the property surface (axes color, background, grid, legend, type, domain/range) maps cleanly to drawing parameters. `EntryClick` maps to a click hit-test against rendered points. The genuinely hard parts are not the chart drawing — they are the **data-source plumbing on `ChartData2D`** and the **regression engine on `Trendline`**, plus the fact that fidelity to MPAndroidChart's exact rendering is unattainable.

What is feasible (realistic approximation):
- Drawing line / scatter / area / bar / pie from in-memory `(x, y)` pairs supplied via `ElementsFromPairs` (designer) and `AddEntry` / `ImportFromList` / `Clear` / `RemoveEntry` (runtime). This is straightforward `<canvas>` or SVG work.
- `Type`, `BackgroundColor`, `GridEnabled`, `LegendEnabled`, `AxesTextColor`, `Description`, `Labels` / `LabelsFromString`, `XFromZero` / `YFromZero`, `PieRadius`, per-series `Color` / `Label` / `PointShape` / `LineType`.
- `SetDomain` / `SetRange` / `ExtendDomainToInclude` / `ExtendRangeToInclude` / `ResetAxes` as axis-bound state that the renderer respects.
- `EntryClick(series, x, y)` via a click hit-test in the renderer, dispatched through `emitEvent`.
- A linear `Trendline` (and, with effort, quadratic / exponential / logarithmic / power) computed in JS via least-squares.

What is NOT feasible / must be stubbed (web-platform limitations):
- **External data sources on `ChartData2D`** — `Source`, `ImportFromCloudDB`, `ImportFromWeb`, `ImportFromSpreadsheet`, `ImportFromDataFile`, `DataFileXColumn`/`DataFileYColumn`, `SpreadsheetXColumn`/`SpreadsheetYColumn`, `WebXColumn`/`WebYColumn`, `DataSourceKey`, `ChangeDataSource`/`RemoveDataSource`. These bind a series to *other components* (CloudDB, Web, Spreadsheet, DataFile, sensors) that themselves are not implemented in the simulator and, in the browser, have no live device backing. `ImportFromTinyDB` is the one exception that could work later (TinyDB *is* implemented). All others must call `h.Unsupported("method", ...)`.
- **MPAndroidChart pixel fidelity** — gesture-driven pan/zoom, exact tick algorithms, curved/stepped `LineType` easing, anti-aliasing, animation. The approximation draws clean axes/lines but will not match the native renderer exactly.
- **`HighlightDataPoints`, `DrawLineOfBestFit` (deprecated)** — feasible-ish but low value; can be stubbed initially.

Because the headline feature (draw a chart from data set in the designer or via `AddEntry`) is fully achievable and only the cross-component data-source integrations are impossible, the honest verdict is **partial**: ship the in-memory chart, stub the external imports with `Unsupported`.

## Properties

Chart (parent) properties:

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Type` | `0` (line) | Visual | Selects the renderer mode: 0=line, 1=scatter, 2=area, 3=bar, 4=pie (ChartTypeEnum order). Drives the draw branch. Note spec marks it `read-only` at runtime (set in designer only). | **High** |
| `BackgroundColor` | `&H00000000` (transparent) | Visual | `colorValue(props.BackgroundColor, 'transparent')` as the plot-area background. | **High** |
| `GridEnabled` | `True` | Visual | Toggles drawing of grid lines behind the series. | **High** |
| `LegendEnabled` | `True` | Visual | Toggles a legend listing each series' `Label` + `Color`. | **Medium** |
| `AxesTextColor` | `&H00000000` | Visual | `colorValue(...)` for axis tick labels and axis lines. Default is transparent-black ARGB `00`; treat alpha-0 default as opaque black for legibility (AI renders it black). | **Medium** |
| `Description` | `""` | Visual | A title string drawn above the plot. | **Medium** |
| `Labels` / `LabelsFromString` | `""` | Visual | Comma-separated X-axis category labels (reuse `elementsFromString`). For bar/pie these are the category names. | **Medium** |
| `XFromZero` | `False` | Behavioral | Forces X-axis origin to 0 vs. min data X when computing auto-bounds. | **Medium** |
| `YFromZero` | `False` | Behavioral | Forces Y-axis origin to 0 vs. min data Y. | **Medium** |
| `PieRadius` | `100` | Visual | Pie-chart hole fill 0–100% (donut). Only meaningful when `Type===4`. | **Low** |
| `ValueFormat` | `""` | Behavioral | Number-format string for axis labels/point values (`chart_value_type`). Best-effort numeric formatting; low priority. | **Low** |
| `Visible` | `True` | Visual | Gated upstream by `isSimulationVisible`. No branch work. | **High** |
| `Width` / `Height` | `""` (automatic) | Visual | `sizeStyle` via `baseStyle()`. Give the canvas a sensible intrinsic size (e.g. 320×220) for automatic. | **High** |
| `Left` / `Top` | `""` | Behavioral | Only inside `AbsoluteArrangement`; handled by `baseStyle()` positioning. | **Low** |

ChartData2D (child) properties worth simulating:

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `ElementsFromPairs` | `""` | Behavioral→Visual | Comma-separated `x1,y1,x2,y2,...`; parsed into the series' initial points. The single most important data entry path in the designer. | **High** |
| `Color` | `&HFF000000` | Visual | Series stroke/fill color. | **High** |
| `Label` | `""` | Visual | Series name shown in the legend / EntryClick. | **Medium** |
| `LineType` | `0` | Visual | linear/curved/stepped — approximate (curved→smoothing, stepped→step path). | **Low** |
| `PointShape` | `0` | Visual | circle/square/triangle/cross/x marker for scatter. | **Low** |
| `DataLabelColor` | `&HFF000000` | Visual | Color of per-point value labels (if drawn). | **Low** |
| `Source`, `DataSourceKey`, `*XColumn`/`*YColumn`, `SpreadsheetUseHeaders`, `DataFile*` | `""` | Behavioral | **Cannot be honored** — bind to unimplemented external components. Store but ignore; the import methods that consume them are `Unsupported`. | — |

Trendline (child) properties worth simulating:

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `ChartData` | `""` | Behavioral | Names the `ChartData2D` to fit. Resolve by component name from `state`. | **High** (if Trendline shipped) |
| `Model` | `Linear` | Behavioral | best_fit_model — implement Linear first (least-squares); Quadratic/Exponential/Logarithmic/Power are stretch. | **High** |
| `Color` | `&H00000000` | Visual | Line color (treat alpha-0 default as black). | **Medium** |
| `Extend` | `True` | Behavioral | Extend the fit line beyond data range vs. clip to data domain. | **Medium** |
| `StrokeStyle` | `1` | Visual | solid/dashed/dotted (StrokeStyleEnum). | **Low** |
| `StrokeWidth` | `1.0` | Visual | Line width in px. | **Low** |
| `Visible` | `True` | Visual | Toggle drawing the fit line. | **Medium** |
| `CorrelationCoefficient`, `RSquared`, `LinearCoefficient`, `YIntercept`, `Predictions`, `Results`, `XIntercepts`, etc. | computed | Behavioral (read-only) | Computed from the regression; returned by `GetProperty` / `GetResultValue`. Linear is exact; others depend on which models are implemented. | **Medium** |

Defaults that **cannot be honored in a browser**: all `ChartData2D` external-source props (listed above). They are stored verbatim so blocks reading them get a value, but they drive no behavior.

## Events

Chart:

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `EntryClick` | `series` (component), `x` (any), `y` (number) | Renderer hit-tests a pointer click against drawn data points; on a hit, `emitEvent(chartName, 'EntryClick', [seriesComponentName, x, y])`. The `series` arg is a component reference — pass the `ChartData2D` component name string (the runtime resolves component args by name elsewhere). Fully fireable in a browser. |

ChartData2D:

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `EntryClick` | `x` (any), `y` (number) | Same click hit-test, dispatched on the child series component. Fire both the Chart-level and series-level events on a point hit (matches AI behavior). |

Trendline:

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `Updated` | `results` (dictionary) | Fire after the fit is (re)computed — i.e. when the bound `ChartData2D` data changes or `Model` changes. The simulator recomputes on data mutation and calls `runEvent(trendlineName, 'Updated', [resultsDict])`. |

No event in this set requires a sensor or permission; all are fireable. Add `Chart: { EntryClick: ['series', 'x', 'y'] }` and `ChartData2D: { EntryClick: ['x', 'y'] }` (and `Trendline: { Updated: ['results'] }`) to `EVENT_ARGS` in `simulation-capabilities.js`.

## Methods

Chart methods:

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `SetDomain` | `(minimum:number, maximum:number)` | Store `_domainMin`/`_domainMax` axis-bound state; renderer uses them instead of auto-bounds. Feasible. |
| `SetRange` | `(minimum:number, maximum:number)` | Store `_rangeMin`/`_rangeMax`; renderer respects. Feasible. |
| `ExtendDomainToInclude` | `(x:number)` | Widen stored domain to include `x` (no-op if already inside). Feasible. |
| `ExtendRangeToInclude` | `(y:number)` | Widen stored range to include `y`. Feasible. |
| `ResetAxes` | `()` | Clear stored domain/range overrides → renderer reverts to auto-bounds. Feasible. |

ChartData2D methods:

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `AddEntry` | `(x:text, y:text)` | Append `{x,y}` to the series `_points`; re-render. (Bar rounds x to int; pie x is a label.) Feasible. |
| `RemoveEntry` | `(x:text, y:text)` | Remove first matching point. Feasible. |
| `Clear` | `()` | Empty `_points`. Feasible. |
| `DoesEntryExist` | `(x, y) → boolean` | Scan `_points`; return bool. Feasible. |
| `GetAllEntries` | `() → list` | Return points as list-of-`[x,y]` lists. Feasible. |
| `GetEntriesWithXValue` | `(x) → list` | Filter by x. Feasible. |
| `GetEntriesWithYValue` | `(y) → list` | Filter by y. Feasible. |
| `ImportFromList` | `(list)` | Parse list of `[x,y]` pairs, append. Feasible. |
| `ImportFromTinyDB` | `(tinyDB, tag)` | **Feasible later** — TinyDB is implemented; read the tag's value (a list) and import. Initially can `Unsupported` until wired, then upgrade. |
| `ImportFromCloudDB` | `(cloudDB, tag)` | **Unsupported** — CloudDB not implemented; no browser backing. |
| `ImportFromWeb` | `(web, xCol, yCol)` | **Unsupported** — Web component / network fetch column parsing not implemented. |
| `ImportFromSpreadsheet` | `(sheet, xCol, yCol, useHeaders)` | **Unsupported** — Spreadsheet component not implemented. |
| `ImportFromDataFile` | `(dataFile, xCol, yCol)` | **Unsupported** — DataFile component not implemented. |
| `ChangeDataSource` / `RemoveDataSource` | component-source linkage | **Unsupported** — external data-source binding not modeled. |
| `HighlightDataPoints` | `(dataPoints:list, color:number)` | Stretch — recolor specified points. Stub `Unsupported` initially. |
| `DrawLineOfBestFit` | `(xList, yList)` (deprecated) | **Unsupported** — deprecated; superseded by Trendline. |

Trendline methods:

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `GetResultValue` | `(value:text) → any` | Return a field from the computed fit (slope, Yintercept, r^2, correlation coefficient, predictions, all values, etc. per the LOBFValues option list). Feasible for whatever models are implemented; return `NaN`/null for unimplemented model fields. |
| `DisconnectFromChartData` | `()` | Clear the `ChartData` linkage; stop drawing the fit. Feasible. |

## Implementation Plan

This is large enough that it should ship in phases. **Phase 1**: Chart + ChartData2D in-memory rendering (line/scatter/area/bar/pie) + AddEntry/Clear/ImportFromList + EntryClick + axis methods. **Phase 2**: Trendline (linear) + `Updated`. **Phase 3**: remaining models / ImportFromTinyDB / highlight.

### simulation-capabilities.js

1. Add to `SIMULATION_SUPPORTED_TYPES`: `'Chart'`, `'ChartData2D'`, `'Trendline'`. Do **not** add to `SIMULATION_NONVISIBLE_TYPES` — all three are `nonVisible = false`. (ChartData2D/Trendline render nothing on their own, but they are visible-category children; the Chart branch reads their state and draws them, and their own render branch should produce no DOM.)
2. Defaults blocks in `buildSimulationDefaults()`:
   ```js
   Chart: {
     Visible: true,
     Width: -1,
     Height: -1,
     Type: 0,
     BackgroundColor: 'transparent',
     AxesTextColor: '&HFF000000',
     Description: '',
     GridEnabled: true,
     LegendEnabled: true,
     PieRadius: 100,
     LabelsFromString: '',
     Labels: [],
     XFromZero: false,
     YFromZero: false,
     ValueFormat: '',
   },
   ChartData2D: {
     Visible: true,
     Color: '&HFF000000',
     Label: '',
     LineType: 0,
     PointShape: 0,
     DataLabelColor: '&HFF000000',
     ElementsFromPairs: '',
     Points: [],            // derived series data
   },
   Trendline: {
     Visible: true,
     ChartData: '',
     Color: '&HFF000000',
     Extend: true,
     Model: 'Linear',
     StrokeStyle: 1,
     StrokeWidth: 1,
   },
   ```
3. New `SIMULATION_VISUAL_PROPS` names (props NOT in this set are stripped before reaching the renderer): `'Type'`, `'AxesTextColor'`, `'Description'`, `'GridEnabled'`, `'LegendEnabled'`, `'PieRadius'`, `'XFromZero'`, `'YFromZero'`, `'ValueFormat'`, `'LabelsFromString'`, `'Labels'`, `'Color'`, `'Label'`, `'LineType'`, `'PointShape'`, `'DataLabelColor'`, `'ElementsFromPairs'`, `'Points'`, `'ChartData'`, `'Extend'`, `'Model'`, `'StrokeStyle'`, `'StrokeWidth'`. (`Visible`, `Width`, `Height`, `Left`, `Top`, `BackgroundColor` already present.) Note `Color`/`Label` are generic names — they are gated per-type anyway since only Chart/ChartData2D/Trendline use them here.
4. `isBooleanProp`: add `'GridEnabled'`, `'LegendEnabled'`, `'XFromZero'`, `'YFromZero'`, `'Extend'`.
5. `isNumericProp`: add `'Type'`, `'PieRadius'`, `'LineType'`, `'PointShape'`, `'StrokeStyle'`, `'StrokeWidth'`. Leave color props (`AxesTextColor`, `DataLabelColor`, `Color`) out (color strings, handled by `colorValue`). `Model` is an enum string — leave out.
6. `coerceSimulationValue`: add `if (propName === 'ElementsFromPairs') return String(value ?? '');` and `if (propName === 'Labels') return normalizeElements... / elementsFromString`. Other props fall through fine.
7. `deriveStateFromDesignerProps`: add derived state:
   - When `componentType === 'ChartData2D'` and key is `ElementsFromPairs`, parse the comma list into `next.Points` (pairs `[{x,y},...]`).
   - When `componentType === 'Chart'` and key is `LabelsFromString`, set `next.Labels = elementsFromString(value)` (mirrors the existing `ElementsFromString → Elements` pattern).
   Add a small `parsePairs(str)` helper near `elementsFromString`.

### SimulationComponent.svelte

Add three branches before the final `{:else}` placeholder.

**Chart** branch (the real renderer). It reads its children's live series state from `state[child.name]` (this works because `state` and `node.children` are both in scope, exactly like the arrangement branches recurse over `node.children`):
```svelte
{:else if node.type === 'Chart'}
  <div class="sim-chart" class:sim-unsupported={unsupportedHere}
       style={baseStyle(`background:${colorValue(props.BackgroundColor,'transparent')};`)}
       data-sim-component={node.name}>
    {#if hasValue(props.Description)}
      <div class="sim-chart-desc" style={`color:${colorValue(props.AxesTextColor,'#000')};`}>{props.Description}</div>
    {/if}
    <svg viewBox="0 0 320 220" class="sim-chart-svg"
         on:click={chartClick}>
      <!-- grid (if GridEnabled), axes (AxesTextColor), then per-series paths/points,
           then trendline overlays, computed from chartSeries() -->
    </svg>
    {#if boolValue(props.LegendEnabled, true)}
      <div class="sim-chart-legend">
        {#each chartSeries() as s}
          <span class="sim-chart-legend-item"><i style={`background:${s.color}`}></i>{s.label}</span>
        {/each}
      </div>
    {/if}
    <!-- children rendered for completeness but produce no DOM -->
    {#each node.children || [] as child (child.pathId || child.name)}
      <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
    {/each}
  </div>
```
- Add a `chartSeries()` helper in the script that walks `node.children`, picks `ChartData2D` children, reads `state[child.name].Points`, `Color` (via `colorValue`), `Label`, `LineType`, `PointShape`, and returns drawable series; plus `Trendline` children resolved to their fit lines. SVG is preferable to `<canvas>` for click hit-testing (each `<circle>`/`<rect>` can carry `on:click` with its `(seriesName, x, y)` → `emitEvent`). Use `<polyline>`/`<path>` for line/area, `<rect>` for bar, `<circle>`/`<path>` for scatter markers, and `<path>` arcs for pie.
- Axis bounds: compute min/max from all series points, honoring `XFromZero`/`YFromZero` and any stored `_domainMin/_domainMax/_rangeMin/_rangeMax` (set by host methods). A small linear scale maps data → the 320×220 viewBox.
- `chartClick` hit-tests the nearest point (or rely on per-point `on:click`) and calls `emitEvent(node.name, 'EntryClick', [seriesName, x, y])` and `emitEvent(seriesName, 'EntryClick', [x, y])`.
- Reuse helpers: `baseStyle()`, `colorValue()`, `boolValue()`, `numberOr()`, `hasValue()`, `emitEvent()`. No new external lib — pure SVG + a hand-written linear scale.

**ChartData2D** and **Trendline** branches render nothing visible (they are data carriers); return an empty hidden marker so they don't fall into the unsupported placeholder:
```svelte
{:else if node.type === 'ChartData2D' || node.type === 'Trendline'}
  <span class="sim-chart-data" hidden data-sim-component={node.name}></span>
```
(They must still appear in `SIMULATION_SUPPORTED_TYPES` so `unsupportedSimulationComponents` stops flagging them and the dashed placeholder disappears.)

CSS: a `.sim-chart` flex column (desc / svg / legend), `.sim-chart-svg { width:100%; height:auto; }`, legend chips. Keep it lightweight.

### SimulateOverlay.svelte

**None.** No dialogs, toasts, or modal effects — `EntryClick` is a plain event dispatched via `emitEvent`, and the axis methods are pure state writes. The overlay's existing generic statePatch application is sufficient for `SetDomain`/`SetRange`/`Clear`/`AddEntry` re-renders (they write component state, which flows back to the renderer through the normal patch path).

### simulation_wasm.go

Generic store path covers Chart's plain properties (`BackgroundColor`, `GridEnabled`, `LegendEnabled`, `AxesTextColor`, `Description`, `Type`, `PieRadius`, `XFromZero`, `YFromZero`, `Labels`, `Visible`, `Width`, `Height`). New handlers needed:

- **CallMethod dispatch**: add `case "Chart": return h.callChartMethod(...)`, `case "ChartData2D": return h.callChartData2DMethod(...)`, `case "Trendline": return h.callTrendlineMethod(...)`.
- `callChartMethod`: `SetDomain`/`SetRange`/`ExtendDomainToInclude`/`ExtendRangeToInclude` write `_domainMin`/`_domainMax`/`_rangeMin`/`_rangeMax` via `setProperty` (so they ride the statePatch to the renderer); `ResetAxes` clears them. Anything else → `h.Unsupported`.
- `callChartData2DMethod`: read/write the series' `Points` list in `h.state`:
  - `AddEntry`/`RemoveEntry`/`Clear`/`ImportFromList` mutate `Points` (then `setProperty` so the patch flows out).
  - `DoesEntryExist` → `runtime.BoolVal`; `GetAllEntries`/`GetEntriesWithXValue`/`GetEntriesWithYValue` → `runtime.ListVal` of `[x,y]` lists.
  - After any mutation, if a `Trendline` is bound to this series, recompute and `h.runEvent(trendlineName, "Trendline", "Updated", [resultsDict])`. Also fire `EntryClick` only from the renderer side (not here).
  - `ImportFromCloudDB`/`ImportFromWeb`/`ImportFromSpreadsheet`/`ImportFromDataFile`/`ChangeDataSource`/`RemoveDataSource`/`HighlightDataPoints`/`DrawLineOfBestFit` → `h.Unsupported("method", componentName+"."+method)`. (`ImportFromTinyDB` can be wired to the existing `h.tinyDB` store in a later phase; until then, `Unsupported`.)
- `callTrendlineMethod`: `GetResultValue` returns the requested fit field from a JS-side or Go-side least-squares computation (Linear at minimum); `DisconnectFromChartData` clears `ChartData`.
- **GetProperty**: add a `case "Trendline"` for the computed read-only props (`CorrelationCoefficient`, `RSquared`, `LinearCoefficient`, `YIntercept`, `Predictions`, `Results`, `XIntercepts`, `YIntercept`, etc.) — compute from the bound series' `Points`. For unimplemented models, return `NaN`/null and optionally `Unsupported`.
- **SetProperty**: `Type` is `read-only` per spec, but the designer still sets it at init; the generic store path handles that. `ChartData` (Trendline → series name) and `Model` writes should trigger a `recomputeTrendline` + `Updated` event — add a small `case` similar to the TextBox `Text`→`TextChanged` pattern.
- Where regression is non-trivial (quadratic/exponential/log/power), it is acceptable to implement only Linear in Go and `Unsupported` the rest, or do the math JS-side in the renderer and emit `Updated` from there.

### design-schema-tree.js

**Already handled.** `CHART_CHILD_TYPES = new Set(['ChartData2D', 'Trendline'])` and `canContainDesignComponent` already returns `parent === 'Chart'` for these children, and returns `false` for any other child placed under a `Chart`. `unsupportedSimulationComponents` stops flagging `Chart`/`ChartData2D`/`Trendline` the moment they are added to `SIMULATION_SUPPORTED_TYPES`. `designTreeToInitialState` merges the new `SIMULATION_DEFAULTS[type]` and runs `deriveStateFromDesignerProps` (so `ElementsFromPairs → Points` and `LabelsFromString → Labels` derivations apply at init). No edit required.

## Dependencies & Ordering

- **External libraries:** none required. Hand-rolled SVG + a linear scale draws all five chart types and is preferable to pulling in Chart.js/D3 (heavy, and harder to wire per-point `EntryClick`). A charting lib could be a later refinement but is not needed for the spec surface.
- **Prerequisite components:** `Chart` must be implemented together with `ChartData2D` (the data carrier) — a `Chart` with no `ChartData2D` child draws an empty plot. `Trendline` depends on `ChartData2D` and on the regression math, so it is a clean Phase 2 follow-on. `ImportFromTinyDB` depends on the already-implemented `TinyDB` store (Phase 3). The other external imports depend on **unimplemented** components (CloudDB, Web, Spreadsheet, DataFile) and are intentionally left `Unsupported`.

## Web-Platform Limitations & Fidelity Caveats

- **No MPAndroidChart fidelity.** The SVG renderer approximates axes, grids, lines, bars, and pie slices but will not match the native library's tick selection, curved/stepped line easing, anti-aliasing, label collision handling, or animations.
- **No gesture pan/zoom.** Real device charts support pinch-zoom and drag-pan of the viewport; the simulated chart is static (the `SetDomain`/`SetRange` API still works programmatically).
- **External data sources unavailable.** Any series bound to CloudDB / Web / Spreadsheet / DataFile / a sensor draws nothing — those source components don't exist in the simulator and have no browser backing. The corresponding `ImportFrom*` methods report `Unsupported` (orange warning). Designer props (`Source`, `*XColumn`, `DataSourceKey`, etc.) are stored but inert.
- **Regression scope.** Only the models actually implemented (Linear first) produce a real `Trendline`; other models return `NaN`/null read-only props until added. `XIntercepts` (NaN/single/list per spec) and the full `Results` dictionary are only as complete as the implemented models.
- **`Type` is designer-only.** The spec marks `Type` `read-only` at runtime, so the chart type cannot change after launch — matching AI. Set it in the designer.
- **`ValueFormat` / number formatting** is best-effort; AI's exact format-string semantics are not fully reproduced.
- **`PointShape` / `LineType` / `StrokeStyle`** are approximated with basic SVG markers/dash arrays, not the native shape set.

## Effort Estimate

**XL** — this is three interdependent components (`Chart` container + `ChartData2D` data carrier + `Trendline` regression), a from-scratch SVG multi-type chart renderer (line/scatter/area/bar/pie) with axis scaling and click hit-testing, ~6 Go method handlers across three new `callXMethod` functions plus computed-property `GetProperty` cases and least-squares math, designer-prop derivations (`ElementsFromPairs`/`LabelsFromString`), and a long tail of `Unsupported` external-import stubs. Realistically phased: Phase 1 (Chart + ChartData2D in-memory rendering + EntryClick + axis methods) is **L** on its own; Trendline + remaining imports push the full surface to **XL**.
