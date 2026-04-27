package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type EveryComponent struct {
	Type string
}

func (e *EveryComponent) String() string {
	return sugar.Format("every(%)", e.Type)
}

func (e *EveryComponent) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "component_all_component_block",
		Mutation: &ast.Mutation{ComponentType: e.Type},
		Fields:   []ast.Field{{Name: "COMPONENT_SELECTOR", Value: e.Type}},
	}
}

func (e *EveryComponent) Continuous() bool {
	return true
}

func (e *EveryComponent) Consumable() bool {
	return true
}

func (e *EveryComponent) Signature() []ast.Signature {
	return []ast.Signature{ast.SignList}
}
