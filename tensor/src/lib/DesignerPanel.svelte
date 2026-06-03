<script>
  import { onMount, tick } from 'svelte';
  import {
    activeScreen,
    designCode,
    initialDesignCode,
    updateDesignCode,
  } from './stores.js';
  import { schemaTokenize, tokensToHtml } from './tokenizer.js';
  import DesignerVisual from './DesignerVisual.svelte';

  let visualMode = false;

  let dsgnCodeEl;
  let dsgnHighlightEl;
  let panelEl;
  let resizeHandleEl;
  let schemaValue = initialDesignCode;
  let designerHistory = [];
  let designerHistoryIndex = -1;
  let applyingDesignerHistory = false;

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

  function resetDesignerHistory(caret = 0) {
    designerHistory = [];
    designerHistoryIndex = -1;
    pushDesignerHistory(Math.max(0, Math.min(caret, schemaValue.length)));
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

  function onVisualDesignerChange(e) {
    schemaValue = e.detail.schema;
    updateDesignCode(schemaValue);
    pushDesignerHistory(schemaValue.length);
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

  function handleVisualDesignerKey(e) {
    if (!visualMode || !panelEl?.contains(document.activeElement)) return;
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
    }
  }

  function initResizeHandle() {
    const main = document.getElementById('main');
    let dragging = false, startX = 0, startWidth = 0;

    const resetDrag = () => {
      dragging = false;
      resizeHandleEl?.classList.remove('dragging');
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };

    const onMouseDown = e => {
      dragging = true;
      startX = e.clientX;
      startWidth = panelEl.offsetWidth;
      resizeHandleEl.classList.add('dragging');
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
      e.preventDefault();
    };

    const onMouseMove = e => {
      if (!dragging) return;
      const delta = startX - e.clientX;
      const maxW = (main?.offsetWidth || window.innerWidth) - 300;
      const newW = Math.max(200, Math.min(startWidth + delta, maxW));
      panelEl.style.flex = `0 0 ${newW}px`;
    };

    const onMouseUp = () => {
      if (!dragging) return;
      resetDrag();
    };

    resizeHandleEl.addEventListener('mousedown', onMouseDown);
    document.addEventListener('mousemove', onMouseMove);
    document.addEventListener('mouseup', onMouseUp);

    return () => {
      resizeHandleEl?.removeEventListener('mousedown', onMouseDown);
      document.removeEventListener('mousemove', onMouseMove);
      document.removeEventListener('mouseup', onMouseUp);
      if (dragging) resetDrag();
    };
  }

  onMount(() => {
    let mountedScreen = null;
    const unsubscribeDesign = designCode.subscribe(value => {
      const nextValue = value || '';
      const firstSync = designerHistoryIndex === -1;
      if (!firstSync && nextValue === schemaValue) return;
      schemaValue = nextValue;
      resetDesignerHistory(0);
      tick().then(syncDesignerScroll);
    });
    const unsubscribeScreen = activeScreen.subscribe(name => {
      if (mountedScreen === null) {
        mountedScreen = name;
        return;
      }
      if (name === mountedScreen) return;
      mountedScreen = name;
      resetDesignerHistory(0);
      tick().then(syncDesignerScroll);
    });
    const cleanupResize = initResizeHandle();

    return () => {
      unsubscribeDesign();
      unsubscribeScreen();
      cleanupResize?.();
    };
  });

</script>

<svelte:window on:keydown={handleVisualDesignerKey} />

<div id="resize-handle" bind:this={resizeHandleEl}></div>

<div id="designer-panel" bind:this={panelEl}>
  <div class="dsgn-header">
    <span class="dsgn-filename">Designer</span>
    <div class="dsgn-mode-toggle">
      <button
        class="dsgn-mode-btn"
        class:dsgn-mode-btn--active={!visualMode}
        on:click={() => (visualMode = false)}
        title="Text editor"
      >
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
          <path d="M1 3h10M1 6h7M1 9h9"/>
        </svg>
        Text
      </button>
      <button
        class="dsgn-mode-btn"
        class:dsgn-mode-btn--active={visualMode}
        on:click={() => (visualMode = true)}
        title="Tree editor"
      >
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
          <rect x="4" y="0.5" width="4" height="2.5" rx="0.5"/>
          <line x1="6" y1="3" x2="6" y2="5.5"/>
          <line x1="2.5" y1="5.5" x2="9.5" y2="5.5"/>
          <line x1="2.5" y1="5.5" x2="2.5" y2="7"/>
          <line x1="9.5" y1="5.5" x2="9.5" y2="7"/>
          <rect x="1" y="7" width="3" height="2.5" rx="0.5"/>
          <rect x="8" y="7" width="3" height="2.5" rx="0.5"/>
        </svg>
        Tree
      </button>
    </div>
  </div>

  {#if visualMode}
    <div class="dsgn-visual-wrap">
      <DesignerVisual
        schemaValue={schemaValue}
        on:change={onVisualDesignerChange}
        on:switchText={() => (visualMode = false)}
      />
    </div>
  {:else}
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
  {/if}

</div>
