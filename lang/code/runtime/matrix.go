package runtime

import (
	astmatrix "Falcon/code/ast/matrix"
	"Falcon/code/lex"
	"math"
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
	return MatrixVal(rows)
}

func (i *Interpreter) evalMatrixCreateND(dimsValue Value, initial Value) Value {
	if initial.Type() != Number {
		panic("matrix initial value must be a number, got " + initial.errorStr())
	}
	dimValues := *dimsValue.AsList()
	if len(dimValues) == 0 {
		panic("matrix dimensions must contain at least one dimension")
	}
	dims := make([]int, len(dimValues))
	for idx, dim := range dimValues {
		n := coerceIndex(dim, "matrix dimension "+strconv.Itoa(idx+1))
		if n < 1 {
			panic("matrix dimension " + strconv.Itoa(idx+1) + " must be positive, got " + strconv.Itoa(n))
		}
		dims[idx] = n
	}
	return matrixFromNestedList(buildMatrixND(dims, initial))
}

func buildMatrixND(dims []int, initial Value) Value {
	if len(dims) == 0 {
		return initial
	}
	values := make([]Value, dims[0])
	for idx := range values {
		values[idx] = buildMatrixND(dims[1:], initial)
	}
	return ListVal(values)
}

func matrixFromNestedList(v Value) Value {
	if v.Type() == Matrix {
		return v
	}
	return MatrixVal(*v.AsList())
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
	value := i.Eval(e.Value)
	if value.Type() != Number {
		panic("matrix cell assignment requires a number, got " + value.errorStr())
	}
	(*list)[idx-1] = value
	return VoidVal()
}

func (i *Interpreter) evalMatrixBinary(operator lex.Type, vals []Value) Value {
	if len(vals) == 0 {
		panic("matrix operation requires at least one operand")
	}
	switch operator {
	case lex.MatrixPlus:
		result := vals[0]
		for _, val := range vals[1:] {
			result = matrixElementwise(result, val, "add", func(a, b float64) float64 { return a + b })
		}
		return result
	case lex.MatrixDash:
		if len(vals) != 2 {
			panic("matrix subtraction requires exactly two operands")
		}
		return matrixElementwise(vals[0], vals[1], "subtract", func(a, b float64) float64 { return a - b })
	case lex.MatrixTimes:
		result := vals[0]
		for _, val := range vals[1:] {
			result = matrixMultiply(result, val)
		}
		return result
	case lex.MatrixPower:
		if len(vals) != 2 {
			panic("matrix power requires exactly two operands")
		}
		return matrixPower(vals[0], vals[1])
	default:
		panic("unknown matrix operator")
	}
}

func (i *Interpreter) evalMatrixMethod(name string, on Value, args []Value) Value {
	switch name {
	case "row":
		return matrixRow(on, args[0])
	case "col":
		return matrixColumn(on, args[0])
	case "dimension":
		dims, ok := matrixDimensions(on)
		if !ok {
			panic("matrix dimensions require a rectangular numeric matrix, got " + on.errorStr())
		}
		values := make([]Value, len(dims))
		for idx, dim := range dims {
			values[idx] = NumVal(float64(dim))
		}
		return ListVal(values)
	case "inverse":
		return matrixInverse(on)
	case "transpose":
		return matrixTranspose(on)
	case "rotateLeft":
		return matrixRotateLeft(on)
	case "rotateRight":
		return matrixRotateRight(on)
	default:
		panic("unknown matrix method ." + name + "()")
	}
}

func isMatrixValue(v Value) bool {
	return v.Type() == Matrix
}

func matrixDimensions(v Value) ([]int, bool) {
	if v.Type() != List && v.Type() != Matrix {
		return nil, false
	}
	list := *v.AsList()
	if len(list) == 0 {
		return nil, false
	}

	first := list[0]
	if first.Type() == Number {
		for _, elem := range list[1:] {
			if elem.Type() != Number {
				return nil, false
			}
		}
		return []int{len(list)}, true
	}
	if first.Type() != List {
		return nil, false
	}

	childDims, ok := matrixDimensions(first)
	if !ok {
		return nil, false
	}
	for _, elem := range list[1:] {
		dims, ok := matrixDimensions(elem)
		if !ok || !sameDims(dims, childDims) {
			return nil, false
		}
	}
	return append([]int{len(list)}, childDims...), true
}

func sameDims(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for idx := range a {
		if a[idx] != b[idx] {
			return false
		}
	}
	return true
}

