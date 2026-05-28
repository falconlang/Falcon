# Tensor — Design Documentation

## Philosophy

Tensor is a notebook interface designed around one core conviction: **the code is the product, not the chrome**. Every structural decision pulls the UI toward the background so the content — cells, outputs, charts — occupies the foreground. There are no gradients, no decorative shadows, no glassmorphism. The aesthetic is warm-minimal: off-white surfaces, tinted neutrals, and a single cool accent that only appears when something is selected or interactive.

The interface draws from editorial and developer-tool traditions simultaneously. It has the spatial calm of a printed document and the density of a code editor.

---

## Color System

All colors are defined as CSS custom properties on `:root`. Nothing is hardcoded in component rules except the canvas drawing code and one tinted-dark hover value.

### Palette

| Token | Value | Role |
|---|---|---|
| `--bg` | `#FAFAF8` | Page background — slightly warm, never pure white |
| `--surface` | `#FFFFFF` | Card/cell surfaces, topbar, sidebar |
| `--border` | `#E8E6E0` | Primary structural borders |
| `--border-soft` | `#F0EDE8` | Subtle inner dividers (cell header lines, output gutters) |
| `--text` | `#1A1916` | Primary text — warm near-black, not pure |
| `--text-muted` | `#7A7870` | Secondary labels, descriptions, stream output |
| `--text-faint` | `#B0AEA8` | Tertiary — line numbers, badges, separators |
| `--accent` | `#2A6F97` | Selection, active state, focus ring, links |
| `--accent-soft` | `#EBF3F9` | Active sidebar item background, tag fills |
| `--accent-mid` | `#4A8FB7` | Reserved for accent hover states |
| `--run` | `#1A6B3A` | Run success, stream prefix, completion badges |
| `--run-soft` | `#EBF5EE` | Success badge backgrounds |
| `--run-dot` | `#3ABF6E` | Live kernel indicator dot |
| `--warn` | `#8B5E00` | Warning text |
| `--warn-soft` | `#FBF4E6` | Warning badge backgrounds |
| `--error` | `#8B2020` | Error text, danger menu items |
| `--error-soft` | `#FCE8E8` | Error badge backgrounds, danger hover |
| `--cell-active` | `#F5F3EF` | Hover background for buttons and sidebar items |

### Tint Principle

Every neutral is tinted toward warm yellow-brown (the `8` and `9` in the hex endings). There is no pure gray (`#808080`), no pure black, no pure white. The warmth creates subconscious cohesion — the entire interface reads as one material.

### Semantic Pairing

Colors are always used as token pairs: a foreground token and its soft background counterpart (e.g. `--run` on `--run-soft`, `--error` on `--error-soft`). This enforces consistency across all status states.

---

## Typography

### Typefaces

| Role | Family | Weights |
|---|---|---|
| UI text | DM Sans | 300, 400, 500 |
| Code, output, badges | DM Mono | 300, 400, 500 |

DM Sans is a modern geometric sans with slightly rounded terminals — approachable but precise. DM Mono is its companion: same design DNA, monospace. Using a matched pair means code and prose feel like they belong in the same document rather than colliding.

Both are loaded from Google Fonts with `preconnect` to `fonts.googleapis.com` **and** `fonts.gstatic.com` (with `crossorigin`) to minimize font load latency.

### Scale

| Context | Size | Weight | Notes |
|---|---|---|---|
| Notebook h1 | 22px | 500 | `letter-spacing: -0.025em` — tighter for display |
| Notebook h2 | 16px | 500 | Underlined with `--border-soft` |
| Notebook h3 | 11px | 500 | Uppercase, `letter-spacing: 0.05em` — label style |
| Body paragraph | 14px | 400 | `line-height: 1.75` — generous for readability |
| Code | 13px | 400 | `line-height: 1.65` — tighter for dense code blocks |
| Code output | 12.5px | 400 | Slightly smaller to visually separate from source |
| UI labels (buttons) | 12–12.5px | 400/500 | |
| Micro labels (badges, status) | 10–11px | 500 | Uppercase with letter-spacing |
| Line numbers | 12px | 400 | `color: --text-faint`, non-selectable |

