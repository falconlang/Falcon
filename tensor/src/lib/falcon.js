import { StreamLanguage, HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { tags } from '@lezer/highlight'
import { EditorView, ViewPlugin, Decoration } from '@codemirror/view'
import { RangeSetBuilder } from '@codemirror/state'

// ── Selection highlight (mark decoration, same mechanism as bracketMatching) ──
const selMark = Decoration.mark({ class: 'cm-sel' })

export const selectionHighlight = ViewPlugin.fromClass(
  class {
    constructor(view) { this.decorations = this.build(view) }
    update(u) {
      if (u.selectionSet || u.docChanged || u.viewportChanged)
        this.decorations = this.build(u.view)
    }
    build(view) {
      const builder = new RangeSetBuilder()
      for (const { from, to } of view.state.selection.ranges)
        if (from < to) builder.add(from, to, selMark)
      return builder.finish()
    }
  },
  { decorations: v => v.decorations },
)

// ── Token sets ────────────────────────────────────────────────────
const KW = new Set([
  'global', 'local', 'func', 'if', 'else', 'while', 'for', 'in',
  'break', 'yield', 'this', 'when', 'any', 'step',
])

const ATOMS = new Set(['true', 'false'])

const TYPES = new Set([
  'text', 'number', 'list', 'dict', 'emptyList', 'emptyText', 'base10', 'hexa', 'bin',
])

const BUILTINS = new Set([
  // IO & control
  'println', 'openScreen', 'openScreenWithValue', 'closeScreenWithValue',
  'getStartValue', 'closeScreen', 'closeApp', 'getPlainStartText',
  // Math
  'sqrt', 'abs', 'neg', 'log', 'exp', 'round', 'ceil', 'floor',
  'sin', 'cos', 'tan', 'asin', 'acos', 'atan', 'atan2', 'degrees', 'radians',
  'decToHex', 'decToBin', 'hexToDec', 'binToDec',
  'randInt', 'randFloat', 'setRandSeed',
  'min', 'max', 'avgOf', 'maxOf', 'minOf', 'geoMeanOf', 'stdDevOf', 'stdErrOf', 'modeOf',
  'mod', 'rem', 'quot', 'formatDecimal',
  // Values
  'copyList', 'copyDict', 'makeColor', 'splitColor', 'dec', 'octal', 'set', 'get',
  // List methods & lambdas
  'listLen', 'add', 'containsItem', 'indexOf', 'insert', 'remove',
  'appendList', 'lookupInPairs', 'join', 'slice', 'random', 'reverseList',
  'toCsvRow', 'toCsvTable', 'sort', 'allButFirst', 'allButLast', 'pairsToDict',
  'map', 'filter', 'reduce',
  // Text methods
  'textLen', 'trim', 'uppercase', 'lowercase', 'startsWith', 'contains',
  'containsAny', 'containsAll', 'split', 'splitAtFirst', 'splitAtAny',
  'splitAtFirstOfAny', 'splitAtSpaces', 'reverse',
  'csvRowToList', 'csvTableToList', 'segment',
  'replace', 'replaceFrom', 'replaceFromLongestFirst',
  // Dict methods
  'dictLen', 'delete', 'getAtPath', 'setAtPath', 'containsKey',
  'mergeInto', 'walkTree', 'keys', 'values', 'toPairs',
])

// ── StreamLanguage mode (CM5-style tokenizer, adapted to CM6) ─────
// StreamLanguage.define() accepts the same startState/token interface
// as CodeMirror 5 modes, so the tokenizer is ported verbatim.
export const falconLanguage = StreamLanguage.define({
  startState: () => ({ inString: false, afterFunc: false }),

  token(stream, state) {
    // Finish a string that started on a previous line
    if (state.inString) {
      while (!stream.eol()) {
        const c = stream.next()
        if (c === '\\') { stream.next(); continue }
        if (c === '"')  { state.inString = false; break }
      }
      return 'string'
    }

    if (stream.eatSpace()) return null

    // // line comment
    if (stream.match('//')) { stream.skipToEnd(); return 'comment' }

    // "string literal"
    if (stream.peek() === '"') {
      stream.next()
      state.inString = true
      while (!stream.eol()) {
        const c = stream.next()
        if (c === '\\') { stream.next(); continue }
        if (c === '"')  { state.inString = false; break }
      }
      return 'string'
    }

    // @Component decorator
    if (stream.peek() === '@') {
      stream.next()
      stream.eatWhile(/\w/)
      state.afterFunc = false
      return 'tag'
    }

    // #RRGGBB colour literal
    if (stream.peek() === '#') {
      stream.next()
      stream.eatWhile(/[0-9a-fA-F]/)
      state.afterFunc = false
      return 'number'
    }

    // Numeric literal (no leading minus — that's a unary operator)
    if (stream.match(/^\d+(\.\d+)?/)) {
      state.afterFunc = false
      return 'number'
    }

    // Identifier / keyword
    if (stream.match(/^[a-zA-Z_]\w*/)) {
      const w = stream.current()
      if (KW.has(w))       { state.afterFunc = (w === 'func'); return 'keyword' }
      if (ATOMS.has(w))    { state.afterFunc = false; return 'atom' }
      if (TYPES.has(w))    { state.afterFunc = false; return 'type' }
      if (BUILTINS.has(w)) { state.afterFunc = false; return 'builtin' }
      if (state.afterFunc) { state.afterFunc = false; return 'def' }
      state.afterFunc = false
      return 'variable'
    }

    // Multi-char operators (must come before single-char)
    if (stream.match(/^(===|!==|==|!=|<=|>=|<<|>>|\|\||&&|\.\.|->)/)) return 'operator'
    // Single-char operators
    if (stream.match(/^[+\-*/%^<>&|~!_?=:]/)) return 'operator'

    stream.next()
    return null
  },
})

// ── Syntax highlighting ───────────────────────────────────────────
// StreamLanguage maps CM5 token names to @lezer/highlight tags:
//   keyword  → tags.keyword
//   string   → tags.string
//   comment  → tags.lineComment
//   number   → tags.number
//   builtin  → tags.standard(tags.name)
//   tag      → tags.tagName
//   type     → tags.typeName
//   atom     → tags.atom
//   operator → tags.operator
//   variable → tags.variableName
//   def      → tags.definition(tags.variableName)
export const falconHighlight = syntaxHighlighting(
  HighlightStyle.define([
    { tag: tags.keyword,                       color: 'var(--syn-keyword)',  fontWeight: '500' },
    { tag: tags.string,                        color: 'var(--syn-string)' },
    { tag: tags.number,                        color: 'var(--syn-number)' },
    { tag: tags.atom,                          color: 'var(--syn-atom)',     fontWeight: '500' },

    { tag: tags.standard(tags.name),           color: 'var(--syn-builtin)' },
    { tag: tags.definition(tags.variableName), color: 'var(--syn-def)',      fontWeight: '500' },

    { tag: tags.tagName,                       color: 'var(--syn-tag)',      fontWeight: '500' },
    { tag: tags.typeName,                      color: 'var(--syn-type)' },

    { tag: tags.comment,                       color: 'var(--syn-comment)',  fontStyle: 'italic' },
    { tag: tags.operator,                      color: 'var(--syn-operator)' },
    { tag: tags.variableName,                  color: 'var(--syn-variable)' },
  ])
)

// ── Editor chrome theme ───────────────────────────────────────────
export const editorTheme = EditorView.theme(
  {
    '&': {
      height: '100%',
      backgroundColor: 'var(--bg-editor)',
      color: 'var(--text)',
    },
    '.cm-scroller': {
      fontFamily: "'Droid Sans Mono', 'JetBrains Mono', monospace",
      fontSize: '13.5px',
      lineHeight: '1.18',
      fontVariantLigatures: 'none',
      overflow: 'auto',
    },
    '.cm-content': {
      paddingBottom: '80px',
      caretColor: 'var(--caret)',
    },
    '.cm-gutters': {
      backgroundColor: 'var(--bg-gutter)',
      borderRight: '1px solid var(--border)',
      color: 'var(--text-muted)',
    },
    '.cm-lineNumbers .cm-gutterElement': {
      fontSize: '12px',
      padding: '0 16px 0 10px',
      minWidth: '40px',
    },
    '.cm-activeLine': {
      backgroundColor: 'var(--bg-activeline)',
    },
    '.cm-activeLineGutter': {
      backgroundColor: 'var(--bg-activeline)',
      color: 'var(--text-sec)',
    },
    '.cm-sel': {
      background: 'var(--bg-selection)',
    },
    '& .cm-content ::selection': {
      background: 'transparent',
    },
    '.cm-matchingBracket': {
      backgroundColor: 'var(--bg-match-bracket)',
      outline: '1px solid var(--bg-match-bracket-outline)',
    },
    '.cm-cursor, .cm-dropCursor': {
      borderLeftColor: 'var(--caret)',
      borderLeftWidth: '2px',
    },
  },
  { dark: false }
)
