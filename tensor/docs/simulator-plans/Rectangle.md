# Rectangle Simulator Implementation Plan

## Overview

`Rectangle` (category **MAPS**, `nonVisible = false`) is an App Inventor *map feature*. Per the spec's
`helpString`: "Rectangles are polygons with fixed latitudes and longitudes for the north, south, east,
and west boundaries. Moving a vertex of the Rectangle updates the appropriate edges accordingly." In
other words a `Rectangle` is an *axis-aligned geographic box* drawn on a `Map`, defined by four edge
coordinates — `NorthLatitude`, `SouthLatitude`, `EastLongitude`, `WestLongitude` (all decimal degrees)
— rather than by a center + radius (`Circle`) or an arbitrary vertex list (`Polygon`). It is styled by
`FillColor`/`FillOpacity` and `StrokeColor`/`StrokeOpacity`/`StrokeWidth`, carries a `Title`/`Description`
infobox (gated by `EnableInfobox`), can be `Draggable`, and exposes geodesic methods (`Bounds`,
`Center`, `DistanceToFeature`, `DistanceToPoint`, `SetCenter`). `Type` is the read-only string
`"Rectangle"`.

- **Category:** MAPS
- **Visible:** Yes (`nonVisible: false`) — but, like `Circle`/`Ball`, it is **not** a top-level visible
  component. It is a geographic primitive that only has meaning rendered on a map surface.
- **Container relationship:** `Rectangle` is a **child of `Map`** (and of `FeatureCollection`, which is
  itself a child of `Map`). `canContainDesignComponent` already encodes this:
  `MAP_CHILD_TYPES = {Circle, FeatureCollection, LineString, Marker, Polygon, Rectangle}` and
  `FEATURE_COLLECTION_CHILD_TYPES` (which also contains `Rectangle`); a `Rectangle` is therefore only
  placeable inside a `Map` or a `FeatureCollection` (see `design-schema-tree.js` lines 24-25, 65-67). It
  is never a parent. It **hard-depends on a `Map` implementation**: the four edges are geographic, so
  without a map projection (lat/lng → screen pixels at a given zoom) there is no correct way to place or
  size the box.

In the simulator `Rectangle` currently renders as a dashed "unsupported" placeholder because both
`Rectangle` *and its parent `Map`* are absent from `SIMULATION_SUPPORTED_TYPES`.

## Feasibility Verdict

**Partially feasible (in a browser simulator) — and hard-blocked on a `Map` implementation.**

Drawing a styled rectangle in a browser is trivial (SVG `<rect>`, a CSS box, or a Leaflet
`L.rectangle`). The difficulty is not the shape — it is that **`Rectangle` is geographic, and the
Tensor simulator has no map surface at all**. There is no `Map` type in `SIMULATION_SUPPORTED_TYPES`, no
map library in `package.json`, and therefore no projection from (lat, lng, zoom) to screen coordinates.
Every visually meaningful property (`NorthLatitude`, `SouthLatitude`, `EastLongitude`, `WestLongitude`)
is defined **relative to that missing surface**. So:

1. **No projection / no tiles.** Placing the box correctly requires converting each edge latitude/
   longitude to a pixel coordinate at the map's current zoom. That math is the Web Mercator projection —
   supplied by a map engine. Real map *tiles* additionally require a tile provider (OSM/USGS) over the
   network, a heavy dependency and a network/licensing concern. Neither exists today.

2. **The box is geographic, not pixel-sized.** Unlike a fixed-pixel SVG `<rect>`, a `Rectangle`'s
   on-screen size depends on the zoom level and on Mercator latitude distortion (a 1°-tall box is taller
   in screen pixels near the equator than near the poles). Honest rendering needs the projection; a
   pixel-box fallback is visibly wrong at most zoom/latitude combinations.

3. **Drag is geographic.** `Draggable` + `Drag`/`StartDrag`/`StopDrag` translate all four edges
   together, which only makes sense against a pannable, projected map.

4. **Degenerate / wrapped defaults.** Every edge defaults to `0`, so a freshly added `Rectangle` is a
   **zero-area box at (0,0)** (on the equator/prime-meridian) — invisible until the edges are set. Also,
   AI does not validate that `North > South` or `West < East` at construction; a box with crossed edges
   or one straddling the antimeridian needs care (Leaflet normalizes most of this, but it is a fidelity
   note).

