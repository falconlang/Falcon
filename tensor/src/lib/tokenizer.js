export function escHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}

export function lineNums(code) {
  const lines = (code || '').replace(/\n$/, '').split('\n');
  return lines.map((_, i) => `<div>${i + 1}</div>`).join('');
}

export function tokensToHtml(tokens) {
  return tokens.map(tok => {
    if (tok === '\n') return '\n';
    const raw = typeof tok === 'string' ? tok : tok.v;
    const v = escHtml(raw);
    const cls = typeof tok === 'object' ? (tok.t || '') : '';
    return cls ? `<span class="${cls}">${v}</span>` : v;
  }).join('');
}

export function falconTokenize(src) {
  const KW = new Set(['global','local','func','if','else','while','for','in','step','break','yield','when','any','this']);
  const BI = new Set(['println','print','sqrt','abs','neg','log','exp','round','ceil','floor','sin','cos','tan','asin','acos','atan','degrees','radians','decToHex','decToBin','hexToDec','binToDec','randInt','randFloat','setRandSeed','min','max','avgOf','maxOf','minOf','geoMeanOf','stdDevOf','stdErrOf','modeOf','mod','rem','quot','atan2','formatDecimal','dec','octal','openScreen','openScreenWithValue','closeScreenWithValue','getStartValue','closeScreen','closeApp','getPlainStartText','copyList','copyDict','makeColor','splitColor','set','get']);
  const TW = new Set(['text','number','list','dict','base10','hexa','bin','emptyList','emptyText']);
  const toks = [];
  let i = 0, lastQ = false, lastDot = false, nextFn = false;
  const push = (t, v) => toks.push(t ? {t,v} : {t:'',v});

  while (i < src.length) {
    const ch = src[i];
    if (ch === '\n') { toks.push('\n'); i++; lastQ = false; lastDot = false; continue; }
    if (ch === ' ' || ch === '\t') {
      let j = i; while (j < src.length && (src[j]===' '||src[j]==='\t')) j++;
      push('', src.slice(i,j)); i=j; continue;
    }
    if (ch==='/' && src[i+1]==='/') {
      let j=i; while (j<src.length && src[j]!=='\n') j++;
      push('c', src.slice(i,j)); i=j; lastQ=false; lastDot=false; nextFn=false; continue;
    }
    if (ch==='"') {
      let j=i+1; while (j<src.length && src[j]!=='"') { if(src[j]==='\\') j++; j++; }
      if (j<src.length) j++;
      push('s', src.slice(i,j)); i=j; lastQ=false; lastDot=false; nextFn=false; continue;
    }
    if (ch==='#' && i+1<src.length && /[0-9A-Fa-f]/.test(src[i+1])) {
      let j=i+1; while (j<src.length && /[0-9A-Fa-f]/.test(src[j])) j++;
      push('s', src.slice(i,j)); i=j; lastQ=false; lastDot=false; nextFn=false; continue;
    }
    if (ch==='@') {
      let j=i+1; while (j<src.length && /[A-Za-z0-9_]/.test(src[j])) j++;
      push('t', src.slice(i,j)); i=j; lastQ=false; lastDot=false; nextFn=false; continue;
    }
    if (/[0-9]/.test(ch)) {
      let j=i; while (j<src.length && /[0-9]/.test(src[j])) j++;
      if (src[j]==='.' && j+1<src.length && /[0-9]/.test(src[j+1])) { j++; while (j<src.length && /[0-9]/.test(src[j])) j++; }
      push('n', src.slice(i,j)); i=j; lastQ=false; lastDot=false; nextFn=false; continue;
    }
    const s3=src.slice(i,i+3);
    if (s3==='===' || s3==='!==') { push('o',s3); i+=3; lastQ=false; lastDot=false; nextFn=false; continue; }
    const s2=src.slice(i,i+2);
    if (['->','..','==','!=','<<','>>','<=','>=','&&','||'].includes(s2)) { push('o',s2); i+=2; lastQ=false; lastDot=false; nextFn=false; continue; }
    if (ch==='.') { push('p','.'); i++; lastDot=true; lastQ=false; nextFn=false; continue; }
    if ('+-*/%^!<>=?~|&'.includes(ch)) {
      push('o',ch); i++;
      lastQ=(ch==='?'); lastDot=false; nextFn=false; continue;
    }
    if ('(){}[],;:'.includes(ch)) { push('p',ch); i++; lastQ=false; lastDot=false; nextFn=false; continue; }
    if (/[A-Za-z_]/.test(ch)) {
      let j=i; while (j<src.length && /[A-Za-z0-9_]/.test(src[j])) j++;
      const w=src.slice(i,j);
      let ai=j; while (ai<src.length && (src[ai]===' '||src[ai]==='\t')) ai++;
      const nc=src[ai];
      let cls;
      if (w==='_')                               cls='o';
      else if (lastQ && TW.has(w))              cls='t';
      else if (lastDot)                          cls='f';
      else if (nextFn)                           cls='f';
      else if (KW.has(w))                        cls='k';
      else if (w==='true'||w==='false')          cls='n';
      else if (BI.has(w))                        cls='b';
      else if (nc==='(')                         cls='f';
      else if (/^[A-Z]/.test(w))                cls='t';
      else                                       cls='';
      push(cls,w); i=j;
      lastDot=false; lastQ=false; nextFn=(cls==='k'&&w==='func');
      continue;
    }
    push('',ch); i++; lastQ=false; lastDot=false;
  }
  return toks;
}

