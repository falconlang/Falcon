package method

import (
	"Falcon/code/ast"
	"Falcon/code/fzf"
	"Falcon/code/lex"
	"Falcon/code/sugar"
	"strconv"
	"strings"
)

type Call struct {
	Where      *lex.Token
	On         ast.Expr
	Name       string
	Args       []ast.Expr
	hintOutput *ast.Signature // nearest valid successor's input type, set by CorrectChain
}

type CallSignature struct {
	Module      string
	BlocklyName string
	ParamCount  int
	Params      string // human-readable parameter list, e.g. "key, notFound"
	Consumable  bool
	Signature   ast.Signature
	ParamSigs   []ast.Signature // nil = accept any; non-nil = required type per positional param
}

func makeSignature(
	module string,
	blocklyName string,
	paramCount int,
	params string,
	consumable bool,
	signature ast.Signature,
) *CallSignature {
	return &CallSignature{
		Module:      module,
		BlocklyName: blocklyName,
		ParamCount:  paramCount,
		Params:      params,
		Consumable:  consumable,
		Signature:   signature,
	}
}

func makeSignatureTyped(
	module string,
	blocklyName string,
	paramCount int,
	params string,
	consumable bool,
	signature ast.Signature,
	paramSigs []ast.Signature,
) *CallSignature {
	cs := makeSignature(module, blocklyName, paramCount, params, consumable, signature)
	cs.ParamSigs = paramSigs
	return cs
}

var signatures = map[string]*CallSignature{
	"textLen":                 makeSignature("text", "text_length", 0, "", true, ast.SignNumb),
	"trim":                    makeSignature("text", "text_trim", 0, "", true, ast.SignText),
	"uppercase":               makeSignature("text", "text_changeCase", 0, "", true, ast.SignText),
	"lowercase":               makeSignature("text", "text_changeCase", 0, "", true, ast.SignText),
	"startsWith":              makeSignatureTyped("text", "text_starts_at", 1, "prefix", true, ast.SignBool, []ast.Signature{ast.SignText}),
	"contains":                makeSignatureTyped("text", "text_contains", 1, "piece", true, ast.SignBool, []ast.Signature{ast.SignText}),
	"containsAny":             makeSignatureTyped("text", "text_contains", 1, "pieces", true, ast.SignBool, []ast.Signature{ast.SignList}),
	"containsAll":             makeSignatureTyped("text", "text_contains", 1, "pieces", true, ast.SignBool, []ast.Signature{ast.SignList}),
	"split":                   makeSignatureTyped("text", "text_split", 1, "separator", true, ast.SignList, []ast.Signature{ast.SignText}),
	"splitAtFirst":            makeSignatureTyped("text", "text_split", 1, "separator", true, ast.SignList, []ast.Signature{ast.SignText}),
	"splitAtAny":              makeSignature("text", "text_split", 1, "separators", true, ast.SignList),
	"splitAtFirstOfAny":       makeSignature("text", "text_split", 1, "separators", true, ast.SignList),
	"splitAtSpaces":           makeSignature("text", "text_split_at_spaces", 0, "", true, ast.SignList),
	"reverse":                 makeSignature("text", "text_reverse", 0, "", true, ast.SignText),
	"csvRowToList":            makeSignature("text", "lists_from_csv_row", 0, "", true, ast.SignList),
	"csvTableToList":          makeSignature("text", "lists_from_csv_table", 0, "", true, ast.SignList),
	"segment":                 makeSignatureTyped("text", "text_segment", 2, "start, length", true, ast.SignText, []ast.Signature{ast.SignNumb, ast.SignNumb}),
	"replace":                 makeSignatureTyped("text", "text_replace_all", 2, "from, to", true, ast.SignText, []ast.Signature{ast.SignText, ast.SignText}),
	"replaceFrom":             makeSignatureTyped("text", "text_replace_mappings", 1, "mappingDict", true, ast.SignText, []ast.Signature{ast.SignDict}),
	"replaceFromLongestFirst": makeSignatureTyped("text", "text_replace_mappings", 1, "mappingDict", true, ast.SignText, []ast.Signature{ast.SignDict}),

	"listLen":       makeSignature("list", "lists_length", 0, "", true, ast.SignNumb),
	"add":           makeSignature("list", "lists_add_items", -1, "item...", false, ast.SignVoid),
	"containsItem":  makeSignature("list", "lists_is_in", 1, "item", true, ast.SignBool),
	"indexOf":       makeSignature("list", "lists_position_in", 1, "item", true, ast.SignNumb),
	"insert":        makeSignatureTyped("list", "lists_insert_item", 2, "index, item", false, ast.SignVoid, []ast.Signature{ast.SignNumb, ast.SignAny}),
	"remove":        makeSignatureTyped("list", "lists_remove_item", 1, "index", false, ast.SignVoid, []ast.Signature{ast.SignNumb}),
	"appendList":    makeSignatureTyped("list", "lists_append_list", 1, "other", false, ast.SignVoid, []ast.Signature{ast.SignList}),
	"lookupInPairs": makeSignature("list", "lists_lookup_in_pairs", 2, "key, notFound", true, ast.SignAny),
	"join":          makeSignatureTyped("list", "lists_join_with_separator", 1, "separator", true, ast.SignText, []ast.Signature{ast.SignText}),
	"slice":         makeSignatureTyped("list", "lists_slice", 2, "from, to", true, ast.SignList, []ast.Signature{ast.SignNumb, ast.SignNumb}),
	"random":        makeSignature("list", "lists_pick_random_item", 0, "", true, ast.SignAny),
	"reverseList":   makeSignature("list", "lists_reverse", 0, "", true, ast.SignList),
	"toCsvRow":      makeSignature("list", "lists_to_csv_row", 0, "", true, ast.SignText),
	"toCsvTable":    makeSignature("list", "lists_to_csv_table", 0, "", true, ast.SignText),
	"sort":          makeSignature("list", "lists_sort", 0, "", true, ast.SignList),
	"allButFirst":   makeSignature("list", "lists_but_first", 0, "", true, ast.SignList),
	"allButLast":    makeSignature("list", "lists_but_last", 0, "", true, ast.SignList),
	"pairsToDict":   makeSignature("list", "dictionaries_alist_to_dict", 0, "", true, ast.SignDict),

	"dictLen":     makeSignature("dict", "dictionaries_length", 0, "", true, ast.SignNumb),
	"get":         makeSignature("dict", "dictionaries_lookup", 2, "key, notFound", true, ast.SignAny),
	"set":         makeSignature("dict", "dictionaries_set_pair", 2, "key, value", false, ast.SignVoid),
	"delete":      makeSignature("dict", "dictionaries_delete_pair", 1, "key", false, ast.SignVoid),
	"getAtPath":   makeSignatureTyped("dict", "dictionaries_recursive_lookup", 2, "keys, notFound", true, ast.SignAny, []ast.Signature{ast.SignList, ast.SignAny}),
	"setAtPath":   makeSignatureTyped("dict", "dictionaries_recursive_set", 2, "keys, value", false, ast.SignVoid, []ast.Signature{ast.SignList, ast.SignAny}),
	"containsKey": makeSignature("dict", "dictionaries_is_key_in", 1, "key", true, ast.SignBool),
	"mergeInto":   makeSignatureTyped("dict", "dictionaries_combine_dicts", 1, "other", true, ast.SignDict, []ast.Signature{ast.SignDict}),
	"walkTree":    makeSignature("dict", "dictionaries_walk_tree", 1, "procedure", true, ast.SignAny),
	"keys":        makeSignature("dict", "dictionaries_getters", 0, "", true, ast.SignList),
	"values":      makeSignature("dict", "dictionaries_getters", 0, "", true, ast.SignList),
	"toPairs":     makeSignature("dict", "dictionaries_dict_to_alist", 0, "", true, ast.SignList),
}

