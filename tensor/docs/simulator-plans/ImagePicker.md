# ImagePicker Simulator Implementation Plan

## Overview

ImagePicker is a MEDIA-category, **visible** component in App Inventor. Per the spec helpString: "A special-purpose button. When the user taps an image picker, the device's image gallery appears, and the user can choose an image. After an image is picked, it is saved, and the `Selected` property will be the name of the file where the image is stored. In order to not fill up storage, a maximum of 10 images will be stored. Picking more images will delete previous images, in order from oldest to newest."

Architecturally it is a `Picker → ButtonBase` subclass: it renders exactly like a Button (same `BackgroundColor`, `Font*`, `Image`, `Shape`, `ShowFeedback`, `Text`, `TextAlignment`, `TextColor` surface) and adds a single read-only output property, `Selection` (the file path of the picked image). When tapped it fires `BeforePicking`, launches the system gallery activity, and on return fires `AfterPicking` after filling `Selection`.

**Container relationship:** standalone visible component (no children, no special parent). It can be placed in any standard arrangement/Screen — there are no Canvas/Map/Chart/FeatureCollection containment rules to encode. This is identical to how Button/ListPicker are placed.

## Feasibility Verdict

**Partially feasible.**

The *visual button* surface is fully feasible — it is byte-for-byte the same ButtonBase rendering already implemented for Button and ListPicker, so all designer properties render with full fidelity.

The *picking flow* is partially feasible. A browser cannot access the device's photo **gallery** (no native gallery API, no MediaStore, no permission grant). However, a browser **can** open a native file chooser via `<input type="file" accept="image/*">`, and that file chooser *is* the OS gallery/Photos picker on mobile devices and the file-open dialog on desktop. So the realistic approximation is:

- Tap ImagePicker -> fire `BeforePicking` -> programmatically `.click()` a hidden `<input type="file" accept="image/*">`.
- User selects a local image file -> read it as an object URL (`URL.createObjectURL`) or data URL (`FileReader`) -> set `Selection` to a synthesized path (e.g. `picked_image_<n>.<ext>` or the file name) -> fire `AfterPicking`.

What **cannot** be faithfully simulated:
- **`Selection` is a real on-device file path** (e.g. `/storage/.../Pictures/_app_inventor_image_picker/picked_image1.jpg`). A browser has no filesystem path for the chosen file (the `value` is sanitized to `C:\fakepath\...`). The simulator must synthesize a stable placeholder string. Blocks that parse the path or feed it to `File`/`TinyDB` will get a non-device-shaped string.
- **The "max 10 images, oldest-deleted" storage rotation** is a device storage behaviour with no browser analogue; not simulated (irrelevant to most apps).
- **Selecting the *same* path the app could later load with an `Image.Picture`** — the synthesized object-URL path will not resolve through `resolveAssetUrl()` (which only matches design-time assets). If the user wires `set Image1.Picture to ImagePicker1.Selection`, the Image component will not display the picked photo unless we also register the object URL in the asset map (see Limitations).

The pick *can* fire (it is a user-initiated browser action, no fake sensor), so unlike Camera/LocationSensor this is genuinely interactive, not a stub. Hence **partially feasible** rather than not-feasible: the button + BeforePicking/AfterPicking + a real file pick all work; only the device-path fidelity and storage semantics diverge.

## Properties

