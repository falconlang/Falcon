import { useEffect, useRef } from 'react'
import {
  createBlocklyPreviewWorkspace,
  fitBlocklyWorkspace,
  loadXmlIntoBlocklyWorkspace,
  safeSvgResize,
  waitForBlockly,
} from '../lib/blockly-preview'

export default function BlocklyPanel({ entries }) {
  const containerRef = useRef(null)
  const wsRef       = useRef(null)
  const entriesRef  = useRef(entries)

  // Keep latest entries accessible inside the async init closure
  useEffect(() => {
    entriesRef.current = entries
    const ws = wsRef.current
    if (!ws) return
    const xml = entries.map(e => e.xml).join('\0')
    loadXmlIntoBlocklyWorkspace(ws, xml, { fit: true })
    requestAnimationFrame(() => {
      if (wsRef.current) {
        safeSvgResize(wsRef.current)
        fitBlocklyWorkspace(wsRef.current)
      }
    })
  }, [entries])

  // Create workspace once on mount, dispose on unmount
  useEffect(() => {
    let disposed = false
    let resizeObserver = null

    ;(async () => {
      try {
        await waitForBlockly()
        await new Promise(r => requestAnimationFrame(r))
        if (disposed || !containerRef.current) return

        const ws = createBlocklyPreviewWorkspace(containerRef.current, { wheelZoom: true })
        wsRef.current = ws

        resizeObserver = new ResizeObserver(() => {
          safeSvgResize(ws)
          fitBlocklyWorkspace(ws)
        })
        resizeObserver.observe(containerRef.current)

        const xml = entriesRef.current.map(e => e.xml).join('\0')
        if (xml) {
          loadXmlIntoBlocklyWorkspace(ws, xml, { fit: true })
          requestAnimationFrame(() => {
            if (!disposed && wsRef.current) {
              safeSvgResize(ws)
              fitBlocklyWorkspace(ws)
            }
          })
        }
      } catch {
        // Blockly not yet available
      }
    })()

    return () => {
      disposed = true
      resizeObserver?.disconnect()
      if (wsRef.current) {
        wsRef.current.dispose()
        wsRef.current = null
      }
    }
  }, [])

  return (
    <div className="pane blockly-panel-pane">
      <div className="pane-header">
        Blocks
        {entries.length > 0 && (
          <span className="blockly-panel-count">{entries.length}</span>
        )}
      </div>
      <div className="pane-editor">
        <div className="blockly-panel-workspace" ref={containerRef} />
      </div>
    </div>
  )
}
