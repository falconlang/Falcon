package runtime

// deffuncs.go dispatches built-in (default) function calls.
// Numeric specials (base conversions, colour, random, statistics) live in numops.go.

import (
	"Falcon/code/ast/common"
	"math"
)

func (i *Interpreter) evalFuncCall(e *common.FuncCall) Value {
	args := make([]Value, len(e.Args))
	for k, a := range e.Args {
		args[k] = i.eval(a)
	}

	switch e.Name {
	// --- Output ---
	case "println":
		printLine(args[0].AsStr())
		return NullVal()

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
		src := args[0].AsList()
		cp := make([]Value, len(*src))
		copy(cp, *src)
		return Value{vtype: List, listVal: &cp}
	case "copyDict":
		return DictVal(args[0].AsDict().Clone())

	// --- Color (numops.go) ---
	case "makeColor":
		return evalMakeColor(args)
	case "splitColor":
		return evalSplitColor(args)

	// --- App Inventor screen stubs ---
	case "openScreen":
		stub("openScreen(" + args[0].AsStr() + ")")
		return NullVal()
	case "openScreenWithValue":
		stub("openScreenWithValue(" + args[0].AsStr() + ", ...)")
		return NullVal()
	case "closeScreen":
		stub("closeScreen()")
		return NullVal()
	case "closeScreenWithValue":
		stub("closeScreenWithValue(...)")
		return NullVal()
	case "closeApp":
		stub("closeApp()")
		return NullVal()
	case "getPlainStartText":
		stub("getPlainStartText()")
		return StrVal("")
	case "closeScreenWithPlainText":
		stub("closeScreenWithPlainText(...)")
		return NullVal()
	case "getStartValue":
		stub("getStartValue()")
		return StrVal("")

	// --- Generic component function stubs ---
	case "set", "get", "call", "vcall", "every":
		stub("component function '" + e.Name + "'")
		return NullVal()

	default:
		panic("unknown built-in function: " + e.Name)
	}
}
