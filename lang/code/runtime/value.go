package runtime

import (
	"math"
	"strconv"
	"strings"
)

type ValueType int

const (
	Null ValueType = iota
	Bool
	Number
	String
	List
	Dict
	Color
	NonConsumable // result of a statement; consuming it is a runtime error
)

type Value struct {
	vtype   ValueType
	boolVal bool
	numVal  float64
	strVal  string
	listVal *[]Value
	dictVal *OrderedDict
}

// --- Constructors ---

func NullVal() Value            { return Value{vtype: Null} }
func BoolVal(b bool) Value      { return Value{vtype: Bool, boolVal: b} }
func NumVal(n float64) Value    { return Value{vtype: Number, numVal: n} }
func StrVal(s string) Value     { return Value{vtype: String, strVal: s} }
func ColorVal(hex string) Value { return Value{vtype: Color, strVal: hex} }
func VoidVal() Value            { return Value{vtype: NonConsumable} }

func ListVal(elems []Value) Value {
	cp := make([]Value, len(elems))
	copy(cp, elems)
	return Value{vtype: List, listVal: &cp}
}

func EmptyList() Value {
	elems := make([]Value, 0)
	return Value{vtype: List, listVal: &elems}
}

func DictVal(d *OrderedDict) Value {
	return Value{vtype: Dict, dictVal: d}
}

func EmptyDict() Value {
	return Value{vtype: Dict, dictVal: NewOrderedDict()}
}

// --- Accessors ---

func (v Value) Type() ValueType { return v.vtype }

// TypeName returns a human-readable type name for use in error messages.
func (v Value) TypeName() string {
	switch v.vtype {
	case Null:
		return "null"
	case Bool:
		return "boolean"
	case Number:
		return "number"
	case String:
		return "text"
	case List:
		return "list"
	case Dict:
		return "dict"
	case Color:
		return "color"
	default:
		return "unknown"
	}
}

// errorStr returns a short description of the value for use in error messages.
func (v Value) errorStr() string {
	switch v.vtype {
	case Number:
		return "number " + v.String()
	case Bool:
		return "boolean " + v.String()
	case String:
		s := v.strVal
		if len(s) > 24 {
			s = s[:24] + "..."
		}
		return "text \"" + s + "\""
	case List:
		return "list (length " + strconv.Itoa(len(*v.listVal)) + ")"
	case Dict:
		return "dict (length " + strconv.Itoa(v.dictVal.Len()) + ")"
	case Null:
		return "null"
	case Color:
		return "color " + v.strVal
	default:
		return "unknown"
	}
}

func (v Value) AsBool() bool {
	if v.vtype == NonConsumable {
		panic("expected a boolean value but got a statement result (void)")
	}
	if v.vtype != Bool {
		panic("expected a boolean value but got " + v.errorStr())
	}
	return v.boolVal
}

func (v Value) AsNum() float64 {
	if v.vtype == NonConsumable {
		panic("expected a number value but got a statement result (void)")
	}
	switch v.vtype {
	case Number:
		return v.numVal
	case String:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v.strVal), 64); err == nil {
			return f
		}
		panic("cannot convert text to number: \"" + v.strVal + "\"")
	default:
		panic("expected a number value but got " + v.errorStr())
	}
}

func (v Value) AsStr() string {
	if v.vtype == NonConsumable {
		panic("cannot consume a statement result as a string")
	}
	switch v.vtype {
	case String:
		return v.strVal
	case Color:
		return v.strVal
	case Number:
		return formatNum(v.numVal)
	case Bool:
		if v.boolVal {
			return "true"
		}
		return "false"
	case Null:
		return ""
	case List:
		return v.String()
	case Dict:
		return v.String()
	default:
		return v.String()
	}
}

func (v Value) AsList() *[]Value {
	if v.vtype == NonConsumable {
		panic("expected a list value but got a statement result (void)")
	}
	if v.vtype != List {
		panic("expected a list value but got " + v.errorStr())
	}
	return v.listVal
}

func (v Value) AsDict() *OrderedDict {
	if v.vtype == NonConsumable {
		panic("expected a dict value but got a statement result (void)")
	}
	if v.vtype != Dict {
		panic("expected a dict value but got " + v.errorStr())
	}
	return v.dictVal
}

func CoerceNum(v Value) (float64, bool) {
	if v.vtype == Number {
		return v.numVal, true
	}
	if v.vtype == String {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v.strVal), 64); err == nil {
			if !math.IsNaN(f) && !math.IsInf(f, 0) {
				return f, true
			}
		}
	}
	return 0, false
}

func (v Value) String() string {
	return v.stringDepth(0)
}

