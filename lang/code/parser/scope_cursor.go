package parser

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
	"strconv"
)

type ScopeType int

const (
	ScopeRoot ScopeType = iota
	ScopeRetProc
	ScopeProc
	ScopeGenericEvent
	ScopeEvent
	ScopeLoop
	ScopeIfBody
	ScopeSmartBody
)

type ScopeCursor struct {
	allScopes []*Scope
	headScope *Scope
	currScope *Scope
}

func MakeScopeCursor() *ScopeCursor {
	headScope := &Scope{Type: ScopeRoot, Parent: nil, Variables: map[string][]ast.Signature{}}
	return &ScopeCursor{allScopes: []*Scope{headScope}, headScope: headScope, currScope: headScope}
}

func (s *ScopeCursor) Enter(where *lex.Token, t ScopeType) {
	s.checkScope(where, t)
	newScope := &Scope{Type: t, Parent: s.currScope, Variables: map[string][]ast.Signature{}}
	s.allScopes = append(s.allScopes, newScope)
	s.currScope = newScope
}

func (s *ScopeCursor) checkScope(where *lex.Token, t ScopeType) {
	depth := len(s.allScopes)
	if t == ScopeRetProc || t == ScopeProc {
		if depth != 1 {
			where.Error("Functions can only be defined at the root.")
		}
	} else if t == ScopeGenericEvent || t == ScopeEvent {
		if depth != 1 {
			where.Error("Events can only be defined at the root.")
		}
	}
}

func (s *ScopeCursor) Exit(t ScopeType) {
	if len(s.allScopes) == 1 {
		panic("Cannot exit the global scope!")
	}
	topIndex := len(s.allScopes) - 1
	current := s.allScopes[topIndex]
	s.allScopes = s.allScopes[:topIndex]
	if current.Type != t {
		panic("Bad scope exit! Expected scope type " + strconv.Itoa(int(current.Type)) + " but got type " + strconv.Itoa(int(t)))
	}
	s.currScope = s.allScopes[len(s.allScopes)-1]
}

func (s *ScopeCursor) DefineVariable(name string, signature []ast.Signature) {
	s.currScope.DefineVariable(name, signature)
}

func (s *ScopeCursor) ResolveVariable(name string) ([]ast.Signature, bool) {
	return s.currScope.ResolveVariable(name)
}

func (s *ScopeCursor) In(t ScopeType) bool {
	for _, scope := range s.allScopes {
		if scope.Type == t {
			return true
		}
	}
	return false
}

func (s *ScopeCursor) AtRoot() bool {
	return len(s.allScopes) == 1
}
