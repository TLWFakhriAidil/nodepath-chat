package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

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
		mysqlURI = "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway"
	} else {
		// Convert mysql:// format to go-sql-driver format
		if strings.HasPrefix(mysqlURI, "mysql://") {
			mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
			// Replace @ with @tcp( and add ) before /
			parts := strings.Split(mysqlURI, "@")
			if len(parts) == 2 {
				userPass := parts[0]
				hostDbParts := strings.Split(parts[1], "/")
				if len(hostDbParts) == 2 {
					host := hostDbParts[0]
					db := hostDbParts[1]
					mysqlURI = fmt.Sprintf("%s@tcp(%s)/%s", userPass, host, db)
				}
			}
		}
	}

	log.Printf("Connecting to database with DSN: %s", mysqlURI)

	// Connect to database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("=== Checking chatbot_flows_nodepath table structure ===")

	// Get table structure
	rows, err := db.Query("DESCRIBE chatbot_flows_nodepath")
	if err != nil {
		log.Fatalf("Failed to describe table: %v", err)
	}
	defer rows.Close()

	fmt.Printf("%-20s %-20s %-10s %-10s %-15s %-10s\n", "Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println(strings.Repeat("-", 90))

	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra string
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		fmt.Printf("%-20s %-20s %-10s %-10s %-15s %-10s\n", field, fieldType, null, key, defaultVal, extra)
	}

	fmt.Println("\n=== Sample data from chatbot_flows_nodepath ===")
	// Get sample data to see what's actually in the table
	sampleRows, err := db.Query("SELECT * FROM chatbot_flows_nodepath LIMIT 1")
	if err != nil {
		log.Printf("Failed to get sample data: %v", err)
		return
	}
	defer sampleRows.Close()

	// Get column names
	columns, err := sampleRows.Columns()
	if err != nil {
		log.Printf("Failed to get columns: %v", err)
		return
	}

	fmt.Printf("Available columns: %v\n", columns)

	if sampleRows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		err := sampleRows.Scan(valuePtrs...)
		if err != nil {
			log.Printf("Error scanning sample row: %v", err)
			return
		}

		fmt.Println("\nSample row:")
		for i, col := range columns {
			val := values[i]
			if val == nil {
				fmt.Printf("%s: <NULL>\n", col)
			} else {
				fmt.Printf("%s: %v\n", col, val)
			}
		}
	} else {
		fmt.Println("No data found in table")
	}
}