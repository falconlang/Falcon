# Polygon Simulator Implementation Plan

## Overview

`Polygon` (category **MAPS**, `nonVisible = false`) is, per the spec's `helpString`, a map feature that
"encloses an arbitrary 2-dimensional area on a Map. Polygons can be used for drawing a perimeter, such
as a campus, city, or country." Its geometry is a list of `(latitude, longitude)` vertices
(`Points`), optionally with interior `HolePoints` (a list of vertex-lists punched out of the fill). Its
appearance is a stroked, filled shape: `FillColor`/`FillOpacity` for the interior, `StrokeColor`/
`StrokeOpacity`/`StrokeWidth` for the outline. It can carry an info window (`Title` + `Description`,
gated by `EnableInfobox`) and can be made `Draggable`.

- **Category:** MAPS.
- **Visible:** Yes (`nonVisible: false`), but it is **not** a top-level visible component — it is a
  geographic overlay primitive that only has meaning when rendered on a **Map** surface with a known
  lat/lng-to-pixel projection.
- **Container relationship:** `Polygon` is a **child of `Map` or `FeatureCollection`** (a
  `FeatureCollection` is itself a child of `Map`). `canContainDesignComponent` already encodes this in
  `design-schema-tree.js`: `MAP_CHILD_TYPES` and `FEATURE_COLLECTION_CHILD_TYPES` both include
  `'Polygon'`, and a Polygon's parent must be `Map` or `FeatureCollection` (lines 24-25, 65-67). It is
  never a parent. It therefore **hard-depends on a `Map` implementation**: every vertex is a lat/lng
  pair that must be projected into Map pixel space, so without a Map there is no coordinate frame and
  nothing to draw against.

It currently renders as a dashed "unsupported" placeholder because `Polygon` is absent from
`SIMULATION_SUPPORTED_TYPES`. **No mapping library is present** in the project (`package.json` has no
Leaflet/MapLibre/OpenLayers).

## Feasibility Verdict

**Partially feasible (in a browser simulator) — and blocked on Map.**

The *shape itself* is trivially feasible in the browser: an SVG `<polygon>`/`<path>` (with even-odd
fill-rule for holes) styled by `FillColor`/`FillOpacity`/`StrokeColor`/`StrokeWidth` is the exact web
analog of a map polygon overlay, and `pointerdown`/`pointermove`/`pointerup`/`click` on that SVG path
give us `Click`/`LongClick`/drag events directly. The hard parts are entirely about the **Map context
the polygon lives in**, and they split into two concrete gaps:

1. **No Map surface / no projection (primary blocker).** A polygon's vertices are geographic
   `(lat, lng)`. To draw them you must project lat/lng to on-screen pixels, which requires a Map that
   defines a center, zoom, and projection (Web Mercator). The simulator has **no Map component**
   (`Map` is not in `SIMULATION_SUPPORTED_TYPES`), no projection code, and no tile layer. Until Map
   exists, the polygon has no coordinate frame. This is the same class of dependency that makes `Ball`
   depend on `Canvas`, except heavier: a faithful Map needs **either a mapping library
   (Leaflet/MapLibre) or a from-scratch Web-Mercator projection + raster tile fetch**, both
   non-trivial.

