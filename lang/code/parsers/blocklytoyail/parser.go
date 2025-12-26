package blocklytoyail

import (
	"Falcon/code/ast"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

type ValueMap struct {
	valueMap map[string]string
}

func (v *ValueMap) getUnsafe(name string) string {
	return v.valueMap[name]
}

func (v *ValueMap) get(name string) string {
	value, ok := v.valueMap[name]
	if !ok {
		panic("Block not found: " + name)
	}
	return value
}

type StatementMap struct {
	statementMap map[string]string
}

func (s *StatementMap) get(name string) string {
	value, ok := s.statementMap[name]
	if !ok {
		return ""
	}
	return value
}

type Parser struct {
	xmlContent string
}

func NewParser(xmlContent string) *Parser {
	return &Parser{xmlContent: xmlContent}
}

func (p *Parser) GenerateYAIL() string {
	blocks := p.decodeXML()
	var code []string

	for _, block := range blocks {
		c := p.parseBlock(block)
		if c != "" {
			code = append(code, c)
		}
	}
	return strings.Join(code, "\n\n")
}

func (p *Parser) decodeXML() []ast.Block {
	decoder := xml.NewDecoder(strings.NewReader(p.xmlContent))
	decoder.Strict = false
	decoder.DefaultSpace = ""

	var root ast.XmlRoot
	if err := decoder.Decode(&root); err != nil {
		return []ast.Block{}
	}
	return root.Blocks
}

func (p *Parser) parseBlock(block ast.Block) string {
	code := p.parseBlockSingle(block)

	// Handle Next block for sequence
	if block.Next != nil && block.Next.Block != nil {
		nextCode := p.parseBlock(*block.Next.Block)
		if nextCode != "" {
			if code == "" {
				code = nextCode
			} else {
				code = code + "\n" + nextCode
			}
		}
	}
	return code
}

func (p *Parser) parseBlockSingle(block ast.Block) string {
	switch block.Type {
	// --- Control ---
	case "controls_if":
		return p.ctrlIf(block)
	case "controls_forRange":
		return p.ctrlForRange(block)
	case "controls_forEach":
		return p.ctrlForEach(block)
	case "controls_for_each_dict":
		return p.ctrlForEachDict(block)
	case "controls_while":
		return p.ctrlWhile(block)
	case "controls_choose":
		return p.ctrlChoose(block)
	case "controls_do_then_return":
		return p.ctrlDoThenReturn(block)
	case "controls_eval_but_ignore":
		return p.ctrlEvalButIgnore(block)
	case "controls_openAnotherScreen":
		return p.ctrlOpenScreen(block)
	case "controls_openAnotherScreenWithStartValue":
		return p.ctrlOpenScreenWithValue(block)
	case "controls_getStartValue":
		return p.makePrimitiveCall("get-start-value", []string{}, "", "get start value")
	case "controls_closeScreen":
		return p.makePrimitiveCall("close-screen", []string{}, "", "close screen")
	case "controls_closeScreenWithValue":
		return p.makePrimitiveCall("close-screen-with-value", p.fromMinVals(block.Values, 1), "any", "close screen with value")
	case "controls_closeApplication":
		return p.makePrimitiveCall("close-application", []string{}, "", "close application")
	case "controls_getPlainStartText":
		return p.makePrimitiveCall("get-plain-start-text", []string{}, "", "get plain start text")
	case "controls_closeScreenWithPlainText":
		return p.makePrimitiveCall("close-screen-with-plain-text", p.fromMinVals(block.Values, 1), "text", "close screen with plain text")
	case "controls_break":
		return "(break #f)"

	// --- Logic ---
	case "logic_boolean":
		if block.SingleField() == "TRUE" {
			return "#t"
		}
		return "#f"
	case "logic_true":
		return "#t"
	case "logic_false":
		return "#f"
	case "logic_negate":
		return p.makePrimitiveCall("yail-not", p.fromMinVals(block.Values, 1), "boolean", "not")
	case "logic_compare":
		return p.logicCompare(block)
	case "logic_operation", "logic_or":
		return p.logicOperation(block)

	// --- Math ---
	case "math_number":
		return block.SingleField()
	case "math_compare":
		return p.mathCompare(block)
	case "math_add":
		return p.makePrimitiveCall("+", p.fromMinVals(block.Values, 2), "number number", "+")
	case "math_subtract":
		return p.makePrimitiveCall("-", p.fromMinVals(block.Values, 2), "number number", "-")
	case "math_multiply":
		return p.makePrimitiveCall("*", p.fromMinVals(block.Values, 2), "number number", "*")
	case "math_division":
		return p.makePrimitiveCall("/", p.fromMinVals(block.Values, 2), "number number", "/")
	case "math_power":
		return p.makePrimitiveCall("expt", p.fromMinVals(block.Values, 2), "number number", "^")
	case "math_random_int":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("random-integer", []string{pVals.get("FROM"), pVals.get("TO")}, "number number", "random integer")
	case "math_random_float":
		return p.makePrimitiveCall("random-fraction", []string{}, "", "random fraction")
	case "math_random_set_seed":
		return p.makePrimitiveCall("random-set-seed", p.fromMinVals(block.Values, 1), "number", "random set seed")
	case "math_number_radix":
		return p.mathRadix(block)
	case "math_on_list":
		return p.mathOnList(block)
	case "math_on_list2":
		return p.mathOnList2(block)
	case "math_mode_of_list":
		return p.makePrimitiveCall("mode", p.fromMinVals(block.Values, 1), "list", "mode")
	case "math_trig", "math_sin", "math_cos", "math_tan":
		return p.mathTrig(block)
	case "math_single":
		return p.mathSingle(block)
	case "math_atan2":
		return p.makePrimitiveCall("atan2-degrees", p.fromVals(block.Values), "number number", "atan2")
	case "math_format_as_decimal":
		return p.makePrimitiveCall("format-as-decimal", p.fromMinVals(block.Values, 2), "number number", "format as decimal")
	case "math_divide":
		return p.mathDivideOther(block)
	case "math_is_a_number":
		return p.mathIsNumber(block)
	case "math_convert_number":
		return p.mathConvertNumber(block)
	case "math_convert_angles":
		return p.mathConvertAngles(block)
	case "math_bitwise":
		return p.mathBitwise(block)

	// --- Text ---
	case "text":
		return p.quote(block.SingleField())
	case "text_join":
		return p.makePrimitiveCall("string-append", p.fromMinVals(block.Values, 0), "text...", "join")
	case "text_length":
		return p.makePrimitiveCall("string-length", p.fromMinVals(block.Values, 1), "text", "length")
	case "text_isEmpty":
		return p.makePrimitiveCall("string-empty?", p.fromMinVals(block.Values, 1), "text", "is empty")
	case "text_trim":
		return p.makePrimitiveCall("string-trim", p.fromMinVals(block.Values, 1), "text", "trim")
	case "text_reverse":
		return p.makePrimitiveCall("string-reverse", p.fromMinVals(block.Values, 1), "text", "reverse")
	case "text_split_at_spaces":
		return p.makePrimitiveCall("string-split-at-spaces", p.fromMinVals(block.Values, 1), "text", "split at spaces")
	case "text_compare":
		return p.textCompare(block)
	case "text_changeCase":
		return p.textChangeCase(block)
	case "text_starts_at":
		return p.makePrimitiveCall("string-starts-at", p.fromMinVals(block.Values, 2), "text text", "starts at")
	case "text_contains":
		return p.makePrimitiveCall("string-contains", p.fromMinVals(block.Values, 2), "text text", "contains")
	case "text_split":
		return p.makePrimitiveCall("string-split", p.fromMinVals(block.Values, 2), "text text", "split")
	case "text_segment":
		return p.makePrimitiveCall("string-substring", p.fromMinVals(block.Values, 3), "text number number", "segment")
	case "text_replace_all":
		return p.makePrimitiveCall("string-replace-all", p.fromMinVals(block.Values, 3), "text text text", "replace all")
	case "obfuscated_text":
		return p.textObfuscated(block)
	case "text_replace_mappings":
		return p.makePrimitiveCall("string-replace-mappings", p.fromMinVals(block.Values, 2), "text dictionary", "replace mappings")
	case "text_is_string":
		return p.makePrimitiveCall("string?", p.fromMinVals(block.Values, 1), "any", "is string?")

	// --- Lists ---
	case "lists_create_with":
		vals := p.fromMinVals(block.Values, 0)
		return p.makePrimitiveCall("make-yail-list", vals, strings.Repeat("any ", len(vals)), "make a list")
	case "lists_add_items":
		return p.listAddItems(block)
	case "lists_is_in":
		return p.makePrimitiveCall("yail-list-member?", p.fromMinVals(block.Values, 2), "any list", "is in list?")
	case "lists_length":
		return p.makePrimitiveCall("yail-list-length", p.fromMinVals(block.Values, 1), "list", "length of list")
	case "lists_is_empty":
		return p.makePrimitiveCall("yail-list-empty?", p.fromMinVals(block.Values, 1), "list", "is list empty?")
	case "lists_pick_random_item":
		return p.makePrimitiveCall("yail-list-pick-random", p.fromMinVals(block.Values, 1), "list", "pick random item")
	case "lists_position_in":
		return p.makePrimitiveCall("yail-list-index", p.fromMinVals(block.Values, 2), "any list", "position in list")
	case "lists_select_item":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-list-get-item", []string{pVals.get("LIST"), pVals.get("NUM")}, "list number", "select list item")
	case "lists_insert_item":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-list-insert-item!", []string{pVals.get("LIST"), pVals.get("INDEX"), pVals.get("ITEM")}, "list number any", "insert list item")
	case "lists_replace_item":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-list-set-item!", []string{pVals.get("LIST"), pVals.get("NUM"), pVals.get("ITEM")}, "list number any", "replace list item")
	case "lists_remove_item":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-list-remove-item!", []string{pVals.get("LIST"), pVals.get("INDEX")}, "list number", "remove list item")
	case "lists_copy":
		return p.makePrimitiveCall("yail-list-copy", p.fromMinVals(block.Values, 1), "list", "copy list")
	case "lists_reverse":
		return p.makePrimitiveCall("yail-list-reverse", p.fromMinVals(block.Values, 1), "list", "reverse list")
	case "lists_to_csv_row":
		return p.makePrimitiveCall("yail-list-to-csv-row", p.fromMinVals(block.Values, 1), "list", "list to csv row")
	case "lists_to_csv_table":
		return p.makePrimitiveCall("yail-list-to-csv-table", p.fromMinVals(block.Values, 1), "list", "list to csv table")
	case "lists_sort":
		return p.makePrimitiveCall("yail-list-sort", p.fromMinVals(block.Values, 1), "list", "sort list")
	case "lists_is_list":
		return p.makePrimitiveCall("yail-list?", p.fromMinVals(block.Values, 1), "any", "is a list?")
	case "lists_from_csv_row":
		return p.makePrimitiveCall("yail-list-from-csv-row", p.fromMinVals(block.Values, 1), "text", "list from csv row")
	case "lists_from_csv_table":
		return p.makePrimitiveCall("yail-list-from-csv-table", p.fromMinVals(block.Values, 1), "text", "list from csv table")
	case "lists_but_first":
		return p.makePrimitiveCall("yail-list-but-first", p.fromMinVals(block.Values, 1), "list", "but first")
	case "lists_but_last":
		return p.makePrimitiveCall("yail-list-but-last", p.fromMinVals(block.Values, 1), "list", "but last")
	case "lists_slice":
		return p.makePrimitiveCall("yail-list-slice", p.fromMinVals(block.Values, 3), "list number number", "list slice")
	case "lists_lookup_in_pairs":
		return p.makePrimitiveCall("yail-lookup-in-pairs", p.fromMinVals(block.Values, 3), "any list any", "lookup in pairs")
	case "lists_join_with_separator":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-list-join-with-separator", []string{pVals.get("LIST"), pVals.get("SEPARATOR")}, "list text", "join with separator")
	case "lists_map", "lists_filter", "lists_reduce", "lists_sort_comparator", "lists_sort_key", "lists_minimum_value", "lists_maximum_value":
		return p.listHigherOrder(block)
	case "lists_append_list":
		return p.makePrimitiveCall("yail-list-append!", p.fromMinVals(block.Values, 2), "list list", "append list")

	// --- Dictionaries ---
	case "pair":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("make-yail-pair", []string{pVals.get("KEY"), pVals.get("VALUE")}, "any any", "make pair")
	case "dictionaries_create_with":
		vals := p.fromMinVals(block.Values, 0)
		return p.makePrimitiveCall("make-yail-dictionary", vals, strings.Repeat("pair ", len(vals)), "make a dictionary")
	case "dictionaries_lookup":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-dictionary-lookup", []string{pVals.get("KEY"), pVals.get("DICT"), pVals.get("NOTFOUND")}, "any dictionary any", "lookup in dictionary")
	case "dictionaries_set_pair":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-dictionary-set-pair", []string{pVals.get("KEY"), pVals.get("DICT"), pVals.get("VALUE")}, "any dictionary any", "set pair")
	case "dictionaries_delete_pair":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-dictionary-delete-pair", []string{pVals.get("KEY"), pVals.get("DICT")}, "any dictionary", "delete pair")
	case "dictionaries_recursive_lookup":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-dictionary-recursive-lookup", []string{pVals.get("KEYS"), pVals.get("DICT"), pVals.get("NOTFOUND")}, "list dictionary any", "recursive lookup")
	case "dictionaries_recursive_set":
		pVals := p.makeValueMap(block.Values)
		return p.makePrimitiveCall("yail-dictionary-recursive-set", []string{pVals.get("KEYS"), pVals.get("DICT"), pVals.get("VALUE")}, "list dictionary any", "recursive set")
	case "dictionaries_getters":
		return p.dictGetters(block)
	case "dictionaries_is_key_in":
		return p.makePrimitiveCall("yail-dictionary-is-key-in", p.fromMinVals(block.Values, 2), "any dictionary", "is key in?")
	case "dictionaries_length":
		return p.makePrimitiveCall("yail-dictionary-length", p.fromMinVals(block.Values, 1), "dictionary", "dictionary length")
	case "dictionaries_alist_to_dict":
		return p.makePrimitiveCall("yail-dictionary-alist-to-dict", p.fromMinVals(block.Values, 1), "list", "alist to dict")
	case "dictionaries_dict_to_alist":
		return p.makePrimitiveCall("yail-dictionary-dict-to-alist", p.fromMinVals(block.Values, 1), "dictionary", "dict to alist")
	case "dictionaries_copy":
		return p.makePrimitiveCall("yail-dictionary-copy", p.fromMinVals(block.Values, 1), "dictionary", "copy dict")
	case "dictionaries_combine_dicts":
		return p.makePrimitiveCall("yail-dictionary-combine-dicts", p.fromMinVals(block.Values, 2), "dictionary dictionary", "combine dicts")
	case "dictionaries_walk_tree":
		return p.makePrimitiveCall("yail-dictionary-walk", p.fromMinVals(block.Values, 2), "list dictionary", "walk tree")
	case "dictionaries_walk_all":
		return "'all"
	case "dictionaries_is_dict":
		return p.makePrimitiveCall("yail-dictionary?", p.fromMinVals(block.Values, 1), "any", "is dict?")

	// --- Colors ---
	case "color_black":
		return "-16777216"
	case "color_white":
		return "-1"
	case "color_red":
		return "-65536"
	case "color_pink":
		return "-6381921"
	case "color_orange":
		return "-26368"
	case "color_yellow":
		return "-256"
	case "color_green":
		return "-16711936"
	case "color_cyan":
		return "-16711681"
	case "color_blue":
		return "-16776961"
	case "color_magenta":
		return "-65281"
	case "color_light_gray":
		return "-3355444"
	case "color_dark_gray":
		return "-12303292"
	case "color_make_color":
		return p.makePrimitiveCall("make-color", p.fromMinVals(block.Values, 1), "list", "make color")
	case "color_split_color":
		return p.makePrimitiveCall("split-color", p.fromMinVals(block.Values, 1), "number", "split color")

	// --- Variables ---
	case "global_declaration":
		pVals := p.makeValueMap(block.Values)
		return fmt.Sprintf("(def g$%s %s)", block.SingleField(), pVals.get("VALUE"))
	case "lexical_variable_set":
		return p.variableSet(block)
	case "lexical_variable_get":
		return p.variableGet(block)
	case "local_declaration_statement", "local_declaration_expression":
		return p.localVariable(block)

	// --- Procedures ---
	case "procedures_defnoreturn", "procedures_defreturn":
		return p.procedureDef(block)
	case "procedures_callnoreturn", "procedures_callreturn":
		return p.procedureCall(block)

	// --- Components ---
	case "component_event":
		return p.componentEvent(block)
	case "component_method":
		return p.componentMethod(block)
	case "component_set_get":
		return p.componentSetGet(block)
	case "component_component_block":
		return fmt.Sprintf("(get-component %s)", block.SingleField())
	case "component_all_component_block":
		return fmt.Sprintf("(get-component-type '%s)", block.Mutation.ComponentType)

	// --- Helpers ---
	case "helpers_assets":
		return p.quote(block.SingleField())
	case "helpers_dropdown":
		return p.quote(block.SingleField())

	default:
		// Fallback for unsupported blocks
		return fmt.Sprintf(";; Unsupported block: %s", block.Type)
	}
}

// Helpers

func (p *Parser) quote(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return fmt.Sprintf("\"%s\"", s)
}

func (p *Parser) makePrimitiveCall(name string, args []string, types string, niceName string) string {
	processedArgs := make([]string, len(args))
	for i, arg := range args {
		if arg == "" {
			if strings.Contains(types, "number") {
				processedArgs[i] = "0"
			} else if strings.Contains(types, "boolean") {
				processedArgs[i] = "#f"
			} else {
				processedArgs[i] = "\"\""
			}
		} else {
			processedArgs[i] = arg
		}
	}

	return fmt.Sprintf("(call-yail-primitive %s (*list-for-runtime* %s) '(%s) \"%s\")",
		name, strings.Join(processedArgs, " "), types, niceName)
}

func (p *Parser) fromVals(values []ast.Value) []string {
	var res []string
	for _, val := range values {
		res = append(res, p.parseBlock(val.Block))
	}
	return res
}

func (p *Parser) fromMinVals(values []ast.Value, min int) []string {
	var res []string
	for _, val := range values {
		res = append(res, p.parseBlock(val.Block))
	}
	for len(res) < min {
		res = append(res, "")
	}
	return res
}

func (p *Parser) makeValueMap(values []ast.Value) *ValueMap {
	vMap := make(map[string]string)
	for _, val := range values {
		vMap[val.Name] = p.parseBlock(val.Block)
	}
	return &ValueMap{valueMap: vMap}
}

func (p *Parser) makeFieldMap(fields []ast.Field) map[string]string {
	fMap := make(map[string]string)
	for _, field := range fields {
		fMap[field.Name] = field.Value
	}
	return fMap
}

func (p *Parser) makeStatementMap(statements []ast.Statement) *StatementMap {
	sMap := make(map[string]string)
	for _, stmt := range statements {
		if stmt.Block != nil {
			sMap[stmt.Name] = p.parseBlock(*stmt.Block)
		}
	}
	return &StatementMap{statementMap: sMap}
}

// Implementations

func (p *Parser) mathCompare(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 2)
	var prim, nice string
	switch op {
	case "EQ":
		prim = "yail-equal?"
		nice = "="
	case "NEQ":
		prim = "yail-not-equal?"
		nice = "!="
	case "LT":
		prim = "<"
		nice = "<"
	case "LTE":
		prim = "<="
		nice = "<="
	case "GT":
		prim = ">"
		nice = ">"
	case "GTE":
		prim = ">="
		nice = ">="
	default:
		prim = "yail-equal?"
		nice = "unknown"
	}
	types := "number number"
	if op == "EQ" || op == "NEQ" {
		types = "any any"
	}
	return p.makePrimitiveCall(prim, vals, types, nice)
}

