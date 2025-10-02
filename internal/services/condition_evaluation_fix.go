package services

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"nodepath-chat/internal/models"
	"strconv"
	"strings"
)

// EvaluateConditionNodeFixed properly matches user input with condition edges
// This fixes the issue where edges are numbered 1,2,3,4 but user input "3" was calling edge 2
func (s *FlowService) EvaluateConditionNodeFixed(flow *models.ChatbotFlow, conditionNodeID string, userInput string) (*models.FlowNode, error) {
	// Get the condition node
	conditionNode, err := s.FindNodeByID(flow, conditionNodeID)
	if err != nil || conditionNode == nil {
		return nil, fmt.Errorf("condition node not found: %s", conditionNodeID)
	}

	// Get all edges from the flow
	edges, err := s.GetFlowEdges(flow)
	if err != nil {
		return nil, err
	}

	// Get all outgoing edges from this condition node
	var outgoingEdges []models.FlowEdge
	for _, edge := range edges {
		if edge.Source == conditionNodeID {
			outgoingEdges = append(outgoingEdges, *edge)
		}
	}

	if len(outgoingEdges) == 0 {
		return nil, fmt.Errorf("no outgoing edges from condition node %s", conditionNodeID)
	}

	// Get conditions from node data
	conditions, ok := conditionNode.Data["conditions"].([]interface{})
	if !ok {
		// No conditions defined - check if user input is a number matching edge index
		return s.evaluateDirectEdgeSelection(flow, userInput, outgoingEdges)
	}

	// Normalize user input
	userInputLower := strings.ToLower(strings.TrimSpace(userInput))

	logrus.WithFields(logrus.Fields{
		"user_input":       userInput,
		"conditions_count": len(conditions),
		"edges_count":      len(outgoingEdges),
		"node_id":          conditionNodeID,
	}).Info("🔍 CONDITION: Evaluating user input against conditions")

	// CRITICAL FIX: First check if user input is a direct edge number (1, 2, 3, 4, etc.)
	if edgeNum, err := strconv.Atoi(userInput); err == nil {
		// User entered a number - treat it as direct edge selection
		edgeIndex := edgeNum - 1 // Convert to 0-based index
		if edgeIndex >= 0 && edgeIndex < len(outgoingEdges) {
			targetNodeID := outgoingEdges[edgeIndex].Target
			logrus.WithFields(logrus.Fields{
				"user_input":  userInput,
				"edge_number": edgeNum,
				"edge_index":  edgeIndex,
				"target_node": targetNodeID,
			}).Info("✅ CONDITION: User selected edge by number")
			return s.FindNodeByID(flow, targetNodeID)
		}
	}

	// Check each condition for matches
	for i, conditionInterface := range conditions {
		condition, ok := conditionInterface.(map[string]interface{})
		if !ok {
			continue
		}

		conditionType, _ := condition["type"].(string)
		conditionValue, _ := condition["value"].(string)
		conditionLabel, _ := condition["label"].(string) // Edge label like "1", "2", "3", "4"

		// Skip default conditions for now
		if conditionType == "default" {
			continue
		}

		// CRITICAL FIX: Check if user input matches the condition label exactly
		if conditionLabel != "" && userInput == conditionLabel {
			// User input matches the edge label exactly (e.g., user says "3" for edge labeled "3")
			if i < len(outgoingEdges) {
				targetNodeID := outgoingEdges[i].Target
				logrus.WithFields(logrus.Fields{
					"matched_label":   conditionLabel,
					"condition_index": i,
					"target_node":     targetNodeID,
				}).Info("✅ CONDITION: Matched by edge label")
				return s.FindNodeByID(flow, targetNodeID)
			}
		}

		// Check condition value match
		if conditionValue != "" {
			conditionValueLower := strings.ToLower(strings.TrimSpace(conditionValue))
			var matches bool

			switch conditionType {
			case "equals":
				// Exact match (case-insensitive)
				matches = userInputLower == conditionValueLower
			case "contains":
				// Contains match (case-insensitive)
				matches = strings.Contains(userInputLower, conditionValueLower)
			default:
				// Default to contains for backward compatibility
				matches = strings.Contains(userInputLower, conditionValueLower)
			}

			if matches {
				// Match found - use the corresponding edge
				if i < len(outgoingEdges) {
					targetNodeID := outgoingEdges[i].Target
					logrus.WithFields(logrus.Fields{
						"condition_type":  conditionType,
						"condition_value": conditionValue,
						"condition_index": i,
						"target_node":     targetNodeID,
					}).Info("✅ CONDITION: Matched by condition value")
					return s.FindNodeByID(flow, targetNodeID)
				}
			}
		}
	}

	// No match found - look for default condition
	for i, conditionInterface := range conditions {
		condition, ok := conditionInterface.(map[string]interface{})
		if !ok {
			continue
		}

		conditionType, _ := condition["type"].(string)
		if conditionType == "default" {
			if i < len(outgoingEdges) {
				targetNodeID := outgoingEdges[i].Target
				logrus.WithFields(logrus.Fields{
					"condition_index": i,
					"target_node":     targetNodeID,
				}).Info("⚡ CONDITION: Using default condition")
				return s.FindNodeByID(flow, targetNodeID)
			}
		}
	}

	// Fallback to first edge if no conditions match and no default
	if len(outgoingEdges) > 0 {
		targetNodeID := outgoingEdges[0].Target
		logrus.WithFields(logrus.Fields{
			"target_node": targetNodeID,
		}).Warn("⚠️ CONDITION: No match found, using first edge as fallback")
		return s.FindNodeByID(flow, targetNodeID)
	}

	return nil, fmt.Errorf("no valid next node found from condition evaluation")
}

// evaluateDirectEdgeSelection handles cases where user input is a direct edge number
func (s *FlowService) evaluateDirectEdgeSelection(flow *models.ChatbotFlow, userInput string, outgoingEdges []models.FlowEdge) (*models.FlowNode, error) {
	// Try to parse user input as a number
	if edgeNum, err := strconv.Atoi(userInput); err == nil {
		// User entered a number - treat it as edge selection (1-based)
		edgeIndex := edgeNum - 1 // Convert to 0-based index
		if edgeIndex >= 0 && edgeIndex < len(outgoingEdges) {
			targetNodeID := outgoingEdges[edgeIndex].Target
			logrus.WithFields(logrus.Fields{
				"user_input":  userInput,
				"edge_number": edgeNum,
				"target_node": targetNodeID,
			}).Info("✅ CONDITION: Direct edge selection by number")
			return s.FindNodeByID(flow, targetNodeID)
		}
	}

	// Not a number or out of range - use first edge as fallback
	if len(outgoingEdges) > 0 {
		return s.FindNodeByID(flow, outgoingEdges[0].Target)
	}

	return nil, fmt.Errorf("no valid edge found")
}
