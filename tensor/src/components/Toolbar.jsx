function CompanionIcon({ status }) {
  const props = {
    viewBox: '0 0 24 24', width: 13, height: 13, fill: 'none',
    stroke: 'currentColor', strokeWidth: '1.9',
    strokeLinecap: 'round', strokeLinejoin: 'round',
    'aria-hidden': 'true',
  }

  if (status === 'connected') return (
    <svg {...props}><polyline points="20 6 9 17 4 12" /></svg>
  )

  if (status === 'connecting') return (
    <svg {...props} className="companion-chip-spin">
      <path d="M21 12a9 9 0 1 1-6.219-8.56" />
    </svg>
  )

  if (status === 'polling') return (
    <svg {...props}>
      <circle cx="12" cy="12" r="10" />
      <polyline points="12 6 12 12 16 14" />
    </svg>
  )

  if (status === 'error') return (
    <svg {...props}>
      <circle cx="12" cy="12" r="10" />
      <line x1="12" y1="8" x2="12" y2="12" />
      <line x1="12" y1="16" x2="12.01" y2="16" />
    </svg>
  )

  return (
    <svg {...props}>
      <rect x="3"  y="3"  width="7" height="7" />
      <rect x="14" y="3"  width="7" height="7" />
      <rect x="3"  y="14" width="7" height="7" />
      <path d="M14 14h3v3M20 14v7M14 20h3" />
    </svg>
  )
}

const CHIP_LABEL = {
  connected:  'Connected',
  connecting: 'Connecting…',
  polling:    'Waiting…',
  error:      'Error',
  idle:       'Connect',
}

const ARIA_LABEL = {
  connected:  'Companion Connected',
  connecting: 'Negotiating…',
  polling:    'Waiting for Companion…',
  error:      'Companion Error',
  idle:       'Connect Companion',
}

function ToolButton({ label, title, onClick, children }) {
  return (
    <button type="button" className="btn toolbar-tool" onClick={onClick} title={title} aria-label={title}>
      {children}
      <span>{label}</span>
    </button>
  )
}

export default function Toolbar({
  onCompanion,
  companionStatus,
  theme,
  onToggleTheme,
  onFormat,
  onSearch,
  onSymbols,
  onQuickFix,
}) {
  const ariaLabel = ARIA_LABEL[companionStatus] ?? ARIA_LABEL.idle

  return (
    <header className="toolbar">
      <div className="logo">
        <div className="logo-mark" aria-hidden="true">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path d="M3 3.5h10M8 3.5v9" stroke="white" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </div>
        <span className="logo-name">Tensor</span>
      </div>

      <div className="spacer" />

      <div className="actions">
        <ToolButton label="Format" title="Format document or selection" onClick={onFormat}>
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M4 7h16" />
            <path d="M4 12h10" />
            <path d="M4 17h13" />
          </svg>
        </ToolButton>
        <ToolButton label="Find" title="Search project" onClick={onSearch}>
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <circle cx="11" cy="11" r="7" />
            <path d="m20 20-3.5-3.5" />
          </svg>
        </ToolButton>
        <ToolButton label="Symbols" title="Open symbol list" onClick={onSymbols}>
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M8 6h13" />
            <path d="M8 12h13" />
            <path d="M8 18h13" />
            <path d="M3 6h.01" />
            <path d="M3 12h.01" />
            <path d="M3 18h.01" />
          </svg>
        </ToolButton>
        <ToolButton label="Fix" title="Show quick fixes" onClick={onQuickFix}>
          <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
            <path d="M9 18h6" />
            <path d="M10 22h4" />
            <path d="M8.5 14.5A6 6 0 1 1 15.5 14.5c-.9.7-1.5 1.7-1.5 2.5h-4c0-.8-.6-1.8-1.5-2.5z" />
          </svg>
        </ToolButton>

        <label className="theme-switch-group" title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}>
          {theme === 'dark' ? (
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
            </svg>
          ) : (
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
              <circle cx="12" cy="12" r="4"/>
              <path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/>
            </svg>
          )}
          <span className="theme-switch-label">{theme === 'dark' ? 'Dark' : 'Light'}</span>
          <button
            type="button"
            className="theme-switch"
            role="switch"
            aria-checked={theme === 'dark'}
            aria-label={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
            onClick={onToggleTheme}
          />
        </label>

        <button
          type="button"
          className={`companion-chip companion-chip--${companionStatus}`}
          onClick={onCompanion}
          aria-label={ariaLabel}
          title={ariaLabel}
        >
          <span className="companion-chip-icon" aria-hidden="true">
            <CompanionIcon status={companionStatus} />
          </span>
          <span className="companion-chip-label">{CHIP_LABEL[companionStatus] ?? CHIP_LABEL.idle}</span>
        </button>
      </div>
    </header>
  )
}