2. **Basemap tiles need a network dependency or a library.** Real fidelity ("draw a perimeter around a
   city") requires a basemap underneath the polygon. Tiles come from a tile server (OSM/Mapbox/etc.)
   and a projection library. That is a **heavy dependency** and a network/licensing concern in a
   client-side simulator. The realistic approximation is to render the polygon **without a basemap**
   (or over a flat neutral background / coarse static graticule), which loses geographic context but
   correctly shows the shape, fill, stroke, and relative vertex layout.

**Realistic approximation (what we will simulate), assuming Map lands first as a projected surface:**
- Render the polygon as an SVG path inside the Map's projected coordinate layer, honoring
  `FillColor` + `FillOpacity`, `StrokeColor` + `StrokeOpacity` + `StrokeWidth`, `Points`,
  `HolePoints` (even-odd holes), and `Visible`.
- Parse `PointsFromString` / `HolePointsFromString` (write-only string forms) into the point lists.
- Fire `Click` on tap and `LongClick` on long-press-when-not-draggable from real DOM pointer events.
- Show the infobox (`Title` + `Description`) as a small DOM popover on click when `EnableInfobox` is
  true, and implement `ShowInfobox`/`HideInfobox` to toggle it.
- Implement `Centroid` (polygon centroid math, returns `[lat, lng]`) purely in the host — no Map
  needed for the math.
- Support `Draggable` + `Drag`/`StartDrag`/`StopDrag` **only if** the Map exposes an inverse projection
  (pixel delta -> lat/lng delta); otherwise mark drag `Unsupported`.
- Mark `DistanceToFeature`/`DistanceToPoint` as best-effort (haversine on centroids/points) or
  `Unsupported` depending on how much geodesy we want to commit to.

This is a credible "shows the shape and its styling, fires click/longclick, computes centroid" cut that
is honest that **there is no basemap and the whole thing is meaningless until Map is implemented.**

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
| --- | --- | --- | --- | --- |
| `Points` | `[]` (empty until set) | Visual — list of `[lat, lng]` vertices | Parse to point array; project each lat/lng via the Map; build SVG path `d`. **Needs Map projection.** | High |
| `PointsFromString` | `""` | Visual (write-only) — string form of `Points` | Parse the AI list-string (e.g. `[[lat1,lng1],[lat2,lng2],...]`) into `Points`; derive in `deriveStateFromDesignerProps`. | High |
| `HolePoints` | `[]` | Visual — list of vertex-lists punched out | Append each hole as a subpath; SVG `fill-rule: evenodd`. **Needs Map projection.** | Medium |
| `HolePointsFromString` | `""` | Visual (write-only) | Parse list-of-lists string into `HolePoints`. | Medium |
| `FillColor` | `&HFFFF0000` (red) | Visual | `colorValue(props.FillColor)` -> SVG `fill`. | High |
| `FillOpacity` | `1.0` | Visual | SVG `fill-opacity`; clamp `[0,1]`. | High |
| `StrokeColor` | `&HFF000000` (black) | Visual | `colorValue(props.StrokeColor)` -> SVG `stroke`. | High |
| `StrokeOpacity` | `1.0` | Visual | SVG `stroke-opacity`; clamp `[0,1]`. | High |
| `StrokeWidth` | `1` | Visual | SVG `stroke-width` in px. | High |
| `Visible` | `True` | Visual | Standard `isSimulationVisible` gate on the branch. | High |
| `Title` | `""` | Behavioral | Infobox header text. | Medium |
| `Description` | `""` | Behavioral | Infobox body text. | Medium |
| `EnableInfobox` | `False` | Behavioral | Whether a tap opens the infobox popover. | Medium |
| `Draggable` | `False` | Behavioral | Enables pointer-drag of the whole polygon. **Only meaningful if Map exposes inverse projection** to translate pixel delta -> lat/lng delta applied to every vertex. | Low |
| `Type` | (read-only) | Behavioral | Constant `"Polygon"`; host `GetProperty` returns `runtime.StrVal("Polygon")`. | Low |

Notes:
- Defaults are pulled from the spec's `designerProperties`. `FillColor`/`StrokeColor` are AI ARGB hex
  (`&HFFFF0000`/`&HFF000000`) which `colorValue()` already parses. Note `FillColor`'s alpha byte (`FF`)
  is independent of `FillOpacity`; the SVG should use `colorValue` for the RGB and `*-opacity` for the
  opacity rather than double-applying.
- `Points`/`HolePoints` are AI lists of lists. The block runtime stores them as lists, so coercion must
  pass arrays through unchanged (they are not `Elements`). The string forms are write-only conveniences
  that must be parsed into the list forms.
- **Cannot be honored without Map:** the actual on-screen *position* of the polygon (everything that
  needs lat/lng -> pixel projection). With no Map, only the styling props are renderable in the
  abstract; the shape cannot be placed.

## Events

| Event | Args | How/when the simulator fires it |
| --- | --- | --- |
| `Click` | (none) | `click`/short `pointerup` on the SVG path. `emitEvent(node.name, 'Click', [])`. Also opens the infobox if `EnableInfobox`. **Feasible** (real DOM event). |
| `LongClick` | (none) | Long-press (>~600ms `pointerdown` with no significant move) **and only when `Draggable` is false** (matches spec). Reuse the existing `setTimeout` long-press pattern from Button. **Feasible.** |
| `StartDrag` | (none) | `pointerdown` that begins a drag, when `Draggable` is true. **Feasible only if Map provides inverse projection** so the drag can move vertices; otherwise the event can still fire but the polygon will not move (degraded). |
| `Drag` | (none) | `pointermove` during a drag, when `Draggable` is true. Per-move dispatch. Same Map-projection caveat as `StartDrag`. |
| `StopDrag` | (none) | `pointerup` ending a drag, when `Draggable` is true. Same caveat. |

All five events are user-interaction events with no sensor/permission involvement, so all are
**firable in a browser** once there is an SVG path to attach handlers to. The drag trio is only
*meaningful* (visually moves the polygon and reports new vertices) when the Map exposes a
pixel->lat/lng inverse projection; without it, drag should be marked `Unsupported` rather than firing
events that do nothing.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
| --- | --- | --- |
| `Centroid()` | `() -> list` | **Feasible, host-only.** Compute the polygon centroid from `Points` (standard signed-area centroid formula over the vertex ring) and return `runtime.ListVal([lat, lng])`. No Map projection needed — pure math on the stored vertices. |
| `ShowInfobox()` | `() -> void` | **Feasible.** Emit a `component-action` effect (`componentActionWith(name, "show-infobox", {...})`) the renderer turns into an open popover with `Title`/`Description`. Shows even if `EnableInfobox` is false (per spec). |
| `HideInfobox()` | `() -> void` | **Feasible.** Emit `component-action` `hide-infobox`; renderer closes the popover. No-op if not shown. |
| `DistanceToPoint(latitude, longitude, centroid)` | `(num, num, bool) -> number` | **Partial.** With `centroid=true`, haversine distance (meters) from the polygon centroid to the point is computable host-side and reasonably faithful. With `centroid=false`, AI returns the distance to the *nearest edge* — computable with point-to-segment geodesic math but more work; v1 can approximate with nearest-vertex/centroid and `h.logs` a note, or mark `Unsupported`. |
| `DistanceToFeature(mapFeature, centroids)` | `(component, bool) -> number` | **Partial / mostly Unsupported in v1.** Requires reading the *other* feature's geometry, which only exists once sibling Map features (Marker/Circle/LineString/Rectangle/Polygon) are implemented and their points are host-readable. With `centroids=true` and a sibling whose centroid is known, haversine is computable; otherwise `h.Unsupported("method", "Polygon.DistanceToFeature")`. |

`Centroid`, `ShowInfobox`, `HideInfobox` are the high-value, fully feasible methods. The two
distance methods are geodesy that is feasible in principle but should be scoped carefully (centroid-mode
haversine first; edge/feature distance later or `Unsupported`).

## Implementation Plan

### simulation-capabilities.js
- **`SIMULATION_SUPPORTED_TYPES`:** add `'Polygon'`. It is **visible**, so do **not** add it to
  `SIMULATION_NONVISIBLE_TYPES`.
- **`buildSimulationDefaults()`** — add a defaults block (Polygon has no `Width`/`Height`/`Enabled`
  designer props; it uses geometry + paint props):
  ```js
  Polygon: {
    Visible: true,
    Points: [],
    HolePoints: [],
    PointsFromString: '',
    HolePointsFromString: '',
    FillColor: '&HFFFF0000',
    FillOpacity: 1,
    StrokeColor: '&HFF000000',
    StrokeOpacity: 1,
    StrokeWidth: 1,
    Title: '',
    Description: '',
    EnableInfobox: false,
    Draggable: false,
  },
  ```
- **`SIMULATION_VISUAL_PROPS`:** add the new names `'Points'`, `'HolePoints'`, `'PointsFromString'`,
  `'HolePointsFromString'`, `'FillColor'`, `'FillOpacity'`, `'StrokeColor'`, `'StrokeOpacity'`,
  `'StrokeWidth'`, `'Description'`, `'EnableInfobox'`, `'Draggable'`. (`Visible` and `Title` already
  present.) Without these the props are **stripped** before reaching the renderer.
- **`isBooleanProp`:** add `'EnableInfobox'`, `'Draggable'`.
- **`isNumericProp`:** add `'FillOpacity'`, `'StrokeOpacity'`, `'StrokeWidth'`. (Do **not** add
  `FillColor`/`StrokeColor` — they stay ARGB strings handled by `colorValue`.)
- **`coerceSimulationValue`:** `Points`/`HolePoints` arrive as arrays (lists). The existing
  `if (Array.isArray(value)) return propName === 'Elements' ? normalizeElements(value) : value;` line
  already passes non-`Elements` arrays through unchanged — good, no change needed for the array case.
  Add string-parse handling for `PointsFromString`/`HolePointsFromString` (see derive below) so a
  designer string yields the list form.
- **`deriveStateFromDesignerProps`:** add a Polygon branch (mirroring the `ElementsFromString ->
  Elements` pattern): when `key === 'PointsFromString'`, parse the AI list-string into `next.Points`;
  when `key === 'HolePointsFromString'`, parse into `next.HolePoints`. Add a small
  `parsePointList(str)` helper (tolerant parse of `[[lat,lng],...]`, falling back to `[]` on bad input)
  alongside `parseListData`.

### SimulationComponent.svelte
This branch must render **inside the Map's projected layer** (the Map plan must provide a positioned
SVG/overlay surface and a `project(lat, lng) -> {x, y}` function passed to children, analogous to how
the Canvas plan provides a positioned surface for sprites). Sketch the branch (place near the other
leaf components, before the `{:else}` placeholder):
```svelte
{:else if node.type === 'Polygon'}
  {#if isSimulationVisible(state, node.name)}
    <svg
      class="sim-polygon"
      class:sim-unsupported={unsupportedHere}
      data-sim-component={node.name}
      style="position:absolute; inset:0; overflow:visible; pointer-events:none;"
    >
      <path
        d={polygonPathData()}
        fill={colorValue(props.FillColor, '#f00')}
        fill-opacity={clamp01(numberOr(props.FillOpacity, 1))}
        fill-rule="evenodd"
        stroke={colorValue(props.StrokeColor, '#000')}
        stroke-opacity={clamp01(numberOr(props.StrokeOpacity, 1))}
        stroke-width={numberOr(props.StrokeWidth, 1)}
        style="pointer-events:auto; cursor:pointer;"
        on:click={polygonClick}
        on:pointerdown={polygonPointerDown}
        on:pointermove={polygonPointerMove}
        on:pointerup={polygonPointerUp}
      />
    </svg>
    {#if infoboxOpen}
      <div class="sim-polygon-infobox" style={infoboxStyle()}>
        {#if props.Title}<strong>{props.Title}</strong>{/if}
        {#if props.Description}<span>{props.Description}</span>{/if}
      </div>
    {/if}
  {/if}
{/if}
```
- **`polygonPathData()`** (new helper): take `props.Points` (and each `HolePoints` ring), project each
  `[lat, lng]` through the Map-provided `project()` to `{x, y}`, and assemble an SVG path `M x0 y0 L x1
  y1 ... Z` for the outer ring followed by one `M...Z` subpath per hole. If no `project()` is available
  (Map not implemented / not passing a projection), return `''` and the polygon draws nothing (graceful
  degradation). Reuse `colorValue`, `numberOr`; add a tiny `clamp01()`.
