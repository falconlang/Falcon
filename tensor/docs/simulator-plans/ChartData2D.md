# ChartData2D Simulator Implementation Plan

## Overview

`ChartData2D` is a visible Charts component (categoryString `CHARTS`, `nonVisible = false`). Per the spec helpString it is "A component that holds (x, y)-coordinate based data." In App Inventor it is **not a standalone widget**: it is a *data series* that must be placed inside a `Chart` parent. The `Chart` owns the plot area (axes, grid, legend, type) and each child `ChartData2D` contributes one series (a labelled set of x/y points) to that plot, styled by its `Color`, `Label`, `LineType`, and `PointShape`. How a series is drawn depends entirely on the parent `Chart.Type` (0=line, 1=scatter, 2=area, 3=bar, 4=pie).

Container relationship: **child of `Chart`** (siblings: `Trendline`). It has no children of its own. It has no independent on-screen box — it has no `Width`/`Height`/`Visible`/`Left`/`Top` designer properties (confirmed: none appear in the spec's `designerProperties`). Its visual contribution exists only as marks inside the parent Chart's plot rectangle.

Critically: **`Chart` is itself currently UNIMPLEMENTED** (not in `SIMULATION_SUPPORTED_TYPES`). `ChartData2D` cannot render anything without a Chart host, so this plan is gated on a Chart implementation and is best executed as one combined Chart + ChartData2D effort.

## Feasibility Verdict

**Partially feasible (in a browser simulator).**

The *core* — plotting (x, y) series as line/scatter/area/bar/pie inside a Chart, with per-series color, point shape, line type, a legend, axes and grid — is fully achievable client-side with inline SVG (or `<canvas>`). No native APIs are involved for the static/blocks-driven case. `AddEntry`, `Clear`, `ImportFromList`, `RemoveEntry`, `DoesEntryExist`, `GetAllEntries`, `GetEntriesWithXValue`/`YValue`, `ElementsFromPairs`, and the `EntryClick` event are all pure data/geometry operations that map cleanly to SVG hit-testing and a re-render. This part is a faithful approximation.

What is **not feasible / only stubbable**, and why:

- **Live data-source binding** (`Source`, `DataSourceKey`, `ChangeDataSource`, `RemoveDataSource`): App Inventor lets a series subscribe to a *live sensor or DB* (AccelerometerSensor, GyroscopeSensor, LocationSensor, OrientationSensor, Pedometer, ProximitySensor, BluetoothClient, CloudDB, Web, TinyDB) and append points in real time as the source emits. In a browser there are **no real sensors/Bluetooth/CloudDB**, and most of those source components are themselves unimplemented in the simulator. `ImportFromTinyDB` is the one realistic exception (TinyDB *is* implemented and is a synchronous client-side store). The rest must be marked `Unsupported`.
- **`ImportFromDataFile` / `ImportFromSpreadsheet` / `ImportFromWeb`** and their column props (`DataFileXColumn`, `DataFileYColumn`, `SpreadsheetXColumn`/`YColumn`, `SpreadsheetUseHeaders`, `WebXColumn`/`YColumn`): depend on `DataFile`, `Spreadsheet`, and `Web` components that do not exist in the simulator and which themselves involve file/network access. `Unsupported`.
- **`ImportFromCloudDB`**: no CloudDB (network service). `Unsupported`.
- **`DrawLineOfBestFit`**: deprecated; superseded by the `Trendline` sibling. Skip (treat as `Unsupported`/no-op).

Realistic simulated approximation: implement the Chart host and ChartData2D as an **inline-SVG plot** that renders points provided statically (`ElementsFromPairs` in the designer) and mutated at runtime via `AddEntry`/`RemoveEntry`/`Clear`/`ImportFromList`/`ImportFromTinyDB`. Mark every external/sensor source path `Unsupported`. This covers the overwhelmingly common tutorial use of Chart (plot fixed or block-appended data) while honestly refusing the device-bound paths.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Color` | `&HFF000000` (black) | Visual | Series stroke/point/bar fill color. `colorValue(props.Color, '#000000')` -> SVG `stroke`/`fill`. | **High** |
| `Label` | `""` | Visual | Series name shown in the parent Chart's legend; read by the Chart host. | **High** |
| `ElementsFromPairs` | `""` | Behavioral (seed data) | Comma-separated `x1,y1,x2,y2,...`; parse into `[{x,y},...]` to seed the series `entries`. This is the only designer way to give a series static data. | **High** |
| `PointShape` | `0` | Visual | `chart_point_shape` enum (0=circle,1=square,2=triangle,3=cross,4=x). Only meaningful when `Chart.Type` is scatter. Render the SVG point marker accordingly. | **Medium** |
| `LineType` | `0` | Visual | `chart_line_type` enum (0=linear,1=curved,2=stepped). Only meaningful for line/area charts. Linear = polyline; stepped = step path; curved = approximate with a smoothed path (Catmull-Rom/`Q` segments). | **Medium** |
| `DataLabelColor` | `&HFF000000` | Visual | Color of per-point value labels (when the parent draws them). `colorValue(...)`. | **Low** |
| `Colors` | n/a (block-only, list) | Visual | Read-write list of per-entry colors; honor at render if set, else fall back to `Color`. | **Low** |
| `Source` | `""` (`chart_data_source`) | Behavioral | Designer-bound live source component. **Cannot be honored** (no sensors); ignore at designer level, `Unsupported` at runtime. | — (unsupported) |
| `DataSourceKey` | `""` | Behavioral | Source key (TinyDB tag, sensor axis, etc.). Only meaningful if `Source` is a TinyDB; otherwise unsupported. | — (mostly unsupported) |
| `DataFileXColumn` / `DataFileYColumn` | `""` | Behavioral | DataFile column refs. No DataFile component. **Cannot be honored.** | — (unsupported) |
| `SpreadsheetUseHeaders` | `""` (boolean) | Behavioral | Spreadsheet import flag. No Spreadsheet component. **Cannot be honored.** | — (unsupported) |
| `SpreadsheetXColumn` / `SpreadsheetYColumn` | `""` | Behavioral | Spreadsheet column refs. **Cannot be honored.** | — (unsupported) |
| `WebXColumn` / `WebYColumn` | `""` | Behavioral | Web import column refs. No Web component / network. **Cannot be honored.** | — (unsupported) |

Note: the import-column properties are `rw: invisible` (no block getter) and are pure inputs to the corresponding `ImportFrom*` methods; since those imports are unsupported, the columns are dead state — store them generically but they never affect the render.

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `EntryClick` | `x` (any), `y` (number) | **Feasible.** Each rendered point/bar/slice carries an `on:click`/`on:pointerup` that calls `emitEvent(node.name, 'EntryClick', [x, y])`. Hit-testing is trivial with discrete SVG point elements. Note the parent `Chart` also has its own `EntryClick(series, x, y)`; the Chart host should fire *both* (the series-level `EntryClick(x,y)` on the clicked `ChartData2D` and the chart-level `EntryClick(series, x, y)` on the parent) to match AI semantics. |

Add `ChartData2D: { EntryClick: ['x', 'y'] }` to `EVENT_ARGS`.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `AddEntry` | `(x text, y text) -> void` | **Feasible.** Push `{x,y}` onto the series `entries`, coerce y (and x for non-pie) to number, re-render. For bar charts round x to nearest int (per spec). |
| `Clear` | `() -> void` | **Feasible.** Empty `entries`, re-render. |
| `RemoveEntry` | `(x text, y text) -> void` | **Feasible.** Remove first matching `{x,y}` if present; re-render. |
| `DoesEntryExist` | `(x text, y text) -> boolean` | **Feasible.** Return whether a matching entry exists. Synchronous return value via `CallMethod`. |
| `GetAllEntries` | `() -> list` | **Feasible.** Return `entries` as a list of `[x,y]` lists. |
| `GetEntriesWithXValue` | `(x text) -> list` | **Feasible.** Filter `entries` by x, return list of `[x,y]`. |
| `GetEntriesWithYValue` | `(y text) -> list` | **Feasible.** Filter `entries` by y, return list of `[x,y]`. |
| `ImportFromList` | `(list) -> void` | **Feasible.** Each element is a 2-item list `[x,y]`; skip invalid; append (does not clear). Re-render. |
| `ImportFromTinyDB` | `(tinyDB component, tag text) -> void` | **Partially feasible.** TinyDB is implemented and client-side; read the tagged value (expected to be a list-of-pairs), parse, append. Implement if the host can resolve the referenced TinyDB's store; otherwise `Unsupported`. |
| `ChangeDataSource` | `(source component, keyValue text) -> void` | **Unsupported** (except TinyDB). Re-binding to a live sensor/DB source has no browser equivalent. If `source` is a TinyDB, treat like a one-shot `ImportFromTinyDB`; else `h.Unsupported`. |
| `RemoveDataSource` | `() -> void` | **No-op / Unsupported.** No live binding exists to remove. Safe no-op. |
| `ImportFromDataFile` | `(dataFile component, xCol text, yCol text) -> void` | **Unsupported.** No DataFile component / file access. |
| `ImportFromSpreadsheet` | `(spreadsheet component, xCol, yCol, useHeaders bool) -> void` | **Unsupported.** No Spreadsheet component / network. |
| `ImportFromWeb` | `(web component, xCol, yCol) -> void` | **Unsupported.** No Web component / network. |
| `ImportFromCloudDB` | `(cloudDB component, tag text) -> void` | **Unsupported.** No CloudDB (network service). |
| `HighlightDataPoints` | `(dataPoints list, color number) -> void` | **Partially feasible.** Override the fill of the listed `[index, value]` points with `color`; re-render. Cosmetic, low priority. |
| `DrawLineOfBestFit` | `(xList list, yList list) -> void` | **Unsupported (deprecated).** Superseded by `Trendline`. No-op. |

## Implementation Plan

> This component cannot be implemented alone. The plan below assumes the **Chart host** is implemented in the same pass (Chart owns the SVG plot, axes/grid/legend, `Type`, and reads its `ChartData2D` children's series state). Steps specific to ChartData2D are called out; Chart-host steps are noted where the series depends on them.

### simulation-capabilities.js

1. Add `'ChartData2D'` **and** `'Chart'` to `SIMULATION_SUPPORTED_TYPES` (both, or ChartData2D stays a dashed placeholder host-less). Do **not** add either to `SIMULATION_NONVISIBLE_TYPES` (both are visible). ChartData2D has no standalone box but it is still a "visible" series; the renderer simply outputs nothing on its own (the Chart draws it).
2. Add a defaults block inside `buildSimulationDefaults()`:
   ```js
   ChartData2D: {
     Color: '&HFF000000',
     Label: '',
     ElementsFromPairs: '',
     LineType: 0,
     PointShape: 0,
     DataLabelColor: '&HFF000000',
     Colors: [],
   },
   ```
   (Plus a sibling `Chart: { Visible: true, Width: -1, Height: -1, Type: 0, BackgroundColor: 'transparent', Description: '', GridEnabled: true, LegendEnabled: true, LabelsFromString: '', PieRadius: 100, XFromZero: false, YFromZero: false }` if implementing Chart here.) Note: do **not** add `Visible`/`Width`/`Height`/`Enabled` to the ChartData2D defaults — the spec has none.
3. `SIMULATION_VISUAL_PROPS` additions (props NOT in this set are stripped before reaching the renderer): `'Color'`, `'Label'`, `'ElementsFromPairs'`, `'LineType'`, `'PointShape'`, `'DataLabelColor'`, `'Colors'`. (For Chart also add `'Type'`, `'GridEnabled'`, `'LegendEnabled'`, `'LabelsFromString'`, `'PieRadius'`, `'XFromZero'`, `'YFromZero'`, `'Description'` — `Width`/`Height`/`Left`/`Top`/`Visible`/`BackgroundColor` already present.)
4. `isBooleanProp`: add `'GridEnabled'`, `'LegendEnabled'`, `'XFromZero'`, `'YFromZero'`, `'SpreadsheetUseHeaders'` (Chart + series booleans).
5. `isNumericProp`: add `'LineType'`, `'PointShape'`, `'Type'`, `'PieRadius'` (enum/int values coerced to number).
6. `coerceSimulationValue`: add a case so `'ElementsFromPairs'` is kept as a string (parsed in `deriveStateFromDesignerProps`), and `'Colors'` is normalized to an array. Other props fall through (colors stay ARGB strings, enums become numbers via `isNumericProp`).
7. `deriveStateFromDesignerProps`: add a `ChartData2D` branch that, when `ElementsFromPairs` is set, parses `"x1,y1,x2,y2,..."` into `next.entries = [{x,y},...]` (the runtime series state the Chart host reads). Mirror the existing `ElementsFromString -> Elements` derivation pattern.

### SimulationComponent.svelte

ChartData2D has **no standalone render branch** — a bare series draws nothing. Two coordinated pieces:

1. **ChartData2D branch (no-op visual):** add a branch that renders an empty/zero-size marker so it is recognized (not "unsupported") but contributes no box:
   ```svelte
   {:else if node.type === 'ChartData2D'}
     <!-- series data only; drawn by the parent Chart. emits nothing on its own -->
     {''}
   ```
   The series' live state (`entries`, `Color`, `Label`, `LineType`, `PointShape`) is read by the Chart host from `state[node.name]`.

2. **Chart host branch (the actual plot):** add a `{:else if node.type === 'Chart'}` branch that renders an inline `<svg>` plot. It iterates `node.children` of type `ChartData2D`, reads each child's state from the shared `state` map, computes the data-domain -> pixel transform (honoring `XFromZero`/`YFromZero`), and draws:
   - axes + grid (`GridEnabled`), legend (`LegendEnabled`, using each series' `Label` + `Color`);
   - per series, by `Chart.Type`: line (`<polyline>`/path, respecting `LineType`), scatter (point markers per `PointShape`), area (filled path), bar (`<rect>` groups), pie (`<path>` arcs using `PieRadius`).
   - Each point/bar/slice gets `on:click`/`on:pointerup` -> `emitEvent(childName, 'EntryClick', [x, y])` and `emitEvent(node.name /*Chart*/, 'EntryClick', [childName, x, y])`.

   Reuse helpers: `baseStyle()` (Width/Height/Left/Top/Visible for the Chart box), `colorValue()` (series colors, axes/grid), `numberOr()` (enum/numeric props), `boolValue()` (Grid/Legend/XFromZero/YFromZero). Note the children render via `<svelte:self ... parentType={node.type}>` recursion (lines ~848-880); the Chart branch should *not* rely on that recursion to draw series (it draws them itself from `state`), but it should still mount children so their state is registered. Pattern reference: the SVG/markup approach is closest to building a small custom visual like the Slider's CSS-var track, plus ListView's "read structured child data from state" pattern.

   CSS: add `.sim-chart { display:inline-block; }` and SVG element classes; small, since SVG carries its own geometry.

### SimulateOverlay.svelte

**None.** No dialogs, toasts, or host "effects" are required. `EntryClick` is a normal event (via `emitEvent`), and the plot mutates through property/method round-trips that re-render reactively. (One optional exception: if `AddEntry` etc. are implemented as host-side state writes that the overlay must echo back as a property patch, that flows through the existing generic property-patch path, not a new effect.)

### simulation_wasm.go

A per-type `callChartData2DMethod` handler is needed (the methods are stateful and some return values):

- **`GetProperty`**: generic store path suffices for `Color`, `Label`, `LineType`, `PointShape`, `DataLabelColor`, `Colors`. The series `entries` should be exposed as host state so the renderer can read it; store it under the component's state alongside the props.
- **`SetProperty`**: generic store path for the styling props. Setting `ElementsFromPairs`/`Colors` updates state and triggers re-render (no special event).
- **`CallMethod`** (add `case "ChartData2D": return h.callChartData2DMethod(...)`):
  - `AddEntry`, `Clear`, `RemoveEntry`, `ImportFromList`: mutate the host-held `entries` slice, write back to state, return nil.
  - `DoesEntryExist` -> boolean; `GetAllEntries`/`GetEntriesWithXValue`/`GetEntriesWithYValue` -> list (return `runtime.Value` lists).
  - `ImportFromTinyDB` / `ChangeDataSource(TinyDB,...)`: if the host can resolve the named TinyDB's namespace store, parse the list-of-pairs and append; else `h.Unsupported("method", name+".ImportFromTinyDB")`.
  - `RemoveDataSource`, `DrawLineOfBestFit`: no-op (DrawLineOfBestFit also acceptable as `Unsupported`, deprecated).
  - `HighlightDataPoints`: store an override-color map in state; re-render.
  - `ImportFromDataFile`, `ImportFromSpreadsheet`, `ImportFromWeb`, `ImportFromCloudDB`, and `ChangeDataSource` to a non-TinyDB source: `h.Unsupported("method", name+"."+method)` (orange warning).
- **Events**: fire `EntryClick` via `h.runEvent(componentName, componentType, "EntryClick", args)` when the renderer dispatches the click; also the Chart-level `EntryClick`.
- The Chart host needs its own `GetProperty`/`SetProperty`/`callChartMethod` (Type read-only, SetDomain/SetRange/Extend*/ResetAxes adjust axis bounds in state). Compile with `npm run build:wasm`.

### design-schema-tree.js

**Already handled.** `CHART_CHILD_TYPES = new Set(['ChartData2D', 'Trendline'])` and `if (CHART_CHILD_TYPES.has(child)) return parent === 'Chart';` already restrict ChartData2D to a Chart parent (line 26, 69), and line 72 forbids a Chart from holding ordinary children. `unsupportedSimulationComponents` stops flagging ChartData2D (and Chart) once they are in `SIMULATION_SUPPORTED_TYPES`; `designTreeToInitialState` merges `SIMULATION_DEFAULTS.ChartData2D` automatically. No edit required here.

## Dependencies & Ordering

- **Hard prerequisite: `Chart`.** ChartData2D has no rendering surface without a Chart host. Implement Chart (plot area, axes/grid/legend, `Type`, axis-bound methods) **first or in the same pass**. Without it, ChartData2D is dead state.
- **Soft dependency: `TinyDB`** (already implemented) — only needed for `ImportFromTinyDB`/`ChangeDataSource` to be more than `Unsupported`.
- **Sibling: `Trendline`** is independent; not required for ChartData2D.
- **External libraries: none required.** Inline SVG (or `<canvas>`) is sufficient for line/scatter/area/bar/pie. A charting lib (Chart.js / uPlot) would improve fidelity and reduce hand-rolled geometry but is an optional heavier dependency, not a necessity for the supported (static + block-appended) data paths.

## Web-Platform Limitations & Fidelity Caveats

- **No live data sources.** Real-time binding to sensors (Accelerometer/Gyroscope/Location/Orientation/Pedometer/Proximity), Bluetooth, CloudDB, and Web are impossible in a browser; the whole `Source`/`DataSourceKey`/`ChangeDataSource`/`RemoveDataSource` surface and `ImportFromDataFile`/`ImportFromSpreadsheet`/`ImportFromWeb`/`ImportFromCloudDB` are `Unsupported`. Only static (`ElementsFromPairs`), block-driven (`AddEntry`/`ImportFromList`), and TinyDB import paths work.
- **Rendering is an approximation of MPAndroidChart.** App Inventor uses MPAndroidChart on device; an SVG re-implementation will differ in exact axis tick selection, label placement, animation, curve smoothing (`LineType=curved` is interpolated, not MPAndroidChart's cubic), and pie-hole rendering (`PieRadius`).
- **`PointShape`/`LineType` only matter for matching `Chart.Type`.** Per the spec these are honored only on scatter (point shape) and line/area (line type) charts; on other types they are silently ignored, matching AI.
- **No determinate device density / pixel sizing.** The plot fills the Chart box; intrinsic px sizing on automatic Width/Height is a simulator choice, not device-accurate.
- **`Colors`/`HighlightDataPoints`** are cosmetic best-effort; exact MPAndroidChart per-entry coloring semantics may diverge.

## Effort Estimate

**XL** — ChartData2D alone is M (series state + a per-type Go method handler + EVENT_ARGS + capability wiring), but it is unusable without the **Chart host**, which is the bulk of the work: an inline-SVG plot engine supporting five chart types, axes/grid/legend, domain/range math and the axis-bound methods, plus hit-testing for `EntryClick`. Combined Chart + ChartData2D, with honest `Unsupported` stubs for every device-bound source/import path, is XL. (If scope is cut to *line + scatter only, static + AddEntry/ImportFromList data*, it drops to L.)
