import { Lightbulb, ListBullets, MagnifyingGlass, TextAlignLeft, X } from '@phosphor-icons/react'

function PanelIcon({ type }) {
  const props = { size: 14, 'aria-hidden': true }

  if (type === 'fix') return <Lightbulb {...props} />
  if (type === 'search') return <MagnifyingGlass {...props} />
  if (type === 'symbols') return <ListBullets {...props} />
  return <TextAlignLeft {...props} />
}

export default function IdePanel({ panel, onClose }) {
  if (!panel) return null

  return (
    <aside className="ide-panel" aria-label={panel.title}>
      <div className="ide-panel-header">
        <div className="ide-panel-title">
          <PanelIcon type={panel.type} />
          <span>{panel.title}</span>
        </div>
        {panel.subtitle && <div className="ide-panel-subtitle">{panel.subtitle}</div>}
        <button type="button" className="ide-panel-close" onClick={onClose} aria-label="Close panel">
          <X size={14} aria-hidden="true" />
        </button>
      </div>

      <div className="ide-panel-list">
        {panel.items?.length ? panel.items.map((item, index) => (
          <button
            type="button"
            className="ide-panel-item"
            key={`${item.pane ?? 'action'}-${item.line ?? index}-${item.col ?? index}-${item.label}`}
            onClick={() => item.onSelect?.()}
          >
            <span className="ide-panel-item-main">
              <span className="ide-panel-item-label">{item.label}</span>
              {item.detail && <span className="ide-panel-item-detail">{item.detail}</span>}
            </span>
            {(item.pane || item.line) && (
              <span className="ide-panel-location">
                {item.pane === 'falcon' ? 'falcon' : item.pane === 'design' ? 'design' : 'action'}
                {item.line ? `:${item.line}` : ''}
              </span>
            )}
          </button>
        )) : (
          <div className="ide-panel-empty">No results</div>
        )}
      </div>
    </aside>
  )
}
