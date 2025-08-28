package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(157.245.206.124:3306)/admin_railway"
	} else {
		// Convert from mysql:// format to Go driver format
		if strings.HasPrefix(mysqlURI, "mysql://") {
			mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
		}
	}

	// Connect to database
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Setting Flow State to AI Prompt Node ===")

	// Update flow state to AI prompt node
	prospectNum := "60179645043"
	idDevice := "FakhriAidilTLW-001"
	flowID := "flow_ai_1756016272"
	aiPromptNodeID := "prompt-1754883603395"

	// Update the flow execution to point to AI prompt node
	updateQuery := `
		UPDATE ai_whatsapp_nodepath 
		SET current_node_id = ?, 
		    flow_id = ?,
		    waiting_for_reply = 0
		WHERE prospect_num = ? AND id_device = ?
	`

	result, err := db.Exec(updateQuery, aiPromptNodeID, flowID, prospectNum, idDevice)
	if err != nil {
		log.Fatal("Failed to update flow state:", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal("Failed to get rows affected:", err)
	}

	fmt.Printf("✅ Updated %d row(s)\n", rowsAffected)
	fmt.Printf("Flow state set to AI prompt node: %s\n", aiPromptNodeID)
	fmt.Printf("Device: %s\n", idDevice)
	fmt.Printf("Prospect: %s\n", prospectNum)
	fmt.Printf("Flow: %s\n", flowID)

	// Verify the update
	verifyQuery := `
		SELECT current_node_id, flow_id, waiting_for_reply 
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = ? AND id_device = ?
	`

	var currentNodeID, currentFlowID sql.NullString
	var waitingForReply sql.NullBool

	err = db.QueryRow(verifyQuery, prospectNum, idDevice).Scan(
		&currentNodeID, &currentFlowID, &waitingForReply)
	if err != nil {
		log.Fatal("Failed to verify flow state:", err)
	}

	fmt.Println("\n=== Verification ===")
	fmt.Printf("Current Node ID: %s\n", currentNodeID.String)
	fmt.Printf("Flow ID: %s\n", currentFlowID.String)
	fmt.Printf("Waiting for Reply: %t\n", waitingForReply.Bool)

	fmt.Println("\n🎯 Ready to test AI prompt node! Run test_ai_prompt.go next.")
}