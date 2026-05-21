package fundamentals

import "Falcon/code/ast"

type Yield struct {
	Expr            ast.Expr
	TransformedExpr ast.Expr
	UseTransformed  bool
}

func (y *Yield) GetExpr() ast.Expr {
	if y.UseTransformed {
		return y.TransformedExpr
	}
	return y.Expr
}

func (y *Yield) String() string {
	if y.UseTransformed {
		return y.TransformedExpr.String()
	}
	return y.Expr.String()
}

func (y *Yield) Blockly(flags ...bool) ast.Block {
	if y.UseTransformed {
		return y.TransformedExpr.Blockly(flags...)
	}
	return y.Expr.Blockly(flags...)
}

func (y *Yield) Continuous() bool {
	if y.UseTransformed {
		return y.TransformedExpr.Continuous()
	}
	return y.Expr.Continuous()
}

func (y *Yield) Consumable() bool {
	if y.UseTransformed {
		return y.TransformedExpr.Consumable()
	}
	return y.Expr.Consumable()
}

func (y *Yield) Signature() []ast.Signature {
	if y.UseTransformed {
		return y.TransformedExpr.Signature()
	}
	return y.Expr.Signature()
}
