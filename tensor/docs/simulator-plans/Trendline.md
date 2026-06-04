# Trendline Simulator Implementation Plan

## Overview

`Trendline` (category **CHARTS**, `nonVisible = false`) is a visible-but-non-standalone component. Per the spec's `helpString` it is "A component that predicts a best fit model for a given data series." It attaches to a single `ChartData2D` series (its `ChartData` designer property is a `component:...ChartData2D` reference), computes a regression curve over that series' points using a selectable `Model` (Linear, Quadratic, Cubic, Exponential, Logarithmic, etc.), and draws that curve as a styled line *inside the parent `Chart`*. It exposes the fitted parameters as read-only block properties (`CorrelationCoefficient`, `RSquared`, `LinearCoefficient`, `QuadraticCoefficient`, `ExponentialBase`/`ExponentialCoefficient`, `LogarithmConstant`/`LogarithmCoefficient`, `YIntercept`, `XIntercepts`, `Predictions`, `Results`) and fires `Updated(results)` whenever the underlying data changes and the fit is recomputed.

Container relationship: **child of `Chart`** (CHARTS category). `Trendline` has no independent on-screen geometry — it is rendered as an overlay line on the parent `Chart`'s plotting surface, exactly like `ChartData2D` is a series rendered on that same surface. It is meaningless without (a) a `Chart` to draw on and (b) a `ChartData2D` series to fit. It is a leaf within the Chart subtree (it has no children).

In the simulator it currently renders as a dashed "unsupported" placeholder because `Trendline` is absent from `SIMULATION_SUPPORTED_TYPES`. Crucially, **`Chart` and `ChartData2D` are also absent** from the simulator — there is no plotting surface, no axis mapping, and no series data anywhere in `SimulationComponent.svelte` or `simulation_wasm.go`. `design-schema-tree.js` already encodes the containment rule (`CHART_CHILD_TYPES = {ChartData2D, Trendline}`, `canContainDesignComponent(... child) => parent === 'Chart'`), but nothing consumes it at render time.

## Feasibility Verdict

**Partially feasible — and strictly blocked on `Chart` + `ChartData2D` first.**

Two separable facts:

1. **The math is fully feasible in a browser.** Every model in the spec (linear, quadratic/cubic polynomial least-squares, exponential `y = a·b^x` via log-linearization, logarithmic `y = a + b·ln(x)`) is closed-form ordinary least squares — a handful of summations or a small normal-equations solve, all trivially done in JS with no library. `CorrelationCoefficient`, `RSquared`, `XIntercepts`, `YIntercept`, `Predictions`, and the `Results` dictionary all fall straight out of the fit. Drawing the resulting curve as an SVG/`<canvas>` polyline over a plotting area is also trivial. So the *Trendline-specific* work is genuinely achievable.

2. **There is nothing to draw it on.** `Trendline` produces a line in the parent `Chart`'s data-coordinate space, mapped through that chart's X/Y axes onto pixels. With no `Chart` render surface, no `ChartData2D` series points, and no axis/scale model in the simulator, a `Trendline` has no coordinate system, no data to fit, and no canvas to render onto. Implementing `Trendline` in isolation would yield a component that can compute nothing (no source series) and display nothing (no plot). This is a hard ordering dependency, not a web-platform limitation.

