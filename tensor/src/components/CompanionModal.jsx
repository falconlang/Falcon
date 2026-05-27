import { useEffect, useRef } from 'react'
import { QRCodeSVG } from 'qrcode.react'

function statusText(status, error, count) {
  switch (status) {
    case 'polling':    return 'Waiting for companion…'
    case 'connecting': return 'Negotiating connection…'
    case 'connected':  return count > 0 ? `${count} ${count === 1 ? 'message' : 'messages'} received` : 'Ready'
    case 'error':      return error || 'Connection error'
    default:           return ''
  }
}

export default function CompanionModal({ open, status, code, error, messageCount, theme, onClose, onRetry, onDisconnect }) {
  const closeRef = useRef(null)
  const modalRef = useRef(null)

  useEffect(() => {
    if (!open) return
    const previousFocus = document.activeElement
    closeRef.current?.focus()

    const onKey = (e) => {
      if (e.key === 'Escape') {
        onClose()
        return
      }
      if (e.key !== 'Tab') return

      const focusable = modalRef.current?.querySelectorAll(
        'a[href], button:not([disabled]), textarea, input, select, [tabindex]:not([tabindex="-1"])',
      )
      if (!focusable?.length) {
        e.preventDefault()
        return
      }

      const first = focusable[0]
      const last = focusable[focusable.length - 1]
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault()
        last.focus()
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault()
        first.focus()
      }
    }

    window.addEventListener('keydown', onKey)
    return () => {
      window.removeEventListener('keydown', onKey)
      if (previousFocus instanceof HTMLElement) previousFocus.focus()
    }
  }, [open, onClose])

  if (!open) return null

  const isConnected = status === 'connected'
  const title       = isConnected ? 'Companion Connected' : 'Connect Companion'

  return (
    <div className="companion-backdrop" onClick={onClose}>
      <div
        ref={modalRef}
        className="companion-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="companion-title"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="companion-header">
          <div>
            <div className="companion-eyebrow">Live Testing</div>
            <h2 id="companion-title" className="companion-title">{title}</h2>
          </div>
          <button ref={closeRef} type="button" className="companion-close" onClick={onClose} aria-label="Close">
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
              <path d="M6 6l12 12M18 6L6 18" />
            </svg>
          </button>
        </div>

        {isConnected ? (
          <div className="companion-connected">
            <div className="companion-connected-icon" aria-hidden="true">
              <svg viewBox="0 0 24 24" width="34" height="34" fill="none" stroke="currentColor" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round">
                <path d="M5 12l5 5L20 7" />
              </svg>
            </div>
            <div className="companion-connected-title">Live link active</div>
            <div className="companion-connected-meta">{statusText(status, error, messageCount)}</div>
            <button type="button" className="btn companion-action companion-disconnect" onClick={onDisconnect}>
              Disconnect
            </button>
          </div>
        ) : (
          <>
            <div className="companion-qr-frame">
              {code && (
                <QRCodeSVG
                  value={code}
                  size={196}
                  bgColor="transparent"
                  fgColor="#F2F6F1"
                  level="M"
                  marginSize={0}
                />
              )}
            </div>

            <div className="companion-divider"><span>or enter code</span></div>

            {code && (
              <div className="companion-code" aria-label={`Companion code ${code.split('').join(' ')}`}>
                {code.split('').map((d, i) => (
                  <span key={i} className="companion-code-digit">{d}</span>
                ))}
              </div>
            )}

            {status === 'idle' ? (
              <p className="companion-hint">
                Scan with the <b>MIT AI2 Companion</b> app, or enter the code manually.
              </p>
            ) : (
              <div className={`companion-status status-${status}`} aria-live="polite">
                <span className="companion-status-dot" aria-hidden="true" />
                <span className="companion-status-text">{statusText(status, error, messageCount)}</span>
              </div>
            )}

            {status === 'error' && (
              <button type="button" className="btn btn-tonal companion-action" onClick={onRetry}>
                Try Again
              </button>
            )}
          </>
        )}
      </div>
    </div>
  )
}
