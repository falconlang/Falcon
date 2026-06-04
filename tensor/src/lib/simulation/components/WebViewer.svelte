<script>
  import { createEventDispatcher } from 'svelte';
  import { runAction } from '../actions.js';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  let webViewerFrame;
  let actionTokens = {};

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: handleActions(actions?.[node?.name] ?? {});

  function handleActions(actionState) {
    runAction(actionTokens, actionState, 'navigate', ['navigate'], () => {
      const url = actionState.url ?? '';
      if (webViewerFrame) webViewerFrame.src = url;
    });
    runAction(actionTokens, actionState, 'goback', ['goback'], () => {
      try { webViewerFrame?.contentWindow?.history?.back(); } catch {}
    });
    runAction(actionTokens, actionState, 'goforward', ['goforward'], () => {
      try { webViewerFrame?.contentWindow?.history?.forward(); } catch {}
    });
    runAction(actionTokens, actionState, 'reload', ['reload'], () => {
      try { webViewerFrame?.contentWindow?.location?.reload(); } catch {}
    });
  }

  function webViewerLoad(e) {
    try {
      const url = e.currentTarget.contentWindow?.location?.href || props.CurrentUrl || '';
      emitInteraction(dispatch, [{ component: node.name, property: 'CurrentUrl', value: url }], null);
      emitEvent(dispatch, eventRunner, node.name, 'PageLoaded', [url]);
    } catch {}
  }

  function webViewerError() {
    emitEvent(dispatch, eventRunner, node.name, 'ErrorOccurred', [-6, 'Failed to load', props.CurrentUrl || '']);
  }
</script>

<div class="sim-webviewer" class:sim-unsupported={unsupportedHere} style={baseStyle(context)} data-sim-component={node.name}>
  {#if props.HomeUrl || props.CurrentUrl}
    <iframe
      bind:this={webViewerFrame}
      title="WebViewer"
      src={props.CurrentUrl || props.HomeUrl || ''}
      class="sim-webviewer-frame"
      sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
      allow="autoplay; camera; microphone"
      on:load={webViewerLoad}
      on:error={webViewerError}
    ></iframe>
  {:else}
    <div class="sim-webviewer-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M3 8h18M8 3v18"/></svg>
      <span>WebViewer</span>
    </div>
  {/if}
</div>
