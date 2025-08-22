package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Get database URL from environment - using MYSQL_URI exclusively
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		fmt.Println("❌ MYSQL_URI environment variable is not set")
		return
	}

	fmt.Printf("Original URI: %s\n", mysqlURI)

	// Convert mysql:// to proper DSN format (same logic as config.GetDSN())
	var dsn string
	if strings.HasPrefix(mysqlURI, "mysql://") {
		// Remove mysql:// prefix and add tcp() wrapper
		dsn = strings.TrimPrefix(mysqlURI, "mysql://")
		// Parse user:password@host:port/database format
		parts := strings.Split(dsn, "/")
		if len(parts) >= 2 {
			userHostPart := parts[0]
			databasePart := parts[1]
			// Split user:password@host:port
			atIndex := strings.LastIndex(userHostPart, "@")
			if atIndex > 0 {
				userPass := userHostPart[:atIndex]
				hostPort := userHostPart[atIndex+1:]
				// Reconstruct with tcp() wrapper for go-sql-driver/mysql
				dsn = userPass + "@tcp(" + hostPort + ")/" + databasePart
				if !strings.Contains(dsn, "?") {
					dsn += "?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
				} else {
					dsn += "&charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
				}
			}
		}
	} else {
		// Return as-is if already in proper format
		dsn = mysqlURI
	}

	fmt.Printf("Converted DSN: %s\n", dsn)
	fmt.Println("🔗 Attempting to connect to database...")

	db, err := sql.Open("mysql", dsn)
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