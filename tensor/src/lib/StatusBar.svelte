<script>
  import { cells } from './stores.js';
</script>

<div id="statusbar">
  <div class="status-item">
    <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.3"><circle cx="6" cy="6" r="4"/><path d="M6 4v2l1 1"/></svg>
    Last run: 2m ago
  </div>
  <span class="status-sep">·</span>
  <div class="status-item">Ln 1, Col 1</div>
  <span class="status-sep">·</span>
  <div class="status-item">UTF-8</div>
  <span class="status-sep">·</span>
  <div class="status-item">{$cells.length} cells</div>
  <div class="tb-spacer"></div>
  <div class="status-item">
    <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.3"><path d="M2 9l3-3 2 2 3-4"/></svg>
    Memory: 248 MB
  </div>
</div>
