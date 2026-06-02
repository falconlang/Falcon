<script>
  import { cells, activeCellId, addCodeCell, notebookMode, unifiedSelectionActive, debugExecutionState, companionCommand } from './stores.js';
  import CodeCell from './CodeCell.svelte';
  import MarkdownCell from './MarkdownCell.svelte';
  import UnifiedEditor from './UnifiedEditor.svelte';

  let unifiedEditor;
  let modeSwitching = false;

  $: isUnified = $notebookMode === 'unified';

  async function showCells() {
    if (modeSwitching) return;
    if (!isUnified) {
      notebookMode.set('cells');
      return;
    }
    modeSwitching = true;
    try {
      const committed = await unifiedEditor?.commitToCells?.();
      if (committed !== false) notebookMode.set('cells');
    } finally {
      modeSwitching = false;
    }
  }

  function showScript() {
    if (!modeSwitching) notebookMode.set('unified');
  }

  function runUnifiedDoItMouseDown(e) {
    e.preventDefault();
    e.stopPropagation();
    unifiedEditor?.runDoIt();
  }

  function runUnifiedDoItClick(e) {
    e.stopPropagation();
    if (e.detail === 0) unifiedEditor?.runDoIt();
  }

  $: isPaused = $debugExecutionState.status === 'paused';

  function continueDebug() {
    companionCommand.set({ type: 'debug-continue', hitId: $debugExecutionState.hitId });
  }
</script>

<div class="dsgn-header">
  <span class="dsgn-filename">Functionality</span>
  {#if isPaused}
    <button
      type="button"
      class="dsgn-continue-btn"
      title="Continue execution"
      on:click={continueDebug}
    >
      <svg viewBox="0 0 14 14" fill="currentColor"><path d="M2 2l6 5-6 5V2z"/><path d="M8 2l5 5-5 5V2z"/></svg>
      Continue
    </button>
  {/if}
  {#if isUnified && $unifiedSelectionActive}
    <button
      type="button"
      class="dsgn-do-it-btn"
      title="Do it"
      on:mousedown={runUnifiedDoItMouseDown}
      on:click={runUnifiedDoItClick}
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M2 6h8M7 3l3 3-3 3"/></svg>
      Do it
    </button>
  {/if}
  <div class="dsgn-mode-toggle">
    <button
      class="dsgn-mode-btn"
      class:dsgn-mode-btn--active={!isUnified}
      disabled={modeSwitching}
      on:click={showCells}
      title="Cells view"
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><rect x="1" y="1" width="10" height="3.5" rx="0.8"/><rect x="1" y="7.5" width="10" height="3.5" rx="0.8"/></svg>
      Cells
    </button>
    <button
      class="dsgn-mode-btn"
      class:dsgn-mode-btn--active={isUnified}
      disabled={modeSwitching}
      on:click={showScript}
      title="Script view"
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><path d="M1 2h10M1 5h7M1 8h10M1 11h4"/></svg>
      Script
    </button>
  </div>
</div>

<div id="notebook-wrap">
  {#if $notebookMode === 'unified'}
    <UnifiedEditor bind:this={unifiedEditor} />
  {:else}
    <div id="notebook">
      {#each $cells as cell (cell.id)}
        {#if cell.type === 'code'}
          <CodeCell {cell} active={$activeCellId === cell.id} />
        {:else if cell.type === 'markdown'}
          <MarkdownCell {cell} active={$activeCellId === cell.id} />
        {/if}
      {/each}
      {#if $cells.length === 0}
        <div class="notebook-empty">
          <p class="notebook-empty-label">Empty screen</p>
          <button class="notebook-empty-btn" on:click={addCodeCell}>
            <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 1v12M1 7h12"/></svg>
            Add code cell
          </button>
        </div>
      {/if}
    </div>
  {/if}
</div>
