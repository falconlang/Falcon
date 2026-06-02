//go:build !js && !wasm

package main

import (
	"Falcon/code/ast"
	"Falcon/code/context"
	"Falcon/code/lex"
	yailParser "Falcon/code/parsers/blocklytoyail"
	"Falcon/design"
	"encoding/json"
	"encoding/xml"
	"errors"
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
	lp := newLangParser(true, tokens)
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
			err = panicErr(r)
		}
	}()
	yail = compileMistToYail(code, "<eval>", typeMap, reverseMap)
	return
}

// ── Shared types ────────────────────────────────────────────────────────────

type evalReq struct {
	Code    string `json:"code,omitempty"`
	Refresh bool   `json:"refresh,omitempty"`
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
		writeLine(os.Stderr, "error reading ann file:", err)
		os.Exit(1)
	}
	annSource := string(annBytes)

	screen, err := design.ParseAnn(annSource)
	if err != nil {
		writeLine(os.Stderr, "error parsing ann file:", err)
		os.Exit(1)
	}
	typeMap, reverseMap := design.ExtractComponents(screen)

	mistBytes, err := os.ReadFile(mistFile)
	if err != nil {
		writeLine(os.Stderr, "error reading mist file:", err)
		os.Exit(1)
	}
	codeYail := compileMistToYail(string(mistBytes), filepath.Base(mistFile), typeMap, reverseMap)

	fullYail, err := design.NewAnnYailConverter().ConvertAnnToReplYail(annSource, codeYail)
	if err != nil {
		writeLine(os.Stderr, "error generating YAIL:", err)
		os.Exit(1)
	}
	replYail := wrapForRepl(fullYail)

	writeLine(os.Stdout, "=== Generated REPL YAIL ===")
	writeLine(os.Stdout, replYail)
	writeLine(os.Stdout, "===========================")

	writeText(os.Stdout, "Enter companion code: ")
	code, err := readWord(os.Stdin)
	if err != nil {
		writeLine(os.Stderr, "error reading code:", err)
		os.Exit(1)
	}

	done := make(chan struct{})
	var doneOnce sync.Once
	repl := NewRepl(code, DefaultRendezvous, 60,
		func(c *webrtc.DataChannel) {
			writeLine(os.Stdout, "Companion connected! Sending YAIL...")
			if err := c.SendText(replYail); err != nil {
				writeLine(os.Stderr, "send error:", err)
			}
		},
		func(graceful bool) {
			if graceful {
				writeLine(os.Stdout, "Companion disconnected gracefully.")
			} else {
				writeLine(os.Stdout, "Companion disconnected unexpectedly.")
			}
			doneOnce.Do(func() { close(done) })
		},
		func(msg webrtc.DataChannelMessage) {
			writeLine(os.Stdout, "=> "+formatReplResponse(msg.Data))
		},
	)

	if err := repl.Connect(); err != nil {
		writeLine(os.Stderr, "connect error:", err)
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
		writeLine(os.Stderr, "error reading ann file:", err)
		os.Exit(1)
	}
	annSource := string(annBytes)

	screen, err := design.ParseAnn(annSource)
	if err != nil {
		writeLine(os.Stderr, "error parsing ann file:", err)
		os.Exit(1)
	}
	typeMap, reverseMap := design.ExtractComponents(screen)

	// Build initial design+code YAIL
	var codeYail string
	if mistFile != "" {
		mistBytes, err := os.ReadFile(mistFile)
		if err != nil {
			writeLine(os.Stderr, "error reading mist file:", err)
			os.Exit(1)
		}
		codeYail = compileMistToYail(string(mistBytes), filepath.Base(mistFile), typeMap, reverseMap)
	}
	fullYail, err := design.NewAnnYailConverter().ConvertAnnToReplYail(annSource, codeYail)
	if err != nil {
		writeLine(os.Stderr, "error generating YAIL:", err)
		os.Exit(1)
	}
	initYail := wrapForRepl(fullYail)

	writeText(os.Stdout, "Enter companion code: ")
	code, err := readWord(os.Stdin)
	if err != nil {
		writeLine(os.Stderr, "error reading code:", err)
		os.Exit(1)
	}

	connected := make(chan *webrtc.DataChannel, 1)
	disconnected := make(chan struct{})
	var disconnectedOnce sync.Once
	repl := NewRepl(code, DefaultRendezvous, 60,
		func(c *webrtc.DataChannel) {
			writeLine(os.Stdout, "Companion connected! Sending initial YAIL...")
			if err := c.SendText(initYail); err != nil {
				writeLine(os.Stderr, "init send error:", err)
			}
			connected <- c
		},
		func(graceful bool) {
			writeLine(os.Stdout, "Companion disconnected.")
			disconnectedOnce.Do(func() { close(disconnected) })
		},
		func(msg webrtc.DataChannelMessage) {
			select {
			case companionRespCh <- string(msg.Data):
			default:
				writeLine(os.Stdout, "[companion] "+formatReplResponse(msg.Data))
			}
		},
	)

	if err := repl.Connect(); err != nil {
		writeLine(os.Stderr, "connect error:", err)
		os.Exit(1)
	}

	ch, err := waitForDataChannel(connected, 15*time.Second)
	if err != nil {
		writeLine(os.Stderr, "connect error:", err)
		os.Exit(1)
	}

	// Drain the initial setup response so it doesn't pollute the first eval
	select {
	case resp := <-companionRespCh:
		writeLine(os.Stdout, "[setup] "+formatReplResponse([]byte(resp)))
	case <-time.After(5 * time.Second):
		writeLine(os.Stdout, "[setup] No response from companion (continuing)")
	}

	ln, err := listenUnixSocket(socketPath)
	if err != nil {
		writeLine(os.Stderr, "socket listen error:", err)
		os.Exit(1)
	}

	writeLine(os.Stdout, "Ready. Listening on", socketPath)

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
		go handleEvalConn(conn, ch, mistFile, annFile, typeMap, reverseMap)
	}
}

