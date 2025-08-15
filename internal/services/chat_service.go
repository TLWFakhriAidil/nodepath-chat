package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"nodepath-chat/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// ChatService handles chatbot conversation execution
type ChatService struct {
	db    *sql.DB
	redis *redis.Client
}

// NewChatService creates a new chat service
func NewChatService(db *sql.DB, redis *redis.Client) *ChatService {
	return &ChatService{
		db:    db,
		redis: redis,
	}
}

// StartExecution starts a new flow execution
func (s *ChatService) StartExecution(flowReference, phoneNumber, staffID string) (*models.ChatbotExecution, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}
	
	execution := &models.ChatbotExecution{
		ID:            uuid.New().String(),
		FlowReference: flowReference,
		PhoneNumber:   phoneNumber,
		StaffID:       staffID,
		Status:        models.ExecutionStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Initialize empty conversation and variables
	emptyConv, _ := json.Marshal([]models.ConversationMessage{})
	emptyVars, _ := json.Marshal(map[string]interface{}{})
	execution.ConvLast = emptyConv
	execution.Variables = emptyVars

	query := `
		INSERT INTO chatbot_executions_nodepath 
		(id, flow_reference, phone_number, staff_id, conv_last, conv_current, current_node, variables, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		execution.ID, execution.FlowReference, execution.PhoneNumber, execution.StaffID,
		execution.ConvLast, execution.ConvCurrent, execution.CurrentNode,
		execution.Variables, execution.Status, execution.CreatedAt, execution.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"execution_id":   execution.ID,
		"flow_reference": flowReference,
		"phone_number":   phoneNumber,
	}).Info("Execution started")

	return execution, nil
}

// GetExecution retrieves an execution by ID
func (s *ChatService) GetExecution(executionID string) (*models.ChatbotExecution, error) {
	query := `
		SELECT id, flow_reference, phone_number, staff_id, conv_last, conv_current, current_node, 
		       variables, status, created_at, updated_at
		FROM chatbot_executions_nodepath 
		WHERE id = ?
	`

	var execution models.ChatbotExecution
	err := s.db.QueryRow(query, executionID).Scan(
		&execution.ID, &execution.FlowReference, &execution.PhoneNumber, &execution.StaffID,
		&execution.ConvLast, &execution.ConvCurrent, &execution.CurrentNode,
		&execution.Variables, &execution.Status, &execution.CreatedAt, &execution.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get execution: %w", err)
	}

	return &execution, nil
}

// GetActiveExecution retrieves active execution for phone number and staff
func (s *ChatService) GetActiveExecution(phoneNumber, staffID string) (*models.ChatbotExecution, error) {
	query := `
		SELECT id, flow_reference, phone_number, staff_id, conv_last, conv_current, current_node, 
		       variables, status, created_at, updated_at
		FROM chatbot_executions_nodepath 
		WHERE phone_number = ? AND staff_id = ? AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1
	`

	var execution models.ChatbotExecution
	err := s.db.QueryRow(query, phoneNumber, staffID).Scan(
		&execution.ID, &execution.FlowReference, &execution.PhoneNumber, &execution.StaffID,
		&execution.ConvLast, &execution.ConvCurrent, &execution.CurrentNode,
		&execution.Variables, &execution.Status, &execution.CreatedAt, &execution.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active execution: %w", err)
	}

	return &execution, nil
}

// UpdateExecution updates an execution
func (s *ChatService) UpdateExecution(execution *models.ChatbotExecution) error {
	execution.UpdatedAt = time.Now()

	query := `
		UPDATE chatbot_executions_nodepath 
		SET conv_last = ?, conv_current = ?, current_node = ?, variables = ?, status = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		execution.ConvLast, execution.ConvCurrent, execution.CurrentNode,
		execution.Variables, execution.Status, execution.UpdatedAt, execution.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update execution: %w", err)
	}

	return nil
}

// AddConversationMessage adds a message to the conversation history
func (s *ChatService) AddConversationMessage(execution *models.ChatbotExecution, role, content string) error {
	// Parse existing conversation
	var messages []models.ConversationMessage
	if execution.ConvLast != nil {
		if err := json.Unmarshal(execution.ConvLast, &messages); err != nil {
			logrus.WithError(err).Warn("Failed to parse conversation history")
			messages = []models.ConversationMessage{}
		}
	}

	// Add new message
	newMessage := models.ConversationMessage{
		Role:    role,
		Content: content,
	}
	messages = append(messages, newMessage)

	// Update conversation
	updatedConv, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	execution.ConvLast = updatedConv
	if role == "BOT" {
		execution.ConvCurrent = content
	}

	return s.UpdateExecution(execution)
}

