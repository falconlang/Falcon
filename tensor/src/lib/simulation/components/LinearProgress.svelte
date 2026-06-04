<script>
  import { assetName, baseStyle, boolValue, colorValue, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));

  function linearProgressPct() {
    const min = numberOr(props.Minimum, 0);
    const max = numberOr(props.Maximum, 100);
    const val = numberOr(props.Progress, 0);
    if (max <= min) return 0;
    return Math.max(0, Math.min(100, ((val - min) / (max - min)) * 100));
  }
</script>

<div
  class="sim-linear-progress"
  class:sim-linear-progress--indeterminate={boolValue(props.Indeterminate, true)}
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(context, `--lp-color: ${colorValue(props.ProgressColor, '#0000ff')}; --lp-ind-color: ${colorValue(props.IndeterminateColor, '#0000ff')}; --lp-pct: ${linearProgressPct()}%;`, { typography: false, arrangement: false, backgroundImage: false })}
  role="progressbar"
  aria-valuenow={boolValue(props.Indeterminate, true) ? undefined : numberOr(props.Progress, 0)}
  aria-valuemin={numberOr(props.Minimum, 0)}
  aria-valuemax={numberOr(props.Maximum, 100)}
  data-sim-component={node.name}
>
  <div class="sim-lp-bar" aria-hidden="true"></div>
</div>
