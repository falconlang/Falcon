function stripLineComment(line) {
  let quote = null;
  let escaped = false;
  for (let i = 0; i < line.length; i += 1) {
    const ch = line[i];
    const next = line[i + 1];
    if (escaped) {
      escaped = false;
      continue;
    }
    if (quote) {
      if (ch === '\\') escaped = true;
      else if (ch === quote) quote = null;
      continue;
    }
    if (ch === '"' || ch === "'") {
      quote = ch;
      continue;
    }
    if (ch === '/' && next === '/') return line.slice(0, i);
  }
  return line;
}

function inferValueType(rawValue) {
  const value = String(rawValue || '').trim();
  if (!value) return 'var';
  if (/^["']/.test(value)) return 'text';
  if (/^\[/.test(value)) return 'list';
  if (/^\{/.test(value)) return 'map';
  if (/^(true|false)\b/i.test(value)) return 'bool';
  if (/^[+-]?(?:\d+(?:\.\d*)?|\.\d+)\b/.test(value)) return 'number';
  if (/^(?:#[0-9a-f]{3,8}|&H[0-9a-f]+)\b/i.test(value)) return 'color';
  return 'var';
}

function parseParamNames(rawParams) {
  return String(rawParams || '')
    .split(',')
    .map(part => part.trim().match(/^([A-Za-z_][A-Za-z0-9_]*)\b/)?.[1])
    .filter(Boolean);
}

export function collectFalconSymbols(cellList) {
  const symbols = [];
  const seen = new Set();
  for (const cell of cellList || []) {
    if (cell?.type !== 'code') continue;
    const lines = String(cell.code || '').split(/\r?\n/);
    lines.forEach((rawLine, index) => {
      const line = stripLineComment(rawLine);
      const lineNumber = index + 1;
      const declaration = line.match(/^\s*(global|local)\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:=\s*(.*))?$/);
      if (declaration) {
        const [, scope, name, value] = declaration;
        symbols.push({
          id: `${cell.id}:${lineNumber}:${scope}:${name}`,
          name,
          kind: inferValueType(value),
          scope,
          cellId: cell.id,
          line: lineNumber,
        });
        return;
      }

      const func = line.match(/^\s*func\s+([A-Za-z_][A-Za-z0-9_]*)\s*\(([^)]*)\)\s*(.*)$/);
      if (func) {
        const [, name, rawParams, tail] = func;
        symbols.push({
          id: `${cell.id}:${lineNumber}:func:${name}`,
          name,
          kind: tail.trim().startsWith('=') ? 'returning' : 'void',
          scope: 'func',
          cellId: cell.id,
          line: lineNumber,
        });
        for (const paramName of parseParamNames(rawParams)) {
          const id = `${cell.id}:${lineNumber}:param:${paramName}`;
          if (seen.has(id)) continue;
          seen.add(id);
          symbols.push({
            id,
            name: paramName,
            kind: 'param',
            scope: 'param',
            cellId: cell.id,
            line: lineNumber,
          });
        }
      }
    });
  }
  return symbols;
}
