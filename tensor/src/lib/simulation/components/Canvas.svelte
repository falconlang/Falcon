<script>
  import { createEventDispatcher, onDestroy } from 'svelte';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, boolValue, colorValue, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let state = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let enabled = true;
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  const EDGE_NORTH = 1;
  const EDGE_NORTHEAST = 2;
  const EDGE_EAST = 3;
  const EDGE_SOUTHEAST = 4;
  const EDGE_SOUTH = -1;
  const EDGE_SOUTHWEST = -2;
  const EDGE_WEST = -3;
  const EDGE_NORTHWEST = -4;
  const FLING_MIN_SPEED_PX_PER_MS = 1;
  const FINGER_WIDTH_PX = 24;
  const FINGER_HEIGHT_PX = 24;

  let canvasEl;
  let canvasCtx = null;
  let canvasDragStart = null;
  let canvasPrev = null;
  let canvasTouchStart = null;
  let canvasTouchTime = null;
  let canvasTouchedSprites = [];
  let canvasDrawOps = [];
  let lastCanvasActionSeq = 0;
  let canvasImageCache = new Map();
  let spriteAnimationFrame = null;
  let spriteLastTick = {};
  let spriteCollisionPairs = new Set();
  let canvasDraggedSprites = [];
  let canvasIsDrag = false;
  let canvasPointerHistory = [];
  let lastCanvasBackgroundKey = null;
  let canvasBackgroundSignature = '';
  let _lastCanvasRenderKey = '';

  $: assetUrl = resolveAssetUrl(assets, assetName(props));
  $: context = styleContext(node, props, parentType, assetUrl);
  $: if (canvasEl) {
    const _key = JSON.stringify([props, ...(node.children ?? []).map(c => state?.[c.name])]);
    if (_key !== _lastCanvasRenderKey) {
      _lastCanvasRenderKey = _key;
      applyCanvasOps();
    }
  }
  $: updateSpriteAnimationLoop();
  $: canvasBackgroundSignature = [
    props.BackgroundColor ?? '',
    props.BackgroundImage ?? '',
    props.BackgroundImageinBase64 ?? '',
  ].join('\0');
  $: clearCanvasDrawingLayerOnBackgroundChange(canvasBackgroundSignature);
  $: handleCanvasActionState(actions?.[node?.name] ?? {});

  onDestroy(() => {
    stopSpriteAnimationLoop();
  });

  function canvasWidth() {
    const v = props.Width;
    if (v === -2 || v === undefined || v === null) return 360;
    if (v === -1 || v === '') return 360;
    return Math.max(1, Number(v) || 360);
  }

  function canvasHeight() {
    const v = props.Height;
    if (v === -1 || v === undefined || v === null || v === '') return 300;
    if (v === -2) return canvasEl?.parentElement?.clientHeight || 300;
    return Math.max(1, Number(v) || 300);
  }

  function canvasSprites() {
    return (node?.children || []).filter(c => c.type === 'Ball' || c.type === 'ImageSprite');
  }

  function getCanvas() {
    if (!canvasEl) return null;
    if (!canvasCtx) canvasCtx = canvasEl.getContext('2d');
    return canvasCtx;
  }

  function canvasPoint(e) {
    if (!canvasEl) return { x: 0, y: 0 };
    const rect = canvasEl.getBoundingClientRect();
    const scaleX = canvasEl.width / rect.width;
    const scaleY = canvasEl.height / rect.height;
    return {
      x: Math.max(0, Math.round((e.clientX - rect.left) * scaleX)),
      y: Math.max(0, Math.round((e.clientY - rect.top) * scaleY)),
    };
  }

  function unitValue(value, fallback = 0) {
    const n = numberOr(value, fallback);
    return Math.max(0, Math.min(1, n));
  }

  function spriteImageElement(sprite) {
    if (sprite.type !== 'ImageSprite') return null;
    const sp = state?.[sprite.name] ?? {};
    return cachedCanvasImage(resolveAssetUrl(assets, sp.Picture));
  }

  function spriteLength(value, natural = 0) {
    const n = numberOr(value, -1);
    if (n < 0) return Math.max(0, numberOr(natural, 0));
    return Math.max(0, n);
  }

  function spriteGeometry(sprite, override = null) {
    const sp = state?.[sprite.name] ?? {};
    const originX = numberOr(override?.X ?? sp.X, 0);
    const originY = numberOr(override?.Y ?? sp.Y, 0);
    if (sprite.type === 'Ball') {
      const r = Math.max(0, numberOr(sp.Radius, 5));
      const u = boolValue(sp.OriginAtCenter, false) ? 0.5 : 0;
      const v = boolValue(sp.OriginAtCenter, false) ? 0.5 : 0;
      const width = r * 2;
      const height = r * 2;
      const x = originX - width * u;
      const y = originY - height * v;
      return {
        x,
        y,
        width,
        height,
        cx: x + r,
        cy: y + r,
        originX,
        originY,
        u,
        v,
        radius: r,
      };
    }
    const img = spriteImageElement(sprite);
    const width = spriteLength(sp.Width, img?.naturalWidth || 0);
    const height = spriteLength(sp.Height, img?.naturalHeight || 0);
    const u = unitValue(sp.OriginX, 0);
    const v = unitValue(sp.OriginY, 0);
    const x = originX - width * u;
    const y = originY - height * v;
    return { x, y, width, height, cx: x + width / 2, cy: y + height / 2, originX, originY, u, v, radius: 0 };
  }

  function canvasTouchRect(pt) {
    const halfWidth = FINGER_WIDTH_PX / 2;
    const halfHeight = FINGER_HEIGHT_PX / 2;
    return {
      left: Math.max(0, pt.x - halfWidth),
      top: Math.max(0, pt.y - halfHeight),
      right: Math.min(canvasWidth() - 1, pt.x + halfWidth),
      bottom: Math.min(canvasHeight() - 1, pt.y + halfHeight),
    };
  }

  function rectsIntersect(a, b) {
    return a.left <= b.right && a.right >= b.left && a.top <= b.bottom && a.bottom >= b.top;
  }

  function spriteBoundingRect(sprite) {
    const geom = spriteGeometry(sprite);
    return {
      left: geom.x,
      top: geom.y,
      right: geom.x + Math.max(0, geom.width) - 1,
      bottom: geom.y + Math.max(0, geom.height) - 1,
      geom,
    };
  }

  function ballIntersectsRect(geom, rect) {
    const x = Math.max(rect.left, Math.min(geom.cx, rect.right));
    const y = Math.max(rect.top, Math.min(geom.cy, rect.bottom));
    return Math.hypot(x - geom.cx, y - geom.cy) <= geom.radius;
  }

  function spriteIntersectsTouchRect(sprite, rect) {
    const sp = state?.[sprite.name] ?? {};
    if (sp.Visible === false || sp.Enabled === false) return false;
    const bounds = spriteBoundingRect(sprite);
    if (bounds.right < bounds.left || bounds.bottom < bounds.top || !rectsIntersect(bounds, rect)) return false;
    if (sprite.type === 'Ball') return ballIntersectsRect(bounds.geom, rect);
    return true;
  }

  function hitSpritesAt(pt) {
    const rect = canvasTouchRect(pt);
    return canvasSprites()
      .filter(sprite => spriteIntersectsTouchRect(sprite, rect))
      .sort((a, b) => numberOr(state?.[b.name]?.Z, 1) - numberOr(state?.[a.name]?.Z, 1));
  }

  function uniqueSprites(primary = [], fallback = []) {
    const seen = new Set();
    const out = [];
    for (const sprite of [...primary, ...fallback]) {
      if (!sprite?.name || seen.has(sprite.name)) continue;
      seen.add(sprite.name);
      out.push(sprite);
    }
    return out;
  }

  async function canvasPointerDown(e) {
    if (!enabled) return;
    canvasEl?.setPointerCapture(e.pointerId);
    const pt = canvasPoint(e);
    canvasTouchedSprites = hitSpritesAt(pt);
    canvasDraggedSprites = [...canvasTouchedSprites];
    canvasDragStart = pt;
    canvasPrev = pt;
    canvasTouchStart = pt;
    canvasTouchTime = Date.now();
    for (const sprite of canvasTouchedSprites) await emitEvent(dispatch, eventRunner, sprite.name, 'TouchDown', [pt.x, pt.y]);
    await emitEvent(dispatch, eventRunner, node.name, 'TouchDown', [pt.x, pt.y]);
    canvasIsDrag = false;
    canvasPointerHistory = [{ ...pt, t: canvasTouchTime }];
  }

  async function canvasPointerMove(e) {
    if (!enabled || !canvasDragStart) return;
    const pt = canvasPoint(e);
    const threshold = numberOr(props.TapThreshold, 15);
    if (!canvasIsDrag
      && Math.abs(pt.x - canvasDragStart.x) < threshold
      && Math.abs(pt.y - canvasDragStart.y) < threshold) {
      return;
    }
    canvasIsDrag = true;
    if (!boolValue(props.ExtendMovesOutsideCanvas, false)
      && (pt.x < 0 || pt.x >= canvasWidth() || pt.y < 0 || pt.y >= canvasHeight())) {
      return;
    }

    canvasDraggedSprites = uniqueSprites(canvasDraggedSprites, hitSpritesAt(pt));
    let handled = false;
    for (const sprite of canvasDraggedSprites) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false || sp.Enabled === false) continue;
      handled = true;
      await emitEvent(dispatch, eventRunner, sprite.name, 'Dragged', [
        canvasDragStart.x, canvasDragStart.y,
        canvasPrev.x, canvasPrev.y,
        pt.x, pt.y,
      ]);
    }
    await emitEvent(dispatch, eventRunner, node.name, 'Dragged', [
      canvasDragStart.x, canvasDragStart.y,
      canvasPrev.x, canvasPrev.y,
      pt.x, pt.y,
      handled,
    ]);
    canvasPrev = pt;
    canvasPointerHistory = [...canvasPointerHistory, { ...pt, t: Date.now() }].slice(-8);
  }

  async function canvasPointerUp(e) {
    if (!enabled) return;
    const pt = canvasPoint(e);
    const upTime = Date.now();
    canvasPointerHistory = [...canvasPointerHistory, { ...pt, t: upTime }].slice(-8);
    let handled = false;
    for (const sprite of canvasDraggedSprites) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false || sp.Enabled === false) continue;
      await emitEvent(dispatch, eventRunner, sprite.name, 'Touched', [pt.x, pt.y]);
      await emitEvent(dispatch, eventRunner, sprite.name, 'TouchUp', [pt.x, pt.y]);
      handled = true;
    }
    if (!canvasIsDrag) await emitEvent(dispatch, eventRunner, node.name, 'Touched', [pt.x, pt.y, handled]);
    await emitEvent(dispatch, eventRunner, node.name, 'TouchUp', [pt.x, pt.y]);
    await maybeEmitCanvasFling(pt, upTime);
    resetCanvasPointerState();
  }

  function canvasPointerCancel() {
    resetCanvasPointerState();
  }

  async function maybeEmitCanvasFling(pt, upTime) {
    if (!canvasTouchStart || canvasPointerHistory.length < 2) return;
    const recent = [...canvasPointerHistory].reverse().find(point => upTime - point.t >= 16 && upTime - point.t <= 140)
      || canvasPointerHistory[0];
    const dt = Math.max(1, upTime - recent.t);
    const vx = (pt.x - recent.x) / dt;
    const vy = (pt.y - recent.y) / dt;
    const speed = Math.hypot(vx, vy);
    const totalDist = Math.hypot(pt.x - canvasTouchStart.x, pt.y - canvasTouchStart.y);
    if (speed < FLING_MIN_SPEED_PX_PER_MS || totalDist < numberOr(props.TapThreshold, 15)) return;
    const heading = -Math.atan2(vy, vx) * 180 / Math.PI;
    let handled = false;
    for (const sprite of canvasTouchedSprites) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false || sp.Enabled === false) continue;
      await emitEvent(dispatch, eventRunner, sprite.name, 'Flung', [canvasTouchStart.x, canvasTouchStart.y, speed, heading, vx, vy]);
      handled = true;
    }
    await emitEvent(dispatch, eventRunner, node.name, 'Flung', [canvasTouchStart.x, canvasTouchStart.y, speed, heading, vx, vy, handled]);
  }

  function resetCanvasPointerState() {
    canvasDragStart = null;
    canvasPrev = null;
    canvasTouchStart = null;
    canvasTouchTime = null;
    canvasTouchedSprites = [];
    canvasDraggedSprites = [];
    canvasIsDrag = false;
    canvasPointerHistory = [];
  }

  function applyCanvasOps() {
    const ctx = getCanvas();
    if (!ctx || !canvasEl) return;
    ctx.clearRect(0, 0, canvasEl.width, canvasEl.height);
    const bg = colorValue(props.BackgroundColor, '#ffffff');
    if (bg && bg !== 'transparent') {
      ctx.fillStyle = bg;
      ctx.fillRect(0, 0, canvasEl.width, canvasEl.height);
    }

    const bgImage = cachedCanvasImage(canvasBackgroundUrl());
    if (bgImage) ctx.drawImage(bgImage, 0, 0, canvasEl.width, canvasEl.height);

    for (const op of canvasDrawOps) drawCanvasOp(ctx, op);

    for (const sprite of canvasSprites().slice().sort((a, b) => numberOr(state?.[a.name]?.Z, 1) - numberOr(state?.[b.name]?.Z, 1))) {
      const sp = state?.[sprite.name] ?? {};
      if (sp.Visible === false) continue;
      ctx.save();
      if (sprite.type === 'Ball') drawBall(ctx, sprite, sp);
      else if (sprite.type === 'ImageSprite') drawImageSprite(ctx, sprite, sp);
      ctx.restore();
    }
  }

  function canvasBackgroundUrl() {
    const base64 = String(props.BackgroundImageinBase64 ?? '').trim();
    if (base64) return base64.startsWith('data:') ? base64 : `data:image/png;base64,${base64}`;
    return props.BackgroundImage ? assetUrl : '';
  }

  function canvasBackgroundKey() {
    return [
      props.BackgroundColor ?? '',
      props.BackgroundImage ?? '',
      props.BackgroundImageinBase64 ?? '',
    ].join('\0');
  }

  function clearCanvasDrawingLayerOnBackgroundChange(key = canvasBackgroundKey()) {
    if (lastCanvasBackgroundKey === null) {
      lastCanvasBackgroundKey = key;
      return;
    }
    if (key === lastCanvasBackgroundKey) return;
    lastCanvasBackgroundKey = key;
    if (!canvasDrawOps.length) return;
    canvasDrawOps = [];
    applyCanvasOps();
  }

  function cachedCanvasImage(url) {
    if (!url) return null;
    let entry = canvasImageCache.get(url);
    if (!entry) {
      const img = new Image();
      entry = { img };
      canvasImageCache.set(url, entry);
      img.onload = () => applyCanvasOps();
      img.onerror = () => canvasImageCache.delete(url);
      img.src = url;
    }
    return entry.img.complete && entry.img.naturalWidth > 0 ? entry.img : null;
  }

  function drawBall(ctx, sprite, sp) {
    const geom = spriteGeometry(sprite);
    ctx.fillStyle = colorValue(sp.PaintColor, '#000000');
    ctx.beginPath();
    ctx.arc(geom.cx, geom.cy, geom.radius, 0, Math.PI * 2);
    ctx.fill();
  }

  function drawImageSprite(ctx, sprite, sp) {
    const img = spriteImageElement(sprite);
    const geom = spriteGeometry(sprite);
    if (!img || geom.width <= 0 || geom.height <= 0) return;
    if (boolValue(sp.Rotates, true)) {
      const angle = numberOr(sp.Heading, 0) * Math.PI / 180;
      ctx.translate(geom.originX, geom.originY);
      ctx.rotate(-angle);
      ctx.drawImage(img, geom.x - geom.originX, geom.y - geom.originY, geom.width, geom.height);
    } else {
      ctx.drawImage(img, geom.x, geom.y, geom.width, geom.height);
    }
  }

  function handleCanvasActionState(actionState) {
    const ordered = Array.isArray(actionState?.__actions) ? actionState.__actions : [];
    const nextActions = ordered.filter(action => (
      action.seq > lastCanvasActionSeq
      && (action.action === 'canvas-draw' || action.action === 'canvas-clear')
    ));
    if (!nextActions.length) return;

    let nextOps = canvasDrawOps;
    for (const action of nextActions) {
      if (action.action === 'canvas-clear') nextOps = [];
      else if (action.action === 'canvas-draw') nextOps = [...nextOps, action];
      lastCanvasActionSeq = Math.max(lastCanvasActionSeq, action.seq);
    }
    canvasDrawOps = nextOps;
    applyCanvasOps();
  }

  function drawCanvasOp(ctx, op) {
    const color = colorValue(op.color ?? props.PaintColor, '#000000');
    const lineWidth = Math.max(1, numberOr(op.lineWidth ?? props.LineWidth, 2));
    ctx.lineWidth = lineWidth;
    ctx.strokeStyle = color;
    ctx.fillStyle = color;
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';

    switch (op.op) {
      case 'line':
        ctx.beginPath();
        ctx.moveTo(numberOr(op.x1, 0), numberOr(op.y1, 0));
        ctx.lineTo(numberOr(op.x2, 0), numberOr(op.y2, 0));
        ctx.stroke();
        break;
      case 'circle':
        ctx.beginPath();
        ctx.arc(numberOr(op.cx, 0), numberOr(op.cy, 0), Math.max(0, numberOr(op.r, 0)), 0, Math.PI * 2);
        if (boolValue(op.fill, false)) ctx.fill();
        else ctx.stroke();
        break;
      case 'point': {
        const size = Math.max(1, lineWidth);
        ctx.fillRect(numberOr(op.x, 0) - size / 2, numberOr(op.y, 0) - size / 2, size, size);
        break;
      }
      case 'text':
        drawCanvasText(ctx, op, 0);
        break;
      case 'textAngle':
        drawCanvasText(ctx, op, -numberOr(op.angle, 0));
        break;
      case 'arc':
        drawCanvasArc(ctx, op);
        break;
      case 'shape':
        drawCanvasShape(ctx, op);
        break;
      default:
        break;
    }
  }

  function drawCanvasText(ctx, op, angle = 0) {
    const x = numberOr(op.x, 0);
    const y = numberOr(op.y, 0);
    ctx.save();
    ctx.translate(x, y);
    if (angle) ctx.rotate(angle * Math.PI / 180);
    ctx.font = `${Math.max(1, numberOr(op.fontSize ?? props.FontSize, 14))}px sans-serif`;
    ctx.textAlign = canvasTextAlign();
    ctx.textBaseline = 'alphabetic';
    ctx.fillText(String(op.text ?? ''), 0, 0);
    ctx.restore();
  }

  function canvasTextAlign() {
    const align = numberOr(props.TextAlignment, 1);
    if (align === 0) return 'left';
    if (align === 2) return 'right';
    return 'center';
  }

  function drawCanvasArc(ctx, op) {
    const left = numberOr(op.left, 0);
    const top = numberOr(op.top, 0);
    const right = numberOr(op.right, left);
    const bottom = numberOr(op.bottom, top);
    const cx = (left + right) / 2;
    const cy = (top + bottom) / 2;
    const rx = Math.max(0, Math.abs(right - left) / 2);
    const ry = Math.max(0, Math.abs(bottom - top) / 2);
    const start = numberOr(op.startAngle, 0) * Math.PI / 180;
    const sweep = numberOr(op.sweepAngle, 0) * Math.PI / 180;
    ctx.beginPath();
    if (boolValue(op.useCenter, false)) ctx.moveTo(cx, cy);
    ctx.ellipse(cx, cy, rx, ry, 0, start, start + sweep, sweep < 0);
    if (boolValue(op.useCenter, false)) ctx.closePath();
    if (boolValue(op.fill, false)) ctx.fill();
    else ctx.stroke();
  }

  function drawCanvasShape(ctx, op) {
    const pts = normalizeCanvasPoints(op.points);
    if (!pts.length) return;
    ctx.beginPath();
    ctx.moveTo(pts[0].x, pts[0].y);
    for (const pt of pts.slice(1)) ctx.lineTo(pt.x, pt.y);
    ctx.closePath();
    if (boolValue(op.fill, false)) ctx.fill();
    else ctx.stroke();
  }

  function normalizeCanvasPoints(value) {
    if (!Array.isArray(value)) return [];
    if (value.every(item => !Array.isArray(item) && typeof item !== 'object')) {
      const out = [];
      for (let i = 0; i + 1 < value.length; i += 2) {
        out.push({ x: numberOr(value[i], 0), y: numberOr(value[i + 1], 0) });
      }
      return out;
    }
    return value
      .map(item => {
        if (Array.isArray(item)) return { x: numberOr(item[0], 0), y: numberOr(item[1], 0) };
        if (item && typeof item === 'object') return { x: numberOr(item.x ?? item.X, 0), y: numberOr(item.y ?? item.Y, 0) };
        return null;
      })
      .filter(Boolean);
  }

  function movingCanvasSprites() {
    return canvasSprites().filter(sprite => {
      const sp = state?.[sprite.name] ?? {};
      return sp.Visible !== false && sp.Enabled !== false && numberOr(sp.Speed, 0) > 0;
    });
  }

  function updateSpriteAnimationLoop() {
    if (!canvasEl) return;
    if (movingCanvasSprites().length > 0) {
      if (spriteAnimationFrame == null) spriteAnimationFrame = requestAnimationFrame(spriteAnimationTick);
    } else {
      stopSpriteAnimationLoop();
    }
  }

  function stopSpriteAnimationLoop() {
    if (spriteAnimationFrame != null) cancelAnimationFrame(spriteAnimationFrame);
    spriteAnimationFrame = null;
    spriteLastTick = {};
  }

  function spriteAnimationTick(timestamp) {
    spriteAnimationFrame = null;
    animateCanvasSprites(timestamp);
    detectSpriteCollisions();
    if (movingCanvasSprites().length > 0) spriteAnimationFrame = requestAnimationFrame(spriteAnimationTick);
    else stopSpriteAnimationLoop();
  }

  function animateCanvasSprites(timestamp) {
    const batchPatches = [];
    for (const sprite of movingCanvasSprites()) {
      const sp = state?.[sprite.name] ?? {};
      const interval = Math.max(1, numberOr(sp.Interval, 100));
      const last = spriteLastTick[sprite.name];
      if (last == null) {
        spriteLastTick[sprite.name] = timestamp;
        continue;
      }
      const elapsed = timestamp - last;
      if (elapsed < interval) continue;
      const steps = Math.max(1, Math.floor(elapsed / interval));
      spriteLastTick[sprite.name] = last + steps * interval;

      const heading = numberOr(sp.Heading, 0) * Math.PI / 180;
      const speed = numberOr(sp.Speed, 0) * steps;
      const next = clampSpritePosition(
        sprite,
        numberOr(sp.X, 0) + speed * Math.cos(heading),
        numberOr(sp.Y, 0) - speed * Math.sin(heading),
      );
      const patches = [
        { component: sprite.name, property: 'X', value: next.x },
        { component: sprite.name, property: 'Y', value: next.y },
      ];
      if (next.edge) {
        emitInteraction(dispatch, patches, { component: sprite.name, event: 'EdgeReached', args: [next.edge] });
      } else {
        batchPatches.push(...patches);
      }
    }
    if (batchPatches.length) emitInteraction(dispatch, batchPatches, null);
  }

  function clampSpritePosition(sprite, x, y) {
    const geom = spriteGeometry(sprite, { X: x, Y: y });
    const over = {
      west: geom.x < 0,
      north: geom.y < 0,
      east: geom.x + geom.width > canvasWidth(),
      south: geom.y + geom.height > canvasHeight(),
    };
    const edge = edgeDirection(over);
    if (!edge) return { x, y, edge: 0 };

    const moved = moveSpriteGeometryIntoCanvas(sprite, geom);
    return { ...moved, edge };
  }

  function edgeDirection({ west, north, east, south }) {
    if (west) {
      if (north) return EDGE_NORTHWEST;
      if (south) return EDGE_SOUTHWEST;
      return EDGE_WEST;
    }
    if (east) {
      if (north) return EDGE_NORTHEAST;
      if (south) return EDGE_SOUTHEAST;
      return EDGE_EAST;
    }
    if (north) return EDGE_NORTH;
    if (south) return EDGE_SOUTH;
    return 0;
  }

  function moveSpriteGeometryIntoCanvas(sprite, geom) {
    const width = canvasWidth();
    const height = canvasHeight();
    let left = geom.x;
    let top = geom.y;
    if (geom.width > width) left = 0;
    else if (geom.x < 0) left = 0;
    else if (geom.x + geom.width > width) left = width - geom.width;

    if (geom.height > height) top = 0;
    else if (geom.y < 0) top = 0;
    else if (geom.y + geom.height > height) top = height - geom.height;

    return {
      x: left + geom.width * geom.u,
      y: top + geom.height * geom.v,
    };
  }

  function detectSpriteCollisions() {
    const sprites = canvasSprites().filter(sprite => {
      const sp = state?.[sprite.name] ?? {};
      return sp.Visible !== false && sp.Enabled !== false;
    });
    const current = new Set();
    for (let i = 0; i < sprites.length; i += 1) {
      for (let j = i + 1; j < sprites.length; j += 1) {
        if (!spritesOverlap(sprites[i], sprites[j])) continue;
        current.add(collisionKey(sprites[i], sprites[j]));
      }
    }
    for (const key of current) {
      if (spriteCollisionPairs.has(key)) continue;
      const [a, b] = key.split('|');
      emitEvent(dispatch, eventRunner, a, 'CollidedWith', [b]);
      emitEvent(dispatch, eventRunner, b, 'CollidedWith', [a]);
    }
    for (const key of spriteCollisionPairs) {
      if (current.has(key)) continue;
      const [a, b] = key.split('|');
      emitEvent(dispatch, eventRunner, a, 'NoLongerCollidingWith', [b]);
      emitEvent(dispatch, eventRunner, b, 'NoLongerCollidingWith', [a]);
    }
    spriteCollisionPairs = current;
  }

  function collisionKey(a, b) {
    return [a.name, b.name].sort().join('|');
  }

  function spritesOverlap(a, b) {
    const ga = spriteGeometry(a);
    const gb = spriteGeometry(b);
    if (a.type === 'Ball' && b.type === 'Ball') {
      const dx = ga.cx - gb.cx;
      const dy = ga.cy - gb.cy;
      const sumR = ga.radius + gb.radius;
      return dx * dx + dy * dy <= sumR * sumR;
    }
    if (a.type === 'Ball') return ballIntersectsRect(ga, spriteBoundingRect(b));
    if (b.type === 'Ball') return ballIntersectsRect(gb, spriteBoundingRect(a));
    return ga.x < gb.x + gb.width
      && ga.x + ga.width > gb.x
      && ga.y < gb.y + gb.height
      && ga.y + ga.height > gb.y;
  }
</script>

<div
  class="sim-canvas-wrap"
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(context)}
  data-sim-component={node.name}
>
  <canvas
    bind:this={canvasEl}
    class="sim-canvas"
    width={canvasWidth()}
    height={canvasHeight()}
    on:pointerdown={canvasPointerDown}
    on:pointermove={canvasPointerMove}
    on:pointerup={canvasPointerUp}
    on:pointercancel={canvasPointerCancel}
  ></canvas>
  <slot />
</div>
