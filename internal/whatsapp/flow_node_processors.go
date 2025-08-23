package whatsapp

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"nodepath-chat/internal/models"
)

// processNode processes a single node based on its type
func (e *FlowEngine) processNode(ctx *ExecutionContext) error {
	logrus.WithFields(logrus.Fields{
		"node_id":   ctx.CurrentNode.ID,
		"node_type": ctx.CurrentNode.Type,
	}).Info("🎯 FLOW_ENGINE: Processing node")

	switch ctx.CurrentNode.Type {
	case models.NodeTypeStart:
		return e.processStartNode(ctx)
	case models.NodeTypeMessage:
		return e.processMessageNode(ctx)
	case models.NodeTypeImage:
		return e.processImageNode(ctx)
	case models.NodeTypeAudio:
		return e.processAudioNode(ctx)
	case models.NodeTypeVideo:
		return e.processVideoNode(ctx)
	case models.NodeTypeAIPrompt:
		return e.processAIPromptNode(ctx)
	case models.NodeTypeAdvancedAIPrompt:
		return e.processAdvancedAIPromptNode(ctx)
	case models.NodeTypeUserReply:
		return e.processUserReplyNode(ctx)
	case models.NodeTypeDelay:
		return e.processDelayNode(ctx)
	case models.NodeTypeCondition:
		return e.processConditionNode(ctx)
	case models.NodeTypeStage:
		return e.processStageNode(ctx)
	case models.NodeTypeManual:
		return e.processManualNode(ctx)
	case models.NodeTypeWaitingReplyTimes:
		return e.processWaitingReplyTimesNode(ctx)
	default:
		logrus.WithField("node_type", ctx.CurrentNode.Type).Warn("⚠️ FLOW_ENGINE: Unknown node type, skipping")
		return nil
	}
}

// processStartNode processes start node - initializes the flow
func (e *FlowEngine) processStartNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🚀 FLOW_ENGINE: Processing start node")
	
	// Start node just initializes the flow, no response needed
	// Continue to next node immediately
	return nil
}

// processMessageNode processes message node - sends a text message
func (e *FlowEngine) processMessageNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("💬 FLOW_ENGINE: Processing message node")
	
	if ctx.CurrentNode.Data == nil {
		return nil
	}
	
	// Get message content
	message, ok := ctx.CurrentNode.Data["message"].(string)
	if !ok || message == "" {
		// Try alternative field names
		if label, exists := ctx.CurrentNode.Data["label"].(string); exists {
			message = label
		}
	}
	
	if message != "" {
		// Replace variables in message
		message = e.flowService.ReplaceVariables(message, ctx.Variables)
		
		// Add to response queue
		ctx.Response = append(ctx.Response, message)
		
		logrus.WithFields(logrus.Fields{
			"node_id": ctx.CurrentNode.ID,
			"message": message,
		}).Info("📝 FLOW_ENGINE: Message queued for sending")
	}
	
	return nil
}

// processImageNode processes image node - sends an image
func (e *FlowEngine) processImageNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🖼️ FLOW_ENGINE: Processing image node")
	
	if ctx.CurrentNode.Data == nil {
		return nil
	}
	
	// Get image URL and caption
	imageURL, _ := ctx.CurrentNode.Data["imageUrl"].(string)
	if imageURL == "" {
		imageURL, _ = ctx.CurrentNode.Data["mediaUrl"].(string)
	}
	
	caption, _ := ctx.CurrentNode.Data["caption"].(string)
	
	if imageURL != "" {
		// Get device settings for sending message
		deviceSettings, err := e.deviceSettingsService.GetByIDDevice(ctx.Execution.IDDevice)
		if err != nil {
			return fmt.Errorf("failed to get device settings: %w", err)
		}
		
		// Send image using provider service
		err = e.providerService.SendMediaMessage(deviceSettings, ctx.Execution.ProspectNum, caption, imageURL)
		if err != nil {
			return fmt.Errorf("failed to send image: %w", err)
		}
		
		logrus.WithFields(logrus.Fields{
			"node_id":   ctx.CurrentNode.ID,
			"image_url": imageURL,
			"caption":   caption,
		}).Info("📸 FLOW_ENGINE: Image sent")
		
		// Caption is already sent with the media, no need to add as separate text response
	}
	
	return nil
}

