<script>
  import FilePickerBase from './FilePickerBase.svelte';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let enabled = true;
  export let visible = true;
  export let actions = {};
  export let eventRunner = null;
</script>

<FilePickerBase
  {node}
  {props}
  {assets}
  {parentType}
  {unsupportedHere}
  {enabled}
  {visible}
  {actions}
  {eventRunner}
  on:event
  on:property
  on:interaction
/>
