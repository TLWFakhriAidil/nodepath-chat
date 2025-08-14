import re

# Fix the QueueMessage method to properly handle empty string pointers
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'r') as f:
    content = f.read()

# Find the QueueMessage method and fix the nullable field handling
# The issue is when msg.SequenceID points to an empty string
old_pattern = r'var sequenceID interface{}\s*\n\s*if msg\.SequenceID != nil \{\s*\n\s*sequenceID = \*msg\.SequenceID'
new_code = '''var sequenceID interface{}
	if msg.SequenceID != nil && *msg.SequenceID != "" {
		sequenceID = *msg.SequenceID'''

content = re.sub(old_pattern, new_code, content, flags=re.MULTILINE)

# Do the same for sequenceStepID
old_pattern2 = r'var sequenceStepID interface{}\s*\n\s*if msg\.SequenceStepID != nil \{\s*\n\s*sequenceStepID = \*msg\.SequenceStepID'
new_code2 = '''var sequenceStepID interface{}
	if msg.SequenceStepID != nil && *msg.SequenceStepID != "" {
		sequenceStepID = *msg.SequenceStepID'''

content = re.sub(old_pattern2, new_code2, content, flags=re.MULTILINE)

# Do the same for groupID
old_pattern3 = r'var groupID interface{}\s*\n\s*if msg\.GroupID != nil \{\s*\n\s*groupID = \*msg\.GroupID'
new_code3 = '''var groupID interface{}
	if msg.GroupID != nil && *msg.GroupID != "" {
		groupID = *msg.GroupID'''

content = re.sub(old_pattern3, new_code3, content, flags=re.MULTILINE)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\broadcast_repository.go', 'w') as f:
    f.write(content)

print("Fixed QueueMessage to handle empty string pointers")
