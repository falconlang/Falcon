package common

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/variables"
	"Falcon/code/lex"
	"Falcon/code/sugar"
	"strconv"
)

type FuncCall struct {
	Where *lex.Token
	Name  string
	Args  []ast.Expr
}

type FuncCallSignature struct {
	Name       string
	ParamCount int
	Signature  ast.Signature
}

func makeSignature(name string, paramCount int, signature ast.Signature) *FuncCallSignature {
	return &FuncCallSignature{Name: name, ParamCount: paramCount, Signature: signature}
}

var signatures = map[string]*FuncCallSignature{
	"sqrt":     makeSignature("sqrt", 1, ast.SignNumb),
	"abs":      makeSignature("abs", 1, ast.SignNumb),
	"neg":      makeSignature("neg", 1, ast.SignNumb),
	"log":      makeSignature("log", 1, ast.SignNumb),
	"exp":      makeSignature("exp", 1, ast.SignNumb),
	"round":    makeSignature("round", 1, ast.SignNumb),
	"ceil":     makeSignature("ceil", 1, ast.SignNumb),
	"floor":    makeSignature("floor", 1, ast.SignNumb),
	"sin":      makeSignature("sin", 1, ast.SignNumb),
	"cos":      makeSignature("cos", 1, ast.SignNumb),
	"tan":      makeSignature("tan", 1, ast.SignNumb),
	"asin":     makeSignature("asin", 1, ast.SignNumb),
	"acos":     makeSignature("acos", 1, ast.SignNumb),
	"atan":     makeSignature("atan", 1, ast.SignNumb),
	"degrees":  makeSignature("degrees", 1, ast.SignNumb),
	"radians":  makeSignature("radians", 1, ast.SignNumb),
	"decToHex": makeSignature("decToHex", 1, ast.SignNumb),
	"decToBin": makeSignature("decToBin", 1, ast.SignNumb),
	"hexToDec": makeSignature("hexToDec", 1, ast.SignNumb),
	"binToDec": makeSignature("binToDec", 1, ast.SignNumb),

	"dec":         makeSignature("dec", 1, ast.SignNumb),
	"bin":         makeSignature("bin", 1, ast.SignNumb),
	"octal":       makeSignature("octal", 1, ast.SignNumb),
	"hexa":        makeSignature("hexa", 1, ast.SignNumb),
	"randInt":     makeSignature("randInt", 2, ast.SignNumb),
	"randFloat":   makeSignature("randFloat", 0, ast.SignNumb),
	"setRandSeed": makeSignature("setRandSeed", 1, ast.SignVoid),
	"min":         makeSignature("min", -1, ast.SignNumb),
	"max":         makeSignature("max", -1, ast.SignNumb),

	"avgOf":     makeSignature("avgOf", 1, ast.SignNumb),
	"maxOf":     makeSignature("maxOf", 1, ast.SignNumb),
	"minOf":     makeSignature("minOf", 1, ast.SignNumb),
	"geoMeanOf": makeSignature("geoMeanOf", 1, ast.SignNumb),
	"stdDevOf":  makeSignature("stdDevOf", 1, ast.SignNumb),
	"stdErrOf":  makeSignature("stdErrOf", 1, ast.SignNumb),

	"println":              makeSignature("println", 1, ast.SignVoid),
	"openScreen":           makeSignature("openScreen", 1, ast.SignVoid),
	"openScreenWithValue":  makeSignature("openScreenWithValue", 2, ast.SignVoid),
	"closeScreenWithValue": makeSignature("closeScreenWithValue", 1, ast.SignVoid),
	"getStartValue":        makeSignature("getStartValue", 0, ast.SignText),

	"closeScreen":              makeSignature("closeScreen", 0, ast.SignVoid),
	"closeApp":                 makeSignature("closeApp", 0, ast.SignVoid),
	"getPlainStartText":        makeSignature("getPlainStartText", 0, ast.SignText),
	"closeScreenWithPlainText": makeSignature("closeScreenWithPlainText", 1, ast.SignVoid),

	"copyList":   makeSignature("copyList", 1, ast.SignList),
	"copyDict":   makeSignature("copyDict", 1, ast.SignDict),
	"makeColor":  makeSignature("makeColor", 1, ast.SignNumb),
	"splitColor": makeSignature("splitColor", 1, ast.SignList),

	"set":   makeSignature("set", 4, ast.SignVoid),
	"get":   makeSignature("get", 3, ast.SignAny),
	"call":  makeSignature("call", -1-(3), ast.SignVoid),
	"vcall": makeSignature("vcall", -1-(3), ast.SignAny),
	"every": makeSignature("every", 1, ast.SignAny),
}

