<script>
  import { onMount } from 'svelte';
  import { liveTestOpen, simulateOpen, liveTestState, companionCommand, loadProjectState, projectName } from './stores.js';
  import {
    downloadBlob,
    exportCurrentProjectToAia,
    importAiaFile,
    sanitizeProjectName,
  } from './appinventor-aia.js';
  $: liveTestStatus = $liveTestState.status || 'idle';
  $: isConnected = liveTestStatus === 'connected';
  $: isActive = liveTestStatus !== 'idle';

  let testMenuOpen = false;
  let testMenuX = 0;
  let testMenuY = 0;
  let testMenuBtn;

  function positionTestMenu() {
    const rect = testMenuBtn?.getBoundingClientRect();
    if (!rect) return;
    testMenuX = rect.left;
    testMenuY = rect.bottom + 4;
  }

  function toggleTestMenu(e) {
    e.stopPropagation();
    if (testMenuOpen) { testMenuOpen = false; return; }
    fileMenuOpen = false;
    positionTestMenu();
    testMenuOpen = true;
  }

  function testAction(action) {
    testMenuOpen = false;
    if (action === 'connect') { liveTestOpen.set(true); return; }
    if (action === 'simulate') { simulateOpen.set(true); return; }
    if (action === 'refresh') { companionCommand.set(action); return; }
    liveTestOpen.set(true);
    companionCommand.set(action);
  }

  function handleMenuKeydown(e) {
    e.stopPropagation();
    if (e.key !== 'Escape') return;
    fileMenuOpen = false;
    testMenuOpen = false;
  }

  function toggleTheme() {
    const isDark = document.documentElement.dataset.theme === 'dark';
    const next = isDark ? 'light' : 'dark';
    document.documentElement.classList.add('theme-transitioning');
    document.documentElement.dataset.theme = next;
    localStorage.setItem('tensor-theme', next);
    updateIcon();
    setTimeout(() => document.documentElement.classList.remove('theme-transitioning'), 300);
  }

  function updateIcon() {
    isDark = document.documentElement.dataset.theme === 'dark';
  }

  let isDark = document.documentElement.dataset.theme === 'dark';
  let fileMenuOpen = false;
  let fileMenuX = 0;
  let fileMenuY = 0;
  let fileMenuBtn;
  let fileInput;
  let fileBusy = false;
  let toast = null; // { message, type: 'success'|'error' }
  let toastTimer = null;

  function showToast(message, type = 'success') {
    clearTimeout(toastTimer);
    toast = { message, type };
    toastTimer = setTimeout(() => { toast = null; }, 3500);
  }

  function positionFileMenu() {
    const rect = fileMenuBtn?.getBoundingClientRect();
    if (!rect) return;
    fileMenuX = rect.left;
    fileMenuY = rect.bottom + 4;
  }

  function toggleFileMenu(e) {
    e.stopPropagation();
    if (fileMenuOpen) {
      fileMenuOpen = false;
      return;
    }
    testMenuOpen = false;
    positionFileMenu();
    fileMenuOpen = true;
  }

  function openImportPicker() {
    fileInput?.click();
  }

  async function handleImportFile(e) {
    const file = e.currentTarget.files?.[0];
    e.currentTarget.value = '';
    if (!file) return;

    fileBusy = true;
    fileMenuOpen = false;

    try {
      const project = await importAiaFile(file);
      loadProjectState(project);
      showToast(`Loaded “${project.projectName}”`);
    } catch (error) {
      showToast(error?.message || String(error), 'error');
      console.error('[aia-import]', error);
    } finally {
      fileBusy = false;
    }
  }

  async function exportProject() {
    fileBusy = true;
    fileMenuOpen = false;

    try {
      const blob = await exportCurrentProjectToAia();
      const safeName = sanitizeProjectName($projectName);
      downloadBlob(blob, `${safeName}.aia`);
      showToast(`Exported ${safeName}.aia`);
    } catch (error) {
      showToast(error?.message || String(error), 'error');
      console.error('[aia-export]', error);
    } finally {
      fileBusy = false;
    }
  }

  onMount(() => {
    function onDocClick() {
      if (fileMenuOpen) fileMenuOpen = false;
      if (testMenuOpen) testMenuOpen = false;
    }
    function onResize() {
      if (fileMenuOpen) positionFileMenu();
      if (testMenuOpen) positionTestMenu();
    }
    document.addEventListener('click', onDocClick);
    window.addEventListener('resize', onResize);
    window.addEventListener('scroll', onResize, true);
    return () => {
      document.removeEventListener('click', onDocClick);
      window.removeEventListener('resize', onResize);
      window.removeEventListener('scroll', onResize, true);
    };
  });
</script>

