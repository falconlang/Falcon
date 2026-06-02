package runtime

import (
	astmatrix "Falcon/code/ast/matrix"
	"strconv"
)

func (i *Interpreter) evalMatrixCreate(e *astmatrix.Create) Value {
	rows := make([]Value, len(e.Rows))
	for rowIndex, row := range e.Rows {
		cells := make([]Value, len(row))
		for colIndex, cell := range row {
			value := i.Eval(cell)
			if value.Type() != Number {
				panic("matrix cell " + strconv.Itoa(rowIndex+1) + "," + strconv.Itoa(colIndex+1) + " must be a number, got " + value.errorStr())
			}
			cells[colIndex] = value
		}
		rows[rowIndex] = ListVal(cells)
	}
	return ListVal(rows)
}

func (i *Interpreter) evalMatrixGetCell(e *astmatrix.GetCell) Value {
	curr := i.Eval(e.Matrix)
	for dimIndex, dim := range e.Dims {
		list := curr.AsList()
		idx := coerceIndex(i.Eval(dim), "matrix dimension "+strconv.Itoa(dimIndex+1))
		if idx < 1 || idx > len(*list) {
			panic("matrix dimension " + strconv.Itoa(dimIndex+1) + " index " + strconv.Itoa(idx) + " out of bounds (len=" + strconv.Itoa(len(*list)) + ")")
		}
		curr = (*list)[idx-1]
	}
	return curr
}

func (i *Interpreter) evalMatrixSetCell(e *astmatrix.SetCell) Value {
	if len(e.Dims) == 0 {
		panic("matrix cell assignment requires at least one dimension")
	}

	curr := i.Eval(e.Matrix)
	for dimIndex, dim := range e.Dims[:len(e.Dims)-1] {
		list := curr.AsList()
		idx := coerceIndex(i.Eval(dim), "matrix dimension "+strconv.Itoa(dimIndex+1))
		if idx < 1 || idx > len(*list) {
			panic("matrix dimension " + strconv.Itoa(dimIndex+1) + " index " + strconv.Itoa(idx) + " out of bounds (len=" + strconv.Itoa(len(*list)) + ")")
		}
		curr = (*list)[idx-1]
	}

	list := curr.AsList()
	lastDimIndex := len(e.Dims)
	idx := coerceIndex(i.Eval(e.Dims[lastDimIndex-1]), "matrix dimension "+strconv.Itoa(lastDimIndex))
	if idx < 1 || idx > len(*list) {
		panic("matrix dimension " + strconv.Itoa(lastDimIndex) + " index " + strconv.Itoa(idx) + " out of bounds (len=" + strconv.Itoa(len(*list)) + ")")
	}
	(*list)[idx-1] = i.Eval(e.Value)
	return VoidVal()
}
