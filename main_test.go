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
	// Start HTTP server for Railway deployment testing
	http.HandleFunc("/test-db", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		
		// Get DSN
		dsn := GetDSN()
		fmt.Fprintf(w, "Railway Database Connection Test\n")
		fmt.Fprintf(w, "================================\n\n")
		fmt.Fprintf(w, "DSN: %s\n\n", dsn)
		
		// Test database connection
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			fmt.Fprintf(w, "❌ Failed to open database: %v\n", err)
			return
		}
		defer db.Close()
		
		err = db.Ping()
		if err != nil {
			fmt.Fprintf(w, "❌ Failed to ping database: %v\n", err)
			fmt.Fprintf(w, "\nThis error indicates Railway's IP is not whitelisted.\n")
			fmt.Fprintf(w, "Please check your database access management and add Railway's IP range.\n")
			return
		}
		
		fmt.Fprintf(w, "✅ Database connection successful!\n")
		
		// Test a simple query
		var version string
		err = db.QueryRow("SELECT VERSION()").Scan(&version)
		if err != nil {
			fmt.Fprintf(w, "❌ Failed to query database: %v\n", err)
			return
		}
		
		fmt.Fprintf(w, "✅ Database query successful!\n")
		fmt.Fprintf(w, "MySQL Version: %s\n", version)
		
		// Test if we can access the admin_railway database
		var dbName string
		err = db.QueryRow("SELECT DATABASE()").Scan(&dbName)
		if err != nil {
			fmt.Fprintf(w, "❌ Failed to get current database: %v\n", err)
			return
		}
		
		fmt.Fprintf(w, "✅ Current database: %s\n", dbName)
		fmt.Fprintf(w, "\n🎉 All database tests passed! Railway can access the database.\n")
	})
	
	http.HandleFunc("/api/flows", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Test database connection for flows endpoint
		dsn := GetDSN()
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"success":false,"error":"Failed to open database: %v"}`, err)
			return
		}
		defer db.Close()
		
		err = db.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"success":false,"error":"database not available: %v"}`, err)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"success":true,"message":"Database connection successful","flows":[]}`)
	})
	
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Railway Database Test</title>
		</head>
		<body>
			<h1>Railway Database Connection Test</h1>
			<p><a href="/test-db">Test Database Connection</a></p>
			<p><a href="/api/flows">Test API Flows Endpoint</a></p>
		</body>
		</html>
		`)
	})
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	fmt.Printf("🚀 Starting Railway Database Test Server on port %s...\n", port)
	fmt.Printf("Visit /test-db to test database connection\n")
	fmt.Printf("Visit /api/flows to test the flows endpoint\n")
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}