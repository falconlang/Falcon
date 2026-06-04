<script>
  import { createEventDispatcher } from 'svelte';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, boolValue, colorValue, styleContext } from '../style.js';
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

  function switchStyle() {
    return baseStyle(context, [
      `--switch-thumb-active: ${colorValue(props.ThumbColorActive, '#ffffff')};`,
      `--switch-thumb-inactive: ${colorValue(props.ThumbColorInactive, '#cccccc')};`,
      `--switch-track-active: ${colorValue(props.TrackColorActive, '#00ff00')};`,
      `--switch-track-inactive: ${colorValue(props.TrackColorInactive, '#444444')};`,
    ].join(' '));
  }

  function checkboxInput(e) {
    emitInteraction(
      dispatch,
      [{ component: node.name, property: 'On', value: e.currentTarget.checked }],
      { component: node.name, event: 'Changed', args: [] },
    );
  }
</script>

<label class="sim-check sim-switch" class:sim-unsupported={unsupportedHere} style={switchStyle()} data-sim-component={node.name}>
  <input
    type="checkbox"
    checked={boolValue(props.On, false)}
    disabled={!enabled}
    on:change={checkboxInput}
    on:focus={() => emitEvent(dispatch, eventRunner, node.name, 'GotFocus')}
    on:blur={() => emitEvent(dispatch, eventRunner, node.name, 'LostFocus')}
  />
  <span class="sim-switch-ui" aria-hidden="true"></span>
  <span>{props.Text ?? ''}</span>
</label>
