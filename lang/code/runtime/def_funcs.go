package runtime

// deffuncs.go dispatches built-in (default) function calls.
// Numeric specials (base conversions, colour, random, statistics) live in numops.go.

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/variables"
	"math"
	"strings"
)

func (i *Interpreter) evalFuncCall(e *common.FuncCall) Value {
	savedToken := i.lastToken
	savedHighlight := i.lastHighlight
	if i.componentHost != nil && isGenericComponentFunc(e.Name) {
		i.lastToken = savedToken
		i.lastHighlight = savedHighlight
		return i.evalGenericComponentFunc(e)
	}
	args := make([]Value, len(e.Args))
	for k, a := range e.Args {
		args[k] = i.Eval(a)
	}
	i.lastToken = savedToken
	i.lastHighlight = savedHighlight

	switch e.Name {
	// --- Output ---
	case "println":
		i.printLine(args[0].AsStr())
		return VoidVal()

	// --- Math single-arg ---
	case "sqrt":
		return NumVal(math.Sqrt(args[0].AsNum()))
	case "abs":
		return NumVal(math.Abs(args[0].AsNum()))
	case "neg":
		return NumVal(-args[0].AsNum())
	case "log":
		return NumVal(math.Log(args[0].AsNum()))
	case "exp":
		return NumVal(math.Exp(args[0].AsNum()))
	case "round":
		return NumVal(math.Round(args[0].AsNum()))
	case "ceil":
		return NumVal(math.Ceil(args[0].AsNum()))
	case "floor":
		return NumVal(math.Floor(args[0].AsNum()))
	case "sin":
		return NumVal(math.Sin(args[0].AsNum()))
	case "cos":
		return NumVal(math.Cos(args[0].AsNum()))
	case "tan":
		return NumVal(math.Tan(args[0].AsNum()))
	case "asin":
		return NumVal(math.Asin(args[0].AsNum()))
	case "acos":
		return NumVal(math.Acos(args[0].AsNum()))
	case "atan":
		return NumVal(math.Atan(args[0].AsNum()))
	case "degrees":
		return NumVal(args[0].AsNum() * 180 / math.Pi)
	case "radians":
		return NumVal(args[0].AsNum() * math.Pi / 180)

	// --- Base conversions (numops.go) ---
	case "decToHex":
		return evalDecToHex(args)
	case "decToBin":
		return evalDecToBin(args)
	case "hexToDec":
		return evalHexToDec(args)
	case "binToDec":
		return evalBinToDec(args)
	case "dec":
		return evalDec(args)
	case "bin":
		return evalBin(args)
	case "octal":
		return evalOctal(args)
	case "hexa":
		return evalHexa(args)

	// --- Random (numops.go) ---
	case "randInt":
		return evalRandInt(args)
	case "randFloat":
		return evalRandFloat(args)
	case "setRandSeed":
		return evalSetRandSeed(args)

	// --- Variadic min/max ---
	case "min":
		result := args[0].AsNum()
		for _, v := range args[1:] {
			if v.AsNum() < result {
				result = v.AsNum()
			}
		}
		return NumVal(result)
	case "max":
		result := args[0].AsNum()
		for _, v := range args[1:] {
			if v.AsNum() > result {
				result = v.AsNum()
			}
		}
		return NumVal(result)

	// --- List statistics (numops.go) ---
	case "avgOf":
		return evalAvgOf(args)
	case "maxOf":
		return evalMaxOf(args)
	case "minOf":
		return evalMinOf(args)
	case "geoMeanOf":
		return evalGeoMeanOf(args)
	case "stdDevOf":
		return evalStdDevOf(args)
	case "stdErrOf":
		return evalStdErrOf(args)
	case "modeOf":
		return evalModeOf(args)

	// --- Secondary math ops (numops.go) ---
	case "mod":
		return evalMod(args)
	case "rem":
		return evalRem(args)
	case "quot":
		return evalQuot(args)
	case "atan2":
		return evalAtan2(args)
	case "formatDecimal":
		return evalFormatDecimal(args)

	// --- List helpers ---
	case "copyList":
		if args[0].Type() != List {
			panic("copyList requires a list, got " + args[0].TypeName())
		}
		return deepCopyValue(args[0])
	case "copyDict":
		if args[0].Type() != Dict {
			panic("copyDict requires a dict, got " + args[0].TypeName())
		}
		return deepCopyValue(args[0])
	case "makeNdArray":
		return i.evalMatrixCreateND(args[0], args[1])

	// --- Color (numops.go) ---
	case "makeColor":
		return evalMakeColor(args)
	case "splitColor":
		return evalSplitColor(args)

	// --- App Inventor screen stubs ---
	case "openScreen":
		i.stub("openScreen(" + args[0].AsStr() + ")")
		return VoidVal()
	case "openScreenWithValue":
		i.stub("openScreenWithValue(" + args[0].AsStr() + ", ...)")
		return VoidVal()
	case "closeScreen":
		i.stub("closeScreen()")
		return VoidVal()
	case "closeScreenWithValue":
		i.stub("closeScreenWithValue(...)")
		return VoidVal()
	case "closeApp":
		i.stub("closeApp()")
		return VoidVal()
	case "getPlainStartText":
		i.stub("getPlainStartText()")
		return StrVal("")
	case "closeScreenWithPlainText":
		i.stub("closeScreenWithPlainText(...)")
		return VoidVal()
	case "getStartValue":
		i.stub("getStartValue()")
		return StrVal("")

	// --- Generic component function stubs ---
	case "set", "call", "vcall":
		i.stub("component function '" + e.Name + "'")
		return VoidVal()
	case "get", "every":
		i.stub("component function '" + e.Name + "'")
		return NullVal()

	default:
		panic("unknown built-in function: " + e.Name)
	}
}

