package runtime

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
	"Falcon/code/ast/components"
	"Falcon/code/ast/control"
	"Falcon/code/ast/fundamentals"
	astlist "Falcon/code/ast/list"
	astmethod "Falcon/code/ast/method"
	"Falcon/code/ast/procedures"
	"Falcon/code/ast/variables"
	"Falcon/code/lex"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// storedProc holds a parsed procedure definition for later invocation.
type storedProc struct {
	params   []string
	voidBody []ast.Expr // set for void procedures
	retExpr  ast.Expr   // set for returning procedures
}

// Interpreter implements Visitor and holds the runtime state.
type Interpreter struct {
	globalEnv *Env
	currEnv   *Env
	procs     map[string]*storedProc
}

func NewInterpreter() *Interpreter {
	env := NewEnv(nil)
	return &Interpreter{
		globalEnv: env,
		currEnv:   env,
		procs:     make(map[string]*storedProc),
	}
}

// inEnv temporarily switches currEnv to env, calls fn, then restores currEnv.
// The restore happens even if fn panics, so scoping is always correct.
func (i *Interpreter) inEnv(env *Env, fn func() Value) Value {
	prev := i.currEnv
	i.currEnv = env
	defer func() { i.currEnv = prev }()
	return fn()
}

// RunGetLast runs the program like Run but returns the value of the last
// non-definition top-level expression. Returns NullVal if none.
func (i *Interpreter) RunGetLast(exprs []ast.Expr) Value {
	for _, e := range exprs {
		switch n := e.(type) {
		case *procedures.VoidProcedure:
			i.procs[n.Name] = &storedProc{params: n.Parameters, voidBody: n.Body}
		case *procedures.RetProcedure:
			i.procs[n.Name] = &storedProc{params: n.Parameters, retExpr: n.Result}
		case *variables.Global:
			val := i.eval(n.Value)
			i.globalEnv.Define(n.Name, val)
		}
	}
	var last Value = NullVal()
	for _, e := range exprs {
		switch e.(type) {
		case *procedures.VoidProcedure, *procedures.RetProcedure, *variables.Global:
		default:
			last = i.eval(e)
		}
	}
	return last
}

// Run executes a top-level list of expressions (a full program).
func (i *Interpreter) Run(exprs []ast.Expr) {
	// First pass: register all procedures and globals so forward calls work.
	for _, e := range exprs {
		switch n := e.(type) {
		case *procedures.VoidProcedure:
			i.procs[n.Name] = &storedProc{params: n.Parameters, voidBody: n.Body}
		case *procedures.RetProcedure:
			i.procs[n.Name] = &storedProc{params: n.Parameters, retExpr: n.Result}
		case *variables.Global:
			val := i.eval(n.Value)
			i.globalEnv.Define(n.Name, val)
		}
	}
	// Second pass: execute non-definition top-level statements.
	for _, e := range exprs {
		switch e.(type) {
		case *procedures.VoidProcedure, *procedures.RetProcedure, *variables.Global:
			// already handled above
		default:
			i.eval(e)
		}
	}
}

// Eval is the public entry point for evaluating a single expression.
func (i *Interpreter) Eval(expr ast.Expr) Value {
	return i.eval(expr)
}