**Web platform itself imposes no real barrier** — `<canvas>`, SVG, and JS regression math are all available client-side; no native API is involved. The only honest blocker is the missing Chart subsystem. Therefore: *feasible as a follow-on to a `Chart`/`ChartData2D` implementation; not implementable on its own.* The realistic approximation, once `Chart`+`ChartData2D` exist, is a faithful one: compute the OLS fit over the parent series' points in JS and stroke it on the same surface the series is drawn on, with all read-only fitted properties exposed exactly.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `ChartData` | `""` | Behavioral (the source series) | Designer ref to a sibling `ChartData2D` (write-only at runtime). Resolve to that series' point list inside the parent `Chart`'s state. **Hard-depends on `ChartData2D` existing.** Empty → no fit, no line. | High (blocking) |
| `Model` | `"Linear"` | Behavioral (which regression) | `best_fit_model` enum → select the JS fit function. Map enum tokens (Linear, Quadratic, Cubic, Exponential, Logarithmic, …) to OLS routines. read-write at runtime. | High |
| `Color` | `&H00000000` (transparent black) | Visual (line color) | `colorValue()`. Note default alpha is `00` → fully transparent; honor it (a line set to transparent is invisible, matching AI). read-write. | High |
| `Extend` | `True` | Behavioral/Visual | Whether the fitted curve is drawn beyond the data's X-range to the chart's axis bounds. With `Extend=true` stroke across the full X-axis; else clip to `[min(x), max(x)]`. Needs the parent Chart's X-range. read-write boolean. | Medium |
| `StrokeStyle` | `"1"` | Visual (line dash) | `stroke_style` enum: 1=Solid, 2=Dashed, 3=Dotted (AI `StrokeStyleEnum`). Map to SVG `stroke-dasharray` / canvas `setLineDash`. read-write. | Medium |
| `StrokeWidth` | `1.0` | Visual (line thickness) | `non_negative_float` → SVG `stroke-width` / canvas `lineWidth` in px. read-write number. | Medium |
| `Visible` | `True` | Visual | Gates whether the line is stroked at all. read-write boolean. | High |
| `CorrelationCoefficient` | (computed) | Behavioral, read-only | Pearson r of fit. Computed from the fit; exposed via `GetProperty`. | Medium |
| `RSquared` | (computed) | read-only | r² coefficient of determination. From fit. | Medium |
| `LinearCoefficient` | (computed) | read-only | Slope (linear/poly leading-into-linear term). From fit. | Medium |
| `QuadraticCoefficient` | (computed) | read-only | x² coefficient (quadratic/cubic). From fit; `0`/absent for linear. | Low |
| `ExponentialBase` / `ExponentialCoefficient` | (computed) | read-only | `b` and `a` in `y = a·b^x`. From log-linearized exponential fit. | Low |
| `LogarithmConstant` / `LogarithmCoefficient` | (computed) | read-only | `a` and `b` in `y = a + b·ln(x)`. From log fit. | Low |
| `YIntercept` | (computed) | read-only | Constant term of the fit. | Low |
| `XIntercepts` | (computed) | read-only | Roots where the curve crosses y=0: `NaN` (none), a single number, or a list. Solve per-model (linear root, quadratic formula, etc.). | Low |
| `Predictions` | (computed) | read-only (list) | Fitted ŷ values at each source x. | Low |
| `Results` | (computed) | read-only (dictionary) | Snapshot dict of all fitted values (slope, intercept, r², predictions, …) — the same payload passed to `Updated`. | Medium |

None of these is *web-impossible*; the only "cannot be honored" cases are entirely downstream of the missing Chart/series data (empty `ChartData` ⇒ every computed property is undefined/`NaN`, which is the correct AI behavior anyway).

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `Updated` | `results` (dictionary) | Fire whenever the fit is recomputed: i.e. when the bound `ChartData2D` series' points change (data added/imported), or when `Model`/`Extend` is set at runtime. The simulator recomputes the OLS fit in JS, builds the `Results` dictionary, restrokes the line, and calls `runEvent(name, "Trendline", "Updated", [resultsDict])`. **Requires the `ChartData2D` series to emit a "points changed" signal** that the Trendline subscribes to — i.e. this event cannot fire until `ChartData2D` data mutation exists in the simulator. No sensor/permission involved, so it is fireable once the data plumbing exists. |

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `DisconnectFromChartData()` | `() → void` | Detach the Trendline from its bound series: clear the internal series reference, stop recomputing/redrawing, hide the line. Pure state change in the host + a render patch. Feasible once the binding exists. |
| `GetResultValue(value)` | `(text: LOBFValues option) → any` | Return one field of the most recent fit. `value` is an `OPTION_LIST` over `LOBFValues` whose underlying strings are `slope`, `Yintercept`, `correlation coefficient`, `predictions`, `all values`, `Quadratic Coefficient`, `a`, `b`, `Xintercepts`, `r^2`. Implement as a lookup into the cached `Results` dictionary keyed by those strings (`all values` → the whole dict). Feasible; returns `NaN`/null when no fit has been computed. |