func isGenericComponentFunc(name string) bool {
	switch name {
	case "set", "get", "call", "vcall", "every":
		return true
	default:
		return false
	}
}

func (i *Interpreter) evalGenericComponentFunc(e *common.FuncCall) Value {
	switch e.Name {
	case "every":
		componentType := i.componentTypeArgument(e.Args[0])
		return componentList(i.componentHost.ComponentNames(componentType))
	case "get":
		componentType := i.componentTypeArgument(e.Args[0])
		componentName := i.resolveComponentName(i.Eval(e.Args[1]), componentType)
		property := i.Eval(e.Args[2]).AsStr()
		return i.componentHost.GetProperty(componentName, componentType, property)
	case "set":
		componentType := i.componentTypeArgument(e.Args[0])
		componentName := i.resolveComponentName(i.Eval(e.Args[1]), componentType)
		property := i.Eval(e.Args[2]).AsStr()
		i.componentHost.SetProperty(componentName, componentType, property, i.Eval(e.Args[3]))
		return VoidVal()
	case "call", "vcall":
		componentType := i.componentTypeArgument(e.Args[0])
		componentName := i.resolveComponentName(i.Eval(e.Args[1]), componentType)
		method := i.Eval(e.Args[2]).AsStr()
		args := make([]Value, 0, len(e.Args)-3)
		for _, arg := range e.Args[3:] {
			args = append(args, i.Eval(arg))
		}
		return i.componentHost.CallMethod(componentName, componentType, method, args)
	default:
		panic("unknown generic component function: " + e.Name)
	}
}

func (i *Interpreter) componentTypeArgument(expr ast.Expr) string {
	switch n := ast.UnwrapAnnotated(expr).(type) {
	case *fundamentals.Text:
		return strings.TrimSpace(n.Content)
	case *variables.Get:
		if !n.Global && len(i.componentHost.ComponentNames(n.Name)) > 0 {
			return n.Name
		}
	}
	return strings.TrimSpace(i.Eval(expr).AsStr())
}
