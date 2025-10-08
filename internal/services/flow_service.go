package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nodepath-chat/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// FlowService handles chatbot flow operations
type FlowService struct {
	db    *sql.DB
	redis *redis.Client
}

// GetDB returns the database connection
func (s *FlowService) GetDB() *sql.DB {
	return s.db
}

// NewFlowService creates a new flow service
func NewFlowService(db *sql.DB, redis *redis.Client) *FlowService {
	return &FlowService{
		db:    db,
		redis: redis,
	}
}

// CreateFlow creates a new chatbot flow
func (s *FlowService) CreateFlow(flow *models.ChatbotFlow) error {
	if s.db == nil {
		logrus.Warn("Database not available, flow creation skipped (fallback mode)")
		return nil // Return success in fallback mode
	}

	if flow.ID == "" {
		flow.ID = uuid.New().String()
	}

	flow.CreatedAt = time.Now()
	flow.UpdatedAt = time.Now()

	query := `
		INSERT INTO chatbot_flows_nodepath 
		(id, name, niche, id_device,
		 nodes, edges, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		flow.ID, flow.Name, flow.Niche, flow.IdDevice, flow.Nodes, flow.Edges,
		flow.CreatedAt, flow.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create flow: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"flow_reference": flow.ID,
		"name":           flow.Name,
	}).Info("Flow created successfully")

	return nil
}

// GetFlow retrieves a flow by ID
func (s *FlowService) GetFlow(flowID string) (*models.ChatbotFlow, error) {
	if s.db == nil {
		logrus.Warn("Database not available, returning nil flow (fallback mode)")
		return nil, nil // Return nil flow in fallback mode
	}

	query := `
		SELECT id, name, niche, id_device,
		       nodes, edges, created_at, updated_at
		FROM chatbot_flows_nodepath 
		WHERE id = ?
		LIMIT 1
	`

	var flow models.ChatbotFlow
	err := s.db.QueryRow(query, flowID).Scan(
		&flow.ID, &flow.Name, &flow.Niche, &flow.IdDevice, &flow.Nodes, &flow.Edges,
		&flow.CreatedAt, &flow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	return &flow, nil
}

// DetermineTableByFlowName determines which table to use based on flow name
func (s *FlowService) DetermineTableByFlowName(flowName string) string {
	// Check if flow name is "WasapBot Exama"
	if flowName == "WasapBot Exama" {
		logrus.WithField("flow_name", flowName).Info("📊 TABLE SELECTION: Using wasapBot_nodepath for WasapBot Exama flow")
		return "wasapBot_nodepath"
	}
	// Default to ai_whatsapp_nodepath for "Chatbot AI" or any other name
	logrus.WithField("flow_name", flowName).Info("📊 TABLE SELECTION: Using ai_whatsapp_nodepath for Chatbot AI flow")
	return "ai_whatsapp_nodepath"
}

// GetFlowAndDetermineTable retrieves a flow and determines which table to use for processing
func (s *FlowService) GetFlowAndDetermineTable(flowID string) (*models.ChatbotFlow, string, error) {
	flow, err := s.GetFlow(flowID)
	if err != nil {
		return nil, "", err
	}
	if flow == nil {
		return nil, "", fmt.Errorf("flow not found")
	}

	// Determine which table to use based on flow name
	tableName := s.DetermineTableByFlowName(flow.Name)

	logrus.WithFields(logrus.Fields{
		"flow_id":    flowID,
		"flow_name":  flow.Name,
		"table_name": tableName,
	}).Info("Determined table for flow processing")

	return flow, tableName, nil
}

// GetAllFlows retrieves all flows
func (s *FlowService) GetAllFlows() ([]*models.ChatbotFlow, error) {
	if s.db == nil {
		logrus.Warn("Database not available, returning empty flows list (fallback mode)")
		return []*models.ChatbotFlow{}, nil // Return empty list in fallback mode
	}

	query := `
		SELECT id, name, niche, id_device,
		       nodes, edges, created_at, updated_at
		FROM chatbot_flows_nodepath 
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows: %w", err)
	}
	defer rows.Close()

	var flows []*models.ChatbotFlow
	for rows.Next() {
		var flow models.ChatbotFlow
		err := rows.Scan(
			&flow.ID, &flow.Name, &flow.Niche, &flow.IdDevice, &flow.Nodes, &flow.Edges,
			&flow.CreatedAt, &flow.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}
		flows = append(flows, &flow)
	}

	return flows, nil
}

