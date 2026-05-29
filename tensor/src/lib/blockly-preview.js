import { get } from 'svelte/store';
import simpleComponentsJsonText from '../../../lang/code/compdb/simple_components.json?raw';
import { listComponents, mistToXmlResult } from './falcon-wasm.js';
import { cells, designCode } from './stores.js';
import { injectFalconCommentsIntoBlocklyXml } from './blockly-comments.js';

const BLOCKLY_CORE_SCRIPTS = [
  '/compiled_blockly/messages.js',
  '/compiled_blockly/blockly-all.js',
];

const BLOCKLY_EXTENSION_SCRIPTS = [
  '/compiled_blockly/extensions/scroll-options-5.0.11.min.js',
  '/compiled_blockly/extensions/workspace-multiselect-0.1.14-beta2.min.js',
  '/compiled_blockly/extensions/workspace-search.min.js',
];

let blocklyRuntimePromise = null;
let blocklyWarmupPromise = null;
let componentTypesPromise = null;
let simpleComponentAliases = null;
let renderHostId = 0;

function runtimeLoadError(promise) {
  return promise.then(() => null, error => error);
}

function xmlChunks(xml) {
  return String(xml || '')
    .split('\0')
    .map(chunk => chunk.trim())
    .filter(Boolean);
}

function lastBlocklyXmlChunk(xml) {
  const chunks = xmlChunks(xml);
  if (!chunks.length) throw new Error('No Blockly XML was generated');
  return chunks[chunks.length - 1];
}

function componentTypeAliases() {
  if (simpleComponentAliases) return simpleComponentAliases;

  const aliases = new Map();
  const descriptors = JSON.parse(simpleComponentsJsonText);
  for (const descriptor of descriptors) {
    const name = String(descriptor?.name || '').trim();
    if (!name) continue;

    aliases.set(name, name);

    const javaType = String(descriptor?.type || '').trim();
    if (javaType) {
      aliases.set(javaType, name);
      aliases.set(javaType.split('.').pop(), name);
    }
  }

  if (aliases.has('Form')) {
    aliases.set('Screen', 'Form');
  }

  simpleComponentAliases = aliases;
  return aliases;
}

function componentDescriptorName(typeName) {
  const clean = String(typeName || '').trim();
  if (!clean) return '';
  return componentTypeAliases().get(clean) || clean;
}

function addComponentInstance(instances, typeName, instanceName) {
  const cleanInstanceName = String(instanceName || '').trim();
  const cleanTypeName = componentDescriptorName(typeName);
  if (!cleanInstanceName || !cleanTypeName) return;
  instances.set(cleanInstanceName, cleanTypeName);
}

function componentInstancesFromDefinitions(componentDefinitions) {
  const instances = new Map();
  if (!componentDefinitions || typeof componentDefinitions !== 'object') return instances;

  if (Array.isArray(componentDefinitions)) {
    for (const entry of componentDefinitions) {
      if (!entry || typeof entry !== 'object') continue;
      addComponentInstance(
        instances,
        entry.typeName ?? entry.type ?? entry.componentType ?? entry.component_type,
        entry.instanceName ?? entry.name ?? entry.id ?? entry.instance_name,
      );
    }
    return instances;
  }

  for (const [typeName, instanceNames] of Object.entries(componentDefinitions)) {
    const names = Array.isArray(instanceNames) || instanceNames instanceof Set
      ? Array.from(instanceNames)
      : [instanceNames];
    for (const instanceName of names) {
      addComponentInstance(instances, typeName, instanceName);
    }
  }

  return instances;
}

function collectComponentInstancesFromXml(xmlGenerated, instances = new Map()) {
  const parser = new DOMParser();
  for (const xmlString of xmlChunks(xmlGenerated)) {
    const xml = parser.parseFromString(xmlString, 'text/xml');
    for (const mutation of xml.querySelectorAll('mutation[component_type]')) {
      const isGeneric = mutation.getAttribute('is_generic') === 'true';
      if (isGeneric) continue;
      addComponentInstance(
        instances,
        mutation.getAttribute('component_type'),
        mutation.getAttribute('instance_name'),
      );
    }
  }
  return instances;
}

