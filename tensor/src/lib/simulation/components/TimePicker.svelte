<script>
  import { createEventDispatcher, tick } from 'svelte';
  import { runAction } from '../actions.js';
  import { blurElement, focusElement, hideKeyboardFor, triggerNativePicker } from '../dom.js';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, buttonInnerStyle, containerStyle, numberOr, styleContext } from '../style.js';
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
  let timeInput;
  let buttonEl;
  let actionTokens = {};
  let suppressNextBlurEvent = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: handleActions(actions?.[node?.name] ?? {});

  function handleActions(actionState) {
    runAction(actionTokens, actionState, 'open', ['open', 'Open', 'DisplayDropdown', 'LaunchPicker'], () => openNativePicker(true));
    runAction(actionTokens, actionState, 'focus', ['focus', 'Focus', 'RequestFocus'], () => focusElement(buttonEl));
    runAction(actionTokens, actionState, 'hideKeyboard', ['hideKeyboard', 'hide-keyboard', 'HideKeyboard'], () => {
      suppressNextBlurEvent = true;
      hideKeyboardFor(buttonEl, timeInput).finally(() => setTimeout(() => { suppressNextBlurEvent = false; }, 0));
    });
    runAction(actionTokens, actionState, 'blur', ['blur', 'Blur'], () => blurElement(buttonEl));
  }

  function timeText() {
    return `${String(numberOr(props.Hour, 0)).padStart(2, '0')}:${String(numberOr(props.Minute, 0)).padStart(2, '0')}`;
  }

  function timeInstant(hour, minute) {
    const pad = n => String(n).padStart(2, '0');
    const iso = `1970-01-01T${pad(hour)}:${pad(minute)}:00`;
    const millis = new Date(1970, 0, 1, hour, minute, 0, 0).getTime();
    return { iso, millis };
  }

  function timeChange(e) {
    const nextTime = e.currentTarget.value;
    if (!nextTime) return;
    const [hour, minute] = nextTime.split(':').map(Number);
    if (![hour, minute].every(Number.isFinite)) return;
    const instant = timeInstant(hour, minute);
    emitInteraction(
      dispatch,
      [
        { component: node.name, property: 'Hour', value: hour },
        { component: node.name, property: 'Minute', value: minute },
        { component: node.name, property: 'Instant', value: instant.iso },
        { component: node.name, property: 'InstantMillis', value: instant.millis },
      ],
      { component: node.name, event: 'AfterTimeSet', args: [] },
    );
  }

  async function openNativePicker(fromAction = false) {
    if (!enabled) return;
    if (!fromAction) {
      await emitEvent(dispatch, eventRunner, node.name, 'Click');
      await tick();
      if (!enabled || !visible) return;
    }
    triggerNativePicker(timeInput);
  }

  function pointerDown() {
    if (!enabled) return;
    emitEvent(dispatch, eventRunner, node.name, 'TouchDown');
  }

  function pointerUp() {
    if (!enabled) return;
    emitEvent(dispatch, eventRunner, node.name, 'TouchUp');
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
</script>

<div class="sim-native-picker-wrap" class:sim-unsupported={unsupportedHere} style={containerStyle(props, parentType)} data-sim-component={node.name}>
  <button
    bind:this={buttonEl}
    type="button"
    class="sim-button"
    style={buttonInnerStyle(context, 'width: 100%; height: 100%;')}
    disabled={!enabled}
    on:pointerdown={pointerDown}
    on:pointerup={pointerUp}
    on:focus={focusEvent}
    on:blur={blurEvent}
    on:click={() => openNativePicker(false)}
  >{props.Text ?? ''}</button>
  <input bind:this={timeInput} class="sim-native-picker-input" type="time" disabled={!enabled} value={timeText()} on:change={timeChange} tabindex="-1" aria-hidden="true" />
</div>
