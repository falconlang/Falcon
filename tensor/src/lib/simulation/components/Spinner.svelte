<script>
  import { createEventDispatcher } from 'svelte';
  import { runAction } from '../actions.js';
  import { emitEvent, emitInteraction } from '../events.js';
  import { elementItems, filterIndexed, itemMainText, selectionByIndex } from '../items.js';
  import { assetName, boolValue, buttonInnerStyle, colorValue, containerStyle, firstNonEmpty, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let enabled = true;
  export let visible = true;
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  let pickerOpen = false;
  let pickerFilter = '';
  let pickerWrapEl;
  let actionTokens = {};

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: elements = elementItems(props).map(itemMainText);
  $: filteredPickerItems = filterIndexed(elements, pickerFilter);
  $: handleActions(actions?.[node?.name] ?? {});

  function handleActions(actionState) {
    runAction(actionTokens, actionState, 'open', ['open', 'Open', 'DisplayDropdown', 'LaunchPicker'], () => {
      if (enabled && visible) {
        pickerFilter = '';
        pickerOpen = true;
      }
    });
  }

  function pickerMenuStyle() {
    return [
      `--picker-item-color: ${colorValue(firstNonEmpty(props.ItemTextColor, props.TextColor), '#202124')};`,
      `--picker-item-bg: ${colorValue(props.ItemBackgroundColor, '#ffffff')};`,
      `--picker-selection-bg: ${colorValue(props.SelectionColor, '#e8f0fe')};`,
    ].join(' ');
  }

  function dismissPicker() {
    pickerOpen = false;
    pickerFilter = '';
  }

  function onWindowPointerDown(e) {
    if (!pickerOpen) return;
    if (pickerWrapEl && pickerWrapEl.contains(e.target)) return;
    dismissPicker();
  }

  function spinnerChange(e) {
    const index = Number(e.currentTarget.value);
    if (!Number.isInteger(index) || index < 0) return;
    pickSpinnerItem(index);
  }

  function pickSpinnerItem(index) {
    const next = selectionByIndex(index, elements);
    if (!next) return;
    pickerOpen = false;
    const suppressInitial = numberOr(props.SelectionIndex, 0) === 0 && index === 0;
    emitInteraction(
      dispatch,
      [
        { component: node.name, property: 'Selection', value: next.selection },
        { component: node.name, property: 'SelectionIndex', value: next.selectionIndex },
      ],
      suppressInitial ? null : { component: node.name, event: 'AfterSelecting', args: [next.selection] },
    );
  }
</script>

<svelte:window on:pointerdown={onWindowPointerDown} />

<div bind:this={pickerWrapEl} class="sim-picker" class:sim-unsupported={unsupportedHere} style={containerStyle(props, parentType)} data-sim-component={node.name}>
  <select
    class="sim-select"
    style={buttonInnerStyle(context, 'width: 100%;')}
    disabled={!enabled}
    on:change={spinnerChange}
    on:pointerdown={() => emitEvent(dispatch, eventRunner, node.name, 'TouchDown')}
    on:pointerup={() => emitEvent(dispatch, eventRunner, node.name, 'TouchUp')}
    on:focus={() => emitEvent(dispatch, eventRunner, node.name, 'GotFocus')}
    on:blur={() => emitEvent(dispatch, eventRunner, node.name, 'LostFocus')}
  >
    {#if props.Prompt != null && props.Prompt !== ''}
      <option value="-1" disabled selected={numberOr(props.SelectionIndex, 0) < 1}>{props.Prompt}</option>
    {/if}
    {#each elements as item, index}
      <option value={index} selected={numberOr(props.SelectionIndex, 0) === index + 1 || (!props.Prompt && numberOr(props.SelectionIndex, 0) === 0 && index === 0)}>{item}</option>
    {/each}
  </select>
  {#if pickerOpen}
    <div class="sim-picker-menu" style={pickerMenuStyle()}>
      {#if props.Prompt != null && props.Prompt !== ''}
        <div class="sim-picker-title">{props.Prompt}</div>
      {/if}
      {#if boolValue(props.ShowFilterBar, false)}
        <input class="sim-picker-filter" value={pickerFilter} placeholder={props.HintText ?? 'Search list...'} on:input={e => pickerFilter = e.currentTarget.value} />
      {/if}
      {#each filteredPickerItems as item (item.index)}
        <button type="button" class:selected={numberOr(props.SelectionIndex, 0) === item.index + 1} on:click={() => pickSpinnerItem(item.index)}>{item.text}</button>
      {:else}
        <div class="sim-picker-empty"></div>
      {/each}
    </div>
  {/if}
</div>
