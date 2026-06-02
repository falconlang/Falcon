package mistparser

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
	"Falcon/code/ast/components"
	"Falcon/code/ast/control"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/list"
	astmatrix "Falcon/code/ast/matrix"
	"Falcon/code/ast/method"
	"Falcon/code/ast/procedures"
	"Falcon/code/ast/variables"
	codecontext "Falcon/code/context"
	"Falcon/code/sugar"
	"strconv"
	"strings"

	l "Falcon/code/lex"
)

// EventValidator is called during event parsing to validate the event against
// the component database. Returning a non-nil error causes a compile error at
// the event-name token.
type EventValidator func(compType, eventName string, params []string) error

// PropertyValidator is called during component property get/set parsing.
type PropertyValidator func(compType, propName string) error

// MethodValidator is called during component method-call parsing.
type MethodValidator func(compType, methodName string, argsCount int) error

type LangParser struct {
	Tokens    []*l.Token
	currIndex int
	tokenSize int

	strict      bool
	autoCorrect bool

	Resolver          *NameResolver
	ScopeCursor       *ScopeCursor
	aggregator        *ErrorAggregator
	patches           []SourcePatch
	eventValidator    EventValidator
	propertyValidator PropertyValidator
	methodValidator   MethodValidator
}

// EnableAutoCorrect turns on the auto-correction pass. Disabled by default.
func (p *LangParser) EnableAutoCorrect() { p.autoCorrect = true }

// ReconstructedSource returns the original source code with all auto-corrections
// applied in-place.
func (p *LangParser) ReconstructedSource() string {
	if len(p.Tokens) == 0 || p.Tokens[0].Context == nil {
		return ""
	}
	return ApplyPatches(*p.Tokens[0].Context.SourceCode, p.patches)
}

func NewLangParser(strict bool, tokens []*l.Token) *LangParser {
	return &LangParser{
		Tokens:    tokens,
		tokenSize: len(tokens),
		currIndex: 0,
		strict:    strict,
		Resolver: &NameResolver{
			Procedures:        map[string]*Procedure{},
			ComponentTypesMap: map[string]string{},
			ComponentNameMap:  map[string][]string{},
		},
		ScopeCursor: MakeScopeCursor(),
		aggregator:  &ErrorAggregator{Errors: map[*l.Token]ParseError{}},
	}
}

// SetEventValidator installs a validator that is called for every event
// declaration. Use compdb.GlobalDB.ValidateEvent to validate against
// simple_components.json.
func (p *LangParser) SetEventValidator(v EventValidator) { p.eventValidator = v }

func (p *LangParser) SetPropertyValidator(v PropertyValidator) { p.propertyValidator = v }

func (p *LangParser) SetMethodValidator(v MethodValidator) { p.methodValidator = v }

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

func (p *LangParser) ParseDefinitions() {
	if p.notEOF() {
		p.defineStatements()
	}
}

func (p *LangParser) ParseAll() []ast.Expr {
	expressions, _ := p.ParseTopLevel()
	return expressions
}

func (p *LangParser) ParseTopLevel() ([]ast.Expr, []int) {
	var expressions []ast.Expr
	var lineNumbers []int
	if p.notEOF() {
		p.defineStatements()
		p.predeclareTopLevelSymbols()
	}
	for p.notEOF() {
		lineNumbers = append(lineNumbers, p.peek().Column)
		e := p.parse()
		expressions = append(expressions, e)
	}
	for _, e := range expressions {
		p.walkAndCorrect(e)
	}
	p.checkPendingSymbols()
	for _, e := range expressions {
		e.Signature()
	}
	return expressions, lineNumbers
}

func (p *LangParser) predeclareTopLevelSymbols() {
	curlyDepth := 0
	for i := p.currIndex; i < p.tokenSize; i++ {
		tok := p.Tokens[i]
		switch tok.Type {
		case l.OpenCurly:
			curlyDepth++
		case l.CloseCurly:
			if curlyDepth > 0 {
				curlyDepth--
			}
		case l.Func:
			if curlyDepth == 0 {
				p.predeclareProcedureAt(i)
			}
		case l.Global:
			if curlyDepth == 0 {
				p.predeclareGlobalAt(i)
			}
		}
	}
}

