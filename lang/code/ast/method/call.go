package method

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
	"Falcon/code/sugar"
	"strconv"
)

type Call struct {
	Where *lex.Token
	On    ast.Expr
	Name  string
	Args  []ast.Expr
}

type CallSignature struct {
	Module      string
	BlocklyName string
	ParamCount  int
	Params      string // human-readable parameter list, e.g. "key, notFound"
	Consumable  bool
	Signature   ast.Signature
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

var signatures = map[string]*CallSignature{
	"textLen":                 makeSignature("text", "text_length", 0, "", true, ast.SignNumb),
	"trim":                    makeSignature("text", "text_trim", 0, "", true, ast.SignText),
	"uppercase":               makeSignature("text", "text_changeCase", 0, "", true, ast.SignText),
	"lowercase":               makeSignature("text", "text_changeCase", 0, "", true, ast.SignText),
	"startsWith":              makeSignature("text", "text_starts_at", 1, "prefix", true, ast.SignBool),
	"contains":                makeSignature("text", "text_contains", 1, "piece", true, ast.SignBool),
	"containsAny":             makeSignature("text", "text_contains", 1, "pieces", true, ast.SignBool),
	"containsAll":             makeSignature("text", "text_contains", 1, "pieces", true, ast.SignBool),
	"split":                   makeSignature("text", "text_split", 1, "separator", true, ast.SignList),
	"splitAtFirst":            makeSignature("text", "text_split", 1, "separator", true, ast.SignList),
	"splitAtAny":              makeSignature("text", "text_split", 1, "separators", true, ast.SignList),
	"splitAtFirstOfAny":       makeSignature("text", "text_split", 1, "separators", true, ast.SignList),
	"splitAtSpaces":           makeSignature("text", "text_split_at_spaces", 0, "", true, ast.SignList),
	"reverse":                 makeSignature("text", "text_reverse", 0, "", true, ast.SignText),
	"csvRowToList":            makeSignature("text", "lists_from_csv_row", 0, "", true, ast.SignList),
	"csvTableToList":          makeSignature("text", "lists_from_csv_table", 0, "", true, ast.SignList),
	"segment":                 makeSignature("text", "text_segment", 2, "start, length", true, ast.SignText),
	"replace":                 makeSignature("text", "text_replace_all", 2, "from, to", true, ast.SignText),
	"replaceFrom":             makeSignature("text", "text_replace_mappings", 1, "mappingDict", true, ast.SignText),
	"replaceFromLongestFirst": makeSignature("text", "text_replace_mappings", 1, "mappingDict", true, ast.SignText),

	"listLen":       makeSignature("list", "lists_length", 0, "", true, ast.SignNumb),
	"add":           makeSignature("list", "lists_add_items", -1, "item...", false, ast.SignVoid),
	"containsItem":  makeSignature("list", "lists_is_in", 1, "item", true, ast.SignBool),
	"indexOf":       makeSignature("list", "lists_position_in", 1, "item", true, ast.SignNumb),
	"insert":        makeSignature("list", "lists_insert_item", 2, "index, item", false, ast.SignVoid),
	"remove":        makeSignature("list", "lists_remove_item", 1, "index", false, ast.SignVoid),
	"appendList":    makeSignature("list", "lists_append_list", 1, "other", false, ast.SignVoid),
	"lookupInPairs": makeSignature("list", "lists_lookup_in_pairs", 2, "key, notFound", true, ast.SignAny),
	"join":          makeSignature("list", "lists_join_with_separator", 1, "separator", true, ast.SignText),
	"slice":         makeSignature("list", "lists_slice", 2, "from, to", true, ast.SignList),
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
	"getAtPath":   makeSignature("dict", "dictionaries_recursive_lookup", 2, "keys, notFound", true, ast.SignAny),
	"setAtPath":   makeSignature("dict", "dictionaries_recursive_set", 2, "keys, value", false, ast.SignVoid),
	"containsKey": makeSignature("dict", "dictionaries_is_key_in", 1, "key", true, ast.SignBool),
	"mergeInto":   makeSignature("dict", "dictionaries_combine_dicts", 1, "other", true, ast.SignDict),
	"walkTree":    makeSignature("dict", "dictionaries_walk_tree", 1, "procedure", true, ast.SignAny),
	"keys":        makeSignature("dict", "dictionaries_getters", 0, "", true, ast.SignList),
	"values":      makeSignature("dict", "dictionaries_getters", 0, "", true, ast.SignList),
	"toPairs":     makeSignature("dict", "dictionaries_dict_to_alist", 0, "", true, ast.SignList),
}

// sigString returns a display string like ".get(key, notFound)" for error messages.
func sigString(name string, sig *CallSignature) string {
	return "." + name + "(" + sig.Params + ")"
}

func TestSignature(methodName string, argsCount int) (string, *CallSignature) {
	signature, ok := signatures[methodName]
	if !ok {
		return sugar.Format("Cannot find method .%()", methodName), nil
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
	errorMessage, signature := TestSignature(c.Name, len(c.Args))
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

func (c *Call) Consumable(flags ...bool) bool {
	signature, ok := signatures[c.Name]
	if !ok {
		c.Where.Error("Cannot find method .%()", c.Name)
	}
	return signature.Consumable
}

func (c *Call) Signature() []ast.Signature {
	errorMessage, signature := TestSignature(c.Name, len(c.Args))
	if signature == nil {
		c.Where.Error(errorMessage)
	}
	return []ast.Signature{signature.Signature}
}

func (c *Call) simpleOperand(blockType string, valueName string) ast.Block {
	return ast.Block{Type: blockType, Values: []ast.Value{{Name: valueName, Block: c.On.Blockly(false)}}}
}
