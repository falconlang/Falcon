<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import {
    activeScreen, cells, replaceCodeCells,
    liveTestState, companionCommand,
    doItResults, clearDoItResult, unifiedSelectionActive,
    debugModeEnabled, debugActiveLocation, debugRuntimeErrors,
  } from './stores.js';
  import { falconTokenize, tokensToHtml } from './tokenizer.js';
  import { mistToXmlResult, runCodeDiagnosticResult } from './falcon-wasm.js';
  import { blocklyXmlToPng, componentDefinitionsFromDesigner, ensureBlocklyRuntime, copyPngBlobToClipboard, downloadPngBlob } from './blockly-preview.js';
  import { splitFalconSourceByTopLevelLines } from './cell-splitting.js';

  const AUTO_PAIRS = { '{': '}', '(': ')', '[': ']', '"': '"' };
  const HISTORY_LIMIT = 100;

  let codeEl, highlightEl, wrapEl, editorContainerEl;
  let editorValue = '';
  let history = [];
  let historyIndex = -1;
  let applyingHistory = false;
  let syncingToStore = false;
  let unsubscribeCells = null;
  let unsubscribeScreen = null;
  let lastCellsSignature = '';

  // ── Blockly callout ──
  // null when hidden, or { x, y, status: 'loading'|'ready'|'error', imgUrl }
  let callout = null;
  let calloutRunId = 0;
  let calloutDebounceTimer = null;

  // ── Do it callout ──
  // null when hidden, or { x, y, startLine, endLine, maxWidth, status: 'running'|'ready', lines, ok }
  let doItCallout = null;
  let doItCalloutEl = null;
  let pendingDoItPos = null;

  $: isCompanionConnected = $liveTestState.status === 'connected';
  $: debugActiveUnifiedLine = $debugModeEnabled ? ($debugActiveLocation?.unifiedLine ?? null) : null;
  $: debugErrorEntries = $debugModeEnabled
    ? Object.values($debugRuntimeErrors || {})
        .filter(error => error?.unifiedLine)
        .sort((a, b) => a.unifiedLine - b.unifiedLine)
    : [];
  $: debugErrorUnifiedLines = new Set(debugErrorEntries.map(error => error.unifiedLine));
  $: debugFirstError = debugErrorEntries[0] ?? null;
  $: debugFirstErrorLine = debugFirstError?.unifiedLine ?? null;
  $: debugFirstErrorLines = debugFirstError?.message
    ? String(debugFirstError.message).split(/\r?\n/)
    : [];

  $: {
    const result = $doItResults['unified'];
    if (result && pendingDoItPos) {
      const pos = doItCalloutPos(pendingDoItPos.startLine, pendingDoItPos.selStart);
      doItCallout = { ...pos, startLine: pendingDoItPos.startLine, endLine: pendingDoItPos.endLine, selStart: pendingDoItPos.selStart, status: 'ready', lines: result.lines, ok: result.ok };
      pendingDoItPos = null;
      clearDoItResult('unified');
    }
  }

  const CALLOUT_DEBOUNCE_MS = 280;
  const LINE_HEIGHT = 13 * 1.65; // must match .code-area font-size × line-height
  const CODE_PAD_TOP = 12;        // must match .code-area padding-top
  const CALLOUT_GAP = 8;
  const CALLOUT_PADDING = 12; // 6px × 2

  $: lineCount = Math.max(editorValue.split('\n').length, 1);
  $: lineNumbers = Array.from({ length: lineCount }, (_, i) => i + 1);
  $: highlightedCode = tokensToHtml(falconTokenize(editorValue));

  function buildContent(cellList) {
    const codeCells = cellList.filter(c => c.type === 'code');
    return codeCells.map(c => c.code || '').join('\n\n');
  }

  function codeCellsSignature(cellList) {
    return JSON.stringify((cellList || [])
      .filter(c => c.type === 'code')
      .map(c => [c.id, c.code || '']));
  }

  function resetHistory(caret = 0) {
    history = [];
    historyIndex = -1;
    pushHistory(Math.max(0, Math.min(caret, editorValue.length)));
  }

  function syncFromCells(cellList, force = false) {
    const signature = codeCellsSignature(cellList);
    if (!force && signature === lastCellsSignature) return;
    const nextContent = buildContent(cellList || []);
    lastCellsSignature = signature;
    editorValue = nextContent;
    dismissCallout();
    resetHistory(0);
    tick().then(syncScroll);
  }

  function syncToStore() {
    syncingToStore = true;
    try {
      replaceCodeCells(editorValue.trim() ? [editorValue] : []);
    } finally {
      syncingToStore = false;
    }
  }

  export async function commitToCells() {
    const source = editorValue;
    if (!source.trim()) {
      syncingToStore = true;
      try {
        replaceCodeCells([]);
      } finally {
        syncingToStore = false;
      }
      return true;
    }

    try {
      const componentDefinitions = await componentDefinitionsFromDesigner();
      const result = await mistToXmlResult(source, componentDefinitions);
      const chunks = splitFalconSourceByTopLevelLines(source, result.lineNumbers);
      syncingToStore = true;
      try {
        replaceCodeCells(chunks);
      } finally {
        syncingToStore = false;
      }
      return true;
    } catch (e) {
      console.warn('[unified-editor] unable to split source by top-level expressions:', e);
      syncingToStore = true;
      try {
        replaceCodeCells([source]);
      } finally {
        syncingToStore = false;
      }
      return false;
    }
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
    editorValue = text;
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
    editorValue = e.currentTarget.value;
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

  function doItCalloutPos(startLine, selStart) {
    if (!codeEl) return { x: 0, y: 0, maxWidth: 280 };
    const rect = codeEl.getBoundingClientRect();

    // Measure the pixel column of the selection start within its line
    const lines = editorValue.split('\n');
    const lineCharStart = lines.slice(0, startLine - 1).reduce((acc, l) => acc + l.length + 1, 0);
    const textToSel = editorValue.slice(lineCharStart, Math.max(lineCharStart, Math.min(selStart ?? lineCharStart, lineCharStart + (lines[startLine - 1] || '').length)));
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.font = getComputedStyle(codeEl).font;
    const colPx = ctx.measureText(textToSel).width;

    // x = left content edge + column offset - horizontal scroll
    const x = rect.left + 4 + colPx - (codeEl.scrollLeft || 0);
    const y = rect.top + CODE_PAD_TOP + (startLine - 1) * LINE_HEIGHT - codeEl.scrollTop;
    const frameRight = wrapEl ? wrapEl.getBoundingClientRect().right : window.innerWidth;
    const maxWidth = Math.max(120, Math.min(320, frameRight - x - 16));
    return { x, y, maxWidth };
  }

  function updateCalloutY() {
    if (!codeEl || !editorContainerEl) return;
    if (callout) {
      const { x, overflowed } = calloutX(callout.startLine, callout.endLine);
      callout = { ...callout, y: calloutY(callout.startLine, callout.endLine), x, overflowed };
    }
    if (doItCallout) {
      const pos = doItCalloutPos(doItCallout.startLine, doItCallout.selStart);
      doItCallout = { ...doItCallout, ...pos };
    }
  }

  function dismissDoItCallout() {
    doItCallout = null;
  }

  function debugLineTop(line) {
    return `${CODE_PAD_TOP + (Math.max(1, Number(line) || 1) - 1) * LINE_HEIGHT}px`;
  }

  function debugMessageTop(line) {
    return `${CODE_PAD_TOP + Math.max(1, Number(line) || 1) * LINE_HEIGHT + 2}px`;
  }

  function onWindowPointerDown(e) {
    if (doItCallout?.status === 'ready' && doItCalloutEl && !doItCalloutEl.contains(e.target)) {
      dismissDoItCallout();
    }
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
    if (selStart === selEnd) { unifiedSelectionActive.set(false); dismissCallout(); return; }
    unifiedSelectionActive.set(true);

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
    let mountedScreen = null;
    unsubscribeCells = cells.subscribe(cellList => {
      const signature = codeCellsSignature(cellList);
      if (syncingToStore) {
        lastCellsSignature = signature;
        return;
      }
      syncFromCells(cellList);
    });
    unsubscribeScreen = activeScreen.subscribe(name => {
      if (mountedScreen === null) {
        mountedScreen = name;
        return;
      }
      if (name === mountedScreen) return;
      mountedScreen = name;
      syncFromCells($cells, true);
    });
    document.addEventListener('selectionchange', onDocumentSelectionChange);
  });

  export async function runDoIt() {
    if (!codeEl) return;
    const selStart = codeEl.selectionStart ?? 0;
    const selEnd = codeEl.selectionEnd ?? 0;
    if (selStart === selEnd) return;
    const start = Math.min(selStart, selEnd);
    const end = Math.max(selStart, selEnd);
    const source = editorValue.slice(start, end);
    if (!source.trim()) return;

    const { startLine, endLine } = selectionLineRange(start, end);
    const pos = doItCalloutPos(startLine, start);
    pendingDoItPos = { ...pos, startLine, endLine, selStart: start };
    doItCallout = { ...pendingDoItPos, status: 'running', lines: [], ok: true };

    if (isCompanionConnected) {
      clearDoItResult('unified');
      companionCommand.set({ type: 'do-it', source, label: 'Do it (Script)', cellId: 'unified' });
      return;
    }

    try {
      const result = await runCodeDiagnosticResult(source);
      const lines = [];
      for (const line of result.output || []) lines.push(line);
      if (!result.ok) {
        const diags = result.diagnostics?.length
          ? result.diagnostics
          : [{ message: result.error || 'Execution failed' }];
        for (const d of diags) {
          const loc = d.line ? `:${d.line}${d.column ? `:${d.column}` : ''}` : '';
          lines.push(`${loc ? loc + ' ' : ''}${d.message}`);
        }
      }
      doItCallout = { ...doItCalloutPos(pendingDoItPos.startLine, pendingDoItPos.selStart), startLine: pendingDoItPos.startLine, endLine: pendingDoItPos.endLine, selStart: pendingDoItPos.selStart, status: 'ready', lines, ok: result.ok };
      pendingDoItPos = null;
    } catch (e) {
      doItCallout = { ...doItCalloutPos(pendingDoItPos.startLine, pendingDoItPos.selStart), startLine: pendingDoItPos.startLine, endLine: pendingDoItPos.endLine, selStart: pendingDoItPos.selStart, status: 'ready', lines: [e?.message || String(e)], ok: false };
      pendingDoItPos = null;
    }
  }

  onDestroy(() => {
    unsubscribeCells?.();
    unsubscribeScreen?.();
    document.removeEventListener('selectionchange', onDocumentSelectionChange);
    unifiedSelectionActive.set(false);
    dismissCallout();
    dismissDoItCallout();
  });
