// Fix 1: Update the modal title dynamically based on context
function showSequenceDeviceReport(sequence) {
    // Store sequence for use in other functions
    currentSequenceForReport = sequence;
    
    // Update modal title to "Sequence Device Report"
    document.querySelector('#campaignDeviceReportModal .modal-title').textContent = 'Sequence Device Report';
    
    // Update card header to "Sequence Details"
    document.querySelector('#campaignDeviceReportModal .card-header h6').textContent = 'Sequence Details';
    
    // Use the same modal as campaign report but populate with sequence data
    document.getElementById('reportTitle').textContent = sequence.name || '-';
    document.getElementById('reportNiche').textContent = sequence.niche || '-';
    document.getElementById('reportTarget').innerHTML = `<span class="badge bg-info">Sequence</span>`;
    document.getElementById('reportDate').textContent = sequence.created_at ? new Date(sequence.created_at).toLocaleDateString() : '-';
    document.getElementById('reportTime').textContent = sequence.trigger || '-';
    document.getElementById('reportStatus').innerHTML = `<span class="badge bg-${getStatusColor(sequence.status)}">${sequence.status}</span>`;
    
    // Show loading state
    document.getElementById('deviceReportTableBody').innerHTML = '<tr><td colspan="7" class="text-center"><div class="spinner-border spinner-border-sm"></div> Loading device report...</td></tr>';
    
    // Fetch device report data for sequence
    fetch(`/api/sequences/${sequence.id}/device-report`, { credentials: 'include' })
        .then(response => response.json())
        .then(data => {
            if (data.code === 'SUCCESS' && data.results) {
                currentDeviceReport = data.results; // Store the report globally
                displayDeviceReport(data.results);
            } else {
                document.getElementById('deviceReportTableBody').innerHTML = '<tr><td colspan="8" class="text-center text-danger">Failed to load device report</td></tr>';
            }
        })
        .catch(error => {
            console.error('Error loading device report:', error);
            document.getElementById('deviceReportTableBody').innerHTML = '<tr><td colspan="8" class="text-center text-danger">Error loading device report</td></tr>';
        });
    
    // Show modal
    const modal = new bootstrap.Modal(document.getElementById('campaignDeviceReportModal'));
    modal.show();
}

// Fix 2: Update showCampaignDeviceReport to reset titles
function showCampaignDeviceReport(campaign) {
    // Store campaign for use in other functions
    currentCampaignForReport = campaign;
    
    // Reset modal title to "Campaign Device Report"
    document.querySelector('#campaignDeviceReportModal .modal-title').textContent = 'Campaign Device Report';
    
    // Reset card header to "Campaign Details"
    document.querySelector('#campaignDeviceReportModal .card-header h6').textContent = 'Campaign Details';
    
    // ... rest of the existing code
}
