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
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/pion/webrtc/v3"
)

const (
	replRequirePrefix = "(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 "
	socketPath        = "/tmp/falcon-companion.sock"
)

var (
	companionRespCh = make(chan string, 1)
	companionEvalMu sync.Mutex
)

func wrapForRepl(yail string) string {
	return replRequirePrefix + yail + "))"
}

// compileMistToYail converts a .mist source file to raw YAIL using the supplied component maps.
func compileMistToYail(source, fileName string, typeMap map[string][]string, reverseMap map[string]string) string {
	codeContext := &context.CodeContext{SourceCode: &source, FileName: fileName}
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
	return yailParser.NewParser(string(xmlBytes)).GenerateYAIL()
}

// compileFalconSnippet compiles a Falcon code snippet to raw YAIL, recovering from panics.
func compileFalconSnippet(code string, typeMap map[string][]string, reverseMap map[string]string) (yail string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	yail = compileMistToYail(code, "<eval>", typeMap, reverseMap)
	return
}

// ── Shared types ────────────────────────────────────────────────────────────

type evalReq struct {
	Code string `json:"code"`
}

type evalResp struct {
	Value string `json:"value"`
	Error string `json:"error,omitempty"`
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

// ── companionRun ─────────────────────────────────────────────────────────────

func companionRun(mistFile, annFile string) {
	annBytes, err := os.ReadFile(annFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading ann file:", err)
		os.Exit(1)
	}
	annSource := string(annBytes)

	screen, err := design.ParseAnn(annSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing ann file:", err)
		os.Exit(1)
	}
	typeMap, reverseMap := design.ExtractComponents(screen)

	mistBytes, err := os.ReadFile(mistFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading mist file:", err)
		os.Exit(1)
	}
	codeYail := compileMistToYail(string(mistBytes), filepath.Base(mistFile), typeMap, reverseMap)

	fullYail, err := design.NewAnnYailConverter().ConvertAnnToReplYail(annSource, codeYail)
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

	done := make(chan struct{})
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
			close(done)
		},
		func(msg webrtc.DataChannelMessage) {
			fmt.Printf("=> %s\n", formatReplResponse(msg.Data))
		},
	)

	if err := repl.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "connect error:", err)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigChan:
	case <-done:
	}
}

// ── companionServe ───────────────────────────────────────────────────────────

func companionServe(mistFile, annFile string) {
	annBytes, err := os.ReadFile(annFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error reading ann file:", err)
		os.Exit(1)
	}
	annSource := string(annBytes)

	screen, err := design.ParseAnn(annSource)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error parsing ann file:", err)
		os.Exit(1)
	}
	typeMap, reverseMap := design.ExtractComponents(screen)

	// Build initial design+code YAIL
	var codeYail string
	if mistFile != "" {
		mistBytes, err := os.ReadFile(mistFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error reading mist file:", err)
			os.Exit(1)
		}
		codeYail = compileMistToYail(string(mistBytes), filepath.Base(mistFile), typeMap, reverseMap)
	}
	fullYail, err := design.NewAnnYailConverter().ConvertAnnToReplYail(annSource, codeYail)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error generating YAIL:", err)
		os.Exit(1)
	}
	initYail := wrapForRepl(fullYail)

	var code string
	fmt.Print("Enter companion code: ")
	if _, err := fmt.Scan(&code); err != nil {
		fmt.Fprintln(os.Stderr, "error reading code:", err)
		os.Exit(1)
	}

	connected := make(chan *webrtc.DataChannel, 1)
	disconnected := make(chan struct{})
	repl := NewRepl(code, DefaultRendezvous, 60,
		func(c *webrtc.DataChannel) {
			fmt.Println("Companion connected! Sending initial YAIL...")
			if err := c.SendText(initYail); err != nil {
				fmt.Fprintln(os.Stderr, "init send error:", err)
			}
			connected <- c
		},
		func(graceful bool) {
			fmt.Println("Companion disconnected.")
			close(disconnected)
		},
		func(msg webrtc.DataChannelMessage) {
			select {
			case companionRespCh <- string(msg.Data):
			default:
				fmt.Printf("[companion] %s\n", formatReplResponse(msg.Data))
			}
		},
	)

	if err := repl.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, "connect error:", err)
		os.Exit(1)
	}

	ch := <-connected

	// Drain the initial setup response so it doesn't pollute the first eval
	select {
	case resp := <-companionRespCh:
		fmt.Printf("[setup] %s\n", formatReplResponse([]byte(resp)))
	case <-time.After(5 * time.Second):
		fmt.Println("[setup] No response from companion (continuing)")
	}

	os.Remove(socketPath)
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "socket listen error:", err)
		os.Exit(1)
	}

	fmt.Println("Ready. Listening on", socketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigChan:
		case <-disconnected:
		}
		ln.Close()
		os.Remove(socketPath)
		os.Exit(0)
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		}
		go handleEvalConn(conn, ch, typeMap, reverseMap)
	}
}

func handleEvalConn(conn net.Conn, ch *webrtc.DataChannel, typeMap map[string][]string, reverseMap map[string]string) {
	defer conn.Close()

	var req evalReq
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		json.NewEncoder(conn).Encode(evalResp{Error: "invalid request: " + err.Error()})
		return
	}

	companionEvalMu.Lock()
	defer companionEvalMu.Unlock()

	// Drain any stale response before sending
	select {
	case <-companionRespCh:
	default:
	}

	yail, err := compileFalconSnippet(req.Code, typeMap, reverseMap)
	if err != nil {
		json.NewEncoder(conn).Encode(evalResp{Error: err.Error()})
		return
	}

	replYail := wrapForRepl("(begin " + yail + ")")
	if err := ch.SendText(replYail); err != nil {
		json.NewEncoder(conn).Encode(evalResp{Error: "send error: " + err.Error()})
		return
	}

	select {
	case raw := <-companionRespCh:
		json.NewEncoder(conn).Encode(evalResp{Value: formatReplResponse([]byte(raw))})
	case <-time.After(5 * time.Second):
		json.NewEncoder(conn).Encode(evalResp{Error: "timeout waiting for companion"})
	}
}

func formatReplResponse(data []byte) string {
	var response replResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return string(data)
	}
	if response.Status != "OK" {
		return "error: " + string(data)
	}
	var parts []string
	for _, val := range response.Values {
		if val.Status == "OK" {
			if val.Value != "" && val.Value != "*nothing*" {
				parts = append(parts, val.Value)
			}
		} else {
			parts = append(parts, "error: "+val.Value)
		}
	}
	return strings.Join(parts, "\n")
}
