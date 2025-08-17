// FIX: Public Device Sequence Summary Shows 0
// Problem: Frontend is calling wrong endpoint URL

// ISSUE IDENTIFIED:
// 1. Frontend calls: /api/public/device/${publicDeviceId}/sequence-summary
// 2. Backend endpoint is: /api/public/device/:device_name/sequences
// 3. Also, frontend uses publicDeviceId but backend expects device_name

// SOLUTION 1: Fix the frontend to use correct endpoint
// In public_device.html, change line 6212:

// FROM:
// let url = isPublicView
//     ? `/api/public/device/${publicDeviceId}/sequence-summary`
//     : '/api/sequences/summary';

// TO:
// let url = isPublicView
//     ? `/api/public/device/${publicDeviceName}/sequences`  // Use device name, not ID
//     : '/api/sequences/summary';

// SOLUTION 2: Make sure publicDeviceName is available
// Check if publicDeviceName is defined in the HTML. If not, add it:

// In the HTML head or script section, add:
const publicDeviceName = '<?php echo $device_name; ?>'; // or however the device name is passed

// OR extract from URL:
const pathParts = window.location.pathname.split('/');
const publicDeviceName = pathParts[pathParts.length - 1]; // Gets last part of URL

// COMPLETE FIX for loadSequenceSummary function:
function loadSequenceSummary() {
    // For public view, extract device name from URL
    let deviceName = '';
    if (isPublicView) {
        const pathParts = window.location.pathname.split('/');
        // URL format is typically: /Device-Name or /public/Device-Name
        deviceName = pathParts[pathParts.length - 1];
        
        // Remove any query parameters if present
        if (deviceName.includes('?')) {
            deviceName = deviceName.split('?')[0];
        }
    }
    
    // Use different API endpoint for public view
    let url = isPublicView
        ? `/api/public/device/${deviceName}/sequences`  // Correct endpoint
        : '/api/sequences/summary';
    
    // Build query parameters
    const params = new URLSearchParams();
    
    // If we have date filters, use them
    if (sequenceStartDate || sequenceEndDate) {
        if (sequenceStartDate) params.append('start_date', sequenceStartDate);
        if (sequenceEndDate) params.append('end_date', sequenceEndDate);
    } else if (sequenceShowTodayOnly) {
        // Only apply today filter if explicitly requested
        const today = new Date().toISOString().split('T')[0];
        params.append('start_date', today);
        params.append('end_date', today);
        
        // Set the date inputs to today
        document.getElementById('sequenceStartDate').value = today;
        document.getElementById('sequenceEndDate').value = today;
    }
    
    if (params.toString()) {
        url += '?' + params.toString();
    }
    
    console.log('Loading sequence summary with URL:', url);
    console.log('Device Name:', deviceName);
    
    const fetchOptions = isPublicView ? {} : { credentials: 'include' };
    
    fetch(url, fetchOptions)
        .then(response => response.json())
        .then(data => {
            console.log('Sequence Summary Response:', data); // Debug log
            if (data.code === 'SUCCESS' && data.results) {
                displaySequenceSummary(data.results);
            } else if (data.sequences || data.total !== undefined) {
                // Handle public API response format
                displaySequenceSummary(data);
            } else {
                document.getElementById('sequenceSummaryContent').innerHTML =
                    '<div class="alert alert-danger">Failed to load sequence summary</div>';
            }
        })
        .catch(error => {
            console.error('Error loading sequence summary:', error);
            document.getElementById('sequenceSummaryContent').innerHTML =
                '<div class="alert alert-danger">Error loading data</div>';
        });
}

// ALSO CHECK: The response format from GetPublicDeviceSequences
// The backend returns data in this format:
/*
{
    "status": 200,
    "code": "SUCCESS",
    "message": "Sequences retrieved successfully",
    "results": [
        {
            "id": "seq-123",
            "name": "Welcome Sequence",
            "trigger": "welcome",
            "total_flows": 7,
            "total_contacts": 350,
            "contacts_done": 300,
            "contacts_failed": 20,
            "success_rate": "85.7"
        }
    ]
}
*/

// Make sure displaySequenceSummary function handles this format correctly