package design

import (
	"os"
	"strings"
	"testing"
)

func TestParseAnnRejectsTrailingInput(t *testing.T) {
	source := `@Screen { id: "Screen1" } @Button { Text: "ignored" }`

	if _, err := ParseAnn(source); err == nil || !strings.Contains(err.Error(), "trailing input") {
		t.Fatalf("ParseAnn() error = %v, want trailing input error", err)
	}
	if _, err := NewAimlParser(source).ConvertAimlToSchema(); err == nil || !strings.Contains(err.Error(), "trailing input") {
		t.Fatalf("ConvertAimlToSchema() error = %v, want trailing input error", err)
	}
}

func TestScreenAnnFixtureValidates(t *testing.T) {
	source, err := os.ReadFile("../testing/Screen1.ann")
	if err != nil {
		t.Fatal(err)
	}
	yail, err := NewAnnYailConverter().ConvertAnnToYail(string(source))
	if err != nil {
		t.Fatalf("ConvertAnnToYail() error = %v", err)
	}
	if !strings.Contains(yail, "NumbersOnly") || !strings.Contains(yail, "Hint") {
		t.Fatalf("generated YAIL does not include canonical TextBox properties:\n%s", yail)
	}
}

func TestSchemaAndXmlConvertersReturnErrors(t *testing.T) {
	if _, err := NewSchemaParser(`{"Properties":[]}`).ConvertSchemaToXml(); err == nil {
		t.Fatal("ConvertSchemaToXml() error = nil, want malformed schema error")
	}
	if _, err := NewSchemaParser(`{"Properties":{"$Name":"Screen1","$Components":[42]}}`).ConvertSchemaToAiml(); err == nil {
		t.Fatal("ConvertSchemaToAiml() error = nil, want malformed component error")
	}
	if _, err := NewXmlParser(`<Screen><Button></Screen>`).ConvertXmlToSchema(); err == nil {
		t.Fatal("ConvertXmlToSchema() error = nil, want malformed XML error")
	}
}
