import simpleComponents from '../../../lang/code/compdb/simple_components.json' with { type: 'json' };

export const CURRENT_YA_VERSION = '233';
export const CURRENT_BLOCKS_LANGUAGE_VERSION = '37';

const SCM_JSON_PREFIX = '#|\n$JSON\n';
const SCM_JSON_SUFFIX = '\n|#';
const OLD_PROJECT_YA_VERSION = 150;

const COMPONENT_META = new Map(simpleComponents.map(component => [component.name, component]));
const TYPE_ALIASES = {
  Screen: 'Form',
  Form: 'Form',
};

const TYPE_RENAMES = [
  { beforeYa: 2, from: 'Logger', to: 'Notifier' },
  { beforeYa: 216, from: 'YandexTranslate', to: 'Translator' },
  { beforeYa: 218, from: 'GoogleSheets', to: 'Spreadsheet' },
  { beforeYa: 228, from: 'LineOfBestFit', to: 'Trendline' },
];

const PROPERTY_RENAMES = {
  Button: [{ before: 2, from: 'Alignment', to: 'TextAlignment' }],
  CheckBox: [{ before: 2, from: 'Value', to: 'Checked' }],
  ContactPicker: [{ before: 2, from: 'Alignment', to: 'TextAlignment' }],
  EmailPicker: [{ before: 2, from: 'Alignment', to: 'TextAlignment' }],
  ImagePicker: [
    { before: 2, from: 'Alignment', to: 'TextAlignment' },
    { before: 5, from: 'ImagePath', to: 'Selection' },
  ],
  Label: [{ before: 2, from: 'Alignment', to: 'TextAlignment' }],
  ListPicker: [{ before: 2, from: 'Alignment', to: 'TextAlignment' }],
  ListView: [{ before: 10, from: 'TextSize', to: 'FontSize' }],
  OrientationSensor: [{ before: 2, from: 'Yaw', to: 'Azimuth' }],
  PasswordTextBox: [{ before: 2, from: 'Alignment', to: 'TextAlignment' }],
  PhoneNumberPicker: [{ before: 2, from: 'Alignment', to: 'TextAlignment' }],
  Player: [{ before: 5, from: 'IsLooping', to: 'Loop' }],
  TextBox: [{ before: 3, from: 'Alignment', to: 'TextAlignment' }],
};

function cloneJson(value) {
  return JSON.parse(JSON.stringify(value ?? null));
}

function scmType(type) {
  return TYPE_ALIASES[type] || type;
}

function currentComponentVersion(type) {
  return COMPONENT_META.get(scmType(type))?.version || null;
}

