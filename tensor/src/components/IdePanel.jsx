function PanelIcon({ type }) {
  const common = {
    viewBox: '0 0 24 24',
    width: 14,
    height: 14,
    fill: 'none',
    stroke: 'currentColor',
    strokeWidth: '1.9',
    strokeLinecap: 'round',
    strokeLinejoin: 'round',
    'aria-hidden': 'true',
  }

  if (type === 'fix') {
    return (
      <svg {...common}>
        <path d="M9 18h6" />
        <path d="M10 22h4" />
        <path d="M8.5 14.5A6 6 0 1 1 15.5 14.5c-.9.7-1.5 1.7-1.5 2.5h-4c0-.8-.6-1.8-1.5-2.5z" />
      </svg>
    )
  }

  if (type === 'search') {
    return (
      <svg {...common}>
        <circle cx="11" cy="11" r="7" />
        <path d="m20 20-3.5-3.5" />
      </svg>
    )
  }

  return (
    <svg {...common}>
      <path d="M4 7h16" />
      <path d="M4 12h10" />
      <path d="M4 17h13" />
    </svg>
  )
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
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
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
