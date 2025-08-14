import re

# Fix the direct_broadcast_processor to properly set sequence_id and sequence_stepid
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\direct_broadcast_processor.go', 'r') as f:
    content = f.read()

# Find the section where we create broadcast message and fix it
old_code = '''			// Create broadcast message WITHOUT SequenceStepID to avoid UUID errors
			msg := domainBroadcast.BroadcastMessage{
				UserID:         lead.UserID,
				DeviceID:       lead.DeviceID,
				SequenceID:     &currentSequenceID,
				// Don't set SequenceStepID - it's causing UUID errors
				RecipientPhone: lead.Phone,
				RecipientName:  lead.Name,
				Message:        step.Content,
				Content:        step.Content,
				Type:           step.MessageType,
				MinDelay:       step.MinDelay,
				MaxDelay:       step.MaxDelay,
				ScheduledAt:    scheduledAt,
				Status:         "pending",
			}'''

new_code = '''			// Validate sequence ID is not empty
			if currentSequenceID == "" {
				logrus.Errorf("Current sequence ID is empty - skipping message creation")
				continue
			}

			// Create broadcast message with proper sequence references
			msg := domainBroadcast.BroadcastMessage{
				UserID:         lead.UserID,
				DeviceID:       lead.DeviceID,
				SequenceID:     &currentSequenceID,
				SequenceStepID: &step.ID,  // Re-enable this since we validated step.ID above
				RecipientPhone: lead.Phone,
				RecipientName:  lead.Name,
				Message:        step.Content,
				Content:        step.Content,
				Type:           step.MessageType,
				MinDelay:       step.MinDelay,
				MaxDelay:       step.MaxDelay,
				ScheduledAt:    scheduledAt,
				Status:         "pending",
			}'''

content = content.replace(old_code, new_code)

# Also update the debug log
old_debug = '''logrus.Debugf("Queueing message - UserID: '%s', DeviceID: '%s', SequenceID: '%s'", 
				msg.UserID, msg.DeviceID, *msg.SequenceID)'''

new_debug = '''logrus.Debugf("Queueing message - UserID: '%s', DeviceID: '%s', SequenceID: '%s', StepID: '%s'", 
				msg.UserID, msg.DeviceID, *msg.SequenceID, *msg.SequenceStepID)'''

content = content.replace(old_debug, new_debug)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\direct_broadcast_processor.go', 'w') as f:
    f.write(content)

print("Fixed direct_broadcast_processor to properly set sequence_id and sequence_stepid")
