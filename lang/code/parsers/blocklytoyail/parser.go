package blocklytoyail

import (
	"Falcon/code/ast"
	"encoding/xml"
	"strconv"
	"strings"
)

const (
	yailNull          = `(get-var *the-null-value*)`
	yailEmptyListCode = `(call-yail-primitive make-yail-list (*list-for-runtime*) '() "make a list")`
	yailEmptyYailList = `'(*list*)`
	yailEmptyDict     = `(make com.google.appinventor.components.runtime.util.YailDictionary)`
)

type Parser struct {
	xmlContent string
}

func NewParser(xmlContent string) *Parser {
	return &Parser{xmlContent: xmlContent}
}

func (p *Parser) GenerateYAIL() string {
	blocks := p.decodeXML()
	parts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		parts = append(parts, p.genChain(b))
	}
	return strings.Join(parts, "\n")
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

func (p *Parser) genChain(block ast.Block) string {
	var parts []string
	curr := &block
	for curr != nil {
		parts = append(parts, p.genBlock(*curr))
		if curr.Next == nil || curr.Next.Block == nil {
			break
		}
		curr = curr.Next.Block
	}
	return strings.Join(parts, "\n")
}

func (p *Parser) genValue(val ast.Value) string {
	if val.Block.Type != "" {
		return p.genBlock(val.Block)
	}
	if val.Shadow != nil {
		return p.genBlock(val.Shadow.BlockValue())
	}
	return ""
}

func (p *Parser) genValueSlot(values []ast.Value, name string) string {
	for _, v := range values {
		if v.Name == name {
			return p.genValue(v)
		}
	}
	return ""
}

func (p *Parser) vs(values []ast.Value, name, def string) string {
	if s := p.genValueSlot(values, name); s != "" {
		return s
	}
	return def
}

func (p *Parser) genStmt(stmts []ast.Statement, name string) string {
	for _, s := range stmts {
		if s.Name == name && s.Block != nil {
			return p.genChain(*s.Block)
		}
	}
	return ""
}

func (p *Parser) ss(stmts []ast.Statement, name, def string) string {
	if s := p.genStmt(stmts, name); s != "" {
		return s
	}
	return def
}

