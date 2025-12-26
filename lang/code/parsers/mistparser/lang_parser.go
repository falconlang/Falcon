package mistparser

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
	"Falcon/code/ast/components"
	"Falcon/code/ast/control"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/list"
	"Falcon/code/ast/method"
	"Falcon/code/ast/procedures"
	"Falcon/code/ast/variables"
	"Falcon/code/sugar"
	"strconv"
	"strings"

	l "Falcon/code/lex"
)

type LangParser struct {
	Tokens         []*l.Token
	currIndex      int
	tokenSize      int
	currCheckpoint int

	strict bool

	Resolver    *NameResolver
	ScopeCursor *ScopeCursor
	aggregator  *ErrorAggregator
}

func NewLangParser(strict bool, tokens []*l.Token) *LangParser {
	return &LangParser{
		Tokens:         tokens,
		tokenSize:      len(tokens),
		currIndex:      0,
		currCheckpoint: 0,
		strict:         strict,
		Resolver: &NameResolver{
			Procedures:        map[string]*Procedure{},
			ComponentTypesMap: map[string]string{},
			ComponentNameMap:  map[string][]string{},
		},
		ScopeCursor: MakeScopeCursor(),
		aggregator:  &ErrorAggregator{Errors: map[*l.Token]ParseError{}},
	}
}

func (p *LangParser) SetComponentDefinitions(definitions map[string][]string, reverseDefinitions map[string]string) {
	p.Resolver.ComponentNameMap = definitions
	p.Resolver.ComponentTypesMap = reverseDefinitions
}

func (p *LangParser) GetComponentDefinitionsCode() string {
	// convert the AST back to syntax
	var definitions strings.Builder
	for key, value := range p.Resolver.ComponentNameMap {
		definitions.WriteString(sugar.Format("@% { % }\n", key, strings.Join(value, ", ")))
	}
	return definitions.String()
}

func (p *LangParser) ParseAll() []ast.Expr {
	var expressions []ast.Expr
	if p.notEOF() {
		p.defineStatements()
	}
	for p.notEOF() {
		e := p.parse()
		expressions = append(expressions, e)
	}
	if p.strict {
		p.checkPendingSymbols()
	}
	return expressions
}

func (p *LangParser) checkPendingSymbols() {
	var errorMessages []string
	for token, parseError := range p.aggregator.Errors {
		// try resolve global variables again
		if get, ok := parseError.Owner.(*variables.Get); ok && get.Global {
			signatures, resolved := p.ScopeCursor.ResolveVariable(get.Name)
			println("Failed to resolve: " + get.String())
			if resolved {
				get.ValueSignature = signatures
				continue
			}
		} else if procCall, ok := parseError.Owner.(*procedures.Call); ok {
			// a late resolution of procedure calls
			procedureErrorMessage, procedureSignature := p.Resolver.ResolveProcedure(procCall.Name, len(procCall.Arguments))
			if procedureSignature != nil {
				procCall.Parameters = procedureSignature.Parameters
				procCall.Returning = procedureSignature.Returning
				continue
			}
			parseError.ErrorMessage = procedureErrorMessage
		}
		errorMessages = append(errorMessages, token.BuildError(false, parseError.ErrorMessage))
	}
	if len(errorMessages) > 0 {
		var errorWriter strings.Builder
		errorWriter.WriteString(sugar.Format("compile failed with % syntax errors", strconv.Itoa(len(errorMessages))))
		errorWriter.WriteString(strings.Join(errorMessages, ""))
		panic(errorWriter.String())
	}
}

func (p *LangParser) defineStatements() {
	for p.notEOF() && p.consume(l.At) {
		compType := p.name()
		p.expect(l.OpenCurly)
		if !p.consume(l.CloseCurly) {
			var componentNames []string
			for {
				name := p.name()
				componentNames = append(componentNames, name)
				p.Resolver.ComponentTypesMap[name] = compType
				if !p.consume(l.Comma) {
					break
				}
			}
			p.Resolver.ComponentNameMap[compType] = componentNames
			p.expect(l.CloseCurly)
		}
	}
}

