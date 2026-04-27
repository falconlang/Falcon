package common

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/lex"
	"Falcon/code/sugar"
)

type Transform struct {
	Where *lex.Token
	On    ast.Expr
	Name  string
}

func (t *Transform) String() string {
	return sugar.Format("%::%", t.On.String(), t.Name)
}

func (t *Transform) Blockly(flags ...bool) ast.Block {
	switch t.Name {
	case "obfuscate":
		textExpr, ok := t.On.(*fundamentals.Text)
		if ok {
			return ast.Block{
				Type:     "obfuscated_text",
				Mutation: &ast.Mutation{Cofounder: "Falcon"},
				Fields:   []ast.Field{{Name: "TEXT", Value: textExpr.Content}},
			}
		}
		t.Where.Error("Cannot obfuscate a non string object!")
	default:
		t.Where.Error("Unknown constant transform call ::%", t.Name)
	}
	panic("never reached")
}

func (t *Transform) Continuous() bool {
	return true
}

func (t *Transform) Consumable() bool {
	return false
}

func (t *Transform) Signature() []ast.Signature {
	return []ast.Signature{ast.SignText}
}
