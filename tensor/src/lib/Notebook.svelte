<script>
  import { cells, activeCellId } from './stores.js';
  import CodeCell from './CodeCell.svelte';
  import MarkdownCell from './MarkdownCell.svelte';
</script>

<div id="notebook-wrap">
  <div id="notebook">
    {#each $cells as cell (cell.id)}
      {#if cell.type === 'code'}
        <CodeCell {cell} active={$activeCellId === cell.id} />
      {:else if cell.type === 'markdown'}
        <MarkdownCell {cell} active={$activeCellId === cell.id} />
      {/if}
    {/each}
  </div>
</div>