Pull from spec `designerProperties`. All ButtonBase props are already implemented for Button/ListPicker, so simulation is "reuse Button rendering".

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Text` | `""` | Visual | Button label; reuse `props.Text` render path. | High |
| `BackgroundColor` | `&H00000000` (COLOR_DEFAULT) | Visual | `baseStyle()`/`buttonStyle()` via `colorValue()`. | High |
| `TextColor` | `&H00000000` (COLOR_DEFAULT) | Visual | `baseStyle()` color. | High |
| `Enabled` | `True` | Behavioral | `disabled` attr; gates the pick handler. | High |
| `Visible` | `True` | Behavioral | Gates whole render (existing). | High |
| `Image` | `""` | Visual | Button background image via `backgroundImageStyle()`/`resolveAssetUrl(assets, props.Image)`. | Medium |
| `Shape` | `0` | Visual | `shapeStyle()` (default/rounded/rect/oval) — already implemented. | Medium |
| `ShowFeedback` | `True` | Visual | `sim-no-feedback` class toggling press feedback — already implemented. | Medium |
| `FontSize` | `14.0` | Visual | `baseStyle()` typography. | Medium |
| `FontBold` | `False` | Visual | `baseStyle()` `font-weight`. | Medium |
| `FontItalic` | `False` | Visual | `baseStyle()` `font-style`. | Medium |
| `FontTypeface` | `0` | Visual | `typefaceStyle()`. | Low |
| `TextAlignment` | `1` (center) | Visual | `textAlignStyle()`. | Medium |
| `Width` / `Height` | `""` (auto) | Visual | `sizeStyle()` (`-1` auto). | High |
| `Left` / `Top` | `""` | Visual | `positionStyle()` (AbsoluteArrangement only). | Low |
| `Selection` | n/a (read-only block prop) | Behavioral output | Synthesized placeholder path set on pick; stored in host state. **Cannot be a real device path.** | High |

Note: `Column`/`Row`/`HeightPercent`/`WidthPercent` are invisible/write-only layout helpers handled by the generic store path; no special work.

## Events

From spec `events` (all params empty).

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `BeforePicking` | none | Fired on button click (in `openImagePicker()`) *before* the file input opens, and by the Go host when `Open()` is called programmatically. Mirrors `openListPicker()`. |
| `AfterPicking` | none | Fired after the user selects a file in the native `<input type="file">` chooser and `Selection` has been set. Emitted via `emitInteraction([{property:'Selection',...}], {event:'AfterPicking'})`. If the user **cancels** the chooser, no event fires (matches AI `RESULT_CANCELED`). |
| `GotFocus` | none | Wire `on:focus` on the button -> `focusEvent` (existing helper). |
| `LostFocus` | none | Wire `on:blur` -> `blurEvent`. |
| `TouchDown` | none | `on:pointerdown` -> `pointerDown(false)` (existing). |
| `TouchUp` | none | `on:pointerup` -> `pointerUp` (existing). |

None of these are blocked by the browser — all are user-initiated. There is no sensor/permission gate, so every ImagePicker event can fire. (One fidelity gap: AI's `BeforePicking` is a fallible event that can cancel opening; like the existing ListPicker, we will not honor the cancel return — documented as a known limitation.)

## Methods

From spec `methods`.

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `Open()` | `Open() -> void` | Opens the picker as though tapped. Go host fires `BeforePicking` then emits a `component-action` effect with `action: "open"`; the overlay relays it; the component's `handleComponentActions()` triggers the file input `.click()`. Mirrors the existing `ListPicker.Open()` path exactly. Note: a programmatic `.click()` on a hidden file input may be blocked by browser pop-up/gesture policies if not within a user-gesture stack — acceptable degradation. |

## Implementation Plan

### simulation-capabilities.js

1. Add `'ImagePicker'` to `SIMULATION_SUPPORTED_TYPES`. (Do **not** add to `SIMULATION_NONVISIBLE_TYPES` — it is visible.)

2. Add a defaults block in `buildSimulationDefaults()`. It is structurally identical to `ListPicker`/`Button` (ButtonBase props), plus `Selection`:

```js
ImagePicker: {
  ...COMMON_VISIBLE_PROPS,
  ...BUTTON_STYLE_PROPS,
  Text: '',
  BackgroundColor: '',
  TextColor: '&HFF000000',
  Selection: '',
},
```
(`BUTTON_STYLE_PROPS` already supplies `FontSize/FontBold/FontItalic/FontTypeface/TextColor/TextAlignment/Image/Shape/ShowFeedback`. The explicit `BackgroundColor: ''` / `TextColor` mirror the COLOR_DEFAULT convention used elsewhere.)

3. `SIMULATION_VISUAL_PROPS`: no new names required. `Selection`, `Image`, `Shape`, `ShowFeedback`, `TextAlignment`, all `Font*` props, `Text`, colors, sizing are **already** in the set. (`Selection` is present from the Spinner/ListPicker work.)

4. `isBooleanProp` / `isNumericProp`: no additions — `Enabled`/`Visible`/`ShowFeedback`/`FontBold`/`FontItalic` are already booleans; `Shape`/`TextAlignment`/`FontSize` are already numeric.

5. `coerceSimulationValue`: no additions — `Selection` is a text passthrough (default branch returns the value). No `deriveStateFromDesignerProps` derived state needed.

### SimulationComponent.svelte

Add a render branch after the `ListPicker` branch. It is the `ListPicker` button minus the dropdown menu, plus a hidden file input:

```svelte
{:else if node.type === 'ImagePicker'}
  <div class="sim-picker" class:sim-unsupported={unsupportedHere} style={containerStyle()} data-sim-component={node.name}>
    <button
      bind:this={buttonEl}
      type="button"
      class="sim-button"
      class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
      style={buttonInnerStyle('width: 100%;')}
      disabled={!enabled}
      on:pointerdown={() => pointerDown(false)}
      on:pointerup={pointerUp}
      on:focus={focusEvent}
      on:blur={blurEvent}
      on:click={openImagePicker}
    >{props.Text ?? ''}</button>
    <input
      bind:this={imagePickerInput}
      class="sim-native-picker-input"
      type="file"
      accept="image/*"
      on:change={imagePicked}
      tabindex="-1"
      aria-hidden="true"
    />
  </div>
