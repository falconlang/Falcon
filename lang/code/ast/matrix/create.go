package matrix

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/lex"
	"strconv"
	"strings"
)

type Create struct {
	Where *lex.Token
	Rows  [][]ast.Expr
}

func (c *Create) String() string {
	rows := make([]string, len(c.Rows))
	for i, row := range c.Rows {
		rows[i] = "[" + ast.JoinExprs(", ", row) + "]"
	}
	return "matrix[" + strings.Join(rows, ", ") + "]"
}

func (c *Create) Blockly(flags ...bool) ast.Block {
	rows, cols := c.dimensions()
	fields := make([]ast.Field, 0, 2+rows*cols)
	fields = append(fields,
		ast.Field{Name: "ROWS", Value: strconv.Itoa(rows)},
		ast.Field{Name: "COLS", Value: strconv.Itoa(cols)},
	)
	for i, row := range c.Rows {
		for j, cell := range row {
			value, ok := matrixLiteralCell(cell)
			if !ok {
				c.Where.TypeError("Matrix literal cells must be numeric literals")
			}
			fields = append(fields, ast.Field{
				Name:  "MATRIX_" + strconv.Itoa(i) + "_" + strconv.Itoa(j),
				Value: value,
			})
		}
	}
	return ast.Block{
		Type:     "matrices_create",
		Mutation: &ast.Mutation{Rows: rows, Cols: cols, Matrix: c.matrixMutation()},
		Fields:   fields,
	}
}

func (c *Create) Continuous() bool {
	return true
}

func (c *Create) Consumable() bool {
	return true
}

func (c *Create) Signature() []ast.Signature {
	c.dimensions()
	for _, row := range c.Rows {
		for _, cell := range row {
			if _, ok := matrixLiteralCell(cell); !ok {
				c.Where.TypeError("Matrix literal cells must be numeric literals")
			}
			cell.Signature()
		}
	}
	return []ast.Signature{ast.SignList}
}

func (c *Create) dimensions() (int, int) {
	rows := len(c.Rows)
	if rows == 0 {
		c.Where.TypeError("Matrix literal requires at least one row")
	}
	cols := len(c.Rows[0])
	if cols == 0 {
		c.Where.TypeError("Matrix literal requires at least one column")
	}
	for i, row := range c.Rows {
		if len(row) != cols {
			c.Where.TypeError("Matrix literal row % must have % columns, got %", strconv.Itoa(i+1), strconv.Itoa(cols), strconv.Itoa(len(row)))
		}
	}
	return rows, cols
}

func (c *Create) matrixMutation() string {
	rows := make([]string, len(c.Rows))
	for i, row := range c.Rows {
		cells := make([]string, len(row))
		for j, cell := range row {
			value, ok := matrixLiteralCell(cell)
			if !ok {
				c.Where.TypeError("Matrix literal cells must be numeric literals")
			}
			cells[j] = value
		}
		rows[i] = "[" + strings.Join(cells, ",") + "]"
	}
	return "[" + strings.Join(rows, ",") + "]"
}

func matrixLiteralCell(cell ast.Expr) (string, bool) {
	if number, ok := cell.(*fundamentals.Number); ok {
		return number.Content, true
	}
	if neg, ok := cell.(*common.FuncCall); ok && neg.Name == "neg" && len(neg.Args) == 1 {
		if number, ok := neg.Args[0].(*fundamentals.Number); ok {
			return "-" + number.Content, true
		}
	}
	return "", false
}
