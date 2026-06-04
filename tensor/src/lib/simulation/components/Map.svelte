<script>
  import { createEventDispatcher, onDestroy } from 'svelte';
  import 'leaflet/dist/leaflet.css';
  import { emitEvent, emitInteraction } from '../events.js';
  import { assetName, baseStyle, boolValue, colorValue, numberOr, styleContext } from '../style.js';
  import { resolveAssetUrl } from '../../simulation-capabilities.js';

  export let node;
  export let props = {};
  export let state = {};
  export let assets = [];
  export let parentType = '';
  export let unsupportedHere = false;
  export let actions = {};
  export let eventRunner = null;

  const dispatch = createEventDispatcher();
  let mapEl;
  let mapInstance = null;
  let mapTileLayer = null;
  let mapZoomControl = null;
  let mapScaleControl = null;
  let mapCompassControl = null;
  let mapLayers = {};
  let mapFeatureActionSeq = {};
  let leafletLib = null;

  $: context = styleContext(node, props, parentType, resolveAssetUrl(assets, assetName(props)));
  $: if (mapEl) initOrUpdateMap();
  $: if (mapInstance && leafletLib && props) {
    mapInstance.setView(
      [numberOr(props.Latitude, 42.359144), numberOr(props.Longitude, -71.093612)],
      numberOr(props.ZoomLevel, 13),
      { animate: false },
    );
    updateMapTileLayer(leafletLib);
    updateMapInteractions();
    updateMapControls(leafletLib);
    fitMapBoundsIfNeeded(leafletLib);
    updateMapFeatures(leafletLib);
  }
  $: if (mapInstance) handleMapFeatureActions();

  onDestroy(() => {
    if (mapInstance) {
      mapInstance.remove();
      mapInstance = null;
    }
  });

  async function initOrUpdateMap() {
    const L = await import('leaflet').catch(() => null);
    if (!L || !mapEl) return;
    if (!mapInstance) {
      delete L.Icon.Default.prototype._getIconUrl;
      L.Icon.Default.mergeOptions({
        iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
        iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
        shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
      });
      mapInstance = L.map(mapEl, {
        center: [numberOr(props.Latitude, 42.359144), numberOr(props.Longitude, -71.093612)],
        zoom: numberOr(props.ZoomLevel, 13),
        zoomControl: false,
        dragging: boolValue(props.EnablePan, true),
        scrollWheelZoom: boolValue(props.EnableZoom, true),
        touchZoom: boolValue(props.EnableZoom, true),
        doubleClickZoom: boolValue(props.EnableZoom, true),
        attributionControl: true,
      });
      updateMapTileLayer(L);
      mapInstance.on('moveend', () => emitEvent(dispatch, eventRunner, node.name, 'BoundsChange'));
      mapInstance.on('zoomend', () => emitEvent(dispatch, eventRunner, node.name, 'ZoomChange'));
      mapInstance.on('click', (e) => emitEvent(dispatch, eventRunner, node.name, 'TapAtPoint', [e.latlng.lat, e.latlng.lng]));
      mapInstance.on('dblclick', (e) => emitEvent(dispatch, eventRunner, node.name, 'DoubleTapAtPoint', [e.latlng.lat, e.latlng.lng]));
      mapInstance.on('contextmenu', (e) => emitEvent(dispatch, eventRunner, node.name, 'LongPressAtPoint', [e.latlng.lat, e.latlng.lng]));
      leafletLib = L;
      emitEvent(dispatch, eventRunner, node.name, 'Ready');
    }
  }

  function mapTileConfig() {
    const custom = String(props.CustomUrl ?? '').trim();
    const defaultUrl = 'https://tile.openstreetmap.org/{z}/{x}/{y}.png';
    if (custom && custom !== defaultUrl) return { url: custom, attribution: '' };
    switch (numberOr(props.MapType, 1)) {
      case 2:
        return {
          url: 'https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png',
          attribution: '&copy; OpenTopoMap contributors',
        };
      case 3:
        return {
          url: 'https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png',
          attribution: '&copy; OpenStreetMap contributors &copy; CARTO',
        };
      default:
        return {
          url: defaultUrl,
          attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
        };
    }
  }

  function updateMapTileLayer(L) {
    if (!mapInstance || !L) return;
    const config = mapTileConfig();
    if (mapTileLayer?._simUrl === config.url) return;
    if (mapTileLayer) mapTileLayer.remove();
    mapTileLayer = L.tileLayer(config.url, {
      attribution: config.attribution,
      maxZoom: numberOr(props.MapType, 1) === 2 ? 17 : 19,
    });
    mapTileLayer._simUrl = config.url;
    mapTileLayer.addTo(mapInstance);
  }

  function updateMapInteractions() {
    if (!mapInstance) return;
    const pan = boolValue(props.EnablePan, true);
    const zoom = boolValue(props.EnableZoom, true);
    if (pan) mapInstance.dragging?.enable();
    else mapInstance.dragging?.disable();
    for (const handler of ['scrollWheelZoom', 'touchZoom', 'doubleClickZoom', 'boxZoom', 'keyboard']) {
      if (zoom) mapInstance[handler]?.enable?.();
      else mapInstance[handler]?.disable?.();
    }
  }

  function updateMapControls(L) {
    if (!mapInstance || !L) return;
    if (boolValue(props.ShowZoom, false)) {
      if (!mapZoomControl) mapZoomControl = L.control.zoom({ position: 'topleft' }).addTo(mapInstance);
    } else if (mapZoomControl) {
      mapZoomControl.remove();
      mapZoomControl = null;
    }

    if (boolValue(props.ShowScale, false)) {
      if (!mapScaleControl) mapScaleControl = L.control.scale({ position: 'bottomleft', metric: true, imperial: true }).addTo(mapInstance);
    } else if (mapScaleControl) {
      mapScaleControl.remove();
      mapScaleControl = null;
    }

    if (boolValue(props.ShowCompass, false)) {
      if (!mapCompassControl) {
        const CompassControl = L.Control.extend({
          options: { position: 'topright' },
          onAdd() {
            const el = L.DomUtil.create('div', 'sim-map-compass');
            el.innerHTML = '<span>N</span>';
            return el;
          },
        });
        mapCompassControl = new CompassControl().addTo(mapInstance);
      }
      const el = mapCompassControl.getContainer?.();
      if (el) el.style.setProperty('--sim-map-rotation', `${numberOr(props.Rotation, 0)}deg`);
    } else if (mapCompassControl) {
      mapCompassControl.remove();
      mapCompassControl = null;
    }
  }

  function fitMapBoundsIfNeeded(L) {
    if (!mapInstance || !L) return;
    const bounds = parseBoundingBox(props.BoundingBox);
    if (!bounds) return;
    mapInstance.fitBounds(bounds, { animate: false });
  }

  function updateMapFeatures(L) {
    if (!mapInstance || !L) return;
    const nextKeys = new Set();
    for (const { child, collection } of mapFeatureEntries()) {
      const sp = state?.[child.name] ?? {};
      if (sp.Visible === false) continue;
      const key = child.name;
      const collectionName = collection?.name || '';
      nextKeys.add(key);
      if (mapLayers[key]?._simType !== child.type || mapLayers[key]?._simCollectionName !== collectionName) {
        mapLayers[key]?.remove();
        delete mapLayers[key];
      }
      if (mapLayers[key]) {
        updateMapLayer(L, child, sp, mapLayers[key]);
      } else {
        mapLayers[key] = createMapLayer(L, child, sp, collection);
        if (mapLayers[key]) {
          mapLayers[key]._simType = child.type;
          mapLayers[key]._simCollectionName = collectionName;
          mapLayers[key].addTo(mapInstance);
        }
      }
    }
    for (const k of Object.keys(mapLayers)) {
      if (!nextKeys.has(k)) {
        mapLayers[k]?.remove();
        delete mapLayers[k];
      }
    }
  }

  function mapFeatureEntries(children = node?.children || [], collection = null) {
    const out = [];
    for (const child of children) {
      if (child.type === 'FeatureCollection') {
        const collectionState = state?.[child.name] ?? {};
        if (collectionState.Visible === false) continue;
        out.push(...mapFeatureEntries(child.children || [], child));
      } else if (['Marker', 'Circle', 'LineString', 'Polygon', 'Rectangle'].includes(child.type)) {
        out.push({ child, collection });
      }
    }
    return out;
  }

  function featureStyle(sp) {
    return {
      color: colorValue(sp.StrokeColor, '#000000'),
      opacity: numberOr(sp.StrokeOpacity, 1),
      weight: numberOr(sp.StrokeWidth, 1),
      fillColor: colorValue(sp.FillColor, '#ff0000'),
      fillOpacity: numberOr(sp.FillOpacity, 1),
    };
  }

  function createMapLayer(L, child, sp, collection = null) {
    switch (child.type) {
      case 'Marker': {
        const icon = markerIcon(L, sp);
        const marker = L.marker([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)], { icon, draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(marker, sp);
        bindMapFeatureEvents(L, marker, child, collection);
        marker.on('dragend', () => {
          const latlng = marker.getLatLng();
          emitInteraction(dispatch, [
            { component: child.name, property: 'Latitude', value: latlng.lat },
            { component: child.name, property: 'Longitude', value: latlng.lng },
          ], null);
        });
        return marker;
      }
      case 'Circle': {
        const circle = L.circle([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)], {
          radius: numberOr(sp.Radius, 10),
          ...featureStyle(sp),
          draggable: boolValue(sp.Draggable, false),
        });
        bindFeaturePopup(circle, sp);
        bindMapFeatureEvents(L, circle, child, collection);
        return circle;
      }
      case 'LineString': {
        const pts = featurePoints(sp);
        if (!pts.length) return null;
        const line = L.polyline(pts, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(line, sp);
        bindMapFeatureEvents(L, line, child, collection);
        return line;
      }
      case 'Polygon': {
        const pts = polygonLatLngs(sp);
        if (!pts.length) return null;
        const polygon = L.polygon(pts, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(polygon, sp);
        bindMapFeatureEvents(L, polygon, child, collection);
        return polygon;
      }
      case 'Rectangle': {
        const bounds = [
          [numberOr(sp.SouthLatitude, 0), numberOr(sp.WestLongitude, 0)],
          [numberOr(sp.NorthLatitude, 0), numberOr(sp.EastLongitude, 0)],
        ];
        const rectangle = L.rectangle(bounds, { ...featureStyle(sp), draggable: boolValue(sp.Draggable, false) });
        bindFeaturePopup(rectangle, sp);
        bindMapFeatureEvents(L, rectangle, child, collection);
        return rectangle;
      }
      default:
        return null;
    }
  }

  function updateMapLayer(L, child, sp, layer) {
    if (child.type === 'Marker') {
      layer.setLatLng([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)]);
      layer.setIcon(markerIcon(L, sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'Circle') {
      layer.setLatLng([numberOr(sp.Latitude, 0), numberOr(sp.Longitude, 0)]);
      layer.setRadius(numberOr(sp.Radius, 10));
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'LineString') {
      const pts = featurePoints(sp);
      if (pts.length) layer.setLatLngs(pts);
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'Polygon') {
      const pts = polygonLatLngs(sp);
      if (pts.length) layer.setLatLngs(pts);
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    } else if (child.type === 'Rectangle') {
      const bounds = [
        [numberOr(sp.SouthLatitude, 0), numberOr(sp.WestLongitude, 0)],
        [numberOr(sp.NorthLatitude, 0), numberOr(sp.EastLongitude, 0)],
      ];
      layer.setBounds(bounds);
      layer.setStyle(featureStyle(sp));
      bindFeaturePopup(layer, sp);
    }
  }

  function bindFeaturePopup(layer, sp) {
    const title = String(sp.Title ?? '');
    const description = String(sp.Description ?? '');
    if (!title && !description) {
      layer.unbindPopup?.();
      return;
    }
    layer.bindPopup(`<b>${escapeHtml(title)}</b><br>${escapeHtml(description)}`);
  }

  function escapeHtml(value) {
    return String(value ?? '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function markerIcon(L, sp) {
    const imageUrl = resolveAssetUrl(assets, sp.ImageAsset);
    if (imageUrl) {
      const size = [32, 32];
      return L.icon({
        iconUrl: imageUrl,
        iconSize: size,
        iconAnchor: markerAnchor(sp, size[0], size[1]),
        popupAnchor: [0, -size[1] / 2],
      });
    }
    const size = [24, 36];
    return L.divIcon({
      html: `<svg viewBox="0 0 24 36" xmlns="http://www.w3.org/2000/svg" width="24" height="36"><path d="M12 0C5.4 0 0 5.4 0 12c0 9 12 24 12 24s12-15 12-24c0-6.6-5.4-12-12-12z" fill="${colorValue(sp.FillColor, '#ff0000')}" stroke="${colorValue(sp.StrokeColor, '#000')}"/><circle cx="12" cy="12" r="5" fill="white"/></svg>`,
      className: '',
      iconSize: size,
      iconAnchor: markerAnchor(sp, size[0], size[1]),
      popupAnchor: [0, -size[1]],
    });
  }

  function markerAnchor(sp, width, height) {
    const horizontal = numberOr(sp.AnchorHorizontal, 3);
    const vertical = numberOr(sp.AnchorVertical, 3);
    const x = horizontal === 1 ? 0 : horizontal === 2 ? width : width / 2;
    const y = vertical === 1 ? 0 : vertical === 2 ? height / 2 : height;
    return [x, y];
  }

  function bindMapFeatureEvents(L, layer, child, collection = null) {
    layer.on('click', async (e) => {
      stopLeafletPropagation(L, e);
      await emitMapFeatureEvent(child, collection, 'Click', 'FeatureClick');
    });
    layer.on('contextmenu', async (e) => {
      stopLeafletPropagation(L, e);
      await emitMapFeatureEvent(child, collection, 'LongClick', 'FeatureLongClick');
    });
    layer.on('dragstart', () => emitMapFeatureEvent(child, collection, 'StartDrag', 'FeatureStartDrag'));
    layer.on('drag', () => emitMapFeatureEvent(child, collection, 'Drag', 'FeatureDrag'));
    layer.on('dragend', () => emitMapFeatureEvent(child, collection, 'StopDrag', 'FeatureStopDrag'));
  }

  async function emitMapFeatureEvent(child, collection, featureEvent, aggregateEvent) {
    await emitEvent(dispatch, eventRunner, child.name, featureEvent);
    if (collection?.name) await emitEvent(dispatch, eventRunner, collection.name, aggregateEvent, [child.name]);
    await emitEvent(dispatch, eventRunner, node.name, aggregateEvent, [child.name]);
  }

  function handleMapFeatureActions() {
    const nextSeq = { ...mapFeatureActionSeq };
    for (const { child } of mapFeatureEntries()) {
      const layer = mapLayers[child.name];
      if (!layer) continue;
      const actionState = actions?.[child.name] || {};
      for (const action of ['show-infobox', 'hide-infobox']) {
        const seq = numberOr(actionState[action], 0);
        const key = `${child.name}:${action}`;
        if (seq <= 0 || seq === nextSeq[key]) continue;
        nextSeq[key] = seq;
        if (action === 'show-infobox') layer.openPopup?.();
        else layer.closePopup?.();
      }
    }
    mapFeatureActionSeq = nextSeq;
  }

  function stopLeafletPropagation(L, e) {
    if (e?.originalEvent) L.DomEvent.stopPropagation(e.originalEvent);
  }

  function featurePoints(sp) {
    return parseLatLngList(sp.Points ?? sp.PointsFromString);
  }

  function polygonLatLngs(sp) {
    const outer = featurePoints(sp);
    if (!outer.length) return [];
    const holes = parseLatLngRings(sp.HolePoints ?? sp.HolePointsFromString);
    return holes.length ? [outer, ...holes] : outer;
  }

  function parseLatLngRings(value) {
    if (Array.isArray(value) && Array.isArray(value[0]) && Array.isArray(value[0][0])) {
      return value.map(ring => parseLatLngList(ring)).filter(ring => ring.length);
    }
    const text = String(value ?? '').trim();
    if (!text) return [];
    try {
      const parsed = JSON.parse(text);
      if (Array.isArray(parsed) && Array.isArray(parsed[0]) && Array.isArray(parsed[0][0])) {
        return parsed.map(ring => parseLatLngList(ring)).filter(ring => ring.length);
      }
      const single = parseLatLngList(parsed);
      return single.length ? [single] : [];
    } catch {
      const single = parseLatLngList(text);
      return single.length ? [single] : [];
    }
  }

  function parseLatLngList(value) {
    if (Array.isArray(value)) {
      if (value.every(point => !Array.isArray(point) && typeof point !== 'object')) {
        const nums = value.map(Number).filter(Number.isFinite);
        const pts = [];
        for (let i = 0; i + 1 < nums.length; i += 2) pts.push([nums[i], nums[i + 1]]);
        return pts;
      }
      return value
        .map(point => {
          if (Array.isArray(point)) return [Number(point[0]), Number(point[1])];
          if (point && typeof point === 'object') return [Number(point.Latitude ?? point.latitude ?? point.lat), Number(point.Longitude ?? point.longitude ?? point.lng)];
          return null;
        })
        .filter(point => point && point.every(Number.isFinite));
    }
    const text = String(value ?? '').trim();
    if (!text) return [];
    try {
      const parsed = JSON.parse(text);
      if (Array.isArray(parsed)) return parseLatLngList(parsed);
    } catch {}
    const nums = text.split(/[\s,]+/).map(Number).filter(Number.isFinite);
    const pts = [];
    for (let i = 0; i + 1 < nums.length; i += 2) pts.push([nums[i], nums[i + 1]]);
    return pts;
  }

  function parseBoundingBox(value) {
    if (Array.isArray(value)) {
      if (Array.isArray(value[0]) && Array.isArray(value[1])) return [value[0], value[1]];
      if (value.length >= 4) {
        const nums = value.map(Number);
        if (nums.every(Number.isFinite)) return [[nums[2], nums[1]], [nums[0], nums[3]]];
      }
    }
    const text = String(value ?? '').trim();
    if (!text) return null;
    try {
      return parseBoundingBox(JSON.parse(text));
    } catch {
      const nums = text.split(/[\s,]+/).map(Number).filter(Number.isFinite);
      if (nums.length >= 4) return [[nums[2], nums[1]], [nums[0], nums[3]]];
    }
    return null;
  }
</script>

<div
  bind:this={mapEl}
  class="sim-map"
  class:sim-unsupported={unsupportedHere}
  style={baseStyle(context)}
  data-sim-component={node.name}
></div>