```

New script glue (reuse existing helpers — `emitEvent`, `emitInteraction`, `pointerDown`, `pointerUp`, `focusEvent`, `blurEvent`, `buttonInnerStyle`, `containerStyle`, the `sim-native-picker-input` CSS already exists for DatePicker/TimePicker):

```js
let imagePickerInput;

async function openImagePicker() {
  if (!enabled || consumeLongClick()) return;
  await emitEvent(node.name, 'BeforePicking');
  await tick();
  if (!enabled || !visible) return;
  imagePickerInput?.click();
}

function imagePicked(e) {
  const file = e.currentTarget.files?.[0];
  if (!file) return; // user cancelled -> no AfterPicking
  const name = file.name || `picked_image_${Date.now()}.jpg`;
  emitInteraction(
    [{ component: node.name, property: 'Selection', value: name }],
    { component: node.name, event: 'AfterPicking', args: [] },
  );
  e.currentTarget.value = ''; // allow re-picking the same file
}
```

Add `ImagePicker` to the `handleComponentActions()` `'open'` branch so the programmatic `Open()` effect calls `imagePickerInput?.click()`:

```js
if (node?.type === 'ImagePicker') imagePickerInput?.click();
```

CSS: reuse `.sim-button`, `.sim-picker`, `.sim-no-feedback`, `.sim-native-picker-input` — no new classes needed.

Optional fidelity upgrade (Medium): also stash the picked image's object URL so a downstream `Image.Picture = ImagePicker.Selection` can display it. This requires registering `{ name, url }` into the `assets` array passed to `resolveAssetUrl` (or a session-scoped picked-images map). Defer to a follow-up — see Limitations.

### SimulateOverlay.svelte

**None required for the basic flow** beyond what already exists. The `Open()` -> `component-action` / `action:"open"` effect is already relayed by `applyEffects()` (it increments the action counter that `handleComponentActions` reads). No new dialog, toast, or effect type is needed because the file chooser is native browser chrome, not an in-app modal.

(If the optional object-URL-into-assets upgrade is pursued, the overlay is where the picked-image map would live, since it owns the session/asset context — but that is out of scope for the minimal implementation.)

### simulation_wasm.go

Add an `ImagePicker` case to `CallMethod` mirroring `ListPicker.Open()`:

```go
case "ImagePicker":
    if method == "Open" {
        if h.runEvent != nil {
            h.runEvent(componentName, componentType, "BeforePicking", nil)
        }
        h.effects = append(h.effects, componentAction(componentName, "open"))
        return runtime.VoidVal()
    }
