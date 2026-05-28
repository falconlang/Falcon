<script>
  import { addCodeCell, sidebarVisible, debugCollapsed } from './stores.js';

  function toggleSidebar() {
    sidebarVisible.update(v => !v);
  }

  function toggleDebugPanel() {
    debugCollapsed.update(v => !v);
  }

  $: debugActive = !$debugCollapsed;
</script>

<div id="toolbar">
  <div class="add-btns">
    <button class="add-btn" on:click={addCodeCell}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 1v12M1 7h12"/></svg>
      Code
    </button>
  </div>
  <div class="tl-sep"></div>
  <button class="tl-btn" title="Clear outputs">
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4h10M5 4V2h4v2M11 4l-.8 8H3.8L3 4"/></svg>
    Clear outputs
  </button>
  <div class="tl-sep"></div>
  <button
    class="tl-btn"
    class:active={debugActive}
    id="debug-toolbar-btn"
    title="Toggle debug panel"
    on:click={toggleDebugPanel}
  >
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="7" cy="5" r="2.5"/><path d="M4 13c0-1.657 1.343-3 3-3s3 1.343 3 3"/><path d="M1 5h2M11 5h2M7 1v1M7 8v1"/><circle cx="7" cy="5" r="1" fill="currentColor" stroke="none"/></svg>
    Debug
  </button>
  <div class="tb-spacer"></div>
  <button class="tl-btn" title="Search">
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="6" cy="6" r="4"/><path d="M10 10l2.5 2.5"/></svg>
  </button>
  <button class="tl-btn" title="Toggle sidebar" on:click={toggleSidebar}>
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1" y="1" width="12" height="12" rx="1.5"/><path d="M5 1v12"/></svg>
  </button>
</div>
