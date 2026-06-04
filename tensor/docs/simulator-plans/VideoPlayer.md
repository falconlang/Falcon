# VideoPlayer Simulator Implementation Plan

## Overview

`VideoPlayer` (category **MEDIA**, `nonVisible = false`) is a visible component that plays video files. Per the spec's `helpString`, it "is displayed as a rectangle on-screen. If the user touches the rectangle, controls will appear to play/pause, skip ahead, and skip backward within the video." The app can also drive playback programmatically with `Start`, `Pause`, `Stop`, and `SeekTo`. Sources may be a bundled asset (`.3gp`/`.mp4`) added in the Designer, or a URL pointing directly at a streaming video file.

Container relationship: standalone visible component. It is a leaf — it has no children and may be placed inside `Form`/`Screen` or any arrangement, exactly like `Image`. There is no parent/child containment rule to encode.

In the simulator it currently renders as a dashed "unsupported" placeholder because `VideoPlayer` is absent from `SIMULATION_SUPPORTED_TYPES`.

## Feasibility Verdict

**Feasible.**

The HTML `<video>` element is a near-exact web-platform analog for an Android `VideoPlayer`. Unlike `WebViewer`'s `<iframe>` (which is crippled by cross-origin embedding policy), `<video>` is genuinely usable in a browser: it loads from an asset blob URL or a remote URL, displays the picture in a rectangle, exposes native player controls on demand, fires precise lifecycle events (`loadedmetadata`, `ended`, `error`, `timeupdate`), and is fully scriptable (`.play()`, `.pause()`, `.currentTime`, `.volume`, `.duration`, `.requestFullscreen()`). Every property, event, and method in the spec maps to a real `<video>` capability with only minor caveats:

- **Codec/format coverage differs.** App Inventor recommends `.3gp`/`.mp4`. Browsers reliably decode `.mp4` (H.264/AAC) and WebM/Ogg; `.3gp` support is spotty (Chrome/Firefox generally will not play `.3gp`). A `.3gp` asset may fail to decode and fire the `error` path. This is a codec-availability gap, not an architectural blocker.
- **Remote URLs are subject to CORS for some operations.** Plain playback of a cross-origin URL works (media elements are not same-origin-restricted for playback), so this is not a real limitation for VideoPlayer's surface; `crossorigin` is only needed for canvas pixel access, which VideoPlayer does not expose.
- **Browser autoplay/volume policy.** Browsers block autoplay-with-sound until a user gesture, and on iOS/Safari `video.volume` is read-only (system-controlled). `Start()` invoked from a non-gesture block may be deferred/muted by the browser, and `Volume` may not audibly change on iOS. Functionally the simulator still tracks state correctly; only the audible result diverges.

These are fidelity caveats, not feasibility blockers. The component renders and behaves correctly with a single `<video>` element plus host-tracked playback state.

## Properties

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Source` | `""` | Behavioral (the media to load) | `resolveAssetUrl(assets, props.Source)` for bundled assets; if the value looks like a URL (`http(s)://`), use it verbatim. Set as `<video src>`. Empty → render the placeholder rectangle. Designer editor is `asset`; runtime block is **write-only text**. | High |
| `Volume` | `50` | Behavioral (0–100) | Map to `<video>.volume` = `clamp(Volume,0,100)/100`. **Write-only** at runtime per spec. Volume 0 should mute. | High |
| `Visible` | `True` | Visual | Standard — gates the render branch via shared `isSimulationVisible`. | High |
| `Width` / `Height` | `""` (auto) | Visual | `sizeStyle()` handles `-1`/`-2`/percent. Auto height gets a sensible min-height fallback (VideoPlayer is usually explicitly sized or Fill). `HeightPercent`/`WidthPercent` are write-only → map to the percent size encoding `sizeStyle()` understands. | High |
| `Left` / `Top` | `""` | Visual (AbsoluteArrangement only) | Already handled by `positionStyle()`; `Left`/`Top` are already in `SIMULATION_VISUAL_PROPS`. | Low |
| `FullScreen` | (block-only, default `false`) | Behavioral | **read-write boolean.** Setting `true` → call `video.requestFullscreen()` via an effect/action; setting `false` → `document.exitFullscreen()`. Reading reflects host-tracked fullscreen state. Subject to browser fullscreen-needs-user-gesture policy (see caveats). | Medium |
| `Column` / `Row` | (block-only, `rw: invisible`) | Designer grid placement | Invisible getters used only inside `TableArrangement`; no runtime semantics to simulate. Ignore. | N/A |