function ensureWorkspaceComponentDatabase(workspace) {
  const { Blockly } = window;
  if (!workspace.componentDb_ && Blockly?.ComponentDatabase) {
    workspace.componentDb_ = new Blockly.ComponentDatabase();
  }
  return workspace.componentDb_;
}

function populateWorkspaceComponentTypes(workspace) {
  const { Blockly } = window;
  const componentDb = ensureWorkspaceComponentDatabase(workspace);

  if (typeof workspace.populateComponentTypes === 'function') {
    workspace.populateComponentTypes(simpleComponentsJsonText, {});
  } else if (componentDb?.populateTypes) {
    componentDb.populateTypes(JSON.parse(simpleComponentsJsonText));
    componentDb.populateTranslations?.({});
  } else if (Blockly?.ComponentDatabase) {
    workspace.componentDb_ = new Blockly.ComponentDatabase();
    workspace.componentDb_.populateTypes(JSON.parse(simpleComponentsJsonText));
    workspace.componentDb_.populateTranslations?.({});
  } else {
    throw new Error('Blockly component database is unavailable');
  }
}

function ensureWorkspaceAddComponentReady(workspace) {
  workspace.typeBlock_ ||= { needsReload: {} };
  workspace.typeBlock_.needsReload ||= {};
  workspace.typeBlock_.hide ||= () => {};
}

function registerWorkspaceComponentInstances(workspace, componentDefinitions, xmlGenerated = '') {
  const componentDb = ensureWorkspaceComponentDatabase(workspace);
  const instances = collectComponentInstancesFromXml(
    xmlGenerated,
    componentInstancesFromDefinitions(componentDefinitions),
  );

  for (const [instanceName, typeName] of instances) {
    const uid = `blockly-preview-${typeName}-${instanceName}`;
    if (componentDb?.getInstance?.(uid) || componentDb?.getInstance?.(instanceName)) continue;

    if (typeof workspace.addComponent === 'function') {
      ensureWorkspaceAddComponentReady(workspace);
      workspace.addComponent(uid, instanceName, typeName);
    } else if (componentDb?.addInstance) {
      componentDb.addInstance(uid, instanceName, typeName);
    }
  }
}

export function joinedFalconCodeThroughCell(cellId, targetSource = null) {
  const notebookCells = get(cells);
  const targetIndex = notebookCells.findIndex(cell => cell.id === cellId);
  if (targetIndex === -1) {
    return String(targetSource || '').trim();
  }

  return notebookCells
    .slice(0, targetIndex + 1)
    .filter(cell => cell.type === 'code')
    .map(cell => {
      if (cell.id === cellId && targetSource !== null) return String(targetSource || '');
      return String(cell.code || '');
    })
    .join('\n\n')
    .trim();
}

async function knownComponentTypes() {
  if (!componentTypesPromise) {
    componentTypesPromise = listComponents().catch(error => {
      componentTypesPromise = null;
      throw error;
    });
  }
  return componentTypesPromise;
}

function addComponentDefinition(defs, type, instanceName, knownTypes) {
  const cleanType = String(type || '').trim();
  const cleanInstance = String(instanceName || '').trim();
  if (!cleanType || !cleanInstance) return;
  if (knownTypes.size && !knownTypes.has(cleanType)) return;
  const defType = cleanType === 'Form' ? 'Screen' : cleanType;
  defs[defType] ||= [];
  if (!defs[defType].includes(cleanInstance)) {
    defs[defType].push(cleanInstance);
  }
}