func (p *Parser) mathSingle(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	var prim, nice string
	switch op {
	case "ROOT":
		prim = "sqrt"
		nice = "sqrt"
	case "ABS":
		prim = "abs"
		nice = "abs"
	case "NEG":
		prim = "-"
		nice = "negate"
	case "LN":
		prim = "log"
		nice = "log"
	case "EXP":
		prim = "exp"
		nice = "exp"
	case "ROUND":
		prim = "round"
		nice = "round"
	case "CEILING":
		prim = "ceil"
		nice = "ceiling"
	case "FLOOR":
		prim = "floor"
		nice = "floor"
	default:
		prim = strings.ToLower(op)
		nice = strings.ToLower(op)
	}
	return p.makePrimitiveCall(prim, vals, "number", nice)
}

func (p *Parser) logicCompare(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 2)
	if op == "EQ" {
		return p.makePrimitiveCall("yail-equal?", vals, "any any", "=")
	}
	return p.makePrimitiveCall("yail-not-equal?", vals, "any any", "!=")
}

func (p *Parser) logicOperation(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 2)
	funcName := "and-delayed"
	if op == "OR" {
		funcName = "or-delayed"
	}
	if vals[0] == "" {
		vals[0] = "#f"
	}
	if vals[1] == "" {
		vals[1] = "#f"
	}
	return fmt.Sprintf("(%s %s %s)", funcName, vals[0], vals[1])
}

