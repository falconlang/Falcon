<script>
  import { createEventDispatcher } from 'svelte';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, boolValue, colorValue, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let enabled = true;
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));

  function sliderMin() {
    return numberOr(props.MinValue, 0);
  }

  function sliderMax() {
    return numberOr(props.MaxValue, 100);
  }

  function sliderValue() {
    return Math.max(sliderMin(), Math.min(sliderMax(), numberOr(props.ThumbPosition, sliderMin())));
  }

  function sliderStep() {
    const steps = numberOr(props.NumberOfSteps, 100);
    const span = Math.abs(sliderMax() - sliderMin());
    if (steps <= 0 || span === 0) return 'any';
    return String(span / steps);
  }

  function sliderStyle() {
    const span = sliderMax() - sliderMin();
    const progress = span === 0 ? 0 : Math.max(0, Math.min(100, ((sliderValue() - sliderMin()) / span) * 100));
    return baseStyle(context, [
      `--slider-left: ${colorValue(props.ColorLeft, '#ffc800')};`,
      `--slider-right: ${colorValue(props.ColorRight, '#888888')};`,
      `--slider-thumb: ${colorValue(props.ThumbColor, '#444444')};`,
      `--slider-progress: ${progress}%;`,
    ].join(' '));
  }

  function sliderInput(e) {
    if (!boolValue(props.ThumbEnabled, true)) {
      e.preventDefault();
      return;
    }
    const value = Number(e.currentTarget.value);
    emitInteraction(
      dispatch,
      [{ component: node.name, property: 'ThumbPosition', value }],
      { component: node.name, event: 'PositionChanged', args: [value] },
    );
  }

  function sliderPointerDown(e) {
    if (!boolValue(props.ThumbEnabled, true)) {
      e.preventDefault();
      return;
    }
    emitEvent(dispatch, eventRunner, node.name, 'TouchDown');
  }

  function sliderPointerUp(e) {
    if (!boolValue(props.ThumbEnabled, true)) {
      e.preventDefault();
      return;
    }
    emitEvent(dispatch, eventRunner, node.name, 'TouchUp');
  }

  function sliderKeydown(e) {
    if (boolValue(props.ThumbEnabled, true)) return;
    e.preventDefault();
  }
</script>

<input
  class="sim-slider"
  class:sim-slider-thumb-disabled={!boolValue(props.ThumbEnabled, true)}
  class:sim-unsupported={unsupportedHere}
  style={sliderStyle()}
  type="range"
  min={sliderMin()}
  max={sliderMax()}
  step={sliderStep()}
  value={sliderValue()}
  disabled={!enabled}
  data-sim-component={node.name}
  on:pointerdown={sliderPointerDown}
  on:pointerup={sliderPointerUp}
  on:keydown={sliderKeydown}
  on:input={sliderInput}
/>
