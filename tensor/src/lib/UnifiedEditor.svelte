<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import { cells, updateCellCode } from './stores.js';
  import { falconTokenize, tokensToHtml } from './tokenizer.js';
  import { mistToXmlResult } from './falcon-wasm.js';
  import { blocklyXmlToPng, componentDefinitionsFromDesigner, ensureBlocklyRuntime, copyPngBlobToClipboard, downloadPngBlob } from './blockly-preview.js';

  const AUTO_PAIRS = { '{': '}', '(': ')', '[': ']', '"': '"' };
  const HISTORY_LIMIT = 100;

  let codeEl, highlightEl, wrapEl, editorContainerEl;
  let editorValue = '';
  let history = [];
  let historyIndex = -1;
  let applyingHistory = false;

  // ── Blockly callout ──
  // null when hidden, or { x, y, status: 'loading'|'ready'|'error', imgUrl }
  let callout = null;
  let calloutRunId = 0;
  let calloutDebounceTimer = null;

  const CALLOUT_DEBOUNCE_MS = 280;
  const LINE_HEIGHT = 13 * 1.65; // must match .code-area font-size × line-height
  const CODE_PAD_TOP = 12;        // must match .code-area padding-top
  const CALLOUT_GAP = 8;
  const CALLOUT_PADDING = 12; // 6px × 2

  // Parallel arrays: one entry per code cell in order.
  // cellIds[i]    — the cell's id
  // cellStarts[i] — character offset in editorValue where cell i's code begins
  // Cell i's code occupies editorValue.slice(cellStarts[i], cellStarts[i+1] - 2)
  // The last cell's code occupies editorValue.slice(cellStarts[last])
  // The 2-char gap between cells is the '\n\n' separator.
  let cellIds = [];
  let cellStarts = [];

  $: lineCount = Math.max(editorValue.split('\n').length, 1);
  $: lineNumbers = Array.from({ length: lineCount }, (_, i) => i + 1);
  $: highlightedCode = tokensToHtml(falconTokenize(editorValue));

  function buildContent(cellList) {
    const codeCells = cellList.filter(c => c.type === 'code');
    cellIds = codeCells.map(c => c.id);
    cellStarts = [];
    let offset = 0;
    const parts = codeCells.map(c => {
      cellStarts.push(offset);
      const code = c.code || '';
      offset += code.length + 2; // +2 for the '\n\n' separator that join() inserts
      return code;
    });
    return parts.join('\n\n');
  }

  function adjustCellStarts(prevContent, nextContent) {
    const delta = nextContent.length - prevContent.length;
    if (delta === 0 || cellStarts.length <= 1) return;
    // Find first differing position
    let changeAt = 0;
    const minLen = Math.min(prevContent.length, nextContent.length);
    while (changeAt < minLen && prevContent[changeAt] === nextContent[changeAt]) changeAt++;
    // Shift every cell start that lies strictly after the edit point (skip index 0 — always 0)
    for (let i = 1; i < cellStarts.length; i++) {
      if (cellStarts[i] > changeAt) {
        cellStarts[i] = Math.max(cellStarts[i - 1] + 1, cellStarts[i] + delta);
      }
    }
  }

  function syncToStore() {
    const content = editorValue;
    cellIds.forEach((id, i) => {
      const start = cellStarts[i] ?? 0;
      const end = i < cellIds.length - 1
        ? Math.max(start, (cellStarts[i + 1] ?? content.length) - 2)
        : content.length;
      updateCellCode(id, content.slice(start, end));
    });
  }

  function selectionRange() {
    if (!codeEl) return { start: editorValue.length, end: editorValue.length };
    return { start: codeEl.selectionStart ?? editorValue.length, end: codeEl.selectionEnd ?? editorValue.length };
  }

  function selectionCaret() { return selectionRange().end; }

  function setSelection(start, end = start) {
    if (!codeEl) return;
    const s = Math.max(0, Math.min(start, editorValue.length));
    const e = Math.max(s, Math.min(end, editorValue.length));
    tick().then(() => {
      if (!codeEl) return;
      codeEl.focus({ preventScroll: true });
      codeEl.setSelectionRange(s, e);
    });
  }

  function render(text, caret = null) {
    const prev = editorValue;
    editorValue = text;
    adjustCellStarts(prev, text);
    syncToStore();
    if (caret !== null) setSelection(caret);
  }

  function pushHistory(caret = selectionCaret()) {
    if (applyingHistory) return;
    const current = history[historyIndex];
    if (current?.text === editorValue) { history[historyIndex] = { text: editorValue, caret }; return; }
    history = history.slice(0, historyIndex + 1);
    history.push({ text: editorValue, caret });
    if (history.length > HISTORY_LIMIT) history.shift();
    historyIndex = history.length - 1;
  }

  function restoreHistory(idx) {
    if (idx < 0 || idx >= history.length) return false;
    applyingHistory = true;
    historyIndex = idx;
    const entry = history[historyIndex];
    render(entry.text, entry.caret);
    applyingHistory = false;
    return true;
  }

  function isInsideString(text, offset) {
    let inString = false, inLineComment = false;
    for (let i = 0; i < offset; i++) {
      const ch = text[i], next = text[i + 1];
      if (inLineComment) { if (ch === '\n') inLineComment = false; continue; }
      if (inString) { if (ch === '\\') { i++; continue; } if (ch === '"') inString = false; continue; }
      if (ch === '/' && next === '/') { inLineComment = true; i++; continue; }
      if (ch === '"') inString = true;
    }
    return inString;
  }

  function isPlainTextKey(e) { return !e.ctrlKey && !e.metaKey && !e.altKey; }

  function autoPair(e) {
    if (!isPlainTextKey(e) || !(e.key in AUTO_PAIRS)) return false;
    const text = editorValue;
    const { start, end } = selectionRange();
    const s = Math.max(0, Math.min(start, text.length));
    const en = Math.max(s, Math.min(end, text.length));
    if (isInsideString(text, s)) return false;
    const open = e.key, close = AUTO_PAIRS[open];
    const selected = text.slice(s, en);
    const next = text.slice(0, s) + open + selected + close + text.slice(en);
    const caret = s === en ? s + 1 : en + 2;
    e.preventDefault(); e.stopPropagation();
    render(next, caret); pushHistory(caret);
    return true;
  }

  function skipExistingClose(e) {
    if (!isPlainTextKey(e) || !Object.values(AUTO_PAIRS).includes(e.key)) return false;
    const text = editorValue;
    const { start, end } = selectionRange();
    if (start !== end) return false;
    const caret = Math.max(0, Math.min(start, text.length));
    if (text[caret] !== e.key) return false;
    if (e.key === '"' ? !isInsideString(text, caret) : isInsideString(text, caret)) return false;
    e.preventDefault(); e.stopPropagation();
    setSelection(caret + 1); pushHistory(caret + 1);
    return true;
  }

  function currentLinePrefix(text, offset) {
    const lineStart = text.lastIndexOf('\n', Math.max(0, offset - 1)) + 1;
    return text.slice(lineStart, offset);
  }

  function indentForNewLine(text, offset) {
    const prefix = currentLinePrefix(text, offset);
    const baseIndent = prefix.match(/^\s*/)[0];
    const shouldIncrease = !isInsideString(text, offset) && prefix.trimEnd().endsWith('{');
    return shouldIncrease ? `${baseIndent}  ` : baseIndent;
  }

  function insertIndentedNewline() {
    const text = editorValue;
    const { start, end } = selectionRange();
    const indent = indentForNewLine(text, start);
    const insert = `\n${indent}`;
    const next = text.slice(0, start) + insert + text.slice(end);
    const caret = start + insert.length;
    render(next, caret); pushHistory(caret);
  }

  function deleteIndentOnlyLine(e) {
    if (e.key !== 'Backspace' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const text = editorValue;
    const { start, end } = selectionRange();
    if (start !== end || start === 0) return false;
    const lineStart = text.lastIndexOf('\n', Math.max(0, start - 1)) + 1;
    const linePrefix = text.slice(lineStart, start);
    if (lineStart === 0 || linePrefix.length === 0 || /\S/.test(linePrefix)) return false;
    const prevEnd = lineStart - 1;
    e.preventDefault(); e.stopPropagation();
    render(text.slice(0, prevEnd) + text.slice(start), prevEnd); pushHistory(prevEnd);
    return true;
  }

  function deleteEmptyPair(e) {
    if (e.key !== 'Backspace' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const text = editorValue;
    const { start, end } = selectionRange();
    if (start !== end || start === 0) return false;
    const open = text[start - 1], close = text[start];
    if (AUTO_PAIRS[open] !== close) return false;
    e.preventDefault(); e.stopPropagation();
    const caret = start - 1;
    render(text.slice(0, caret) + text.slice(start + 1), caret); pushHistory(caret);
    return true;
  }

  function expandBraceOnEnter(e) {
    if (e.key !== 'Enter' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const text = editorValue;
    const { start, end } = selectionRange();
    if (start !== end) return false;
    const caret = Math.max(0, Math.min(start, text.length));
    if (text[caret - 1] !== '{' || text[caret] !== '}') return false;
    if (isInsideString(text, caret)) return false;
    const lineStart = text.lastIndexOf('\n', Math.max(0, caret - 1)) + 1;
    const baseIndent = text.slice(lineStart, caret).match(/^\s*/)[0];
    const innerIndent = `${baseIndent}  `;
    const insert = `\n${innerIndent}\n${baseIndent}`;
    const next = text.slice(0, caret) + insert + text.slice(caret);
    const nextCaret = caret + 1 + innerIndent.length;
    e.preventDefault(); e.stopPropagation();
    render(next, nextCaret); pushHistory(nextCaret);
    return true;
  }

  function selectedLineRange(text, start, end) {
    const activeEnd = end > start && text[end - 1] === '\n' ? end - 1 : end;
    const lineStart = text.lastIndexOf('\n', Math.max(0, start - 1)) + 1;
    const nextBreak = text.indexOf('\n', activeEnd);
    const lineEnd = nextBreak === -1 ? text.length : nextBreak;
    return { lineStart, lineEnd };
  }

  function toggleLineComments() {
    if (!codeEl) return;
    const text = editorValue;
    const { start, end } = selectionRange();
    const s = Math.max(0, Math.min(start, text.length));
    const en = Math.max(s, Math.min(end, text.length));
    const { lineStart, lineEnd } = selectedLineRange(text, s, en);
    const lines = text.slice(lineStart, lineEnd).split('\n');
    const candidates = lines.filter(l => l.trim().length > 0);
    const shouldUncomment = candidates.length > 0 && candidates.every(l => /^\s*\/\//.test(l));
    const nextLines = lines.map(line => {
      if (shouldUncomment) {
        const match = line.match(/^(\s*)\/\/ ?/);
        return match ? match[1] + line.slice(match[0].length) : line;
      }
      const indent = line.match(/^\s*/)[0];
      return `${indent}// ${line.slice(indent.length)}`;
    });
    const toggled = nextLines.join('\n');
    const next = text.slice(0, lineStart) + toggled + text.slice(lineEnd);
    const toggledEnd = lineStart + toggled.length;
    const nextLineStart = next[toggledEnd] === '\n' ? toggledEnd + 1 : toggledEnd;
    render(next, nextLineStart); pushHistory(nextLineStart);
  }

  function isUndoShortcut(e) { return (e.ctrlKey || e.metaKey) && !e.altKey && e.key?.toLowerCase() === 'z' && !e.shiftKey; }
  function isRedoShortcut(e) { const k = e.key?.toLowerCase(); return (e.ctrlKey || e.metaKey) && !e.altKey && (k === 'y' || (k === 'z' && e.shiftKey)); }
  function isCommentShortcut(e) {
    const code = e.keyCode || e.which;
    return (e.ctrlKey || e.metaKey) && !e.altKey && (e.key === '/' || e.key === '?' || e.code === 'Slash' || e.code === 'NumpadDivide' || code === 191 || code === 111);
  }

  function handleKey(e) {
    if (isUndoShortcut(e)) { e.preventDefault(); e.stopPropagation(); restoreHistory(historyIndex - 1); return; }
    if (isRedoShortcut(e)) { e.preventDefault(); e.stopPropagation(); restoreHistory(historyIndex + 1); return; }
    if (isCommentShortcut(e)) { e.preventDefault(); e.stopPropagation(); e.stopImmediatePropagation(); toggleLineComments(); return; }
    if (skipExistingClose(e) || autoPair(e)) return;
    if (deleteEmptyPair(e)) return;
    if (deleteIndentOnlyLine(e)) return;
    if (e.key === 'Tab') {
      e.preventDefault();
      const text = editorValue;
      const { start, end } = selectionRange();
      const next = text.slice(0, start) + '  ' + text.slice(end);
      const caret = start + 2;
      render(next, caret); pushHistory(caret);
      return;
    }
    if (e.key === 'Enter') {
      if (expandBraceOnEnter(e)) return;
      e.preventDefault();
      insertIndentedNewline();
    }
  }

  function syncScroll() {
    if (!codeEl || !highlightEl) return;
    highlightEl.scrollLeft = codeEl.scrollLeft;
    highlightEl.scrollTop = codeEl.scrollTop;
    updateCalloutY();
  }

  function onInput(e) {
    const prev = editorValue;
    editorValue = e.currentTarget.value;
    adjustCellStarts(prev, editorValue);
    syncToStore();
    dismissCallout();
    pushHistory();
    syncScroll();
  }

  // ── Callout helpers ──

  function lineOfOffset(offset) {
    const boundedOffset = Math.max(0, Math.min(offset, editorValue.length));
    return (editorValue.slice(0, boundedOffset).match(/\n/g) || []).length + 1;
  }

  function selectionLineRange(selStart, selEnd) {
    const start = Math.max(0, Math.min(selStart, editorValue.length));
    const end = Math.max(start, Math.min(selEnd, editorValue.length));
    const activeEnd = end > start && editorValue[end - 1] === '\n' ? end - 1 : end;
    const startLine = lineOfOffset(start);
    const endLine = Math.max(startLine, lineOfOffset(activeEnd));
    return { startLine, endLine };
  }

  function firstCodeLineInSelection(startLine, endLine) {
    const lines = editorValue.split('\n');
    for (let line = startLine; line <= endLine; line += 1) {
      const trimmed = (lines[line - 1] || '').trim();
      if (!trimmed || trimmed.startsWith('//')) continue;
      return line;
    }
    return null;
  }

  function topLevelIndexForSelection(lineNumbers, startLine, endLine) {
    const firstCodeLine = firstCodeLineInSelection(startLine, endLine);
    if (firstCodeLine === null) return -1;
    return lineNumbers.findIndex(ln => ln === firstCodeLine);
  }

  function calloutY(startLine, endLine) {
    if (!codeEl || !editorContainerEl) return 0;
    const rect = codeEl.getBoundingClientRect();
    const centerLine = (startLine + endLine) / 2;
    return rect.top + CODE_PAD_TOP + (centerLine - 1) * LINE_HEIGHT - codeEl.scrollTop;
  }

  function calloutX(startLine, endLine) {
    if (!codeEl) return { x: 0, overflowed: false };
    const lines = editorValue.split('\n');
    const selLines = lines.slice(startLine - 1, endLine);
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.font = getComputedStyle(codeEl).font;
    const maxPx = Math.max(0, ...selLines.map(l => ctx.measureText(l).width));
    const rect = codeEl.getBoundingClientRect();
    const rawX = rect.left + 4 + maxPx; // 4px matches .code-area padding-left
    const frameRight = wrapEl ? wrapEl.getBoundingClientRect().right : window.innerWidth;
    const overflowed = rawX + CALLOUT_GAP > frameRight;
    return { x: overflowed ? frameRight : rawX, overflowed };
  }

  function updateCalloutY() {
    if (!callout || !codeEl || !editorContainerEl) return;
    const { x, overflowed } = calloutX(callout.startLine, callout.endLine);
    callout = { ...callout, y: calloutY(callout.startLine, callout.endLine), x, overflowed };
  }

  function dismissCallout() {
    ++calloutRunId;
    clearTimeout(calloutDebounceTimer);
    if (callout?.imgUrl) URL.revokeObjectURL(callout.imgUrl);
    callout = null;
  }

  function scheduleCalloutCheck() {
    const runId = ++calloutRunId;
    clearTimeout(calloutDebounceTimer);
    calloutDebounceTimer = setTimeout(() => runCalloutCheck(runId), CALLOUT_DEBOUNCE_MS);
  }

  function onDocumentSelectionChange() {
    if (document.activeElement === codeEl) scheduleCalloutCheck();
  }

  async function runCalloutCheck(runId) {
    if (!codeEl || runId !== calloutRunId) return;
    const selStart = codeEl.selectionStart;
    const selEnd   = codeEl.selectionEnd;
    if (selStart === selEnd) { dismissCallout(); return; }

    const { startLine: selectionStartLine, endLine } = selectionLineRange(selStart, selEnd);

    let xmlResult;
    let componentDefinitions;
    try {
      componentDefinitions = await componentDefinitionsFromDesigner();
      if (runId !== calloutRunId) return;
      await ensureBlocklyRuntime();
      if (runId !== calloutRunId) return;
      xmlResult = await mistToXmlResult(editorValue, componentDefinitions);
      if (runId !== calloutRunId) return;
    } catch { return; }

    const { xml, lineNumbers } = xmlResult;
    const idx = topLevelIndexForSelection(lineNumbers, selectionStartLine, endLine);
    if (idx === -1) { if (runId === calloutRunId) callout = null; return; }
    const startLine = lineNumbers[idx];

    const chunks = String(xml || '').split('\0').map(s => s.trim()).filter(Boolean);
    const xmlChunk = chunks[idx];
    if (!xmlChunk) { callout = null; return; }
    const contextXml = chunks.slice(0, idx).join('\0');

    const imgHeight = Math.round((endLine - startLine + 1) * LINE_HEIGHT);
    const { x: cx, overflowed } = calloutX(startLine, endLine);
    const maxWidth = Math.max(80, window.innerWidth - cx - CALLOUT_GAP - CALLOUT_PADDING);
    if (callout?.imgUrl) URL.revokeObjectURL(callout.imgUrl);
    callout = { x: cx, y: calloutY(startLine, endLine), startLine, endLine, imgHeight, maxWidth, overflowed, status: 'loading', imgUrl: null };

    try {
      const result = await blocklyXmlToPng(xmlChunk, componentDefinitions, { contextXml });
      if (runId !== calloutRunId) return;
      callout = { ...callout, status: 'ready', blob: result.blob, imgUrl: URL.createObjectURL(result.blob) };
    } catch {
      if (runId !== calloutRunId) return;
      callout = null;
    }
  }

  onMount(() => {
    editorValue = buildContent($cells);
    pushHistory(0);
    document.addEventListener('selectionchange', onDocumentSelectionChange);
  });

  onDestroy(() => {
    document.removeEventListener('selectionchange', onDocumentSelectionChange);
    dismissCallout();
  });
</script>

<div class="unified-wrap" bind:this={wrapEl}>
  <div class="unified-editor" bind:this={editorContainerEl}>
    <div class="line-nums" aria-hidden="true">
      {#each lineNumbers as n}
        <div>{n}</div>
      {/each}
    </div>
    <div class="code-stack">
      <pre class="code-highlight" aria-hidden="true" bind:this={highlightEl}>{@html highlightedCode}</pre>
      <textarea
        class="code-area"
        aria-multiline="true"
        spellcheck="false"
        wrap="off"
        rows={lineCount}
        bind:this={codeEl}
        value={editorValue}
        on:keydown={handleKey}
        on:input={onInput}
        on:scroll={syncScroll}
        on:select={scheduleCalloutCheck}
        on:mouseup={scheduleCalloutCheck}
        on:keyup={scheduleCalloutCheck}
      ></textarea>
    </div>
  </div>
</div>

{#if callout}
  <div
    class="blocks-callout"
    class:blocks-callout--loading={callout.status === 'loading'}
    style="top:{callout.y}px; left:{callout.x + 8}px; max-width:{callout.maxWidth}px;"
  >
    {#if callout.status === 'loading'}
      <svg class="callout-spinner" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5">
        <circle cx="7" cy="7" r="4.5" stroke-opacity="0.2"/>
        <path d="M11.5 7A4.5 4.5 0 007 2.5"/>
      </svg>
    {:else if callout.status === 'ready' && callout.imgUrl}
      <img class="callout-img" src={callout.imgUrl} alt="Blockly preview" style="max-height:{callout.imgHeight}px;" />
      <div class="callout-actions">
        <button class="callout-action-btn" on:click={() => copyPngBlobToClipboard(callout.blob)} title="Copy blocks as PNG">
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.4"><rect x="4" y="4" width="8" height="8" rx="1.2"/><path d="M2 10V2h8"/></svg>
          Copy blocks
        </button>
        <button class="callout-action-btn" on:click={() => downloadPngBlob(callout.blob)} title="Download blocks as PNG">
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.4"><path d="M7 2v7M4.5 6.5L7 9l2.5-2.5"/><path d="M2 11h10"/></svg>
          Download
        </button>
      </div>
    {/if}
  </div>
{/if}

<style>
  .unified-wrap {
    display: flex;
    flex-direction: column;
    min-height: 100%;
  }

  .unified-editor {
    flex: 1;
    display: flex;
    background: var(--surface);
    overflow: hidden;
  }

  /* ── Blockly callout ── */
  .blocks-callout {
    position: fixed;
    z-index: 80;
    transform: translateY(-50%);
    background: var(--surface);
    border: 1.5px solid var(--border);
    border-radius: 10px;
    padding: 6px;
    pointer-events: auto;
  }

  /* Arrow pointing left — rotated square on left edge */
  .blocks-callout::after {
    content: '';
    position: absolute;
    left: -5px;
    top: 50%;
    width: 8px;
    height: 8px;
    background: var(--surface);
    border-bottom: 1.5px solid var(--border);
    border-left: 1.5px solid var(--border);
    transform: translateY(-50%) rotate(45deg);
    border-radius: 0 0 0 2px;
  }

  .blocks-callout--loading {
    padding: 10px 12px;
  }

  .callout-spinner {
    display: block;
    width: 20px;
    height: 20px;
    animation: callout-spin 0.9s linear infinite;
    color: var(--text-faint);
  }

  @keyframes callout-spin {
    to { transform: rotate(360deg); }
  }

  .callout-img {
    display: block;
    width: auto;
    max-width: 100%;
    height: auto;
    border-radius: 6px;
  }

  .callout-actions {
    display: flex;
    gap: 4px;
    margin-top: 6px;
  }

  .callout-action-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    height: 26px;
    padding: 0 8px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font);
    font-size: 11px;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.1s, color 0.1s, border-color 0.1s;
    white-space: nowrap;
  }

  .callout-action-btn:hover {
    background: var(--cell-active);
    color: var(--text);
    border-color: var(--text-faint);
  }

  .callout-action-btn svg {
    width: 12px;
    height: 12px;
    flex-shrink: 0;
  }
</style>
