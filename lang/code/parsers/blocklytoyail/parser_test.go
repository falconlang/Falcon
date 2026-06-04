package blocklytoyail

import (
	"strings"
	"testing"
)

func TestAllComponentsBlockUsesFullyQualifiedComponentType(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_all_component_block">
    <mutation component_type="Button"></mutation>
    <field name="COMPONENT_SELECTOR">Button</field>
  </block>
</xml>`

	yail, err := NewParser(xml).TryGenerateYAIL()
	if err != nil {
		t.Fatalf("TryGenerateYAIL() error = %v", err)
	}
	want := "(get-all-components com.google.appinventor.components.runtime.Button)"
	if yail != want {
		t.Fatalf("TryGenerateYAIL() = %q, want %q", yail, want)
	}
}

func TestComponentEventUsesMetadataParamsWhenMutationArgsMissing(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_event">
    <mutation component_type="Form" instance_name="Screen1" event_name="PermissionGranted"></mutation>
    <field name="COMPONENT_SELECTOR">Screen1</field>
  </block>
</xml>`

	yail, err := NewParser(xml).TryGenerateYAIL()
	if err != nil {
		t.Fatalf("TryGenerateYAIL() error = %v", err)
	}
	want := "(define-event Screen1 PermissionGranted ($permissionName)"
	if !strings.Contains(yail, want) {
		t.Fatalf("TryGenerateYAIL() = %q, want it to contain %q", yail, want)
	}
}

func TestGenericComponentEventUsesMetadataParamsWhenMutationArgsMissing(t *testing.T) {
	xml := `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_event">
    <mutation component_type="Button" is_generic="true" event_name="Click"></mutation>
    <field name="COMPONENT_SELECTOR">Button</field>
  </block>
</xml>`

	yail, err := NewParser(xml).TryGenerateYAIL()
	if err != nil {
		t.Fatalf("TryGenerateYAIL() error = %v", err)
	}
	want := "(define-generic-event com.google.appinventor.components.runtime.Button Click ($component $notAlreadyHandled)"
	if !strings.Contains(yail, want) {
		t.Fatalf("TryGenerateYAIL() = %q, want it to contain %q", yail, want)
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

	_, err := NewParser(xml).TryGenerateYAIL()
	if err == nil {
		t.Fatal("TryGenerateYAIL() error = nil, want event parameter mismatch")
	}
	want := `event Form.PermissionGranted parameter 1 must be "permissionName", got "permission"`
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("TryGenerateYAIL() error = %q, want substring %q", err.Error(), want)
	}
}

func TestTryGenerateYAILReturnsMalformedXmlError(t *testing.T) {
	_, err := NewParser(`<xml><block></xml>`).TryGenerateYAIL()
	if err == nil {
		t.Fatal("TryGenerateYAIL() error = nil, want malformed XML error")
	}
}

