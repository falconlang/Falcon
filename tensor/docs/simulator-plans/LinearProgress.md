# LinearProgress Simulator Implementation Plan

## Overview
LinearProgress is a visible User Interface component (categoryString `USERINTERFACE`, `nonVisible: false`). Per the spec helpString it "indicates the progress of an operation using an animated linear bar." It has two modes:

- **Indeterminate** (`Indeterminate = True`, the default): an infinitely animating bar with no fixed fill value, used to indicate "working" without a known percentage. In this mode `Progress` is ignored.
- **Determinate** (`Indeterminate = False`): a filled bar whose fill is `Progress` clamped to `[Minimum, Maximum]`.

Container relationship: standalone visible component. It is not a container and has no children. It lives directly under a Screen/Form or inside any arrangement, exactly like Slider/Label/Button. No special parent/child rules apply.

Closest existing analog in the simulator: **Slider** (numeric range + clamping host + CSS-var color theming). LinearProgress is strictly simpler — it is read-only/display (the user does not drag it) and maps almost 1:1 onto the native HTML `<progress>` element, which is explicitly listed as available in this environment.

## Feasibility Verdict
**Feasible.** Both modes render natively in the browser:

- Determinate: `<progress max={...} value={...}>` gives a real filled bar. `Minimum`/`Maximum`/`Progress` map to a normalized `value`/`max` (offset by `Minimum`).
- Indeterminate: `<progress>` with no `value` attribute already renders the browser's built-in indeterminate animation; for full control over color/animation we render a custom `<div>` track with a CSS keyframe sweep (so `IndeterminateColor` is honored). Both `<progress>`, CSS, and JS animation are available here.
- `ProgressColor`, `IndeterminateColor`, `Minimum`, `Maximum`, `Progress`, `Visible`, `Width`/`WidthPercent` all have direct browser equivalents.
- `ProgressChanged` is a plain state-derived event (no sensor/permission), and `IncrementProgressBy` is pure arithmetic on host state.

One real-device divergence worth noting (not a blocker): the spec says `Minimum` "works only for devices with API >= 26." The simulator will always honor `Minimum` (it has no API-level gate), so a determinate bar with a non-zero `Minimum` may look slightly more correct in the simulator than on an old Android device. This is a fidelity caveat, not a limitation.

## Properties
| Property | AI default | Visual/Behavioral | How to simulate | Priority |
| --- | --- | --- | --- | --- |
| Indeterminate | `True` (boolean) | Behavioral + Visual | Boolean; selects render branch (CSS sweep vs. filled `<progress>`). Default true means the bar animates out of the box. | High |
| IndeterminateColor | `&HFF0000FF` (blue) | Visual | `colorValue()` -> CSS custom prop `--lp-indeterminate`; colors the animated sweep. | High |
| ProgressColor | `&HFF0000FF` (blue) | Visual | `colorValue()` -> CSS custom prop `--lp-progress`; colors the `<progress>` fill (`::-webkit-progress-value` / `::-moz-progress-bar`). | High |
| Maximum | `100` (integer) | Behavioral | Numeric; upper bound. Used to normalize fill = `(Progress - Minimum) / (Maximum - Minimum)`. | High |
| Minimum | `0` (integer) | Behavioral | Numeric; lower bound. Clamp + normalize. (No API-26 gate in simulator.) | High |
| Progress | (not a designer prop; blockProperty, default effectively `Minimum`/0) | Behavioral + Visual | Numeric runtime value, clamped to `[Minimum, Maximum]`. Drives determinate fill. Add to defaults as `0`. | High |
| Visible | `True` (visibility) | Visual | Standard `visible` handling already wired via COMMON props. | High |
| Width | `-1` (automatic) | Visual | Standard `sizeStyle('Width')`. For a progress bar, default automatic; AI bars are typically full-width. | Medium |
| WidthPercent | (write-only blockProperty) | Visual | Optional runtime write; convert percent of screen width to px. Low priority; can defer. | Low |
| Left / Top | `""` (integer) | Visual | Only meaningful inside AbsoluteArrangement; already handled by `positionStyle()`. | Low |
| Height | `-1` | Visual | Standard; a thin bar, height rarely set. | Low |
| Column / Row | invisible blockProperties | n/a | Not surfaced; ignore. | n/a |

Notes: `Progress` is the only runtime-meaningful value that is NOT a designer property, so it must be seeded in `buildSimulationDefaults` (default `0`) so the host has a starting value for `IncrementProgressBy` and for the determinate fill.

## Events
| Event | Args | How/when the simulator fires it |
| --- | --- | --- |
| ProgressChanged | `progress` (number) | Fired by the host whenever `Progress` changes (via `SetProperty("Progress", ...)` or `IncrementProgressBy`). Per the spec, if `Indeterminate` is true the event reports `0`. The component is display-only, so there is no user-driven DOM event — the event originates entirely from host-side property writes (mirrors how TextBox `TextChanged` fires from `SetProperty`). No browser limitation. |

