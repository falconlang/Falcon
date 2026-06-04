# ContactPicker Simulator Implementation Plan

## Overview

ContactPicker is a **SOCIAL**-category, **visible** component in App Inventor. Per the spec helpString: "A button that, when clicked on, displays a list of the contacts to choose among. After the user has made a selection, the following properties will be set to information about the chosen contact: `ContactName`, `EmailAddress`, `ContactUri`, `EmailAddressList`, `PhoneNumber`, `PhoneNumberList`, `Picture`." Other properties affect the appearance of the button (`TextAlignment`, `BackgroundColor`, etc.) and whether it can be clicked (`Enabled`).

Architecturally it is a `Picker → ButtonBase` subclass — the same lineage as ListPicker and ImagePicker. It renders exactly like a Button (same `BackgroundColor`, `Font*`, `Image`, `Shape`, `ShowFeedback`, `Text`, `TextAlignment`, `TextColor` surface) and adds seven read-only output block properties that are filled after a pick: `ContactName` (text), `EmailAddress` (text), `ContactUri` (text), `EmailAddressList` (list), `PhoneNumber` (text), `PhoneNumberList` (list), `Picture` (text URI). When tapped it fires `BeforePicking`, launches the device contacts activity, and on return fires `AfterPicking` after filling those properties.

**Container relationship:** standalone visible button-style picker (no children, no special parent). It is placed in any standard arrangement/Screen exactly like Button/ListPicker/ImagePicker — there are no Canvas/Map/Chart/FeatureCollection containment rules to encode.

## Feasibility Verdict

**Partially feasible.**

The **visual button** surface is fully feasible — it is byte-for-byte the ButtonBase rendering already implemented for Button/ListPicker/DatePicker, so all designer properties render with full fidelity.

The **picking flow** is partially feasible, and the limiting factor is different from ImagePicker. ImagePicker can use a *real* native `<input type="file" accept="image/*">`, which on mobile surfaces the actual gallery. **There is no equivalent for contacts.** The web platform has nothing that reliably reads device contacts:

- The **Contact Picker API** (`navigator.contacts.select`) is non-standard, Chromium-only, restricted to **Android Chrome over HTTPS in a top-level secure context with a user gesture**, and is absent in desktop Chrome, all of Firefox, and all of Safari. It is not a dependable cross-browser primitive and would yield real PII through a permission prompt — inappropriate for a design-time IDE preview.
- There is no file-chooser fallback that returns contact data the way ImagePicker's chooser returns a real file.

Therefore the realistic approximation is a **simulator-provided fake contact roster**: tapping the button fires `BeforePicking`, then shows an in-app dropdown/menu listing a small set of seeded mock contacts (e.g. "Ada Lovelace", "Grace Hopper", "Alan Turing"). Selecting one fills the result properties from that mock record and fires `AfterPicking`. This reuses the existing `sim-picker-menu` dropdown already built for ListPicker/Spinner — no native API, no permission, no real PII, fully client-side and genuinely interactive.

What **cannot** be faithfully simulated:
- **Real device contacts.** The roster is fabricated; it never reflects the host machine's address book (and should not, for privacy).
- **`Picture` as a usable image URI.** On device this is a content URI that `Image.Picture`/`ImageSprite.Picture` can load. The mock `Picture` is at best a design-time asset name or a generated placeholder; wiring `set Image1.Picture to ContactPicker1.Picture` will only display something if the mock points at an existing project asset (otherwise it shows the placeholder string). `resolveAssetUrl()` only matches design-time assets.
- **`ContactUri` shape.** On device this is an Android `content://com.android.contacts/...` URI; the mock value is a synthesized stand-in, so `ViewContact(uri)` and any block that parses the URI will diverge.
- **Android-version conditionals** in the helpString (phone numbers / multiple emails only on later Android versions) are not modeled — the mock simply always provides them.

Because the button + BeforePicking/AfterPicking + an interactive selection all work (it is a user-initiated action, not a fake sensor), this is **partially feasible**, not not-feasible. Only the *source and shape* of the contact data diverges.

## Properties

