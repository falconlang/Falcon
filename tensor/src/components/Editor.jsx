import { forwardRef, useCallback, useImperativeHandle, useMemo, useRef } from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { Decoration, EditorView, GutterMarker, WidgetType, gutter, keymap, layer } from '@codemirror/view'
import { EditorSelection, Prec, RangeSetBuilder, StateField } from '@codemirror/state'
import { copyLineDown, deleteLine, indentMore, indentLess, moveLineDown, moveLineUp } from '@codemirror/commands'
import { indentUnit } from '@codemirror/language'
import { acceptCompletion, autocompletion, completionStatus } from '@codemirror/autocomplete'
import { falconLanguage, falconHighlight, editorTheme, selectionHighlight } from '../lib/falcon'
import { designLanguage, designHighlight } from '../lib/design'
import { designLint } from '../lib/design-lint'
import {
  createBlocklyPreviewWorkspace,
  fitBlocklyWorkspace,
  loadXmlIntoBlocklyWorkspace,
  safeSvgResize,
  waitForBlockly,
} from '../lib/blockly-preview'

const CLOSE_DELIMITERS = new Set([')', ']', '}', '"', "'"])
const EMPTY_EXTENSIONS = []
const PAIRS = {
  '(': ')',
  '[': ']',
  '{': '}',
  '"': '"',
  "'": "'",
}

const tabKeymap = Prec.high(keymap.of([
  {
    key: 'Ctrl-/',
    run(view) {
      return toggleSelectedLineComments(view)
    },
  },
  {
    key: '(',
    run(view) {
      return insertSmartPair(view, '(', ')')
    },
  },
  {
    key: '[',
    run(view) {
      return insertSmartPair(view, '[', ']')
    },
  },
  {
    key: '{',
    run(view) {
      return insertSmartPair(view, '{', '}')
    },
  },
  {
    key: '"',
    run(view) {
      return insertSmartPair(view, '"', '"')
    },
  },
  {
    key: "'",
    run(view) {
      return insertSmartPair(view, "'", "'")
    },
  },
  {
    key: ')',
    run(view) {
      return skipExistingCloser(view, ')')
    },
  },
  {
    key: ']',
    run(view) {
      return skipExistingCloser(view, ']')
    },
  },
  {
    key: '}',
    run(view) {
      return skipExistingCloser(view, '}')
    },
  },
  {
    key: 'Enter',
    run(view) {
      if (completionStatus(view.state) === 'active') return acceptCompletion(view)
      return smartEnter(view)
    },
  },
  {
    key: 'Tab',
    run(view) {
      if (completionStatus(view.state) === 'active') return acceptCompletion(view)
      if (view.state.selection.ranges.some(r => !r.empty)) return indentMore(view)
      if (tabOutOfClosingDelimiter(view)) return true
      const { from } = view.state.selection.main
      view.dispatch({ changes: { from, insert: '  ' }, userEvent: 'input' })
      return true
    },
    shift: indentLess,
  },
  {
    key: 'Mod-d',
    run: copyLineDown,
  },
  {
    key: 'Alt-ArrowUp',
    run: moveLineUp,
  },
  {
    key: 'Alt-ArrowDown',
    run: moveLineDown,
  },
  {
    key: 'Ctrl-y',
    run: deleteLine,
  },
  {
    key: 'Mod-Backspace',
    run: deleteLine,
  },
]))

function textAt(doc, pos) {
  return pos >= 0 && pos < doc.length ? doc.sliceString(pos, pos + 1) : ''
}

function rangesAreEmpty(state) {
  return state.selection.ranges.every(range => range.empty)
}

function tabOutOfClosingDelimiter(view) {
  if (!rangesAreEmpty(view.state)) return false

  const ranges = []
  for (const range of view.state.selection.ranges) {
    if (!CLOSE_DELIMITERS.has(textAt(view.state.doc, range.head))) return false
    ranges.push(EditorSelection.cursor(range.head + 1))
  }

  view.dispatch({
    selection: EditorSelection.create(ranges, view.state.selection.mainIndex),
    scrollIntoView: true,
  })
  return true
}

