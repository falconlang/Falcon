import { useState, useRef, useEffect, useCallback, useMemo } from 'react'

import Toolbar from './components/Toolbar.jsx'
import Editor, { falconExtensions, designExtensions } from './components/Editor.jsx'
import StatusBar from './components/StatusBar.jsx'
import ComponentPopup from './components/ComponentPopup.jsx'
import PropertyPopup from './components/PropertyPopup.jsx'
import CompanionModal from './components/CompanionModal.jsx'
import IdePanel from './components/IdePanel.jsx'
import {
  compileForCompanion,
  connectCompanion,
  extractComponentDefs,
  pollRendezvous,
  validateSources,
} from './lib/companion.js'
import { loadIdeCatalog, mistToXmlResult } from './lib/falcon-wasm.js'
import {
  breadcrumbsForPosition,
  diagnosticLocations,
  documentSymbols,
  extractFunction,
  findDefinition,
  findUsages,
  formatSource,
  lineColToOffset,
  projectSearch,
  quickFixesAt,
  renameSymbol,
  surroundSelection,
} from './lib/ide-actions.js'
import { createDesignCompletionSource, createFalconCompletionSource } from './lib/ide-completions.js'
import { createHoverDocumentationExtension, createSignatureHelpExtension } from './lib/ide-intelligence.js'

function generateCode() {
  return String(Math.floor(10000 + Math.random() * 90000))
}

function getInitialTheme() {
  try {
    const stored = localStorage.getItem('tensor-theme')
    if (stored === 'light' || stored === 'dark') return stored
    return window.matchMedia?.('(prefers-color-scheme: dark)')?.matches ? 'dark' : 'light'
  } catch {
    return 'light'
  }
}

function handleCompanionResponse(data) {
  try {
    const resp = JSON.parse(data)
    if (resp.status !== 'OK') { console.error('[companion]', data); return }
    for (const v of resp.values ?? []) {
      if (v.status === 'OK') {
        if (v.value && v.value !== '*nothing*') console.log('[companion] →', v.value)
      } else {
        console.error('[companion] eval error:', v.value)
      }
    }
  } catch {
    console.warn('[companion] non-JSON:', data)
  }
}

function logCompanionPayload(label, result) {
  const screenList = result.screenIds?.length ? result.screenIds.join(', ') : '(none)'
  const eventList = result.eventDefs?.length
    ? result.eventDefs.map(({ component, event }) => `${component}.${event}`).join(', ')
    : '(none)'
  console.log(`[companion] ${label} screens in compiled YAIL: ${screenList}`)
  console.log(`[companion] ${label} events in compiled YAIL: ${eventList}`)
  console.log(`[companion] ${label} full YAIL (blocks + Screen/design code):\n${result.fullYail}`)
  console.log(`[companion] ${label} REPL payload sent:\n${result.replPayload}`)
}

function companionSourceKey(falconCode, designCode) {
  return `${falconCode}\0${designCode}`
}

function splitBlocklyXml(xml) {
  return String(xml ?? '').split('\0').map(chunk => chunk.trim()).filter(Boolean)
}

function buildBlocklyPreviewEntries(result) {
  const lineNumbers = result?.lineNumbers ?? []
  return splitBlocklyXml(result?.xml)
    .map((xml, index) => ({
      index,
      xml,
      line: Number(lineNumbers[index]),
    }))
    .filter(entry => Number.isFinite(entry.line))
    .sort((a, b) => a.line - b.line || a.index - b.index)
    .map((entry, index, entries) => ({
      ...entry,
      nextLine: entries[index + 1]?.line ?? null,
    }))
}

