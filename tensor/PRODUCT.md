# Product

## Register

product

## Users

The primary user is an **App Inventor builder leveling up**: someone fluent in block-based programming who is ready for the speed, power, and reuse that text gives, but who does not want to abandon the App Inventor ecosystem to get it.

They arrive already knowing the mental model (screens, components, events, blocks) and are crossing a bridge, not starting over. Their context is hands-on building: they open Tensor to write a Falcon program, see it become real App Inventor blocks, wire up a screen in the visual designer, and test it live on a phone over the companion. The job to be done is **build a working App Inventor app faster and with more control than blocks alone allow**, while keeping perfect fidelity with `.aia` projects and the existing block format underneath.

Secondary contexts the product also serves: experienced developers who find Blockly slow, and AI-assisted building (Falcon was designed to make App Inventor apps agent-editable).

## Product Purpose

Tensor is the web IDE for **Falcon**, a text-based language designed for MIT App Inventor. It lets people program App Inventor apps in syntax instead of blocks, then converts losslessly in both directions (Falcon ↔ Blockly XML ↔ YAIL). Around the editor it provides a visual component designer, `.aia` import, a live-test companion (paired to a device over WebRTC via the MIT rendezvous server), and an in-browser simulation runtime. The language compiles to WASM and runs entirely client-side.

It exists because block programming hits a ceiling: large apps become unwieldy to author, diff, reuse, and automate in a purely visual form. Falcon removes that ceiling without breaking compatibility. Success looks like a builder who used to drag blocks now writing Falcon comfortably, moving between code and the designer without friction, and shipping a `.aia` that is indistinguishable from one made the traditional way.

## Brand Personality

Calm, editorial, precise. Three words: **composed, exact, unobtrusive.**

The interface should feel like a quiet, well-set document that happens to be a development environment, with the spatial calm of print and the density of a code editor. It is confident without being loud, friendly to someone arriving from blocks without being childish, and dense with capability without being busy. The voice is plain and exact: it names things correctly, never over-explains, and trusts the user. The content (code, blocks, the designed screen) is always the foreground; the chrome stays in the background.

## Anti-references

- **Toy / childish block UI.** No candy colors, oversized rounded everything, or cartoonish chrome in the style of Scratch or Blockly's default toolbox. Tensor is a deliberate step up in seriousness from the blocks editor it complements.
- **Generic dark IDE.** No default VS-Code-style blue-on-near-black. Tensor's warm off-white identity is a deliberate choice, not a theme to be swapped for the expected developer-tool darkness.
- Also avoid, by extension of the above: gradient/glassmorphism SaaS-landing aesthetics, hero-metric template cards, and cluttered enterprise toolbars that bury the work.

## Design Principles

1. **The work is the foreground.** Code, blocks, and the designed screen occupy the eye; toolbars, gutters, and status chrome recede until needed. Every structural decision pulls the UI toward the background.
2. **Lossless by default.** The product's core promise is round-trip fidelity with App Inventor. The interface should make that trustworthy and visible, never make the user wonder whether something was lost in translation.
3. **A bridge, not a cliff.** Meet the App Inventor user where they are. Text and blocks are two views of one truth; moving between code, designer, and live device should feel continuous, not like crossing between separate tools.
4. **Quiet confidence.** Restraint over decoration. One warm material, one accent that appears only when something is selected or interactive. Precision is the personality.
5. **Keyboard-first, instant feedback.** A real editor for real building: keyboard operability throughout, immediate state feedback (run, selection, live indicators), and no decorative motion that slows the work.

## Accessibility & Inclusion

Target **WCAG 2.1 AA** with a keyboard-first stance:

- AA contrast for text and meaningful UI against the warm off-white surfaces; verify the muted/faint neutral tiers and status pairs (`run`, `warn`, `error` on their soft backgrounds) meet ratio.
- Full keyboard operability for every action, with a visible focus indicator that is never removed without an equivalent replacement (the existing 2px accent `:focus-visible` outline).
- Icon-only controls carry accessible names (current `title` attributes should be backed by `aria-label` in production).
- Status is never conveyed by color alone: pair every state color with an icon, label, or text so it survives color-blindness and low-color displays.
- Honor `prefers-reduced-motion` for the cell-flash, context-menu, and running-spinner animations.
