package fzf

import "strings"

// synonymGraph maps a token to its direct synonym neighbours.
// The graph is intentionally sparse: clusters are connected through hub nodes
// (e.g. "length" links "len", "size", and "count") so that transitivity
// (size→length→len) is resolved by BFS in Canonical, never listed manually.
var synonymGraph = map[string][]string{
	// ── length / size / count ────────────────────────────────────────────────
	"length": {"len", "size", "count"},
	"len":    {"length"},
	"size":   {"length"},
	"count":  {"length"},

	// ── trim / strip ─────────────────────────────────────────────────────────
	"trim":  {"strip", "clean"},
	"strip": {"trim", "clean"},
	"clean": {"trim", "strip"},

	// ── uppercase ─────────────────────────────────────────────────────────────
	"uppercase": {"upper", "upcase", "toupper"},
	"upper":     {"uppercase", "upcase", "toupper"},
	"upcase":    {"upper", "uppercase"},
	"toupper":   {"upper", "uppercase"},

	// ── lowercase ─────────────────────────────────────────────────────────────
	"lowercase": {"lower", "downcase", "tolower"},
	"lower":     {"lowercase", "downcase", "tolower"},
	"downcase":  {"lower", "lowercase"},
	"tolower":   {"lower", "lowercase"},

	// ── starts / begins / prefix  (for startsWith) ───────────────────────────
	"starts": {"begins", "prefix"},
	"begins": {"starts", "prefix"},
	"prefix": {"starts", "begins"},

	// ── contains / includes / has ─────────────────────────────────────────────
	"contains": {"includes", "has", "include"},
	"includes": {"contains", "has", "include"},
	"has":      {"contains", "includes", "include"},
	"include":  {"contains", "includes", "has"},

	// ── split / explode / divide / partition / tokenize ───────────────────────
	"split":     {"explode", "divide", "partition", "tokenize"},
	"explode":   {"split", "divide"},
	"divide":    {"split", "explode", "partition"},
	"partition": {"split", "divide"},
	"tokenize":  {"split"},

	// ── reverse / flip / invert ───────────────────────────────────────────────
	"reverse":   {"flip", "invert", "backward"},
	"flip":      {"reverse", "invert"},
	"invert":    {"reverse", "flip"},
	"backward":  {"reverse"},
	"backwards": {"reverse"},

	// ── segment / substring / substr / extract  (text sub-range) ─────────────
	"segment":   {"substring", "substr", "extract"},
	"substring": {"segment", "substr", "extract"},
	"substr":    {"segment", "substring"},
	"extract":   {"segment", "substring"},

	// ── slice / cut / range  (list sub-range) ────────────────────────────────
	"slice": {"cut", "range"},
	"cut":   {"slice", "range"},
	"range": {"slice", "cut"},

	// ── replace / substitute / swap / subst ──────────────────────────────────
	"replace":    {"substitute", "swap", "subst"},
	"substitute": {"replace", "swap", "subst"},
	"swap":       {"replace", "substitute"},
	"subst":      {"replace", "substitute"},

	// ── add / push / append ───────────────────────────────────────────────────
	"add":    {"push", "append"},
	"push":   {"add", "append"},
	"append": {"add", "push"},

	// ── insert / prepend / unshift ────────────────────────────────────────────
	"insert":  {"prepend", "unshift"},
	"prepend": {"insert", "unshift"},
	"unshift": {"insert", "prepend"},

	// ── indexOf: index / find / position / search / locate / pos ─────────────
	"index":    {"find", "position", "search", "locate", "pos"},
	"find":     {"index", "position", "search", "locate"},
	"position": {"index", "find", "locate", "pos"},
	"search":   {"index", "find", "locate"},
	"locate":   {"index", "find", "position"},
	"pos":      {"index", "position"},

	// ── remove / delete / erase / pop / del ──────────────────────────────────
	"remove": {"delete", "erase", "pop", "del"},
	"delete": {"remove", "erase", "del"},
	"erase":  {"remove", "delete"},
	"pop":    {"remove"},
	"del":    {"remove", "delete"},

	// ── join / concat / concatenate / glue / implode ─────────────────────────
	"join":        {"concat", "concatenate", "glue", "implode", "connect"},
	"concat":      {"join", "concatenate", "glue"},
	"concatenate": {"join", "concat"},
	"glue":        {"join", "concat"},
	"implode":     {"join"},
	"connect":     {"join"},

	// ── random / pick / sample / rand / choose ────────────────────────────────
	"random": {"pick", "sample", "rand", "choose"},
	"pick":   {"random", "sample", "choose"},
	"sample": {"random", "pick", "choose"},
	"rand":   {"random"},
	"choose": {"random", "pick", "sample"},

	// ── sort / order / arrange / rank ─────────────────────────────────────────
	"sort":    {"order", "arrange", "rank"},
	"order":   {"sort", "arrange", "rank"},
	"arrange": {"sort", "order"},
	"rank":    {"sort", "order"},

	// ── first / head / front ──────────────────────────────────────────────────
	"first": {"head", "front"},
	"head":  {"first", "front"},
	"front": {"first", "head"},

	// ── last / tail / end / final ─────────────────────────────────────────────
	"last":  {"tail", "end", "final"},
	"tail":  {"last", "end"},
	"end":   {"last", "tail", "final"},
	"final": {"last", "end"},

	// ── but / except / without / excluding  (for allBut*) ────────────────────
	"but":       {"except", "without", "excluding"},
	"except":    {"but", "without", "excluding"},
	"without":   {"but", "except"},
	"excluding": {"but", "except"},

	// ── get / fetch / retrieve / lookup / read ────────────────────────────────
	"get":      {"fetch", "retrieve", "lookup", "read"},
	"fetch":    {"get", "retrieve", "lookup"},
	"retrieve": {"get", "fetch", "lookup"},
	"lookup":   {"get", "fetch", "retrieve"},
	"read":     {"get", "fetch"},

	// ── set / put / store / assign / write ────────────────────────────────────
	"set":    {"put", "store", "assign", "write"},
	"put":    {"set", "store", "assign"},
	"store":  {"set", "put"},
	"assign": {"set", "put"},
	"write":  {"set", "store"},

	// ── merge / combine / extend / union / mixin ──────────────────────────────
	"merge":   {"combine", "extend", "union", "mixin"},
	"combine": {"merge", "extend", "union"},
	"extend":  {"merge", "combine"},
	"union":   {"merge", "combine"},
	"mixin":   {"merge"},

	// ── walk / traverse / iterate / visit / scan ──────────────────────────────
	"walk":     {"traverse", "iterate", "visit", "scan"},
	"traverse": {"walk", "iterate", "visit"},
	"iterate":  {"walk", "traverse"},
	"visit":    {"walk", "traverse"},
	"scan":     {"walk"},

	// ── keys / keyset ────────────────────────────────────────────────────────
	"keys":   {"keyset", "keynames"},
	"keyset": {"keys"},

	// ── values / vals ────────────────────────────────────────────────────────
	"values": {"vals"},
	"vals":   {"values"},

	// ── pairs / entries / kvpairs  (for toPairs, lookupInPairs, pairsToDict) ──
	"pairs":   {"entries", "kvpairs"},
	"entries": {"pairs", "kvpairs"},
	"kvpairs": {"pairs", "entries"},

	// ── text / string / str ──────────────────────────────────────────────────
	"text":   {"string", "str"},
	"string": {"text", "str"},
	"str":    {"text", "string"},

	// ── list / array / arr / vec / vector ────────────────────────────────────
	"list":   {"array", "arr", "vec", "vector"},
	"array":  {"list", "arr", "vec"},
	"arr":    {"list", "array"},
	"vec":    {"list", "array", "vector"},
	"vector": {"list", "vec"},

	// ── dict / map / hash / hashmap / object / obj ───────────────────────────
	"dict":    {"map", "hash", "hashmap", "object", "obj"},
	"map":     {"dict", "hash", "hashmap"},
	"hash":    {"dict", "map", "hashmap"},
	"hashmap": {"dict", "map", "hash"},
	"object":  {"dict", "obj"},
	"obj":     {"dict", "object"},

	// ── spaces / whitespace / ws / blanks ────────────────────────────────────
	"spaces":     {"whitespace", "ws", "blanks"},
	"whitespace": {"spaces", "ws"},
	"ws":         {"spaces", "whitespace"},
	"blanks":     {"spaces"},

	// ── separator / delimiter / sep / delim ──────────────────────────────────
	"separator": {"delimiter", "sep", "delim"},
	"delimiter": {"separator", "sep", "delim"},
	"sep":       {"separator", "delimiter"},
	"delim":     {"separator", "delimiter"},

	// ── item / element / elem ────────────────────────────────────────────────
	"item":    {"element", "elem"},
	"element": {"item", "elem"},
	"elem":    {"item", "element"},

	// ── path / route / nested ────────────────────────────────────────────────
	"path":   {"route", "nested"},
	"route":  {"path"},
	"nested": {"path"},

	// ── tree / graph / structure ──────────────────────────────────────────────
	"tree":      {"graph", "structure"},
	"graph":     {"tree"},
	"structure": {"tree"},

	// ── all / every / each  (for containsAll, allButFirst, allButLast) ────────
	"all":   {"every", "each"},
	"every": {"all", "each"},
	"each":  {"all", "every"},

	// ── any / some / either  (for containsAny, splitAtAny) ───────────────────
	"any":    {"some", "either"},
	"some":   {"any", "either"},
	"either": {"any", "some"},

	// ── csv / comma / separated  (for csvRowToList, toCsvRow, etc.) ──────────
	"csv":       {"comma", "separated"},
	"comma":     {"csv", "separated"},
	"separated": {"csv", "comma"},

	// ── row / line / record  (for csvRowToList, toCsvRow) ─────────────────────
	"row":    {"line", "record"},
	"line":   {"row", "record"},
	"record": {"row", "line"},

	// ── table / matrix / grid  (for csvTableToList, toCsvTable) ──────────────
	"table":  {"matrix", "grid", "spreadsheet"},
	"matrix": {"table", "grid"},
	"grid":   {"table", "matrix"},
	"spreadsheet": {"table"},

	// ── key / prop / field / property / attr  (for containsKey, getAtPath) ───
	"key":       {"prop", "field", "property", "attr", "attribute"},
	"prop":      {"key", "field", "property", "attr"},
	"field":     {"key", "prop", "property", "attr"},
	"property":  {"key", "prop", "field", "attr", "attribute"},
	"attr":      {"key", "prop", "field", "property"},
	"attribute": {"key", "property"},

	// ── longest / greedy / maximal  (for replaceFromLongestFirst) ────────────
	"longest": {"greedy", "maximal"},
	"greedy":  {"longest", "maximal"},
	"maximal": {"longest", "greedy"},
}

// Canonical returns the alphabetically smallest token reachable from word
// in the synonym cluster via BFS. Words absent from the graph return themselves.
// This provides a stable canonical form so that semantically equivalent tokens
// (e.g. "size", "len", "count") all collapse to the same representative.
func Canonical(word string) string {
	word = strings.ToLower(word)
	visited := make(map[string]bool)
	queue := []string{word}
	visited[word] = true
	smallest := word

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr < smallest {
			smallest = curr
		}

		for _, neighbour := range synonymGraph[curr] {
			n := strings.ToLower(neighbour)
			if !visited[n] {
				visited[n] = true
				queue = append(queue, n)
			}
		}
	}

	return smallest
}

// CanonicalTokens splits identifier by camelCase boundaries (via SplitCamel),
// canonicalises each token, and deduplicates. This is the semantic-aware
// token set that TokenOverlap uses for scoring.
func CanonicalTokens(identifier string) []string {
	tokens := SplitCamel(identifier)
	seen := make(map[string]bool)
	result := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		c := Canonical(strings.ToLower(tok))
		if !seen[c] {
			seen[c] = true
			result = append(result, c)
		}
	}
	return result
}
