//go:build js && wasm
// +build js,wasm

// GOOS=js GOARCH=wasm go build -o web/falcon.wasm

package main

import (
	"Falcon/code/ast"
	astcommon "Falcon/code/ast/common"
	astmethod "Falcon/code/ast/method"
	"Falcon/code/compdb"
	"Falcon/code/context"
	"Falcon/code/jsonutil"
	"Falcon/code/lex"
	blocklyParser "Falcon/code/parsers/blocklytomist"
	blocklyYail "Falcon/code/parsers/blocklytoyail"
	"Falcon/code/parsers/mistparser"
	"Falcon/code/runtime"
	"Falcon/design"
	"errors"
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

func panicMessage(r any) string {
	switch v := r.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return "unknown error"
	}
}

func diagnosticsForRecover(r any, phase, fileName string) (string, []wasmDiagnostic) {
	raw := panicMessage(r)
	switch v := r.(type) {
	case *context.DiagnosticError:
		return raw, wasmDiagnosticsFromContext([]context.Diagnostic{v.Diagnostic}, phase, raw, fileName)
	case *context.DiagnosticListError:
		return raw, wasmDiagnosticsFromContext(v.Diagnostics, phase, raw, fileName)
	}
	return raw, fallbackDiagnostic(raw, phase, fileName)
}

func jsDiagnostics(diagnostics []wasmDiagnostic) []any {
	values := make([]any, len(diagnostics))
	for i, diagnostic := range diagnostics {
		values[i] = map[string]any{
			"message":  diagnostic.Message,
			"severity": diagnostic.Severity,
			"phase":    diagnostic.Phase,
			"file":     diagnostic.File,
			"line":     diagnostic.Line,
			"column":   diagnostic.Column,
			"length":   diagnostic.Length,
			"raw":      diagnostic.Raw,
		}
	}
	return values
}

func diagnosticResult(ok bool, phase, fileName, raw string, diagnostics []wasmDiagnostic, fields map[string]any) js.Value {
	if diagnostics == nil && raw != "" {
		diagnostics = fallbackDiagnostic(raw, phase, fileName)
	}
	result := map[string]any{
		"ok":          ok,
		"error":       raw,
		"diagnostics": jsDiagnostics(diagnostics),
	}
	for k, v := range fields {
		result[k] = v
	}
	return js.ValueOf(result)
}

func parseMistToXmlStrictArg(value js.Value) bool {
	if value.IsUndefined() || value.IsNull() {
		return true
	}

	switch value.Type() {
	case js.TypeBoolean:
		return value.Bool()
	case js.TypeObject:
		strict := value.Get("strict")
		if strict.Type() == js.TypeBoolean {
			return strict.Bool()
		}

		allowUnsafeParsing := value.Get("allowUnsafeParsing")
		if allowUnsafeParsing.Type() == js.TypeBoolean {
			return !allowUnsafeParsing.Bool()
		}
	}

	return true
}

func componentDefinitionsFromJS(obj js.Value) (map[string][]string, map[string]string) {
	componentContextMap := make(map[string][]string) // Button -> [Button1, Button2]
	reverseComponentMap := make(map[string]string)   // Button1 -> Button, Button2 -> Button
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
	return componentContextMap, reverseComponentMap
}

func newWasmLangParser(strict bool, tokens []*lex.Token) *mistparser.LangParser {
	langParser := mistparser.NewLangParser(strict, tokens)
	langParser.SetEventValidator(compdb.GlobalDB.ValidateEvent)
	langParser.SetPropertyValidator(compdb.GlobalDB.ValidateProperty)
	langParser.SetMethodValidator(compdb.GlobalDB.ValidateMethod)
	return langParser
}

