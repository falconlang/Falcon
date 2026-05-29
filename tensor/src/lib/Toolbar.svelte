<script>
  import { onMount } from 'svelte';
  import {
    addCodeCell,
    sidebarVisible,
    debugCollapsed,
    debugModeEnabled,
    companionCommand,
    screenList,
    activeScreen,
    switchScreen,
    addScreen,
    removeScreen,
    projectName,
    clearDebugLogs,
  } from './stores.js';
  import ProjectPropertiesDialog from './ProjectPropertiesDialog.svelte';

  let projectPropsOpen = false;

  function toggleSidebar() { sidebarVisible.update(v => !v); }
  function toggleDebugPanel() { debugCollapsed.update(v => !v); }
  function toggleDebugMode() { companionCommand.set({ type: 'toggle-debug' }); }
  $: debugActive = !$debugCollapsed;
  $: debugModeActive = $debugModeEnabled;
  $: canRemove = $activeScreen !== 'Screen1';

  // ── Custom spinner — position:fixed, same pattern as ctx-menu ──
  let spinnerOpen = false;
  let dropdownX = 0;
  let dropdownY = 0;
  let triggerBtn;

  function openSpinner() {
    const rect = triggerBtn.getBoundingClientRect();
    dropdownX = rect.left;
    dropdownY = rect.bottom + 4;
    spinnerOpen = true;
  }

  function closeSpinner() { spinnerOpen = false; }

  function toggleSpinner(e) {
    e.stopPropagation();
    spinnerOpen ? closeSpinner() : openSpinner();
  }

  function selectScreen(name) {
    closeSpinner();
    switchScreen(name);
  }

  function handleSpinnerMenuKey(e) {
    if (e.key === 'Escape') {
      e.stopPropagation();
      closeSpinner();
      triggerBtn?.focus();
    }
  }

  onMount(() => {
    function onDocClick() { if (spinnerOpen) spinnerOpen = false; }
    document.addEventListener('click', onDocClick);
    return () => document.removeEventListener('click', onDocClick);
  });

  // ── Add Screen dialog ──
  let addDialogOpen = false;
  let addName = '';
  let addInput;

  function openAddDialog() {
    const list = $screenList;
    let n = list.length + 1;
    const existing = new Set(list);
    while (existing.has(`Screen${n}`)) n++;
    addName = `Screen${n}`;
    addDialogOpen = true;
    setTimeout(() => { addInput?.focus(); addInput?.select(); }, 50);
  }

  const SCREEN_NAME_RE = /^[a-zA-Z][a-zA-Z0-9_]*$/;

  function confirmAdd() {
    const trimmed = addName.trim();
    if (!trimmed || !SCREEN_NAME_RE.test(trimmed) || $screenList.includes(trimmed)) return;
    addScreen(trimmed);
    addDialogOpen = false;
  }

  function cancelAdd() { addDialogOpen = false; }

  function handleAddKey(e) {
    if (e.key === 'Enter') confirmAdd();
    if (e.key === 'Escape') cancelAdd();
  }

  // ── Remove Screen dialog ──
  let removeDialogOpen = false;
  let removeConfirmInput = '';
  let removeInput;

  function openRemoveDialog() {
    if (!canRemove) return;
    removeConfirmInput = '';
    removeDialogOpen = true;
    setTimeout(() => removeInput?.focus(), 50);
  }

  function confirmRemove() {
    if (removeConfirmInput.trim() !== $activeScreen) return;
    removeScreen($activeScreen);
    removeDialogOpen = false;
  }

  function cancelRemove() { removeDialogOpen = false; }

  function handleRemoveKey(e) {
    if (e.key === 'Enter') confirmRemove();
    if (e.key === 'Escape') cancelRemove();
  }

  $: removeValid = removeConfirmInput.trim() === $activeScreen;
</script>

