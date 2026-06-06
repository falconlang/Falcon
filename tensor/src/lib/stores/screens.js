import { writable, get } from 'svelte/store';
import { normalizeProjectProperties } from '../project-properties.js';
import {
  isValidAppInventorAssetName,
  normalizeAppInventorAssetName,
} from '../appinventor-validation.js';
import { setProjectExtensionComponentDescriptors } from '../appinventor-component-registry.js';
import {
  cells,
  designCode,
  designAssets,
  projectExtensionComponents,
  projectName,
  projectProperties,
  activeCellId,
  execCounter,
  lastRunAt,
  rawBlocklyXml,
  sourceScm,
  sourceDesignCode,
  sourceScmUpgradeWarnings,
  screenList,
  activeScreen,
  deletedCellUndoStack,
  deletedCellRedoStack,
} from './state.js';
import { clearDebugRuntimeState, disableDebugMode } from './debug.js';
import { clearCopiedCell } from './cells.js';

// ── Private screen persistence ──
const screenSavedStates = writable({});

// ── Internal helpers ──

function cloneCells(cellList) {
  return JSON.parse(JSON.stringify(cellList || []));
}

function uniqueNameFrom(baseName, existingNames) {
  const fallback = String(baseName || 'Screen1').trim() || 'Screen1';
  if (!existingNames.has(fallback)) return fallback;
  const match = fallback.match(/^(.*?)(\d+)$/);
  const stem = match ? match[1] : fallback;
  let n = match ? Number(match[2]) + 1 : 2;
  let next = `${stem}${n}`;
  while (existingNames.has(next)) {
    n += 1;
    next = `${stem}${n}`;
  }
  return next;
}

function currentScreenState() {
  return {
    cells: cloneCells(get(cells)),
    designCode: get(designCode),
    rawBlocklyXml: get(rawBlocklyXml),
    sourceScm: get(sourceScm),
    sourceDesignCode: get(sourceDesignCode),
    sourceScmUpgradeWarnings: Array.from(get(sourceScmUpgradeWarnings) || []),
  };
}

function stateForScreen(name, savedStates = get(screenSavedStates)) {
  const curr = get(activeScreen);
  if (name === curr) return currentScreenState();
  const saved = savedStates[name];
  return {
    cells: cloneCells(saved?.cells || []),
    designCode: saved?.designCode || '',
    rawBlocklyXml: saved?.rawBlocklyXml || '',
    sourceScm: saved?.sourceScm || '',
    sourceDesignCode: saved?.sourceDesignCode || '',
    sourceScmUpgradeWarnings: Array.from(saved?.sourceScmUpgradeWarnings || []),
  };
}

function applyScreenState(state) {
  const nextCells = cloneCells(state?.cells || []);
  clearDebugRuntimeState();
  deletedCellUndoStack.set([]);
  deletedCellRedoStack.set([]);
  cells.set(nextCells);
  designCode.set(state?.designCode || '');
  rawBlocklyXml.set(state?.rawBlocklyXml || '');
  sourceScm.set(state?.sourceScm || '');
  sourceDesignCode.set(state?.sourceDesignCode || '');
  sourceScmUpgradeWarnings.set(Array.from(state?.sourceScmUpgradeWarnings || []));
  activeCellId.set(nextCells[0]?.id || null);
}

function replaceDesignAssets(nextAssets) {
  for (const asset of get(designAssets)) {
    if (typeof asset !== 'string' && asset.url && typeof URL !== 'undefined') {
      URL.revokeObjectURL(asset.url);
    }
  }
  designAssets.set(nextAssets || []);
}

// ── Screen management ──

export function switchScreen(name) {
  const curr = get(activeScreen);
  if (name === curr) return;
  screenSavedStates.update(s => ({
    ...s,
    [curr]: currentScreenState(),
  }));
  const saved = get(screenSavedStates)[name];
  applyScreenState(saved || {});
  activeScreen.set(name);
}

