package mistparser

import (
	"Falcon/code/ast"
	"Falcon/code/ast/control"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/list"
	"Falcon/code/ast/variables"
	l "Falcon/code/lex"
)

type path struct {
	frames []Frame
	yield  *fundamentals.Yield
}

func makePath(frames []Frame, yield *fundamentals.Yield) path {
	p := path{}
	p.frames = append(p.frames, frames...)
	p.yield = yield
	return p
}

func (p *path) hasLoopInPath() bool {
	for _, ft := range p.frames {
		if ft.FrameType == FrameTypeLoop {
			return true
		}
	}
	return false
}

type YieldParser struct {
	Exprs []ast.Expr

	paths     []path
	pathIndex int

	localResultDeclared bool //  if `_result = [true, false]` has been added
}

func (y *YieldParser) ParseYield() []ast.Expr {
	y.pathIndex = 0

	y.mapRouteToYields(y.Exprs, []Frame{})
	//for _, path := range y.paths {
	//	for i, frameType := range path {
	//		print(frameType.String())
	//		if i != len(path)-1 {
	//			print("  -->  ")
	//		}
	//		println()
	//	}
	//}
	return y.edits(y.Exprs)
}

func (y *YieldParser) edits(exprs []ast.Expr) []ast.Expr {
	var newExprs []ast.Expr
outerLoop:
	for k, expr := range exprs {
		switch e := expr.(type) {
		case *control.If:
			allBodiesYield := true
			for j := range e.Bodies {
				currPath := y.nextPath()
				var lastPath *path = nil
				if j > 0 {
					lastPath = &y.paths[y.pathIndex-2]
				}
				if currPath.yield == nil {
					allBodiesYield = false
				}
				// a loop makes for a potential yield, not confirmed, so debranch If
				if currPath.hasLoopInPath() && currPath.yield != nil {
					currPath.yield.UseTransformed = true
					requiresDeclaration := !y.localResultDeclared
					if requiresDeclaration {
						// this ensures the child edit() calls do not declare it themselves
						y.localResultDeclared = true
					}
					var addedExprs []ast.Expr
					if j+1 < len(e.Bodies) {
						// decompose if branch
						nextIf := e.Decompose(j + 1)
						// wrap rest of the code in yield var check
						nextIf.ElseBody = append(nextIf.ElseBody, y.yieldResultQuery(y.edits(exprs[k+1:])))
						addedExprs = append(addedExprs, y.addLoopBreaking(currPath, e))
						addedExprs = append(addedExprs, nextIf)
					} else {
						// already decomposed position, simply wrap rest of the code in yield var check
						addedExprs = append(addedExprs, y.addLoopBreaking(currPath, e))
						addedExprs = append(addedExprs, y.yieldResultQuery(y.edits(exprs[k+1:])))
					}
					if !requiresDeclaration {
						newExprs = append(newExprs, addedExprs...)
					} else {
						newExprs = append(newExprs, y.declareLocalResult(addedExprs))
					}
					break outerLoop
				}
				if lastPath != nil {
					// last path didn't have yield, but this one does, so debranchIf
					if lastPath.yield == nil && currPath.yield != nil {
						newExprs = y.wrapInResultCheck(currPath.yield, exprs[k+1:], newExprs, e)
						break outerLoop
					} else if lastPath.yield != nil && currPath.yield == nil {
						// OR last path had yield, but this one doesn't, so debranch to else
						nextIf := e.Decompose(j)
						e.ElseBody = []ast.Expr{nextIf}
						e.ElseBody = append(e.ElseBody, exprs[k+1:]...)
						y.pathIndex--
						e.ElseBody = y.edits(e.ElseBody)
						newExprs = append(newExprs, e)
						break outerLoop
					}
				}
				if j+1 == len(e.Bodies) && currPath.yield == nil {
					newExprs = append(newExprs, e)
					continue outerLoop
				}
			}
			// set rest of the body as else branch
			if e.ElseBody == nil {
				e.ElseBody = y.edits(exprs[k+1:])
				newExprs = append(newExprs, e)
				break outerLoop
			}
			// else body exists
			elsePath := y.nextPath()
			// append rest of the body to else branch
			if elsePath.yield == nil {
				e.ElseBody = append(e.ElseBody, y.edits(exprs[k+1:])...)
				newExprs = append(newExprs, e)
				break outerLoop
			} else if len(e.Bodies) == 1 && !allBodiesYield {
				// only else has a yield, inverse it
				thenBranch := e.Bodies[0]
				e.Bodies[0] = e.ElseBody
				// append rest of the body to else branch
				e.ElseBody = append(thenBranch, y.edits(exprs[k+1:])...)
				// invert the condition logic
				switch c := e.Conditions[0].(type) {
				case *fundamentals.Not:
					e.Conditions[0] = c.Expr
					break
				case *fundamentals.Boolean:
					e.Conditions[0] = &fundamentals.Boolean{Value: !c.Value}
					break
				default:
					e.Conditions[0] = &fundamentals.Not{Expr: c}
				}
				newExprs = append(newExprs, e)
				break outerLoop
			} else {
				// wrap rest of the body in result check
				newExprs = y.wrapInResultCheck(elsePath.yield, exprs[k+1:], newExprs, e)
				break outerLoop
			}
		case *control.For, *control.While, *control.Each, *control.EachPair:
			// wrap rest of the body in a check
			p := y.nextPath()
			if p.yield != nil {
				newExprs = y.wrapInResultCheck(p.yield, exprs[k+1:], newExprs, y.addLoopBreaking(p, e))
				break outerLoop
			}
			newExprs = append(newExprs, expr)
		case *variables.Var:
			e.Body = y.edits(e.Body)
			newExprs = append(newExprs, e)
			break outerLoop
		default:
			newExprs = append(newExprs, expr)
		}
	}
	return newExprs
}

