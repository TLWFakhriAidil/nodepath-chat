// Update sendTextMessage in WhatsAppMessageSender to process spintax

func (w *WhatsAppMessageSender) sendTextMessage(waClient *whatsmeow.Client, recipient types.JID, msg *broadcast.BroadcastMessage) error {
	// Process content with spintax and greeting
	processedContent := w.greetingProcessor.PrepareMessageWithGreeting(
		msg.Content,           // Use Content field (same as Message)
		msg.RecipientName,
		msg.DeviceID,
		msg.RecipientPhone,
	)
	
	// Apply message randomization (10% variation instead of 20%)
	finalContent := w.messageRandomizer.RandomizeMessage(processedContent)
	
	// Fix line breaks for WhatsApp
	// Replace \n with %0A for proper WhatsApp formatting
	finalContent = strings.ReplaceAll(finalContent, "\n", "%0A")
	
	// Create message
	message := &waE2E.Message{
		Conversation: proto.String(finalContent),
	}
	
	// Send message
	resp, err := waClient.SendMessage(context.Background(), recipient, message)
	if err != nil {
		return fmt.Errorf("failed to send text message: %v", err)
	}
	
	logrus.Infof("Text message sent to %s (ID: %s)", recipient.String(), resp.ID)
	return nil
}

// Also update sendImageMessage to process caption with spintax

func (w *WhatsAppMessageSender) sendImageMessage(waClient *whatsmeow.Client, recipient types.JID, msg *broadcast.BroadcastMessage) error {
	// ... existing image download/upload code ...
	
	// Process caption with spintax and greeting
	caption := msg.Message
	if caption != "" {
		processedCaption := w.greetingProcessor.PrepareMessageWithGreeting(
			caption,
			msg.RecipientName,
			msg.DeviceID,
			msg.RecipientPhone,
		)
		
		// Apply message randomization
		caption = w.messageRandomizer.RandomizeMessage(processedCaption)
		
		// Fix line breaks for WhatsApp
		caption = strings.ReplaceAll(caption, "\n", "%0A")
	}
	
	// Create image message
	message := &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       proto.String(caption),
			// ... rest of image fields ...
		},
	}
	
	// ... rest of sending logic ...
}