func (i *Interpreter) eval(expr ast.Expr) Value {
	switch e := expr.(type) {
	// --- Fundamentals ---
	case *fundamentals.Boolean:
		return BoolVal(e.Value)
	case *fundamentals.Not:
		return BoolVal(!i.eval(e.Expr).AsBool())
	case *fundamentals.Number:
		f, err := strconv.ParseFloat(e.Content, 64)
		if err != nil {
			panic("invalid number literal: " + e.Content)
		}
		return NumVal(f)
	case *fundamentals.Text:
		return StrVal(e.Content)
	case *fundamentals.Color:
		return ColorVal(e.Hex)
	case *fundamentals.List:
		elems := make([]Value, len(e.Elements))
		for k, el := range e.Elements {
			elems[k] = i.eval(el)
		}
		return ListVal(elems)
	case *fundamentals.Dictionary:
		return i.evalDictionary(e)
	case *fundamentals.Pair:
		// Bare pair evaluates to a two-element list [key, value].
		return ListVal([]Value{i.eval(e.Key), i.eval(e.Value)})
	case *fundamentals.SmartBody:
		return i.execBody(e.Body)
	case *fundamentals.Component:
		fmt.Printf("[stub] component reference @%s (%s) is not supported outside App Inventor\n", e.Name, e.Type)
		return NullVal()

	case *common.EmptySocket:
		return NumVal(0)
	case *common.Transform:
		return i.eval(e.On)

	// --- Common ---
	case *common.BinaryExpr:
		return i.evalBinary(e)
	case *common.FuncCall:
		return i.evalFuncCall(e)
	case *common.Question:
		return i.evalQuestion(e)

	// --- Control ---
	case *control.If:
		return i.evalIf(e)
	case *control.SimpleIf:
		return i.evalSimpleIf(e)
	case *control.For:
		return i.evalFor(e)
	case *control.While:
		return i.evalWhile(e)
	case *control.Each:
		return i.evalEach(e)
	case *control.EachPair:
		return i.evalEachPair(e)
	case *control.Break:
		panic(BreakSignal{})
	case *control.Do:
		i.execBody(e.Body)
		return i.eval(e.Result)

	// --- Variables ---
	case *variables.Get:
		if e.Global {
			return i.currEnv.GetGlobal(e.Name)
		}
		return i.currEnv.Get(e.Name)
	case *variables.Set:
		val := i.eval(e.Expr)
		if e.Global {
			i.currEnv.SetGlobal(e.Name, val)
		} else {
			i.currEnv.Set(e.Name, val)
		}
		return NullVal()
	case *variables.Global:
		val := i.eval(e.Value)
		i.currEnv.root().Define(e.Name, val)
		return NullVal()
	case *variables.Var:
		return i.evalVar(e)
	case *variables.SimpleVar:
		return i.evalSimpleVar(e)
	case *variables.VarResult:
		return i.evalVarResult(e)

	// --- Procedures ---
	case *procedures.VoidProcedure:
		i.procs[e.Name] = &storedProc{params: e.Parameters, voidBody: e.Body}
		return NullVal()
	case *procedures.RetProcedure:
		i.procs[e.Name] = &storedProc{params: e.Parameters, retExpr: e.Result}
		return NullVal()
	case *procedures.Call:
		return i.evalProcedureCall(e)

	// --- Methods ---
	case *astmethod.Call:
		return i.evalMethodCall(e)

	// --- List indexing ---
	case *astlist.Get:
		list := i.eval(e.List).AsList()
		idx := int(i.eval(e.Index).AsNum())
		if idx < 1 || idx > len(*list) {
			panic(fmt.Sprintf("list index %d out of bounds (len=%d)", idx, len(*list)))
		}
		return (*list)[idx-1]
	case *astlist.Set:
		list := i.eval(e.List).AsList()
		idx := int(i.eval(e.Index).AsNum())
		val := i.eval(e.Value)
		if idx < 1 || idx > len(*list) {
			panic(fmt.Sprintf("list index %d out of bounds (len=%d)", idx, len(*list)))
		}
		(*list)[idx-1] = val
		return NullVal()
	case *astlist.Transformer:
		return i.evalTransformer(e)

	// --- App Inventor component stubs ---
	case *components.Event:
		fmt.Printf("[stub] event handler %s.%s is not supported outside App Inventor\n", e.ComponentName, e.Event)
		return NullVal()
	case *components.GenericEvent:
		fmt.Printf("[stub] generic event handler %s.%s is not supported outside App Inventor\n", e.ComponentType, e.Event)
		return NullVal()
	case *components.MethodCall:
		fmt.Printf("[stub] component method %s.%s(...) is not supported outside App Inventor\n", e.ComponentName, e.Method)
		return NullVal()
	case *components.GenericMethodCall:
		fmt.Printf("[stub] generic component method %s.%s(...) is not supported outside App Inventor\n", e.ComponentType, e.Method)
		return NullVal()
	case *components.PropertyGet:
		fmt.Printf("[stub] property get %s.%s is not supported outside App Inventor\n", e.ComponentName, e.Property)
		return NullVal()
	case *components.GenericPropertyGet:
		fmt.Printf("[stub] generic property get %s.%s is not supported outside App Inventor\n", e.ComponentType, e.Property)
		return NullVal()
	case *components.PropertySet:
		fmt.Printf("[stub] property set %s.%s is not supported outside App Inventor\n", e.ComponentName, e.Property)
		return NullVal()
	case *components.GenericPropertySet:
		fmt.Printf("[stub] generic property set %s.%s is not supported outside App Inventor\n", e.ComponentType, e.Property)
		return NullVal()
	case *components.EveryComponent:
		fmt.Printf("[stub] every(%s) is not supported outside App Inventor\n", e.Type)
		return EmptyList()

	default:
		panic(fmt.Sprintf("unknown AST node type: %T", expr))
	}
}