function skipExistingCloser(view, closer) {
  if (!rangesAreEmpty(view.state)) return false

  const ranges = []
  for (const range of view.state.selection.ranges) {
    if (textAt(view.state.doc, range.head) !== closer) return false
    ranges.push(EditorSelection.cursor(range.head + 1))
  }

  view.dispatch({
    selection: EditorSelection.create(ranges, view.state.selection.mainIndex),
    scrollIntoView: true,
  })
  return true
}

function currentLineIndent(state, pos) {
  return state.doc.lineAt(pos).text.match(/^\s*/)?.[0] ?? ''
}

function blockIntroPrefix(prefix) {
  const trimmed = prefix.trim()
  return (
    /^if\s*\(.+\)$/.test(trimmed) ||
    /^while\s*\(.+\)$/.test(trimmed) ||
    /^for\s+\w+.+$/.test(trimmed) ||
    /^when\s+[A-Za-z_]\w*\.[A-Za-z_]\w*(?:\([^)]*\))?$/.test(trimmed) ||
    /^func\s+[A-Za-z_]\w*\s*\([^)]*\)\s*=$/.test(trimmed) ||
    /^else$/.test(trimmed) ||
    /^@?[A-Z][A-Za-z0-9_]*(?:\.[A-Za-z_]\w*)?$/.test(trimmed)
  )
}

function selectionWrapChange(state, range, opener, closer) {
  const selected = state.doc.sliceString(range.from, range.to)
  return { from: range.from, to: range.to, insert: opener + selected + closer }
}

function insertSmartPair(view, opener, closer) {
  const state = view.state

  if (state.selection.ranges.some(range => !range.empty)) {
    const changes = []
    const ranges = []
    let offset = 0
    for (const range of state.selection.ranges) {
      changes.push(selectionWrapChange(state, range, opener, closer))
      ranges.push(EditorSelection.range(
        range.from + offset + opener.length,
        range.to + offset + opener.length,
      ))
      offset += opener.length + closer.length
    }
    view.dispatch({
      changes,
      selection: EditorSelection.create(ranges, state.selection.mainIndex),
      userEvent: 'input.surround',
    })
    return true
  }

  if (opener === closer && rangesAreEmpty(state) && state.selection.ranges.length === 1) {
    const head = state.selection.main.head
    if (textAt(state.doc, head) === closer) return skipExistingCloser(view, closer)
  }

  if (opener === '{' && state.selection.ranges.length === 1) {
    const head = state.selection.main.head
    const line = state.doc.lineAt(head)
    const prefix = line.text.slice(0, head - line.from)
    if (blockIntroPrefix(prefix)) {
      const indent = currentLineIndent(state, head)
      const needsSpace = prefix.length > 0 && !/\s$/.test(prefix)
      const insert = `${needsSpace ? ' ' : ''}{\n${indent}  \n${indent}}`
      const cursor = head + (needsSpace ? 1 : 0) + 2 + indent.length + 2
      view.dispatch({
        changes: { from: head, insert },
        selection: EditorSelection.cursor(cursor),
        scrollIntoView: true,
        userEvent: 'input.block',
      })
      return true
    }
  }

  const changes = []
  const ranges = []
  let offset = 0
  for (const range of state.selection.ranges) {
    changes.push({ from: range.head, insert: opener + closer })
    ranges.push(EditorSelection.cursor(range.head + offset + opener.length))
    offset += opener.length + closer.length
  }

  view.dispatch({
    changes,
    selection: EditorSelection.create(ranges, state.selection.mainIndex),
    userEvent: 'input.pair',
  })
  return true
}