### Font Smoothing

```css
-webkit-font-smoothing: antialiased;
-moz-osx-font-smoothing: grayscale;
```

Applied globally. Without this, DM Sans renders heavier than intended on macOS and some Linux configurations.

---

## Layout

### Shell Structure

```
#app (column flex, 100vh)
├── #topbar       (48px, fixed height)
├── #toolbar      (38px, fixed height)
├── #main         (flex: 1, overflow hidden)
│   ├── #sidebar  (220px, fixed width)
│   └── #notebook-wrap  (flex: 1, overflow-y: auto)
│       └── #notebook   (max-width: 780px, centered)
└── #statusbar    (24px, fixed height)
```

The shell uses `height: 100vh` with `overflow: hidden` on `#app`. Only `#notebook-wrap` scrolls. This keeps the topbar, toolbar, sidebar, and statusbar anchored at all times — the notebook is a viewport into a document, not a full-page scroll.

### Notebook Column

`max-width: 780px` centered with `margin: 0 auto`. This is a deliberate reading width — wide enough for code, narrow enough that the eye doesn't have to travel far. The `padding: 32px 24px 80px` gives breathing room at the top and a generous bottom clearance so the last cell doesn't feel pinned to the viewport edge.

The 24px horizontal padding is deliberately less than the 32px top padding — horizontal space is more valuable in a code context.

### Sidebar

Fixed at `220px`. Contains three sections: Contents (table of contents), Variables (active namespace), Files. Sections are separated by hairline dividers (`--border-soft`). The sidebar can be toggled via the toolbar button, collapsing to `display: none`.

### Gutter

Cells have an absolute-positioned `.cell-gutter` at `left: -28px` — outside the cell's box. This means gutter controls (move up/down/delete) don't occupy any of the cell's own width and appear only on hover, keeping the notebook visually clean. The gutter fades in via `opacity` transition, not layout shift.

---

## Component Design

### Topbar

Height: 48px. Contains the logo mark, file name input, and right-aligned actions. The file name is an `<input>` that looks like text until hovered or focused — progressive disclosure of editability.

The **Run all** button (`tb-btn.primary`) uses `--text` (near-black) as its background — the darkest color in the palette — which makes it the single most visually authoritative element in the topbar without needing a colorful accent. `margin-left: 6px` provides breathing room from the runtime badge without a spacer element.

The **runtime badge** uses a pill shape (`border-radius: 99px`) with the `--run-dot` green indicator. The badge is intentionally not interactive — it's a status display, not a control.

### Toolbar

Height: 38px. Secondary chrome below the topbar. Uses slightly smaller buttons (`tl-btn`) than the topbar. Button groups are separated by 1px vertical separators (`tl-sep`). Add cell buttons use pill shapes to visually differentiate them from the action buttons — they create rather than modify.

### Cells

Every cell has three layers of border state:

