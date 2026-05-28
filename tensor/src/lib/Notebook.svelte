<script>
  import { cells, activeCellId, addCodeCell } from './stores.js';
  import CodeCell from './CodeCell.svelte';
  import MarkdownCell from './MarkdownCell.svelte';
</script>

<div class="dsgn-header">
  <span class="dsgn-filename">Functionality</span>
</div>

<div id="notebook-wrap">
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
</div>
