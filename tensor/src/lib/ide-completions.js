import { snippetCompletion } from '@codemirror/autocomplete'
import { extractComponentDefs } from './companion.js'

const FALCON_KEYWORDS = [
  'func', 'when', 'if', 'else', 'while', 'for', 'in', 'step',
  'break', 'yield', 'local', 'global', 'this',
]

const FALCON_TYPES = [
  'text', 'number', 'list', 'dict', 'emptyList', 'emptyText', 'base10', 'hexa', 'bin',
]

const ATOMS = ['true', 'false']

const DEFAULT_CATALOG = {
  functions: [],
  methods: [],
  componentNames: [],
  components: {},
}

const BOOST = {
  snippet: 70,
  keyword: 35,
  type: 20,
  atom: 25,
  builtInFunction: 10,
  builtInMethod: 20,
  projectFunction: 95,
  variable: 100,
  component: 92,
  componentMember: 88,
  event: 96,
  designProperty: 92,
  designComponent: 45,
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

function paramPlaceholders(params, fallbackCount = 0) {
  const names = splitParams(params)
  const count = names.length || Math.max(0, Number(fallbackCount) || 0)
  return Array.from({ length: count }, (_, index) => {
    const name = names[index] || `arg${index + 1}`
    return `\${${name}}`
  }).join(', ')
}

function callSnippet(name, params, fallbackCount = 0) {
  return `${name}(${paramPlaceholders(params, fallbackCount)})`
}

function functionDetail(paramCount, result) {
  const count = Number(paramCount)
  const arity = count < 0 ? 'variadic' : `${count} arg${count === 1 ? '' : 's'}`
  return `${arity} -> ${result || 'any'}`
}

function functionCompletion(fn, source = 'built-in') {
  return snippetCompletion(callSnippet(fn.name, '', fn.paramCount), {
    label: fn.name,
    type: 'function',
    detail: functionDetail(fn.paramCount, fn.result),
    info: `${source} function`,
    boost: source === 'project' ? BOOST.projectFunction : BOOST.builtInFunction,
  })
}

function methodCompletion(method) {
  return snippetCompletion(callSnippet(method.name, method.params, method.paramCount), {
    label: method.name,
    type: 'method',
    detail: `${method.module || 'value'} -> ${method.result || 'any'}`,
    info: method.params ? `.${method.name}(${method.params})` : `.${method.name}()`,
    boost: BOOST.builtInMethod,
  })
}

function componentMethodCompletion(method) {
  const params = splitParams(method.params)
  return snippetCompletion(callSnippet(method.name, params, params.length), {
    label: method.name,
    type: 'method',
    detail: params.length ? `${params.length} arg${params.length === 1 ? '' : 's'}` : 'no args',
    info: stripHtml(method.description),
    boost: BOOST.componentMember,
  })
}

function componentPropertyCompletion(prop) {
  return {
    label: prop.name,
    type: 'property',
    detail: prop.type || prop.rw || 'property',
    info: stripHtml(prop.description),
    boost: BOOST.componentMember + (String(prop.rw ?? '').toLowerCase().includes('write') ? 4 : 0),
  }
}

function eventSnippet(event) {
  const params = splitParams(event.params)
  const signature = params.length ? `(${params.join(', ')})` : ''
  return `${event.name}${signature} {\n  \${}\n}`
}

function componentEventCompletion(event) {
  const params = splitParams(event.params)
  return snippetCompletion(eventSnippet(event), {
    label: event.name,
    type: 'event',
    detail: params.length ? params.join(', ') : 'event',
    info: stripHtml(event.description),
    boost: BOOST.event,
  })
}

function eventHandlerCompletion(componentId, event) {
  const params = splitParams(event.params)
  const signature = params.length ? `(${params.join(', ')})` : ''
  return snippetCompletion(`${componentId}.${event.name}${signature} {\n  \${}\n}`, {
    label: `${componentId}.${event.name}`,
    type: 'event',
    detail: params.length ? params.join(', ') : 'event',
    info: stripHtml(event.description),
    boost: BOOST.event,
  })
}

function uniqueByLabel(options) {
  const seen = new Set()
  return options.filter(option => {
    if (!option?.label || seen.has(option.label)) return false
    seen.add(option.label)
    return true
  })
}

function reverseComponentDefs(componentDefs) {
  const reverse = {}
  for (const [type, ids] of Object.entries(componentDefs ?? {})) {
    for (const id of ids ?? []) reverse[id] = type
  }
  return reverse
}

function functionSymbols(source) {
  const options = []
  const re = /\bfunc\s+([A-Za-z_]\w*)\s*\(([^)]*)\)/g
  let match
  while ((match = re.exec(source)) !== null) {
    const params = splitParams(match[2])
    options.push(snippetCompletion(callSnippet(match[1], params, params.length), {
      label: match[1],
      type: 'function',
      detail: params.length ? params.join(', ') : 'no args',
      info: 'project function',
      boost: BOOST.projectFunction,
    }))
  }
  return uniqueByLabel(options)
}

