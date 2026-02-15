package blocklytomist

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
	"Falcon/code/ast/components"
	"Falcon/code/ast/control"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/list"
	"Falcon/code/ast/method"
	"Falcon/code/ast/procedures"
	"Falcon/code/ast/variables"
	"Falcon/code/lex"
	"encoding/xml"
	"strconv"
	"strings"
)

type ValueMap struct {
	valueMap map[string]ast.Expr
}

func (v *ValueMap) getUnsafe(name string) ast.Expr {
	return v.valueMap[name]
}

func (v *ValueMap) get(name string) ast.Expr {
	value := v.valueMap[name]
	if value == nil {
		return &common.EmptySocket{}
	}
	return value
}

type StatementMap struct {
	statementMap map[string][]ast.Expr
}

func (s *StatementMap) getUnsafe(name string) []ast.Expr {
	return s.statementMap[name]
}

func (s *StatementMap) get(name string) []ast.Expr {
	value := s.statementMap[name]
	if value == nil {
		return []ast.Expr{}
	}
	return value
}

type Parser struct {
	xmlContent string
}

func NewParser(xmlContent string) *Parser {
	return &Parser{xmlContent: xmlContent}
}

func (p *Parser) GenerateAST() []ast.Expr {
	return p.parseAllBlocks(p.decodeXML())
}

func (p *Parser) decodeXML() []ast.Block {
	decoder := xml.NewDecoder(strings.NewReader(p.xmlContent))
	decoder.Strict = false
	decoder.DefaultSpace = ""

	var root ast.XmlRoot
	if err := decoder.Decode(&root); err != nil {
		panic(err)
	}
	return root.Blocks
}

func (p *Parser) parseAllBlocks(allBlocks []ast.Block) []ast.Expr {
	var parsedBlocks []ast.Expr
	for i := range allBlocks {
		parsedBlocks = append(parsedBlocks, p.parseBlock(allBlocks[i]))
	}
	return parsedBlocks
}

func (p *Parser) singleExpr(block ast.Block) ast.Expr {
	if len(block.Values) == 0 {
		return &common.EmptySocket{}
	}
	return p.parseBlock(block.Values[0].Block)
}