Note: VideoPlayer has **no `Enabled`, `BackgroundColor`, font, or text properties** in the spec — do not add any of those to its defaults (same discipline as the `Image` correction noted in the reviews; `Image` deliberately omits `Enabled`).

## Events

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `Completed` | (none) | Fired from the frontend on the `<video>` `ended` event → `emitEvent(node.name, 'Completed')`. Maps exactly to playback reaching the end. |
| `VideoPlayerError` | `message` (text) | **Deprecated in App Inventor** ("no longer used. Please use Screen.ErrorOccurred instead"). For fidelity we wire the `<video>` `error` event and route it to `Screen.ErrorOccurred` (the real runtime path); optionally also fire `VideoPlayerError(message)` for legacy apps that still handle it. `message` derived from `video.error.code`/`MediaError`. A `.3gp` or otherwise undecodable source is the most common trigger. |

There are no per-frame/position-change events in the spec (no `PositionChanged`), so no `timeupdate`-driven event is required; `timeupdate` is only used internally to keep `currentTime` for `GetDuration`/`SeekTo` bookkeeping.

## Methods

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `Start()` | `() → void` | Emit a `media-play` component-action effect; overlay/renderer calls `video.play()`. Subject to autoplay policy (may be deferred until a user gesture). Resumes from current position (App Inventor semantics). |
| `Pause()` | `() → void` | Emit a `media-pause` action → `video.pause()`, leaving `currentTime` intact so `Start` resumes. |
| `Stop()` | `() → void` | Emit a `media-stop` action → `video.pause()` then set `video.currentTime = 0` (App Inventor: "resets to start of video and pauses it"). |
| `SeekTo(ms)` | `(number) → void` | Emit a `media-seek` action carrying `ms` → set `video.currentTime = ms/1000`. Honors the spec note that a paused video's displayed frame may not update for non-keyframe seeks (browser behavior approximates this naturally). |
| `GetDuration()` | `() → number` | Returns duration in **milliseconds**. Frontend reports `video.duration*1000` back to the host (via a state patch / `loadedmetadata`); `GetDuration` returns the host-tracked value, or `0` if metadata not yet loaded. Because `CallMethod` is synchronous in the host while the actual duration lives in the DOM, the host must already hold a cached `__duration` (pushed from the frontend on `loadedmetadata`); return that. |

`GetDuration` is the only method with a non-void return and is the one place the DOM→host direction matters: the frontend must patch the measured duration into host state (e.g. a hidden `Duration` value) so the synchronous Go getter can return it.

## Implementation Plan

### simulation-capabilities.js

- Add `'VideoPlayer'` to `SIMULATION_SUPPORTED_TYPES`. It is **not** non-visible, so do **not** add it to `SIMULATION_NONVISIBLE_TYPES`.
- Add a defaults block inside `buildSimulationDefaults()`:

```js
VideoPlayer: {
  Visible: true,
  Width: -1,        // VideoPlayer is usually Fill/explicit, not auto
  Height: -1,
  Source: '',
  Volume: 50,
  FullScreen: false,
},
```

- Add new prop names to `SIMULATION_VISUAL_PROPS`: `Source`, `Volume`, `FullScreen`. (`Width`, `Height`, `Left`, `Top`, `Visible` are already present.) Props not in this set are stripped before reaching the renderer, so `Source`/`Volume`/`FullScreen` must be added or the `<video>` wiring will never see them.
- `isBooleanProp`: add `FullScreen`.
- `isNumericProp`: add `Volume`. (`Width`/`Height`/`Left`/`Top` already covered.)
- `coerceSimulationValue`: no special case beyond the above — `Source` is a plain string handled by the default `return value`; `Volume` flows through `coerceNumber`; `FullScreen` through `coerceBoolean`. No `deriveStateFromDesignerProps` branch is required.

### SimulationComponent.svelte

Add a branch after the `Image` branch (closest leaf/asset pattern). Reuse `containerStyle()` for sizing/position and `resolveAssetUrl(assets, …)` for asset resolution (the `$: assetUrl` reactive already calls `resolveAssetUrl(assets, assetName())` — but `assetName()` looks at `Picture/Image/BackgroundImage`, not `Source`, so add a dedicated `videoSrc(props)` helper rather than reusing `assetUrl`).

