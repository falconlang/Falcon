package mistparser

import "Falcon/code/ast"

//go:generate stringer -type=FrameType
type FrameType int

const (
	FrameTypeIf FrameType = iota
	FrameTypeLoop
	FrameTypeYield
	FrameTypeNoYield
)

type Frame struct {
	FrameType FrameType
	Expr      ast.Expr
}

func AppendFrame(frames []Frame, ft FrameType, expr ast.Expr) []Frame {
	var newFrames []Frame
	newFrames = append(newFrames, frames...)
	newFrames = append(newFrames, Frame{FrameType: ft, Expr: expr})
	return newFrames
}
