# FilePicker Simulator Implementation Plan

## Overview
`FilePicker` is a button-like component in the **Media** category. Per the spec's
`helpString`: "a button-like component that when clicked by the user will prompt
them to select a file from the system. The picker can also be programmatically
opened by calling its `Open` method." It is **visible** (`nonVisible: "false"`).

In App Inventor it renders exactly like a `Button` (same `BackgroundColor`,
`Text`, `Image`, `Shape`, `FontSize`, `TextColor`, `TextAlignment`,
`ShowFeedback`, `Enabled` surface), and when tapped opens the platform file
chooser. After the user selects, it fills its read-only `Selection` property
with a content URI / path and fires `AfterPicking`. The `Action` property
(`Pick Existing File` / `Pick Directory` / `Pick New File`) and `MimeType`
(`*/*` default) shape what the native chooser presents.

Container relationship: a standalone visible button-style picker. It has no
children and is placed inside any arrangement/screen — exactly like `Button`,
`ListPicker`, `DatePicker`. No special containment.

## Feasibility Verdict
**Partially feasible.**

The visual surface is 100% feasible — it is a `Button` clone and reuses every
existing button helper. The interaction is *more* feasible than most pickers
because the browser has a genuine native file chooser: a hidden
`<input type="file">` actually lets the user pick a real file. That gives us a
true `BeforePicking` -> open chooser -> `AfterPicking` flow with a real
`Selection` value.

What is **not** faithfully reproducible:

- **`Selection` is a browser object URL / filename, not an Android content URI.**
  On device, `Selection` is something like `content://.../document/...` or a
  file path that `File`/`TinyDB`/image components can re-open. In the browser we
  can only expose a `blob:` object URL (revoked on reload) or the bare filename.
  Apps that pass `Selection` to a `File.ReadFrom`-style read will not get real
  bytes back unless those components are also simulated against the same blob.
- **`Action = "Pick Directory"`** has no portable web equivalent. The
  `showDirectoryPicker()` File System Access API exists only in Chromium and is
  not available cross-browser; the standard `<input type="file" webkitdirectory>`
  selects a folder's *contents*, not a directory handle. Realistic
  approximation: treat `Pick Directory` like `Pick Existing File` and surface
  the chosen file's parent name, or mark it `Unsupported` and fall back.
- **`Action = "Pick New File"`** (save-as) maps to `showSaveFilePicker()` —
  Chromium-only, not portable. Approximation: prompt for a name via a dialog or
  just synthesize a `Selection` filename without writing anything to disk.
- **`MimeType`** maps cleanly to the `<input accept="...">` attribute (`*/*`
  becomes no filter), so this one *is* honorable.

The realistic baseline (and what this plan implements): render the button,
fire `BeforePicking` on click, open a hidden `<input type="file" accept={MimeType}>`,
and on `change` set `Selection` to the file's object URL (or name) and fire
`AfterPicking`. `Pick Directory`/`Pick New File` degrade gracefully (filename
synthesized, no real FS write) with an honest caveat.

## Properties
| Property | AI default | Visual/Behavioral | How to simulate | Priority |
| --- | --- | --- | --- | --- |
| `Text` | `""` | Visual | Button label, reuse `props.Text` like Button | High |
| `BackgroundColor` | `&H00000000` (none) | Visual | `colorValue()` via `baseStyle()` | High |
| `TextColor` | `&H00000000` | Visual | `baseStyle()` color | High |
| `Enabled` | `True` | Behavioral | `disabled={!enabled}`, gate click | High |
| `Visible` | `True` | Visual | handled by common `visible` gate | High |
| `FontSize` | `14.0` | Visual | `baseStyle()` typography | Med |
| `FontBold` | `False` | Visual | `baseStyle()` | Med |
| `FontItalic` | `False` | Visual | `baseStyle()` | Med |
| `FontTypeface` | `0` | Visual | `typefaceStyle()` | Low |
| `Shape` | `0` | Visual | `shapeStyle()` (rounded/rect/oval) | Med |
| `TextAlignment` | `1` (center) | Visual | `textAlignStyle()` | Med |
| `ShowFeedback` | `True` | Visual | `sim-no-feedback` class | Low |
| `Image` | `""` | Visual | `resolveAssetUrl(assets, props.Image)` as bg, like ListPicker/Button image | Med |
| `MimeType` | `*/*` | Behavioral | maps to `<input accept>`; `*/*` -> no filter | High |
| `Action` | `Pick Existing File` | Behavioral | picks chooser mode; only "Pick Existing File" is faithful (see Feasibility) | Med |
| `Selection` | — (read-only block prop) | Behavioral | object URL / filename of chosen file; set on `change` | High |
| `Left` / `Top` | `""` | Visual | `positionStyle()` (AbsoluteArrangement only) | Low |
| `Width` / `Height` | `-1` | Visual | `sizeStyle()` | High |

