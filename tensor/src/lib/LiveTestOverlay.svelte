<script>
  import { onMount, tick } from 'svelte';
  import {
    cells,
    designCode,
    getDesignSource,
    getFalconSource,
    liveTestOpen,
    liveTestState,
  } from './stores.js';
  import {
    compileForCompanion,
    connectCompanion,
    pollRendezvous,
  } from './companion.js';
  import { createQrSvg } from './qr-code.js';

  let liveCode = null;
  let digitsEl;
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
  let refreshTimer = null;
  let wasOpen = false;

  $: liveTestState.set({
    status,
    code: liveCode,
    error,
    messageCount,
  });

  function genCode() {
    return String(Math.floor(10000 + Math.random() * 90000));
  }

  function companionSourceKey() {
    return `${getFalconSource()}\0${getDesignSource()}`;
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

  function resetConnectionLogs(message = null) {
    connectionLogs = [];
    logId = 0;
    if (message) addConnectionLog(message);
  }

  function ltRenderDigits(code) {
    if (!digitsEl) return;
    digitsEl.innerHTML = code.split('').map(d => `<span class="lt-digit">${d}</span>`).join('');
  }

  function ltRenderQR(code) {
    qrSvg = createQrSvg(code, {
      size: 148,
      fgColor: '#1A1916',
      bgColor: '#ffffff',
    });
  }

  function renderConnectionCode() {
    tick().then(() => {
      if (!liveCode) return;
      ltRenderDigits(liveCode);
      ltRenderQR(liveCode);
    });
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
      for (const value of resp.values ?? []) {
        if (value.status === 'OK') {
          if (value.value && value.value !== '*nothing*') {
            console.log('[companion] ->', value.value);
            addConnectionLog(`Companion value: ${value.value}`, 'success');
          }
        } else {
          console.error('[companion] eval error:', value.value);
          addConnectionLog(`Companion eval error: ${value.value}`, 'error');
        }
      }
    } catch {
      console.warn('[companion] non-JSON:', data);
      addConnectionLog('Companion response: non-JSON payload', 'warn');
    }
  }

  function logCompanionPayload(label, result) {
    const screenList = result.screenIds?.length ? result.screenIds.join(', ') : '(none)';
    const eventList = result.eventDefs?.length
      ? result.eventDefs.map(({ component, event }) => `${component}.${event}`).join(', ')
      : '(none)';
    console.log(`[companion] ${label} screens in compiled YAIL: ${screenList}`);
    console.log(`[companion] ${label} events in compiled YAIL: ${eventList}`);
    console.log(`[companion] ${label} REPL payload sent:\n${result.replPayload}`);
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

  function resetDisconnected() {
    cancelAttempt();
    closeTransport();
    liveCode = null;
    qrSvg = '';
    status = 'idle';
    error = null;
    messageCount = 0;
    resetConnectionLogs();
  }

  function close() {
    if (status !== 'connected') resetDisconnected();
    liveTestOpen.set(false);
  }

  function disconnectCompanion() {
    resetDisconnected();
    liveTestOpen.set(false);
  }

  async function refreshCompanion() {
    refreshTimer = null;
    if (status !== 'connected') return;

    const activeChannel = channel;
    const sourceKey = companionSourceKey();
    if (!activeChannel || activeChannel.readyState !== 'open' || sourceKey === lastSentSourceKey) return;

    try {
      const result = await compileForCompanion(getFalconSource(), getDesignSource());
      if (channel === activeChannel && activeChannel.readyState === 'open') {
        logCompanionPayload('auto-refresh', result);
        activeChannel.send(result.replPayload);
        lastSentSourceKey = sourceKey;
      }
    } catch (e) {
      console.error('[companion] auto-refresh failed:', e.message || String(e));
      error = e.message || String(e);
    }

    scheduleRefresh();
  }

  function scheduleRefresh() {
    if (status !== 'connected') return;
    if (companionSourceKey() === lastSentSourceKey) return;
    if (refreshTimer) clearTimeout(refreshTimer);
    refreshTimer = setTimeout(refreshCompanion, 600);
  }

  async function startCompanion() {
    cancelAttempt();
    closeTransport();

    const token = attemptToken + 1;
    attemptToken = token;
    liveCode = genCode();
    status = 'polling';
    error = null;
    messageCount = 0;
    lastSentSourceKey = null;
    resetConnectionLogs(`Live test code ${liveCode} generated`);
    addConnectionLog('Compiling Falcon and Designer sources');
    renderConnectionCode();

    const abort = new AbortController();
    abortController = abort;

    try {
      const sourceKey = companionSourceKey();
      const result = await compileForCompanion(getFalconSource(), getDesignSource());
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

      logCompanionPayload('initial send', result);
      addConnectionLog('Sending REPL payload to companion');
      conn.channel.send(result.replPayload);
      lastSentSourceKey = sourceKey;
      status = 'connected';
      error = null;
      addConnectionLog('REPL payload accepted by data channel', 'success');
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

  function backdropClick(e) {
    if (e.target === e.currentTarget) close();
  }

  onMount(() => {
    const unsubCells = cells.subscribe(() => scheduleRefresh());
    const unsubDesign = designCode.subscribe(() => scheduleRefresh());

    return () => {
      unsubCells();
      unsubDesign();
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
    <div class="lt-title">{status === 'connected' ? 'Companion Connected' : 'Live Test'}</div>
    {#if status === 'connected'}
      <div class="lt-connected-icon" aria-hidden="true">
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 8l3 3 7-7"/></svg>
      </div>
      <p class="lt-sub lt-centered">Live link active. Source edits will refresh on the device.</p>
      <div class="lt-status status-connected" aria-live="polite">
        <span class="lt-dot"></span>
        {statusText}
      </div>
      <button class="lt-cancel lt-disconnect" on:click={disconnectCompanion}>
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M2 3h8M4 6h4M5 9h2"/></svg>
        Disconnect
      </button>
    {:else}
      <p class="lt-sub">Scan with MIT AI2 Companion or enter the code manually.</p>
      <div class="lt-qr-wrap">
        <div class="lt-qr-box">{@html qrSvg}</div>
      </div>
      <div class="lt-code-label">or enter code</div>
      <div class="lt-code-display" bind:this={digitsEl}></div>
      <div class="lt-status status-{status}" aria-live="polite">
        <span class="lt-dot"></span>
        <div class="lt-status-primary">{statusText}</div>
      </div>
      {#if status === 'error'}
        <button class="lt-cancel lt-retry" on:click={retry}>
          <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.6"><path d="M10 4a4 4 0 10.6 2M10 1v3H7"/></svg>
          Try Again
        </button>
      {/if}
      <button class="lt-cancel" on:click={close}>
        <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M2 2l8 8M10 2l-8 8"/></svg>
        Cancel
      </button>
    {/if}
  </div>
</div>
