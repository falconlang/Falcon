# FeatureCollection Simulator Implementation Plan

## Overview

`FeatureCollection` (category **MAPS**, `nonVisible = false`) is, per the spec's `helpString`, "a FeatureCollection contains one or more map features as a group. Any events fired on a feature in the collection will also trigger the corresponding event on the collection object. FeatureCollections can be loaded from external resources as a means of populating a Map with content."

It is a **visible container**, but a peculiar one: it has no visual surface of its own. It is a logical grouping/event-aggregation layer that sits *inside* a `Map` and *holds* feature primitives. Its only rendering is whatever its child features (`Circle`, `LineString`, `Marker`, `Polygon`, `Rectangle`) draw on the parent map's coordinate space. Its designer `Left`/`Top`/`Visible`/`Width`/`Height` properties exist on the block surface but, in App Inventor, a FeatureCollection occupies the full bounds of its parent Map; the geometry that matters lives on the children.

Its defining behaviors are:
- **Event re-dispatch:** when a child feature fires `FeatureClick`/`FeatureDrag`/`FeatureLongClick`/`FeatureStartDrag`/`FeatureStopDrag`, the *same* event also fires on the FeatureCollection with the child component as the `feature` argument. This is the App Inventor "bubble to the collection" contract.
- **GeoJSON ingestion:** `FeaturesFromGeoJSON` (write-only) and `LoadFromURL` parse GeoJSON and materialize features, raising `GotFeatures` / `LoadError` / `ErrorLoadingFeatureCollection`.
- **Feature list:** `Features` (read-write list) is the set of child features, including any created via `FeatureFromDescription`.

**Container relationship.** It is a **child of `Map`** and a **container of the feature primitives** (`Circle`, `LineString`, `Marker`, `Polygon`, `Rectangle`) — a strict subset of Map's children (it cannot nest another FeatureCollection or a Map). This is already encoded in `design-schema-tree.js`: `FEATURE_COLLECTION_CHILD_TYPES = {Circle, LineString, Marker, Polygon, Rectangle}` (line 25) and `canContainDesignComponent('Map','FeatureCollection') === true`, `canContainDesignComponent('FeatureCollection', <featurePrimitive>) === true` (lines 65-68).

It currently renders as a dashed "unsupported" placeholder because `FeatureCollection` is absent from `SIMULATION_SUPPORTED_TYPES`.

## Feasibility Verdict

**Partially feasible (in a browser simulator) — and entirely gated on `Map`.**

The FeatureCollection itself is trivial web-platform-wise: it draws nothing, so there is no rendering obstacle. The honest constraints are upstream and around GeoJSON:

1. **It has no value without `Map`, which is not yet implemented and is the hard dependency.** FeatureCollection only matters as a layer inside a Map: a feature's screen position is `(latitude, longitude) → pixel` via the map projection at the current center/zoom. With no Map there is no coordinate space, no tiles, and nowhere for child features to render. The simulator architecture note is explicit that map tiles need a library like Leaflet/MapLibre "or are a heavy dependency." So FeatureCollection cannot be meaningfully demonstrated until Map (and the feature primitives) exist. This is a *dependency* limitation, not a web-platform one — Leaflet/MapLibre render tiles and vector overlays in a browser perfectly well.
2. **Event re-dispatch is fully feasible** once child features fire their own events: the collection subscribes to child `Feature*` events and re-emits them on itself with the child as `feature`. Pure JS plumbing, no platform limit.
3. **GeoJSON parsing is feasible** in-browser (`JSON.parse` + a small GeoJSON-geometry walker). `LoadFromURL` is **partially feasible**: a `fetch()` can retrieve GeoJSON from a CORS-permitting URL, but cross-origin URLs without CORS headers will fail (a real Android device has no such restriction). The realistic approximation: attempt `fetch`; on CORS/network failure raise `LoadError` with a synthetic responseCode, exactly as the AI contract allows. `FeaturesFromGeoJSON` (a literal string, no network) is fully feasible.
4. **`FeatureFromDescription`** (build a feature from a `((tag value)...)` list) is feasible as pure data transformation, but the *returned* component only renders if the feature primitives render — i.e. again gated on Map + primitives.