func (p *LangParser) parse() ast.Expr {
	switch p.peek().Type {
	case l.If:
		return p.ifSmt()
	case l.For:
		return p.forExpr()
	case l.While:
		return p.whileExpr()
	case l.Break:
		p.skip()
		return &control.Break{}
	case l.Local:
		return p.varExpr()
	case l.Global:
		return p.globVar()
	case l.Func:
		return p.funcSmt()
	case l.When:
		p.skip()
		if p.consume(l.Any) {
			return p.genericEvent()
		}
		return p.event()
	default:
		// It cannot be consumable
		return p.expr(0)
	}
}

func (p *LangParser) genericEvent() ast.Expr {
	componentType := p.componentType()
	p.expect(l.Dot)
	eventName := p.name()
	var parameters []string
	if p.isNext(l.OpenCurve) {
		parameters = p.parameters()
	}
	where := p.expect(l.OpenCurly)
	p.ScopeCursor.Enter(where, ScopeEvent)
	for _, param := range parameters {
		p.ScopeCursor.DefineVariable(param, []ast.Signature{ast.SignOfEvent, ast.SignAny})
	}
	body := p.bodyUntilCurly()
	p.ScopeCursor.Exit(ScopeEvent)
	p.expect(l.CloseCurly)
	return &components.GenericEvent{ComponentType: componentType, Event: eventName, Parameters: parameters, Body: body}
}

func (p *LangParser) event() ast.Expr {
	component := p.component()
	p.expect(l.Dot)
	eventName := p.name()
	var parameters []string
	if p.isNext(l.OpenCurve) {
		parameters = p.parameters()
	}
	where := p.expect(l.OpenCurly)
	p.ScopeCursor.Enter(where, ScopeEvent)
	for _, param := range parameters {
		p.ScopeCursor.DefineVariable(param, []ast.Signature{ast.SignOfEvent, ast.SignAny})
	}
	body := p.bodyUntilCurly()
	p.ScopeCursor.Exit(ScopeEvent)
	p.expect(l.CloseCurly)

	return &components.Event{
		ComponentName: component.Name,
		ComponentType: component.Type,
		Event:         eventName,
		Parameters:    parameters,
		Body:          body,
	}
}

func (p *LangParser) funcSmt() ast.Expr {
	where := p.next()
	name := p.name()
	var parameters = p.parameters()
	returning := p.consume(l.Assign)
	p.Resolver.Procedures[name] = &Procedure{Name: name, Parameters: parameters, Returning: returning}
	if returning {
		p.ScopeCursor.Enter(where, ScopeSmartBody)
		for _, parameter := range parameters {
			p.ScopeCursor.DefineVariable(parameter, []ast.Signature{ast.SignAny})
		}
		var result ast.Expr
		if p.isNext(l.OpenCurly) {
			result = p.smartBody()
		} else {
			result = p.parse()
		}
		p.ScopeCursor.Exit(ScopeSmartBody)
		return &procedures.RetProcedure{Name: name, Parameters: parameters, Result: result}
	} else {
		where := p.expect(l.OpenCurly)
		p.ScopeCursor.Enter(where, ScopeProc)
		for _, parameter := range parameters {
			p.ScopeCursor.DefineVariable(parameter, []ast.Signature{ast.SignAny})
		}
		body := p.bodyUntilCurly()
		p.ScopeCursor.Exit(ScopeProc)
		p.expect(l.CloseCurly)
		return &procedures.VoidProcedure{Name: name, Parameters: parameters, Body: body}
	}
}

func (p *LangParser) globVar() ast.Expr {
	where := p.next()
	if !p.ScopeCursor.AtRoot() {
		where.Error("Global variables can only be defined at the root.")
	}
	name := p.name()
	p.expect(l.Assign)
	value := p.parse()
	p.ScopeCursor.DefineVariable(name, value.Signature())
	return &variables.Global{Name: name, Value: value}
}

