package control

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
	"Falcon/code/sugar"
)

type For struct {
	Where *lex.Token
	IName string
	From  ast.Expr
	To    ast.Expr
	By    ast.Expr
	Body  []ast.Expr
}

func (f *For) String() string {
	return sugar.Format("for (%: % .. % step %) {\n%}",
		ast.FormatName(f.IName), f.From.String(), f.To.String(), f.By.String(), ast.PadBody(f.Body))
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

func (f *For) Consumable() bool {
	return false
}

func (f *For) Signature() []ast.Signature {
	f.From.Signature()
	f.To.Signature()
	f.By.Signature()
	for _, expr := range f.Body {
		expr.Signature()
	}
	return []ast.Signature{ast.SignVoid}
}
