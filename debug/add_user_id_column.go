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

	// Check if user_id column exists
	var columnExists int
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.columns 
					   WHERE table_schema = DATABASE() 
					   AND table_name = 'device_setting_nodepath' 
					   AND column_name = 'user_id'`).Scan(&columnExists)
	if err != nil {
		log.Fatalf("Failed to check if user_id column exists: %v", err)
	}

	if columnExists > 0 {
		fmt.Println("✅ user_id column already exists")
	} else {
		fmt.Println("🔄 Adding user_id column to device_setting_nodepath table...")
		
		// Add user_id column
		addColumnQuery := `ALTER TABLE device_setting_nodepath 
						   ADD COLUMN user_id INT DEFAULT NULL 
						   COMMENT 'User ID linked to this device'`
		
		_, err = db.Exec(addColumnQuery)
		if err != nil {
			log.Fatalf("Failed to add user_id column: %v", err)
		}
		
		fmt.Println("✅ Successfully added user_id column")
	}

	// Add index for user_id if it doesn't exist
	fmt.Println("🔄 Checking for user_id index...")
	var indexExists int
	err = db.QueryRow(`SELECT COUNT(*) FROM information_schema.statistics 
					   WHERE table_schema = DATABASE() 
					   AND table_name = 'device_setting_nodepath' 
					   AND index_name = 'idx_user_id'`).Scan(&indexExists)
	if err != nil {
		log.Printf("Warning: Failed to check for user_id index: %v", err)
	} else if indexExists == 0 {
		fmt.Println("🔄 Adding index for user_id column...")
		
		addIndexQuery := `ALTER TABLE device_setting_nodepath 
						  ADD INDEX idx_user_id (user_id)`
		
		_, err = db.Exec(addIndexQuery)
		if err != nil {
			log.Printf("Warning: Failed to add user_id index: %v", err)
		} else {
			fmt.Println("✅ Successfully added user_id index")
		}
	} else {
		fmt.Println("✅ user_id index already exists")
	}

	fmt.Println("\n=== Migration completed successfully ===")
}