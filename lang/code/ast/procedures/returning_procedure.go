package procedures

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/variables"
	"Falcon/code/sugar"
)

type RetProcedure struct {
	Name       string
	Parameters []string
	Result     ast.Expr
}

func (v *RetProcedure) String() string {
	var resultString string
	switch v.Result.(type) {
	case *variables.VarResult:
		resultString = ast.Pad("{\n" + ast.Pad(v.Result.String()) + "}")
	default:
		if sb, ok := v.Result.(*fundamentals.SmartBody); ok && len(sb.Body) == 1 {
			switch sb.Body[0].(type) {
			case *variables.VarResult, *variables.Var:
				resultString = ast.Pad("{\n" + ast.Pad(sb.Body[0].String()) + "}")
				break
			}
		}
		if resultString == "" {
			resultString = ast.Pad(v.Result.String())
		}
	}
	return sugar.Format("func %(%) =\n%", ast.FormatName(v.Name), ast.JoinNames(", ", v.Parameters), resultString)
}

func (v *RetProcedure) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "procedures_defreturn",
		Mutation: &ast.Mutation{Args: ast.ToArgs(v.Parameters)},
		Fields:   append(ast.ToFields("VAR", v.Parameters), ast.Field{Name: "NAME", Value: v.Name}),
		Values:   []ast.Value{{Name: "RETURN", Block: v.Result.Blockly(false)}},
	}
}

func (v *RetProcedure) Continuous() bool {
	return false
}

func (v *RetProcedure) Consumable() bool {
	return false
}

func (v *RetProcedure) Signature() []ast.Signature {
	return v.Result.Signature()
}
