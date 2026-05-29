/**
 * CodeMirror 6 language support for Falcon.
 * Uses StreamLanguage so the existing tokeniser logic is reused directly.
 */
import { StreamLanguage, HighlightStyle, syntaxHighlighting } from '@codemirror/language';
import { tags as t } from '@lezer/highlight';

const KW = new Set(['global','local','func','if','else','while','for','in','step','break','yield','when','any','this']);
const BI = new Set(['println','print','sqrt','abs','neg','log','exp','round','ceil','floor','sin','cos','tan','asin','acos','atan','degrees','radians','decToHex','decToBin','hexToDec','binToDec','randInt','randFloat','setRandSeed','min','max','avgOf','maxOf','minOf','geoMeanOf','stdDevOf','stdErrOf','modeOf','mod','rem','quot','atan2','formatDecimal','dec','octal','openScreen','openScreenWithValue','closeScreenWithValue','getStartValue','closeScreen','closeApp','getPlainStartText','copyList','copyDict','makeColor','splitColor','set','get']);
const TW = new Set(['text','number','list','dict','base10','hexa','bin','emptyList','emptyText']);

// ── StreamLanguage definition ───────────────────────────────────────────────

export const falconLang = StreamLanguage.define({
  name: 'falcon',

  startState() {
    return { lastQ: false, lastDot: false, nextFn: false };
  },

  token(stream, state) {
    // whitespace
    if (stream.eatWhile(c => c === ' ' || c === '\t')) return null;

    const ch = stream.peek();
    if (ch == null) return null;

    // line comment
    if (ch === '/' && stream.string[stream.pos + 1] === '/') {
      stream.skipToEnd();
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'comment';
    }

    // string literal
    if (ch === '"') {
      stream.next();
      while (!stream.eol()) {
        const c = stream.next();
        if (c === '\\') stream.next();
        else if (c === '"') break;
      }
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'string';
    }

    // hex colour literal  #RRGGBB
    if (ch === '#' && /[0-9A-Fa-f]/.test(stream.string[stream.pos + 1] || '')) {
      stream.next();
      stream.eatWhile(c => /[0-9A-Fa-f]/.test(c));
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'string';
    }

    // @TypeTag
    if (ch === '@') {
      stream.next();
      stream.eatWhile(c => /[A-Za-z0-9_]/.test(c));
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'typeName';
    }

    // number
    if (/[0-9]/.test(ch)) {
      stream.eatWhile(c => /[0-9]/.test(c));
      if (stream.peek() === '.' && /[0-9]/.test(stream.string[stream.pos + 1] || '')) {
        stream.next();
        stream.eatWhile(c => /[0-9]/.test(c));
      }
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'number';
    }

    // three-char operators
    const s3 = stream.string.slice(stream.pos, stream.pos + 3);
    if (s3 === '===' || s3 === '!==') {
      stream.pos += 3;
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'operator';
    }

    // two-char operators
    const s2 = stream.string.slice(stream.pos, stream.pos + 2);
    if (['->','..','==','!=','<<','>>','<=','>=','&&','||'].includes(s2)) {
      stream.pos += 2;
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'operator';
    }

    // dot
    if (ch === '.') {
      stream.next();
      state.lastDot = true; state.lastQ = false; state.nextFn = false;
      return 'punctuation';
    }

    // single-char operators
    if ('+-*/%^!<>=?~|&'.includes(ch)) {
      stream.next();
      state.lastQ = ch === '?'; state.lastDot = false; state.nextFn = false;
      return 'operator';
    }

    // brackets / punctuation
    if ('(){}[],;:'.includes(ch)) {
      stream.next();
      state.lastQ = false; state.lastDot = false; state.nextFn = false;
      return 'punctuation';
    }

    // identifiers / keywords
    if (/[A-Za-z_]/.test(ch)) {
      const start = stream.pos;
      stream.eatWhile(c => /[A-Za-z0-9_]/.test(c));
      const w = stream.string.slice(start, stream.pos);

      // peek past whitespace for following '('
      let ai = stream.pos;
      while (stream.string[ai] === ' ' || stream.string[ai] === '\t') ai++;
      const nc = stream.string[ai];

      let cls;
      if (w === '_')                          cls = 'operator';
      else if (state.lastQ && TW.has(w))      cls = 'typeName';
      else if (state.lastDot)                 cls = 'function';
      else if (state.nextFn)                  cls = 'function';
      else if (KW.has(w))                     cls = 'keyword';
      else if (w === 'true' || w === 'false') cls = 'number';
      else if (BI.has(w))                     cls = 'builtin';
      else if (nc === '(')                    cls = 'function';
      else if (/^[A-Z]/.test(w))             cls = 'typeName';
      else                                    cls = null;

      state.lastDot = false;
      state.lastQ = false;
      state.nextFn = (cls === 'keyword' && w === 'func');
      return cls;
    }

    stream.next();
    state.lastQ = false; state.lastDot = false;
    return null;
  },

  blankLine(state) {
    state.lastQ = false; state.lastDot = false;
  },

  // Map string token names returned by token() to Lezer tags
  tokenTable: {
    keyword:     t.keyword,
    builtin:     t.standard(t.name),
    string:      t.string,
    number:      t.number,
    comment:     t.lineComment,
    function:    t.function(t.name),
    operator:    t.operator,
    punctuation: t.punctuation,
    typeName:    t.typeName,
  },
});

// ── HighlightStyle ──────────────────────────────────────────────────────────
// Uses CSS variables so light/dark themes are handled by app.css, not here.

const falconHighlight = HighlightStyle.define([
  { tag: t.keyword,          color: 'var(--tok-k)' },
  { tag: t.standard(t.name), color: 'var(--tok-b)', fontWeight: '500' },
  { tag: t.string,           color: 'var(--tok-s)' },
  { tag: t.number,           color: 'var(--tok-n)' },
  { tag: t.lineComment,      color: 'var(--tok-c)', fontStyle: 'italic' },
  { tag: t.function(t.name), color: 'var(--tok-f)' },
  { tag: t.operator,         color: 'var(--tok-o)' },
  { tag: t.punctuation,      color: 'var(--tok-p)' },
  { tag: t.typeName,         color: 'var(--tok-t)' },
]);

export const falconSyntaxHighlighting = syntaxHighlighting(falconHighlight);
