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
