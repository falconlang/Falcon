<script>
  import { createEventDispatcher, tick } from 'svelte';
  import { runAction } from '../actions.js';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, boolValue, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  let videoEl;
  let actionTokens = {};

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: handleActions(actions?.[node?.name] ?? {});

  function handleActions(actionState) {
    runAction(actionTokens, actionState, 'play', ['play'], () => { videoEl?.play().catch(() => {}); });
    runAction(actionTokens, actionState, 'pause', ['pause'], () => { videoEl?.pause(); });
    runAction(actionTokens, actionState, 'fullscreen', ['fullscreen'], () => requestVideoFullscreen());
    runAction(actionTokens, actionState, 'seek', ['seek'], () => {
      if (videoEl) videoEl.currentTime = numberOr(actionState.ms, 0) / 1000;
    });
  }

  function videoLoadedMetadata() {
    if (!videoEl) return;
    const duration = Math.round((videoEl.duration || 0) * 1000);
    emitInteraction(dispatch, [{ component: node.name, property: 'Duration', value: duration }], null);
  }

  async function requestVideoFullscreen() {
    await tick();
    const target = videoEl?.parentElement || videoEl;
    try {
      if (target?.requestFullscreen) await target.requestFullscreen();
      else if (videoEl?.webkitEnterFullscreen) videoEl.webkitEnterFullscreen();
    } catch {}
  }
</script>

<div class="sim-videoplayer" class:sim-unsupported={unsupportedHere} style={baseStyle(context)} data-sim-component={node.name}>
  {#if props.Source}
    <video
      bind:this={videoEl}
      class="sim-videoplayer-video"
      src={resolveAssetUrl(assets, props.Source) || props.Source}
      volume={Math.max(0, Math.min(1, numberOr(props.Volume, 50) / 100))}
      loop={boolValue(props.Loop, false)}
      preload="metadata"
      controls
      on:ended={() => emitEvent(dispatch, eventRunner, node.name, 'Completed')}
      on:error={(e) => emitEvent(dispatch, eventRunner, node.name, 'VideoPlayerError', [e.currentTarget.error?.message || 'Error'])}
      on:loadedmetadata={videoLoadedMetadata}
    >
      <track kind="captions" />
    </video>
  {:else}
    <div class="sim-videoplayer-empty">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><circle cx="12" cy="12" r="9"/><polygon points="10,8 16,12 10,16" fill="currentColor" stroke="none"/></svg>
      <span>VideoPlayer</span>
    </div>
  {/if}
</div>
