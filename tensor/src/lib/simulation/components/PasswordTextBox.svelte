<script>
  import { createEventDispatcher, tick } from 'svelte';
  import { runAction } from '../actions.js';
  import { blurElement, focusElement, hideKeyboardFor } from '../dom.js';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, boolValue, hintColorStyle, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let enabled = true;
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  let textInputEl;
  let actionTokens = {};
  let suppressNextBlurEvent = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: handleActions(actions?.[node?.name] ?? {});

  function handleActions(actionState) {
    runAction(actionTokens, actionState, 'focus', ['focus', 'Focus', 'RequestFocus'], () => focusElement(textInputEl));
    runAction(actionTokens, actionState, 'hideKeyboard', ['hideKeyboard', 'hide-keyboard', 'HideKeyboard'], () => {
      suppressNextBlurEvent = true;
      hideKeyboardFor(textInputEl).finally(() => setTimeout(() => { suppressNextBlurEvent = false; }, 0));
    });
    runAction(actionTokens, actionState, 'blur', ['blur', 'Blur'], () => blurElement(textInputEl));
    runAction(actionTokens, actionState, 'cursorStart', ['MoveCursorToStart', 'cursorStart', 'cursor-start'], () => setTextCursor(0));
    runAction(actionTokens, actionState, 'cursorEnd', ['MoveCursorToEnd', 'cursorEnd', 'cursor-end'], () => setTextCursor(textValue().length));
    runAction(actionTokens, actionState, 'cursorTo', ['MoveCursorTo', 'cursorTo', 'cursor-position'], () => {
      const position = Number(actionState.position ?? props.CursorPosition ?? props.SelectionStart ?? 1);
      setTextCursor(Math.max(0, position - 1));
    });
  }

  function textValue() {
    return String(props.Text ?? '');
  }

  async function setTextCursor(position) {
    await tick();
    if (!textInputEl || typeof textInputEl.setSelectionRange !== 'function') return;
    const next = Math.max(0, Math.min(position, textValue().length));
    textInputEl.focus();
    textInputEl.setSelectionRange(next, next);
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

  function textInput(e) {
    emitInteraction(
      dispatch,
      [{ component: node.name, property: 'Text', value: e.currentTarget.value }],
      { component: node.name, event: 'TextChanged', args: [] },
    );
  }
</script>

<input
  bind:this={textInputEl}
  class="sim-textbox"
  class:sim-unsupported={unsupportedHere}
  style={`${baseStyle(context)} ${hintColorStyle(props)}`}
  type={boolValue(props.PasswordVisible, false) ? 'text' : 'password'}
  value={props.Text ?? ''}
  placeholder={props.Hint ?? ''}
  disabled={!enabled}
  readonly={boolValue(props.ReadOnly, false)}
  inputmode={boolValue(props.NumbersOnly, false) ? 'decimal' : 'text'}
  data-sim-component={node.name}
  on:input={textInput}
  on:focus={focusEvent}
  on:blur={blurEvent}
/>
