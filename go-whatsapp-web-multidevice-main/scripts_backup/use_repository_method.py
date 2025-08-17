import re

# Replace direct SQL insert with repository method
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\sequence_trigger_processor.go', 'r') as f:
    content = f.read()

# Add repository import if not present
if '"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"' not in content:
    # Add after other imports
    content = content.replace(
        '"github.com/sirupsen/logrus"',
        '"github.com/sirupsen/logrus"\n\t"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"'
    )

# Replace the insert loop
old_insert = '''	// Insert all messages into broadcast_messages
	for _, msg := range allMessages {
		// Generate UUID for message ID
		messageID := uuid.New().String()
		
		insertQuery := `
			INSERT INTO broadcast_messages (
				id, user_id, device_id, sequence_id, sequence_stepid,
				recipient_phone, recipient_name, message_type,
				content, media_url, status, scheduled_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		
		// Handle potential nil values for media_url
		var mediaURL interface{} = nil
		if msg.MediaURL != "" {
			mediaURL = msg.MediaURL
		}
		
		_, err = tx.Exec(insertQuery,
			messageID, msg.UserID, msg.DeviceID, msg.SequenceID, msg.SequenceStepID,
			msg.RecipientPhone, msg.RecipientName, msg.Type,
			msg.Content, mediaURL, msg.Status, msg.ScheduledAt)
		
		if err != nil {
			logrus.Errorf("Failed to insert broadcast message: %v", err)
			return fmt.Errorf("failed to insert broadcast message: %w", err)
		}
	}'''

new_insert = '''	// Get broadcast repository
	broadcastRepo := repository.GetBroadcastRepository()
	
	// Insert all messages using repository (which handles UUIDs properly)
	for _, msg := range allMessages {
		// The repository QueueMessage will handle ID generation and null values
		err := broadcastRepo.QueueMessage(msg)
		if err != nil {
			logrus.Errorf("Failed to queue broadcast message for %s: %v", msg.RecipientPhone, err)
			return fmt.Errorf("failed to queue broadcast message: %w", err)
		}
	}'''

content = content.replace(old_insert, new_insert)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\sequence_trigger_processor.go', 'w') as f:
    f.write(content)

print("Fixed - now using repository method which handles UUIDs properly")
