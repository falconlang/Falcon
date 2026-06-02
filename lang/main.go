//go:build !js && !wasm

package main

import (
	"Falcon/code/ast"
	"Falcon/code/compdb"
	"Falcon/code/context"
	"Falcon/code/lex"
	blocklyParser "Falcon/code/parsers/blocklytomist"
	yailParser "Falcon/code/parsers/blocklytoyail"
	mistParser "Falcon/code/parsers/mistparser"
	"Falcon/code/runtime"
	designAnalysis "Falcon/design"
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) == 1 {
		printUsage(os.Stdout)
		return
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		printUsage(os.Stdout)
		return
	case "repl":
		repl()
		return
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: Falcon run <file.mist>")
			os.Exit(1)
		}
		runFile(os.Args[2])
		return
	case "format":
		formatStdin()
		return
	case "reformat":
		reformatStdin()
		return
	case "roundtrip":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: Falcon roundtrip <file.mist>")
			os.Exit(1)
		}
		roundtripFile(os.Args[2])
		return
	case "exec":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: Falcon exec <file.mist>")
			os.Exit(1)
		}
		execFile(os.Args[2])
		return
	case "correct":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: Falcon correct <file.mist>")
			os.Exit(1)
		}
		correctFile(os.Args[2])
		return
	case "companion":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: Falcon companion <file.mist> <Screen.ann>")
			os.Exit(1)
		}
		companionRun(os.Args[2], os.Args[3])
		return
	case "companion-serve":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "usage: Falcon companion-serve <file.mist> <Screen.ann>")
			os.Exit(1)
		}
		companionServe(os.Args[2], os.Args[3])
		return
	case "eval":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: Falcon eval \"<falcon code>\"")
			os.Exit(1)
		}
		evalSend(strings.Join(os.Args[2:], " "))
		return
	case "refresh":
		refreshSend()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", os.Args[1])
		printUsage(os.Stderr)
		os.Exit(1)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: Falcon <command> [args]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  repl")
	fmt.Fprintln(w, "  run <file.mist>")
	fmt.Fprintln(w, "  exec <file.mist>")
	fmt.Fprintln(w, "  format")
	fmt.Fprintln(w, "  reformat")
	fmt.Fprintln(w, "  roundtrip <file.mist>")
	fmt.Fprintln(w, "  correct <file.mist>")
	fmt.Fprintln(w, "  companion <file.mist> <Screen.ann>")
	fmt.Fprintln(w, "  companion-serve <file.mist> <Screen.ann>")
	fmt.Fprintln(w, "  eval \"<falcon code>\"")
	fmt.Fprintln(w, "  refresh")
}

func newLangParser(strict bool, tokens []*lex.Token) *mistParser.LangParser {
	langParser := mistParser.NewLangParser(strict, tokens)
	langParser.SetEventValidator(compdb.GlobalDB.ValidateEvent)
	langParser.SetPropertyValidator(compdb.GlobalDB.ValidateProperty)
	langParser.SetMethodValidator(compdb.GlobalDB.ValidateMethod)
	return langParser
}

func formatExecutionError(interp *runtime.Interpreter, r any) string {
	switch err := r.(type) {
	case *context.DiagnosticError:
		return err.Error()
	case *context.DiagnosticListError:
		return err.Error()
	}
	if interp != nil {
		return interp.FormatRuntimeError(r)
	}
	if err, ok := r.(error); ok {
		return err.Error()
	}
	if msg, ok := r.(string); ok {
		return msg
	}
	return "unknown error"
}

func reformatStdin() {
	codeBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	source := string(codeBytes)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Error:", r)
			os.Exit(1)
		}
	}()
	codeContext := &context.CodeContext{SourceCode: &source, FileName: "<reformat>"}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := newLangParser(true, tokens)
	exprs := langParser.ParseAll()
	blocks := make([]ast.Block, len(exprs))
	for i, expr := range exprs {
		blocks[i] = expr.Blockly(true)
	}
	xmlBlock := ast.XmlRoot{
		Blocks: blocks,
		XMLNS:  "https://developers.google.com/blockly/xml",
	}
	xmlBytes, _ := xml.MarshalIndent(xmlBlock, "", "  ")
	roundtripped := blocklyParser.NewParser(string(xmlBytes)).GenerateAST()
	var out strings.Builder
	for i, expr := range roundtripped {
		if i > 0 {
			out.WriteRune('\n')
		}
		out.WriteString(expr.String())
	}
	fmt.Print(out.String())
}