- **Pointer/long-press wiring:** reuse the existing Button long-press pattern (`pointerDown(longClick)`
  with the `setTimeout` -> `LongClick`, `suppressClick`, `consumeLongClick`). `polygonClick` fires
  `Click` (skipping it if a long-press just fired) and, if `EnableInfobox`, sets `infoboxOpen = true`.
  Long-press fires `LongClick` only when `!boolValue(props.Draggable)`. When `Draggable` is true,
  `pointerdown`/`move`/`up` drive `StartDrag`/`Drag`/`StopDrag` (and, if the Map exposes an inverse
  projection, translate the pixel delta into a lat/lng delta applied to every vertex via a host method
  / state patch so the shape follows the pointer).
- **Infobox** is a small absolutely-positioned DOM popover (not an SVG element), positioned near the
  polygon centroid pixel. Process `show-infobox`/`hide-infobox` component-action effects via the
  existing `handleComponentActions(actions?.[node.name])` reactive path to toggle `infoboxOpen`.
- **CSS:** `.sim-polygon { position: absolute; inset: 0; }` (the SVG overlays the Map layer);
  `.sim-polygon-infobox` is a small white rounded card with shadow. Do **not** use `baseStyle()` (it
  injects width/height/background/alignment that do not apply to a geographic overlay).
