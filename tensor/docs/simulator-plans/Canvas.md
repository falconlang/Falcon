# Canvas Simulator Implementation Plan

## Overview

`Canvas` (category **ANIMATION** — "Drawing & Animation" in the AI palette, `nonVisible = false`) is, per the spec's `helpString`, "a two-dimensional touch-sensitive rectangular panel on which drawing can be done and sprites can be moved." It is the drawing surface of App Inventor: it exposes a coordinate space (X = pixels from the left edge, Y = pixels from the top edge), `Draw*` methods for points/lines/circles/arcs/shapes/text, pixel read/write methods, touch/drag/fling events, and per-Canvas paint state (`PaintColor`, `LineWidth`, `FontSize`, `TextAlignment`, `BackgroundColor`, `BackgroundImage`).

It is a **visible container**: it is the AI parent of `Ball` and `ImageSprite`. `canContainDesignComponent('Canvas', 'Ball'|'ImageSprite')` already returns `true` in `design-schema-tree.js` (line 64), and `CANVAS_CHILD_TYPES` is already defined (line 23). Sprites are positioned absolutely within the Canvas via their own `X`/`Y` (and `Z`) coordinates. Canvas itself can be placed on `Form`/`Screen` or in any arrangement, like `Image`.

In the simulator it currently renders as a dashed "unsupported" placeholder because `Canvas` is absent from `SIMULATION_SUPPORTED_TYPES`. (Note: `EVENT_ARGS.Canvas` already exists in `simulation-capabilities.js` at lines 433-437 for `TouchDown`/`TouchUp`/`Touched`, but is dormant until the type is supported.)

## Feasibility Verdict

**Feasible.**

The HTML `<canvas>` 2D context is the direct web-platform analog of the Android `Canvas`, and pointer events (`pointerdown`/`pointermove`/`pointerup`) give exact (x, y) coordinates relative to the element via `getBoundingClientRect()`. Every drawing method maps to a `CanvasRenderingContext2D` call:

- `DrawPoint`/`DrawLine`/`DrawCircle`/`DrawArc`/`DrawShape`/`DrawText`/`DrawTextAtAngle`/`Clear` → `fillRect`/`moveTo`+`lineTo`/`arc`/`fillText`/`rotate`+`fillText`/`clearRect`. All directly implementable.
- `GetPixelColor`/`GetBackgroundPixelColor`/`SetBackgroundPixelColor` → `ctx.getImageData`/`putImageData`. **Caveat:** `getImageData` taints and throws a `SecurityException` if a cross-origin `BackgroundImage` without CORS has been drawn; in this simulator all assets resolve to same-origin blob/object URLs via `resolveAssetUrl`, so the canvas stays untainted and pixel reads work.
- `TouchDown`/`TouchUp`/`Touched`/`Dragged`/`Flung` → derived from a pointer gesture state machine on the canvas element. `TapThreshold` (default 15 px) distinguishes a tap (`Touched`) from a drag (`Dragged`); fling speed/heading/velocity are computed from the last pointer-move delta over elapsed time. All firable in-browser.
- Sprite hit-testing (`touchedAnySprite`, `draggedAnySprite`, `flungSprite`) is computable from child Ball/ImageSprite geometry the simulator already holds in state.

The only honest gaps are device-storage methods: `Save`/`SaveAs` write a PNG/JPEG to the device's external storage. A browser can produce the image (`canvas.toDataURL`/`toBlob`) but cannot write to an arbitrary device path; the realistic approximation is a browser download (or returning a synthetic path string) — these are partial, everything else is fully feasible.

