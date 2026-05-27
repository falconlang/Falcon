import {
  Check,
  CircleNotch,
  Clock,
  Moon,
  QrCode,
  Sun,
  TextT,
  Warning,
} from '@phosphor-icons/react'

function CompanionIcon({ status }) {
  const props = { size: 13, 'aria-hidden': true }

  if (status === 'connected') return <Check {...props} weight="bold" />
  if (status === 'connecting') return <CircleNotch {...props} className="companion-chip-spin" />
  if (status === 'polling') return <Clock {...props} />
  if (status === 'error') return <Warning {...props} />
  return <QrCode {...props} />
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

export default function Toolbar({
  onCompanion,
  companionStatus,
  theme,
  onToggleTheme,
}) {
  const ariaLabel = ARIA_LABEL[companionStatus] ?? ARIA_LABEL.idle

  return (
    <header className="toolbar">
      <div className="logo">
        <div className="logo-mark" aria-hidden="true">
          <TextT size={17} weight="bold" />
        </div>
        <span className="logo-name">Tensor</span>
      </div>

      <div className="spacer" />

      <div className="actions">
        <label className="theme-switch-group" title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}>
          {theme === 'dark'
            ? <Moon size={16} aria-hidden="true" />
            : <Sun size={16} aria-hidden="true" />
          }
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
