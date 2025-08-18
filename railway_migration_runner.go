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

// createMissingTables creates all missing tables in the database
func createMissingTables(db *sql.DB) error {
	log.Println("📋 Creating missing tables...")

	// Define all required tables
	tables := []struct {
		name string
		sql  string
	}{
		{
			name: "conversation_log_nodepath",
			sql: `CREATE TABLE IF NOT EXISTS conversation_log_nodepath (
				id INT AUTO_INCREMENT PRIMARY KEY,
				prospect_num VARCHAR(255) NOT NULL,
				id_staff VARCHAR(255) NOT NULL,
				message TEXT NOT NULL,
				sender VARCHAR(10) NOT NULL COMMENT 'user or bot',
				stage VARCHAR(255) DEFAULT NULL,
				timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				
				INDEX idx_prospect_num (prospect_num),
				INDEX idx_id_staff (id_staff),
				INDEX idx_sender (sender),
				INDEX idx_stage (stage),
				INDEX idx_timestamp (timestamp),
				INDEX idx_created_at (created_at)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		},
		{
			name: "ai_settings_nodepath",
			sql: `CREATE TABLE IF NOT EXISTS ai_settings_nodepath (
				id INT AUTO_INCREMENT PRIMARY KEY,
				id_staff VARCHAR(255) NOT NULL,
				system_prompt TEXT DEFAULT NULL,
				closing_prompt TEXT DEFAULT NULL,
				instance_prompt TEXT DEFAULT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				
				INDEX idx_id_staff (id_staff)
			) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		},
	}

	// Create each table
	for _, table := range tables {
		log.Printf("🔄 Creating table: %s", table.name)
		
		// Check if table exists
		var tableExists bool
		err := db.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table.name).Scan(&tableExists)
		if err != nil {
			log.Printf("⚠️ Failed to check if table %s exists: %v", table.name, err)
			continue
		}

		if tableExists {
			log.Printf("⏭️ Table %s already exists, skipping", table.name)
			continue
		}

		// Create the table
		_, err = db.Exec(table.sql)
		if err != nil {
			log.Printf("❌ Failed to create table %s: %v", table.name, err)
			return fmt.Errorf("failed to create table %s: %w", table.name, err)
		}
		log.Printf("✅ Successfully created table: %s", table.name)
	}

	log.Println("✅ All missing tables created successfully!")
	return nil
}

func main() {
	log.Println("🚀 Railway Migration Runner - Comprehensive Database Migration")

	// Get MySQL URI from environment (Railway provides this)
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Println("❌ MYSQL_URI environment variable not found")
		log.Println("ℹ️ This script should be run in Railway environment")
		os.Exit(1)
	}

	log.Printf("📡 Connecting to Railway database...")

	// Convert URI to DSN format
	dsn := convertMySQLURI(mysqlURI)

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

	// Create missing tables first
	if err := createMissingTables(db); err != nil {
		log.Printf("❌ Failed to create missing tables: %v", err)
		os.Exit(1)
	}

	// Check if table exists
	log.Println("📋 Checking if ai_whatsapp_nodepath table exists...")
	var tableExists bool
	err = db.QueryRow("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'ai_whatsapp_nodepath'").Scan(&tableExists)
	if err != nil {
		log.Fatalf("❌ Failed to check table existence: %v", err)
	}

	if !tableExists {
		log.Println("❌ Table ai_whatsapp_nodepath does not exist!")
		os.Exit(1)
	}
	log.Println("✅ Table ai_whatsapp_nodepath exists")

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

	// Define columns to add with their SQL statements
	columnsToAdd := []struct {
		name string
		sql  string
	}{
		{"jam", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN jam VARCHAR(255) DEFAULT NULL COMMENT 'Jam field for AI WhatsApp conversations'"},
		{"intro", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN intro VARCHAR(255) DEFAULT NULL COMMENT 'Introduction field'"},
		{"date_order", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN date_order DATETIME DEFAULT NULL COMMENT 'Order date field'"},
		{"balas", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN balas VARCHAR(255) DEFAULT NULL COMMENT 'Reply field'"},
		{"data_image", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN data_image TEXT DEFAULT NULL COMMENT 'Image data field'"},
		{"conv_stage", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN conv_stage VARCHAR(100) DEFAULT NULL COMMENT 'Conversation stage field'"},
		{"keywordiklan", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN keywordiklan VARCHAR(255) DEFAULT NULL COMMENT 'Advertisement keyword field'"},
		{"marketer", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN marketer VARCHAR(255) DEFAULT NULL COMMENT 'Marketer field'"},
		{"update_today", "ALTER TABLE ai_whatsapp_nodepath ADD COLUMN update_today TINYINT(1) DEFAULT 0 COMMENT 'Update today flag'"},
	}

	// Add missing columns
	addedColumns := []string{}
	skippedColumns := []string{}

	for _, column := range columnsToAdd {
		if existingColumns[column.name] {
			log.Printf("⏭️ Column '%s' already exists, skipping", column.name)
			skippedColumns = append(skippedColumns, column.name)
			continue
		}

		log.Printf("➕ Adding column: %s", column.name)
		_, err := db.Exec(column.sql)
		if err != nil {
			log.Printf("❌ Failed to add column '%s': %v", column.name, err)
		} else {
			log.Printf("✅ Successfully added column: %s", column.name)
			addedColumns = append(addedColumns, column.name)
		}
	}

	// Fix data types for existing columns
	log.Println("🔧 Fixing data types for existing columns...")

	// Fix id_prospect data type
	if existingColumns["id_prospect"] {
		log.Println("🔄 Modifying id_prospect to INT...")
		_, err = db.Exec("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN id_prospect INT DEFAULT NULL COMMENT 'Prospect ID as integer'")
		if err != nil {
			log.Printf("⚠️ Could not modify id_prospect: %v", err)
		} else {
			log.Println("✅ Successfully modified id_prospect to INT")
		}
	}

	// Fix bot_balas data type
	if existingColumns["bot_balas"] {
		log.Println("🔄 Modifying bot_balas to TIMESTAMP...")
		_, err = db.Exec("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN bot_balas TIMESTAMP NULL DEFAULT NULL COMMENT 'Bot reply timestamp'")
		if err != nil {
			log.Printf("⚠️ Could not modify bot_balas: %v", err)
		} else {
			log.Println("✅ Successfully modified bot_balas to TIMESTAMP")
		}
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
	log.Printf("✅ Added columns: %v (%d)", addedColumns, len(addedColumns))
	log.Printf("⏭️ Skipped columns: %v (%d)", skippedColumns, len(skippedColumns))

	if len(addedColumns) > 0 {
		log.Println("🚀 Migration completed successfully! New columns added.")
	} else {
		log.Println("ℹ️ No new columns were added (all already exist).")
	}

	// Test if jam column exists now
	log.Println("\n🧪 Testing jam column access...")
	_, err = db.Query("SELECT jam FROM ai_whatsapp_nodepath LIMIT 1")
	if err != nil {
		log.Printf("❌ Jam column test failed: %v", err)
		os.Exit(1)
	} else {
		log.Println("✅ Jam column is accessible!")
	}

	// Test if conversation_log_nodepath table exists and is accessible
	log.Println("🧪 Testing conversation_log_nodepath table access...")
	_, err = db.Query("SELECT id_staff FROM conversation_log_nodepath LIMIT 1")
	if err != nil {
		log.Printf("❌ conversation_log_nodepath table test failed: %v", err)
		os.Exit(1)
	} else {
		log.Println("✅ conversation_log_nodepath table is accessible!")
	}

	log.Println("\n🎉 Comprehensive migration completed successfully!")
}