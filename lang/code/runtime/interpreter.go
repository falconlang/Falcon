package runtime

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
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
	procs     map[string]*storedProc
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		globalEnv: NewEnv(nil),
		procs:     make(map[string]*storedProc),
	}
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
			val := i.evalInEnv(n.Value, i.globalEnv)
			i.globalEnv.Define(n.Name, val)
		}
	}
	// Second pass: execute non-definition top-level statements.
	for _, e := range exprs {
		switch e.(type) {
		case *procedures.VoidProcedure, *procedures.RetProcedure, *variables.Global:
			// already handled above
		default:
			i.evalInEnv(e, i.globalEnv)
		}
	}
}

// Eval dispatches to the correct eval method via type switch (visitor pattern).
func (i *Interpreter) Eval(expr ast.Expr) Value {
	return i.evalInEnv(expr, i.globalEnv)
}

func (i *Interpreter) evalInEnv(expr ast.Expr, env *Env) Value {
	switch e := expr.(type) {
	// --- Fundamentals ---
	case *fundamentals.Boolean:
		return BoolVal(e.Value)
	case *fundamentals.Not:
		return BoolVal(!i.evalInEnv(e.Expr, env).AsBool())
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
			elems[k] = i.evalInEnv(el, env)
		}
		return ListVal(elems)
	case *fundamentals.Dictionary:
		return i.evalDictionary(e, env)
	case *fundamentals.Pair:
		panic("bare Pair expression not expected at runtime")
	case *fundamentals.SmartBody:
		return i.execBody(e.Body, env)
	case *common.EmptySocket:
		return NumVal(0)
	case *common.Transform:
		return i.evalInEnv(e.On, env)

	// --- Common ---
	case *common.BinaryExpr:
		return i.evalBinary(e, env)
	case *common.FuncCall:
		return i.evalFuncCall(e, env)
	case *common.Question:
		return i.evalQuestion(e, env)

	// --- Control ---
	case *control.If:
		return i.evalIf(e, env)
	case *control.SimpleIf:
		return i.evalSimpleIf(e, env)
	case *control.For:
		return i.evalFor(e, env)
	case *control.While:
		return i.evalWhile(e, env)
	case *control.Each:
		return i.evalEach(e, env)
	case *control.EachPair:
		return i.evalEachPair(e, env)
	case *control.Break:
		panic(BreakSignal{})
	case *control.Do:
		i.execBody(e.Body, env)
		return i.evalInEnv(e.Result, env)

	// --- Variables ---
	case *variables.Get:
		if e.Global {
			return env.GetGlobal(e.Name)
		}
		return env.Get(e.Name)
	case *variables.Set:
		val := i.evalInEnv(e.Expr, env)
		if e.Global {
			env.SetGlobal(e.Name, val)
		} else {
			env.Set(e.Name, val)
		}
		return NullVal()
	case *variables.Global:
		val := i.evalInEnv(e.Value, env)
		env.root().Define(e.Name, val)
		return NullVal()
	case *variables.Var:
		return i.evalVar(e, env)
	case *variables.SimpleVar:
		return i.evalSimpleVar(e, env)
	case *variables.VarResult:
		return i.evalVarResult(e, env)

	// --- Procedures ---
	case *procedures.VoidProcedure:
		i.procs[e.Name] = &storedProc{params: e.Parameters, voidBody: e.Body}
		return NullVal()
	case *procedures.RetProcedure:
		i.procs[e.Name] = &storedProc{params: e.Parameters, retExpr: e.Result}
		return NullVal()
	case *procedures.Call:
		return i.evalProcedureCall(e, env)

	// --- Methods ---
	case *astmethod.Call:
		return i.evalMethodCall(e, env)

	// --- List indexing ---
	case *astlist.Get:
		list := i.evalInEnv(e.List, env).AsList()
		idx := int(i.evalInEnv(e.Index, env).AsNum())
		if idx < 1 || idx > len(*list) {
			panic(fmt.Sprintf("list index %d out of bounds (len=%d)", idx, len(*list)))
		}
		return (*list)[idx-1]
	case *astlist.Set:
		list := i.evalInEnv(e.List, env).AsList()
		idx := int(i.evalInEnv(e.Index, env).AsNum())
		val := i.evalInEnv(e.Value, env)
		if idx < 1 || idx > len(*list) {
			panic(fmt.Sprintf("list index %d out of bounds (len=%d)", idx, len(*list)))
		}
		(*list)[idx-1] = val
		return NullVal()
	case *astlist.Transformer:
		return i.evalTransformer(e, env)

	default:
		panic(fmt.Sprintf("unknown AST node type: %T", expr))
	}
}

// --- Dictionary literal ---

func (i *Interpreter) evalDictionary(e *fundamentals.Dictionary, env *Env) Value {
	d := NewOrderedDict()
	for _, el := range e.Elements {
		pair, ok := el.(*fundamentals.Pair)
		if !ok {
			panic("dictionary element is not a Pair")
		}
		key := i.evalInEnv(pair.Key, env).AsStr()
		val := i.evalInEnv(pair.Value, env)
		d.Set(key, val)
	}
	return DictVal(d)
}

// --- Binary expressions ---

