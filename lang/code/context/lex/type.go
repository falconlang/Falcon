package lex

//go:generate stringer -type=Type
type Type int

const (
	Plus Type = iota
	Dash
	Times
	Slash
	Power
	Remainder

	LogicOr
	LogicAnd
	BitwiseOr
	BitwiseAnd
	BitwiseXor

	Equals
	NotEquals

	LessThan
	LessThanEqual
	GreatThan
	GreaterThanEqual

	TextEquals
	TextNotEquals

	TextLessThan
	TextGreaterThan

	OpenCurve
	CloseCurve
	OpenSquare
	CloseSquare
	OpenCurly
	CloseCurly

	Assign
	Dot
	Comma
	Question
	Not
	Colon
	DoubleColon
	DoubleDot
	RightArrow
	Underscore
	At

	True
	False
	Text
	Number
	Name
	ColorCode

	If
	Else
	For
	Step
	In
	While
	Do
	Break
	WalkAll
	Global
	Local
	Compute
	This
	Func
	When
	Any
	Undefined
)
