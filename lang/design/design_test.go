package design

import (
	"encoding/json"
	"errors"
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

func TestParseAnnSkipsLineComments(t *testing.T) {
	source := `// screen comment
Screen.Screen1 {
  Title: "Calculator // not a comment",
  // button comment
  Button.AddButton {
    Text: "+"
  }
}`

	screen, err := ParseAnn(source)
	if err != nil {
		t.Fatalf("ParseAnn() error = %v", err)
	}
	if screen.Id != "Screen1" {
		t.Fatalf("screen.Id = %q, want Screen1", screen.Id)
	}
	if got := screen.Properties["Title"]; got != "Calculator // not a comment" {
		t.Fatalf("screen Title = %q, want string content preserved", got)
	}
	if len(screen.Children) != 1 || screen.Children[0].Id != "AddButton" {
		t.Fatalf("screen.Children = %#v, want AddButton", screen.Children)
	}
}

func TestParseAnnNormalizesColorLiterals(t *testing.T) {
	source := `Screen.Screen1 {
	  BackgroundColor: &HFF446A98,
	  Button.AddButton {
	    BackgroundColor: #0FFFFF,
	    TextColor: #336699CC
	  }
	}`

	screen, err := ParseAnn(source)
	if err != nil {
		t.Fatalf("ParseAnn() error = %v", err)
	}
	if got := screen.Properties["BackgroundColor"]; got != "-12293480" {
		t.Fatalf("screen BackgroundColor = %q, want -12293480", got)
	}
	if len(screen.Children) != 1 {
		t.Fatalf("len(screen.Children) = %d, want 1", len(screen.Children))
	}
	button := screen.Children[0]
	if got := button.Properties["BackgroundColor"]; got != "-15728641" {
		t.Fatalf("button BackgroundColor = %q, want -15728641", got)
	}
	if got := button.Properties["TextColor"]; got != "-869046631" {
		t.Fatalf("button TextColor = %q, want -869046631", got)
	}
}

func TestAnnYailConvertsColorLiteralsToNumbers(t *testing.T) {
	source := `Screen.Screen1 {
  BackgroundColor: &HFF446A98,
  Button.AddButton {
    BackgroundColor: #FFFFFF,
    TextColor: "#336699CC"
  }
}`

	yail, err := NewAnnYailConverter().ConvertAnnToYail(source)
	if err != nil {
		t.Fatalf("ConvertAnnToYail() error = %v", err)
	}
	for _, want := range []string{
		"(set-and-coerce-property! 'Screen1 'BackgroundColor -12293480 'number)",
		"(set-and-coerce-property! 'AddButton 'BackgroundColor -1 'number)",
		"(set-and-coerce-property! 'AddButton 'TextColor -869046631 'number)",
	} {
		if !strings.Contains(yail, want) {
			t.Fatalf("generated YAIL does not contain %q:\n%s", want, yail)
		}
	}
	if strings.Contains(yail, `"&H`) || strings.Contains(yail, `"#`) {
		t.Fatalf("generated YAIL quoted a color literal:\n%s", yail)
	}
}

func TestAnnYailUsesComponentMetadataForValues(t *testing.T) {
	source := `Screen.Screen1 {
  Button.HideButton { Text: "Café", Visible: False }
  BluetoothClient.BluetoothClient1
  NxtDrive.Drive { BluetoothClient: BluetoothClient1 }
}`

	yail, err := NewAnnYailConverter().ConvertAnnToYail(source)
	if err != nil {
		t.Fatalf("ConvertAnnToYail() error = %v", err)
	}
	for _, want := range []string{
		`(set-and-coerce-property! 'HideButton 'Visible #f 'boolean)`,
		`(set-and-coerce-property! 'Drive 'BluetoothClient (get-component BluetoothClient1) 'component)`,
		`"Caf\u00e9"`,
	} {
		if !strings.Contains(yail, want) {
			t.Fatalf("generated YAIL does not contain %q:\n%s", want, yail)
		}
	}
	if strings.Contains(yail, `\u00c3\u00a9`) {
		t.Fatalf("generated YAIL used UTF-8 byte escapes instead of UTF-16 code units:\n%s", yail)
	}
}

func TestAnnYailSkipsInternalPropertiesAndAlwaysSendsDefaults(t *testing.T) {
	source := `Screen.Screen1 {
  Uuid: "0",
  TutorialURL: "https://example.invalid/tutorial",
  BlocksToolkit: "{}"
}`

	yail, err := NewAnnYailConverter().ConvertAnnToYail(source)
	if err != nil {
		t.Fatalf("ConvertAnnToYail() error = %v", err)
	}
	for _, forbidden := range []string{"Uuid", "TutorialURL", "BlocksToolkit"} {
		if strings.Contains(yail, forbidden) {
			t.Fatalf("generated YAIL contains internal property %q:\n%s", forbidden, yail)
		}
	}
	for _, want := range []string{
		`(set-and-coerce-property! 'Screen1 'ShowListsAsJson #t 'boolean)`,
		`(set-and-coerce-property! 'Screen1 'Sizing "Responsive" 'text)`,
	} {
		if !strings.Contains(yail, want) {
			t.Fatalf("generated YAIL does not contain always-send default %q:\n%s", want, yail)
		}
	}
}

func TestAnnYailSetsFormNameBeforeOptionalRename(t *testing.T) {
	source := `Screen.DetailScreen { Title: "Detail" }`

	yail, err := NewAnnYailConverter().ConvertAnnToYail(source)
	if err != nil {
		t.Fatalf("ConvertAnnToYail() error = %v", err)
	}

	setForm := `(set-form-name "DetailScreen")`
	rename := `(rename-component "Screen1" "DetailScreen")`
	if !strings.Contains(yail, setForm) {
		t.Fatalf("generated YAIL does not contain %q:\n%s", setForm, yail)
	}
	if !strings.Contains(yail, rename) {
		t.Fatalf("generated YAIL does not contain %q:\n%s", rename, yail)
	}
	if strings.Index(yail, setForm) > strings.Index(yail, rename) {
		t.Fatalf("generated YAIL sets form name after rename:\n%s", yail)
	}
}

func TestAnnValidationRejectsUnknownDuplicateAndIllegalContainment(t *testing.T) {
	source := `Screen.Screen1 {
  Button.Dup
  Label.Dup
  Button.Bad { Label.Child }
  Mystery.Thing
}`

	_, err := NewAnnYailConverter().ConvertAnnToYail(source)
	if err == nil {
		t.Fatal("ConvertAnnToYail() error = nil, want validation diagnostics")
	}
	for _, want := range []string{
		`duplicate component name "Dup"`,
		`Label.Child cannot be placed inside Button.Bad`,
		`unknown component type "Mystery"`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("validation error = %q, want substring %q", err.Error(), want)
		}
	}
}