```

- **GetProperty:** generic store path suffices. `Selection` is read from `h.state[componentName]` like any stored prop; no computed override needed (unlike DatePicker `Instant`).
- **SetProperty:** generic path suffices. `Selection` is read-only in blocks, but if set it falls into `setSelection()`'s `default` case (`h.setProperty(componentName, "Selection", value)`) which is harmless. Other props use the generic `h.setProperty`.
- **Events:** `AfterPicking`/`GotFocus`/`LostFocus`/`TouchDown`/`TouchUp` are dispatched from the frontend via `emitEvent`/`emitInteraction` -> `runEvent`; no host-side firing needed except `BeforePicking` on the programmatic `Open()` path above.
- Do **not** call `h.Unsupported` — the component is genuinely interactive.

### design-schema-tree.js

**Already handled.** `canContainDesignComponent()` falls through to `containerTypes().has(parent)` for any standard visible component, and ImagePicker is neither a Canvas/Map/Chart/FeatureCollection child nor a non-visible type. Once `'ImagePicker'` is added to `SIMULATION_SUPPORTED_TYPES`, `unsupportedSimulationComponents()` stops flagging it and `designTreeToInitialState()` merges `SIMULATION_DEFAULTS.ImagePicker` automatically. No changes to this file.

## Dependencies & Ordering

- **External libraries:** none. Uses the native `<input type="file" accept="image/*">` and the existing style/event helpers.
- **Prerequisite components:** none strictly required. The implementation directly reuses Button/ListPicker patterns that are already in place, so those serve as the reference (no ordering dependency — they are already shipped).
- **WASM rebuild:** the Go `CallMethod` change requires `npm run build:wasm` (GOOS=js GOARCH=wasm) and refreshing `public/falcon.wasm` / `web/falcon.wasm`.

## Web-Platform Limitations & Fidelity Caveats

- **`Selection` is not a real device file path.** It is a synthesized name (the file's `name`, or a generated `picked_image_<ts>.jpg`). Blocks that expect an Android-shaped absolute path, or that feed `Selection` to `File`/sharing components, will behave differently than on device. The browser deliberately hides the true path (`C:\fakepath`).
- **No gallery, only a file chooser.** On desktop this is the OS file-open dialog; on mobile browsers it surfaces the Photos/gallery picker. The "device image gallery activity" look-and-feel is replaced by native browser chrome.
- **Picked image will not auto-display in an `Image` component** unless the optional object-URL-into-assets upgrade is implemented, because `resolveAssetUrl()` only matches design-time assets. By default `Image1.Picture = ImagePicker1.Selection` shows the placeholder string, not the photo.
- **No 10-image storage rotation.** The "max 10 images, oldest deleted" persistence is a device-storage behaviour with no browser equivalent; not simulated.
- **`BeforePicking` cannot cancel the pick.** Like the existing ListPicker, the simulator fires `BeforePicking` and always proceeds to open the chooser; AI's fallible-event cancel semantics are not honored.
- **Programmatic `Open()` may be gesture-blocked.** A `.click()` on a hidden file input that is not inside a fresh user-gesture stack can be silently ignored by browser pop-up policies. User-initiated taps always work.
- **COLOR_DEFAULT appearance.** Like Button/ListPicker, the baseline `sim-button` CSS does not reproduce the Android theme's default button color when `BackgroundColor` is `&H00000000`; it uses the simulator's button styling.

## Effort Estimate

**S** — Reuses the ListPicker/DatePicker button + hidden-input pattern wholesale; net new work is one defaults block, one render branch (~25 lines), two small handlers, and a 6-line Go `Open()` case. The only nuance is honestly documenting `Selection` path fidelity. Optional object-URL-into-assets display upgrade would push it to **M**.
