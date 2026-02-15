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
	default:
		panic("Unknown matrix method: " + signature.BlocklyName)
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
