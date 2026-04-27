package common

import (
	"Falcon/code/ast"
)

type EmptySocket struct{}

func (e *EmptySocket) String() string {
	return "undefined"
}

func (e *EmptySocket) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:   "math_number",
		Fields: ast.FieldsFromMap(map[string]string{"NUM": "0"}),
	}
}

func (e *EmptySocket) Continuous() bool {
	return true
}

func (e *EmptySocket) Consumable() bool {
	return false
}

func (e *EmptySocket) Signature() []ast.Signature {
	return []ast.Signature{ast.SignNumb}
}