**Realistic approximation (what we will simulate), in two tiers** (identical strategy to the `Circle`
plan, since they share the Map host):

- **Tier A — depends on a `Map` host (the only honest path).** Implement `Map` first (see
  Dependencies), backed by Leaflet. Then a `Rectangle` is a thin child rendered as
  `L.rectangle([[SouthLatitude, WestLongitude], [NorthLatitude, EastLongitude]], { color, fillColor,
  fillOpacity, weight, opacity })`, getting correct geographic placement, zoom-accurate size, infobox
  popups, and (with a path-drag plugin) real drag for free. `Bounds`/`Center`/`DistanceToPoint`/
  `DistanceToFeature` become exact. This is the *feasible* outcome, gated entirely on the Map work.

- **Tier B — standalone fallback (low fidelity, only if Map is deferred).** Render the rectangle as a
  plain SVG/CSS box inside the (currently unsupported) Map placeholder, with **no geographic meaning**: a
  fixed-size box styled by the fill/stroke props, optionally showing `Title` as a tooltip. The four edge
  coordinates round-trip through the store and read back correctly from blocks, and `Bounds`/`Center`/
  `DistanceToPoint`/`DistanceToFeature` are computed *purely numerically* (they do not need the
  renderer). But position and size on screen are arbitrary and drag is meaningless. Tier B is only a
  stopgap; the primary recommendation is Tier A.

Bottom line: the component itself is straightforward; feasibility is entirely a function of whether a
projected `Map` surface exists. Without `Map`, only the non-visual, computable surface (property
round-trip + `Bounds`/`Center`/distance math) is honest; the visual is a placeholder box, not a map
rectangle.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
| --- | --- | --- | --- | --- |
| `NorthLatitude` | `0` | Visual/Behavioral — north edge (decimal degrees) | Tier A: `[N,E]` corner of `L.rectangle` bounds; Tier B: stored only, not placeable | High |
| `SouthLatitude` | `0` | Visual/Behavioral — south edge | Tier A: `[S,W]` corner of bounds; Tier B: stored only | High |
| `EastLongitude` | `0` | Visual/Behavioral — east edge | Tier A: `[N,E]` corner of bounds; Tier B: stored only | High |
| `WestLongitude` | `0` | Visual/Behavioral — west edge | Tier A: `[S,W]` corner of bounds; Tier B: stored only | High |
| `FillColor` | `&HFFFF0000` (red) | Visual — interior fill | `colorValue()` → SVG `fill` / Leaflet `fillColor` | High |
| `FillOpacity` | `1.0` | Visual — interior alpha 0..1 | clamp 0..1 → `fill-opacity` / Leaflet `fillOpacity` | High |
| `StrokeColor` | `&HFF000000` (black) | Visual — outline color | `colorValue()` → SVG `stroke` / Leaflet `color` | High |
| `StrokeOpacity` | `1.0` | Visual — outline alpha 0..1 | clamp 0..1 → `stroke-opacity` / Leaflet `opacity` | Medium |
| `StrokeWidth` | `1` | Visual — outline width (px) | `stroke-width` / Leaflet `weight` | Medium |
| `Title` | `""` | Behavioral — infobox heading | Popup/tooltip content (bound title) | Medium |
| `Description` | `""` | Behavioral — infobox body | Popup body text | Medium |
| `EnableInfobox` | `False` | Behavioral — whether tap opens the infobox | gate the click→popup behavior | Medium |
| `Draggable` | `False` | Behavioral — long-press drag to move the box | Tier A: Leaflet path-drag plugin / proxy; Tier B: not meaningful | Low |
| `Visible` | `True` | Visual | generic `Visible` handling (already wired) | High |
| `Type` | (block-only, read-only `"Rectangle"`) | Behavioral — feature type string | Computed `GetProperty` constant `"Rectangle"`; never stored | Low |

Notes:
- Defaults are pulled verbatim from the spec's `designerProperties`. `FillColor`/`StrokeColor` are AI
  ARGB hex (`&HFFFF0000`, `&HFF000000`) — `colorValue()` already parses that format. **Do not** fold the
  opacity props into the color alpha: AI keeps fill/stroke color and `FillOpacity`/`StrokeOpacity`
  separate (the ARGB alpha is typically `FF`), so render `fill = colorValue(FillColor)` and apply
  `FillOpacity` as a *separate* opacity channel.
