package mistparser

import (
	"Falcon/code/context"
	"Falcon/code/lex"
	"reflect"
	"strings"
	"testing"
)

func TestParseAllWithLineNumbersTracksTopLevelExpressionStarts(t *testing.T) {
	source := strings.Join([]string{
		"@Button { Button1 }",
		"",
		"true",
		"if(false) {",
		"  true",
		"}",
		"123",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())

	exprs, lineNumbers := parser.ParseTopLevel()

	if len(exprs) != 3 {
		t.Fatalf("ParseTopLevel() expressions = %d, want 3", len(exprs))
	}
	want := []int{3, 4, 7}
	if !reflect.DeepEqual(lineNumbers, want) {
		t.Fatalf("ParseTopLevel() line numbers = %v, want %v", lineNumbers, want)
	}
}

func TestKnownFunctionWrongArityReportsArgumentCount(t *testing.T) {
	source := "println()"
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want arity diagnostic")
		}
		err, ok := r.(*context.DiagnosticListError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticListError", r, r)
		}
		want := "Expected 1 args but got 0 for function println()"
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
		if len(err.Diagnostics) != 1 || err.Diagnostics[0].Message != want {
			t.Fatalf("diagnostics = %#v, want message %q", err.Diagnostics, want)
		}
	}()

	parser.ParseTopLevel()
}

func TestKnownMethodWrongArityReportsArgumentCount(t *testing.T) {
	source := `"abc".replace()`
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want arity diagnostic")
		}
		err, ok := r.(*context.DiagnosticListError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticListError", r, r)
		}
		want := ".replace(from, to) expects 2 arg(s) but got 0"
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
		if len(err.Diagnostics) != 1 || err.Diagnostics[0].Message != want {
			t.Fatalf("diagnostics = %#v, want message %q", err.Diagnostics, want)
		}
	}()

	parser.ParseTopLevel()
}
