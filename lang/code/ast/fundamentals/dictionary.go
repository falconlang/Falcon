package fundamentals

import (
	"Falcon/code/ast"
	"Falcon/code/sugar"
)

type Dictionary struct {
	Elements []ast.Expr
}

func (d *Dictionary) String() string {
	return sugar.Format("{ % }", ast.JoinExprs(", ", d.Elements))
}

func (d *Dictionary) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "dictionaries_create_with",
		Mutation: &ast.Mutation{ItemCount: len(d.Elements)},
		Values:   ast.ValuesByPrefix("ADD", d.Elements),
	}
}

func (d *Dictionary) Continuous() bool {
	return true
}

func (d *Dictionary) Consumable() bool {
	return true
}

func (d *Dictionary) Signature() []ast.Signature {
	return []ast.Signature{ast.SignDict}
}

type Pair struct {
	Key   ast.Expr
	Value ast.Expr
}

func (p *Pair) String() string {
	return sugar.Format("% : %", p.Key.String(), p.Value.String())
}

func (p *Pair) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "pair",
		Values: ast.MakeValues([]ast.Expr{p.Key, p.Value}, "KEY", "VALUE"),
	}
}

func (p *Pair) Continuous() bool {
	return false
}

func (p *Pair) Consumable() bool {
	return true
}

func (p *Pair) Signature() []ast.Signature {
	return []ast.Signature{ast.SignList}
}

type WalkAll struct {
}

func (w *WalkAll) String() string {
	return "walkAll"
}

func (w *WalkAll) Blockly(flags ...bool) ast.Block {
	return ast.Block{Type: "dictionaries_walk_all"}
}

func (w *WalkAll) Continuous() bool {
	return true
}

func (w *WalkAll) Consumable() bool {
	return true
}

func (w *WalkAll) Signature() []ast.Signature {
	return []ast.Signature{ast.SignList}
}
