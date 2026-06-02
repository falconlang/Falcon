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
			name: "create multidim",
			xml:  matrixBlockXML("matrices_create_multidim", matrixNumberValue("DIM", "2"), matrixNumberValue("INITIAL", "9")),
			want: `make-yail-matrix-multidim (*list-for-runtime* 2 9) '(list number) "create multidimensional matrix"`,
		},
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
			want: `yail-matrix-set-cell! (*list-for-runtime* 1 9 2 3) '(matrix number number number) "set matrix cell"`,
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
		{op: "ROTATE_LEFT", want: `yail-matrix-rotate-left (*list-for-runtime* 1) '(matrix) "rotate_left"`},
		{op: "ROTATE_RIGHT", want: `yail-matrix-rotate-right (*list-for-runtime* 1) '(matrix) "rotate_right"`},
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
