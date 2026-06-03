package variables

import (
	"Falcon/code/ast"
)

type Global struct {
	Name  string
	Value ast.Expr
}

func (g *Global) String() string {
	return "global " + ast.FormatName(g.Name) + " = " + g.Value.String()
}

func (g *Global) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "global_declaration",
		Fields: []ast.Field{{Name: "NAME", Value: g.Name}},
		Values: []ast.Value{{Name: "VALUE", Block: g.Value.Blockly(false)}},
	}
}

func (g *Global) Continuous() bool {
	return false
}

func (g *Global) Consumable() bool {
	return false
}

func (g *Global) Signature() []ast.Signature {
	g.Value.Signature()
	return []ast.Signature{ast.SignVoid}
}