func (p *LangParser) predeclareProcedureAt(index int) {
	if index+2 >= p.tokenSize {
		return
	}
	nameTok := p.Tokens[index+1]
	if nameTok.Type != l.Name || nameTok.Content == nil || p.Tokens[index+2].Type != l.OpenCurve {
		return
	}

	var parameters []string
	closeIndex := -1
	for i := index + 3; i < p.tokenSize; i++ {
		tok := p.Tokens[i]
		if tok.Type == l.CloseCurve {
			closeIndex = i
			break
		}
		if tok.Type == l.Name && tok.Content != nil {
			parameters = append(parameters, *tok.Content)
		}
	}
	if closeIndex == -1 {
		return
	}

	returning := closeIndex+1 < p.tokenSize && p.Tokens[closeIndex+1].Type == l.Assign
	name := *nameTok.Content
	p.Resolver.Procedures[name] = &Procedure{Name: name, Parameters: parameters, Returning: returning}
}

func (p *LangParser) isAnonymousProcedureStart() bool {
	return p.currIndex+1 < p.tokenSize && p.Tokens[p.currIndex+1].Type == l.OpenCurve
}

func (p *LangParser) isProcedureDropdownStart() bool {
	return p.currIndex+2 < p.tokenSize &&
		p.Tokens[p.currIndex+1].Type == l.Dot &&
		p.Tokens[p.currIndex+2].Type == l.Name
}

func (p *LangParser) predeclareGlobalAt(index int) {
	if index+1 >= p.tokenSize {
		return
	}
	nameTok := p.Tokens[index+1]
	if nameTok.Type != l.Name || nameTok.Content == nil {
		return
	}
	p.ScopeCursor.DefineGlobalVariable(*nameTok.Content, []ast.Signature{ast.SignAny})
}

