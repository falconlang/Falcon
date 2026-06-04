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
  const CANVAS_PAD_PX = 0;

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
      x: Math.round((e.clientX - rect.left) * scaleX),
      y: Math.round((e.clientY - rect.top) * scaleY),
    };
  }

  function canvasPointerDown(e) {
    if (!enabled) return;
    canvasEl?.setPointerCapture(e.pointerId);
    const pt = canvasPoint(e);
    canvasDragStart = pt;
    canvasPrev = pt;
    canvasTouchStart = pt;
    canvasTouchTime = Date.now();
    emitEvent(node.name, 'TouchDown', [pt.x, pt.y]);
  }

  function canvasPointerMove(e) {
    if (!enabled || !canvasDragStart) return;
    const pt = canvasPoint(e);
    emitEvent(node.name, 'Dragged', [
      canvasDragStart.x, canvasDragStart.y,
      canvasPrev.x, canvasPrev.y,
      pt.x, pt.y,
      false,
    ]);
    canvasPrev = pt;
  }

  function canvasPointerUp(e) {
    if (!enabled) return;
    const pt = canvasPoint(e);
    emitEvent(node.name, 'TouchUp', [pt.x, pt.y]);
    const elapsed = Date.now() - (canvasTouchTime || 0);
    const dx = pt.x - (canvasTouchStart?.x ?? pt.x);
    const dy = pt.y - (canvasTouchStart?.y ?? pt.y);
    const dist = Math.hypot(dx, dy);
    if (dist < numberOr(props.TapThreshold, 15)) {
      emitEvent(node.name, 'Touched', [pt.x, pt.y, false]);
    } else if (elapsed > 0) {
      const speed = dist / elapsed * 1000;
      const heading = (Math.atan2(-dy, dx) * 180 / Math.PI + 360) % 360;
      emitEvent(node.name, 'Flung', [pt.x, pt.y, speed, heading, dx / elapsed * 1000, -dy / elapsed * 1000, false]);
    }
    canvasDragStart = null;
    canvasPrev = null;
  }

  function canvasPointerCancel() {
    canvasDragStart = null;
    canvasPrev = null;
  }

  // Re-render canvas drawing ops when state changes
  $: if (canvasEl && state) applyCanvasOps();

  function applyCanvasOps() {
    const ctx = getCanvas();
    if (!ctx || !canvasEl) return;
    // Clear + background
    ctx.clearRect(0, 0, canvasEl.width, canvasEl.height);
    const bg = colorValue(props.BackgroundColor, '#ffffff');
    if (bg && bg !== 'transparent') {
      ctx.fillStyle = bg;
      ctx.fillRect(0, 0, canvasEl.width, canvasEl.height);
    }
    // Background image
    if (assetUrl && props.BackgroundImage) {
      const img = new Image();
      img.onload = () => { ctx.drawImage(img, 0, 0, canvasEl.width, canvasEl.height); };
      img.src = assetUrl;
    }
    // Draw sprites (Ball/ImageSprite) on canvas
    for (const sprite of canvasSprites()) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false) continue;
      ctx.save();
      if (sprite.type === 'Ball') {
        const r = numberOr(sp.Radius, 5);
        const cx = numberOr(sp.X, 0) + (boolValue(sp.OriginAtCenter, false) ? 0 : r);
        const cy = numberOr(sp.Y, 0) + (boolValue(sp.OriginAtCenter, false) ? 0 : r);
        ctx.fillStyle = colorValue(sp.PaintColor, '#000000');
        ctx.beginPath();
        ctx.arc(cx, cy, r, 0, Math.PI * 2);
        ctx.fill();
      } else if (sprite.type === 'ImageSprite') {
        const spUrl = resolveAssetUrl(assets, sp.Picture);
        const w = numberOr(sp.Width, 0);
        const h = numberOr(sp.Height, 0);
        const x = numberOr(sp.X, 0);
        const y = numberOr(sp.Y, 0);
        if (spUrl && w > 0 && h > 0) {
          const img = new Image();
          img.onload = () => { ctx.drawImage(img, x, y, w, h); };
          img.src = spUrl;
        }
      }
      ctx.restore();
    }
  }

  // Canvas draw op effects: the Go host emits 'canvas-draw' effects
  $: if (node?.type === 'Canvas') handleCanvasEffects();
  let lastCanvasDrawSeq = 0;
  function handleCanvasEffects() { /* ops arrive via applyCanvasOps reactive */ }

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
        return {
          label: sp.Label || c.name,
          color: colorValue(sp.Color, CHART_COLORS[i % CHART_COLORS.length]),
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

  function chartXRange() {
    const all = chartDataSeries().flatMap(s => s.points.map(p => p[0]));
    if (!all.length) return [0, 1];
    const min = boolValue(props.XFromZero, false) ? 0 : Math.min(...all);
    return [min, Math.max(...all, min + 1)];
  }

  function chartYRange() {
    const all = chartDataSeries().flatMap(s => s.points.map(p => p[1]));
    if (!all.length) return [0, 1];
    const min = boolValue(props.YFromZero, false) ? 0 : Math.min(...all);
    return [min, Math.max(...all, min + 1)];
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
    const r = Math.min(cx, cy) - CHART_PAD;
    let angle = -Math.PI / 2;
    return pts.map((p, i) => {
      const sweep = (Math.abs(p[1]) / total) * Math.PI * 2;
      const x1 = cx + r * Math.cos(angle);
      const y1 = cy + r * Math.sin(angle);
      const x2 = cx + r * Math.cos(angle + sweep);
      const y2 = cy + r * Math.sin(angle + sweep);
      const large = sweep > Math.PI ? 1 : 0;
      const d = `M${cx},${cy}L${x1},${y1}A${r},${r},0,${large},1,${x2},${y2}Z`;
      const fill = CHART_COLORS[(si * pts.length + i) % CHART_COLORS.length];
      angle += sweep;
      return { d, fill };
    });
  }

  // ── Map helpers (Leaflet) ───────────────────────────────────────────────
  let mapEl;
  let mapInstance = null;
  let mapLayers = {};

  // Map setup runs reactively via the $: if (mapEl) initOrUpdateMap() block below

  $: if (node?.type === 'Map' && mapEl) initOrUpdateMap();

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
        zoomControl: boolValue(props.ShowZoom, false),
        dragging: boolValue(props.EnablePan, true),
        scrollWheelZoom: boolValue(props.EnableZoom, true),
        touchZoom: boolValue(props.EnableZoom, true),
        doubleClickZoom: boolValue(props.EnableZoom, true),
        attributionControl: true,
      });
      const tileUrl = props.CustomUrl || 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
      L.tileLayer(tileUrl, { attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>' }).addTo(mapInstance);
      mapInstance.on('moveend', () => emitEvent(node.name, 'BoundsChange'));
      mapInstance.on('zoomend', () => emitEvent(node.name, 'ZoomChange'));
      mapInstance.on('click', (e) => emitEvent(node.name, 'TapAtPoint', [e.latlng.lat, e.latlng.lng]));
      mapInstance.on('dblclick', (e) => emitEvent(node.name, 'DoubleTapAtPoint', [e.latlng.lat, e.latlng.lng]));
      mapInstance.on('contextmenu', (e) => emitEvent(node.name, 'LongPressAtPoint', [e.latlng.lat, e.latlng.lng]));
      emitEvent(node.name, 'Ready');
    } else {
      mapInstance.setView(
        [numberOr(props.Latitude, 42.359144), numberOr(props.Longitude, -71.093612)],
        numberOr(props.ZoomLevel, 13),
        { animate: false },
      );
    }
    updateMapFeatures(L);
  }

  function updateMapFeatures(L) {
    if (!mapInstance || !L) return;
    const nextKeys = new Set();
    for (const child of node?.children || []) {
      const sp = state?.[child.name] ?? {};
      if (sp.Visible === false) continue;
      const key = child.name;
      nextKeys.add(key);
      if (mapLayers[key]) {
        updateMapLayer(L, child, sp, mapLayers[key]);
      } else {
        mapLayers[key] = createMapLayer(L, child, sp);
        if (mapLayers[key]) mapLayers[key].addTo(mapInstance);
      }
    }
    for (const k of Object.keys(mapLayers)) {
      if (!nextKeys.has(k)) { mapLayers[k]?.remove(); delete mapLayers[k]; }
    }
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

  function createMapLayer(L, child, sp) {
    switch (child.type) {
      case 'Marker': {
        const icon = L.divIcon({
          html: `<svg viewBox="0 0 24 36" xmlns="http://www.w3.org/2000/svg" width="24" height="36"><path d="M12 0C5.4 0 0 5.4 0 12c0 9 12 24 12 24s12-15 12-24c0-6.6-5.4-12-12-12z" fill="${colorValue(sp.FillColor,'#ff0000')}" stroke="${colorValue(sp.StrokeColor,'#000')}"/><circle cx="12" cy="12" r="5" fill="white"/></svg>`,
          className: '',
          iconSize: [24, 36],
          iconAnchor: [12, 36],
          popupAnchor: [0, -36],
        });
        const m = L.marker([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)], { icon, draggable: boolValue(sp.Draggable, false) });
        if (sp.Title || sp.Description) m.bindPopup(`<b>${sp.Title || ''}</b><br>${sp.Description || ''}`);
        m.on('click', () => emitEvent(node.name, 'FeatureClick', [child.name]));
        m.on('contextmenu', () => emitEvent(node.name, 'FeatureLongClick', [child.name]));
        return m;
      }
      case 'Circle': {
        const c = L.circle([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)], { radius: numberOr(sp.Radius, 10), ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        c.on('click', () => emitEvent(node.name, 'FeatureClick', [child.name]));
        c.on('contextmenu', () => emitEvent(node.name, 'FeatureLongClick', [child.name]));
        return c;
      }
      case 'LineString': {
        const pts = parseLatLngList(sp.PointsFromString);
        if (!pts.length) return null;
        const l = L.polyline(pts, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        l.on('click', () => emitEvent(node.name, 'FeatureClick', [child.name]));
        return l;
      }
      case 'Polygon': {
        const pts = parseLatLngList(sp.PointsFromString);
        if (!pts.length) return null;
        const pg = L.polygon(pts, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        pg.on('click', () => emitEvent(node.name, 'FeatureClick', [child.name]));
        return pg;
      }
      case 'Rectangle': {
        const bounds = [
          [numberOr(sp.SouthLatitude, 0), numberOr(sp.WestLongitude, 0)],
          [numberOr(sp.NorthLatitude, 0), numberOr(sp.EastLongitude, 0)],
        ];
        const r = L.rectangle(bounds, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        r.on('click', () => emitEvent(node.name, 'FeatureClick', [child.name]));
        return r;
      }
      default: return null;
    }
  }

  function updateMapLayer(L, child, sp, layer) {
    if (child.type === 'Marker') {
      layer.setLatLng([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)]);
    } else if (child.type === 'Circle') {
      layer.setLatLng([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)]);
      layer.setRadius(numberOr(sp.Radius, 10));
      layer.setStyle(featureStyle(sp));
    } else if (child.type === 'LineString') {
      const pts = parseLatLngList(sp.PointsFromString);
      if (pts.length) layer.setLatLngs(pts);
      layer.setStyle(featureStyle(sp));
    } else if (child.type === 'Polygon') {
      const pts = parseLatLngList(sp.PointsFromString);
      if (pts.length) layer.setLatLngs(pts);
      layer.setStyle(featureStyle(sp));
    } else if (child.type === 'Rectangle') {
      const bounds = [
        [numberOr(sp.SouthLatitude, 0), numberOr(sp.WestLongitude, 0)],
        [numberOr(sp.NorthLatitude, 0), numberOr(sp.EastLongitude, 0)],
      ];
      layer.setBounds(bounds);
      layer.setStyle(featureStyle(sp));
    }
  }

  function parseLatLngList(value) {
    const text = String(value ?? '').trim();
    if (!text) return [];
    try {
      const parsed = JSON.parse(text);
      if (Array.isArray(parsed)) return parsed.map(p => Array.isArray(p) ? [Number(p[0]), Number(p[1])] : [0, 0]);
    } catch {}
    const nums = text.split(/[\s,]+/).map(Number).filter(Number.isFinite);
    const pts = [];
    for (let i = 0; i + 1 < nums.length; i += 2) pts.push([nums[i], nums[i + 1]]);
    return pts;
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
        {#each chartDataSeries() as series, si}
          {#if numberOr(props.Type, 0) === 4}
            <!-- Pie chart -->
            {#each pieSectors(series.points, si) as sector}
              <path d={sector.d} fill={sector.fill} stroke="white" stroke-width="1" />
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
                fill={series.color}
                on:click={() => emitEvent(node.name, 'EntryClick', [series.label, pt[0], pt[1]])}
                on:keydown={(e) => e.key === 'Enter' && emitEvent(node.name, 'EntryClick', [series.label, pt[0], pt[1]])}
              />
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
            {#each series.points as pt}
              <circle
                role="button"
                tabindex="0"
                cx={chartX(pt[0])}
                cy={chartY(pt[1])}
                r="4"
                fill={series.color}
                on:click={() => emitEvent(node.name, 'EntryClick', [series.label, pt[0], pt[1]])}
                on:keydown={(e) => e.key === 'Enter' && emitEvent(node.name, 'EntryClick', [series.label, pt[0], pt[1]])}
              />
            {/each}
          {/if}
        {/each}
      </svg>
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
</style>
