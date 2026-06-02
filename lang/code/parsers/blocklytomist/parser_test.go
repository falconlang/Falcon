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
