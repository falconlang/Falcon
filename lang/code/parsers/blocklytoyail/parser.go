// Package blocklytoyail translates AppInventor Blockly XML into YAIL
// (the Kawa/Scheme dialect consumed by the App Inventor runtime).
//
// Each block type handled in genBlock has a 1-to-1 counterpart in the
// AppInventor blocklyeditor YAIL generators located at:
//
//	appinventor/blocklyeditor/src/generators/yail/
//
// The files in that directory are the canonical reference for what YAIL
// each block type is expected to produce.
package blocklytoyail

import (
	"Falcon/code/ast"
	"encoding/xml"
	"errors"
	"strconv"
	"strings"
	"unicode/utf16"
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
	yail, err := p.TryGenerateYAIL()
	if err != nil {
		panic(err)
	}
	return yail
}

func (p *Parser) TryGenerateYAIL() (yail string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoveredError(r)
			yail = ""
		}
	}()
	blocks := p.decodeXML()
	parts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		parts = append(parts, p.genChain(b))
	}
	return strings.Join(parts, "\n"), nil
}

func recoveredError(r any) error {
	if err, ok := r.(error); ok {
		return err
	}
	if msg, ok := r.(string); ok {
		return errors.New(msg)
	}
	return errors.New("unknown parser error")
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
		if s := p.genBlock(*curr); s != "" {
			parts = append(parts, s)
		}
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
	if b.Disabled {
		return ""
	}
	switch b.Type {
	// generators/yail/control.js
	case "controls_if":
		return p.genControlsIf(b)
	case "controls_forRange":
		loopVar := "$" + fieldByName(b, "VAR")
		start := p.vs(b.Values, "START", "0")
		end := p.vs(b.Values, "END", "0")
		step := p.vs(b.Values, "STEP", "0")
		body := p.ss(b.Statements, "DO", "#f")
		return "(forrange " + loopVar + " (begin " + body + ") " + start + " " + end + " " + step + ")"
	case "controls_forEach":
		loopVar := "$" + fieldByName(b, "VAR")
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		body := p.ss(b.Statements, "DO", "#f")
		return "(foreach " + loopVar + " (begin " + body + ") " + list + ")"
	case "controls_for_each_dict":
		loopVar := "$item"
		keyVar := "$" + fieldByName(b, "KEY")
		valVar := "$" + fieldByName(b, "VALUE")
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		body := p.ss(b.Statements, "DO", "#f")
		getKey := callPrimitive("yail-list-get-item",
			[]string{"(lexical-value " + loopVar + ")", "1"},
			[]string{"list", "number"}, "select list item")
		getVal := callPrimitive("yail-list-get-item",
			[]string{"(lexical-value " + loopVar + ")", "2"},
			[]string{"list", "number"}, "select list item")
		return "(foreach " + loopVar + " (let ((" + keyVar + " " + getKey + ")(" + valVar + " " + getVal + ")) (begin " + body + ")) " + dict + ")"
	case "controls_while":
		test := p.vs(b.Values, "TEST", "#f")
		body := p.ss(b.Statements, "DO", "#f")
		return "(while " + test + " (begin " + body + "))"
	case "controls_choose":
		test := p.vs(b.Values, "TEST", "#f")
		thenRet := p.vs(b.Values, "THENRETURN", "#f")
		elseRet := p.vs(b.Values, "ELSERETURN", "#f")
		return "(if " + test + " " + thenRet + " " + elseRet + ")"
	case "controls_do_then_return", "procedures_do_then_return":
		stm := p.ss(b.Statements, "STM", "#f")
		val := p.vs(b.Values, "VALUE", "#f")
		return "(begin " + stm + " " + val + ")"
	case "controls_eval_but_ignore":
		val := p.vs(b.Values, "VALUE", "#f")
		return "(begin " + val + " \"ignored\")"
	case "controls_break":
		return "(*yail-break* #f)"
	case "controls_nothing":
		return "*the-null-value*"
	case "controls_run_in_background":
		proc := p.vs(b.Values, "PROCEDURE", yailNull)
		cb := p.vs(b.Values, "CALLBACK", yailNull)
		return callPrimitive("run-in-background", []string{proc, cb}, []string{"any", "any"}, "run in background")
	case "controls_run_after_period":
		millis := p.vs(b.Values, "MILLIS", yailNull)
		proc := p.vs(b.Values, "PROCEDURE", yailNull)
		return callPrimitive("run-after-period", []string{millis, proc}, []string{"any", "any"}, "run after period")
	case "controls_openAnotherScreen":
		name := p.vs(b.Values, "SCREEN", "null")
		return callPrimitive("open-another-screen", []string{name}, []string{"text"}, "open another screen")
	case "controls_openAnotherScreenWithStartValue":
		screen := p.vs(b.Values, "SCREENNAME", "null")
		start := p.vs(b.Values, "STARTVALUE", "null")
		return callPrimitive("open-another-screen-with-start-value",
			[]string{screen, start}, []string{"text", "any"}, "open another screen with start value")
	case "controls_getStartValue":
		return callPrimitive("get-start-value", nil, nil, "get start value")
	case "controls_closeScreen":
		return callPrimitive("close-screen", nil, nil, "close screen")
	case "controls_closeScreenWithValue":
		val := p.vs(b.Values, "SCREEN", "null")
		return callPrimitive("close-screen-with-value", []string{val}, []string{"any"}, "close screen with value")
	case "controls_closeApplication":
		return callPrimitive("close-application", nil, nil, "close application")
	case "controls_getPlainStartText":
		return callPrimitive("get-plain-start-text", nil, nil, "get plain start text")
	case "controls_closeScreenWithPlainText":
		txt := p.vs(b.Values, "TEXT", "#f")
		return callPrimitive("close-screen-with-plain-text", []string{txt}, []string{"text"}, "close screen with plain text")

	// generators/yail/logic.js
	case "logic_boolean":
		if field(b) == "TRUE" {
			return "#t"
		}
		return "#f"
	case "logic_true":
		return "#t"
	case "logic_false":
		if fieldByName(b, "BOOL") == "TRUE" {
			return "#t"
		}
		return "#f"
	case "logic_negate":
		val := p.vs(b.Values, "BOOL", "#f")
		return callPrimitive("yail-not", []string{val}, []string{"boolean"}, "not")
	case "logic_compare":
		a := p.vs(b.Values, "A", "#f")
		bv := p.vs(b.Values, "B", "#f")
		if fieldByName(b, "OP") == "NEQ" {
			return callPrimitive("yail-not-equal?", []string{a, bv}, []string{"any", "any"}, "=")
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

	// generators/yail/text.js
	case "text":
		return quoteStr(field(b))
	case "text_join":
		n := mutItemCountOr(b, 2)
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
		mappings := p.vs(b.Values, "MAPPINGS", `""`)
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

	// generators/yail/math.js
	case "math_number":
		s := field(b)
		if s == "" {
			return "0"
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return strconv.FormatFloat(f, 'f', -1, 64)
		}
		return "+nan.0"
	case "math_number_radix":
		fm := fieldMap(b)
		num := fm["NUM"]
		if num == "" {
			return "0"
		}
		radixParse := func(base int) string {
			val, err := strconv.ParseUint(num, base, 64)
			if err != nil {
				if numErr, ok := err.(*strconv.NumError); ok && numErr.Err == strconv.ErrRange {
					return "+inf.0"
				}
				return "+nan.0"
			}
			return strconv.FormatUint(val, 10)
		}
		switch fm["OP"] {
		case "HEX":
			return radixParse(16)
		case "BIN":
			return radixParse(2)
		case "OCT":
			return radixParse(8)
		default:
			if f, err := strconv.ParseFloat(num, 64); err == nil {
				return strconv.FormatFloat(f, 'f', -1, 64)
			}
			return "+nan.0"
		}
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
		n := mutItemCountOr(b, 2)
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
		n := mutItemCountOr(b, 2)
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
		n := mutItemCountOr(b, 2)
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
	case "math_single", "math_abs", "math_neg", "math_round", "math_ceiling", "math_floor":
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
		n := mutItemCountOr(b, 2)
		var op, identityDef string
		switch fieldByName(b, "OP") {
		case "MIN":
			op, identityDef = "min", "+inf.0"
		case "MAX":
			op, identityDef = "max", "-inf.0"
		default:
			panic("blocklytoyail: unsupported math_on_list OP: " + fieldByName(b, "OP"))
		}
		if n == 0 {
			return callPrimitive(op, []string{identityDef}, []string{"number"}, op)
		}
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = p.vs(b.Values, "NUM"+strconv.Itoa(i), "0")
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

	// generators/yail/matrices.js
	case "matrices_create":
		return p.genMatrixCreate(b)
	case "matrices_create_multidim":
		dims := p.vs(b.Values, "DIM", yailNull)
		initial := p.vs(b.Values, "INITIAL", "0")
		return callPrimitive("make-yail-matrix-multidim", []string{dims, initial},
			[]string{"list", "number"}, "create multidimensional matrix")
	case "matrices_get_row":
		matrix := p.vs(b.Values, "MATRIX", yailNull)
		row := p.vs(b.Values, "ROW", "1")
		return callPrimitive("yail-matrix-get-row", []string{matrix, row},
			[]string{"matrix", "number"}, "get matrix row")
	case "matrices_get_column":
		matrix := p.vs(b.Values, "MATRIX", yailNull)
		col := p.vs(b.Values, "COLUMN", "1")
		return callPrimitive("yail-matrix-get-column", []string{matrix, col},
			[]string{"matrix", "number"}, "get matrix column")
	case "matrices_get_cell":
		return p.genMatrixCell(b, false)
	case "matrices_set_cell":
		return p.genMatrixCell(b, true)
	case "matrices_get_dims":
		matrix := p.vs(b.Values, "MATRIX", yailNull)
		return callPrimitive("yail-matrix-get-dims", []string{matrix},
			[]string{"matrix"}, "get matrix dimensions")
	case "matrices_operations", "matrices_transpose", "matrices_rotate_left", "matrices_rotate_right":
		return p.genMatrixOperation(b)
	case "matrices_subtract":
		return p.genMatrixArithmetic(b, "MINUS")
	case "matrices_power":
		return p.genMatrixArithmetic(b, "POWER")
	case "matrices_add":
		return p.genMatrixArithmeticList(b, "ADD")
	case "matrices_multiply":
		return p.genMatrixArithmeticList(b, "MULTIPLY")
	case "matrices_is_matrix":
		val := p.vs(b.Values, "VALUE", yailNull)
		return callPrimitive("yail-matrix?", []string{val}, []string{"any"}, "is matrix?")

	// generators/yail/lists.js
	case "lists_create_with":
		n := mutItemCountOr(b, 2)
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
		n := mutItemCountOr(b, 1)
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
		thing := p.vs(b.Values, "ITEM", "1")
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
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		to := p.vs(b.Values, "TO", "#f")
		return "(map_nondest " + loopVar + " " + to + " " + list + ")"
	case "lists_filter":
		loopVar := "$" + field(b)
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		test := p.vs(b.Values, "TEST", "#f")
		return "(filter_nondest " + loopVar + " " + test + " " + list + ")"
	case "lists_reduce":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		init := p.genValueSlot(b.Values, "INITANSWER")
		combine := p.vs(b.Values, "COMBINE", "#f")
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		return "(reduceovereach " + init + " " + var2 + " " + var1 + " " + combine + " " + list + ")"
	case "lists_sort_comparator":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		cmp := p.vs(b.Values, "COMPARE", "#f")
		return "(sortcomparator_nondest " + var1 + " " + var2 + " " + cmp + " " + list + ")"
	case "lists_sort_key":
		loopVar := "$" + field(b)
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		key := p.vs(b.Values, "KEY", "#f")
		return "(sortkey_nondest " + loopVar + " " + key + " " + list + ")"
	case "lists_minimum_value":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		defaultCmp := callPrimitive("<", []string{"(lexical-value " + var1 + ")", "(lexical-value " + var2 + ")"}, []string{"number", "number"}, "<")
		cmp := p.vs(b.Values, "COMPARE", defaultCmp)
		return "(mincomparator-nondest " + var1 + " " + var2 + " " + cmp + " " + list + ")"
	case "lists_maximum_value":
		fm := fieldMap(b)
		var1 := "$" + fm["VAR1"]
		var2 := "$" + fm["VAR2"]
		list := p.vs(b.Values, "LIST", yailEmptyListCode)
		defaultCmp := callPrimitive("<", []string{"(lexical-value " + var1 + ")", "(lexical-value " + var2 + ")"}, []string{"number", "number"}, "<")
		cmp := p.vs(b.Values, "COMPARE", defaultCmp)
		return "(maxcomparator-nondest " + var1 + " " + var2 + " " + cmp + " " + list + ")"

	// generators/yail/dictionaries.js
	case "pair":
		key := p.vs(b.Values, "KEY", yailNull)
		val := p.vs(b.Values, "VALUE", yailNull)
		return callPrimitive("make-dictionary-pair", []string{key, val}, []string{"key", "any"}, "make a pair")
	case "dictionaries_create_with":
		n := mutItemCountOr(b, 2)
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
		key := p.vs(b.Values, "KEY", yailNull)
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		val := p.vs(b.Values, "VALUE", yailNull)
		return callPrimitive("yail-dictionary-set-pair",
			[]string{key, dict, val}, []string{"key", "dictionary", "any"}, "set value for key")
	case "dictionaries_delete_pair":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		key := p.vs(b.Values, "KEY", yailNull)
		return callPrimitive("yail-dictionary-delete-pair",
			[]string{dict, key}, []string{"dictionary", "key"}, "delete entry for key")
	case "dictionaries_recursive_lookup":
		keys := p.vs(b.Values, "KEYS", "#f")
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		notFound := p.vs(b.Values, "NOTFOUND", yailNull)
		return callPrimitive("yail-dictionary-recursive-lookup",
			[]string{keys, dict, notFound}, []string{"list", "dictionary", "any"}, "get value at key path")
	case "dictionaries_recursive_set":
		keys := p.vs(b.Values, "KEYS", "'()")
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
		pairs := p.vs(b.Values, "PAIRS", yailEmptyDict)
		return callPrimitive("yail-dictionary-alist-to-dict", []string{pairs}, []string{"list"}, "list of pairs to dictionary")
	case "dictionaries_dict_to_alist":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		return callPrimitive("yail-dictionary-dict-to-alist", []string{dict}, []string{"dictionary"}, "dictionary to list of pairs")
	case "dictionaries_copy":
		dict := p.vs(b.Values, "DICT", yailEmptyDict)
		return callPrimitive("yail-dictionary-copy", []string{dict}, []string{"dictionary"}, "copy dictionary")
	case "dictionaries_combine_dicts":
		dict1 := p.vs(b.Values, "DICT1", "#f")
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

	// generators/yail/colors.js
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

	// generators/yail/variables.js
	case "global_declaration":
		name := "g$" + fieldByName(b, "NAME")
		val := p.vs(b.Values, "VALUE", "0")
		return "(def " + name + " " + val + ")"
	case "lexical_variable_get", "for_lexical_variable_get", "procedure_lexical_variable_get":
		return genVarGet(b)
	case "lexical_variable_set":
		name := fieldByName(b, "VAR")
		if b.Mutation != nil && len(b.Mutation.EventParams) > 0 {
			name = b.Mutation.EventParams[0].Name
		}
		val := p.vs(b.Values, "VALUE", "0")
		return genVarSet(name, val)
	case "local_declaration_statement":
		return p.genLocalDecl(b, false)
	case "local_declaration_expression":
		return p.genLocalDecl(b, true)

	// generators/yail/procedures.js
	case "procedures_defnoreturn":
		return p.genProcDef(b, false)
	case "procedures_defreturn":
		return p.genProcDef(b, true)
	case "procedures_defanonnoreturn":
		return p.genAnonProcDef(b, false)
	case "procedures_defanonreturn":
		return p.genAnonProcDef(b, true)
	case "procedures_callnoreturn", "procedures_callreturn":
		return p.genProcCall(b)
	case "procedures_callanonnoreturn", "procedures_callanonreturn":
		return p.genAnonProcCall(b)
	case "procedures_callanonnoreturn_inputlist", "procedures_callanonreturn_inputlist":
		return p.genAnonProcCallInputList(b)
	case "procedures_numArgs":
		proc := p.vs(b.Values, "PROCEDURE", "#f")
		return callPrimitive("num-args-yail-procedure", []string{proc}, []string{"any"}, "get number of arguments")
	case "procedures_getWithName":
		name := p.vs(b.Values, "PROCEDURENAME", "#f")
		return callPrimitive("create-yail-procedure-with-name", []string{name}, []string{"any"}, "get procedure")
	case "procedures_getWithDropdown":
		procName := fieldByName(b, "PROCNAME")
		if procName == "" {
			procName = field(b)
		}
		return callPrimitive("create-yail-procedure", []string{"(get-var p$" + procName + ")"},
			[]string{"any"}, "get procedure")

	// generators/yail/componentblock.js
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
		if fieldType := fieldByName(b, "COMPONENT_TYPE_SELECTOR"); fieldType != "" {
			ct = fieldType
		}
		if fieldType := fieldByName(b, "COMPONENT_SELECTOR"); fieldType != "" {
			ct = fieldType
		}
		if b.Mutation != nil {
			ct = b.Mutation.ComponentType
		}
		return "(get-all-components " + globalDB.GetFQCN(ct) + ")"

	// generators/yail/helpers.js
	case "helpers_assets", "helpers_screen_names", "helpers_provider", "helpers_providermodel":
		return quoteStr(field(b))
	case "helpers_dropdown":
		key := ""
		if b.Mutation != nil {
			key = b.Mutation.Key
		}
		return getDropdownYAIL(key, field(b))

	default:
		if strings.HasPrefix(b.Type, "color_") && len(b.Fields) > 0 {
			return genColor(b)
		}
		if strings.HasPrefix(b.Type, "helpers_") {
			return quoteStr(field(b))
		}
		panic("blocklytoyail: unsupported block type " + b.Type)
	}
}

func (p *Parser) genMatrixCreate(b ast.Block) string {
	rows := matrixDimension(b, "ROWS", 2)
	cols := matrixDimension(b, "COLS", 2)
	args := []string{strconv.Itoa(rows), strconv.Itoa(cols)}
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			value := fieldByName(b, "MATRIX_"+strconv.Itoa(i)+"_"+strconv.Itoa(j))
			if value == "" {
				value = "0"
			}
			args = append(args, value)
		}
	}
	return callPrimitive("make-yail-matrix", args, repeatStr("number", len(args)), "create a matrix")
}

