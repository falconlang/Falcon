# Ball Simulator Implementation Plan

## Overview
`Ball` is an App Inventor ANIMATION sprite: "A round 'sprite' that can be placed on a `Canvas`, where
it can react to touches and drags, interact with other sprites (`ImageSprite`s and other `Ball`s) and
the edge of the Canvas, and move according to its property values." Unlike `ImageSprite`, a `Ball`
has no image asset — its entire appearance is a filled circle driven by `PaintColor` and `Radius`.
It can move autonomously: with `Enabled = True`, `Speed > 0`, and a heading, App Inventor's internal
animation clock moves the ball `Speed` pixels along `Heading` every `Interval` milliseconds, firing
`EdgeReached` at canvas borders and `CollidedWith` / `NoLongerCollidingWith` against other sprites.

- **Category:** ANIMATION
- **Visible:** Yes (`nonVisible: false`), but it is **not** a top-level visible component — it is a
  drawing primitive that exists only on a `Canvas` surface.
- **Container relationship:** `Ball` is a **child of `Canvas`** (Drawing & Animation). It is never a
  parent. `canContainDesignComponent` already encodes `CANVAS_CHILD_TYPES = {Ball, ImageSprite}` ->
  only `Canvas` may contain it (see `design-schema-tree.js`). It therefore **hard-depends on a Canvas
  implementation**: the ball is positioned (`X`/`Y`) relative to the Canvas coordinate space, sized in
  Canvas pixels, and its touch/fling/edge semantics are all defined relative to the Canvas rectangle.

## Feasibility Verdict
**Partially feasible (in a browser simulator) — and blocked on Canvas.**

The *static rendering* of a Ball is trivially feasible: a CSS circle (`border-radius: 50%`,
`background`, `width/height = 2*Radius`) absolutely positioned at `(X, Y)` inside a positioned Canvas
container. Touch/drag/fling event wiring is feasible with DOM pointer events. The hard parts are the
two things that make a sprite a *sprite*, and both run into a concrete architectural gap:

1. **Autonomous movement (the animation clock).** App Inventor moves the ball every `Interval` ms
   when `Enabled && Speed > 0`. The Tensor simulator **has no timer/animation loop**: grepping
   `SimulateOverlay.svelte` and `simulation_wasm.go` shows no `setInterval`, no
   `requestAnimationFrame`, and no Clock-style tick. The whole simulator is **event-driven** — the Go
   host runs blocks only in response to a user interaction (`runEvent`) or applies one-shot `effects`;
   it does not "advance time." So a ball with `Speed = 4, Interval = 500` will **sit still** in the
   simulator. This is the same class of limitation that makes the `Clock` component's `Timer` event
   un-fireable. Honest stance: autonomous motion is **not feasible** without building a new ticking
   subsystem; the realistic approximation is to render the ball at its *designer/runtime* `X`/`Y` and
   let blocks reposition it imperatively via `MoveTo` / setting `X`/`Y` (which we *can* reflect live).

2. **Collision detection.** `CollidedWith` / `NoLongerCollidingWith` / `CollidingWith` require a
   per-frame geometric overlap test between every enabled sprite pair on the Canvas. With no tick loop
   and (initially) no other sprites implemented, there is nothing to drive or test collisions. These
   are **not feasible** in the first cut and should be marked `Unsupported` at the host.

**Realistic approximation (what we will simulate):**
- Render the ball as a positioned circle inside the Canvas, honoring `PaintColor`, `Radius`, `X`,
  `Y`, `Z` (z-index), `Visible`.
- Wire `TouchDown` / `TouchUp` / `Touched` / `Dragged` / `Flung` from real DOM pointer events on the
  circle, computing canvas-relative `(x, y)` and the fling vector. These are genuine user
  interactions and map cleanly to the event-driven model.
- Support imperative repositioning: `MoveTo`, `MoveToPoint`, `PointInDirection` and writes to
  `X`/`Y`/`Heading` update stored state and re-render live (generic store path + a small Ball method
  handler).
- Mark autonomous motion, `Bounce`/`EdgeReached`, and all collision surface as `Unsupported` (orange
  warning) so the limitation is visible rather than silently wrong.

This is a credible, useful approximation for the large class of apps that drag a ball or move it on
button-press, while being honest that "set Speed and watch it fly" will not animate.

