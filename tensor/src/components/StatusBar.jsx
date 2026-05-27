import { Warning } from '@phosphor-icons/react'

const DOT_COLOR = {
  Falcon: 'var(--accent)',
  Design: 'var(--syn-number)',
}

export default function StatusBar({
  line,
  col,
  charCount,
  lang = 'Falcon',
  error = null,
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
            <Warning size={11} aria-hidden="true" />
            {diagnosticsCount > 1 ? `${diagnosticsCount} errors` : 'Parse error'}
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
