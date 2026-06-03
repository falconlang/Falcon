import { get, writable } from 'svelte/store';
import simpleComponents from '../../../lang/code/compdb/simple_components.json' with { type: 'json' };

export const extensionComponentDescriptors = writable([]);

const BUILTIN_DESCRIPTORS = Object.freeze(simpleComponents.map(normalizeComponentDescriptor));
let currentExtensionDescriptors = [];

extensionComponentDescriptors.subscribe(value => {
  currentExtensionDescriptors = normalizeComponentDescriptors(value);
});

function arrayValue(value) {
  return Array.isArray(value) ? value : [];
}

export function normalizeComponentDescriptor(descriptor) {
  const source = descriptor && typeof descriptor === 'object' ? descriptor : {};
  const name = String(source.name || '').trim();
  const type = String(source.type || '').trim();
  if (!name || !type) return null;

  return {
    ...source,
    name,
    type,
    external: String(source.external ?? 'false'),
    version: String(source.version ?? '1'),
    categoryString: String(source.categoryString || (source.external === 'true' ? 'EXTENSION' : '')),
    helpString: String(source.helpString || name),
    showOnPalette: String(source.showOnPalette ?? 'true'),
    nonVisible: String(source.nonVisible ?? 'true'),
    iconName: String(source.iconName || ''),
    events: arrayValue(source.events),
    methods: arrayValue(source.methods),
    properties: arrayValue(source.properties),
    blockProperties: arrayValue(source.blockProperties),
  };
}

export function normalizeComponentDescriptors(descriptors) {
  const seen = new Set();
  const normalized = [];
  const list = Array.isArray(descriptors) ? descriptors : [];
  for (const descriptor of list) {
    const next = normalizeComponentDescriptor(descriptor);
    if (!next || seen.has(next.name)) continue;
    seen.add(next.name);
    normalized.push(next);
  }
  return normalized;
}

export function setProjectExtensionComponentDescriptors(descriptors) {
  extensionComponentDescriptors.set(normalizeComponentDescriptors(descriptors));
}

export function readComponentDescriptorFile(text) {
  const content = String(text || '').trim();
  if (!content) return [];
  const parsed = JSON.parse(content);
  return normalizeComponentDescriptors(Array.isArray(parsed) ? parsed : [parsed]);
}

export function projectExtensionComponentDescriptors() {
  return get(extensionComponentDescriptors);
}

export function allComponentDescriptors(extraDescriptors = currentExtensionDescriptors) {
  return [
    ...BUILTIN_DESCRIPTORS,
    ...normalizeComponentDescriptors(extraDescriptors),
  ].filter(Boolean);
}

export function allComponentDescriptorsJsonText(extraDescriptors = currentExtensionDescriptors) {
  const extensions = normalizeComponentDescriptors(extraDescriptors);
  if (!extensions.length) return JSON.stringify(BUILTIN_DESCRIPTORS);
  return JSON.stringify(allComponentDescriptors(extensions));
}

export function componentMetaMap(extraDescriptors = currentExtensionDescriptors) {
  return new Map(allComponentDescriptors(extraDescriptors).map(component => [component.name, component]));
}

export function componentTypeAliases(extraDescriptors = currentExtensionDescriptors) {
  const aliases = new Map();
  for (const descriptor of allComponentDescriptors(extraDescriptors)) {
    aliases.set(descriptor.name, descriptor.name);
    aliases.set(descriptor.type, descriptor.name);
    const simpleName = descriptor.type.split('.').pop();
    if (simpleName) aliases.set(simpleName, descriptor.name);
  }
  if (aliases.has('Form')) aliases.set('Screen', 'Form');
  return aliases;
}

export function componentDescriptorName(typeName, extraDescriptors = currentExtensionDescriptors) {
  const clean = String(typeName || '').trim();
  if (!clean) return '';
  return componentTypeAliases(extraDescriptors).get(clean) || clean;
}

export function knownComponentTypeSet(extraDescriptors = currentExtensionDescriptors) {
  return new Set([...allComponentDescriptors(extraDescriptors).map(component => component.name), 'Screen']);
}
