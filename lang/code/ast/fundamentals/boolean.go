package fundamentals

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type Boolean struct {
	Value bool
}

func (b *Boolean) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

func (b *Boolean) Blockly(flags ...bool) ast.Block {
	var bText string
	if b.Value {
		bText = "TRUE"
	} else {
		bText = "FALSE"
	}
	return ast.Block{
		Type:   "logic_boolean",
		Fields: ast.FieldsFromMap(map[string]string{"BOOL": bText}),
	}
}

func (b *Boolean) Continuous() bool {
	return true
}

func (b *Boolean) Consumable() bool {
	return true
}

func (b *Boolean) Signature() []ast.Signature {
	return []ast.Signature{ast.SignBool}
}

type Not struct {
	Expr ast.Expr
}

func (n *Not) String() string {
	return sugar.Format("!%", n.Expr.String())
}

func (n *Not) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "logic_negate",
		Values: []ast.Value{{Name: "BOOL", Block: n.Expr.Blockly(false)}},
	}
}

func (n *Not) Continuous() bool {
	return false
}

func (n *Not) Consumable() bool {
	return true
}

func (n *Not) Signature() []ast.Signature {
	return []ast.Signature{ast.SignBool}
}
