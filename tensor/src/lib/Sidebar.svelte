<script>
  import { cells, navigateToCellLine } from './stores.js';
  import { collectFalconSymbols } from './source-symbols.js';

  $: symbols = collectFalconSymbols($cells);
  $: variableSymbols = symbols.filter(symbol => symbol.scope === 'global');
  $: functionSymbols = symbols.filter(symbol => symbol.scope === 'func');

  // All paths inherit fill="none" stroke="currentColor" stroke-width="1.35" from the SVG wrapper.
  // Override fill/stroke inline where needed.
  const VAR_ICONS = {
    // T shape — typography/text
    text:   `<path d="M2.5 3.5h7M6 3.5v5.5" stroke-linecap="round"/>`,
    // three bullet-rows
    list:   `<circle cx="2.8" cy="3.8" r="0.85" fill="currentColor" stroke="none"/>
             <circle cx="2.8" cy="6.5" r="0.85" fill="currentColor" stroke="none"/>
             <circle cx="2.8" cy="9.2" r="0.85" fill="currentColor" stroke="none"/>
             <path d="M5 3.8h5.5M5 6.5h4.5M5 9.2h3.5" stroke-linecap="round"/>`,
    // open book — two pages fanning from centre spine
    map:    `<path d="M6 3.5C4.5 2.5 3 2.5 2 3V9.5C3 9 4.5 9 6 10C7.5 9 9 9 10 9.5V3C9 2.5 7.5 2.5 6 3.5Z" stroke-linejoin="round"/>
             <path d="M6 3.5V10" stroke-linecap="round"/>`,
    // half-filled circle — left half = false/0, right half = true/1
    bool:   `<path d="M6 2a4 4 0 010 8Z" fill="currentColor" stroke="none"/>
             <circle cx="6" cy="6" r="4"/>`,
    // π — mathematical constant, universal number indicator
    number: `<path d="M2.5 3.5h7M4.5 3.5v6M8 3.5v5c0 1 0.5 1.5 1.5 1.5" stroke-linecap="round" stroke-linejoin="round"/>`,
    // raindrop / color swatch
    color:  `<path d="M6 1.5C4 4 2.5 5.8 2.5 8a3.5 3.5 0 007 0C9.5 5.8 8 4 6 1.5Z"
                   fill="currentColor" fill-opacity="0.3" stroke-linejoin="round"/>`,
    // three-way asterisk — wildcard / any type
    var:    `<path d="M6 2v8M2.5 4L9.5 8M9.5 4L2.5 8" stroke-linecap="round"/>`,
  };

  const FN_ICONS = {
    // right arrow → produces a value
    returning: `<path d="M2.5 6h7M7 3.5L9.5 6 7 8.5" stroke-linecap="round" stroke-linejoin="round"/>`,
    // ▽ downward triangle — funnel, no output
    void:      `<path d="M2 3h8L6 9.5Z" stroke-linejoin="round"/>`,
  };
</script>

<div id="sidebar">
  <div class="sidebar-section">
    <div class="sidebar-label">
      <!-- curly braces { } — universal variable/binding symbol -->
      <svg class="sidebar-label-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
        <path d="M6.5 3C5.5 3 5 3.6 5 4.6V7C5 7.8 4.4 8 4 8C4.4 8 5 8.2 5 9V11.4C5 12.4 5.5 13 6.5 13"/>
        <path d="M9.5 3C10.5 3 11 3.6 11 4.6V7C11 7.8 11.6 8 12 8C11.6 8 11 8.2 11 9V11.4C11 12.4 10.5 13 9.5 13"/>
      </svg>
      Global Variables
    </div>
    {#if variableSymbols.length}
      {#each variableSymbols as symbol (symbol.id)}
        <button
          type="button"
          class="sidebar-item sidebar-item--var"
          title={`${symbol.name} (${symbol.kind}) — cell ${symbol.cellId}, line ${symbol.line}`}
          on:click={() => navigateToCellLine(symbol.cellId, symbol.line)}
        >
          <span class="sidebar-symbol-name">{symbol.name}</span>
          <svg class="sidebar-kind-icon" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.35">
            {@html VAR_ICONS[symbol.kind] ?? VAR_ICONS.var}
          </svg>
        </button>
      {/each}
    {:else}
      <div class="sidebar-empty">No variables declared</div>
    {/if}
  </div>

  <div class="sidebar-section">
    <div class="sidebar-label">
      <!-- ƒ — mathematical function symbol -->
      <svg class="sidebar-label-icon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
        <path d="M11 3.5C11 2.2 9 1.8 8 2.5C7.2 3 7 4 7 5L6.5 13"/>
        <path d="M4.5 8H9"/>
      </svg>
      Procedures
    </div>
    {#if functionSymbols.length}
      {#each functionSymbols as symbol (symbol.id)}
        <button
          type="button"
          class="sidebar-item sidebar-item--fn"
          title={`${symbol.name} (${symbol.kind}) — cell ${symbol.cellId}, line ${symbol.line}`}
          on:click={() => navigateToCellLine(symbol.cellId, symbol.line)}
        >
          <span class="sidebar-symbol-name">{symbol.name}</span>
          <svg class="sidebar-kind-icon" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5">
            {@html FN_ICONS[symbol.kind] ?? FN_ICONS.void}
          </svg>
        </button>
      {/each}
    {:else}
      <div class="sidebar-empty">No procedures declared</div>
    {/if}
  </div>

</div>
