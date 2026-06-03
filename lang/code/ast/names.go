package ast

import "strings"

var reservedNames = map[string]struct{}{
	"true":      {},
	"false":     {},
	"if":        {},
	"else":      {},
	"for":       {},
	"step":      {},
	"in":        {},
	"while":     {},
	"do":        {},
	"break":     {},
	"walkAll":   {},
	"global":    {},
	"local":     {},
	"this":      {},
	"func":      {},
	"when":      {},
	"any":       {},
	"undefined": {},
	"yield":     {},
}

func FormatName(name string) string {
	if isPlainName(name) {
		if _, reserved := reservedNames[name]; !reserved {
			return name
		}
	}
	return "`" + strings.ReplaceAll(name, "`", "\\`") + "`"
}

func JoinNames(sep string, names []string) string {
	formatted := make([]string, len(names))
	for i, name := range names {
		formatted[i] = FormatName(name)
	}
	return strings.Join(formatted, sep)
}

func isPlainName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' || (i > 0 && c >= '0' && c <= '9') {
			continue
		}
		return false
	}
	return true
}
