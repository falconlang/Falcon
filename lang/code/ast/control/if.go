package control

import (
	"Falcon/code/ast"
	"strings"
)

type If struct {
	Conditions []ast.Expr
	Bodies     [][]ast.Expr
	ElseBody   []ast.Expr

	mutated  bool
	mutation ast.Expr
}

func (i *If) String() string {
	// TODO: accept flags here too
	var builder strings.Builder

	numConditions := len(i.Conditions)
	currI := 0

	builder.WriteString("if ")
	for {
		builder.WriteString("(")
		builder.WriteString(i.Conditions[currI].String())
		builder.WriteString(") {\n")
		builder.WriteString(ast.PadBody(i.Bodies[currI]))
		builder.WriteString("}")
		currI++
		if currI < numConditions {
			builder.WriteString(" else if ")
		} else {
			break
		}
	}
	if i.ElseBody != nil {
		builder.WriteString(" else {\n")
		builder.WriteString(ast.PadBody(i.ElseBody))
		builder.WriteString("}")
	}
	return builder.String()
}

func (i *If) Blockly(flags ...bool) ast.Block {
	if len(flags) > 0 && !flags[0] {
		// Default to an if expression
		return i.createSimpleIf()
	}
	conditions := ast.ValuesByPrefix("IF", i.Conditions)
	bodies := ast.ToStatements("DO", i.Bodies)

	numbElifs := len(conditions) - 1
	var numbElse int

	if i.ElseBody != nil {
		bodies = append(bodies, ast.CreateStatement("ELSE", i.ElseBody))
		numbElse = 1
	} else {
		numbElse = 0
	}
	return ast.Block{
		Type:       "controls_if",
		Mutation:   &ast.Mutation{ElseIfCount: numbElifs, ElseCount: numbElse},
		Values:     conditions,
		Statements: bodies,
	}
}

func (i *If) createSimpleIf() ast.Block {
	var currElseBlock []ast.Expr
	if i.ElseBody != nil {
		currElseBlock = i.ElseBody
	}
	for k := len(i.Conditions) - 1; k >= 0; k-- {
		condition := i.Conditions[k]
		then := i.Bodies[k]
		simpleIf := MakeSimpleIf(condition, then, currElseBlock)
		currElseBlock = []ast.Expr{simpleIf}
	}
	i.mutated = true
	i.mutation = currElseBlock[0]
	return currElseBlock[0].Blockly()
}

func (i *If) Continuous() bool {
	return false
}

func (i *If) Consumable() bool {
	if i.mutated {
		return true
	}
	if i.ElseBody == nil || len(i.ElseBody) == 0 {
		return false
	}
	for _, body := range i.Bodies {
		if len(body) == 0 || !body[len(body)-1].Consumable() {
			return false
		}
	}
	if !i.ElseBody[len(i.ElseBody)-1].Consumable() {
		return false
	}
	return true
}

func (i *If) Signature() []ast.Signature {
	for _, cond := range i.Conditions {
		cond.Signature()
	}
	for _, body := range i.Bodies {
		for _, expr := range body {
			expr.Signature()
		}
	}
	for _, expr := range i.ElseBody {
		expr.Signature()
	}
	if i.ElseBody == nil || len(i.ElseBody) == 0 {
		return []ast.Signature{ast.SignVoid}
	}
	for _, body := range i.Bodies {
		if len(body) == 0 || !body[len(body)-1].Consumable() {
			return []ast.Signature{ast.SignVoid}
		}
	}
	if !i.ElseBody[len(i.ElseBody)-1].Consumable() {
		return []ast.Signature{ast.SignVoid}
	}
	var result []ast.Signature
	for _, body := range i.Bodies {
		result = ast.CombineSignatures(result, body[len(body)-1].Signature())
	}
	result = ast.CombineSignatures(result, i.ElseBody[len(i.ElseBody)-1].Signature())
	return result
}
