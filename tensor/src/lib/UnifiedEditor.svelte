<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import {
    activeScreen, cells, replaceCodeCells,
    liveTestState, companionCommand,
    doItResults, clearDoItResult, unifiedSelectionActive,
    debugModeEnabled, debugActiveLocation, debugRuntimeErrors,
  } from './stores.js';
  import { mistToXmlResult, runCodeDiagnosticResult } from './falcon-wasm.js';
  import { blocklyXmlToPng, componentDefinitionsFromDesigner, ensureBlocklyRuntime, copyPngBlobToClipboard, downloadPngBlob } from './blockly-preview.js';
  import { splitFalconSourceByTopLevelLines } from './cell-splitting.js';
  import { falconLang, falconSyntaxHighlighting } from './falcon-cm.js';

  // CodeMirror imports
  import { EditorView, keymap, lineNumbers, Decoration, WidgetType, gutter, GutterMarker } from '@codemirror/view';
  import { EditorState, StateField, StateEffect, RangeSetBuilder } from '@codemirror/state';
  import { defaultKeymap, historyKeymap, history, toggleComment, indentWithTab } from '@codemirror/commands';
  import { closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete';
  import { indentOnInput } from '@codemirror/language';

  // ── State ─────────────────────────────────────────────────────────────────
  let cmEl;
  let wrapEl;
  let view;
  let syncingToStore = false;
  let unsubscribeCells = null;
  let unsubscribeScreen = null;
  let lastCellsSignature = '';

  // Callout state
  let callout = null;
  let calloutRunId = 0;
  let calloutDebounceTimer = null;
  let doItCallout = null;
  let doItCalloutEl = null;
  let pendingDoItPos = null;

  const CALLOUT_DEBOUNCE_MS = 280;
  const LINE_HEIGHT = 13 * 1.65;
  const CODE_PAD_TOP = 12;
  const CALLOUT_GAP = 8;
  const CALLOUT_PADDING = 12;

  $: isCompanionConnected = $liveTestState.status === 'connected';

  $: debugActiveUnifiedLine = $debugModeEnabled ? ($debugActiveLocation?.unifiedLine ?? null) : null;
  $: debugErrorEntries = $debugModeEnabled
    ? Object.values($debugRuntimeErrors || {})
        .filter(e => e?.unifiedLine)
        .sort((a, b) => a.unifiedLine - b.unifiedLine)
    : [];
  $: debugErrorUnifiedLines = new Set(debugErrorEntries.map(e => e.unifiedLine));
  $: debugFirstError = debugErrorEntries[0] ?? null;
  $: debugFirstErrorLine = debugFirstError?.unifiedLine ?? null;
  $: debugFirstErrorLines = debugFirstError?.message
    ? String(debugFirstError.message).split(/\r?\n/) : [];

  // Push debug state into CM
  $: if (view) {
    view.dispatch({ effects: [
      setDebugActiveLine.of(debugActiveUnifiedLine),
      setDebugErrorEntries.of(debugErrorEntries),
    ]});
  }

  // Handle do-it result arriving via store
  $: {
    const result = $doItResults['unified'];
    if (result && pendingDoItPos) {
      const pos = doItCalloutPos(pendingDoItPos.startLine, pendingDoItPos.selStart);
      doItCallout = { ...pos, startLine: pendingDoItPos.startLine, endLine: pendingDoItPos.endLine, selStart: pendingDoItPos.selStart, status: 'ready', lines: result.lines, ok: result.ok };
      pendingDoItPos = null;
      clearDoItResult('unified');
    }
  }

  // ── CodeMirror effects & state ────────────────────────────────────────────
  const externalChange      = StateEffect.define();
  const setDebugActiveLine  = StateEffect.define();
  const setDebugErrorEntries = StateEffect.define();

  class ErrorWidget extends WidgetType {
    constructor(lines) { super(); this.lines = lines; }
    eq(other) { return JSON.stringify(other.lines) === JSON.stringify(this.lines); }
    toDOM() {
      const wrap = document.createElement('div');
      wrap.className = 'cm-debug-error-widget';
      const icon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      icon.setAttribute('viewBox', '0 0 14 14'); icon.setAttribute('fill', 'none');
      icon.setAttribute('stroke', 'currentColor'); icon.setAttribute('stroke-width', '1.4');
      icon.setAttribute('stroke-linecap', 'round'); icon.setAttribute('stroke-linejoin', 'round');
      icon.setAttribute('width', '13'); icon.setAttribute('height', '13');
      icon.setAttribute('aria-hidden', 'true');
      icon.innerHTML = '<ellipse cx="7" cy="7.5" rx="3" ry="3.5"/><path d="M7 4V2.5"/><path d="M4 7.5H2M12 7.5h-2"/><path d="M4.5 5.2l-1.2-1.2M9.5 5.2l1.2-1.2"/><path d="M4.5 9.8l-1.2 1.2M9.5 9.8l1.2 1.2"/>';
      const body = document.createElement('div');
      body.className = 'cm-debug-error-widget-body';
      for (const msg of (this.lines.length ? this.lines : ['Runtime error'])) {
        const d = document.createElement('div'); d.textContent = msg; body.appendChild(d);
      }
      wrap.appendChild(icon); wrap.appendChild(body);
      return wrap;
    }
    ignoreEvent() { return true; }
  }

  class DebugCurrentMarker extends GutterMarker {
    toDOM() { return document.createTextNode(''); }
    elementClass = 'debug-gutter-current';
  }
  class DebugErrorMarker extends GutterMarker {
    toDOM() { return document.createTextNode(''); }
    elementClass = 'debug-gutter-error';
  }
  const debugCurrentMarker = new DebugCurrentMarker();
  const debugErrorMarker = new DebugErrorMarker();

  const debugState = StateField.define({
    create() { return { activeLine: null, errorEntries: [] }; },
    update(val, tr) {
      let next = val;
      for (const e of tr.effects) {
        if (e.is(setDebugActiveLine))  next = { ...next, activeLine: e.value };
        if (e.is(setDebugErrorEntries)) next = { ...next, errorEntries: e.value };
      }
      return next;
    },
  });

  const debugDecorations = EditorView.decorations.compute([debugState], state => {
    const { activeLine, errorEntries } = state.field(debugState);
    const builder = new RangeSetBuilder();
    const entries = [];

    if (activeLine != null && activeLine >= 1 && activeLine <= state.doc.lines) {
      const line = state.doc.line(activeLine);
      entries.push({ from: line.from, to: line.from, dec: Decoration.line({ class: 'cm-debug-active-line' }) });
    }

    for (const err of (errorEntries || [])) {
      const ln = err.unifiedLine;
      if (ln == null || ln < 1 || ln > state.doc.lines) continue;
      const line = state.doc.line(ln);
      entries.push({ from: line.from, to: line.from, dec: Decoration.line({ class: 'cm-debug-error-line-bg' }) });
      const msgs = err.message ? String(err.message).split(/\r?\n/) : [];
      entries.push({
        from: line.to, to: line.to,
        dec: Decoration.widget({ widget: new ErrorWidget(msgs), block: true, side: 1 }),
      });
    }

    entries.sort((a, b) => a.from - b.from || a.to - b.to);
    for (const e of entries) builder.add(e.from, e.to, e.dec);
    return builder.finish();
  });

  const debugGutter = gutter({
    class: 'cm-debug-gutter',
    lineMarker(view, line) {
      const { activeLine, errorEntries } = view.state.field(debugState);
      const lineNo = view.state.doc.lineAt(line.from).number;
      if (lineNo === activeLine) return debugCurrentMarker;
      if ((errorEntries || []).some(e => e.unifiedLine === lineNo)) return debugErrorMarker;
      return null;
    },
    lineMarkerChange: update => update.transactions.some(tr =>
      tr.effects.some(e => e.is(setDebugActiveLine) || e.is(setDebugErrorEntries))),
  });

  // Track selection for callout + unifiedSelectionActive
  const selectionListener = EditorView.updateListener.of(update => {
    if (!update.selectionSet && !update.docChanged) return;
    const sel = update.state.selection.main;
    const hasSelection = sel.from !== sel.to;
    unifiedSelectionActive.set(hasSelection);
    if (hasSelection) scheduleCalloutCheck();
    else { unifiedSelectionActive.set(false); dismissCallout(); }
  });

  // ── Cell ↔ editor sync helpers ────────────────────────────────────────────
  function buildContent(cellList) {
    return cellList.filter(c => c.type === 'code').map(c => c.code || '').join('\n\n');
  }

  function codeCellsSignature(cellList) {
    return JSON.stringify((cellList || []).filter(c => c.type === 'code').map(c => [c.id, c.code || '']));
  }

  function syncFromCells(cellList, force = false) {
    const sig = codeCellsSignature(cellList);
    if (!force && sig === lastCellsSignature) return;
    lastCellsSignature = sig;
    const next = buildContent(cellList || []);
    if (!view) return;
    const current = view.state.doc.toString();
    if (next === current) return;
    view.dispatch({
      changes: { from: 0, to: current.length, insert: next },
      annotations: [externalChange.of(true)],
    });
    dismissCallout();
  }

  function syncToStore() {
    if (!view) return;
    const text = view.state.doc.toString();
    syncingToStore = true;
    try { replaceCodeCells(text.trim() ? [text] : []); }
    finally { syncingToStore = false; }
  }

  export async function commitToCells() {
    if (!view) return false;
    const source = view.state.doc.toString();
    if (!source.trim()) {
      syncingToStore = true;
      try { replaceCodeCells([]); } finally { syncingToStore = false; }
      return true;
    }
    try {
      const componentDefinitions = await componentDefinitionsFromDesigner();
      const result = await mistToXmlResult(source, componentDefinitions);
      const chunks = splitFalconSourceByTopLevelLines(source, result.lineNumbers);
      syncingToStore = true;
      try { replaceCodeCells(chunks); } finally { syncingToStore = false; }
      return true;
    } catch (e) {
      console.warn('[unified-editor] unable to split source by top-level expressions:', e);
      syncingToStore = true;
      try { replaceCodeCells([source]); } finally { syncingToStore = false; }
      return false;
    }
  }

  // ── Callout helpers ───────────────────────────────────────────────────────
  function editorValue() { return view?.state.doc.toString() || ''; }

  function lineOfOffset(offset) {
    if (!view) return 1;
    return view.state.doc.lineAt(Math.max(0, Math.min(offset, view.state.doc.length))).number;
  }

  function selectionLineRange(selStart, selEnd) {
    if (!view) return { startLine: 1, endLine: 1 };
    const doc = view.state.doc;
    const s = Math.max(0, Math.min(selStart, doc.length));
    const e = Math.max(s, Math.min(selEnd, doc.length));
    const activeEnd = e > s && doc.sliceString(e - 1, e) === '\n' ? e - 1 : e;
    const startLine = doc.lineAt(s).number;
    const endLine = Math.max(startLine, doc.lineAt(activeEnd).number);
    return { startLine, endLine };
  }

  function firstCodeLineInSelection(startLine, endLine) {
    const lines = editorValue().split('\n');
    for (let line = startLine; line <= endLine; line++) {
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
    if (!view || !wrapEl) return 0;
    const rect = view.scrollDOM.getBoundingClientRect();
    const centerLine = (startLine + endLine) / 2;
    return rect.top + CODE_PAD_TOP + (centerLine - 1) * LINE_HEIGHT - view.scrollDOM.scrollTop;
  }

  function calloutX(startLine, endLine) {
    if (!view) return { x: 0, overflowed: false };
    const lines = editorValue().split('\n');
    const selLines = lines.slice(startLine - 1, endLine);
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.font = getComputedStyle(view.contentDOM).font;
    const maxPx = Math.max(0, ...selLines.map(l => ctx.measureText(l).width));
    const rect = view.scrollDOM.getBoundingClientRect();
    const rawX = rect.left + 4 + maxPx;
    const frameRight = wrapEl ? wrapEl.getBoundingClientRect().right : window.innerWidth;
    const overflowed = rawX + CALLOUT_GAP > frameRight;
    return { x: overflowed ? frameRight : rawX, overflowed };
  }

  function doItCalloutPos(startLine, selStart) {
    if (!view) return { x: 0, y: 0, maxWidth: 280 };
    const rect = view.scrollDOM.getBoundingClientRect();
    const lines = editorValue().split('\n');
    const lineCharStart = lines.slice(0, startLine - 1).reduce((acc, l) => acc + l.length + 1, 0);
    const textToSel = editorValue().slice(lineCharStart, Math.max(lineCharStart, Math.min(selStart ?? lineCharStart, lineCharStart + (lines[startLine - 1] || '').length)));
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    ctx.font = getComputedStyle(view.contentDOM).font;
    const colPx = ctx.measureText(textToSel).width;
    const x = rect.left + 4 + colPx - (view.scrollDOM.scrollLeft || 0);
    const y = rect.top + CODE_PAD_TOP + (startLine - 1) * LINE_HEIGHT - view.scrollDOM.scrollTop;
    const frameRight = wrapEl ? wrapEl.getBoundingClientRect().right : window.innerWidth;
    const maxWidth = Math.max(120, Math.min(320, frameRight - x - 16));
    return { x, y, maxWidth };
  }

  function updateCalloutY() {
    if (callout) {
      const { x, overflowed } = calloutX(callout.startLine, callout.endLine);
      callout = { ...callout, y: calloutY(callout.startLine, callout.endLine), x, overflowed };
    }
    if (doItCallout) {
      const pos = doItCalloutPos(doItCallout.startLine, doItCallout.selStart);
      doItCallout = { ...doItCallout, ...pos };
    }
  }

  function dismissDoItCallout() { doItCallout = null; }

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

  async function runCalloutCheck(runId) {
    if (!view || runId !== calloutRunId) return;
    const { from: selStart, to: selEnd } = view.state.selection.main;
    if (selStart === selEnd) { unifiedSelectionActive.set(false); dismissCallout(); return; }
    unifiedSelectionActive.set(true);

    const { startLine: selectionStartLine, endLine } = selectionLineRange(selStart, selEnd);

    let xmlResult, componentDefinitions;
    try {
      componentDefinitions = await componentDefinitionsFromDesigner();
      if (runId !== calloutRunId) return;
      await ensureBlocklyRuntime();
      if (runId !== calloutRunId) return;
      xmlResult = await mistToXmlResult(editorValue(), componentDefinitions);
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

  // ── Do it (run selection) ─────────────────────────────────────────────────
  export async function runDoIt() {
    if (!view) return;
    const { from, to } = view.state.selection.main;
    if (from === to) return;
    const source = view.state.sliceDoc(from, to);
    if (!source.trim()) return;

    const { startLine, endLine } = selectionLineRange(from, to);
    const pos = doItCalloutPos(startLine, from);
    pendingDoItPos = { ...pos, startLine, endLine, selStart: from };
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
        const diags = result.diagnostics?.length ? result.diagnostics : [{ message: result.error || 'Execution failed' }];
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

  // ── Build EditorView ──────────────────────────────────────────────────────
  function buildView(container, doc) {
    return new EditorView({
      parent: container,
      state: EditorState.create({
        doc,
        extensions: [
          history(),
          falconLang,
          falconSyntaxHighlighting,
          closeBrackets(),
          indentOnInput(),
          lineNumbers(),
          debugState,
          debugDecorations,
          debugGutter,
          selectionListener,
          keymap.of([
            { key: 'Mod-/', run: toggleComment },
            { key: 'Tab', run: indentWithTab.run, shift: indentWithTab.shift },
            ...closeBracketsKeymap,
            ...defaultKeymap,
            ...historyKeymap,
          ]),
          EditorView.updateListener.of(update => {
            if (!update.docChanged) return;
            if (update.transactions.some(tr => tr.annotation(externalChange))) return;
            syncToStore();
            dismissCallout();
          }),
          EditorView.domEventHandlers({
            scroll() { updateCalloutY(); },
            blur()   { unifiedSelectionActive.set(false); },
          }),
          EditorView.theme({
            '&': { background: 'transparent', height: '100%' },
            '.cm-scroller': {
              fontFamily: 'var(--mono)',
              fontSize: '13px',
              lineHeight: '1.65',
              overflow: 'auto',
            },
            '.cm-content': { padding: '12px 16px 12px 4px', caretColor: 'var(--text)' },
            '.cm-line': { padding: '0' },
            '&.cm-focused .cm-cursor': { borderLeftColor: 'var(--text)' },
            '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': {
              background: 'color-mix(in srgb, var(--accent) 24%, transparent)',
            },
            '&.cm-focused': { outline: 'none' },
            '.cm-gutters': {
              background: 'var(--surface)',
              border: 'none',
              color: 'var(--text-faint)',
              fontFamily: 'var(--mono)',
              fontSize: '13px',
              lineHeight: '1.65',
              userSelect: 'none',
              minWidth: '40px',
            },
            '.cm-lineNumbers .cm-gutterElement': { padding: '0 10px 0 8px', minWidth: '32px', textAlign: 'right' },
            '.cm-debug-gutter': { width: '3px', background: 'transparent' },
            '.cm-debug-gutter .cm-gutterElement': { padding: '0' },
          }),
        ],
      }),
    });
  }

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  onMount(() => {
    let mountedScreen = null;
    view = buildView(cmEl, '');

    unsubscribeCells = cells.subscribe(cellList => {
      const sig = codeCellsSignature(cellList);
      if (syncingToStore) { lastCellsSignature = sig; return; }
      syncFromCells(cellList);
    });

    unsubscribeScreen = activeScreen.subscribe(name => {
      if (mountedScreen === null) { mountedScreen = name; return; }
      if (name === mountedScreen) return;
      mountedScreen = name;
      syncFromCells($cells, true);
    });
  });

  onDestroy(() => {
    unsubscribeCells?.();
    unsubscribeScreen?.();
    unifiedSelectionActive.set(false);
    dismissCallout();
    dismissDoItCallout();
    view?.destroy();
  });
</script>

<div
  class="unified-wrap"
  class:debug-active-editor={Boolean(debugActiveUnifiedLine)}
  class:debug-error-editor={debugFirstErrorLine !== null}
  data-debug-active-line={debugActiveUnifiedLine || undefined}
  bind:this={wrapEl}
>
  <div class="unified-editor">
    <div class="cm-host" bind:this={cmEl}></div>
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
    flex: 1;
    display: flex;
    background: var(--surface);
    overflow: hidden;
  }

  .cm-host {
    flex: 1;
    min-width: 0;
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

  .blocks-callout--loading { padding: 10px 12px; }

  .callout-spinner {
    display: block;
    width: 20px;
    height: 20px;
    animation: callout-spin 0.9s linear infinite;
    color: var(--text-faint);
  }

  @keyframes callout-spin { to { transform: rotate(360deg); } }

  .callout-img {
    display: block;
    width: auto;
    max-width: 100%;
    height: auto;
    border-radius: 6px;
  }

  .callout-actions { display: flex; gap: 4px; margin-top: 6px; }

  .callout-action-btn {
    display: flex; align-items: center; justify-content: center;
    gap: 4px; height: 26px; padding: 0 8px;
    border: 1px solid var(--border); background: transparent;
    color: var(--text-muted); font-family: var(--font); font-size: 11px;
    border-radius: 6px; cursor: pointer;
    transition: background 0.1s, color 0.1s, border-color 0.1s;
    white-space: nowrap;
  }
  .callout-action-btn:hover { background: var(--cell-active); color: var(--text); border-color: var(--text-faint); }
  .callout-action-btn svg { width: 12px; height: 12px; flex-shrink: 0; }

  /* ── Do it callout ── */
  .do-it-callout {
    position: fixed; z-index: 80;
    transform: translateY(calc(-100% - 6px));
    border-radius: 10px; padding: 7px 8px 7px 10px;
    pointer-events: auto; display: flex; align-items: flex-start;
    gap: 6px; font-family: var(--mono); font-size: 12px;
    line-height: 1.5; min-width: 60px;
  }
  .do-it-callout::after {
    content: ''; position: absolute; bottom: -5px; left: 10px;
    width: 8px; height: 8px;
    border-bottom: 1.5px solid var(--accent); border-right: 1.5px solid var(--accent);
    transform: rotate(45deg); border-radius: 0 0 2px 0;
  }
  .do-it-callout--ok { background: var(--run-soft); border: 1.5px solid var(--accent); color: var(--text); }
  .do-it-callout--ok::after { background: var(--run-soft); border-color: var(--accent); }
  .do-it-callout--error { background: var(--error-soft); border: 1.5px solid var(--error); color: var(--error); }
  .do-it-callout--error::after { background: var(--error-soft); border-color: var(--error); }
  .do-it-callout--running { background: var(--surface); border: 1.5px solid var(--border); padding: 10px 12px; }
  .do-it-callout--running::after { background: var(--surface); border-color: var(--border); }

  .do-it-callout-arrow { flex-shrink: 0; font-weight: 600; color: var(--accent); user-select: none; padding-top: 1px; }
  .do-it-callout--error .do-it-callout-arrow { color: var(--error); }
  .do-it-callout-body { flex: 1; min-width: 0; }
  .do-it-callout-line { white-space: pre-wrap; word-break: break-word; }
  .do-it-callout-line--empty { color: var(--text-faint); font-style: italic; }
  .do-it-callout--error .do-it-callout-line { color: var(--error); }

  .do-it-callout-dismiss {
    flex-shrink: 0; background: none; border: none; cursor: pointer;
    color: var(--text-faint); font-size: 14px; line-height: 1;
    padding: 1px 3px; border-radius: 3px;
    transition: background 0.1s, color 0.1s; margin-top: 1px;
  }
  .do-it-callout-dismiss:hover { background: var(--border-soft); color: var(--text-muted); }

  .do-it-callout-spinner {
    display: block; width: 16px; height: 16px;
    animation: callout-spin 0.9s linear infinite; color: var(--text-faint);
  }
</style>
