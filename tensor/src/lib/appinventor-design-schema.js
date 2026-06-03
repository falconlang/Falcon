import { componentMetaMap } from './appinventor-component-registry.js';
import { PROJECT_PROPERTY_NAMES } from './project-properties.js';

const TYPE_ALIASES = {
  Screen: 'Form',
  Form: 'Form',
};

const HIDDEN_FORM_SCHEMA_PROPERTIES = new Set([
  ...PROJECT_PROPERTY_NAMES,
  'ActionBar',
  'BuildNumber',
  'PhonePreview',
  'PhoneTablet',
  'ProjectColors',
  'UsesLocation',
]);

function scmType(type) {
  return TYPE_ALIASES[type] || type;
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
  const component = componentMetaMap().get(scmType(componentType));
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

export function scmComponentToSchema(component, depth = 0) {
  const indent = '  '.repeat(depth);
  const type = component.$Type === 'Form' ? 'Screen' : component.$Type;
  const name = component.$Name || type;
  const props = Object.entries(component)
    .filter(([key]) => {
      if (key.startsWith('$') || key === 'Uuid') return false;
      if (depth === 0 && scmType(type) === 'Form' && HIDDEN_FORM_SCHEMA_PROPERTIES.has(key)) return false;
      return true;
    })
    .map(([key, value]) => `${indent}  ${key}: ${schemaValueText(value, type, key)}`);
  const children = (component.$Components || [])
    .map(child => scmComponentToSchema(child, depth + 1));
  const body = [...props, ...children];

  if (!body.length) return `${indent}${type}.${name}`;
  return `${indent}${type}.${name} {\n${body.join(',\n')}\n${indent}}`;
}

export function scmToDesignSchema(text, fallbackScreenName) {
  const scm = extractScmJson(text);
  const properties = scm?.Properties;
  if (!properties || typeof properties !== 'object') {
    throw new Error(`Unable to read designer properties for ${fallbackScreenName}`);
  }
  if (!properties.$Name) properties.$Name = fallbackScreenName;
  return scmComponentToSchema(properties);
}