func (p *Parser) genBlock(b ast.Block) string {
	switch b.Type {
	case "controls_if":
		return p.genControlsIf(b)
	case "controls_forRange":
		loopVar := "$" + fieldByName(b, "VAR")
		start := p.vs(b.Values, "START", "0")
		end := p.vs(b.Values, "END", "0")
		step := p.vs(b.Values, "STEP", "0")
		body := p.ss(b.Statements, "DO", yailNull)
		return "(forrange " + loopVar + " (begin " + body + ") " + start + " " + end + " " + step + ")"
	case "controls_forEach":
		loopVar := "$" + fieldByName(b, "VAR")
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		body := p.ss(b.Statements, "DO", yailNull)
		return "(foreach " + loopVar + " (begin " + body + ") " + list + ")"
	case "controls_for_each_dict":
		loopVar := "$item"
		keyVar := "$" + fieldByName(b, "KEY")
		valVar := "$" + fieldByName(b, "VALUE")
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		body := p.ss(b.Statements, "DO", yailNull)
		getKey := callPrimitive("yail-list-get-item",
			[]string{"(lexical-value " + loopVar + ")", "1"},
			[]string{"list", "number"}, "select list item")
		getVal := callPrimitive("yail-list-get-item",
			[]string{"(lexical-value " + loopVar + ")", "2"},
			[]string{"list", "number"}, "select list item")
		return "(foreach " + loopVar + " (let ((" + keyVar + " " + getKey + ")(" + valVar + " " + getVal + ")) (begin " + body + ")) " + dict + ")"
	case "controls_while":
		test := p.vs(b.Values, "TEST", "#f")
		body := p.ss(b.Statements, "DO", yailNull)
		return "(while " + test + " (begin " + body + "))"
	case "controls_choose":
		test := p.vs(b.Values, "TEST", "#f")
		thenRet := p.vs(b.Values, "THENRETURN", "#f")
		elseRet := p.vs(b.Values, "ELSERETURN", "#f")
		return "(if " + test + " " + thenRet + " " + elseRet + ")"
	case "controls_do_then_return", "procedures_do_then_return":
		stm := p.ss(b.Statements, "STM", yailNull)
		val := p.vs(b.Values, "VALUE", "#f")
		return "(begin " + stm + " " + val + ")"
	case "controls_eval_but_ignore":
		val := p.vs(b.Values, "VALUE", "#f")
		return "(begin " + val + " \"ignored\")"
	case "controls_break":
		return "(*yail-break* #f)"
	case "controls_nothing":
		return yailNull
	case "controls_run_in_background":
		proc := p.vs(b.Values, "PROCEDURE", yailNull)
		cb := p.vs(b.Values, "CALLBACK", yailNull)
		return callPrimitive("run-in-background", []string{proc, cb}, []string{"any", "any"}, "run in background")
	case "controls_run_after_period":
		millis := p.vs(b.Values, "MILLIS", yailNull)
		proc := p.vs(b.Values, "PROCEDURE", yailNull)
		return callPrimitive("run-after-period", []string{millis, proc}, []string{"any", "any"}, "run after period")
	case "controls_openAnotherScreen":
		name := p.vs(b.Values, "SCREEN", `""`)
		return callPrimitive("open-another-screen", []string{name}, []string{"text"}, "open another screen")
	case "controls_openAnotherScreenWithStartValue":
		screen := p.vs(b.Values, "SCREENNAME", `""`)
		start := p.vs(b.Values, "STARTVALUE", "#f")
		return callPrimitive("open-another-screen-with-start-value",
			[]string{screen, start}, []string{"text", "any"}, "open another screen with start value")
	case "controls_getStartValue":
		return callPrimitive("get-start-value", nil, nil, "get start value")
	case "controls_closeScreen":
		return callPrimitive("close-screen", nil, nil, "close screen")
	case "controls_closeScreenWithValue":
		val := p.vs(b.Values, "SCREEN", "#f")
		return callPrimitive("close-screen-with-value", []string{val}, []string{"any"}, "close screen with value")
	case "controls_closeApplication":
		return callPrimitive("close-application", nil, nil, "close application")
	case "controls_getPlainStartText":
		return callPrimitive("get-plain-start-text", nil, nil, "get plain start text")
	case "controls_closeScreenWithPlainText":
		txt := p.vs(b.Values, "TEXT", `""`)
		return callPrimitive("close-screen-with-plain-text", []string{txt}, []string{"text"}, "close screen with plain text")

	case "logic_boolean":
		if field(b) == "TRUE" {
			return "#t"
		}
		return "#f"
	case "logic_true":
		return "#t"
	case "logic_false":
		return "#f"
	case "logic_negate":
		val := p.vs(b.Values, "BOOL", "#f")
		return callPrimitive("yail-not", []string{val}, []string{"boolean"}, "not")
	case "logic_compare":
		a := p.vs(b.Values, "A", "#f")
		bv := p.vs(b.Values, "B", "#f")
		if fieldByName(b, "OP") == "NEQ" {
			return callPrimitive("yail-not-equal?", []string{a, bv}, []string{"any", "any"}, "not =")
		}
		return callPrimitive("yail-equal?", []string{a, bv}, []string{"any", "any"}, "=")
	case "logic_operation", "logic_or":
		op := fieldByName(b, "OP")
		def := "#f"
		yailOp := "or-delayed"
		if op == "AND" {
			def = "#t"
			yailOp = "and-delayed"
		}
		args := []string{p.vs(b.Values, "A", def), p.vs(b.Values, "B", def)}
		for i := 2; i < mutItemCount(b); i++ {
			args = append(args, p.vs(b.Values, "BOOL"+strconv.Itoa(i), "#f"))
		}
		return "(" + yailOp + " " + strings.Join(args, " ") + ")"

	case "text":
		return quoteStr(field(b))
	case "text_join":
		n := mutItemCount(b)
		if n < 1 {
			n = len(b.Values)
		}
		args := make([]string, n)
		types := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = p.vs(b.Values, "ADD"+strconv.Itoa(i), `""`)
			types[i] = "text"
		}
		return callPrimitive("string-append", args, types, "join")
	case "text_length":
		val := p.vs(b.Values, "VALUE", `""`)
		return callPrimitive("string-length", []string{val}, []string{"text"}, "length")
	case "text_isEmpty":
		val := p.vs(b.Values, "VALUE", `""`)
		return callPrimitive("string-empty?", []string{val}, []string{"text"}, "is text empty?")
	case "text_trim":
		txt := p.vs(b.Values, "TEXT", `""`)
		return callPrimitive("string-trim", []string{txt}, []string{"text"}, "trim")
	case "text_changeCase":
		txt := p.vs(b.Values, "TEXT", `""`)
		if field(b) == "UPCASE" {
			return callPrimitive("string-to-upper-case", []string{txt}, []string{"text"}, "upcase")
		}
		return callPrimitive("string-to-lower-case", []string{txt}, []string{"text"}, "downcase")
	case "text_starts_at":
		txt := p.vs(b.Values, "TEXT", `""`)
		piece := p.vs(b.Values, "PIECE", `""`)
		return callPrimitive("string-starts-at", []string{txt, piece}, []string{"text", "text"}, "starts at")
	case "text_contains":
		txt := p.vs(b.Values, "TEXT", `""`)
		piece := p.vs(b.Values, "PIECE", `""`)
		switch field(b) {
		case "CONTAINS_ANY":
			return callPrimitive("string-contains-any", []string{txt, piece}, []string{"text", "list"}, "string contains any")
		case "CONTAINS_ALL":
			return callPrimitive("string-contains-all", []string{txt, piece}, []string{"text", "list"}, "string contains all")
		}
		return callPrimitive("string-contains", []string{txt, piece}, []string{"text", "text"}, "string contains")
	case "text_split":
		txt := p.vs(b.Values, "TEXT", `""`)
		at := p.vs(b.Values, "AT", "1")
		switch field(b) {
		case "SPLITATFIRST":
			return callPrimitive("string-split-at-first", []string{txt, at}, []string{"text", "text"}, "split at first")
		case "SPLITATFIRSTOFANY":
			return callPrimitive("string-split-at-first-of-any", []string{txt, at}, []string{"text", "list"}, "split at first of any")
		case "SPLITATANY":
			return callPrimitive("string-split-at-any", []string{txt, at}, []string{"text", "list"}, "split at any")
		}
		return callPrimitive("string-split", []string{txt, at}, []string{"text", "text"}, "split")
	case "text_split_at_spaces":
		txt := p.vs(b.Values, "TEXT", `""`)
		return callPrimitive("string-split-at-spaces", []string{txt}, []string{"text"}, "split at spaces")
	case "text_segment":
		txt := p.vs(b.Values, "TEXT", `""`)
		start := p.vs(b.Values, "START", "1")
		length := p.vs(b.Values, "LENGTH", "1")
		return callPrimitive("string-substring", []string{txt, start, length}, []string{"text", "number", "number"}, "segment")
	case "text_replace_all":
		txt := p.vs(b.Values, "TEXT", `""`)
		seg := p.vs(b.Values, "SEGMENT", `""`)
		rep := p.vs(b.Values, "REPLACEMENT", `""`)
		return callPrimitive("string-replace-all", []string{txt, seg, rep}, []string{"text", "text", "text"}, "replace all")
	case "text_compare":
		t1 := p.vs(b.Values, "TEXT1", `""`)
		t2 := p.vs(b.Values, "TEXT2", `""`)
		switch field(b) {
		case "LT":
			return callPrimitive("string<?", []string{t1, t2}, []string{"text", "text"}, "text<")
		case "GT":
			return callPrimitive("string>?", []string{t1, t2}, []string{"text", "text"}, "text>")
		case "NEQ":
			return "(not " + callPrimitive("string=?", []string{t1, t2}, []string{"text", "text"}, "not =") + ")"
		}
		return callPrimitive("string=?", []string{t1, t2}, []string{"text", "text"}, "text=")
	case "text_replace_mappings":
		txt := p.vs(b.Values, "TEXT", `""`)
		mappings := p.vs(b.Values, "MAPPINGS", yailEmptyDict)
		if field(b) == "DICTIONARY_ORDER" {
			return callPrimitive("string-replace-mappings-dictionary",
				[]string{txt, mappings}, []string{"text", "dictionary"}, "replace with mappings")
		}
		return callPrimitive("string-replace-mappings-longest-string",
			[]string{txt, mappings}, []string{"text", "dictionary"}, "replace with mappings")
	case "text_is_string":
		item := p.vs(b.Values, "ITEM", "#f")
		return callPrimitive("string?", []string{item}, []string{"any"}, "is a string?")
	case "text_reverse":
		val := p.vs(b.Values, "VALUE", `""`)
		return callPrimitive("string-reverse", []string{val}, []string{"text"}, "reverse")
	case "obfuscated_text":
		return p.genObfuscatedText(b)

	case "math_number":
		return field(b)
	case "math_number_radix":
		fm := fieldMap(b)
		num := fm["NUM"]
		var val int64
		switch fm["OP"] {
		case "HEX":
			val, _ = strconv.ParseInt(num, 16, 64)
		case "BIN":
			val, _ = strconv.ParseInt(num, 2, 64)
		case "OCT":
			val, _ = strconv.ParseInt(num, 8, 64)
		default:
			val, _ = strconv.ParseInt(num, 10, 64)
		}
		return strconv.FormatInt(val, 10)
	case "math_compare":
		a := p.vs(b.Values, "A", "0")
		bv := p.vs(b.Values, "B", "0")
		switch fieldByName(b, "OP") {
		case "NEQ":
			return callPrimitive("yail-not-equal?", []string{a, bv}, []string{"any", "any"}, "not =")
		case "EQ":
			return callPrimitive("yail-equal?", []string{a, bv}, []string{"any", "any"}, "=")
		case "LT":
			return callPrimitive("<", []string{a, bv}, []string{"number", "number"}, "<")
		case "LTE":
			return callPrimitive("<=", []string{a, bv}, []string{"number", "number"}, "<=")
		case "GT":
			return callPrimitive(">", []string{a, bv}, []string{"number", "number"}, ">")
		case "GTE":
			return callPrimitive(">=", []string{a, bv}, []string{"number", "number"}, ">=")
		}
		return callPrimitive("=", []string{a, bv}, []string{"number", "number"}, "=")
	case "math_add":
		n := mutItemCount(b)
		if n < 2 {
			n = 2
		}
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = p.vs(b.Values, "NUM"+strconv.Itoa(i), "0")
		}
		return callPrimitive("+", args, repeatStr("number", n), "+")
	case "math_subtract":
		a := p.vs(b.Values, "A", "0")
		bv := p.vs(b.Values, "B", "0")
		return callPrimitive("-", []string{a, bv}, []string{"number", "number"}, "-")
	case "math_multiply":
		n := mutItemCount(b)
		if n < 2 {
			n = 2
		}
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = p.vs(b.Values, "NUM"+strconv.Itoa(i), "0")
		}
		return callPrimitive("*", args, repeatStr("number", n), "*")
	case "math_division":
		a := p.vs(b.Values, "A", "0")
		bv := p.vs(b.Values, "B", "0")
		return callPrimitive("yail-divide", []string{a, bv}, []string{"number", "number"}, "yail-divide")
	case "math_power":
		a := p.vs(b.Values, "A", "0")
		bv := p.vs(b.Values, "B", "0")
		return callPrimitive("expt", []string{a, bv}, []string{"number", "number"}, "expt")
	case "math_bitwise":
		n := mutItemCount(b)
		if n < 2 {
			n = 2
		}
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = p.vs(b.Values, "NUM"+strconv.Itoa(i), "0")
		}
		switch field(b) {
		case "BITAND":
			return callPrimitive("bitwise-and", args, repeatStr("number", n), "bitwise-and")
		case "BITIOR":
			return callPrimitive("bitwise-ior", args, repeatStr("number", n), "bitwise-ior")
		}
		return callPrimitive("bitwise-xor", args, repeatStr("number", n), "bitwise-xor")
	case "math_single":
		val := p.vs(b.Values, "NUM", "1")
		switch field(b) {
		case "ROOT":
			return callPrimitive("sqrt", []string{val}, []string{"number"}, "sqrt")
		case "ABS":
			return callPrimitive("abs", []string{val}, []string{"number"}, "abs")
		case "NEG":
			return callPrimitive("-", []string{val}, []string{"number"}, "negate")
		case "LN":
			return callPrimitive("log", []string{val}, []string{"number"}, "log")
		case "EXP":
			return callPrimitive("exp", []string{val}, []string{"number"}, "exp")
		case "ROUND":
			return callPrimitive("yail-round", []string{val}, []string{"number"}, "round")
		case "CEILING":
			return callPrimitive("yail-ceiling", []string{val}, []string{"number"}, "ceiling")
		case "FLOOR":
			return callPrimitive("yail-floor", []string{val}, []string{"number"}, "floor")
		}
		return callPrimitive("abs", []string{val}, []string{"number"}, "abs")
	case "math_trig", "math_sin", "math_cos", "math_tan":
		val := p.vs(b.Values, "NUM", "0")
		switch field(b) {
		case "COS":
			return callPrimitive("cos-degrees", []string{val}, []string{"number"}, "cos")
		case "TAN":
			return callPrimitive("tan-degrees", []string{val}, []string{"number"}, "tan")
		case "ASIN":
			return callPrimitive("asin-degrees", []string{val}, []string{"number"}, "asin")
		case "ACOS":
			return callPrimitive("acos-degrees", []string{val}, []string{"number"}, "acos")
		case "ATAN":
			return callPrimitive("atan-degrees", []string{val}, []string{"number"}, "atan")
		}
		return callPrimitive("sin-degrees", []string{val}, []string{"number"}, "sin")
	case "math_on_list":
		n := mutItemCount(b)
		op := strings.ToLower(field(b))
		var identityDef string
		switch op {
		case "min":
			identityDef = "+inf.0"
		case "max":
			identityDef = "-inf.0"
		default:
			identityDef = "0"
		}
		if n == 0 {
			return callPrimitive(op, []string{identityDef}, []string{"number"}, op)
		}
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = p.vs(b.Values, "NUM"+strconv.Itoa(i), identityDef)
		}
		return callPrimitive(op, args, repeatStr("number", n), op)
	case "math_on_list2":
		val := p.vs(b.Values, "LIST", yailEmptyYailList)
		switch field(b) {
		case "MIN":
			return callPrimitive("minl", []string{val}, []string{"list-of-number"}, "min")
		case "MAX":
			return callPrimitive("maxl", []string{val}, []string{"list-of-number"}, "max")
		case "GM":
			return callPrimitive("gm", []string{val}, []string{"list-of-number"}, "gm")
		case "SD":
			return callPrimitive("std-dev", []string{val}, []string{"list-of-number"}, "std-dev")
		case "SE":
			return callPrimitive("std-err", []string{val}, []string{"list-of-number"}, "std-err")
		}
		return callPrimitive("avg", []string{val}, []string{"list-of-number"}, "avg")
	case "math_mode_of_list":
		val := p.vs(b.Values, "LIST", yailEmptyYailList)
		return callPrimitive("mode", []string{val}, []string{"list-of-number"}, "mode")
	case "math_atan2":
		y := p.vs(b.Values, "Y", "1")
		x := p.vs(b.Values, "X", "1")
		return callPrimitive("atan2-degrees", []string{y, x}, []string{"number", "number"}, "atan2")
	case "math_divide":
		dividend := p.vs(b.Values, "DIVIDEND", "0")
		divisor := p.vs(b.Values, "DIVISOR", "1")
		switch field(b) {
		case "MODULO":
			return callPrimitive("modulo", []string{dividend, divisor}, []string{"number", "number"}, "modulo")
		case "REMAINDER":
			return callPrimitive("remainder", []string{dividend, divisor}, []string{"number", "number"}, "remainder")
		}
		return callPrimitive("quotient", []string{dividend, divisor}, []string{"number", "number"}, "quotient")
	case "math_format_as_decimal":
		num := p.vs(b.Values, "NUM", "0")
		places := p.vs(b.Values, "PLACES", "0")
		return callPrimitive("format-as-decimal", []string{num, places}, []string{"number", "number"}, "format as decimal")
	case "math_is_a_number":
		val := p.vs(b.Values, "NUM", "#f")
		switch field(b) {
		case "BASE10":
			return callPrimitive("is-base10?", []string{val}, []string{"text"}, "is base10?")
		case "HEXADECIMAL":
			return callPrimitive("is-hexadecimal?", []string{val}, []string{"text"}, "is hexadecimal?")
		case "BINARY":
			return callPrimitive("is-binary?", []string{val}, []string{"text"}, "is binary?")
		}
		return callPrimitive("is-number?", []string{val}, []string{"text"}, "is a number?")
	case "math_convert_number":
		val := p.vs(b.Values, "NUM", "0")
		switch field(b) {
		case "HEX_TO_DEC":
			return callPrimitive("math-convert-hex-dec", []string{val}, []string{"text"}, "convert hex to dec")
		case "DEC_TO_BIN":
			return callPrimitive("math-convert-dec-bin", []string{val}, []string{"text"}, "convert dec to bin")
		case "BIN_TO_DEC":
			return callPrimitive("math-convert-bin-dec", []string{val}, []string{"text"}, "convert bin to dec")
		}
		return callPrimitive("math-convert-dec-hex", []string{val}, []string{"text"}, "convert dec to hex")
	case "math_convert_angles":
		val := p.vs(b.Values, "NUM", "0")
		if field(b) == "DEGREES_TO_RADIANS" {
			return callPrimitive("degrees->radians", []string{val}, []string{"number"}, "convert degrees to radians")
		}
		return callPrimitive("radians->degrees", []string{val}, []string{"number"}, "convert radians to degrees")
	case "math_random_int":
		from := p.vs(b.Values, "FROM", "0")
		to := p.vs(b.Values, "TO", "0")
		return callPrimitive("random-integer", []string{from, to}, []string{"number", "number"}, "random integer")
	case "math_random_float":
		return callPrimitive("random-fraction", nil, nil, "random fraction")
	case "math_random_set_seed":
		num := p.vs(b.Values, "NUM", "0")
		return callPrimitive("random-set-seed", []string{num}, []string{"number"}, "random set seed")

	case "lists_create_with":
		n := mutItemCount(b)
		var args, types []string
		for i := 0; i < n; i++ {
			s := p.genValueSlot(b.Values, "ADD"+strconv.Itoa(i))
			if s != "" {
				args = append(args, s)
				types = append(types, "any")
			}
		}
		return callPrimitive("make-yail-list", args, types, "make a list")
	case "lists_add_items":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		n := mutItemCount(b)
		args := make([]string, 1+n)
		types := make([]string, 1+n)
		args[0] = list
		types[0] = "list"
		for i := 0; i < n; i++ {
			args[1+i] = p.vs(b.Values, "ITEM"+strconv.Itoa(i), "#f")
			types[1+i] = "any"
		}
		return callPrimitive("yail-list-add-to-list!", args, types, "add items to list")
	case "lists_is_in":
		thing := p.vs(b.Values, "ITEM", "#f")
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-member?", []string{thing, list}, []string{"any", "list"}, "is in list?")
	case "lists_length":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-length", []string{list}, []string{"list"}, "length of list")
	case "lists_is_empty":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-empty?", []string{list}, []string{"list"}, "is list empty?")
	case "lists_pick_random_item":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-pick-random", []string{list}, []string{"list"}, "pick a random item")
	case "lists_position_in":
		thing := p.vs(b.Values, "ITEM", "#f")
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-index", []string{thing, list}, []string{"any", "list"}, "index in list")
	case "lists_select_item":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		idx := p.vs(b.Values, "NUM", "1")
		return callPrimitive("yail-list-get-item", []string{list, idx}, []string{"list", "number"}, "select list item")
	case "lists_insert_item":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		idx := p.vs(b.Values, "INDEX", "1")
		item := p.vs(b.Values, "ITEM", "#f")
		return callPrimitive("yail-list-insert-item!", []string{list, idx, item}, []string{"list", "number", "any"}, "insert list item")
	case "lists_replace_item":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		idx := p.vs(b.Values, "NUM", "1")
		item := p.vs(b.Values, "ITEM", "#f")
		return callPrimitive("yail-list-set-item!", []string{list, idx, item}, []string{"list", "number", "any"}, "replace list item")
	case "lists_remove_item":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		idx := p.vs(b.Values, "INDEX", "1")
		return callPrimitive("yail-list-remove-item!", []string{list, idx}, []string{"list", "number"}, "remove list item")
	case "lists_copy":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-copy", []string{list}, []string{"list"}, "copy list")
	case "lists_reverse":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-reverse", []string{list}, []string{"list"}, "reverse list")
	case "lists_to_csv_row":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-to-csv-row", []string{list}, []string{"list"}, "list to csv row")
	case "lists_to_csv_table":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-to-csv-table", []string{list}, []string{"list"}, "list to csv table")
	case "lists_from_csv_row":
		txt := p.vs(b.Values, "TEXT", `""`)
		return callPrimitive("yail-list-from-csv-row", []string{txt}, []string{"text"}, "list from csv row")
	case "lists_from_csv_table":
		txt := p.vs(b.Values, "TEXT", `""`)
		return callPrimitive("yail-list-from-csv-table", []string{txt}, []string{"text"}, "list from csv table")
	case "lists_lookup_in_pairs":
		key := p.vs(b.Values, "KEY", "#f")
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		notFound := p.vs(b.Values, "NOTFOUND", yailNull)
		return callPrimitive("yail-alist-lookup", []string{key, list, notFound}, []string{"any", "list", "any"}, "lookup in pairs")
	case "lists_join_with_separator":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		sep := p.vs(b.Values, "SEPARATOR", `""`)
		return callPrimitive("yail-list-join-with-separator", []string{list, sep}, []string{"list", "text"}, "join with separator")
	case "lists_sort":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-sort", []string{list}, []string{"list"}, "sort")
	case "lists_is_list":
		thing := p.vs(b.Values, "ITEM", "#f")
		return callPrimitive("yail-list?", []string{thing}, []string{"any"}, "is a list?")
	case "lists_but_first":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-but-first", []string{list}, []string{"list"}, "but first of list")
	case "lists_but_last":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return callPrimitive("yail-list-but-last", []string{list}, []string{"list"}, "but last of list")
	case "lists_slice":
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		idx1 := p.vs(b.Values, "INDEX1", "1")
		idx2 := p.vs(b.Values, "INDEX2", "1")
		return callPrimitive("yail-list-slice", []string{list, idx1, idx2}, []string{"list", "number", "number"}, "slice list")
	case "lists_append_list":
		list1 := p.vs(b.Values, "LIST0", yailEmptyListCode)
		list2 := p.vs(b.Values, "LIST1", yailEmptyListCode)
		return callPrimitive("yail-list-append!", []string{list1, list2}, []string{"list", "list"}, "append to list")
	case "lists_map":
		loopVar := "$" + field(b)
		list := p.vs(b.Values, "LIST", yailEmptyYailList)
		to := p.vs(b.Values, "TO", "#f")
		return "(map_nondest " + loopVar + " " + to + " " + list + ")"
	case "lists_filter":
		loopVar := "$" + field(b)
		list := p.vs(b.Values, "LIST", yailEmptyYailList)
		test := p.vs(b.Values, "TEST", "#f")
		return "(filter_nondest " + loopVar + " " + test + " " + list + ")"
	case "lists_reduce":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		init := p.vs(b.Values, "INITANSWER", "#f")
		combine := p.vs(b.Values, "COMBINE", "#f")
		list := p.vs(b.Values, "LIST", yailEmptyYailList)
		return "(reduceovereach " + init + " " + var2 + " " + var1 + " " + combine + " " + list + ")"
	case "lists_sort_comparator":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		list := p.vs(b.Values, "LIST", yailEmptyYailList)
		cmp := p.vs(b.Values, "COMPARE", "#f")
		return "(sortcomparator_nondest " + var1 + " " + var2 + " " + cmp + " " + list + ")"
	case "lists_sort_key":
		loopVar := "$" + field(b)
		list := p.vs(b.Values, "LIST", yailEmptyYailList)
		key := p.vs(b.Values, "KEY", "#f")
		return "(sortkey_nondest " + loopVar + " " + key + " " + list + ")"
	case "lists_minimum_value":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		list := p.vs(b.Values, "LIST", yailEmptyYailList)
		cmp := p.vs(b.Values, "COMPARE", "#f")
		return "(mincomparator-nondest " + var1 + " " + var2 + " " + cmp + " " + list + ")"
	case "lists_maximum_value":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		list := p.vs(b.Values, "LIST", yailEmptyYailList)
		cmp := p.vs(b.Values, "COMPARE", "#f")
		return "(maxcomparator-nondest " + var1 + " " + var2 + " " + cmp + " " + list + ")"

	case "pair":
		key := p.vs(b.Values, "KEY", "#f")
		val := p.vs(b.Values, "VALUE", "#f")
		return callPrimitive("make-dictionary-pair", []string{key, val}, []string{"key", "any"}, "make a pair")
	case "dictionaries_create_with":
		n := mutItemCount(b)
		args := make([]string, n)
		types := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = p.vs(b.Values, "ADD"+strconv.Itoa(i), yailNull)
			types[i] = "pair"
		}
		return callPrimitive("make-yail-dictionary", args, types, "make a dictionary")
	case "dictionaries_lookup":
		key := p.vs(b.Values, "KEY", "#f")
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		notFound := p.vs(b.Values, "NOTFOUND", yailNull)
		return callPrimitive("yail-dictionary-lookup",
			[]string{key, dict, notFound}, []string{"key", "dictionary", "any"}, "get value for key")
	case "dictionaries_set_pair":
		key := p.vs(b.Values, "KEY", "#f")
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		val := p.vs(b.Values, "VALUE", "#f")
		return callPrimitive("yail-dictionary-set-pair",
			[]string{key, dict, val}, []string{"key", "dictionary", "any"}, "set value for key")
	case "dictionaries_delete_pair":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		key := p.vs(b.Values, "KEY", "#f")
		return callPrimitive("yail-dictionary-delete-pair",
			[]string{dict, key}, []string{"dictionary", "key"}, "delete entry for key")
	case "dictionaries_recursive_lookup":
		keys := p.vs(b.Values, "KEYS", yailEmptyYailList)
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		notFound := p.vs(b.Values, "NOTFOUND", yailNull)
		return callPrimitive("yail-dictionary-recursive-lookup",
			[]string{keys, dict, notFound}, []string{"list", "dictionary", "any"}, "get value at key path")
	case "dictionaries_recursive_set":
		keys := p.vs(b.Values, "KEYS", yailEmptyYailList)
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		val := p.vs(b.Values, "VALUE", "#f")
		return callPrimitive("yail-dictionary-recursive-set",
			[]string{keys, dict, val}, []string{"list", "dictionary", "any"}, "set value for key path")
	case "dictionaries_getters", "dictionaries_get_values":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		if field(b) == "VALUES" {
			return callPrimitive("yail-dictionary-get-values", []string{dict}, []string{"dictionary"}, "get values")
		}
		return callPrimitive("yail-dictionary-get-keys", []string{dict}, []string{"dictionary"}, "get keys")
	case "dictionaries_is_key_in":
		key := p.vs(b.Values, "KEY", "#f")
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		return callPrimitive("yail-dictionary-is-key-in",
			[]string{key, dict}, []string{"key", "dictionary"}, "is key in dictionary?")
	case "dictionaries_length":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		return callPrimitive("yail-dictionary-length", []string{dict}, []string{"dictionary"}, "number of pairs")
	case "dictionaries_alist_to_dict":
		pairs := p.vs(b.Values, "PAIRS", yailEmptyYailList)
		return callPrimitive("yail-dictionary-alist-to-dict", []string{pairs}, []string{"list"}, "list of pairs to dictionary")
	case "dictionaries_dict_to_alist":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		return callPrimitive("yail-dictionary-dict-to-alist", []string{dict}, []string{"dictionary"}, "dictionary to list of pairs")
	case "dictionaries_copy":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		return callPrimitive("yail-dictionary-copy", []string{dict}, []string{"dictionary"}, "copy dictionary")
	case "dictionaries_combine_dicts":
		dict1 := p.vs(b.Values, "DICT1", yailEmptyDict)
		dict2 := p.vs(b.Values, "DICT2", yailEmptyDict)
		return callPrimitive("yail-dictionary-combine-dicts",
			[]string{dict1, dict2}, []string{"dictionary", "dictionary"}, "combine 2 dictionaries")
	case "dictionaries_walk_tree":
		path := p.vs(b.Values, "PATH", yailEmptyYailList)
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		return callPrimitive("yail-dictionary-walk",
			[]string{path, dict}, []string{"list", "any"}, "list by walking key path in dictionary")
	case "dictionaries_walk_all":
		return "(static-field com.google.appinventor.components.runtime.util.YailDictionary 'ALL)"
	case "dictionaries_is_dict":
		thing := p.vs(b.Values, "THING", yailEmptyDict)
		return callPrimitive("yail-dictionary?", []string{thing}, []string{"any"}, "is a dictionary?")

	case "color_black", "color_white", "color_red", "color_pink", "color_orange",
		"color_yellow", "color_green", "color_cyan", "color_blue", "color_magenta",
		"color_light_gray", "color_dark_gray", "color_gray", "color_light_green":
		return genColor(b)
	case "color_make_color":
		colorList := p.vs(b.Values, "COLORLIST",
			`(call-yail-primitive make-yail-list (*list-for-runtime* 0 0 0) '(any any any) "make a list")`)
		return callPrimitive("make-color", []string{colorList}, []string{"list"}, "make-color")
	case "color_split_color":
		color := p.vs(b.Values, "COLOR", "-1")
		return callPrimitive("split-color", []string{color}, []string{"number"}, "split-color")

	case "global_declaration":
		name := "g$" + fieldByName(b, "NAME")
		val := p.vs(b.Values, "VALUE", "0")
		return "(def " + name + " " + val + ")"
	case "lexical_variable_get", "for_lexical_variable_get", "procedure_lexical_variable_get":
		return genVarGet(b)
	case "lexical_variable_set":
		name := fieldByName(b, "VAR")
		val := p.vs(b.Values, "VALUE", "0")
		return genVarSet(name, val)
	case "local_declaration_statement":
		return p.genLocalDecl(b, false)
	case "local_declaration_expression":
		return p.genLocalDecl(b, true)

	case "procedures_defnoreturn":
		return p.genProcDef(b, false)
	case "procedures_defreturn":
		return p.genProcDef(b, true)
	case "procedures_callnoreturn", "procedures_callreturn":
		return p.genProcCall(b)

	case "component_event":
		return p.genComponentEvent(b)
	case "component_method":
		return p.genComponentMethod(b)
	case "component_set_get":
		return p.genComponentProp(b)
	case "component_component_block":
		return "(get-component " + fieldByName(b, "COMPONENT_SELECTOR") + ")"
	case "component_all_component_block":
		ct := ""
		if b.Mutation != nil {
			ct = b.Mutation.ComponentType
		}
		return "(get-all-components " + ct + ")"

	case "helpers_assets", "helpers_screen_names", "helpers_provider", "helpers_providermodel":
		return quoteStr(field(b))
	case "helpers_dropdown":
		key := ""
		if b.Mutation != nil {
			key = b.Mutation.Key
		}
		return "(static-field " + key + " \"" + field(b) + "\")"

	default:
		if strings.HasPrefix(b.Type, "color_") && len(b.Fields) > 0 {
			return genColor(b)
		}
		if strings.HasPrefix(b.Type, "helpers_") {
			return quoteStr(field(b))
		}
		return "; unsupported block: " + b.Type
	}
}

