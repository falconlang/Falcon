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
		if _, ok := s.Body[0].(*SmartBody); ok {
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
	return s.createDoSmt(s.Body[len(s.Body)-1], s.Body[:len(s.Body)-1])
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
			if !doResult.Consumable() {
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

func (s *SmartBody) Continuous() bool {
	return false
}

func (s *SmartBody) Consumable() bool {
	return true
}

func (s *SmartBody) Signature() []ast.Signature {
	for _, expr := range s.Body {
		expr.Signature()
	}
	return s.Body[len(s.Body)-1].Signature()
}
