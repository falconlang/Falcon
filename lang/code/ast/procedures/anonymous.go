package procedures

import (
	"Falcon/code/ast"
	"Falcon/code/ast/control"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/variables"
	"Falcon/code/sugar"
	"strings"
)

type AnonProcedure struct {
	Parameters []string
	Body       []ast.Expr
	Result     ast.Expr
	Returning  bool
}

func (a *AnonProcedure) String() string {
	if a.Returning {
		return sugar.Format("func(%) =\n%", strings.Join(a.Parameters, ", "), formatProcedureResult(a.Result))
	}
	return sugar.Format("func(%) {\n%}", strings.Join(a.Parameters, ", "), ast.PadBody(a.Body))
}

func (a *AnonProcedure) Blockly(flags ...bool) ast.Block {
	block := ast.Block{
		Type:     "procedures_defanonnoreturn",
		Mutation: &ast.Mutation{Args: ast.ToArgs(a.Parameters)},
		Fields:   ast.ToFields("VAR", a.Parameters),
	}
	if a.Returning {
		block.Type = "procedures_defanonreturn"
		block.Values = []ast.Value{{Name: "RETURN", Block: a.Result.Blockly(false)}}
	} else {
		block.Statements = ast.OptionalStatement("STACK", a.Body)
	}
	return block
}

func (a *AnonProcedure) Continuous() bool {
	return true
}

func (a *AnonProcedure) Consumable() bool {
	return true
}

func (a *AnonProcedure) Signature() []ast.Signature {
	if a.Returning {
		a.Result.Signature()
	} else {
		for _, expr := range a.Body {
			expr.Signature()
		}
	}
	return []ast.Signature{ast.SignAny}
}

type AnonCall struct {
	Procedure ast.Expr
	Arguments []ast.Expr
}

func (a *AnonCall) String() string {
	return sugar.Format("%(%)", callTargetString(a.Procedure), ast.JoinExprs(", ", a.Arguments))
}

func (a *AnonCall) Blockly(flags ...bool) ast.Block {
	blockType := "procedures_callanonreturn"
	if len(flags) > 0 && flags[0] {
		blockType = "procedures_callanonnoreturn"
	}
	return ast.Block{
		Type:     blockType,
		Mutation: &ast.Mutation{ItemCount: len(a.Arguments)},
		Values:   ast.ValueArgsByPrefix(a.Procedure, "PROCEDURE", "ARG", a.Arguments),
	}
}

func (a *AnonCall) Continuous() bool {
	return true
}

func (a *AnonCall) Consumable() bool {
	return false
}

func (a *AnonCall) Signature() []ast.Signature {
	a.Procedure.Signature()
	for _, arg := range a.Arguments {
		arg.Signature()
	}
	return []ast.Signature{ast.SignAny}
}

type AnonCallInputList struct {
	Procedure ast.Expr
	InputList ast.Expr
}

func (a *AnonCallInputList) String() string {
	return sugar.Format("%.call(%)", callTargetString(a.Procedure), a.InputList.String())
}

func (a *AnonCallInputList) Blockly(flags ...bool) ast.Block {
	blockType := "procedures_callanonreturn_inputlist"
	if len(flags) > 0 && flags[0] {
		blockType = "procedures_callanonnoreturn_inputlist"
	}
	return ast.Block{
		Type:   blockType,
		Values: ast.MakeValueArgs(a.Procedure, "PROCEDURE", []ast.Expr{a.InputList}, "INPUTLIST"),
	}
}

func (a *AnonCallInputList) Continuous() bool {
	return true
}

func (a *AnonCallInputList) Consumable() bool {
	return false
}

func (a *AnonCallInputList) Signature() []ast.Signature {
	a.Procedure.Signature()
	a.InputList.Signature()
	return []ast.Signature{ast.SignAny}
}

type NumArgs struct {
	Procedure ast.Expr
}

func (n *NumArgs) String() string {
	return sugar.Format("%.numArgs()", callTargetString(n.Procedure))
}

func (n *NumArgs) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "procedures_numArgs",
		Values: []ast.Value{{Name: "PROCEDURE", Block: n.Procedure.Blockly(false)}},
	}
}

func (n *NumArgs) Continuous() bool {
	return true
}

func (n *NumArgs) Consumable() bool {
	return true
}

func (n *NumArgs) Signature() []ast.Signature {
	n.Procedure.Signature()
	return []ast.Signature{ast.SignNumb}
}

type GetWithName struct {
	Name ast.Expr
}

func (g *GetWithName) String() string {
	return sugar.Format("getFunc(%)", g.Name.String())
}

func (g *GetWithName) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "procedures_getWithName",
		Values: []ast.Value{{Name: "PROCEDURENAME", Block: g.Name.Blockly(false)}},
	}
}

func (g *GetWithName) Continuous() bool {
	return true
}

func (g *GetWithName) Consumable() bool {
	return true
}

func (g *GetWithName) Signature() []ast.Signature {
	g.Name.Signature()
	return []ast.Signature{ast.SignAny}
}

type GetWithDropdown struct {
	Name string
}

func (g *GetWithDropdown) String() string {
	return "func." + g.Name
}

func (g *GetWithDropdown) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "procedures_getWithDropdown",
		Fields: []ast.Field{{Name: "PROCNAME", Value: g.Name}},
	}
}

func (g *GetWithDropdown) Continuous() bool {
	return true
}

func (g *GetWithDropdown) Consumable() bool {
	return true
}

func (g *GetWithDropdown) Signature() []ast.Signature {
	return []ast.Signature{ast.SignAny}
}

func formatProcedureResult(result ast.Expr) string {
	var resultString string
	switch result.(type) {
	case *control.Do, *variables.VarResult:
		resultString = ast.Pad("{\n" + ast.Pad(result.String()) + "}")
	default:
		if sb, ok := result.(*fundamentals.SmartBody); ok && len(sb.Body) == 1 {
			switch sb.Body[0].(type) {
			case *variables.VarResult, *variables.Var:
				resultString = ast.Pad("{\n" + ast.Pad(sb.Body[0].String()) + "}")
			}
		}
		if resultString == "" {
			resultString = ast.Pad(result.String())
		}
	}
	return resultString
}

func callTargetString(expr ast.Expr) string {
	if expr.Continuous() {
		return expr.String()
	}
	return "(" + expr.String() + ")"
}
