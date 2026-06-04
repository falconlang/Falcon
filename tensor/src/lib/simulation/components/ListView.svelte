<script>
  import { createEventDispatcher } from 'svelte';
  import { emitInteraction } from '../events.js';
  import { listViewRows, textIncludes } from '../items.js';
  import { assetName, baseStyle, boolValue, colorValue, firstNonEmpty, hasValue, numberOr, styleContext, typefaceStyleFor } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;

  const dispatch = createEventDispatcher();
  let listFilter = '';

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: listRows = listViewRows(props);
  $: filteredListRows = listRows.filter(row => textIncludes(`${row.text1} ${row.text2}`, listFilter));

  function listViewDetailStyle() {
    return [
      `color: ${colorValue(firstNonEmpty(props.TextColorDetail, props.TextColor), '#202124')};`,
      hasValue(props.FontSizeDetail) ? `font-size: ${numberOr(props.FontSizeDetail, 14)}px;` : '',
      typefaceStyleFor(props.FontTypefaceDetail),
    ].filter(Boolean).join(' ');
  }

  function listViewImageStyle() {
    return `width: ${numberOr(props.ImageWidth, 200)}px; height: ${numberOr(props.ImageHeight, 200)}px; max-width: 100%;`;
  }

  function listViewItemStyle(selected = false) {
    const bg = selected
      ? colorValue(props.SelectionColor, '#d3d3d3')
      : colorValue(firstNonEmpty(props.ElementColor, props.ItemBackgroundColor), '#ffffff');
    return [
      `color: ${colorValue(firstNonEmpty(props.TextColor, props.ItemTextColor), '#202124')};`,
      `background: ${bg};`,
      hasValue(props.DividerColor) ? `border-bottom-color: ${colorValue(props.DividerColor, '#eef1f4')};` : '',
      `border-bottom-width: ${Math.max(0, numberOr(props.DividerThickness, 0))}px;`,
      'border-bottom-style: solid;',
      `border-radius: ${Math.max(0, numberOr(props.ElementCornerRadius, 0))}px;`,
      `margin: ${Math.max(0, numberOr(props.ElementMarginsWidth, 0))}px;`,
    ].filter(Boolean).join(' ');
  }

  function rowImageUrl(row) {
    if (!row?.image || row.image === 'None') return '';
    return resolveAssetUrl(assets, row.image);
  }

  function listViewPick(index) {
    const row = listRows[index];
    if (!row) return;
    emitInteraction(
      dispatch,
      [
        { component: node.name, property: 'Selection', value: row.text1 },
        { component: node.name, property: 'SelectionIndex', value: index + 1 },
      ],
      { component: node.name, event: 'AfterPicking', args: [] },
    );
  }
</script>

<div class="sim-listview" class:sim-unsupported={unsupportedHere} style={baseStyle(context)} data-sim-component={node.name}>
  {#if boolValue(props.ShowFilterBar, false)}
    <input class="sim-list-filter" value={listFilter} placeholder={props.HintText ?? 'Search list...'} on:input={e => listFilter = e.currentTarget.value} />
  {/if}
  <div class="sim-listview-items" class:horizontal={numberOr(props.Orientation, 1) === 0}>
    {#each filteredListRows as row (row.index)}
      <button
        type="button"
        class:selected={numberOr(props.SelectionIndex, 0) === row.index + 1}
        class:with-image={!!rowImageUrl(row)}
        style={listViewItemStyle(numberOr(props.SelectionIndex, 0) === row.index + 1)}
        on:click={() => listViewPick(row.index)}
      >
        {#if rowImageUrl(row)}
          <img src={rowImageUrl(row)} alt="" style={listViewImageStyle()} />
        {/if}
        <span class="sim-list-text">
          <span>{row.text1}</span>
          {#if row.text2}
            <small style={listViewDetailStyle()}>{row.text2}</small>
          {/if}
        </span>
      </button>
    {:else}
      <div class="sim-picker-empty"></div>
    {/each}
  </div>
</div>