func (p *Parser) genControlsIf(b ast.Block) string {
	elseIfCount := mutElseIfCount(b)
	hasElse := mutHasElse(b)
	numConds := 1 + elseIfCount

	var sb strings.Builder
	for i := 0; i < numConds; i++ {
		cond := p.vs(b.Values, "IF"+strconv.Itoa(i), "#f")
		body := p.ss(b.Statements, "DO"+strconv.Itoa(i), yailNull)
		if i < numConds-1 || hasElse {
			sb.WriteString("(if " + cond + "\n  (begin " + body + ")\n")
		} else {
			sb.WriteString("(if " + cond + "\n  (begin " + body + ")")
		}
	}
	if hasElse {
		sb.WriteString("  (begin " + p.ss(b.Statements, "ELSE", yailNull) + ")")
	}
	for i := 0; i < numConds; i++ {
		sb.WriteByte(')')
	}
	return sb.String()
}

func (p *Parser) genLocalDecl(b ast.Block, isExpr bool) string {
	n := 0
	if b.Mutation != nil {
		n = len(b.Mutation.LocalNames)
	}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			sb.WriteByte(' ')
		}
		varName := fieldByName(b, "VAR"+strconv.Itoa(i))
		decl := p.vs(b.Values, "DECL"+strconv.Itoa(i), "0")
		sb.WriteString("($" + varName + " " + decl + ")")
	}
	if isExpr {
		return "(let (" + sb.String() + ") " + p.vs(b.Values, "RETURN", "0") + ")"
	}
	return "(let (" + sb.String() + ") (begin " + p.ss(b.Statements, "STACK", yailNull) + "))"
}

