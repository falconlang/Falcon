package runtime

import (
	astlist "Falcon/code/ast/list"
	astmethod "Falcon/code/ast/method"
	"strings"
)

// methodCall dispatches method calls on text, list, and dict values.
func (i *Interpreter) methodCall(e *astmethod.Call) Value {
	on := i.Eval(e.On)
	args := i.evalExprs(e.Args)

	switch e.Name {
	// ============ Text methods ============
	case "textLen":
		return NumVal(float64(len([]rune(on.AsStr()))))
	case "trim":
		return StrVal(strings.TrimSpace(on.AsStr()))
	case "uppercase":
		return StrVal(strings.ToUpper(on.AsStr()))
	case "lowercase":
		return StrVal(strings.ToLower(on.AsStr()))
	case "startsWith":
		return BoolVal(strings.HasPrefix(on.AsStr(), args[0].AsStr()))
	case "contains":
		return BoolVal(strings.Contains(on.AsStr(), args[0].AsStr()))
	case "containsAny":
		haystack := on.AsStr()
		for _, v := range *args[0].AsList() {
			if strings.Contains(haystack, v.AsStr()) {
				return BoolVal(true)
			}
		}
		return BoolVal(false)
	case "containsAll":
		haystack := on.AsStr()
		for _, v := range *args[0].AsList() {
			if !strings.Contains(haystack, v.AsStr()) {
				return BoolVal(false)
			}
		}
		return BoolVal(true)
	case "split":
		sep := args[0].AsStr()
		parts := strings.Split(on.AsStr(), sep)
		elems := make([]Value, len(parts))
		for k, p := range parts {
			elems[k] = StrVal(p)
		}
		return ListVal(elems)
	case "splitAtFirst":
		sep := args[0].AsStr()
		idx := strings.Index(on.AsStr(), sep)
		if idx == -1 {
			return ListVal([]Value{StrVal(on.AsStr()), StrVal("")})
		}
		return ListVal([]Value{StrVal(on.AsStr()[:idx]), StrVal(on.AsStr()[idx+len(sep):])})
	case "splitAtAny":
		haystack := on.AsStr()
		var result []Value
		remaining := haystack
		for _, v := range *args[0].AsList() {
			sep := v.AsStr()
			parts := strings.Split(remaining, sep)
			if len(parts) > 1 {
				for k, p := range parts {
					if k < len(parts)-1 {
						result = append(result, StrVal(p))
					} else {
						remaining = p
					}
				}
			}
		}
		result = append(result, StrVal(remaining))
		return ListVal(result)
	case "splitAtFirstOfAny":
		haystack := on.AsStr()
		earliest := len(haystack)
		sep := ""
		for _, v := range *args[0].AsList() {
			s := v.AsStr()
			idx := strings.Index(haystack, s)
			if idx != -1 && idx < earliest {
				earliest = idx
				sep = s
			}
		}
		if sep == "" {
			return ListVal([]Value{StrVal(haystack), StrVal("")})
		}
		return ListVal([]Value{StrVal(haystack[:earliest]), StrVal(haystack[earliest+len(sep):])})
	case "splitAtSpaces":
		parts := strings.Fields(on.AsStr())
		elems := make([]Value, len(parts))
		for k, p := range parts {
			elems[k] = StrVal(p)
		}
		return ListVal(elems)
	case "reverse":
		runes := []rune(on.AsStr())
		for a, b := 0, len(runes)-1; a < b; a, b = a+1, b-1 {
			runes[a], runes[b] = runes[b], runes[a]
		}
		return StrVal(string(runes))
	case "csvRowToList":
		records := parseCSVRow(on.AsStr())
		elems := make([]Value, len(records))
		for k, rec := range records {
			elems[k] = StrVal(rec)
		}
		return ListVal(elems)
	case "csvTableToList":
		allRecords := parseCSVTable(on.AsStr())
		rows := make([]Value, len(allRecords))
		for k, rec := range allRecords {
			elems := make([]Value, len(rec))
			for j, cell := range rec {
				elems[j] = StrVal(cell)
			}
			rows[k] = ListVal(elems)
		}
		return ListVal(rows)
	case "segment":
		s := []rune(on.AsStr())
		from := int(args[0].AsNum()) - 1 // 1-based
		length := int(args[1].AsNum())
		if from < 0 {
			from = 0
		}
		end := from + length
		if end > len(s) {
			end = len(s)
		}
		return StrVal(string(s[from:end]))
	case "replace":
		return StrVal(strings.ReplaceAll(on.AsStr(), args[0].AsStr(), args[1].AsStr()))
	case "replaceFrom":
		result := on.AsStr()
		d := args[0].AsDict()
		for _, entry := range d.entries {
			result = strings.ReplaceAll(result, entry.Key, entry.Val.AsStr())
		}
		return StrVal(result)
	case "replaceFromLongestFirst":
		result := on.AsStr()
		d := args[0].AsDict()
		keys := make([]string, len(d.entries))
		for k, entry := range d.entries {
			keys[k] = entry.Key
		}
		insertionSort(len(keys), func(a, b int) bool { return len(keys[a]) > len(keys[b]) }, func(a, b int) { keys[a], keys[b] = keys[b], keys[a] })
		for _, key := range keys {
			if val, ok := d.Get(key); ok {
				result = strings.ReplaceAll(result, key, val.AsStr())
			}
		}
		return StrVal(result)

	// ============ List methods ============
	case "listLen":
		return NumVal(float64(len(*on.AsList())))
	case "add":
		list := on.AsList()
		*list = append(*list, args...)
		return VoidVal()
	case "containsItem":
		for _, v := range *on.AsList() {
			if DeepEqual(v, args[0]) {
				return BoolVal(true)
			}
		}
		return BoolVal(false)
	case "indexOf":
		for k, v := range *on.AsList() {
			if DeepEqual(v, args[0]) {
				return NumVal(float64(k + 1)) // 1-based
			}
		}
		return NumVal(0)
	case "insert":
		list := on.AsList()
		idx := int(args[0].AsNum()) - 1 // 1-based
		val := args[1]
		*list = append(*list, NullVal())
		copy((*list)[idx+1:], (*list)[idx:])
		(*list)[idx] = val
		return VoidVal()
	case "remove":
		list := on.AsList()
		idx := int(args[0].AsNum()) - 1 // 1-based
		*list = append((*list)[:idx], (*list)[idx+1:]...)
		return VoidVal()
	case "appendList":
		list := on.AsList()
		other := args[0].AsList()
		*list = append(*list, *other...)
		return VoidVal()
	case "lookupInPairs":
		key := args[0].AsStr()
		notFound := args[1]
		for _, v := range *on.AsList() {
			pair := *v.AsList()
			if len(pair) >= 2 && pair[0].AsStr() == key {
				return pair[1]
			}
		}
		return notFound
	case "join":
		parts := make([]string, len(*on.AsList()))
		for k, v := range *on.AsList() {
			parts[k] = v.AsStr()
		}
		return StrVal(strings.Join(parts, args[0].AsStr()))
	case "slice":
		list := *on.AsList()
		idx1 := int(args[0].AsNum()) - 1 // 1-based
		idx2 := int(args[1].AsNum())     // 1-based inclusive → exclusive
		if idx1 < 0 {
			idx1 = 0
		}
		if idx2 > len(list) {
			idx2 = len(list)
		}
		return ListVal(list[idx1:idx2])
	case "random":
		list := *on.AsList()
		if len(list) == 0 {
			panic("random: empty list")
		}
		return list[rngIntn(len(list))]
	case "reverseList":
		list := *on.AsList()
		cp := make([]Value, len(list))
		for k, v := range list {
			cp[len(list)-1-k] = v
		}
		return ListVal(cp)
	case "toCsvRow":
		list := *on.AsList()
		fields := make([]string, len(list))
		for k, v := range list {
			fields[k] = v.AsStr()
		}
		return StrVal(formatCSVRow(fields))
	case "toCsvTable":
		var sb strings.Builder
		for k, row := range *on.AsList() {
			rowList := *row.AsList()
			fields := make([]string, len(rowList))
			for j, v := range rowList {
				fields[j] = v.AsStr()
			}
			if k > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(formatCSVRow(fields))
		}
		return StrVal(sb.String())
	case "sort":
		list := *on.AsList()
		cp := make([]Value, len(list))
		copy(cp, list)
		sortValues(cp)
		return ListVal(cp)
	case "allButFirst":
		list := *on.AsList()
		if len(list) == 0 {
			return ListVal(nil)
		}
		return ListVal(list[1:])
	case "allButLast":
		list := *on.AsList()
		if len(list) == 0 {
			return ListVal(nil)
		}
		return ListVal(list[:len(list)-1])
	case "pairsToDict":
		d := NewOrderedDict()
		for _, v := range *on.AsList() {
			pair := *v.AsList()
			if len(pair) >= 2 {
				d.Set(pair[0].AsStr(), pair[1])
			}
		}
		return DictVal(d)

	// ============ Dict methods ============
	case "dictLen":
		return NumVal(float64(on.AsDict().Len()))
	case "get":
		d := on.AsDict()
		if val, ok := d.Get(args[0].AsStr()); ok {
			return val
		}
		return args[1] // notFound
	case "set":
		on.AsDict().Set(args[0].AsStr(), args[1])
		return VoidVal()
	case "delete":
		on.AsDict().Delete(args[0].AsStr())
		return VoidVal()
	case "getAtPath":
		return dictGetAtPath(on.AsDict(), args[0].AsList(), args[1])
	case "setAtPath":
		dictSetAtPath(on.AsDict(), args[0].AsList(), args[1])
		return VoidVal()
	case "containsKey":
		return BoolVal(on.AsDict().ContainsKey(args[0].AsStr()))
	case "mergeInto":
		dst := on.AsDict()
		src := args[0].AsDict()
		for _, entry := range src.entries {
			dst.Set(entry.Key, entry.Val)
		}
		return DictVal(dst)
	case "walkTree":
		// Simplified: treat path as a list of keys
		path := args[0].AsList()
		var cur Value = on
		for _, key := range *path {
			if cur.Type() == Dict {
				if v, ok := cur.AsDict().Get(key.AsStr()); ok {
					cur = v
				} else {
					return NullVal()
				}
			} else {
				return NullVal()
			}
		}
		return cur
	case "keys":
		return ListVal(on.AsDict().Keys())
	case "values":
		return ListVal(on.AsDict().Values())
	case "toPairs":
		d := on.AsDict()
		pairs := make([]Value, len(d.entries))
		for k, entry := range d.entries {
			pairs[k] = ListVal([]Value{StrVal(entry.Key), entry.Val})
		}
		return ListVal(pairs)

	default:
		panic("unknown method: " + e.Name)
	}
}

