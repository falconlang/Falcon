package procedures

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
	"strings"
)

type VoidProcedure struct {
	Name       string
	Parameters []string
	Body       []ast.Expr
}

func (v *VoidProcedure) String() string {
	return sugar.Format("func %(%) {\n%}", v.Name, strings.Join(v.Parameters, ", "), ast.PadBody(v.Body))
}

func (v *VoidProcedure) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:       "procedures_defnoreturn",
		Mutation:   &ast.Mutation{Args: ast.ToArgs(v.Parameters)},
		Fields:     append(ast.ToFields("VAR", v.Parameters), ast.Field{Name: "NAME", Value: v.Name}),
		Statements: ast.OptionalStatement("STACK", v.Body),
	}
}

func (v *VoidProcedure) Continuous() bool {
	return false
}

func (v *VoidProcedure) Consumable(flags ...bool) bool {
	return false
}

func (v *VoidProcedure) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
