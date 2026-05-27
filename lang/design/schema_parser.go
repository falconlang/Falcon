package design

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
)

type SchemaParser struct {
	schemaJson string
}

func NewSchemaParser(schemaJson string) *SchemaParser {
	return &SchemaParser{schemaJson: schemaJson}
}

func (p *SchemaParser) ConvertSchemaToAiml() (string, error) {
	root, err := p.parseSchemaRoot()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = root.WriteAiml(&buf, 0)
	return buf.String(), err
}

func (p *SchemaParser) ConvertSchemaToXml() (string, error) {
	root, err := p.parseSchemaRoot()
	if err != nil {
		return "", err
	}
	root.XMLName = xml.Name{Local: "Screen"}

	var buf bytes.Buffer
	err = root.WriteXML(&buf, 0)
	return buf.String(), err
}

func (p *SchemaParser) parseSchemaRoot() (Component, error) {
	var jsonStruct map[string]interface{}
	if err := json.Unmarshal([]byte(p.schemaJson), &jsonStruct); err != nil {
		return Component{}, err
	}
	properties, err := objectField(jsonStruct, "Properties")
	if err != nil {
		return Component{}, err
	}
	screenId, err := requiredStringField(properties, "$Name")
	if err != nil {
		return Component{}, err
	}

	children, err := schemaChildren(properties)
	if err != nil {
		return Component{}, err
	}
	props, err := filterDesignerProperties(properties)
	if err != nil {
		return Component{}, err
	}
	return Component{
		Id:         screenId,
		Type:       "Screen",
		Properties: props,
		Children:   children,
	}, nil
}

func schemaComponentToXml(schemaJson map[string]interface{}) (Component, error) {
	compType, err := requiredStringField(schemaJson, "$Type")
	if err != nil {
		return Component{}, err
	}
	name, err := requiredStringField(schemaJson, "$Name")
	if err != nil {
		return Component{}, err
	}
	properties, err := filterDesignerProperties(schemaJson)
	if err != nil {
		return Component{}, err
	}

	var xmlChildren []Component
	children, err := schemaChildren(schemaJson)
	if err != nil {
		return Component{}, err
	}
	xmlChildren = append(xmlChildren, children...)

	return Component{
		XMLName:    xml.Name{Local: compType},
		Id:         name,
		Type:       compType,
		Properties: properties,
		Children:   xmlChildren,
	}, nil
}

func schemaChildren(componentProps map[string]interface{}) ([]Component, error) {
	rawComponents, ok := componentProps["$Components"]
	if !ok || rawComponents == nil {
		return nil, nil
	}
	componentList, ok := rawComponents.([]interface{})
	if !ok {
		return nil, fmt.Errorf("$Components must be an array")
	}
	children := make([]Component, 0, len(componentList))
	for i, rawComponent := range componentList {
		componentMap, ok := rawComponent.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("$Components[%d] must be an object", i)
		}
		child, err := schemaComponentToXml(componentMap)
		if err != nil {
			return nil, fmt.Errorf("$Components[%d]: %w", i, err)
		}
		children = append(children, child)
	}
	return children, nil
}

func filterDesignerProperties(componentProps map[string]interface{}) (map[string]string, error) {
	filteredProperties := make(map[string]string)
	for key, value := range componentProps {
		if len(key) > 0 && key[0] == '$' {
			continue
		}
		text, err := schemaScalarToString(value)
		if err != nil {
			return nil, fmt.Errorf("property %q: %w", key, err)
		}
		filteredProperties[key] = text
	}
	return filteredProperties, nil
}

func objectField(parent map[string]interface{}, name string) (map[string]interface{}, error) {
	value, ok := parent[name]
	if !ok {
		return nil, fmt.Errorf("missing %q", name)
	}
	child, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%q must be an object", name)
	}
	return child, nil
}

func requiredStringField(parent map[string]interface{}, name string) (string, error) {
	value, ok := parent[name]
	if !ok {
		return "", fmt.Errorf("missing %q", name)
	}
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%q must be a string", name)
	}
	return text, nil
}

func schemaScalarToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("expected string, number, boolean, or null")
	}
}
