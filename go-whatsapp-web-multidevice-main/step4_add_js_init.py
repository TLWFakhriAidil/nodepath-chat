# Step 4: Add missing JavaScript functions
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find where to add the missing variables and functions
# Add after "let currentTriggerFilter = 'all';"
add_after_variables = "let currentTriggerFilter = 'all';"
new_variables = """let currentTriggerFilter = 'all';
        let leadModal;
        let importModal;
        let bulkUpdateModal;
        let selectedLeads = new Set();"""

content = content.replace(add_after_variables, new_variables)

# Add modal initialization after DOMContentLoaded
add_after_dom = "document.getElementById('deviceId').textContent = deviceId;"
new_init = """document.getElementById('deviceId').textContent = deviceId;
            
            // Initialize modals
            leadModal = new bootstrap.Modal(document.getElementById('leadModal'));
            importModal = new bootstrap.Modal(document.getElementById('importModal'));
            bulkUpdateModal = new bootstrap.Modal(document.getElementById('bulkUpdateModal'));"""

content = content.replace(add_after_dom, new_init)

# Update createLeadCard to include checkbox and make name clickable
old_card = """return `
                <div class="lead-card" data-lead-id="${lead.id}">
                    <div class="d-flex align-items-start">
                        <div class="lead-card-content">
                            <div class="flex-grow-1">
                                <h5 class="mb-1">${lead.name}</h5>"""

new_card = """return `
                <div class="lead-card \${selectedLeads.has(lead.id) ? 'selected' : ''}" data-lead-id="\${lead.id}">
                    <div class="d-flex align-items-start">
                        <input type="checkbox" class="form-check-input lead-checkbox" 
                               value="\${lead.id}" 
                               \${selectedLeads.has(lead.id) ? 'checked' : ''}
                               onchange="toggleLeadSelection('\${lead.id}')"
                               onclick="event.stopPropagation()">
                        <div class="lead-card-content" onclick="editLead('\${lead.id}')">
                            <div class="flex-grow-1">
                                <h5 class="mb-1" style="cursor: pointer; color: var(--primary);">\${lead.name}</h5>"""

content = content.replace(old_card, new_card)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Step 4: Added JavaScript variables and modal initialization!")
