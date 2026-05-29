import {
  DEBUG_BREAK_PREFIX,
  DEBUG_TRACE_PREFIX,
  DEBUG_VALUE_PREFIX,
  buildFalconSourceMap,
} from './debug-source-map.js';

const DEFAULT_DEBUG_NOTIFIER = 'TensorDebugNotifier';
const DEBUG_CONTINUE_GLOBAL = 'tensorDebugContinueFlag';
const DEBUG_VALUE_FUNC = 'tensorDebugValue';
const DEBUG_BREAK_FUNC = 'tensorDebugBreak';

export function createDebugSessionId() {
  const random = Math.random().toString(36).slice(2, 8);
  return `dbg-${Date.now().toString(36)}-${random}`;
}

function falconString(value) {
  return `"${String(value ?? '')
    .replace(/\\/g, '\\\\')
    .replace(/"/g, '\\"')
    .replace(/\n/g, '\\n')
    .replace(/\r/g, '\\r')}"`;
}

function stripLineComment(line) {
  let inString = false;
  for (let i = 0; i < line.length; i += 1) {
    const ch = line[i];
    const next = line[i + 1];
    if (inString) {
      if (ch === '\\') {
        i += 1;
        continue;
      }
      if (ch === '"') inString = false;
      continue;
    }
    if (ch === '"') {
      inString = true;
      continue;
    }
    if (ch === '/' && next === '/') return line.slice(0, i);
  }
  return line;
}

function splitLineComment(line) {
  const code = stripLineComment(line);
  return {
    code,
    comment: line.slice(code.length),
  };
}

function maskIgnoredSpans(line) {
  const chars = String(line ?? '').split('');
  let inString = false;
  let inLineComment = false;

  for (let i = 0; i < chars.length; i += 1) {
    const ch = chars[i];
    const next = chars[i + 1];

    if (inLineComment) {
      if (ch === '\n') inLineComment = false;
      else chars[i] = ' ';
      continue;
    }

    if (inString) {
      if (ch === '\\') {
        chars[i] = ' ';
        if (i + 1 < chars.length) chars[++i] = ' ';
        continue;
      }
      if (ch === '"') inString = false;
      chars[i] = ' ';
      continue;
    }

    if (ch === '/' && next === '/') {
      chars[i] = ' ';
      chars[++i] = ' ';
      inLineComment = true;
      continue;
    }

    if (ch === '"') {
      chars[i] = ' ';
      inString = true;
    }
  }

  return chars.join('');
}

function braceDelta(line) {
  const searchable = stripLineComment(line);
  let inString = false;
  let delta = 0;

  for (let i = 0; i < searchable.length; i += 1) {
    const ch = searchable[i];
    if (inString) {
      if (ch === '\\') {
        i += 1;
        continue;
      }
      if (ch === '"') inString = false;
      continue;
    }
    if (ch === '"') {
      inString = true;
      continue;
    }
    if (ch === '{') delta += 1;
    if (ch === '}') delta -= 1;
  }

  return delta;
}

function isSkippableLine(trimmed) {
  if (!trimmed || trimmed.startsWith('//')) return true;
  return /^[{};]+$/.test(trimmed);
}

function startsClosingBranch(trimmed) {
  return /^}\s*else\b/.test(trimmed);
}

function startsElseBranch(trimmed) {
  return /^else\b/.test(trimmed);
}

