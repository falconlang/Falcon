# Tensor IDE Simulator — Unimplemented Visible Components: Implementation Plans

**Generated:** 2026-06-04
**Scope:** every *visible* App Inventor component that the simulator does not yet render (currently shown as a dashed "unsupported" placeholder). Non-visible components (sensors, connectivity, media capture, storage, etc.) are out of scope for this pass.
**Components planned:** 23 — **4 feasible, 19 partially feasible, 0 impossible.**

Each component has a detailed plan in this directory covering: overview, feasibility verdict, properties/events/methods to simulate, a file-by-file implementation plan against the simulator architecture, dependencies/ordering, web-platform caveats, and an effort estimate.

> The simulator architecture (where each piece lives) is summarized in every plan. In short: add the type to `SIMULATION_SUPPORTED_TYPES` + a defaults block + visual-prop allowlist entries in `src/lib/simulation-capabilities.js`; add a render branch + CSS + DOM-event wiring in `src/lib/SimulationComponent.svelte`; add host get/set/method/event/effect handling in `lang/simulation_wasm.go` (rebuild with `npm run build:wasm`); use `src/lib/SimulateOverlay.svelte` only for dialogs/effects. Containment for Canvas/Chart/Map families is **already** encoded in `src/lib/design-schema-tree.js`.

## All components

| Component | Category | Feasibility | Effort | Depends on | Verdict |
|-----------|----------|-------------|--------|------------|---------|
| [CircularProgress](./CircularProgress.md) | User Interface | ✅ Feasible | S | — | Pure-CSS rotating ring tinted by `Color`; no events/methods, generic store path. |
| [LinearProgress](./LinearProgress.md) | User Interface | ✅ Feasible | S | — | Native `<progress>` (determinate) + CSS sweep (indeterminate); tiny Go handler for `Progress`/`ProgressChanged`. |
| [TableArrangement](./TableArrangement.md) | Layout | ⚠️ Partial | S | — | CSS-Grid container; only gap is the schema has no per-child Row/Column, so children fill cells in document order. |
| [EmailPicker](./EmailPicker.md) | Social | ⚠️ Partial | S | TextBox | A single-line TextBox clone reusing the existing path; only the device-contacts autocomplete dropdown has no browser equivalent. |
| [ImagePicker](./ImagePicker.md) | Media | ⚠️ Partial | S | Button, ListPicker | ButtonBase picker + real `<input type=file>` chooser; `Selection` can't be a real device path. |
| [FilePicker](./FilePicker.md) | Media | ⚠️ Partial | S | — | Button-clone picker with a real file chooser; `Selection` is a blob URL, and Pick Directory/New File have no portable web API. |
| [ContactPicker](./ContactPicker.md) | Social | ⚠️ Partial | S | ListPicker, ImagePicker | Button renders fully; pick flow approximated with a seeded mock-contact roster (no browser contacts API). |
| [PhoneNumberPicker](./PhoneNumberPicker.md) | Social | ⚠️ Partial | S | Button, ListPicker, Spinner | Same as ContactPicker; `ContactUri`/`Picture` can't be honored. |
| [VideoPlayer](./VideoPlayer.md) | Media | ✅ Feasible | M | — | Maps cleanly to `<video>` (controls, play/pause/seek/volume/fullscreen, ended/error); codec/autoplay caveats. |
| [WebViewer](./WebViewer.md) | User Interface | ⚠️ Partial | M | — | `<iframe>`; loading/sizing/app-driven nav work, but X-Frame-Options/CSP, history reads, and the WebViewString bridge are blocked by browser security. |
| [Canvas](./Canvas.md) | Drawing & Animation | ✅ Feasible | L | (parent of Ball, ImageSprite) | Native `<canvas>` 2D + pointer-gesture machine; needs a frontend animation/tick loop; `GetPixelColor`/`Save` are partial. |
| [Ball](./Ball.md) | Drawing & Animation | ⚠️ Partial | M | **Canvas**, tick loop | Static circle + touch/drag/fling work; autonomous motion/bounce/collision need a tick subsystem. |
| [ImageSprite](./ImageSprite.md) | Drawing & Animation | ⚠️ Partial | L | **Canvas**, Ball | Render + touch/drag/fling map cleanly; gated on Canvas host + shared collision/edge physics. |
| [Chart](./Chart.md) | Charts | ⚠️ Partial | XL | ChartData2D, Trendline | Hand-rolled SVG (line/scatter/area/bar/pie) + `EntryClick`; external data sources (CloudDB/Web/Spreadsheet/DataFile) unsupported. |
| [ChartData2D](./ChartData2D.md) | Charts | ⚠️ Partial | XL | **Chart**, TinyDB | (x,y) series; static + block-appended data render, but live sensor/DB/file/web sources are impossible client-side. |
| [Trendline](./Trendline.md) | Charts | ⚠️ Partial | L | **Chart, ChartData2D** | OLS regression math is trivial in JS; strictly blocked on Chart + ChartData2D landing first. |
| [Map](./Map.md) | Maps | ⚠️ Partial | XL | **Leaflet/MapLibre** + feature children | Slippy map with tiles/pan/zoom/point-events; rotation/compass/GPS have no browser sensor. |
| [Marker](./Marker.md) | Maps | ⚠️ Partial | L | **Map** | Pin/infobox/drag/distance renderable; gated on a Map host (projection + viewport). |
| [Circle](./Circle.md) | Maps | ⚠️ Partial | L | **Map**, Leaflet | Geographic disc; trivial shape but blocked on a Map surface; distance methods honest standalone. |
| [LineString](./LineString.md) | Maps | ⚠️ Partial | L | **Map**, FeatureCollection | SVG polyline + click/long-click; blocked on a Map projection/viewport. |
| [Polygon](./Polygon.md) | Maps | ⚠️ Partial | L | **Map**, FeatureCollection | SVG path + click/longclick/infobox/centroid; everything positional blocked on a Map surface. |
| [Rectangle](./Rectangle.md) | Maps | ⚠️ Partial | L | **Map**, Leaflet | Geographic box; renderable as `L.rectangle` but blocked on a Map; edge round-trip + distance math honest without it. |
| [FeatureCollection](./FeatureCollection.md) | Maps | ⚠️ Partial | M | **Map** + feature primitives | Non-rendering logical layer that groups features and re-dispatches their events; no payoff until Map + primitives exist. |

