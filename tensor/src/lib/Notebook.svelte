<script>
  import { cells, activeCellId, addCodeCell, notebookMode } from './stores.js';
  import CodeCell from './CodeCell.svelte';
  import MarkdownCell from './MarkdownCell.svelte';
  import UnifiedEditor from './UnifiedEditor.svelte';

  $: isUnified = $notebookMode === 'unified';
  function toggleMode() { notebookMode.update(m => m === 'cells' ? 'unified' : 'cells'); }
</script>

<div class="dsgn-header">
  <span class="dsgn-filename">Functionality</span>
  <div class="dsgn-mode-toggle">
    <button
      class="dsgn-mode-btn"
      class:dsgn-mode-btn--active={!isUnified}
      on:click={() => notebookMode.set('cells')}
      title="Cells view"
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><rect x="1" y="1" width="10" height="3.5" rx="0.8"/><rect x="1" y="7.5" width="10" height="3.5" rx="0.8"/></svg>
      Cells
    </button>
    <button
      class="dsgn-mode-btn"
      class:dsgn-mode-btn--active={isUnified}
      on:click={() => notebookMode.set('unified')}
      title="Script view"
    >
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round"><path d="M1 2h10M1 5h7M1 8h10M1 11h4"/></svg>
      Script
    </button>
  </div>
</div>

<div id="notebook-wrap">
  {#if $notebookMode === 'unified'}
    <UnifiedEditor />
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
