export default function ErrorPanel({ falconError, designError }) {
  if (!falconError && !designError) return null

  const entries = []
  if (falconError) entries.push({ file: 'Screen1.falcon', msg: falconError })
  if (designError) entries.push({ file: 'Screen1.design', msg: designError })

  return (
    <div className="error-panel" role="alert" aria-label="Parse errors">
      {entries.map(({ file, msg }) => (
        <div key={file} className="error-panel-entry">
          <span className="error-panel-file">{file}</span>
          <pre className="error-panel-msg">{msg}</pre>
        </div>
      ))}
    </div>
  )
}
