package variables

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
)

type Get struct {
	Where          *lex.Token
	Global         bool
	Name           string
	ValueSignature []ast.Signature
}

func (g *Get) String() string {
	if g.Global {
		return "this." + g.Name
	}
	return g.Name
}

func (g *Get) Blockly(flags ...bool) ast.Block {
	// check if this refers to an event parameter
	if len(g.ValueSignature) > 0 && g.ValueSignature[0] == ast.SignOfEvent {
		return ast.Block{
			Type:     "lexical_variable_get",
			Mutation: &ast.Mutation{EventParams: []ast.EventParam{{Name: g.Name}}},
			Fields:   []ast.Field{{Name: "VAR", Value: g.Name}},
		}
	}
	var name string
	if g.Global {
		name = "global " + g.Name
	} else {
		name = g.Name
	}
	return ast.Block{
		Type:   "lexical_variable_get",
		Fields: []ast.Field{{Name: "VAR", Value: name}},
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
