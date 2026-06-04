<script>
  import { createEventDispatcher, onDestroy, onMount, tick } from 'svelte';
  import 'leaflet/dist/leaflet.css';
  import {
    elementsFromString,
    isSimulationNonVisibleType,
    resolveAssetUrl,
  } from './simulation-capabilities.js';
  import {
    isSimulationEnabled,
    isSimulationVisible,
  } from './design-schema-tree.js';

  export let node;
  export let state = {};
  export let unsupported = [];
  export let assets = [];
  export let parentType = '';
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  const MONTHS = [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December',
  ];

  let pickerOpen = false;
  let pickerFilter = '';
  let listFilter = '';
  let actionTokens = {};
  let dateInput;
  let timeInput;
  let textInputEl;
  let buttonEl;
  let pickerWrapEl;
  let webViewerFrame;
  let longClickTimer = null;
  let suppressClick = false;

  $: props = state?.[node?.name] ?? {};
  $: visible = isSimulationVisible(state, node?.name);
  $: enabled = isSimulationEnabled(state, node?.name);
  $: nonVisible = isSimulationNonVisibleType(node?.type);
  $: unsupportedHere = unsupported.some(entry => entry.detail === `${node?.type}.${node?.name}`);
  $: rawElementItems = elementItems();
  $: elements = rawElementItems.map(itemMainText);
  $: assetUrl = resolveAssetUrl(assets, assetName());
  $: isMultilineTextBox = node?.type === 'TextBox' && boolValue(props.MultiLine, false);
  $: filteredPickerItems = filterIndexed(elements, pickerFilter);
  $: listRows = listViewRows();
  $: filteredListRows = listRows.filter(row => textIncludes(`${row.text1} ${row.text2}`, listFilter));
  $: handleComponentActions(actions?.[node?.name] ?? {});

  onDestroy(() => {
    clearLongClick();
    stopSpriteAnimationLoop();
    if (mapInstance) { mapInstance.remove(); mapInstance = null; }
  });

  function hasValue(value) {
    return value !== undefined && value !== null && value !== '';
  }

  function firstNonEmpty(...values) {
    return values.find(value => hasValue(value) && String(value).trim() !== '') ?? '';
  }

  function boolValue(value, fallback = false) {
    if (value === undefined || value === null || value === '') return fallback;
    if (typeof value === 'boolean') return value;
    if (typeof value === 'number') return value !== 0;
    const text = String(value).trim().toLowerCase();
    if (['true', '1', 'yes'].includes(text)) return true;
    if (['false', '0', 'no'].includes(text)) return false;
    return fallback;
  }

  function numberOr(value, fallback = 0) {
    if (value === undefined || value === null || value === '') return fallback;
    const numberValue = Number(value);
    return Number.isFinite(numberValue) ? numberValue : fallback;
  }

  function cssUrl(url) {
    return String(url ?? '').replace(/\\/g, '\\\\').replace(/"/g, '\\"');
  }

  function rgbaFromParts(r, g, b, alpha) {
    const a = Math.max(0, Math.min(1, alpha));
    const rounded = Math.round(a * 1000) / 1000;
    return `rgba(${r}, ${g}, ${b}, ${rounded})`;
  }

  function rgbaFromArgb(hex) {
    const alpha = parseInt(hex.slice(0, 2), 16) / 255;
    const r = parseInt(hex.slice(2, 4), 16);
    const g = parseInt(hex.slice(4, 6), 16);
    const b = parseInt(hex.slice(6, 8), 16);
    return rgbaFromParts(r, g, b, alpha);
  }

  function rgbaFromRgbaHex(hex) {
    const r = parseInt(hex.slice(0, 2), 16);
    const g = parseInt(hex.slice(2, 4), 16);
    const b = parseInt(hex.slice(4, 6), 16);
    const alpha = parseInt(hex.slice(6, 8), 16) / 255;
    return rgbaFromParts(r, g, b, alpha);
  }

  function colorValue(value, fallback = 'transparent') {
    if (value === undefined || value === null || value === '') return fallback;
    if (typeof value === 'number' && Number.isFinite(value)) {
      const unsigned = value >>> 0;
      // Component.COLOR_DEFAULT (&H00000000) means "use the theme default",
      // not transparent black, so fall back rather than rendering invisibly.
      if (unsigned === 0) return fallback;
      return rgbaFromArgb(unsigned.toString(16).padStart(8, '0'));
    }

    const text = String(value).trim();
    const ai = text.match(/^&H([0-9A-Fa-f]{8})$/);
    if (ai) {
      if (ai[1].toLowerCase() === '00000000') return fallback;
      return rgbaFromArgb(ai[1]);
    }

    const hex = text.match(/^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$/);
    if (hex?.[1]?.length === 8) return rgbaFromRgbaHex(hex[1]);
    if (hex) return text;

    if (/^(rgba?|hsla?|oklch|transparent|currentColor|inherit|var\()/i.test(text)) return text;
    return fallback;
  }

  function sizeStyle(prop) {
    const value = props[prop];
    const cssProp = prop === 'Width' ? 'width' : 'height';
    if (value === -2 || value === '-2') return `${cssProp}: 100%;`;
    if (value === -1 || value === '-1' || value === undefined || value === null || value === '') return '';
    const numberValue = Number(value);
    if (!Number.isFinite(numberValue)) return '';
    if (numberValue <= -1000) return `${cssProp}: ${Math.max(0, -numberValue - 1000)}%;`;
    if (numberValue >= 0) return `${cssProp}: ${numberValue}px;`;
    return '';
  }

  function positionStyle() {
    if (parentType !== 'AbsoluteArrangement') return '';
    return `position: absolute; left: ${numberOr(props.Left, 0)}px; top: ${numberOr(props.Top, 0)}px;`;
  }

  function typefaceStyleFor(value) {
    const raw = String(value ?? '').trim();
    if (!raw || raw === '0' || raw.toLowerCase() === 'default') return '';
    const lower = raw.toLowerCase();
    if (raw === '1' || lower === 'serif') return 'font-family: Georgia, "Times New Roman", serif;';
    if (raw === '2' || lower === 'sans' || lower === 'sans-serif') return 'font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif;';
    if (raw === '3' || lower === 'monospace' || lower === 'mono') return 'font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;';
    return `font-family: "${raw.replace(/"/g, '\\"')}", inherit;`;
  }

  function typefaceStyle() {
    return typefaceStyleFor(props.FontTypeface);
  }

  function textAlignStyle() {
    if (!hasValue(props.TextAlignment)) return '';
    const value = numberOr(props.TextAlignment, 0);
    if (value === 1) return 'text-align: center;';
    if (value === 2) return 'text-align: right;';
    return 'text-align: left;';
  }

  function isVerticalLayout(type = node?.type) {
    return ['Screen', 'Form', 'VerticalArrangement', 'VerticalScrollArrangement'].includes(type);
  }

  function isHorizontalLayout(type = node?.type) {
    return ['HorizontalArrangement', 'HorizontalScrollArrangement'].includes(type);
  }

  function alignmentStyle() {
    if (!isVerticalLayout() && !isHorizontalLayout()) return '';

    const horizontal = numberOr(props.AlignHorizontal, 1);
    const vertical = numberOr(props.AlignVertical, 1);
    const horizontalCss = horizontal === 2 ? 'flex-end' : horizontal === 3 ? 'center' : 'flex-start';
    const verticalCss = vertical === 2 ? 'center' : vertical === 3 ? 'flex-end' : 'flex-start';

    if (isHorizontalLayout()) {
      return `justify-content: ${horizontalCss}; align-items: ${verticalCss};`;
    }
    return `align-items: ${horizontalCss}; justify-content: ${verticalCss};`;
  }

  function isAutoSize(value) {
    return value === -1 || value === '-1' || value === undefined || value === null || value === '';
  }

  function emptyArrangementStyle() {
    if (node?.type === 'Screen' || node?.type === 'Form') return '';
    if (!isVerticalLayout() && !isHorizontalLayout()) return '';
    if ((node?.children?.length ?? 0) > 0) return '';
    // App Inventor renders an empty Automatic-sized arrangement at a 100px floor
    // (ComponentConstants.EMPTY_HV_ARRANGEMENT_WIDTH/HEIGHT), per dimension.
    const rules = [];
    if (isAutoSize(props.Width)) rules.push('min-width: 100px;');
    if (isAutoSize(props.Height)) rules.push('min-height: 100px;');
    return rules.join(' ');
  }

  function hintColorStyle() {
    return hasValue(props.HintColor) ? `--sim-placeholder-color: ${colorValue(props.HintColor, '#9aa0a6')};` : '';
  }

  function backgroundImageStyle() {
    if (!assetUrl || !firstNonEmpty(props.Image, props.BackgroundImage)) return '';
    return `background-image: url("${cssUrl(assetUrl)}"); background-repeat: no-repeat; background-position: center; background-size: cover;`;
  }

  function baseStyle(extra = '', options = {}) {
    const {
      size = true,
      position = true,
      typography = true,
      arrangement = true,
      backgroundImage = true,
    } = options;

    const rules = [
      size ? sizeStyle('Width') : '',
      size ? sizeStyle('Height') : '',
      hasValue(props.BackgroundColor) ? `background-color: ${colorValue(props.BackgroundColor)};` : '',
      hasValue(props.TextColor) ? `color: ${colorValue(props.TextColor, 'inherit')};` : '',
      typography && hasValue(props.FontSize) ? `font-size: ${numberOr(props.FontSize, 14)}px;` : '',
      typography && boolValue(props.FontBold, false) ? 'font-weight: 700;' : '',
      typography && boolValue(props.FontItalic, false) ? 'font-style: italic;' : '',
      typography ? typefaceStyle() : '',
      typography ? textAlignStyle() : '',
      arrangement ? alignmentStyle() : '',
      backgroundImage ? backgroundImageStyle() : '',
      position ? positionStyle() : '',
      extra,
    ];
    return rules.filter(Boolean).join(' ');
  }

  function containerStyle(extra = '') {
    return [sizeStyle('Width'), sizeStyle('Height'), positionStyle(), extra].filter(Boolean).join(' ');
  }

  function shapeStyle() {
    const shape = String(props.Shape ?? '0');
    if (shape === '1') return 'border-radius: 9px;';
    if (shape === '2') return 'border-radius: 0;';
    if (shape === '3') return 'border-radius: 999px;';
    return '';
  }

  function buttonStyle(extra = '') {
    return baseStyle(`${shapeStyle()} ${extra}`);
  }

  function buttonInnerStyle(extra = '') {
    return baseStyle(`${shapeStyle()} ${extra}`, { size: false, position: false, arrangement: false });
  }

  function assetName() {
    return firstNonEmpty(props.Picture, props.Image, props.BackgroundImage);
  }

  function hasElements(value) {
    if (Array.isArray(value)) return value.length > 0;
    if (value && typeof value === 'object') return Object.keys(value).length > 0;
    return String(value ?? '').trim() !== '';
  }

  function elementItems() {
    const source = hasElements(props.Elements) ? props.Elements : props.ElementsFromString;
    if (Array.isArray(source)) return source;
    if (source && typeof source === 'object') return Object.values(source);
    return elementsFromString(source);
  }

  function itemMainText(item) {
    if (Array.isArray(item)) return String(item[0] ?? '');
    if (item && typeof item === 'object') {
      return String(
        item.Text1 ?? item.text1 ?? item.MainText ?? item.mainText ?? item.Text ?? item.text ?? Object.values(item)[0] ?? '',
      );
    }
    return String(item ?? '');
  }

  function parseMaybeRow(value) {
    if (typeof value !== 'string') return value;
    const text = value.trim();
    if (!text.startsWith('{') && !text.startsWith('[')) return value;
    try {
      return JSON.parse(text);
    } catch {
      return value;
    }
  }

  function normalizeListRow(item, index) {
    const parsed = parseMaybeRow(item);
    if (Array.isArray(parsed)) {
      return {
        index,
        text1: String(parsed[0] ?? ''),
        text2: String(parsed[1] ?? ''),
        image: String(parsed[2] ?? ''),
      };
    }
    if (parsed && typeof parsed === 'object') {
      return {
        index,
        text1: String(parsed.Text1 ?? parsed.text1 ?? parsed.MainText ?? parsed.mainText ?? parsed.Text ?? ''),
        text2: String(parsed.Text2 ?? parsed.text2 ?? parsed.DetailText ?? parsed.detailText ?? ''),
        image: String(parsed.Image ?? parsed.image ?? parsed.ImageName ?? parsed.imageName ?? ''),
      };
    }
    return { index, text1: String(parsed ?? ''), text2: '', image: '' };
  }

  function listDataRows() {
    if (!String(props.ListData ?? '').trim()) return [];
    try {
      const parsed = JSON.parse(props.ListData);
      if (!Array.isArray(parsed)) return [];
      return parsed.map((row, index) => normalizeListRow(row, index));
    } catch {
      return [];
    }
  }

  function listViewRows() {
    const fromListData = listDataRows();
    if (fromListData.length) return fromListData;
    return rawElementItems.map((item, index) => normalizeListRow(item, index));
  }

  function textIncludes(value, filter) {
    const needle = String(filter ?? '').trim().toLowerCase();
    if (!needle) return true;
    return String(value ?? '').toLowerCase().includes(needle);
  }

  function filterIndexed(list, filter) {
    return list
      .map((text, index) => ({ text, index }))
      .filter(item => textIncludes(item.text, filter));
  }

  function actionValue(actionState, names) {
    for (const name of names) {
      const value = Number(actionState?.[name] ?? 0);
      if (value > 0) return value;
    }
    return 0;
  }

  function runAction(actionState, key, names, fn) {
    const value = actionValue(actionState, names);
    if (value === (actionTokens[key] ?? 0)) return;
    actionTokens[key] = value;
    if (value > 0) fn();
  }

  function handleComponentActions(actionState) {
    if (node?.type === 'Canvas') {
      clearCanvasDrawingLayerOnBackgroundChange();
      handleCanvasActionState(actionState);
    }
    runAction(actionState, 'open', ['open', 'Open', 'DisplayDropdown', 'LaunchPicker'], () => {
      if ((node?.type === 'ListPicker' || node?.type === 'Spinner') && enabled && visible) {
        pickerFilter = '';
        pickerOpen = true;
      }
      if (node?.type === 'DatePicker') openNativePicker(dateInput, true);
      if (node?.type === 'TimePicker') openNativePicker(timeInput, true);
    });
    runAction(actionState, 'navigate', ['navigate'], () => {
      const url = actionState.url ?? '';
      if (node?.type === 'WebViewer' && webViewerFrame) {
        webViewerFrame.src = url;
      }
    });
    runAction(actionState, 'goback', ['goback'], () => { try { webViewerFrame?.contentWindow?.history?.back(); } catch {} });
    runAction(actionState, 'goforward', ['goforward'], () => { try { webViewerFrame?.contentWindow?.history?.forward(); } catch {} });
    runAction(actionState, 'reload', ['reload'], () => { try { webViewerFrame?.contentWindow?.location?.reload(); } catch {} });
    runAction(actionState, 'play', ['play'], () => { videoEl?.play().catch(() => {}); });
    runAction(actionState, 'pause', ['pause'], () => { videoEl?.pause(); });
    runAction(actionState, 'fullscreen', ['fullscreen'], () => requestVideoFullscreen());
    runAction(actionState, 'seek', ['seek'], () => {
      if (videoEl) videoEl.currentTime = (numberOr(actionState.ms, 0)) / 1000;
    });
    runAction(actionState, 'focus', ['focus', 'Focus', 'RequestFocus'], () => focusCurrentInput());
    runAction(actionState, 'blur', ['blur', 'Blur', 'HideKeyboard'], () => blurCurrentInput());
    runAction(actionState, 'cursorStart', ['MoveCursorToStart', 'cursorStart', 'cursor-start'], () => setTextCursor(0));
    runAction(actionState, 'cursorEnd', ['MoveCursorToEnd', 'cursorEnd', 'cursor-end'], () => setTextCursor(textValue().length));
    runAction(actionState, 'cursorTo', ['MoveCursorTo', 'cursorTo', 'cursor-position'], () => {
      const position = numberOr(actionState.position ?? props.CursorPosition ?? props.SelectionStart, 1);
      setTextCursor(Math.max(0, position - 1));
    });
  }

  async function focusCurrentInput() {
    await tick();
    const target = textInputEl || buttonEl || dateInput || timeInput;
    if (target && !target.disabled) target.focus();
  }

  async function blurCurrentInput() {
    await tick();
    const target = textInputEl || buttonEl || dateInput || timeInput;
    if (target) target.blur();
  }

  async function setTextCursor(position) {
    await tick();
    if (!textInputEl || typeof textInputEl.setSelectionRange !== 'function') return;
    const next = Math.max(0, Math.min(position, textValue().length));
    textInputEl.focus();
    textInputEl.setSelectionRange(next, next);
  }

  function textValue() {
    return String(props.Text ?? '');
  }

  const LABEL_ALLOWED_TAGS = new Set([
    'a', 'b', 'big', 'blockquote', 'br', 'cite', 'code', 'dfn', 'div', 'em', 'font',
    'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'i', 'li', 'ol', 'p', 'pre', 's', 'small',
    'span', 'strike', 'strong', 'sub', 'sup', 'tt', 'u', 'ul',
  ]);
  const LABEL_ALLOWED_ATTRS = {
    a: new Set(['href']),
    font: new Set(['color', 'face']),
  };

  function labelEscapeHtml(value) {
    return String(value ?? '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function labelSanitizeNode(domNode) {
    if (domNode.nodeType === 3) return labelEscapeHtml(domNode.nodeValue);
    if (domNode.nodeType !== 1) return '';
    const tag = domNode.tagName.toLowerCase();
    const children = Array.from(domNode.childNodes).map(labelSanitizeNode).join('');
    if (!LABEL_ALLOWED_TAGS.has(tag)) return children;
    if (tag === 'br') return '<br>';
    let attrs = '';
    const allowed = LABEL_ALLOWED_ATTRS[tag];
    if (allowed) {
      for (const name of allowed) {
        const value = domNode.getAttribute(name);
        if (value == null) continue;
        if (name === 'href' && !/^(https?:|mailto:|tel:)/i.test(value.trim())) continue;
        attrs += ` ${name}="${labelEscapeHtml(value)}"`;
      }
    }
    return `<${tag}${attrs}>${children}</${tag}>`;
  }

  function labelHtml(text) {
    const raw = String(text ?? '');
    if (typeof document === 'undefined') return labelEscapeHtml(raw);
    const template = document.createElement('template');
    template.innerHTML = raw;
    return Array.from(template.content.childNodes).map(labelSanitizeNode).join('');
  }

  function dismissPicker() {
    pickerOpen = false;
    pickerFilter = '';
  }

  function onWindowPointerDown(e) {
    if (!pickerOpen) return;
    if (pickerWrapEl && pickerWrapEl.contains(e.target)) return;
    dismissPicker();
  }

  function listViewDetailStyle() {
    return [
      `color: ${colorValue(firstNonEmpty(props.TextColorDetail, props.TextColor), '#202124')};`,
      hasValue(props.FontSizeDetail) ? `font-size: ${numberOr(props.FontSizeDetail, 14)}px;` : '',
      typefaceStyleFor(props.FontTypefaceDetail),
    ].filter(Boolean).join(' ');
  }

  function listViewImageStyle() {
    return `width: ${numberOr(props.ImageWidth, 200)}px; height: ${numberOr(props.ImageHeight, 200)}px; max-width: 100%;`;
  }

  // ── LinearProgress helpers ──────────────────────────────────────────────
  function linearProgressPct() {
    const min = numberOr(props.Minimum, 0);
    const max = numberOr(props.Maximum, 100);
    const val = numberOr(props.Progress, 0);
    if (max <= min) return 0;
    return Math.max(0, Math.min(100, ((val - min) / (max - min)) * 100));
  }

  // ── WebViewer helpers ───────────────────────────────────────────────────
  function webViewerLoad(e) {
    try {
      const url = e.currentTarget.contentWindow?.location?.href || props.CurrentUrl || '';
      emitInteraction([{ component: node.name, property: 'CurrentUrl', value: url }], null);
      emitEvent(node.name, 'PageLoaded', [url]);
    } catch {}
  }
  function webViewerError() {
    emitEvent(node.name, 'ErrorOccurred', [-6, 'Failed to load', props.CurrentUrl || '']);
  }

  // ── VideoPlayer helpers ─────────────────────────────────────────────────
  let videoEl;
  function videoLoadedMetadata() {
    if (!videoEl) return;
    const dur = Math.round((videoEl.duration || 0) * 1000);
    emitInteraction([{ component: node.name, property: 'Duration', value: dur }], null);
  }

  async function requestVideoFullscreen() {
    await tick();
    const target = videoEl?.parentElement || videoEl;
    try {
      if (target?.requestFullscreen) await target.requestFullscreen();
      else if (videoEl?.webkitEnterFullscreen) videoEl.webkitEnterFullscreen();
    } catch {}
  }

  // ── File / Image / Contact pickers ─────────────────────────────────────
  let fileInput;
  const MOCK_CONTACTS = [
    { name: 'Alice Example', phone: '555-0101', email: 'alice@example.com' },
    { name: 'Bob Example',   phone: '555-0102', email: 'bob@example.com' },
    { name: 'Carol Example', phone: '555-0103', email: 'carol@example.com' },
  ];

  function filePickerAccept() {
    if (node?.type === 'ImagePicker') return 'image/*';
    if (node?.type === 'FilePicker') return props.MimeType && props.MimeType !== '*/*' ? props.MimeType : '';
    return '';
  }

  async function openFilePicker() {
    if (!enabled || consumeLongClick()) return;
    await emitEvent(node.name, 'BeforePicking');
    await tick();
    if (!enabled || !visible) return;
    if (node?.type === 'ContactPicker' || node?.type === 'PhoneNumberPicker') {
      pickerFilter = '';
      pickerOpen = true;
    } else {
      fileInput?.click();
    }
  }

  async function filePickerChange(e) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;
    const url = URL.createObjectURL(file);
    const patches = [{ component: node.name, property: 'Selection', value: file.name }];
    if (node?.type === 'ImagePicker') patches.push({ component: node.name, property: 'ImagePath', value: url });
    emitInteraction(patches, { component: node.name, event: 'AfterPicking', args: [] });
    e.currentTarget.value = '';
  }

  function pickContact(contact) {
    pickerOpen = false;
    const patches = [];
    if (node?.type === 'PhoneNumberPicker') {
      patches.push({ component: node.name, property: 'PhoneNumber', value: contact.phone });
      patches.push({ component: node.name, property: 'ContactName', value: contact.name });
    } else {
      patches.push({ component: node.name, property: 'ContactName', value: contact.name });
      patches.push({ component: node.name, property: 'EmailAddress', value: contact.email });
      patches.push({ component: node.name, property: 'PhoneNumber', value: contact.phone });
    }
    patches.push({ component: node.name, property: 'Selection', value: contact.name });
    emitInteraction(patches, { component: node.name, event: 'AfterPicking', args: [] });
  }

  // ── Canvas helpers ──────────────────────────────────────────────────────
  let canvasEl;
  let canvasCtx = null;
  let canvasDragStart = null;
  let canvasPrev = null;
  let canvasTouchStart = null;
  let canvasTouchTime = null;
  let canvasTouchedSprites = [];
  let canvasDrawOps = [];
  let lastCanvasActionSeq = 0;
  let canvasImageCache = new Map();
  let spriteAnimationFrame = null;
  let spriteLastTick = {};
  let spriteCollisionPairs = new Set();
  let canvasDraggedSprites = [];
  let canvasIsDrag = false;
  let canvasPointerHistory = [];
  let lastCanvasBackgroundKey = null;
  let canvasBackgroundSignature = '';
  const CANVAS_PAD_PX = 0;
  const EDGE_NORTH = 1;
  const EDGE_NORTHEAST = 2;
  const EDGE_EAST = 3;
  const EDGE_SOUTHEAST = 4;
  const EDGE_SOUTH = -1;
  const EDGE_SOUTHWEST = -2;
  const EDGE_WEST = -3;
  const EDGE_NORTHWEST = -4;
  const FLING_MIN_SPEED_PX_PER_MS = 1;

  function canvasWidth() {
    const v = props.Width;
    if (v === -2 || v === undefined || v === null) return 360;
    if (v === -1 || v === '') return 360;
    return Math.max(1, Number(v) || 360);
  }

  function canvasHeight() {
    const v = props.Height;
    if (v === -1 || v === undefined || v === null || v === '') return 300;
    if (v === -2) {
      // fill parent — use parent rendered height if available, else 300
      return canvasEl?.parentElement?.clientHeight || 300;
    }
    return Math.max(1, Number(v) || 300);
  }

  function canvasSprites() {
    return (node?.children || []).filter(c => c.type === 'Ball' || c.type === 'ImageSprite');
  }

  function getCanvas() {
    if (!canvasEl) return null;
    if (!canvasCtx) canvasCtx = canvasEl.getContext('2d');
    return canvasCtx;
  }

  function canvasPoint(e) {
    if (!canvasEl) return { x: 0, y: 0 };
    const rect = canvasEl.getBoundingClientRect();
    const scaleX = canvasEl.width / rect.width;
    const scaleY = canvasEl.height / rect.height;
    return {
      x: Math.max(0, Math.round((e.clientX - rect.left) * scaleX)),
      y: Math.max(0, Math.round((e.clientY - rect.top) * scaleY)),
    };
  }

  function unitValue(value, fallback = 0) {
    const n = numberOr(value, fallback);
    return Math.max(0, Math.min(1, n));
  }

  function spriteImageElement(sprite) {
    if (sprite.type !== 'ImageSprite') return null;
    const sp = state?.[sprite.name] ?? {};
    return cachedCanvasImage(resolveAssetUrl(assets, sp.Picture));
  }

  function spriteLength(value, natural = 0) {
    const n = numberOr(value, -1);
    if (n < 0) return Math.max(0, numberOr(natural, 0));
    return Math.max(0, n);
  }

  function spriteGeometry(sprite, override = null) {
    const sp = state?.[sprite.name] ?? {};
    const originX = numberOr(override?.X ?? sp.X, 0);
    const originY = numberOr(override?.Y ?? sp.Y, 0);
    if (sprite.type === 'Ball') {
      const r = Math.max(0, numberOr(sp.Radius, 5));
      const u = boolValue(sp.OriginAtCenter, false) ? 0.5 : 0;
      const v = boolValue(sp.OriginAtCenter, false) ? 0.5 : 0;
      const width = r * 2;
      const height = r * 2;
      const x = originX - width * u;
      const y = originY - height * v;
      return {
        x,
        y,
        width,
        height,
        cx: x + r,
        cy: y + r,
        originX,
        originY,
        u,
        v,
        radius: r,
      };
    }
    const img = spriteImageElement(sprite);
    const width = spriteLength(sp.Width, img?.naturalWidth || 0);
    const height = spriteLength(sp.Height, img?.naturalHeight || 0);
    const u = unitValue(sp.OriginX, 0);
    const v = unitValue(sp.OriginY, 0);
    const x = originX - width * u;
    const y = originY - height * v;
    return { x, y, width, height, cx: x + width / 2, cy: y + height / 2, originX, originY, u, v, radius: 0 };
  }

  function spriteContainsPoint(sprite, pt) {
    const sp = state?.[sprite.name] ?? {};
    if (sp.Visible === false || sp.Enabled === false) return false;
    const geom = spriteGeometry(sprite);
    if (sprite.type === 'Ball') {
      return Math.hypot(pt.x - geom.cx, pt.y - geom.cy) <= geom.radius;
    }
    if (boolValue(sp.Rotates, true) && numberOr(sp.Heading, 0) !== 0) {
      const angle = numberOr(sp.Heading, 0) * Math.PI / 180;
      const dx = pt.x - geom.originX;
      const dy = pt.y - geom.originY;
      const localX = geom.originX + dx * Math.cos(angle) - dy * Math.sin(angle);
      const localY = geom.originY + dx * Math.sin(angle) + dy * Math.cos(angle);
      return localX >= geom.x && localX <= geom.x + geom.width
        && localY >= geom.y && localY <= geom.y + geom.height;
    }
    return pt.x >= geom.x && pt.x <= geom.x + geom.width && pt.y >= geom.y && pt.y <= geom.y + geom.height;
  }

  function hitSpritesAt(pt) {
    return canvasSprites()
      .filter(sprite => spriteContainsPoint(sprite, pt))
      .sort((a, b) => numberOr(state?.[b.name]?.Z, 1) - numberOr(state?.[a.name]?.Z, 1));
  }

  function uniqueSprites(primary = [], fallback = []) {
    const seen = new Set();
    const out = [];
    for (const sprite of [...primary, ...fallback]) {
      if (!sprite?.name || seen.has(sprite.name)) continue;
      seen.add(sprite.name);
      out.push(sprite);
    }
    return out;
  }

  async function canvasPointerDown(e) {
    if (!enabled) return;
    canvasEl?.setPointerCapture(e.pointerId);
    const pt = canvasPoint(e);
    canvasTouchedSprites = hitSpritesAt(pt);
    canvasDraggedSprites = [...canvasTouchedSprites];
    canvasDragStart = pt;
    canvasPrev = pt;
    canvasTouchStart = pt;
    canvasTouchTime = Date.now();
    for (const sprite of canvasTouchedSprites) await emitEvent(sprite.name, 'TouchDown', [pt.x, pt.y]);
    await emitEvent(node.name, 'TouchDown', [pt.x, pt.y]);
    canvasIsDrag = false;
    canvasPointerHistory = [{ ...pt, t: canvasTouchTime }];
  }

  async function canvasPointerMove(e) {
    if (!enabled || !canvasDragStart) return;
    const pt = canvasPoint(e);
    const threshold = numberOr(props.TapThreshold, 15);
    if (!canvasIsDrag
      && Math.abs(pt.x - canvasDragStart.x) < threshold
      && Math.abs(pt.y - canvasDragStart.y) < threshold) {
      return;
    }
    canvasIsDrag = true;
    if (!boolValue(props.ExtendMovesOutsideCanvas, false)
      && (pt.x <= 0 || pt.x > canvasWidth() || pt.y <= 0 || pt.y > canvasHeight())) {
      return;
    }

    canvasDraggedSprites = uniqueSprites(canvasDraggedSprites, hitSpritesAt(pt));
    let handled = false;
    for (const sprite of canvasDraggedSprites) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false || sp.Enabled === false) continue;
      handled = true;
      await emitEvent(sprite.name, 'Dragged', [
        canvasDragStart.x, canvasDragStart.y,
        canvasPrev.x, canvasPrev.y,
        pt.x, pt.y,
      ]);
    }
    await emitEvent(node.name, 'Dragged', [
      canvasDragStart.x, canvasDragStart.y,
      canvasPrev.x, canvasPrev.y,
      pt.x, pt.y,
      handled,
    ]);
    canvasPrev = pt;
    canvasPointerHistory = [...canvasPointerHistory, { ...pt, t: Date.now() }].slice(-8);
  }

  async function canvasPointerUp(e) {
    if (!enabled) return;
    const pt = canvasPoint(e);
    const upTime = Date.now();
    canvasPointerHistory = [...canvasPointerHistory, { ...pt, t: upTime }].slice(-8);
    let handled = false;
    for (const sprite of canvasDraggedSprites) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false || sp.Enabled === false) continue;
      await emitEvent(sprite.name, 'Touched', [pt.x, pt.y]);
      await emitEvent(sprite.name, 'TouchUp', [pt.x, pt.y]);
      handled = true;
    }
    if (!canvasIsDrag) await emitEvent(node.name, 'Touched', [pt.x, pt.y, handled]);
    await emitEvent(node.name, 'TouchUp', [pt.x, pt.y]);
    await maybeEmitCanvasFling(pt, upTime);
    resetCanvasPointerState();
  }

  function canvasPointerCancel() {
    resetCanvasPointerState();
  }

  async function maybeEmitCanvasFling(pt, upTime) {
    if (!canvasTouchStart || canvasPointerHistory.length < 2) return;
    const recent = [...canvasPointerHistory].reverse().find(point => upTime - point.t >= 16 && upTime - point.t <= 140)
      || canvasPointerHistory[0];
    const dt = Math.max(1, upTime - recent.t);
    const vx = (pt.x - recent.x) / dt;
    const vy = (pt.y - recent.y) / dt;
    const speed = Math.hypot(vx, vy);
    const totalDist = Math.hypot(pt.x - canvasTouchStart.x, pt.y - canvasTouchStart.y);
    if (speed < FLING_MIN_SPEED_PX_PER_MS || totalDist < numberOr(props.TapThreshold, 15)) return;
    const heading = normalizeHeading(-Math.atan2(vy, vx) * 180 / Math.PI);
    let handled = false;
    for (const sprite of canvasTouchedSprites) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false || sp.Enabled === false) continue;
      await emitEvent(sprite.name, 'Flung', [canvasTouchStart.x, canvasTouchStart.y, speed, heading, vx, vy]);
      handled = true;
    }
    await emitEvent(node.name, 'Flung', [canvasTouchStart.x, canvasTouchStart.y, speed, heading, vx, vy, handled]);
  }

  function normalizeHeading(value) {
    const normalized = value % 360;
    return normalized < 0 ? normalized + 360 : normalized;
  }

  function resetCanvasPointerState() {
    canvasDragStart = null;
    canvasPrev = null;
    canvasTouchStart = null;
    canvasTouchTime = null;
    canvasTouchedSprites = [];
    canvasDraggedSprites = [];
    canvasIsDrag = false;
    canvasPointerHistory = [];
  }

  // Re-render canvas drawing ops when state changes
  $: if (canvasEl && state) { canvasDrawOps; applyCanvasOps(); }
  $: if (node?.type === 'Canvas') updateSpriteAnimationLoop();
  $: canvasBackgroundSignature = node?.type === 'Canvas'
    ? [props.BackgroundColor ?? '', props.BackgroundImage ?? '', props.BackgroundImageinBase64 ?? ''].join('\0')
    : '';
  $: if (node?.type === 'Canvas') clearCanvasDrawingLayerOnBackgroundChange(canvasBackgroundSignature);

  function applyCanvasOps() {
    const ctx = getCanvas();
    if (!ctx || !canvasEl) return;
    ctx.clearRect(0, 0, canvasEl.width, canvasEl.height);
    const bg = colorValue(props.BackgroundColor, '#ffffff');
    if (bg && bg !== 'transparent') {
      ctx.fillStyle = bg;
      ctx.fillRect(0, 0, canvasEl.width, canvasEl.height);
    }

    const bgImage = cachedCanvasImage(canvasBackgroundUrl());
    if (bgImage) {
      ctx.drawImage(bgImage, 0, 0, canvasEl.width, canvasEl.height);
    }

    for (const op of canvasDrawOps) drawCanvasOp(ctx, op);

    for (const sprite of canvasSprites().slice().sort((a, b) => numberOr(state?.[a.name]?.Z, 1) - numberOr(state?.[b.name]?.Z, 1))) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false) continue;
      ctx.save();
      if (sprite.type === 'Ball') {
        drawBall(ctx, sprite, sp);
      } else if (sprite.type === 'ImageSprite') {
        drawImageSprite(ctx, sprite, sp);
      }
      ctx.restore();
    }
  }

  function canvasBackgroundUrl() {
    const base64 = String(props.BackgroundImageinBase64 ?? '').trim();
    if (base64) return base64.startsWith('data:') ? base64 : `data:image/png;base64,${base64}`;
    return props.BackgroundImage ? assetUrl : '';
  }

  function canvasBackgroundKey() {
    return [
      props.BackgroundColor ?? '',
      props.BackgroundImage ?? '',
      props.BackgroundImageinBase64 ?? '',
    ].join('\0');
  }

  function clearCanvasDrawingLayerOnBackgroundChange(key = canvasBackgroundKey()) {
    if (lastCanvasBackgroundKey === null) {
      lastCanvasBackgroundKey = key;
      return;
    }
    if (key === lastCanvasBackgroundKey) return;
    lastCanvasBackgroundKey = key;
    if (!canvasDrawOps.length) return;
    canvasDrawOps = [];
    applyCanvasOps();
  }

  function cachedCanvasImage(url) {
    if (!url) return null;
    let entry = canvasImageCache.get(url);
    if (!entry) {
      const img = new Image();
      entry = { img };
      canvasImageCache.set(url, entry);
      img.onload = () => applyCanvasOps();
      img.onerror = () => canvasImageCache.delete(url);
      img.src = url;
    }
    return entry.img.complete && entry.img.naturalWidth > 0 ? entry.img : null;
  }

  function drawBall(ctx, sprite, sp) {
    const geom = spriteGeometry(sprite);
    ctx.fillStyle = colorValue(sp.PaintColor, '#000000');
    ctx.beginPath();
    ctx.arc(geom.cx, geom.cy, geom.radius, 0, Math.PI * 2);
    ctx.fill();
  }

  function drawImageSprite(ctx, sprite, sp) {
    const img = spriteImageElement(sprite);
    const geom = spriteGeometry(sprite);
    if (!img || geom.width <= 0 || geom.height <= 0) return;
    if (boolValue(sp.Rotates, true)) {
      const angle = numberOr(sp.Heading, 0) * Math.PI / 180;
      ctx.translate(geom.originX, geom.originY);
      ctx.rotate(-angle);
      ctx.drawImage(img, geom.x - geom.originX, geom.y - geom.originY, geom.width, geom.height);
    } else {
      ctx.drawImage(img, geom.x, geom.y, geom.width, geom.height);
    }
  }

  function handleCanvasActionState(actionState) {
    const ordered = Array.isArray(actionState?.__actions) ? actionState.__actions : [];
    const nextActions = ordered.filter(action => (
      action.seq > lastCanvasActionSeq
      && (action.action === 'canvas-draw' || action.action === 'canvas-clear')
    ));
    if (!nextActions.length) return;

    let nextOps = canvasDrawOps;
    for (const action of nextActions) {
      if (action.action === 'canvas-clear') {
        nextOps = [];
      } else if (action.action === 'canvas-draw') {
        nextOps = [...nextOps, action];
      }
      lastCanvasActionSeq = Math.max(lastCanvasActionSeq, action.seq);
    }
    canvasDrawOps = nextOps;
    applyCanvasOps();
  }

  function drawCanvasOp(ctx, op) {
    const color = colorValue(op.color ?? props.PaintColor, '#000000');
    const lineWidth = Math.max(1, numberOr(op.lineWidth ?? props.LineWidth, 2));
    ctx.lineWidth = lineWidth;
    ctx.strokeStyle = color;
    ctx.fillStyle = color;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';

    switch (op.op) {
      case 'line':
        ctx.beginPath();
        ctx.moveTo(numberOr(op.x1, 0), numberOr(op.y1, 0));
        ctx.lineTo(numberOr(op.x2, 0), numberOr(op.y2, 0));
        ctx.stroke();
        break;
      case 'circle':
        ctx.beginPath();
        ctx.arc(numberOr(op.cx, 0), numberOr(op.cy, 0), Math.max(0, numberOr(op.r, 0)), 0, Math.PI * 2);
        if (boolValue(op.fill, false)) ctx.fill();
        else ctx.stroke();
        break;
      case 'point': {
        const size = Math.max(1, lineWidth);
        ctx.fillRect(numberOr(op.x, 0) - size / 2, numberOr(op.y, 0) - size / 2, size, size);
        break;
      }
      case 'text':
        drawCanvasText(ctx, op, 0);
        break;
      case 'textAngle':
        drawCanvasText(ctx, op, -numberOr(op.angle, 0));
        break;
      case 'arc':
        drawCanvasArc(ctx, op);
        break;
      case 'shape':
        drawCanvasShape(ctx, op);
        break;
      default:
        break;
    }
  }

  function drawCanvasText(ctx, op, angle = 0) {
    const x = numberOr(op.x, 0);
    const y = numberOr(op.y, 0);
    ctx.save();
    ctx.translate(x, y);
    if (angle) ctx.rotate(angle * Math.PI / 180);
    ctx.font = `${Math.max(1, numberOr(op.fontSize ?? props.FontSize, 14))}px sans-serif`;
    ctx.textAlign = canvasTextAlign();
    ctx.textBaseline = 'alphabetic';
    ctx.fillText(String(op.text ?? ''), 0, 0);
    ctx.restore();
  }

  function canvasTextAlign() {
    const align = numberOr(props.TextAlignment, 1);
    if (align === 0) return 'left';
    if (align === 2) return 'right';
    return 'center';
  }

  function drawCanvasArc(ctx, op) {
    const left = numberOr(op.left, 0);
    const top = numberOr(op.top, 0);
    const right = numberOr(op.right, left);
    const bottom = numberOr(op.bottom, top);
    const cx = (left + right) / 2;
    const cy = (top + bottom) / 2;
    const rx = Math.max(0, Math.abs(right - left) / 2);
    const ry = Math.max(0, Math.abs(bottom - top) / 2);
    const start = numberOr(op.startAngle, 0) * Math.PI / 180;
    const sweep = numberOr(op.sweepAngle, 0) * Math.PI / 180;
    ctx.beginPath();
    if (boolValue(op.useCenter, false)) ctx.moveTo(cx, cy);
    ctx.ellipse(cx, cy, rx, ry, 0, start, start + sweep, sweep < 0);
    if (boolValue(op.useCenter, false)) ctx.closePath();
    if (boolValue(op.fill, false)) ctx.fill();
    else ctx.stroke();
  }

  function drawCanvasShape(ctx, op) {
    const pts = normalizeCanvasPoints(op.points);
    if (!pts.length) return;
    ctx.beginPath();
    ctx.moveTo(pts[0].x, pts[0].y);
    for (const pt of pts.slice(1)) ctx.lineTo(pt.x, pt.y);
    ctx.closePath();
    if (boolValue(op.fill, false)) ctx.fill();
    else ctx.stroke();
  }

  function normalizeCanvasPoints(value) {
    if (!Array.isArray(value)) return [];
    if (value.every(item => !Array.isArray(item) && typeof item !== 'object')) {
      const out = [];
      for (let i = 0; i + 1 < value.length; i += 2) {
        out.push({ x: numberOr(value[i], 0), y: numberOr(value[i + 1], 0) });
      }
      return out;
    }
    return value
      .map(item => {
        if (Array.isArray(item)) return { x: numberOr(item[0], 0), y: numberOr(item[1], 0) };
        if (item && typeof item === 'object') return { x: numberOr(item.x ?? item.X, 0), y: numberOr(item.y ?? item.Y, 0) };
        return null;
      })
      .filter(Boolean);
  }

  function movingCanvasSprites() {
    return canvasSprites().filter(sprite => {
      const sp = state?.[sprite.name] ?? {};
      return sp.Visible !== false && sp.Enabled !== false && numberOr(sp.Speed, 0) > 0;
    });
  }

  function updateSpriteAnimationLoop() {
    if (!canvasEl || node?.type !== 'Canvas') return;
    if (movingCanvasSprites().length > 0) {
      if (spriteAnimationFrame == null) {
        spriteAnimationFrame = requestAnimationFrame(spriteAnimationTick);
      }
    } else {
      stopSpriteAnimationLoop();
    }
  }

  function stopSpriteAnimationLoop() {
    if (spriteAnimationFrame != null) cancelAnimationFrame(spriteAnimationFrame);
    spriteAnimationFrame = null;
    spriteLastTick = {};
  }

  function spriteAnimationTick(timestamp) {
    spriteAnimationFrame = null;
    animateCanvasSprites(timestamp);
    detectSpriteCollisions();
    if (movingCanvasSprites().length > 0) {
      spriteAnimationFrame = requestAnimationFrame(spriteAnimationTick);
    } else {
      stopSpriteAnimationLoop();
    }
  }

  function animateCanvasSprites(timestamp) {
    const batchPatches = [];
    for (const sprite of movingCanvasSprites()) {
      const sp = state?.[sprite.name] ?? {};
      const interval = Math.max(1, numberOr(sp.Interval, 100));
      const last = spriteLastTick[sprite.name];
      if (last == null) {
        spriteLastTick[sprite.name] = timestamp;
        continue;
      }
      const elapsed = timestamp - last;
      if (elapsed < interval) continue;
      const steps = Math.max(1, Math.floor(elapsed / interval));
      spriteLastTick[sprite.name] = last + steps * interval;

      const heading = numberOr(sp.Heading, 0) * Math.PI / 180;
      const speed = numberOr(sp.Speed, 0) * steps;
      const next = clampSpritePosition(
        sprite,
        numberOr(sp.X, 0) + speed * Math.cos(heading),
        numberOr(sp.Y, 0) - speed * Math.sin(heading),
      );
      const patches = [
        { component: sprite.name, property: 'X', value: next.x },
        { component: sprite.name, property: 'Y', value: next.y },
      ];
      if (next.edge) {
        emitInteraction(patches, { component: sprite.name, event: 'EdgeReached', args: [next.edge] });
      } else {
        batchPatches.push(...patches);
      }
    }
    if (batchPatches.length) emitInteraction(batchPatches, null);
  }

  function clampSpritePosition(sprite, x, y) {
    const geom = spriteGeometry(sprite, { X: x, Y: y });
    const over = {
      west: geom.x < 0,
      north: geom.y < 0,
      east: geom.x + geom.width > canvasWidth(),
      south: geom.y + geom.height > canvasHeight(),
    };
    const edge = edgeDirection(over);
    if (!edge) return { x, y, edge: 0 };

    const moved = moveSpriteGeometryIntoCanvas(sprite, geom);
    return { ...moved, edge };
  }

  function edgeDirection({ west, north, east, south }) {
    if (west) {
      if (north) return EDGE_NORTHWEST;
      if (south) return EDGE_SOUTHWEST;
      return EDGE_WEST;
    }
    if (east) {
      if (north) return EDGE_NORTHEAST;
      if (south) return EDGE_SOUTHEAST;
      return EDGE_EAST;
    }
    if (north) return EDGE_NORTH;
    if (south) return EDGE_SOUTH;
    return 0;
  }

  function moveSpriteGeometryIntoCanvas(sprite, geom) {
    const width = canvasWidth();
    const height = canvasHeight();
    let left = geom.x;
    let top = geom.y;
    if (geom.width > width) left = 0;
    else if (geom.x < 0) left = 0;
    else if (geom.x + geom.width > width) left = width - geom.width;

    if (geom.height > height) top = 0;
    else if (geom.y < 0) top = 0;
    else if (geom.y + geom.height > height) top = height - geom.height;

    return {
      x: left + geom.width * geom.u,
      y: top + geom.height * geom.v,
    };
  }

  function detectSpriteCollisions() {
    const sprites = canvasSprites().filter(sprite => {
      const sp = state?.[sprite.name] ?? {};
      return sp.Visible !== false && sp.Enabled !== false;
    });
    const current = new Set();
    for (let i = 0; i < sprites.length; i += 1) {
      for (let j = i + 1; j < sprites.length; j += 1) {
        if (!spritesOverlap(sprites[i], sprites[j])) continue;
        current.add(collisionKey(sprites[i], sprites[j]));
      }
    }
    for (const key of current) {
      if (spriteCollisionPairs.has(key)) continue;
      const [a, b] = key.split('|');
      emitEvent(a, 'CollidedWith', [b]);
      emitEvent(b, 'CollidedWith', [a]);
    }
    for (const key of spriteCollisionPairs) {
      if (current.has(key)) continue;
      const [a, b] = key.split('|');
      emitEvent(a, 'NoLongerCollidingWith', [b]);
      emitEvent(b, 'NoLongerCollidingWith', [a]);
    }
    spriteCollisionPairs = current;
  }

  function collisionKey(a, b) {
    return [a.name, b.name].sort().join('|');
  }

  function spritesOverlap(a, b) {
    const ga = spriteGeometry(a);
    const gb = spriteGeometry(b);
    return ga.x < gb.x + gb.width
      && ga.x + ga.width > gb.x
      && ga.y < gb.y + gb.height
      && ga.y + ga.height > gb.y;
  }

  // ── Chart helpers ───────────────────────────────────────────────────────
  const CHART_PAD = 32;
  const CHART_COLORS = ['#2196F3','#F44336','#4CAF50','#FF9800','#9C27B0','#00BCD4'];

  function chartWidth() { return Math.max(80, numberOr(props.Width, 300)); }
  function chartHeight() { return Math.max(60, numberOr(props.Height, 300)); }

  function chartDataSeries() {
    if (!node?.children) return [];
    return node.children
      .filter(c => c.type === 'ChartData2D')
      .map((c, i) => {
        const sp = state?.[c.name] ?? {};
        const pts = parseChartPoints(sp.Elements || sp.ElementsFromPairs || '');
        const colors = parseChartColors(sp.Colors);
        return {
          label: sp.Label || c.name,
          color: colorValue(sp.Color, CHART_COLORS[i % CHART_COLORS.length]),
          colors,
          dataLabelColor: colorValue(sp.DataLabelColor, '#000000'),
          highlightColor: colorValue(sp.HighlightColor, ''),
          points: pts,
          name: c.name,
        };
      });
  }

  function parseChartPoints(value) {
    if (Array.isArray(value)) return value.map(p => Array.isArray(p) ? [Number(p[0])||0, Number(p[1])||0] : [0,0]);
    const text = String(value ?? '').trim();
    if (!text) return [];
    const pairs = text.split(',');
    const pts = [];
    for (let i = 0; i + 1 < pairs.length; i += 2) {
      pts.push([Number(pairs[i].trim()) || 0, Number(pairs[i + 1].trim()) || 0]);
    }
    return pts;
  }

  function parseChartColors(value) {
    if (Array.isArray(value)) return value.map(item => colorValue(item, '')).filter(Boolean);
    const text = String(value ?? '').trim();
    if (!text) return [];
    return text.split(',').map(item => colorValue(item.trim(), '')).filter(Boolean);
  }

  function chartXRange() {
    const all = chartDataSeries().flatMap(s => s.points.map(p => p[0]));
    const manualMin = finiteNumber(props.XMin);
    const manualMax = finiteNumber(props.XMax);
    if (manualMin != null && manualMax != null && manualMax > manualMin) return [manualMin, manualMax];
    if (!all.length) return [0, 1];
    const min = manualMin ?? (boolValue(props.XFromZero, false) ? 0 : Math.min(...all));
    const max = manualMax ?? Math.max(...all, min + 1);
    return [min, Math.max(max, min + 1)];
  }

  function chartYRange() {
    const all = chartDataSeries().flatMap(s => s.points.map(p => p[1]));
    const manualMin = finiteNumber(props.YMin);
    const manualMax = finiteNumber(props.YMax);
    if (manualMin != null && manualMax != null && manualMax > manualMin) return [manualMin, manualMax];
    if (!all.length) return [0, 1];
    const min = manualMin ?? (boolValue(props.YFromZero, false) ? 0 : Math.min(...all));
    const max = manualMax ?? Math.max(...all, min + 1);
    return [min, Math.max(max, min + 1)];
  }

  function finiteNumber(value) {
    if (value === null || value === undefined || value === '') return null;
    const n = Number(value);
    return Number.isFinite(n) ? n : null;
  }

  function chartX(v) {
    const [xMin, xMax] = chartXRange();
    const w = chartWidth() - CHART_PAD * 2;
    return CHART_PAD + ((v - xMin) / (xMax - xMin)) * w;
  }

  function chartY(v) {
    const [yMin, yMax] = chartYRange();
    const h = chartHeight() - CHART_PAD * 2;
    return chartHeight() - CHART_PAD - ((v - yMin) / (yMax - yMin)) * h;
  }

  function chartPolylinePoints(pts) {
    return pts.map(p => `${chartX(p[0])},${chartY(p[1])}`).join(' ');
  }

  function chartTicks(range, count = 5) {
    const [min, max] = range;
    const span = max - min;
    if (!Number.isFinite(span) || span <= 0) return [];
    return Array.from({ length: count }, (_, i) => min + (span * i) / (count - 1));
  }

  function chartXTicks() {
    return chartTicks(chartXRange());
  }

  function chartYTicks() {
    return chartTicks(chartYRange());
  }

  function chartTickText(value) {
    if (Math.abs(value) >= 1000 || (Math.abs(value) > 0 && Math.abs(value) < 0.01)) return value.toExponential(1);
    return Number.isInteger(value) ? String(value) : String(Math.round(value * 100) / 100);
  }

  function chartLabels() {
    if (Array.isArray(props.Labels)) return props.Labels.map(item => String(item ?? ''));
    return elementsFromString(props.Labels);
  }

  function chartXTickText(value, index) {
    return chartLabels()[index] || chartTickText(value);
  }

  function chartPointColor(series, index) {
    return series.colors[index] || series.color;
  }

  function chartAreaPath(pts) {
    if (!pts.length) return '';
    const base = chartHeight() - CHART_PAD;
    const points = [`${chartX(pts[0][0])},${base}`]
      .concat(pts.map(p => `${chartX(p[0])},${chartY(p[1])}`))
      .concat([`${chartX(pts[pts.length - 1][0])},${base}`]);
    return 'M' + points.join('L') + 'Z';
  }

  function chartBarX(si, pi, numSeries, numPts) {
    const w = chartWidth() - CHART_PAD * 2;
    const groupW = w / Math.max(1, numPts);
    const barW = chartBarWidth(numSeries, numPts);
    return CHART_PAD + pi * groupW + si * barW + (groupW - numSeries * barW) / 2;
  }

  function chartBarWidth(numSeries, numPts) {
    const w = chartWidth() - CHART_PAD * 2;
    return Math.max(2, (w / Math.max(1, numPts)) / Math.max(1, numSeries) - 2);
  }

  function pieSectors(pts, si) {
    const total = pts.reduce((s, p) => s + Math.abs(p[1]), 0) || 1;
    const cx = chartWidth() / 2;
    const cy = chartHeight() / 2;
    const radiusPct = Math.max(0, Math.min(100, numberOr(props.PieRadius, 100))) / 100;
    const r = (Math.min(cx, cy) - CHART_PAD) * radiusPct;
    let angle = -Math.PI / 2;
    return pts.map((p, i) => {
      const sweep = (Math.abs(p[1]) / total) * Math.PI * 2;
      const x1 = cx + r * Math.cos(angle);
      const y1 = cy + r * Math.sin(angle);
      const x2 = cx + r * Math.cos(angle + sweep);
      const y2 = cy + r * Math.sin(angle + sweep);
      const large = sweep > Math.PI ? 1 : 0;
      const d = `M${cx},${cy}L${x1},${y1}A${r},${r},0,${large},1,${x2},${y2}Z`;
      const series = chartDataSeries()[si];
      const fill = series?.colors?.[i] || CHART_COLORS[(si * pts.length + i) % CHART_COLORS.length];
      angle += sweep;
      return { d, fill };
    });
  }

  function chartEntryClick(series, pt) {
    emitEvent(node.name, 'EntryClick', [series.label, pt[0], pt[1]]);
    emitEvent(series.name, 'EntryClick', [pt[0], pt[1]]);
  }

  function chartTrendlines() {
    const series = chartDataSeries();
    return (node?.children || [])
      .filter(c => c.type === 'Trendline')
      .map((trend, index) => {
        const sp = state?.[trend.name] ?? {};
        if (sp.Visible === false) return null;
        const targetName = String(sp.ChartData ?? '').trim();
        const target = series.find(item => item.name === targetName || item.label === targetName) || series[index] || series[0];
        if (!target || target.points.length < 2) return null;
        const regression = linearRegression(target.points);
        if (!regression) return null;
        const [xMin, xMax] = chartXRange();
        const y1 = regression.slope * xMin + regression.intercept;
        const y2 = regression.slope * xMax + regression.intercept;
        return {
          name: trend.name,
          d: `M${chartX(xMin)},${chartY(y1)}L${chartX(xMax)},${chartY(y2)}`,
          color: colorValue(sp.Color, target.color),
          width: Math.max(1, numberOr(sp.StrokeWidth, 1)),
          dash: numberOr(sp.StrokeStyle, 1) === 2 ? '6 4' : numberOr(sp.StrokeStyle, 1) === 3 ? '2 4' : '',
        };
      })
      .filter(Boolean);
  }

  function linearRegression(points) {
    const n = points.length;
    if (n < 2) return null;
    let sumX = 0;
    let sumY = 0;
    let sumXY = 0;
    let sumXX = 0;
    for (const [x, y] of points) {
      sumX += x;
      sumY += y;
      sumXY += x * y;
      sumXX += x * x;
    }
    const denom = n * sumXX - sumX * sumX;
    if (denom === 0) return null;
    const slope = (n * sumXY - sumX * sumY) / denom;
    const intercept = (sumY - slope * sumX) / n;
    return { slope, intercept };
  }

  // ── Map helpers (Leaflet) ───────────────────────────────────────────────
  let mapEl;
  let mapInstance = null;
  let mapTileLayer = null;
  let mapZoomControl = null;
  let mapScaleControl = null;
  let mapCompassControl = null;
  let mapLayers = {};
  let mapFeatureActionSeq = {};

  // Map setup runs reactively via the $: if (mapEl) initOrUpdateMap() block below

  $: if (node?.type === 'Map' && mapEl) initOrUpdateMap();
  $: if (node?.type === 'Map' && mapInstance) handleMapFeatureActions();

  async function initOrUpdateMap() {
    const L = await import('leaflet').catch(() => null);
    if (!L || !mapEl) return;
    if (!mapInstance) {
      // Fix Leaflet default icon path for Vite bundling
      delete L.Icon.Default.prototype._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
        iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
      });
      mapInstance = L.map(mapEl, {
        center: [numberOr(props.Latitude, 42.359144), numberOr(props.Longitude, -71.093612)],
        zoom: numberOr(props.ZoomLevel, 13),
        zoomControl: false,
        dragging: boolValue(props.EnablePan, true),
        scrollWheelZoom: boolValue(props.EnableZoom, true),
        touchZoom: boolValue(props.EnableZoom, true),
        doubleClickZoom: boolValue(props.EnableZoom, true),
        attributionControl: true,
      });
      updateMapTileLayer(L);
      mapInstance.on('moveend', () => emitEvent(node.name, 'BoundsChange'));
      mapInstance.on('zoomend', () => emitEvent(node.name, 'ZoomChange'));
      mapInstance.on('click', (e) => emitEvent(node.name, 'TapAtPoint', [e.latlng.lat, e.latlng.lng]));
      mapInstance.on('dblclick', (e) => emitEvent(node.name, 'DoubleTapAtPoint', [e.latlng.lat, e.latlng.lng]));
      mapInstance.on('contextmenu', (e) => emitEvent(node.name, 'LongPressAtPoint', [e.latlng.lat, e.latlng.lng]));
      emitEvent(node.name, 'Ready');
    } else {
      updateMapTileLayer(L);
      mapInstance.setView(
        [numberOr(props.Latitude, 42.359144), numberOr(props.Longitude, -71.093612)],
        numberOr(props.ZoomLevel, 13),
        { animate: false },
      );
    }
    updateMapInteractions();
    updateMapControls(L);
    fitMapBoundsIfNeeded(L);
    updateMapFeatures(L);
  }

  function mapTileConfig() {
    const custom = String(props.CustomUrl ?? '').trim();
    const defaultUrl = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
    if (custom && custom !== defaultUrl) {
      return { url: custom, attribution: '' };
    }
    switch (numberOr(props.MapType, 1)) {
      case 2:
        return {
          url: 'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
          attribution: '© OpenTopoMap contributors',
        };
      case 3:
        return {
          url: 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png',
          attribution: '© OpenStreetMap contributors © CARTO',
        };
      default:
        return {
          url: defaultUrl,
          attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
        };
    }
  }

  function updateMapTileLayer(L) {
    if (!mapInstance || !L) return;
    const config = mapTileConfig();
    if (mapTileLayer?._simUrl === config.url) return;
    if (mapTileLayer) mapTileLayer.remove();
    mapTileLayer = L.tileLayer(config.url, {
      attribution: config.attribution,
      maxZoom: numberOr(props.MapType, 1) === 2 ? 17 : 19,
    });
    mapTileLayer._simUrl = config.url;
    mapTileLayer.addTo(mapInstance);
  }

  function updateMapInteractions() {
    if (!mapInstance) return;
    const pan = boolValue(props.EnablePan, true);
    const zoom = boolValue(props.EnableZoom, true);
    if (pan) mapInstance.dragging?.enable(); else mapInstance.dragging?.disable();
    for (const handler of ['scrollWheelZoom', 'touchZoom', 'doubleClickZoom', 'boxZoom', 'keyboard']) {
      if (zoom) mapInstance[handler]?.enable?.();
      else mapInstance[handler]?.disable?.();
    }
  }

  function updateMapControls(L) {
    if (!mapInstance || !L) return;
    if (boolValue(props.ShowZoom, false)) {
      if (!mapZoomControl) mapZoomControl = L.control.zoom({ position: 'topleft' }).addTo(mapInstance);
    } else if (mapZoomControl) {
      mapZoomControl.remove();
      mapZoomControl = null;
    }

    if (boolValue(props.ShowScale, false)) {
      if (!mapScaleControl) mapScaleControl = L.control.scale({ position: 'bottomleft', metric: true, imperial: true }).addTo(mapInstance);
    } else if (mapScaleControl) {
      mapScaleControl.remove();
      mapScaleControl = null;
    }

    if (boolValue(props.ShowCompass, false)) {
      if (!mapCompassControl) {
        const CompassControl = L.Control.extend({
          options: { position: 'topright' },
          onAdd() {
            const el = L.DomUtil.create('div', 'sim-map-compass');
            el.innerHTML = '<span>N</span>';
            return el;
          },
        });
        mapCompassControl = new CompassControl().addTo(mapInstance);
      }
      const el = mapCompassControl.getContainer?.();
      if (el) el.style.setProperty('--sim-map-rotation', `${numberOr(props.Rotation, 0)}deg`);
    } else if (mapCompassControl) {
      mapCompassControl.remove();
      mapCompassControl = null;
    }
  }

  function fitMapBoundsIfNeeded(L) {
    if (!mapInstance || !L) return;
    const bounds = parseBoundingBox(props.BoundingBox);
    if (!bounds) return;
    mapInstance.fitBounds(bounds, { animate: false });
  }

  function updateMapFeatures(L) {
    if (!mapInstance || !L) return;
    const nextKeys = new Set();
    for (const { child, collection } of mapFeatureEntries()) {
      const sp = state?.[child.name] ?? {};
      if (sp.Visible === false) continue;
      const key = child.name;
      const collectionName = collection?.name || '';
      nextKeys.add(key);
      if (mapLayers[key]?._simType !== child.type || mapLayers[key]?._simCollectionName !== collectionName) {
        mapLayers[key]?.remove();
        delete mapLayers[key];
      }
      if (mapLayers[key]) {
        updateMapLayer(L, child, sp, mapLayers[key]);
      } else {
        mapLayers[key] = createMapLayer(L, child, sp, collection);
        if (mapLayers[key]) {
          mapLayers[key]._simType = child.type;
          mapLayers[key]._simCollectionName = collectionName;
          mapLayers[key].addTo(mapInstance);
        }
      }
    }
    for (const k of Object.keys(mapLayers)) {
      if (!nextKeys.has(k)) { mapLayers[k]?.remove(); delete mapLayers[k]; }
    }
  }

  function mapFeatureEntries(children = node?.children || [], collection = null) {
    const out = [];
    for (const child of children) {
      if (child.type === 'FeatureCollection') {
        const collectionState = state?.[child.name] ?? {};
        if (collectionState.Visible === false) continue;
        out.push(...mapFeatureEntries(child.children || [], child));
      } else if (['Marker', 'Circle', 'LineString', 'Polygon', 'Rectangle'].includes(child.type)) {
        out.push({ child, collection });
      }
    }
    return out;
  }

  function featureStyle(sp) {
    return {
      color: colorValue(sp.StrokeColor, '#000000'),
      opacity: numberOr(sp.StrokeOpacity, 1),
      weight: numberOr(sp.StrokeWidth, 1),
      fillColor: colorValue(sp.FillColor, '#ff0000'),
      fillOpacity: numberOr(sp.FillOpacity, 1),
    };
  }

  function createMapLayer(L, child, sp, collection = null) {
    switch (child.type) {
      case 'Marker': {
        const icon = markerIcon(L, sp);
        const m = L.marker([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)], { icon, draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(m, sp);
        bindMapFeatureEvents(L, m, child, collection);
        m.on('dragend', () => {
          const latlng = m.getLatLng();
          emitInteraction([
            { component: child.name, property: 'Latitude', value: latlng.lat },
            { component: child.name, property: 'Longitude', value: latlng.lng },
          ], null);
        });
        return m;
      }
      case 'Circle': {
        const c = L.circle([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)], { radius: numberOr(sp.Radius, 10), ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(c, sp);
        bindMapFeatureEvents(L, c, child, collection);
        return c;
      }
      case 'LineString': {
        const pts = featurePoints(sp);
        if (!pts.length) return null;
        const l = L.polyline(pts, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(l, sp);
        bindMapFeatureEvents(L, l, child, collection);
        return l;
      }
      case 'Polygon': {
        const pts = polygonLatLngs(sp);
        if (!pts.length) return null;
        const pg = L.polygon(pts, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(pg, sp);
        bindMapFeatureEvents(L, pg, child, collection);
        return pg;
      }
      case 'Rectangle': {
        const bounds = [
          [numberOr(sp.SouthLatitude, 0), numberOr(sp.WestLongitude, 0)],
          [numberOr(sp.NorthLatitude, 0), numberOr(sp.EastLongitude, 0)],
        ];
        const r = L.rectangle(bounds, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(r, sp);
        bindMapFeatureEvents(L, r, child, collection);
        return r;
      }
      default: return null;
    }
  }

  function updateMapLayer(L, child, sp, layer) {
    if (child.type === 'Marker') {
      layer.setLatLng([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)]);
      layer.setIcon(markerIcon(L, sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'Circle') {
      layer.setLatLng([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)]);
      layer.setRadius(numberOr(sp.Radius, 10));
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'LineString') {
      const pts = featurePoints(sp);
      if (pts.length) layer.setLatLngs(pts);
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'Polygon') {
      const pts = polygonLatLngs(sp);
      if (pts.length) layer.setLatLngs(pts);
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'Rectangle') {
      const bounds = [
        [numberOr(sp.SouthLatitude, 0), numberOr(sp.WestLongitude, 0)],
        [numberOr(sp.NorthLatitude, 0), numberOr(sp.EastLongitude, 0)],
      ];
      layer.setBounds(bounds);
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    }
  }

  function bindFeaturePopup(layer, sp) {
    const title = String(sp.Title ?? '');
    const description = String(sp.Description ?? '');
    if (!title && !description) {
      layer.unbindPopup?.();
      return;
    }
    layer.bindPopup(`<b>${escapeHtml(title)}</b><br>${escapeHtml(description)}`);
  }

  function escapeHtml(value) {
    return String(value ?? '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function markerIcon(L, sp) {
    const imageUrl = resolveAssetUrl(assets, sp.ImageAsset);
    if (imageUrl) {
      const size = [32, 32];
      return L.icon({
        iconUrl: imageUrl,
        iconSize: size,
        iconAnchor: markerAnchor(sp, size[0], size[1]),
        popupAnchor: [0, -size[1] / 2],
      });
    }
    const size = [24, 36];
    return L.divIcon({
      html: `<svg viewBox="0 0 24 36" xmlns="http://www.w3.org/2000/svg" width="24" height="36"><path d="M12 0C5.4 0 0 5.4 0 12c0 9 12 24 12 24s12-15 12-24c0-6.6-5.4-12-12-12z" fill="${colorValue(sp.FillColor,'#ff0000')}" stroke="${colorValue(sp.StrokeColor,'#000')}"/><circle cx="12" cy="12" r="5" fill="white"/></svg>`,
      className: '',
      iconSize: size,
      iconAnchor: markerAnchor(sp, size[0], size[1]),
      popupAnchor: [0, -size[1]],
    });
  }

  function markerAnchor(sp, width, height) {
    const horizontal = numberOr(sp.AnchorHorizontal, 3);
    const vertical = numberOr(sp.AnchorVertical, 3);
    const x = horizontal === 1 ? 0 : horizontal === 2 ? width : width / 2;
    const y = vertical === 1 ? 0 : vertical === 2 ? height / 2 : height;
    return [x, y];
  }

  function bindMapFeatureEvents(L, layer, child, collection = null) {
    layer.on('click', async (e) => {
      stopLeafletPropagation(L, e);
      await emitMapFeatureEvent(child, collection, 'Click', 'FeatureClick');
    });
    layer.on('contextmenu', async (e) => {
      stopLeafletPropagation(L, e);
      await emitMapFeatureEvent(child, collection, 'LongClick', 'FeatureLongClick');
    });
    layer.on('dragstart', () => emitMapFeatureEvent(child, collection, 'StartDrag', 'FeatureStartDrag'));
    layer.on('drag', () => emitMapFeatureEvent(child, collection, 'Drag', 'FeatureDrag'));
    layer.on('dragend', () => emitMapFeatureEvent(child, collection, 'StopDrag', 'FeatureStopDrag'));
  }

  async function emitMapFeatureEvent(child, collection, featureEvent, aggregateEvent) {
    await emitEvent(child.name, featureEvent);
    if (collection?.name) await emitEvent(collection.name, aggregateEvent, [child.name]);
    await emitEvent(node.name, aggregateEvent, [child.name]);
  }

  function handleMapFeatureActions() {
    const nextSeq = { ...mapFeatureActionSeq };
    for (const { child } of mapFeatureEntries()) {
      const layer = mapLayers[child.name];
      if (!layer) continue;
      const actionState = actions?.[child.name] || {};
      for (const action of ['show-infobox', 'hide-infobox']) {
        const seq = numberOr(actionState[action], 0);
        const key = `${child.name}:${action}`;
        if (seq <= 0 || seq === nextSeq[key]) continue;
        nextSeq[key] = seq;
        if (action === 'show-infobox') layer.openPopup?.();
        else layer.closePopup?.();
      }
    }
    mapFeatureActionSeq = nextSeq;
  }

  function stopLeafletPropagation(L, e) {
    if (e?.originalEvent) L.DomEvent.stopPropagation(e.originalEvent);
  }

  function featurePoints(sp) {
    return parseLatLngList(sp.Points ?? sp.PointsFromString);
  }

  function polygonLatLngs(sp) {
    const outer = featurePoints(sp);
    if (!outer.length) return [];
    const holes = parseLatLngRings(sp.HolePoints ?? sp.HolePointsFromString);
    return holes.length ? [outer, ...holes] : outer;
  }

  function parseLatLngRings(value) {
    if (Array.isArray(value) && Array.isArray(value[0]) && Array.isArray(value[0][0])) {
      return value.map(ring => parseLatLngList(ring)).filter(ring => ring.length);
    }
    const text = String(value ?? '').trim();
    if (!text) return [];
    try {
      const parsed = JSON.parse(text);
      if (Array.isArray(parsed) && Array.isArray(parsed[0]) && Array.isArray(parsed[0][0])) {
        return parsed.map(ring => parseLatLngList(ring)).filter(ring => ring.length);
      }
      const single = parseLatLngList(parsed);
      return single.length ? [single] : [];
    } catch {
      const single = parseLatLngList(text);
      return single.length ? [single] : [];
    }
  }

  function parseLatLngList(value) {
    if (Array.isArray(value)) {
      if (value.every(point => !Array.isArray(point) && typeof point !== 'object')) {
        const nums = value.map(Number).filter(Number.isFinite);
        const pts = [];
        for (let i = 0; i + 1 < nums.length; i += 2) pts.push([nums[i], nums[i + 1]]);
        return pts;
      }
      return value
        .map(point => {
          if (Array.isArray(point)) return [Number(point[0]), Number(point[1])];
          if (point && typeof point === 'object') return [Number(point.Latitude ?? point.latitude ?? point.lat), Number(point.Longitude ?? point.longitude ?? point.lng)];
          return null;
        })
        .filter(point => point && point.every(Number.isFinite));
    }
    const text = String(value ?? '').trim();
    if (!text) return [];
    try {
      const parsed = JSON.parse(text);
      if (Array.isArray(parsed)) return parseLatLngList(parsed);
    } catch {}
    const nums = text.split(/[\s,]+/).map(Number).filter(Number.isFinite);
    const pts = [];
    for (let i = 0; i + 1 < nums.length; i += 2) pts.push([nums[i], nums[i + 1]]);
    return pts;
  }

  function parseBoundingBox(value) {
    if (Array.isArray(value)) {
      if (Array.isArray(value[0]) && Array.isArray(value[1])) return [value[0], value[1]];
      if (value.length >= 4) {
        const nums = value.map(Number);
        if (nums.every(Number.isFinite)) return [[nums[2], nums[1]], [nums[0], nums[3]]];
      }
    }
    const text = String(value ?? '').trim();
    if (!text) return null;
    try {
      return parseBoundingBox(JSON.parse(text));
    } catch {
      const nums = text.split(/[\s,]+/).map(Number).filter(Number.isFinite);
      if (nums.length >= 4) return [[nums[2], nums[1]], [nums[0], nums[3]]];
    }
    return null;
  }

  function childEvent(event) {
    dispatch(event.type, event.detail);
  }

  function emitInteraction(properties = [], event = null) {
    dispatch('interaction', { properties, event });
  }

  async function emitEvent(component, event, args = []) {
    const detail = { component, event, args };
    if (eventRunner) {
      await eventRunner(detail);
    } else {
      dispatch('event', detail);
    }
  }

  async function triggerNativePicker(input) {
    await tick();
    if (!input || input.disabled) return;
    input.focus();
    try {
      if (typeof input.showPicker === 'function') input.showPicker();
      else input.click();
    } catch {}
  }

  function focusEvent() {
    emitEvent(node.name, 'GotFocus');
  }

  function blurEvent() {
    emitEvent(node.name, 'LostFocus');
  }

  function clearLongClick() {
    if (longClickTimer) clearTimeout(longClickTimer);
    longClickTimer = null;
  }

  function pointerDown(longClick = false) {
    if (!enabled) return;
    emitEvent(node.name, 'TouchDown');
    if (!longClick) return;
    suppressClick = false;
    clearLongClick();
    longClickTimer = setTimeout(() => {
      longClickTimer = null;
      suppressClick = true;
      emitEvent(node.name, 'LongClick');
    }, 600);
  }

  function pointerUp() {
    if (!enabled) return;
    clearLongClick();
    emitEvent(node.name, 'TouchUp');
  }

  function consumeLongClick() {
    if (!suppressClick) return false;
    suppressClick = false;
    return true;
  }

  function textInput(e) {
    const value = e.currentTarget.value;
    emitInteraction(
      [{ component: node.name, property: 'Text', value }],
      { component: node.name, event: 'TextChanged', args: [] },
    );
  }

  function checkboxInput(e) {
    const property = node.type === 'Switch' ? 'On' : 'Checked';
    emitInteraction(
      [{ component: node.name, property, value: e.currentTarget.checked }],
      { component: node.name, event: 'Changed', args: [] },
    );
  }

  function sliderMin() {
    return numberOr(props.MinValue, 0);
  }

  function sliderMax() {
    return numberOr(props.MaxValue, 100);
  }

  function sliderValue() {
    return Math.max(sliderMin(), Math.min(sliderMax(), numberOr(props.ThumbPosition, sliderMin())));
  }

  function sliderStep() {
    const steps = numberOr(props.NumberOfSteps, 100);
    const span = Math.abs(sliderMax() - sliderMin());
    if (steps <= 0 || span === 0) return 'any';
    return String(span / steps);
  }

  function sliderStyle() {
    const span = sliderMax() - sliderMin();
    const progress = span === 0 ? 0 : Math.max(0, Math.min(100, ((sliderValue() - sliderMin()) / span) * 100));
    return baseStyle([
      `--slider-left: ${colorValue(props.ColorLeft, '#ffc800')};`,
      `--slider-right: ${colorValue(props.ColorRight, '#888888')};`,
      `--slider-thumb: ${colorValue(props.ThumbColor, '#444444')};`,
      `--slider-progress: ${progress}%;`,
    ].join(' '));
  }

  function sliderInput(e) {
    if (!boolValue(props.ThumbEnabled, true)) {
      e.preventDefault();
      return;
    }
    const value = Number(e.currentTarget.value);
    emitInteraction(
      [{ component: node.name, property: 'ThumbPosition', value }],
      { component: node.name, event: 'PositionChanged', args: [value] },
    );
  }

  function sliderPointerDown(e) {
    if (!boolValue(props.ThumbEnabled, true)) {
      e.preventDefault();
      return;
    }
    emitEvent(node.name, 'TouchDown');
  }

  function sliderPointerUp(e) {
    if (!boolValue(props.ThumbEnabled, true)) {
      e.preventDefault();
      return;
    }
    emitEvent(node.name, 'TouchUp');
  }

  function sliderKeydown(e) {
    if (boolValue(props.ThumbEnabled, true)) return;
    e.preventDefault();
  }

  function buttonClick() {
    if (!enabled || !visible || consumeLongClick()) return;
    emitEvent(node.name, 'Click');
  }

  function imageClick() {
    if (!visible || !boolValue(props.Clickable, false)) return;
    emitEvent(node.name, 'Click');
  }

  function validIndex(index) {
    return Number.isInteger(index) && index >= 0 && index < elements.length;
  }

  function selectionByIndex(index) {
    if (!validIndex(index)) return null;
    const selectionIndex = index + 1;
    const selection = elements[index] ?? '';
    return { selection, selectionIndex };
  }

  function spinnerChange(e) {
    const index = Number(e.currentTarget.value);
    if (!Number.isInteger(index) || index < 0) return;
    pickSpinnerItem(index);
  }

  function pickSpinnerItem(index) {
    const next = selectionByIndex(index);
    if (!next) return;
    pickerOpen = false;
    const suppressInitial = numberOr(props.SelectionIndex, 0) === 0 && index === 0;
    emitInteraction(
      [
        { component: node.name, property: 'Selection', value: next.selection },
        { component: node.name, property: 'SelectionIndex', value: next.selectionIndex },
      ],
      suppressInitial ? null : { component: node.name, event: 'AfterSelecting', args: [next.selection] },
    );
  }

  async function openListPicker() {
    if (!enabled || consumeLongClick()) return;
    await emitEvent(node.name, 'BeforePicking');
    await tick();
    if (!enabled || !visible) return;
    pickerFilter = '';
    pickerOpen = true;
  }

  function pickListItem(index) {
    const next = selectionByIndex(index);
    if (!next) return;
    pickerOpen = false;
    emitInteraction(
      [
        { component: node.name, property: 'Selection', value: next.selection },
        { component: node.name, property: 'SelectionIndex', value: next.selectionIndex },
      ],
      { component: node.name, event: 'AfterPicking', args: [] },
    );
  }

  function listViewPick(index) {
    const row = listRows[index];
    if (!row) return;
    emitInteraction(
      [
        { component: node.name, property: 'Selection', value: row.text1 },
        { component: node.name, property: 'SelectionIndex', value: index + 1 },
      ],
      { component: node.name, event: 'AfterPicking', args: [] },
    );
  }

  function currentDate() {
    const now = new Date();
    return {
      year: now.getFullYear(),
      month: now.getMonth() + 1,
      day: now.getDate(),
    };
  }

  function dateParts() {
    const today = currentDate();
    return {
      year: numberOr(props.Year, today.year),
      month: numberOr(props.Month, today.month),
      day: numberOr(props.Day, today.day),
    };
  }

  function dateText() {
    const { year, month, day } = dateParts();
    return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
  }

  function dateInstant(year, month, day) {
    const pad = n => String(n).padStart(2, '0');
    // Build the ISO string from local parts so the date never shifts across the
    // UTC offset (toISOString() would roll back a day in positive-offset locales),
    // matching the local-midnight format used by the defaults and the Go host.
    const iso = `${year}-${pad(month)}-${pad(day)}T00:00:00`;
    const millis = new Date(year, month - 1, day, 0, 0, 0, 0).getTime();
    return { iso, millis };
  }

  function dateChange(e) {
    const nextDate = e.currentTarget.value;
    if (!nextDate) return;
    const [year, month, day] = nextDate.split('-').map(Number);
    if (![year, month, day].every(Number.isFinite)) return;
    const instant = dateInstant(year, month, day);
    emitInteraction(
      [
        { component: node.name, property: 'Year', value: year },
        { component: node.name, property: 'Month', value: month },
        { component: node.name, property: 'Day', value: day },
        { component: node.name, property: 'MonthInText', value: MONTHS[month - 1] ?? '' },
        { component: node.name, property: 'Instant', value: instant.iso },
        { component: node.name, property: 'InstantMillis', value: instant.millis },
      ],
      { component: node.name, event: 'AfterDateSet', args: [] },
    );
  }

  function timeText() {
    return `${String(numberOr(props.Hour, 0)).padStart(2, '0')}:${String(numberOr(props.Minute, 0)).padStart(2, '0')}`;
  }

  function timeInstant(hour, minute) {
    const pad = n => String(n).padStart(2, '0');
    const date = new Date();
    date.setHours(hour, minute, 0, 0);
    // Format from local parts so the hour/minute never shift across the UTC offset,
    // matching capabilities.js timeInstant and the Go host's timeInstantString.
    const iso = `1970-01-01T${pad(hour)}:${pad(minute)}:00`;
    return { iso, millis: date.getTime() };
  }

  function timeChange(e) {
    const nextTime = e.currentTarget.value;
    if (!nextTime) return;
    const [hour, minute] = nextTime.split(':').map(Number);
    if (![hour, minute].every(Number.isFinite)) return;
    const instant = timeInstant(hour, minute);
    emitInteraction(
      [
        { component: node.name, property: 'Hour', value: hour },
        { component: node.name, property: 'Minute', value: minute },
        { component: node.name, property: 'Instant', value: instant.iso },
        { component: node.name, property: 'InstantMillis', value: instant.millis },
      ],
      { component: node.name, event: 'AfterTimeSet', args: [] },
    );
  }

  async function openNativePicker(input, fromAction = false) {
    if (!enabled) return;
    if (!fromAction) {
      await emitEvent(node.name, 'Click');
      await tick();
      if (!enabled || !visible) return;
    }
    triggerNativePicker(input);
  }

  function switchStyle() {
    return baseStyle([
      `--switch-thumb-active: ${colorValue(props.ThumbColorActive, '#ffffff')};`,
      `--switch-thumb-inactive: ${colorValue(props.ThumbColorInactive, '#cccccc')};`,
      `--switch-track-active: ${colorValue(props.TrackColorActive, '#00ff00')};`,
      `--switch-track-inactive: ${colorValue(props.TrackColorInactive, '#444444')};`,
    ].join(' '));
  }

  function pickerMenuStyle() {
    return [
      `--picker-item-color: ${colorValue(firstNonEmpty(props.ItemTextColor, props.TextColor), '#202124')};`,
      `--picker-item-bg: ${colorValue(props.ItemBackgroundColor, '#ffffff')};`,
      `--picker-selection-bg: ${colorValue(props.SelectionColor, '#e8f0fe')};`,
    ].join(' ');
  }

  function listViewItemStyle(selected = false) {
    const bg = selected
      ? colorValue(props.SelectionColor, '#d3d3d3')
      : colorValue(firstNonEmpty(props.ElementColor, props.ItemBackgroundColor), '#ffffff');
    return [
      `color: ${colorValue(firstNonEmpty(props.TextColor, props.ItemTextColor), '#202124')};`,
      `background: ${bg};`,
      hasValue(props.DividerColor) ? `border-bottom-color: ${colorValue(props.DividerColor, '#eef1f4')};` : '',
      `border-bottom-width: ${Math.max(0, numberOr(props.DividerThickness, 0))}px;`,
      `border-bottom-style: solid;`,
      `border-radius: ${Math.max(0, numberOr(props.ElementCornerRadius, 0))}px;`,
      `margin: ${Math.max(0, numberOr(props.ElementMarginsWidth, 0))}px;`,
    ].filter(Boolean).join(' ');
  }

  function rowImageUrl(row) {
    if (!row?.image || row.image === 'None') return '';
    return resolveAssetUrl(assets, row.image);
  }
</script>

<svelte:window on:pointerdown={onWindowPointerDown} />

{#if visible && !nonVisible}
  {#if node.type === 'Screen' || node.type === 'Form'}
    <div class="sim-screen-root" data-sim-component={node.name}>
      {#if boolValue(props.TitleVisible, true) && hasValue(props.Title)}
        <div class="sim-screen-titlebar" style="background: {colorValue(props.PrimaryColor, '#3f51b5')};">{props.Title}</div>
      {/if}
      <div
        class="sim-screen"
        class:sim-screen--scrollable={boolValue(props.Scrollable, false)}
        class:sim-unsupported={unsupportedHere}
        style={baseStyle('', { size: false })}
      >
        {#each node.children || [] as child (child.pathId || child.name)}
          <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
        {/each}
      </div>
    </div>
  {:else if node.type === 'VerticalArrangement' || node.type === 'VerticalScrollArrangement'}
    <div
      class="sim-arrangement"
      class:sim-scroll-vertical={node.type === 'VerticalScrollArrangement'}
      class:sim-unsupported={unsupportedHere}
      style={`${baseStyle()} ${emptyArrangementStyle()}`}
      data-sim-component={node.name}
    >
      {#each node.children || [] as child (child.pathId || child.name)}
        <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
      {/each}
    </div>
  {:else if node.type === 'HorizontalArrangement' || node.type === 'HorizontalScrollArrangement'}
    <div
      class="sim-arrangement sim-arrangement--horizontal"
      class:sim-scroll-horizontal={node.type === 'HorizontalScrollArrangement'}
      class:sim-unsupported={unsupportedHere}
      style={`${baseStyle()} ${emptyArrangementStyle()}`}
      data-sim-component={node.name}
    >
      {#each node.children || [] as child (child.pathId || child.name)}
        <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
      {/each}
    </div>
  {:else if node.type === 'AbsoluteArrangement'}
    <div class="sim-absolute" class:sim-unsupported={unsupportedHere} style={baseStyle('position: relative;')} data-sim-component={node.name}>
      {#each node.children || [] as child (child.pathId || child.name)}
        <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
      {/each}
    </div>
  {:else if node.type === 'Label'}
    <div class="sim-label" class:sim-unsupported={unsupportedHere} style={baseStyle(boolValue(props.HasMargins, true) ? 'margin: 2px;' : 'margin: 0;')} data-sim-component={node.name}>
      {#if boolValue(props.HTMLFormat, false)}{@html labelHtml(props.Text)}{:else}{props.Text ?? ''}{/if}
    </div>
  {:else if node.type === 'TextBox' || node.type === 'PasswordTextBox'}
    {#if isMultilineTextBox}
      <textarea
        bind:this={textInputEl}
        class="sim-textbox sim-textarea"
        class:sim-unsupported={unsupportedHere}
        style={`${baseStyle()} ${hintColorStyle()}`}
        value={props.Text ?? ''}
        placeholder={props.Hint ?? ''}
        disabled={!enabled}
        readonly={boolValue(props.ReadOnly, false)}
        inputmode={boolValue(props.NumbersOnly, false) ? 'decimal' : 'text'}
        rows="3"
        data-sim-component={node.name}
        on:input={textInput}
        on:focus={focusEvent}
        on:blur={blurEvent}
      ></textarea>
    {:else}
      <input
        bind:this={textInputEl}
        class="sim-textbox"
        class:sim-unsupported={unsupportedHere}
        style={`${baseStyle()} ${hintColorStyle()}`}
        type={node.type === 'PasswordTextBox' && !boolValue(props.PasswordVisible, false) ? 'password' : 'text'}
        value={props.Text ?? ''}
        placeholder={props.Hint ?? ''}
        disabled={!enabled}
        readonly={boolValue(props.ReadOnly, false)}
        inputmode={boolValue(props.NumbersOnly, false) ? 'decimal' : 'text'}
        data-sim-component={node.name}
        on:input={textInput}
        on:focus={focusEvent}
        on:blur={blurEvent}
      />
    {/if}
  {:else if node.type === 'Button'}
    <button
      bind:this={buttonEl}
      type="button"
      class="sim-button"
      class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
      class:sim-unsupported={unsupportedHere}
      style={buttonStyle()}
      disabled={!enabled}
      data-sim-component={node.name}
      on:pointerdown={() => pointerDown(true)}
      on:pointerup={pointerUp}
      on:pointercancel={clearLongClick}
      on:focus={focusEvent}
      on:blur={blurEvent}
      on:click={buttonClick}
    >
      {props.Text ?? ''}
    </button>
  {:else if node.type === 'CheckBox'}
    <label class="sim-check" class:sim-unsupported={unsupportedHere} style={baseStyle()} data-sim-component={node.name}>
      <input
        type="checkbox"
        checked={boolValue(props.Checked, false)}
        disabled={!enabled}
        on:change={checkboxInput}
        on:focus={focusEvent}
        on:blur={blurEvent}
      />
      <span>{props.Text ?? ''}</span>
    </label>
  {:else if node.type === 'Switch'}
    <label class="sim-check sim-switch" class:sim-unsupported={unsupportedHere} style={switchStyle()} data-sim-component={node.name}>
      <input
        type="checkbox"
        checked={boolValue(props.On, false)}
        disabled={!enabled}
        on:change={checkboxInput}
        on:focus={focusEvent}
        on:blur={blurEvent}
      />
      <span class="sim-switch-ui" aria-hidden="true"></span>
      <span>{props.Text ?? ''}</span>
    </label>
  {:else if node.type === 'Slider'}
    <input
      class="sim-slider"
      class:sim-slider-thumb-disabled={!boolValue(props.ThumbEnabled, true)}
      class:sim-unsupported={unsupportedHere}
      style={sliderStyle()}
      type="range"
      min={sliderMin()}
      max={sliderMax()}
      step={sliderStep()}
      value={sliderValue()}
      disabled={!enabled}
      data-sim-component={node.name}
      on:pointerdown={sliderPointerDown}
      on:pointerup={sliderPointerUp}
      on:keydown={sliderKeydown}
      on:input={sliderInput}
    />
  {:else if node.type === 'Image'}
    <button
      type="button"
      class="sim-image"
      class:sim-image-clickable={boolValue(props.Clickable, false)}
      class:sim-unsupported={unsupportedHere}
      style={baseStyle(`transform: rotate(${numberOr(props.RotationAngle, 0)}deg);`, { backgroundImage: false })}
      disabled={!boolValue(props.Clickable, false)}
      data-sim-component={node.name}
      on:click={imageClick}
    >
      {#if assetUrl}
        <img class:sim-image-fill={boolValue(props.ScalePictureToFit, false) || numberOr(props.Scaling, 0) === 1} src={assetUrl} alt={props.AlternateText ?? ''} />
      {:else}
        <span>{props.Picture ?? ''}</span>
      {/if}
    </button>
  {:else if node.type === 'Spinner'}
    <div bind:this={pickerWrapEl} class="sim-picker" class:sim-unsupported={unsupportedHere} style={containerStyle()} data-sim-component={node.name}>
      <select
        class="sim-select"
        style={buttonInnerStyle('width: 100%;')}
        disabled={!enabled}
        on:change={spinnerChange}
        on:pointerdown={() => emitEvent(node.name, 'TouchDown')}
        on:pointerup={() => emitEvent(node.name, 'TouchUp')}
        on:focus={focusEvent}
        on:blur={blurEvent}
      >
        {#if props.Prompt != null && props.Prompt !== ''}
          <option value="-1" disabled selected={numberOr(props.SelectionIndex, 0) < 1}>{props.Prompt}</option>
        {/if}
        {#each elements as item, index}
          <option value={index} selected={numberOr(props.SelectionIndex, 0) === index + 1 || (!props.Prompt && numberOr(props.SelectionIndex, 0) === 0 && index === 0)}>{item}</option>
        {/each}
      </select>
      {#if pickerOpen}
        <div class="sim-picker-menu" style={pickerMenuStyle()}>
          {#if props.Prompt != null && props.Prompt !== ''}
            <div class="sim-picker-title">{props.Prompt}</div>
          {/if}
          {#if boolValue(props.ShowFilterBar, false)}
            <input class="sim-picker-filter" value={pickerFilter} placeholder={props.HintText ?? 'Search list...'} on:input={e => pickerFilter = e.currentTarget.value} />
          {/if}
          {#each filteredPickerItems as item (item.index)}
            <button type="button" class:selected={numberOr(props.SelectionIndex, 0) === item.index + 1} on:click={() => pickSpinnerItem(item.index)}>{item.text}</button>
          {:else}
            <div class="sim-picker-empty"></div>
          {/each}
        </div>
      {/if}
    </div>
  {:else if node.type === 'ListPicker'}
    <div bind:this={pickerWrapEl} class="sim-picker" class:sim-unsupported={unsupportedHere} style={containerStyle()} data-sim-component={node.name}>
      <button
        bind:this={buttonEl}
        type="button"
        class="sim-button"
        class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
        style={buttonInnerStyle('width: 100%;')}
        disabled={!enabled}
        on:pointerdown={() => pointerDown(true)}
        on:pointerup={pointerUp}
        on:pointercancel={clearLongClick}
        on:focus={focusEvent}
        on:blur={blurEvent}
        on:click={openListPicker}
      >{props.Text ?? ''}</button>
      {#if pickerOpen}
        <div class="sim-picker-menu" style={pickerMenuStyle()}>
          {#if props.Title != null && props.Title !== ''}
            <div class="sim-picker-title">{props.Title}</div>
          {/if}
          {#if boolValue(props.ShowFilterBar, false)}
            <input class="sim-picker-filter" value={pickerFilter} placeholder={props.HintText ?? 'Search list...'} on:input={e => pickerFilter = e.currentTarget.value} />
          {/if}
          {#each filteredPickerItems as item (item.index)}
            <button type="button" on:click={() => pickListItem(item.index)}>{item.text}</button>
          {:else}
            <div class="sim-picker-empty"></div>
          {/each}
        </div>
      {/if}
    </div>
  {:else if node.type === 'DatePicker'}
    <div class="sim-native-picker-wrap" class:sim-unsupported={unsupportedHere} style={containerStyle()} data-sim-component={node.name}>
      <button
        bind:this={buttonEl}
        type="button"
        class="sim-button"
        style={buttonInnerStyle('width: 100%; height: 100%;')}
        disabled={!enabled}
        on:pointerdown={() => pointerDown(false)}
        on:pointerup={pointerUp}
        on:focus={focusEvent}
        on:blur={blurEvent}
        on:click={() => openNativePicker(dateInput)}
      >{props.Text ?? ''}</button>
      <input bind:this={dateInput} class="sim-native-picker-input" type="date" disabled={!enabled} value={dateText()} on:change={dateChange} tabindex="-1" aria-hidden="true" />
    </div>
  {:else if node.type === 'TimePicker'}
    <div class="sim-native-picker-wrap" class:sim-unsupported={unsupportedHere} style={containerStyle()} data-sim-component={node.name}>
      <button
        bind:this={buttonEl}
        type="button"
        class="sim-button"
        style={buttonInnerStyle('width: 100%; height: 100%;')}
        disabled={!enabled}
        on:pointerdown={() => pointerDown(false)}
        on:pointerup={pointerUp}
        on:focus={focusEvent}
        on:blur={blurEvent}
        on:click={() => openNativePicker(timeInput)}
      >{props.Text ?? ''}</button>
      <input bind:this={timeInput} class="sim-native-picker-input" type="time" disabled={!enabled} value={timeText()} on:change={timeChange} tabindex="-1" aria-hidden="true" />
    </div>
  {:else if node.type === 'ListView'}
    <div class="sim-listview" class:sim-unsupported={unsupportedHere} style={baseStyle()} data-sim-component={node.name}>
      {#if boolValue(props.ShowFilterBar, false)}
        <input class="sim-list-filter" value={listFilter} placeholder={props.HintText ?? 'Search list...'} on:input={e => listFilter = e.currentTarget.value} />
      {/if}
      <div class="sim-listview-items" class:horizontal={numberOr(props.Orientation, 1) === 0}>
        {#each filteredListRows as row (row.index)}
          <button
            type="button"
            class:selected={numberOr(props.SelectionIndex, 0) === row.index + 1}
            class:with-image={!!rowImageUrl(row)}
            style={listViewItemStyle(numberOr(props.SelectionIndex, 0) === row.index + 1)}
            on:click={() => listViewPick(row.index)}
          >
            {#if rowImageUrl(row)}
              <img src={rowImageUrl(row)} alt="" style={listViewImageStyle()} />
            {/if}
            <span class="sim-list-text">
              <span>{row.text1}</span>
              {#if row.text2}
                <small style={listViewDetailStyle()}>{row.text2}</small>
              {/if}
            </span>
          </button>
        {:else}
          <div class="sim-picker-empty"></div>
        {/each}
      </div>
    </div>
  {:else if node.type === 'CircularProgress'}
    <div
      class="sim-circular-progress"
      class:sim-unsupported={unsupportedHere}
      style={baseStyle(`--cp-color: ${colorValue(props.Color, '#0000ff')};`, { typography: false, arrangement: false, backgroundImage: false })}
      role="progressbar"
      aria-label="Loading"
      data-sim-component={node.name}
    >
      <span class="sim-cp-ring" aria-hidden="true"></span>
    </div>
  {:else if node.type === 'LinearProgress'}
    <div
      class="sim-linear-progress"
      class:sim-linear-progress--indeterminate={boolValue(props.Indeterminate, true)}
      class:sim-unsupported={unsupportedHere}
      style={baseStyle(`--lp-color: ${colorValue(props.ProgressColor, '#0000ff')}; --lp-ind-color: ${colorValue(props.IndeterminateColor, '#0000ff')}; --lp-pct: ${linearProgressPct()}%;`, { typography: false, arrangement: false, backgroundImage: false })}
      role="progressbar"
      aria-valuenow={boolValue(props.Indeterminate, true) ? undefined : numberOr(props.Progress, 0)}
      aria-valuemin={numberOr(props.Minimum, 0)}
      aria-valuemax={numberOr(props.Maximum, 100)}
      data-sim-component={node.name}
    >
      <div class="sim-lp-bar" aria-hidden="true"></div>
    </div>
  {:else if node.type === 'TableArrangement'}
    <div
      class="sim-table"
      class:sim-unsupported={unsupportedHere}
      style={baseStyle(`grid-template-columns: repeat(${numberOr(props.Columns, 2)}, 1fr); grid-template-rows: repeat(${numberOr(props.Rows, 2)}, auto);`, { typography: false, arrangement: false, backgroundImage: false })}
      data-sim-component={node.name}
    >
      {#each node.children || [] as child (child.pathId || child.name)}
        <svelte:self node={child} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
      {/each}
    </div>
  {:else if node.type === 'WebViewer'}
    <div class="sim-webviewer" class:sim-unsupported={unsupportedHere} style={baseStyle()} data-sim-component={node.name}>
      {#if props.HomeUrl || props.CurrentUrl}
        <iframe
          bind:this={webViewerFrame}
          title="WebViewer"
          src={props.CurrentUrl || props.HomeUrl || ''}
          class="sim-webviewer-frame"
          sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
          allow="autoplay; camera; microphone"
          on:load={webViewerLoad}
          on:error={webViewerError}
        ></iframe>
      {:else}
        <div class="sim-webviewer-empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 8h18M8 3v18"/></svg>
          <span>WebViewer</span>
        </div>
      {/if}
    </div>
  {:else if node.type === 'VideoPlayer'}
    <div class="sim-videoplayer" class:sim-unsupported={unsupportedHere} style={baseStyle()} data-sim-component={node.name}>
      {#if props.Source}
        <video
          bind:this={videoEl}
          class="sim-videoplayer-video"
          src={resolveAssetUrl(assets, props.Source) || props.Source}
          volume={Math.max(0, Math.min(1, numberOr(props.Volume, 50) / 100))}
          loop={boolValue(props.Loop, false)}
          preload="metadata"
          controls
          on:ended={() => emitEvent(node.name, 'Completed')}
          on:error={(e) => emitEvent(node.name, 'VideoPlayerError', [e.currentTarget.error?.message || 'Error'])}
          on:loadedmetadata={videoLoadedMetadata}
        >
          <track kind="captions" />
        </video>
      {:else}
        <div class="sim-videoplayer-empty">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><circle cx="12" cy="12" r="9"/><polygon points="10,8 16,12 10,16" fill="currentColor" stroke="none"/></svg>
          <span>VideoPlayer</span>
        </div>
      {/if}
    </div>
  {:else if node.type === 'EmailPicker'}
    <input
      bind:this={textInputEl}
      class="sim-textbox"
      class:sim-unsupported={unsupportedHere}
      style={`${baseStyle()} ${hintColorStyle()}`}
      type="email"
      autocomplete="email"
      value={props.Text ?? ''}
      placeholder={props.Hint ?? ''}
      disabled={!enabled}
      data-sim-component={node.name}
      on:input={textInput}
      on:focus={focusEvent}
      on:blur={blurEvent}
    />
  {:else if node.type === 'ImagePicker' || node.type === 'FilePicker' || node.type === 'ContactPicker' || node.type === 'PhoneNumberPicker'}
    <div bind:this={pickerWrapEl} class="sim-picker" class:sim-unsupported={unsupportedHere} style={containerStyle()} data-sim-component={node.name}>
      <button
        bind:this={buttonEl}
        type="button"
        class="sim-button"
        class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
        style={buttonInnerStyle('width: 100%;')}
        disabled={!enabled}
        on:pointerdown={() => pointerDown(true)}
        on:pointerup={pointerUp}
        on:pointercancel={clearLongClick}
        on:focus={focusEvent}
        on:blur={blurEvent}
        on:click={openFilePicker}
      >{props.Text ?? ''}</button>
      <input
        bind:this={fileInput}
        type="file"
        accept={filePickerAccept()}
        class="sim-native-picker-input"
        tabindex="-1"
        aria-hidden="true"
        on:change={filePickerChange}
      />
      {#if node.type === 'ContactPicker' || node.type === 'PhoneNumberPicker'}
        {#if pickerOpen}
          <div class="sim-picker-menu" style={pickerMenuStyle()}>
            <div class="sim-picker-title">
              {node.type === 'PhoneNumberPicker' ? 'Pick Phone Number' : 'Pick Contact'}
            </div>
            {#each MOCK_CONTACTS as contact, i (i)}
              <button type="button" on:click={() => pickContact(contact)}>{contact.name}</button>
            {/each}
          </div>
        {/if}
      {/if}
    </div>
  {:else if node.type === 'Canvas'}
    <div
      class="sim-canvas-wrap"
      class:sim-unsupported={unsupportedHere}
      style={baseStyle()}
      data-sim-component={node.name}
    >
      <canvas
        bind:this={canvasEl}
        class="sim-canvas"
        width={canvasWidth()}
        height={canvasHeight()}
        on:pointerdown={canvasPointerDown}
        on:pointermove={canvasPointerMove}
        on:pointerup={canvasPointerUp}
        on:pointercancel={canvasPointerCancel}
      ></canvas>
      {#each canvasSprites() as sprite (sprite.name)}
        <svelte:self node={sprite} {state} {unsupported} {assets} {actions} {eventRunner} parentType={node.type} on:event={childEvent} on:property={childEvent} on:interaction={childEvent} />
      {/each}
    </div>
  {:else if node.type === 'Ball'}
    <!-- Ball is rendered by the Canvas parent onto its canvas element; this node renders nothing itself -->
  {:else if node.type === 'ImageSprite'}
    <!-- ImageSprite is rendered by the Canvas parent onto its canvas element; this node renders nothing itself -->
  {:else if node.type === 'Chart'}
    <div
      class="sim-chart"
      class:sim-unsupported={unsupportedHere}
      style={baseStyle()}
      data-sim-component={node.name}
    >
      <svg class="sim-chart-svg" viewBox={`0 0 ${chartWidth()} ${chartHeight()}`} preserveAspectRatio="none">
        {#if numberOr(props.Type, 0) !== 4}
          {#if boolValue(props.GridEnabled, true)}
            {#each chartXTicks() as tick}
              <line class="sim-chart-grid-line" x1={chartX(tick)} y1={CHART_PAD} x2={chartX(tick)} y2={chartHeight() - CHART_PAD} />
            {/each}
            {#each chartYTicks() as tick}
              <line class="sim-chart-grid-line" x1={CHART_PAD} y1={chartY(tick)} x2={chartWidth() - CHART_PAD} y2={chartY(tick)} />
            {/each}
          {/if}
          <line class="sim-chart-axis-line" x1={CHART_PAD} y1={chartHeight() - CHART_PAD} x2={chartWidth() - CHART_PAD} y2={chartHeight() - CHART_PAD} />
          <line class="sim-chart-axis-line" x1={CHART_PAD} y1={CHART_PAD} x2={CHART_PAD} y2={chartHeight() - CHART_PAD} />
          {#each chartXTicks() as tick, ti}
            <text class="sim-chart-axis-text" x={chartX(tick)} y={chartHeight() - 8} text-anchor="middle" fill={colorValue(props.AxesTextColor, '#000000')}>{chartXTickText(tick, ti)}</text>
          {/each}
          {#each chartYTicks() as tick}
            <text class="sim-chart-axis-text" x={CHART_PAD - 5} y={chartY(tick) + 3} text-anchor="end" fill={colorValue(props.AxesTextColor, '#000000')}>{chartTickText(tick)}</text>
          {/each}
        {/if}
        {#each chartDataSeries() as series, si}
          {#if numberOr(props.Type, 0) === 4}
            <!-- Pie chart -->
            {#each pieSectors(series.points, si) as sector, pi}
              <path
                role="button"
                tabindex="0"
                d={sector.d}
                fill={sector.fill}
                stroke="white"
                stroke-width="1"
                on:click={() => chartEntryClick(series, series.points[pi])}
                on:keydown={(e) => e.key === 'Enter' && chartEntryClick(series, series.points[pi])}
              />
            {/each}
          {:else if numberOr(props.Type, 0) === 3}
            <!-- Bar chart -->
            {#each series.points as pt, pi}
              <rect
                role="button"
                tabindex="0"
                x={chartBarX(si, pi, chartDataSeries().length, series.points.length)}
                y={chartY(pt[1])}
                width={chartBarWidth(chartDataSeries().length, series.points.length)}
                height={chartHeight() - CHART_PAD - chartY(pt[1])}
                fill={chartPointColor(series, pi)}
                on:click={() => chartEntryClick(series, pt)}
                on:keydown={(e) => e.key === 'Enter' && chartEntryClick(series, pt)}
              />
              <text class="sim-chart-data-label" x={chartBarX(si, pi, chartDataSeries().length, series.points.length) + chartBarWidth(chartDataSeries().length, series.points.length) / 2} y={chartY(pt[1]) - 4} text-anchor="middle" fill={series.dataLabelColor}>{pt[1]}</text>
            {/each}
          {:else}
            <!-- Line / scatter / area -->
            {#if numberOr(props.Type, 0) === 2}
              <!-- Area fill -->
              <path d={chartAreaPath(series.points)} fill={series.color} opacity="0.3" />
            {/if}
            {#if numberOr(props.Type, 0) !== 1}
              <!-- Line -->
              <polyline points={chartPolylinePoints(series.points)} fill="none" stroke={series.color} stroke-width="2" />
            {/if}
            {#each series.points as pt, pi}
              <circle
                role="button"
                tabindex="0"
                cx={chartX(pt[0])}
                cy={chartY(pt[1])}
                r="4"
                fill={chartPointColor(series, pi)}
                stroke={series.highlightColor || 'none'}
                stroke-width={series.highlightColor ? 2 : 0}
                on:click={() => chartEntryClick(series, pt)}
                on:keydown={(e) => e.key === 'Enter' && chartEntryClick(series, pt)}
              />
              <text class="sim-chart-data-label" x={chartX(pt[0])} y={chartY(pt[1]) - 7} text-anchor="middle" fill={series.dataLabelColor}>{pt[1]}</text>
            {/each}
          {/if}
        {/each}
        {#each chartTrendlines() as trendline}
          <path d={trendline.d} fill="none" stroke={trendline.color} stroke-width={trendline.width} stroke-dasharray={trendline.dash} />
        {/each}
      </svg>
      {#if props.Description}
        <div class="sim-chart-description">{props.Description}</div>
      {/if}
      {#if boolValue(props.LegendEnabled, true) && chartDataSeries().length > 0}
        <div class="sim-chart-legend">
          {#each chartDataSeries() as series}
            <span><span class="sim-chart-legend-dot" style="background:{series.color}"></span>{series.label}</span>
          {/each}
        </div>
      {/if}
    </div>
  {:else if node.type === 'Map'}
    <div
      bind:this={mapEl}
      class="sim-map"
      class:sim-unsupported={unsupportedHere}
      style={baseStyle()}
      data-sim-component={node.name}
    ></div>
  {:else if node.type === 'FeatureCollection'}
    <!-- FeatureCollection is a non-rendering logical container; map features render inside the Map host -->
  {:else}
    <div class="sim-unsupported sim-placeholder" style={baseStyle()} data-sim-component={node.name}>
      {node.type}.{node.name}
    </div>
  {/if}
{/if}

<style>
  .sim-screen-root {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }

  .sim-screen-titlebar {
    flex: 0 0 auto;
    padding: 9px 14px;
    font-size: 14px;
    font-weight: 600;
    line-height: 1.3;
    color: #fff;
    overflow-wrap: anywhere;
  }

  .sim-screen {
    flex: 1 1 auto;
    min-height: 0;
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 14px;
    overflow: hidden;
    background: #fff;
    color: #202124;
  }

  .sim-screen--scrollable {
    overflow-y: auto;
    overflow-x: hidden;
  }

  .sim-arrangement {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
  }

  .sim-arrangement--horizontal {
    flex-direction: row;
    flex-wrap: nowrap;
  }

  .sim-scroll-horizontal {
    overflow-x: auto;
  }

  .sim-scroll-vertical {
    overflow-y: auto;
    overflow-x: hidden;
    max-height: 100%;
  }

  .sim-absolute {
    min-height: 0;
    overflow: visible;
  }

  .sim-label {
    min-height: 20px;
    line-height: 1.35;
    overflow-wrap: anywhere;
  }

  .sim-textbox,
  .sim-select,
  .sim-picker-filter,
  .sim-list-filter {
    min-height: 34px;
    padding: 6px 8px;
    border: 1px solid #c9cdd3;
    border-radius: 3px;
    background: #fff;
    color: #202124;
    font: inherit;
  }

  .sim-textarea {
    min-height: 72px;
    resize: vertical;
    white-space: pre-wrap;
  }

  .sim-textbox:disabled,
  .sim-button:disabled,
  .sim-select:disabled,
  .sim-slider:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .sim-textbox[readonly] {
    cursor: default;
  }

  .sim-textbox::placeholder {
    color: var(--sim-placeholder-color, #9aa0a6);
    opacity: 1;
  }

  .sim-button {
    min-height: 34px;
    min-width: 38px;
    padding: 6px 12px;
    border: 1px solid #7b8794;
    border-radius: 4px;
    background-color: #5b6470;
    color: #fff;
    font: inherit;
    cursor: pointer;
    background-repeat: no-repeat;
    background-position: center;
    background-size: cover;
  }

  .sim-button:not(:disabled):not(.sim-no-feedback):active {
    transform: translateY(1px);
  }

  .sim-check {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 28px;
    color: #202124;
  }

  .sim-check input {
    width: 18px;
    height: 18px;
  }

  .sim-switch {
    position: relative;
  }

  .sim-switch input {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
  }

  .sim-switch-ui {
    position: relative;
    display: inline-block;
    width: 44px;
    height: 24px;
    flex: 0 0 auto;
    border-radius: 999px;
    background: var(--switch-track-inactive, #444444);
    transition: background 0.12s ease;
  }

  .sim-switch-ui::after {
    content: "";
    position: absolute;
    top: 3px;
    left: 3px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: var(--switch-thumb-inactive, #cccccc);
    transition: transform 0.12s ease, background 0.12s ease;
  }

  .sim-switch input:checked + .sim-switch-ui {
    background: var(--switch-track-active, #00ff00);
  }

  .sim-switch input:checked + .sim-switch-ui::after {
    transform: translateX(20px);
    background: var(--switch-thumb-active, #ffffff);
  }

  .sim-switch input:focus-visible + .sim-switch-ui,
  .sim-button:focus-visible,
  .sim-textbox:focus-visible,
  .sim-select:focus-visible,
  .sim-picker-filter:focus-visible,
  .sim-list-filter:focus-visible {
    outline: 2px solid #2a6f97;
    outline-offset: 2px;
  }

  .sim-slider {
    min-height: 28px;
    appearance: none;
    -webkit-appearance: none;
    background: transparent;
  }

  .sim-slider::-webkit-slider-runnable-track {
    height: 8px;
    border-radius: 999px;
    background: linear-gradient(
      to right,
      var(--slider-left, #ffc800) 0%,
      var(--slider-left, #ffc800) var(--slider-progress, 50%),
      var(--slider-right, #888888) var(--slider-progress, 50%),
      var(--slider-right, #888888) 100%
    );
  }

  .sim-slider::-moz-range-track {
    height: 8px;
    border-radius: 999px;
    background: var(--slider-right, #888888);
  }

  .sim-slider::-moz-range-progress {
    height: 8px;
    border-radius: 999px;
    background: var(--slider-left, #ffc800);
  }

  .sim-slider::-webkit-slider-thumb {
    width: 18px;
    height: 18px;
    margin-top: -5px;
    border: 0;
    border-radius: 50%;
    background: var(--slider-thumb, #444444);
    -webkit-appearance: none;
  }

  .sim-slider::-moz-range-thumb {
    width: 18px;
    height: 18px;
    border: 0;
    border-radius: 50%;
    background: var(--slider-thumb, #444444);
  }

  .sim-slider-thumb-disabled {
    pointer-events: none;
  }

  .sim-slider-thumb-disabled::-webkit-slider-thumb {
    opacity: 0;
  }

  .sim-slider-thumb-disabled::-moz-range-thumb {
    opacity: 0;
  }

  .sim-image {
    display: grid;
    place-items: center;
    min-height: 52px;
    border: 1px dashed #c9cdd3;
    background: #f8fafc;
    color: #6b7280;
    overflow: hidden;
    padding: 0;
  }

  .sim-image-clickable {
    cursor: pointer;
  }

  .sim-image:disabled {
    cursor: default;
  }

  .sim-image img {
    width: 100%;
    height: 100%;
    object-fit: contain;
  }

  .sim-image img.sim-image-fill {
    object-fit: fill;
  }

  .sim-picker,
  .sim-native-picker-wrap {
    position: relative;
    display: inline-block;
  }

  .sim-picker-menu {
    position: absolute;
    z-index: 2;
    top: calc(100% + 4px);
    left: 0;
    min-width: 100%;
    max-height: 220px;
    overflow: auto;
    display: grid;
    border: 1px solid #c9cdd3;
    border-radius: 4px;
    background: #fff;
    box-shadow: 0 10px 24px rgba(15, 23, 42, 0.16);
  }

  .sim-picker-title {
    padding: 7px 9px;
    border-bottom: 1px solid #eef1f4;
    font-weight: 600;
    color: #202124;
    background: #f8fafc;
  }

  .sim-picker-filter {
    margin: 8px;
    min-height: 30px;
  }

  .sim-picker-menu button,
  .sim-listview button {
    min-height: 30px;
    border: 0;
    border-bottom: 1px solid #eef1f4;
    background: var(--picker-item-bg, #fff);
    color: var(--picker-item-color, #202124);
    text-align: left;
    padding: 6px 9px;
    font: inherit;
  }

  .sim-picker-menu button:hover,
  .sim-listview button:hover,
  .sim-picker-menu button.selected,
  .sim-listview button.selected {
    background: var(--picker-selection-bg, #e8f0fe);
  }

  .sim-picker-empty {
    min-height: 12px;
  }

  .sim-native-picker-input {
    position: absolute;
    left: 0;
    bottom: 0;
    width: 1px;
    height: 1px;
    opacity: 0;
    pointer-events: none;
  }

  .sim-listview {
    display: flex;
    flex-direction: column;
    min-height: 80px;
    max-height: 180px;
    overflow: hidden;
    border: 1px solid #d7dce2;
    background: #fff;
  }

  .sim-list-filter {
    margin: 8px;
    flex: 0 0 auto;
  }

  .sim-listview-items {
    display: grid;
    overflow: auto;
  }

  .sim-listview-items.horizontal {
    grid-auto-flow: column;
    grid-auto-columns: max-content;
    overflow-x: auto;
    overflow-y: hidden;
  }

  .sim-listview button {
    display: flex;
    align-items: center;
    gap: 9px;
  }

  .sim-listview button img {
    width: 36px;
    height: 36px;
    object-fit: cover;
    flex: 0 0 auto;
  }

  .sim-list-text {
    display: grid;
    gap: 2px;
    min-width: 0;
  }

  .sim-list-text small {
    color: inherit;
  }

  .sim-unsupported {
    outline: 1px dashed #d97706;
    outline-offset: 2px;
  }

  .sim-placeholder {
    min-height: 30px;
    padding: 6px 8px;
    border: 1px dashed #d97706;
    background: #fffbeb;
    color: #92400e;
    font-size: 12px;
    overflow-wrap: anywhere;
  }

  /* ── CircularProgress ────────────────────────────────────────── */
  .sim-circular-progress {
    display: inline-grid;
    place-items: center;
    min-width: 36px;
    min-height: 36px;
  }

  .sim-cp-ring {
    width: 28px;
    height: 28px;
    border: 3px solid rgba(0, 0, 255, 0.2);
    border-top-color: var(--cp-color, #0000ff);
    border-radius: 50%;
    animation: sim-cp-spin 0.9s linear infinite;
  }

  @keyframes sim-cp-spin {
    to { transform: rotate(360deg); }
  }

  @media (prefers-reduced-motion: reduce) {
    .sim-cp-ring { animation-duration: 2.4s; }
  }

  /* ── LinearProgress ──────────────────────────────────────────── */
  .sim-linear-progress {
    width: 100%;
    height: 6px;
    background: rgba(0, 0, 255, 0.15);
    border-radius: 3px;
    overflow: hidden;
    min-height: 6px;
    position: relative;
  }

  .sim-lp-bar {
    height: 100%;
    background: var(--lp-color, #0000ff);
    border-radius: 3px;
    width: var(--lp-pct, 0%);
    transition: width 0.2s;
  }

  .sim-linear-progress--indeterminate .sim-lp-bar {
    width: 40%;
    animation: sim-lp-slide 1.2s ease-in-out infinite;
    background: var(--lp-ind-color, #0000ff);
  }

  @keyframes sim-lp-slide {
    0% { transform: translateX(-100%); }
    100% { transform: translateX(300%); }
  }

  @media (prefers-reduced-motion: reduce) {
    .sim-linear-progress--indeterminate .sim-lp-bar { animation-duration: 3s; }
  }

  /* ── TableArrangement ────────────────────────────────────────── */
  .sim-table {
    display: grid;
    gap: 0;
    min-width: 0;
    min-height: 0;
  }

  /* ── WebViewer ───────────────────────────────────────────────── */
  .sim-webviewer {
    min-height: 80px;
    border: 1px solid #c9cdd3;
    border-radius: 3px;
    overflow: hidden;
    background: #fff;
    position: relative;
    display: flex;
    flex-direction: column;
  }

  .sim-webviewer-frame {
    flex: 1 1 auto;
    width: 100%;
    border: none;
    min-height: 0;
  }

  .sim-webviewer-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 6px;
    min-height: 80px;
    color: #9aa0a6;
    font-size: 11px;
  }

  .sim-webviewer-empty svg {
    width: 28px;
    height: 28px;
    opacity: 0.5;
  }

  /* ── VideoPlayer ─────────────────────────────────────────────── */
  .sim-videoplayer {
    min-height: 60px;
    background: #000;
    border-radius: 3px;
    overflow: hidden;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
  }

  .sim-videoplayer-video {
    width: 100%;
    height: 100%;
    display: block;
  }

  .sim-videoplayer-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    color: #9aa0a6;
    font-size: 11px;
    padding: 14px;
  }

  .sim-videoplayer-empty svg {
    width: 32px;
    height: 32px;
    opacity: 0.55;
    color: #fff;
  }

  /* ── Canvas ──────────────────────────────────────────────────── */
  .sim-canvas-wrap {
    position: relative;
    overflow: hidden;
  }

  .sim-canvas {
    display: block;
    width: 100%;
    height: 100%;
    touch-action: none;
    cursor: crosshair;
  }

  /* ── Chart ───────────────────────────────────────────────────── */
  .sim-chart {
    display: flex;
    flex-direction: column;
    min-height: 80px;
    background: #fff;
    border: 1px solid #e0e0e0;
    border-radius: 3px;
    overflow: hidden;
  }

  .sim-chart-svg {
    flex: 1 1 auto;
    width: 100%;
    height: 100%;
    cursor: pointer;
  }

  .sim-chart-grid-line {
    stroke: rgba(32, 33, 36, 0.12);
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
  }

  .sim-chart-axis-line {
    stroke: rgba(32, 33, 36, 0.5);
    stroke-width: 1;
    vector-effect: non-scaling-stroke;
  }

  .sim-chart-axis-text,
  .sim-chart-data-label {
    font: 9px/1 system-ui, sans-serif;
    paint-order: stroke;
    stroke: rgba(255, 255, 255, 0.88);
    stroke-width: 2px;
    vector-effect: non-scaling-stroke;
  }

  .sim-chart-data-label {
    font-size: 8px;
    pointer-events: none;
  }

  .sim-chart-description {
    padding: 3px 8px 4px;
    border-top: 1px solid #eee;
    color: #555;
    font-size: 10px;
    line-height: 1.25;
    text-align: center;
    overflow-wrap: anywhere;
  }

  .sim-chart-legend {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 12px;
    padding: 4px 8px;
    font-size: 10px;
    color: #555;
    border-top: 1px solid #eee;
  }

  .sim-chart-legend-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    margin-right: 3px;
    vertical-align: middle;
  }

  /* ── Map ─────────────────────────────────────────────────────── */
  .sim-map {
    min-height: 80px;
    border: 1px solid #c9cdd3;
    border-radius: 3px;
    z-index: 0;
  }

  :global(.sim-map-compass) {
    width: 34px;
    height: 34px;
    display: grid;
    place-items: center;
    border: 1px solid rgba(32, 33, 36, 0.22);
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.92);
    color: #202124;
    font: 700 12px/1 system-ui, sans-serif;
    box-shadow: 0 2px 8px rgba(15, 23, 42, 0.18);
  }

  :global(.sim-map-compass span) {
    display: block;
    transform: rotate(calc(-1 * var(--sim-map-rotation, 0deg)));
  }
</style>
