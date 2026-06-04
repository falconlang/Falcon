# Circle Simulator Implementation Plan

## Overview

`Circle` (category **MAPS**, `nonVisible = false`) is an App Inventor *map feature*. The spec's
`helpString` is only the bare word "Circle", but the `blockProperties` make its semantics precise: a
`Circle` is a geographic disc drawn on a `Map`, centered at a (`Latitude`, `Longitude`) point with a
`Radius` **measured in meters** (not pixels), styled with `FillColor`/`FillOpacity` and
`StrokeColor`/`StrokeOpacity`/`StrokeWidth`. It carries a `Title`/`Description` infobox, can be made
`Draggable`, and exposes geodesic methods (`DistanceToFeature`, `DistanceToPoint`, `SetLocation`).
`Type` is the read-only string `"Circle"`.

- **Category:** MAPS
- **Visible:** Yes (`nonVisible: false`) — but, like `Ball` on a `Canvas`, it is **not** a top-level
  visible component. It is a geographic primitive that only has meaning rendered on a map surface.
- **Container relationship:** `Circle` is a **child of `Map`** (and of `FeatureCollection`, which is
  itself a child of `Map`). `canContainDesignComponent` already encodes this:
  `MAP_CHILD_TYPES = {Circle, FeatureCollection, LineString, Marker, Polygon, Rectangle}` and
  `FEATURE_COLLECTION_CHILD_TYPES` (which also contains `Circle`); a `Circle` is therefore only
  placeable inside a `Map` or a `FeatureCollection` (see `design-schema-tree.js` lines 24-25, 65-67).
  It is never a parent. It therefore **hard-depends on a `Map` implementation**: the circle's center
  and radius are geographic, so without a map projection (lat/lng → screen pixels at a given zoom)
  there is no correct way to place or size it.

In the simulator `Circle` currently renders as a dashed "unsupported" placeholder because both
`Circle` *and its parent `Map`* are absent from `SIMULATION_SUPPORTED_TYPES`.

## Feasibility Verdict

**Partially feasible (in a browser simulator) — and hard-blocked on a `Map` implementation.**

Drawing a styled circle in a browser is trivial (SVG `<circle>`, a CSS `border-radius: 50%` div, or a
Leaflet `L.circle`). The difficulty is not the shape — it is that **`Circle` is geographic, and the
Tensor simulator has no map surface at all**. There is no `Map` type in `SIMULATION_SUPPORTED_TYPES`,
no map library in `package.json`, and therefore no projection from (lat, lng, zoom) to screen
coordinates and no meters→pixels scale. Every visually meaningful property of a `Circle`
(`Latitude`, `Longitude`, `Radius` in meters) is defined **relative to that missing surface**. So:

1. **No projection / no tiles.** Placing the circle correctly requires converting `Latitude`/
   `Longitude` to a pixel position and `Radius` (meters) to a pixel radius at the map's current zoom.
   That math is the Web Mercator projection plus a meters-per-pixel scale — both supplied by a map
   engine. Real map *tiles* additionally require a tile provider (OpenStreetMap/Mapbox) over the
   network, which is a heavy external dependency and a network/licensing concern. Neither exists today.

2. **`Radius` in meters is not directly renderable.** Even with a fixed-zoom flat backdrop,
   meters-per-pixel varies with latitude (Mercator distortion). Honest rendering needs the projection;
   a crude approximation (treat meters as pixels at some nominal scale) will be visibly wrong at most
   zoom/latitude combinations.

3. **Drag is geographic.** `Draggable` + `Drag`/`StartDrag`/`StopDrag` move the circle's *center
   lat/lng*, which only makes sense against a pannable, projected map.

**Realistic approximation (what we will simulate), in two tiers:**

- **Tier A — depends on a `Map` host (the only honest path).** Implement `Map` first (see
  Dependencies), backed by a slippy-map library (Leaflet or MapLibre GL). Then a `Circle` is a thin
  child that renders as `L.circle([lat, lng], { radius, color, fillColor, fillOpacity, weight,
  opacity })` and gets correct geographic placement, meters-accurate radius, infobox popups, and real
  drag for free. `DistanceToPoint`/`DistanceToFeature` become exact via the Haversine formula. This is
  the *feasible* outcome, but it is gated entirely on the Map work.

