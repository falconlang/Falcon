<script>
  import { liveTestOpen, liveTestState } from './stores.js';

  const LIVE_TEST_LABELS = {
    idle: 'Live Test',
    polling: 'Waiting...',
    connecting: 'Connecting...',
    connected: 'Connected',
    error: 'Error',
  };

  const LIVE_TEST_TITLES = {
    idle: 'Connect Companion',
    polling: 'Waiting for Companion',
    connecting: 'Negotiating connection',
    connected: 'Companion connected',
    error: 'Companion error',
  };

  function toggleTheme() {
    const isDark = document.documentElement.dataset.theme === 'dark';
    const next = isDark ? 'light' : 'dark';
    document.documentElement.classList.add('theme-transitioning');
    document.documentElement.dataset.theme = next;
    localStorage.setItem('tensor-theme', next);
    updateIcon();
    setTimeout(() => document.documentElement.classList.remove('theme-transitioning'), 300);
  }

  function updateIcon() {
    isDark = document.documentElement.dataset.theme === 'dark';
  }

  let isDark = document.documentElement.dataset.theme === 'dark';

  function openLiveTest() {
    liveTestOpen.set(true);
  }

  $: liveTestStatus = $liveTestState.status || 'idle';
  $: liveTestLabel = LIVE_TEST_LABELS[liveTestStatus] || LIVE_TEST_LABELS.idle;
  $: liveTestTitle = LIVE_TEST_TITLES[liveTestStatus] || LIVE_TEST_TITLES.idle;
</script>

<div id="topbar">
  <a class="logo" href="#">
    <div class="logo-mark">
      <svg viewBox="0 0 14 14" fill="none">
        <rect x="1" y="1" width="5" height="5" rx="1" fill="white" fill-opacity="0.9"/>
        <rect x="8" y="1" width="5" height="5" rx="1" fill="white" fill-opacity="0.5"/>
        <rect x="1" y="8" width="5" height="5" rx="1" fill="white" fill-opacity="0.5"/>
        <rect x="8" y="8" width="5" height="5" rx="1" fill="white" fill-opacity="0.8"/>
      </svg>
    </div>
    tensor
  </a>
  <div class="divider"></div>
  <input class="file-name" value="falcon_tour.fc" spellcheck="false" />
  <div class="tb-spacer"></div>
  <div class="topbar-actions">
    <button
      class="tb-btn live-test-chip live-test-chip--{liveTestStatus}"
      title={liveTestTitle}
      aria-label={liveTestTitle}
      on:click={openLiveTest}
    >
      {#if liveTestStatus === 'connected'}
        <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 7l2.5 2.5L11 4"/></svg>
      {:else if liveTestStatus === 'connecting'}
        <svg class="live-test-spin" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="7" cy="7" r="4.5" stroke-opacity="0.25"/><path d="M11.5 7A4.5 4.5 0 007 2.5"/></svg>
      {:else if liveTestStatus === 'polling'}
        <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="7" cy="7" r="4.5"/><path d="M7 4.5V7l1.8 1.3"/></svg>
      {:else if liveTestStatus === 'error'}
        <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M7 2l5 9H2l5-9z"/><path d="M7 5.4v2.4M7 10h.01"/></svg>
      {:else}
        <svg viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="1" width="8" height="12" rx="1.5"/><path d="M5 10.5h4"/><path d="M7 1V0M4 0h6"/></svg>
      {/if}
      {liveTestLabel}
    </button>
    <button class="tb-btn" title="Share">
      <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 3l3 3-3 3M13 6H6a3 3 0 000 6h1"/></svg>
      Share
    </button>
    <button class="tb-btn" title={isDark ? 'Switch to light mode' : 'Switch to dark mode'} on:click={toggleTheme}>
      {#if isDark}
        <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="8" cy="8" r="3"/><path d="M8 2v1M8 13v1M2 8h1M13 8h1M3.7 3.7l.7.7M11.6 11.6l.7.7M3.7 12.3l.7-.7M11.6 4.4l.7-.7"/></svg>
      {:else}
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
      {/if}
    </button>
  </div>
</div>
