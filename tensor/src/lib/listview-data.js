export const LISTVIEW_LAYOUTS = Object.freeze([
  { value: '0', label: 'Single Text', columns: ['Text1'] },
  { value: '1', label: 'Two Text', columns: ['Text1', 'Text2'] },
  { value: '2', label: 'Two Text Linear', columns: ['Text1', 'Text2'] },
  { value: '3', label: 'Image, Single Text', columns: ['Text1', 'Image'] },
  { value: '4', label: 'Image, Two Text', columns: ['Text1', 'Text2', 'Image'] },
  { value: '5', label: 'Image Top, Two Text', columns: ['Text1', 'Text2', 'Image'] },
]);

const LISTVIEW_LAYOUT_BY_VALUE = new Map(LISTVIEW_LAYOUTS.map(layout => [layout.value, layout]));
const LISTVIEW_KEYS = ['Text1', 'Text2', 'Image'];

export function listViewLayoutDefinition(layoutValue = '0') {
  return LISTVIEW_LAYOUT_BY_VALUE.get(String(layoutValue)) || LISTVIEW_LAYOUTS[0];
}

export function listViewColumnsForLayout(layoutValue = '0') {
  return [...listViewLayoutDefinition(layoutValue).columns];
}

export function emptyListViewRow(layoutValue = '0') {
  const row = {};
  for (const column of listViewColumnsForLayout(layoutValue)) {
    row[column] = column === 'Image' ? 'None' : '';
  }
  return row;
}

function normalizeListViewRow(input, layoutValue = '0') {
  const source = input && typeof input === 'object' ? input : {};
  const row = {};
  for (const column of listViewColumnsForLayout(layoutValue)) {
    const fallback = column === 'Image' ? 'None' : '';
    row[column] = String(source[column] ?? fallback);
  }
  return row;
}

export function parseListViewData(value, layoutValue = '0') {
  if (!String(value || '').trim()) return [];
  let parsed;
  try {
    parsed = JSON.parse(value);
  } catch {
    return [];
  }
  if (!Array.isArray(parsed)) return [];
  return parsed.map(row => normalizeListViewRow(row, layoutValue));
}

export function serializeListViewData(rows = [], layoutValue = '0') {
  const normalized = Array.isArray(rows)
    ? rows.map(row => normalizeListViewRow(row, layoutValue))
    : [];
  return JSON.stringify(normalized);
}

export function listViewDataSummary(value, layoutValue = '0') {
  const rows = parseListViewData(value, layoutValue);
  const count = rows.length;
  if (!count) return 'No rows';
  return count === 1 ? '1 row' : `${count} rows`;
}

export function pruneListViewDataForLayout(value, layoutValue = '0') {
  return serializeListViewData(parseListViewData(value, layoutValue), layoutValue);
}

export function listViewDataAssetNames(value) {
  const names = new Set();
  let parsed;
  try {
    parsed = JSON.parse(value || '[]');
  } catch {
    return [];
  }
  if (!Array.isArray(parsed)) return [];
  for (const row of parsed) {
    const image = row && typeof row === 'object' ? String(row.Image || '') : '';
    if (image && image !== 'None') names.add(image);
  }
  return [...names];
}

export function renameListViewDataAsset(value, oldName, nextName) {
  let parsed;
  try {
    parsed = JSON.parse(value || '[]');
  } catch {
    return value || '';
  }
  if (!Array.isArray(parsed)) return value || '';
  let changed = false;
  const rows = parsed.map(row => {
    if (!row || typeof row !== 'object' || row.Image !== oldName) return row;
    changed = true;
    return { ...row, Image: nextName || 'None' };
  });
  return changed ? JSON.stringify(rows) : value || '';
}