function smartEnter(view) {
  if (!rangesAreEmpty(view.state) || view.state.selection.ranges.length !== 1) return false

  const state = view.state
  const head = state.selection.main.head
  const previous = textAt(state.doc, head - 1)
  const next = textAt(state.doc, head)
  const indent = currentLineIndent(state, head)

  if (previous === '{' && next === '}') {
    const insert = `\n${indent}  \n${indent}`
    view.dispatch({
      changes: { from: head, insert },
      selection: EditorSelection.cursor(head + 1 + indent.length + 2),
      scrollIntoView: true,
      userEvent: 'input.block',
    })
    return true
  }

  const line = state.doc.lineAt(head)
  const prefix = line.text.slice(0, head - line.from)
  if (head === line.to && blockIntroPrefix(prefix)) {
    const needsSpace = prefix.length > 0 && !/\s$/.test(prefix)
    const insert = `${needsSpace ? ' ' : ''}{\n${indent}  \n${indent}}`
    const cursor = head + (needsSpace ? 1 : 0) + 2 + indent.length + 2
    view.dispatch({
      changes: { from: head, insert },
      selection: EditorSelection.cursor(cursor),
      scrollIntoView: true,
      userEvent: 'input.block',
    })
    return true
  }

  return false
}

function selectedLineRanges(state) {
  const ranges = []

  for (const range of state.selection.ranges) {
    if (range.empty) continue

    let fromLine = state.doc.lineAt(range.from).number
    let toLine = state.doc.lineAt(range.to).number
    const endLine = state.doc.line(toLine)
    if (range.to > range.from && range.to === endLine.from) {
      toLine = Math.max(fromLine, toLine - 1)
    }

    const previous = ranges[ranges.length - 1]
    if (previous && fromLine <= previous.to + 1) {
      previous.to = Math.max(previous.to, toLine)
    } else {
      ranges.push({ from: fromLine, to: toLine })
    }
  }

  return ranges
}

function lineCommentOffset(text) {
  const indent = text.match(/^\s*/)?.[0].length ?? 0
  return text.slice(indent).startsWith('//') ? indent : -1
}

function toggleSelectedLineComments(view) {
  const ranges = selectedLineRanges(view.state)
  if (!ranges.length) return false

  const lines = []
  for (const range of ranges) {
    for (let number = range.from; number <= range.to; number++) {
      const line = view.state.doc.line(number)
      if (line.text.trim()) lines.push(line)
    }
  }
  if (!lines.length) return true

  const shouldUncomment = lines.every(line => lineCommentOffset(line.text) >= 0)
  const changes = []

  for (const line of lines) {
    if (shouldUncomment) {
      const offset = lineCommentOffset(line.text)
      let to = line.from + offset + 2
      if (line.text.slice(offset + 2, offset + 3) === ' ') to++
      changes.push({ from: line.from + offset, to, insert: '' })
    } else {
      const indent = line.text.match(/^\s*/)?.[0].length ?? 0
      changes.push({ from: line.from + indent, insert: '// ' })
    }
  }

  const changeSet = view.state.changes(changes)
  view.dispatch({
    changes: changeSet,
    selection: view.state.selection.map(changeSet),
    userEvent: shouldUncomment ? 'delete.comment' : 'input.comment',
  })
  return true
}

const hiddenPreviewLine = Decoration.line({ class: 'cm-blockly-hidden-line' })
const diagnosticMark = Decoration.mark({ class: 'cm-diagnostic-error' })
const blocklyOverlayPadding = { top: 6, right: 12, bottom: 12, left: 8 }

class DiagnosticWidget extends WidgetType {
  constructor(diagnostics) {
    super()
    this.diagnostics = diagnostics
  }

  eq(other) {
    return JSON.stringify(other.diagnostics) === JSON.stringify(this.diagnostics)
  }

  toDOM() {
    const wrap = document.createElement('div')
    wrap.className = 'cm-diagnostic-widget'
    wrap.setAttribute('role', 'alert')

    const rail = document.createElement('div')
    rail.className = 'cm-diagnostic-rail'
    wrap.appendChild(rail)

    const body = document.createElement('div')
    body.className = 'cm-diagnostic-body'
    for (const diagnostic of this.diagnostics) {
      const line = document.createElement('div')
      line.className = 'cm-diagnostic-message'
      line.textContent = diagnostic.message || 'Falcon error'
      body.appendChild(line)
    }
    wrap.appendChild(body)
    return wrap
  }
}

