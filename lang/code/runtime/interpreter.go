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
	"math"
	"strconv"
	"strings"
)

type Procedure struct {
	params   []string
	voidBody []ast.Expr // for void procedures
	retExpr  ast.Expr   // for returning procedures
}

// Interpreter implements Visitor and holds the runtime state.
type Interpreter struct {
	globalEnv  *Env
	currEnv    *Env
	procedures map[string]*Procedure
}

func NewInterpreter() *Interpreter {
	env := NewEnv(nil)
	return &Interpreter{
		globalEnv:  env,
		currEnv:    env,
		procedures: make(map[string]*Procedure),
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
			i.procedures[n.Name] = &Procedure{params: n.Parameters, voidBody: n.Body}
		case *procedures.RetProcedure:
			i.procedures[n.Name] = &Procedure{params: n.Parameters, retExpr: n.Result}
		case *variables.Global:
			val := i.Eval(n.Value)
			i.globalEnv.Define(n.Name, val)
		}
	}
	var last Value = NullVal()
	for _, e := range exprs {
		switch e.(type) {
		case *procedures.VoidProcedure, *procedures.RetProcedure, *variables.Global:
		default:
			last = i.Eval(e)
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
			i.procedures[n.Name] = &Procedure{params: n.Parameters, voidBody: n.Body}
		case *procedures.RetProcedure:
			i.procedures[n.Name] = &Procedure{params: n.Parameters, retExpr: n.Result}
		case *variables.Global:
			val := i.Eval(n.Value)
			i.globalEnv.Define(n.Name, val)
		}
	}
	// Second pass: execute non-definition top-level statements.
	for _, e := range exprs {
		switch e.(type) {
		case *procedures.VoidProcedure, *procedures.RetProcedure, *variables.Global:
			// already handled above
		default:
			i.Eval(e)
		}
	}
}

func (i *Interpreter) Eval(expr ast.Expr) Value {
	switch e := expr.(type) {
	// fundamentals
	case *fundamentals.Boolean:
		return BoolVal(e.Value)
	case *fundamentals.Not:
		return BoolVal(!i.Eval(e.Expr).AsBool())
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
		return ListVal(i.evalExprs(e.Elements))
	case *fundamentals.Dictionary:
		return i.dictionary(e)
	case *fundamentals.Pair:
		// Bare pair evaluates to a two-element list [key, value].
		return ListVal([]Value{i.Eval(e.Key), i.Eval(e.Value)})
	case *fundamentals.SmartBody:
		return i.execBody(e.Body)
	case *fundamentals.Component:
		stub("component reference @" + e.Name + " (" + e.Type + ")")
		return NullVal()

	// falcon specific features
	case *common.EmptySocket:
		return NumVal(0)
	case *common.Transform:
		return i.Eval(e.On) // for "obfuscate" unwrapping

	case *common.BinaryExpr:
		return i.binary(e)
	case *common.FuncCall:
		return i.evalFuncCall(e)
	case *common.Question:
		return i.question(e)
	case *astmethod.Call:
		return i.evalMethodCall(e)

	// Control blocks
	case *control.If:
		return i.ifExpr(e)
	case *control.SimpleIf:
		return i.simpleIfExpr(e)
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
		return i.Eval(e.Result)

	// Variable blocks
	case *variables.Get:
		if e.Global {
			return i.currEnv.GetGlobal(e.Name)
		}
		return i.currEnv.Get(e.Name)
	case *variables.Set:
		val := i.Eval(e.Expr)
		if e.Global {
			i.currEnv.SetGlobal(e.Name, val)
		} else {
			i.currEnv.Set(e.Name, val)
		}
		return NullVal()
	case *variables.Global:
		val := i.Eval(e.Value)
		i.currEnv.root().Define(e.Name, val)
		return NullVal()
	case *variables.Var:
		return i.evalVar(e)
	case *variables.SimpleVar:
		return i.evalSimpleVar(e)
	case *variables.VarResult:
		return i.evalVarResult(e)

	// Procedure definitions
	case *procedures.VoidProcedure:
		i.procedures[e.Name] = &Procedure{params: e.Parameters, voidBody: e.Body}
		return NullVal()
	case *procedures.RetProcedure:
		i.procedures[e.Name] = &Procedure{params: e.Parameters, retExpr: e.Result}
		return NullVal()
	case *procedures.Call:
		return i.evalProcedureCall(e)

	// List manipulation
	case *astlist.Get:
		list := i.Eval(e.List).AsList()
		idx := int(i.Eval(e.Index).AsNum())
		if idx < 1 || idx > len(*list) {
			panic("list index " + strconv.Itoa(idx) + " out of bounds (len=" + strconv.Itoa(len(*list)) + ")")
		}
		return (*list)[idx-1]
	case *astlist.Set:
		list := i.Eval(e.List).AsList()
		idx := int(i.Eval(e.Index).AsNum())
		val := i.Eval(e.Value)
		if idx < 1 || idx > len(*list) {
			panic("list index " + strconv.Itoa(idx) + " out of bounds (len=" + strconv.Itoa(len(*list)) + ")")
		}
		(*list)[idx-1] = val
		return NullVal()
	case *astlist.Transformer:
		return i.evalTransformer(e)

	// Component blocks
	case *components.Event:
		stub("event handler " + e.ComponentName + "." + e.Event)
		return NullVal()
	case *components.GenericEvent:
		stub("generic event handler " + e.ComponentType + "." + e.Event)
		return NullVal()
	case *components.MethodCall:
		stub("component method " + e.ComponentName + "." + e.Method + "(...)")
		return NullVal()
	case *components.GenericMethodCall:
		stub("generic component method " + e.ComponentType + "." + e.Method + "(...)")
		return NullVal()
	case *components.PropertyGet:
		stub("property get " + e.ComponentName + "." + e.Property)
		return NullVal()
	case *components.GenericPropertyGet:
		stub("generic property get " + e.ComponentType + "." + e.Property)
		return NullVal()
	case *components.PropertySet:
		stub("property set " + e.ComponentName + "." + e.Property)
		return NullVal()
	case *components.GenericPropertySet:
		stub("generic property set " + e.ComponentType + "." + e.Property)
		return NullVal()
	case *components.EveryComponent:
		stub("every(" + e.Type + ")")
		return EmptyList()

	default:
		panic("unknown AST node type")
	}
}

