<script>
  import { assetName, baseStyle, colorValue, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
</script>

<div
  class="sim-circular-progress"
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(context, `--cp-color: ${colorValue(props.Color, '#0000ff')};`, { typography: false, arrangement: false, backgroundImage: false })}
  role="progressbar"
  aria-label="Loading"
  data-sim-component={node.name}
>
  <span class="sim-cp-ring" aria-hidden="true"></span>
</div>
