//go:build !js && !wasm

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

// refreshSend connects to a running companion-serve instance and triggers a
// full recompile + resend of the current mist/ann files to the companion.
func refreshSend() {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: companion-serve is not running ("+socketPath+")")
		os.Exit(1)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(evalReq{Refresh: true}); err != nil {
		fmt.Fprintln(os.Stderr, "error sending:", err)
		os.Exit(1)
	}

	var resp evalResp
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		fmt.Fprintln(os.Stderr, "error reading response:", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintln(os.Stderr, resp.Error)
		os.Exit(1)
	}
	fmt.Println("Companion refreshed.")
}

// evalSend connects to a running companion-serve instance via the Unix socket,
// sends Falcon code, and prints the evaluated result.
func evalSend(code string) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: companion-serve is not running ("+socketPath+")")
		os.Exit(1)
	}
	defer conn.Close()

	if err := json.NewEncoder(conn).Encode(evalReq{Code: code}); err != nil {
		fmt.Fprintln(os.Stderr, "error sending:", err)
		os.Exit(1)
	}

	var resp evalResp
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		fmt.Fprintln(os.Stderr, "error reading response:", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintln(os.Stderr, resp.Error)
		os.Exit(1)
	}
	fmt.Println(resp.Value)
}
