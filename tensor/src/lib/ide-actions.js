const IDENT_RE = /^[A-Za-z_]\w*$/
const WORD_CHAR_RE = /[A-Za-z0-9_]/

export function maskIgnoredSpans(text) {
  const chars = String(text ?? '').split('')
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
      if (ch === '\n') continue
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

export function offsetToLineCol(source, offset) {
  const text = String(source ?? '')
  const safeOffset = Math.max(0, Math.min(text.length, Number(offset) || 0))
  let line = 1
  let lineStart = 0
  for (let i = 0; i < safeOffset; i++) {
    if (text[i] === '\n') {
      line++
      lineStart = i + 1
    }
  }
  return { line, col: safeOffset - lineStart + 1 }
}

export function lineColToOffset(source, line, col) {
  const lines = String(source ?? '').split('\n')
  const lineIndex = Math.max(0, Math.min(lines.length - 1, (Number(line) || 1) - 1))
  let offset = 0
  for (let i = 0; i < lineIndex; i++) offset += lines[i].length + 1
  return offset + Math.max(0, Math.min(lines[lineIndex].length, (Number(col) || 1) - 1))
}

export function wordAt(source, pos) {
  const text = String(source ?? '')
  const safePos = Math.max(0, Math.min(text.length, Number(pos) || 0))
  let from = safePos
  let to = safePos
  if (from === to && from > 0 && WORD_CHAR_RE.test(text[from - 1]) && !WORD_CHAR_RE.test(text[from] ?? '')) {
    from--
    to--
  }
  while (from > 0 && WORD_CHAR_RE.test(text[from - 1])) from--
  while (to < text.length && WORD_CHAR_RE.test(text[to])) to++
  if (from === to) return null
  return { from, to, text: text.slice(from, to) }
}

function escapeRegExp(value) {
  return String(value).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function stripHtml(value = '') {
  return String(value)
    .replace(/<[^>]*>/g, ' ')
    .replace(/&quot;/g, '"')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/\s+/g, ' ')
    .trim()
}

function splitParams(params) {
  if (Array.isArray(params)) return params.map(param => param?.name ?? param).filter(Boolean)
  return String(params ?? '').split(',').map(part => part.trim()).filter(Boolean)
}

function paramsForCount(paramCount) {
  const count = Number(paramCount)
  if (!Number.isFinite(count)) return []
  if (count >= 0) return Array.from({ length: count }, (_, index) => `arg${index + 1}`)
  const min = Math.max(0, -count - 1)
  return [
    ...Array.from({ length: min }, (_, index) => `arg${index + 1}`),
    'args...',
  ]
}

function findRangesOutsideIgnored(source, re, buildRange) {
  const searchable = maskIgnoredSpans(source)
  const ranges = []
  let match
  while ((match = re.exec(searchable)) !== null) {
    const range = buildRange(match)
    if (range && range.to > range.from) ranges.push(range)
    if (match[0] === '') re.lastIndex++
  }
  return ranges
}

function findWordRanges(source, word) {
  if (!word) return []
  return findRangesOutsideIgnored(
    source,
    new RegExp(`\\b${escapeRegExp(word)}\\b`, 'g'),
    match => ({ from: match.index, to: match.index + match[0].length }),
  )
}

function findCallNameRanges(source, name) {
  if (!name) return []
  return findRangesOutsideIgnored(
    source,
    new RegExp(`\\b${escapeRegExp(name)}\\b(?=\\s*\\()`, 'g'),
    match => ({ from: match.index, to: match.index + match[0].length }),
  )
}

export function parseFalconSymbols(source) {
  const text = String(source ?? '')
  const functions = []
  const events = []
  const variables = []
  const searchable = maskIgnoredSpans(text)

  let match
  const funcRe = /\bfunc\s+([A-Za-z_]\w*)\s*\(([^)]*)\)/g
  while ((match = funcRe.exec(searchable)) !== null) {
    const name = match[1]
    const from = match.index + match[0].indexOf(name)
    functions.push({
      kind: 'function',
      name,
      params: splitParams(match[2]),
      pane: 'falcon',
      from,
      to: from + name.length,
      ...offsetToLineCol(text, from),
    })
    for (const param of splitParams(match[2])) {
      const paramFrom = match.index + match[0].indexOf(param)
      if (paramFrom >= match.index) {
        variables.push({
          kind: 'parameter',
          name: param,
          pane: 'falcon',
          from: paramFrom,
          to: paramFrom + param.length,
          ...offsetToLineCol(text, paramFrom),
        })
      }
    }
  }

  const eventRe = /\bwhen\s+([A-Za-z_]\w*)\s*\.\s*([A-Za-z_]\w*)\s*(?:\(([^)]*)\))?/g
  while ((match = eventRe.exec(searchable)) !== null) {
    const name = `${match[1]}.${match[2]}`
    const from = match.index + match[0].indexOf(match[1])
    events.push({
      kind: 'event',
      name,
      component: match[1],
      event: match[2],
      params: splitParams(match[3]),
      pane: 'falcon',
      from,
      to: from + name.length,
      ...offsetToLineCol(text, from),
    })
    for (const param of splitParams(match[3])) {
      const paramFrom = match.index + match[0].indexOf(param)
      if (paramFrom >= match.index) {
        variables.push({
          kind: 'parameter',
          name: param,
          pane: 'falcon',
          from: paramFrom,
          to: paramFrom + param.length,
          ...offsetToLineCol(text, paramFrom),
        })
      }
    }
  }

  const variableRe = /\b(?:local|global)\s+([A-Za-z_]\w*)/g
  while ((match = variableRe.exec(searchable)) !== null) {
    const name = match[1]
    const from = match.index + match[0].indexOf(name)
    variables.push({
      kind: match[0].trimStart().startsWith('global') ? 'global' : 'local',
      name,
      pane: 'falcon',
      from,
      to: from + name.length,
      ...offsetToLineCol(text, from),
    })
  }

  return { functions, events, variables }
}

