function waitForReady(timeout = 10000) {
  if (typeof window.describeComponent === 'function') return Promise.resolve()
  return new Promise((resolve, reject) => {
    const deadline = Date.now() + timeout
    const check = () => {
      if (typeof window.describeComponent === 'function') resolve()
      else if (Date.now() > deadline) reject(new Error('[falcon-wasm] timeout waiting for WASM'))
      else setTimeout(check, 50)
    }
    setTimeout(check, 0)
  })
}

function withMistErrors(fn) {
  const errors = []
  const original = window.mistError
  window.mistError = (msg) => errors.push(msg)
  try {
    const result = fn()
    if (errors.length) throw new Error(errors.join('\n'))
    return result
  } finally {
    window.mistError = original
  }
}

export class FalconDiagnosticError extends Error {
  constructor(result, fallbackMessage = 'Falcon compilation failed') {
    super(result?.error || result?.diagnostics?.[0]?.message || fallbackMessage)
    this.name = 'FalconDiagnosticError'
    this.result = result
    this.diagnostics = result?.diagnostics ?? []
    this.raw = result?.error || ''
  }
}

function normalizeDiagnostic(diagnostic) {
  if (!diagnostic || typeof diagnostic !== 'object') return null
  const line = Number(diagnostic.line)
  const column = Number(diagnostic.column)
  const length = Number(diagnostic.length)
  return {
    message: String(diagnostic.message || 'Falcon error'),
    severity: diagnostic.severity || 'error',
    phase: diagnostic.phase || 'compile',
    file: diagnostic.file || '',
    line: Number.isFinite(line) && line > 0 ? line : 1,
    column: Number.isFinite(column) && column > 0 ? column : 1,
    length: Number.isFinite(length) && length > 0 ? length : 1,
    raw: diagnostic.raw || '',
  }
}

function normalizeDiagnosticResult(result, extra = {}) {
  const obj = result && typeof result === 'object' ? result : {}
  const diagnostics = Array.from(obj.diagnostics ?? [])
    .map(normalizeDiagnostic)
    .filter(Boolean)
  return {
    ok: Boolean(obj.ok),
    error: typeof obj.error === 'string' ? obj.error : '',
    diagnostics,
    ...extra(obj),
  }
}

function normalizeMistToXmlResult(result) {
  if (typeof result === 'string') {
    try {
      const parsed = JSON.parse(result)
      if (parsed && typeof parsed === 'object' && typeof parsed.xml === 'string') {
        return normalizeMistToXmlResult(parsed)
      }
    } catch {}
    return { xml: result, lineNumbers: [] }
  }

  if (result && typeof result === 'object') {
    return {
      xml: typeof result.xml === 'string' ? result.xml : '',
      lineNumbers: Array.from(result.lineNumbers ?? [])
        .map(Number)
        .filter(Number.isFinite),
    }
  }

  return { xml: '', lineNumbers: [] }
}

function normalizeMistDiagnosticResult(result) {
  return normalizeDiagnosticResult(result, obj => {
    const normalized = normalizeMistToXmlResult(obj)
    return {
      xml: normalized.xml,
      lineNumbers: normalized.lineNumbers,
    }
  })
}

function normalizeYailDiagnosticResult(result) {
  return normalizeDiagnosticResult(result, obj => ({
    yail: typeof obj.yail === 'string' ? obj.yail : '',
  }))
}

export async function describeComponent(name) {
  await waitForReady()
  const json = window.describeComponent(name)
  return json ? JSON.parse(json) : null
}

export async function listComponents() {
  await waitForReady()
  const json = window.listComponents?.()
  return json ? JSON.parse(json) : []
}

export async function falconCompletionCatalog() {
  await waitForReady()
  if (typeof window.falconCompletionCatalog !== 'function') {
    return { functions: [], methods: [] }
  }
  const json = window.falconCompletionCatalog()
  return json ? JSON.parse(json) : { functions: [], methods: [] }
}

export async function loadIdeCatalog() {
  const [falcon, componentNames] = await Promise.all([
    falconCompletionCatalog(),
    listComponents(),
  ])
  const componentEntries = await Promise.all(
    componentNames.map(async name => [name, await describeComponent(name)]),
  )
  return {
    functions: falcon.functions ?? [],
    methods: falcon.methods ?? [],
    componentNames,
    components: Object.fromEntries(componentEntries.filter(([, desc]) => desc)),
  }
}

export async function mistToXmlResult(sourceCode, componentDefs, safe = true) {
  await waitForReady()
  if (typeof window.mistToXmlWithDiagnostics === 'function') {
    const result = normalizeMistDiagnosticResult(window.mistToXmlWithDiagnostics(sourceCode, componentDefs, safe))
    if (!result.ok) throw new FalconDiagnosticError(result)
    return { xml: result.xml, lineNumbers: result.lineNumbers }
  }
  return withMistErrors(() => normalizeMistToXmlResult(window.mistToXml(sourceCode, componentDefs, safe)))
}

export async function mistToXmlDiagnosticResult(sourceCode, componentDefs, safe = true) {
  await waitForReady()
  if (typeof window.mistToXmlWithDiagnostics === 'function') {
    return normalizeMistDiagnosticResult(window.mistToXmlWithDiagnostics(sourceCode, componentDefs, safe))
  }
  try {
    const result = await mistToXmlResult(sourceCode, componentDefs, safe)
    return { ok: true, error: '', diagnostics: [], ...result }
  } catch (e) {
    return { ok: false, error: e.message || String(e), diagnostics: [], xml: '', lineNumbers: [] }
  }
}

export async function mistToXml(sourceCode, componentDefs, safe = true) {
  return (await mistToXmlResult(sourceCode, componentDefs, safe)).xml
}

export async function getComponentDefinitionsCode(sourceCode) {
  await waitForReady()
  return withMistErrors(() => window.getComponentDefinitionsCode(sourceCode))
}

export async function blocklyToYail(xml) {
  await waitForReady()
  return withMistErrors(() => {
    const xmlText = xml && typeof xml === 'object' && typeof xml.xml === 'string' ? xml.xml : xml
    const chunks = String(xmlText).split('\0').map(chunk => chunk.trim()).filter(Boolean)
    return chunks.map(chunk => window.blocklyToYail(chunk)).filter(Boolean).join('\n')
  })
}

export async function annToYail(annSource, codeYail = '') {
  await waitForReady()
  if (typeof window.annToYailWithDiagnostics === 'function') {
    const result = normalizeYailDiagnosticResult(window.annToYailWithDiagnostics(annSource, codeYail))
    if (!result.ok) throw new FalconDiagnosticError(result)
    return result.yail
  }
  return withMistErrors(() => window.annToYail(annSource, codeYail))
}

export async function annToYailDiagnosticResult(annSource, codeYail = '') {
  await waitForReady()
  if (typeof window.annToYailWithDiagnostics === 'function') {
    return normalizeYailDiagnosticResult(window.annToYailWithDiagnostics(annSource, codeYail))
  }
  try {
    const yail = await annToYail(annSource, codeYail)
    return { ok: true, error: '', diagnostics: [], yail }
  } catch (e) {
    return { ok: false, error: e.message || String(e), diagnostics: [], yail: '' }
  }
}
