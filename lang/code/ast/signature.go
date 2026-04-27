package ast

import "strings"

//go:generate stringer -type=Signature
type Signature int

const (
	SignBool Signature = iota
	SignNumb
	SignText
	SignList
	SignDict
	SignComponent
	SignHelper
	SignAny
	SignOfEvent
	SignVoid
)

func CombineSignatures(first []Signature, second []Signature) []Signature {
	seen := make(map[Signature]bool)
	unique := make([]Signature, 0, len(first)+len(second))

	for _, sig := range first {
		if !seen[sig] {
			seen[sig] = true
			unique = append(unique, sig)
		}
	}
	for _, sig := range second {
		if !seen[sig] {
			seen[sig] = true
			unique = append(unique, sig)
		}
	}
	return unique
}

// HasSignature reports whether signatures contain target or SignAny.
func HasSignature(signatures []Signature, target Signature) bool {
	for _, s := range signatures {
		if s == SignAny || s == target {
			return true
		}
	}
	return false
}

// FormatSignatures returns a human-readable string for a slice of signatures.
func FormatSignatures(signatures []Signature) string {
	if len(signatures) == 0 {
		return "unknown"
	}
	parts := make([]string, len(signatures))
	for i, s := range signatures {
		parts[i] = s.String()
	}
	return strings.Join(parts, " | ")
}
