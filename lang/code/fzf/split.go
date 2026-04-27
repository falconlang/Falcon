package fzf

import "strings"

func SplitCamel(s string) []string {
	if s == "" {
		return nil
	}
	var tokens []string
	start := 0
	for i := 1; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			if s[i-1] >= 'a' && s[i-1] <= 'z' {
				tokens = append(tokens, strings.ToLower(s[start:i]))
				start = i
			} else if i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z' {
				tokens = append(tokens, strings.ToLower(s[start:i]))
				start = i
			}
		}
	}
	tokens = append(tokens, strings.ToLower(s[start:]))
	return tokens
}
