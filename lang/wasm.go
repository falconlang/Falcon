//go:build js && wasm
// +build js,wasm

// GOOS=js GOARCH=wasm go build -o web/falcon.wasm

package main

import (
	"Falcon/code/ast"
	"Falcon/code/context"
	"Falcon/code/lex"
	"Falcon/code/parser"
	"Falcon/design"
	"encoding/xml"
	"strings"
	"syscall/js"
)

func safeExec(fn func() js.Value) (ret js.Value) {
	ret = js.Undefined()

	defer func() {
		if r := recover(); r != nil {
			if msg, ok := r.(string); ok {
				js.Global().Get("console").Call("error", msg)
			} else if err, ok := r.(error); ok {
				js.Global().Get("console").Call("error", err.Error())
			} else {
				// last-resort: print the raw value
				js.Global().Get("console").Call("error", r)
			}
		}
	}()

	ret = fn()
	return
}

// Code -> Blocks
func mistToXml(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 2 {
			return js.ValueOf("mistToXML(sourceCode string, componentDefinitions map[string][]string) not provided!")
		}
		sourceCode := p[0].String()

		// Parse the Component Definition Context
		componentContextMap := make(map[string][]string) // Button -> [Button1, Button2]
		reverseComponentMap := make(map[string]string)   // Button1 -> Button, Button2 -> Button
		obj := p[1]
		keys := js.Global().Get("Object").Call("keys", obj)
		length := keys.Length()
		for i := 0; i < length; i++ {
			compType := keys.Index(i).String()
			jsArr := obj.Get(compType)
			var compNames []string
			for j := 0; j < jsArr.Length(); j++ {
				instanceName := jsArr.Index(j).String()
				compNames = append(compNames, instanceName)
				reverseComponentMap[instanceName] = compType
			}
			componentContextMap[compType] = compNames
		}

		// Parse Mist To XML Blockly
		codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: "appinventor.live"}

		tokens := lex.NewLexer(codeContext).Lex()
		langParser := parser.NewLangParser(true, tokens)
		langParser.SetComponentDefinitions(componentContextMap, reverseComponentMap)
		expressions := langParser.ParseAll()

		var xmlCode strings.Builder

		for _, expression := range expressions {
			xmlBlock := ast.XmlRoot{
				Blocks: []ast.Block{expression.Blockly(true)},
				XMLNS:  "https://developers.google.com/blockly/xml",
			}
			bytes, _ := xml.MarshalIndent(xmlBlock, "", "  ")

			xmlCode.WriteString(string(bytes))
			xmlCode.WriteByte(0)
		}

		return js.ValueOf(xmlCode.String())
	})
}

// Blocks -> Code
func xmlToMist(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("No XML content provided")
		}
		xmlContent := p[0].String()
		exprs := parser.NewXMLParser(xmlContent).ParseBlockly()
		var builder strings.Builder

		for _, expr := range exprs {
			builder.WriteString(expr.String())
			builder.WriteString("\n")

			block := expr.Blockly(true)
			if block.Order() > 0 {
				builder.WriteString("\n")
			}
		}
		return js.ValueOf(builder.String())
	})
}

func convertSchemaToXml(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("No schema provided")
		}
		schemaString, err := design.NewSchemaParser(p[0].String()).ConvertSchemaToXml()
		if err != nil {
			panic(err)
		}
		return js.ValueOf(schemaString)
	})
}

func convertXmlToSchema(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("No schema provided")
		}
		schemaString, err := design.NewXmlParser(p[0].String()).ConvertXmlToSchema()
		if err != nil {
			panic(err)
		}
		return js.ValueOf(schemaString)
	})
}

func main() {
	println("Hello from wasm.go!")

	c := make(chan struct{}, 0)
	js.Global().Set("mistToXml", js.FuncOf(mistToXml))
	js.Global().Set("xmlToMist", js.FuncOf(xmlToMist))
	js.Global().Set("schemaToXml", js.FuncOf(convertSchemaToXml))
	js.Global().Set("xmlToSchema", js.FuncOf(convertXmlToSchema))
	<-c
}
