package variables

import (
	"Falcon/code/ast"
	"strings"
)

type Var struct {
	Names  []string
	Values []ast.Expr
	Body   []ast.Expr
}

func (v *Var) String() string {
	var builder strings.Builder
	localLines := make([]string, len(v.Names))
	for k, name := range v.Names {
		localLines[k] = "local " + name + " = " + v.Values[k].String()
	}
	builder.WriteString(strings.Join(localLines, "\n"))
	builder.WriteString("\n")
	builder.WriteString(ast.JoinExprs("\n", v.Body))
	return builder.String()
}

func (v *Var) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:       "local_declaration_statement",
		Mutation:   &ast.Mutation{LocalNames: ast.MakeLocalNames(v.Names...)},
		Fields:     ast.ToFields("VAR", v.Names),
		Values:     ast.ValuesByPrefix("DECL", v.Values),
		Statements: ast.OptionalStatement("STACK", v.Body),
	}
}

func (v *Var) Continuous() bool {
	return false
}

func (v *Var) Consumable(flags ...bool) bool {
	return false
}

func (v *Var) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