There are no user-interaction events (no Click/Touch) on LinearProgress — it does not respond to taps. So no `on:click`/`on:pointer*` wiring is needed in the render branch.

## Methods
| Method | Signature | Simulated behavior |
| --- | --- | --- |
| IncrementProgressBy | `IncrementProgressBy(value: number)` -> void | Host reads current `Progress`, adds `value`, clamps to `[Minimum, Maximum]`, writes it back via the same path that fires `ProgressChanged`. In indeterminate mode App Inventor still tracks the underlying progress value but the bar ignores it; the simulator should update the stored `Progress` and (matching the `ProgressChanged` spec) report `0` while indeterminate. Pure arithmetic, fully simulable. |

## Implementation Plan

### simulation-capabilities.js
- Add `'LinearProgress'` to `SIMULATION_SUPPORTED_TYPES`. (NOT added to `SIMULATION_NONVISIBLE_TYPES` — it is visible.)
- Add a defaults block in `buildSimulationDefaults()`:
  ```js
  LinearProgress: {
    Visible: true,
    Width: -1,
    Height: -1,
    Indeterminate: true,
    IndeterminateColor: '&HFF0000FF',
    ProgressColor: '&HFF0000FF',
    Minimum: 0,
    Maximum: 100,
    Progress: 0,
  },
  ```
  (No `Enabled`/font/text props — LinearProgress has none.)
- Add new prop names to `SIMULATION_VISUAL_PROPS`: `'Indeterminate'`, `'IndeterminateColor'`, `'ProgressColor'`, `'Minimum'`, `'Maximum'`, `'Progress'`. (`Visible`/`Width`/`Height`/`Left`/`Top` are already in the set.)
- `isBooleanProp`: add `'Indeterminate'`.
- `isNumericProp`: add `'Minimum'`, `'Maximum'`, `'Progress'`. (`Width`/`Height`/`Left`/`Top` already present.)
- `coerceSimulationValue`: no special-case needed — booleans/numbers are handled by the `isBooleanProp`/`isNumericProp` branches once the names are registered. Color strings (`IndeterminateColor`/`ProgressColor`) fall through as raw strings, matching Slider's `ColorLeft`/`ColorRight`.
- `deriveStateFromDesignerProps`: no derived state required (no Elements/Instant-style derivation). The generic loop suffices.

### SimulationComponent.svelte
Add helpers near the Slider helpers (reuse `numberOr`, `boolValue`, `colorValue`, `baseStyle`):
  ```js
  function lpFraction() {
    const min = numberOr(props.Minimum, 0);
    const max = numberOr(props.Maximum, 100);
    const span = max - min;
    if (span <= 0) return 0;
    const value = Math.max(min, Math.min(max, numberOr(props.Progress, min)));
    return (value - min) / span;
  }
  function lpStyle() {
    return baseStyle([
      `--lp-progress: ${colorValue(props.ProgressColor, '#0000ff')};`,
      `--lp-indeterminate: ${colorValue(props.IndeterminateColor, '#0000ff')};`,
    ].join(' '), { typography: false });
  }
  ```
Render branch (display-only; no event wiring):
  ```svelte
  {:else if node.type === 'LinearProgress'}
    {#if boolValue(props.Indeterminate, true)}
      <div
        class="sim-linearprogress sim-linearprogress--indeterminate"
        class:sim-unsupported={unsupportedHere}
        style={lpStyle()}
        data-sim-component={node.name}
        role="progressbar"
        aria-busy="true"
      ><span class="sim-lp-bar"></span></div>
    {:else}
      <progress
        class="sim-linearprogress"
        class:sim-unsupported={unsupportedHere}
        style={lpStyle()}
        data-sim-component={node.name}
        max="1"
        value={lpFraction()}
      ></progress>
    {/if}
  ```
CSS (add near `.sim-slider` rules), reusing the CSS-var theming pattern:
  ```css
  .sim-linearprogress { display: block; width: 100%; min-width: 80px; height: 6px;
    border: 0; border-radius: 999px; appearance: none; -webkit-appearance: none;
    background: rgba(0,0,0,0.12); overflow: hidden; }
  .sim-linearprogress::-webkit-progress-bar { background: rgba(0,0,0,0.12); border-radius: 999px; }
  .sim-linearprogress::-webkit-progress-value { background: var(--lp-progress, #0000ff); border-radius: 999px; }
  .sim-linearprogress::-moz-progress-bar { background: var(--lp-progress, #0000ff); border-radius: 999px; }
  .sim-linearprogress--indeterminate { position: relative; }
  .sim-linearprogress--indeterminate .sim-lp-bar {
    position: absolute; top: 0; bottom: 0; width: 40%; border-radius: 999px;
    background: var(--lp-indeterminate, #0000ff);
    animation: sim-lp-sweep 1.4s ease-in-out infinite; }
  @keyframes sim-lp-sweep {
    0%   { left: -40%; width: 40%; }
    50%  { left: 30%;  width: 60%; }
    100% { left: 100%; width: 40%; }
  }
  ```
  Reuse `Visible` handling (the wrapper already hides via existing `visible` logic). Note the `<progress>` value is normalized to `max="1"` with `lpFraction()` so `Minimum`/`Maximum`/`Progress` math stays in JS and avoids per-browser quirks. Pattern reuse: CSS-custom-prop color theming from Slider; `baseStyle()` for width/position/visibility.

