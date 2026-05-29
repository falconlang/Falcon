import { DEBUG_TRACE_PREFIX, buildFalconSourceMap } from './debug-source-map.js';

const DEFAULT_DEBUG_NOTIFIER = 'TensorDebugNotifier';

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
  const mapEntries = lineMapEntries.length
    ? lineMapEntries
    : buildFalconSourceMap([{ id: 'source', type: 'code', code: source }]).entries;
  const entryByLine = new Map(mapEntries.map(entry => [entry.unifiedLine, entry]));
  const lines = String(source ?? '').split('\n');
  const instrumented = [];
  const tracePoints = [];
  const earlyTracedLines = new Set();
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

    if (canTraceBefore) {
      const trace = tracePayload(sessionId, entry, unifiedLine);
      instrumented.push(traceLine(notifierName, trace, indent));
      tracePoints.push(trace);
    }

    instrumented.push(line);

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
    lineMap: mapEntries,
  };
}
