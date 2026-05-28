<script>
  import { tick } from 'svelte';
  import { clearDebugLogs, debugCollapsed, debugLogs, debugOpenHeight } from './stores.js';

  let dbgScrollEl;
  let lastDebugLogId = 0;
  let didDrag = false;
  let resizing = false;
  let resizeStartY = 0;
  let resizeStartH = 0;

  function toggle() {
    if (didDrag) { didDrag = false; return; }
    debugCollapsed.update(v => !v);
  }

  function handleKey(e) {
    if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggle(); }
  }

  function initResize(e) {
    if ($debugCollapsed) return;
    resizing = true;
    didDrag = false;
    resizeStartY = e.clientY;
    resizeStartH = $debugOpenHeight;
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
    e.preventDefault();
  }

  function onMove(e) {
    if (!resizing) return;
    didDrag = true;
    debugOpenHeight.set(Math.max(60, Math.min(500, resizeStartH + (resizeStartY - e.clientY))));
  }

  function onUp() {
    resizing = false;
    document.removeEventListener('mousemove', onMove);
    document.removeEventListener('mouseup', onUp);
  }

  function levelLabel(level) {
    if (level === 'high') return 'High';
    if (level === 'warn') return 'Warning';
    return level ? level[0].toUpperCase() + level.slice(1) : 'Info';
  }

  $: panelHeight  = $debugCollapsed ? 0 : $debugOpenHeight;
  $: caretPath    = $debugCollapsed ? 'M2 4l3 3 3-3' : 'M2 6l3-3 3 3';
  $: logCount     = $debugLogs.length;
  $: logStatus    = logCount === 0 ? 'No logs' : `${logCount} log${logCount === 1 ? '' : 's'}`;

  $: {
    const newestId = $debugLogs[$debugLogs.length - 1]?.id ?? 0;
    if (dbgScrollEl && newestId !== lastDebugLogId) {
      lastDebugLogId = newestId;
      tick().then(() => { if (dbgScrollEl) dbgScrollEl.scrollTop = dbgScrollEl.scrollHeight; });
    }
  }
</script>

<div
  id="debug-handle"
  on:mousedown={initResize}
  on:click={toggle}
  role="button"
  tabindex="0"
  on:keydown={handleKey}
>
  <span class="dbg-handle-label">
    <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4">
      <circle cx="6" cy="4" r="2"/>
      <path d="M3 11c0-1.657 1.343-3 3-3s3 1.343 3 3"/>
      <path d="M1 4h1M10 4h1M6 1v1"/>
    </svg>
    <span class="dbg-handle-title">Debug</span>
    <span class="dbg-count">{logStatus}</span>
  </span>
  <div class="dbg-actions">
    <button
      class="dbg-clear-btn"
      on:mousedown|stopPropagation
      on:click|stopPropagation={clearDebugLogs}
      title="Clear debug logs"
      aria-label="Clear debug logs"
      disabled={logCount === 0}
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
        <path d="M2 3h8M4.5 3V2h3v1M9 3l-.5 7h-5L3 3"/>
      </svg>
    </button>
    <button
      class="dbg-toggle-btn"
      on:mousedown|stopPropagation
      on:click|stopPropagation={toggle}
      title="Toggle debug panel"
    >
      <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.5">
        <path d={caretPath}/>
      </svg>
    </button>
  </div>
</div>

<div id="debug-panel" style="height: {panelHeight}px">
  <div
    class="dbg-scroll"
    bind:this={dbgScrollEl}
    role="log"
    aria-live="polite"
    aria-label="Debug logs"
  >
    {#if logCount === 0}
      <div class="dbg-empty">No logs yet</div>
    {:else}
      {#each $debugLogs as log (log.id)}
        <div class="dbg-line dbg-line--{log.level}" title={log.source}>
          <span class="dbg-ts">{log.time}</span>
          <span class="dbg-level">{levelLabel(log.level)}</span>
          <span class="dbg-msg">{log.message}</span>
        </div>
      {/each}
    {/if}
  </div>
</div>
