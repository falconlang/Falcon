package mistparser

import (
	"Falcon/code/ast/common"
	"Falcon/code/ast/procedures"
	"Falcon/code/ast/variables"
	"Falcon/code/compdb"
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

func TestMatrixOperatorsGenerateMatrixBlocks(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		block string
	}{
		{
			name:  "add",
			src:   "makeNdArray([1, 2], 0) [+] makeNdArray([1, 2], 1)",
			block: "matrices_add",
		},
		{
			name:  "subtract",
			src:   "makeNdArray([1, 2], 0) [-] makeNdArray([1, 2], 1)",
			block: "matrices_subtract",
		},
		{
			name:  "multiply",
			src:   "makeNdArray([1, 2], 0) [*] makeNdArray([1, 2], 1)",
			block: "matrices_multiply",
		},
		{
			name:  "power",
			src:   "makeNdArray([2, 2], 0) [^] 2",
			block: "matrices_power",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.CodeContext{SourceCode: &tt.src, FileName: "test.mist"}
			parser := NewLangParser(false, lex.NewLexer(ctx).Lex())

			exprs := parser.ParseAll()
			if len(exprs) != 1 {
				t.Fatalf("ParseAll() expressions = %d, want 1", len(exprs))
			}
			got := exprs[0].Blockly(false).Type
			if got != tt.block {
				t.Fatalf("Blockly type = %q, want %q", got, tt.block)
			}
		})
	}
}

func TestNumericPowerStillGeneratesMathPower(t *testing.T) {
	src := "2 ^ 3"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())

	exprs := parser.ParseAll()
	if len(exprs) != 1 {
		t.Fatalf("ParseAll() expressions = %d, want 1", len(exprs))
	}
	got := exprs[0].Blockly(false).Type
	if got != "math_power" {
		t.Fatalf("Blockly type = %q, want math_power", got)
	}
}

func TestScreenInitializeEventValidatesAgainstFormMetadata(t *testing.T) {
	source := strings.Join([]string{
		"@Screen { Screen1 }",
		"",
		"when Screen1.Initialize() {",
		"  true",
		"}",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())
	parser.SetComponentDefinitions(
		map[string][]string{"Screen": {"Screen1"}},
		map[string]string{"Screen1": "Screen"},
	)
	parser.SetEventValidator(compdb.GlobalDB.ValidateEvent)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ParseTopLevel() rejected Screen.Initialize: %v", r)
		}
	}()

	exprs, lineNumbers := parser.ParseTopLevel()
	if len(exprs) != 1 {
		t.Fatalf("ParseTopLevel() expressions = %d, want 1", len(exprs))
	}
	want := []int{3}
	if !reflect.DeepEqual(lineNumbers, want) {
		t.Fatalf("ParseTopLevel() line numbers = %v, want %v", lineNumbers, want)
	}
}

func TestInvalidComponentEventReportsDiagnostic(t *testing.T) {
	source := strings.Join([]string{
		"@Button { Button1 }",
		"",
		"when Button1.NotAnEvent() {",
		"  println(1)",
		"}",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())
	parser.SetComponentDefinitions(
		map[string][]string{"Button": {"Button1"}},
		map[string]string{"Button1": "Button"},
	)
	parser.SetEventValidator(compdb.GlobalDB.ValidateEvent)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want invalid event diagnostic")
		}
		err, ok := r.(*context.DiagnosticError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticError", r, r)
		}
		want := "component Button has no event NotAnEvent"
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
	}()

	parser.ParseTopLevel()
}

func TestComponentEventParameterNameMismatchReportsDiagnostic(t *testing.T) {
	source := strings.Join([]string{
		"@Screen { Screen1 }",
		"",
		"when Screen1.PermissionGranted(permission) {",
		"  println(permission)",
		"}",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())
	parser.SetComponentDefinitions(
		map[string][]string{"Screen": {"Screen1"}},
		map[string]string{"Screen1": "Screen"},
	)
	parser.SetEventValidator(compdb.GlobalDB.ValidateEvent)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want event parameter diagnostic")
		}
		err, ok := r.(*context.DiagnosticError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticError", r, r)
		}
		want := `event Form.PermissionGranted parameter 1 must be "permissionName", got "permission"`
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
	}()

	parser.ParseTopLevel()
}

