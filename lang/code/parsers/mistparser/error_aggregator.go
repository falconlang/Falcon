package mistparser

import (
	"Falcon/code/ast"
	l "Falcon/code/lex"
)

type ErrorAggregator struct {
	Errors map[*l.Token]ParseError
}

type ParseError struct {
	Owner        ast.Expr
	ErrorMessage string
	Deferred     bool
}

func (e *ErrorAggregator) EnqueueSymbol(where *l.Token, owner ast.Expr, message string) {
	e.Errors[where] = ParseError{Owner: owner, ErrorMessage: message, Deferred: true}
}

func (e *ErrorAggregator) EnqueueError(where *l.Token, owner ast.Expr, message string) {
	e.Errors[where] = ParseError{Owner: owner, ErrorMessage: message}
}

func (e *ErrorAggregator) MarkResolved(where *l.Token) {
	delete(e.Errors, where)
}
