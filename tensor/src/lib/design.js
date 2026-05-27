import { StreamLanguage, HighlightStyle, syntaxHighlighting } from '@codemirror/language'
import { tags } from '@lezer/highlight'

// ── StreamLanguage tokenizer for .design files ────────────────────
export const designLanguage = StreamLanguage.define({
  startState: () => ({ inString: false, expectComponentInstance: false }),

  token(stream, state) {
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

    // "string"
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

    // Legacy @ComponentName form
    if (stream.peek() === '@') {
      stream.next()
      stream.eatWhile(/\w/)
      state.expectComponentInstance = false
      return 'tag'
    }

    // Numeric literal
    if (stream.match(/^\d+(\.\d+)?/)) {
      state.expectComponentInstance = false
      return 'number'
    }

    // Component instance after Type.instance
    if (state.expectComponentInstance && stream.match(/^[a-zA-Z_]\w*/)) {
      state.expectComponentInstance = false
      return 'def'
    }

    // Identifier — peek ahead for ':' to classify as attribute name
    if (stream.match(/^[a-zA-Z_]\w*/)) {
      const w = stream.current()
      if (w === 'true' || w === 'false') {
        state.expectComponentInstance = false
        return 'atom'
      }
      if (stream.match(/^\s*:/, false)) {
        state.expectComponentInstance = false
        return 'attribute'
      }
      if (/^[A-Z]/.test(w) && stream.match(/^\s*(?:\.[a-zA-Z_]\w*)?\s*(?=\{|$)/, false)) {
        state.expectComponentInstance = stream.match(/^\s*\./, false)
        return 'tag'
      }
      state.expectComponentInstance = false
      return 'variable'
    }

    // Brackets
    if (stream.match(/^[{}]/)) {
      state.expectComponentInstance = false
      return 'bracket'
    }

    // Component type/instance separator
    if (stream.match(/^\./)) return 'operator'

    // Punctuation
    if (stream.match(/^[,:]/)) {
      state.expectComponentInstance = false
      return 'operator'
    }

    stream.next()
    state.expectComponentInstance = false
    return null
  },
})

// ── Syntax highlight ──────────────────────────────────────────────
export const designHighlight = syntaxHighlighting(
  HighlightStyle.define([
    { tag: tags.tagName,       color: 'var(--syn-tag)' },
    { tag: tags.attributeName, color: 'var(--syn-attribute)', fontWeight: '500' },
    { tag: tags.string,        color: 'var(--syn-string)' },
    { tag: tags.number,        color: 'var(--syn-number)' },
    { tag: tags.atom,          color: 'var(--syn-atom)',      fontWeight: '500' },
    { tag: tags.comment,       color: 'var(--syn-comment)',   fontStyle: 'italic' },
    { tag: tags.operator,      color: 'var(--syn-operator)' },
    { tag: tags.bracket,       color: 'var(--syn-bracket)' },
    { tag: tags.variableName,  color: 'var(--syn-variable)' },
    { tag: tags.definition(tags.variableName), color: 'var(--syn-def)', fontWeight: '500' },
  ])
)
