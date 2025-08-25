package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database URL from environment or use default
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway"
	}

	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Database State Check ===")
	fmt.Println("Checking flow tracking fields for test device...")
	fmt.Println()

	// Query flow tracking fields for our test device
	query := `
		SELECT prospect_num, id_device, current_node_id, waiting_for_reply, 
		       flow_id, execution_status, execution_id, last_node_id
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = '601137508067' AND id_device = 'FakhriAidilTLW-001'
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to query database:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var prospectNum, idDevice, currentNodeID, flowID, executionStatus, executionID, lastNodeID sql.NullString
		var waitingForReply sql.NullInt32

		err := rows.Scan(&prospectNum, &idDevice, &currentNodeID, &waitingForReply, 
			&flowID, &executionStatus, &executionID, &lastNodeID)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		fmt.Printf("ProspectNum: %s\n", getStringValue(prospectNum))
		fmt.Printf("IDDevice: %s\n", getStringValue(idDevice))
		fmt.Printf("CurrentNodeID: %s\n", getStringValue(currentNodeID))
		fmt.Printf("WaitingForReply: %d\n", getIntValue(waitingForReply))
		fmt.Printf("FlowID: %s\n", getStringValue(flowID))
		fmt.Printf("ExecutionStatus: %s\n", getStringValue(executionStatus))
		fmt.Printf("ExecutionID: %s\n", getStringValue(executionID))
		fmt.Printf("LastNodeID: %s\n", getStringValue(lastNodeID))
		fmt.Println()
	}

	if !found {
		fmt.Println("No records found for the test device.")
	}

	err = rows.Err()
	if err != nil {
		log.Fatal("Error iterating rows:", err)
	}
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}

func getIntValue(ni sql.NullInt32) int32 {
	if ni.Valid {
		return ni.Int32
	}
	return 0
}