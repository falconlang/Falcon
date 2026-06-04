<script>
  import { createEventDispatcher, onDestroy } from 'svelte';
  import { runAction } from '../actions.js';
  import { blurElement, focusElement, hideKeyboardFor } from '../dom.js';
  import { emitEvent } from '../events.js';
  import { assetName, boolValue, buttonStyle, styleContext } from '../style.js';
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
  let buttonEl;
  let actionTokens = {};
  let longClickTimer = null;
  let suppressClick = false;
  let suppressNextBlurEvent = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: handleActions(actions?.[node?.name] ?? {});

  onDestroy(clearLongClick);

  function handleActions(actionState) {
    runAction(actionTokens, actionState, 'focus', ['focus', 'Focus', 'RequestFocus'], () => focusElement(buttonEl));
    runAction(actionTokens, actionState, 'hideKeyboard', ['hideKeyboard', 'hide-keyboard', 'HideKeyboard'], () => {
      suppressNextBlurEvent = true;
      hideKeyboardFor(buttonEl).finally(() => setTimeout(() => { suppressNextBlurEvent = false; }, 0));
    });
    runAction(actionTokens, actionState, 'blur', ['blur', 'Blur'], () => blurElement(buttonEl));
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

  function buttonClick() {
    if (!enabled || !visible || consumeLongClick()) return;
    emitEvent(dispatch, eventRunner, node.name, 'Click');
  }
</script>

<button
  bind:this={buttonEl}
  type="button"
  class="sim-button"
  class:sim-no-feedback={!boolValue(props.ShowFeedback, true)}
  class:sim-unsupported={unsupportedHere}
  style={buttonStyle(context)}
  disabled={!enabled}
  data-sim-component={node.name}
  on:pointerdown={pointerDown}
  on:pointerup={pointerUp}
  on:pointercancel={clearLongClick}
  on:focus={focusEvent}
  on:blur={blurEvent}
  on:click={buttonClick}
>
  {props.Text ?? ''}
</button>
