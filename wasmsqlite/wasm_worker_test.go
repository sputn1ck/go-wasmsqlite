//go:build js && wasm

package wasmsqlite

import (
	"syscall/js"
	"testing"
	"time"
)

func TestWorkerCreation(t *testing.T) {
	// Test creating a simple inline worker
	workerCode := `
		self.addEventListener("message", (e) => {
			if (e.data.type === "ping") {
				self.postMessage({ type: "pong", id: e.data.id });
			}
		});
	`
	
	// Create a blob URL for the worker
	blob := js.Global().Get("Blob").New(
		js.ValueOf([]interface{}{workerCode}),
		map[string]interface{}{"type": "application/javascript"},
	)
	
	url := js.Global().Get("URL").Call("createObjectURL", blob)
	defer js.Global().Get("URL").Call("revokeObjectURL", url)
	
	// Create the worker
	worker := js.Global().Get("Worker").New(url)
	defer worker.Call("terminate")
	
	// Set up message handler
	done := make(chan bool, 1)
	onMessage := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		data := args[0].Get("data")
		if data.Get("type").String() == "pong" {
			done <- true
		}
		return nil
	})
	defer onMessage.Release()
	
	worker.Call("addEventListener", "message", onMessage)
	
	// Send ping
	worker.Call("postMessage", map[string]interface{}{
		"type": "ping",
		"id":   1,
	})
	
	// Wait for pong
	select {
	case <-done:
		t.Log("Worker communication successful")
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for worker response")
	}
}