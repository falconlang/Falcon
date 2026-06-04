<script>
  import MapFeaturePlaceholder from './MapFeaturePlaceholder.svelte';
  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
</script>

<MapFeaturePlaceholder {node} {props} {assets} {parentType} />
