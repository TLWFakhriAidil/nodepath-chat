package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	DSN      string
}

// TestResult holds test results
type TestResult struct {
	Test        string    `json:"test"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    string    `json:"duration,omitempty"`
	Data        interface{} `json:"data,omitempty"`
}

// TableInfo holds table information
type TableInfo struct {
	TableName string `json:"table_name"`
	RowCount  int    `json:"row_count"`
}

// FlowData represents chatbot flow data
type FlowData struct {
	ID       int    `json:"id"`
	IDDevice string `json:"id_device"`
	FlowName string `json:"flow_name"`
	Status   string `json:"status"`
}

var db *sql.DB
var dbConfig DatabaseConfig

// convertMySQLURL converts mysql:// URL to DSN format
func convertMySQLURL(mysqlURL string) (string, error) {
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		return "", fmt.Errorf("invalid MySQL URL format: must start with mysql://")
	}

	// Remove mysql:// prefix
	url := strings.TrimPrefix(mysqlURL, "mysql://")

	// Split user:password@host:port/database
	parts := strings.Split(url, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid MySQL URL format: missing @ separator")
	}

	userPass := parts[0]
	hostPortDB := parts[1]

	// Split user:password
	userParts := strings.Split(userPass, ":")
	if len(userParts) != 2 {
		return "", fmt.Errorf("invalid MySQL URL format: missing user:password")
	}
	user := userParts[0]
	password := userParts[1]

	// Split host:port/database
	hostParts := strings.Split(hostPortDB, "/")
	if len(hostParts) != 2 {
		return "", fmt.Errorf("invalid MySQL URL format: missing database")
	}
	database := hostParts[1]

	hostPort := hostParts[0]
	hostPortParts := strings.Split(hostPort, ":")
	if len(hostPortParts) != 2 {
		return "", fmt.Errorf("invalid MySQL URL format: missing port")
	}
	host := hostPortParts[0]
	port := hostPortParts[1]

	// Store config for later use
	dbConfig = DatabaseConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}

	// Create DSN: user:password@tcp(host:port)/database?parseTime=true&charset=utf8mb4
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&timeout=30s&readTimeout=30s&writeTimeout=30s",
		user, password, host, port, database)

	dbConfig.DSN = dsn
	return dsn, nil
}

// initDatabase initializes database connection
func initDatabase() error {
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		return fmt.Errorf("MYSQL_URI environment variable is not set")
	}

	dsn, err := convertMySQLURL(mysqlURI)
	if err != nil {
		return fmt.Errorf("failed to convert MySQL URL: %v", err)
	}

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Configure connection pool for high performance
	db.SetMaxOpenConns(100)    // Maximum open connections
	db.SetMaxIdleConns(10)     // Maximum idle connections
	db.SetConnMaxLifetime(time.Hour) // Connection lifetime

	return nil
}

// testDatabaseConnection tests basic database connectivity
func testDatabaseConnection() TestResult {
	start := time.Now()
	result := TestResult{
		Test:      "Database Connection",
		Timestamp: start,
	}

	err := db.Ping()
	if err != nil {
		result.Status = "FAILED"
		result.Message = fmt.Sprintf("Failed to ping database: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}

	result.Status = "SUCCESS"
	result.Message = fmt.Sprintf("Successfully connected to MySQL server at %s:%s", dbConfig.Host, dbConfig.Port)
	result.Duration = time.Since(start).String()
	return result
}

// testMySQLVersion tests MySQL version retrieval
func testMySQLVersion() TestResult {
	start := time.Now()
	result := TestResult{
		Test:      "MySQL Version",
		Timestamp: start,
	}

	var version string
	err := db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		result.Status = "FAILED"
		result.Message = fmt.Sprintf("Failed to get MySQL version: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}

	result.Status = "SUCCESS"
	result.Message = fmt.Sprintf("MySQL Version: %s", version)
	result.Data = map[string]string{"version": version}
	result.Duration = time.Since(start).String()
	return result
}

// testDatabaseExists tests if the target database exists
func testDatabaseExists() TestResult {
	start := time.Now()
	result := TestResult{
		Test:      "Database Existence",
		Timestamp: start,
	}

	var dbName string
	err := db.QueryRow("SELECT DATABASE()").Scan(&dbName)
	if err != nil {
		result.Status = "FAILED"
		result.Message = fmt.Sprintf("Failed to get current database: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}

	if dbName != dbConfig.Database {
		result.Status = "FAILED"
		result.Message = fmt.Sprintf("Connected to wrong database. Expected: %s, Got: %s", dbConfig.Database, dbName)
		result.Duration = time.Since(start).String()
		return result
	}

	result.Status = "SUCCESS"
	result.Message = fmt.Sprintf("Successfully connected to database: %s", dbName)
	result.Data = map[string]string{"database": dbName}
	result.Duration = time.Since(start).String()
	return result
}

// testNodepathTables tests for tables ending with _nodepath
func testNodepathTables() TestResult {
	start := time.Now()
	result := TestResult{
		Test:      "Nodepath Tables",
		Timestamp: start,
	}

	query := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? 
		AND TABLE_NAME LIKE '%_nodepath'
		ORDER BY TABLE_NAME
	`

	rows, err := db.Query(query, dbConfig.Database)
	if err != nil {
		result.Status = "FAILED"
		result.Message = fmt.Sprintf("Failed to query tables: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	if len(tables) == 0 {
		result.Status = "WARNING"
		result.Message = "No tables ending with '_nodepath' found"
	} else {
		result.Status = "SUCCESS"
		result.Message = fmt.Sprintf("Found %d tables ending with '_nodepath'", len(tables))
	}

	result.Data = map[string]interface{}{"tables": tables, "count": len(tables)}
	result.Duration = time.Since(start).String()
	return result
}

// testTableData tests data retrieval from nodepath tables
func testTableData() TestResult {
	start := time.Now()
	result := TestResult{
		Test:      "Table Data Retrieval",
		Timestamp: start,
	}

	// Get all nodepath tables
	query := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? 
		AND TABLE_NAME LIKE '%_nodepath'
		ORDER BY TABLE_NAME
	`

	rows, err := db.Query(query, dbConfig.Database)
	if err != nil {
		result.Status = "FAILED"
		result.Message = fmt.Sprintf("Failed to query tables: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}
	defer rows.Close()

	var tableInfo []TableInfo
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		// Count rows in each table
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		var count int
		if err := db.QueryRow(countQuery).Scan(&count); err != nil {
			count = -1 // Error getting count
		}

		tableInfo = append(tableInfo, TableInfo{
			TableName: tableName,
			RowCount:  count,
		})
	}

	if len(tableInfo) == 0 {
		result.Status = "WARNING"
		result.Message = "No nodepath tables found to test data retrieval"
	} else {
		result.Status = "SUCCESS"
		result.Message = fmt.Sprintf("Successfully retrieved data info from %d tables", len(tableInfo))
	}

	result.Data = map[string]interface{}{"tables": tableInfo}
	result.Duration = time.Since(start).String()
	return result
}

// testChatbotFlows tests specific chatbot flows table
func testChatbotFlows() TestResult {
	start := time.Now()
	result := TestResult{
		Test:      "Chatbot Flows Data",
		Timestamp: start,
	}

	// Check if chatbot_flows_nodepath table exists
	var exists int
	existsQuery := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? 
		AND TABLE_NAME = 'chatbot_flows_nodepath'
	`
	err := db.QueryRow(existsQuery, dbConfig.Database).Scan(&exists)
	if err != nil || exists == 0 {
		result.Status = "WARNING"
		result.Message = "chatbot_flows_nodepath table not found"
		result.Duration = time.Since(start).String()
		return result
	}

	// Get sample data from chatbot_flows_nodepath
	flowQuery := `
		SELECT id, id_device, flow_name, status 
		FROM chatbot_flows_nodepath 
		LIMIT 5
	`

	rows, err := db.Query(flowQuery)
	if err != nil {
		result.Status = "FAILED"
		result.Message = fmt.Sprintf("Failed to query chatbot flows: %v", err)
		result.Duration = time.Since(start).String()
		return result
	}
	defer rows.Close()

	var flows []FlowData
	for rows.Next() {
		var flow FlowData
		if err := rows.Scan(&flow.ID, &flow.IDDevice, &flow.FlowName, &flow.Status); err != nil {
			continue
		}
		flows = append(flows, flow)
	}

	result.Status = "SUCCESS"
	result.Message = fmt.Sprintf("Successfully retrieved %d chatbot flows", len(flows))
	result.Data = map[string]interface{}{"flows": flows, "count": len(flows)}
	result.Duration = time.Since(start).String()
	return result
}

// runAllTests runs all database tests
func runAllTests() []TestResult {
	var results []TestResult

	// Test 1: Database Connection
	results = append(results, testDatabaseConnection())

	// Test 2: MySQL Version
	results = append(results, testMySQLVersion())

	// Test 3: Database Existence
	results = append(results, testDatabaseExists())

	// Test 4: Nodepath Tables
	results = append(results, testNodepathTables())

	// Test 5: Table Data Retrieval
	results = append(results, testTableData())

	// Test 6: Chatbot Flows
	results = append(results, testChatbotFlows())
	return results
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"message": "Database not initialized",
		})
		return
	}

	err := db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"message": fmt.Sprintf("Database ping failed: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"message": "Database connection is working",
		"database": dbConfig.Database,
		"host": dbConfig.Host,
		"timestamp": time.Now(),
	})
}

