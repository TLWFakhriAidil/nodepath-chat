package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// TestTunnelConnection tests the SSH tunnel connection to MySQL database
// This script verifies that the tunnel is working and database is accessible
func main() {
	fmt.Println("🔍 Testing SSH Tunnel Connection to MySQL Database")
	fmt.Println("=================================================")

	// Test 1: Check if tunnel port is accessible
	fmt.Println("\n📡 Test 1: Checking tunnel port accessibility...")
	tunnelHost := getEnvOrDefault("TUNNEL_HOST", "mysql-ssh-tunnel.railway.internal")
	tunnelPort := getEnvOrDefault("TUNNEL_PORT", "3307")
	
	if testPortConnection(tunnelHost, tunnelPort) {
		fmt.Printf("✅ Tunnel port %s:%s is accessible\n", tunnelHost, tunnelPort)
	} else {
		fmt.Printf("❌ Tunnel port %s:%s is NOT accessible\n", tunnelHost, tunnelPort)
		return
	}

	// Test 2: Test database connection through tunnel
	fmt.Println("\n🗄️  Test 2: Testing database connection through tunnel...")
	dbURL := getEnvOrDefault("DATABASE_URL", 
		fmt.Sprintf("mysql://admin_aqil:admin_aqil@%s:%s/admin_railway?charset=utf8mb4&parseTime=True&loc=Local", 
			tunnelHost, tunnelPort))
	
	if testDatabaseConnection(dbURL) {
		fmt.Println("✅ Database connection through tunnel successful")
	} else {
		fmt.Println("❌ Database connection through tunnel failed")
		return
	}

	// Test 3: Test basic database operations
	fmt.Println("\n🔧 Test 3: Testing basic database operations...")
	if testDatabaseOperations(dbURL) {
		fmt.Println("✅ Basic database operations successful")
	} else {
		fmt.Println("❌ Basic database operations failed")
		return
	}

	// Test 4: Test connection pool performance
	fmt.Println("\n⚡ Test 4: Testing connection pool performance...")
	if testConnectionPool(dbURL) {
		fmt.Println("✅ Connection pool performance test passed")
	} else {
		fmt.Println("❌ Connection pool performance test failed")
		return
	}

	fmt.Println("\n🎉 All tests passed! SSH tunnel is working correctly.")
	fmt.Println("\n📋 Summary:")
	fmt.Println("   ✅ Tunnel port accessible")
	fmt.Println("   ✅ Database connection established")
	fmt.Println("   ✅ Basic operations working")
	fmt.Println("   ✅ Connection pool performing well")
	fmt.Println("\n🚀 Your application is ready for deployment!")
}

// testPortConnection tests if the tunnel port is accessible
func testPortConnection(host, port string) bool {
	address := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		fmt.Printf("   ❌ Port connection failed: %v\n", err)
		return false
	}
	defer conn.Close()
	fmt.Printf("   ✅ Successfully connected to %s\n", address)
	return true
}

// testDatabaseConnection tests the database connection through the tunnel
func testDatabaseConnection(dbURL string) bool {
	// Convert mysql:// URL to DSN format
	dsn := convertMySQLURL(dbURL)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("   ❌ Failed to open database: %v\n", err)
		return false
	}
	defer db.Close()

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	err = db.PingContext(ctx)
	if err != nil {
		fmt.Printf("   ❌ Failed to ping database: %v\n", err)
		return false
	}

	fmt.Println("   ✅ Database ping successful")
	return true
}

