package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// verifyColumnExists checks if a column exists in the specified table
func verifyColumnExists(db *sql.DB, tableName, columnName string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = ? 
		  AND COLUMN_NAME = ?
	`
	
	var count int
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// getColumnDataType retrieves the data type of a column
func getColumnDataType(db *sql.DB, tableName, columnName string) (string, error) {
	query := `
		SELECT DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		  AND TABLE_NAME = ? 
		  AND COLUMN_NAME = ?
	`
	
	var dataType string
	err := db.QueryRow(query, tableName, columnName).Scan(&dataType)
	if err != nil {
		return "", err
	}
	
	return dataType, nil
}

func main() {
	// Get database URL from environment - using MYSQL_URI exclusively
	dbURL := os.Getenv("MYSQL_URI")
	if dbURL == "" {
		log.Fatal("MYSQL_URI environment variable is required")
	}

	// Connect to database
	db, err := sql.Open("mysql", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🔍 Verifying ai_whatsapp_nodepath table schema...")
	fmt.Println("================================================")

	// List of columns that should exist
	requiredColumns := []struct {
		name         string
		expectedType string
	}{
		{"id", "int"},
		{"id_prospect", "int"},
		{"id_staff", "varchar"},
		{"prospect_num", "varchar"},
		{"stage", "varchar"},
		{"conv_last", "text"},
		{"conv_current", "text"},
		{"human", "int"},
		{"catatan_staff", "text"},
		{"bot_balas", "timestamp"},
		{"niche", "varchar"},
		{"created_at", "timestamp"},
		{"updated_at", "timestamp"},
		{"jam", "varchar"},           // Previously missing
		{"intro", "text"},            // Previously missing
		{"date_order", "date"},       // Previously missing
		{"balas", "text"},            // Previously missing
		{"data_image", "text"},       // Previously missing
		{"conv_stage", "varchar"},     // Previously missing
		{"keywordiklan", "varchar"},  // Previously missing
		{"marketer", "varchar"},      // Previously missing
		{"update_today", "int"},      // Previously missing
	}

	allColumnsExist := true
	missingColumns := []string{}
	typeIssues := []string{}

	// Check each required column
	for _, col := range requiredColumns {
		exists, err := verifyColumnExists(db, "ai_whatsapp_nodepath", col.name)
		if err != nil {
			log.Printf("❌ Error checking column %s: %v", col.name, err)
			continue
		}

		if !exists {
			fmt.Printf("❌ Column '%s' is MISSING\n", col.name)
			allColumnsExist = false
			missingColumns = append(missingColumns, col.name)
		} else {
			// Check data type
			actualType, err := getColumnDataType(db, "ai_whatsapp_nodepath", col.name)
			if err != nil {
				log.Printf("❌ Error checking data type for column %s: %v", col.name, err)
				continue
			}

			if actualType == col.expectedType {
				fmt.Printf("✅ Column '%s' exists with correct type (%s)\n", col.name, actualType)
			} else {
				fmt.Printf("⚠️  Column '%s' exists but has type '%s' (expected '%s')\n", col.name, actualType, col.expectedType)
				typeIssues = append(typeIssues, fmt.Sprintf("%s: %s (expected %s)", col.name, actualType, col.expectedType))
			}
		}
	}

	fmt.Println("\n================================================")
	fmt.Println("📊 VERIFICATION SUMMARY")
	fmt.Println("================================================")

	if allColumnsExist && len(typeIssues) == 0 {
		fmt.Println("🎉 SUCCESS: All required columns exist with correct data types!")
		fmt.Println("✅ The 'Unknown column jam' error should now be resolved.")
	} else {
		if len(missingColumns) > 0 {
			fmt.Printf("❌ MISSING COLUMNS (%d): %v\n", len(missingColumns), missingColumns)
		}
		if len(typeIssues) > 0 {
			fmt.Printf("⚠️  TYPE ISSUES (%d):\n", len(typeIssues))
			for _, issue := range typeIssues {
				fmt.Printf("   - %s\n", issue)
			}
		}
		fmt.Println("\n🔧 Migration may need to be re-run or checked for errors.")
	}

	fmt.Println("\n================================================")
}