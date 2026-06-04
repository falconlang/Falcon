const LABEL_ALLOWED_TAGS = new Set([
  'a', 'b', 'big', 'blockquote', 'br', 'cite', 'code', 'dfn', 'div', 'em', 'font',
  'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'i', 'li', 'ol', 'p', 'pre', 's', 'small',
  'span', 'strike', 'strong', 'sub', 'sup', 'tt', 'u', 'ul',
]);

const LABEL_ALLOWED_ATTRS = {
  a: new Set(['href']),
  font: new Set(['color', 'face']),
};

function labelEscapeHtml(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function labelSanitizeNode(domNode) {
  if (domNode.nodeType === 3) return labelEscapeHtml(domNode.nodeValue);
  if (domNode.nodeType !== 1) return '';
  const tag = domNode.tagName.toLowerCase();
  const children = Array.from(domNode.childNodes).map(labelSanitizeNode).join('');
  if (!LABEL_ALLOWED_TAGS.has(tag)) return children;
  if (tag === 'br') return '<br>';
  let attrs = '';
  const allowed = LABEL_ALLOWED_ATTRS[tag];
  if (allowed) {
    for (const name of allowed) {
      const value = domNode.getAttribute(name);
      if (value == null) continue;
      if (name === 'href' && !/^(https?:|mailto:|tel:)/i.test(value.trim())) continue;
      attrs += ` ${name}="${labelEscapeHtml(value)}"`;
    }
  }
  return `<${tag}${attrs}>${children}</${tag}>`;
}

export function labelHtml(text) {
  const raw = String(text ?? '');
  if (typeof document === 'undefined') return labelEscapeHtml(raw);
  const template = document.createElement('template');
  template.innerHTML = raw;
  return Array.from(template.content.childNodes).map(labelSanitizeNode).join('');
}