func (p *Parser) parseBlock(block ast.Block) ast.Expr {
	switch block.Type {
	case "controls_if":
		return p.ctrlIf(block)
	case "controls_forRange":
		return p.ctrlForRange(block)
	case "controls_forEach":
		return &control.Each{
			IName:    block.SingleField(),
			Iterable: p.singleExpr(block),
			Body:     p.optSingleBody(block)}
	case "controls_for_each_dict":
		return p.ctrlForEachDict(block)
	case "controls_while":
		return &control.While{
			Condition: p.singleExpr(block),
			Body:      p.optSingleBody(block)}
	case "controls_choose":
		return p.ctrlChoose(block)
	case "controls_do_then_return":
		return &control.Do{Body: p.optSingleBody(block), Result: p.singleExpr(block)}
	case "controls_eval_but_ignore":
		return common.MakeFuncCall("println", p.singleExpr(block))
	case "controls_openAnotherScreen":
		return common.MakeFuncCall("openScreen", p.singleExpr(block))
	case "controls_openAnotherScreenWithStartValue":
		return common.MakeFuncCall("openScreenWithValue", p.singleExpr(block))
	case "controls_getStartValue":
		return common.MakeFuncCall("getStartValue")
	case "controls_closeScreen":
		return common.MakeFuncCall("closeScreen")
	case "controls_closeScreenWithValue":
		return common.MakeFuncCall("closeScreenWithValue", p.singleExpr(block))
	case "controls_closeApplication":
		return common.MakeFuncCall("closeApp")
	case "controls_getPlainStartText":
		return common.MakeFuncCall("getPlainStartText")
	case "controls_closeScreenWithPlainText":
		return common.MakeFuncCall("closeScreenWithPlainText", p.singleExpr(block))
	case "controls_break":
		return &control.Break{}

	case "logic_boolean", "logic_true", "logic_false":
		return &fundamentals.Boolean{Value: block.SingleField() == "TRUE"}
	case "logic_negate":
		return &fundamentals.Not{Expr: p.singleExpr(block)}
	case "logic_compare", "logic_operation", "logic_or":
		return p.logicExpr(block)

	case "text":
		return &fundamentals.Text{Content: block.SingleField()}
	case "text_join":
		return p.makeBinary("_", p.fromMinVals(block.Values, 1))
	case "text_length":
		return p.makePropCall("textLen", p.singleExpr(block))
	case "text_isEmpty":
		return p.makeQuestion(lex.Text, block, "emptyText")
	case "text_trim":
		return p.makePropCall("trim", p.singleExpr(block))
	case "text_reverse":
		return p.makePropCall("reverse", p.singleExpr(block))
	case "text_split_at_spaces":
		return p.makePropCall("splitAtSpaces", p.singleExpr(block))
	case "text_compare":
		return p.textCompare(block)
	case "text_changeCase":
		return p.textChangeCase(block)
	case "text_starts_at":
		return p.textStartsWith(block)
	case "text_contains":
		return p.textContains(block)
	case "text_split":
		return p.textSplit(block)
	case "text_segment":
		return p.textSegment(block)
	case "text_replace_all":
		return p.textReplace(block)
	case "obfuscated_text":
		return p.textObfuscate(block)
	case "text_replace_mappings":
		return p.textReplaceMap(block)
	case "text_is_string":
		return p.makeQuestion(lex.Text, block, "text")

	case "math_number":
		return &fundamentals.Number{Content: block.SingleField()}
	case "math_compare", "math_bitwise":
		return p.mathExpr(block)
	case "math_add":
		return p.makeBinary("+", p.fromMinVals(block.Values, 2))
	case "math_subtract":
		return p.makeBinary("-", p.fromMinVals(block.Values, 2))
	case "math_multiply":
		return p.makeBinary("*", p.fromMinVals(block.Values, 2))
	case "math_division":
		return p.makeBinary("/", p.fromMinVals(block.Values, 2))
	case "math_power":
		return p.makeBinary("^", p.fromMinVals(block.Values, 2))
	case "math_random_int":
		return p.mathRandom(block)
	case "math_random_float":
		return common.MakeFuncCall("randFloat")
	case "math_random_set_seed":
		return common.MakeFuncCall("setRandSeed", p.singleExpr(block))
	case "math_number_radix":
		return p.mathRadix(block)
	case "math_on_list": // min() and max()
		return common.MakeFuncCall(strings.ToLower(block.SingleField()), p.fromMinVals(block.Values, 1)...)
	case "math_on_list2":
		return p.mathOnList2(block)
	case "math_mode_of_list":
		return common.MakeFuncCall("modeOf", p.singleExpr(block))
	case "math_trig", "math_sin", "math_cos", "math_tan":
		return common.MakeFuncCall(strings.ToLower(block.SingleField()), p.singleExpr(block))
	case "math_single":
		return p.mathSingle(block)
	case "math_atan2":
		return common.MakeFuncCall("aTan2", p.fromVals(block.Values)...)
	case "math_format_as_decimal":
		return common.MakeFuncCall("formatDecimal", p.fromMinVals(block.Values, 2)...)
	case "math_divide":
		return p.mathDivide(block)
	case "math_is_a_number":
		return p.mathIsNumber(block)
	case "math_convert_number":
		return p.mathConvertNumber(block)
	case "math_convert_angles":
		return p.mathConvertAngles(block)

	case "matrices_create":
		return p.matricesCreate(block)
	case "matrices_create_multidim":
		return p.matricesNdArray(block)

	case "lists_create_with":
		return &fundamentals.List{Elements: p.fromMinVals(block.Values, 0)}
	case "lists_add_items":
		return p.listAddItem(block)
	case "lists_is_in":
		return p.listContainsItem(block)
	case "lists_length":
		return p.makePropCall("listLen", p.singleExpr(block))
	case "lists_is_empty":
		return p.makeQuestion(lex.OpenSquare, block, "emptyList")
	case "lists_pick_random_item":
		return p.makePropCall("random", p.singleExpr(block))
	case "lists_position_in":
		return p.listIndexOf(block)
	case "lists_select_item":
		return p.listSelectItem(block)
	case "lists_insert_item":
		return p.listInsertItem(block)
	case "lists_replace_item":
		return p.listReplaceItem(block)
	case "lists_remove_item":
		return p.listRemoveItem(block)
	case "lists_copy":
		return common.MakeFuncCall("copyList", p.singleExpr(block))
	case "lists_reverse":
		return p.makePropCall("reverseList", p.singleExpr(block))
	case "lists_to_csv_row":
		return p.makePropCall("toCsvRow", p.singleExpr(block))
	case "lists_to_csv_table":
		return p.makePropCall("toCsvTable", p.singleExpr(block))
	case "lists_sort":
		return p.makePropCall("sort", p.singleExpr(block))
	case "lists_is_list":
		return p.makeQuestion(lex.OpenSquare, block, "list")
	case "lists_from_csv_row":
		return p.makePropCall("csvRowToList", p.singleExpr(block))
	case "lists_from_csv_table":
		return p.makePropCall("csvTableToList", p.singleExpr(block))
	case "lists_but_first":
		return p.makePropCall("allButFirst", p.singleExpr(block))
	case "lists_but_last":
		return p.makePropCall("allButLast", p.singleExpr(block))
	case "lists_lookup_in_pairs":
		return p.listLookupPairs(block)
	case "lists_join_with_separator":
		return p.listJoin(block)
	case "lists_slice":
		return p.listSlice(block)
	case "lists_map":
		return p.listMap(block)
	case "lists_filter":
		return p.listFilter(block)
	case "lists_reduce":
		return p.listReduce(block)
	case "lists_sort_comparator":
		return p.listSortComparator(block)
	case "lists_sort_key":
		return p.listSortKeyComparator(block)
	case "lists_minimum_value":
		return p.listTransMin(block)
	case "lists_maximum_value":
		return p.listTransMax(block)
	case "lists_append_list":
		return p.listAppend(block)

	case "pair":
		return p.dictPair(block)
	case "dictionaries_create_with":
		return &fundamentals.Dictionary{Elements: p.fromMinVals(block.Values, 0)}
	case "dictionaries_lookup":
		return p.dictLookup(block)
	case "dictionaries_set_pair":
		return p.dictSet(block)
	case "dictionaries_delete_pair":
		return p.dictRemove(block)
	case "dictionaries_recursive_lookup":
		return p.dictLookupPath(block)
	case "dictionaries_recursive_set":
		return p.dictSetPath(block)
	case "dictionaries_getters":
		return p.dictGetters(block)
	case "dictionaries_is_key_in":
		return p.dictHasKey(block)
	case "dictionaries_length":
		return p.makePropCall("dictLen", p.singleExpr(block))
	case "dictionaries_alist_to_dict":
		return p.makePropCall("pairsToDict", p.singleExpr(block))
	case "dictionaries_dict_to_alist":
		return p.makePropCall("toPairs", p.singleExpr(block))
	case "dictionaries_copy":
		return common.MakeFuncCall("copyDict", p.singleExpr(block))
	case "dictionaries_combine_dicts":
		return p.dictCombine(block)
	case "dictionaries_walk_tree":
		return p.dictWalkTree(block)
	case "dictionaries_walk_all":
		return &fundamentals.WalkAll{}
	case "dictionaries_is_dict":
		return p.makeQuestion(lex.OpenCurly, block, "dict")

	case "color_black":
		return p.makeColor(block)
	case "color_white":
		return p.makeColor(block)
	case "color_red":
		return p.makeColor(block)
	case "color_pink":
		return p.makeColor(block)
	case "color_orange":
		return p.makeColor(block)
	case "color_yellow":
		return p.makeColor(block)
	case "color_green":
		return p.makeColor(block)
	case "color_cyan":
		return p.makeColor(block)
	case "color_blue":
		return p.makeColor(block)
	case "color_magenta":
		return p.makeColor(block)
	case "color_light_gray":
		return p.makeColor(block)
	case "color_dark_gray":
		return p.makeColor(block)
	case "color_light_green":
		return p.makeColor(block)
	case "color_gray":
		return p.makeColor(block)
	case "color_make_color":
		return common.MakeFuncCall("makeColor", p.singleExpr(block))
	case "color_split_color":
		return common.MakeFuncCall("splitColor", p.singleExpr(block))

	case "global_declaration":
		return &variables.Global{Name: block.SingleField(), Value: p.singleExpr(block)}
	case "lexical_variable_get":
		return p.variableGet(block)
	case "lexical_variable_set":
		return p.variableSet(block)
	case "local_declaration_statement", "local_declaration_expression":
		return p.variableSmts(block)

	case "procedures_defnoreturn":
		return p.voidProcedure(block)
	case "procedures_defreturn":
		return p.returnProcedure(block)
	case "procedures_callnoreturn", "procedures_callreturn":
		return p.procedureCall(block)

	case "helpers_assets":
		return &fundamentals.Text{Content: block.SingleField()}
	case "helpers_dropdown":
		return &fundamentals.HelperDropdown{Key: block.Mutation.Key, Option: block.SingleField()}

	case "component_component_block":
		return &fundamentals.Component{Name: block.SingleField(), Type: block.Mutation.ComponentType}
	case "component_set_get":
		return p.componentProp(block)
	case "component_event":
		return p.componentEvent(block)
	case "component_method":
		return p.componentMethod(block)
	case "component_all_component_block":
		return &components.EveryComponent{Type: block.Mutation.ComponentType}
	default:
		if strings.HasPrefix(block.Type, "color_") && len(block.Fields) > 0 {
			return p.makeColor(block)
		}
		if strings.HasPrefix(block.Type, "helpers_") {
			return &fundamentals.Text{Content: block.SingleField()}
		}
		println("Unsupported block type: " + block.Type)
		panic("Unsupported block type: " + block.Type)
	}
}