func matrixElementwise(a, b Value, opName string, op func(float64, float64) float64) Value {
	aDims, aOK := matrixDimensions(a)
	bDims, bOK := matrixDimensions(b)
	if !aOK || !bOK {
		panic("matrix " + opName + " requires rectangular numeric matrices")
	}
	if !sameDims(aDims, bDims) {
		panic("matrix " + opName + " requires matching dimensions, got " + formatDims(aDims) + " and " + formatDims(bDims))
	}
	return matrixFromNestedList(matrixElementwiseSameShape(a, b, op))
}

func matrixElementwiseSameShape(a, b Value, op func(float64, float64) float64) Value {
	if a.Type() == Number && b.Type() == Number {
		return NumVal(op(a.AsNum(), b.AsNum()))
	}
	aList := *a.AsList()
	bList := *b.AsList()
	values := make([]Value, len(aList))
	for idx := range aList {
		values[idx] = matrixElementwiseSameShape(aList[idx], bList[idx], op)
	}
	return ListVal(values)
}

func matrixMultiply(a, b Value) Value {
	if b.Type() == Number {
		return matrixScalarMultiply(a, b.AsNum())
	}
	if a.Type() == Number {
		return matrixScalarMultiply(b, a.AsNum())
	}
	return matrixMatMul(a, b)
}

func matrixScalarMultiply(matrix Value, scalar float64) Value {
	if matrix.Type() == Number {
		return NumVal(matrix.AsNum() * scalar)
	}
	if _, ok := matrixDimensions(matrix); !ok {
		panic("matrix scalar multiplication requires a rectangular numeric matrix")
	}
	values := *matrix.AsList()
	result := make([]Value, len(values))
	for idx, elem := range values {
		result[idx] = matrixScalarMultiply(elem, scalar)
	}
	if matrix.Type() == Matrix {
		return MatrixVal(result)
	}
	return ListVal(result)
}

func matrixMatMul(a, b Value) Value {
	aNums, aRows, aCols := matrix2DNumbers(a, "matrix multiply left operand")
	bNums, bRows, bCols := matrix2DNumbers(b, "matrix multiply right operand")
	if aCols != bRows {
		panic("matrix multiply requires left columns to equal right rows, got " + strconv.Itoa(aCols) + " and " + strconv.Itoa(bRows))
	}
	result := make([]Value, aRows)
	for row := 0; row < aRows; row++ {
		cells := make([]Value, bCols)
		for col := 0; col < bCols; col++ {
			sum := 0.0
			for k := 0; k < aCols; k++ {
				sum += aNums[row][k] * bNums[k][col]
			}
			cells[col] = NumVal(sum)
		}
		result[row] = ListVal(cells)
	}
	return MatrixVal(result)
}

func matrixPower(matrix Value, exponent Value) Value {
	exp := coerceIndex(exponent, "matrix power")
	if exp < 0 {
		panic("matrix power exponent must be non-negative, got " + strconv.Itoa(exp))
	}
	_, rows, cols := matrix2DNumbers(matrix, "matrix power base")
	if rows != cols {
		panic("matrix power requires a square matrix, got " + strconv.Itoa(rows) + "x" + strconv.Itoa(cols))
	}
	result := matrixIdentity(rows)
	base := matrix
	for exp > 0 {
		if exp%2 == 1 {
			result = matrixMatMul(result, base)
		}
		exp /= 2
		if exp > 0 {
			base = matrixMatMul(base, base)
		}
	}
	return result
}

func matrixIdentity(size int) Value {
	rows := make([]Value, size)
	for row := 0; row < size; row++ {
		cells := make([]Value, size)
		for col := 0; col < size; col++ {
			if row == col {
				cells[col] = NumVal(1)
			} else {
				cells[col] = NumVal(0)
			}
		}
		rows[row] = ListVal(cells)
	}
	return MatrixVal(rows)
}

func matrixRow(matrix Value, rowValue Value) Value {
	values, rows, _ := matrix2DNumbers(matrix, "matrix row")
	row := coerceIndex(rowValue, "matrix row")
	if row < 1 || row > rows {
		panic("matrix row " + strconv.Itoa(row) + " out of bounds (rows=" + strconv.Itoa(rows) + ")")
	}
	cells := make([]Value, len(values[row-1]))
	for idx, n := range values[row-1] {
		cells[idx] = NumVal(n)
	}
	return ListVal(cells)
}

