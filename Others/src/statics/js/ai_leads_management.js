// AI Lead Management Functions - Exact copy of device leads UI

// Function to load AI leads
function loadAILeads() {
    fetch('/api/leads-ai')
        .then(response => response.json())
        .then(data => {
            if (data.code === 'SUCCESS') {
                aiLeads = data.results || [];
                renderAILeads();
                updateAILeadStats();
            } else {
                console.error('Failed to load AI leads:', data.message);
                showAlert('danger', 'Failed to load AI leads: ' + data.message);
            }
        })
        .catch(error => {
            console.error('Error loading AI leads:', error);
            showAlert('danger', 'Error loading AI leads');
        });
}

// Render AI leads table
function renderAILeads() {
    const tbody = $('#aiLeadsTableBody');
    tbody.empty();
    
    aiLeads.forEach(lead => {
        const row = `
            <tr>
                <td>${lead.name}</td>
                <td>${lead.phone}</td>
                <td>${lead.email || '-'}</td>
                <td><span class="badge bg-info">${lead.niche || '-'}</span></td>
                <td><span class="badge ${lead.target_status === 'customer' ? 'bg-success' : 'bg-warning'}">${lead.target_status}</span></td>
                <td><span class="badge ${getStatusBadgeClass(lead.status)}">${lead.status}</span></td>
                <td>${lead.device_id || 'Unassigned'}</td>
                <td>
                    <button class="btn btn-sm btn-outline-primary" onclick="editAILead(${lead.id})">
                        <i class="bi bi-pencil"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteAILead(${lead.id})">
                        <i class="bi bi-trash"></i>
                    </button>
                </td>
            </tr>
        `;
        tbody.append(row);
    });
}

// Add new AI lead modal
function showAddAILeadModal() {
    $('#aiLeadModal').modal('show');
    $('#aiLeadModalTitle').text('Add AI Lead');
    $('#aiLeadForm')[0].reset();
    $('#aiLeadId').val('');
}

// Save AI lead
function saveAILead() {
    const leadId = $('#aiLeadId').val();
    const leadData = {
        name: $('#aiLeadName').val(),
        phone: $('#aiLeadPhone').val(),
        email: $('#aiLeadEmail').val(),
        niche: $('#aiLeadNiche').val(),
        target_status: $('#aiLeadTargetStatus').val(),
        notes: $('#aiLeadNotes').val()
    };
    
    const url = leadId ? `/api/leads-ai/${leadId}` : '/api/leads-ai';
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
            $('#aiLeadModal').modal('hide');
            loadAILeads();
            showAlert('success', leadId ? 'AI Lead updated successfully' : 'AI Lead added successfully');
        } else {
            showAlert('danger', data.message || 'Failed to save AI lead');
        }
    })
    .catch(error => {
        showAlert('danger', 'Error: ' + error.message);
    });
}

