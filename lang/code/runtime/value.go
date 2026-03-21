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

func (v Value) AsBool() bool {
	if v.vtype != Bool {
		panic("expected a boolean value")
	}
	return v.boolVal
}

func (v Value) AsNum() float64 {
	switch v.vtype {
	case Number:
		return v.numVal
	case String:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v.strVal), 64); err == nil {
			return f
		}
		panic("cannot convert string to number: " + v.strVal)
	default:
		panic("expected a number value")
	}
}

func (v Value) AsStr() string {
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
	if v.vtype != List {
		panic("expected a list value")
	}
	return v.listVal
}

func (v Value) AsDict() *OrderedDict {
	if v.vtype != Dict {
		panic("expected a dict value")
	}
	return v.dictVal
}

func CoerceNum(v Value) (float64, bool) {
	if v.vtype == Number {
		return v.numVal, true
	}
	if v.vtype == String {
		if f, err := strconv.ParseFloat(strings.TrimSpace(v.strVal), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

func (v Value) String() string {
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
			parts[i] = e.String()
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case Dict:
		return v.dictVal.String()
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

// DeepEqual checks structural equality of two values.
func DeepEqual(a, b Value) bool {
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
		la, lb := *a.listVal, *b.listVal
		if len(la) != len(lb) {
			return false
		}
		for i := range la {
			if !DeepEqual(la[i], lb[i]) {
				return false
			}
		}
		return true
	case Dict:
		da, db := a.dictVal, b.dictVal
		if da.Len() != db.Len() {
			return false
		}
		for _, e := range da.entries {
			bv, ok := db.Get(e.Key)
			if !ok || !DeepEqual(e.Val, bv) {
				return false
			}
		}
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
	parts := make([]string, len(d.entries))
	for i, e := range d.entries {
		parts[i] = e.Key + ": " + e.Val.String()
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
