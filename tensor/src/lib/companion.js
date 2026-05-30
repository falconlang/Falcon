import {
  mistToXml,
  mistToXmlDiagnosticResult,
  blocklyToYail,
  annToYail,
  annToYailDiagnosticResult,
  listComponents,
} from './falcon-wasm.js'
import {
  DEBUG_CONTINUE_GLOBAL,
  createDebugSessionId,
  ensureDebugNotifierDesignSource,
  instrumentFalconSourceForDebug,
} from './debug-instrumentation.js'

const RENDEZVOUS = 'https://rendezvous.appinventor.mit.edu/rendezvous/'
const RENDEZVOUS2 = 'https://rendezvous.appinventor.mit.edu/rendezvous2/'
const DEFAULT_ICE_SERVERS = [
  {
    server: 'turn:turn.appinventor.mit.edu:3478',
    username: 'oh',
    password: 'boy',
  },
]
const WEBRTC_CHUNK_LENGTH = 15000
const COMPANION_CODE_ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
const COMPANION_CODE_LENGTH = 5
const KNOWN_COMPONENT_TYPES = new Set([
  'Screen',
  'AbsoluteArrangement', 'AccelerometerSensor', 'ActivityStarter', 'AnomalyDetection',
  'Ball', 'BarcodeScanner', 'Barometer', 'BluetoothClient', 'BluetoothServer', 'Button',
  'Camcorder', 'Camera', 'Canvas', 'Chart', 'ChartData2D', 'ChatBot', 'CheckBox',
  'Circle', 'CircularProgress', 'Clock', 'CloudDB', 'ContactPicker', 'DataFile',
  'DatePicker', 'EmailPicker', 'Ev3ColorSensor', 'Ev3Commands', 'Ev3GyroSensor',
  'Ev3Motors', 'Ev3Sound', 'Ev3TouchSensor', 'Ev3UI', 'Ev3UltrasonicSensor',
  'FeatureCollection', 'File', 'FilePicker', 'FirebaseDB', 'FusiontablesControl',
  'GameClient', 'GyroscopeSensor', 'HorizontalArrangement', 'HorizontalScrollArrangement',
  'Hygrometer', 'Image', 'ImageBot', 'ImagePicker', 'ImageSprite', 'Label',
  'LightSensor', 'LineString', 'LinearProgress', 'ListPicker', 'ListView',
  'LocationSensor', 'MagneticFieldSensor', 'Map', 'Marker', 'MediaStore', 'Navigation',
  'NearField', 'Notifier', 'NxtColorSensor', 'NxtDirectCommands', 'NxtDrive',
  'NxtLightSensor', 'NxtSoundSensor', 'NxtTouchSensor', 'NxtUltrasonicSensor',
  'OrientationSensor', 'PasswordTextBox', 'Pedometer', 'PhoneCall', 'PhoneNumberPicker',
  'PhoneStatus', 'Player', 'Polygon', 'ProximitySensor', 'Rectangle', 'Regression',
  'Serial', 'Sharing', 'Slider', 'Sound', 'SoundRecorder', 'SpeechRecognizer',
  'Spinner', 'Spreadsheet', 'Switch', 'TableArrangement', 'TextBox', 'TextToSpeech',
  'Texting', 'Thermometer', 'TimePicker', 'TinyDB', 'TinyWebDB', 'Translator',
  'Trendline', 'Twitter', 'VerticalArrangement', 'VerticalScrollArrangement',
  'VideoPlayer', 'Voting', 'Web', 'WebViewer', 'YandexTranslate',
])
KNOWN_COMPONENT_TYPES.add('Form')

let chunkSequence = 0

function emitLog(onLog, message, level = 'info') {
  if (typeof onLog === 'function') onLog({ message, level })
}

