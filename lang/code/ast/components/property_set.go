package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type PropertySet struct {
	ComponentName string
	ComponentType string
	Property      string
	Value         ast.Expr
}

func (p *PropertySet) String() string {
	return sugar.Format("%.% = %", p.ComponentName, p.Property, p.Value.String())
}

func (p *PropertySet) Blockly(flags ...bool) ast.Block {
	newValue := p.Value.Blockly(false)
	// explicitly mark as value for consumption
	if newValue.Mutation != nil {
		newValue.Mutation.Shape = "value"
	}
	return ast.Block{
		Type: "component_set_get",
		Mutation: &ast.Mutation{
			SetOrGet:      "set",
			PropertyName:  p.Property,
			IsGeneric:     false,
			InstanceName:  p.ComponentName,
			ComponentType: p.ComponentType,
		},
		Fields: ast.FieldsFromMap(map[string]string{
			"COMPONENT_SELECTOR": p.ComponentName,
			"PROP":               p.Property,
		}),
		Values: []ast.Value{{Name: "VALUE", Block: newValue}},
	}
}

func (p *PropertySet) Continuous() bool {
	return false
}

func (p *PropertySet) Consumable(flags ...bool) bool {
	return false
}

func (p *PropertySet) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
