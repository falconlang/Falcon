<script>
  import { createEventDispatcher } from 'svelte';
  import { emitEvent } from '../events.js';
  import { assetName, baseStyle, boolValue, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let visible = true;
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  $: assetUrl = resolveAssetUrl(assets, assetName(props));
  $: context = styleContext(node, props, parentType, assetUrl);

  function imageClick() {
    if (!visible || !boolValue(props.Clickable, false)) return;
    emitEvent(dispatch, eventRunner, node.name, 'Click');
  }
</script>

<button
  type="button"
  class="sim-image"
  class:sim-image-clickable={boolValue(props.Clickable, false)}
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(context, `transform: rotate(${numberOr(props.RotationAngle, 0)}deg);`, { backgroundImage: false })}
  disabled={!boolValue(props.Clickable, false)}
  data-sim-component={node.name}
  on:click={imageClick}
>
  {#if assetUrl}
    <img class:sim-image-fill={boolValue(props.ScalePictureToFit, false) || numberOr(props.Scaling, 0) === 1} src={assetUrl} alt={props.AlternateText ?? ''} />
  {:else}
    <span>{props.Picture ?? ''}</span>
  {/if}
</button>