No methods are web-impossible; both are blocked only by the absence of the series binding.

## Implementation Plan

> Ordering note: the steps below assume `Chart` and `ChartData2D` already exist in the simulator. If they do not (current state), they must be implemented first — this plan adds `Trendline` on top of that surface. Where a step needs something Chart/ChartData2D must provide, it is called out.

### simulation-capabilities.js

- Add `'Trendline'` to `SIMULATION_SUPPORTED_TYPES`. It is **not** in `SIMULATION_NONVISIBLE_TYPES` (it draws on the chart, so it's visible), but it is also **not** an independently positioned box — its render is delegated to the parent `Chart` branch (see below), so it never goes through the normal leaf layout path.
- Add a defaults block inside `buildSimulationDefaults()` (pull values straight from the spec's `designerProperties`):

```js
Trendline: {
  Visible: true,
  ChartData: '',
  Model: 'Linear',
  Color: '&H00000000',     // AI default: transparent black
  Extend: true,
  StrokeStyle: 1,          // 1=Solid (StrokeStyleEnum)
  StrokeWidth: 1.0,
},
```

- Add new prop names to `SIMULATION_VISUAL_PROPS`: `ChartData`, `Model`, `Color`, `Extend`, `StrokeStyle`, `StrokeWidth`. (`Visible` is already present.) Props not in this set are stripped before reaching the renderer, so each of these must be added or the line styling/binding will never reach the Chart render code. (`Color` is currently *not* in the visual set — confirm it is added here.)
- `isBooleanProp`: add `Extend`.
- `isNumericProp`: add `StrokeStyle` and `StrokeWidth`. (`Color` is left as its `&H…` string and resolved by `colorValue()` like every other color prop — do **not** put it through `coerceNumber`, matching how `BackgroundColor`/`TextColor` are handled.)
- `coerceSimulationValue`: no special case beyond the above. `Model` is a plain enum string (default `return value`); `ChartData` is a plain string component-name reference.
- `deriveStateFromDesignerProps`: no derived state is needed at *designer* time — the fit can only be computed once series points exist at runtime. (Optionally seed `Results: {}` / `Predictions: []` so read-only getters return empty rather than null before any data arrives.)

### SimulationComponent.svelte

`Trendline` does **not** get a standalone top-level `{#if node.type === 'Trendline'}` leaf branch with its own box — it has no geometry. Instead:

- Implement it as part of the **`Chart` render branch** (which this plan presumes exists or is being added alongside). The `Chart` branch already iterates `node.children` to draw each `ChartData2D` series onto an SVG (or `<canvas>`) plotting area with an X/Y scale. Within that same loop, for each child where `child.type === 'Trendline'`:
  1. Resolve the bound series: find the sibling `ChartData2D` child whose name === `childProps.ChartData`; read its point array.
  2. Compute the fit in a small pure helper module (see "fit helper" below) for the selected `Model`.
  3. Sample the fitted curve across the X-domain — clipped to `[xmin, xmax]` when `Extend` is false, or across the full axis range when true — map data coords → pixels using the Chart's existing scale, and emit an SVG `<path>`/`<polyline>` (or a `ctx.stroke()` on canvas) using `colorValue(childProps.Color)`, `stroke-width: numberOr(childProps.StrokeWidth, 1)`, and a `stroke-dasharray` chosen from `childProps.StrokeStyle` (1→none, 2→`6 4`, 3→`2 3`).
  4. Skip stroking entirely when `!boolValue(childProps.Visible, true)` or when `colorValue` resolves to fully transparent.
- Reuse helpers: `colorValue()` (line color), `numberOr()` (stroke width), `boolValue()` (Visible/Extend). The X/Y → pixel scale must come from the Chart branch (not Trendline-specific).
- Add a pure fit module, e.g. `src/lib/chart-trendline.js`, exporting `computeTrendline(points, model)` → `{ predictions, results, sample(x), corrCoef, rSquared, coefficients, xIntercepts, yIntercept }`. Keep regression math out of the Svelte file. This module is the reusable core for both the render path and the host's computed properties.
- Recompute reactively when the bound series points, `Model`, `Extend`, or styling change (`$:` over the resolved series + childProps), and on each recompute, signal the host to fire `Updated` (via `emitInteraction`/`emitEvent`, mirroring how other components report DOM-originated state — see ListView's selection reporting). The host then runs `Updated(results)`.

### SimulateOverlay.svelte

**None.** `Trendline` needs no dialogs, toasts, or runtime effects. `Updated` is an ordinary event routed through the existing `runEvent`/event-dispatch path; `DisconnectFromChartData`/`GetResultValue` are state reads/writes. No new overlay handler is required.

### simulation_wasm.go

A dedicated `Trendline` path is needed because almost all of its blockProperties are **computed read-only** values and there are two methods.

- **`GetProperty`** — add `case "Trendline":` returning computed fields from the cached fit. The fit results must live in `h.state[name]` (e.g. under a `Results` map plus flat keys), populated either (a) by the frontend reporting the JS-computed fit back via an interaction patch when the series changes, or (b) by porting the same OLS math into Go and computing on demand from the bound series' points held in host state. Option (a) keeps one source of truth (the JS fit module) and matches the VideoPlayer "frontend computes, patches host" precedent; prefer it. Handle: `CorrelationCoefficient`, `RSquared`, `LinearCoefficient`, `QuadraticCoefficient`, `ExponentialBase`, `ExponentialCoefficient`, `LogarithmConstant`, `LogarithmCoefficient`, `YIntercept`, `XIntercepts`, `Predictions`, `Results` — each reading the cached value, returning `runtime.NullVal()`/`NaN`/empty list when no fit exists.
- **`SetProperty`** — the generic `h.setProperty` store path suffices for `Color`, `Extend`, `Model`, `StrokeStyle`, `StrokeWidth`, `Visible`, and write-only `ChartData` (all just patch state, which the renderer reacts to and refits). Optionally, on `Model`/`Extend`/`ChartData` writes, the host can also fire nothing itself and let the frontend recompute + report `Updated` (keeps the single fit source). `ChartData` being write-only means there is no getter case for it.
- **`CallMethod`** — add `case "Trendline": return h.callTrendlineMethod(...)`:
  - `DisconnectFromChartData` → clear the stored `ChartData`/cached fit for the component (patch state so the renderer drops the line). No return value.
  - `GetResultValue` → read `args[0].AsStr()` (one of the `LOBFValues` strings) and return the matching field from the cached `Results` dict; `all values` → the whole dict; unknown/empty fit → `runtime.NullVal()`.
  - default → `h.Unsupported("method", name+"."+method)`.
- **Events** — `Updated(results)` is fired via `h.runEvent(name, "Trendline", "Updated", []runtime.Value{resultsDict})`. The natural trigger is the frontend reporting "series changed, here is the new fit"; the host then both caches the results and calls `runEvent`. (If the Go-side computes the fit, fire it directly when the bound series mutates.)
- **Effects** — none.
- Recompile with `npm run build:wasm` (`GOOS=js GOARCH=wasm`).

### design-schema-tree.js

**Already handled.** `canContainDesignComponent` already encodes `CHART_CHILD_TYPES = new Set(['ChartData2D', 'Trendline'])` and returns `parent === 'Chart'` for both (line 26 + line 69), so the designer already lets a `Trendline` live only under a `Chart`. `designTreeToInitialState()` will merge `SIMULATION_DEFAULTS.Trendline` automatically once the defaults block is added, and `unsupportedSimulationComponents()` will stop flagging it once it is in `SIMULATION_SUPPORTED_TYPES`. **No change required here.**

## Dependencies & Ordering

- **External libraries: none.** OLS regression (linear, polynomial via normal equations, exponential via log-linearization, logarithmic) is a few dozen lines of plain JS; SVG/`<canvas>` line drawing is native. No charting library is required for the trendline itself (and none should be pulled in just for this).
- **Hard prerequisite: `Chart` and `ChartData2D` must be implemented first.** `Trendline` has no coordinate system, no source data, and no surface without them:
  - `Chart` provides the plotting area, the X/Y axis ranges, and the data-coord → pixel scale the line is mapped through.
  - `ChartData2D` provides the point series the fit is computed over, and must expose a "points changed" signal so `Updated` can fire and the line can refit.
- **Recommended build order:** `Chart` → `ChartData2D` → `Trendline`. Attempting `Trendline` before the other two yields a component that computes nothing and renders nothing.

## Web-Platform Limitations & Fidelity Caveats

- **No standalone render.** Unlike most components, `Trendline` cannot be shown by itself; fidelity is entirely inherited from the `Chart`/`ChartData2D` implementation's axis scaling and surface. If the Chart's pixel mapping is approximate, the line's position is approximate.
- **Numeric divergence from MPAndroidChart.** App Inventor's real charts use MPAndroidChart's regression and curve-sampling. A from-scratch JS OLS will match analytically for linear/polynomial/log, but exponential fits via log-linearization (the common cheap method) weight residuals differently than a nonlinear least-squares fit, so `ExponentialBase`/`ExponentialCoefficient`, `r`, and `r²` can differ slightly from the device for exponential models. Curve smoothness/sampling density and rounding of the `Results` fields may also differ at the last digits.
- **`XIntercepts` shape.** AI returns `NaN`, a single number, or a list depending on root count. Reproducing the exact discriminant edge cases (e.g. tangent quadratics, cubic with one vs three real roots) requires care to match AI's exact branching; minor disagreements at degenerate inputs are possible.
- **`Color` default is transparent.** The spec default `&H00000000` has zero alpha, so a freshly-dropped Trendline is invisible until `Color` is set — faithful, but can read as "broken" to users; the simulator should honor it rather than substituting an opaque color.
- **Recompute timing.** On a device the fit updates as data streams in; in the simulator `Updated` only fires on the discrete state mutations the simulator models (series set/import, `Model`/`Extend` change). Continuous/streaming data sources (sensors, real-time imports) are not simulated, so high-frequency `Updated` cadence will not match a device.
- **`Extend` against axis bounds.** Faithful extension requires the Chart to expose its current X-axis min/max; if the Chart auto-scales, the extended line endpoints depend on that auto-scale and may shift as data changes.

## Effort Estimate

**L** (assuming `Chart` + `ChartData2D` already exist; **XL** if counted together with building that Chart subsystem, which is the real prerequisite).

Breakdown: writing a correct, well-tested OLS fit module covering all `best_fit_model` variants plus `r`/`r²`/intercepts (M on its own); integrating the line render into the Chart branch with stroke-style/extend/transparent-color handling (S–M); the host `callTrendlineMethod`, the ~12 computed read-only `GetProperty` cases, the `Updated` round-trip, and a `build:wasm` cycle (M); and the numeric-fidelity care for exponential/intercept edge cases. The capabilities/containment wiring is trivial. The dominating cost — and the honest blocker — is that none of this is useful until `Chart` and `ChartData2D` land first.
