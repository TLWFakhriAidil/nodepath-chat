# Fix the loadLeads function to load all leads and use target_status
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Update loadLeads to load all leads with proper pagination
old_load_leads = r'function loadLeads\(\) \{[\s\S]*?\}[\s\S]*?\.catch\(error => \{[\s\S]*?\}\);[\s\S]*?\}'

new_load_leads = '''function loadLeads() {
            // Load all leads by default
            const url = `/api/public/device/${deviceId}/leads?per_page=1000`;
                
            fetch(url)
                .then(response => response.json())
                .then(data => {
                    // Handle the response format from public API
                    if (data.code === 'SUCCESS' && data.results && data.results.leads && Array.isArray(data.results.leads)) {
                        leads = data.results.leads;
                        document.getElementById('totalLeads').textContent = data.results.pagination ? data.results.pagination.total : leads.length;
                    } else if (data.leads && Array.isArray(data.leads)) {
                        leads = data.leads;
                        document.getElementById('totalLeads').textContent = leads.length;
                    }
                    buildNicheFilters();
                    displayLeads();
                    updateSelectAllCheckbox();
                })
                .catch(error => {
                    console.error('Error loading leads:', error);
                    leads = [];
                    displayLeads();
                });
        }'''

content = re.sub(old_load_leads, new_load_leads, content, flags=re.DOTALL)

# Fix the saveLead function to use target_status
content = re.sub(
    r'status: document\.getElementById\(\'leadStatus\'\)\.value',
    'target_status: document.getElementById(\'leadStatus\').value',
    content
)

# Fix the saveBulkUpdate to include name and use target_status
old_bulk_update = r'const updates = \{[\s\S]*?\};'
new_bulk_update = '''const updates = {
                name: document.getElementById('bulkName').value,
                niche: document.getElementById('bulkNiche').value,
                target_status: document.getElementById('bulkStatus').value,
                trigger: document.getElementById('bulkTrigger').value
            };'''

content = re.sub(old_bulk_update, new_bulk_update, content)

# Update the bulk update data to use target_status
content = re.sub(
    r'status: updates\.status \|\| lead\.target_status \|\| lead\.status \|\| \'prospect\'',
    'target_status: updates.target_status || lead.target_status || lead.status || \'prospect\'',
    content
)

# Also update CSV import to use target_status
content = re.sub(
    r'status: lead\.target_status \|\| lead\.status \|\| \'prospect\'',
    'target_status: lead.target_status || lead.status || \'prospect\'',
    content
)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed data loading and status fields!")