func TestAnnValidationReportsPropertyPosition(t *testing.T) {
	source := "Screen.Screen1 {\n  Button.AddButton { Texxt: \"+\" }\n}"
	_, err := NewAnnYailConverter().ConvertAnnToYail(source)
	if err == nil {
		t.Fatal("ConvertAnnToYail() error = nil, want property diagnostic")
	}

	var diagnosticErr *AnnDiagnosticListError
	if !errors.As(err, &diagnosticErr) {
		t.Fatalf("ConvertAnnToYail() error = %T %v, want *AnnDiagnosticListError", err, err)
	}
	if len(diagnosticErr.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %#v, want one diagnostic", diagnosticErr.Diagnostics)
	}

	diagnostic := diagnosticErr.Diagnostics[0]
	wantPosition := strings.Index(source, "Texxt")
	if diagnostic.Position != wantPosition {
		t.Fatalf("diagnostic.Position = %d, want %d", diagnostic.Position, wantPosition)
	}
	if diagnostic.Length != len("Texxt") {
		t.Fatalf("diagnostic.Length = %d, want %d", diagnostic.Length, len("Texxt"))
	}
	if !strings.Contains(diagnostic.Message, `unknown property "Texxt"`) {
		t.Fatalf("diagnostic.Message = %q, want unknown property message", diagnostic.Message)
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

func TestSchemaToAimlSkipsInternalDesignerProperties(t *testing.T) {
	source := `{"Properties":{"$Name":"Screen1","Uuid":"0","TutorialURL":"https://example.invalid","BlocksToolkit":"{}","$Components":[]}}`

	ann, err := NewSchemaParser(source).ConvertSchemaToAiml()
	if err != nil {
		t.Fatalf("ConvertSchemaToAiml() error = %v", err)
	}
	for _, forbidden := range []string{"Uuid", "TutorialURL", "BlocksToolkit"} {
		if strings.Contains(ann, forbidden) {
			t.Fatalf("ConvertSchemaToAiml() contains internal property %q:\n%s", forbidden, ann)
		}
	}
}

func TestAimlToSchemaUsesCurrentAppInventorVersions(t *testing.T) {
	source := `Screen.Screen1 { Button.AddButton }`

	schemaText, err := NewAimlParser(source).ConvertAimlToSchema()
	if err != nil {
		t.Fatalf("ConvertAimlToSchema() error = %v", err)
	}

	var schema struct {
		YaVersion  string `json:"YaVersion"`
		Properties struct {
			Version    string `json:"$Version"`
			Components []struct {
				Type    string `json:"$Type"`
				Version string `json:"$Version"`
			} `json:"$Components"`
		} `json:"Properties"`
	}
	if err := json.Unmarshal([]byte(schemaText), &schema); err != nil {
		t.Fatalf("generated schema is not JSON: %v\n%s", err, schemaText)
	}
	if schema.YaVersion != appInventorYaVersion {
		t.Fatalf("YaVersion = %q, want %q", schema.YaVersion, appInventorYaVersion)
	}
	if schema.Properties.Version != appInventorComponentVersion("Screen") {
		t.Fatalf("Form $Version = %q, want %q", schema.Properties.Version, appInventorComponentVersion("Screen"))
	}
	if len(schema.Properties.Components) != 1 || schema.Properties.Components[0].Version != appInventorComponentVersion("Button") {
		t.Fatalf("Button component version not current: %#v", schema.Properties.Components)
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
