package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// GetDSN constructs the database connection string from environment variables
// This replicates the exact logic from internal/config/config.go
func GetDSN() string {
	// Priority 1: Use DATABASE_URL if available
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		// Convert mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
		if strings.HasPrefix(databaseURL, "mysql://") {
			// Remove mysql:// prefix
			dsn := strings.TrimPrefix(databaseURL, "mysql://")
			
			// Find the @ symbol that separates credentials from host
			parts := strings.Split(dsn, "@")
			if len(parts) >= 2 {
				credentials := parts[0]
				hostAndDB := parts[1]
				
				// Split host:port/database
				hostParts := strings.Split(hostAndDB, "/")
				if len(hostParts) >= 2 {
					hostPort := hostParts[0]
					database := hostParts[1]
					
					// Remove query parameters from database name
					if queryIndex := strings.Index(database, "?"); queryIndex != -1 {
						queryParams := database[queryIndex:]
						database = database[:queryIndex]
						return fmt.Sprintf("%s@tcp(%s)/%s%s", credentials, hostPort, database, queryParams)
					} else {
						return fmt.Sprintf("%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", credentials, hostPort, database)
					}
				}
			}
		}
		return databaseURL
	}
	
	// Priority 2: Fallback to individual environment variables
	host := getEnvWithDefault("MYSQL_HOST", "159.89.198.71")
	port := getEnvWithDefault("MYSQL_PORT", "3306")
	user := getEnvWithDefault("MYSQL_USER", "admin_aqil")
	password := getEnvWithDefault("MYSQL_PASSWORD", "admin_aqil")
	database := getEnvWithDefault("MYSQL_DATABASE", "admin_railway")
	
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, database)
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Test database connection
	dsn := GetDSN()
	fmt.Printf("Testing database connection...\n")
	fmt.Printf("DSN: %s\n\n", dsn)
	
	// Attempt to connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("❌ Failed to open database connection: %v\n", err)
		return
	}
	defer db.Close()
	
	// Test the connection
	err = db.Ping()
	if err != nil {
		fmt.Printf("❌ Failed to ping database: %v\n", err)
		fmt.Printf("\nThis indicates the database server is rejecting the connection.\n")
		fmt.Printf("Possible causes:\n")
		fmt.Printf("1. IP address not whitelisted\n")
		fmt.Printf("2. Incorrect credentials\n")
		fmt.Printf("3. Database server is down\n")
		fmt.Printf("4. Network connectivity issues\n")
		return
	}
	
	fmt.Printf("✅ Database connection successful!\n")
	
	// Test a simple query
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		fmt.Printf("❌ Failed to query database: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Database query successful!\n")
	fmt.Printf("MySQL Version: %s\n", version)
	
	// Test if we can access the admin_railway database
	var dbName string
	err = db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		fmt.Printf("❌ Failed to get current database: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Current database: %s\n", dbName)
	
	// Start HTTP server for Railway deployment testing
	http.HandleFunc("/test-db", func(w http.ResponseWriter, r *http.Request) {
		// Test database connection in HTTP context
		db, err := sql.Open("mysql", GetDSN())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to open database: %v", err)
			return
		}
		defer db.Close()
		
		err = db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Failed to ping database: %v", err)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Database connection successful from Railway!")
	})
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Railway Database Test Server - Visit /test-db to test database connection")
	})
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("\n🚀 Starting HTTP server on port %s...\n", port)
	fmt.Printf("Visit /test-db to test database connection from Railway\n")
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}