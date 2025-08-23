package whatsapp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"
)

// FlowEngine handles the execution of chatbot flows with proper node sequencing
type FlowEngine struct {
	flowService           *services.FlowService
	aiWhatsappService     services.AIWhatsappService
	aiService             *services.AIService
	providerService       *services.ProviderService
	deviceSettingsService *services.DeviceSettingsService
}

// NewFlowEngine creates a new flow engine instance
func NewFlowEngine(
	flowService *services.FlowService,
	aiWhatsappService services.AIWhatsappService,
	aiService *services.AIService,
	providerService *services.ProviderService,
	deviceSettingsService *services.DeviceSettingsService,
) *FlowEngine {
	return &FlowEngine{
		flowService:           flowService,
		aiWhatsappService:     aiWhatsappService,
		aiService:             aiService,
		providerService:       providerService,
		deviceSettingsService: deviceSettingsService,
	}
}

// ExecutionContext holds the context for flow execution
type ExecutionContext struct {
	Flow        *models.ChatbotFlow
	Execution   *models.AIWhatsapp
	UserInput   string
	Variables   map[string]interface{}
	CurrentNode *models.FlowNode
	Response    []string
	ShouldStop  bool
}

// ExecuteFlow executes a flow from the current node position
func (e *FlowEngine) ExecuteFlow(execution *models.AIWhatsapp, userInput string) error {
	logrus.WithFields(logrus.Fields{
		"execution_id": execution.ExecutionID.String,
		"flow_ref":     execution.FlowReference.String,
		"phone_number": execution.ProspectNum,
	}).Info("🚀 FLOW_ENGINE: Starting flow execution")

	// Get flow data
	flow, err := e.flowService.GetFlow(execution.FlowReference.String)
	if err != nil {
		return fmt.Errorf("failed to get flow: %w", err)
	}

	if flow == nil {
		return fmt.Errorf("flow not found: %s", execution.FlowReference.String)
	}

	// Initialize execution context
	ctx := &ExecutionContext{
		Flow:      flow,
		Execution: execution,
		UserInput: userInput,
		Variables: make(map[string]interface{}),
		Response:  []string{},
	}

	// Load existing variables if any
	if execution.Variables != nil {
		err = json.Unmarshal(execution.Variables, &ctx.Variables)
		if err != nil {
			logrus.WithError(err).Warn("Failed to load execution variables")
		}
	}

	// Execute the flow step by step
	err = e.executeFlowSteps(ctx)
	if err != nil {
		return fmt.Errorf("flow execution failed: %w", err)
	}

	// Send accumulated responses
	if len(ctx.Response) > 0 {
		err = e.sendFlowResponses(ctx)
		if err != nil {
			return fmt.Errorf("failed to send responses: %w", err)
		}
	}

	// Note: Conversation history is saved by the AI conversation processing in whatsapp_service.go
	// to prevent duplicate saves and data conflicts

	return nil
}

