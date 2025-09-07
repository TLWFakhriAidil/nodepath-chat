package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✅ Connected to database successfully")

	// Check all devices and their user IDs
	fmt.Println("\n=== All devices and their user IDs ===")
	query := `SELECT id_device, user_id, provider FROM device_setting_nodepath ORDER BY user_id`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query devices: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var idDevice, provider string
		var userID sql.NullInt32
		err := rows.Scan(&idDevice, &userID, &provider)
		if err != nil {
			log.Printf("Error scanning device: %v", err)
			continue
		}
		
		if userID.Valid {
			fmt.Printf("Device: %s (%s) -> User ID: %d\n", idDevice, provider, userID.Int32)
		} else {
			fmt.Printf("Device: %s (%s) -> User ID: NULL\n", idDevice, provider)
		}
	}

	// Check if there are any users in the users table
	fmt.Println("\n=== Checking users table ===")
	userQuery := `SELECT id, email FROM users LIMIT 5`
	userRows, err := db.Query(userQuery)
	if err != nil {
		fmt.Printf("Error querying users table: %v\n", err)
	} else {
		defer userRows.Close()
		fmt.Println("Users in database:")
		for userRows.Next() {
			var id, email string
			err := userRows.Scan(&id, &email)
			if err != nil {
				log.Printf("Error scanning user: %v", err)
				continue
			}
			fmt.Printf("  User ID: %s, Email: %s\n", id, email)
		}
	}
}