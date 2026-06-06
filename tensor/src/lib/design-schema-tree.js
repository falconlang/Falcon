import {
  allComponentDescriptors,
  componentMetaMap,
  knownComponentTypeSet,
} from './appinventor-component-registry.js';
import {
  appInventorNameError,
  isValidComponentName,
  isValidScreenName,
} from './appinventor-validation.js';
const TYPE_ALIASES = {
  Screen: 'Form',
  Form: 'Form',
};

const CANVAS_CHILD_TYPES = new Set(['Ball', 'ImageSprite']);
const MAP_CHILD_TYPES = new Set(['Circle', 'FeatureCollection', 'LineString', 'Marker', 'Polygon', 'Rectangle']);
const FEATURE_COLLECTION_CHILD_TYPES = new Set(['Circle', 'LineString', 'Marker', 'Polygon', 'Rectangle']);
const CHART_CHILD_TYPES = new Set(['ChartData2D', 'Trendline']);

function dbType(type) {
  return TYPE_ALIASES[type] || type;
}

function componentTypeName(ident) {
  const dotIdx = ident.indexOf('.');
  return dotIdx === -1 ? ident : ident.slice(0, dotIdx);
}

function isKnownComponentType(type) {
  const known = knownComponentTypeSet();
  return known.has(type) || known.has(dbType(type));
}

function isNonVisibleComponentType(type) {
  const resolved = dbType(type);
  if (resolved === 'Form') return false;
  return componentMetaMap().get(resolved)?.nonVisible === 'true';
}

function containerTypes() {
  return new Set([
    'Screen',
    'Form',
    'ScrollHorizontal',
    'ScrollVertical',
    ...allComponentDescriptors()
      .filter(component => component.categoryString === 'LAYOUT')
      .map(component => component.name),
  ]);
}

export function canContainDesignComponent(parentType, childType) {
  const parent = dbType(parentType);
  const child = dbType(childType);
  if (!parent || !child) return false;
  if (CANVAS_CHILD_TYPES.has(child)) return parent === 'Canvas';
  if (MAP_CHILD_TYPES.has(child)) {
    if (parent === 'Map') return true;
    return parent === 'FeatureCollection' && FEATURE_COLLECTION_CHILD_TYPES.has(child);
  }
  if (CHART_CHILD_TYPES.has(child)) return parent === 'Chart';
  if (isNonVisibleComponentType(child)) return parent === 'Form';
  if (isNonVisibleComponentType(parent)) return false;
  if (parent === 'Canvas' || parent === 'Map' || parent === 'FeatureCollection' || parent === 'Chart') return false;
  return containerTypes().has(parent);
}

export function stripDesignerLineComments(text) {
  let out = '';
  let inString = false;
  let escaped = false;

  for (let i = 0; i < String(text || '').length; i += 1) {
    const ch = text[i];
    const next = text[i + 1];
    if (inString) {
      out += ch;
      if (escaped) escaped = false;
      else if (ch === '\\') escaped = true;
      else if (ch === '"') inString = false;
      continue;
    }
    if (ch === '"') {
      inString = true;
      out += ch;
      continue;
    }
    if (ch === '/' && next === '/') {
      while (i < text.length && text[i] !== '\n') i += 1;
      if (i < text.length) out += text[i];
      continue;
    }
    out += ch;
  }

  return out;
}

export function validateDesignTreePlacement(node, parent = null) {
  if (parent && !canContainDesignComponent(parent.type, node.type)) {
    throw new Error(`${node.type}.${node.name} cannot be placed inside ${parent.type}.${parent.name}`);
  }
  for (const child of node.children || []) validateDesignTreePlacement(child, node);
}

export function validateUniqueDesignNames(node, seen = new Set()) {
  if (seen.has(node.name)) throw new Error(`Duplicate component name "${node.name}"`);
  seen.add(node.name);
  for (const child of node.children || []) validateUniqueDesignNames(child, seen);
}

