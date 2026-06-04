<script>
  import { tick } from 'svelte';
  import { fade } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import {
    cells,
    notebookMode,
    designCode,
    unifiedSearchSource,
    designerSearchIndex,
    designerTreeActive,
    searchOpen,
    closeSearch,
    navigateToCellLine,
    requestSearchNavigation,
  } from './stores.js';

  const GROUP_CAP = 50;
  const DISPLAY_MAX = 220;

  // All icons share viewBox 0 0 12 12, stroke=currentColor, stroke-width 1.5.
  const ICONS = {
    cell: `<rect x="1.5" y="2" width="9" height="8" rx="1.2"/><path d="M1.5 4.3h9" />`,
    'cell-md': `<path d="M2 3h8M2 6h6M2 9h7" stroke-linecap="round"/>`,
    script: `<path d="M2 2.5h8M2 5.5h5.5M2 8.5h8M2 11h3.5" stroke-linecap="round"/>`,
    'designer-text': `<path d="M2 3h8M2 6h7M2 9h5" stroke-linecap="round"/>`,
    component: `<rect x="2" y="2" width="8" height="8" rx="1.2"/>`,
    property: `<circle cx="4" cy="6" r="2.1"/><path d="M6.1 6H10M8.5 6v2" stroke-linecap="round" stroke-linejoin="round"/>`,
  };

  function iconSvg(name) {
    return `<svg class="search-result-icon" viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5">${ICONS[name] || ''}</svg>`;
  }

  let query = '';
  let inputEl;
  let activeIndex = 0;
  let wasOpen = false;

  $: if ($searchOpen && !wasOpen) {
    wasOpen = true;
    query = '';
    activeIndex = 0;
    tick().then(() => inputEl?.focus());
  }
  $: if (!$searchOpen && wasOpen) wasOpen = false;

  function stripHtml(html) {
    if (typeof document === 'undefined') return String(html || '').replace(/<[^>]*>/g, ' ');
    const tmp = document.createElement('div');
    tmp.innerHTML = String(html || '');
    return (tmp.textContent || '').replace(/\s+/g, ' ').trim();
  }

  function display(raw) {
    const s = String(raw ?? '').replace(/^\s+/, '');
    const clipped = s.length > DISPLAY_MAX ? `${s.slice(0, DISPLAY_MAX)}…` : s;
    return clipped || '(blank line)';
  }

  // Split a string into [{text, hit}] segments highlighting every case-insensitive
  // occurrence of the query.
  function segments(text, q) {
    const out = [];
    if (!q) return [{ text, hit: false }];
    const lower = text.toLowerCase();
    const ql = q.toLowerCase();
    let i = 0;
    while (i < text.length) {
      const found = lower.indexOf(ql, i);
      if (found === -1) { out.push({ text: text.slice(i), hit: false }); break; }
      if (found > i) out.push({ text: text.slice(i, found), hit: false });
      out.push({ text: text.slice(found, found + ql.length), hit: true });
      i = found + ql.length;
    }
    return out;
  }

  function scanLines(src, ql, cb) {
    const lines = String(src || '').split('\n');
    let offset = 0;
    for (let i = 0; i < lines.length; i += 1) {
      const line = lines[i];
      const col = line.toLowerCase().indexOf(ql);
      if (col !== -1) cb(line, i + 1, offset, col);
      offset += line.length + 1;
    }
  }

  function buildResults(rawQuery) {
    const q = rawQuery;
    const ql = q.toLowerCase();
    const groups = [];

    // ── Functionality ──────────────────────────────────────────────
    const funcItems = [];
    if ($notebookMode === 'unified') {
      const src = $unifiedSearchSource || '';
      scanLines(src, ql, (line, lineNo, start, col) => {
        const sourceStart = start + col;
        funcItems.push({
          icon: 'script',
          primary: display(line),
          secondary: `line ${lineNo}`,
          navigate: () => requestSearchNavigation({ scope: 'script', sourceStart, sourceEnd: sourceStart + q.length }),
        });
      });
      groups.push({ key: 'functionality', label: 'Script', items: funcItems });
    } else {
      $cells.forEach((cell, idx) => {
        if (cell.type === 'code') {
          String(cell.code || '').split('\n').forEach((line, i) => {
            if (line.toLowerCase().includes(ql)) {
              funcItems.push({
                icon: 'cell',
                primary: display(line),
                secondary: `cell ${idx + 1} · line ${i + 1}`,
                navigate: () => navigateToCellLine(cell.id, i + 1),
              });
            }
          });
        } else if (cell.type === 'markdown') {
          const text = stripHtml(cell.content);
          if (text.toLowerCase().includes(ql)) {
            funcItems.push({
              icon: 'cell-md',
              primary: display(text),
              secondary: `cell ${idx + 1} · text`,
              navigate: () => navigateToCellLine(cell.id, 1),
            });
          }
        }
      });
      groups.push({ key: 'functionality', label: 'Cells', items: funcItems });
    }

    // ── Designer ───────────────────────────────────────────────────
    const desItems = [];
    if ($designerTreeActive) {
      if ($designerSearchIndex.length) {
        for (const comp of $designerSearchIndex) {
          const nameHit = comp.name && comp.name.toLowerCase().includes(ql);
          const typeHit = comp.type && comp.type.toLowerCase().includes(ql);
          if (nameHit || typeHit) {
            desItems.push({
              icon: 'component',
              primary: comp.name || comp.type,
              secondary: comp.type,
              navigate: () => requestSearchNavigation({ scope: 'designer-tree', pathId: comp.pathId }),
            });
          }
          for (const p of comp.props || []) {
            const valueText = p.value == null ? '' : String(p.value);
            const pNameHit = p.name.toLowerCase().includes(ql);
            const pValHit = valueText.toLowerCase().includes(ql);
            if (pNameHit || pValHit) {
              const primary = valueText ? `${p.name}: ${valueText}` : p.name;
              desItems.push({
                icon: 'property',
                primary: display(primary),
                secondary: `${comp.name} · ${comp.type}`,
                navigate: () => requestSearchNavigation({
                  scope: 'designer-tree',
                  pathId: comp.pathId,
                  propName: p.name,
                  category: p.category,
                }),
              });
            }
          }
        }
        groups.push({ key: 'designer', label: 'Designer · Tree', items: desItems });
      } else {
        groups.push({ key: 'designer', label: 'Designer · Tree', items: [], note: 'Designer unavailable (parse error)' });
      }
    } else {
      scanLines($designCode || '', ql, (line, lineNo, start, col) => {
        const startOffset = start + col;
        desItems.push({
          icon: 'designer-text',
          primary: display(line),
          secondary: `line ${lineNo}`,
          navigate: () => requestSearchNavigation({ scope: 'designer-text', start: startOffset, end: startOffset + q.length }),
        });
      });
      groups.push({ key: 'designer', label: 'Designer · Text', items: desItems });
    }

    // ── Assign flat indices (in display order) for keyboard nav ──────
    let flat = [];
    for (const g of groups) {
      g.total = g.items.length;
      g.shown = g.items.slice(0, GROUP_CAP);
      g.overflow = Math.max(0, g.total - GROUP_CAP);
      for (const item of g.shown) {
        item.flatIndex = flat.length;
        flat.push(item);
      }
    }
    return { groups, flat };
  }

  // List the store dependencies explicitly so results recompute if any source
  // changes (buildResults reads them inside a function call, which Svelte's
  // reactivity does not track on its own).
  $: result = (
    $cells, $notebookMode, $unifiedSearchSource,
    $designerSearchIndex, $designerTreeActive, $designCode, query,
    query.trim() ? buildResults(query.trim()) : null
  );
  $: flat = result?.flat ?? [];
  $: totalMatches = flat.length;
  // Reset the active row to the top whenever the query text changes.
  $: { query; activeIndex = 0; }
  $: if (activeIndex >= flat.length) activeIndex = Math.max(0, flat.length - 1);

  // Keep the active row in view.
  $: if (result) scrollActiveIntoView(activeIndex);
  function scrollActiveIntoView(idx) {
    tick().then(() => {
      document.querySelector(`.search-results [data-result-index="${idx}"]`)
        ?.scrollIntoView({ block: 'nearest' });
    });
  }

  function dismiss() {
    closeSearch();
    document.getElementById('toolbar-search-btn')?.focus();
  }

  function go(item) {
    if (!item) return;
    closeSearch();
    tick().then(() => item.navigate());
  }

  function cardIn(node) {
    return {
      duration: 150,
      easing: cubicOut,
      css: (t, u) => `opacity:${t};transform:scale(${0.96 + 0.04 * t}) translateY(${u * 5}px)`,
    };
  }
  function cardOut(node) {
    return {
      duration: 100,
      easing: cubicOut,
      css: (t) => `opacity:${t};transform:scale(${0.97 + 0.03 * t})`,
    };
  }

  function onInputKey(e) {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (flat.length) activeIndex = (activeIndex + 1) % flat.length;
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (flat.length) activeIndex = (activeIndex - 1 + flat.length) % flat.length;
    } else if (e.key === 'Enter') {
      e.preventDefault();
      go(flat[activeIndex]);
    } else if (e.key === 'Escape') {
      e.preventDefault();
      dismiss();
    }
  }
