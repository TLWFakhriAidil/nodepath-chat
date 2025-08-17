import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Fix 1: Replace all lead.status with lead.target_status
content = re.sub(r'lead\.status', 'lead.target_status', content)
content = re.sub(r'lead\.target_status \|\| lead\.target_status', 'lead.target_status || lead.status', content)  # Fix double replacement

# Fix 2: Replace .lead-status CSS class with .lead-target_status
content = re.sub(r'\.lead-status', '.lead-target_status', content)
content = re.sub(r'<span class="lead-status', '<span class="lead-target_status', content)

# Fix 3: Fix the Select All checkbox - it says "Hide Select All for public view"
content = content.replace(
    '<!-- Hide Select All for public view -->',
    '''<div class="form-check">
                        <input class="form-check-input" type="checkbox" id="selectAll" onchange="toggleSelectAll()">
                        <label class="form-check-label" for="selectAll">
                            Select All
                        </label>
                    </div>'''
)

# Fix 4: Update all save functions to use target_status
# Already done in saveLead, but need to ensure consistency

# Fix 5: Fix CSV export to use target_status correctly
content = re.sub(
    r"lead\.target_status \|\| lead\.status \|\| ''\s*\n\s*lead\.trigger",
    "lead.target_status || lead.status || '',\n                lead.trigger",
    content
)

# Fix 6: Fix CSV import to use target_status
content = re.sub(
    r"target_status: lead\.target_status \|\| lead\.status \|\| 'prospect'\s*\n\s*trigger:",
    "target_status: lead.target_status || lead.status || 'prospect',\n                    trigger:",
    content
)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed all status references and Select All checkbox!")
