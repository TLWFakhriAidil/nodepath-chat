// FIX 1: Don't save user message again for existing executions
// In ProcessIncomingMessageFromWebhook function, around line 319-325

// Change this section:
if aiExecution.CurrentNodeID.Valid && aiExecution.CurrentNodeID.String != "" {
    // Process through the flow using current node
    // DON'T call processNewFlowExecution for existing executions
    // Instead, process directly without re-saving user message
    
    // Get the flow data
    flow, err := s.flowService.GetFlow(aiExecution.FlowReference.String)
    if err != nil || flow == nil {
        return fmt.Errorf("flow not found")
    }
    
    // Process the message WITHOUT saving user input again
    response, err := s.processFlowMessage(flow, aiExecution, content)
    if err != nil {
        return err
    }
    
    // Send response if not empty
    if response != "" && strings.TrimSpace(response) != "" {
        // Process and send response
        s.SendMessageFromDevice(deviceID, phoneNumber, response)
    }
    
    return nil
}

// FIX 2: In processUserReplyNode, ensure we only advance when there's actual user input
// This is already fixed in our previous change

// FIX 3: For message nodes followed by user_reply, ensure we set waiting state
// In processMessageNode, after sending message, check if next node is user_reply

func (s *Service) processMessageNode(flow *models.ChatbotFlow, execution *models.AIWhatsapp, node *models.FlowNode, userInput string) (string, error) {
    // Get message from node data
    message := ""
    if msg, ok := node.Data["message"].(string); ok {
        message = msg
    }

    // Replace variables
    variables, _ := s.aiWhatsappService.GetFlowExecutionVariables(execution.ProspectNum, execution.IDDevice)
    message = s.flowService.ReplaceVariables(message, variables)

    // Check next node
    nextNode, err := s.flowService.GetNextNode(flow, node.ID)
    if err == nil && nextNode != nil {
        // If next node is user_reply, set waiting state
        if nextNode.Type == models.NodeTypeUserReply || nextNode.Type == "user_reply" {
            // Update to user_reply node and set waiting
            s.updateCurrentNode(execution, nextNode.ID)
            s.updateFlowTrackingFields(execution, nextNode.ID, flow.ID, true)
            s.aiWhatsappService.UpdateFlowExecution(execution.ProspectNum, execution.IDDevice, nextNode.ID, make(map[string]interface{}), "active")
            
            // Return message to send but don't continue processing
            return message, nil
        }
        
        // For other node types, continue as before
        // ... existing logic ...
    }
    
    return message, nil
}
