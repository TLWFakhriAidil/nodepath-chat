// AI Lead Management JavaScript Functions
// Add these functions to your dashboard JavaScript

// Load AI Leads
function loadAILeads() {
    fetch('/api/leads-ai')
        .then(response => response.json())
        .then(data => {
            if (data.code === 'SUCCESS' && data.results) {
                displayAILeads(data.results);
                updateAILeadStats(data.results);
            }
        })
        .catch(error => {
            console.error('Error loading AI leads:', error);
            showToast('Failed to load AI leads', 'error');
        });
}

// Display AI Leads in table
function displayAILeads(leads) {
    const tbody = document.getElementById('aiLeadsTableBody');
    tbody.innerHTML = '';
    
    if (!leads || leads.length === 0) {
        tbody.innerHTML = '<tr><td colspan="8" class="text-center">No AI leads found</td></tr>';
        return;
    }
    
    leads.forEach(lead => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${lead.name}</td>
            <td>${lead.phone}</td>
            <td>${lead.niche || '-'}</td>
            <td><span class="badge bg-${lead.target_status === 'customer' ? 'success' : 'primary'}">${lead.target_status}</span></td>
            <td>${getStatusBadge(lead.status)}</td>
            <td>${lead.device_id || '-'}</td>
            <td>${formatDate(lead.created_at)}</td>
            <td>
                <button class="btn btn-sm btn-primary" onclick="editAILead(${lead.id})">
                    <i class="bi bi-pencil"></i>
                </button>
                <button class="btn btn-sm btn-danger" onclick="deleteAILead(${lead.id})">
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        `;
        tbody.appendChild(row);
    });
}
// Update AI Lead statistics
function updateAILeadStats(leads) {
    const stats = {
        total: leads.length,
        pending: 0,
        sent: 0,
        failed: 0
    };
    
    leads.forEach(lead => {
        switch(lead.status) {
            case 'pending':
                stats.pending++;
                break;
            case 'sent':
                stats.sent++;
                break;
            case 'failed':
                stats.failed++;
                break;
        }
    });
    
    document.getElementById('aiTotalLeads').textContent = stats.total;
    document.getElementById('aiPendingLeads').textContent = stats.pending;
    document.getElementById('aiSentLeads').textContent = stats.sent;
    document.getElementById('aiFailedLeads').textContent = stats.failed;
}

// Get status badge HTML
function getStatusBadge(status) {
    const badges = {
        'pending': '<span class="badge bg-warning">Pending</span>',
        'assigned': '<span class="badge bg-info">Assigned</span>',
        'sent': '<span class="badge bg-success">Sent</span>',
        'failed': '<span class="badge bg-danger">Failed</span>'
    };
    return badges[status] || `<span class="badge bg-secondary">${status}</span>`;
}
// Show Add AI Lead Modal
function showAddAILeadModal() {
    document.getElementById('aiLeadModalTitle').textContent = 'Add AI Lead';
    document.getElementById('aiLeadForm').reset();
    document.getElementById('aiLeadId').value = '';
    $('#aiLeadModal').modal('show');
}

// Save AI Lead
function saveAILead() {
    const leadId = document.getElementById('aiLeadId').value;
    const isEdit = !!leadId;
    
    const leadData = {
        name: document.getElementById('aiLeadName').value,
        phone: document.getElementById('aiLeadPhone').value,
        email: document.getElementById('aiLeadEmail').value,
        niche: document.getElementById('aiLeadNiche').value,
        target_status: document.getElementById('aiLeadTargetStatus').value,
        notes: document.getElementById('aiLeadNotes').value
    };
    
    // Validate required fields
    if (!leadData.name || !leadData.phone) {
        showToast('Name and phone are required', 'error');
        return;
    }
    
    const url = isEdit ? `/api/leads-ai/${leadId}` : '/api/leads-ai';
    const method = isEdit ? 'PUT' : 'POST';
    
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
            showToast(isEdit ? 'AI Lead updated successfully' : 'AI Lead added successfully', 'success');
        } else {
            showToast(data.message || 'Failed to save AI lead', 'error');
        }
    })
    .catch(error => {
        console.error('Error saving AI lead:', error);
        showToast('Failed to save AI lead', 'error');
    });
}
// Edit AI Lead
function editAILead(leadId) {
    fetch(`/api/leads-ai`)
        .then(response => response.json())
        .then(data => {
            if (data.code === 'SUCCESS' && data.results) {
                const lead = data.results.find(l => l.id === leadId);
                if (lead) {
                    document.getElementById('aiLeadModalTitle').textContent = 'Edit AI Lead';
                    document.getElementById('aiLeadId').value = lead.id;
                    document.getElementById('aiLeadName').value = lead.name;
                    document.getElementById('aiLeadPhone').value = lead.phone;
                    document.getElementById('aiLeadEmail').value = lead.email || '';
                    document.getElementById('aiLeadNiche').value = lead.niche || '';
                    document.getElementById('aiLeadTargetStatus').value = lead.target_status;
                    document.getElementById('aiLeadNotes').value = lead.notes || '';
                    $('#aiLeadModal').modal('show');
                }
            }
        })
        .catch(error => {
            console.error('Error loading lead details:', error);
            showToast('Failed to load lead details', 'error');
        });
}

// Delete AI Lead
function deleteAILead(leadId) {
    Swal.fire({
        title: 'Delete AI Lead?',
        text: "This action cannot be undone!",
        icon: 'warning',
        showCancelButton: true,
        confirmButtonColor: '#d33',
        cancelButtonColor: '#3085d6',
        confirmButtonText: 'Yes, delete it!'
    }).then((result) => {
        if (result.isConfirmed) {
            fetch(`/api/leads-ai/${leadId}`, {
                method: 'DELETE'
            })
            .then(response => response.json())
            .then(data => {
                if (data.code === 'SUCCESS') {
                    loadAILeads();
                    showToast('AI Lead deleted successfully', 'success');
                } else {
                    showToast(data.message || 'Failed to delete AI lead', 'error');
                }
            })
            .catch(error => {
                console.error('Error deleting AI lead:', error);
                showToast('Failed to delete AI lead', 'error');
            });
        }
    });
}
// Show Create AI Campaign Modal
function showCreateAICampaignModal() {
    document.getElementById('aiCampaignForm').reset();
    // Set default date to today
    document.getElementById('aiCampaignDate').value = new Date().toISOString().split('T')[0];
    $('#aiCampaignModal').modal('show');
}

// Save AI Campaign
function saveAICampaign() {
    const campaignData = {
        title: document.getElementById('aiCampaignTitle').value,
        niche: document.getElementById('aiCampaignNiche').value,
        target_status: document.getElementById('aiCampaignTargetStatus').value,
        message: document.getElementById('aiCampaignMessage').value,
        image_url: document.getElementById('aiCampaignImageUrl').value,
        campaign_date: document.getElementById('aiCampaignDate').value,
        time_schedule: document.getElementById('aiCampaignTime').value,
        limit: parseInt(document.getElementById('aiDeviceLimit').value),
        min_delay_seconds: parseInt(document.getElementById('aiMinDelay').value),
        max_delay_seconds: parseInt(document.getElementById('aiMaxDelay').value),
        ai: 'ai' // Mark as AI campaign
    };
    
    // Validate required fields
    if (!campaignData.title || !campaignData.niche || !campaignData.message || 
        !campaignData.campaign_date || !campaignData.limit) {
        showToast('Please fill all required fields', 'error');
        return;
    }
    
    if (campaignData.limit <= 0) {
        showToast('Device limit must be greater than 0', 'error');
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
            loadCampaigns(); // Reload campaigns list
            showToast('AI Campaign created successfully', 'success');
            
            // Ask if user wants to trigger campaign now
            Swal.fire({
                title: 'Campaign Created!',
                text: 'Do you want to trigger the campaign now?',
                icon: 'success',
                showCancelButton: true,
                confirmButtonText: 'Yes, trigger now',
                cancelButtonText: 'No, trigger later'
            }).then((result) => {
                if (result.isConfirmed) {
                    triggerAICampaign(data.results.id);
                }
            });
        } else {
            showToast(data.message || 'Failed to create AI campaign', 'error');
        }
    })
    .catch(error => {
        console.error('Error creating AI campaign:', error);
        showToast('Failed to create AI campaign', 'error');
    });
}
// Trigger AI Campaign
function triggerAICampaign(campaignId) {
    showToast('Triggering AI campaign...', 'info');
    
    fetch(`/api/campaigns-ai/${campaignId}/trigger`, {
        method: 'POST'
    })
    .then(response => response.json())
    .then(data => {
        if (data.code === 'SUCCESS') {
            showToast('AI Campaign triggered successfully! Processing leads...', 'success');
            loadCampaigns(); // Refresh campaign list
            
            // Show progress modal
            showAICampaignProgress(campaignId);
        } else {
            showToast(data.message || 'Failed to trigger AI campaign', 'error');
        }
    })
    .catch(error => {
        console.error('Error triggering AI campaign:', error);
        showToast('Failed to trigger AI campaign', 'error');
    });
}

// Show AI Campaign Progress (optional - for real-time monitoring)
function showAICampaignProgress(campaignId) {
    // This could open a modal or redirect to a progress page
    console.log('Monitoring campaign progress:', campaignId);
    // Implementation depends on whether you want real-time updates via WebSocket
}

// Helper function to format date
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString();
}

// Helper function to show toast notifications
function showToast(message, type = 'info') {
    // If using a toast library
    if (window.toastr) {
        toastr[type](message);
    } else if (window.Swal) {
        // Fallback to SweetAlert2
        const Toast = Swal.mixin({
            toast: true,
            position: 'top-end',
            showConfirmButton: false,
            timer: 3000,
            timerProgressBar: true
        });
        
        Toast.fire({
            icon: type === 'error' ? 'error' : type === 'success' ? 'success' : 'info',
            title: message
        });
    } else {
        // Fallback to console
        console.log(`[${type.toUpperCase()}] ${message}`);
    }
}

// Initialize AI Management when tab is clicked
document.addEventListener('DOMContentLoaded', function() {
    const aiTab = document.getElementById('manage-ai-tab');
    if (aiTab) {
        aiTab.addEventListener('shown.bs.tab', function() {
            loadAILeads();
        });
    }
});