function clampDiagnosticRange(state, diagnostic) {
  const lineNumber = Math.min(Math.max(Number(diagnostic.line) || 1, 1), state.doc.lines)
  const line = state.doc.line(lineNumber)
  const column = Math.max(Number(diagnostic.column) || 1, 1)
  const length = Math.max(Number(diagnostic.length) || 1, 1)
  const from = Math.min(line.to, line.from + column - 1)
  const to = Math.max(from, Math.min(line.to, from + length))
  return { line, from, to: to === from && from < line.to ? from + 1 : to }
}

function buildDiagnosticDecorationSet(state, activeDiagnostics) {
  const builder = new RangeSetBuilder()
  const byLine = new Map()
  const entries = []

  for (const diagnostic of activeDiagnostics) {
    const range = clampDiagnosticRange(state, diagnostic)
    entries.push({ from: range.from, to: range.to, decoration: diagnosticMark })
    const key = range.line.number
    const list = byLine.get(key) ?? []
    list.push(diagnostic)
    byLine.set(key, list)
  }

  for (const [lineNumber, lineDiagnostics] of [...byLine.entries()].sort((a, b) => a[0] - b[0])) {
    const line = state.doc.line(lineNumber)
    entries.push({
      from: line.to,
      to: line.to,
      decoration: Decoration.widget({
        widget: new DiagnosticWidget(lineDiagnostics),
        block: true,
        side: 1,
      }),
    })
  }

  entries
    .sort((a, b) => a.from - b.from || a.to - b.to)
    .forEach(entry => builder.add(entry.from, entry.to, entry.decoration))
  return builder.finish()
}

function diagnosticDecorations(diagnostics = []) {
  const activeDiagnostics = diagnostics.filter(Boolean)
  if (!activeDiagnostics.length) return []

  const diagnosticField = StateField.define({
    create(state) {
      return buildDiagnosticDecorationSet(state, activeDiagnostics)
    },
    update(value, transaction) {
      return transaction.docChanged
        ? buildDiagnosticDecorationSet(transaction.state, activeDiagnostics)
        : value
    },
    provide: field => EditorView.decorations.from(field),
  })

  return diagnosticField
}

function definitionClickExtension(onGoToDefinition) {
  if (!onGoToDefinition) return []
  return EditorView.domEventHandlers({
    mousedown(event, view) {
      if (event.button !== 0 || (!event.ctrlKey && !event.metaKey)) return false
      const pos = view.posAtCoords({ x: event.clientX, y: event.clientY })
      if (pos == null) return false
      event.preventDefault()
      return onGoToDefinition(pos) !== false
    },
  })
}

class BlockPreviewMarker extends GutterMarker {
  constructor(active) {
    super()
    this.active = active
  }

  eq(other) {
    return other.active === this.active
  }

  toDOM() {
    const button = document.createElement('button')
    button.type = 'button'
    button.tabIndex = -1
    button.className = `cm-block-preview-button${this.active ? ' active' : ''}`
    button.setAttribute('aria-label', 'Preview Blockly blocks')
    button.title = 'Preview Blockly blocks'

    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg')
    svg.setAttribute('viewBox', '0 0 24 24')
    svg.setAttribute('width', '16')
    svg.setAttribute('height', '16')
    svg.setAttribute('fill', this.active ? 'currentColor' : 'none')
    svg.setAttribute('stroke', this.active ? 'none' : 'currentColor')
    svg.setAttribute('stroke-width', '1.9')
    svg.setAttribute('stroke-linecap', 'round')
    svg.setAttribute('stroke-linejoin', 'round')
    svg.setAttribute('aria-hidden', 'true')
    svg.style.display = 'block'
    const path = document.createElementNS('http://www.w3.org/2000/svg', 'path')
    path.setAttribute('d', 'M4 7h3a1 1 0 0 0 1-1V5a2 2 0 0 1 4 0v1a1 1 0 0 0 1 1h3a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1a2 2 0 0 0 0 4h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-3a1 1 0 0 1-1-1v-1a2 2 0 0 0-4 0v1a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h1a2 2 0 0 0 0-4H4a1 1 0 0 1-1-1V8a1 1 0 0 1 1-1z')
    svg.appendChild(path)
    button.appendChild(svg)

    return button
  }
}

