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

type stackFrame struct {
	token *lex.Token
	name  string
}

// Interpreter implements Visitor and holds the runtime state.
type Interpreter struct {
	globalEnv      *Env
	currEnv        *Env
	procedures     map[string]*Procedure
	lastToken      *lex.Token   // last source token seen during Eval — used for runtime error reporting
	lastHighlight  int          // override highlight width (0 = use token's own content length)
	stackTrace     []stackFrame // populated as panics propagate up through procedure calls
	outputCallback func(string) // if non-nil, receives each printed line instead of stdout
}

func NewInterpreter() *Interpreter {
	env := NewEnv(nil)
	return &Interpreter{
		globalEnv:  env,
		currEnv:    env,
		procedures: make(map[string]*Procedure),
	}
}

// NewInterpreterWithOutput creates an Interpreter that calls callback for every
// printed line instead of writing to stdout. Used by the WASM runtime to stream
// output back to JavaScript.
func NewInterpreterWithOutput(callback func(string)) *Interpreter {
	env := NewEnv(nil)
	return &Interpreter{
		globalEnv:      env,
		currEnv:        env,
		procedures:     make(map[string]*Procedure),
		outputCallback: callback,
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
	// Register all procedures first so globals can call them.
	for _, e := range exprs {
		switch n := e.(type) {
		case *procedures.VoidProcedure:
			i.procedures[n.Name] = &Procedure{params: n.Parameters, voidBody: n.Body}
		case *procedures.RetProcedure:
			i.procedures[n.Name] = &Procedure{params: n.Parameters, retExpr: n.Result}
		}
	}
	for _, e := range exprs {
		if n, ok := e.(*variables.Global); ok {
			val := i.Eval(n.Value)
			i.globalEnv.Define(n.Name, val)
		}
	}
	var last = NullVal()
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
	// Register all procedures first so global initializers can call them.
	for _, e := range exprs {
		switch n := e.(type) {
		case *procedures.VoidProcedure:
			i.procedures[n.Name] = &Procedure{params: n.Parameters, voidBody: n.Body}
		case *procedures.RetProcedure:
			i.procedures[n.Name] = &Procedure{params: n.Parameters, retExpr: n.Result}
		}
	}
	// Evaluate globals after all procedures are registered.
	for _, e := range exprs {
		if n, ok := e.(*variables.Global); ok {
			val := i.Eval(n.Value)
			i.globalEnv.Define(n.Name, val)
		}
	}
	// Execute non-definition top-level statements.
	for _, e := range exprs {
		switch e.(type) {
		case *procedures.VoidProcedure, *procedures.RetProcedure, *variables.Global:
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
	case *fundamentals.Yield:
		return i.Eval(e.GetExpr())
	case *fundamentals.Component:
		i.stub("component reference @" + e.Name + " (" + e.Type + ")")
		return NullVal()

	// falcon specific features
	case *common.EmptySocket:
		return NumVal(0)
	case *common.Transform:
		return i.Eval(e.On) // for "obfuscate" unwrapping

	case *common.BinaryExpr:
		i.lastToken = e.Where
		return i.binary(e)
	case *common.FuncCall:
		i.lastToken = e.Where
		return i.evalFuncCall(e)
	case *common.Question:
		i.lastToken = e.Where
		return i.question(e)
	case *astmethod.Call:
		i.lastToken = e.Where
		return i.methodCall(e)

	// Control blocks
	case *control.If:
		return i.ifSmt(e)
	case *control.SimpleIf:
		return i.simpleIfExpr(e)
	case *control.For:
		i.lastToken = e.Where
		return i.forSmt(e)
	case *control.While:
		i.lastToken = e.Where
		return i.whileSmt(e)
	case *control.Each:
		i.lastToken = e.Where
		return i.eachSmt(e)
	case *control.EachPair:
		i.lastToken = e.Where
		return i.eachPairSmt(e)
	case *control.Break:
		panic(BreakSignal{})
	case *control.Yield:
		panic(YieldSignal{Val: i.Eval(e.Expr)})
	case *control.Do:
		i.execBody(e.Body)
		return i.Eval(e.Result)

	// Variable blocks
	case *variables.Get:
		i.lastToken = e.Where
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
		return VoidVal()
	case *variables.Global:
		val := i.Eval(e.Value)
		i.currEnv.root().Define(e.Name, val)
		return VoidVal()
	case *variables.Var:
		return i.evalVar(e)
	case *variables.SimpleVar:
		return i.evalSimpleVar(e)
	case *variables.VarResult:
		return i.evalVarResult(e)

	// Procedure definitions
	case *procedures.VoidProcedure:
		i.procedures[e.Name] = &Procedure{params: e.Parameters, voidBody: e.Body}
		return VoidVal()
	case *procedures.RetProcedure:
		i.procedures[e.Name] = &Procedure{params: e.Parameters, retExpr: e.Result}
		return VoidVal()
	case *procedures.Call:
		i.lastToken = e.Where
		return i.evalProcedureCall(e)

	// List manipulation
	case *astlist.Get:
		listVal := i.Eval(e.List)
		i.lastToken = e.Where
		hl := 1 + len(e.Index.String()) + 1 // covers full [index]
		i.lastHighlight = hl
		if listVal.Type() == String {
			panic("expected a list value but got " + listVal.errorStr() + " — use .segment(start, length) to extract characters from text")
		}
		list := listVal.AsList()
		i.lastHighlight = 0
		idx := int(i.Eval(e.Index).AsNum())
		if idx < 1 || idx > len(*list) {
			i.lastToken = e.Where
			i.lastHighlight = hl
			panic("list index " + strconv.Itoa(idx) + " out of bounds (len=" + strconv.Itoa(len(*list)) + ")")
		}
		return (*list)[idx-1]
	case *astlist.Set:
		listVal := i.Eval(e.List)
		i.lastToken = e.Where
		hl := 1 + len(e.Index.String()) + 1 // covers full [index]
		i.lastHighlight = hl
		list := listVal.AsList()
		i.lastHighlight = 0
		idx := int(i.Eval(e.Index).AsNum())
		val := i.Eval(e.Value)
		if idx < 1 || idx > len(*list) {
			i.lastToken = e.Where
			i.lastHighlight = hl
			panic("list index " + strconv.Itoa(idx) + " out of bounds (len=" + strconv.Itoa(len(*list)) + ")")
		}
		(*list)[idx-1] = val
		return VoidVal()
	case *astlist.Transformer:
		i.lastToken = e.Where
		return i.evalTransformer(e)

	// Component blocks
	case *components.Event:
		i.stub("event handler " + e.ComponentName + "." + e.Event)
		return VoidVal()
	case *components.GenericEvent:
		i.stub("generic event handler " + e.ComponentType + "." + e.Event)
		return VoidVal()
	case *components.MethodCall:
		i.stub("component method " + e.ComponentName + "." + e.Method + "(...)")
		return VoidVal()
	case *components.GenericMethodCall:
		i.stub("generic component method " + e.ComponentType + "." + e.Method + "(...)")
		return VoidVal()
	case *components.PropertyGet:
		i.stub("property get " + e.ComponentName + "." + e.Property)
		return NullVal() // property reads are expressions
	case *components.GenericPropertyGet:
		i.stub("generic property get " + e.ComponentType + "." + e.Property)
		return NullVal() // property reads are expressions
	case *components.PropertySet:
		i.stub("property set " + e.ComponentName + "." + e.Property)
		return VoidVal()
	case *components.GenericPropertySet:
		i.stub("generic property set " + e.ComponentType + "." + e.Property)
		return VoidVal()
	case *components.EveryComponent:
		i.stub("every(" + e.Type + ")")
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
		n := v.AsNum()
		if n != math.Trunc(n) {
			panic("? even requires an integer value, got " + formatNum(n))
		}
		return BoolVal(int64(n)%2 == 0)
	case "odd":
		n := v.AsNum()
		if n != math.Trunc(n) {
			panic("? odd requires an integer value, got " + formatNum(n))
		}
		return BoolVal(int64(n)%2 != 0)
	default:
		panic("unknown ? question: " + e.Question)
	}
}

func (i *Interpreter) ifSmt(e *control.If) Value {
	for k, cond := range e.Conditions {
		if i.Eval(cond).AsBool() {
			return i.execBody(e.Bodies[k])
		}
	}
	if e.ElseBody != nil {
		return i.execBody(e.ElseBody)
	}
	return VoidVal()
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

func (i *Interpreter) forSmt(e *control.For) Value {
	from := i.Eval(e.From).AsNum()
	to := i.Eval(e.To).AsNum()
	by := i.Eval(e.By).AsNum()
	if by == 0 {
		panic("for loop step cannot be 0")
	}
	loopEnv := NewEnv(i.currEnv)
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
				i.execBody(e.Body)
			}
			return VoidVal()
		})
	}()
	return VoidVal()
}