func dictGetAtPath(d *OrderedDict, path *[]Value, notFound Value) Value {
	var cur Value = DictVal(d)
	for _, key := range *path {
		if cur.Type() != Dict {
			return notFound
		}
		v, ok := cur.AsDict().Get(key.AsStr())
		if !ok {
			return notFound
		}
		cur = v
	}
	return cur
}

func dictSetAtPath(d *OrderedDict, path *[]Value, val Value) {
	if len(*path) == 0 {
		return
	}
	keys := *path
	var cur Value = DictVal(d)
	for _, key := range keys[:len(keys)-1] {
		if cur.Type() != Dict {
			return
		}
		v, ok := cur.AsDict().Get(key.AsStr())
		if !ok {
			nd := NewOrderedDict()
			cur.AsDict().Set(key.AsStr(), DictVal(nd))
			cur = DictVal(nd)
		} else {
			cur = v
		}
	}
	if cur.Type() == Dict {
		cur.AsDict().Set(keys[len(keys)-1].AsStr(), val)
	}
}

// evalTransformer handles list lambda operations.
func (i *Interpreter) evalTransformer(e *astlist.Transformer) Value {
	list := i.Eval(e.List).AsList()
	outerEnv := i.currEnv

	switch e.Name {
	case "map":
		varName := e.Names[0]
		result := make([]Value, len(*list))
		for k, elem := range *list {
			lambdaEnv := NewEnv(outerEnv)
			lambdaEnv.Define(varName, elem)
			result[k] = i.inEnv(lambdaEnv, func() Value { return i.Eval(e.Transformer) })
		}
		return ListVal(result)

	case "filter":
		varName := e.Names[0]
		var result []Value
		for _, elem := range *list {
			lambdaEnv := NewEnv(outerEnv)
			lambdaEnv.Define(varName, elem)
			if i.inEnv(lambdaEnv, func() Value { return i.Eval(e.Transformer) }).AsBool() {
				result = append(result, elem)
			}
		}
		return ListVal(result)

	case "reduce":
		varName := e.Names[0] // element
		accName := e.Names[1] // accumulator
		acc := i.Eval(e.Args[0])
		for _, elem := range *list {
			lambdaEnv := NewEnv(outerEnv)
			lambdaEnv.Define(varName, elem)
			lambdaEnv.Define(accName, acc)
			acc = i.inEnv(lambdaEnv, func() Value { return i.Eval(e.Transformer) })
		}
		return acc

	case "sort":
		varM := e.Names[0]
		varN := e.Names[1]
		cp := make([]Value, len(*list))
		copy(cp, *list)
		insertionSort(len(cp), func(a, b int) bool {
			lambdaEnv := NewEnv(outerEnv)
			lambdaEnv.Define(varM, cp[a])
			lambdaEnv.Define(varN, cp[b])
			return i.inEnv(lambdaEnv, func() Value { return i.Eval(e.Transformer) }).AsBool()
		}, func(a, b int) { cp[a], cp[b] = cp[b], cp[a] })
		return ListVal(cp)

	case "sortByKey":
		varName := e.Names[0]
		cp := make([]Value, len(*list))
		copy(cp, *list)
		insertionSort(len(cp), func(a, b int) bool {
			envA := NewEnv(outerEnv)
			envA.Define(varName, cp[a])
			keyA := i.inEnv(envA, func() Value { return i.Eval(e.Transformer) })
			envB := NewEnv(outerEnv)
			envB.Define(varName, cp[b])
			keyB := i.inEnv(envB, func() Value { return i.Eval(e.Transformer) })
			na, aOk := CoerceNum(keyA)
			nb, bOk := CoerceNum(keyB)
			if aOk && bOk {
				return na < nb
			}
			return keyA.AsStr() < keyB.AsStr()
		}, func(a, b int) { cp[a], cp[b] = cp[b], cp[a] })
		return ListVal(cp)

	case "min":
		varM := e.Names[0]
		varN := e.Names[1]
		if len(*list) == 0 {
			panic("min transformer: empty list")
		}
		best := (*list)[0]
		for _, elem := range (*list)[1:] {
			lambdaEnv := NewEnv(outerEnv)
			lambdaEnv.Define(varM, best)
			lambdaEnv.Define(varN, elem)
			if i.inEnv(lambdaEnv, func() Value { return i.Eval(e.Transformer) }).AsBool() {
				best = elem
			}
		}
		return best

	case "max":
		varM := e.Names[0]
		varN := e.Names[1]
		if len(*list) == 0 {
			panic("max transformer: empty list")
		}
		best := (*list)[0]
		for _, elem := range (*list)[1:] {
			lambdaEnv := NewEnv(outerEnv)
			lambdaEnv.Define(varM, best)
			lambdaEnv.Define(varN, elem)
			if i.inEnv(lambdaEnv, func() Value { return i.Eval(e.Transformer) }).AsBool() {
				best = elem
			}
		}
		return best

	default:
		panic("unknown list transformer: " + e.Name)
	}
}