class BlocklyHoverPreviewMarker {
  constructor(preview, rect, onHover, hoverTimerRef) {
    this.preview = preview
    this.rect = rect
    this.onHover = onHover
    this.hoverTimerRef = hoverTimerRef
  }

  eq(other) {
    return (
      other.preview?.index === this.preview.index &&
      other.preview?.xml === this.preview.xml &&
      other.rect.left === this.rect.left &&
      other.rect.top === this.rect.top &&
      other.rect.width === this.rect.width &&
      other.rect.height === this.rect.height
    )
  }

  adjust(dom) {
    dom.style.left = `${this.rect.left}px`
    dom.style.top = `${this.rect.top}px`
    dom.style.width = `${this.rect.width}px`
    dom.style.height = `${this.rect.height}px`
  }

  draw() {
    const host = document.createElement('div')
    host.className = 'cm-blockly-hover-preview'
    this.adjust(host)

    // Hover-mode only: keep preview alive while pointer is over it
    if (this.onHover && this.hoverTimerRef) {
      host.style.pointerEvents = 'auto'
      host.addEventListener('mouseenter', () => {
        clearTimeout(this.hoverTimerRef.current)
        this.hoverTimerRef.current = null
      })
      host.addEventListener('mouseleave', () => {
        this.hoverTimerRef.current = setTimeout(() => {
          this.hoverTimerRef.current = null
          this.onHover?.(null)
        }, 120)
      })
    }

    const workspaceDiv = document.createElement('div')
    workspaceDiv.className = 'cm-blockly-hover-workspace'
    host.appendChild(workspaceDiv)

    let disposed = false
    let workspace = null
    let resizeObserver = null

    const mount = async () => {
      try {
        await waitForBlockly()
        await new Promise(resolve => requestAnimationFrame(resolve))
        if (disposed || !document.contains(workspaceDiv)) return

        workspace = createBlocklyPreviewWorkspace(workspaceDiv, { wheelZoom: false })
        host.__blocklyPreviewWorkspace = workspace
        resizeObserver = new ResizeObserver(() => {
          safeSvgResize(workspace)
          fitBlocklyWorkspace(workspace, blocklyOverlayPadding)
        })
        resizeObserver.observe(workspaceDiv)
        loadXmlIntoBlocklyWorkspace(workspace, this.preview.xml, { fit: true, padding: blocklyOverlayPadding })
        requestAnimationFrame(() => {
          if (!disposed) {
            safeSvgResize(workspace)
            fitBlocklyWorkspace(workspace, blocklyOverlayPadding)
          }
        })
      } catch (e) {
        if (!disposed) {
          host.classList.add('error')
          host.textContent = e.message || String(e)
        }
      }
    }

    mount()

    host.__disposeBlocklyPreview = () => {
      disposed = true
      resizeObserver?.disconnect()
      resizeObserver = null
      if (workspace) {
        workspace.dispose()
        workspace = null
        host.__blocklyPreviewWorkspace = null
      }
    }

    return host
  }

  update(dom, previous) {
    if (previous.preview?.index !== this.preview.index || previous.preview?.xml !== this.preview.xml) {
      dom.__disposeBlocklyPreview?.()
      const replacement = this.draw()
      dom.replaceWith(replacement)
      return true
    }
    this.adjust(dom)
    const workspace = dom.__blocklyPreviewWorkspace
    if (workspace) {
      safeSvgResize(workspace)
      fitBlocklyWorkspace(workspace, blocklyOverlayPadding)
    }
    return true
  }
}