func (y *YieldParser) addLoopBreaking(p path, currFor ast.Expr) ast.Expr {
	totalLoops := 0
	currIndexMatch := 0
	for k, frame := range p.frames {
		if frame.FrameType == FrameTypeLoop {
			totalLoops++
		}
		if frame.Expr == currFor {
			currIndexMatch = k
		}
	}
	loopIndex := 0
	// `if (!_result[1]) break`
	conditionalBreaking := &control.If{
		Conditions: []ast.Expr{&fundamentals.Not{Expr: y.localResultQuery("1")}},
		Bodies:     [][]ast.Expr{{&control.Break{}}},
	}
	for _, frame := range p.frames {
		if frame.FrameType != FrameTypeLoop {
			continue
		}
		// add breaking for all loops except the most inner one (has a break)
		if loopIndex != totalLoops-1 {
			switch a := frame.Expr.(type) {
			case *control.For:
				a.Body = append(a.Body, conditionalBreaking)
				break
			case *control.While:
				a.Body = append(a.Body, conditionalBreaking)
				break
			case *control.Each:
				a.Body = append(a.Body, conditionalBreaking)
				break
			case *control.EachPair:
				a.Body = append(a.Body, conditionalBreaking)
				break
			}
			loopIndex++
		}
	}
	return p.frames[currIndexMatch].Expr
}

func (y *YieldParser) wrapInResultCheck(yield *fundamentals.Yield, restOfTheBody []ast.Expr, newExprs []ast.Expr, currExpr ast.Expr) []ast.Expr {
	yield.UseTransformed = true
	yieldQuery := y.yieldResultQuery(y.edits(restOfTheBody))
	if y.localResultDeclared {
		newExprs = append(newExprs, currExpr)
		newExprs = append(newExprs, yieldQuery)
	} else {
		newExprs = append(newExprs, y.declareLocalResult([]ast.Expr{currExpr, yieldQuery}))
	}
	return newExprs
}

