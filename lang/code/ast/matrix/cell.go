package matrix

import (
	"Falcon/code/ast"
	"Falcon/code/lex"
	"Falcon/code/sugar"
	"strconv"
)

type GetCell struct {
	Where  *lex.Token
	Matrix ast.Expr
	Dims   []ast.Expr
}

func (g *GetCell) String() string {
	pFormat := "%[[%]]"
	if !g.Matrix.Continuous() {
		pFormat = "(%)[[%]]"
	}
	return sugar.Format(pFormat, g.Matrix.String(), ast.JoinExprs(", ", g.Dims))
}

func (g *GetCell) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "matrices_get_cell",
		Mutation: &ast.Mutation{ItemCount: len(g.Dims)},
		Values:   matrixCellValues(g.Matrix, g.Dims, nil),
	}
}

func (g *GetCell) Continuous() bool {
	return true
}

func (g *GetCell) Consumable() bool {
	return true
}

func (g *GetCell) Signature() []ast.Signature {
	matrixSigs := g.Matrix.Signature()
	if !ast.HasSignature(matrixSigs, ast.SignList) {
		g.Where.TypeError("Matrix cell access requires a matrix value, but got %", ast.FormatSignatures(matrixSigs))
	}
	for _, dim := range g.Dims {
		dimSigs := dim.Signature()
		if !ast.HasSignature(dimSigs, ast.SignNumb) {
			g.Where.TypeError("Matrix cell dimension requires a number value, but got %", ast.FormatSignatures(dimSigs))
		}
	}
	return []ast.Signature{ast.SignAny}
}

type SetCell struct {
	Where  *lex.Token
	Matrix ast.Expr
	Dims   []ast.Expr
	Value  ast.Expr
}

func (s *SetCell) String() string {
	pFormat := "%[[%]] = %"
	if !s.Matrix.Continuous() {
		pFormat = "(%)[[%]] = %"
	}
	return sugar.Format(pFormat, s.Matrix.String(), ast.JoinExprs(", ", s.Dims), s.Value.String())
}

func (s *SetCell) Blockly(flags ...bool) ast.Block {
	return ast.Block{
		Type:     "matrices_set_cell",
		Mutation: &ast.Mutation{ItemCount: len(s.Dims)},
		Values:   matrixCellValues(s.Matrix, s.Dims, s.Value),
	}
}

func (s *SetCell) Continuous() bool {
	return false
}

func (s *SetCell) Consumable() bool {
	return false
}

func (s *SetCell) Signature() []ast.Signature {
	matrixSigs := s.Matrix.Signature()
	if !ast.HasSignature(matrixSigs, ast.SignList) {
		s.Where.TypeError("Matrix cell assignment requires a matrix value, but got %", ast.FormatSignatures(matrixSigs))
	}
	for _, dim := range s.Dims {
		dimSigs := dim.Signature()
		if !ast.HasSignature(dimSigs, ast.SignNumb) {
			s.Where.TypeError("Matrix cell dimension requires a number value, but got %", ast.FormatSignatures(dimSigs))
		}
	}
	s.Value.Signature()
	return []ast.Signature{ast.SignVoid}
}

func matrixCellValues(matrix ast.Expr, dims []ast.Expr, value ast.Expr) []ast.Value {
	values := make([]ast.Value, 0, len(dims)+2)
	values = append(values, ast.Value{Name: "MATRIX", Block: matrix.Blockly(false)})
	for i, dim := range dims {
		values = append(values, ast.Value{Name: "DIM" + strconv.Itoa(i), Block: dim.Blockly(false)})
	}
	if value != nil {
		values = append(values, ast.Value{Name: "VALUE", Block: value.Blockly(false)})
	}
	return values
}
