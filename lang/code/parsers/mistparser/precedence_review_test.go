package mistparser

import (
	"Falcon/code/ast"
	"Falcon/code/ast/common"
	"Falcon/code/ast/fundamentals"
	"Falcon/code/ast/procedures"
	"Falcon/code/ast/variables"
	"Falcon/code/context"
	"Falcon/code/lex"
	blocklytomist "Falcon/code/parsers/blocklytomist"
	"Falcon/code/runtime"
	"strings"
	"testing"
)

func parseReviewExprs(t *testing.T, src string) []ast.Expr {
	t.Helper()
	ctx := &context.CodeContext{SourceCode: &src, FileName: "precedence-review.mist"}
	parser := NewLangParser(true, lex.NewLexer(ctx).Lex())
	return parser.ParseAll()
}

func evalReviewExpr(t *testing.T, src string) string {
	t.Helper()
	interp := runtime.NewInterpreter()
	return interp.RunGetLast(parseReviewExprs(t, src)).String()
}

func serializeReviewXML(t *testing.T, src string) string {
	t.Helper()
	exprs := parseReviewExprs(t, src)
	blocks := make([]ast.Block, len(exprs))
	for i, expr := range exprs {
		blocks[i] = expr.Blockly(true)
	}
	xmlBytes := ast.XmlRoot{
		Blocks: blocks,
		XMLNS:  "https://developers.google.com/blockly/xml",
	}.MarshalIndent("", "  ")
	return string(xmlBytes)
}

func roundTripReviewSource(t *testing.T, src string) string {
	t.Helper()
	xmlText := serializeReviewXML(t, src)
	exprs, err := blocklytomist.NewParser(xmlText).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("round-trip expressions = %d, want 1", len(exprs))
	}
	return exprs[0].String()
}

func childBlockByValueName(t *testing.T, block ast.Block, name string) ast.Block {
	t.Helper()
	for _, value := range block.Values {
		if value.Name == name {
			return value.Block
		}
	}
	t.Fatalf("block %q missing value %q", block.Type, name)
	return ast.Block{}
}

func numberValue(t *testing.T, block ast.Block) string {
	t.Helper()
	if block.Type != "math_number" {
		t.Fatalf("block type = %q, want math_number", block.Type)
	}
	for _, field := range block.Fields {
		if field.Name == "NUM" {
			return field.Value
		}
	}
	t.Fatalf("math_number missing NUM field")
	return ""
}

