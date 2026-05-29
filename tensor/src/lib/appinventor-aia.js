import simpleComponents from '../../../lang/code/compdb/simple_components.json';
import { getProjectSnapshot } from './stores.js';
import { mistToXml } from './falcon-wasm.js';
import { componentDefinitionsFromDesigner } from './blockly-preview.js';
import {
  actionBarForTheme,
  applyProjectPropertiesToScmProperties,
  extractProjectPropertiesFromScmProperties,
  normalizeProjectProperties,
  projectPropertiesToAiaProperties,
} from './project-properties.js';

const YA_VERSION = '233';
const BLOCKS_LANGUAGE_VERSION = '37';
const DEFAULT_USER = 'ai_tensor';
const DEFAULT_PROJECT = 'TensorProject';
const PROJECT_PROPERTIES_PATH = 'youngandroidproject/project.properties';
const SCM_JSON_PREFIX = '#|\n$JSON\n';
const SCM_JSON_SUFFIX = '\n|#';

let jsZipPromise = null;

const COMPONENT_META = new Map(simpleComponents.map(component => [component.name, component]));
const KNOWN_COMPONENT_TYPES = new Set([...COMPONENT_META.keys(), 'Screen']);
const TYPE_ALIASES = {
  Screen: 'Form',
  Form: 'Form',
};
const CANVAS_CHILD_TYPES = new Set(['Ball', 'ImageSprite']);
const MAP_CHILD_TYPES = new Set(['Circle', 'FeatureCollection', 'LineString', 'Marker', 'Polygon', 'Rectangle']);
const FEATURE_COLLECTION_CHILD_TYPES = new Set(['Circle', 'LineString', 'Marker', 'Polygon', 'Rectangle']);
const CHART_CHILD_TYPES = new Set(['ChartData2D', 'Trendline']);

function loadScript(src) {
  const existing = document.querySelector(`script[data-tensor-src="${src}"]`);
  if (existing?.dataset.loaded === 'true') return Promise.resolve();
  if (existing?.dataset.loading === 'true') {
    return new Promise((resolve, reject) => {
      existing.addEventListener('load', resolve, { once: true });
      existing.addEventListener('error', reject, { once: true });
    });
  }
  if (existing) existing.remove();

  return new Promise((resolve, reject) => {
    const script = document.createElement('script');
    script.src = src;
    script.async = false;
    script.dataset.tensorSrc = src;
    script.dataset.loading = 'true';
    script.onload = () => {
      script.dataset.loading = 'false';
      script.dataset.loaded = 'true';
      resolve();
    };
    script.onerror = () => {
      script.dataset.loading = 'false';
      script.dataset.error = 'true';
      script.remove();
      reject(new Error(`Failed to load ${src}`));
    };
    document.head.appendChild(script);
  });
}

async function ensureJSZip() {
  if (window.JSZip) return window.JSZip;
  if (!jsZipPromise) {
    jsZipPromise = loadScript('/jszip.min.js').then(() => {
      if (!window.JSZip) throw new Error('JSZip did not initialize');
      return window.JSZip;
    }).catch(error => {
      jsZipPromise = null;
      throw error;
    });
  }
  return jsZipPromise;
}

function stripKnownExtension(name) {
  return String(name || '')
    .replace(/\.(aia|zip|fc)$/i, '')
    .trim();
}

export function sanitizeProjectName(name) {
  const clean = stripKnownExtension(name)
    .replace(/[^A-Za-z0-9_]/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_+|_+$/g, '');
  const safe = clean || DEFAULT_PROJECT;
  return /^[A-Za-z]/.test(safe) ? safe : `Project_${safe}`;
}

function sanitizeScreenName(name, fallback = 'Screen1') {
  const clean = String(name || '')
    .replace(/[^A-Za-z0-9_]/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_+|_+$/g, '');
  const safe = clean || fallback;
  return /^[A-Za-z]/.test(safe) ? safe : `Screen_${safe}`;
}

