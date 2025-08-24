package services

import (
	"encoding/json"
	"fmt"
	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// FlowExecutionService handles the execution of chatbot flows
type FlowExecutionService interface {
	// ProcessFlowForNewUser processes the flow for a new user and returns AI prompt if found
	ProcessFlowForNewUser(idDevice, prospectNum string) (*FlowExecutionResult, error)
	
	// ProcessFlowForExistingUser processes the flow for an existing user based on current stage
	ProcessFlowForExistingUser(idDevice, prospectNum, currentStage, userInput string) (*FlowExecutionResult, error)
	
	// GetAIPromptFromNode extracts AI prompt content from a flow node
	GetAIPromptFromNode(node *models.FlowNode) (string, error)
	
	// HasAIPromptNodes checks if a flow contains any AI prompt nodes
	HasAIPromptNodes(flow *models.ChatbotFlow) bool
}

// FlowExecutionResult represents the result of flow execution
type FlowExecutionResult struct {
	ShouldUseAI     bool   `json:"should_use_ai"`
	AIPromptContent string `json:"ai_prompt_content,omitempty"`
	CurrentStage    string `json:"current_stage"`
	NextStage       string `json:"next_stage,omitempty"`
	FlowID          string `json:"flow_id,omitempty"`
	NodeID          string `json:"node_id,omitempty"`
	Message         string `json:"message,omitempty"`
}

// flowExecutionService implements FlowExecutionService
type flowExecutionService struct {
	flowService *FlowService
}

// NewFlowExecutionService creates a new flow execution service
func NewFlowExecutionService(flowService *FlowService) FlowExecutionService {
	return &flowExecutionService{
		flowService: flowService,
	}
}

// ProcessFlowForNewUser processes the flow for a new user and returns AI prompt if found
func (s *flowExecutionService) ProcessFlowForNewUser(idDevice, prospectNum string) (*FlowExecutionResult, error) {
	logrus.WithFields(logrus.Fields{
		"id_device":    idDevice,
		"prospect_num": prospectNum,
	}).Info("🔄 FLOW: Processing flow for new user")

	// Get default flow for the device
	defaultFlow, err := s.flowService.GetDefaultFlowForDevice(idDevice)
	if err != nil {
		logrus.WithError(err).Warn("⚠️ FLOW: Failed to get default flow for device")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: "welcome",
		}, nil
	}

	if defaultFlow == nil {
		logrus.WithField("id_device", idDevice).Info("ℹ️ FLOW: No default flow found for device")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: "welcome",
		}, nil
	}

	// Check if flow has AI prompt nodes
	if !s.HasAIPromptNodes(defaultFlow) {
		logrus.WithFields(logrus.Fields{
			"flow_id":   defaultFlow.ID,
			"flow_name": defaultFlow.Name,
		}).Info("ℹ️ FLOW: Flow has no AI prompt nodes, using fallback")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: "welcome",
			FlowID:       defaultFlow.ID,
		}, nil
	}

	// Get start node from the flow
	startNode, err := s.flowService.GetStartNode(defaultFlow)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get start node")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: "welcome",
			FlowID:       defaultFlow.ID,
		}, nil
	}

	if startNode == nil {
		logrus.WithField("flow_id", defaultFlow.ID).Error("❌ FLOW: No start node found in flow")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: "welcome",
			FlowID:       defaultFlow.ID,
		}, nil
	}

	// Check if start node is an AI prompt node
	if startNode.Type == "ai_prompt" || startNode.Type == "advanced_ai_prompt" || startNode.Type == "prompt" {
		aiPromptContent, err := s.GetAIPromptFromNode(startNode)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to extract AI prompt from start node")
			return &FlowExecutionResult{
				ShouldUseAI:  false,
				CurrentStage: startNode.ID,
				FlowID:       defaultFlow.ID,
				NodeID:       startNode.ID,
			}, nil
		}

		logrus.WithFields(logrus.Fields{
			"flow_id":   defaultFlow.ID,
			"node_id":   startNode.ID,
			"node_type": startNode.Type,
		}).Info("✅ FLOW: Found AI prompt in start node for new user")

		return &FlowExecutionResult{
			ShouldUseAI:     true,
			AIPromptContent: aiPromptContent,
			CurrentStage:    startNode.ID,
			FlowID:          defaultFlow.ID,
			NodeID:          startNode.ID,
		}, nil
	}

	// Start node is not an AI prompt, check if we should navigate to next AI prompt node
	logrus.WithFields(logrus.Fields{
		"flow_id":        defaultFlow.ID,
		"start_node_id":  startNode.ID,
		"start_node_type": startNode.Type,
	}).Info("ℹ️ FLOW: Start node is not AI prompt, looking for next AI prompt node")

	// For now, return the start node as current stage without AI processing
	// This can be extended later to navigate through the flow
	return &FlowExecutionResult{
		ShouldUseAI:  false,
		CurrentStage: startNode.ID,
		FlowID:       defaultFlow.ID,
		NodeID:       startNode.ID,
	}, nil
}

