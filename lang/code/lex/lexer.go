package lex

import (
	"Falcon/code/context"
	"Falcon/code/sugar"
	"strconv"
	"strings"
)

type Lexer struct {
	ctx        *context.CodeContext
	source     string
	sourceLen  int
	currIndex  int
	currColumn int
	currRow    int
	Tokens     []*Token
}

func NewLexer(ctx *context.CodeContext) *Lexer {
	return &Lexer{
		ctx:        ctx,
		source:     *ctx.SourceCode,
		sourceLen:  len(*ctx.SourceCode),
		currIndex:  0,
		currColumn: 1, // current line
		currRow:    0, // nth character of current line
		Tokens:     []*Token{},
	}
}

func (l *Lexer) Lex() []*Token {
	for l.notEOF() {
		l.parse()
	}
	return l.Tokens
}

func (l *Lexer) parse() {
	c := l.next()

	if c == '/' && l.consume('/') {
		// comment, skip the current line
		for l.notEOF() {
			n := l.next()
			if n == '\n' {
				l.currColumn++
				l.currRow = 0
				break
			}
		}
		return
	}
	if c == '\n' {
		l.currColumn++
		l.currRow = 0
		return
	} else if c == ' ' || c == '\t' {
		return
	}
	switch c {
	case '+':
		l.createOp("+")
	case '-':
		if l.consume('>') {
			l.createOp("->")
		} else {
			l.createOp("-")
		}
	case '*':
		l.createOp("*")
	case '/':
		l.createOp("/")
	case '%':
		l.createOp("%")
	case '^':
		l.createOp("^")
	case '|':
		if l.consume('|') {
			l.createOp("||")
		} else {
			l.createOp("|")
		}
	case '&':
		if l.consume('&') {
			l.createOp("&&")
		} else {
			l.createOp("&")
		}
	case '~':
		l.createOp("~")
	case '<':
		if l.consume('=') {
			l.createOp("<=")
		} else if l.consume('<') {
			l.createOp("<<")
		} else {
			l.createOp("<")
		}
	case '>':
		if l.consume('=') {
			l.createOp(">=")
		} else if l.consume('>') {
			l.createOp(">>")
		} else {
			l.createOp(">")
		}
	case '(':
		l.createOp("(")
	case ')':
		l.createOp(")")
	case '[':
		l.createOp("[")
	case ']':
		l.createOp("]")
	case '{':
		l.createOp("{")
	case '}':
		l.createOp("}")
	case '=':
		if l.consume('=') {
			if l.consume('=') {
				l.createOp("===")
			} else {
				l.createOp("==")
			}
		} else {
			l.createOp("=")
		}
	case '.':
		if l.consume('.') {
			l.createOp("..")
		} else {
			l.createOp(".")
		}
	case ',':
		l.createOp(",")
	case '?':
		l.createOp("?")
	case '!':
		if l.consume('=') {
			if l.consume('=') {
				l.createOp("!==")
			} else {
				l.createOp("!=")
			}
		} else {
			l.createOp("!")
		}
	case ':':
		if l.consume(':') {
			l.createOp("::")
		} else {
			l.createOp(":")
		}
	case '_':
		l.createOp("_")
	case '@':
		l.createOp("@")
	case '"':
		l.text()
	case '#':
		l.colorCode()
	default:
		l.back()
		if l.isAlpha() {
			l.alpha()
		} else if l.isDigit() {
			l.numeric()
		} else {
			l.error("Unexpected character '%'", string(c))
		}
	}
}

func (l *Lexer) createOp(op string) {
	sToken, ok := Symbols[op]
	if !ok {
		l.error("Bad createOp('%')", op)
	} else {
		l.appendToken(sToken.Normal(l.currColumn, l.currRow, l.ctx, op))
	}
}