export function parseDesignSchema(source, { pathIds = true, validate = true } = {}) {
  const raw = String(source || '');
  if (!raw.trim()) throw new Error('The design schema is empty.');
  const text = stripDesignerLineComments(raw);
  let pos = 0;
  const typeCounts = {};

  function skipWs() {
    while (pos < text.length && /\s/.test(text[pos])) pos += 1;
  }
  function describePos() {
    const before = text.slice(0, pos);
    const line = before.split('\n').length;
    const col = before.length - before.lastIndexOf('\n');
    return `line ${line}, column ${col}`;
  }
  function fail(message) {
    throw new Error(`${message} (${describePos()})`);
  }
  function readIdent() {
    skipWs();
    let ident = '';
    while (pos < text.length && /[\w.]/.test(text[pos])) ident += text[pos++];
    return ident;
  }
  function readString() {
    pos += 1;
    let value = '';
    while (pos < text.length && text[pos] !== '"') {
      if (text[pos] === '\\') {
        pos += 1;
        if (pos >= text.length) fail('Unterminated string literal');
        const escaped = text[pos++];
        if (escaped === 'b') value += '\b';
        else if (escaped === 'f') value += '\f';
        else if (escaped === 'n') value += '\n';
        else if (escaped === 'r') value += '\r';
        else if (escaped === 't') value += '\t';
        else if (escaped === 'u') {
          const hex = text.slice(pos, pos + 4);
          if (!/^[0-9A-Fa-f]{4}$/.test(hex)) fail('Invalid unicode escape');
          value += String.fromCharCode(parseInt(hex, 16));
          pos += 4;
        } else {
          value += escaped;
        }
        continue;
      }
      value += text[pos++];
    }
    if (pos >= text.length) fail('Unterminated string literal');
    pos += 1;
    return value;
  }
  function readValue() {
    skipWs();
    if (text[pos] === '"') return readString();
    let value = '';
    while (pos < text.length && !/[,\n{}]/.test(text[pos])) value += text[pos++];
    value = value.trim();
    if (!value) fail('Expected a property value');
    return value;
  }
  function makeComponent(ident, pathId) {
    const dotIdx = ident.indexOf('.');
    const type = componentTypeName(ident);
    let name = dotIdx === -1 ? '' : ident.slice(dotIdx + 1);
    if (!type) fail('Expected a component type');
    if (!isKnownComponentType(type)) fail(`Unknown component type "${type}"`);
    if (!name) {
      typeCounts[type] = (typeCounts[type] || 0) + 1;
      name = `${type}${typeCounts[type]}`;
    }
    const isRootType = type === 'Screen' || type === 'Form';
    const valid = isRootType ? isValidScreenName(name) : isValidComponentName(name);
    if (!valid) fail(appInventorNameError(name, { component: !isRootType }));
    const node = { type, name, props: {}, children: [] };
    if (pathIds) node.pathId = pathId;
    return node;
  }
  function parseComponent(pathId) {
    skipWs();
    const ident = readIdent();
    if (!ident) fail('Expected a component');
    const component = makeComponent(ident, pathId);
    skipWs();
    if (text[pos] !== '{') return component;
    pos += 1;
    let childIndex = 0;
    while (true) {
      skipWs();
      if (text[pos] === '}') {
        pos += 1;
        break;
      }
      if (pos >= text.length) fail(`Expected "}" to close ${component.type}.${component.name}`);
      const save = pos;
      const token = readIdent();
      if (!token) fail('Expected a property or child component');
      skipWs();
      if (text[pos] === ':') {
        pos += 1;
        component.props[token] = readValue();
      } else {
        if (!isKnownComponentType(componentTypeName(token))) fail(`Expected ":" after property "${token}"`);
        pos = save;
        component.children.push(parseComponent(`${pathId}-${childIndex++}`));
      }
      skipWs();
      if (text[pos] === ',') pos += 1;
    }
    return component;
  }

  const root = parseComponent('0');
  skipWs();
  if (pos < text.length) fail(`Unexpected token "${text[pos]}"`);
  if (root.type !== 'Screen' && root.type !== 'Form') fail('The root designer component must be Screen');
  if (validate) {
    validateDesignTreePlacement(root);
    validateUniqueDesignNames(root);
  }
  return root;
}

export function parseDesignSchemaResult(source, options) {
  try {
    return { root: parseDesignSchema(source, options), error: null };
  } catch (error) {
    return { root: null, error: error?.message || 'Unable to parse design schema.' };
  }
}

function quoteSchemaString(value) {
  return JSON.stringify(String(value));
}

function isColorProp(node, propName) {
  return /(Color|Colour)$/.test(propName) || propName === 'PaintColor';
}

function needsQuotes(value, node, propName) {
  const text = String(value).trim();
  if (/^(true|false|True|False|-?\d+\.?\d*|&H[0-9A-Fa-f]{8})$/.test(text)) return false;
  if (isColorProp(node, propName) && /^#[0-9A-Fa-f]{3,8}$/.test(text)) return false;
  return true;
}

export function serializeDesignComponent(node, depth = 0) {
  const indent = '  '.repeat(depth);
  const ident = `${node.type}.${node.name}`;
  const propLines = Object.entries(node.props || {}).map(
    ([key, value]) => `${indent}  ${key}: ${needsQuotes(value, node, key) ? quoteSchemaString(value) : value}`,
  );
  const childLines = (node.children || []).map(child => serializeDesignComponent(child, depth + 1));
  const body = [...propLines, ...childLines];
  if (!body.length) return `${indent}${ident}`;
  return `${indent}${ident} {\n${body.join(',\n')}\n${indent}}`;
}

export function serializeDesignTree(root) {
  return root ? serializeDesignComponent(root, 0) : '';
}

export function designTreeToComponentDefinitions(root) {
  const defs = {};
  function visit(node) {
    if (!node) return;
    const type = node.type === 'Form' ? 'Screen' : node.type;
    defs[type] ||= [];
    if (!defs[type].includes(node.name)) defs[type].push(node.name);
    for (const child of node.children || []) visit(child);
  }
  visit(root);
  return defs;
}