function getPreviewTopBlocks(view) {
  const previewEl = view.dom.querySelector('.cm-blockly-hover-preview')
  if (!previewEl) return null
  const workspace = previewEl.__blocklyPreviewWorkspace
  if (!workspace) return null
  const blocks = workspace.getTopBlocks(true)
  return blocks.length ? blocks : null
}

function blockPreviewGutter(lines, activeLine, onHover, hoverTimerRef) {
  const lineSet = new Set(lines.filter(Number.isFinite))

  const getLineNumber = (view, line) => view.state.doc.lineAt(line.from).number
  const hasMarker = (view, line) => lineSet.has(getLineNumber(view, line))

  return gutter({
    class: 'cm-block-preview-gutter',
    initialSpacer: () => new BlockPreviewMarker(false),
    lineMarker(view, line) {
      const lineNumber = getLineNumber(view, line)
      return lineSet.has(lineNumber) ? new BlockPreviewMarker(lineNumber === activeLine) : null
    },
    lineMarkerChange: update => update.docChanged || update.viewportChanged,
    domEventHandlers: {
      mousedown(view, line, event) {
        if (!hasMarker(view, line)) return false
        event.preventDefault()
        return true
      },
      click(view, line) {
        if (!hasMarker(view, line)) return false
        const blocks = getPreviewTopBlocks(view)
        if (blocks) blocks.forEach(block => window.Blockly?.exportBlockAsPng(block))
        return true
      },
      contextmenu(view, line, event) {
        if (!hasMarker(view, line)) return false
        event.preventDefault()
        const blocks = getPreviewTopBlocks(view)
        if (blocks?.[0]) {
          window.Blockly?.blockToPngBlob(blocks[0])
            .then(blob => navigator.clipboard.write([new ClipboardItem({ 'image/png': blob })]))
            .catch(() => {})
        }
        return true
      },
      mousemove(view, line) {
        if (!hasMarker(view, line)) return false
        clearTimeout(hoverTimerRef.current)
        hoverTimerRef.current = null
        const lineNumber = getLineNumber(view, line)
        if (lineNumber !== activeLine) onHover?.(lineNumber)
        return true
      },
      mouseleave() {
        hoverTimerRef.current = setTimeout(() => {
          hoverTimerRef.current = null
          onHover?.(null)
        }, 120)
        return false
      },
    },
  })
}

function expressionPreviewRange(state, preview) {
  if (!preview || !Number.isFinite(preview.line)) return null
  if (preview.line < 1 || preview.line > state.doc.lines) return null

  const endLineNumber = Number.isFinite(preview.nextLine)
    ? Math.max(preview.line, Math.min(preview.nextLine - 1, state.doc.lines))
    : state.doc.lines
  const lineCount = endLineNumber - preview.line + 1
  return {
    fromLine: preview.line,
    toLine: endLineNumber,
    height: Math.max(28, Math.round(lineCount * 16)),
  }
}

function blockPreviewHiddenLines(previews) {
  const visiblePreviews = previews?.filter(Boolean) ?? []
  if (!visiblePreviews.length) return []

  return EditorView.decorations.of(view => {
    const lines = new Set()
    for (const preview of visiblePreviews) {
      const range = expressionPreviewRange(view.state, preview)
      if (!range) continue
      for (let line = range.fromLine; line <= range.toLine; line++) lines.add(line)
    }

    const builder = new RangeSetBuilder()
    for (const line of [...lines].sort((a, b) => a - b)) {
      builder.add(view.state.doc.line(line).from, view.state.doc.line(line).from, hiddenPreviewLine)
    }
    return builder.finish()
  })
}

