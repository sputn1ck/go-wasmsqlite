//go:build js && wasm

package wasmsqlite

import (
	"context"
	"database/sql"
	"testing"

	"github.com/lightninglabs/nautilus/walletdk/walletdb/sqlc"
)

func TestWasmSqliteDriver_SQLCQueries(t *testing.T) {
	// Note: No need to set WASM_SQLITE_WORKER_URL since we embed worker.js
	// The bridge.go will use the embedded worker.js via blob URL

	// Use in-memory DB for portability across runners (no OPFS needed).
	db, err := sql.Open("wasmsqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Minimal migration (matches your 000001.test.up.sql)
	t.Log("Running migration...")
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	t.Log("Migration completed")

	q := sqlc.New(db)

	// Insert (your InsertName returns :exec with RETURNING id in SQL; driver returns RowsAffected
	// so we just validate via follow-up GetName on rowid 1 for this test)
	t.Logf("Inserting name 'Alice'...")
	if err := q.InsertName(ctx, "Alice"); err != nil {
		t.Fatalf("InsertName: %v", err)
	}

	// Debug: check if the row was inserted
	var count int
	err2 := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if err2 != nil {
		t.Fatalf("Failed to count rows: %v", err2)
	}
	t.Logf("Found %d rows after insert", count)
	
	// Debug: try a direct query
	var debugID int64
	var debugName string
	err3 := db.QueryRowContext(ctx, "SELECT id, name FROM test WHERE id = 1").Scan(&debugID, &debugName)
	if err3 != nil {
		t.Logf("Direct query failed: %v", err3)
	} else {
		t.Logf("Direct query found: id=%d, name=%s", debugID, debugName)
	}

	t.Logf("Getting name with ID 1...")
	one, err := q.GetName(ctx, 1)
	if err != nil {
		t.Fatalf("GetName: %v", err)
	}
	if one.Name != "Alice" {
		t.Fatalf("expected name=Alice, got %q", one.Name)
	}

	if err := q.UpdateName(ctx, sqlc.UpdateNameParams{Name: "Bob", ID: one.ID}); err != nil {
		t.Fatalf("UpdateName: %v", err)
	}

	one2, err := q.GetName(ctx, one.ID)
	if err != nil {
		t.Fatalf("GetName after update: %v", err)
	}
	if one2.Name != "Bob" {
		t.Fatalf("expected updated name=Bob, got %q", one2.Name)
	}

	all, err := q.GetNames(ctx)
	if err != nil {
		t.Logf("GetNames failed, let's check if the data exists first")
		
		// Try a simpler query first
		var count int
		err2 := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
		if err2 != nil {
			t.Fatalf("Failed to count rows: %v", err2)
		}
		t.Logf("Found %d rows in test table", count)
		
		t.Fatalf("GetNames: %v", err)
	}
	if len(all) < 1 {
		t.Fatalf("expected at least 1 row, got %d", len(all))
	}
}