func (v Value) stringDepth(depth int) string {
	if depth > 50 {
		return "..."
	}
	switch v.vtype {
	case Null:
		return "null"
	case Bool:
		if v.boolVal {
			return "true"
		}
		return "false"
	case Number:
		return formatNum(v.numVal)
	case String:
		return v.strVal
	case Color:
		return v.strVal
	case List:
		parts := make([]string, len(*v.listVal))
		for i, e := range *v.listVal {
			parts[i] = e.stringDepth(depth + 1)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case Dict:
		return v.dictVal.stringDepth(depth)
	case NonConsumable:
		return "<void>"
	default:
		return "<unknown>"
	}
}

func formatNum(f float64) string {
	if math.IsInf(f, 1) {
		return "Infinity"
	}
	if math.IsInf(f, -1) {
		return "-Infinity"
	}
	if math.IsNaN(f) {
		return "NaN"
	}
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// deepCopyValue returns a fully independent deep copy of v.
// Primitive values (null, bool, number, string, color) are immutable and returned as-is.
// Lists and dicts are recursively cloned.
func deepCopyValue(v Value) Value {
	switch v.vtype {
	case List:
		src := *v.listVal
		cp := make([]Value, len(src))
		for i, elem := range src {
			cp[i] = deepCopyValue(elem)
		}
		return Value{vtype: List, listVal: &cp}
	case Dict:
		nd := NewOrderedDict()
		for _, entry := range v.dictVal.entries {
			nd.Set(entry.Key, deepCopyValue(entry.Val))
		}
		return DictVal(nd)
	default:
		return v
	}
}

// DeepEqual checks structural equality of two values.
func DeepEqual(a, b Value) bool {
	return deepEqualDepth(a, b, 0)
}

func deepEqualDepth(a, b Value, depth int) bool {
	if depth > 50 {
		panic("circular reference detected in equality comparison")
	}
	if a.vtype != b.vtype {
		return false
	}
	switch a.vtype {
	case Null:
		return true
	case Bool:
		return a.boolVal == b.boolVal
	case Number:
		return a.numVal == b.numVal
	case String, Color:
		return a.strVal == b.strVal
	case List:
		if a.listVal == b.listVal {
			return true // same pointer — same list (handles self-referential equality)
		}
		la, lb := *a.listVal, *b.listVal
		if len(la) != len(lb) {
			return false
		}
		for i := range la {
			if !deepEqualDepth(la[i], lb[i], depth+1) {
				return false
			}
		}
		return true
	case Dict:
		da, db := a.dictVal, b.dictVal
		if da == db {
			return true // same pointer — same dict
		}
		if da.Len() != db.Len() {
			return false
		}
		for _, e := range da.entries {
			bv, ok := db.Get(e.Key)
			if !ok || !deepEqualDepth(e.Val, bv, depth+1) {
				return false
			}
		}
		return true
	case NonConsumable:
		return true
	}
	return false
}

// --- OrderedDict ---

type DictEntry struct {
	Key string
	Val Value
}

type OrderedDict struct {
	entries []DictEntry
}

func NewOrderedDict() *OrderedDict {
	return &OrderedDict{}
}

func (d *OrderedDict) Len() int { return len(d.entries) }

func (d *OrderedDict) Get(key string) (Value, bool) {
	for _, e := range d.entries {
		if e.Key == key {
			return e.Val, true
		}
	}
	return NullVal(), false
}

func (d *OrderedDict) Set(key string, val Value) {
	for i, e := range d.entries {
		if e.Key == key {
			d.entries[i].Val = val
			return
		}
	}
	d.entries = append(d.entries, DictEntry{Key: key, Val: val})
}

func (d *OrderedDict) Delete(key string) {
	for i, e := range d.entries {
		if e.Key == key {
			d.entries = append(d.entries[:i], d.entries[i+1:]...)
			return
		}
	}
}

func (d *OrderedDict) ContainsKey(key string) bool {
	_, ok := d.Get(key)
	return ok
}

func (d *OrderedDict) Keys() []Value {
	keys := make([]Value, len(d.entries))
	for i, e := range d.entries {
		keys[i] = StrVal(e.Key)
	}
	return keys
}

func (d *OrderedDict) Values() []Value {
	vals := make([]Value, len(d.entries))
	for i, e := range d.entries {
		vals[i] = e.Val
	}
	return vals
}

func (d *OrderedDict) Clone() *OrderedDict {
	nd := NewOrderedDict()
	nd.entries = make([]DictEntry, len(d.entries))
	copy(nd.entries, d.entries)
	return nd
}

func (d *OrderedDict) String() string {
	return d.stringDepth(0)
}

func (d *OrderedDict) stringDepth(depth int) string {
	if depth > 50 {
		return "{...}"
	}
	parts := make([]string, len(d.entries))
	for i, e := range d.entries {
		parts[i] = e.Key + ": " + e.Val.stringDepth(depth+1)
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