func (p *Parser) componentMethod(block ast.Block) ast.Expr {
	if block.Mutation.IsGeneric {
		pVals := p.makeValueMap(block.Values)
		var callArgs []ast.Expr

		for i := 0; ; i++ {
			aArg := pVals.getUnsafe("ARG" + strconv.Itoa(i))
			if aArg == nil {
				break
			}
			callArgs = append(callArgs, aArg)
		}
		return &components.GenericMethodCall{
			Component:     pVals.get("COMPONENT"),
			ComponentType: block.Mutation.ComponentType,
			Method:        block.Mutation.MethodName,
			Args:          callArgs,
			Returning:     block.Mutation.Shape == "value",
		}
	}
	return &components.MethodCall{
		ComponentName: block.Mutation.InstanceName,
		ComponentType: block.Mutation.ComponentType,
		Method:        block.Mutation.MethodName,
		Args:          p.fromVals(block.Values),
	}
}

func (p *Parser) componentEvent(block ast.Block) ast.Expr {
	var mutArgsNames []ast.Arg
	if block.Mutation != nil {
		mutArgsNames = block.Mutation.Args
	}
	paramNames := make([]string, len(mutArgsNames))
	for i := range mutArgsNames {
		paramNames[i] = mutArgsNames[i].Name
	}
	if block.Mutation.IsGeneric {
		return &components.GenericEvent{
			ComponentType: block.Mutation.ComponentType,
			Event:         block.Mutation.EventName,
			Parameters:    paramNames,
			Body:          p.optSingleBody(block),
		}
	}
	return &components.Event{
		ComponentName: block.Mutation.InstanceName,
		ComponentType: block.Mutation.ComponentType,
		Event:         block.Mutation.EventName,
		Parameters:    paramNames,
		Body:          p.optSingleBody(block),
	}
}

