<script>
  import { assetName, baseStyle, boolValue, styleContext } from '../style.js';
  import { labelHtml } from '../label-html.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
</script>

<div
  class="sim-label"
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(context, boolValue(props.HasMargins, true) ? 'margin: 2px;' : 'margin: 0;')}
  data-sim-component={node.name}
>
  {#if boolValue(props.HTMLFormat, false)}{@html labelHtml(props.Text)}{:else}{props.Text ?? ''}{/if}
</div>
