<script>
  import { createEventDispatcher } from 'svelte';
  import { elementsFromString } from '../../simulation-capabilities.js';
  import { emitEvent } from '../events.js';
  import { assetName, baseStyle, boolValue, colorValue, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let state = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  const CHART_PAD = 32;
  const CHART_COLORS = ['#2196F3', '#F44336', '#4CAF50', '#FF9800', '#9C27B0', '#00BCD4'];

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));

  function chartWidth() {
    return Math.max(80, numberOr(props.Width, 300));
  }

  function chartHeight() {
    return Math.max(60, numberOr(props.Height, 300));
  }

  function chartDataSeries() {
    if (!node?.children) return [];
    return node.children
      .filter(c => c.type === 'ChartData2D')
      .map((c, i) => {
        const sp = state?.[c.name] ?? {};
        const pts = parseChartPoints(sp.Elements || sp.ElementsFromPairs || '');
        const colors = parseChartColors(sp.Colors);
        return {
          label: sp.Label || c.name,
          color: colorValue(sp.Color, CHART_COLORS[i % CHART_COLORS.length]),
          colors,
          dataLabelColor: colorValue(sp.DataLabelColor, '#000000'),
          highlightColor: colorValue(sp.HighlightColor, ''),
          points: pts,
          name: c.name,
        };
      });
  }

  function parseChartPoints(value) {
    if (Array.isArray(value)) return value.map(p => Array.isArray(p) ? [Number(p[0]) || 0, Number(p[1]) || 0] : [0, 0]);
    const text = String(value ?? '').trim();
    if (!text) return [];
    const pairs = text.split(',');
    const pts = [];
    for (let i = 0; i + 1 < pairs.length; i += 2) {
      pts.push([Number(pairs[i].trim()) || 0, Number(pairs[i + 1].trim()) || 0]);
    }
    return pts;
  }

  function parseChartColors(value) {
    if (Array.isArray(value)) return value.map(item => colorValue(item, '')).filter(Boolean);
    const text = String(value ?? '').trim();
    if (!text) return [];
    return text.split(',').map(item => colorValue(item.trim(), '')).filter(Boolean);
  }

  function chartXRange() {
    const all = chartDataSeries().flatMap(s => s.points.map(p => p[0]));
    const manualMin = finiteNumber(props.XMin);
    const manualMax = finiteNumber(props.XMax);
    if (manualMin != null && manualMax != null && manualMax > manualMin) return [manualMin, manualMax];
    if (!all.length) return [0, 1];
    const min = manualMin ?? (boolValue(props.XFromZero, false) ? 0 : Math.min(...all));
    const max = manualMax ?? Math.max(...all, min + 1);
    return [min, Math.max(max, min + 1)];
  }

  function chartYRange() {
    const all = chartDataSeries().flatMap(s => s.points.map(p => p[1]));
    const manualMin = finiteNumber(props.YMin);
    const manualMax = finiteNumber(props.YMax);
    if (manualMin != null && manualMax != null && manualMax > manualMin) return [manualMin, manualMax];
    if (!all.length) return [0, 1];
    const min = manualMin ?? (boolValue(props.YFromZero, false) ? 0 : Math.min(...all));
    const max = manualMax ?? Math.max(...all, min + 1);
    return [min, Math.max(max, min + 1)];
  }

  function finiteNumber(value) {
    if (value === null || value === undefined || value === '') return null;
    const n = Number(value);
    return Number.isFinite(n) ? n : null;
  }

  function chartX(v) {
    const [xMin, xMax] = chartXRange();
    const w = chartWidth() - CHART_PAD * 2;
    return CHART_PAD + ((v - xMin) / (xMax - xMin)) * w;
  }

  function chartY(v) {
    const [yMin, yMax] = chartYRange();
    const h = chartHeight() - CHART_PAD * 2;
    return chartHeight() - CHART_PAD - ((v - yMin) / (yMax - yMin)) * h;
  }

  function chartPolylinePoints(pts) {
    return pts.map(p => `${chartX(p[0])},${chartY(p[1])}`).join(' ');
  }

  function chartTicks(range, count = 5) {
    const [min, max] = range;
    const span = max - min;
    if (!Number.isFinite(span) || span <= 0) return [];
    return Array.from({ length: count }, (_, i) => min + (span * i) / (count - 1));
  }

  function chartXTicks() {
    return chartTicks(chartXRange());
  }

  function chartYTicks() {
    return chartTicks(chartYRange());
  }

  function chartTickText(value) {
    if (Math.abs(value) >= 1000 || (Math.abs(value) > 0 && Math.abs(value) < 0.01)) return value.toExponential(1);
    return Number.isInteger(value) ? String(value) : String(Math.round(value * 100) / 100);
  }

  function chartLabels() {
    if (Array.isArray(props.Labels)) return props.Labels.map(item => String(item ?? ''));
    return elementsFromString(props.Labels);
  }

  function chartXTickText(value, index) {
    return chartLabels()[index] || chartTickText(value);
  }

  function chartPointColor(series, index) {
    return series.colors[index] || series.color;
  }

  function chartAreaPath(pts) {
    if (!pts.length) return '';
    const base = chartHeight() - CHART_PAD;
    const points = [`${chartX(pts[0][0])},${base}`]
      .concat(pts.map(p => `${chartX(p[0])},${chartY(p[1])}`))
      .concat([`${chartX(pts[pts.length - 1][0])},${base}`]);
    return `M${points.join('L')}Z`;
  }

  function chartBarX(si, pi, numSeries, numPts) {
    const w = chartWidth() - CHART_PAD * 2;
    const groupW = w / Math.max(1, numPts);
    const barW = chartBarWidth(numSeries, numPts);
    return CHART_PAD + pi * groupW + si * barW + (groupW - numSeries * barW) / 2;
  }

  function chartBarWidth(numSeries, numPts) {
    const w = chartWidth() - CHART_PAD * 2;
    return Math.max(2, (w / Math.max(1, numPts)) / Math.max(1, numSeries) - 2);
  }

  function pieSectors(pts, si) {
    const total = pts.reduce((s, p) => s + Math.abs(p[1]), 0) || 1;
    const cx = chartWidth() / 2;
    const cy = chartHeight() / 2;
    const radiusPct = Math.max(0, Math.min(100, numberOr(props.PieRadius, 100))) / 100;
    const r = (Math.min(cx, cy) - CHART_PAD) * radiusPct;
    let angle = -Math.PI / 2;
    return pts.map((p, i) => {
      const sweep = (Math.abs(p[1]) / total) * Math.PI * 2;
      const x1 = cx + r * Math.cos(angle);
      const y1 = cy + r * Math.sin(angle);
      const x2 = cx + r * Math.cos(angle + sweep);
      const y2 = cy + r * Math.sin(angle + sweep);
      const large = sweep > Math.PI ? 1 : 0;
      const d = `M${cx},${cy}L${x1},${y1}A${r},${r},0,${large},1,${x2},${y2}Z`;
      const series = chartDataSeries()[si];
      const fill = series?.colors?.[i] || CHART_COLORS[(si * pts.length + i) % CHART_COLORS.length];
      angle += sweep;
      return { d, fill };
    });
  }

  function chartEntryClick(series, pt) {
    emitEvent(dispatch, eventRunner, node.name, 'EntryClick', [series.label, pt[0], pt[1]]);
    emitEvent(dispatch, eventRunner, series.name, 'EntryClick', [pt[0], pt[1]]);
  }

  function chartTrendlines() {
    const series = chartDataSeries();
    return (node?.children || [])
      .filter(c => c.type === 'Trendline')
      .map((trend, index) => {
        const sp = state?.[trend.name] ?? {};
        if (sp.Visible === false) return null;
        const targetName = String(sp.ChartData ?? '').trim();
        const target = series.find(item => item.name === targetName || item.label === targetName) || series[index] || series[0];
        if (!target || target.points.length < 2) return null;
        const regression = linearRegression(target.points);
        if (!regression) return null;
        const [xMin, xMax] = chartXRange();
        const y1 = regression.slope * xMin + regression.intercept;
        const y2 = regression.slope * xMax + regression.intercept;
        return {
          name: trend.name,
          d: `M${chartX(xMin)},${chartY(y1)}L${chartX(xMax)},${chartY(y2)}`,
          color: colorValue(sp.Color, target.color),
          width: Math.max(1, numberOr(sp.StrokeWidth, 1)),
          dash: numberOr(sp.StrokeStyle, 1) === 2 ? '6 4' : numberOr(sp.StrokeStyle, 1) === 3 ? '2 4' : '',
        };
      })
      .filter(Boolean);
  }

  function linearRegression(points) {
    const n = points.length;
    if (n < 2) return null;
    let sumX = 0;
    let sumY = 0;
    let sumXY = 0;
    let sumXX = 0;
    for (const [x, y] of points) {
      sumX += x;
      sumY += y;
      sumXY += x * y;
      sumXX += x * x;
    }
    const denom = n * sumXX - sumX * sumX;
    if (denom === 0) return null;
    const slope = (n * sumXY - sumX * sumY) / denom;
    const intercept = (sumY - slope * sumX) / n;
    return { slope, intercept };
  }
