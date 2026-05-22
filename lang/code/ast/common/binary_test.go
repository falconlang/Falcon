package common

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/lex"
	"testing"
)

func TestBinaryExprRemainderSignature(t *testing.T) {
	b := &BinaryExpr{
		Where:    lex.MakeFakeToken(lex.Plus),
		Operands: []ast.Expr{&fundamentals.Number{Content: "5"}, &fundamentals.Number{Content: "2"}},
		Operator: lex.Remainder,
	}
	sig := b.Signature()
	if len(sig) != 1 || sig[0] != ast.SignNumb {
		t.Errorf("Expected SignNumb for Remainder operator, got %v", sig)
	}
}

func TestBinaryExprPlusSignature(t *testing.T) {
	b := &BinaryExpr{
		Where:    lex.MakeFakeToken(lex.Plus),
		Operands: []ast.Expr{&fundamentals.Number{Content: "1"}, &fundamentals.Number{Content: "2"}},
		Operator: lex.Plus,
	}
	sig := b.Signature()
	if len(sig) != 1 || sig[0] != ast.SignNumb {
		t.Errorf("Expected SignNumb for Plus operator, got %v", sig)
	}
}

func TestBinaryExprTextJoinSignature(t *testing.T) {
	b := &BinaryExpr{
		Where:    lex.MakeFakeToken(lex.Underscore),
		Operands: []ast.Expr{&fundamentals.Number{Content: "1"}},
		Operator: lex.Underscore,
	}
	sig := b.Signature()
	if len(sig) != 1 || sig[0] != ast.SignText {
		t.Errorf("Expected SignText for Underscore (text join) operator, got %v", sig)
	}
}
