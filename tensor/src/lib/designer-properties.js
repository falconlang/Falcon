import { PROJECT_PROPERTY_NAMES } from './project-properties.js';
import {
  LISTVIEW_LAYOUTS,
  emptyListViewRow,
  listViewColumnsForLayout,
  listViewDataSummary,
  parseListViewData,
  pruneListViewDataForLayout,
  serializeListViewData,
} from './listview-data.js';

export const CUSTOM_DESIGNER_EDITORS = Object.freeze({
  ListViewAddData: Object.freeze({
    editorType: 'ListViewAddData',
    component: 'ListView',
    property: 'ListData',
    kind: 'dialog',
    label: 'List Data',
    triggerLabel: 'Edit rows',
    valueType: 'json-array',
    columnsForLayout: listViewColumnsForLayout,
    parse: parseListViewData,
    serialize: serializeListViewData,
    emptyRow: emptyListViewRow,
    pruneForLayout: pruneListViewDataForLayout,
    summary: listViewDataSummary,
    layoutProperty: 'ListViewLayout',
  }),
});

export const DESIGNER_ENUM_OPTIONS = Object.freeze({
  typeface: [
    { value: '0', label: 'Default' },
    { value: '1', label: 'Serif' },
    { value: '2', label: 'Sans serif' },
    { value: '3', label: 'Monospace' },
    { value: '4', label: 'Custom' },
  ],
  visibility: [
    { value: 'true', label: 'Show' },
    { value: 'false', label: 'Hide' },
    { value: 'responsive', label: 'Responsive' },
  ],
  textalignment: [
    { value: '0', label: 'Left' },
    { value: '1', label: 'Center' },
    { value: '2', label: 'Right' },
  ],
  button_shape: [
    { value: '0', label: 'Default' },
    { value: '1', label: 'Rounded' },
    { value: '2', label: 'Rectangular' },
    { value: '3', label: 'Oval' },
  ],
  horizontal_alignment: [
    { value: '1', label: 'Left' },
    { value: '3', label: 'Center' },
    { value: '2', label: 'Right' },
  ],
  vertical_alignment: [
    { value: '1', label: 'Top' },
    { value: '2', label: 'Center' },
    { value: '3', label: 'Bottom' },
  ],
  screen_orientation: [
    { value: 'unspecified', label: 'Unspecified' },
    { value: 'portrait', label: 'Portrait' },
    { value: 'landscape', label: 'Landscape' },
    { value: 'sensor', label: 'Sensor' },
  ],
  toast_length: [
    { value: '0', label: 'Short' },
    { value: '1', label: 'Long' },
  ],
  theme: [
    { value: 'Classic', label: 'Classic' },
    { value: 'Device Default', label: 'Device Default' },
    { value: 'Dark', label: 'Dark' },
    { value: 'Black', label: 'Black' },
    { value: 'AppTheme.Light.DarkActionBar', label: 'Light with action bar' },
  ],
  screen_animation: [
    { value: 'default', label: 'Default' },
    { value: 'fade', label: 'Fade' },
    { value: 'zoom', label: 'Zoom' },
    { value: 'slidehorizontal', label: 'Slide horizontal' },
    { value: 'slidevertical', label: 'Slide vertical' },
    { value: 'none', label: 'None' },
  ],
  sizing: [
    { value: 'Responsive', label: 'Responsive' },
    { value: 'Fixed', label: 'Fixed' },
  ],
  file_scope: [
    { value: 'App', label: 'App' },
    { value: 'Asset', label: 'Asset' },
    { value: 'Cache', label: 'Cache' },
    { value: 'Legacy', label: 'Legacy' },
    { value: 'Private', label: 'Private' },
    { value: 'Shared', label: 'Shared' },
  ],
  ListViewLayout: LISTVIEW_LAYOUTS.map(({ value, label }) => ({ value, label })),
});

export const COMPONENT_ALIASES = Object.freeze({
  Screen: 'Form',
});

const HIDDEN_FORM_PROPERTIES = new Set([
  ...PROJECT_PROPERTY_NAMES,
  'ActionBar',
  'BuildNumber',
  'PhonePreview',
  'PhoneTablet',
  'ProjectColors',
  'UsesLocation',
]);

const INTERNAL_PLACEMENT_PROPERTIES = new Set(['Column', 'Row']);

export function componentTypeName(ident) {
  const dotIdx = String(ident || '').indexOf('.');
  return dotIdx === -1 ? String(ident || '') : String(ident || '').slice(0, dotIdx);
}

export function optionsFromHelper(blockProp) {
  const opts = blockProp?.helper?.data?.options;
  if (!Array.isArray(opts)) return [];
  return opts
    .filter(opt => opt?.deprecated !== 'true')
    .map(opt => ({ value: String(opt.value), label: opt.name }));
}

export function optionsForDesignerEditorType(editorType, editorArgs = []) {
  if (editorType === 'choices') {
    return editorArgs.map(arg => ({ value: arg, label: arg }));
  }
  return DESIGNER_ENUM_OPTIONS[editorType] || [];
}

