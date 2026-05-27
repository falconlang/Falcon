package design

import (
	"os"
	"strings"
	"testing"
)

func TestParseAnnRejectsTrailingInput(t *testing.T) {
	source := `Screen.Screen1 {} Button { Text: "ignored" }`

	if _, err := ParseAnn(source); err == nil || !strings.Contains(err.Error(), "trailing input") {
		t.Fatalf("ParseAnn() error = %v, want trailing input error", err)
	}
	if _, err := NewAimlParser(source).ConvertAimlToSchema(); err == nil || !strings.Contains(err.Error(), "trailing input") {
		t.Fatalf("ConvertAimlToSchema() error = %v, want trailing input error", err)
	}
}

func TestParseAnnReadsDotIds(t *testing.T) {
	source := `Screen.Screen1 { Title: "Calculator", Button.AddButton { Text: "+" }, Notifier.Notifier1 }`

	screen, err := ParseAnn(source)
	if err != nil {
		t.Fatalf("ParseAnn() error = %v", err)
	}
	if screen.Id != "Screen1" {
		t.Fatalf("screen.Id = %q, want Screen1", screen.Id)
	}
	if got := screen.Properties["Title"]; got != "Calculator" {
		t.Fatalf("screen Title = %q, want Calculator", got)
	}
	if len(screen.Children) != 2 {
		t.Fatalf("len(screen.Children) = %d, want 2", len(screen.Children))
	}
	if got := screen.Children[0].Id; got != "AddButton" {
		t.Fatalf("button id = %q, want AddButton", got)
	}
	if got := screen.Children[0].Properties["Text"]; got != "+" {
		t.Fatalf("button Text = %q, want +", got)
	}
	if got := screen.Children[1].Id; got != "Notifier1" {
		t.Fatalf("notifier id = %q, want Notifier1", got)
	}
}

func TestSchemaToAimlWritesDotIds(t *testing.T) {
	source := `{"Properties":{"$Name":"Screen1","$Components":[{"$Type":"Button","$Name":"AddButton","Text":"+"},{"$Type":"Notifier","$Name":"Notifier1"}]}}`

	ann, err := NewSchemaParser(source).ConvertSchemaToAiml()
	if err != nil {
		t.Fatalf("ConvertSchemaToAiml() error = %v", err)
	}
	for _, want := range []string{"Screen.Screen1 {", "Button.AddButton { Text: \"+\" }", "Notifier.Notifier1"} {
		if !strings.Contains(ann, want) {
			t.Fatalf("ConvertSchemaToAiml() =\n%s\nwant substring %q", ann, want)
		}
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