func (p *Parser) componentProp(block ast.Block) ast.Expr {
	pFields := p.makeFieldMap(block.Fields)
	property := pFields["PROP"]
	isSet := block.Mutation.SetOrGet == "set"

	if block.Mutation.IsGeneric {
		pVals := p.makeValueMap(block.Values)
		if isSet {
			return &components.GenericPropertySet{
				Component:     pVals.get("COMPONENT"),
				ComponentType: block.Mutation.ComponentType,
				Property:      property,
				Value:         pVals.get("VALUE"),
			}
		}
		return &components.GenericPropertyGet{
			Component:     pVals.get("COMPONENT"),
			ComponentType: block.Mutation.ComponentType,
			Property:      property,
		}
	}

	if isSet {
		return &components.PropertySet{
			ComponentName: pFields["COMPONENT_SELECTOR"],
			ComponentType: block.Mutation.ComponentType,
			Property:      property,
			Value:         p.singleExpr(block),
		}
	}
	return &components.PropertyGet{
		ComponentName: pFields["COMPONENT_SELECTOR"],
		ComponentType: block.Mutation.ComponentType,
		Property:      property,
	}
}

func (p *Parser) ctrlChoose(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	then := pVals.get("THENRETURN")
	elze := pVals.get("ELSERETURN")
	return control.MakeSimpleIf(pVals.get("TEST"), []ast.Expr{then}, []ast.Expr{elze})
}

func (p *Parser) ctrlForEachDict(block ast.Block) ast.Expr {
	pFields := p.makeFieldMap(block.Fields)
	return &control.EachPair{
		KeyName:   pFields["KEY"],
		ValueName: pFields["VALUE"],
		Iterable:  p.singleExpr(block),
		Body:      p.optSingleBody(block),
	}
}

func (p *Parser) ctrlForRange(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return &control.For{
		IName: block.SingleField(),
		From:  pVals.get("START"),
		To:    pVals.get("END"),
		By:    pVals.get("STEP"),
		Body:  p.optSingleBody(block),
	}
}

func (p *Parser) ctrlIf(block ast.Block) ast.Expr {
	conditions := p.fromVals(block.Values)
	statementMap := p.makeStatementMap(block.Statements)

	var bodies [][]ast.Expr
	elseBody := statementMap.getUnsafe("ELSE")

	for i := range conditions {
		bodies = append(bodies, statementMap.get("DO"+strconv.Itoa(i)))
	}
	return &control.If{Conditions: conditions, Bodies: bodies, ElseBody: elseBody}
}

func (p *Parser) logicExpr(block ast.Block) ast.Expr {
	var pOperation string
	switch block.SingleField() {
	case "EQ":
		pOperation = "=="
	case "NEQ":
		pOperation = "!="
	case "AND":
		pOperation = "&&"
	case "OR":
		pOperation = "||"
	default:
		panic("Unknown Logic Compare operation: " + block.SingleField())
	}
	return p.makeBinary(pOperation, p.fromMinVals(block.Values, 2))
}

func (p *Parser) procedureCall(block ast.Block) ast.Expr {
	var mutArgsNames []ast.Arg
	if block.Mutation != nil {
		mutArgsNames = block.Mutation.Args
	}
	paramNames := make([]string, len(mutArgsNames))
	for i := range mutArgsNames {
		paramNames[i] = mutArgsNames[i].Name
	}
	procedureName := block.SingleField()
	args := p.fromVals(block.Values)
	return &procedures.Call{
		Name:       procedureName,
		Parameters: paramNames,
		Arguments:  args,
		Returning:  block.Type == "procedures_callreturn",
	}
}

func (p *Parser) returnProcedure(block ast.Block) ast.Expr {
	procedureName := p.makeFieldMap(block.Fields)["NAME"]
	var mutArgs []ast.Arg
	if block.Mutation != nil {
		mutArgs = block.Mutation.Args
	}
	paramNames := make([]string, len(mutArgs))
	for i := range mutArgs {
		paramNames[i] = mutArgs[i].Name
	}
	return &procedures.RetProcedure{
		Name:       procedureName,
		Parameters: paramNames,
		Result:     p.singleExpr(block),
	}
}

