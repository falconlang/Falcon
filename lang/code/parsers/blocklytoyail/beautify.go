package blocklytoyail

import "strings"

const beautifyMaxWidth = 80

type bToken struct {
	kind byte
	text string
}

type bNode struct {
	text      string
	children  []bNode
	isList    bool
	quoted    bool
	isComment bool
}

func Beautify(yail string) string {
	nodes := bParse(bTokenize(yail))
	var sb strings.Builder
	for i, n := range nodes {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(bRender(n, 0))
	}
	return sb.String()
}

func bTokenize(s string) []bToken {
	var toks []bToken
	i, n := 0, len(s)
	for i < n {
		c := s[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '(':
			toks = append(toks, bToken{'(', "("})
			i++
		case c == ')':
			toks = append(toks, bToken{')', ")"})
			i++
		case c == '\'':
			toks = append(toks, bToken{'q', "'"})
			i++
		case c == ';':
			start := i
			for i < n && s[i] != '\n' {
				i++
			}
			toks = append(toks, bToken{'c', s[start:i]})
		case c == '"':
			start := i
			i++
			for i < n {
				if s[i] == '\\' && i+1 < n {
					i += 2
					continue
				}
				if s[i] == '"' {
					i++
					break
				}
				i++
			}
			toks = append(toks, bToken{'s', s[start:i]})
		default:
			start := i
			for i < n {
				ch := s[i]
				if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' || ch == '(' || ch == ')' {
					break
				}
				i++
			}
			toks = append(toks, bToken{'a', s[start:i]})
		}
	}
	return toks
}

func bParse(toks []bToken) []bNode {
	pos := 0
	var parseNode func() bNode
	parseNode = func() bNode {
		if pos >= len(toks) {
			return bNode{}
		}
		t := toks[pos]
		switch t.kind {
		case 'q':
			pos++
			n := parseNode()
			n.quoted = true
			return n
		case '(':
			pos++
			var children []bNode
			for pos < len(toks) && toks[pos].kind != ')' {
				children = append(children, parseNode())
			}
			if pos < len(toks) {
				pos++
			}
			return bNode{children: children, isList: true}
		case 'a', 's':
			pos++
			return bNode{text: t.text}
		case 'c':
			pos++
			return bNode{text: t.text, isComment: true}
		}
		pos++
		return bNode{}
	}
	var roots []bNode
	for pos < len(toks) {
		roots = append(roots, parseNode())
	}
	return roots
}

func bInline(n bNode) string {
	if n.isComment {
		return n.text
	}
	var sb strings.Builder
	if n.quoted {
		sb.WriteByte('\'')
	}
	if !n.isList {
		sb.WriteString(n.text)
		return sb.String()
	}
	sb.WriteByte('(')
	for i, c := range n.children {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(bInline(c))
	}
	sb.WriteByte(')')
	return sb.String()
}

func bHasComment(n bNode) bool {
	if n.isComment {
		return true
	}
	for _, c := range n.children {
		if bHasComment(c) {
			return true
		}
	}
	return false
}

func bRender(n bNode, indent int) string {
	if n.isComment {
		return n.text
	}
	inline := bInline(n)
	if !n.isList || len(n.children) == 0 || (indent+len(inline) <= beautifyMaxWidth && !bHasComment(n)) {
		return inline
	}

	var sb strings.Builder
	if n.quoted {
		sb.WriteByte('\'')
	}
	sb.WriteByte('(')

	headLine := bInline(n.children[0])
	childIndent := indent + 2
	pad := strings.Repeat(" ", childIndent)

	i := 1
	sawList := false
	for i < len(n.children) {
		child := n.children[i]
		if child.isComment {
			break
		}
		argStr := bInline(child)
		if child.isList && sawList {
			break
		}
		// +1 for opening paren, +1 for space before this arg
		if indent+1+len(headLine)+1+len(argStr) > beautifyMaxWidth {
			break
		}
		headLine += " " + argStr
		if child.isList {
			sawList = true
		}
		i++
	}
	sb.WriteString(headLine)

	for ; i < len(n.children); i++ {
		sb.WriteByte('\n')
		sb.WriteString(pad)
		sb.WriteString(bRender(n.children[i], childIndent))
	}
	sb.WriteByte(')')
	return sb.String()
}
