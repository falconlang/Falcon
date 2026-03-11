package runtime

// BreakSignal is panicked by a break statement and recovered by loops.
type BreakSignal struct{}

// ReturnSignal is panicked by a returning procedure and recovered by the caller.
type ReturnSignal struct {
	Val Value
}
