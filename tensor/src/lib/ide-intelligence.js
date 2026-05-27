import { StateField } from '@codemirror/state'
import { hoverTooltip, showTooltip } from '@codemirror/view'
import { extractComponentDefs } from './companion.js'

const DEFAULT_CATALOG = {
  functions: [],
  methods: [],
  componentNames: [],
  components: {},
}

function normalizeCatalog(catalog) {
  return {
    ...DEFAULT_CATALOG,
    ...(catalog ?? {}),
    functions: catalog?.functions ?? [],
    methods: catalog?.methods ?? [],
    componentNames: catalog?.componentNames ?? [],
    components: catalog?.components ?? {},
  }
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

function reverseComponentDefs(componentDefs) {
  const reverse = {}
  for (const [type, ids] of Object.entries(componentDefs ?? {})) {
    for (const id of ids ?? []) reverse[id] = type
  }
  return reverse
}

function functionSymbols(source) {
  const symbols = []
  const re = /\bfunc\s+([A-Za-z_]\w*)\s*\(([^)]*)\)/g
  let match
  while ((match = re.exec(source)) !== null) {
    symbols.push({
      name: match[1],
      params: splitParams(match[2]),
      kind: 'project function',
      from: match.index,
    })
  }
  return symbols
}

function linePrefix(state, pos) {
  const line = state.doc.lineAt(pos)
  return line.text.slice(0, pos - line.from)
}

function inLineComment(prefix) {
  let inString = false
  for (let i = 0; i < prefix.length; i++) {
    const ch = prefix[i]
    const next = prefix[i + 1]
    if (!inString && ch === '/' && next === '/') return true
    if (ch === '"' && prefix[i - 1] !== '\\') inString = !inString
  }
  return false
}

function inLineCommentOrString(prefix) {
  let inString = false
  for (let i = 0; i < prefix.length; i++) {
    const ch = prefix[i]
    const next = prefix[i + 1]
    if (!inString && ch === '/' && next === '/') return true
    if (ch === '"' && prefix[i - 1] !== '\\') inString = !inString
  }
  return inString
}

function findOpenParen(source, pos) {
  let depth = 0
  for (let i = pos - 1; i >= 0; i--) {
    const ch = source[i]
    if (ch === ')') depth++
    else if (ch === '(') {
      if (depth === 0) return i
      depth--
    } else if (ch === '\n' && depth === 0) {
      return -1
    }
  }
  return -1
}

function activeParamIndex(source, open, pos) {
  let depth = 0
  let index = 0
  for (let i = open + 1; i < pos; i++) {
    const ch = source[i]
    if (ch === '(' || ch === '[' || ch === '{') depth++
    else if (ch === ')' || ch === ']' || ch === '}') depth = Math.max(0, depth - 1)
    else if (ch === ',' && depth === 0) index++
  }
  return index
}

function callBeforeOpen(source, open) {
  const before = source.slice(0, open)
  const member = before.match(/([A-Za-z_]\w*)\s*\.\s*([A-Za-z_]\w*)\s*$/)
  if (member) return { kind: 'member', object: member[1], name: member[2] }
  const chain = before.match(/\.\s*([A-Za-z_]\w*)\s*$/)
  if (chain) return { kind: 'chain', name: chain[1] }
  const fn = before.match(/([A-Za-z_]\w*)\s*$/)
  if (fn) return { kind: 'function', name: fn[1] }
  return null
}

function componentMethodSignature(component, name) {
  const method = component?.methods?.find(candidate => candidate.name === name)
  if (!method) return null
  const params = splitParams(method.params)
  return {
    name,
    params,
    detail: params.length ? `${params.length} arg${params.length === 1 ? '' : 's'}` : 'no args',
    description: stripHtml(method.description),
  }
}

function componentEventSignature(component, name) {
  const event = component?.events?.find(candidate => candidate.name === name)
  if (!event) return null
  const params = splitParams(event.params)
  return {
    name,
    params,
    detail: 'event',
    description: stripHtml(event.description),
  }
}

