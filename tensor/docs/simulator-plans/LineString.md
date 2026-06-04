# LineString Simulator Implementation Plan

## Overview

`LineString` (category **MAPS**, `nonVisible = false`) is a visible map feature that draws "an open, continuous sequence of lines on a Map." Per the spec's `helpString`, in the AI designer you add vertices by dragging the midpoint of a segment, move a vertex by dragging it, and delete a vertex by clicking it (down to a minimum of two). At runtime it is defined by an ordered list of `[latitude, longitude]` points (`Points` / `PointsFromString`) rendered as a stroked polyline on the map, with a clickable/long-clickable/draggable body and an optional infobox (`Title` + `Description`) popup.

Container relationship: **child of `Map` or `FeatureCollection`** (the parents). `canContainDesignComponent` already encodes this in `design-schema-tree.js`: `MAP_CHILD_TYPES` (line 24) and `FEATURE_COLLECTION_CHILD_TYPES` (line 25) both include `'LineString'`, and `canContainDesignComponent('Map'|'FeatureCollection', 'LineString')` returns `true`. LineString has no children (it is a leaf, valid only inside a `Map` or a `FeatureCollection`, which itself lives in a `Map`).

It currently renders as a dashed "unsupported" placeholder because neither `LineString` nor its required parent `Map` is in `SIMULATION_SUPPORTED_TYPES`, so `unsupportedSimulationComponents()` flags it.

## Feasibility Verdict

**Partially feasible — and strictly blocked on a `Map` implementation that does not yet exist.**

A polyline itself is trivial to draw in a browser (an SVG `<polyline>`/`<path>` with `stroke`, `stroke-width`, `stroke-opacity`). The hard part is everything a polyline *needs around it*: a geographic coordinate system (latitude/longitude), a projection from lat/lng to screen pixels, a zoom/pan-able tiled basemap, and a parent `Map` host that owns the viewport (`CenterFromString`, `ZoomLevel`, bounds) and the child-feature coordinate space. None of that exists in the simulator today:

1. **No `Map` host exists.** LineString's `Points` are geographic coordinates, not pixels. To render them you must project each `[lat, lng]` through the Map's current center/zoom into the Map's pixel box and reproject on every pan/zoom. There is no `Map` branch in `SimulationComponent.svelte`, no Map defaults, and no Map in the Go host. **`Map` must be implemented first** — it owns the basemap tiles, the viewport, the lat/lng↔pixel projection, and the stacking context the polyline draws into. Without a Map, a LineString has no coordinate frame and cannot be placed meaningfully.

2. **Map tiles are a heavy dependency.** A faithful Map requires a tile-rendering library (Leaflet or MapLibre GL) plus a tile source. That is a real external dependency and, for OSM/MapLibre, network tile fetches — which a sandboxed, possibly-offline client simulator may not want. The realistic v1 approximation for the *Map* is either (a) embed Leaflet and let it own projection + a `Polyline` layer (LineString becomes a thin Leaflet `L.polyline` wrapper), or (b) a tile-less "blank canvas with a gray graticule" where lat/lng is linearly mapped to the Map's pixel box (no real geography, but vertices, stroke, and events all work). Option (b) keeps the dependency at zero and is the honest minimum that still exercises the blocks.

3. **The drag/vertex editing in the `helpString` is a designer authoring feature, not a runtime behavior** — the simulator does not need to reproduce midpoint-drag vertex insertion. It only needs to render the resulting `Points` and fire the runtime events. That reduces scope.

