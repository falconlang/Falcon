package design

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
)

type SchemaParser struct {
	schemaJson string
}

func NewSchemaParser(schemaJson string) *SchemaParser {
	return &SchemaParser{schemaJson: schemaJson}
}

func (p *SchemaParser) ConvertSchemaToAiml() (string, error) {
	var jsonStruct map[string]interface{}
	err := json.Unmarshal([]byte(p.schemaJson), &jsonStruct)
	if err != nil {
		return "", err
	}
	properties := jsonStruct["Properties"].(map[string]interface{})
	screenId := properties["$Name"].(string)

	schemaComponents := properties["$Components"].([]interface{})
	var children []Component
	for _, schemaComponent := range schemaComponents {
		children = append(children, schemaComponentToXml(schemaComponent.(map[string]interface{})))
	}

	root := Component{
		Id:         screenId,
		Type:       "Screen",
		Properties: filterDesignerProperties(properties),
		Children:   children,
	}
	var buf bytes.Buffer
	err = root.WriteAiml(&buf, 0)
	return buf.String(), err
}

func (p *SchemaParser) ConvertSchemaToXml() (string, error) {
	var jsonStruct map[string]interface{}
	err := json.Unmarshal([]byte(p.schemaJson), &jsonStruct)
	if err != nil {
		return "", err
	}
	properties := jsonStruct["Properties"].(map[string]interface{})
	screenId := properties["$Name"].(string)

	schemaComponents := properties["$Components"].([]interface{})
	var xmlChildren []Component
	for _, schemaComponent := range schemaComponents {
		xmlChildren = append(xmlChildren, schemaComponentToXml(schemaComponent.(map[string]interface{})))
	}

	root := Component{
		XMLName:    xml.Name{Local: "Screen"},
		Id:         screenId,
		Type:       "Screen",
		Properties: filterDesignerProperties(properties),
		Children:   xmlChildren,
	}
	var buf bytes.Buffer
	err = root.WriteXML(&buf, 0)
	return buf.String(), err
}

func schemaComponentToXml(schemaJson map[string]interface{}) Component {
	compType := schemaJson["$Type"].(string)
	schemaComponents := schemaJson["$Components"]
	var xmlChildren []Component
	if schemaComponents != nil {
		for _, schemaComponent := range schemaComponents.([]interface{}) {
			xmlChildren = append(xmlChildren, schemaComponentToXml(schemaComponent.(map[string]interface{})))
		}
	}
	return Component{
		XMLName:    xml.Name{Local: compType},
		Id:         schemaJson["$Name"].(string),
		Type:       compType,
		Properties: filterDesignerProperties(schemaJson),
		Children:   xmlChildren,
	}
}

func filterDesignerProperties(componentProps map[string]interface{}) map[string]string {
	filteredProperties := make(map[string]string)
	for key, value := range componentProps {
		if len(key) > 0 && key[0] == '$' {
			continue
		}
		filteredProperties[key] = value.(string)
	}
	return filteredProperties
}