</script>

<div
  class="sim-chart"
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(context)}
  data-sim-component={node.name}
>
  <svg class="sim-chart-svg" viewBox={`0 0 ${chartWidth()} ${chartHeight()}`} preserveAspectRatio="none">
    {#if numberOr(props.Type, 0) !== 4}
      {#if boolValue(props.GridEnabled, true)}
        {#each chartXTicks() as tick}
          <line class="sim-chart-grid-line" x1={chartX(tick)} y1={CHART_PAD} x2={chartX(tick)} y2={chartHeight() - CHART_PAD} />
        {/each}
        {#each chartYTicks() as tick}
          <line class="sim-chart-grid-line" x1={CHART_PAD} y1={chartY(tick)} x2={chartWidth() - CHART_PAD} y2={chartY(tick)} />
        {/each}
      {/if}
      <line class="sim-chart-axis-line" x1={CHART_PAD} y1={chartHeight() - CHART_PAD} x2={chartWidth() - CHART_PAD} y2={chartHeight() - CHART_PAD} />
      <line class="sim-chart-axis-line" x1={CHART_PAD} y1={CHART_PAD} x2={CHART_PAD} y2={chartHeight() - CHART_PAD} />
      {#each chartXTicks() as tick, ti}
        <text class="sim-chart-axis-text" x={chartX(tick)} y={chartHeight() - 8} text-anchor="middle" fill={colorValue(props.AxesTextColor, '#000000')}>{chartXTickText(tick, ti)}</text>
      {/each}
      {#each chartYTicks() as tick}
        <text class="sim-chart-axis-text" x={CHART_PAD - 5} y={chartY(tick) + 3} text-anchor="end" fill={colorValue(props.AxesTextColor, '#000000')}>{chartTickText(tick)}</text>
      {/each}
    {/if}
    {#each chartDataSeries() as series, si}
      {#if numberOr(props.Type, 0) === 4}
        {#each pieSectors(series.points, si) as sector, pi}
          <path
            role="button"
            tabindex="0"
            d={sector.d}
            fill={sector.fill}
            stroke="white"
            stroke-width="1"
            on:click={() => chartEntryClick(series, series.points[pi])}
            on:keydown={(e) => e.key === 'Enter' && chartEntryClick(series, series.points[pi])}
          />
        {/each}
      {:else if numberOr(props.Type, 0) === 3}
        {#each series.points as pt, pi}
          <rect
            role="button"
            tabindex="0"
            x={chartBarX(si, pi, chartDataSeries().length, series.points.length)}
            y={chartY(pt[1])}
            width={chartBarWidth(chartDataSeries().length, series.points.length)}
            height={chartHeight() - CHART_PAD - chartY(pt[1])}
            fill={chartPointColor(series, pi)}
            on:click={() => chartEntryClick(series, pt)}
            on:keydown={(e) => e.key === 'Enter' && chartEntryClick(series, pt)}
          />
          <text class="sim-chart-data-label" x={chartBarX(si, pi, chartDataSeries().length, series.points.length) + chartBarWidth(chartDataSeries().length, series.points.length) / 2} y={chartY(pt[1]) - 4} text-anchor="middle" fill={series.dataLabelColor}>{pt[1]}</text>
        {/each}
      {:else}
        {#if numberOr(props.Type, 0) === 2}
          <path d={chartAreaPath(series.points)} fill={series.color} opacity="0.3" />
        {/if}
        {#if numberOr(props.Type, 0) !== 1}
          <polyline points={chartPolylinePoints(series.points)} fill="none" stroke={series.color} stroke-width="2" />
        {/if}
        {#each series.points as pt, pi}
          <circle
            role="button"
            tabindex="0"
            cx={chartX(pt[0])}
            cy={chartY(pt[1])}
            r="4"
            fill={chartPointColor(series, pi)}
            stroke={series.highlightColor || 'none'}
            stroke-width={series.highlightColor ? 2 : 0}
            on:click={() => chartEntryClick(series, pt)}
            on:keydown={(e) => e.key === 'Enter' && chartEntryClick(series, pt)}
          />
          <text class="sim-chart-data-label" x={chartX(pt[0])} y={chartY(pt[1]) - 7} text-anchor="middle" fill={series.dataLabelColor}>{pt[1]}</text>
        {/each}
      {/if}
    {/each}
    {#each chartTrendlines() as trendline}
      <path d={trendline.d} fill="none" stroke={trendline.color} stroke-width={trendline.width} stroke-dasharray={trendline.dash} />
    {/each}
  </svg>
  {#if props.Description}
    <div class="sim-chart-description">{props.Description}</div>
  {/if}
  {#if boolValue(props.LegendEnabled, true) && chartDataSeries().length > 0}
    <div class="sim-chart-legend">
      {#each chartDataSeries() as series}
        <span><span class="sim-chart-legend-dot" style="background:{series.color}"></span>{series.label}</span>
      {/each}
    </div>
  {/if}
</div>
