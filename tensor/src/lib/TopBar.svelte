<script>
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
