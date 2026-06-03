package components

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type PropertyGet struct {
	ComponentName string
	ComponentType string
	Property      string
}

func (p *PropertyGet) String() string {
	return sugar.Format("%.%", ast.FormatName(p.ComponentName), ast.FormatName(p.Property))
}

func (p *PropertyGet) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type: "component_set_get",
		Mutation: &ast.Mutation{
			SetOrGet:      "get",
			PropertyName:  p.Property,
			IsGeneric:     false,
			InstanceName:  p.ComponentName,
			ComponentType: p.ComponentType,
		},
		Fields: []ast.Field{
			{Name: "COMPONENT_SELECTOR", Value: p.ComponentName},
			{Name: "PROP", Value: p.Property},
		},
	}
}

func (p *PropertyGet) Continuous() bool {
	return false
}

func (p *PropertyGet) Consumable() bool {
	return true
}

func (p *PropertyGet) Signature() []ast.Signature {
	return []ast.Signature{ast.SignAny}
}
