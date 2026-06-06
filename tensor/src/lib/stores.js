import { writable, derived, get } from 'svelte/store';
import {
  defaultProjectProperties,
  normalizeProjectPropertyName,
  normalizeProjectProperties,
  withProjectPropertyValue,
} from './project-properties.js';
import {
  isValidAppInventorAssetName,
  normalizeAppInventorAssetName,
} from './appinventor-validation.js';
import { setProjectExtensionComponentDescriptors } from './appinventor-component-registry.js';
import { isDebugTraceValue, lineMapEntryForUnifiedLine } from './debug-source-map.js';

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
export const projectExtensionComponents = writable([]);
export const projectName = writable('falcon_tour');
export const projectProperties = writable(defaultProjectProperties());
export const activeCellId = writable('c1');
export const execCounter = writable(6);
export const ctxMenu = writable({ show: false, x: 0, y: 0, cellId: null });
export const liveTestOpen = writable(false);
export const companionCommand = writable(null);
export const liveTestState = writable({
  status: 'idle',
  code: null,
  error: null,
  messageCount: 0,
});
export const doItCellId = writable(null);
export const doItResults = writable({});
export const unifiedSelectionActive = writable(false);
export const blocklyPreviewRequest = writable(null);
export const lastRunAt = writable(null);
export const sidebarVisible = writable(typeof window !== 'undefined' ? window.innerWidth > 1024 : true);
export const debugCollapsed = writable(true);
export const notebookMode = writable('cells'); // 'cells' | 'unified'
export const debugOpenHeight = writable(200);
export const debugLogs = writable([]);
export const debugModeEnabled = writable(false);
export const debugUserEnabled = writable(false);
export const runtimeErrorNotice = writable({ show: false, error: null });
export const debugExecutionState = writable({
  status: 'idle',
  sessionId: null,
  startedAt: null,
  pausedAt: null,
  hitId: null,
});
export const debugLineMap = writable([]);
export const debugActiveLocation = writable(null);
export const debugRuntimeErrors = writable({});
export const debugAnnotationActive = writable(false);
export const debugBreakpoints = writable({});
export const debugPausedFrame = writable(null);
export const debugExpressionCatalog = writable([]);
export const debugExpressionValues = writable({});
export const copiedCellAvailable = writable(false);
export const sourceNavigationHighlight = writable(null);

// ── Universal search ──
export const searchOpen = writable(false);
// Live source of the unified Script editor, published while it is mounted.
export const unifiedSearchSource = writable('');
// Searchable component/property index, published by the designer tree editor while mounted.
export const designerSearchIndex = writable([]);
// True while the designer tree editor is mounted (designer is in Tree mode).
export const designerTreeActive = writable(false);
// Token-stamped navigation request consumed by the active editors.
export const searchNavigation = writable(null);

// ── Deleted-cell undo/redo history (per screen) ──
export const deletedCellUndoStack = writable([]); // [{ cell, index }]
export const deletedCellRedoStack = writable([]);
export const canUndoDeletedCell = derived(deletedCellUndoStack, s => s.length > 0);
export const canRedoDeletedCell = derived(deletedCellRedoStack, s => s.length > 0);
const DELETED_CELL_HISTORY_LIMIT = 50;

let blocklyPreviewRequestId = 0;
let cellIdSeed = Date.now();
let copiedCell = null;
let sourceNavigationHighlightTimer = null;
let sourceNavigationHighlightId = 0;
let searchNavigationId = 0;

export function openSearch() { searchOpen.set(true); }
export function closeSearch() { searchOpen.set(false); }
export function toggleSearch() { searchOpen.update(v => !v); }

// Dispatch a navigation request to whichever editor owns the result's coordinates.
// No mode flipping: the search overlay blocks the pane mode toggles, so results
// always match the current mode and navigation stays within it.
export function requestSearchNavigation(payload = {}) {
  searchNavigation.set({ ...payload, token: ++searchNavigationId });
}

