package mistparser

import (
	"Falcon/code/ast"
)

type Scope struct {
	Type      ScopeType
	Parent    *Scope
	Variables map[string][]ast.Signature

	YieldIndex          int
	ChildYieldScopeType ScopeType
	YieldName           *string
	//Yield               *fundamentals.Yield
}

func (s *Scope) DefineVariable(name string, signature []ast.Signature) {
	s.Variables[name] = signature
}

func (s *Scope) ResolveVariable(name string) ([]ast.Signature, bool) {
	signature, ok := s.Variables[name]
	if ok {
		return signature, true
	}
	if s.Parent != nil {
		return s.Parent.ResolveVariable(name)
	}
	return make([]ast.Signature, 0), false
}

func (s *Scope) InLoop() bool {
	var currScope = s
	for {
		if currScope.Type == ScopeLoop {
			return true
		}
		currScope = currScope.Parent
		if currScope == nil {
			return false
		}
	}
}

func (s *Scope) GetLoopScope() *Scope {
	var currScope = s
	for {
		if currScope.Type == ScopeLoop {
			return currScope
		}
		currScope = currScope.Parent
		if currScope == nil {
			return nil
		}
	}
}

func (s *Scope) IsRoot() bool {
	return s.Parent == nil
}
