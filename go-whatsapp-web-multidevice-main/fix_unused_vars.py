# Quick fix for unused variables
import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Replace unused variables in GetSequenceSummary
content = re.sub(
    r'startDate := c\.Query\("start_date", c\.Query\("start", ""\)\)',
    '_ = c.Query("start_date", c.Query("start", ""))',
    content
)
content = re.sub(
    r'endDate := c\.Query\("end_date", c\.Query\("end", ""\)\)',
    '_ = c.Query("end_date", c.Query("end", ""))',
    content
)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed unused variables!")
