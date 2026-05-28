import {
  mistToXml,
  mistToXmlDiagnosticResult,
  blocklyToYail,
  annToYail,
  annToYailDiagnosticResult,
} from './falcon-wasm.js'

const RENDEZVOUS = 'https://rendezvous.appinventor.mit.edu/rendezvous/'

function emitLog(onLog, message, level = 'info') {
  if (typeof onLog === 'function') onLog({ message, level })
}

function describeIceCandidate(candidate) {
  if (!candidate) return 'complete'
  const text = candidate.candidate || ''
  const type = text.match(/\btyp\s+(\w+)/)?.[1] || 'unknown'
  const protocol = text.match(/\b(udp|tcp)\b/i)?.[1]?.toUpperCase() || 'transport'
  return `${type} ${protocol} candidate`
}

function maskIgnoredSpans(text) {
  const chars = text.split('')
  let inString = false
  let inLineComment = false

  for (let i = 0; i < chars.length; i++) {
    const ch = chars[i]
    const next = chars[i + 1]

    if (inLineComment) {
      if (ch === '\n') inLineComment = false
      else chars[i] = ' '
      continue
    }

    if (inString) {
      if (ch === '\n') continue
      if (ch === '\\') {
        chars[i] = ' '
        if (i + 1 < chars.length) chars[++i] = ' '
        continue
      }
      if (ch === '"') inString = false
      chars[i] = ' '
      continue
    }

    if (ch === '/' && next === '/') {
      chars[i] = ' '
      chars[++i] = ' '
      inLineComment = true
      continue
    }

    if (ch === '"') {
      chars[i] = ' '
      inString = true
    }
  }

  return chars.join('')
}

