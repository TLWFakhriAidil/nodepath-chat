# Step 1: Add Import/Export buttons and fix Add Lead button
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find and replace the button section
old_buttons = r'<div>\s*<!-- Hide buttons for public view -->\s*</div>'
new_buttons = '''<div>
                    <button class="btn btn-outline-primary me-2" onclick="exportLeads()">
                        <i class="bi bi-download"></i> Export
                    </button>
                    <button class="btn btn-outline-primary me-2" onclick="document.getElementById('importFile').click()">
                        <i class="bi bi-upload"></i> Import
                    </button>
                    <input type="file" id="importFile" accept=".csv" style="display: none;" onchange="importLeads(this)">
                    <button class="add-lead-btn" onclick="openAddLeadModal()">
                        <i class="bi bi-plus-circle"></i> Add Lead
                    </button>
                </div>'''

content = re.sub(old_buttons, new_buttons, content)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Step 1: Added Import/Export buttons!")
