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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get MYSQL_URI
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is required")
	}

	fmt.Printf("MYSQL_URI: %s\n", mysqlURI)

	// Convert to DSN format
	dsn := convertToDSN(mysqlURI)
	fmt.Printf("DSN: %s\n", dsn)

	// Test connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test ping
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Database connection successful!")
}

func convertToDSN(mysqlURI string) string {
	if mysqlURI == "" {
		return ""
	}
	
	// Convert mysql:// to proper DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		// Remove mysql:// prefix and add tcp() wrapper
		dsn := strings.TrimPrefix(mysqlURI, "mysql://")
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
				return dsn
			}
		}
	}
	
	// Return as-is if already in proper format
	return mysqlURI
}