function maskDesignerIgnoredSpans(text) {
  const chars = String(text || '').split('');
  let inString = false;
  let inLineComment = false;

  for (let i = 0; i < chars.length; i += 1) {
    const ch = chars[i];
    const next = chars[i + 1];

    if (inLineComment) {
      if (ch === '\n') inLineComment = false;
      else chars[i] = ' ';
      continue;
    }

    if (inString) {
      if (ch === '\\') {
        chars[i] = ' ';
        if (i + 1 < chars.length) chars[++i] = ' ';
        continue;
      }
      if (ch === '"') inString = false;
      chars[i] = ' ';
      continue;
    }

    if (ch === '/' && next === '/') {
      chars[i] = ' ';
      chars[++i] = ' ';
      inLineComment = true;
      continue;
    }

    if (ch === '"') {
      chars[i] = ' ';
      inString = true;
    }
  }

  return chars.join('');
}

export async function componentDefinitionsFromDesigner(source = get(designCode)) {
  const componentTypes = await knownComponentTypes();
  const knownTypes = new Set(componentTypes);
  knownTypes.add('Screen');

  const defs = {};
  const typeCounts = {};
  const original = String(source || '');

  const legacyRe = /@([A-Za-z]\w*)\s*\{(?:[^{}@]*?[,\n])?\s*id\s*:\s*"([^"]+)"/g;
  let match;
  while ((match = legacyRe.exec(original)) !== null) {
    addComponentDefinition(defs, match[1], match[2], knownTypes);
  }

  const searchable = maskDesignerIgnoredSpans(original);
  const componentRe = /(?:^|[{},\n])\s*([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*([A-Za-z_]\w*))?\s*(?=\s*(?:\{|[,}\n]|$))/gm;
  while ((match = componentRe.exec(searchable)) !== null) {
    const type = match[1];
    if (knownTypes.size && !knownTypes.has(type)) continue;
    let instanceName = match[2];
    if (!instanceName) {
      typeCounts[type] = (typeCounts[type] || 0) + 1;
      instanceName = `${type}${typeCounts[type]}`;
    }
    addComponentDefinition(defs, type, instanceName, knownTypes);
  }

  return defs;
}

function installBlocklyHostStubs() {
  const roots = [window];
  try {
    if (window.parent && window.parent !== window) roots.push(window.parent);
  } catch {}
  try {
    if (window.top && window.top !== window && window.top !== window.parent) roots.push(window.top);
  } catch {}

  for (const root of roots) {
    try {
      root.BlocklyPanel_checkIsAdmin ||= () => true;
      root.BlocklyPanel_getSnapEnabled ||= () => true;
      root.BlocklyPanel_getGridEnabled ||= () => false;
      root.BlocklyPanel_setGridEnabled ||= () => {};
      root.BlocklyPanel_setSnapEnabled ||= () => {};
      root.BlocklyPanel_saveUserSettings ||= () => {};
      root.BlocklyPanel_getOdeMessage ||= key => key;
      root.BlocklyPanel_callToggleWarning ||= () => {};
      root.BlocklyPanel_getComponentContainerUuid ||= () => '';
    } catch {}
  }
}

function loadScript(src) {
  const existing = document.querySelector(`script[data-tensor-blockly-src="${src}"]`);
  if (existing?.dataset.loaded === 'true') return Promise.resolve();
  if (existing?.dataset.loading === 'true') {
    return new Promise((resolve, reject) => {
      existing.addEventListener('load', resolve, { once: true });
      existing.addEventListener('error', reject, { once: true });
    });
  }
  if (existing) existing.remove();

  return new Promise((resolve, reject) => {
    const script = document.createElement('script');
    script.src = src;
    script.async = false;
    script.dataset.tensorBlocklySrc = src;
    script.dataset.loading = 'true';
    script.onload = () => {
      script.dataset.loaded = 'true';
      script.dataset.loading = 'false';
      resolve();
    };
    script.onerror = () => {
      script.dataset.loading = 'false';
      script.dataset.error = 'true';
      script.remove();
      reject(new Error(`Failed to load ${src}`));
    };
    document.head.appendChild(script);
  });
}

