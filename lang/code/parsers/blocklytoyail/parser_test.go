package blocklytoyail

import "testing"

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

func TestTryGenerateYAILReturnsMalformedXmlError(t *testing.T) {
	_, err := NewParser(`<xml><block></xml>`).TryGenerateYAIL()
	if err == nil {
		t.Fatal("TryGenerateYAIL() error = nil, want malformed XML error")
	}
}
