// Package xmlutil provides a minimal XML tokenizer and tree builder that does
// not import fmt, reflect, or encoding/xml, keeping WASM binary size small.
package xmlutil

import (
	"errors"
	"strconv"
	"strings"
)

// ── public types ──────────────────────────────────────────────────────────────

// Attr is an XML attribute.
type Attr struct {
	Name  string
	Value string
}

// Element is a parsed XML element node.
type Element struct {
	Name     string
	Attrs    []Attr
	Children []*Element
	Text     string // concatenated character data
}

// AttrVal returns the value of the named attribute, or "" if absent.
func (e *Element) AttrVal(name string) string {
	for _, a := range e.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

// AttrBool returns true if the named attribute equals "true".
func (e *Element) AttrBool(name string) bool {
	return e.AttrVal(name) == "true"
}

// AttrInt returns the named attribute parsed as an integer.
func (e *Element) AttrInt(name string) int {
	v, _ := strconv.Atoi(e.AttrVal(name))
	return v
}

// Children with the given tag name.
func (e *Element) ChildrenNamed(name string) []*Element {
	var out []*Element
	for _, c := range e.Children {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// FirstChild returns the first child with the given name, or nil.
func (e *Element) FirstChild(name string) *Element {
	for _, c := range e.Children {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// EscapeText returns s with XML special characters replaced by entities.
func EscapeText(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		case '"':
			sb.WriteString("&quot;")
		case '\'':
			sb.WriteString("&apos;")
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// ParseDocument parses an XML string and returns the root element.
func ParseDocument(src string) (*Element, error) {
	t := &tokenizer{src: src}
	root, err := t.parseDocument()
	if err != nil {
		return nil, err
	}
	return root, nil
}

// ── tokenizer ─────────────────────────────────────────────────────────────────

type tokKind int

const (
	tokStart tokKind = iota // start element (may be self-closing)
	tokEnd                  // end element
	tokChar                 // character data
	tokEOF
)

type rawToken struct {
	kind      tokKind
	name      string
	attrs     []Attr
	data      string
	selfClose bool
}

type tokenizer struct {
	src string
	pos int
	// buffered token for self-closing elements
	pending *rawToken
}

func (t *tokenizer) next() (rawToken, error) {
	if t.pending != nil {
		tok := *t.pending
		t.pending = nil
		return tok, nil
	}
	for {
		t.skipWS()
		if t.pos >= len(t.src) {
			return rawToken{kind: tokEOF}, nil
		}
		if t.src[t.pos] != '<' {
			// character data
			data, err := t.readCharData()
			if err != nil {
				return rawToken{}, err
			}
			if strings.TrimSpace(data) == "" {
				continue // skip whitespace-only text nodes
			}
			return rawToken{kind: tokChar, data: data}, nil
		}
		// We're at '<'
		if t.pos+1 >= len(t.src) {
			return rawToken{}, errors.New("unexpected end after '<'")
		}
		switch {
		case t.src[t.pos+1] == '/':
			// end element
			tok, err := t.readEndElement()
			if err != nil {
				return rawToken{}, err
			}
			return tok, nil
		case t.src[t.pos+1] == '!':
			// comment or CDATA
			if strings.HasPrefix(t.src[t.pos:], "<!--") {
				if err := t.skipComment(); err != nil {
					return rawToken{}, err
				}
				continue
			}
			if strings.HasPrefix(t.src[t.pos:], "<![CDATA[") {
				data, err := t.readCDATA()
				if err != nil {
					return rawToken{}, err
				}
				return rawToken{kind: tokChar, data: data}, nil
			}
			return rawToken{}, errors.New("unknown '<!' construct at position " + strconv.Itoa(t.pos))
		case t.src[t.pos+1] == '?':
			// processing instruction / XML declaration
			if err := t.skipPI(); err != nil {
				return rawToken{}, err
			}
			continue
		default:
			tok, err := t.readStartElement()
			if err != nil {
				return rawToken{}, err
			}
			if tok.selfClose {
				// Buffer an implicit end element
				t.pending = &rawToken{kind: tokEnd, name: tok.name}
				tok.selfClose = false
			}
			return tok, nil
		}
	}
}

func (t *tokenizer) skipWS() {
	for t.pos < len(t.src) {
		switch t.src[t.pos] {
		case ' ', '\t', '\n', '\r':
			t.pos++
		default:
			return
		}
	}
}

func (t *tokenizer) readCharData() (string, error) {
	start := t.pos
	for t.pos < len(t.src) && t.src[t.pos] != '<' {
		t.pos++
	}
	return expandEntities(t.src[start:t.pos]), nil
}

func (t *tokenizer) readEndElement() (rawToken, error) {
	t.pos += 2 // consume '</'
	name := t.readName()
	t.skipWS()
	if t.pos >= len(t.src) || t.src[t.pos] != '>' {
		return rawToken{}, errors.New("expected '>' in end element")
	}
	t.pos++ // consume '>'
	return rawToken{kind: tokEnd, name: name}, nil
}

func (t *tokenizer) readStartElement() (rawToken, error) {
	t.pos++ // consume '<'
	name := t.readName()
	var attrs []Attr
	for {
		t.skipWS()
		if t.pos >= len(t.src) {
			return rawToken{}, errors.New("unexpected end in start element")
		}
		if t.src[t.pos] == '>' {
			t.pos++
			return rawToken{kind: tokStart, name: name, attrs: attrs}, nil
		}
		if t.pos+1 < len(t.src) && t.src[t.pos] == '/' && t.src[t.pos+1] == '>' {
			t.pos += 2
			return rawToken{kind: tokStart, name: name, attrs: attrs, selfClose: true}, nil
		}
		// read attribute
		attr, err := t.readAttr()
		if err != nil {
			return rawToken{}, err
		}
		attrs = append(attrs, attr)
	}
}

func (t *tokenizer) readAttr() (Attr, error) {
	name := t.readName()
	t.skipWS()
	if t.pos >= len(t.src) || t.src[t.pos] != '=' {
		// boolean attribute (no value) – treat value as name
		return Attr{Name: name, Value: name}, nil
	}
	t.pos++ // consume '='
	t.skipWS()
	if t.pos >= len(t.src) {
		return Attr{}, errors.New("unexpected end after '='")
	}
	quote := t.src[t.pos]
	if quote != '"' && quote != '\'' {
		return Attr{}, errors.New("expected '\"' or \"'\" in attribute at position " + strconv.Itoa(t.pos))
	}
	t.pos++ // consume opening quote
	start := t.pos
	for t.pos < len(t.src) && t.src[t.pos] != quote {
		t.pos++
	}
	if t.pos >= len(t.src) {
		return Attr{}, errors.New("unterminated attribute value")
	}
	value := expandEntities(t.src[start:t.pos])
	t.pos++ // consume closing quote
	return Attr{Name: name, Value: value}, nil
}

func (t *tokenizer) readName() string {
	start := t.pos
	for t.pos < len(t.src) {
		c := t.src[t.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' || c == '/' || c == '=' || c == '"' || c == '\'' {
			break
		}
		t.pos++
	}
	name := t.src[start:t.pos]
	// Strip namespace prefix (we don't care about namespaces)
	if idx := strings.IndexByte(name, ':'); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

func (t *tokenizer) skipComment() error {
	t.pos += 4 // consume '<!--'
	for t.pos+2 < len(t.src) {
		if t.src[t.pos] == '-' && t.src[t.pos+1] == '-' && t.src[t.pos+2] == '>' {
			t.pos += 3
			return nil
		}
		t.pos++
	}
	return errors.New("unterminated XML comment")
}

func (t *tokenizer) readCDATA() (string, error) {
	t.pos += 9 // consume '<![CDATA['
	start := t.pos
	for t.pos+2 < len(t.src) {
		if t.src[t.pos] == ']' && t.src[t.pos+1] == ']' && t.src[t.pos+2] == '>' {
			data := t.src[start:t.pos]
			t.pos += 3
			return data, nil
		}
		t.pos++
	}
	return "", errors.New("unterminated CDATA section")
}

func (t *tokenizer) skipPI() error {
	t.pos += 2 // consume '<?'
	for t.pos+1 < len(t.src) {
		if t.src[t.pos] == '?' && t.src[t.pos+1] == '>' {
			t.pos += 2
			return nil
		}
		t.pos++
	}
	return errors.New("unterminated processing instruction")
}

// ── tree builder ──────────────────────────────────────────────────────────────

func (t *tokenizer) parseDocument() (*Element, error) {
	// skip any leading whitespace / PI / comments until we find the root start element
	for {
		tok, err := t.next()
		if err != nil {
			return nil, err
		}
		if tok.kind == tokEOF {
			return nil, errors.New("empty XML document")
		}
		if tok.kind == tokStart {
			elem := &Element{Name: tok.name, Attrs: tok.attrs}
			if err := t.parseChildren(elem); err != nil {
				return nil, err
			}
			return elem, nil
		}
		// skip char data or other tokens before root element
	}
}

func (t *tokenizer) parseChildren(parent *Element) error {
	for {
		tok, err := t.next()
		if err != nil {
			return err
		}
		switch tok.kind {
		case tokEOF:
			return errors.New("unexpected end of XML: missing </" + parent.Name + ">")
		case tokEnd:
			return nil // done with this element
		case tokChar:
			parent.Text += tok.data
		case tokStart:
			child := &Element{Name: tok.name, Attrs: tok.attrs}
			if err := t.parseChildren(child); err != nil {
				return err
			}
			parent.Children = append(parent.Children, child)
		}
	}
}

// ── entity expansion ──────────────────────────────────────────────────────────

func expandEntities(s string) string {
	if !strings.ContainsRune(s, '&') {
		return s
	}
	var sb strings.Builder
	i := 0
	for i < len(s) {
		if s[i] != '&' {
			sb.WriteByte(s[i])
			i++
			continue
		}
		// find ';'
		j := i + 1
		for j < len(s) && s[j] != ';' {
			j++
		}
		if j >= len(s) {
			sb.WriteByte('&')
			i++
			continue
		}
		ref := s[i+1 : j]
		i = j + 1
		switch ref {
		case "amp":
			sb.WriteByte('&')
		case "lt":
			sb.WriteByte('<')
		case "gt":
			sb.WriteByte('>')
		case "quot":
			sb.WriteByte('"')
		case "apos":
			sb.WriteByte('\'')
		default:
			if len(ref) > 1 && ref[0] == '#' {
				// numeric character reference
				var code int64
				var err error
				if len(ref) > 2 && ref[1] == 'x' {
					code, err = strconv.ParseInt(ref[2:], 16, 32)
				} else {
					code, err = strconv.ParseInt(ref[1:], 10, 32)
				}
				if err == nil {
					sb.WriteRune(rune(code))
				} else {
					sb.WriteByte('&')
					sb.WriteString(ref)
					sb.WriteByte(';')
				}
			} else {
				// unknown entity — emit as-is
				sb.WriteByte('&')
				sb.WriteString(ref)
				sb.WriteByte(';')
			}
		}
	}
	return sb.String()
}
