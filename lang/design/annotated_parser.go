package design

import (
	"Falcon/code/jsonutil"
	"strconv"
	"strings"
	"unicode"
)

// AimlParser parses the Component.Id { key: value, Child { } } design format.
type AimlParser struct {
	source      []rune
	pos         int
	autoIdCount map[string]int
}

type AnnParseError struct {
	Message  string
	Position int
}

func (e *AnnParseError) Error() string {
	return e.Message + " at position " + strconv.Itoa(e.Position)
}

type AnnDiagnostic struct {
	Message  string
	Position int
	Length   int
}

type AnnDiagnosticListError struct {
	Message     string
	Raw         string
	Diagnostics []AnnDiagnostic
}

func (e *AnnDiagnosticListError) Error() string {
	return e.Raw
}

func NewAimlParser(source string) *AimlParser {
	return &AimlParser{
		source:      []rune(source),
		autoIdCount: make(map[string]int),
	}
}

func (p *AimlParser) ConvertAimlToSchema() (string, error) {
	screen, err := p.parseDocument()
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
		"$Version":    appInventorComponentVersion("Screen"),
		"$Components": components,
	}
	for k, v := range screen.Properties {
		props[k] = v
	}

	schema := map[string]interface{}{
		"authURL":    []interface{}{"ai2.appinventor.mit.edu"},
		"YaVersion":  appInventorYaVersion,
		"Source":     "Form",
		"Properties": props,
	}
	jsonBytes, err := jsonutil.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func (p *AimlParser) parseDocument() (Component, error) {
	p.skipWhitespace()
	component, err := p.parseComponent()
	if err != nil {
		return Component{}, err
	}
	p.skipWhitespace()
	if p.pos < len(p.source) {
		return Component{}, p.parseError("unexpected trailing input")
	}
	return component, nil
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
		"$Type":    dbCompType(component.Type),
		"$Version": appInventorComponentVersion(component.Type),
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
	if p.pos < len(p.source) && p.source[p.pos] == '@' {
		p.pos++
	}

	typeStart := p.pos
	typeName := p.readIdentifier()
	if typeName == "" {
		return Component{}, p.parseError("expected component type name")
	}

	comp := Component{
		Type:              typeName,
		Properties:        make(map[string]string),
		typePosition:      typeStart,
		propertyPositions: make(map[string]int),
		propertyLengths:   make(map[string]int),
	}

	p.skipWhitespace()
	if p.pos < len(p.source) && p.source[p.pos] == '.' {
		p.pos++
		idStart := p.pos
		comp.Id = p.readIdentifier()
		if comp.Id == "" {
			return Component{}, p.parseError("expected component id after " + typeName + ".")
		}
		comp.idPosition = idStart
		p.skipWhitespace()
	}

	if p.pos >= len(p.source) || p.source[p.pos] != '{' {
		if comp.Id != "" {
			return comp, nil
		}
		return Component{}, p.parseError("expected '{' after " + typeName)
	}
	p.pos++

	for {
		p.skipWhitespace()
		if p.pos >= len(p.source) {
			return Component{}, p.parseError("unexpected end of input, expected '}'")
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
			if p.pos < len(p.source) && p.source[p.pos] == '/' {
				p.pos++
				p.readIdentifier()
				p.skipWhitespace()
			}
			if p.pos < len(p.source) && p.source[p.pos] == ',' {
				p.pos++
			}
			continue
		}

		entryStart := p.pos
		keyStart := p.pos
		key := p.readIdentifier()
		if key == "" {
			return Component{}, p.parseError("unexpected character " + strconv.Quote(string(p.source[p.pos])))
		}
		p.skipWhitespace()
		if p.pos >= len(p.source) || p.source[p.pos] != ':' {
			p.pos = entryStart
			child, err := p.parseComponent()
			if err != nil {
				return Component{}, err
			}
			comp.Children = append(comp.Children, child)
			p.skipWhitespace()
			if p.pos < len(p.source) && p.source[p.pos] == '/' {
				p.pos++
				p.readIdentifier()
				p.skipWhitespace()
			}
			if p.pos < len(p.source) && p.source[p.pos] == ',' {
				p.pos++
			}
			continue
		}

		p.pos++
		p.skipWhitespace()

		value, err := p.readValue()
		if err != nil {
			return Component{}, err
		}
		value, err = normalizeAnnPropertyValue(key, value)
		if err != nil {
			return Component{}, p.parseError(err.Error())
		}

		if key == "id" {
			if comp.Id != "" {
				return Component{}, p.parseError("duplicate id for " + typeName)
			}
			comp.Id = value
			comp.idPosition = keyStart
		} else {
			comp.Properties[key] = value
			comp.propertyPositions[key] = keyStart
			comp.propertyLengths[key] = len([]rune(key))
		}

		p.skipWhitespace()
		if p.pos < len(p.source) && p.source[p.pos] == ',' {
			p.pos++
		}
	}

	return comp, nil
}