func (p *Parser) ctrlIf(block ast.Block) string {
	return p.buildIf(0, &block)
}

func (p *Parser) buildIf(index int, block *ast.Block) string {
	var testVal *ast.Value
	var valName = "IF" + strconv.Itoa(index)
	for i := range block.Values {
		if block.Values[i].Name == valName {
			testVal = &block.Values[i]
			break
		}
	}

	if testVal == nil {
		stmts := p.makeStatementMap(block.Statements)
		elseBody := stmts.get("ELSE")
		if elseBody != "" {
			return fmt.Sprintf("(begin %s)", elseBody)
		}
		return "#f"
	}

	stmts := p.makeStatementMap(block.Statements)
	doBody := stmts.get("DO" + strconv.Itoa(index))
	if doBody == "" {
		doBody = "#f"
	}

	elseCode := p.buildIf(index+1, block)
	testCode := p.parseBlock(testVal.Block)
	if testCode == "" {
		testCode = "#f"
	}

	return fmt.Sprintf("(if %s (begin %s) %s)", testCode, doBody, elseCode)
}

func (p *Parser) ctrlForRange(block ast.Block) string {
	pVals := p.makeValueMap(block.Values)
	stmts := p.makeStatementMap(block.Statements)
	varName := block.SingleField()
	start := pVals.get("START")
	end := pVals.get("END")
	step := pVals.get("STEP")
	body := stmts.get("DO")
	return fmt.Sprintf("(for-range $%s (begin %s) %s %s %s)", varName, body, start, end, step)
}

