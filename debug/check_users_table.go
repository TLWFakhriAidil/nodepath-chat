package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/go-sql-driver/mysql"
)

/**
 * Check users_nodepath table structure
 */
func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		// Fallback to MYSQL_URI if DATABASE_URL is not set
		databaseURL = os.Getenv("MYSQL_URI")
	}
	if databaseURL == "" {
		log.Fatal("DATABASE_URL or MYSQL_URI environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✅ Database connection successful")

	fmt.Println("\n=== CHECKING USERS TABLE ===")

	// Check if users table exists
	var tableExists int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_name = 'users'
	`).Scan(&tableExists)
	if err != nil {
		log.Printf("❌ Error checking if users table exists: %v", err)
		return
	}

	if tableExists == 0 {
		fmt.Println("❌ users table does not exist")
		
		// Check for alternative table names
		fmt.Println("\nChecking for alternative user table names:")
		rows, err := db.Query(`
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = DATABASE() 
			AND table_name LIKE '%user%'
		`)
		if err != nil {
			log.Printf("❌ Error checking for user tables: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var tableName string
			err := rows.Scan(&tableName)
			if err != nil {
				log.Printf("❌ Error scanning table name: %v", err)
				continue
			}
			fmt.Printf("   Found table: %s\n", tableName)
		}
		return
	}

	fmt.Println("✅ users table exists")

	// Get table structure
	fmt.Println("\nTable structure:")
	rows, err := db.Query("DESCRIBE users")
	if err != nil {
		log.Printf("❌ Error describing users_nodepath table: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra string
		// Handle NULL values in default column
		var defaultPtr *string
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultPtr, &extra)
		if err != nil {
			log.Printf("❌ Error scanning column info: %v", err)
			continue
		}
		if defaultPtr != nil {
			defaultVal = *defaultPtr
		} else {
			defaultVal = "NULL"
		}
		fmt.Printf("   - %s: %s (Null: %s, Key: %s, Default: %s)\n", field, fieldType, null, key, defaultVal)
	}

	// Check current user count
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Printf("❌ Error counting users: %v", err)
		return
	}
	fmt.Printf("\nCurrent user count: %d\n", userCount)

	// Show sample users if any exist
	if userCount > 0 {
		fmt.Println("\nSample users:")
		rows, err := db.Query("SELECT id, email FROM users LIMIT 3")
		if err != nil {
			log.Printf("❌ Error fetching sample users: %v", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id, email string
			err := rows.Scan(&id, &email)
			if err != nil {
				log.Printf("❌ Error scanning user: %v", err)
				continue
			}
			fmt.Printf("   - ID: %s, Email: %s\n", id, email)
		}
	}
}