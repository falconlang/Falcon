<script>
  import { createEventDispatcher, onDestroy, tick } from 'svelte';
  import { runAction } from '../actions.js';
  import { blurElement, focusElement, hideKeyboardFor } from '../dom.js';
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
  let buttonEl;
  let actionTokens = {};
  let longClickTimer = null;
  let suppressClick = false;
  let suppressNextBlurEvent = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: elements = elementItems(props).map(itemMainText);
  $: filteredPickerItems = filterIndexed(elements, pickerFilter);
  $: handleActions(actions?.[node?.name] ?? {});

  onDestroy(clearLongClick);

  function handleActions(actionState) {
    runAction(actionTokens, actionState, 'open', ['open', 'Open', 'DisplayDropdown', 'LaunchPicker'], () => {
      if (enabled && visible) {
        pickerFilter = '';
        pickerOpen = true;
      }
    });
    runAction(actionTokens, actionState, 'focus', ['focus', 'Focus', 'RequestFocus'], () => focusElement(buttonEl));
    runAction(actionTokens, actionState, 'hideKeyboard', ['hideKeyboard', 'hide-keyboard', 'HideKeyboard'], () => {
      suppressNextBlurEvent = true;
      hideKeyboardFor(buttonEl).finally(() => setTimeout(() => { suppressNextBlurEvent = false; }, 0));
    });
    runAction(actionTokens, actionState, 'blur', ['blur', 'Blur'], () => blurElement(buttonEl));
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

  function focusEvent() {
    emitEvent(dispatch, eventRunner, node.name, 'GotFocus');
  }

  function blurEvent() {
    if (suppressNextBlurEvent) {
      suppressNextBlurEvent = false;
      return;
    }
    emitEvent(dispatch, eventRunner, node.name, 'LostFocus');
  }

  function clearLongClick() {
    if (longClickTimer) clearTimeout(longClickTimer);
    longClickTimer = null;
  }

  function pointerDown() {
    if (!enabled) return;
    emitEvent(dispatch, eventRunner, node.name, 'TouchDown');
    suppressClick = false;
    clearLongClick();
    longClickTimer = setTimeout(() => {
      longClickTimer = null;
      suppressClick = true;
      emitEvent(dispatch, eventRunner, node.name, 'LongClick');
    }, 600);
  }

  function pointerUp() {
    if (!enabled) return;
    clearLongClick();
    emitEvent(dispatch, eventRunner, node.name, 'TouchUp');
  }

  function consumeLongClick() {
    if (!suppressClick) return false;
    suppressClick = false;
    return true;
  }

  async function openListPicker() {
    if (!enabled || consumeLongClick()) return;
    await emitEvent(dispatch, eventRunner, node.name, 'BeforePicking');
    await tick();
    if (!enabled || !visible) return;
    pickerFilter = '';
    pickerOpen = true;
  }

  function pickListItem(index) {
    const next = selectionByIndex(index, elements);
    if (!next) return;
    pickerOpen = false;
    emitInteraction(
      dispatch,
      [
        { component: node.name, property: 'Selection', value: next.selection },
        { component: node.name, property: 'SelectionIndex', value: next.selectionIndex },
      ],
      { component: node.name, event: 'AfterPicking', args: [] },
    );
  }
</script>

<svelte:window on:pointerdown={onWindowPointerDown} />

<div bind:this={pickerWrapEl} class="sim-picker" class:sim-unsupported={unsupportedHere} style={containerStyle(props, parentType)} data-sim-component={node.name}>
  <button
    bind:this={buttonEl}
    type="button"
    class="sim-button"
    class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
    style={buttonInnerStyle(context, 'width: 100%;')}
    disabled={!enabled}
    on:pointerdown={pointerDown}
    on:pointerup={pointerUp}
    on:pointercancel={clearLongClick}
    on:focus={focusEvent}
    on:blur={blurEvent}
    on:click={openListPicker}
  >{props.Text ?? ''}</button>
  {#if pickerOpen}
    <div class="sim-picker-menu" style={pickerMenuStyle()}>
      {#if props.Title != null && props.Title !== ''}
        <div class="sim-picker-title">{props.Title}</div>
      {/if}
      {#if boolValue(props.ShowFilterBar, false)}
        <input class="sim-picker-filter" value={pickerFilter} placeholder={props.HintText ?? 'Search list...'} on:input={e => pickerFilter = e.currentTarget.value} />
      {/if}
      {#each filteredPickerItems as item (item.index)}
        <button type="button" on:click={() => pickListItem(item.index)}>{item.text}</button>
      {:else}
        <div class="sim-picker-empty"></div>
      {/each}
    </div>
  {/if}
</div>