func (p *Parser) voidProcedure(block ast.Block) ast.Expr {
	procedureName := p.makeFieldMap(block.Fields)["NAME"]
	var mutArgs []ast.Arg
	if block.Mutation != nil {
		mutArgs = block.Mutation.Args
	}
	paramNames := make([]string, len(mutArgs))
	for i := range mutArgs {
		paramNames[i] = mutArgs[i].Name
	}
	return &procedures.VoidProcedure{
		Name:       procedureName,
		Parameters: paramNames,
		Body:       p.optSingleBody(block),
	}
}

func (p *Parser) variableSmts(block ast.Block) ast.Expr {
	numOfVars := len(block.Mutation.LocalNames)
	fieldMap := p.makeFieldMap(block.Fields)
	valueMap := p.makeValueMap(block.Values)

	varNames := make([]string, numOfVars)
	varValues := make([]ast.Expr, numOfVars)

	for i := 0; i < numOfVars; i++ {
		varNames[i] = fieldMap["VAR"+strconv.Itoa(i)]
		varValues[i] = valueMap.get("DECL" + strconv.Itoa(i))
	}
	if block.GetType() == "local_declaration_statement" {
		return &variables.Var{
			Names:  varNames,
			Values: varValues,
			Body:   p.optSingleBody(block),
		}
	}
	return &variables.VarResult{Names: varNames, Values: varValues, Result: valueMap.get("RETURN")}
}

func (p *Parser) variableSet(block ast.Block) ast.Expr {
	varName := block.SingleField()
	isGlobal := strings.HasPrefix(varName, "global ")
	if isGlobal {
		varName = varName[len("global "):]
	}
	return variables.Set{Global: isGlobal, Name: varName, Expr: p.singleExpr(block)}
}

func (p *Parser) variableGet(block ast.Block) ast.Expr {
	varName := block.Fields[0].Name
	if varName == "VAR" {
		varName = block.SingleField()
	}
	isGlobal := strings.HasPrefix(varName, "global ")
	if isGlobal {
		varName = varName[len("global "):]
	}
	return &variables.Get{Where: lex.MakeFakeToken(lex.Global), Global: isGlobal, Name: varName}
}

func (p *Parser) dictWalkTree(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("walkTree", pVals.get("DICT"), pVals.get("PATH"))
}

func (p *Parser) dictCombine(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("mergeInto", pVals.get("DICT2"), pVals.get("DICT1"))
}

func (p *Parser) dictHasKey(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("containsKey", pVals.get("DICT"), pVals.get("KEY"))
}

func (p *Parser) dictGetters(block ast.Block) ast.Expr {
	var pOperation string
	switch block.SingleField() {
	case "KEYS":
		pOperation = "keys"
	case "VALUES":
		pOperation = "values"
	default:
		panic("Unknown DictGetters operation: " + block.SingleField())
	}
	return p.makePropCall(pOperation, p.singleExpr(block))
}

func (p *Parser) dictSetPath(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("setAtPath", pVals.get("DICT"), pVals.get("KEYS"), pVals.get("VALUE"))
}

func (p *Parser) dictLookupPath(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("getAtPath", pVals.get("DICT"), pVals.get("KEYS"), pVals.get("NOTFOUND"))
}

func (p *Parser) dictRemove(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("remove", pVals.get("DICT"), pVals.get("KEY"))
}

func (p *Parser) dictSet(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("set", pVals.get("KEY"), pVals.get("VALUE"))
}

func (p *Parser) dictLookup(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("get", pVals.get("DICT"), pVals.get("KEY"), pVals.get("NOTFOUND"))
}

func (p *Parser) dictPair(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return &fundamentals.Pair{Key: pVals.get("KEY"), Value: pVals.get("VALUE")}
}

func (p *Parser) listAppend(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("appendList", pVals.get("LIST0"), pVals.get("LIST1"))
}

func (p *Parser) listTransMax(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	pFields := p.makeFieldMap(block.Fields)
	return &list.Transformer{
		Where:       lex.MakeFakeToken(lex.OpenSquare),
		List:        pVals.get("LIST"),
		Name:        "max",
		Args:        []ast.Expr{},
		Names:       []string{pFields["VAR1"], pFields["VAR2"]},
		Transformer: pVals.get("COMPARE"),
	}
}

func (p *Parser) listTransMin(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	pFields := p.makeFieldMap(block.Fields)
	return &list.Transformer{
		Where:       lex.MakeFakeToken(lex.OpenSquare),
		List:        pVals.get("LIST"),
		Name:        "min",
		Args:        []ast.Expr{},
		Names:       []string{pFields["VAR1"], pFields["VAR2"]},
		Transformer: pVals.get("COMPARE"),
	}
}

