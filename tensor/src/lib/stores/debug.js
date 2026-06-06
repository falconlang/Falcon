import { get } from 'svelte/store';
import { isDebugTraceValue, lineMapEntryForUnifiedLine } from '../debug-source-map.js';
import {
  activeScreen,
  debugLogs,
  debugModeEnabled,
  debugUserEnabled,
  runtimeErrorNotice,
  debugExecutionState,
  debugLineMap,
  debugActiveLocation,
  debugRuntimeErrors,
  debugAnnotationActive,
  debugBreakpoints,
  debugPausedFrame,
  debugExpressionCatalog,
  debugExpressionValues,
  deletedCellUndoStack,
  deletedCellRedoStack,
} from './state.js';

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

export function clearDebugRuntimeState() {
  debugActiveLocation.set(null);
  debugPausedFrame.set(null);
  debugExpressionValues.set({});
  debugRuntimeErrors.set({});
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
