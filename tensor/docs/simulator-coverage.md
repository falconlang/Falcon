# Tensor Simulator — Component Feature Coverage

**Generated:** 2026-06-04  
**Components:** 44 simulated (Screen/Form counted as one)  
**Purpose:** Precise record of what is implemented, what is partial, and exactly why each gap exists. Use this as the authority when deciding what to build next.

---

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully implemented and functional |
| ⚠️ | Accepted in state / sent as effect, but behaviour is approximated or incomplete |
| ❌ | Not implemented — falls to `h.Unsupported()` or silently no-ops |

---

## Cross-Cutting Gaps (affect every component)

These gaps exist across all simulated components and should be fixed once rather than per-component:

### 1. WidthPercent / HeightPercent (❌ on every component)
App Inventor converts `WidthPercent=50` → `Width = -(1000+50) = -1050`. The Svelte `sizeStyle()` already handles that sentinel (`value <= -1000 → ${-value-1000}%`). The gap is that the Go `SetProperty` host never intercepts `WidthPercent`/`HeightPercent` writes to perform this conversion — it just stores them as raw numbers. Fix: add a Go `SetProperty` case for these two properties on every component type that converts to the Width/Height sentinel.

### 2. Canvas block-driven drawing (⚠️ effects dispatched, never applied)
`DrawLine`, `DrawCircle`, etc. in the Go host emit `component-action` effects with `action='canvas-draw'`. `SimulateOverlay.applyEffects` routes these into `actionTokens`, but `SimulationComponent.handleComponentActions` has no case for `'canvas-draw'`. The ops are lost. The canvas only re-renders sprites and background when reactive state changes. Fix: add a `runAction(actionState, 'canvas-draw', ...)` handler in `handleComponentActions` that calls `ctx.drawLine` / `ctx.arc` etc. from the effect payload fields, and handle sequential ops (the current token counter deduplication must be bypassed for sequential draw calls — consider using an incremental op-log in state instead).

### 3. Sprite touch/interaction events (❌ Ball, ImageSprite)
Canvas pointer events dispatch only to the Canvas component. Individual sprites receive no `TouchDown`, `TouchUp`, `Touched`, `Dragged`, or `Flung` events because there is no per-sprite hit-test on pointer events. Fix: in `canvasPointerDown/Move/Up` iterate sprites, compute bounding boxes at their stored X/Y/Radius/Width/Height, and dispatch events to matching sprites.

### 4. Animation loop (❌ Ball, ImageSprite, Canvas physics)
The simulator is purely event-driven. `Speed`, `Interval`, `Heading`, `Bounce`, `EdgeReached`, `CollidedWith`, and `NoLongerCollidingWith` all require a `requestAnimationFrame` loop running at each sprite's `Interval` ms. No such loop exists. Fix: when any Ball or ImageSprite with `Speed > 0` is present, start a shared `setInterval`/`rAF` loop in `SimulationComponent` that moves sprites, calls `MoveIntoBounds`, detects edge collisions, detects inter-sprite AABB overlaps, and dispatches the appropriate Go events.

### 5. Map feature direct events (⚠️ design-level semantic gap)
Leaflet routes click/context-menu events to map layers. The current implementation fires `FeatureClick`/`FeatureLongClick` on the **Map component**, not on the clicked feature component itself. Falcon blocks written as `when Marker1.Click do...` never fire; you must use `when Map1.FeatureClick do...`. This matches the Leaflet model but diverges from App Inventor, where feature components dispatch their own events. Fix: in `createMapLayer`, resolve which component name owns each Leaflet layer and call `emitEvent(featureComponentName, 'Click')` instead of routing through the parent map.

---

## Summary Table