// testHandler handles database test requests
func testHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Database not initialized",
		})
		return
	}

	results := runAllTests()

	// Check if any critical tests failed
	criticalFailed := false
	for _, result := range results {
		if result.Test == "Database Connection" && result.Status == "FAILED" {
			criticalFailed = true
			break
		}
	}

	if criticalFailed {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"timestamp": time.Now(),
		"database_config": map[string]string{
			"host":     dbConfig.Host,
			"port":     dbConfig.Port,
			"database": dbConfig.Database,
			"user":     dbConfig.User,
		},
		"tests": results,
	})
}

func main() {
	// Initialize database
	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Database initialized successfully")
	log.Printf("Connected to: %s@%s:%s/%s", dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.Database)

	// Run tests once at startup
	log.Println("Running initial database tests...")
	results := runAllTests()
	for _, result := range results {
		log.Printf("[%s] %s: %s", result.Status, result.Test, result.Message)
	}

	// Setup HTTP handlers
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/db-test", testHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<!DOCTYPE html>
			<html>
			<head>
				<title>MySQL URI Database Test</title>
				<style>
					body { font-family: Arial, sans-serif; margin: 40px; }
					.status { padding: 10px; margin: 10px 0; border-radius: 5px; }
					.success { background-color: #d4edda; color: #155724; }
					.warning { background-color: #fff3cd; color: #856404; }
					.failed { background-color: #f8d7da; color: #721c24; }
				</style>
			</head>
			<body>
				<h1>MySQL URI Database Test Service</h1>
				<p>Database: <strong>%s</strong></p>
				<p>Host: <strong>%s:%s</strong></p>
				<h2>Available Endpoints:</h2>
				<ul>
					<li><a href="/health">/health</a> - Health check</li>
					<li><a href="/db-test">/db-test</a> - Comprehensive database tests</li>
				</ul>
			</body>
			</html>
		`, dbConfig.Database, dbConfig.Host, dbConfig.Port)
	})

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Printf("Health check: http://localhost:%s/health", port)
	log.Printf("Database test: http://localhost:%s/db-test", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}