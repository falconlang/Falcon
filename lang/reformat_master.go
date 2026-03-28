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

type ReformatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ReformatEntry struct {
	Messages []ReformatMessage `json:"messages"`
}

func reformatCode(code string, index int) (result string, err interface{}) {
	defer func() {
		if r := recover(); r != nil {
			result = ""
			err = r
		}
	}()

	codeContext := &context.CodeContext{SourceCode: &code, FileName: fmt.Sprintf("<reformat_%d>", index)}
	tokens := lex.NewLexer(codeContext).Lex()
	langParser := mistParser.NewLangParser(false, tokens)
	exprs := langParser.ParseAll()

	// AST -> Blockly XML
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

	// Blockly XML -> AST -> String
	roundtripped := blocklyParser.NewParser(xmlContent).GenerateAST()
	var out strings.Builder
	for i, expr := range roundtripped {
		if i > 0 {
			out.WriteRune('\n')
		}
		out.WriteString(expr.String())
	}
	return out.String(), nil
}

var falconCodeRe = regexp.MustCompile("(?s)```falcon\\s*\n(.*?)```")

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run reformat_master.go <master.jsonl>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimRight(string(fileBytes), "\n"), "\n")

	reformatted := 0
	failed := 0
	skipped := 0

	var outLines []string

	for i, line := range lines {
		lineNum := i + 1
		if strings.TrimSpace(line) == "" {
			outLines = append(outLines, line)
			continue
		}

		var entry ReformatEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			fmt.Fprintf(os.Stderr, "Line %d: JSON parse error: %v\n", lineNum, err)
			outLines = append(outLines, line)
			skipped++
			continue
		}

		changed := false
		for j, msg := range entry.Messages {
			if msg.Role != "assistant" {
				continue
			}
			match := falconCodeRe.FindStringSubmatchIndex(msg.Content)
			if match == nil {
				continue
			}
			// match[2]:match[3] is the capture group (the code body)
			codeStart := match[2]
			codeEnd := match[3]
			oldCode := strings.TrimSpace(msg.Content[codeStart:codeEnd])

			newCode, errVal := reformatCode(oldCode, lineNum)
			if errVal != nil {
				fmt.Fprintf(os.Stderr, "Line %d: reformat error: %v\n", lineNum, errVal)
				failed++
				continue
			}

			// Reconstruct the content with the reformatted code block
			newContent := msg.Content[:match[0]] +
				"```falcon\n" + newCode + "\n```" +
				msg.Content[match[1]:]
			entry.Messages[j].Content = newContent
			changed = true
		}

		newLine, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			fmt.Fprintf(os.Stderr, "Line %d: marshal error: %v\n", lineNum, marshalErr)
			outLines = append(outLines, line)
			skipped++
			continue
		}
		outLines = append(outLines, string(newLine))
		if changed {
			reformatted++
		}

		if lineNum%200 == 0 {
			fmt.Printf("Processed %d lines...\n", lineNum)
		}
	}

	output := strings.Join(outLines, "\n") + "\n"
	if err := os.WriteFile(filePath, []byte(output), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Error writing file:", err)
		os.Exit(1)
	}

	fmt.Printf("\nDone.\n")
	fmt.Printf("Reformatted: %d\n", reformatted)
	fmt.Printf("Failed:      %d\n", failed)
	fmt.Printf("Skipped:     %d\n", skipped)
	fmt.Printf("Total lines: %d\n", len(outLines))
}
