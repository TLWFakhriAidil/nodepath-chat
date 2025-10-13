package services

import (
	"database/sql"
	"fmt"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// UnifiedFlowService handles flow execution with table routing based on flow name
type UnifiedFlowService struct {
	flowService    *FlowService
	aiWhatsappRepo repository.AIWhatsappRepository
	wasapBotRepo   repository.WasapBotRepository
}

// NewUnifiedFlowService creates a new unified flow service
func NewUnifiedFlowService(
	flowService *FlowService,
	aiWhatsappRepo repository.AIWhatsappRepository,
	wasapBotRepo repository.WasapBotRepository,
) *UnifiedFlowService {
	return &UnifiedFlowService{
		flowService:    flowService,
		aiWhatsappRepo: aiWhatsappRepo,
		wasapBotRepo:   wasapBotRepo,
	}
}

// AcquireAIWhatsappSession attempts to acquire a session lock for AI WhatsApp flows
func (s *UnifiedFlowService) AcquireAIWhatsappSession(phoneNumber, deviceID string) (bool, error) {
	if s.aiWhatsappRepo == nil {
		return false, fmt.Errorf("aiWhatsappRepo is not initialized")
	}

	acquired, err := s.aiWhatsappRepo.TryAcquireSession(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Error("Failed to acquire AI WhatsApp session lock")
		return false, err
	}

	if !acquired {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Warn("AI WhatsApp session lock already active - ignoring duplicate message")
	}

	return acquired, nil
}

// ReleaseAIWhatsappSession releases a session lock for AI WhatsApp flows
func (s *UnifiedFlowService) ReleaseAIWhatsappSession(phoneNumber, deviceID string) error {
	if s.aiWhatsappRepo == nil {
		return fmt.Errorf("aiWhatsappRepo is not initialized")
	}

	if err := s.aiWhatsappRepo.ReleaseSession(phoneNumber, deviceID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Error("Failed to release AI WhatsApp session lock")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Debug("AI WhatsApp session lock released")
	return nil
}

// AcquireWasapBotSession attempts to acquire a session lock for WasapBot flows
func (s *UnifiedFlowService) AcquireWasapBotSession(phoneNumber, deviceID string) (bool, error) {
	if s.wasapBotRepo == nil {
		return false, fmt.Errorf("wasapBotRepo is not initialized")
	}

	acquired, err := s.wasapBotRepo.TryAcquireSession(phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Error("Failed to acquire WasapBot session lock")
		return false, err
	}

	if !acquired {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Warn("WasapBot session lock already active - ignoring duplicate message")
	}

	return acquired, nil
}

// ReleaseWasapBotSession releases a session lock for WasapBot flows
func (s *UnifiedFlowService) ReleaseWasapBotSession(phoneNumber, deviceID string) error {
	if s.wasapBotRepo == nil {
		return fmt.Errorf("wasapBotRepo is not initialized")
	}

	if err := s.wasapBotRepo.ReleaseSession(phoneNumber, deviceID); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Error("Failed to release WasapBot session lock")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Debug("WasapBot session lock released")
	return nil
}

// GetActiveExecutionByFlow retrieves active execution based on flow name
func (s *UnifiedFlowService) GetActiveExecutionByFlow(phoneNumber, deviceID, flowID string) (interface{}, string, error) {
	// Get flow to determine which table to use
	flow, tableName, err := s.flowService.GetFlowAndDetermineTable(flowID)
	if err != nil {
		return nil, "", err
	}

	logrus.WithFields(logrus.Fields{
		"flow_name":    flow.Name,
		"table_name":   tableName,
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("Checking for active execution in determined table")

	// Route to appropriate table
	if tableName == "wasapBot_nodepath" {
		execution, err := s.wasapBotRepo.GetActiveExecution(phoneNumber, deviceID)
		return execution, tableName, err
	}

	// Default to ai_whatsapp_nodepath - get any execution with active status
	execution, err := s.aiWhatsappRepo.GetAIWhatsappByProspectAndDevice(phoneNumber, deviceID)
	if err != nil {
		return nil, "ai_whatsapp_nodepath", err
	}

	// Check if execution is active
	if execution != nil && execution.ExecutionStatus.Valid && execution.ExecutionStatus.String == "active" {
		return execution, "ai_whatsapp_nodepath", nil
	}

	return nil, "ai_whatsapp_nodepath", nil
}

// CreateExecutionByFlow creates new execution in appropriate table based on flow name
func (s *UnifiedFlowService) CreateExecutionByFlow(phoneNumber, deviceID, flowID, startNodeID, prospectName string) (string, string, error) {
	// Get flow to determine which table to use
	flow, tableName, err := s.flowService.GetFlowAndDetermineTable(flowID)
	if err != nil {
		return "", "", err
	}

	executionID := fmt.Sprintf("exec_%s_%s", flowID, uuid.New().String())

	logrus.WithFields(logrus.Fields{
		"flow_name":    flow.Name,
		"table_name":   tableName,
		"execution_id": executionID,
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("Creating new execution in determined table")

	// Route to appropriate table
	if tableName == "wasapBot_nodepath" {
		// Default prospect name if empty
		if prospectName == "" {
			prospectName = "Sis"
		}

		wasapBot := &models.WasapBot{
			FlowReference:   sql.NullString{String: flowID, Valid: true},
			ExecutionID:     sql.NullString{String: executionID, Valid: true},
			ExecutionStatus: sql.NullString{String: "active", Valid: true},
			FlowID:          sql.NullString{String: flowID, Valid: true},
			CurrentNodeID:   sql.NullString{String: startNodeID, Valid: true},
			WaitingForReply: 0,
			IDDevice:        sql.NullString{String: deviceID, Valid: true},
			ProspectNum:     sql.NullString{String: phoneNumber, Valid: true},
			Nama:            sql.NullString{String: prospectName, Valid: true},
			Niche:           sql.NullString{String: flow.Niche, Valid: flow.Niche != ""},
			Stage:           sql.NullString{String: "welcome", Valid: true},
			Status:          sql.NullString{String: "Prospek", Valid: true},
		}

		err = s.wasapBotRepo.Create(wasapBot)
		if err != nil {
			return "", "", fmt.Errorf("failed to create WasapBot execution: %w", err)
		}

		return executionID, tableName, nil
	}

	// Default to ai_whatsapp_nodepath
	// Set intro based on flow name
	var introText string
	if flow.Name == "Chatbot AI" {
		introText = "Welcome to Chatbot AI flow"
	} else {
		introText = "Welcome" // Default intro for other flows
	}

	// Default prospect name if empty
	if prospectName == "" {
		prospectName = "Sis"
	}

	aiWhatsapp := &models.AIWhatsapp{
		FlowReference:   sql.NullString{String: flowID, Valid: true},
		ExecutionID:     sql.NullString{String: executionID, Valid: true},
		ExecutionStatus: sql.NullString{String: "active", Valid: true},
		FlowID:          sql.NullString{String: flowID, Valid: true},
		CurrentNodeID:   sql.NullString{String: startNodeID, Valid: true},
		WaitingForReply: sql.NullInt32{Int32: 0, Valid: true},
		ProspectNum:     phoneNumber,
		IDDevice:        deviceID,
		ProspectName:    sql.NullString{String: prospectName, Valid: true},
		Intro:           sql.NullString{String: introText, Valid: true}, // Set intro based on flow
		Niche:           flow.Niche,
		Stage:           sql.NullString{}, // Leave stage as NULL initially
		Human:           0,
	}

	err = s.aiWhatsappRepo.CreateAIWhatsapp(aiWhatsapp)
	if err != nil {
		return "", "", fmt.Errorf("failed to create AI WhatsApp execution: %w", err)
	}

	return executionID, "ai_whatsapp_nodepath", nil
}

// UpdateExecutionNodeByFlow updates current node in appropriate table
func (s *UnifiedFlowService) UpdateExecutionNodeByFlow(executionID, nodeID, flowID string) error {
	// Get flow to determine which table to use
	_, tableName, err := s.flowService.GetFlowAndDetermineTable(flowID)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"table_name":   tableName,
		"execution_id": executionID,
		"node_id":      nodeID,
	}).Info("Updating execution node in determined table")

	// Route to appropriate table
	if tableName == "wasapBot_nodepath" {
		return s.wasapBotRepo.UpdateCurrentNode(executionID, nodeID)
	}

	// Default to ai_whatsapp_nodepath
	// Since we don't have a direct method to get by execution ID,
	// we'll need to add one or work around it
	// For now, let's just log an error
	logrus.WithField("execution_id", executionID).Error("Update by execution ID not fully implemented for ai_whatsapp_nodepath")
	return fmt.Errorf("update by execution ID not fully implemented for ai_whatsapp_nodepath")
}

// SaveConversationByFlow saves conversation in appropriate table
func (s *UnifiedFlowService) SaveConversationByFlow(phoneNumber, deviceID, userMessage, botResponse, stage, prospectName, flowID string) error {
	// Get flow to determine which table to use
	flow, tableName, err := s.flowService.GetFlowAndDetermineTable(flowID)
	var flowName string

	if err != nil {
		// If flow not found, try to determine by checking existing records
		logrus.WithError(err).Warn("Flow not found, checking existing records")

		// Check wasapBot first
		wasapBot, _ := s.wasapBotRepo.GetByProspectAndDevice(phoneNumber, deviceID)
		if wasapBot != nil {
			tableName = "wasapBot_nodepath"
			flowName = "WasapBot Exama (inferred)"
		} else {
			tableName = "ai_whatsapp_nodepath"
			flowName = "Chatbot AI (inferred)"
		}
	} else {
		if flow != nil {
			flowName = flow.Name
			tableName = s.flowService.DetermineTableByFlowName(flow.Name)
		} else {
			// Fallback if flow is nil
			tableName = "ai_whatsapp_nodepath"
			flowName = "Unknown"
		}
	}

	logrus.WithFields(logrus.Fields{
		"table_name":   tableName,
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"flow_id":      flowID,
		"flow_name":    flowName,
	}).Info("🗄️ SAVING CONVERSATION: Determined table for saving conversation")

	// Route to appropriate table
	if tableName == "wasapBot_nodepath" {
		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"flow_id":      flowID,
			"flow_name":    flowName,
		}).Info("💾 DATABASE: Saving to wasapBot_nodepath table")
		return s.wasapBotRepo.SaveConversationHistory(phoneNumber, deviceID, userMessage, botResponse, stage, prospectName)
	}

	// Default to ai_whatsapp_nodepath
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"flow_id":      flowID,
		"flow_name":    flowName,
	}).Info("💾 DATABASE: Saving to ai_whatsapp_nodepath table")
	return s.aiWhatsappRepo.SaveConversationHistory(phoneNumber, deviceID, userMessage, botResponse, stage, prospectName)
}

// UpdateWaitingStatusByFlow updates waiting status in appropriate table
func (s *UnifiedFlowService) UpdateWaitingStatusByFlow(executionID string, waitingValue int32, flowID string) error {
	// Get flow to determine which table to use
	_, tableName, err := s.flowService.GetFlowAndDetermineTable(flowID)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"table_name":    tableName,
		"execution_id":  executionID,
		"waiting_value": waitingValue,
	}).Info("Updating waiting status in determined table")

	// Route to appropriate table
	if tableName == "wasapBot_nodepath" {
		return s.wasapBotRepo.UpdateWaitingStatus(executionID, int(waitingValue))
	}

	// Default to ai_whatsapp_nodepath
	return s.aiWhatsappRepo.UpdateWaitingStatus(executionID, waitingValue)
}
