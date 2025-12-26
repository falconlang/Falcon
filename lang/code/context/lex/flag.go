package lex

//go:generate stringer -type=Flag
type Flag int

const (
	Operator Flag = iota
	LLogicOr
	LLogicAnd
	BBitwiseOr
	BBitwiseAnd
	BBitwiseXor

	Relational
	Equality
	Binary
	BinaryL1
	BinaryL2
	TextJoin
	Pair
	AssignmentType
	Unary

	Value
	ConstantValue

	PreserveOrder
	Compoundable
)

func PrecedenceOf(flag Flag) int {
	switch flag {
	case AssignmentType:
		return 0
	case Pair:
		return 1
	case TextJoin:
		return 2
	case LLogicOr:
		return 3
	case LLogicAnd:
		return 4
	case BBitwiseOr:
		return 5
	case BBitwiseAnd:
		return 6
	case BBitwiseXor:
		return 7
	case Equality:
		return 8
	case Relational:
		return 9
	case Binary:
		return 10
	case BinaryL1:
		return 11
	case BinaryL2:
		return 12
	default:
		return -1
	}
}