func TestInvalidComponentPropertyReportsDiagnostic(t *testing.T) {
	source := strings.Join([]string{
		"@Button { Button1 }",
		"Button1.NotAProperty = 3",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())
	parser.SetComponentDefinitions(
		map[string][]string{"Button": {"Button1"}},
		map[string]string{"Button1": "Button"},
	)
	parser.SetPropertyValidator(compdb.GlobalDB.ValidateProperty)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want invalid property diagnostic")
		}
		err, ok := r.(*context.DiagnosticError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticError", r, r)
		}
		want := `Button: unknown property "NotAProperty"`
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
	}()

	parser.ParseTopLevel()
}

func TestInvalidComponentMethodReportsDiagnostic(t *testing.T) {
	source := strings.Join([]string{
		"@Web { Web1 }",
		"Web1.NotAMethod()",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())
	parser.SetComponentDefinitions(
		map[string][]string{"Web": {"Web1"}},
		map[string]string{"Web1": "Web"},
	)
	parser.SetMethodValidator(compdb.GlobalDB.ValidateMethod)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want invalid method diagnostic")
		}
		err, ok := r.(*context.DiagnosticError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticError", r, r)
		}
		want := `Web: unknown method "NotAMethod"`
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
	}()

	parser.ParseTopLevel()
}

func TestComponentMethodWrongArityReportsDiagnostic(t *testing.T) {
	source := strings.Join([]string{
		"@Web { Web1 }",
		"Web1.Get(1)",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())
	parser.SetComponentDefinitions(
		map[string][]string{"Web": {"Web1"}},
		map[string]string{"Web1": "Web"},
	)
	parser.SetMethodValidator(compdb.GlobalDB.ValidateMethod)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want method arity diagnostic")
		}
		err, ok := r.(*context.DiagnosticError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticError", r, r)
		}
		want := "method Web.Get takes no arguments"
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
	}()

	parser.ParseTopLevel()
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

func TestForwardProcedureReferenceParsesBeforeDefinition(t *testing.T) {
	source := strings.Join([]string{
		"println(addOne(41))",
		"",
		"func addOne(n) = n + 1",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())

	exprs, _ := parser.ParseTopLevel()
	printCall, ok := exprs[0].(*common.FuncCall)
	if !ok {
		t.Fatalf("first expression = %T, want *common.FuncCall", exprs[0])
	}
	procCall, ok := printCall.Args[0].(*procedures.Call)
	if !ok {
		t.Fatalf("println argument = %T, want *procedures.Call", printCall.Args[0])
	}
	if procCall.Name != "addOne" || !procCall.Returning || !reflect.DeepEqual(procCall.Parameters, []string{"n"}) {
		t.Fatalf("procedure call = %#v, want returning addOne(n)", procCall)
	}
}

func TestForwardGlobalReferenceParsesBeforeDefinition(t *testing.T) {
	source := strings.Join([]string{
		"println(this.message)",
		"",
		"global message = \"hello\"",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())

	exprs, _ := parser.ParseTopLevel()
	printCall, ok := exprs[0].(*common.FuncCall)
	if !ok {
		t.Fatalf("first expression = %T, want *common.FuncCall", exprs[0])
	}
	get, ok := printCall.Args[0].(*variables.Get)
	if !ok {
		t.Fatalf("println argument = %T, want *variables.Get", printCall.Args[0])
	}
	if !get.Global || get.Name != "message" {
		t.Fatalf("global get = %#v, want this.message", get)
	}
}

func TestUndefinedGlobalStillReportsDiagnostic(t *testing.T) {
	source := "println(this.missing)"
	ctx := &context.CodeContext{SourceCode: &source, FileName: "test.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("ParseTopLevel() panic = nil, want missing global diagnostic")
		}
		err, ok := r.(*context.DiagnosticListError)
		if !ok {
			t.Fatalf("ParseTopLevel() panic = %T %v, want *context.DiagnosticListError", r, r)
		}
		want := "Cannot find symbol 'missing'"
		if !strings.Contains(err.Raw, want) {
			t.Fatalf("diagnostic raw = %q, want substring %q", err.Raw, want)
		}
	}()

	parser.ParseTopLevel()
}
