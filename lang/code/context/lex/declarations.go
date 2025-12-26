package lex

var Symbols = map[string]StaticToken{
	"+": staticOf(Plus, Binary, Operator, Compoundable),
	"-": staticOf(Dash, Binary, Operator, Unary, Compoundable),
	"*": staticOf(Times, BinaryL1, Operator, PreserveOrder, Compoundable),
	"/": staticOf(Slash, BinaryL1, Operator, PreserveOrder, Compoundable),
	"%": staticOf(Remainder, BinaryL1, Operator, PreserveOrder, Compoundable),
	"^": staticOf(Power, BinaryL2, Operator, PreserveOrder, Compoundable),

	"||": staticOf(LogicOr, LLogicOr, Operator),
	"&&": staticOf(LogicAnd, LLogicAnd, Operator),
	"|":  staticOf(BitwiseOr, BBitwiseOr, Operator, Compoundable),
	"&":  staticOf(BitwiseAnd, BBitwiseAnd, Operator, Compoundable),
	"~":  staticOf(BitwiseXor, BBitwiseXor, Operator, Compoundable),

	"==":  staticOf(Equals, Equality, Operator),
	"!=":  staticOf(NotEquals, Equality, Operator),
	"===": staticOf(TextEquals, Equality, Operator),
	"!==": staticOf(TextNotEquals, Equality, Operator),

	"<":  staticOf(LessThan, Relational, Operator),
	"<=": staticOf(LessThanEqual, Relational, Operator),
	">":  staticOf(GreatThan, Relational, Operator),
	">=": staticOf(GreaterThanEqual, Relational, Operator),

	"<<": staticOf(TextLessThan, Relational, Operator),
	">>": staticOf(TextGreaterThan, Relational, Operator),

	"_":  staticOf(Underscore, TextJoin, Operator, Compoundable),
	"@":  staticOf(At),
	":":  staticOf(Colon, Pair, Operator),
	"::": staticOf(DoubleColon),
	"..": staticOf(DoubleDot),

	"(": staticOf(OpenCurve),
	")": staticOf(CloseCurve),
	"[": staticOf(OpenSquare),
	"]": staticOf(CloseSquare),
	"{": staticOf(OpenCurly),
	"}": staticOf(CloseCurly),

	"=":  staticOf(Assign, AssignmentType, Operator),
	".":  staticOf(Dot),
	",":  staticOf(Comma),
	"?":  staticOf(Question),
	"!":  staticOf(Not),
	"->": staticOf(RightArrow),
}

var Keywords = map[string]StaticToken{
	"true":  staticOf(True, Value, ConstantValue),
	"false": staticOf(False, Value, ConstantValue),

	"if":        staticOf(If),
	"else":      staticOf(Else),
	"for":       staticOf(For),
	"step":      staticOf(Step),
	"in":        staticOf(In),
	"while":     staticOf(While),
	"do":        staticOf(Do),
	"break":     staticOf(Break),
	"walkAll":   staticOf(WalkAll),
	"global":    staticOf(Global),
	"local":     staticOf(Local),
	"compute":   staticOf(Compute),
	"this":      staticOf(This, Value),
	"func":      staticOf(Func),
	"when":      staticOf(When),
	"any":       staticOf(Any),
	"undefined": staticOf(Undefined),
}
