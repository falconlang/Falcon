<script>
  import { tick } from 'svelte';
  import { setActive, moveCellById, deleteCellById, sourceNavigationHighlight } from './stores.js';

  export let cell;
  export let active = false;

  let handledSourceHighlightToken = null;
  let searchFlash = false;

  // Reveal + flash when universal search navigates to this markdown cell.
  $: if (
    $sourceNavigationHighlight?.cellId === cell.id
    && $sourceNavigationHighlight?.token !== handledSourceHighlightToken
  ) {
    handledSourceHighlightToken = $sourceNavigationHighlight.token;
    revealCell();
  }

  async function revealCell() {
    await tick();
    document.getElementById(`cell-${cell.id}`)?.scrollIntoView({ block: 'center', behavior: 'smooth' });
    searchFlash = false;
    await tick();
    searchFlash = true;
    setTimeout(() => { searchFlash = false; }, 1400);
  }

  const ALLOWED_TAGS = new Set([
    'a', 'br', 'code', 'div', 'em', 'h1', 'h2', 'h3', 'h4',
    'li', 'ol', 'p', 'pre', 'span', 'strong', 'ul',
  ]);
  const ALLOWED_ATTRS = {
    a: new Set(['href', 'rel', 'target', 'title']),
    div: new Set(['class']),
    span: new Set(['class']),
    code: new Set(['class']),
    pre: new Set(['class']),
  };

  function escapeHtml(value) {
    return String(value ?? '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function safeHref(value) {
    const raw = String(value || '').trim();
    if (!raw) return '';
    if (raw.startsWith('#')) return raw;
    if (raw.startsWith('/') && !raw.startsWith('//') && !raw.startsWith('/\\')) return raw;
    try {
      const url = new URL(raw, window.location.href);
      return ['http:', 'https:', 'mailto:', 'tel:'].includes(url.protocol) ? raw : '';
    } catch {
      return '';
    }
  }

  function sanitizeNode(node) {
    if (node.nodeType === Node.TEXT_NODE) return escapeHtml(node.textContent || '');
    if (node.nodeType !== Node.ELEMENT_NODE) return '';

    const tag = node.tagName.toLowerCase();
    const children = Array.from(node.childNodes).map(sanitizeNode).join('');
    if (!ALLOWED_TAGS.has(tag)) return children;

    const allowedAttrs = ALLOWED_ATTRS[tag] || new Set();
    const attrs = [];
    for (const attr of Array.from(node.attributes)) {
      const name = attr.name.toLowerCase();
      if (name.startsWith('on') || !allowedAttrs.has(name)) continue;
      if (name === 'href') {
        const href = safeHref(attr.value);
        if (href) attrs.push(`href="${escapeHtml(href)}"`);
        continue;
      }
      if (name === 'target') {
        if (attr.value === '_blank') attrs.push('target="_blank"');
        continue;
      }
      if (name === 'rel') {
        attrs.push('rel="noopener noreferrer"');
        continue;
      }
      if (name === 'class' && !/^[A-Za-z0-9_\-\s]+$/.test(attr.value)) continue;
      attrs.push(`${name}="${escapeHtml(attr.value)}"`);
    }

    if (tag === 'a' && attrs.some(attr => attr === 'target="_blank"') && !attrs.some(attr => attr.startsWith('rel='))) {
      attrs.push('rel="noopener noreferrer"');
    }

    const attrText = attrs.length ? ` ${attrs.join(' ')}` : '';
    return tag === 'br' ? '<br>' : `<${tag}${attrText}>${children}</${tag}>`;
  }

  function sanitizeHtml(value) {
    if (typeof document === 'undefined' || typeof Node === 'undefined') return escapeHtml(value);
    const template = document.createElement('template');
    template.innerHTML = String(value || '');
    return Array.from(template.content.childNodes).map(sanitizeNode).join('');
  }

  function activateCellFromKeyboard(e) {
    if (e.target !== e.currentTarget) return;
    if (e.key !== 'Enter' && e.key !== ' ') return;
    e.preventDefault();
    setActive(cell.id);
  }

  $: safeContent = sanitizeHtml(cell.content);
</script>

<div
  class="cell md-cell"
  class:active
  class:search-flash={searchFlash}
  id="cell-{cell.id}"
  role="button"
  tabindex="0"
  on:click={() => setActive(cell.id)}
  on:keydown={activateCellFromKeyboard}
>
  <div class="cell-gutter">
    <button class="gutter-btn" title="Move up" on:click|stopPropagation={() => moveCellById(cell.id, -1)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 8l4-4 4 4"/></svg>
    </button>
    <button class="gutter-btn" title="Move down" on:click|stopPropagation={() => moveCellById(cell.id, 1)}>
      <svg viewBox="0 0 12 12" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M2 4l4 4 4-4"/></svg>
    </button>
  </div>
  <div class="md-content">{@html safeContent}</div>
</div>
