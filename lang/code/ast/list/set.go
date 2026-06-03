package list

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/lex"
	"Falcon/code/sugar"
)

type Set struct {
	Where *lex.Token
	List  ast.Expr
	Index ast.Expr
	Value ast.Expr
}

func (s *Set) String() string {
	pFormat := "%[%] = %"
	if !s.List.Continuous() {
		pFormat = "(%)[%] = %"
	}
	index := s.Index.String()
	if _, ok := ast.UnwrapAnnotated(s.Index).(*fundamentals.List); ok {
		index = "(" + index + ")"
	}
	return sugar.Format(pFormat, s.List.String(), index, s.Value.String())
}

func (s *Set) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "lists_replace_item",
		Values: ast.MakeValues([]ast.Expr{s.List, s.Index, s.Value}, "LIST", "NUM", "ITEM"),
	}
}

func (s *Set) Continuous() bool {
	return false
}

func (s *Set) Consumable() bool {
	return false
}

func (s *Set) Signature() []ast.Signature {
	s.List.Signature()
	s.Index.Signature()
	s.Value.Signature()
	listSigs := s.List.Signature()
	if !ast.HasSignature(listSigs, ast.SignList) {
		s.Where.TypeError("List index assignment requires a list value, but got %", ast.FormatSignatures(listSigs))
	}
	return []ast.Signature{ast.SignVoid}
}
