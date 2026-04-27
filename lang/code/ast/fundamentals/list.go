package fundamentals

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type List struct {
	Elements []ast.Expr
}

func (l *List) String() string {
	return sugar.Format("[%]", ast.JoinExprs(", ", l.Elements))
}

func (l *List) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "lists_create_with",
		Mutation: &ast.Mutation{ItemCount: len(l.Elements)},
		Values:   ast.ValuesByPrefix("ADD", l.Elements),
	}
}

func (l *List) Continuous() bool {
	return true
}

func (l *List) Consumable() bool {
	return true
}

func (l *List) Signature() []ast.Signature {
	return []ast.Signature{ast.SignList}
}
