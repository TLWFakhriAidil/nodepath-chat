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

// convertMySQLURLToDSN converts mysql://user:pass@host:port/db to DSN format
func convertMySQLURLToDSN(mysqlURL string) (string, error) {
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		return mysqlURL, nil // Already in DSN format
	}

	// Remove mysql:// prefix
	dsn := strings.TrimPrefix(mysqlURL, "mysql://")
	
	// Parse user:password@host:port/database format
	parts := strings.Split(dsn, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid MySQL URL format")
	}
	
	userHostPart := parts[0]
	databasePart := parts[1]
	
	// Split user:password@host:port
	atIndex := strings.LastIndex(userHostPart, "@")
	if atIndex <= 0 {
		return "", fmt.Errorf("invalid MySQL URL format: missing @ separator")
	}
	
	userPass := userHostPart[:atIndex]
	hostPort := userHostPart[atIndex+1:]
	
	// Reconstruct with tcp() wrapper for go-sql-driver/mysql
	dsn = userPass + "@tcp(" + hostPort + ")/" + databasePart
	if !strings.Contains(dsn, "?") {
		dsn += "?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
	} else {
		dsn += "&charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
	}
	
	return dsn, nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get MySQL URI from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is not set")
	}

	log.Printf("Testing connection with MYSQL_URI: %s", mysqlURI)

	// Convert to DSN format
	dsn, err := convertMySQLURLToDSN(mysqlURI)
	if err != nil {
		log.Fatalf("Failed to convert MySQL URL to DSN: %v", err)
	}

	log.Printf("Converted DSN: %s", dsn)

	// Test database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test ping
	log.Println("Testing database ping...")
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Database connection successful!")

	// Test a simple query
	log.Println("Testing simple query...")
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query database version: %v", err)
	}

	log.Printf("✅ Database version: %s", version)

	// Test chatbot_flows_nodepath table
	log.Println("Testing chatbot_flows_nodepath table...")
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM chatbot_flows_nodepath").Scan(&count)
	if err != nil {
		log.Printf("❌ Failed to query chatbot_flows_nodepath table: %v", err)
	} else {
		log.Printf("✅ chatbot_flows_nodepath table exists with %d rows", count)
	}

	log.Println("\n🎉 All database tests completed successfully!")
}