function uniqueScreenName(name, usedNames, fallback = 'Screen1') {
  const clean = sanitizeScreenName(name, fallback);
  if (!usedNames.has(clean)) return clean;
  const match = clean.match(/^(.*?)(\d+)$/);
  const stem = match ? match[1] : clean;
  let n = match ? Number(match[2]) + 1 : 2;
  let next = `${stem}${n}`;
  while (usedNames.has(next)) {
    n += 1;
    next = `${stem}${n}`;
  }
  return next;
}

function zipPathForAsset(name) {
  const parts = String(name || '')
    .replace(/\\/g, '/')
    .split('/')
    .map(part => part.trim())
    .filter(part => part && part !== '.' && part !== '..');
  return parts.length ? `assets/${parts.join('/')}` : '';
}

function sourceBasePath(project) {
  return `src/appinventor/${DEFAULT_USER}/${project}`;
}

function qualifiedMain(project, screen = 'Screen1') {
  return `appinventor.${DEFAULT_USER}.${project}.${screen}`;
}

function parseProperties(text) {
  const props = {};
  for (const rawLine of String(text || '').split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!line || line.startsWith('#') || line.startsWith('!')) continue;
    const index = findPropertySeparator(line);
    const key = index === -1 ? line : line.slice(0, index).trim();
    const value = index === -1 ? '' : line.slice(index + 1).trim();
    if (key) props[unescapePropertyText(key)] = unescapePropertyText(value);
  }
  return props;
}

function findPropertySeparator(line) {
  let escaped = false;
  for (let i = 0; i < line.length; i += 1) {
    const ch = line[i];
    if (escaped) {
      escaped = false;
      continue;
    }
    if (ch === '\\') {
      escaped = true;
      continue;
    }
    if (ch === ':' || ch === '=') return i;
  }
  return -1;
}

function unescapePropertyText(value) {
  let out = '';
  const text = String(value ?? '');
  for (let i = 0; i < text.length; i += 1) {
    const ch = text[i];
    if (ch !== '\\' || i + 1 >= text.length) {
      out += ch;
      continue;
    }
    const next = text[++i];
    if (next === 'n') out += '\n';
    else if (next === 'r') out += '\r';
    else if (next === 't') out += '\t';
    else out += next;
  }
  return out;
}

function escapePropertyText(value, { key = false } = {}) {
  let out = '';
  const text = String(value ?? '');
  for (let i = 0; i < text.length; i += 1) {
    const ch = text[i];
    if (ch === '\\') out += '\\\\';
    else if (ch === '\n') out += '\\n';
    else if (ch === '\r') out += '\\r';
    else if (ch === '\t') out += '\\t';
    else if (ch === ':' || ch === '=' || ch === '#' || ch === '!' || (key && /\s/.test(ch))) out += `\\${ch}`;
    else out += ch;
  }
  return out;
}

function projectPropertiesText(project, baseProperties = {}) {
  const aiaProperties = projectPropertiesToAiaProperties(baseProperties);
  const theme = aiaProperties.theme || 'AppTheme.Light.DarkActionBar';
  const actionBar = actionBarForTheme(theme);
  const props = {
    ...aiaProperties,
    main: qualifiedMain(project, 'Screen1'),
    name: project,
    assets: '../assets',
    source: '../src',
    build: '../build',
    versioncode: aiaProperties.versioncode || '1',
    versionname: aiaProperties.versionname || '1.0',
    useslocation: aiaProperties.useslocation || 'False',
    aname: aiaProperties.aname ?? '',
    sizing: aiaProperties.sizing || 'Responsive',
    showlistsasjson: aiaProperties.showlistsasjson || 'True',
    actionbar: actionBar,
    theme,
    defaultfilescope: aiaProperties.defaultfilescope || 'App',
    projectcolors: aiaProperties.projectcolors || '{}',
    'color.primary': aiaProperties['color.primary'] || '&HFF3F51B5',
    'color.primary.dark': aiaProperties['color.primary.dark'] || '&HFF303F9F',
    'color.accent': aiaProperties['color.accent'] || '&HFFFF4081',
  };

  const preferredOrder = [
    'main',
    'name',
    'assets',
    'source',
    'build',
    'versioncode',
    'versionname',
    'useslocation',
    'aname',
    'sizing',
    'showlistsasjson',
    'tutorialurl',
    'subsetjson',
    'actionbar',
    'theme',
    'defaultfilescope',
    'color.primary',
    'color.primary.dark',
    'color.accent',
    'aiversioning',
    'lastopened',
    'buildnumber',
    'NSBluetoothAlwaysUsageDescription',
    'NSBluetoothPeripheralUsageDescription',
    'NSContactsUsageDescription',
    'NSMicrophoneUsageDescription',
    'NSCameraUsageDescription',
    'NSSpeechRecognitionUsageDescription',
    'NSLocationWhenInUseUsageDescription',
    'projectcolors',
  ];
  const keys = [
    ...preferredOrder.filter(key => key in props),
    ...Object.keys(props).filter(key => !preferredOrder.includes(key)).sort(),
  ];
  return `${keys.map(key => `${escapePropertyText(key, { key: true })}=${escapePropertyText(props[key])}`).join('\n')}\n`;
}

