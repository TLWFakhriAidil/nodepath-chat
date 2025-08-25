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
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is not set")
	}

	// Convert mysql:// format to Go driver format
	dsn := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Database Table Structure Check ===")
	fmt.Println()

	// Check chatbot_flows_nodepath table structure
	fmt.Println("=== chatbot_flows_nodepath table structure ===")
	flowStructQuery := `DESCRIBE chatbot_flows_nodepath`
	rows, err := db.Query(flowStructQuery)
	if err != nil {
		fmt.Printf("Error checking chatbot_flows_nodepath structure: %v\n", err)
	} else {
		defer rows.Close()
		fmt.Println("Field\t\tType\t\tNull\tKey\tDefault\tExtra")
		fmt.Println("-----\t\t----\t\t----\t---\t-------\t-----")
		for rows.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
			if err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}
			fmt.Printf("%s\t\t%s\t\t%s\t%s\t%s\t%s\n", 
				getValue(field), getValue(fieldType), getValue(null), 
				getValue(key), getValue(defaultVal), getValue(extra))
		}
	}
	fmt.Println()

	// Try to get flow information with available columns
	fmt.Println("=== Flow Information ===")
	flowQuery := `SELECT * FROM chatbot_flows_nodepath WHERE id = ? LIMIT 1`
	rows2, err := db.Query(flowQuery, "flow_ai_1756016272")
	if err != nil {
		fmt.Printf("Error querying flow: %v\n", err)
	} else {
		defer rows2.Close()
		if rows2.Next() {
			// Get column names
			columns, err := rows2.Columns()
			if err != nil {
				log.Printf("Error getting columns: %v", err)
			} else {
				// Create a slice of interface{} to hold the values
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				// Scan the row
				err = rows2.Scan(valuePtrs...)
				if err != nil {
					log.Printf("Error scanning row: %v", err)
				} else {
					// Print the results
					for i, col := range columns {
						val := values[i]
						if val == nil {
							fmt.Printf("%s: NULL\n", col)
						} else {
							fmt.Printf("%s: %v\n", col, val)
						}
					}
				}
			}
		} else {
			fmt.Println("Flow flow_ai_1756016272 not found")
		}
	}
	fmt.Println()

	fmt.Println("=== Current State Analysis ===")
	fmt.Println("Device: FakhriAidilTLW-001")
	fmt.Println("Phone: 60179645043")
	fmt.Println("Current Node: user_reply-1756015760720")
	fmt.Println("Waiting for Reply: Yes (1)")
	fmt.Println("Flow ID: flow_ai_1756016272")
	fmt.Println("Execution Status: active")
	fmt.Println("Execution ID: NULL")
	fmt.Println("Last Updated: 2025-08-25T04:09:38+08:00")
}

func getValue(v interface{}) string {
	switch val := v.(type) {
	case sql.NullString:
		if val.Valid {
			return val.String
		}
		return "NULL"
	case sql.NullInt64:
		if val.Valid {
			return fmt.Sprintf("%d", val.Int64)
		}
		return "NULL"
	default:
		if val == nil {
			return "NULL"
		}
		return fmt.Sprintf("%v", val)
	}
}