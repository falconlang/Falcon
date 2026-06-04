export function hasValue(value) {
  return value !== undefined && value !== null && value !== '';
}

export function firstNonEmpty(...values) {
  return values.find(value => hasValue(value) && String(value).trim() !== '') ?? '';
}

export function boolValue(value, fallback = false) {
  if (value === undefined || value === null || value === '') return fallback;
  if (typeof value === 'boolean') return value;
  if (typeof value === 'number') return value !== 0;
  const text = String(value).trim().toLowerCase();
  if (['true', '1', 'yes'].includes(text)) return true;
  if (['false', '0', 'no'].includes(text)) return false;
  return fallback;
}

export function numberOr(value, fallback = 0) {
  if (value === undefined || value === null || value === '') return fallback;
  const numberValue = Number(value);
  return Number.isFinite(numberValue) ? numberValue : fallback;
}

export function cssUrl(url) {
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

export function colorValue(value, fallback = 'transparent') {
  if (value === undefined || value === null || value === '') return fallback;
  if (typeof value === 'number' && Number.isFinite(value)) {
    const unsigned = value >>> 0;
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

export function sizeStyle(props, prop) {
  const value = props?.[prop];
  const cssProp = prop === 'Width' ? 'width' : 'height';
  if (value === -2 || value === '-2') return `${cssProp}: 100%;`;
  if (value === -1 || value === '-1' || value === undefined || value === null || value === '') return '';
  const numberValue = Number(value);
  if (!Number.isFinite(numberValue)) return '';
  if (numberValue <= -1000) return `${cssProp}: ${Math.max(0, -numberValue - 1000)}%;`;
  if (numberValue >= 0) return `${cssProp}: ${numberValue}px;`;
  return '';
}

export function positionStyle(props, parentType) {
  if (parentType !== 'AbsoluteArrangement') return '';
  return `position: absolute; left: ${numberOr(props?.Left, 0)}px; top: ${numberOr(props?.Top, 0)}px;`;
}

export function typefaceStyleFor(value) {
  const raw = String(value ?? '').trim();
  if (!raw || raw === '0' || raw.toLowerCase() === 'default') return '';
  const lower = raw.toLowerCase();
  if (raw === '1' || lower === 'serif') return 'font-family: Georgia, "Times New Roman", serif;';
  if (raw === '2' || lower === 'sans' || lower === 'sans-serif') return 'font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif;';
  if (raw === '3' || lower === 'monospace' || lower === 'mono') return 'font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;';
  return `font-family: "${raw.replace(/"/g, '\\"')}", inherit;`;
}

export function textAlignStyle(props) {
  if (!hasValue(props?.TextAlignment)) return '';
  const value = numberOr(props.TextAlignment, 0);
  if (value === 1) return 'text-align: center;';
  if (value === 2) return 'text-align: right;';
  return 'text-align: left;';
}

export function isVerticalLayout(type) {
  return ['Screen', 'Form', 'VerticalArrangement', 'VerticalScrollArrangement'].includes(type);
}

export function isHorizontalLayout(type) {
  return ['HorizontalArrangement', 'HorizontalScrollArrangement'].includes(type);
}

export function alignmentStyle(props, type) {
  if (!isVerticalLayout(type) && !isHorizontalLayout(type)) return '';

  const horizontal = numberOr(props?.AlignHorizontal, 1);
  const vertical = numberOr(props?.AlignVertical, 1);
  const horizontalCss = horizontal === 2 ? 'flex-end' : horizontal === 3 ? 'center' : 'flex-start';
  const verticalCss = vertical === 2 ? 'center' : vertical === 3 ? 'flex-end' : 'flex-start';

  if (isHorizontalLayout(type)) {
    return `justify-content: ${horizontalCss}; align-items: ${verticalCss};`;
  }
  return `align-items: ${horizontalCss}; justify-content: ${verticalCss};`;
}

export function isAutoSize(value) {
  return value === -1 || value === '-1' || value === undefined || value === null || value === '';
}

export function emptyArrangementStyle(node, props) {
  if (node?.type === 'Screen' || node?.type === 'Form') return '';
  if (!isVerticalLayout(node?.type) && !isHorizontalLayout(node?.type)) return '';
  if ((node?.children?.length ?? 0) > 0) return '';
  const rules = [];
  if (isAutoSize(props?.Width)) rules.push('min-width: 100px;');
  if (isAutoSize(props?.Height)) rules.push('min-height: 100px;');
  return rules.join(' ');
}

export function hintColorStyle(props) {
  return hasValue(props?.HintColor) ? `--sim-placeholder-color: ${colorValue(props.HintColor, '#9aa0a6')};` : '';
}

export function backgroundImageStyle(props, assetUrl) {
  if (!assetUrl || !firstNonEmpty(props?.Image, props?.BackgroundImage)) return '';
  return `background-image: url("${cssUrl(assetUrl)}"); background-repeat: no-repeat; background-position: center; background-size: cover;`;
}

export function assetName(props) {
  return firstNonEmpty(props?.Picture, props?.Image, props?.BackgroundImage);
}

export function shapeStyle(props) {
  const shape = String(props?.Shape ?? '0');
  if (shape === '1') return 'border-radius: 9px;';
  if (shape === '2') return 'border-radius: 0;';
  if (shape === '3') return 'border-radius: 999px;';
  return '';
}

export function baseStyle({ node, props, parentType, assetUrl = '' }, extra = '', options = {}) {
  const {
    size = true,
    position = true,
    typography = true,
    arrangement = true,
    backgroundImage = true,
  } = options;

  const rules = [
    size ? sizeStyle(props, 'Width') : '',
    size ? sizeStyle(props, 'Height') : '',
    hasValue(props?.BackgroundColor) ? `background-color: ${colorValue(props.BackgroundColor)};` : '',
    hasValue(props?.TextColor) ? `color: ${colorValue(props.TextColor, 'inherit')};` : '',
    typography && hasValue(props?.FontSize) ? `font-size: ${numberOr(props.FontSize, 14)}px;` : '',
    typography && boolValue(props?.FontBold, false) ? 'font-weight: 700;' : '',
    typography && boolValue(props?.FontItalic, false) ? 'font-style: italic;' : '',
    typography ? typefaceStyleFor(props?.FontTypeface) : '',
    typography ? textAlignStyle(props) : '',
    arrangement ? alignmentStyle(props, node?.type) : '',
    backgroundImage ? backgroundImageStyle(props, assetUrl) : '',
    position ? positionStyle(props, parentType) : '',
    extra,
  ];
  return rules.filter(Boolean).join(' ');
}

export function containerStyle(props, parentType, extra = '') {
  return [
    sizeStyle(props, 'Width'),
    sizeStyle(props, 'Height'),
    positionStyle(props, parentType),
    extra,
  ].filter(Boolean).join(' ');
}

export function buttonStyle(context, extra = '') {
  return baseStyle(context, `${shapeStyle(context.props)} ${extra}`);
}

export function buttonInnerStyle(context, extra = '') {
  return baseStyle(context, `${shapeStyle(context.props)} ${extra}`, {
    size: false,
    position: false,
    arrangement: false,
  });
}

export function styleContext(node, props, parentType, assetUrl = '') {
  return { node, props, parentType, assetUrl };
}
