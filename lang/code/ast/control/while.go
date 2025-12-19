package control

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type While struct {
	Condition ast.Expr
	Body      []ast.Expr
}

func (w *While) String() string {
	return sugar.Format("while (%) {\n%}", w.Condition.String(), ast.PadBody(w.Body))
}

func (w *While) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:       "controls_while",
		Values:     []ast.Value{{Name: "TEST", Block: w.Condition.Blockly(false)}},
		Statements: ast.OptionalStatement("DO", w.Body),
	}
}

func (w *While) Continuous() bool {
	return false
}

func (w *While) Consumable(flags ...bool) bool {
	return false
}

func (w *While) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