func (p *LangParser) checkPendingSymbols() {
	var errorMessages []string
	var diagnostics []codecontext.Diagnostic
	methodErrors := make(map[int][]pendingCallError) // keyed by line number
	funcErrors := make(map[int][]pendingCallError)
	questionErrors := make(map[int][]pendingCallError)
	methodErrorCount := 0
	funcErrorCount := 0
	questionErrorCount := 0

	for token, parseError := range p.aggregator.Errors {
		if !parseError.Deferred {
			errorMessages = append(errorMessages, token.BuildError(false, parseError.ErrorMessage))
			diagnostics = append(diagnostics, token.Diagnostic(parseError.ErrorMessage))
			continue
		}
		// try resolve global variables again
		if get, ok := parseError.Owner.(*variables.Get); ok && get.Global {
			signatures, resolved := p.ScopeCursor.ReferGlobalVariable(get.Name)
			if resolved {
				get.ValueSignature = signatures
				continue
			}
		} else if mc, ok := parseError.Owner.(*method.Call); ok {
			if _, sig := method.TestSignature(mc.Name, len(mc.Args)); sig != nil {
				continue
			}
			inputSigs := safeSignature(mc.On)
			allowedModules := method.DeriveAllowedModules(inputSigs)
			suggestion := method.FindBestSuggestion(mc.Name, allowedModules, mc.HintOutput())
			methodErrors[token.Column] = append(methodErrors[token.Column], pendingCallError{token: token, name: mc.Name, suggestion: suggestion})
			methodErrorCount++
			continue
		} else if q, ok := parseError.Owner.(*common.Question); ok {
			if common.IsKnownQuestion(q.Question) {
				continue
			}
			suggestion := common.FindBestQuestionSuggestion(q.Question)
			hint := ""
			if suggestion != "" {
				hint = "? " + suggestion
			}
			questionErrors[token.Column] = append(questionErrors[token.Column], pendingCallError{token: token, name: q.Question, suggestion: hint})
			questionErrorCount++
			continue
		} else if fc, ok := parseError.Owner.(*common.FuncCall); ok {
			if _, sig := common.TestSignature(fc.Name, len(fc.Args)); sig != nil {
				continue
			}
			suggestion := common.FindBestSuggestion(fc.Name)
			funcErrors[token.Column] = append(funcErrors[token.Column], pendingCallError{token: token, name: fc.Name, suggestion: suggestion})
			funcErrorCount++
			continue
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
		diagnostics = append(diagnostics, token.Diagnostic(parseError.ErrorMessage))
	}

	p.reportCompileErrors(errorMessages, diagnostics, methodErrors, funcErrors, questionErrors, methodErrorCount, funcErrorCount, questionErrorCount)
}

func (p *LangParser) reportCompileErrors(errorMessages []string, diagnostics []codecontext.Diagnostic, methodErrors, funcErrors, questionErrors map[int][]pendingCallError, methodErrorCount, funcErrorCount, questionErrorCount int) {
	methodBlocks := renderCallErrorGroups(methodErrors)
	funcBlocks := renderCallErrorGroups(funcErrors)
	questionBlocks := renderCallErrorGroups(questionErrors)
	errorMessages = append(errorMessages, methodBlocks...)
	errorMessages = append(errorMessages, funcBlocks...)
	errorMessages = append(errorMessages, questionBlocks...)
	diagnostics = append(diagnostics, callDiagnostics(methodErrors, "method")...)
	diagnostics = append(diagnostics, callDiagnostics(funcErrors, "function")...)
	diagnostics = append(diagnostics, callDiagnostics(questionErrors, "question")...)

	// Count each individual bad call, not each group block.
	groupBlocks := len(methodBlocks) + len(funcBlocks) + len(questionBlocks)
	totalErrors := len(errorMessages) - groupBlocks + methodErrorCount + funcErrorCount + questionErrorCount
	if p.strict && totalErrors > 0 {
		var errorWriter strings.Builder
		message := sugar.Format("compile failed with % syntax errors", strconv.Itoa(totalErrors))
		errorWriter.WriteString(message)
		errorWriter.WriteString(strings.Join(errorMessages, ""))
		panic(&codecontext.DiagnosticListError{
			Message:     message,
			Raw:         errorWriter.String(),
			Diagnostics: diagnostics,
		})
	}
}

func callDiagnostics(byLine map[int][]pendingCallError, kind string) []codecontext.Diagnostic {
	var diagnostics []codecontext.Diagnostic
	for _, items := range byLine {
		for _, item := range items {
			message := "No " + kind + " named " + item.name
			if kind == "method" {
				message = "No method named ." + item.name + "()"
			} else if kind == "function" {
				message = "No function named " + item.name + "()"
			} else if kind == "question" {
				message = "No question named ? " + item.name
			}
			if item.suggestion != "" && item.suggestion != "?" {
				message += ", did you mean " + item.suggestion + "?"
			}
			diagnostics = append(diagnostics, item.token.Diagnostic(message))
		}
	}
	return diagnostics
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
		tok := p.next()
		if !p.ScopeCursor.In(ScopeLoop) {
			tok.Error("break can only be used inside a loop")
		}
		return &control.Break{}
	case l.Yield:
		return p.yieldSmt()
	case l.Local:
		return p.localSmt()
	case l.Global:
		return p.globalSmt()
	case l.Func:
		if p.isAnonymousProcedureStart() || p.isProcedureDropdownStart() {
			return p.expr(0)
		}
		return p.funcSmt()
	case l.When:
		p.skip()
		if p.consume(l.Any) {
			return p.genericEvent()
		}
		return p.event()
	default:
		return p.expr(0)
	}
}

func (p *LangParser) yieldSmt() ast.Expr {
	tok := p.next()
	if !p.ScopeCursor.In(ScopeRetProc) {
		tok.Error("yield can only be used inside a returning procedure")
	}
	yieldName := "_result"
	expr := p.parse()
	if p.ScopeCursor.currScope.Type == ScopeRetProc && p.isNext(l.CloseCurly) {
		// just return the expr as is
		return expr
	}
	// _result = [  false, <expr>  ]
	transformedExpr := &variables.Set{
		Global: false,
		Name:   yieldName,
		Expr: &fundamentals.List{
			Elements: []ast.Expr{
				&fundamentals.Boolean{Value: false},
				expr,
			},
		},
	}
	return &fundamentals.Yield{
		Expr:            expr,
		TransformedExpr: transformedExpr,
		UseTransformed:  false,
	}
}

func (p *LangParser) genericEvent() ast.Expr {
	componentType := p.componentType()
	p.expect(l.Dot)
	eventTok := p.peek()
	eventName := p.name()
	var parameters []string
	if p.isNext(l.OpenCurve) {
		parameters = p.parameters()
	}
	if p.eventValidator != nil {
		if err := p.eventValidator(componentType, eventName, parameters); err != nil {
			eventTok.Error("%", err.Error())
		}
	}
	body := p.parseEventBody(parameters)
	return &components.GenericEvent{ComponentType: componentType, Event: eventName, Parameters: parameters, Body: body}
}

