package design

import (
	"Falcon/code/xmlutil"
	"io"
	"sort"
	"strconv"
	"strings"
)

type Component struct {
	Id         string
	Type       string
	Properties map[string]string
	Children   []Component

	// source position tracking for annotated design parser diagnostics
	typePosition      int
	idPosition        int
	propertyPositions map[string]int
	propertyLengths   map[string]int
}

// ParseDesignXML parses an App Inventor design XML string into a Component tree.
func ParseDesignXML(src string) (Component, error) {
	doc, err := xmlutil.ParseDocument(src)
	if err != nil {
		return Component{}, err
	}
	return elementToComponent(doc), nil
}

func elementToComponent(e *xmlutil.Element) Component {
	c := Component{
		Type:       e.Name,
		Id:         e.AttrVal("id"),
		Properties: make(map[string]string),
	}
	for _, attr := range e.Attrs {
		if attr.Name != "id" {
			c.Properties[attr.Name] = attr.Value
		}
	}
	for _, child := range e.Children {
		c.Children = append(c.Children, elementToComponent(child))
	}
	return c
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

// WriteXML converts Component structure to XML.
func (c *Component) WriteXML(w io.Writer, indent int) error {
	indentStr := strings.Repeat("  ", indent)

	tag := indentStr + "<" + c.Type
	if c.Id != "" {
		tag += ` id="` + xmlutil.EscapeText(c.Id) + `"`
	}

	keys := make([]string, 0, len(c.Properties))
	for k := range c.Properties {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := c.Properties[k]
		if k == "id" || k == "type" {
			continue
		}
		tag += ` ` + k + `="` + xmlutil.EscapeText(v) + `"`
	}

	if len(c.Children) == 0 {
		tag += "/>\n"
		_, err := w.Write([]byte(tag))
		return err
	}

	tag += ">\n"
	if _, err := w.Write([]byte(tag)); err != nil {
		return err
	}

	for _, child := range c.Children {
		if err := child.WriteXML(w, indent+1); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte(indentStr + "</" + c.Type + ">\n"))
	return err
}
