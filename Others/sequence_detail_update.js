// Add this function to sequence_detail.html after loadSequenceDetails function

function loadDeviceReport() {
    // Build URL with date filters
    let url = `/api/sequences/${sequenceId}/device-report`;
    const startDate = document.getElementById('filterStartDate').value;
    const endDate = document.getElementById('filterEndDate').value;
    
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    
    if (params.toString()) {
        url += '?' + params.toString();
    }
    
    fetch(url)
        .then(response => response.json())
        .then(data => {
            if (data.code === 'SUCCESS' && data.results) {
                const report = data.results;
                
                // Update overall statistics
                document.getElementById('totalShouldSend').textContent = report.shouldSend || 0;
                document.getElementById('totalDoneSend').textContent = report.doneSend || 0;
                document.getElementById('totalFailedSend').textContent = report.failedSend || 0;
                document.getElementById('totalRemainingSend').textContent = report.remainingSend || 0;
                document.getElementById('totalLeads').textContent = report.totalLeads || 0;
                
                // Update step-wise statistics if needed
                if (report.devices && report.devices.length > 0) {
                    // Aggregate step data from all devices
                    const stepTotals = {};
                    
                    report.devices.forEach(device => {
                        if (device.steps) {
                            device.steps.forEach(step => {
                                if (!stepTotals[step.step_id]) {
                                    stepTotals[step.step_id] = {
                                        shouldSend: 0,
                                        doneSend: 0,
                                        failedSend: 0,
                                        remainingSend: 0,
                                        totalLeads: 0
                                    };
                                }
                                stepTotals[step.step_id].shouldSend += step.should_send || 0;
                                stepTotals[step.step_id].doneSend += step.done_send || 0;
                                stepTotals[step.step_id].failedSend += step.failed_send || 0;
                                stepTotals[step.step_id].remainingSend += step.remaining_send || 0;
                                stepTotals[step.step_id].totalLeads += step.total_leads || 0;
                            });
                        }
                    });
                    
                    // Update flowStats with the aggregated data
                    Object.keys(stepTotals).forEach(stepId => {
                        flowStats[stepId] = stepTotals[stepId];
                    });
                    
                    // Refresh the display
                    displayFlowCards();
                }
            }
        })
        .catch(error => {
            console.error('Error loading device report:', error);
        });
}

// Update loadSequenceDetails to also load device report
// Replace the existing Promise.all section with:
Promise.all([
    loadLeadsCount(),
    loadSequenceContacts(),
    loadDeviceReport()  // Add this
]).then(() => {
    displayFlowCards();
    displayTimeline();
});

// Update applyDateFilter to reload device report
function applyDateFilter() {
    const startDate = document.getElementById('filterStartDate').value;
    const endDate = document.getElementById('filterEndDate').value;
    
    // Reload device report with new date filters
    loadDeviceReport();
    
    // ... rest of existing filter logic
}