The largest *implementation* dependency is that a fully useful Canvas needs **Ball** and **ImageSprite** child rendering and their motion/animation timer to be meaningful. The drawing surface and touch/drag/fling events are independently implementable today; sprite interaction (`touchedAnySprite` etc.) degrades gracefully to `false` until sprites land.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `BackgroundColor` | `&HFFFFFFFF` (white) | Visual | `colorValue()` → canvas element `background` (or a `fillRect` base layer the `Clear` method preserves). | High |
| `BackgroundImage` | `""` | Visual | `resolveAssetUrl(assets, name)` (Image pattern); draw as a background layer beneath drawings, `background-size: cover`-style or 1:1. Empty → just `BackgroundColor`. | High |
| `PaintColor` | `&HFF000000` (black) | Behavioral | Held in host paint state; used as `ctx.strokeStyle`/`fillStyle` for subsequent `Draw*` calls. Not directly visible. | High |
| `LineWidth` | `2.0` | Behavioral | `ctx.lineWidth` for `DrawLine`/`DrawCircle`(outline)/`DrawArc`. | High |
| `FontSize` | `14.0` | Behavioral | `ctx.font` px size for `DrawText`/`DrawTextAtAngle`. | Medium |
| `TextAlignment` | `1` (center) | Behavioral | Maps to `ctx.textAlign` (`0`=left/`start`, `1`=center, `2`=right/`end`) for drawn text. | Medium |
| `Width` | `""` (auto) | Visual | `sizeStyle()` handles `-1`/`-2`/percent. Canvas with no explicit size in AI is small; render with a sensible default min size and set the backing-store `width`/`height` attributes to the laid-out CSS px so the coordinate space matches. | High |
| `Height` | `""` (auto) | Visual | As above. | High |
| `Left` / `Top` | `""` | Visual (AbsoluteArrangement only) | Already handled via `positionStyle()`; `Left`/`Top` already in `SIMULATION_VISUAL_PROPS`. | Low |
| `ExtendMovesOutsideCanvas` | `False` | Behavioral | When `false`, clamp reported drag coordinates to `[0, Width]`/`[0, Height]`; when `true`, allow negative / over-size coords. | Low |
| `TapThreshold` | `15` | Behavioral | Pixel distance threshold in the pointer gesture state machine separating `Touched` (tap) from `Dragged`. | Medium |
| `Visible` | `True` | Visual | Standard `isSimulationVisible` gate on the branch. | High |
| `Column` / `Row` | (block-only, `rw: invisible`) | n/a | Layout artifacts; not simulated. | — |

`HeightPercent`/`WidthPercent` (write-only) map to the percent size encoding `sizeStyle()` already understands. `BackgroundImageinBase64` (write-only) can be accepted and, since it is already a data URI, used directly as the background source.

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `TouchDown` | `x`, `y` (number) | On `pointerdown` over the canvas. Compute `(x, y)` from `e.clientX/Y` minus `getBoundingClientRect()` top/left. Fires immediately. |
| `TouchUp` | `x`, `y` (number) | On `pointerup`/`pointercancel`, with the release coordinates. |
| `Touched` | `x`, `y`, `touchedAnySprite` (boolean) | On `pointerup` when the total movement since `pointerdown` stayed within `TapThreshold` (a tap). `touchedAnySprite` = whether the down point hit a child Ball/ImageSprite bounding region (`false` until sprites are implemented). |
| `Dragged` | `startX`, `startY`, `prevX`, `prevY`, `currentX`, `currentY`, `draggedAnySprite` (boolean) | On each `pointermove` while the pointer is down once cumulative movement exceeds `TapThreshold`. `startX/Y` = down point, `prevX/Y` = previous move point, `currentX/Y` = this point. `draggedAnySprite` = whether the start point hit a sprite (and, when it did, the sprite should follow the drag and its own `Dragged` fires). |
| `Flung` | `x`, `y`, `speed`, `heading`, `xvel`, `yvel`, `flungSprite` (boolean) | On `pointerup` when the final pointer-move velocity exceeds a fling threshold. `speed` = px/ms from the last move delta / elapsed ms; `heading` = `atan2(-dy, dx)` in degrees (-180..180); `xvel`/`yvel` = component velocities; `flungSprite` = whether the start point was near a sprite. |

