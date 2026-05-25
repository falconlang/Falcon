package design

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// AimlParser parses the @Component { key: value, @Child { } } design format.
type AimlParser struct {
	source      []rune
	pos         int
	autoIdCount map[string]int
}

func NewAimlParser(source string) *AimlParser {
	return &AimlParser{
		source:      []rune(source),
		autoIdCount: make(map[string]int),
	}
}

func (p *AimlParser) ConvertAimlToSchema() (string, error) {
	p.skipWhitespace()
	screen, err := p.parseComponent()
	if err != nil {
		return "", err
	}

	var components []interface{}
	for _, child := range screen.Children {
		components = append(components, p.componentToJson(child))
	}

	props := map[string]interface{}{
		"$Name":       screen.Id,
		"$Type":       "Form",
		"$Version":    "31",
		"$Components": components,
	}
	for k, v := range screen.Properties {
		props[k] = v
	}

	schema := map[string]interface{}{
		"authURL":    []interface{}{"ai2.appinventor.mit.edu"},
		"YaVersion":  "200",
		"Source":     "Form",
		"Properties": props,
	}
	jsonBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (p *AimlParser) componentToJson(component Component) interface{} {
	var children []interface{}
	for _, child := range component.Children {
		children = append(children, p.componentToJson(child))
	}
	compId := component.Id
	if compId == "" {
		compId = component.Type + strconv.Itoa(p.autoIdCount[component.Type]+1)
		p.autoIdCount[component.Type]++
	}
	schema := map[string]interface{}{
		"$Name":    compId,
		"$Type":    component.Type,
		"$Version": "32",
	}
	if len(children) > 0 {
		schema["$Components"] = children
	}
	for k, v := range component.Properties {
		schema[k] = v
	}
	return schema
}

func (p *AimlParser) parseComponent() (Component, error) {
	if p.pos >= len(p.source) || p.source[p.pos] != '@' {
		return Component{}, fmt.Errorf("expected '@' at position %d", p.pos)
	}
	p.pos++

	typeName := p.readIdentifier()
	if typeName == "" {
		return Component{}, fmt.Errorf("expected component type name at position %d", p.pos)
	}

	p.skipWhitespace()
	if p.pos >= len(p.source) || p.source[p.pos] != '{' {
		return Component{}, fmt.Errorf("expected '{' after @%s at position %d", typeName, p.pos)
	}
	p.pos++

	comp := Component{
		Type:       typeName,
		Properties: make(map[string]string),
	}

	for {
		p.skipWhitespace()
		if p.pos >= len(p.source) {
			return Component{}, fmt.Errorf("unexpected end of input, expected '}'")
		}
		if p.source[p.pos] == '}' {
			p.pos++
			break
		}
		if p.source[p.pos] == '@' {
			child, err := p.parseComponent()
			if err != nil {
				return Component{}, err
			}
			comp.Children = append(comp.Children, child)
			p.skipWhitespace()
			if p.pos < len(p.source) && p.source[p.pos] == ',' {
				p.pos++
			}
			continue
		}
		// Parse property key: value
		key := p.readIdentifier()
		if key == "" {
			return Component{}, fmt.Errorf("unexpected character %q at position %d", string(p.source[p.pos]), p.pos)
		}
		p.skipWhitespace()
		if p.pos >= len(p.source) || p.source[p.pos] != ':' {
			return Component{}, fmt.Errorf("expected ':' after key %q", key)
		}
		p.pos++
		p.skipWhitespace()

		value, err := p.readValue()
		if err != nil {
			return Component{}, err
		}

		if key == "id" {
			comp.Id = value
		} else {
			comp.Properties[key] = value
		}

		p.skipWhitespace()
		if p.pos < len(p.source) && p.source[p.pos] == ',' {
			p.pos++
		}
	}

	return comp, nil
}

func (p *AimlParser) readIdentifier() string {
	start := p.pos
	for p.pos < len(p.source) {
		c := p.source[p.pos]
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			p.pos++
		} else {
			break
		}
	}
	return string(p.source[start:p.pos])
}

func (p *AimlParser) readValue() (string, error) {
	if p.pos >= len(p.source) {
		return "", fmt.Errorf("unexpected end of input reading value")
	}
	if p.source[p.pos] == '"' {
		return p.readString()
	}
	// Unquoted value: read until comma, }, or whitespace
	start := p.pos
	for p.pos < len(p.source) {
		c := p.source[p.pos]
		if c == ',' || c == '}' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(string(p.source[start:p.pos])), nil
}

func (p *AimlParser) readString() (string, error) {
	p.pos++ // skip opening '"'
	var sb strings.Builder
	for p.pos < len(p.source) {
		c := p.source[p.pos]
		if c == '\\' {
			p.pos++
			if p.pos >= len(p.source) {
				return "", fmt.Errorf("unexpected end of input after escape")
			}
			switch p.source[p.pos] {
			case '"':
				sb.WriteRune('"')
			case '\\':
				sb.WriteRune('\\')
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(p.source[p.pos])
			}
			p.pos++
		} else if c == '"' {
			p.pos++
			return sb.String(), nil
		} else {
			sb.WriteRune(c)
			p.pos++
		}
	}
	return "", fmt.Errorf("unterminated string literal")
}

func (p *AimlParser) skipWhitespace() {
	for p.pos < len(p.source) {
		c := p.source[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.pos++
		} else {
			break
		}
	}
}
