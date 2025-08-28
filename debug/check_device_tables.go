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

// checkDeviceTables lists all tables related to device settings
func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get MySQL URI from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		// Use the remote database connection as fallback
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
	}

	// Convert mysql:// URL to DSN format if needed
	if strings.HasPrefix(mysqlURI, "mysql://") {
		// Remove mysql:// prefix and convert to proper DSN format
		mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		// Format: user:password@host:port/database
		// Convert to: user:password@tcp(host:port)/database
		parts := strings.Split(mysqlURI, "/")
		if len(parts) >= 2 {
			userHostPart := parts[0]
			databasePart := parts[1]
			// Split user:password@host:port
			atIndex := strings.LastIndex(userHostPart, "@")
			if atIndex != -1 {
				userPass := userHostPart[:atIndex]
				hostPort := userHostPart[atIndex+1:]
				mysqlURI = fmt.Sprintf("%s@tcp(%s)/%s", userPass, hostPort, databasePart)
			}
		}
	}

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

	fmt.Println("✅ Connected to database successfully!")
	fmt.Println("===========================================")

	// List all tables
	fmt.Println("📋 All tables in database:")
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		log.Fatalf("Failed to show tables: %v", err)
	}
	defer rows.Close()

	var deviceTables []string
	var allTables []string

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			log.Fatalf("Failed to scan table name: %v", err)
		}
		allTables = append(allTables, tableName)
		if strings.Contains(strings.ToLower(tableName), "device") {
			deviceTables = append(deviceTables, tableName)
		}
	}

	fmt.Printf("Total tables: %d\n", len(allTables))
	fmt.Println("\n🔍 Device-related tables:")
	for _, table := range deviceTables {
		fmt.Printf("   - %s\n", table)
	}

	fmt.Println("\n📋 All tables:")
	for _, table := range allTables {
		fmt.Printf("   - %s\n", table)
	}

	// If we found device tables, check their structure
	if len(deviceTables) > 0 {
		fmt.Println("\n🔍 Checking structure of device tables:")
		for _, table := range deviceTables {
			fmt.Printf("\n📊 Structure of %s:\n", table)
			columns, err := db.Query(fmt.Sprintf("DESCRIBE %s", table))
			if err != nil {
				fmt.Printf("   Error describing table: %v\n", err)
				continue
			}
			defer columns.Close()

			for columns.Next() {
				var field, fieldType, null, key, defaultVal, extra sql.NullString
				err := columns.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
				if err != nil {
					fmt.Printf("   Error scanning column: %v\n", err)
					continue
				}
				fmt.Printf("   - %s (%s)\n", getStringValue(field), getStringValue(fieldType))
			}
		}
	}

	// Check for FakhriAidilTLW-001 in device tables
	fmt.Println("\n🔍 Searching for FakhriAidilTLW-001 in device tables:")
	for _, table := range deviceTables {
		// Try to find the device in this table
		query := fmt.Sprintf("SELECT * FROM %s WHERE id_device = ? LIMIT 1", table)
		row := db.QueryRow(query, "FakhriAidilTLW-001")
		
		// Get column names
		columns, err := db.Query(fmt.Sprintf("DESCRIBE %s", table))
		if err != nil {
			continue
		}
		
		var columnNames []string
		for columns.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			err := columns.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
			if err != nil {
				continue
			}
			columnNames = append(columnNames, getStringValue(field))
		}
		columns.Close()
		
		// Try to scan the row
		values := make([]interface{}, len(columnNames))
		valuePtrs := make([]interface{}, len(columnNames))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		
		err = row.Scan(valuePtrs...)
		if err == nil {
			fmt.Printf("\n✅ Found FakhriAidilTLW-001 in table: %s\n", table)
			for i, col := range columnNames {
				var val string
				if values[i] != nil {
					val = fmt.Sprintf("%v", values[i])
				} else {
					val = "NULL"
				}
				fmt.Printf("   %s: %s\n", col, val)
			}
		} else if err != sql.ErrNoRows {
			fmt.Printf("   Error querying %s: %v\n", table, err)
		}
	}
}

// getStringValue safely extracts string from sql.NullString
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}