// executeFlowSteps executes flow steps following the designed node sequence
func (e *FlowEngine) executeFlowSteps(ctx *ExecutionContext) error {
	logrus.Info("🔄 FLOW_ENGINE: Starting flow step execution")

	// Get current node or start from beginning
	currentNode, err := e.getCurrentOrStartNode(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current node: %w", err)
	}

	ctx.CurrentNode = currentNode
	maxSteps := 50 // Prevent infinite loops
	stepCount := 0

	// Execute nodes in sequence until we hit a stopping point
	for stepCount < maxSteps && !ctx.ShouldStop {
		stepCount++

		logrus.WithFields(logrus.Fields{
			"step":      stepCount,
			"node_id":   ctx.CurrentNode.ID,
			"node_type": ctx.CurrentNode.Type,
		}).Info("🎯 FLOW_ENGINE: Processing node")

		// Process current node
		err = e.processNode(ctx)
		if err != nil {
			return fmt.Errorf("failed to process node %s: %w", ctx.CurrentNode.ID, err)
		}

		// If node requested stop, break the loop
		if ctx.ShouldStop {
			logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🛑 FLOW_ENGINE: Node requested stop")
			break
		}

		// Move to next node
		nextNode, err := e.getNextNode(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next node: %w", err)
		}

		if nextNode == nil {
			// End of flow reached
			logrus.Info("🏁 FLOW_ENGINE: End of flow reached")
			err = e.completeFlowExecution(ctx)
			if err != nil {
				logrus.WithError(err).Error("Failed to complete flow execution")
			}
			break
		}

		// Update current node and execution state
		ctx.CurrentNode = nextNode
		err = e.updateExecutionState(ctx)
		if err != nil {
			logrus.WithError(err).Error("Failed to update execution state")
		}

		logrus.WithFields(logrus.Fields{
			"from_node": ctx.CurrentNode.ID,
			"to_node":   nextNode.ID,
		}).Info("➡️ FLOW_ENGINE: Advanced to next node")
	}

	if stepCount >= maxSteps {
		logrus.Warn("⚠️ FLOW_ENGINE: Maximum steps reached, stopping execution")
	}

	return nil
}

// getCurrentOrStartNode gets the current node or starts from the beginning
func (e *FlowEngine) getCurrentOrStartNode(ctx *ExecutionContext) (*models.FlowNode, error) {
	// If we have a current node, try to get it
	if ctx.Execution.CurrentNode.Valid && ctx.Execution.CurrentNode.String != "" {
		node, err := e.flowService.FindNodeByID(ctx.Flow, ctx.Execution.CurrentNode.String)
		if err == nil && node != nil {
			// Special handling for user_reply nodes - advance to next node when user sends input
			if node.Type == models.NodeTypeUserReply {
				logrus.WithField("node_id", node.ID).Info("📍 FLOW_ENGINE: User replied to user_reply node, advancing to next node")
				
				// Get the next node after user_reply
				nextNode, err := e.flowService.GetNextNode(ctx.Flow, node.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to get next node after user_reply: %w", err)
				}
				
				if nextNode != nil {
					logrus.WithFields(logrus.Fields{
						"from_node": node.ID,
						"to_node": nextNode.ID,
						"next_type": nextNode.Type,
					}).Info("➡️ FLOW_ENGINE: Advanced from user_reply to next node")
					return nextNode, nil
				}
				
				// If no next node, flow is complete
				logrus.WithField("node_id", node.ID).Info("🏁 FLOW_ENGINE: No next node after user_reply, flow complete")
				return nil, nil
			}
			
			logrus.WithField("node_id", node.ID).Info("📍 FLOW_ENGINE: Resuming from current node")
			return node, nil
		}
		logrus.WithError(err).Warn("Current node not found, starting from beginning")
	}

	// Get start node
	startNode, err := e.flowService.GetStartNode(ctx.Flow)
	if err != nil {
		return nil, fmt.Errorf("failed to get start node: %w", err)
	}

	logrus.WithField("node_id", startNode.ID).Info("🚀 FLOW_ENGINE: Starting from start node")
	return startNode, nil
}

// getNextNode determines the next node based on current node and flow logic
func (e *FlowEngine) getNextNode(ctx *ExecutionContext) (*models.FlowNode, error) {
	// For condition nodes, we need special handling
	if ctx.CurrentNode.Type == models.NodeTypeCondition {
		return e.getNextNodeFromCondition(ctx)
	}

	// For regular nodes, follow the edge
	return e.flowService.GetNextNode(ctx.Flow, ctx.CurrentNode.ID)
}

