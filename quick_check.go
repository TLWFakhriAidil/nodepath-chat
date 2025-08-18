package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	fmt.Printf("Connecting to database...\n")

	// Connect to database
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Printf("Connected successfully!\n")

	// Check if jam column exists
	query := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = 'ai_whatsapp_nodepath' 
		  AND COLUMN_NAME = 'jam'
	`
	
	var count int
	err = db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Fatalf("Failed to check jam column: %v", err)
	}

	if count > 0 {
		fmt.Printf("✅ SUCCESS: 'jam' column exists in ai_whatsapp_nodepath table!\n")
	} else {
		fmt.Printf("❌ FAILED: 'jam' column does NOT exist in ai_whatsapp_nodepath table!\n")
	}

	// Also check a few other critical columns
	criticalColumns := []string{"intro", "date_order", "balas", "conv_stage"}
	for _, col := range criticalColumns {
		query = `
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			  AND TABLE_NAME = 'ai_whatsapp_nodepath' 
			  AND COLUMN_NAME = ?
		`
		
		err = db.QueryRow(query, col).Scan(&count)
		if err != nil {
			log.Printf("Error checking column %s: %v", col, err)
			continue
		}

		if count > 0 {
			fmt.Printf("✅ Column '%s' exists\n", col)
		} else {
			fmt.Printf("❌ Column '%s' missing\n", col)
		}
	}
}