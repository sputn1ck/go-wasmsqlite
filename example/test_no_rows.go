//go:build js && wasm

package main

import (
	"context"
	"database/sql"
	"fmt"
	"syscall/js"
)

func testNoRowsJS(this js.Value, p []js.Value) interface{} {
	go func() {
		fmt.Println("\n🧪 Testing SELECT with no rows...")
		ctx := context.Background()
		
		// First, clear the database
		fmt.Println("🧹 Clearing database...")
		db.ExecContext(ctx, "DELETE FROM posts")
		db.ExecContext(ctx, "DELETE FROM users")
		
		// Test 1: Try to get a non-existent user
		fmt.Println("\n📍 Test 1: Getting non-existent user (ID: 999999)...")
		user, err := queries.GetUser(ctx, 999999)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("✅ Correctly returned sql.ErrNoRows")
			} else {
				fmt.Printf("❌ Unexpected error: %v\n", err)
			}
		} else {
			fmt.Printf("⚠️ Expected error but got user: %+v\n", user)
		}
		
		// Test 2: Query with QueryRow that returns no results
		fmt.Println("\n📍 Test 2: Direct QueryRow with no results...")
		var username string
		err = db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = ?", 999999).Scan(&username)
		if err != nil {
			if err == sql.ErrNoRows {
				fmt.Println("✅ Correctly returned sql.ErrNoRows for QueryRow")
			} else {
				fmt.Printf("❌ Unexpected error from QueryRow: %v\n", err)
			}
		} else {
			fmt.Printf("⚠️ Expected error but got username: %s\n", username)
		}
		
		// Test 3: Query that returns empty result set
		fmt.Println("\n📍 Test 3: Query with empty result set...")
		rows, err := db.QueryContext(ctx, "SELECT * FROM users WHERE id > 999999")
		if err != nil {
			fmt.Printf("❌ Error executing query: %v\n", err)
			return
		}
		defer rows.Close()
		
		count := 0
		for rows.Next() {
			count++
		}
		if err = rows.Err(); err != nil {
			fmt.Printf("❌ Error iterating rows: %v\n", err)
		} else {
			fmt.Printf("✅ Successfully handled empty result set (found %d rows)\n", count)
		}
		
		// Test 4: List users when table is empty
		fmt.Println("\n📍 Test 4: Listing users from empty table...")
		users, err := queries.ListUsers(ctx)
		if err != nil {
			fmt.Printf("❌ Error listing users: %v\n", err)
		} else {
			fmt.Printf("✅ Successfully listed users from empty table (found %d users)\n", len(users))
		}
		
		fmt.Println("\n✅ All no-rows tests completed!")
	}()
	
	return nil
}