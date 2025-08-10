import re

# Read the ultra_scale_broadcast_manager.go file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\broadcast\ultra_scale_broadcast_manager.go', 'r') as f:
    content = f.read()

# Update the QueueMessageToBroadcast function to check sequence_stepid
old_logic = r'''// Determine which pool this message belongs to
	if msg\.CampaignID != nil \{
		poolKey = fmt\.Sprintf\("campaign:%d", \*msg\.CampaignID\)
	\} else if msg\.SequenceID != nil \{
		poolKey = fmt\.Sprintf\("sequence:%s", \*msg\.SequenceID\)
	\} else \{
		return fmt\.Errorf\("message has no campaign or sequence ID"\)
	\}'''

new_logic = '''// Determine which pool this message belongs to
	if msg.CampaignID != nil {
		poolKey = fmt.Sprintf("campaign:%d", *msg.CampaignID)
	} else if msg.SequenceStepID != nil {
		// Use sequence_stepid for sequences instead of sequence_id
		poolKey = fmt.Sprintf("sequence:step:%s", *msg.SequenceStepID)
	} else if msg.SequenceID != nil {
		// Fallback to sequence_id if no step_id
		poolKey = fmt.Sprintf("sequence:%s", *msg.SequenceID)
	} else {
		return fmt.Errorf("message has no campaign ID or sequence step ID")
	}'''

content = re.sub(old_logic, new_logic, content, flags=re.DOTALL)

# Also update the pool creation logic
old_pool_logic = r'''broadcastType := "campaign"
		broadcastID := ""
		if msg\.CampaignID != nil \{
			broadcastID = fmt\.Sprintf\("%d", \*msg\.CampaignID\)
		\} else if msg\.SequenceID != nil \{
			broadcastType = "sequence"
			broadcastID = \*msg\.SequenceID
		\}'''

new_pool_logic = '''broadcastType := "campaign"
		broadcastID := ""
		if msg.CampaignID != nil {
			broadcastID = fmt.Sprintf("%d", *msg.CampaignID)
		} else if msg.SequenceStepID != nil {
			broadcastType = "sequence"
			broadcastID = *msg.SequenceStepID
		} else if msg.SequenceID != nil {
			broadcastType = "sequence"
			broadcastID = *msg.SequenceID
		}'''

content = re.sub(old_pool_logic, new_pool_logic, content, flags=re.DOTALL)

# Write the updated content
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\broadcast\ultra_scale_broadcast_manager.go', 'w') as f:
    f.write(content)

print("Successfully updated broadcast manager to use sequence_stepid!")
