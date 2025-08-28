package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is not set")
	}

	// Convert mysql:// format to Go driver format
	// mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway
	// becomes: admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local
	dsn := "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== AI WhatsApp Record Check ===")
	fmt.Println("Checking ai_whatsapp_nodepath for device FakhriAidilTLW-001 and phone 60179645043...")
	fmt.Println()

	// Query the specific record
	query := `SELECT 
		id_prospect,
		prospect_num,
		id_device,
		current_node_id,
		waiting_for_reply,
		flow_id,
		execution_status,
		execution_id,
		last_node_id,
		created_at,
		updated_at
	FROM ai_whatsapp_nodepath 
	WHERE id_device = ? AND prospect_num = ?`

	rows, err := db.Query(query, "FakhriAidilTLW-001", "60179645043")
	if err != nil {
		log.Fatal("Failed to execute query:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var idProspect sql.NullInt64
		var prospectNum, idDevice sql.NullString
		var currentNodeID, flowID, executionStatus, executionID, lastNodeID sql.NullString
		var waitingForReply sql.NullInt64
		var createdAt, updatedAt sql.NullString

		err := rows.Scan(
			&idProspect,
			&prospectNum,
			&idDevice,
			&currentNodeID,
			&waitingForReply,
			&flowID,
			&executionStatus,
			&executionID,
			&lastNodeID,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		fmt.Printf("ID Prospect: %v\n", getValue(idProspect))
		fmt.Printf("ProspectNum: %v\n", getValue(prospectNum))
		fmt.Printf("IDDevice: %v\n", getValue(idDevice))
		fmt.Printf("CurrentNodeID: %v\n", getValue(currentNodeID))
		fmt.Printf("WaitingForReply: %v\n", getValue(waitingForReply))
		fmt.Printf("FlowID: %v\n", getValue(flowID))
		fmt.Printf("ExecutionStatus: %v\n", getValue(executionStatus))
		fmt.Printf("ExecutionID: %v\n", getValue(executionID))
		fmt.Printf("LastNodeID: %v\n", getValue(lastNodeID))
		fmt.Printf("CreatedAt: %v\n", getValue(createdAt))
		fmt.Printf("UpdatedAt: %v\n", getValue(updatedAt))
		fmt.Println()
	}

	if !found {
		fmt.Println("No record found for device FakhriAidilTLW-001 and phone 60179645043")
	}

	if err := rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
	}
}

func getValue(v interface{}) string {
	switch val := v.(type) {
	case sql.NullString:
		if val.Valid {
			return val.String
		}
		return "NULL"
	case sql.NullInt64:
		if val.Valid {
			return fmt.Sprintf("%d", val.Int64)
		}
		return "NULL"
	default:
		return "NULL"
	}
}