func (p *LangParser) varExpr() ast.Expr {
	// a clean full scope variable
	var names []string
	var values []ast.Expr
	for {
		p.createCheckpoint()
		if !p.consume(l.Local) {
			break
		}
		name := p.name()
		p.expect(l.Assign)
		value := p.parse()

		if ast.DependsOnVariables(value, names) {
			// Since this variable depends on the last variable, we cannot include
			// it in the current set.
			p.backToPast()
			break
		}

		names = append(names, name)
		values = append(values, value)
		p.ScopeCursor.DefineVariable(name, value.Signature())
	}
	// we have to parse rest of the body here
	return &variables.Var{Names: names, Values: values, Body: p.bodyUntilCurly()}
}

func (p *LangParser) whileExpr() *control.While {
	p.skip()
	p.expect(l.OpenCurve)
	condition := p.parse()
	p.expect(l.CloseCurve)
	body := p.body(ScopeLoop)
	return &control.While{Condition: condition, Body: body}
}

func (p *LangParser) forExpr() ast.Expr {
	// TODO:
	//  We could refactor this later to reuse declaring variables inside body
	p.skip()
	p.expect(l.OpenCurve)
	firstName := p.name()
	if p.consume(l.Comma) {
		// Dictionary For each loop
		valueName := p.name()
		p.expect(l.In)
		iterable := p.parse()
		p.expect(l.CloseCurve)

		where := p.expect(l.OpenCurly)
		p.ScopeCursor.Enter(where, ScopeLoop)
		p.ScopeCursor.DefineVariable(firstName, iterable.Signature())
		p.ScopeCursor.DefineVariable(valueName, iterable.Signature())
		body := p.bodyUntilCurly()
		p.ScopeCursor.Exit(ScopeLoop)
		p.expect(l.CloseCurly)

		return &control.EachPair{KeyName: firstName, ValueName: valueName, Iterable: iterable, Body: body}
	} else if p.consume(l.In) {
		// For each loop
		iterable := p.parse()
		p.expect(l.CloseCurve)

		where := p.expect(l.OpenCurly)
		p.ScopeCursor.Enter(where, ScopeLoop)
		p.ScopeCursor.DefineVariable(firstName, iterable.Signature())
		body := p.bodyUntilCurly()
		p.ScopeCursor.Exit(ScopeLoop)
		p.expect(l.CloseCurly)

		return &control.Each{IName: firstName, Iterable: iterable, Body: body}
	}
	// For I loop
	p.expect(l.Colon)
	// Earlier we were using p.element(), check the side effects
	from := p.parse()
	p.expect(l.DoubleDot)
	to := p.parse()
	var by ast.Expr
	if p.consume(l.Step) {
		by = p.parse()
	} else {
		by = &fundamentals.Number{Content: "1"}
	}
	p.expect(l.CloseCurve)

	where := p.expect(l.OpenCurly)
	p.ScopeCursor.Enter(where, ScopeLoop)
	p.ScopeCursor.DefineVariable(firstName, []ast.Signature{ast.SignNumb})
	body := p.bodyUntilCurly()
	p.ScopeCursor.Exit(ScopeLoop)
	p.expect(l.CloseCurly)

	return &control.For{IName: firstName, From: from, To: to, By: by, Body: body}
}

func (p *LangParser) ifSmt() ast.Expr {
	p.skip()
	var conditions []ast.Expr
	var bodies [][]ast.Expr

	conditions = append(conditions, p.expr(0))
	if p.isNext(l.OpenCurly) {
		bodies = append(bodies, p.body(ScopeIfBody))
	} else {
		bodies = append(bodies, []ast.Expr{p.parse()})
	}

	var elseBody []ast.Expr
	for p.notEOF() && p.consume(l.Else) {
		if p.consume(l.If) {
			conditions = append(conditions, p.expr(0))
			if p.isNext(l.OpenCurly) {
				bodies = append(bodies, p.body(ScopeIfBody))
			} else {
				bodies = append(bodies, []ast.Expr{p.parse()})
			}
		} else {
			if p.isNext(l.OpenCurly) {
				elseBody = p.body(ScopeIfBody)
			} else {
				elseBody = []ast.Expr{p.parse()}
			}
			break
		}
	}
	return &control.If{Conditions: conditions, Bodies: bodies, ElseBody: elseBody}
}

