//go:build js && wasm

package wasmsqlite

import (
	"mime"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Ensure correct MIME for WASM & ESM when served by the test HTTP server.
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".mjs", "text/javascript")

	// Optional: set your worker URL here so bridge.go can find it.
	// Match the relative path of your assets in the repo.
	os.Setenv("WASM_SQLITE_WORKER_URL", "/walletdb/wasmsqlite/worker.js")

	os.Exit(m.Run())
}
