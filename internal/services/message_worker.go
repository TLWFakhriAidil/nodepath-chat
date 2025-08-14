package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// MessageWorker handles individual message processing
type MessageWorker struct {
	id        int
	processor *MessageProcessor
	logger    *logrus.Entry
}

// NewMessageWorker creates a new message worker
func NewMessageWorker(id int, processor *MessageProcessor) *MessageWorker {
	return &MessageWorker{
		id:        id,
		processor: processor,
		logger: logrus.WithFields(logrus.Fields{
			"component": "message_worker",
			"worker_id": id,
		}),
	}
}

// Start starts the worker
func (mw *MessageWorker) Start(ctx context.Context) {
	mw.logger.Info("Starting message worker")
	
	for {
		select {
		case msg := <-mw.processor.incomingChan:
			if msg == nil {
				mw.logger.Info("Worker shutting down - incoming channel closed")
				return
			}
			
			// Process message with timeout
			result := mw.processMessageWithTimeout(ctx, msg)
			
			// Send result
			select {
			case mw.processor.resultChan <- result:
			case <-ctx.Done():
				return
			}
			
		case <-ctx.Done():
			mw.logger.Info("Worker shutting down - context cancelled")
			return
		}
	}
}

// processMessageWithTimeout processes a message with timeout
func (mw *MessageWorker) processMessageWithTimeout(ctx context.Context, msg *IncomingMessage) *ProcessingResult {
	start := time.Now()
	
	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, mw.processor.config.ProcessingTimeout)
	defer cancel()
	
	result := &ProcessingResult{
		MessageID:   msg.ID,
		ProcessedAt: time.Now(),
	}
	
	// Process message in goroutine to handle timeout
	done := make(chan bool, 1)
	var response string
	var err error
	
	go func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic during message processing: %v", r)
				mw.logger.WithField("panic", r).Error("Worker panic recovered")
			}
			done <- true
		}()
		
		response, err = mw.processMessage(timeoutCtx, msg)
	}()
	
	// Wait for completion or timeout
	select {
	case <-done:
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			
			// Retry logic
			if msg.RetryCount < msg.MaxRetries {
				msg.RetryCount++
				mw.logger.WithFields(logrus.Fields{
					"message_id":   msg.ID,
					"retry_count":  msg.RetryCount,
					"max_retries":  msg.MaxRetries,
					"error":        err.Error(),
				}).Warn("Retrying message processing")
				
				// Schedule retry with exponential backoff
				retryDelay := time.Duration(msg.RetryCount) * mw.processor.config.RetryDelay
				go func() {
					time.Sleep(retryDelay)
					select {
					case mw.processor.incomingChan <- msg:
					case <-ctx.Done():
					}
				}()
			}
		} else {
			result.Success = true
			result.Response = response
		}
		
	case <-timeoutCtx.Done():
		result.Success = false
		result.Error = "processing timeout"
		mw.logger.WithFields(logrus.Fields{
			"message_id": msg.ID,
			"timeout":    mw.processor.config.ProcessingTimeout,
		}).Error("Message processing timeout")
	}
	
	result.ProcessingTime = time.Since(start)
	return result
}