func (p *AimlParser) parseError(message string) error {
	return &AnnParseError{Message: message, Position: p.pos}
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
		return "", p.parseError("unexpected end of input reading value")
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

func normalizeAnnPropertyValue(propName, value string) (string, error) {
	if !isAnnColorProperty(propName) {
		return value, nil
	}
	colorInt, ok, err := annColorLiteralToIntString(value)
	if err != nil {
		return "", err
	}
	if ok {
		return colorInt, nil
	}
	return value, nil
}

func isAnnColorProperty(propName string) bool {
	return strings.Contains(strings.ToLower(propName), "color")
}

func annColorLiteralToIntString(value string) (string, bool, error) {
	argbHex, ok, err := annColorLiteralToARGBHex(value)
	if err != nil || !ok {
		return "", ok, err
	}
	color, err := strconv.ParseUint(argbHex, 16, 32)
	if err != nil {
		return "", true, invalidColorLiteral(value)
	}
	return strconv.FormatInt(int64(int32(color)), 10), true, nil
}

func annColorLiteralToARGBHex(value string) (string, bool, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return "", false, nil
	}
	lower := strings.ToLower(raw)

	if strings.HasPrefix(lower, "&h") || strings.HasPrefix(lower, "#x") {
		hex := raw[2:]
		if len(hex) != 8 || !isHexDigits(hex) {
			return "", true, invalidColorLiteral(value)
		}
		return strings.ToUpper(hex), true, nil
	}

	if !strings.HasPrefix(raw, "#") {
		return "", false, nil
	}

	hex := raw[1:]
	if len(hex) == 0 || !isHexDigits(hex) {
		return "", true, invalidColorLiteral(value)
	}

	switch len(hex) {
	case 3:
		// CSS #RGB shorthand, with an implicit opaque alpha channel.
		r := strings.Repeat(hex[0:1], 2)
		g := strings.Repeat(hex[1:2], 2)
		b := strings.Repeat(hex[2:3], 2)
		return strings.ToUpper("FF" + r + g + b), true, nil
	case 4:
		// CSS #RGBA shorthand. App Inventor persists colors as AARRGGBB.
		r := strings.Repeat(hex[0:1], 2)
		g := strings.Repeat(hex[1:2], 2)
		b := strings.Repeat(hex[2:3], 2)
		a := strings.Repeat(hex[3:4], 2)
		return strings.ToUpper(a + r + g + b), true, nil
	case 6:
		// CSS #RRGGBB, with an implicit opaque alpha channel.
		return strings.ToUpper("FF" + hex), true, nil
	case 8:
		// App Inventor's picker works with #RRGGBBAA, then serializes AARRGGBB.
		return strings.ToUpper(hex[6:8] + hex[0:6]), true, nil
	default:
		return "", true, invalidColorLiteral(value)
	}
}

func isHexDigits(value string) bool {
	for _, c := range value {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func (p *AimlParser) readString() (string, error) {
	p.pos++ // skip opening '"'
	var sb strings.Builder
	for p.pos < len(p.source) {
		c := p.source[p.pos]
		if c == '\\' {
			p.pos++
			if p.pos >= len(p.source) {
				return "", p.parseError("unexpected end of input after escape")
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
	return "", p.parseError("unterminated string literal")
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
