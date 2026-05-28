<script>
  import { setActive, moveCellById, deleteCellById } from './stores.js';

  export let cell;
  export let active = false;
</script>

<div
  class="cell md-cell"
  class:active
  id="cell-{cell.id}"
  on:click={() => setActive(cell.id)}
>
  <div class="cell-gutter">
    <button class="gutter-btn" title="Move up" on:click|stopPropagation={() => moveCellById(cell.id, -1)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 8l4-4 4 4"/></svg>
    </button>
    <button class="gutter-btn" title="Move down" on:click|stopPropagation={() => moveCellById(cell.id, 1)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4l4 4 4-4"/></svg>
    </button>
    <button class="gutter-btn" title="Delete" on:click|stopPropagation={() => deleteCellById(cell.id)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 3h8M5 3V2h2v1M9 3l-.6 6H3.6L3 3"/></svg>
    </button>
  </div>
  <div class="md-content">{@html cell.content}</div>
</div>