func waitForDataChannel(connected <-chan *webrtc.DataChannel, timeout time.Duration) (*webrtc.DataChannel, error) {
	select {
	case ch := <-connected:
		return ch, nil
	case <-time.After(timeout):
		return nil, errors.New("timeout waiting for WebRTC data channel")
	}
}

func listenUnixSocket(path string) (net.Listener, error) {
	ln, err := net.Listen("unix", path)
	if err == nil {
		return ln, nil
	}

	conn, dialErr := net.DialTimeout("unix", path, 200*time.Millisecond)
	if dialErr == nil {
		conn.Close()
		return nil, errors.New(path + " is already in use")
	}

	if removeErr := os.Remove(path); removeErr != nil && !os.IsNotExist(removeErr) {
		return nil, wrapError("remove stale socket", removeErr)
	}
	ln, listenErr := net.Listen("unix", path)
	if listenErr != nil {
		return nil, listenErr
	}
	return ln, nil
}

func handleEvalConn(conn net.Conn, ch *webrtc.DataChannel, mistFile, annFile string, typeMap map[string][]string, reverseMap map[string]string) {
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

	if req.Refresh {
		refreshYail, err := buildRefreshYail(mistFile, annFile, typeMap, reverseMap)
		if err != nil {
			json.NewEncoder(conn).Encode(evalResp{Error: err.Error()})
			return
		}
		if err := ch.SendText(refreshYail); err != nil {
			json.NewEncoder(conn).Encode(evalResp{Error: "send error: " + err.Error()})
			return
		}
		select {
		case raw := <-companionRespCh:
			json.NewEncoder(conn).Encode(evalResp{Value: formatReplResponse([]byte(raw))})
		case <-time.After(5 * time.Second):
			json.NewEncoder(conn).Encode(evalResp{Value: "refreshed"})
		}
		return
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

func buildRefreshYail(mistFile, annFile string, typeMap map[string][]string, reverseMap map[string]string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = panicErr(r)
		}
	}()

	annBytes, readErr := os.ReadFile(annFile)
	if readErr != nil {
		return "", wrapError("read ann", readErr)
	}
	annSource := string(annBytes)

	var codeYail string
	if mistFile != "" {
		mistBytes, readErr := os.ReadFile(mistFile)
		if readErr != nil {
			return "", wrapError("read mist", readErr)
		}
		codeYail = compileMistToYail(string(mistBytes), filepath.Base(mistFile), typeMap, reverseMap)
	}

	fullYail, convErr := design.NewAnnYailConverter().ConvertAnnToReplYail(annSource, codeYail)
	if convErr != nil {
		return "", convErr
	}
	return wrapForRepl(fullYail), nil
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
