//go:build js && wasm

package wasmsqlite

import (
	"context"
	"database/sql"
	"testing"
)

func TestDebugSQLite(t *testing.T) {
	// Use in-memory DB for testing
	db, err := sql.Open("wasmsqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	// Create table
	t.Log("Creating test table...")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS test (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
	t.Log("Table created successfully")

	// Insert a row - use exact SQLC query
	t.Log("Inserting row with name='TestUser' using SQLC query...")
	sqlcQuery := `INSERT INTO test (name) VALUES (?) RETURNING id`
	t.Logf("SQL: %q", sqlcQuery)
	result, err := db.ExecContext(ctx, sqlcQuery, "TestUser")
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}
	
	lastID, err := result.LastInsertId()
	if err != nil {
		t.Logf("LastInsertId error: %v", err)
	} else {
		t.Logf("LastInsertId: %d", lastID)
	}
	
	rowsAff, err := result.RowsAffected()
	if err != nil {
		t.Logf("RowsAffected error: %v", err)
	} else {
		t.Logf("RowsAffected: %d", rowsAff)
	}

	// Count rows
	t.Log("Counting rows...")
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Fatalf("COUNT failed: %v", err)
	}
	t.Logf("Row count: %d", count)

	if count != 1 {
		t.Fatalf("Expected 1 row, got %d", count)
	}

	t.Log("Test passed!")
}