// --- Dictionary literal ---

func (i *Interpreter) evalDictionary(e *fundamentals.Dictionary) Value {
	d := NewOrderedDict()
	for _, el := range e.Elements {
		pair, ok := el.(*fundamentals.Pair)
		if !ok {
			panic("dictionary element is not a Pair")
		}
		key := i.eval(pair.Key).AsStr()
		val := i.eval(pair.Value)
		d.Set(key, val)
	}
	return DictVal(d)
}

// --- Binary expressions ---

func (i *Interpreter) evalBinary(e *common.BinaryExpr) Value {
	// Short-circuit logical operators
	switch e.Operator {
	case lex.LogicAnd:
		for _, op := range e.Operands {
			if !i.eval(op).AsBool() {
				return BoolVal(false)
			}
		}
		return BoolVal(true)
	case lex.LogicOr:
		for _, op := range e.Operands {
			if i.eval(op).AsBool() {
				return BoolVal(true)
			}
		}
		return BoolVal(false)
	}

	vals := make([]Value, len(e.Operands))
	for k, op := range e.Operands {
		vals[k] = i.eval(op)
	}

	switch e.Operator {
	case lex.Plus:
		// If all operands are numeric-coercible, add as numbers; otherwise join as strings
		allNum := true
		for _, v := range vals {
			if _, ok := TryNum(v); !ok {
				allNum = false
				break
			}
		}
		if allNum {
			result := vals[0].AsNum()
			for _, v := range vals[1:] {
				result += v.AsNum()
			}
			return NumVal(result)
		}
		// Fall back to text join
		var sb strings.Builder
		for _, v := range vals {
			sb.WriteString(v.AsStr())
		}
		return StrVal(sb.String())
	case lex.Dash:
		return NumVal(vals[0].AsNum() - vals[1].AsNum())
	case lex.Times:
		result := vals[0].AsNum()
		for _, v := range vals[1:] {
			result *= v.AsNum()
		}
		return NumVal(result)
	case lex.Slash:
		return NumVal(vals[0].AsNum() / vals[1].AsNum())
	case lex.Remainder:
		a, b := vals[0].AsNum(), vals[1].AsNum()
		return NumVal(math.Mod(a, b))
	case lex.Power:
		return NumVal(math.Pow(vals[0].AsNum(), vals[1].AsNum()))

	case lex.BitwiseAnd:
		result := int64(vals[0].AsNum())
		for _, v := range vals[1:] {
			result &= int64(v.AsNum())
		}
		return NumVal(float64(result))
	case lex.BitwiseOr:
		result := int64(vals[0].AsNum())
		for _, v := range vals[1:] {
			result |= int64(v.AsNum())
		}
		return NumVal(float64(result))
	case lex.BitwiseXor:
		result := int64(vals[0].AsNum())
		for _, v := range vals[1:] {
			result ^= int64(v.AsNum())
		}
		return NumVal(float64(result))

	case lex.Equals:
		return BoolVal(DeepEqual(vals[0], vals[1]))
	case lex.NotEquals:
		return BoolVal(!DeepEqual(vals[0], vals[1]))

	case lex.LessThan:
		return BoolVal(vals[0].AsNum() < vals[1].AsNum())
	case lex.LessThanEqual:
		return BoolVal(vals[0].AsNum() <= vals[1].AsNum())
	case lex.GreatThan:
		return BoolVal(vals[0].AsNum() > vals[1].AsNum())
	case lex.GreaterThanEqual:
		return BoolVal(vals[0].AsNum() >= vals[1].AsNum())

	case lex.TextEquals:
		return BoolVal(vals[0].AsStr() == vals[1].AsStr())
	case lex.TextNotEquals:
		return BoolVal(vals[0].AsStr() != vals[1].AsStr())
	case lex.TextLessThan:
		return BoolVal(vals[0].AsStr() < vals[1].AsStr())
	case lex.TextGreaterThan:
		return BoolVal(vals[0].AsStr() > vals[1].AsStr())

	case lex.Underscore:
		var sb strings.Builder
		for _, v := range vals {
			sb.WriteString(v.AsStr())
		}
		return StrVal(sb.String())

	default:
		panic(fmt.Sprintf("unknown binary operator: %v", e.Operator))
	}
}

