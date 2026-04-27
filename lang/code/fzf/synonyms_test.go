package fzf

import (
	"strings"
	"testing"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func sharedCanonical(a, b string) bool {
	return Canonical(a) == Canonical(b)
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// ── 1. Full method-token coverage ─────────────────────────────────────────────

// allMethods is the complete list of method names from lang/code/ast/method/call.go.
var allMethods = []string{
	"textLen", "trim", "uppercase", "lowercase", "startsWith",
	"contains", "containsAny", "containsAll",
	"split", "splitAtFirst", "splitAtAny", "splitAtFirstOfAny", "splitAtSpaces",
	"reverse", "csvRowToList", "csvTableToList",
	"segment", "replace", "replaceFrom", "replaceFromLongestFirst",
	"listLen", "add", "containsItem", "indexOf", "insert", "remove",
	"appendList", "lookupInPairs", "join", "slice", "random",
	"reverseList", "toCsvRow", "toCsvTable", "sort",
	"allButFirst", "allButLast", "pairsToDict",
	"dictLen", "get", "set", "delete",
	"getAtPath", "setAtPath", "containsKey", "mergeInto",
	"walkTree", "keys", "values", "toPairs",
}

// exemptTokens are prepositions / conjunctions / articles that carry no
// semantic information useful to a synonym graph.
var exemptTokens = map[string]bool{
	"at": true, "in": true, "into": true, "to": true, "of": true,
	"with": true, "by": true, "on": true, "for": true, "from": true,
}

// TestAllMethodTokensCovered verifies that every meaningful token produced by
// SplitCamel on every Falcon method name has at least one entry in synonymGraph.
func TestAllMethodTokensCovered(t *testing.T) {
	for _, method := range allMethods {
		for _, tok := range SplitCamel(method) {
			tok = strings.ToLower(tok)
			if exemptTokens[tok] {
				continue
			}
			if _, ok := synonymGraph[tok]; !ok {
				t.Errorf("method %q: token %q not in synonymGraph", method, tok)
			}
		}
	}
}

// ── 2. Canonical correctness ──────────────────────────────────────────────────

func TestCanonicalAlphabeticallySmallest(t *testing.T) {
	// {len, length, size, count} → alphabetically smallest is "count"
	for _, w := range []string{"len", "length", "size", "count"} {
		if got := Canonical(w); got != "count" {
			t.Errorf("Canonical(%q)=%q, want %q", w, got, "count")
		}
	}
}

func TestCanonicalReverseSynonyms(t *testing.T) {
	// {reverse, flip, invert, backward, backwards} → alphabetically "backward"
	want := "backward"
	for _, w := range []string{"reverse", "flip", "invert", "backward", "backwards"} {
		if got := Canonical(w); got != want {
			t.Errorf("Canonical(%q)=%q, want %q", w, got, want)
		}
	}
}

func TestCanonicalUnknownIsItself(t *testing.T) {
	for _, w := range []string{"xyzzy_unknown", "frobnicator", "snafubar"} {
		if got := Canonical(w); got != w {
			t.Errorf("Canonical(%q)=%q, want the word itself", w, got)
		}
	}
}

func TestCanonicalTransitivity(t *testing.T) {
	// "size" → "length" → "len" → they all share one canonical
	if !sharedCanonical("size", "len") {
		t.Error("expected size and len to share a canonical via length")
	}
	if !sharedCanonical("flip", "backwards") {
		t.Error("expected flip and backwards to share a canonical via reverse")
	}
}

func TestCanonicalCaseInsensitive(t *testing.T) {
	pairs := [][2]string{{"Flip", "flip"}, {"SIZE", "size"}, {"REVERSE", "reverse"}}
	for _, p := range pairs {
		if Canonical(p[0]) != Canonical(p[1]) {
			t.Errorf("Canonical(%q) != Canonical(%q): not case-insensitive", p[0], p[1])
		}
	}
}

func TestCanonicalExemptPrepositions(t *testing.T) {
	// Prepositions are not in the graph; each returns itself.
	for tok := range exemptTokens {
		if got := Canonical(tok); got != tok {
			t.Errorf("Canonical(%q)=%q, want itself (preposition not in graph)", tok, got)
		}
	}
}

// ── 3. CanonicalTokens ────────────────────────────────────────────────────────

func TestCanonicalTokensSplitsAndCanonicalises(t *testing.T) {
	// "reverseList" → ["reverse","list"] → [Canonical("reverse"), Canonical("list")]
	ct := CanonicalTokens("reverseList")
	flipCan := Canonical("flip")   // same cluster as reverse
	listCan := Canonical("list")   // "arr"
	if !containsStr(ct, flipCan) {
		t.Errorf("CanonicalTokens(reverseList)=%v missing reverse canonical %q", ct, flipCan)
	}
	if !containsStr(ct, listCan) {
		t.Errorf("CanonicalTokens(reverseList)=%v missing list canonical %q", ct, listCan)
	}
}

func TestCanonicalTokensDeduplication(t *testing.T) {
	// If two tokens collapse to the same canonical, the result must deduplicate.
	// "sizeLen" (hypothetical): both "size" and "len" → same canonical → one entry.
	ct := CanonicalTokens("sizeLen")
	seen := make(map[string]int)
	for _, tok := range ct {
		seen[tok]++
	}
	for tok, cnt := range seen {
		if cnt > 1 {
			t.Errorf("duplicate canonical %q (×%d) in CanonicalTokens(sizeLen)", tok, cnt)
		}
	}
	// Result must have exactly one entry (both "size" and "len" share one canonical).
	if len(ct) != 1 {
		t.Errorf("CanonicalTokens(sizeLen)=%v, want exactly 1 entry (deduped)", ct)
	}
}

// ── 4. TokenOverlap semantic sharing ─────────────────────────────────────────

func testOverlapPositive(t *testing.T, a, b string) {
	t.Helper()
	ov := TokenOverlap(CanonicalTokens(a), CanonicalTokens(b))
	if ov <= 0 {
		t.Errorf("TokenOverlap(%q, %q)=%v; want > 0", a, b, ov)
	}
}

func testOverlapZero(t *testing.T, a, b string) {
	t.Helper()
	ov := TokenOverlap(CanonicalTokens(a), CanonicalTokens(b))
	if ov > 0 {
		t.Errorf("TokenOverlap(%q, %q)=%v; want 0", a, b, ov)
	}
}

// — spec examples —
func TestOverlapSpec_SizeVsTextLen(t *testing.T)      { testOverlapPositive(t, "size", "textLen") }
func TestOverlapSpec_FlipVsReverseList(t *testing.T)  { testOverlapPositive(t, "flip", "reverseList") }
func TestOverlapSpec_SizeVsIndexOf(t *testing.T)      { testOverlapZero(t, "size", "indexOf") }

// — length cluster —
func TestOverlap_CountVsListLen(t *testing.T)   { testOverlapPositive(t, "count", "listLen") }
func TestOverlap_LengthVsDictLen(t *testing.T)  { testOverlapPositive(t, "length", "dictLen") }

// — trim cluster —
func TestOverlap_StripVsTrim(t *testing.T) { testOverlapPositive(t, "strip", "trim") }

// — uppercase cluster —
func TestOverlap_UpperVsUppercase(t *testing.T)    { testOverlapPositive(t, "upper", "uppercase") }
func TestOverlap_ToupperVsUppercase(t *testing.T)  { testOverlapPositive(t, "toupper", "uppercase") }
func TestOverlap_UpcaseVsUppercase(t *testing.T)   { testOverlapPositive(t, "upcase", "uppercase") }

// — lowercase cluster —
func TestOverlap_LowerVsLowercase(t *testing.T)    { testOverlapPositive(t, "lower", "lowercase") }
func TestOverlap_DowncaseVsLowercase(t *testing.T) { testOverlapPositive(t, "downcase", "lowercase") }

// — contains cluster —
func TestOverlap_HasVsContains(t *testing.T)      { testOverlapPositive(t, "has", "contains") }
func TestOverlap_IncludesVsContains(t *testing.T) { testOverlapPositive(t, "includes", "contains") }
func TestOverlap_HasVsContainsAny(t *testing.T)   { testOverlapPositive(t, "has", "containsAny") }
func TestOverlap_HasKeyVsContainsKey(t *testing.T){ testOverlapPositive(t, "hasKey", "containsKey") }

// — split cluster —
func TestOverlap_ExplodeVsSplit(t *testing.T)    { testOverlapPositive(t, "explode", "split") }
func TestOverlap_DivideVsSplit(t *testing.T)     { testOverlapPositive(t, "divide", "split") }
func TestOverlap_TokenizeVsSplit(t *testing.T)   { testOverlapPositive(t, "tokenize", "split") }

// — reverse / list —
func TestOverlap_InvertVsReverse(t *testing.T)       { testOverlapPositive(t, "invert", "reverse") }
func TestOverlap_BackwardVsReverseList(t *testing.T) { testOverlapPositive(t, "backward", "reverseList") }

// — segment / substring —
func TestOverlap_SubstringVsSegment(t *testing.T) { testOverlapPositive(t, "substring", "segment") }
func TestOverlap_SubstrVsSegment(t *testing.T)    { testOverlapPositive(t, "substr", "segment") }
func TestOverlap_ExtractVsSegment(t *testing.T)   { testOverlapPositive(t, "extract", "segment") }

// — slice (list) —
func TestOverlap_CutVsSlice(t *testing.T)   { testOverlapPositive(t, "cut", "slice") }
func TestOverlap_RangeVsSlice(t *testing.T) { testOverlapPositive(t, "range", "slice") }

// — replace —
func TestOverlap_SubstituteVsReplace(t *testing.T) { testOverlapPositive(t, "substitute", "replace") }
func TestOverlap_SwapVsReplace(t *testing.T)       { testOverlapPositive(t, "swap", "replace") }

// — add / push / append —
func TestOverlap_PushVsAdd(t *testing.T)   { testOverlapPositive(t, "push", "add") }
func TestOverlap_AppendVsAdd(t *testing.T) { testOverlapPositive(t, "append", "add") }
func TestOverlap_PushVsAppendList(t *testing.T) { testOverlapPositive(t, "push", "appendList") }

// — insert —
func TestOverlap_PrependVsInsert(t *testing.T) { testOverlapPositive(t, "prepend", "insert") }
func TestOverlap_UnshiftVsInsert(t *testing.T) { testOverlapPositive(t, "unshift", "insert") }

// — indexOf —
func TestOverlap_FindVsIndexOf(t *testing.T)     { testOverlapPositive(t, "find", "indexOf") }
func TestOverlap_SearchVsIndexOf(t *testing.T)   { testOverlapPositive(t, "search", "indexOf") }
func TestOverlap_PositionVsIndexOf(t *testing.T) { testOverlapPositive(t, "position", "indexOf") }

// — remove / delete —
func TestOverlap_EraseVsRemove(t *testing.T) { testOverlapPositive(t, "erase", "remove") }
func TestOverlap_PopVsRemove(t *testing.T)   { testOverlapPositive(t, "pop", "remove") }
func TestOverlap_DelVsDelete(t *testing.T)   { testOverlapPositive(t, "del", "delete") }

// — join —
func TestOverlap_ConcatVsJoin(t *testing.T)       { testOverlapPositive(t, "concat", "join") }
func TestOverlap_ImplodeVsJoin(t *testing.T)      { testOverlapPositive(t, "implode", "join") }
func TestOverlap_GlueVsJoin(t *testing.T)         { testOverlapPositive(t, "glue", "join") }
func TestOverlap_ConcatenateVsJoin(t *testing.T)  { testOverlapPositive(t, "concatenate", "join") }

// — random —
func TestOverlap_PickVsRandom(t *testing.T)   { testOverlapPositive(t, "pick", "random") }
func TestOverlap_SampleVsRandom(t *testing.T) { testOverlapPositive(t, "sample", "random") }
func TestOverlap_RandVsRandom(t *testing.T)   { testOverlapPositive(t, "rand", "random") }
func TestOverlap_ChooseVsRandom(t *testing.T) { testOverlapPositive(t, "choose", "random") }

// — sort —
func TestOverlap_OrderVsSort(t *testing.T)   { testOverlapPositive(t, "order", "sort") }
func TestOverlap_ArrangeVsSort(t *testing.T) { testOverlapPositive(t, "arrange", "sort") }
func TestOverlap_RankVsSort(t *testing.T)    { testOverlapPositive(t, "rank", "sort") }


// — get / fetch / retrieve / lookup —
func TestOverlap_FetchVsGet(t *testing.T)    { testOverlapPositive(t, "fetch", "get") }
func TestOverlap_RetrieveVsGet(t *testing.T) { testOverlapPositive(t, "retrieve", "get") }
func TestOverlap_LookupVsGet(t *testing.T)   { testOverlapPositive(t, "lookup", "get") }
func TestOverlap_ReadVsGet(t *testing.T)     { testOverlapPositive(t, "read", "get") }

// — set / put / store —
func TestOverlap_PutVsSet(t *testing.T)    { testOverlapPositive(t, "put", "set") }
func TestOverlap_StoreVsSet(t *testing.T)  { testOverlapPositive(t, "store", "set") }
func TestOverlap_AssignVsSet(t *testing.T) { testOverlapPositive(t, "assign", "set") }
func TestOverlap_WriteVsSet(t *testing.T)  { testOverlapPositive(t, "write", "set") }

// — mergeInto —
func TestOverlap_CombineVsMerge(t *testing.T) { testOverlapPositive(t, "combine", "mergeInto") }
func TestOverlap_ExtendVsMerge(t *testing.T)  { testOverlapPositive(t, "extend", "mergeInto") }
func TestOverlap_UnionVsMerge(t *testing.T)   { testOverlapPositive(t, "union", "mergeInto") }

// — walkTree —
func TestOverlap_TraverseVsWalkTree(t *testing.T) { testOverlapPositive(t, "traverse", "walkTree") }
func TestOverlap_IterateVsWalkTree(t *testing.T)  { testOverlapPositive(t, "iterate", "walkTree") }

// — keys / values —
func TestOverlap_KeysetVsKeys(t *testing.T) { testOverlapPositive(t, "keyset", "keys") }
func TestOverlap_ValsVsValues(t *testing.T) { testOverlapPositive(t, "vals", "values") }

// — toPairs / lookupInPairs / pairsToDict —
func TestOverlap_EntriesToPairs(t *testing.T)  { testOverlapPositive(t, "entries", "toPairs") }
func TestOverlap_KvpairsToPairs(t *testing.T)  { testOverlapPositive(t, "kvpairs", "lookupInPairs") }

// — csv / row / table —
func TestOverlap_CommaVsCsvRow(t *testing.T)     { testOverlapPositive(t, "comma", "csvRowToList") }
func TestOverlap_LineVsCsvRow(t *testing.T)      { testOverlapPositive(t, "line", "toCsvRow") }
func TestOverlap_RecordVsCsvRow(t *testing.T)    { testOverlapPositive(t, "record", "csvRowToList") }
func TestOverlap_MatrixVsCsvTable(t *testing.T)  { testOverlapPositive(t, "matrix", "csvTableToList") }
func TestOverlap_GridVsCsvTable(t *testing.T)    { testOverlapPositive(t, "grid", "toCsvTable") }

// — key / prop / field (containsKey, getAtPath) —
func TestOverlap_PropVsContainsKey(t *testing.T)  { testOverlapPositive(t, "prop", "containsKey") }
func TestOverlap_FieldVsContainsKey(t *testing.T) { testOverlapPositive(t, "field", "containsKey") }
func TestOverlap_AttrVsContainsKey(t *testing.T)  { testOverlapPositive(t, "attr", "containsKey") }

// — all / any clusters —
func TestOverlap_EveryVsContainsAll(t *testing.T) { testOverlapPositive(t, "every", "containsAll") }
func TestOverlap_EachVsContainsAll(t *testing.T)  { testOverlapPositive(t, "each", "containsAll") }
func TestOverlap_SomeVsContainsAny(t *testing.T)  { testOverlapPositive(t, "some", "containsAny") }
func TestOverlap_EitherVsContainsAny(t *testing.T){ testOverlapPositive(t, "either", "containsAny") }

// — longest / greedy —
func TestOverlap_GreedyVsReplaceFromLongestFirst(t *testing.T) {
	testOverlapPositive(t, "greedy", "replaceFromLongestFirst")
}

// — type-module tokens (text/list/dict) —
func TestOverlap_StrVsTextLen(t *testing.T)    { testOverlapPositive(t, "str", "textLen") }
func TestOverlap_ArrayVsListLen(t *testing.T)  { testOverlapPositive(t, "array", "listLen") }
func TestOverlap_MapVsDictLen(t *testing.T)    { testOverlapPositive(t, "map", "dictLen") }
func TestOverlap_HashVsDictLen(t *testing.T)   { testOverlapPositive(t, "hash", "dictLen") }

// — spaces / whitespace —
func TestOverlap_WhitespaceVsSplitAtSpaces(t *testing.T) {
	testOverlapPositive(t, "whitespace", "splitAtSpaces")
}
func TestOverlap_WsVsSplitAtSpaces(t *testing.T) { testOverlapPositive(t, "ws", "splitAtSpaces") }

// ── 5. Non-interference: semantically unrelated pairs score zero ──────────────

func TestNoInterference_SizeVsSort(t *testing.T)     { testOverlapZero(t, "size", "sort") }
func TestNoInterference_FlipVsJoin(t *testing.T)     { testOverlapZero(t, "flip", "join") }
func TestNoInterference_UpperVsReverseList(t *testing.T) { testOverlapZero(t, "upper", "reverseList") }
func TestNoInterference_TrimVsIndexOf(t *testing.T)  { testOverlapZero(t, "trim", "indexOf") }
func TestNoInterference_RandomVsKeys(t *testing.T)   { testOverlapZero(t, "random", "keys") }
func TestNoInterference_SizeVsWalkTree(t *testing.T) { testOverlapZero(t, "size", "walkTree") }

// ── 6. Score end-to-end (semantic boost visible in combined score) ────────────

func assertScoreAbove(t *testing.T, input, candidate string, threshold float64) {
	t.Helper()
	s := Score(input, candidate)
	if s < threshold {
		t.Errorf("Score(%q, %q)=%.3f; want >= %.3f", input, candidate, s, threshold)
	}
}

func assertScoreBeats(t *testing.T, input, better, worse string) {
	t.Helper()
	sb := Score(input, better)
	sw := Score(input, worse)
	if sb <= sw {
		t.Errorf("Score(%q, %q)=%.3f should beat Score(%q, %q)=%.3f",
			input, better, sb, input, worse, sw)
	}
}

// spec examples
func TestScore_SizeVsTextLen(t *testing.T)     { assertScoreAbove(t, "size", "textLen", 0.1) }
func TestScore_FlipVsReverseList(t *testing.T) { assertScoreAbove(t, "flip", "reverseList", 0.1) }
func TestScore_SizeTextLenBeatsIndexOf(t *testing.T) {
	assertScoreBeats(t, "size", "textLen", "indexOf")
}
func TestScore_FlipReverseListBeatsAllButFirst(t *testing.T) {
	assertScoreBeats(t, "flip", "reverseList", "allButFirst")
}

// per-method alias smoke tests
func TestScore_StripVsTrim(t *testing.T)          { assertScoreAbove(t, "strip", "trim", 0.1) }
func TestScore_UpperVsUppercase(t *testing.T)     { assertScoreAbove(t, "upper", "uppercase", 0.1) }
func TestScore_LowerVsLowercase(t *testing.T)     { assertScoreAbove(t, "lower", "lowercase", 0.1) }
func TestScore_HasVsContains(t *testing.T)        { assertScoreAbove(t, "has", "contains", 0.1) }
func TestScore_ExplodeVsSplit(t *testing.T)       { assertScoreAbove(t, "explode", "split", 0.1) }
func TestScore_SubstringVsSegment(t *testing.T)   { assertScoreAbove(t, "substring", "segment", 0.1) }
func TestScore_SubstituteVsReplace(t *testing.T)  { assertScoreAbove(t, "substitute", "replace", 0.1) }
func TestScore_PushVsAdd(t *testing.T)            { assertScoreAbove(t, "push", "add", 0.1) }
func TestScore_FindVsIndexOf(t *testing.T)        { assertScoreAbove(t, "find", "indexOf", 0.1) }
func TestScore_EraseVsRemove(t *testing.T)        { assertScoreAbove(t, "erase", "remove", 0.1) }
func TestScore_ConcatVsJoin(t *testing.T)         { assertScoreAbove(t, "concat", "join", 0.1) }
func TestScore_PickVsRandom(t *testing.T)         { assertScoreAbove(t, "pick", "random", 0.1) }
func TestScore_OrderVsSort(t *testing.T)          { assertScoreAbove(t, "order", "sort", 0.1) }
func TestScore_FetchVsGet(t *testing.T)           { assertScoreAbove(t, "fetch", "get", 0.1) }
func TestScore_PutVsSet(t *testing.T)             { assertScoreAbove(t, "put", "set", 0.1) }
func TestScore_CombineVsMergeInto(t *testing.T)   { assertScoreAbove(t, "combine", "mergeInto", 0.1) }
func TestScore_TraverseVsWalkTree(t *testing.T)   { assertScoreAbove(t, "traverse", "walkTree", 0.1) }
func TestScore_CountVsListLen(t *testing.T)       { assertScoreAbove(t, "count", "listLen", 0.1) }
func TestScore_HasKeyVsContainsKey(t *testing.T)  { assertScoreAbove(t, "hasKey", "containsKey", 0.1) }
func TestScore_EntriesVsToPairs(t *testing.T)     { assertScoreAbove(t, "entries", "toPairs", 0.1) }
func TestScore_LineVsToCsvRow(t *testing.T)       { assertScoreAbove(t, "line", "toCsvRow", 0.1) }
func TestScore_GridVsToCsvTable(t *testing.T)     { assertScoreAbove(t, "grid", "toCsvTable", 0.1) }
func TestScore_PrependVsInsert(t *testing.T)      { assertScoreAbove(t, "prepend", "insert", 0.1) }
func TestScore_CutVsSlice(t *testing.T)           { assertScoreAbove(t, "cut", "slice", 0.1) }
func TestScore_SearchVsIndexOf(t *testing.T)      { assertScoreAbove(t, "search", "indexOf", 0.1) }
