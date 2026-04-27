package fundamentals

import (
	"Falcon/code/ast"
)

type Component struct {
	Name string
	Type string
}

func (c *Component) String() string {
	return c.Name
}

func (c *Component) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "component_component_block",
		Mutation: &ast.Mutation{InstanceName: c.Name, ComponentType: c.Type},
		Fields:   []ast.Field{{Name: "COMPONENT_SELECTOR", Value: c.Name}},
	}
}

func (c *Component) Continuous() bool {
	return true
}

func (c *Component) Consumable() bool {
	return true
}

func (c *Component) Signature() []ast.Signature {
	return []ast.Signature{ast.SignComponent}
}
