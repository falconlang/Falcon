import simpleComponents from '../../../lang/code/compdb/simple_components.json' with { type: 'json' };

export const DEFAULT_AI_PRIMARY_COLOR_DARK = '&HFF303F9F';
export const DEFAULT_AI_PRIMARY_COLOR = '&HFF3F51B5';
export const DEFAULT_AI_ACCENT_COLOR = '&HFFFF4081';

const MAX_ASSET_FILENAME_LENGTH = 100;

const JAVA_NAMES = new Set([
  'abstract', 'continue', 'for', 'new', 'switch', 'assert', 'default', 'goto',
  'package', 'synchronized', 'boolean', 'do', 'if', 'private', 'this', 'break',
  'double', 'implements', 'protected', 'throw', 'byte', 'else', 'import',
  'public', 'throws', 'case', 'enum', 'instanceof', 'return', 'transient',
  'catch', 'extends', 'int', 'short', 'try', 'char', 'final', 'interface',
  'static', 'void', 'class', 'finally', 'long', 'strictfp', 'volatile',
  'const', 'float', 'native', 'super', 'while',
]);

const YAIL_NAMES = new Set([
  'CsvUtil', 'Double', 'Float', 'Integer', 'JavaCollection', 'JavaIterator',
  'KawaEnvironment', 'Long', 'Short', 'SimpleForm', 'String', 'Pattern',
  'YailDictionary', 'YailList', 'YailNumberToString', 'YailRuntimeError',
]);

const SCHEME_NAMES = new Set(['begin', 'def', 'foreach', 'forrange', 'JavaStringUtils', 'quote']);
const COMPONENT_TYPE_NAMES = new Set(simpleComponents.map(component => component.name));

export function isReservedAppInventorName(name) {
  const text = String(name || '');
  return JAVA_NAMES.has(text) || YAIL_NAMES.has(text) || SCHEME_NAMES.has(text);
}

export function isAppInventorIdentifier(name) {
  return /^[A-Za-z][A-Za-z0-9_]*$/.test(String(name || ''));
}

export function isValidScreenName(name) {
  const text = String(name || '');
  return isAppInventorIdentifier(text) && !isReservedAppInventorName(text);
}

export function isValidComponentName(name) {
  const text = String(name || '');
  return isAppInventorIdentifier(text)
    && !isReservedAppInventorName(text)
    && !COMPONENT_TYPE_NAMES.has(text);
}

export function appInventorNameError(name, { component = false } = {}) {
  const text = String(name || '');
  if (!text) return 'Name is required.';
  if (!isAppInventorIdentifier(text)) {
    return 'Must start with a letter and contain only letters, numbers, or underscores.';
  }
  if (isReservedAppInventorName(text)) return 'Name is reserved by App Inventor.';
  if (component && COMPONENT_TYPE_NAMES.has(text)) return 'Name cannot match a component type.';
  return '';
}

export function normalizeAppInventorAssetName(name) {
  return String(name || '').trim().replace(/[\\/]+/g, '-');
}

export function isValidAppInventorAssetName(name) {
  const text = String(name || '');
  if (!text || text.length > MAX_ASSET_FILENAME_LENGTH) return false;
  if (text.includes("'") || /[/:\\]/.test(text)) return false;
  for (let i = 0; i < text.length; i += 1) {
    const code = text.charCodeAt(i);
    if (code < 0x20 || code > 0x7e) return false;
  }
  return encodeURIComponent(text) === text;
}

export function appInventorAssetNameError(name) {
  const text = String(name || '');
  if (!text) return 'Name is required.';
  if (text.length > MAX_ASSET_FILENAME_LENGTH) {
    return `Asset names must be ${MAX_ASSET_FILENAME_LENGTH} characters or fewer.`;
  }
  if (!isValidAppInventorAssetName(text)) {
    return 'Use printable ASCII only and avoid path, quote, or URL-reserved characters.';
  }
  return '';
}
