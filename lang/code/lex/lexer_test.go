package lex

import (
	"Falcon/code/ast"
	"Falcon/code/context"
	"strings"
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

func TestBlockCommentLexesAsToken(t *testing.T) {
	src := "/* hello */ 1"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "lexer_test.mist"}

	tokens := NewLexer(ctx).Lex()

	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2", len(tokens))
	}
	if tokens[0].Type != BlockComment {
		t.Fatalf("token[0].Type = %s, want BlockComment", tokens[0].Type)
	}
	if got := *tokens[0].Content; got != " hello " {
		t.Fatalf("token[0].Content = %q, want %q", got, " hello ")
	}
	if tokens[1].Type != Number {
		t.Fatalf("token[1].Type = %s, want Number", tokens[1].Type)
	}
}

func TestBlockCommentCanSpanLines(t *testing.T) {
	src := "/* one\n two */ 3"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "lexer_test.mist"}

	tokens := NewLexer(ctx).Lex()

	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2", len(tokens))
	}
	if tokens[0].Type != BlockComment {
		t.Fatalf("token[0].Type = %s, want BlockComment", tokens[0].Type)
	}
	if got := *tokens[0].Content; got != " one\n two " {
		t.Fatalf("token[0].Content = %q, want multiline block comment", got)
	}
	if tokens[1].Column != 2 {
		t.Fatalf("token[1].Column = %d, want 2", tokens[1].Column)
	}
}

func TestUnterminatedBlockCommentReportsError(t *testing.T) {
	src := "/* missing"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "lexer_test.mist"}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("NewLexer().Lex() did not panic")
		}
		if !strings.Contains(r.(error).Error(), "Unterminated block comment") {
			t.Fatalf("panic = %v, want unterminated block comment", r)
		}
	}()

	NewLexer(ctx).Lex()
}