<div id="toolbar">
  <span class="toolbar-project-title">{$projectName}</span>
  <div class="add-btns">
    <button class="add-btn" on:click={addCodeCell}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 1v12M1 7h12"/></svg>
      Code
    </button>
  </div>
  <div class="tl-sep"></div>
  <button class="tl-btn" title="Clear outputs" on:click={clearDebugLogs}>
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4h10M5 4V2h4v2M11 4l-.8 8H3.8L3 4"/></svg>
    Clear outputs
  </button>
  <div class="tl-sep"></div>
  <button
    class="tl-btn"
    class:active={debugActive}
    title="Toggle logger panel"
    on:click={toggleDebugPanel}
  >
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1.5" y="2.5" width="11" height="9" rx="1.5"/><path d="M4 6.5l2 1.5-2 1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M8.5 9.5h2" stroke-linecap="round"/></svg>
    Logger
  </button>
  <div class="tl-sep"></div>
  <button
    class="tl-btn"
    class:active={debugModeActive}
    id="debug-toolbar-btn"
    title="Debug"
    aria-pressed={debugModeActive}
    on:click={toggleDebugMode}
  >
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="7" cy="5" r="2.5"/><path d="M4 13c0-1.657 1.343-3 3-3s3 1.343 3 3"/><path d="M1 5h2M11 5h2M7 1v1M7 8v1"/><circle cx="7" cy="5" r="1" fill="currentColor" stroke="none"/></svg>
    Debug
  </button>

  <div class="tb-spacer"></div>

  <!-- Screen selector cluster, absolutely centered -->
  <div class="screen-selector">

    <!-- Spinner trigger -->
    <button
      class="screen-spinner-btn"
      class:open={spinnerOpen}
      bind:this={triggerBtn}
      on:click={toggleSpinner}
      aria-haspopup="listbox"
      aria-expanded={spinnerOpen}
      title="Select screen"
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" class="screen-spinner-icon"><rect x="2.5" y="0.5" width="7" height="11" rx="1.5" stroke-width="1.3"/><line x1="4" y1="9.5" x2="8" y2="9.5" stroke-width="1.3" stroke-linecap="round"/></svg>
      <span class="screen-spinner-label">{$activeScreen}</span>
      <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.5" class="screen-spinner-chevron" class:rotated={spinnerOpen}><path d="M2 3.5l3 3 3-3"/></svg>
    </button>
    <button class="screen-spinner-btn" title="Add Screen" on:click={openAddDialog}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" class="screen-spinner-icon"><path d="M7 2v10M2 7h10"/></svg>
      Add Screen
    </button>
    <button class="screen-spinner-btn" title="Remove Screen" disabled={!canRemove} on:click={openRemoveDialog}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" class="screen-spinner-icon"><circle cx="7" cy="7" r="5"/><path d="M4.8 4.8l4.4 4.4M9.2 4.8l-4.4 4.4" stroke-linecap="round"/></svg>
      Remove Screen
    </button>
    <button class="screen-spinner-btn" title="Project Properties" on:click={() => projectPropsOpen = true}>
      <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" class="screen-spinner-icon"><circle cx="4.5" cy="4.5" r="1.5"/><path d="M1 4.5h1.5M6 4.5H13" stroke-linecap="round"/><circle cx="9.5" cy="9.5" r="1.5"/><path d="M1 9.5h7M11 9.5H13" stroke-linecap="round"/></svg>
      Project Properties
    </button>

  </div>

  <div class="tb-spacer"></div>
  <button class="tl-btn" title="Search">
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="6" cy="6" r="4"/><path d="M10 10l2.5 2.5"/></svg>
  </button>
  <button class="tl-btn" title="Toggle sidebar" on:click={toggleSidebar}>
    <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="1" y="1" width="12" height="12" rx="1.5"/><path d="M5 1v12"/></svg>
  </button>
</div>

