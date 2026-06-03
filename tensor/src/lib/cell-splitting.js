function codePartBeforeLineComment(line) {
  let out = '';
  let inString = false;
  let escaped = false;

  for (let i = 0; i < line.length; i += 1) {
    const ch = line[i];
    const next = line[i + 1];

    if (inString) {
      out += ch;
      if (escaped) {
        escaped = false;
      } else if (ch === '\\') {
        escaped = true;
      } else if (ch === '"') {
        inString = false;
      }
      continue;
    }

    if (ch === '"') {
      inString = true;
      out += ch;
      continue;
    }

    if (ch === '/' && next === '/') break;
    out += ch;
  }

  return out;
}

function updateDepthFromLine(line, depth) {
  const code = codePartBeforeLineComment(line);
  let inString = false;
  let escaped = false;
  let nextDepth = depth;

  for (let i = 0; i < code.length; i += 1) {
    const ch = code[i];

    if (inString) {
      if (escaped) {
        escaped = false;
      } else if (ch === '\\') {
        escaped = true;
      } else if (ch === '"') {
        inString = false;
      }
      continue;
    }

    if (ch === '"') {
      inString = true;
    } else if (ch === '{' || ch === '[' || ch === '(') {
      nextDepth += 1;
    } else if (ch === '}' || ch === ']' || ch === ')') {
      nextDepth = Math.max(0, nextDepth - 1);
    }
  }

  return nextDepth;
}

function isCommentOrBlank(line) {
  const trimmed = String(line ?? '').trim();
  return !trimmed || trimmed.startsWith('//');
}

function adjustedStartLine(lines, lineNumber, minLine = 1) {
  let start = lineNumber;
  while (start > minLine && isCommentOrBlank(lines[start - 2])) {
    start -= 1;
  }
  return start;
}

function lineStartsExpressionBody(code) {
  return /^(func|global)\b/.test(code) && /(?:^|[^=!<>])=\s*$/.test(code);
}

function inferTopLevelLineNumbers(lines) {
  const starts = [];
  let depth = 0;
  let pendingExpressionBody = false;

  for (let i = 0; i < lines.length; i += 1) {
    const line = lines[i];
    const code = codePartBeforeLineComment(line).trim();
    const depthBefore = depth;
    if (depthBefore === 0 && code && !code.startsWith('}') && !pendingExpressionBody) {
      starts.push(i + 1);
    }
    depth = updateDepthFromLine(line, depth);
    if (code) {
      if (depthBefore === 0 && lineStartsExpressionBody(code)) {
        pendingExpressionBody = true;
      } else if (pendingExpressionBody) {
        pendingExpressionBody = depth > 0;
      }
    }
  }

  return starts;
}

function firstCodeLine(chunk) {
  return String(chunk ?? '')
    .split('\n')
    .map(line => codePartBeforeLineComment(line).trim())
    .find(Boolean) || '';
}

function isGlobalChunk(chunk) {
  return /^global\b/.test(firstCodeLine(chunk));
}

function groupGlobalChunksFirst(chunks) {
  const globals = [];
  const rest = [];

  for (const chunk of chunks) {
    if (isGlobalChunk(chunk)) globals.push(chunk);
    else rest.push(chunk);
  }

  return globals.length ? [globals.join('\n\n'), ...rest] : rest;
}

export function splitFalconSourceByTopLevelLines(sourceCode, lineNumbers = []) {
  const source = String(sourceCode ?? '');
  if (!source.trim()) return [];

  const lines = source.split('\n');
  let starts = [...new Set(
    Array.from(lineNumbers ?? [])
      .map(Number)
      .filter(line => Number.isFinite(line) && line >= 1 && line <= lines.length)
      .map(line => Math.trunc(line))
  )].sort((a, b) => a - b);

  if (!starts.length) {
    starts = inferTopLevelLineNumbers(lines);
  }

  if (!starts.length) return [source.trim()];

  const adjustedStarts = starts.map((start, index) =>
    adjustedStartLine(lines, start, index > 0 ? starts[index - 1] + 1 : 1)
  );

  const chunks = [];
  for (let i = 0; i < adjustedStarts.length; i += 1) {
    const startIndex = adjustedStarts[i] - 1;
    const endIndex = i < adjustedStarts.length - 1 ? adjustedStarts[i + 1] - 1 : lines.length;
    const chunk = lines.slice(startIndex, endIndex).join('\n').trim();
    if (chunk) chunks.push(chunk);
  }

  return chunks.length ? groupGlobalChunksFirst(chunks) : [source.trim()];
}

export function splitFalconSourceIntoCells(sourceCode, lineNumbers = []) {
  const parserChunks = splitFalconSourceByTopLevelLines(sourceCode, lineNumbers);
  if (parserChunks.length !== 1) return parserChunks;

  const inferredChunks = splitFalconSourceByTopLevelLines(sourceCode);
  return inferredChunks.length > parserChunks.length ? inferredChunks : parserChunks;
}
