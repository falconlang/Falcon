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
  import { falconCellToBlocklyPng, warmBlocklyPreviewRuntime } from './blockly-preview.js';
  import { runCodeDiagnosticResult } from './falcon-wasm.js';
  import { falconLang, falconSyntaxHighlighting } from './falcon-cm.js';

  // CodeMirror imports
  import { EditorView, keymap, lineNumbers, Decoration, WidgetType, gutter, GutterMarker } from '@codemirror/view';
  import { EditorState, StateField, StateEffect, RangeSetBuilder } from '@codemirror/state';
  import { defaultKeymap, historyKeymap, history, toggleComment, indentWithTab } from '@codemirror/commands';
  import { closeBrackets, closeBracketsKeymap } from '@codemirror/autocomplete';
  import { indentOnInput } from '@codemirror/language';

  export let cell;
  export let active = false;

  // DOM refs
  let editorWrapEl;   // outer .code-editor div
  let cmEl;           // div that CodeMirror mounts into
  let view;           // EditorView instance

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

  const BLOCKLY_HEIGHT_TRANSITION_MS = 180;

  $: isDoItVisible = $doItCellId === cell.id;
  $: isCompanionConnected = $liveTestState.status === 'connected';
  $: doItResult = $doItResults[cell.id] ?? null;
  $: debugActiveLine = $debugModeEnabled && $debugActiveLocation?.cellId === cell.id
    ? $debugActiveLocation.cellLine : null;
  $: debugError = $debugModeEnabled ? ($debugRuntimeErrors[cell.id] ?? null) : null;
  $: debugErrorLines = debugError?.message ? String(debugError.message).split(/\r?\n/) : [];

  // Sync cell.code → CM when changed externally
  $: if (view && !blocklyActive) {
    const incoming = cell.code || '';
    const current = view.state.doc.toString();
    if (incoming !== current) {
      view.dispatch({
        changes: { from: 0, to: current.length, insert: incoming },
        annotations: [externalChange.of(true)],
      });
    }
  }

  // Update debug decorations reactively
  $: if (view) {
    view.dispatch({ effects: [
      setDebugActiveLine.of(debugActiveLine),
      setDebugErrorLine.of(debugError?.cellLine ?? null),
      setDebugErrorMessages.of(debugErrorLines),
    ]});
  }

  $: blocklyEditorStyle = blocklyPreviewHeight ? `height: ${blocklyPreviewHeight}px;` : '';

  $: if (
    $blocklyPreviewRequest
    && $blocklyPreviewRequest.cellId === cell.id
    && $blocklyPreviewRequest.id !== handledBlocklyRequestId
  ) {
    handledBlocklyRequestId = $blocklyPreviewRequest.id;
    previewPastedBlocks($blocklyPreviewRequest.blob);
  }

  // ── CodeMirror state effects ─────────────────────────────────────────────
  const externalChange      = StateEffect.define();
  const setDebugActiveLine  = StateEffect.define();
  const setDebugErrorLine   = StateEffect.define();
  const setDebugErrorMessages = StateEffect.define();

  // ── Debug decorations ─────────────────────────────────────────────────────

  // Gutter markers
  class DebugCurrentMarker extends GutterMarker {
    toDOM() { return document.createTextNode(''); }
    elementClass = 'debug-gutter-current';
  }
  class DebugErrorMarker extends GutterMarker {
    toDOM() { return document.createTextNode(''); }
    elementClass = 'debug-gutter-error';
  }
  const debugCurrentMarker = new DebugCurrentMarker();
  const debugErrorMarker   = new DebugErrorMarker();

  const debugState = StateField.define({
    create() { return { activeLine: null, errorLine: null, errorMessages: [] }; },
    update(val, tr) {
      let next = val;
      for (const e of tr.effects) {
        if (e.is(setDebugActiveLine))   next = { ...next, activeLine: e.value };
        if (e.is(setDebugErrorLine))    next = { ...next, errorLine: e.value };
        if (e.is(setDebugErrorMessages)) next = { ...next, errorMessages: e.value };
      }
      return next;
    },
  });

  // Line backdrops + error inline widget
  class ErrorWidget extends WidgetType {
    constructor(lines) { super(); this.lines = lines; }
    eq(other) { return JSON.stringify(other.lines) === JSON.stringify(this.lines); }
    toDOM() {
      const wrap = document.createElement('div');
      wrap.className = 'cm-debug-error-widget';
      const icon = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
      icon.setAttribute('viewBox', '0 0 14 14');
      icon.setAttribute('fill', 'none');
      icon.setAttribute('stroke', 'currentColor');
      icon.setAttribute('stroke-width', '1.4');
      icon.setAttribute('stroke-linecap', 'round');
      icon.setAttribute('stroke-linejoin', 'round');
      icon.setAttribute('width', '13');
      icon.setAttribute('height', '13');
      icon.setAttribute('aria-hidden', 'true');
      icon.innerHTML = '<ellipse cx="7" cy="7.5" rx="3" ry="3.5"/><path d="M7 4V2.5"/><path d="M4 7.5H2M12 7.5h-2"/><path d="M4.5 5.2l-1.2-1.2M9.5 5.2l1.2-1.2"/><path d="M4.5 9.8l-1.2 1.2M9.5 9.8l1.2 1.2"/>';
      const body = document.createElement('div');
      body.className = 'cm-debug-error-widget-body';
      const msgs = this.lines.length ? this.lines : ['Runtime error'];
      for (const msg of msgs) {
        const d = document.createElement('div'); d.textContent = msg; body.appendChild(d);
      }
      wrap.appendChild(icon);
      wrap.appendChild(body);
      return wrap;
    }
    ignoreEvent() { return true; }
  }

  const debugDecorations = EditorView.decorations.compute([debugState], state => {
    const { activeLine, errorLine, errorMessages } = state.field(debugState);
    const builder = new RangeSetBuilder();
    const entries = [];

    if (activeLine != null && activeLine >= 1 && activeLine <= state.doc.lines) {
      const line = state.doc.line(activeLine);
      entries.push({ from: line.from, to: line.from, dec: Decoration.line({ class: 'cm-debug-active-line' }) });
    }

    if (errorLine != null && errorLine >= 1 && errorLine <= state.doc.lines) {
      const line = state.doc.line(errorLine);
      entries.push({ from: line.from, to: line.from, dec: Decoration.line({ class: 'cm-debug-error-line-bg' }) });
      entries.push({
        from: line.to, to: line.to,
        dec: Decoration.widget({ widget: new ErrorWidget(errorMessages), block: true, side: 1 }),
      });
    }

    entries.sort((a, b) => a.from - b.from || a.to - b.to);
    for (const e of entries) builder.add(e.from, e.to, e.dec);
    return builder.finish();
  });

  const debugGutter = gutter({
    class: 'cm-debug-gutter',
    lineMarker(view, line) {
      const { activeLine, errorLine } = view.state.field(debugState);
      const lineNo = view.state.doc.lineAt(line.from).number;
      if (lineNo === activeLine) return debugCurrentMarker;
      if (lineNo === errorLine)  return debugErrorMarker;
      return null;
    },
    lineMarkerChange: update => update.transactions.some(tr => tr.effects.some(e => e.is(setDebugActiveLine) || e.is(setDebugErrorLine))),
  });

  // ── Selection tracking for "Do it" ───────────────────────────────────────
  const selectionListener = EditorView.updateListener.of(update => {
    if (!update.selectionSet && !update.docChanged) return;
    const sel = update.state.selection.main;
    doItCellId.set(sel.from !== sel.to ? cell.id : null);
  });

  // ── Keyboard extensions ───────────────────────────────────────────────────
  function makeKeymap() {
    return keymap.of([
      // Shift+Enter → run cell
      {
        key: 'Shift-Enter',
        run(v) {
          runCell();
          const currentCells = get(cells);
          const idx = currentCells.findIndex(c => c.id === cell.id);
          if (idx < currentCells.length - 1) setActive(currentCells[idx + 1].id);
          return true;
        },
      },
      // Cmd/Ctrl+/ → toggle comment (handled by toggleComment from @codemirror/commands)
      { key: 'Mod-/', run: toggleComment },
      // Tab → 2 spaces
      { key: 'Tab', run: indentWithTab.run, shift: indentWithTab.shift },
      ...closeBracketsKeymap,
      ...defaultKeymap,
      ...historyKeymap,
    ]);
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
          makeKeymap(),
          // Write changes back to store
          EditorView.updateListener.of(update => {
            if (!update.docChanged) return;
            if (update.transactions.some(tr => tr.annotation(externalChange))) return;
            const code = update.state.doc.toString();
            updateCellCode(cell.id, code);
          }),
          // Focus → set active cell
          EditorView.domEventHandlers({
            focus() { setActive(cell.id); },
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
              background: 'transparent',
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
    view = buildView(cmEl, cell.code || '');
    // Trigger initial debug decoration state
    if (debugActiveLine != null || debugError?.cellLine != null) {
      view.dispatch({ effects: [
        setDebugActiveLine.of(debugActiveLine),
        setDebugErrorLine.of(debugError?.cellLine ?? null),
        setDebugErrorMessages.of(debugErrorLines),
      ]});
    }
  });

  onDestroy(() => {
    blocklyPreviewRun += 1;
    blocklyActive = false;
    blocklyStatus = 'idle';
    clearBlocklyHeightReleaseTimer();
    clearBlocklyPreviewAsset();
    view?.destroy();
  });

  // ── Helpers ───────────────────────────────────────────────────────────────
  function selectedSource() {
    if (!view) return '';
    const { from, to } = view.state.selection.main;
    if (from === to) return '';
    return view.state.sliceDoc(from, to);
  }

  function appendRunOutput(result, sourceLabel) {
    const entries = [];
    for (const line of result.output || []) entries.push({ level: 'info', source: sourceLabel, message: line });
    if (!result.ok) {
      const diagnostics = result.diagnostics?.length
        ? result.diagnostics : [{ message: result.error || 'Falcon execution failed' }];
      for (const diagnostic of diagnostics) {
        const location = diagnostic.line
          ? `:${diagnostic.line}${diagnostic.column ? `:${diagnostic.column}` : ''}` : '';
        entries.push({ level: 'error', source: `${sourceLabel}${location}`, message: diagnostic.message || result.error || 'Falcon execution failed' });
      }
    }
    if (entries.length) { appendDebugLogs(entries); debugCollapsed.set(false); }
  }

  async function runCell(sourceOverride = null, sourceLabel = `Cell ${cell.id}`) {
    if (running) return;
    const source = typeof sourceOverride === 'string'
      ? sourceOverride
      : (blocklyActive ? blocklySourceSnapshot : (view?.state.doc.toString() || ''));
    if (!source.trim()) return;

    running = true;
    const icon = document.getElementById(`run-icon-${cell.id}`);
    if (icon) icon.innerHTML = `<circle cx="7" cy="7" r="4" fill="none" stroke="currentColor" stroke-width="1.5" stroke-dasharray="10" stroke-dashoffset="10" class="running"/>`;

    try {
      const result = await runCodeDiagnosticResult(source);
      appendRunOutput(result, sourceLabel);
      updateCellExecCount(cell.id);
      flashing = true;
      setTimeout(() => { flashing = false; }, 400);
    } catch (e) {
      appendRunOutput({ ok: false, error: e?.message || String(e), diagnostics: [], output: [] }, sourceLabel);
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
          ? result.diagnostics : [{ message: result.error || 'Execution failed' }];
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
      companionCommand.set({ type: 'do-it', source, label: `Selection ${cell.id}`, cellId: cell.id });
      doItCellId.set(null);
      return;
    }
    runDoIt(source);
    doItCellId.set(null);
  }

  function handleDoItMouseDown(e) { e.preventDefault(); e.stopPropagation(); runSelection(); }
  function handleDoItClick(e) { e.stopPropagation(); if (e.detail === 0) runSelection(); }

  // ── Blockly preview ───────────────────────────────────────────────────────
  function clearBlocklyPreviewAsset() {
    if (blocklyPreviewUrl) URL.revokeObjectURL(blocklyPreviewUrl);
    blocklyPreviewUrl = ''; blocklyPreviewBlob = null;
  }

  function clearBlocklyHeightReleaseTimer() {
    if (!blocklyHeightReleaseTimer) return;
    clearTimeout(blocklyHeightReleaseTimer); blocklyHeightReleaseTimer = null;
  }

  function captureBlocklyPreviewHeight() {
    clearBlocklyHeightReleaseTimer();
    const rect = editorWrapEl?.getBoundingClientRect?.();
    blocklyPreviewHeight = rect?.height ? Math.max(1, Math.ceil(rect.height)) : 0;
  }

  function previewImageTargetHeight(img) {
    const naturalWidth = img?.naturalWidth || 0, naturalHeight = img?.naturalHeight || 0;
    if (!editorWrapEl || !naturalWidth || !naturalHeight) return 0;
    const pane = editorWrapEl.querySelector('.blockly-preview-pane');
    const styles = pane ? getComputedStyle(pane) : null;
    const pL = Number.parseFloat(styles?.paddingLeft || '0') || 0;
    const pR = Number.parseFloat(styles?.paddingRight || '0') || 0;
    const pT = Number.parseFloat(styles?.paddingTop || '0') || 0;
    const pB = Number.parseFloat(styles?.paddingBottom || '0') || 0;
    const avail = Math.max(1, editorWrapEl.clientWidth - pL - pR);
    const scale = Math.min(1, avail / naturalWidth);
    const limit = Math.min(360, Math.max(96, Math.floor(window.innerHeight * 0.56)));
    return Math.max(1, Math.ceil(Math.min(Math.ceil(naturalHeight * scale), limit) + pT + pB));
  }

  function onBlocklyPreviewImageLoad(e) {
    if (!blocklyActive) return;
    const next = previewImageTargetHeight(e.currentTarget);
    if (!next || Math.abs(next - blocklyPreviewHeight) < 1) return;
    blocklyPreviewHeight = next;
  }

  function codeEditorTargetHeight() {
    return Math.max(1, Math.ceil(view?.scrollDOM?.scrollHeight || 0));
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
    blocklyActive = false; blocklyStatus = 'idle'; blocklyError = '';
    clearBlocklyPreviewAsset();
    blocklySourceSnapshot = '';
    // Restore content into CM
    if (view) {
      const current = view.state.doc.toString();
      if (current !== restored) {
        view.dispatch({
          changes: { from: 0, to: current.length, insert: restored },
          annotations: [externalChange.of(true)],
        });
      }
      updateCellCode(cell.id, restored);
    }
    tick().then(() => {
      if (!blocklyPreviewHeight) return;
      blocklyPreviewHeight = codeEditorTargetHeight();
      releaseBlocklyHeightAfterTransition(runId);
    });
  }

  async function startBlocklyPreview() {
    if (blocklyActive || blocklyStatus === 'loading') return;
    const source = view?.state.doc.toString() || cell.code || '';
    if (!source.trim()) return;

    setActive(cell.id);
    const runId = ++blocklyPreviewRun;
    blocklySourceSnapshot = source;
    captureBlocklyPreviewHeight();
    blocklyActive = true; blocklyStatus = 'loading'; blocklyError = '';
    clearBlocklyPreviewAsset();
    doItCellId.set(null);
    await tick();

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
      ? blocklySourceSnapshot : (view?.state.doc.toString() || cell.code || '');
    blocklyPreviewBlob = blob;
    blocklyPreviewUrl = URL.createObjectURL(blob);
    blocklyStatus = 'ready'; blocklyError = ''; blocklyActive = true;
    doItCellId.set(null);
  }

  function activateCellFromKeyboard(e) {
    if (e.target !== e.currentTarget) return;
    if (e.key !== 'Enter' && e.key !== ' ') return;
    e.preventDefault(); setActive(cell.id);
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
    bind:this={editorWrapEl}
    class:blockly-previewing={blocklyActive}
    style={blocklyEditorStyle}
  >
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
    <div class="cm-host" bind:this={cmEl} class:cm-hidden={blocklyActive}></div>
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
</div>
