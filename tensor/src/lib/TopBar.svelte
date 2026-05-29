<script>
  import { onMount } from 'svelte';
  import { liveTestOpen, liveTestState, loadProjectState, projectName } from './stores.js';
  import {
    downloadBlob,
    exportCurrentProjectToAia,
    importAiaFile,
    sanitizeProjectName,
  } from './appinventor-aia.js';

  const LIVE_TEST_LABELS = { idle: 'Test', polling: 'Waiting...', connecting: 'Connecting...', connected: 'Connected', error: 'Error' };
  const LIVE_TEST_TITLES = { idle: 'Connect Companion', polling: 'Waiting for Companion', connecting: 'Negotiating connection', connected: 'Companion connected', error: 'Companion error' };

  $: liveTestStatus = $liveTestState.status || 'idle';
  $: liveTestLabel = LIVE_TEST_LABELS[liveTestStatus] || LIVE_TEST_LABELS.idle;
  $: liveTestTitle = LIVE_TEST_TITLES[liveTestStatus] || LIVE_TEST_TITLES.idle;

  function openLiveTest() {
    liveTestOpen.set(true);
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
  let fileMenuStatus = '';
  let fileMenuError = '';

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
    positionFileMenu();
    fileMenuOpen = true;
  }

  function openImportPicker() {
    fileMenuError = '';
    fileMenuStatus = '';
    fileInput?.click();
  }

  async function handleImportFile(e) {
    const file = e.currentTarget.files?.[0];
    e.currentTarget.value = '';
    if (!file) return;

    fileBusy = true;
    fileMenuOpen = true;
    fileMenuError = '';
    fileMenuStatus = 'Importing...';

    try {
      const project = await importAiaFile(file);
      loadProjectState(project);
      fileMenuStatus = `Loaded ${project.projectName}`;
    } catch (error) {
      fileMenuError = error?.message || String(error);
      fileMenuStatus = '';
      console.error('[aia-import]', error);
    } finally {
      fileBusy = false;
    }
  }

  async function exportProject() {
    fileBusy = true;
    fileMenuOpen = true;
    fileMenuError = '';
    fileMenuStatus = 'Exporting...';

    try {
      const blob = await exportCurrentProjectToAia();
      const safeName = sanitizeProjectName($projectName);
      downloadBlob(blob, `${safeName}.aia`);
      fileMenuStatus = `Exported ${safeName}.aia`;
    } catch (error) {
      fileMenuError = error?.message || String(error);
      fileMenuStatus = '';
      console.error('[aia-export]', error);
    } finally {
      fileBusy = false;
    }
  }

  onMount(() => {
    function onDocClick() {
      if (fileMenuOpen) fileMenuOpen = false;
    }
    function onResize() {
      if (fileMenuOpen) positionFileMenu();
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
  <a class="logo" href="#">
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
      class="tb-btn live-test-chip live-test-chip--{liveTestStatus}"
      title={liveTestTitle}
      aria-label={liveTestTitle}
      on:click={openLiveTest}
    >
      {liveTestLabel}
    </button>
  </div>
  <div class="tb-spacer"></div>
  <div class="topbar-actions">
    <button class="tb-btn" title="Share">
      <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 3l3 3-3 3M13 6H6a3 3 0 000 6h1"/></svg>
      Share
    </button>
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
    on:click|stopPropagation
  >
    <button
      class="ctx-item"
      role="menuitem"
      disabled={fileBusy}
      on:click={openImportPicker}
    >
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 2v7M4.5 6.5L7 9l2.5-2.5"/><path d="M2.5 12h9"/></svg>
      Import project (.aia) from my computer
    </button>
    <button
      class="ctx-item"
      role="menuitem"
      disabled={fileBusy}
      on:click={exportProject}
    >
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 12V5M4.5 7.5L7 5l2.5 2.5"/><path d="M2.5 12h9"/></svg>
      Export selected project (.aia) to my computer
    </button>
    {#if fileMenuStatus || fileMenuError}
      <div class="ctx-sep"></div>
      <div class:error={fileMenuError} class="file-menu-status">
        {fileMenuError || fileMenuStatus}
      </div>
    {/if}
  </div>
{/if}