function numericVersion(value, fallback = 0) {
  const version = Number.parseInt(String(value ?? ''), 10);
  return Number.isFinite(version) ? version : fallback;
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

function renameProperty(properties, from, to) {
  if (!Object.prototype.hasOwnProperty.call(properties, from)) return;
  properties[to] = properties[from];
  delete properties[from];
}

function unquoteAiString(value) {
  if (typeof value !== 'string' || value.length < 2) return value;
  if (!value.startsWith('"') || !value.endsWith('"')) return value;
  try {
    return JSON.parse(value);
  } catch {
    return value.slice(1, -1).replace(/\\"/g, '"').replace(/\\\\/g, '\\');
  }
}

function unquoteLegacyPropertyValues(properties) {
  for (const [key, value] of Object.entries(properties)) {
    if (key.startsWith('$') || key === 'Uuid') continue;
    if (typeof value === 'string') properties[key] = unquoteAiString(value);
  }
}

function knownPropertyNames(type) {
  const meta = COMPONENT_META.get(scmType(type));
  if (!meta) return null;
  const names = new Set(['$Name', '$Type', '$Version', '$Components', 'Uuid']);
  for (const property of meta.properties || []) names.add(property.name);
  for (const property of meta.blockProperties || []) names.add(property.name);
  return names;
}

function pruneUnsupportedKnownProperties(properties, type) {
  const names = knownPropertyNames(type);
  if (!names) return;
  for (const key of Object.keys(properties)) {
    if (!names.has(key)) delete properties[key];
  }
}

function directChildFillsParent(component) {
  for (const child of component.$Components || []) {
    if (String(child?.Height ?? '') === '-2') return true;
  }
  return false;
}

function upgradeFormProperties(properties, version) {
  if (version < 2 && directChildFillsParent(properties)) {
    properties.Scrollable = 'False';
  }
  if (version < 13) {
    if (Object.prototype.hasOwnProperty.call(properties, 'Scrollable')) {
      if (properties.Scrollable === 'False') delete properties.Scrollable;
    } else {
      properties.Scrollable = 'True';
    }
  }
  if (version < 17) {
    properties.CompatibilityMode = 'True';
  }
  if (version < 18) {
    if (Object.prototype.hasOwnProperty.call(properties, 'CompatibilityMode')) {
      delete properties.CompatibilityMode;
    } else {
      properties.Sizing = 'Responsive';
    }
  }
  if (version < 23) {
    if (Object.prototype.hasOwnProperty.call(properties, 'Theme') && properties.Theme !== 'Classic') {
      properties.ActionBar = 'True';
    } else if (Object.prototype.hasOwnProperty.call(properties, 'ActionBar')) {
      delete properties.ActionBar;
    }
  }
  if (version < 25) {
    if (Object.prototype.hasOwnProperty.call(properties, 'Sizing')) {
      if (properties.Sizing === 'Responsive') delete properties.Sizing;
    } else {
      properties.Sizing = 'Fixed';
    }
  }
  if (version < 26) {
    if (Object.prototype.hasOwnProperty.call(properties, 'ShowListsAsJson')) {
      if (properties.ShowListsAsJson === 'True') delete properties.ShowListsAsJson;
    } else {
      properties.ShowListsAsJson = 'False';
    }
  }
  if (version < 31) {
    if (Object.prototype.hasOwnProperty.call(properties, 'Theme')) {
      if (properties.Theme === 'AppTheme.Light.DarkActionBar') delete properties.Theme;
    } else {
      properties.Theme = 'Classic';
    }
  }
}

function cleanupImageSpriteOrigin(properties, version) {
  if (version >= 10) return;
  const markOrigin = String(properties.MarkOrigin ?? '');
  const match = markOrigin.match(/^\(([-+]?\d+(?:\.\d+)?),\s*([-+]?\d+(?:\.\d+)?)\)$/);
  if (match && Number(match[1]) === 0 && Number(match[2]) === 0) delete properties.MarkOrigin;
  if (Number(properties.OriginX) === 0) delete properties.OriginX;
  if (Number(properties.OriginY) === 0) delete properties.OriginY;
}

function applySpecialPropertyUpgrades(properties, type, version) {
  for (const rename of PROPERTY_RENAMES[type] || []) {
    if (version < rename.before) renameProperty(properties, rename.from, rename.to);
  }

  if (type === 'Camera' && version < 3) {
    delete properties.UseFront;
  } else if (type === 'Canvas' && version < 10) {
    properties.TextAlignment = '0';
  } else if (type === 'File' && version < 4 && Object.prototype.hasOwnProperty.call(properties, 'LegacyMode')) {
    delete properties.LegacyMode;
    properties.DefaultScope = 'Legacy';
  } else if (type === 'Form') {
    upgradeFormProperties(properties, version);
  } else if (type === 'ImageSprite') {
    cleanupImageSpriteOrigin(properties, version);
  } else if (type === 'Label' && version < 4) {
    properties.HasMargins = 'False';
  } else if (type === 'ListView' && version >= 8 && version < 9 && properties.ElementColor === '&H00FFFFFF') {
    delete properties.ElementColor;
  } else if (type === 'Marker' && version < 3) {
    delete properties.ShowShadow;
  } else if (type === 'TextBox' && version < 4) {
    properties.MultiLine = 'True';
  } else if (type === 'Texting' && version < 3 && Object.prototype.hasOwnProperty.call(properties, 'ReceivingEnabled')) {
    properties.ReceivingEnabled = String(properties.ReceivingEnabled).toLowerCase() === 'true' ? '2' : '1';
  }
}

function upgradeComponent(properties, srcYaVersion, path = '0') {
  if (!properties || typeof properties !== 'object') return;

  let type = scmType(properties.$Type || '');
  for (const rename of TYPE_RENAMES) {
    if (srcYaVersion < rename.beforeYa && type === rename.from) {
      type = rename.to;
      properties.$Type = rename.to;
      break;
    }
  }

  const currentVersion = currentComponentVersion(type);
  const sourceVersion = numericVersion(properties.$Version, 0);
  const effectiveVersion = sourceVersion === 0 ? 1 : sourceVersion;

  if (currentVersion) {
    applySpecialPropertyUpgrades(properties, type, effectiveVersion);
  }
  if (srcYaVersion < 26) {
    unquoteLegacyPropertyValues(properties);
  }

  if (!properties.$Name && path === '0') properties.$Name = 'Screen1';
  if (currentVersion) properties.$Version = String(currentVersion);
  if (!Object.prototype.hasOwnProperty.call(properties, 'Uuid')) {
    properties.Uuid = path === '0' ? '0' : stableUuid(`${path}:${type}:${properties.$Name || type}`);
  }

  pruneUnsupportedKnownProperties(properties, type);

  for (const [index, child] of (properties.$Components || []).entries()) {
    upgradeComponent(child, srcYaVersion, `${path}-${index}`);
  }
}

function addLegacyAuthUrlTags(scm, srcYaVersion, host) {
  if (Array.isArray(scm.authURL)) return;
  scm.authURL = [];
  if (srcYaVersion > OLD_PROJECT_YA_VERSION) scm.authURL.push('*UNKNOWN*');
  if (host && !scm.authURL.includes(host)) scm.authURL.push(host);
}

export function extractScmJson(text) {
  const source = String(text || '').trim();
  const match = source.match(/\$JSON\s*([\s\S]*?)\s*\|#/);
  const jsonText = match ? match[1].trim() : source;
  if (!jsonText.startsWith('{')) {
    throw new Error('Only JSON-format App Inventor .scm files are supported');
  }
  return JSON.parse(jsonText);
}

export function serializeScmJson(scm) {
  return `${SCM_JSON_PREFIX}${JSON.stringify(scm)}${SCM_JSON_SUFFIX}`;
}

export function upgradeLegacyScmObject(input, { host = '' } = {}) {
  const scm = cloneJson(input);
  if (!scm || typeof scm !== 'object') {
    throw new Error('Invalid App Inventor SCM JSON');
  }

  const before = JSON.stringify(scm);
  const srcYaVersion = numericVersion(scm.YaVersion, 0);
  if (!scm.Source) scm.Source = 'Form';
  addLegacyAuthUrlTags(scm, srcYaVersion, host);
  if (srcYaVersion <= numericVersion(CURRENT_YA_VERSION)) {
    upgradeComponent(scm.Properties, srcYaVersion);
    scm.YaVersion = CURRENT_YA_VERSION;
  }

  return {
    scm,
    upgraded: before !== JSON.stringify(scm),
    sourceYaVersion: srcYaVersion,
  };
}

export function upgradeLegacyScmText(text, options = {}) {
  const original = extractScmJson(text);
  const result = upgradeLegacyScmObject(original, options);
  return {
    ...result,
    text: serializeScmJson(result.scm),
    original,
  };
}

function fillMissingComponentVersions(properties) {
  if (!properties || typeof properties !== 'object') return;
  if (!Object.prototype.hasOwnProperty.call(properties, '$Version')) {
    properties.$Version = '1';
  }
  for (const child of properties.$Components || []) {
    fillMissingComponentVersions(child);
  }
}

export function formJsonForBlockUpgrade(scm) {
  const formJson = cloneJson(scm);
  fillMissingComponentVersions(formJson?.Properties);
  return JSON.stringify(formJson);
}

export function componentDefinitionsFromScmProperties(properties, defs = {}) {
  if (!properties || typeof properties !== 'object') return defs;
  const type = scmType(properties.$Type || '');
  const name = String(properties.$Name || '').trim();
  if (type && name) {
    const defType = type === 'Form' ? 'Screen' : type;
    defs[defType] ||= [];
    if (!defs[defType].includes(name)) defs[defType].push(name);
  }
  for (const child of properties.$Components || []) {
    componentDefinitionsFromScmProperties(child, defs);
  }
  return defs;
}
