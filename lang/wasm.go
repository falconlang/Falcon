//go:build js && wasm
// +build js,wasm

// GOOS=js GOARCH=wasm go build -o web/falcon.wasm

package main

import (
	"Falcon/code/ast"
	"Falcon/code/compdb"
	"Falcon/code/context"
	"Falcon/code/lex"
	blocklyParser "Falcon/code/parsers/blocklytomist"
	blocklyYail "Falcon/code/parsers/blocklytoyail"
	"Falcon/code/parsers/mistparser"
	"Falcon/code/runtime"
	"Falcon/design"
	"encoding/json"
	"encoding/xml"
	"strings"
	"syscall/js"
)

func safeExec(fn func() js.Value) (ret js.Value) {
	ret = js.Undefined()

	defer func() {
		if r := recover(); r != nil {
			if msg, ok := r.(string); ok {
				js.Global().Call("mistError", msg)
			} else if err, ok := r.(error); ok {
				js.Global().Call("mistError", err.Error())
			} else {
				// last-resort: print the raw value
				js.Global().Call("mistError", r)
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
		langParser := mistparser.NewLangParser(true, tokens)
		langParser.SetComponentDefinitions(componentContextMap, reverseComponentMap)
		langParser.SetEventValidator(compdb.GlobalDB.ValidateEvent)
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
		exprs := blocklyParser.NewParser(xmlContent).GenerateAST()
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

func convertSchemaToAiml(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("No schema provided")
		}
		aimlString, err := design.NewSchemaParser(p[0].String()).ConvertSchemaToAiml()
		if err != nil {
			panic(err)
		}
		return js.ValueOf(aimlString)
	})
}

func convertAimlToSchema(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("No schema provided")
		}
		src := p[0].String()
		if err := design.ValidateAnnSource(src); err != nil {
			panic(err)
		}
		schemaString, err := design.NewAimlParser(src).ConvertAimlToSchema()
		if err != nil {
			panic(err)
		}
		return js.ValueOf(schemaString)
	})
}

// annToYail converts a .ann design source and pre-compiled code YAIL into the
// combined design+code YAIL block ready for the REPL (without the outer require wrapper).
func annToYail(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("annToYail(annSource [, codeYail]) not provided!")
		}
		annSource := p[0].String()
		codeYail := ""
		if len(p) >= 2 {
			codeYail = p[1].String()
		}
		yail, err := design.NewAnnYailConverter().ConvertAnnToReplYail(annSource, codeYail)
		if err != nil {
			panic(err)
		}
		return js.ValueOf(yail)
	})
}

// blocklyToYail compiles Blockly XML into YAIL for the App Inventor companion.
func blocklyToYail(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("blocklyToYail(xmlContent) not provided!")
		}
		yail := blocklyYail.NewParser(p[0].String()).GenerateYAIL()
		return js.ValueOf(yail)
	})
}

// describeComponent returns a JSON string with the full metadata for a component
// type (properties, methods, events). Returns an error string if unknown.
func describeComponent(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("describeComponent(componentName) not provided!")
		}
		name := p[0].String()
		desc, ok := compdb.GlobalDB.DescribeComponent(name)
		if !ok {
			return js.ValueOf("")
		}
		return js.ValueOf(desc)
	})
}

// listComponents returns a JSON array of all known component type names.
func listComponents(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		names := compdb.GlobalDB.ListComponentNames()
		data, err := json.Marshal(names)
		if err != nil {
			return js.ValueOf("[]")
		}
		return js.ValueOf(string(data))
	})
}

// runCode executes Falcon source code and streams each printed line to JS via
// the falconPrint(line) callback. Parse and runtime errors are sent to mistError(msg).
func runCode(this js.Value, p []js.Value) any {
	if len(p) < 1 {
		js.Global().Call("mistError", "runCode(sourceCode) not provided!")
		return js.Undefined()
	}
	sourceCode := p[0].String()

	var interp *runtime.Interpreter
	defer func() {
		if r := recover(); r != nil {
			var msg string
			if interp != nil {
				msg = interp.FormatRuntimeError(r)
			} else if s, ok := r.(string); ok {
				msg = s
			} else if err, ok := r.(error); ok {
				msg = err.Error()
			} else {
				msg = "unknown error"
			}
			js.Global().Call("mistError", msg)
		}
	}()

	codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: "wasm"}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := mistparser.NewLangParser(false, tokens)
	expressions := langParser.ParseAll()

	interp = runtime.NewInterpreterWithOutput(func(line string) {
		js.Global().Call("falconPrint", line)
	})
	interp.Run(expressions)
	return js.Undefined()
}

func main() {
	println("Hello from wasm.go!")

	c := make(chan struct{}, 0)
	js.Global().Set("mistToXml", js.FuncOf(mistToXml))
	js.Global().Set("xmlToMist", js.FuncOf(xmlToMist))
	js.Global().Set("schemaToAiml", js.FuncOf(convertSchemaToAiml))
	js.Global().Set("aimlToSchema", js.FuncOf(convertAimlToSchema))
	js.Global().Set("annToYail", js.FuncOf(annToYail))
	js.Global().Set("blocklyToYail", js.FuncOf(blocklyToYail))
	js.Global().Set("runCode", js.FuncOf(runCode))
	js.Global().Set("describeComponent", js.FuncOf(describeComponent))
	js.Global().Set("listComponents", js.FuncOf(listComponents))
	<-c
}