func (p *Parser) genProcDef(b ast.Block, hasReturn bool) string {
	name := "p$" + fieldByName(b, "NAME")
	params := make([]string, 0)
	if b.Mutation != nil {
		for _, a := range b.Mutation.Args {
			params = append(params, "$"+a.Name)
		}
	}
	sig := "(" + name + " " + strings.Join(params, " ") + ")"
	if hasReturn {
		return "(def " + sig + " " + p.vs(b.Values, "RETURN", "#f") + ")"
	}
	return "(def " + sig + " (begin " + p.ss(b.Statements, "STACK", yailNull) + "))"
}

func (p *Parser) genProcCall(b ast.Block) string {
	name := "p$" + field(b)
	numArgs := 0
	if b.Mutation != nil {
		numArgs = len(b.Mutation.Args)
	}
	if numArgs == 0 {
		return "((get-var " + name + "))"
	}
	args := make([]string, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = p.vs(b.Values, "ARG"+strconv.Itoa(i), "#f")
	}
	return "((get-var " + name + ") " + strings.Join(args, " ") + ")"
}

func (p *Parser) genComponentEvent(b ast.Block) string {
	mut := b.Mutation
	if mut == nil {
		return ""
	}
	params := make([]string, len(mut.Args))
	for i, a := range mut.Args {
		params[i] = "$" + a.Name
	}
	paramStr := strings.Join(params, " ")
	body := p.ss(b.Statements, "DO", yailNull)
	if mut.IsGeneric {
		return "(define-generic-event " + mut.ComponentType + " " + mut.EventName +
			" (" + paramStr + ")\n  (set-this-form)\n  " + body + ")"
	}
	compName := fieldByName(b, "COMPONENT_SELECTOR")
	if compName == "" {
		compName = mut.InstanceName
	}
	return "(define-event " + compName + " " + mut.EventName +
		" (" + paramStr + ")\n  (set-this-form)\n  " + body + ")"
}