So: the FeatureCollection's own logic (grouping, event bubbling, GeoJSON parse, feature-list management) is feasible and cheap; its visible payoff is blocked until `Map`, `Marker`, `Circle`, `LineString`, `Polygon`, `Rectangle` land. The realistic simulated approximation for a standalone v1 is a **non-rendering logical layer** that aggregates child events and manages the feature list, becoming visually meaningful the moment Map renders.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Visible` | `True` | Behavioral (gates whether child features show) | When `false`, the renderer hides all child features in this collection. There is no surface of its own to hide. | Medium |
| `Features` | (read-write list) | Behavioral | The list of child feature components (designer children + any added via `FeatureFromDescription`). Held in host state; getter returns component refs. | Medium |
| `FeaturesFromGeoJSON` | `""` (write-only) | Behavioral | Parse the GeoJSON string and materialize features into `Features`. Invalid GeoJSON → `ErrorLoadingFeatureCollection` (helpString) with `url = <string>`. | Medium |
| `Source` | `""` (read-only) | Behavioral | The URL the collection was loaded from (set by `LoadFromURL`), else empty string. Plain stored read. | Low |
| `Left` / `Top` | `""` | n/a in browser | Block/AbsoluteArrangement layout artifacts. A FeatureCollection occupies the parent Map's full bounds; these are accepted and stored but not honored visually. | Low |
| `Width` / `Height` | (read-write number) | n/a in browser | Same — the collection has no independent box; it spans the parent Map. Accept/store, do not lay out. | Low |
| `WidthPercent` / `HeightPercent` | (write-only) | n/a | Accept and no-op for the same reason. | Low |
| `Column` / `Row` | (block-only, `rw: invisible`) | n/a | Layout artifacts; not simulated. | — |

Note: none of the size/position properties can be *honored* in a browser the way AI documents them, because in AI itself they are effectively inert for a map layer — the collection always covers its Map. They are accepted (so designer schemas parse) but visually no-ops.

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `FeatureClick` | `feature` (component) | Re-dispatched when any child feature in this collection fires its own `FeatureClick` (i.e. the child marker/shape was clicked). The simulator listens to child-feature click events and emits `FeatureClick` on the collection with the child as `feature`. **Requires child feature primitives to be implemented and emitting clicks.** |
| `FeatureLongClick` | `feature` (component) | Same bubbling, on a child long-press (pointer held > ~500 ms). Requires child primitives. |
| `FeatureDrag` | `feature` (component) | Re-dispatched on each child drag move. Requires draggable child primitives (Marker `EnableDrag`). |
| `FeatureStartDrag` | `feature` (component) | Re-dispatched when a child feature drag begins. Requires draggable child primitives. |
| `FeatureStopDrag` | `feature` (component) | Re-dispatched when a child feature drag ends. Requires draggable child primitives. |
| `GotFeatures` | `url` (text), `features` (list) | Fired after `LoadFromURL` (or `FeaturesFromGeoJSON`) successfully parses GeoJSON. `features` is the list of (key, value) pairs parsed from the document. Feasible in-browser via `fetch` + `JSON.parse`. |
| `LoadError` | `url` (text), `responseCode` (number), `errorMessage` (text) | Fired when `LoadFromURL` fails — network error, non-2xx HTTP, **CORS rejection (browser-specific)**, or invalid GeoJSON. `responseCode` will be the HTTP status when available, else a synthetic code (e.g. `0` for network/CORS). |

All seven events are firable in a browser. The five `Feature*` events are **gated on the feature primitives existing** (nothing emits child clicks/drags until `Marker`/`Circle`/etc. are implemented); `GotFeatures`/`LoadError` are independent of the primitives and can be exercised purely from GeoJSON parsing.