- **Reuse:** `colorValue`, `numberOr`, `boolValue`, the Button long-press machinery, and the
  `handleComponentActions` effect path. The SVG-overlay approach parallels the Canvas plan's
  positioned-surface-with-children model.

### SimulateOverlay.svelte
**None strictly required.** The infobox is rendered locally in `SimulationComponent.svelte` and toggled
via `component-action` effects (`show-infobox`/`hide-infobox`) that `applyEffects()` already routes
through `actionTokens` to the renderer — the same mechanism Canvas/TextBox use for `open`/`focus`. No
modal dialog or toast is needed (the infobox is an inline popover, not a Notifier-style dialog). If a
future iteration wants a modal-style infobox it could move here, but that is not necessary for v1.

### simulation_wasm.go
A `Polygon` type needs a small handler set; most paint/geometry props ride the **generic store path**:
- **`GetProperty`:** add `case "Polygon"` with one computed property — `Type` returns
  `runtime.StrVal("Polygon")` (read-only constant). `Points`/`HolePoints`/`FillColor`/`FillOpacity`/
  `StrokeColor`/`StrokeOpacity`/`StrokeWidth`/`Title`/`Description`/`EnableInfobox`/`Draggable`/`Visible`
  all read straight from `h.state` via the generic fallthrough.
