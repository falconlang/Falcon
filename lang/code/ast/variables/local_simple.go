package variables

import (
	"Falcon/code/ast"
	"strings"
)

type SimpleVar struct {
	Name  string
	Value ast.Expr
	Body  []ast.Expr
}

func (v *SimpleVar) String() string {
	var builder strings.Builder
	builder.WriteString("local ")
	builder.WriteString(v.Name)
	builder.WriteString(" = ")
	builder.WriteString(v.Value.String())
	builder.WriteString("\n")
	builder.WriteString(ast.JoinExprs("\n", v.Body))
	return builder.String()
}

func (v *SimpleVar) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:       "local_declaration_statement",
		Mutation:   &ast.Mutation{LocalNames: ast.MakeLocalNames(v.Name)},
		Fields:     []ast.Field{{Name: "VAR0", Value: v.Name}},
		Values:     []ast.Value{{Name: "DECL0", Block: v.Value.Blockly(false)}},
		Statements: ast.OptionalStatement("STACK", v.Body),
	}
}

func (v *SimpleVar) Continuous() bool {
	return false
}

func (v *SimpleVar) Consumable(flags ...bool) bool {
	return false
}

func (v *SimpleVar) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
