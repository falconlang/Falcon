export function xmlChunks(xml) {
  return String(xml || '')
    .split('\0')
    .map(chunk => chunk.trim())
    .filter(Boolean);
}

export function blocklyTargetXmlWithContextParts(xml, targetIndex = -1) {
  const chunks = xmlChunks(xml);
  if (!chunks.length) throw new Error('No Blockly XML was generated');

  const index = Number.isInteger(targetIndex) && targetIndex >= 0 && targetIndex < chunks.length
    ? targetIndex
    : chunks.length - 1;

  return {
    targetXml: chunks[index],
    contextXml: chunks.filter((_, chunkIndex) => chunkIndex !== index).join('\0'),
    targetIndex: index,
  };
}

export function blocklyTargetXmlsWithContextParts(xml, targetIndexes = []) {
  const chunks = xmlChunks(xml);
  if (!chunks.length) throw new Error('No Blockly XML was generated');

  const selected = new Set();
  for (const index of targetIndexes) {
    if (Number.isInteger(index) && index >= 0 && index < chunks.length) {
      selected.add(index);
    }
  }

  return {
    targetXml: chunks.filter((_, index) => selected.has(index)).join('\0'),
    contextXml: chunks.filter((_, index) => !selected.has(index)).join('\0'),
    targetIndexes: chunks.map((_, index) => index).filter(index => selected.has(index)),
  };
}

function firstCodeLineInRange(source, startLine, endLine) {
  const lines = String(source ?? '').split('\n');
  for (let line = startLine; line <= endLine; line += 1) {
    const trimmed = (lines[line - 1] || '').trim();
    if (!trimmed || trimmed.startsWith('//')) continue;
    return line;
  }
  return null;
}

export function topLevelIndexesFullyContainedInLineRange(source, lineNumbers, startLine, endLine) {
  const lines = String(source ?? '').split('\n');
  const starts = Array.isArray(lineNumbers) ? lineNumbers : [];
  const selectionStart = Math.max(1, Math.trunc(Number(startLine) || 1));
  const selectionEnd = Math.max(selectionStart, Math.trunc(Number(endLine) || selectionStart));
  const firstCodeLine = firstCodeLineInRange(source, selectionStart, selectionEnd);
  if (firstCodeLine === null) return [];

  return starts
    .map((line, index) => {
      const exprStart = Math.trunc(Number(line) || 0);
      if (exprStart < 1) return null;
      const nextStart = Math.trunc(Number(starts[index + 1]) || 0);
      let exprEnd = nextStart > exprStart ? nextStart - 1 : lines.length;
      while (exprEnd > exprStart) {
        const trimmed = (lines[exprEnd - 1] || '').trim();
        if (trimmed && !trimmed.startsWith('//')) break;
        exprEnd -= 1;
      }
      return { index, startLine: exprStart, endLine: exprEnd };
    })
    .filter(Boolean)
    .filter(entry => (
      entry.startLine >= selectionStart
      && entry.endLine <= selectionEnd
      && entry.endLine >= firstCodeLine
    ))
    .map(entry => entry.index);
}
