package services

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"nodepath-chat/internal/models"
	"strings"
)

// Simple patch for EvaluateConditionNode to handle ALL conditions properly
func (s *FlowService) EvaluateConditionNodePatch(flow *models.ChatbotFlow, conditionNodeID string, userInput string) (*models.FlowNode, error) {
	// Get the condition node
	conditionNode, err := s.FindNodeByID(flow, conditionNodeID)
	if err != nil || conditionNode == nil {
		return nil, fmt.Errorf("condition node not found: %s", conditionNodeID)
	}

	// Get all edges
	edges, err := s.GetFlowEdges(flow)
	if err != nil {
		return nil, err
	}

	// Get conditions from node data
	conditions, ok := conditionNode.Data["conditions"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("no conditions found in condition node %s", conditionNodeID)
	}

	// Collect ALL outgoing edges (not limited by condition count)
	var outgoingEdges []models.FlowEdge
	for _, edge := range edges {
		if edge.Source == conditionNodeID {
			outgoingEdges = append(outgoingEdges, *edge)
		}
	}

	// Log the actual counts
	logrus.WithFields(logrus.Fields{
		"conditions_count": len(conditions),
		"edges_count":      len(outgoingEdges),
		"node_id":          conditionNodeID,
	}).Info("CONDITION FIX: Processing ALL conditions")

	// Normalize user input
	userInputLower := strings.ToLower(strings.TrimSpace(userInput))

	// Process ALL conditions (not limited by edge count)
	for i, conditionInterface := range conditions {
		condition, ok := conditionInterface.(map[string]interface{})
		if !ok {
			continue
		}

		conditionType, _ := condition["type"].(string)
		conditionValue, _ := condition["value"].(string)

		if conditionType == "default" {
			continue // Handle default later
		}

		// Check if condition matches
		conditionValueLower := strings.ToLower(strings.TrimSpace(conditionValue))
		var matches bool

		if conditionType == "contains" || conditionValue != "" {
			// For contains or when value exists, check if input contains the value
			matches = strings.Contains(userInputLower, conditionValueLower)
		}

		if matches {
			// Use modulo to handle cases where conditions > edges
			edgeIndex := i % len(outgoingEdges)
			if edgeIndex < len(outgoingEdges) {
				logrus.WithFields(logrus.Fields{
					"condition_index": i,
					"edge_index":      edgeIndex,
					"matched_value":   conditionValue,
				}).Info("CONDITION FIX: Matched condition")
				return s.FindNodeByID(flow, outgoingEdges[edgeIndex].Target)
			}
		}
	}

	// Handle default condition
	for i, conditionInterface := range conditions {
		condition, ok := conditionInterface.(map[string]interface{})
		if !ok {
			continue
		}

		conditionType, _ := condition["type"].(string)
		if conditionType == "default" {
			edgeIndex := i % len(outgoingEdges)
			if edgeIndex < len(outgoingEdges) {
				logrus.Info("CONDITION FIX: Using default condition")
				return s.FindNodeByID(flow, outgoingEdges[edgeIndex].Target)
			}
		}
	}

	// Fallback to first edge
	if len(outgoingEdges) > 0 {
		logrus.Warn("CONDITION FIX: No match, using first edge")
		return s.FindNodeByID(flow, outgoingEdges[0].Target)
	}

	return nil, fmt.Errorf("no valid next node found")
}
