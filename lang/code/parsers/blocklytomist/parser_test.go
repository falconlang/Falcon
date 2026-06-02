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
