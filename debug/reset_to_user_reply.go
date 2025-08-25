package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

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

	fmt.Println("=== Reset Flow to User Reply Node ===")
	fmt.Println("Resetting flow for phone 601137508067 back to user_reply node...")
	fmt.Println()

	// Update the record to reset it back to user_reply node and set waiting_for_reply = 1
	updateQuery := `
		UPDATE ai_whatsapp_nodepath 
		SET current_node_id = 'user_reply-1756015760720',
		    waiting_for_reply = 1,
		    last_node_id = 'message-1755584833903',
		    updated_at = ?
		WHERE prospect_num = '601137508067' AND id_device = 'FakhriAidilTLW-001'
	`

	result, err := db.Exec(updateQuery, time.Now())
	if err != nil {
		log.Fatalf("Failed to update record: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Could not get rows affected: %v", err)
	} else {
		fmt.Printf("Rows affected: %d\n", rowsAffected)
	}

	if rowsAffected == 0 {
		fmt.Println("No records were updated. Record might not exist.")
		return
	}

	fmt.Println("✅ Flow reset successfully!")
	fmt.Println()

	// Verify the update by querying the record
	fmt.Println("Verifying the update...")
	verifyQuery := `
		SELECT prospect_num, id_device, current_node_id, waiting_for_reply, 
		       flow_id, execution_status, last_node_id, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = '601137508067' AND id_device = 'FakhriAidilTLW-001'
	`

	var prospectNum, idDevice string
	var currentNodeID, flowID, executionStatus, lastNodeID sql.NullString
	var waitingForReply sql.NullInt32
	var updatedAt time.Time

	err = db.QueryRow(verifyQuery).Scan(&prospectNum, &idDevice, &currentNodeID, &waitingForReply,
		&flowID, &executionStatus, &lastNodeID, &updatedAt)
	if err != nil {
		log.Fatalf("Failed to verify update: %v", err)
	}

	fmt.Printf("ProspectNum: %s\n", prospectNum)
	fmt.Printf("IDDevice: %s\n", idDevice)
	
	if currentNodeID.Valid {
		fmt.Printf("CurrentNodeID: %s\n", currentNodeID.String)
	} else {
		fmt.Printf("CurrentNodeID: NULL\n")
	}
	
	if waitingForReply.Valid {
		fmt.Printf("WaitingForReply: %d\n", waitingForReply.Int32)
	} else {
		fmt.Printf("WaitingForReply: NULL\n")
	}
	
	if flowID.Valid {
		fmt.Printf("FlowID: %s\n", flowID.String)
	} else {
		fmt.Printf("FlowID: NULL\n")
	}
	
	if executionStatus.Valid {
		fmt.Printf("ExecutionStatus: %s\n", executionStatus.String)
	} else {
		fmt.Printf("ExecutionStatus: NULL\n")
	}
	
	if lastNodeID.Valid {
		fmt.Printf("LastNodeID: %s\n", lastNodeID.String)
	} else {
		fmt.Printf("LastNodeID: NULL\n")
	}
	
	fmt.Printf("UpdatedAt: %s\n", updatedAt.Format("2006-01-02T15:04:05Z07:00"))
	fmt.Println()
	fmt.Println("✅ Flow is now ready to test user reply functionality!")
	fmt.Println("You can now send a webhook with 'anak' or 'sendiri' to test flow continuation.")
}