// roundtripFile round-trips a Falcon source file through:
//
//	Stage 1 (exit 1): Falcon source → mist parser → AST
//	Stage 2 (exit 2): AST → Blockly serializer → XML
//	Stage 3 (exit 3): XML → Blockly parser → AST → Falcon source
func roundtripFile(path string) {
	codeBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	source := string(codeBytes)
	fileName := filepath.Base(path)

	// Stage 1: Falcon source → AST
	var exprs []ast.Expr
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "stage 1 (mist parser):", r)
				os.Exit(1)
			}
		}()
		codeContext := &context.CodeContext{SourceCode: &source, FileName: fileName}
		tokens := lex.NewLexer(codeContext).Lex()
		exprs = newLangParser(true, tokens).ParseAll()
	}()

	// Stage 2: AST → Blockly XML
	var xmlBytes []byte
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "stage 2 (blockly serializer):", r)
				os.Exit(2)
			}
		}()
		blocks := make([]ast.Block, len(exprs))
		for i, expr := range exprs {
			blocks[i] = expr.Blockly(true)
		}
		xmlBlock := ast.XmlRoot{
			Blocks: blocks,
			XMLNS:  "https://developers.google.com/blockly/xml",
		}
		xmlBytes, _ = xml.MarshalIndent(xmlBlock, "", "  ")
	}()

	// Stage 3: Blockly XML → AST → Falcon source
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintln(os.Stderr, "stage 3 (blockly parser):", r)
				os.Exit(3)
			}
		}()
		roundtripped := blocklyParser.NewParser(string(xmlBytes)).GenerateAST()
		var out strings.Builder
		for i, expr := range roundtripped {
			if i > 0 {
				out.WriteRune('\n')
			}
			out.WriteString(expr.String())
		}
		fmt.Print(out.String())
	}()
}

func formatStdin() {
	codeBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	source := string(codeBytes)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Error:", r)
			os.Exit(1)
		}
	}()
	codeContext := &context.CodeContext{SourceCode: &source, FileName: "<format>"}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := newLangParser(true, tokens)
	exprs := langParser.ParseAll()
	var out strings.Builder
	for i, expr := range exprs {
		if i > 0 {
			out.WriteRune('\n')
		}
		out.WriteString(expr.String())
	}
	fmt.Print(out.String())
}

func runFile(path string) {
	codeBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	source := string(codeBytes)
	fileName := filepath.Base(path)
	interp := runtime.NewInterpreter()
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, formatExecutionError(interp, r))
			//fmt.Fprintln(os.Stderr, "\n--- Go stack trace ---")
			//fmt.Fprintln(os.Stderr, string(debug.Stack()))
			os.Exit(1)
		}
	}()
	codeContext := &context.CodeContext{SourceCode: &source, FileName: fileName}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := newLangParser(true, tokens)
	exprs := langParser.ParseAll()
	//fmt.Println("--- corrected source ---")
	//fmt.Println(langParser.ReconstructedSource())
	//fmt.Println("------------------------")
	for _, e := range exprs {
		e.Blockly()
	}
	interp.Run(exprs)
}

// execFile runs a Falcon source file without the Blockly validation stage.
// Use this when the program uses features (e.g. yield) that are not yet
// representable in the Blockly XML serializer.
func execFile(path string) {
	codeBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	source := string(codeBytes)
	fileName := filepath.Base(path)
	interp := runtime.NewInterpreter()
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, formatExecutionError(interp, r))
			os.Exit(1)
		}
	}()
	codeContext := &context.CodeContext{SourceCode: &source, FileName: fileName}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := newLangParser(true, tokens)
	exprs := langParser.ParseAll()
	interp.Run(exprs)
}

func correctFile(path string) {
	codeBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	source := string(codeBytes)
	fileName := filepath.Base(path)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, "Error:", r)
			os.Exit(1)
		}
	}()
	codeContext := &context.CodeContext{SourceCode: &source, FileName: fileName}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := mistParser.NewLangParser(false, tokens)
	//langParser.EnableAutoCorrect()
	langParser.ParseAll()
	fmt.Print(langParser.ReconstructedSource())
}

func repl() {
	fmt.Println("Falcon REPL  (type :exit to quit, Ctrl+D to exit)")
	fmt.Println()

	interp := runtime.NewInterpreter()
	reader := bufio.NewReader(os.Stdin)

	var inputBuf strings.Builder
	openBraces := 0

	for {
		if openBraces == 0 {
			fmt.Print(">>>> ")
		} else {
			fmt.Print(".. ")
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				fmt.Println("Goodbye!")
				break
			}
			fmt.Fprintln(os.Stderr, "read error:", err)
			continue
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == ":exit" {
			fmt.Println("Goodbye!")
			break
		}

		inString := false
		escaped := false
		for _, c := range line {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' && inString {
				escaped = true
				continue
			}
			if c == '"' {
				inString = !inString
			} else if !inString && c == '{' {
				openBraces++
			} else if !inString && c == '}' {
				if openBraces > 0 {
					openBraces--
				}
			}
		}
		inputBuf.WriteString(line)

		if openBraces > 0 {
			continue
		}
		openBraces = 0

		source := inputBuf.String()
		inputBuf.Reset()

		if strings.TrimSpace(source) == "" {
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintln(os.Stderr, "Error:", r)
				}
			}()
			codeContext := &context.CodeContext{SourceCode: &source, FileName: "<repl>"}
			tokens := lex.NewLexer(codeContext).Lex()
			langParser := newLangParser(true, tokens)
			exprs := langParser.ParseAll()
			for _, e := range exprs {
				e.Blockly()
			}
			result := interp.RunGetLast(exprs)
			if result.Type() != runtime.Null && result.Type() != runtime.NonConsumable {
				fmt.Println("=", result.String())
			}
		}()
	}
}

