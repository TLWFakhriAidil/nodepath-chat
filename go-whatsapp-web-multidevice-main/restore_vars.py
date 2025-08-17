# Restore the variables
import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Restore the variables in GetSequenceSummary
content = re.sub(
    r'_ = c\.Query\("start_date", c\.Query\("start", ""\)\)',
    'startDate := c.Query("start_date", c.Query("start", ""))',
    content
)
content = re.sub(
    r'_ = c\.Query\("end_date", c\.Query\("end", ""\)\)',
    'endDate := c.Query("end_date", c.Query("end", ""))',
    content
)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\public_device_routes.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Restored variables!")
