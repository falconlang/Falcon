# Falcon IDE Feature Checklist

This document defines the editing behaviors needed to make the Tensor web IDE feel like a real Falcon IDE, closer to IntelliJ than a plain text editor.

## Editing Flow

### 1. Tab Out Of Closing Delimiters

When the cursor is immediately before a closing delimiter, `Tab` should move past it instead of inserting spaces.

Example:

```falcon
if (checkTextBoxes()|)
```

Pressing `Tab` should become:

```falcon
if (checkTextBoxes())|
```

Rules:
- Works for `)`, `]`, `}`, `"`, and `'`.
- If multiple adjacent closing delimiters are ahead, `Tab` moves across the next one only.
- If a completion menu is open, `Tab` still accepts the completion first.
- If text is selected, existing indentation behavior still wins.

### 2. Smart Pair Insertion

Typing an opening delimiter should insert the matching closer and place the cursor inside.

Pairs:
- `(` -> `(|)`
- `[` -> `[|]`
- `{` -> `{|}`
- `"` -> `"|"`

Selection behavior:
- Selecting `value` and typing `(` becomes `(value)`.
- Selecting `value` and typing `"` becomes `"value"`.
- Selecting one or more lines and typing `{` wraps them in a block with indentation.

### 3. Smart Skip Over Existing Closers

Typing a closing delimiter when that same delimiter is already under the cursor should move over it instead of inserting a duplicate.

Example:

```falcon
println("hello"|)
```

Typing `)` should become:

```falcon
println("hello")|
```

### 4. Smart Enter In Blocks

Pressing `Enter` between braces should create a properly indented block.

Example:

```falcon
if (checkTextBoxes()) {|}
```

Enter:

```falcon
if (checkTextBoxes()) {
  |
}
```

Also applies to:
- `func name() = {|`
- `when Button.Click {|`
- `while (...) {|`
- `for item in list {|`

### 5. Complete Statement Into Block

When the cursor is at the end of a block-introducing statement, pressing `{` or `Enter` should complete the block shape.

Example:

```falcon
if (checkTextBoxes())|
```

Typing `{`:

```falcon
if (checkTextBoxes()) {
  |
}
```

### 6. Line Comment Toggle

With selected text, `Ctrl+/` toggles `//` comments for all selected non-empty lines.

Already implemented baseline:
- Selection + `Ctrl+/` comments selected lines.
- Repeating `Ctrl+/` uncomments if all selected non-empty lines are commented.
- Plain `/` remains normal text input.

Needed refinements:
- Optional: support no-selection `Ctrl+/` for the current line.

### 7. Duplicate Line Or Selection

Shortcut should duplicate the current line or selected block.

Suggested binding:
- `Ctrl+D` / `Cmd+D`

Behavior:
- No selection duplicates the current line below.
- Selection duplicates the selected text directly below.
- Cursor/selection moves to the duplicate.

### 8. Move Line Or Selection

Move selected lines up or down without breaking indentation.

Suggested bindings:
- `Alt+ArrowUp`
- `Alt+ArrowDown`

Behavior:
- Works with no selection by moving the current line.
- Works with multi-line selections.
- Preserves selection around the moved block.

### 9. Delete Line

Suggested binding:
- `Ctrl+Y` or `Cmd+Backspace`, depending on platform expectations.

Behavior:
- Deletes current line when there is no selection.
- Deletes selected lines when selection spans lines.

### 10. Expand And Shrink Selection

Selection should expand by syntax-aware regions.

Suggested bindings:
- `Ctrl+W` / `Cmd+W`: expand
- `Ctrl+Shift+W` / `Cmd+Shift+W`: shrink

Expansion order examples:
- word
- function call arguments
- parenthesized expression
- statement
- block body
- whole block

## Code Intelligence

### 11. Context-Aware Completion

Completion should be semantic, not just text matching.

Falcon completions:
- Keywords: `func`, `when`, `if`, `else`, `while`, `for`, `local`, `global`, `yield`, `break`
- Built-in functions from Falcon WASM metadata
- Built-in chain methods from Falcon WASM metadata
- User-defined functions
- Local variables
- Global variables
- Event parameters
- Component names from `Screen1.design`
- Component properties after `ComponentName.`
- Component methods after `ComponentName.`
- Component events after `when ComponentName.`

Design completions:
- Component type names
- Component IDs
- Component properties inside the current component block
- Boolean values for boolean properties
- Common property value snippets

