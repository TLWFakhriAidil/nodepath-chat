package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// convertMySQLURI converts mysql:// URI to DSN format
func convertMySQLURI(uri string) string {
	if !strings.HasPrefix(uri, "mysql://") {
		return uri
	}

	// Remove mysql:// prefix
	uri = strings.TrimPrefix(uri, "mysql://")

	// Split into parts
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return uri
	}

	connectionPart := parts[0]
	database := parts[1]

	// Split connection part into user:pass@host:port
	atIndex := strings.LastIndex(connectionPart, "@")
	if atIndex == -1 {
		return uri
	}

	userPass := connectionPart[:atIndex]
	hostPort := connectionPart[atIndex+1:]

	// Format as DSN
	dsn := fmt.Sprintf("%s@tcp(%s)/%s?parseTime=true&timeout=30s", userPass, hostPort, database)
	return dsn
}

func main() {
	log.Println("🚀 Starting direct migration to Railway production database...")

	// Get MySQL URI from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
		log.Println("⚠️ Using default MySQL URI")
	}

	log.Printf("📡 Connecting to: %s", strings.Replace(mysqlURI, "admin_aqil", "***", -1))

	// Convert URI to DSN format
	dsn := convertMySQLURI(mysqlURI)
	log.Printf("🔗 DSN: %s", strings.Replace(dsn, "admin_aqil", "***", -1))

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test connection
	log.Println("🔍 Testing database connection...")
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}
	log.Println("✅ Database connection successful!")

	// Check current table structure
	log.Println("📋 Checking current table structure...")
	rows, err := db.Query("DESCRIBE ai_whatsapp_nodepath")
	if err != nil {
		log.Fatalf("❌ Failed to describe table: %v", err)
	}

	existingColumns := make(map[string]bool)
	for rows.Next() {
		var field, fieldType, null, key, defaultVal, extra sql.NullString
		if err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra); err != nil {
			log.Printf("⚠️ Error scanning row: %v", err)
			continue
		}
		if field.Valid {
			existingColumns[field.String] = true
			log.Printf("📝 Found column: %s (%s)", field.String, fieldType.String)
		}
	}
	rows.Close()

	// Define columns to add
	columnsToAdd := map[string]string{
		"jam":          "ADD COLUMN jam VARCHAR(255) DEFAULT NULL COMMENT 'Jam field for AI WhatsApp conversations'",
		"intro":        "ADD COLUMN intro VARCHAR(255) DEFAULT NULL COMMENT 'Introduction field'",
		"date_order":   "ADD COLUMN date_order DATETIME DEFAULT NULL COMMENT 'Order date field'",
		"balas":        "ADD COLUMN balas VARCHAR(255) DEFAULT NULL COMMENT 'Reply field'",
		"data_image":   "ADD COLUMN data_image TEXT DEFAULT NULL COMMENT 'Image data field'",
		"conv_stage":   "ADD COLUMN conv_stage VARCHAR(100) DEFAULT NULL COMMENT 'Conversation stage field'",
		"keywordiklan": "ADD COLUMN keywordiklan VARCHAR(255) DEFAULT NULL COMMENT 'Advertisement keyword field'",
		"marketer":     "ADD COLUMN marketer VARCHAR(255) DEFAULT NULL COMMENT 'Marketer field'",
		"update_today": "ADD COLUMN update_today TINYINT(1) DEFAULT 0 COMMENT 'Update today flag'",
	}

	// Add missing columns
	addedColumns := []string{}
	skippedColumns := []string{}

	for columnName, alterStatement := range columnsToAdd {
		if existingColumns[columnName] {
			log.Printf("⏭️ Column '%s' already exists, skipping", columnName)
			skippedColumns = append(skippedColumns, columnName)
			continue
		}

		log.Printf("➕ Adding column: %s", columnName)
		query := fmt.Sprintf("ALTER TABLE ai_whatsapp_nodepath %s", alterStatement)
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("❌ Failed to add column '%s': %v", columnName, err)
		} else {
			log.Printf("✅ Successfully added column: %s", columnName)
			addedColumns = append(addedColumns, columnName)
		}
	}

	// Fix data types for existing columns
	log.Println("🔧 Fixing data types for existing columns...")

	// Fix id_prospect data type
	log.Println("🔄 Modifying id_prospect to INT...")
	_, err = db.Exec("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN id_prospect INT DEFAULT NULL COMMENT 'Prospect ID as integer'")
	if err != nil {
		log.Printf("⚠️ Could not modify id_prospect: %v", err)
	} else {
		log.Println("✅ Successfully modified id_prospect to INT")
	}

	// Fix bot_balas data type
	log.Println("🔄 Modifying bot_balas to TIMESTAMP...")
	_, err = db.Exec("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN bot_balas TIMESTAMP NULL DEFAULT NULL COMMENT 'Bot reply timestamp'")
	if err != nil {
		log.Printf("⚠️ Could not modify bot_balas: %v", err)
	} else {
		log.Println("✅ Successfully modified bot_balas to TIMESTAMP")
	}

	// Verify final table structure
	log.Println("🔍 Verifying final table structure...")
	rows, err = db.Query("DESCRIBE ai_whatsapp_nodepath")
	if err != nil {
		log.Printf("⚠️ Failed to verify table structure: %v", err)
	} else {
		finalColumns := []string{}
		for rows.Next() {
			var field, fieldType, null, key, defaultVal, extra sql.NullString
			if err := rows.Scan(&field, &fieldType, &null, &key, &defaultVal, &extra); err != nil {
				continue
			}
			if field.Valid {
				finalColumns = append(finalColumns, field.String)
			}
		}
		rows.Close()
		log.Printf("📊 Final table has %d columns: %v", len(finalColumns), finalColumns)
	}

	// Summary
	log.Println("\n🎉 Migration Summary:")
	log.Printf("✅ Added columns: %v", addedColumns)
	log.Printf("⏭️ Skipped columns: %v", skippedColumns)
	log.Println("🚀 Migration completed successfully!")

	// Test if jam column exists now
	log.Println("\n🧪 Testing jam column access...")
	_, err = db.Query("SELECT jam FROM ai_whatsapp_nodepath LIMIT 1")
	if err != nil {
		log.Printf("❌ Jam column test failed: %v", err)
	} else {
		log.Println("✅ Jam column is accessible!")
	}
}