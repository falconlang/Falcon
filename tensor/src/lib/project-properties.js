import {
  DEFAULT_AI_ACCENT_COLOR,
  DEFAULT_AI_PRIMARY_COLOR,
  DEFAULT_AI_PRIMARY_COLOR_DARK,
} from './appinventor-validation.js';

const PROJECT_PROPERTY_CATEGORY_ORDER = ['General', 'Theming', 'Publishing', 'iOS Settings'];

export const PROJECT_PROPERTY_CATEGORIES = Object.freeze([
  { id: 'General', label: 'General' },
  { id: 'Theming', label: 'Theming' },
  { id: 'Publishing', label: 'Publishing' },
  { id: 'iOS Settings', label: 'iOS Settings' },
]);

export const PROJECT_PROPERTY_DEFINITIONS = Object.freeze([
  {
    name: 'AppName',
    category: 'General',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'aname',
    writeEmpty: true,
    description: 'This is the display name of the installed application in the phone. If the AppName is blank, it will be set to the name of the project when the project is built.',
  },
  {
    name: 'Icon',
    category: 'General',
    editorType: 'asset',
    defaultValue: '',
    aiaKey: 'icon',
    description: 'The image used for your App display icon should be a square png or jpeg image with dimensions up to 1024x1024 pixels.',
  },
  {
    name: 'DefaultFileScope',
    category: 'General',
    editorType: 'file_scope',
    defaultValue: 'App',
    aiaKey: 'defaultfilescope',
    options: ['App', 'Asset', 'Cache', 'Legacy', 'Private', 'Shared'],
    description: 'Specifies the default scope used when components access files.',
  },
  {
    name: 'ShowListsAsJson',
    category: 'General',
    editorType: 'boolean',
    defaultValue: 'True',
    aiaKey: 'showlistsasjson',
    alwaysSend: true,
    description: 'Controls whether lists are displayed as JSON/Python notation or Lisp notation.',
  },
  {
    name: 'Sizing',
    category: 'General',
    editorType: 'sizing',
    defaultValue: 'Responsive',
    aiaKey: 'sizing',
    alwaysSend: true,
    options: ['Fixed', 'Responsive'],
    description: 'Controls whether screen layouts use a fixed-size screen or the device resolution.',
  },
  {
    name: 'TutorialURL',
    category: 'General',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'tutorialurl',
    description: 'A URL to use to populate the Tutorial Sidebar while editing a project.',
  },
  {
    name: 'BlocksToolkit',
    category: 'General',
    editorType: 'subset_json',
    defaultValue: '',
    aiaKey: 'subsetjson',
    uiVisible: false,
    description: 'JSON controlling which components, designer properties, and blocks are available in the project.',
  },
  {
    name: 'PrimaryColor',
    category: 'Theming',
    editorType: 'color',
    defaultValue: DEFAULT_AI_PRIMARY_COLOR,
    aiaKey: 'color.primary',
    description: 'The primary color used for Material UI elements, such as the ActionBar.',
  },
  {
    name: 'PrimaryColorDark',
    category: 'Theming',
    editorType: 'color',
    defaultValue: DEFAULT_AI_PRIMARY_COLOR_DARK,
    aiaKey: 'color.primary.dark',
    description: 'The primary color used for darker Material UI elements.',
  },
  {
    name: 'AccentColor',
    category: 'Theming',
    editorType: 'color',
    defaultValue: DEFAULT_AI_ACCENT_COLOR,
    aiaKey: 'color.accent',
    description: 'The accent color used for highlights and other user interface accents.',
  },
  {
    name: 'Theme',
    category: 'Theming',
    editorType: 'theme',
    defaultValue: 'AppTheme.Light.DarkActionBar',
    aiaKey: 'theme',
    options: ['Classic', 'AppTheme.Light.DarkActionBar', 'AppTheme.Light', 'AppTheme'],
    description: 'Sets the theme used by the application.',
  },
  {
    name: 'VersionCode',
    category: 'Publishing',
    editorType: 'non_negative_integer',
    defaultValue: '1',
    aiaKey: 'versioncode',
    description: 'An integer value that should be incremented each time a new APK is created for distribution.',
  },
  {
    name: 'VersionName',
    category: 'Publishing',
    editorType: 'string',
    defaultValue: '1.0',
    aiaKey: 'versionname',
    description: 'A string that lets users distinguish between different versions of the app.',
  },
  {
    name: 'NSBluetoothAlwaysUsageDescription',
    category: 'iOS Settings',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'NSBluetoothAlwaysUsageDescription',
    description: 'Specifies the iOS privacy explanation for Bluetooth access on iOS 13 and later.',
  },
  {
    name: 'NSBluetoothPeripheralUsageDescription',
    category: 'iOS Settings',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'NSBluetoothPeripheralUsageDescription',
    description: 'Specifies the iOS privacy explanation for Bluetooth peripheral access before iOS 13.',
  },
  {
    name: 'NSContactsUsageDescription',
    category: 'iOS Settings',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'NSContactsUsageDescription',
    description: 'Specifies the iOS privacy explanation for Contacts access.',
  },
  {
    name: 'NSMicrophoneUsageDescription',
    category: 'iOS Settings',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'NSMicrophoneUsageDescription',
    description: 'Specifies the iOS privacy explanation for microphone access.',
  },
  {
    name: 'NSCameraUsageDescription',
    category: 'iOS Settings',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'NSCameraUsageDescription',
    description: 'Specifies the iOS privacy explanation for camera access.',
  },
  {
    name: 'NSSpeechRecognitionUsageDescription',
    category: 'iOS Settings',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'NSSpeechRecognitionUsageDescription',
    description: 'Specifies the iOS privacy explanation for speech recognition access.',
  },
  {
    name: 'NSLocationWhenInUseUsageDescription',
    category: 'iOS Settings',
    editorType: 'string',
    defaultValue: '',
    aiaKey: 'NSLocationWhenInUseUsageDescription',
    description: 'Specifies the iOS privacy explanation for user location access.',
  },
].map(definition => Object.freeze({
  uiVisible: true,
  includeInScm: true,
  alwaysSend: false,
  writeEmpty: false,
  ...definition,
})));