// processMessage processes a single message based on its type
func (mw *MessageWorker) processMessage(ctx context.Context, msg *IncomingMessage) (string, error) {
	mw.logger.WithFields(logrus.Fields{
		"message_id":   msg.ID,
		"message_type": msg.Type,
		"device_id":    msg.DeviceID,
		"phone_number": msg.PhoneNumber,
	}).Debug("Processing message")
	
	switch msg.Type {
	case MessageTypeUserReply:
		return mw.processUserReply(ctx, msg)
	case MessageTypeCustomerReply:
		return mw.processCustomerReply(ctx, msg)
	case MessageTypeAIReply:
		return mw.processAIReply(ctx, msg)
	case MessageTypeSystem:
		return mw.processSystemMessage(ctx, msg)
	default:
		return "", fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// processUserReply processes user reply messages
func (mw *MessageWorker) processUserReply(ctx context.Context, msg *IncomingMessage) (string, error) {
	// Get or create execution for the user
	execution, err := mw.processor.chatService.GetExecution(msg.PhoneNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get execution: %w", err)
	}
	
	// Get the chatbot flow
	_, err = mw.processor.flowService.GetFlow(execution.FlowReference)
	if err != nil {
		return "", fmt.Errorf("failed to get flow: %w", err)
	}
	
	// Add user message to conversation
	mw.processor.chatService.AddConversationMessage(execution, "user", msg.Content)
	
	// Process through flow engine - simplified flow processing
	response := "Thank you for your message. We'll get back to you soon."
	
	// Add bot response to conversation
	mw.processor.chatService.AddConversationMessage(execution, "assistant", response)
	
	// Update execution in database
	if err := mw.processor.chatService.UpdateExecution(execution); err != nil {
		mw.logger.WithError(err).Warn("Failed to update execution")
	}
	
	// Send response back to user
	if err := mw.sendResponse(ctx, msg.DeviceID, msg.PhoneNumber, response); err != nil {
		return response, fmt.Errorf("failed to send response: %w", err)
	}
	
	// Update metrics
	mw.processor.metrics.mu.Lock()
	mw.processor.metrics.UserRepliesProcessed++
	mw.processor.metrics.mu.Unlock()
	
	return response, nil
}

// processCustomerReply processes customer reply messages
func (mw *MessageWorker) processCustomerReply(ctx context.Context, msg *IncomingMessage) (string, error) {
	// Store customer message using AddConversationMessage
	execution, err := mw.processor.chatService.GetExecution(msg.PhoneNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get execution: %w", err)
	}
	
	if err := mw.processor.chatService.AddConversationMessage(execution, "user", msg.Content); err != nil {
		return "", fmt.Errorf("failed to store customer message: %w", err)
	}
	
	// Notify relevant staff/agents about new customer message
	if err := mw.notifyStaff(ctx, msg); err != nil {
		mw.logger.WithError(err).Warn("Failed to notify staff")
	}
	
	// Update metrics
	mw.processor.metrics.mu.Lock()
	mw.processor.metrics.CustomerRepliesProcessed++
	mw.processor.metrics.mu.Unlock()
	
	return "Customer message received", nil
}

// processAIReply processes AI-generated reply messages
func (mw *MessageWorker) processAIReply(ctx context.Context, msg *IncomingMessage) (string, error) {
	// Check circuit breaker for AI service
	if mw.processor.config.EnableCircuitBreaker && !mw.processor.circuitBreaker.CanExecute() {
		return "", fmt.Errorf("AI service circuit breaker is open")
	}
	
	// Get conversation history
	execution, err := mw.processor.chatService.GetExecution(msg.PhoneNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get execution: %w", err)
	}
	
	conversationHistory, err := mw.processor.chatService.GetConversationHistory(execution)
	if err != nil {
		return "", fmt.Errorf("failed to get conversation history: %w", err)
	}
	
	// Generate AI response with correct parameters
	aiResponse, err := mw.processor.aiService.GenerateResponse("You are a helpful assistant", msg.Content, "your-api-key", conversationHistory)
	if err != nil {
		// Record failure in circuit breaker
		if mw.processor.config.EnableCircuitBreaker {
			mw.processor.circuitBreaker.RecordFailure()
		}
		return "", fmt.Errorf("failed to generate AI response: %w", err)
	}
	
	// Record success in circuit breaker
	if mw.processor.config.EnableCircuitBreaker {
		mw.processor.circuitBreaker.RecordSuccess()
	}
	
	// Send AI response
	if err := mw.sendResponse(ctx, msg.DeviceID, msg.PhoneNumber, aiResponse); err != nil {
		return aiResponse, fmt.Errorf("failed to send AI response: %w", err)
	}
	
	// Store AI conversation
	if err := mw.processor.chatService.AddConversationMessage(execution, "assistant", aiResponse); err != nil {
		mw.logger.WithError(err).Warn("Failed to store AI message")
	}
	
	// Update metrics
	mw.processor.metrics.mu.Lock()
	mw.processor.metrics.AIRepliesProcessed++
	mw.processor.metrics.mu.Unlock()
	
	return aiResponse, nil
}

// processSystemMessage processes system messages
func (mw *MessageWorker) processSystemMessage(ctx context.Context, msg *IncomingMessage) (string, error) {
	// Handle system-level operations
	mw.logger.WithFields(logrus.Fields{
		"message_id": msg.ID,
		"content":    msg.Content,
	}).Info("Processing system message")
	
	// System messages don't typically require responses
	return "System message processed", nil
}

// sendResponse sends a response message
func (mw *MessageWorker) sendResponse(ctx context.Context, deviceID, phoneNumber, content string) error {
	// Use the queue service to send the message
	if mw.processor.queueService != nil {
		return mw.processor.queueService.QueueMessage(phoneNumber, content)
	}
	
	// Fallback: log the response (in a real implementation, this would send via WhatsApp API)
	mw.logger.WithFields(logrus.Fields{
		"device_id":    deviceID,
		"phone_number": phoneNumber,
		"content":      content,
	}).Info("Sending response message")
	
	return nil
}

// notifyStaff notifies staff about new customer messages
func (mw *MessageWorker) notifyStaff(ctx context.Context, msg *IncomingMessage) error {
	// In a real implementation, this would notify staff via WebSocket, email, etc.
	mw.logger.WithFields(logrus.Fields{
		"phone_number": msg.PhoneNumber,
		"content":      msg.Content,
	}).Info("Notifying staff about customer message")
	
	return nil
}