func (p *Parser) ctrlForEach(block ast.Block) string {
	pVals := p.makeValueMap(block.Values)
	stmts := p.makeStatementMap(block.Statements)
	varName := block.SingleField()
	list := pVals.get("LIST")
	body := stmts.get("DO")
	return fmt.Sprintf("(foreach $%s (begin %s) %s)", varName, body, list)
}

func (p *Parser) ctrlForEachDict(block ast.Block) string {
	pFields := p.makeFieldMap(block.Fields)
	pVals := p.makeValueMap(block.Values)
	stmts := p.makeStatementMap(block.Statements)

	key := pFields["KEY"]
	val := pFields["VALUE"]
	body := stmts.get("DO")
	dict := pVals.get("DICT")

	return fmt.Sprintf("(yail-dictionary-foreach %s (lambda ($%s $%s) (begin %s)))", dict, key, val, body)
}

func (p *Parser) ctrlWhile(block ast.Block) string {
	pVals := p.makeValueMap(block.Values)
	stmts := p.makeStatementMap(block.Statements)
	return fmt.Sprintf("(while %s (begin %s))", pVals.get("TEST"), stmts.get("DO"))
}

func (p *Parser) ctrlChoose(block ast.Block) string {
	pVals := p.makeValueMap(block.Values)
	test := pVals.get("TEST")
	thenVal := pVals.get("THENRETURN")
	elseVal := pVals.get("ELSERETURN")
	return fmt.Sprintf("(if %s %s %s)", test, thenVal, elseVal)
}