All five events are fully firable in a browser — no sensor or permission is involved. `EVENT_ARGS.Canvas` already lists `TouchDown`/`TouchUp`/`Touched`; it must be **extended** to add the `Dragged` and `Flung` arg lists.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `Clear()` | `() → void` | Emit a `canvas-draw` effect `{op:'clear'}`; overlay/renderer `clearRect`s the drawing layer (background color/image preserved by re-painting them). |
| `DrawPoint(x, y)` | `(num, num) → void` | `fillRect(x, y, 1, 1)` (or a `LineWidth`-sized dot) in `PaintColor`. |
| `DrawLine(x1, y1, x2, y2)` | `(num×4) → void` | `beginPath`/`moveTo`/`lineTo`/`stroke` with `PaintColor`, `LineWidth`. |
| `DrawCircle(centerX, centerY, radius, fill)` | `(num×3, bool) → void` | `arc(...,0,2π)`; `fill` → `fill()`, else `stroke()`. |
| `DrawArc(left, top, right, bottom, startAngle, sweepAngle, useCenter, fill)` | `(num×6, bool, bool) → void` | Derive center/radii from the bounding box; `ctx.ellipse`/`arc` with start/sweep in radians (AI degrees, 0 = right, clockwise → no Y-flip needed). `useCenter` → line to center for a sector; `fill` chooses fill/stroke. |
| `DrawShape(pointList, fill)` | `(list, bool) → void` | Parse `((x1 y1)(x2 y2)...)` into points; `moveTo` first, `lineTo` rest, `closePath`; `fill`/`stroke`. |
| `DrawText(text, x, y)` | `(text, num, num) → void` | `ctx.font` from `FontSize`, `ctx.textAlign` from `TextAlignment`, `fillText` in `PaintColor`. |
| `DrawTextAtAngle(text, x, y, angle)` | `(text, num×2, num) → void` | `save`/`translate(x,y)`/`rotate(-angle°)`/`fillText`/`restore`. |
| `GetPixelColor(x, y)` | `(num, num) → number` | `getImageData(x,y,1,1)` → pack RGBA into the AI `&HAARRGGBB` number. Includes drawings, background, **and** sprites (sprites are composited into the canvas in our render). |
| `GetBackgroundPixelColor(x, y)` | `(num, num) → number` | Same, but read from an offscreen background+drawings layer that excludes sprites. Requires keeping the drawing layer separate from the sprite layer. |
| `SetBackgroundPixelColor(x, y, color)` | `(num, num, num) → void` | `putImageData` a single pixel of the unpacked color onto the drawing layer. |
| `Save()` | `() → text` | **Partial.** Browser cannot write to device external storage. Approximation: `canvas.toDataURL('image/png')` → trigger a download via a `canvas-save` effect; return a synthetic filename string (e.g. `Canvas1.png`). Surface a `h.logs`/`h.Unsupported`-style note that the real device path is unavailable. |
| `SaveAs(fileName)` | `(text) → text` | **Partial.** Same as `Save` but honor the requested name/extension (`.png`/`.jpg`/`.jpeg` → `toDataURL` mime). Cannot write to a device path; download instead. Return the requested `fileName`. |

The `Draw*`/`Clear`/`Set*` methods return void and must persist their effect on a retained drawing buffer (drawing is cumulative; the canvas state must survive re-renders and `runEvent` round-trips).

## Implementation Plan

### simulation-capabilities.js

- Add `'Canvas'` to `SIMULATION_SUPPORTED_TYPES`. It is **visible**, so do **not** add it to `SIMULATION_NONVISIBLE_TYPES`.
- Add a defaults block inside `buildSimulationDefaults()`:

```js
Canvas: {
  Visible: true,
  Width: -1,
  Height: -1,
  BackgroundColor: '&HFFFFFFFF',
  BackgroundImage: '',
  PaintColor: '&HFF000000',
  LineWidth: 2,
  FontSize: 14,
  TextAlignment: 1,
  ExtendMovesOutsideCanvas: false,
  TapThreshold: 15,
},
```

