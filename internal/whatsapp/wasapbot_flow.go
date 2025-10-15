package whatsapp

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"nodepath-chat/internal/models"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// processWasapBotExamaFlow handles the WasapBot Exama flow with dynamic stage processing
func (s *Service) processWasapBotExamaFlow(phoneNumber, content, deviceID, senderName string, flow *models.ChatbotFlow) error {
	flowName := flow.Name
	logrus.WithFields(logrus.Fields{
		"phone":   phoneNumber,
		"device":  deviceID,
		"flow":    flowName,
		"message": content,
	}).Info("🎯 WASAPBOT: Starting WasapBot Exama flow processing")

	// CRITICAL: Phone number validation - must be <= 13 digits OR start with 601
	if len(phoneNumber) > 13 && !strings.HasPrefix(phoneNumber, "601") {
		logrus.WithFields(logrus.Fields{
			"phone":  phoneNumber,
			"reason": "Invalid phone number format (>13 digits and not 601 prefix)",
		}).Warn("🚫 WASAPBOT: Phone validation failed, terminating without saving")
		return nil // Terminate immediately without any database operations
	}

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

	// Helper function to replace template messages for WasapBot Exama flow
	replaceTemplateMessage := func(flowName string, originalMessage string) string {
		// Only apply to WasapBot Exama flow
		if flowName != "WasapBot Exama" {
			return originalMessage
		}

		// Get current prospect data from database
		var nama, alamat, noFon, pakej, caraBayaran, tarikhGaji sql.NullString
		err := db.QueryRow(`
			SELECT nama, alamat, no_fon, pakej, cara_bayaran, tarikh_gaji 
			FROM wasapBot_nodepath 
			WHERE id_prospect = ?`, idProspect).Scan(&nama, &alamat, &noFon, &pakej, &caraBayaran, &tarikhGaji)

		if err != nil {
			logrus.WithError(err).Warn("Failed to get prospect data for template replacement")
			return originalMessage // Return original if can't get data
		}

		// Check for exact template matches and replace
		switch strings.TrimSpace(originalMessage) {
		case "SEND DETAIL PEMBELI":
			if nama.Valid && alamat.Valid {
				replaced := fmt.Sprintf("Pengesahan Detail:\nNAMA : %s\nALAMAT : %s\nNO FON : %s",
					nama.String, alamat.String, content) // Use current user input as phone number
				logrus.WithFields(logrus.Fields{
					"template": "SEND DETAIL PEMBELI",
					"flow":     flowName,
				}).Info("📝 WASAPBOT: Replaced template with dynamic data")
				return replaced
			}

		case "DETAIL PEMBELI COD":
			if nama.Valid && alamat.Valid && noFon.Valid && pakej.Valid {
				replaced := fmt.Sprintf("Baik, ini ringkasan tempahan Cik yaa...\nNAMA : %s\nALAMAT : %s\nNO FONE : %s\nPAKEJ : %s\n*COD @ POSTAGE PERCUMA*\nCARA BAYARAN : COD",
					nama.String, alamat.String, noFon.String, pakej.String)
				logrus.WithFields(logrus.Fields{
					"template": "DETAIL PEMBELI COD",
					"flow":     flowName,
				}).Info("📝 WASAPBOT: Replaced template with dynamic data")
				return replaced
			}

		case "DETAIL PEMBELI GAJI":
			if nama.Valid && alamat.Valid && noFon.Valid && pakej.Valid && caraBayaran.Valid {
				replaced := fmt.Sprintf("Baik, ini ringkasan tempahan Cik yaa...\nNAMA : %s\nALAMAT : %s\nNO FONE : %s\nPAKEJ : %s\n*COD @ POSTAGE PERCUMA*\nCARA BAYARAN : %s\nTARIKH GAJI : %s",
					nama.String, alamat.String, noFon.String, pakej.String, caraBayaran.String, content) // Use current user input as date
				logrus.WithFields(logrus.Fields{
					"template": "DETAIL PEMBELI GAJI",
					"flow":     flowName,
				}).Info("📝 WASAPBOT: Replaced template with dynamic data")
				return replaced
			}

		case "DETAIL PEMBELI CASH":
			if nama.Valid && alamat.Valid && noFon.Valid && pakej.Valid {
				replaced := fmt.Sprintf("Baik, ini ringkasan tempahan Cik yaa...\nNAMA : %s\nALAMAT : %s\nNO FONE : %s\nPAKEJ : %s\n*COD @ POSTAGE PERCUMA*\nCARA BAYARAN : Online Transfer",
					nama.String, alamat.String, noFon.String, pakej.String)
				logrus.WithFields(logrus.Fields{
					"template": "DETAIL PEMBELI CASH",
					"flow":     flowName,
				}).Info("📝 WASAPBOT: Replaced template with dynamic data")
				return replaced
			}
		}

		// If no exact match, return original message
		return originalMessage
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
			"user_input":     userInput,
			"upper_input":    upperInput,
		}).Debug("Processing condition node")

		// Debug: Log all edges from this node
		var availableEdges []string
		for _, edge := range edges {
			if source, ok := edge["source"].(string); ok && source == nodeID {
				sourceHandle, _ := edge["sourceHandle"].(string)
				target, _ := edge["target"].(string)
				// More detailed edge info for debugging
				availableEdges = append(availableEdges, fmt.Sprintf("sourceHandle:'%s'->target:%s", sourceHandle, target))
			}
		}
		logrus.WithField("available_edges", availableEdges).Info("📊 WASAPBOT: Available edges from condition node")

		if data, ok := node["data"].(map[string]interface{}); ok {
			// First check if user input is a number that matches a condition label
			// This is different from edge index - it should match the actual label like "1", "2", "3", etc.
			if userNum, err := strconv.Atoi(strings.TrimSpace(userInput)); err == nil {
				// User entered a number - look for matching sourceHandle (edge label)
				userNumStr := strconv.Itoa(userNum)
				for _, edge := range edges {
					if source, ok := edge["source"].(string); ok && source == nodeID {
						sourceHandle, _ := edge["sourceHandle"].(string)
						if sourceHandle == userNumStr {
							target, _ := edge["target"].(string)
							logrus.WithFields(logrus.Fields{
								"user_input":  userInput,
								"edge_label":  sourceHandle,
								"target_node": target,
							}).Info("✅ WASAPBOT: Direct edge selection by label match")
							return target
						}
					}
				}
			}

			if conditions, ok := data["conditions"].([]interface{}); ok {
				// Check each condition
				for i, cond := range conditions {
					if condMap, ok := cond.(map[string]interface{}); ok {
						condType, _ := condMap["type"].(string)
						condValue, _ := condMap["value"].(string)
						condLabel, _ := condMap["label"].(string)

						logrus.WithFields(logrus.Fields{
							"cond_type":  condType,
							"cond_value": condValue,
							"cond_label": condLabel,
							"cond_index": i,
							"user_input": userInput,
						}).Debug("📋 WASAPBOT: Checking condition")

						// Variable to track if this condition matches
						var conditionMatched bool = false

						// PRIORITY 1: Check if user input exactly matches the condition label
						// This handles numbered options like "1", "2", "3", "4"
						if condLabel != "" && strings.TrimSpace(userInput) == condLabel {
							conditionMatched = true
							logrus.WithFields(logrus.Fields{
								"matched_label":   condLabel,
								"user_input":      userInput,
								"condition_index": i,
							}).Info("✅ WASAPBOT: Exact label match")
						} else if condType == "contains" && condValue != "" {
							// Check if user input contains any of the comma-separated values
							values := strings.Split(condValue, ",")
							for _, v := range values {
								v = strings.TrimSpace(strings.ToUpper(v))
								// Check both the input and the value
								if strings.Contains(upperInput, v) || upperInput == v {
									conditionMatched = true
									logrus.WithFields(logrus.Fields{
										"matched_value":   v,
										"condition_id":    condMap["id"],
										"condition_label": condLabel,
										"condition_index": i,
									}).Info("✅ WASAPBOT: Condition matched (contains)")
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
										"matched_value":   v,
										"condition_id":    condMap["id"],
										"condition_label": condLabel,
										"condition_index": i,
									}).Info("✅ WASAPBOT: Condition matched (equals)")
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
										"condition_id":  condMap["id"],
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
										"condition_id":  condMap["id"],
									}).Info("🎯 WASAPBOT: Condition matched (ends_with)")
									break
								}
							}
						}

						// If condition matched, find and return the edge
						if conditionMatched {
							condID, _ := condMap["id"].(string)
							condIndex := i                            // Current condition index
							condLabel := strings.TrimSpace(condLabel) // Ensure label is trimmed
							condValue := strings.TrimSpace(condValue) // Ensure value is trimmed

							// Debug: Log what we're looking for and all available edges
							logrus.WithFields(logrus.Fields{
								"condition_id":    condID,
								"condition_label": condLabel,
								"condition_value": condValue,
								"condition_index": condIndex,
								"available_edges": availableEdges,
								"looking_for":     fmt.Sprintf("ID=%s OR Label=%s OR Value=%s", condID, condLabel, condValue),
							}).Info("🔍 WASAPBOT: Looking for matching edge for condition")

							// DYNAMIC MATCHING STRATEGY
							// We need to dynamically detect what sourceHandle pattern is being used

							foundEdge := false
							var targetNode string

							// First, let's collect all edges from this node to understand the pattern
							var edgesFromNode []map[string]string
							for _, edge := range edges {
								if source, ok := edge["source"].(string); ok && source == nodeID {
									sourceHandle, _ := edge["sourceHandle"].(string)
									target, _ := edge["target"].(string)
									edgesFromNode = append(edgesFromNode, map[string]string{
										"sourceHandle": sourceHandle,
										"target":       target,
									})
								}
							}

							// STRATEGY 1: Try direct matching with condition-X pattern (common in flow builders)
							expectedHandle := fmt.Sprintf("condition-%d", condIndex)
							for _, edgeInfo := range edgesFromNode {
								if edgeInfo["sourceHandle"] == expectedHandle {
									targetNode = edgeInfo["target"]
									foundEdge = true
									logrus.WithFields(logrus.Fields{
										"matched_handle": expectedHandle,
										"target_node":    targetNode,
										"strategy":       "condition-index-pattern",
									}).Info("✅ WASAPBOT: Found edge by condition-X pattern")
									return targetNode
								}
							}

							// STRATEGY 2: Try matching by position (if edges are in same order as conditions)
							// This works when sourceHandles don't follow a predictable pattern
							if !foundEdge && condIndex < len(edgesFromNode) {
								targetNode = edgesFromNode[condIndex]["target"]
								foundEdge = true
								logrus.WithFields(logrus.Fields{
									"condition_index": condIndex,
									"sourceHandle":    edgesFromNode[condIndex]["sourceHandle"],
									"target_node":     targetNode,
									"strategy":        "position-based",
								}).Info("✅ WASAPBOT: Found edge by position matching")
								return targetNode
							}

							// STRATEGY 3: Try all possible label/value/ID matches
							var possibleHandles []string

							// Add all variations of label
							if condLabel != "" {
								possibleHandles = append(possibleHandles, condLabel)
								possibleHandles = append(possibleHandles, strings.ToUpper(condLabel))
								possibleHandles = append(possibleHandles, strings.ToLower(condLabel))
							}

							// Add ID
							if condID != "" {
								possibleHandles = append(possibleHandles, condID)
							}

							// Add all variations of value
							if condValue != "" {
								possibleHandles = append(possibleHandles, condValue)
								possibleHandles = append(possibleHandles, strings.ToUpper(condValue))
								possibleHandles = append(possibleHandles, strings.ToLower(condValue))
							}

							// Add index as string
							possibleHandles = append(possibleHandles, strconv.Itoa(condIndex))

							// Try each possible handle
							for _, possibleHandle := range possibleHandles {
								for _, edgeInfo := range edgesFromNode {
									if strings.EqualFold(edgeInfo["sourceHandle"], possibleHandle) {
										targetNode = edgeInfo["target"]
										foundEdge = true
										logrus.WithFields(logrus.Fields{
											"matched_handle": possibleHandle,
											"sourceHandle":   edgeInfo["sourceHandle"],
											"target_node":    targetNode,
											"strategy":       "flexible-match",
										}).Info("✅ WASAPBOT: Found edge by flexible matching")
										return targetNode
									}
								}
							}

							// STRATEGY 4: Pattern detection - try to understand the pattern
							// Look for patterns like "condition-X", "cond_X", "X", etc.
							for _, edgeInfo := range edgesFromNode {
								handle := edgeInfo["sourceHandle"]
								// Check if handle contains the index number
								if strings.Contains(handle, strconv.Itoa(condIndex)) {
									targetNode = edgeInfo["target"]
									foundEdge = true
									logrus.WithFields(logrus.Fields{
										"sourceHandle":    handle,
										"condition_index": condIndex,
										"target_node":     targetNode,
										"strategy":        "pattern-detection",
									}).Info("✅ WASAPBOT: Found edge by pattern detection")
									return targetNode
								}
							}

							if foundEdge {
								return targetNode
							}

							// Log error if no edge found
							logrus.WithFields(logrus.Fields{
								"condition_id":    condID,
								"condition_label": condLabel,
								"condition_value": condValue,
								"condition_index": condIndex,
								"available_edges": availableEdges,
								"edges_from_node": edgesFromNode,
							}).Error("❌ WASAPBOT: No edge found for matched condition despite trying all strategies")
						}
					}
				}

				// If no condition matched, look for default
				logrus.Info("🔍 WASAPBOT: No condition matched, looking for default condition")
				for i, cond := range conditions {
					if condMap, ok := cond.(map[string]interface{}); ok {
						if condType, _ := condMap["type"].(string); condType == "default" {
							condID, _ := condMap["id"].(string)
							condLabel, _ := condMap["label"].(string)

							logrus.WithFields(logrus.Fields{
								"default_id":    condID,
								"default_label": condLabel,
								"default_index": i,
							}).Debug("🔍 WASAPBOT: Found default condition, looking for edge")

							// Try multiple ways to find the default edge
							// 1. Try matching by label "default" or "Default" or "DEFAULT"
							for _, edge := range edges {
								if source, ok := edge["source"].(string); ok && source == nodeID {
									sourceHandle, _ := edge["sourceHandle"].(string)
									if strings.EqualFold(sourceHandle, "default") {
										target, _ := edge["target"].(string)
										logrus.WithField("default_target", target).Info("✅ WASAPBOT: Using default condition path (by 'default' label)")
										return target
									}
								}
							}

							// 2. Try matching by condition ID
							for _, edge := range edges {
								if source, ok := edge["source"].(string); ok && source == nodeID {
									sourceHandle, _ := edge["sourceHandle"].(string)
									if sourceHandle == condID {
										target, _ := edge["target"].(string)
										logrus.WithField("default_target", target).Info("✅ WASAPBOT: Using default condition path (by ID)")
										return target
									}
								}
							}

							// 3. Try matching by condition label if exists
							if condLabel != "" {
								for _, edge := range edges {
									if source, ok := edge["source"].(string); ok && source == nodeID {
										sourceHandle, _ := edge["sourceHandle"].(string)
										if strings.EqualFold(sourceHandle, condLabel) {
											target, _ := edge["target"].(string)
											logrus.WithField("default_target", target).Info("✅ WASAPBOT: Using default condition path (by label)")
											return target
										}
									}
								}
							}

							// 4. Try position-based (default is often the last edge)
							edgeCount := 0
							var lastEdgeTarget string
							for _, edge := range edges {
								if source, ok := edge["source"].(string); ok && source == nodeID {
									target, _ := edge["target"].(string)
									lastEdgeTarget = target
									if edgeCount == i {
										logrus.WithField("default_target", target).Info("✅ WASAPBOT: Using default condition path (by position)")
										return target
									}
									edgeCount++
								}
							}

							// 5. If nothing else worked, use the last edge as default
							if lastEdgeTarget != "" {
								logrus.WithField("default_target", lastEdgeTarget).Warn("⚠️ WASAPBOT: Using last edge as default fallback")
								return lastEdgeTarget
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

	// Helper function to get last user input from database
	getLastUserInput := func(prospectID int64) string {
		var lastInput sql.NullString
		err := db.QueryRow(`SELECT conv_last FROM wasapBot_nodepath WHERE id_prospect = ?`, prospectID).Scan(&lastInput)
		if err != nil || !lastInput.Valid {
			return ""
		}
		return lastInput.String
	}

	// Helper function to save data based on stage - DYNAMIC DATABASE-DRIVEN APPROACH
	saveDataByStage := func(stageValue, userInput string) map[string]interface{} {
		updates := make(map[string]interface{})

		// Only apply dynamic data storage for WasapBot Exama flow
		if flowName != "WasapBot Exama" {
			// For non-WasapBot Exama flows, return basic updates
			updates["conv_last"] = userInput
			return updates
		}

		logrus.WithFields(logrus.Fields{
			"stage":     stageValue,
			"userInput": userInput,
			"deviceID":  deviceID,
		}).Info("🔄 WASAPBOT: Processing dynamic stage data storage")

		// Always save user input to conv_last
		updates["conv_last"] = userInput

		// Query stageSetValue_nodepath for dynamic configuration
		query := `
			SELECT type_inputData, inputHardCode, columnsData 
			FROM stageSetValue_nodepath 
			WHERE id_device = ? AND stage = ?
		`

		logrus.WithFields(logrus.Fields{
			"id_device": deviceID,
			"stage":     stageValue,
			"query":     "SELECT FROM stageSetValue_nodepath WHERE id_device=? AND stage=?",
		}).Debug("🔍 WASAPBOT: Querying stage configuration for specific device and stage")

		rows, err := db.Query(query, deviceID, stageValue)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"device": deviceID,
				"stage":  stageValue,
			}).Warn("Failed to query stage configuration from stageSetValue_nodepath")
			// Fall back to basic user input storage
			return updates
		}
		defer rows.Close()

		// Process each stage configuration
		hasConfig := false
		for rows.Next() {
			hasConfig = true
			var typeInputData string
			var inputHardCode sql.NullString
			var columnsData string

			if err := rows.Scan(&typeInputData, &inputHardCode, &columnsData); err != nil {
				logrus.WithError(err).Warn("Failed to scan stage config row")
				continue
			}

			// Determine the value to store based on type_inputData
			var valueToStore string
			if typeInputData == "Set" && inputHardCode.Valid {
				// Use the hardcoded value from database
				valueToStore = inputHardCode.String
				logrus.WithFields(logrus.Fields{
					"stage":           stageValue,
					"column":          columnsData,
					"hardcoded_value": valueToStore,
					"type":            "Set",
				}).Info("📝 WASAPBOT: Using hardcoded value from stageSetValue_nodepath")
			} else if typeInputData == "User Input" {
				// Use the user's actual input
				valueToStore = userInput
				logrus.WithFields(logrus.Fields{
					"stage":      stageValue,
					"column":     columnsData,
					"user_value": valueToStore,
					"type":       "User Input",
				}).Info("📝 WASAPBOT: Using user input value")
			} else {
				// Unknown type, skip
				logrus.WithFields(logrus.Fields{
					"stage":  stageValue,
					"type":   typeInputData,
					"column": columnsData,
				}).Warn("Unknown type_inputData in stageSetValue_nodepath, skipping")
				continue
			}

			// Map columnsData to actual database column names
			// This mapping ensures columnsData values match wasapBot_nodepath columns
			columnMap := map[string]string{
				"nama":         "nama",
				"alamat":       "alamat",
				"pakej":        "pakej",
				"no_fon":       "no_fon",
				"tarikh_gaji":  "tarikh_gaji",
				"cara_bayaran": "cara_bayaran",
				"status":       "status",
				"niche":        "niche",
				"umur":         "umur",
				"kerja":        "kerja",
				"sijil":        "sijil",
				"alasan":       "alasan",
				"nota":         "nota",
			}

			// Store the value in the mapped column
			if dbColumn, ok := columnMap[columnsData]; ok {
				updates[dbColumn] = valueToStore
				logrus.WithFields(logrus.Fields{
					"stage":   stageValue,
					"column":  dbColumn,
					"value":   valueToStore,
					"mapping": columnsData + " -> " + dbColumn,
				}).Info("💾 WASAPBOT: Dynamic data saved to wasapBot_nodepath column")
			} else {
				logrus.WithFields(logrus.Fields{
					"stage":  stageValue,
					"column": columnsData,
				}).Warn("Unknown column mapping in columnsData, skipping. Add mapping if needed.")
			}
		}

		// Log if no configuration was found
		if !hasConfig {
			logrus.WithFields(logrus.Fields{
				"stage":  stageValue,
				"device": deviceID,
			}).Info("📋 WASAPBOT: No stage configuration found in stageSetValue_nodepath, using fallback")

			// Fallback: Check for special completion stages
			upperStage := strings.ToUpper(strings.TrimSpace(stageValue))
			if upperStage == "HABIS" || upperStage == "COMPLETE" || upperStage == "DONE" || upperStage == "END" {
				updates["status"] = "Customer"
				updates["current_node_id"] = "end"
				logrus.Info("✅ WASAPBOT: Marked as complete/customer based on stage name")
			}

			// Fallback: Check for payment method keywords in user input (backward compatibility)
			upperInput := strings.ToUpper(strings.TrimSpace(userInput))
			if strings.Contains(upperInput, "CASH") {
				updates["cara_bayaran"] = "Online Transfer"
				logrus.Info("💳 WASAPBOT: Payment method set to Online Transfer (CASH keyword)")
			} else if strings.Contains(upperInput, "COD") {
				updates["cara_bayaran"] = "COD"
				logrus.Info("💳 WASAPBOT: Payment method set to COD")
			} else if strings.Contains(upperInput, "GAJI") {
				updates["cara_bayaran"] = "COD Time Gaji"
				logrus.Info("💳 WASAPBOT: Payment method set to COD Time Gaji")
			}
		}

		logrus.WithFields(logrus.Fields{
			"stage":   stageValue,
			"updates": updates,
		}).Info("📊 WASAPBOT: Stage data processing complete")

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
				"node":        currentNode,
				"type":        nodeType,
				"stage":       stageVal,
				"has_message": msg != "",
				"has_media":   mediaURL != "",
			}).Debug("Processing initial node")

			// Update current node in database before processing
			db.Exec(`UPDATE wasapBot_nodepath SET current_node_id = ? WHERE id_prospect = ?`, currentNode, idProspect)

			// Handle different node types dynamically
			switch nodeType {
			case "stage":
				// Update stage in database
				if stageVal != "" {
					// First update the stage in database
					db.Exec(`UPDATE wasapBot_nodepath SET stage = ? WHERE id_prospect = ?`, stageVal, idProspect)
					logrus.WithField("stage", stageVal).Info("🎯 WASAPBOT: Stage updated from node")

					// For WasapBot Exama flow, check stageSetValue_nodepath for dynamic data configuration
					if flowName == "WasapBot Exama" {
						logrus.WithFields(logrus.Fields{
							"stage":    stageVal,
							"deviceID": deviceID,
							"flowName": flowName,
						}).Info("🔍 WASAPBOT: Checking stageSetValue_nodepath for stage configuration")

						// Query stageSetValue_nodepath to check if configuration exists
						checkQuery := `
							SELECT COUNT(*) 
							FROM stageSetValue_nodepath 
							WHERE id_device = ? AND stage = ?
						`
						var configCount int
						err := db.QueryRow(checkQuery, deviceID, stageVal).Scan(&configCount)
						if err != nil {
							logrus.WithError(err).Warn("Failed to check stage configuration count")
						} else if configCount > 0 {
							logrus.WithFields(logrus.Fields{
								"stage":       stageVal,
								"deviceID":    deviceID,
								"configCount": configCount,
							}).Info("✅ WASAPBOT: Found stage configuration in stageSetValue_nodepath")

							// If user has provided input previously, use it for dynamic data storage
							// Otherwise, we'll wait for user input
							if lastUserInput := getLastUserInput(idProspect); lastUserInput != "" {
								logrus.WithField("lastInput", lastUserInput).Info("📝 WASAPBOT: Processing stage data with last user input")
								stageUpdates := saveDataByStage(stageVal, lastUserInput)
								for k, v := range stageUpdates {
									updateQuery := fmt.Sprintf("UPDATE wasapBot_nodepath SET %s = ? WHERE id_prospect = ?", k)
									db.Exec(updateQuery, v, idProspect)
								}
							}
						} else {
							logrus.WithFields(logrus.Fields{
								"stage":    stageVal,
								"deviceID": deviceID,
							}).Info("⚠️ WASAPBOT: No stage configuration found in stageSetValue_nodepath")
						}
					}
				}

			case "message":
				// Apply template replacement for WasapBot Exama flow
				if msg != "" {
					replacedMsg := replaceTemplateMessage(flowName, msg)
					err := s.SendMessageFromDevice(deviceID, phoneNumber, replacedMsg)
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
				// CRITICAL FIX: Update waiting_for_reply IMMEDIATELY and SYNCHRONOUSLY
				result, err := db.Exec(`UPDATE wasapBot_nodepath SET waiting_for_reply = 1 WHERE id_prospect = ?`, idProspect)
				if err != nil {
					logrus.WithError(err).Error("❌ WASAPBOT: Failed to set waiting_for_reply=1")
				} else {
					rowsAffected, _ := result.RowsAffected()
					logrus.WithFields(logrus.Fields{
						"id_prospect":   idProspect,
						"rows_affected": rowsAffected,
					}).Info("✅ WASAPBOT: Set waiting_for_reply=1 immediately at user_reply node")
				}
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

		// DON'T process stage data here - wait until we know the NEXT stage
		// Just log what stage we're coming FROM
		if stage.Valid && stage.String != "" {
			logrus.WithFields(logrus.Fields{
				"previous_stage": stage.String,
				"user_input":     content,
			}).Info("📋 WASAPBOT: User replied while at stage (will process with next stage)")
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
				"user_input":     content,
				"next_node":      nextNodeID,
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
					"node":  currentNode,
					"type":  nodeType,
					"stage": stageVal,
				}).Info("🎯 WASAPBOT: Processing next node")

				updates["current_node_id"] = currentNode

				switch nodeType {
				case "stage":
					updates["stage"] = stageVal
					logrus.WithField("stage", stageVal).Info("🎯 WASAPBOT: Stage updated from node")

					// For WasapBot Exama flow, process dynamic data storage when stage changes
					if flowName == "WasapBot Exama" && stageVal != "" {
						logrus.WithFields(logrus.Fields{
							"new_stage": stageVal,
							"deviceID":  deviceID,
							"userInput": content,
							"info":      "Processing with NEW stage from node, not old stage from DB",
						}).Info("🔍 WASAPBOT: Processing stage data for WasapBot Exama with CURRENT NODE stage")

						// Check if stage configuration exists in stageSetValue_nodepath
						checkQuery := `
							SELECT COUNT(*) 
							FROM stageSetValue_nodepath 
							WHERE id_device = ? AND stage = ?
						`
						logrus.WithFields(logrus.Fields{
							"query":     "WHERE id_device = ? AND stage = ?",
							"id_device": deviceID,
							"stage":     stageVal,
						}).Debug("🔎 WASAPBOT: Checking stageSetValue_nodepath with device AND stage")

						var configCount int
						err := db.QueryRow(checkQuery, deviceID, stageVal).Scan(&configCount)
						if err != nil {
							logrus.WithError(err).Warn("Failed to check stage configuration count")
						} else if configCount > 0 {
							logrus.WithFields(logrus.Fields{
								"stage":       stageVal,
								"id_device":   deviceID,
								"configCount": configCount,
								"query_used":  fmt.Sprintf("WHERE id_device='%s' AND stage='%s'", deviceID, stageVal),
							}).Info("✅ WASAPBOT: Found stage configuration in stageSetValue_nodepath for this device and stage")

							// Process dynamic data storage with user input
							stageUpdates := saveDataByStage(stageVal, content)
							for k, v := range stageUpdates {
								updates[k] = v
								logrus.WithFields(logrus.Fields{
									"field": k,
									"value": v,
								}).Info("💾 WASAPBOT: Stage data field set")
							}
						} else {
							logrus.WithFields(logrus.Fields{
								"stage":      stageVal,
								"id_device":  deviceID,
								"query_used": fmt.Sprintf("WHERE id_device='%s' AND stage='%s'", deviceID, stageVal),
							}).Info("⚠️ WASAPBOT: No stage configuration found in stageSetValue_nodepath for this device and stage combination")
							// Just save the stage, no dynamic data processing
						}
					}

				case "message":
					// Apply template replacement for WasapBot Exama flow
					if msg != "" {
						replacedMsg := replaceTemplateMessage(flowName, msg)
						err := s.SendMessageFromDevice(deviceID, phoneNumber, replacedMsg)
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
					// CRITICAL FIX: Update waiting_for_reply IMMEDIATELY and SYNCHRONOUSLY
					// Don't just add to updates map - apply immediately so next message sees it
					_, execErr := db.Exec(`UPDATE wasapBot_nodepath SET waiting_for_reply = 1 WHERE phone_number = ? AND id_device = ?`, phoneNumber, deviceID)
					if execErr != nil {
						logrus.WithError(execErr).Error("❌ WASAPBOT: Failed to set waiting_for_reply=1")
					} else {
						logrus.WithFields(logrus.Fields{
							"phone_number": phoneNumber,
							"device_id":    deviceID,
						}).Info("✅ WASAPBOT: Set waiting_for_reply=1 immediately at user_reply node (existing prospect)")
					}
					updates["waiting_for_reply"] = 1 // Also keep in updates for consistency
					logrus.Info("🎯 WASAPBOT: Waiting for user input")
					break

				case "condition":
					// We've moved to a condition node - evaluate it immediately with current user input
					nextCondNode := processConditionNode(currentNode, content)
					logrus.WithFields(logrus.Fields{
						"condition_node":       currentNode,
						"user_input":           content,
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
			setClauses = append(setClauses, field+" = ?")
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
		"stage":   stage,
		"updates": updates,
	}).Info("🎯 WASAPBOT: Flow processing completed")

	return nil
}