<!-- Spinner dropdown — position:fixed, same layer as ctx-menu -->
<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
{#if spinnerOpen}
  <div
    class="ctx-menu screen-spinner-menu show"
    style="left:{dropdownX}px; top:{dropdownY}px"
    role="listbox"
    tabindex="-1"
    on:click|stopPropagation
    on:keydown={handleSpinnerMenuKey}
  >
    {#each $screenList as name}
      <button
        type="button"
        class="ctx-item"
        class:screen-spinner-active={name === $activeScreen}
        role="option"
        aria-selected={name === $activeScreen}
        tabindex={name === $activeScreen ? 0 : -1}
        on:click={() => selectScreen(name)}
      >
        {name}
      </button>
    {/each}
  </div>
{/if}

<!-- ── Add Screen dialog ── -->
<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
{#if addDialogOpen}
  <div class="sd-overlay" on:click|self={cancelAdd}>
    <div class="sd-card" role="dialog" aria-modal="true" aria-labelledby="sd-add-title">
      <div class="sd-header">
        <div class="sd-header-icon">
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.4"><rect x="1" y="2" width="12" height="9" rx="1"/><path d="M5 11v1.5M9 11v1.5M3.5 12.5h7"/></svg>
        </div>
        <span class="sd-title" id="sd-add-title">New Screen</span>
        <button class="sd-close" on:click={cancelAdd} title="Cancel">
          <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M2 2l6 6M8 2l-6 6"/></svg>
        </button>
      </div>
      <div class="sd-body">
        <p class="sd-desc">Choose a name for the new screen.</p>
        <input
          class="sd-input"
          type="text"
          placeholder="e.g. Screen2"
          bind:value={addName}
          bind:this={addInput}
          on:keydown={handleAddKey}
          spellcheck="false"
          maxlength="64"
        />
        {#if addName.trim() && !SCREEN_NAME_RE.test(addName.trim())}
          <p class="sd-error">Must start with a letter and contain only letters, numbers, or underscores.</p>
        {:else if addName.trim() && $screenList.includes(addName.trim())}
          <p class="sd-error">A screen named "{addName.trim()}" already exists.</p>
        {/if}
      </div>
      <div class="sd-footer">
        <button class="sd-btn sd-btn--ghost" on:click={cancelAdd}>Cancel</button>
        <button
          class="sd-btn sd-btn--accent"
          on:click={confirmAdd}
          disabled={!addName.trim() || !SCREEN_NAME_RE.test(addName.trim()) || $screenList.includes(addName.trim())}
        >
          <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M6 1v10M1 6h10"/></svg>
          Add Screen
        </button>
      </div>
    </div>
  </div>
{/if}

<!-- ── Remove Screen dialog ── -->
<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
{#if removeDialogOpen}
  <div class="sd-overlay" on:click|self={cancelRemove}>
    <div class="sd-card" role="dialog" aria-modal="true" aria-labelledby="sd-remove-title">
      <div class="sd-header">
        <div class="sd-header-icon sd-header-icon--danger">
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.4"><path d="M2 4h10M5 4V2h4v2M11 4l-.8 8H3.8L3 4"/></svg>
        </div>
        <span class="sd-title" id="sd-remove-title">Remove Screen</span>
        <button class="sd-close" on:click={cancelRemove} title="Cancel">
          <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M2 2l6 6M8 2l-6 6"/></svg>
        </button>
      </div>
      <div class="sd-body">
        <p class="sd-desc">This will permanently delete <strong>{$activeScreen}</strong> and all its content. Type the screen name to confirm.</p>
        <input
          class="sd-input"
          class:valid={removeValid}
          type="text"
          placeholder={$activeScreen}
          bind:value={removeConfirmInput}
          bind:this={removeInput}
          on:keydown={handleRemoveKey}
          spellcheck="false"
          maxlength="64"
        />
      </div>
      <div class="sd-footer">
        <button class="sd-btn sd-btn--ghost" on:click={cancelRemove}>Cancel</button>
        <button
          class="sd-btn sd-btn--danger"
          on:click={confirmRemove}
          disabled={!removeValid}
        >
          <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M1.5 3h9M4 3V1.5h4V3M9.5 3l-.6 7H3.1L2.5 3"/></svg>
          Remove Screen
        </button>
      </div>
    </div>
  </div>
{/if}

<ProjectPropertiesDialog bind:open={projectPropsOpen} />
