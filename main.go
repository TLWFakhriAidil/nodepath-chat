package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Database connection pool
var db *sql.DB

// Response structures for API endpoints
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
}

type DatabaseTestResponse struct {
	Status         string                 `json:"status"`
	Connection     string                 `json:"connection"`
	MySQLVersion   string                 `json:"mysql_version"`
	Database       string                 `json:"database"`
	NodepathTables []string               `json:"nodepath_tables"`
	TableData      map[string]interface{} `json:"table_data"`
	Error          string                 `json:"error,omitempty"`
}

type TableInfo struct {
	TableName string `json:"table_name"`
	RowCount  int    `json:"row_count"`
}

// convertMySQLURLToDSN converts MySQL URL to DSN format
func convertMySQLURLToDSN(mysqlURL string) (string, error) {
	parsedURL, err := url.Parse(mysqlURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse MySQL URL: %v", err)
	}

	// Extract components
	username := parsedURL.User.Username()
	password, _ := parsedURL.User.Password()
	host := parsedURL.Host
	database := strings.TrimPrefix(parsedURL.Path, "/")

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, database)

	return dsn, nil
}

// initDatabase initializes the database connection
func initDatabase() error {
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		return fmt.Errorf("MYSQL_URI environment variable is not set")
	}

	log.Printf("Connecting to database with URI: %s", mysqlURI)

	// Convert MySQL URL to DSN
	dsn, err := convertMySQLURLToDSN(mysqlURI)
	if err != nil {
		return fmt.Errorf("failed to convert MySQL URL to DSN: %v", err)
	}

	// Open database connection
	var dbErr error
	db, dbErr = sql.Open("mysql", dsn)
	if dbErr != nil {
		return fmt.Errorf("failed to open database: %v", dbErr)
	}

	// Configure connection pool for high performance
	db.SetMaxOpenConns(100)    // Maximum number of open connections
	db.SetMaxIdleConns(10)     // Maximum number of idle connections
	db.SetConnMaxLifetime(time.Hour) // Maximum connection lifetime

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Message:   "NodePath Chat API is running",
	}

	json.NewEncoder(w).Encode(response)
}

// databaseTestHandler performs comprehensive database tests
func databaseTestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := DatabaseTestResponse{
		Status:    "testing",
		TableData: make(map[string]interface{}),
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		response.Status = "failed"
		response.Error = fmt.Sprintf("Database connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	response.Connection = "successful"

	// Get MySQL version
	var version string
	if err := db.QueryRow("SELECT VERSION()").Scan(&version); err != nil {
		response.Error = fmt.Sprintf("Failed to get MySQL version: %v", err)
	} else {
		response.MySQLVersion = version
	}

	// Get current database name
	var dbName string
	if err := db.QueryRow("SELECT DATABASE()").Scan(&dbName); err != nil {
		response.Error = fmt.Sprintf("Failed to get database name: %v", err)
	} else {
		response.Database = dbName
	}

	// Get all tables ending with _nodepath
	rows, err := db.Query("SHOW TABLES LIKE '%_nodepath'")
	if err != nil {
		response.Error = fmt.Sprintf("Failed to query tables: %v", err)
	} else {
		defer rows.Close()
		var tables []string
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err == nil {
				tables = append(tables, tableName)
			}
		}
		response.NodepathTables = tables

		// Get sample data from each table
		for _, table := range tables {
			var count int
			countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", table)
			if err := db.QueryRow(countQuery).Scan(&count); err == nil {
				tableInfo := TableInfo{
					TableName: table,
					RowCount:  count,
				}
				response.TableData[table] = tableInfo

				// Get sample records (limit 5)
				sampleQuery := fmt.Sprintf("SELECT * FROM %s LIMIT 5", table)
				sampleRows, sampleErr := db.Query(sampleQuery)
				if sampleErr == nil {
					defer sampleRows.Close()
					// Note: For production, you'd want to properly handle column types
					// This is a simplified version for testing
				}
			}
		}
	}

	if response.Error == "" {
		response.Status = "success"
	} else {
		response.Status = "partial_success"
	}

	json.NewEncoder(w).Encode(response)
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize database connection
	if err := initDatabase(); err != nil {
		log.Printf("Database initialization failed: %v", err)
		log.Println("Server will start but database operations will fail")
	}

	// Setup HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/db-test", databaseTestHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "NodePath Chat API",
			"version": "1.0.0",
			"endpoints": "/health, /db-test",
		})
	})

	// Apply CORS middleware
	handler := corsMiddleware(mux)

	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("Database test: http://localhost:%s/db-test", port)

	// Start server
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}