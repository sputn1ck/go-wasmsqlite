//go:build js && wasm

package wasmsqlite

import (
	"syscall/js"
	"testing"
)

func TestSimpleWASM(t *testing.T) {
	// Simple test to verify WASM environment is working
	global := js.Global()
	
	// Check if we're in a browser environment
	if !global.Get("window").Truthy() {
		t.Fatal("Not running in browser environment")
	}
	
	// Check if Worker is available
	if !global.Get("Worker").Truthy() {
		t.Fatal("Worker not available in this environment")
	}
	
	t.Log("Basic WASM environment check passed")
}

func TestWorkerURL(t *testing.T) {
	// Test if we can set and get the worker URL
	url := "/walletdb/wasmsqlite/worker.js"
	js.Global().Set("WASM_SQLITE_WORKER_URL", js.ValueOf(url))
	
	result := js.Global().Get("WASM_SQLITE_WORKER_URL")
	if !result.Truthy() || result.String() != url {
		t.Fatalf("Failed to set worker URL, got: %v", result.String())
	}
	
	t.Logf("Worker URL set successfully: %s", url)
}