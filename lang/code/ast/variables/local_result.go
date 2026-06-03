package variables

import (
	"Falcon/code/ast"
	"strings"
)

type VarResult struct {
	Names  []string
	Values []ast.Expr
	Result ast.Expr
}

func (v *VarResult) String() string {
	var result ast.Expr
	var combinedNames []string
	var combinedValues []ast.Expr

	result = v.Result
	combinedNames = v.Names
	combinedValues = v.Values

	for {
		// check for nested var results!
		if vr, ok := result.(*VarResult); ok {
			combinedNames = append(combinedNames, vr.Names...)
			combinedValues = append(combinedValues, vr.Values...)
			result = vr.Result
		} else {
			break
		}
	}

	var builder strings.Builder
	localLines := make([]string, len(combinedNames))
	for k, name := range combinedNames {
		localLines[k] = "local " + ast.FormatName(name) + " = " + combinedValues[k].String()
	}
	builder.WriteString(strings.Join(localLines, "\n"))
	builder.WriteString("\n")
	builder.WriteString(result.String())
	return builder.String()
}

func (v *VarResult) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "local_declaration_expression",
		Mutation: &ast.Mutation{LocalNames: ast.MakeLocalNames(v.Names...)},
		Fields:   ast.ToFields("VAR", v.Names),
		Values: append(ast.ValuesByPrefix("DECL", v.Values),
			ast.Value{Name: "RETURN", Block: v.Result.Blockly(false)}),
	}
}

func (v *VarResult) Continuous() bool {
	return true
}

func (v *VarResult) Consumable() bool {
	return true
}

func (v *VarResult) Signature() []ast.Signature {
	for _, value := range v.Values {
		value.Signature()
	}
	return v.Result.Signature()
}
