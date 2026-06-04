# ImageSprite Simulator Implementation Plan

## Overview

`ImageSprite` (category **ANIMATION** / "Drawing and Animation", `nonVisible = false`) is a visible sprite that lives **on a `Canvas`**. Per the spec's `helpString`, it is "a 'sprite' that can be placed on a `Canvas`, where it can react to touches and drags, interact with other sprites (`Ball`s and other `ImageSprite`s) and the edge of the Canvas, and move according to its property values. Its appearance is that of the image specified in its `Picture` property." It supports autonomous motion: setting `Speed` (pixels/interval), `Interval` (ms), `Heading` (degrees CCW from +x), and `Enabled = True` makes the sprite move every interval; `Rotates = True` rotates the image to match `Heading`. The sprite has an `origin` (default top-left, controlled by `OriginX`/`OriginY`/`MarkOrigin`) about which it is positioned (`X`/`Y`) and rotated.

Container relationship: **child of `Canvas`** (the parent), alongside `Ball`. `canContainDesignComponent` already encodes `Canvas -> { Ball, ImageSprite }` via `CANVAS_CHILD_TYPES`, so containment is wired — but **`Canvas` itself is not yet a supported simulated type**. ImageSprite has no children (it is a leaf in the design tree, but only valid inside a Canvas).

In the simulator it currently renders as a dashed "unsupported" placeholder because `ImageSprite` (and `Canvas`) are absent from `SIMULATION_SUPPORTED_TYPES`, so `unsupportedSimulationComponents()` flags them.

## Feasibility Verdict

**Partially feasible — and strictly blocked on a `Canvas` implementation that does not yet exist.**

The rendering and motion model is a clean fit for the browser: an absolutely-positioned `<img>` (or DOM node) inside the Canvas's positioned coordinate space, with `transform: translate()` + `rotate()`, and a `requestAnimationFrame`/interval-driven motion loop, reproduces position, heading, rotation, and the autonomous `Speed`/`Interval` drift very accurately. Touch, drag, and fling gestures map directly to pointer events. The hard parts are not the sprite itself but the **environment it depends on and the inter-sprite physics**:

1. **No `Canvas` host exists.** ImageSprite's `X`/`Y` are coordinates *relative to the Canvas's top-left*, its edge events depend on the Canvas's pixel bounds, and it must render inside the Canvas's stacking context (`Z`-ordering against `Ball`s and other sprites). None of that exists today — there is no `Canvas` branch in `SimulationComponent.svelte`, no Canvas defaults, and no Canvas in the Go host. **`Canvas` must be implemented first** (a sized, positioned, optionally background-image/`BackgroundColor` `<div>` or `<canvas>` that owns a sprite coordinate system and reports its measured pixel size back to the host). This plan assumes Canvas lands first and documents the ImageSprite-specific work layered on top.

2. **Collision detection is real work, not a web-platform limitation.** `CollidedWith`, `NoLongerCollidingWith`, `CollidingWith()`, and `PointTowards()` require a per-frame pairwise collision engine over all sprites on the same Canvas (the spec even notes AI uses unrotated bounding boxes for rotated sprites). This is implementable in JS (axis-aligned bounding-box / circle-vs-box overlap, recomputed each animation frame), but it is genuinely new physics code that belongs in the Canvas layer, shared by `Ball`. It is feasible but raises effort.

3. **The motion loop is timer-driven and must run in the renderer.** AI moves the sprite "every `Interval` ms by `Speed` pixels along `Heading`". The synchronous Go host cannot run a wall-clock animation loop, so the motion integration, edge detection (`EdgeReached`), and collision checks must live frontend-side (JS `setInterval`/`rAF`), with the renderer writing back `X`/`Y`/`Heading` and firing events through `emitEvent`/`emitInteraction`. The Go host owns property reads/writes and method dispatch; it cannot own the tick.

