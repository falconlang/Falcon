const DOT_COLOR = {
  Falcon: 'var(--accent)',
  Design: 'var(--syn-number)',
}

function BlocksIcon() {
  return (
    <svg viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" style={{ display: 'block', flexShrink: 0 }}>
      <path d="M4 7h3a1 1 0 0 0 1-1V5a2 2 0 0 1 4 0v1a1 1 0 0 0 1 1h3a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-1a2 2 0 0 0 0 4h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1h-3a1 1 0 0 1-1-1v-1a2 2 0 0 0-4 0v1a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1h1a2 2 0 0 0 0-4H4a1 1 0 0 1-1-1V8a1 1 0 0 1 1-1z"/>
    </svg>
  )
}

function WarningIcon() {
  return (
    <svg viewBox="0 0 24 24" width="11" height="11" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" style={{ display: 'block', flexShrink: 0 }}>
      <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
      <line x1="12" y1="9" x2="12" y2="13"/>
      <line x1="12" y1="17" x2="12.01" y2="17"/>
    </svg>
  )
}

export default function StatusBar({
  line,
  col,
  charCount,
  lang = 'Falcon',
  error = null,
  blockCount = 0,
  diagnosticsCount = 0,
  breadcrumb = '',
}) {
  return (
    <footer className="statusbar">
      <div className="status-left">
        <div className="lang-badge">
          <div className="lang-dot" style={{ background: DOT_COLOR[lang] ?? 'var(--accent)' }} aria-hidden="true" />
          {lang}
        </div>
        {error && (
          <span className="status-error-badge" aria-live="polite">
            <WarningIcon />
            {diagnosticsCount > 1 ? `${diagnosticsCount} errors` : 'Parse error'}
          </span>
        )}
        {lang === 'Falcon' && blockCount > 0 && (
          <span
            className="status-blocks"
            title={`${blockCount} Blockly block${blockCount !== 1 ? 's' : ''}`}
          >
            <BlocksIcon />
            {blockCount}
          </span>
        )}
        {breadcrumb && <span className="status-breadcrumb">{breadcrumb}</span>}
      </div>
      <div className="status-right">
        <span>Ln {line}, Col {col}</span>
        <span>{charCount} chars</span>
        {!error && <span className="status-meta">2 spaces</span>}
      </div>
    </footer>
  )
}