| Component | Category | Props ✅/total | Events ✅/total | Methods ✅/total |
|-----------|----------|---------------|----------------|-----------------|
| Screen / Form | UI | 12/20 | 1/5 | 2/8 |
| Label | UI | 12/13 | 0/0 | 0/0 |
| TextBox | UI | 15/17 | 3/3 | 5/5 |
| PasswordTextBox | UI | 14/16 | 3/3 | 4/4 |
| Button | UI | 14/16 | 6/6 | 0/0 |
| CheckBox | UI | 11/13 | 3/3 | 0/0 |
| Switch | UI | 15/17 | 3/3 | 0/0 |
| Slider | UI | 12/14 | 3/3 | 0/0 |
| Image | UI | 11/14 | 1/1 | 0/0 |
| Spinner | UI | 17/20 | 1/3 | 1/1 |
| ListPicker | UI | 22/24 | 6/6 | 1/1 |
| DatePicker | UI | 19/21 | 5/5 | 3/3 |
| TimePicker | UI | 17/19 | 5/5 | 3/3 |
| ListView | UI | 25/31 | 1/1 | 9/9 |
| CircularProgress | UI | 6/8 | 0/0 | 0/0 |
| LinearProgress | UI | 9/11 | 1/1 | 1/1 |
| WebViewer | UI | 8/14 | 2/4 | 7/12 |
| VideoPlayer | Media | 8/10 | 2/2 | 5/5 |
| HorizontalArrangement | Layout | 9/11 | 0/0 | 0/0 |
| VerticalArrangement | Layout | 9/11 | 0/0 | 0/0 |
| HorizontalScrollArrangement | Layout | 9/11 | 0/0 | 0/0 |
| VerticalScrollArrangement | Layout | 9/11 | 0/0 | 0/0 |
| AbsoluteArrangement | Layout | 7/9 | 0/0 | 0/0 |
| TableArrangement | Layout | 5/9 | 0/0 | 0/0 |
| Notifier | UI | 2/2 | 4/4 | 10/10 |
| TinyDB | Storage | 1/1 | 0/0 | 6/6 |
| EmailPicker | Social | 13/14 | 3/3 | 4/4 |
| ImagePicker | Media | 14/17 | 6/6 | 1/1 |
| FilePicker | Media | 13/19 | 6/6 | 1/1 |
| ContactPicker | Social | 17/23 | 6/6 | 1/2 |
| PhoneNumberPicker | Social | 17/23 | 6/6 | 1/2 |
| Canvas | Drawing | 11/16 | 5/5 | 8/13 |
| Ball | Drawing | 7/10 | 0/8 | 3/7 |
| ImageSprite | Drawing | 10/15 | 0/8 | 3/7 |
| Chart | Charts | 8/16 | 1/1 | 3/5 |
| ChartData2D | Charts | 3/4 | 0/1 | 10/16 |
| Trendline | Charts | 2/19 | 0/1 | 0/2 |
| Map | Maps | 11/27 | 6/14 | 1/5 |
| Marker | Maps | 14/20 | 0/5† | 3/7 |
| Circle | Maps | 12/14 | 0/5† | 3/5 |
| LineString | Maps | 8/11 | 0/5† | 2/4 |
| Polygon | Maps | 10/15 | 0/5† | 2/5 |
| Rectangle | Maps | 13/15 | 0/5† | 4/7 |
| FeatureCollection | Maps | 2/10 | 0/7 | 0/2 |

