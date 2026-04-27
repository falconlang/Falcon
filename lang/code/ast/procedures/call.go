package procedures

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
	"Falcon/code/sugar"
)

type Call struct {
	Where      *lex.Token
	Name       string
	Parameters []string
	Arguments  []ast.Expr
	Returning  bool
}

func (v *Call) String() string {
	return sugar.Format("%(%)", v.Name, ast.JoinExprs(", ", v.Arguments))
}

func (v *Call) Blockly(flags ...bool) ast.Block {
	var blockType string
	if v.Returning {
		blockType = "procedures_callreturn"
	} else {
		blockType = "procedures_callnoreturn"
	}
	return ast.Block{
		Type:     blockType,
		Mutation: &ast.Mutation{Name: v.Name, Args: ast.ToArgs(v.Parameters)},
		Fields:   []ast.Field{{Name: "PROCNAME", Value: v.Name}},
		Values:   ast.ValuesByPrefix("ARG", v.Arguments),
	}
}

func (v *Call) Continuous() bool {
	return true
}

func (v *Call) Consumable() bool {
	return v.Returning
}

func (v *Call) Signature() []ast.Signature {
	for _, arg := range v.Arguments {
		arg.Signature()
	}
	return []ast.Signature{ast.SignAny}
}