## Properties
| Property | AI default | Visual/Behavioral | How to simulate | Priority |
| --- | --- | --- | --- | --- |
| `PaintColor` | `&HFF000000` | Visual — fill color of the circle | `background: colorValue(props.PaintColor, '#000')` | High |
| `Radius` | `5` | Visual — circle radius in px | `width/height = 2*Radius`; clamp >= 0 | High |
| `X` | `0.0` | Visual/Behavioral — left (or center if `OriginAtCenter`) in Canvas px | `position:absolute; left:` derived from `X` (and `Radius` if `OriginAtCenter`) | High |
| `Y` | `0.0` | Visual/Behavioral — top (or center) in Canvas px | `position:absolute; top:` derived from `Y` | High |
| `Z` | `1.0` | Visual — layer order vs other sprites | `z-index: round(Z)` | Medium |
| `Visible` | `True` | Visual | Generic `Visible` handling (already wired) | High |
| `Enabled` | `True` | Behavioral — gates motion + interaction | Store it; gates whether touch/drag/fling events fire. Motion itself not simulated | Medium |
| `Heading` | `0` | Behavioral — degrees CCW from +x axis | Stored; consumed by `PointInDirection`/`Flung` math; no visual unless motion existed | Medium |
| `OriginAtCenter` | `False` | Behavioral — whether `X`/`Y` mean center vs top-left | Affects the left/top derivation: if true, `left = X - Radius` | Medium |
| `Speed` | `0.0` | Behavioral — px per interval | Stored; **cannot animate** (no tick loop) — caveat | Low |
| `Interval` | `100` | Behavioral — ms between moves | Stored; **no tick loop** to honor it — caveat | Low |

Notes:
- All defaults are pulled from the spec's `designerProperties`. `PaintColor` is an AI ARGB hex
  (`&HFF000000`) — `colorValue()` already parses that format.
- `OriginAtCenter` is `rw: invisible` (not in the designer property editor as a runtime-visible prop,
  but it has a default of `False`); it must still be honored when converting `X`/`Y` to CSS left/top.
- `Speed`/`Interval`/`Heading` are read-write and must round-trip through the store, but only `Heading`
  has any simulated *consumer* (the pointing methods); `Speed`/`Interval` are inert without a clock.

