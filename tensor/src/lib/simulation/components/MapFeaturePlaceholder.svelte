<script>
  import UnsupportedComponent from './UnsupportedComponent.svelte';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
</script>

<UnsupportedComponent {node} {props} {assets} {parentType} />