func runProgram() {
	fileName := "run.mist"
	filePath := "/home/kumaraswamy/Documents/falcon/lang/testing/" + fileName
	codeBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	sourceCode := string(codeBytes)
	codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: fileName}

	tokens := lex.NewLexer(codeContext).Lex()
	langParser := newLangParser(true, tokens)
	exprs := langParser.ParseAll()

	interp := runtime.NewInterpreter()
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, formatExecutionError(interp, r))
			os.Exit(1)
		}
	}()
	interp.Run(exprs)
}

func designTest() {
	xmlFile := "Screen1.aiml"
	xmlPath := "/home/kumaraswamy/Documents/falcon/lang/testing/" + xmlFile
	codeBytes, err := os.ReadFile(xmlPath)
	if err != nil {
		panic(err)
	}
	xmlString := string(codeBytes)
	schemaString, err := designAnalysis.NewXmlParser(xmlString).ConvertXmlToSchema()
	if err != nil {
		panic(err)
	}
	println(schemaString)
	xmlString, err = designAnalysis.NewSchemaParser(schemaString).ConvertSchemaToXml()
	if err != nil {
		panic(err)
	}
	println("Produced XML: ")
	println(xmlString)
}

func annTest() {
	annFile := "Screen1.ann"
	annPath := "/home/kumaraswamy/Documents/falcon/lang/testing/" + annFile
	codeBytes, err := os.ReadFile(annPath)
	if err != nil {
		panic(err)
	}
	yailCode, err := designAnalysis.NewAnnYailConverter().ConvertAnnToYail(string(codeBytes))
	if err != nil {
		panic(err)
	}
	println("\n=== Companion REPL YAIL ===\n")
	println(yailCode)
}

func xmlTest() {
	xmlFile := "xml.txt"
	xmlPath := "/home/kumaraswamy/Documents/falcon/lang/testing/" + xmlFile
	codeBytes, err := os.ReadFile(xmlPath)
	if err != nil {
		panic(err)
	}
	xmlString := string(codeBytes)
	exprs := blocklyParser.NewParser(xmlString).GenerateAST()
	var machineSourceCode strings.Builder
	for _, expr := range exprs {
		machineSourceCode.WriteString(expr.String())
		machineSourceCode.WriteRune('\n')
	}
	println(machineSourceCode.String())
}

func analyzeSyntax() {
	fileName := "hi.mist"
	filePath := "/home/kumaraswamy/Documents/falcon/lang/testing/" + fileName
	codeBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	sourceCode := string(codeBytes)
	codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: fileName}

	// lexical parser
	tokens := lex.NewLexer(codeContext).Lex()
	for _, token := range tokens {
		println(token.Debug())
	}

	println("\n=== AST ===\n")

	// conversion of Falcon -> Blockly XML
	langParser := newLangParser(true, tokens)
	expressions := langParser.ParseAll()
	println(langParser.GetComponentDefinitionsCode())
	for _, expression := range expressions {
		println(expression.String())
	}

	println("\n=== Blockly XML ===\n")

	blocks := make([]ast.Block, len(expressions))
	for i, expression := range expressions {
		blocks[i] = expression.Blockly(true)
	}
	xmlBlock := ast.XmlRoot{
		Blocks: blocks,
		XMLNS:  "https://developers.google.com/blockly/xml",
	}
	bytes, _ := xml.MarshalIndent(xmlBlock, "", "  ")
	xmlContent := string(bytes)

	println(xmlContent)
	println()

	// reconversion of Blockly XML -> Falcon
	exprs := blocklyParser.NewParser(xmlContent).GenerateAST()
	var machineSourceCode strings.Builder
	for _, expr := range exprs {
		machineSourceCode.WriteString(expr.String())
		machineSourceCode.WriteRune('\n')
	}
	println(machineSourceCode.String())

	println("\n=== YAIL ===\n")

	yailCode := yailParser.NewParser(xmlContent).GenerateYAIL()
	println(yailParser.Beautify(yailCode))
}
