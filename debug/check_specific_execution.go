package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

// Helper function to get string value from sql.NullString
func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "<NULL>"
}

func main() {
	// Database connection
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway"
	}

	log.Printf("Connecting to database with DSN: %s", mysqlURI)
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Check for the specific phone number from Railway logs
	phoneNumber := "120363306218165529"
	deviceID := "FakhriAidilTLW-001"

	log.Printf("=== Checking for execution record: Phone %s, Device %s ===", phoneNumber, deviceID)

	// Query ai_whatsapp_nodepath for the specific phone number
	query := `SELECT id_prospect, id_device, prospect_num, stage, human,
					flow_reference, current_node, variables, execution_status, execution_id,
					conv_last, conv_current, conv_stage, created_at, updated_at
			  FROM ai_whatsapp_nodepath 
			  WHERE prospect_num = ? AND id_device = ?
			  ORDER BY updated_at DESC LIMIT 5`

	rows, err := db.Query(query, phoneNumber, deviceID)
	if err != nil {
		log.Fatalf("Failed to query ai_whatsapp_nodepath: %v", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var idProspect, idDevice, prospectNum, stage string
		var human int
		var flowReference, currentNode, executionStatus, executionID sql.NullString
		var variables sql.NullString
		var convLast, convCurrent, convStage sql.NullString
		var createdAt, updatedAt sql.NullString

		err := rows.Scan(&idProspect, &idDevice, &prospectNum, &stage, &human,
			&flowReference, &currentNode, &variables, &executionStatus, &executionID,
			&convLast, &convCurrent, &convStage, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		fmt.Printf("\n=== Flow Execution Record Found ===\n")
		fmt.Printf("ID Prospect: %s\n", idProspect)
		fmt.Printf("ID Device: %s\n", idDevice)
		fmt.Printf("Prospect Num: %s\n", prospectNum)
		fmt.Printf("Stage: %s\n", stage)
		fmt.Printf("Human: %d\n", human)
		fmt.Printf("Flow Reference: %s\n", getStringValue(flowReference))
		fmt.Printf("Current Node: %s\n", getStringValue(currentNode))
		fmt.Printf("Execution Status: %s\n", getStringValue(executionStatus))
		fmt.Printf("Execution ID: %s\n", getStringValue(executionID))
		fmt.Printf("Conv Last: %s\n", getStringValue(convLast))
		fmt.Printf("Conv Current: %s\n", getStringValue(convCurrent))
		fmt.Printf("Conv Stage: %s\n", getStringValue(convStage))
		fmt.Printf("Created At: %s\n", getStringValue(createdAt))
		fmt.Printf("Updated At: %s\n", getStringValue(updatedAt))
	}

	if !found {
		fmt.Printf("\n❌ NO EXECUTION RECORD FOUND for phone %s and device %s\n", phoneNumber, deviceID)
		fmt.Printf("This explains why delayed messages are failing with 'execution not found'\n")
		
		// Check if there are any records for this device with different phone numbers
		fmt.Printf("\n=== Checking all records for device %s ===\n", deviceID)
		allQuery := `SELECT prospect_num, stage, execution_status, created_at, updated_at
					 FROM ai_whatsapp_nodepath 
					 WHERE id_device = ?
					 ORDER BY updated_at DESC LIMIT 10`
		
		allRows, err := db.Query(allQuery, deviceID)
		if err != nil {
			log.Printf("Failed to query all records: %v", err)
			return
		}
		defer allRows.Close()
		
		for allRows.Next() {
			var prospectNum, stage string
			var executionStatus sql.NullString
			var createdAt, updatedAt sql.NullString
			
			err := allRows.Scan(&prospectNum, &stage, &executionStatus, &createdAt, &updatedAt)
			if err != nil {
				log.Printf("Error scanning all records row: %v", err)
				continue
			}
			
			fmt.Printf("Phone: %s, Stage: %s, Status: %s, Updated: %s\n", 
				prospectNum, stage, getStringValue(executionStatus), getStringValue(updatedAt))
		}
	} else {
		fmt.Printf("\n✅ Execution record found\n")
	}
}