There are **no device limitations** here — no GPS, sensors, camera, or permissions are involved (LineString does not read the device location; that is the `Map`'s `LocationSensor`/user-location feature, a separate concern). The blocker is purely architectural: it is feasible only after a `Map` exists, and fidelity depends on whether that Map uses a real tile library or a blank projection. Standalone (without Map): **not feasible**.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Points` | (block-only, list) | Behavioral + Visual | The ordered `[[lat,lng], …]` vertex list. Drives the polyline geometry. Read-write at runtime. Store as an array; project each pair through the Map to build the SVG `points`/`d`. | High |
| `PointsFromString` | `""` | Behavioral (write-only) | Designer + runtime text form of `Points`: `[[lat1,lng1],[lat2,lng2],…]`. Parse JSON in `deriveStateFromDesignerProps` → set `Points`. Write-only at runtime per spec (no getter). | High |
| `StrokeColor` | `&HFF000000` | Visual | Polyline stroke color. `colorValue(props.StrokeColor, '#000')` → SVG `stroke`. Read-write. | High |
| `StrokeWidth` | `3` | Visual | Stroke width in px. `numberOr(props.StrokeWidth, 3)` → SVG `stroke-width`. Read-write. | High |
| `StrokeOpacity` | `1.0` | Visual | Stroke alpha 0..1. → SVG `stroke-opacity` (clamp 0..1). Read-write. | Medium |
| `Visible` | `True` | Visual | Show/hide the polyline. Standard `Visible` handling. | High |
| `Title` | `""` | Behavioral | Heading shown in the infobox popup. Used only when the infobox opens. | Medium |
| `Description` | `""` | Behavioral | Body text in the infobox popup. | Medium |
| `EnableInfobox` | `False` | Behavioral | If true, clicking the line opens the infobox (Title/Description popup) in addition to firing `Click`. If false, click only fires the event. | Medium |
| `Draggable` | `False` | Behavioral | If true, the user can long-press-and-drag the whole line; gates `LongClick` (which only fires when `Draggable = false`) and enables `StartDrag`/`Drag`/`StopDrag`. In v1, dragging the whole polyline by pointer is implementable but low priority; honoring the gate is what matters. | Low |
| `Type` | (read-only) | Behavioral | Constant `"LineString"`. Serve from the Go host `GetProperty` (computed, no stored value). | Low |

Notes: LineString has **no font/text-display/`BackgroundColor`/`TextColor`/`Width`/`Height` properties** — its only geometry is `Points`, and its size is implied by them; do not add layout-size props. `StrokeColor` is AI-encoded as `&HAARRGGBB` (the `&HFF000000` default = opaque black); reuse `colorValue()`. `StrokeWidth`/`StrokeOpacity` are numeric; `Visible`/`Draggable`/`EnableInfobox` are boolean. `Points` is a list and must NOT be coerced like a scalar.

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `Click` | (none) | `pointerup`/click on the polyline body. `emitEvent(node.name, 'Click')`. If `EnableInfobox` is true, also open the infobox. Firable in a browser via SVG path pointer events. |
| `LongClick` | (none) | A press held > ~500 ms on the polyline **while `Draggable = false`** (per spec, only fires when not draggable). Track pointerdown time; emit on a timer if still pressed. Firable. |
| `Drag` | (none) | Continuous pointer-move while dragging the line, **only when `Draggable = true`**. Emit on move. Firable, but low priority (needs whole-line drag). |
| `StartDrag` | (none) | First move after a long-press begins a drag (when `Draggable = true`). Emit once at drag start. Firable, low priority. |
| `StopDrag` | (none) | `pointerup` ending a drag (when `Draggable = true`). Emit once. Firable, low priority. |

All events are firable in a browser — there are no sensor/permission gates. The only dependency is the `Map` parent providing the pixel frame the polyline is drawn and hit-tested in. `Click`/`LongClick` need only pointer wiring on the SVG path; `Drag`/`StartDrag`/`StopDrag` additionally need whole-line drag handling and are deferrable past v1.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `ShowInfobox` | `() → void` | Open the infobox popup (Title + Description) anchored near the line, regardless of `EnableInfobox`. Implement as a Go `effect` (`componentAction(name, "show-infobox")`) the overlay renders, or a frontend-local popup toggled by host method dispatch. Feasible. |
| `HideInfobox` | `() → void` | Close the infobox if shown; no-op otherwise. Effect `componentAction(name, "hide-infobox")`. Feasible. |
| `DistanceToPoint` | `(latitude:number, longitude:number, centroid:boolean) → number` | Compute meters between the line and a point. Feasible **purely in JS** with the haversine formula (point-to-polyline = min over segments of point-to-segment great-circle distance; `centroid=true` = distance to the line's centroid). Implement in the Go host (`callLineStringMethod`) using the stored `Points`. No map tiles needed — it is pure geometry. |
| `DistanceToFeature` | `(mapFeature:component, centroids:boolean) → number` | Distance in meters to another map feature. Feasible in principle (read the other feature's `Points`/geometry and run the same haversine min-distance), but requires the other feature types (`Marker`/`Polygon`/etc.) to exist and expose geometry. **Mark `Unsupported` for v1** (`h.Unsupported("method", name+".DistanceToFeature")`) until sibling features land; revisit after Map + other features exist. |

## Implementation Plan

### simulation-capabilities.js

- **`SIMULATION_SUPPORTED_TYPES`**: add `'LineString'` (and add `'Map'` when the Map host lands — LineString is meaningless without it). LineString is **visible**, so do NOT add it to `SIMULATION_NONVISIBLE_TYPES`.
- **`buildSimulationDefaults()`** — add a defaults block:
  ```js
  LineString: {
    Visible: true,
    Points: [],
    PointsFromString: '',
    StrokeColor: '&HFF000000',
    StrokeWidth: 3,
    StrokeOpacity: 1.0,
    Title: '',
    Description: '',
    EnableInfobox: false,
    Draggable: false,
  },
  ```
- **`SIMULATION_VISUAL_PROPS`**: add `'Points'`, `'PointsFromString'`, `'StrokeColor'`, `'StrokeWidth'`, `'StrokeOpacity'`, `'Description'`, `'EnableInfobox'`, `'Draggable'`. (`Visible` and `Title` are already present.) Props not in this set are stripped before reaching the renderer, so the polyline geometry and stroke would vanish without these.
- **`isBooleanProp`**: add `'EnableInfobox'`, `'Draggable'`.
- **`isNumericProp`**: add `'StrokeWidth'`. Do **not** add `'StrokeOpacity'` to `isNumericProp` only if you need fractional preservation — `coerceNumber` keeps floats fine, so adding it is safe and preferred (so `"1.0"` → `1`). Add `'StrokeOpacity'` too.
- **`coerceSimulationValue`**: add a `Points` branch that accepts an array of `[lat,lng]` pairs (pass through as-is, like `Elements`/arrays) and a `PointsFromString` branch that JSON-parses the `[[lat,lng],…]` string into the points array. Keep `Points` out of the numeric/boolean coercion paths (it is a list).
- **`deriveStateFromDesignerProps`**: when `PointsFromString` is set, parse it and also set `next.Points` (mirror the `ElementsFromString → Elements` pattern at line 658). Add a small `parsePointsFromString(value)` helper (JSON.parse, validate array-of-2-number-arrays, fallback `[]`).

### SimulationComponent.svelte

Add a render branch keyed on `node.type === 'LineString'`. Because a LineString only has meaning inside a Map's projected pixel box, the branch renders an SVG polyline positioned in the Map's coordinate space. v1 (blank-projection Map) sketch:

```svelte
{:else if node.type === 'LineString'}
  <svg class="sim-linestring" class:sim-unsupported={unsupportedHere}
       style={baseStyle('position:absolute; inset:0; overflow:visible; pointer-events:none;')}
       data-sim-component={node.name}>
    {#if boolValue(props.Visible, true) && projectedPoints(props.Points).length >= 2}
      <polyline
        points={projectedPoints(props.Points).map(p => `${p.x},${p.y}`).join(' ')}
        fill="none"
        stroke={colorValue(props.StrokeColor, '#000')}
        stroke-width={numberOr(props.StrokeWidth, 3)}
        stroke-opacity={Math.max(0, Math.min(1, numberOr(props.StrokeOpacity, 1)))}
        stroke-linejoin="round" stroke-linecap="round"
        style="pointer-events:stroke; cursor:pointer;"
        on:pointerdown={onLineStringDown}
        on:pointerup={onLineStringUp}
        on:click={onLineStringClick} />
    {/if}
  </svg>
{/if}
```

- **Projection helper**: `projectedPoints(points)` converts each `[lat,lng]` to `{x,y}` in the Map's pixel box. v1 = linear map of lat/lng across the Map bounds the parent Map supplies (via context/prop); a real Leaflet Map would instead use `map.latLngToContainerPoint`.
- **Event wiring**: reuse the long-press pattern already used for `LongClick`/`TouchDown` (see the Slider/Button handlers, lines 530–544). `onLineStringClick` → `emitEvent(node.name, 'Click')` and, if `EnableInfobox`, open the infobox; a held-press timer → `emitEvent(node.name, 'LongClick')` when `!Draggable`. Drag events are gated behind `Draggable` and deferred.
- **Helpers to reuse**: `baseStyle()`, `colorValue()`, `numberOr()`, `boolValue()`, `emitEvent()`. No `sizeStyle`/`alignmentStyle` (no box layout). Keep the `sim-unsupported` class wiring so an in-Map-but-unprojectable LineString still degrades visibly.
- **CSS**: minimal — `.sim-linestring { position:absolute; inset:0; }`; the visible stroke is driven by SVG attributes. Infobox popup styling can mirror the Notifier/picker popup CSS.

### SimulateOverlay.svelte

Needed **only for the infobox**. `ShowInfobox`/`HideInfobox` and the `EnableInfobox` click-popup are best rendered as a small anchored popup. Two options: (a) the Go host emits `componentAction(name, "show-infobox" | "hide-infobox")` effects and the overlay renders a Notifier-style popup with `Title`/`Description` (consistent with the existing effect→overlay pattern, lines 671–722); or (b) a purely frontend-local popup inside the LineString branch. Prefer (a) for `ShowInfobox`/`HideInfobox` (host-driven), reuse the frontend for the `EnableInfobox` click case. If the infobox is descoped from v1, this file is **none**.

### simulation_wasm.go

- **`GetProperty`**: add a `case "LineString"` returning the computed read-only `Type` → `runtime.StrVal("LineString")`. All stored props (`Points`, `StrokeColor`, …) fall through to the generic `h.state` path already at the top of `GetProperty` (lines 86–90).
- **`SetProperty`**: add a `PointsFromString` special-case (parse string → set `Points`, mirroring the `ElementsFromString → setElements` branch at lines 120–123) so a runtime `set PointsFromString` updates the rendered `Points`. Everything else (`Points`, `StrokeColor`, `StrokeWidth`, `StrokeOpacity`, `Visible`, `Title`, `Description`, `EnableInfobox`, `Draggable`) uses the generic `h.setProperty` store path — no special handler needed.
- **`CallMethod`**: add `case "LineString": return h.callLineStringMethod(...)`. In `callLineStringMethod`:
  - `ShowInfobox`/`HideInfobox` → append `componentAction(name, "show-infobox"|"hide-infobox")` to `h.effects`.
  - `DistanceToPoint` → compute haversine point-to-polyline distance from stored `Points` and the args; return `runtime.NumVal(meters)`.
  - `DistanceToFeature` → `h.Unsupported("method", name+".DistanceToFeature")` for v1 (other features absent); return `runtime.VoidVal()`/`NumVal(0)`.
  - default → `h.Unsupported("method", name+"."+method)`.
- **Events**: `Click`/`LongClick`/`Drag`/`StartDrag`/`StopDrag` originate from the frontend pointer wiring via `emitEvent` → `h.runEvent`; no Go-side synthesis required. Compile with `npm run build:wasm`.

### design-schema-tree.js

**No change required.** Containment is already encoded: `MAP_CHILD_TYPES` and `FEATURE_COLLECTION_CHILD_TYPES` (lines 24–25) include `'LineString'`, and `canContainDesignComponent` returns `true` for `Map → LineString` and `FeatureCollection → LineString` (lines 65–68). `designTreeToInitialState()` will merge `SIMULATION_DEFAULTS.LineString` automatically once the defaults block exists, and `unsupportedSimulationComponents()` stops flagging it once `'LineString'` is in `SIMULATION_SUPPORTED_TYPES`. (It will still render as an unsupported placeholder while its `Map` parent remains unimplemented.)

## Dependencies & Ordering

1. **`Map` MUST be implemented first.** It is the hard blocker: it owns the basemap, the viewport (`CenterFromString`/`ZoomLevel`/bounds), the lat/lng↔pixel projection, and the stacking context the polyline draws into. Until Map exists, LineString has no coordinate frame.
2. **`FeatureCollection`** (optional grouping parent) should follow Map; LineString works directly under Map without it.
3. **External library (for high fidelity):** a tile-map library — **Leaflet** (lighter, raster tiles) or **MapLibre GL** (heavier, vector). With Leaflet, LineString becomes a thin wrapper over `L.polyline`. **Zero-dependency alternative:** a blank-projection Map (linear lat/lng → pixel box, no tiles), which keeps the bundle clean and still exercises `Points`/stroke/events. Choose at the Map layer; this plan works under either.
4. `DistanceToFeature` additionally depends on sibling feature types (`Marker`, `Polygon`, `Rectangle`, `Circle`) existing and exposing geometry — defer.

## Web-Platform Limitations & Fidelity Caveats

- **No real geography without a tile library.** With the zero-dependency blank-projection Map, there is no basemap imagery and lat/lng is mapped linearly (not a true Web Mercator projection), so distances and shapes will distort vs a real device. A Leaflet-based Map fixes this but adds a dependency and (for online tiles) network fetches.
- **Designer vertex editing is not reproduced.** The `helpString`'s midpoint-drag-to-add / click-to-delete vertex authoring is an AI *designer* affordance; the simulator only renders the resulting `Points` and fires runtime events.
- **Whole-line dragging (`Draggable`/`Drag`/`StartDrag`/`StopDrag`) is approximate/deferred.** Faithful map-drag (re-projecting the line as the user drags it across the basemap) is fiddly; v1 honors the `Draggable` gate (so `LongClick` fires only when not draggable) but may not implement live drag-to-relocate.
- **`DistanceToFeature` returns nothing useful in v1** (other features unimplemented); `DistanceToPoint` is accurate (haversine) but only as good as the projection used for hit/render.
- **Hit-testing is stroke-width-limited.** Clicking a thin SVG polyline uses `pointer-events:stroke`; a very thin `StrokeWidth` makes the line hard to click — same as a real device but more noticeable on desktop.

## Effort Estimate

**L** — the LineString branch itself is **S–M** (SVG polyline + stroke props + click/long-click wiring + a haversine `DistanceToPoint`), but it is gated behind implementing a **`Map` host** (projection, viewport, optional tile library), which is the dominant cost and pushes the realistic end-to-end effort to **L** (and toward **XL** if a full Leaflet/MapLibre basemap with pan/zoom is required).
