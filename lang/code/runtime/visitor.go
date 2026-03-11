package runtime

import "Falcon/code/ast"

// Visitor is the interface for the interpreted runtime.
// Because ast.Expr does not have an Accept method, dispatch is done via
// a type switch inside Eval — the idiomatic Go visitor approach.
type Visitor interface {
	Eval(expr ast.Expr) Value
}
