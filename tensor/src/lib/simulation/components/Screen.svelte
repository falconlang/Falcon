<script>
  import { assetName, baseStyle, boolValue, colorValue, hasValue, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
</script>

<div class="sim-screen-root" data-sim-component={node.name}>
  {#if boolValue(props.TitleVisible, true) && hasValue(props.Title)}
    <div class="sim-screen-titlebar" style="background: {colorValue(props.PrimaryColor, '#3f51b5')};">{props.Title}</div>
  {/if}
  <div
    class="sim-screen"
    class:sim-screen--scrollable={boolValue(props.Scrollable, false)}
    class:sim-unsupported={unsupportedHere}
    style={baseStyle(context, '', { size: false })}
  >
    <slot />
  </div>
</div>