func (p *Parser) ctrlDoThenReturn(block ast.Block) string {
	pVals := p.makeValueMap(block.Values)
	stmts := p.makeStatementMap(block.Statements)
	return fmt.Sprintf("(begin %s %s)", stmts.get("STM"), pVals.get("VALUE"))
}

func (p *Parser) ctrlEvalButIgnore(block ast.Block) string {
	return p.makePrimitiveCall("begin", p.fromMinVals(block.Values, 1), "any", "eval but ignore")
}

func (p *Parser) ctrlOpenScreen(block ast.Block) string {
	return p.makePrimitiveCall("open-another-screen", p.fromMinVals(block.Values, 1), "text", "open another screen")
}

func (p *Parser) ctrlOpenScreenWithValue(block ast.Block) string {
	return p.makePrimitiveCall("open-another-screen-with-start-value", p.fromMinVals(block.Values, 2), "text any", "open another screen with start value")
}

func (p *Parser) variableSet(block ast.Block) string {
	name := block.SingleField()
	val := p.pSingle(block)
	if val == "" {
		val = "0"
	}
	if strings.HasPrefix(name, "global ") {
		realName := strings.TrimPrefix(name, "global ")
		return fmt.Sprintf("(set-this-form-environment-variable! \"%s\" %s)", realName, val)
	}
	return fmt.Sprintf("(set-lexical! $%s %s)", name, val)
}

