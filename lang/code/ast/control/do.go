package control

import (
	"Falcon/code/ast"
)

type Do struct {
	Body   []ast.Expr
	Result ast.Expr
}

func (d *Do) String() string {
	return ast.JoinExprs("\n", d.Body) + "\n" + d.Result.String()
}

func (d *Do) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:       "controls_do_then_return",
		Statements: ast.OptionalStatement("STM", d.Body),
		Values:     []ast.Value{{Name: "VALUE", Block: d.Result.Blockly(false)}},
	}
}

func (d *Do) Continuous() bool {
	return false
}

func (d *Do) Consumable(flags ...bool) bool {
	return false
}

func (d *Do) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