// GetConversationHistory returns the conversation history
func (s *ChatService) GetConversationHistory(execution *models.ChatbotExecution) ([]models.ConversationMessage, error) {
	var messages []models.ConversationMessage
	if execution.ConvLast != nil {
		if err := json.Unmarshal(execution.ConvLast, &messages); err != nil {
			return nil, fmt.Errorf("failed to parse conversation history: %w", err)
		}
	}
	return messages, nil
}

// SetExecutionVariable sets a variable in the execution context
func (s *ChatService) SetExecutionVariable(execution *models.ChatbotExecution, key string, value interface{}) error {
	// Parse existing variables
	var variables map[string]interface{}
	if execution.Variables != nil {
		if err := json.Unmarshal(execution.Variables, &variables); err != nil {
			logrus.WithError(err).Warn("Failed to parse execution variables")
			variables = make(map[string]interface{})
		}
	} else {
		variables = make(map[string]interface{})
	}

	// Set the variable
	variables[key] = value

	// Update variables
	updatedVars, err := json.Marshal(variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	execution.Variables = updatedVars
	return s.UpdateExecution(execution)
}

// GetExecutionVariables returns the execution variables
func (s *ChatService) GetExecutionVariables(execution *models.ChatbotExecution) (map[string]interface{}, error) {
	var variables map[string]interface{}
	if execution.Variables != nil {
		if err := json.Unmarshal(execution.Variables, &variables); err != nil {
			return nil, fmt.Errorf("failed to parse execution variables: %w", err)
		}
	} else {
		variables = make(map[string]interface{})
	}
	return variables, nil
}

// CompleteExecution marks an execution as completed
func (s *ChatService) CompleteExecution(executionID string) error {
	query := `UPDATE chatbot_executions_nodepath SET status = 'completed', updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, time.Now(), executionID)
	if err != nil {
		return fmt.Errorf("failed to complete execution: %w", err)
	}

	logrus.WithField("execution_id", executionID).Info("Execution completed")
	return nil
}

// FailExecution marks an execution as failed
func (s *ChatService) FailExecution(executionID string) error {
	query := `UPDATE chatbot_executions_nodepath SET status = 'failed', updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, time.Now(), executionID)
	if err != nil {
		return fmt.Errorf("failed to fail execution: %w", err)
	}

	logrus.WithField("execution_id", executionID).Info("Execution failed")
	return nil
}

// CreateOrUpdateLead creates or updates a lead
func (s *ChatService) CreateOrUpdateLead(phoneNumber, staffID, name, email string) error {
	leadID := uuid.New().String()
	now := time.Now()

	query := `
		INSERT INTO chatbot_leads (id, phone_number, staff_id, name, email, status, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'new', 'whatsapp', ?, ?)
		ON DUPLICATE KEY UPDATE
		name = VALUES(name), email = VALUES(email), updated_at = VALUES(updated_at)
	`

	_, err := s.db.Exec(query, leadID, phoneNumber, staffID, name, email, now, now)
	if err != nil {
		return fmt.Errorf("failed to create/update lead: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"staff_id":     staffID,
		"name":         name,
	}).Info("Lead created/updated")

	return nil
}

// GetExecutionsByFlow retrieves all executions for a flow
func (s *ChatService) GetExecutionsByFlow(flowReference string) ([]*models.ChatbotExecution, error) {
	query := `
		SELECT id, flow_reference, phone_number, staff_id, conv_last, conv_current, current_node, 
		       variables, status, created_at, updated_at
		FROM chatbot_executions_nodepath 
		WHERE flow_reference = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, flowReference)
	if err != nil {
		return nil, fmt.Errorf("failed to get executions: %w", err)
	}
	defer rows.Close()

	var executions []*models.ChatbotExecution
	for rows.Next() {
		var execution models.ChatbotExecution
		err := rows.Scan(
			&execution.ID, &execution.FlowReference, &execution.PhoneNumber, &execution.StaffID,
			&execution.ConvLast, &execution.ConvCurrent, &execution.CurrentNode,
			&execution.Variables, &execution.Status, &execution.CreatedAt, &execution.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}
		executions = append(executions, &execution)
	}

	return executions, nil
}