export const PROJECT_PROPERTY_NAMES = Object.freeze(PROJECT_PROPERTY_DEFINITIONS.map(definition => definition.name));
export const VISIBLE_PROJECT_PROPERTY_DEFINITIONS = Object.freeze(
  PROJECT_PROPERTY_DEFINITIONS.filter(definition => definition.uiVisible)
);

const DEFINITIONS_BY_NAME = new Map(PROJECT_PROPERTY_DEFINITIONS.map(definition => [definition.name, definition]));
const NAME_ALIASES = new Map();

for (const definition of PROJECT_PROPERTY_DEFINITIONS) {
  NAME_ALIASES.set(definition.name.toLowerCase(), definition.name);
  NAME_ALIASES.set(definition.aiaKey.toLowerCase(), definition.name);
}

NAME_ALIASES.set('blockstoolkit', 'BlocksToolkit');
NAME_ALIASES.set('blocks_toolkit', 'BlocksToolkit');
NAME_ALIASES.set('default_file_scope', 'DefaultFileScope');
NAME_ALIASES.set('show_lists_as_json', 'ShowListsAsJson');

function sortedDefinitions(definitions = PROJECT_PROPERTY_DEFINITIONS) {
  return [...definitions].sort((a, b) => {
    const categoryOrder = PROJECT_PROPERTY_CATEGORY_ORDER.indexOf(a.category)
      - PROJECT_PROPERTY_CATEGORY_ORDER.indexOf(b.category);
    return categoryOrder || PROJECT_PROPERTY_DEFINITIONS.indexOf(a) - PROJECT_PROPERTY_DEFINITIONS.indexOf(b);
  });
}

export function projectPropertyDefinition(name) {
  return DEFINITIONS_BY_NAME.get(normalizeProjectPropertyName(name));
}

export function normalizeProjectPropertyName(name) {
  if (name == null) return null;
  return NAME_ALIASES.get(String(name).trim().toLowerCase()) || null;
}

export function projectPropertiesByCategory({ includeHidden = true } = {}) {
  const definitions = includeHidden
    ? PROJECT_PROPERTY_DEFINITIONS
    : VISIBLE_PROJECT_PROPERTY_DEFINITIONS;
  return PROJECT_PROPERTY_CATEGORIES.map(category => ({
    ...category,
    properties: sortedDefinitions(definitions).filter(definition => definition.category === category.id),
  }));
}

function normalizeBoolean(value, defaultValue) {
  if (typeof value === 'boolean') return value ? 'True' : 'False';
  const text = String(value ?? '').trim();
  if (/^(true|false)$/i.test(text)) return text.toLowerCase() === 'true' ? 'True' : 'False';
  if (text === '1') return 'True';
  if (text === '0') return 'False';
  return defaultValue;
}

function normalizeNonNegativeInteger(value, defaultValue) {
  const text = String(value ?? '').trim();
  if (/^\d+$/.test(text)) return String(Number(text));
  return defaultValue;
}