func (p *LangParser) body(scope ScopeType) []ast.Expr {
	where := p.expect(l.OpenCurly)
	p.ScopeCursor.Enter(where, scope)
	expressions := p.bodyUntilCurly()
	p.ScopeCursor.Exit(scope)
	p.expect(l.CloseCurly)
	return expressions
}

func (p *LangParser) bodyUntilCurly() []ast.Expr {
	var expressions []ast.Expr
	if p.isNext(l.CloseCurly) {
		return expressions
	}
	for p.notEOF() && !p.isNext(l.CloseCurly) {
		expressions = append(expressions, p.parse())
		p.consume(l.Comma)
	}
	return expressions
}

func (p *LangParser) expr(minPrecedence int) ast.Expr {
	left := p.element()
	for p.notEOF() {
		opToken := p.peek()
		if !opToken.HasFlag(l.Operator) {
			break
		}
		precedence := l.PrecedenceOf(opToken.Flags[0])
		if precedence == -1 || precedence < minPrecedence {
			break
		}
		p.skip()
		if p.isNext(l.Assign) && opToken.HasFlag(l.Compoundable) {
			// a compound operator e.g. a += 3
			p.skip()
			left = p.compoundOperator(opToken, left)
			break
		}
		var right ast.Expr
		if opToken.HasFlag(l.PreserveOrder) {
			right = p.element()
		} else {
			right = p.expr(precedence)
		}
		if rBinExpr, ok := right.(*common.BinaryExpr); ok && rBinExpr.CanRepeat(opToken.Type) {
			// for NoPreserveOrder: merge binary expr with same operator (towards right)
			rBinExpr.Operands = append([]ast.Expr{left}, rBinExpr.Operands...)
			left = rBinExpr
		} else if lBinExpr, ok := left.(*common.BinaryExpr); ok && lBinExpr.CanRepeat(opToken.Type) {
			// for PreserveOder: merge binary expr with same operator (towards left)
			lBinExpr.Operands = append(lBinExpr.Operands, right)
		} else {
			// a new binary node
			left = p.makeBinary(opToken, left, right)
		}
	}
	return left
}

func (p *LangParser) compoundOperator(opToken *l.Token, left ast.Expr) ast.Expr {
	right := p.parse()
	var binaryOperator ast.Expr
	if opToken.Type == l.Remainder {
		binaryOperator = common.MakeFuncCall("rem", left, right)
	} else {
		binaryOperator = p.makeBinary(opToken, left, right)
	}
	expr, done := p.assignSmt(left, binaryOperator)
	if done {
		return expr
	} else {
		opToken.Error("Unknown compound operator '%='", *opToken.Content)
		panic("unreached")
	}
}

func (p *LangParser) makeBinary(opToken *l.Token, left ast.Expr, right ast.Expr) ast.Expr {
	switch opToken.Type {
	case l.Colon:
		return &fundamentals.Pair{Key: left, Value: right}
	case l.Assign:
		expr, done := p.assignSmt(left, right)
		if done {
			return expr
		}
	case l.Remainder:
		return common.MakeFuncCall("rem", left, right)
	}
	return &common.BinaryExpr{Where: opToken, Operands: []ast.Expr{left, right}, Operator: opToken.Type}
}

func (p *LangParser) assignSmt(left ast.Expr, right ast.Expr) (ast.Expr, bool) {
	if nameExpr, ok := left.(*variables.Get); ok {
		p.aggregator.MarkResolved(nameExpr.Where)
		return &variables.Set{Global: nameExpr.Global, Name: nameExpr.Name, Expr: right}, true
	} else if listGet, ok := left.(*list.Get); ok {
		return &list.Set{List: listGet.List, Index: listGet.Index, Value: right}, true
	}
	return nil, false
}