func (p *LangParser) event() ast.Expr {
	component := p.component()
	p.expect(l.Dot)
	eventTok := p.peek()
	eventName := p.name()
	var parameters []string
	if p.isNext(l.OpenCurve) {
		parameters = p.parameters()
	}
	if p.eventValidator != nil {
		if err := p.eventValidator(component.Type, eventName, parameters); err != nil {
			eventTok.Error("%", err.Error())
		}
	}
	body := p.parseEventBody(parameters)
	return &components.Event{
		ComponentName: component.Name,
		ComponentType: component.Type,
		Event:         eventName,
		Parameters:    parameters,
		Body:          body,
	}
}

func (p *LangParser) parseEventBody(parameters []string) []ast.Expr {
	vars := make([]ScopeVar, len(parameters))
	for i, param := range parameters {
		vars[i] = scopeVar(param, ast.SignOfEvent, ast.SignAny)
	}
	return p.body(ScopeEvent, vars...)
}

func (p *LangParser) funcSmt() ast.Expr {
	where := p.next()
	name := p.name()
	parameters := p.parameters()
	returning := p.consume(l.Assign)
	p.Resolver.Procedures[name] = &Procedure{Name: name, Parameters: parameters, Returning: returning}
	if returning {
		return p.retProcedure(where, name, parameters)
	}
	return p.voidProcedure(name, parameters)
}

func (p *LangParser) anonProcedure() ast.Expr {
	p.next()
	parameters := p.parameters()
	returning := p.consume(l.Assign)
	if returning {
		p.ScopeCursor.EnterAnonymousProcedure(ScopeRetProc)
		for _, parameter := range parameters {
			p.ScopeCursor.DefineVariable(parameter, []ast.Signature{ast.SignAny})
		}
		var result ast.Expr
		if p.consume(l.OpenCurly) {
			yieldParser := &YieldParser{Exprs: p.bodyUntilCurly()}
			result = &fundamentals.SmartBody{Body: yieldParser.ParseYield()}
			p.expect(l.CloseCurly)
		} else {
			result = p.parse()
		}
		p.ScopeCursor.Exit(ScopeRetProc)
		return &procedures.AnonProcedure{Parameters: parameters, Result: result, Returning: true}
	}
	vars := make([]ScopeVar, len(parameters))
	for i, param := range parameters {
		vars[i] = scopeVar(param, ast.SignAny)
	}
	p.expect(l.OpenCurly)
	p.ScopeCursor.EnterAnonymousProcedure(ScopeProc)
	for _, v := range vars {
		p.ScopeCursor.DefineVariable(v.Name, v.Sigs)
	}
	body := p.bodyUntilCurly()
	p.ScopeCursor.Exit(ScopeProc)
	p.expect(l.CloseCurly)
	return &procedures.AnonProcedure{Parameters: parameters, Body: body}
}

func (p *LangParser) retProcedure(where *l.Token, name string, parameters []string) ast.Expr {
	p.ScopeCursor.Enter(where, ScopeRetProc)
	for _, parameter := range parameters {
		p.ScopeCursor.DefineVariable(parameter, []ast.Signature{ast.SignAny})
	}
	var result ast.Expr
	if p.consume(l.OpenCurly) {
		yieldParser := &YieldParser{Exprs: p.bodyUntilCurly()}
		result = &fundamentals.SmartBody{Body: yieldParser.ParseYield()}
		p.expect(l.CloseCurly)
	} else {
		result = p.parse()
	}
	p.ScopeCursor.Exit(ScopeRetProc)
	return &procedures.RetProcedure{Name: name, Parameters: parameters, Result: result}
}

func (p *LangParser) voidProcedure(name string, parameters []string) ast.Expr {
	vars := make([]ScopeVar, len(parameters))
	for i, param := range parameters {
		vars[i] = scopeVar(param, ast.SignAny)
	}
	body := p.body(ScopeProc, vars...)
	return &procedures.VoidProcedure{Name: name, Parameters: parameters, Body: body}
}

