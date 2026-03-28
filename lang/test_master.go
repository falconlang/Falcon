//go:build !js && !wasm

package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"

	"Falcon/code/ast"
	"Falcon/code/context"
	"Falcon/code/lex"
	blocklyParser "Falcon/code/parsers/blocklytomist"
	mistParser "Falcon/code/parsers/mistparser"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Entry struct {
	Messages []Message `json:"messages"`
}

func extractFalconCode(content string) string {
	re := regexp.MustCompile("(?s)```falcon\\s*\n(.*?)```")
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func processCode(code string, index int) (success bool, err interface{}) {
	defer func() {
		if r := recover(); r != nil {
			success = false
			err = r
		}
	}()

	codeContext := &context.CodeContext{SourceCode: &code, FileName: fmt.Sprintf("<test_%d>", index)}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := mistParser.NewLangParser(false, tokens)
	exprs := langParser.ParseAll()

	blocks := make([]ast.Block, len(exprs))
	for i, expr := range exprs {
		blocks[i] = expr.Blockly()
	}

	xmlBlock := ast.XmlRoot{
		Blocks: blocks,
		XMLNS:  "https://developers.google.com/blockly/xml",
	}
	xmlBytes, _ := xml.MarshalIndent(xmlBlock, "", "  ")
	xmlContent := string(xmlBytes)

	blocklyParser.NewParser(xmlContent).GenerateAST()

	return true, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run test_master.go <master.jsonl>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		os.Exit(1)
	}

	lines := strings.Split(string(fileBytes), "\n")

	passed := 0
	failed := 0
	skipped := 0

	type Failure struct {
		Index   int
		Code    string
		Prompt  string
		Error   interface{}
	}

	var failures []Failure

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			fmt.Fprintf(os.Stderr, "Line %d: JSON parse error: %v\n", i+1, err)
			skipped++
			continue
		}

		var falconCode, prompt string
		for _, msg := range entry.Messages {
			if msg.Role == "user" {
				prompt = msg.Content
			} else if msg.Role == "assistant" {
				falconCode = extractFalconCode(msg.Content)
			}
		}

		if falconCode == "" {
			fmt.Fprintf(os.Stderr, "Line %d: No falcon code found\n", i+1)
			skipped++
			continue
		}

		success, errVal := processCode(falconCode, i+1)
		if success {
			passed++
		} else {
			failed++
			failures = append(failures, Failure{
				Index:  i + 1,
				Code:   falconCode,
				Prompt: prompt,
				Error:  errVal,
			})
		}

		if (i+1)%100 == 0 {
			fmt.Printf("Processed %d: passed=%d, failed=%d, skipped=%d\n", i+1, passed+failed+skipped, passed, failed)
		}
	}

	fmt.Println("\n=== FINAL RESULTS ===")
	fmt.Printf("Passed:  %d\n", passed)
	fmt.Printf("Failed:  %d\n", failed)
	fmt.Printf("Skipped: %d\n", skipped)
	fmt.Printf("Total:   %d\n", passed+failed+skipped)

	if len(failures) > 0 && len(failures) <= 50 {
		fmt.Println("\n=== FAILURES ===")
		for _, f := range failures {
			fmt.Printf("\n--- Line %d ---\n", f.Index)
			fmt.Printf("Prompt: %.100s...\n", f.Prompt)
			fmt.Printf("Code: %.200s...\n", f.Code)
			fmt.Printf("Error: %v\n", f.Error)
		}
	} else if len(failures) > 50 {
		fmt.Printf("\n%d failures (showing first 20):\n", len(failures))
		for _, f := range failures[:20] {
			fmt.Printf("Line %d: %v\n", f.Index, f.Error)
		}
	}
}