- All four edge defaults are `0`, so a freshly added `Rectangle` is a **zero-area box at the
  equator/prime-meridian** — invisible until edges are set (expected AI behavior; worth noting so it does
  not look "broken").
- `Type` is read-only and always `"Rectangle"` — handle it as a computed `GetProperty` constant, not a
  stored value (mirror the `Circle` `Type` handling).
- In **Tier B** the geographic edges are *stored and readable* but cannot be honored visually; flag that
  clearly rather than faking a position/size.

## Events

| Event | Args | How/when the simulator fires it |
| --- | --- | --- |
| `Click` | (none) | Tap on the rectangle. Feasible in both tiers via DOM/Leaflet click → `emitEvent(node.name, 'Click')`. If `EnableInfobox`, also open the popup. |
| `LongClick` | (none) | Long-press on the rectangle **when `Draggable` is false** (per spec). Reuse the existing 600ms long-press machinery (`pointerDown(true)` / `longClickTimer` / `consumeLongClick`). Feasible. |
| `StartDrag` | (none) | Fires when a drag gesture begins. **Tier A only** (needs a pannable projected map + draggable path); fire on Leaflet `dragstart`. **Not meaningful in Tier B.** |
| `Drag` | (none) | Fires repeatedly while dragging. **Tier A only**, on Leaflet `drag`; update all four edges live. **Not meaningful in Tier B.** |
| `StopDrag` | (none) | Fires when the drag ends. **Tier A only**, on Leaflet `dragend`, committing the new edge coordinates. **Not meaningful in Tier B.** |

