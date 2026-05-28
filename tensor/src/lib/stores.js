import { writable, get } from 'svelte/store';

export const initialDesignCode = `Screen.Screen1 { Title: "Calculator",
  Label { Text: "First number: " }
  TextBox.firstNumberTextBox { NumbersOnly: true, Hint: "Enter first number" }

  Label { Text: "Second number: " }
  TextBox.secondNumberTextBox { NumbersOnly: true, Hint: "Enter second number" }

  HorizontalArrangement {
    Button.AddButton { Text: "+" }
    Button.SubtractButton { Text: "-" }
    Button.MultiplyButton { Text: "*" }
    Button.DivideButton { Text: "/" }
  }

  Notifier.Notifier1
}`;

const initialCells = [
  {
    id: 'c1', type: 'code', execCount: 1,
    code: `func checkTextBoxes() = {
  if (!(firstNumberTextBox.Text ? number) || !(secondNumberTextBox.Text ? number)) {
    Notifier1.ShowAlert("Please enter numeric values in the textbox!")
    yield false
  }
  true
}`,
  },
  {
    id: 'c2', type: 'code', execCount: 2,
    code: `when AddButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text + secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'c3', type: 'code', execCount: 3,
    code: `when SubtractButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text - secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'c4', type: 'code', execCount: 4,
    code: `when MultiplyButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text * secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
  {
    id: 'c5', type: 'code', execCount: 5,
    code: `when DivideButton.Click {
  if (checkTextBoxes()) {
    firstNumberTextBox.Text = firstNumberTextBox.Text / secondNumberTextBox.Text
    secondNumberTextBox.Text = ""
  }
}`,
  },
];

export const cells = writable(initialCells);
export const designCode = writable(initialDesignCode);
export const designAssets = writable([]);
export const activeCellId = writable('c1');
export const execCounter = writable(6);
export const ctxMenu = writable({ show: false, x: 0, y: 0, cellId: null });
export const liveTestOpen = writable(false);
export const liveTestState = writable({
  status: 'idle',
  code: null,
  error: null,
  messageCount: 0,
});
export const doItCellId = writable(null);
export const sidebarVisible = writable(true);
export const debugCollapsed = writable(true);
export const debugOpenHeight = writable(200);
export const debugLogs = writable([]);

let debugLogId = 0;

function debugTimestamp(date = new Date()) {
  const pad = n => String(n).padStart(2, '0');
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}.${String(date.getMilliseconds()).padStart(3, '0')}`;
}

export function normalizeDebugLevel(level) {
  const normalized = String(level || 'info').toLowerCase();
  if (normalized === 'high' || normalized === 'critical' || normalized === 'fatal') return 'high';
  if (normalized === 'error') return 'error';
  if (normalized === 'warning' || normalized === 'warn') return 'warn';
  return 'info';
}

function stringifyDebugValue(value) {
  if (value == null) return '';
  if (typeof value === 'string') return value;
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

export function extractDebugLogsFromCompanionResponse(data) {
  const payload = typeof data === 'string' ? JSON.parse(data) : data;
  const values = Array.isArray(payload?.values)
    ? payload.values
    : Array.isArray(payload) ? payload : [];

  return values
    .map(value => {
      if (value?.type === 'log') {
        return {
          level: normalizeDebugLevel(value.level),
          message: stringifyDebugValue(value.item ?? value.value),
          source: stringifyDebugValue(value.blockid || 'Notifier'),
          status: value.status || null,
        };
      }

      if (value?.type === 'error') {
        return {
          level: 'high',
          message: stringifyDebugValue(value.value ?? value.item),
          source: stringifyDebugValue(value.blockid || 'Runtime'),
          status: value.status || null,
        };
      }

      return null;
    })
    .filter(Boolean);
}

export function appendDebugLogs(entries) {
  const nextEntries = entries
    .map(entry => ({
      id: ++debugLogId,
      time: debugTimestamp(),
      level: normalizeDebugLevel(entry.level),
      source: entry.source || 'Notifier',
      message: stringifyDebugValue(entry.message),
      status: entry.status || null,
    }))
    .filter(entry => entry.message.length > 0);

  if (!nextEntries.length) return [];
  debugLogs.update(logs => [...logs, ...nextEntries].slice(-500));
  return nextEntries;
}

export function appendDebugLogsFromCompanionResponse(data) {
  try {
    return appendDebugLogs(extractDebugLogsFromCompanionResponse(data));
  } catch {
    return [];
  }
}

export function clearDebugLogs() {
  debugLogs.set([]);
}

export function setActive(id) {
  activeCellId.set(id);
}

export function addCodeCell() {
  const id = 'c' + Date.now();
  const activeId = get(activeCellId);
  const currentCells = get(cells);
  const idx = activeId ? currentCells.findIndex(c => c.id === activeId) + 1 : currentCells.length;
  cells.update(cs => {
    const next = [...cs];
    next.splice(idx, 0, { id, type: 'code', code: '', execCount: null });
    return next;
  });
  setActive(id);
  return id;
}

export function addMarkdownCell() {
  const id = 'c' + Date.now();
  const activeId = get(activeCellId);
  const currentCells = get(cells);
  const idx = activeId ? currentCells.findIndex(c => c.id === activeId) + 1 : currentCells.length;
  cells.update(cs => {
    const next = [...cs];
    next.splice(idx, 0, { id, type: 'markdown', content: '<div class="md-p">New text cell</div>' });
    return next;
  });
  setActive(id);
  return id;
}

export function moveCellById(id, dir) {
  cells.update(cs => {
    const idx = cs.findIndex(c => c.id === id);
    const newIdx = idx + dir;
    if (newIdx < 0 || newIdx >= cs.length) return cs;
    const next = [...cs];
    [next[idx], next[newIdx]] = [next[newIdx], next[idx]];
    return next;
  });
  setActive(id);
}

export function deleteCellById(id) {
  const currentCells = get(cells);
  if (currentCells.length <= 1) return;
  const idx = currentCells.findIndex(c => c.id === id);
  cells.update(cs => cs.filter(c => c.id !== id));
  const nextCells = get(cells);
  const nextActive = nextCells[Math.min(idx, nextCells.length - 1)]?.id || null;
  if (nextActive) setActive(nextActive);
}

export function updateCellExecCount(id) {
  const count = get(execCounter);
  cells.update(cs => cs.map(c => c.id === id ? { ...c, execCount: count } : c));
  execCounter.update(n => n + 1);
}

export function updateCellCode(id, code) {
  cells.update(cs => cs.map(c => c.id === id ? { ...c, code } : c));
}

export function updateDesignCode(code) {
  designCode.set(code);
}

function assetNameFrom(input) {
  return typeof input === 'string' ? input : input?.name;
}

function normalizeAssetName(name) {
  return String(name || '').trim().replace(/[\\/]+/g, '-');
}

function splitAssetName(name) {
  const dot = name.lastIndexOf('.');
  if (dot <= 0) return { base: name, ext: '' };
  return { base: name.slice(0, dot), ext: name.slice(dot) };
}

function uniqueAssetName(name, assets, exceptId = null) {
  const clean = normalizeAssetName(name);
  if (!clean) return '';
  const existing = new Set(
    assets
      .filter(asset => (typeof asset === 'string' ? asset : asset.id) !== exceptId)
      .map(assetNameFrom)
  );
  if (!existing.has(clean)) return clean;

  const { base, ext } = splitAssetName(clean);
  let n = 2;
  let next = `${base} ${n}${ext}`;
  while (existing.has(next)) {
    n += 1;
    next = `${base} ${n}${ext}`;
  }
  return next;
}

function createAssetRecord(fileOrName, name, existingAssets) {
  const isFile = typeof File !== 'undefined' && fileOrName instanceof File;
  const cleanName = uniqueAssetName(name, existingAssets);
  if (!cleanName) return null;
  const id = `asset-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`;
  const record = {
    id,
    name: cleanName,
    size: isFile ? fileOrName.size : 0,
    type: isFile ? fileOrName.type : '',
    url: '',
  };
  if (isFile && typeof URL !== 'undefined') {
    record.url = URL.createObjectURL(fileOrName);
  }
  return record;
}

export function addDesignAsset(fileOrName) {
  let added = null;
  designAssets.update(assets => {
    const name = normalizeAssetName(assetNameFrom(fileOrName));
    if (!name) return assets;
    const isFile = typeof File !== 'undefined' && fileOrName instanceof File;
    const existingIndex = assets.findIndex(asset => assetNameFrom(asset) === name);
    const existing = existingIndex === -1 ? null : assets[existingIndex];
    if (existing) {
      if (typeof existing === 'string' && isFile) {
        added = createAssetRecord(fileOrName, name, assets.filter((_, index) => index !== existingIndex));
        if (!added) return assets;
        return assets.map((asset, index) => index === existingIndex ? added : asset);
      }
      added = typeof existing === 'string'
        ? { id: existing, name: existing, size: 0, type: '', url: '' }
        : existing;
      return assets;
    }
    added = createAssetRecord(fileOrName, name, assets);
    return added ? [...assets, added] : assets;
  });
  return added;
}

export function renameDesignAsset(assetId, nextName) {
  let renamed = null;
  designAssets.update(assets => {
    const clean = uniqueAssetName(nextName, assets, assetId);
    if (!clean) return assets;
    return assets.map(asset => {
      const id = typeof asset === 'string' ? asset : asset.id;
      if (id !== assetId && assetNameFrom(asset) !== assetId) return asset;
      renamed = typeof asset === 'string'
        ? { id: clean, name: clean, size: 0, type: '', url: '' }
        : { ...asset, name: clean };
      return renamed;
    });
  });
  return renamed;
}

export function deleteDesignAsset(assetId) {
  let removed = null;
  designAssets.update(assets => {
    const next = [];
    for (const asset of assets) {
      const id = typeof asset === 'string' ? asset : asset.id;
      if (id === assetId || assetNameFrom(asset) === assetId) {
        removed = asset;
        if (typeof asset !== 'string' && asset.url && typeof URL !== 'undefined') {
          URL.revokeObjectURL(asset.url);
        }
      } else {
        next.push(asset);
      }
    }
    return next;
  });
  return typeof removed === 'string'
    ? { id: removed, name: removed, size: 0, type: '', url: '' }
    : removed;
}

export function getFalconSource() {
  return get(cells)
    .filter(cell => cell.type === 'code')
    .map(cell => cell.code || '')
    .join('\n\n');
}

export function getDesignSource() {
  return get(designCode);
}

export function showCtx(e, id) {
  e.preventDefault();
  e.stopPropagation();
  setActive(id);
  ctxMenu.set({
    show: true,
    x: Math.min(e.clientX, window.innerWidth - 180),
    y: Math.min(e.clientY, window.innerHeight - 260),
    cellId: id,
  });
}

export function hideCtx() {
  ctxMenu.update(m => ({ ...m, show: false, cellId: null }));
}
