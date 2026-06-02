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
