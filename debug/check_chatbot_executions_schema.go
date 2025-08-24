package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// convertMySQLURLToDSN converts a MySQL URL to a DSN format
func convertMySQLURLToDSN(mysqlURL string) (string, error) {
	u, err := url.Parse(mysqlURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse MySQL URL: %w", err)
	}

	// Extract user info
	username := u.User.Username()
	password, _ := u.User.Password()

	// Extract host and port
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "3306" // Default MySQL port
	}

	// Extract database name
	dbName := strings.TrimPrefix(u.Path, "/")

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", username, password, host, port, dbName)
	return dsn, nil
}

func main() {
	// Get database URL from environment or use default
	mysqlURL := os.Getenv("MYSQL_URI")
	if mysqlURL == "" {
		mysqlURL = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
	}

	// Convert MySQL URL to DSN
	dsn, err := convertMySQLURLToDSN(mysqlURL)
	if err != nil {
		log.Fatalf("Failed to convert MySQL URL to DSN: %v", err)
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

	fmt.Println("Connected to database successfully!")
	fmt.Println("\n=== Checking chatbot_executions_nodepath table schema ===")

	// Check if table exists and get its structure
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'chatbot_executions_nodepath'
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to query table schema: %v", err)
	}
	defer rows.Close()

	fmt.Println("\nTable structure:")
	fmt.Println("Column Name\t\tData Type\tNullable\tDefault\t\tExtra")
	fmt.Println("----------\t\t---------\t--------\t-------\t\t-----")

	hasStaffID := false
	hasIDDevice := false

	for rows.Next() {
		var columnName, dataType, isNullable, extra string
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault, &extra)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		defaultValue := "NULL"
		if columnDefault.Valid {
			defaultValue = columnDefault.String
		}

		fmt.Printf("%-20s\t%-12s\t%-8s\t%-15s\t%s\n", columnName, dataType, isNullable, defaultValue, extra)

		// Check for staff_id and id_device columns
		if columnName == "staff_id" {
			hasStaffID = true
		}
		if columnName == "id_device" {
			hasIDDevice = true
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	fmt.Println("\n=== Migration Status ===")
	if hasStaffID {
		fmt.Println("✗ staff_id column found - MIGRATION NEEDED")
	} else {
		fmt.Println("✓ staff_id column not found")
	}

	if hasIDDevice {
		fmt.Println("✓ id_device column found")
	} else {
		fmt.Println("✗ id_device column not found")
	}

	if hasStaffID && !hasIDDevice {
		fmt.Println("\n>>> ACTION REQUIRED: Need to rename staff_id to id_device <<<")
	} else if hasStaffID && hasIDDevice {
		fmt.Println("\n>>> ACTION REQUIRED: Both columns exist, need to migrate data and drop staff_id <<<")
	} else if !hasStaffID && hasIDDevice {
		fmt.Println("\n>>> MIGRATION COMPLETE: Only id_device column exists <<<")
	} else {
		fmt.Println("\n>>> WARNING: Neither column found - table may not exist <<<")
	}
}