func TestSignature(funcName string, argsCount int) (string, *FuncCallSignature) {
	callSignature, ok := signatures[funcName]
	if !ok {
		return sugar.Format("Cannot find function .%()", funcName), nil
	}
	if callSignature.ParamCount == -1 {
		if argsCount == 0 {
			return sugar.Format("Expected a positive number of args for function %()", funcName), nil
		}
	} else if callSignature.ParamCount >= 0 {
		if argsCount != callSignature.ParamCount {
			return sugar.Format("Expected % args but got % for function %()",
				strconv.Itoa(callSignature.ParamCount), strconv.Itoa(argsCount), funcName), nil
		}
	} else {
		minArgs := -callSignature.ParamCount - 1 // -1 offset
		if argsCount < minArgs {
			return sugar.Format("Expected at least % args but got only % for function %()",
				strconv.Itoa(minArgs), strconv.Itoa(argsCount), funcName), nil
		}
	}
	return "", callSignature
}

func (f *FuncCall) String() string {
	if f.Name == "rem" {
		return f.Args[0].String() + " % " + f.Args[1].String()
	}
	if f.Name == "neg" {
		if f.Args[0].Continuous() {
			return "-" + f.Args[0].String()
		}
		return "-(" + f.Args[0].String() + ")"
	}
	return sugar.Format("%(%)", f.Name, ast.JoinExprs(", ", f.Args))
}

func (f *FuncCall) Blockly(flags ...bool) ast.Block {
	errorMessage, signature := TestSignature(f.Name, len(f.Args))
	if signature == nil {
		panic(errorMessage)
	}
	if len(flags) > 0 && !flags[0] && !f.Consumable() {
		f.Where.Error("Expected a consumable but got a statement")
	}
	switch f.Name {
	case "sqrt", "abs", "neg", "log", "exp", "round", "ceil", "floor",
		"sin", "cos", "tan", "asin", "acos", "atan", "degrees", "radians",
		"decToHex", "decToBin", "hexToDec", "binToDec":
		return f.mathConversions()

	case "dec", "bin", "octal", "hexa":
		return f.mathRadix()
	case "randInt":
		return f.randInt()
	case "randFloat":
		return f.randFloat()
	case "setRandSeed":
		return f.setRandSeed()
	case "min", "max":
		return f.minOrMax()
	case "avgOf", "maxOf", "minOf", "geoMeanOf", "stdDevOf", "stdErrOf":
		return f.mathOnList()
	case "modeOf":
		return f.modeOf()
	case "mod", "rem", "quot":
		return f.mathDivide()
	case "aTan2":
		return f.atan2()
	case "formatDecimal":
		return f.formatDecimal()

	case "println":
		return f.println()
	case "openScreen":
		return f.openScreen()
	case "openScreenWithValue":
		return f.openScreenWithValue()
	case "closeScreenWithValue":
		return f.closeScreenWithValue()
	case "getStartValue":
		return f.ctrlSimpleBlock("controls_getStartValue")
	case "closeScreen":
		return f.ctrlSimpleBlock("controls_closeScreen")
	case "closeApp":
		return f.ctrlSimpleBlock("controls_closeApplication")
	case "getPlainStartText":
		return f.ctrlSimpleBlock("controls_getPlainStartText")
	case "closeScreenWithPlainText":
		return f.closeScreenWithPlainText()
	case "copyList":
		return f.copyList()
	case "copyDict":
		return f.copyDict()

	case "makeColor":
		return f.makeColor()
	case "splitColor":
		return f.splitColor()

	case "set":
		return f.genericSet()
	case "get":
		return f.genericGet()
	case "call":
		return f.genericCall(false)
	case "vcall":
		return f.genericCall(true)
	case "every":
		return f.everyComponent()
	default:
		f.Where.Error("Cannot find %()", f.Name)
		panic("never reached")
	}
}