func (p *Parser) variableGet(block ast.Block) string {
	name := block.Fields[0].Name
	if name == "VAR" {
		name = block.SingleField()
	}
	if strings.HasPrefix(name, "global ") {
		realName := strings.TrimPrefix(name, "global ")
		return fmt.Sprintf("(get-var g$%s)", realName)
	}
	return fmt.Sprintf("(lexical-value $%s)", name)
}

func (p *Parser) localVariable(block ast.Block) string {
	var decls []string
	fMap := p.makeFieldMap(block.Fields)
	vMap := p.makeValueMap(block.Values)
	i := 0
	for {
		vNameKey := "VAR" + strconv.Itoa(i)
		vName, ok := fMap[vNameKey]
		if !ok {
			break
		}
		valKey := "DECL" + strconv.Itoa(i)
		val := vMap.get(valKey)
		if val == "" {
			val = "0"
		}
		decls = append(decls, fmt.Sprintf("($%s %s)", vName, val))
		i++
	}
	stmts := p.makeStatementMap(block.Statements)
	body := stmts.get("STACK")
	if block.Type == "local_declaration_expression" {
		ret := vMap.get("RETURN")
		if ret != "" {
			return fmt.Sprintf("(let ( %s ) %s)", strings.Join(decls, " "), ret)
		}
	}
	return fmt.Sprintf("(let ( %s ) (begin %s))", strings.Join(decls, " "), body)
}

