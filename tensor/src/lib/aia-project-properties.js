import {
  actionBarForTheme,
  projectPropertiesToAiaProperties,
} from './project-properties.js';
import {
  DEFAULT_AI_ACCENT_COLOR,
  DEFAULT_AI_PRIMARY_COLOR,
  DEFAULT_AI_PRIMARY_COLOR_DARK,
} from './appinventor-validation.js';

const DEFAULT_USER = 'ai_tensor';

export function appInventorUserFromProperties(properties = {}) {
  const main = String(properties?.main || '').trim();
  const match = main.match(/^appinventor\.([A-Za-z][A-Za-z0-9_]*)\.[A-Za-z][A-Za-z0-9_]*\.[A-Za-z][A-Za-z0-9_]*$/);
  return match ? match[1] : DEFAULT_USER;
}

export function sourceBasePath(project, properties = {}) {
  return `src/appinventor/${appInventorUserFromProperties(properties)}/${project}`;
}

export function qualifiedMain(project, screen = 'Screen1', properties = {}) {
  return `appinventor.${appInventorUserFromProperties(properties)}.${project}.${screen}`;
}

export function decodeJavaPropertiesBytes(bytes) {
  if (typeof bytes === 'string') return bytes;
  if (typeof TextDecoder !== 'undefined') {
    return new TextDecoder('iso-8859-1').decode(bytes);
  }
  let out = '';
  const chunkSize = 8192;
  for (let i = 0; i < bytes.length; i += chunkSize) {
    out += String.fromCharCode(...bytes.slice(i, i + chunkSize));
  }
  return out;
}

export function parseProperties(text) {
  const props = {};
  for (const rawLine of logicalPropertyLines(text)) {
    let pos = 0;
    while (pos < rawLine.length && /[ \t\f]/.test(rawLine[pos])) pos += 1;
    if (pos >= rawLine.length || rawLine[pos] === '#' || rawLine[pos] === '!') continue;

    const { keyEnd, valueStart } = propertyKeyValueBounds(rawLine, pos);
    const key = rawLine.slice(pos, keyEnd);
    const value = rawLine.slice(valueStart);
    if (key) props[unescapePropertyText(key)] = unescapePropertyText(value);
  }
  return props;
}

function logicalPropertyLines(text) {
  const lines = String(text || '').replace(/\r\n/g, '\n').replace(/\r/g, '\n').split('\n');
  const logical = [];
  let current = '';

  for (const physicalLine of lines) {
    current = current
      ? current + physicalLine.replace(/^[ \t\f]*/, '')
      : physicalLine;
    if (hasOddTrailingBackslashes(current)) {
      current = current.slice(0, -1);
      continue;
    }
    logical.push(current);
    current = '';
  }
  if (current) logical.push(current);
  return logical;
}

function hasOddTrailingBackslashes(line) {
  let count = 0;
  for (let i = line.length - 1; i >= 0 && line[i] === '\\'; i -= 1) count += 1;
  return count % 2 === 1;
}

function propertyKeyValueBounds(line, start) {
  let escaped = false;
  let keyEnd = line.length;
  let separator = -1;
  for (let i = start; i < line.length; i += 1) {
    const ch = line[i];
    if (escaped) {
      escaped = false;
      continue;
    }
    if (ch === '\\') {
      escaped = true;
      continue;
    }
    if (ch === ':' || ch === '=' || /[ \t\f]/.test(ch)) {
      keyEnd = i;
      separator = i;
      break;
    }
  }

  let valueStart = separator === -1 ? line.length : separator;
  while (valueStart < line.length && /[ \t\f]/.test(line[valueStart])) valueStart += 1;
  if (valueStart < line.length && (line[valueStart] === ':' || line[valueStart] === '=')) valueStart += 1;
  while (valueStart < line.length && /[ \t\f]/.test(line[valueStart])) valueStart += 1;
  return { keyEnd, valueStart };
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
    if (next === 'u') {
      const hex = text.slice(i + 1, i + 5);
      if (!/^[0-9A-Fa-f]{4}$/.test(hex)) {
        throw new Error(`Malformed Java properties unicode escape "\\u${hex}"`);
      }
      out += String.fromCharCode(parseInt(hex, 16));
      i += 4;
    } else if (next === 'n') out += '\n';
    else if (next === 'r') out += '\r';
    else if (next === 't') out += '\t';
    else if (next === 'f') out += '\f';
    else out += next;
  }
  return out;
}

function escapePropertyText(value, { key = false } = {}) {
  let out = '';
  const text = String(value ?? '');
  for (let i = 0; i < text.length; i += 1) {
    const ch = text[i];
    const code = text.charCodeAt(i);
    if (ch === '\\') out += '\\\\';
    else if (ch === '\n') out += '\\n';
    else if (ch === '\r') out += '\\r';
    else if (ch === '\t') out += '\\t';
    else if (ch === '\f') out += '\\f';
    else if ((i === 0 && ch === ' ') || ch === ':' || ch === '=' || ch === '#' || ch === '!' || (key && /\s/.test(ch))) out += `\\${ch}`;
    else if (code < 0x20 || code > 0x7e) out += `\\u${code.toString(16).padStart(4, '0')}`;
    else out += ch;
  }
  return out;
}

export function projectPropertiesText(project, baseProperties = {}) {
  const aiaProperties = projectPropertiesToAiaProperties(baseProperties);
  const theme = aiaProperties.theme || 'AppTheme.Light.DarkActionBar';
  const actionBar = actionBarForTheme(theme);
  const props = {
    ...aiaProperties,
    main: qualifiedMain(project, 'Screen1', aiaProperties),
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
    'color.primary': aiaProperties['color.primary'] || DEFAULT_AI_PRIMARY_COLOR,
    'color.primary.dark': aiaProperties['color.primary.dark'] || DEFAULT_AI_PRIMARY_COLOR_DARK,
    'color.accent': aiaProperties['color.accent'] || DEFAULT_AI_ACCENT_COLOR,
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
