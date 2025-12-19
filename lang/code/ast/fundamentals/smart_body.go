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
	// prepare a do expression out of the then
	doExpr := s.createDoSmt(s.Body[len(s.Body)-1], s.Body[:len(s.Body)-1])

	var namesLocal = s.mutateVars()
	if len(namesLocal) == 0 {
		// no variables declared in the then, a do expression is enough
		return doExpr
	}
	// We'd need to use a local result expression
	var defaultLocalVals []ast.Expr
	for k := range defaultLocalVals {
		defaultLocalVals[k] = &Boolean{Value: false}
	}
	return s.createLocalResult(namesLocal, defaultLocalVals, doExpr)
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
		if !doResult.Consumable() {
			panic("Cannot include a statement for the required variable result")
		}
		doExpr = ast.Block{
			Type:       "controls_do_then_return",
			Statements: ast.OptionalStatement("STM", doBody),
			// TODO: we have set the flag to false, previously was true, verify effects
			Values: []ast.Value{{Name: "VALUE", Block: doResult.Blockly(false)}},
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

// mutateVars returns a name list of declared variables, and the declarations are mutated to a set call.
// The variables will later be defined at the top.
func (s *SmartBody) mutateVars() []string {
	var names []string
	for k, expr := range s.Body {
		// We only have simple variables
		if e, ok := expr.(*variables.SimpleVar); ok {
			names = append(names, e.Name)
			// Mutate it to a set function
			s.Body[k] = &variables.Set{Global: false, Name: e.Name, Expr: e.Value}
		}
	}
	return names
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