func (i *Interpreter) evalBinary(e *common.BinaryExpr, env *Env) Value {
	// Short-circuit logical operators
	switch e.Operator {
	case lex.LogicAnd:
		for _, op := range e.Operands {
			if !i.evalInEnv(op, env).AsBool() {
				return BoolVal(false)
			}
		}
		return BoolVal(true)
	case lex.LogicOr:
		for _, op := range e.Operands {
			if i.evalInEnv(op, env).AsBool() {
				return BoolVal(true)
			}
		}
		return BoolVal(false)
	}

	vals := make([]Value, len(e.Operands))
	for k, op := range e.Operands {
		vals[k] = i.evalInEnv(op, env)
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

func (i *Interpreter) evalQuestion(e *common.Question, env *Env) Value {
	v := i.evalInEnv(e.On, env)
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

func (i *Interpreter) evalIf(e *control.If, env *Env) Value {
	for k, cond := range e.Conditions {
		if i.evalInEnv(cond, env).AsBool() {
			return i.execBody(e.Bodies[k], env)
		}
	}
	if e.ElseBody != nil {
		return i.execBody(e.ElseBody, env)
	}
	return NullVal()
}

func (i *Interpreter) evalSimpleIf(e *control.SimpleIf, env *Env) Value {
	if i.evalInEnv(e.Condition(), env).AsBool() {
		return i.execBody(e.Then(), env)
	}
	if els := e.Else(); len(els) > 0 {
		return i.execBody(els, env)
	}
	return NullVal()
}

func (i *Interpreter) evalFor(e *control.For, env *Env) Value {
	from := i.evalInEnv(e.From, env).AsNum()
	to := i.evalInEnv(e.To, env).AsNum()
	by := i.evalInEnv(e.By, env).AsNum()
	if by == 0 {
		panic("for loop step cannot be 0")
	}
	loopEnv := NewEnv(env)
	var last Value = NullVal()
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(BreakSignal); !ok {
					panic(r)
				}
			}
		}()
		for cur := from; (by > 0 && cur <= to) || (by < 0 && cur >= to); cur += by {
			loopEnv.Define(e.IName, NumVal(cur))
			last = i.execBody(e.Body, loopEnv)
		}
	}()
	return last
}

func (i *Interpreter) evalWhile(e *control.While, env *Env) Value {
	var last Value = NullVal()
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(BreakSignal); !ok {
					panic(r)
				}
			}
		}()
		for i.evalInEnv(e.Condition, env).AsBool() {
			last = i.execBody(e.Body, env)
		}
	}()
	return last
}

func (i *Interpreter) evalEach(e *control.Each, env *Env) Value {
	list := i.evalInEnv(e.Iterable, env).AsList()
	loopEnv := NewEnv(env)
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
		for _, elem := range *list {
			loopEnv.Define(e.IName, elem)
			last = i.execBody(e.Body, loopEnv)
		}
	}()
	return last
}

func (i *Interpreter) evalEachPair(e *control.EachPair, env *Env) Value {
	d := i.evalInEnv(e.Iterable, env).AsDict()
	loopEnv := NewEnv(env)
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
		for _, entry := range d.entries {
			loopEnv.Define(e.KeyName, StrVal(entry.Key))
			loopEnv.Define(e.ValueName, entry.Val)
			last = i.execBody(e.Body, loopEnv)
		}
	}()
	return last
}

// --- Variables ---

func (i *Interpreter) evalVar(e *variables.Var, env *Env) Value {
	childEnv := NewEnv(env)
	for k, name := range e.Names {
		val := i.evalInEnv(e.Values[k], env) // values evaluated in parent scope
		childEnv.Define(name, val)
	}
	return i.execBody(e.Body, childEnv)
}

func (i *Interpreter) evalSimpleVar(e *variables.SimpleVar, env *Env) Value {
	val := i.evalInEnv(e.Value, env)
	childEnv := NewEnv(env)
	childEnv.Define(e.Name, val)
	return i.execBody(e.Body, childEnv)
}

func (i *Interpreter) evalVarResult(e *variables.VarResult, env *Env) Value {
	childEnv := NewEnv(env)
	for k, name := range e.Names {
		val := i.evalInEnv(e.Values[k], env)
		childEnv.Define(name, val)
	}
	return i.evalInEnv(e.Result, childEnv)
}

// --- Procedure calls ---

func (i *Interpreter) evalProcedureCall(e *procedures.Call, env *Env) Value {
	proc, ok := i.procs[e.Name]
	if !ok {
		panic("undefined procedure: " + e.Name)
	}
	callEnv := NewEnv(i.globalEnv)
	for k, param := range proc.params {
		callEnv.Define(param, i.evalInEnv(e.Arguments[k], env))
	}

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
			result = i.evalInEnv(proc.retExpr, callEnv)
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
		i.execBody(proc.voidBody, callEnv)
	}()
	return ret
}

// --- execBody: run a slice of expressions, return value of last ---

func (i *Interpreter) execBody(body []ast.Expr, env *Env) Value {
	if len(body) == 0 {
		return NullVal()
	}
	var last Value
	for _, expr := range body {
		last = i.evalInEnv(expr, env)
	}
	return last
}