```svelte
{:else if node.type === 'VideoPlayer'}
  <div
    class="sim-videoplayer"
    class:sim-unsupported={unsupportedHere}
    style={containerStyle('min-height: 120px;')}
    data-sim-component={node.name}
  >
    {#if videoSrc(props)}
      <video
        bind:this={videoEl}
        class="sim-video"
        src={videoSrc(props)}
        controls
        playsinline
        preload="metadata"
        on:loadedmetadata={onVideoMeta}
        on:ended={() => emitEvent(node.name, 'Completed')}
        on:error={onVideoError}
      ></video>
    {:else}
      <div class="sim-videoplayer-empty" aria-hidden="true">
        <span>VideoPlayer</span>
        <small>Set Source to a .mp4 asset or URL</small>
      </div>
    {/if}
  </div>
```

- Helper `videoSrc(props)`: if `props.Source` matches `^https?://`, return it verbatim; else `resolveAssetUrl(assets, props.Source)`; else `''`.
- `onVideoMeta`: read `videoEl.duration`, set `videoEl.volume = clamp(numberOr(props.Volume,50),0,100)/100`, and report the duration back to the host so `GetDuration` works — via `emitInteraction([{ component: node.name, property: 'Duration', value: Math.round(videoEl.duration*1000) }])` (the overlay already applies interaction property patches into state).
- `onVideoError`: derive a message from `videoEl.error?.code` and `emitEvent(node.name, 'Completed')` is **not** fired; instead fire the error path. Because `VideoPlayerError` is deprecated, dispatch to the Screen via the host — simplest is `emitEvent(node.name, 'VideoPlayerError', [message])` and let the host re-route to `Screen.ErrorOccurred`.
- Reactive volume: a `$:` statement `if (videoEl) videoEl.volume = clamp(numberOr(props.Volume,50),0,100)/100;` keeps the element in sync when blocks set `Volume`.
- DOM controls: the `controls` attribute provides the native play/pause/skip UI matching the spec's "controls will appear" behavior — no manual control bar needed.
- Wire `Start`/`Pause`/`Stop`/`SeekTo`/`FullScreen` from the host through the existing `handleComponentActions(actions?.[node.name] ?? {})` reactive path (the same mechanism `open`/`focus`/`cursor-position` use). Add cases for the new action tokens (`media-play`, `media-pause`, `media-stop`, `media-seek`, `fullscreen`) inside `handleComponentActions`, each calling the matching `videoEl` API. `media-seek` reads the carried `ms` from the action state; `fullscreen` calls `videoEl.requestFullscreen()` / `document.exitFullscreen()` based on the carried boolean.
- Add `videoEl` to the `let` declarations near `buttonEl`.
- CSS: `.sim-videoplayer` is a positioned block with `overflow:hidden; background:#000;`; `.sim-video` fills it (`width:100%; height:100%; object-fit:contain; background:#000;`); `.sim-videoplayer-empty` is a centered dashed placeholder on a dark rectangle (mirrors the real "black rectangle until touched" look).

### SimulateOverlay.svelte

The `media-play` / `media-pause` / `media-stop` / `media-seek` / `fullscreen` actions can be modeled as `component-action` effects (exactly like `open`/`focus`/`cursor-position`), which `applyEffects()` already converts into `actionTokens` consumed by `handleComponentActions`. `media-seek`/`fullscreen` carry an extra field (`ms` / `value`) the same way `cursor-position` carries `position` via `componentActionWith`. The duration round-trip uses the existing **interaction property patch** path (the overlay already applies `interaction.properties` into state). **Result: no new dialog/toast/effect handler is required in the overlay** — reuse the existing component-action + interaction plumbing.

### simulation_wasm.go

A dedicated `VideoPlayer` path is needed for the methods and the deprecated-event reroute; property reads/writes are otherwise the generic store path.

