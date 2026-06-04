<script>
  import { createEventDispatcher, tick, onMount, onDestroy } from 'svelte';
  import { addDesignAsset, deleteDesignAsset, designAssets, renameDesignAsset, designerSearchIndex, designerTreeActive, searchNavigation } from './stores.js';
  import {
    customDesignerEditorMetadata,
    optionsForDesignerEditorType,
    shouldIncludeDesignerProperty,
  } from './designer-properties.js';
  import {
    emptyListViewRow,
    listViewColumnsForLayout,
    listViewDataSummary,
    parseListViewData,
    pruneListViewDataForLayout,
    renameListViewDataAsset,
    serializeListViewData,
  } from './listview-data.js';
  import {
    appInventorAssetNameError,
    appInventorNameError,
    isValidComponentName,
    isValidScreenName,
    normalizeAppInventorAssetName,
  } from './appinventor-validation.js';
  import {
    allComponentDescriptors,
    componentMetaMap,
    extensionComponentDescriptors,
  } from './appinventor-component-registry.js';
  import {
    parseDesignSchemaResult as parseSharedDesignSchemaResult,
    serializeDesignTree as serializeSharedDesignTree,
  } from './design-schema-tree.js';

  export let schemaValue = '';
  const dispatch = createEventDispatcher();

  // ── Component property database ────────────────────────────────────
  const LEGACY_COMP_PROPS = {
    Screen: [
      { name: 'Title',             editorType: 'string',             category: 'General'    },
      { name: 'BackgroundColor',   editorType: 'color',              category: 'Appearance' },
      { name: 'BackgroundImage',   editorType: 'asset',              category: 'Appearance' },
      { name: 'Scrollable',        editorType: 'boolean',            category: 'Behavior'   },
      { name: 'ScreenOrientation', editorType: 'screen_orientation', category: 'Behavior'   },
      { name: 'Theme',             editorType: 'theme',              category: 'General'    },
    ],
    Button: [
      { name: 'BackgroundColor', editorType: 'color',              category: 'Appearance' },
      { name: 'Enabled',         editorType: 'boolean',            category: 'Behavior'   },
      { name: 'FontBold',        editorType: 'boolean',            category: 'Appearance' },
      { name: 'FontItalic',      editorType: 'boolean',            category: 'Appearance' },
      { name: 'FontSize',        editorType: 'non_negative_float', category: 'Appearance' },
      { name: 'FontTypeface',    editorType: 'typeface',           category: 'Appearance' },
      { name: 'Height',          editorType: 'layout_size',        category: 'Appearance' },
      { name: 'Image',           editorType: 'asset',              category: 'Appearance' },
      { name: 'Shape',           editorType: 'button_shape',       category: 'Appearance' },
      { name: 'ShowFeedback',    editorType: 'boolean',            category: 'Appearance' },
      { name: 'Text',            editorType: 'string',             category: 'Appearance' },
      { name: 'TextAlignment',   editorType: 'textalignment',      category: 'Appearance' },
      { name: 'TextColor',       editorType: 'color',              category: 'Appearance' },
      { name: 'Visible',         editorType: 'visibility',         category: 'Appearance' },
      { name: 'Width',           editorType: 'layout_size',        category: 'Appearance' },
    ],
    Label: [
      { name: 'BackgroundColor', editorType: 'color',              category: 'Appearance' },
      { name: 'FontBold',        editorType: 'boolean',            category: 'Appearance' },
      { name: 'FontItalic',      editorType: 'boolean',            category: 'Appearance' },
      { name: 'FontSize',        editorType: 'non_negative_float', category: 'Appearance' },
      { name: 'FontTypeface',    editorType: 'typeface',           category: 'Appearance' },
      { name: 'HasMargins',      editorType: 'boolean',            category: 'Appearance' },
      { name: 'Height',          editorType: 'layout_size',        category: 'Appearance' },
      { name: 'HTMLFormat',      editorType: 'boolean',            category: 'Behavior'   },
      { name: 'Text',            editorType: 'string',             category: 'Appearance' },
      { name: 'TextAlignment',   editorType: 'textalignment',      category: 'Appearance' },
      { name: 'TextColor',       editorType: 'color',              category: 'Appearance' },
      { name: 'Visible',         editorType: 'visibility',         category: 'Appearance' },
      { name: 'Width',           editorType: 'layout_size',        category: 'Appearance' },
    ],
    TextBox: [
      { name: 'BackgroundColor', editorType: 'color',              category: 'Appearance' },
      { name: 'Enabled',         editorType: 'boolean',            category: 'Behavior'   },
      { name: 'FontBold',        editorType: 'boolean',            category: 'Appearance' },
      { name: 'FontItalic',      editorType: 'boolean',            category: 'Appearance' },
      { name: 'FontSize',        editorType: 'non_negative_float', category: 'Appearance' },
      { name: 'FontTypeface',    editorType: 'typeface',           category: 'Appearance' },
      { name: 'Height',          editorType: 'layout_size',        category: 'Appearance' },
      { name: 'Hint',            editorType: 'string',             category: 'Appearance' },
      { name: 'MultiLine',       editorType: 'boolean',            category: 'Behavior'   },
      { name: 'NumbersOnly',     editorType: 'boolean',            category: 'Behavior'   },
      { name: 'Text',            editorType: 'string',             category: 'Appearance' },
      { name: 'TextAlignment',   editorType: 'textalignment',      category: 'Appearance' },
      { name: 'TextColor',       editorType: 'color',              category: 'Appearance' },
      { name: 'Visible',         editorType: 'visibility',         category: 'Appearance' },
      { name: 'Width',           editorType: 'layout_size',        category: 'Appearance' },
    ],
    HorizontalArrangement: [
      { name: 'AlignHorizontal', editorType: 'horizontal_alignment', category: 'Appearance' },
      { name: 'AlignVertical',   editorType: 'vertical_alignment',   category: 'Appearance' },
      { name: 'BackgroundColor', editorType: 'color',                category: 'Appearance' },
      { name: 'Height',          editorType: 'layout_size',          category: 'Appearance' },
      { name: 'Image',           editorType: 'asset',                category: 'Appearance' },
      { name: 'Visible',         editorType: 'visibility',           category: 'Appearance' },
      { name: 'Width',           editorType: 'layout_size',          category: 'Appearance' },
    ],
    VerticalArrangement: [
      { name: 'AlignHorizontal', editorType: 'horizontal_alignment', category: 'Appearance' },
      { name: 'AlignVertical',   editorType: 'vertical_alignment',   category: 'Appearance' },
      { name: 'BackgroundColor', editorType: 'color',                category: 'Appearance' },
      { name: 'Height',          editorType: 'layout_size',          category: 'Appearance' },
      { name: 'Image',           editorType: 'asset',                category: 'Appearance' },
      { name: 'Visible',         editorType: 'visibility',           category: 'Appearance' },
      { name: 'Width',           editorType: 'layout_size',          category: 'Appearance' },
    ],
    Notifier: [
      { name: 'BackgroundColor', editorType: 'color',        category: 'Appearance' },
      { name: 'NotifierLength',  editorType: 'toast_length', category: 'Behavior'   },
      { name: 'TextColor',       editorType: 'color',        category: 'Appearance' },
    ],
    ListView: [
      { name: 'BackgroundColor', editorType: 'color',              category: 'Appearance' },
      { name: 'Elements',        editorType: 'textArea',           category: 'Behavior'   },
      { name: 'FontSize',        editorType: 'non_negative_float', category: 'Appearance' },
      { name: 'Height',          editorType: 'layout_size',        category: 'Appearance' },
      { name: 'SelectionColor',  editorType: 'color',              category: 'Appearance' },
      { name: 'ShowFilterBar',   editorType: 'boolean',            category: 'Behavior'   },
      { name: 'TextColor',       editorType: 'color',              category: 'Appearance' },
      { name: 'Visible',         editorType: 'visibility',         category: 'Appearance' },
      { name: 'Width',           editorType: 'layout_size',        category: 'Appearance' },
    ],
  };

  const ENUM_OPTIONS = {
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
  };

  const COMPONENT_ALIASES = {
    Screen: 'Form',
  };

  function optionsFromHelper(blockProp) {
    const opts = blockProp?.helper?.data?.options;
    if (!Array.isArray(opts)) return [];
    return opts
      .filter(opt => opt?.deprecated !== 'true')
      .map(opt => ({ value: String(opt.value), label: opt.name }));
  }

  function inferEditorType(propName, blockProp, designerProp) {
    if (propName === 'Height' || propName === 'Width') return 'layout_size';
    if (designerProp?.editorType) return designerProp.editorType;
    if (/(Color|Colour)$/.test(propName) || propName === 'PaintColor') return 'color';
    if (/Image|Icon|Picture|Sound|Source|FileName|ResponseFileName/.test(propName)) return 'asset';
    if (blockProp?.type === 'boolean') return 'boolean';
    if (blockProp?.type === 'number') return 'integer';
    if (blockProp?.type === 'list') return 'textArea';
    return 'string';
  }

  function normalizePropertyCategory(category) {
    if (!category || category === 'Unspecified') return 'General';
    return category;
  }

  function cleanDescription(description) {
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

  function isSettableBlockProperty(prop) {
    return prop?.rw !== 'invisible' && prop?.rw !== 'read-only';
  }

  function buildPropsForComponent(component) {
    const designerByName = new Map((component.properties || []).map(prop => [prop.name, prop]));
    const blockByName = new Map((component.blockProperties || []).map(prop => [prop.name, prop]));
    const orderedNames = [];

    for (const prop of component.blockProperties || []) {
      if (!isSettableBlockProperty(prop)) continue;
      orderedNames.push(prop.name);
    }

    for (const prop of component.properties || []) {
      if (!shouldIncludeDesignerProperty(component.name, prop)) continue;
      if (!orderedNames.includes(prop.name)) orderedNames.push(prop.name);
    }

    return orderedNames.map(name => {
      const designerProp = designerByName.get(name);
      const blockProp = blockByName.get(name);
      const options = optionsFromHelper(blockProp);
      const editorType = inferEditorType(name, blockProp, designerProp);
      const editorArgs = designerProp?.editorArgs || [];
      const fallbackOptions = optionsForDesignerEditorType(editorType, editorArgs);
      const customEditor = customDesignerEditorMetadata(editorType);

      return {
        name,
        editorType,
        category: normalizePropertyCategory(blockProp?.category),
        description: cleanDescription(blockProp?.description || designerProp?.description),
        defaultValue: designerProp?.defaultValue ?? '',
        options: options.length ? options : fallbackOptions,
        customEditor,
        designerOnly: Boolean(designerProp && blockProp?.rw === 'invisible'),
      };
    });
  }

  function buildComponentProps(components) {
    const props = {};
    for (const component of components) {
      props[component.name] = buildPropsForComponent(component);
    }
    for (const [alias, target] of Object.entries(COMPONENT_ALIASES)) {
      props[alias] = props[target] || [];
    }
    return props;
  }

  let componentDescriptors = allComponentDescriptors();
  let COMP_PROPS = buildComponentProps(componentDescriptors);
  let KNOWN_COMPONENT_TYPES = new Set(Object.keys(COMP_PROPS));
  let COMPONENT_META = componentMetaMap();
  const CANVAS_CHILD_TYPES = new Set(['Ball', 'ImageSprite']);
  const MAP_CHILD_TYPES = new Set(['Circle', 'FeatureCollection', 'LineString', 'Marker', 'Polygon', 'Rectangle']);
  const FEATURE_COLLECTION_CHILD_TYPES = new Set(['Circle', 'LineString', 'Marker', 'Polygon', 'Rectangle']);
  const CHART_CHILD_TYPES = new Set(['ChartData2D', 'Trendline']);

  function componentTypeName(ident) {
    const dotIdx = ident.indexOf('.');
    return dotIdx === -1 ? ident : ident.slice(0, dotIdx);
  }

  function isKnownComponentIdent(ident) {
    return KNOWN_COMPONENT_TYPES.has(componentTypeName(ident));
  }

  function dbComponentType(type) {
    return COMPONENT_ALIASES[type] || type;
  }

  const FALLBACK_PROPS = [
    { name: 'Visible', editorType: 'visibility',  category: 'Appearance' },
    { name: 'Height',  editorType: 'layout_size', category: 'Appearance' },
    { name: 'Width',   editorType: 'layout_size', category: 'Appearance' },
  ];


  let CONTAINER_TYPES = new Set();
  $: componentDescriptors = allComponentDescriptors($extensionComponentDescriptors);
  $: COMP_PROPS = buildComponentProps(componentDescriptors);
  $: KNOWN_COMPONENT_TYPES = new Set(Object.keys(COMP_PROPS));
  $: COMPONENT_META = componentMetaMap($extensionComponentDescriptors);
  $: CONTAINER_TYPES = new Set([
    'Screen',
    'Form',
    'ScrollHorizontal',
    'ScrollVertical',
    ...componentDescriptors
      .filter(component => component.categoryString === 'LAYOUT')
      .map(component => component.name),
  ]);

  function isNonVisibleComponentType(type) {
    const dbType = dbComponentType(type);
    if (dbType === 'Form') return false;
    return COMPONENT_META.get(dbType)?.nonVisible === 'true';
  }

  function canContainComponent(parentType, childType) {
    const parent = dbComponentType(parentType);
    const child = dbComponentType(childType);
    if (!parent || !child) return false;
    if (CANVAS_CHILD_TYPES.has(child)) return parent === 'Canvas';
    if (MAP_CHILD_TYPES.has(child)) {
      if (parent === 'Map') return true;
      return parent === 'FeatureCollection' && FEATURE_COLLECTION_CHILD_TYPES.has(child);
    }
    if (CHART_CHILD_TYPES.has(child)) return parent === 'Chart';
    if (isNonVisibleComponentType(child)) {
      return parent === 'Form';
    }
    if (isNonVisibleComponentType(parent)) return false;
    if (parent === 'Canvas' || parent === 'Map' || parent === 'FeatureCollection' || parent === 'Chart') return false;
    return CONTAINER_TYPES.has(parent);
  }

  // ── Schema parser ──────────────────────────────────────────────────
  function stripLineComments(text) {
    let out = '';
    let inString = false;
    let escaped = false;

    for (let i = 0; i < text.length; i += 1) {
      const ch = text[i];
      const next = text[i + 1];

      if (inString) {
        out += ch;
        if (escaped) {
          escaped = false;
        } else if (ch === '\\') {
          escaped = true;
        } else if (ch === '"') {
          inString = false;
        }
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

  function parseSchema(text) {
    if (!text?.trim()) return { root: null, error: 'The design schema is empty.' };
    text = stripLineComments(text);
    let pos = 0;
    const typeCounts = {};

    function skipWs() { while (pos < text.length && /\s/.test(text[pos])) pos++; }
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
      let s = '';
      while (pos < text.length && /[\w.]/.test(text[pos])) s += text[pos++];
      return s;
    }
    function readStr() {
      pos++;
      let s = '';
      while (pos < text.length && text[pos] !== '"') {
        if (text[pos] === '\\') {
          pos++;
          if (pos >= text.length) fail('Unterminated string literal');
        }
        s += text[pos++];
      }
      if (pos >= text.length) fail('Unterminated string literal');
      pos++;
      return s;
    }
    function readValue() {
      skipWs();
      if (text[pos] === '"') return readStr();
      let s = '';
      while (pos < text.length && !/[,\n{}]/.test(text[pos])) s += text[pos++];
      const value = s.trim();
      if (!value) fail('Expected a property value');
      return value;
    }
    function makeComponent(ident, pathId) {
      const dotIdx = ident.indexOf('.');
      const type = componentTypeName(ident);
      let name = dotIdx === -1 ? null : ident.slice(dotIdx + 1);
      if (!type) fail('Expected a component type');
      if (!isKnownComponentIdent(type)) fail(`Unknown component type "${type}"`);
      if (!name) {
        typeCounts[type] = (typeCounts[type] || 0) + 1;
        name = `${type}${typeCounts[type]}`;
      }
      const isRootType = type === 'Screen' || type === 'Form';
      const validName = isRootType ? isValidScreenName(name) : isValidComponentName(name);
      if (!validName) fail(appInventorNameError(name, { component: !isRootType }));
      return { type, name, props: {}, children: [], pathId };
    }

    function parseComp(pathId) {
      skipWs();
      if (pos >= text.length) fail('Expected a component');
      const ident = readIdent();
      if (!ident) fail('Expected a component or property name');
      const comp = makeComponent(ident, pathId);
      skipWs();
      if (text[pos] !== '{') return comp;
      pos++;
      let ci = 0;
      while (true) {
        skipWs();
        if (text[pos] === '}') { pos++; break; }
        if (pos >= text.length) fail(`Expected "}" to close ${comp.type}.${comp.name}`);
        const save = pos;
        const tok = readIdent();
        if (!tok) fail('Expected a property or child component');
        skipWs();
        if (pos < text.length && text[pos] === ':') {
          pos++;
          skipWs();
          comp.props[tok] = readValue();
          skipWs();
          if (pos < text.length && text[pos] === ',') pos++;
        } else if (pos < text.length && text[pos] === '{') {
          if (!isKnownComponentIdent(tok)) fail(`Expected ":" after property "${tok}"`);
          pos = save;
          const child = parseComp(`${pathId}-${ci++}`);
          if (child) comp.children.push(child);
          skipWs();
          if (pos < text.length && text[pos] === ',') pos++;
        } else {
          if (!isKnownComponentIdent(tok)) fail(`Expected ":" after property "${tok}"`);
          pos = save;
          const child = parseComp(`${pathId}-${ci++}`);
          if (child) comp.children.push(child);
          skipWs();
          if (pos < text.length && text[pos] === ',') pos++;
        }
      }
      return comp;
    }

    try {
      const root = parseComp('0');
      skipWs();
      if (pos < text.length) fail(`Unexpected token "${text[pos]}"`);
      if (root.type !== 'Screen' && root.type !== 'Form') fail('The root designer component must be Screen');
      validateTreePlacement(root);
      validateUniqueComponentNames(root);
      return { root, error: null };
    } catch (err) {
      return { root: null, error: err.message || 'Unable to parse design schema.' };
    }
  }

  function validateTreePlacement(node, parent = null) {
    if (parent && !canContainComponent(parent.type, node.type)) {
      throw new Error(`${node.type}.${node.name} cannot be placed inside ${parent.type}.${parent.name}`);
    }
    for (const child of node.children || []) validateTreePlacement(child, node);
  }

  function validateUniqueComponentNames(node, seen = new Set()) {
    if (seen.has(node.name)) throw new Error(`Duplicate component name "${node.name}"`);
    seen.add(node.name);
    for (const child of node.children || []) validateUniqueComponentNames(child, seen);
  }

  // ── Serializer ─────────────────────────────────────────────────────
  function isColorProp(node, propName) {
    const prop = (COMP_PROPS[node?.type] ?? FALLBACK_PROPS).find(p => p.name === propName);
    return prop?.editorType === 'color' || /(Color|Colour)$/.test(propName) || propName === 'PaintColor';
  }

  function needsQuotes(val, node, propName) {
    const text = String(val).trim();
    if (/^(true|false|True|False|-?\d+\.?\d*|&H[0-9A-Fa-f]{8})$/.test(text)) return false;
    if (isColorProp(node, propName) && /^#[0-9A-Fa-f]{3,8}$/.test(text)) return false;
    return true;
  }

  function quoteSchemaString(val) {
    return `"${String(val).replace(/\\/g, '\\\\').replace(/"/g, '\\"')}"`;
  }

  function serializeComp(node, depth = 0) {
    const ind = '  '.repeat(depth);
    const ident = `${node.type}.${node.name}`;
    const propLines = Object.entries(node.props || {}).map(
      ([k, v]) => `${ind}  ${k}: ${needsQuotes(v, node, k) ? quoteSchemaString(v) : v}`
    );
    const childLines = (node.children || []).map(c => serializeComp(c, depth + 1));
    const body = [...propLines, ...childLines];
    if (body.length === 0) return `${ind}${ident}`;
    return `${ind}${ident} {\n${body.join(',\n')}\n${ind}}`;
  }

  function serializeTree(root) {
    return root ? serializeComp(root, 0) : '';
  }

  // ── Tree utilities ─────────────────────────────────────────────────
  function flattenTree(node, depth, collapsed, parentPathId = null, sibIdx = 0, sibCount = 1) {
    if (!node) return [];
    const rows = [{ ...node, depth, parentPathId, sibIdx, sibCount }];
    if (!collapsed.has(node.pathId)) {
      node.children.forEach((c, i) =>
        rows.push(...flattenTree(c, depth + 1, collapsed, node.pathId, i, node.children.length))
      );
    }
    return rows;
  }

  function groupByCategory(props) {
    const map = {};
    for (const p of props) { const cat = p.category || 'Appearance'; (map[cat] ??= []).push(p); }
    return Object.entries(map).map(([cat, items]) => ({ cat, items }));
  }

  function cloneTree(node) { return JSON.parse(JSON.stringify(node)); }

  function findNode(root, pathId) {
    if (!root) return null;
    if (root.pathId === pathId) return root;
    for (const c of root.children) { const f = findNode(c, pathId); if (f) return f; }
    return null;
  }

  function findParentOf(root, pathId) {
    for (let i = 0; i < root.children.length; i++) {
      if (root.children[i].pathId === pathId) return { parent: root, index: i };
      const f = findParentOf(root.children[i], pathId);
      if (f) return f;
    }
    return null;
  }

  function rebuildPathIds(node, pid = '0') {
    node.pathId = pid;
    node.children.forEach((c, i) => rebuildPathIds(c, `${pid}-${i}`));
    return node;
  }

  function collectNames(root) {
    if (!root) return [];
    return [root.name, ...root.children.flatMap(collectNames)];
  }

  function nameExists(name, root, exceptPathId = null) {
    if (!root) return false;
    if (root.pathId !== exceptPathId && root.name === name) return true;
    return root.children.some(child => nameExists(name, child, exceptPathId));
  }

  function uniqueName(type, root) {
    const existing = new Set(collectNames(root));
    let n = 1;
    while (existing.has(`${type}${n}`)) n++;
    return `${type}${n}`;
  }

  // ── Mutable tree state ─────────────────────────────────────────────
  let prevSchema = '';
  let mutableTree = null;
  let parseError = '';

  $: {
    if (schemaValue !== prevSchema) {
      prevSchema = schemaValue;
      const parsed = parseSharedDesignSchemaResult(schemaValue);
      parseError = parsed.error || '';
      mutableTree = parsed.root;
      if (parsed.root && !findNode(parsed.root, selectedPathId)) selectedPathId = '0';
    }
  }

  function applyChange(newTree) {
    mutableTree = newTree;
    const schema = serializeSharedDesignTree(newTree);
    prevSchema = schema;
    parseError = '';
    dispatch('change', { schema });
  }

  // ── Display state ──────────────────────────────────────────────────
  let collapsed = new Set();
  let selectedPathId = '0';
  let collapsedCategories = new Set();
  let treeTab = 'components';
  let showPicker = false;
  let addCompValue = '';
  let addCompError = '';
  let addCompInputEl;
  let addCompHighlight = 0;
  let addCompSelecting = false; // true while clicking a suggestion, suppresses blur

  $: ALL_COMPONENT_NAMES = Array.from(KNOWN_COMPONENT_TYPES)
    .filter(name => name !== 'Screen' && name !== 'Form')
    .sort((a, b) => a.localeCompare(b));

  $: addCompSuggestions = addCompValue.trim()
    ? ALL_COMPONENT_NAMES.filter(n => n.toLowerCase().startsWith(addCompValue.trim().toLowerCase())).slice(0, 7)
    : [];
  $: { addCompSuggestions; addCompHighlight = 0; }
  let ctxMenu = null; // { x, y, pathId, isContainer, isRoot }
  let assetCtxMenu = null; // { x, y, assetId }
  let helpPopup = null; // { x, y, description }
  let renamingAssetId = null;
  let assetRenameValue = '';
  let assetRenameError = '';
  let assetRenameInputEl;
  let renamingPathId = null;
  let renameValue = '';
  let renameError = '';
  let renameInputEl;

  let dragPathId = null;
  let dropTarget = null; // { pathId, position: 'before' | 'after' | 'into' }

  $: flatList = mutableTree ? flattenTree(mutableTree, 0, collapsed) : [];
  $: mediaAssets = $designAssets.map(normalizeAssetRecord);
  $: activeAsset = assetCtxMenu ? mediaAssets.find(asset => asset.id === assetCtxMenu.assetId) : null;
  $: if (mutableTree && flatList.length && !flatList.some(n => n.pathId === selectedPathId)) {
    selectedPathId = flatList[0].pathId;
  }
  $: selectedNode = flatList.find(n => n.pathId === selectedPathId) ?? flatList[0] ?? null;
  $: propGroups = selectedNode
    ? groupByCategory(COMP_PROPS[selectedNode.type] ?? FALLBACK_PROPS)
    : [];
  $: isRoot = selectedNode?.pathId === '0';

  // ── Universal search: publish a full (collapse-independent) component/property index ──
  let flashPropName = null;
  let treeFlashPathId = null;
  let handledDesignerNavToken = null;
  let propFlashTimer = null;
  let treeFlashTimer = null;

  $: designerSearchIndex.set(
    (mutableTree ? flattenTree(mutableTree, 0, new Set()) : []).map(node => ({
      pathId: node.pathId,
      name: node.name,
      type: node.type,
      props: (COMP_PROPS[node.type] ?? FALLBACK_PROPS).map(p => ({
        name: p.name,
        category: p.category || 'Appearance',
        value: node.props?.[p.name] ?? '',
      })),
    }))
  );

  $: if (
    $searchNavigation
    && $searchNavigation.token !== handledDesignerNavToken
    && $searchNavigation.scope === 'designer-tree'
  ) {
    handledDesignerNavToken = $searchNavigation.token;
    revealTreeMatch($searchNavigation.pathId, $searchNavigation.propName, $searchNavigation.category);
  }

  function expandAncestors(pathId) {
    const parts = String(pathId || '').split('-');
    if (parts.length <= 1) return;
    const ancestors = [];
    for (let i = 1; i < parts.length; i += 1) ancestors.push(parts.slice(0, i).join('-'));
    if (!ancestors.some(a => collapsed.has(a))) return;
    const next = new Set(collapsed);
    ancestors.forEach(a => next.delete(a));
    collapsed = next;
  }

  async function revealTreeMatch(pathId, propName, category) {
    if (!mutableTree || !findNode(mutableTree, pathId)) return;
    expandAncestors(pathId);
    selectedPathId = pathId;
    treeTab = 'components';
    if (propName && category && collapsedCategories.has(category)) {
      const next = new Set(collapsedCategories);
      next.delete(category);
      collapsedCategories = next;
    }
    await tick();

    const treeRow = document.querySelector('.vis-tree-scroll .vis-item.selected');
    treeRow?.scrollIntoView({ block: 'nearest' });

    if (propName) {
      await tick();
      const row = document.querySelector(`.vis-props-scroll [data-prop-name="${cssAttrEscape(propName)}"]`);
      row?.scrollIntoView({ block: 'center' });
      flashPropName = propName;
      clearTimeout(propFlashTimer);
      propFlashTimer = setTimeout(() => { flashPropName = null; }, 1400);
    } else {
      treeFlashPathId = pathId;
      clearTimeout(treeFlashTimer);
      treeFlashTimer = setTimeout(() => { treeFlashPathId = null; }, 1400);
    }
  }

  function cssAttrEscape(value) {
    return String(value).replace(/["\\]/g, '\\$&');
  }

  onMount(() => { designerTreeActive.set(true); });

  onDestroy(() => {
    designerTreeActive.set(false);
    designerSearchIndex.set([]);
    clearTimeout(propFlashTimer);
    clearTimeout(treeFlashTimer);
  });

  function normalizeAssetRecord(asset, index) {
    if (typeof asset === 'string') {
      return { id: asset || `asset-${index}`, name: asset, size: 0, type: '', url: '' };
    }
    return {
      id: asset?.id || asset?.name || `asset-${index}`,
      name: asset?.name || '',
      size: asset?.size || 0,
      type: asset?.type || '',
      url: asset?.url || '',
    };
  }

  function toggleCollapse(pathId) {
    const next = new Set(collapsed);
    next.has(pathId) ? next.delete(pathId) : next.add(pathId);
    collapsed = next;
  }

  function toggleCategory(cat) {
    const next = new Set(collapsedCategories);
    next.has(cat) ? next.delete(cat) : next.add(cat);
    collapsedCategories = next;
  }

  function openHelp(e, description) {
    e.stopPropagation();
    if (helpPopup) { helpPopup = null; return; }
    const r = e.currentTarget.getBoundingClientRect();
    const calloutW = 220;
    const btnCenterX = r.left + r.width / 2;
    const left = Math.max(8, Math.min(btnCenterX - calloutW / 2, window.innerWidth - calloutW - 8));
    const arrowLeft = btnCenterX - left;
    const bottomPx = window.innerHeight - r.top + 6;
    helpPopup = { left, arrowLeft, bottomPx, description: description || 'No description available.' };
  }

  function closeHelp() { helpPopup = null; }

  // ── Tree operations ────────────────────────────────────────────────
  function addComponent(type) {
    if (!mutableTree) return;
    const newTree = cloneTree(mutableTree);
    const newName = uniqueName(type, newTree);
    const newComp = { type, name: newName, props: {}, children: [], pathId: '' };

    const sel = findNode(newTree, selectedPathId);
    if (isNonVisibleComponentType(type)) {
      newTree.children.push(newComp);
    } else if (sel && canContainComponent(sel.type, type)) {
      sel.children.push(newComp);
    } else if (sel && selectedPathId !== '0') {
      const p = findParentOf(newTree, selectedPathId);
      if (p && canContainComponent(p.parent.type, type)) p.parent.children.splice(p.index + 1, 0, newComp);
      else newTree.children.push(newComp);
    } else {
      newTree.children.push(newComp);
    }

    rebuildPathIds(newTree);
    applyChange(newTree);

    const added = flattenTree(newTree, 0, new Set()).find(n => n.name === newName);
    if (added) {
      // Expand any collapsed ancestor so the new node is visible
      if (added.parentPathId) {
        const next = new Set(collapsed);
        next.delete(added.parentPathId);
        collapsed = next;
      }
      selectedPathId = added.pathId;
    }
    showPicker = false;
  }

  // ── Drag and drop ─────────────────────────────────────────────────
  function isDescendantOrSelf(node, targetPathId) {
    if (node.pathId === targetPathId) return true;
    return node.children.some(c => isDescendantOrSelf(c, targetPathId));
  }

  function handleDragStart(e, pathId) {
    dragPathId = pathId;
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', pathId);
  }

  function handleDragOver(e, pathId) {
    e.preventDefault();
    if (!dragPathId || !mutableTree) return;
    if (pathId === dragPathId) { dropTarget = null; return; }
    const dragged = findNode(mutableTree, dragPathId);
    if (dragged && isDescendantOrSelf(dragged, pathId)) { dropTarget = null; return; }

    const rect = e.currentTarget.getBoundingClientRect();
    const relY = (e.clientY - rect.top) / rect.height;
    const targetNode = flatList.find(n => n.pathId === pathId);
    const targetParent = pathId !== '0' ? flatList.find(n => n.pathId === targetNode?.parentPathId) : null;
    const targetAcceptsDragged = canContainComponent(targetNode?.type, dragged?.type);
    const targetParentAcceptsDragged = targetParent
      ? canContainComponent(targetParent.type, dragged?.type)
      : pathId !== '0' && canContainComponent(mutableTree.type, dragged?.type);

    let position;
    if (pathId === '0') {
      position = 'into';
    } else if (targetAcceptsDragged && relY > 0.25 && relY < 0.75) {
      position = 'into';
    } else if (targetParentAcceptsDragged) {
      position = relY < 0.5 ? 'before' : 'after';
    } else {
      dropTarget = null;
      e.dataTransfer.dropEffect = 'none';
      return;
    }

    if (position === 'into' && !targetAcceptsDragged) {
      dropTarget = null;
      e.dataTransfer.dropEffect = 'none';
      return;
    }
    dropTarget = { pathId, position };
    e.dataTransfer.dropEffect = 'move';
  }

  function handleScrollDragLeave(e) {
    if (!e.relatedTarget || !e.currentTarget.contains(e.relatedTarget)) {
      dropTarget = null;
    }
  }

  function handleDrop(e, pathId) {
    e.preventDefault();
    if (!dragPathId || !dropTarget || !mutableTree) { handleDragEnd(); return; }
    const { pathId: targetPathId, position } = dropTarget;
    if (targetPathId === dragPathId) { handleDragEnd(); return; }

    const newTree = cloneTree(mutableTree);
    const draggedNode = findNode(newTree, dragPathId);
    const targetNode  = findNode(newTree, targetPathId);
    if (!draggedNode || !targetNode) { handleDragEnd(); return; }

    const targetParentResult = targetPathId !== '0' ? findParentOf(newTree, targetPathId) : null;
    const targetParentNode   = targetParentResult?.parent ?? null;

    const draggedParentResult = findParentOf(newTree, dragPathId);
    if (!draggedParentResult) { handleDragEnd(); return; }
    draggedParentResult.parent.children.splice(draggedParentResult.index, 1);

    if (position === 'into') {
      if (!canContainComponent(targetNode.type, draggedNode.type)) { handleDragEnd(); return; }
      targetNode.children.push(draggedNode);
    } else if (targetParentNode) {
      if (!canContainComponent(targetParentNode.type, draggedNode.type)) { handleDragEnd(); return; }
      const idx = targetParentNode.children.indexOf(targetNode);
      targetParentNode.children.splice(idx + (position === 'after' ? 1 : 0), 0, draggedNode);
    } else {
      if (!canContainComponent(newTree.type, draggedNode.type)) { handleDragEnd(); return; }
      newTree.children.push(draggedNode);
    }

    const movedName = draggedNode.name;
    rebuildPathIds(newTree);

    if (position === 'into') {
      const newFlat = flattenTree(newTree, 0, new Set());
      const newTarget = newFlat.find(n => n.name === targetNode.name);
      if (newTarget) {
        const next = new Set(collapsed);
        next.delete(newTarget.pathId);
        collapsed = next;
      }
    }

    applyChange(newTree);
    const moved = flattenTree(newTree, 0, collapsed).find(n => n.name === movedName);
    if (moved) selectedPathId = moved.pathId;
    handleDragEnd();
  }

  function handleDragEnd() {
    dragPathId = null;
    dropTarget = null;
  }

  // ── Context menu ───────────────────────────────────────────────────
  function handleItemCtxMenu(e, node) {
    e.preventDefault();
    e.stopPropagation();
    selectedPathId = node.pathId;
    showPicker = false;
    assetCtxMenu = null;
    ctxMenu = {
      x: Math.min(e.clientX, window.innerWidth - 172),
      y: Math.min(e.clientY, window.innerHeight - 150),
      pathId: node.pathId,
      isContainer: CONTAINER_TYPES.has(node.type),
      isRoot: node.pathId === '0',
    };
  }

  function closeCtxMenu() {
    ctxMenu = null;
    assetCtxMenu = null;
  }

  // ── Property dropdown state ────────────────────────────────────────
  let visDropdownKey = null;
  let visDropdownX = 0;
  let visDropdownY = 0;
  let visDropdownMinWidth = 80;
  let visDropdownOptions = [];
  let visDropdownCurrent = null;
  let visDropdownOnSelect = null;

  function openVisDropdown(e, key, options, current, onSelect) {
    e.stopPropagation();
    if (visDropdownKey === key) { visDropdownKey = null; return; }
    closeCtxMenu();
    const rect = e.currentTarget.getBoundingClientRect();
    visDropdownX = rect.left;
    visDropdownY = rect.bottom + 4;
    visDropdownMinWidth = rect.width;
    visDropdownOptions = options;
    visDropdownCurrent = String(current ?? '');
    visDropdownOnSelect = onSelect;
    visDropdownKey = key;
  }

  function closeVisDropdown() { visDropdownKey = null; }

  function ctxRename()  { closeCtxMenu(); startRename(); }
  function ctxDelete()  { closeCtxMenu(); deleteSelected(); }
  function ctxAddComp() {
    closeCtxMenu();
    addCompValue = '';
    addCompError = '';
    showPicker = true;
    tick().then(() => addCompInputEl?.focus());
  }

  function pickSuggestion(name) {
    addCompSelecting = false;
    addComponent(name);
    showPicker = false;
    addCompValue = '';
    addCompError = '';
  }

  function commitAddComp() {
    const resolved = addCompSuggestions[addCompHighlight] ?? addCompValue.trim();
    if (!resolved) {
      addCompError = 'Enter a component name.';
      tick().then(() => addCompInputEl?.focus());
      return;
    }
    if (!KNOWN_COMPONENT_TYPES.has(resolved)) {
      addCompError = 'Unknown component type.';
      tick().then(() => addCompInputEl?.focus());
      return;
    }
    pickSuggestion(resolved);
  }

  function handleAddCompKey(e) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      addCompHighlight = Math.min(addCompHighlight + 1, addCompSuggestions.length - 1);
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      addCompHighlight = Math.max(addCompHighlight - 1, 0);
    } else if (e.key === 'Tab' && addCompSuggestions.length) {
      e.preventDefault();
      pickSuggestion(addCompSuggestions[addCompHighlight]);
    } else if (e.key === 'Enter') {
      commitAddComp();
    } else if (e.key === 'Escape') {
      showPicker = false;
      addCompValue = '';
      addCompError = '';
    }
  }

  function deleteSelected() {
    if (!selectedPathId || selectedPathId === '0' || !mutableTree) return;
    const newTree = cloneTree(mutableTree);
    const p = findParentOf(newTree, selectedPathId);
    if (!p) return;
    const parentPid = p.parent.pathId;
    p.parent.children.splice(p.index, 1);
    rebuildPathIds(newTree);
    applyChange(newTree);
    selectedPathId = parentPid;
  }

  function startRename() {
    if (!selectedNode || isRoot) return;
    renamingPathId = selectedNode.pathId;
    renameValue = selectedNode.name;
    renameError = '';
    tick().then(() => { renameInputEl?.focus(); renameInputEl?.select(); });
  }

  function commitRename() {
    if (!renamingPathId || !mutableTree) { renamingPathId = null; renameError = ''; return; }
    const targetPathId = renamingPathId;
    const value = renameValue.trim();
    const current = findNode(mutableTree, renamingPathId)?.name;
    if (!value) {
      renameError = 'Name is required.';
      tick().then(() => renameInputEl?.focus());
      return;
    }
    const nameError = appInventorNameError(value, { component: true });
    if (nameError || !isValidComponentName(value)) {
      renameError = nameError || 'Invalid App Inventor component name.';
      tick().then(() => renameInputEl?.focus());
      return;
    }
    if (value !== current && nameExists(value, mutableTree, targetPathId)) {
      renameError = 'Name already exists.';
      tick().then(() => renameInputEl?.focus());
      return;
    }
    renamingPathId = null;
    renameError = '';
    if (value === current) return;
    const newTree = cloneTree(mutableTree);
    const target = findNode(newTree, targetPathId);
    if (target) target.name = value;
    applyChange(newTree);
  }

  function handleRenameKey(e) {
    if (e.key === 'Enter') commitRename();
    if (e.key === 'Escape') { renamingPathId = null; renameError = ''; }
  }

  // ── Property value helpers ─────────────────────────────────────────
  function propVal(node, name) { return node?.props?.[name] ?? null; }

  function propDisplayVal(prop, val) {
    return val ?? prop.defaultValue ?? '';
  }

  function updateProp(pathId, propName, value) {
    if (!mutableTree || !pathId) return;
    const newTree = cloneTree(mutableTree);
    const target = findNode(newTree, pathId);
    if (!target) return;
    target.props = target.props || {};
    target.props[propName] = String(value);
    if (target.type === 'ListView' && propName === 'ListViewLayout' && target.props.ListData) {
      target.props.ListData = pruneListViewDataForLayout(target.props.ListData, value);
    }
    applyChange(newTree);
    selectedPathId = pathId;
  }

  function aiColorToHex(val) {
    const m = String(val || '').match(/^&H([0-9A-Fa-f]{8})$/);
    if (!m || parseInt(m[1].slice(0, 2), 16) === 0) return '#1a1916';
    const h = m[1];
    const r = h.slice(2, 4);
    const g = h.slice(4, 6);
    const b = h.slice(6, 8);
    return `#${r}${g}${b}`.toLowerCase();
  }

  function hexToAiColor(hex) {
    const m = String(hex || '').match(/^#?([0-9A-Fa-f]{6})$/);
    if (!m) return '&H00000000';
    const h = m[1].toUpperCase();
    const r = h.slice(0, 2);
    const g = h.slice(2, 4);
    const b = h.slice(4, 6);
    return `&HFF${r}${g}${b}`;
  }

  function colorDisplay(val) {
    if (!val || val === '&H00000000') return { hex: '#1a1916', label: 'Default' };
    const m = String(val).match(/^&H([0-9A-Fa-f]{8})$/);
    if (m) {
      const h = m[1];
      if (parseInt(h.slice(0, 2), 16) === 0) return { hex: '#1a1916', label: 'Default' };
      return { hex: aiColorToHex(val), label: val };
    }
    return { hex: '#1a1916', label: val || 'Default' };
  }

  function boolVal(val, def = false) {
    if (val === null || val === undefined) return def;
    return val === 'true' || val === 'True';
  }

  function boolDefault(prop) {
    return prop?.defaultValue === true
      || prop?.defaultValue === 'true'
      || prop?.defaultValue === 'True';
  }

  function layoutMode(val) {
    if (!val || val === '-2') return 'automatic';
    if (val === '-1') return 'fill';
    return 'custom';
  }

  function layoutPixels(val) {
    const n = Number(val);
    return Number.isFinite(n) && n > 0 ? String(Math.round(n)) : '100';
  }

  function updateLayoutMode(pathId, propName, mode, currentVal) {
    if (mode === 'automatic') updateProp(pathId, propName, '-2');
    else if (mode === 'fill') updateProp(pathId, propName, '-1');
    else updateProp(pathId, propName, layoutPixels(currentVal));
  }

  function updateLayoutPixels(pathId, propName, value) {
    const n = Math.max(1, Math.round(Number(value) || 1));
    updateProp(pathId, propName, String(n));
  }

  function selectOptions(prop) {
    return prop.options || [];
  }

  function selectValue(prop, val) {
    if (val !== null && val !== undefined && val !== '') return String(val);
    if (prop.editorType === 'visibility') return 'true';
    if (prop.defaultValue !== undefined && prop.defaultValue !== '') return String(prop.defaultValue);
    return String(selectOptions(prop)[0]?.value ?? '');
  }

  function assetOptions(currentValue) {
    const names = new Set(mediaAssets.map(asset => asset.name).filter(Boolean));
    if (currentValue) names.add(currentValue);
    return [...names].sort((a, b) => a.localeCompare(b));
  }

  function handleAssetPicked(e, pathId, propName) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;
    const added = addDesignAsset(file);
    updateProp(pathId, propName, added?.name || cleanAssetName(file.name));
    e.currentTarget.value = '';
  }

  function cleanAssetName(name) {
    return normalizeAppInventorAssetName(name);
  }

  function assetNameExists(name, exceptId = null) {
    return mediaAssets.some(asset => asset.id !== exceptId && asset.name === name);
  }

  function handleMediaUpload(e) {
    const files = Array.from(e.currentTarget.files || []);
    for (const file of files) addDesignAsset(file);
    e.currentTarget.value = '';
  }

  function handleAssetCtxMenu(e, asset) {
    e.preventDefault();
    e.stopPropagation();
    ctxMenu = null;
    assetCtxMenu = {
      x: Math.min(e.clientX, window.innerWidth - 172),
      y: Math.min(e.clientY, window.innerHeight - 150),
      assetId: asset.id,
    };
  }

  function startAssetRename(asset) {
    assetCtxMenu = null;
    renamingAssetId = asset.id;
    assetRenameValue = asset.name;
    assetRenameError = '';
    tick().then(() => { assetRenameInputEl?.focus(); assetRenameInputEl?.select(); });
  }

  function updateAssetReferences(oldName, nextName) {
    if (!mutableTree || !oldName) return;
    const newTree = cloneTree(mutableTree);
    let changed = false;

    function walk(node) {
      for (const propName of Object.keys(node.props || {})) {
        if (node.props[propName] === oldName) {
          node.props[propName] = nextName;
          changed = true;
        } else if (node.type === 'ListView' && propName === 'ListData') {
          const nextListData = renameListViewDataAsset(node.props[propName], oldName, nextName);
          if (nextListData !== node.props[propName]) {
            node.props[propName] = nextListData;
            changed = true;
          }
        }
      }
      for (const child of node.children || []) walk(child);
    }

    walk(newTree);
    if (changed) applyChange(newTree);
  }

  function commitAssetRename() {
    if (!renamingAssetId) return;
    const current = mediaAssets.find(asset => asset.id === renamingAssetId);
    if (!current) {
      renamingAssetId = null;
      assetRenameError = '';
      return;
    }

    const clean = cleanAssetName(assetRenameValue);
    if (!clean) {
      assetRenameError = 'Name is required.';
      tick().then(() => assetRenameInputEl?.focus());
      return;
    }
    const assetNameError = appInventorAssetNameError(clean);
    if (assetNameError) {
      assetRenameError = assetNameError;
      tick().then(() => assetRenameInputEl?.focus());
      return;
    }
    if (clean !== current.name && assetNameExists(clean, current.id)) {
      assetRenameError = 'Asset already exists.';
      tick().then(() => assetRenameInputEl?.focus());
      return;
    }

    renamingAssetId = null;
    assetRenameError = '';
    if (clean === current.name) return;
    const renamed = renameDesignAsset(current.id, clean);
    if (renamed?.name) updateAssetReferences(current.name, renamed.name);
  }

  function handleAssetRenameKey(e) {
    if (e.key === 'Enter') commitAssetRename();
    if (e.key === 'Escape') {
      renamingAssetId = null;
      assetRenameError = '';
    }
  }

  function removeAsset(asset) {
    const removed = deleteDesignAsset(asset.id);
    if (removed?.name) updateAssetReferences(removed.name, '');
    if (renamingAssetId === asset.id) {
      renamingAssetId = null;
      assetRenameError = '';
    }
    closeCtxMenu();
  }

  function downloadAsset(asset) {
    if (!asset?.url || typeof document === 'undefined') return;
    const link = document.createElement('a');
    link.href = asset.url;
    link.download = asset.name;
    document.body.appendChild(link);
    link.click();
    link.remove();
    closeCtxMenu();
  }

  function formatAssetSize(size) {
    if (!size) return '';
    if (size < 1024) return `${size} B`;
    if (size < 1024 * 1024) return `${Math.round(size / 102.4) / 10} KB`;
    return `${Math.round(size / 104857.6) / 10} MB`;
  }

  // ── ListData dialog ────────────────────────────────────────────────
  let listDataDialogOpen = false;
  let listDataDialogPathId = null;
  let listDataDialogRows = [];
  let listDataDialogColumns = [];
  let listDataDialogLayoutValue = '0';

  function openListDataDialog(pathId) {
    const node = findNode(mutableTree, pathId);
    if (!node) return;
    const layoutValue = String(node.props?.ListViewLayout ?? '0');
    const currentData = node.props?.ListData ?? '';
    listDataDialogPathId = pathId;
    listDataDialogLayoutValue = layoutValue;
    listDataDialogColumns = listViewColumnsForLayout(layoutValue);
    listDataDialogRows = parseListViewData(currentData, layoutValue).map(r => ({ ...r }));
    listDataDialogOpen = true;
  }

  function addListDataRow() {
    listDataDialogRows = [...listDataDialogRows, emptyListViewRow(listDataDialogLayoutValue)];
  }

  function removeListDataRow(i) {
    listDataDialogRows = listDataDialogRows.filter((_, idx) => idx !== i);
  }

  function moveListDataRow(i, dir) {
    const j = i + dir;
    if (j < 0 || j >= listDataDialogRows.length) return;
    const rows = [...listDataDialogRows];
    [rows[i], rows[j]] = [rows[j], rows[i]];
    listDataDialogRows = rows;
  }

  function updateListDataCell(i, col, value) {
    listDataDialogRows = listDataDialogRows.map((r, idx) =>
      idx === i ? { ...r, [col]: value } : r
    );
  }

  function saveListDataDialog() {
    const serialized = serializeListViewData(listDataDialogRows, listDataDialogLayoutValue);
    updateProp(listDataDialogPathId, 'ListData', serialized);
    listDataDialogOpen = false;
  }

  function cancelListDataDialog() {
    listDataDialogOpen = false;
  }

  function handleListDataDialogKey(e) {
    if (!listDataDialogOpen || e.key !== 'Escape') return;
    e.stopPropagation();
    cancelListDataDialog();
  }

  // ── Component icons (12×12 viewBox inner markup) ───────────────────
  function compIcon(type) {
    switch (type) {
      case 'Screen':              return '<rect x="2.5" y="0.5" width="7" height="11" rx="1.5" stroke-width="1.3"/><line x1="4" y1="9.5" x2="8" y2="9.5" stroke-width="1.3" stroke-linecap="round"/>';
      case 'Button':              return '<rect x="0.5" y="3.5" width="11" height="5" rx="2" stroke-width="1.3"/><line x1="3.5" y1="6" x2="8.5" y2="6" stroke-width="1.5" stroke-linecap="round"/>';
      case 'Label':               return '<path d="M2 2.5h8M6 2.5v7" stroke-width="1.5" stroke-linecap="round"/>';
      case 'TextBox':             return '<rect x="0.5" y="3" width="11" height="6" rx="1" stroke-width="1.3"/><line x1="2.5" y1="6" x2="5.5" y2="6" stroke-width="1.3" stroke-linecap="round"/>';
      case 'HorizontalArrangement': return '<rect x="0.5" y="1.5" width="11" height="9" rx="1" stroke-width="1.3"/><line x1="4.5" y1="1.5" x2="4.5" y2="10.5" stroke-width="1.1"/><line x1="8" y1="1.5" x2="8" y2="10.5" stroke-width="1.1"/>';
      case 'VerticalArrangement': return '<rect x="1.5" y="0.5" width="9" height="11" rx="1" stroke-width="1.3"/><line x1="1.5" y1="4.5" x2="10.5" y2="4.5" stroke-width="1.1"/><line x1="1.5" y1="8" x2="10.5" y2="8" stroke-width="1.1"/>';
      case 'ListView':            return '<rect x="0.5" y="0.5" width="11" height="11" rx="1" stroke-width="1.3"/><line x1="2.5" y1="3.5" x2="9.5" y2="3.5" stroke-width="1.1"/><line x1="2.5" y1="6" x2="9.5" y2="6" stroke-width="1.1"/><line x1="2.5" y1="8.5" x2="9.5" y2="8.5" stroke-width="1.1"/>';
      case 'Notifier':            return '<path d="M6 1L11 10H1L6 1z" stroke-width="1.3" stroke-linejoin="round"/><line x1="6" y1="4.5" x2="6" y2="7" stroke-width="1.3" stroke-linecap="round"/><circle cx="6" cy="8.8" r="0.7" fill="currentColor" stroke="none"/>';
      case 'ChatBot':             return '<path d="M1.5 2h9a1 1 0 011 1v5a1 1 0 01-1 1H7l-2 2.5V9H1.5a1 1 0 01-1-1V3a1 1 0 011-1z" stroke-width="1.3"/>';
      case 'Image':               return '<rect x="0.5" y="1.5" width="11" height="9" rx="1.5" stroke-width="1.3"/><path d="M0.5 8l3-3 2 2 2-3 4 4.5" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'CheckBox':            return '<rect x="1" y="1" width="10" height="10" rx="1.5" stroke-width="1.3"/><path d="M3 6l2 2 4-4" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'Clock':               return '<circle cx="6" cy="6" r="5" stroke-width="1.3"/><path d="M6 3.5V6l2 1.5" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'WebViewer':           return '<rect x="0.5" y="1.5" width="11" height="9" rx="1" stroke-width="1.3"/><path d="M0.5 4h11" stroke-width="1.1"/><circle cx="2.5" cy="2.8" r="0.6" fill="currentColor" stroke="none"/><circle cx="4.5" cy="2.8" r="0.6" fill="currentColor" stroke="none"/>';
      case 'AbsoluteArrangement': return '<rect x="0.5" y="0.5" width="11" height="11" rx="1" stroke-width="1.3"/><path d="M3 0.5v3M0.5 3h3" stroke-width="1.1" stroke-linecap="round"/><rect x="4" y="4" width="5.5" height="5" rx="0.5" stroke-width="1.1"/>';
      case 'AccelerometerSensor': return '<rect x="3.5" y="3.5" width="5" height="5" rx="1" stroke-width="1.3"/><path d="M6 1v2.5M6 8.5V11M1 6h2.5M8.5 6H11" stroke-width="1.2" stroke-linecap="round"/>';
      case 'ActivityStarter':     return '<path d="M4.5 2H2a1 1 0 00-1 1v7a1 1 0 001 1h8a1 1 0 001-1V7.5" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M7 1.5h3.5V5M10.5 1.5L6.5 5.5" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'AnomalyDetection':    return '<polyline points="1,9 3,7.5 5,7.5 7,3 9,7.5 11,7.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><circle cx="7" cy="3" r="1.5" stroke-width="1.2"/>';
      case 'Ball':                return '<circle cx="6" cy="6" r="4.5" stroke-width="1.3"/><path d="M4 4c.5-1 1.8-1.5 2.8-1" stroke-width="1" stroke-linecap="round"/>';
      case 'BarcodeScanner':      return '<line x1="2" y1="2" x2="2" y2="10" stroke-width="1.5" stroke-linecap="round"/><line x1="3.5" y1="2" x2="3.5" y2="10" stroke-width="0.8" stroke-linecap="round"/><line x1="5" y1="2" x2="5" y2="10" stroke-width="1.5" stroke-linecap="round"/><line x1="6.5" y1="2" x2="6.5" y2="10" stroke-width="0.8" stroke-linecap="round"/><line x1="8" y1="2" x2="8" y2="10" stroke-width="1.5" stroke-linecap="round"/><line x1="9.5" y1="2" x2="9.5" y2="10" stroke-width="0.8" stroke-linecap="round"/><line x1="0.5" y1="6" x2="11" y2="6" stroke-width="1" stroke-linecap="round"/>';
      case 'Barometer':           return '<path d="M2 9.5 a4 4 0 0 1 8 0" stroke-width="1.3" stroke-linecap="round"/><line x1="2" y1="9.5" x2="10" y2="9.5" stroke-width="1.3" stroke-linecap="round"/><line x1="6" y1="9.5" x2="8.5" y2="6" stroke-width="1.3" stroke-linecap="round"/><circle cx="6" cy="9.5" r="0.8" fill="currentColor" stroke="none"/>';
      case 'BluetoothClient':     return '<path d="M4 2v8M4 2l3.5 2-3.5 2M4 10l3.5-2-3.5-2" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M9 6h2M10 5.2l1 .8-1 .8" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'BluetoothServer':     return '<path d="M3 2v8M3 2l3 2-3 2M3 10l3-2-3-2" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><rect x="7.5" y="3.5" width="4" height="2" rx="0.4" stroke-width="1.1"/><rect x="7.5" y="6.5" width="4" height="2" rx="0.4" stroke-width="1.1"/>';
      case 'Camcorder':           return '<rect x="0.5" y="3" width="7.5" height="6.5" rx="1" stroke-width="1.3"/><path d="M8 5.5l3.5-2v5l-3.5-2z" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><circle cx="3.5" cy="6.25" r="1.5" stroke-width="1.1"/>';
      case 'Camera':              return '<rect x="0.5" y="3" width="11" height="7.5" rx="1" stroke-width="1.3"/><path d="M4 3V2h4v1" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><circle cx="6" cy="6.5" r="2" stroke-width="1.3"/>';
      case 'Canvas':              return '<rect x="0.5" y="0.5" width="11" height="11" rx="1" stroke-width="1.3"/><path d="M2.5 9l2.5-4 2 2.5 1.5-2.5 2.5 4" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'Chart':               return '<line x1="1" y1="11" x2="11" y2="11" stroke-width="1.1" stroke-linecap="round"/><rect x="1.5" y="6" width="2" height="5" rx="0.3" stroke-width="1.2"/><rect x="5" y="3.5" width="2" height="7.5" rx="0.3" stroke-width="1.2"/><rect x="8.5" y="5" width="2" height="6" rx="0.3" stroke-width="1.2"/>';
      case 'ChartData2D':         return '<line x1="1" y1="11" x2="11" y2="11" stroke-width="1.1" stroke-linecap="round"/><line x1="1" y1="1" x2="1" y2="11" stroke-width="1.1" stroke-linecap="round"/><path d="M2.5 8.5l2.5-3 2.5 1.5 3-4" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><circle cx="2.5" cy="8.5" r="0.8" fill="currentColor" stroke="none"/><circle cx="5" cy="5.5" r="0.8" fill="currentColor" stroke="none"/><circle cx="7.5" cy="7" r="0.8" fill="currentColor" stroke="none"/><circle cx="10.5" cy="3" r="0.8" fill="currentColor" stroke="none"/>';
      case 'Circle':              return '<circle cx="6" cy="6" r="4.5" stroke-width="1.3"/><line x1="3" y1="6" x2="9" y2="6" stroke-width="0.8" stroke-linecap="round"/><line x1="6" y1="3" x2="6" y2="9" stroke-width="0.8" stroke-linecap="round"/>';
      case 'CircularProgress':    return '<path d="M6 1.5 a4.5 4.5 0 1 1 -4.5 4.5" stroke-width="1.5" stroke-linecap="round"/>';
      case 'CloudDB':             return '<path d="M1 6.5 Q2 4 3.5 6.5 Q6 1.5 8.5 6.5 Q10 4 11 6.5 H1z" stroke-width="1.3" stroke-linejoin="round"/><line x1="2.5" y1="8.5" x2="9.5" y2="8.5" stroke-width="1.2" stroke-linecap="round"/><line x1="2.5" y1="10.5" x2="8" y2="10.5" stroke-width="1.2" stroke-linecap="round"/>';
      case 'ContactPicker':       return '<circle cx="5" cy="3.5" r="2" stroke-width="1.3"/><path d="M1 10.5a4 4 0 018 0" stroke-width="1.3" stroke-linecap="round"/><path d="M9.5 7v3.5M8 9l1.5 1.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'DataFile':            return '<path d="M2.5 1.5h5L10.5 5v5.5a1 1 0 01-1 1h-7a1 1 0 01-1-1v-8a1 1 0 011-1z" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M7.5 1.5V5h3" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/><line x1="4" y1="7" x2="8.5" y2="7" stroke-width="1.1" stroke-linecap="round"/><line x1="4" y1="9" x2="7.5" y2="9" stroke-width="1.1" stroke-linecap="round"/>';
      case 'DatePicker':          return '<rect x="0.5" y="2" width="11" height="9.5" rx="1" stroke-width="1.3"/><line x1="0.5" y1="5.5" x2="11.5" y2="5.5" stroke-width="1.1"/><line x1="3.5" y1="0.5" x2="3.5" y2="3.5" stroke-width="1.3" stroke-linecap="round"/><line x1="8.5" y1="0.5" x2="8.5" y2="3.5" stroke-width="1.3" stroke-linecap="round"/><circle cx="4" cy="8.5" r="0.8" fill="currentColor" stroke="none"/><circle cx="6" cy="8.5" r="0.8" fill="currentColor" stroke="none"/><circle cx="8" cy="8.5" r="0.8" fill="currentColor" stroke="none"/>';
      case 'EmailPicker':         return '<rect x="0.5" y="2.5" width="9" height="7" rx="1" stroke-width="1.3"/><path d="M0.5 3l4.5 3.5 4.5-3.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><path d="M10.5 7v3.5M9 9l1.5 1.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'Ev3ColorSensor':      return '<rect x="2" y="2" width="8" height="8" rx="1.5" stroke-width="1.3"/><circle cx="6" cy="6" r="2" stroke-width="1.2"/><circle cx="6" cy="6" r="0.8" fill="currentColor" stroke="none"/>';
      case 'Ev3Commands':         return '<rect x="0.5" y="0.5" width="11" height="11" rx="1.5" stroke-width="1.3"/><path d="M3 4l2.5 2-2.5 2" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><line x1="7" y1="8" x2="9.5" y2="8" stroke-width="1.3" stroke-linecap="round"/>';
      case 'Ev3GyroSensor':       return '<path d="M3 2.5C1.5 3.5 1 5 1.5 6.5c.5 2 2.5 3.5 5 3.5a4 4 0 003.5-6" stroke-width="1.3" stroke-linecap="round"/><path d="M10 4l-1.5-.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><circle cx="6" cy="6" r="1" fill="currentColor" stroke="none"/>';
      case 'Ev3Motors':           return '<circle cx="6" cy="6" r="4" stroke-width="1.3"/><circle cx="6" cy="6" r="1.5" stroke-width="1.2"/><line x1="6" y1="2" x2="6" y2="4.5" stroke-width="1.1" stroke-linecap="round"/><line x1="6" y1="7.5" x2="6" y2="10" stroke-width="1.1" stroke-linecap="round"/><line x1="2" y1="6" x2="4.5" y2="6" stroke-width="1.1" stroke-linecap="round"/><line x1="7.5" y1="6" x2="10" y2="6" stroke-width="1.1" stroke-linecap="round"/>';
      case 'Ev3Sound':            return '<path d="M1.5 4.5h2l3-3v9l-3-3H1.5z" stroke-width="1.3" stroke-linejoin="round"/><path d="M9 4a3 3 0 010 4" stroke-width="1.3" stroke-linecap="round"/>';
      case 'Ev3TouchSensor':      return '<rect x="1.5" y="5" width="9" height="5.5" rx="1" stroke-width="1.3"/><path d="M6 1.5v3.5" stroke-width="1.3" stroke-linecap="round"/><path d="M4.5 3.5l1.5-2 1.5 2" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><circle cx="6" cy="7.5" r="1.2" stroke-width="1.1"/>';
      case 'Ev3UI':               return '<rect x="0.5" y="2" width="11" height="8" rx="1" stroke-width="1.3"/><rect x="1.5" y="3" width="5.5" height="5" rx="0.5" stroke-width="1.1"/><circle cx="9" cy="5" r="0.9" stroke-width="1.1"/><circle cx="9" cy="7.5" r="0.9" stroke-width="1.1"/>';
      case 'Ev3UltrasonicSensor': return '<circle cx="2.5" cy="6" r="1.5" stroke-width="1.3"/><path d="M5 5 a2 2 0 0 1 0 2" stroke-width="1.2" stroke-linecap="round"/><path d="M6.5 3.5 a3 3 0 0 1 0 5" stroke-width="1.1" stroke-linecap="round"/>';
      case 'FeatureCollection':   return '<path d="M1 6L3.5 2.5 6 6 8 4l3 3.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><circle cx="2.5" cy="9.5" r="1" stroke-width="1.2"/><path d="M5.5 8v3M4 9.5h3" stroke-width="1.1" stroke-linecap="round"/><rect x="8.5" y="8.5" width="3" height="2.5" rx="0.4" stroke-width="1.1"/>';
      case 'File':                return '<path d="M2.5 1.5h5L10.5 5v5.5a1 1 0 01-1 1h-7a1 1 0 01-1-1v-8a1 1 0 011-1z" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M7.5 1.5V5h3" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'FilePicker':          return '<path d="M2 1.5h4L8.5 5v5.5a1 1 0 01-1 1H3a1 1 0 01-1-1V2.5a1 1 0 011-1z" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M6 1.5V5h2.5" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/><path d="M10 7.5v3.5M8.5 9.5l1.5 1.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'FirebaseDB':          return '<ellipse cx="6" cy="3" rx="4" ry="1.2" stroke-width="1.2"/><path d="M2 3v6M10 3v6" stroke-width="1.2" stroke-linecap="round"/><ellipse cx="6" cy="9" rx="4" ry="1.2" stroke-width="1.2"/><path d="M7 4.5L5 7h2.5L5.5 9.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'FusiontablesControl': return '<rect x="0.5" y="0.5" width="11" height="11" rx="1" stroke-width="1.3"/><line x1="0.5" y1="4" x2="11.5" y2="4" stroke-width="1.1"/><line x1="0.5" y1="7.5" x2="11.5" y2="7.5" stroke-width="1.1"/><line x1="4" y1="0.5" x2="4" y2="11.5" stroke-width="1.1"/>';
      case 'GameClient':          return '<rect x="0.5" y="3.5" width="11" height="6" rx="2" stroke-width="1.3"/><path d="M3.5 5.5v2M2.5 6.5h2" stroke-width="1.2" stroke-linecap="round"/><circle cx="8" cy="6" r="0.8" fill="currentColor" stroke="none"/><circle cx="9.5" cy="6" r="0.8" fill="currentColor" stroke="none"/>';
      case 'GyroscopeSensor':     return '<circle cx="6" cy="6" r="4.5" stroke-width="1.3"/><ellipse cx="6" cy="6" rx="4.5" ry="1.8" stroke-width="1.1"/><circle cx="6" cy="6" r="1" fill="currentColor" stroke="none"/>';
      case 'HorizontalScrollArrangement': return '<rect x="0.5" y="1" width="11" height="8.5" rx="1" stroke-width="1.3"/><line x1="4.5" y1="1" x2="4.5" y2="9.5" stroke-width="1.1"/><line x1="8" y1="1" x2="8" y2="9.5" stroke-width="1.1"/><rect x="1" y="10.5" width="5" height="1" rx="0.5" stroke-width="1.1"/>';
      case 'Hygrometer':          return '<path d="M6 1.5C6 1.5 2.5 5.5 2.5 7.5a3.5 3.5 0 007 0C9.5 5.5 6 1.5 6 1.5z" stroke-width="1.3" stroke-linejoin="round"/><path d="M4.5 7.5a1.5 1.5 0 012.5-1" stroke-width="1" stroke-linecap="round"/>';
      case 'ImageBot':            return '<rect x="0.5" y="1.5" width="11" height="9" rx="1.5" stroke-width="1.3"/><path d="M0.5 8l3-3 2 2 2-3 4 4.5" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/><path d="M8.5 1l.5 1.5 1.5.5-1.5.5L8.5 5l-.5-1.5L6.5 3l1.5-.5z" fill="currentColor" stroke="none"/>';
      case 'ImagePicker':         return '<rect x="0.5" y="1.5" width="9" height="8" rx="1.5" stroke-width="1.3"/><path d="M0.5 7l2.5-3 2 2 1.5-2.5 3 4" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/><path d="M10.5 6.5v3.5M9 8.5l1.5 1.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'ImageSprite':         return '<rect x="2" y="2" width="7" height="7" rx="1" stroke-width="1.3"/><path d="M9.5 6h2M10.5 5l1 1-1 1" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><path d="M6 9.5v2M5 10.5l1 1 1-1" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'LightSensor':         return '<circle cx="6" cy="6" r="2.5" stroke-width="1.3"/><path d="M6 1v1.5M6 9.5V11M1 6h1.5M9.5 6H11M2.6 2.6l1 1M8.4 8.4l1 1M9.4 2.6l-1 1M3.6 8.4l-1 1" stroke-width="1.2" stroke-linecap="round"/>';
      case 'LineString':          return '<path d="M1.5 9l3-5 2.5 3 2.5-3 2.5 3" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><circle cx="1.5" cy="9" r="0.9" fill="currentColor" stroke="none"/><circle cx="10.5" cy="7" r="0.9" fill="currentColor" stroke="none"/>';
      case 'LinearProgress':      return '<rect x="0.5" y="4.5" width="11" height="3" rx="1.5" stroke-width="1.3"/><rect x="0.5" y="4.5" width="7" height="3" rx="1.5" fill="currentColor" stroke="none"/>';
      case 'ListPicker':          return '<rect x="0.5" y="0.5" width="9" height="11" rx="1" stroke-width="1.3"/><line x1="2.5" y1="3.5" x2="7.5" y2="3.5" stroke-width="1.1"/><line x1="2.5" y1="6" x2="7.5" y2="6" stroke-width="1.1"/><line x1="2.5" y1="8.5" x2="7.5" y2="8.5" stroke-width="1.1"/><path d="M10.5 5v3.5M9 7l1.5 1.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'LocationSensor':      return '<path d="M6 1a4 4 0 00-4 4c0 3 4 7 4 7s4-4 4-7a4 4 0 00-4-4z" stroke-width="1.3" stroke-linejoin="round"/><circle cx="6" cy="5" r="1.5" stroke-width="1.2"/>';
      case 'MagneticFieldSensor': return '<path d="M3 2v6a3 3 0 006 0V2" stroke-width="1.3" stroke-linecap="round"/><line x1="2" y1="2" x2="4" y2="2" stroke-width="1.5" stroke-linecap="round"/><line x1="8" y1="2" x2="10" y2="2" stroke-width="1.5" stroke-linecap="round"/>';
      case 'Map':                 return '<rect x="0.5" y="0.5" width="11" height="11" rx="1" stroke-width="1.3"/><path d="M0.5 8c2-2 4 1 6 0s2-3 5-2" stroke-width="1.1" stroke-linecap="round"/><circle cx="7" cy="3" r="1.3" stroke-width="1.2"/><line x1="7" y1="4.3" x2="7" y2="5.5" stroke-width="1.2" stroke-linecap="round"/>';
      case 'Marker':              return '<path d="M6 1a3.5 3.5 0 00-3.5 3.5C2.5 7.5 6 11 6 11s3.5-3.5 3.5-6.5A3.5 3.5 0 006 1z" stroke-width="1.3" stroke-linejoin="round"/><circle cx="6" cy="4.5" r="1.2" fill="currentColor" stroke="none"/>';
      case 'MediaStore':          return '<path d="M1 4.5h9a1 1 0 011 1v5a1 1 0 01-1 1H1a1 1 0 01-1-1V5.5a1 1 0 011-1z" stroke-width="1.3" stroke-linejoin="round"/><path d="M1 4.5l1.5-2h3l1.5 2" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M4 6.5v3M3 8h2" stroke-width="1.1" stroke-linecap="round"/><path d="M7 7l2 1-2 1" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'Navigation':          return '<path d="M2 10L6 2l4 8-4-2.5L2 10z" stroke-width="1.3" stroke-linejoin="round"/>';
      case 'NearField':           return '<circle cx="6" cy="6" r="1.5" stroke-width="1.3"/><path d="M3.5 3.5a3.5 3.5 0 015 5" stroke-width="1.2" stroke-linecap="round"/><path d="M1.5 1.5a6 6 0 019 9" stroke-width="1.1" stroke-linecap="round"/>';
      case 'NxtColorSensor':      return '<rect x="1.5" y="2" width="9" height="8" rx="1" stroke-width="1.3"/><circle cx="6" cy="6" r="2" stroke-width="1.2"/><circle cx="6" cy="6" r="0.8" fill="currentColor" stroke="none"/><circle cx="3" cy="4.5" r="0.5" fill="currentColor" stroke="none"/><circle cx="9" cy="4.5" r="0.5" fill="currentColor" stroke="none"/>';
      case 'NxtDirectCommands':   return '<rect x="0.5" y="1.5" width="11" height="9" rx="1" stroke-width="1.3"/><path d="M3 5l2 1.5-2 1.5" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><line x1="6.5" y1="7.5" x2="9" y2="7.5" stroke-width="1.3" stroke-linecap="round"/>';
      case 'NxtDrive':            return '<circle cx="3" cy="8" r="2.5" stroke-width="1.3"/><circle cx="9" cy="8" r="2.5" stroke-width="1.3"/><rect x="3.5" y="2.5" width="5" height="3.5" rx="0.5" stroke-width="1.2"/><line x1="3" y1="5.5" x2="3" y2="8" stroke-width="1" stroke-linecap="round"/><line x1="9" y1="5.5" x2="9" y2="8" stroke-width="1" stroke-linecap="round"/>';
      case 'NxtLightSensor':      return '<rect x="1.5" y="2" width="9" height="8" rx="1" stroke-width="1.3"/><circle cx="6" cy="6" r="2" stroke-width="1.2"/><path d="M6 3.5v1M6 8.5v1M3.5 6h1M7.5 6h1" stroke-width="1" stroke-linecap="round"/>';
      case 'NxtSoundSensor':      return '<rect x="1.5" y="2" width="9" height="8" rx="1" stroke-width="1.3"/><circle cx="5" cy="6" r="1.5" stroke-width="1.2"/><path d="M7 4.5a2 2 0 010 3" stroke-width="1.2" stroke-linecap="round"/>';
      case 'NxtTouchSensor':      return '<rect x="2" y="5" width="8" height="5" rx="1" stroke-width="1.3"/><rect x="4" y="2" width="4" height="3.5" rx="1.5" stroke-width="1.3"/>';
      case 'NxtUltrasonicSensor': return '<rect x="1" y="3" width="5" height="6" rx="1" stroke-width="1.3"/><circle cx="3.5" cy="6" r="1.5" stroke-width="1.2"/><path d="M7 4.5a2 2 0 010 3" stroke-width="1.2" stroke-linecap="round"/><path d="M8 4a2.5 2.5 0 010 4" stroke-width="1.1" stroke-linecap="round"/>';
      case 'OrientationSensor':   return '<rect x="3.5" y="1" width="5" height="8.5" rx="1.5" stroke-width="1.3"/><path d="M1.5 5C1 6.5 1.5 8.5 3.5 9.5" stroke-width="1.2" stroke-linecap="round"/><path d="M1.5 5l1 .5-.5-1.5" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/><path d="M10.5 5C11 6.5 10.5 8.5 8.5 9.5" stroke-width="1.2" stroke-linecap="round"/><path d="M10.5 5l-1 .5.5-1.5" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'PasswordTextBox':     return '<rect x="0.5" y="3" width="11" height="6" rx="1" stroke-width="1.3"/><circle cx="3.5" cy="6" r="0.8" fill="currentColor" stroke="none"/><circle cx="6" cy="6" r="0.8" fill="currentColor" stroke="none"/><circle cx="8.5" cy="6" r="0.8" fill="currentColor" stroke="none"/>';
      case 'Pedometer':           return '<circle cx="6" cy="2.5" r="1.5" stroke-width="1.2"/><path d="M6 4v3.5M4 5.5l2 1.5 2-1.5M4.5 8l1.5 3M7.5 8l-1.5 3" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'PhoneCall':           return '<circle cx="4" cy="3" r="2" stroke-width="1.3"/><circle cx="8" cy="9" r="2" stroke-width="1.3"/><path d="M4 5v3h4v-1" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'PhoneNumberPicker':   return '<circle cx="3.5" cy="3" r="1.5" stroke-width="1.2"/><circle cx="7" cy="8" r="1.5" stroke-width="1.2"/><path d="M3.5 4.5v3h3.5v-.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><path d="M9.5 6.5v4M8 9l1.5 1.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'PhoneStatus':         return '<rect x="3" y="0.5" width="6" height="11" rx="1.5" stroke-width="1.3"/><line x1="3" y1="3" x2="9" y2="3" stroke-width="1.1"/><rect x="4.5" y="7" width="1" height="2" rx="0.2" fill="currentColor" stroke="none"/><rect x="6" y="6" width="1" height="3" rx="0.2" fill="currentColor" stroke="none"/><rect x="7.5" y="5" width="1" height="4" rx="0.2" fill="currentColor" stroke="none"/>';
      case 'Player':              return '<circle cx="6" cy="6" r="5" stroke-width="1.3"/><path d="M4.5 4l4 2-4 2z" fill="currentColor" stroke="none"/>';
      case 'Polygon':             return '<path d="M6 1l4 3.5-1.5 5.5H3.5L2 4.5z" stroke-width="1.3" stroke-linejoin="round"/>';
      case 'ProximitySensor':     return '<rect x="0.5" y="4" width="4.5" height="4" rx="0.8" stroke-width="1.3"/><path d="M6.5 5.5a1 1 0 010 1" stroke-width="1.2" stroke-linecap="round"/><path d="M7.5 4.5a2 2 0 010 3" stroke-width="1.2" stroke-linecap="round"/><path d="M9 3.5a2.5 2.5 0 010 5" stroke-width="1.1" stroke-linecap="round"/>';
      case 'Rectangle':           return '<rect x="2" y="3" width="8" height="6" stroke-width="1.3"/><rect x="1" y="2" width="2" height="2" rx="0.3" fill="currentColor" stroke="none"/><rect x="9" y="2" width="2" height="2" rx="0.3" fill="currentColor" stroke="none"/><rect x="1" y="8" width="2" height="2" rx="0.3" fill="currentColor" stroke="none"/><rect x="9" y="8" width="2" height="2" rx="0.3" fill="currentColor" stroke="none"/>';
      case 'Regression':          return '<line x1="1" y1="11" x2="11" y2="11" stroke-width="1.1" stroke-linecap="round"/><line x1="1" y1="1" x2="1" y2="11" stroke-width="1.1" stroke-linecap="round"/><circle cx="3" cy="8" r="0.8" fill="currentColor" stroke="none"/><circle cx="4.5" cy="6.5" r="0.8" fill="currentColor" stroke="none"/><circle cx="6" cy="5" r="0.8" fill="currentColor" stroke="none"/><circle cx="8" cy="4" r="0.8" fill="currentColor" stroke="none"/><circle cx="10" cy="3" r="0.8" fill="currentColor" stroke="none"/><line x1="2" y1="9" x2="11" y2="2.5" stroke-width="1.2" stroke-linecap="round"/>';
      case 'Serial':              return '<rect x="1" y="3.5" width="10" height="5" rx="1" stroke-width="1.3"/><circle cx="3.5" cy="6" r="0.8" fill="currentColor" stroke="none"/><circle cx="6" cy="6" r="0.8" fill="currentColor" stroke="none"/><circle cx="8.5" cy="6" r="0.8" fill="currentColor" stroke="none"/><line x1="2" y1="8.5" x2="2" y2="10" stroke-width="1.2" stroke-linecap="round"/><line x1="10" y1="8.5" x2="10" y2="10" stroke-width="1.2" stroke-linecap="round"/>';
      case 'Sharing':             return '<circle cx="9.5" cy="2.5" r="1.5" stroke-width="1.2"/><circle cx="9.5" cy="9.5" r="1.5" stroke-width="1.2"/><circle cx="2.5" cy="6" r="1.5" stroke-width="1.2"/><line x1="4" y1="6" x2="8" y2="3" stroke-width="1.2"/><line x1="4" y1="6" x2="8" y2="9" stroke-width="1.2"/>';
      case 'Slider':              return '<line x1="0.5" y1="6" x2="11.5" y2="6" stroke-width="1.3" stroke-linecap="round"/><circle cx="7" cy="6" r="2.5" stroke-width="1.3"/>';
      case 'Sound':               return '<line x1="8" y1="1.5" x2="8" y2="9" stroke-width="1.2" stroke-linecap="round"/><circle cx="6" cy="9" r="2" stroke-width="1.2"/><path d="M8 1.5l3.5 1.5" stroke-width="1.2" stroke-linecap="round"/>';
      case 'SoundRecorder':       return '<rect x="4.5" y="1" width="3" height="6" rx="1.5" stroke-width="1.3"/><path d="M3 7a3 3 0 006 0" stroke-width="1.3" stroke-linecap="round"/><line x1="6" y1="10" x2="6" y2="11.5" stroke-width="1.3" stroke-linecap="round"/><line x1="4" y1="11.5" x2="8" y2="11.5" stroke-width="1.3" stroke-linecap="round"/>';
      case 'SpeechRecognizer':    return '<rect x="3" y="1" width="3" height="5.5" rx="1.5" stroke-width="1.3"/><path d="M2 6.5a2.5 2.5 0 015 0" stroke-width="1.3" stroke-linecap="round"/><line x1="4.5" y1="9" x2="4.5" y2="10" stroke-width="1.2" stroke-linecap="round"/><path d="M8.5 4h2.5M8.5 6.5h3M8.5 9h2" stroke-width="1.1" stroke-linecap="round"/>';
      case 'Spinner':             return '<rect x="0.5" y="3.5" width="11" height="5" rx="1" stroke-width="1.3"/><line x1="8" y1="3.5" x2="8" y2="8.5" stroke-width="1.1"/><path d="M9.5 5.5l1 1 1-1" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/><line x1="2" y1="6" x2="6" y2="6" stroke-width="1.1" stroke-linecap="round"/>';
      case 'Spreadsheet':         return '<rect x="0.5" y="0.5" width="11" height="11" rx="1" stroke-width="1.3"/><line x1="0.5" y1="3.5" x2="11.5" y2="3.5" stroke-width="1.1"/><line x1="0.5" y1="7" x2="11.5" y2="7" stroke-width="1.1"/><line x1="4" y1="3.5" x2="4" y2="11.5" stroke-width="1.1"/><line x1="7.5" y1="3.5" x2="7.5" y2="11.5" stroke-width="1.1"/>';
      case 'Switch':              return '<rect x="0.5" y="4" width="11" height="4" rx="2" stroke-width="1.3"/><circle cx="8.5" cy="6" r="1.8" stroke-width="1.3"/>';
      case 'TableArrangement':    return '<rect x="0.5" y="0.5" width="11" height="11" rx="1" stroke-width="1.3"/><line x1="0.5" y1="4.5" x2="11.5" y2="4.5" stroke-width="1.1"/><line x1="0.5" y1="8.5" x2="11.5" y2="8.5" stroke-width="1.1"/><line x1="4.5" y1="0.5" x2="4.5" y2="11.5" stroke-width="1.1"/><line x1="8.5" y1="0.5" x2="8.5" y2="11.5" stroke-width="1.1"/>';
      case 'TextToSpeech':        return '<path d="M1 3h5M1 5.5h4M1 8h5" stroke-width="1.2" stroke-linecap="round"/><path d="M7 4.5h1l3-2v6l-3-2H7z" stroke-width="1.2" stroke-linejoin="round"/>';
      case 'Texting':             return '<path d="M1.5 2h9a1 1 0 011 1v5a1 1 0 01-1 1H3.5L1 11V9H1.5a1 1 0 01-1-1V3a1 1 0 011-1z" stroke-width="1.3"/><line x1="3.5" y1="4.5" x2="9.5" y2="4.5" stroke-width="1" stroke-linecap="round"/><line x1="3.5" y1="6.5" x2="8" y2="6.5" stroke-width="1" stroke-linecap="round"/>';
      case 'Thermometer':         return '<rect x="5" y="1" width="2" height="8" rx="1" stroke-width="1.3"/><circle cx="6" cy="9.5" r="2" stroke-width="1.3"/><rect x="5.5" y="4" width="1" height="6" rx="0.5" fill="currentColor" stroke="none"/>';
      case 'TimePicker':          return '<circle cx="4.5" cy="6" r="4" stroke-width="1.3"/><path d="M4.5 3.5V6l1.5 1.2" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><path d="M9.5 3.5v4.5M8 6.5l1.5 1.5 1.5-1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'TinyDB':              return '<ellipse cx="6" cy="3" rx="3.5" ry="1" stroke-width="1.2"/><path d="M2.5 3v6M9.5 3v6" stroke-width="1.2" stroke-linecap="round"/><ellipse cx="6" cy="9" rx="3.5" ry="1" stroke-width="1.2"/><circle cx="6" cy="6" r="1.2" stroke-width="1.1"/>';
      case 'TinyWebDB':           return '<ellipse cx="6" cy="3.5" rx="3.5" ry="1" stroke-width="1.2"/><path d="M2.5 3.5v5.5M9.5 3.5v5.5" stroke-width="1.2" stroke-linecap="round"/><ellipse cx="6" cy="9" rx="3.5" ry="1" stroke-width="1.2"/><circle cx="6" cy="6.5" r="1.5" stroke-width="1.1"/><line x1="4.5" y1="6.5" x2="7.5" y2="6.5" stroke-width="0.9" stroke-linecap="round"/>';
      case 'Translator':          return '<path d="M1 1.5h6v4.5H3L1 7.5z" stroke-width="1.2" stroke-linejoin="round"/><path d="M5 5.5h6v4.5H9l-2 2v-2H5z" stroke-width="1.2" stroke-linejoin="round"/>';
      case 'Trendline':           return '<line x1="1" y1="11" x2="11" y2="11" stroke-width="1.1" stroke-linecap="round"/><line x1="1" y1="1" x2="1" y2="11" stroke-width="1.1" stroke-linecap="round"/><path d="M2 9.5C3 8 4 6.5 6 5S9 3.5 11 3" stroke-width="1.3" stroke-linecap="round"/>';
      case 'Twitter':             return '<path d="M11 2a5 5 0 01-7.5 6.5C2 10 1 11 1 11c2-1 3.5-2.5 4-4A5 5 0 0111 2z" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      case 'VerticalScrollArrangement': return '<rect x="1" y="0.5" width="8.5" height="11" rx="1" stroke-width="1.3"/><line x1="1" y1="4.5" x2="9.5" y2="4.5" stroke-width="1.1"/><line x1="1" y1="8" x2="9.5" y2="8" stroke-width="1.1"/><rect x="10.5" y="1.5" width="1" height="4" rx="0.5" stroke-width="1.1"/>';
      case 'VideoPlayer':         return '<rect x="0.5" y="2" width="11" height="8" rx="1" stroke-width="1.3"/><path d="M4.5 5l4 2-4 2z" fill="currentColor" stroke="none"/>';
      case 'Voting':              return '<rect x="1.5" y="1.5" width="9" height="9" rx="1" stroke-width="1.3"/><path d="M3.5 4.5l1.5 1.5 3-3" stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"/><line x1="3.5" y1="7.5" x2="8.5" y2="7.5" stroke-width="1.2" stroke-linecap="round"/>';
      case 'Web':                 return '<circle cx="6" cy="6" r="5" stroke-width="1.3"/><path d="M6 1c-1.5 1.5-2.5 3-2.5 5S4.5 9.5 6 11M6 1c1.5 1.5 2.5 3 2.5 5S7.5 9.5 6 11" stroke-width="1" stroke-linecap="round"/><line x1="1" y1="6" x2="11" y2="6" stroke-width="1" stroke-linecap="round"/>';
      case 'YandexTranslate':     return '<circle cx="5" cy="6" r="4.5" stroke-width="1.3"/><path d="M3 3.5c1 .5 2.5.5 4 0M2.5 6h5M3 8.5c1-.5 2.5-.5 4 0" stroke-width="0.9" stroke-linecap="round"/><path d="M9.5 4.5l2 1.5-2 1.5" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>';
      default:                    return '<rect x="1.5" y="1.5" width="9" height="9" rx="1.5" stroke-width="1.3"/>';
    }
  }
</script>

<svelte:window on:keydown={handleListDataDialogKey} />

<div class="vis-body">
  {#if parseError}
    <div class="vis-parse-error">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true">
        <path d="M12 3l9 16H3L12 3z" stroke-linejoin="round"/>
        <path d="M12 9v4M12 17h.01" stroke-linecap="round"/>
      </svg>
      <div class="vis-parse-copy">
        <h3>Schema needs a text edit</h3>
        <p>{parseError}</p>
      </div>
      <button class="vis-parse-btn" on:click={() => dispatch('switchText')}>
        Switch to Text
      </button>
    </div>
  {:else}
  <!-- ── Component tree ──────────────────────────────────────────────── -->
  <div class="vis-tree">
    <div class="vis-tree-tabs">
      <button
        class="vis-tab-btn"
        class:active={treeTab === 'components'}
        on:click={() => { treeTab = 'components'; closeCtxMenu(); }}
      >Components</button>
      <button
        class="vis-tab-btn"
        class:active={treeTab === 'media'}
        on:click={() => { treeTab = 'media'; closeCtxMenu(); }}
      >Media</button>
      <div class="vis-tab-spacer"></div>
      {#if treeTab === 'components'}
        <button
          class="vis-tab-add-btn"
          title="Add component"
          on:click={ctxAddComp}
        >
          <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round">
            <path d="M5 1v8M1 5h8"/>
          </svg>
        </button>
      {/if}
    </div>

  {#if treeTab === 'components'}
    <div
      class="vis-tree-scroll"
      role="tree"
      aria-label="Components"
      tabindex="-1"
      on:dragleave={handleScrollDragLeave}
    >
      {#if flatList.length}
        {#each flatList as node (node.pathId)}
          {@const hasKids  = node.children.length > 0}
          {@const isOpen   = !collapsed.has(node.pathId)}
          {@const isSel    = selectedNode?.pathId === node.pathId}
          {@const renaming = renamingPathId === node.pathId}
          {@const dropPos  = dropTarget?.pathId === node.pathId ? dropTarget.position : null}

          <div
            class="vis-item"
            class:selected={isSel}
            class:search-flash-row={treeFlashPathId === node.pathId}
            class:drag-source={dragPathId === node.pathId}
            class:drop-before={dropPos === 'before'}
            class:drop-after={dropPos === 'after'}
            class:drop-into={dropPos === 'into'}
            draggable={node.pathId !== '0' ? 'true' : 'false'}
            style="padding-left: {6 + node.depth * 14}px"
            role="treeitem"
            aria-selected={isSel}
            tabindex="0"
            on:click={() => { selectedPathId = node.pathId; showPicker = false; closeCtxMenu(); }}
            on:keydown={e => e.key === 'Enter' && (selectedPathId = node.pathId)}
            on:contextmenu={e => handleItemCtxMenu(e, node)}
            on:dragstart={e => node.pathId !== '0' && handleDragStart(e, node.pathId)}
            on:dragover={e => handleDragOver(e, node.pathId)}
            on:drop={e => handleDrop(e, node.pathId)}
            on:dragend={handleDragEnd}
          >
            <button
              class="vis-arrow"
              class:vis-arrow--active={hasKids}
              class:vis-arrow--open={hasKids && isOpen}
              tabindex={hasKids ? 0 : -1}
              on:click|stopPropagation={() => hasKids && toggleCollapse(node.pathId)}
            >
              {#if hasKids}
                <svg viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M2 3l2 2 2-2"/>
                </svg>
              {/if}
            </button>

            <svg class="vis-type-icon" viewBox="0 0 12 12" fill="none" stroke="currentColor">
              {@html compIcon(node.type)}
            </svg>

            {#if renaming}
              <input
                class="vis-rename-input"
                class:error={!!renameError}
                bind:this={renameInputEl}
                bind:value={renameValue}
                on:input={() => { renameError = ''; }}
                on:blur={commitRename}
                on:keydown={handleRenameKey}
                on:click|stopPropagation
              />
              {#if renameError}
                <span class="vis-rename-error">{renameError}</span>
              {/if}
            {:else}
              <span class="vis-item-name" title={node.name}>{node.name}</span>
            {/if}

            <button
              class="vis-item-menu-btn"
              tabindex="-1"
              title="More options"
              aria-label="More options for {node.name}"
              on:click|stopPropagation={e => handleItemCtxMenu(e, node)}
            >
              <svg viewBox="0 0 12 12" fill="currentColor">
                <circle cx="2.5" cy="6" r="1"/>
                <circle cx="6" cy="6" r="1"/>
                <circle cx="9.5" cy="6" r="1"/>
              </svg>
            </button>
          </div>
        {/each}
      {:else}
        <div class="vis-tree-empty">No components yet.</div>
      {/if}
    </div>

    <!-- Add component input -->
    {#if showPicker}
      {#if addCompSuggestions.length}
        <div class="vis-add-suggestions">
          {#each addCompSuggestions as name, i}
            <button
              class="vis-add-suggestion"
              class:highlighted={i === addCompHighlight}
              on:mousedown={() => { addCompSelecting = true; }}
              on:click={() => pickSuggestion(name)}
              tabindex="-1"
            >
              <svg class="vis-add-suggestion-icon" viewBox="0 0 12 12" fill="none" stroke="currentColor">{@html compIcon(name)}</svg>
              {name}
            </button>
          {/each}
        </div>
      {/if}
      <div class="vis-add-bar">
        <input
          class="vis-add-input"
          bind:this={addCompInputEl}
          bind:value={addCompValue}
          placeholder="Component name, e.g. Button"
          class:error={!!addCompError}
          on:input={() => { addCompError = ''; }}
          on:keydown={handleAddCompKey}
          on:blur={() => { if (addCompSelecting) return; showPicker = false; addCompValue = ''; addCompError = ''; }}
          spellcheck="false"
          autocomplete="off"
        />
        {#if addCompError}
          <div class="vis-add-error">{addCompError}</div>
        {/if}
      </div>
    {/if}

  {:else}
    <!-- Media tab -->
    <div class="vis-media-tab">
      <div class="vis-media-header">
        <span>Media</span>
        <label class="vis-media-upload">
          Upload file
          <input
            class="vis-file-native"
            type="file"
            multiple
            on:change={handleMediaUpload}
          />
        </label>
      </div>
      <div class="vis-media-list" role="list" aria-label="Media assets">
        {#if mediaAssets.length}
          {#each mediaAssets as asset (asset.id)}
            {@const assetSize = formatAssetSize(asset.size)}
            <div
              class="vis-media-item"
              role="listitem"
              title={assetSize ? `${asset.name} - ${assetSize}` : asset.name}
              on:contextmenu={e => handleAssetCtxMenu(e, asset)}
            >
              <svg class="vis-media-icon" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.3" stroke-linejoin="round">
                <path d="M2.5 1.5h4L9.5 4v6.5h-7z"/>
                <path d="M6.5 1.5V4h3"/>
              </svg>
              {#if renamingAssetId === asset.id}
                <div class="vis-media-rename">
                  <input
                    class="vis-media-rename-input"
                    class:error={!!assetRenameError}
                    bind:this={assetRenameInputEl}
                    bind:value={assetRenameValue}
                    on:input={() => { assetRenameError = ''; }}
                    on:blur={commitAssetRename}
                    on:keydown={handleAssetRenameKey}
                    on:click|stopPropagation
                  />
                  {#if assetRenameError}
                    <span class="vis-media-rename-error">{assetRenameError}</span>
                  {/if}
                </div>
              {:else}
                <span class="vis-media-name">{asset.name}</span>
                {#if assetSize}
                  <span class="vis-media-size">{assetSize}</span>
                {/if}
                <button
                  class="vis-item-menu-btn"
                  tabindex="-1"
                  title="More options"
                  aria-label="More options for {asset.name}"
                  on:click|stopPropagation={e => handleAssetCtxMenu(e, asset)}
                >
                  <svg viewBox="0 0 12 12" fill="currentColor">
                    <circle cx="2.5" cy="6" r="1"/>
                    <circle cx="6" cy="6" r="1"/>
                    <circle cx="9.5" cy="6" r="1"/>
                  </svg>
                </button>
              {/if}
            </div>
          {/each}
        {:else}
          <div class="vis-media-empty">No media files</div>
        {/if}
      </div>
    </div>
  {/if}

  </div>

  <!-- ── Context menu ──────────────────────────────────────────────────── -->
  {#if ctxMenu}
    <div class="vis-ctx-backdrop" role="presentation" on:click={closeCtxMenu} on:keydown></div>
    <div
      class="vis-ctx-menu"
      style="left:{ctxMenu.x}px; top:{ctxMenu.y}px"
      role="menu"
      tabindex="-1"
      on:click|stopPropagation
      on:keydown|stopPropagation
    >
      {#if ctxMenu.isContainer}
        <button class="vis-ctx-item" role="menuitem" on:click={ctxAddComp}>
          <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
            <path d="M6 1v10M1 6h10"/>
          </svg>
          Add component
        </button>
        <div class="vis-ctx-sep"></div>
      {/if}
      <button class="vis-ctx-item" role="menuitem" disabled={ctxMenu.isRoot} on:click={ctxRename}>
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M8.5 1.5a1.414 1.414 0 012 2L4 10l-3 1 1-3 6.5-6.5z"/>
        </svg>
        Rename
      </button>
      <div class="vis-ctx-sep"></div>
      <button class="vis-ctx-item vis-ctx-item--danger" role="menuitem" disabled={ctxMenu.isRoot} on:click={ctxDelete}>
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M2 3h8M5 3V2h2v1M10 3l-.6 7H2.6L2 3"/>
        </svg>
        Delete
      </button>
    </div>
  {/if}

  {#if assetCtxMenu && activeAsset}
    <div class="vis-ctx-backdrop" role="presentation" on:click={closeCtxMenu} on:keydown></div>
    <div
      class="vis-ctx-menu"
      style="left:{assetCtxMenu.x}px; top:{assetCtxMenu.y}px"
      role="menu"
      tabindex="-1"
      on:click|stopPropagation
      on:keydown|stopPropagation
    >
      <button class="vis-ctx-item" role="menuitem" on:click={() => startAssetRename(activeAsset)}>
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M8.5 1.5a1.414 1.414 0 012 2L4 10l-3 1 1-3 6.5-6.5z"/>
        </svg>
        Rename
      </button>
      <button class="vis-ctx-item" role="menuitem" disabled={!activeAsset.url} on:click={() => downloadAsset(activeAsset)}>
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M6 1v6M3.5 4.5L6 7l2.5-2.5M2 10h8"/>
        </svg>
        Download
      </button>
      <div class="vis-ctx-sep"></div>
      <button class="vis-ctx-item vis-ctx-item--danger" role="menuitem" on:click={() => removeAsset(activeAsset)}>
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M2 3h8M5 3V2h2v1M10 3l-.6 7H2.6L2 3"/>
        </svg>
        Delete
      </button>
    </div>
  {/if}

  <!-- ── Property editor ──────────────────────────────────────────────── -->
  <div class="vis-props">
    {#if selectedNode}
      <div class="vis-props-header">
        <svg class="vis-props-type-icon" viewBox="0 0 12 12" fill="none" stroke="currentColor">
          {@html compIcon(selectedNode.type)}
        </svg>
        <span class="vis-props-type">{selectedNode.type}</span>
        <span class="vis-props-dot">·</span>
        <span class="vis-props-title">{selectedNode.name}</span>
      </div>

      <div class="vis-props-scroll">
        {#each propGroups as group}
          <div class="vis-section">
            <button
              class="vis-section-hd"
              on:click={() => toggleCategory(group.cat)}
              aria-expanded={!collapsedCategories.has(group.cat)}
            >
              <svg
                class="vis-section-caret"
                class:vis-section-caret--open={!collapsedCategories.has(group.cat)}
                viewBox="0 0 8 8" fill="none" stroke="currentColor"
                stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"
              ><path d="M2 3l2 2 2-2"/></svg>
              <span class="vis-section-title">{group.cat}</span>
            </button>

            {#if !collapsedCategories.has(group.cat)}
              <div class="vis-section-body">
                {#each group.items as prop}
                  {@const val = propVal(selectedNode, prop.name)}
                  {@const editorVal = propDisplayVal(prop, val)}
                  {@const options = selectOptions(prop)}
                  <div class="vis-row" class:search-flash-row={flashPropName === prop.name} data-prop-name={prop.name}>
                    <div class="vis-row-label">
                      <span class="vis-prop-name">{prop.name}</span>
                      <button
                        class="vis-help-btn"
                        aria-label="Show property description"
                        on:click={e => openHelp(e, prop.description)}
                      >
                        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4">
                          <circle cx="6" cy="6" r="5"/>
                          <path d="M5 4.5C5 4 5.4 3.5 6 3.5c.7 0 1 .5 1 1 0 .8-1 1-1 2" stroke-linecap="round"/>
                          <circle cx="6" cy="8.5" r=".5" fill="currentColor" stroke="none"/>
                        </svg>
                      </button>
                    </div>
                    <div class="vis-row-editor">
                      {#if prop.editorType === 'color'}
                        {@const cd = colorDisplay(editorVal)}
                        <div class="vis-color-control">
                          <label class="vis-color-pill">
                            <input
                              class="vis-color-native"
                              type="color"
                              value={aiColorToHex(editorVal)}
                              on:input={e => updateProp(selectedNode.pathId, prop.name, hexToAiColor(e.currentTarget.value))}
                            />
                            <span class="vis-color-dot" style="background:{cd.hex}"></span>
                            <span class="vis-color-text">{cd.label}</span>
                            <svg class="vis-pill-caret" viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M1.5 3l2.5 2.5 2.5-2.5"/></svg>
                          </label>
                          <button
                            class="vis-color-none"
                            title="Use default color"
                            on:click={() => updateProp(selectedNode.pathId, prop.name, '&H00000000')}
                          >
                            None
                          </button>
                        </div>
                      {:else if prop.editorType === 'boolean'}
                        <label class="vis-check-wrap">
                          <input
                            type="checkbox"
                            class="vis-check-native"
                            checked={boolVal(val, boolDefault(prop))}
                            on:change={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.checked ? 'true' : 'false')}
                          />
                          <span class="vis-check-box"></span>
                        </label>
                      {:else if prop.editorType === 'non_negative_float' || prop.editorType === 'float'}
                        <input
                          type="number"
                          class="vis-input"
                          value={editorVal || '0'}
                          step="0.1"
                          min={prop.editorType === 'non_negative_float' ? '0' : null}
                          on:input={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.value)}
                        />
                      {:else if prop.editorType === 'integer' || prop.editorType === 'non_negative_integer'}
                        <input
                          type="number"
                          class="vis-input"
                          value={editorVal || '0'}
                          step="1"
                          min={prop.editorType === 'non_negative_integer' ? '0' : null}
                          on:input={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.value)}
                        />
                      {:else if prop.editorType === 'layout_size'}
                        {@const mode = layoutMode(val)}
                        {@const layoutLabels = { automatic: 'Automatic', fill: 'Fill parent', custom: 'Pixels' }}
                        {@const layoutKey = `layout-${selectedNode.pathId}-${prop.name}`}
                        <div class="vis-layout-edit">
                          <button
                            class="vis-select-btn"
                            class:open={visDropdownKey === layoutKey}
                            on:click={e => openVisDropdown(e, layoutKey,
                              [{ value: 'automatic', label: 'Automatic' }, { value: 'fill', label: 'Fill parent' }, { value: 'custom', label: 'Pixels' }],
                              mode,
                              v => updateLayoutMode(selectedNode.pathId, prop.name, v, val)
                            )}
                          >
                            <span class="vis-select-btn-label">{layoutLabels[mode] ?? mode}</span>
                            <svg class="vis-select-btn-chevron" class:rotated={visDropdownKey === layoutKey} viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M1.5 3l2.5 2.5 2.5-2.5"/></svg>
                          </button>
                          {#if mode === 'custom'}
                            <input
                              type="number"
                              class="vis-input vis-layout-pixels"
                              value={layoutPixels(val)}
                              min="1"
                              step="1"
                              aria-label="{prop.name} pixels"
                              on:input={e => updateLayoutPixels(selectedNode.pathId, prop.name, e.currentTarget.value)}
                            />
                          {/if}
                        </div>
                      {:else if prop.editorType === 'string' || prop.editorType === 'text'}
                        <input
                          type="text"
                          class="vis-input"
                          value={editorVal}
                          on:input={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.value)}
                        />
                      {:else if prop.editorType === 'textArea'}
                        <textarea
                          class="vis-textarea"
                          value={editorVal}
                          on:input={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.value)}
                        ></textarea>
                      {:else if prop.editorType === 'asset'}
                        {@const assetKey = `asset-${selectedNode.pathId}-${prop.name}`}
                        <div class="vis-asset-control">
                          <button
                            class="vis-select-btn"
                            class:open={visDropdownKey === assetKey}
                            on:click={e => openVisDropdown(e, assetKey,
                              [{ value: '', label: 'None...' }, ...assetOptions(editorVal).map(a => ({ value: a, label: a }))],
                              editorVal || '',
                              v => updateProp(selectedNode.pathId, prop.name, v)
                            )}
                          >
                            <span class="vis-select-btn-label">{editorVal || 'None...'}</span>
                            <svg class="vis-select-btn-chevron" class:rotated={visDropdownKey === assetKey} viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M1.5 3l2.5 2.5 2.5-2.5"/></svg>
                          </button>
                          <label class="vis-asset-upload" title="Choose asset file">
                            <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
                              <path d="M6 1v7M3.5 3.5L6 1l2.5 2.5M2 10h8"/>
                            </svg>
                            <input
                              class="vis-file-native"
                              type="file"
                              on:change={e => handleAssetPicked(e, selectedNode.pathId, prop.name)}
                            />
                          </label>
                        </div>
                      {:else if options.length > 0}
                        {@const optKey = `opt-${selectedNode.pathId}-${prop.name}`}
                        {@const curOptVal = selectValue(prop, val)}
                        {@const curOptLabel = options.find(o => String(o.value) === curOptVal)?.label ?? curOptVal}
                        <button
                          class="vis-select-btn"
                          class:open={visDropdownKey === optKey}
                          on:click={e => openVisDropdown(e, optKey,
                            options.map(o => ({ value: String(o.value), label: o.label })),
                            curOptVal,
                            v => updateProp(selectedNode.pathId, prop.name, v)
                          )}
                        >
                          <span class="vis-select-btn-label">{curOptLabel}</span>
                          <svg class="vis-select-btn-chevron" class:rotated={visDropdownKey === optKey} viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"><path d="M1.5 3l2.5 2.5 2.5-2.5"/></svg>
                        </button>
                      {:else if prop.customEditor?.editorType === 'ListViewAddData'}
                        {@const layoutValue = String(propVal(selectedNode, 'ListViewLayout') ?? '0')}
                        <div class="vis-listdata-control">
                          <span class="vis-listdata-summary">{listViewDataSummary(editorVal, layoutValue)}</span>
                          <button
                            class="vis-listdata-edit-btn"
                            on:click={() => openListDataDialog(selectedNode.pathId)}
                          >Edit rows</button>
                        </div>
                      {:else}
                        <input
                          type="text"
                          class="vis-input"
                          value={editorVal}
                          on:input={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.value)}
                        />
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {:else}
      <div class="vis-empty">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.2">
          <rect x="3" y="3" width="18" height="18" rx="2"/>
          <path d="M3 9h18M9 3v18"/>
        </svg>
        <span>Select a component</span>
      </div>
    {/if}
  </div>
  {/if}

  <!-- ── Property dropdown ─────────────────────────────────────────────── -->
  {#if visDropdownKey}
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div class="vis-ctx-backdrop" role="presentation" on:click={closeVisDropdown} on:keydown></div>
    <!-- svelte-ignore a11y-click-events-have-key-events -->
    <!-- svelte-ignore a11y-no-static-element-interactions -->
    <div
      class="vis-ctx-menu vis-dropdown"
      style="left:{visDropdownX}px; top:{visDropdownY}px; min-width:{visDropdownMinWidth}px"
      role="listbox"
      tabindex="-1"
      on:click|stopPropagation
      on:keydown={e => e.key === 'Escape' && closeVisDropdown()}
    >
      {#each visDropdownOptions as opt}
        <button
          type="button"
          class="vis-ctx-item"
          class:vis-option-active={opt.value === visDropdownCurrent}
          role="option"
          aria-selected={opt.value === visDropdownCurrent}
          on:click={() => { visDropdownOnSelect?.(opt.value); closeVisDropdown(); }}
        >{opt.label}</button>
      {/each}
    </div>
  {/if}

  <!-- ── Help callout ─────────────────────────────────────────────────── -->
  {#if helpPopup}
    <div class="vis-ctx-backdrop" role="presentation" on:click={closeHelp} on:keydown></div>
    <div
      class="vis-help-callout"
      style="left:{helpPopup.left}px; bottom:{helpPopup.bottomPx}px; --arrow-left:{helpPopup.arrowLeft}px;"
      role="tooltip"
    >{helpPopup.description}</div>
  {/if}

  <!-- ── ListData dialog ──────────────────────────────────────────────── -->
  {#if listDataDialogOpen}
    <button
      type="button"
      class="vis-listdata-backdrop"
      aria-label="Close List Data dialog"
      on:click={cancelListDataDialog}
    ></button>
    <div
      class="vis-listdata-dialog"
      role="dialog"
      aria-modal="true"
      aria-label="Edit List Data"
    >
      <div class="vis-listdata-hd">
        <span class="vis-listdata-title">Edit List Data</span>
        <button class="vis-listdata-close" on:click={cancelListDataDialog} aria-label="Close">
          <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
            <path d="M2 2l8 8M10 2l-8 8"/>
          </svg>
        </button>
      </div>

      <div class="vis-listdata-table-wrap">
        <table class="vis-listdata-table">
          <thead>
            <tr>
              <th class="vis-listdata-th vis-listdata-th-order"></th>
              {#each listDataDialogColumns as col}
                <th class="vis-listdata-th">{col}</th>
              {/each}
              <th class="vis-listdata-th vis-listdata-th-del"></th>
            </tr>
          </thead>
          <tbody>
            {#each listDataDialogRows as row, i (i)}
              <tr class="vis-listdata-tr">
                <td class="vis-listdata-td vis-listdata-td-order">
                  <div class="vis-listdata-reorder">
                    <button
                      class="vis-listdata-reorder-btn"
                      disabled={i === 0}
                      on:click={() => moveListDataRow(i, -1)}
                      aria-label="Move row up"
                    >
                      <svg viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M4 6V2M2 4l2-2 2 2"/>
                      </svg>
                    </button>
                    <button
                      class="vis-listdata-reorder-btn"
                      disabled={i === listDataDialogRows.length - 1}
                      on:click={() => moveListDataRow(i, 1)}
                      aria-label="Move row down"
                    >
                      <svg viewBox="0 0 8 8" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M4 2v4M2 4l2 2 2-2"/>
                      </svg>
                    </button>
                  </div>
                </td>
                {#each listDataDialogColumns as col}
                  <td class="vis-listdata-td">
                    {#if col === 'Image'}
                      <select
                        class="vis-listdata-select"
                        value={row[col] ?? 'None'}
                        on:change={e => updateListDataCell(i, col, e.currentTarget.value)}
                      >
                        <option value="None">None</option>
                        {#each mediaAssets.map(a => a.name).filter(Boolean) as assetName}
                          <option value={assetName}>{assetName}</option>
                        {/each}
                      </select>
                    {:else}
                      <input
                        type="text"
                        class="vis-listdata-input"
                        value={row[col] ?? ''}
                        on:input={e => updateListDataCell(i, col, e.currentTarget.value)}
                        placeholder={col}
                      />
                    {/if}
                  </td>
                {/each}
                <td class="vis-listdata-td vis-listdata-td-del">
                  <button
                    class="vis-listdata-del-btn"
                    on:click={() => removeListDataRow(i)}
                    aria-label="Remove row"
                  >
                    <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M2 3h8M5 3V2h2v1M10 3l-.6 7H2.6L2 3"/>
                    </svg>
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
        {#if listDataDialogRows.length === 0}
          <div class="vis-listdata-empty-rows">No rows yet. Click "Add row" to begin.</div>
        {/if}
      </div>

      <div class="vis-listdata-footer">
        <button class="vis-listdata-add-btn" on:click={addListDataRow}>
          <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round">
            <path d="M5 1v8M1 5h8"/>
          </svg>
          Add row
        </button>
        <div class="vis-listdata-footer-actions">
          <button class="vis-listdata-cancel-btn" on:click={cancelListDataDialog}>Cancel</button>
          <button class="vis-listdata-save-btn" on:click={saveListDataDialog}>Save</button>
        </div>
      </div>
    </div>
  {/if}
</div>

<style>
  /* ── Shell ──────────────────────────────────────────────────────────── */
  .vis-body {
    display: flex;
    flex: 1;
    overflow: hidden;
    height: 100%;
  }

  .vis-parse-error {
    flex: 1;
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 18px;
    background: var(--surface);
    color: var(--text);
  }

  .vis-parse-error > svg {
    width: 22px;
    height: 22px;
    flex-shrink: 0;
    color: var(--error);
  }

  .vis-parse-copy {
    flex: 1;
    min-width: 0;
  }

  .vis-parse-copy h3 {
    margin: 0 0 4px;
    font-family: var(--font);
    font-size: 13px;
    font-weight: 500;
    color: var(--text);
  }

  .vis-parse-copy p {
    margin: 0;
    color: var(--text-muted);
    font-family: var(--mono);
    font-size: 11.5px;
    line-height: 1.45;
  }

  .vis-parse-btn {
    flex-shrink: 0;
    height: 26px;
    padding: 0 10px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text);
    font-family: var(--font);
    font-size: 12px;
    cursor: pointer;
    transition: background 0.1s, border-color 0.1s;
  }
  .vis-parse-btn:hover { border-color: var(--text-faint); background: var(--cell-active); }
  .vis-parse-btn:active { opacity: 0.7; transition: opacity 0.05s; }

  /* ── Tree panel ─────────────────────────────────────────────────────── */
  .vis-tree {
    flex: 0 1 45%;
    max-width: 45%;
    min-width: 0;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border);
    background: var(--bg);
    overflow: hidden;
    position: relative;
  }

  /* ── Tab bar ─────────────────────────────────────────────────────── */
  .vis-tree-tabs {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 6px;
    border-bottom: 1px solid var(--border-soft);
    background: var(--surface);
  }

  .vis-tab-btn {
    height: 22px;
    padding: 0 8px;
    border: none;
    border-radius: var(--radius);
    background: transparent;
    color: var(--text-faint);
    font-family: var(--font);
    font-size: 11.5px;
    cursor: pointer;
    transition: background 0.1s, color 0.1s;
  }
  .vis-tab-btn:hover { background: var(--cell-active); color: var(--text-muted); }
  .vis-tab-btn.active { background: var(--cell-active); color: var(--text); font-weight: 500; }

  .vis-tab-spacer { flex: 1; }

  .vis-tab-add-btn {
    width: 22px;
    height: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    border-radius: var(--radius);
    background: transparent;
    color: var(--text-faint);
    cursor: pointer;
    transition: background 0.1s, color 0.1s;
    flex-shrink: 0;
  }
  .vis-tab-add-btn:hover { background: var(--cell-active); color: var(--accent); }
  .vis-tab-add-btn svg { width: 10px; height: 10px; }

  .vis-tree-scroll {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    padding: 5px 0;
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
    min-height: 0;
  }
  .vis-tree-scroll::-webkit-scrollbar { width: 4px; }
  .vis-tree-scroll::-webkit-scrollbar-track { background: transparent; }
  .vis-tree-scroll::-webkit-scrollbar-thumb { background: var(--border); border-radius: 99px; }

  .vis-tree-empty {
    padding: 16px 12px;
    color: var(--text-faint);
    font-family: var(--font);
    font-size: 12px;
    text-align: center;
  }

  /* Tree items */
  .vis-item {
    position: relative;
    display: flex;
    align-items: center;
    gap: 4px;
    height: 26px;
    padding-right: 6px;
    cursor: pointer;
    user-select: none;
    outline: none;
    transition: background 0.1s;
  }
  .vis-item:hover { background: var(--cell-active); }
  .vis-item:focus-visible { outline: 1px solid var(--accent); outline-offset: -1px; }
  .vis-item.selected { background: var(--accent-soft); }
  .vis-item.selected .vis-item-name { color: var(--accent); }
  .vis-item.selected .vis-type-icon { color: var(--accent-mid); }

  .vis-arrow {
    width: 14px;
    height: 14px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--text-faint);
    cursor: default;
    padding: 0;
    border-radius: 3px;
    transition: color 0.1s, background 0.1s;
  }
  .vis-arrow--active { cursor: pointer; }
  .vis-arrow--active:hover { background: var(--border-soft); color: var(--text-muted); }
  .vis-arrow svg {
    width: 8px; height: 8px;
    transition: transform 0.14s cubic-bezier(0.16, 1, 0.3, 1);
    transform: rotate(-90deg);
  }
  .vis-arrow--open svg { transform: rotate(0deg); }

  .vis-type-icon {
    width: 12px; height: 12px;
    flex-shrink: 0;
    color: var(--text-faint);
    transition: color 0.1s;
  }

  .vis-item-name {
    font-size: 12px;
    font-family: var(--font);
    color: var(--text-muted);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    flex: 1;
    min-width: 0;
    line-height: 1;
    transition: color 0.1s;
  }

  /* Inline rename input */
  .vis-rename-input {
    flex: 1;
    min-width: 0;
    height: 18px;
    padding: 0 4px;
    border: 1px solid var(--accent);
    border-radius: 3px;
    background: var(--surface);
    color: var(--text);
    font-family: var(--font);
    font-size: 12px;
    outline: none;
  }
  .vis-rename-input.error {
    border-color: var(--error);
  }

  .vis-rename-error {
    flex-shrink: 0;
    max-width: 84px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--error);
    font-family: var(--font);
    font-size: 10.5px;
  }

  /* Tree item more-options button */
  .vis-item-menu-btn {
    width: 18px;
    height: 18px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--text-faint);
    border-radius: 3px;
    cursor: pointer;
    padding: 0;
    opacity: 0;
    transition: opacity 0.1s, background 0.1s, color 0.1s;
  }
  .vis-item:hover .vis-item-menu-btn,
  .vis-item.selected .vis-item-menu-btn,
  .vis-media-item:hover .vis-item-menu-btn { opacity: 1; }
  .vis-item-menu-btn:hover { background: var(--border-soft); color: var(--text-muted); }
  .vis-item.selected .vis-item-menu-btn:hover { background: color-mix(in srgb, var(--accent) 15%, transparent); color: var(--accent); }
  .vis-item-menu-btn svg { width: 12px; height: 12px; }

  /* Drag-and-drop states */
  .vis-item[draggable="true"] { cursor: grab; }
  .vis-item[draggable="true"]:active { cursor: grabbing; }

  .vis-item.drag-source { opacity: 0.38; }

  .vis-item.drop-before::before,
  .vis-item.drop-after::after {
    content: '';
    position: absolute;
    left: 8px; right: 8px;
    height: 2px;
    background: var(--accent);
    border-radius: 1px;
    pointer-events: none;
    z-index: 2;
  }
  .vis-item.drop-before::before { top: 0; }
  .vis-item.drop-after::after  { bottom: 0; }

  .vis-item.drop-into {
    background: var(--accent-soft) !important;
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--accent) 35%, transparent);
  }

  /* ── Add component suggestions ──────────────────────────────────── */
  .vis-add-suggestions {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    border-top: 1px solid var(--border-soft);
    background: var(--surface);
    overflow: hidden;
  }

  .vis-add-suggestion {
    display: flex;
    align-items: center;
    width: 100%;
    height: 26px;
    padding: 0 12px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 12px;
    text-align: left;
    cursor: pointer;
    transition: background 0.08s, color 0.08s;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .vis-add-suggestion:hover,
  .vis-add-suggestion.highlighted {
    background: var(--accent-soft);
    color: var(--accent);
  }
  .vis-add-suggestion-icon {
    width: 12px;
    height: 12px;
    flex-shrink: 0;
    margin-right: 7px;
    color: var(--text-faint);
    transition: color 0.08s;
  }
  .vis-add-suggestion:hover .vis-add-suggestion-icon,
  .vis-add-suggestion.highlighted .vis-add-suggestion-icon {
    color: var(--accent);
  }

  /* ── Add component input ─────────────────────────────────────────── */
  .vis-add-bar {
    flex-shrink: 0;
    padding: 6px 8px;
    border-top: 1px solid var(--border-soft);
    background: var(--surface);
  }

  .vis-add-input {
    width: 100%;
    height: 26px;
    padding: 0 8px;
    border: 1px solid var(--accent);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text);
    font-family: var(--font);
    font-size: 12px;
    outline: none;
    box-sizing: border-box;
    transition: border-color 0.1s;
  }
  .vis-add-input::placeholder { color: var(--text-faint); }
  .vis-add-input.error {
    border-color: var(--error);
  }
  .vis-add-error {
    margin-top: 4px;
    color: var(--error);
    font-family: var(--font);
    font-size: 11px;
  }

  /* ── Media tab ────────────────────────────────────────────────────── */
  .vis-media-tab {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: var(--surface);
    overflow: hidden;
  }

  .vis-media-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    min-height: 32px;
    padding: 5px 8px;
    border-bottom: 1px solid var(--border-soft);
    background: var(--bg);
    box-sizing: border-box;
  }

  .vis-media-header span {
    min-width: 0;
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
  }

  .vis-media-upload {
    position: relative;
    flex-shrink: 0;
    height: 24px;
    padding: 0 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--surface);
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 11.5px;
    cursor: pointer;
    overflow: hidden;
    box-sizing: border-box;
    transition: background 0.1s, border-color 0.1s, color 0.1s;
  }

  .vis-media-upload:hover {
    border-color: var(--text-faint);
    background: var(--cell-active);
    color: var(--text);
  }

  .vis-media-list {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    padding: 4px 0;
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
  }
  .vis-media-list::-webkit-scrollbar { width: 4px; }
  .vis-media-list::-webkit-scrollbar-track { background: transparent; }
  .vis-media-list::-webkit-scrollbar-thumb { background: var(--border); border-radius: 99px; }

  .vis-media-item {
    display: flex;
    align-items: center;
    gap: 6px;
    min-height: 26px;
    padding: 3px 8px;
    color: var(--text-muted);
    cursor: default;
    outline: none;
    box-sizing: border-box;
  }

  .vis-media-item:hover,
  .vis-media-item:focus-visible {
    background: var(--cell-active);
    color: var(--text);
  }

  .vis-media-icon {
    width: 13px;
    height: 13px;
    flex-shrink: 0;
    color: var(--text-faint);
  }

  .vis-media-name {
    min-width: 0;
    flex: 1;
    color: inherit;
    font-family: var(--font);
    font-size: 12px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .vis-media-empty {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-faint);
    font-family: var(--font);
    font-size: 11.5px;
  }

  .vis-media-size {
    flex-shrink: 0;
    font-family: var(--mono);
    font-size: 10px;
    color: var(--text-faint);
    margin-left: auto;
    padding-left: 6px;
  }

  .vis-media-rename {
    flex: 1;
    min-width: 0;
    display: grid;
    gap: 3px;
  }

  .vis-media-rename-input {
    min-width: 0;
    width: 100%;
    height: 22px;
    padding: 0 5px;
    border: 1px solid var(--accent);
    border-radius: 3px;
    background: var(--surface);
    color: var(--text);
    font-family: var(--font);
    font-size: 12px;
    outline: none;
    box-sizing: border-box;
  }

  .vis-media-rename-input.error {
    border-color: var(--error);
  }

  .vis-media-rename-error {
    min-width: 0;
    color: var(--error);
    font-family: var(--font);
    font-size: 10.5px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  /* ── Context menu ─────────────────────────────────────────────────── */
  .vis-ctx-backdrop {
    position: fixed;
    inset: 0;
    z-index: 49;
  }

  .vis-ctx-menu {
    position: fixed;
    z-index: 50;
    min-width: 160px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    padding: 4px;
    opacity: 0;
    transform: scale(0.96) translateY(-4px);
    transform-origin: top left;
    animation: visCtxIn 0.13s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }

  @keyframes visCtxIn {
    to { opacity: 1; transform: scale(1) translateY(0); }
  }

  .vis-ctx-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 5px 10px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 12.5px;
    border-radius: var(--radius);
    cursor: pointer;
    text-align: left;
    transition: background 0.08s, color 0.08s;
  }
  .vis-ctx-item:hover:not(:disabled) { background: var(--cell-active); color: var(--text); }
  .vis-ctx-item:disabled { opacity: 0.38; cursor: default; }
  .vis-ctx-item svg { width: 13px; height: 13px; flex-shrink: 0; }

  .vis-ctx-item--danger { color: var(--error); }
  .vis-ctx-item--danger:hover:not(:disabled) { background: var(--error-soft); color: var(--error); }

  .vis-ctx-sep {
    height: 1px;
    background: var(--border-soft);
    margin: 3px 0;
  }

  /* ── Properties panel ────────────────────────────────────────────── */
  .vis-props {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--surface);
    min-width: 0;
  }

  .vis-props-header {
    display: flex;
    align-items: center;
    gap: 7px;
    padding: 8px 14px;
    border-bottom: 1px solid var(--border-soft);
    flex-shrink: 0;
  }

  .vis-props-type-icon { width: 13px; height: 13px; color: var(--text-faint); flex-shrink: 0; }

  .vis-props-type {
    font-family: var(--font);
    font-size: 12px;
    color: var(--text-faint);
    white-space: nowrap;
  }

  .vis-props-dot {
    font-size: 12px;
    color: var(--text-faint);
    padding: 0 3px;
    flex-shrink: 0;
  }

  .vis-props-title {
    font-family: var(--font);
    font-size: 12.5px;
    font-weight: 500;
    color: var(--text);
    letter-spacing: -0.01em;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    min-width: 0;
  }

  .vis-props-scroll {
    flex: 1;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
    min-height: 0;
  }
  .vis-props-scroll::-webkit-scrollbar { width: 4px; }
  .vis-props-scroll::-webkit-scrollbar-track { background: transparent; }
  .vis-props-scroll::-webkit-scrollbar-thumb { background: var(--border); border-radius: 99px; }

  /* ── Sections ────────────────────────────────────────────────────── */
  .vis-section { border-bottom: 1px solid var(--border-soft); }

  .vis-section-hd {
    display: flex; align-items: center; gap: 6px;
    width: 100%; height: 28px; padding: 0 12px;
    border: none; background: var(--bg);
    cursor: pointer; text-align: left;
    transition: background 0.1s;
  }
  .vis-section-hd:hover { background: var(--cell-active); }
  .vis-section-hd:focus-visible { outline: 1px solid var(--accent); outline-offset: -1px; }

  .vis-section-caret {
    width: 8px; height: 8px; flex-shrink: 0;
    color: var(--text-faint);
    transition: transform 0.14s cubic-bezier(0.16, 1, 0.3, 1);
    transform: rotate(-90deg);
  }
  .vis-section-caret--open { transform: rotate(0deg); }

  .vis-section-title {
    font-family: var(--font); font-size: 10px; font-weight: 500;
    text-transform: uppercase; letter-spacing: 0.07em; color: var(--text-faint);
  }

  .vis-section-body { background: var(--surface); }

  /* ── Property rows ───────────────────────────────────────────────── */
  .vis-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    align-items: center;
    padding: 5px 12px;
    min-height: 30px;
    border-bottom: 1px solid var(--border-soft);
  }
  .vis-row:last-child { border-bottom: none; }

  .vis-row-label { display: flex; align-items: center; gap: 4px; min-width: 0; }

  .vis-prop-name {
    font-family: var(--font); font-size: 11.5px; color: var(--text-muted);
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  }

  .vis-help-btn {
    width: 15px; height: 15px; flex-shrink: 0;
    border: none; background: transparent; color: var(--text-faint);
    cursor: pointer; padding: 0;
    display: flex; align-items: center; justify-content: center;
    border-radius: 50%; transition: color 0.1s, background 0.1s;
  }
  .vis-help-btn:hover { background: var(--border-soft); color: var(--text-muted); }
  .vis-help-btn svg { width: 13px; height: 13px; }

  .vis-row-editor { min-width: 0; }

  .vis-help-callout {
    position: fixed;
    z-index: 60;
    width: 220px;
    padding: 8px 10px;
    background: var(--surface);
    border: 1.5px solid var(--border);
    border-radius: 10px;
    box-shadow: var(--shadow-md);
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 12px;
    line-height: 1.5;
    pointer-events: none;
    opacity: 0;
    transform: translateY(3px);
    animation: visHelpCalloutIn 0.14s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }

  @keyframes visHelpCalloutIn {
    to { opacity: 1; transform: translateY(0); }
  }

  /* Arrow pointing downward toward the button */
  .vis-help-callout::after {
    content: '';
    position: absolute;
    bottom: -5px;
    left: var(--arrow-left, 50%);
    width: 8px;
    height: 8px;
    background: var(--surface);
    border-bottom: 1.5px solid var(--border);
    border-right: 1.5px solid var(--border);
    transform: translateX(-50%) rotate(45deg);
    border-radius: 0 0 2px 0;
  }

  /* Color */
  .vis-color-control {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 5px;
    align-items: center;
  }

  .vis-color-pill {
    position: relative;
    display: flex; align-items: center; gap: 6px;
    height: 24px; padding: 0 7px;
    border: 1px solid var(--border); border-radius: var(--radius);
    background: var(--bg); cursor: pointer;
    transition: border-color 0.1s, background 0.1s; overflow: hidden;
  }
  .vis-color-pill:hover { border-color: var(--text-faint); background: var(--cell-active); }
  .vis-color-native {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
  }
  .vis-color-dot {
    width: 11px; height: 11px; flex-shrink: 0;
    border-radius: 50%; border: 1px solid var(--border);
  }
  .vis-color-text {
    flex: 1; font-family: var(--font); font-size: 11.5px; color: var(--text-muted);
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis; min-width: 0;
  }
  .vis-pill-caret { width: 8px; height: 8px; flex-shrink: 0; color: var(--text-faint); }

  .vis-color-none {
    height: 24px;
    padding: 0 7px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 11px;
    cursor: pointer;
  }
  .vis-color-none:hover {
    border-color: var(--text-faint);
    background: var(--cell-active);
    color: var(--text);
  }

  /* Checkbox */
  .vis-check-wrap { display: flex; align-items: center; cursor: pointer; }
  .vis-check-native { display: none; }
  .vis-check-box {
    width: 14px; height: 14px;
    border: 1.5px solid var(--border); border-radius: 3px;
    background: var(--bg); flex-shrink: 0; position: relative;
    transition: border-color 0.1s, background 0.1s;
  }
  .vis-check-wrap:hover .vis-check-box { border-color: var(--text-faint); }
  .vis-check-native:checked + .vis-check-box { background: var(--accent); border-color: var(--accent); }
  .vis-check-native:checked + .vis-check-box::after {
    content: '';
    position: absolute; inset: 1.5px;
    background: url("data:image/svg+xml,%3Csvg viewBox='0 0 10 10' fill='none' stroke='white' stroke-width='2.2' stroke-linecap='round' stroke-linejoin='round' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1.5 5l2.5 2.5 4.5-4'/%3E%3C/svg%3E") center / contain no-repeat;
  }

  /* Inputs */
  .vis-input {
    width: 100%; height: 24px; padding: 0 7px;
    border: 1px solid var(--border); border-radius: var(--radius);
    background: var(--bg); color: var(--text);
    font-family: var(--font); font-size: 11.5px;
    outline: none; box-sizing: border-box; transition: border-color 0.1s;
  }
  .vis-input:focus { border-color: var(--accent); }
  input[type='number'].vis-input { -moz-appearance: textfield; }
  input[type='number'].vis-input::-webkit-inner-spin-button,
  input[type='number'].vis-input::-webkit-outer-spin-button { -webkit-appearance: none; }

  .vis-layout-edit {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 54px;
    gap: 5px;
    align-items: center;
  }

  .vis-layout-edit .vis-select-btn:only-child {
    grid-column: 1 / -1;
  }

  .vis-layout-pixels {
    text-align: right;
  }

  .vis-textarea {
    width: 100%; min-height: 52px; padding: 5px 7px;
    border: 1px solid var(--border); border-radius: var(--radius);
    background: var(--bg); color: var(--text);
    font-family: var(--font); font-size: 11.5px;
    outline: none; resize: vertical; box-sizing: border-box; transition: border-color 0.1s;
  }
  .vis-textarea:focus { border-color: var(--accent); }

  /* Assets */
  .vis-asset-control {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 26px;
    gap: 5px;
    align-items: center;
  }

  .vis-asset-upload {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 24px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text-muted);
    cursor: pointer;
    box-sizing: border-box;
    transition: border-color 0.1s, background 0.1s, color 0.1s;
  }

  .vis-asset-upload:hover {
    border-color: var(--text-faint);
    background: var(--cell-active);
    color: var(--text);
  }

  .vis-asset-upload svg {
    width: 12px;
    height: 12px;
  }

  .vis-file-native {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
  }

  /* Custom select trigger */
  .vis-select-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    width: 100%;
    height: 24px;
    padding: 0 7px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 11.5px;
    text-align: left;
    cursor: pointer;
    box-sizing: border-box;
    transition: border-color 0.1s, background 0.1s;
  }
  .vis-select-btn:hover { background: var(--cell-active); border-color: var(--text-faint); }
  .vis-select-btn.open { border-color: var(--accent); }
  .vis-select-btn-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .vis-select-btn-chevron {
    width: 8px;
    height: 8px;
    flex-shrink: 0;
    color: var(--text-faint);
    transition: transform 0.15s;
  }
  .vis-select-btn-chevron.rotated { transform: rotate(180deg); }

  /* Active option in dropdown */
  .vis-option-active { color: var(--accent) !important; font-weight: 500; }

  /* Empty */
  .vis-empty {
    flex: 1; display: flex; flex-direction: column;
    align-items: center; justify-content: center;
    gap: 10px; color: var(--text-faint);
    font-size: 12px; font-family: var(--font); padding: 24px; text-align: center;
  }
  .vis-empty svg { width: 28px; height: 28px; opacity: 0.6; }

  /* ── ListData property row control ──────────────────────────────────── */
  .vis-listdata-control {
    display: flex;
    align-items: center;
    gap: 5px;
    min-width: 0;
  }

  .vis-listdata-summary {
    flex: 1;
    min-width: 0;
    font-family: var(--font);
    font-size: 11.5px;
    color: var(--text-faint);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .vis-listdata-edit-btn {
    flex-shrink: 0;
    height: 22px;
    padding: 0 8px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 11px;
    cursor: pointer;
    transition: border-color 0.1s, background 0.1s, color 0.1s;
  }
  .vis-listdata-edit-btn:hover {
    border-color: var(--accent);
    background: var(--accent-soft);
    color: var(--accent);
  }

  /* ── ListData dialog ─────────────────────────────────────────────────── */
  .vis-listdata-backdrop {
    position: fixed;
    inset: 0;
    z-index: 90;
    border: 0;
    padding: 0;
    background: rgba(0, 0, 0, 0.45);
    cursor: default;
  }

  .vis-listdata-dialog {
    position: fixed;
    z-index: 91;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: min(600px, calc(100vw - 32px));
    max-height: min(520px, calc(100vh - 64px));
    display: flex;
    flex-direction: column;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: 0 8px 40px rgba(0, 0, 0, 0.32);
    overflow: hidden;
    animation: visCtxIn 0.14s cubic-bezier(0.16, 1, 0.3, 1) forwards;
  }

  .vis-listdata-hd {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 9px 12px;
    border-bottom: 1px solid var(--border-soft);
    background: var(--bg);
  }

  .vis-listdata-title {
    flex: 1;
    font-family: var(--font);
    font-size: 12.5px;
    font-weight: 500;
    color: var(--text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .vis-listdata-close {
    width: 20px;
    height: 20px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--text-faint);
    border-radius: 3px;
    cursor: pointer;
    padding: 0;
    transition: background 0.1s, color 0.1s;
  }
  .vis-listdata-close:hover { background: var(--border-soft); color: var(--text-muted); }
  .vis-listdata-close svg { width: 11px; height: 11px; }

  .vis-listdata-table-wrap {
    flex: 1;
    min-height: 0;
    overflow: auto;
    scrollbar-width: thin;
    scrollbar-color: var(--border) transparent;
  }
  .vis-listdata-table-wrap::-webkit-scrollbar { width: 4px; height: 4px; }
  .vis-listdata-table-wrap::-webkit-scrollbar-track { background: transparent; }
  .vis-listdata-table-wrap::-webkit-scrollbar-thumb { background: var(--border); border-radius: 99px; }

  .vis-listdata-table {
    width: 100%;
    border-collapse: collapse;
    font-family: var(--font);
    font-size: 12px;
    table-layout: fixed;
  }

  .vis-listdata-th {
    position: sticky;
    top: 0;
    background: var(--bg);
    z-index: 1;
    padding: 5px 8px;
    text-align: left;
    font-size: 10px;
    font-weight: 500;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--text-faint);
    border-bottom: 1px solid var(--border-soft);
    white-space: nowrap;
  }

  .vis-listdata-th-order { width: 44px; }
  .vis-listdata-th-del { width: 34px; }

  .vis-listdata-tr:hover { background: var(--cell-active); }

  .vis-listdata-td {
    padding: 4px 6px;
    border-bottom: 1px solid var(--border-soft);
    vertical-align: middle;
  }

  .vis-listdata-td-order {
    width: 44px;
    padding: 3px 4px;
    text-align: center;
  }

  .vis-listdata-td-del {
    width: 34px;
    text-align: center;
    padding: 3px 4px;
  }

  .vis-listdata-reorder {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1px;
  }

  .vis-listdata-reorder-btn {
    width: 20px;
    height: 17px;
    display: flex;
    align-items: center;
    justify-content: center;
    border: none;
    background: transparent;
    color: var(--text-faint);
    cursor: pointer;
    padding: 0;
    border-radius: 2px;
    transition: background 0.08s, color 0.08s;
  }
  .vis-listdata-reorder-btn:hover:not(:disabled) { background: var(--border-soft); color: var(--text-muted); }
  .vis-listdata-reorder-btn:disabled { opacity: 0.22; cursor: default; }
  .vis-listdata-reorder-btn svg { width: 8px; height: 8px; }

  .vis-listdata-del-btn {
    width: 22px;
    height: 22px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: auto;
    border: none;
    background: transparent;
    color: var(--text-faint);
    cursor: pointer;
    padding: 0;
    border-radius: 3px;
    transition: background 0.08s, color 0.08s;
  }
  .vis-listdata-del-btn:hover {
    background: color-mix(in srgb, var(--error) 12%, transparent);
    color: var(--error);
  }
  .vis-listdata-del-btn svg { width: 12px; height: 12px; }

  .vis-listdata-input {
    width: 100%;
    height: 24px;
    padding: 0 6px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text);
    font-family: var(--font);
    font-size: 11.5px;
    outline: none;
    box-sizing: border-box;
    transition: border-color 0.1s;
  }
  .vis-listdata-input:focus { border-color: var(--accent); }
  .vis-listdata-input::placeholder { color: var(--text-faint); }

  .vis-listdata-select {
    width: 100%;
    height: 24px;
    padding: 0 5px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text);
    font-family: var(--font);
    font-size: 11.5px;
    outline: none;
    cursor: pointer;
    box-sizing: border-box;
    transition: border-color 0.1s;
  }
  .vis-listdata-select:focus { border-color: var(--accent); }

  .vis-listdata-empty-rows {
    padding: 28px 16px;
    text-align: center;
    color: var(--text-faint);
    font-family: var(--font);
    font-size: 12px;
  }

  .vis-listdata-footer {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    padding: 7px 10px;
    gap: 8px;
    border-top: 1px solid var(--border-soft);
    background: var(--bg);
  }

  .vis-listdata-footer-actions {
    margin-left: auto;
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .vis-listdata-add-btn {
    display: flex;
    align-items: center;
    gap: 5px;
    height: 26px;
    padding: 0 10px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 12px;
    cursor: pointer;
    transition: border-color 0.1s, background 0.1s, color 0.1s;
  }
  .vis-listdata-add-btn:hover {
    border-color: var(--accent);
    background: var(--accent-soft);
    color: var(--accent);
  }
  .vis-listdata-add-btn svg { width: 10px; height: 10px; }

  .vis-listdata-cancel-btn {
    height: 26px;
    padding: 0 12px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 12px;
    cursor: pointer;
    transition: border-color 0.1s, background 0.1s;
  }
  .vis-listdata-cancel-btn:hover { border-color: var(--text-faint); background: var(--cell-active); }

  .vis-listdata-save-btn {
    height: 26px;
    padding: 0 14px;
    border: 1px solid var(--accent);
    border-radius: var(--radius);
    background: var(--accent);
    color: #fff;
    font-family: var(--font);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: opacity 0.1s;
  }
  .vis-listdata-save-btn:hover { opacity: 0.86; }
</style>