// ProcessFlowForExistingUser processes the flow for an existing user based on current stage
func (s *flowExecutionService) ProcessFlowForExistingUser(idDevice, prospectNum, currentStage, userInput string) (*FlowExecutionResult, error) {
	logrus.WithFields(logrus.Fields{
		"id_device":     idDevice,
		"prospect_num":  prospectNum,
		"current_stage": currentStage,
	}).Info("🔄 FLOW: Processing flow for existing user")

	// Get flow for the device
	flows, err := s.flowService.GetFlowsByDevice(idDevice)
	if err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to get flows for device")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: currentStage,
		}, nil
	}

	if len(flows) == 0 {
		logrus.WithField("id_device", idDevice).Info("ℹ️ FLOW: No flows found for device")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: currentStage,
		}, nil
	}

	// Find the flow that contains the current stage
	var targetFlow *models.ChatbotFlow
	var currentNode *models.FlowNode

	for _, flow := range flows {
		node, err := s.flowService.FindNodeByID(flow, currentStage)
		if err == nil && node != nil {
			targetFlow = flow
			currentNode = node
			break
		}
	}

	if targetFlow == nil || currentNode == nil {
		logrus.WithFields(logrus.Fields{
			"current_stage": currentStage,
			"id_device":     idDevice,
		}).Warn("⚠️ FLOW: Current stage not found in any flow, using fallback")
		return &FlowExecutionResult{
			ShouldUseAI:  false,
			CurrentStage: currentStage,
		}, nil
	}

	// Check if current node is an AI prompt node
	if currentNode.Type == "ai_prompt" || currentNode.Type == "advanced_ai_prompt" || currentNode.Type == "prompt" {
		aiPromptContent, err := s.GetAIPromptFromNode(currentNode)
		if err != nil {
			logrus.WithError(err).Error("❌ FLOW: Failed to extract AI prompt from current node")
			return &FlowExecutionResult{
				ShouldUseAI:  false,
				CurrentStage: currentStage,
				FlowID:       targetFlow.ID,
				NodeID:       currentNode.ID,
			}, nil
		}

		logrus.WithFields(logrus.Fields{
			"flow_id":   targetFlow.ID,
			"node_id":   currentNode.ID,
			"node_type": currentNode.Type,
		}).Info("✅ FLOW: Found AI prompt in current node for existing user")

		return &FlowExecutionResult{
			ShouldUseAI:     true,
			AIPromptContent: aiPromptContent,
			CurrentStage:    currentStage,
			FlowID:          targetFlow.ID,
			NodeID:          currentNode.ID,
		}, nil
	}

	// Current node is not an AI prompt, return without AI processing
	logrus.WithFields(logrus.Fields{
		"flow_id":   targetFlow.ID,
		"node_id":   currentNode.ID,
		"node_type": currentNode.Type,
	}).Info("ℹ️ FLOW: Current node is not AI prompt")

	return &FlowExecutionResult{
		ShouldUseAI:  false,
		CurrentStage: currentStage,
		FlowID:       targetFlow.ID,
		NodeID:       currentNode.ID,
	}, nil
}

// GetAIPromptFromNode extracts AI prompt content from a flow node
func (s *flowExecutionService) GetAIPromptFromNode(node *models.FlowNode) (string, error) {
	if node == nil {
		return "", fmt.Errorf("node is nil")
	}

	if node.Data == nil {
		return "", fmt.Errorf("node data is nil")
	}

	// Extract system prompt from node data
	var systemPrompt string
	if promptData, exists := node.Data["systemPrompt"]; exists {
		if promptStr, ok := promptData.(string); ok {
			systemPrompt = promptStr
		}
	}

	// Also check for 'prompt' field as fallback
	if systemPrompt == "" {
		if promptData, exists := node.Data["prompt"]; exists {
			if promptStr, ok := promptData.(string); ok {
				systemPrompt = promptStr
			}
		}
	}

	// Check for 'content' field as another fallback
	if systemPrompt == "" {
		if contentData, exists := node.Data["content"]; exists {
			if contentStr, ok := contentData.(string); ok {
				systemPrompt = contentStr
			}
		}
	}

	if systemPrompt == "" {
		return "", fmt.Errorf("no AI prompt content found in node %s", node.ID)
	}

	logrus.WithFields(logrus.Fields{
		"node_id":   node.ID,
		"node_type": node.Type,
		"prompt_length": len(systemPrompt),
	}).Debug("🔍 FLOW: Extracted AI prompt from node")

	return systemPrompt, nil
}

// HasAIPromptNodes checks if a flow contains any AI prompt nodes
func (s *flowExecutionService) HasAIPromptNodes(flow *models.ChatbotFlow) bool {
	if flow == nil {
		return false
	}

	// Parse nodes from JSON
	var nodes []models.FlowNode
	if err := json.Unmarshal(*flow.Nodes, &nodes); err != nil {
		logrus.WithError(err).Error("❌ FLOW: Failed to parse flow nodes")
		return false
	}

	// Check each node for AI prompt types
	for _, node := range nodes {
		if node.Type == "ai_prompt" || node.Type == "advanced_ai_prompt" || node.Type == "prompt" {
			logrus.WithFields(logrus.Fields{
				"flow_id":   flow.ID,
				"node_id":   node.ID,
				"node_type": node.Type,
			}).Debug("🔍 FLOW: Found AI prompt node in flow")
			return true
		}
	}

	logrus.WithFields(logrus.Fields{
		"flow_id":   flow.ID,
		"flow_name": flow.Name,
		"total_nodes": len(nodes),
	}).Debug("🔍 FLOW: No AI prompt nodes found in flow")

	return false
}