func (p *LangParser) globalSmt() ast.Expr {
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

func (p *LangParser) localSmt() ast.Expr {
	// a clean full scope variable
	var names []string
	var values []ast.Expr
	for {
		locCurrIndex := p.currIndex
		if !p.consume(l.Local) {
			break
		}
		name := p.name()
		p.expect(l.Assign)
		preParseSumVarRefCount := p.GetSummatedVarRefCount(names)
		value := p.parse()
		postParseSumVarRefCount := p.GetSummatedVarRefCount(names)

		if postParseSumVarRefCount > preParseSumVarRefCount {
			// Since this variable depends on the last variable, we cannot include
			// it in the current set.
			p.currIndex = locCurrIndex
			break
		}

		names = append(names, name)
		values = append(values, value)
		p.ScopeCursor.DefineVariable(name, value.Signature())
	}
	// we have to parse rest of the body here
	body := p.bodyUntilCurly()
	if len(body) == 1 && body[0].Consumable() {
		return &variables.VarResult{Names: names, Values: values, Result: body[0]}
	}
	return &variables.Var{Names: names, Values: values, Body: body}
	//return &variables.VarStack{Names: names, Values: values}
}

func (p *LangParser) GetSummatedVarRefCount(names []string) int {
	summatedCount := 0
	for _, name := range names {
		refCount := p.ScopeCursor.GetVariableReferCount(name)
		if refCount == -1 {
			panic("Trying to query ref count of undeclared variable: " + name)
		}
		summatedCount += refCount
	}
	return summatedCount
}

func (p *LangParser) whileExpr() *control.While {
	whileTok := p.next()
	p.expect(l.OpenCurve)
	condition := p.parse()
	p.expect(l.CloseCurve)
	body := p.body(ScopeLoop)
	return &control.While{Where: whileTok, Condition: condition, Body: body}
}

func (p *LangParser) forExpr() ast.Expr {
	forTok := p.next()
	p.expect(l.OpenCurve)
	firstName := p.name()
	if p.consume(l.Comma) {
		return p.forEachPair(forTok, firstName)
	} else if p.consume(l.In) {
		return p.forEach(forTok, firstName)
	}
	return p.forRange(forTok, firstName)
}

func (p *LangParser) forEachPair(forTok *l.Token, keyName string) ast.Expr {
	valueName := p.name()
	p.expect(l.In)
	iterable := p.parse()
	p.expect(l.CloseCurve)
	body := p.body(ScopeLoop, scopeVar(keyName, ast.SignAny), scopeVar(valueName, ast.SignAny))
	return &control.EachPair{Where: forTok, KeyName: keyName, ValueName: valueName, Iterable: iterable, Body: body}
}

func (p *LangParser) forEach(forTok *l.Token, iName string) ast.Expr {
	iterable := p.parse()
	p.expect(l.CloseCurve)
	body := p.body(ScopeLoop, scopeVar(iName, ast.SignAny))
	return &control.Each{Where: forTok, IName: iName, Iterable: iterable, Body: body}
}

func (p *LangParser) forRange(forTok *l.Token, iName string) ast.Expr {
	p.expect(l.Colon)
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
	body := p.body(ScopeLoop, scopeVar(iName, ast.SignNumb))
	return &control.For{Where: forTok, IName: iName, From: from, To: to, By: by, Body: body}
}

func (p *LangParser) ifSmt() ast.Expr {
	where := p.next()
	var conditions []ast.Expr
	var bodies [][]ast.Expr

	p.expect(l.OpenCurve)
	conditions = append(conditions, p.expr(0))
	p.expect(l.CloseCurve)
	bodies = append(bodies, p.parseConditionBody())

	var elseBody []ast.Expr
	for p.notEOF() && p.consume(l.Else) {
		if p.consume(l.If) {
			p.expect(l.OpenCurve)
			conditions = append(conditions, p.expr(0))
			p.expect(l.CloseCurve)
			bodies = append(bodies, p.parseConditionBody())
		} else {
			elseBody = p.parseConditionBody()
			break
		}
	}
	return &control.If{Where: where, Conditions: conditions, Bodies: bodies, ElseBody: elseBody}
}

func (p *LangParser) parseConditionBody() []ast.Expr {
	if p.isNext(l.OpenCurly) {
		return p.body(ScopeIfBody)
	}
	return []ast.Expr{p.parse()}
}

// ScopeVar declares a variable immediately after entering a new scope.
type ScopeVar struct {
	Name string
	Sigs []ast.Signature
}

func scopeVar(name string, sigs ...ast.Signature) ScopeVar {
	return ScopeVar{Name: name, Sigs: sigs}
}

func (p *LangParser) body(scope ScopeType, vars ...ScopeVar) []ast.Expr {
	where := p.expect(l.OpenCurly)
	p.ScopeCursor.Enter(where, scope)
	for _, v := range vars {
		p.ScopeCursor.DefineVariable(v.Name, v.Sigs)
	}
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
		expr := p.parse()
		expressions = append(expressions, expr)
		p.consume(l.Comma)

		// no statements allowed after `break`
		switch expr.(type) {
		case *control.Break, *fundamentals.Yield:
			if p.notEOF() && !p.isNext(l.CloseCurly) {
				p.peek().Error("unreachable code after '%'", expr.String())
			}
		}
		switch expr.(type) {
		case *fundamentals.Yield:
			if p.ScopeCursor.In(ScopeLoop) {
				// we have to inject a break statement here
				expressions = append(expressions, &control.Break{})
			}
		}
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
		if p.isOnNewLine() {
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
		rightMinPrecedence := precedence + 1
		if isRightAssociativeOperator(opToken.Type) {
			rightMinPrecedence = precedence
		}
		right := p.expr(rightMinPrecedence)
		lBinExpr, leftRepeats := left.(*common.BinaryExpr)
		leftRepeats = leftRepeats && lBinExpr.CanRepeat(opToken.Type)
		rBinExpr, rightRepeats := right.(*common.BinaryExpr)
		rightRepeats = rightRepeats && rBinExpr.CanRepeat(opToken.Type)
		if opToken.HasFlag(l.PreserveOrder) && leftRepeats {
			// PreserveOrder operators prefer extending the left-hand chain.
			lBinExpr.Operands = append(lBinExpr.Operands, right)
		} else if rightRepeats {
			// Other repeatable operators can merge into the right-hand chain.
			rBinExpr.Operands = append([]ast.Expr{left}, rBinExpr.Operands...)
			left = rBinExpr
		} else if leftRepeats {
			lBinExpr.Operands = append(lBinExpr.Operands, right)
		} else {
			// a new binary node
			left = p.makeBinary(opToken, left, right)
		}
	}
	return left
}

func isRightAssociativeOperator(operator l.Type) bool {
	return operator == l.Power || operator == l.MatrixPower
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
	}
	opToken.Error("Unknown compound operator '%='", *opToken.Content)
	panic("unreached")
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
		return &list.Set{Where: listGet.Where, List: listGet.List, Index: listGet.Index, Value: right}, true
	} else if matrixGet, ok := left.(*astmatrix.GetCell); ok {
		return &astmatrix.SetCell{Where: matrixGet.Where, Matrix: matrixGet.Matrix, Dims: matrixGet.Dims, Value: right}, true
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
		case l.OpenCurve:
			if p.isOnNewLine() {
				break
			}
			left = &procedures.AnonCall{Procedure: left, Arguments: p.arguments()}
			continue
		case l.OpenSquare:
			if p.isOnNewLine() {
				break // '[' is on a new line — new statement, not index access
			}
			openSquare := p.next()
			if p.isNext(l.OpenSquare) {
				p.next()
				left = &astmatrix.GetCell{Where: openSquare, Matrix: left, Dims: p.matrixCellDims()}
				p.expect(l.CloseSquare)
				continue
			}
			// an index element access
			left = &list.Get{Where: openSquare, List: left, Index: p.parse()}
			p.expect(l.CloseSquare)
			continue
		}
		break
	}
	return left
}

