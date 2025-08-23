package main

import (
	"database/sql"
	"encoding/json"
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
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway"
	} else {
		// Convert mysql:// format to go-sql-driver format
		if strings.HasPrefix(mysqlURI, "mysql://") {
			mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
			// Replace @ with @tcp( and add ) before /
			parts := strings.Split(mysqlURI, "@")
			if len(parts) == 2 {
				userPass := parts[0]
				hostDbParts := strings.Split(parts[1], "/")
				if len(hostDbParts) == 2 {
					host := hostDbParts[0]
					db := hostDbParts[1]
					mysqlURI = fmt.Sprintf("%s@tcp(%s)/%s", userPass, host, db)
				}
			}
		}
	}

	log.Printf("Connecting to database with DSN: %s", mysqlURI)

	// Connect to database
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

	log.Println("=== Checking ai_whatsapp_nodepath for flow execution ===")

	// Check for flow execution record
	query := `SELECT id_prospect, id_device, prospect_num, stage, human, 
					flow_reference, current_node, variables, execution_status, execution_id,
					conv_last, conv_current, conv_stage, created_at, updated_at
			  FROM ai_whatsapp_nodepath 
			  WHERE prospect_num = '60179645043' AND id_device = 'FakhriAidilTLW-001'
			  ORDER BY updated_at DESC LIMIT 5`

	rows, err := db.Query(query)
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

		fmt.Printf("\n=== Flow Execution Record ===\n")
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

		// Parse and display variables if present
		if variables.Valid && variables.String != "" {
			fmt.Printf("\nVariables:\n")
			var varsMap map[string]interface{}
			err := json.Unmarshal([]byte(variables.String), &varsMap)
			if err != nil {
				fmt.Printf("Error parsing variables: %v\n", err)
				fmt.Printf("Raw variables: %s\n", variables.String)
			} else {
				for key, value := range varsMap {
					fmt.Printf("  %s: %v\n", key, value)
				}
			}
		}
		fmt.Println("\n" + strings.Repeat("=", 50))
	}

	if !found {
		fmt.Println("❌ No flow execution records found for phone 60179645043 and device FakhriAidilTLW-001")
	} else {
		fmt.Println("✅ Flow execution records found")
	}

	// Check if there are any flows configured for this device
	log.Println("\n=== Checking chatbot_flows_nodepath for device ===")
	flowQuery := `SELECT id_device, niche, created_at, updated_at 
				  FROM chatbot_flows_nodepath 
				  WHERE id_device = 'FakhriAidilTLW-001' 
				  ORDER BY created_at DESC`

	flowRows, err := db.Query(flowQuery)
	if err != nil {
		log.Printf("Failed to query chatbot_flows_nodepath: %v", err)
	} else {
		defer flowRows.Close()
		flowFound := false
		for flowRows.Next() {
			flowFound = true
			var idDevice, niche string
			var createdAt, updatedAt sql.NullString

			err := flowRows.Scan(&idDevice, &niche, &createdAt, &updatedAt)
			if err != nil {
				log.Printf("Error scanning flow row: %v", err)
				continue
			}

			fmt.Printf("\n=== Flow Configuration ===\n")
			fmt.Printf("ID Device: %s\n", idDevice)
			fmt.Printf("Niche: %s\n", niche)
			fmt.Printf("Created At: %s\n", getStringValue(createdAt))
			fmt.Printf("Updated At: %s\n", getStringValue(updatedAt))
		}

		if !flowFound {
			fmt.Println("❌ No flows configured for device FakhriAidilTLW-001")
		} else {
			fmt.Println("✅ Flows found for device")
		}
	}
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "<NULL>"
}