func (p *Parser) procedureDef(block ast.Block) string {
	name := block.SingleField()
	var args []string
	if block.Mutation != nil {
		for _, arg := range block.Mutation.Args {
			args = append(args, "$"+arg.Name)
		}
	}
	stmts := p.makeStatementMap(block.Statements)
	body := stmts.get("STACK")
	var bodyContent string
	if block.Type == "procedures_defnoreturn" {
		bodyContent = fmt.Sprintf("(begin %s)", body)
	} else {
		vMap := p.makeValueMap(block.Values)
		bodyContent = vMap.get("RETURN")
	}
	return fmt.Sprintf("(def (p$%s %s) %s)", name, strings.Join(args, " "), bodyContent)
}

func (p *Parser) procedureCall(block ast.Block) string {
	name := block.SingleField()
	vals := p.fromMinVals(block.Values, 0)
	return fmt.Sprintf("(p$%s %s)", name, strings.Join(vals, " "))
}

func (p *Parser) componentEvent(block ast.Block) string {
	compName := block.Mutation.InstanceName
	eventName := block.Mutation.EventName
	var args []string
	if block.Mutation != nil {
		for _, arg := range block.Mutation.Args {
			args = append(args, "$"+arg.Name)
		}
	}
	stmts := p.makeStatementMap(block.Statements)
	body := stmts.get("DO")
	return fmt.Sprintf("(define-event %s %s (%s) %s)", compName, eventName, strings.Join(args, " "), body)
}

func (p *Parser) componentMethod(block ast.Block) string {
	compName := block.Mutation.InstanceName
	methodName := block.Mutation.MethodName
	vals := p.fromMinVals(block.Values, 0)
	types := make([]string, len(vals))
	for i := range types {
		types[i] = "any"
	}
	return fmt.Sprintf("(call-component-method '%s '%s (*list-for-runtime* %s) '(%s))",
		compName, methodName, strings.Join(vals, " "), strings.Join(types, " "))
}

func (p *Parser) componentSetGet(block ast.Block) string {
	compName := block.Mutation.InstanceName
	prop := p.makeFieldMap(block.Fields)["PROP"]
	if block.Mutation.SetOrGet == "set" {
		val := p.pSingle(block)
		return fmt.Sprintf("(set-and-coerce-property! '%s '%s %s 'any)", compName, prop, val)
	} else {
		return fmt.Sprintf("(get-property '%s '%s)", compName, prop)
	}
}

func (p *Parser) mathRadix(block ast.Block) string {
	pFields := p.makeFieldMap(block.Fields)
	pVals := p.makeValueMap(block.Values)
	num := pVals.get("NUM")
	switch pFields["OP"] {
	case "DEC":
		return p.makePrimitiveCall("math-convert-dec", []string{num}, "text", "convert to dec")
	case "BIN":
		return p.makePrimitiveCall("math-convert-bin", []string{num}, "text", "convert to bin")
	case "HEX":
		return p.makePrimitiveCall("math-convert-hex", []string{num}, "text", "convert to hex")
	case "OCT":
		return p.makePrimitiveCall("math-convert-oct", []string{num}, "text", "convert to oct")
	}
	return ""
}

func (p *Parser) mathOnList(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	funcName := strings.ToLower(op)
	return p.makePrimitiveCall(funcName, vals, "list", funcName)
}

func (p *Parser) mathOnList2(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	switch op {
	case "AVG":
		return p.makePrimitiveCall("yail-list-avg", vals, "list", "avg")
	case "MIN":
		return p.makePrimitiveCall("min", vals, "list", "min")
	case "MAX":
		return p.makePrimitiveCall("max", vals, "list", "max")
	}
	return p.makePrimitiveCall(strings.ToLower(op), vals, "list", strings.ToLower(op))
}

func (p *Parser) mathTrig(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	funcName := strings.ToLower(op) + "-degrees"
	return p.makePrimitiveCall(funcName, vals, "number", strings.ToLower(op))
}

func (p *Parser) mathDivideOther(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 2)
	switch op {
	case "MODULO":
		return p.makePrimitiveCall("modulo", vals, "number number", "modulo")
	case "REMAINDER":
		return p.makePrimitiveCall("remainder", vals, "number number", "remainder")
	case "QUOTIENT":
		return p.makePrimitiveCall("quotient", vals, "number number", "quotient")
	}
	return ""
}

func (p *Parser) mathIsNumber(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	switch op {
	case "NUMBER":
		return p.makePrimitiveCall("is-number?", vals, "any", "is number?")
	case "BASE10":
		return p.makePrimitiveCall("is-base10?", vals, "text", "is base10?")
	case "HEXADECIMAL":
		return p.makePrimitiveCall("is-hexadecimal?", vals, "text", "is hex?")
	case "BINARY":
		return p.makePrimitiveCall("is-binary?", vals, "text", "is binary?")
	}
	return ""
}

