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

  onMount(() => {
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
    {/if}
    <Notebook />
    <DesignerPanel />
    <LiveTestOverlay />
  </div>

  <DebugPanel />
  <StatusBar />
</div>

<ContextMenu />
