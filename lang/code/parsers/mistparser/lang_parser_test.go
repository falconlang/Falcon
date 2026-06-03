package mistparser

import (
	"Falcon/code/ast"
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

func parseOneForTest(t *testing.T, src string) ast.Expr {
	t.Helper()
	ctx := &context.CodeContext{SourceCode: &src, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())
	exprs := parser.ParseAll()
	if len(exprs) != 1 {
		t.Fatalf("ParseAll() expressions = %d, want 1", len(exprs))
	}
	return exprs[0]
}

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

func serializeBlocklyForTest(t *testing.T, src string) string {
	t.Helper()
	ctx := &context.CodeContext{SourceCode: &src, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())
	exprs := parser.ParseAll()
	blocks := make([]ast.Block, len(exprs))
	for i, expr := range exprs {
		blocks[i] = expr.Blockly(true)
	}
	return string(ast.XmlRoot{
		XMLNS:  "https://developers.google.com/blockly/xml",
		Blocks: blocks,
	}.MarshalIndent("", "  "))
}

func TestBlockCommentInlineSerializesRootBlockInline(t *testing.T) {
	xmlText := serializeBlocklyForTest(t, "/*inline*/ 1 + 2")

	if !strings.Contains(xmlText, `<block type="math_add" inline="true">`) {
		t.Fatalf("serialized XML = %s, want root math_add inline", xmlText)
	}
}

func TestBlockCommentInlineInsideExpressionSerializesMarkedOperandOnly(t *testing.T) {
	xmlText := serializeBlocklyForTest(t, "1 + /*inline*/ 2")

	if strings.Contains(xmlText, `<block type="math_add" inline="true">`) {
		t.Fatalf("serialized XML = %s, root math_add should not be inline", xmlText)
	}
	if !strings.Contains(xmlText, `<block type="math_number" inline="true">`) {
		t.Fatalf("serialized XML = %s, want marked operand inline", xmlText)
	}
}

func TestBlockCommentSerializesAsBlocklyComment(t *testing.T) {
	xmlText := serializeBlocklyForTest(t, "/* explain */ 123")

	want := `<comment pinned="false" h="80" w="160">explain</comment>`
	if !strings.Contains(xmlText, want) {
		t.Fatalf("serialized XML = %s, want %s", xmlText, want)
	}
}

func TestBlockCommentInlineAndCommentCanAnnotateSameBlock(t *testing.T) {
	xmlText := serializeBlocklyForTest(t, "/*inline*/ /* explain */ 123")

	if !strings.Contains(xmlText, `<block type="math_number" inline="true">`) {
		t.Fatalf("serialized XML = %s, want inline number block", xmlText)
	}
	want := `<comment pinned="false" h="80" w="160">explain</comment>`
	if !strings.Contains(xmlText, want) {
		t.Fatalf("serialized XML = %s, want %s", xmlText, want)
	}
}

func TestMultipleBlockCommentsJoinAsBlocklyComment(t *testing.T) {
	xmlText := serializeBlocklyForTest(t, "/* first */ /* second */ 123")

	want := "<comment pinned=\"false\" h=\"80\" w=\"160\">first\nsecond</comment>"
	if !strings.Contains(xmlText, want) {
		t.Fatalf("serialized XML = %s, want joined comment", xmlText)
	}
}

func TestAnonymousProcedureDefinitionsGenerateProcedureBlocks(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "void",
			src:  `func(x) { println(x) }`,
			want: "procedures_defanonnoreturn",
		},
		{
			name: "returning",
			src:  `func(r) = { 3.14 * r ^ 2 }`,
			want: "procedures_defanonreturn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOneForTest(t, tt.src).Blockly(false)
			if got.Type != tt.want {
				t.Fatalf("Blockly type = %q, want %q", got.Type, tt.want)
			}
			if len(got.Mutation.Args) != 1 || got.Mutation.Args[0].Name == "" {
				t.Fatalf("mutation args = %#v, want one named argument", got.Mutation.Args)
			}
		})
	}
}

func TestEscapedKeywordIdentifiersParseAndRoundTrip(t *testing.T) {
	src := strings.Join([]string{
		"local `when` = 4",
		"`when` + 2",
	}, "\n")
	expr := parseOneForTest(t, src)
	if got := expr.String(); got != src {
		t.Fatalf("String() = %q, want %q", got, src)
	}
}

func TestEscapedKeywordProcedureNamesAndParametersParse(t *testing.T) {
	src := strings.Join([]string{
		"func `when`(`global`) = `global` + 1",
		"`when`(4)",
	}, "\n")
	ctx := &context.CodeContext{SourceCode: &src, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())

	exprs := parser.ParseAll()
	if len(exprs) != 2 {
		t.Fatalf("ParseAll() expressions = %d, want 2", len(exprs))
	}
	if got := exprs[0].String(); got != "func `when`(`global`) =\n  `global` + 1\n" {
		t.Fatalf("procedure String() = %q", got)
	}
	if got := exprs[1].String(); got != "`when`(4)" {
		t.Fatalf("call String() = %q", got)
	}
}