func (p *LangParser) element() ast.Expr {
	left := p.term()
	for p.notEOF() {
		pe := p.peek()
		// check if it's a variable Get, if so, check if it refers to a component
		if getExpr, ok := left.(*fundamentals.Component); ok && pe.Type == l.Dot {
			if compType, exists := p.Resolver.ComponentTypesMap[getExpr.Name]; exists {
				// a specific component call (MethodCall, PropertyGet, PropertySet)
				left = p.componentCall(getExpr.Name, compType)
				continue
			}
		}

		switch pe.Type {
		case l.At:
			left = p.helperDropdown(left)
		case l.Dot:
			left = p.objectCall(left)
			continue
		case l.Question:
			left = &common.Question{Where: p.next(), On: left, Question: p.name()}
			continue
		case l.DoubleColon:
			// constant value transformer
			left = &common.Transform{Where: p.next(), On: left, Name: p.name()}
		case l.OpenSquare:
			p.skip()
			// an index element access
			left = &list.Get{List: left, Index: p.parse()}
			p.expect(l.CloseSquare)
			continue
		}
		break
	}
	return left
}

func (p *LangParser) componentCall(compName string, compType string) ast.Expr {
	p.expect(l.Dot)
	resource := p.name()
	if p.isNext(l.OpenCurve) {
		return &components.MethodCall{
			ComponentName: compName,
			ComponentType: compType,
			Method:        resource,
			Args:          p.arguments(),
		}
	} else if p.consume(l.Assign) {
		assignment := p.expr(0)
		return &components.PropertySet{
			ComponentName: compName,
			ComponentType: compType,
			Property:      resource,
			Value:         assignment,
		}
	}
	return &components.PropertyGet{ComponentName: compName, ComponentType: compType, Property: resource}
}

func (p *LangParser) helperDropdown(keyExpr ast.Expr) ast.Expr {
	where := p.next()
	if key, ok := keyExpr.(*variables.Get); ok && !key.Global {
		p.aggregator.MarkResolved(key.Where)
		return &fundamentals.HelperDropdown{Key: key.Name, Option: p.name()}
	}
	where.Error("Invalid Helper Access operation ")
	panic("")
}

func (p *LangParser) objectCall(object ast.Expr) ast.Expr {
	p.skip()
	where := p.next()
	name := *where.Content

	var args []ast.Expr
	if p.isNext(l.OpenCurve) {
		args = p.arguments()
		if !p.isNext(l.OpenCurly) {
			// he's a simple call!
			errorMessage, signature := method.TestSignature(name, len(args))
			if signature == nil {
				p.aggregator.EnqueueSymbol(where, object, errorMessage)
			} else {
				p.aggregator.MarkResolved(where)
			}
			return &method.Call{Where: where, On: object, Name: name, Args: args}
		}
	}
	p.expect(l.OpenCurly)
	// oh, no! he's a transformer >_>
	p.ScopeCursor.Enter(where, ScopeTypeTransform)
	var namesUsed []string
	if !p.consume(l.RightArrow) {
		for {
			mName := p.name()
			p.ScopeCursor.DefineVariable(mName, []ast.Signature{ast.SignAny})
			namesUsed = append(namesUsed, mName)
			if !p.consume(l.Comma) {
				break
			}
		}
		p.consume(l.RightArrow)
	}
	transformer := p.parse()
	p.ScopeCursor.Exit(ScopeTypeTransform)
	p.consume(l.CloseCurly)
	errorMessage, signature := list.TestSignature(name, len(args), len(namesUsed))
	if signature == nil {
		p.aggregator.EnqueueSymbol(where, object, errorMessage)
	} else {
		p.aggregator.MarkResolved(where)
	}
	return &list.Transformer{
		Where:       where,
		List:        object,
		Name:        name,
		Args:        args,
		Names:       namesUsed,
		Transformer: transformer}
}

func (p *LangParser) term() ast.Expr {
	token := p.next()
	switch token.Type {
	case l.Undefined:
		return &common.EmptySocket{}
	case l.OpenSquare:
		return p.list()
	case l.OpenCurly:
		p.back()
		return p.smartBody()
	case l.OpenCurve:
		e := p.parse()
		p.expect(l.CloseCurve)
		return e
	case l.Not:
		return &fundamentals.Not{Expr: p.element()}
	case l.Dash:
		return common.MakeFuncCall("neg", p.element())
	case l.If:
		p.back()
		return p.ifSmt()
	case l.Compute:
		return p.computeExpr()
	case l.WalkAll:
		return &fundamentals.WalkAll{}
	default:
		if token.HasFlag(l.Value) {
			return p.checkCall(token)
		}
		token.Error("Unexpected! %", token.String())
		panic("") // unreachable
	}
}

