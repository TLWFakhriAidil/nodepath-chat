package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

// FlowEdge represents an edge connecting nodes
type FlowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type,omitempty"`
}

// EvaluateConditionNode simulates the condition evaluation logic
func EvaluateConditionNode(nodes []FlowNode, edges []FlowEdge, conditionNodeID string, userInput string) (*FlowNode, error) {
	// Find the condition node
	var conditionNode *FlowNode
	for _, node := range nodes {
		if node.ID == conditionNodeID {
			conditionNode = &node
			break
		}
	}

	if conditionNode == nil {
		return nil, fmt.Errorf("condition node not found: %s", conditionNodeID)
	}

	// Get conditions from node data
	conditions, ok := conditionNode.Data["conditions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no conditions found in condition node %s", conditionNodeID)
	}

	// Find outgoing edges from this condition node
	var outgoingEdges []FlowEdge
	for _, edge := range edges {
		if edge.Source == conditionNodeID {
			outgoingEdges = append(outgoingEdges, edge)
		}
	}

	if len(outgoingEdges) == 0 {
		return nil, fmt.Errorf("no outgoing edges found for condition node %s", conditionNodeID)
	}

	// Normalize user input for comparison
	userInputLower := strings.ToLower(strings.TrimSpace(userInput))

	fmt.Printf("Evaluating user input: '%s' (normalized: '%s')\n", userInput, userInputLower)
	fmt.Printf("Found %d conditions and %d outgoing edges\n", len(conditions), len(outgoingEdges))

	// Evaluate each condition
	for i, conditionInterface := range conditions {
		condition, ok := conditionInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Get condition properties
		conditionType, _ := condition["type"].(string)
		conditionValue, _ := condition["value"].(string)
		conditionLabel, _ := condition["label"].(string)

		// Normalize condition value for comparison
		conditionValueLower := strings.ToLower(strings.TrimSpace(conditionValue))

		fmt.Printf("Condition %d: label='%s', type='%s', value='%s' (normalized: '%s')\n", 
			i+1, conditionLabel, conditionType, conditionValue, conditionValueLower)

		// Evaluate condition based on type
		var matches bool
		switch conditionType {
		case "equals":
			matches = userInputLower == conditionValueLower
		case "contains":
			matches = strings.Contains(userInputLower, conditionValueLower)
		case "default":
			// Default condition matches if no other conditions match
			continue
		default:
			// Fallback: treat as equals
			matches = userInputLower == conditionValueLower
		}

		fmt.Printf("  Match result: %v\n", matches)

		// If condition matches, find the corresponding edge
		if matches && i < len(outgoingEdges) {
			targetNodeID := outgoingEdges[i].Target
			fmt.Printf("  ✅ MATCH! Using edge %d to target node: %s\n", i, targetNodeID)
			
			// Find and return the target node
			for _, node := range nodes {
				if node.ID == targetNodeID {
					return &node, nil
				}
			}
		}
	}

	// If no conditions match, try to find a default condition
	for i, conditionInterface := range conditions {
		condition, ok := conditionInterface.(map[string]interface{})
		if !ok {
			continue
		}

		conditionType, _ := condition["type"].(string)
		if conditionType == "default" && i < len(outgoingEdges) {
			targetNodeID := outgoingEdges[i].Target
			fmt.Printf("  🔄 Using default condition, edge %d to target node: %s\n", i, targetNodeID)
			
			// Find and return the target node
			for _, node := range nodes {
				if node.ID == targetNodeID {
					return &node, nil
				}
			}
		}
	}

	// If no conditions match and no default, use the first edge as fallback
	if len(outgoingEdges) > 0 {
		targetNodeID := outgoingEdges[0].Target
		fmt.Printf("  ⚠️ No matches, using fallback edge 0 to target node: %s\n", targetNodeID)
		
		// Find and return the target node
		for _, node := range nodes {
			if node.ID == targetNodeID {
				return &node, nil
			}
		}
	}

	return nil, fmt.Errorf("no valid next node found for condition node %s", conditionNodeID)
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

	fmt.Println("=== Testing Condition Evaluation for flow_ai_1756016272 ===")
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

	// Find the condition node
	conditionNodeID := "condition-1755585116033"

	// Test cases
	testInputs := []string{"anak", "sendiri", "Anak", "Sendiri", "ANAK", "SENDIRI", "other"}

	for _, testInput := range testInputs {
		fmt.Printf("\n=== Testing input: '%s' ===\n", testInput)
		nextNode, err := EvaluateConditionNode(nodes, edges, conditionNodeID, testInput)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		if nextNode != nil {
			fmt.Printf("✅ Result: Next node ID = %s, Type = %s\n", nextNode.ID, nextNode.Type)
			if message, ok := nextNode.Data["message"].(string); ok {
				fmt.Printf("📝 Message: %s\n", message)
			}
		} else {
			fmt.Printf("❌ No next node found\n")
		}
	}

	fmt.Println("\n=== Test Complete ===")
}