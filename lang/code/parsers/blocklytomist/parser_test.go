package blocklytomist

import (
	"strings"
	"testing"
)

func TestComponentEventUsesMetadataParamsWhenMutationArgsMissing(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_event">
    <mutation component_type="Form" instance_name="Screen1" event_name="PermissionGranted"></mutation>
    <field name="COMPONENT_SELECTOR">Screen1</field>
  </block>
</xml>`

	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}

	got := exprs[0].String()
	want := "when Screen1.PermissionGranted(permissionName)"
	if !strings.Contains(got, want) {
		t.Fatalf("event source = %q, want it to contain %q", got, want)
	}
}

func TestGenericComponentEventUsesMetadataParamsWhenMutationArgsMissing(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_event">
    <mutation component_type="Button" is_generic="true" event_name="Click"></mutation>
    <field name="COMPONENT_SELECTOR">Button</field>
  </block>
</xml>`

	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}

	got := exprs[0].String()
	want := "when any Button.Click(component, notAlreadyHandled)"
	if !strings.Contains(got, want) {
		t.Fatalf("event source = %q, want it to contain %q", got, want)
	}
}

func TestOpenScreenWithStartValueBlockStringUsesComma(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="controls_openAnotherScreenWithStartValue">
    <value name="SCREENNAME"><block type="text"><field name="TEXT">Screen2</field></block></value>
    <value name="STARTVALUE"><block type="text"><field name="TEXT">payload</field></block></value>
  </block>
</xml>`

	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}

	if got, want := exprs[0].String(), `openScreen("Screen2", "payload")`; got != want {
		t.Fatalf("source = %q, want %q", got, want)
	}
}

func TestOpenScreenDoThenReturnArgumentStringIsSmartBody(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="controls_openAnotherScreen">
    <value name="SCREEN">
      <block type="controls_do_then_return">
        <statement name="STM">
          <block type="controls_closeScreen"></block>
        </statement>
        <value name="VALUE">
          <block type="helpers_screen_names"><field name="SCREEN">Menu</field></block>
        </value>
      </block>
    </value>
  </block>
</xml>`

	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}

	want := "openScreen({\n  closeScreen()\n  \"Menu\"\n})"
	if got := exprs[0].String(); got != want {
		t.Fatalf("source = %q, want %q", got, want)
	}
}

