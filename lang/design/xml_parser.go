package design

import (
	"Falcon/code/jsonutil"
	"errors"
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
	screen, err := ParseDesignXML(p.xmlContent)
	if err != nil {
		return "", err
	}
	if screen.Type != "Screen" {
		return "", errors.New("root XML element must be Screen")
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

func (p *XmlParser) componentToJson(component Component) interface{} {
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
