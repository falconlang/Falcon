package design

import (
	"bytes"
	"encoding/xml"
	"io"
	"sort"
	"strconv"
	"strings"
)

type XmlRoot struct {
	XMLName xml.Name  `xml:"xml"`
	XMLNS   string    `xml:"xmlns,attr"`
	Screen  Component `xml:"Screen"`
}

type Component struct {
	XMLName    xml.Name          `xml:""`
	Id         string            `xml:"id,attr,omitempty"`
	Type       string            `xml:"-"`
	Properties map[string]string `xml:"-"`
	Children   []Component       `xml:",any"`

	typePosition      int
	idPosition        int
	propertyPositions map[string]int
	propertyLengths   map[string]int
}

func (c *Component) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	c.XMLName = start.Name
	c.Type = start.Name.Local
	c.Properties = make(map[string]string)

	for _, attr := range start.Attr {
		if attr.Name.Local == "id" {
			c.Id = attr.Value
		} else {
			c.Properties[attr.Name.Local] = attr.Value
		}
	}
	for {
		tok, err := d.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			var child Component
			if err := d.DecodeElement(&child, &tok); err != nil {
				return err
			}
			c.Children = append(c.Children, child)
		case xml.EndElement:
			if tok.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
	return nil
}

// WriteAiml converts a Component to the Component.Id { key: value } format.
func (c *Component) WriteAiml(w io.Writer, indent int) error {
	indentStr := strings.Repeat("  ", indent)
	hasChildren := len(c.Children) > 0

	type kv struct{ k, v string }
	var props []kv
	keys := make([]string, 0, len(c.Properties))
	for k := range c.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		props = append(props, kv{k, c.Properties[k]})
	}

	var sb strings.Builder
	sb.WriteString(indentStr)
	sb.WriteString(c.Type)
	if c.Id != "" {
		sb.WriteString(".")
		sb.WriteString(c.Id)
	}

	if c.Id != "" && !hasChildren && len(props) == 0 {
		sb.WriteString("\n")
		_, err := w.Write([]byte(sb.String()))
		return err
	}

	sb.WriteString(" {")

	for i, p := range props {
		sb.WriteString(" ")
		sb.WriteString(p.k)
		sb.WriteString(": ")
		sb.WriteString(aimlFormatValue(p.v))
		if i < len(props)-1 || hasChildren {
			sb.WriteString(",")
		}
	}

	if !hasChildren {
		if len(props) > 0 {
			sb.WriteString(" }")
		} else {
			sb.WriteString("}")
		}
		sb.WriteString("\n")
		_, err := w.Write([]byte(sb.String()))
		return err
	}

	sb.WriteString("\n")
	if _, err := w.Write([]byte(sb.String())); err != nil {
		return err
	}
	for _, child := range c.Children {
		if err := child.WriteAiml(w, indent+1); err != nil {
			return err
		}
	}
	_, err := w.Write([]byte(indentStr + "}\n"))
	return err
}

func aimlFormatValue(v string) string {
	if v == "true" || v == "false" {
		return v
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return v
	}
	return `"` + aimlEscapeString(v) + `"`
}

func aimlEscapeString(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch c {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

// WriteXML manually converts Component structure to XML, a workaround for now, since
// Go lang does not support self-closing tags
func (c *Component) WriteXML(w io.Writer, indent int) error {
	indentStr := strings.Repeat("  ", indent)

	// Start tag
	tag := indentStr + "<" + c.Type
	if c.Id != "" {
		var buf bytes.Buffer
		err := xml.EscapeText(&buf, []byte(c.Id))
		if err != nil {
			return err
		}
		tag += ` id="` + buf.String() + `"`
	}

	for k, v := range c.Properties {
		if k == "id" || k == "type" {
			continue
		}
		var buf bytes.Buffer
		err := xml.EscapeText(&buf, []byte(v))
		if err != nil {
			return err
		}
		tag += ` ` + k + `="` + buf.String() + `"`
	}

	if len(c.Children) == 0 {
		tag += "/>\n"
		_, err := w.Write([]byte(tag))
		return err
	}

	// Open tag
	tag += ">\n"
	if _, err := w.Write([]byte(tag)); err != nil {
		return err
	}

	// Recursively write children
	for _, child := range c.Children {
		if err := child.WriteXML(w, indent+1); err != nil {
			return err
		}
	}

	// Close tag
	_, err := w.Write([]byte(indentStr + "</" + c.Type + ">\n"))
	return err
}