func (p *Parser) genComponentMethod(b ast.Block) string {
	mut := b.Mutation
	if mut == nil {
		return ""
	}
	var args []string
	for i := 0; ; i++ {
		slot := "ARG" + strconv.Itoa(i)
		found := false
		for _, v := range b.Values {
			if v.Name == slot {
				found = true
				s := p.genValue(v)
				if s == "" {
					s = "#f"
				}
				args = append(args, s)
				break
			}
		}
		if !found {
			break
		}
	}
	argList := ""
	if len(args) > 0 {
		argList = " " + strings.Join(args, " ")
	}
	types := strings.Join(repeatStr("any", len(args)), " ")
	if mut.IsGeneric {
		comp := p.genValueSlot(b.Values, "COMPONENT")
		allTypes := "component"
		if len(args) > 0 {
			allTypes += " " + types
		}
		return "(call-component-type-method " + comp + " '" + mut.ComponentType +
			" '" + mut.MethodName + " (*list-for-runtime*" + argList + ") '(" + allTypes + "))"
	}
	compName := fieldByName(b, "COMPONENT_SELECTOR")
	if compName == "" {
		compName = mut.InstanceName
	}
	return "(call-component-method '" + compName + " '" + mut.MethodName +
		" (*list-for-runtime*" + argList + ") '(" + types + "))"
}

