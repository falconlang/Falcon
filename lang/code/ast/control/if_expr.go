package control

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/sugar"
	"strings"
)

type SimpleIf struct {
	condition ast.Expr

	smartThen ast.Expr
	smartElse ast.Expr

	normalThen []ast.Expr
	normalElse []ast.Expr
}

func (s *SimpleIf) Condition() ast.Expr    { return s.condition }
func (s *SimpleIf) Then() []ast.Expr       { return s.normalThen }
func (s *SimpleIf) Else() []ast.Expr       { return s.normalElse }

func MakeSimpleIf(condition ast.Expr, then []ast.Expr, elze []ast.Expr) *SimpleIf {
	return &SimpleIf{
		condition:  condition,
		smartThen:  &fundamentals.SmartBody{Body: then},
		smartElse:  &fundamentals.SmartBody{Body: elze},
		normalThen: then,
		normalElse: elze,
	}
}

func (s *SimpleIf) String() string {
	var branches []string
	currIf := s
	var hasDiscontinuity = false

	for {
		if !currIf.normalThen[0].Continuous() || !currIf.normalElse[0].Continuous() {
			hasDiscontinuity = true
		}
		// append If branch
		ifFormat := ""
		if currIf != s {
			// we are in a nested if
			ifFormat += "else "
		}
		var thenString string
		if len(currIf.normalThen) == 1 {
			thenString = currIf.normalThen[0].String()
			if strings.ContainsRune(thenString, '\n') || strings.HasPrefix(thenString, "{") {
				ifFormat += "if (%) {\n%} "
				thenString = ast.PadBody(currIf.normalThen)
			} else {
				ifFormat += "if (%) % "
			}
		} else {
			ifFormat += "if (%) {\n%} "
			thenString = ast.PadBody(currIf.normalThen)
		}
		branches = append(branches, sugar.Format(ifFormat, currIf.condition.String(), thenString))
		// check for nested If branch
		nextIf, hasNextIf := currIf.normalElse[0].(*SimpleIf)
		if len(currIf.normalElse) == 1 && hasNextIf {
			// break it, let it be handled in the next iteration
			currIf = nextIf
			continue
		}
		// append Else branch
		var elseFormat string
		var elseString string
		if len(currIf.normalElse) == 1 {
			elseString = currIf.normalElse[0].String()
			if strings.ContainsRune(elseString, '\n') || strings.HasPrefix(elseString, "{") {
				elseFormat = "else {\n%}"
				elseString = ast.PadBody(currIf.normalElse)
			} else {
				elseFormat = "else %"
			}
		} else {
			elseFormat = "else {\n%}"
			elseString = ast.PadBody(currIf.normalElse)
		}
		branches = append(branches, sugar.Format(elseFormat, elseString))
		if !hasNextIf {
			break
		}
	}
	if len(branches) > 2 || hasDiscontinuity {
		return strings.Join(branches, "\n")
	}
	return strings.Join(branches, "")
}

func (s *SimpleIf) Blockly(flags ...bool) ast.Block {
	// SimpleIf is always an expression (controls_choose).
	// It should never appear in a statement position.
	// if len(flags) > 0 && flags[0] {
	// 	// Blockly expects a statement here, we have to mutate
	// 	fullIf := If{
	// 		Conditions: []ast.Expr{s.condition},
	// 		Bodies:     [][]ast.Expr{s.normalThen},
	// 		ElseBody:   s.normalElse,
	// 	}
	// 	return fullIf.Blockly()
	// }
	return ast.Block{
		Type:   "controls_choose",
		Values: ast.MakeValues([]ast.Expr{s.condition, s.smartThen, s.smartElse}, "TEST", "THENRETURN", "ELSERETURN"),
	}
}

func (s *SimpleIf) Continuous() bool {
	return false
}

func (s *SimpleIf) Consumable(flags ...bool) bool {
	return true
}

func (s *SimpleIf) Signature() []ast.Signature {
	return ast.CombineSignatures(s.smartThen.Signature(), s.smartElse.Signature())
}