All designer properties are ButtonBase props already implemented for Button/ListPicker, so the visual side is "reuse Button rendering." The result block properties are read-only outputs stored in host state and filled on pick.

| Property | AI default | Visual/Behavioral | How to simulate | Priority |
|---|---|---|---|---|
| `Text` | `""` | Visual | Button label; reuse `props.Text` render path. | High |
| `BackgroundColor` | `&H00000000` (COLOR_DEFAULT) | Visual | `buttonInnerStyle()` via `colorValue()`. | High |
| `TextColor` | `&H00000000` (COLOR_DEFAULT) | Visual | `buttonInnerStyle()` color. | High |
| `Enabled` | `True` | Behavioral | `disabled` attr; gates the pick handler. | High |
| `Visible` | `True` | Behavioral | Gates whole render (existing). | High |
| `Image` | `""` | Visual | Button background image via `resolveAssetUrl(assets, props.Image)`. | Medium |
| `Shape` | `0` | Visual | Shape styling (default/rounded/rect/oval) — already implemented for Button. | Medium |
| `ShowFeedback` | `True` | Visual | `sim-no-feedback` class toggling press feedback — already implemented. | Medium |
| `FontSize` | `14.0` | Visual | typography in `buttonInnerStyle()`. | Medium |
| `FontBold` | `False` | Visual | `font-weight`. | Medium |
| `FontItalic` | `False` | Visual | `font-style`. | Medium |
| `FontTypeface` | `0` | Visual | typeface mapping. | Low |
| `TextAlignment` | `1` (center) | Visual | text-align style. | Medium |
| `Width` / `Height` | `""` (auto) | Visual | `sizeStyle()`/`containerStyle()` (`-1` auto). | High |
| `Left` / `Top` | `""` | Visual | position (AbsoluteArrangement only). | Low |
| `ContactName` | n/a (read-only output) | Behavioral output | Set from chosen mock record on pick; stored in host state. | High |
| `EmailAddress` | n/a (read-only output) | Behavioral output | Mock primary email. | High |
| `EmailAddressList` | n/a (read-only output, list) | Behavioral output | Mock list of emails; stored as a list value. | Medium |
| `PhoneNumber` | n/a (read-only output) | Behavioral output | Mock primary phone. | Medium |
| `PhoneNumberList` | n/a (read-only output, list) | Behavioral output | Mock list of phones. | Medium |
| `ContactUri` | n/a (read-only output) | Behavioral output | Synthesized stand-in URI. **Not a real device content URI.** | Medium |
| `Picture` | n/a (read-only output) | Behavioral output | Mock asset name or placeholder. **Not a real device photo URI.** | Low |

Note: `Column`/`Row`/`HeightPercent`/`WidthPercent`/`FontTypeface` are invisible/write-only layout helpers handled by the generic store path; no special work. The result properties have no designer default (they only exist as block getters), so they are initialized to empty (`""` / empty list).

## Events

From spec `events` (all params empty).

| Event | Args | How/when the simulator fires it |
|---|---|---|
| `BeforePicking` | none | Fired on button click in `openContactPicker()` *before* the mock contact menu opens, and by the Go host when `Open()` is called programmatically. Mirrors `openListPicker()`. |
| `AfterPicking` | none | Fired after the user selects a contact from the mock dropdown and all result properties have been set. Emitted via `emitInteraction([...property patches], {event:'AfterPicking'})`. If the user dismisses the menu without choosing, no event fires (matches AI `RESULT_CANCELED`). |
| `GotFocus` | none | Wire `on:focus` on the button → `focusEvent` (existing helper). |
| `LostFocus` | none | Wire `on:blur` → `blurEvent`. |
| `TouchDown` | none | `on:pointerdown` → `pointerDown(false)` (existing). |
| `TouchUp` | none | `on:pointerup` → `pointerUp` (existing). |