// --- Question (? operator) ---

func (i *Interpreter) evalQuestion(e *common.Question) Value {
	v := i.eval(e.On)
	switch e.Question {
	case "number":
		_, ok := TryNum(v)
		return BoolVal(ok)
	case "base10":
		s := strings.TrimSpace(v.AsStr())
		for _, c := range s {
			if c < '0' || c > '9' {
				return BoolVal(false)
			}
		}
		return BoolVal(len(s) > 0)
	case "hexa":
		s := strings.TrimSpace(strings.ToLower(v.AsStr()))
		for _, c := range s {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				return BoolVal(false)
			}
		}
		return BoolVal(len(s) > 0)
	case "bin":
		s := strings.TrimSpace(v.AsStr())
		for _, c := range s {
			if c != '0' && c != '1' {
				return BoolVal(false)
			}
		}
		return BoolVal(len(s) > 0)
	case "text":
		return BoolVal(v.Type() == String)
	case "list":
		return BoolVal(v.Type() == List)
	case "dict":
		return BoolVal(v.Type() == Dict)
	case "emptyText":
		return BoolVal(v.Type() == String && v.strVal == "")
	case "emptyList":
		return BoolVal(v.Type() == List && len(*v.listVal) == 0)
	case "even":
		return BoolVal(int64(v.AsNum())%2 == 0)
	case "odd":
		return BoolVal(int64(v.AsNum())%2 != 0)
	default:
		panic("unknown ? question: " + e.Question)
	}
}

// --- Control flow ---

func (i *Interpreter) evalIf(e *control.If) Value {
	for k, cond := range e.Conditions {
		if i.eval(cond).AsBool() {
			return i.execBody(e.Bodies[k])
		}
	}
	if e.ElseBody != nil {
		return i.execBody(e.ElseBody)
	}
	return NullVal()
}

func (i *Interpreter) evalSimpleIf(e *control.SimpleIf) Value {
	if i.eval(e.Condition()).AsBool() {
		return i.execBody(e.Then())
	}
	if els := e.Else(); len(els) > 0 {
		return i.execBody(els)
	}
	return NullVal()
}

func (i *Interpreter) evalFor(e *control.For) Value {
	from := i.eval(e.From).AsNum()
	to := i.eval(e.To).AsNum()
	by := i.eval(e.By).AsNum()
	if by == 0 {
		panic("for loop step cannot be 0")
	}
	loopEnv := NewEnv(i.currEnv)
	var last Value = NullVal()
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(BreakSignal); !ok {
					panic(r)
				}
			}
		}()
		i.inEnv(loopEnv, func() Value {
			for cur := from; (by > 0 && cur <= to) || (by < 0 && cur >= to); cur += by {
				loopEnv.Define(e.IName, NumVal(cur))
				last = i.execBody(e.Body)
			}
			return NullVal()
		})
	}()
	return last
}

