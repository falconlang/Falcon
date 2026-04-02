package list

import (
	"Falcon/code/ast"
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
	return sugar.Format(pFormat, s.List.String(), s.Index.String(), s.Value.String())
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

func (s *Set) Consumable(flags ...bool) bool {
	return false
}

func (s *Set) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