func (p *LangParser) smartBody() ast.Expr {
	body := p.body(ScopeSmartBody)
	k := 0
	for ; k < len(body); k++ {
		if _, ok := body[k].(*fundamentals.Pair); !ok {
			break
		}
	}
	if k == len(body) {
		// It's actually a dictionary!
		return &fundamentals.Dictionary{Elements: body}
	}
	return &fundamentals.SmartBody{Body: body}
}

func (p *LangParser) checkCall(token *l.Token) ast.Expr {
	value := p.value(token)
	if nameExpr, ok := value.(*variables.Get); ok && !nameExpr.Global && p.isNext(l.OpenCurve) {
		arguments := p.arguments()
		// check for in-built function call
		_, funcCallSignature := common.TestSignature(nameExpr.Name, len(arguments))
		if funcCallSignature != nil {
			p.aggregator.MarkResolved(nameExpr.Where)
			return &common.FuncCall{Where: nameExpr.Where, Name: nameExpr.Name, Args: arguments}
		}
		// check for a user defined procedure
		procedureErrorMessage, procedureSignature := p.Resolver.ResolveProcedure(nameExpr.Name, len(arguments))
		var funcCall *procedures.Call
		if procedureSignature != nil {
			funcCall = &procedures.Call{
				Name:       nameExpr.Name,
				Parameters: procedureSignature.Parameters,
				Arguments:  arguments,
				Returning:  procedureSignature.Returning,
			}
			p.aggregator.MarkResolved(nameExpr.Where)
		} else {
			// just fill in a template, could be resolved later
			funcCall = &procedures.Call{Name: nameExpr.Name, Arguments: arguments}
			p.aggregator.EnqueueSymbol(nameExpr.Where, funcCall, procedureErrorMessage)
		}
		return funcCall
	}
	return value
}

func (p *LangParser) computeExpr() *variables.VarResult {
	var varNames []string
	var varValues []ast.Expr
	p.expect(l.OpenCurve)

	// a result local var
	for p.notEOF() && !p.isNext(l.CloseCurve) {
		name := p.name()
		p.expect(l.Assign)
		value := p.parse()

		varNames = append(varNames, name)
		varValues = append(varValues, value)

		if !p.consume(l.Comma) {
			break
		}
	}
	p.expect(l.CloseCurve)
	p.expect(l.RightArrow)
	return &variables.VarResult{Names: varNames, Values: varValues, Result: p.parse()}
}

func (p *LangParser) dictionary() *fundamentals.Dictionary {
	var elements []ast.Expr
	if !p.consume(l.CloseCurly) {
		for p.notEOF() {
			elements = append(elements, p.expr(0))
			if !p.consume(l.Comma) {
				break
			}
		}
		p.expect(l.CloseCurly)
	}
	return &fundamentals.Dictionary{Elements: elements}
}

func (p *LangParser) list() *fundamentals.List {
	var elements []ast.Expr
	if !p.consume(l.CloseSquare) {
		for p.notEOF() {
			elements = append(elements, p.expr(0))
			if !p.consume(l.Comma) {
				break
			}
		}
		p.expect(l.CloseSquare)
	}
	return &fundamentals.List{Elements: elements}
}

func (p *LangParser) parameters() []string {
	p.expect(l.OpenCurve)
	var parameters []string
	if !p.consume(l.CloseCurve) {
		for p.notEOF() && !p.isNext(l.CloseCurve) {
			parameters = append(parameters, p.name())
			if !p.consume(l.Comma) {
				break
			}
		}
		p.expect(l.CloseCurve)
	}
	return parameters
}