func (p *LangParser) matrixCellDims() []ast.Expr {
	if p.isNext(l.CloseSquare) {
		p.peek().Error("Matrix cell access requires at least one dimension")
	}
	var dims []ast.Expr
	for p.notEOF() {
		dims = append(dims, p.expr(0))
		if !p.consume(l.Comma) {
			break
		}
	}
	p.expect(l.CloseSquare)
	return dims
}

func (p *LangParser) componentCall(compName string, compType string) ast.Expr {
	p.expect(l.Dot)
	resourceTok := p.peek()
	resource := p.name()
	if p.isNext(l.OpenCurve) {
		args := p.arguments()
		if p.methodValidator != nil {
			if err := p.methodValidator(compType, resource, len(args)); err != nil {
				resourceTok.Error("%", err.Error())
			}
		}
		return &components.MethodCall{
			ComponentName: compName,
			ComponentType: compType,
			Method:        resource,
			Args:          args,
		}
	} else if p.consume(l.Assign) {
		if p.propertyValidator != nil {
			if err := p.propertyValidator(compType, resource); err != nil {
				resourceTok.Error("%", err.Error())
			}
		}
		assignment := p.expr(0)
		return &components.PropertySet{
			ComponentName: compName,
			ComponentType: compType,
			Property:      resource,
			Value:         assignment,
		}
	}
	if p.propertyValidator != nil {
		if err := p.propertyValidator(compType, resource); err != nil {
			resourceTok.Error("%", err.Error())
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
		// Only treat as a transformer if the name is a known transformer signature
		// AND the arg count matches — otherwise the '{' belongs to the surrounding expression.
		if !p.isNext(l.OpenCurly) || !list.IsTransformer(name, len(args)) {
			return p.parseMethodCall(object, where, name, args)
		}
	}
	return p.parseTransformer(object, where, name, args)
}

func (p *LangParser) parseMethodCall(object ast.Expr, where *l.Token, name string, args []ast.Expr) ast.Expr {
	if name == "call" && len(args) == 1 {
		p.aggregator.MarkResolved(where)
		return &procedures.AnonCallInputList{Procedure: object, InputList: args[0]}
	}
	if name == "numArgs" && len(args) == 0 {
		p.aggregator.MarkResolved(where)
		return &procedures.NumArgs{Procedure: object}
	}
	call := &method.Call{Where: where, On: object, Name: name, Args: args}
	errorMessage, signature := method.TestSignature(name, len(args))
	if signature == nil {
		// Before treating as a bad method, check if it looks like a question (? keyword).
		if len(args) == 0 {
			if common.FindBestQuestionSuggestion(name) != "" {
				q := &common.Question{Where: where, On: object, Question: name, MethodCallSyntax: true}
				if common.IsKnownQuestion(name) {
					p.aggregator.MarkResolved(where)
				} else {
					p.aggregator.EnqueueSymbol(where, q, "")
				}
				return q
			}
		}
		if method.IsKnownMethod(name) {
			p.aggregator.EnqueueError(where, call, errorMessage)
		} else {
			p.aggregator.EnqueueSymbol(where, call, errorMessage)
		}
	} else {
		p.aggregator.MarkResolved(where)
	}
	return call
}

func (p *LangParser) parseTransformer(object ast.Expr, where *l.Token, name string, args []ast.Expr) ast.Expr {
	p.expect(l.OpenCurly)
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
		p.expect(l.RightArrow)
	}
	transformer := p.exprOrSmartBody()
	p.ScopeCursor.Exit(ScopeTypeTransform)
	p.expect(l.CloseCurly)
	errorMessage, signature := list.TestSignature(name, len(args), len(namesUsed))
	if signature == nil {
		p.aggregator.EnqueueError(where, object, errorMessage)
	} else {
		p.aggregator.MarkResolved(where)
	}
	return &list.Transformer{
		Where:       where,
		List:        object,
		Name:        name,
		Args:        args,
		Names:       namesUsed,
		Transformer: transformer,
	}
}