function findSignature(source, call, activeCatalog, reverseComponents) {
  if (!call) return null

  if (call.kind === 'member') {
    const componentType = reverseComponents[call.object] || (activeCatalog.components[call.object] ? call.object : null)
    const component = componentType ? activeCatalog.components[componentType] : null
    return componentMethodSignature(component, call.name)
      ?? componentEventSignature(component, call.name)
  }

  if (call.kind === 'chain') {
    const method = activeCatalog.methods.find(candidate => candidate.name === call.name)
    if (!method) return null
    return {
      name: `.${method.name}`,
      params: splitParams(method.params),
      detail: `${method.module || 'value'} -> ${method.result || 'any'}`,
      description: method.params ? `.${method.name}(${method.params})` : `.${method.name}()`,
    }
  }

  const projectFunction = functionSymbols(source).find(symbol => symbol.name === call.name)
  if (projectFunction) {
    return {
      name: projectFunction.name,
      params: projectFunction.params,
      detail: 'project function',
      description: 'Defined in this Falcon file.',
    }
  }

  const fn = activeCatalog.functions.find(candidate => candidate.name === call.name)
  if (!fn) return null
  return {
    name: fn.name,
    params: paramsForCount(fn.paramCount),
    detail: `built-in -> ${fn.result || 'any'}`,
    description: 'Falcon built-in function.',
  }
}

function signatureAtState(state, config) {
  if (config.mode !== 'falcon') return null
  const selection = state.selection.main
  if (!selection.empty || state.selection.ranges.length !== 1) return null
  if (inLineComment(linePrefix(state, selection.head))) return null

  const source = state.doc.toString()
  const open = findOpenParen(source, selection.head)
  if (open < 0) return null

  const activeCatalog = normalizeCatalog(config.catalog)
  const reverseComponents = reverseComponentDefs(extractComponentDefs(config.designCode ?? ''))
  const call = callBeforeOpen(source, open)
  const signature = findSignature(source, call, activeCatalog, reverseComponents)
  if (!signature) return null

  return {
    pos: open,
    active: activeParamIndex(source, open, selection.head),
    signature,
  }
}

function signatureDom(data) {
  const wrap = document.createElement('div')
  wrap.className = 'cm-signature-help'

  const line = document.createElement('div')
  line.className = 'cm-signature-line'
  const name = document.createElement('span')
  name.className = 'cm-signature-name'
  name.textContent = data.signature.name
  line.appendChild(name)
  line.append('(')
  data.signature.params.forEach((param, index) => {
    if (index > 0) line.append(', ')
    const paramNode = document.createElement('span')
    paramNode.className = `cm-signature-param${index === data.active ? ' active' : ''}`
    paramNode.textContent = param
    line.appendChild(paramNode)
  })
  line.append(')')
  wrap.appendChild(line)

  if (data.signature.detail) {
    const detail = document.createElement('div')
    detail.className = 'cm-signature-detail'
    detail.textContent = data.signature.detail
    wrap.appendChild(detail)
  }
  return wrap
}

export function createSignatureHelpExtension(config) {
  const field = StateField.define({
    create(state) {
      return signatureAtState(state, config)
    },
    update(value, transaction) {
      if (!transaction.docChanged && !transaction.selectionSet) return value
      return signatureAtState(transaction.state, config)
    },
    provide: fieldRef => showTooltip.computeN([fieldRef], state => {
      const data = state.field(fieldRef)
      if (!data) return []
      return [{
        pos: data.pos,
        above: true,
        strictSide: false,
        create() {
          return { dom: signatureDom(data) }
        },
      }]
    }),
  })
  return field
}

function wordAt(doc, pos) {
  const source = doc.toString()
  let from = pos
  let to = pos
  while (from > 0 && /[A-Za-z0-9_]/.test(source[from - 1])) from--
  while (to < source.length && /[A-Za-z0-9_]/.test(source[to])) to++
  if (from === to) return null
  return { from, to, text: source.slice(from, to), source }
}