## Recommended build order

The work splits cleanly into **independent quick wins** and **three container subsystems** that should each be built parent-first.

### Tier 1 — Independent quick wins (no new infrastructure, reuse existing patterns)
`CircularProgress`, `LinearProgress`, `TableArrangement`, `EmailPicker`, `FilePicker`, `ImagePicker`, `ContactPicker`, `PhoneNumberPicker`, `VideoPlayer`, `WebViewer`.

These have no inter-dependencies (the listed "depends on" are *already-implemented* components whose patterns they reuse) and need no shared infrastructure. `EmailPicker` is nearly free — it is a `TextBox` with the type added to the existing guards. The four pickers (`ImagePicker`/`FilePicker`/`ContactPicker`/`PhoneNumberPicker`) share a "button that opens a picker, then fires `BeforePicking`/`AfterPicking` and sets `Selection`" pattern worth factoring once; the two contact pickers also share a seeded **mock-contact roster** (browsers have no contacts API).

### Tier 2 — Canvas subsystem  →  `Canvas`, then `Ball`, `ImageSprite`
Build `Canvas` first (it is the coordinate space + draw surface). **Key shared infrastructure:** the simulator is currently purely event-driven with no `setInterval`/`requestAnimationFrame`/Clock tick, so autonomous sprite motion, `Bounce`/`EdgeReached`, and per-frame collision (`CollidedWith`/`CollidingWith`) require a new **frontend animation/tick loop** (the synchronous Go host cannot run a wall-clock loop). Without it, `Ball`/`ImageSprite` are limited to static placement + touch/drag/fling.

### Tier 3 — Chart subsystem  →  `Chart`, then `ChartData2D`, `Trendline`
Build `Chart` (the SVG plot surface + axes) first; `ChartData2D` (data carrier) and `Trendline` (regression overlay) have no standalone surface. Feasible parts: static + block-driven data, OLS regression in JS, `EntryClick` hit-testing. **Not feasible:** live data-source binding (`Source`/`ImportFromCloudDB`/`Web`/`Spreadsheet`/`DataFile`) — those bind to non-visible components with no browser backing and must be marked `Unsupported` (only `ImportFromList`/`ImportFromTinyDB`/`AddEntry` etc. work).

### Tier 4 — Map subsystem  →  `Map`, then `Marker`/`Circle`/`LineString`/`Polygon`/`Rectangle`/`FeatureCollection`
Build `Map` first. **Key dependency:** this is the first simulator component to need a heavy third-party library — **Leaflet** (or MapLibre GL) + its CSS for tiles/pan/zoom + a lazy-load/teardown lifecycle. The six feature children then render as Leaflet layers / SVG overlays. **Not feasible:** `EnableRotation`/`Rotation` (Leaflet can't rotate the base layer), `ShowUser`/user-GPS (no device sensor), and `Save` to a device path; `LoadFromURL` is CORS-constrained.

## Cross-cutting themes

- **No native device APIs (the recurring "partial" cause):** contacts, gallery/MediaStore, real filesystem paths/content URIs, GPS/compass, and camera don't exist in the browser sandbox. Affected components stay functional via realistic stand-ins (a real `<input type=file>` chooser, a seeded mock-contact roster, a synthetic map center) and surface honest `Unsupported` markers / empty values where there is no analogue.
- **No tick/animation loop yet:** the biggest *infrastructure* gap. Adding a frontend `requestAnimationFrame`/interval loop unlocks the Canvas animation family (and would also enable a future `Clock` component).
- **First external dependency:** the Map family is the only set that requires bundling a third-party library (Leaflet/MapLibre).
- **Container families are already wired for containment:** `design-schema-tree.js` `canContainDesignComponent` already permits Canvas→Ball/ImageSprite, Chart→ChartData2D/Trendline, and Map(/FeatureCollection)→feature types, so no schema changes are needed to *place* these — only rendering/host work.
