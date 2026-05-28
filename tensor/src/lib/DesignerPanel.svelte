<script>
  import { onMount, tick } from 'svelte';
  import {
    clearDebugLogs,
    debugCollapsed,
    debugLogs,
    debugOpenHeight,
    initialDesignCode,
    updateDesignCode,
  } from './stores.js';
  import { schemaTokenize, tokensToHtml } from './tokenizer.js';

  let dsgnCodeEl;
  let dsgnHighlightEl;
  let dbgScrollEl;
  let panelEl;
  let resizeHandleEl;
  let schemaValue = initialDesignCode;
  let designerHistory = [];
  let designerHistoryIndex = -1;
  let applyingDesignerHistory = false;

  let debugDidDrag = false;
  let debugResizing = false;
  let debugResizeStartY = 0;
  let debugResizeStartH = 0;
  let lastDebugLogId = 0;

  const DESIGNER_HISTORY_LIMIT = 100;
  const AUTO_PAIRS = {
    '{': '}',
    '(': ')',
    '[': ']',
    '"': '"',
  };

  $: designerLineCount = Math.max(schemaValue.split('\n').length, 1);
  $: designerLineNumbers = Array.from({ length: designerLineCount }, (_, i) => i + 1);
  $: highlightedSchema = tokensToHtml(schemaTokenize(schemaValue));

  function designerSelectionRange() {
    if (!dsgnCodeEl) return { start: schemaValue.length, end: schemaValue.length };
    return {
      start: dsgnCodeEl.selectionStart ?? schemaValue.length,
      end: dsgnCodeEl.selectionEnd ?? schemaValue.length,
    };
  }

  function setDesignerSelection(start, end = start) {
    if (!dsgnCodeEl) return;
    const safeStart = Math.max(0, Math.min(start, schemaValue.length));
    const safeEnd = Math.max(safeStart, Math.min(end, schemaValue.length));
    tick().then(() => {
      if (!dsgnCodeEl) return;
      dsgnCodeEl.focus({ preventScroll: true });
      dsgnCodeEl.setSelectionRange(safeStart, safeEnd);
    });
  }

  function renderDesigner(text, caret = null) {
    schemaValue = text;
    updateDesignCode(text);
    if (caret !== null) setDesignerSelection(caret);
  }

  function pushDesignerHistory(caret = designerSelectionRange().end) {
    if (applyingDesignerHistory) return;
    const current = designerHistory[designerHistoryIndex];

    if (current?.text === schemaValue) {
      designerHistory[designerHistoryIndex] = { text: schemaValue, caret };
      return;
    }

    designerHistory = designerHistory.slice(0, designerHistoryIndex + 1);
    designerHistory.push({ text: schemaValue, caret });
    if (designerHistory.length > DESIGNER_HISTORY_LIMIT) designerHistory.shift();
    designerHistoryIndex = designerHistory.length - 1;
  }

  function restoreDesignerHistory(nextIndex) {
    if (nextIndex < 0 || nextIndex >= designerHistory.length) return false;
    applyingDesignerHistory = true;
    designerHistoryIndex = nextIndex;
    const entry = designerHistory[designerHistoryIndex];
    renderDesigner(entry.text, entry.caret);
    applyingDesignerHistory = false;
    return true;
  }

  function undoDesigner() {
    return restoreDesignerHistory(designerHistoryIndex - 1);
  }

  function redoDesigner() {
    return restoreDesignerHistory(designerHistoryIndex + 1);
  }

  function insertDesignerText(text) {
    const { start, end } = designerSelectionRange();
    const nextValue = schemaValue.slice(0, start) + text + schemaValue.slice(end);
    const nextCaret = start + text.length;
    renderDesigner(nextValue, nextCaret);
    pushDesignerHistory(nextCaret);
  }

  function onDesignerInput(e) {
    schemaValue = e.currentTarget.value;
    updateDesignCode(schemaValue);
    pushDesignerHistory();
    syncDesignerScroll();
  }

  function syncDesignerScroll() {
    if (!dsgnCodeEl || !dsgnHighlightEl) return;
    dsgnHighlightEl.scrollLeft = dsgnCodeEl.scrollLeft;
    dsgnHighlightEl.scrollTop = dsgnCodeEl.scrollTop;
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

  function selectedLineRange(text, start, end) {
    const activeEnd = end > start && text[end - 1] === '\n' ? end - 1 : end;
    const lineStart = text.lastIndexOf('\n', Math.max(0, start - 1)) + 1;
    const nextBreak = text.indexOf('\n', activeEnd);
    const lineEnd = nextBreak === -1 ? text.length : nextBreak;
    return { lineStart, lineEnd };
  }

  function toggleDesignerLineComments() {
    const { start: rawStart, end: rawEnd } = designerSelectionRange();
    const start = Math.max(0, Math.min(rawStart, schemaValue.length));
    const end = Math.max(start, Math.min(rawEnd, schemaValue.length));
    const { lineStart, lineEnd } = selectedLineRange(schemaValue, start, end);
    const lines = schemaValue.slice(lineStart, lineEnd).split('\n');
    const uncommentCandidates = lines.filter(line => line.trim().length > 0);
    const shouldUncomment = uncommentCandidates.length > 0
      && uncommentCandidates.every(line => /^\s*\/\//.test(line));

    const nextLines = lines.map(line => {
      if (shouldUncomment) {
        const match = line.match(/^(\s*)\/\/ ?/);
        return match ? match[1] + line.slice(match[0].length) : line;
      }

      const indent = line.match(/^\s*/)[0];
      return `${indent}// ${line.slice(indent.length)}`;
    });

    const toggledText = nextLines.join('\n');
    const nextValue = schemaValue.slice(0, lineStart) + toggledText + schemaValue.slice(lineEnd);
    const toggledEnd = lineStart + toggledText.length;
    const nextLineStart = nextValue[toggledEnd] === '\n' ? toggledEnd + 1 : toggledEnd;

    renderDesigner(nextValue, nextLineStart);
    pushDesignerHistory(nextLineStart);
  }

  function autoPairDesigner(e) {
    if (!isPlainTextKey(e) || !(e.key in AUTO_PAIRS)) return false;
    const { start: rawStart, end: rawEnd } = designerSelectionRange();
    const start = Math.max(0, Math.min(rawStart, schemaValue.length));
    const end = Math.max(start, Math.min(rawEnd, schemaValue.length));
    if (isInsideString(schemaValue, start)) return false;

    const close = AUTO_PAIRS[e.key];
    const selectedText = schemaValue.slice(start, end);
    const nextValue = schemaValue.slice(0, start) + e.key + selectedText + close + schemaValue.slice(end);
    const nextCaret = start === end ? start + 1 : end + 2;

    e.preventDefault();
    e.stopPropagation();
    renderDesigner(nextValue, nextCaret);
    pushDesignerHistory(nextCaret);
    return true;
  }

  function skipExistingDesignerClose(e) {
    if (!isPlainTextKey(e) || !Object.values(AUTO_PAIRS).includes(e.key)) return false;
    const { start, end } = designerSelectionRange();
    if (start !== end) return false;

    const caret = Math.max(0, Math.min(start, schemaValue.length));
    if (schemaValue[caret] !== e.key) return false;
    if (e.key === '"' ? !isInsideString(schemaValue, caret) : isInsideString(schemaValue, caret)) return false;

    e.preventDefault();
    e.stopPropagation();
    setDesignerSelection(caret + 1);
    pushDesignerHistory(caret + 1);
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

  function insertDesignerIndentedNewline() {
    const { start, end } = designerSelectionRange();
    const indent = indentForNewLine(schemaValue, start);
    const insert = `\n${indent}`;
    const nextValue = schemaValue.slice(0, start) + insert + schemaValue.slice(end);
    const nextCaret = start + insert.length;

    renderDesigner(nextValue, nextCaret);
    pushDesignerHistory(nextCaret);
  }

  function deleteDesignerIndentOnlyLine(e) {
    if (e.key !== 'Backspace' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const { start, end } = designerSelectionRange();
    if (start !== end || start === 0) return false;

    const lineStart = schemaValue.lastIndexOf('\n', Math.max(0, start - 1)) + 1;
    const linePrefix = schemaValue.slice(lineStart, start);
    if (lineStart === 0 || linePrefix.length === 0 || /\S/.test(linePrefix)) return false;

    const previousLineEnd = lineStart - 1;
    const nextValue = schemaValue.slice(0, previousLineEnd) + schemaValue.slice(start);

    e.preventDefault();
    e.stopPropagation();
    renderDesigner(nextValue, previousLineEnd);
    pushDesignerHistory(previousLineEnd);
    return true;
  }

  function deleteDesignerEmptyPair(e) {
    if (e.key !== 'Backspace' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const { start, end } = designerSelectionRange();
    if (start !== end || start === 0) return false;

    const open = schemaValue[start - 1];
    const close = schemaValue[start];
    if (AUTO_PAIRS[open] !== close) return false;

    const nextValue = schemaValue.slice(0, start - 1) + schemaValue.slice(start + 1);
    const nextCaret = start - 1;

    e.preventDefault();
    e.stopPropagation();
    renderDesigner(nextValue, nextCaret);
    pushDesignerHistory(nextCaret);
    return true;
  }

  function expandDesignerBraceOnEnter(e) {
    if (e.key !== 'Enter' || e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return false;
    const { start, end } = designerSelectionRange();
    if (start !== end) return false;

    const caret = Math.max(0, Math.min(start, schemaValue.length));
    if (schemaValue[caret - 1] !== '{' || schemaValue[caret] !== '}') return false;
    if (isInsideString(schemaValue, caret)) return false;

    const baseIndent = currentLineIndent(schemaValue, caret);
    const innerIndent = `${baseIndent}  `;
    const insert = `\n${innerIndent}\n${baseIndent}`;
    const nextValue = schemaValue.slice(0, caret) + insert + schemaValue.slice(caret);
    const nextCaret = caret + 1 + innerIndent.length;

    e.preventDefault();
    e.stopPropagation();
    renderDesigner(nextValue, nextCaret);
    pushDesignerHistory(nextCaret);
    return true;
  }

  function handleDesignerKey(e) {
    if (isUndoShortcut(e)) {
      e.preventDefault();
      e.stopPropagation();
      undoDesigner();
      return;
    }
    if (isRedoShortcut(e)) {
      e.preventDefault();
      e.stopPropagation();
      redoDesigner();
      return;
    }
    if (isCommentShortcut(e)) {
      e.preventDefault();
      e.stopPropagation();
      e.stopImmediatePropagation();
      toggleDesignerLineComments();
      return;
    }
    if (skipExistingDesignerClose(e) || autoPairDesigner(e)) return;
    if (deleteDesignerEmptyPair(e)) return;
    if (deleteDesignerIndentOnlyLine(e)) return;
    if (e.key === 'Tab') {
      e.preventDefault();
      insertDesignerText('  ');
      return;
    }
    if (e.key === 'Enter') {
      if (expandDesignerBraceOnEnter(e)) return;
      e.preventDefault();
      insertDesignerIndentedNewline();
    }
  }

  export function toggleDebugPanel() {
    if (debugDidDrag) { debugDidDrag = false; return; }
    debugCollapsed.update(v => !v);
  }

  function debugLevelLabel(level) {
    if (level === 'high') return 'High';
    if (level === 'warn') return 'Warning';
    return level ? level[0].toUpperCase() + level.slice(1) : 'Info';
  }

  function handleDebugKey(e) {
    if (e.key !== 'Enter' && e.key !== ' ') return;
    e.preventDefault();
    toggleDebugPanel();
  }

  function initDebugResize(e) {
    if ($debugCollapsed) return;
    debugResizing = true;
    debugDidDrag = false;
    debugResizeStartY = e.clientY;
    debugResizeStartH = $debugOpenHeight;
    document.addEventListener('mousemove', onDebugResizeMove);
    document.addEventListener('mouseup', onDebugResizeUp);
    e.preventDefault();
  }

  function onDebugResizeMove(e) {
    if (!debugResizing) return;
    debugDidDrag = true;
    const delta = debugResizeStartY - e.clientY;
    debugOpenHeight.set(Math.max(60, Math.min(500, debugResizeStartH + delta)));
  }

  function onDebugResizeUp() {
    debugResizing = false;
    document.removeEventListener('mousemove', onDebugResizeMove);
    document.removeEventListener('mouseup', onDebugResizeUp);
  }

  function initResizeHandle() {
    const main = document.getElementById('main');
    let dragging = false, startX = 0, startWidth = 0;

    resizeHandleEl.addEventListener('mousedown', e => {
      dragging = true;
      startX = e.clientX;
      startWidth = panelEl.offsetWidth;
      resizeHandleEl.classList.add('dragging');
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
      e.preventDefault();
    });

    document.addEventListener('mousemove', e => {
      if (!dragging) return;
      const delta = startX - e.clientX;
      const maxW = main.offsetWidth - 300;
      const newW = Math.max(200, Math.min(startWidth + delta, maxW));
      panelEl.style.width = newW + 'px';
    });

    document.addEventListener('mouseup', () => {
      if (!dragging) return;
      dragging = false;
      resizeHandleEl.classList.remove('dragging');
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    });
  }

  onMount(() => {
    updateDesignCode(schemaValue);
    pushDesignerHistory(0);
    initResizeHandle();
  });

  $: debugPanelHeight = $debugCollapsed ? 0 : $debugOpenHeight;
  $: toggleIconPath = $debugCollapsed ? 'M2 4l3 3 3-3' : 'M2 6l3-3 3 3';
  $: debugLogCount = $debugLogs.length;
  $: debugLogStatus = debugLogCount === 0
    ? 'No logs'
    : `${debugLogCount} log${debugLogCount === 1 ? '' : 's'}`;
  $: {
    const newestId = $debugLogs[$debugLogs.length - 1]?.id ?? 0;
    if (dbgScrollEl && newestId !== lastDebugLogId) {
      lastDebugLogId = newestId;
      tick().then(() => {
        if (dbgScrollEl) dbgScrollEl.scrollTop = dbgScrollEl.scrollHeight;
      });
    }
  }
</script>

<div id="resize-handle" bind:this={resizeHandleEl}></div>

<div id="designer-panel" bind:this={panelEl}>
  <div class="dsgn-header">
    <span class="dsgn-filename">Designer</span>
  </div>

  <div class="dsgn-editor">
    <div class="dsgn-scroll">
      <div class="dsgn-row">
        <div class="dsgn-line-nums" aria-hidden="true">
          {#each designerLineNumbers as n}
            <div>{n}</div>
          {/each}
        </div>
        <div class="dsgn-stack">
          <pre class="dsgn-highlight" aria-hidden="true" bind:this={dsgnHighlightEl}>{@html highlightedSchema}</pre>
          <textarea
            class="dsgn-code"
            aria-multiline="true"
            spellcheck="false"
            wrap="off"
            rows={designerLineCount}
            bind:this={dsgnCodeEl}
            value={schemaValue}
            on:input={onDesignerInput}
            on:scroll={syncDesignerScroll}
            on:keydown={handleDesignerKey}
          ></textarea>
        </div>
      </div>
    </div>
  </div>

  <div
    id="debug-handle"
    on:mousedown={initDebugResize}
    on:click={toggleDebugPanel}
    role="button"
    tabindex="0"
    on:keydown={handleDebugKey}
  >
    <span class="dbg-handle-label">
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4"><circle cx="6" cy="4" r="2"/><path d="M3 11c0-1.657 1.343-3 3-3s3 1.343 3 3"/><path d="M1 4h1M10 4h1M6 1v1"/></svg>
      <span class="dbg-handle-title">Debug</span>
      <span class="dbg-count">{debugLogStatus}</span>
    </span>
    <div class="dbg-actions">
      <button
        class="dbg-clear-btn"
        on:mousedown|stopPropagation
        on:click|stopPropagation={clearDebugLogs}
        title="Clear debug logs"
        aria-label="Clear debug logs"
        disabled={debugLogCount === 0}
      >
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <path d="M2 3h8M4.5 3V2h3v1M9 3l-.5 7h-5L3 3"/>
        </svg>
      </button>
      <button
        class="dbg-toggle-btn"
        on:mousedown|stopPropagation
        on:click|stopPropagation={toggleDebugPanel}
        title="Toggle debugger"
      >
        <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d={toggleIconPath}/>
        </svg>
      </button>
    </div>
  </div>

  <div id="debug-panel" style="height: {debugPanelHeight}px">
    <div
      class="dbg-scroll"
      bind:this={dbgScrollEl}
      role="log"
      aria-live="polite"
      aria-label="Notifier debug logs"
    >
      {#if debugLogCount === 0}
        <div class="dbg-empty">No Notifier logs yet</div>
      {:else}
        {#each $debugLogs as log (log.id)}
          <div class="dbg-line dbg-line--{log.level}" title={log.source}>
            <span class="dbg-ts">{log.time}</span>
            <span class="dbg-level">{debugLevelLabel(log.level)}</span>
            <span class="dbg-msg">{log.message}</span>
          </div>
        {/each}
      {/if}
    </div>
  </div>
</div>