Needed refinements:
- Rank local/project symbols above global built-ins.
- Rank component events only in `when` headers.
- Rank writable properties above read-only properties in assignment contexts.
- Avoid noisy suggestions in comments and strings.

### 12. Signature Help

When typing inside a function or method call, show parameter help.

Examples:
- `println(|)` shows `println(value)`
- `randInt(|)` shows `randInt(from, to)`
- `Notifier1.ShowAlert(|)` shows App Inventor method parameters
- `"abc".replace(|)` shows `.replace(from, to)`

Behavior:
- Highlight the active parameter based on commas.
- Show return type where available.
- Use Falcon engine metadata where possible.

### 13. Hover Documentation

Hovering symbols should show useful documentation.

Targets:
- Built-in functions
- Chain methods
- Component properties
- Component methods
- Component events
- Component types in Screen1.design
- User-defined functions
- Local/global variables

Content:
- Signature
- Return type
- Description
- Source location for user-defined symbols

### 14. Go To Definition

`Ctrl+Click` / `Cmd+Click` should jump to definitions.

Implemented baseline:
- `Ctrl+Click` / `Cmd+Click` jumps from Falcon component references to `Screen1.design`.
- `F12` jumps to local function, component, or variable definitions where known.

Targets:
- User-defined `func`
- Local/global variables
- Component IDs from Falcon to Screen1.design
- Component references from Screen1.design to usage list in Falcon

### 15. Find Usages

Find all references for:
- Functions
- Component IDs
- Variables
- Event handlers
- Design properties where possible

Implemented baseline:
- `Alt+F7` opens a flat usages panel.
- Component usages search across both Falcon and Screen1.design.
- Function and variable usages search the relevant Falcon scope/file.

### 16. Rename Symbol

Rename should update all safe references.

Targets:
- Function names
- Component IDs across Falcon and Screen1.design
- Local/global variables
- Event parameters

Rules:
- Rename component ID in design updates all Falcon references.
- Rename function updates calls.
- Rename local variable only updates references in scope.

Implemented baseline:
- `F2` renames component IDs across both tabs.
- `F2` renames function call sites and declarations.
- `F2` renames local word references in the active file.

### 17. Quick Fixes

Diagnostics should offer actions.

Examples:
- Unknown function with suggestion: rename to suggested function.
- Wrong arg count: insert missing argument placeholders or remove extras.
- Unknown component: create component in Screen1.design.
- Unknown property casing: replace with canonical property name.
- Missing event parameter list: insert canonical parameters.
- Design property typo: replace with suggested property.

Implemented baseline:
- `Alt+Enter` opens quick fixes near the cursor.
- Similar-symbol fixes are generated from the project symbols and WASM/catalog metadata.
- Function/method call signature fixes can fill argument placeholders from known metadata.

## Formatting

### 18. Format Document

A formatter should normalize both tabs.

Falcon formatting:
- 2-space indentation
- Spaces around binary operators
- Consistent brace placement
- Stable blank lines between top-level declarations

Design formatting:
- 2-space indentation
- One component child per line
- Compact one-line leaf components where appropriate
- Stable property ordering only if we explicitly choose that rule

Implemented baseline:
- Toolbar Format and `Shift+Alt+F` format the active document.
- Falcon/design indentation is normalized to 2 spaces.
- Common Falcon operators and design property separators are normalized.

### 19. Format Selection

Same as document formatting, but scoped to selected lines or selected block.

Implemented baseline:
- If text is selected, Toolbar Format / `Shift+Alt+F` formats the selected line range only.

### 20. Auto Indentation

Indentation should react while typing:
- New line after `{` indents one level.
- Line starting with `}` dedents.
- `else` aligns with matching `if`.
- Continued expressions indent predictably.

## Navigation And Structure

### 21. Document Symbol Outline

Show structure for the active file.

Falcon outline:
- Functions
- Event handlers
- Globals

Design outline:
- Screen
- Component tree
- IDs and component types

### 22. Breadcrumbs

Show current context:
- Falcon: `when AddButton.Click > if`
- Design: `Screen1 > HorizontalArrangement > Button.DivideButton`

### 23. Error Navigation

Shortcuts:
- Next error
- Previous error

Behavior:
- Moves between inline diagnostics in both editors.
- Keeps focus in the relevant pane.

Implemented baseline:
- `F8` moves to the next diagnostic.
- `Shift+F8` moves to the previous diagnostic.
- Status bar shows active context breadcrumbs and error counts.

## Cross-File Falcon And Design Awareness