func (p *Parser) listSortKeyComparator(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return &list.Transformer{
		Where:       lex.MakeFakeToken(lex.OpenSquare),
		List:        pVals.get("LIST"),
		Name:        "sortByKey",
		Args:        []ast.Expr{},
		Names:       []string{block.SingleField()},
		Transformer: pVals.get("KEY"),
	}
}

func (p *Parser) listSortComparator(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	pFields := p.makeFieldMap(block.Fields)
	return &list.Transformer{
		Where:       lex.MakeFakeToken(lex.OpenSquare),
		List:        pVals.get("LIST"),
		Name:        "sort",
		Args:        []ast.Expr{},
		Names:       []string{pFields["VAR1"], pFields["VAR2"]},
		Transformer: pVals.get("COMPARE"),
	}
}

func (p *Parser) listReduce(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	pFields := p.makeFieldMap(block.Fields)
	return &list.Transformer{
		Where:       lex.MakeFakeToken(lex.OpenSquare),
		List:        pVals.get("LIST"),
		Name:        "reduce",
		Args:        []ast.Expr{pVals.get("INITANSWER")},
		Names:       []string{pFields["VAR1"], pFields["VAR2"]},
		Transformer: pVals.get("COMBINE"),
	}
}

func (p *Parser) listFilter(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return &list.Transformer{
		Where:       lex.MakeFakeToken(lex.OpenSquare),
		List:        pVals.get("LIST"),
		Name:        "filter",
		Args:        []ast.Expr{},
		Names:       []string{block.SingleField()},
		Transformer: pVals.get("TEST"),
	}
}

func (p *Parser) listMap(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return &list.Transformer{
		Where:       lex.MakeFakeToken(lex.OpenSquare),
		List:        pVals.get("LIST"),
		Name:        "map",
		Args:        []ast.Expr{},
		Names:       []string{block.SingleField()},
		Transformer: pVals.get("TO"),
	}
}

func (p *Parser) listSlice(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("slice", pVals.get("LIST"), pVals.get("INDEX1"), pVals.get("INDEX2"))
}

func (p *Parser) listJoin(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("join", pVals.get("LIST"), pVals.get("SEPARATOR"))
}

func (p *Parser) listLookupPairs(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("lookupInPairs", pVals.get("LIST"), pVals.get("KEY"), pVals.get("NOTFOUND"))
}

func (p *Parser) listRemoveItem(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("remove", pVals.get("LIST"), pVals.get("INDEX"))
}

func (p *Parser) listReplaceItem(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return &list.Set{List: pVals.get("LIST"), Index: pVals.get("NUM"), Value: pVals.get("ITEM")}
}

func (p *Parser) listInsertItem(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("insert", pVals.get("LIST"), pVals.get("INDEX"), pVals.get("ITEM"))
}

func (p *Parser) listSelectItem(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return &list.Get{List: pVals.get("LIST"), Index: pVals.get("NUM")}
}

func (p *Parser) listIndexOf(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("indexOf", pVals.get("LIST"), pVals.get("ITEM"))
}

func (p *Parser) listContainsItem(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("containsItem", pVals.get("LIST"), pVals.get("ITEM"))
}

func (p *Parser) listAddItem(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	numElements := block.Mutation.ItemCount
	arrElements := make([]ast.Expr, numElements)
	for i := 0; i < numElements; i++ {
		arrElements[i] = pVals.get("ITEM" + strconv.Itoa(i))
	}
	return p.makePropCall("add", pVals.get("LIST"), arrElements...)
}

func (p *Parser) textReplaceMap(block ast.Block) ast.Expr {
	var pOperation string
	switch block.SingleField() {
	case "LONGEST_STRING_FIRST":
		pOperation = "replaceFromLongestFirst"
	case "DICTIONARY_ORDER":
		pOperation = "replaceFrom"
	default:
		panic("Unknown Text Replace Map operation: " + block.SingleField())
	}
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall(pOperation, pVals.get("TEXT"), pVals.get("MAPPINGS"))
}

func (p *Parser) textObfuscate(block ast.Block) ast.Expr {
	return &common.Transform{
		Where: lex.MakeFakeToken(lex.Text),
		On:    &fundamentals.Text{Content: block.SingleField()},
		Name:  "obfuscate"}
}

func (p *Parser) textSegment(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("segment", pVals.get("TEXT"), pVals.get("START"), pVals.get("LENGTH"))
}

func (p *Parser) textReplace(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("replace", pVals.get("TEXT"), pVals.get("SEGMENT"), pVals.get("REPLACEMENT"))
}

func (p *Parser) textSplit(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	var pOperation string
	switch block.SingleField() {
	case "SPLIT":
		pOperation = "split"
	case "SPLITATFIRST":
		pOperation = "splitAtFirst"
	case "SPLITATANY":
		pOperation = "splitAtAny"
	case "SPLITATFIRSTOFANY":
		pOperation = "splitAtFirstOfAny"
	default:
		panic("Unsupported Text Split operation: " + block.SingleField())
	}
	return p.makePropCall(pOperation, pVals.get("TEXT"), pVals.get("AT"))
}

