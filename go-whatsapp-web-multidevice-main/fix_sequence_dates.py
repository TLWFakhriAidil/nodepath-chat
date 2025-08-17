# Add comment about unused date filters
import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Find the GetSequenceSummary function and add comment
pattern = r'(// Get date filters - handle both formats\s+startDate := c\.Query\("start_date", c\.Query\("start", ""\)\)\s+endDate := c\.Query\("end_date", c\.Query\("end", ""\)\))'
replacement = r'\1\n\t// TODO: Add date filtering to sequence summary query if needed\n\t_ = startDate // Currently unused\n\t_ = endDate   // Currently unused'

content = re.sub(pattern, replacement, content, flags=re.MULTILINE)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Added comments for unused variables!")