// processAudioNode processes audio node - sends an audio file
func (e *FlowEngine) processAudioNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🎵 FLOW_ENGINE: Processing audio node")
	
	if ctx.CurrentNode.Data == nil {
		return nil
	}
	
	// Get audio URL and caption
	audioURL, _ := ctx.CurrentNode.Data["audioUrl"].(string)
	if audioURL == "" {
		audioURL, _ = ctx.CurrentNode.Data["mediaUrl"].(string)
	}
	
	caption, _ := ctx.CurrentNode.Data["caption"].(string)
	
	if audioURL != "" {
		// Get device settings for sending message
		deviceSettings, err := e.deviceSettingsService.GetByIDDevice(ctx.Execution.IDDevice)
		if err != nil {
			return fmt.Errorf("failed to get device settings: %w", err)
		}
		
		// Send audio using provider service
		err = e.providerService.SendMediaMessage(deviceSettings, ctx.Execution.ProspectNum, caption, audioURL)
		if err != nil {
			return fmt.Errorf("failed to send audio: %w", err)
		}
		
		logrus.WithFields(logrus.Fields{
			"node_id":   ctx.CurrentNode.ID,
			"audio_url": audioURL,
			"caption":   caption,
		}).Info("🎵 FLOW_ENGINE: Audio sent")
		
		// Caption is already sent with the media, no need to add as separate text response
	}
	
	return nil
}

// processVideoNode processes video node - sends a video file
func (e *FlowEngine) processVideoNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🎬 FLOW_ENGINE: Processing video node")
	
	if ctx.CurrentNode.Data == nil {
		return nil
	}
	
	// Get video URL and caption
	videoURL, _ := ctx.CurrentNode.Data["videoUrl"].(string)
	if videoURL == "" {
		videoURL, _ = ctx.CurrentNode.Data["mediaUrl"].(string)
	}
	
	caption, _ := ctx.CurrentNode.Data["caption"].(string)
	
	if videoURL != "" {
		// Get device settings for sending message
		deviceSettings, err := e.deviceSettingsService.GetByIDDevice(ctx.Execution.IDDevice)
		if err != nil {
			return fmt.Errorf("failed to get device settings: %w", err)
		}
		
		// Send video using provider service
		err = e.providerService.SendMediaMessage(deviceSettings, ctx.Execution.ProspectNum, caption, videoURL)
		if err != nil {
			return fmt.Errorf("failed to send video: %w", err)
		}
		
		logrus.WithFields(logrus.Fields{
			"node_id":   ctx.CurrentNode.ID,
			"video_url": videoURL,
			"caption":   caption,
		}).Info("🎬 FLOW_ENGINE: Video sent")
		
		// Caption is already sent with the media, no need to add as separate text response
	}
	
	return nil
}

// processAIPromptNode processes AI prompt node - generates AI response
func (e *FlowEngine) processAIPromptNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🤖 FLOW_ENGINE: Processing AI prompt node")
	
	if ctx.CurrentNode.Data == nil {
		logrus.Warn("AI prompt node has no data configured")
		fallbackResponse := "I'm here to help! Could you please rephrase your question?"
		ctx.Response = append(ctx.Response, fallbackResponse)
		return nil
	}
	
	// Get AI prompt from node data
	prompt, ok := ctx.CurrentNode.Data["prompt"].(string)
	if !ok {
		prompt, _ = ctx.CurrentNode.Data["systemPrompt"].(string)
	}
	
	if prompt == "" {
		logrus.Warn("AI prompt node has no prompt configured")
		fallbackResponse := "I'm here to help! Could you please rephrase your question?"
		ctx.Response = append(ctx.Response, fallbackResponse)
		return nil
	}
	
	// Get device settings for AI configuration
	deviceSettings, err := e.deviceSettingsService.GetByIDDevice(ctx.Execution.IDDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings for AI prompt node")
		fallbackResponse := "I apologize, but I'm having trouble processing your request right now. Please try again."
		ctx.Response = append(ctx.Response, fallbackResponse)
		return nil
	}
	
	// Validate device settings exist
	if deviceSettings == nil {
		logrus.WithField("id_device", ctx.Execution.IDDevice).Error("Device settings not found for AI prompt node")
		fallbackResponse := "I apologize, but I'm having trouble processing your request right now. Please try again."
		ctx.Response = append(ctx.Response, fallbackResponse)
		return nil
	}
	
	// Get API key from device settings with validation
	apiKey := ""
	if deviceSettings.APIKey.Valid && deviceSettings.APIKey.String != "" {
		apiKey = deviceSettings.APIKey.String
	} else {
		// Check for special devices that use hardcoded API key
		if ctx.Execution.IDDevice == "SCHQ-S94" || ctx.Execution.IDDevice == "SCHQ-S12" {
			apiKey = "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"
			logrus.WithField("id_device", ctx.Execution.IDDevice).Info("Using hardcoded API key for special device")
		} else {
			logrus.WithField("id_device", ctx.Execution.IDDevice).Warn("No API key configured for device, using fallback response")
			fallbackResponse := "I'm here to help! Could you please rephrase your question?"
			ctx.Response = append(ctx.Response, fallbackResponse)
			return nil
		}
	}
	
	// Replace variables in prompt
	prompt = e.flowService.ReplaceVariables(prompt, ctx.Variables)
	
	// Validate prompt after variable replacement
	if prompt == "" {
		logrus.Warn("Prompt is empty after variable replacement")
		fallbackResponse := "I'm here to help! Could you please rephrase your question?"
		ctx.Response = append(ctx.Response, fallbackResponse)
		return nil
	}
	
	// Generate AI response with enhanced error handling
	response, err := e.aiService.GenerateResponse(prompt, ctx.UserInput, apiKey, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"node_id": ctx.CurrentNode.ID,
			"error": err.Error(),
			"has_api_key": apiKey != "",
		}).Error("AI generation failed")
		// Add fallback response instead of returning error
		fallbackResponse := "I apologize, but I'm having trouble processing your request right now. Please try again."
		ctx.Response = append(ctx.Response, fallbackResponse)
		return nil
	}
	
	// Enhanced response validation to prevent <nil> responses
	if response != "" && response != "<nil>" && response != "null" && response != "nil" && len(response) > 0 {
		// Additional validation to ensure response is meaningful
		trimmedResponse := strings.TrimSpace(response)
		if trimmedResponse != "" && trimmedResponse != "<nil>" && trimmedResponse != "null" && trimmedResponse != "nil" {
			ctx.Response = append(ctx.Response, trimmedResponse)
			
			logrus.WithFields(logrus.Fields{
				"node_id":  ctx.CurrentNode.ID,
				"response_length": len(trimmedResponse),
			}).Info("🤖 FLOW_ENGINE: AI response generated successfully")
			return nil
		}
	}
	
	// Handle empty, nil, or invalid response
	logrus.WithFields(logrus.Fields{
		"node_id": ctx.CurrentNode.ID,
		"raw_response": response,
		"response_length": len(response),
	}).Warn("AI returned empty or invalid response, using fallback")
	
	fallbackResponse := "I'm here to help! Could you please rephrase your question?"
	ctx.Response = append(ctx.Response, fallbackResponse)
	
	return nil
}

