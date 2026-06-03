package lex

import (
	"Falcon/code/context"
	"testing"
)

func TestEscapedIdentifierLexesAsName(t *testing.T) {
	src := "`when` `global` `has space` `has\\`tick`"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "lexer_test.mist"}

	tokens := NewLexer(ctx).Lex()

	want := []string{"when", "global", "has space", "has`tick"}
	if len(tokens) != len(want) {
		t.Fatalf("tokens = %d, want %d", len(tokens), len(want))
	}
	for i, token := range tokens {
		if token.Type != Name {
			t.Fatalf("token[%d].Type = %s, want Name", i, token.Type)
		}
		if got := *token.Content; got != want[i] {
			t.Fatalf("token[%d].Content = %q, want %q", i, got, want[i])
		}
	}
}
