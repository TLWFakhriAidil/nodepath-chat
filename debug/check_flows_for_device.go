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

func main() {
	// Get database connection
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
	}

	// Convert to proper DSN format
	dsn := "admin_aqil:admin_aqil@tcp(159.89.198.71:3306)/admin_railway"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	fmt.Println("=== Checking flows for device FakhriAidilTLW-001 ===")

	// Check if flows exist for this device
	query := `SELECT id, name, niche, id_device, nodes, edges, created_at, updated_at 
			  FROM chatbot_flows_nodepath 
			  WHERE id_device = ? 
			  ORDER BY created_at DESC`

	rows, err := db.Query(query, "FakhriAidilTLW-001")
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	found := false
	for rows.Next() {
		found = true
		var id, name, niche, idDevice string
		var nodes, edges []byte
		var createdAt, updatedAt sql.NullTime

		err := rows.Scan(&id, &name, &niche, &idDevice, &nodes, &edges, &createdAt, &updatedAt)
		if err != nil {
			log.Fatal("Scan failed:", err)
		}

		fmt.Printf("\n=== Flow Found ===")
		fmt.Printf("\nFlow ID: %s", id)
		fmt.Printf("\nFlow Name: %s", name)
		fmt.Printf("\nNiche: %s", niche)
		fmt.Printf("\nDevice ID: %s", idDevice)
		fmt.Printf("\nCreated: %v", createdAt)
		fmt.Printf("\nUpdated: %v", updatedAt)
		fmt.Printf("\nNodes Length: %d bytes", len(nodes))
		fmt.Printf("\nEdges Length: %d bytes", len(edges))

		// Parse and display nodes
		if len(nodes) > 0 {
			fmt.Printf("\n\n=== Flow Nodes ===")
			var nodeList []map[string]interface{}
			err := json.Unmarshal(nodes, &nodeList)
			if err != nil {
				fmt.Printf("\nError parsing nodes: %v", err)
				fmt.Printf("\nRaw nodes: %s", string(nodes)[:min(500, len(nodes))])
			} else {
				fmt.Printf("\nTotal nodes: %d", len(nodeList))
				for i, node := range nodeList {
					if i < 10 { // Show first 10 nodes
						nodeID, _ := node["id"].(string)
						nodeType, _ := node["type"].(string)
						fmt.Printf("\n  Node %d: ID=%s, Type=%s", i+1, nodeID, nodeType)
						
						// Show node data if available
						if data, ok := node["data"].(map[string]interface{}); ok {
							if label, exists := data["label"]; exists {
								fmt.Printf(", Label=%v", label)
							}
							if message, exists := data["message"]; exists {
								fmt.Printf(", Message=%v", message)
							}
						}
					}
				}
				if len(nodeList) > 10 {
					fmt.Printf("\n  ... and %d more nodes", len(nodeList)-10)
				}
			}
		}

		// Parse and display edges
		if len(edges) > 0 {
			fmt.Printf("\n\n=== Flow Edges ===")
			var edgeList []map[string]interface{}
			err := json.Unmarshal(edges, &edgeList)
			if err != nil {
				fmt.Printf("\nError parsing edges: %v", err)
			} else {
				fmt.Printf("\nTotal edges: %d", len(edgeList))
				for i, edge := range edgeList {
					if i < 10 { // Show first 10 edges
						edgeID, _ := edge["id"].(string)
						source, _ := edge["source"].(string)
						target, _ := edge["target"].(string)
						fmt.Printf("\n  Edge %d: ID=%s, %s -> %s", i+1, edgeID, source, target)
					}
				}
				if len(edgeList) > 10 {
					fmt.Printf("\n  ... and %d more edges", len(edgeList)-10)
				}
			}
		}
		fmt.Println("\n" + strings.Repeat("=", 50))
	}

	if !found {
		fmt.Println("❌ No flows found for device FakhriAidilTLW-001")
		fmt.Println("This explains why the system falls back to AI conversation.")
		fmt.Println("\nTo fix this issue, you need to:")
		fmt.Println("1. Create a flow for device FakhriAidilTLW-001 in the chatbot_flows_nodepath table")
		fmt.Println("2. Or update an existing flow to use this device ID")
	} else {
		fmt.Println("✅ Flows found for device FakhriAidilTLW-001")
		fmt.Println("The flow engine should be processing these flows.")
		fmt.Println("\nIf AI conversation is still happening instead of flow execution,")
		fmt.Println("check the flow execution logs for errors.")
	}

	// Also check current flow execution status
	fmt.Println("\n=== Checking current flow execution status ===")
	execQuery := `SELECT flow_reference, current_node, execution_status, variables 
				  FROM ai_whatsapp_nodepath 
				  WHERE prospect_num = '60179645043' AND id_device = 'FakhriAidilTLW-001'`

	var flowRef, currentNode, execStatus sql.NullString
	var variables []byte
	err = db.QueryRow(execQuery).Scan(&flowRef, &currentNode, &execStatus, &variables)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("❌ No execution record found for phone 60179645043")
		} else {
			fmt.Printf("❌ Error checking execution: %v\n", err)
		}
	} else {
		fmt.Printf("✅ Execution record found:\n")
		fmt.Printf("  Flow Reference: %s\n", getStringValue(flowRef))
		fmt.Printf("  Current Node: %s\n", getStringValue(currentNode))
		fmt.Printf("  Execution Status: %s\n", getStringValue(execStatus))
		fmt.Printf("  Variables: %s\n", string(variables))
	}
}

func getStringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "<NULL>"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}