const SAMPLE_FALCON_CELLS = [
  {
    id: 'cell-check-text-boxes',
    title: 'Validate inputs',
    code: `func checkTextBoxes() = {
  if (!(firstNumberTextBox.Text ? number) || !(secondNumberTextBox.Text ? number)) {
    Notifier1.ShowAlert("Please enter numeric values in the textbox!")
    yield false
  }
  true
}`,
  },
  {
    id: 'cell-add-button',
    title: 'Add button',
    code: `when AddButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text + secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'cell-subtract-button',
    title: 'Subtract button',
    code: `when SubtractButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text - secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'cell-multiply-button',
    title: 'Multiply button',
    code: `when MultiplyButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text * secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'cell-divide-button',
    title: 'Divide button',
    code: `when DivideButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text / secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
]

const SAMPLE_DESIGN = `Screen.Screen1 { Title: "Calculator",
  Label { Text: "First number: " }
  TextBox.firstNumberTextBox { NumbersOnly: true, Hint: "Enter first number" }

  Label { Text: "Second number: " }
  TextBox.secondNumberTextBox { NumbersOnly: true, Hint: "Enter second number" }

  HorizontalArrangement {
    Button.AddButton { Text: "+" }
    Button.SubtractButton { Text: "-" }
    Button.MultiplyButton { Text: "*" }
    Button.DivideButton { Text: "/" }
  }

  Notifier.Notifier1
}
`

const VALIDATION_DELAY_MS = 1000

function createNotebookCell(afterTitle = 'New cell') {
  const suffix = `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}`
  return {
    id: `cell-${suffix}`,
    title: afterTitle,
    code: '',
  }
}

function combineFalconCells(cells) {
  const source = cells.map(cell => cell.code.trimEnd()).join('\n\n')
  return source ? `${source}\n` : ''
}

function cellLineCount(code) {
  return String(code ?? '').split('\n').length
}

function buildCellPositions(cells) {
  let charStart = 0
  let lineStart = 1
  return cells.map((cell, index) => {
    const code = cell.code ?? ''
    const lineCount = cellLineCount(code)
    const position = {
      id: cell.id,
      index,
      charStart,
      charEnd: charStart + code.length,
      lineStart,
      lineEnd: lineStart + lineCount - 1,
      lineCount,
    }
    charStart += code.length + (index < cells.length - 1 ? 2 : 1)
    lineStart += lineCount + (index < cells.length - 1 ? 2 : 1)
    return position
  })
}

function mapCombinedLineToCell(positions, line) {
  const lineNumber = Number(line) || 1
  return positions.find(position => lineNumber >= position.lineStart && lineNumber <= position.lineEnd)
    ?? positions[positions.length - 1]
}

function mapCombinedOffsetToCell(positions, offset) {
  const pos = Number(offset) || 0
  return positions.find(position => pos >= position.charStart && pos <= position.charEnd)
    ?? positions[positions.length - 1]
}

function splitCombinedAcrossCells(nextCode, previousCells) {
  const lines = String(nextCode ?? '').replace(/\n$/, '').split('\n')
  let lineCursor = 0
  return previousCells.map((cell, index) => {
    const count = cellLineCount(cell.code)
    const code = lines.slice(lineCursor, lineCursor + count).join('\n')
    lineCursor += count
    while (lineCursor < lines.length && lines[lineCursor] === '' && index < previousCells.length - 1) lineCursor++
    return { ...cell, code }
  })
}

function findFalconComponentAtCursor(code, line, col, designCode = '') {
  const lineText = code.split('\n')[line - 1] ?? ''
  const colIdx = col - 1
  const re = /@([A-Za-z]\w*)/g
  let m
  while ((m = re.exec(lineText)) !== null) {
    if (colIdx >= m.index && colIdx <= m.index + m[0].length) return m[1]
  }

  const componentDefs = extractComponentDefs(designCode)
  const componentById = {}
  for (const [type, ids] of Object.entries(componentDefs)) {
    for (const id of ids ?? []) componentById[id] = type
  }

  const wordRe = /\b[A-Za-z_]\w*\b/g
  while ((m = wordRe.exec(lineText)) !== null) {
    if (colIdx >= m.index && colIdx <= m.index + m[0].length) {
      return componentById[m[0]] ?? null
    }
  }
  return null
}

function findDesignComponentAtCursor(code, line, col) {
  const lineText = code.split('\n')[line - 1] ?? ''
  const colIdx = col - 1

  const oldRe = /@([A-Za-z]\w*)/g
  let m
  while ((m = oldRe.exec(lineText)) !== null) {
    if (colIdx >= m.index && colIdx <= m.index + m[0].length) return m[1]
  }

  const re = /\b([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*([A-Za-z_]\w*))?(?=\s*(?:\{|$))/g
  while ((m = re.exec(lineText)) !== null) {
    const start = m.index
    const end = m.index + m[0].length
    if (colIdx >= start && colIdx <= end) return m[1]
  }
  return null
}

function findPropertyAtCursor(code, line, col) {
  const lineText = code.split('\n')[line - 1] ?? ''
  const colIdx = col - 1
  const re = /([A-Za-z_]\w*)\s*:/g
  let m
  while ((m = re.exec(lineText)) !== null) {
    const start = m.index
    const end = m.index + m[1].length
    if (colIdx >= start && colIdx <= end) return m[1]
  }
  return null
}

function componentNameBeforeOpening(text, openingIndex) {
  const before = text.slice(0, openingIndex)
  const oldMatch = before.match(/@([A-Za-z]\w*)\s*$/)
  if (oldMatch) return oldMatch[1]
  const match = before.match(/([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*[A-Za-z_]\w*)?\s*$/)
  return match ? match[1] : null
}

function findEnclosingComponent(code, line, col) {
  const lines = code.split('\n')
  let pos = 0
  for (let i = 0; i < line - 1; i++) pos += lines[i].length + 1
  pos += col - 1
  const text = code.slice(0, pos)
  let depth = 0
  for (let i = text.length - 1; i >= 0; i--) {
    const ch = text[i]
    if (ch === '}') depth++
    else if (ch === '{') {
      if (depth === 0) {
        return componentNameBeforeOpening(text, i)
      }
      depth--
    }
  }
  return null
}

export default function App() {
  const [falconCells, setFalconCells] = useState(SAMPLE_FALCON_CELLS)
  const [designCode, setDesignCode]  = useState(SAMPLE_DESIGN)
  const [cursor, setCursor]          = useState({ line: 1, col: 1 })
  const [charCount, setCharCount]    = useState(SAMPLE_FALCON_CELLS[0].code.length)
  const [activePane, setActivePane]  = useState('falcon')
  const [activeCellId, setActiveCellId] = useState(SAMPLE_FALCON_CELLS[0].id)
  const [focusedPane, setFocusedPane] = useState(null)
  const [selectionEmpty, setSelectionEmpty] = useState(true)
  const [blocklyPreviewEntries, setBlocklyPreviewEntries] = useState([])
  const [selectedBlocklyPreview, setSelectedBlocklyPreview] = useState(null)
  const [theme, setTheme] = useState(getInitialTheme)
  const [blocklyMode, setBlocklyMode]     = useState(false)
  const [falconError, setFalconError]     = useState(null)
  const [designError, setDesignError]     = useState(null)
  const [falconDiagnostics, setFalconDiagnostics] = useState([])
  const [designDiagnostics, setDesignDiagnostics] = useState([])
  const [ideCatalog, setIdeCatalog]       = useState(null)
  const [idePanel, setIdePanel]           = useState(null)
  const [toast, setToast]                 = useState({ msg: '', visible: false })
  const [companionOpen, setCompanionOpen] = useState(false)
  const [companion, setCompanion]         = useState({ status: 'idle', code: null, error: null, messageCount: 0 })
  const falconEditorRefs = useRef(new Map())
  const designEditorRef = useRef(null)
  const toastTimer = useRef(null)
  const channelRef = useRef(null)
  const peerRef    = useRef(null)
  const lastSentSourceKeyRef = useRef(null)

  const falconCode = useMemo(() => combineFalconCells(falconCells), [falconCells])
  const cellPositions = useMemo(() => buildCellPositions(falconCells), [falconCells])
  const activeCell = useMemo(
    () => falconCells.find(cell => cell.id === activeCellId) ?? falconCells[0],
    [activeCellId, falconCells],
  )

  const showToast = useCallback((msg, ms = 2000) => {
    setToast({ msg, visible: true })
    clearTimeout(toastTimer.current)
    toastTimer.current = setTimeout(() => setToast(t => ({ ...t, visible: false })), ms)
  }, [])

  const setCellCode = useCallback((cellId, code) => {
    setFalconCells(cells => cells.map(cell => (
      cell.id === cellId ? { ...cell, code } : cell
    )))
  }, [])

  const setCellTitle = useCallback((cellId, title) => {
    setFalconCells(cells => cells.map(cell => (
      cell.id === cellId ? { ...cell, title } : cell
    )))
  }, [])

  const focusCell = useCallback((cellId) => {
    setActivePane('falcon')
    setActiveCellId(cellId)
    requestAnimationFrame(() => falconEditorRefs.current.get(cellId)?.focus())
  }, [])

  const addCellAfter = useCallback((cellId) => {
    const nextCell = createNotebookCell()
    setFalconCells(cells => {
      const index = Math.max(0, cells.findIndex(cell => cell.id === cellId))
      return [
        ...cells.slice(0, index + 1),
        nextCell,
        ...cells.slice(index + 1),
      ]
    })
    requestAnimationFrame(() => focusCell(nextCell.id))
  }, [focusCell])

  const deleteCell = useCallback((cellId) => {
    setFalconCells(cells => {
      if (cells.length <= 1) {
        showToast('Notebook needs at least one cell')
        return cells
      }
      const index = cells.findIndex(cell => cell.id === cellId)
      const nextCells = cells.filter(cell => cell.id !== cellId)
      const nextActive = nextCells[Math.max(0, Math.min(index, nextCells.length - 1))]
      requestAnimationFrame(() => focusCell(nextActive.id))
      return nextCells
    })
  }, [focusCell, showToast])

  const moveCell = useCallback((cellId, direction) => {
    setFalconCells(cells => {
      const index = cells.findIndex(cell => cell.id === cellId)
      const nextIndex = index + direction
      if (index < 0 || nextIndex < 0 || nextIndex >= cells.length) return cells
      const nextCells = [...cells]
      const [cell] = nextCells.splice(index, 1)
      nextCells.splice(nextIndex, 0, cell)
      return nextCells
    })
    focusCell(cellId)
  }, [focusCell])

  const renameCell = useCallback((cellId) => {
    const cell = falconCells.find(candidate => candidate.id === cellId)
    if (!cell) return
    const title = window.prompt('Cell title:', cell.title)
    if (title == null) return
    setCellTitle(cellId, title.trim() || cell.title)
  }, [falconCells, setCellTitle])

  const setFalconCodeFromCombined = useCallback((nextCode) => {
    setFalconCells(cells => splitCombinedAcrossCells(nextCode, cells))
  }, [])

  useEffect(() => {
    return () => clearTimeout(toastTimer.current)
  }, [])

  useEffect(() => {
    document.documentElement.dataset.theme = theme
    document.documentElement.style.colorScheme = theme
    try { localStorage.setItem('tensor-theme', theme) } catch {}
  }, [theme])

  useEffect(() => {
    let cancelled = false
    loadIdeCatalog()
      .then(catalog => {
        if (!cancelled) setIdeCatalog(catalog)
      })
      .catch(error => {
        console.warn('[ide] completion catalog unavailable:', error)
      })
    return () => { cancelled = true }
  }, [])

  const toggleTheme = useCallback(() => {
    setTheme(t => t === 'dark' ? 'light' : 'dark')
  }, [])

  const handleFalconCursor = useCallback((cellId, line, col, length, empty) => {
    setActivePane('falcon')
    setActiveCellId(cellId)
    setCursor({ line, col })
    setCharCount(length)
    setSelectionEmpty(empty)
  }, [])

  const handleDesignCursor = useCallback((line, col, length, empty) => {
    setActivePane('design')
    setCursor({ line, col })
    setCharCount(length)
    setSelectionEmpty(empty)
  }, [])

  const handleFalconFocus = useCallback((focused) => {
    setFocusedPane(prev => focused ? 'falcon' : (prev === 'falcon' ? null : prev))
  }, [])

  const handleDesignFocus = useCallback((focused) => {
    setFocusedPane(prev => focused ? 'design' : (prev === 'design' ? null : prev))
  }, [])

  const openCompanion = useCallback(() => {
    setCompanionOpen(true)
    setCompanion(c => {
      if (c.status === 'idle' || c.status === 'error') {
        return { status: 'polling', code: generateCode(), error: null, messageCount: 0 }
      }
      return c
    })
  }, [])

  const closeCompanion = useCallback(() => {
    setCompanionOpen(false)
    setCompanion(c => (
      c.status === 'connected' ? c : { status: 'idle', code: null, error: null, messageCount: 0 }
    ))
  }, [])

  const retryCompanion = useCallback(() => {
    lastSentSourceKeyRef.current = null
    setCompanion({ status: 'polling', code: generateCode(), error: null, messageCount: 0 })
  }, [])

  const disconnectCompanion = useCallback(() => {
    const ch = channelRef.current
    const pc = peerRef.current
    if (ch) { ch.onclose = null; ch.onerror = null; ch.onmessage = null; try { ch.close() } catch {} }
    if (pc) { try { pc.close() } catch {} }
    channelRef.current = null
    peerRef.current = null
    lastSentSourceKeyRef.current = null
    setCompanion({ status: 'idle', code: null, error: null, messageCount: 0 })
    setCompanionOpen(false)
    showToast('Companion disconnected')
  }, [showToast])

  // Debounced parse validation — surfaces errors in the status bar
  useEffect(() => {
    let cancelled = false
    const timer = setTimeout(async () => {
      const {
        falconError: fe,
        designError: de,
        falconDiagnostics: fd = [],
        designDiagnostics: dd = [],
      } = await validateSources(falconCode, designCode)
      if (!cancelled) {
        setFalconError(fe)
        setDesignError(de)
        setFalconDiagnostics(fd)
        setDesignDiagnostics(dd)
      }
    }, VALIDATION_DELAY_MS)
    return () => { cancelled = true; clearTimeout(timer) }
  }, [falconCode, designCode])

  // Debounced Blockly preview index — parses the full Falcon source so markers
  // use the same context as normal compilation.
  useEffect(() => {
    let cancelled = false
    const timer = setTimeout(async () => {
      try {
        const result = await mistToXmlResult(falconCode, extractComponentDefs(designCode))
        if (cancelled) return
        const entries = buildBlocklyPreviewEntries(result)
        setBlocklyPreviewEntries(entries)
        setSelectedBlocklyPreview(prev => {
          if (!prev) return null
          return entries.find(entry => entry.index === prev.index && entry.line === prev.line)
            ?? entries.find(entry => entry.line === prev.line)
            ?? null
        })
      } catch {
        if (!cancelled) {
          setBlocklyPreviewEntries([])
          setSelectedBlocklyPreview(null)
        }
      }
    }, 500)
    return () => { cancelled = true; clearTimeout(timer) }
  }, [falconCode, designCode])

  // Connection lifecycle — runs each time the code changes (new attempt)
  useEffect(() => {
    if (!companion.code) return
    if (companion.status === 'connected' || companion.status === 'error') return

    const abort = new AbortController()
    let alive = true
    let localPeer = null
    let localChannel = null

    ;(async () => {
      try {
        const result = await compileForCompanion(falconCode, designCode)
        if (!alive) return

        const { digest, config } = await pollRendezvous(companion.code, abort.signal)
        if (!alive) return

        setCompanion(c => ({ ...c, status: 'connecting' }))
        const conn = await connectCompanion(companion.code, digest, config, abort.signal)
        if (!alive) { try { conn.channel.close() } catch {} ; try { conn.peer.close() } catch {} ; return }

        localPeer = conn.peer
        localChannel = conn.channel
        channelRef.current = conn.channel
        peerRef.current    = conn.peer

        conn.channel.onmessage = ({ data }) => {
          setCompanion(c => ({ ...c, messageCount: c.messageCount + 1 }))
          handleCompanionResponse(data)
        }
        conn.channel.onerror = () => {
          setCompanion(c => ({ ...c, status: 'error', error: 'Data channel error' }))
        }
        conn.channel.onclose = () => {
          if (channelRef.current !== conn.channel) return
          channelRef.current = null
          peerRef.current = null
          lastSentSourceKeyRef.current = null
          try { conn.peer.close() } catch {}
          setCompanion(c => (c.status === 'error' ? c : { ...c, status: 'idle' }))
          showToast('Companion disconnected')
        }

        logCompanionPayload('initial send', result)
        conn.channel.send(result.replPayload)
        lastSentSourceKeyRef.current = companionSourceKey(falconCode, designCode)
        setCompanion(c => ({ ...c, status: 'connected' }))
        setCompanionOpen(false)
        showToast('Companion connected')
      } catch (e) {
        if (e.name === 'AbortError' || !alive) return
        console.error('[companion]', e)
        setCompanion(c => ({ ...c, status: 'error', error: e.message || String(e) }))
      }
    })()

    return () => {
      alive = false
      abort.abort()
      if (localChannel) {
        localChannel.onclose = null
        localChannel.onerror = null
        localChannel.onmessage = null
      }
      if (localChannel) try { localChannel.close() } catch {}
      if (localPeer)    try { localPeer.close() }    catch {}
      if (channelRef.current === localChannel) channelRef.current = null
      if (peerRef.current === localPeer)       peerRef.current    = null
    }
  }, [companion.code]) // eslint-disable-line react-hooks/exhaustive-deps

  // Auto-refresh — debounced recompile + send when sources change while connected
  useEffect(() => {
    if (companion.status !== 'connected') return
    const sourceKey = companionSourceKey(falconCode, designCode)
    if (lastSentSourceKeyRef.current === sourceKey) return
    const timer = setTimeout(async () => {
      const ch = channelRef.current
      if (!ch || ch.readyState !== 'open') return
      try {
        const result = await compileForCompanion(falconCode, designCode)
        if (channelRef.current === ch && ch.readyState === 'open') {
          logCompanionPayload('auto-refresh', result)
          ch.send(result.replPayload)
          lastSentSourceKeyRef.current = sourceKey
        }
      } catch (e) {
        console.error('[companion] auto-refresh failed:', e.message)
      }
    }, 600)
    return () => clearTimeout(timer)
  }, [falconCode, designCode, companion.status])

  const popupActive = focusedPane !== null && selectionEmpty

  const falconComponentAtCursor = useMemo(
    () => (popupActive && focusedPane === 'falcon' ? findFalconComponentAtCursor(falconCode, cursor.line, cursor.col, designCode) : null),
    [popupActive, focusedPane, falconCode, designCode, cursor.line, cursor.col],
  )
  const designComponentAtCursor = useMemo(
    () => (popupActive && focusedPane === 'design' ? findDesignComponentAtCursor(designCode, cursor.line, cursor.col) : null),
    [popupActive, focusedPane, designCode, cursor.line, cursor.col],
  )

  const designPropertyAtCursor = useMemo(() => {
    if (!popupActive || focusedPane !== 'design' || designComponentAtCursor) return null
    const propName = findPropertyAtCursor(designCode, cursor.line, cursor.col)
    if (!propName) return null
    const compName = findEnclosingComponent(designCode, cursor.line, cursor.col)
    return compName ? { propName, compName } : null
  }, [popupActive, focusedPane, designComponentAtCursor, designCode, cursor.line, cursor.col])

  const blocklyPreviewEntriesByCell = useMemo(() => {
    const grouped = new Map()
    for (const entry of blocklyPreviewEntries) {
      const position = mapCombinedLineToCell(cellPositions, entry.line)
      if (!position) continue
      const localEntry = {
        ...entry,
        combinedLine: entry.line,
        line: entry.line - position.lineStart + 1,
        nextLine: Number.isFinite(entry.nextLine) && entry.nextLine <= position.lineEnd
          ? entry.nextLine - position.lineStart + 1
          : null,
      }
      const entries = grouped.get(position.id) ?? []
      entries.push(localEntry)
      grouped.set(position.id, entries)
    }
    return grouped
  }, [blocklyPreviewEntries, cellPositions])

  const falconDiagnosticsByCell = useMemo(() => {
    const grouped = new Map()
    for (const diagnostic of falconDiagnostics) {
      const position = mapCombinedLineToCell(cellPositions, diagnostic.line)
      if (!position) continue
      const localDiagnostic = {
        ...diagnostic,
        combinedLine: diagnostic.line,
        line: Math.max(1, diagnostic.line - position.lineStart + 1),
      }
      const diagnostics = grouped.get(position.id) ?? []
      diagnostics.push(localDiagnostic)
      grouped.set(position.id, diagnostics)
    }
    return grouped
  }, [falconDiagnostics, cellPositions])

  const falconCompletionSource = useMemo(
    () => createFalconCompletionSource({ catalog: ideCatalog, designCode, projectCode: falconCode }),
    [ideCatalog, designCode, falconCode],
  )

  const designCompletionSource = useMemo(
    () => createDesignCompletionSource({ catalog: ideCatalog }),
    [ideCatalog],
  )

  const falconIntelligenceExtensions = useMemo(
    () => [
      createSignatureHelpExtension({ catalog: ideCatalog, designCode, mode: 'falcon' }),
      createHoverDocumentationExtension({ catalog: ideCatalog, designCode, mode: 'falcon' }),
    ],
    [ideCatalog, designCode],
  )

  const designIntelligenceExtensions = useMemo(
    () => [
      createHoverDocumentationExtension({ catalog: ideCatalog, mode: 'design' }),
    ],
    [ideCatalog],
  )

  const diagnosticsCount = falconDiagnostics.length + designDiagnostics.length
  const activeBreadcrumb = useMemo(() => {
    if (activePane === 'falcon') {
      const position = cellPositions.find(cell => cell.id === activeCellId)
      const cell = falconCells.find(candidate => candidate.id === activeCellId)
      if (!position || !cell) return 'Notebook'
      const localOffset = lineColToOffset(cell.code, cursor.line, cursor.col)
      return `${cell.title} > ${breadcrumbsForPosition(falconCode, 'falcon', position.charStart + localOffset)}`
    }
    const pos = lineColToOffset(designCode, cursor.line, cursor.col)
    return breadcrumbsForPosition(designCode, activePane, pos)
  }, [activePane, activeCellId, cellPositions, cursor.line, cursor.col, falconCells, falconCode, designCode])

  const goToLocation = useCallback((location) => {
    if (!location) return false
    const pane = location.pane === 'design' ? 'design' : 'falcon'
    const code = pane === 'falcon' ? falconCode : designCode
    const pos = Number.isFinite(location.from)
      ? location.from
      : lineColToOffset(code, location.line, location.col)
    setActivePane(pane)
    if (pane === 'falcon') {
      const position = mapCombinedOffsetToCell(cellPositions, pos)
      if (!position) return false
      setActiveCellId(position.id)
      requestAnimationFrame(() => {
        falconEditorRefs.current.get(position.id)?.goTo(Math.max(0, pos - position.charStart))
      })
    } else {
      requestAnimationFrame(() => designEditorRef.current?.goTo(pos))
    }
    return true
  }, [cellPositions, falconCode, designCode])

  const getActionContext = useCallback(() => {
    const pane = focusedPane ?? activePane
    const cell = falconCells.find(candidate => candidate.id === activeCellId) ?? falconCells[0]
    const position = cellPositions.find(candidate => candidate.id === cell?.id)
    const editor = pane === 'falcon' ? falconEditorRefs.current.get(cell?.id) : designEditorRef.current
    const view = editor?.getView()
    const code = pane === 'falcon' ? (cell?.code ?? '') : designCode
    const fallbackPos = lineColToOffset(code, cursor.line, cursor.col)
    const selection = view?.state.selection.main ?? { from: fallbackPos, to: fallbackPos, head: fallbackPos, empty: true }
    const pos = selection.head ?? selection.from
    return {
      pane,
      cellId: cell?.id,
      cell,
      editor,
      view,
      code,
      pos,
      combinedPos: pane === 'falcon' && position ? position.charStart + pos : pos,
      selection,
      cellPosition: position,
    }
  }, [activeCellId, activePane, cellPositions, cursor.line, cursor.col, designCode, falconCells, focusedPane])

  const applyCodeChange = useCallback((pane, from, to, insert, selectTo = null) => {
    if (pane === 'falcon') {
      const position = mapCombinedOffsetToCell(cellPositions, from)
      if (!position || position.id !== mapCombinedOffsetToCell(cellPositions, to)?.id) return
      const cell = falconCells.find(candidate => candidate.id === position.id)
      if (!cell) return
      const localFrom = Math.max(0, from - position.charStart)
      const localTo = Math.max(localFrom, to - position.charStart)
      const next = cell.code.slice(0, localFrom) + insert + cell.code.slice(localTo)
      setCellCode(position.id, next)
      requestAnimationFrame(() => {
        falconEditorRefs.current.get(position.id)?.goTo(
          selectTo == null ? localFrom + String(insert ?? '').length : Math.max(0, selectTo - position.charStart),
        )
      })
      return
    }

    const next = designCode.slice(0, from) + insert + designCode.slice(to)
    setDesignCode(next)
    requestAnimationFrame(() => designEditorRef.current?.goTo(selectTo ?? from + String(insert ?? '').length))
  }, [cellPositions, designCode, falconCells, setCellCode])

  const handleFormat = useCallback(() => {
    const { pane, view, code, selection, cellId } = getActionContext()
    const mode = pane === 'design' ? 'design' : 'falcon'

    if (view && selection && !selection.empty) {
      const fromLine = view.state.doc.lineAt(selection.from)
      const toLine = view.state.doc.lineAt(selection.to)
      const to = selection.to === toLine.from && toLine.number > fromLine.number
        ? view.state.doc.line(toLine.number - 1).to
        : toLine.to
      const formatted = formatSource(code.slice(fromLine.from, to), mode).trimEnd()
      if (pane === 'falcon') {
        setCellCode(cellId, code.slice(0, fromLine.from) + formatted + code.slice(to))
        requestAnimationFrame(() => falconEditorRefs.current.get(cellId)?.goTo(fromLine.from))
      } else {
        applyCodeChange(pane, fromLine.from, to, formatted, fromLine.from)
      }
      showToast('Selection formatted')
      return
    }

    const formatted = formatSource(code, mode)
    if (pane === 'falcon') setCellCode(activeCellId, formatted.trimEnd())
    else setDesignCode(formatted)
    showToast(`${pane === 'falcon' ? 'Cell' : 'Design'} formatted`)
  }, [activeCellId, applyCodeChange, getActionContext, setCellCode, showToast])

  const handleGoToDefinition = useCallback((pane, pos, cellId = activeCellId) => {
    const position = pane === 'falcon'
      ? cellPositions.find(candidate => candidate.id === cellId)
      : null
    const combinedPos = pane === 'falcon' && position ? position.charStart + pos : pos
    const location = findDefinition({ falconCode, designCode, pane, pos: combinedPos, catalog: ideCatalog })
    if (!location) {
      showToast('No local definition found')
      return true
    }
    goToLocation(location)
    return true
  }, [activeCellId, cellPositions, falconCode, designCode, ideCatalog, goToLocation, showToast])

  const handleFindUsages = useCallback(() => {
    const { pane, pos, combinedPos } = getActionContext()
    const { symbol, usages } = findUsages({ falconCode, designCode, pane, pos: pane === 'falcon' ? combinedPos : pos, catalog: ideCatalog })
    if (!symbol || !usages.length) {
      showToast('No usages found')
      return
    }
    setIdePanel({
      type: 'search',
      title: `Usages of ${symbol.name}`,
      subtitle: `${usages.length} occurrence${usages.length === 1 ? '' : 's'}`,
      items: usages.map(usage => ({
        label: usage.excerpt || usage.name,
        detail: usage.kind,
        pane: usage.pane,
        line: usage.line,
        col: usage.col,
        onSelect: () => goToLocation(usage),
      })),
    })
  }, [falconCode, designCode, ideCatalog, getActionContext, goToLocation, showToast])

  const handleRename = useCallback(() => {
    const { pane, pos, combinedPos } = getActionContext()
    const targetPos = pane === 'falcon' ? combinedPos : pos
    const { symbol } = findUsages({ falconCode, designCode, pane, pos: targetPos, catalog: ideCatalog })
    if (!symbol) {
      showToast('No renameable symbol at cursor')
      return
    }
    const nextName = window.prompt(`Rename ${symbol.name} to:`, symbol.name)
    if (nextName == null) return
    const result = renameSymbol({ falconCode, designCode, pane, pos: targetPos, catalog: ideCatalog, newName: nextName })
    if (!result.ok) {
      showToast(result.error)
      return
    }
    setFalconCodeFromCombined(result.falconCode)
    setDesignCode(result.designCode)
    setIdePanel(null)
    showToast(`Renamed ${result.count} reference${result.count === 1 ? '' : 's'}`)
  }, [falconCode, designCode, ideCatalog, getActionContext, setFalconCodeFromCombined, showToast])

  const handleQuickFix = useCallback(() => {
    const { pane, pos, combinedPos } = getActionContext()
    const diagnostics = pane === 'falcon' ? falconDiagnostics : designDiagnostics
    const fixes = quickFixesAt({
      falconCode,
      designCode,
      pane,
      pos: pane === 'falcon' ? combinedPos : pos,
      diagnostics,
      catalog: ideCatalog,
    })
    if (!fixes.length) {
      showToast('No quick fixes available')
      return
    }
    setIdePanel({
      type: 'fix',
      title: 'Quick Fixes',
      subtitle: `${fixes.length} available`,
      items: fixes.map(fix => ({
        label: fix.title,
        detail: fix.pane === 'falcon' ? 'Screen1.falcon' : 'Screen1.design',
        pane: fix.pane,
        onSelect: () => {
          applyCodeChange(fix.pane, fix.from, fix.to, fix.insert)
          setIdePanel(null)
          showToast('Quick fix applied')
        },
      })),
    })
  }, [applyCodeChange, designCode, designDiagnostics, falconCode, falconDiagnostics, getActionContext, ideCatalog, showToast])

  const handleProjectSearch = useCallback(() => {
    const query = window.prompt('Search project:')
    if (!query) return
    const results = projectSearch(falconCode, designCode, query)
    setIdePanel({
      type: 'search',
      title: `Search: ${query}`,
      subtitle: `${results.length} match${results.length === 1 ? '' : 'es'}`,
      items: results.map(result => ({
        label: result.excerpt || result.name,
        detail: 'match',
        pane: result.pane,
        line: result.line,
        col: result.col,
        onSelect: () => goToLocation(result),
      })),
    })
  }, [falconCode, designCode, goToLocation])

  const handleSymbols = useCallback(() => {
    const symbols = documentSymbols(falconCode, designCode)
    setIdePanel({
      type: 'symbols',
      title: 'Project Symbols',
      subtitle: `${symbols.length} symbols`,
      items: symbols.map(symbol => ({
        label: symbol.label,
        detail: symbol.detail,
        pane: symbol.pane,
        line: symbol.line,
        col: symbol.col,
        onSelect: () => goToLocation(symbol),
      })),
    })
  }, [falconCode, designCode, goToLocation])

  const handleNavigateDiagnostic = useCallback((direction = 1) => {
    const locations = diagnosticLocations(falconCode, designCode, falconDiagnostics, designDiagnostics)
    if (!locations.length) {
      showToast('No errors')
      return
    }
    const activePosition = cellPositions.find(position => position.id === activeCellId)
    const activeCode = activePane === 'falcon'
      ? (activeCell?.code ?? '')
      : designCode
    const currentFrom = activePane === 'falcon' && activePosition
      ? activePosition.charStart + lineColToOffset(activeCode, cursor.line, cursor.col)
      : lineColToOffset(activeCode, cursor.line, cursor.col)
    const paneOrder = pane => pane === 'falcon' ? 0 : 1
    const currentKey = [paneOrder(activePane), currentFrom]
    const afterCurrent = loc => {
      const key = [paneOrder(loc.pane), loc.from]
      return key[0] > currentKey[0] || (key[0] === currentKey[0] && key[1] > currentKey[1])
    }
    const beforeCurrent = loc => {
      const key = [paneOrder(loc.pane), loc.from]
      return key[0] < currentKey[0] || (key[0] === currentKey[0] && key[1] < currentKey[1])
    }
    const target = direction > 0
      ? locations.find(afterCurrent) ?? locations[0]
      : [...locations].reverse().find(beforeCurrent) ?? locations[locations.length - 1]
    goToLocation(target)
    showToast(target.message)
  }, [activeCell, activeCellId, activePane, cellPositions, cursor.line, cursor.col, designCode, designDiagnostics, falconCode, falconDiagnostics, goToLocation, showToast])

  const handleExtractFunction = useCallback(() => {
    const { pane, selection, code, cellId } = getActionContext()
    if (pane !== 'falcon') {
      showToast('Extract function is available in Falcon code')
      return
    }
    if (!selection || selection.empty) {
      showToast('Select statements to extract')
      return
    }
    const name = window.prompt('New function name:', 'extractedFunction')
    if (!name) return
    const result = extractFunction(code, selection.from, selection.to, name)
    if (!result.ok) {
      showToast(result.error)
      return
    }
    setCellCode(cellId, result.code)
    requestAnimationFrame(() => falconEditorRefs.current.get(cellId)?.goTo(result.replacementOffset))
    showToast(`Extracted ${name}()`)
  }, [getActionContext, setCellCode, showToast])

  const handleSurroundWith = useCallback(() => {
    const { pane, selection, code, cellId } = getActionContext()
    if (pane !== 'falcon') {
      showToast('Surround-with is available in Falcon code')
      return
    }
    if (!selection || selection.empty) {
      showToast('Select code to surround')
      return
    }
    const kind = window.prompt('Surround with: if, while, for, function', 'if')
    if (!kind) return
    const result = surroundSelection(code, selection.from, selection.to, kind.trim())
    if (!result.ok) {
      showToast(result.error)
      return
    }
    setCellCode(cellId, result.code)
    requestAnimationFrame(() => falconEditorRefs.current.get(cellId)?.goTo(result.from))
    showToast(`Surrounded with ${kind}`)
  }, [getActionContext, setCellCode, showToast])

  useEffect(() => {
    const onKeyDown = (event) => {
      const key = event.key.toLowerCase()
      if (event.key === 'F12') {
        event.preventDefault()
        const { pane, pos } = getActionContext()
        handleGoToDefinition(pane, pos)
      } else if (event.key === 'F2') {
        event.preventDefault()
        handleRename()
      } else if (event.key === 'F8') {
        event.preventDefault()
        handleNavigateDiagnostic(event.shiftKey ? -1 : 1)
      } else if (event.altKey && event.key === 'Enter') {
        event.preventDefault()
        handleQuickFix()
      } else if (event.altKey && event.key === 'F7') {
        event.preventDefault()
        handleFindUsages()
      } else if (event.altKey && event.shiftKey && key === 'f') {
        event.preventDefault()
        handleFormat()
      } else if ((event.ctrlKey || event.metaKey) && event.shiftKey && key === 'f') {
        event.preventDefault()
        handleProjectSearch()
      } else if ((event.ctrlKey || event.metaKey) && event.shiftKey && key === 'o') {
        event.preventDefault()
        handleSymbols()
      } else if (event.ctrlKey && event.altKey && key === 'm') {
        event.preventDefault()
        handleExtractFunction()
      } else if (event.ctrlKey && event.altKey && key === 't') {
        event.preventDefault()
        handleSurroundWith()
      }
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [
    getActionContext,
    handleExtractFunction,
    handleFindUsages,
    handleFormat,
    handleGoToDefinition,
    handleNavigateDiagnostic,
    handleProjectSearch,
    handleQuickFix,
    handleRename,
    handleSurroundWith,
    handleSymbols,
  ])

  const handleBlocklyPreviewLineHover = useCallback((cellId, line) => {
    if (line == null) {
      setSelectedBlocklyPreview(null)
      return
    }
    const position = cellPositions.find(candidate => candidate.id === cellId)
    if (!position) return
    const combinedLine = position.lineStart + line - 1
    const entry = blocklyPreviewEntries.find(candidate => candidate.line === combinedLine)
    if (!entry) return
    setSelectedBlocklyPreview(entry)
  }, [blocklyPreviewEntries, cellPositions])

  return (
    <div className="app">
      <Toolbar
        onCompanion={openCompanion}
        companionStatus={companionOpen && companion.status === 'polling' ? 'idle' : companion.status}
        theme={theme}
        onToggleTheme={toggleTheme}
        onFormat={handleFormat}
        onSearch={handleProjectSearch}
        onSymbols={handleSymbols}
        onQuickFix={handleQuickFix}
      />

      <div className="notebook-workspace">
        <main className="notebook-pane">
          <div className={`pane-header notebook-header${activePane === 'falcon' ? ' active' : ''}`}>
            <span>Screen1.falcon notebook</span>
            <span className="notebook-count">{falconCells.length} cells</span>
            <button type="button" className="notebook-add-top" onClick={() => addCellAfter(falconCells[falconCells.length - 1]?.id)}>
              <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                <path d="M12 5v14M5 12h14" />
              </svg>
              Cell
            </button>
            <label className="blockly-switch-group" title={blocklyMode ? 'Switch to code view' : 'Switch to blocks view'}>
              {blocklyMode ? (
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                  <path d="M4 7h3a1 1 0 0 0 1-1V5a2 2 0 0 1 4 0v1a1 1 0 0 0 1 1h3a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1a2 2 0 0 0 0 4h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-3a1 1 0 0 1-1-1v-1a2 2 0 0 0-4 0v1a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h1a2 2 0 0 0 0-4H4a1 1 0 0 1-1-1V8a1 1 0 0 1 1-1z"/>
                </svg>
              ) : (
                <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                  <polyline points="16 18 22 12 16 6"/>
                  <polyline points="8 6 2 12 8 18"/>
                </svg>
              )}
              <span className="blockly-switch-label">{blocklyMode ? 'Blocks' : 'Code'}</span>
              <button
                type="button"
                className="blockly-switch"
                role="switch"
                aria-checked={blocklyMode}
                aria-label={blocklyMode ? 'Switch to code view' : 'Switch to blocks view'}
                onClick={() => setBlocklyMode(m => !m)}
              />
            </label>
          </div>

          <div className="notebook-scroll">
            {falconCells.map((cell, index) => {
              const entries = blocklyPreviewEntriesByCell.get(cell.id) ?? []
              const selectedPosition = selectedBlocklyPreview
                ? mapCombinedLineToCell(cellPositions, selectedBlocklyPreview.line)
                : null
              const activePreview = selectedPosition?.id === cell.id
                ? {
                    ...selectedBlocklyPreview,
                    combinedLine: selectedBlocklyPreview.line,
                    line: selectedBlocklyPreview.line - selectedPosition.lineStart + 1,
                    nextLine: Number.isFinite(selectedBlocklyPreview.nextLine) && selectedBlocklyPreview.nextLine <= selectedPosition.lineEnd
                      ? selectedBlocklyPreview.nextLine - selectedPosition.lineStart + 1
                      : null,
                  }
                : null
              const editorHeight = Math.max(118, Math.min(360, (cellLineCount(cell.code) + 2) * 22))

              return (
                <section className={`notebook-cell${cell.id === activeCellId ? ' active' : ''}`} key={cell.id}>
                  <div className="notebook-cell-rail">
                    <span className="notebook-cell-number">{index + 1}</span>
                  </div>
                  <div className="notebook-cell-main">
                    <div className="notebook-cell-header">
                      <button type="button" className="notebook-cell-title" onClick={() => renameCell(cell.id)}>
                        {cell.title}
                      </button>
                      <div className="notebook-cell-actions">
                        <button type="button" onClick={() => moveCell(cell.id, -1)} disabled={index === 0} aria-label="Move cell up" title="Move cell up">
                          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                            <path d="m18 15-6-6-6 6" />
                          </svg>
                        </button>
                        <button type="button" onClick={() => moveCell(cell.id, 1)} disabled={index === falconCells.length - 1} aria-label="Move cell down" title="Move cell down">
                          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                            <path d="m6 9 6 6 6-6" />
                          </svg>
                        </button>
                        <button type="button" onClick={() => addCellAfter(cell.id)} aria-label="Add cell below" title="Add cell below">
                          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                            <path d="M12 5v14M5 12h14" />
                          </svg>
                        </button>
                        <button type="button" onClick={() => deleteCell(cell.id)} disabled={falconCells.length <= 1} aria-label="Delete cell" title="Delete cell">
                          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                            <path d="M3 6h18" />
                            <path d="M8 6V4h8v2" />
                            <path d="M19 6 18 20H6L5 6" />
                          </svg>
                        </button>
                      </div>
                    </div>
                    <div className="notebook-cell-editor" style={{ height: `${editorHeight}px` }}>
                      <Editor
                        ref={node => {
                          if (node) falconEditorRefs.current.set(cell.id, node)
                          else falconEditorRefs.current.delete(cell.id)
                        }}
                        value={cell.code}
                        onChange={code => setCellCode(cell.id, code)}
                        onCursorChange={(line, col, length, empty) => handleFalconCursor(cell.id, line, col, length, empty)}
                        onFocusChange={handleFalconFocus}
                        extensions={falconExtensions}
                        blockPreviewLines={entries.map(entry => entry.line)}
                        activeBlockPreviewLine={activePreview?.line ?? null}
                        activeBlockPreview={activePreview}
                        onBlockPreviewLineHover={line => handleBlocklyPreviewLineHover(cell.id, line)}
                        showBlockPreviewGutter
                        blocklyMode={blocklyMode}
                        allBlockPreviews={entries}
                        diagnostics={falconDiagnosticsByCell.get(cell.id) ?? []}
                        completionSource={falconCompletionSource}
                        extraExtensions={falconIntelligenceExtensions}
                        onGoToDefinition={pos => handleGoToDefinition('falcon', pos, cell.id)}
                      />
                    </div>
                  </div>
                </section>
              )
            })}
          </div>
          <ComponentPopup name={falconComponentAtCursor} side="left" />
        </main>

        <aside className="design-pane">
          <div className={`pane-header${activePane === 'design' ? ' active' : ''}`}>Screen1.design</div>
          <div className="pane-editor">
            <Editor
              ref={designEditorRef}
              value={designCode}
              onChange={setDesignCode}
              onCursorChange={handleDesignCursor}
              onFocusChange={handleDesignFocus}
              extensions={designExtensions}
              diagnostics={designDiagnostics}
              completionSource={designCompletionSource}
              extraExtensions={designIntelligenceExtensions}
              onGoToDefinition={pos => handleGoToDefinition('design', pos)}
            />
            {designPropertyAtCursor
              ? <PropertyPopup propName={designPropertyAtCursor.propName} compName={designPropertyAtCursor.compName} side="right" />
              : <ComponentPopup name={designComponentAtCursor} side="right" />
            }
          </div>
        </aside>
      </div>

      <StatusBar
        line={cursor.line}
        col={cursor.col}
        charCount={charCount}
        lang={activePane === 'falcon' ? 'Falcon' : 'Design'}
        error={activePane === 'falcon' ? falconError : designError}
        blockCount={activePane === 'falcon' ? blocklyPreviewEntries.length : 0}
        diagnosticsCount={diagnosticsCount}
        breadcrumb={activeBreadcrumb}
      />

      <IdePanel panel={idePanel} onClose={() => setIdePanel(null)} />

      <div
        role="status"
        aria-live="polite"
        className={`toast${toast.visible ? ' show' : ''}`}
      >
        {toast.msg}
      </div>

      <CompanionModal
        open={companionOpen}
        status={companion.status}
        code={companion.code}
        error={companion.error}
        messageCount={companion.messageCount}
        theme={theme}
        onClose={closeCompanion}
        onRetry={retryCompanion}
        onDisconnect={disconnectCompanion}
      />
    </div>
  )
}