- Add new prop names to `SIMULATION_VISUAL_PROPS`: `PaintColor`, `LineWidth`, `ExtendMovesOutsideCanvas`, `TapThreshold`. (`Visible`, `Width`, `Height`, `Left`, `Top`, `BackgroundColor`, `BackgroundImage`, `FontSize`, `TextAlignment` are already present.) `PaintColor`/`LineWidth` must be visual so the renderer's draw-command processing can read current paint state from props; props not in this set are stripped before reaching the renderer.
- `isBooleanProp`: add `ExtendMovesOutsideCanvas`.
- `isNumericProp`: add `LineWidth` and `TapThreshold`. (`Width`/`Height`/`Left`/`Top`/`FontSize`/`TextAlignment` already covered.)
- `coerceSimulationValue`: no special case needed — colors are stored as AI `&H...` strings (the renderer's `colorValue()` converts), numbers go through `coerceNumber`. No `deriveStateFromDesignerProps` branch is required.

### SimulationComponent.svelte

Add a branch modeled on a hybrid of the **Image** branch (asset resolution via `assetName()`/`resolveAssetUrl`/`assetUrl`) and the **arrangement** branches (it hosts children). Sketch:

```svelte
{:else if node.type === 'Canvas'}
  <div
    class="sim-canvas"
    class:sim-unsupported={unsupportedHere}
    style={baseStyle('position: relative; overflow: hidden;', { backgroundImage: false })}
    data-sim-component={node.name}
  >
    {#if assetUrl}
      <img class="sim-canvas-bg" src={assetUrl} alt="" />
    {/if}
    <canvas
      bind:this={canvasEl}
      class="sim-canvas-surface"
      on:pointerdown={canvasPointerDown}
      on:pointermove={canvasPointerMove}
      on:pointerup={canvasPointerUp}
      on:pointercancel={canvasPointerUp}
    ></canvas>
    <div class="sim-canvas-sprites">
      {#each node.children || [] as child (child.pathId || child.name)}
        <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
      {/each}
    </div>
  </div>
```

- **`assetName()`** already returns `props.BackgroundImage` for non-Image types if extended; confirm/extend it to return `props.BackgroundImage` for `Canvas` so `assetUrl` resolves the background. (Background can be an `<img>` layer or drawn into the canvas; an `<img>` layer is simpler and keeps `GetBackgroundPixelColor` honest if we read from the canvas drawing layer separately.)
- **Backing store sizing:** on mount and on size change, set `canvasEl.width`/`canvasEl.height` (attributes) to the element's laid-out CSS pixel size so canvas coordinate space == AI coordinate space (1:1). Use a `ResizeObserver` or react to `props.Width`/`Height`.
- **Pointer gesture state machine** (new `let` vars near `buttonEl`: `canvasEl`, plus gesture state `downPt`, `prevPt`, `lastMoveTs`, `dragging`, `startSprite`): `canvasPointerDown` records the down point (coords from `getBoundingClientRect`), `setPointerCapture`, fires `TouchDown(x,y)`, hit-tests sprites. `canvasPointerMove` (only while down): if cumulative distance > `TapThreshold` mark `dragging`, fire `Dragged(startX,startY,prevX,prevY,curX,curY,draggedAnySprite)` each move, update `prevPt`/`lastMoveTs`/velocity. `canvasPointerUp`: fire `TouchUp(x,y)`; if not `dragging` fire `Touched(x,y,touchedAnySprite)`; if final velocity > fling threshold fire `Flung(...)`. Clamp coords to canvas bounds unless `ExtendMovesOutsideCanvas`. Use `emitEvent(node.name, 'TouchDown', [x,y])` etc. (the multi-arg `emitEvent` path already exists).
- **Draw command processing:** the cumulative drawing buffer must be retained. Reuse the existing `handleComponentActions(actions?.[node.name] ?? {})` reactive path: the Go host emits ordered `canvas-draw` component-action effects (op + params); the renderer replays new draw ops onto the 2D context. Maintain a `drawnCount`/last-applied index so re-renders only apply *new* ops. `Clear` re-paints background then drops the op log. (Alternative: keep the draw op log in host state and replay the full log each render — simpler but O(n); fine for a simulator.)
- **CSS:** `.sim-canvas` is a positioned block (`sizeStyle`), `overflow:hidden`; `.sim-canvas-bg` and `.sim-canvas-surface` are absolutely positioned to fill it; `.sim-canvas-sprites` is an absolute overlay (`pointer-events: none` so canvas pointer events pass through, except sprites re-enable `pointer-events` for their own drag — that interplay is finalized when Ball/ImageSprite land). Add `canvasEl` to the `let` declarations and clean up the `ResizeObserver` in `onDestroy`.

### SimulateOverlay.svelte

**None required for v1.** Draw operations (`Clear`, `Draw*`, `Set*`) are modeled as `component-action` effects (like `open`/`focus`/`cursor-position`) that `applyEffects()` already turns into `actionTokens`; the renderer consumes them via the existing `actions` prop / `handleComponentActions` path. For the v1 cumulative-buffer approach, draw ops can also be carried via `statePatch` (e.g. an incrementing op log on the Canvas component) so no new overlay handler is needed. `Save`/`SaveAs` would need a small `canvas-save` effect that triggers a download; that is optional polish — for v1 they can be `h.Unsupported` or a logged no-op.

### simulation_wasm.go

A `Canvas` type needs real handlers (the generic store path is insufficient for paint state, draw commands, computed pixel reads, and gesture-derived events):

- **`GetProperty`:** add `case "Canvas"`. Most props (`PaintColor`, `LineWidth`, `FontSize`, `TextAlignment`, `BackgroundColor`, etc.) are plain reads from `h.state` (the generic fallthrough already handles these). `GetPixelColor`/`GetBackgroundPixelColor` are **methods**, not properties, so no computed-property work here. Width/Height read back the stored layout size.
- **`SetProperty`:** add `case "Canvas"` only if a write needs side effects. `BackgroundColor`/`BackgroundImage` changes should patch state (renderer repaints background). `BackgroundImageinBase64` (write-only) maps to setting `BackgroundImage` with the data URI. `PaintColor`/`LineWidth`/`FontSize`/`TextAlignment`/`ExtendMovesOutsideCanvas`/`TapThreshold` are plain `h.setProperty`. No `*Changed` event on Canvas property writes.
- **`CallMethod`:** add `case "Canvas": return h.callCanvasMethod(...)`. For each `Draw*`/`Clear`/`Set*` method, append a `canvas-draw` effect via `componentActionWith(componentName, "draw", {op, ...params})` (or push onto an op-log in `h.statePatch`). The host reads current `PaintColor`/`LineWidth`/`FontSize`/`TextAlignment` from `h.state` to bake into each effect so the renderer does not need to re-derive paint state. `GetPixelColor`/`GetBackgroundPixelColor` **cannot** be computed host-side (no canvas in Go); the realistic approach is to maintain a host-side software model of drawn pixels, OR return a best-effort value (e.g. the `BackgroundColor` for an empty canvas) and `h.logs` a note — full fidelity would require the frontend to read back via a round-trip the current synchronous method API does not support. Document this limitation. `Save`/`SaveAs` → emit a `canvas-save` effect and return the filename string (partial). `DrawArc`/`DrawShape` parse their numeric/list args (`argList`-style helper).
- **Events:** Canvas events (`TouchDown`/`TouchUp`/`Touched`/`Dragged`/`Flung`) are fired **frontend-side** via `emitEvent`/`eventRunner` from the pointer state machine — the host does not synthesize them. No `runEvent` calls are needed on the host for Canvas input.
- Recompile with `npm run build:wasm`.

### design-schema-tree.js

**No change required.** Containment is already encoded: `CANVAS_CHILD_TYPES = {Ball, ImageSprite}` and `canContainDesignComponent('Canvas', child)` returns `true` for those (line 64); `canContainDesignComponent('Canvas', other)` returns `false` (line 72). `designTreeToInitialState()` will merge the new `SIMULATION_DEFAULTS.Canvas` automatically, and `unsupportedSimulationComponents()` will stop flagging Canvas once it is in `SIMULATION_SUPPORTED_TYPES`. (Ball/ImageSprite remain unsupported until implemented and will still render as placeholders inside the Canvas.)

## Dependencies & Ordering

- **No external libraries.** Native `<canvas>` 2D context + pointer events — no Leaflet/MapLibre-class dependency.
- **Implementation order:** the Canvas drawing surface + touch/drag/fling events are independently implementable and valuable on their own (drawing apps, signature pads, touch demos). **Ball** and **ImageSprite** child components should be implemented after Canvas (Canvas is their required parent and provides the coordinate space, hit-testing entry points, and `touchedAnySprite`/`draggedAnySprite`/`flungSprite` plumbing). Until they land, sprite-hit booleans degrade to `false` and the `sim-canvas-sprites` overlay is empty — no functional breakage. So: **Canvas first, then Ball/ImageSprite.**

## Web-Platform Limitations & Fidelity Caveats

- **`Save`/`SaveAs` cannot write to device storage.** A browser can generate the image (`toDataURL`/`toBlob`) and trigger a download, but there is no arbitrary device file path. The returned path string is synthetic.
- **`GetPixelColor`/`GetBackgroundPixelColor` are awkward across the Go↔JS boundary.** Pixel reads are inherently a frontend (`getImageData`) operation, but `CallMethod` returns synchronously in Go without a canvas. v1 returns a best-effort value (background color / logged note); exact pixel fidelity needs a host-side software pixel model or an async round-trip the current API does not provide.
- **Canvas tainting:** if a future cross-origin `BackgroundImage` without CORS headers were drawn, `getImageData` would throw. In-simulator assets are same-origin blob URLs, so this does not arise in practice — but a user-supplied remote URL would.
- **Fling thresholds are heuristic.** `speed`/`heading`/`xvel`/`yvel` are derived from the last pointer-move delta and timing; exact values will differ from Android's `VelocityTracker`. Direction/order-of-magnitude are faithful; precise numbers are not.
- **Sprite interaction is gated on Ball/ImageSprite.** `touchedAnySprite`/`draggedAnySprite`/`flungSprite` are `false`, and sprite-following-drag does not occur, until those components are implemented.
- **Sub-pixel/anti-aliasing and font metrics** differ between the browser 2D context and Android's skia-backed Canvas; drawn text width/position and line rendering will not be pixel-identical.

## Effort Estimate

**L.** The drawing surface (canvas element + backing-store sizing + ~8 `Draw*` ops replayed from host effects), the pointer gesture state machine (tap/drag/fling discrimination with `TapThreshold`, 5 events with full arg vectors), background-image layering, and the host `callCanvasMethod` with paint-state baking add up well beyond M. The cumulative-draw-buffer/op-replay design and the `GetPixelColor` boundary problem are the genuinely tricky parts; no external dependency and already-handled containment keep it out of XL. Ball/ImageSprite are separate follow-on efforts.
