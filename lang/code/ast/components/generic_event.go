package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
	"strings"
)

type GenericEvent struct {
	ComponentType string
	Event         string
	Parameters    []string
	Body          []ast.Expr
}

func (g *GenericEvent) String() string {
	pFormat := "when any %.%(%) {\n%}"
	return sugar.Format(pFormat, g.ComponentType, g.Event, strings.Join(g.Parameters, ", "), ast.PadBody(g.Body))
}

func (g *GenericEvent) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type: "component_event",
		Mutation: &ast.Mutation{
			IsGeneric:     true,
			EventName:     g.Event,
			ComponentType: g.ComponentType,
			Args:          ast.ToArgs(g.Parameters),
		},
		Statements: ast.OptionalStatement("DO", g.Body),
	}
}

func (g *GenericEvent) Continuous() bool {
	return false
}

func (g *GenericEvent) Consumable(flags ...bool) bool {
	return false
}

func (g *GenericEvent) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
