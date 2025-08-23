package whatsapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/services"
)

// FlowEngine handles the execution of chatbot flows with proper node sequencing
type FlowEngine struct {
	flowService           *services.FlowService
	aiWhatsappService     *services.AIWhatsappService
	aiService             *services.AIService
	providerService       *services.ProviderService
	deviceSettingsService *services.DeviceSettingsService
}

// NewFlowEngine creates a new flow engine instance
func NewFlowEngine(
	flowService *services.FlowService,
	aiWhatsappService *services.AIWhatsappService,
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

	// Save conversation history
	err = e.saveConversationHistory(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to save conversation history")
	}

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

// getNextNodeFromCondition handles condition node logic
func (e *FlowEngine) getNextNodeFromCondition(ctx *ExecutionContext) (*models.FlowNode, error) {
	// TODO: Implement condition evaluation logic
	// For now, just get the first available next node
	return e.flowService.GetNextNode(ctx.Flow, ctx.CurrentNode.ID)
}

// updateExecutionState updates the execution state in the database
func (e *FlowEngine) updateExecutionState(ctx *ExecutionContext) error {
	// Serialize variables
	variablesJSON, err := json.Marshal(ctx.Variables)
	if err != nil {
		return fmt.Errorf("failed to serialize variables: %w", err)
	}

	ctx.Execution.Variables = variablesJSON
	ctx.Execution.CurrentNode.String = ctx.CurrentNode.ID
	ctx.Execution.CurrentNode.Valid = true

	// Update in database
	return e.aiWhatsappService.UpdateFlowExecution(
		ctx.Execution.ProspectNum,
		ctx.Execution.IDDevice,
		ctx.CurrentNode.ID,
		ctx.Variables,
		"active",
	)
}

// completeFlowExecution marks the flow execution as completed
func (e *FlowEngine) completeFlowExecution(ctx *ExecutionContext) error {
	return e.aiWhatsappService.CompleteFlowExecution(
		ctx.Execution.ProspectNum,
		ctx.Execution.IDDevice,
	)
}

// sendFlowResponses sends all accumulated responses
func (e *FlowEngine) sendFlowResponses(ctx *ExecutionContext) error {
	for _, response := range ctx.Response {
		if response != "" {
			err := e.sendSingleResponse(ctx.Execution.ProspectNum, ctx.Execution.IDDevice, response)
			if err != nil {
				return err
			}
			// Add small delay between messages
			time.Sleep(500 * time.Millisecond)
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
	err = e.providerService.SendMessage(phoneNumber, message, deviceSettings)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// saveConversationHistory saves the conversation to history
func (e *FlowEngine) saveConversationHistory(ctx *ExecutionContext) error {
	responseText := ""
	if len(ctx.Response) > 0 {
		responseText = ctx.Response[0] // Use first response for history
	}

	return e.aiWhatsappService.SaveConversationHistory(
		ctx.Execution.ProspectNum,
		ctx.Execution.IDDevice,
		ctx.UserInput,
		responseText,
		"", // Stage will be managed separately
	)
}