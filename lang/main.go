//go:build !js && !wasm

package main

import (
	"Falcon/code/ast"
	"Falcon/code/context"
	"Falcon/code/lex"
	blocklyParser "Falcon/code/parsers/blocklytomist"
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
	if len(os.Args) > 1 {
		switch os.Args[1] {
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
		}
	}
	//repl()
	//diffTest()
	analyzeSyntax()
	//xmlTest()
	//designTest()
	//runProgram()
	//runFile("/home/kumaraswamy/Documents/falcon/testing/run.mist")
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
	langParser := mistParser.NewLangParser(false, tokens)
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
		exprs = mistParser.NewLangParser(false, tokens).ParseAll()
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
	langParser := mistParser.NewLangParser(false, tokens)
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
			fmt.Fprintln(os.Stderr, interp.FormatRuntimeError(r))
			//fmt.Fprintln(os.Stderr, "\n--- Go stack trace ---")
			//fmt.Fprintln(os.Stderr, string(debug.Stack()))
			os.Exit(1)
		}
	}()
	codeContext := &context.CodeContext{SourceCode: &source, FileName: fileName}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := mistParser.NewLangParser(false, tokens)
	exprs := langParser.ParseAll()
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
			fmt.Fprintln(os.Stderr, interp.FormatRuntimeError(r))
			os.Exit(1)
		}
	}()
	codeContext := &context.CodeContext{SourceCode: &source, FileName: fileName}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := mistParser.NewLangParser(false, tokens)
	exprs := langParser.ParseAll()
	interp.Run(exprs)
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

		for _, c := range line {
			if c == '{' {
				openBraces++
			} else if c == '}' {
				openBraces--
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
			langParser := mistParser.NewLangParser(false, tokens)
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
	filePath := "/home/kumaraswamy/Documents/falcon/testing/" + fileName
	codeBytes, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	sourceCode := string(codeBytes)
	codeContext := &context.CodeContext{SourceCode: &sourceCode, FileName: fileName}

	tokens := lex.NewLexer(codeContext).Lex()
	langParser := mistParser.NewLangParser(false, tokens)
	exprs := langParser.ParseAll()

	interp := runtime.NewInterpreter()
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(os.Stderr, interp.FormatRuntimeError(r))
			os.Exit(1)
		}
	}()
	interp.Run(exprs)
}

func designTest() {
	xmlFile := "Screen1.aiml"
	xmlPath := "/home/ekina/Documents/Falcon/testing/" + xmlFile
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

func xmlTest() {
	xmlFile := "xml.txt"
	xmlPath := "/home/kumaraswamy/Documents/falcon/testing/" + xmlFile
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
	filePath := "/home/kumaraswamy/Documents/falcon/testing/" + fileName
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
	langParser := mistParser.NewLangParser(true, tokens)
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

	//// Generate a merged syntax
	//println("\n=== DIFF ===\n")
	//syntaxDiff := diff.MakeSyntaxDiff(sourceCode, machineSourceCode.String())
	//println(syntaxDiff.Merge())
}