</script>

<div
  class="unified-wrap"
  class:debug-active-editor={Boolean(debugActiveUnifiedLine)}
  class:debug-error-editor={debugFirstErrorLine !== null}
  data-debug-active-line={debugActiveUnifiedLine || undefined}
  bind:this={wrapEl}
>
  <div class="unified-editor" bind:this={editorContainerEl}>
    {#if debugActiveUnifiedLine !== null}
      <div
        class="debug-line-backdrop debug-line-backdrop--active"
        style:top={debugLineTop(debugActiveUnifiedLine)}
      ></div>
    {/if}
    {#each debugErrorEntries as error (error.cellId)}
      <div
        class="debug-line-backdrop debug-line-backdrop--error"
        style:top={debugLineTop(error.unifiedLine)}
      ></div>
    {/each}
    {#if debugFirstError}
      <div
        class="debug-inline-error"
        style:top={debugMessageTop(debugFirstErrorLine)}
      >
        <span class="debug-inline-error-icon" aria-hidden="true">✗</span>
        <div class="debug-inline-error-body">
          {#if debugFirstErrorLines.length}
            {#each debugFirstErrorLines as line}
              <div>{line}</div>
            {/each}
          {:else}
            <div>Runtime error</div>
          {/if}
        </div>
      </div>
    {/if}
    <div class="line-nums" aria-hidden="true">
      {#each lineNumbers as n}
        <div
          class:debug-current-line={debugActiveUnifiedLine === n}
          class:debug-error-line={debugErrorUnifiedLines.has(n)}
          data-line={n}
        >{n}</div>
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
        on:blur={() => unifiedSelectionActive.set(false)}
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

<svelte:window on:pointerdown={onWindowPointerDown} />

{#if doItCallout}
  <div
    bind:this={doItCalloutEl}
    class="do-it-callout"
    class:do-it-callout--ok={doItCallout.ok}
    class:do-it-callout--error={!doItCallout.ok}
    class:do-it-callout--running={doItCallout.status === 'running'}
    style="top:{doItCallout.y}px; left:{doItCallout.x + 8}px; max-width:{doItCallout.maxWidth}px;"
  >
    {#if doItCallout.status === 'running'}
      <svg class="do-it-callout-spinner" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5">
        <circle cx="7" cy="7" r="4.5" stroke-opacity="0.2"/>
        <path d="M11.5 7A4.5 4.5 0 007 2.5"/>
      </svg>
    {:else}
      <span class="do-it-callout-arrow" aria-hidden="true">{doItCallout.ok ? '→' : '✗'}</span>
      <div class="do-it-callout-body">
        {#if doItCallout.lines.length}
          {#each doItCallout.lines as line}
            <div class="do-it-callout-line">{line}</div>
          {/each}
        {:else}
          <div class="do-it-callout-line do-it-callout-line--empty">ok</div>
        {/if}
      </div>
      <button class="do-it-callout-dismiss" title="Dismiss" on:click={dismissDoItCallout}>×</button>
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
    position: relative;
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

  /* ── Do it callout ── */
  .do-it-callout {
    position: fixed;
    z-index: 80;
    /* sits above the first selected line with a small gap */
    transform: translateY(calc(-100% - 6px));
    border-radius: 10px;
    padding: 7px 8px 7px 10px;
    pointer-events: auto;
    display: flex;
    align-items: flex-start;
    gap: 6px;
    font-family: var(--mono);
    font-size: 12px;
    line-height: 1.5;
    min-width: 60px;
  }

  /* Arrow pointing down toward the selection */
  .do-it-callout::after {
    content: '';
    position: absolute;
    bottom: -5px;
    left: 10px;
    width: 8px;
    height: 8px;
    border-bottom: 1.5px solid var(--accent);
    border-right: 1.5px solid var(--accent);
    transform: rotate(45deg);
    border-radius: 0 0 2px 0;
  }

  .do-it-callout--ok {
    background: var(--run-soft);
    border: 1.5px solid var(--accent);
    color: var(--text);
  }
  .do-it-callout--ok::after {
    background: var(--run-soft);
    border-color: var(--accent);
  }

  .do-it-callout--error {
    background: var(--error-soft);
    border: 1.5px solid var(--error);
    color: var(--error);
  }
  .do-it-callout--error::after {
    background: var(--error-soft);
    border-color: var(--error);
  }

  .do-it-callout--running {
    background: var(--surface);
    border: 1.5px solid var(--border);
    padding: 10px 12px;
  }
  .do-it-callout--running::after {
    background: var(--surface);
    border-color: var(--border);
  }

  .do-it-callout-arrow {
    flex-shrink: 0;
    font-weight: 600;
    color: var(--accent);
    user-select: none;
    padding-top: 1px;
  }
  .do-it-callout--error .do-it-callout-arrow { color: var(--error); }

  .do-it-callout-body {
    flex: 1;
    min-width: 0;
  }

  .do-it-callout-line {
    white-space: pre-wrap;
    word-break: break-word;
  }
  .do-it-callout-line--empty {
    color: var(--text-faint);
    font-style: italic;
  }
  .do-it-callout--error .do-it-callout-line { color: var(--error); }

  .do-it-callout-dismiss {
    flex-shrink: 0;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-faint);
    font-size: 14px;
    line-height: 1;
    padding: 1px 3px;
    border-radius: 3px;
    transition: background 0.1s, color 0.1s;
    margin-top: 1px;
  }
  .do-it-callout-dismiss:hover {
    background: var(--border-soft);
    color: var(--text-muted);
  }

  .do-it-callout-spinner {
    display: block;
    width: 16px;
    height: 16px;
    animation: callout-spin 0.9s linear infinite;
    color: var(--text-faint);
  }
</style>
