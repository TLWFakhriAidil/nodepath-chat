import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix line 622 - missing function call
content = content.replace(
    'emptyState.style.display = \'none\';\n            leadsList.innerHTML = filteredLeads.map(lead => createLeadCard(lead)).join(\'\');\n        }',
    'emptyState.style.display = \'none\';\n            leadsList.innerHTML = filteredLeads.map(lead => createLeadCard(lead)).join(\'\');\n        }'
)

# Fix line 701 - missing function declaration
content = content.replace(
    '// Create lead card HTML (simplified for public view)\n        function createLeadCard(lead) {',
    '// Create lead card HTML (simplified for public view)\n        function createLeadCard(lead) {'
)

# Also ensure the function call on line 622 is correct
content = re.sub(
    r'createLeadCard\s*\n',
    'leadsList.innerHTML = filteredLeads.map(lead => createLeadCard(lead)).join(\'\');\n',
    content
)

# And fix the function definition
content = re.sub(
    r'// Create lead card HTML \(simplified for public view\)\s*\n\s*createLeadCard\s*\n',
    '// Create lead card HTML (simplified for public view)\n        function createLeadCard(lead) {\n',
    content
)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed function definitions!")