func (p *LangParser) exprOrSmartBody() ast.Expr {
	locCurrIndex := p.currIndex
	simpleExpr := p.parse()
	if simpleExpr.Consumable() {
		return simpleExpr
	}
	p.currIndex = locCurrIndex
	// we gotta do a manual smart body here
	p.ScopeCursor.Enter(p.peek(), ScopeSmartBody)
	smartBody := &fundamentals.SmartBody{Body: p.bodyUntilCurly()}
	p.ScopeCursor.Exit(ScopeSmartBody)
	return smartBody
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
		return common.MakeFuncCall("neg", p.expr(l.PrecedenceOf(l.BinaryL2)))
	case l.If:
		p.back()
		return p.ifSmt()
	case l.WalkAll:
		return &fundamentals.WalkAll{}
	case l.Func:
		if p.isNext(l.Dot) {
			p.expect(l.Dot)
			return &procedures.GetWithDropdown{Name: p.name()}
		}
		p.back()
		return p.anonProcedure()
	default:
		if p.isMatrixLiteralStart(token) {
			return p.matrixLiteral(token)
		}
		if token.HasFlag(l.Value) {
			return p.checkCall(token)
		}
		token.Error("Unexpected! %", token.String())
		panic("") // unreachable
	}
}

func (p *LangParser) isMatrixLiteralStart(token *l.Token) bool {
	return token.Type == l.Name &&
		token.Content != nil &&
		*token.Content == "matrix" &&
		p.currIndex+1 < p.tokenSize &&
		p.Tokens[p.currIndex].Type == l.OpenSquare &&
		p.Tokens[p.currIndex+1].Type == l.OpenSquare &&
		p.Tokens[p.currIndex].Column == token.Column
}

