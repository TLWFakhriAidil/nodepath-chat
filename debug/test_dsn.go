package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Could not load .env file: %v\n", err)
	}

	// Get the MYSQL_URI
	uri := os.Getenv("MYSQL_URI")
	fmt.Printf("Original URI: %s\n", uri)

	// Convert mysql:// to proper DSN format
	if strings.HasPrefix(uri, "mysql://") {
		// Remove mysql:// prefix
		dsn := strings.TrimPrefix(uri, "mysql://")
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
				fmt.Printf("Converted DSN: %s\n", dsn)
			}
		}
	}
}