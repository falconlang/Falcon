<script>
  import Screen from './Screen.svelte';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
</script>

<Screen {node} {props} {assets} {parentType} {unsupportedHere}>
  <slot />
</Screen>