func compileMistToXml(sourceCode string, componentDefinitions js.Value, strict bool) (string, []int) {
	componentContextMap, reverseComponentMap := componentDefinitionsFromJS(componentDefinitions)

	codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: "appinventor.live"}

	tokens := lex.NewLexer(codeContext).Lex()
	langParser := newWasmLangParser(strict, tokens)
	langParser.SetComponentDefinitions(componentContextMap, reverseComponentMap)
	expressions, lineNumbers := langParser.ParseTopLevel()

	var xmlCode strings.Builder

	for _, expression := range expressions {
		xmlBlock := ast.XmlRoot{
			Blocks: []ast.Block{expression.Blockly(true)},
			XMLNS:  "https://developers.google.com/blockly/xml",
		}
		xmlCode.Write(xmlBlock.MarshalIndent("", "  "))
		xmlCode.WriteByte(0)
	}

	return xmlCode.String(), lineNumbers
}

func jsLineNumbers(lineNumbers []int) []any {
	values := make([]any, len(lineNumbers))
	for i, lineNumber := range lineNumbers {
		values[i] = lineNumber
	}
	return values
}

// Code -> Blocks
func mistToXml(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 2 {
			return js.ValueOf("mistToXML(sourceCode string, componentDefinitions map[string][]string [, strict bool | { strict?: bool, allowUnsafeParsing?: bool }]) not provided!")
		}
		sourceCode := p[0].String()
		strict := true
		if len(p) >= 3 {
			strict = parseMistToXmlStrictArg(p[2])
		}

		xmlCode, lineNumbers := compileMistToXml(sourceCode, p[1], strict)
		return js.ValueOf(map[string]any{
			"xml":         xmlCode,
			"lineNumbers": jsLineNumbers(lineNumbers),
		})
	})
}

func mistToXmlWithDiagnostics(this js.Value, p []js.Value) any {
	sourceCode := ""
	if len(p) >= 1 {
		sourceCode = p[0].String()
	}
	if len(p) < 2 {
		msg := "mistToXML(sourceCode string, componentDefinitions map[string][]string [, strict bool | { strict?: bool, allowUnsafeParsing?: bool }]) not provided!"
		return diagnosticResult(false, "compile", "appinventor.live", msg, nil, map[string]any{
			"xml":         "",
			"lineNumbers": []any{},
		})
	}

	strict := true
	if len(p) >= 3 {
		strict = parseMistToXmlStrictArg(p[2])
	}

	var xmlCode string
	var lineNumbers []int
	var raw string
	var diagnostics []wasmDiagnostic
	ok := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				raw, diagnostics = diagnosticsForRecover(r, "compile", "appinventor.live")
				ok = false
			}
		}()
		xmlCode, lineNumbers = compileMistToXml(sourceCode, p[1], strict)
	}()

	return diagnosticResult(ok, "compile", "appinventor.live", raw, diagnostics, map[string]any{
		"xml":         xmlCode,
		"lineNumbers": jsLineNumbers(lineNumbers),
	})
}

func getComponentDefinitionsCode(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("getComponentDefinitionsCode(sourceCode string) not provided!")
		}
		sourceCode := p[0].String()
		codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: "appinventor.live"}
		tokens := lex.NewLexer(codeContext).Lex()
		langParser := mistparser.NewLangParser(false, tokens)
		langParser.ParseDefinitions()
		return js.ValueOf(langParser.GetComponentDefinitionsCode())
	})
}

// Blocks -> Code
func xmlToMist(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("No XML content provided")
		}
		xmlContent := p[0].String()
		exprs, err := blocklyParser.NewParser(xmlContent).TryGenerateAST()
		if err != nil {
			panic(err)
		}
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

func xmlToMistWithDiagnostics(this js.Value, p []js.Value) any {
	if len(p) < 1 {
		msg := "No XML content provided"
		return diagnosticResult(false, "blocks", "Screen1.bky", msg, nil, map[string]any{"source": ""})
	}

	xmlContent := p[0].String()
	var source string
	var raw string
	var diagnostics []wasmDiagnostic
	ok := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				raw, diagnostics = diagnosticsForRecover(r, "blocks", "Screen1.bky")
				ok = false
			}
		}()

		exprs, err := blocklyParser.NewParser(xmlContent).TryGenerateAST()
		if err != nil {
			raw = err.Error()
			diagnostics = fallbackDiagnostic(raw, "blocks", "Screen1.bky")
			ok = false
			return
		}

		var builder strings.Builder
		for _, expr := range exprs {
			builder.WriteString(expr.String())
			builder.WriteString("\n")

			block := expr.Blockly(true)
			if block.Order() > 0 {
				builder.WriteString("\n")
			}
		}
		source = builder.String()
	}()

	return diagnosticResult(ok, "blocks", "Screen1.bky", raw, diagnostics, map[string]any{"source": source})
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