func (p *Parser) genComponentProp(b ast.Block) string {
	mut := b.Mutation
	if mut == nil {
		return ""
	}
	propName := fieldByName(b, "PROP")
	isSet := mut.SetOrGet == "set"
	if mut.IsGeneric {
		comp := p.genValueSlot(b.Values, "COMPONENT")
		if isSet {
			val := p.vs(b.Values, "VALUE", "#f")
			return "(set-and-coerce-property-and-check! " + comp +
				" '" + mut.ComponentType + " '" + propName + " " + val + " 'any)"
		}
		return "(get-property-and-check " + comp + " '" + mut.ComponentType + " '" + propName + ")"
	}
	compName := fieldByName(b, "COMPONENT_SELECTOR")
	if compName == "" {
		compName = mut.InstanceName
	}
	if isSet {
		val := p.vs(b.Values, "VALUE", "#f")
		return "(set-and-coerce-property! '" + compName + " '" + propName + " " + val + " 'any)"
	}
	return "(get-property '" + compName + " '" + propName + ")"
}

func (p *Parser) genObfuscatedText(b ast.Block) string {
	text := field(b)
	confounder := "aaa"
	if b.Mutation != nil && b.Mutation.Cofounder != "" {
		confounder = b.Mutation.Cofounder
	}
	return callPrimitive("text-deobfuscate",
		[]string{quoteStr(obfuscate(text, confounder)), quoteStr(confounder)},
		[]string{"text", "text"}, "deobfuscate text")
}