func (i *Interpreter) whileSmt(e *control.While) Value {
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, ok := r.(BreakSignal); !ok {
					panic(r)
				}
			}
		}()
		for i.Eval(e.Condition).AsBool() {
			i.execBody(e.Body)
		}
	}()
	return VoidVal()
}

func (i *Interpreter) eachSmt(e *control.Each) Value {
	list := i.Eval(e.Iterable).AsList()
	loopEnv := NewEnv(i.currEnv)
	loopEnv.Define(e.IName, NullVal())
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
				i.execBody(e.Body)
			}
			return VoidVal()
		})
	}()
	return VoidVal()
}

func (i *Interpreter) eachPairSmt(e *control.EachPair) Value {
	d := i.Eval(e.Iterable).AsDict()
	loopEnv := NewEnv(i.currEnv)
	loopEnv.Define(e.KeyName, NullVal())
	loopEnv.Define(e.ValueName, NullVal())
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
				i.execBody(e.Body)
			}
			return VoidVal()
		})
	}()
	return VoidVal()
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
	i.currEnv.Define(e.Name, i.Eval(e.Value))
	return VoidVal()
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
	savedToken := i.lastToken
	savedHighlight := i.lastHighlight
	argVals := i.evalExprs(e.Arguments)
	i.lastToken = savedToken
	i.lastHighlight = savedHighlight
	if len(argVals) != len(proc.params) {
		panic("procedure " + e.Name + " expects " + strconv.Itoa(len(proc.params)) + " argument(s) but got " + strconv.Itoa(len(argVals)))
	}
	callEnv := NewEnv(i.globalEnv)
	for k, param := range proc.params {
		callEnv.Define(param, argVals[k])
	}

	var result Value
	func() {
		defer func() {
			if r := recover(); r != nil {
				i.stackTrace = append(i.stackTrace, stackFrame{e.Where, e.Name})
				panic(r)
			}
		}()
		result = i.inEnv(callEnv, func() Value {
			if proc.retExpr != nil {
				// a returning procedure
				var res Value
				func() {
					defer func() {
						if r := recover(); r != nil {
							switch rs := r.(type) {
							case ReturnSignal:
								res = rs.Val
							case YieldSignal:
								res = rs.Val
							default:
								panic(r)
							}
						}
					}()
					res = i.Eval(proc.retExpr)
				}()
				return res
			}
			// a void procedure
			func() {
				defer func() {
					if r := recover(); r != nil {
						if _, ok := r.(ReturnSignal); !ok {
							panic(r)
						}
					}
				}()
				i.execBody(proc.voidBody)
			}()
			return VoidVal()
		})
	}()
	return result
}

