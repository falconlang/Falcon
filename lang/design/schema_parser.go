package design

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
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
	if source, ok := jsonStruct["Source"]; ok {
		sourceText, ok := source.(string)
		if !ok || sourceText != "Form" {
			return Component{}, errors.New(`"Source" must be "Form"`)
		}
	}
	properties, err := objectField(jsonStruct, "Properties")
	if err != nil {
		return Component{}, err
	}
	if rootType, ok := properties["$Type"]; ok {
		rootTypeText, ok := rootType.(string)
		if !ok || rootTypeText != "Form" {
			return Component{}, errors.New(`root "$Type" must be "Form"`)
		}
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
		return nil, errors.New("$Components must be an array")
	}
	children := make([]Component, 0, len(componentList))
	for i, rawComponent := range componentList {
		componentMap, ok := rawComponent.(map[string]interface{})
		if !ok {
			return nil, errors.New("$Components[" + strconv.Itoa(i) + "] must be an object")
		}
		child, err := schemaComponentToXml(componentMap)
		if err != nil {
			return nil, wrapError("$Components["+strconv.Itoa(i)+"]", err)
		}
		children = append(children, child)
	}
	return children, nil
}

func filterDesignerProperties(componentProps map[string]interface{}) (map[string]string, error) {
	filteredProperties := make(map[string]string)
	for key, value := range componentProps {
		if !shouldSendDesignerProperty(key) {
			continue
		}
		text, err := schemaScalarToString(value)
		if err != nil {
			return nil, wrapError("property "+strconv.Quote(key), err)
		}
		filteredProperties[key] = text
	}
	return filteredProperties, nil
}

func objectField(parent map[string]interface{}, name string) (map[string]interface{}, error) {
	value, ok := parent[name]
	if !ok {
		return nil, errors.New("missing " + strconv.Quote(name))
	}
	child, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.New(strconv.Quote(name) + " must be an object")
	}
	return child, nil
}

func requiredStringField(parent map[string]interface{}, name string) (string, error) {
	value, ok := parent[name]
	if !ok {
		return "", errors.New("missing " + strconv.Quote(name))
	}
	text, ok := value.(string)
	if !ok {
		return "", errors.New(strconv.Quote(name) + " must be a string")
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
		return "", errors.New("expected string, number, boolean, or null")
	}
}
