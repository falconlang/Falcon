package variables

import (
	"Falcon/code/ast"
	"strings"
)

type VarStack struct {
	Names  []string
	Values []ast.Expr
}

func (v *VarStack) String() string {
	var builder strings.Builder
	localLines := make([]string, len(v.Names))
	for k, name := range v.Names {
		localLines[k] = "local " + name + " = " + v.Values[k].String()
	}
	builder.WriteString(strings.Join(localLines, "\n"))
	return builder.String()
}

func (v *VarStack) Blockly(flags ...bool) ast.Block {
	panic("cannot call Blockly() on a VarStack")
}

func (v *VarStack) Continuous() bool {
	return false
}

func (v *VarStack) Consumable() bool {
	return false
}

func (v *VarStack) Signature() []ast.Signature {
	for _, value := range v.Values {
		value.Signature()
	}
	return []ast.Signature{ast.SignVoid}
}
