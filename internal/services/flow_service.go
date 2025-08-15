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
		return fmt.Errorf("database not available")
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
		"name":    flow.Name,
	}).Info("Flow created successfully")

	return nil
}

// GetFlow retrieves a flow by ID
func (s *FlowService) GetFlow(flowID string) (*models.ChatbotFlow, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
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

// GetAllFlows retrieves all flows
func (s *FlowService) GetAllFlows() ([]*models.ChatbotFlow, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
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

// UpdateFlow updates an existing flow
func (s *FlowService) UpdateFlow(flow *models.ChatbotFlow) error {
	flow.UpdatedAt = time.Now()

	query := `
		UPDATE chatbot_flows_nodepath 
		SET name = ?, niche = ?, id_device = ?, nodes = ?, edges = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.Exec(query,
		flow.Name, flow.Niche, flow.IdDevice, flow.Nodes, flow.Edges, 
		flow.UpdatedAt, flow.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"flow_reference": flow.ID,
		"name":    flow.Name,
	}).Info("Flow updated successfully")

	return nil
}

// DeleteFlow deletes a flow
func (s *FlowService) DeleteFlow(flowID string) error {
	query := `DELETE FROM chatbot_flows_nodepath WHERE id = ?`
	_, err := s.db.Exec(query, flowID)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}

	logrus.WithField("flow_reference", flowID).Info("Flow deleted successfully")
	return nil
}

// GetFlowNodes parses and returns the nodes from a flow
func (s *FlowService) GetFlowNodes(flow *models.ChatbotFlow) ([]*models.FlowNode, error) {
	if flow.Nodes == nil {
		return []*models.FlowNode{}, nil
	}

	var nodes []*models.FlowNode
	err := json.Unmarshal(*flow.Nodes, &nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flow nodes: %w", err)
	}

	return nodes, nil
}

// GetFlowEdges parses and returns the edges from a flow
func (s *FlowService) GetFlowEdges(flow *models.ChatbotFlow) ([]*models.FlowEdge, error) {
	if flow.Edges == nil {
		return []*models.FlowEdge{}, nil
	}

	var edges []*models.FlowEdge
	err := json.Unmarshal(*flow.Edges, &edges)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flow edges: %w", err)
	}

	return edges, nil
}

// FindNodeByID finds a node by its ID in the flow
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

	return nil, fmt.Errorf("node not found: %s", nodeID)
}

// GetNextNode finds the next node in the flow based on current node and edges
func (s *FlowService) GetNextNode(flow *models.ChatbotFlow, currentNodeID string) (*models.FlowNode, error) {
	edges, err := s.GetFlowEdges(flow)
	if err != nil {
		return nil, err
	}

	// Find the edge that starts from the current node
	var nextNodeID string
	for _, edge := range edges {
		if edge.Source == currentNodeID {
			nextNodeID = edge.Target
			break
		}
	}

	if nextNodeID == "" {
		return nil, nil // No next node (end of flow)
	}

	return s.FindNodeByID(flow, nextNodeID)
}

// GetStartNode finds the starting node of the flow
func (s *FlowService) GetStartNode(flow *models.ChatbotFlow) (*models.FlowNode, error) {
	nodes, err := s.GetFlowNodes(flow)
	if err != nil {
		return nil, err
	}

	edges, err := s.GetFlowEdges(flow)
	if err != nil {
		return nil, err
	}

	// Find nodes that are not targets of any edge (start nodes)
	targetNodes := make(map[string]bool)
	for _, edge := range edges {
		targetNodes[edge.Target] = true
	}

	for _, node := range nodes {
		if !targetNodes[node.ID] {
			return node, nil // This is a start node
		}
	}

	// If no start node found, return the first node
	if len(nodes) > 0 {
		return nodes[0], nil
	}

	return nil, fmt.Errorf("no start node found in flow")
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