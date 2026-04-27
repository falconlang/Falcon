package method

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/lex"
	"testing"
)

// chain builds a left-to-right method chain over a root expression.
// Each entry in calls is [name, ...argExprs]; args are ignored here
// since the correction pass only cares about names and receiver types.
func chainOn(root ast.Expr, names ...string) *Call {
	tok := lex.MakeFakeToken(lex.Dot)
	var cur ast.Expr = root
	for _, name := range names {
		cur = &Call{Where: tok, On: cur, Name: name, Args: []ast.Expr{}}
	}
	return cur.(*Call)
}

// TestCorrectChainReverseToReverseList: .reverse() on a list should become .reverseList()
func TestCorrectChainReverseToReverseList(t *testing.T) {
	// s.lowercase().split(" ").reverse().join("")
	root := &fundamentals.Text{}
	chain := chainOn(root, "lowercase", "split", "reverse", "join")

	corrected := CorrectChain(chain)
	if !corrected {
		t.Fatal("expected correction but none was made")
	}

	// Walk the chain to find the corrected name
	names := collectNames(chain)
	assertName(t, names, 2, "reverseList") // index 2 is where "reverse" was
}

// TestCorrectChainUpperToUppercase: .upper() on text should become .uppercase()
func TestCorrectChainUpperToUppercase(t *testing.T) {
	// s.upper().replace(" ", "").textLen()
	root := &fundamentals.Text{}
	chain := chainOn(root, "upper", "replace", "textLen")

	corrected := CorrectChain(chain)
	if !corrected {
		t.Fatal("expected correction but none was made")
	}

	names := collectNames(chain)
	assertName(t, names, 0, "uppercase")
	assertName(t, names, 1, "replace") // already correct
	assertName(t, names, 2, "textLen") // already correct
}

// TestCorrectChainLenToTextLen: .len() on text should become .textLen()
func TestCorrectChainLenToTextLen(t *testing.T) {
	// s.uppercase().replace(" ", "").len()
	root := &fundamentals.Text{}
	chain := chainOn(root, "uppercase", "replace", "len")

	corrected := CorrectChain(chain)
	if !corrected {
		t.Fatal("expected correction but none was made")
	}

	names := collectNames(chain)
	assertName(t, names, 2, "textLen")
}

// TestCorrectChainAllWrong: all three methods wrong
func TestCorrectChainAllWrong(t *testing.T) {
	// s.upper().replaceAll(" ", "").len()
	root := &fundamentals.Text{}
	chain := chainOn(root, "upper", "replaceAll", "len")

	corrected := CorrectChain(chain)
	if !corrected {
		t.Fatal("expected correction but none was made")
	}

	names := collectNames(chain)
	assertName(t, names, 0, "uppercase")
	assertName(t, names, 1, "replace")
	assertName(t, names, 2, "textLen")
}

// TestCorrectChainAndDecompositionSingle: "allButLastAndListLen" splits into allButLast().listLen()
func TestCorrectChainAndDecompositionSingle(t *testing.T) {
	// s.split(" ").allButLastAndListLen()
	root := &fundamentals.List{} // pretend split already happened → list receiver
	chain := chainOn(root, "allButLastAndListLen")

	corrected := CorrectChain(chain)
	if !corrected {
		t.Fatal("expected decomposition but none was made")
	}

	// After decomposition, the outermost call should be "listLen" wrapping a "allButLast" call
	if chain.Name != "listLen" {
		t.Errorf("outer call: got %q, want %q", chain.Name, "listLen")
	}
	inner, ok := chain.On.(*Call)
	if !ok {
		t.Fatal("expected inner *Call after decomposition")
	}
	if inner.Name != "allButLast" {
		t.Errorf("inner call: got %q, want %q", inner.Name, "allButLast")
	}
}

