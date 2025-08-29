//go:build js && wasm

package wasmsqlite

import (
	"database/sql/driver"
	"errors"
)

type rows struct {
	cols []string
	data [][]any // each row = []any aligned with cols
	idx  int
}

func (r *rows) Columns() []string { return append([]string(nil), r.cols...) }
func (r *rows) Close() error      { return nil }

func (r *rows) Next(dest []driver.Value) error {
	r.idx++
	if r.idx >= len(r.data) {
		return errors.New("EOF")
	}
	row := r.data[r.idx]
	for i := range dest {
		if i < len(row) {
			dest[i] = row[i]
		} else {
			dest[i] = nil
		}
	}
	return nil
}
