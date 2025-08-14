package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection details
	dsn := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
	
	fmt.Println("Testing MySQL connection...")
	fmt.Println("Host: 159.89.198.71")
	fmt.Println("Port: 3306")
	fmt.Println("Database: admin_railway")
	fmt.Println("Username: admin_aqil")
	fmt.Println("")
	
	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()
	
	// Test the connection
	fmt.Println("Attempting to ping database...")
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	
	fmt.Println("✅ Database connection successful!")
	
	// Test a simple query
	fmt.Println("\nTesting database query...")
	var version string
	err = db.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query database version: %v", err)
	}
	
	fmt.Printf("✅ MySQL Version: %s\n", version)
	
	// List tables in the database
	fmt.Println("\nListing tables in admin_railway database...")
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("Tables found:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("Error scanning table name: %v", err)
			continue
		}
		fmt.Printf("  - %s\n", tableName)
	}
	
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}
	
	fmt.Println("\n✅ Database connection test completed successfully!")
}