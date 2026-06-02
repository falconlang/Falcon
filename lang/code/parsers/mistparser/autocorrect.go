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

// walkAndCorrect recursively visits every expression in the tree, corrects
// auto-correctable names in-place, and records each change as a SourcePatch
// so the original source can be reconstructed via ReconstructedSource().
func (p *LangParser) walkAndCorrect(expr ast.Expr) {
	if !p.autoCorrect || expr == nil {
		return
	}
	switch e := expr.(type) {

	case *method.Call:
		var corrections []method.Correction
		method.CorrectChainAndCollect(e, &corrections)
		for _, c := range corrections {
			p.patches = append(p.patches, SourcePatch{
				Line:  c.Where.Column,
				Start: c.Where.Row - len(c.OldName),
				End:   c.Where.Row,
				Text:  c.Replacement,
			})
		}
		for _, arg := range e.Args {
			p.walkAndCorrect(arg)
		}

	// Variable declarations
	case *variables.Var:
		for _, v := range e.Values {
			p.walkAndCorrect(v)
		}
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *variables.VarResult:
		for _, v := range e.Values {
			p.walkAndCorrect(v)
		}
		p.walkAndCorrect(e.Result)
	case *variables.Set:
		p.walkAndCorrect(e.Expr)
	case *variables.Global:
		p.walkAndCorrect(e.Value)

	// Procedures
	case *procedures.RetProcedure:
		p.walkAndCorrect(e.Result)
	case *procedures.VoidProcedure:
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *procedures.Call:
		for _, arg := range e.Arguments {
			p.walkAndCorrect(arg)
		}

	// Common expressions
	case *common.FuncCall:
		if !common.IsKnownFunction(e.Name) {
			if best := common.FindBestSuggestion(e.Name); best != "" {
				oldName := e.Name
				e.Name = best
				p.patches = append(p.patches, SourcePatch{
					Line:  e.Where.Column,
					Start: e.Where.Row - len(oldName),
					End:   e.Where.Row,
					Text:  best,
				})
			}
		}
		for _, arg := range e.Args {
			p.walkAndCorrect(arg)
		}
	case *common.BinaryExpr:
		for _, op := range e.Operands {
			p.walkAndCorrect(op)
		}
	case *common.Question:
		oldQuestion := e.Question
		if !common.IsKnownQuestion(e.Question) {
			if best := common.FindBestQuestionSuggestion(e.Question); best != "" {
				e.Question = best
			}
		}
		if e.MethodCallSyntax {
			// Source had .name() or .isName() — patch replaces dot + name + () with ? keyword.
			// The leading space preserves the visual gap that the dot occupied (e.g. s.isString() → s ? text).
			p.patches = append(p.patches, SourcePatch{
				Line:  e.Where.Column,
				Start: e.Where.Row - len(oldQuestion) - 1, // include the dot
				End:   e.Where.Row + 2,                    // include ()
				Text:  " ? " + e.Question,
			})
		} else if e.Question != oldQuestion {
			// Source used ? syntax but with wrong keyword — simple in-place rename.
			p.patches = append(p.patches, SourcePatch{
				Line:  e.Where.Column,
				Start: e.Where.Row - len(oldQuestion),
				End:   e.Where.Row,
				Text:  e.Question,
			})
		}
		p.walkAndCorrect(e.On)

	// Control flow
	case *control.If:
		for _, cond := range e.Conditions {
			p.walkAndCorrect(cond)
		}
		for _, body := range e.Bodies {
			for _, b := range body {
				p.walkAndCorrect(b)
			}
		}
		for _, b := range e.ElseBody {
			p.walkAndCorrect(b)
		}
	case *control.For:
		p.walkAndCorrect(e.From)
		p.walkAndCorrect(e.To)
		p.walkAndCorrect(e.By)
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *control.While:
		p.walkAndCorrect(e.Condition)
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *control.Each:
		p.walkAndCorrect(e.Iterable)
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *control.EachPair:
		p.walkAndCorrect(e.Iterable)
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *control.Do:
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
		p.walkAndCorrect(e.Result)

	// List operations
	case *list.Transformer:
		p.walkAndCorrect(e.List)
		for _, arg := range e.Args {
			p.walkAndCorrect(arg)
		}
		p.walkAndCorrect(e.Transformer)
	case *list.Get:
		p.walkAndCorrect(e.List)
		p.walkAndCorrect(e.Index)
	case *list.Set:
		p.walkAndCorrect(e.List)
		p.walkAndCorrect(e.Index)
		p.walkAndCorrect(e.Value)

	// Matrix operations
	case *astmatrix.GetCell:
		p.walkAndCorrect(e.Matrix)
		for _, dim := range e.Dims {
			p.walkAndCorrect(dim)
		}
	case *astmatrix.SetCell:
		p.walkAndCorrect(e.Matrix)
		for _, dim := range e.Dims {
			p.walkAndCorrect(dim)
		}
		p.walkAndCorrect(e.Value)

	// Fundamentals that can contain sub-expressions
	case *fundamentals.SmartBody:
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *fundamentals.List:
		for _, item := range e.Elements {
			p.walkAndCorrect(item)
		}
	case *fundamentals.Dictionary:
		for _, item := range e.Elements {
			p.walkAndCorrect(item)
		}
	case *fundamentals.Pair:
		p.walkAndCorrect(e.Key)
		p.walkAndCorrect(e.Value)
	case *fundamentals.Not:
		p.walkAndCorrect(e.Expr)
	case *common.Transform:
		p.walkAndCorrect(e.On)

	// Component event handlers
	case *components.Event:
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *components.GenericEvent:
		for _, b := range e.Body {
			p.walkAndCorrect(b)
		}
	case *components.PropertySet:
		p.walkAndCorrect(e.Value)
	case *components.GenericPropertySet:
		p.walkAndCorrect(e.Value)
	case *components.MethodCall:
		for _, arg := range e.Args {
			p.walkAndCorrect(arg)
		}
	case *components.GenericMethodCall:
		for _, arg := range e.Args {
			p.walkAndCorrect(arg)
		}
	}
}
