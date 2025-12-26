package lex

import (
	"Falcon/code/context"
	"Falcon/code/sugar"
	"strconv"
)

type Token struct {
	Column  int
	Row     int
	Context *context.CodeContext

	Type    Type
	Flags   []Flag
	Content *string
}

func (t *Token) String() string {
	return sugar.Format("(% %)", t.Type.String(), *t.Content)
}

func (t *Token) Debug() string {
	return sugar.Format("(%:% % %)", strconv.Itoa(t.Column), strconv.Itoa(t.Row), t.Type.String(), *t.Content)
}

func (t *Token) HasFlag(flag Flag) bool {
	for _, f := range t.Flags {
		if f == flag {
			return true
		}
	}
	return false
}

func (t *Token) Error(message string, args ...string) {
	if t.Context != nil {
		(*t.Context).ReportError(t.Column, t.Row, len(*t.Content), message, args...)
	} else {
		panic(sugar.Format(message, args...))
	}
}

func (t *Token) BuildError(decorate bool, message string, args ...string) string {
	if t.Context != nil {
		return (*t.Context).BuildError(decorate, t.Column, t.Row, len(*t.Content), message, args...)
	} else {
		return sugar.Format(message, args...)
	}
}

type StaticToken struct {
	Type  Type
	Flags []Flag
}

func staticOf(t Type, flags ...Flag) StaticToken {
	return StaticToken{t, flags}
}

func (s *StaticToken) Normal(
	column int,
	row int,
	ctx *context.CodeContext,
	content string,
) *Token {
	return &Token{
		Column:  column,
		Row:     row,
		Context: ctx,

		Type:    s.Type,
		Flags:   s.Flags,
		Content: &content,
	}
}

// TODO: (future) it'll point to something meaningful
func MakeFakeToken(t Type) *Token {
	return &Token{
		Column:  -1,
		Row:     -1,
		Context: nil,
		Type:    t,
		Flags:   make([]Flag, 0),
		Content: nil,
	}
}
