package runtime

// BreakSignal is panicked by a break statement and recovered by loops.
type BreakSignal struct{}

// ReturnSignal is panicked by a returning procedure and recovered by the caller.
type ReturnSignal struct {
	Val Value
}

// YieldSignal is panicked by a yield statement and recovered by the enclosing
// function call, which returns the carried value to the caller.
type YieldSignal struct {
	Val Value
}