func annToYailWithDiagnostics(this js.Value, p []js.Value) any {
	annSource := ""
	if len(p) >= 1 {
		annSource = p[0].String()
	}
	if len(p) < 1 {
		msg := "annToYail(annSource [, codeYail]) not provided!"
		return diagnosticResult(false, "design", "Screen1.design", msg, nil, map[string]any{"yail": ""})
	}
	codeYail := ""
	if len(p) >= 2 {
		codeYail = p[1].String()
	}

	var yail string
	var raw string
	var diagnostics []wasmDiagnostic
	ok := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				raw, diagnostics = diagnosticsForRecover(r, "design", "Screen1.design")
				ok = false
			}
		}()
		var err error
		yail, err = design.NewAnnYailConverter().ConvertAnnToReplYail(annSource, codeYail)
		if err != nil {
			raw = err.Error()
			var diagnosticListErr *design.AnnDiagnosticListError
			var parseErr *design.AnnParseError
			if errors.As(err, &diagnosticListErr) {
				diagnostics = wasmDiagnosticsFromAnn(annSource, diagnosticListErr.Diagnostics, "design", raw, "Screen1.design")
			} else if errors.As(err, &parseErr) {
				line, column := lineColumnForRuneIndex(annSource, parseErr.Position)
				diagnostics = []wasmDiagnostic{{
					Message:  parseErr.Message,
					Severity: "error",
					Phase:    "design",
					File:     "Screen1.design",
					Line:     line,
					Column:   column,
					Length:   1,
					Raw:      raw,
				}}
			}
			ok = false
		}
	}()

	return diagnosticResult(ok, "design", "Screen1.design", raw, diagnostics, map[string]any{"yail": yail})
}

// blocklyToYail compiles Blockly XML into YAIL for the App Inventor companion.
func blocklyToYail(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		if len(p) < 1 {
			return js.ValueOf("blocklyToYail(xmlContent) not provided!")
		}
		yail, err := blocklyYail.NewParser(p[0].String()).TryGenerateYAIL()
		if err != nil {
			panic(err)
		}
		return js.ValueOf(yail)
	})
}

func blocklyToYailWithDiagnostics(this js.Value, p []js.Value) any {
	if len(p) < 1 {
		msg := "blocklyToYail(xmlContent) not provided!"
		return diagnosticResult(false, "blocks", "Screen1.bky", msg, nil, map[string]any{"yail": ""})
	}

	var yail string
	var raw string
	var diagnostics []wasmDiagnostic
	ok := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				raw, diagnostics = diagnosticsForRecover(r, "blocks", "Screen1.bky")
				ok = false
			}
		}()

		var err error
		yail, err = blocklyYail.NewParser(p[0].String()).TryGenerateYAIL()
		if err != nil {
			raw = err.Error()
			diagnostics = fallbackDiagnostic(raw, "blocks", "Screen1.bky")
			ok = false
		}
	}()

	return diagnosticResult(ok, "blocks", "Screen1.bky", raw, diagnostics, map[string]any{"yail": yail})
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
		namesAny := make([]interface{}, len(names))
		for i, n := range names {
			namesAny[i] = n
		}
		data, err := jsonutil.Marshal(namesAny)
		if err != nil {
			return js.ValueOf("[]")
		}
		return js.ValueOf(string(data))
	})
}

