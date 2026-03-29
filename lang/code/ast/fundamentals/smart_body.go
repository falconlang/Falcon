package fundamentals

import (
	"Falcon/code/ast"
	"Falcon/code/ast/variables"
	"Falcon/code/sugar"
)

type SmartBody struct {
	Body []ast.Expr
}

func (s *SmartBody) String() string {
	if len(s.Body) == 1 {
		if _, ok := s.Body[0].(*variables.VarResult); ok {
			return s.Body[0].String()
		}
	}
	return sugar.Format("{\n%}", ast.PadBody(s.Body))
}

func (s *SmartBody) Blockly(flags ...bool) ast.Block {
	// a single expression, just inline it
	if v, ok := s.Body[0].(*variables.Var); ok {
		// it's a var body, but we want a var result!
		var doExpr ast.Block
		if len(v.Body) > 0 {
			doExpr = s.createDoSmt(v.Body[len(v.Body)-1], v.Body[:len(v.Body)-1])
		} else {
			doExpr = createEmptyDoSmt(v)
		}
		return s.createLocalResult(v.Names, v.Values, doExpr)
	}
	if len(s.Body) == 1 {
		return s.Body[0].Blockly(flags...)
	}
	// Hoist SimpleVar declarations first so doBody no longer re-declares them.
	namesLocal, valsLocal := s.mutateVars()

	// prepare a do expression out of the then
	doExpr := s.createDoSmt(s.Body[len(s.Body)-1], s.Body[:len(s.Body)-1])

	if len(namesLocal) == 0 {
		// no variables declared in the then, a do expression is enough
		return doExpr
	}
	// We'd need to use a local result expression
	return s.createLocalResult(namesLocal, valsLocal, doExpr)
}

func (s *SmartBody) createLocalResult(names []string, values []ast.Expr, doExpr ast.Block) ast.Block {
	return ast.Block{
		Type:     "local_declaration_expression",
		Mutation: &ast.Mutation{LocalNames: ast.MakeLocalNames(names...)},
		Fields:   ast.ToFields("VAR", names),
		Values: append(ast.ValuesByPrefix("DECL", values),
			ast.Value{Name: "RETURN", Block: doExpr}),
	}
}

func (s *SmartBody) createDoSmt(doResult ast.Expr, doBody []ast.Expr) ast.Block {
	var doExpr ast.Block
	if len(doBody) == 0 {
		if v, ok := doResult.(*variables.Var); ok {
			// it's a var body, but we want a var result!
			if len(v.Body) > 0 {
				doExpr = s.createDoSmt(v.Body[len(v.Body)-1], v.Body[:len(v.Body)-1])
			} else {
				doExpr = createEmptyDoSmt(v)
			}
			return s.createLocalResult(v.Names, v.Values, doExpr)
		}
		doExpr = doResult.Blockly(false)
	} else {
		if v, ok := doResult.(*variables.Var); ok {
			// The result is a var-body (non-consumable), but it contains a
			// consumable result buried inside. Build a local_declaration_expression
			// for the Var, then wrap the whole thing in a controls_do_then_return.
			var innerDoExpr ast.Block
			if len(v.Body) > 0 {
				innerDoExpr = s.createDoSmt(v.Body[len(v.Body)-1], v.Body[:len(v.Body)-1])
			} else {
				innerDoExpr = createEmptyDoSmt(v)
			}
			valueExpr := s.createLocalResult(v.Names, v.Values, innerDoExpr)
			doExpr = ast.Block{
				Type:       "controls_do_then_return",
				Statements: ast.OptionalStatement("STM", doBody),
				Values:     []ast.Value{{Name: "VALUE", Block: valueExpr}},
			}
		} else {
			resultExpr := doResult.Blockly(false)
			if !doResult.Consumable(false) {
				panic("Cannot include a statement for the required variable result")
			}
			doExpr = ast.Block{
				Type:       "controls_do_then_return",
				Statements: ast.OptionalStatement("STM", doBody),
				// TODO: we have set the flag to false, previously was true, verify effects
				Values: []ast.Value{{Name: "VALUE", Block: resultExpr}},
			}
		}
	}
	return doExpr
}

func createEmptyDoSmt(v *variables.Var) ast.Block {
	return ast.Block{
		Type:   "lexical_variable_get",
		Fields: []ast.Field{{Name: "VAR", Value: v.Names[len(v.Names)-1]}},
	}
}

// mutateVars hoists SimpleVar declarations out of the body into DECL values.
// It removes SimpleVar entries from the body (they become DECL0, DECL1, …) and
// returns the extracted names and initial values. The last body element (result) is preserved.
func (s *SmartBody) mutateVars() ([]string, []ast.Expr) {
	var names []string
	var values []ast.Expr
	last := s.Body[len(s.Body)-1]
	var filtered []ast.Expr
	for _, expr := range s.Body[:len(s.Body)-1] {
		if e, ok := expr.(*variables.SimpleVar); ok {
			names = append(names, e.Name)
			values = append(values, e.Value)
			// removed from body — hoisted to DECL
		} else {
			filtered = append(filtered, expr)
		}
	}
	s.Body = append(filtered, last)
	return names, values
}

func (s *SmartBody) Continuous() bool {
	return false
}

func (s *SmartBody) Consumable(flags ...bool) bool {
	return true
}

func (s *SmartBody) Signature() []ast.Signature {
	return s.Body[len(s.Body)-1].Signature()
}
