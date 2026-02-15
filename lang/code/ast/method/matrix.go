package method

import "Falcon/code/ast"

func (c *Call) matrixMethods(signature *CallSignature) ast.Block {
	switch signature.BlocklyName {
	case "matrices_get_row":
		return c.matrixGetRow()
	case "matrices_get_column":
		return c.matrixGetColumn()
	case "matrices_get_dims":
		return c.matrixGetDimensions()
	case "matrices_operations":
		return c.matrixOperations()
	default:
		panic("Unknown matrix method: " + signature.BlocklyName)
	}
}

func (c *Call) matrixOperations() ast.Block {
	var blocklyOp string
	switch c.Name {
	case "inverse":
		blocklyOp = "INVERSE"
	case "transpose":
		blocklyOp = "TRANSPOSE"
	case "rotateLeft":
		blocklyOp = "ROTATE_LEFT"
	case "rotateRight":
		blocklyOp = "ROTATE_RIGHT"
	default:
		panic("Unknown matrix operation: " + c.Name)
	}
	return ast.Block{
		Type:   "matrices_operations",
		Fields: []ast.Field{{Name: "OP", Value: blocklyOp}},
		Values: []ast.Value{{Name: "MATRIX", Block: c.On.Blockly()}},
	}
}

func (c *Call) matrixGetDimensions() ast.Block {
	return ast.Block{
		Type:   "matrices_get_dims",
		Values: []ast.Value{{Name: "MATRIX", Block: c.On.Blockly()}},
	}
}

func (c *Call) matrixGetRow() ast.Block {
	return ast.Block{
		Type:   "matrices_get_row",
		Values: ast.MakeValueArgs(c.On, "MATRIX", c.Args, "ROW"),
	}
}

func (c *Call) matrixGetColumn() ast.Block {
	return ast.Block{
		Type:   "matrices_get_column",
		Values: ast.MakeValueArgs(c.On, "MATRIX", c.Args, "COLUMN"),
	}
}