export async function ensureBlocklyRuntime() {
  if (window.Blockly?.Xml && window.AI?.Blockly) return;
  if (!blocklyRuntimePromise) {
    blocklyRuntimePromise = (async () => {
      installBlocklyHostStubs();
      for (const src of BLOCKLY_CORE_SCRIPTS) {
        await loadScript(src);
      }
      await Promise.all(BLOCKLY_EXTENSION_SCRIPTS.map(src => loadScript(src)));
      if (!window.Blockly?.Xml || !window.AI?.Blockly) {
        throw new Error('Blockly runtime did not initialize');
      }
    })().catch(error => {
      blocklyRuntimePromise = null;
      throw error;
    });
  }
  await blocklyRuntimePromise;
}

export function warmBlocklyPreviewRuntime() {
  if (!blocklyWarmupPromise) {
    blocklyWarmupPromise = (async () => {
      await Promise.all([
        ensureBlocklyRuntime(),
        knownComponentTypes(),
      ]);

      if (!document.body) return;
      const host = createRenderHost();
      const workspace = createWorkspace(host);
      try {
        window.Blockly.svgResize(workspace);
        workspace.resizeContents?.();
      } finally {
        workspace.dispose();
        host.remove();
      }
    })().catch(error => {
      blocklyWarmupPromise = null;
      console.debug('[blockly-preview] warmup skipped', error);
    });
  }
  return blocklyWarmupPromise;
}

function createRenderHost() {
  const host = document.createElement('div');
  host.id = `blockly-render-host-${++renderHostId}`;
  host.style.position = 'fixed';
  host.style.left = '-10000px';
  host.style.top = '0';
  host.style.width = '1200px';
  host.style.height = '800px';
  host.style.opacity = '0';
  host.style.pointerEvents = 'none';
  host.style.overflow = 'hidden';
  document.body.appendChild(host);
  return host;
}

function createWorkspace(host) {
  const { Blockly, AI } = window;
  const options = {
    toolbox: { kind: 'flyoutToolbox', contents: [] },
    trashcan: false,
    readOnly: true,
    useDoubleClick: true,
    bumpNeighbours: true,
    renderer: 'geras2_renderer',
    grid: { spacing: 20, length: 5, snap: false, colour: '#ccc' },
    zoom: {
      controls: false,
      wheel: false,
      scaleSpeed: 1.1,
      maxScale: 3,
      minScale: 0.1,
    },
    multiselectIcon: {
      hideIcon: true,
      enabledIcon: 'static/images/select.svg',
      disabledIcon: 'static/images/unselect.svg',
    },
    multiselectCopyPaste: {
      crossTab: true,
      menu: true,
    },
  };

  const workspace = Blockly.inject(host, options);
  if (window.Multiselect && AI?.Blockly) {
    AI.Blockly.multiselect = window.Multiselect;
    new AI.Blockly.multiselect(workspace).init(options);
  }
  if (window.LexicalVariablesPlugin) {
    window.LexicalVariablesPlugin.init(workspace);
  }
  if (window.WorkspaceSearch) {
    new window.WorkspaceSearch(workspace).init();
  }
  if (window.ScrollOptions) {
    new window.ScrollOptions(workspace).init({
      edgeScrollOptions: {
        oversizeBlockMargin: 0,
        oversizeBlockThreshold: 0,
        slowBlockSpeed: 0.15,
      },
    });
  }

  workspace.formName = 'CellScreen';
  workspace.screenList_ = [];
  workspace.assetList_ = [];
  workspace.componentDb_ = new Blockly.ComponentDatabase();
  populateWorkspaceComponentTypes(workspace);
  workspace.procedureDb_ = new Blockly.ProcedureDatabase(workspace);
  workspace.variableDb_ = new Blockly.VariableDatabase();
  workspace.blocksNeedingRendering = [];
  workspace.flyout_ = workspace.getFlyout();
  workspace.injecting = false;
  workspace.injected = true;
  workspace.notYetRendered = false;
  Blockly.svgResize(workspace);
  return workspace;
}

