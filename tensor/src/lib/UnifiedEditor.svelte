<script>
  import { onMount, tick } from 'svelte';
  import { cells, updateCellCode } from './stores.js';
  import { falconTokenize, tokensToHtml } from './tokenizer.js';

  const SEP_RE_G = /\/\/ ─── \[cell:([^\]]+)\] ───\n?/g;
  const SEP_LINE = id => `// ─── [cell:${id}] ───`;

  const AUTO_PAIRS = { '{': '}', '(': ')', '[': ']', '"': '"' };
  const HISTORY_LIMIT = 100;

  let codeEl, highlightEl, wrapEl;
  let editorValue = '';
  let history = [];
  let historyIndex = -1;
  let applyingHistory = false;

  $: lineCount = Math.max(editorValue.split('\n').length, 1);
  $: lineNumbers = Array.from({ length: lineCount }, (_, i) => i + 1);
  $: highlightedCode = tokensToHtml(falconTokenize(editorValue));

  function buildContent(cellList) {
    const code = cellList.filter(c => c.type === 'code');
    if (!code.length) return '';
    return code.map(c => `${SEP_LINE(c.id)}\n${c.code || ''}`).join('\n\n');
  }

  function parseAndSync(content) {
    const re = new RegExp(SEP_RE_G.source, 'g');
    let lastIndex = 0, lastId = null, match;
    while ((match = re.exec(content)) !== null) {
      if (lastId !== null) {
        updateCellCode(lastId, content.slice(lastIndex, match.index).replace(/\n+$/, ''));
      }
      lastId = match[1];
      lastIndex = match.index + match[0].length;
    }
    if (lastId !== null) {
      updateCellCode(lastId, content.slice(lastIndex).replace(/\n+$/, ''));
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
    parseAndSync(text);
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
  }

  function onInput(e) {
    editorValue = e.currentTarget.value;
    parseAndSync(editorValue);
    pushHistory();
    syncScroll();
  }

  onMount(() => {
    editorValue = buildContent($cells);
    pushHistory(0);
  });
</script>

<div class="unified-wrap" bind:this={wrapEl}>
  <div class="unified-editor">
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
      ></textarea>
    </div>
  </div>
</div>

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
</style>
