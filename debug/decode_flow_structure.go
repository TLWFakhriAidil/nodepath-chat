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

// FlowEdge represents an edge in the chatbot flow
type FlowEdge struct {
	ID           string `json:"id"`
	Source       string `json:"source"`
	Target       string `json:"target"`
	SourceHandle string `json:"sourceHandle"`
	TargetHandle string `json:"targetHandle"`
}

func main() {
	// Get database connection string from environment
	mysqlURI := os.Getenv("MYSQL_URI")
	if mysqlURI == "" {
		mysqlURI = "mysql://admin_aqil:admin_aqil@159.89.198.71:3306/admin_railway"
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

	fmt.Println("=== Flow Structure Analysis for flow_ai_1756016272 ===")

	// Get flow data
	flowQuery := `SELECT nodes, edges FROM chatbot_flows_nodepath WHERE id = ?`
	var nodesData, edgesData []byte
	err = db.QueryRow(flowQuery, "flow_ai_1756016272").Scan(&nodesData, &edgesData)
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

	// Parse edges
	var edges []FlowEdge
	err = json.Unmarshal(edgesData, &edges)
	if err != nil {
		fmt.Printf("Error parsing edges: %v\n", err)
		return
	}

	fmt.Printf("\n=== NODES (%d total) ===\n", len(nodes))
	for i, node := range nodes {
		fmt.Printf("%d. Node ID: %s\n", i+1, node.ID)
		fmt.Printf("   Type: %s\n", node.Type)
		if node.Data != nil {
			if title, ok := node.Data["title"].(string); ok {
				fmt.Printf("   Title: %s\n", title)
			}
			if systemPrompt, ok := node.Data["system_prompt"].(string); ok && systemPrompt != "" {
				fmt.Printf("   System Prompt: %s\n", systemPrompt[:min(100, len(systemPrompt))])
			}
			if message, ok := node.Data["message"].(string); ok && message != "" {
				fmt.Printf("   Message: %s\n", message[:min(100, len(message))])
			}
		}
		fmt.Println()
	}

	fmt.Printf("\n=== EDGES (%d total) ===\n", len(edges))
	for i, edge := range edges {
		fmt.Printf("%d. Edge ID: %s\n", i+1, edge.ID)
		fmt.Printf("   From: %s -> To: %s\n", edge.Source, edge.Target)
		fmt.Printf("   Handles: %s -> %s\n", edge.SourceHandle, edge.TargetHandle)
		fmt.Println()
	}

	// Analyze flow path from user_reply-1756105605330
	fmt.Println("\n=== FLOW PATH ANALYSIS ===")
	currentNode := "user_reply-1756105605330"
	fmt.Printf("Starting from current node: %s\n", currentNode)

	// Find next nodes
	for step := 1; step <= 10; step++ {
		nextNode := findNextNode(edges, currentNode)
		if nextNode == "" {
			fmt.Printf("Step %d: No next node found (end of flow)\n", step)
			break
		}

		nodeInfo := findNodeInfo(nodes, nextNode)
		fmt.Printf("Step %d: %s -> %s (Type: %s)\n", step, currentNode, nextNode, nodeInfo.Type)
		
		// Show node details
		if nodeInfo.Data != nil {
			if title, ok := nodeInfo.Data["title"].(string); ok {
				fmt.Printf("        Title: %s\n", title)
			}
			if systemPrompt, ok := nodeInfo.Data["system_prompt"].(string); ok && systemPrompt != "" {
				fmt.Printf("        Has System Prompt: Yes (%d chars)\n", len(systemPrompt))
			}
			if message, ok := nodeInfo.Data["message"].(string); ok && message != "" {
				fmt.Printf("        Message: %s\n", message[:min(50, len(message))])
			}
		}

		currentNode = nextNode
	}
}

func findNextNode(edges []FlowEdge, currentNode string) string {
	for _, edge := range edges {
		if edge.Source == currentNode {
			return edge.Target
		}
	}
	return ""
}

func findNodeInfo(nodes []FlowNode, nodeID string) FlowNode {
	for _, node := range nodes {
		if node.ID == nodeID {
			return node
		}
	}
	return FlowNode{}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}