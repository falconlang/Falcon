//go:build !js && !wasm

package main

import (
	"Falcon/code/ast"
	"Falcon/code/context"
	"Falcon/code/lex"
	yailParser "Falcon/code/parsers/blocklytoyail"
	mistParser "Falcon/code/parsers/mistparser"
	"Falcon/design"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/pion/webrtc/v3"
)

const replRequirePrefix = "(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 "

func wrapForRepl(yail string) string {
	return replRequirePrefix + yail + "))"
}

func companionRun(mistFile, annFile string) {
	annBytes, err := os.ReadFile(annFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading ann file:", err)
		os.Exit(1)
	}
	annSource := string(annBytes)

	// Parse .ann to extract component definitions for the Mist parser
	screen, err := design.ParseAnn(annSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing ann file:", err)
		os.Exit(1)
	}
	typeMap, reverseMap := design.ExtractComponents(screen)

	// Compile .mist → AST → Blockly XML → YAIL
	mistBytes, err := os.ReadFile(mistFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading mist file:", err)
		os.Exit(1)
	}
	mistSource := string(mistBytes)
	fileName := filepath.Base(mistFile)

	codeContext := &context.CodeContext{SourceCode: &mistSource, FileName: fileName}
	tokens := lex.NewLexer(codeContext).Lex()
	lp := mistParser.NewLangParser(true, tokens)
	lp.SetComponentDefinitions(typeMap, reverseMap)
	exprs := lp.ParseAll()

	blocks := make([]ast.Block, len(exprs))
	for i, expr := range exprs {
		blocks[i] = expr.Blockly(true)
	}
	xmlRoot := ast.XmlRoot{Blocks: blocks, XMLNS: "https://developers.google.com/blockly/xml"}
	xmlBytes, _ := xml.Marshal(xmlRoot)
	codeYail := yailParser.NewParser(string(xmlBytes)).GenerateYAIL()

	// Combine: design YAIL (with code YAIL embedded at the correct position)
	conv := design.NewAnnYailConverter()
	fullYail, err := conv.ConvertAnnToReplYail(annSource, codeYail)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error generating YAIL:", err)
		os.Exit(1)
	}
	replYail := wrapForRepl(fullYail)

	fmt.Println("=== Generated REPL YAIL ===")
	fmt.Println(replYail)
	fmt.Println("===========================")

	var code string
	fmt.Print("Enter companion code: ")
	if _, err := fmt.Scan(&code); err != nil {
		fmt.Fprintln(os.Stderr, "error reading code:", err)
		os.Exit(1)
	}

	repl := NewRepl(code, DefaultRendezvous, 60,
		func(c *webrtc.DataChannel) {
			fmt.Println("Companion connected! Sending YAIL...")
			if err := c.SendText(replYail); err != nil {
				fmt.Fprintln(os.Stderr, "send error:", err)
			}
		},
		func(graceful bool) {
			if graceful {
				fmt.Println("Companion disconnected gracefully.")
			} else {
				fmt.Println("Companion disconnected unexpectedly.")
			}
		},
		companionHandleMessage,
	)

	if err := repl.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "connect error:", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

type replValue struct {
	Status  string `json:"status"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	BlockID string `json:"blockid"`
}

type replResponse struct {
	Status string      `json:"status"`
	Values []replValue `json:"values"`
}

func companionHandleMessage(msg webrtc.DataChannelMessage) {
	var response replResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		fmt.Printf("Response (raw): %s\n", msg.Data)
		return
	}
	if response.Status != "OK" {
		fmt.Printf("REPL error: %s\n", msg.Data)
		return
	}
	for _, val := range response.Values {
		if val.Status == "OK" {
			if val.Value != "" && val.Value != "*nothing*" {
				fmt.Printf("=> %s\n", val.Value)
			}
		} else {
			fmt.Printf("Error: %s\n", val.Value)
		}
	}
}