There are no *device* limitations here (no sensors, camera, GPS, contacts) — everything is pure geometry the browser handles well. The feasibility caveat is purely architectural: **it is feasible only after Canvas, and the collision/edge/fling physics are substantial new code.** Without Canvas it is **not feasible** at all.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Picture` | `""` | Visual | The sprite's appearance. `resolveAssetUrl(assets, props.Picture)` -> `<img src>`. Empty -> faint placeholder box sized to `Width`/`Height` (matches AI showing nothing/empty until a picture is set). | High |
| `X` | `0.0` | Behavioral + Visual | Horizontal coord of the **origin** relative to Canvas left. Rendered as `left = X - OriginX*Width` (origin offset). Read-write at runtime; the motion loop writes it back. | High |
| `Y` | `0.0` | Behavioral + Visual | Vertical coord of origin relative to Canvas top. `top = Y - OriginY*Height`. | High |
| `Z` | `1.0` | Visual | Layering vs other sprites/balls. Map to CSS `z-index` (rounded/scaled); higher = front. | Medium |
| `Heading` | `0` | Behavioral + Visual | Degrees CCW from +x (0 = right, 90 = up). Drives motion vector `(cos, -sin)` (screen-y is down) and, if `Rotates`, the image `rotate(-Heading deg)`. | High |
| `Speed` | `0.0` | Behavioral | Pixels moved per `Interval` while `Enabled`. Drives the frontend motion loop. | High |
| `Interval` | `100` | Behavioral | ms between motion updates. The loop steps every `Interval` ms (or accumulates time in a `rAF` loop). | High |
| `Enabled` | `True` | Behavioral | Gates motion **and** interaction (touch/drag/fling/collisions). When false, sprite is frozen and ignores gestures. | High |
| `Rotates` | `True` | Visual | If true, image rotates with `Heading` about the origin (`transform-origin: OriginX*100% OriginY*100%`). If false, image stays upright. | Medium |
| `Visible` | `True` | Visual | Standard show/hide of the sprite node. Invisible sprites still exist for collisions in AI; for v1 hide visually but keep in motion/collision set. | High |
| `OriginX` | `0.0` | Behavioral | Unit X (0..1) of the origin within the sprite. Affects positioning and `transform-origin`. | Medium |
| `OriginY` | `0.0` | Behavioral | Unit Y (0..1) of the origin. Same. | Medium |
| `MarkOrigin` | `(0.0, 0.0)` | Designer-only (write-only block) | A designer marker for picking the origin; parsed into `OriginX`/`OriginY`. No runtime render of the marker. Parse `"(ox, oy)"` -> set `OriginX`/`OriginY` in `deriveStateFromDesignerProps`. | Low |
| `Width` / `Height` | (block-only) | Visual | Sprite pixel size. Not in designerProperties (sprite usually sizes to its picture); at runtime read-write. Auto -> natural image size; explicit -> fixed px. Needed for origin math, edge detection, and collisions. | High |

Notes: ImageSprite has **no font/text/`BackgroundColor`/`TextColor` properties** — do not add any (same discipline as `Image`/`VideoPlayer`). `X`/`Y`/`Heading`/`Speed`/`Width`/`Height`/`Z`/`OriginX`/`OriginY` are all numeric; `Enabled`/`Rotates`/`Visible` are boolean.

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `Touched` | `x, y` | Pointer down+up on the sprite without a drag, while `Enabled`. Fire `emitEvent(name, 'Touched', [x, y])` with coords relative to the **Canvas** top-left (convert from the pointer event minus Canvas bounding rect). |
| `TouchDown` | `x, y` | `pointerdown` on the sprite. `emitEvent(name, 'TouchDown', [x, y])`. |
| `TouchUp` | `x, y` | `pointerup` after a TouchDown. `emitEvent(name, 'TouchUp', [x, y])`. |
| `Dragged` | `startX, startY, prevX, prevY, currentX, currentY` | `pointermove` while pressed and `Enabled`. Track start coord, previous coord, current coord (all Canvas-relative). Fire each move. Per spec the sprite does **not** move unless the app calls `MoveTo` — so the simulator must NOT auto-move the sprite on drag; it only emits the event. (Most apps wire `MoveTo(currentX, currentY)` themselves.) |
| `Flung` | `x, y, speed, heading, xvel, yvel` | A quick `pointerdown`->`pointermove`->`pointerup` swipe over a short time. Compute velocity from displacement/elapsed-ms; `speed = hypot(xvel,yvel)`, `heading = atan2(-yvel, xvel)` in degrees (-180..180). Threshold on speed to distinguish from a Touched/Dragged. |
| `EdgeReached` | `edge` (number, N=1 NE=2 E=3 SE=4 S=-1 SW=-2 W=-3 NW=-4) | The frontend motion loop detects when the sprite's bounding box crosses a Canvas edge; classify which edge(s) and emit the AI direction code. Requires Canvas pixel bounds. |
| `CollidedWith` | `other` (component) | Per-frame collision engine detects the sprite's box newly overlapping another sprite/ball on the same Canvas. `other` is passed as the component reference. Requires the shared Canvas collision pass. |
| `NoLongerCollidingWith` | `other` (component) | Emitted the frame a previously-colliding pair stops overlapping. |

All events are firable in a browser — there are no sensor/permission gates. The only dependency is that `Canvas` must exist to provide the coordinate origin, edge bounds, and the sprite set for collisions. `CollidedWith`/`NoLongerCollidingWith`/`EdgeReached` additionally require the new physics loop; `Touched`/`TouchDown`/`TouchUp`/`Dragged`/`Flung` need only pointer wiring on the sprite node.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `MoveTo(x, y)` | `(number, number) -> void` | Set the sprite's origin to `(x, y)`: patch `X`/`Y` state (host `setProperty`), renderer repositions. The primary way drag actually moves a sprite. |
| `MoveToPoint(coords)` | `(list) -> void` | `coords = [x, y]`; same as `MoveTo(coords[0], coords[1])`. Parse the 2-element list in the Go host. |
| `MoveIntoBounds()` | `() -> void` | Clamp the sprite so it lies within Canvas bounds (left-align if too wide, top-align if too tall). Needs Canvas pixel size — compute frontend-side and patch `X`/`Y` back, OR have the host emit a `sprite-move-into-bounds` action the renderer resolves using measured Canvas size. Recommended: renderer-side (it knows the bounds). |
| `PointInDirection(x, y)` | `(number, number) -> void` | Set `Heading` so the sprite points at Canvas point `(x, y)`: `Heading = atan2(-(y - originY), (x - originX))` in degrees. Pure geometry; can be computed in the host from current `X`/`Y` or in the renderer. |
| `PointTowards(target)` | `(component) -> void` | Set `Heading` toward another sprite's origin: needs the target's `X`/`Y`. The host can read the target's stored `X`/`Y` and compute heading directly. |
| `Bounce(edge)` | `(number) -> void` | Reflect `Heading` off the given edge code (the value from `EdgeReached`): N/S edges negate the y-component, E/W edges negate the x-component. Pure geometry; update `Heading`. |
| `CollidingWith(other)` | `(component) -> boolean` | Return whether this sprite currently overlaps `other`. **Synchronous return** — the host must read the live collision state the renderer maintains (the renderer patches a per-pair collision set back into state, or the host recomputes from cached `X`/`Y`/`Width`/`Height` of both sprites). Recomputing in the host from cached geometry is the most reliable for a synchronous boolean. |

Every method is feasible (all geometry). The synchronous-return method `CollidingWith` and the bounds-aware `MoveIntoBounds` are the ones that need cached Canvas/sprite geometry in the host (or a renderer round-trip); the rest are simple state patches the renderer reacts to.

## Implementation Plan

> Prerequisite: implement `Canvas` first (supported type + defaults `{ Visible, Enabled, Width, Height, BackgroundColor, BackgroundImage, PaintColor, ... }`, a positioned render branch that establishes a sprite coordinate space and measures its pixel size, and Go handlers for Canvas methods/events). The steps below are the ImageSprite-specific additions on top of Canvas.

### simulation-capabilities.js

- Add `'Canvas'` and `'ImageSprite'` to `SIMULATION_SUPPORTED_TYPES`. Neither is non-visible — do **not** add to `SIMULATION_NONVISIBLE_TYPES`.
- Add a defaults block inside `buildSimulationDefaults()` (pulling defaults from `designerProperties`; add `Width`/`Height` as block-level sizing):

```js
ImageSprite: {
  Visible: true,
  Enabled: true,
  Picture: '',
  X: 0,
  Y: 0,
  Z: 1,
  Heading: 0,
  Speed: 0,
  Interval: 100,
  Rotates: true,
  OriginX: 0,
  OriginY: 0,
  Width: -1,   // auto -> natural image size
  Height: -1,
},
```

- Add new prop names to `SIMULATION_VISUAL_PROPS`: `X`, `Y`, `Z`, `Heading`, `Speed`, `Interval`, `Rotates`, `OriginX`, `OriginY`. (`Picture`, `Visible`, `Enabled`, `Width`, `Height` are already present.) Props not in this set are stripped before reaching the renderer, so the motion/positioning props must be added or the renderer never sees them.
- `isBooleanProp`: add `Rotates`. (`Visible`/`Enabled` already covered.)
- `isNumericProp`: add `X`, `Y`, `Z`, `Heading`, `Speed`, `Interval`, `OriginX`, `OriginY`. (`Width`/`Height` already covered.)
- `coerceSimulationValue`: numbers flow through `coerceNumber`, `Rotates` through `coerceBoolean`. Add a `MarkOrigin` -> `OriginX`/`OriginY` derivation in `deriveStateFromDesignerProps` (parse `"(ox, oy)"` and set `OriginX`/`OriginY`; `MarkOrigin` itself need not be a visual prop).

### SimulationComponent.svelte

The sprite must render **inside the Canvas branch's positioned box**, not as a free-floating component, because its coordinates are Canvas-relative and it competes for `z-index` with siblings. The Canvas branch iterates its children with `<svelte:self ... parentType="Canvas">` (like the arrangement branches). Add an `ImageSprite` branch that uses absolute positioning keyed off Canvas-relative `X`/`Y`:

```svelte
{:else if node.type === 'ImageSprite'}
  <div
    class="sim-sprite"
    class:sim-unsupported={unsupportedHere}
    class:sim-sprite-hidden={!boolValue(props.Visible, true)}
    style={spriteStyle()}
    data-sim-component={node.name}
    on:pointerdown={spritePointerDown}
    on:pointermove={spritePointerMove}
    on:pointerup={spritePointerUp}
    on:pointercancel={spritePointerCancel}
  >
    {#if spriteAssetUrl()}
      <img src={spriteAssetUrl()} alt="" draggable="false" />
    {/if}
  </div>
```

- `spriteStyle()`: `position: absolute;` + `left: ${X - OriginX*W}px; top: ${Y - OriginY*H}px;` + size from `Width`/`Height` (or natural image size when auto) + `z-index: ${Math.round(numberOr(Z,1)*10)}` + `transform: rotate(${boolValue(Rotates,true) ? -numberOr(Heading,0) : 0}deg); transform-origin: ${OriginX*100}% ${OriginY*100}%;`. Reuse `numberOr`/`boolValue`/`cssUrl`.
- `spriteAssetUrl()`: `resolveAssetUrl(assets, props.Picture)` (mirror the `assetName()`/`assetUrl` pattern; add a dedicated helper because `assetName()` already covers `Picture` but the sprite needs it independent of the generic `assetUrl` reactive).
- Pointer wiring (Canvas-relative coords): on `pointerdown`/move/up compute `x,y` from the event minus the **Canvas element's** `getBoundingClientRect()` (the Canvas branch should expose its bounds, e.g. via a context or by the sprite walking up to `[data-sim-component]` of the Canvas). Track press start + previous coords to drive `Touched`/`TouchDown`/`TouchUp`/`Dragged`/`Flung` exactly as the events table describes. Gate all on `enabled`. Mirror the existing `pointerDown`/`pointerUp`/`longClickTimer` pattern but with coordinate + velocity tracking.
- Motion loop: when `Enabled` and `Speed > 0`, run a per-sprite `setInterval(Interval)` (or a shared `rAF` accumulator) that advances `X += Speed*cos(Heading)`, `Y -= Speed*sin(Heading)`, emits `EdgeReached` on bound crossings, and patches `X`/`Y` back via `emitInteraction([{component, property:'X', value}, {component, property:'Y', value}])`. Tear it down in `onDestroy` and restart it reactively when `Enabled`/`Speed`/`Interval`/`Heading` change. Edge detection and collisions need the Canvas pixel bounds + the sibling sprite set, which is why this logic is best centralized in the Canvas layer (a Canvas-level animation tick iterating its sprite children) rather than per-sprite — recommended refactor: the **Canvas branch owns the tick and the collision pass**, ImageSprite exposes its geometry. v1 can ship per-sprite motion + edge events and defer collisions.
- CSS: `.sim-sprite { position:absolute; pointer-events:auto; user-select:none; touch-action:none; } .sim-sprite img { display:block; width:100%; height:100%; pointer-events:none; } .sim-sprite-hidden { visibility:hidden; }`.

### SimulateOverlay.svelte

**None required for v1.** No dialogs/toasts. Method calls (`MoveTo`, `Bounce`, etc.) are state patches or `component-action` effects the existing `applyEffects()` -> `handleComponentActions` plumbing already routes. If `MoveIntoBounds`/`CollidingWith` are implemented as renderer round-trips, they reuse the existing `interaction.properties` patch path (no new overlay handler). Only revisit the overlay if Canvas needs a new effect kind.

### simulation_wasm.go

A dedicated `ImageSprite` path is needed for the motion methods and the geometry-returning method; plain property reads/writes use the generic store path.

- **`GetProperty`:** generic store path suffices for `X`/`Y`/`Z`/`Heading`/`Speed`/`Interval`/`Enabled`/`Rotates`/`Picture`/`OriginX`/`OriginY`/`Width`/`Height` (all stored, kept current by renderer patches during motion). No computed case strictly required.
- **`SetProperty`:** generic `h.setProperty` suffices — the renderer reacts to `X`/`Y`/`Heading`/`Speed`/`Interval`/`Enabled` changes. (Optionally clamp `OriginX`/`OriginY` to `[0,1]` like the Slider clamp pattern.)
- **`CallMethod`:** add `case "ImageSprite": return h.callImageSpriteMethod(...)`:
  - `MoveTo` -> `setProperty(name,"X",args[0])`, `setProperty(name,"Y",args[1])`.
  - `MoveToPoint` -> parse the 2-element list arg, then same as `MoveTo`.
  - `PointInDirection` -> read current `X`/`Y`, compute `Heading = atan2(-(y-Y),(x-X))` deg, `setProperty(name,"Heading",...)`.
  - `PointTowards` -> read target sprite's `X`/`Y` (target component name from the arg), compute heading, set it.
  - `Bounce` -> read `Heading`, reflect per edge code, set `Heading`.
  - `MoveIntoBounds` -> emit `componentAction(name, "sprite-move-into-bounds")` (renderer clamps using measured Canvas bounds and patches `X`/`Y` back).
  - `CollidingWith` -> return `runtime.BoolVal(...)` computed from cached `X`/`Y`/`Width`/`Height` of this sprite and the target (AABB overlap), or read a renderer-maintained collision set; default to `false` if geometry unknown.
  - default -> `h.Unsupported("method", name+"."+method)`.
- **Events:** `Touched`/`TouchDown`/`TouchUp`/`Dragged`/`Flung`/`EdgeReached`/`CollidedWith`/`NoLongerCollidingWith` originate **frontend-side** via `emitEvent` (pointer + motion loop), so no host `runEvent` call is needed to originate them. Add `ImageSprite` (and `Canvas`) entries to `EVENT_ARGS` in `simulation-capabilities.js` so the arg names match the spec (e.g. `Touched: ['x','y']`, `Dragged: ['startX','startY','prevX','prevY','currentX','currentY']`, `Flung: ['x','y','speed','heading','xvel','yvel']`, `EdgeReached: ['edge']`, `CollidedWith: ['other']`, `NoLongerCollidingWith: ['other']`).
- Recompile with `npm run build:wasm` (`GOOS=js GOARCH=wasm`).

### design-schema-tree.js

**Containment already handled.** `CANVAS_CHILD_TYPES = { Ball, ImageSprite }` and `canContainDesignComponent` already enforce `ImageSprite` only inside `Canvas` (and `Canvas` accepts no other children). `designTreeToInitialState()` will merge `SIMULATION_DEFAULTS.ImageSprite` automatically. `unsupportedSimulationComponents()` stops flagging ImageSprite (and Canvas) once both are added to `SIMULATION_SUPPORTED_TYPES`. No change to this file is required for ImageSprite itself.

## Dependencies & Ordering

- **No external libraries.** Pure DOM + CSS transforms + JS timers/`rAF`. No Leaflet/MapLibre-class dependency.
- **Hard prerequisite: `Canvas` must be implemented first.** ImageSprite cannot render, position, detect edges, or collide without a Canvas providing the coordinate origin, pixel bounds, stacking context, and the sprite set. Implement Canvas (render branch + defaults + measured bounds + Go handlers + the shared sprite tick/collision pass) before ImageSprite.
- **Shared with `Ball`.** The motion loop, edge detection, and collision engine should be built at the Canvas layer so `Ball` reuses them. Implementing ImageSprite and `Ball` together (sharing the Canvas physics) is more efficient than doing either alone. `CollidedWith`/`PointTowards`/`CollidingWith` are cross-sprite and need both types present to be meaningful.

## Web-Platform Limitations & Fidelity Caveats

- **Requires Canvas; not feasible standalone.** Until Canvas exists, ImageSprite stays a dashed placeholder. This is an internal architecture gap, not a browser limitation.
- **Collision model is approximate.** AI checks collisions using the sprite's **unrotated** bounding box (the spec explicitly warns this is inaccurate for rotated tall/narrow or short/wide sprites). The simulator should replicate the same unrotated-AABB approximation for parity — meaning collisions will be visually "wrong" for rotated sprites in exactly the way AI is. Pixel-perfect/alpha-mask collision is out of scope.
- **Timing fidelity.** `Interval` is honored by a JS timer; browser timer throttling (background tabs, `setInterval` clamping to ~4ms minimum, and rAF pausing when the tab is hidden) means motion may stutter or pause when the simulator tab is not focused. A real device keeps ticking. Sub-frame `Interval` values (<16ms) cannot run faster than the display refresh.
- **`CollidingWith` synchronicity.** The Go `CallMethod` returns synchronously, but the authoritative geometry lives in the DOM/renderer. Returning a boolean requires the host to cache each sprite's `X`/`Y`/`Width`/`Height` (kept current by motion patches) and recompute overlap; immediately after a `MoveTo` in the same block tick, the cached geometry may lag one frame behind the rendered position.
- **`Dragged` does not move the sprite.** Per spec, dragging only fires the event; the app must call `MoveTo`. Apps that forget this will see the event but no movement — faithful to AI, but can look "broken" to someone expecting drag-to-move.
- **Fling heuristics differ.** `speed`/`heading`/velocity from a pointer swipe are derived from displacement/elapsed-ms; the exact thresholds distinguishing a fling from a drag/touch will not match Android's `GestureDetector` numbers, so borderline gestures may classify differently than on a device.
- **`Z`-ordering granularity.** AI `Z` is a float; CSS `z-index` is an integer. Very close `Z` values may collapse to the same layer after rounding.
- **Invisible-but-active sprites.** In AI an invisible sprite can still collide; v1 may simplify by excluding hidden sprites from interaction — note this divergence if so.

## Effort Estimate

**L** (XL if Canvas + Ball + full collision engine are counted as one unit). ImageSprite *in isolation* atop an existing Canvas is roughly **M** (render branch with absolute positioning + rotation, pointer/drag/fling wiring, a per-sprite motion loop with edge events, and a `callImageSpriteMethod` with seven geometry methods plus one `build:wasm`). But it is **gated on building `Canvas` first** (coordinate space, measured bounds, render branch, Go handlers) and the high-value features (`CollidedWith`/`NoLongerCollidingWith`/`CollidingWith`/`PointTowards`/`EdgeReached`) require a **shared Canvas-level motion + collision engine** also serving `Ball` — that physics layer is the bulk of the work and pushes the realistic end-to-end effort to **L**. A pragmatic v1 (render + position + touch/drag/fling + autonomous motion + edge events, deferring collisions) is **M** once Canvas exists.
