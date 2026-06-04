<script>
  import { assetName, baseStyle, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
</script>

<div
  class="sim-table"
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(
    context,
    `grid-template-columns: repeat(${numberOr(props.Columns, 2)}, 1fr); grid-template-rows: repeat(${numberOr(props.Rows, 2)}, auto);`,
    { typography: false, arrangement: false, backgroundImage: false },
  )}
  data-sim-component={node.name}
>
  <slot />
</div>
