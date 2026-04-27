package fundamentals

import (
	"Falcon/code/ast"
	"testing"
)

func TestWalkAllSignature(t *testing.T) {
	w := &WalkAll{}
	sig := w.Signature()
	if len(sig) != 1 || sig[0] != ast.SignList {
		t.Errorf("Expected SignList for WalkAll, got %v", sig)
	}
}

func TestDictionarySignature(t *testing.T) {
	d := &Dictionary{}
	sig := d.Signature()
	if len(sig) != 1 || sig[0] != ast.SignDict {
		t.Errorf("Expected SignDict for Dictionary, got %v", sig)
	}
}

func TestPairSignature(t *testing.T) {
	p := &Pair{}
	sig := p.Signature()
	if len(sig) != 1 || sig[0] != ast.SignList {
		t.Errorf("Expected SignList for Pair, got %v", sig)
	}
}
