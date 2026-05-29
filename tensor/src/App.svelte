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
    if (window.requestIdleCallback) {
      window.requestIdleCallback(warmBlockly, { timeout: 2500 });
    } else {
      window.setTimeout(warmBlockly, 800);
    }

    document.addEventListener('click', hideCtx);

    document.addEventListener('keydown', e => {
      if (e.key === 'Escape') {
        hideCtx();
        liveTestOpen.set(false);
      }
    });

    document.addEventListener('selectionchange', () => {
      const activeEl = document.activeElement;
      if (
        activeEl?.classList?.contains('code-area')
        && activeEl.id?.startsWith('code-')
      ) {
        doItCellId.set(activeEl.selectionStart !== activeEl.selectionEnd ? activeEl.id.slice(5) : null);
        return;
      }

      const sel = window.getSelection();
      const hasText = sel && sel.toString().length > 0;
      let selCellId = null;
      if (hasText && sel.rangeCount) {
        const node = sel.getRangeAt(0).commonAncestorContainer;
        const area = node.nodeType === 1
          ? node.closest('.code-area')
          : node.parentElement?.closest('.code-area');
        if (area && area.id.startsWith('code-')) selCellId = area.id.slice(5);
      }
      doItCellId.set(selCellId);
    });
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