func TestAnonymousProcedureCallsChooseBlockByConsumption(t *testing.T) {
	statementCall := parseOneForTest(t, `(func(x) { println(x) })("Melon")`).Blockly(true)
	if statementCall.Type != "procedures_callanonnoreturn" {
		t.Fatalf("statement call block = %q, want procedures_callanonnoreturn", statementCall.Type)
	}
	if statementCall.Mutation.ItemCount != 1 {
		t.Fatalf("statement call item count = %d, want 1", statementCall.Mutation.ItemCount)
	}

	printCall := parseOneForTest(t, `println((func(x) = x)(5))`).(*common.FuncCall)
	valueBlock := printCall.Args[0].Blockly(false)
	if valueBlock.Type != "procedures_callanonreturn" {
		t.Fatalf("value call block = %q, want procedures_callanonreturn", valueBlock.Type)
	}
}

func TestAnonymousProcedureInputListCallsChooseBlockByConsumption(t *testing.T) {
	statementCall := parseOneForTest(t, `(func(x) { println(x) }).call(["Melon"])`).Blockly(true)
	if statementCall.Type != "procedures_callanonnoreturn_inputlist" {
		t.Fatalf("statement input-list call block = %q, want procedures_callanonnoreturn_inputlist", statementCall.Type)
	}

	printCall := parseOneForTest(t, `println((func(x) = x).call([5]))`).(*common.FuncCall)
	valueBlock := printCall.Args[0].Blockly(false)
	if valueBlock.Type != "procedures_callanonreturn_inputlist" {
		t.Fatalf("value input-list call block = %q, want procedures_callanonreturn_inputlist", valueBlock.Type)
	}
}

func TestAnonymousProcedureHelperBlocks(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "num args",
			src:  `println((func(x) { println(x) }).numArgs())`,
			want: "procedures_numArgs",
		},
		{
			name: "get with name",
			src:  `println(getFunc("sayHello"))`,
			want: "procedures_getWithName",
		},
		{
			name: "get with dropdown",
			src:  `println(func.sayHello)`,
			want: "procedures_getWithDropdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printCall := parseOneForTest(t, tt.src).(*common.FuncCall)
			got := printCall.Args[0].Blockly(false)
			if got.Type != tt.want {
				t.Fatalf("helper block = %q, want %q", got.Type, tt.want)
			}
		})
	}
}

func TestMatrixCellSyntaxGeneratesMatrixBlocks(t *testing.T) {
	tests := []struct {
		name  string
		src   string
		block string
	}{
		{
			name:  "get",
			src:   "makeNdArray([2, 2], 0)[[1, 2]]",
			block: "matrices_get_cell",
		},
		{
			name:  "set",
			src:   "makeNdArray([2, 2], 0)[[1, 2]] = 9",
			block: "matrices_set_cell",
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

func TestParenthesizedListIndexStaysListIndex(t *testing.T) {
	src := "[[1, 2], [3, 4]][([1, 2])]"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())

	exprs := parser.ParseAll()
	if len(exprs) != 1 {
		t.Fatalf("ParseAll() expressions = %d, want 1", len(exprs))
	}
	if got := exprs[0].Blockly(false).Type; got != "lists_select_item" {
		t.Fatalf("Blockly type = %q, want lists_select_item", got)
	}
	if got := exprs[0].String(); got != src {
		t.Fatalf("source = %q, want %q", got, src)
	}
}

func TestMatrixCreateSyntaxGeneratesMatrixCreateBlock(t *testing.T) {
	src := "matrix[[1, 2], [3, 4]]"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())

	exprs := parser.ParseAll()
	if len(exprs) != 1 {
		t.Fatalf("ParseAll() expressions = %d, want 1", len(exprs))
	}
	block := exprs[0].Blockly(false)
	if block.Type != "matrices_create" {
		t.Fatalf("Blockly type = %q, want matrices_create", block.Type)
	}
	if got := exprs[0].String(); got != src {
		t.Fatalf("source = %q, want %q", got, src)
	}
}

func TestNestedListLiteralStaysListBlock(t *testing.T) {
	src := "[[1, 2], [3, 4]]"
	ctx := &context.CodeContext{SourceCode: &src, FileName: "test.mist"}
	parser := NewLangParser(false, lex.NewLexer(ctx).Lex())

	exprs := parser.ParseAll()
	if len(exprs) != 1 {
		t.Fatalf("ParseAll() expressions = %d, want 1", len(exprs))
	}
	if got := exprs[0].Blockly(false).Type; got != "lists_create_with" {
		t.Fatalf("Blockly type = %q, want lists_create_with", got)
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