func TestAnonymousProcedureBlocksGenerateYAIL(t *testing.T) {
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
        <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">greet</field></block></value>
      </block>
    </statement>
  </block>
</xml>`,
			want: `(call-yail-primitive create-yail-procedure (*list-for-runtime* (lambda ($x) (call-yail-primitive call-yail-procedure (*list-for-runtime* (lexical-value $greet)) '(any) "call procedure"))) '(any) "create procedure")`,
		},
		{
			name: "returning definition",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_defanonreturn">
    <mutation><arg name="r"></arg></mutation>
    <value name="RETURN"><block type="math_number"><field name="NUM">5</field></block></value>
  </block>
</xml>`,
			want: `(call-yail-primitive create-yail-procedure (*list-for-runtime* (lambda ($r) 5)) '(any) "create procedure")`,
		},
		{
			name: "direct call",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_callanonreturn">
    <mutation items="1"></mutation>
    <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">f</field></block></value>
    <value name="ARG0"><block type="math_number"><field name="NUM">5</field></block></value>
  </block>
</xml>`,
			want: `(call-yail-primitive call-yail-procedure (*list-for-runtime* (lexical-value $f) 5) '(any any) "call procedure")`,
		},
		{
			name: "input list call",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_callanonreturn_inputlist">
    <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">f</field></block></value>
    <value name="INPUTLIST"><block type="lists_create_with"><mutation items="0"></mutation></block></value>
  </block>
</xml>`,
			want: `(call-yail-primitive call-yail-procedure-input-list (*list-for-runtime* (lexical-value $f) (call-yail-primitive make-yail-list (*list-for-runtime*) '() "make a list")) '(any any) "call procedure(with input list)")`,
		},
		{
			name: "num args",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_numArgs">
    <value name="PROCEDURE"><block type="lexical_variable_get"><field name="VAR">f</field></block></value>
  </block>
</xml>`,
			want: `(call-yail-primitive num-args-yail-procedure (*list-for-runtime* (lexical-value $f)) '(any) "get number of arguments")`,
		},
		{
			name: "get with name",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_getWithName">
    <value name="PROCEDURENAME"><block type="text"><field name="TEXT">sayHello</field></block></value>
  </block>
</xml>`,
			want: `(call-yail-primitive create-yail-procedure-with-name (*list-for-runtime* "sayHello") '(any) "get procedure")`,
		},
		{
			name: "get with dropdown",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_getWithDropdown">
    <field name="PROCNAME">sayHello</field>
  </block>
</xml>`,
			want: `(call-yail-primitive create-yail-procedure (*list-for-runtime* (get-var p$sayHello)) '(any) "get procedure")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yail := generateYAILForTest(t, tt.xml)
			if yail != tt.want {
				t.Fatalf("YAIL = %q, want %q", yail, tt.want)
			}
		})
	}
}

func TestAppInventorPrefixedVariableNamesGenerateCanonicalYAIL(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "prefixed local declaration",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="local_declaration_expression">
    <mutation>
      <localname name="item"></localname>
    </mutation>
    <field name="VAR0">local item</field>
    <value name="DECL0"><block type="math_number"><field name="NUM">7</field></block></value>
    <value name="RETURN"><block type="lexical_variable_get"><field name="VAR">local item</field></block></value>
  </block>
</xml>`,
			want: `(let (($local_item 7)) (lexical-value $local_item))`,
		},
		{
			name: "prefixed procedure parameter",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="procedures_defreturn">
    <mutation><arg name="param value"></arg></mutation>
    <field name="NAME">echo</field>
    <value name="RETURN"><block type="procedure_lexical_variable_get"><field name="VAR">param value</field></block></value>
  </block>
</xml>`,
			want: `(def (p$echo $param_value) (lexical-value $param_value))`,
		},
		{
			name: "prefixed list transform variable",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="lists_map">
    <field name="VAR">local item</field>
    <value name="TO"><block type="lexical_variable_get"><field name="VAR">local item</field></block></value>
  </block>
</xml>`,
			want: `(map_nondest $local_item (lexical-value $local_item)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yail := generateYAILForTest(t, tt.xml)
			if !strings.Contains(yail, tt.want) {
				t.Fatalf("YAIL = %q, want substring %q", yail, tt.want)
			}
		})
	}
}

func TestMatrixCreateYAILUsesFields(t *testing.T) {
	yail := generateYAILForTest(t, `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="matrices_create">
    <mutation rows="2" cols="2" matrix="[[1,2],[3,4]]"></mutation>
    <field name="ROWS">2</field>
    <field name="COLS">2</field>
    <field name="MATRIX_0_0">1</field>
    <field name="MATRIX_0_1">2</field>
    <field name="MATRIX_1_0">3</field>
    <field name="MATRIX_1_1">4</field>
  </block>
</xml>`)

	want := `(call-yail-primitive make-yail-matrix (*list-for-runtime* 2 2 1 2 3 4) '(number number number number number number) "create a matrix")`
	if yail != want {
		t.Fatalf("matrix create YAIL = %q, want %q", yail, want)
	}
}

func TestMatrixCreateYAILFallsBackToMutationMatrix(t *testing.T) {
	yail := generateYAILForTest(t, `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="matrices_create">
    <mutation rows="2" cols="2" matrix="[[1.5,-2],[3,4]]"></mutation>
  </block>
</xml>`)

	want := `(call-yail-primitive make-yail-matrix (*list-for-runtime* 2 2 1.5 -2 3 4) '(number number number number number number) "create a matrix")`
	if yail != want {
		t.Fatalf("matrix create mutation YAIL = %q, want %q", yail, want)
	}
}

func TestMatrixBlocksGenerateYAILPrimitives(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "get row",
			xml:  matrixBlockXML("matrices_get_row", matrixNumberValue("MATRIX", "1"), matrixNumberValue("ROW", "2")),
			want: `yail-matrix-get-row (*list-for-runtime* 1 2) '(matrix number) "get matrix row"`,
		},
		{
			name: "get column",
			xml:  matrixBlockXML("matrices_get_column", matrixNumberValue("MATRIX", "1"), matrixNumberValue("COLUMN", "3")),
			want: `yail-matrix-get-column (*list-for-runtime* 1 3) '(matrix number) "get matrix column"`,
		},
		{
			name: "get cell",
			xml: matrixBlockXML("matrices_get_cell",
				`<mutation items="2"></mutation>`,
				matrixNumberValue("MATRIX", "1"),
				matrixNumberValue("DIM0", "2"),
				matrixNumberValue("DIM1", "3")),
			want: `yail-matrix-get-cell (*list-for-runtime* 1 2 3) '(matrix number number) "get matrix cell"`,
		},
		{
			name: "set cell",
			xml: matrixBlockXML("matrices_set_cell",
				`<mutation items="2"></mutation>`,
				matrixNumberValue("MATRIX", "1"),
				matrixNumberValue("DIM0", "2"),
				matrixNumberValue("DIM1", "3"),
				matrixNumberValue("VALUE", "9")),
			want: `yail-matrix-set-cell! (*list-for-runtime* 1 2 3 9) '(matrix number number number) "set matrix cell"`,
		},
		{
			name: "dimensions",
			xml:  matrixBlockXML("matrices_get_dims", matrixNumberValue("MATRIX", "1")),
			want: `yail-matrix-get-dims (*list-for-runtime* 1) '(matrix) "get matrix dimensions"`,
		},
		{
			name: "is matrix",
			xml:  matrixBlockXML("matrices_is_matrix", matrixNumberValue("VALUE", "1")),
			want: `yail-matrix? (*list-for-runtime* 1) '(any) "is matrix?"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yail := generateYAILForTest(t, tt.xml)
			if !strings.Contains(yail, tt.want) {
				t.Fatalf("YAIL = %q, want substring %q", yail, tt.want)
			}
		})
	}
}

func TestMatrixOperationYAILPrimitives(t *testing.T) {
	tests := []struct {
		op   string
		want string
	}{
		{op: "INVERSE", want: `yail-matrix-inverse (*list-for-runtime* 1) '(matrix) "inverse"`},
		{op: "TRANSPOSE", want: `yail-matrix-transpose (*list-for-runtime* 1) '(matrix) "transpose"`},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			yail := generateYAILForTest(t, matrixBlockXML("matrices_operations",
				`<field name="OP">`+tt.op+`</field>`,
				matrixNumberValue("MATRIX", "1")))
			if !strings.Contains(yail, tt.want) {
				t.Fatalf("YAIL = %q, want substring %q", yail, tt.want)
			}
		})
	}
}

