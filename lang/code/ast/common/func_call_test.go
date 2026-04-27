package common

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"testing"
)

func TestFuncCallEverySignature(t *testing.T) {
	f := MakeFuncCall("every", &fundamentals.Text{Content: "Button"})
	sig := f.Signature()
	if len(sig) != 1 || sig[0] != ast.SignList {
		t.Errorf("Expected SignList for every(), got %v", sig)
	}
}

func TestFuncCallDecToHexSignature(t *testing.T) {
	f := MakeFuncCall("decToHex", &fundamentals.Number{Content: "255"})
	sig := f.Signature()
	if len(sig) != 1 || sig[0] != ast.SignText {
		t.Errorf("Expected SignText for decToHex(), got %v", sig)
	}
}

func TestFuncCallDecToBinSignature(t *testing.T) {
	f := MakeFuncCall("decToBin", &fundamentals.Number{Content: "5"})
	sig := f.Signature()
	if len(sig) != 1 || sig[0] != ast.SignText {
		t.Errorf("Expected SignText for decToBin(), got %v", sig)
	}
}

func TestFuncCallSqrtSignature(t *testing.T) {
	f := MakeFuncCall("sqrt", &fundamentals.Number{Content: "4"})
	sig := f.Signature()
	if len(sig) != 1 || sig[0] != ast.SignNumb {
		t.Errorf("Expected SignNumb for sqrt(), got %v", sig)
	}
}