// GetFlowsByUserDevices retrieves flows filtered by user's device IDs from device_setting_nodepath
// GetFlowsByUserDevicesString gets flows by user devices using string UUID user_id
func (s *FlowService) GetFlowsByUserDevicesString(userID string) ([]*models.ChatbotFlow, error) {
	if s.db == nil {
		logrus.Warn("Database not available, returning empty flows list for user devices (fallback mode)")
		return []*models.ChatbotFlow{}, nil // Return empty list in fallback mode
	}

	// First get all device IDs for this user
	deviceQuery := `SELECT id_device FROM device_setting_nodepath WHERE user_id = ?`
	deviceRows, err := s.db.Query(deviceQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user devices: %w", err)
	}
	defer deviceRows.Close()

	var deviceIDs []string
	for deviceRows.Next() {
		var deviceID string
		err := deviceRows.Scan(&deviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device ID: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	// If user has no devices, return empty list
	if len(deviceIDs) == 0 {
		return []*models.ChatbotFlow{}, nil
	}

	// Build query with IN clause for device filtering
	placeholders := strings.Repeat("?,", len(deviceIDs))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf(`
		SELECT id, name, niche, id_device,
		       nodes, edges, created_at, updated_at
		FROM chatbot_flows_nodepath 
		WHERE id_device IN (%s)
		ORDER BY created_at DESC
	`, placeholders)

	// Convert deviceIDs to interface{} slice for query
	args := make([]interface{}, len(deviceIDs))
	for i, deviceID := range deviceIDs {
		args[i] = deviceID
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows for user devices: %w", err)
	}
	defer rows.Close()

	var flows []*models.ChatbotFlow
	for rows.Next() {
		var flow models.ChatbotFlow
		err := rows.Scan(
			&flow.ID, &flow.Name, &flow.Niche, &flow.IdDevice, &flow.Nodes, &flow.Edges,
			&flow.CreatedAt, &flow.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}
		flows = append(flows, &flow)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating flow rows: %w", err)
	}

	return flows, nil
}

// GetFlowsByUserDevices gets flows by user devices using int user_id (deprecated, use GetFlowsByUserDevicesString)
func (s *FlowService) GetFlowsByUserDevices(userID int) ([]*models.ChatbotFlow, error) {
	if s.db == nil {
		logrus.Warn("Database not available, returning empty flows list for user devices (fallback mode)")
		return []*models.ChatbotFlow{}, nil // Return empty list in fallback mode
	}

	// First get all device IDs for this user
	deviceQuery := `SELECT id_device FROM device_setting_nodepath WHERE user_id = ?`
	deviceRows, err := s.db.Query(deviceQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user devices: %w", err)
	}
	defer deviceRows.Close()

	var deviceIDs []string
	for deviceRows.Next() {
		var deviceID string
		err := deviceRows.Scan(&deviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device ID: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	// If user has no devices, return empty list
	if len(deviceIDs) == 0 {
		return []*models.ChatbotFlow{}, nil
	}

	// Build query with IN clause for device filtering
	placeholders := strings.Repeat("?,", len(deviceIDs))
	placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma

	query := fmt.Sprintf(`
		SELECT id, name, niche, id_device,
		       nodes, edges, created_at, updated_at
		FROM chatbot_flows_nodepath 
		WHERE id_device IN (%s)
		ORDER BY created_at DESC
	`, placeholders)

	// Convert deviceIDs to interface{} slice for query
	args := make([]interface{}, len(deviceIDs))
	for i, deviceID := range deviceIDs {
		args[i] = deviceID
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows for user devices: %w", err)
	}
	defer rows.Close()

	var flows []*models.ChatbotFlow
	for rows.Next() {
		var flow models.ChatbotFlow
		err := rows.Scan(
			&flow.ID, &flow.Name, &flow.Niche, &flow.IdDevice, &flow.Nodes, &flow.Edges,
			&flow.CreatedAt, &flow.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}
		flows = append(flows, &flow)
	}

	return flows, nil
}

// GetFlowsByDevice retrieves flows by device ID
func (s *FlowService) GetFlowsByDevice(idDevice string) ([]*models.ChatbotFlow, error) {
	if s.db == nil {
		logrus.Warn("Database not available, returning empty flows list for device (fallback mode)")
		return []*models.ChatbotFlow{}, nil // Return empty list in fallback mode
	}

	query := `
		SELECT id, name, niche, id_device,
		       nodes, edges, created_at, updated_at
		FROM chatbot_flows_nodepath 
		WHERE id_device = ?
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, idDevice)
	if err != nil {
		return nil, fmt.Errorf("failed to get flows by device: %w", err)
	}
	defer rows.Close()

	var flows []*models.ChatbotFlow
	for rows.Next() {
		var flow models.ChatbotFlow
		err := rows.Scan(
			&flow.ID, &flow.Name, &flow.Niche, &flow.IdDevice, &flow.Nodes, &flow.Edges,
			&flow.CreatedAt, &flow.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan flow: %w", err)
		}
		flows = append(flows, &flow)
	}

	return flows, nil
}

// GetDefaultFlowForDevice retrieves the first/default flow for a device
func (s *FlowService) GetDefaultFlowForDevice(idDevice string) (*models.ChatbotFlow, error) {
	flows, err := s.GetFlowsByDevice(idDevice)
	if err != nil {
		return nil, err
	}

	if len(flows) == 0 {
		return nil, nil
	}

	return flows[0], nil // Return the first flow as default
}

// GetStartNode extracts the start node from a flow's nodes JSON
func (s *FlowService) GetStartNode(flow *models.ChatbotFlow) (*models.FlowNode, error) {
	if flow.Nodes == nil || len(*flow.Nodes) == 0 {
		return nil, fmt.Errorf("flow has no nodes")
	}

	var nodes []*models.FlowNode
	if err := json.Unmarshal(*flow.Nodes, &nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	// Find the start node
	for _, node := range nodes {
		if node.Type == "start" {
			return node, nil
		}
	}

	return nil, fmt.Errorf("no start node found in flow")
}

// GetFlowNodes extracts all nodes from a flow's nodes JSON
func (s *FlowService) GetFlowNodes(flow *models.ChatbotFlow) ([]*models.FlowNode, error) {
	if flow.Nodes == nil || len(*flow.Nodes) == 0 {
		return nil, fmt.Errorf("flow has no nodes")
	}

	var nodes []*models.FlowNode
	if err := json.Unmarshal(*flow.Nodes, &nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nodes: %w", err)
	}

	return nodes, nil
}

// GetFlowEdges extracts edges from a flow's edges JSON
func (s *FlowService) GetFlowEdges(flow *models.ChatbotFlow) ([]*models.FlowEdge, error) {
	if flow.Edges == nil || len(*flow.Edges) == 0 {
		return []*models.FlowEdge{}, nil // Return empty array if no edges
	}

	var edges []*models.FlowEdge
	if err := json.Unmarshal(*flow.Edges, &edges); err != nil {
		return nil, fmt.Errorf("failed to unmarshal edges: %w", err)
	}

	return edges, nil
}

// FindNodeByID finds a node by its ID in the flow's nodes
func (s *FlowService) FindNodeByID(flow *models.ChatbotFlow, nodeID string) (*models.FlowNode, error) {
	nodes, err := s.GetFlowNodes(flow)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes {
		if node.ID == nodeID {
			return node, nil
		}
	}

	return nil, fmt.Errorf("node with ID %s not found", nodeID)
}

// UpdateFlow updates an existing flow
func (s *FlowService) UpdateFlow(flow *models.ChatbotFlow) error {
	if s.db == nil {
		logrus.Warn("Database not available, flow update skipped (fallback mode)")
		return nil // Return success in fallback mode
	}

	flow.UpdatedAt = time.Now()

	query := `
		UPDATE chatbot_flows_nodepath 
		SET name = ?, niche = ?, id_device = ?,
		    nodes = ?, edges = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		flow.Name, flow.Niche, flow.IdDevice, flow.Nodes, flow.Edges,
		flow.UpdatedAt, flow.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	return nil
}

// DeleteFlow deletes a flow by ID
func (s *FlowService) DeleteFlow(flowID string) error {
	if s.db == nil {
		logrus.Warn("Database not available, flow deletion skipped (fallback mode)")
		return nil // Return success in fallback mode
	}

	query := `DELETE FROM chatbot_flows_nodepath WHERE id = ?`
	_, err := s.db.Exec(query, flowID)

	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}

	return nil
}

// GetNextNode finds the next node in the flow based on the current node
func (s *FlowService) GetNextNode(flow *models.ChatbotFlow, currentNodeID string) (*models.FlowNode, error) {
	edges, err := s.GetFlowEdges(flow)
	if err != nil {
		return nil, err
	}

	// Find edge from current node
	var nextNodeID string
	for _, edge := range edges {
		if edge.Source == currentNodeID {
			nextNodeID = edge.Target
			break
		}
	}

	if nextNodeID == "" {
		return nil, fmt.Errorf("no next node found for node %s", currentNodeID)
	}

	return s.FindNodeByID(flow, nextNodeID)
}

// EvaluateConditionNode evaluates a condition node and returns the appropriate next node based on user input
func (s *FlowService) EvaluateConditionNode(flow *models.ChatbotFlow, conditionNodeID string, userInput string) (*models.FlowNode, error) {
	// Use the fixed version from condition_evaluation_fix.go
	return s.EvaluateConditionNodeFixed(flow, conditionNodeID, userInput)
}

// ReplaceVariables replaces variables in text with actual values
func (s *FlowService) ReplaceVariables(text string, variables map[string]interface{}) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		if valueStr, ok := value.(string); ok {
			result = strings.ReplaceAll(result, placeholder, valueStr)
		} else {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}
	return result
}