- **`GetProperty`:** no computed-property case strictly required. `GetDuration` is a method, not a property; the cached duration (patched from the frontend into `h.state[component]["Duration"]`) is read inside `CallMethod`. `FullScreen` reads the stored boolean. (Optional: add a `case "VideoPlayer"` returning `0` for `Duration` before metadata arrives.)
- **`SetProperty`:** the generic `h.setProperty` path suffices for `Source`, `Volume`, and `FullScreen` (all just patch state, which the frontend reacts to). To make `set FullScreen` actually trigger fullscreen, add a small `componentType == "VideoPlayer"` check: on `FullScreen`, also append a `fullscreen` component-action effect carrying the boolean (`componentActionWith(name, "fullscreen", {"value": ...})`) so the renderer calls `requestFullscreen`/`exitFullscreen`. Volume/Source need no special handling beyond the patch.
- **`CallMethod`:** add `case "VideoPlayer": return h.callVideoPlayerMethod(...)`. Implement:
  - `Start` → append `componentAction(name, "media-play")`.
  - `Pause` → `componentAction(name, "media-pause")`.
  - `Stop` → `componentAction(name, "media-stop")`.
  - `SeekTo` → `componentActionWith(name, "media-seek", {"ms": args[0].AsNum()})`.
  - `GetDuration` → return `runtime.NumVal(valueAsNumber(h.GetProperty(name, "VideoPlayer", "Duration"), 0))` (the duration the frontend cached on `loadedmetadata`); `0` if not yet known.
  - default → `h.Unsupported("method", name+"."+method)`.
- **Events:** `Completed` and `VideoPlayerError` are fired frontend-side (from the `<video>` `ended`/`error` DOM events) through `emitEvent`, so no host `runEvent` call is required to originate them. If routing the deprecated error to `Screen.ErrorOccurred` is desired, handle it where events are dispatched (host or overlay) — but firing `VideoPlayerError` directly is sufficient for v1.
- Recompile with `npm run build:wasm` (`GOOS=js GOARCH=wasm`).

### design-schema-tree.js

No change required. `VideoPlayer` is a leaf with no special containment: `canContainDesignComponent` already lets any ordinary container hold non-container leaves, and `unsupportedSimulationComponents()` stops flagging it once it is in `SIMULATION_SUPPORTED_TYPES`. `designTreeToInitialState()` merges the new `SIMULATION_DEFAULTS.VideoPlayer` automatically.

## Dependencies & Ordering

- **No external libraries.** The render uses a native `<video>` element — no Leaflet/MapLibre-style heavy dependency.
- **No prerequisite components.** VideoPlayer is a standalone leaf; it does not depend on Canvas/Map/Chart child plumbing or on any other component being implemented first. It can be implemented independently. It shares the asset-resolution pattern with `Image` (already implemented), so that pattern is available to copy.

## Web-Platform Limitations & Fidelity Caveats

- **`.3gp` (and some `.mp4` codec profiles) may not decode.** Browsers reliably play H.264/AAC `.mp4` and WebM/Ogg; `.3gp` is generally unsupported in Chrome/Firefox. An undecodable asset fires the `error` path instead of playing. App Inventor's recommended `.3gp` is the weakest spot.
- **Autoplay policy.** A `Start()` issued without a preceding user gesture may be blocked or forced to mute by the browser; the simulator tracks "playing" state correctly but audio/playback may not begin until the user interacts. On a real device `Start()` plays immediately.
- **Volume on iOS/Safari.** `HTMLMediaElement.volume` is read-only on iOS (system-controlled), so `Volume` writes may not change audible level there; on desktop and Android browsers it works.
- **Fullscreen requires a user gesture.** `video.requestFullscreen()` invoked from a non-gesture block may be rejected by the browser; reading `FullScreen` back reflects the simulator's tracked intent, which can briefly diverge from the actual fullscreen state if the request was denied.
- **`GetDuration` is asynchronous in reality but synchronous in the host.** It returns `0` until the frontend has fired `loadedmetadata` and patched the measured duration into state; a block that calls `GetDuration` immediately after setting `Source` (before metadata loads) will see `0`. On a device the value would also be unavailable until the media is prepared, so this is close to real behavior.
- **`VideoPlayerError` is deprecated upstream;** the real path is `Screen.ErrorOccurred`. The simulator fires the error event but the exact `errorCode`/`message` text will differ from Android's `MediaPlayer` error codes.

## Effort Estimate

**M.** The frontend branch is small (single `<video>` element, `videoSrc`/volume helpers, a handful of `handleComponentActions` cases, CSS) and reuses the existing component-action + interaction plumbing, so the overlay needs no new dialogs. The host work is modest (`callVideoPlayerMethod` with five methods, the `GetDuration` cache read, and the optional `fullscreen`/`SetProperty` hook) plus one `npm run build:wasm` cycle. The duration DOM→host round-trip and the codec/autoplay caveats add some care, keeping it above S; the absence of any external dependency, modal dialogs, or containment work keeps it out of L.
