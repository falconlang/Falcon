package ast

import (
	"strings"
	"testing"
)

func TestParseBlocklyXMLReadsBlockInlineAndComment(t *testing.T) {
	root, err := ParseBlocklyXML(`<xml xmlns="https://developers.google.com/blockly/xml">
  <block type="math_number" inline="true">
    <comment pinned="false" h="80" w="160">explain</comment>
    <field name="NUM">1</field>
  </block>
</xml>`)
	if err != nil {
		t.Fatalf("ParseBlocklyXML() error = %v", err)
	}
	if len(root.Blocks) != 1 {
		t.Fatalf("blocks = %d, want 1", len(root.Blocks))
	}
	if !root.Blocks[0].Inline {
		t.Fatal("block Inline = false, want true")
	}
	if got := root.Blocks[0].Comment; got != "explain" {
		t.Fatalf("block Comment = %q, want explain", got)
	}
}

func TestBlocklyXMLPreservesBlockInlineAndComment(t *testing.T) {
	root := XmlRoot{
		XMLNS: "https://developers.google.com/blockly/xml",
		Blocks: []Block{{
			Type:    "math_number",
			Inline:  true,
			Comment: "a < b",
			Fields:  []Field{{Name: "NUM", Value: "1"}},
		}},
	}

	xmlText := string(root.MarshalIndent("", "  "))

	if !strings.Contains(xmlText, `<block type="math_number" inline="true">`) {
		t.Fatalf("serialized XML = %s, want inline block", xmlText)
	}
	if !strings.Contains(xmlText, `<comment pinned="false" h="80" w="160">a &lt; b</comment>`) {
		t.Fatalf("serialized XML = %s, want escaped block comment", xmlText)
	}
}