Cannot be honored in a browser: `Selection` as a real content URI; `Action`
values other than `Pick Existing File` (no portable directory/save picker).

## Events
| Event | Args | How/when the simulator fires it |
| --- | --- | --- |
| `BeforePicking` | none | On button `click` (and on `Open` method), before opening the file input. Reuse `await emitEvent('BeforePicking')` then trigger the hidden input — same shape as `openListPicker`. |
| `AfterPicking` | none | On the hidden `<input>` `change` event after a file is chosen: emit `Selection` property patch + `AfterPicking` via `emitInteraction`. If the user cancels the chooser, **nothing fires** (correct — matches device). |
| `GotFocus` | none | On button `focus` -> `focusEvent()`. |
| `LostFocus` | none | On button `blur` -> `blurEvent()`. |
| `TouchDown` | none | On `pointerdown` -> `pointerDown(false)`. |
| `TouchUp` | none | On `pointerup` -> `pointerUp()`. |

All six events are firable in the browser. None require a real sensor or
permission. The only caveat: browser file-chooser cancellation is not
observable for `Pick New File`/`Pick Directory` fallbacks, but for the standard
`<input type="file">` cancel simply yields no `change`, which is the right
behavior.

## Methods
| Method | Signature | Simulated behavior (or Unsupported + why) |
| --- | --- | --- |
| `Open` | `Open()` -> void | Programmatically opens the picker exactly as a click would: fire `BeforePicking` then trigger the hidden file input. Implemented as a Go `CallMethod` case that runs `BeforePicking` and pushes a `component-action` `open` effect; the renderer's `handleComponentActions` already maps `open`/`Open` to an action token, so wire the FilePicker open handler into that path (same as ListPicker's `Open`). |

## Implementation Plan

### simulation-capabilities.js
- Add `'FilePicker'` to `SIMULATION_SUPPORTED_TYPES`. (It is visible, so do
  **not** add to `SIMULATION_NONVISIBLE_TYPES`.)
- Add a defaults block in `buildSimulationDefaults()`. It is a button clone plus
  two behavioral props:
  ```js
  FilePicker: {
    ...COMMON_VISIBLE_PROPS,
    ...BUTTON_STYLE_PROPS,
    Text: '',
    BackgroundColor: '',          // &H00000000 -> none, matches BUTTON_STYLE_PROPS
    TextColor: '&HFF000000',      // AI default &H00000000 is "none"; use black for legibility
    Action: 'Pick Existing File',
    MimeType: '*/*',
    Selection: '',
  },
  ```
- Add new prop names to `SIMULATION_VISUAL_PROPS`: `Action`, `MimeType`.
  (`Selection`, `Text`, `Width`, `Height`, `Shape`, `ShowFeedback`, `Image`,
  font props, `TextAlignment` are already in the set.)
- `isBooleanProp`: no additions (no new booleans).
- `isNumericProp`: no additions (`Action`/`MimeType` are strings; `Selection` is text).
- `coerceSimulationValue`: no special case needed — `Action`, `MimeType`,
  `Selection` are plain strings and fall through to the default `return value`.
- `deriveStateFromDesignerProps`: no derived state needed.

### SimulationComponent.svelte
Add a render branch modeled on the `ListPicker` branch (button + hidden trigger),
combining the `DatePicker` hidden-`<input>` pattern for the native chooser:

```svelte
{:else if node.type === 'FilePicker'}
  <div class="sim-native-picker-wrap" class:sim-unsupported={unsupportedHere}
       style={containerStyle()} data-sim-component={node.name}>
    <button
      bind:this={buttonEl}
      type="button"
      class="sim-button"
      class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
      style={buttonInnerStyle('width: 100%; height: 100%;')}
      disabled={!enabled}
      on:pointerdown={() => pointerDown(false)}
      on:pointerup={pointerUp}
      on:focus={focusEvent}
      on:blur={blurEvent}
      on:click={openFilePicker}
    >{props.Text ?? ''}</button>
    <input bind:this={fileInput} class="sim-native-picker-input" type="file"
           accept={props.MimeType && props.MimeType !== '*/*' ? props.MimeType : undefined}
           disabled={!enabled} on:change={fileChange}
           tabindex="-1" aria-hidden="true" />
  </div>
```

New script functions (mirror `openListPicker` + `dateChange`):
```js
let fileInput;

async function openFilePicker() {
  if (!enabled || consumeLongClick()) return;
  await emitEvent(node.name, 'BeforePicking');
  await tick();
  if (!enabled || !visible) return;
  triggerNativePicker(fileInput);   // existing helper: focus + showPicker()/click()
}

function fileChange(e) {
  const file = e.currentTarget.files && e.currentTarget.files[0];
  if (!file) return;                 // user cancelled -> no AfterPicking
  const selection = URL.createObjectURL(file); // or file.name for plain-text fidelity
  emitInteraction(
    [{ component: node.name, property: 'Selection', value: selection }],
    { component: node.name, event: 'AfterPicking', args: [] },
  );
}
```
- Reuse helpers: `buttonInnerStyle`, `containerStyle`, `boolValue`,
  `triggerNativePicker`, `pointerDown`/`pointerUp`, `focusEvent`/`blurEvent`,
  `consumeLongClick`, `emitEvent`, `emitInteraction`.
