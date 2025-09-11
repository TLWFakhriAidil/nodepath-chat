package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Column definition for expected table structure
type ColumnDef struct {
	Name       string
	DataType   string
	Nullable   bool
	Default    *string
	Comment    string
}

// Expected columns based on AIWhatsapp model
var expectedColumns = []ColumnDef{
	{Name: "id_prospect", DataType: "INT(11)", Nullable: false, Comment: "Primary key"},
	{Name: "flow_reference", DataType: "VARCHAR(255)", Nullable: true, Comment: "Reference to chatbot flow being executed"},
	{Name: "execution_id", DataType: "VARCHAR(255)", Nullable: true, Comment: "Unique execution identifier"},
	{Name: "date_order", DataType: "DATETIME", Nullable: true, Comment: "Order date"},
	{Name: "id_device", DataType: "VARCHAR(255)", Nullable: true, Comment: "Device identifier"},
	{Name: "niche", DataType: "VARCHAR(255)", Nullable: true, Comment: "Niche category"},
	{Name: "prospect_name", DataType: "VARCHAR(255)", Nullable: true, Comment: "Prospect name"},
	{Name: "prospect_num", DataType: "VARCHAR(255)", Nullable: true, Comment: "Prospect phone number"},
	{Name: "intro", DataType: "VARCHAR(255)", Nullable: true, Comment: "Introduction text"},
	{Name: "stage", DataType: "VARCHAR(255)", Nullable: true, Comment: "Current stage"},
	{Name: "conv_last", DataType: "TEXT", Nullable: true, Comment: "Last conversation"},
	{Name: "conv_current", DataType: "TEXT", Nullable: true, Comment: "Current conversation"},
	{Name: "execution_status", DataType: "ENUM('active','completed','failed')", Nullable: true, Comment: "Flow execution status"},
	{Name: "flow_id", DataType: "VARCHAR(255)", Nullable: true, Comment: "ID of the current chatbot flow"},
	{Name: "current_node_id", DataType: "VARCHAR(255)", Nullable: true, Comment: "Current node ID in the chatbot flow"},
	{Name: "last_node_id", DataType: "VARCHAR(255)", Nullable: true, Comment: "Previous node ID for flow tracking"},
	{Name: "waiting_for_reply", DataType: "TINYINT(1)", Nullable: true, Default: stringPtr("0"), Comment: "1 = waiting for user reply"},
	{Name: "balas", DataType: "VARCHAR(255)", Nullable: true, Comment: "Reply field"},
	{Name: "human", DataType: "INT(11)", Nullable: true, Default: stringPtr("0"), Comment: "0 = AI active, 1 = human takeover"},
	{Name: "keywordiklan", DataType: "VARCHAR(255)", Nullable: true, Comment: "Advertisement keyword"},
	{Name: "marketer", DataType: "VARCHAR(255)", Nullable: true, Comment: "Marketer name"},
	{Name: "created_at", DataType: "TIMESTAMP", Nullable: false, Default: stringPtr("CURRENT_TIMESTAMP"), Comment: "Creation timestamp"},
	{Name: "updated_at", DataType: "TIMESTAMP", Nullable: false, Default: stringPtr("CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"), Comment: "Update timestamp"},
	{Name: "update_today", DataType: "DATETIME", Nullable: true, Comment: "Today update timestamp"},
}

// Columns to drop (not in the model)
var columnsToRemove = []string{
	"jam",
	"catatan_staff",
	"data_image",
	"conv_stage",
	"variables",
	"bot_balas",
	"current_node",
}

func stringPtr(s string) *string {
	return &s
}

func connectDB() (*sql.DB, error) {
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		return nil, fmt.Errorf("MYSQL_URI environment variable is not set")
	}

	// Convert mysql:// URL to DSN format
	if strings.HasPrefix(mysqlURI, "mysql://") {
		mysqlURI = convertMySQLURIToDSN(mysqlURI)
	}

	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func convertMySQLURIToDSN(uri string) string {
	uri = strings.TrimPrefix(uri, "mysql://")
	parts := strings.Split(uri, "@")
	if len(parts) != 2 {
		return uri
	}
	userPass := parts[0]
	hostDB := parts[1]
	hostDB = strings.Replace(hostDB, "/", ")/", 1)
	return userPass + "@tcp(" + hostDB + "?parseTime=true&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci"
}