func TestPrecedenceReviewRuntimeResults(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"global x = 1\nx = 2\nx", "2"},
		{"1 : 2 + 3", "[1, 5]"},
		{`"a" _ "b" _ "c"`, "abc"},
		{`"a" _ 1 + 2`, "a3"},
		{"2 + 3 * 4", "14"},
		{"(2 + 3) * 4", "20"},
		{"2 + 3 * 4 ^ 2", "50"},
		{"(2 + 3) * 4 ^ 2", "80"},
		{"2 * 3 ^ 2", "18"},
		{"3.14 * 2 ^ 2", "12.56"},
		{"3.14 * (2 ^ 2)", "12.56"},
		{"7 % 4 ^ 2", "7"},
		{"7 % (4 ^ 2)", "7"},
		{"2 ^ 3 * 4", "32"},
		{"20 / 2 * 5", "50"},
		{"20 / 2 / 5", "2"},
		{"20 - 3 - 2", "15"},
		{"20 - (3 - 2)", "19"},
		{"2 ^ 3 ^ 2", "512"},
		{"(2 ^ 3) ^ 2", "64"},
		{"2 ^ (3 ^ 2)", "512"},
		{"1 + 2 < 4", "true"},
		{"1 < 2 == true", "true"},
		{"1 < 2 != false", "true"},
		{`"a" === "a"`, "true"},
		{`"a" !== "b"`, "true"},
		{`"a" << "b"`, "true"},
		{`"b" >> "a"`, "true"},
		{"6 & 3 ~ 1", "3"},
		{"6 & (3 ~ 1)", "2"},
		{"(6 & 3) ~ 1", "3"},
		{"1 | 2 & 0", "1"},
		{"(1 | 2) & 0", "0"},
		{"true || false && false", "true"},
		{"(true || false) && false", "false"},
		{"!false && true", "true"},
		{"-2 ^ 2", "-4"},
		{"-(2 ^ 2)", "-4"},
		{"2 * -3 ^ 2", "-18"},
		{"2 * (-3) ^ 2", "18"},
		{"matrix[[2, 0], [0, 2]] [^] 2 ^ 3", "[[256, 0], [0, 256]]"},
		{"(matrix[[2, 0], [0, 2]] [^] 2) [^] 3", "[[64, 0], [0, 64]]"},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			if got := evalReviewExpr(t, tt.src); got != tt.want {
				t.Fatalf("eval = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrecedenceReviewAllOperatorASTAndBlockly(t *testing.T) {
	tests := []struct {
		name          string
		src           string
		checkRoot     bool
		wantRoot      lex.Type
		wantBlockType string
	}{
		{name: "assignment", src: "global x = 1\nx = 2", wantBlockType: "lexical_variable_set"},
		{name: "pair", src: "1 : 2 + 3", wantBlockType: "pair"},
		{name: "text join", src: `"a" _ "b"`, checkRoot: true, wantRoot: lex.Underscore, wantBlockType: "text_join"},
		{name: "logic or", src: "true || false && false", checkRoot: true, wantRoot: lex.LogicOr, wantBlockType: "logic_operation"},
		{name: "logic and", src: "true && false", checkRoot: true, wantRoot: lex.LogicAnd, wantBlockType: "logic_operation"},
		{name: "bitwise or", src: "1 | 2 & 0", checkRoot: true, wantRoot: lex.BitwiseOr, wantBlockType: "math_bitwise"},
		{name: "bitwise and", src: "6 & 3", checkRoot: true, wantRoot: lex.BitwiseAnd, wantBlockType: "math_bitwise"},
		{name: "bitwise xor over and", src: "6 & 3 ~ 1", checkRoot: true, wantRoot: lex.BitwiseXor, wantBlockType: "math_bitwise"},
		{name: "bitwise xor", src: "3 ~ 1", checkRoot: true, wantRoot: lex.BitwiseXor, wantBlockType: "math_bitwise"},
		{name: "equals", src: "1 < 2 == true", checkRoot: true, wantRoot: lex.Equals, wantBlockType: "logic_compare"},
		{name: "not equals", src: "1 < 2 != false", checkRoot: true, wantRoot: lex.NotEquals, wantBlockType: "logic_compare"},
		{name: "text equals", src: `"a" === "a"`, checkRoot: true, wantRoot: lex.TextEquals, wantBlockType: "text_compare"},
		{name: "text not equals", src: `"a" !== "b"`, checkRoot: true, wantRoot: lex.TextNotEquals, wantBlockType: "text_compare"},
		{name: "less than", src: "1 + 2 < 4", checkRoot: true, wantRoot: lex.LessThan, wantBlockType: "math_compare"},
		{name: "less than equal", src: "1 + 2 <= 3", checkRoot: true, wantRoot: lex.LessThanEqual, wantBlockType: "math_compare"},
		{name: "greater than", src: "4 > 1 + 2", checkRoot: true, wantRoot: lex.GreatThan, wantBlockType: "math_compare"},
		{name: "greater than equal", src: "4 >= 1 + 2", checkRoot: true, wantRoot: lex.GreaterThanEqual, wantBlockType: "math_compare"},
		{name: "text less than", src: `"a" << "b"`, checkRoot: true, wantRoot: lex.TextLessThan, wantBlockType: "text_compare"},
		{name: "text greater than", src: `"b" >> "a"`, checkRoot: true, wantRoot: lex.TextGreaterThan, wantBlockType: "text_compare"},
		{name: "plus", src: "1 + 2 * 3", checkRoot: true, wantRoot: lex.Plus, wantBlockType: "math_add"},
		{name: "dash", src: "10 - 3 - 2", checkRoot: true, wantRoot: lex.Dash, wantBlockType: "math_subtract"},
		{name: "times", src: "2 * 3", checkRoot: true, wantRoot: lex.Times, wantBlockType: "math_multiply"},
		{name: "slash", src: "20 / 2 / 5", checkRoot: true, wantRoot: lex.Slash, wantBlockType: "math_division"},
		{name: "power", src: "2 ^ 3 ^ 2", checkRoot: true, wantRoot: lex.Power, wantBlockType: "math_power"},
		{name: "matrix plus", src: "matrix[[1, 0], [0, 1]] [+] matrix[[1, 0], [0, 1]]", checkRoot: true, wantRoot: lex.MatrixPlus, wantBlockType: "matrices_add"},
		{name: "matrix dash", src: "matrix[[1, 0], [0, 1]] [-] matrix[[1, 0], [0, 1]]", checkRoot: true, wantRoot: lex.MatrixDash, wantBlockType: "matrices_subtract"},
		{name: "matrix times", src: "matrix[[1, 0], [0, 1]] [*] matrix[[1, 0], [0, 1]]", checkRoot: true, wantRoot: lex.MatrixTimes, wantBlockType: "matrices_multiply"},
		{name: "matrix power", src: "matrix[[1, 0], [0, 1]] [^] 2", checkRoot: true, wantRoot: lex.MatrixPower, wantBlockType: "matrices_power"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprs := parseReviewExprs(t, tt.src)
			expr := exprs[len(exprs)-1]
			if set, ok := expr.(*variables.Set); ok {
				if got := set.Blockly(false).Type; got != tt.wantBlockType {
					t.Fatalf("Blockly type = %q, want %q", got, tt.wantBlockType)
				}
				return
			}
			if tt.checkRoot {
				bin, ok := expr.(*common.BinaryExpr)
				if !ok {
					t.Fatalf("expr = %T, want *common.BinaryExpr", expr)
				}
				if bin.Operator != tt.wantRoot {
					t.Fatalf("operator = %v, want %v", bin.Operator, tt.wantRoot)
				}
			}
			if got := expr.Blockly(false).Type; got != tt.wantBlockType {
				t.Fatalf("Blockly type = %q, want %q", got, tt.wantBlockType)
			}
		})
	}
}

