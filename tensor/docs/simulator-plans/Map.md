# Map Simulator Implementation Plan

## Overview

`Map` (category **MAPS**, `nonVisible = false`) is, per the spec's `helpString`, "a two-dimensional container that renders map tiles in the background and allows for multiple Marker elements to identify points on the map. Map tiles are supplied by OpenStreetMap contributors and the United States Geological Survey, or a custom basemap URL can be provided." It exposes a slippy-map viewport with a center (`CenterFromString` / `SetCenter` / `Latitude` / `Longitude`), a `ZoomLevel`, a `BoundingBox` (`[[North, West], [South, East]]`), a base tile layer selected by `MapType` (1 Roads / 2 Aerial / 3 Terrain / 4 Custom + `CustomUrl`), pan/zoom/rotation gesture toggles, optional compass / scale / zoom-control / user-location overlays, and a set of map-interaction events (`TapAtPoint`, `DoubleTapAtPoint`, `LongPressAtPoint`, `BoundsChange`, `ZoomChange`, `Ready`) plus feature-interaction events (`FeatureClick`, `FeatureDrag`, ...). It can also ingest GeoJSON via `LoadFromURL` / `FeatureFromDescription` and create markers via `CreateMarker`.

It is a **visible container**: it is the AI parent of the map-feature components. `canContainDesignComponent('Map', child)` already returns `true` for `MAP_CHILD_TYPES = {Circle, FeatureCollection, LineString, Marker, Polygon, Rectangle}` in `design-schema-tree.js` (lines 24, 65-67), and a nested `FeatureCollection` may contain the subset without `FeatureCollection` itself. Features are positioned geographically (lat/lng), not by pixel layout. Map itself can be placed on `Form`/`Screen` or in any arrangement, like `Image` or `Canvas`.

In the simulator it currently renders as a dashed "unsupported" placeholder because `Map` is absent from `SIMULATION_SUPPORTED_TYPES`.

## Feasibility Verdict

**Partially feasible.**

A slippy map is a well-solved web problem: a tiles library (Leaflet or MapLibre GL) renders OSM/USGS tiles, supports pan/zoom gestures, click/dblclick/contextmenu (long-press) events with `latlng`, programmatic `setView`/`flyTo`, and draggable feature markers/shapes. The core visible surface, the center/zoom/bounds model, the point-interaction events, and `CreateMarker`/`PanTo` are all directly implementable. The honest gaps are tile-source fidelity, device-only features, and file/storage methods:

- **A map needs a tiles library — a non-trivial dependency.** Native `<canvas>`/`<iframe>` cannot render an interactive slippy map alone. The realistic path is **Leaflet** (~42 KB gzipped, no GL, simplest API, raster tiles match AI's tile model) or **MapLibre GL** (heavier, needs WebGL + a style/vector source, but supports true `EnableRotation`). Leaflet is the recommended choice; it maps almost 1:1 onto AI's raster-tile + marker model. Either way this is the first AI component in this simulator to require a real third-party rendering library.
- **Tiles require network access and respect provider usage policies.** OSM's public tile server (the default `CustomUrl`) works for light use but has a usage policy; heavy/automated use can be blocked. Offline, no tiles load (gray grid). `MapType` 2 (Aerial) / 3 (Terrain) in AI map to specific tile providers (USGS, etc.) that may need their own URLs/keys; the simulator can ship sensible public raster sources for Roads/Terrain and a satellite source for Aerial, but exact tile imagery will differ from the device.
- **`ShowUser` / `UserLatitude` / `UserLongitude` / `LocationSensor`** depend on a real location provider. The browser Geolocation API exists but requires an explicit user permission prompt and returns the *developer's* machine location, not a simulated device. The realistic approximation is: leave `ShowUser` off by default, and if enabled, either (a) request browser geolocation (honest but prompts and leaks real location) or (b) place a synthetic user marker at the map center and report center as `UserLatitude`/`UserLongitude`. Option (b) is recommended for fidelity-without-surprise; document it.
- **`EnableRotation`** (rotate map to device orientation) needs a GL renderer (Leaflet does not rotate the base layer) **and** a device orientation/compass sensor that does not exist in a desktop browser. With Leaflet this is unsupported; with MapLibre the map *can* rotate programmatically but there is no orientation sensor to drive it. Treat as accept-but-no-op (or a manual rotation slider) and surface via `h.Unsupported`.
- **`ShowCompass`** similarly has no real digital compass; render a static decorative compass overlay at most.
- **`Save(path)`** writes a map image to device storage — no arbitrary device file path in a browser (same limitation as `Canvas.Save`). Approximation: trigger a download or `h.Unsupported`.
- **`LoadFromURL(url)`** fetches a GeoJSON document; cross-origin fetches are subject to CORS, so arbitrary URLs may fail (`LoadError`) where a device succeeds.

**Realistic simulated approximation:** render a Leaflet map bound to `CenterFromString`/`ZoomLevel`/`MapType`/`CustomUrl`; wire Leaflet's `click`/`dblclick`/`contextmenu`/`moveend`/`zoomend`/`load` to the AI point-interaction events; honor `EnablePan`/`EnableZoom`/`ShowZoom`/`ScaleUnits`/`ShowScale` via Leaflet options/controls; render feature children (Marker/Circle/LineString/Polygon/Rectangle) as Leaflet layers once those components exist; implement `CreateMarker`/`PanTo`/`SetCenter`; treat `ShowUser`/`EnableRotation`/`ShowCompass`/`LocationSensor` as best-effort/no-op with honest caveats; `Save` as download-or-unsupported; `LoadFromURL` as a real `fetch` with CORS caveats.

This is **the heaviest visible component to land in this simulator so far** — comparable to WebViewer in honesty caveats but with a hard external-library dependency on top.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `CenterFromString` | `"42.359144, -71.093612"` | Behavioral (write-only at runtime; sets initial center) | Parse `"lat, lng"` → `map.setView([lat,lng], zoom)`. Invalid → fire `InvalidPoint`. Drives initial `Latitude`/`Longitude`. | High |
| `ZoomLevel` | `13` | Visual/Behavioral | `map.setZoom()`; read back on `zoomend`. Clamp to provider min/max (1–18 typical). | High |
| `MapType` | `1` (Roads) | Visual | Selects the tile layer: 1 Roads (OSM), 2 Aerial (a satellite raster source), 3 Terrain (a topo raster source), 4 Custom (`CustomUrl`). Swap the Leaflet `tileLayer` on change. | High |
| `CustomUrl` | `https://tile.openstreetmap.org/{z}/{x}/{y}.png` | Behavioral | The `{z}/{x}/{y}` template for `tileLayer`. Used directly when `MapType=4` (and as the default Roads source). | High |
| `EnablePan` | `True` | Behavioral | `map.dragging.enable()/disable()`. | High |
| `EnableZoom` | `True` | Behavioral | `map.touchZoom`/`scrollWheelZoom`/`doubleClickZoom` enable/disable (note: AI says this does not affect the `ShowZoom` control buttons). | High |
| `ShowZoom` | `False` | Visual | Add/remove Leaflet `zoomControl`. | Medium |
| `ShowScale` | `False` | Visual | Add/remove `L.control.scale({ metric, imperial })` driven by `ScaleUnits`. | Medium |
| `ScaleUnits` | `1` (metric) | Behavioral | 1 → metric scale, 2 → imperial; configures the scale control. | Low |
| `Visible` | `True` | Visual | Standard `isSimulationVisible` gate; call `map.invalidateSize()` when re-shown. | High |
| `Width` / `Height` | `""` (auto) | Visual | `sizeStyle()` handles `-1`/`-2`/percent. A map with no explicit size is tiny in AI; render with a sensible min size and `invalidateSize()` after layout so tiles fill the box. | High |
| `Left` / `Top` | `""` | Visual (AbsoluteArrangement only) | Already handled via `baseStyle`/position; both already in `SIMULATION_VISUAL_PROPS`. | Low |
| `Rotation` | `0.0` | Behavioral | Leaflet cannot rotate the base layer; with MapLibre `setBearing()`. Accept; no-op under Leaflet. | Low |
| `EnableRotation` | `False` | Behavioral | No orientation sensor in browser; accept + no-op + `h.Unsupported`. | Low |
| `ShowCompass` | `False` | Visual | No digital compass; static decorative overlay at most. | Low |
| `ShowUser` | `False` | Visual/Behavioral | No device GPS; synthetic user marker at center (or browser geolocation with prompt). Drives `UserLatitude`/`UserLongitude`. | Low |
| `LocationSensor` | `""` | Behavioral (write-only) | References a LocationSensor component; that component is itself a device sensor unavailable in-browser. Accept + no-op. | Low |
| `BoundingBox` | (block-only, read-write, `[[N,W],[S,E]]`) | Behavioral | Read from `map.getBounds()`; write via `map.fitBounds([[S,W],[N,E]])`. | Medium |
| `Latitude` / `Longitude` | (block-only, read-only) | Behavioral | `map.getCenter().lat`/`.lng` — computed, not stored. | High |
| `UserLatitude` / `UserLongitude` | (block-only, read-only) | Behavioral | Synthetic (center) or browser geolocation; only meaningful when `ShowUser` on. | Low |
| `Features` | (block-only, read-write list) | Behavioral | The list of feature child components on the map. Backed by the design tree's Map children plus any created via `CreateMarker`/`FeatureFromDescription`. | Medium |
| `Column` / `Row` | (block-only, `rw: invisible`) | n/a | Layout artifacts; not simulated. | — |
| `HeightPercent` / `WidthPercent` | (write-only) | Visual | Map to the percent size encoding `sizeStyle()` already understands. | Low |

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `Ready` | (none) | Once, on Leaflet's `load` event after the map and first tiles initialize. |
| `TapAtPoint` | `latitude`, `longitude` (number) | On Leaflet `click` over the map base (not on a feature). `e.latlng.lat`/`.lng`. Fully firable. |
| `DoubleTapAtPoint` | `latitude`, `longitude` (number) | On Leaflet `dblclick`. Followed by `ZoomChange` if zoom gestures enabled and not at max zoom. Fully firable. |
| `LongPressAtPoint` | `latitude`, `longitude` (number) | On Leaflet `contextmenu` (right-click / touch long-press maps to contextmenu), or a pointerdown+hold timer. `e.latlng`. Firable. |
| `BoundsChange` | (none) | On Leaflet `moveend` after a user pan/zoom changes the viewport. Firable. |
| `ZoomChange` | (none) | On Leaflet `zoomend`. Firable. |
| `FeatureClick` | `feature` (component) | On `click` of a feature layer. **Gated on feature components** (Marker/Circle/...) being implemented; the `feature` arg is a component reference, so the host must map the clicked layer back to its AI component name. |
| `FeatureLongClick` | `feature` (component) | On `contextmenu` of a feature layer. Gated on feature components. |
| `FeatureDrag` / `FeatureStartDrag` / `FeatureStopDrag` | `feature` (component) | On Leaflet marker `drag`/`dragstart`/`dragend` (draggable markers). Gated on feature components. |
| `GotFeatures` | `url`, `features` (text, list) | On successful `LoadFromURL` GeoJSON fetch+parse. Firable but subject to CORS (see `LoadError`). |
| `LoadError` | `url`, `responseCode`, `errorMessage` (text, number, text) | On failed `LoadFromURL` (network/CORS/parse error). Firable; cross-origin failures will be common. |
| `InvalidPoint` | `message` (text) | When a coordinate parse/validation fails (`CenterFromString`, `PanTo`, `CreateMarker`, `FeatureFromDescription`). Fully firable. |

No event is blocked by a device sensor (unlike, e.g., AccelerometerSensor). The feature-interaction events are blocked only by the **feature child components not yet existing**, not by the platform. `GotFeatures`/`LoadError` are subject to CORS but are firable.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `PanTo(latitude, longitude, zoom)` | `(num, num, num) → void` | `map.flyTo([lat,lng], zoom)` (or `setView`). Validate coords → else `InvalidPoint`. Fires `BoundsChange`/`ZoomChange` via Leaflet's own `moveend`/`zoomend`. Fully feasible. |
| `CreateMarker(latitude, longitude)` | `(num, num) → component` | Add an `L.marker([lat,lng])` to the map and register a new synthetic Marker component in host/sim state; return its component reference. **Partial**: full return-value fidelity requires the `Marker` component to exist so the returned component can be addressed by later blocks; until then it can return a placeholder/handle. Validate coords → `InvalidPoint`. |
| `FeatureFromDescription(description)` | `(list) → any` | Parse a feature description list (GeoJSON-ish `(("type" "Point")("coordinates" (...)))`) into a feature layer + component. **Partial**: depends on feature components; parse + add layer is feasible, returning an addressable component is gated. Invalid → `InvalidPoint`. |
| `LoadFromURL(url)` | `(text) → void` | `fetch(url)` the GeoJSON; on success parse, add layers, fire `GotFeatures(url, features)`; on failure fire `LoadError(url, code, msg)`. **Partial — CORS**: cross-origin URLs without permissive headers fail where a device succeeds. |
| `Save(path)` | `(text) → void` | **Unsupported / partial.** No arbitrary device file path in a browser. Approximation: a `leaflet-image`-style export → download, or `h.Unsupported("method", "Map.Save")`. Low value in a simulator. |

## Implementation Plan

### simulation-capabilities.js

- Add `'Map'` to `SIMULATION_SUPPORTED_TYPES`. It is **visible**, so do **not** add it to `SIMULATION_NONVISIBLE_TYPES`.
- Add a defaults block inside `buildSimulationDefaults()`:

```js
Map: {
  Visible: true,
  Width: -1,
  Height: -1,
  CenterFromString: '42.359144, -71.093612',
  CustomUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
  MapType: 1,
  ZoomLevel: 13,
  EnablePan: true,
  EnableZoom: true,
  EnableRotation: false,
  Rotation: 0,
  ScaleUnits: 1,
  ShowCompass: false,
  ShowScale: false,
  ShowUser: false,
  ShowZoom: false,
},
```

- Add new prop names to `SIMULATION_VISUAL_PROPS` so they are not stripped before reaching the renderer: `CenterFromString`, `CustomUrl`, `MapType`, `ZoomLevel`, `EnablePan`, `EnableZoom`, `EnableRotation`, `Rotation`, `ScaleUnits`, `ShowCompass`, `ShowScale`, `ShowUser`, `ShowZoom`. (`Visible`, `Width`, `Height`, `Left`, `Top` already present.)
- `isBooleanProp`: add `EnablePan`, `EnableZoom`, `EnableRotation`, `ShowCompass`, `ShowScale`, `ShowUser`, `ShowZoom`.
- `isNumericProp`: add `MapType`, `ZoomLevel`, `Rotation`, `ScaleUnits`. (`Width`/`Height`/`Left`/`Top` already covered.)
- `coerceSimulationValue`: no special case needed for the simple props — booleans/numbers go through the existing `coerceBoolean`/`coerceNumber`. `CenterFromString`/`CustomUrl` stay strings (default fallthrough). Optionally add a `deriveStateFromDesignerProps` branch for `Map` that parses `CenterFromString` into derived `_centerLat`/`_centerLng` numbers for the renderer — but this can equally be done in the Svelte branch; not required for v1.

### SimulationComponent.svelte

Add a branch modeled on a hybrid of the **arrangement** branches (it hosts feature children) and a library-backed container (like the planned WebViewer iframe). Sketch:

```svelte
{:else if node.type === 'Map'}
  <div
    bind:this={mapEl}
    class="sim-map"
    class:sim-unsupported={unsupportedHere}
    style={baseStyle('position: relative; overflow: hidden;', { backgroundImage: false })}
    data-sim-component={node.name}
  >
    <!-- Leaflet renders tiles + controls into mapEl via JS -->
    <!-- feature children are added as Leaflet layers, not DOM, so they are
         iterated to register layers rather than rendered as <svelte:self> DOM -->
    {#each node.children || [] as child (child.pathId || child.name)}
      <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
    {/each}
  </div>
```

- **Library init:** in an `onMount`/`afterUpdate` (or a reactive `mapEl` effect), lazily `import('leaflet')`, create `L.map(mapEl, { dragging, zoomControl, scrollWheelZoom, doubleClickZoom, touchZoom })` from props, add the `tileLayer` for the current `MapType`/`CustomUrl`, `setView(parseLatLng(props.CenterFromString), props.ZoomLevel)`. Keep the `L.Map` instance in a `let mapInstance` and **tear it down in `onDestroy`** (`mapInstance.remove()`).
- **Sizing:** Leaflet needs a sized container before init and an `invalidateSize()` after layout/visibility changes. Use a `ResizeObserver` on `mapEl` → `mapInstance.invalidateSize()`. The map box uses `sizeStyle()` via `baseStyle`; supply a min-height fallback.
- **Reactive prop → map sync:** a reactive block diffs `props.ZoomLevel`/`CenterFromString`/`MapType`/`CustomUrl`/`EnablePan`/`EnableZoom`/`ShowZoom`/`ShowScale`/`ScaleUnits` and calls the matching Leaflet API (`setZoom`, `setView`, swap `tileLayer`, `dragging.enable/disable`, add/remove `zoomControl`/scale control). Guard against feedback loops with the host (suppress emitting events for programmatic moves, mirroring the `suppressTextChanged` pattern in the host).
- **Event wiring:** attach Leaflet handlers → `emitEvent`: `load`→`Ready`; `click`→`TapAtPoint([lat,lng])`; `dblclick`→`DoubleTapAtPoint`; `contextmenu`→`LongPressAtPoint`; `moveend`→`BoundsChange` (and patch back `Latitude`/`Longitude`/`BoundingBox`); `zoomend`→`ZoomChange` (patch `ZoomLevel`). Use the multi-arg `emitEvent(node.name, 'TapAtPoint', [lat, lng])` path (already supported).
- **Feature children:** Marker/Circle/LineString/Polygon/Rectangle children should NOT render as DOM; they must register Leaflet layers on the parent map. Since `<svelte:self>` recursion renders DOM, the cleanest v1 is for the Map branch to **read its feature children from `node.children`** and create the corresponding Leaflet layers itself (or have feature components emit a "register layer" interaction the Map listens for). Either way, feature components are a **follow-on**; until they exist, the map renders with no features and feature events do not fire.
- **CSS:** import Leaflet's stylesheet (`leaflet/dist/leaflet.css`) — Leaflet is unusable without it. `.sim-map` is a positioned block (`sizeStyle`), `overflow:hidden`. Note marker-icon asset URLs: Leaflet's default marker images must be bundled/resolved (a common Leaflet+bundler gotcha — set `L.Icon.Default.imagePath` or import the icon assets).

### SimulateOverlay.svelte

**Likely none for the map surface.** Tiles, pan/zoom, controls, and point-events are handled entirely in the component branch via Leaflet. A small overlay handler is only needed if `CreateMarker`/`FeatureFromDescription`/`LoadFromURL` are modeled as host **effects** that the overlay must apply (e.g. "add a layer at lat/lng" pushed from Go). If those methods instead patch sim state (the new feature's props) and the Map branch reconstructs layers from state on the next render, no overlay change is required. `Save`-as-download (if implemented) would need a small `map-save` effect, but that is optional polish.

### simulation_wasm.go

A `Map` type needs real handlers (the generic store path covers plain prop reads/writes but not computed center/bounds, coordinate-method validation, GeoJSON loading, or feature creation):

- **`GetProperty`:** add `case "Map"`. Most designer props (`MapType`, `ZoomLevel`, `EnablePan`, `CustomUrl`, ...) fall through to the generic `h.state` read. Computed read-only props need synthesis: `Latitude`/`Longitude` from the stored center (parsed `CenterFromString` or the last `moveend`-patched center), `BoundingBox` from the last viewport patch, `UserLatitude`/`UserLongitude` from center (synthetic-user approximation) or `NullVal` when `ShowUser` is off, `Features` from the registered feature list. If the frontend patches center/zoom/bounds back via `statePatch` on `moveend`/`zoomend`, these reduce to plain reads.
- **`SetProperty`:** add `case "Map"`. `CenterFromString` write → validate; on invalid, `h.runEvent(name, "Map", "InvalidPoint", [msg])` and skip. `BoundingBox` write → store + emit an effect/patch so the renderer `fitBounds`. `MapType`/`CustomUrl`/`ZoomLevel`/`EnablePan`/`EnableZoom`/`ShowZoom`/`ShowScale`/`ScaleUnits`/`ShowUser` → plain `h.setProperty` (renderer reacts via the prop diff). `Rotation`/`EnableRotation`/`ShowCompass`/`LocationSensor` → accept + `h.Unsupported("property", "Map."+property+" (no browser sensor)")` once, then store. No `*Changed` event on Map property writes.
- **`CallMethod`:** add `case "Map": return h.callMapMethod(...)`:
  - `PanTo(lat, lng, zoom)` → validate; patch center/zoom into state + emit a `component-action` effect `{action:'panTo', lat, lng, zoom}` the renderer applies (`flyTo`). Invalid → `InvalidPoint`.
  - `CreateMarker(lat, lng)` → validate; register a synthetic Marker component (name, props) in `h.state` + `h.componentTypes`, push a `map-add-feature` effect, return a component reference. **Partial** until the Marker component exists.
  - `FeatureFromDescription(desc)` → parse list; on success add feature + return component; invalid → `InvalidPoint`. **Partial.**
  - `LoadFromURL(url)` → this is async (network); model as an effect `{action:'loadFromUrl', url}` the overlay/renderer performs with `fetch`, then the renderer calls back `GotFeatures`/`LoadError` via `eventRunner`. (The Go host cannot `fetch` synchronously.) **Partial — CORS.**
  - `Save(path)` → `h.Unsupported("method", "Map.Save")` (or a `map-save` download effect).
- **Events:** point-interaction and bounds/zoom events are fired **frontend-side** from Leaflet via `emitEvent`/`eventRunner`; the host does not synthesize them. `InvalidPoint`/`GotFeatures`/`LoadError` are fired from method handling (host-side for `InvalidPoint`, renderer-side for the async load).
- Recompile with `npm run build:wasm`.

### design-schema-tree.js

**No change required.** Containment is already encoded: `MAP_CHILD_TYPES = {Circle, FeatureCollection, LineString, Marker, Polygon, Rectangle}` and `canContainDesignComponent('Map', child)` returns `true` for those (lines 65-67), with the `FeatureCollection` subset rule already handled. `designTreeToInitialState()` will merge the new `SIMULATION_DEFAULTS.Map` automatically, and `unsupportedSimulationComponents()` will stop flagging Map once it is in `SIMULATION_SUPPORTED_TYPES`. (The feature child types remain unsupported and will render as placeholders / fail to produce Leaflet layers until they are implemented.)

## Dependencies & Ordering

- **External library required: Leaflet** (recommended, ~42 KB gzipped, raster tiles, simplest fit to AI's tile model) or MapLibre GL (heavier, WebGL, only needed if true `EnableRotation` is desired). This is the first simulator component to require a third-party rendering library; it must be added to `package.json`, and its CSS imported. Lazy-load it (`import('leaflet')`) so it does not bloat the initial bundle for apps without a Map.
- **Network access** for tiles (OSM/USGS/satellite providers) and for `LoadFromURL` (CORS-permitting).
- **Implementation order:** the **Map surface + point-interaction events + center/zoom/bounds model + `PanTo`** are independently implementable and valuable on their own (a pannable, zoomable, tappable map with no features). The **feature child components — `Marker`, then `Circle` / `LineString` / `Polygon` / `Rectangle`, and the `FeatureCollection` container — are a required follow-on** for `FeatureClick`/`FeatureDrag`/`CreateMarker`/`Features`/`FeatureFromDescription` to be meaningful. So: **Map first, then Marker, then the other feature types + FeatureCollection.** Until features land, feature events do not fire and feature-returning methods are partial.

## Web-Platform Limitations & Fidelity Caveats

- **Tiles require network and a tile provider; offline = blank/gray map.** OSM's public tiles have a usage policy; `MapType` Aerial/Terrain map to providers that may need their own URLs/keys, so exact imagery differs from the device.
- **`EnableRotation` / `Rotation` is effectively unsupported.** Leaflet does not rotate the base layer, and there is no browser device-orientation/compass to drive rotation. MapLibre can rotate programmatically but still has no sensor input.
- **`ShowCompass` has no real digital compass** — decorative overlay at best.
- **`ShowUser` / `UserLatitude` / `UserLongitude` / `LocationSensor` have no simulated device GPS.** Either a synthetic user marker at center (recommended) or browser geolocation (prompts, leaks the developer's real location). Neither matches a real device's moving GPS fix.
- **`LoadFromURL` is CORS-constrained.** Cross-origin GeoJSON without permissive headers fails (`LoadError`) where a device's HTTP client succeeds.
- **`Save` cannot write to device storage** — no arbitrary file path in a browser (same as `Canvas.Save`); download-or-unsupported.
- **`CreateMarker` / `FeatureFromDescription` return component references** that are only fully usable once the feature child components exist; until then they are partial (placeholder handles).
- **Feature interaction events are gated on the feature components**, not on the platform.
- **Tile rendering, projection precision, and gesture feel** (Web Mercator via Leaflet vs. Android's map renderer) will be close but not pixel/behavior-identical; zoom-level min/max bounds depend on the chosen tile provider.

## Effort Estimate

**XL.** The hard external-library integration (add Leaflet + its CSS + lazy load + marker-icon asset resolution + lifecycle teardown + `invalidateSize`/`ResizeObserver`), a bidirectional prop↔Leaflet sync layer with feedback-loop suppression, six point-interaction events plus bounds/zoom patch-back, the `callMapMethod` host handler (`PanTo`/`CreateMarker`/`FeatureFromDescription`/`LoadFromURL`/`Save`) with coordinate validation and an async `LoadFromURL` round-trip via effects, and the honest no-op surfaces for rotation/compass/user-location — combined with the fact that the component is only *complete* once the entire **Marker / Circle / LineString / Polygon / Rectangle / FeatureCollection** family is also implemented — push this clearly beyond L. The standalone map surface alone is ~L; the full feature ecosystem makes the component XL.