func (f *FuncCall) Continuous() bool {
	return true
}

func (f *FuncCall) Consumable(flags ...bool) bool {
	if f.Name == "setRandSeed" || f.Name == "println" ||
		f.Name == "openScreen" || f.Name == "openScreenWithValue" ||
		f.Name == "closeScreen" || f.Name == "closeScreenWithValue" ||
		f.Name == "closeApp" || f.Name == "closeScreenWithPlainText" ||
		f.Name == "set" || f.Name == "call" {
		return false
	}
	return true
}

func (f *FuncCall) Signature() []ast.Signature {
	errorMessage, signature := TestSignature(f.Name, len(f.Args)) // signatures are already verified
	if signature == nil {
		panic(errorMessage)
	}
	return []ast.Signature{signature.Signature}
}

func (f *FuncCall) everyComponent() ast.Block {
	compType, ok := f.Args[0].(*variables.Get)
	if !ok || compType.Global {
		f.Where.Error("Expected a component type for every() 1st argument!")
	}
	return ast.Block{
		Type:     "component_all_component_block",
		Mutation: &ast.Mutation{ComponentType: compType.Name},
		Fields:   []ast.Field{{Name: "COMPONENT_SELECTOR", Value: compType.Name}},
	}
}

func (f *FuncCall) genericCall(vcall bool) ast.Block {
	// arg[0] 	 compType
	// arg[1] 	 component (any object)
	// arg[2] 	 method name
	// arg[4->n] invoke args
	// if {vcall}, it is a returning method
	compType, ok := f.Args[0].(*fundamentals.Text)
	if !ok {
		f.Where.Error("Expected a component type for call() 1st argument!")
	}
	vGet, ok := f.Args[2].(*fundamentals.Text)
	if !ok {
		f.Where.Error("Expected a method name for call() 3rd argument!")
	}
	var shape string
	if vcall {
		shape = "value"
	} else {
		shape = "statement"
	}
	return ast.Block{
		Type: "component_method",
		Mutation: &ast.Mutation{
			MethodName:    vGet.Content,
			IsGeneric:     true,
			ComponentType: compType.Content,
			Shape:         shape,
		},
		Values: ast.ValueArgsByPrefix(f.Args[1], "COMPONENT", "ARG", f.Args[3:]),
	}
}

func (f *FuncCall) genericGet() ast.Block {
	compType, ok := f.Args[0].(*fundamentals.Text)
	if !ok {
		f.Where.Error("Expected a component type for get() 1st argument!")
	}
	vGet, ok := f.Args[2].(*fundamentals.Text)
	if !ok {
		f.Where.Error("Expected a property type for get() 3rd argument!")
	}
	return ast.Block{
		Type: "component_set_get",
		Mutation: &ast.Mutation{
			SetOrGet:      "get",
			PropertyName:  vGet.Content,
			IsGeneric:     true,
			ComponentType: compType.Content,
		},
		Fields: []ast.Field{{Name: "PROP", Value: vGet.Content}},
		Values: []ast.Value{{Name: "COMPONENT", Block: f.Args[1].Blockly(false)}},
	}
}

