package design

import (
	"encoding/json"
	"encoding/xml"
	"strconv"
)

type XmlParser struct {
	xmlContent  string
	autoIdCount map[string]int
}

func NewXmlParser(xmlContent string) *XmlParser {
	return &XmlParser{xmlContent: xmlContent, autoIdCount: make(map[string]int)}
}

func (p *XmlParser) ConvertXmlToSchema() (string, error) {
	var screen Component
	if err := xml.Unmarshal([]byte(p.xmlContent), &screen); err != nil {
		panic(err)
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
	// add Screen's properties here
	for k, v := range screen.Properties {
		props[k] = v
	}

	schema := map[string]interface{}{
		"authURL":    []interface{}{"ai2.appinventor.mit.edu"},
		"YaVersion":  "200",
		"Source":     "Form",
		"Properties": props,
	}
	jsonBytes, _ := json.MarshalIndent(schema, "", "  ")
	return string(jsonBytes), nil
}

func (p *XmlParser) componentToJson(component Component) interface{} {
	var children []interface{}
	for _, child := range component.Children {
		children = append(children, p.componentToJson(child))
	}
	compId := component.Id
	if compId == "" {
		// dynamically generate an Id
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
