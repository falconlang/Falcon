package runtime

// numops.go holds numeric operations that are less common or specialised:
// base conversions, colour manipulation, random numbers, and list statistics.
// Keeping them here lets deffuncs.go stay focused on the core built-in dispatch.

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// --- Base-format predicates (used by the ? questionnaire operator) ---

func isBase10(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
		if len(s) == 0 {
			return false
		}
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
	n := args[0].AsNum()
	if n != math.Trunc(n) {
		panic("decToHex requires an integer, got " + formatNum(n))
	}
	return StrVal(strings.ToUpper(strconv.FormatInt(int64(n), 16)))
}

func evalDecToBin(args []Value) Value {
	n := args[0].AsNum()
	if n != math.Trunc(n) {
		panic("decToBin requires an integer, got " + formatNum(n))
	}
	return StrVal(strconv.FormatInt(int64(n), 2))
}

func evalHexToDec(args []Value) Value {
	s := strings.TrimPrefix(args[0].AsStr(), "0x")
	s = strings.TrimPrefix(s, "0X")
	n, err := strconv.ParseInt(s, 16, 64)
	if err != nil {
		panic("invalid hex string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalBinToDec(args []Value) Value {
	s := strings.TrimPrefix(args[0].AsStr(), "0b")
	s = strings.TrimPrefix(s, "0B")
	n, err := strconv.ParseInt(s, 2, 64)
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
	s := strings.TrimPrefix(args[0].AsStr(), "0b")
	s = strings.TrimPrefix(s, "0B")
	n, err := strconv.ParseInt(s, 2, 64)
	if err != nil {
		panic("invalid binary string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalOctal(args []Value) Value {
	s := strings.TrimPrefix(args[0].AsStr(), "0o")
	s = strings.TrimPrefix(s, "0O")
	n, err := strconv.ParseInt(s, 8, 64)
	if err != nil {
		panic("invalid octal string: " + args[0].AsStr())
	}
	return NumVal(float64(n))
}

func evalHexa(args []Value) Value {
	s := strings.TrimPrefix(args[0].AsStr(), "0x")
	s = strings.TrimPrefix(s, "0X")
	n, err := strconv.ParseInt(s, 16, 64)
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

func clampByte(n int) int {
	if n < 0 {
		return 0
	}
	if n > 255 {
		return 255
	}
	return n
}

func evalMakeColor(args []Value) Value {
	list := args[0].AsList()
	if len(*list) < 3 {
		panic("makeColor requires a list of at least 3 elements [r, g, b]")
	}
	r := clampByte(int((*list)[0].AsNum()))
	g := clampByte(int((*list)[1].AsNum()))
	b := clampByte(int((*list)[2].AsNum()))
	a := 255
	if len(*list) >= 4 {
		a = clampByte(int((*list)[3].AsNum()))
	}
	return ColorVal("#" + hexByte(a) + hexByte(r) + hexByte(g) + hexByte(b))
}

func evalSplitColor(args []Value) Value {
	var a, r, g, b int64
	// A numeric argument is treated as a 32-bit ARGB integer.
	if args[0].Type() == Number {
		nf := args[0].AsNum()
		if nf < 0 {
			panic("splitColor: color number must be non-negative, got " + formatNum(nf))
		}
		n := uint64(int64(nf))
		a = int64((n >> 24) & 0xFF)
		r = int64((n >> 16) & 0xFF)
		g = int64((n >> 8) & 0xFF)
		b = int64(n & 0xFF)
		return ListVal([]Value{NumVal(float64(a)), NumVal(float64(r)), NumVal(float64(g)), NumVal(float64(b))})
	}
	raw := args[0].AsStr()
	hex := strings.TrimPrefix(raw, "#")
	for _, c := range hex {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			panic("splitColor: invalid color value: " + raw)
		}
	}
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
	} else {
		panic("splitColor: color must be 6 or 8 hex digits, got: " + raw)
	}
	// Return [a, r, g, b] to match App Inventor channel order
	return ListVal([]Value{NumVal(float64(a)), NumVal(float64(r)), NumVal(float64(g)), NumVal(float64(b))})
}

// --- Random ---

// --- RNG (xorshift64) ---

var rngState uint64 = 0x9e3779b97f4a7c15 // non-zero default seed; overwritten by init

func init() {
	s := uint64(time.Now().UnixNano())
	if s == 0 {
		s = 0x9e3779b97f4a7c15
	}
	rngState = s
}

func rngNext() uint64 {
	rngState ^= rngState << 13
	rngState ^= rngState >> 7
	rngState ^= rngState << 17
	return rngState
}

func rngIntn(n int) int {
	return int(rngNext()>>1) % n
}

func evalRandInt(args []Value) Value {
	lo, hi := int(args[0].AsNum()), int(args[1].AsNum())
	if lo > hi {
		lo, hi = hi, lo
	}
	return NumVal(float64(lo + rngIntn(hi-lo+1)))
}

func evalRandFloat(_ []Value) Value {
	// Use the high 53 bits for a uniform [0,1) float64.
	return NumVal(float64(rngNext()>>11) / (1 << 53))
}

func evalSetRandSeed(args []Value) Value {
	s := args[0].AsNum()
	if s < 0 {
		panic("setRandSeed requires a non-negative seed, got " + formatNum(s))
	}
	seed := uint64(s)
	if seed == 0 {
		seed = 1 // xorshift64 must not be zero
	}
	rngState = seed
	return NullVal()
}

// --- List statistics ---

func evalAvgOf(args []Value) Value {
	list := args[0].AsList()
	if len(*list) == 0 {
		panic("avgOf() requires a non-empty list")
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
		panic("maxOf() requires a non-empty list")
	}
	mx := (*list)[0].AsNum()
	for _, v := range (*list)[1:] {
		n := v.AsNum()
		if math.IsNaN(n) {
			mx = n
			break
		}
		if n > mx {
			mx = n
		}
	}
	return NumVal(mx)
}

func evalMinOf(args []Value) Value {
	list := args[0].AsList()
	if len(*list) == 0 {
		panic("minOf() requires a non-empty list")
	}
	mn := (*list)[0].AsNum()
	for _, v := range (*list)[1:] {
		n := v.AsNum()
		if math.IsNaN(n) {
			mn = n
			break
		}
		if n < mn {
			mn = n
		}
	}
	return NumVal(mn)
}

func evalGeoMeanOf(args []Value) Value {
	list := args[0].AsList()
	if len(*list) == 0 {
		panic("geoMeanOf() requires a non-empty list")
	}
	product := 1.0
	for _, v := range *list {
		n := v.AsNum()
		if n < 0 {
			panic("geoMeanOf() requires non-negative values")
		}
		product *= n
	}
	return NumVal(math.Pow(product, 1.0/float64(len(*list))))
}

func evalStdDevOf(args []Value) Value {
	return NumVal(listStdDev(args[0].AsList(), false))
}

func evalStdErrOf(args []Value) Value {
	list := args[0].AsList()
	n := float64(len(*list))
	if n == 0 {
		panic("stdErrOf() requires a non-empty list")
	}
	if n <= 1 {
		return NumVal(0)
	}
	return NumVal(listStdDev(list, false) / math.Sqrt(n))
}

func evalModeOf(args []Value) Value {
	list := args[0].AsList()
	if len(*list) == 0 {
		panic("modeOf() requires a non-empty list")
	}
	type entry struct {
		val   Value
		count int
	}
	var entries []*entry
outer:
	for _, v := range *list {
		for _, e := range entries {
			if DeepEqual(v, e.val) {
				e.count++
				continue outer
			}
		}
		entries = append(entries, &entry{val: v, count: 1})
	}
	maxCount := 0
	for _, e := range entries {
		if e.count > maxCount {
			maxCount = e.count
		}
	}
	var modes []Value
	for _, e := range entries {
		if e.count == maxCount {
			modes = append(modes, e.val)
		}
	}
	return ListVal(modes)
}

// --- Secondary math ops ---

func evalFormatDecimal(args []Value) Value {
	n, rawPlaces := args[0].AsNum(), args[1].AsNum()
	places := int(rawPlaces)
	if rawPlaces != float64(places) || places < 0 {
		panic("formatDecimal: places must be a non-negative integer, got " + formatNum(rawPlaces))
	}
	if places > 100 {
		panic("formatDecimal: places cannot exceed 100")
	}
	return StrVal(strconv.FormatFloat(n, 'f', places, 64))
}

func evalMod(args []Value) Value {
	a, b := args[0].AsNum(), args[1].AsNum()
	if b == 0 {
		panic("modulus by zero")
	}
	// floor modulo: result has the same sign as the divisor
	r := math.Mod(a, b)
	if r != 0 && (r < 0) != (b < 0) {
		r += b
	}
	return NumVal(r)
}

func evalRem(args []Value) Value {
	a, b := args[0].AsNum(), args[1].AsNum()
	if b == 0 {
		panic("remainder by zero")
	}
	return NumVal(math.Mod(a, b))
}

func evalQuot(args []Value) Value {
	a, b := args[0].AsNum(), args[1].AsNum()
	if b == 0 {
		panic("quotient by zero")
	}
	return NumVal(math.Trunc(a / b))
}

func evalAtan2(args []Value) Value {
	return NumVal(math.Atan2(args[0].AsNum(), args[1].AsNum()))
}

// --- Shared helpers ---

func listStdDev(list *[]Value, population bool) float64 {
	n := float64(len(*list))
	if n == 0 {
		panic("stdDevOf() / stdErrOf() requires a non-empty list")
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

// insertionSort performs a stable in-place sort using the provided less function.
func insertionSort(n int, less func(i, j int) bool, swap func(i, j int)) {
	for i := 1; i < n; i++ {
		for j := i; j > 0 && less(j, j-1); j-- {
			swap(j, j-1)
		}
	}
}

// sortValues sorts a slice of Values numerically if possible, else lexicographically.
func sortValues(list []Value) {
	insertionSort(len(list), func(a, b int) bool {
		va, vb := list[a], list[b]
		na, aOk := CoerceNum(va)
		nb, bOk := CoerceNum(vb)
		if aOk && bOk {
			return na < nb
		}
		return va.AsStr() < vb.AsStr()
	}, func(a, b int) { list[a], list[b] = list[b], list[a] })
}
