import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# The issue is with template literals inside the createLeadCard function
# We need to build the HTML string differently to avoid Go template parsing issues

# Find the createLeadCard function and replace it
old_function = r'''function createLeadCard\(lead\) \{
        const lastInteraction = lead\.last_interaction
            \? new Date\(lead\.last_interaction\)\.toLocaleDateString\(\)
            : 'Never';
        
        return `
            <div class="lead-card \$\{selectedLeads\.has\(lead\.id\) \? 'selected' : ''\}" data-lead-id="\$\{lead\.id\}">
                <div class="d-flex align-items-start">
                    <input type="checkbox" class="form-check-input lead-checkbox" 
                           value="\$\{lead\.id\}" 
                           \$\{selectedLeads\.has\(lead\.id\) \? 'checked' : ''\}
                           onchange="toggleLeadSelection\('\$\{lead\.id\}'\)"
                           onclick="event\.stopPropagation\(\)">
                    <div class="lead-card-content" onclick="editLead\('\$\{lead\.id\}'\)">
                        <div class="flex-grow-1">
                            <h5 class="mb-1" style="cursor: pointer; color: var\(--primary\);">\$\{lead\.name\}</h5>
                            <p class="mb-2">
                                <i class="bi bi-telephone text-muted"></i> \$\{lead\.phone\}
                                \$\{lead\.niche \? `<span class="ms-3"><i class="bi bi-briefcase text-muted"></i> \$\{lead\.niche\}</span>` : ''\}
                            </p>
                            <div class="d-flex align-items-center">
                                <span class="lead-status \$\{lead\.target_status \|\| lead\.status\}">\$\{\(lead\.target_status \|\| lead\.status \|\| ''\)\.toUpperCase\(\)\}</span>
                                <span class="text-muted small ms-3">Last interaction: \$\{lastInteraction\}</span>
                            </div>
                            \$\{lead\.journey \? `<div class="lead-journey">\$\{lead\.journey\}</div>` : ''\}
                            \$\{lead\.trigger \? `<div class="mt-2"><span class="badge bg-info"><i class="bi bi-lightning"></i> Trigger: \$\{lead\.trigger\}</span></div>` : ''\}
                        </div>
                    </div>
                </div>
            </div>
        `;
    \}'''

# Use string concatenation instead of template literals
new_function = '''function createLeadCard(lead) {
        const lastInteraction = lead.last_interaction
            ? new Date(lead.last_interaction).toLocaleDateString()
            : 'Never';
        
        let html = '<div class="lead-card';
        if (selectedLeads.has(lead.id)) html += ' selected';
        html += '" data-lead-id="' + lead.id + '">';
        html += '<div class="d-flex align-items-start">';
        html += '<input type="checkbox" class="form-check-input lead-checkbox" value="' + lead.id + '"';
        if (selectedLeads.has(lead.id)) html += ' checked';
        html += ' onchange="toggleLeadSelection(\'' + lead.id + '\')" onclick="event.stopPropagation()">';
        html += '<div class="lead-card-content" onclick="editLead(\'' + lead.id + '\')">';
        html += '<div class="flex-grow-1">';
        html += '<h5 class="mb-1" style="cursor: pointer; color: var(--primary);">' + lead.name + '</h5>';
        html += '<p class="mb-2">';
        html += '<i class="bi bi-telephone text-muted"></i> ' + lead.phone;
        if (lead.niche) {
            html += '<span class="ms-3"><i class="bi bi-briefcase text-muted"></i> ' + lead.niche + '</span>';
        }
        html += '</p>';
        html += '<div class="d-flex align-items-center">';
        html += '<span class="lead-status ' + (lead.target_status || lead.status) + '">' + (lead.target_status || lead.status || '').toUpperCase() + '</span>';
        html += '<span class="text-muted small ms-3">Last interaction: ' + lastInteraction + '</span>';
        html += '</div>';
        if (lead.journey) {
            html += '<div class="lead-journey">' + lead.journey + '</div>';
        }
        if (lead.trigger) {
            html += '<div class="mt-2"><span class="badge bg-info"><i class="bi bi-lightning"></i> Trigger: ' + lead.trigger + '</span></div>';
        }
        html += '</div></div></div></div>';
        
        return html;
    }'''

# First try to find with escaped $
pattern1 = r'function createLeadCard\(lead\) \{[^}]+return `[^`]+`;\s*\}'
if re.search(pattern1, content, re.DOTALL):
    content = re.sub(pattern1, new_function, content, flags=re.DOTALL)
else:
    # Try without escaped $
    pattern2 = r'function createLeadCard\(lead\) \{[^}]+return `[^`]+`;\s*\}'
    content = re.sub(pattern2, new_function, content, flags=re.DOTALL)

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed createLeadCard to use string concatenation!")
