package control

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type Each struct {
	IName    string
	Iterable ast.Expr
	Body     []ast.Expr
}

func (e *Each) String() string {
	return sugar.Format("for (% in %) {\n%}", e.IName, e.Iterable.String(), ast.PadBody(e.Body))
}

func (e *Each) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:       "controls_forEach",
		Fields:     []ast.Field{{Name: "VAR", Value: e.IName}},
		Values:     []ast.Value{{Name: "LIST", Block: e.Iterable.Blockly(false)}},
		Statements: ast.OptionalStatement("DO", e.Body),
	}
}

func (e *Each) Continuous() bool {
	return false
}

func (e *Each) Consumable(flags ...bool) bool {
	return false
}

func (e *Each) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