- If `props.Image` resolves to an asset, render it as a button background the
  same way the Button/ListPicker branch handles `Image` (via `assetUrl` /
  `resolveAssetUrl(assets, props.Image)`) — optional polish.
- Wire the programmatic `Open` into `handleComponentActions`: the existing
  `runAction(actionState, 'open', ['open','Open','DisplayDropdown','LaunchPicker'], ...)`
  block should also, for `node.type === 'FilePicker'`, call `openFilePicker()`
  (or directly `triggerNativePicker(fileInput)` after `BeforePicking`). Add a
  `node?.type === 'FilePicker'` clause there.
- CSS: none new — reuse `.sim-button`, `.sim-no-feedback`,
  `.sim-native-picker-wrap`, `.sim-native-picker-input`.

### SimulateOverlay.svelte
**None required** for the standard path. The overlay already routes
`component-action` effects with action `open` into `handleComponentActions`,
which is all `Open` needs. (Only if we later add a "Pick New File" save-name
dialog would we touch the overlay; out of scope for the baseline.)

### simulation_wasm.go
- `GetProperty`: generic store path suffices. `Selection` is written via the
  `Selection` SetProperty path (already routed through `setSelection`); but note
  FilePicker `Selection` is a *plain text* value, not list-index-derived. The
  existing `setSelection` switch only special-cases `Spinner`/`ListPicker`/`ListView`;
  for any other type it should fall through to a plain `setProperty(componentName,
  "Selection", value)`. Verify the `default` branch of `setSelection` stores the
  raw value (add a `default:` if absent) so FilePicker `Selection` round-trips.
- `SetProperty`: `Action`, `MimeType`, `Text`, colors, font props all use the
  generic `setProperty` path — no new case needed.
- `CallMethod`: add a `case "FilePicker":` mirroring ListPicker's `Open`:
  ```go
  case "FilePicker":
      if method == "Open" {
          if h.runEvent != nil {
              h.runEvent(componentName, componentType, "BeforePicking", nil)
          }
          h.effects = append(h.effects, componentAction(componentName, "open"))
          return runtime.VoidVal()
      }
  ```
- Events: `BeforePicking`/`AfterPicking`/`GotFocus`/`LostFocus`/`TouchDown`/
  `TouchUp` all flow through the generic `h.runEvent` / front-end `emitEvent`
  path; no per-event Go code beyond the `Open` method above.
- No new effect type needed (reuses `component-action` / `open`).
- Rebuild with `npm run build:wasm`.

### design-schema-tree.js
Containment already handled. FilePicker is a leaf (no children) and lives in any
arrangement/screen; `canContainDesignComponent` needs no entry. Once `FilePicker`
is in `SIMULATION_SUPPORTED_TYPES`, `unsupportedSimulationComponents()` stops
flagging it and `designTreeToInitialState()` will merge the new
`SIMULATION_DEFAULTS.FilePicker` block automatically.

## Dependencies & Ordering
- **No external libraries.** Uses the native `<input type="file">` and
  `URL.createObjectURL` — both universally available.
- **No prerequisite components.** It is self-contained. (Button/ListPicker/
  DatePicker patterns are already in place and only serve as references, not
  runtime dependencies.)
- Optional future dependency: if `Selection`'s blob URL is meant to be consumed
  by a future `File`/`Image` read, those components would need to resolve the
  same object URL — but that is out of scope here.

## Web-Platform Limitations & Fidelity Caveats
- `Selection` is a `blob:` object URL or filename, **not** an Android
  `content://` URI. Apps that string-match or re-open the URI on device will
  diverge.
- `Action = "Pick Directory"` and `"Pick New File"` have no portable web
  equivalent (File System Access API is Chromium-only). The baseline treats
  both like "Pick Existing File" with a synthesized selection; no real directory
  handle and no actual file is written to disk.
- `MimeType` maps to `<input accept>` which is a *hint* the browser may ignore;
  it does not strictly enforce file type the way the Android chooser does.
- Object URLs created via `URL.createObjectURL` are not revoked here and are lost
  on reload — acceptable for a session-scoped simulator but not persistent.
- The browser file chooser is a real OS dialog (not the in-frame phone UI), so
  the picker UI itself does not render inside the phone frame — a minor visual
  divergence from the on-device chooser.
- No file size/permission constraints are simulated; the browser silently allows
  whatever the OS file dialog returns.

## Effort Estimate
**S** — one button-clone render branch + two small handlers (`openFilePicker`,
`fileChange`), a defaults block, two visual-prop names, and one Go `CallMethod`
case for `Open`; no new CSS, no overlay work, no external deps.
