# Insert all missing functions before the closing script tag
import re

with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'r', encoding='utf-8') as f:
    content = f.read()

# Find where to insert - before the closing script tag
insert_position = content.rfind('    </script>')

if insert_position > 0:
    # All the missing functions
    missing_functions = '''
        // Toggle select all
        function toggleSelectAll() {
            const selectAll = document.getElementById('selectAll');
            const checkboxes = document.querySelectorAll('.lead-checkbox');
            
            if (selectAll.checked) {
                checkboxes.forEach(cb => {
                    cb.checked = true;
                    selectedLeads.add(cb.value);
                });
            } else {
                checkboxes.forEach(cb => {
                    cb.checked = false;
                });
                selectedLeads.clear();
            }
            
            updateBulkActionsVisibility();
            updateSelectedCount();
        }
        
        // Toggle individual lead selection
        function toggleLeadSelection(leadId) {
            if (selectedLeads.has(leadId)) {
                selectedLeads.delete(leadId);
            } else {
                selectedLeads.add(leadId);
            }
            
            updateBulkActionsVisibility();
            updateSelectedCount();
            updateSelectAllCheckbox();
        }
        
        // Update bulk actions visibility
        function updateBulkActionsVisibility() {
            const bulkActions = document.getElementById('bulkActions');
            if (selectedLeads.size > 0) {
                bulkActions.style.display = 'flex';
            } else {
                bulkActions.style.display = 'none';
            }
        }
        
        // Update selected count
        function updateSelectedCount() {
            const element = document.getElementById('selectedCount');
            if (element) element.textContent = selectedLeads.size;
            const bulkElement = document.getElementById('bulkUpdateCount');
            if (bulkElement) bulkElement.textContent = selectedLeads.size;
        }
        
        // Update select all checkbox
        function updateSelectAllCheckbox() {
            const selectAll = document.getElementById('selectAll');
            const checkboxes = document.querySelectorAll('.lead-checkbox');
            
            if (!selectAll) return;
            
            if (checkboxes.length === 0) {
                selectAll.checked = false;
                selectAll.indeterminate = false;
            } else if (selectedLeads.size === 0) {
                selectAll.checked = false;
                selectAll.indeterminate = false;
            } else if (selectedLeads.size === checkboxes.length) {
                selectAll.checked = true;
                selectAll.indeterminate = false;
            } else {
                selectAll.checked = false;
                selectAll.indeterminate = true;
            }
        }
        
        // Open add lead modal
        function openAddLeadModal() {
            document.getElementById('leadModalTitle').textContent = 'Add New Lead';
            document.getElementById('leadForm').reset();
            document.getElementById('leadId').value = '';
            document.getElementById('leadStatus').value = 'prospect';
            leadModal.show();
        }
        
        // Edit lead
        function editLead(leadId) {
            const lead = leads.find(l => l.id === leadId);
            if (lead) {
                document.getElementById('leadModalTitle').textContent = 'Edit Lead';
                document.getElementById('leadId').value = lead.id;
                document.getElementById('leadName').value = lead.name;
                document.getElementById('leadPhone').value = lead.phone;
                document.getElementById('leadNiche').value = lead.niche || '';
                document.getElementById('leadStatus').value = lead.target_status || lead.status || 'prospect';
                document.getElementById('leadTrigger').value = lead.trigger || '';
                document.getElementById('leadJourney').value = lead.journey || '';
                leadModal.show();
            }
        }
        
        // Save lead
        function saveLead() {
            const leadId = document.getElementById('leadId').value;
            const leadData = {
                device_id: deviceId,
                name: document.getElementById('leadName').value,
                phone: document.getElementById('leadPhone').value,
                niche: document.getElementById('leadNiche').value,
                status: document.getElementById('leadStatus').value,
                trigger: document.getElementById('leadTrigger').value,
                journey: document.getElementById('leadJourney').value
            };
            
            const url = leadId 
                ? `/api/public/device/${deviceId}/lead/${leadId}`
                : `/api/public/device/${deviceId}/lead`;
            const method = leadId ? 'PUT' : 'POST';
            
            fetch(url, {
                method: method,
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(leadData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    leadModal.hide();
                    loadLeads();
                    showAlert('success', leadId ? 'Lead updated successfully!' : 'Lead added successfully!');
                } else {
                    showAlert('danger', data.message || 'Failed to save lead');
                }
            })
            .catch(error => {
                showAlert('danger', 'Error: ' + error.message);
            });
        }
        
        // Bulk delete
        function bulkDelete() {
            if (selectedLeads.size === 0) return;
            
            if (confirm(`Are you sure you want to delete ${selectedLeads.size} lead(s)?`)) {
                const promises = Array.from(selectedLeads).map(leadId => 
                    fetch(`/api/public/device/${deviceId}/lead/${leadId}`, {
                        method: 'DELETE'
                    })
                );
                
                Promise.all(promises)
                    .then(() => {
                        selectedLeads.clear();
                        loadLeads();
                        showAlert('success', 'Selected leads deleted successfully!');
                    })
                    .catch(error => {
                        showAlert('danger', 'Error deleting leads: ' + error.message);
                    });
            }
        }
        
        // Open bulk update modal
        function bulkUpdate() {
            if (selectedLeads.size === 0) return;
            document.getElementById('bulkUpdateForm').reset();
            bulkUpdateModal.show();
        }
        
        // Save bulk update
        function saveBulkUpdate() {
            const updates = {
                niche: document.getElementById('bulkNiche').value,
                status: document.getElementById('bulkStatus').value,
                trigger: document.getElementById('bulkTrigger').value
            };
            
            const promises = Array.from(selectedLeads).map(leadId => {
                const lead = leads.find(l => l.id === leadId);
                if (lead) {
                    const updateData = {
                        device_id: deviceId,
                        name: lead.name,
                        phone: lead.phone,
                        niche: updates.niche || lead.niche || '',
                        status: updates.status || lead.target_status || lead.status || 'prospect',
                        trigger: updates.trigger || lead.trigger || '',
                        journey: lead.journey || ''
                    };
                    
                    return fetch(`/api/public/device/${deviceId}/lead/${leadId}`, {
                        method: 'PUT',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify(updateData)
                    });
                }
            });
            
            Promise.all(promises)
                .then(() => {
                    bulkUpdateModal.hide();
                    selectedLeads.clear();
                    loadLeads();
                    showAlert('success', 'Selected leads updated successfully!');
                })
                .catch(error => {
                    showAlert('danger', 'Error updating leads: ' + error.message);
                });
        }
        
        // Export leads
        function exportLeads() {
            const csvContent = createCSV(leads);
            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const link = document.createElement('a');
            const url = URL.createObjectURL(blob);
            
            link.setAttribute('href', url);
            link.setAttribute('download', `leads_${deviceId}_${new Date().toISOString().split('T')[0]}.csv`);
            link.style.visibility = 'hidden';
            
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        }
        
        // Create CSV content
        function createCSV(leadsData) {
            const headers = ['name', 'phone', 'niche', 'target_status', 'trigger'];
            const rows = leadsData.map(lead => [
                lead.name || '',
                lead.phone || '',
                lead.niche || '',
                lead.target_status || lead.status || '',
                lead.trigger || ''
            ]);
            
            const csvContent = [
                headers.join(','),
                ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
            ].join('\\n');
            
            return csvContent;
        }
        
        // Import leads
        function importLeads(input) {
            const file = input.files[0];
            if (file) {
                importModal.show();
                document.getElementById('importFileInput').files = input.files;
            }
            input.value = ''; // Clear input
        }
        
        // Process import
        function processImport() {
            const fileInput = document.getElementById('importFileInput');
            const file = fileInput.files[0];
            
            if (!file) {
                showAlert('danger', 'Please select a file');
                return;
            }
            
            const reader = new FileReader();
            reader.onload = function(e) {
                const csv = e.target.result;
                const lines = csv.split(/\\r?\\n/);
                
                if (lines.length < 2) {
                    showAlert('danger', 'CSV file is empty or invalid');
                    return;
                }
                
                const headers = lines[0].toLowerCase().split(',').map(h => h.trim().replace(/"/g, ''));
                const newLeads = [];
                
                for (let i = 1; i < lines.length; i++) {
                    if (!lines[i].trim()) continue;
                    
                    const values = parseCSVLine(lines[i]);
                    const lead = {};
                    
                    headers.forEach((header, index) => {
                        if (values[index]) {
                            lead[header] = values[index].trim();
                        }
                    });
                    
                    if (lead.phone && lead.name) {
                        newLeads.push({
                            device_id: deviceId,
                            phone: lead.phone,
                            name: lead.name,
                            niche: lead.niche || '',
                            status: lead.target_status || lead.status || 'prospect',
                            trigger: lead.trigger || ''
                        });
                    }
                }
                
                if (newLeads.length === 0) {
                    showAlert('danger', 'No valid leads found in the CSV file');
                    return;
                }
                
                // Import leads one by one
                const promises = newLeads.map(lead => 
                    fetch(`/api/public/device/${deviceId}/lead`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json'
                        },
                        body: JSON.stringify(lead)
                    })
                );
                
                Promise.all(promises)
                    .then(() => {
                        importModal.hide();
                        loadLeads();
                        showAlert('success', `Successfully imported ${newLeads.length} leads!`);
                    })
                    .catch(error => {
                        showAlert('danger', 'Error importing leads: ' + error.message);
                    });
            };
            
            reader.readAsText(file);
        }
        
        // Parse CSV line
        function parseCSVLine(line) {
            const values = [];
            let current = '';
            let inQuotes = false;
            
            for (let i = 0; i < line.length; i++) {
                const char = line[i];
                
                if (char === '"') {
                    inQuotes = !inQuotes;
                } else if (char === ',' && !inQuotes) {
                    values.push(current.trim());
                    current = '';
                } else {
                    current += char;
                }
            }
            
            values.push(current.trim());
            return values;
        }
        
        // Show alert
        function showAlert(type, message) {
            const alertDiv = document.createElement('div');
            alertDiv.className = `alert alert-${type} alert-dismissible fade show position-fixed top-0 start-50 translate-middle-x mt-3`;
            alertDiv.style.zIndex = '9999';
            alertDiv.innerHTML = `
                ${message}
                <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            `;
            document.body.appendChild(alertDiv);
            
            setTimeout(() => {
                alertDiv.remove();
            }, 5000);
        }
'''

    # Insert the functions
    new_content = content[:insert_position] + missing_functions + '\n' + content[insert_position:]
    
    with open(r'C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\views\public_device_leads.html', 'w', encoding='utf-8') as f:
        f.write(new_content)
    
    print("Added all missing functions!")
else:
    print("Could not find insertion point")