function nextFrame() {
  return new Promise(resolve => requestAnimationFrame(() => resolve()));
}

function workspaceIsAttached(workspace) {
  try {
    const parentSvg = workspace?.getParentSvg?.();
    return Boolean(parentSvg && document.documentElement.contains(parentSvg));
  } catch {
    return false;
  }
}

function restoreMainWorkspace(previousWorkspace, renderWorkspace) {
  const common = window.Blockly?.common;
  if (!common?.getMainWorkspace || !common?.setMainWorkspace) return;
  if (common.getMainWorkspace() !== renderWorkspace) return;
  common.setMainWorkspace(workspaceIsAttached(previousWorkspace) ? previousWorkspace : null);
}

function renderXmlIntoWorkspace(xmlGenerated, workspace, { clear = true } = {}) {
  const { Blockly } = window;
  const blocks = [];
  if (clear) workspace.clear();

  for (const xmlString of xmlChunks(xmlGenerated)) {
    const xml = Blockly.utils.xml.textToDom(xmlString);
    const xmlBlock = xml.tagName?.toLowerCase() === 'block' ? xml : xml.firstElementChild;
    if (!xmlBlock || xmlBlock.tagName?.toLowerCase() !== 'block') continue;

    const block = Blockly.Xml.domToBlock(xmlBlock, workspace);
    block.initSvg();
    blocks.push(block);
  }

  for (const block of blocks) {
    if (typeof workspace.requestRender === 'function') {
      workspace.requestRender(block);
    } else if (typeof block.render === 'function') {
      block.render();
    }
  }

  return blocks;
}

function arrangeBlocksVertically(workspace) {
  const item = window.Blockly?.ContextMenuRegistry?.registry?.getItem?.('appinventor_arrange_vertical');
  if (!item || typeof item.callback !== 'function') return;
  item.callback({ workspace }, null);
}

function dataUriToBlob(dataUri) {
  const [header, encoded] = dataUri.split(',');
  const mime = header.match(/data:([^;]+)/)?.[1] || 'image/png';
  const binary = window.atob(encoded);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i += 1) {
    bytes[i] = binary.charCodeAt(i);
  }
  return new Blob([bytes], { type: mime });
}

function workspaceToPngBlob(workspace, blocks) {
  const { Blockly, AI } = window;
  const renderedBlocks = Array.isArray(blocks) && blocks.length
    ? blocks
    : workspace.getTopBlocks(true);
  if (!renderedBlocks.length) throw new Error('No Blockly blocks were generated');

  if (renderedBlocks.length === 1 && typeof Blockly.blockToPngBlob === 'function') {
    return Blockly.blockToPngBlob(renderedBlocks[0]);
  }

  return new Promise((resolve, reject) => {
    const getUri = AI?.Blockly?.ExportBlocksImage?.getUri;
    if (typeof getUri !== 'function') {
      reject(new Error('Blockly PNG exporter is unavailable'));
      return;
    }
    getUri(uri => {
      if (!uri) reject(new Error('Blockly PNG exporter returned an empty image'));
      else resolve(dataUriToBlob(uri));
    }, workspace);
  });
}

export async function blocklyXmlToPng(xml, componentDefinitions = undefined, options = {}) {
  await ensureBlocklyRuntime();
  let defs = componentDefinitions;
  if (defs === undefined) {
    try {
      defs = await componentDefinitionsFromDesigner();
    } catch {
      defs = null;
    }
  }
  const host = createRenderHost();
  const previousMainWorkspace = window.Blockly?.common?.getMainWorkspace?.() || null;
  const workspace = createWorkspace(host);
  const contextXml = options?.contextXml || '';

  try {
    registerWorkspaceComponentInstances(workspace, defs, `${contextXml}\0${xml}`);
    const contextBlocks = contextXml
      ? renderXmlIntoWorkspace(contextXml, workspace)
      : [];
    const blocks = renderXmlIntoWorkspace(xml, workspace, { clear: !contextBlocks.length });
    await nextFrame();
    arrangeBlocksVertically(workspace);
    window.Blockly.svgResize(workspace);
    workspace.resizeContents?.();
    await nextFrame();
    const blob = await workspaceToPngBlob(workspace, blocks);
    return { blob, xml, blockCount: blocks.length, contextBlockCount: contextBlocks.length };
  } finally {
    workspace.dispose();
    host.remove();
    restoreMainWorkspace(previousMainWorkspace, workspace);
  }
}

