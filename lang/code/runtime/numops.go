package runtime

// numops.go holds numeric operations that are less common or specialised:
// base conversions, colour manipulation, random numbers, and list statistics.
// Keeping them here lets deffuncs.go stay focused on the core built-in dispatch.

import (
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

// --- Base-format predicates (used by the ? questionnaire operator) ---

func isBase10(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func isHex(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}

func isBinary(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c != '0' && c != '1' {
			return false
		}
	}
	return true
}

// --- Base conversions ---

func evalDecToHex(args []Value) Value {
	n := int64(args[0].AsNum())
	return StrVal(strings.ToUpper(strconv.FormatInt(n, 16)))
}

func evalDecToBin(args []Value) Value {
	return StrVal(strconv.FormatInt(int64(args[0].AsNum()), 2))
}

func evalHexToDec(args []Value) Value {
	n, err := strconv.ParseInt(args[0].AsStr(), 16, 64)
	if err != nil {
		panic("invalid hex string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalBinToDec(args []Value) Value {
	n, err := strconv.ParseInt(args[0].AsStr(), 2, 64)
	if err != nil {
		panic("invalid binary string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalDec(args []Value) Value {
	n, err := strconv.ParseInt(args[0].AsStr(), 10, 64)
	if err != nil {
		panic("invalid decimal string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalBin(args []Value) Value {
	n, err := strconv.ParseInt(args[0].AsStr(), 2, 64)
	if err != nil {
		panic("invalid binary string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalOctal(args []Value) Value {
	n, err := strconv.ParseInt(args[0].AsStr(), 8, 64)
	if err != nil {
		panic("invalid octal string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalHexa(args []Value) Value {
	n, err := strconv.ParseInt(args[0].AsStr(), 16, 64)
	if err != nil {
		panic("invalid hex string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

// --- Color manipulation ---

// hexByte formats a single byte (0–255) as a two-character uppercase hex string.
func hexByte(n int) string {
	s := strconv.FormatInt(int64(n&0xFF), 16)
	if len(s) < 2 {
		return "0" + strings.ToUpper(s)
	}
	return strings.ToUpper(s)
}

func evalMakeColor(args []Value) Value {
	list := args[0].AsList()
	r := int((*list)[0].AsNum())
	g := int((*list)[1].AsNum())
	b := int((*list)[2].AsNum())
	a := 255
	if len(*list) >= 4 {
		a = int((*list)[3].AsNum())
	}
	return ColorVal("#" + hexByte(a) + hexByte(r) + hexByte(g) + hexByte(b))
}

func evalSplitColor(args []Value) Value {
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
}

// --- Random ---

func evalRandInt(args []Value) Value {
	lo, hi := int(args[0].AsNum()), int(args[1].AsNum())
	return NumVal(float64(lo + rand.Intn(hi-lo+1)))
}

func evalRandFloat(_ []Value) Value {
	return NumVal(rand.Float64())
}

func evalSetRandSeed(args []Value) Value {
	rand.Seed(int64(args[0].AsNum()))
	return NullVal()
}

// --- List statistics ---

func evalAvgOf(args []Value) Value {
	list := args[0].AsList()
	if len(*list) == 0 {
		return NumVal(0)
	}
	sum := 0.0
	for _, v := range *list {
		sum += v.AsNum()
	}
	return NumVal(sum / float64(len(*list)))
}

func evalMaxOf(args []Value) Value {
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
}

func evalMinOf(args []Value) Value {
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
}

func evalGeoMeanOf(args []Value) Value {
	list := args[0].AsList()
	if len(*list) == 0 {
		return NumVal(0)
	}
	product := 1.0
	for _, v := range *list {
		product *= v.AsNum()
	}
	return NumVal(math.Pow(product, 1.0/float64(len(*list))))
}

func evalStdDevOf(args []Value) Value {
	return NumVal(listStdDev(args[0].AsList(), false))
}

func evalStdErrOf(args []Value) Value {
	list := args[0].AsList()
	n := float64(len(*list))
	if n <= 1 {
		return NumVal(0)
	}
	return NumVal(listStdDev(list, false) / math.Sqrt(n))
}

func evalModeOf(args []Value) Value {
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
	return ListVal([]Value{StrVal(modeVal)})
}

// --- Secondary math ops ---

func evalMod(args []Value) Value {
	a, b := args[0].AsNum(), args[1].AsNum()
	return NumVal(math.Mod(math.Abs(a), math.Abs(b)) * sign(a))
}

func evalRem(args []Value) Value {
	a, b := args[0].AsNum(), args[1].AsNum()
	return NumVal(math.Mod(a, b))
}

func evalQuot(args []Value) Value {
	a, b := args[0].AsNum(), args[1].AsNum()
	return NumVal(math.Trunc(a / b))
}

func evalAtan2(args []Value) Value {
	return NumVal(math.Atan2(args[0].AsNum(), args[1].AsNum()))
}

func evalFormatDecimal(args []Value) Value {
	n, places := args[0].AsNum(), int(args[1].AsNum())
	return StrVal(strconv.FormatFloat(n, 'f', places, 64))
}

// --- Shared helpers ---

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
