<script>
  import { assetName, baseStyle, emptyArrangementStyle, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
</script>

<div
  class="sim-arrangement"
  class:sim-unsupported={unsupportedHere}
  style={`${baseStyle(context)} ${emptyArrangementStyle(node, props)}`}
  data-sim-component={node.name}
>
  <slot />
</div>