export function extractComponentDefs(annSource) {
  const defs = {}
  const addDef = (type, id) => {
    if (!defs[type]) defs[type] = []
    if (!defs[type].includes(id)) defs[type].push(id)
  }

  // Legacy @Type { id: "name" } syntax
  const re = /@([A-Za-z]\w*)\s*\{(?:[^{}@]*?[,\n])?\s*id\s*:\s*"([^"]+)"/g
  let m
  while ((m = re.exec(annSource)) !== null) {
    addDef(m[1], m[2])
  }

  // Current Type.name syntax
  const searchable = maskIgnoredSpans(annSource)
  const dotRe = /(?:^|[{}\n])\s*([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*([A-Za-z_]\w*))\s*(?=\{|$)/gm
  while ((m = dotRe.exec(searchable)) !== null) {
    addDef(m[1], m[2])
  }

  return defs
}

function findMissingScreens(componentDefs, yail) {
  const screenIds = componentDefs.Screen ?? []
  return screenIds.filter(id => !yail.includes(id))
}

function extractEventDefs(blocklyXml) {
  const defs = []
  const re = /<mutation\b[^>]*\binstance_name="([^"]+)"[^>]*\bevent_name="([^"]+)"/g
  let m
  while ((m = re.exec(blocklyXml)) !== null) {
    defs.push({ component: m[1], event: m[2] })
  }
  return defs
}

function findMissingEvents(eventDefs, yail) {
  return eventDefs.filter(({ component, event }) => !yail.includes(`(define-event ${component} ${event}`))
}

function wrapForRepl(yail) {
  return `(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 ${yail}))`
}

export function wrapSnippet(yail) {
  return `(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 (begin ${yail})))`
}

export async function compileForCompanion(mistSource, annSource) {
  const componentDefs = extractComponentDefs(annSource)
  const blocklyXml    = await mistToXml(mistSource, componentDefs)
  const codeYail      = await blocklyToYail(blocklyXml)
  const eventDefs     = extractEventDefs(blocklyXml)
  const missingEvents = findMissingEvents(eventDefs, codeYail)
  if (missingEvents.length) {
    throw new Error(`Compiled YAIL payload is missing event code for: ${missingEvents.map(({ component, event }) => `${component}.${event}`).join(', ')}`)
  }
  const fullYail      = await annToYail(annSource, codeYail)
  const missingScreens = findMissingScreens(componentDefs, fullYail)
  if (missingScreens.length) {
    throw new Error(`Compiled YAIL payload is missing screen code for: ${missingScreens.join(', ')}`)
  }
  const replPayload   = wrapForRepl(fullYail)
  return { componentDefs, screenIds: componentDefs.Screen ?? [], eventDefs, blocklyXml, codeYail, fullYail, replPayload }
}

export async function validateSources(mistSource, annSource) {
  try {
    const annPreflight = await annToYailDiagnosticResult(annSource, '')
    if (!annPreflight.ok) {
      return {
        falconError: null,
        designError: annPreflight.error,
        falconDiagnostics: [],
        designDiagnostics: annPreflight.diagnostics,
      }
    }

    const componentDefs = extractComponentDefs(annSource)
    const mistResult = await mistToXmlDiagnosticResult(mistSource, componentDefs)
    if (!mistResult.ok) {
      return {
        falconError: mistResult.error,
        designError: null,
        falconDiagnostics: mistResult.diagnostics,
        designDiagnostics: [],
      }
    }
    const xml      = mistResult.xml
    const codeYail = await blocklyToYail(xml)
    const annCompile = await annToYailDiagnosticResult(annSource, codeYail)
    if (!annCompile.ok) {
      return {
        falconError: null,
        designError: annCompile.error,
        falconDiagnostics: [],
        designDiagnostics: annCompile.diagnostics,
      }
    }
    return { falconError: null, designError: null, falconDiagnostics: [], designDiagnostics: [] }
  } catch (e) {
    if (e.message?.includes('timeout waiting for WASM')) {
      return { falconError: null, designError: null, falconDiagnostics: [], designDiagnostics: [] }
    }
    return {
      falconError: e.message,
      designError: null,
      falconDiagnostics: e.diagnostics ?? [],
      designDiagnostics: [],
    }
  }
}

export async function compileSnippet(mistSnippet, componentDefs) {
  const xml  = await mistToXml(mistSnippet, componentDefs)
  const yail = await blocklyToYail(xml)
  return wrapSnippet(yail)
}

async function sha1Hex(text) {
  const buf  = new TextEncoder().encode(text)
  const hash = await crypto.subtle.digest('SHA-1', buf)
  return Array.from(new Uint8Array(hash)).map(b => b.toString(16).padStart(2, '0')).join('')
}

function sleep(ms, signal) {
  return new Promise((resolve, reject) => {
    const t = setTimeout(resolve, ms)
    if (signal) {
      const onAbort = () => { clearTimeout(t); reject(new DOMException('aborted', 'AbortError')) }
      if (signal.aborted) onAbort()
      else signal.addEventListener('abort', onAbort, { once: true })
    }
  })
}

export async function pollRendezvous(code, signal, onLog) {
  emitLog(onLog, 'Rendezvous: computing connection key')
  const digest = await sha1Hex(code)
  emitLog(onLog, `Rendezvous: contacting ${RENDEZVOUS}`)
  for (let i = 0; i < 60; i++) {
    if (signal?.aborted) throw new DOMException('aborted', 'AbortError')
    emitLog(onLog, `Rendezvous: poll ${i + 1}/60`)
    const resp = await fetch(RENDEZVOUS + digest, { signal })
    emitLog(onLog, `Rendezvous: HTTP ${resp.status}`)
    if (!resp.ok) throw new Error(`rendezvous: HTTP ${resp.status}`)
    const body = await resp.arrayBuffer()
    if (body.byteLength > 0) {
      emitLog(onLog, `Rendezvous: companion response received (${body.byteLength} bytes)`)
      const config = JSON.parse(new TextDecoder().decode(body))
      const iceCount = (config.iceServers ?? config.iceservers ?? []).length
      emitLog(onLog, `Rendezvous: WebRTC config includes ${iceCount} ICE server${iceCount === 1 ? '' : 's'}`)
      return { digest, config }
    }
    emitLog(onLog, 'Rendezvous: no companion yet, retrying in 1s', 'muted')
    await sleep(1000, signal)
  }
  throw new Error('Companion not found (timeout)')
}

async function postSignal(url, body, signal, onLog, label = 'signal') {
  emitLog(onLog, `Rendezvous2: POST ${label}`)
  const resp = await fetch(url, {
    method:  'POST',
    headers: { 'Content-Type': 'application/json' },
    body:    JSON.stringify(body),
    signal,
  })
  emitLog(onLog, `Rendezvous2: ${label} HTTP ${resp.status}`)
  if (!resp.ok) throw new Error(`rendezvous signal: HTTP ${resp.status}`)
}

function extractSessionDescription(hunk) {
  if (hunk.offer)  return hunk.offer
  if (hunk.answer) return hunk.answer
  if (hunk.sdp && hunk.type) return { sdp: hunk.sdp, type: hunk.type }
  return null
}

async function receiveOfferResponse(peer, responseUrl, signal, onLog) {
  let answerSet = false
  const pending = []
  const seen = new Set()

  emitLog(onLog, 'WebRTC: waiting for SDP answer from companion')
  for (let i = 0; i < 60 && peer.connectionState !== 'connected' && peer.connectionState !== 'failed'; i++) {
    if (signal?.aborted) throw new DOMException('aborted', 'AbortError')
    try {
      emitLog(onLog, `Rendezvous2: answer poll ${i + 1}/60`)
      const resp = await fetch(responseUrl, { signal })
      emitLog(onLog, `Rendezvous2: answer poll HTTP ${resp.status}`)
      if (resp.ok) {
        const text = await resp.text()
        if (text) {
          const hunks = JSON.parse(text)
          emitLog(onLog, `Rendezvous2: received ${hunks.length} signaling hunk${hunks.length === 1 ? '' : 's'}`)
          for (const hunk of hunks) {
            if (hunk.nonce !== undefined && seen.has(hunk.nonce)) continue
            if (hunk.nonce !== undefined) seen.add(hunk.nonce)

            const desc = extractSessionDescription(hunk)
            if (desc && !answerSet) {
              await peer.setRemoteDescription(new RTCSessionDescription(desc))
              answerSet = true
              emitLog(onLog, `WebRTC: remote ${desc.type || 'answer'} applied`)
              for (const c of pending) await peer.addIceCandidate(new RTCIceCandidate(c))
              if (pending.length) emitLog(onLog, `ICE: added ${pending.length} queued remote candidate${pending.length === 1 ? '' : 's'}`)
              pending.length = 0
            }
            if (hunk.candidate) {
              const candidateLabel = describeIceCandidate(hunk.candidate)
              if (answerSet) {
                await peer.addIceCandidate(new RTCIceCandidate(hunk.candidate))
                emitLog(onLog, `ICE: added remote ${candidateLabel}`)
              } else {
                pending.push(hunk.candidate)
                emitLog(onLog, `ICE: queued remote ${candidateLabel}`)
              }
            }
          }
        }
      }
    } catch (e) {
      if (e.name === 'AbortError') throw e
    }
    await sleep(1000, signal)
  }

  if (peer.connectionState !== 'connected') {
    throw new Error(`Negotiation timed out (state: ${peer.connectionState}, SDP answer: ${answerSet ? 'received' : 'missing'})`)
  }
}

export async function connectCompanion(code, digest, config, signal, onLog) {
  const rawServers = config.iceServers ?? config.iceservers
  if (!Array.isArray(rawServers)) {
    throw new Error('Rendezvous response is missing iceServers — companion may be in legacy (non-WebRTC) mode')
  }
  if (!config.rendezvous2) {
    throw new Error('Rendezvous response is missing rendezvous2 URL')
  }

  const iceServers = rawServers
    .map(s => {
      const rawUrls = s.urls ?? s.url ?? s.server
      const urls = Array.isArray(rawUrls) ? rawUrls.filter(Boolean) : [rawUrls].filter(Boolean)
      return {
        urls,
        username:   s.username,
        credential: s.password ?? s.credential,
      }
    })
    .filter(s => s.urls.length > 0)
  if (!iceServers.length) throw new Error('Rendezvous response did not include usable ICE server URLs')

  const iceUrlCount = iceServers.reduce((count, server) => count + server.urls.length, 0)
  emitLog(onLog, `ICE: using ${iceServers.length} server entr${iceServers.length === 1 ? 'y' : 'ies'} (${iceUrlCount} URL${iceUrlCount === 1 ? '' : 's'})`)

  const peer    = new RTCPeerConnection({ iceServers })
  const channel = peer.createDataChannel('data', { ordered: true })
  emitLog(onLog, 'WebRTC: peer created, data channel requested')

  peer.onicegatheringstatechange = () => emitLog(onLog, `ICE: gathering state ${peer.iceGatheringState}`)
  peer.oniceconnectionstatechange = () => emitLog(onLog, `ICE: connection state ${peer.iceConnectionState}`)
  peer.onconnectionstatechange = () => emitLog(onLog, `WebRTC: peer connection state ${peer.connectionState}`)
  peer.onsignalingstatechange = () => emitLog(onLog, `WebRTC: signaling state ${peer.signalingState}`)

  peer.onicecandidate = ({ candidate }) => {
    if (!candidate) {
      emitLog(onLog, 'ICE: local candidate gathering complete')
      return
    }
    emitLog(onLog, `ICE: local ${describeIceCandidate(candidate)} discovered`)
    postSignal(config.rendezvous2, {
      key:       digest + '-s',
      webrtc:    true,
      nonce:     Math.floor(Math.random() * 10000) + 1,
      candidate: candidate.toJSON(),
    }, signal, onLog, 'local ICE candidate').catch(() => {})
  }

  emitLog(onLog, 'WebRTC: creating SDP offer')
  const offer = await peer.createOffer()
  await peer.setLocalDescription(offer)
  emitLog(onLog, 'WebRTC: local offer set')
  await postSignal(config.rendezvous2, {
    key:       code + '-s',
    webrtc:    true,
    nonce:     Math.floor(Math.random() * 10000) + 1,
    offer,
    candidate: null,
  }, signal, onLog, 'SDP offer')

  await receiveOfferResponse(peer, config.rendezvous2 + code + '-r', signal, onLog)
  await waitForChannelOpen(channel, signal, onLog)
  return { peer, channel }
}

function waitForChannelOpen(channel, signal, onLog, timeoutMs = 5000) {
  if (channel.readyState === 'open') return Promise.resolve()
  emitLog(onLog, 'WebRTC: waiting for data channel open')
  return new Promise((resolve, reject) => {
    const cleanup = () => {
      channel.removeEventListener('open',  onOpen)
      channel.removeEventListener('error', onErr)
      channel.removeEventListener('close', onErr)
      signal?.removeEventListener('abort', onAbort)
      clearTimeout(timer)
    }
    const onOpen  = () => { emitLog(onLog, 'WebRTC: data channel open'); cleanup(); resolve() }
    const onErr   = () => { emitLog(onLog, `WebRTC: data channel ${channel.readyState}`, 'error'); cleanup(); reject(new Error('Data channel closed before opening')) }
    const onAbort = () => { cleanup(); reject(new DOMException('aborted', 'AbortError')) }
    const timer = setTimeout(() => { emitLog(onLog, 'WebRTC: data channel open timeout', 'error'); cleanup(); reject(new Error('Data channel open timeout')) }, timeoutMs)

    channel.addEventListener('open',  onOpen,  { once: true })
    channel.addEventListener('error', onErr,   { once: true })
    channel.addEventListener('close', onErr,   { once: true })
    signal?.addEventListener('abort', onAbort, { once: true })
  })
}