- **Tier B — standalone fallback (low fidelity, only if Map is deferred).** Render the circle as a
  plain SVG/CSS disc inside the (currently unsupported) Map placeholder, with **no geographic
  meaning**: a fixed-size circle styled by the fill/stroke props, optionally showing the `Title` as a
  tooltip. `Latitude`/`Longitude`/`Radius` round-trip through the store and read back correctly from
  blocks, and `DistanceToPoint`/`DistanceToFeature` can still be computed *purely numerically* via
  Haversine (they do not need the renderer). But position and size on screen would be arbitrary and
  drag would be meaningless. This tier is only worth shipping as a stopgap; the plan's primary
  recommendation is Tier A.

Bottom line: the component itself is straightforward; its feasibility is entirely a function of
whether a projected `Map` surface exists. Without `Map`, only the non-visual, computable surface
(property round-trip + distance math) is honest; the visual is a placeholder disc, not a map circle.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
| --- | --- | --- | --- | --- |
| `Latitude` | `0` | Visual/Behavioral — circle center latitude | Tier A: pass to `L.circle` center; Tier B: stored only, not placeable | High |
| `Longitude` | `0` | Visual/Behavioral — circle center longitude | Tier A: `L.circle` center; Tier B: stored only | High |
| `Radius` | `0` | Visual/Behavioral — radius in **meters** | Tier A: `L.circle` `radius` (meters, projected). Tier B: cannot honor meters; fall back to a fixed pixel disc. Note default `0` = invisible until set | High |
| `FillColor` | `&HFFFF0000` (red) | Visual — interior fill | `colorValue()` → SVG `fill` / Leaflet `fillColor` | High |
| `FillOpacity` | `1.0` | Visual — interior alpha 0..1 | clamp 0..1 → `fill-opacity` / Leaflet `fillOpacity` | High |
| `StrokeColor` | `&HFF000000` (black) | Visual — outline color | `colorValue()` → SVG `stroke` / Leaflet `color` | High |
| `StrokeOpacity` | `1.0` | Visual — outline alpha 0..1 | clamp 0..1 → `stroke-opacity` / Leaflet `opacity` | Medium |
| `StrokeWidth` | `1` | Visual — outline width (px) | `stroke-width` / Leaflet `weight` | Medium |
| `Title` | `""` | Behavioral — infobox heading | Popup/tooltip content (bound title) | Medium |
| `Description` | `""` | Behavioral — infobox body | Popup body text | Medium |
| `EnableInfobox` | `False` | Behavioral — whether tap opens the infobox | gate the click→popup behavior | Medium |
| `Draggable` | `False` | Behavioral — long-press drag to move center | Tier A: Leaflet `draggable` (needs Leaflet.Path drag plugin or a Marker proxy); Tier B: not meaningful | Low |
| `Visible` | `True` | Visual | generic `Visible` handling (already wired) | High |

Notes:
- Defaults are pulled verbatim from the spec's `designerProperties`. `FillColor`/`StrokeColor` are AI
  ARGB hex (`&HFFFF0000`, `&HFF000000`) — `colorValue()` already parses that format. **Do not** route
  the colors through the opacity props blindly: AI keeps fill color and `FillOpacity` separate (the
  ARGB alpha is typically `FF`), so render `fill = colorValue(FillColor)` and apply `FillOpacity` as a
  *separate* opacity channel.
- `Type` is read-only and always `"Circle"` — handle it as a computed `GetProperty` constant, not a
  stored value.
- `Radius` default `0` means a freshly added circle is **zero-size / invisible** until a `Radius` is
  set (designer or blocks) — expected AI behavior, worth noting so it does not look "broken".
- In **Tier B** the geographic props (`Latitude`/`Longitude`/`Radius`) are *stored and readable* but
  cannot be honored visually; flag that clearly rather than faking a position.

## Events

| Event | Args | How/when the simulator fires it |
| --- | --- | --- |
| `Click` | (none) | Tap on the circle. Feasible in both tiers via a DOM/`Leaflet` click → `emitEvent(node.name, 'Click')`. If `EnableInfobox`, also open the popup. |
| `LongClick` | (none) | Long-press on the circle **when `Draggable` is false** (per spec). Reuse the existing 600ms long-press timer pattern (see `pointerDown(true)`/`longClickTimer`). Feasible. |
| `StartDrag` | (none) | Fires when a drag gesture begins. **Tier A only** (needs a pannable projected map + draggable path); fire on Leaflet `dragstart`. **Not meaningful in Tier B.** |
| `Drag` | (none) | Fires repeatedly while dragging. **Tier A only**, on Leaflet `drag`; update center lat/lng live. **Not meaningful in Tier B.** |
| `StopDrag` | (none) | Fires when the drag ends. **Tier A only**, on Leaflet `dragend`, committing the new center lat/lng. **Not meaningful in Tier B.** |