export async function falconCodeToBlocklyPng(sourceCode, componentDefinitions = null) {
  const source = String(sourceCode || '').trim();
  if (!source) throw new Error('No Falcon code to convert');
  const runtimeReady = runtimeLoadError(ensureBlocklyRuntime());
  const defs = componentDefinitions ?? await componentDefinitionsFromDesigner();
  const result = await mistToXmlResult(source, defs);
  const xml = injectFalconCommentsIntoBlocklyXml(result.xml, source, result.lineNumbers);
  const runtimeError = await runtimeReady;
  if (runtimeError) throw runtimeError;
  return blocklyXmlToPng(xml, defs);
}

export async function falconCellToBlocklyPng(cellId, targetSource = null, componentDefinitions = null) {
  const target = get(cells).find(cell => cell.id === cellId);
  const currentTargetSource = targetSource ?? target?.code ?? '';
  if (!String(currentTargetSource || '').trim()) {
    throw new Error('No Falcon code to convert');
  }

  const source = joinedFalconCodeThroughCell(cellId, currentTargetSource);
  if (!source) throw new Error('No Falcon code to convert');

  const runtimeReady = runtimeLoadError(ensureBlocklyRuntime());
  const defs = componentDefinitions ?? await componentDefinitionsFromDesigner();
  const xmlResult = await mistToXmlResult(source, defs);
  const fullXml = injectFalconCommentsIntoBlocklyXml(xmlResult.xml, source, xmlResult.lineNumbers);
  const chunks = xmlChunks(fullXml);
  const targetXml = chunks.length ? chunks[chunks.length - 1] : lastBlocklyXmlChunk(fullXml);
  const contextXml = chunks.slice(0, -1).join('\0');
  const runtimeError = await runtimeReady;
  if (runtimeError) throw runtimeError;
  const result = await blocklyXmlToPng(targetXml, defs, { contextXml });
  return {
    ...result,
    xml: targetXml,
    fullXml,
  };
}

export async function copyPngBlobToClipboard(blob) {
  if (!navigator.clipboard?.write || typeof ClipboardItem === 'undefined') {
    throw new Error('PNG clipboard writes are not supported in this browser');
  }
  const png = blob.type === 'image/png' ? blob : new Blob([blob], { type: 'image/png' });
  await navigator.clipboard.write([
    new ClipboardItem({ 'image/png': png }),
  ]);
}

export async function copyBlocksBlobToClipboard(blob, xml = '') {
  if (!navigator.clipboard?.write || typeof ClipboardItem === 'undefined') {
    throw new Error('Clipboard writes are not supported in this browser');
  }
  const png = blob.type === 'image/png' ? blob : new Blob([blob], { type: 'image/png' });
  const item = {
    'image/png': png,
  };
  if (String(xml || '').trim()) {
    item['text/plain'] = new Blob([String(xml)], { type: 'text/plain' });
  }
  await navigator.clipboard.write([new ClipboardItem(item)]);
}

export async function readPngBlobFromClipboard() {
  if (!navigator.clipboard?.read) {
    throw new Error('PNG clipboard reads are not supported in this browser');
  }

  const items = await navigator.clipboard.read();
  for (const item of items) {
    if (item.types.includes('image/png')) {
      return item.getType('image/png');
    }
  }
  throw new Error('Clipboard does not contain a PNG image');
}

export function downloadPngBlob(blob, filename = 'falcon-blocks.png') {
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}
