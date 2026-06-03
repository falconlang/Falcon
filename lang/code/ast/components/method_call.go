package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type MethodCall struct {
	ComponentName string
	ComponentType string
	Method        string
	Args          []ast.Expr
}

func (m *MethodCall) String() string {
	return sugar.Format("%.%(%)", ast.FormatName(m.ComponentName), ast.FormatName(m.Method), ast.JoinExprs(", ", m.Args))
}

func (m *MethodCall) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type: "component_method",
		Mutation: &ast.Mutation{
			MethodName:    m.Method,
			IsGeneric:     false,
			InstanceName:  m.ComponentName,
			ComponentType: m.ComponentType,
		},
		Fields: []ast.Field{{Name: "COMPONENT_SELECTOR", Value: m.ComponentName}},
		Values: ast.ValuesByPrefix("ARG", m.Args),
	}
}

func (m *MethodCall) Continuous() bool {
	return false
}

func (m *MethodCall) Consumable() bool {
	return false
}

func (m *MethodCall) Signature() []ast.Signature {
	for _, arg := range m.Args {
		arg.Signature()
	}
	return []ast.Signature{ast.SignAny}
}