function componentNameBeforeOpening(text, openingIndex) {
  const before = text.slice(0, openingIndex)
  const legacy = before.match(/@([A-Za-z]\w*)\s*$/)
  if (legacy) return legacy[1]
  const match = before.match(/([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*[A-Za-z_]\w*)?\s*$/)
  return match ? match[1] : null
}

function enclosingComponentType(source, pos) {
  const text = source.slice(0, pos)
  let depth = 0
  for (let i = text.length - 1; i >= 0; i--) {
    const ch = text[i]
    if (ch === '}') depth++
    else if (ch === '{') {
      if (depth === 0) return componentNameBeforeOpening(text, i)
      depth--
    }
  }
  return null
}

function hoverInfoForWord(word, activeCatalog, config, view) {
  if (config.mode === 'design') {
    const component = activeCatalog.components[word.text]
    if (component) {
      return {
        title: word.text,
        meta: component.categoryString || 'component',
        body: stripHtml(component.helpString),
      }
    }

    const enclosingType = enclosingComponentType(word.source, word.from)
    const prop = activeCatalog.components[enclosingType]?.blockProperties?.find(candidate => candidate.name === word.text)
    if (prop) {
      return {
        title: prop.name,
        meta: `${enclosingType} property${prop.type ? ` -> ${prop.type}` : ''}`,
        body: stripHtml(prop.description),
      }
    }
    return null
  }

  const componentDefs = extractComponentDefs(config.designCode ?? '')
  const reverseComponents = reverseComponentDefs(componentDefs)
  const componentType = reverseComponents[word.text]
  if (componentType) {
    const component = activeCatalog.components[componentType]
    return {
      title: word.text,
      meta: componentType,
      body: stripHtml(component?.helpString) || 'Component defined in Screen1.design.',
    }
  }

  const before = word.source.slice(0, word.from)
  const member = before.match(/([A-Za-z_]\w*)\s*\.\s*$/)
  if (member) {
    const ownerType = reverseComponents[member[1]] || (activeCatalog.components[member[1]] ? member[1] : null)
    const component = ownerType ? activeCatalog.components[ownerType] : null
    const method = component?.methods?.find(candidate => candidate.name === word.text)
    if (method) {
      return {
        title: `${member[1]}.${method.name}()`,
        meta: `${ownerType} method`,
        body: stripHtml(method.description),
      }
    }
    const prop = component?.blockProperties?.find(candidate => candidate.name === word.text)
    if (prop) {
      return {
        title: `${member[1]}.${prop.name}`,
        meta: `${ownerType} property${prop.type ? ` -> ${prop.type}` : ''}`,
        body: stripHtml(prop.description),
      }
    }
    const event = component?.events?.find(candidate => candidate.name === word.text)
    if (event) {
      return {
        title: `${member[1]}.${event.name}`,
        meta: 'event',
        body: stripHtml(event.description),
      }
    }
  }

  const projectFunction = functionSymbols(view.state.doc.toString()).find(symbol => symbol.name === word.text)
  if (projectFunction) {
    return {
      title: `${projectFunction.name}(${projectFunction.params.join(', ')})`,
      meta: 'project function',
      body: 'Defined in this Falcon file.',
    }
  }

  const fn = activeCatalog.functions.find(candidate => candidate.name === word.text)
  if (fn) {
    return {
      title: `${fn.name}(${paramsForCount(fn.paramCount).join(', ')})`,
      meta: `built-in -> ${fn.result || 'any'}`,
      body: 'Falcon built-in function.',
    }
  }

  const chainMethod = activeCatalog.methods.find(candidate => candidate.name === word.text)
  if (chainMethod) {
    return {
      title: `.${chainMethod.name}(${splitParams(chainMethod.params).join(', ')})`,
      meta: `${chainMethod.module || 'value'} -> ${chainMethod.result || 'any'}`,
      body: chainMethod.params ? `Chain method for ${chainMethod.module} values.` : 'Chain method.',
    }
  }

  return null
}

function hoverDom(info) {
  const wrap = document.createElement('div')
  wrap.className = 'cm-ide-hover'

  const title = document.createElement('div')
  title.className = 'cm-ide-hover-title'
  title.textContent = info.title
  wrap.appendChild(title)

  if (info.meta) {
    const meta = document.createElement('div')
    meta.className = 'cm-ide-hover-meta'
    meta.textContent = info.meta
    wrap.appendChild(meta)
  }

  if (info.body) {
    const body = document.createElement('div')
    body.className = 'cm-ide-hover-body'
    body.textContent = info.body
    wrap.appendChild(body)
  }

  return wrap
}

export function createHoverDocumentationExtension(config) {
  return hoverTooltip((view, pos) => {
    const prefix = linePrefix(view.state, pos)
    if (inLineCommentOrString(prefix)) return null

    const word = wordAt(view.state.doc, pos)
    if (!word) return null

    const activeCatalog = normalizeCatalog(config.catalog)
    const info = hoverInfoForWord(word, activeCatalog, config, view)
    if (!info) return null

    return {
      pos: word.from,
      end: word.to,
      above: true,
      create() {
        return { dom: hoverDom(info) }
      },
    }
  }, { hoverTime: 350 })
}
