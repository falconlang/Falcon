export function splitFalconSourceByTopLevelLines(sourceCode, lineNumbers = []) {
  const source = String(sourceCode ?? '');
  if (!source.trim()) return [];

  const lines = source.split('\n');
  const starts = [...new Set(
    Array.from(lineNumbers ?? [])
      .map(Number)
      .filter(line => Number.isFinite(line) && line >= 1 && line <= lines.length)
      .map(line => Math.trunc(line))
  )].sort((a, b) => a - b);

  if (!starts.length) return [source.trim()];

  if (starts[0] > 1 && lines.slice(0, starts[0] - 1).join('\n').trim()) {
    starts.unshift(1);
  }

  const chunks = [];
  for (let i = 0; i < starts.length; i += 1) {
    const startIndex = starts[i] - 1;
    const endIndex = i < starts.length - 1 ? starts[i + 1] - 1 : lines.length;
    const chunk = lines.slice(startIndex, endIndex).join('\n').trim();
    if (chunk) chunks.push(chunk);
  }

  return chunks.length ? chunks : [source.trim()];
}