// Export AI leads
function exportAILeads() {
    const csv = convertAILeadsToCSV(aiLeads);
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `ai_leads_${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
}

// Convert AI leads to CSV
function convertAILeadsToCSV(data) {
    const headers = ['name', 'phone', 'email', 'niche', 'target_status', 'status', 'notes'];
    const rows = data.map(lead => [
        lead.name,
        lead.phone,
        lead.email || '',
        lead.niche || '',
        lead.target_status || 'prospect',
        lead.status || 'pending',
        lead.notes || ''
    ]);
    
    const csvContent = [
        headers.join(','),
        ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
    ].join('\n');
    
    return csvContent;
}

// Import AI leads
function importAILeads() {
    $('#aiImportModal').modal('show');
}

// Process AI leads import
function processAILeadsImport() {
    const file = document.getElementById('aiImportFile').files[0];
    if (!file) {
        showAlert('danger', 'Please select a file');
        return;
    }
    
    const reader = new FileReader();
    reader.onload = function(e) {
        const csv = e.target.result;
        const lines = csv.split('\n').filter(line => line.trim());
        const headers = lines[0].split(',').map(h => h.trim().toLowerCase().replace(/"/g, ''));
        
        const nameIndex = headers.indexOf('name');
        const phoneIndex = headers.indexOf('phone');
        
        if (nameIndex === -1 || phoneIndex === -1) {
            showAlert('danger', 'CSV must have "name" and "phone" columns');
            return;
        }
        
        const emailIndex = headers.indexOf('email');
        const nicheIndex = headers.indexOf('niche');
        const targetStatusIndex = headers.indexOf('target_status');
        const notesIndex = headers.indexOf('notes') !== -1 ? headers.indexOf('notes') : headers.indexOf('additional_note');
        
        const leadsToImport = [];
        for (let i = 1; i < lines.length; i++) {
            const values = lines[i].split(',').map(v => v.trim().replace(/"/g, ''));
            
            const lead = {
                name: values[nameIndex],
                phone: values[phoneIndex],
                email: emailIndex !== -1 ? values[emailIndex] : '',
                niche: nicheIndex !== -1 ? values[nicheIndex] : '',
                target_status: targetStatusIndex !== -1 ? values[targetStatusIndex] : 'prospect',
                notes: notesIndex !== -1 ? values[notesIndex] : ''
            };
            
            if (lead.name && lead.phone) {
                leadsToImport.push(lead);
            }
        }
        
        if (leadsToImport.length === 0) {
            showAlert('danger', 'No valid leads found in CSV');
            return;
        }
        
        // Import leads one by one
        let imported = 0;
        let failed = 0;
        
        const importNext = (index) => {
            if (index >= leadsToImport.length) {
                $('#aiImportModal').modal('hide');
                loadAILeads();
                showAlert('success', `Import complete: ${imported} imported, ${failed} failed`);
                return;
            }
            
            fetch('/api/leads-ai', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(leadsToImport[index])
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    imported++;
                } else {
                    failed++;
                }
                importNext(index + 1);
            })
            .catch(() => {
                failed++;
                importNext(index + 1);
            });
        };
        
        importNext(0);
    };
    
    reader.readAsText(file);
}

// Edit AI lead
function editAILead(leadId) {
    const lead = aiLeads.find(l => l.id === leadId);
    if (!lead) return;
    
    $('#aiLeadModal').modal('show');
    $('#aiLeadModalTitle').text('Edit AI Lead');
    $('#aiLeadId').val(lead.id);
    $('#aiLeadName').val(lead.name);
    $('#aiLeadPhone').val(lead.phone);
    $('#aiLeadEmail').val(lead.email || '');
    $('#aiLeadNiche').val(lead.niche || '');
    $('#aiLeadTargetStatus').val(lead.target_status || 'prospect');
    $('#aiLeadNotes').val(lead.notes || '');
}

// Delete AI lead
function deleteAILead(leadId) {
    if (!confirm('Are you sure you want to delete this AI lead?')) {
        return;
    }
    
    fetch(`/api/leads-ai/${leadId}`, {
        method: 'DELETE'
    })
    .then(response => response.json())
    .then(data => {
        if (data.code === 'SUCCESS') {
            loadAILeads();
            showAlert('success', 'AI Lead deleted successfully');
        } else {
            showAlert('danger', data.message || 'Failed to delete AI lead');
        }
    })
    .catch(error => {
        showAlert('danger', 'Error: ' + error.message);
    });
}

// Update AI lead stats
function updateAILeadStats() {
    $('#totalAILeads').text(aiLeads.length);
    const pending = aiLeads.filter(l => l.status === 'pending').length;
    const sent = aiLeads.filter(l => l.status === 'sent').length;
    $('#pendingAILeads').text(pending);
    $('#sentAILeads').text(sent);
}

// Search AI leads
function searchAILeads() {
    const searchTerm = $('#aiSearchInput').val().toLowerCase();
    const filtered = aiLeads.filter(lead => 
        lead.name.toLowerCase().includes(searchTerm) ||
        lead.phone.includes(searchTerm) ||
        (lead.niche && lead.niche.toLowerCase().includes(searchTerm))
    );
    
    const tbody = $('#aiLeadsTableBody');
    tbody.empty();
    
    filtered.forEach(lead => {
        const row = `
            <tr>
                <td>${lead.name}</td>
                <td>${lead.phone}</td>
                <td>${lead.email || '-'}</td>
                <td><span class="badge bg-info">${lead.niche || '-'}</span></td>
                <td><span class="badge ${lead.target_status === 'customer' ? 'bg-success' : 'bg-warning'}">${lead.target_status}</span></td>
                <td><span class="badge ${getStatusBadgeClass(lead.status)}">${lead.status}</span></td>
                <td>${lead.device_id || 'Unassigned'}</td>
                <td>
                    <button class="btn btn-sm btn-outline-primary" onclick="editAILead(${lead.id})">
                        <i class="bi bi-pencil"></i>
                    </button>
                    <button class="btn btn-sm btn-outline-danger" onclick="deleteAILead(${lead.id})">
                        <i class="bi bi-trash"></i>
                    </button>
                </td>
            </tr>
        `;
        tbody.append(row);
    });
}

// Get status badge class
function getStatusBadgeClass(status) {
    switch(status) {
        case 'sent': return 'bg-success';
        case 'pending': return 'bg-warning';
        case 'failed': return 'bg-danger';
        case 'assigned': return 'bg-info';
        default: return 'bg-secondary';
    }
}

// Initialize on page load
$(document).ready(function() {
    loadAILeads();
});
// AI Campaign Management Functions

function showCreateAICampaignModal() {
    $('#aiCampaignModal').modal('show');
    $('#aiCampaignForm')[0].reset();
    
    // Set default date to today
    const today = new Date().toISOString().split('T')[0];
    $('#aiCampaignDate').val(today);
}

function saveAICampaign() {
    const campaignData = {
        title: $('#aiCampaignTitle').val(),
        niche: $('#aiCampaignNiche').val(),
        target_status: $('#aiCampaignTargetStatus').val(),
        message: $('#aiCampaignMessage').val(),
        image_url: $('#aiCampaignImageUrl').val(),
        campaign_date: $('#aiCampaignDate').val(),
        time_schedule: $('#aiCampaignTime').val(),
        limit: parseInt($('#aiDeviceLimit').val()), // Device limit for AI campaign
        min_delay_seconds: parseInt($('#aiMinDelay').val()),
        max_delay_seconds: parseInt($('#aiMaxDelay').val()),
        ai: 'ai' // Mark as AI campaign
    };
    
    // Validate
    if (!campaignData.title || !campaignData.message || !campaignData.limit) {
        showAlert('danger', 'Please fill all required fields');
        return;
    }
    
    fetch('/api/campaigns', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(campaignData)
    })
    .then(response => response.json())
    .then(data => {
        if (data.code === 'SUCCESS') {
            $('#aiCampaignModal').modal('hide');
            loadCampaigns(); // Refresh campaigns list
            showAlert('success', 'AI Campaign created successfully');
        } else {
            showAlert('danger', data.message || 'Failed to create AI campaign');
        }
    })
    .catch(error => {
        showAlert('danger', 'Error: ' + error.message);
    });
}

// Variable to store AI leads
let aiLeads = [];

// Alert function
function showAlert(type, message) {
    const alertHtml = `
        <div class="alert alert-${type} alert-dismissible fade show" role="alert">
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;
    $('#alertContainer').html(alertHtml);
    
    // Auto dismiss after 5 seconds
    setTimeout(() => {
        $('.alert').alert('close');
    }, 5000);
}