function blockPreviewViewportRect(view) {
  const scrollerRect = view.scrollDOM.getBoundingClientRect()
  const gutters = view.scrollDOM.querySelector('.cm-gutters-before')
  const gutterRect = gutters?.getBoundingClientRect()
  const gutterWidth = gutterRect
    ? Math.max(0, gutterRect.right - scrollerRect.left)
    : Math.max(0, view.contentDOM.getBoundingClientRect().left - scrollerRect.left)
  return {
    left: view.scrollDOM.scrollLeft + Math.ceil(gutterWidth),
    width: Math.max(1, Math.floor(view.scrollDOM.clientWidth - gutterWidth)),
  }
}

function blockPreviewHoverLayer(preview, onHover, hoverTimerRef) {
  if (!preview) return []

  return layer({
    above: true,
    class: 'cm-blockly-hover-layer',
    update(update) {
      return update.docChanged || update.viewportChanged || update.geometryChanged
    },
    markers(view) {
      const range = expressionPreviewRange(view.state, preview)
      if (!range) return []
      const startLine = view.state.doc.line(range.fromLine)
      const endLine = view.state.doc.line(range.toLine)
      const startBlock = view.lineBlockAt(startLine.from)
      const endBlock = view.lineBlockAt(endLine.to)
      const height = Math.max(range.height, endBlock.bottom - startBlock.top)
      const viewportRect = blockPreviewViewportRect(view)
      return [
        new BlocklyHoverPreviewMarker(preview, {
          left: viewportRect.left,
          top: startBlock.top,
          width: viewportRect.width,
          height,
        }, onHover, hoverTimerRef),
      ]
    },
    destroy(layerDom) {
      for (const node of layerDom.querySelectorAll('.cm-blockly-hover-preview')) {
        node.__disposeBlocklyPreview?.()
      }
    },
  })
}

function blockPreviewAllLayer(entries) {
  if (!entries?.length) return []

  return layer({
    above: true,
    class: 'cm-blockly-hover-layer',
    update(update) {
      return update.docChanged || update.viewportChanged || update.geometryChanged
    },
    markers(view) {
      const viewportRect = blockPreviewViewportRect(view)
      return entries.flatMap(entry => {
        const range = expressionPreviewRange(view.state, entry)
        if (!range) return []
        const startLine = view.state.doc.line(range.fromLine)
        const endLine   = view.state.doc.line(range.toLine)
        const startBlock = view.lineBlockAt(startLine.from)
        const endBlock   = view.lineBlockAt(endLine.to)
        const height = Math.max(range.height, endBlock.bottom - startBlock.top)
        return [new BlocklyHoverPreviewMarker(entry, {
          left: viewportRect.left,
          top: startBlock.top,
          width: viewportRect.width,
          height,
        })]
      })
    },
    destroy(layerDom) {
      for (const node of layerDom.querySelectorAll('.cm-blockly-hover-preview')) {
        node.__disposeBlocklyPreview?.()
      }
    },
  })
}

export const falconExtensions = [
  falconLanguage,
  falconHighlight,
  editorTheme,
  selectionHighlight,
  indentUnit.of('  '),
  tabKeymap,
]

export const designExtensions = [
  designLanguage,
  designHighlight,
  editorTheme,
  selectionHighlight,
  indentUnit.of('  '),
  tabKeymap,
  designLint,
]

