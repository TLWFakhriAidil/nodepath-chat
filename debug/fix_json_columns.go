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

// fixJSONColumns converts JSON columns to TEXT in ai_whatsapp_nodepath table
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

	log.Println("🔄 Starting JSON to TEXT column migration for ai_whatsapp_nodepath...")

	// Check current JSON columns
	jsonColumns := checkJSONColumns(db)
	if len(jsonColumns) == 0 {
		log.Println("✅ No JSON columns found - migration not needed.")
		return
	}

	log.Printf("Found %d JSON columns to migrate: %v\n", len(jsonColumns), jsonColumns)

	// Migrate each JSON column to TEXT
	for _, column := range jsonColumns {
		log.Printf("🔄 Migrating column: %s\n", column)
		
		// Alter column from JSON to TEXT
		alterQuery := fmt.Sprintf("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN %s TEXT COLLATE utf8mb4_unicode_ci", column)
		log.Printf("Executing: %s\n", alterQuery)
		
		_, err := db.Exec(alterQuery)
		if err != nil {
			log.Printf("❌ Error migrating column %s: %v\n", column, err)
			continue
		}
		
		log.Printf("✅ Successfully migrated %s from JSON to TEXT\n", column)
	}

	// Verify migration
	log.Println("\n🔍 Verifying migration...")
	remainingJSONColumns := checkJSONColumns(db)
	if len(remainingJSONColumns) == 0 {
		log.Println("✅ All JSON columns successfully migrated to TEXT!")
	} else {
		log.Printf("⚠️  Still have %d JSON columns: %v\n", len(remainingJSONColumns), remainingJSONColumns)
	}

	log.Println("🎉 JSON to TEXT migration completed!")
}

// checkJSONColumns returns list of columns that are JSON type
func checkJSONColumns(db *sql.DB) []string {
	query := `SELECT COLUMN_NAME 
			  FROM information_schema.columns 
			  WHERE table_schema = DATABASE() 
			  AND table_name = 'ai_whatsapp_nodepath' 
			  AND DATA_TYPE = 'json'`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error checking JSON columns: %v", err)
		return nil
	}
	defer rows.Close()

	var jsonColumns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			log.Printf("Error scanning column name: %v", err)
			continue
		}
		jsonColumns = append(jsonColumns, columnName)
	}

	return jsonColumns
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