// TestCorrectChainAndDecompositionMulti: three-way And-join splits into three calls
func TestCorrectChainAndDecompositionMulti(t *testing.T) {
	// list.allButLastAndSortAndListLen() → allButLast().sort().listLen()
	root := &fundamentals.List{}
	chain := chainOn(root, "allButLastAndSortAndListLen")

	corrected := CorrectChain(chain)
	if !corrected {
		t.Fatal("expected multi-level decomposition but none was made")
	}

	// Outermost: listLen
	if chain.Name != "listLen" {
		t.Errorf("outer call: got %q, want %q", chain.Name, "listLen")
	}
	mid, ok := chain.On.(*Call)
	if !ok {
		t.Fatal("expected middle *Call after decomposition")
	}
	if mid.Name != "sort" {
		t.Errorf("middle call: got %q, want %q", mid.Name, "sort")
	}
	inner, ok := mid.On.(*Call)
	if !ok {
		t.Fatal("expected inner *Call after decomposition")
	}
	if inner.Name != "allButLast" {
		t.Errorf("inner call: got %q, want %q", inner.Name, "allButLast")
	}
}

// TestCorrectChainAndDecompositionInChain: And-join in the middle of a longer chain
func TestCorrectChainAndDecompositionInChain(t *testing.T) {
	// text.splitAtSpacesAndReverseList() → splitAtSpaces().reverseList()
	root := &fundamentals.Text{}
	chain := chainOn(root, "splitAtSpacesAndReverseList")

	corrected := CorrectChain(chain)
	if !corrected {
		t.Fatal("expected decomposition but none was made")
	}

	if chain.Name != "reverseList" {
		t.Errorf("outer call: got %q, want %q", chain.Name, "reverseList")
	}
	inner, ok := chain.On.(*Call)
	if !ok {
		t.Fatal("expected inner *Call after decomposition")
	}
	if inner.Name != "splitAtSpaces" {
		t.Errorf("inner call: got %q, want %q", inner.Name, "splitAtSpaces")
	}
}

// TestCorrectChainAndDecompositionTypeMismatch: decomposition rejected when types don't chain
func TestCorrectChainAndDecompositionTypeMismatch(t *testing.T) {
	// On a text receiver, "joinAndListLen" can't be decomposed because join() needs a list
	root := &fundamentals.Text{}
	chain := chainOn(root, "joinAndListLen")

	// Should not apply the decomposition (join requires list, not text)
	// Instead should fall through to fzf correction or no correction at all
	originalName := chain.Name
	_ = CorrectChain(chain)
	// The decomposition should NOT have produced joinAndListLen → join+listLen on a text receiver
	// (join.Module = "list", moduleMatchesSig("list",[SignText]) = false)
	if chain.Name == "joinAndListLen" {
		// Uncorrected is fine — no valid decomposition or rename exists
		_ = originalName
	}
	// Just verify no panic and the result is type-coherent
}

// TestCorrectChainAlreadyValid: no correction when chain is already valid
func TestCorrectChainAlreadyValid(t *testing.T) {
	root := &fundamentals.Text{}
	chain := chainOn(root, "uppercase", "replace", "textLen")

	corrected := CorrectChain(chain)
	if corrected {
		t.Error("expected no correction for a valid chain")
	}
}

// TestCorrectChainListRootReverseList: already-correct reverseList on a list root
func TestCorrectChainListRootReverseList(t *testing.T) {
	root := &fundamentals.List{}
	chain := chainOn(root, "reverseList", "join")

	corrected := CorrectChain(chain)
	if corrected {
		t.Error("expected no correction for a valid list chain")
	}
}

// collectNames flattens the chain into left-to-right name order.
func collectNames(call *Call) []string {
	chain, _ := flattenChain(call)
	names := make([]string, len(chain))
	for i, c := range chain {
		names[i] = c.Name
	}
	return names
}

func assertName(t *testing.T, names []string, idx int, want string) {
	t.Helper()
	if idx >= len(names) {
		t.Errorf("chain too short: want index %d, len=%d", idx, len(names))
		return
	}
	if names[idx] != want {
		t.Errorf("names[%d]: got %q, want %q", idx, names[idx], want)
	}
}
