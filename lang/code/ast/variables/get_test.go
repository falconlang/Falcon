package variables

import (
	"Falcon/code/ast"
	"testing"
)

func TestGetSignatureUsesValueSignature(t *testing.T) {
	g := &Get{
		Global:         false,
		Name:           "x",
		ValueSignature: []ast.Signature{ast.SignNumb},
	}
	sig := g.Signature()
	if len(sig) != 1 || sig[0] != ast.SignNumb {
		t.Errorf("Expected SignNumb from ValueSignature, got %v", sig)
	}
}

func TestGetSignatureUsesValueSignatureText(t *testing.T) {
	g := &Get{
		Global:         false,
		Name:           "msg",
		ValueSignature: []ast.Signature{ast.SignText},
	}
	sig := g.Signature()
	if len(sig) != 1 || sig[0] != ast.SignText {
		t.Errorf("Expected SignText from ValueSignature, got %v", sig)
	}
}

func TestGetSignatureFallsBackToAny(t *testing.T) {
	g := &Get{
		Global:         false,
		Name:           "unknown",
		ValueSignature: nil,
	}
	sig := g.Signature()
	if len(sig) != 1 || sig[0] != ast.SignAny {
		t.Errorf("Expected SignAny fallback, got %v", sig)
	}
}

func TestGetSignatureFallsBackToAnyForEmpty(t *testing.T) {
	g := &Get{
		Global:         false,
		Name:           "unknown",
		ValueSignature: []ast.Signature{},
	}
	sig := g.Signature()
	if len(sig) != 1 || sig[0] != ast.SignAny {
		t.Errorf("Expected SignAny fallback for empty ValueSignature, got %v", sig)
	}
}