// evaluates all expressions and returns the last expr result
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

// FormatRuntimeError formats a recovered panic value as a runtime error message.
// If a source token was recorded, the message includes the source line and a caret.
func (i *Interpreter) FormatRuntimeError(r any) string {
	// If the panic value is already a fully-formatted traceback (e.g. from
	// Token.TypeError / Token.Error), return it unchanged.
	msg := "runtime error"
	switch v := r.(type) {
	case string:
		if strings.HasPrefix(v, "Traceback (most recent call last):") {
			return v
		}
		msg = v
	case error:
		msg = v.Error()
	}

	var sb strings.Builder
	sb.WriteString("Traceback (most recent call last):\n")

	// stackTrace is stored innermost -> outermost; print outermost -> innermost
	for j := len(i.stackTrace) - 1; j >= 0; j-- {
		frame := i.stackTrace[j]
		var funcName string
		if j == len(i.stackTrace)-1 {
			funcName = "<module>"
		} else {
			funcName = i.stackTrace[j+1].name
		}
		sb.WriteString(i.formatTraceFrame(frame.token, funcName))
	}

	// The actual error location
	if i.lastToken != nil && i.lastToken.Column >= 0 {
		var funcName string
		if len(i.stackTrace) > 0 {
			funcName = i.stackTrace[0].name
		} else {
			funcName = "<module>"
		}
		sb.WriteString(i.formatTraceFrame(i.lastToken, funcName, true))
	}

	sb.WriteString("RuntimeError: " + msg + "\n")
	return sb.String()
}
func (i *Interpreter) formatTraceFrame(token *lex.Token, funcName string, isLast ...bool) string {
	last := len(isLast) > 0 && isLast[0]
	fileName := "<unknown>"
	if token.Context != nil {
		fileName = token.Context.FileName
	}
	hlSize := 1
	if token.Content != nil {
		hlSize = len(*token.Content)
	}
	if last && i.lastHighlight > 0 {
		hlSize = i.lastHighlight
	}
	if token.Context != nil {
		return token.Context.FormatTracebackFrame(fileName, token.Column, token.Row, hlSize, funcName)
	}
	return "  File \"" + fileName + "\", line " + strconv.Itoa(token.Column) + ", in " + funcName + "\n"
}

// evaluates all exprs in a given list
func (i *Interpreter) evalExprs(exprs []ast.Expr) []Value {
	vals := make([]Value, len(exprs))
	for k, expr := range exprs {
		vals[k] = i.Eval(expr)
	}
	return vals
}
