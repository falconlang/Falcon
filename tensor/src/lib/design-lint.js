import { ViewPlugin, Decoration } from '@codemirror/view'
import { RangeSetBuilder, StateEffect } from '@codemirror/state'

const triggerEffect = StateEffect.define()
const errorMark = Decoration.mark({ class: 'cm-design-error' })

const COMPONENT_EXCEPTIONS = new Set(['Screen'])

function maskIgnoredSpans(text) {
  const chars = text.split('')
  let inString = false
  let inLineComment = false

  for (let i = 0; i < chars.length; i++) {
    const ch = chars[i]
    const next = chars[i + 1]

    if (inLineComment) {
      if (ch === '\n') inLineComment = false
      else chars[i] = ' '
      continue
    }

    if (inString) {
      if (ch === '\n') {
        continue
      }
      if (ch === '\\') {
        chars[i] = ' '
        if (i + 1 < chars.length) chars[++i] = ' '
        continue
      }
      if (ch === '"') inString = false
      chars[i] = ' '
      continue
    }

    if (ch === '/' && next === '/') {
      chars[i] = ' '
      chars[++i] = ' '
      inLineComment = true
      continue
    }

    if (ch === '"') {
      chars[i] = ' '
      inString = true
    }
  }

  return chars.join('')
}

function getCompInfo(name, cache) {
  if (cache.has(name)) return cache.get(name)
  try {
    const json = window.describeComponent(name)
    const info = json ? JSON.parse(json) : null
    cache.set(name, info)
    return info
  } catch {
    cache.set(name, null)
    return null
  }
}

function componentNameBeforeOpening(text, openingIndex) {
  const before = text.slice(0, openingIndex)
  const oldMatch = before.match(/@([A-Za-z]\w*)\s*$/)
  if (oldMatch) return oldMatch[1]
  const match = before.match(/([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*[A-Za-z_]\w*)?\s*$/)
  return match ? match[1] : null
}

function computeErrors(text) {
  if (typeof window.describeComponent !== 'function') return []

  const errors = []
  const cache = new Map()
  const searchable = maskIgnoredSpans(text)

  // Unknown legacy @ComponentName
  const compRe = /@([A-Za-z]\w*)/g
  let m
  while ((m = compRe.exec(searchable)) !== null) {
    if (COMPONENT_EXCEPTIONS.has(m[1])) continue
    if (!getCompInfo(m[1], cache)) {
      errors.push({ from: m.index + 1, to: m.index + m[0].length })
    }
  }

  // Unknown Type or Type.instance component declarations
  const newCompRe = /(?:^|[{}\n])\s*([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*[A-Za-z_]\w*)?\s*(?=\{|$)/gm
  while ((m = newCompRe.exec(searchable)) !== null) {
    const compName = m[1]
    if (COMPONENT_EXCEPTIONS.has(compName)) continue
    if (!getCompInfo(compName, cache)) {
      const from = m.index + m[0].indexOf(compName)
      errors.push({ from, to: from + compName.length })
    }
  }

  // Unknown PascalCase property names (skip lowercase like `id`)
  const propRe = /(?:^|[,{\n])\s*([A-Z]\w*)\s*:/gm
  while ((m = propRe.exec(searchable)) !== null) {
    const propName = m[1]
    const propStart = m.index + m[0].indexOf(propName)
    const propEnd = propStart + propName.length

    // Find enclosing component declaration by scanning backward
    let depth = 0
    let compName = null
    for (let i = propStart - 1; i >= 0; i--) {
      const ch = searchable[i]
      if (ch === '}') depth++
      else if (ch === '{') {
        if (depth === 0) {
          compName = componentNameBeforeOpening(searchable, i)
          break
        }
        depth--
      }
    }

    if (!compName) continue
    const compInfo = getCompInfo(compName, cache)
    if (!compInfo) continue // component already flagged

    const known = compInfo.blockProperties?.some(p => p.name === propName)
    if (!known) errors.push({ from: propStart, to: propEnd })
  }

  errors.sort((a, b) => a.from - b.from)
  return errors
}

function buildDecorations(view) {
  const text = view.state.doc.toString()
  const errors = computeErrors(text)
  const builder = new RangeSetBuilder()
  for (const { from, to } of errors) {
    builder.add(from, to, errorMark)
  }
  return builder.finish()
}

export const designLint = ViewPlugin.fromClass(
  class {
    constructor(view) {
      this.decorations = buildDecorations(view)
      this._watchWasm(view)
    }

    _watchWasm(view) {
      if (typeof window.describeComponent === 'function') return
      const poll = () => {
        if (typeof window.describeComponent === 'function') {
          view.dispatch({ effects: triggerEffect.of(null) })
        } else {
          setTimeout(poll, 200)
        }
      }
      setTimeout(poll, 200)
    }

    update(update) {
      const triggered = update.transactions.some(t =>
        t.effects.some(e => e.is(triggerEffect))
      )
      if (update.docChanged || triggered) {
        this.decorations = buildDecorations(update.view)
      }
    }
  },
  { decorations: v => v.decorations },
)
