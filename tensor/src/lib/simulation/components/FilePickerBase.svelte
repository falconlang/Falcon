<script>
  import { createEventDispatcher, onDestroy, tick } from 'svelte';
  import { blurElement, focusElement, hideKeyboardFor } from '../dom.js';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, boolValue, buttonInnerStyle, colorValue, containerStyle, firstNonEmpty, styleContext } from '../style.js';
  import { runAction } from '../actions.js';
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
  const MOCK_CONTACTS = [
    { name: 'Alice Example', phone: '555-0101', email: 'alice@example.com' },
    { name: 'Bob Example', phone: '555-0102', email: 'bob@example.com' },
    { name: 'Carol Example', phone: '555-0103', email: 'carol@example.com' },
  ];

  let fileInput;
  let pickerWrapEl;
  let buttonEl;
  let pickerOpen = false;
  let imagePickerObjectUrl = '';
  let longClickTimer = null;
  let suppressClick = false;
  let suppressNextBlurEvent = false;
  let actionTokens = {};

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: handleActions(actions?.[node?.name] ?? {});

  onDestroy(() => {
    clearLongClick();
    if (imagePickerObjectUrl) URL.revokeObjectURL(imagePickerObjectUrl);
  });

  function handleActions(actionState) {
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

  function filePickerAccept() {
    if (node?.type === 'ImagePicker') return 'image/*';
    if (node?.type === 'FilePicker') return props.MimeType && props.MimeType !== '*/*' ? props.MimeType : '';
    return '';
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

  function dismissPicker() {
    pickerOpen = false;
  }

  function onWindowPointerDown(e) {
    if (!pickerOpen) return;
    if (pickerWrapEl && pickerWrapEl.contains(e.target)) return;
    dismissPicker();
  }

  async function openFilePicker() {
    if (!enabled || consumeLongClick()) return;
    await emitEvent(dispatch, eventRunner, node.name, 'BeforePicking');
    await tick();
    if (!enabled || !visible) return;
    if (node?.type === 'ContactPicker' || node?.type === 'PhoneNumberPicker') {
      pickerOpen = true;
    } else {
      fileInput?.click();
    }
  }

  function filePickerChange(e) {
    const file = e.currentTarget.files?.[0];
    if (!file) return;
    const patches = [{ component: node.name, property: 'Selection', value: file.name }];
    if (node?.type === 'ImagePicker') {
      const url = URL.createObjectURL(file);
      if (imagePickerObjectUrl) URL.revokeObjectURL(imagePickerObjectUrl);
      imagePickerObjectUrl = url;
      patches.push({ component: node.name, property: 'ImagePath', value: url });
    }
    emitInteraction(dispatch, patches, { component: node.name, event: 'AfterPicking', args: [] });
    e.currentTarget.value = '';
  }

  function pickContact(contact) {
    pickerOpen = false;
    const patches = [];
    if (node?.type === 'PhoneNumberPicker') {
      patches.push({ component: node.name, property: 'PhoneNumber', value: contact.phone });
      patches.push({ component: node.name, property: 'ContactName', value: contact.name });
    } else {
      patches.push({ component: node.name, property: 'ContactName', value: contact.name });
      patches.push({ component: node.name, property: 'EmailAddress', value: contact.email });
      patches.push({ component: node.name, property: 'PhoneNumber', value: contact.phone });
    }
    patches.push({ component: node.name, property: 'Selection', value: contact.name });
    emitInteraction(dispatch, patches, { component: node.name, event: 'AfterPicking', args: [] });
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
    on:click={openFilePicker}
  >{props.Text ?? ''}</button>
  <input
    bind:this={fileInput}
    type="file"
    accept={filePickerAccept()}
    class="sim-native-picker-input"
    tabindex="-1"
    aria-hidden="true"
    on:change={filePickerChange}
  />
  {#if node.type === 'ContactPicker' || node.type === 'PhoneNumberPicker'}
    {#if pickerOpen}
      <div class="sim-picker-menu" style={pickerMenuStyle()}>
        <div class="sim-picker-title">
          {node.type === 'PhoneNumberPicker' ? 'Pick Phone Number' : 'Pick Contact'}
        </div>
        {#each MOCK_CONTACTS as contact, i (i)}
          <button type="button" on:click={() => pickContact(contact)}>{contact.name}</button>
        {/each}
      </div>
    {/if}
  {/if}
</div>
