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
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database connection string
	mysqlURI := os.Getenv("DATABASE_URL")
	if mysqlURI == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully!")

	// Query the ai_whatsapp_nodepath table for our test data
	query := `
		SELECT id_prospect, prospect_num, prospect_name, id_device, stage, 
		       SUBSTRING(conv_last, 1, 100) as conv_last_preview,
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = '601137508067' 
		ORDER BY id_prospect DESC 
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	fmt.Println("\n=== AI WhatsApp Records for Phone: 601137508067 ===")
	fmt.Printf("%-12s %-15s %-20s %-20s %-20s %-30s\n", 
		"ID", "Phone", "Prospect Name", "Device ID", "Stage", "Conv Preview")
	fmt.Println(strings.Repeat("-", 120))

	recordCount := 0
	for rows.Next() {
		var idProspect int
		var prospectNum, prospectName, idDevice, stage, convLastPreview string
		var createdAt, updatedAt string

		err := rows.Scan(&idProspect, &prospectNum, &prospectName, &idDevice, &stage, &convLastPreview, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		// Handle NULL values
		if prospectName == "" {
			prospectName = "<NULL>"
		}
		if stage == "" {
			stage = "<NULL>"
		}
		if convLastPreview == "" {
			convLastPreview = "<NULL>"
		}

		fmt.Printf("%-12d %-15s %-20s %-20s %-20s %-30s\n", 
			idProspect, prospectNum, prospectName, idDevice, stage, convLastPreview)
		fmt.Printf("  Created: %s, Updated: %s\n", createdAt, updatedAt)
		fmt.Println()

		recordCount++
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	if recordCount == 0 {
		fmt.Println("No records found for phone number 601137508067")
	} else {
		fmt.Printf("Found %d record(s)\n", recordCount)
	}

	// Also check if prospect_name column exists and its structure
	fmt.Println("\n=== Table Structure Check ===")
	descQuery := "DESCRIBE ai_whatsapp_nodepath"
	descRows, err := db.Query(descQuery)
	if err != nil {
		log.Printf("Failed to describe table: %v", err)
		return
	}
	defer descRows.Close()

	fmt.Printf("%-20s %-15s %-10s %-10s %-10s %-10s\n", 
		"Field", "Type", "Null", "Key", "Default", "Extra")
	fmt.Println(strings.Repeat("-", 80))

	for descRows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString

		err := descRows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra)
		if err != nil {
			log.Printf("Error scanning describe row: %v", err)
			continue
		}

		fmt.Printf("%-20s %-15s %-10s %-10s %-10s %-10s\n", 
			field.String, fieldType.String, null.String, key.String, defaultVal.String, extra.String)
	}
}