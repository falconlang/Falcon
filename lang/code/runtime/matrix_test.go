package runtime

import (
	"Falcon/code/context"
	"Falcon/code/lex"
	"Falcon/code/parsers/mistparser"
	"strings"
	"testing"
)

func TestMatrixCellRuntimeGet(t *testing.T) {
	got := evalMatrixTestSource(t, "[[1, 2], [3, 4]][[2, 1]]")
	if got.String() != "3" {
		t.Fatalf("matrix cell get = %q, want 3", got.String())
	}
}

func TestMatrixCellRuntimeSet(t *testing.T) {
	src := strings.Join([]string{
		"global m = [[1, 2], [3, 4]]",
		"m[[2, 1]] = 9",
		"m[[2, 1]]",
	}, "\n")

	got := evalMatrixTestSource(t, src)
	if got.String() != "9" {
		t.Fatalf("matrix cell set = %q, want 9", got.String())
	}
}

func evalMatrixTestSource(t *testing.T, src string) Value {
	t.Helper()
	ctx := &context.CodeContext{SourceCode: &src, FileName: "matrix_test.mist"}
	parser := mistparser.NewLangParser(false, lex.NewLexer(ctx).Lex())
	return NewInterpreter().RunGetLast(parser.ParseAll())
}