function normalizeColor(value, defaultValue) {
  const text = String(value ?? '').trim();
  if (!text) return defaultValue;
  const hex = text.replace(/^#/, '');
  if (/^[0-9A-Fa-f]{6}$/.test(hex)) return `&HFF${hex.toUpperCase()}`;
  if (/^[0-9A-Fa-f]{8}$/.test(hex)) return `&H${hex.toUpperCase()}`;
  if (/^&H[0-9A-Fa-f]{8}$/.test(text)) return text.toUpperCase();
  return text;
}

export function normalizeProjectPropertyValue(name, value) {
  const definition = projectPropertyDefinition(name);
  if (!definition) return String(value ?? '');

  if (definition.editorType === 'boolean') {
    return normalizeBoolean(value, definition.defaultValue);
  }

  if (definition.editorType === 'non_negative_integer') {
    return normalizeNonNegativeInteger(value, definition.defaultValue);
  }

  if (definition.editorType === 'color') {
    return normalizeColor(value, definition.defaultValue);
  }

  const text = String(value ?? '').trim();
  if (!text && definition.defaultValue) return definition.defaultValue;
  return text;
}

function readProjectPropertyValue(source, definition) {
  if (!source || typeof source !== 'object') return undefined;
  const directKeys = [definition.name, definition.aiaKey];
  for (const key of directKeys) {
    if (Object.prototype.hasOwnProperty.call(source, key)) return source[key];
  }

  const wanted = new Set(directKeys.map(key => key.toLowerCase()));
  for (const [key, value] of Object.entries(source)) {
    if (wanted.has(key.toLowerCase())) return value;
  }
  return undefined;
}

export function defaultProjectProperties(overrides = {}) {
  return normalizeProjectProperties(overrides);
}

export function normalizeProjectProperties(source = {}, {
  includeDefaults = true,
  preserveUnknown = true,
} = {}) {
  const input = source && typeof source === 'object' ? source : {};
  const result = {};

  if (preserveUnknown) {
    for (const [key, value] of Object.entries(input)) {
      if (!normalizeProjectPropertyName(key)) {
        result[key] = String(value ?? '');
      }
    }
  }

  for (const definition of PROJECT_PROPERTY_DEFINITIONS) {
    const value = readProjectPropertyValue(input, definition);
    if (value !== undefined) {
      result[definition.name] = normalizeProjectPropertyValue(definition.name, value);
    } else if (includeDefaults) {
      result[definition.name] = definition.defaultValue;
    }
  }

  return result;
}

export function withProjectPropertyValue(source, name, value) {
  const propertyName = normalizeProjectPropertyName(name);
  if (!propertyName) {
    throw new Error(`Unknown project property "${name}"`);
  }
  return {
    ...normalizeProjectProperties(source),
    [propertyName]: normalizeProjectPropertyValue(propertyName, value),
  };
}

export function projectPropertiesToAiaProperties(source = {}, { includeDefaults = true } = {}) {
  const input = source && typeof source === 'object' ? source : {};
  const normalized = normalizeProjectProperties(input, { includeDefaults, preserveUnknown: true });
  const result = {};

  for (const [key, value] of Object.entries(input)) {
    if (!normalizeProjectPropertyName(key)) {
      result[key] = String(value ?? '');
    }
  }

  for (const definition of PROJECT_PROPERTY_DEFINITIONS) {
    const value = normalized[definition.name];
    if (value == null) continue;
    if (value === '' && !definition.writeEmpty) continue;
    if (!includeDefaults && value === definition.defaultValue && !definition.alwaysSend) continue;
    result[definition.aiaKey] = value;
  }

  return result;
}

export function actionBarForTheme(theme) {
  return String(theme || '') === 'Classic' ? 'False' : 'True';
}

export function projectPropertiesForScm(source = {}, projectName = '') {
  const normalized = normalizeProjectProperties(source, { includeDefaults: true, preserveUnknown: false });
  const result = {};
  for (const definition of PROJECT_PROPERTY_DEFINITIONS) {
    if (!definition.includeInScm) continue;
    const value = normalized[definition.name];
    if (value == null) continue;
    if (value === '' && !definition.writeEmpty && !definition.alwaysSend) continue;
    result[definition.name] = value;
  }

  if (!Object.prototype.hasOwnProperty.call(result, 'AppName')) {
    result.AppName = projectName;
  }
  result.ActionBar = actionBarForTheme(normalized.Theme);
  return result;
}

export function applyProjectPropertiesToScmProperties(properties = {}, projectProperties = {}, projectName = '') {
  const result = { ...(properties || {}) };
  const values = projectPropertiesForScm(projectProperties, projectName);

  for (const definition of PROJECT_PROPERTY_DEFINITIONS) {
    delete result[definition.name];
    delete result[definition.aiaKey];
  }

  for (const [key, value] of Object.entries(values)) {
    result[key] = value;
  }

  return result;
}

export function extractProjectPropertiesFromScmProperties(properties = {}) {
  if (!properties || typeof properties !== 'object') return {};
  const result = {};
  for (const definition of PROJECT_PROPERTY_DEFINITIONS) {
    if (Object.prototype.hasOwnProperty.call(properties, definition.name)) {
      result[definition.name] = normalizeProjectPropertyValue(definition.name, properties[definition.name]);
    }
  }
  return result;
}
