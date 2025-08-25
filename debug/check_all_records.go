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

	fmt.Println("=== All AI WhatsApp Records Check ===")
	fmt.Println("Checking all records for device FakhriAidilTLW-001...")
	fmt.Println()

	// Query all records for the device
	query := `
		SELECT id_prospect, prospect_num, id_device, current_node_id, waiting_for_reply, 
		       flow_id, execution_status, execution_id, last_node_id, created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE id_device = ? 
		ORDER BY updated_at DESC
	`

	rows, err := db.Query(query, "FakhriAidilTLW-001")
	if err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}
	defer rows.Close()

	recordCount := 0
	for rows.Next() {
		recordCount++
		var idProspect int
		var prospectNum, idDevice string
		var currentNodeID, flowID, executionStatus, executionID, lastNodeID sql.NullString
		var waitingForReply sql.NullInt32
		var createdAt, updatedAt time.Time

		err := rows.Scan(&idProspect, &prospectNum, &idDevice, &currentNodeID, &waitingForReply,
			&flowID, &executionStatus, &executionID, &lastNodeID, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		fmt.Printf("Record %d:\n", recordCount)
		fmt.Printf("ID Prospect: %d\n", idProspect)
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
		
		if executionID.Valid {
			fmt.Printf("ExecutionID: %s\n", executionID.String)
		} else {
			fmt.Printf("ExecutionID: NULL\n")
		}
		
		if lastNodeID.Valid {
			fmt.Printf("LastNodeID: %s\n", lastNodeID.String)
		} else {
			fmt.Printf("LastNodeID: NULL\n")
		}
		
		fmt.Printf("CreatedAt: %s\n", createdAt.Format("2006-01-02T15:04:05Z07:00"))
		fmt.Printf("UpdatedAt: %s\n", updatedAt.Format("2006-01-02T15:04:05Z07:00"))
		fmt.Println("---")
	}

	if recordCount == 0 {
		fmt.Println("No records found for device FakhriAidilTLW-001.")
	} else {
		fmt.Printf("\nTotal records found: %d\n", recordCount)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}
}