function randomByte() {
  const cryptoApi = globalThis.crypto
  if (cryptoApi?.getRandomValues) {
    const bytes = new Uint8Array(1)
    cryptoApi.getRandomValues(bytes)
    return bytes[0]
  }
  return Math.floor(Math.random() * 256)
}

export function generateCompanionCode() {
  let code = ''
  const maxUnbiasedByte = 256 - (256 % COMPANION_CODE_ALPHABET.length)
  while (code.length < COMPANION_CODE_LENGTH) {
    const byte = randomByte()
    if (byte >= maxUnbiasedByte) continue
    code += COMPANION_CODE_ALPHABET[byte % COMPANION_CODE_ALPHABET.length]
  }
  return code
}

function normalizeRendezvous2Url(value) {
  let url = String(value || RENDEZVOUS2).trim()
  if (!url) url = RENDEZVOUS2
  if (!/^https?:\/\//i.test(url)) url = `https://${url}`
  url = url.replace(/\/+$/, '')
  if (!/\/rendezvous2$/i.test(url)) url = `${url}/rendezvous2`
  return `${url}/`
}

function nextChunkSymbol() {
  chunkSequence += 1
  return `Q${chunkSequence}`
}

export function chunkCompanionMessage(input) {
  const message = String(input ?? '')
  if (message.length <= WEBRTC_CHUNK_LENGTH) return [message]

  let remaining = message
  const chunks = []
  while (remaining.length > 0) {
    chunks.push(remaining.slice(0, WEBRTC_CHUNK_LENGTH))
    remaining = remaining.slice(WEBRTC_CHUNK_LENGTH)
  }

  const symbol = nextChunkSymbol()
  const out = [`(define ${symbol} "")`]
  for (let item of chunks) {
    item = item.replace(/\\/g, '\\\\').replace(/"/g, '\\"')
    out.push(`(set! ${symbol} (string-append ${symbol} "${item}"))`)
  }
  out.push(`(eval (read (open-input-string ${symbol})))`)
  out.push(`(set! ${symbol} #!null)`)
  return out
}

export function sendCompanionMessage(channel, input) {
  if (!channel || channel.readyState !== 'open') {
    throw new Error('Companion data channel is not open')
  }
  const chunks = chunkCompanionMessage(input)
  for (const chunk of chunks) channel.send(chunk)
  return chunks.length
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

function canonicalComponentDefType(type) {
  return type === 'Form' ? 'Screen' : type
}

async function knownComponentTypesForExtraction() {
  const names = await listComponents()
  return new Set([...names, 'Screen', 'Form'])
}

export function extractComponentDefs(annSource, knownTypes = KNOWN_COMPONENT_TYPES) {
  const defs = {}
  const typeCounts = {}
  const known = knownTypes instanceof Set ? knownTypes : new Set(knownTypes || [])
  const addDef = (type, id) => {
    if (known.size && !known.has(type)) return
    const defType = canonicalComponentDefType(type)
    if (!defs[defType]) defs[defType] = []
    if (!defs[defType].includes(id)) defs[defType].push(id)
  }

  // Legacy @Type { id: "name" } syntax
  const re = /@([A-Za-z]\w*)\s*\{(?:[^{}@]*?[,\n])?\s*id\s*:\s*"([^"]+)"/g
  let m
  while ((m = re.exec(annSource)) !== null) {
    addDef(m[1], m[2])
  }

  // Current Type.name syntax
  const searchable = maskIgnoredSpans(annSource)
  const dotRe = /(?:^|[{},\n])\s*([A-Z][A-Za-z0-9_]*)(?:\s*\.\s*([A-Za-z_]\w*))?\s*(?=\s*(?:\{|[,}\n]|$))/gm
  while ((m = dotRe.exec(searchable)) !== null) {
    const type = m[1]
    let id = m[2]
    if (!id) {
      typeCounts[type] = (typeCounts[type] || 0) + 1
      id = `${type}${typeCounts[type]}`
    }
    addDef(type, id)
  }

  return defs
}

function findMissingScreens(componentDefs, yail) {
  const screenIds = componentDefs.Screen ?? []
  return screenIds.filter(id => !yailHasToken(yail, id))
}

function extractEventDefs(blocklyXml) {
  const defs = []
  const mutationRe = /<mutation\b[^>]*>/g
  const attrRe = /\b([A-Za-z_][\w:-]*)\s*=\s*"([^"]*)"/g
  let m
  while ((m = mutationRe.exec(blocklyXml)) !== null) {
    const attrs = {}
    let attr
    while ((attr = attrRe.exec(m[0])) !== null) attrs[attr[1]] = attr[2]
    if (attrs.instance_name && attrs.event_name) {
      defs.push({ component: attrs.instance_name, event: attrs.event_name })
    }
  }
  return defs
}

