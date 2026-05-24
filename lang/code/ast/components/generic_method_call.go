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
	Returning     bool
}

func (g *GenericMethodCall) String() string {
	var callType string
	if g.Returning {
		callType = "vcall"
	} else {
		callType = "call"
	}
	return sugar.Format("%(\"%\", %, \"%\", %)", callType, g.ComponentType, g.Component.String(), g.Method, ast.JoinExprs(", ", g.Args))
}

func (g *GenericMethodCall) Blockly(flags ...bool) ast.Block {
	shape := "statement"
	if g.Returning {
		shape = "value"
	}
	return ast.Block{
		Type: "component_method",
		Mutation: &ast.Mutation{
			MethodName:    g.Method,
			IsGeneric:     true,
			ComponentType: g.ComponentType,
			Shape:         shape,
		},
		Values: ast.ValueArgsByPrefix(g.Component, "COMPONENT", "ARG", g.Args),
	}
}

func (g *GenericMethodCall) Continuous() bool {
	return false
}

func (g *GenericMethodCall) Consumable() bool {
	return g.Returning
}

func (g *GenericMethodCall) Signature() []ast.Signature {
	g.Component.Signature()
	for _, arg := range g.Args {
		arg.Signature()
	}
	return []ast.Signature{ast.SignAny}
}
