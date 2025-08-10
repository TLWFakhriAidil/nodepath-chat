import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\campaign_trigger.go', 'r') as f:
    content = f.read()

# Find and comment out the ProcessSequenceTriggers section
lines = content.split('\n')
new_lines = []
in_sequence_section = False
comment_count = 0

for i, line in enumerate(lines):
    if 'Process sequence triggers for new leads' in line:
        in_sequence_section = True
        new_lines.append('\t\t\t// REMOVED: ProcessSequenceTriggers - now handled by dedicated processor')
        new_lines.append('\t\t\t// This was causing duplicate message creation!')
    
    if in_sequence_section:
        if 'ProcessSequenceTriggers' in line or 'Error processing sequence triggers' in line or ('}' in line and comment_count < 3):
            new_lines.append('\t\t\t// ' + line.strip())
            comment_count += 1
            if '}' in line:
                in_sequence_section = False
        else:
            new_lines.append(line)
    else:
        new_lines.append(line)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\campaign_trigger.go', 'w') as f:
    f.write('\n'.join(new_lines))

print('Successfully commented out ProcessSequenceTriggers in campaign_trigger.go')
