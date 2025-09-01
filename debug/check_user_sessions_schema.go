package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
	}

	// Convert mysql:// to proper format
	if mysqlURI[:8] == "mysql://" {
		mysqlURI = mysqlURI[8:] // Remove mysql:// prefix
	}

	// Format: user:password@tcp(host:port)/database
	dsn := mysqlURI + "?parseTime=true"
	if mysqlURI == "admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway" {
		dsn = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?parseTime=true"
	}

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully")

	// Check if user_sessions table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'admin_railway' AND table_name = 'user_sessions'").Scan(&tableExists)
	if err != nil {
		log.Fatalf("Failed to check table existence: %v", err)
	}

	if tableExists == 0 {
		fmt.Println("❌ user_sessions table does NOT exist in the database")
		return
	}

	fmt.Println("✅ user_sessions table exists in the database")

	// Describe the table structure
	fmt.Println("\n📋 Current user_sessions table structure:")
	fmt.Printf("%-20s %-25s %-8s %-8s %-15s %s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println("==================================================================================")

	rows, err := db.Query("DESCRIBE user_sessions")
	if err != nil {
		log.Fatalf("Failed to describe table: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		fmt.Printf("%-20s %-25s %-8s %-8s %-15s %s\n",
			getStringValue(field),
			getStringValue(fieldType),
			getStringValue(null),
			getStringValue(key),
			getStringValue(defaultVal),
			getStringValue(extra))
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	// Show the CREATE TABLE statement
	fmt.Println("\n🔧 CREATE TABLE statement:")
	var tableName, createStatement sql.NullString
	err = db.QueryRow("SHOW CREATE TABLE user_sessions").Scan(&tableName, &createStatement)
	if err != nil {
		log.Fatalf("Failed to get CREATE TABLE statement: %v", err)
	}

	fmt.Println(getStringValue(createStatement))

	// Count records
	var recordCount int
	err = db.QueryRow("SELECT COUNT(*) FROM user_sessions").Scan(&recordCount)
	if err != nil {
		log.Fatalf("Failed to count records: %v", err)
	}

	fmt.Printf("\n📊 Total records in user_sessions: %d\n", recordCount)
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}