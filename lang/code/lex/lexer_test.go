package lex

import (
	"Falcon/code/ast"
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

func TestFormattedEscapedIdentifierLexesBackToOriginalName(t *testing.T) {
	want := `a\b` + "`" + `c`
	src := ast.FormatName(want)
	ctx := &context.CodeContext{SourceCode: &src, FileName: "lexer_test.mist"}

	tokens := NewLexer(ctx).Lex()

	if len(tokens) != 1 {
		t.Fatalf("tokens = %d, want 1", len(tokens))
	}
	if tokens[0].Type != Name {
		t.Fatalf("token type = %s, want Name", tokens[0].Type)
	}
	if got := *tokens[0].Content; got != want {
		t.Fatalf("token content = %q, want %q; source was %q", got, want, src)
	}
}