// sigString returns a display string like ".get(key, notFound)" for error messages.
func sigString(name string, sig *CallSignature) string {
	return "." + name + "(" + sig.Params + ")"
}

// HintOutput returns the lookahead output constraint stored by CorrectChain,
// or nil if no valid successor was found in the chain.
func (c *Call) HintOutput() *ast.Signature { return c.hintOutput }

// DeriveAllowedModules is exported for use in mistparser's checkPendingSymbols.
func DeriveAllowedModules(onSigs []ast.Signature) []string {
	return deriveAllowedModules(onSigs)
}

func deriveAllowedModules(onSigs []ast.Signature) []string {
	var modules []string
	if ast.HasSignature(onSigs, ast.SignText) {
		modules = append(modules, "text")
	}
	if ast.HasSignature(onSigs, ast.SignList) {
		modules = append(modules, "list")
	}
	if ast.HasSignature(onSigs, ast.SignDict) {
		modules = append(modules, "dict")
	}
	return modules
}

func joinOr(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		return parts[0] + " or " + parts[1]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + " or " + parts[len(parts)-1]
}

// methodSuggestions returns a "Did you mean" suffix filtered by input module
// and optionally by neededOutput (the output type the caller expects).
// When neededOutput filtering leaves no candidates it falls back without it.
func methodSuggestions(methodName string, allowedModules []string, neededOutput *ast.Signature) string {
	candidates := collectCandidates(methodName, allowedModules, neededOutput)
	if len(candidates) == 0 && neededOutput != nil {
		candidates = collectCandidates(methodName, allowedModules, nil)
	}
	suggestions := fzf.Top(methodName, candidates, 3)
	if len(suggestions) > 0 {
		parts := make([]string, len(suggestions))
		for i, s := range suggestions {
			parts[i] = "." + s + "()"
		}
		return ". Did you mean " + joinOr(parts) + "?"
	}
	return ""
}