func (p *Parser) genMatrixCell(b ast.Block, set bool) string {
	matrix := p.vs(b.Values, "MATRIX", yailNull)
	dimCount := mutItemCountOrValues(b, "DIM", 2)

	args := []string{matrix}
	types := []string{"matrix"}
	if set {
		args = append(args, p.vs(b.Values, "VALUE", "1"))
		types = append(types, "number")
	}
	for i := 0; i < dimCount; i++ {
		args = append(args, p.vs(b.Values, "DIM"+strconv.Itoa(i), "1"))
		types = append(types, "number")
	}
	if set {
		return callPrimitive("yail-matrix-set-cell!", args, types, "set matrix cell")
	}
	return callPrimitive("yail-matrix-get-cell", args, types, "get matrix cell")
}

func (p *Parser) genMatrixOperation(b ast.Block) string {
	mode := matrixOperationMode(b)
	operator, display := matrixOperation(mode)
	matrix := p.vs(b.Values, "MATRIX", yailNull)
	return callPrimitive(operator, []string{matrix}, []string{"matrix"}, display)
}

func (p *Parser) genMatrixArithmetic(b ast.Block, mode string) string {
	operator := matrixArithmeticOperator(mode)
	a := p.vs(b.Values, "A", "0")
	bv := p.vs(b.Values, "B", "0")
	types := []string{"matrix", "matrix"}
	if mode == "POWER" {
		types = []string{"matrix", "number"}
	} else if mode == "MULTIPLY" {
		types = []string{"matrix", "any"}
	}
	return callPrimitive(operator, []string{a, bv}, types, operator)
}