func TestIOSUnsupportedBlocksReturnErrors(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "run in background",
			xml:  matrixBlockXML("controls_run_in_background"),
			want: "controls_run_in_background is not supported by the App Inventor iOS runtime",
		},
		{
			name: "run after period",
			xml:  matrixBlockXML("controls_run_after_period"),
			want: "controls_run_after_period is not supported by the App Inventor iOS runtime",
		},
		{
			name: "multidimensional matrix",
			xml:  matrixBlockXML("matrices_create_multidim", matrixNumberValue("DIM", "2"), matrixNumberValue("INITIAL", "9")),
			want: "matrices_create_multidim is not supported by the App Inventor iOS runtime",
		},
		{
			name: "rotate left",
			xml: matrixBlockXML("matrices_operations",
				`<field name="OP">ROTATE_LEFT</field>`,
				matrixNumberValue("MATRIX", "1")),
			want: "matrices_rotate_left is not supported by the App Inventor iOS runtime",
		},
		{
			name: "rotate right",
			xml: matrixBlockXML("matrices_operations",
				`<field name="OP">ROTATE_RIGHT</field>`,
				matrixNumberValue("MATRIX", "1")),
			want: "matrices_rotate_right is not supported by the App Inventor iOS runtime",
		},
		{
			name: "3d matrix cell",
			xml: matrixBlockXML("matrices_get_cell",
				`<mutation items="3"></mutation>`,
				matrixNumberValue("MATRIX", "1"),
				matrixNumberValue("DIM0", "2"),
				matrixNumberValue("DIM1", "3"),
				matrixNumberValue("DIM2", "4")),
			want: "matrices_get_cell is not supported by the App Inventor iOS runtime",
		},
		{
			name: "generic continuation method",
			xml: `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="component_method">
    <mutation is_generic="true" component_type="File" method_name="Exists">
      <arg name="scope" type="com.google.appinventor.components.common.FileScopeEnum"></arg>
      <arg name="path" type="text"></arg>
    </mutation>
    <value name="COMPONENT"><block type="component_component_block"><field name="COMPONENT_SELECTOR">File1</field></block></value>
    <value name="ARG0"><block type="text"><field name="TEXT">App</field></block></value>
    <value name="ARG1"><block type="text"><field name="TEXT">data.txt</field></block></value>
  </block>
</xml>`,
			want: "component_method is not supported by the App Inventor iOS runtime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewParser(tt.xml).TryGenerateYAIL()
			if err == nil {
				t.Fatal("TryGenerateYAIL() error = nil, want unsupported iOS block error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("TryGenerateYAIL() error = %q, want substring %q", err.Error(), tt.want)
			}
		})
	}
}

