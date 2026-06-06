import { get } from 'svelte/store';
import {
  normalizeProjectPropertyName,
  normalizeProjectProperties,
  withProjectPropertyValue,
  defaultProjectProperties,
} from '../project-properties.js';
import {
  cells,
  designCode,
  projectProperties,
  activeCellId,
  execCounter,
  lastRunAt,
  ctxMenu,
  doItResults,
  blocklyPreviewRequest,
  notebookMode,
  copiedCellAvailable,
  sourceNavigationHighlight,
  searchOpen,
  searchNavigation,
  deletedCellUndoStack,
  deletedCellRedoStack,
} from './state.js';
import { clearDebugRuntimeState } from './debug.js';

const DELETED_CELL_HISTORY_LIMIT = 50;

let blocklyPreviewRequestId = 0;
let cellIdSeed = Date.now();
let copiedCell = null;
let sourceNavigationHighlightTimer = null;
let sourceNavigationHighlightId = 0;
let searchNavigationId = 0;

function nextCellId() {
  cellIdSeed += 1;
  return `c${cellIdSeed}`;
}

// Called by screens.js when loading a new project to reset clipboard state.
export function clearCopiedCell() {
  copiedCell = null;
  copiedCellAvailable.set(false);
}

// ── Project properties ──

export function setProjectProperty(name, value) {
  let nextValue = null;
  projectProperties.update(properties => {
    const next = withProjectPropertyValue(properties, name, value);
    const propertyName = normalizeProjectPropertyName(name);
    nextValue = next[propertyName] ?? null;
    return next;
  });
  return nextValue;
}

export function updateProjectProperties(values) {
  projectProperties.update(properties => normalizeProjectProperties({
    ...properties,
    ...(values || {}),
  }));
}

export function resetProjectProperties(values = {}) {
  projectProperties.set(defaultProjectProperties(values));
}

// ── Blockly preview ──

export function requestBlocklyPreview(cellId, payload = {}) {
  blocklyPreviewRequest.set({
    id: ++blocklyPreviewRequestId,
    cellId,
    ...payload,
  });
}

// ── Universal search ──

export function openSearch() { searchOpen.set(true); }
export function closeSearch() { searchOpen.set(false); }
export function toggleSearch() { searchOpen.update(v => !v); }

export function requestSearchNavigation(payload = {}) {
  searchNavigation.set({ ...payload, token: ++searchNavigationId });
}

// ── Cell navigation ──

export function setActive(id) {
  activeCellId.set(id);
}

export function navigateToCellLine(cellId, line = 1) {
  const targetCell = get(cells).find(cell => cell?.id === cellId);
  if (!targetCell) return false;
  const targetLine = Math.max(1, Math.trunc(Number(line) || 1));
  const token = ++sourceNavigationHighlightId;
  notebookMode.set('cells');
  setActive(cellId);
  sourceNavigationHighlight.set({ cellId, line: targetLine, token });
  if (sourceNavigationHighlightTimer) clearTimeout(sourceNavigationHighlightTimer);
  sourceNavigationHighlightTimer = setTimeout(() => {
    sourceNavigationHighlight.update(current => (
      current?.token === token ? null : current
    ));
    sourceNavigationHighlightTimer = null;
  }, 2800);
  return true;
}

// ── Cell CRUD ──

