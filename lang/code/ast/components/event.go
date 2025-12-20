package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
	"strings"
)

type Event struct {
	ComponentName string
	ComponentType string
	Event         string
	Parameters    []string
	Body          []ast.Expr
}

func (e *Event) String() string {
	pFormat := "when %.%(%) {\n%}"
	return sugar.Format(pFormat, e.ComponentName, e.Event, strings.Join(e.Parameters, ", "), ast.PadBody(e.Body))
}

func (e *Event) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type: "component_event",
		Mutation: &ast.Mutation{
			IsGeneric:     false,
			InstanceName:  e.ComponentName,
			EventName:     e.Event,
			ComponentType: e.ComponentType,
			Args:          ast.ToArgs(e.Parameters),
		},
		Fields:     []ast.Field{{Name: "COMPONENT_SELECTOR", Value: e.ComponentName}},
		Statements: ast.OptionalStatement("DO", e.Body),
	}
}

func (e *Event) Continuous() bool {
	return false
}

func (e *Event) Consumable(flags ...bool) bool {
	return false
}

func (e *Event) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