func (p *LangParser) arguments() []ast.Expr {
	p.expect(l.OpenCurve)
	var args []ast.Expr
	if p.consume(l.CloseCurve) {
		return args
	}
	for p.notEOF() {
		args = append(args, p.expr(0))
		if !p.consume(l.Comma) {
			break
		}
	}
	p.expect(l.CloseCurve)
	return args
}

func (p *LangParser) value(t *l.Token) ast.Expr {
	switch t.Type {
	case l.True, l.False:
		return &fundamentals.Boolean{Value: t.Type == l.True}
	case l.Number:
		return &fundamentals.Number{Content: *t.Content}
	case l.Text:
		return &fundamentals.Text{Content: *t.Content}
	case l.Name:
		if compType, exists := p.Resolver.ComponentTypesMap[*t.Content]; exists {
			return &fundamentals.Component{Name: *t.Content, Type: compType}
		}
		// May not be variable reference always. It could be a func or a method call.
		signatures, found := p.ScopeCursor.ResolveVariable(*t.Content)
		get := &variables.Get{Where: t, Global: false, Name: *t.Content, ValueSignature: signatures}
		if !found {
			p.aggregator.EnqueueSymbol(t, get, "Cannot find symbol '"+*t.Content+"'")
		}
		return get
	case l.This:
		p.expect(l.Dot)
		nameToken := p.expect(l.Name)
		name := *nameToken.Content
		signatures, found := p.ScopeCursor.ResolveVariable(name)
		get := &variables.Get{Where: t, Global: true, Name: name, ValueSignature: signatures}
		if !found {
			p.aggregator.EnqueueSymbol(nameToken, get, "Cannot find symbol '"+*nameToken.Content+"'")
		}
		return get
	case l.ColorCode:
		return &fundamentals.Color{Where: t, Hex: *t.Content}
	default:
		t.Error("Unknown value type '%'", t.String())
		panic("") // unreachable
	}
}

func (p *LangParser) componentType() string {
	token := p.expect(l.Name)
	name := *token.Content
	if _, exists := p.Resolver.ComponentNameMap[name]; exists {
		return name
	}
	token.Error("Undefined component group %", name)
	panic("")
}

func (p *LangParser) component() fundamentals.Component {
	token := p.expect(l.Name)
	name := *token.Content
	if compType, exists := p.Resolver.ComponentTypesMap[name]; exists {
		return fundamentals.Component{Name: name, Type: compType}
	}
	token.Error("Undefined component %", name)
	panic("")
}

func (p *LangParser) name() string {
	return *p.expect(l.Name).Content
}

func (p *LangParser) consume(t l.Type) bool {
	if p.notEOF() && p.peek().Type == t {
		p.currIndex++
		return true
	}
	return false
}

func (p *LangParser) expect(t l.Type) *l.Token {
	if p.isEOF() {
		panic("Early EOF! Was expecting type " + t.String())
	}
	got := p.next()
	if got.Type != t {
		got.Error("Expected type % but got %", t.String(), got.String())
	}
	return got
}

func (p *LangParser) isNext(checkTypes ...l.Type) bool {
	if p.isEOF() {
		return false
	}
	pType := p.peek().Type
	for _, checkType := range checkTypes {
		if checkType == pType {
			return true
		}
	}
	return false
}

func (p *LangParser) peek() *l.Token {
	if p.isEOF() {
		panic("Early EOF!")
	}
	return p.Tokens[p.currIndex]
}

func (p *LangParser) next() *l.Token {
	if p.isEOF() {
		panic("Early EOF!")
	}
	token := p.Tokens[p.currIndex]
	p.currIndex++
	return token
}

func (p *LangParser) createCheckpoint() {
	p.currCheckpoint = p.currIndex
}

func (p *LangParser) backToPast() {
	p.currIndex = p.currCheckpoint
}

func (p *LangParser) back() {
	p.currIndex--
}

func (p *LangParser) skip() {
	p.currIndex++
}

func (p *LangParser) notEOF() bool {
	return p.currIndex < p.tokenSize
}

func (p *LangParser) isEOF() bool {
	return p.currIndex >= p.tokenSize
}
