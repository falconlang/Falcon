package control

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type For struct {
	IName string
	From  ast.Expr
	To    ast.Expr
	By    ast.Expr
	Body  []ast.Expr
}

func (f *For) String() string {
	return sugar.Format("for (%: % .. % step %) {\n%}",
		f.IName, f.From.String(), f.To.String(), f.By.String(), ast.PadBody(f.Body))
}

func (f *For) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:       "controls_forRange",
		Fields:     []ast.Field{{Name: "VAR", Value: f.IName}},
		Values:     ast.MakeValues([]ast.Expr{f.From, f.To, f.By}, "START", "END", "STEP"),
		Statements: ast.OptionalStatement("DO", f.Body),
	}
}

func (f *For) Continuous() bool {
	return false
}

func (f *For) Consumable(flags ...bool) bool {
	return false
}

func (f *For) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
