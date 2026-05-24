package method

import (
	"Falcon/code/ast"
)

func (c *Call) textMethods(signature *CallSignature) ast.Block {
	switch signature.BlocklyName {
	case "text_length":
		return c.simpleOperand("text_length", "VALUE")
	case "text_trim":
		return c.simpleOperand("text_trim", "TEXT")
	case "text_split_at_spaces":
		return c.simpleOperand("text_split_at_spaces", "TEXT")
	case "text_reverse":
		return c.simpleOperand("text_reverse", "VALUE")
	case "lists_from_csv_row":
		return c.simpleOperand("lists_from_csv_row", "TEXT")
	case "lists_from_csv_table":
		return c.simpleOperand("lists_from_csv_table", "TEXT")
	case "text_changeCase":
		return c.textChangeCase()
	case "text_starts_at":
		return c.textStartsAt()
	case "text_contains":
		return c.textContains()
	case "text_split":
		return c.textSplit()
	case "text_segment":
		return c.textSegment()
	case "text_replace_all":
		return c.textReplace()
	case "text_replace_mappings":
		return c.textReplaceFrom()
	default:
		c.Where.Error("Unknown text method: %", signature.BlocklyName)
		panic("")
	}
}

func (c *Call) textChangeCase() ast.Block {
	var fieldOp string
	if c.Name == "uppercase" {
		fieldOp = "UPCASE"
	} else {
		fieldOp = "DOWNCASE"
	}
	return ast.Block{
		Type:   "text_changeCase",
		Fields: []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: []ast.Value{{Name: "TEXT", Block: c.On.Blockly(false)}},
	}
}

func (c *Call) textReplaceFrom() ast.Block {
	var fieldOp string
	if c.Name == "replaceFrom" {
		fieldOp = "DICTIONARY_ORDER"
	} else {
		fieldOp = "LONGEST_STRING_FIRST"
	}
	return ast.Block{
		Type:   "text_replace_mappings",
		Fields: []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: []ast.Value{
			{Name: "MAPPINGS", Block: c.Args[0].Blockly(false)},
			{Name: "TEXT", Block: c.On.Blockly(false)},
		},
	}
}

func (c *Call) textReplace() ast.Block {
	return ast.Block{
		Type: "text_replace_all",
		Values: []ast.Value{
			{Name: "TEXT", Block: c.On.Blockly(false)},
			{Name: "SEGMENT", Block: c.Args[0].Blockly(false)},
			{Name: "REPLACEMENT", Block: c.Args[1].Blockly(false)},
		},
	}
}

func (c *Call) textSegment() ast.Block {
	return ast.Block{
		Type: "text_segment",
		Values: []ast.Value{
			{Name: "TEXT", Block: c.On.Blockly(false)},
			{Name: "START", Block: c.Args[0].Blockly(false)},
			{Name: "LENGTH", Block: c.Args[1].Blockly(false)},
		},
	}
}

func (c *Call) textSplit() ast.Block {
	var fieldOp string
	switch c.Name {
	case "split":
		fieldOp = "SPLIT"
	case "splitAtFirst":
		fieldOp = "SPLITATFIRST"
	case "splitAtAny":
		fieldOp = "SPLITATANY"
	case "splitAtFirstOfAny":
		fieldOp = "SPLITATFIRSTOFANY"
	}
	return ast.Block{
		Type:     "text_split",
		Mutation: &ast.Mutation{Mode: fieldOp},
		Fields:   []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: []ast.Value{
			{Name: "TEXT", Block: c.On.Blockly(false)},
			{Name: "AT", Block: c.Args[0].Blockly(false)},
		},
	}
}

func (c *Call) textContains() ast.Block {
	var fieldOp string
	switch c.Name {
	case "contains":
		fieldOp = "CONTAINS"
	case "containsAny":
		fieldOp = "CONTAINS_ANY"
	case "containsAll":
		fieldOp = "CONTAINS_ALL"
	}
	return ast.Block{
		Type:     "text_contains",
		Mutation: &ast.Mutation{Mode: fieldOp},
		Fields:   []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: []ast.Value{
			{Name: "TEXT", Block: c.On.Blockly(false)},
			{Name: "PIECE", Block: c.Args[0].Blockly(false)},
		},
	}
}

func (c *Call) textStartsAt() ast.Block {
	return ast.Block{
		Type: "text_starts_at",
		Values: []ast.Value{
			{Name: "TEXT", Block: c.On.Blockly(false)},
			{Name: "PIECE", Block: c.Args[0].Blockly(false)},
		},
	}
}
