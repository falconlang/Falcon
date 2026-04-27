package common

import (
	"Falcon/code/ast"
	"testing"
)

func TestEmptySocketSignatureIsNumb(t *testing.T) {
	e := &EmptySocket{}
	sig := e.Signature()
	if len(sig) != 1 || sig[0] != ast.SignNumb {
		t.Errorf("Expected SignNumb for EmptySocket, got %v", sig)
	}
}