export function inferDesignerEditorType(propName, blockProp, designerProp) {
  if (propName === 'Height' || propName === 'Width') return 'layout_size';
  if (designerProp?.editorType) return designerProp.editorType;
  if (/(Color|Colour)$/.test(propName) || propName === 'PaintColor') return 'color';
  if (/Image|Icon|Picture|Sound|Source|FileName|ResponseFileName/.test(propName)) return 'asset';
  if (blockProp?.type === 'boolean') return 'boolean';
  if (blockProp?.type === 'number') return 'integer';
  if (blockProp?.type === 'list') return 'textArea';
  return 'string';
}

export function normalizePropertyCategory(category) {
  if (!category || category === 'Unspecified') return 'General';
  return category;
}

export function cleanDescription(description) {
  return String(description || '')
    .replace(/<br\s*\/?>/gi, ' ')
    .replace(/<\/p>/gi, ' ')
    .replace(/<[^>]+>/g, '')
    .replace(/&nbsp;/g, ' ')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&amp;/g, '&')
    .replace(/\s+/g, ' ')
    .trim();
}

export function isSettableBlockProperty(prop) {
  return prop?.rw !== 'invisible' && prop?.rw !== 'read-only';
}

export function shouldIncludeDesignerProperty(componentName, designerProp) {
  if (!designerProp?.name || INTERNAL_PLACEMENT_PROPERTIES.has(designerProp.name)) return false;
  if (componentName === 'Form' && HIDDEN_FORM_PROPERTIES.has(designerProp.name)) return false;
  return true;
}

export function customDesignerEditorMetadata(editorType) {
  return CUSTOM_DESIGNER_EDITORS[editorType] || null;
}

export function buildPropsForComponent(component) {
  const designerByName = new Map((component.properties || []).map(prop => [prop.name, prop]));
  const blockByName = new Map((component.blockProperties || []).map(prop => [prop.name, prop]));
  const orderedNames = [];

  for (const prop of component.properties || []) {
    if (!shouldIncludeDesignerProperty(component.name, prop)) continue;
    orderedNames.push(prop.name);
  }

  for (const prop of component.blockProperties || []) {
    if (!isSettableBlockProperty(prop)) continue;
    if (!orderedNames.includes(prop.name)) orderedNames.push(prop.name);
  }

  return orderedNames.map(name => {
    const designerProp = designerByName.get(name);
    const blockProp = blockByName.get(name);
    const editorType = inferDesignerEditorType(name, blockProp, designerProp);
    const editorArgs = designerProp?.editorArgs || [];
    const helperOptions = optionsFromHelper(blockProp);
    const editorOptions = optionsForDesignerEditorType(editorType, editorArgs);
    const customEditor = customDesignerEditorMetadata(editorType);

    return {
      name,
      editorType,
      category: normalizePropertyCategory(blockProp?.category),
      description: cleanDescription(blockProp?.description || designerProp?.description),
      defaultValue: designerProp?.defaultValue ?? '',
      options: helperOptions.length ? helperOptions : editorOptions,
      customEditor,
      designerOnly: Boolean(designerProp && blockProp?.rw === 'invisible'),
    };
  });
}

export function buildComponentProps(components = []) {
  const props = {};
  for (const component of components) {
    props[component.name] = buildPropsForComponent(component);
  }
  for (const [alias, target] of Object.entries(COMPONENT_ALIASES)) {
    props[alias] = props[target] || [];
  }
  return props;
}

export function customDesignerPropertyReport(components = []) {
  const records = [];
  for (const component of components) {
    for (const prop of component.properties || []) {
      const customEditor = customDesignerEditorMetadata(prop.editorType);
      if (!customEditor) continue;
      records.push({
        component: component.name,
        property: prop.name,
        editorType: prop.editorType,
        kind: customEditor.kind,
      });
    }
  }
  return records;
}

export function unknownDesignerEditorTypes(components = []) {
  const known = new Set([
    ...Object.keys(DESIGNER_ENUM_OPTIONS),
    ...Object.keys(CUSTOM_DESIGNER_EDITORS),
    'asset',
    'boolean',
    'button_shape',
    'color',
    'choices',
    'file_scope',
    'float',
    'horizontal_alignment',
    'integer',
    'layout_size',
    'non_negative_float',
    'non_negative_integer',
    'screen_animation',
    'screen_orientation',
    'sizing',
    'string',
    'text',
    'textArea',
    'textalignment',
    'theme',
    'toast_length',
    'typeface',
    'vertical_alignment',
    'visibility',
  ]);
  const uses = new Map();
  for (const component of components) {
    for (const prop of component.properties || []) {
      if (known.has(prop.editorType)) continue;
      const rows = uses.get(prop.editorType) || [];
      rows.push(`${component.name}.${prop.name}`);
      uses.set(prop.editorType, rows);
    }
  }
  return [...uses.entries()].map(([editorType, properties]) => ({ editorType, properties }));
}