There is no `ErrorLoadingFeatureCollection` in the spec's `events` array (it is referenced only in the `FeaturesFromGeoJSON` helpString); on invalid literal GeoJSON, surface it as a `LoadError`-style notice or an `h.Unsupported`/log entry rather than inventing an event.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `LoadFromURL(url)` | `(text) → void` | **Partial.** `fetch(url)` the GeoJSON; on success parse and materialize features, set `Source = url`, fire `GotFeatures(url, features)`. On failure fire `LoadError(url, responseCode, errorMessage)`. Cross-origin URLs **without CORS headers fail in a browser** (a real device does not have this limit) — those surface as `LoadError` with a synthetic `responseCode` (0) and a CORS message. |
| `FeatureFromDescription(description)` | `(list) → component` | **Partial / feasible as data.** Parse the `((tag value)...)` GeoJSON-ish description into a feature spec, append to `Features`, return a synthetic feature component reference. The returned feature **only renders once Map + the matching primitive (Marker/LineString/Polygon) are implemented**; until then it is a data-only object. |

`FeaturesFromGeoJSON` (write-only property, listed under properties above) behaves like `LoadFromURL` minus the network step: parse the literal string, materialize features, fire `GotFeatures("", features)` on success or signal an error on invalid GeoJSON.

## Implementation Plan

This plan describes FeatureCollection's own surface. It is only **renderable** once `Map` (parent coordinate space + tiles) and the feature primitives exist; the logic below can be landed first and degrade to a non-rendering aggregation layer.

### simulation-capabilities.js

- Add `'FeatureCollection'` to `SIMULATION_SUPPORTED_TYPES`. It is **visible** (a layer inside the visible Map), so do **not** add it to `SIMULATION_NONVISIBLE_TYPES`.
- Add a defaults block inside `buildSimulationDefaults()`:

```js
FeatureCollection: {
  Visible: true,
  Features: [],
  Source: '',
  FeaturesFromGeoJSON: '',
  Left: -1,
  Top: -1,
  Width: -2,
  Height: -2,
},
```

- Add new prop names to `SIMULATION_VISUAL_PROPS`: `Features`, `Source`, `FeaturesFromGeoJSON`. (`Visible`, `Width`, `Height`, `Left`, `Top` are already present.) Props not in this set are stripped before reaching the renderer; the renderer needs `Features`/`Visible` to decide which child features to show.
- `isBooleanProp`: nothing new (`Visible` already covered).
- `isNumericProp`: nothing strictly required — `Left`/`Top`/`Width`/`Height` are already covered, and they are visual no-ops here anyway.
- `coerceSimulationValue`: `Features` is a list — let it pass through as an array (the existing `Array.isArray(value)` branch returns it unchanged for non-`Elements` props). `Source`/`FeaturesFromGeoJSON` are strings (default pass-through). No new coercion case is strictly required; optionally add a `deriveStateFromDesignerProps` branch so that a designer-supplied `FeaturesFromGeoJSON` is parsed into an initial `Features` array at load time (mirrors how `ElementsFromString` derives `Elements`).

### SimulationComponent.svelte

FeatureCollection draws nothing itself; it is a transparent wrapper that forwards its children to the Map's coordinate space and re-emits their `Feature*` events. Add a branch modeled on the **arrangement** branches (which host children via `<svelte:self>`), but with `display: contents`-style passthrough so it adds no box:

```svelte
{:else if node.type === 'FeatureCollection'}
  <div
    class="sim-feature-collection"
    class:sim-hidden={!boolValue(props.Visible, true)}
    data-sim-component={node.name}
    style="display: contents;"
  >
    {#each node.children || [] as child (child.pathId || child.name)}
      <svelte:self
        node={child} {state} {unsupported} {assets} {actions} {eventRunner}
        parentType={node.type}
        on:event={featureBubble}
        on:property={childEvent}
        on:interaction={childEvent}
      />
    {/each}
  </div>
```

