package ast

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
