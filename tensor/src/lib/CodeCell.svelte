<script>
  import { onDestroy, onMount, tick } from 'svelte';
  import { get } from 'svelte/store';
  import {
    setActive, moveCellById, deleteCellById, showCtx,
    updateCellExecCount, updateCellCode, cells, doItCellId, blocklyPreviewRequest,
    appendDebugLogs, debugCollapsed, companionCommand, liveTestState,
    doItResults, setDoItResult, clearDoItResult,
    debugModeEnabled, debugActiveLocation, debugRuntimeErrors,
  } from './stores.js';
  import { falconTokenize, tokensToHtml } from './tokenizer.js';
  import { falconCellToBlocklyPng, warmBlocklyPreviewRuntime } from './blockly-preview.js';
  import { runCodeDiagnosticResult } from './falcon-wasm.js';

  export let cell;
  export let active = false;

  let editorEl;
  let codeEl;
  let highlightEl;
  let running = false;
  let flashing = false;
  let blocklyActive = false;
  let blocklyStatus = 'idle';
  let blocklyError = '';
  let blocklyPreviewUrl = '';
  let blocklyPreviewBlob = null;
  let blocklySourceSnapshot = '';
  let blocklyPreviewRun = 0;
  let blocklyPreviewHeight = 0;
  let blocklyHeightReleaseTimer = null;
  let handledBlocklyRequestId = null;
  let codeValue = cell.code || '';
  let history = [];
  let historyIndex = -1;
  let applyingHistory = false;
  let localEditPending = false;

  const HISTORY_LIMIT = 100;
  const BLOCKLY_HEIGHT_TRANSITION_MS = 180;
  const CODE_LINE_HEIGHT = 13 * 1.65;
  const CODE_PAD_TOP = 12;
  const AUTO_PAIRS = {
    '{': '}',
    '(': ')',
    '[': ']',
    '"': '"',
  };

  $: isDoItVisible = $doItCellId === cell.id;
  $: isCompanionConnected = $liveTestState.status === 'connected';
  $: doItResult = $doItResults[cell.id] ?? null;
  $: debugActiveLine = $debugModeEnabled && $debugActiveLocation?.cellId === cell.id
    ? $debugActiveLocation.cellLine
    : null;
  $: debugError = $debugModeEnabled ? ($debugRuntimeErrors[cell.id] ?? null) : null;
  $: debugErrorLines = debugError?.message
    ? String(debugError.message).split(/\r?\n/)
    : [];
  $: if (!blocklyActive) {
    const incomingCode = cell.code || '';
    if (localEditPending) {
      if (incomingCode === codeValue) localEditPending = false;
    } else if (incomingCode !== codeValue) {
      codeValue = incomingCode;
    }
  }
  $: sourceLineCount = Math.max(codeValue.split('\n').length, 1);
  $: previewLineCount = Math.max((blocklySourceSnapshot || codeValue).split('\n').length, 1);
  $: lineCount = blocklyActive ? previewLineCount : sourceLineCount;
  $: lineNumbers = Array.from({ length: lineCount }, (_, i) => i + 1);
  $: highlightedCode = tokensToHtml(falconTokenize(codeValue));
  $: blocklyEditorStyle = blocklyPreviewHeight
    ? `height: ${blocklyPreviewHeight}px;`
    : '';
  $: if (
    $blocklyPreviewRequest
    && $blocklyPreviewRequest.cellId === cell.id
    && $blocklyPreviewRequest.id !== handledBlocklyRequestId
  ) {
    handledBlocklyRequestId = $blocklyPreviewRequest.id;
    previewPastedBlocks($blocklyPreviewRequest.blob);
  }

  function codeText() {
    return codeValue;
  }

  function selectionRange() {
    if (!codeEl) return { start: codeValue.length, end: codeValue.length };
    return {
      start: codeEl.selectionStart ?? codeValue.length,
      end: codeEl.selectionEnd ?? codeValue.length,
    };
  }

  function selectionCaret() {
    return selectionRange().end;
  }

  function setSelection(start, end = start) {
    if (!codeEl) return;
    const safeStart = Math.max(0, Math.min(start, codeValue.length));
    const safeEnd = Math.max(safeStart, Math.min(end, codeValue.length));

    tick().then(() => {
      if (!codeEl) return;
      codeEl.focus({ preventScroll: true });
      codeEl.setSelectionRange(safeStart, safeEnd);
    });
  }

  function renderCode(text, caret = null) {
    codeValue = text;
    localEditPending = true;
    updateCellCode(cell.id, text);
    if (caret !== null) setSelection(caret);
  }

  function pushHistory(caret = selectionCaret()) {
    if (applyingHistory) return;
    const text = codeValue;
    const current = history[historyIndex];

    if (current?.text === text) {
      history[historyIndex] = { text, caret };
      return;
    }

    history = history.slice(0, historyIndex + 1);
    history.push({ text, caret });
    if (history.length > HISTORY_LIMIT) history.shift();
    historyIndex = history.length - 1;
  }

  function restoreHistory(nextIndex) {
    if (nextIndex < 0 || nextIndex >= history.length) return false;
    applyingHistory = true;
    historyIndex = nextIndex;
    const entry = history[historyIndex];
    renderCode(entry.text, entry.caret);
    applyingHistory = false;
    return true;
  }

  function undoCode() {
    return restoreHistory(historyIndex - 1);
  }

  function redoCode() {
    return restoreHistory(historyIndex + 1);
  }

  function isUndoShortcut(e) {
    return (e.ctrlKey || e.metaKey) && !e.altKey && e.key?.toLowerCase() === 'z' && !e.shiftKey;
  }

  function isRedoShortcut(e) {
    const key = e.key?.toLowerCase();
    return (e.ctrlKey || e.metaKey)
      && !e.altKey
      && (key === 'y' || (key === 'z' && e.shiftKey));
  }

  function isCommentShortcut(e) {
    const legacyCode = e.keyCode || e.which;
    return (e.ctrlKey || e.metaKey)
      && !e.altKey
      && (
        e.key === '/'
        || e.key === '?'
        || e.code === 'Slash'
        || e.code === 'NumpadDivide'
        || legacyCode === 191
        || legacyCode === 111
      );
  }

  function isPlainTextKey(e) {
    return !e.ctrlKey && !e.metaKey && !e.altKey;
  }

  function isInsideString(text, offset) {
    let inString = false;
    let inLineComment = false;

    for (let i = 0; i < offset; i += 1) {
      const ch = text[i];
      const next = text[i + 1];

      if (inLineComment) {
        if (ch === '\n') inLineComment = false;
        continue;
      }

      if (inString) {
        if (ch === '\\') {
          i += 1;
          continue;
        }
        if (ch === '"') inString = false;
        continue;
      }

      if (ch === '/' && next === '/') {
        inLineComment = true;
        i += 1;
        continue;
      }
      if (ch === '"') inString = true;
    }

    return inString;
  }

  function autoPair(e) {
    if (!isPlainTextKey(e) || !(e.key in AUTO_PAIRS)) return false;
    const text = codeText();
    const selection = selectionRange();
    const start = Math.max(0, Math.min(selection.start, text.length));
    const end = Math.max(start, Math.min(selection.end, text.length));
    if (isInsideString(text, start)) return false;

    const open = e.key;
    const close = AUTO_PAIRS[open];
    const selectedText = text.slice(start, end);
    const nextText = text.slice(0, start) + open + selectedText + close + text.slice(end);
    const nextCaret = start === end ? start + 1 : end + 2;

    e.preventDefault();
    e.stopPropagation();
    renderCode(nextText, nextCaret);
    pushHistory(nextCaret);
    return true;
  }

  function skipExistingClose(e) {
    if (!isPlainTextKey(e) || !Object.values(AUTO_PAIRS).includes(e.key)) return false;
    const text = codeText();
    const selection = selectionRange();
    if (selection.start !== selection.end) return false;

    const caret = Math.max(0, Math.min(selection.start, text.length));
    if (text[caret] !== e.key) return false;
    if (e.key === '"' ? !isInsideString(text, caret) : isInsideString(text, caret)) return false;

    e.preventDefault();
    e.stopPropagation();
    setSelection(caret + 1);
    pushHistory(caret + 1);
    return true;
  }

  function currentLineIndent(text, offset) {
    const lineStart = text.lastIndexOf('\n', Math.max(0, offset - 1)) + 1;
    return text.slice(lineStart, offset).match(/^\s*/)[0];
  }

  function currentLinePrefix(text, offset) {
    const lineStart = text.lastIndexOf('\n', Math.max(0, offset - 1)) + 1;
    return text.slice(lineStart, offset);
  }

  function indentForNewLine(text, offset) {
    const prefix = currentLinePrefix(text, offset);
    const baseIndent = prefix.match(/^\s*/)[0];
    const shouldIncreaseIndent = !isInsideString(text, offset) && prefix.trimEnd().endsWith('{');
    return shouldIncreaseIndent ? `${baseIndent}  ` : baseIndent;
  }

  function insertIndentedNewline() {
    const text = codeText();
    const { start, end } = selectionRange();
    const indent = indentForNewLine(text, start);
    const insert = `\n${indent}`;
    const nextText = text.slice(0, start) + insert + text.slice(end);
    const nextCaret = start + insert.length;

    renderCode(nextText, nextCaret);
    pushHistory(nextCaret);
  }

  function deleteIndentOnlyLine(e) {
    if (e.key !== 'Backspace' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const text = codeText();
    const { start, end } = selectionRange();
    if (start !== end || start === 0) return false;

    const lineStart = text.lastIndexOf('\n', Math.max(0, start - 1)) + 1;
    const linePrefix = text.slice(lineStart, start);
    if (lineStart === 0 || linePrefix.length === 0 || /\S/.test(linePrefix)) return false;

    const previousLineEnd = lineStart - 1;
    const nextText = text.slice(0, previousLineEnd) + text.slice(start);

    e.preventDefault();
    e.stopPropagation();
    renderCode(nextText, previousLineEnd);
    pushHistory(previousLineEnd);
    return true;
  }

  function deleteEmptyPair(e) {
    if (e.key !== 'Backspace' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const text = codeText();
    const { start, end } = selectionRange();
    if (start !== end || start === 0) return false;

    const open = text[start - 1];
    const close = text[start];
    if (AUTO_PAIRS[open] !== close) return false;

    const nextText = text.slice(0, start - 1) + text.slice(start + 1);
    const nextCaret = start - 1;

    e.preventDefault();
    e.stopPropagation();
    renderCode(nextText, nextCaret);
    pushHistory(nextCaret);
    return true;
  }

  function expandBraceOnEnter(e) {
    if (e.key !== 'Enter' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const text = codeText();
    const selection = selectionRange();
    if (selection.start !== selection.end) return false;

    const caret = Math.max(0, Math.min(selection.start, text.length));
    if (text[caret - 1] !== '{' || text[caret] !== '}') return false;
    if (isInsideString(text, caret)) return false;

    const baseIndent = currentLineIndent(text, caret);
    const innerIndent = `${baseIndent}  `;
    const insert = `\n${innerIndent}\n${baseIndent}`;
    const nextText = text.slice(0, caret) + insert + text.slice(caret);
    const nextCaret = caret + 1 + innerIndent.length;

    e.preventDefault();
    e.stopPropagation();
    renderCode(nextText, nextCaret);
    pushHistory(nextCaret);
    return true;
  }

  function hasSelectionInCode() {
    if (!codeEl) return false;
    return document.activeElement === codeEl && codeEl.selectionStart !== codeEl.selectionEnd;
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
    const text = codeText();
    const selection = selectionRange();
    const start = Math.max(0, Math.min(selection.start, text.length));
    const end = Math.max(start, Math.min(selection.end, text.length));
    const { lineStart, lineEnd } = selectedLineRange(text, start, end);
    const lines = text.slice(lineStart, lineEnd).split('\n');
    const uncommentCandidates = lines.filter(line => line.trim().length > 0);
    const shouldUncomment = uncommentCandidates.length > 0
      && uncommentCandidates.every(line => /^\s*\/\//.test(line));

    const nextLines = [];

    for (const line of lines) {
      if (shouldUncomment) {
        const match = line.match(/^(\s*)\/\/ ?/);
        if (match) {
          nextLines.push(match[1] + line.slice(match[0].length));
        } else {
          nextLines.push(line);
        }
      } else {
        const indent = line.match(/^\s*/)[0];
        nextLines.push(`${indent}// ${line.slice(indent.length)}`);
      }
    }

    const toggledText = nextLines.join('\n');
    const nextText = text.slice(0, lineStart) + toggledText + text.slice(lineEnd);
    const toggledEnd = lineStart + toggledText.length;
    const nextLineStart = nextText[toggledEnd] === '\n' ? toggledEnd + 1 : toggledEnd;

    renderCode(nextText, nextLineStart);
    pushHistory(nextLineStart);
  }

  function onInput(e) {
    codeValue = e.currentTarget.value;
    localEditPending = true;
    updateCellCode(cell.id, codeValue);
    pushHistory();
    syncEditorScroll();
  }

  function syncDoItSelection() {
    if (!codeEl) return;
    doItCellId.set(codeEl.selectionStart !== codeEl.selectionEnd ? cell.id : null);
  }

  function syncEditorScroll() {
    if (!codeEl || !highlightEl) return;
    highlightEl.scrollLeft = codeEl.scrollLeft;
    highlightEl.scrollTop = codeEl.scrollTop;
  }

  function clearBlocklyPreviewAsset() {
    if (blocklyPreviewUrl) URL.revokeObjectURL(blocklyPreviewUrl);
    blocklyPreviewUrl = '';
    blocklyPreviewBlob = null;
  }

  function clearBlocklyHeightReleaseTimer() {
    if (!blocklyHeightReleaseTimer) return;
    clearTimeout(blocklyHeightReleaseTimer);
    blocklyHeightReleaseTimer = null;
  }

  function captureBlocklyPreviewHeight() {
    clearBlocklyHeightReleaseTimer();
    const rect = editorEl?.getBoundingClientRect?.();
    blocklyPreviewHeight = rect?.height ? Math.max(1, Math.ceil(rect.height)) : 0;
  }

  function previewImageTargetHeight(img) {
    const naturalWidth = img?.naturalWidth || 0;
    const naturalHeight = img?.naturalHeight || 0;
    if (!editorEl || !naturalWidth || !naturalHeight) return 0;

    const pane = editorEl.querySelector('.blockly-preview-pane');
    const styles = pane ? getComputedStyle(pane) : null;
    const paddingLeft = Number.parseFloat(styles?.paddingLeft || '0') || 0;
    const paddingRight = Number.parseFloat(styles?.paddingRight || '0') || 0;
    const paddingTop = Number.parseFloat(styles?.paddingTop || '0') || 0;
    const paddingBottom = Number.parseFloat(styles?.paddingBottom || '0') || 0;
    const availableWidth = Math.max(1, editorEl.clientWidth - paddingLeft - paddingRight);
    const widthScale = Math.min(1, availableWidth / naturalWidth);
    const viewportLimit = Math.min(360, Math.max(96, Math.floor(window.innerHeight * 0.56)));
    const imageHeight = Math.min(Math.ceil(naturalHeight * widthScale), viewportLimit);

    return Math.max(1, Math.ceil(imageHeight + paddingTop + paddingBottom));
  }

  function onBlocklyPreviewImageLoad(e) {
    if (!blocklyActive) return;
    const nextHeight = previewImageTargetHeight(e.currentTarget);
    if (!nextHeight || Math.abs(nextHeight - blocklyPreviewHeight) < 1) return;
    blocklyPreviewHeight = nextHeight;
  }

  function codeEditorTargetHeight() {
    const codeHeight = codeEl?.scrollHeight || 0;
    const lineNumbersHeight = editorEl?.querySelector('.line-nums')?.scrollHeight || 0;
    return Math.max(1, Math.ceil(codeHeight), Math.ceil(lineNumbersHeight));
  }

  function releaseBlocklyHeightAfterTransition(runId) {
    clearBlocklyHeightReleaseTimer();
    blocklyHeightReleaseTimer = setTimeout(() => {
      if (runId === blocklyPreviewRun && !blocklyActive) blocklyPreviewHeight = 0;
      blocklyHeightReleaseTimer = null;
    }, BLOCKLY_HEIGHT_TRANSITION_MS + 40);
  }

  function restoreCodeAfterBlocklyPreview() {
    const restored = blocklySourceSnapshot;
    const runId = ++blocklyPreviewRun;
    blocklyActive = false;
    blocklyStatus = 'idle';
    blocklyError = '';
    clearBlocklyPreviewAsset();
    codeValue = restored;
    blocklySourceSnapshot = '';
    localEditPending = true;
    updateCellCode(cell.id, restored);
    tick().then(() => {
      syncEditorScroll();
      if (!blocklyPreviewHeight) return;
      blocklyPreviewHeight = codeEditorTargetHeight();
      releaseBlocklyHeightAfterTransition(runId);
    });
  }

  async function startBlocklyPreview() {
    if (blocklyActive || blocklyStatus === 'loading') return;
    const source = codeValue || cell.code || '';
    if (!source.trim()) return;

    setActive(cell.id);
    const runId = ++blocklyPreviewRun;
    blocklySourceSnapshot = source;
    captureBlocklyPreviewHeight();
    blocklyActive = true;
    blocklyStatus = 'loading';
    blocklyError = '';
    clearBlocklyPreviewAsset();
    codeValue = '';
    doItCellId.set(null);
    await tick();
    syncEditorScroll();

    try {
      const result = await falconCellToBlocklyPng(cell.id, source);
      if (runId !== blocklyPreviewRun || !blocklyActive) return;
      blocklyPreviewBlob = result.blob;
      blocklyPreviewUrl = URL.createObjectURL(result.blob);
      blocklyStatus = 'ready';
    } catch (e) {
      if (runId !== blocklyPreviewRun) return;
      blocklyError = e?.message || String(e);
      restoreCodeAfterBlocklyPreview();
      console.error('[blockly-preview]', e);
    }
  }

  function stopBlocklyPreview() {
    if (!blocklyActive && blocklyStatus !== 'loading') return;
    restoreCodeAfterBlocklyPreview();
  }

  function toggleBlocklyPreview() {
    if (blocklyActive || blocklyStatus === 'loading') stopBlocklyPreview();
    else startBlocklyPreview();
  }

  function previewPastedBlocks(blob) {
    if (!blob) return;
    setActive(cell.id);
    blocklyPreviewRun += 1;
    clearBlocklyPreviewAsset();
    if (!blocklyActive || !blocklyPreviewHeight) captureBlocklyPreviewHeight();
    blocklySourceSnapshot = blocklyActive
      ? blocklySourceSnapshot
      : (codeValue || cell.code || '');
    blocklyPreviewBlob = blob;
    blocklyPreviewUrl = URL.createObjectURL(blob);
    blocklyStatus = 'ready';
    blocklyError = '';
    blocklyActive = true;
    codeValue = '';
    doItCellId.set(null);
    tick().then(syncEditorScroll);
  }

  function selectedSource() {
    if (!codeEl) return '';
    const start = codeEl.selectionStart ?? 0;
    const end = codeEl.selectionEnd ?? 0;
    if (start === end) return '';
    return codeValue.slice(Math.min(start, end), Math.max(start, end));
  }

  function appendRunOutput(result, sourceLabel) {
    const entries = [];
    for (const line of result.output || []) {
      entries.push({ level: 'info', source: sourceLabel, message: line });
    }

    if (!result.ok) {
      const diagnostics = result.diagnostics?.length
        ? result.diagnostics
        : [{ message: result.error || 'Falcon execution failed' }];
      for (const diagnostic of diagnostics) {
        const location = diagnostic.line
          ? `:${diagnostic.line}${diagnostic.column ? `:${diagnostic.column}` : ''}`
          : '';
        entries.push({
          level: 'error',
          source: `${sourceLabel}${location}`,
          message: diagnostic.message || result.error || 'Falcon execution failed',
        });
      }
    }

    if (entries.length) {
      appendDebugLogs(entries);
      debugCollapsed.set(false);
    }
  }

  async function runCell(sourceOverride = null, sourceLabel = `Cell ${cell.id}`) {
    if (running) return;
    const source = typeof sourceOverride === 'string'
      ? sourceOverride
      : (blocklyActive ? blocklySourceSnapshot : codeValue);
    if (!source.trim()) return;

    running = true;
    const icon = document.getElementById(`run-icon-${cell.id}`);
    if (icon) {
      icon.innerHTML = `<circle cx="7" cy="7" r="4" fill="none" stroke="currentColor" stroke-width="1.5" stroke-dasharray="10" stroke-dashoffset="10" class="running"/>`;
    }

    try {
      const result = await runCodeDiagnosticResult(source);
      appendRunOutput(result, sourceLabel);
      updateCellExecCount(cell.id);
      flashing = true;
      setTimeout(() => { flashing = false; }, 400);
    } catch (e) {
      appendRunOutput({
        ok: false,
        error: e?.message || String(e),
        diagnostics: [],
        output: [],
      }, sourceLabel);
    } finally {
      running = false;
      if (icon) icon.innerHTML = `<path d="M2 2l10 5-10 5V2z"/>`;
    }
  }

  async function runDoIt(source) {
    if (running) return;
    running = true;
    try {
      const result = await runCodeDiagnosticResult(source);
      const lines = [];
      for (const line of result.output || []) lines.push(line);
      if (!result.ok) {
        const diagnostics = result.diagnostics?.length
          ? result.diagnostics
          : [{ message: result.error || 'Execution failed' }];
        for (const d of diagnostics) {
          const loc = d.line ? `:${d.line}${d.column ? `:${d.column}` : ''}` : '';
          lines.push(`${loc ? loc + ' ' : ''}${d.message}`);
        }
      }
      setDoItResult(cell.id, lines, result.ok);
      updateCellExecCount(cell.id);
      flashing = true;
      setTimeout(() => { flashing = false; }, 400);
    } catch (e) {
      setDoItResult(cell.id, [e?.message || String(e)], false);
    } finally {
      running = false;
    }
  }

  function runSelection() {
    const source = selectedSource();
    if (!source.trim()) return;
    if (isCompanionConnected) {
      companionCommand.set({
        type: 'do-it',
        source,
        label: `Selection ${cell.id}`,
        cellId: cell.id,
      });
      doItCellId.set(null);
      return;
    }
    runDoIt(source);
    doItCellId.set(null);
  }

  function handleDoItMouseDown(e) {
    e.preventDefault();
    e.stopPropagation();
    runSelection();
  }

  function handleDoItClick(e) {
    e.stopPropagation();
    if (e.detail === 0) runSelection();
  }

  function handleCodeKey(e) {
    if (isUndoShortcut(e)) {
      e.preventDefault();
      e.stopPropagation();
      undoCode();
      return;
    }
    if (isRedoShortcut(e)) {
      e.preventDefault();
      e.stopPropagation();
      redoCode();
      return;
    }
    if (isCommentShortcut(e)) {
      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation();
      toggleLineComments();
      return;
    }
    if (skipExistingClose(e) || autoPair(e)) return;
    if (deleteEmptyPair(e)) return;
    if (deleteIndentOnlyLine(e)) return;
    if (e.key === 'Tab') {
      e.preventDefault();
      const text = codeText();
      const { start, end } = selectionRange();
      const nextText = text.slice(0, start) + '  ' + text.slice(end);
      const nextCaret = start + 2;
      renderCode(nextText, nextCaret);
      pushHistory(nextCaret);
      return;
    }
    if (e.key === 'Enter') {
      if (e.shiftKey) {
        e.preventDefault();
        runCell();
        const currentCells = get(cells);
        const idx = currentCells.findIndex(c => c.id === cell.id);
        if (idx < currentCells.length - 1) setActive(currentCells[idx + 1].id);
      } else {
        if (expandBraceOnEnter(e)) return;
        e.preventDefault();
        insertIndentedNewline();
      }
    }
  }

  onMount(() => {
    pushHistory(0);

    function handleDocumentKey(e) {
      if (e.defaultPrevented) return;
      const wantsUndo = isUndoShortcut(e);
      const wantsRedo = isRedoShortcut(e);
      const wantsComment = isCommentShortcut(e);
      if (!wantsUndo && !wantsRedo && !wantsComment) return;

      const targetEditor = e.target?.closest?.('input, textarea, select, [contenteditable="true"]');
      if (targetEditor && targetEditor !== codeEl) return;
      if (!active && !hasSelectionInCode() && document.activeElement !== codeEl) return;

      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation();
      setActive(cell.id);
      if (wantsUndo) undoCode();
      else if (wantsRedo) redoCode();
      else toggleLineComments();
    }

    document.addEventListener('keydown', handleDocumentKey, true);
    return () => document.removeEventListener('keydown', handleDocumentKey, true);
  });

  onDestroy(() => {
    blocklyPreviewRun += 1;
    blocklyActive = false;
    blocklyStatus = 'idle';
    clearBlocklyHeightReleaseTimer();
    clearBlocklyPreviewAsset();
  });

  function activateCellFromKeyboard(e) {
    if (e.target !== e.currentTarget) return;
    if (e.key !== 'Enter' && e.key !== ' ') return;
    e.preventDefault();
    setActive(cell.id);
  }

  function debugLineTop(line) {
    return `${CODE_PAD_TOP + (Math.max(1, Number(line) || 1) - 1) * CODE_LINE_HEIGHT}px`;
  }
</script>

<div
  class="cell code-cell"
  class:active
  class:flash={flashing}
  class:debug-active-cell={Boolean(debugActiveLine)}
  class:debug-error-cell={Boolean(debugError)}
  data-debug-active-line={debugActiveLine || undefined}
  data-debug-error-line={debugError?.cellLine || undefined}
  id="cell-{cell.id}"
  role="button"
  tabindex="0"
  on:click={() => setActive(cell.id)}
  on:keydown={activateCellFromKeyboard}
  on:contextmenu={(e) => showCtx(e, cell.id)}
>
  <div class="cell-gutter">
    <button class="gutter-btn" title="Move up" on:click|stopPropagation={() => moveCellById(cell.id, -1)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 8l4-4 4 4"/></svg>
    </button>
    <button class="gutter-btn" title="Move down" on:click|stopPropagation={() => moveCellById(cell.id, 1)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4l4 4 4-4"/></svg>
    </button>
    <button class="gutter-btn" title="Delete" on:click|stopPropagation={() => deleteCellById(cell.id)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 3h8M5 3V2h2v1M9 3l-.6 6H3.6L3 3"/></svg>
    </button>
  </div>

  <div class="cell-header">
    <button class="run-btn" title="Run (Shift+Enter)" on:click|stopPropagation={() => runCell()}>
      <svg id="run-icon-{cell.id}" viewBox="0 0 14 14" fill="currentColor">
        <path d="M2 2l10 5-10 5V2z"/>
      </svg>
    </button>
    <span class="cell-type-badge">fc</span>
    <span class="exec-count">[{cell.execCount ?? ' '}]</span>
    <div class="cell-header-spacer"></div>
    <button
      type="button"
      class="cell-menu-btn do-it-btn"
      class:visible={isDoItVisible}
      title="Do it"
      on:mousedown={handleDoItMouseDown}
      on:click={handleDoItClick}
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M2 6h8M7 3l3 3-3 3"/></svg>
      Do it
    </button>
    <button
      class="cell-menu-btn blockly-btn"
      class:active={blocklyActive}
      title="Toggle Blockly view"
      on:click|stopPropagation={toggleBlocklyPreview}
      on:pointerenter={warmBlocklyPreviewRuntime}
    >
      <svg viewBox="0 0 100 100" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
        <path stroke="currentColor" stroke-width="4" stroke-linejoin="round" d="M 50 9 C 43.936593 9 39 13.936593 39 20 C 39 22.259221 39.844873 24.249457 41.017578 26 L 26 26 C 22.698375 26 20 28.698375 20 32 L 20 47 L 20 48.859375 A 1.0001 1.0001 0 0 0 21.699219 49.572266 C 23.324036 47.978972 25.540569 47 28 47 C 32.982593 47 37 51.017407 37 56 C 37 60.982593 32.982593 65 28 65 C 25.540569 65 23.324036 64.021028 21.699219 62.427734 A 1.0001 1.0001 0 0 0 20 63.140625 L 20 64 L 20 80 C 20 83.301625 22.698375 86 26 86 L 74 86 C 77.301625 86 80 83.301625 80 80 L 80 32 C 80 28.698375 77.301625 26 74 26 L 58.982422 26 C 60.154932 24.249881 61 22.259603 61 20 C 61 13.936593 56.063407 9 50 9 z M 50 11 C 54.982593 11 59 15.017407 59 20 C 59 22.459431 58.02067 24.675276 56.427734 26.298828 A 1.0001 1.0001 0 0 0 57.140625 28 L 74 28 C 76.220375 28 78 29.779625 78 32 L 78 80 C 78 82.220375 76.220375 84 74 84 L 26 84 C 23.779625 84 22 82.220375 22 80 L 22 64.982422 C 23.750312 66.155107 25.740098 67 28 67 C 34.063407 67 39 62.063407 39 56 C 39 49.936593 34.063407 45 28 45 C 25.740098 45 23.750312 45.844893 22 47.017578 L 22 32 C 22 29.779625 23.779625 28 26 28 L 42.859375 28 A 1.0001 1.0001 0 0 0 43.572266 26.300781 C 41.979101 24.676095 41 22.458333 41 20 C 41 15.017407 45.017407 11 50 11 z"/>
      </svg>
      Blocks
    </button>
    <button class="cell-menu-btn" on:click|stopPropagation={(e) => showCtx(e, cell.id, true)}>
      <svg viewBox="0 0 14 14" fill="currentColor"><circle cx="3" cy="7" r="1.2"/><circle cx="7" cy="7" r="1.2"/><circle cx="11" cy="7" r="1.2"/></svg>
    </button>
  </div>

  <div
    class="code-editor"
    bind:this={editorEl}
    class:blockly-previewing={blocklyActive}
    style={blocklyEditorStyle}
  >
    {#if !blocklyActive && debugActiveLine !== null}
      <div
        class="debug-line-backdrop debug-line-backdrop--active"
        style:top={debugLineTop(debugActiveLine)}
      ></div>
    {/if}
    {#if !blocklyActive && debugError?.cellLine != null}
      <div
        class="debug-line-backdrop debug-line-backdrop--error"
        style:top={debugLineTop(debugError.cellLine)}
      ></div>
    {/if}
    <div class="line-nums" aria-hidden="true">
      {#each lineNumbers as n}
        <div
          class:debug-current-line={debugActiveLine === n}
          class:debug-error-line={debugError?.cellLine === n}
          data-line={n}
        >{n}</div>
      {/each}
    </div>
    <div class="code-stack">
      {#if blocklyActive}
        <div class="blockly-preview-pane">
          {#if blocklyStatus === 'loading'}
            <div class="blockly-preview-state">Converting blocks...</div>
          {:else if blocklyPreviewUrl}
            <img class="blockly-preview-img" src={blocklyPreviewUrl} alt="" on:load={onBlocklyPreviewImageLoad} />
          {:else if blocklyError}
            <div class="blockly-preview-state blockly-preview-error">{blocklyError}</div>
          {/if}
        </div>
      {/if}
      <pre class="code-highlight" aria-hidden="true" bind:this={highlightEl}>{@html highlightedCode}</pre>
      <textarea
        class="code-area"
        aria-multiline="true"
        id="code-{cell.id}"
        spellcheck="false"
        wrap="off"
        rows={lineCount}
        bind:this={codeEl}
        value={codeValue}
        on:focus={() => setActive(cell.id)}
        on:keydown={handleCodeKey}
        on:input={onInput}
        on:scroll={syncEditorScroll}
        on:select={syncDoItSelection}
        on:keyup={syncDoItSelection}
        on:mouseup={syncDoItSelection}
      ></textarea>
    </div>
  </div>

  {#if doItResult}
    <div class="do-it-result" class:do-it-result--error={!doItResult.ok}>
      <span class="do-it-result-arrow" aria-hidden="true">{doItResult.ok ? '→' : '✗'}</span>
      <div class="do-it-result-body">
        {#if doItResult.lines.length}
          {#each doItResult.lines as line}
            <div class="do-it-result-line">{line}</div>
          {/each}
        {:else}
          <div class="do-it-result-line do-it-result-line--empty">ok</div>
        {/if}
      </div>
      <button
        class="do-it-result-dismiss"
        title="Dismiss"
        on:click|stopPropagation={() => clearDoItResult(cell.id)}
      >×</button>
    </div>
  {/if}

  {#if debugError}
    <div class="do-it-result do-it-result--error debug-error-result">
      <span class="do-it-result-arrow" aria-hidden="true">
        <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" width="13" height="13">
          <ellipse cx="7" cy="7.5" rx="3" ry="3.5"/>
          <path d="M7 4V2.5"/>
          <path d="M4 7.5H2M12 7.5h-2"/>
          <path d="M4.5 5.2l-1.2-1.2M9.5 5.2l1.2-1.2"/>
          <path d="M4.5 9.8l-1.2 1.2M9.5 9.8l1.2 1.2"/>
        </svg>
      </span>
      <div class="do-it-result-body">
        {#if debugErrorLines.length}
          {#each debugErrorLines as line}
            <div class="do-it-result-line">{line}</div>
          {/each}
        {:else}
          <div class="do-it-result-line">Runtime error</div>
        {/if}
      </div>
    </div>
  {/if}
</div>