<div id="topbar">
  <a class="logo" href="/" on:click|preventDefault>
    <div class="logo-mark">
      <svg viewBox="0 0 14 14" fill="none">
        <rect x="1" y="1" width="5" height="5" rx="1" fill="white" fill-opacity="0.9"/>
        <rect x="8" y="1" width="5" height="5" rx="1" fill="white" fill-opacity="0.5"/>
        <rect x="1" y="8" width="5" height="5" rx="1" fill="white" fill-opacity="0.5"/>
        <rect x="8" y="8" width="5" height="5" rx="1" fill="white" fill-opacity="0.8"/>
      </svg>
    </div>
    tensor
  </a>
  <div class="divider"></div>
  <div class="topbar-primary-actions">
    <div class="file-menu-anchor">
      <button
        class="tb-btn file-menu-btn"
        class:open={fileMenuOpen}
        bind:this={fileMenuBtn}
        on:click={toggleFileMenu}
        aria-haspopup="menu"
        aria-expanded={fileMenuOpen}
      >
        File
      </button>
      <input
        class="visually-hidden-file"
        type="file"
        accept=".aia,.zip,application/zip,application/x-zip-compressed"
        bind:this={fileInput}
        on:change={handleImportFile}
      />
    </div>
    <button
      class="tb-btn file-menu-btn live-test-chip live-test-chip--{liveTestStatus}"
      class:open={testMenuOpen}
      bind:this={testMenuBtn}
      on:click={toggleTestMenu}
      aria-haspopup="menu"
      aria-expanded={testMenuOpen}
    >
      Connect
    </button>
  </div>
  <div class="tb-spacer"></div>
  <div class="topbar-actions">
    <button class="tb-btn" title={isDark ? 'Switch to light mode' : 'Switch to dark mode'} on:click={toggleTheme}>
      {#if isDark}
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="3"/><path d="M8 2v1M8 13v1M2 8h1M13 8h1M3.7 3.7l.7.7M11.6 11.6l.7.7M3.7 12.3l.7-.7M11.6 4.4l.7-.7"/></svg>
      {:else}
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
      {/if}
    </button>
  </div>
</div>

{#if fileMenuOpen}
  <div
    class="ctx-menu file-dropdown show"
    style="left:{fileMenuX}px; top:{fileMenuY}px"
    role="menu"
    tabindex="-1"
    on:click|stopPropagation
    on:keydown={handleMenuKeydown}
  >
    <button
      class="ctx-item"
      role="menuitem"
      disabled={fileBusy}
      on:click={openImportPicker}
    >
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1.5" y="8.5" width="11" height="4" rx="1"/><path d="M7 1v7" stroke-linecap="round"/><path d="M4.5 6l2.5 2.5 2.5-2.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
      Import project (.aia) from my computer
    </button>
    <button
      class="ctx-item"
      role="menuitem"
      disabled={fileBusy}
      on:click={exportProject}
    >
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1.5" y="8.5" width="11" height="4" rx="1"/><path d="M7 8V2" stroke-linecap="round"/><path d="M4.5 4L7 1.5 9.5 4" stroke-linecap="round" stroke-linejoin="round"/></svg>
      Export selected project (.aia) to my computer
    </button>
  </div>
{/if}

{#if testMenuOpen}
  <div
    class="ctx-menu file-dropdown show"
    style="left:{testMenuX}px; top:{testMenuY}px"
    role="menu"
    tabindex="-1"
    on:click|stopPropagation
    on:keydown={handleMenuKeydown}
  >
    <button class="ctx-item" role="menuitem" on:click={() => testAction('connect')}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="4" y="1" width="6" height="12" rx="1.5"/><path d="M5.5 3.5h3" stroke-linecap="round"/><path d="M5.5 10.5h3" stroke-linecap="round"/></svg>
      Live test
    </button>
    <button class="ctx-item" role="menuitem" on:click={() => testAction('simulate')}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1.5" y="2" width="11" height="10" rx="1.5"/><path d="M4 5h6M4 7.5h3" stroke-linecap="round"/><path d="M9 8.5l2 1-2 1z" fill="currentColor" stroke="none"/></svg>
      Simulate
    </button>
    <div class="ctx-sep"></div>
    <button class="ctx-item" role="menuitem" disabled={!isConnected} on:click={() => testAction('refresh')}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4.5h10M9.5 2l2.5 2.5-2.5 2.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M12 9.5H2M4.5 7l-2.5 2.5 2.5 2.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
      Refresh Companion
    </button>
    <button class="ctx-item" role="menuitem" disabled={!isActive} on:click={() => testAction('reset')}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 4H6a4 4 0 000 8h2" stroke-linecap="round"/><path d="M7 1.5L10 4 7 6.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
      Reset Connection
    </button>
    <button class="ctx-item" role="menuitem" on:click={() => testAction('hard-reset')}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 1.5L1.5 12h11z" stroke-linecap="round" stroke-linejoin="round"/><path d="M7 5.5v3" stroke-linecap="round"/><circle cx="7" cy="10.5" r=".7" fill="currentColor" stroke="none"/></svg>
      Hard Reset
    </button>
    <div class="ctx-sep"></div>
    <button class="ctx-item" role="menuitem" disabled={!isConnected} on:click={() => testAction('save')}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3.5" y="1" width="7" height="12" rx="1.5"/><path d="M7 9V5M5 7l2-2 2 2" stroke-linecap="round" stroke-linejoin="round"/></svg>
      Save Project to Companion
    </button>
  </div>
{/if}

{#if toast}
  <div class="toast toast--{toast.type}" role="status" aria-live="polite">
    {#if toast.type === 'success'}
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 7l3 3 6-6"/></svg>
    {:else}
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"><path d="M7 2v5"/><circle cx="7" cy="11" r=".8" fill="currentColor" stroke="none"/></svg>
    {/if}
    {toast.message}
  </div>
{/if}