func (p *Parser) textContains(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	var pOperation string
	if len(block.Fields) > 0 {
		switch block.SingleField() {
		case "CONTAINS":
			pOperation = "contains"
		case "CONTAINS_ANY":
			pOperation = "containsAny"
		case "CONTAINS_ALL":
			pOperation = "containsAll"
		default:
			panic("Unsupported Text Contains operation: " + block.SingleField())
		}
	} else {
		pOperation = "contains"
	}
	return p.makePropCall(pOperation, pVals.get("TEXT"), pVals.get("PIECE"))
}

func (p *Parser) textStartsWith(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return p.makePropCall("startsWith", pVals.get("TEXT"), pVals.get("PIECE"))
}

func (p *Parser) textChangeCase(block ast.Block) ast.Expr {
	var pOperation string
	switch block.SingleField() {
	case "UPCASE":
		pOperation = "uppercase"
	case "DOWNCASE":
		pOperation = "lowercase"
	default:
		panic("Unsupported Text Change Case operation type: " + block.SingleField())
	}
	return p.makePropCall(pOperation, p.singleExpr(block))
}

func (p *Parser) textCompare(block ast.Block) ast.Expr {
	var pOperation string
	switch block.SingleField() {
	case "EQUAL":
		pOperation = "==="
	case "NEQ":
		pOperation = "!=="
	case "LT":
		pOperation = "<<"
	case "GT":
		pOperation = ">>"
	default:
		panic("Unknown Text Compare operation: " + block.SingleField())
	}
	return p.makeBinary(pOperation, p.fromMinVals(block.Values, 2))
}

func (p *Parser) matricesNdArray(block ast.Block) ast.Expr {
	pVals := p.makeValueMap(block.Values)
	return common.MakeFuncCall("makeNdArray", pVals.get("DIM"), pVals.get("INITIAL"))
}

func (p *Parser) matricesCreate(block ast.Block) ast.Expr {
	pFields := p.makeFieldMap(block.Fields)
	numRows, err := strconv.Atoi(pFields["ROWS"])
	if err != nil {
		panic(err)
	}
	numCols, err := strconv.Atoi(pFields["COLS"])
	if err != nil {
		panic(err)
	}
	matrix := make([]ast.Expr, numRows)
	for i := range matrix {
		row := make([]ast.Expr, numCols)
		for j := range row {
			row[j] = &fundamentals.Number{Content: pFields["MATRIX_"+strconv.Itoa(i)+"_"+strconv.Itoa(j)]}
		}
		matrix[i] = &fundamentals.List{Elements: row}
	}
	return &fundamentals.List{Elements: matrix}
}

func (p *Parser) mathConvertAngles(block ast.Block) ast.Expr {
	var funcName string
	switch block.SingleField() {
	case "RADIANS_TO_DEGREES":
		funcName = "degrees"
	case "DEGREES_TO_RADIANS":
		funcName = "radians"
	}
	return common.MakeFuncCall(funcName, p.singleExpr(block))
}

func (p *Parser) mathConvertNumber(block ast.Block) ast.Expr {
	var funcName string
	switch block.SingleField() {
	case "DEC_TO_HEX":
		funcName = "decToHex"
	case "DEC_TO_BIN":
		funcName = "decToBin"
	case "HEX_TO_DEC":
		funcName = "hexToDec"
	case "BIN_TO_DEC":
		funcName = "binToDec"
	default:
		panic("Unknown MathConvertNumber type: " + block.SingleField())
	}
	return common.MakeFuncCall(funcName, p.singleExpr(block))
}

func (p *Parser) mathIsNumber(block ast.Block) ast.Expr {
	var question string
	switch block.SingleField() {
	case "NUMBER":
		question = "number"
	case "BINARY":
		question = "bin"
	case "HEXADECIMAL":
		question = "hexa"
	case "BASE10":
		question = "base10"
	default:
		panic("Unknown MathIsNumber type: " + block.SingleField())
	}
	return p.makeQuestion(lex.Number, block, question)
}

func (p *Parser) mathDivide(block ast.Block) ast.Expr {
	var funcName string
	switch block.SingleField() {
	case "MODULO":
		funcName = "mod"
	case "REMAINDER":
		funcName = "rem"
	case "QUOTIENT":
		funcName = "quot"
	default:
		panic("Unsupported math divide type: " + block.SingleField())
	}
	return common.MakeFuncCall(funcName, p.fromMinVals(block.Values, 2)...)
}

func (p *Parser) mathSingle(block ast.Block) ast.Expr {
	funcName := strings.ToLower(block.SingleField())
	switch funcName {
	case "ln":
		funcName = "log"
	case "ceiling":
		funcName = "ceil"
	}
	return common.MakeFuncCall(funcName, p.singleExpr(block))
}

