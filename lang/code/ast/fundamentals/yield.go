package fundamentals

import "Falcon/code/ast"

type Yield struct {
	Expr            ast.Expr
	TransformedExpr ast.Expr
	Revert          bool
}

func (y *Yield) GetExpr() ast.Expr {
	if y.Revert {
		return y.Expr
	}
	return y.TransformedExpr
}

func (y *Yield) String() string {
	if y.Revert {
		return y.Expr.String()
	}
	return y.TransformedExpr.String()
}

func (y *Yield) Blockly(flags ...bool) ast.Block {
	if y.Revert {
		return y.Expr.Blockly(flags...)
	}
	return y.TransformedExpr.Blockly(flags...)
}

func (y *Yield) Continuous() bool {
	if y.Revert {
		return y.Expr.Continuous()
	}
	return y.TransformedExpr.Continuous()
}

func (y *Yield) Consumable() bool {
	if y.Revert {
		return y.Expr.Consumable()
	}
	return y.TransformedExpr.Consumable()
}

func (y *Yield) Signature() []ast.Signature {
	if y.Revert {
		return y.Expr.Signature()
	}
	return y.TransformedExpr.Signature()
}