### 24. Component Synchronization

The IDE should treat Falcon and Screen1.design as one project.

Behaviors:
- Component IDs created in design immediately appear in Falcon completions.
- Falcon unknown-component errors can create components in design.
- Renaming component IDs updates both tabs.
- Deleting a component warns if Falcon still references it.

Implemented baseline:
- Component completions, hover docs, go-to-definition, usages, and rename share the same `Screen1.design` component index.

### 25. Event Handler Generation

From a component in Screen1.design, generate a Falcon event handler.

Example:
- On `Button.AddButton`, offer `when AddButton.Click { }`.

From Falcon:
- Typing `when AddButton.` should suggest only events valid for `Button`.

### 26. Component Property Editing Assistance

Inside Screen1.design:
- Suggest only valid properties for the enclosing component type.
- Suggest expected value types.
- Show property documentation.
- Warn on read-only or invisible properties if they are not valid design-time properties.

## Refactoring And Productivity

### 27. Extract Function

Select Falcon statements and extract them into:

```falcon
func newName(params) = {
  ...
}
```

Then replace selection with `newName(args)`.

Implemented baseline:
- `Ctrl+Alt+M` extracts selected Falcon statements into a new zero-argument function and replaces the selection with a function call.

### 28. Surround With

Selection can be wrapped with:
- `if (...) { selection }`
- `while (...) { selection }`
- `for item in list { selection }`
- `func name() = { selection }`

Implemented baseline:
- `Ctrl+Alt+T` surrounds selected Falcon code with `if`, `while`, `for`, or `function` templates.

### 29. Generate Code

Available generation actions:
- Function template
- Event handler
- Local variable assignment
- Component definition in design
- Common UI layout patterns in design

## Search

### 30. Project Search

Search both Falcon and Screen1.design.

Features:
- Case-sensitive toggle
- Whole-word toggle
- Regex toggle
- Replace across both tabs

Implemented baseline:
- Toolbar Find and `Ctrl+Shift+F` search both Falcon and Screen1.design.
- Results open in the flat IDE panel and jump back to the source.

### 31. Symbol Search

Fast palette for:
- Functions
- Components
- Events
- Properties

Implemented baseline:
- Toolbar Symbols and `Ctrl+Shift+O` show functions, event handlers, and design components.

## Implementation Order

Implemented batch 1 baseline:
- Tab out of closing delimiters.
- Smart skip over existing closers.
- Smart pair insertion and surround selection.
- Smart Enter in blocks.
- Complete block-introducing statements into block skeletons.
- Preserve selection through `Ctrl+/` line comment toggling.
- Duplicate line/selection with `Ctrl+D` / `Cmd+D`.
- Move line/selection with `Alt+ArrowUp` and `Alt+ArrowDown`.
- Delete line with `Ctrl+Y` and `Cmd+Backspace`.

Implemented batch 2 baseline:
- Semantic completion uses Falcon WASM metadata plus project functions, variables, event parameters, and component IDs from `Screen1.design`.
- Local/project symbols are ranked above built-ins.
- `when Component.` suggests only events for that component.
- `when ` can generate valid component event handler snippets.
- Signature help appears while editing Falcon function, method, and component method calls.
- Hover documentation appears for Falcon symbols and Screen1.design components/properties.

Implemented batch 3 baseline:
- Format document/selection with Toolbar Format or `Shift+Alt+F`.
- Go to definition with `Ctrl+Click` / `Cmd+Click` and `F12`.
- Find usages with `Alt+F7`.
- Rename with `F2`.
- Quick fixes with `Alt+Enter`.
- Next/previous diagnostics with `F8` / `Shift+F8`.
- Project search with Toolbar Find or `Ctrl+Shift+F`.
- Project symbol list with Toolbar Symbols or `Ctrl+Shift+O`.
- Extract function with `Ctrl+Alt+M`.
- Surround with `Ctrl+Alt+T`.

Recommended batch 1:
1. Tab out of closing delimiters.
2. Smart skip over existing closers.
3. Smart pair insertion and surround selection.
4. Smart Enter in blocks.
5. Preserve selection for comment toggling.
6. Duplicate line/selection.
7. Move line/selection.

Recommended batch 2:
1. Signature help.
2. Hover documentation.
3. Better completion ranking.
4. Event handler generation.
5. Component rename across Falcon and design.

Recommended batch 3:
1. Formatter.
2. Go to definition.
3. Find usages.
4. Quick fixes.
5. Extract function and surround-with refactors.