function nextCellId() {
  cellIdSeed += 1;
  return `c${cellIdSeed}`;
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

export function requestBlocklyPreview(cellId, payload = {}) {
  blocklyPreviewRequest.set({
    id: ++blocklyPreviewRequestId,
    cellId,
    ...payload,
  });
}

// ── Screen management ──
export const screenList = writable(['Screen1']);
export const activeScreen = writable('Screen1');
export const rawBlocklyXml = writable('');
export const sourceScm = writable('');
export const sourceDesignCode = writable('');
export const sourceScmUpgradeWarnings = writable([]);
// Saved state for non-active screens:
// { [screenName]: { cells, designCode, rawBlocklyXml, sourceScm, sourceDesignCode, sourceScmUpgradeWarnings } }
const screenSavedStates = writable({});

function cloneCells(cellList) {
  return JSON.parse(JSON.stringify(cellList || []));
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
  // Save current state
  screenSavedStates.update(s => ({
    ...s,
    [curr]: currentScreenState(),
  }));
  // Use provided name or find unique default
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
  // Switch to empty new screen
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
  copiedCell = null;
  copiedCellAvailable.set(false);
}

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
      if (isDebugTraceValue(value)) return null;

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

export function enableDebugMode() {
  debugModeEnabled.set(true);
  debugUserEnabled.set(true);
  clearDebugRuntimeState();
  runtimeErrorNotice.set({ show: false, error: null });
}

export function disableDebugMode() {
  debugModeEnabled.set(false);
  debugUserEnabled.set(false);
  clearDebugRuntimeState();
  runtimeErrorNotice.set({ show: false, error: null });
  debugExecutionState.set({ status: 'idle', sessionId: null, startedAt: null, pausedAt: null, hitId: null });
  debugLineMap.set([]);
  debugExpressionCatalog.set([]);
  debugExpressionValues.set({});
}

export function dismissRuntimeErrorNotice() {
  runtimeErrorNotice.set({ show: false, error: null });
  clearDebugRuntimeState();
}

export function startDebugSession({ sessionId, lineMap = [], expressionCatalog = [] } = {}) {
  debugLineMap.set(Array.from(lineMap || []));
  debugExpressionCatalog.set(Array.from(expressionCatalog || []));
  debugExpressionValues.set({});
  debugPausedFrame.set(null);
  debugRuntimeErrors.set({});
  debugActiveLocation.set(null);
  debugExecutionState.set({
    status: 'active',
    sessionId: sessionId || null,
    startedAt: Date.now(),
    pausedAt: null,
    hitId: null,
  });
}

export function clearDebugActiveLocation(sessionId = null) {
  const current = get(debugExecutionState);
  if (sessionId && current.sessionId && sessionId !== current.sessionId) return;
  debugActiveLocation.set(null);
}

export function clearDebugRuntimeState() {
  debugActiveLocation.set(null);
  debugPausedFrame.set(null);
  debugExpressionValues.set({});
  debugRuntimeErrors.set({});
}

function normalizedDebugLocation(location) {
  const mapEntry = location?.unifiedLine != null
    ? lineMapEntryForUnifiedLine(get(debugLineMap), location.unifiedLine)
    : null;

  return {
    sessionId: location?.sessionId || get(debugExecutionState).sessionId || null,
    traceId: location?.traceId || '',
    cellId: mapEntry?.cellId ?? location?.cellId ?? null,
    cellLine: mapEntry?.cellLine ?? location?.cellLine ?? null,
    unifiedLine: mapEntry?.unifiedLine ?? location?.unifiedLine ?? null,
  };
}

export function setDebugTraceLocation(location) {
  if (!location?.sessionId) return false;
  const current = get(debugExecutionState);
  if (current.sessionId && location.sessionId !== current.sessionId) return false;

  const normalized = {
    ...normalizedDebugLocation(location),
    timestamp: Date.now(),
  };

  debugActiveLocation.set(normalized);
  return true;
}

function debugBreakpointKey(input = {}) {
  const screen = String(input.screen || get(activeScreen) || 'Screen1');
  const cellId = input.cellId == null ? '' : String(input.cellId);
  const cellLine = Number(input.cellLine);
  if (!cellId || !Number.isFinite(cellLine) || cellLine < 1) return '';
  return `${screen}:${cellId}:${Math.trunc(cellLine)}`;
}

export function toggleDebugBreakpoint(input = {}) {
  const key = debugBreakpointKey(input);
  if (!key) return null;

  let nextValue = null;
  debugBreakpoints.update(current => {
    const next = { ...(current || {}) };
    if (next[key]) {
      delete next[key];
      nextValue = null;
      return next;
    }

    nextValue = {
      screen: String(input.screen || get(activeScreen) || 'Screen1'),
      cellId: String(input.cellId),
      cellLine: Math.trunc(Number(input.cellLine)),
      unifiedLine: Number.isFinite(Number(input.unifiedLine)) ? Math.trunc(Number(input.unifiedLine)) : null,
      enabled: input.enabled !== false,
      createdAt: Date.now(),
    };
    next[key] = nextValue;
    return next;
  });
  return nextValue;
}

export function hasDebugBreakpoint(input = {}) {
  const key = debugBreakpointKey(input);
  return Boolean(key && get(debugBreakpoints)[key]);
}

export function debugBreakpointsForCompile(screen = get(activeScreen)) {
  return Object.values(get(debugBreakpoints) || {})
    .filter(bp => bp?.enabled !== false && (!screen || bp.screen === screen));
}

export function setDebugExpressionValue(event) {
  if (!event?.exprId) return false;
  debugExpressionValues.update(values => ({
    ...(values || {}),
    [event.exprId]: {
      value: stringifyDebugValue(event.value),
      timestamp: event.timestamp || Date.now(),
    },
  }));
  return true;
}

export function setDebugBreakpointHit(hit) {
  if (!hit?.sessionId) return false;
  const current = get(debugExecutionState);
  if (current.sessionId && hit.sessionId !== current.sessionId) return false;

  const normalized = {
    ...normalizedDebugLocation(hit),
    hitId: hit.hitId || hit.traceId || '',
    timestamp: hit.timestamp || Date.now(),
  };

  debugPausedFrame.set({
    ...normalized,
    values: get(debugExpressionValues),
  });
  debugActiveLocation.set(normalized);
  debugExecutionState.set({
    status: 'paused',
    sessionId: normalized.sessionId,
    startedAt: current.startedAt || Date.now(),
    pausedAt: normalized.timestamp,
    hitId: normalized.hitId,
  });
  return true;
}

export function continueDebugExecution(hitId = null) {
  const current = get(debugExecutionState);
  if (hitId && current.hitId && hitId !== current.hitId) return false;
  debugPausedFrame.set(null);
  debugActiveLocation.set(null);
  debugExecutionState.set({
    status: current.sessionId ? 'active' : 'idle',
    sessionId: current.sessionId || null,
    startedAt: current.startedAt || null,
    pausedAt: null,
    hitId: null,
  });
  return true;
}

function sourceLineColumn(source, offset) {
  const value = String(source ?? '');
  const bounded = Math.max(0, Math.min(Number(offset) || 0, value.length));
  const before = value.slice(0, bounded);
  const lines = before.split('\n');
  return {
    line: lines.length,
    column: lines[lines.length - 1].length,
  };
}

export function lookupPausedSelection({ cellId = null, source = '', start = 0, end = 0, unified = false } = {}) {
  const frame = get(debugPausedFrame);
  if (!frame) return null;

  const value = String(source ?? '');
  const s = Math.max(0, Math.min(Number(start) || 0, value.length));
  const e = Math.max(s, Math.min(Number(end) || 0, value.length));
  if (s === e) return null;

  const selectedText = value.slice(s, e).trim();
  if (!selectedText) return null;

  const startPos = sourceLineColumn(value, s);
  const endPos = sourceLineColumn(value, e);
  const entries = get(debugExpressionCatalog) || [];
  const candidates = entries
    .filter(entry => {
      if (unified) {
        if (entry.unifiedLine !== startPos.line || entry.unifiedLine !== endPos.line) return false;
      } else if (cellId != null) {
        if (entry.cellId !== cellId || entry.cellLine !== startPos.line || entry.cellLine !== endPos.line) return false;
      } else {
        return false;
      }
      return entry.sourceText === selectedText
        || (
          entry.startColumn >= startPos.column
          && entry.endColumn <= endPos.column
          && selectedText.includes(entry.sourceText)
        );
    })
    .sort((a, b) => {
      const aExact = a.sourceText === selectedText ? 1 : 0;
      const bExact = b.sourceText === selectedText ? 1 : 0;
      if (aExact !== bExact) return bExact - aExact;
      return (b.endColumn - b.startColumn) - (a.endColumn - a.startColumn);
    });

  const entry = candidates[0];
  if (!entry) {
    return { status: 'out-of-scope', sourceText: selectedText, lines: ['out of scope'], ok: false };
  }

  const captured = frame.values?.[entry.exprId] || get(debugExpressionValues)[entry.exprId];
  if (!captured) {
    return { status: 'not-evaluated', sourceText: entry.sourceText, expressionId: entry.exprId, lines: ['not evaluated'], ok: false };
  }

  return {
    status: 'ready',
    sourceText: entry.sourceText,
    expressionId: entry.exprId,
    value: captured.value,
    lines: [captured.value],
    ok: true,
  };
}

export function setDebugRuntimeError(error) {
  const active = get(debugActiveLocation);
  const location = normalizedDebugLocation(error?.location || active);
  if (!location?.cellId) return false;

  const entry = {
    sessionId: location.sessionId,
    traceId: location.traceId || '',
    cellId: location.cellId,
    cellLine: location.cellLine ?? null,
    unifiedLine: location.unifiedLine ?? null,
    message: stringifyDebugValue(error?.message || 'Runtime error'),
    source: stringifyDebugValue(error?.source || 'Runtime'),
    status: error?.status || null,
    timestamp: Date.now(),
  };

  debugRuntimeErrors.update(errors => ({
    ...errors,
    [entry.cellId]: entry,
  }));
  debugActiveLocation.set(entry);
  return true;
}

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
  if (idx === -1) return false; // stale entry; nothing to re-delete

  cells.update(cs => cs.filter(c => c.id !== entry.cell.id));
  // Push back onto the undo stack directly so the redo stack is preserved.
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

function assetNameFrom(input) {
  return typeof input === 'string' ? input : input?.name;
}

function normalizeAssetName(name) {
  return normalizeAppInventorAssetName(name);
}

function splitAssetName(name) {
  const dot = name.lastIndexOf('.');
  if (dot <= 0) return { base: name, ext: '' };
  return { base: name.slice(0, dot), ext: name.slice(dot) };
}

function uniqueAssetName(name, assets, exceptId = null) {
  const clean = normalizeAssetName(name);
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
  return cloneCells(cellList || []).map(cell => (
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

export function getFalconSource() {
  return get(cells)
    .filter(cell => cell.type === 'code')
    .map(cell => cell.code || '')
    .join('\n\n');
}

export function getDesignSource() {
  return get(designCode);
}

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
