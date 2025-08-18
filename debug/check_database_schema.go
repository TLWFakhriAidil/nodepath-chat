package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// checkTableSchema checks the schema of a specific table
func checkTableSchema(db *sql.DB, tableName string) error {
	fmt.Printf("\n=== Checking schema for table: %s ===\n", tableName)
	
	// Check if table exists
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", tableName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking table existence: %v", err)
	}
	
	if exists == 0 {
		fmt.Printf("❌ Table %s does not exist\n", tableName)
		return nil
	}
	
	fmt.Printf("✅ Table %s exists\n", tableName)
	
	// Get table structure
	rows, err := db.Query("DESCRIBE " + tableName)
	if err != nil {
		return fmt.Errorf("error describing table: %v", err)
	}
	defer rows.Close()
	
	fmt.Println("\nTable Structure:")
	fmt.Println("Field\t\t\tType\t\t\tNull\tKey\tDefault\tExtra")
	fmt.Println("-----\t\t\t----\t\t\t----\t---\t-------\t-----")
	
	hasIDStaff := false
	hasIDDevice := false
	
	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			return fmt.Errorf("error scanning row: %v", err)
		}
		
		// Check for id_staff or id_device columns
		if field.String == "id_staff" {
			hasIDStaff = true
		}
		if field.String == "id_device" {
			hasIDDevice = true
		}
		
		fmt.Printf("%-20s\t%-20s\t%s\t%s\t%s\t%s\n",
			field.String,
			fieldType.String,
			null.String,
			key.String,
			defaultVal.String,
			extra.String)
	}
	
	// Report findings
	fmt.Println("\nColumn Analysis:")
	if hasIDStaff {
		fmt.Printf("⚠️  Found id_staff column - NEEDS MIGRATION\n")
	}
	if hasIDDevice {
		fmt.Printf("✅ Found id_device column\n")
	}
	if !hasIDStaff && !hasIDDevice {
		fmt.Printf("ℹ️  No id_staff or id_device columns found\n")
	}
	
	return nil
}

// convertMySQLURL converts mysql:// URL to DSN format
func convertMySQLURL(mysqlURL string) string {
	if !strings.HasPrefix(mysqlURL, "mysql://") {
		return mysqlURL
	}
	
	// Remove mysql:// prefix
	dsn := strings.TrimPrefix(mysqlURL, "mysql://")
	
	// Replace the first @ with @ and add tcp() wrapper for host:port
	parts := strings.Split(dsn, "@")
	if len(parts) == 2 {
		userPass := parts[0]
		hostDB := parts[1]
		
		// Split host and database
		hostParts := strings.Split(hostDB, "/")
		if len(hostParts) == 2 {
			host := hostParts[0]
			db := hostParts[1]
			return fmt.Sprintf("%s@tcp(%s)/%s", userPass, host, db)
		}
	}
	
	return dsn
}

func main() {
	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		// Fallback to default connection for local testing
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
		fmt.Println("Using default MySQL URI for testing")
	}
	
	fmt.Printf("Connecting to database...\n")
	
	// Convert MySQL URL to DSN format
	dsn := convertMySQLURL(mysqlURI)
	
	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()
	
	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	
	fmt.Println("✅ Database connection successful")
	
	// Tables to check
	tables := []string{
		"ai_whatsapp_nodepath",
		"conversation_log_nodepath",
		"conversation_log_nodepath_backup",
	}
	
	// Check each table
	for _, table := range tables {
		err := checkTableSchema(db, table)
		if err != nil {
			fmt.Printf("❌ Error checking table %s: %v\n", table, err)
		}
	}
	
	fmt.Println("\n=== Schema Check Complete ===")
}