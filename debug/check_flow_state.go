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
		// Convert mysql:// format to Go driver format
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
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("Connected to database successfully")

	// Query all records for this device to see what's in the database
	query := `
		SELECT prospect_num, id_device, current_node_id, waiting_for_reply, 
		       flow_id, execution_status, conv_last
		FROM ai_whatsapp_nodepath 
		WHERE id_device = 'FakhriAidilTLW-001'
		ORDER BY updated_at DESC
		LIMIT 5
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal("Failed to execute query:", err)
	}
	defer rows.Close()

	fmt.Println("\n=== Flow State for Test Conversation ===")
	fmt.Printf("%-15s %-20s %-15s %-15s %-20s %-15s\n", 
		"ProspectNum", "IDDevice", "CurrentNodeID", "WaitingReply", "FlowID", "ExecStatus")
	fmt.Println(strings.Repeat("-", 100))

	for rows.Next() {
		var prospectNum, idDevice, currentNodeID, flowID, execStatus, convLast sql.NullString
		var waitingReply sql.NullInt32

		err := rows.Scan(&prospectNum, &idDevice, &currentNodeID, &waitingReply, 
			&flowID, &execStatus, &convLast)
		if err != nil {
			log.Fatal("Failed to scan row:", err)
		}

		// Display values
		fmt.Printf("%-15s %-20s %-15s %-15d %-20s %-15s\n",
			getStringValue(prospectNum),
			getStringValue(idDevice),
			getStringValue(currentNodeID),
			getIntValue(waitingReply),
			getStringValue(flowID),
			getStringValue(execStatus))

		// Show conversation history
		fmt.Println("\nConversation History:")
		if convLast.Valid {
			fmt.Println(convLast.String)
		} else {
			fmt.Println("No conversation history")
		}
	}

	if err = rows.Err(); err != nil {
		log.Fatal("Row iteration error:", err)
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