func collectCandidates(methodName string, allowedModules []string, neededOutput *ast.Signature) []string {
	candidates := make([]string, 0, len(signatures))
	for name, sig := range signatures {
		if name == methodName {
			continue
		}
		if len(allowedModules) > 0 {
			moduleMatch := false
			for _, mod := range allowedModules {
				if sig.Module == mod {
					moduleMatch = true
					break
				}
			}
			if !moduleMatch {
				continue
			}
		}
		if neededOutput != nil && sig.Signature != *neededOutput && sig.Signature != ast.SignAny {
			continue
		}
		candidates = append(candidates, name)
	}
	return candidates
}

// BuildSuggestions is exported for use in mistparser's checkPendingSymbols.
func BuildSuggestions(methodName string, allowedModules []string, neededOutput *ast.Signature) string {
	return methodSuggestions(methodName, allowedModules, neededOutput)
}

// FindBestSuggestion returns the single highest-scoring replacement name, or ""
// if no candidate clears the scoring threshold.
func FindBestSuggestion(methodName string, allowedModules []string, neededOutput *ast.Signature) string {
	candidates := collectCandidates(methodName, allowedModules, neededOutput)
	if len(candidates) == 0 && neededOutput != nil {
		candidates = collectCandidates(methodName, allowedModules, nil)
	}
	if tops := fzf.Top(methodName, candidates, 1); len(tops) > 0 {
		return tops[0]
	}
	return ""
}

func TestSignature(methodName string, argsCount int, allowedModules ...string) (string, *CallSignature) {
	signature, ok := signatures[methodName]
	if !ok {
		return "No method named ." + methodName + "()" + methodSuggestions(methodName, allowedModules, nil), nil
	}
	sig := sigString(methodName, signature)
	if signature.ParamCount >= 0 {
		if signature.ParamCount != argsCount {
			return sugar.Format("% expects % arg(s) but got %",
				sig, strconv.Itoa(signature.ParamCount), strconv.Itoa(argsCount)), nil
		}
	} else {
		minArgs := -signature.ParamCount
		if argsCount < minArgs {
			return sugar.Format("% expects at least % arg(s) but got %",
				sig, strconv.Itoa(minArgs), strconv.Itoa(argsCount)), nil
		}
	}
	return "", signature
}

func (c *Call) String() string {
	pFormat := "%.%(%)"
	if !c.On.Continuous() {
		pFormat = "(%).%(%)"
	}
	return sugar.Format(pFormat, c.On.String(), c.Name, ast.JoinExprs(", ", c.Args))
}

func (c *Call) Blockly(flags ...bool) ast.Block {
	onSigs := c.On.Signature()
	errorMessage, signature := TestSignature(c.Name, len(c.Args), deriveAllowedModules(onSigs)...)
	if signature == nil {
		c.Where.Error(errorMessage)
	}
	switch signature.Module {
	case "text":
		return c.textMethods(signature)
	case "list":
		return c.listMethods(signature)
	case "dict":
		return c.dictMethods(signature)
	default:
		c.Where.Error("Unknown method module: %", signature.Module)
		panic("")
	}
}

func (c *Call) Continuous() bool {
	return true
}

func (c *Call) Consumable() bool {
	signature, ok := signatures[c.Name]
	if !ok {
		return true
	}
	return signature.Consumable
}

func (c *Call) Signature() []ast.Signature {
	onSigs := c.On.Signature()
	errorMessage, signature := TestSignature(c.Name, len(c.Args), deriveAllowedModules(onSigs)...)
	if signature == nil {
		c.Where.Error(errorMessage)
	}
	for i, arg := range c.Args {
		argSigs := arg.Signature()
		if i < len(signature.ParamSigs) {
			expected := signature.ParamSigs[i]
			if expected != ast.SignAny && !ast.HasSignature(argSigs, expected) {
				c.Where.TypeError(".%() argument % expects %, not %", c.Name, strconv.Itoa(i+1), expected.String(), ast.FormatSignatures(argSigs))
			}
		}
	}
	intendedOutput := signature.Signature
	switch signature.Module {
	case "text":
		if !ast.HasSignature(onSigs, ast.SignText) {
			c.Where.TypeError(".%() operates on text, not %"+methodSuggestions(c.Name, deriveAllowedModules(onSigs), &intendedOutput), c.Name, ast.FormatSignatures(onSigs))
		}
	case "list":
		if !ast.HasSignature(onSigs, ast.SignList) {
			c.Where.TypeError(".%() operates on lists, not %"+methodSuggestions(c.Name, deriveAllowedModules(onSigs), &intendedOutput), c.Name, ast.FormatSignatures(onSigs))
		}
	case "dict":
		if !ast.HasSignature(onSigs, ast.SignDict) {
			c.Where.TypeError(".%() operates on dictionaries, not %"+methodSuggestions(c.Name, deriveAllowedModules(onSigs), &intendedOutput), c.Name, ast.FormatSignatures(onSigs))
		}
	}
	return []ast.Signature{signature.Signature}
}

func (c *Call) simpleOperand(blockType string, valueName string) ast.Block {
	return ast.Block{Type: blockType, Values: []ast.Value{{Name: valueName, Block: c.On.Blockly(false)}}}
}
