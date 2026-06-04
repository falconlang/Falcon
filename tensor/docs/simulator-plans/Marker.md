# Marker Simulator Implementation Plan

## Overview
`Marker` (category **MAPS**, `nonVisible="false"`) is, per the spec helpString, "An icon positioned at a point to indicate information on a map. Markers can be used to provide an info window, custom fill and stroke colors, and custom images to convey information to the user."

It is a **visible** map-feature component. Crucially it is **not a top-level visible component** — it is a *child* of a `Map` (or of a `FeatureCollection` nested inside a `Map`). It has no standalone visual representation: its `Latitude`/`Longitude` only have meaning relative to the geographic projection and viewport of its parent `Map`. The default marker is a teardrop pin icon anchored at its tip; it can show an info window (title + description) on tap and optionally be dragged to a new lat/lng.

Containment (`design-schema-tree.js`): `MAP_CHILD_TYPES` and `FEATURE_COLLECTION_CHILD_TYPES` already include `Marker`, and `canContainDesignComponent('Map'|'FeatureCollection', 'Marker')` already returns `true`. So the tree/placement layer is ready — but the renderer and host are not, and **neither `Map` nor `FeatureCollection` is in `SIMULATION_SUPPORTED_TYPES`**, so today a `Map` (and therefore everything under it, including `Marker`) renders as the dashed unsupported placeholder.

## Feasibility Verdict
**Partially feasible (and strictly gated on a `Map` implementation first).**

A `Marker` is meaningless without a coordinate space to position it in. The simulator currently has no `Map` host: there is no tile layer, no projection (lat/lng -> pixel), no pan/zoom viewport, and no container into which a child feature can be absolutely positioned. Therefore a `Marker` cannot be rendered, dragged, or hit-tested in isolation. Its feasibility is entirely a function of how `Map` is implemented.

What is realistically achievable **once a Map host exists**:
- Rendering a pin (default teardrop, or `ImageAsset`) at a projected pixel position with a configurable anchor — fully feasible (DOM/CSS, absolute positioning inside the map viewport).
- Custom `FillColor`/`StrokeColor`/opacity/`StrokeWidth` on the default pin — feasible if the default pin is drawn as inline SVG (so fill/stroke are themeable); a raster `ImageAsset` cannot be re-tinted, matching real AI behavior.
- `Click`, `LongClick`, and an info window (`Title`/`Description`) toggled by `EnableInfobox`/`ShowInfobox`/`HideInfobox` — feasible with DOM events + a small popover.
- `Draggable` + `StartDrag`/`Drag`/`StopDrag` and `SetLocation` — feasible *only* if the Map host exposes an inverse projection (pixel -> lat/lng); a pointer drag updates pixel position and back-projects to new `Latitude`/`Longitude`.

What needs the inverse-projection math (so depends on the chosen map library):
- `DistanceToPoint`/`DistanceToFeature` (great-circle / haversine) and `BearingToPoint`/`BearingToFeature` are pure math and can be implemented in Go regardless of the renderer — feasible and even useful independent of tiles.

Honest limitation: drawing actual map tiles needs a heavy dependency (Leaflet / MapLibre GL) or an OSM raster tile fetch, which is a `Map`-level decision. A `Marker` plan can only specify the child piece; everything geographic flows from the `Map` host. If `Map` is implemented as a "blank panel with a projection" (no real tiles), the `Marker` still works visually and behaviorally; it just floats over a gray panel instead of a real map.