func TestPrecedenceReviewRemainderOperator(t *testing.T) {
	expr := parseOneForTest(t, "7 % 4 ^ 2")
	call, ok := expr.(*common.FuncCall)
	if !ok || call.Name != "rem" {
		t.Fatalf("expr = %#v, want rem() func call", expr)
	}
	if _, ok := call.Args[1].(*common.BinaryExpr); !ok {
		t.Fatalf("remainder divisor = %#v, want nested power", call.Args[1])
	}
	if got := expr.Blockly(false).Type; got != "math_divide" {
		t.Fatalf("Blockly type = %q, want math_divide", got)
	}
	if got := evalReviewExpr(t, "7 % 4 ^ 2"); got != "7" {
		t.Fatalf("eval = %q, want 7", got)
	}
}

func TestPrecedenceReviewThreePointOneFourTimesRSquared(t *testing.T) {
	src := "func area(r) = 3.14 * r ^ 2"
	expr := parseOneForTest(t, src)
	proc, ok := expr.(*procedures.RetProcedure)
	if !ok {
		t.Fatalf("expr = %T, want *procedures.RetProcedure", expr)
	}
	times, ok := proc.Result.(*common.BinaryExpr)
	if !ok || times.Operator != lex.Times {
		t.Fatalf("result = %#v, want multiplication root", proc.Result)
	}
	power, ok := times.Operands[1].(*common.BinaryExpr)
	if !ok || power.Operator != lex.Power {
		t.Fatalf("multiplication right operand = %#v, want r ^ 2", times.Operands[1])
	}
	if got := proc.Result.Blockly(false).Type; got != "math_multiply" {
		t.Fatalf("Blockly type = %q, want math_multiply", got)
	}

	if got := evalReviewExpr(t, "3.14 * 2 ^ 2"); got != "12.56" {
		t.Fatalf("eval = %q, want 12.56", got)
	}
	if got := evalReviewExpr(t, "3.14 * (2 ^ 2)"); got != "12.56" {
		t.Fatalf("eval parenthesized grouping = %q, want 12.56", got)
	}
}

func TestPrecedenceReviewASTShape(t *testing.T) {
	root := parseOneForTest(t, "2 + 3 * 4 ^ 2").(*common.BinaryExpr)
	if root.Operator != lex.Plus {
		t.Fatalf("root operator = %v, want Plus", root.Operator)
	}
	multiplyOnRight := root.Operands[1].(*common.BinaryExpr)
	if multiplyOnRight.Operator != lex.Times {
		t.Fatalf("right operator = %v, want Times", multiplyOnRight.Operator)
	}
	powerInMultiply := multiplyOnRight.Operands[1].(*common.BinaryExpr)
	if powerInMultiply.Operator != lex.Power {
		t.Fatalf("nested operator = %v, want Power", powerInMultiply.Operator)
	}

	rightAssocPower := parseOneForTest(t, "2 ^ 3 ^ 2").(*common.BinaryExpr)
	if rightAssocPower.Operator != lex.Power {
		t.Fatalf("root operator = %v, want Power", rightAssocPower.Operator)
	}
	rightNestedPower, ok := rightAssocPower.Operands[1].(*common.BinaryExpr)
	if !ok || rightNestedPower.Operator != lex.Power {
		t.Fatalf("2 ^ 3 ^ 2 should parse as right-nested power, got %#v", rightAssocPower.Operands[1])
	}
}

func TestPrecedenceReviewBlocklySerializationShape(t *testing.T) {
	expr := parseOneForTest(t, "2 + 3 * 4 ^ 2")
	block := expr.Blockly(false)
	if block.Type != "math_add" {
		t.Fatalf("root block = %q, want math_add", block.Type)
	}
	multiply := childBlockByValueName(t, block, "NUM1")
	if multiply.Type != "math_multiply" {
		t.Fatalf("right block = %q, want math_multiply", multiply.Type)
	}
	power := childBlockByValueName(t, multiply, "NUM1")
	if power.Type != "math_power" {
		t.Fatalf("multiply right block = %q, want math_power", power.Type)
	}

	powerChain := parseOneForTest(t, "2 ^ 3 ^ 2").Blockly(false)
	if powerChain.Type != "math_power" {
		t.Fatalf("power chain root = %q, want math_power", powerChain.Type)
	}
	left := childBlockByValueName(t, powerChain, "A")
	right := childBlockByValueName(t, powerChain, "B")
	if got := numberValue(t, left); got != "2" {
		t.Fatalf("left child number = %q, want 2", got)
	}
	if right.Type != "math_power" {
		t.Fatalf("right child = %q, want math_power for right-associative grouping", right.Type)
	}
}

