<script>
  import { assetName, baseStyle, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
</script>

<div class="sim-unsupported sim-placeholder" style={baseStyle(context)} data-sim-component={node.name}>
  {node.type}.{node.name}
</div>
