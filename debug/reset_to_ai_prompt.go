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
		mysqlURI = "mysql://admin_aqil:admin_aqil@157.245.206.124:3306/admin_railway"
	}

	// Parse the MySQL URI to get connection string
	// Format: mysql://user:password@host:port/database
	// Convert to: user:password@tcp(host:port)/database
	connStr := mysqlURI[8:] // Remove "mysql://" prefix
	// Replace @ with @tcp( and add ) before /
	if idx := strings.LastIndex(connStr, "@"); idx != -1 {
		if dbIdx := strings.LastIndex(connStr, "/"); dbIdx != -1 {
			userPass := connStr[:idx]
			hostPort := connStr[idx+1:dbIdx]
			dbName := connStr[dbIdx+1:]
			connStr = fmt.Sprintf("%s@tcp(%s)/%s", userPass, hostPort, dbName)
		}
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging database:", err)
	}

	fmt.Println("=== Resetting Flow State to Test AI Prompt ===")

	// Reset flow state to user_reply node and set waiting_for_reply to true
	updateQuery := `
		UPDATE ai_whatsapp_nodepath 
		SET 
			current_node_id = 'user_reply-1756105605330',
			waiting_for_reply = 1,
			execution_status = 'active',
			updated_at = NOW()
		WHERE prospect_num = '60179645043' 
		AND id_device = 'FakhriAidilTLW-001'
	`

	result, err := db.Exec(updateQuery)
	if err != nil {
		fmt.Printf("Error updating flow state: %v\n", err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("Error getting rows affected: %v\n", err)
		return
	}

	fmt.Printf("✅ Flow state reset successfully. Rows affected: %d\n", rowsAffected)

	// Verify the update
	fmt.Println("\n=== Verifying Flow State ===")
	verifyQuery := `
		SELECT current_node_id, waiting_for_reply, execution_status, flow_id 
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = '60179645043' AND id_device = 'FakhriAidilTLW-001'
	`

	var currentNodeID, execStatus, flowID sql.NullString
	var waitingForReply sql.NullBool
	err = db.QueryRow(verifyQuery).Scan(&currentNodeID, &waitingForReply, &execStatus, &flowID)
	if err != nil {
		fmt.Printf("Error verifying flow state: %v\n", err)
		return
	}

	fmt.Printf("Current Node ID: %s\n", getStringValue(currentNodeID))
	fmt.Printf("Waiting for Reply: %v\n", getBoolValue(waitingForReply))
	fmt.Printf("Execution Status: %s\n", getStringValue(execStatus))
	fmt.Printf("Flow ID: %s\n", getStringValue(flowID))

	fmt.Println("\n✅ Ready to test AI prompt node execution!")
	fmt.Println("Send a message to trigger the flow and it should process the AI prompt node.")
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "<NULL>"
}

func getBoolValue(nb sql.NullBool) bool {
	if nb.Valid {
		return nb.Bool
	}
	return false
}