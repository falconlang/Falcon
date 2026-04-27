package method

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/lex"
	"testing"
)

func TestCallAllButFirstSignature(t *testing.T) {
	c := &Call{
		Where: lex.MakeFakeToken(lex.Dot),
		On:    &fundamentals.List{},
		Name:  "allButFirst",
		Args:  []ast.Expr{},
	}
	sig := c.Signature()
	if len(sig) != 1 || sig[0] != ast.SignList {
		t.Errorf("Expected SignList for allButFirst, got %v", sig)
	}
}

func TestCallAllButLastSignature(t *testing.T) {
	c := &Call{
		Where: lex.MakeFakeToken(lex.Dot),
		On:    &fundamentals.List{},
		Name:  "allButLast",
		Args:  []ast.Expr{},
	}
	sig := c.Signature()
	if len(sig) != 1 || sig[0] != ast.SignList {
		t.Errorf("Expected SignList for allButLast, got %v", sig)
	}
}

func TestCallMergeIntoConsumable(t *testing.T) {
	c := &Call{
		Where: lex.MakeFakeToken(lex.Dot),
		On:    &fundamentals.Dictionary{},
		Name:  "mergeInto",
		Args:  []ast.Expr{&fundamentals.Dictionary{}},
	}
	if !c.Consumable() {
		t.Error("Expected mergeInto to be consumable")
	}
}

func TestCallMergeIntoSignature(t *testing.T) {
	c := &Call{
		Where: lex.MakeFakeToken(lex.Dot),
		On:    &fundamentals.Dictionary{},
		Name:  "mergeInto",
		Args:  []ast.Expr{&fundamentals.Dictionary{}},
	}
	sig := c.Signature()
	if len(sig) != 1 || sig[0] != ast.SignDict {
		t.Errorf("Expected SignDict for mergeInto, got %v", sig)
	}
}