// getNextNodeFromCondition handles condition node logic with proper evaluation
func (e *FlowEngine) getNextNodeFromCondition(ctx *ExecutionContext) (*models.FlowNode, error) {
	// Get all possible next nodes for this condition
	nextNodes, err := e.flowService.GetAllNextNodes(ctx.Flow, ctx.CurrentNode.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next nodes for condition: %w", err)
	}
	
	if len(nextNodes) == 0 {
		return nil, nil // End of flow
	}
	
	// Evaluate conditions based on user input
	selectedNode := e.evaluateConditions(ctx, nextNodes)
	if selectedNode != nil {
		logrus.WithFields(logrus.Fields{
			"condition_node": ctx.CurrentNode.ID,
			"user_input": ctx.UserInput,
			"available_paths": len(nextNodes),
			"selected_path": selectedNode.ID,
		}).Info("🔀 FLOW_ENGINE: Condition evaluated - selected matching path")
		return selectedNode, nil
	}
	
	// If no condition matches, look for default path or use first available
	defaultNode := e.findDefaultPath(ctx, nextNodes)
	if defaultNode != nil {
		logrus.WithFields(logrus.Fields{
			"condition_node": ctx.CurrentNode.ID,
			"user_input": ctx.UserInput,
			"selected_path": defaultNode.ID,
		}).Info("🔀 FLOW_ENGINE: No condition matched - using default path")
		return defaultNode, nil
	}
	
	// Fallback to first available node
	logrus.WithFields(logrus.Fields{
		"condition_node": ctx.CurrentNode.ID,
		"user_input": ctx.UserInput,
		"available_paths": len(nextNodes),
		"selected_path": nextNodes[0].ID,
	}).Warn("🔀 FLOW_ENGINE: No conditions or default found - using first available path")
	
	return nextNodes[0], nil
}

// updateExecutionState updates the execution state in the database
func (e *FlowEngine) updateExecutionState(ctx *ExecutionContext) error {
	// Update current node in execution record
	ctx.Execution.CurrentNode.String = ctx.CurrentNode.ID
	ctx.Execution.CurrentNode.Valid = true
	
	// Parse existing variables from execution
	variables := make(map[string]interface{})
	if len(ctx.Execution.Variables) > 0 {
		err := json.Unmarshal(ctx.Execution.Variables, &variables)
		if err != nil {
			logrus.WithError(err).Warn("Failed to parse existing variables, using empty map")
		}
	}
	
	// Add any new variables from context
	for key, value := range ctx.Variables {
		variables[key] = value
	}
	
	// Update flow execution state in database
	err := e.aiWhatsappService.UpdateFlowExecution(
		ctx.Execution.ProspectNum,
		ctx.Execution.IDDevice,
		ctx.CurrentNode.ID,
		variables,
		"active",
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to update flow execution state")
		return fmt.Errorf("failed to update flow execution state: %w", err)
	}
	
	// Update conversation stage if changed
	if ctx.Execution.Stage != "" {
		stageErr := e.aiWhatsappService.UpdateConversationStage(ctx.Execution.ProspectNum, ctx.Execution.Stage)
		if stageErr != nil {
			logrus.WithError(stageErr).Warn("Failed to update conversation stage")
		}
	}
	
	logrus.WithFields(logrus.Fields{
		"prospect_num": ctx.Execution.ProspectNum,
		"current_node": ctx.CurrentNode.ID,
		"variables_count": len(variables),
	}).Info("💾 FLOW_ENGINE: Execution state updated successfully")
	
	return nil
}

// completeFlowExecution marks the flow execution as completed
func (e *FlowEngine) completeFlowExecution(ctx *ExecutionContext) error {
	// Update conversation stage to completed if needed
	if ctx.Execution.Stage != "" {
		return e.aiWhatsappService.UpdateConversationStage(ctx.Execution.ProspectNum, ctx.Execution.Stage)
	}
	return nil
}