All five events have **no params** (confirmed from the spec) — no `EVENT_ARGS` entry is needed (only
positional-arg events like Slider/Canvas need one). The drag trio is honest only once a real projected
map exists; in Tier B they cannot fire because there is nothing to drag against.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
| --- | --- | --- |
| `SetLocation` | `SetLocation(latitude, longitude)` → void | Feasible. Set stored `Latitude`/`Longitude`, patch state. Tier A re-positions the `L.circle` center live; Tier B updates readable state only. |
| `DistanceToPoint` | `DistanceToPoint(latitude, longitude, centroid)` → number | Feasible **numerically in both tiers** (no renderer needed). Compute great-circle (Haversine) distance in meters from the circle's center to the point; with `centroid = false` subtract `Radius` to get edge distance (clamp ≥ 0), with `centroid = true` use center-to-point. |
| `DistanceToFeature` | `DistanceToFeature(mapFeature, centroids)` → number | Partially feasible — **iff the target feature exposes a readable center** (i.e. another implemented `Circle`/`Marker`). Read the target's `Latitude`/`Longitude` from state, Haversine to it; honor `centroids` like `DistanceToPoint`. If the target type is unimplemented/unreadable: `h.Unsupported("method", "Circle.DistanceToFeature")` and return `0`. |
| `ShowInfobox` | `ShowInfobox()` → void | Tier A: open the Leaflet popup (even if `EnableInfobox` is false, per spec) via a `component-action` effect (`action: "show-infobox"`). Tier B: emit a toast-style notice with `Title`/`Description`, or `Unsupported` if no infobox surface. |
| `HideInfobox` | `HideInfobox()` → void | Tier A: close the popup via a `component-action` effect (`action: "hide-infobox"`); no-op if not shown (per spec). Tier B: dismiss the notice or `Unsupported`. |

`DistanceToPoint`/`DistanceToFeature` are the genuinely useful, fully-honest methods here because
geodesic math needs only the stored center + radius, independent of whether the map renders.

## Implementation Plan

### simulation-capabilities.js
- **`SIMULATION_SUPPORTED_TYPES`:** add `'Circle'`. (Do **not** add to `SIMULATION_NONVISIBLE_TYPES` —
  it is visible.) This is meaningful only once `'Map'` is also supported; otherwise the parent stays a
  placeholder and the child has nowhere to render.
- **`buildSimulationDefaults()`** — add a defaults block (Circle has no `Width`/`Height`/font props, so
  it does **not** spread `COMMON_VISIBLE_PROPS`):
  ```js
  Circle: {
    Visible: true,
    Latitude: 0,
    Longitude: 0,
    Radius: 0,
    FillColor: '&HFFFF0000',
    FillOpacity: 1.0,
    StrokeColor: '&HFF000000',
    StrokeOpacity: 1.0,
    StrokeWidth: 1,
    Title: '',
    Description: '',
    EnableInfobox: false,
    Draggable: false,
  },
  ```
- **`SIMULATION_VISUAL_PROPS`:** add the new names `'Latitude'`, `'Longitude'`, `'Radius'`,
  `'FillColor'`, `'FillOpacity'`, `'StrokeColor'`, `'StrokeOpacity'`, `'StrokeWidth'`,
  `'Description'`, `'EnableInfobox'`, `'Draggable'`. (`Title` and `Visible` are already present.)
  Without these the props are **stripped** before reaching the renderer.
- **`isBooleanProp`:** add `'EnableInfobox'`, `'Draggable'`.
- **`isNumericProp`:** add `'Latitude'`, `'Longitude'`, `'Radius'`, `'FillOpacity'`, `'StrokeOpacity'`,
  `'StrokeWidth'`. (`FillColor`/`StrokeColor` stay strings/ARGB — handled by `colorValue` — do **not**
  make them numeric, even though `blockProperties` type them `number`.)
- **`coerceSimulationValue`:** no new special case — numeric props go through `coerceNumber`, booleans
  through `coerceBoolean`; colors fall through as strings. (Optionally clamp `FillOpacity`/
  `StrokeOpacity` to 0..1 in the renderer rather than at coercion.)
- **`deriveStateFromDesignerProps`:** no derived state required.

### SimulationComponent.svelte
A standalone branch is only honest **inside a projected Map surface**. Two cases:

