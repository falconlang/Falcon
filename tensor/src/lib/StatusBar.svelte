<script>
  import { activeCellId, cells, lastRunAt, liveTestOpen, liveTestState } from './stores.js';

  const COMPANION_LABELS = {
    idle: 'No companion',
    polling: 'Awaiting companion',
    connecting: 'Connecting',
    connected: 'Companion connected',
    error: 'Companion error',
  };

  $: companionStatus = $liveTestState.status || 'idle';
  $: companionLabel = COMPANION_LABELS[companionStatus] || COMPANION_LABELS.idle;
  $: lastRunLabel = $lastRunAt
    ? `Last run: ${new Date($lastRunAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
    : 'Not run yet';

  function openCompanionOverlay() {
    liveTestOpen.set(true);
  }
</script>

<div id="statusbar">
  <div class="status-item">
    <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.3"><circle cx="6" cy="6" r="4"/><path d="M6 4v2l1 1"/></svg>
    {lastRunLabel}
  </div>
  <span class="status-sep">·</span>
  <div class="status-item">{$activeCellId ? `Active: ${$activeCellId}` : 'No active cell'}</div>
  <span class="status-sep">·</span>
  <div class="status-item">UTF-8</div>
  <span class="status-sep">·</span>
  <div class="status-item">{$cells.length} cells</div>
  <div class="tb-spacer"></div>
  <button
    type="button"
    class="status-item status-companion status-companion--{companionStatus}"
    on:click={openCompanionOverlay}
    aria-label="Open Companion status"
    title="Open Companion status"
  >
    <svg viewBox="0 0 8 8" fill="currentColor"><circle cx="4" cy="4" r="3"/></svg>
    {companionLabel}
  </button>
</div>