export function schemaTokenize(src) {
  const toks = [];
  let i = 0, lastDot = false;
  const push = (t, v) => toks.push(t ? {t,v} : {t:'',v});

  while (i < src.length) {
    const ch = src[i];
    if (ch==='\n') { toks.push('\n'); i++; lastDot=false; continue; }
    if (ch===' '||ch==='\t') {
      let j=i; while (j<src.length && (src[j]===' '||src[j]==='\t')) j++;
      push('',src.slice(i,j)); i=j; continue;
    }
    if (ch==='/' && src[i+1]==='/') {
      let j=i; while (j<src.length && src[j]!=='\n') j++;
      push('c',src.slice(i,j)); i=j; lastDot=false; continue;
    }
    if (ch==='"') {
      let j=i+1; while (j<src.length && src[j]!=='"') { if(src[j]==='\\') j++; j++; }
      if (j<src.length) j++;
      push('s',src.slice(i,j)); i=j; lastDot=false; continue;
    }
    if (/[0-9]/.test(ch)) {
      let j=i; while (j<src.length && /[0-9.]/.test(src[j])) j++;
      push('n',src.slice(i,j)); i=j; lastDot=false; continue;
    }
    if (ch==='.') { push('p','.'); i++; lastDot=true; continue; }
    if ('{}:,'.includes(ch)) { push('p',ch); i++; lastDot=false; continue; }
    if (/[A-Za-z_]/.test(ch)) {
      let j=i; while (j<src.length && /[A-Za-z0-9_]/.test(src[j])) j++;
      const w=src.slice(i,j);
      let ai=j; while (ai<src.length && (src[ai]===' '||src[ai]==='\t')) ai++;
      let cls;
      if (lastDot)                       cls='f';
      else if (src[ai]===':')            cls='b';
      else if (w==='true'||w==='false')  cls='n';
      else if (/^[A-Z]/.test(w))        cls='t';
      else                               cls='';
      push(cls,w); i=j; lastDot=false; continue;
    }
    push('',ch); i++; lastDot=false;
  }
  return toks;
}

export function getCaretOffset(el) {
  return getSelectionOffsets(el)?.start ?? 0;
}

export function getSelectionOffsets(el) {
  const sel = window.getSelection();
  if (!sel.rangeCount) return null;
  const range = sel.getRangeAt(0);
  if (!el.contains(range.startContainer) || !el.contains(range.endContainer)) return null;

  const preStart = range.cloneRange();
  preStart.selectNodeContents(el);
  preStart.setEnd(range.startContainer, range.startOffset);

  const preEnd = range.cloneRange();
  preEnd.selectNodeContents(el);
  preEnd.setEnd(range.endContainer, range.endOffset);

  return {
    start: preStart.toString().length,
    end: preEnd.toString().length,
  };
}

export function setCaretOffset(el, target) {
  setSelectionOffsets(el, target, target);
}

export function setSelectionOffsets(el, start, end = start) {
  const sel = window.getSelection();
  const range = document.createRange();
  if (document.activeElement !== el) el.focus({ preventScroll: true });

  function findPosition(target) {
    let rem = Math.max(0, target);
    let lastText = null;

    function walk(node) {
      if (node.nodeType === 3) {
        lastText = node;
        if (rem <= node.length) return { node, offset: rem };
        rem -= node.length;
        return null;
      }
      for (const child of node.childNodes) {
        const found = walk(child);
        if (found) return found;
      }
      return null;
    }

    const found = walk(el);
    if (found) return found;
    if (lastText) return { node: lastText, offset: lastText.length };
    return { node: el, offset: 0 };
  }

  const startPos = findPosition(start);
  const endPos = findPosition(Math.max(start, end));

  range.setStart(startPos.node, startPos.offset);
  range.setEnd(endPos.node, endPos.offset);
  if (start === end) range.collapse(true);
  sel.removeAllRanges();
  sel.addRange(range);
}

export function highlight(el, tokenizer) {
  const off = getCaretOffset(el);
  el.innerHTML = tokensToHtml(tokenizer(el.textContent || ''));
  setCaretOffset(el, off);
}

export function insertTextAtCursor(text) {
  const sel = window.getSelection();
  if (!sel.rangeCount) return;
  const range = sel.getRangeAt(0);
  range.deleteContents();
  const node = document.createTextNode(text);
  range.insertNode(node);
  range.setStartAfter(node);
  range.collapse(true);
  sel.removeAllRanges();
  sel.addRange(range);
}
