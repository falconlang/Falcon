package runtime

import (
	"Falcon/code/ast/common"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

// evalFuncCall dispatches built-in (default) function calls.
func (i *Interpreter) evalFuncCall(e *common.FuncCall) Value {
	// evaluate arguments eagerly (except short-circuit is handled in BinaryExpr)
	args := make([]Value, len(e.Args))
	for k, a := range e.Args {
		args[k] = i.eval(a)
	}

	switch e.Name {
	// --- Output ---
	case "println":
		fmt.Println(args[0].AsStr())
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

	// --- Base conversions ---
	case "decToHex":
		return StrVal(fmt.Sprintf("%X", int64(args[0].AsNum())))
	case "decToBin":
		return StrVal(strconv.FormatInt(int64(args[0].AsNum()), 2))
	case "hexToDec":
		n, err := strconv.ParseInt(args[0].AsStr(), 16, 64)
		if err != nil {
			panic("invalid hex string: " + args[0].AsStr())
		}
		return NumVal(float64(n))
	case "binToDec":
		n, err := strconv.ParseInt(args[0].AsStr(), 2, 64)
		if err != nil {
			panic("invalid binary string: " + args[0].AsStr())
		}
		return NumVal(float64(n))
	case "dec":
		n, err := strconv.ParseInt(args[0].AsStr(), 10, 64)
		if err != nil {
			panic("invalid decimal string: " + args[0].AsStr())
		}
		return NumVal(float64(n))
	case "bin":
		n, err := strconv.ParseInt(args[0].AsStr(), 2, 64)
		if err != nil {
			panic("invalid binary string: " + args[0].AsStr())
		}
		return NumVal(float64(n))
	case "octal":
		n, err := strconv.ParseInt(args[0].AsStr(), 8, 64)
		if err != nil {
			panic("invalid octal string: " + args[0].AsStr())
		}
		return NumVal(float64(n))
	case "hexa":
		n, err := strconv.ParseInt(args[0].AsStr(), 16, 64)
		if err != nil {
			panic("invalid hex string: " + args[0].AsStr())
		}
		return NumVal(float64(n))

	// --- Random ---
	case "randInt":
		lo, hi := int(args[0].AsNum()), int(args[1].AsNum())
		return NumVal(float64(lo + rand.Intn(hi-lo+1)))
	case "randFloat":
		return NumVal(rand.Float64())
	case "setRandSeed":
		rand.Seed(int64(args[0].AsNum()))
		return NullVal()

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

	// --- List stats ---
	case "avgOf":
		list := args[0].AsList()
		if len(*list) == 0 {
			return NumVal(0)
		}
		sum := 0.0
		for _, v := range *list {
			sum += v.AsNum()
		}
		return NumVal(sum / float64(len(*list)))
	case "maxOf":
		list := args[0].AsList()
		if len(*list) == 0 {
			panic("maxOf: empty list")
		}
		mx := (*list)[0].AsNum()
		for _, v := range (*list)[1:] {
			if v.AsNum() > mx {
				mx = v.AsNum()
			}
		}
		return NumVal(mx)
	case "minOf":
		list := args[0].AsList()
		if len(*list) == 0 {
			panic("minOf: empty list")
		}
		mn := (*list)[0].AsNum()
		for _, v := range (*list)[1:] {
			if v.AsNum() < mn {
				mn = v.AsNum()
			}
		}
		return NumVal(mn)
	case "geoMeanOf":
		list := args[0].AsList()
		if len(*list) == 0 {
			return NumVal(0)
		}
		product := 1.0
		for _, v := range *list {
			product *= v.AsNum()
		}
		return NumVal(math.Pow(product, 1.0/float64(len(*list))))
	case "stdDevOf":
		return NumVal(listStdDev(args[0].AsList(), false))
	case "stdErrOf":
		list := args[0].AsList()
		n := float64(len(*list))
		if n <= 1 {
			return NumVal(0)
		}
		return NumVal(listStdDev(list, false) / math.Sqrt(n))
	case "modeOf":
		list := args[0].AsList()
		counts := make(map[string]int)
		var order []string
		for _, v := range *list {
			s := v.String()
			if counts[s] == 0 {
				order = append(order, s)
			}
			counts[s]++
		}
		maxCount := 0
		var modeVal string
		for _, s := range order {
			if counts[s] > maxCount {
				maxCount = counts[s]
				modeVal = s
			}
		}
		// Return as list of modes (App Inventor returns a list)
		return ListVal([]Value{StrVal(modeVal)})

	// --- Math ops ---
	case "mod":
		a, b := args[0].AsNum(), args[1].AsNum()
		return NumVal(math.Mod(math.Abs(a), math.Abs(b)) * sign(a))
	case "rem":
		a, b := args[0].AsNum(), args[1].AsNum()
		return NumVal(math.Mod(a, b))
	case "quot":
		a, b := args[0].AsNum(), args[1].AsNum()
		return NumVal(math.Trunc(a / b))
	case "atan2":
		return NumVal(math.Atan2(args[0].AsNum(), args[1].AsNum()))
	case "formatDecimal":
		n, places := args[0].AsNum(), int(args[1].AsNum())
		return StrVal(strconv.FormatFloat(n, 'f', places, 64))

	// --- List ---
	case "copyList":
		src := args[0].AsList()
		cp := make([]Value, len(*src))
		copy(cp, *src)
		return Value{vtype: List, listVal: &cp}
	case "copyDict":
		return DictVal(args[0].AsDict().Clone())

	// --- Color ---
	case "makeColor":
		// arg is a list [r, g, b] or [r, g, b, a]
		list := args[0].AsList()
		r := int((*list)[0].AsNum()) & 0xFF
		g := int((*list)[1].AsNum()) & 0xFF
		b := int((*list)[2].AsNum()) & 0xFF
		a := 255
		if len(*list) >= 4 {
			a = int((*list)[3].AsNum()) & 0xFF
		}
		hex := fmt.Sprintf("#%02X%02X%02X%02X", a, r, g, b)
		return ColorVal(hex)
	case "splitColor":
		hex := strings.TrimPrefix(args[0].AsStr(), "#")
		var a, r, g, b int64
		if len(hex) == 8 {
			a, _ = strconv.ParseInt(hex[0:2], 16, 64)
			r, _ = strconv.ParseInt(hex[2:4], 16, 64)
			g, _ = strconv.ParseInt(hex[4:6], 16, 64)
			b, _ = strconv.ParseInt(hex[6:8], 16, 64)
		} else if len(hex) == 6 {
			a = 255
			r, _ = strconv.ParseInt(hex[0:2], 16, 64)
			g, _ = strconv.ParseInt(hex[2:4], 16, 64)
			b, _ = strconv.ParseInt(hex[4:6], 16, 64)
		}
		return ListVal([]Value{NumVal(float64(r)), NumVal(float64(g)), NumVal(float64(b)), NumVal(float64(a))})

	// --- App Inventor screen stubs ---
	case "openScreen":
		fmt.Printf("[stub] openScreen(%s) is not supported outside App Inventor\n", args[0].AsStr())
		return NullVal()
	case "openScreenWithValue":
		fmt.Printf("[stub] openScreenWithValue(%s, ...) is not supported outside App Inventor\n", args[0].AsStr())
		return NullVal()
	case "closeScreen":
		fmt.Println("[stub] closeScreen() is not supported outside App Inventor")
		return NullVal()
	case "closeScreenWithValue":
		fmt.Println("[stub] closeScreenWithValue(...) is not supported outside App Inventor")
		return NullVal()
	case "closeApp":
		fmt.Println("[stub] closeApp() is not supported outside App Inventor")
		return NullVal()
	case "getPlainStartText":
		fmt.Println("[stub] getPlainStartText() is not supported outside App Inventor")
		return StrVal("")
	case "closeScreenWithPlainText":
		fmt.Println("[stub] closeScreenWithPlainText(...) is not supported outside App Inventor")
		return NullVal()
	case "getStartValue":
		fmt.Println("[stub] getStartValue() is not supported outside App Inventor")
		return StrVal("")

	// --- Generic component function stubs ---
	case "set", "get", "call", "vcall", "every":
		fmt.Printf("[stub] component function '%s' is not supported outside App Inventor\n", e.Name)
		return NullVal()

	default:
		panic("unknown built-in function: " + e.Name)
	}
}

func listStdDev(list *[]Value, population bool) float64 {
	n := float64(len(*list))
	if n == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range *list {
		sum += v.AsNum()
	}
	mean := sum / n
	variance := 0.0
	for _, v := range *list {
		d := v.AsNum() - mean
		variance += d * d
	}
	denom := n
	if !population && n > 1 {
		denom = n - 1
	}
	return math.Sqrt(variance / denom)
}

func sign(f float64) float64 {
	if f < 0 {
		return -1
	}
	return 1
}

// sortValues sorts a slice of Values numerically if possible, else lexicographically.
func sortValues(list []Value) {
	sort.SliceStable(list, func(a, b int) bool {
		va, vb := list[a], list[b]
		na, aOk := TryNum(va)
		nb, bOk := TryNum(vb)
		if aOk && bOk {
			return na < nb
		}
		return va.AsStr() < vb.AsStr()
	})
}