const Editor = forwardRef(function Editor({
  value,
  onChange,
  onCursorChange,
  onFocusChange,
  extensions = falconExtensions,
  blockPreviewLines = [],
  activeBlockPreviewLine = null,
  activeBlockPreview = null,
  onBlockPreviewLineHover,
  showBlockPreviewGutter = false,
  blocklyMode = false,
  allBlockPreviews = [],
  diagnostics = [],
  completionSource = null,
  extraExtensions = EMPTY_EXTENSIONS,
  onGoToDefinition,
}, ref) {
  const cmRef = useRef(null)
  const viewRef = useRef(null)
  const hoverTimerRef = useRef(null)
  const hiddenBlockPreviews = useMemo(
    () => blocklyMode ? allBlockPreviews : activeBlockPreview ? [activeBlockPreview] : [],
    [blocklyMode, allBlockPreviews, activeBlockPreview],
  )

  useImperativeHandle(ref, () => ({
    getView() {
      return viewRef.current ?? cmRef.current?.view ?? null
    },
    focus() {
      const view = viewRef.current ?? cmRef.current?.view
      view?.focus()
    },
    goTo(pos) {
      const view = viewRef.current ?? cmRef.current?.view
      if (!view) return false
      const safePos = Math.max(0, Math.min(view.state.doc.length, Number(pos) || 0))
      view.dispatch({
        selection: EditorSelection.cursor(safePos),
        scrollIntoView: true,
      })
      view.focus()
      return true
    },
    replaceRange(from, to, insert) {
      const view = viewRef.current ?? cmRef.current?.view
      if (!view) return false
      view.dispatch({
        changes: { from, to, insert },
        selection: EditorSelection.cursor(from + String(insert ?? '').length),
        scrollIntoView: true,
      })
      view.focus()
      return true
    },
  }), [])

  const handleCreateEditor = useCallback((view) => {
    viewRef.current = view
  }, [])

  const editorExtensions = useMemo(
    () => [
      ...extensions,
      showBlockPreviewGutter
        ? blockPreviewGutter(blockPreviewLines, activeBlockPreviewLine, onBlockPreviewLineHover, hoverTimerRef)
        : [],
      blockPreviewHiddenLines(hiddenBlockPreviews),
      blocklyMode
        ? blockPreviewAllLayer(allBlockPreviews)
        : blockPreviewHoverLayer(activeBlockPreview, onBlockPreviewLineHover, hoverTimerRef),
      completionSource
        ? autocompletion({
          override: [completionSource],
          activateOnTyping: true,
          maxRenderedOptions: 14,
          closeOnBlur: false,
        })
        : [],
      definitionClickExtension(onGoToDefinition),
      diagnosticDecorations(diagnostics),
      ...extraExtensions,
    ],
    [extensions, blockPreviewLines, activeBlockPreviewLine, activeBlockPreview, onBlockPreviewLineHover, showBlockPreviewGutter, blocklyMode, allBlockPreviews, hiddenBlockPreviews, completionSource, onGoToDefinition, diagnostics, extraExtensions],
  )

  const handleUpdate = useCallback(
    (viewUpdate) => {
      if (viewUpdate.selectionSet || viewUpdate.docChanged) {
        const sel = viewUpdate.state.selection.main
        const lineInfo = viewUpdate.state.doc.lineAt(sel.head)
        onCursorChange(
          lineInfo.number,
          sel.head - lineInfo.from + 1,
          viewUpdate.state.doc.length,
          sel.empty,
        )
      }
    },
    [onCursorChange],
  )

  const handleFocus = useCallback(() => onFocusChange?.(true),  [onFocusChange])
  const handleBlur  = useCallback(() => onFocusChange?.(false), [onFocusChange])

  return (
    <div className="editor-wrap">
      <CodeMirror
        ref={cmRef}
        value={value}
        extensions={editorExtensions}
        onChange={onChange}
        onUpdate={handleUpdate}
        onCreateEditor={handleCreateEditor}
        onFocus={handleFocus}
        onBlur={handleBlur}
        basicSetup={{
          lineNumbers: true,
          highlightActiveLineGutter: true,
          highlightSpecialChars: false,
          history: true,
          foldGutter: false,
          drawSelection: false,
          dropCursor: false,
          allowMultipleSelections: false,
          indentOnInput: true,
          syntaxHighlighting: false,
          bracketMatching: true,
          closeBrackets: false,
          autocompletion: false,
          rectangularSelection: false,
          crosshairCursor: false,
          highlightActiveLine: true,
          highlightSelectionMatches: false,
          tabSize: 2,
        }}
        style={{ height: '100%' }}
      />
    </div>
  )
})

export default Editor