## Events
| Event | Args | How/when the simulator fires it |
| --- | --- | --- |
| `Touched` | `x, y` | Pointer tap on the ball (down+up with no drag); fire with canvas-relative coords. Only if `Enabled`. Feasible. |
| `TouchDown` | `x, y` | `pointerdown` on the ball; canvas-relative coords. Feasible. |
| `TouchUp` | `x, y` | `pointerup` on the ball; canvas-relative coords. Feasible. |
| `Dragged` | `startX, startY, prevX, prevY, currentX, currentY` | `pointermove` while pressed: track start point, previous point, current point in canvas coords and emit per move. Feasible (matches AI's "MoveTo not called automatically" semantics — we only fire the event; blocks decide whether to move). |
| `Flung` | `x, y, speed, heading, xvel, yvel` | On a quick `pointerup` after movement: compute velocity from the last pointer delta / elapsed time, `speed = hypot(xvel,yvel)`, `heading = atan2(-yvel, xvel)` in degrees. Feasible (approximation of native fling thresholds). |
| `EdgeReached` | `edge` (Direction 1..-4) | Fires when the moving ball hits a Canvas border. **Cannot fire** — depends on autonomous motion (no tick loop). Mark conceptually unsupported. |
| `CollidedWith` | `other` (component) | Fires when two enabled sprites overlap. **Cannot fire** — needs a per-frame collision test + other sprites + a tick loop. Unsupported. |
| `NoLongerCollidingWith` | `other` (component) | Inverse of collision. **Cannot fire** — same reason. Unsupported. |

Event arg ordering must be registered in `EVENT_ARGS` (see Implementation Plan) so the runtime
receives positional args correctly, mirroring the existing `Canvas` entry.

## Methods
| Method | Signature | Simulated behavior (or Unsupported + why) |
| --- | --- | --- |
| `MoveTo` | `MoveTo(x, y)` -> void | Feasible. Set stored `X = x`, `Y = y` (honor `OriginAtCenter`), patch state, re-render at new position. |
| `MoveToPoint` | `MoveToPoint(coordinates: list)` -> void | Feasible. Parse `[x, y]` from the list arg, then same as `MoveTo`. |
| `PointInDirection` | `PointInDirection(x, y)` -> void | Feasible. Compute `Heading = atan2(yOrigin - y, x - xOrigin)` in degrees (AI's y-down convention) and store it. No visual change (ball has no orientation), but `Heading` reads back correctly. |
| `MoveIntoBounds` | `MoveIntoBounds()` -> void | Partially feasible **iff Canvas exposes its pixel size**. Clamp `X`/`Y` so the circle stays within the Canvas rect; if the ball is larger than the canvas, align its left/top to the canvas left/top (per spec). Requires Canvas width/height; otherwise `Unsupported`. |
| `Bounce` | `Bounce(edge)` -> void | **Unsupported.** Bouncing reflects the heading/velocity off an edge as part of the motion model; with no autonomous motion and no `EdgeReached` source, there is nothing to bounce. Mark `h.Unsupported("method", "Ball.Bounce")`. |
| `CollidingWith` | `CollidingWith(other)` -> boolean | **Unsupported** (return `false`). No collision system; returning a constant `false` is the least-surprising stub, plus an `Unsupported` notice. |
| `PointTowards` | `PointTowards(target: component)` -> void | Partially feasible **iff the target sprite is implemented and its `X`/`Y` are readable**. Compute heading from this ball's origin to the target's origin. Until `ImageSprite`/multi-`Ball` reads are wired, treat as `Unsupported` or best-effort if both are `Ball`s with known positions. |

## Implementation Plan

### simulation-capabilities.js
- **`SIMULATION_SUPPORTED_TYPES`:** add `'Ball'`. (Do **not** add to `SIMULATION_NONVISIBLE_TYPES` — it
  is visible.)
- **`buildSimulationDefaults()`** — add a defaults block (Ball does not use `COMMON_VISIBLE_PROPS`
  because it has no `Width`/`Height` designer props; it uses `Radius`):
  ```js
  Ball: {
    Visible: true,
    Enabled: true,
    PaintColor: '&HFF000000',
    Radius: 5,
    X: 0,
    Y: 0,
    Z: 1,
    Heading: 0,
    Speed: 0,
    Interval: 100,
    OriginAtCenter: false,
  },
  ```
- **`SIMULATION_VISUAL_PROPS`:** add the new names `'PaintColor'`, `'Radius'`, `'X'`, `'Y'`, `'Z'`,
  `'Heading'`, `'Speed'`, `'Interval'`, `'OriginAtCenter'`. (`Visible`/`Enabled` already present.)
  Without these the props would be **stripped** before reaching the renderer.
- **`isBooleanProp`:** add `'OriginAtCenter'`.
- **`isNumericProp`:** add `'Radius'`, `'X'`, `'Y'`, `'Z'`, `'Heading'`, `'Speed'`, `'Interval'`.
  (`PaintColor` stays a string/ARGB and is handled by `colorValue` in the renderer — do **not** make
  it numeric.)
- **`coerceSimulationValue`:** no new special case needed — `isNumericProp` -> `coerceNumber`, and
  `isBooleanProp` -> `coerceBoolean` cover the additions. `PaintColor` falls through as a string.
- **`deriveStateFromDesignerProps`:** no derived state required.

### SimulationComponent.svelte
This branch is unusual: the ball must be **absolutely positioned inside the Canvas surface**, so it
relies on the Canvas branch rendering a `position: relative` container (the Canvas plan must provide
that). Sketch the branch (place it near the other leaf components):
```svelte
{:else if node.type === 'Ball'}
  <button
    type="button"
    class="sim-ball"
    class:sim-unsupported={unsupportedHere}
    style={ballStyle()}
    disabled={!enabled}
    data-sim-component={node.name}
    on:pointerdown={ballPointerDown}
    on:pointermove={ballPointerMove}
    on:pointerup={ballPointerUp}
    on:pointercancel={ballPointerCancel}
  ></button>
{/if}
```
- **`ballStyle()`** (new helper): compute `r = max(0, numberOr(props.Radius, 5))`, the diameter
  `2*r`, and left/top from `X`/`Y` with `OriginAtCenter`:
  ```js
  function ballStyle() {
    const r = Math.max(0, numberOr(props.Radius, 5));
    const center = boolValue(props.OriginAtCenter, false);
    const left = numberOr(props.X, 0) - (center ? r : 0);
    const top  = numberOr(props.Y, 0) - (center ? r : 0);
    return [
      'position: absolute;',
      `left: ${left}px; top: ${top}px;`,
      `width: ${2*r}px; height: ${2*r}px;`,
      'border-radius: 50%; padding: 0; border: 0; cursor: pointer;',
      `background: ${colorValue(props.PaintColor, '#000')};`,
      `z-index: ${Math.round(numberOr(props.Z, 1))};`,
    ].join(' ');
  }
  ```
- **Pointer wiring** (new functions, modeled loosely on the existing `pointerDown`/`pointerUp` and the
  Canvas `EVENT_ARGS` coords convention): on `pointerdown` capture the pointer, record `start`/`prev`
  in **canvas-relative** coords (use `e.currentTarget.parentElement.getBoundingClientRect()` or a
  Canvas-provided rect), set a pressed flag, fire `TouchDown`. On `pointermove` while pressed, fire
  `Dragged(startX, startY, prevX, prevY, curX, curY)` and advance `prev`. On `pointerup` fire
  `TouchUp`; if total movement is below a small threshold fire `Touched`, else compute the fling
  vector from the last delta and elapsed time and fire `Flung`. All gated on `enabled`. Use
  `emitEvent(node.name, 'Dragged', [startX, startY, prevX, prevY, curX, curY])` etc.
- **Reuse:** `colorValue`, `numberOr`, `boolValue` helpers already exist; the long-press/`pointerDown`
  pattern from Button is the closest existing reference for pointer event dispatch. Do **not** use
  `baseStyle()` (it injects size/background-image/alignment rules that do not apply to a sprite).
- **CSS:** add a minimal `.sim-ball { position: absolute; }` rule; the inline style drives geometry.
- **Static-only fallback:** if Canvas is not yet implemented, the Ball still renders inside the
  unsupported Canvas placeholder but without a positioned parent — so this branch is only meaningful
  **after** the Canvas `position: relative` surface exists.

### SimulateOverlay.svelte
None required for the feasible subset (no dialogs/toasts). The `Unsupported` notices for
`Bounce`/collision/motion surface through the existing orange-warning path automatically. If a future
iteration adds a tick loop, that subsystem would likely live here (it hosts the session) — out of
scope for this plan.

### simulation_wasm.go
Add a `case "Ball":` to `CallMethod` dispatching to a new `callBallMethod`, plus targeted
`GetProperty`/`SetProperty` handling. Most properties ride the **generic store path** already.
- **`GetProperty`:** generic store read returns `PaintColor`/`Radius`/`X`/`Y`/`Z`/`Heading`/`Speed`/
  `Interval`/`Enabled`/`Visible`/`OriginAtCenter`. No computed-property case needed.
- **`SetProperty`:** generic `h.setProperty` suffices for all read-write props (`X`/`Y`/`Heading`/
  etc.) — each patches state and re-renders the circle live. No special derivation needed.
- **`CallMethod` -> `callBallMethod`:**
  - `MoveTo(x, y)`: `h.setProperty(name, "X", args[0]); h.setProperty(name, "Y", args[1])`.
  - `MoveToPoint(coords)`: read list elem 0/1, then same as `MoveTo`.
  - `PointInDirection(x, y)`: compute heading via `math.Atan2(originY - y, x - originX) * 180/π`
    (reading current `X`/`Y` and `Radius`/`OriginAtCenter` for the origin), store `Heading`.
  - `MoveIntoBounds()`: clamp against Canvas size **if the Canvas host exposes width/height**;
    otherwise `h.Unsupported("method", "Ball.MoveIntoBounds")`.
  - `PointTowards(target)`: if target position is readable, compute heading to it; else
    `h.Unsupported(...)`.
  - `Bounce(edge)`: `h.Unsupported("method", "Ball.Bounce")` — no motion model.
  - `CollidingWith(other)`: `h.Unsupported("method", "Ball.CollidingWith")`, return
    `runtime.BoolVal(false)`.
  - `default`: `h.Unsupported("method", componentName+"."+method)`.
- **Events:** `Touched`/`TouchDown`/`TouchUp`/`Dragged`/`Flung` are dispatched from the renderer via
  `emitEvent` -> `runEvent`; no host-side firing logic needed. `EdgeReached`/`CollidedWith`/
  `NoLongerCollidingWith` have no driver and are simply never fired (optionally note via `Unsupported`
  the first time motion-dependent blocks are reached, though there is no clean hook — leaving them
  unfired is acceptable).

### design-schema-tree.js
**Already handled.** `CANVAS_CHILD_TYPES` includes `'Ball'`, and `canContainDesignComponent` returns
`parent === 'Canvas'` for it — so containment validation is correct with no edit.
`designTreeToInitialState` will merge `SIMULATION_DEFAULTS.Ball` with derived designer props once the
defaults block exists. `unsupportedSimulationComponents` stops flagging `Ball` once it is added to
`SIMULATION_SUPPORTED_TYPES`. No code change in this file.

## Dependencies & Ordering
- **External libraries:** none. A CSS circle + DOM pointer events suffice; no `<canvas>` element is
  strictly required for the Ball itself (the Canvas surface can be a positioned `<div>`).
- **Hard prerequisite: `Canvas` must be implemented first.** The Ball is meaningless without a
  positioned Canvas surface that (a) renders children with `position: relative`, (b) defines the
  coordinate origin and pixel size, and (c) ideally exposes its bounding rect to children for
  canvas-relative coordinate math and `MoveIntoBounds`. Implement Canvas, then Ball.
- **Soft prerequisite: a tick/animation subsystem** for autonomous motion + collisions. Not required
  for the feasible subset; required to lift `Speed`/`Interval`/`Bounce`/`EdgeReached`/collisions out
  of `Unsupported`. Treat as a separate, larger initiative.
- **Related sibling: `ImageSprite`** shares nearly all of this surface (minus `PaintColor`/`Radius`,
  plus an image asset). Implementing Ball establishes the sprite pattern ImageSprite will reuse, and
  full `CollidedWith`/`PointTowards` fidelity needs both sprites present.

## Web-Platform Limitations & Fidelity Caveats
- **No autonomous motion (primary caveat):** the simulator has no Clock/tick loop, so `Speed`,
  `Interval`, and `Heading`-driven movement do not animate. A ball configured to "drift" sits still.
  Blocks can still reposition it imperatively (`MoveTo`, set `X`/`Y`), which renders live.
- **No edge detection:** `EdgeReached` never fires and `Bounce` is a no-op/`Unsupported`, both because
  they are byproducts of the missing motion model.
- **No collision system:** `CollidedWith`, `NoLongerCollidingWith`, and `CollidingWith` cannot be
  computed — there is no per-frame overlap test, and in the first cut no other sprite types are even
  present. `CollidingWith` returns a constant `false`.
- **Coordinate-space coupling to Canvas:** `X`/`Y` and all event coords are Canvas-relative; accuracy
  depends entirely on the Canvas surface reporting a correct origin/size. `OriginAtCenter` must be
  honored consistently in both the renderer (left/top) and the host methods (heading origin).
- **Fling fidelity:** native fling uses platform velocity thresholds; the simulated `speed`/`heading`/
  `xvel`/`yvel` are derived from the last pointer delta over elapsed time and are an approximation,
  not pixel/ms-exact device parity.
- **Z-layering vs other sprites/Canvas drawing:** `Z` maps to CSS `z-index` among sprites, but
  interleaving with Canvas freehand drawing (`DrawLine`, `DrawCircle`, etc.) depends on how the Canvas
  layer is rendered and may not match device stacking exactly.

## Effort Estimate
**M** — the static circle, prop plumbing (capabilities entries, defaults, visual/numeric/boolean
sets), and the imperative `MoveTo`/`MoveToPoint`/`PointInDirection` host methods are straightforward
(S-sized in isolation). The cost driver is the **pointer-event sprite interaction** (canvas-relative
coordinate math, drag tracking with start/prev/current, fling vector computation) and the careful
`Unsupported` marking of the motion/collision surface — plus the hard dependency that **Canvas must
land first**. Lifting autonomous motion/collisions out of `Unsupported` would be a separate **L/XL**
effort (a new tick/animation subsystem) and is explicitly out of scope here.