func (p *LangParser) matrixLiteral(where *l.Token) ast.Expr {
	p.expect(l.OpenSquare)
	var rows [][]ast.Expr
	for p.notEOF() {
		p.expect(l.OpenSquare)
		rows = append(rows, p.matrixLiteralRow())
		if !p.consume(l.Comma) {
			break
		}
	}
	p.expect(l.CloseSquare)
	return &astmatrix.Create{Where: where, Rows: rows}
}

func (p *LangParser) matrixLiteralRow() []ast.Expr {
	if p.isNext(l.CloseSquare) {
		p.peek().Error("Matrix literal rows require at least one cell")
	}
	var row []ast.Expr
	for p.notEOF() {
		row = append(row, p.expr(0))
		if !p.consume(l.Comma) {
			break
		}
	}
	p.expect(l.CloseSquare)
	return row
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
	if nameExpr, ok := value.(*variables.Get); ok && !nameExpr.Global && p.isNext(l.OpenCurve) && !p.isOnNewLine() {
		arguments := p.arguments()
		if nameExpr.Name == "getFunc" && len(arguments) == 1 {
			p.aggregator.MarkResolved(nameExpr.Where)
			return &procedures.GetWithName{Name: arguments[0]}
		}
		// check for in-built function call
		errorMessage, funcCallSignature := common.TestSignature(nameExpr.Name, len(arguments))
		if funcCallSignature != nil {
			p.aggregator.MarkResolved(nameExpr.Where)
			return &common.FuncCall{Where: nameExpr.Where, Name: nameExpr.Name, Args: arguments}
		}
		if common.IsKnownFunction(nameExpr.Name) {
			fc := &common.FuncCall{Where: nameExpr.Where, Name: nameExpr.Name, Args: arguments}
			p.aggregator.EnqueueError(nameExpr.Where, fc, errorMessage)
			return fc
		}
		if len(nameExpr.ValueSignature) > 0 {
			p.aggregator.MarkResolved(nameExpr.Where)
			return &procedures.AnonCall{Procedure: nameExpr, Arguments: arguments}
		}
		// check for a user defined procedure
		procedureErrorMessage, procedureSignature := p.Resolver.ResolveProcedure(nameExpr.Name, len(arguments))
		if procedureSignature != nil {
			p.aggregator.MarkResolved(nameExpr.Where)
			return &procedures.Call{
				Where:      nameExpr.Where,
				Name:       nameExpr.Name,
				Parameters: procedureSignature.Parameters,
				Arguments:  arguments,
				Returning:  procedureSignature.Returning,
			}
		}
		// Unknown calls may be forward-declared procedures. Keep them resolvable, but
		// retain the built-in spelling hint if late resolution still fails.
		if common.FindBestSuggestion(nameExpr.Name) != "" {
			funcCall := &procedures.Call{Where: nameExpr.Where, Name: nameExpr.Name, Arguments: arguments}
			p.aggregator.EnqueueSymbol(nameExpr.Where, funcCall, "No function named "+nameExpr.Name+"()")
			return funcCall
		}
		// Unknown — fill in a template that may be resolved later (forward-declared procedure).
		funcCall := &procedures.Call{Where: nameExpr.Where, Name: nameExpr.Name, Arguments: arguments}
		p.aggregator.EnqueueSymbol(nameExpr.Where, funcCall, procedureErrorMessage)
		return funcCall
	}
	return value
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
		signatures, found := p.ScopeCursor.ReferVariable(*t.Content)
		get := &variables.Get{Where: t, Global: false, Name: *t.Content, ValueSignature: signatures}
		if !found {
			p.aggregator.EnqueueSymbol(t, get, "Cannot find symbol '"+*t.Content+"'")
		}
		return get
	case l.This:
		p.expect(l.Dot)
		nameToken := p.expect(l.Name)
		name := *nameToken.Content
		signatures, found := p.ScopeCursor.ReferGlobalVariable(name)
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

// isOnNewLine reports whether the current (peeked) token is on a different
// line than the last consumed token. Used to prevent cross-line postfix
// continuation — if '[' or '(' appears on a new line it is a new statement,
// not a continuation of the previous expression (mirrors Kotlin's rule).
func (p *LangParser) isOnNewLine() bool {
	if p.currIndex == 0 {
		return false
	}
	return p.peek().Column > p.Tokens[p.currIndex-1].Column
}
