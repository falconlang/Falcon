export const DEBUG_TRACE_PREFIX = '__TENSOR_DEBUG__';
export const DEBUG_VALUE_PREFIX = '__TENSOR_DEBUG_VALUE__';
export const DEBUG_BREAK_PREFIX = '__TENSOR_DEBUG_BREAK__';
export const DEBUG_IDLE_CLEAR_MS = 900;

export function buildFalconSourceMap(cellList = []) {
  const codeCells = Array.from(cellList || []).filter(cell => cell?.type === 'code');
  const sourceParts = codeCells.map(cell => String(cell.code ?? ''));
  const entries = [];
  let unifiedLine = 1;

  for (let cellIndex = 0; cellIndex < codeCells.length; cellIndex += 1) {
    const cell = codeCells[cellIndex];
    const code = String(cell.code ?? '');
    const lines = code.split('\n');

    for (let lineIndex = 0; lineIndex < lines.length; lineIndex += 1) {
      entries.push({
        cellId: cell.id,
        cellIndex,
        cellLine: lineIndex + 1,
        unifiedLine,
        text: lines[lineIndex],
      });
      unifiedLine += 1;
    }

    if (cellIndex < codeCells.length - 1) {
      unifiedLine += 1;
    }
  }

  return {
    source: sourceParts.join('\n\n'),
    entries,
  };
}

export function lineMapEntryForUnifiedLine(lineMap, unifiedLine) {
  const line = Number(unifiedLine);
  if (!Number.isFinite(line)) return null;
  return Array.from(lineMap || []).find(entry => entry.unifiedLine === line) || null;
}

export function parseDebugTraceValue(value) {
  const raw = value?.item ?? value?.value ?? value?.message;
  if (typeof raw !== 'string' || !raw.startsWith(DEBUG_TRACE_PREFIX)) return null;

  try {
    const payload = JSON.parse(raw.slice(DEBUG_TRACE_PREFIX.length));
    const unifiedLine = Number(payload.unifiedLine);
    const cellLine = Number(payload.cellLine);
    return {
      type: 'trace',
      sessionId: String(payload.sessionId || ''),
      traceId: String(payload.traceId || ''),
      cellId: payload.cellId == null ? null : String(payload.cellId),
      cellLine: Number.isFinite(cellLine) ? cellLine : null,
      unifiedLine: Number.isFinite(unifiedLine) ? unifiedLine : null,
      timestamp: Date.now(),
    };
  } catch {
    return null;
  }
}

export function parseDebugValueCapture(value) {
  const raw = value?.item ?? value?.value ?? value?.message;
  if (typeof raw !== 'string' || !raw.startsWith(DEBUG_VALUE_PREFIX)) return null;

  const body = raw.slice(DEBUG_VALUE_PREFIX.length);
  const separator = body.indexOf('\t');
  if (separator === -1) return null;
  const exprId = body.slice(0, separator);
  if (!exprId) return null;

  return {
    type: 'value',
    exprId,
    value: body.slice(separator + 1),
    timestamp: Date.now(),
  };
}

export function parseDebugBreakpointHit(value) {
  const raw = value?.item ?? value?.value ?? value?.message;
  if (typeof raw !== 'string' || !raw.startsWith(DEBUG_BREAK_PREFIX)) return null;

  try {
    const payload = JSON.parse(raw.slice(DEBUG_BREAK_PREFIX.length));
    const unifiedLine = Number(payload.unifiedLine);
    const cellLine = Number(payload.cellLine);
    return {
      type: 'breakpoint-hit',
      sessionId: String(payload.sessionId || ''),
      hitId: String(payload.hitId || ''),
      traceId: String(payload.traceId || payload.hitId || ''),
      cellId: payload.cellId == null ? null : String(payload.cellId),
      cellLine: Number.isFinite(cellLine) ? cellLine : null,
      unifiedLine: Number.isFinite(unifiedLine) ? unifiedLine : null,
      timestamp: Date.now(),
    };
  } catch {
    return null;
  }
}

export function parseDebugRuntimeEvent(value) {
  return parseDebugTraceValue(value)
    || parseDebugValueCapture(value)
    || parseDebugBreakpointHit(value);
}

export function isDebugTraceValue(value) {
  return Boolean(parseDebugRuntimeEvent(value));
}
