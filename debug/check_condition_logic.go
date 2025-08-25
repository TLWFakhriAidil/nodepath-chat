package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// FlowNode represents a node in the chatbot flow
type FlowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Position map[string]float64    `json:"position"`
	Data     map[string]interface{} `json:"data"`
}

// FlowEdge represents an edge connecting nodes
type FlowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type,omitempty"`
}

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

	fmt.Println("=== Condition Logic Analysis for flow_ai_1756016272 ===")
	fmt.Println("Analyzing condition nodes and their evaluation logic...")
	fmt.Println()

	// Get the flow data
	query := `SELECT nodes, edges FROM chatbot_flows_nodepath WHERE id = ?`
	var nodesJSON, edgesJSON []byte
	err = db.QueryRow(query, "flow_ai_1756016272").Scan(&nodesJSON, &edgesJSON)
	if err != nil {
		log.Fatalf("Failed to get flow data: %v", err)
	}

	// Parse nodes
	var nodes []FlowNode
	err = json.Unmarshal(nodesJSON, &nodes)
	if err != nil {
		log.Fatalf("Failed to parse nodes: %v", err)
	}

	// Parse edges
	var edges []FlowEdge
	err = json.Unmarshal(edgesJSON, &edges)
	if err != nil {
		log.Fatalf("Failed to parse edges: %v", err)
	}

	fmt.Printf("Total nodes: %d\n", len(nodes))
	fmt.Printf("Total edges: %d\n", len(edges))
	fmt.Println()

	// Find condition nodes
	fmt.Println("=== CONDITION NODES ===")
	conditionNodes := make([]FlowNode, 0)
	for _, node := range nodes {
		if node.Type == "condition" {
			conditionNodes = append(conditionNodes, node)
			fmt.Printf("Condition Node ID: %s\n", node.ID)
			fmt.Printf("Node Data: %+v\n", node.Data)
			fmt.Println("---")
		}
	}

	if len(conditionNodes) == 0 {
		fmt.Println("No condition nodes found in this flow.")
		return
	}

	fmt.Printf("Found %d condition node(s)\n\n", len(conditionNodes))

	// Analyze edges from condition nodes
	fmt.Println("=== CONDITION NODE EDGES ===")
	for _, condNode := range conditionNodes {
		fmt.Printf("Condition Node: %s\n", condNode.ID)
		
		// Find all edges from this condition node
		outgoingEdges := make([]FlowEdge, 0)
		for _, edge := range edges {
			if edge.Source == condNode.ID {
				outgoingEdges = append(outgoingEdges, edge)
			}
		}
		
		fmt.Printf("Outgoing edges: %d\n", len(outgoingEdges))
		for i, edge := range outgoingEdges {
			fmt.Printf("  Edge %d: %s -> %s (Type: %s)\n", i+1, edge.Source, edge.Target, edge.Type)
			
			// Find target node details
			for _, targetNode := range nodes {
				if targetNode.ID == edge.Target {
					fmt.Printf("    Target Node Type: %s\n", targetNode.Type)
					if targetNode.Type == "message" {
						if message, ok := targetNode.Data["message"].(string); ok {
							fmt.Printf("    Target Message: %s\n", message)
						}
					}
					break
				}
			}
		}
		fmt.Println("---")
	}

	// Find user_reply nodes that might lead to conditions
	fmt.Println("\n=== USER REPLY NODES ===")
	for _, node := range nodes {
		if node.Type == "user_reply" {
			fmt.Printf("User Reply Node ID: %s\n", node.ID)
			
			// Find what comes after this user_reply node
			for _, edge := range edges {
				if edge.Source == node.ID {
					fmt.Printf("  Next Node: %s\n", edge.Target)
					
					// Find the next node details
					for _, nextNode := range nodes {
						if nextNode.ID == edge.Target {
							fmt.Printf("  Next Node Type: %s\n", nextNode.Type)
							if nextNode.Type == "condition" {
								fmt.Printf("  Condition Data: %+v\n", nextNode.Data)
							}
							break
						}
					}
					break
				}
			}
			fmt.Println("---")
		}
	}

	fmt.Println("\n=== ANALYSIS COMPLETE ===")
	fmt.Println("This shows the condition node structure and how user replies should be evaluated.")
}