func genColor(b ast.Block) string {
	hex := field(b)
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	n, _ := strconv.ParseInt(hex, 16, 64)
	return strconv.FormatInt(n-16777216, 10)
}

func genVarGet(b ast.Block) string {
	name := field(b)
	if strings.HasPrefix(name, "global ") {
		return "(get-var g$" + name[7:] + ")"
	}
	return "(lexical-value $" + name + ")"
}

func genVarSet(name, val string) string {
	if strings.HasPrefix(name, "global ") {
		return "(set-var! g$" + name[7:] + " " + val + ")"
	}
	return "(set-lexical! $" + name + " " + val + ")"
}

func mutItemCount(b ast.Block) int {
	if b.Mutation != nil {
		return b.Mutation.ItemCount
	}
	return 0
}

func mutElseIfCount(b ast.Block) int {
	if b.Mutation != nil {
		return b.Mutation.ElseIfCount
	}
	return 0
}

func mutHasElse(b ast.Block) bool {
	return b.Mutation != nil && b.Mutation.ElseCount > 0
}

func field(b ast.Block) string {
	if len(b.Fields) > 0 {
		return b.Fields[0].Value
	}
	return ""
}

func fieldByName(b ast.Block, name string) string {
	for _, f := range b.Fields {
		if f.Name == name {
			return f.Value
		}
	}
	return ""
}