</script>

<!-- svelte-ignore a11y-click-events-have-key-events -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
{#if $searchOpen}
  <div class="sd-overlay" transition:fade={{ duration: 160 }} on:click|self={dismiss}>
    <div class="sd-card sd-card--search" role="dialog" aria-modal="true" aria-label={'Universal search'}
      in:cardIn out:cardOut>
      <div class="search-input-row">
        <svg class="search-input-icon" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5">
          <circle cx="6" cy="6" r="4" /><path d="M10 10l2.5 2.5" stroke-linecap="round" />
        </svg>
        <input
          class="search-input"
          type="text"
          placeholder="Search cells, script, and designer on this screen…"
          bind:value={query}
          bind:this={inputEl}
          on:keydown={onInputKey}
          spellcheck="false"
          autocomplete="off"
          aria-label="Search query"
        />
        {#if query.trim()}
          <span class="search-count">{totalMatches}</span>
        {/if}
        <button class="search-close" on:click={dismiss} title="Close">
          <svg viewBox="0 0 10 10" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M2 2l6 6M8 2l-6 6" /></svg>
        </button>
      </div>

      <div class="search-results" role="listbox" aria-label={'Search results'}>
        {#if !query.trim()}
          <div class="search-hint">Search cells, script, and designer</div>
        {:else if totalMatches === 0 && !result.groups.some(g => g.note)}
          <div class="search-hint">No results for "{query.trim()}"</div>
        {:else}
          {#each result.groups as group (group.key)}
            {#if group.shown.length || group.note}
              <div class="search-group-label">{group.label}{#if group.total} <span class="search-group-count">{group.total}</span>{/if}</div>
              {#if group.note}
                <div class="search-note">{group.note}</div>
              {/if}
              {#each group.shown as item (item.flatIndex)}
                <button
                  type="button"
                  class="search-result"
                  class:active={item.flatIndex === activeIndex}
                  data-result-index={item.flatIndex}
                  role="option"
                  aria-selected={item.flatIndex === activeIndex}
                  on:mousemove={() => (activeIndex = item.flatIndex)}
                  on:click={() => go(item)}
                >
                  {@html iconSvg(item.icon)}
                  <span class="search-result-text">{#each segments(item.primary, query.trim()) as seg}{#if seg.hit}<b>{seg.text}</b>{:else}{seg.text}{/if}{/each}</span>
                  <span class="search-result-meta">{item.secondary}</span>
                </button>
              {/each}
              {#if group.overflow}
                <div class="search-more">+{group.overflow} more {group.overflow === 1 ? 'match' : 'matches'} — refine your query</div>
              {/if}
            {/if}
          {/each}
        {/if}
      </div>

      <div class="search-footer">
        <span class="search-kbd-hints">
          <span><kbd>↑↓</kbd> navigate</span>
          <span><kbd>↵</kbd> jump</span>
          <span><kbd>esc</kbd> close</span>
        </span>
        {#if totalMatches > 0}
          <span class="search-footer-count">{totalMatches} {totalMatches === 1 ? 'match' : 'matches'}</span>
        {/if}
      </div>
    </div>
  </div>
{/if}