function findMissingEvents(eventDefs, yail) {
  return eventDefs.filter(({ component, event }) => !yailHasDefineEvent(yail, component, event))
}

function escapeRegex(text) {
  return String(text).replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function yailHasToken(yail, token) {
  return new RegExp(`(^|[^A-Za-z0-9_.$-])${escapeRegex(token)}([^A-Za-z0-9_.$-]|$)`).test(String(yail || ''))
}

function yailHasDefineEvent(yail, component, event) {
  return new RegExp(`\\(define-event\\s+${escapeRegex(component)}\\s+${escapeRegex(event)}(?:\\s|\\))`).test(String(yail || ''))
}

export const __companionInternalsForTests = {
  extractEventDefs,
  findMissingEvents,
  findMissingScreens,
}

function wrapForRepl(yail) {
  return `(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 ${yail}))`
}

export function wrapSnippet(yail) {
  return `(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 (begin ${yail})))`
}

function safeScreenName(name = 'Screen1') {
  const clean = String(name || '').trim().replace(/[^A-Za-z0-9_]/g, '_').replace(/_+/g, '_').replace(/^_+|_+$/g, '')
  const fallback = clean || 'Screen1'
  return /^[A-Za-z]/.test(fallback) ? fallback : `Screen_${fallback}`
}

export function normalizeCompanionDesignSource(annSource, screenName = 'Screen1') {
  const source = String(annSource || '')
  const cleanName = safeScreenName(screenName)
  if (!source.trim()) return `Screen.${cleanName} { Title: "${cleanName}" }`
  return source.replace(
    /^(\s*(?:(?:\/\/[^\n]*(?:\n|$))\s*)*(?:Screen|Form))(?:\s*\.\s*[A-Za-z_]\w*)?/,
    `$1.${cleanName}`,
  )
}

async function companionDesignContext(annSource, screenName, { debug = null } = {}) {
  let normalizedAnnSource = normalizeCompanionDesignSource(annSource, screenName)
  let debugNotifierName = null

  if (debug?.enabled) {
    const injected = ensureDebugNotifierDesignSource(normalizedAnnSource, debug.notifierName)
    normalizedAnnSource = injected.designSource
    debugNotifierName = injected.notifierName
  }

  const componentDefs = extractComponentDefs(normalizedAnnSource, await knownComponentTypesForExtraction())
  return { normalizedAnnSource, componentDefs, debugNotifierName }
}

export async function compileForCompanion(mistSource, annSource, { screenName = 'Screen1', debug = null } = {}) {
  const debugEnabled = Boolean(debug?.enabled)
  const sessionId = debug?.sessionId || (debugEnabled ? createDebugSessionId() : null)
  const { normalizedAnnSource, componentDefs, debugNotifierName } = await companionDesignContext(annSource, screenName, { debug })
  const annPreflight = await annToYailDiagnosticResult(normalizedAnnSource, '')
  if (!annPreflight.ok) {
    throw new Error(annPreflight.error || 'Design validation failed')
  }

  let compileSource = mistSource
  let debugInfo = null

  if (debugEnabled) {
    await mistToXml(mistSource, componentDefs)
    const instrumented = instrumentFalconSourceForDebug(mistSource, debug.lineMap || [], {
      sessionId,
      notifierName: debugNotifierName,
      breakpoints: debug.breakpoints || [],
    })
    compileSource = instrumented.source
    debugInfo = {
      sessionId: instrumented.sessionId,
      notifierName: instrumented.notifierName,
      lineMap: instrumented.lineMap,
      tracePoints: instrumented.tracePoints,
      breakpointPoints: instrumented.breakpointPoints,
      expressionCatalog: instrumented.expressionCatalog,
    }
  }

  const blocklyXml    = await mistToXml(compileSource, componentDefs)
  const codeYail      = await blocklyToYail(blocklyXml)
  const eventDefs     = extractEventDefs(blocklyXml)
  const missingEvents = findMissingEvents(eventDefs, codeYail)
  if (missingEvents.length) {
    throw new Error(`Compiled YAIL payload is missing event code for: ${missingEvents.map(({ component, event }) => `${component}.${event}`).join(', ')}`)
  }
  const fullYail      = await annToYail(normalizedAnnSource, codeYail)
  const missingScreens = findMissingScreens(componentDefs, fullYail)
  if (missingScreens.length) {
    throw new Error(`Compiled YAIL payload is missing screen code for: ${missingScreens.join(', ')}`)
  }
  const replPayload   = wrapForRepl(fullYail)
  return { componentDefs, screenIds: componentDefs.Screen ?? [], eventDefs, blocklyXml, codeYail, fullYail, replPayload, debug: debugInfo }
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

export async function compileSnippetForCompanion(mistSnippet, annSource, { screenName = 'Screen1' } = {}) {
  const { normalizedAnnSource, componentDefs } = await companionDesignContext(annSource, screenName)
  const annPreflight = await annToYailDiagnosticResult(normalizedAnnSource, '')
  if (!annPreflight.ok) {
    throw new Error(annPreflight.error || 'Design validation failed')
  }
  const replPayload = await compileSnippet(mistSnippet, componentDefs)
  return { componentDefs, replPayload }
}

export function debugContinueReplPayload() {
  return wrapSnippet(`(set-var! g$${DEBUG_CONTINUE_GLOBAL} #t)`)
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

export async function connectCompanion(code, _digest, config = {}, signal, onLog) {
  const rendezvous2 = normalizeRendezvous2Url(config.rendezvous2)
  const rawServers = config.iceServers ?? config.iceservers ?? DEFAULT_ICE_SERVERS
  if (!Array.isArray(rawServers)) {
    throw new Error('Rendezvous response included malformed ICE server data')
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
    postSignal(rendezvous2, {
      key:       code + '-s',
      webrtc:    true,
      nonce:     Math.floor(Math.random() * 10000) + 1,
      candidate: candidate.toJSON(),
    }, signal, onLog, 'local ICE candidate').catch(() => {})
  }

  try {
    emitLog(onLog, 'WebRTC: creating SDP offer')
    const offer = await peer.createOffer()
    await peer.setLocalDescription(offer)
    emitLog(onLog, 'WebRTC: local offer set')
    await postSignal(rendezvous2, {
      key:       code + '-s',
      webrtc:    true,
      nonce:     Math.floor(Math.random() * 10000) + 1,
      offer,
      candidate: null,
    }, signal, onLog, 'SDP offer')

    await receiveOfferResponse(peer, rendezvous2 + code + '-r', signal, onLog)
    await waitForChannelOpen(channel, signal, onLog)
    return { peer, channel }
  } catch (e) {
    peer.onicecandidate = null
    peer.onicegatheringstatechange = null
    peer.oniceconnectionstatechange = null
    peer.onconnectionstatechange = null
    peer.onsignalingstatechange = null
    try { channel.close() } catch {}
    try { peer.close() } catch {}
    throw e
  }
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