### SimulateOverlay.svelte
None. No dialogs, toasts, or runtime effects are required. LinearProgress renders purely from component state; `ProgressChanged` flows through the normal `runEvent` path, not via overlay effects.

### simulation_wasm.go
A small per-type handler is warranted for the `Progress` write + `ProgressChanged` event and for the method:

- **SetProperty**: add a `LinearProgress` clause that intercepts `Progress` (and ideally `Minimum`/`Maximum` for clamping symmetry, like `setSliderProperty`). When `Progress` changes, clamp to `[Minimum, Maximum]`, store it, then fire `ProgressChanged` via `h.runEvent` with the reported value — `0` if `Indeterminate` is true, otherwise the clamped progress. Mirror the TextBox `TextChanged` "only fire if changed and not suppressed" guard.
  ```go
  if componentType == "LinearProgress" {
      switch property {
      case "Progress", "Minimum", "Maximum":
          h.setLinearProgressProperty(componentName, property, value)
          return
      }
  }
  ```
  `setLinearProgressProperty` reads `Minimum`/`Maximum`/`Progress` (like `setSliderProperty`), clamps via `clampFloat`, stores, and on a `Progress` change calls `runEvent("ProgressChanged", [reported])`.
- **GetProperty**: generic store path suffices (`Progress`, `Indeterminate`, colors, `Minimum`, `Maximum` all live in `h.state`). No computed getters needed (`Column`/`Row` are invisible and can stay unsupported/null).
- **CallMethod**: add a `case "LinearProgress"` that handles `IncrementProgressBy`: read current `Progress`, add `args[0]`, route through `setLinearProgressProperty` (so clamping + `ProgressChanged` reuse the same path). Any other method -> `h.Unsupported`.
- **Effects**: none.

If you prefer minimal Go, the generic `setProperty` path already stores `Progress` and the renderer would update — but then `ProgressChanged` would never fire and `IncrementProgressBy` would be `Unsupported`. The small handler above is the correct, low-cost choice.

### design-schema-tree.js
Containment already handled. LinearProgress is a normal visible non-container leaf: `canContainDesignComponent` returns true for any standard container parent (it is not Canvas/Map/Chart/non-visible), and it itself is never a parent. `designTreeToInitialState` will merge `SIMULATION_DEFAULTS.LinearProgress` automatically once the defaults block exists. `unsupportedSimulationComponents` will stop flagging it once it is in `SIMULATION_SUPPORTED_TYPES`. No edits needed.

## Dependencies & Ordering
No external libraries (native `<progress>` + CSS only). No other components must be implemented first — LinearProgress is self-contained and depends only on existing helpers (`numberOr`, `boolValue`, `colorValue`, `baseStyle`, `clampFloat`). Requires the standard `npm run build:wasm` after the Go change.

## Web-Platform Limitations & Fidelity Caveats
- **Minimum on old Android**: real devices ignore `Minimum` below API 26; the simulator always honors it. Behavior may look more correct in-sim than on a legacy device.
- **Indeterminate animation timing/easing** will not be pixel-identical to Android's Material `LinearProgressIndicator` sweep; the CSS keyframe is an approximation (single sweeping bar) and tuned for plausibility, not exact parity.
- **Bar thickness / corner radius / track color** are simulator styling choices (Material-ish) since the spec exposes no height/track-color/corner props; they will not match a specific device theme exactly.
- **`<progress>` shadow-part styling** (`::-webkit-progress-value`, `::-moz-progress-bar`) is browser-specific; the fill color is honored on Chromium/Firefox but theming nuances differ slightly between engines. The indeterminate branch sidesteps this by using a custom div.
- `Column`/`Row` getters are not simulated (invisible blockProperties; return null) — negligible.

## Effort Estimate
**S.** One defaults block + six visual-prop/coercion registrations in capabilities; one display-only render branch + ~2 helpers + CSS in SimulationComponent (no event wiring); one small `setLinearProgressProperty` + SetProperty/CallMethod clauses in Go for `Progress`/`ProgressChanged`/`IncrementProgressBy`. No overlay, no library, no containment work.