- **Event bubbling (`featureBubble`):** wrap the child `on:event` handler so that when a child emits one of `FeatureClick`/`FeatureLongClick`/`FeatureDrag`/`FeatureStartDrag`/`FeatureStopDrag` (or the child primitive's `Click`/`LongClick`/`Drag` events that map to those), it (a) forwards the original child event upward via `childEvent` and (b) additionally calls `emitEvent(node.name, 'FeatureClick', [childComponentName])` etc. so the collection-level event fires with the child as the `feature` arg. This is the App Inventor "events bubble to the collection" contract.
- **`Visible` gate:** when `Visible === false`, hide the children (e.g. `class:sim-hidden` toggling `display:none`, or simply don't render the `{#each}`). The collection has no surface of its own to hide.
- **No box of its own:** use `display: contents` (or render children directly) so the collection contributes no layout — child features position themselves on the parent Map. Do **not** call `baseStyle()`/`sizeStyle()`; the size/position props are inert here.
- **Reuse:** the `<svelte:self>` child-forwarding pattern from the arrangement branches; `boolValue()` for the `Visible` gate; `emitEvent()` for the re-dispatch. No new CSS beyond a `.sim-hidden { display: none; }` helper (if not already present).
- Until the feature primitives are implemented, the children render as their own dashed "unsupported" placeholders inside the Map — no functional breakage, and the bubbling wiring is inert (nothing emits child `Feature*` events yet).

### SimulateOverlay.svelte

**None required for v1.** FeatureCollection has no modal dialogs or toasts. `GotFeatures`/`LoadError` are ordinary events dispatched to the Go host via `eventRunner`, not overlay effects. (If `LoadFromURL` fetches client-side and we want a visible "load failed" notice, that could reuse the existing notice path — but it is optional polish, not required.)

### simulation_wasm.go

A `FeatureCollection` type needs light handlers; most of its property surface is plain store reads/writes, but GeoJSON ingestion and event re-dispatch need code:

- **`GetProperty`:** add `case "FeatureCollection"`. `Features` returns the stored feature list; `Source` returns the stored source URL (empty if never loaded). `Visible`/`Left`/`Top`/`Width`/`Height` are plain `h.state` reads (generic fallthrough suffices for these).
- **`SetProperty`:** add `case "FeatureCollection"` for the side-effecting writes. `FeaturesFromGeoJSON` (write-only) → parse the GeoJSON string host-side; on success update `Features` (statePatch) and `h.runEvent(name, "GotFeatures", "", featuresList)`; on parse failure log/`h.Unsupported` an invalid-GeoJSON note (no `ErrorLoadingFeatureCollection` event exists in the spec). `Visible` → patch state (renderer hides children). `WidthPercent`/`HeightPercent`/`Width`/`Height`/`Left`/`Top` → plain `h.setProperty` (visually inert). `Features` (read-write) → store the provided list.
- **`CallMethod`:** add `case "FeatureCollection": return h.callFeatureCollectionMethod(...)`.
  - `LoadFromURL(url)` — **decision point:** either (a) `fetch` client-side in the overlay and post the result back, or (b) the host attempts the fetch via the WASM `fetch` bridge if one exists. Simpler v1: emit a `feature-load` effect carrying the URL; the overlay performs `fetch`, parses, and posts `GotFeatures`/`LoadError` back through `eventRunner`. On success set `Source`. On CORS/network/HTTP failure raise `LoadError(url, responseCode, errorMessage)`.
  - `FeatureFromDescription(description)` — parse the description list into a feature spec, append to `Features`, return a synthetic feature ref. Data-only until primitives render.
- **Event re-dispatch:** the bubbling of child `Feature*` events to the collection can be done **frontend-side** (the `featureBubble` wrapper in SimulationComponent.svelte) so the host need not synthesize it — but the host must still deliver the collection-level event to user blocks via `h.runEvent`. Whichever side originates it, the `feature` arg is the child component name. The five `Feature*` events do not fire until the feature primitives emit child events.
- Recompile with `npm run build:wasm` (`GOOS=js GOARCH=wasm`).

### design-schema-tree.js

**No change required.** Containment is already fully encoded: `MAP_CHILD_TYPES` includes `FeatureCollection`, `FEATURE_COLLECTION_CHILD_TYPES = {Circle, LineString, Marker, Polygon, Rectangle}` (line 25), and `canContainDesignComponent` already returns `true` for `Map → FeatureCollection` and for `FeatureCollection → <feature primitive>`, while rejecting a nested FeatureCollection/Map (lines 65-72). `designTreeToInitialState()` will merge the new `SIMULATION_DEFAULTS.FeatureCollection` automatically, and `unsupportedSimulationComponents()` will stop flagging it once it is in `SIMULATION_SUPPORTED_TYPES`.

## Dependencies & Ordering

- **Hard dependency: `Map`.** A FeatureCollection has no coordinate space, no tiles, and no way to position child features without a rendered Map. **Map must be implemented first.** Map itself needs a tile/vector library — **Leaflet** or **MapLibre GL** — which is the single significant external dependency for this whole MAPS family (the architecture note flags map tiles as "a heavy dependency").
- **Feature primitives next:** `Marker`, `Circle`, `LineString`, `Polygon`, `Rectangle`. Until at least one exists, FeatureCollection has no children to group and no child events to bubble; the five `Feature*` events stay inert. `Marker` is the most common and the natural first primitive.
- **FeatureCollection's own logic (this plan)** can be landed in parallel/early as a non-rendering aggregation layer — its event-bubbling wrapper and GeoJSON parsing are independent of tiles — but it produces no *visible* result until Map + ≥1 primitive land.
- **Realistic order:** `Map` (with Leaflet/MapLibre) → `Marker` (+ other primitives) → `FeatureCollection` becomes meaningful. Implementing FeatureCollection before Map yields only the inert logical layer.

## Web-Platform Limitations & Fidelity Caveats

- **No payoff without Map.** This is the dominant caveat: FeatureCollection is invisible and inert until a Map with tiles renders beneath it and ≥1 feature primitive exists. This is a dependency/ordering reality, not a hard browser limit (Leaflet/MapLibre render fine in-browser).
- **`LoadFromURL` is subject to CORS.** A real Android device fetches any GeoJSON URL; a browser can only fetch cross-origin URLs that send permissive CORS headers. Non-CORS URLs fail and surface as `LoadError` with a synthetic `responseCode` (0). This diverges from device behavior for the same URL.
- **`FeatureFromDescription` returns a data-only feature** until the matching primitive renders; the returned component reference will not draw on the map until `Marker`/`LineString`/`Polygon`/etc. are implemented.
- **Size/position properties are inert.** `Left`/`Top`/`Width`/`Height`/`WidthPercent`/`HeightPercent` are accepted and stored but never honored — the collection always spans its parent Map, matching AI's effective behavior but offering no independent box to inspect.
- **`Feature*` event fidelity depends on primitive input mapping.** Long-press timing, drag thresholds, and click hit-testing are defined by the child primitives' pointer handling; the collection only mirrors whatever they emit, so any imprecision there propagates to the collection-level events.
- **`ErrorLoadingFeatureCollection` has no event.** Invalid literal GeoJSON (via `FeaturesFromGeoJSON`) is mentioned only in a helpString; the simulator surfaces it as a log/notice rather than a real event, since it is absent from the spec's `events`.

## Effort Estimate

**M** (for FeatureCollection's own surface) — but **only after the L/XL `Map` + primitives prerequisites land**. The collection itself is a thin passthrough container plus event re-dispatch wiring (`featureBubble`), a GeoJSON parser shared with `LoadFromURL`/`FeaturesFromGeoJSON`, and `GotFeatures`/`LoadError` plumbing — all modest. The real cost is upstream: `Map` (Leaflet/MapLibre integration, projection, tiles) is **L–XL** and the feature primitives are several **M** efforts; FeatureCollection contributes nothing visible until they exist.