export function addCodeCell() {
  const id = nextCellId();
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
  const id = nextCellId();
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

export function copyCellById(id) {
  const cell = get(cells).find(c => c.id === id);
  if (!cell) return false;
  copiedCell = JSON.parse(JSON.stringify(cell));
  copiedCellAvailable.set(true);
  return true;
}

export function pasteCopiedCellBelow(id) {
  if (!copiedCell) return null;
  const currentCells = get(cells);
  const sourceIndex = currentCells.findIndex(c => c.id === id);
  const insertIndex = sourceIndex === -1 ? currentCells.length : sourceIndex + 1;
  const pasted = {
    ...JSON.parse(JSON.stringify(copiedCell)),
    id: nextCellId(),
  };

  cells.update(cs => {
    const next = [...cs];
    next.splice(insertIndex, 0, pasted);
    return next;
  });
  setActive(pasted.id);
  return pasted.id;
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

function pushDeletedCell(cell, index) {
  deletedCellUndoStack.update(stack => {
    const next = [...stack, { cell: JSON.parse(JSON.stringify(cell)), index }];
    if (next.length > DELETED_CELL_HISTORY_LIMIT) next.shift();
    return next;
  });
  deletedCellRedoStack.set([]);
}

export function deleteCellById(id) {
  const currentCells = get(cells);
  const idx = currentCells.findIndex(c => c.id === id);
  if (idx === -1) return;
  pushDeletedCell(currentCells[idx], idx);
  cells.update(cs => cs.filter(c => c.id !== id));
  const nextCells = get(cells);
  const nextActive = nextCells[Math.min(idx, nextCells.length - 1)]?.id || null;
  activeCellId.set(nextActive);
}

export function undoDeleteCell() {
  const stack = get(deletedCellUndoStack);
  if (!stack.length) return false;
  const entry = stack[stack.length - 1];
  deletedCellUndoStack.set(stack.slice(0, -1));

  const collision = get(cells).some(c => c.id === entry.cell.id);
  const restored = collision ? { ...entry.cell, id: nextCellId() } : entry.cell;

  cells.update(cs => {
    const next = [...cs];
    const insertAt = Math.max(0, Math.min(entry.index, next.length));
    next.splice(insertAt, 0, JSON.parse(JSON.stringify(restored)));
    return next;
  });
  deletedCellRedoStack.update(s => [...s, { cell: restored, index: entry.index }]);
  setActive(restored.id);
  return true;
}

export function redoDeleteCell() {
  const stack = get(deletedCellRedoStack);
  if (!stack.length) return false;
  const entry = stack[stack.length - 1];
  deletedCellRedoStack.set(stack.slice(0, -1));

  const idx = get(cells).findIndex(c => c.id === entry.cell.id);
  if (idx === -1) return false;

  cells.update(cs => cs.filter(c => c.id !== entry.cell.id));
  deletedCellUndoStack.update(s => {
    const next = [...s, { cell: JSON.parse(JSON.stringify(entry.cell)), index: idx }];
    if (next.length > DELETED_CELL_HISTORY_LIMIT) next.shift();
    return next;
  });
  const nextCells = get(cells);
  const nextActive = nextCells[Math.min(idx, nextCells.length - 1)]?.id || null;
  activeCellId.set(nextActive);
  return true;
}

export function updateCellExecCount(id) {
  const count = get(execCounter);
  cells.update(cs => cs.map(c => c.id === id ? { ...c, execCount: count } : c));
  execCounter.update(n => n + 1);
  lastRunAt.set(Date.now());
}

export function updateCellCode(id, code) {
  clearDebugRuntimeState();
  cells.update(cs => cs.map(c => c.id === id ? { ...c, code } : c));
}

export function replaceCodeCells(codeChunks, { activeIndex = 0, skipDebugClear = false } = {}) {
  if (!skipDebugClear) clearDebugRuntimeState();
  const chunks = Array.from(codeChunks || [])
    .map(code => String(code ?? ''))
    .filter(code => code.trim().length > 0);
  let nextActive = null;

  cells.update(current => {
    const existingCodeCells = current.filter(cell => cell.type === 'code');
    const nextCodeCells = chunks.map((code, index) => ({
      id: existingCodeCells[index]?.id || nextCellId(),
      type: 'code',
      code,
      execCount: existingCodeCells[index]?.execCount ?? null,
    }));

    const next = [];
    let codeIndex = 0;
    let lastCodeOutputIndex = -1;
    for (const cell of current) {
      if (cell.type !== 'code') {
        next.push(cell);
        continue;
      }
      if (codeIndex < nextCodeCells.length) {
        next.push(nextCodeCells[codeIndex]);
        lastCodeOutputIndex = next.length - 1;
      }
      codeIndex += 1;
    }
    if (codeIndex === 0) {
      next.push(...nextCodeCells);
    } else if (codeIndex < nextCodeCells.length) {
      next.splice(lastCodeOutputIndex + 1, 0, ...nextCodeCells.slice(codeIndex));
    }

    const safeActiveIndex = Math.max(0, Math.min(activeIndex, nextCodeCells.length - 1));
    nextActive = nextCodeCells[safeActiveIndex]?.id || next[0]?.id || null;
    return next;
  });

  activeCellId.set(nextActive);
  return nextActive;
}

export function updateDesignCode(code) {
  clearDebugRuntimeState();
  designCode.set(code);
}

// ── Source accessors ──

export function getFalconSource() {
  return get(cells)
    .filter(cell => cell.type === 'code')
    .map(cell => cell.code || '')
    .join('\n\n');
}

export function getDesignSource() {
  return get(designCode);
}

// ── Do-it results ──

export function setDoItResult(cellId, lines, ok) {
  doItResults.update(r => ({ ...r, [cellId]: { lines, ok } }));
}

export function clearDoItResult(cellId) {
  doItResults.update(r => {
    const next = { ...r };
    delete next[cellId];
    return next;
  });
}

// ── Context menu ──

export function showCtx(e, id, toggle = false) {
  e.preventDefault();
  e.stopPropagation();
  setActive(id);
  ctxMenu.update(current => {
    if (toggle && current.show && current.cellId === id) {
      return { ...current, show: false, cellId: null };
    }

    return {
      show: true,
      x: Math.min(e.clientX, window.innerWidth - 180),
      y: Math.min(e.clientY, window.innerHeight - 260),
      cellId: id,
    };
  });
}

export function hideCtx() {
  ctxMenu.update(m => ({ ...m, show: false, cellId: null }));
}
