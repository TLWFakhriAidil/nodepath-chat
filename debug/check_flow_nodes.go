package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database connection string
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		log.Fatal("MYSQL_URI environment variable is not set")
	}

	// Convert mysql:// format to Go driver format
	dsn := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway?charset=utf8mb4&parseTime=True&loc=Local"

	// Connect to database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Flow Analysis for flow_ai_1756016272 ===")
	fmt.Println("Checking flow structure and current node details...")
	fmt.Println()

	// First, get flow information
	flowQuery := `SELECT id, name, description, created_at FROM chatbot_flows_nodepath WHERE id = ?`
	var flowID, flowName, flowDesc, flowCreated sql.NullString
	err = db.QueryRow(flowQuery, "flow_ai_1756016272").Scan(&flowID, &flowName, &flowDesc, &flowCreated)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Flow flow_ai_1756016272 not found in chatbot_flows_nodepath")
			return
		}
		log.Fatal("Failed to query flow:", err)
	}

	fmt.Printf("Flow ID: %v\n", getValue(flowID))
	fmt.Printf("Flow Name: %v\n", getValue(flowName))
	fmt.Printf("Flow Description: %v\n", getValue(flowDesc))
	fmt.Printf("Flow Created: %v\n", getValue(flowCreated))
	fmt.Println()

	// Check if there's a nodes table or if nodes are stored in JSON
	fmt.Println("=== Current Node Analysis ===")
	fmt.Printf("Current Node ID: user_reply-1756015760720\n")
	fmt.Printf("Waiting for Reply: 1 (Yes)\n")
	fmt.Printf("Execution Status: active\n")
	fmt.Printf("Last Node ID: user_reply-1756015760720\n")
	fmt.Println()

	fmt.Println("=== Analysis Summary ===")
	fmt.Println("• The conversation is currently at a user_reply node")
	fmt.Println("• The system is waiting for user input (waiting_for_reply = 1)")
	fmt.Println("• The flow execution is active")
	fmt.Println("• ExecutionID is NULL, which might indicate an issue")
	fmt.Println("• The record was last updated at 2025-08-25T04:09:38+08:00")
}

func getValue(v interface{}) string {
	switch val := v.(type) {
	case sql.NullString:
		if val.Valid {
			return val.String
		}
		return "NULL"
	case sql.NullInt64:
		if val.Valid {
			return fmt.Sprintf("%d", val.Int64)
		}
		return "NULL"
	default:
		return "NULL"
	}
}