func (f *FuncCall) genericSet() ast.Block {
	compType, ok := f.Args[0].(*fundamentals.Text)
	if !ok {
		f.Where.Error("Expected a component type for set() 1st argument!")
	}
	propName, ok := f.Args[2].(*fundamentals.Text)
	if !ok {
		f.Where.Error("Expected a property type for set() 3rd argument!")
	}
	return ast.Block{
		Type: "component_set_get",
		Mutation: &ast.Mutation{
			SetOrGet:      "set",
			PropertyName:  propName.Content,
			IsGeneric:     true,
			ComponentType: compType.Content,
		},
		Fields: []ast.Field{{Name: "PROP", Value: propName.Content}},
		Values: ast.MakeValues([]ast.Expr{f.Args[1], f.Args[3]}, "COMPONENT", "VALUE"),
	}
}

func (f *FuncCall) splitColor() ast.Block {
	return ast.Block{
		Type:   "color_make_color",
		Values: ast.MakeValues(f.Args, "COLOR"),
	}
}

func (f *FuncCall) makeColor() ast.Block {
	return ast.Block{
		Type:   "color_make_color",
		Values: ast.MakeValues(f.Args, "COLORLIST"),
	}
}

func (f *FuncCall) copyDict() ast.Block {
	return ast.Block{
		Type:   "dictionaries_copy",
		Values: ast.MakeValues(f.Args, "DICT"),
	}
}

func (f *FuncCall) copyList() ast.Block {
	return ast.Block{
		Type:   "lists_copy",
		Values: ast.MakeValues(f.Args, "LIST"),
	}
}

func (f *FuncCall) ctrlSimpleBlock(blockType string) ast.Block {
	return ast.Block{Type: blockType}
}

func (f *FuncCall) closeScreenWithPlainText() ast.Block {
	return ast.Block{
		Type:   "controls_closeScreenWithPlainText",
		Values: ast.MakeValues(f.Args, "TEXT"),
	}
}

func (f *FuncCall) closeScreenWithValue() ast.Block {
	return ast.Block{
		Type:   "controls_closeScreenWithValue",
		Values: ast.MakeValues(f.Args, "SCREEN"),
	}
}

func (f *FuncCall) openScreenWithValue() ast.Block {
	return ast.Block{
		Type:   "controls_openAnotherScreenWithStartValue",
		Values: ast.MakeValues(f.Args, "SCREENNAME", "STARTVALUE"),
	}
}

func (f *FuncCall) openScreen() ast.Block {
	return ast.Block{
		Type:   "controls_openAnotherScreen",
		Values: ast.MakeValues(f.Args, "SCREEN"),
	}
}

func (f *FuncCall) println() ast.Block {
	return ast.Block{Type: "controls_eval_but_ignore", Values: ast.MakeValues(f.Args, "VALUE")}
}

var mathFuncMap = map[string]string{
	"sqrt":     "ROOT",
	"abs":      "ABS",
	"neg":      "NEG",
	"log":      "LN",
	"exp":      "EXP",
	"round":    "ROUND",
	"ceil":     "CEILING",
	"floor":    "FLOOR",
	"sin":      "SIN",
	"cos":      "COS",
	"tan":      "TAN",
	"asin":     "ASIN",
	"acos":     "ACOS",
	"atan":     "ATAN",
	"degrees":  "RADIANS_TO_DEGREES",
	"radians":  "DEGREES_TO_RADIANS",
	"decToHex": "DEC_TO_HEX",
	"decToBin": "DEC_TO_BIN",
	"hexToDec": "HEX_TO_DEC",
	"binToDec": "BIN_TO_DEC",
}

