package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type GenericMethodCall struct {
	Component     ast.Expr
	ComponentType string
	Method        string
	Args          []ast.Expr
}

func (g *GenericMethodCall) String() string {
	return sugar.Format("call(\"%\", %, \"%\", %)", g.ComponentType, g.Component.String(), g.Method, ast.JoinExprs(", ", g.Args))
}

func (g *GenericMethodCall) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type: "component_method",
		Mutation: &ast.Mutation{
			MethodName:    g.Method,
			IsGeneric:     true,
			ComponentType: g.ComponentType,
		},
		Values: ast.ValueArgsByPrefix(g.Component, "COMPONENT", "ARG", g.Args),
	}
}

func (g *GenericMethodCall) Continuous() bool {
	return false
}

func (g *GenericMethodCall) Consumable(flags ...bool) bool {
	return false // play safe, may be consumable too
}

func (g *GenericMethodCall) Signature() []ast.Signature {
	return []ast.Signature{ast.SignAny}
}