- **`SetProperty`:** generic `h.setProperty` suffices for all read-write props. `PointsFromString`/
  `HolePointsFromString` (write-only) should be handled like `ElementsFromString`: parse the string into
  a list and store it under `Points`/`HolePoints` (so the renderer re-projects and redraws live).
- **`CallMethod` -> add `case "Polygon": return h.callPolygonMethod(...)`:**
  - `Centroid()`: compute the signed-area centroid from stored `Points`; return `runtime.ListVal`
    of `[lat, lng]`. Pure Go math, no Map needed.
  - `ShowInfobox()`: `h.effects = append(h.effects, componentAction(name, "show-infobox"))`.
  - `HideInfobox()`: `h.effects = append(h.effects, componentAction(name, "hide-infobox"))`.
  - `DistanceToPoint(lat, lng, centroid)`: if `centroid` true, haversine(centroid, point) in meters;
    else best-effort (nearest vertex) + `h.logs` note, or `h.Unsupported("method",
    "Polygon.DistanceToPoint")` for the edge-distance case.
  - `DistanceToFeature(feature, centroids)`: `h.Unsupported("method", "Polygon.DistanceToFeature")`
    in v1 (sibling-feature geometry reads not yet available); upgrade to haversine once sibling Map
    features expose centroids.
  - `default`: `h.Unsupported("method", componentName+"."+method)`.
- **Events:** `Click`/`LongClick`/`StartDrag`/`Drag`/`StopDrag` are dispatched **frontend-side** via
  `emitEvent`/`eventRunner` from the SVG pointer handlers; no host-side `runEvent` synthesis needed.
- Recompile with `npm run build:wasm`.

### design-schema-tree.js
**Already handled.** `MAP_CHILD_TYPES` and `FEATURE_COLLECTION_CHILD_TYPES` already include `'Polygon'`,
and `canContainDesignComponent` returns `true` for `parent === 'Map'` (line 66) and `parent ===
'FeatureCollection'` with the subset check (line 67) — so containment validation is correct with no
edit. `designTreeToInitialState` will merge `SIMULATION_DEFAULTS.Polygon` once the defaults block
exists, and `unsupportedSimulationComponents` stops flagging `Polygon` once it is added to
`SIMULATION_SUPPORTED_TYPES`. No code change in this file.