export function addScreen(name) {
  const curr = get(activeScreen);
  const list = get(screenList);
  screenSavedStates.update(s => ({
    ...s,
    [curr]: currentScreenState(),
  }));
  let newName = name;
  if (!newName) {
    let n = list.length + 1;
    const existing = new Set(list);
    while (existing.has(`Screen${n}`)) n++;
    newName = `Screen${n}`;
  } else {
    newName = uniqueNameFrom(newName, new Set(list));
  }
  screenList.update(l => [...l, newName]);
  applyScreenState({});
  activeScreen.set(newName);
}

export function removeScreen(name) {
  if (name === 'Screen1') return;
  const curr = get(activeScreen);
  screenList.update(l => l.filter(s => s !== name));
  screenSavedStates.update(s => {
    const next = { ...s };
    delete next[name];
    return next;
  });
  if (curr === name) {
    const saved = get(screenSavedStates)['Screen1'];
    applyScreenState(saved || currentScreenState());
    activeScreen.set('Screen1');
  }
}

export function getProjectSnapshot() {
  const saved = {
    ...get(screenSavedStates),
    [get(activeScreen)]: currentScreenState(),
  };
  const screens = get(screenList).map(name => ({
    name,
    ...stateForScreen(name, saved),
  }));

  return {
    projectName: get(projectName),
    projectProperties: normalizeProjectProperties(get(projectProperties)),
    activeScreen: get(activeScreen),
    screens,
    assets: get(designAssets),
    extensionComponents: get(projectExtensionComponents),
  };
}

export function loadProjectState(project) {
  const inputScreens = Array.isArray(project?.screens) && project.screens.length
    ? project.screens
    : [{ name: 'Screen1', cells: [], designCode: '' }];
  const usedNames = new Set();
  const screens = inputScreens.map(screen => {
    const name = uniqueNameFrom(screen.name || 'Screen1', usedNames);
    usedNames.add(name);
    return { ...screen, name };
  });
  const names = screens.map(screen => screen.name);
  const active = names.includes(project?.activeScreen) ? project.activeScreen : names[0];
  const saved = {};

  for (const screen of screens) {
    const name = screen.name || 'Screen1';
    saved[name] = {
      cells: cloneCells(screen.cells || []),
      designCode: screen.designCode || '',
      rawBlocklyXml: screen.rawBlocklyXml || '',
      sourceScm: screen.sourceScm || '',
      sourceDesignCode: screen.sourceDesignCode || '',
      sourceScmUpgradeWarnings: Array.from(screen.sourceScmUpgradeWarnings || []),
    };
  }

  projectName.set(project?.projectName || 'ImportedProject');
  projectProperties.set(normalizeProjectProperties(project?.projectProperties || {}));
  projectExtensionComponents.set(project?.extensionComponents || []);
  setProjectExtensionComponentDescriptors(project?.extensionComponents || []);
  screenList.set(names);
  screenSavedStates.set(saved);
  activeScreen.set(active);
  applyScreenState(saved[active]);
  replaceDesignAssets(project?.assets || []);
  execCounter.set(1);
  lastRunAt.set(null);
  disableDebugMode();
  clearCopiedCell();
}

// ── Asset management ──
// Lives here because replaceAssetReferences must update screenSavedStates
// for non-active screens as well as the live stores.

function assetNameFrom(input) {
  return typeof input === 'string' ? input : input?.name;
}

function splitAssetName(name) {
  const dot = name.lastIndexOf('.');
  if (dot <= 0) return { base: name, ext: '' };
  return { base: name.slice(0, dot), ext: name.slice(dot) };
}

function uniqueAssetName(name, assets, exceptId = null) {
  const clean = normalizeAppInventorAssetName(name);
  if (!clean || !isValidAppInventorAssetName(clean)) return '';
  const existing = new Set(
    assets
      .filter(asset => (typeof asset === 'string' ? asset : asset.id) !== exceptId)
      .map(assetNameFrom)
  );
  if (!existing.has(clean)) return clean;

  const { base, ext } = splitAssetName(clean);
  let n = 2;
  let next = `${base}_${n}${ext}`;
  while (existing.has(next)) {
    n += 1;
    next = `${base}_${n}${ext}`;
  }
  return next;
}