// processAdvancedAIPromptNode processes advanced AI prompt node with structured output
func (e *FlowEngine) processAdvancedAIPromptNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🧠 FLOW_ENGINE: Processing advanced AI prompt node")
	
	if ctx.CurrentNode.Data == nil {
		return nil
	}
	
	// Get AI prompt from node data
	prompt, ok := ctx.CurrentNode.Data["prompt"].(string)
	if !ok {
		prompt, _ = ctx.CurrentNode.Data["systemPrompt"].(string)
	}
	
	if prompt == "" {
		logrus.Warn("Advanced AI prompt node has no prompt configured")
		return nil
	}
	
	// Get device settings for AI configuration
	deviceSettings, err := e.deviceSettingsService.GetByIDDevice(ctx.Execution.IDDevice)
	if err != nil {
		return fmt.Errorf("failed to get device settings: %w", err)
	}
	
	// Get API key from device settings
	apiKey := ""
	if deviceSettings.APIKey.Valid {
		apiKey = deviceSettings.APIKey.String
	}
	
	// Replace variables in prompt
	prompt = e.flowService.ReplaceVariables(prompt, ctx.Variables)
	
	// Get closing prompt if available
	closingPrompt, _ := ctx.CurrentNode.Data["closingPrompt"].(string)
	
	// Generate advanced AI response
	response, err := e.aiService.GenerateAdvancedResponse(prompt, ctx.UserInput, apiKey, nil, closingPrompt)
	if err != nil {
		logrus.WithError(err).Error("Advanced AI generation failed")
		// Add fallback response instead of returning error
		fallbackResponse := "I apologize, but I'm having trouble processing your request right now. Please try again."
		ctx.Response = append(ctx.Response, fallbackResponse)
		return nil
	}
	
	if response != nil && len(response.Response) > 0 {
		// Process structured response
		validResponseAdded := false
		for _, item := range response.Response {
			if item.Type == "text" && item.Content != "" && item.Content != "<nil>" && item.Content != "null" {
				ctx.Response = append(ctx.Response, item.Content)
				validResponseAdded = true
			}
			// TODO: Handle image responses
		}
		
		if validResponseAdded {
			logrus.WithFields(logrus.Fields{
				"node_id": ctx.CurrentNode.ID,
				"stage":   response.Stage,
				"items":   len(response.Response),
			}).Info("🧠 FLOW_ENGINE: Advanced AI response generated")
		} else {
			// No valid responses found, add fallback
			logrus.WithFields(logrus.Fields{
				"node_id": ctx.CurrentNode.ID,
				"response_items": len(response.Response),
			}).Warn("Advanced AI returned no valid text responses, using fallback")
			
			fallbackResponse := "I'm here to help! Could you please rephrase your question?"
			ctx.Response = append(ctx.Response, fallbackResponse)
		}
	} else {
		// Handle empty or nil response
		logrus.WithField("node_id", ctx.CurrentNode.ID).Warn("Advanced AI returned empty response, using fallback")
		fallbackResponse := "I'm here to help! Could you please rephrase your question?"
		ctx.Response = append(ctx.Response, fallbackResponse)
	}
	
	return nil
}