## Dependencies & Ordering
- **Hard prerequisite: `Map` must be implemented first.** A Polygon is meaningless without a Map
  surface that (a) renders a positioned overlay layer for its features, (b) defines center/zoom and a
  `project(lat, lng) -> {x, y}` (Web Mercator) passed to children, and (c) ideally an inverse
  `unproject(x, y) -> {lat, lng}` for drag. This is the dominant dependency — implement Map, then
  Polygon (and its sibling features Marker/Circle/LineString/Rectangle, which share the same
  projection plumbing).
- **External libraries:** the *shape* needs none — SVG `<path>` is built in. A faithful **basemap**
  (tiles) needs **Leaflet or MapLibre GL** (each a heavyweight client-side dependency, plus a tile
  source / licensing). The realistic v1 is **no basemap** (flat neutral background under the projected
  overlay), which keeps the simulator dependency-free at the cost of geographic context. Adding a
  basemap library is a separate, larger decision owned by the Map plan, not this one.
- **Soft prerequisite: sibling Map features** (Marker/Circle/LineString/Rectangle/Polygon) for
  `DistanceToFeature` to be more than `Unsupported`.
- **Related:** `FeatureCollection` as an intermediate container (it is itself a Map child) should be
  handled by the Map work so a Polygon nested in a FeatureCollection still resolves to the Map's
  projection.

## Web-Platform Limitations & Fidelity Caveats
- **No basemap (primary caveat).** Without Leaflet/MapLibre + tiles, the polygon draws over a flat
  background with no streets/satellite imagery, so the user sees the shape's outline/fill but not
  *where on Earth* it is. "Draw a perimeter around a campus" shows a red blob, not the campus.
- **Blocked on Map / projection.** Everything positional depends on a Map providing a lat/lng->pixel
  projection. With no Map, the polygon cannot be placed at all (the branch draws nothing). The styling
  props are renderable in the abstract but have no geometry to apply to.
- **Drag fidelity.** `Draggable` + `Drag`/`StartDrag`/`StopDrag` only move the polygon if the Map
  exposes an inverse projection; the lat/lng delta from a pixel drag is an approximation (Mercator is
  not equal-area, so dragging far from the original latitude distorts), and without the inverse the
  drag events fire but the shape does not move.
- **Distance geodesy is approximate.** `DistanceToPoint`/`DistanceToFeature` use haversine on a sphere;
  AI uses the same family of approximations, but edge-distance (`centroid=false`) and nearest-feature
  distance are more involved and are scoped to `Unsupported`/best-effort in v1.
- **Vertex editing is not interactive.** AI's designer lets the user drag midpoints to add vertices and
  click vertices to delete them (min 3). The simulator renders the polygon as given by `Points`; it
  does **not** offer interactive vertex editing — vertices change only via block-set `Points`/
  `PointsFromString`. This matches the runtime, not the designer-edit, behavior.
- **No marker clustering / z-ordering subtleties.** Overlap order among multiple features depends on
  DOM/SVG order; it will not exactly match Android map-feature z-ordering.

## Effort Estimate
**L** (XL if a real basemap is in scope). The SVG path rendering, paint-prop plumbing
(capabilities entries, visual/numeric/boolean sets, `PointsFromString` parsing), the
`Click`/`LongClick` wiring (reusing Button's long-press), the infobox popover + `ShowInfobox`/
`HideInfobox` effects, and the host `callPolygonMethod` with `Centroid` math are each modest (M in
isolation). The cost drivers are (1) the **hard dependency on a Map projection surface that does not
yet exist** — most of the real work is actually the Map, and Polygon is a thin layer on top — and (2)
the geodesy for the distance methods + drag inverse-projection. A **basemap (Leaflet/MapLibre + tiles)
pushes the overall Maps initiative to XL**, but that cost belongs to the Map plan; the Polygon-specific
work, given a projecting Map, is **L**.
