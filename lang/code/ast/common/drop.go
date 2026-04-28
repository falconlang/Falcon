package common

import "strings"

// dropTypeWords lists the type-name suffixes (and direct function names) that
// constitute a no-op type-conversion in Falcon's dynamic type system.
// ".toString()", ".toStr()", ".toInt()", ".toNumber()" etc. are all identity
// operations that should be silently dropped from the AST.
var dropTypeWords = map[string]bool{
	"string": true, "str": true, "text": true,
	"int": true, "integer": true,
	"num": true, "number": true, "float": true, "double": true,
	"bool": true, "boolean": true,
	"list": true, "array": true, "arr": true,
	"dict": true, "map": true, "object": true, "obj": true,
}

// IsDropMethod reports whether a 0-arg method call named name is a no-op
// type conversion that should be silently dropped.
// Matches the "to<Type>" camelCase prefix pattern (e.g. toString, toStr,
// toInt, toNumber, toList, toDict, toBoolean, …).
func IsDropMethod(name string) bool {
	lower := strings.ToLower(name)
	if !strings.HasPrefix(lower, "to") || len(lower) <= 2 {
		return false
	}
	return dropTypeWords[lower[2:]]
}

// IsDropFunction reports whether a 1-arg function call named name is a no-op
// type-cast wrapper that should be dropped, keeping only its argument.
// Matches both bare type names (string, int, number, …) and the "to<Type>"
// prefix form (toString, toInt, …).
func IsDropFunction(name string) bool {
	lower := strings.ToLower(name)
	if dropTypeWords[lower] {
		return true
	}
	if strings.HasPrefix(lower, "to") && len(lower) > 2 {
		return dropTypeWords[lower[2:]]
	}
	return false
}
