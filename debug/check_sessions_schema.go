package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Check if user_sessions table exists
	var tableName string
	err = db.QueryRow("SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'user_sessions'").Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ user_sessions table does not exist in the database")
			return
		}
		log.Fatal("Error checking table existence:", err)
	}

	fmt.Println("✅ user_sessions table exists in the database")

	// Get table structure
	fmt.Println("\n📋 Current user_sessions table structure:")
	rows, err := db.Query("DESCRIBE user_sessions")
	if err != nil {
		log.Fatal("Error describing table:", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-25s %-8s %-8s %-15s %-10s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println("==================================================================================")

	for rows.Next() {
		var field, fieldType, null, key, extra string
		var defaultValue sql.NullString

		err := rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra)
		if err != nil {
			log.Fatal("Error scanning row:", err)
		}

		defaultStr := "NULL"
		if defaultValue.Valid {
			defaultStr = defaultValue.String
		}

		fmt.Printf("%-20s %-25s %-8s %-8s %-15s %-10s\n", field, fieldType, null, key, defaultStr, extra)
	}

	// Get CREATE TABLE statement
	fmt.Println("\n🔧 CREATE TABLE statement:")
	var createTable string
	err = db.QueryRow("SHOW CREATE TABLE user_sessions").Scan(&tableName, &createTable)
	if err != nil {
		log.Fatal("Error getting CREATE TABLE statement:", err)
	}
	fmt.Println(createTable)

	// Count records
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM user_sessions").Scan(&count)
	if err != nil {
		log.Fatal("Error counting records:", err)
	}
	fmt.Printf("\n📊 Total records in user_sessions: %d\n", count)
}