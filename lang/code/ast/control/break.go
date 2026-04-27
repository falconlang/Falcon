package control

import (
	"Falcon/code/ast"
)

type Break struct {
	// Hola Amigo!
}

func (b *Break) String() string {
	return "break"
}

func (b *Break) Blockly(flags ...bool) ast.Block {
	return ast.Block{Type: "controls_break"}
}

func (b *Break) Continuous() bool {
	return true
}

func (b *Break) Consumable() bool {
	return false
}

func (b *Break) Signature() []ast.Signature {
	return []ast.Signature{ast.SignVoid}
}