func matrixColumn(matrix Value, colValue Value) Value {
	values, rows, cols := matrix2DNumbers(matrix, "matrix column")
	col := coerceIndex(colValue, "matrix column")
	if col < 1 || col > cols {
		panic("matrix column " + strconv.Itoa(col) + " out of bounds (cols=" + strconv.Itoa(cols) + ")")
	}
	cells := make([]Value, rows)
	for row := 0; row < rows; row++ {
		cells[row] = NumVal(values[row][col-1])
	}
	return ListVal(cells)
}

func matrixTranspose(matrix Value) Value {
	values, rows, cols := matrix2DNumbers(matrix, "matrix transpose")
	result := make([]Value, cols)
	for col := 0; col < cols; col++ {
		cells := make([]Value, rows)
		for row := 0; row < rows; row++ {
			cells[row] = NumVal(values[row][col])
		}
		result[col] = ListVal(cells)
	}
	return MatrixVal(result)
}

func matrixRotateLeft(matrix Value) Value {
	values, rows, cols := matrix2DNumbers(matrix, "matrix rotateLeft")
	result := make([]Value, cols)
	for col := cols - 1; col >= 0; col-- {
		cells := make([]Value, rows)
		for row := 0; row < rows; row++ {
			cells[row] = NumVal(values[row][col])
		}
		result[cols-1-col] = ListVal(cells)
	}
	return MatrixVal(result)
}

func matrixRotateRight(matrix Value) Value {
	values, rows, cols := matrix2DNumbers(matrix, "matrix rotateRight")
	result := make([]Value, cols)
	for col := 0; col < cols; col++ {
		cells := make([]Value, rows)
		for row := rows - 1; row >= 0; row-- {
			cells[rows-1-row] = NumVal(values[row][col])
		}
		result[col] = ListVal(cells)
	}
	return MatrixVal(result)
}

func matrixInverse(matrix Value) Value {
	values, rows, cols := matrix2DNumbers(matrix, "matrix inverse")
	if rows != cols {
		panic("matrix inverse requires a square matrix, got " + strconv.Itoa(rows) + "x" + strconv.Itoa(cols))
	}
	n := rows
	aug := make([][]float64, n)
	for row := 0; row < n; row++ {
		aug[row] = make([]float64, 2*n)
		copy(aug[row], values[row])
		aug[row][n+row] = 1
	}

	for col := 0; col < n; col++ {
		pivot := col
		for row := col + 1; row < n; row++ {
			if math.Abs(aug[row][col]) > math.Abs(aug[pivot][col]) {
				pivot = row
			}
		}
		if math.Abs(aug[pivot][col]) < 1e-12 {
			panic("matrix inverse requires a non-singular matrix")
		}
		aug[col], aug[pivot] = aug[pivot], aug[col]

		pivotValue := aug[col][col]
		for j := 0; j < 2*n; j++ {
			aug[col][j] /= pivotValue
		}
		for row := 0; row < n; row++ {
			if row == col {
				continue
			}
			factor := aug[row][col]
			for j := 0; j < 2*n; j++ {
				aug[row][j] -= factor * aug[col][j]
			}
		}
	}

	result := make([]Value, n)
	for row := 0; row < n; row++ {
		cells := make([]Value, n)
		for col := 0; col < n; col++ {
			cells[col] = NumVal(cleanMatrixFloat(aug[row][n+col]))
		}
		result[row] = ListVal(cells)
	}
	return MatrixVal(result)
}

func matrix2DNumbers(matrix Value, context string) ([][]float64, int, int) {
	dims, ok := matrixDimensions(matrix)
	if !ok || len(dims) != 2 {
		if ok {
			panic(context + " requires a 2D matrix, got dimensions " + formatDims(dims))
		}
		panic(context + " requires a rectangular numeric matrix, got " + matrix.errorStr())
	}
	rows, cols := dims[0], dims[1]
	values := make([][]float64, rows)
	rowValues := *matrix.AsList()
	for row := 0; row < rows; row++ {
		values[row] = make([]float64, cols)
		colValues := *rowValues[row].AsList()
		for col := 0; col < cols; col++ {
			values[row][col] = colValues[col].AsNum()
		}
	}
	return values, rows, cols
}

func cleanMatrixFloat(n float64) float64 {
	const epsilon = 1e-9
	if math.Abs(n) < epsilon {
		return 0
	}
	scaled := math.Round(n*1e12) / 1e12
	if math.Abs(n-scaled) < epsilon {
		return scaled
	}
	return n
}

func formatDims(dims []int) string {
	if len(dims) == 0 {
		return "[]"
	}
	out := "["
	for idx, dim := range dims {
		if idx > 0 {
			out += "x"
		}
		out += strconv.Itoa(dim)
	}
	return out + "]"
}
