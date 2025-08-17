import re

# Read the file
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix 1: loadLeads function declaration
content = re.sub(r'// Load leads\s*\n\s*function loadLeads\s*\n\s*const url', 
                 '// Load leads\n        function loadLeads() {\n            const url', content)

# Fix 2: Fix the leads response handling to match the API structure
content = re.sub(r'if \(data\.leads && Array\.isArray\(data\.leads\)\) \{',
                 'if (data.code === "SUCCESS" && data.results && data.results.leads) {', content)

content = re.sub(r'leads = data\.leads;',
                 'leads = data.results.leads;', content)

content = re.sub(r'document\.getElementById\(\'totalLeads\'\)\.textContent = leads\.length;',
                 'document.getElementById(\'totalLeads\').textContent = data.results.pagination ? data.results.pagination.total : leads.length;', content)

# Fix 3: Close the loadLeads function
content = re.sub(r'(\s*displayLeads\(\);\s*\}\s*\}\)\s*\.catch\(error => \{[^}]*\}\);)',
                 r'\1\n        }', content)

# Fix 4: Fix function calls that are missing parentheses
content = re.sub(r'loadLeads\s*;', 'loadLeads();', content)
content = re.sub(r'loadLeads\s*\n', 'loadLeads();\n', content)

# Fix 5: Fix any remaining standalone function names
content = re.sub(r'function buildNicheFilters\s*\n', 'function buildNicheFilters() {\n', content)
content = re.sub(r'function getFilteredLeads\s*\n', 'function getFilteredLeads() {\n', content)
content = re.sub(r'function createLeadCard\s*\n', 'function createLeadCard(lead) {\n', content)

# Write back
with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Applied minimal fixes to make the page work!")
