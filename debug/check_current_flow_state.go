package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get database connection string from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("mysql", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Connected to database successfully!")
	fmt.Println(strings.Repeat("=", 60))

	// Check current flow execution state for FakhriAidilTLW-001 and 60179645043
	checkCurrentFlowState(db)

	// Check available flows for the device
	checkAvailableFlows(db)

	// Check conversation history
	checkConversationHistory(db)
}

func checkCurrentFlowState(db *sql.DB) {
	fmt.Println("\n=== CURRENT FLOW EXECUTION STATE ===")
	
	query := `
		SELECT id_prospect, id_device, prospect_num, stage, human,
			   flow_reference, current_node, variables, execution_status, execution_id,
			   conv_last, conv_current, conv_stage, created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = '60179645043' AND id_device = 'FakhriAidilTLW-001'
		ORDER BY updated_at DESC LIMIT 1
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to query ai_whatsapp_nodepath: %v", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var idProspect, idDevice, prospectNum, stage string
		var human int
		var flowReference, currentNode, executionStatus, executionID sql.NullString
		var variables sql.NullString
		var convLast, convCurrent, convStage sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&idProspect, &idDevice, &prospectNum, &stage, &human,
			&flowReference, &currentNode, &variables, &executionStatus, &executionID,
			&convLast, &convCurrent, &convStage, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return
		}

		fmt.Printf("ID Prospect: %s\n", idProspect)
		fmt.Printf("ID Device: %s\n", idDevice)
		fmt.Printf("Prospect Num: %s\n", prospectNum)
		fmt.Printf("Stage: %s\n", stage)
		fmt.Printf("Human: %d\n", human)
		fmt.Printf("Flow Reference: %s\n", getStringValue(flowReference))
		fmt.Printf("Current Node: %s\n", getStringValue(currentNode))
		fmt.Printf("Execution Status: %s\n", getStringValue(executionStatus))
		fmt.Printf("Execution ID: %s\n", getStringValue(executionID))
		fmt.Printf("Variables: %s\n", getStringValue(variables))
		fmt.Printf("Conv Stage: %s\n", getStringValue(convStage))
		fmt.Printf("Created At: %s\n", getTimeValue(createdAt))
		fmt.Printf("Updated At: %s\n", getTimeValue(updatedAt))

		// Analysis
		fmt.Println("\n--- ANALYSIS ---")
		if !flowReference.Valid || flowReference.String == "" {
			fmt.Println("❌ NO FLOW REFERENCE: Flow execution not initialized")
		} else {
			fmt.Printf("✅ FLOW REFERENCE: %s\n", flowReference.String)
		}

		if !executionStatus.Valid || executionStatus.String == "" {
			fmt.Println("❌ NO EXECUTION STATUS: Flow execution not active")
		} else {
			fmt.Printf("✅ EXECUTION STATUS: %s\n", executionStatus.String)
		}

		if !currentNode.Valid || currentNode.String == "" {
			fmt.Println("❌ NO CURRENT NODE: Flow position not tracked")
		} else {
			fmt.Printf("✅ CURRENT NODE: %s\n", currentNode.String)
		}
	} else {
		fmt.Println("❌ NO RECORD FOUND for prospect 60179645043 and device FakhriAidilTLW-001")
	}
}

func checkAvailableFlows(db *sql.DB) {
	fmt.Println("\n=== AVAILABLE FLOWS FOR DEVICE ===")
	
	query := `
		SELECT id, id_device, name, nodes, created_at, updated_at
		FROM chatbot_flows_nodepath 
		WHERE id_device = 'FakhriAidilTLW-001'
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to query chatbot_flows_nodepath: %v", err)
		return
	}
	defer rows.Close()

	flowCount := 0
	for rows.Next() {
		flowCount++
		var id, idDevice, flowName string
		var flowData sql.NullString
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &idDevice, &flowName, &flowData, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("Error scanning flow row: %v", err)
			continue
		}

		fmt.Printf("\nFlow #%d:\n", flowCount)
		fmt.Printf("  ID: %s\n", id)
		fmt.Printf("  Device: %s\n", idDevice)
		fmt.Printf("  Name: %s\n", flowName)
		fmt.Printf("  Created: %s\n", getTimeValue(createdAt))
		fmt.Printf("  Updated: %s\n", getTimeValue(updatedAt))
		
		// Check if flow data contains the nodes from the conversation
		if flowData.Valid {
			flowDataStr := flowData.String
			fmt.Printf("  Has Flow Data: Yes (%d chars)\n", len(flowDataStr))
			
			// Check for specific nodes mentioned in the conversation
			if strings.Contains(flowDataStr, "message-1755584833903") {
				fmt.Println("  ✅ Contains expected message node")
			}
			if strings.Contains(flowDataStr, "delay-1755584872621") {
				fmt.Println("  ✅ Contains expected delay node")
			}
			if strings.Contains(flowDataStr, "image-1755584904039") {
				fmt.Println("  ✅ Contains expected image node")
			}
			if strings.Contains(flowDataStr, "condition-1755585116033") {
				fmt.Println("  ✅ Contains expected condition node")
			}
		} else {
			fmt.Println("  ❌ No Flow Data")
		}
	}

	if flowCount == 0 {
		fmt.Println("❌ NO FLOWS FOUND for device FakhriAidilTLW-001")
	} else {
		fmt.Printf("\n✅ FOUND %d FLOW(S) for device FakhriAidilTLW-001\n", flowCount)
	}
}

func checkConversationHistory(db *sql.DB) {
	fmt.Println("\n=== RECENT CONVERSATION HISTORY ===")
	
	query := `
		SELECT conv_last, conv_current
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = '60179645043' AND id_device = 'FakhriAidilTLW-001'
		ORDER BY updated_at DESC LIMIT 1
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Failed to query conversation history: %v", err)
		return
	}
	defer rows.Close()

	if rows.Next() {
		var convLast, convCurrent sql.NullString

		err := rows.Scan(&convLast, &convCurrent)
		if err != nil {
			log.Printf("Error scanning conversation row: %v", err)
			return
		}

		fmt.Printf("Conv Last: %s\n", getStringValue(convLast))
		fmt.Printf("Conv Current: %s\n", getStringValue(convCurrent))
		
		// Check if conversation contains the messages from the log
		if convLast.Valid {
			convStr := convLast.String
			if strings.Contains(convStr, "Hai saya nak") {
				fmt.Println("✅ Contains user message: 'Hai saya nak'")
			}
			if strings.Contains(convStr, "Hai, Assalamualaikum, Saya Fakhri") {
				fmt.Println("✅ Contains bot message: 'Hai, Assalamualaikum, Saya Fakhri'")
			}
			if strings.Contains(convStr, "Ni untuk anak atau sendiri") {
				fmt.Println("✅ Contains bot message: 'Ni untuk anak atau sendiri'")
			}
			if strings.Contains(convStr, "Ohh, anak rupanya...kejpa nyea") {
				fmt.Println("✅ Contains bot message: 'Ohh, anak rupanya...kejpa nyea'")
			}
		}
	} else {
		fmt.Println("❌ NO CONVERSATION HISTORY FOUND")
	}
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "NULL"
}

func getTimeValue(nt sql.NullTime) string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02 15:04:05")
	}
	return "NULL"
}