func (p *Parser) genMatrixArithmeticList(b ast.Block, mode string) string {
	operator := matrixArithmeticOperator(mode)
	n := mutItemCountOrValues(b, "MAT", 2)
	var args []string
	for i := 0; i < n; i++ {
		if arg := p.genValueSlot(b.Values, "MAT"+strconv.Itoa(i)); arg != "" {
			args = append(args, arg)
		}
	}

	if mode == "ADD" {
		if len(args) == 0 {
			return "0"
		}
		if len(args) == 1 {
			return args[0]
		}
		return callPrimitive(operator, args, repeatStr("matrix", len(args)), operator)
	}

	if n == 0 {
		return "1"
	}
	if len(args) == 0 {
		return "0"
	}
	if len(args) == 1 {
		return args[0]
	}
	types := make([]string, len(args))
	types[0] = "matrix"
	for i := 1; i < len(types); i++ {
		types[i] = "any"
	}
	return callPrimitive(operator, args, types, operator)
}

func (p *Parser) genControlsIf(b ast.Block) string {
	elseIfCount := mutElseIfCount(b)
	hasElse := mutHasElse(b)
	numConds := 1 + elseIfCount

	// Build from the innermost branch outward so each elseif is properly
	// wrapped in (begin ...) as the else-clause of its parent (if ...).
	lastIdx := numConds - 1
	lastCond := p.vs(b.Values, "IF"+strconv.Itoa(lastIdx), "#f")
	lastBody := p.ss(b.Statements, "DO"+strconv.Itoa(lastIdx), "#f")

	var result string
	if hasElse {
		elseBody := p.ss(b.Statements, "ELSE", "#f")
		result = "(if " + lastCond + "\n  (begin " + lastBody + ")\n  (begin " + elseBody + "))"
	} else {
		result = "(if " + lastCond + "\n  (begin " + lastBody + "))"
	}

	for i := numConds - 2; i >= 0; i-- {
		cond := p.vs(b.Values, "IF"+strconv.Itoa(i), "#f")
		body := p.ss(b.Statements, "DO"+strconv.Itoa(i), "#f")
		result = "(if " + cond + "\n  (begin " + body + ")\n  (begin " + result + "))"
	}
	return result
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
	return "(let (" + sb.String() + ") " + p.ss(b.Statements, "STACK", "#f") + ")"
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
	return "(def " + sig + " " + p.ss(b.Statements, "STACK", "#f") + ")"
}

