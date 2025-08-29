//go:build js && wasm

package wasmsqlite

import (
	"context"
	"database/sql"
	"testing"
)

func TestGeneralizedWorker(t *testing.T) {
	// Use in-memory DB for testing
	db, err := sql.Open("wasmsqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	
	ctx := context.Background()
	
	// Test CREATE TABLE
	t.Log("Creating table...")
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT
		)
	`)
	if err != nil {
		t.Fatalf("CREATE TABLE failed: %v", err)
	}
	
	// Test INSERT
	t.Log("Inserting row...")
	result, err := db.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Alice", "alice@example.com")
	if err != nil {
		t.Fatalf("INSERT failed: %v", err)
	}
	
	lastID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId failed: %v", err)
	}
	t.Logf("Inserted row with ID: %d", lastID)
	
	// Test SELECT single row
	t.Log("Selecting single row...")
	var id int64
	var name, email string
	err = db.QueryRowContext(ctx, "SELECT id, name, email FROM users WHERE id = ?", lastID).Scan(&id, &name, &email)
	if err != nil {
		t.Fatalf("SELECT single row failed: %v", err)
	}
	
	if name != "Alice" || email != "alice@example.com" {
		t.Fatalf("Unexpected data: name=%q, email=%q", name, email)
	}
	t.Logf("Retrieved: id=%d, name=%q, email=%q", id, name, email)
	
	// Test INSERT another row
	t.Log("Inserting second row...")
	_, err = db.ExecContext(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "Bob", "bob@example.com")
	if err != nil {
		t.Fatalf("Second INSERT failed: %v", err)
	}
	
	// Test SELECT all rows
	t.Log("Selecting all rows...")
	rows, err := db.QueryContext(ctx, "SELECT id, name, email FROM users")
	if err != nil {
		t.Fatalf("SELECT all failed: %v", err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var id int64
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		t.Logf("Row %d: id=%d, name=%q, email=%q", count+1, id, name, email)
		count++
	}
	
	if count != 2 {
		t.Fatalf("Expected 2 rows, got %d", count)
	}
	
	// Test UPDATE
	t.Log("Updating row...")
	_, err = db.ExecContext(ctx, "UPDATE users SET email = ? WHERE name = ?", "alice.new@example.com", "Alice")
	if err != nil {
		t.Fatalf("UPDATE failed: %v", err)
	}
	
	// Verify UPDATE
	err = db.QueryRowContext(ctx, "SELECT email FROM users WHERE name = ?", "Alice").Scan(&email)
	if err != nil {
		t.Fatalf("SELECT after UPDATE failed: %v", err)
	}
	
	if email != "alice.new@example.com" {
		t.Fatalf("UPDATE didn't work, email is still %q", email)
	}
	
	t.Log("All tests passed!")
}