<script>
  import { onDestroy } from 'svelte';
  import { get } from 'svelte/store';
  import { cells, designAssets, designCode, simulateOpen, simulateState } from './stores.js';
  import {
    createSimulationSession,
    dispatchSimulationEvent,
    disposeSimulationSession,
    setSimulationProperty,
  } from './falcon-wasm.js';
  import {
    designTreeToComponentDefinitions,
    designTreeToInitialState,
    mergeSimulationStatePatch,
    parseDesignSchemaResult,
    unsupportedSimulationComponents,
  } from './design-schema-tree.js';
  import SimulationComponent from './SimulationComponent.svelte';

  const REBUILD_DELAY = 600;
  const DEFAULT_NOTICE_DURATION = 3500;
  const DIALOG_KINDS = new Set(['message', 'choose', 'text', 'password', 'progress']);

  let sessionId = null;
  let root = null;
  let componentState = {};
  let rebuildTimer = null;
  let rebuildSeq = 0;
  let rebuilding = false;
  let actionTokens = {};
  let unsupportedEntries = [];
  let notices = [];
  let noticeSeed = 0;
  let noticeTimers = new Map();
  let activeDialog = null;
  let dialogSeed = 0;
  let dialogInput = '';
  let posX = 0;
  let posY = 0;
  let dragging = false;
  let dragAnchor = { mx: 0, my: 0, px: 0, py: 0 };

  $: sourceSignature = JSON.stringify({
    open: $simulateOpen,
    design: $designCode,
    code: joinedSource($cells),
    assets: ($designAssets || []).map(asset => asset?.name || asset).join('|'),
  });

  $: if ($simulateOpen && sourceSignature) scheduleRebuild();
  $: if (!$simulateOpen && sessionId !== null) disposeCurrentSession();

  function joinedSource(list) {
    return (list || [])
      .filter(cell => cell.type === 'code')
      .map(cell => cell.code || '')
      .join('\n\n')
      .trim();
  }

  function scheduleRebuild() {
    clearTimeout(rebuildTimer);
    rebuildTimer = setTimeout(() => rebuildSimulation(), REBUILD_DELAY);
  }

  function startDrag(e) {
    e.preventDefault();
    dragging = true;
    dragAnchor = { mx: e.clientX, my: e.clientY, px: posX, py: posY };
    window.addEventListener('pointermove', onPointerMove);
    window.addEventListener('pointerup', stopDrag);
    window.addEventListener('pointercancel', stopDrag);
  }

  function onPointerMove(e) {
    posX = dragAnchor.px + (e.clientX - dragAnchor.mx);
    posY = dragAnchor.py + (e.clientY - dragAnchor.my);
  }

  function stopDrag() {
    dragging = false;
    window.removeEventListener('pointermove', onPointerMove);
    window.removeEventListener('pointerup', stopDrag);
    window.removeEventListener('pointercancel', stopDrag);
  }

  function clearNoticeTimer(id) {
    const timer = noticeTimers.get(id);
    if (noticeTimers.has(id)) clearTimeout(timer);
    noticeTimers.delete(id);
  }

  function clearNoticeTimers() {
    for (const timer of noticeTimers.values()) clearTimeout(timer);
    noticeTimers.clear();
  }

  function clearTransientEffects() {
    clearNoticeTimers();
    notices = [];
    activeDialog = null;
    dialogInput = '';
  }

  async function disposeCurrentSession() {
    const old = sessionId;
    sessionId = null;
    root = null;
    componentState = {};
    actionTokens = {};
    unsupportedEntries = [];
    clearTransientEffects();
    simulateState.set({ status: 'idle', sessionId: null, error: null, diagnostics: [] });
    if (old !== null) {
      try { await disposeSimulationSession(old); } catch {}
    }
  }

  async function clearSessionAfterFailure(message, diagnostics = []) {
    const old = sessionId;
    sessionId = null;
    root = null;
    componentState = {};
    actionTokens = {};
    unsupportedEntries = [];
    clearTransientEffects();
    simulateState.set({ status: 'error', sessionId: null, error: message, diagnostics });
    if (old !== null) {
      try { await disposeSimulationSession(old); } catch {}
    }
  }

  async function rebuildSimulation() {
    if (!$simulateOpen) return;
    const seq = ++rebuildSeq;
    rebuilding = true;

    const parsed = parseDesignSchemaResult(get(designCode));
    if (parsed.error) {
      rebuilding = false;
      await clearSessionAfterFailure(parsed.error);
      return;
    }

    const nextRoot = parsed.root;
    const nextInitialState = designTreeToInitialState(nextRoot);
    const componentDefs = designTreeToComponentDefinitions(nextRoot);
    const staticUnsupported = unsupportedSimulationComponents(nextRoot);
    const source = joinedSource(get(cells));
    const result = await createSimulationSession(source, componentDefs, nextInitialState);
    if (seq !== rebuildSeq) {
      if (result.sessionId) await disposeSimulationSession(result.sessionId);
      return;
    }

    rebuilding = false;
    if (!result.ok) {
      const errorText = result.error || result.diagnostics?.[0]?.message || 'Simulation failed to build.';
      await clearSessionAfterFailure(errorText, result.diagnostics || []);
      return;
    }

    const oldSession = sessionId;
    sessionId = result.sessionId;
    root = nextRoot;
    componentState = mergeSimulationStatePatch(nextInitialState, result.statePatch);
    unsupportedEntries = mergeUnsupported([], [...staticUnsupported, ...(result.unsupported || [])]);
    clearTransientEffects();
    applyEffects(result.effects);
    simulateState.set({ status: 'running', sessionId, error: null, diagnostics: [] });
    if (oldSession !== null && oldSession !== sessionId) {
      try { await disposeSimulationSession(oldSession); } catch {}
    }
  }

  async function applyPropertyPatch(component, property, value) {
    componentState = mergeSimulationStatePatch(componentState, { [component]: { [property]: value } });
    if (sessionId === null) return null;
    const result = await setSimulationProperty(sessionId, component, property, value);
    if (result.ok) {
      componentState = mergeSimulationStatePatch(componentState, result.statePatch);
      addUnsupported(result.unsupported);
      applyEffects(result.effects);
    } else {
      simulateState.set({
        status: 'error',
        sessionId,
        error: result.error || 'Unable to update property.',
        diagnostics: result.diagnostics || [],
      });
    }
    return result;
  }

  function applyEffects(effects = []) {
    let next = actionTokens;
    for (const effect of effects || []) {
      if (effect?.type === 'notice') {
        addNotice(effect);
        continue;
      }
      if (effect?.type === 'dialog') {
        activeDialog = normalizeDialog(effect);
        dialogInput = '';
        continue;
      }
      if (effect?.type === 'dialog-dismiss') {
        dismissDialogEffect(effect);
        continue;
      }
      if (effect?.type !== 'component-action' || !effect.component || !effect.action) continue;
      next = {
        ...next,
        [effect.component]: {
          ...(next[effect.component] || {}),
          [effect.action]: ((next[effect.component]?.[effect.action] || 0) + 1),
          ...(effect.position !== undefined ? { position: effect.position } : {}),
        },
      };
    }
    actionTokens = next;
  }

  function addNotice(effect) {
    const notice = {
      id: ++noticeSeed,
      component: effect.component || '',
      text: String(effect.text ?? ''),
      backgroundColor: colorValue(effect.backgroundColor),
      textColor: colorValue(effect.textColor),
    };
    const nextNotices = [...notices.slice(-2), notice];
    const nextIds = new Set(nextNotices.map(item => item.id));
    for (const oldNotice of notices) {
      if (!nextIds.has(oldNotice.id)) clearNoticeTimer(oldNotice.id);
    }
    notices = nextNotices;
    scheduleNoticeDismiss(notice.id, effect.duration);
  }

  function scheduleNoticeDismiss(id, duration) {
    const rawDuration = duration === undefined || duration === null || duration === ''
      ? DEFAULT_NOTICE_DURATION
      : Number(duration);
    if (!Number.isFinite(rawDuration)) return;
    const timeout = setTimeout(() => dismissNotice(id), Math.max(0, rawDuration));
    noticeTimers.set(id, timeout);
  }

  function colorValue(value) {
    if (value === undefined || value === null || value === '') return null;
    if (typeof value === 'number' && Number.isFinite(value)) return argbToCss(value >>> 0);

    const text = String(value).trim();
    const ai = text.match(/^&H([0-9A-Fa-f]{2})([0-9A-Fa-f]{2})([0-9A-Fa-f]{2})([0-9A-Fa-f]{2})$/);
    if (ai) {
      return argbToCss([
        ai[1],
        ai[2],
        ai[3],
        ai[4],
      ].map(part => parseInt(part, 16)));
    }
    if (/^-?\d+$/.test(text)) return argbToCss(Number(text) >>> 0);
    if (/^#[0-9A-Fa-f]{3,8}$/.test(text)) return text;
    if (/^(?:rgb|rgba|hsl|hsla|oklch|oklab|color)\(/i.test(text)) return text;
    if (/^[a-z]+$/i.test(text)) return text;
    return text;
  }

  function argbToCss(value) {
    let a;
    let r;
    let g;
    let b;
    if (Array.isArray(value)) {
      [a, r, g, b] = value;
    } else {
      a = (value >>> 24) & 0xff;
      r = (value >>> 16) & 0xff;
      g = (value >>> 8) & 0xff;
      b = value & 0xff;
    }
    if (a >= 255) return `#${hexByte(r)}${hexByte(g)}${hexByte(b)}`;
    return `rgba(${r}, ${g}, ${b}, ${roundAlpha(a / 255)})`;
  }

  function hexByte(value) {
    return Math.max(0, Math.min(255, Number(value) || 0)).toString(16).padStart(2, '0');
  }

  function roundAlpha(value) {
    return Math.round(value * 1000) / 1000;
  }

  function normalizeDialog(effect) {
    const requestedKind = effect.dialogKind || effect.dialogType;
    const kind = DIALOG_KINDS.has(requestedKind) ? requestedKind : 'message';
    return {
      id: ++dialogSeed,
      kind,
      component: effect.component || '',
      title: String(effect.title ?? ''),
      message: String(effect.message ?? ''),
      buttonText: String(effect.buttonText ?? 'OK'),
      button1Text: String(effect.button1Text ?? 'OK'),
      button2Text: String(effect.button2Text ?? 'Cancel'),
      cancelable: dialogCancelable(effect.cancelable, kind),
    };
  }

  function dialogCancelable(value, kind) {
    if (kind === 'progress' || kind === 'message') return false;
    if (value === undefined || value === null || value === '') return true;
    return !(value === false || value === 0 || value === 'false' || value === '0');
  }

  function dismissDialogEffect(effect) {
    if (!activeDialog || activeDialog.kind !== 'progress') return;
    if (!effect.component || activeDialog.component === effect.component) closeActiveDialog(activeDialog);
  }

  function mergeUnsupported(current = [], entries = []) {
    const out = [...current];
    const seen = new Set(out.map(entry => `${entry?.kind || 'unsupported'}:${entry?.detail || ''}`));
    for (const entry of entries || []) {
      if (!entry) continue;
      const normalized = {
        kind: String(entry.kind || 'unsupported'),
        detail: String(entry.detail || ''),
      };
      const key = `${normalized.kind}:${normalized.detail}`;
      if (!normalized.detail || seen.has(key)) continue;
      seen.add(key);
      out.push(normalized);
    }
    return out;
  }

  function addUnsupported(entries = []) {
    if (!entries?.length) return;
    unsupportedEntries = mergeUnsupported(unsupportedEntries, entries);
  }

  function dismissNotice(id) {
    clearNoticeTimer(id);
    notices = notices.filter(notice => notice.id !== id);
  }

  function closeActiveDialog(dialog = activeDialog) {
    if (!dialog || !activeDialog || dialog.id === activeDialog.id) {
      activeDialog = null;
      dialogInput = '';
    }
  }

  async function emitDialogEvent(dialog, event, args = []) {
    if (!dialog?.component) return;
    await handleEvent({ detail: { component: dialog.component, event, args } });
  }

  function closeMessageDialog() {
    closeActiveDialog(activeDialog);
  }

  async function chooseDialog(choice) {
    const dialog = activeDialog;
    closeActiveDialog(dialog);
    await emitDialogEvent(dialog, 'AfterChoosing', [choice]);
  }

  async function cancelChooseDialog() {
    const dialog = activeDialog;
    closeActiveDialog(dialog);
    await emitDialogEvent(dialog, 'ChoosingCanceled');
    await emitDialogEvent(dialog, 'AfterChoosing', ['Cancel']);
  }

  async function submitTextDialog() {
    const dialog = activeDialog;
    const response = dialogInput;
    closeActiveDialog(dialog);
    await emitDialogEvent(dialog, 'AfterTextInput', [response]);
  }

  async function cancelTextDialog() {
    const dialog = activeDialog;
    closeActiveDialog(dialog);
    await emitDialogEvent(dialog, 'TextInputCanceled');
    // App Inventor's Notifier also fires AfterTextInput("Cancel") for
    // backward compatibility, mirroring the choose-dialog cancel path.
    await emitDialogEvent(dialog, 'AfterTextInput', ['Cancel']);
  }

  async function handleProperty(e) {
    const { component, property, value } = e.detail;
    await applyPropertyPatch(component, property, value);
  }

  async function handleEvent(e) {
    if (sessionId === null) return;
    const { component, event, args } = e.detail;
    const result = await dispatchSimulationEvent(sessionId, component, event, args || []);
    if (result.ok) {
      componentState = mergeSimulationStatePatch(componentState, result.statePatch);
      addUnsupported(result.unsupported);
      applyEffects(result.effects);
    } else {
      simulateState.set({
        status: 'error',
        sessionId,
        error: result.error || 'Simulation runtime error.',
        diagnostics: result.diagnostics || [],
      });
    }
  }

  async function runSimulationEvent(detail) {
    await handleEvent({ detail });
  }

  async function handleInteraction(e) {
    for (const change of e.detail.properties || []) {
      const result = await applyPropertyPatch(change.component, change.property, change.value);
      if (result && !result.ok) return;
    }
    if (e.detail.event) {
      await handleEvent({ detail: e.detail.event });
    }
  }

  function restart() {
    clearTimeout(rebuildTimer);
    rebuildSimulation();
  }

  onDestroy(() => {
    clearTimeout(rebuildTimer);
    clearNoticeTimers();
    window.removeEventListener('pointermove', onPointerMove);
    window.removeEventListener('pointerup', stopDrag);
    window.removeEventListener('pointercancel', stopDrag);
    if (sessionId !== null) disposeSimulationSession(sessionId);
  });
</script>

{#if $simulateOpen}
  <section
    class="sim-root"
    class:is-dragging={dragging}
    style="transform: translate(calc(-50% + {posX}px), calc(-50% + {posY}px))"
    aria-label="Simulator"
  >
    <div class="sim-inner">
      <div class="sim-bar">
        <button
          class="sim-drag-btn"
          on:pointerdown={startDrag}
          aria-label="Drag to reposition simulator"
          title="Drag to move"
        >
          <svg viewBox="0 0 10 14" fill="none" aria-hidden="true">
            <circle cx="3" cy="2" r="1.2" fill="currentColor" />
            <circle cx="7" cy="2" r="1.2" fill="currentColor" />
            <circle cx="3" cy="7" r="1.2" fill="currentColor" />
            <circle cx="7" cy="7" r="1.2" fill="currentColor" />
            <circle cx="3" cy="12" r="1.2" fill="currentColor" />
            <circle cx="7" cy="12" r="1.2" fill="currentColor" />
          </svg>
        </button>

        <span class="sim-label">Simulator</span>
        <div class="sim-spacer" />

        <span
          class="sim-status-dot"
          class:dot-live={$simulateState.status === 'running' && !rebuilding}
          class:dot-building={rebuilding}
          class:dot-error={$simulateState.status === 'error' && !rebuilding}
          title={rebuilding
            ? 'Building...'
            : $simulateState.status === 'error'
            ? 'Error'
            : $simulateState.status === 'running'
            ? 'Live'
            : 'Idle'}
        />

        <div class="sim-sep" aria-hidden="true" />

        <button
          class="sim-icon-btn"
          on:click={restart}
          title="Restart simulation"
          aria-label="Restart simulation"
          disabled={rebuilding}
        >
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path d="M11.5 7A4.5 4.5 0 1 1 7 2.5" />
            <path d="M5 1l2 1.5-2 1.5" />
          </svg>
        </button>

        <button
          class="sim-icon-btn sim-close-btn"
          on:click={() => simulateOpen.set(false)}
          title="Close simulator"
          aria-label="Close simulator"
        >
          <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" aria-hidden="true">
            <path d="M2 2l6 6M8 2l-6 6" />
          </svg>
        </button>
      </div>

      <div class="phone-frame">
        <div class="phone-notch" aria-hidden="true">
          <div class="phone-camera" />
        </div>

        <div class="simulate-stage" aria-busy={rebuilding}>
          {#if root}
            <SimulationComponent
              node={root}
              state={componentState}
              assets={$designAssets}
              actions={actionTokens}
              unsupported={unsupportedEntries}
              eventRunner={runSimulationEvent}
              on:event={handleEvent}
              on:property={handleProperty}
              on:interaction={handleInteraction}
            />
          {/if}

          {#if notices.length}
            <div class="sim-notices">
              {#each notices as notice (notice.id)}
                <div
                  class="sim-notice"
                  style:background={notice.backgroundColor}
                  style:color={notice.textColor}
                >
                  <span>{notice.text}</span>
                  <button type="button" aria-label="Dismiss notice" on:click={() => dismissNotice(notice.id)}>x</button>
                </div>
              {/each}
            </div>
          {/if}

          {#if activeDialog}
            <div class="sim-dialog-layer">
              <div
                class="sim-dialog"
                class:sim-dialog--progress={activeDialog.kind === 'progress'}
                role="dialog"
                aria-modal="true"
                aria-label={activeDialog.title || 'Notifier dialog'}
                aria-describedby={activeDialog.message ? 'sim-dialog-message' : undefined}
              >
                {#if activeDialog.title}
                  <div class="sim-dialog-title">{activeDialog.title}</div>
                {/if}

                {#if activeDialog.kind === 'progress'}
                  <div class="sim-dialog-progress-row">
                    <div class="sim-dialog-spinner" aria-hidden="true" />
                    {#if activeDialog.message}
                      <div class="sim-dialog-message" id="sim-dialog-message">{activeDialog.message}</div>
                    {/if}
                  </div>
                {:else}
                  {#if activeDialog.message}
                    <div class="sim-dialog-message" id="sim-dialog-message">{activeDialog.message}</div>
                  {/if}

                  {#if activeDialog.kind === 'message'}
                    <div class="sim-dialog-actions">
                      <button
                        type="button"
                        class="sim-dialog-btn sim-dialog-btn--primary"
                        on:click={closeMessageDialog}
                      >
                        {activeDialog.buttonText}
                      </button>
                    </div>
                  {:else if activeDialog.kind === 'choose'}
                    <div class="sim-dialog-actions">
                      {#if activeDialog.cancelable}
                        <button
                          type="button"
                          class="sim-dialog-btn"
                          on:click={cancelChooseDialog}
                        >
                          Cancel
                        </button>
                      {/if}
                      <button
                        type="button"
                        class="sim-dialog-btn"
                        on:click={() => chooseDialog(activeDialog.button2Text)}
                      >
                        {activeDialog.button2Text}
                      </button>
                      <button
                        type="button"
                        class="sim-dialog-btn sim-dialog-btn--primary"
                        on:click={() => chooseDialog(activeDialog.button1Text)}
                      >
                        {activeDialog.button1Text}
                      </button>
                    </div>
                  {:else if activeDialog.kind === 'text' || activeDialog.kind === 'password'}
                    <form class="sim-dialog-form" on:submit|preventDefault={submitTextDialog}>
                      {#if activeDialog.kind === 'password'}
                        <input
                          class="sim-dialog-input"
                          type="password"
                          bind:value={dialogInput}
                          aria-label="Dialog response"
                        />
                      {:else}
                        <input
                          class="sim-dialog-input"
                          type="text"
                          bind:value={dialogInput}
                          aria-label="Dialog response"
                        />
                      {/if}
                      <div class="sim-dialog-actions">
                        {#if activeDialog.cancelable}
                          <button
                            type="button"
                            class="sim-dialog-btn"
                            on:click={cancelTextDialog}
                          >
                            Cancel
                          </button>
                        {/if}
                        <button type="submit" class="sim-dialog-btn sim-dialog-btn--primary">
                          {activeDialog.buttonText}
                        </button>
                      </div>
                    </form>
                  {/if}
                {/if}
              </div>
            </div>
          {/if}

          {#if unsupportedEntries.length}
            <div class="sim-unsupported-panel">
              {#each unsupportedEntries.slice(0, 3) as entry}
                <div>{entry.kind}: {entry.detail}</div>
              {/each}
              {#if unsupportedEntries.length > 3}
                <div>+{unsupportedEntries.length - 3} more unsupported item{unsupportedEntries.length - 3 === 1 ? '' : 's'}</div>
              {/if}
            </div>
          {/if}
        </div>

        <div class="phone-chin" aria-hidden="true">
          <div class="phone-home-pill" />
        </div>
      </div>
    </div>
  </section>
{/if}

<style>
  .sim-root {
    position: fixed;
    z-index: 80;
    top: 50%;
    left: 50%;
    width: min(320px, calc(100vw - 32px), calc((100vh - 110px) * 9 / 20 + 16px));
    user-select: none;
    -webkit-user-select: none;
  }

  .sim-root.is-dragging {
    cursor: grabbing;
  }

  .sim-inner {
    display: flex;
    flex-direction: column;
    gap: 5px;
  }

  .sim-bar {
    display: flex;
    align-items: center;
    height: 34px;
    padding: 0 6px;
    gap: 2px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 8px;
    box-shadow:
      0 2px 8px oklch(10% 0.01 55 / 0.08),
      0 1px 2px oklch(10% 0.01 55 / 0.05);
  }

  .sim-drag-btn,
  .sim-icon-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    padding: 0;
    border: none;
    background: none;
    color: var(--text-muted);
    border-radius: 4px;
    flex-shrink: 0;
  }

  .sim-drag-btn {
    cursor: grab;
    color: var(--text-faint);
  }

  .sim-icon-btn {
    cursor: pointer;
  }

  .sim-drag-btn:hover,
  .sim-icon-btn:hover {
    color: var(--text);
    background: var(--cell-active);
  }

  .is-dragging .sim-drag-btn {
    cursor: grabbing;
    color: var(--text-muted);
  }

  .sim-drag-btn:focus-visible,
  .sim-icon-btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .sim-drag-btn svg {
    width: 10px;
    height: 14px;
  }

  .sim-icon-btn svg {
    width: 14px;
    height: 14px;
  }

  .sim-icon-btn:disabled {
    opacity: 0.32;
    cursor: default;
  }

  .sim-close-btn:hover {
    color: var(--error);
    background: var(--error-soft);
  }

  .sim-close-btn svg {
    width: 10px;
    height: 10px;
  }

  .sim-label {
    font-family: 'DM Sans', sans-serif;
    font-size: 11px;
    font-weight: 500;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--text-muted);
    margin-left: 2px;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .sim-spacer {
    flex: 1;
  }

  .sim-status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--border);
    flex-shrink: 0;
    transition: background 0.25s;
  }

  .sim-status-dot.dot-live {
    background: var(--run-dot);
  }

  .sim-status-dot.dot-error {
    background: var(--error);
  }

  .sim-status-dot.dot-building {
    background: #c9900a;
    animation: dot-pulse 1.1s ease-in-out infinite;
  }

  @keyframes dot-pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.3; }
  }

  .sim-sep {
    width: 1px;
    height: 16px;
    background: var(--border-soft);
    margin: 0 3px;
    flex-shrink: 0;
  }

  .phone-frame {
    background: #1a1916;
    border-radius: 32px;
    padding: 14px 8px 20px;
    position: relative;
    box-shadow:
      0 0 0 1px #2d2b28,
      0 28px 60px oklch(8% 0.025 45 / 0.5),
      0 8px 22px oklch(8% 0.025 45 / 0.28);
  }

  .phone-notch {
    display: flex;
    justify-content: center;
    padding-bottom: 10px;
  }

  .phone-camera {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #0f0e0d;
    border: 1.5px solid #292724;
    box-shadow: 0 0 0 1px #0a0908 inset;
  }

  .simulate-stage {
    width: 100%;
    aspect-ratio: 9 / 20;
    background: #fff;
    overflow: hidden;
    position: relative;
    border-radius: 3px;
  }

  .sim-notices,
  .sim-unsupported-panel {
    position: absolute;
    left: 10px;
    right: 10px;
    z-index: 3;
    display: grid;
    gap: 6px;
    pointer-events: none;
  }

  .sim-notices {
    bottom: 10px;
  }

  .sim-notice {
    display: grid;
    grid-template-columns: 1fr auto;
    align-items: center;
    gap: 8px;
    min-height: 34px;
    padding: 7px 8px 7px 10px;
    border-radius: 5px;
    background: rgba(32, 33, 36, 0.94);
    color: #fff;
    font-size: 12px;
    line-height: 1.25;
    box-shadow: 0 8px 22px rgba(0, 0, 0, 0.22);
    pointer-events: auto;
  }

  .sim-notice span {
    min-width: 0;
    overflow-wrap: anywhere;
  }

  .sim-notice button {
    width: 22px;
    height: 22px;
    border: 0;
    border-radius: 3px;
    background: color-mix(in srgb, currentColor 14%, transparent);
    color: inherit;
    cursor: pointer;
    line-height: 1;
  }

  .sim-unsupported-panel {
    top: 10px;
    padding: 7px 8px;
    border: 1px solid #f59e0b;
    border-radius: 5px;
    background: rgba(255, 251, 235, 0.96);
    color: #92400e;
    font-size: 11px;
    line-height: 1.3;
  }

  .sim-dialog-layer {
    position: absolute;
    inset: 0;
    z-index: 4;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 18px;
    background: rgba(26, 25, 22, 0.34);
    pointer-events: auto;
  }

  .sim-dialog {
    width: min(100%, 244px);
    max-height: 88%;
    overflow: auto;
    box-sizing: border-box;
    padding: 14px 14px 12px;
    border: 1px solid rgba(26, 25, 22, 0.14);
    border-radius: 6px;
    background: #fffefa;
    color: #1a1916;
    font-family: 'DM Sans', sans-serif;
    box-shadow:
      0 18px 36px rgba(26, 25, 22, 0.28),
      0 2px 8px rgba(26, 25, 22, 0.16);
  }

  .sim-dialog--progress {
    padding: 15px 14px;
  }

  .sim-dialog-title {
    margin-bottom: 7px;
    color: #1a1916;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.25;
    overflow-wrap: anywhere;
  }

  .sim-dialog-message {
    color: #403e38;
    font-size: 12px;
    line-height: 1.45;
    overflow-wrap: anywhere;
    white-space: pre-wrap;
  }

  .sim-dialog-form {
    display: grid;
    gap: 10px;
    margin-top: 10px;
  }

  .sim-dialog-input {
    width: 100%;
    height: 32px;
    box-sizing: border-box;
    border: 1px solid #d9d5cb;
    border-radius: 4px;
    background: #fafaf8;
    color: #1a1916;
    font: 400 12px/1.2 'DM Sans', sans-serif;
    padding: 0 8px;
  }

  .sim-dialog-input:focus {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
    border-color: var(--accent);
  }

  .sim-dialog-actions {
    display: flex;
    justify-content: flex-end;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 13px;
  }

  .sim-dialog-form .sim-dialog-actions {
    margin-top: 0;
  }

  .sim-dialog-btn {
    min-width: 48px;
    min-height: 30px;
    max-width: 100%;
    padding: 5px 10px;
    border: 1px solid #d9d5cb;
    border-radius: 4px;
    background: #fafaf8;
    color: #403e38;
    font: 500 12px/1.2 'DM Sans', sans-serif;
    cursor: pointer;
    overflow-wrap: anywhere;
  }

  .sim-dialog-btn:hover {
    background: #f0ede8;
    color: #1a1916;
  }

  .sim-dialog-btn:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .sim-dialog-btn--primary {
    border-color: #1a1916;
    background: #1a1916;
    color: #fffefa;
  }

  .sim-dialog-btn--primary:hover {
    background: #2d2b28;
    color: #fffefa;
  }

  .sim-dialog-progress-row {
    display: flex;
    align-items: center;
    gap: 10px;
    min-height: 28px;
  }

  .sim-dialog-spinner {
    width: 18px;
    height: 18px;
    flex: 0 0 auto;
    border: 2px solid #d9d5cb;
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: sim-dialog-spin 0.85s linear infinite;
  }

  @keyframes sim-dialog-spin {
    to { transform: rotate(360deg); }
  }

  .phone-chin {
    display: flex;
    justify-content: center;
    padding-top: 12px;
  }

  .phone-home-pill {
    width: 34px;
    height: 4px;
    border-radius: 2px;
    background: #363330;
  }
</style>