† Feature events fire on the parent **Map** component, not on individual feature components (see Cross-Cutting Gap #5).

---

## Component Details

---

### Screen / Form

**Implemented props:** Width (360 px constant), Height (640 px constant), BackgroundColor, BackgroundImage, Title (rendered in action bar when TitleVisible=true), TitleVisible, AlignHorizontal, AlignVertical, Scrollable (non-scrollable screens clip; scrollable screens overflow-y:auto), ShowStatusBar (no-op — no status bar DOM element), ScreenOrientation (no-op — browser controls orientation), Platform (returns "Android"), PlatformVersion (returns "simulator").

| Gap | Type | Why |
|-----|------|-----|
| Sizing, Theme, AboutScreen, HighContrast, BigDefaultText, ActionBar, PrimaryColorDark, AccentColor, OpenScreenAnimation, CloseScreenAnimation | ⚠️ accepted, ignored | No theme engine or animation system in the simulator |
| BackPressed event | ❌ | No browser back-button intercept; would need `popstate` listener |
| ErrorOccurred event | ❌ | No component-level error dispatch infrastructure |
| OtherScreenClosed, ScreenOrientationChanged events | ❌ | Multi-screen and device rotation unsupported |
| Initialize | ✅ fires at session start | — |
| Navigation methods (OpenScreen, CloseScreen, etc.) | ❌ | Multi-screen not supported; emits informative Unsupported notice |
| AskForPermission, HideKeyboard | ⚠️ no-op | No browser permission API bridge; keyboard is virtual |

---

### Label

**Implemented:** Text, BackgroundColor, TextColor, FontSize, FontBold, FontItalic, FontTypeface, TextAlignment, HasMargins (2 px margin), HTMLFormat (sanitized HTML subset rendered via `{@html}`), Visible, Width, Height, Left, Top.

| Gap | Type | Why |
|-----|------|-----|
| HTMLContent (read-only) | ⚠️ returns null | Go GetProperty has no special case; would need to return the HTML-formatted Text from state |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### TextBox

**Implemented:** All visual props (Text, Hint, HintColor, BackgroundColor, TextColor, fonts, TextAlignment, MultiLine, NumbersOnly, ReadOnly, Enabled). Events: GotFocus, LostFocus, TextChanged (fires on user input AND on programmatic Text set). Methods: HideKeyboard (no-op), RequestFocus, MoveCursorTo, MoveCursorToStart, MoveCursorToEnd.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |
| FontBold, FontItalic (missing from DB BlockProperties) | n/a | Not exposed as block properties in AI — no gap |

---

### PasswordTextBox

Same coverage as TextBox with PasswordVisible (toggles `type=password`) added. All 4 methods and all 3 events implemented.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Button

**Implemented:** All appearance props (Text, BackgroundColor, TextColor, fonts, TextAlignment, Image, Shape, ShowFeedback, Enabled). All 6 events: Click, GotFocus, LostFocus, TouchDown, TouchUp, LongClick (600 ms timer).

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### CheckBox

**Implemented:** Checked, Text, BackgroundColor, TextColor, FontSize, FontBold, FontItalic, FontTypeface, Enabled. Events: Changed (user toggle AND programmatic Checked set), GotFocus, LostFocus.

| Gap | Type | Why |
|-----|------|-----|
| TextAlignment | ⚠️ applied but cosmetically limited | `text-align` on a `display:inline-flex` container does not visually align the label text; AI aligns the text in the `CompoundButton` label span |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Switch

Same as CheckBox plus: ThumbColorActive, ThumbColorInactive, TrackColorActive, TrackColorInactive all fully rendered via CSS custom properties. On prop fires Changed programmatically.

| Gap | Type | Why |
|-----|------|-----|
| TextAlignment | ⚠️ same as CheckBox | — |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Slider

**Implemented:** MinValue, MaxValue, ThumbPosition (clamped to [Min,Max]), ThumbEnabled (hidden thumb + pointer-events:none), ColorLeft, ColorRight, ThumbColor, NumberOfSteps, Enabled, Visible. Events: PositionChanged, TouchDown, TouchUp. No methods.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Image

**Implemented:** Picture (resolved from design assets), Visible, Width, Height, Left, Top, Clickable (Click event), RotationAngle, ScalePictureToFit, Scaling (both map to object-fit), AlternateText.

| Gap | Type | Why |
|-----|------|-----|
| Animation | ⚠️ stored, not animated | AI uses Android `AnimationUtils`; no CSS equivalent without mapping string values |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Spinner

**Implemented:** Elements/ElementsFromString, Selection, SelectionIndex (auto-selects first item when populated), Prompt, BackgroundColor, TextColor, FontSize, FontBold, FontItalic, FontTypeface, ShowFeedback, Enabled. Event: AfterSelecting (first-item pick suppressed). Method: DisplayDropdown.

| Gap | Type | Why |
|-----|------|-----|
| TouchDown, TouchUp | ❌ | `<select>` element has no pointer events wired |
| Image | ⚠️ stored | Spinner dropdown button has no backgroundImageStyle applied |
| TextAlignment | ⚠️ limited | `text-align` on `<select>` is inconsistent cross-browser |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### ListPicker

**Fully implemented.** All 24 props, all 6 events (including BeforePicking before opening), method Open. Filter bar, item colors, title all render.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### DatePicker

**Fully implemented.** Day/Month/Year/MonthInText/Instant all derived correctly. AfterDateSet, GotFocus, LostFocus, TouchDown, TouchUp all fire. LaunchPicker, SetDateToDisplay, SetDateToDisplayFromInstant all work.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### TimePicker

**Fully implemented.** Same coverage as DatePicker. Hour/Minute/Instant correct. AfterTimeSet, LaunchPicker, SetTimeToDisplay, SetTimeToDisplayFromInstant all work.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### ListView

**Implemented:** Elements/ElementsFromString/ListData, Selection/SelectionIndex, BackgroundColor, TextColor, TextColorDetail, FontSize, FontSizeDetail, FontTypeface, FontTypefaceDetail, SelectionColor, ElementColor, DividerColor, ShowFilterBar, HintText, ListViewLayout (affects detail+image rendering), Orientation (horizontal carousel), ImageWidth/ImageHeight. All 9 methods. Event: AfterPicking.

| Gap | Type | Why |
|-----|------|-----|
| DividerThickness | ⚠️ color applied, thickness not | Border-bottom width not driven from prop — add `border-bottom-width: ${DividerThickness}px` |
| ElementCornerRadius | ⚠️ stored | `border-radius` not applied to row buttons |
| ElementMarginsWidth | ⚠️ stored | `margin` not applied to row buttons |
| BounceEdgeEffect | ⚠️ stored | No browser overscroll bounce control (`overscroll-behavior` exists but isn't the same) |
| MultiSelect | ⚠️ stored | Multi-select mode not rendered; would need checkbox-per-row + separate SelectionList state |
| SelectionDetailText (read-only) | ⚠️ returns null | Go GetProperty has no special case — would need to look up the selected row's Text2 |
| TextSize | ⚠️ accepted | Duplicate of FontSize from an older API version; could alias to FontSize |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### CircularProgress

**Implemented:** Color (CSS ring tint), Visible, Width, Height, Left, Top. Pure CSS animation. No events or methods.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### LinearProgress

**Implemented:** Indeterminate (CSS sweep animation), IndeterminateColor, ProgressColor, Progress (clamped to [Min,Max]), Minimum, Maximum, Visible, Width. Event: ProgressChanged. Method: IncrementProgressBy.

| Gap | Type | Why |
|-----|------|-----|
| Height | ⚠️ defaults to 6 px | Progress bar height is fixed in CSS; `sizeStyle('Height')` isn't applied to the bar track |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### TableArrangement

**Implemented:** Visible, Width, Height, Left, Top. CSS Grid container (`grid-template-columns: repeat(Columns, 1fr)`). Children fill cells in document order.

| Gap | Type | Why |
|-----|------|-----|
| Columns, Rows | ⚠️ designer-only, not block-settable | These are designer properties only (`properties[]` in DB, not `blockProperties[]`) — no block can change them at runtime |
| Per-child Row/Column placement | ⚠️ document order only | The design schema stores children as a flat ordered list with no cell coordinates; sparse/out-of-order layouts cannot be reproduced |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### WebViewer

**Implemented:** HomeUrl (navigates iframe), CurrentUrl (navigates + stored), PageLoaded event, ErrorOccurred event (best-effort on iframe error). Methods: GoToUrl, GoHome, GoBack (effect sent), GoForward (effect sent), Reload (effect sent), ClearCaches/ClearCookies/ClearLocations (no-op).

| Gap | Type | Why |
|-----|------|-----|
| CurrentPageTitle (read) | ⚠️ always empty | Same-origin policy prevents reading cross-origin iframe's `document.title` |
| FollowLinks | ⚠️ ignored | Browser always follows links inside iframe; `sandbox` attribute can block navigation but breaks many sites |
| IgnoreSslErrors | ⚠️ no-op | Browser enforces SSL; no API to override |
| PromptforPermission | ⚠️ no-op | `sandbox="allow-scripts"` already grants limited permissions; cannot replicate Android runtime permission model |
| WebViewString (r/w) | ⚠️ stored, no bridge | Requires injecting JS into the frame — only works same-origin; cross-origin pages block it |
| WebViewStringChange event | ❌ | Requires same-origin `window.postMessage` bridge; most target URLs are cross-origin |
| BeforePageLoad event | ⚠️ fires on load, not before | `iframe.onload` fires after the page has already loaded; no `onbeforeunload`-equivalent intercept |
| CanGoBack, CanGoForward | ❌ always false | Cross-origin `contentWindow.history` is inaccessible |
| RunJavaScript | ❌ | Marked Unsupported — cross-origin restriction |
| StopLoading | ⚠️ no-op | `contentWindow.stop()` is cross-origin blocked |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### VideoPlayer

**Fully implemented.** Source, Volume, Loop, FullScreen (stored). Events: Completed, VideoPlayerError. Methods: Start, Pause, Stop, SeekTo, GetDuration (returns state-patched duration after `loadedmetadata`).

| Gap | Type | Why |
|-----|------|-----|
| FullScreen trigger | ⚠️ stored, not triggered | Setting `FullScreen=true` via blocks does not call `requestFullscreen()` — would need a Go effect handler + Svelte action |
| .3gp codec | ⚠️ browser-dependent | Some formats may not decode in all browsers; falls to VideoPlayerError |
| Autoplay policy | ⚠️ browser-dependent | Chrome/Safari may mute or block auto-start without a prior user gesture |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### HorizontalArrangement / VerticalArrangement / HorizontalScrollArrangement / VerticalScrollArrangement

**Implemented:** AlignHorizontal, AlignVertical (CSS flexbox justify/align), BackgroundColor, Image (background-image), Visible, Width, Height, Left, Top. Empty-arrangement 100 px floor. Scroll variants overflow correctly.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### AbsoluteArrangement

**Implemented:** BackgroundColor, Image, Visible, Width (defaults -2 = fill parent), Height (defaults 100 px), Left, Top. Children position with `position:absolute` using their Left/Top.

| Gap | Type | Why |
|-----|------|-----|
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Notifier

**Fully implemented.** ShowAlert (auto-dismiss via NotifierLength duration), ShowMessageDialog, ShowChooseDialog (dual-fire ChoosingCanceled+AfterChoosing on cancel), ShowTextDialog, ShowPasswordDialog (dual-fire TextInputCanceled+AfterTextInput on cancel), ShowProgressDialog, DismissProgressDialog. LogError/Warning/Info route to host logs. BackgroundColor/TextColor applied to toasts.

No gaps.

---

### TinyDB

**Fully implemented.** Namespace, StoreValue, GetValue, ClearTag, ClearAll, GetTags, GetEntries. Persisted across sessions via `localStorage` under key `tensor:simulation:tinydb:v1`. Deep-copy on read/write for mutation isolation.

No gaps.

---

### EmailPicker

**Implemented as a TextBox clone** with `type=email`. Text, Hint, HintColor, BackgroundColor, TextColor, FontSize, FontBold, FontItalic, FontTypeface, TextAlignment, Enabled. Events: GotFocus, LostFocus, TextChanged. Methods: RequestFocus, MoveCursorTo, MoveCursorToStart, MoveCursorToEnd.

| Gap | Type | Why |
|-----|------|-----|
| Contacts autocomplete dropdown | ❌ | Browser has no address-book API; `<input type=email>` only auto-suggests from browser history |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### ImagePicker

**Implemented:** Button appearance (Text, BackgroundColor, TextColor, fonts, Image, Shape, ShowFeedback, Enabled). Uses a hidden `<input type=file accept=image/*>` as the picker. Events: BeforePicking, AfterPicking, GotFocus, LostFocus, TouchDown, TouchUp. Method: Open.

| Gap | Type | Why |
|-----|------|-----|
| Selection | ⚠️ file.name, not device URI | `file.name` is returned, not a `content://` URI; a blob: URL is generated internally but not exposed as a block-readable property |
| Picked image in Image component | ⚠️ not auto-linked | The blob URL is stored as `ImagePath` (internal only), not wired to any Image component's `Picture` — app must do this manually with `set Image1.Picture to ImagePicker1.Selection` |
| 10-image storage rotation | ❌ | AI rotates through 10 saved filenames; browser file chooser has no persistent storage |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### FilePicker

**Implemented:** Button appearance, BeforePicking, AfterPicking, GotFocus, LostFocus, TouchDown, TouchUp. Method: Open. MimeType maps to input `accept` attribute (non-binding hint only).

| Gap | Type | Why |
|-----|------|-----|
| Selection | ⚠️ blob URL, not content:// | Browser creates an `ObjectURL`; actual filesystem path is inaccessible |
| Action = Pick Directory | ⚠️ falls back to Pick Existing File | File System Access API (`showDirectoryPicker`) is Chromium-only and not universally available |
| Action = Pick New File | ⚠️ falls back to Pick Existing File | `showSaveFilePicker` is Chromium-only |
| MimeType enforcement | ⚠️ advisory only | `<input accept>` is a hint; OS file dialog may show all files |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### ContactPicker / PhoneNumberPicker

**Implemented:** Button appearance, BeforePicking, AfterPicking, GotFocus, LostFocus, TouchDown, TouchUp, Open. After pick: ContactName, PhoneNumber, EmailAddress set from a seeded mock roster (3 contacts: Alice/Bob/Carol with placeholder numbers and emails). Selection = ContactName.

| Gap | Type | Why |
|-----|------|-----|
| Real device contacts | ❌ | Contact Picker API is Chromium-Android-only, permission-gated, and absent in desktop browsers |
| ContactUri | ❌ always empty | No browser `content://` URI for contacts |
| Picture | ❌ always empty | No browser API to fetch contact photos |
| EmailAddressList, PhoneNumberList | ⚠️ single-item only | Mock contacts have one email/phone each; AI returns all associated entries |
| ViewContact(uri) | ❌ | Marked Unsupported — no browser contact-viewer activity |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Canvas

**Implemented:** Render as `<canvas>` element. BackgroundColor, BackgroundImage, PaintColor (used in draw effects), LineWidth, FontSize, TextAlignment, TapThreshold (tap/drag/fling discrimination), ExtendMovesOutsideCanvas (pointer capture). Pointer events correctly dispatch TouchDown, TouchUp, Touched, Dragged, Flung (with tap/drag/fling geometry). Sprites (Ball, ImageSprite) rendered on canvas from their state. Draw methods send typed `canvas-draw` effects to Go.

| Gap | Type | Why |
|-----|------|-----|
| DrawLine, DrawCircle, DrawArc, DrawPoint, DrawShape, DrawText, DrawTextAtAngle, Clear | ⚠️ effects dispatched, not applied | See cross-cutting gap #2 — `handleComponentActions` has no canvas-draw handler |
| draggedAnySprite, flungSprite args | ⚠️ always false | No per-sprite hit-testing (see cross-cutting gap #3) |
| GetPixelColor, GetBackgroundPixelColor | ⚠️ always return 0 | Canvas cross-origin taint from external images; would need sync transfer across WASM boundary |
| SetBackgroundPixelColor | ❌ | Marked Unsupported — WASM can't access canvas ImageData synchronously |
| BackgroundImageinBase64 | ❌ | Write-only prop; Go SetProperty has no case to decode and apply it |
| Save, SaveAs | ❌ | Marked Unsupported — `canvas.toBlob()` can trigger a download, not write to device storage |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Ball

**Rendered** on the parent Canvas as a filled circle using stored X, Y, Radius, PaintColor, OriginAtCenter.

| Gap | Type | Why |
|-----|------|-----|
| Touched, TouchDown, TouchUp, Dragged, Flung events | ❌ | See cross-cutting gap #3 |
| CollidedWith, NoLongerCollidingWith, EdgeReached events | ❌ | See cross-cutting gap #4 |
| Speed, Interval (autonomous motion) | ❌ | See cross-cutting gap #4 |
| Bounce, MoveIntoBounds, PointTowards | ❌ | Require animation loop + canvas bounds |
| CollidingWith | ⚠️ always returns false | No per-frame AABB overlap test |
| Z (draw order) | ⚠️ ignored | Sprites drawn in children-array order, not by Z value |
| Heading rotation visual | ⚠️ not rendered | Ball is a circle; rotation is unused visually (matters for direction of autonomous motion) |

---

### ImageSprite

**Rendered** on the parent Canvas as an image at X, Y, Width, Height using stored Picture (resolved from assets).

| Gap | Type | Why |
|-----|------|-----|
| Touched, TouchDown, TouchUp, Dragged, Flung events | ❌ | See cross-cutting gap #3 |
| CollidedWith, NoLongerCollidingWith, EdgeReached events | ❌ | See cross-cutting gap #4 |
| Speed, Interval (autonomous motion) | ❌ | See cross-cutting gap #4 |
| Bounce, MoveIntoBounds, PointTowards | ❌ | Require animation loop + canvas bounds |
| CollidingWith | ⚠️ always returns false | No per-frame AABB overlap test |
| Heading rotation (Rotates=true) | ⚠️ stored, not applied | `ctx.rotate()` not called before `drawImage` |
| MarkOrigin (write-only) | ❌ | Sets OriginX/OriginY visually in designer; write-only block prop, Go has no case |
| Z (draw order) | ⚠️ ignored | Same as Ball |

---

### Chart

**Implemented:** Type (0=Line, 1=Scatter, 2=Area, 3=Bar, 4=Pie), BackgroundColor, LegendEnabled, XFromZero, YFromZero. SVG renderer reads ChartData2D children's live `Elements` state. EntryClick fires on point/bar click. SetDomain, SetRange, ResetAxes.

| Gap | Type | Why |
|-----|------|-----|
| GridEnabled | ⚠️ stored, not rendered | No grid lines drawn in SVG; would need to compute tick positions and draw `<line>` elements |
| AxesTextColor | ⚠️ stored, not rendered | No axis labels; would need tick labels with this color |
| Labels (x-axis) | ⚠️ stored, not applied | Label list for x-axis tick labels; rendering requires axis scale engine |
| Description | ⚠️ stored, not rendered | A subtitle below the chart |
| PieRadius | ⚠️ stored | Pie always fills available space; `PieRadius` as a percentage is not applied |
| ExtendDomainToInclude, ExtendRangeToInclude | ⚠️ logged | These should expand the manual domain/range bounds; currently no-op |
| Colors (per-series, from ChartData2D) | ⚠️ not used | Per-bar or per-slice color arrays not applied |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### ChartData2D

**Implemented:** Color, Label (used in Chart legend). AddEntry, RemoveEntry, Clear, ImportFromList, ImportFromTinyDB, GetAllEntries, DoesEntryExist, GetEntriesWithXValue, GetEntriesWithYValue.

| Gap | Type | Why |
|-----|------|-----|
| EntryClick | ⚠️ event fires on Chart not ChartData2D | The SVG click handlers target the parent Chart component; Falcon blocks must use `when Chart1.EntryClick` |
| Colors (per-point list) | ⚠️ stored, not applied | Per-point color list not used in SVG rendering |
| DataLabelColor | ⚠️ stored | Data labels (value annotations on bars/points) not rendered |
| HighlightDataPoints | ⚠️ stores color, not applied | No highlighted-point rendering |
| ImportFromCloudDB, ImportFromDataFile, ImportFromSpreadsheet, ImportFromWeb | ❌ | Marked Unsupported — those source components are not simulated |
| ChangeDataSource, RemoveDataSource | ❌ | No live data-source binding infrastructure |

---

### Trendline

Registered as non-visible. Stores Color, Model, StrokeWidth, StrokeStyle, Extend, Visible, ChartData. **No regression is computed and no line is rendered.**

| Gap | Type | Why |
|-----|------|-----|
| All regression output props (CorrelationCoefficient, LinearCoefficient, RSquared, YIntercept, XIntercepts, Predictions, Results, ExponentialBase, ExponentialCoefficient, LogarithmCoefficient, LogarithmConstant, QuadraticCoefficient) | ❌ | No OLS/regression engine implemented |
| Updated event | ❌ | No regression computation to trigger it |
| Trendline rendering on Chart | ❌ | Chart SVG has no trendline path; it would need to read this component's state and draw a best-fit line over the ChartData2D series |
| GetResultValue | ⚠️ always returns null | No results computed |
| DisconnectFromChartData | ⚠️ no-op | No data binding to disconnect |

---

### Map

**Implemented:** Leaflet slippy map with OpenStreetMap tiles. Latitude/Longitude/ZoomLevel (center+zoom), EnablePan, EnableZoom, ShowZoom (Leaflet zoomControl), CustomUrl (tile URL). Events: Ready, TapAtPoint, DoubleTapAtPoint, LongPressAtPoint, BoundsChange, ZoomChange. Map feature children (Marker/Circle/LineString/Polygon/Rectangle) rendered as Leaflet layers and updated reactively. PanTo method.

| Gap | Type | Why |
|-----|------|-----|
| MapType (satellite/terrain/etc.) | ⚠️ stored, ignored | Only OpenStreetMap tiles; satellite/terrain require licensed tile sources (Google/Mapbox) |
| EnableRotation, Rotation | ⚠️ stored | Leaflet cannot rotate the base tile layer |
| ShowCompass | ⚠️ stored | No compass widget; would require a CSS overlay |
| ShowScale | ⚠️ stored | Leaflet has a scale control (`L.control.scale()`) — easy to add |
| ShowUser, UserLatitude, UserLongitude | ⚠️ stored | No device GPS; `navigator.geolocation` could approximate but requires permission |
| BoundingBox | ⚠️ accepted | Not used for `fitBounds()` — needs Go SetProperty case + Svelte `fitBounds` call |
| CenterFromString | ⚠️ stored without parsing | Write-only "lat, lng" string; needs a Go SetProperty case to parse and apply to Latitude/Longitude |
| Features (list) | ⚠️ not managed | Feature list managed as design-tree children; programmatic add/remove not wired |
| LocationSensor | ⚠️ ignored | Component reference binding; not meaningful in the browser |
| FeatureDrag, FeatureStartDrag, FeatureStopDrag events | ❌ | Leaflet drag events not wired (Leaflet fires `move`/`dragstart`/`dragend` but not mapped to Svelte emitEvent) |
| GotFeatures event | ❌ | LoadFromURL is not implemented |
| LoadError, InvalidPoint events | ❌ | No URL loading; no coordinate validation |
| CreateMarker | ⚠️ effect sent, not rendered | Effect does not dynamically add a Leaflet marker layer — would need the Svelte side to handle a `map-create-marker` effect |
| FeatureFromDescription, LoadFromURL | ❌ | Marked Unsupported |
| Save | ❌ | Cannot write to device storage |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Marker

**Implemented:** Latitude, Longitude, FillColor (SVG pin fill), StrokeColor, StrokeWidth, StrokeOpacity, FillOpacity, Title + Description (Leaflet popup), Draggable. SetLocation, ShowInfobox, HideInfobox. FeatureClick fires via parent Map when marker is clicked.

| Gap | Type | Why |
|-----|------|-----|
| Click, LongClick, Drag, StartDrag, StopDrag (feature-direct events) | ❌ | See cross-cutting gap #5 — events fire on the parent Map, not this component |
| ImageAsset | ⚠️ ignored | Custom icon image from assets not resolved; always uses SVG divIcon |
| AnchorHorizontal, AnchorVertical | ⚠️ stored | `iconAnchor` not computed from these values; hardcoded to [12,36] |
| Type (read-only) | ⚠️ returns null | Not set in state; Go GetProperty would need a special case |
| BearingToPoint, BearingToFeature | ❌ | Marked Unsupported — geodesic bearing not implemented |
| DistanceToPoint, DistanceToFeature | ❌ | Marked Unsupported — Haversine not implemented |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

### Circle

**Implemented:** Latitude, Longitude, Radius, FillColor, FillOpacity, StrokeColor, StrokeOpacity, StrokeWidth, Draggable. SetLocation, ShowInfobox, HideInfobox. FeatureClick fires via parent Map.

| Gap | Type | Why |
|-----|------|-----|
| Click, LongClick, Drag, StartDrag, StopDrag | ❌ | See cross-cutting gap #5 |
| Type | ⚠️ returns null | Not set in state |
| DistanceToPoint, DistanceToFeature | ❌ | Marked Unsupported |

---

### LineString

**Implemented:** PointsFromString (JSON or space-delimited lat,lng pairs parsed to Leaflet polyline), StrokeColor, StrokeOpacity, StrokeWidth, Draggable. ShowInfobox, HideInfobox. FeatureClick fires via parent Map.

| Gap | Type | Why |
|-----|------|-----|
| Click, LongClick, Drag, StartDrag, StopDrag | ❌ | See cross-cutting gap #5 |
| Points (read-write list) | ⚠️ not synced | Setting `Points` via blocks does not update the Leaflet layer; only `PointsFromString` is watched on the Svelte side |
| Type | ⚠️ returns null | Not set in state |
| DistanceToPoint, DistanceToFeature | ❌ | Marked Unsupported |

---

### Polygon

**Implemented:** PointsFromString, FillColor, FillOpacity, StrokeColor, StrokeOpacity, StrokeWidth, Draggable. ShowInfobox, HideInfobox. FeatureClick fires via parent Map. Centroid returns empty list.

| Gap | Type | Why |
|-----|------|-----|
| Click, LongClick, Drag, StartDrag, StopDrag | ❌ | See cross-cutting gap #5 |
| HolePoints, HolePointsFromString | ⚠️ stored, not applied | Holes not passed to `L.polygon([outerRing, holeRing1, ...])` |
| Points (read-write list) | ⚠️ not synced | Same as LineString |
| Centroid | ⚠️ returns empty list | Centroid computation not implemented |
| Type | ⚠️ returns null | Not set in state |
| DistanceToPoint, DistanceToFeature | ❌ | Marked Unsupported |

---

### Rectangle

**Implemented:** NorthLatitude, SouthLatitude, EastLongitude, WestLongitude, FillColor, FillOpacity, StrokeColor, StrokeOpacity, StrokeWidth, Draggable. SetCenter (adjusts all four bounds), Bounds and Center return computed values. ShowInfobox, HideInfobox.

| Gap | Type | Why |
|-----|------|-----|
| Click, LongClick, Drag, StartDrag, StopDrag | ❌ | See cross-cutting gap #5 |
| Type | ⚠️ returns null | Not set in state |
| DistanceToPoint, DistanceToFeature | ❌ | Marked Unsupported |

---

### FeatureCollection

Non-rendering logical container. Children render via their parent Map's Leaflet layer management.

| Gap | Type | Why |
|-----|------|-----|
| FeatureClick, FeatureLongClick, FeatureDrag, FeatureStartDrag, FeatureStopDrag, GotFeatures, LoadError | ❌ | No event re-dispatch infrastructure from child-feature clicks; FeatureCollection itself has no Leaflet surface |
| FeaturesFromGeoJSON | ⚠️ stored | GeoJSON is not parsed; no dynamic layer creation from JSON |
| Features (list) | ⚠️ not managed | Design-tree children only; no programmatic add/remove |
| FeatureFromDescription | ❌ | Marked Unsupported |
| LoadFromURL | ❌ | Marked Unsupported — CORS restricted |
| WidthPercent, HeightPercent | ❌ | See cross-cutting gap #1 |

---

## Prioritised Next Steps

Based purely on impact-to-effort ratio:

| Priority | Work | Components fixed | Notes |
|----------|------|------------------|-------|
| **1** | Fix WidthPercent/HeightPercent Go handler (one-time) | All 44 | Add intercept in `SetProperty` for these two write-only props on all types |
| **2** | Canvas draw-op replay (`handleComponentActions` canvas-draw handler) | Canvas, Ball, ImageSprite | Unblocks all drawing-based apps |
| **3** | Sprite hit-testing + sprite events | Ball, ImageSprite | Needed for any interactive game/animation app |
| **4** | Map feature direct events (Marker.Click etc.) | Marker, Circle, LineString, Polygon, Rectangle | Semantic parity with AI event model |
| **5** | Map ShowScale (`L.control.scale()`), ShowCompass, MapType | Map | Small additions in `initOrUpdateMap` |
| **6** | ListView `DividerThickness`, `ElementCornerRadius`, `ElementMarginsWidth` | ListView | CSS one-liners |
| **7** | Trendline regression + Chart rendering | Trendline, Chart | OLS in JS + SVG path on Chart |
| **8** | Animation loop (rAF/setInterval) | Ball, ImageSprite | Unblocks autonomous motion |
| **9** | Canvas GetPixelColor (sync WASM readback) | Canvas | Needs ImageData transfer across WASM boundary |
| **10** | WebViewer WebViewString bridge (same-origin only) | WebViewer | Inject `__AppInventorWebViewString` via `srcdoc` approach |