func TestAppInventorArithmeticAliasBlocksGenerateYAIL(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "math binary alias",
			xml: matrixBlockXML("math_arithmetic",
				`<field name="OP">DIVIDE</field>`,
				matrixNumberValue("A", "8"),
				matrixNumberValue("B", "2")),
			want: `yail-divide (*list-for-runtime* 8 2) '(number number) "yail-divide"`,
		},
		{
			name: "math variadic alias",
			xml: matrixBlockXML("math_arithmetic_list",
				`<field name="OP">MULTIPLY</field>`,
				`<mutation items="3"></mutation>`,
				matrixNumberValue("NUM0", "2"),
				matrixNumberValue("NUM1", "3"),
				matrixNumberValue("NUM2", "4")),
			want: `* (*list-for-runtime* 2 3 4) '(number number number) "*"`,
		},
		{
			name: "matrix binary alias",
			xml: matrixBlockXML("matrices_arithmetic",
				`<field name="OP">POWER</field>`,
				matrixNumberValue("A", "2"),
				matrixNumberValue("B", "3")),
			want: `yail-matrix-power (*list-for-runtime* 2 3) '(matrix number) "yail-matrix-power"`,
		},
		{
			name: "matrix variadic alias",
			xml: matrixBlockXML("matrices_arithmetic_list",
				`<field name="OP">MULTIPLY</field>`,
				`<mutation items="2"></mutation>`,
				matrixNumberValue("MAT0", "5"),
				matrixNumberValue("MAT1", "6")),
			want: `yail-matrix-multiply (*list-for-runtime* 5 6) '(matrix any) "yail-matrix-multiply"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yail := generateYAILForTest(t, tt.xml)
			if !strings.Contains(yail, tt.want) {
				t.Fatalf("YAIL = %q, want substring %q", yail, tt.want)
			}
		})
	}
}

func TestMatrixArithmeticYAILPrimitives(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "add variadic",
			xml: matrixBlockXML("matrices_add",
				`<mutation items="3"></mutation>`,
				matrixNumberValue("MAT0", "1"),
				matrixNumberValue("MAT1", "2"),
				matrixNumberValue("MAT2", "3")),
			want: `yail-matrix-add (*list-for-runtime* 1 2 3) '(matrix matrix matrix) "yail-matrix-add"`,
		},
		{
			name: "add binary fallback",
			xml:  matrixBlockXML("matrices_add", matrixNumberValue("A", "1"), matrixNumberValue("B", "2")),
			want: `yail-matrix-add (*list-for-runtime* 1 2) '(matrix matrix) "yail-matrix-add"`,
		},
		{
			name: "subtract",
			xml:  matrixBlockXML("matrices_subtract", matrixNumberValue("A", "1"), matrixNumberValue("B", "2")),
			want: `yail-matrix-subtract (*list-for-runtime* 1 2) '(matrix matrix) "yail-matrix-subtract"`,
		},
		{
			name: "multiply variadic",
			xml: matrixBlockXML("matrices_multiply",
				`<mutation items="2"></mutation>`,
				matrixNumberValue("MAT0", "1"),
				matrixNumberValue("MAT1", "2")),
			want: `yail-matrix-multiply (*list-for-runtime* 1 2) '(matrix any) "yail-matrix-multiply"`,
		},
		{
			name: "multiply binary fallback",
			xml:  matrixBlockXML("matrices_multiply", matrixNumberValue("A", "1"), matrixNumberValue("B", "2")),
			want: `yail-matrix-multiply (*list-for-runtime* 1 2) '(matrix any) "yail-matrix-multiply"`,
		},
		{
			name: "power",
			xml:  matrixBlockXML("matrices_power", matrixNumberValue("A", "1"), matrixNumberValue("B", "2")),
			want: `yail-matrix-power (*list-for-runtime* 1 2) '(matrix number) "yail-matrix-power"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yail := generateYAILForTest(t, tt.xml)
			if !strings.Contains(yail, tt.want) {
				t.Fatalf("YAIL = %q, want substring %q", yail, tt.want)
			}
		})
	}
}

func generateYAILForTest(t *testing.T, xml string) string {
	t.Helper()
	yail, err := NewParser(xml).TryGenerateYAIL()
	if err != nil {
		t.Fatalf("TryGenerateYAIL() error = %v", err)
	}
	return yail
}

func matrixBlockXML(blockType string, body ...string) string {
	return `<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="` + blockType + `">
    ` + strings.Join(body, "\n    ") + `
  </block>
</xml>`
}

func matrixNumberValue(name, value string) string {
	return `<value name="` + name + `">
      <block type="math_number">
        <field name="NUM">` + value + `</field>
      </block>
    </value>`
}
