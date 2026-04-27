package fundamentals

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
)

type Color struct {
	Where *lex.Token
	Hex   string
}

func (c *Color) String() string {
	return c.Hex
}

func (c *Color) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "color_black",
		Fields: []ast.Field{{Name: "COLOR", Value: c.Hex}},
	}
}

func (c *Color) Continuous() bool {
	return true
}

func (c *Color) Consumable() bool {
	return true
}

func (c *Color) Signature() []ast.Signature {
	return []ast.Signature{ast.SignNumb}
}
