package control

import (
	"Falcon/code/ast"
)

// Yield is a statement that immediately exits the enclosing function and
// returns the given value to the caller. It behaves like an early return.
type Yield struct {
	Expr ast.Expr
}

func (y *Yield) String() string {
	return "yield " + y.Expr.String()
}

func (y *Yield) Blockly(flags ...bool) ast.Block {
	return ast.Block{Type: "controls_yield"}
}

func (y *Yield) Continuous() bool {
	return true
}

func (y *Yield) Consumable(flags ...bool) bool {
	return false
}

func (y *Yield) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
