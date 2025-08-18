package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// fixDatabaseSchema ensures the ai_whatsapp_nodepath table has the correct schema
func fixDatabaseSchema() error {
	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		return fmt.Errorf("MYSQL_URI environment variable not set")
	}

	// Connect to database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Database connection successful")

	// Read the SQL fix file
	sqlContent, err := ioutil.ReadFile("fix_jam_column.sql")
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %w", err)
	}

	// Execute the SQL fix
	_, err = db.Exec(string(sqlContent))
	if err != nil {
		return fmt.Errorf("failed to execute SQL fix: %w", err)
	}

	log.Println("✅ Database schema fix applied successfully")

	// Verify the table structure
	rows, err := db.Query(`
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_DEFAULT
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = 'ai_whatsapp_nodepath' 
		ORDER BY ORDINAL_POSITION
	`)
	if err != nil {
		return fmt.Errorf("failed to query table structure: %w", err)
	}
	defer rows.Close()

	log.Println("\n📋 Current ai_whatsapp_nodepath table structure:")
	log.Println("Column Name\t\tData Type\tNullable\tDefault")
	log.Println("--------------------------------------------------------")

	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		defaultValue := "NULL"
		if columnDefault.Valid {
			defaultValue = columnDefault.String
		}

		log.Printf("%-20s\t%-15s\t%-8s\t%s", columnName, dataType, isNullable, defaultValue)
	}

	return nil
}

func main() {
	log.Println("🔧 Starting database schema fix...")

	if err := fixDatabaseSchema(); err != nil {
		log.Fatalf("❌ Database schema fix failed: %v", err)
	}

	log.Println("\n✅ Database schema fix completed successfully!")
	log.Println("The 'jam' column should now be available in ai_whatsapp_nodepath table.")
}