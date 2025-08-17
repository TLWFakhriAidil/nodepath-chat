# Fix: Public Device Sequence Summary Shows 0

## Root Cause Analysis

The sequence summary shows 0 in the public device view due to multiple issues:

1. **Incorrect API Endpoint**
   - Frontend calls: `/api/public/device/${publicDeviceId}/sequence-summary`
   - Backend provides: `/api/public/device/:device_name/sequences`

2. **Wrong Parameter Type**
   - Frontend uses: `publicDeviceId` (UUID)
   - Backend expects: `device_name` (string name like "Device-123")

3. **Response Format Mismatch**
   - Backend returns: Array of sequence objects
   - Frontend expects: Summary object with specific structure

## Solution

### Step 1: Fix the Frontend URL (in public_device.html)

Find the `loadSequenceSummary` function around line 6209 and replace:

```javascript
// OLD CODE (line 6211-6213):
let url = isPublicView
    ? `/api/public/device/${publicDeviceId}/sequence-summary`
    : '/api/sequences/summary';
```

With:

```javascript
// NEW CODE:
// Extract device name from URL for public view
let deviceName = '';
if (isPublicView) {
    // URL is typically like: /Device-Name or ends with /Device-Name
    const pathParts = window.location.pathname.split('/');
    deviceName = pathParts[pathParts.length - 1];
    
    // Remove any query parameters
    if (deviceName.includes('?')) {
        deviceName = deviceName.split('?')[0];
    }
}

let url = isPublicView
    ? `/api/public/device/${deviceName}/sequences`  // Correct endpoint with device name
    : '/api/sequences/summary';
```

### Step 2: Fix the Response Handling

The `displaySequenceSummary` function needs to properly handle the array response from the public API.

The backend returns:
```json
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
```

Update the response handling in `loadSequenceSummary` (around line 6246):

```javascript
.then(data => {
    console.log('Sequence Summary Response:', data); // Debug log
    if (data.code === 'SUCCESS' && data.results) {
        // For public view, transform the array response
        if (isPublicView && Array.isArray(data.results)) {
            const sequences = data.results;
            const transformedData = {
                sequences: sequences,
                total: sequences.length,
                total_should_send: sequences.reduce((sum, s) => sum + (s.total_contacts || 0), 0),
                total_done_send: sequences.reduce((sum, s) => sum + (s.contacts_done || 0), 0),
                total_failed_send: sequences.reduce((sum, s) => sum + (s.contacts_failed || 0), 0),
                total_remaining_send: sequences.reduce((sum, s) => sum + ((s.total_contacts || 0) - (s.contacts_done || 0) - (s.contacts_failed || 0)), 0)
            };
            displaySequenceSummary(transformedData);
        } else {
            displaySequenceSummary(data.results);
        }
    } else {
        document.getElementById('sequenceSummaryContent').innerHTML =
            '<div class="alert alert-danger">Failed to load sequence summary</div>';
    }
})
```

### Step 3: Update displaySequenceSummary to Handle Public View Data

In the `displaySequenceSummary` function, add handling for the public view format:

```javascript
function displaySequenceSummary(summary) {
    console.log('DisplaySequenceSummary called with:', summary); // Debug log
    
    // Handle public view format
    if (summary.sequences && Array.isArray(summary.sequences)) {
        const sequences = summary.sequences;
        
        // Build the HTML for public view
        let html = `
            <div class="row mb-4">
                <div class="col-md-3">
                    <div class="card">
                        <div class="card-body">
                            <h5 class="card-title">Total Sequences</h5>
                            <h2>${summary.total || sequences.length}</h2>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card">
                        <div class="card-body">
                            <h5 class="card-title">Should Send</h5>
                            <h2>${summary.total_should_send || 0}</h2>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card">
                        <div class="card-body">
                            <h5 class="card-title">Done Send</h5>
                            <h2 class="text-success">${summary.total_done_send || 0}</h2>
                        </div>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="card">
                        <div class="card-body">
                            <h5 class="card-title">Failed Send</h5>
                            <h2 class="text-danger">${summary.total_failed_send || 0}</h2>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="table-responsive">
                <table class="table table-hover">
                    <thead>
                        <tr>
                            <th>Sequence Name</th>
                            <th>Trigger</th>
                            <th>Total Flows</th>
                            <th>Total Contacts</th>
                            <th>Done</th>
                            <th>Failed</th>
                            <th>Success Rate</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${sequences.map(seq => `
                            <tr>
                                <td>${seq.name}</td>
                                <td><span class="badge bg-info">${seq.trigger || 'N/A'}</span></td>
                                <td>${seq.total_flows || 0}</td>
                                <td>${seq.total_contacts || 0}</td>
                                <td class="text-success">${seq.contacts_done || 0}</td>
                                <td class="text-danger">${seq.contacts_failed || 0}</td>
                                <td>
                                    <div class="progress" style="height: 20px;">
                                        <div class="progress-bar bg-success" style="width: ${seq.success_rate || 0}%">
                                            ${seq.success_rate || 0}%
                                        </div>
                                    </div>
                                </td>
                            </tr>
                        `).join('')}
                    </tbody>
                </table>
            </div>
        `;
        
        document.getElementById('sequenceSummaryContent').innerHTML = html;
        return;
    }
    
    // Continue with existing display logic for non-public views...
}
```

## Summary

The issue occurs because:
1. Frontend calls wrong endpoint URL
2. Uses device ID instead of device name
3. Expects different response format

The fix:
1. Use correct endpoint: `/api/public/device/${deviceName}/sequences`
2. Extract device name from URL
3. Transform the array response to match expected format
4. Update display function to handle public view data