func getExistingColumns(db *sql.DB) (map[string]bool, error) {
	query := `
		SELECT COLUMN_NAME 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'ai_whatsapp_nodepath'
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	columns := make(map[string]bool)
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, err
		}
		columns[colName] = true
	}
	
	return columns, nil
}

func dropColumn(db *sql.DB, columnName string) error {
	query := fmt.Sprintf("ALTER TABLE ai_whatsapp_nodepath DROP COLUMN `%s`", columnName)
	log.Printf("Executing: %s\n", query)
	
	_, err := db.Exec(query)
	if err != nil {
		// Check if column doesn't exist (error 1091)
		if strings.Contains(err.Error(), "1091") || strings.Contains(err.Error(), "Can't DROP") {
			log.Printf("Column %s doesn't exist (skipping)\n", columnName)
			return nil
		}
		return fmt.Errorf("failed to drop column %s: %w", columnName, err)
	}
	
	log.Printf("✅ Successfully dropped column: %s\n", columnName)
	return nil
}

func addColumn(db *sql.DB, col ColumnDef) error {
	var query strings.Builder
	query.WriteString(fmt.Sprintf("ALTER TABLE ai_whatsapp_nodepath ADD COLUMN `%s` %s", col.Name, col.DataType))
	
	if !col.Nullable {
		query.WriteString(" NOT NULL")
	} else {
		query.WriteString(" NULL")
	}
	
	if col.Default != nil {
		if strings.Contains(*col.Default, "CURRENT_TIMESTAMP") {
			query.WriteString(fmt.Sprintf(" DEFAULT %s", *col.Default))
		} else {
			query.WriteString(fmt.Sprintf(" DEFAULT '%s'", *col.Default))
		}
	}
	
	if col.Comment != "" {
		query.WriteString(fmt.Sprintf(" COMMENT '%s'", col.Comment))
	}
	
	log.Printf("Executing: %s\n", query.String())
	
	_, err := db.Exec(query.String())
	if err != nil {
		// Check if column already exists (error 1060)
		if strings.Contains(err.Error(), "1060") || strings.Contains(err.Error(), "Duplicate column") {
			log.Printf("Column %s already exists (skipping)\n", col.Name)
			return nil
		}
		return fmt.Errorf("failed to add column %s: %w", col.Name, err)
	}
	
	log.Printf("✅ Successfully added column: %s\n", col.Name)
	return nil
}

func modifyColumn(db *sql.DB, col ColumnDef) error {
	var query strings.Builder
	query.WriteString(fmt.Sprintf("ALTER TABLE ai_whatsapp_nodepath MODIFY COLUMN `%s` %s", col.Name, col.DataType))
	
	if !col.Nullable {
		query.WriteString(" NOT NULL")
	} else {
		query.WriteString(" NULL")
	}
	
	if col.Default != nil {
		if strings.Contains(*col.Default, "CURRENT_TIMESTAMP") {
			query.WriteString(fmt.Sprintf(" DEFAULT %s", *col.Default))
		} else {
			query.WriteString(fmt.Sprintf(" DEFAULT '%s'", *col.Default))
		}
	}
	
	if col.Comment != "" {
		query.WriteString(fmt.Sprintf(" COMMENT '%s'", col.Comment))
	}
	
	log.Printf("Executing: %s\n", query.String())
	
	_, err := db.Exec(query.String())
	if err != nil {
		return fmt.Errorf("failed to modify column %s: %w", col.Name, err)
	}
	
	log.Printf("✅ Successfully modified column: %s\n", col.Name)
	return nil
}

func main() {
	log.Println("🔧 Starting AI WhatsApp Table Column Migration")
	log.Println("================================================")
	
	// Connect to database
	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	log.Println("✅ Connected to database successfully")
	
	// Get existing columns
	existingColumns, err := getExistingColumns(db)
	if err != nil {
		log.Fatalf("❌ Failed to get existing columns: %v", err)
	}
	
	log.Printf("📊 Found %d existing columns in ai_whatsapp_nodepath\n", len(existingColumns))
	
	// Step 1: Drop unwanted columns
	log.Println("\n🗑️  Step 1: Dropping unwanted columns...")
	log.Println("----------------------------------------")
	for _, colName := range columnsToRemove {
		if existingColumns[colName] {
			if err := dropColumn(db, colName); err != nil {
				log.Printf("⚠️  Warning: %v\n", err)
			}
		} else {
			log.Printf("Column %s doesn't exist (skipping)\n", colName)
		}
	}
	
	// Step 2: Add missing columns
	log.Println("\n➕ Step 2: Adding missing columns...")
	log.Println("------------------------------------")
	for _, col := range expectedColumns {
		if !existingColumns[col.Name] {
			if err := addColumn(db, col); err != nil {
				log.Printf("⚠️  Warning: %v\n", err)
			}
		} else {
			log.Printf("Column %s already exists (checking data type)\n", col.Name)
			// Optionally modify column if it exists but might have wrong data type
			// Uncomment if you want to ensure data types match
			// if err := modifyColumn(db, col); err != nil {
			//     log.Printf("⚠️  Warning: %v\n", err)
			// }
		}
	}
	
	// Step 3: Verify final structure
	log.Println("\n✅ Step 3: Verifying final table structure...")
	log.Println("---------------------------------------------")
	
	finalColumns, err := getExistingColumns(db)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to verify final structure: %v\n", err)
	} else {
		log.Printf("📊 Final table has %d columns\n", len(finalColumns))
		
		// Check for expected columns
		missingColumns := []string{}
		for _, col := range expectedColumns {
			if !finalColumns[col.Name] {
				missingColumns = append(missingColumns, col.Name)
			}
		}
		
		if len(missingColumns) > 0 {
			log.Printf("⚠️  Missing columns: %v\n", missingColumns)
		} else {
			log.Println("✅ All expected columns are present!")
		}
		
		// Check for unwanted columns still present
		unwantedPresent := []string{}
		for _, colName := range columnsToRemove {
			if finalColumns[colName] {
				unwantedPresent = append(unwantedPresent, colName)
			}
		}
		
		if len(unwantedPresent) > 0 {
			log.Printf("⚠️  Unwanted columns still present: %v\n", unwantedPresent)
		} else {
			log.Println("✅ All unwanted columns have been removed!")
		}
	}
	
	// Show final table structure
	log.Println("\n📋 Final Table Structure:")
	log.Println("------------------------")
	rows, err := db.Query("DESCRIBE ai_whatsapp_nodepath")
	if err != nil {
		log.Printf("⚠️  Failed to describe table: %v\n", err)
	} else {
		defer rows.Close()
		
		for rows.Next() {
			var field, colType, null, key, defaultVal, extra sql.NullString
			if err := rows.Scan(&field, &colType, &null, &key, &defaultVal, &extra); err != nil {
				continue
			}
			log.Printf("  - %s: %s %s\n", field.String, colType.String, null.String)
		}
	}
	
	log.Println("\n✅ Migration completed successfully!")
	log.Println("=====================================")
}
