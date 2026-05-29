<script>
  import { onMount } from 'svelte';
  import { hideCtx, liveTestOpen, doItCellId, sidebarVisible } from './lib/stores.js';
  import TopBar from './lib/TopBar.svelte';
  import Toolbar from './lib/Toolbar.svelte';
  import Sidebar from './lib/Sidebar.svelte';
  import Notebook from './lib/Notebook.svelte';
  import DesignerPanel from './lib/DesignerPanel.svelte';
  import DebugPanel from './lib/DebugPanel.svelte';
  import StatusBar from './lib/StatusBar.svelte';
  import ContextMenu from './lib/ContextMenu.svelte';
  import LiveTestOverlay from './lib/LiveTestOverlay.svelte';
  import { warmBlocklyPreviewRuntime } from './lib/blockly-preview.js';

  function closeSidebar() { sidebarVisible.set(false); }

  onMount(() => {
    const warmBlockly = () => warmBlocklyPreviewRuntime();
    let idleId = null;
    let warmTimer = null;
    if (window.requestIdleCallback) {
      idleId = window.requestIdleCallback(warmBlockly, { timeout: 2500 });
    } else {
      warmTimer = window.setTimeout(warmBlockly, 800);
    }

    const onKeyDown = e => {
      if (e.key === 'Escape') {
        hideCtx();
        liveTestOpen.set(false);
      }
    };

    // Selection tracking is now handled per-cell in CodeMirror's updateListener.
    const onSelectionChange = () => {};

    document.addEventListener('click', hideCtx);
    document.addEventListener('keydown', onKeyDown);
    document.addEventListener('selectionchange', onSelectionChange);

    return () => {
      document.removeEventListener('click', hideCtx);
      document.removeEventListener('keydown', onKeyDown);
      document.removeEventListener('selectionchange', onSelectionChange);
      if (idleId !== null && window.cancelIdleCallback) window.cancelIdleCallback(idleId);
      if (warmTimer !== null) window.clearTimeout(warmTimer);
    };
  });
</script>

<div id="app">
  <TopBar />
  <Toolbar />

  <div id="main" style="position: relative">
    {#if $sidebarVisible}
      <Sidebar />
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div class="sidebar-backdrop" on:click={closeSidebar}></div>
    {/if}
    <div id="notebook-col">
      <Notebook />
      <DebugPanel />
    </div>
    <DesignerPanel />
    <LiveTestOverlay />
  </div>

  <StatusBar />
</div>

<ContextMenu />