All five events have **no params** (confirmed from the spec) — no `EVENT_ARGS` entry is needed (only
positional-arg events like Slider/Canvas need one). The drag trio is honest only once a real projected
map exists; in Tier B they cannot fire because there is nothing to drag against. No event is blocked by a
device sensor.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
| --- | --- | --- |
| `Bounds` | `Bounds()` → list | Feasible **numerically in both tiers** (no renderer needed). Return `((NorthLatitude WestLongitude) (SouthLatitude EastLongitude))` — a list of two `(lat lng)` pairs, per the spec's `((North West) (South East))` format. Pure read of stored edges. |
| `Center` | `Center()` → list | Feasible numerically in both tiers. Return `((North+South)/2, (East+West)/2)` as a `(Latitude Longitude)` list. (Antimeridian-straddling boxes need longitude wrapping; document as a fidelity edge case.) |
| `SetCenter` | `SetCenter(latitude, longitude)` → void | Feasible. Compute current half-height `(N-S)/2` and half-width `(E-W)/2`, then set `North = lat + halfH`, `South = lat - halfH`, `East = lng + halfW`, `West = lng - halfW` (preserving size, per the spec). Patch all four edges. Tier A re-positions the `L.rectangle`; Tier B updates readable state only. |
| `DistanceToPoint` | `DistanceToPoint(latitude, longitude, centroid)` → number | Feasible **numerically in both tiers**. With `centroid = true`, Haversine from the box `Center` to the point. With `centroid = false`, distance from the *nearest point on the rectangle's edge/interior* to the target (0 if the point is inside the box); a correct edge-distance is more involved than Circle's `center − radius`, but is pure geodesic math. |
| `DistanceToFeature` | `DistanceToFeature(mapFeature, centroids)` → number | Partially feasible — **iff the target feature exposes a readable center/geometry** (another implemented `Rectangle`/`Circle`/`Marker`). With `centroids = true`, Haversine between the two centroids (read the target's coords from `h.state`). With `centroids = false`, nearest-geometry distance (involved). If the target type is unimplemented/unreadable: `h.Unsupported("method", "Rectangle.DistanceToFeature")` and return `0`. |
| `ShowInfobox` | `ShowInfobox()` → void | Tier A: open the Leaflet popup (even if `EnableInfobox` is false, per spec) via a `component-action` effect (`action: "show-infobox"`). Tier B: emit a toast-style notice with `Title`/`Description`, or `Unsupported` if no infobox surface. |
| `HideInfobox` | `HideInfobox()` → void | Tier A: close the popup via a `component-action` effect (`action: "hide-infobox"`); no-op if not shown (per spec). Tier B: dismiss the notice or `Unsupported`. |

`Bounds`/`Center`/`SetCenter`/`DistanceToPoint` are the genuinely useful, fully-honest methods here
because they need only the stored edges, independent of whether the map renders. `DistanceToFeature` is
the only one coupled to other components existing.

## Implementation Plan

### simulation-capabilities.js
- **`SIMULATION_SUPPORTED_TYPES`:** add `'Rectangle'`. (Do **not** add to `SIMULATION_NONVISIBLE_TYPES`
  — it is visible.) Meaningful only once `'Map'` is also supported; otherwise the parent stays a
  placeholder and the child has nowhere to render.
- **`buildSimulationDefaults()`** — add a defaults block (Rectangle has no `Width`/`Height`/font props,
  so it does **not** spread `COMMON_VISIBLE_PROPS`):
  ```js
  Rectangle: {
    Visible: true,
    NorthLatitude: 0,
    SouthLatitude: 0,
    EastLongitude: 0,
    WestLongitude: 0,
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
- **`SIMULATION_VISUAL_PROPS`:** add the new names `'NorthLatitude'`, `'SouthLatitude'`,
  `'EastLongitude'`, `'WestLongitude'`. (`FillColor`, `FillOpacity`, `StrokeColor`, `StrokeOpacity`,
  `StrokeWidth`, `Description`, `EnableInfobox`, `Draggable` are shared with `Circle` and should already
  be present if `Circle` landed first — add any that are still missing; `Title`/`Visible` are already
  present.) Without these the props are **stripped** before reaching the renderer.
- **`isBooleanProp`:** add `'EnableInfobox'`, `'Draggable'` (shared with `Circle`; add only if absent).
- **`isNumericProp`:** add `'NorthLatitude'`, `'SouthLatitude'`, `'EastLongitude'`, `'WestLongitude'`,
  plus `'FillOpacity'`, `'StrokeOpacity'`, `'StrokeWidth'` (shared with `Circle`; add only if absent).
  (`FillColor`/`StrokeColor` stay strings/ARGB — handled by `colorValue` — do **not** make them numeric,
  even though `blockProperties` type them `number`.)
- **`coerceSimulationValue`:** no new special case — numeric props go through `coerceNumber`, booleans
  through `coerceBoolean`; colors fall through as strings. (Optionally clamp `FillOpacity`/
  `StrokeOpacity` to 0..1 in the renderer rather than at coercion.)
- **`deriveStateFromDesignerProps`:** no derived state required.

### SimulationComponent.svelte
A standalone branch is only honest **inside a projected Map surface**. Two cases (mirroring `Circle`):

- **Preferred (Tier A):** the `Rectangle` is rendered *by the Map branch via Leaflet*, not by a generic
  `<svelte:self>` child div — Leaflet owns the overlay layer. In that model the Rectangle branch here may
  render no DOM (the Map registers an `L.rectangle` per child feature and binds its props/events), or a
  tiny invisible anchor. Click/long-press/drag wiring attaches to the Leaflet layer in the Map component
  and dispatches through `emitEvent`. This keeps geometry correct.

- **Fallback (Tier B):** add a leaf branch rendering a styled SVG box with **no geographic meaning**
  (clearly a stand-in). Sketch:
  ```svelte
  {:else if node.type === 'Rectangle'}
    <svg class="sim-mapfeature sim-rectangle" class:sim-unsupported={unsupportedHere}
         data-sim-component={node.name} viewBox="0 0 100 70" width="100" height="70"
         on:click={() => emitEvent(node.name, 'Click')}
         on:pointerdown={() => pointerDown(true)}
         on:pointerup={pointerUp}
         on:pointercancel={clearLongClick}>
      <title>{props.Title || node.name}</title>
      <rect x="6" y="6" width="88" height="58"
            fill={colorValue(props.FillColor, '#f00')}
            fill-opacity={clamp01(numberOr(props.FillOpacity, 1))}
            stroke={colorValue(props.StrokeColor, '#000')}
            stroke-opacity={clamp01(numberOr(props.StrokeOpacity, 1))}
            stroke-width={numberOr(props.StrokeWidth, 1)} />
    </svg>
  {/if}
  ```
- **Reuse:** `colorValue`, `numberOr`, `boolValue` already exist; the Button `pointerDown(true)` /
  `longClickTimer` / `consumeLongClick` long-press machinery is the closest reference for `LongClick`. A
  `clamp01(n)` helper (introduced for `Circle`) is shared. Do **not** use `baseStyle()` (it injects
  size/background/font/alignment rules irrelevant to a map feature).
- **CSS:** minimal `.sim-rectangle { display: block; }`; the SVG attributes drive appearance.
- Honest note: in Tier B the box is a fixed stand-in; the four edge coordinates do **not** affect it.
  Ship Tier A whenever Map exists.

### SimulateOverlay.svelte
**Needed only for infobox surfacing (and only modestly).** `ShowInfobox`/`HideInfobox` (and a `Click`
with `EnableInfobox`) emit `component-action` effects (`action: "show-infobox"` / `"hide-infobox"`) the
overlay routes — in Tier A to the Leaflet popup hosted by the Map, in Tier B to a transient notice using
the existing Notifier-style toast/dialog path. If infoboxes are deferred, **none required** —
`Click`/`LongClick` flow through the normal `emitEvent`→`runEvent` path with no overlay involvement.

### simulation_wasm.go
Add a `case "Rectangle":` to `CallMethod` dispatching to a new `callRectangleMethod`, plus a computed
`GetProperty` for `Type`. Most properties ride the **generic store path** already.
- **`GetProperty`:** add a `case "Rectangle":` returning the constant `runtime.StrVal("Rectangle")` for
  `property == "Type"` (read-only). All other reads (the four edges, colors/opacities, `StrokeWidth`,
  `Title`/`Description`/`EnableInfobox`/`Draggable`/`Visible`) come from the generic store read at the
  top of `GetProperty` — no extra cases.
- **`SetProperty`:** generic `h.setProperty` suffices for every read-write prop; each patches state and
  (Tier A) re-positions/re-styles the Leaflet layer live. No special derivation. (`Type` is read-only —
  if a `SetProperty("Type")` arrives, ignore or `h.Unsupported`.)
- **`CallMethod` → `callRectangleMethod`:**
  - `Bounds()`: build `runtime.ListVal([ ListVal([North, West]), ListVal([South, East]) ])` from the
    stored edges.
  - `Center()`: `runtime.ListVal([ NumVal((N+S)/2), NumVal((E+W)/2) ])`.
  - `SetCenter(lat, lng)`: read current edges, compute half-height/half-width, set the four edges to
    keep size centered on `(lat, lng)`, patch state.
  - `DistanceToPoint(lat, lng, centroid)`: `centroid=true` → Haversine center→point; `centroid=false` →
    nearest-point-on-box distance (0 if inside). Pure math; works regardless of renderer.
  - `DistanceToFeature(feature, centroids)`: resolve the target's coords/geometry from `h.state` if
    readable, Haversine (centroid or nearest); else `h.Unsupported("method", "Rectangle.DistanceToFeature")`
    and return `runtime.NumVal(0)`.
  - `ShowInfobox()` / `HideInfobox()`: append a `component-action` effect (`"show-infobox"` /
    `"hide-infobox"`); no-op semantics for hide-when-not-shown.
  - `default`: `h.Unsupported("method", componentName+"."+method)`.
- **Events:** `Click`/`LongClick` dispatched from the renderer/map layer via `emitEvent`→`runEvent` — no
  host-side firing. `StartDrag`/`Drag`/`StopDrag` are fired by the Map's Leaflet drag handlers (Tier A)
  and simply never fire in Tier B.
- Recompile with `npm run build:wasm`.

### design-schema-tree.js
**Already handled.** `MAP_CHILD_TYPES` and `FEATURE_COLLECTION_CHILD_TYPES` both include `'Rectangle'`,
and `canContainDesignComponent` returns `true` for `Map`→`Rectangle` and `FeatureCollection`→`Rectangle`
(lines 65-67) — containment validation is correct with **no edit**. `designTreeToInitialState` will merge
`SIMULATION_DEFAULTS.Rectangle` once the defaults block exists, and `unsupportedSimulationComponents`
stops flagging `Rectangle` once it is added to `SIMULATION_SUPPORTED_TYPES`. No code change in this file.
(Note: it will *still* be flagged in practice until `Map` is also supported, since the parent placeholder
swallows children.)

## Dependencies & Ordering

- **Hard prerequisite: `Map` must be implemented first** (see `Map.md`, effort XL). A `Rectangle` has no
  honest visual without a projected map surface (lat/lng→pixels, zoom, pan). Implement `Map`, then the
  feature children become thin styled overlays.
- **External library:** **Leaflet** — `L.rectangle([[south, west], [north, east]], options)` maps almost
  1:1 onto the four-edge model, with popups and (via a path-drag plugin) drag. This is the new runtime
  dependency added by the Map work; Rectangle reuses it.
- **Sibling features:** `Rectangle` shares the overlay/long-press/infobox/distance pattern with `Circle`,
  `Polygon`, `LineString`, `Marker`. Implement `Circle` (or `Marker`) first to establish the shared
  helpers (`clamp01`, the SVG-feature branch shape, the `callXMethod` + `Type`-constant host pattern,
  the shared `SIMULATION_VISUAL_PROPS`/`isBooleanProp`/`isNumericProp` additions); `Rectangle` then
  follows the same template with edge-coordinate geometry instead of center+radius.
  `DistanceToFeature` reaches full fidelity only once sibling features expose readable geometry.
- **Ordering:** `Map` → (`Circle`/`Marker` to establish shared pattern) → `Rectangle` (+ `Polygon`/
  `LineString`) → `FeatureCollection`. Without Map, only the non-visual Tier B subset (edge round-trip +
  `Bounds`/`Center`/`SetCenter`/distance math) is shippable, and it is low value.

## Web-Platform Limitations & Fidelity Caveats

- **No map without a library (primary caveat):** the simulator has no projection/tiles today; a
  `Rectangle` cannot be placed or sized geographically until `Map` + Leaflet land. Until then the box is
  a non-geographic stand-in.
- **Geographic size, not pixel size:** correct rendering needs the projection at the current zoom, and
  the box's screen height varies with latitude (Web Mercator distortion). A pixel-box fallback is
  visibly inaccurate.
- **Tiles are a network/licensing dependency:** real basemap tiles require an external provider over the
  network; offline/blank-canvas mode is the safe default and diverges from the on-device map look.
- **Drag fidelity:** Leaflet has no first-class draggable polygon/rectangle; it needs a path-drag plugin
  or a vertex/marker proxy, so `Draggable`/`StartDrag`/`Drag`/`StopDrag` are approximations even in Tier
  A and entirely inert in Tier B.
- **Degenerate / antimeridian boxes:** all edges default to `0` (zero-area box at (0,0) — invisible until
  set); AI does not enforce `North > South` / `West < East`, and a box straddling the antimeridian needs
  longitude wrapping for `Center`/rendering. These are fidelity edge cases to document.
- **Infobox styling:** `Title`/`Description` map to a Leaflet popup, which will not match the native
  Android infobox chrome exactly.
- **`DistanceToFeature` coupling:** exact only when the *target* feature is implemented and its geometry
  is readable from simulator state; otherwise it is `Unsupported`/`0`. The `centroid=false` edge-distance
  is also more involved than the centroid case and may ship centroid-only initially.
- **Color vs opacity separation:** AI keeps `FillColor`/`StrokeColor` ARGB separate from `FillOpacity`/
  `StrokeOpacity`; the renderer must apply opacity as a distinct channel, not fold it into the color
  alpha, to match device output.

## Effort Estimate

**L** — the `Rectangle` *itself* is S/M (capabilities entries, defaults, a styled SVG/Leaflet overlay,
the `Type` constant, and the pure-math `Bounds`/`Center`/`SetCenter`/`DistanceToPoint` host methods are
all straightforward, and most of the shared scaffolding lands with `Circle`/`Marker`). The effort is
dominated by the **hard `Map` prerequisite** (Leaflet integration, projected surface, pan/zoom, tile/
offline strategy, feature-child rendering + drag/infobox routing) which is **L/XL** on its own and must
precede a faithful Rectangle. As a standalone Tier-B stopgap (no Map) the Rectangle is **S**, but low
fidelity and not recommended as the end state.
