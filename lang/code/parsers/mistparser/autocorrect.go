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
)

// safeSignature returns expr.Signature() without panicking; falls back to [SignAny].
func safeSignature(expr ast.Expr) (sigs []ast.Signature) {
	defer func() {
		if recover() != nil {
			sigs = []ast.Signature{ast.SignAny}
		}
	}()
	return expr.Signature()
}

// walkAndCorrect recursively visits every expression in the tree and corrects
// method chain type mismatches (via method.CorrectChain) before the main
// Signature() pass fires any type errors.
func walkAndCorrect(expr ast.Expr) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {

	case *method.Call:
		method.CorrectChain(e)
		for _, arg := range e.Args {
			walkAndCorrect(arg)
		}

	// Variable declarations
	case *variables.SimpleVar:
		walkAndCorrect(e.Value)
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *variables.Var:
		for _, v := range e.Values {
			walkAndCorrect(v)
		}
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *variables.VarResult:
		for _, v := range e.Values {
			walkAndCorrect(v)
		}
		walkAndCorrect(e.Result)
	case *variables.Set:
		walkAndCorrect(e.Expr)
	case *variables.Global:
		walkAndCorrect(e.Value)

	// Procedures
	case *procedures.RetProcedure:
		walkAndCorrect(e.Result)
	case *procedures.VoidProcedure:
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *procedures.Call:
		for _, arg := range e.Arguments {
			walkAndCorrect(arg)
		}

	// Common expressions
	case *common.FuncCall:
		if !common.IsKnownFunction(e.Name) {
			if best := common.FindBestSuggestion(e.Name); best != "" {
				e.Name = best
			}
		}
		for _, arg := range e.Args {
			walkAndCorrect(arg)
		}
	case *common.BinaryExpr:
		for _, op := range e.Operands {
			walkAndCorrect(op)
		}
	case *common.Question:
		if !common.IsKnownQuestion(e.Question) {
			if best := common.FindBestQuestionSuggestion(e.Question); best != "" {
				e.Question = best
			}
		}
		walkAndCorrect(e.On)

	// Control flow
	case *control.If:
		for _, cond := range e.Conditions {
			walkAndCorrect(cond)
		}
		for _, body := range e.Bodies {
			for _, b := range body {
				walkAndCorrect(b)
			}
		}
		for _, b := range e.ElseBody {
			walkAndCorrect(b)
		}
	case *control.For:
		walkAndCorrect(e.From)
		walkAndCorrect(e.To)
		walkAndCorrect(e.By)
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *control.While:
		walkAndCorrect(e.Condition)
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *control.Each:
		walkAndCorrect(e.Iterable)
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *control.EachPair:
		walkAndCorrect(e.Iterable)
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *control.Do:
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
		walkAndCorrect(e.Result)

	// List operations
	case *list.Transformer:
		walkAndCorrect(e.List)
		for _, arg := range e.Args {
			walkAndCorrect(arg)
		}
		walkAndCorrect(e.Transformer)
	case *list.Get:
		walkAndCorrect(e.List)
		walkAndCorrect(e.Index)
	case *list.Set:
		walkAndCorrect(e.List)
		walkAndCorrect(e.Index)
		walkAndCorrect(e.Value)

	// Fundamentals that can contain sub-expressions
	case *fundamentals.SmartBody:
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *fundamentals.List:
		for _, item := range e.Elements {
			walkAndCorrect(item)
		}

	// Component event handlers
	case *components.Event:
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *components.GenericEvent:
		for _, b := range e.Body {
			walkAndCorrect(b)
		}
	case *components.PropertySet:
		walkAndCorrect(e.Value)
	case *components.GenericPropertySet:
		walkAndCorrect(e.Value)
	case *components.MethodCall:
		for _, arg := range e.Args {
			walkAndCorrect(arg)
		}
	case *components.GenericMethodCall:
		for _, arg := range e.Args {
			walkAndCorrect(arg)
		}
	}
}