function textFileBaseName(path) {
  return String(path || '').split('/').pop()?.replace(/\.[^.]+$/, '') || '';
}

function screenSort(a, b) {
  if (a.name === 'Screen1') return -1;
  if (b.name === 'Screen1') return 1;
  return a.name.localeCompare(b.name, undefined, { numeric: true, sensitivity: 'base' });
}

function extractScmJson(text) {
  const source = String(text || '').trim();
  const match = source.match(/\$JSON\s*([\s\S]*?)\s*\|#/);
  const jsonText = match ? match[1].trim() : source;
  if (!jsonText.startsWith('{')) {
    throw new Error('Only JSON-format App Inventor .scm files are supported');
  }
  return JSON.parse(jsonText);
}

function quoteSchemaString(value) {
  return `"${String(value)
    .replace(/\\/g, '\\\\')
    .replace(/"/g, '\\"')}"`;
}

function designerEditorTypeToValueKind(editorType) {
  if (editorType === 'boolean') return 'boolean';
  if (['color', 'float', 'integer', 'layout_size', 'non_negative_float', 'non_negative_integer'].includes(editorType)) return 'number';
  if (editorType === 'ListViewAddData') return 'list';
  return 'text';
}

function componentPropertyValueKind(componentType, propName) {
  const component = COMPONENT_META.get(scmType(componentType));
  const blockProp = component?.blockProperties?.find(prop => prop.name === propName);
  if (blockProp?.type) return blockProp.type;
  const designerProp = component?.properties?.find(prop => prop.name === propName);
  if (designerProp?.editorType) return designerEditorTypeToValueKind(designerProp.editorType);
  return 'any';
}

function schemaValueText(value, componentType = '', propName = '') {
  const text = String(value ?? '');
  const kind = componentPropertyValueKind(componentType, propName);
  if (kind === 'boolean') {
    if (/^(true|1)$/i.test(text)) return 'true';
    if (/^(false|0)$/i.test(text)) return 'false';
    return quoteSchemaString(text);
  }
  if (kind === 'number') {
    if (/^-?\d+(?:\.\d+)?$/.test(text) || /^&H[0-9A-Fa-f]{8}$/.test(text)) return text;
    return quoteSchemaString(text);
  }
  if (kind === 'text' || kind === 'list' || kind === 'component') {
    return quoteSchemaString(text);
  }
  if (/^(true|false|-?\d+(?:\.\d+)?|&H[0-9A-Fa-f]{8})$/i.test(text)) {
    return /^(true|false)$/i.test(text) ? text.toLowerCase() : text;
  }
  return quoteSchemaString(text);
}

function scmComponentToSchema(component, depth = 0) {
  const indent = '  '.repeat(depth);
  const type = component.$Type === 'Form' ? 'Screen' : component.$Type;
  const name = component.$Name || type;
  const props = Object.entries(component)
    .filter(([key]) => !key.startsWith('$') && key !== 'Uuid')
    .map(([key, value]) => `${indent}  ${key}: ${schemaValueText(value, type, key)}`);
  const children = (component.$Components || [])
    .map(child => scmComponentToSchema(child, depth + 1));
  const body = [...props, ...children];

  if (!body.length) return `${indent}${type}.${name}`;
  return `${indent}${type}.${name} {\n${body.join(',\n')}\n${indent}}`;
}

function scmToDesignSchema(text, fallbackScreenName) {
  const scm = extractScmJson(text);
  const properties = scm?.Properties;
  if (!properties || typeof properties !== 'object') {
    throw new Error(`Unable to read designer properties for ${fallbackScreenName}`);
  }
  if (!properties.$Name) properties.$Name = fallbackScreenName;
  return scmComponentToSchema(properties);
}

function stripLineComments(text) {
  let out = '';
  let inString = false;
  let escaped = false;

  for (let i = 0; i < text.length; i += 1) {
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

function parseDesignSchema(source) {
  const text = stripLineComments(source || '');
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

  function makeComponent(ident) {
    const dotIdx = ident.indexOf('.');
    const type = dotIdx === -1 ? ident : ident.slice(0, dotIdx);
    let name = dotIdx === -1 ? '' : ident.slice(dotIdx + 1);
    if (!type) fail('Expected a component type');
    if (!isKnownComponentType(type)) fail(`Unknown component type "${type}"`);
    if (!name) {
      typeCounts[type] = (typeCounts[type] || 0) + 1;
      name = `${type}${typeCounts[type]}`;
    }
    return { type, name, props: {}, children: [] };
  }

  function parseComponent() {
    skipWs();
    const ident = readIdent();
    if (!ident) fail('Expected a component');
    const component = makeComponent(ident);
    skipWs();
    if (text[pos] !== '{') return component;

    pos += 1;
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
        pos = save;
        component.children.push(parseComponent());
      }

      skipWs();
      if (text[pos] === ',') pos += 1;
    }

    return component;
  }

  const root = parseComponent();
  skipWs();
  if (pos < text.length) fail(`Unexpected token "${text[pos]}"`);
  if (root.type !== 'Screen' && root.type !== 'Form') fail('The root designer component must be Screen');
  validateDesignTree(root);
  validateUniqueComponentNames(root);
  return root;
}

function scmType(type) {
  return TYPE_ALIASES[type] || type;
}

function isKnownComponentType(type) {
  return KNOWN_COMPONENT_TYPES.has(type) || KNOWN_COMPONENT_TYPES.has(scmType(type));
}

function isNonVisibleComponentType(type) {
  if (type === 'Screen' || type === 'Form') return false;
  return COMPONENT_META.get(scmType(type))?.nonVisible === 'true';
}

const CONTAINER_TYPES = new Set([
  'Screen',
  'Form',
  'ScrollHorizontal',
  'ScrollVertical',
  ...simpleComponents
    .filter(component => component.categoryString === 'LAYOUT')
    .map(component => component.name),
]);

function canContainComponent(parentType, childType) {
  const parent = scmType(parentType);
  const child = scmType(childType);
  if (!parent || !child) return false;
  if (CANVAS_CHILD_TYPES.has(child)) return parent === 'Canvas';
  if (MAP_CHILD_TYPES.has(child)) {
    if (parent === 'Map') return true;
    return parent === 'FeatureCollection' && FEATURE_COLLECTION_CHILD_TYPES.has(child);
  }
  if (CHART_CHILD_TYPES.has(child)) return parent === 'Chart';
  if (isNonVisibleComponentType(child)) {
    return parentType === 'Screen' || parentType === 'Form';
  }
  if (isNonVisibleComponentType(parent)) return false;
  if (parent === 'Canvas' || parent === 'Map' || parent === 'FeatureCollection' || parent === 'Chart') return false;
  return CONTAINER_TYPES.has(parent);
}

function validateDesignTree(node, parent = null) {
  if (parent && !canContainComponent(parent.type, node.type)) {
    throw new Error(`${node.type}.${node.name} cannot be placed inside ${parent.type}.${parent.name}`);
  }
  for (const child of node.children || []) validateDesignTree(child, node);
}

function validateUniqueComponentNames(node, seen = new Set()) {
  if (seen.has(node.name)) {
    throw new Error(`Duplicate component name "${node.name}"`);
  }
  seen.add(node.name);
  for (const child of node.children || []) validateUniqueComponentNames(child, seen);
}

function componentVersion(type) {
  return COMPONENT_META.get(scmType(type))?.version || '1';
}

function stableUuid(seed) {
  let hash = 0;
  const text = String(seed || '');
  for (let i = 0; i < text.length; i += 1) {
    hash = Math.imul(31, hash) + text.charCodeAt(i);
    hash |= 0;
  }
  return String(hash || 1);
}

function nodeToScmProperties(node, path, screenName, project) {
  const type = scmType(node.type);
  const props = {
    $Name: path === '0' ? screenName : node.name,
    $Type: type,
    $Version: componentVersion(node.type),
    Uuid: path === '0' ? '0' : stableUuid(`${path}:${node.type}:${node.name}`),
  };

  if (path === '0') {
    props.Title = node.props?.Title || screenName;
    props.AppName = project;
  }

  for (const [key, value] of Object.entries(node.props || {})) {
    props[key] = String(value);
  }

  if (node.children?.length) {
    props.$Components = node.children.map((child, index) =>
      nodeToScmProperties(child, `${path}-${index}`, child.name, project)
    );
  }

  return props;
}

function designSchemaToScm(schema, screenName, project, projectProperties = null) {
  const trimmed = String(schema || '').trim();
  const root = trimmed
    ? parseDesignSchema(trimmed)
    : { type: 'Screen', name: screenName, props: { Title: screenName }, children: [] };
  let properties = nodeToScmProperties(root, '0', screenName, project);
  if (projectProperties && screenName === 'Screen1') {
    properties = applyProjectPropertiesToScmProperties(properties, projectProperties, project);
  }
  const scm = {
    authURL: [],
    YaVersion: YA_VERSION,
    Source: 'Form',
    Properties: properties,
  };
  return `${SCM_JSON_PREFIX}${JSON.stringify(scm)}${SCM_JSON_SUFFIX}`;
}

function scmTextWithProjectProperties(text, projectProperties, project) {
  const scm = extractScmJson(text);
  scm.Properties = applyProjectPropertiesToScmProperties(scm.Properties, projectProperties, project);
  return `${SCM_JSON_PREFIX}${JSON.stringify(scm)}${SCM_JSON_SUFFIX}`;
}

function xmlChunks(xml) {
  return String(xml || '')
    .split('\0')
    .map(chunk => chunk.trim())
    .filter(Boolean);
}

function hasParserError(doc) {
  return doc.getElementsByTagName('parsererror').length > 0;
}

function emptyBlocklyXml() {
  return `<xml xmlns="http://www.w3.org/1999/xhtml">\n  <yacodeblocks ya-version="${YA_VERSION}" language-version="${BLOCKS_LANGUAGE_VERSION}"></yacodeblocks>\n</xml>`;
}

function generatedBlocklyXml(xmlText) {
  const chunks = xmlChunks(xmlText);
  if (!chunks.length) return emptyBlocklyXml();

  const doc = document.implementation.createDocument('http://www.w3.org/1999/xhtml', 'xml');
  const root = doc.documentElement;
  const parser = new DOMParser();

  for (const chunk of chunks) {
    const sourceDoc = parser.parseFromString(chunk, 'text/xml');
    if (hasParserError(sourceDoc)) {
      throw new Error('Falcon generated invalid Blockly XML');
    }
    const sourceRoot = sourceDoc.documentElement;
    const children = sourceRoot.tagName.toLowerCase() === 'xml'
      ? Array.from(sourceRoot.childNodes)
      : [sourceRoot];

    for (const child of children) {
      if (child.nodeType !== Node.ELEMENT_NODE) continue;
      if (child.tagName.toLowerCase() === 'yacodeblocks') continue;
      root.appendChild(doc.importNode(child, true));
      root.appendChild(doc.createTextNode('\n  '));
    }
  }

  const meta = doc.createElement('yacodeblocks');
  meta.setAttribute('ya-version', YA_VERSION);
  meta.setAttribute('language-version', BLOCKS_LANGUAGE_VERSION);
  root.appendChild(meta);

  return new XMLSerializer()
    .serializeToString(doc)
    .replace(/(<xml[^>]*>)/, '$1\n  ')
    .replace(/(<yacodeblocks[^>]*><\/yacodeblocks>)/, '$1\n');
}

function falconSourceFromCells(cells) {
  return (cells || [])
    .filter(cell => cell.type === 'code')
    .map(cell => cell.code || '')
    .join('\n\n')
    .trim();
}

async function blocklyXmlForScreen(screen) {
  const source = falconSourceFromCells(screen.cells);
  if (!source) {
    return String(screen.rawBlocklyXml || '').trim() || emptyBlocklyXml();
  }

  const defs = await componentDefinitionsFromDesigner(screen.designCode || '');
  const xml = await mistToXml(source, defs);
  return generatedBlocklyXml(xml);
}

async function blobForAsset(asset) {
  if (asset?.blob instanceof Blob) return asset.blob;
  if (asset?.url) {
    const response = await fetch(asset.url);
    if (response.ok) return response.blob();
  }
  if (asset?.bytes) return new Blob([asset.bytes], { type: asset.type || '' });
  return new Blob([]);
}

function normalizedAssetRecord(name, blob) {
  const id = `asset-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  return {
    id,
    name,
    size: blob.size,
    type: blob.type || '',
    blob,
    url: URL.createObjectURL(blob),
  };
}

function importedBlocksCell(screenName, rawBlocklyXml) {
  if (!String(rawBlocklyXml || '').trim()) return [];
  return [{
    id: `imported-${screenName}-${Date.now()}`,
    type: 'markdown',
    content: '<div class="md-p">Imported App Inventor blocks are preserved for export. Add Falcon code to replace this screen&apos;s blocks.</div>',
  }];
}

export async function importAiaFile(file) {
  const JSZip = await ensureJSZip();
  const zip = await JSZip.loadAsync(file);
  const propsEntry = zip.file(PROJECT_PROPERTIES_PATH);
  if (!propsEntry) {
    throw new Error('This file is not an App Inventor project archive: missing youngandroidproject/project.properties');
  }

  const projectProperties = parseProperties(await propsEntry.async('string'));
  const project = sanitizeProjectName(projectProperties.name || stripKnownExtension(file?.name) || DEFAULT_PROJECT);
  const screenFiles = new Map();

  for (const [path, entry] of Object.entries(zip.files)) {
    if (entry.dir || !path.startsWith('src/') || !/\.(scm|bky)$/i.test(path)) continue;
    const base = path.replace(/\.(scm|bky)$/i, '');
    const record = screenFiles.get(base) || {};
    if (/\.scm$/i.test(path)) record.scm = path;
    if (/\.bky$/i.test(path)) record.bky = path;
    screenFiles.set(base, record);
  }

  const screens = [];
  const usedScreenNames = new Set();
  let screen1ProjectProperties = {};
  for (const [base, record] of screenFiles) {
    if (!record.scm) continue;
    const rawScm = await zip.file(record.scm).async('string');
    const rawBlocklyXml = record.bky ? await zip.file(record.bky).async('string') : '';
    const fallbackName = sanitizeScreenName(textFileBaseName(base), `Screen${screens.length + 1}`);
    let designCode = '';
    let screenName = fallbackName;

    try {
      const scm = extractScmJson(rawScm);
      screenName = sanitizeScreenName(scm?.Properties?.$Name || fallbackName, fallbackName);
      designCode = scmComponentToSchema(scm.Properties);
      if (screenName === 'Screen1' && !usedScreenNames.has('Screen1')) {
        screen1ProjectProperties = extractProjectPropertiesFromScmProperties(scm.Properties);
      }
    } catch {
      designCode = scmToDesignSchema(rawScm, fallbackName);
    }
    screenName = uniqueScreenName(screenName, usedScreenNames, fallbackName);
    usedScreenNames.add(screenName);

    screens.push({
      name: screenName,
      cells: importedBlocksCell(screenName, rawBlocklyXml),
      designCode,
      rawBlocklyXml,
      sourceScm: rawScm,
      sourceDesignCode: designCode,
    });
  }

  if (!screens.length) {
    throw new Error('No App Inventor screens were found in the project archive');
  }

  const assets = [];
  for (const [path, entry] of Object.entries(zip.files)) {
    if (entry.dir || !path.startsWith('assets/')) continue;
    const name = path.slice('assets/'.length);
    if (!name) continue;
    const blob = await entry.async('blob');
    assets.push(normalizedAssetRecord(name, blob));
  }

  screens.sort(screenSort);
  const screenNames = new Set(screens.map(screen => screen.name));
  const activeScreen = screenNames.has(projectProperties.lastopened)
    ? projectProperties.lastopened
    : screens[0].name;

  return {
    projectName: project,
    projectProperties: normalizeProjectProperties({
      ...screen1ProjectProperties,
      ...projectProperties,
    }),
    activeScreen,
    screens,
    assets,
  };
}

export async function exportCurrentProjectToAia() {
  const JSZip = await ensureJSZip();
  const snapshot = getProjectSnapshot();
  const project = sanitizeProjectName(snapshot.projectName);
  const zip = new JSZip();
  const basePath = sourceBasePath(project);
  const sortedScreens = [...snapshot.screens].sort(screenSort);
  const normalizedProjectProperties = normalizeProjectProperties(snapshot.projectProperties);
  const usedScreenNames = new Set();
  const screenRecords = sortedScreens.map((screen, index) => {
    const sanitized = sanitizeScreenName(screen.name, `Screen${index + 1}`);
    const screenName = uniqueScreenName(sanitized, usedScreenNames, `Screen${index + 1}`);
    usedScreenNames.add(screenName);
    return { screen, screenName, sanitized };
  });
  const activeExportScreen = screenRecords.find(record => record.screen.name === snapshot.activeScreen)?.screenName
    || screenRecords[0]?.screenName
    || 'Screen1';

  zip.file(PROJECT_PROPERTIES_PATH, projectPropertiesText(project, {
    ...normalizedProjectProperties,
    lastopened: activeExportScreen,
  }));

  for (const { screen, screenName, sanitized } of screenRecords) {
    const screenBase = `${basePath}/${screenName}`;
    const designUnchanged = screen.sourceScm
      && screen.sourceDesignCode
      && screen.designCode === screen.sourceDesignCode;
    let scm;
    if (designUnchanged && screenName === sanitized) {
      scm = screen.sourceScm;
      if (screenName === 'Screen1') {
        try {
          scm = scmTextWithProjectProperties(scm, normalizedProjectProperties, project);
        } catch {
          scm = designSchemaToScm(screen.designCode, screenName, project, normalizedProjectProperties);
        }
      }
    } else {
      scm = designSchemaToScm(
        screen.designCode,
        screenName,
        project,
        screenName === 'Screen1' ? normalizedProjectProperties : null
      );
    }
    const bky = await blocklyXmlForScreen(screen);

    zip.file(`${screenBase}.scm`, scm);
    zip.file(`${screenBase}.bky`, bky);
  }

  for (const asset of snapshot.assets || []) {
    const name = typeof asset === 'string' ? asset : asset?.name;
    const path = zipPathForAsset(name);
    if (!path) continue;
    zip.file(path, await blobForAsset(asset));
  }

  return zip.generateAsync({
    type: 'blob',
    compression: 'DEFLATE',
    comment: 'Built with MIT App Inventor',
  });
}

export function downloadBlob(blob, filename) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}
