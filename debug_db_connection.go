package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// GetDSN mimics the exact logic from internal/config/config.go
func GetDSN() string {
	// Check for DATABASE_URL first
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		fmt.Printf("Found DATABASE_URL: %s\n", databaseURL)
		// Convert mysql:// to proper DSN format if needed
		if strings.HasPrefix(databaseURL, "mysql://") {
			// Remove mysql:// prefix and add tcp() wrapper
			dsn := strings.TrimPrefix(databaseURL, "mysql://")
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
					// Reconstruct with tcp() wrapper
					dsn = userPass + "@tcp(" + hostPort + ")/" + databasePart
					if !strings.Contains(dsn, "?") {
						dsn += "?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
					} else {
						dsn += "&charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
					}
					fmt.Printf("Converted DSN: %s\n", dsn)
					return dsn
				}
			}
		}
		return databaseURL
	}
	
	// Fallback to individual MySQL environment variables
	mysqlHost := getEnv("MYSQL_HOST", "159.89.198.71")
	mysqlPort := getEnvAsInt("MYSQL_PORT", 3306)
	mysqlUser := getEnv("MYSQL_USER", "admin_aqil")
	mysqlPassword := getEnv("MYSQL_PASSWORD", "admin_aqil")
	mysqlDatabase := getEnv("MYSQL_DATABASE", "admin_railway")
	
	fallbackDSN := mysqlUser + ":" + mysqlPassword + "@tcp(" + mysqlHost + ":" + strconv.Itoa(mysqlPort) + ")/" + mysqlDatabase + "?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
	fmt.Printf("Using fallback DSN: %s\n", fallbackDSN)
	return fallbackDSN
}

// Helper functions
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	fmt.Println("=== Database Connection Debug ===")
	fmt.Printf("DATABASE_URL env var: %s\n", os.Getenv("DATABASE_URL"))
	fmt.Printf("MYSQL_URI env var: %s\n", os.Getenv("MYSQL_URI"))
	fmt.Printf("MYSQL_HOST env var: %s\n", os.Getenv("MYSQL_HOST"))
	fmt.Printf("MYSQL_USER env var: %s\n", os.Getenv("MYSQL_USER"))
	fmt.Printf("MYSQL_PASSWORD env var: %s\n", os.Getenv("MYSQL_PASSWORD"))
	fmt.Printf("MYSQL_DATABASE env var: %s\n", os.Getenv("MYSQL_DATABASE"))
	fmt.Printf("MYSQL_PORT env var: %s\n", os.Getenv("MYSQL_PORT"))
	fmt.Println()

	// Get the DSN using the same logic as the application
	dsn := GetDSN()
	fmt.Printf("Final DSN: %s\n", dsn)
	fmt.Println()

	// Test the database connection
	fmt.Println("Testing database connection...")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ Failed to open database connection: %v\n", err)
		return
	}
	defer db.Close()

	// Configure connection pool (same as application)
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// Test the connection with ping
	fmt.Println("Pinging database...")
	if err := db.Ping(); err != nil {
		fmt.Printf("❌ Failed to ping database: %v\n", err)
		return
	}

	fmt.Println("✅ Database connection successful!")

	// Test a simple query
	fmt.Println("Testing simple query...")
	var version string
	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		fmt.Printf("❌ Failed to query database: %v\n", err)
		return
	}

	fmt.Printf("✅ Database query successful! MySQL version: %s\n", version)

	// Test if the chatbot_flows_nodepath table exists
	fmt.Println("Checking if chatbot_flows_nodepath table exists...")
	var tableExists bool
	query := "SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'chatbot_flows_nodepath'"
	if err := db.QueryRow(query).Scan(&tableExists); err != nil {
		fmt.Printf("❌ Failed to check table existence: %v\n", err)
		return
	}

	if tableExists {
		fmt.Println("✅ chatbot_flows_nodepath table exists")
		
		// Count rows in the table
		var rowCount int
		if err := db.QueryRow("SELECT COUNT(*) FROM chatbot_flows_nodepath").Scan(&rowCount); err != nil {
			fmt.Printf("❌ Failed to count rows: %v\n", err)
		} else {
			fmt.Printf("✅ Table has %d rows\n", rowCount)
		}
	} else {
		fmt.Println("⚠️ chatbot_flows_nodepath table does not exist")
	}
}