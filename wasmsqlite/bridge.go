//go:build js && wasm

package wasmsqlite

import (
	"context"
	_ "embed"
	"errors"
	"math"
	"syscall/js"
)

//go:embed worker.js
var workerJS string

func jsToDriverValue(v js.Value) any {
	switch v.Type() {
	case js.TypeNull, js.TypeUndefined:
		return nil
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeNumber:
		f := v.Float()
		// Prefer int64 when it’s integral (helps scans into sql.NullInt64 etc.)
		if f == math.Trunc(f) && f >= math.MinInt64 && f <= math.MaxInt64 {
			return int64(f)
		}
		return f
	case js.TypeString:
		return v.String()
	case js.TypeObject:
		// Handle Uint8Array -> []byte for BLOBs
		ua := js.Global().Get("Uint8Array")
		if ua.Truthy() && v.InstanceOf(ua) {
			n := v.Get("length").Int()
			b := make([]byte, n)
			// Requires Go 1.20+: copies JS typed array into Go slice
			js.CopyBytesToGo(b, v)
			return b
		}
		// Fallback: try toString (rarely needed for our sqlite rows)
		return v.String()
	default:
		return nil
	}
}

type bridge struct {
	worker js.Value
	nextID int
	wait   map[int]chan js.Value
}

type resp struct {
	OK           bool
	Error        string
	LastInsertID int64
	RowsAffected int64
	Columns      []string
	Rows         [][]any
}

func newBridge(dsn string) (*bridge, error) {
	w := js.Global().Get("Worker")
	if w.IsUndefined() {
		return nil, errors.New("Worker unsupported")
	}
	
	// Create a blob URL for the embedded worker
	blob := js.Global().Get("Blob").New(
		js.ValueOf([]interface{}{workerJS}),
		map[string]interface{}{"type": "application/javascript"},
	)
	url := js.Global().Get("URL").Call("createObjectURL", blob)
	
	worker := w.New(url)
	
	// Clean up the blob URL after creating the worker
	js.Global().Get("URL").Call("revokeObjectURL", url)
	
	b := &bridge{worker: worker, wait: map[int]chan js.Value{}}

	onMsg := js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].Get("data")
		id := data.Get("id").Int()
		if ch, ok := b.wait[id]; ok {
			ch <- data
			close(ch)
			delete(b.wait, id)
		}
		return nil
	})
	worker.Call("addEventListener", "message", onMsg)

	// init db (dsn may be "opfs:/app.db" or ":memory:")
	_, err := b.Call(context.Background(), "open", map[string]any{"dsn": dsn})
	return b, err
}

func (b *bridge) Close() error {
	b.worker.Call("terminate")
	return nil
}

func (b *bridge) Call(ctx context.Context, typ string, payload map[string]any) (resp, error) {
	b.nextID++
	id := b.nextID
	msg := map[string]any{"id": id, "type": typ}
	for k, v := range payload {
		msg[k] = v
	}
	ch := make(chan js.Value, 1)
	b.wait[id] = ch
	b.worker.Call("postMessage", msg)

	select {
	case <-ctx.Done():
		delete(b.wait, id)
		return resp{}, ctx.Err()
	case v := <-ch:
		if !v.Truthy() {
			return resp{}, errors.New("no response")
		}
		if e := v.Get("error"); e.Truthy() {
			return resp{OK: false, Error: e.String()}, nil
		}
		out := resp{OK: true}
		if li := v.Get("lastInsertId"); li.Truthy() {
			out.LastInsertID = int64(li.Int())
		}
		if ra := v.Get("rowsAffected"); ra.Truthy() {
			out.RowsAffected = int64(ra.Int())
		}

		cols := []string{}
		cv := v.Get("columns")
		if cv.Truthy() {
			for i := 0; i < cv.Length(); i++ {
				cols = append(cols, cv.Index(i).String())
			}
		}
		out.Columns = cols

		rows := [][]any{}
		rv := v.Get("rows")
		if rv.Truthy() {
			for i := 0; i < rv.Length(); i++ {
				rowV := rv.Index(i) // JS array for the row
				row := make([]any, len(cols))
				for c := 0; c < len(cols); c++ {
					cell := rowV.Index(c)
					row[c] = jsToDriverValue(cell)
				}
				rows = append(rows, row)
			}
		}
		out.Rows = rows
		return out, nil
	}
}
