import { useEffect, useState } from 'react'
import { describeComponent } from '../lib/falcon-wasm.js'

function stripHtml(html) {
  if (typeof DOMParser !== 'undefined') {
    const doc = new DOMParser().parseFromString(html, 'text/html')
    return (doc.body.textContent || '').replace(/\s+/g, ' ').trim()
  }
  return html.replace(/<[^>]*>/g, ' ').replace(/\s+/g, ' ').trim()
}

export default function ComponentPopup({ name, side }) {
  const [result, setResult] = useState({ info: null, resolvedFor: null })

  useEffect(() => {
    if (!name) { setResult({ info: null, resolvedFor: null }); return }
    let cancelled = false
    describeComponent(name)
      .then(info  => { if (!cancelled) setResult({ info, resolvedFor: name }) })
      .catch(()   => { if (!cancelled) setResult({ info: null, resolvedFor: name }) })
    return () => { cancelled = true }
  }, [name])

  if (!name) return null

  const loading  = result.resolvedFor !== name
  const { info } = result
  const helpText = info?.helpString ? stripHtml(info.helpString) : null
  const docsUrl  = info?.categoryString
    ? `https://ai2.appinventor.mit.edu/reference/components/${info.categoryString.toLowerCase()}.html#${name}`
    : null

  return (
    <aside className={`component-info ${side}`}>
      <div className="component-info-name">@{name}</div>
      {loading ? (
        <div className="component-info-desc ci-loading">Loading…</div>
      ) : !info ? (
        <div className="component-info-desc">No documentation found.</div>
      ) : (
        <div className="ci-body">
          {helpText && <p className="ci-help">{helpText}</p>}
          {docsUrl && (
            <a className="ci-learn-more" href={docsUrl} target="_blank" rel="noreferrer" aria-label={`Learn more about ${name}`}>
              Learn more ↗
            </a>
          )}
        </div>
      )}
    </aside>
  )
}
