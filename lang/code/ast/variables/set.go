package variables

import "Falcon/code/ast"

type Set struct {
	Global bool
	Name   string
	Expr   ast.Expr
}

func (s Set) String() string {
	if s.Global {
		return "this." + s.Name + " = " + s.Expr.String()
	}
	return s.Name + " = " + s.Expr.String()
}

func (s Set) Blockly(flags ...bool) ast.Block {
	var name string
	if s.Global {
		name = "global " + s.Name
	} else {
		name = s.Name
	}
	return ast.Block{
		Type:   "lexical_variable_set",
		Fields: []ast.Field{{Name: "VAR", Value: name}},
		Values: []ast.Value{{Name: "VALUE", Block: s.Expr.Blockly(false)}},
	}
}

func (s Set) Continuous() bool {
	return false
}

func (s Set) Consumable() bool {
	return false
}

func (s Set) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