| State | Border |
|---|---|
| Default | `transparent` (invisible, but reserves 1.5px of space) |
| Hover | `--border` (shows the cell's bounding box) |
| Active/focused | `--accent` (blue — the only non-neutral color on the cell itself) |

The 1.5px border is always present in the layout (transparent) to prevent layout shift between states. `transition: border-color 0.15s` keeps the state change smooth.

**Code cells** (`code-cell`) have a white surface with a cell header bar. The header contains the run button, a `py` type badge, the execution count `[n]`, and a three-dot overflow menu. The overflow menu is `opacity: 0` until the cell is hovered or active.

**Markdown cells** (`md-cell`) have a transparent background — they read as document content, not UI containers. Their only affordance is `md-content:hover { background: rgba(0,0,0,0.02) }` — the faintest possible hover feedback.

### Code Editor Area

The code editor is a `display: flex` row:
- Left column: `.line-nums` — fixed-width, right-aligned, non-selectable, `color: --text-faint`
- Right column: `.code-area` — `contenteditable="true"`, `spellcheck="false"`, `white-space: pre`

`white-space: pre` preserves indentation. The `contenteditable` approach means no textarea is needed — the browser handles cursor, selection, and text input natively. Line numbers sync on every `input` event via `syncLineNums()`.

Background is `var(--bg)` (slightly off-white) rather than `--surface` (pure white) — this creates a micro-contrast between the cell's header/output sections and the code editing area, subtly signalling that the editor is a different zone.

### Syntax Highlighting

A minimal token system using single-letter class names for zero overhead:

| Class | Color | Meaning |
|---|---|---|
| `.k` | `#7C3AED` purple | Keywords (`import`, `def`, `for`, `return`) |
| `.s` | `#1A6B3A` green | String literals |
| `.n` | `#C2410C` orange | Numeric literals |
| `.c` | `#9CA3AF` gray italic | Comments |
| `.f` | `#1D4ED8` blue | Function names |
| `.b` | `#374151` dark, bold | Builtins (`print`, `range`, `int`) |
| `.t` | `#5B21B6` deep purple | Types |
| `.o` | `#6B7280` neutral | Operators |
| `.p` | `#374151` dark | Punctuation |

The syntax data is stored as token arrays in the cell's `code` property — each token is either a plain string (for whitespace/newlines) or an object `{t: 'classname', v: 'text'}`. This format is compact and renders to clean `<span class="k">import</span>` HTML.

### Output Types

Four output variants, all inside `.cell-output` with a shared `.output-gutter` (40px left strip) and `.output-content` (flex: 1) layout:

| Type | Rendering |
|---|---|
| `stream` | Preformatted text in `--text-muted` — `print()` output |
| `table` | `output-table` with uppercase letter-spaced column headers, alternating hover rows |
| `chart` | Canvas element (520×180 logical pixels, HiDPI-aware) |
| `result` | Green "Completed" badge + preformatted text |

The output gutter mirrors the line number gutter in the code editor — visual alignment between input and output sides.

### Loss/Accuracy Chart

Drawn on a `<canvas>` with Canvas 2D API. Fully procedural — data is generated mathematically (exponential decay + noise) rather than hardcoded.

Features:
- **HiDPI rendering**: `devicePixelRatio` scaling with CSS `width`/`height` override keeps it crisp on retina
- **Fill area**: Semi-transparent blue fill under the loss curve (`rgba(42,111,151,0.07)`) for visual weight
- **Dual lines**: Loss in `--accent` blue at full opacity, accuracy in `--run` green at 60% opacity
- **Grid**: 4-row horizontal grid in `--border-soft`
- **Axes**: Left y-axis and bottom x-axis in `--border`
- **Legend**: Inline color swatches at top-left, no border, no box
- **Typography**: `10px DM Mono` matching the rest of the code-adjacent UI

### Context Menu

Triggered by right-click on any code cell. Positioned with `Math.min` clamping to prevent overflow beyond the viewport edges.

Animation: `opacity 0→1` + `translateY(-4px) scale(0.97) → translateY(0) scale(1)` with `cubic-bezier(0.16, 1, 0.3, 1)` (a fast-out exponential curve). This gives it a natural snap-in feel — it decelerates sharply at the end rather than bouncing. Duration: 140ms.

The menu is removed from interaction (`pointer-events: none`) when hidden so it doesn't intercept clicks beneath it.

---

## Interaction States

Every interactive element has all required states defined:

### Buttons (toolbar, add, run)

| State | Treatment |
|---|---|
| Default | Transparent background, `--text-muted` label |
| Hover | `--cell-active` background, `--text` label (0.12s transition) |
| Focus (keyboard) | 2px `--accent` outline, offset 2px — never removed for accessibility |
| Active | `scale(0.95)` transform — kinetic click feedback |
| Disabled | Handled by the browser (Run all button during execution) |

Run buttons additionally use `--run-soft` background and `--run` color on hover, distinguishing the execute action from navigation actions.

### Active State (`:active`)

Transform-based, not color-based:
- Run button: `scale(0.88)` — more aggressive, emphasizes the "fire" action
- Toolbar/add buttons: `scale(0.95)` — subtle
- Sidebar items and context menu items: `opacity: 0.65` — depresses without transforming layout

All `:active` transitions use `0.05s` — fast enough to register as immediate.

### Running Animation

When a cell is executing, the run button's SVG icon is replaced with a spinning dashed circle:

```html
<circle ... stroke-dasharray="10" stroke-dashoffset="10" class="running"/>
```

The `.running` class applies `animation: spin 0.8s linear infinite`. After a simulated delay (600–1000ms randomized), the icon reverts and a `cellFlash` animation plays on the cell — a brief `--accent-soft` background wash fading to transparent over 500ms (`ease-out`).

### Cell Selection

`setActive()` removes `.active` from all cells, adds it to the target. The `.active` class:
- Sets `border-color: var(--accent)` — the cell gains a blue frame
- Makes `.cell-gutter` and `.cell-menu-btn` `opacity: 1` (they're hidden by default)

---

## Transitions

All transitions use design-system timing:

| Duration | Usage |
|---|---|
| 0.05s | Active/click state (needs to feel instant) |
| 0.1s | Hover background/color changes |
| 0.12s | Toolbar button state changes |
| 0.14s | Context menu spring animation |
| 0.15s | Cell border color, gutter opacity |
| 0.5s | Cell flash animation |

Easing:
- State changes (hover): linear or ease (short enough it doesn't matter)
- Context menu entry: `cubic-bezier(0.16, 1, 0.3, 1)` — exponential ease-out, snappy
- Flash animation: `ease-out` — fades naturally

Only `transform` and `opacity` are animated where possible. The cell flash is the exception — it animates `background` on the cell, but since it's a 500ms once-off effect (not continuous), the perf cost is acceptable.

---

## Accessibility

- **Focus indicators**: `button:focus-visible` uses a 2px `--accent` outline. The file name input uses `border-color: var(--accent)` on focus. No focus ring is removed without a replacement.
- **Semantic HTML**: Toolbar and topbar use `<button>` elements throughout, not `<div onclick>`. Sidebar items are `<div>` with `cursor: pointer` (a known trade-off for list items that don't need role).
- **No `[title] { cursor: default }`**: This rule was removed in polish — it was overriding `cursor: pointer` on all titled buttons, breaking pointer feedback across the entire interface.
- **ARIA**: Buttons without visible text have `title` attributes for tooltip-level disclosure. A production build would add `aria-label` to icon-only buttons.
- **Keyboard**: Shift+Enter runs the active cell and advances focus to the next cell (`handleCodeKey`). Escape dismisses the context menu.

---

## Scrollbar

Custom WebKit scrollbar for the notebook column:

```css
width: 6px
track: transparent
thumb: --border, border-radius 99px
thumb:hover: --text-faint
```

Firefox equivalent via `scrollbar-width: thin; scrollbar-color: var(--border) transparent`.

The narrow scrollbar keeps the reading gutter clean without hiding the scroll position entirely.

---

## Status Bar

24px height. Left side: last run time, cursor position, encoding, cell count (live-updated by JS). Right side: memory usage. Uses `·` separators in `--border` color — visible but subordinate.

The status bar uses the same `--surface` background and `--border-soft` top border as the toolbar, making the chrome layers feel consistently layered.

---

## Design Decisions Not Taken

- **Dark mode**: Intentionally omitted. The warm off-white is a deliberate identity choice, not a default. Dark mode would require a complete token retheme and the current palette is specifically tuned for the light surface.
- **Resizable sidebar**: Would add complexity (drag handle, state persistence) without meaningful payoff for a notebook interface where the sidebar is primarily navigational.
- **Real-time syntax highlighting**: The token array format supports static pre-tokenized code. A production implementation would wire a tokenizer (e.g. highlight.js or a Lezer grammar) into the `input` event, but this is out of scope for the UI layer.
- **Virtualized cell list**: At the scale this interface targets (tens of cells), full DOM rendering is appropriate. Virtualization would complicate the `scrollIntoView` and selection model.