func (p *Parser) genAnonProcDef(b ast.Block, hasReturn bool) string {
	params := anonProcParams(b)
	for i, param := range params {
		params[i] = "$" + param
	}
	body := p.ss(b.Statements, "STACK", "#f")
	if hasReturn {
		body = p.vs(b.Values, "RETURN", "#f")
	}
	lambda := "(lambda (" + strings.Join(params, " ") + ") " + body + ")"
	return callPrimitive("create-yail-procedure", []string{lambda}, []string{"any"}, "create procedure")
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

func (p *Parser) genAnonProcCall(b ast.Block) string {
	args := []string{p.vs(b.Values, "PROCEDURE", "#f")}
	types := []string{"any"}
	numArgs := mutItemCountOrValues(b, "ARG", 0)
	for i := 0; i < numArgs; i++ {
		args = append(args, p.vs(b.Values, "ARG"+strconv.Itoa(i), "#f"))
		types = append(types, "any")
	}
	return callPrimitive("call-yail-procedure", args, types, "call procedure")
}

func (p *Parser) genAnonProcCallInputList(b ast.Block) string {
	proc := p.vs(b.Values, "PROCEDURE", "#f")
	inputList := p.vs(b.Values, "INPUTLIST", "#f")
	return callPrimitive("call-yail-procedure-input-list", []string{proc, inputList},
		[]string{"any", "any"}, "call procedure(with input list)")
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
		return "(define-generic-event " + globalDB.GetFQCN(mut.ComponentType) + " " + mut.EventName +
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
	methodName := mut.MethodName
	if methodName == "Add" {
		if tu := fieldByName(b, "TIME_UNIT"); tu != "" {
			methodName = "Add" + tu
		}
	}

	argMeta := p.componentMethodArgs(b, methodName)
	numArgs := len(argMeta)
	args := make([]string, numArgs)
	typeList := make([]string, numArgs)
	for i, a := range argMeta {
		args[i] = p.vs(b.Values, "ARG"+strconv.Itoa(i), yailNull)
		typeList[i] = a.Type
	}
	argList := ""
	if numArgs > 0 {
		argList = " " + strings.Join(args, " ")
	}
	types := strings.Join(typeList, " ")

	if mut.IsGeneric {
		comp := p.vs(b.Values, "COMPONENT", yailNull)
		allTypes := "component"
		if numArgs > 0 {
			allTypes += " " + types
		}
		fqcn := globalDB.GetFQCN(mut.ComponentType)
		if globalDB.IsContinuation(mut.ComponentType, methodName) {
			return "(call-component-type-method-with-blocking-continuation " + comp + " '" + fqcn +
				" '" + methodName + " (*list-for-runtime*" + argList + ") '(" + allTypes + "))"
		}
		return "(call-component-type-method " + comp + " '" + fqcn +
			" '" + methodName + " (*list-for-runtime*" + argList + ") '(" + allTypes + "))"
	}
	compName := fieldByName(b, "COMPONENT_SELECTOR")
	if compName == "" {
		compName = mut.InstanceName
	}
	if globalDB.IsContinuation(mut.ComponentType, methodName) {
		return "(call-component-method-with-blocking-continuation '" + compName + " '" + methodName +
			" (*list-for-runtime*" + argList + ") '(" + types + "))"
	}
	return "(call-component-method '" + compName + " '" + methodName +
		" (*list-for-runtime*" + argList + ") '(" + types + "))"
}

func (p *Parser) componentMethodArgs(b ast.Block, methodName string) []ast.Arg {
	mut := b.Mutation
	if mut == nil {
		return nil
	}

	numArgs := len(mut.Args)
	if valueArgs := valueSlotCount(b.Values, "ARG"); valueArgs > numArgs {
		numArgs = valueArgs
	}
	if numArgs == 0 {
		return nil
	}

	args := make([]ast.Arg, numArgs)
	copy(args, mut.Args)
	dbParams := globalDB.GetMethodParams(mut.ComponentType, methodName)
	for i := range args {
		if args[i].Name == "" {
			args[i].Name = "ARG" + strconv.Itoa(i)
			if i < len(dbParams) && dbParams[i].Name != "" {
				args[i].Name = dbParams[i].Name
			}
		}
		if args[i].Type == "" {
			args[i].Type = "any"
			if i < len(dbParams) && dbParams[i].Type != "" {
				args[i].Type = dbParams[i].Type
			}
		}
	}
	return args
}

func (p *Parser) genComponentProp(b ast.Block) string {
	mut := b.Mutation
	if mut == nil {
		return ""
	}
	propName := fieldByName(b, "PROP")
	isSet := mut.SetOrGet == "set"
	if mut.IsGeneric {
		comp := p.vs(b.Values, "COMPONENT", yailNull)
		fqcn := globalDB.GetFQCN(mut.ComponentType)
		if isSet {
			val := p.vs(b.Values, "VALUE", yailNull)
			propType := globalDB.GetPropType(mut.ComponentType, propName)
			return "(set-and-coerce-property-and-check! " + comp +
				" '" + fqcn + " '" + propName + " " + val + " '" + propType + ")"
		}
		return "(get-property-and-check " + comp + " '" + fqcn + " '" + propName + ")"
	}
	compName := fieldByName(b, "COMPONENT_SELECTOR")
	if compName == "" {
		compName = mut.InstanceName
	}
	if isSet {
		val := p.vs(b.Values, "VALUE", yailNull)
		propType := globalDB.GetPropType(mut.ComponentType, propName)
		return "(set-and-coerce-property! '" + compName + " '" + propName + " " + val + " '" + propType + ")"
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
	if b.Mutation != nil && len(b.Mutation.EventParams) > 0 {
		name = b.Mutation.EventParams[0].Name
	}
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

func mutItemCountOr(b ast.Block, def int) int {
	if b.Mutation == nil {
		return def
	}
	return b.Mutation.ItemCount
}

func mutItemCountOrValues(b ast.Block, prefix string, def int) int {
	if b.Mutation != nil {
		return b.Mutation.ItemCount
	}
	if n := valueSlotCount(b.Values, prefix); n > 0 {
		return n
	}
	return def
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

func matrixDimension(b ast.Block, name string, def int) int {
	if n := positiveInt(fieldByName(b, name)); n > 0 {
		return n
	}
	if b.Mutation != nil {
		switch name {
		case "ROWS":
			if b.Mutation.Rows > 0 {
				return b.Mutation.Rows
			}
		case "COLS":
			if b.Mutation.Cols > 0 {
				return b.Mutation.Cols
			}
		}
	}
	return def
}

func positiveInt(s string) int {
	if s == "" {
		return 0
	}
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 {
		return int(f)
	}
	return 0
}

func matrixOperationMode(b ast.Block) string {
	if mode := fieldByName(b, "OP"); mode != "" {
		return mode
	}
	switch b.Type {
	case "matrices_transpose":
		return "TRANSPOSE"
	case "matrices_rotate_left":
		return "ROTATE_LEFT"
	case "matrices_rotate_right":
		return "ROTATE_RIGHT"
	default:
		return "INVERSE"
	}
}

func matrixOperation(mode string) (operator, display string) {
	switch mode {
	case "INVERSE":
		return "yail-matrix-inverse", "inverse"
	case "TRANSPOSE":
		return "yail-matrix-transpose", "transpose"
	case "ROTATE_LEFT":
		return "yail-matrix-rotate-left", "rotate_left"
	case "ROTATE_RIGHT":
		return "yail-matrix-rotate-right", "rotate_right"
	default:
		panic("blocklytoyail: unsupported matrices_operations OP: " + mode)
	}
}

func matrixArithmeticOperator(mode string) string {
	switch mode {
	case "ADD":
		return "yail-matrix-add"
	case "MINUS":
		return "yail-matrix-subtract"
	case "MULTIPLY":
		return "yail-matrix-multiply"
	case "POWER":
		return "yail-matrix-power"
	default:
		panic("blocklytoyail: unsupported matrix arithmetic OP: " + mode)
	}
}

func anonProcParams(b ast.Block) []string {
	if b.Mutation != nil && len(b.Mutation.Args) > 0 {
		params := make([]string, len(b.Mutation.Args))
		for i, arg := range b.Mutation.Args {
			params[i] = arg.Name
		}
		return params
	}
	var params []string
	for i := 0; ; i++ {
		param := fieldByName(b, "VAR"+strconv.Itoa(i))
		if param == "" {
			break
		}
		params = append(params, param)
	}
	return params
}

func valueSlotCount(values []ast.Value, prefix string) int {
	maxIdx := -1
	for _, v := range values {
		if !strings.HasPrefix(v.Name, prefix) {
			continue
		}
		idx, err := strconv.Atoi(strings.TrimPrefix(v.Name, prefix))
		if err != nil {
			continue
		}
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	return maxIdx + 1
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
	// Operate on UTF-16 code units to match JS charCodeAt/String.fromCharCode behavior.
	inputUnits := utf16.Encode([]rune(input))
	confUnits := utf16.Encode([]rune(confounder))
	for len(confUnits) < len(inputUnits) {
		confUnits = append(confUnits, confUnits...)
	}
	n := len(inputUnits)
	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := (int(inputUnits[i]) ^ int(confUnits[i])) & 0xFF
		b := (c ^ (n - i)) & 0xFF
		sb.WriteRune(rune(b))
	}
	return sb.String()
}

func quoteStr(s string) string {
	// Mirrors AI.Yail.quotifyForREPL:
	//  - backslash followed by n/t/r: pass both chars through unchanged
	//  - other backslash: double it
	//  - code < 32 or > 126: emit \uXXXX (UTF-16 code unit)
	//  - supplementary chars: emit two \uXXXX surrogate halves
	runes := []rune(s)
	n := len(runes)
	var sb strings.Builder
	sb.WriteByte('"')
	for i := 0; i < n; i++ {
		r := runes[i]
		switch {
		case r == '\\':
			if i+1 < n && (runes[i+1] == 'n' || runes[i+1] == 't' || runes[i+1] == 'r') {
				sb.WriteRune('\\')
				sb.WriteRune(runes[i+1])
				i++
			} else {
				sb.WriteString(`\\`)
			}
		case r == '"':
			sb.WriteString(`\"`)
		case r >= 32 && r <= 126:
			sb.WriteRune(r)
		case r < 0x10000:
			hex := "000" + strconv.FormatInt(int64(r), 16)
			sb.WriteString(`\u` + hex[len(hex)-4:])
		default:
			r1, r2 := utf16.EncodeRune(r)
			hi := "000" + strconv.FormatInt(int64(r1), 16)
			lo := "000" + strconv.FormatInt(int64(r2), 16)
			sb.WriteString(`\u` + hi[len(hi)-4:])
			sb.WriteString(`\u` + lo[len(lo)-4:])
		}
	}
	sb.WriteByte('"')
	return sb.String()
}
