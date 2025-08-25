package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// convertMySQLURI converts mysql://user:pass@host:port/db to user:pass@tcp(host:port)/db
func convertMySQLURI(uri string) string {
	if !strings.HasPrefix(uri, "mysql://") {
		return uri
	}

	// Remove mysql:// prefix
	uri = strings.TrimPrefix(uri, "mysql://")

	// Split into parts
	parts := strings.Split(uri, "/")
	if len(parts) != 2 {
		return uri
	}

	connPart := parts[0]
	dbName := parts[1]

	// Split connection part
	userPassHost := strings.Split(connPart, "@")
	if len(userPassHost) != 2 {
		return uri
	}

	userPass := userPassHost[0]
	hostPort := userPassHost[1]

	// Format as DSN
	return fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci", userPass, hostPort, dbName)
}

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

	fmt.Println("✅ Database connection successful")

	// Read the migration file
	sqlContent, err := ioutil.ReadFile("migrations/000017_add_flow_tracking_fields.up.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Split SQL statements by semicolon and execute each one
	sqlStatements := strings.Split(string(sqlContent), ";")

	fmt.Println("🔄 Applying flow tracking migration...")

	for i, statement := range sqlStatements {
		statement = strings.TrimSpace(statement)
		if statement == "" || strings.HasPrefix(statement, "--") {
			continue
		}

		fmt.Printf("Executing statement %d...\n", i+1)
		_, err = db.Exec(statement)
		if err != nil {
			// Check if error is about column or index already existing
			if strings.Contains(err.Error(), "Duplicate column name") {
				fmt.Printf("⚠️ Column already exists, skipping: %v\n", err)
				continue
			}
			if strings.Contains(err.Error(), "Duplicate key name") {
				fmt.Printf("⚠️ Index already exists, skipping: %v\n", err)
				continue
			}
			log.Fatalf("Failed to execute SQL statement: %v\nStatement: %s", err, statement)
		}
	}

	fmt.Println("✅ Flow tracking migration applied successfully!")

	// Verify the new columns were added
	fmt.Println("\n🔍 Verifying new columns...")
	rows, err := db.Query(`
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = 'ai_whatsapp_nodepath' 
		  AND COLUMN_NAME IN ('current_node_id', 'waiting_for_reply', 'flow_id', 'last_node_id')
		ORDER BY COLUMN_NAME
	`)
	if err != nil {
		log.Printf("Warning: Could not verify columns: %v", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-20s %-15s %-10s %-15s %-30s\n", "COLUMN_NAME", "DATA_TYPE", "NULLABLE", "DEFAULT", "COMMENT")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var columnName, dataType, isNullable, columnDefault, columnComment sql.NullString
		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &columnComment)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		defaultVal := "NULL"
		if columnDefault.Valid {
			defaultVal = columnDefault.String
		}

		comment := ""
		if columnComment.Valid {
			comment = columnComment.String
		}

		fmt.Printf("%-20s %-15s %-10s %-15s %-30s\n", 
			columnName.String, dataType.String, isNullable.String, defaultVal, comment)
	}

	fmt.Println("\n🎉 Migration completed successfully!")
	fmt.Println("✅ ai_whatsapp_nodepath table now supports flow tracking")
	fmt.Println("✅ Added fields: current_node_id, waiting_for_reply, flow_id, last_node_id")
}