func fieldMap(b ast.Block) map[string]string {
	m := make(map[string]string, len(b.Fields))
	for _, f := range b.Fields {
		m[f.Name] = f.Value
	}
	return m
}

func repeatStr(s string, n int) []string {
	r := make([]string, n)
	for i := range r {
		r[i] = s
	}
	return r
}

func callPrimitive(name string, args, types []string, display string) string {
	argStr := ""
	if len(args) > 0 {
		argStr = " " + strings.Join(args, " ")
	}
	return "(call-yail-primitive " + name + " (*list-for-runtime*" + argStr + ") '(" +
		strings.Join(types, " ") + `) "` + display + `")`
}

func obfuscate(input, confounder string) string {
	for len(confounder) < len(input) {
		confounder += confounder
	}
	n := len(input)
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		c := (int(input[i]) ^ int(confounder[i])) & 0xFF
		result[i] = byte((c ^ (n - i)) & 0xFF)
	}
	return string(result)
}

func quoteStr(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			sb.WriteString(`\"`)
		case c == '\\':
			sb.WriteString(`\\`)
		case c >= 32 && c <= 126:
			sb.WriteByte(c)
		default:
			hex := "000" + strconv.FormatInt(int64(c), 16)
			sb.WriteString(`\u` + hex[len(hex)-4:])
		}
	}
	sb.WriteByte('"')
	return sb.String()
}
