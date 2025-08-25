package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Use the correct MySQL connection string format for Go driver
	mysqlURI := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("=== Restart Flow for Testing ===")
	fmt.Println("Restarting flow for phone 60179645043 back to user_reply node...")
	fmt.Println()

	// Update the flow state to go back to user_reply node
	query := `UPDATE ai_whatsapp_nodepath 
			  SET current_node_id = ?, 
			      waiting_for_reply = ?, 
			      execution_status = ?,
			      last_node_id = NULL,
			      updated_at = NOW()
			  WHERE prospect_num = ? AND id_device = ?`

	result, err := db.Exec(query, "user_reply-1756015760720", 1, "active", "60179645043", "FakhriAidilTLW-001")
	if err != nil {
		log.Fatalf("Failed to update flow state: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatalf("Failed to get rows affected: %v", err)
	}

	fmt.Printf("Rows affected: %d\n", rowsAffected)

	if rowsAffected > 0 {
		fmt.Println("✅ Flow successfully restarted!")
		fmt.Println("Flow is now at user_reply node and waiting for user input.")
		fmt.Println("You can now test with 'sendiri' input.")
	} else {
		fmt.Println("❌ No records were updated. Record might not exist.")
	}

	// Verify the update
	fmt.Println("\n=== Verification ===")
	verifyQuery := `SELECT current_node_id, waiting_for_reply, execution_status 
					 FROM ai_whatsapp_nodepath 
					 WHERE prospect_num = ? AND id_device = ?`

	var currentNodeID sql.NullString
	var waitingForReply sql.NullInt32
	var executionStatus sql.NullString

	err = db.QueryRow(verifyQuery, "60179645043", "FakhriAidilTLW-001").Scan(&currentNodeID, &waitingForReply, &executionStatus)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}

	fmt.Printf("Current Node ID: %s\n", currentNodeID.String)
	fmt.Printf("Waiting For Reply: %d\n", waitingForReply.Int32)
	fmt.Printf("Execution Status: %s\n", executionStatus.String)
}