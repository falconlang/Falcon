import { useEffect, useState } from 'react'
import { describeComponent } from '../lib/falcon-wasm.js'

function stripHtml(html) {
  if (typeof DOMParser !== 'undefined') {
    const doc = new DOMParser().parseFromString(html, 'text/html')
    return (doc.body.textContent || '').replace(/\s+/g, ' ').trim()
  }
  return html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
}

export default function PropertyPopup({ propName, compName, side }) {
  const [result, setResult] = useState({ info: null, resolvedFor: null })

  const key = propName && compName ? `${compName}.${propName}` : null

  useEffect(() => {
    if (!key) { setResult({ info: null, resolvedFor: null }); return }
    let cancelled = false
    describeComponent(compName)
      .then(info => { if (!cancelled) setResult({ info, resolvedFor: key }) })
      .catch(() => { if (!cancelled) setResult({ info: null, resolvedFor: key }) })
    return () => { cancelled = true }
  }, [key, compName])

  if (!propName || !compName) return null

  const loading = result.resolvedFor !== key
  const { info } = result
  const prop = info?.blockProperties?.find(
    p => p.name === propName && p.deprecated !== 'true' && p.rw !== 'invisible'
  )

  return (
    <aside className={`component-info ${side}`}>
      <div className="component-info-name">
        <span className="ci-name-dim">@{compName}.</span>{propName}
      </div>
      {loading ? (
        <div className="component-info-desc ci-loading">Loading…</div>
      ) : !prop ? (
        <div className="component-info-desc">No documentation found.</div>
      ) : (
        <div className="ci-body">
          <div className="prop-meta">
            <span className="prop-type">{prop.type}</span>
            <span className="prop-rw">{prop.rw}</span>
          </div>
          {prop.description && (
            <p className="ci-help">{stripHtml(prop.description)}</p>
          )}
          {info.categoryString && (
            <a
              className="ci-learn-more"
              href={`https://ai2.appinventor.mit.edu/reference/components/${info.categoryString.toLowerCase()}.html#${compName}.${propName}`}
              target="_blank"
              rel="noreferrer"
              aria-label={`Learn more about ${compName}.${propName}`}
            >
              Learn more ↗
            </a>
          )}
        </div>
      )}
    </aside>
  )
}