func (i *Interpreter) evalWhile(e *control.While) Value {
	var last Value = NullVal()
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(BreakSignal); !ok {
					panic(r)
				}
			}
		}()
		for i.eval(e.Condition).AsBool() {
			last = i.execBody(e.Body)
		}
	}()
	return last
}

func (i *Interpreter) evalEach(e *control.Each) Value {
	list := i.eval(e.Iterable).AsList()
	loopEnv := NewEnv(i.currEnv)
	loopEnv.Define(e.IName, NullVal())
	var last Value = NullVal()
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(BreakSignal); !ok {
					panic(r)
				}
			}
		}()
		i.inEnv(loopEnv, func() Value {
			for _, elem := range *list {
				loopEnv.Define(e.IName, elem)
				last = i.execBody(e.Body)
			}
			return NullVal()
		})
	}()
	return last
}

func (i *Interpreter) evalEachPair(e *control.EachPair) Value {
	d := i.eval(e.Iterable).AsDict()
	loopEnv := NewEnv(i.currEnv)
	loopEnv.Define(e.KeyName, NullVal())
	loopEnv.Define(e.ValueName, NullVal())
	var last Value = NullVal()
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(BreakSignal); !ok {
					panic(r)
				}
			}
		}()
		i.inEnv(loopEnv, func() Value {
			for _, entry := range d.entries {
				loopEnv.Define(e.KeyName, StrVal(entry.Key))
				loopEnv.Define(e.ValueName, entry.Val)
				last = i.execBody(e.Body)
			}
			return NullVal()
		})
	}()
	return last
}

// --- Variables ---

func (i *Interpreter) evalVar(e *variables.Var) Value {
	childEnv := NewEnv(i.currEnv)
	for k, name := range e.Names {
		val := i.eval(e.Values[k]) // values evaluated in parent scope
		childEnv.Define(name, val)
	}
	return i.inEnv(childEnv, func() Value { return i.execBody(e.Body) })
}

func (i *Interpreter) evalSimpleVar(e *variables.SimpleVar) Value {
	val := i.eval(e.Value)
	childEnv := NewEnv(i.currEnv)
	childEnv.Define(e.Name, val)
	return i.inEnv(childEnv, func() Value { return i.execBody(e.Body) })
}

func (i *Interpreter) evalVarResult(e *variables.VarResult) Value {
	childEnv := NewEnv(i.currEnv)
	for k, name := range e.Names {
		val := i.eval(e.Values[k])
		childEnv.Define(name, val)
	}
	return i.inEnv(childEnv, func() Value { return i.eval(e.Result) })
}

// --- Procedure calls ---

func (i *Interpreter) evalProcedureCall(e *procedures.Call) Value {
	proc, ok := i.procs[e.Name]
	if !ok {
		panic("undefined procedure: " + e.Name)
	}
	// Evaluate arguments in the current (caller) env before switching scope.
	argVals := make([]Value, len(proc.params))
	for k := range proc.params {
		argVals[k] = i.eval(e.Arguments[k])
	}
	callEnv := NewEnv(i.globalEnv)
	for k, param := range proc.params {
		callEnv.Define(param, argVals[k])
	}

	return i.inEnv(callEnv, func() Value {
		if proc.retExpr != nil {
			var result Value
			func() {
				defer func() {
					if r := recover(); r != nil {
						if rs, ok := r.(ReturnSignal); ok {
							result = rs.Val
						} else {
							panic(r)
						}
					}
				}()
				result = i.eval(proc.retExpr)
			}()
			return result
		}

		// void procedure
		var ret Value = NullVal()
		func() {
			defer func() {
				if r := recover(); r != nil {
					if rs, ok := r.(ReturnSignal); ok {
						ret = rs.Val
					} else {
						panic(r)
					}
				}
			}()
			i.execBody(proc.voidBody)
		}()
		return ret
	})
}

// --- execBody: run a slice of expressions, return value of last ---

func (i *Interpreter) execBody(body []ast.Expr) Value {
	if len(body) == 0 {
		return NullVal()
	}
	var last Value
	for _, expr := range body {
		last = i.eval(expr)
	}
	return last
}