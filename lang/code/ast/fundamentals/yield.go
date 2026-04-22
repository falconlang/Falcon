package fundamentals

import "Falcon/code/ast"

type Yield struct {
	Expr            ast.Expr
	TransformedExpr ast.Expr
	Confirmed       bool
}

func (y *Yield) GetExpr() ast.Expr {
	if y.Confirmed {
		return y.TransformedExpr
	}
	return y.Expr
}

func (y *Yield) String() string {
	if y.Confirmed {
		return y.TransformedExpr.String()
	}
	return y.Expr.String()
}

func (y *Yield) Blockly(flags ...bool) ast.Block {
	if y.Confirmed {
		return y.TransformedExpr.Blockly(flags...)
	}
	return y.Expr.Blockly(flags...)
}

func (y *Yield) Continuous() bool {
	if y.Confirmed {
		return y.TransformedExpr.Continuous()
	}
	return y.Expr.Continuous()
}

func (y *Yield) Consumable(flags ...bool) bool {
	if y.Confirmed {
		return y.TransformedExpr.Consumable(flags...)
	}
	return y.Expr.Consumable(flags...)
}

func (y *Yield) Signature() []ast.Signature {
	if y.Confirmed {
		return y.TransformedExpr.Signature()
	}
	return y.Expr.Signature()
}