func (l *Lexer) colorCode() {
	startIndex := l.currIndex
	// Read up to 6 hex characters
	for i := 0; i < 6 && l.notEOF(); i++ {
		c := l.peek()
		if (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f') {
			l.skip()
		} else {
			l.error("Invalid color code character '%' in color literal", string(c))
		}
	}

	length := l.currIndex - startIndex
	if length != 6 {
		l.error("Color code must be 6 hexadecimal characters, got %", strconv.Itoa(length))
	}
	content := l.source[startIndex-1 : l.currIndex] // include '#'
	l.appendToken(&Token{
		Context: l.ctx,
		Row:     l.currRow,
		Column:  l.currColumn,

		Type:    ColorCode,
		Content: &content,
		Flags:   []Flag{Value, ConstantValue},
	})
}

func (l *Lexer) text() {
	var writer strings.Builder
	for {
		c := l.next()
		if c == '"' {
			break
		}
		if c == '\\' {
			// Only handle escaping of (")
			e := l.peek()
			if e == '"' || e == '\\' {
				c = e
				l.skip()
			}
		}
		writer.WriteByte(c)
	}
	println(len(writer.String()) == 0)
	content := writer.String()
	l.appendToken(&Token{
		Context: l.ctx,
		Row:     l.currRow,
		Column:  l.currColumn,
		Type:    Text,
		Content: &content,
		Flags:   []Flag{Value, ConstantValue},
	})
}

func (l *Lexer) alpha() {
	startIndex := l.currIndex
	l.skip()
	for l.notEOF() && l.isAlphaNumeric() {
		l.skip()
	}
	content := l.source[startIndex:l.currIndex]
	sToken, ok := Keywords[content]
	if ok {
		l.appendToken(sToken.Normal(l.currColumn, l.currRow, l.ctx, content))
	} else {
		l.appendToken(&Token{
			Context: l.ctx,
			Row:     l.currRow,
			Column:  l.currColumn,

			Type:    Name,
			Content: &content,
			Flags:   []Flag{Value},
		})
	}
}

func (l *Lexer) numeric() {
	var numb strings.Builder
	l.writeNumeric(&numb)
	if l.notEOF() && l.peek() == '.' {
		l.skip()
		numb.WriteByte('.')
		l.writeNumeric(&numb)
	}
	content := numb.String()
	l.appendToken(&Token{
		Context: l.ctx,
		Row:     l.currRow,
		Column:  l.currColumn,

		Type:    Number,
		Content: &content,
		Flags:   []Flag{Value, ConstantValue},
	})
}

func (l *Lexer) appendToken(token *Token) {
	println(token.Debug())
	l.Tokens = append(l.Tokens, token)
}

func (l *Lexer) writeNumeric(builder *strings.Builder) {
	startIndex := l.currIndex
	for l.notEOF() && l.isDigit() {
		l.skip()
	}
	builder.WriteString(l.source[startIndex:l.currIndex])
}

func (l *Lexer) isAlphaNumeric() bool {
	return l.isAlpha() || l.isDigit()
}

func (l *Lexer) isAlpha() bool {
	c := l.peek()
	return c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c == '_'
}

func (l *Lexer) isDigit() bool {
	c := l.peek()
	return c >= '0' && c <= '9'
}

func (l *Lexer) eat(expect uint8) {
	got := l.next()
	if got != expect {
		l.error("Expected '%', but got '%'", string(expect), string(got))
	}
}

func (l *Lexer) error(message string, args ...string) {
	panic("[line " + strconv.Itoa(l.currColumn) + "] " + sugar.Format(message, args...))
}

func (l *Lexer) consume(expect uint8) bool {
	if l.peek() == expect {
		l.currIndex++
		l.currRow++
		return true
	}
	return false
}

func (l *Lexer) back() {
	l.currIndex--
	l.currRow--
}

func (l *Lexer) skip() {
	l.currIndex++
	l.currRow++
}

func (l *Lexer) peek() uint8 {
	return l.source[l.currIndex]
}

func (l *Lexer) next() uint8 {
	c := l.source[l.currIndex]
	l.currIndex++
	l.currRow++
	return c
}

func (l *Lexer) isEOF() bool {
	return l.currIndex >= l.sourceLen
}

func (l *Lexer) notEOF() bool {
	return l.currIndex < l.sourceLen
}
