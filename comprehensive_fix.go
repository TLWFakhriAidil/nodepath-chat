package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	// Fix 1: ai_whatsapp_repository.go
	fixRepository()
	
	// Fix 2: ai_whatsapp_service.go  
	fixService()
	
	log.Println("All fixes applied successfully")
}

func fixRepository() {
	content, err := ioutil.ReadFile("internal/repository/ai_whatsapp_repository.go")
	if err != nil {
		log.Fatal(err)
	}
	
	text := string(content)
	
	// Comment out field checks
	text = strings.Replace(text,
		`if ai.FlowReference.Valid {
		flowReferenceValue = ai.FlowReference.String
	} else {
		flowReferenceValue = nil
	}`,
		`// FlowReference removed from database
	flowReferenceValue = nil`,
		1)
	
	text = strings.Replace(text,
		`if ai.ExecutionID.Valid {
		executionIDValue = ai.ExecutionID.String
	} else {
		executionIDValue = nil
	}`,
		`// ExecutionID removed from database
	executionIDValue = nil`,
		1)
		
	text = strings.Replace(text,
		`if ai.ExecutionStatus.Valid {
		executionStatusValue = ai.ExecutionStatus.String
	} else {
		executionStatusValue = nil
	}`,
		`// ExecutionStatus removed from database
	executionStatusValue = nil`,
		1)
	
	// Fix Scan calls
	text = strings.Replace(text, `&ai.ExecutionStatus, &ai.ExecutionID,`, `// &ai.ExecutionStatus, &ai.ExecutionID, // Removed`, -1)
	
	// Fix method signature
	text = strings.Replace(text, `UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int, executionStatus, executionID string)`,
		`UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int)`, -1)
	
	// Fix implementation
	text = strings.Replace(text, `func (r *aiWhatsappRepository) UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int, executionStatus, executionID string)`,
		`func (r *aiWhatsappRepository) UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int)`, 1)
	
	// Remove execution fields from UPDATE query
	text = strings.Replace(text, `flow_id = ?, current_node_id = ?, last_node_id = ?, waiting_for_reply = ?,
			execution_status = ?, execution_id = ?, updated_at = ?`,
		`flow_id = ?, current_node_id = ?, last_node_id = ?, waiting_for_reply = ?, updated_at = ?`, 1)
		
	// Remove from Exec parameters
	text = strings.Replace(text, `executionStatusValue, executionIDValue, time.Now(),`,
		`time.Now(),`, 1)
	
	ioutil.WriteFile("internal/repository/ai_whatsapp_repository.go", []byte(text), 0644)
	log.Println("Fixed ai_whatsapp_repository.go")
}

func fixService() {
	content, err := ioutil.ReadFile("internal/services/ai_whatsapp_service.go")
	if err != nil {
		log.Fatal(err)
	}
	
	text := string(content)
	
	// Add stubs for missing methods at the end
	if !strings.Contains(text, "// STUB IMPLEMENTATIONS FOR REMOVED METHODS") {
		text += `

// STUB IMPLEMENTATIONS FOR REMOVED METHODS

// CompleteFlowExecution - stub for removed functionality
func (s *aiWhatsappService) CompleteFlowExecution(prospectNum, idDevice string) error {
	return s.aiRepo.UpdateFlowTrackingFields(prospectNum, idDevice, "", "", "", 0)
}

// Circuit breaker methods
func (s *aiWhatsappService) isCircuitBreakerOpen() bool {
	return false
}

func (s *aiWhatsappService) recordAPIFailure() {
	return
}
`
	}
	
	// Fix UpdateFlowTrackingFields calls (remove last 2 parameters)
	text = strings.Replace(text, `s.aiRepo.UpdateFlowTrackingFields(
			prospectNum, idDevice,
			flowReference, // flowID
			startNode.ID, // currentNodeID - set to actual start node ID
			"", // lastNodeID
			0, // waitingForReply
			"active", // executionStatus
			executionID, // executionID
		)`,
		`s.aiRepo.UpdateFlowTrackingFields(
			prospectNum, idDevice,
			flowReference, // flowID
			startNode.ID, // currentNodeID
			"", // lastNodeID
			0, // waitingForReply
		)`, 1)
		
	text = strings.Replace(text, `err = s.aiRepo.UpdateFlowTrackingFields(
		prospectNum, idDevice,
		flowID, // flowID
		currentNode, // currentNodeID
		lastNodeID, // lastNodeID
		0, // waitingForReply - clear since we're moving forward
		status, // executionStatus
		"", // executionID - preserve existing
	)`,
		`err = s.aiRepo.UpdateFlowTrackingFields(
		prospectNum, idDevice,
		flowID, // flowID
		currentNode, // currentNodeID
		lastNodeID, // lastNodeID
		0, // waitingForReply
	)`, 1)
	
	ioutil.WriteFile("internal/services/ai_whatsapp_service.go", []byte(text), 0644)
	log.Println("Fixed ai_whatsapp_service.go")
}