func TestComponentEventParameterNameMismatchReturnsError(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_event">
    <mutation component_type="Form" instance_name="Screen1" event_name="PermissionGranted">
      <arg name="permission"></arg>
    </mutation>
    <field name="COMPONENT_SELECTOR">Screen1</field>
  </block>
</xml>`

	_, err := NewParser(xml).TryGenerateAST()
	if err == nil {
		t.Fatal("TryGenerateAST() error = nil, want event parameter mismatch")
	}
	want := `event Form.PermissionGranted parameter 1 must be "permissionName", got "permission"`
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("TryGenerateAST() error = %q, want substring %q", err.Error(), want)
	}
}

func TestReservedBlocklyNamesGenerateEscapedMistIdentifiers(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "global declaration",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="global_declaration">
    <field name="NAME">when</field>
    <value name="VALUE"><block type="math_number"><field name="NUM">1</field></block></value>
  </block>
</xml>`,
			want: "global `when` = 1",
		},
		{
			name: "lexical get",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="lexical_variable_get">
    <field name="VAR">when</field>
  </block>
</xml>`,
			want: "`when`",
		},
		{
			name: "global set",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="lexical_variable_set">
    <field name="VAR">global when</field>
    <value name="VALUE"><block type="math_number"><field name="NUM">2</field></block></value>
  </block>
</xml>`,
			want: "this.`when` = 2",
		},
		{
			name: "local declaration",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="local_declaration_expression">
    <mutation><localname name="local"></localname></mutation>
    <field name="VAR0">local</field>
    <value name="DECL0"><block type="math_number"><field name="NUM">3</field></block></value>
    <value name="RETURN"><block type="lexical_variable_get"><field name="VAR">local</field></block></value>
  </block>
</xml>`,
			want: "local `local` = 3\n`local`",
		},
		{
			name: "procedure definition",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_defreturn">
    <mutation><arg name="global"></arg></mutation>
    <field name="NAME">func</field>
    <field name="VAR0">global</field>
    <value name="RETURN"><block type="lexical_variable_get"><field name="VAR">global</field></block></value>
  </block>
</xml>`,
			want: "func `func`(`global`) =\n  `global`\n",
		},
		{
			name: "procedure call",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_callreturn">
    <mutation name="when"><arg name="local"></arg></mutation>
    <field name="PROCNAME">when</field>
    <value name="ARG0"><block type="math_number"><field name="NUM">4</field></block></value>
  </block>
</xml>`,
			want: "`when`(4)",
		},
		{
			name: "component reference",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_component_block">
    <mutation component_type="Button" instance_name="when"></mutation>
    <field name="COMPONENT_SELECTOR">when</field>
  </block>
</xml>`,
			want: "`when`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := NewParser(tt.xml).TryGenerateAST()
			if err != nil {
				t.Fatalf("TryGenerateAST() error = %v", err)
			}
			if len(exprs) != 1 {
				t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
			}
			if got := exprs[0].String(); got != tt.want {
				t.Fatalf("source = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatrixArithmeticBlocksUseMatrixOperators(t *testing.T) {
	tests := []struct {
		name  string
		block string
		aName string
		bName string
		want  string
	}{
		{name: "add", block: "matrices_add", aName: "MAT0", bName: "MAT1", want: "1 [+] 2"},
		{name: "subtract", block: "matrices_subtract", aName: "A", bName: "B", want: "1 [-] 2"},
		{name: "multiply", block: "matrices_multiply", aName: "MAT0", bName: "MAT1", want: "1 [*] 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xml := matrixArithmeticXML(tt.block, tt.aName, tt.bName)
			exprs, err := NewParser(xml).TryGenerateAST()
			if err != nil {
				t.Fatalf("TryGenerateAST() error = %v", err)
			}
			if len(exprs) != 1 {
				t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
			}
			if got := exprs[0].String(); got != tt.want {
				t.Fatalf("source = %q, want %q", got, tt.want)
			}
			if got := exprs[0].Blockly(false).Type; got != tt.block {
				t.Fatalf("round-trip block type = %q, want %q", got, tt.block)
			}
		})
	}
}

func TestMatrixPowerBlockRoundTripsAsMatrixPower(t *testing.T) {
	xml := matrixArithmeticXML("matrices_power", "A", "B")
	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}
	if got := exprs[0].String(); got != "1 [^] 2" {
		t.Fatalf("source = %q, want %q", got, "1 [^] 2")
	}
	if got := exprs[0].Blockly(false).Type; got != "matrices_power" {
		t.Fatalf("round-trip block type = %q, want matrices_power", got)
	}
}

