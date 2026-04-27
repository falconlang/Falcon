package control

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"testing"
)

func TestDoSignatureReturnsResultSignature(t *testing.T) {
	d := &Do{
		Body:   []ast.Expr{&fundamentals.Number{Content: "1"}},
		Result: &fundamentals.Text{Content: "hello"},
	}
	sig := d.Signature()
	if len(sig) != 1 || sig[0] != ast.SignText {
		t.Errorf("Expected SignText, got %v", sig)
	}
}

func TestDoSignatureWithNumberResult(t *testing.T) {
	d := &Do{
		Body:   []ast.Expr{},
		Result: &fundamentals.Number{Content: "42"},
	}
	sig := d.Signature()
	if len(sig) != 1 || sig[0] != ast.SignNumb {
		t.Errorf("Expected SignNumb, got %v", sig)
	}
}

func TestDoSignatureWithBoolResult(t *testing.T) {
	d := &Do{
		Body:   []ast.Expr{},
		Result: &fundamentals.Boolean{Value: true},
	}
	sig := d.Signature()
	if len(sig) != 1 || sig[0] != ast.SignBool {
		t.Errorf("Expected SignBool, got %v", sig)
	}
}
