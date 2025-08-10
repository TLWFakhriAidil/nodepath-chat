import re

# Read the file
with open('src/ui/rest/app.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Find GetSequenceDeviceLeads function and remove the misplaced date filter
lines = content.split('\n')
new_lines = []
in_sequence_device_leads = False
found_date_filter = False
skip_lines = 0

for i, line in enumerate(lines):
    # Check if we're in GetSequenceDeviceLeads function
    if 'func (handler *App) GetSequenceDeviceLeads' in line:
        in_sequence_device_leads = True
    elif in_sequence_device_leads and ('func (handler *App)' in line or 'func (' in line):
        in_sequence_device_leads = False
    
    # If we're in the function and find the misplaced date filter
    if in_sequence_device_leads and '// Add date filter if provided' in line and i > 4840:
        found_date_filter = True
        skip_lines = 11  # Skip this line and next 10 lines
        continue
    
    if skip_lines > 0:
        skip_lines -= 1
        continue
        
    new_lines.append(line)

# Write back
with open('src/ui/rest/app.go', 'w', encoding='utf-8') as f:
    f.write('\n'.join(new_lines))

print(f"Removed misplaced date filter from GetSequenceDeviceLeads function")