func TestAnonymousProcedureBlocksRoundTripToMist(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "void definition",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_defanonnoreturn">
    <mutation><arg name="x"></arg></mutation>
    <statement name="STACK">
      <block type="procedures_callanonnoreturn">
        <mutation items="1"></mutation>
        <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">greet</field></block></value>
        <value name="ARG0"><block type="text"><field name="TEXT">Melon</field></block></value>
      </block>
    </statement>
  </block>
</xml>`,
			want: "func(x) {\n  greet(\"Melon\")\n}",
		},
		{
			name: "returning definition",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_defanonreturn">
    <mutation><arg name="r"></arg></mutation>
    <value name="RETURN"><block type="math_number"><field name="NUM">5</field></block></value>
  </block>
</xml>`,
			want: "func(r) =\n  5\n",
		},
		{
			name: "input list call",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_callanonreturn_inputlist">
    <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">greet</field></block></value>
    <value name="INPUTLIST">
      <block type="lists_create_with">
        <mutation items="1"></mutation>
        <value name="ADD0"><block type="text"><field name="TEXT">Melon</field></block></value>
      </block>
    </value>
  </block>
</xml>`,
			want: "greet.call([\"Melon\"])",
		},
		{
			name: "num args",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_numArgs">
    <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">greet</field></block></value>
  </block>
</xml>`,
			want: "greet.numArgs()",
		},
		{
			name: "get with name",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_getWithName">
    <value name="PROCEDURENAME"><block type="text"><field name="TEXT">sayHello</field></block></value>
  </block>
</xml>`,
			want: "getFunc(\"sayHello\")",
		},
		{
			name: "get with dropdown",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_getWithDropdown">
    <field name="PROCNAME">sayHello</field>
  </block>
</xml>`,
			want: "func.sayHello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := NewParser(tt.xml).TryGenerateAST()
			if err != nil {
				t.Fatalf("TryGenerateAST() error = %v", err)
			}
			if len(exprs) != 1 {
				t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
			}
			if got := exprs[0].String(); got != tt.want {
				t.Fatalf("source = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcedureCategoryAliasesRoundTripToMist(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "procedure lexical getter",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedure_lexical_variable_get">
    <field name="VAR">x</field>
  </block>
</xml>`,
			want: "x",
		},
		{
			name: "procedure do then return",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_do_then_return">
    <statement name="STM">
      <block type="procedures_callanonnoreturn">
        <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">greet</field></block></value>
      </block>
    </statement>
    <value name="VALUE"><block type="math_number"><field name="NUM">5</field></block></value>
  </block>
</xml>`,
			want: "{\n  greet()\n  5\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := NewParser(tt.xml).TryGenerateAST()
			if err != nil {
				t.Fatalf("TryGenerateAST() error = %v", err)
			}
			if len(exprs) != 1 {
				t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
			}
			if got := exprs[0].String(); got != tt.want {
				t.Fatalf("source = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatrixCellBlocksUseDoubleSquareCoordinates(t *testing.T) {
	tests := []struct {
		name  string
		block string
		want  string
	}{
		{name: "get", block: "matrices_get_cell", want: "1[[2, 3]]"},
		{name: "set", block: "matrices_set_cell", want: "1[[2, 3]] = 9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := NewParser(matrixCellXML(tt.block)).TryGenerateAST()
			if err != nil {
				t.Fatalf("TryGenerateAST() error = %v", err)
			}
			if len(exprs) != 1 {
				t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
			}
			if got := exprs[0].String(); got != tt.want {
				t.Fatalf("source = %q, want %q", got, tt.want)
			}
			if got := exprs[0].Blockly(false).Type; got != tt.block {
				t.Fatalf("round-trip block type = %q, want %q", got, tt.block)
			}
		})
	}
}

func TestMatrixCreateBlockUsesMatrixLiteral(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="matrices_create">
    <mutation rows="2" cols="2" matrix="[[1,2],[3,4]]"></mutation>
    <field name="ROWS">2</field>
    <field name="COLS">2</field>
    <field name="MATRIX_0_0">1</field>
    <field name="MATRIX_0_1">2</field>
    <field name="MATRIX_1_0">3</field>
    <field name="MATRIX_1_1">4</field>
  </block>
</xml>`

	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}
	if got := exprs[0].String(); got != "matrix[[1, 2], [3, 4]]" {
		t.Fatalf("source = %q, want matrix literal", got)
	}
	if got := exprs[0].Blockly(false).Type; got != "matrices_create" {
		t.Fatalf("round-trip block type = %q, want matrices_create", got)
	}
}

func TestMatrixCreateBlockFallsBackToMutationMatrix(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="matrices_create">
    <mutation rows="2" cols="2" matrix="[[1.5,-2],[3,4]]"></mutation>
  </block>
</xml>`

	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}
	if got := exprs[0].String(); got != "matrix[[1.5, -2], [3, 4]]" {
		t.Fatalf("source = %q, want mutation-backed matrix literal", got)
	}
}

func TestMatrixCellBlocksInferDimensionsWithoutMutation(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="matrices_get_cell">
    <value name="MATRIX"><block type="math_number"><field name="NUM">1</field></block></value>
    <value name="DIM0"><block type="math_number"><field name="NUM">2</field></block></value>
    <value name="DIM1"><block type="math_number"><field name="NUM">3</field></block></value>
  </block>
</xml>`

	exprs, err := NewParser(xml).TryGenerateAST()
	if err != nil {
		t.Fatalf("TryGenerateAST() error = %v", err)
	}
	if len(exprs) != 1 {
		t.Fatalf("TryGenerateAST() produced %d expressions, want 1", len(exprs))
	}
	if got := exprs[0].String(); got != "1[[2, 3]]" {
		t.Fatalf("source = %q, want inferred matrix cell coordinates", got)
	}
}

func matrixArithmeticXML(blockType, aName, bName string) string {
	return `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="` + blockType + `">
    <value name="` + aName + `">
      <block type="math_number">
        <field name="NUM">1</field>
      </block>
    </value>
    <value name="` + bName + `">
      <block type="math_number">
        <field name="NUM">2</field>
      </block>
    </value>
  </block>
</xml>`
}

func matrixCellXML(blockType string) string {
	value := ""
	if blockType == "matrices_set_cell" {
		value = `
    <value name="VALUE">
      <block type="math_number">
        <field name="NUM">9</field>
      </block>
    </value>`
	}
	return `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="` + blockType + `">
    <mutation items="2"></mutation>
    <value name="MATRIX">
      <block type="math_number">
        <field name="NUM">1</field>
      </block>
    </value>
    <value name="DIM0">
      <block type="math_number">
        <field name="NUM">2</field>
      </block>
    </value>
    <value name="DIM1">
      <block type="math_number">
        <field name="NUM">3</field>
      </block>
    </value>` + value + `
  </block>
</xml>`
}