// evaluateConditions evaluates user input against condition edges to find matching path
func (e *FlowEngine) evaluateConditions(ctx *ExecutionContext, nextNodes []*models.FlowNode) *models.FlowNode {
	userInput := strings.ToLower(strings.TrimSpace(ctx.UserInput))
	
	// Get edges from current condition node to evaluate conditions
	edges, err := e.flowService.GetEdgesFromNode(ctx.Flow, ctx.CurrentNode.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get edges for condition evaluation")
		return nil
	}
	
	// Evaluate each edge condition
	for _, edge := range edges {
		if edge.Condition == nil {
			continue
		}
		
		// Get condition data
		conditionType, _ := edge.Condition["type"].(string)
		conditionValue, _ := edge.Condition["value"].(string)
		
		if conditionType == "" || conditionValue == "" {
			continue
		}
		
		// Evaluate condition based on type
		matches := false
		switch conditionType {
		case "contains":
			matches = strings.Contains(userInput, strings.ToLower(conditionValue))
		case "equals":
			matches = userInput == strings.ToLower(conditionValue)
		case "starts_with":
			matches = strings.HasPrefix(userInput, strings.ToLower(conditionValue))
		case "ends_with":
			matches = strings.HasSuffix(userInput, strings.ToLower(conditionValue))
		case "regex":
			// TODO: Implement regex matching if needed
			logrus.Warn("Regex condition type not implemented yet")
		}
		
		if matches {
			// Find the target node for this edge
			for _, node := range nextNodes {
				if node.ID == edge.Target {
					logrus.WithFields(logrus.Fields{
						"condition_type": conditionType,
						"condition_value": conditionValue,
						"user_input": ctx.UserInput,
						"target_node": node.ID,
					}).Info("🎯 FLOW_ENGINE: Condition matched")
					return node
				}
			}
		}
	}
	
	return nil
}

// findDefaultPath finds a default path among the available next nodes by checking edge conditions
func (e *FlowEngine) findDefaultPath(ctx *ExecutionContext, nextNodes []*models.FlowNode) *models.FlowNode {
	// Get edges from current condition node to find default path
	edges, err := e.flowService.GetEdgesFromNode(ctx.Flow, ctx.CurrentNode.ID)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get edges for default path evaluation")
		return nil
	}
	
	// Look for edges with default condition type
	for _, edge := range edges {
		if edge.Condition != nil {
			conditionType, _ := edge.Condition["type"].(string)
			if conditionType == "default" {
				// Find the target node for this default edge
				for _, node := range nextNodes {
					if node.ID == edge.Target {
						logrus.WithFields(logrus.Fields{
							"default_edge": edge.ID,
							"target_node": node.ID,
						}).Info("🎯 FLOW_ENGINE: Found default path")
						return node
					}
				}
			}
		}
	}
	
	// Fallback: look for nodes that are marked as default in their data
	for _, node := range nextNodes {
		if node.Data != nil {
			if isDefault, ok := node.Data["isDefault"].(bool); ok && isDefault {
				return node
			}
			if defaultFlag, ok := node.Data["default"].(bool); ok && defaultFlag {
				return node
			}
		}
	}
	
	return nil
}

// sendFlowResponses sends all accumulated responses
func (e *FlowEngine) sendFlowResponses(ctx *ExecutionContext) error {
	for _, response := range ctx.Response {
		// Filter out empty, nil, or invalid responses
		if response != "" && response != "<nil>" && response != "nil" && strings.TrimSpace(response) != "" {
			err := e.sendSingleResponse(ctx.Execution.ProspectNum, ctx.Execution.IDDevice, response)
			if err != nil {
				return err
			}
			// Add small delay between messages
			time.Sleep(500 * time.Millisecond)
		} else {
			logrus.WithField("response", response).Warn("🚫 FLOW_ENGINE: Filtered out invalid response")
		}
	}
	return nil
}

// sendSingleResponse sends a single response message using the provider service
func (e *FlowEngine) sendSingleResponse(phoneNumber, deviceID, message string) error {
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"message":      message,
	}).Info("📤 FLOW_ENGINE: Sending response")

	// Get device settings for provider configuration
	deviceSettings, err := e.deviceSettingsService.GetByIDDevice(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device settings: %w", err)
	}

	// Send message using provider service
	err = e.providerService.SendMessage(deviceSettings, phoneNumber, message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// saveConversationHistory saves the conversation to history
// saveConversationHistory function removed to prevent duplicate conversation saves
// Conversation history is now handled exclusively by AI conversation processing in whatsapp_service.go