func (y *YieldParser) yieldResultQuery(restOfTheBody []ast.Expr) ast.Expr {
	// `if (_result[1]) { ... <default_yield> } else _result[2]`
	return &control.If{
		Conditions: []ast.Expr{y.localResultQuery("1")},
		Bodies:     [][]ast.Expr{restOfTheBody},
		ElseBody:   []ast.Expr{y.localResultQuery("2")},
	}
}

func (y *YieldParser) localResultQuery(index string) *list.Get {
	// `_result[n]`
	return &list.Get{
		Where: l.MakeFakeToken(l.OpenSquare),
		List: &variables.Get{
			Where:          l.MakeFakeToken(l.Name),
			Global:         false,
			Name:           "_result",
			ValueSignature: []ast.Signature{ast.SignList},
		},
		Index: &fundamentals.Number{Content: index},
	}
}

func (y *YieldParser) declareLocalResult(restOfTheBody []ast.Expr) ast.Expr {
	y.localResultDeclared = true
	// `_result = [true, false]`,
	// where the first var indicates if unset, second var holds the value
	return &variables.Var{
		Names: []string{"_result"},
		Values: []ast.Expr{&fundamentals.List{
			Elements: []ast.Expr{&fundamentals.Boolean{Value: true}, &fundamentals.Boolean{Value: false}},
		}},
		Body: restOfTheBody,
	}
}

func (y *YieldParser) mapRouteToYields(traverseExprs []ast.Expr, frames []Frame) {
	if len(traverseExprs) == 0 {
		return
	}
	// check if the last expression is yield
	switch yield := traverseExprs[len(traverseExprs)-1].(type) {
	case *fundamentals.Yield:
		y.paths = append(y.paths, makePath(frames, yield))
		return
	}
	// or else the last second expression (in case of loop yield)
	if len(traverseExprs) > 1 {
		switch yield := traverseExprs[len(traverseExprs)-2].(type) {
		case *fundamentals.Yield:
			y.paths = append(y.paths, makePath(frames, yield))
			return
		}
	}
	handled := false
	for _, expr := range traverseExprs {
		switch e := expr.(type) {
		case *control.If:
			{
				handled = true
				for _, body := range e.Bodies {
					y.mapRouteToYields(body, AppendFrame(frames, FrameTypeIf, e))
				}
				y.mapRouteToYields(e.ElseBody, AppendFrame(frames, FrameTypeIf, e))
				continue
			}
		case *control.For:
			handled = true
			y.mapRouteToYields(e.Body, AppendFrame(frames, FrameTypeLoop, e))
			continue
		case *control.While:
			handled = true
			y.mapRouteToYields(e.Body, AppendFrame(frames, FrameTypeLoop, e))
			continue
		case *control.Each:
			handled = true
			y.mapRouteToYields(e.Body, AppendFrame(frames, FrameTypeLoop, e))
			continue
		case *control.EachPair:
			handled = true
			y.mapRouteToYields(e.Body, AppendFrame(frames, FrameTypeLoop, e))
			continue
		case *fundamentals.SmartBody:
			handled = true
			y.mapRouteToYields(e.Body, AppendFrame(frames, FrameTypeSmartBody, e))
			continue
		case *variables.VarResult:
			handled = true
			y.mapRouteToYields([]ast.Expr{e.Result}, AppendFrame(frames, FrameTypeVar, e))
			continue
		case *variables.Var:
			handled = true
			y.mapRouteToYields(e.Body, AppendFrame(frames, FrameTypeVar, e))
			continue
		}
	}
	if !handled {
		y.paths = append(y.paths, makePath(frames, nil))
	}
}

func (y *YieldParser) nextPath() path {
	if y.pathIndex >= len(y.paths) {
		panic("yield parser: no more paths available (index out of range)")
	}
	p := y.paths[y.pathIndex]
	y.pathIndex += 1
	return p
}