// testDatabaseOperations tests basic database operations
func testDatabaseOperations(dbURL string) bool {
	dsn := convertMySQLURL(dbURL)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("   ❌ Failed to open database: %v\n", err)
		return false
	}
	defer db.Close()

	// Test 1: Show databases
	fmt.Println("   📋 Testing SHOW DATABASES...")
	rows, err := db.Query("SHOW DATABASES")
	if err != nil {
		fmt.Printf("   ❌ SHOW DATABASES failed: %v\n", err)
		return false
	}
	defer rows.Close()

	databaseCount := 0
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err == nil {
			databaseCount++
		}
	}
	fmt.Printf("   ✅ Found %d databases\n", databaseCount)

	// Test 2: Show tables in admin_railway database
	fmt.Println("   📋 Testing SHOW TABLES...")
	rows, err = db.Query("SHOW TABLES")
	if err != nil {
		fmt.Printf("   ❌ SHOW TABLES failed: %v\n", err)
		return false
	}
	defer rows.Close()

	tableCount := 0
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err == nil {
			tableCount++
			fmt.Printf("   📄 Found table: %s\n", tableName)
		}
	}
	fmt.Printf("   ✅ Found %d tables\n", tableCount)

	// Test 3: Test SELECT 1
	fmt.Println("   📋 Testing SELECT 1...")
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		fmt.Printf("   ❌ SELECT 1 failed: %v\n", err)
		return false
	}
	if result != 1 {
		fmt.Printf("   ❌ SELECT 1 returned %d, expected 1\n", result)
		return false
	}
	fmt.Println("   ✅ SELECT 1 successful")

	return true
}

// testConnectionPool tests connection pool performance
func testConnectionPool(dbURL string) bool {
	dsn := convertMySQLURL(dbURL)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("   ❌ Failed to open database: %v\n", err)
		return false
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("   ⚡ Testing concurrent connections...")
	
	// Test concurrent connections
	start := time.Now()
	concurrentTests := 5
	results := make(chan bool, concurrentTests)

	for i := 0; i < concurrentTests; i++ {
		go func(id int) {
			var result int
			err := db.QueryRow("SELECT ?", id).Scan(&result)
			if err != nil {
				fmt.Printf("   ❌ Concurrent test %d failed: %v\n", id, err)
				results <- false
				return
			}
			if result != id {
				fmt.Printf("   ❌ Concurrent test %d returned %d, expected %d\n", id, result, id)
				results <- false
				return
			}
			results <- true
		}(i + 1)
	}

	// Wait for all tests to complete
	successCount := 0
	for i := 0; i < concurrentTests; i++ {
		if <-results {
			successCount++
		}
	}

	duration := time.Since(start)
	fmt.Printf("   ✅ %d/%d concurrent tests passed in %v\n", successCount, concurrentTests, duration)

	if successCount != concurrentTests {
		return false
	}

	// Test connection pool stats
	stats := db.Stats()
	fmt.Printf("   📊 Connection pool stats:\n")
	fmt.Printf("      - Open connections: %d\n", stats.OpenConnections)
	fmt.Printf("      - In use: %d\n", stats.InUse)
	fmt.Printf("      - Idle: %d\n", stats.Idle)
	fmt.Printf("      - Wait count: %d\n", stats.WaitCount)
	fmt.Printf("      - Wait duration: %v\n", stats.WaitDuration)

	return true
}

// convertMySQLURL converts mysql:// URL to DSN format
func convertMySQLURL(mysqlURL string) string {
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		return mysqlURL
	}

	// Remove mysql:// prefix
	dsn := strings.TrimPrefix(mysqlURL, "mysql://")
	
	// Replace first @ with @ and handle the rest
	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) != 2 {
		return dsn
	}

	userPass := parts[0]
	hostDbParams := parts[1]

	// Split host/db and parameters
	var hostDb, params string
	if strings.Contains(hostDbParams, "?") {
		paramParts := strings.SplitN(hostDbParams, "?", 2)
		hostDb = paramParts[0]
		params = "?" + paramParts[1]
	} else {
		hostDb = hostDbParams
	}

	// Construct DSN: user:pass@tcp(host:port)/database?params
	dsn = fmt.Sprintf("%s@tcp(%s)%s", userPass, hostDb, params)
	
	return dsn
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}