function variableSymbols(source) {
  const options = []
  const patterns = [
    /\b(?:local|global)\s+([A-Za-z_]\w*)/g,
    /\bfor\s+([A-Za-z_]\w*)\s+in\b/g,
    /\bfor\s+([A-Za-z_]\w*)\s*,\s*([A-Za-z_]\w*)\s+in\b/g,
    /\bfunc\s+[A-Za-z_]\w*\s*\(([^)]*)\)/g,
    /\bwhen\s+[A-Za-z_]\w*\.[A-Za-z_]\w*\s*\(([^)]*)\)/g,
  ]

  for (const re of patterns) {
    let match
    while ((match = re.exec(source)) !== null) {
      const names = re.source.includes('([^)]*)')
        ? splitParams(match[1])
        : match.slice(1).filter(Boolean)
      for (const name of names) {
        options.push({ label: name, type: 'variable', detail: 'symbol', boost: BOOST.variable })
      }
    }
  }

  return uniqueByLabel(options)
}

function linePrefix(context) {
  const line = context.state.doc.lineAt(context.pos)
  return {
    line,
    text: line.text.slice(0, context.pos - line.from),
  }
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

function wordRange(context) {
  return context.matchBefore(/[A-Za-z_]\w*/)
}

function completionResult(context, from, options) {
  const filtered = uniqueByLabel(options)
  if (!filtered.length) return null
  return {
    from,
    options: filtered,
    validFor: /^[A-Za-z_]\w*$/,
  }
}

function componentOptions(componentDefs) {
  const options = []
  for (const [type, ids] of Object.entries(componentDefs ?? {})) {
    for (const id of ids ?? []) {
      options.push({ label: id, type: 'variable', detail: type, boost: BOOST.component })
    }
  }
  return options
}

function topLevelFalconOptions(catalog, source, componentDefs) {
  const keywordOptions = FALCON_KEYWORDS.map(label => ({ label, type: 'keyword', boost: BOOST.keyword }))
  const typeOptions = FALCON_TYPES.map(label => ({ label, type: 'type', boost: BOOST.type }))
  const atomOptions = ATOMS.map(label => ({ label, type: 'constant', boost: BOOST.atom }))
  const snippets = [
    snippetCompletion('func ${name}(${params}) = {\n  ${}\n}', {
      label: 'func',
      type: 'keyword',
      detail: 'function',
      boost: BOOST.snippet,
    }),
    snippetCompletion('when ${component}.${event} {\n  ${}\n}', {
      label: 'when',
      type: 'keyword',
      detail: 'event handler',
      boost: BOOST.snippet,
    }),
    snippetCompletion('if (${condition}) {\n  ${}\n}', {
      label: 'if',
      type: 'keyword',
      detail: 'conditional',
      boost: BOOST.snippet,
    }),
    snippetCompletion('local ${name} = ${value}', {
      label: 'local',
      type: 'keyword',
      detail: 'variable',
      boost: BOOST.snippet,
    }),
  ]

  return [
    ...snippets,
    ...variableSymbols(source),
    ...functionSymbols(source),
    ...componentOptions(componentDefs),
    ...keywordOptions,
    ...typeOptions,
    ...atomOptions,
    ...catalog.functions.map(fn => functionCompletion(fn)),
  ]
}

export function createFalconCompletionSource({ catalog, designCode, projectCode }) {
  const activeCatalog = normalizeCatalog(catalog)
  const componentDefs = extractComponentDefs(designCode ?? '')
  const reverseComponents = reverseComponentDefs(componentDefs)

  return context => {
    const { text: prefix } = linePrefix(context)
    if (inLineCommentOrString(prefix)) return null

    const source = context.state.doc.toString()
    const projectSource = projectCode || source
    const dotMatch = prefix.match(/([A-Za-z_]\w*)\.(\w*)$/)
    if (dotMatch) {
      const [, objectName, typed] = dotMatch
      const from = context.pos - typed.length
      const componentType = reverseComponents[objectName] || (activeCatalog.components[objectName] ? objectName : null)
      const component = componentType ? activeCatalog.components[componentType] : null
      const inEventHeader = /^\s*when\s+[A-Za-z_]\w*\.\w*$/.test(prefix)

      if (component && inEventHeader) {
        return completionResult(context, from, (component.events ?? []).map(componentEventCompletion))
      }
      if (component) {
        return completionResult(context, from, [
          ...(component.methods ?? []).map(componentMethodCompletion),
          ...(component.blockProperties ?? []).map(componentPropertyCompletion),
        ])
      }
      return completionResult(context, from, activeCatalog.methods.map(methodCompletion))
    }

    const whenHeaderMatch = prefix.match(/^\s*when\s+([A-Za-z_]\w*)?$/)
    if (whenHeaderMatch && (context.explicit || whenHeaderMatch[1])) {
      const events = []
      for (const [type, ids] of Object.entries(componentDefs)) {
        const component = activeCatalog.components[type]
        for (const id of ids ?? []) {
          for (const event of component?.events ?? []) events.push(eventHandlerCompletion(id, event))
        }
      }
      const typed = whenHeaderMatch[1] ?? ''
      return completionResult(context, context.pos - typed.length, events)
    }

    const word = wordRange(context)
    if (!word || (word.from === word.to && !context.explicit)) return null

    return completionResult(context, word.from, topLevelFalconOptions(activeCatalog, projectSource, componentDefs))
  }
}

function designPropertyValue(propName, prop) {
  const type = prop?.type || ''
  if (type === 'boolean') return `${propName}: true`
  if (type === 'number') return `${propName}: \${0}`
  return `${propName}: "\${}"`
}

function designPropertyCompletion(prop) {
  return snippetCompletion(designPropertyValue(prop.name, prop), {
    label: prop.name,
    type: 'property',
    detail: prop.type || prop.rw || 'property',
    info: stripHtml(prop.description),
    boost: BOOST.designProperty + (String(prop.rw ?? '').toLowerCase().includes('write') ? 4 : 0),
  })
}

function nextComponentId(type, componentDefs) {
  const ids = componentDefs[type] ?? []
  for (let i = ids.length + 1; i < ids.length + 100; i++) {
    const candidate = `${type}${i}`
    if (!ids.includes(candidate)) return candidate
  }
  return `${type}${ids.length + 1}`
}

function componentTypeCompletion(type, componentDefs) {
  const defaultId = nextComponentId(type, componentDefs)
  return snippetCompletion(`${type}.\${${defaultId}} { \${} }`, {
    label: type,
    type: 'class',
    detail: 'component',
    boost: BOOST.designComponent,
  })
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

export function createDesignCompletionSource({ catalog }) {
  const activeCatalog = normalizeCatalog(catalog)

  return context => {
    const { text: prefix } = linePrefix(context)
    if (inLineCommentOrString(prefix)) return null

    const source = context.state.doc.toString()
    const componentDefs = extractComponentDefs(source)

    const valueMatch = prefix.match(/([A-Za-z_]\w*)\s*:\s*([A-Za-z_]*)$/)
    if (valueMatch) {
      const typed = valueMatch[2]
      return completionResult(context, context.pos - typed.length, ATOMS.map(label => ({
        label,
        type: 'constant',
        detail: 'boolean',
        boost: BOOST.atom,
      })))
    }

    const dotMatch = prefix.match(/\b([A-Z][A-Za-z0-9_]*)\.(\w*)$/)
    if (dotMatch) {
      const [, type, typed] = dotMatch
      const existing = (componentDefs[type] ?? []).map(id => ({ label: id, type: 'variable', detail: type, boost: BOOST.component }))
      const fresh = { label: nextComponentId(type, componentDefs), type: 'variable', detail: 'new component id', boost: BOOST.component + 4 }
      return completionResult(context, context.pos - typed.length, [fresh, ...existing])
    }

    const word = wordRange(context)
    if (!word || (word.from === word.to && !context.explicit)) return null

    const enclosingType = enclosingComponentType(source, context.pos)
    const component = enclosingType ? activeCatalog.components[enclosingType] : null
    const propertyOptions = (component?.blockProperties ?? []).map(designPropertyCompletion)
    const componentOptions = activeCatalog.componentNames.map(type => componentTypeCompletion(type, componentDefs))

    return completionResult(context, word.from, [
      ...propertyOptions,
      ...componentOptions,
      ...ATOMS.map(label => ({ label, type: 'constant', detail: 'boolean', boost: BOOST.atom })),
    ])
  }
}