func (f *FuncCall) mathConversions() ast.Block {
	fieldOp, ok := mathFuncMap[f.Name]
	if !ok {
		f.Where.Error("Unknown Math Conversion %()", f.Name)
	}
	var blockType string
	switch fieldOp {
	case "SIN", "COS", "TAN", "ASIN", "ACOS", "ATAN":
		blockType = "math_trig"
	case "RADIANS_TO_DEGREES", "DEGREES_TO_RADIANS":
		blockType = "math_convert_angles"
	case "DEC_TO_HEX", "HEX_TO_DEC", "DEC_TO_BIN", "BIN_TO_DEC":
		blockType = "math_convert_number"
	default:
		blockType = "math_single"
	}
	return ast.Block{
		Type:   blockType,
		Fields: []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: []ast.Value{{Name: "NUM", Block: f.Args[0].Blockly(false)}},
	}
}

func (f *FuncCall) formatDecimal() ast.Block {
	return ast.Block{
		Type:   "math_format_as_decimal",
		Values: ast.MakeValues(f.Args, "NUM", "PLACES"),
	}
}

func (f *FuncCall) atan2() ast.Block {
	return ast.Block{
		Type:   "math_atan2",
		Values: ast.MakeValues(f.Args, "Y", "X"),
	}
}

func (f *FuncCall) mathDivide() ast.Block {
	var fieldOp string
	switch f.Name {
	case "mod":
		fieldOp = "MODULO"
	case "rem":
		fieldOp = "REMAINDER"
	case "quot":
		fieldOp = "QUOTIENT"
	}
	return ast.Block{
		Type:   "math_divide",
		Fields: []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: ast.MakeValues(f.Args, "DIVIDEND", "DIVISOR"),
	}
}

func (f *FuncCall) modeOf() ast.Block {
	return ast.Block{
		Type:   "math_mode_of_list",
		Values: ast.MakeValues(f.Args, "LIST"),
	}
}

func (f *FuncCall) mathOnList() ast.Block {
	var fieldOp string
	switch f.Name {
	case "avgOf":
		fieldOp = "AVG"
	case "maxOf":
		fieldOp = "MAX"
	case "minOf":
		fieldOp = "MIN"
	case "geoMeanOf":
		fieldOp = "GM"
	case "stdDevOf":
		fieldOp = "SD"
	case "stdErrOf":
		fieldOp = "SE"
	}
	return ast.Block{
		Type:   "math_on_list2",
		Fields: []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: ast.MakeValues(f.Args, "LIST"),
	}
}

func (f *FuncCall) minOrMax() ast.Block {
	argSize := len(f.Args)
	if argSize == 0 {
		f.Where.Error("No arguments provided for %()", f.Name)
	}
	var fieldOp string
	switch f.Name {
	case "min":
		fieldOp = "MIN"
	case "max":
		fieldOp = "MAX"
	}
	return ast.Block{
		Type:     "math_on_list",
		Fields:   []ast.Field{{Name: "OP", Value: fieldOp}},
		Mutation: &ast.Mutation{ItemCount: argSize},
		Values:   ast.ValuesByPrefix("NUM", f.Args),
	}
}

func (f *FuncCall) setRandSeed() ast.Block {
	return ast.Block{
		Type:   "math_random_set_seed",
		Values: ast.MakeValues(f.Args, "NUM"),
	}
}

func (f *FuncCall) randFloat() ast.Block {
	return ast.Block{Type: "math_random_float"}
}

func (f *FuncCall) randInt() ast.Block {
	return ast.Block{
		Type:   "math_random_int",
		Values: ast.MakeValues(f.Args, "FROM", "TO"),
	}
}

func (f *FuncCall) mathRadix() ast.Block {
	var fieldOp string
	switch f.Name {
	case "dec":
		fieldOp = "DEC"
	case "bin":
		fieldOp = "BIN"
	case "octal":
		fieldOp = "OCT"
	case "hexa":
		fieldOp = "HEX"
	}
	textExpr, ok := f.Args[0].(*fundamentals.Text)
	if !ok {
		f.Where.Error("Expected a numeric string argument for %()", f.Name)
	}
	return ast.Block{
		Type: "math_number_radix",
		Fields: []ast.Field{
			{Name: "OP", Value: fieldOp},
			{Name: "NUM", Value: textExpr.Content},
		},
	}
}
