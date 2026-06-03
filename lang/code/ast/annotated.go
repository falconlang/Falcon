package ast

type AnnotatedExpr struct {
	Expr    Expr
	Inline  bool
	Comment string
}

func (a *AnnotatedExpr) String() string {
	return a.Expr.String()
}

func (a *AnnotatedExpr) Blockly(flags ...bool) Block {
	block := a.Expr.Blockly(flags...)
	if a.Inline {
		block.Inline = true
	}
	if a.Comment != "" {
		block.Comment = a.Comment
	}
	return block
}

func (a *AnnotatedExpr) Continuous() bool {
	return a.Expr.Continuous()
}

func (a *AnnotatedExpr) Consumable() bool {
	return a.Expr.Consumable()
}

func (a *AnnotatedExpr) Signature() []Signature {
	return a.Expr.Signature()
}

func UnwrapAnnotated(expr Expr) Expr {
	for {
		annotated, ok := expr.(*AnnotatedExpr)
		if !ok {
			return expr
		}
		expr = annotated.Expr
	}
}
