//go:build js && wasm

package wasmsqlite // use the SAME package as your test files

import (
	"mime"
	"syscall/js"
)

func init() {
	// Make the test HTTP server serve correct types
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".mjs", "text/javascript") // ok for Chrome
	_ = mime.AddExtensionType(".js", "text/javascript")

	// Point the bridge to THIS folder’s worker when running tests from here
	js.Global().Set("WASM_SQLITE_WORKER_URL", js.ValueOf("/worker.js"))
}
