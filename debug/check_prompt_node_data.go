package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// FlowNode represents a node in the chatbot flow
type FlowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Position map[string]float64    `json:"position"`
	Data     map[string]interface{} `json:"data"`
}

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

	fmt.Println("=== AI Prompt Node Data Analysis ===")

	// Get flow data
	flowQuery := `SELECT nodes FROM chatbot_flows_nodepath WHERE id = ?`
	var nodesData []byte
	err = db.QueryRow(flowQuery, "flow_ai_1756016272").Scan(&nodesData)
	if err != nil {
		fmt.Printf("Error querying flow: %v\n", err)
		return
	}

	// Parse nodes
	var nodes []FlowNode
	err = json.Unmarshal(nodesData, &nodes)
	if err != nil {
		fmt.Printf("Error parsing nodes: %v\n", err)
		return
	}

	// Find the AI prompt node
	promptNodeID := "prompt-1754883603395"
	var promptNode *FlowNode
	for _, node := range nodes {
		if node.ID == promptNodeID {
			promptNode = &node
			break
		}
	}

	if promptNode == nil {
		fmt.Printf("AI Prompt node %s not found!\n", promptNodeID)
		return
	}

	fmt.Printf("\n=== AI Prompt Node: %s ===\n", promptNodeID)
	fmt.Printf("Type: %s\n", promptNode.Type)
	fmt.Printf("\nNode Data:\n")

	if promptNode.Data == nil {
		fmt.Println("  No data found!")
		return
	}

	// Check all data fields
	for key, value := range promptNode.Data {
		fmt.Printf("  %s: ", key)
		switch v := value.(type) {
		case string:
			if len(v) > 200 {
				fmt.Printf("%s... (truncated, total: %d chars)\n", v[:200], len(v))
			} else {
				fmt.Printf("%s\n", v)
			}
		case nil:
			fmt.Println("<nil>")
		default:
			fmt.Printf("%v (type: %T)\n", v, v)
		}
	}

	// Check specific required fields
	fmt.Printf("\n=== Required Fields Check ===\n")
	requiredFields := []string{"system_prompt", "instance", "apiprovider"}
	for _, field := range requiredFields {
		if value, exists := promptNode.Data[field]; exists {
			if str, ok := value.(string); ok && str != "" {
				fmt.Printf("✅ %s: Present (%d chars)\n", field, len(str))
			} else {
				fmt.Printf("❌ %s: Empty or not string\n", field)
			}
		} else {
			fmt.Printf("❌ %s: Missing\n", field)
		}
	}

	// Check current flow state
	fmt.Printf("\n=== Current Flow State ===\n")
	stateQuery := `SELECT current_node_id, flow_id, execution_status, waiting_for_reply FROM ai_whatsapp_nodepath WHERE prospect_num = ? AND id_device = ?`
	var currentNodeID, flowID, execStatus sql.NullString
	var waitingForReply sql.NullBool
	err = db.QueryRow(stateQuery, "60179645043", "FakhriAidilTLW-001").Scan(&currentNodeID, &flowID, &execStatus, &waitingForReply)
	if err != nil {
		fmt.Printf("Error querying flow state: %v\n", err)
	} else {
		fmt.Printf("Current Node ID: %s\n", getStringValue(currentNodeID))
		fmt.Printf("Flow ID: %s\n", getStringValue(flowID))
		fmt.Printf("Execution Status: %s\n", getStringValue(execStatus))
		fmt.Printf("Waiting for Reply: %v\n", getBoolValue(waitingForReply))
	}
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