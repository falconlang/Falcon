package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type GenericPropertyGet struct {
	Component     ast.Expr
	ComponentType string
	Property      string
}

func (g *GenericPropertyGet) String() string {
	return sugar.Format("get(\"%\", %, \"%\")", g.ComponentType, g.Component.String(), g.Property)
}

func (g *GenericPropertyGet) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type: "component_set_get",
		Mutation: &ast.Mutation{
			SetOrGet:      "get",
			PropertyName:  g.Property,
			IsGeneric:     true,
			ComponentType: g.ComponentType,
		},
		Fields: []ast.Field{{Name: "PROP", Value: g.Property}},
		Values: []ast.Value{{Name: "COMPONENT", Block: g.Component.Blockly(false)}},
	}
}

func (g *GenericPropertyGet) Continuous() bool {
	return false
}

func (g *GenericPropertyGet) Consumable(flags ...bool) bool {
	return true
}

func (g *GenericPropertyGet) Signature() []ast.Signature {
	return []ast.Signature{ast.SignAny}
}
