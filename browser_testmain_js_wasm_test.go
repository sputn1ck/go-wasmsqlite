//go:build js && wasm

package wasmsqlite

import (
	"fmt"
	"os"
	"syscall/js"
	"testing"
)

func TestMain(m *testing.M) {
	code := m.Run()
	js.Global().Set("__wasmTestPassed", code == 0)
	js.Global().Set("__wasmTestDone", true)
	status := js.Global().Get("Object").New()
	status.Set("code", code)
	status.Set("passed", code == 0)
	js.Global().Get("console").Call("log", "wasmsqlite-browser-test-status", status)
	fmt.Printf("wasmsqlite browser tests finished: code=%d\n", code)
	os.Exit(code)
}