function isTopLevelBodyHeader(trimmed, depthBefore) {
  return depthBefore === 0 && /^(func|when)\b.*\{\s*(?:(?:\/\/.*)?)$/.test(trimmed);
}

function traceLine(notifierName, payload, indent) {
  return `${indent}${notifierName}.LogInfo(${falconString(DEBUG_TRACE_PREFIX + JSON.stringify(payload))})`;
}

function debugPrelude(notifierName) {
  return `global ${DEBUG_CONTINUE_GLOBAL} = false

func ${DEBUG_VALUE_FUNC}(tensorDebugExprId, tensorDebugCapturedValue) = {
  ${notifierName}.LogInfo(${falconString(DEBUG_VALUE_PREFIX)} _ tensorDebugExprId _ ${falconString('\t')} _ tensorDebugCapturedValue)
  tensorDebugCapturedValue
}

func ${DEBUG_BREAK_FUNC}(tensorDebugHitPayload) {
  this.${DEBUG_CONTINUE_GLOBAL} = false
  ${notifierName}.LogInfo(${falconString(DEBUG_BREAK_PREFIX)} _ tensorDebugHitPayload)
  while (!this.${DEBUG_CONTINUE_GLOBAL}) {
  }
  this.${DEBUG_CONTINUE_GLOBAL} = false
}`;
}

function captureExpression(exprId, expression) {
  return `${DEBUG_VALUE_FUNC}(${falconString(exprId)}, ${expression})`;
}

function captureValueLine(notifierName, exprId, expression, indent) {
  return `${indent}${notifierName}.LogInfo(${falconString(DEBUG_VALUE_PREFIX)} _ ${falconString(exprId)} _ ${falconString('\t')} _ ${expression})`;
}

function breakpointLine(payload, indent) {
  return `${indent}${DEBUG_BREAK_FUNC}(${falconString(JSON.stringify(payload))})`;
}

function isTraceableExecutionLine(trimmed) {
  return !isSkippableLine(trimmed)
    && !startsClosingBranch(trimmed)
    && !startsElseBranch(trimmed)
    && !trimmed.startsWith('}');
}

function tracePayload(sessionId, entry, unifiedLine) {
  return {
    type: 'trace',
    sessionId,
    traceId: `${sessionId}:${unifiedLine}`,
    cellId: entry.cellId,
    cellLine: entry.cellLine,
    unifiedLine,
  };
}

function breakpointPayload(sessionId, entry, unifiedLine) {
  return {
    type: 'breakpoint-hit',
    sessionId,
    hitId: `${sessionId}:${unifiedLine}`,
    traceId: `${sessionId}:${unifiedLine}`,
    cellId: entry.cellId,
    cellLine: entry.cellLine,
    unifiedLine,
  };
}

function normalizeBreakpointKey(breakpoint) {
  const cellId = breakpoint?.cellId == null ? '' : String(breakpoint.cellId);
  const cellLine = Number(breakpoint?.cellLine);
  const unifiedLine = Number(breakpoint?.unifiedLine);
  if (cellId && Number.isFinite(cellLine)) return `cell:${cellId}:${cellLine}`;
  if (Number.isFinite(unifiedLine)) return `unified:${unifiedLine}`;
  return '';
}

function breakpointSetFrom(options = {}) {
  const breakpoints = Array.from(options.breakpoints || []);
  return new Set(breakpoints.map(normalizeBreakpointKey).filter(Boolean));
}

function hasBreakpointForEntry(breakpointSet, entry, unifiedLine) {
  if (!entry) return false;
  return breakpointSet.has(`cell:${entry.cellId}:${entry.cellLine}`)
    || breakpointSet.has(`unified:${unifiedLine}`);
}

function expressionCatalogBuilder(sessionId) {
  const entries = [];
  const byVarName = new Map();
  let expressionIndex = 0;

  const nextId = (kind, entry, sourceText) => {
    expressionIndex += 1;
    return `${sessionId}:expr:${kind}:${entry.unifiedLine}:${expressionIndex}`;
  };

  const add = (entry, sourceText, startColumn, endColumn, kind, exprId = null) => {
    if (!entry || !String(sourceText ?? '').trim()) return null;
    const id = exprId || nextId(kind, entry, sourceText);
    entries.push({
      exprId: id,
      kind,
      sourceText: String(sourceText ?? '').trim(),
      cellId: entry.cellId,
      cellLine: entry.cellLine,
      unifiedLine: entry.unifiedLine,
      startColumn,
      endColumn,
    });
    return id;
  };

  const declareVar = (entry, name, startColumn, endColumn) => {
    const id = nextId('var', entry, name);
    byVarName.set(name, id);
    add(entry, name, startColumn, endColumn, 'var', id);
    return id;
  };

  const varExprId = name => byVarName.get(name) || null;

  return {
    entries,
    add,
    declareVar,
    varExprId,
    knownVarNames: () => Array.from(byVarName.keys()),
  };
}

function scanVariableOccurrences(line, entry, catalog) {
  if (!entry) return;
  const masked = maskIgnoredSpans(line);
  for (const name of catalog.knownVarNames()) {
    const exprId = catalog.varExprId(name);
    if (!exprId) continue;
    const re = new RegExp(`\\b${name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\b`, 'g');
    let match;
    while ((match = re.exec(masked)) !== null) {
      catalog.add(entry, name, match.index, match.index + name.length, 'var-ref', exprId);
    }
  }
}

function wrapNotExpressions(line, entry, catalog, startOffset = 0) {
  if (!entry) return line;
  const masked = maskIgnoredSpans(line);
  const ranges = [];
  const re = /!\s*([A-Za-z][A-Za-z0-9]*)\b/g;
  let match;
  while ((match = re.exec(masked)) !== null) {
    const prev = masked[match.index - 1] || '';
    if (prev === '=' || prev === '!') continue;
    const sourceText = line.slice(match.index, match.index + match[0].length);
    const exprId = catalog.add(
      entry,
      sourceText,
      startOffset + match.index,
      startOffset + match.index + match[0].length,
      'expression',
    );
    if (exprId) ranges.push({ start: match.index, end: match.index + match[0].length, exprId });
  }

  let next = line;
  for (let i = ranges.length - 1; i >= 0; i -= 1) {
    const range = ranges[i];
    const sourceText = next.slice(range.start, range.end);
    next = `${next.slice(0, range.start)}${captureExpression(range.exprId, sourceText)}${next.slice(range.end)}`;
  }
  return next;
}

function instrumentLocalLine(line, entry, catalog, notifierName) {
  const { code, comment } = splitLineComment(line);
  const match = code.match(/^(\s*local\s+)([A-Za-z][A-Za-z0-9]*)(\s*=\s*)(.+?)\s*$/);
  if (!entry || !match) return { line, after: [] };

  const [, prefix, name, separator, rhs] = match;
  const nameStart = prefix.length;
  const nameEnd = nameStart + name.length;
  const rhsStart = prefix.length + name.length + separator.length;
  const varExprId = catalog.declareVar(entry, name, nameStart, nameEnd);
  const rhsExprId = catalog.add(entry, rhs, rhsStart, rhsStart + rhs.length, 'expression');
  const wrappedRhs = rhsExprId
    ? captureExpression(rhsExprId, wrapNotExpressions(rhs, entry, catalog, rhsStart))
    : rhs;
  const indent = line.match(/^\s*/)?.[0] || '';
  return {
    line: `${prefix}${name}${separator}${wrappedRhs}${comment}`,
    after: [captureValueLine(notifierName, varExprId, name, indent)],
  };
}

function instrumentConditionLine(line, entry, catalog) {
  if (!entry) return line;
  const match = line.match(/^(\s*(?:if|while)\s*\()(.+)(\)\s*\{?\s*(?:(?:\/\/.*)?)$)/);
  if (!match) return wrapNotExpressions(line, entry, catalog, 0);

  const [, prefix, condition, suffix] = match;
  const conditionStart = prefix.length;
  const wrappedCondition = wrapNotExpressions(condition, entry, catalog, conditionStart);
  const conditionExprId = catalog.add(
    entry,
    condition,
    conditionStart,
    conditionStart + condition.length,
    'condition',
  );

  return `${prefix}${conditionExprId ? captureExpression(conditionExprId, wrappedCondition) : wrappedCondition}${suffix}`;
}

function instrumentLineExpressions(line, entry, catalog, notifierName) {
  scanVariableOccurrences(line, entry, catalog);
  const local = instrumentLocalLine(line, entry, catalog, notifierName);
  if (local.after.length) return local;
  return { line: instrumentConditionLine(line, entry, catalog), after: [] };
}

function nextExecutableEntry(lines, entryByLine, startIndex) {
  let localDepth = 1;

  for (let index = startIndex + 1; index < lines.length; index += 1) {
    const unifiedLine = index + 1;
    const line = lines[index];
    const trimmed = stripLineComment(line).trim();
    const entry = entryByLine.get(unifiedLine);

    if (entry && localDepth > 0 && isTraceableExecutionLine(trimmed)) {
      return { entry, unifiedLine };
    }

    localDepth += braceDelta(line);
    if (localDepth <= 0) break;
  }

  return null;
}

function existingComponentNames(source) {
  const names = new Set();
  const searchable = String(source ?? '').split('\n').map(stripLineComment).join('\n');
  const re = /\b[A-Z][A-Za-z0-9_]*\s*\.\s*([A-Za-z_][A-Za-z0-9_]*)\b/g;
  let match;
  while ((match = re.exec(searchable)) !== null) names.add(match[1]);
  return names;
}

function uniqueDebugNotifierName(source, preferredName = DEFAULT_DEBUG_NOTIFIER) {
  const existing = existingComponentNames(source);
  if (!existing.has(preferredName)) return preferredName;
  let index = 2;
  while (existing.has(`${preferredName}${index}`)) index += 1;
  return `${preferredName}${index}`;
}

export function ensureDebugNotifierDesignSource(annSource, preferredName = DEFAULT_DEBUG_NOTIFIER) {
  const source = String(annSource ?? '');
  const notifierName = uniqueDebugNotifierName(source, preferredName);
  const insertion = `Notifier.${notifierName}`;
  const closeIndex = source.lastIndexOf('}');

  if (closeIndex === -1) {
    return {
      designSource: source.trim() ? `${source}\n${insertion}` : insertion,
      notifierName,
    };
  }

  const beforeClose = source.slice(0, closeIndex).replace(/\s*$/, '');
  const afterClose = source.slice(closeIndex);
  const bodyTail = beforeClose;
  const needsSeparator = bodyTail.length > 0 && !/[{,]$/.test(bodyTail);
  const separator = needsSeparator ? ',' : '';

  return {
    designSource: `${beforeClose}${separator}\n  ${insertion}\n${afterClose}`,
    notifierName,
  };
}

export function instrumentFalconSourceForDebug(source, lineMapEntries = [], options = {}) {
  const sessionId = options.sessionId || createDebugSessionId();
  const notifierName = options.notifierName || DEFAULT_DEBUG_NOTIFIER;
  const breakpointSet = breakpointSetFrom(options);
  const mapEntries = lineMapEntries.length
    ? lineMapEntries
    : buildFalconSourceMap([{ id: 'source', type: 'code', code: source }]).entries;
  const entryByLine = new Map(mapEntries.map(entry => [entry.unifiedLine, entry]));
  const lines = String(source ?? '').split('\n');
  const instrumented = [debugPrelude(notifierName), ''];
  const tracePoints = [];
  const breakpointPoints = [];
  const earlyTracedLines = new Set();
  const catalog = expressionCatalogBuilder(sessionId);
  let depth = 0;

  for (let index = 0; index < lines.length; index += 1) {
    const unifiedLine = index + 1;
    const line = lines[index];
    const trimmed = stripLineComment(line).trim();
    const entry = entryByLine.get(unifiedLine);
    const indent = line.match(/^\s*/)?.[0] || '';
    const depthBefore = depth;
    const topLevelHeader = entry && isTopLevelBodyHeader(trimmed, depthBefore);
    const closingBranch = entry && startsClosingBranch(trimmed);
    const elseBranch = entry && startsElseBranch(trimmed);
    const canTraceBefore = entry
      && depthBefore > 0
      && isTraceableExecutionLine(trimmed)
      && !earlyTracedLines.has(unifiedLine);
    const shouldBreak = canTraceBefore && hasBreakpointForEntry(breakpointSet, entry, unifiedLine);

    if (canTraceBefore) {
      const trace = tracePayload(sessionId, entry, unifiedLine);
      instrumented.push(traceLine(notifierName, trace, indent));
      tracePoints.push(trace);
      if (shouldBreak) {
        const breakpoint = breakpointPayload(sessionId, entry, unifiedLine);
        instrumented.push(breakpointLine(breakpoint, indent));
        breakpointPoints.push(breakpoint);
      }
    }

    const expressionLine = instrumentLineExpressions(line, entry, catalog, notifierName);
    instrumented.push(expressionLine.line);
    instrumented.push(...expressionLine.after);

    if (topLevelHeader || closingBranch || elseBranch) {
      const traceIndent = `${indent}  `;
      const target = nextExecutableEntry(lines, entryByLine, index);
      const trace = target
        ? tracePayload(sessionId, target.entry, target.unifiedLine)
        : tracePayload(sessionId, entry, unifiedLine);
      instrumented.push(traceLine(notifierName, trace, traceIndent));
      tracePoints.push(trace);
      earlyTracedLines.add(trace.unifiedLine);
    }

    depth = Math.max(0, depth + braceDelta(line));
  }

  return {
    source: instrumented.join('\n'),
    sessionId,
    notifierName,
    tracePoints,
    breakpointPoints,
    expressionCatalog: catalog.entries,
    lineMap: mapEntries,
  };
}

export { DEBUG_CONTINUE_GLOBAL };