func falconCompletionCatalog(this js.Value, p []js.Value) any {
	return safeExec(func() js.Value {
		fns := astcommon.ListFunctionCompletions()
		ms := astmethod.ListMethodCompletions()
		fnsAny := make([]interface{}, len(fns))
		for i, v := range fns {
			fnsAny[i] = v
		}
		msAny := make([]interface{}, len(ms))
		for i, v := range ms {
			msAny[i] = v
		}
		data, err := jsonutil.Marshal(map[string]interface{}{
			"functions": fnsAny,
			"methods":   msAny,
		})
		if err != nil {
			return js.ValueOf(`{"functions":[],"methods":[]}`)
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
	langParser := newWasmLangParser(true, tokens)
	expressions := langParser.ParseAll()

	interp = runtime.NewInterpreterWithOutput(func(line string) {
		js.Global().Call("falconPrint", line)
	})
	interp.Run(expressions)
	return js.Undefined()
}

func runCodeWithDiagnostics(this js.Value, p []js.Value) any {
	if len(p) < 1 {
		msg := "runCode(sourceCode) not provided!"
		return diagnosticResult(false, "runtime", "wasm", msg, nil, map[string]any{})
	}
	sourceCode := p[0].String()

	var interp *runtime.Interpreter
	var raw string
	var diagnostics []wasmDiagnostic
	ok := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				if interp != nil {
					raw = interp.FormatRuntimeError(r)
					diagnostics = wasmDiagnosticsFromContext(interp.DiagnosticFromRuntimeError(r), "runtime", raw, "wasm")
				} else {
					raw, diagnostics = diagnosticsForRecover(r, "compile", "wasm")
				}
				ok = false
			}
		}()

		codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: "wasm"}
		tokens := lex.NewLexer(codeContext).Lex()
		langParser := newWasmLangParser(true, tokens)
		expressions := langParser.ParseAll()

		interp = runtime.NewInterpreterWithOutput(func(line string) {
			js.Global().Call("falconPrint", line)
		})
		interp.Run(expressions)
	}()

	return diagnosticResult(ok, "runtime", "wasm", raw, diagnostics, map[string]any{})
}

func main() {
	println("Hello from wasm.go!")

	c := make(chan struct{}, 0)
	js.Global().Set("mistToXml", js.FuncOf(mistToXml))
	js.Global().Set("mistToXmlWithDiagnostics", js.FuncOf(mistToXmlWithDiagnostics))
	js.Global().Set("getComponentDefinitionsCode", js.FuncOf(getComponentDefinitionsCode))
	js.Global().Set("xmlToMist", js.FuncOf(xmlToMist))
	js.Global().Set("xmlToMistWithDiagnostics", js.FuncOf(xmlToMistWithDiagnostics))
	js.Global().Set("schemaToAiml", js.FuncOf(convertSchemaToAiml))
	js.Global().Set("aimlToSchema", js.FuncOf(convertAimlToSchema))
	js.Global().Set("annToYail", js.FuncOf(annToYail))
	js.Global().Set("annToYailWithDiagnostics", js.FuncOf(annToYailWithDiagnostics))
	js.Global().Set("blocklyToYail", js.FuncOf(blocklyToYail))
	js.Global().Set("blocklyToYailWithDiagnostics", js.FuncOf(blocklyToYailWithDiagnostics))
	js.Global().Set("runCode", js.FuncOf(runCode))
	js.Global().Set("runCodeWithDiagnostics", js.FuncOf(runCodeWithDiagnostics))
	js.Global().Set("createSimulationSession", js.FuncOf(createSimulationSession))
	js.Global().Set("setSimulationProperty", js.FuncOf(setSimulationProperty))
	js.Global().Set("dispatchSimulationEvent", js.FuncOf(dispatchSimulationEvent))
	js.Global().Set("disposeSimulationSession", js.FuncOf(disposeSimulationSession))
	js.Global().Set("describeComponent", js.FuncOf(describeComponent))
	js.Global().Set("listComponents", js.FuncOf(listComponents))
	js.Global().Set("falconCompletionCatalog", js.FuncOf(falconCompletionCatalog))
	<-c
}
