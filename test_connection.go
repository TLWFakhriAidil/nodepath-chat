package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Println("❌ DATABASE_URL environment variable is not set")
		return
	}

	fmt.Println("🔗 Attempting to connect to database...")

	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		fmt.Printf("❌ Failed to open database connection: %v\n", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Printf("❌ Failed to ping database: %v\n", err)
		return
	}

	fmt.Println("✅ Database connection successful!")

	// Simple query to check if table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath'").Scan(&tableExists)
	if err != nil {
		fmt.Printf("❌ Error checking table existence: %v\n", err)
		return
	}

	if tableExists > 0 {
		fmt.Println("✅ ai_whatsapp_nodepath table exists")
		
		// Check for jam column specifically
		var jamExists int
		err = db.QueryRow("SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'ai_whatsapp_nodepath' AND COLUMN_NAME = 'jam'").Scan(&jamExists)
		if err != nil {
			fmt.Printf("❌ Error checking jam column: %v\n", err)
			return
		}
		
		if jamExists > 0 {
			fmt.Println("✅ SUCCESS: 'jam' column exists!")
		} else {
			fmt.Println("❌ PROBLEM: 'jam' column still missing!")
		}
	} else {
		fmt.Println("❌ ai_whatsapp_nodepath table does not exist")
	}
}