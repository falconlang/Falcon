<script>
  import { createEventDispatcher } from 'svelte';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, boolValue, styleContext } from '../style.js';
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

  function checkboxInput(e) {
    emitInteraction(
      dispatch,
      [{ component: node.name, property: 'Checked', value: e.currentTarget.checked }],
      { component: node.name, event: 'Changed', args: [] },
    );
  }
</script>

<label class="sim-check" class:sim-unsupported={unsupportedHere} style={baseStyle(context)} data-sim-component={node.name}>
  <input
    type="checkbox"
    checked={boolValue(props.Checked, false)}
    disabled={!enabled}
    on:change={checkboxInput}
    on:focus={() => emitEvent(dispatch, eventRunner, node.name, 'GotFocus')}
    on:blur={() => emitEvent(dispatch, eventRunner, node.name, 'LostFocus')}
  />
  <span>{props.Text ?? ''}</span>
</label>