None of these are blocked by the browser — all are user-initiated, no sensor/permission gate. (One fidelity gap: AI's `BeforePicking` is a fallible event that can cancel opening; like the existing ListPicker we will not honor the cancel return — documented as a known limitation.)

## Methods

From spec `methods`.

| Method | Signature | Simulated behavior (or Unsupported + why) |
|---|---|---|
| `Open()` | `Open() → void` | Opens the picker as though tapped. Go host fires `BeforePicking` then emits a `component-action` effect with `action: "open"`; the overlay relays it; the component's `handleComponentActions()` opens the mock contact menu. Mirrors the existing `ListPicker.Open()` path exactly. |
| `ViewContact(uri)` | `ViewContact(text uri) → void` | **Unsupported.** On device this launches the system contacts viewer activity for a `content://` URI. A browser cannot open a device contact viewer, and the mock URIs are synthetic. Call `h.Unsupported("method", componentName+".ViewContact")` (orange warning). Optionally, as a soft approximation, surface a `notice`/toast effect echoing the URI, but the real behavior cannot be reproduced. |

## Implementation Plan

### simulation-capabilities.js

1. Add `'ContactPicker'` to `SIMULATION_SUPPORTED_TYPES`. **Do not** add to `SIMULATION_NONVISIBLE_TYPES` — it is visible.

2. Add a defaults block in `buildSimulationDefaults()`. It is structurally the `ListPicker`/`Button` ButtonBase block plus the read-only result properties initialized empty:

```js
ContactPicker: {
  ...COMMON_VISIBLE_PROPS,
  ...BUTTON_STYLE_PROPS,
  Text: '',
  BackgroundColor: '',
  TextColor: '&HFF000000',
  ContactName: '',
  EmailAddress: '',
  EmailAddressList: [],
  PhoneNumber: '',
  PhoneNumberList: [],
  ContactUri: '',
  Picture: '',
},
```
(`BUTTON_STYLE_PROPS` already supplies `FontSize/FontBold/FontItalic/FontTypeface/TextColor/TextAlignment/Image/Shape/ShowFeedback`.)

3. `SIMULATION_VISUAL_PROPS`: add the result-property names so they are not stripped before reaching the renderer (the renderer does not paint them, but they must survive into component state so the live `props` object carries them and the Go host can read/write them). Add:
   `'ContactName'`, `'EmailAddress'`, `'EmailAddressList'`, `'PhoneNumber'`, `'PhoneNumberList'`, `'ContactUri'`.
   (`Picture` is already in the set; `Image`, `Shape`, `ShowFeedback`, `TextAlignment`, all `Font*`, `Text`, colors, sizing are already present.)

4. `isBooleanProp` / `isNumericProp`: no additions — `Enabled`/`Visible`/`ShowFeedback`/`FontBold`/`FontItalic` are already booleans; `Shape`/`TextAlignment`/`FontSize` are already numeric. The result properties are text/list, not boolean/numeric.

5. `coerceSimulationValue`: `ContactName`/`EmailAddress`/`PhoneNumber`/`ContactUri`/`Picture` are text passthroughs (default branch). `EmailAddressList`/`PhoneNumberList` are lists — they will hit the `Array.isArray(value)` early-return (which only normalizes for `Elements`, otherwise returns the array as-is), which is the desired behavior. No new coercion branch needed. No `deriveStateFromDesignerProps` derived state needed.

### SimulationComponent.svelte

Add a render branch after the `ListPicker` branch. It is the `ListPicker` button + the existing `sim-picker-menu` dropdown, but populated from a fixed mock roster instead of `Elements`:

```svelte
{:else if node.type === 'ContactPicker'}
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
      on:click={openContactPicker}
    >{props.Text ?? ''}</button>
    {#if pickerOpen}
      <div class="sim-picker-menu" style={pickerMenuStyle()}>
        <div class="sim-picker-title">Choose a contact</div>
        {#each MOCK_CONTACTS as contact, i}
          <button type="button" on:click={() => pickContact(i)}>{contact.ContactName}</button>
        {/each}
      </div>
    {/if}
  </div>
```

New script glue (reuse existing helpers — `emitEvent`, `emitInteraction`, `pointerDown`, `pointerUp`, `clearLongClick`, `consumeLongClick`, `focusEvent`, `blurEvent`, `buttonInnerStyle`, `containerStyle`, `pickerMenuStyle`, `pickerWrapEl`, `pickerOpen`; the outside-click close behavior that already closes `pickerOpen` for ListPicker/Spinner applies automatically):

```js
const MOCK_CONTACTS = [
  {
    ContactName: 'Ada Lovelace', EmailAddress: 'ada@example.com',
    EmailAddressList: ['ada@example.com'], PhoneNumber: '+1-555-0100',
    PhoneNumberList: ['+1-555-0100'], ContactUri: 'sim://contacts/1', Picture: '',
  },
  {
    ContactName: 'Grace Hopper', EmailAddress: 'grace@example.com',
    EmailAddressList: ['grace@example.com', 'g.hopper@navy.example'], PhoneNumber: '+1-555-0142',
    PhoneNumberList: ['+1-555-0142', '+1-555-0199'], ContactUri: 'sim://contacts/2', Picture: '',
  },
  {
    ContactName: 'Alan Turing', EmailAddress: 'alan@example.com',
    EmailAddressList: ['alan@example.com'], PhoneNumber: '+44-20-5550-0000',
    PhoneNumberList: ['+44-20-5550-0000'], ContactUri: 'sim://contacts/3', Picture: '',
  },
];

async function openContactPicker() {
  if (!enabled || consumeLongClick()) return;
  await emitEvent(node.name, 'BeforePicking');
  await tick();
  if (!enabled || !visible) return;
  pickerOpen = true;
}

function pickContact(index) {
  const c = MOCK_CONTACTS[index];
  if (!c) return;
  pickerOpen = false;
  emitInteraction(
    [
      { component: node.name, property: 'ContactName', value: c.ContactName },
      { component: node.name, property: 'EmailAddress', value: c.EmailAddress },
      { component: node.name, property: 'EmailAddressList', value: c.EmailAddressList },
      { component: node.name, property: 'PhoneNumber', value: c.PhoneNumber },
      { component: node.name, property: 'PhoneNumberList', value: c.PhoneNumberList },
      { component: node.name, property: 'ContactUri', value: c.ContactUri },
      { component: node.name, property: 'Picture', value: c.Picture },
    ],
    { component: node.name, event: 'AfterPicking', args: [] },
  );
}
```

Add `ContactPicker` to the `handleComponentActions()` `'open'` branch so the programmatic `Open()` effect opens the menu — reuse the existing ListPicker/Spinner condition:

```js
if ((node?.type === 'ListPicker' || node?.type === 'Spinner' || node?.type === 'ContactPicker') && enabled && visible) {
  pickerFilter = '';
  pickerOpen = true;
}
```

CSS: reuse `.sim-button`, `.sim-picker`, `.sim-picker-menu`, `.sim-picker-title`, `.sim-no-feedback` — no new classes needed.

### SimulateOverlay.svelte

**None required.** The `Open()` → `component-action` / `action:"open"` effect is already relayed by `applyEffects()` (it increments the action counter `handleComponentActions` reads). The contact menu is an in-app dropdown rendered by `SimulationComponent`, not an overlay dialog, so no new dialog/notice/effect type is needed. (Only if `ViewContact` is given the optional soft-toast approximation would a `notice` effect be emitted — and `notice` is already handled by the overlay.)

### simulation_wasm.go

Add a `ContactPicker` case to `CallMethod` mirroring `ListPicker.Open()`, plus an explicit `ViewContact` → unsupported:

```go
case "ContactPicker":
    switch method {
    case "Open":
        if h.runEvent != nil {
            h.runEvent(componentName, componentType, "BeforePicking", nil)
        }
        h.effects = append(h.effects, componentAction(componentName, "open"))
        return runtime.VoidVal()
    case "ViewContact":
        h.Unsupported("method", componentName+".ViewContact is not supported in the web simulator")
        return runtime.VoidVal()
    }
```

- **GetProperty:** generic store path suffices. `ContactName`/`EmailAddress`/`ContactUri`/`Picture`/`PhoneNumber` read from `h.state[componentName]` like any stored prop; `EmailAddressList`/`PhoneNumberList` are stored as list values when the frontend patches them via `emitInteraction`. No computed override needed (unlike DatePicker `Instant`).
- **SetProperty:** generic path suffices for the result properties (they are read-only in blocks; the frontend writes them through the standard property-patch path which lands in `h.setProperty`). The ButtonBase visual props (`Text`, `BackgroundColor`, `Enabled`, etc.) also use the generic `h.setProperty`. Note: `Selection`/`Elements`/`SelectionIndex` special-casing in `SetProperty` does not apply — ContactPicker has none of those, so it never hits `setSelection`/`setElements`.
- **Events:** `AfterPicking`/`GotFocus`/`LostFocus`/`TouchDown`/`TouchUp` are dispatched from the frontend via `emitEvent`/`emitInteraction` → `runEvent`; only `BeforePicking` on the programmatic `Open()` path is host-fired (above).

### design-schema-tree.js

**Already handled.** `canContainDesignComponent()` falls through to the generic visible-component containment for any standard component, and ContactPicker is neither a Canvas/Map/Chart/FeatureCollection child nor a non-visible type. Once `'ContactPicker'` is in `SIMULATION_SUPPORTED_TYPES`, `unsupportedSimulationComponents()` stops flagging it and `designTreeToInitialState()` merges `SIMULATION_DEFAULTS.ContactPicker` automatically. No changes to this file.

## Dependencies & Ordering

- **External libraries:** none. Uses the existing `sim-picker-menu` dropdown, style/event helpers, and a hard-coded mock roster. No Contact Picker API, no permissions.
- **Prerequisite components:** none strictly required. It directly reuses the ListPicker button-plus-dropdown pattern and the ImagePicker `Open()`/`BeforePicking`/`AfterPicking` flow, both already shipped — they serve as the reference, not a build-order dependency.
- **WASM rebuild:** the Go `CallMethod` change requires `npm run build:wasm` (GOOS=js GOARCH=wasm) and refreshing `public/falcon.wasm` / `web/falcon.wasm`.

## Web-Platform Limitations & Fidelity Caveats

- **No real device contacts.** The roster is a fixed mock list (Ada Lovelace, Grace Hopper, Alan Turing). It never reflects any real address book — by design, for privacy. The only cross-browser contact API (Contact Picker API) is Chromium-Android-only, gated behind a permission prompt and a secure top-level context, and is deliberately not used.
- **`ContactUri` is synthetic** (`sim://contacts/N`), not an Android `content://com.android.contacts/...` URI. Blocks that parse the URI, or pass it to `ViewContact`, will diverge.
- **`Picture` is not a loadable device photo URI.** It is empty/placeholder unless pointed at an existing project asset; `set Image1.Picture to ContactPicker1.Picture` will show the placeholder string by default (`resolveAssetUrl()` only matches design-time assets).
- **`ViewContact(uri)` is unsupported** — no browser can open a device contact-viewer activity; it triggers an orange "unsupported" warning.
- **Android-version conditionals not modeled.** The helpString notes phone numbers and multi-email only exist on later Android versions; the mock always supplies them.
- **`BeforePicking` cannot cancel the pick.** Like the existing ListPicker, the simulator fires `BeforePicking` and always proceeds to open the menu; AI's fallible-event cancel semantics are not honored.
- **COLOR_DEFAULT appearance.** Like Button/ListPicker, the baseline `sim-button` CSS does not reproduce the Android theme's default button color when `BackgroundColor` is `&H00000000`; it uses the simulator's button styling.
- **The chooser is an in-app dropdown, not the native contacts activity.** Look-and-feel is the simulator's `sim-picker-menu`, not the Android contacts screen.

## Effort Estimate

**S** — Reuses the ListPicker button + `sim-picker-menu` dropdown wholesale and the ImagePicker `Open()` flow. Net new work is one defaults block, six new names in `SIMULATION_VISUAL_PROPS`, one render branch (~25 lines) plus a mock roster and a `pickContact` handler, one line added to the `handleComponentActions` open condition, and a ~12-line Go `CallMethod` case (`Open` + unsupported `ViewContact`). The only nuance is honestly seeding mock contact data and documenting the data-source/URI fidelity gaps.
