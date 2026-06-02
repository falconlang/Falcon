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

    const narrowMq = window.matchMedia('(max-width: 1024px)');
    const onMqChange = (e) => { sidebarVisible.set(!e.matches); };
    narrowMq.addEventListener('change', onMqChange);

    const onKeyDown = e => {
      if (e.key === 'Escape') {
        hideCtx();
        liveTestOpen.set(false);
      }
    };

    const onSelectionChange = () => {
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
    };

    document.addEventListener('click', hideCtx);
    document.addEventListener('keydown', onKeyDown);
    document.addEventListener('selectionchange', onSelectionChange);

    return () => {
      narrowMq.removeEventListener('change', onMqChange);
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
    <div id="notebook-debug-col">
      <div id="notebook-col">
        <Notebook />
      </div>
      <DebugPanel />
    </div>
    <DesignerPanel />
    <LiveTestOverlay />
  </div>

  <StatusBar />
</div>

<ContextMenu />
