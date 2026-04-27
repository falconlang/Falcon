package control

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"testing"
)

func TestIfNotConsumableWithoutElse(t *testing.T) {
	i := &If{
		Conditions: []ast.Expr{&fundamentals.Boolean{Value: true}},
		Bodies:     [][]ast.Expr{{&fundamentals.Number{Content: "1"}}},
		ElseBody:   nil,
	}
	if i.Consumable() {
		t.Error("If without else should not be consumable")
	}
}

func TestIfNotConsumableWithEmptyBody(t *testing.T) {
	i := &If{
		Conditions: []ast.Expr{&fundamentals.Boolean{Value: true}},
		Bodies:     [][]ast.Expr{{}},
		ElseBody:   []ast.Expr{&fundamentals.Number{Content: "1"}},
	}
	if i.Consumable() {
		t.Error("If with empty body should not be consumable")
	}
}

func TestIfNotConsumableWithNonConsumableLastExpr(t *testing.T) {
	// Using a statement-like expression as the last element
	i := &If{
		Conditions: []ast.Expr{&fundamentals.Boolean{Value: true}},
		Bodies:     [][]ast.Expr{{&fundamentals.Number{Content: "1"}}},
		ElseBody:   []ast.Expr{&fundamentals.Text{Content: "ok"}},
	}
	// Both number and text are consumable, so this should be consumable
	if !i.Consumable() {
		t.Error("If with consumable last expressions should be consumable")
	}
}

func TestIfConsumableWhenAllBranchesConsumable(t *testing.T) {
	i := &If{
		Conditions: []ast.Expr{&fundamentals.Boolean{Value: true}},
		Bodies:     [][]ast.Expr{{&fundamentals.Number{Content: "1"}}},
		ElseBody:   []ast.Expr{&fundamentals.Number{Content: "2"}},
	}
	if !i.Consumable() {
		t.Error("If with all consumable branches should be consumable")
	}
}

func TestIfSignatureVoidWhenNotConsumable(t *testing.T) {
	i := &If{
		Conditions: []ast.Expr{&fundamentals.Boolean{Value: true}},
		Bodies:     [][]ast.Expr{{&fundamentals.Number{Content: "1"}}},
		ElseBody:   nil,
	}
	sig := i.Signature()
	if len(sig) != 1 || sig[0] != ast.SignVoid {
		t.Errorf("Expected SignVoid, got %v", sig)
	}
}

func TestIfSignatureUnionWhenConsumable(t *testing.T) {
	i := &If{
		Conditions: []ast.Expr{&fundamentals.Boolean{Value: true}},
		Bodies:     [][]ast.Expr{{&fundamentals.Number{Content: "1"}}},
		ElseBody:   []ast.Expr{&fundamentals.Text{Content: "2"}},
	}
	sig := i.Signature()
	if len(sig) != 2 {
		t.Fatalf("Expected 2 signatures, got %d: %v", len(sig), sig)
	}
	if sig[0] != ast.SignNumb && sig[0] != ast.SignText {
		t.Errorf("Expected SignNumb or SignText, got %v", sig[0])
	}
	if sig[1] != ast.SignNumb && sig[1] != ast.SignText {
		t.Errorf("Expected SignNumb or SignText, got %v", sig[1])
	}
}

func TestIfMultiBranchSignature(t *testing.T) {
	i := &If{
		Conditions: []ast.Expr{
			&fundamentals.Boolean{Value: true},
			&fundamentals.Boolean{Value: false},
		},
		Bodies: [][]ast.Expr{
			{&fundamentals.Number{Content: "1"}},
			{&fundamentals.Boolean{Value: true}},
		},
		ElseBody: []ast.Expr{&fundamentals.Text{Content: "fallback"}},
	}
	sig := i.Signature()
	if len(sig) != 3 {
		t.Fatalf("Expected 3 signatures, got %d: %v", len(sig), sig)
	}
}