func (i *Interpreter) dictionary(e *fundamentals.Dictionary) Value {
	d := NewOrderedDict()
	for _, el := range e.Elements {
		pair := i.Eval(el).AsList()
		if len(*pair) < 2 {
			panic("dictionary entry must be a pair (two-element list)")
		}
		d.Set((*pair)[0].AsStr(), (*pair)[1])
	}
	return DictVal(d)
}

func (i *Interpreter) binary(e *common.BinaryExpr) Value {
	// Short-circuit logical operators
	switch e.Operator {
	case lex.LogicAnd:
		for _, op := range e.Operands {
			if !i.Eval(op).AsBool() {
				return BoolVal(false)
			}
		}
		return BoolVal(true)
	case lex.LogicOr:
		for _, op := range e.Operands {
			if i.Eval(op).AsBool() {
				return BoolVal(true)
			}
		}
		return BoolVal(false)
	}

	vals := i.evalExprs(e.Operands)
	switch e.Operator {
	case lex.Plus:
		result := vals[0].AsNum()
		for _, v := range vals[1:] {
			result += v.AsNum()
		}
		return NumVal(result)
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

	// runs deep comparison for logical equality
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

	// text join operation
	case lex.Underscore:
		var sb strings.Builder
		for _, v := range vals {
			sb.WriteString(v.AsStr())
		}
		return StrVal(sb.String())

	default:
		panic("unknown binary operator: " + strconv.Itoa(int(e.Operator)))
	}
}

func (i *Interpreter) question(e *common.Question) Value {
	v := i.Eval(e.On)
	switch e.Question {
	case "number":
		_, ok := CoerceNum(v)
		return BoolVal(ok)
	case "base10":
		return BoolVal(isBase10(v.AsStr()))
	case "hexa":
		return BoolVal(isHex(v.AsStr()))
	case "bin":
		return BoolVal(isBinary(v.AsStr()))
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

func (i *Interpreter) ifExpr(e *control.If) Value {
	for k, cond := range e.Conditions {
		if i.Eval(cond).AsBool() {
			return i.execBody(e.Bodies[k])
		}
	}
	if e.ElseBody != nil {
		return i.execBody(e.ElseBody)
	}
	return NullVal()
}

func (i *Interpreter) simpleIfExpr(e *control.SimpleIf) Value {
	if i.Eval(e.Condition()).AsBool() {
		return i.execBody(e.Then())
	}
	if els := e.Else(); len(els) > 0 {
		return i.execBody(els)
	}
	return NullVal()
}

func (i *Interpreter) evalFor(e *control.For) Value {
	from := i.Eval(e.From).AsNum()
	to := i.Eval(e.To).AsNum()
	by := i.Eval(e.By).AsNum()
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
		for i.Eval(e.Condition).AsBool() {
			last = i.execBody(e.Body)
		}
	}()
	return last
}

func (i *Interpreter) evalEach(e *control.Each) Value {
	list := i.Eval(e.Iterable).AsList()
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
	d := i.Eval(e.Iterable).AsDict()
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
		val := i.Eval(e.Values[k]) // values evaluated in parent scope
		childEnv.Define(name, val)
	}
	return i.inEnv(childEnv, func() Value { return i.execBody(e.Body) })
}

func (i *Interpreter) evalSimpleVar(e *variables.SimpleVar) Value {
	val := i.Eval(e.Value)
	childEnv := NewEnv(i.currEnv)
	childEnv.Define(e.Name, val)
	return i.inEnv(childEnv, func() Value { return i.execBody(e.Body) })
}

func (i *Interpreter) evalVarResult(e *variables.VarResult) Value {
	childEnv := NewEnv(i.currEnv)
	for k, name := range e.Names {
		val := i.Eval(e.Values[k])
		childEnv.Define(name, val)
	}
	return i.inEnv(childEnv, func() Value { return i.Eval(e.Result) })
}

// --- Procedure calls ---

func (i *Interpreter) evalProcedureCall(e *procedures.Call) Value {
	proc, ok := i.procedures[e.Name]
	if !ok {
		panic("undefined procedure: " + e.Name)
	}
	// Evaluate arguments in the current (caller) env before switching scope.
	argVals := make([]Value, len(proc.params))
	for k := range proc.params {
		argVals[k] = i.Eval(e.Arguments[k])
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
				result = i.Eval(proc.retExpr)
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
		last = i.Eval(expr)
	}
	return last
}

// evalExprs evaluates each expression in exprs and returns the results as a []Value slice.
func (i *Interpreter) evalExprs(exprs []ast.Expr) []Value {
	vals := make([]Value, len(exprs))
	for k, expr := range exprs {
		vals[k] = i.Eval(expr)
	}
	return vals
}