function xmlEscapeText(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function replaceLiteral(text, before, after) {
  if (!before || typeof text !== 'string' || !text.includes(before)) return text;
  return text.split(before).join(after);
}

function replaceAssetNameInText(text, oldName, nextName) {
  if (typeof text !== 'string' || !oldName) return text;
  const rawOld = String(oldName);
  const rawNext = String(nextName || '');
  const pairs = [
    [rawOld, rawNext],
    [JSON.stringify(rawOld).slice(1, -1), JSON.stringify(rawNext).slice(1, -1)],
    [xmlEscapeText(rawOld), xmlEscapeText(rawNext)],
  ];
  const seen = new Set();
  let out = text;
  for (const [before, after] of pairs) {
    if (seen.has(before)) continue;
    seen.add(before);
    out = replaceLiteral(out, before, after);
  }
  return out;
}

function replaceAssetNameInCells(cellList, oldName, nextName) {
  return JSON.parse(JSON.stringify(cellList || [])).map(cell => (
    cell?.type === 'code'
      ? { ...cell, code: replaceAssetNameInText(cell.code || '', oldName, nextName) }
      : cell
  ));
}

function replaceAssetNameInScreenState(state, oldName, nextName) {
  return {
    ...state,
    cells: replaceAssetNameInCells(state?.cells || [], oldName, nextName),
    designCode: replaceAssetNameInText(state?.designCode || '', oldName, nextName),
    rawBlocklyXml: replaceAssetNameInText(state?.rawBlocklyXml || '', oldName, nextName),
    sourceScm: replaceAssetNameInText(state?.sourceScm || '', oldName, nextName),
    sourceDesignCode: replaceAssetNameInText(state?.sourceDesignCode || '', oldName, nextName),
    sourceScmUpgradeWarnings: Array.from(state?.sourceScmUpgradeWarnings || []),
  };
}

function replaceAssetReferences(oldName, nextName) {
  if (!oldName) return;
  projectProperties.update(properties => {
    const normalized = normalizeProjectProperties(properties);
    if (normalized.Icon !== oldName) return properties;
    return { ...normalized, Icon: nextName || '' };
  });
  clearDebugRuntimeState();
  cells.update(cellList => replaceAssetNameInCells(cellList, oldName, nextName));
  designCode.update(code => replaceAssetNameInText(code, oldName, nextName));
  rawBlocklyXml.update(xml => replaceAssetNameInText(xml, oldName, nextName));
  sourceScm.update(text => replaceAssetNameInText(text, oldName, nextName));
  sourceDesignCode.update(code => replaceAssetNameInText(code, oldName, nextName));
  screenSavedStates.update(states => {
    const next = {};
    for (const [screen, state] of Object.entries(states || {})) {
      next[screen] = replaceAssetNameInScreenState(state, oldName, nextName);
    }
    return next;
  });
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
    blob: isFile ? fileOrName : null,
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
    const name = normalizeAppInventorAssetName(assetNameFrom(fileOrName));
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
        ? { id: existing, name: existing, size: 0, type: '', blob: null, url: '' }
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
  let oldName = '';
  designAssets.update(assets => {
    const clean = uniqueAssetName(nextName, assets, assetId);
    if (!clean) return assets;
    return assets.map(asset => {
      const id = typeof asset === 'string' ? asset : asset.id;
      if (id !== assetId && assetNameFrom(asset) !== assetId) return asset;
      oldName = assetNameFrom(asset);
      renamed = typeof asset === 'string'
        ? { id: clean, name: clean, size: 0, type: '', blob: null, url: '' }
        : { ...asset, name: clean };
      return renamed;
    });
  });
  if (renamed?.name && oldName) replaceAssetReferences(oldName, renamed.name);
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
  const removedName = assetNameFrom(removed);
  if (removedName) replaceAssetReferences(removedName, '');
  return typeof removed === 'string'
    ? { id: removed, name: removed, size: 0, type: '', blob: null, url: '' }
    : removed;
}
