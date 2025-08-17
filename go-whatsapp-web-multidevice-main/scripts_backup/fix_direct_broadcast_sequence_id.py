import re

# Fix the direct_broadcast_processor to ensure sequence_id is always populated
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\direct_broadcast_processor.go', 'r') as f:
    content = f.read()

# Add a variable to track the original sequence ID
old_vars = 'currentSequenceID := sequenceID\n\tprocessedSequences := make(map[string]bool)'
new_vars = '''currentSequenceID := sequenceID
	originalSequenceID := sequenceID  // Keep track of the original sequence
	processedSequences := make(map[string]bool)'''

content = content.replace(old_vars, new_vars)

# Update the message creation to use originalSequenceID if currentSequenceID is empty
old_msg_creation = '''// Create broadcast message with proper sequence references
			msg := domainBroadcast.BroadcastMessage{
				UserID:         lead.UserID,
				DeviceID:       lead.DeviceID,
				SequenceID:     &currentSequenceID,'''

new_msg_creation = '''// Create broadcast message with proper sequence references
			// Use current sequence ID, but fallback to original if empty
			sequenceIDToUse := currentSequenceID
			if sequenceIDToUse == "" {
				sequenceIDToUse = originalSequenceID
			}
			
			msg := domainBroadcast.BroadcastMessage{
				UserID:         lead.UserID,
				DeviceID:       lead.DeviceID,
				SequenceID:     &sequenceIDToUse,'''

content = content.replace(old_msg_creation, new_msg_creation)

# Update debug log
old_debug = '''logrus.Debugf("Queueing message - UserID: '%s', DeviceID: '%s', SequenceID: '%s', StepID: '%s'", 
				msg.UserID, msg.DeviceID, *msg.SequenceID, *msg.SequenceStepID)'''

new_debug = '''logrus.Debugf("Queueing message - UserID: '%s', DeviceID: '%s', SequenceID: '%s', StepID: '%s'", 
				msg.UserID, msg.DeviceID, *msg.SequenceID, *msg.SequenceStepID)
			
			// Validate both IDs are set
			if *msg.SequenceID == "" || *msg.SequenceStepID == "" {
				logrus.Errorf("WARNING: Creating message with empty IDs - SequenceID: '%s', StepID: '%s'", 
					*msg.SequenceID, *msg.SequenceStepID)
			}'''

content = content.replace(old_debug, new_debug)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\direct_broadcast_processor.go', 'w') as f:
    f.write(content)

print("Fixed direct_broadcast_processor to always populate sequence_id!")
