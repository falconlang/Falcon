<script>
  import { onMount, tick } from 'svelte';
  import {
    cells,
    designCode,
    appendDebugLogsFromCompanionResponse,
    activeScreen,
    getDesignSource,
    liveTestOpen,
    liveTestState,
    companionCommand,
    projectName,
    setDoItResult,
    debugBreakpoints,
    debugModeEnabled,
    debugUserEnabled,
    runtimeErrorNotice,
    enableDebugMode,
    disableDebugMode,
    dismissRuntimeErrorNotice,
    debugBreakpointsForCompile,
    startDebugSession,
    clearDebugActiveLocation,
    clearDebugRuntimeState,
    setDebugTraceLocation,
    setDebugExpressionValue,
    setDebugBreakpointHit,
    continueDebugExecution,
    setDebugRuntimeError,
    debugAnnotationActive,
    debugExecutionState,
  } from './stores.js';
  import {
    compileForCompanion,
    compileSnippetForCompanion,
    connectCompanion,
    debugContinueReplPayload,
    generateCompanionCode,
    pollRendezvous,
    sendCompanionMessage,
  } from './companion.js';
  import {
    DEBUG_IDLE_CLEAR_MS,
    buildFalconSourceMap,
    parseDebugRuntimeEvent,
  } from './debug-source-map.js';
  import {
    exportCurrentProjectToAia,
    sanitizeProjectName,
  } from './appinventor-aia.js';
  import { createQrSvg } from './qr-code.js';

  let liveCode = null;
  let codeCharsEl;
  let qrSvg = '';
  let status = 'idle';
  let error = null;
  let messageCount = 0;
  let connectionLogs = [];
  let logId = 0;

  let abortController = null;
  let peer = null;
  let channel = null;
  let attemptToken = 0;
  let lastSentSourceKey = null;
  let lastFailedSourceKey = null;
  let refreshTimer = null;
  let wasOpen = false;
  let commandRunning = false;
  let pendingDoItCellId = null;
  let debugIdleTimer = null;
  const COMPANION_DEBUG = Boolean(import.meta.env?.DEV);

  $: liveTestState.set({
    status,
    code: liveCode,
    error,
    messageCount,
  });

  function currentFalconSourceMap() {
    return buildFalconSourceMap($cells);
  }

  function companionSourceKey() {
    const { source } = currentFalconSourceMap();
    // Only include user breakpoints when debug mode is user-enabled; debug instrumentation is always active
    const userBreakpoints = $debugUserEnabled ? debugBreakpointsForCompile($activeScreen) : [];
    const breakpointKey = JSON.stringify(userBreakpoints);
    return `${$activeScreen}\0${source}\0${getDesignSource()}\0breakpoints:${breakpointKey}`;
  }

  function debugCompileOptions(sourceMap) {
    // Always compile with debug instrumentation so runtime errors can be located
    return {
      enabled: true,
      lineMap: sourceMap.entries,
      breakpoints: $debugUserEnabled ? debugBreakpointsForCompile($activeScreen) : [],
    };
  }

  async function compileCurrentForCompanion() {
    const sourceMap = currentFalconSourceMap();
    clearDebugRuntimeState();
    const result = await compileForCompanion(sourceMap.source, getDesignSource(), {
      screenName: $activeScreen,
      debug: debugCompileOptions(sourceMap),
    });

    if (result.debug) {
      startDebugSession({
        sessionId: result.debug.sessionId,
        lineMap: result.debug.lineMap,
        expressionCatalog: result.debug.expressionCatalog,
      });
      debugModeEnabled.set(true);
    }

    return result;
  }

  function clearDebugIdleTimer() {
    if (!debugIdleTimer) return;
    clearTimeout(debugIdleTimer);
    debugIdleTimer = null;
  }

  function scheduleDebugIdleClear(sessionId) {
    clearDebugIdleTimer();
    debugIdleTimer = setTimeout(() => {
      clearDebugActiveLocation(sessionId);
      debugIdleTimer = null;
    }, DEBUG_IDLE_CLEAR_MS);
  }

  $: statusText = status === 'connecting' ? 'Negotiating a connection…'
    : status === 'error' ? (error || 'Connection error')
    : 'Waiting for companion…';

  function logTime() {
    const d = new Date();
    return [
      String(d.getHours()).padStart(2, '0'),
      String(d.getMinutes()).padStart(2, '0'),
      String(d.getSeconds()).padStart(2, '0'),
    ].join(':');
  }

  function addConnectionLog(entry, fallbackLevel = 'info') {
    const item = typeof entry === 'string'
      ? { message: entry, level: fallbackLevel }
      : { message: entry?.message || String(entry), level: entry?.level || fallbackLevel };

    connectionLogs = [
      ...connectionLogs,
      {
        id: ++logId,
        time: logTime(),
        level: item.level || 'info',
        message: item.message,
      },
    ].slice(-8);
  }

  function schemeString(value) {
    return `"${String(value ?? '')
      .replace(/\\/g, '\\\\')
      .replace(/"/g, '\\"')
      .replace(/\n/g, '\\n')
      .replace(/\r/g, '\\r')}"`;
  }

  async function blobToBase64(blob) {
    const bytes = new Uint8Array(await blob.arrayBuffer());
    const chunkSize = 0x8000;
    let binary = '';
    for (let i = 0; i < bytes.length; i += chunkSize) {
      binary += String.fromCharCode(...bytes.subarray(i, i + chunkSize));
    }
    return btoa(binary);
  }

  function projectArchiveSaveMessage(filename, base64) {
    return `(begin
  (require <com.google.youngandroid.runtime>)
  (define-alias AIFileUtil <com.google.appinventor.components.runtime.util.FileUtil>)
  (define-alias AIQUtil <com.google.appinventor.components.runtime.util.QUtil>)
  (define-alias AIForm <com.google.appinventor.components.runtime.Form>)
  (define-alias AIBase64 <android.util.Base64>)
  (process-repl-input -1
    (begin
      (AIFileUtil:writeFile
        (AIBase64:decode ${schemeString(base64)} 0)
        (string-append (AIQUtil:getExternalStoragePath (AIForm:getActiveForm) #t) "/Download/${filename}")))))`;
  }

  function resetConnectionLogs(message = null) {
    connectionLogs = [];
    logId = 0;
    if (message) addConnectionLog(message);
  }

  function renderCodeChars(code) {
    if (!codeCharsEl) return;
    codeCharsEl.innerHTML = code.split('').map(char => `<span class="lt-code-char">${char}</span>`).join('');
  }

  function ltRenderQR(code) {
    qrSvg = createQrSvg(code, {
      size: 148,
      fgColor: '#1A1A21',
      bgColor: '#FFFFFF',
    });
  }

  function renderConnectionCode() {
    tick().then(() => {
      if (!liveCode) return;
      renderCodeChars(liveCode);
      ltRenderQR(liveCode);
    });
  }

  function applyDebugRuntimeEvent(value) {
    const event = parseDebugRuntimeEvent(value);
    if (!event) return false;

    if (event.type === 'value') {
      return setDebugExpressionValue(event);
    }

    if (event.type === 'breakpoint-hit') {
      clearDebugIdleTimer();
      return setDebugBreakpointHit(event);
    }

    if (event.type === 'trace') {
      if (setDebugTraceLocation(event)) {
        if ($debugExecutionState.status !== 'paused') scheduleDebugIdleClear(event.sessionId);
        return true;
      }
    }
    return false;
  }

  function applyDebugRuntimeError(value, source = 'Runtime') {
    if (pendingDoItCellId) return false;
    clearDebugIdleTimer();
    const stored = setDebugRuntimeError({
      message: value?.value ?? value?.item ?? value?.message ?? 'Runtime error',
      source: value?.blockid || source,
      status: value?.status || null,
    });
    if (stored && !$debugUserEnabled) {
      runtimeErrorNotice.set({
        show: true,
        error: {
          message: value?.value ?? value?.item ?? value?.message ?? 'Runtime error',
          source: value?.blockid || source,
        },
      });
    }
    return stored;
  }

  function handleCompanionResponse(data) {
    messageCount += 1;
    try {
      const resp = JSON.parse(data);
      if (resp.status !== 'OK') {
        console.error('[companion]', data);
        addConnectionLog('Companion response: non-OK status', 'warn');
        return;
      }
      appendDebugLogsFromCompanionResponse(resp);
      for (const value of resp.values ?? []) {
        if (value.type === 'log') {
          applyDebugRuntimeEvent(value);
          continue;
        }
        if (value.type === 'startCache') {
          addConnectionLog('Companion requested a project cache; saving archive');
          saveProjectToCompanion();
          continue;
        }
        if (value.type === 'assetTransferred') {
          addConnectionLog(`Companion asset transferred: ${value.value || 'asset'}`, 'info');
          continue;
        }
        if (value.type === 'error') {
          console.error('[companion] runtime error:', value.value);
          applyDebugRuntimeError(value, 'Runtime');
          addConnectionLog(`Runtime error: ${value.value}`, 'error');
          continue;
        }
        if (value.status === 'OK') {
          if (pendingDoItCellId) {
            const lines = value.value && value.value !== '*nothing*' ? [value.value] : [];
            setDoItResult(pendingDoItCellId, lines, true);
            pendingDoItCellId = null;
          } else if (value.value && value.value !== '*nothing*') {
            addConnectionLog(`Companion value: ${value.value}`, 'info');
          }
        } else {
          console.error('[companion] eval error:', value.value);
          if (pendingDoItCellId) {
            setDoItResult(pendingDoItCellId, [value.value || 'Evaluation error'], false);
            pendingDoItCellId = null;
          } else {
            applyDebugRuntimeError(value, 'Evaluation');
            addConnectionLog(`Companion eval error: ${value.value}`, 'error');
          }
        }
      }
    } catch {
      console.warn('[companion] non-JSON:', data);
      addConnectionLog('Companion response: non-JSON payload', 'warn');
    }
  }

  function logCompanionPayload(label, result) {
    if (!COMPANION_DEBUG) return;
    const screenList = result.screenIds?.length ? result.screenIds.join(', ') : '(none)';
    const eventList = result.eventDefs?.length
      ? result.eventDefs.map(({ component, event }) => `${component}.${event}`).join(', ')
      : '(none)';
    console.debug(`[companion] ${label} screens in compiled YAIL: ${screenList}`);
    console.debug(`[companion] ${label} events in compiled YAIL: ${eventList}`);
    console.debug(`[companion] ${label} REPL payload sent:\n${result.replPayload}`);
  }

  function cancelAttempt() {
    attemptToken += 1;
    if (abortController) {
      abortController.abort();
      abortController = null;
    }
    if (refreshTimer) {
      clearTimeout(refreshTimer);
      refreshTimer = null;
    }
  }

  function closeTransport() {
    const localChannel = channel;
    const localPeer = peer;
    channel = null;
    peer = null;
    lastSentSourceKey = null;
    lastFailedSourceKey = null;

    if (localChannel) {
      localChannel.onmessage = null;
      localChannel.onerror = null;
      localChannel.onclose = null;
      try { localChannel.close(); } catch {}
    }
    if (localPeer) {
      try { localPeer.close(); } catch {}
    }
  }

  function sendDoneToCompanion(reason = 'disconnect') {
    const activeChannel = channel;
    if (!activeChannel || activeChannel.readyState !== 'open') return false;
    try {
      activeChannel.send('#DONE#');
      addConnectionLog(`${reason}: sent Companion shutdown signal`);
      return true;
    } catch (e) {
      addConnectionLog(`${reason}: unable to send shutdown signal`, 'warn');
      return false;
    }
  }

  function resetDisconnected() {
    cancelAttempt();
    closeTransport();
    clearDebugIdleTimer();
    clearDebugRuntimeState();
    runtimeErrorNotice.set({ show: false, error: null });
    liveCode = null;
    qrSvg = '';
    status = 'idle';
    error = null;
    messageCount = 0;
    resetConnectionLogs();
  }

  function locateRuntimeError() {
    runtimeErrorNotice.set({ show: false, error: null });
    debugUserEnabled.set(true);
  }

  function dismissRuntimeError() {
    dismissRuntimeErrorNotice();
  }

  async function resetConnection() {
    sendDoneToCompanion('Reset connection');
    cancelAttempt();
    closeTransport();
    liveCode = null;
    qrSvg = '';
    status = 'idle';
    error = null;
    messageCount = 0;
    resetConnectionLogs('Reset connection requested');

    try {
      addConnectionLog('aiStarter: requesting /reset');
      await requestAiStarter('reset', 'aiStarter reset');
      addConnectionLog('Connection reset complete');
    } catch (e) {
      addConnectionLog('aiStarter reset was not available; Companion session was closed locally', 'warn');
    }

    liveTestOpen.set(false);
  }

  function close() {
    if (status !== 'connected') resetDisconnected();
    liveTestOpen.set(false);
  }

  function disconnectCompanion() {
    sendDoneToCompanion('Disconnect');
    resetDisconnected();
    liveTestOpen.set(false);
  }

  function sendCompiledResult(label, result, sourceKey, activeChannel = channel) {
    if (channel !== activeChannel || !activeChannel || activeChannel.readyState !== 'open') {
      throw new Error('Companion disconnected before the update could be sent');
    }
    logCompanionPayload(label, result);
    const chunkCount = sendCompanionMessage(activeChannel, result.replPayload);
    lastSentSourceKey = sourceKey;
    addConnectionLog(`${label}: REPL payload sent${chunkCount > 1 ? ` in ${chunkCount} chunks` : ''}`, 'info');
    return chunkCount;
  }

  async function refreshCompanion({ force = false, label = 'manual refresh' } = {}) {
    refreshTimer = null;
    if (status !== 'connected') {
      error = 'Companion is not connected';
      addConnectionLog(`${label}: Companion is not connected`, 'warn');
      return;
    }

    const activeChannel = channel;
    const sourceKey = companionSourceKey();
    if (!activeChannel || activeChannel.readyState !== 'open') {
      error = 'Companion data channel is not open';
      addConnectionLog(`${label}: data channel is not open`, 'error');
      return;
    }
    if (!force && sourceKey === lastSentSourceKey) return;

    try {
      addConnectionLog(`${label}: compiling Falcon and Designer sources`);
      const result = await compileCurrentForCompanion();
      sendCompiledResult(label, result, sourceKey, activeChannel);
      lastFailedSourceKey = null;
      error = null;
      scheduleRefresh();
    } catch (e) {
      console.error('[companion] refresh failed:', e.message || String(e));
      error = e.message || String(e);
      lastFailedSourceKey = sourceKey;
      addConnectionLog(`${label}: ${error}`, 'error');
    }
  }

  function scheduleRefresh() {
    if (status !== 'connected') return;
    if ($debugExecutionState.status === 'paused') return;
    if ($debugAnnotationActive) return;
    const sourceKey = companionSourceKey();
    if (sourceKey === lastSentSourceKey || sourceKey === lastFailedSourceKey) return;
    if (refreshTimer) clearTimeout(refreshTimer);
    refreshTimer = setTimeout(refreshCompanion, 600);
  }

  async function startCompanion() {
    cancelAttempt();
    closeTransport();

    const token = attemptToken + 1;
    attemptToken = token;
    liveCode = generateCompanionCode();
    status = 'polling';
    error = null;
    messageCount = 0;
    lastSentSourceKey = null;
    lastFailedSourceKey = null;
    resetConnectionLogs(`Live test code ${liveCode} generated`);
    addConnectionLog('Compiling Falcon and Designer sources');
    renderConnectionCode();

    const abort = new AbortController();
    abortController = abort;

    try {
      const sourceKey = companionSourceKey();
      const result = await compileCurrentForCompanion();
      if (token !== attemptToken || abort.signal.aborted) return;
      const eventCount = result.eventDefs?.length || 0;
      const screenCount = result.screenIds?.length || 0;
      addConnectionLog(`Compile complete: ${screenCount} screen${screenCount === 1 ? '' : 's'}, ${eventCount} event handler${eventCount === 1 ? '' : 's'}`);
      addConnectionLog('Waiting for MIT rendezvous response');

      const { digest, config } = await pollRendezvous(liveCode, abort.signal, addConnectionLog);
      if (token !== attemptToken || abort.signal.aborted) return;

      status = 'connecting';
      await tick();
      await new Promise(r => requestAnimationFrame(r));
      addConnectionLog('Companion found; starting WebRTC negotiation');
      const conn = await connectCompanion(liveCode, digest, config, abort.signal, addConnectionLog);
      if (token !== attemptToken || abort.signal.aborted) {
        try { conn.channel.close(); } catch {}
        try { conn.peer.close(); } catch {}
        return;
      }

      channel = conn.channel;
      peer = conn.peer;
      abortController = null;

      conn.channel.onmessage = ({ data }) => handleCompanionResponse(data);
      conn.channel.onerror = () => {
        status = 'error';
        error = 'Data channel error';
      };
      conn.channel.onclose = () => {
        if (channel !== conn.channel) return;
        channel = null;
        peer = null;
        lastSentSourceKey = null;
        try { conn.peer.close(); } catch {}
        if (status !== 'error') {
          status = 'idle';
          liveCode = null;
        }
      };

      addConnectionLog('Sending REPL payload to companion');
      sendCompiledResult('initial send', result, sourceKey, conn.channel);
      status = 'connected';
      error = null;
      addConnectionLog('REPL payload accepted by data channel', 'info');
      liveTestOpen.set(false);
      scheduleRefresh();
    } catch (e) {
      if (e.name === 'AbortError' || token !== attemptToken) return;
      console.error('[companion]', e);
      status = 'error';
      error = e.message || String(e);
      addConnectionLog(`Error: ${error}`, 'error');
      renderConnectionCode();
    }
  }

  function retry() {
    startCompanion();
  }

  function requestAiStarter(path, label) {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 7000);
    return fetch(`http://localhost:8004/${path}/`, { signal: controller.signal })
      .then(response => {
        if (!response.ok) throw new Error(`${label} failed with HTTP ${response.status}`);
        return response.text();
      })
      .finally(() => clearTimeout(timeout));
  }

  async function hardResetCompanion() {
    if (typeof window !== 'undefined') {
      const ok = window.confirm('Hard reset the Companion/emulator connection? This resets the local aiStarter session.');
      if (!ok) return;
    }

    sendDoneToCompanion('Hard reset');
    cancelAttempt();
    closeTransport();
    liveCode = null;
    qrSvg = '';
    messageCount = 0;
    status = 'idle';
    error = null;
    resetConnectionLogs('Hard reset requested');

    try {
      addConnectionLog('aiStarter: requesting /reset');
      await requestAiStarter('reset', 'aiStarter reset');
      addConnectionLog('aiStarter: requesting /emulatorreset');
      await requestAiStarter('emulatorreset', 'aiStarter emulator reset');
      addConnectionLog('Hard reset complete');
      liveTestOpen.set(false);
    } catch (e) {
      status = 'error';
      error = e.name === 'AbortError'
        ? 'aiStarter hard reset timed out. Make sure aiStarter is running on port 8004.'
        : (e.message || String(e));
      addConnectionLog(`Hard reset failed: ${error}`, 'error');
      liveTestOpen.set(true);
    }
  }

  async function saveProjectToCompanion() {
    if (commandRunning) return;
    if (status !== 'connected' || !channel || channel.readyState !== 'open') {
      error = 'Start Live test before saving the project';
      addConnectionLog(error, 'warn');
      liveTestOpen.set(true);
      return;
    }

    commandRunning = true;
    const activeChannel = channel;
    try {
      addConnectionLog('Exporting current project archive');
      const blob = await exportCurrentProjectToAia();
      const safeName = sanitizeProjectName($projectName);
      const filename = `${safeName}.aia`;
      const base64 = await blobToBase64(blob);
      const yail = projectArchiveSaveMessage(filename, base64);
      if (channel !== activeChannel || activeChannel.readyState !== 'open') {
        throw new Error('Companion disconnected before the project archive could be saved');
      }
      const chunkCount = sendCompanionMessage(activeChannel, yail);
      addConnectionLog(`Project archive sent to Companion as ${filename}${chunkCount > 1 ? ` in ${chunkCount} chunks` : ''}`);
      error = null;
    } catch (e) {
      error = e.message || String(e);
      addConnectionLog(`Save project failed: ${error}`, 'error');
      console.error('[companion] save project failed:', e);
    } finally {
      commandRunning = false;
    }
  }

  async function doItCompanion(source, label = 'Do it', cellId = null) {
    if (status !== 'connected' || !channel || channel.readyState !== 'open') {
      error = 'Companion is not connected';
      addConnectionLog(`${label}: Companion is not connected`, 'warn');
      liveTestOpen.set(true);
      return;
    }

    pendingDoItCellId = cellId;
    const activeChannel = channel;
    try {
      addConnectionLog(`${label}: compiling selected Falcon code`);
      const result = await compileSnippetForCompanion(source, getDesignSource(), { screenName: $activeScreen });
      if (channel !== activeChannel || activeChannel.readyState !== 'open') {
        throw new Error('Companion disconnected before the selected code could be sent');
      }
      const chunkCount = sendCompanionMessage(activeChannel, result.replPayload);
      addConnectionLog(`${label}: selected code sent${chunkCount > 1 ? ` in ${chunkCount} chunks` : ''}`, 'info');
      error = null;
    } catch (e) {
      if (pendingDoItCellId === cellId) {
        setDoItResult(cellId, [e.message || String(e)], false);
        pendingDoItCellId = null;
      }
      error = e.message || String(e);
      addConnectionLog(`${label}: ${error}`, 'error');
      console.error('[companion] do it failed:', e);
    }
  }

  async function continueDebugCompanion(hitId = null) {
    if (status !== 'connected' || !channel || channel.readyState !== 'open') {
      error = 'Companion is not connected';
      addConnectionLog('Continue: Companion is not connected', 'warn');
      liveTestOpen.set(true);
      return;
    }

    const activeChannel = channel;
    try {
      const payload = debugContinueReplPayload();
      if (channel !== activeChannel || activeChannel.readyState !== 'open') {
        throw new Error('Companion disconnected before continue could be sent');
      }
      sendCompanionMessage(activeChannel, payload);
      continueDebugExecution(hitId);
      addConnectionLog('Continue: breakpoint resumed', 'info');
      error = null;
      scheduleRefresh();
    } catch (e) {
      error = e.message || String(e);
      addConnectionLog(`Continue: ${error}`, 'error');
      console.error('[companion] continue failed:', e);
    }
  }

  async function setCompanionDebugMode(enabled) {
    if (enabled) {
      enableDebugMode();
      // Re-compile only if breakpoints need to be activated
      if (status === 'connected' && debugBreakpointsForCompile($activeScreen).length > 0) {
        await refreshCompanion({ force: true, label: 'debug enabled' });
      }
      return;
    }

    debugUserEnabled.set(false);
    clearDebugRuntimeState();
    runtimeErrorNotice.set({ show: false, error: null });
    // Re-compile if breakpoints need to be deactivated
    if (status === 'connected' && debugBreakpointsForCompile($activeScreen).length > 0) {
      await refreshCompanion({ force: true, label: 'debug disabled' });
    }
  }

  async function runCompanionCommand(cmd) {
    const action = typeof cmd === 'string' ? cmd : cmd?.type || cmd?.action;
    if (action === 'do-it') {
      await doItCompanion(cmd?.source || '', cmd?.label || 'Do it', cmd?.cellId || null);
    } else if (action === 'debug-continue') {
      await continueDebugCompanion(cmd?.hitId || null);
    } else if (action === 'toggle-debug') {
      await setCompanionDebugMode(!$debugUserEnabled);
    } else if (action === 'debug-enable') {
      await setCompanionDebugMode(true);
    } else if (action === 'debug-disable') {
      await setCompanionDebugMode(false);
    } else if (action === 'refresh') {
      await refreshCompanion({ force: true, label: 'manual refresh' });
    } else if (action === 'reset') {
      await resetConnection();
    } else if (action === 'hard-reset') {
      await hardResetCompanion();
    } else if (action === 'save') {
      await saveProjectToCompanion();
    }
  }

  function backdropClick(e) {
    if (e.target === e.currentTarget) close();
  }

  $: if ($companionCommand) {
    const cmd = $companionCommand;
    companionCommand.set(null);
    runCompanionCommand(cmd);
  }

  onMount(() => {
    const unsubCells = cells.subscribe(() => scheduleRefresh());
    const unsubDesign = designCode.subscribe(() => scheduleRefresh());
    const unsubDebug = debugUserEnabled.subscribe(() => scheduleRefresh());
    const unsubBreakpoints = debugBreakpoints.subscribe(() => scheduleRefresh());

    return () => {
      unsubCells();
      unsubDesign();
      unsubDebug();
      unsubBreakpoints();
      clearDebugIdleTimer();
      resetDisconnected();
    };
  });

  $: {
    const open = $liveTestOpen;
    if (open && !wasOpen) {
      if (status === 'idle' || status === 'error') startCompanion();
      else renderConnectionCode();
    }
    if (!open && wasOpen && status !== 'connected') resetDisconnected();
    wasOpen = open;
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div
  id="live-test-overlay"
  class:show={$liveTestOpen}
  on:click={backdropClick}
>
  <div class="lt-card">

    <!-- Shared header -->
    <div class="lt-header">
      <div class="lt-header-icon" aria-hidden="true">
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
          <rect x="3.5" y="1" width="9" height="14" rx="2"/>
          <circle cx="8" cy="12.5" r="0.9" fill="currentColor" stroke="none"/>
        </svg>
      </div>
      <span class="lt-title">Live test</span>
      <button class="lt-close-btn" on:click={status === 'connected' ? disconnectCompanion : close} title={status === 'connected' ? 'Disconnect' : 'Cancel'}>
        <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
          <path d="M2 2l6 6M8 2l-6 6"/>
        </svg>
      </button>
    </div>

    {#if status === 'connected'}

      <!-- Connected body -->
      <div class="lt-success-area">
        <div class="lt-success-ring" aria-hidden="true">
          <svg viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 10l4.5 4.5 8-8"/>
          </svg>
        </div>
        <div class="lt-success-text">
          <div class="lt-connected-label">Connected</div>
          <div class="lt-connected-desc">Code changes reflect on your device in real time.</div>
          {#if messageCount > 0}
            <div class="lt-msg-count">{messageCount} update{messageCount === 1 ? '' : 's'} sent</div>
          {/if}
        </div>
      </div>

      <div class="lt-footer">
        <button class="lt-btn lt-btn--danger" on:click={disconnectCompanion}>
          <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
            <path d="M4.5 6h5M7 3.5L4 6l3 2.5"/>
            <path d="M2 2v8"/>
          </svg>
          Disconnect
        </button>
      </div>

    {:else}

      <!-- QR + digits body -->
      <div class="lt-qr-section">
        <div class="lt-qr-box">{@html qrSvg}</div>
        <div class="lt-scan-hint">Scan with MIT AI2 Companion</div>
      </div>

      <div class="lt-or-row">
        <span class="lt-or-line"></span>
        <span class="lt-or-text">or enter code</span>
        <span class="lt-or-line"></span>
      </div>

      <div class="lt-code-display" bind:this={codeCharsEl}></div>

      <div class="lt-status lt-status--{status}" aria-live="polite">
        <span class="lt-dot"></span>
        <span>{statusText}</span>
      </div>

      {#if status === 'error'}
        <div class="lt-footer">
          <button class="lt-btn lt-btn--accent" on:click={retry}>
            <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M10 4a4 4 0 10.6 2M10 1v3H7"/></svg>
            Try Again
          </button>
        </div>
      {/if}

    {/if}
  </div>
</div>

{#if $runtimeErrorNotice.show}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="re-backdrop" role="dialog" aria-modal="true" on:click|self={dismissRuntimeError}>
    <div class="re-card">

      <div class="re-header">
        <div class="re-header-icon" aria-hidden="true">
          <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M6 1.5L11 10H1L6 1.5z"/>
            <path d="M6 5v2.2"/>
            <circle cx="6" cy="9" r="0.55" fill="currentColor" stroke="none"/>
          </svg>
        </div>
        <span class="re-title">Runtime Error</span>
        <button class="lt-close-btn" on:click={dismissRuntimeError} title="Dismiss">
          <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round">
            <path d="M2 2l6 6M8 2l-6 6"/>
          </svg>
        </button>
      </div>

      <div class="re-body">
        {#if $runtimeErrorNotice.error?.message}
          <div class="re-message">{$runtimeErrorNotice.error.message}</div>
        {:else}
          <div class="re-message re-message--fallback">An unexpected error occurred at runtime.</div>
        {/if}
        {#if $runtimeErrorNotice.error?.source && $runtimeErrorNotice.error.source !== 'Runtime'}
          <div class="re-source">from {$runtimeErrorNotice.error.source}</div>
        {/if}
      </div>

      <div class="re-footer">
        <button class="lt-btn lt-btn--re-locate" on:click={locateRuntimeError}>
          <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="7" cy="5" r="2.5"/><path d="M4 13c0-1.657 1.343-3 3-3s3 1.343 3 3"/><path d="M1 5h2M11 5h2M7 1v1M7 8v1"/><circle cx="7" cy="5" r="1" fill="currentColor" stroke="none"/></svg>
          Debug error
        </button>
        <button class="lt-btn" on:click={dismissRuntimeError}>Dismiss</button>
      </div>

    </div>
  </div>
{/if}
