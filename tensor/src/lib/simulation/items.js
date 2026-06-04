import { elementsFromString } from '../simulation-capabilities.js';

export function hasElements(value) {
  if (Array.isArray(value)) return value.length > 0;
  if (value && typeof value === 'object') return Object.keys(value).length > 0;
  return String(value ?? '').trim() !== '';
}

export function elementItems(props) {
  const source = hasElements(props?.Elements) ? props.Elements : props?.ElementsFromString;
  if (Array.isArray(source)) return source;
  if (source && typeof source === 'object') return Object.values(source);
  return elementsFromString(source);
}

export function itemMainText(item) {
  if (Array.isArray(item)) return String(item[0] ?? '');
  if (item && typeof item === 'object') {
    return String(
      item.Text1 ?? item.text1 ?? item.MainText ?? item.mainText ?? item.Text ?? item.text ?? Object.values(item)[0] ?? '',
    );
  }
  return String(item ?? '');
}

export function parseMaybeRow(value) {
  if (typeof value !== 'string') return value;
  const text = value.trim();
  if (!text.startsWith('{') && !text.startsWith('[')) return value;
  try {
    return JSON.parse(text);
  } catch {
    return value;
  }
}

export function normalizeListRow(item, index) {
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

export function listDataRows(props) {
  if (!String(props?.ListData ?? '').trim()) return [];
  try {
    const parsed = JSON.parse(props.ListData);
    if (!Array.isArray(parsed)) return [];
    return parsed.map((row, index) => normalizeListRow(row, index));
  } catch {
    return [];
  }
}

export function listViewRows(props) {
  const fromListData = listDataRows(props);
  if (fromListData.length) return fromListData;
  return elementItems(props).map((item, index) => normalizeListRow(item, index));
}

export function textIncludes(value, filter) {
  const needle = String(filter ?? '').trim().toLowerCase();
  if (!needle) return true;
  return String(value ?? '').toLowerCase().includes(needle);
}

export function filterIndexed(list, filter) {
  return list
    .map((text, index) => ({ text, index }))
    .filter(item => textIncludes(item.text, filter));
}

export function hasSelectableIndex(index, items) {
  return Number.isInteger(index) && index >= 0 && index < items.length;
}

export function selectionByIndex(index, items) {
  if (!hasSelectableIndex(index, items)) return null;
  return { selection: items[index] ?? '', selectionIndex: index + 1 };
}