func TestPrecedenceReviewSerializedXMLRoundTrip(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"2 + 3 * 4 ^ 2", "2 + 3 * 4 ^ 2"},
		{"(2 + 3) * 4 ^ 2", "(2 + 3) * 4 ^ 2"},
		{"20 - (3 - 2)", "20 - (3 - 2)"},
		{"2 ^ 3 ^ 2", "2 ^ 3 ^ 2"},
		{"2 ^ (3 ^ 2)", "2 ^ 3 ^ 2"},
		{"(2 ^ 3) ^ 2", "(2 ^ 3) ^ 2"},
		{"matrix[[2, 0], [0, 2]] [^] 2 ^ 3", "matrix[[2, 0], [0, 2]] [^] 2 ^ 3"},
		{"(matrix[[2, 0], [0, 2]] [^] 2) [^] 3", "(matrix[[2, 0], [0, 2]] [^] 2) [^] 3"},
	}

	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			xmlText := serializeReviewXML(t, tt.src)
			if !strings.Contains(xmlText, `xmlns="https://developers.google.com/blockly/xml"`) {
				t.Fatalf("serialized XML missing Blockly namespace: %s", xmlText)
			}
			if !strings.Contains(xmlText, `type="math_`) && !strings.Contains(xmlText, `type="matrices_`) {
				t.Fatalf("serialized XML missing math or matrix block: %s", xmlText)
			}
			t.Logf("serialized Blockly XML for %q:\n%s", tt.src, xmlText)
			if got := roundTripReviewSource(t, tt.src); got != tt.want {
				t.Fatalf("round-trip source = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrecedenceReviewBlocklyXMLNestedRightPowerPreserved(t *testing.T) {
	xmlText := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="math_power">
    <value name="A"><block type="math_number"><field name="NUM">2</field></block></value>
    <value name="B">
      <block type="math_power">
        <value name="A"><block type="math_number"><field name="NUM">3</field></block></value>
        <value name="B"><block type="math_number"><field name="NUM">2</field></block></value>
      </block>
    </value>
  </block>
</xml>`
	exprs, err := blocklytomist.NewParser(xmlText).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if got := exprs[0].String(); got != "2 ^ 3 ^ 2" {
		t.Fatalf("source = %q, want right-nested power", got)
	}
	if got := runtime.NewInterpreter().RunGetLast(exprs).String(); got != "512" {
		t.Fatalf("eval = %q, want 512", got)
	}

	rightNested := exprs[0].(*common.BinaryExpr).Operands[1]
	if _, ok := rightNested.(*common.BinaryExpr); !ok {
		t.Fatalf("right operand = %#v, want nested binary power", rightNested)
	}
}

func TestPrecedenceReviewBlocklyXMLNestedSubtractionPreserved(t *testing.T) {
	xmlText := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="math_subtract">
    <value name="A"><block type="math_number"><field name="NUM">10</field></block></value>
    <value name="B">
      <block type="math_subtract">
        <value name="A"><block type="math_number"><field name="NUM">3</field></block></value>
        <value name="B"><block type="math_number"><field name="NUM">2</field></block></value>
      </block>
    </value>
  </block>
</xml>`
	exprs, err := blocklytomist.NewParser(xmlText).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if got := exprs[0].String(); got != "10 - (3 - 2)" {
		t.Fatalf("source = %q, want right-nested subtraction", got)
	}
	if got := runtime.NewInterpreter().RunGetLast(exprs).String(); got != "9" {
		t.Fatalf("eval = %q, want 9", got)
	}
}

func TestPrecedenceReviewParenthesizedRightOperandSerialization(t *testing.T) {
	text := roundTripReviewSource(t, "2 * (3 + 4)")
	if text != "2 * (3 + 4)" {
		t.Fatalf("round-trip source = %q, want parentheses preserved", text)
	}

	expr := parseOneForTest(t, "2 * (3 + 4)").(*common.BinaryExpr)
	if _, ok := expr.Operands[1].(*common.BinaryExpr); !ok {
		t.Fatalf("right operand = %#v, want nested addition", expr.Operands[1])
	}
	if _, ok := expr.Operands[0].(*fundamentals.Number); !ok {
		t.Fatalf("left operand = %#v, want number", expr.Operands[0])
	}
}