## Properties
Defaults pulled from the spec `designerProperties`.

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Latitude` | `0` | Behavioral (position) | Fed to the Map host's `latLngToPixel` projection to place the pin; round-trips on drag. | High |
| `Longitude` | `0` | Behavioral (position) | Same as Latitude. | High |
| `Visible` | `True` | Visual | Standard `visible` gate (`isSimulationVisible`); hide the pin element. | High |
| `Title` | `""` | Visual (infobox) | Bold heading line in the info window popover. | High |
| `Description` | `""` | Visual (infobox) | Secondary line in the info window popover. | High |
| `EnableInfobox` | `False` | Behavioral | If true, `Click` opens the info window; if false, no auto-open (but `ShowInfobox` still forces it). | High |
| `ImageAsset` | `""` | Visual | `resolveAssetUrl(assets, ImageAsset)`; if set, render `<img>` instead of the default SVG pin. Empty = default pin. | High |
| `FillColor` | `&HFFFF0000` (red) | Visual | Fill of the default SVG pin via `colorValue()`. Ignored when `ImageAsset` is set (matches AI). | Med |
| `StrokeColor` | `&HFF000000` (black) | Visual | SVG `stroke` of the default pin. Ignored when `ImageAsset` set. | Med |
| `StrokeWidth` | `1` | Visual | SVG `stroke-width` px. | Med |
| `FillOpacity` | `1.0` | Visual | SVG fill opacity (0-1). | Med |
| `StrokeOpacity` | `1.0` | Visual | SVG stroke opacity (0-1). | Med |
| `AnchorHorizontal` | `3` (center) | Visual | Which point of the icon sits on the lat/lng: 1=left,2=right,3=center -> CSS `transform-origin`/translate offset. | Med |
| `AnchorVertical` | `3` (bottom) | Visual | 1=top,2=center,3=bottom -> translate offset so a pin's tip lands on the coordinate. | Med |
| `Draggable` | `False` | Behavioral | Enables pointer-drag handling + `StartDrag`/`Drag`/`StopDrag`; when true, `LongClick` does NOT fire (per spec). | Med |
| `Width` / `Height` | `-1` (block-only, no designer default) | Visual | Icon box size in px when set; otherwise intrinsic icon size. | Low |

Cannot be honored / not worth simulating:
- `ShowShadow` (`rw="invisible"`) — drop-shadow under the pin; cosmetic, optional CSS `filter: drop-shadow(...)`. Low value.
- `HeightPercent` / `WidthPercent` (write-only) — percentage sizing relative to Screen; markers are icon-sized, ignore.

## Events
| Event | Args | How/when the simulator fires it |
|---|---|---|
| `Click` | none | Pointer `click` on the pin (when not suppressed by a preceding long-press/drag). If `EnableInfobox` is true, also opens the info window. Reuse the `Click`/`consumeLongClick` pattern. |
| `LongClick` | none | 600ms press timer (reuse `pointerDown(true)`/`longClickTimer`) — **only fires when `Draggable` is false**, per spec. Suppresses the following `Click`. |
| `StartDrag` | none | On drag start (pointerdown that begins moving) when `Draggable` is true. |
| `Drag` | none | During pointer move while dragging; throttle to fire as the pin moves. Updates `Latitude`/`Longitude` via inverse projection. |
| `StopDrag` | none | On pointerup ending a drag. |

None of these require device permissions or sensors — all are pointer-driven and fully fireable in a browser. The only blocker is having the Map viewport to receive the pointer events.

## Methods
| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `SetLocation` | `(latitude:number, longitude:number) -> void` | Write both `Latitude` and `Longitude` to state (one `setProperty` each) and re-project the pin. Fully feasible. |
| `ShowInfobox` | `() -> void` | Force the info window open even if `EnableInfobox` is false (per spec). Emit a state flag / effect the renderer reads. Feasible. |
| `HideInfobox` | `() -> void` | Close the info window if open; no-op otherwise. Feasible. |
| `DistanceToPoint` | `(lat:number, lng:number) -> number` | Haversine great-circle distance in meters, computed in Go from current `Latitude`/`Longitude`. Pure math — fully feasible. |
| `BearingToPoint` | `(lat:number, lng:number) -> number` | Initial bearing (degrees from due north) via standard formula. Fully feasible. |
| `DistanceToFeature` | `(mapFeature:component, centroids:boolean) -> number` | Feasible *only* between markers (point-to-point haversine). For polygon/line features, "nearest point" / centroid requires those features' geometry to exist in the host. Implement marker-to-marker now; `Unsupported("method", ...)` for non-point features (or until those are added). |
| `BearingToFeature` | `(mapFeature:component, centroids:boolean) -> number` | Same caveat as `DistanceToFeature` — point-to-point feasible, polygon/line nearest-point not. |

## Implementation Plan

**Prerequisite:** implement `Map` (and ideally `FeatureCollection`) first. The steps below assume the Map host provides a projection API the child can call (`latLngToPixel(lat,lng)` and `pixelToLatLng(x,y)`), a sized viewport element, and current center/zoom in shared state. Without that, `Marker` has nowhere to render.

- **simulation-capabilities.js**
  - Add `'Marker'` to `SIMULATION_SUPPORTED_TYPES` (it is visible, so NOT added to `SIMULATION_NONVISIBLE_TYPES`). Note: `Map` and `FeatureCollection` must also be added as part of their own work; a `Marker` whose parent is still unsupported will still placeholder.
  - Add a defaults block in `buildSimulationDefaults()`:
    ```js
    Marker: {
      Visible: true,
      Latitude: 0,
      Longitude: 0,
      Title: '',
      Description: '',
      EnableInfobox: false,
      Draggable: false,
      ImageAsset: '',
      FillColor: '&HFFFF0000',
      FillOpacity: 1.0,
      StrokeColor: '&HFF000000',
      StrokeOpacity: 1.0,
      StrokeWidth: 1,
      AnchorHorizontal: 3,
      AnchorVertical: 3,
    },
    ```
  - Add to `SIMULATION_VISUAL_PROPS` the new names not already present: `Latitude`, `Longitude`, `Description`, `EnableInfobox`, `Draggable`, `ImageAsset`, `FillColor`, `FillOpacity`, `StrokeColor`, `StrokeOpacity`, `StrokeWidth`, `AnchorHorizontal`, `AnchorVertical`. (`Visible`, `Title`, `Width`, `Height` are already present.) Props NOT in this set are stripped before reaching the renderer, so all of the above are required.
  - `isBooleanProp`: add `'EnableInfobox'`, `'Draggable'` (and `'ShowShadow'` if you choose to honor it).
  - `isNumericProp`: add `'Latitude'`, `'Longitude'`, `'FillOpacity'`, `'StrokeOpacity'`, `'StrokeWidth'`. (`AnchorHorizontal`/`AnchorVertical` are numeric; add them too. Note these overlap conceptually with alignment but are plain numbers here.)
  - `coerceSimulationValue`: covered by the boolean/numeric additions above; no special-case needed. `ImageAsset` stays a string. No new `deriveStateFromDesignerProps` derivation required.

- **SimulationComponent.svelte**
  - Add a render branch. A `Marker` should NOT render at top level on its own — it must render inside the Map viewport. Practically, the Map branch will iterate its feature children and position each; the `Marker` branch produces the pin element that the Map branch absolutely-positions. Sketch:
    ```svelte
    {:else if node.type === 'Marker'}
      <div class="sim-marker"
           class:sim-unsupported={unsupportedHere}
           style={markerAnchorStyle()}
           data-sim-component={node.name}
           on:pointerdown={markerPointerDown}
           on:click={markerClick}>
        {#if assetUrl}
          <img class="sim-marker-img" src={assetUrl} alt={props.Title ?? ''} />
        {:else}
          <svg class="sim-marker-pin" viewBox="0 0 24 36" width="24" height="36">
            <path d="M12 0 C5 0 0 5 0 12 C0 21 12 36 12 36 C12 36 24 21 24 12 C24 5 19 0 12 0 Z"
                  fill={colorValue(props.FillColor, '#f00')}
                  fill-opacity={numberOr(props.FillOpacity, 1)}
                  stroke={colorValue(props.StrokeColor, '#000')}
                  stroke-opacity={numberOr(props.StrokeOpacity, 1)}
                  stroke-width={numberOr(props.StrokeWidth, 1)} />
          </svg>
        {/if}
        {#if infoboxOpen && (hasValue(props.Title) || hasValue(props.Description))}
          <div class="sim-marker-infobox">
            {#if hasValue(props.Title)}<strong>{props.Title}</strong>{/if}
            {#if hasValue(props.Description)}<div>{props.Description}</div>{/if}
          </div>
        {/if}
      </div>
    ```
  - Helpers to reuse: `colorValue()` (ARGB -> rgba), `numberOr()`, `boolValue()`, `resolveAssetUrl(assets, ...)` via the existing `assetName()` (point `assetName()` at `props.ImageAsset` for the `Marker` type), `hasValue()`. Add a small `markerAnchorStyle()` that maps `AnchorHorizontal`/`AnchorVertical` to a `translate(...)` so the chosen icon point sits on the projected coordinate (default 3/3 = horizontally centered, anchored at bottom tip).
  - Event wiring: reuse the established interaction model. `markerClick` -> `consumeLongClick()` guard, then `emitEvent(node.name,'Click')`, and toggle `infoboxOpen` if `EnableInfobox`. `markerPointerDown` -> reuse `pointerDown(longClick=!Draggable)` so `LongClick` only arms when not draggable; if `Draggable`, attach pointermove/pointerup that emit `StartDrag`/`Drag`/`StopDrag` and call up to the Map host for `pixelToLatLng` then `emitInteraction([{component, property:'Latitude'...},{...Longitude}])`.
  - CSS: `.sim-marker { position:absolute; cursor:pointer; }` (absolute within the Map viewport), `.sim-marker-infobox` a small white popover with shadow, `.sim-unsupported` already styled.

- **SimulateOverlay.svelte**
  - The info window can be a local DOM popover inside the marker branch (above), so **no overlay change is strictly required**. Only touch the overlay if you decide info windows / drag feedback should be host-driven "effects"; for `Marker` that is unnecessary — none.

- **simulation_wasm.go**
  - Add a `callMarkerMethod(componentName string, method string, args []runtime.Value)` and route `case "Marker":` in `CallMethod`:
    - `SetLocation`: `h.setProperty(name,"Latitude",args[0]); h.setProperty(name,"Longitude",args[1])`.
    - `ShowInfobox` / `HideInfobox`: set a transient flag (e.g. `h.setProperty(name,"_InfoboxOpen", bool)`) the renderer reads, OR push an effect; the simplest is a state flag added to `SIMULATION_VISUAL_PROPS` as an internal prop.
    - `DistanceToPoint` / `BearingToPoint`: compute haversine / bearing in Go from the marker's current lat/lng; return `runtime.NumVal(...)`.
    - `DistanceToFeature` / `BearingToFeature`: if the target component is another `Marker`, read its lat/lng from `h.state` and compute point-to-point; otherwise `h.Unsupported("method", name+"."+method+" for non-point features")`.
  - `GetProperty`: add `case "Marker":` for the read-only `Type` -> return `runtime.StrVal("Marker")`. Other reads (`Latitude`, `Longitude`, etc.) fall through to the generic store path.
  - `SetProperty`: the generic store path (`h.setProperty`) suffices for `Latitude`/`Longitude`/colors/`Title`/etc. No special handling needed unless you want a `LocationChanged`-style re-render hook (none in spec). The Map host re-projects on its own reactive pass.
  - Events (`h.runEvent`): `Click`, `LongClick`, `StartDrag`, `Drag`, `StopDrag` are dispatched from the renderer through the existing event pipeline; no Go-side origination needed beyond the generic event routing.

- **design-schema-tree.js**
  - Containment is **already handled**: `MAP_CHILD_TYPES`/`FEATURE_COLLECTION_CHILD_TYPES` include `Marker` and `canContainDesignComponent` returns `true` for `Map`/`FeatureCollection` parents. `designTreeToInitialState()` will automatically merge `SIMULATION_DEFAULTS.Marker` once the defaults block is added. `unsupportedSimulationComponents()` will stop flagging `Marker` once it is in `SIMULATION_SUPPORTED_TYPES`. No edits required here.

## Dependencies & Ordering
1. **`Map` must be implemented first** — it owns the viewport, tile/background, projection (`latLngToPixel`/`pixelToLatLng`), and pan/zoom. `Marker` is a child that positions itself in that space and is non-functional without it.
2. **`FeatureCollection`** ideally next (a Map can nest markers inside a FeatureCollection); not strictly required for a Map-direct marker, but `DistanceToFeature` across collections benefits from it.
3. External library decision lives with `Map`: Leaflet or MapLibre GL for real tiles (heavy), OR a lightweight self-rolled equirectangular/Web-Mercator projection over a blank/gray panel (no extra dependency, lower fidelity). `Marker` itself needs **no external library** — just inline SVG + the projection the Map exposes.
4. Other point features (`Circle`, `LineString`, `Polygon`, `Rectangle`) are independent; `BearingToFeature`/`DistanceToFeature` against them stays `Unsupported` until they exist.

## Web-Platform Limitations & Fidelity Caveats
- **No map tiles by default** in a client-side simulator without a heavy library or external tile fetch; the realistic baseline is a projected blank panel, so markers float over gray rather than real geography. (This is a `Map` limitation that the `Marker` inherits.)
- **`ImageAsset` cannot be tinted** — `FillColor`/`StrokeColor`/opacity apply only to the default SVG pin, identical to real AI behavior, so this is acceptable fidelity.
- **Drag precision depends on the projection**; with a simplified (non-tiled) projection, dragged lat/lng will be approximate and won't match a real Mercator tile grid pixel-for-pixel.
- **`DistanceToFeature`/`BearingToFeature` only between point markers** until polygon/line geometry exists in the host; nearest-point and centroid math for those features is out of scope here.
- **No real GPS / device location** — markers are author-placed or programmatically set; there is no "my location" source. (Not in the spec, but worth noting for map apps.)
- `ShowShadow` is purely cosmetic and may be skipped without behavioral loss.

## Effort Estimate
**L** for `Marker` itself (render branch + SVG pin + anchor math + drag/inverse-projection wiring + 7 methods incl. haversine/bearing in Go + infobox), and it is **gated on the `Map` host (separately XL)**. Standalone, with a Map host already present, the `Marker` work is solidly **L**; counting the mandatory `Map` prerequisite, the realistic combined effort is **XL**.
