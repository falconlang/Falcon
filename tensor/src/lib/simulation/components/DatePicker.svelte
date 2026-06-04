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
  const MONTHS = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December',
  ];

  let dateInput;
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
      hideKeyboardFor(buttonEl, dateInput).finally(() => setTimeout(() => { suppressNextBlurEvent = false; }, 0));
    });
    runAction(actionTokens, actionState, 'blur', ['blur', 'Blur'], () => blurElement(buttonEl));
  }

  function currentDate() {
    const now = new Date();
    return { year: now.getFullYear(), month: now.getMonth() + 1, day: now.getDate() };
  }

  function dateParts() {
    const today = currentDate();
    return {
      year: numberOr(props.Year, today.year),
      month: numberOr(props.Month, today.month),
      day: numberOr(props.Day, today.day),
    };
  }

  function dateText() {
    const { year, month, day } = dateParts();
    return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
  }

  function dateInstant(year, month, day) {
    const pad = n => String(n).padStart(2, '0');
    const iso = `${year}-${pad(month)}-${pad(day)}T00:00:00`;
    const millis = new Date(year, month - 1, day, 0, 0, 0, 0).getTime();
    return { iso, millis };
  }

  function dateChange(e) {
    const nextDate = e.currentTarget.value;
    if (!nextDate) return;
    const [year, month, day] = nextDate.split('-').map(Number);
    if (![year, month, day].every(Number.isFinite)) return;
    const instant = dateInstant(year, month, day);
    emitInteraction(
      dispatch,
      [
        { component: node.name, property: 'Year', value: year },
        { component: node.name, property: 'Month', value: month },
        { component: node.name, property: 'Day', value: day },
        { component: node.name, property: 'MonthInText', value: MONTHS[month - 1] ?? '' },
        { component: node.name, property: 'Instant', value: instant.iso },
        { component: node.name, property: 'InstantMillis', value: instant.millis },
      ],
      { component: node.name, event: 'AfterDateSet', args: [] },
    );
  }

  async function openNativePicker(fromAction = false) {
    if (!enabled) return;
    if (!fromAction) {
      await emitEvent(dispatch, eventRunner, node.name, 'Click');
      await tick();
      if (!enabled || !visible) return;
    }
    triggerNativePicker(dateInput);
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
  <input bind:this={dateInput} class="sim-native-picker-input" type="date" disabled={!enabled} value={dateText()} on:change={dateChange} tabindex="-1" aria-hidden="true" />
</div>
