# PhoneNumberPicker Simulator Implementation Plan

## Overview

PhoneNumberPicker is a SOCIAL-category, **visible** component in App Inventor. Per the spec helpString: "A button that, when clicked on, displays a list of the contacts' phone numbers to choose among. After the user has made a selection, the following properties will be set to information about the chosen contact: `ContactName`, `PhoneNumber`, `EmailAddress`, `Picture` (the file name of the contact's image)." Other properties affect the button's appearance (`TextAlignment`, `BackgroundColor`, etc.) and whether it can be clicked (`Enabled`).

Architecturally it is a `ContactPicker → Picker → ButtonBase` subclass. It renders exactly like a Button (same `BackgroundColor`, `Font*`, `Image`, `Shape`, `ShowFeedback`, `Text`, `TextAlignment`, `TextColor` surface) and adds **read-only output properties** filled from the chosen contact: `ContactName`, `ContactUri`, `EmailAddress`, `EmailAddressList`, `PhoneNumber`, `PhoneNumberList`, `Picture`. When tapped it fires `BeforePicking`, launches the system contacts/phone-number activity, and on return fires `AfterPicking` after filling those properties. The distinction from the base `ContactPicker` is that the displayed list is of **phone numbers**, and `PhoneNumber` is set to the specific number the user tapped (not just the contact's primary number).

**Container relationship:** standalone visible component (no children, no special parent). It can be placed in any standard arrangement/Screen — there are no Canvas/Map/Chart/FeatureCollection containment rules to encode. Identical placement to Button/ListPicker/ImagePicker.

## Feasibility Verdict

**Partially feasible.**

The *visual button* surface is fully feasible — it is byte-for-byte the same ButtonBase rendering already implemented for Button/ListPicker/DatePicker, so all designer properties render with full fidelity. The *pick interaction* (tap -> `BeforePicking` -> choose -> `AfterPicking` with output props filled) is feasible as a user-initiated flow. What is **not** feasible is sourcing the choices from the **device's real contacts**.

Web-platform limitations:

- **No reliable browser contacts API.** The only web API is the experimental **Contact Picker API** (`navigator.contacts.select(['name','email','tel'], …)`), which is Chromium-on-Android **only**, requires a secure context and a user gesture, is absent on desktop Chrome, all of Firefox, and all of Safari/iOS, and returns no `ContactUri` or photo file name. It cannot be relied on as the simulator's mechanism — the simulator targets a phone frame inside a desktop/laptop browser where the API does not exist.
- **No device contact store, no `ContactUri`, no `Picture` file.** `ContactUri` is a `content://com.android.contacts/...` URI and `Picture` is the name of a file holding the contact's photo — neither has any browser analogue. There is no MediaStore, no ContactsContract.

Realistic simulated approximation: render the ButtonBase, fire `BeforePicking` on tap, then present an **in-app picker menu of fabricated sample phone numbers** (the same `.sim-picker-menu` dropdown the Spinner/ListPicker branches already use), grouped/labelled by a small built-in roster of fake contacts. When the user taps a number, fill `PhoneNumber` (the tapped number), `ContactName`, `EmailAddress`, `PhoneNumberList`, `EmailAddressList` from that fake contact, set `Picture`/`ContactUri` to empty strings (honestly unavailable), then fire `AfterPicking`. This makes the component **interactive and demonstrable** (apps that read `PhoneNumberPicker1.PhoneNumber` / `.ContactName` after `AfterPicking` get sensible non-empty values) without pretending to read real contacts.

This is **partially feasible** rather than not-feasible: the button + BeforePicking/AfterPicking + a real interactive selection all work; only the *source of the data* is fabricated and `ContactUri`/`Picture` cannot be honored.

## Properties

Designer properties from spec `designerProperties` (all ButtonBase, already implemented for Button/ListPicker). Read-only output properties from `blockProperties` are filled on pick.

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Text` | `""` | Visual | Button label; reuse `props.Text` render path. | High |
| `BackgroundColor` | `&H00000000` (COLOR_DEFAULT) | Visual | `buttonInnerStyle()` via `colorValue()`. | High |
| `TextColor` | `&H00000000` (COLOR_DEFAULT) | Visual | `buttonInnerStyle()` color. | High |
| `Enabled` | `True` | Behavioral | `disabled` attr; gates the pick handler. | High |
| `Visible` | `True` | Behavioral | Gates whole render (existing). | High |
| `Image` | `""` | Visual | Button background image via `resolveAssetUrl(assets, props.Image)`. | Medium |
| `Shape` | `0` | Visual | Shape styling (default/rounded/rect/oval) — already in ButtonBase path. | Medium |
| `ShowFeedback` | `True` | Visual | `sim-no-feedback` class — already implemented. | Medium |
| `FontSize` | `14.0` | Visual | typography via `buttonInnerStyle()`. | Medium |
| `FontBold` | `False` | Visual | `font-weight`. | Medium |
| `FontItalic` | `False` | Visual | `font-style`. | Medium |
| `FontTypeface` | `0` | Visual | typeface styling. | Low |
| `TextAlignment` | `1` (center) | Visual | text alignment. | Medium |
| `Width` / `Height` | `""` (auto, `-1`) | Visual | `sizeStyle()`. | High |
| `Left` / `Top` | `""` | Visual | position (AbsoluteArrangement only). | Low |
| `PhoneNumber` | n/a (read-only) | Behavioral output | Set to the tapped fake number on pick; stored in host state. | High |
| `ContactName` | n/a (read-only) | Behavioral output | Set to the fake contact's name on pick. | High |
| `EmailAddress` | n/a (read-only) | Behavioral output | Set to the fake contact's primary email on pick (may be `""`). | Medium |
| `PhoneNumberList` | n/a (read-only, list) | Behavioral output | List of the fake contact's numbers; set on pick. | Medium |
| `EmailAddressList` | n/a (read-only, list) | Behavioral output | List of the fake contact's emails; set on pick. | Low |
| `ContactUri` | n/a (read-only) | **Cannot honor** | Set to `""` — no browser content:// URI exists. | Low |
| `Picture` | n/a (read-only) | **Cannot honor** | Set to `""` — no contact photo file in a browser. | Low |

Note: `Column`/`Row`/`Shape`(getter)/`FontTypeface`(getter)/`TextAlignment`(getter) are `invisible` block props, and `HeightPercent`/`WidthPercent` are write-only layout helpers — all handled by the generic store path; no special work.

## Events

From spec `events` (all params empty).

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `BeforePicking` | none | Fired on button click (in `openPhonePicker()`) *before* the picker menu opens, and by the Go host when `Open()` is called programmatically. Mirrors `openListPicker()`. |
| `AfterPicking` | none | Fired after the user taps a number in the fabricated `.sim-picker-menu` and the output properties (`PhoneNumber`/`ContactName`/`EmailAddress`/lists) have been set. Emitted via `emitInteraction([...prop patches], { event:'AfterPicking' })`. Dismissing the menu without choosing fires nothing (matches AI `RESULT_CANCELED`). |
| `GotFocus` | none | Wire `on:focus` on the button -> `focusEvent` (existing helper). |
| `LostFocus` | none | Wire `on:blur` -> `blurEvent`. |
| `TouchDown` | none | `on:pointerdown` -> `pointerDown(false)` (existing). |
| `TouchUp` | none | `on:pointerup` -> `pointerUp` (existing). |

None of these are blocked by the browser — all are user-initiated; there is no sensor/permission gate, so every event can fire. One fidelity gap (same as ListPicker): AI's `BeforePicking` is a fallible event that could cancel opening; the simulator fires it and always proceeds.

## Methods

From spec `methods`.

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `Open()` | `Open() -> void` | Opens the picker as though tapped. Go host fires `BeforePicking`, then emits a `component-action` effect with `action:"open"`; the overlay relays it; the component's `handleComponentActions()` opens the fabricated picker menu. Mirrors the existing `ListPicker.Open()` path exactly. |
| `ViewContact(uri)` | `ViewContact(text uri) -> void` | **Unsupported.** On device this launches the system contacts viewer activity for a `content://` URI. A browser cannot open the Android contacts app, and the simulated `ContactUri` is always empty so there is nothing meaningful to view. Route to `h.Unsupported("method", componentName+".ViewContact")` (shows the orange warning). |

## Implementation Plan

### simulation-capabilities.js

1. Add `'PhoneNumberPicker'` to `SIMULATION_SUPPORTED_TYPES`. (Do **not** add to `SIMULATION_NONVISIBLE_TYPES` — it is visible.)

2. Add a defaults block in `buildSimulationDefaults()`. Structurally the `ListPicker`/`Button` ButtonBase props plus the read-only outputs initialized empty:

```js
PhoneNumberPicker: {
  ...COMMON_VISIBLE_PROPS,
  ...BUTTON_STYLE_PROPS,
  Text: '',
  BackgroundColor: '',
  TextColor: '&HFF000000',
  PhoneNumber: '',
  ContactName: '',
  EmailAddress: '',
  ContactUri: '',
  Picture: '',
  PhoneNumberList: [],
  EmailAddressList: [],
},
```
(`BUTTON_STYLE_PROPS` already supplies `FontSize/FontBold/FontItalic/FontTypeface/TextColor/TextAlignment/Image/Shape/ShowFeedback`. The explicit `BackgroundColor: ''` / `TextColor` mirror the COLOR_DEFAULT convention.)

3. `SIMULATION_VISUAL_PROPS`: add the four new **scalar output** names so they survive the strip-before-render filter and are mirrored into component state — `'PhoneNumber'`, `'ContactName'`, `'EmailAddress'`, `'ContactUri'`. (`Picture`, `Image`, `Shape`, `ShowFeedback`, `TextAlignment`, all `Font*`, `Text`, colors, sizing are already in the set.) The two **list** outputs (`PhoneNumberList`, `EmailAddressList`) are not needed for *rendering*; they are pure block-readable state living in the Go host — do not add them to visual props (they would just be passed through untouched anyway, but they have no render role). If you want them readable on the JS side too, add them, but the host is the source of truth for block reads.

4. `isBooleanProp` / `isNumericProp`: no additions — `Enabled`/`Visible`/`ShowFeedback`/`FontBold`/`FontItalic` are already booleans; `Shape`/`TextAlignment`/`FontSize` already numeric. The output props are text/list, not boolean/numeric.

5. `coerceSimulationValue`: no additions — `PhoneNumber`/`ContactName`/`EmailAddress`/`ContactUri`/`Picture` are text passthrough (default branch). `PhoneNumberList`/`EmailAddressList` are arrays and hit the early `Array.isArray` branch (returned as-is since they are not `Elements`). No `deriveStateFromDesignerProps` derived state needed.

### SimulationComponent.svelte

Add a render branch after the `ListPicker` branch. It is the `ListPicker` button + dropdown menu, but the menu is populated from a fabricated contacts roster instead of `Elements`:

```svelte
{:else if node.type === 'PhoneNumberPicker'}
  <div bind:this={pickerWrapEl} class="sim-picker" class:sim-unsupported={unsupportedHere} style={containerStyle()} data-sim-component={node.name}>
    <button
      bind:this={buttonEl}
      type="button"
      class="sim-button"
      class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
      style={buttonInnerStyle('width: 100%;')}
      disabled={!enabled}
      on:pointerdown={() => pointerDown(false)}
      on:pointerup={pointerUp}
      on:pointercancel={clearLongClick}
      on:focus={focusEvent}
      on:blur={blurEvent}
      on:click={openPhonePicker}
    >{props.Text ?? ''}</button>
    {#if pickerOpen}
      <div class="sim-picker-menu" style={pickerMenuStyle()}>
        <div class="sim-picker-title">Choose a phone number</div>
        {#each FAKE_CONTACTS as contact (contact.uri)}
          {#each contact.numbers as number}
            <button type="button" on:click={() => pickPhoneNumber(contact, number)}>
              {number} — {contact.name}
            </button>
          {/each}
        {/each}
      </div>
    {/if}
  </div>
```

New script glue (reuse existing helpers — `emitEvent`, `emitInteraction`, `pointerDown`, `pointerUp`, `clearLongClick`, `consumeLongClick`, `focusEvent`, `blurEvent`, `buttonInnerStyle`, `containerStyle`, `pickerMenuStyle`, `pickerWrapEl`, the `pickerOpen` flag, and the existing `.sim-picker`/`.sim-picker-menu`/`.sim-picker-title` CSS used by Spinner/ListPicker):

```js
const FAKE_CONTACTS = [
  { name: 'Alex Rivera',  uri: 'sim-contact-1', numbers: ['+1 555-0142', '+1 555-0188'], emails: ['alex.rivera@example.com'] },
  { name: 'Sam Chen',     uri: 'sim-contact-2', numbers: ['+1 555-0177'],                emails: ['sam.chen@example.com'] },
  { name: 'Jordan Patel', uri: 'sim-contact-3', numbers: ['+1 555-0123', '+1 555-0199'], emails: [] },
];

async function openPhonePicker() {
  if (!enabled || consumeLongClick()) return;
  await emitEvent(node.name, 'BeforePicking');
  await tick();
  if (!enabled || !visible) return;
  pickerOpen = true;
}

function pickPhoneNumber(contact, number) {
  pickerOpen = false;
  emitInteraction(
    [
      { component: node.name, property: 'PhoneNumber',      value: number },
      { component: node.name, property: 'ContactName',      value: contact.name },
      { component: node.name, property: 'EmailAddress',     value: contact.emails[0] ?? '' },
      { component: node.name, property: 'PhoneNumberList',  value: contact.numbers },
      { component: node.name, property: 'EmailAddressList', value: contact.emails },
      { component: node.name, property: 'ContactUri',       value: '' },
      { component: node.name, property: 'Picture',          value: '' },
    ],
    { component: node.name, event: 'AfterPicking', args: [] },
  );
}
```

Add `PhoneNumberPicker` to the existing `handleComponentActions()` `'open'` branch so a programmatic `Open()` effect opens the menu (reuse the Spinner/ListPicker condition):

```js
if ((node?.type === 'ListPicker' || node?.type === 'Spinner' || node?.type === 'PhoneNumberPicker') && enabled && visible) {
  pickerFilter = '';
  pickerOpen = true;
}
```

CSS: reuse `.sim-button`, `.sim-picker`, `.sim-no-feedback`, `.sim-picker-menu`, `.sim-picker-title` — no new classes needed. (Use the menu close-on-outside-click logic already wired for `pickerWrapEl`.)

### SimulateOverlay.svelte

**None required.** The `Open()` -> `component-action` / `action:"open"` effect is already relayed by the overlay's effect-application loop (it increments the action counter `handleComponentActions` reads), exactly as for ListPicker/Spinner. No new dialog, toast, or effect type is needed — the fabricated picker is an in-component dropdown, not an overlay-hosted modal.

### simulation_wasm.go

Add a `PhoneNumberPicker` case to `CallMethod` mirroring `ListPicker.Open()`, plus an explicit `ViewContact` unsupported:

```go
case "PhoneNumberPicker":
    switch method {
    case "Open":
        if h.runEvent != nil {
            h.runEvent(componentName, componentType, "BeforePicking", nil)
        }
        h.effects = append(h.effects, componentAction(componentName, "open"))
        return runtime.VoidVal()
    case "ViewContact":
        h.Unsupported("method", componentName+".ViewContact")
        return runtime.VoidVal()
    }
```

- **GetProperty:** generic store path suffices. `PhoneNumber`/`ContactName`/`EmailAddress`/`ContactUri`/`Picture`/`PhoneNumberList`/`EmailAddressList` are read from `h.state[componentName]` after the frontend patches them via `emitInteraction` -> `runEvent`/patch. No computed override needed (unlike DatePicker `Instant`). Defaults are empty until a pick happens, matching AI (empty string / empty list before first pick).
- **SetProperty:** generic path suffices. These outputs are read-only in blocks; the frontend patches them through the normal `setProperty` route. No `Selection`/`Elements`/`Spinner`-style special handling applies (PhoneNumberPicker has no `Elements`/`Selection` surface).
- **Events:** `AfterPicking`/`GotFocus`/`LostFocus`/`TouchDown`/`TouchUp` dispatched from the frontend; only `BeforePicking` is fired host-side on the programmatic `Open()` path above.
- WASM rebuild required: `npm run build:wasm` (GOOS=js GOARCH=wasm); refresh `public/falcon.wasm` and `web/falcon.wasm`.

### design-schema-tree.js

**Already handled.** `canContainDesignComponent()` falls through to the generic visible-component path; PhoneNumberPicker is neither a Canvas/Map/Chart/FeatureCollection child nor a non-visible type. Once `'PhoneNumberPicker'` is in `SIMULATION_SUPPORTED_TYPES`, `unsupportedSimulationComponents()` stops flagging it and `designTreeToInitialState()` merges `SIMULATION_DEFAULTS.PhoneNumberPicker` automatically. No changes to this file.

## Dependencies & Ordering

- **External libraries:** none. The fabricated picker reuses the existing `.sim-picker-menu` dropdown machinery; no Contact Picker API, no native bridge.
- **Prerequisite components:** none strictly required. Directly reuses the ButtonBase rendering and the Spinner/ListPicker dropdown + `handleComponentActions`/`Open()` effect pattern already shipped — they serve as the reference, not a build-order dependency.
- The base **ContactPicker** and **EmailPicker** siblings share the identical pattern (ContactPicker drops the per-number list and tapping selects a whole contact; EmailPicker shows emails). Implementing PhoneNumberPicker first establishes the fabricated-roster approach those can copy.
- **WASM rebuild:** the Go `CallMethod` change requires `npm run build:wasm` and refreshing both `falcon.wasm` copies.

## Web-Platform Limitations & Fidelity Caveats

- **Contacts are fabricated, not real.** A browser has no access to the device contact list, and the only web alternative (the experimental Contact Picker API) is Chromium-Android-only and absent in the desktop browser the simulator runs in. The picker shows a small built-in roster of fake numbers so apps remain demonstrable; values will not match any real device contact.
- **`ContactUri` is always `""`.** There is no `content://com.android.contacts/...` analogue in a browser. Blocks that pass it to `ViewContact` or store/parse it get an empty string.
- **`Picture` is always `""`.** No contact-photo file exists; wiring `Image1.Picture = PhoneNumberPicker1.Picture` shows nothing (it would not resolve through `resolveAssetUrl()` anyway).
- **`ViewContact()` is unsupported** and raises the orange warning — a browser cannot open the system contacts viewer.
- **`BeforePicking` cannot cancel the pick.** Like ListPicker, the simulator fires it and always opens the menu; AI's fallible-event cancel semantics are not honored.
- **`PhoneNumberList`/`EmailAddressList` shapes** are simulator-defined arrays from the fake roster, not device-sourced lists; their length/contents will not match a real contact.
- **COLOR_DEFAULT appearance.** As with Button/ListPicker, the baseline `sim-button` CSS does not reproduce the Android theme's default button color when `BackgroundColor` is `&H00000000`.
- **No Android-version gating.** The helpString warns the list may be empty on pre-3.0 Android; the simulator does not model device/OS variance.

## Effort Estimate

**S** — Reuses the ListPicker button + `.sim-picker-menu` dropdown + `Open()` effect pattern wholesale. Net new work is one defaults block, four new visual-prop names, one render branch (~25 lines) with a fabricated roster constant, a small `pickPhoneNumber()` handler, one line added to `handleComponentActions`, and a ~10-line Go `CallMethod` case (Open + ViewContact-unsupported). The only nuance is honestly fabricating the contact data and documenting the `ContactUri`/`Picture` gaps.
