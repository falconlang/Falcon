export function isBlocklyReady() {
  return typeof window.Blockly !== 'undefined'
}

export function waitForBlockly(timeout = 15000) {
  if (isBlocklyReady()) return Promise.resolve()
  return new Promise((resolve, reject) => {
    const deadline = Date.now() + timeout
    const check = () => {
      if (isBlocklyReady()) resolve()
      else if (Date.now() > deadline) reject(new Error('Blockly not available'))
      else setTimeout(check, 100)
    }
    setTimeout(check, 0)
  })
}

export function isWorkspaceAttached(workspace) {
  const group = workspace?.svgGroup_
  return Boolean(group?.parentElement?.parentElement?.parentElement && document.contains(group))
}

export function safeSvgResize(workspace) {
  if (!isWorkspaceAttached(workspace)) return
  window.Blockly.svgResize(workspace)
}

function normalizeFitPadding(padding) {
  if (typeof padding === 'number') {
    return { top: padding, right: padding, bottom: padding, left: padding }
  }
  return {
    top: padding?.top ?? 12,
    right: padding?.right ?? 12,
    bottom: padding?.bottom ?? 12,
    left: padding?.left ?? 12,
  }
}

export function fitBlocklyWorkspace(workspace, padding = 12) {
  if (!isWorkspaceAttached(workspace)) return
  const svg = workspace.getParentSvg?.()
  const host = svg?.parentElement
  const bounds = workspace.getBlocksBoundingBox?.()
  if (!host || !bounds) return

  const inset = normalizeFitPadding(padding)
  const blockWidth = Math.max(1, bounds.right - bounds.left)
  const blockHeight = Math.max(1, bounds.bottom - bounds.top)
  const availableWidth = Math.max(1, host.clientWidth - inset.left - inset.right)
  const availableHeight = Math.max(1, host.clientHeight - inset.top - inset.bottom)
  const scale = Math.max(0.08, Math.min(1, availableWidth / blockWidth, availableHeight / blockHeight))

  workspace.setScale(scale)
  workspace.scrollX = inset.left - bounds.left * scale
  workspace.scrollY = inset.top - bounds.top * scale
  workspace.translate(workspace.scrollX, workspace.scrollY)
}

function getTopBlock(block) {
  while (block.getParent()) block = block.getParent()
  return block
}

function renderBlocksNow(workspace, blocks) {
  const topBlocks = []
  const seen = new Set()

  for (const block of blocks) {
    const topBlock = getTopBlock(block)
    if (!seen.has(topBlock.id)) {
      seen.add(topBlock.id)
      topBlocks.push(topBlock)
    }
  }

  if (typeof workspace.render === 'function') {
    workspace.render(topBlocks)
  } else {
    topBlocks.forEach(block => block.render(false))
  }
}

function arrangeBlocksVertically(workspace) {
  const topBlocks = workspace.getTopBlocks(false)
  const metrics = workspace.getMetrics?.()
  const scale = workspace.scale || 1
  const padding = 16 / scale
  const x = Number.isFinite(metrics?.viewLeft) ? metrics.viewLeft + padding : padding
  let y = Number.isFinite(metrics?.viewTop) ? metrics.viewTop + padding : padding
  const spacer = 24 / scale

  for (const block of topBlocks) {
    const xy = block.getRelativeToSurfaceXY()
    block.moveBy(x - xy.x, y - xy.y)
    const size = block.getHeightWidth()
    y += Math.max(size.height, 24) + spacer
  }
}

function withBlocklyEventsDisabled(Blockly, fn) {
  const events = Blockly.Events
  const wereEnabled = events?.isEnabled?.() ?? true
  events?.disable?.()
  try {
    return fn()
  } finally {
    if (wereEnabled) events?.enable?.()
  }
}

export function createBlocklyPreviewWorkspace(container, { wheelZoom = false } = {}) {
  const Blockly = window.Blockly
  const workspace = Blockly.inject(container, {
    toolbox: {
      kind: 'flyoutToolbox',
      contents: [],
    },
    readOnly: true,
    trashcan: false,
    useDoubleClick: true,
    bumpNeighbours: true,
    renderer: 'geras2_renderer',
    zoom: { controls: false, wheel: wheelZoom, scaleSpeed: 1.1, maxScale: 3, minScale: 0.1 },
  })

  workspace.formName = 'PreviewScreen'
  workspace.screenList_ = []
  workspace.assetList_ = []
  workspace.componentDb_ = new Blockly.ComponentDatabase()
  workspace.procedureDb_ = new Blockly.ProcedureDatabase(workspace)
  workspace.variableDb_ = new Blockly.VariableDatabase()
  workspace.blocksNeedingRendering = []
  workspace.flyout_ = workspace.getFlyout()
  workspace.injecting = false
  workspace.injected = true
  workspace.notYetRendered = true

  if (window.LexicalVariablesPlugin) {
    window.LexicalVariablesPlugin.init(workspace)
  }

  Blockly.browserEvents.bind(workspace.svgGroup_, 'focus', workspace, workspace.markFocused)

  return workspace
}

export function loadXmlIntoBlocklyWorkspace(workspace, xml, { fit = false, padding } = {}) {
  const Blockly = window.Blockly
  if (!isWorkspaceAttached(workspace)) return

  withBlocklyEventsDisabled(Blockly, () => {
    workspace.clear()
    safeSvgResize(workspace)
    const xmlStrings = String(xml ?? '').split('\0').map(s => s.trim()).filter(Boolean)
    const blocks = []

    for (const xmlString of xmlStrings) {
      const dom = Blockly.utils.xml.textToDom(xmlString)
      const xmlBlock = dom.firstElementChild
      if (!xmlBlock) continue
      const block = Blockly.Xml.domToBlock(xmlBlock, workspace)
      block.initSvg()
      blocks.push(block)
    }

    renderBlocksNow(workspace, blocks)
    arrangeBlocksVertically(workspace)
    workspace.resizeContents()
  })

  safeSvgResize(workspace)
  if (fit) fitBlocklyWorkspace(workspace, padding)
}