export function parseDesignSymbols(source) {
  const text = String(source ?? '')
  const searchable = maskIgnoredSpans(text)
  const components = []
  const legacyRe = /@([A-Za-z]\w*)\s*\{(?:[^{}@]*?[,\n])?\s*id\s*:\s*"([^"]+)"/g
  let match

  while ((match = legacyRe.exec(searchable)) !== null) {
    const type = match[1]
    const id = match[2]
    const typeFrom = match.index + 1
    const idFrom = match.index + match[0].lastIndexOf(id)
    components.push({
      kind: 'component',
      type,
      id,
      name: id,
      pane: 'design',
      from: idFrom,
      to: idFrom + id.length,
      typeFrom,
      typeTo: typeFrom + type.length,
      ...offsetToLineCol(text, idFrom),
    })
  }

  const dotRe = /(?:^|[{}\n])\s*([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*([A-Za-z_]\w*))?\s*(?=\{|$)/gm
  while ((match = dotRe.exec(searchable)) !== null) {
    const type = match[1]
    const id = match[2] ?? null
    const typeFrom = match.index + match[0].indexOf(type)
    const idFrom = id ? match.index + match[0].lastIndexOf(id) : null
    components.push({
      kind: 'component',
      type,
      id,
      name: id || type,
      pane: 'design',
      from: idFrom ?? typeFrom,
      to: (idFrom ?? typeFrom) + (id || type).length,
      typeFrom,
      typeTo: typeFrom + type.length,
      ...offsetToLineCol(text, idFrom ?? typeFrom),
    })
  }

  components.sort((a, b) => a.from - b.from)
  return { components }
}

export function buildProjectIndex(falconCode, designCode, catalog) {
  const falcon = parseFalconSymbols(falconCode)
  const design = parseDesignSymbols(designCode)
  const componentById = new Map()
  for (const component of design.components) {
    if (component.id) componentById.set(component.id, component)
  }
  return {
    falcon,
    design,
    componentById,
    catalog: {
      functions: catalog?.functions ?? [],
      methods: catalog?.methods ?? [],
      componentNames: catalog?.componentNames ?? [],
      components: catalog?.components ?? {},
    },
  }
}

function currentLineText(source, pos) {
  const text = String(source ?? '')
  const lineStart = text.lastIndexOf('\n', Math.max(0, pos - 1)) + 1
  const lineEnd = text.indexOf('\n', pos)
  return text.slice(lineStart, lineEnd < 0 ? text.length : lineEnd)
}

function usageEntry(source, pane, from, to, kind, name) {
  const loc = offsetToLineCol(source, from)
  return {
    pane,
    from,
    to,
    line: loc.line,
    col: loc.col,
    kind,
    name,
    excerpt: currentLineText(source, from).trim(),
  }
}

function symbolAt(project, pane, source, pos) {
  const word = wordAt(source, pos)
  if (!word) return null
  const before = source.slice(0, word.from)
  const after = source.slice(word.to)
  const member = before.match(/([A-Za-z_]\w*)\s*\.\s*$/)

  if (pane === 'falcon') {
    if (project.componentById.has(word.text)) {
      return { kind: 'component', name: word.text, range: word }
    }
    if (member && project.componentById.has(member[1])) {
      return { kind: 'componentMember', owner: member[1], name: word.text, range: word }
    }
    if (project.falcon.functions.some(fn => fn.name === word.text) || /^\s*\(/.test(after)) {
      return { kind: 'function', name: word.text, range: word }
    }
    if (project.falcon.variables.some(variable => variable.name === word.text)) {
      return { kind: 'variable', name: word.text, range: word }
    }
    return { kind: 'word', name: word.text, range: word }
  }

  const designComponent = project.design.components.find(component =>
    word.from >= component.from && word.to <= component.to
  )
  if (designComponent?.id === word.text) return { kind: 'component', name: word.text, range: word }
  if (project.catalog.components[word.text]) return { kind: 'componentType', name: word.text, range: word }
  return { kind: 'word', name: word.text, range: word }
}

export function findDefinition({ falconCode, designCode, pane, pos, catalog }) {
  const source = pane === 'falcon' ? falconCode : designCode
  const project = buildProjectIndex(falconCode, designCode, catalog)
  const symbol = symbolAt(project, pane, source, pos)
  if (!symbol) return null

  if (symbol.kind === 'component') {
    const component = project.componentById.get(symbol.name)
    return component ? { pane: 'design', from: component.from, line: component.line, col: component.col, label: component.id } : null
  }

  if (symbol.kind === 'function') {
    const fn = project.falcon.functions.find(candidate => candidate.name === symbol.name)
    return fn ? { pane: 'falcon', from: fn.from, line: fn.line, col: fn.col, label: fn.name } : null
  }

  if (symbol.kind === 'variable') {
    const variable = [...project.falcon.variables]
      .filter(candidate => candidate.name === symbol.name && candidate.from <= symbol.range.from)
      .sort((a, b) => b.from - a.from)[0]
    return variable ? { pane: 'falcon', from: variable.from, line: variable.line, col: variable.col, label: variable.name } : null
  }

  return null
}

export function findUsages({ falconCode, designCode, pane, pos, catalog }) {
  const source = pane === 'falcon' ? falconCode : designCode
  const project = buildProjectIndex(falconCode, designCode, catalog)
  const symbol = symbolAt(project, pane, source, pos)
  if (!symbol) return { symbol: null, usages: [] }

  if (symbol.kind === 'component') {
    const usages = [
      ...findWordRanges(falconCode, symbol.name).map(range =>
        usageEntry(falconCode, 'falcon', range.from, range.to, 'component', symbol.name),
      ),
      ...findWordRanges(designCode, symbol.name).map(range =>
        usageEntry(designCode, 'design', range.from, range.to, 'component', symbol.name),
      ),
    ]
    return { symbol, usages }
  }

  if (symbol.kind === 'function') {
    const usages = findCallNameRanges(falconCode, symbol.name).map(range =>
      usageEntry(falconCode, 'falcon', range.from, range.to, 'function', symbol.name),
    )
    return { symbol, usages }
  }

  const activeCode = pane === 'falcon' ? falconCode : designCode
  const usages = findWordRanges(activeCode, symbol.name).map(range =>
    usageEntry(activeCode, pane, range.from, range.to, symbol.kind, symbol.name),
  )
  return { symbol, usages }
}

function applyRanges(source, ranges, insert) {
  let text = String(source ?? '')
  for (const range of [...ranges].sort((a, b) => b.from - a.from)) {
    text = text.slice(0, range.from) + insert + text.slice(range.to)
  }
  return text
}

export function renameSymbol({ falconCode, designCode, pane, pos, catalog, newName }) {
  const replacement = String(newName ?? '').trim()
  if (!IDENT_RE.test(replacement)) return { ok: false, error: 'Use a valid Falcon identifier.' }

  const source = pane === 'falcon' ? falconCode : designCode
  const project = buildProjectIndex(falconCode, designCode, catalog)
  const symbol = symbolAt(project, pane, source, pos)
  if (!symbol || symbol.kind === 'componentMember' || symbol.kind === 'componentType') {
    return { ok: false, error: 'This symbol cannot be renamed yet.' }
  }

  if (symbol.kind === 'component') {
    const falconRanges = findWordRanges(falconCode, symbol.name)
    const designRanges = findWordRanges(designCode, symbol.name)
    return {
      ok: true,
      falconCode: applyRanges(falconCode, falconRanges, replacement),
      designCode: applyRanges(designCode, designRanges, replacement),
      count: falconRanges.length + designRanges.length,
      name: symbol.name,
    }
  }

  if (symbol.kind === 'function') {
    const ranges = findCallNameRanges(falconCode, symbol.name)
    return {
      ok: true,
      falconCode: applyRanges(falconCode, ranges, replacement),
      designCode,
      count: ranges.length,
      name: symbol.name,
    }
  }

  const activeCode = pane === 'falcon' ? falconCode : designCode
  const ranges = findWordRanges(activeCode, symbol.name)
  return {
    ok: true,
    falconCode: pane === 'falcon' ? applyRanges(activeCode, ranges, replacement) : falconCode,
    designCode: pane === 'design' ? applyRanges(activeCode, ranges, replacement) : designCode,
    count: ranges.length,
    name: symbol.name,
  }
}

function braceDelta(line) {
  const masked = maskIgnoredSpans(line)
  let delta = 0
  for (const ch of masked) {
    if (ch === '{') delta++
    else if (ch === '}') delta--
  }
  return delta
}

function normalizeFalconLine(line) {
  return line
    .replace(/\s+,/g, ',')
    .replace(/,\s*/g, ', ')
    .replace(/\s*([+*%]|==|!=|<=|>=|\|\||&&)\s*/g, ' $1 ')
    .replace(/\s*(?<![!<>=])=(?![=>])\s*/g, ' = ')
    .replace(/\s+\{/g, ' {')
    .replace(/\s+$/g, '')
}

function normalizeDesignLine(line) {
  return line
    .replace(/\s+,/g, ',')
    .replace(/,\s*/g, ', ')
    .replace(/\s*:\s*/g, ': ')
    .replace(/\s+\{/g, ' {')
    .replace(/\s+$/g, '')
}

export function formatSource(source, mode = 'falcon') {
  const lines = String(source ?? '').replace(/\r\n/g, '\n').split('\n')
  const out = []
  let indent = 0
  let previousBlank = false
  const normalize = mode === 'design' ? normalizeDesignLine : normalizeFalconLine

  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed) {
      if (out.length && !previousBlank) out.push('')
      previousBlank = true
      continue
    }

    previousBlank = false
    const leadingClosers = trimmed.match(/^}+/)?.[0].length ?? 0
    if (leadingClosers) indent = Math.max(0, indent - leadingClosers)

    out.push(`${'  '.repeat(indent)}${normalize(trimmed)}`)
    indent = Math.max(0, indent + braceDelta(trimmed))
  }

  return `${out.join('\n').trimEnd()}\n`
}

function commonIndent(lines) {
  const indents = lines
    .filter(line => line.trim())
    .map(line => line.match(/^\s*/)?.[0].length ?? 0)
  return indents.length ? Math.min(...indents) : 0
}

function stripCommonIndent(text) {
  const lines = String(text ?? '').replace(/\r\n/g, '\n').split('\n')
  const indent = commonIndent(lines)
  return lines.map(line => line.slice(Math.min(indent, line.match(/^\s*/)?.[0].length ?? 0))).join('\n')
}

function indentText(text, indent = '  ') {
  return String(text ?? '').split('\n').map(line => line.trim() ? `${indent}${line}` : line).join('\n')
}

export function extractFunction(source, from, to, name) {
  const fnName = String(name ?? '').trim()
  if (!IDENT_RE.test(fnName)) return { ok: false, error: 'Use a valid Falcon function name.' }
  const selected = String(source ?? '').slice(from, to)
  if (!selected.trim()) return { ok: false, error: 'Select Falcon statements before extracting.' }
  const body = indentText(stripCommonIndent(selected).trimEnd())
  const declaration = `func ${fnName}() = {\n${body}\n}\n\n`
  return {
    ok: true,
    code: declaration + source.slice(0, from) + `${fnName}()` + source.slice(to),
    replacementOffset: declaration.length + from,
  }
}

export function surroundSelection(source, from, to, kind) {
  const selected = String(source ?? '').slice(from, to)
  if (!selected.trim()) return { ok: false, error: 'Select code before surrounding it.' }

  const body = indentText(stripCommonIndent(selected).trimEnd())
  const templates = {
    if: `if (condition) {\n${body}\n}`,
    while: `while (condition) {\n${body}\n}`,
    for: `for item in list {\n${body}\n}`,
    function: `func name() = {\n${body}\n}`,
  }
  const insert = templates[kind]
  if (!insert) return { ok: false, error: 'Unknown surround template.' }
  return { ok: true, code: source.slice(0, from) + insert + source.slice(to), from, to: from + insert.length }
}

function levenshtein(a, b) {
  const left = String(a ?? '').toLowerCase()
  const right = String(b ?? '').toLowerCase()
  const dp = Array.from({ length: left.length + 1 }, () => Array(right.length + 1).fill(0))
  for (let i = 0; i <= left.length; i++) dp[i][0] = i
  for (let j = 0; j <= right.length; j++) dp[0][j] = j
  for (let i = 1; i <= left.length; i++) {
    for (let j = 1; j <= right.length; j++) {
      const cost = left[i - 1] === right[j - 1] ? 0 : 1
      dp[i][j] = Math.min(dp[i - 1][j] + 1, dp[i][j - 1] + 1, dp[i - 1][j - 1] + cost)
    }
  }
  return dp[left.length][right.length]
}

function nearestNames(name, candidates, limit = 4) {
  const unique = [...new Set(candidates.filter(Boolean))]
  return unique
    .map(candidate => ({ candidate, score: levenshtein(name, candidate) }))
    .filter(item => item.candidate !== name && item.score <= Math.max(2, Math.ceil(name.length / 2)))
    .sort((a, b) => a.score - b.score || a.candidate.localeCompare(b.candidate))
    .slice(0, limit)
    .map(item => item.candidate)
}

function diagnosticToRange(source, diagnostic) {
  if (!diagnostic) return null
  const from = lineColToOffset(source, diagnostic.line, diagnostic.column)
  const length = Math.max(1, Number(diagnostic.length) || 1)
  return { from, to: Math.min(String(source ?? '').length, from + length) }
}

function componentTypeAt(source, pos) {
  const searchable = maskIgnoredSpans(source.slice(0, pos))
  let depth = 0
  for (let i = searchable.length - 1; i >= 0; i--) {
    const ch = searchable[i]
    if (ch === '}') depth++
    else if (ch === '{') {
      if (depth === 0) {
        const before = searchable.slice(0, i)
        const match = before.match(/([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*[A-Za-z_]\w*)?\s*$/)
        return match?.[1] ?? null
      }
      depth--
    }
  }
  return null
}

function matchingParen(source, open) {
  let depth = 0
  const masked = maskIgnoredSpans(source)
  for (let i = open; i < masked.length; i++) {
    if (masked[i] === '(') depth++
    else if (masked[i] === ')') {
      depth--
      if (depth === 0) return i
    }
    if (masked[i] === '\n' && depth <= 1) return -1
  }
  return -1
}

function callSignatureFix(source, word, project) {
  const after = source.slice(word.to)
  const openOffset = after.match(/^\s*\(/)?.[0]?.length
  if (!openOffset) return null
  const open = word.to + openOffset - 1
  const close = matchingParen(source, open)
  if (close < 0) return null

  const before = source.slice(0, word.from)
  const member = before.match(/([A-Za-z_]\w*)\s*\.\s*$/)
  let params = null
  let label = word.text

  if (member) {
    const owner = project.componentById.get(member[1])
    const component = owner ? project.catalog.components[owner.type] : null
    const method = component?.methods?.find(candidate => candidate.name === word.text)
    if (method) {
      params = splitParams(method.params)
      label = `${member[1]}.${word.text}`
    } else {
      const chainMethod = project.catalog.methods.find(candidate => candidate.name === word.text)
      if (chainMethod) params = splitParams(chainMethod.params)
    }
  } else {
    const projectFn = project.falcon.functions.find(candidate => candidate.name === word.text)
    const builtIn = project.catalog.functions.find(candidate => candidate.name === word.text)
    params = projectFn?.params ?? (builtIn ? paramsForCount(builtIn.paramCount) : null)
  }

  if (!params) return null
  const insert = params.join(', ')
  if (source.slice(open + 1, close).trim() === insert) return null
  return {
    title: `Use signature ${label}(${insert})`,
    pane: 'falcon',
    from: open + 1,
    to: close,
    insert,
  }
}

export function quickFixesAt({ falconCode, designCode, pane, pos, diagnostics = [], catalog }) {
  const source = pane === 'falcon' ? falconCode : designCode
  const project = buildProjectIndex(falconCode, designCode, catalog)
  const ranges = diagnostics.map(diagnostic => ({ diagnostic, range: diagnosticToRange(source, diagnostic) })).filter(item => item.range)
  const active = ranges.find(item => pos >= item.range.from && pos <= item.range.to)
    ?? ranges.find(item => Math.abs(item.range.from - pos) < 40)
  const baseRange = active?.range ?? wordAt(source, pos)
  const word = wordAt(source, baseRange?.from ?? pos)
  if (!word) return []

  const fixes = []
  const callFix = pane === 'falcon' ? callSignatureFix(source, word, project) : null
  if (callFix) fixes.push(callFix)

  const candidates = pane === 'falcon'
    ? [
        ...project.falcon.functions.map(fn => fn.name),
        ...project.falcon.variables.map(variable => variable.name),
        ...project.componentById.keys(),
        ...project.catalog.functions.map(fn => fn.name),
      ]
    : [
        ...project.catalog.componentNames,
        ...Object.values(project.catalog.components[componentTypeAt(source, word.from)]?.blockProperties ?? {}).map(prop => prop.name),
      ]

  for (const candidate of nearestNames(word.text, candidates)) {
    fixes.push({
      title: `Replace "${word.text}" with "${candidate}"`,
      pane,
      from: word.from,
      to: word.to,
      insert: candidate,
    })
  }

  return fixes
}

export function documentSymbols(falconCode, designCode) {
  const falcon = parseFalconSymbols(falconCode)
  const design = parseDesignSymbols(designCode)
  return [
    ...falcon.functions.map(symbol => ({
      ...symbol,
      label: `func ${symbol.name}(${symbol.params.join(', ')})`,
      detail: 'Screen1.falcon',
    })),
    ...falcon.events.map(symbol => ({
      ...symbol,
      label: `when ${symbol.name}`,
      detail: 'Screen1.falcon',
    })),
    ...design.components.map(symbol => ({
      ...symbol,
      label: symbol.id ? `${symbol.type}.${symbol.id}` : symbol.type,
      detail: 'Screen1.design',
    })),
  ].sort((a, b) => a.pane.localeCompare(b.pane) || a.from - b.from)
}

export function projectSearch(falconCode, designCode, query) {
  const needle = String(query ?? '')
  if (!needle) return []
  const re = new RegExp(escapeRegExp(needle), 'gi')
  const collect = (source, pane) => {
    const results = []
    let match
    while ((match = re.exec(source)) !== null) {
      results.push(usageEntry(source, pane, match.index, match.index + match[0].length, 'match', needle))
      if (match[0] === '') re.lastIndex++
    }
    return results
  }
  return [...collect(falconCode, 'falcon'), ...collect(designCode, 'design')]
}

export function breadcrumbsForPosition(source, pane, pos) {
  const text = String(source ?? '')
  const before = text.slice(0, pos)
  if (pane === 'falcon') {
    const contexts = []
    let match
    const re = /\b(func\s+[A-Za-z_]\w*|when\s+[A-Za-z_]\w*\.[A-Za-z_]\w*)/g
    while ((match = re.exec(before)) !== null) contexts.push(match[1])
    return contexts.slice(-1)[0] ?? 'top level'
  }

  const searchable = maskIgnoredSpans(before)
  const stack = []
  let token = ''
  for (let i = 0; i < searchable.length; i++) {
    const ch = searchable[i]
    if (ch === '{') {
      const name = token.match(/([A-Z][A-Za-z0-9_]*(?:\s*\.\s*[A-Za-z_]\w*)?)\s*$/)?.[1]
      if (name) stack.push(name.replace(/\s+/g, ''))
      token = ''
    } else if (ch === '}') {
      stack.pop()
      token = ''
    } else if (ch === '\n') {
      token = ''
    } else {
      token += ch
      if (token.length > 120) token = token.slice(-120)
    }
  }
  return stack.length ? stack.join(' > ') : 'Screen1.design'
}

export function diagnosticLocations(falconCode, designCode, falconDiagnostics = [], designDiagnostics = []) {
  const mapDiagnostic = (diagnostic, pane, source) => {
    const from = lineColToOffset(source, diagnostic.line, diagnostic.column)
    return {
      pane,
      from,
      line: Number(diagnostic.line) || 1,
      col: Number(diagnostic.column) || 1,
      message: diagnostic.message || 'Falcon error',
    }
  }
  return [
    ...falconDiagnostics.map(diagnostic => mapDiagnostic(diagnostic, 'falcon', falconCode)),
    ...designDiagnostics.map(diagnostic => mapDiagnostic(diagnostic, 'design', designCode)),
  ].sort((a, b) => {
    const paneOrder = pane => pane === 'falcon' ? 0 : 1
    return paneOrder(a.pane) - paneOrder(b.pane) || a.from - b.from
  })
}

export function componentDocumentation(catalog, type, member) {
  const component = catalog?.components?.[type]
  if (!component) return ''
  if (!member) return stripHtml(component.helpString)
  const prop = component.blockProperties?.find(candidate => candidate.name === member)
  if (prop) return stripHtml(prop.description)
  const method = component.methods?.find(candidate => candidate.name === member)
  if (method) return stripHtml(method.description)
  const event = component.events?.find(candidate => candidate.name === member)
  if (event) return stripHtml(event.description)
  return ''
}
