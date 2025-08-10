import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\sequence_trigger_processor.go', 'r') as f:
    content = f.read()

# Remove uuid import since we're using repository method now
content = content.replace('\n\t"github.com/google/uuid"', '')

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase\sequence_trigger_processor.go', 'w') as f:
    f.write(content)

print("Removed unused uuid import")
