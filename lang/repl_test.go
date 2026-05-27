//go:build !js && !wasm

package main

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrintUsage(t *testing.T) {
	var out bytes.Buffer
	printUsage(&out)
	if !strings.Contains(out.String(), "usage: Falcon <command> [args]") {
		t.Fatalf("usage output = %q", out.String())
	}
}

func TestParseCommConfigUsesAppInventorDefaultsAndLowercaseIceServers(t *testing.T) {
	cfg, err := parseCommConfig([]byte(`{"iceservers":[{"server":"stun:example.org","username":"u","password":"p"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Rendezvous2 != DefaultRendezvous2 {
		t.Fatalf("Rendezvous2 = %q, want %q", cfg.Rendezvous2, DefaultRendezvous2)
	}
	if len(cfg.IceServers) != 1 || cfg.IceServers[0].Server != "stun:example.org" {
		t.Fatalf("IceServers = %#v", cfg.IceServers)
	}

	cfg, err = parseCommConfig([]byte(`{"rendezvous2":"https://example.org/rendezvous2"}`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Rendezvous2 != "https://example.org/rendezvous2/" {
		t.Fatalf("Rendezvous2 = %q, want normalized trailing slash", cfg.Rendezvous2)
	}
	if len(cfg.IceServers) != 1 || cfg.IceServers[0].Server != "turn:turn.appinventor.mit.edu:3478" {
		t.Fatalf("default IceServers = %#v", cfg.IceServers)
	}
}

func TestReplUsesRawCodeRendezvousKeysForWebRTC(t *testing.T) {
	repl := NewRepl("123456", DefaultRendezvous, 1, nil, nil, nil)
	if repl.sendKey() != "123456-s" {
		t.Fatalf("sendKey() = %q", repl.sendKey())
	}
	if repl.responseKey() != "123456-r" {
		t.Fatalf("responseKey() = %q", repl.responseKey())
	}
	if repl.sendKey() == repl.sha1Digest+"-s" {
		t.Fatal("send key uses SHA-1 digest, want raw connection code")
	}
}

func TestReplDisconnectCallbackIsIdempotent(t *testing.T) {
	calls := 0
	repl := NewRepl("123456", DefaultRendezvous, 1, nil, func(bool) {
		calls++
	}, nil)
	repl.notifyDisconnect(false)
	repl.notifyDisconnect(true)
	if calls != 1 {
		t.Fatalf("disconnect callback calls = %d, want 1", calls)
	}
}

func TestListenUnixSocketDoesNotRemoveLiveSocket(t *testing.T) {
	path := filepath.Join(t.TempDir(), "falcon.sock")
	ln, err := listenUnixSocket(path)
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("unix sockets are not permitted in this sandbox: %v", err)
		}
		t.Fatal(err)
	}
	defer func() {
		ln.Close()
		os.Remove(path)
	}()

	if second, err := listenUnixSocket(path); err == nil {
		second.Close()
		t.Fatal("second listenUnixSocket() unexpectedly succeeded")
	}

	conn, err := net.DialTimeout("unix", path, time.Second)
	if err != nil {
		t.Fatalf("original socket was not reachable after second listen attempt: %v", err)
	}
	conn.Close()
}
