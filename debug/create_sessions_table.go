package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
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

	// Read SQL file
	sqlContent, err := ioutil.ReadFile("migrations/create_user_sessions_table.sql")
	if err != nil {
		log.Fatalf("Failed to read SQL file: %v", err)
	}

	// Execute SQL
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		log.Fatalf("Failed to execute SQL: %v", err)
	}

	fmt.Println("Successfully created user_sessions table")
}