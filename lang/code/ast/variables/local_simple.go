package variables

import (
	"Falcon/code/ast"
	"strings"
)

type SimpleVar struct {
	Name  string
	Value ast.Expr
}

func (v *SimpleVar) String() string {
	var builder strings.Builder
	builder.WriteString("local ")
	builder.WriteString(v.Name)
	builder.WriteString(" = ")
	builder.WriteString(v.Value.String())
	builder.WriteString("\n")
	return builder.String()
}

func (v *SimpleVar) Blockly(flags ...bool) ast.Block {
	panic("cannot call Blockly() on a SimpleVar")
}

func (v *SimpleVar) Continuous() bool {
	return false
}

func (v *SimpleVar) Consumable() bool {
	return false
}

func (v *SimpleVar) Signature() []ast.Signature {
	v.Value.Signature()
	return []ast.Signature{ast.SignVoid}
}