func (p *Parser) mathOnList2(block ast.Block) ast.Expr {
	var funcName string
	switch block.SingleField() {
	case "AVG":
		funcName = "avgOf"
	case "MIN":
		funcName = "minOf"
	case "MAX":
		funcName = "maxOf"
	case "GM":
		funcName = "geoMeanOf"
	case "SD":
		funcName = "stdDevOf"
	case "SE":
		funcName = "stdErrOf"
	default:
		panic("Unsupported math on list operation: " + block.SingleField())
	}
	return common.MakeFuncCall(funcName, p.singleExpr(block))
}

func (p *Parser) mathRadix(block ast.Block) ast.Expr {
	pFields := p.makeFieldMap(block.Fields)
	var funcName string
	switch pFields["OP"] {
	case "DEC":
		funcName = "dec"
	case "BIN":
		funcName = "bin"
	case "HEX":
		funcName = "hexa"
	case "OCT":
		funcName = "octal"
	default:
		panic("Unknown Math Radix Type: " + pFields["OP"])
	}
	return common.MakeFuncCall(funcName, &fundamentals.Text{Content: pFields["NUM"]})
}

func (p *Parser) mathRandom(block ast.Block) ast.Expr {
	valMap := p.makeValueMap(block.Values)
	return common.MakeFuncCall("randInt", valMap.get("FROM"), valMap.get("TO"))
}

func (p *Parser) mathExpr(block ast.Block) ast.Expr {
	var mathOp string
	switch block.SingleField() {
	case "EQ":
		mathOp = "=="
	case "NEQ":
		mathOp = "!="
	case "LT":
		mathOp = "<"
	case "LTE":
		mathOp = "<="
	case "GT":
		mathOp = ">"
	case "GTE":
		mathOp = ">="
	case "BITAND":
		mathOp = "&"
	case "BITIOR":
		mathOp = "|"
	case "BITXOR":
		mathOp = "~"
	default:
		panic("Unsupported math expression operation: " + block.SingleField())
	}
	return p.makeBinary(mathOp, p.fromMinVals(block.Values, 2))
}

func (p *Parser) makeColor(block ast.Block) ast.Expr {
	return &fundamentals.Color{Where: lex.MakeFakeToken(lex.ColorCode), Hex: block.SingleField()}
}

func (p *Parser) makeQuestion(t lex.Type, on ast.Block, name string) ast.Expr {
	return &common.Question{Where: lex.MakeFakeToken(t), On: p.singleExpr(on), Question: name}
}

func (p *Parser) makePropCall(name string, on ast.Expr, args ...ast.Expr) ast.Expr {
	return &method.Call{Where: lex.MakeFakeToken(lex.Text), Name: name, On: on, Args: args}
}

func (p *Parser) makeBinary(operator string, operands []ast.Expr) ast.Expr {
	token := makeToken(operator)
	return &common.BinaryExpr{Where: token, Operator: token.Type, Operands: operands}
}

func makeToken(symbol string) *lex.Token {
	sToken := lex.Symbols[symbol]
	return sToken.Normal(-1, -1, nil, symbol)
}

func (p *Parser) optSingleBody(block ast.Block) []ast.Expr {
	if len(block.Statements) > 0 {
		return p.recursiveParse(*block.SingleStatement().Block)
	}
	return []ast.Expr{}
}

func (p *Parser) makeStatementMap(allStatements []ast.Statement) StatementMap {
	statementMap := make(map[string][]ast.Expr, len(allStatements))
	for _, stmt := range allStatements {
		statementMap[stmt.Name] = p.recursiveParse(*stmt.Block)
	}
	return StatementMap{statementMap: statementMap}
}

func (p *Parser) recursiveParse(currBlock ast.Block) []ast.Expr {
	var pParsed []ast.Expr
	for {
		pParsed = append(pParsed, p.parseBlock(currBlock))
		if currBlock.Next == nil {
			break
		}
		currBlock = *currBlock.Next.Block
	}
	return pParsed
}

func (p *Parser) makeFieldMap(allFields []ast.Field) map[string]string {
	fieldMap := make(map[string]string, len(allFields))
	for _, fil := range allFields {
		fieldMap[fil.Name] = fil.Value
	}
	return fieldMap
}

func (p *Parser) makeValueMap(allValues []ast.Value) ValueMap {
	valueMap := make(map[string]ast.Expr, len(allValues))
	for _, val := range allValues {
		valueMap[val.Name] = p.parseBlock(val.Block)
	}
	return ValueMap{valueMap: valueMap}
}

func (p *Parser) fromVals(allValues []ast.Value) []ast.Expr {
	arrBlocks := make([]ast.Expr, len(allValues))
	for i := range allValues {
		arrBlocks[i] = p.parseBlock(allValues[i].Block)
	}
	return arrBlocks
}

func (p *Parser) fromMinVals(allValues []ast.Value, minCount int) []ast.Expr {
	size := max(minCount, len(allValues))
	arrExprs := make([]ast.Expr, size)
	for i := range allValues {
		arrExprs[i] = p.parseBlock(allValues[i].Block)
	}
	for i := len(allValues); i < size; i++ {
		arrExprs[i] = &common.EmptySocket{}
	}
	return arrExprs
}
