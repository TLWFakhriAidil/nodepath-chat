package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DatabaseStatus represents the database connection status
type DatabaseStatus struct {
	Connected    bool   `json:"connected"`
	DatabaseURL  string `json:"database_url,omitempty"`
	DSN          string `json:"dsn,omitempty"`
	Error        string `json:"error,omitempty"`
	MySQLVersion string `json:"mysql_version,omitempty"`
	FlowCount    int    `json:"flow_count,omitempty"`
	Message      string `json:"message"`
}

// convertMySQLURL converts mysql:// URL to proper DSN format for go-sql-driver/mysql
func convertMySQLURL(databaseURL string) string {
	if databaseURL == "" {
		return ""
	}
	
	// Convert mysql:// to proper DSN format if needed
	if strings.HasPrefix(databaseURL, "mysql://") {
		// Remove mysql:// prefix
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
				// Reconstruct with tcp() wrapper for go-sql-driver/mysql
				dsn = userPass + "@tcp(" + hostPort + ")/" + databasePart
				// Add essential MySQL parameters
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
	return databaseURL
}

// testDatabaseConnection tests the database connection and returns status
func testDatabaseConnection() DatabaseStatus {
	status := DatabaseStatus{
		Connected: false,
		Message:   "Database connection test",
	}
	
	// Get DATABASE_URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		status.Error = "DATABASE_URL environment variable is not set"
		status.Message = "DATABASE_URL not found in Railway environment"
		return status
	}
	
	status.DatabaseURL = databaseURL
	
	// Convert to proper DSN format
	dsn := convertMySQLURL(databaseURL)
	status.DSN = dsn
	
	// Test database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		status.Error = "Failed to open database connection: " + err.Error()
		status.Message = "Database connection failed"
		return status
	}
	defer db.Close()
	
	// Configure connection pool (same as production settings)
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)
	
	// Test the connection with ping
	if err := db.Ping(); err != nil {
		status.Error = "Failed to ping database: " + err.Error()
		status.Message = "Database ping failed - likely IP whitelist issue"
		return status
	}
	
	// Get MySQL version
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		status.Error = "Failed to query database version: " + err.Error()
		status.Message = "Database query failed"
		return status
	}
	status.MySQLVersion = version
	
	// Test if we can access the chatbot_flows_nodepath table
	var flowCount int
	err = db.QueryRow("SELECT COUNT(*) FROM chatbot_flows_nodepath").Scan(&flowCount)
	if err != nil {
		// Table might not exist, but connection is working
		status.FlowCount = -1
		status.Message = "Database connected but chatbot_flows_nodepath table not found"
	} else {
		status.FlowCount = flowCount
		status.Message = "Database connection successful"
	}
	
	status.Connected = true
	return status
}

// healthHandler provides a health check endpoint for Railway
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := testDatabaseConnection()
	
	if status.Connected {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	
	json.NewEncoder(w).Encode(status)
}

// dbTestHandler provides detailed database test results
func dbTestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := testDatabaseConnection()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func main() {
	// Test database connection on startup
	log.Println("=== Railway Database Test Service ===")
	status := testDatabaseConnection()
	
	if status.Connected {
		log.Printf("✅ Database connection successful: %s", status.Message)
		log.Printf("✅ MySQL version: %s", status.MySQLVersion)
		log.Printf("✅ Flow count: %d", status.FlowCount)
	} else {
		log.Printf("❌ Database connection failed: %s", status.Message)
		log.Printf("❌ Error: %s", status.Error)
	}
	
	// Set up HTTP routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/db-test", dbTestHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Railway Database Test</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .status { padding: 20px; border-radius: 5px; margin: 20px 0; }
        .success { background-color: #d4edda; border: 1px solid #c3e6cb; color: #155724; }
        .error { background-color: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; }
        pre { background-color: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Railway Database Connection Test</h1>
    <p>This service tests the database connection in your Railway environment.</p>
    
    <h2>Available Endpoints:</h2>
    <ul>
        <li><a href="/health">/health</a> - Health check (returns 200 if DB connected)</li>
        <li><a href="/db-test">/db-test</a> - Detailed database test results</li>
    </ul>
    
    <h2>Usage:</h2>
    <pre>
# Test database connection
curl https://your-app.railway.app/health

# Get detailed test results
curl https://your-app.railway.app/db-test
    </pre>
    
    <div class="status ` + func() string {
			if status.Connected {
				return "success">✅ Database Status: Connected"
			}
			return "error">❌ Database Status: Disconnected"
		}() + `</div>
</body>
</html>
		`))
	})
	
	// Get port from environment (Railway provides this)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	
	log.Printf("🚀 Starting server on port %s", port)
	log.Printf("🔗 Health check: http://localhost:%s/health", port)
	log.Printf("🔗 DB test: http://localhost:%s/db-test", port)
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}