- **Preferred (Tier A):** the `Circle` is rendered *by the Map branch via the map library*, not by a
  generic `<svelte:self>` child div — Leaflet owns the overlay layer. In that model the Circle branch
  here may not render DOM at all (the Map registers an `L.circle` per child feature and binds its
  props), or renders a tiny invisible anchor. The wiring (click/long-press/drag) is attached to the
  Leaflet layer in the Map component, dispatching through `emitEvent`. This keeps the Circle's geometry
  correct.

- **Fallback (Tier B):** add a leaf branch rendering a styled SVG disc with **no geographic meaning**
  (clearly a stand-in). Sketch:
  ```svelte
  {:else if node.type === 'Circle'}
    <svg class="sim-mapfeature sim-circle" class:sim-unsupported={unsupportedHere}
         data-sim-component={node.name} viewBox="0 0 80 80" width="80" height="80"
         on:click={() => emitEvent(node.name, 'Click')}
         on:pointerdown={() => pointerDown(true)}
         on:pointerup={pointerUp}
         on:pointercancel={clearLongClick}>
      <title>{props.Title || node.name}</title>
      <circle cx="40" cy="40" r="36"
              fill={colorValue(props.FillColor, '#f00')}
              fill-opacity={clamp01(numberOr(props.FillOpacity, 1))}
              stroke={colorValue(props.StrokeColor, '#000')}
              stroke-opacity={clamp01(numberOr(props.StrokeOpacity, 1))}
              stroke-width={numberOr(props.StrokeWidth, 1)} />
    </svg>
  {/if}
  ```
- **Reuse:** `colorValue`, `numberOr`, `boolValue` already exist; the Button `pointerDown(true)` /
  `longClickTimer` / `consumeLongClick` long-press machinery is the closest reference for `LongClick`.
  A new `clamp01(n)` helper is trivial. Do **not** use `baseStyle()` (it injects size/background/font/
  alignment rules irrelevant to a map feature).
- **CSS:** minimal `.sim-circle { display: block; }`; the SVG attributes drive appearance.
- Honest note: in Tier B the disc is a fixed 80px stand-in; `Latitude`/`Longitude`/`Radius` do **not**
  affect it. Ship Tier A whenever Map exists.

### SimulateOverlay.svelte
**Needed only for infobox surfacing (and only modestly).** `ShowInfobox`/`HideInfobox` (and a
`Click` with `EnableInfobox`) emit `component-action` effects (`action: "show-infobox"` /
`"hide-infobox"`) that the overlay can route — in Tier A to the Leaflet popup hosted by the Map, in
Tier B to a transient notice using the existing Notifier-style toast/dialog path the overlay already
implements. If infoboxes are deferred, **none required** — `Click`/`LongClick` flow through the normal
`emitEvent`→`runEvent` path with no overlay involvement.

### simulation_wasm.go
Add a `case "Circle":` to `CallMethod` dispatching to a new `callCircleMethod`, plus a computed
`GetProperty` for `Type`. Most properties ride the **generic store path** already.
- **`GetProperty`:** add a `case "Circle":` returning the constant `runtime.StrVal("Circle")` for
  `property == "Type"` (read-only). All other reads (`Latitude`/`Longitude`/`Radius`/colors/opacities/
  `StrokeWidth`/`Title`/`Description`/`EnableInfobox`/`Draggable`/`Visible`) come from the generic
  store read at the top of `GetProperty` — no extra cases.
- **`SetProperty`:** generic `h.setProperty` suffices for every read-write prop; each patches state and
  (Tier A) re-styles/repositions the Leaflet layer live. No special derivation. (`Type` is read-only —
  if a `SetProperty("Type")` arrives, ignore or `h.Unsupported`.)
- **`CallMethod` → `callCircleMethod`:**
  - `SetLocation(lat, lng)`: `h.setProperty(name, "Latitude", args[0]); h.setProperty(name,
    "Longitude", args[1])`.
  - `DistanceToPoint(lat, lng, centroid)`: read current center, Haversine in meters; if `!centroid`
    subtract `Radius`, clamp ≥ 0; return `runtime.NumberVal(d)`. (Pure math — works regardless of
    renderer.)
  - `DistanceToFeature(feature, centroids)`: resolve the target's `Latitude`/`Longitude` from
    `h.state` if readable, Haversine; else `h.Unsupported("method", "Circle.DistanceToFeature")` and
    return `runtime.NumberVal(0)`.
  - `ShowInfobox()` / `HideInfobox()`: append a `component-action` effect (`"show-infobox"` /
    `"hide-infobox"`); no-op semantics for hide-when-not-shown.
  - `default`: `h.Unsupported("method", componentName+"."+method)`.