func (p *Parser) mathConvertNumber(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	switch op {
	case "DEC_TO_HEX":
		return p.makePrimitiveCall("math-convert-dec-hex", vals, "number", "dec to hex")
	case "HEX_TO_DEC":
		return p.makePrimitiveCall("math-convert-hex-dec", vals, "text", "hex to dec")
	case "DEC_TO_BIN":
		return p.makePrimitiveCall("math-convert-dec-bin", vals, "number", "dec to bin")
	case "BIN_TO_DEC":
		return p.makePrimitiveCall("math-convert-bin-dec", vals, "text", "bin to dec")
	}
	return ""
}

func (p *Parser) mathConvertAngles(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	if op == "RADIANS_TO_DEGREES" {
		return p.makePrimitiveCall("radians->degrees", vals, "number", "rad to deg")
	}
	return p.makePrimitiveCall("degrees->radians", vals, "number", "deg to rad")
}

func (p *Parser) mathBitwise(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 2)
	switch op {
	case "BITAND":
		return p.makePrimitiveCall("bitwise-and", vals, "number number", "bitwise and")
	case "BITIOR":
		return p.makePrimitiveCall("bitwise-ior", vals, "number number", "bitwise or")
	case "BITXOR":
		return p.makePrimitiveCall("bitwise-xor", vals, "number number", "bitwise xor")
	}
	return ""
}

func (p *Parser) textCompare(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 2)
	switch op {
	case "LT":
		return p.makePrimitiveCall("string<?", vals, "text text", "text <")
	case "GT":
		return p.makePrimitiveCall("string>?", vals, "text text", "text >")
	case "EQUAL":
		return p.makePrimitiveCall("string=?", vals, "text text", "text =")
	case "NEQ":
		return p.makePrimitiveCall("string-not=?", vals, "text text", "text !=")
	}
	return ""
}

func (p *Parser) textChangeCase(block ast.Block) string {
	op := block.SingleField()
	vals := p.fromMinVals(block.Values, 1)
	if op == "UPPERCASE" {
		return p.makePrimitiveCall("string-to-upper-case", vals, "text", "upcase")
	}
	return p.makePrimitiveCall("string-to-lower-case", vals, "text", "downcase")
}

func (p *Parser) textObfuscated(block ast.Block) string {
	text := block.SingleField()
	return p.quote(text)
}

func (p *Parser) listAddItems(block ast.Block) string {
	pVals := p.makeValueMap(block.Values)
	list := pVals.get("LIST")

	var items []string
	items = append(items, list)
	types := "list"

	i := 0
	for {
		key := "ITEM" + strconv.Itoa(i)
		val := pVals.get(key)
		if val == "" {
			break
		}
		items = append(items, val)
		types += " any"
		i++
	}

	return p.makePrimitiveCall("yail-list-add-to-list!", items, types, "add items to list")
}

func (p *Parser) listHigherOrder(block ast.Block) string {
	pFields := p.makeFieldMap(block.Fields)
	pVals := p.makeValueMap(block.Values)

	switch block.Type {
	case "lists_map":
		return fmt.Sprintf("(yail-list-map (lambda ($%s) %s) %s)", pFields["VAR"], pVals.get("TO"), pVals.get("LIST"))
	case "lists_filter":
		return fmt.Sprintf("(yail-list-filter (lambda ($%s) %s) %s)", pFields["VAR"], pVals.get("TEST"), pVals.get("LIST"))
	case "lists_reduce":
		return fmt.Sprintf("(yail-list-reduce (lambda ($%s $%s) %s) %s %s)",
			pFields["VAR1"], pFields["VAR2"], pVals.get("COMBINE"), pVals.get("LIST"), pVals.get("INITANSWER"))
	case "lists_sort_comparator":
		return fmt.Sprintf("(yail-list-sort-comparator (lambda ($%s $%s) %s) %s)",
			pFields["VAR1"], pFields["VAR2"], pVals.get("COMPARE"), pVals.get("LIST"))
	case "lists_sort_key":
		return fmt.Sprintf("(yail-list-sort-key (lambda ($%s) %s) %s)",
			pFields["VAR"], pVals.get("KEY"), pVals.get("LIST"))
	}
	return ""
}

func (p *Parser) dictGetters(block ast.Block) string {
	op := block.SingleField()
	dict := p.makeValueMap(block.Values).get("DICT")

	if op == "KEYS" {
		return p.makePrimitiveCall("yail-dictionary-get-keys", []string{dict}, "dictionary", "get keys")
	}
	return p.makePrimitiveCall("yail-dictionary-get-values", []string{dict}, "dictionary", "get values")
}

func (p *Parser) pSingle(block ast.Block) string {
	if len(block.Values) == 0 {
		return ""
	}
	return p.parseBlock(block.Values[0].Block)
}
