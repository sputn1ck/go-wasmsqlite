//go:build js && wasm

package wasmsqlite

import (
	"context"
	"database/sql/driver"
	"errors"
)

type conn struct {
	br *bridge
}

func (c *conn) Prepare(query string) (driver.Stmt, error) { return nil, driver.ErrSkip } // not using prepared stmts in v1
func (c *conn) Close() error                              { return c.br.Close() }
func (c *conn) Begin() (driver.Tx, error)                 { return nil, driver.ErrSkip }

// Implement direct Exec/Query so database/sql can call without Prepare.
func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	a := toAny(args)
	resp, err := c.br.Call(ctx, "exec", map[string]any{"sql": query, "args": a})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return result{lastID: resp.LastInsertID, rowsAff: resp.RowsAffected}, nil
}

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	a := toAny(args)
	resp, err := c.br.Call(ctx, "query", map[string]any{"sql": query, "args": a})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return &rows{cols: resp.Columns, data: resp.Rows, idx: -1}, nil
}

// Optional interfaces to hint database/sql:
func (c *conn) Ping(ctx context.Context) error { _, err := c.br.Call(ctx, "ping", nil); return err }

type result struct{ lastID, rowsAff int64 }

func (r result) LastInsertId() (int64, error) { return r.lastID, nil }
func (r result) RowsAffected() (int64, error) { return r.rowsAff, nil }

func toAny(in []driver.NamedValue) []any {
	out := make([]any, len(in))
	for i, nv := range in {
		out[i] = nv.Value
	}
	return out
}