- **Events:** `Click`/`LongClick` dispatched from the renderer/map layer via `emitEvent`→`runEvent` —
  no host-side firing. `StartDrag`/`Drag`/`StopDrag` are fired by the Map's Leaflet drag handlers
  (Tier A) and simply never fire in Tier B.

### design-schema-tree.js
**Already handled.** `MAP_CHILD_TYPES` and `FEATURE_COLLECTION_CHILD_TYPES` both include `'Circle'`,
and `canContainDesignComponent` returns `true` for `Map`→`Circle` and `FeatureCollection`→`Circle`
(lines 65-67) — containment validation is correct with **no edit**. `designTreeToInitialState` will
merge `SIMULATION_DEFAULTS.Circle` once the defaults block exists, and
`unsupportedSimulationComponents` stops flagging `Circle` once it is added to
`SIMULATION_SUPPORTED_TYPES`. No code change in this file. (Note: it will *still* be flagged in
practice until `Map` is also supported, since the parent placeholder swallows children.)

## Dependencies & Ordering

- **Hard prerequisite: `Map` must be implemented first.** A `Circle` has no honest visual without a
  projected map surface (lat/lng→pixels, meters→pixels, pan/zoom). Implement `Map`, then the feature
  children (`Circle`, `Marker`, `LineString`, `Polygon`, `Rectangle`) become thin styled overlays.
- **External library:** a slippy-map engine — **Leaflet** (lightweight, `L.circle` supports a
  meters radius natively, popups, drag plugins) or **MapLibre GL** (heavier, WebGL). Leaflet is the
  pragmatic choice. This is a **new runtime dependency** to add to `package.json`; the simulator
  currently has none. Map *tiles* additionally need a tile provider (OpenStreetMap is free but has a
  usage policy; offline/blank-tile mode avoids the network entirely and is the safest default for a
  local IDE simulator).
- **Sibling features:** `Marker`/`LineString`/`Polygon`/`Rectangle` share the same overlay pattern;
  `DistanceToFeature` reaches full fidelity only once sibling features expose readable centers.
  `FeatureCollection` is a grouping container (also a Map child) and should land alongside.
- **Ordering:** `Map` → `Circle` (+ siblings) → `FeatureCollection`. Without Map, only the non-visual
  Tier B subset (property round-trip + Haversine distance methods) is shippable, and it is low value.

## Web-Platform Limitations & Fidelity Caveats

- **No map without a library (primary caveat):** the simulator has no projection/tiles today; a
  `Circle` cannot be placed or sized geographically until `Map` + Leaflet/MapLibre land. Until then
  the disc is a non-geographic stand-in.
- **`Radius` is in meters:** correct rendering requires the projection's meters-per-pixel scale, which
  varies with latitude (Web Mercator distortion). A pixel-radius fallback is visibly inaccurate.
- **Tiles are a network/licensing dependency:** real basemap tiles require an external provider over
  the network; offline/blank-canvas mode is the safe default and diverges from the on-device map look.
- **Drag fidelity:** Leaflet has no first-class draggable-circle; it needs a path-drag plugin or a
  Marker proxy at the center, so `Draggable`/`StartDrag`/`Drag`/`StopDrag` are approximations even in
  Tier A and entirely inert in Tier B.
- **Infobox styling:** `Title`/`Description` map to a Leaflet popup, which will not match the native
  Android infobox chrome exactly.
- **`DistanceToFeature` coupling:** exact only when the *target* feature is implemented and its center
  is readable from simulator state; otherwise it is `Unsupported`/`0`.
- **Color vs opacity separation:** AI keeps `FillColor`/`StrokeColor` ARGB separate from
  `FillOpacity`/`StrokeOpacity`; the renderer must apply opacity as a distinct channel, not fold it
  into the color alpha, to match device output.

## Effort Estimate

**L** — the `Circle` *itself* is S/M (capabilities entries, defaults, a styled SVG/Leaflet overlay,
`SetLocation`, the `Type` constant, and the pure-math `DistanceToPoint`/`DistanceToFeature` host
methods are all straightforward). The effort is dominated by the **hard `Map` prerequisite**:
integrating a slippy-map library (new dependency), wiring a projected surface with pan/zoom, choosing a
tile/offline strategy, and routing feature-child rendering + drag/infobox effects through it — that
Map work is **L/XL** on its own and must precede a faithful Circle. As a standalone Tier-B stopgap (no
Map) the Circle is **S**, but low fidelity and not recommended as the end state.
