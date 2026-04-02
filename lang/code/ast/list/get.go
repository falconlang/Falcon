package list

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
	"Falcon/code/sugar"
)

type Get struct {
	Where *lex.Token
	List  ast.Expr
	Index ast.Expr
}

func (g *Get) String() string {
	pFormat := "%[%]"
	if !g.List.Continuous() {
		pFormat = "(%)[%]"
	}
	return sugar.Format(pFormat, g.List.String(), g.Index.String())
}

func (g *Get) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "lists_select_item",
		Values: ast.MakeValues([]ast.Expr{g.List, g.Index}, "LIST", "NUM"),
	}
}

func (g *Get) Continuous() bool {
	return true
}

func (g *Get) Consumable(flags ...bool) bool {
	return true
}

func (g *Get) Signature() []ast.Signature {
	return []ast.Signature{ast.SignAny}
}
