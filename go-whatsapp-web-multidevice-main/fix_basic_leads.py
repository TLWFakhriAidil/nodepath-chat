# First, let's just fix the basic issues with the current file
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# 1. Fix the loadLeads function to handle the correct API response format
content = content.replace(
    'if (data.leads && Array.isArray(data.leads)) {',
    'if (data.code === \'SUCCESS\' && data.results && data.results.leads && Array.isArray(data.results.leads)) {\n                        leads = data.results.leads;\n                        document.getElementById(\'totalLeads\').textContent = data.results.pagination ? data.results.pagination.total : leads.length;\n                    } else if (data.leads && Array.isArray(data.leads)) {'
)

# 2. Remove credentials from public API calls
content = content.replace(
    "fetch(`/api/public/device/${deviceId}/info`, { credentials: 'include' })",
    "fetch(`/api/public/device/${deviceId}/info`)"
)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed basic API response handling!")
