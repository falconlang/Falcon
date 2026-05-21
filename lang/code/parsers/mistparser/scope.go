package mistparser

import (
	"Falcon/code/ast"
)

type VarEntry struct {
	Signatures []ast.Signature
	Count      int
}

type Scope struct {
	Type      ScopeType
	Parent    *Scope
	Variables map[string]*VarEntry
}

func (s *Scope) DefineVariable(name string, signatures []ast.Signature) {
	s.Variables[name] = &VarEntry{Signatures: signatures, Count: 0}
}

func (s *Scope) ReferVariable(name string) ([]ast.Signature, bool) {
	variable, ok := s.Variables[name]
	if ok {
		variable.Count += 1
		return variable.Signatures, true
	}
	if s.Parent != nil {
		return s.Parent.ReferVariable(name)
	}
	return make([]ast.Signature, 0), false
}

func (s *Scope) ReferGlobalVariable(name string) ([]ast.Signature, bool) {
	variable, ok := s.Variables[name]
	if ok {
		variable.Count += 1
		return variable.Signatures, true
	}
	return make([]ast.Signature, 0), false
}

func (s *Scope) GetVariableReferCount(name string) int {
	variable, ok := s.Variables[name]
	if ok {
		return variable.Count
	}
	return -1
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
