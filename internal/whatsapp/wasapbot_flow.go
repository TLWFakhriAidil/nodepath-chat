package whatsapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"nodepath-chat/internal/models"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// processWasapBotExamaFlow handles the WasapBot Exama flow with dynamic stage processing
func (s *Service) processWasapBotExamaFlow(phoneNumber, content, deviceID, senderName string, flow *models.ChatbotFlow) error {
	logrus.WithFields(logrus.Fields{
		"phone": phoneNumber,
		"device": deviceID,
		"flow": flow.Name,
		"message": content,
	}).Info("🎯 WASAPBOT: Starting WasapBot Exama flow processing")
	
	// Direct database access for WasapBot
	db := s.flowService.GetDB()
	if db == nil {
		logrus.Error("Database not available")
		return fmt.Errorf("database not available")
	}
	
	// Clean message for processing
	waText := strings.ToUpper(strings.TrimSpace(content))
	
	// Check for existing WasapBot record
	var idProspect int64
	var stage sql.NullString
	var currentNodeID sql.NullString
	var waitingForReply int
	
	err := db.QueryRow(`
		SELECT id_prospect, stage, current_node_id, waiting_for_reply
		FROM wasapBot_nodepath 
		WHERE prospect_num = ? AND id_device = ? 
		ORDER BY id_prospect DESC LIMIT 1
	`, phoneNumber, deviceID).Scan(&idProspect, &stage, &currentNodeID, &waitingForReply)
	
	exists := err == nil
	
	// Handle QUIT command
	if waText == "QUITEXAMA" && exists {
		db.Exec(`UPDATE wasapBot_nodepath SET stage = NULL, current_node_id = 'end' WHERE id_prospect = ?`, idProspect)
		s.SendMessageFromDevice(deviceID, phoneNumber, "Terima kasih. Sesi tamat.")
		return nil
	}
	
	// Parse flow nodes and edges
	var nodes []map[string]interface{}
	var edges []map[string]interface{}
	
	if flow.Nodes != nil {
		json.Unmarshal(*flow.Nodes, &nodes)
	}
	if flow.Edges != nil {
		json.Unmarshal(*flow.Edges, &edges)
	}
	
	// Helper function to get next nodes from edges (handles conditions)
	getNextNodes := func(currentID string) []string {
		var nextNodes []string
		for _, edge := range edges {
			if source, ok := edge["source"].(string); ok && source == currentID {
				if target, ok := edge["target"].(string); ok {
					nextNodes = append(nextNodes, target)
				}
			}
		}
		return nextNodes
	}
	
	// Helper function to get node by ID
	getNodeByID := func(nodeID string) map[string]interface{} {
		for _, node := range nodes {
			if id, ok := node["id"].(string); ok && id == nodeID {
				return node
			}
		}
		return nil
	}
	
	// Helper function to process node and extract info
	processNode := func(nodeID string) (message string, nodeType string, stageValue string, mediaURL string) {
		node := getNodeByID(nodeID)
		if node == nil {
			return
		}
		
		// Get node type
		if nt, ok := node["type"].(string); ok {
			nodeType = nt
		}
		
		// Get data from node
		if data, ok := node["data"].(map[string]interface{}); ok {
			// Get message
			if msg, ok := data["message"].(string); ok {
				message = msg
			}
			// Get stage name for stage nodes
			if nodeType == "stage" {
				if stg, ok := data["stageName"].(string); ok {
					stageValue = stg
				}
			}
			// Get media URL for image/video/audio nodes
			if nodeType == "image" {
				if img, ok := data["imageUrl"].(string); ok {
					mediaURL = img
				}
			} else if nodeType == "video" {
				if vid, ok := data["videoUrl"].(string); ok {
					mediaURL = vid
				}
			} else if nodeType == "audio" {
				if aud, ok := data["audioUrl"].(string); ok {
					mediaURL = aud
				}
			}
		}
		return
	}
	
	// Helper function to process condition node
	processConditionNode := func(nodeID string, userInput string) string {
		node := getNodeByID(nodeID)
		if node == nil {
			logrus.WithField("nodeID", nodeID).Error("Condition node not found")
			return ""
		}
		
		upperInput := strings.ToUpper(strings.TrimSpace(userInput))
		
		logrus.WithFields(logrus.Fields{
			"condition_node": nodeID,
			"user_input": userInput,
			"upper_input": upperInput,
		}).Debug("Processing condition node")
		
		// Debug: Log all edges from this node
		var availableEdges []string
		for _, edge := range edges {
			if source, ok := edge["source"].(string); ok && source == nodeID {
				sourceHandle, _ := edge["sourceHandle"].(string)
				target, _ := edge["target"].(string)
				availableEdges = append(availableEdges, fmt.Sprintf("handle:%s->target:%s", sourceHandle, target))
			}
		}
		logrus.WithField("available_edges", availableEdges).Debug("Available edges from condition node")
		
		if data, ok := node["data"].(map[string]interface{}); ok {
			if conditions, ok := data["conditions"].([]interface{}); ok {
				// Check each condition
				for _, cond := range conditions {
					if condMap, ok := cond.(map[string]interface{}); ok {
						condType, _ := condMap["type"].(string)
						condValue, _ := condMap["value"].(string)
						condLabel, _ := condMap["label"].(string)
						
						logrus.WithFields(logrus.Fields{
							"cond_type": condType,
							"cond_value": condValue,
							"cond_label": condLabel,
						}).Debug("Checking condition")
						
						// Variable to track if this condition matches
						var conditionMatched bool = false
						
						if condType == "contains" && condValue != "" {
							// Check if user input contains any of the comma-separated values
							values := strings.Split(condValue, ",")
							for _, v := range values {
								v = strings.TrimSpace(strings.ToUpper(v))
								// Check both the input and the value
								if strings.Contains(upperInput, v) || upperInput == v {
									conditionMatched = true
									logrus.WithFields(logrus.Fields{
										"matched_value": v,
										"condition_id": condMap["id"],
									}).Info("🎯 WASAPBOT: Condition matched (contains)")
									break
								}
							}
						} else if condType == "equals" && condValue != "" {
							// Check for exact match
							values := strings.Split(condValue, ",")
							for _, v := range values {
								v = strings.TrimSpace(strings.ToUpper(v))
								if upperInput == v {
									conditionMatched = true
									logrus.WithFields(logrus.Fields{
										"matched_value": v,
										"condition_id": condMap["id"],
									}).Info("🎯 WASAPBOT: Condition matched (equals)")
									break
								}
							}
						} else if condType == "not_equals" && condValue != "" {
							// Check for not equal
							values := strings.Split(condValue, ",")
							matched := false
							for _, v := range values {
								v = strings.TrimSpace(strings.ToUpper(v))
								if upperInput == v {
									matched = true
									break
								}
							}
							if !matched {
								conditionMatched = true
								logrus.WithFields(logrus.Fields{
									"condition_id": condMap["id"],
								}).Info("🎯 WASAPBOT: Condition matched (not_equals)")
							}
						} else if condType == "starts_with" && condValue != "" {
							// Check if input starts with value
							values := strings.Split(condValue, ",")
							for _, v := range values {
								v = strings.TrimSpace(strings.ToUpper(v))
								if strings.HasPrefix(upperInput, v) {
									conditionMatched = true
									logrus.WithFields(logrus.Fields{
										"matched_value": v,
										"condition_id": condMap["id"],
									}).Info("🎯 WASAPBOT: Condition matched (starts_with)")
									break
								}
							}
						} else if condType == "ends_with" && condValue != "" {
							// Check if input ends with value
							values := strings.Split(condValue, ",")
							for _, v := range values {
								v = strings.TrimSpace(strings.ToUpper(v))
								if strings.HasSuffix(upperInput, v) {
									conditionMatched = true
									logrus.WithFields(logrus.Fields{
										"matched_value": v,
										"condition_id": condMap["id"],
									}).Info("🎯 WASAPBOT: Condition matched (ends_with)")
									break
								}
							}
						}
						
						// If condition matched, find and return the edge
						if conditionMatched {
							condID, _ := condMap["id"].(string)
							
							// Try to find edge
							for _, edge := range edges {
								if source, ok := edge["source"].(string); ok && source == nodeID {
									sourceHandle, _ := edge["sourceHandle"].(string)
									target, _ := edge["target"].(string)
									
									// Check if sourceHandle matches condition ID or label  
									if sourceHandle == condID || (condLabel != "" && sourceHandle == condLabel) {
										logrus.WithFields(logrus.Fields{
											"matched_by": sourceHandle,
											"target_node": target,
										}).Info("🎯 WASAPBOT: Found target node for condition")
										return target
									}
								}
							}
							
							// Log error if no edge found for matched condition
							logrus.WithFields(logrus.Fields{
								"condition_id": condID,
								"condition_label": condLabel,
								"node_id": nodeID,
							}).Error("🎯 WASAPBOT: No edge found for matched condition")
							// Don't return empty here, continue to check default
						}
					}
				}
				
				// If no condition matched, look for default
				logrus.Info("🎯 WASAPBOT: No condition matched, looking for default")
				for _, cond := range conditions {
					if condMap, ok := cond.(map[string]interface{}); ok {
						if condType, _ := condMap["type"].(string); condType == "default" {
							condID, _ := condMap["id"].(string)
							for _, edge := range edges {
								if source, ok := edge["source"].(string); ok && source == nodeID {
									if sourceHandle, ok := edge["sourceHandle"].(string); ok && sourceHandle == condID {
										if target, ok := edge["target"].(string); ok {
											logrus.WithField("default_target", target).Info("🎯 WASAPBOT: Using default condition path")
											return target
										}
									}
								}
							}
						}
					}
				}
			}
		}
		
		// If no conditions matched and no default, just get the first edge
		logrus.Warn("🎯 WASAPBOT: No condition matched and no default found, using first edge")
		nextNodes := getNextNodes(nodeID)
		if len(nextNodes) > 0 {
			return nextNodes[0]
		}
		
		return ""
	}
	
	// Helper function to save data based on stage
	saveDataByStage := func(stageValue, userInput string) map[string]interface{} {
		updates := make(map[string]interface{})
		upperInput := strings.ToUpper(strings.TrimSpace(userInput))
		
		// Dynamic stage processing based on stage value
		switch stageValue {
		case "2", "3", "4", "5", "6":
			// Numeric stages - save user input
			updates["conv_last"] = userInput
			
			// Special handling for package selection (stage 4 or 5)
			if stageValue == "4" || stageValue == "5" {
				if strings.Contains(upperInput, "1") {
					updates["pakej"] = "1 Botol RM79"
				} else if strings.Contains(upperInput, "2") {
					updates["pakej"] = "2 Botol RM140 + Gift"
				} else if strings.Contains(upperInput, "3") {
					updates["pakej"] = "3 Botol RM190 + Gift"
				} else if strings.Contains(upperInput, "4") {
					updates["pakej"] = "4 Botol RM250 + Gift"
				}
			}
			
		case "alamat":
			updates["alamat"] = userInput
			updates["conv_last"] = userInput
			
		case "nama":
			updates["nama"] = userInput
			updates["conv_last"] = userInput
			
		case "no_fon":
			updates["no_fon"] = userInput
			updates["conv_last"] = userInput
			
		case "done":
			updates["conv_last"] = userInput
			
		case "Online Transfer", "Online Transfer (Done)":
			updates["cara_bayaran"] = "Online Transfer"
			updates["conv_last"] = "Online Transfer"
			
		case "Tarikh COD":
			updates["tarikh_gaji"] = userInput
			updates["conv_last"] = userInput
			
		case "HABIS":
			updates["status"] = "Customer"
			updates["conv_last"] = "Completed"
			
		default:
			// For any other stage, just save the input
			updates["conv_last"] = userInput
		}
		
		return updates
	}
	
	var updates map[string]interface{} = make(map[string]interface{})
	
	if !exists {
		// NEW PROSPECT - Find start node and process flow
		var startNodeID string
		for _, node := range nodes {
			if nodeType, ok := node["type"].(string); ok && nodeType == "start" {
				if id, ok := node["id"].(string); ok {
					startNodeID = id
				}
				break
			}
		}
		
		// Find first node after start
		nextNodes := getNextNodes(startNodeID)
		if len(nextNodes) == 0 {
			logrus.Error("No node after start")
			return fmt.Errorf("no node after start")
		}
		
		firstNodeID := nextNodes[0]
		
		// Create WasapBot record - don't set stage yet
		_, err = db.Exec(`
			INSERT INTO wasapBot_nodepath 
			(prospect_num, id_device, nama, current_node_id, conv_start, conv_last, 
			 date_start, date_last, niche, status, flow_reference, flow_id, waiting_for_reply)
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), 'EXAM-A', 'Prospek', ?, ?, 0)
		`, phoneNumber, deviceID, senderName, firstNodeID, content, content, flow.ID, flow.ID)
		
		if err != nil {
			logrus.WithError(err).Error("Failed to create WasapBot record")
			return err
		}
		
		// Get the inserted ID
		err = db.QueryRow(`SELECT LAST_INSERT_ID()`).Scan(&idProspect)
		if err != nil {
			logrus.WithError(err).Warn("Failed to get last insert ID")
		}
		
		// Process nodes until we hit an input node
		currentNode := firstNodeID
		for i := 0; i < 50; i++ { // Max iterations to prevent infinite loop
			msg, nodeType, stageVal, mediaURL := processNode(currentNode)
			
			logrus.WithFields(logrus.Fields{
				"node": currentNode,
				"type": nodeType,
				"stage": stageVal,
				"has_message": msg != "",
				"has_media": mediaURL != "",
			}).Debug("Processing initial node")
			
			// Update current node in database before processing
			db.Exec(`UPDATE wasapBot_nodepath SET current_node_id = ? WHERE id_prospect = ?`, currentNode, idProspect)
			
			// Handle different node types dynamically
			switch nodeType {
			case "stage":
				// Update stage in database
				if stageVal != "" {
					db.Exec(`UPDATE wasapBot_nodepath SET stage = ? WHERE id_prospect = ?`, stageVal, idProspect)
					logrus.WithField("stage", stageVal).Info("🎯 WASAPBOT: Stage updated")
				}
				
			case "message":
				// Send message immediately
				if msg != "" {
					err := s.SendMessageFromDevice(deviceID, phoneNumber, msg)
					if err != nil {
						logrus.WithError(err).Error("Failed to send message")
					}
					time.Sleep(500 * time.Millisecond) // Small delay between messages
				}
				
			case "image", "video", "audio":
				// Send media immediately
				if mediaURL != "" {
					err := s.SendMediaMessage(deviceID, phoneNumber, mediaURL)
					if err != nil {
						logrus.WithError(err).Error("Failed to send media")
					}
					time.Sleep(1 * time.Second) // Small delay after media
				}
				
			case "user_reply", "user-reply", "input", "user-input", "question":
				// Stop processing - wait for user input
				db.Exec(`UPDATE wasapBot_nodepath SET waiting_for_reply = 1 WHERE id_prospect = ?`, idProspect)
				logrus.Info("🎯 WASAPBOT: Waiting for user input")
				return nil // Exit function, waiting for user
				
			case "delay":
				// Apply actual delay
				if data, ok := getNodeByID(currentNode)["data"].(map[string]interface{}); ok {
					if delaySeconds, ok := data["delaySeconds"].(float64); ok {
						logrus.WithField("delay", delaySeconds).Info("🎯 WASAPBOT: Applying delay")
						time.Sleep(time.Duration(delaySeconds) * time.Second)
					} else if delay, ok := data["delay"].(float64); ok {
						logrus.WithField("delay", delay).Info("🎯 WASAPBOT: Applying delay")
						time.Sleep(time.Duration(delay) * time.Second)
					}
				}
				
			case "condition":
				// Condition at start - wait for user input to evaluate
				db.Exec(`UPDATE wasapBot_nodepath SET waiting_for_reply = 1 WHERE id_prospect = ?`, idProspect)
				logrus.Info("🎯 WASAPBOT: Condition node at start - waiting for user input")
				return nil // Exit function, waiting for user
				
			case "end":
				db.Exec(`UPDATE wasapBot_nodepath SET current_node_id = 'end' WHERE id_prospect = ?`, idProspect)
				logrus.Info("🎯 WASAPBOT: Flow ended")
				return nil
				
			default:
				// Unknown node type - log and continue
				logrus.WithField("node_type", nodeType).Warn("Unknown node type encountered")
			}
			
			// Get next node
			nextNodes := getNextNodes(currentNode)
			if len(nextNodes) == 0 || nextNodes[0] == "end" {
				db.Exec(`UPDATE wasapBot_nodepath SET current_node_id = 'end' WHERE id_prospect = ?`, idProspect)
				logrus.Info("🎯 WASAPBOT: No more nodes - flow ended")
				return nil
			}
			
			currentNode = nextNodes[0]
		}
		
	} else {
		// EXISTING PROSPECT - Process user input
		if !currentNodeID.Valid || currentNodeID.String == "end" {
			logrus.Info("🎯 WASAPBOT: Flow already ended")
			return nil
		}
		
		// Save data based on current stage
		if stage.Valid && stage.String != "" {
			stageUpdates := saveDataByStage(stage.String, content)
			for k, v := range stageUpdates {
				updates[k] = v
			}
		}
		
		// Get current node and check its type
		currentNodeType := ""
		currentNode := getNodeByID(currentNodeID.String)
		if currentNode != nil {
			if nt, ok := currentNode["type"].(string); ok {
				currentNodeType = nt
			}
		}
		
		// Determine next node based on current node type
		var nextNodeID string
		
		// Special handling for different waiting node types
		if currentNodeType == "condition" {
			// We're at a condition node - evaluate user input to determine next path
			nextNodeID = processConditionNode(currentNodeID.String, content)
			logrus.WithFields(logrus.Fields{
				"condition_node": currentNodeID.String,
				"user_input": content,
				"next_node": nextNodeID,
			}).Info("🎯 WASAPBOT: Evaluated condition with user input")
			
		} else if currentNodeType == "user_reply" || currentNodeType == "user-reply" || 
		          currentNodeType == "input" || currentNodeType == "user-input" || 
		          currentNodeType == "question" {
			// User has replied to an input node, move to next node
			nextNodes := getNextNodes(currentNodeID.String)
			if len(nextNodes) > 0 {
				nextNodeID = nextNodes[0]
				
				// Check if next node is a condition - if so, evaluate it immediately
				nextNode := getNodeByID(nextNodeID)
				if nextNode != nil {
					if nt, ok := nextNode["type"].(string); ok && nt == "condition" {
						// The next node is a condition, evaluate it with current user input
						logrus.WithField("condition_node", nextNodeID).Info("🎯 WASAPBOT: Next node is condition, evaluating immediately")
						nextNodeID = processConditionNode(nextNodeID, content)
						logrus.WithField("result_node", nextNodeID).Info("🎯 WASAPBOT: Condition evaluated, continuing to result")
					}
				}
			}
			logrus.WithField("next_node", nextNodeID).Info("🎯 WASAPBOT: Moving from user_reply")
			
		} else {
			// For other nodes (shouldn't happen if waiting_for_reply is set correctly)
			nextNodes := getNextNodes(currentNodeID.String)
			if len(nextNodes) > 0 {
				nextNodeID = nextNodes[0]
			}
			logrus.WithField("unexpected_node_type", currentNodeType).Warn("Unexpected node type while waiting for reply")
		}
		
		if nextNodeID == "" || nextNodeID == "end" {
			updates["current_node_id"] = "end"
			logrus.Info("🎯 WASAPBOT: Flow ended")
		} else {
			// Process nodes from next node
			currentNode := nextNodeID
			for i := 0; i < 50; i++ {
				msg, nodeType, stageVal, mediaURL := processNode(currentNode)
				
				logrus.WithFields(logrus.Fields{
					"node": currentNode,
					"type": nodeType,
					"stage": stageVal,
					"has_message": msg != "",
					"has_media": mediaURL != "",
				}).Debug("Processing node after user input")
				
				// Update current node
				updates["current_node_id"] = currentNode
				
				// Handle different node types dynamically
				switch nodeType {
				case "stage":
					// Update stage
					if stageVal != "" {
						updates["stage"] = stageVal
						logrus.WithField("stage", stageVal).Info("🎯 WASAPBOT: Stage updated from node")
					}
					
				case "message":
					// Send message immediately
					if msg != "" {
						err := s.SendMessageFromDevice(deviceID, phoneNumber, msg)
						if err != nil {
							logrus.WithError(err).Error("Failed to send message")
						}
						time.Sleep(500 * time.Millisecond) // Small delay between messages
					}
					
				case "image", "video", "audio":
					// Send media immediately
					if mediaURL != "" {
						err := s.SendMediaMessage(deviceID, phoneNumber, mediaURL)
						if err != nil {
							logrus.WithError(err).Error("Failed to send media")
						}
						time.Sleep(1 * time.Second) // Delay after media
					}
					
				case "user_reply", "user-reply", "input", "user-input", "question":
					// Stop and wait for input
					updates["waiting_for_reply"] = 1
					logrus.Info("🎯 WASAPBOT: Waiting for user input")
					break
					
				case "condition":
					// We've moved to a condition node - evaluate it immediately with current user input
					nextCondNode := processConditionNode(currentNode, content)
					logrus.WithFields(logrus.Fields{
						"condition_node": currentNode,
						"user_input": content,
						"next_after_condition": nextCondNode,
					}).Info("🎯 WASAPBOT: Evaluating condition after user_reply")
					
					if nextCondNode != "" && nextCondNode != "end" {
						// Continue processing from the result of the condition
						currentNode = nextCondNode
						continue // Continue the loop to process the next node
					} else if nextCondNode == "end" {
						updates["current_node_id"] = "end"
						logrus.Info("🎯 WASAPBOT: Flow ended after condition")
						break
					} else {
						// No valid path from condition - this shouldn't happen
						logrus.Error("🎯 WASAPBOT: No valid path from condition node")
						updates["current_node_id"] = currentNode
						updates["waiting_for_reply"] = 1
						break
					}
					
				case "delay":
					// Apply actual delay
					if data, ok := getNodeByID(currentNode)["data"].(map[string]interface{}); ok {
						if delaySeconds, ok := data["delaySeconds"].(float64); ok {
							logrus.WithField("delay", delaySeconds).Info("🎯 WASAPBOT: Applying delay")
							time.Sleep(time.Duration(delaySeconds) * time.Second)
						} else if delay, ok := data["delay"].(float64); ok {
							logrus.WithField("delay", delay).Info("🎯 WASAPBOT: Applying delay")
							time.Sleep(time.Duration(delay) * time.Second)
						}
					}
					
				case "end":
					updates["current_node_id"] = "end"
					logrus.Info("🎯 WASAPBOT: Flow ended")
					break
					
				default:
					// Unknown node type - log and continue
					logrus.WithField("node_type", nodeType).Warn("Unknown node type encountered")
				}
				
				// If we need user input, stop processing
				if nodeType == "user_reply" || nodeType == "user-reply" || nodeType == "input" || 
				   nodeType == "user-input" || nodeType == "question" || nodeType == "condition" || 
				   nodeType == "end" {
					break
				}
				
				// Get next node
				nextNodes := getNextNodes(currentNode)
				if len(nextNodes) == 0 {
					updates["current_node_id"] = "end"
					break
				}
				
				currentNode = nextNodes[0]
			}
		}
	}
	
	// Update WasapBot database
	if len(updates) > 0 && exists {
		var setClauses []string
		var args []interface{}
		
		for field, value := range updates {
			setClauses = append(setClauses, field + " = ?")
			args = append(args, value)
		}
		
		setClauses = append(setClauses, "date_last = NOW()")
		args = append(args, idProspect)
		
		query := fmt.Sprintf("UPDATE wasapBot_nodepath SET %s WHERE id_prospect = ?", strings.Join(setClauses, ", "))
		_, err = db.Exec(query, args...)
		
		if err != nil {
			logrus.WithError(err).Error("Failed to update WasapBot record")
		} else {
			logrus.WithField("updates", updates).Info("🎯 WASAPBOT: Updated database")
		}
	}
	
	logrus.WithFields(logrus.Fields{
		"stage": stage,
		"updates": updates,
	}).Info("🎯 WASAPBOT: Flow processing completed")
	
	return nil
}
