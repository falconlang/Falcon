<script>
  import { createEventDispatcher, tick } from 'svelte';
  import simpleComponents from '../../../lang/code/compdb/simple_components.json';
  import { addDesignAsset, deleteDesignAsset, designAssets, renameDesignAsset } from './stores.js';

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

  function buildPropsForComponent(component) {
    const designerByName = new Map((component.properties || []).map(prop => [prop.name, prop]));
    const blockByName = new Map((component.blockProperties || []).map(prop => [prop.name, prop]));
    const orderedNames = [];

    for (const prop of component.blockProperties || []) {
      if (prop.rw === 'invisible') continue;
      orderedNames.push(prop.name);
    }

    for (const prop of component.properties || []) {
      const blockProp = blockByName.get(prop.name);
      if (blockProp?.rw === 'invisible') continue;
      if (!orderedNames.includes(prop.name)) orderedNames.push(prop.name);
    }

    return orderedNames.map(name => {
      const designerProp = designerByName.get(name);
      const blockProp = blockByName.get(name);
      const options = optionsFromHelper(blockProp);
      const editorType = inferEditorType(name, blockProp, designerProp);
      const editorArgs = designerProp?.editorArgs || [];

      return {
        name,
        editorType,
        category: normalizePropertyCategory(blockProp?.category),
        defaultValue: designerProp?.defaultValue ?? '',
        options: options.length
          ? options
          : editorType === 'choices'
            ? editorArgs.map(arg => ({ value: arg, label: arg }))
            : ENUM_OPTIONS[editorType] || [],
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

  const COMP_PROPS = buildComponentProps(simpleComponents);
  const KNOWN_COMPONENT_TYPES = new Set(Object.keys(COMP_PROPS));
  const COMPONENT_META = new Map(simpleComponents.map(component => [component.name, component]));

  function componentTypeName(ident) {
    const dotIdx = ident.indexOf('.');
    return dotIdx === -1 ? ident : ident.slice(0, dotIdx);
  }

  function isKnownComponentIdent(ident) {
    return KNOWN_COMPONENT_TYPES.has(componentTypeName(ident));
  }

  const FALLBACK_PROPS = [
    { name: 'Visible', editorType: 'visibility',  category: 'Appearance' },
    { name: 'Height',  editorType: 'layout_size', category: 'Appearance' },
    { name: 'Width',   editorType: 'layout_size', category: 'Appearance' },
  ];


  const CONTAINER_TYPES = new Set([
    'Screen',
    'Form',
    'Canvas',
    'Map',
    'FeatureCollection',
    'ScrollHorizontal',
    'ScrollVertical',
    ...simpleComponents
      .filter(component => component.categoryString === 'LAYOUT')
      .map(component => component.name),
  ]);

  function isNonVisibleComponentType(type) {
    if (type === 'Screen' || type === 'Form') return false;
    return COMPONENT_META.get(type)?.nonVisible === 'true';
  }

  // ── Schema parser ──────────────────────────────────────────────────
  function parseSchema(text) {
    if (!text?.trim()) return { root: null, error: 'The design schema is empty.' };
    text = text.replace(/\/\/[^\n]*/g, '');
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
      if (!name) {
        typeCounts[type] = (typeCounts[type] || 0) + 1;
        name = `${type}${typeCounts[type]}`;
      }
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
      return { root, error: null };
    } catch (err) {
      return { root: null, error: err.message || 'Unable to parse design schema.' };
    }
  }

  // ── Serializer ─────────────────────────────────────────────────────
  function needsQuotes(val) {
    return !/^(true|false|True|False|-?\d+\.?\d*|&H[0-9A-Fa-f]{8})$/.test(String(val).trim());
  }

  function serializeComp(node, depth = 0) {
    const ind = '  '.repeat(depth);
    const ident = `${node.type}.${node.name}`;
    const propLines = Object.entries(node.props || {}).map(
      ([k, v]) => `${ind}  ${k}: ${needsQuotes(v) ? `"${v}"` : v}`
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
      const parsed = parseSchema(schemaValue);
      parseError = parsed.error || '';
      mutableTree = parsed.root;
      if (parsed.root && !findNode(parsed.root, selectedPathId)) selectedPathId = '0';
    }
  }

  function applyChange(newTree) {
    mutableTree = newTree;
    const schema = serializeTree(newTree);
    prevSchema = schema;
    parseError = '';
    dispatch('change', { schema });
  }

  // ── Display state ──────────────────────────────────────────────────
  let collapsed = new Set();
  let selectedPathId = '0';
  let collapsedCategories = new Set();
  let componentFilter = 'all';
  let showPicker = false;
  let addCompValue = '';
  let addCompError = '';
  let addCompInputEl;
  let ctxMenu = null; // { x, y, pathId, isContainer, isRoot }
  let assetCtxMenu = null; // { x, y, assetId }
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
  $: visibleFlatList = flatList.filter(componentFilterMatches);
  $: mediaAssets = $designAssets.map(normalizeAssetRecord);
  $: activeAsset = assetCtxMenu ? mediaAssets.find(asset => asset.id === assetCtxMenu.assetId) : null;
  $: if (mutableTree && visibleFlatList.length && !visibleFlatList.some(n => n.pathId === selectedPathId)) {
    selectedPathId = visibleFlatList[0].pathId;
  }
  $: selectedNode = flatList.find(n => n.pathId === selectedPathId) ?? visibleFlatList[0] ?? flatList[0] ?? null;
  $: propGroups = selectedNode
    ? groupByCategory(COMP_PROPS[selectedNode.type] ?? FALLBACK_PROPS)
    : [];
  $: isRoot = selectedNode?.pathId === '0';

  function componentFilterMatches(node) {
    if (componentFilter === 'all') return true;
    const isNonVisible = isNonVisibleComponentType(node.type);
    return componentFilter === 'nonVisible' ? isNonVisible : !isNonVisible;
  }

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

  // ── Tree operations ────────────────────────────────────────────────
  function addComponent(type) {
    if (!mutableTree) return;
    const newTree = cloneTree(mutableTree);
    const newName = uniqueName(type, newTree);
    const newComp = { type, name: newName, props: {}, children: [], pathId: '' };

    const sel = findNode(newTree, selectedPathId);
    if (sel && CONTAINER_TYPES.has(sel.type)) {
      sel.children.push(newComp);
    } else if (sel && selectedPathId !== '0') {
      const p = findParentOf(newTree, selectedPathId);
      if (p) p.parent.children.splice(p.index + 1, 0, newComp);
      else newTree.children.push(newComp);
    } else {
      newTree.children.push(newComp);
    }

    rebuildPathIds(newTree);
    applyChange(newTree);

    const added = flattenTree(newTree, 0, new Set()).find(n => n.name === newName);
    if (added) selectedPathId = added.pathId;
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
    const targetIsContainer = CONTAINER_TYPES.has(targetNode?.type);

    let position;
    if (pathId === '0') {
      position = 'into';
    } else if (targetIsContainer && relY > 0.25 && relY < 0.75) {
      position = 'into';
    } else {
      position = relY < 0.5 ? 'before' : 'after';
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
      targetNode.children.push(draggedNode);
    } else if (targetParentNode) {
      const idx = targetParentNode.children.indexOf(targetNode);
      targetParentNode.children.splice(idx + (position === 'after' ? 1 : 0), 0, draggedNode);
    } else {
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

  function ctxRename()  { closeCtxMenu(); startRename(); }
  function ctxDelete()  { closeCtxMenu(); deleteSelected(); }
  function ctxAddComp() {
    closeCtxMenu();
    addCompValue = '';
    addCompError = '';
    showPicker = true;
    tick().then(() => addCompInputEl?.focus());
  }

  function commitAddComp() {
    const type = addCompValue.trim();
    if (!type) {
      addCompError = 'Enter a component name.';
      tick().then(() => addCompInputEl?.focus());
      return;
    }
    if (!KNOWN_COMPONENT_TYPES.has(type)) {
      addCompError = 'Unknown component type.';
      tick().then(() => addCompInputEl?.focus());
      return;
    }
    addComponent(type);
    showPicker = false;
    addCompValue = '';
    addCompError = '';
  }

  function handleAddCompKey(e) {
    if (e.key === 'Enter') commitAddComp();
    if (e.key === 'Escape') { showPicker = false; addCompValue = ''; addCompError = ''; }
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
    const clean = renameValue.replace(/\s+/g, '_').replace(/[^\w]/g, '');
    const current = findNode(mutableTree, renamingPathId)?.name;
    if (!clean) {
      renameError = 'Name is required.';
      tick().then(() => renameInputEl?.focus());
      return;
    }
    if (clean !== current && nameExists(clean, mutableTree, targetPathId)) {
      renameError = 'Name already exists.';
      tick().then(() => renameInputEl?.focus());
      return;
    }
    renamingPathId = null;
    renameError = '';
    if (clean === current) return;
    const newTree = cloneTree(mutableTree);
    const target = findNode(newTree, targetPathId);
    if (target) target.name = clean;
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
    applyChange(newTree);
    selectedPathId = pathId;
  }

  function aiColorToHex(val) {
    const m = String(val || '').match(/^&H([0-9A-Fa-f]{8})$/);
    if (!m || parseInt(m[1].slice(0, 2), 16) === 0) return '#1a1916';
    const h = m[1];
    const r = h.slice(6, 8);
    const g = h.slice(4, 6);
    const b = h.slice(2, 4);
    return `#${r}${g}${b}`.toLowerCase();
  }

  function hexToAiColor(hex) {
    const m = String(hex || '').match(/^#?([0-9A-Fa-f]{6})$/);
    if (!m) return '&H00000000';
    const h = m[1].toUpperCase();
    const r = h.slice(0, 2);
    const g = h.slice(2, 4);
    const b = h.slice(4, 6);
    return `&HFF${b}${g}${r}`;
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
    return String(name || '').trim().replace(/[\\/]+/g, '-');
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

  // ── Component icons (12×12 viewBox inner markup) ───────────────────
  function compIcon(type) {
    switch (type) {
      case 'Screen':              return '<rect x="2.5" y="0.5" width="7" height="11" rx="1.5" stroke-width="1.3"/><line x1="4" y1="9.5" x2="8" y2="9.5" stroke-width="1.3" stroke-linecap="round"/>';
      case 'Button':              return '<rect x="0.5" y="3.5" width="11" height="5" rx="1.5" stroke-width="1.3"/>';
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
      default:                    return '<rect x="1.5" y="1.5" width="9" height="9" rx="1.5" stroke-width="1.3"/>';
    }
  }
</script>

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
    <div class="vis-tree-filter" aria-label="Component list filter">
      <button
        type="button"
        class:active={componentFilter === 'all'}
        on:click={() => { componentFilter = 'all'; closeCtxMenu(); }}
      >
        All
      </button>
      <button
        type="button"
        class:active={componentFilter === 'visible'}
        on:click={() => { componentFilter = 'visible'; closeCtxMenu(); }}
      >
        Visible
      </button>
      <button
        type="button"
        class:active={componentFilter === 'nonVisible'}
        on:click={() => { componentFilter = 'nonVisible'; closeCtxMenu(); }}
      >
        Non-visible
      </button>
    </div>

    <div
      class="vis-tree-scroll"
      role="tree"
      aria-label="Components"
      tabindex="-1"
      on:dragleave={handleScrollDragLeave}
    >
      {#if visibleFlatList.length}
        {#each visibleFlatList as node (node.pathId)}
          {@const hasKids  = node.children.length > 0}
          {@const isOpen   = !collapsed.has(node.pathId)}
          {@const isSel    = selectedNode?.pathId === node.pathId}
          {@const renaming = renamingPathId === node.pathId}
          {@const dropPos  = dropTarget?.pathId === node.pathId ? dropTarget.position : null}

          <div
            class="vis-item"
            class:selected={isSel}
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

          </div>
        {/each}
      {:else}
        <div class="vis-tree-empty">No components match this filter.</div>
      {/if}
    </div>

    <!-- Add component input -->
    {#if showPicker}
      <div class="vis-add-bar">
        <input
          class="vis-add-input"
          bind:this={addCompInputEl}
          bind:value={addCompValue}
          placeholder="Component name, e.g. Button"
          class:error={!!addCompError}
          on:input={() => { addCompError = ''; }}
          on:keydown={handleAddCompKey}
          on:blur={() => { showPicker = false; addCompValue = ''; addCompError = ''; }}
          spellcheck="false"
        />
        {#if addCompError}
          <div class="vis-add-error">{addCompError}</div>
        {/if}
      </div>
    {/if}

    <div class="vis-media">
      <div class="vis-media-header">
        <span>Media</span>
        <label class="vis-media-upload">
          Upload File ...
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
              {/if}
            </div>
          {/each}
        {:else}
          <div class="vis-media-empty">No media files</div>
        {/if}
      </div>
    </div>

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
                  <div class="vis-row">
                    <div class="vis-row-label">
                      <span class="vis-prop-name">{prop.name}</span>
                      <button class="vis-help-btn" title="{prop.name} ({prop.editorType})" aria-label="Help">
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
                        <div class="vis-layout-edit">
                          <select
                            class="vis-select"
                            on:change={e => updateLayoutMode(selectedNode.pathId, prop.name, e.currentTarget.value, val)}
                          >
                            <option value="automatic" selected={mode === 'automatic'}>Automatic</option>
                            <option value="fill" selected={mode === 'fill'}>Fill parent</option>
                            <option value="custom" selected={mode === 'custom'}>Pixels</option>
                          </select>
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
                        <div class="vis-asset-control">
                          <select
                            class="vis-select"
                            on:change={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.value)}
                          >
                            <option value="" selected={!editorVal}>None...</option>
                            {#each assetOptions(editorVal) as asset}
                              <option value={asset} selected={editorVal === asset}>{asset}</option>
                            {/each}
                          </select>
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
                        <select
                          class="vis-select"
                          on:change={e => updateProp(selectedNode.pathId, prop.name, e.currentTarget.value)}
                        >
                          {#each options as option}
                            <option value={option.value} selected={selectValue(prop, val) === String(option.value)}>
                              {option.label}
                            </option>
                          {/each}
                        </select>
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
    font-family: var(--font-mono);
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
  }

  .vis-parse-btn:hover {
    border-color: var(--text-faint);
    background: var(--cell-active);
  }

  /* ── Tree panel ─────────────────────────────────────────────────────── */
  .vis-tree {
    flex: 1;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border);
    background: var(--bg);
    overflow: hidden;
    position: relative;
  }

  .vis-tree-filter {
    flex-shrink: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 4px;
    padding: 6px;
    border-bottom: 1px solid var(--border-soft);
    background: var(--surface);
  }

  .vis-tree-filter button {
    height: 24px;
    padding: 0 6px;
    border: 1px solid var(--border);
    border-radius: var(--radius);
    background: var(--bg);
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 11px;
    cursor: pointer;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    transition: background 0.1s, border-color 0.1s, color 0.1s;
  }

  .vis-tree-filter button:hover {
    border-color: var(--text-faint);
    background: var(--cell-active);
    color: var(--text);
  }

  .vis-tree-filter button.active {
    border-color: var(--accent);
    background: var(--accent-soft);
    color: var(--accent);
  }

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
    padding: 10px 12px;
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 12px;
    line-height: 1.35;
  }

  /* Tree items */
  .vis-item {
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

  /* Drag-and-drop states */
  .vis-item[draggable="true"] { cursor: grab; }
  .vis-item[draggable="true"]:active { cursor: grabbing; }

  .vis-item.drag-source { opacity: 0.38; }

  .vis-item { position: relative; }

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

  /* ── Media panel ─────────────────────────────────────────────────── */
  .vis-media {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    min-height: 96px;
    max-height: 176px;
    border-top: 1px solid var(--border);
    background: var(--surface);
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
    padding: 7px 8px;
    color: var(--text-faint);
    font-family: var(--font);
    font-size: 11.5px;
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

  .vis-props-title {
    font-family: var(--font);
    font-size: 12.5px;
    font-weight: 500;
    color: var(--text);
    letter-spacing: -0.01em;
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

  .vis-layout-edit .vis-select:only-child {
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

  /* Select */
  .vis-select {
    width: 100%; height: 24px; padding: 0 20px 0 7px;
    border: 1px solid var(--border); border-radius: var(--radius);
    background: var(--bg); color: var(--text-muted);
    font-family: var(--font); font-size: 11.5px;
    outline: none; cursor: pointer; appearance: none;
    background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 8 8' fill='none' stroke='%23B0AEA8' stroke-width='1.4' stroke-linecap='round' stroke-linejoin='round' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M1.5 3l2.5 2.5 2.5-2.5'/%3E%3C/svg%3E");
    background-repeat: no-repeat; background-position: right 6px center; background-size: 10px;
    box-sizing: border-box; transition: border-color 0.1s;
  }
  .vis-select:focus { border-color: var(--accent); }

  /* Empty */
  .vis-empty {
    flex: 1; display: flex; flex-direction: column;
    align-items: center; justify-content: center;
    gap: 10px; color: var(--text-faint);
    font-size: 12px; font-family: var(--font); padding: 24px; text-align: center;
  }
  .vis-empty svg { width: 28px; height: 28px; opacity: 0.6; }
</style>
