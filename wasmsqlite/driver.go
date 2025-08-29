//go:build js && wasm

package wasmsqlite

import (
	"database/sql"
	"database/sql/driver"
)

func init() {
	sql.Register("wasmsqlite", &drv{})
}

type drv struct{}

func (d *drv) Open(name string) (driver.Conn, error) {
	// name can be "opfs:/app.db" or ":memory:"
	br, err := newBridge(name)
	if err != nil {
		return nil, err
	}
	return &conn{br: br}, nil
}