// processUserReplyNode processes user reply node - waits for user input
func (e *FlowEngine) processUserReplyNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("👤 FLOW_ENGINE: Processing user reply node")
	
	// User reply node stops execution and waits for user input
	// The next time a message comes in, it will continue from the next node
	ctx.ShouldStop = true
	
	logrus.Info("⏸️ FLOW_ENGINE: Execution paused for user reply")
	return nil
}

// processDelayNode processes delay node - adds a delay
func (e *FlowEngine) processDelayNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("⏰ FLOW_ENGINE: Processing delay node")
	
	if ctx.CurrentNode.Data == nil {
		return nil
	}
	
	// Get delay duration
	var delaySeconds int
	
	// Try different field names for delay
	if delay, ok := ctx.CurrentNode.Data["delay"].(float64); ok {
		delaySeconds = int(delay)
	} else if delay, ok := ctx.CurrentNode.Data["delaySeconds"].(float64); ok {
		delaySeconds = int(delay)
	} else if delayStr, ok := ctx.CurrentNode.Data["delay"].(string); ok {
		if parsed, err := strconv.Atoi(delayStr); err == nil {
			delaySeconds = parsed
		}
	}
	
	if delaySeconds > 0 {
		logrus.WithFields(logrus.Fields{
			"node_id": ctx.CurrentNode.ID,
			"delay":   delaySeconds,
		}).Info("⏰ FLOW_ENGINE: Applying delay")
		
		// Apply delay (max 30 seconds for safety)
		if delaySeconds > 30 {
			delaySeconds = 30
		}
		time.Sleep(time.Duration(delaySeconds) * time.Second)
	}
	
	return nil
}

// processConditionNode processes condition node - evaluates conditions
func (e *FlowEngine) processConditionNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🔀 FLOW_ENGINE: Processing condition node")
	
	// Condition nodes don't generate responses, they just evaluate conditions
	// The actual condition evaluation happens in the getNextNode method
	// when determining which path to take based on user input
	
	logrus.WithFields(logrus.Fields{
		"node_id": ctx.CurrentNode.ID,
		"user_input": ctx.UserInput,
	}).Info("🔀 FLOW_ENGINE: Condition node processed, evaluation will happen during path selection")
	
	return nil
}

// processStageNode processes stage node - updates conversation stage
func (e *FlowEngine) processStageNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("🏷️ FLOW_ENGINE: Processing stage node")
	
	if ctx.CurrentNode.Data == nil {
		return nil
	}
	
	// Get stage name
	stageName, ok := ctx.CurrentNode.Data["stageName"].(string)
	if !ok || stageName == "" {
		stageName, _ = ctx.CurrentNode.Data["stage"].(string)
	}
	
	if stageName != "" {
		// Update conversation stage
		ctx.Variables["current_stage"] = stageName
		
		logrus.WithFields(logrus.Fields{
			"node_id": ctx.CurrentNode.ID,
			"stage":   stageName,
		}).Info("🏷️ FLOW_ENGINE: Stage updated")
	}
	
	return nil
}

// processManualNode processes manual node - for human intervention
func (e *FlowEngine) processManualNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("👨‍💼 FLOW_ENGINE: Processing manual node")
	
	// Manual node typically requires human intervention
	// For now, just log and continue
	// TODO: Implement proper manual intervention logic
	
	if ctx.CurrentNode.Data != nil {
		if message, ok := ctx.CurrentNode.Data["message"].(string); ok && message != "" {
			message = e.flowService.ReplaceVariables(message, ctx.Variables)
			ctx.Response = append(ctx.Response, message)
		}
	}
	
	logrus.Info("👨‍💼 FLOW_ENGINE: Manual node processed")
	return nil
}

// processWaitingReplyTimesNode processes waiting reply times node
func (e *FlowEngine) processWaitingReplyTimesNode(ctx *ExecutionContext) error {
	logrus.WithField("node_id", ctx.CurrentNode.ID).Info("⏳ FLOW_ENGINE: Processing waiting reply times node")
	
	// TODO: Implement waiting reply times logic
	// This would typically track how many times we've waited for a reply
	// and potentially branch to different paths based on that
	
	logrus.Info("⏳ FLOW_ENGINE: Waiting reply times node processed")
	return nil
}