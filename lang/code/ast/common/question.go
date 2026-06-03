package common

import (
	"Falcon/code/ast"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/fzf"
	"Falcon/code/lex"
	"Falcon/code/sugar"
)

var questionKeywords = []string{
	"number", "base10", "hexa", "bin",
	"text", "list", "dict",
	"emptyText", "emptyList",
	"even", "odd",
}

// IsKnownQuestion reports whether name is a valid question keyword.
func IsKnownQuestion(name string) bool {
	for _, k := range questionKeywords {
		if k == name {
			return true
		}
	}
	return false
}

// FindBestQuestionSuggestion returns the closest question keyword for a method-call-style name.
// It strips a leading "isX" camelCase prefix (e.g. "isNumber" → "number") before fuzzy matching.
// Returns "" when no candidate clears the fzf threshold.
func FindBestQuestionSuggestion(methodName string) string {
	base := stripIsPrefix(methodName)
	if IsKnownQuestion(base) {
		return base
	}
	if tops := fzf.Top(base, questionKeywords, 1); len(tops) > 0 {
		return tops[0]
	}
	if base != methodName {
		if tops := fzf.Top(methodName, questionKeywords, 1); len(tops) > 0 {
			return tops[0]
		}
	}
	return ""
}

// stripIsPrefix removes a leading "is" + uppercase letter prefix from a camelCase name
// and lowercases the first character of the remainder (e.g. "isNumber" → "number").
func stripIsPrefix(name string) string {
	if len(name) > 2 && name[0] == 'i' && name[1] == 's' {
		rest := name[2:]
		if len(rest) > 0 && rest[0] >= 'A' && rest[0] <= 'Z' {
			return string(rest[0]+'a'-'A') + rest[1:]
		}
	}
	return name
}

type Question struct {
	Where            *lex.Token
	On               ast.Expr
	Question         string
	MethodCallSyntax bool // true when parsed from .name() or .isName() notation
}

func (q *Question) String() string {
	pFormat := "% ? %"
	if !q.On.Continuous() {
		pFormat = "(%) ? %"
	}
	return sugar.Format(pFormat, q.On.String(), ast.FormatName(q.Question))
}

func (q *Question) Blockly(flags ...bool) ast.Block {
	switch q.Question {
	case "number", "base10", "hexa", "bin":
		return q.mathQuestion()
	case "text":
		return q.textQuestion()
	case "list":
		return q.listQuestion()
	case "dict":
		return q.dictQuestion()
	case "matrix":
		return q.matrixQuestion()
	case "emptyText":
		return q.textIsEmpty()
	case "emptyList":
		return q.listIsEmpty()
	case "even", "odd":
		return q.evenOrOdd()
	default:
		q.Where.Error("Unknown question ? %", q.Question)
	}
	panic("Unreachable")
}

func (q *Question) Continuous() bool {
	return false
}

func (q *Question) Consumable() bool {
	return true
}

func (q *Question) Signature() []ast.Signature {
	q.On.Signature()
	return []ast.Signature{ast.SignBool}
}

func (q *Question) evenOrOdd() ast.Block {
	var remainder string
	if q.Question == "even" {
		remainder = "0"
	} else {
		remainder = "1"
	}
	remainderCall := &FuncCall{
		Where: q.Where,
		Name:  "rem",
		Args:  []ast.Expr{q.On, &fundamentals.Number{Content: "2"}},
	}
	comparison := BinaryExpr{
		Where:    q.Where,
		Operands: []ast.Expr{remainderCall, &fundamentals.Number{Content: remainder}},
		Operator: lex.Equals,
	}
	return comparison.Blockly(false)
}

func (q *Question) listIsEmpty() ast.Block {
	return ast.Block{
		Type:   "lists_is_empty",
		Values: []ast.Value{{Name: "LIST", Block: q.On.Blockly(false)}},
	}
}

func (q *Question) textIsEmpty() ast.Block {
	return ast.Block{
		Type:   "text_isEmpty",
		Values: []ast.Value{{Name: "VALUE", Block: q.On.Blockly(false)}},
	}
}

func (q *Question) matrixQuestion() ast.Block {
	return ast.Block{
		Type:   "matrices_is_matrix",
		Values: []ast.Value{{Name: "VALUE", Block: q.On.Blockly(false)}},
	}
}

func (q *Question) dictQuestion() ast.Block {
	return ast.Block{
		Type:   "dictionaries_is_dict",
		Values: []ast.Value{{Name: "THING", Block: q.On.Blockly(false)}},
	}
}

func (q *Question) listQuestion() ast.Block {
	return ast.Block{
		Type:   "lists_is_list",
		Values: []ast.Value{{Name: "ITEM", Block: q.On.Blockly(false)}},
	}
}

func (q *Question) textQuestion() ast.Block {
	return ast.Block{
		Type:   "text_is_string",
		Values: []ast.Value{{Name: "ITEM", Block: q.On.Blockly(false)}},
	}
}

func (q *Question) mathQuestion() ast.Block {
	var fieldOp string
	switch q.Question {
	case "number":
		fieldOp = "NUMBER"
	case "base10":
		fieldOp = "BASE10"
	case "hexa":
		fieldOp = "HEXADECIMAL"
	case "bin":
		fieldOp = "BINARY"
	}
	return ast.Block{
		Type:   "math_is_a_number",
		Fields: []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: []ast.Value{{Name: "NUM", Block: q.On.Blockly(false)}},
	}
}
