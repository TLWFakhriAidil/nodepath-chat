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

// checkAIWhatsappSchema checks the current schema of ai_whatsapp_nodepath table
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is required")
	}

	// Convert mysql:// URL to DSN format if needed
	dsn := convertMySQLURI(mysqlURI)

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

	log.Println("🔍 Checking ai_whatsapp_nodepath table schema...")

	// Check table schema
	query := `SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA 
			  FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = 'ai_whatsapp_nodepath' 
			  ORDER BY ORDINAL_POSITION`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query table schema: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n=== ai_whatsapp_nodepath Table Schema ===")
	fmt.Printf("%-20s %-15s %-10s %-15s %-15s\n", "COLUMN_NAME", "DATA_TYPE", "NULLABLE", "DEFAULT", "EXTRA")
	fmt.Println(strings.Repeat("-", 80))

	jsonColumns := []string{}
	for rows.Next() {
		var columnName, dataType, isNullable, columnDefault, extra sql.NullString
		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &extra)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Track JSON columns
		if dataType.String == "json" {
			jsonColumns = append(jsonColumns, columnName.String)
		}

		fmt.Printf("%-20s %-15s %-10s %-15s %-15s\n", 
			columnName.String, 
			dataType.String, 
			isNullable.String, 
			columnDefault.String, 
			extra.String)
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	fmt.Println("\n=== Analysis ===")
	if len(jsonColumns) > 0 {
		fmt.Printf("❌ Found %d JSON columns that need to be converted to TEXT:\n", len(jsonColumns))
		for _, col := range jsonColumns {
			fmt.Printf("   - %s\n", col)
		}
		fmt.Println("\n⚠️  These columns need to be migrated from JSON to TEXT type.")
	} else {
		fmt.Println("✅ No JSON columns found - all columns are using appropriate types.")
	}

	// Check if table exists
	var tableExists int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'ai_whatsapp_nodepath'").Scan(&tableExists)
	if err != nil {
		log.Printf("Error checking table existence: %v", err)
	} else if tableExists == 0 {
		fmt.Println("❌ Table ai_whatsapp_nodepath does not exist!")
	} else {
		fmt.Println("✅ Table ai_whatsapp_nodepath exists.")
	}

	log.Println("🎉 Schema check completed!")
}

// convertMySQLURI converts mysql:// URL to DSN format
func convertMySQLURI(mysqlURI string) string {
	if !strings.HasPrefix(mysqlURI, "mysql://") {
		return mysqlURI
	}

	// Remove mysql:// prefix
	dsn := strings.TrimPrefix(mysqlURI, "mysql://")
	return dsn
}